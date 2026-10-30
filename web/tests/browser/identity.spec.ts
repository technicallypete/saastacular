import { createHash, randomBytes } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import { dirname } from "node:path";

import AxeBuilder from "@axe-core/playwright";
import { request as playwrightRequest, type APIRequestContext, type Browser, type BrowserContext, type Cookie, type Page, type TestInfo } from "@playwright/test";

import { expectAxeClean } from "./support/contrast";
import { expectFocusMotion } from "./support/form-controls";
import { expect, test } from "./support/page-errors";
import { observe, sanitizeHtml, sanitizeUrl } from "./support/sanitize";

// L2 identity scenarios (G004 Plan r5 §8.2, r2 §8.2; docs/foundation/IDENTITY.md) in the disposable HTTPS identity
// environment: the real web and core images, PostgreSQL, the pinned WorkOS emulator and the TLS test proxy
// (infra/compose/identity.yml). Every page request of every browser context is captured and must stay on the app
// origins under test and the emulator (E-NET-1). Direct core, replica, web and emulator calls run only in separate
// driver request contexts without shared storage (E-NET-3). The emulator does not model hosted MFA, SSO reuse or a
// same-sid callback; those are L1, L3 and MQ evidence (IDENTITY.md). Pages that answer an error status are opened in
// extra pages, so the default page's S-ERRORS check covers only flows that end on a normal page.

const ORIGIN_A = "https://org-a.localhost:543";
const ORIGIN_B = "https://org-b.localhost:543";
const ATTACKER = "https://attacker.localhost:543";
const EMULATOR = "http://emulator:4100";
const CORE = "http://core:8180";
const CORE_REPLICA = "http://core-replica:8180";
const WEB_DIRECT = "http://127.0.0.1:3000";
const FORM = "application/x-www-form-urlencoded";
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const SECRET = /^[A-Za-z0-9_-]{43}$/;

type Credentials = { email: string; password: string };
type SeedUser = Credentials & { name: string; first_name: string; last_name: string };
type Seed = { users: SeedUser[]; canaries: Record<string, string>; emulator_api_key: string };
const SEED = JSON.parse(readFileSync("/identity/browser/users.json", "utf8")) as Seed;
function seeded(name: string): SeedUser {
  const found = SEED.users.find((u) => u.name === name);
  if (!found) throw new Error(`seed user ${name} missing`);
  return found;
}
const ALICE = seeded("alice");
const BOB = seeded("bob");
const CAROL = seeded("carol");
const HOSTILE = seeded("hostile");
const FRANK = seeded("frank");

function canary(): string {
  const value = `Cx${randomBytes(10).toString("hex")}`;
  observe(value);
  return value;
}

// --- page traffic (E-NET-1/2), extra contexts and drivers ---------------------------------------------------------------
type Tracking = { origins: Set<string>; allowed: Set<string>; contexts: BrowserContext[]; drivers: APIRequestContext[] };
let tracking: Tracking;

function track(context: BrowserContext) {
  context.on("request", (request) => {
    const url = request.url();
    if (/^https?:/.test(url)) tracking.origins.add(new URL(url).origin);
  });
}

test.beforeEach(async ({ context }) => {
  tracking = { origins: new Set(), allowed: new Set([ORIGIN_A, ORIGIN_B, EMULATOR]), contexts: [], drivers: [] };
  track(context);
});

test.afterEach(async () => {
  for (const c of tracking.contexts) await c.close();
  for (const d of tracking.drivers) await d.dispose();
  const outside = [...tracking.origins].filter((o) => !tracking.allowed.has(o));
  expect(outside, "E-NET-1: page requests outside the app origins under test and the emulator").toEqual([]);
});

// Another browser profile with the project's appearance, captured like the default context.
async function open(browser: Browser, testInfo: TestInfo): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({
    baseURL: ORIGIN_A,
    colorScheme: testInfo.project.use.colorScheme,
    viewport: testInfo.project.use.viewport,
  });
  track(context);
  tracking.contexts.push(context);
  return { context, page: await context.newPage() };
}

// A driver: a separate request context with no shared storage; never a page (E-NET-3).
async function driver(): Promise<APIRequestContext> {
  const d = await playwrightRequest.newContext();
  tracking.drivers.push(d);
  return d;
}

// --- sign-in through the emulator's hosted page ------------------------------------------------------------------------
async function atEmulator(page: Page) {
  await page.waitForURL((u) => u.origin === EMULATOR);
}

async function enterCredentials(page: Page, who: Credentials) {
  await atEmulator(page);
  await page.locator("#email").fill(who.email);
  await page.getByRole("button", { name: "Continue" }).click();
  const password = page.locator('input[name="password"]');
  await password.fill(who.password);
  await password.press("Enter");
}

// Completes the hosted sign-in and waits to be back on an application origin.
async function completeAtEmulator(page: Page, who: Credentials) {
  await enterCredentials(page, who);
  await page.waitForURL((u) => u.origin !== EMULATOR);
}

// Opens ORIGIN's sign-in page (returning to /account), submits it and stops on the emulator page.
async function startSignIn(page: Page, origin = ORIGIN_A) {
  await page.goto(`${origin}/sign-in?return_to=/account`);
  await page.getByRole("button", { name: "Sign in" }).click();
  await atEmulator(page);
}

async function signIn(page: Page, who: Credentials, origin = ORIGIN_A) {
  await startSignIn(page, origin);
  await completeAtEmulator(page, who);
}

// Completes the emulator sign-in but never delivers the callback: the final emulator POST is fetched without following
// its redirect (redirect hops are not routable) and the browser is shown a placeholder. Returns the callback URL.
async function captureCallback(page: Page, who: Credentials): Promise<string> {
  let captured = "";
  const pattern = `${EMULATOR}/user_management/authorize`;
  await page.route(pattern, async (route) => {
    if (route.request().method() !== "POST") return route.continue();
    const response = await route.fetch({ maxRedirects: 0 });
    const location = response.headers()["location"] ?? "";
    if (/^https:\/\/org-[ab]\.localhost:543\/auth\/callback\?/.test(location)) {
      captured = location;
      return route.fulfill({ status: 200, contentType: "text/plain", body: "callback captured" });
    }
    return route.fulfill({ response });
  });
  await enterCredentials(page, who);
  await expect.poll(() => captured).not.toBe("");
  await page.unroute(pattern);
  const url = new URL(captured);
  observe(url.searchParams.get("code"));
  observe(url.searchParams.get("state"));
  return captured;
}

function statusOf(url: string): string | null {
  return new URL(url).searchParams.get("status");
}

function pathOf(url: string): string {
  return new URL(url).pathname;
}

// --- observations -------------------------------------------------------------------------------------------------------
async function sessionCookie(context: BrowserContext, origin = ORIGIN_A): Promise<Cookie | undefined> {
  const cookie = (await context.cookies(origin)).find((c) => c.name === "__Host-sid");
  observe(cookie?.value);
  return cookie;
}

async function sessionValue(context: BrowserContext, origin = ORIGIN_A): Promise<string> {
  const cookie = await sessionCookie(context, origin);
  expect(cookie?.value).toMatch(SECRET);
  return cookie!.value;
}

async function bindingCookie(context: BrowserContext, origin = ORIGIN_A): Promise<Cookie | undefined> {
  const cookie = (await context.cookies(origin)).find((c) => c.name === "__Host-login");
  observe(cookie?.value);
  return cookie;
}

// Browser-side calls run as same-origin fetches from a helper page of the context: Chromium resolves *.localhost and
// trusts the pinned key (the Node request client does neither), and the calls carry exactly the browser's cookies.
const apiPages = new WeakMap<BrowserContext, Map<string, Page>>();
async function apiPage(context: BrowserContext, origin: string): Promise<Page> {
  let byOrigin = apiPages.get(context);
  if (!byOrigin) apiPages.set(context, (byOrigin = new Map()));
  let p = byOrigin.get(origin);
  if (!p || p.isClosed()) {
    p = await context.newPage();
    await p.goto(`${origin}/api/health`);
    byOrigin.set(origin, p);
  }
  return p;
}

type Account = { status: number; body: Record<string, unknown> };
async function account(context: BrowserContext, origin = ORIGIN_A): Promise<Account> {
  const p = await apiPage(context, origin);
  const r = await p.evaluate(async () => {
    const res = await fetch("/api/account", { cache: "no-store" });
    return { status: res.status, cacheControl: res.headers.get("cache-control"), body: (await res.json()) as Record<string, unknown> };
  });
  expect(r.cacheControl).toBe("no-store");
  return { status: r.status, body: r.body };
}

async function accountId(context: BrowserContext, origin = ORIGIN_A): Promise<string> {
  const a = await account(context, origin);
  expect(a.status).toBe(200);
  expect(a.body.id).toMatch(UUID);
  return a.body.id as string;
}

// G016 A2 cookie replay: a fresh browser profile holding only `value` as its __Host-sid for origin, so the browser sends
// exactly that value as the Cookie header to origin's /api/account (the Node request client neither resolves *.localhost
// nor trusts the pinned key). Returns /api/account's answer there; the profile is tracked and closed like any other.
async function replayAt(browser: Browser, testInfo: TestInfo, value: string, origin: string): Promise<Account> {
  const replay = await open(browser, testInfo);
  await replay.context.addCookies([{ name: "__Host-sid", value, url: origin, secure: true, httpOnly: true, sameSite: "Lax" }]);
  return account(replay.context, origin);
}

type Posted = { status: number; redirected: boolean; cacheControl: string | null; text: string };
// A same-origin POST from the browser with this body and Content-Type (the browser adds Origin itself).
async function browserPost(context: BrowserContext, path: string, body: string, contentType: string | null, origin = ORIGIN_A): Promise<Posted> {
  const p = await apiPage(context, origin);
  return p.evaluate(
    async ({ path, body, contentType }) => {
      const headers: Record<string, string> = {};
      if (contentType !== null) headers["content-type"] = contentType;
      const res = await fetch(path, { method: "POST", body, headers, redirect: "manual", cache: "no-store" });
      const redirected = res.type === "opaqueredirect";
      return { status: res.status, redirected, cacheControl: res.headers.get("cache-control"), text: redirected ? "" : await res.text() };
    },
    { path, body, contentType },
  );
}

// The CSRF token rendered into this page's account forms (all three carry the same value). G063: scoped to the Sign-out
// forms, since /account also embeds the MCP connection manager's forms.
async function csrfOf(page: Page): Promise<string> {
  const values = await page.locator('#sign-out input[name="csrf_token"]').evaluateAll((els) => els.map((e) => (e as HTMLInputElement).value));
  expect(values).toHaveLength(3);
  expect(new Set(values).size).toBe(1);
  expect(values[0]).toMatch(SECRET);
  observe(values[0]);
  return values[0];
}

// G016 A2: core answers a session only for the tenant origin it is bound to, so the caller names that origin
// explicitly (no default); it is encoded once, as web's own identity client does.
async function coreSession(d: APIRequestContext, base: string, credential: string, origin: string): Promise<number> {
  const r = await d.get(`${base}/internal/v1/auth/session?origin=${encodeURIComponent(origin)}`, { headers: { Authorization: `Session ${credential}` } });
  return r.status();
}

function signOutButton(page: Page, which: "local" | "all" | "provider") {
  const name = { local: "Sign out", all: "Sign out everywhere", provider: "Sign out and revoke WorkOS session" }[which];
  return page.getByRole("button", { name, exact: true });
}

// --- the emulator's administration (driver; the run's per-run emulator API key) ---------------------------------------
async function emulatorAdmin() {
  const d = await driver();
  const headers = { Authorization: `Bearer ${SEED.emulator_api_key}` };
  return {
    async userId(email: string): Promise<string> {
      const r = await d.get(`${EMULATOR}/user_management/users`, { headers, params: { email } });
      expect(r.status()).toBe(200);
      const data = ((await r.json()) as { data: { id: string }[] }).data;
      expect(data).toHaveLength(1);
      return data[0].id;
    },
    // The user's live provider sessions; the emulator drops a session from this list when it is revoked.
    async sessions(userId: string): Promise<Set<string>> {
      const r = await d.get(`${EMULATOR}/user_management/users/${userId}/sessions`, { headers, params: { limit: 100 } });
      expect(r.status()).toBe(200);
      const body = (await r.json()) as { data: { id: string; status: string }[]; list_metadata: { after: string | null } };
      expect(body.list_metadata.after, "one page of sessions").toBeNull();
      expect(body.data.every((s) => s.status === "active")).toBe(true);
      return new Set(body.data.map((s) => s.id));
    },
    async create(user: Credentials & { first_name: string; last_name: string }): Promise<string> {
      const r = await d.post(`${EMULATOR}/user_management/users`, { headers, data: { ...user, email_verified: true } });
      expect(r.status(), "emulator user creation").toBe(201);
      return ((await r.json()) as { id: string }).id;
    },
    async rename(userId: string, first: string, last: string) {
      const r = await d.put(`${EMULATOR}/user_management/users/${userId}`, { headers, data: { first_name: first, last_name: last } });
      expect(r.status(), "emulator user update").toBe(200);
    },
    async remove(userId: string) {
      const r = await d.delete(`${EMULATOR}/user_management/users/${userId}`, { headers });
      expect(r.status(), "emulator user deletion").toBe(204);
    },
    async failOnce(method: string, path: string, status: number) {
      const r = await d.post(`${EMULATOR}/_emulate/hooks`, { data: { method, path, status, count: 1 } });
      expect(r.status(), "emulator error hook").toBeLessThan(300);
    },
  };
}

// Sessions added to and removed from a user's live provider sessions between two observations.
function changes(before: Set<string>, after: Set<string>): { added: number; removed: number } {
  return { added: [...after].filter((id) => !before.has(id)).length, removed: [...before].filter((id) => !after.has(id)).length };
}

function freshUser(prefix: string): Credentials & { first_name: string; last_name: string } {
  const password = `Pw-${randomBytes(18).toString("hex")}`;
  observe(password);
  return { email: `${prefix}-${randomBytes(6).toString("hex")}@identity.test`, password, first_name: "Attr", last_name: "One" };
}

// =======================================================================================================================

test("[E-LOGIN] [A-1] interactive sign-in on org-a: session and binding cookies, account, no token in the page", async ({ page, context }, testInfo) => {
  // Anonymous /account goes to sign-in, returning to /account.
  await page.goto("/account");
  expect(pathOf(page.url())).toBe("/sign-in");
  expect(new URL(page.url()).searchParams.get("return_to")).toBe("/account");
  expect(await sessionCookie(context)).toBeUndefined();

  const authorize: string[] = [];
  page.on("request", (r) => {
    if (r.url().startsWith(`${EMULATOR}/user_management/authorize?`)) authorize.push(r.url());
  });
  await signIn(page, ALICE);
  expect(pathOf(page.url())).toBe("/account");
  expect(statusOf(page.url())).toBe("signed_in");
  await expect(page.getByTestId("status-message")).toHaveText("You are signed in.");
  await expect(page.getByTestId("display-name")).toHaveText("Alice Example");
  await expect(page.getByTestId("mfa-status")).toHaveText("Not verified by this application");

  // The hosted authorization request carries exactly the reviewed parameters (P-POS-1: the client ID is intended).
  expect(authorize).toHaveLength(1);
  const q = new URL(authorize[0]).searchParams;
  expect([...q.keys()].sort()).toEqual(["client_id", "code_challenge", "code_challenge_method", "provider", "redirect_uri", "response_type", "state"]);
  expect(q.get("redirect_uri")).toBe(`${ORIGIN_A}/auth/callback`);
  expect(q.get("code_challenge_method")).toBe("S256");
  expect(q.get("provider")).toBe("authkit");
  await testInfo.attach("authorize-request.txt", { body: sanitizeUrl(authorize[0]).text, contentType: "text/plain" });

  const sid = await sessionCookie(context);
  expect(sid).toMatchObject({ domain: "org-a.localhost", path: "/", httpOnly: true, secure: true, sameSite: "Lax" });
  expect(sid?.value).toMatch(SECRET);
  const login = await bindingCookie(context);
  expect(login).toMatchObject({ domain: "org-a.localhost", path: "/", httpOnly: true, secure: true, sameSite: "Lax" });
  expect(login!.expires - Date.now() / 1000).toBeLessThanOrEqual(600);

  const a = await account(context);
  expect(a.status).toBe(200);
  expect(Object.keys(a.body).sort()).toEqual(["auth_time", "display_name", "id", "mfa_assurance", "session_expires_at"]);
  expect(a.body).toMatchObject({ display_name: "Alice Example", mfa_assurance: "unknown" });

  // No credential or provider token reaches the URL, storage or HTML.
  const html = await page.content();
  expect(html).not.toContain(sid!.value);
  expect(html).not.toMatch(/eyJ[A-Za-z0-9_-]{10,}\./);
  expect(page.url()).not.toMatch(/[?&](code|state)=/);
  const storage = await page.evaluate(() => JSON.stringify({ l: { ...localStorage }, s: { ...sessionStorage } }));
  expect(storage).not.toContain(sid!.value);
  expect(storage).not.toMatch(/eyJ/);
  // The three Sign-out forms plus G063's embedded MCP create form (this member has no connections, so no row forms).
  const saved = sanitizeHtml(html);
  expect(saved.replaced).toBe(4);
  await testInfo.attach("account.html", { body: saved.text, contentType: "text/html" });
});

test("[E-ORIGINS] the same user on org-b has the same account id; org-a's cookie never reaches org-b", async ({ page, context, browser }, testInfo) => {
  await signIn(page, ALICE);
  const idA = await accountId(context);
  expect(await sessionCookie(context, ORIGIN_B)).toBeUndefined();
  expect((await account(context, ORIGIN_B)).status).toBe(401);
  // G016 A2: org-a's session value replayed as org-b's cookie is refused (bound to org-a); the same replay at org-a works.
  const sidA = await sessionValue(context);
  expect((await replayAt(browser, testInfo, sidA, ORIGIN_B)).status).toBe(401);
  const control = await replayAt(browser, testInfo, sidA, ORIGIN_A);
  expect(control.status).toBe(200);
  expect(control.body.id).toBe(idA);
  await signIn(page, ALICE, ORIGIN_B);
  expect(new URL(page.url()).origin).toBe(ORIGIN_B);
  expect(await accountId(context, ORIGIN_B)).toBe(idA);
  const b = await sessionCookie(context, ORIGIN_B);
  expect(b?.domain).toBe("org-b.localhost");
  expect(b?.value).not.toBe((await sessionCookie(context, ORIGIN_A))?.value);
});

// G016 D-2(a): a tenant-host sign-in completes only for a current member, and a principal without an account cannot
// become one at a tenant host (nothing is created). The rename half uses frank, a seeded member whose name the test
// restores; a new provider user (a fresh address) is refused with not_member and gets no session. This denial is not
// email-reuse evidence: that mapping is by (issuer, subject), never email, stays TestMappingIsByIssuerAndSubject's.
test("[E-ATTR] a renamed user keeps its account; a new provider user without membership gets no account", async ({ page, context, browser }, testInfo) => {
  const admin = await emulatorAdmin();
  const frankId = await admin.userId(FRANK.email);
  await signIn(page, FRANK);
  const id1 = await accountId(context);
  expect((await account(context)).body.display_name).toBe(`${FRANK.first_name} ${FRANK.last_name}`);

  const renamed = `Renamed${randomBytes(3).toString("hex")}`;
  await admin.rename(frankId, renamed, "Person");
  try {
    const again = await open(browser, testInfo);
    await signIn(again.page, FRANK);
    expect(await accountId(again.context)).toBe(id1);
    expect((await account(again.context)).body.display_name).toBe(`${renamed} Person`);
  } finally {
    await admin.rename(frankId, FRANK.first_name, FRANK.last_name);
  }

  const user = freshUser("attr");
  const fresh = await admin.create(user);
  try {
    const reused = await open(browser, testInfo);
    await signIn(reused.page, user);
    expect(pathOf(reused.page.url())).toBe("/sign-in");
    expect(statusOf(reused.page.url())).toBe("not_member");
    await expect(reused.page.getByTestId("status-message")).toContainText("not a member of this workspace");
    expect(await sessionCookie(reused.context)).toBeUndefined();
    expect((await account(reused.context)).status).toBe(401);
  } finally {
    await admin.remove(fresh);
  }
});

test("[E-FIX] a planted session cookie is replaced at sign-in and stays unknown", async ({ page, context }) => {
  const planted = randomBytes(32).toString("base64url");
  observe(planted);
  await context.addCookies([{ name: "__Host-sid", value: planted, url: `${ORIGIN_A}/`, secure: true, httpOnly: true, sameSite: "Lax" }]);
  expect((await account(context)).status).toBe(401);
  await signIn(page, ALICE);
  const now = await sessionValue(context);
  expect(now).not.toBe(planted);
  const d = await driver();
  expect(await coreSession(d, CORE, planted, ORIGIN_A)).toBe(401);
  expect(await coreSession(d, CORE, now, ORIGIN_A)).toBe(200);
});

test("[E-CB-1] a pending sign-in for another user completing in a signed-in browser is refused and revoked", async ({ page, context }) => {
  await startSignIn(page); // pending transaction, to be completed as Carol
  const other = await context.newPage();
  await signIn(other, ALICE);
  const idA = await accountId(context);
  const before = await sessionValue(context);
  const admin = await emulatorAdmin();
  const carolId = await admin.userId(CAROL.email);
  const carolBefore = await admin.sessions(carolId);

  await completeAtEmulator(page, CAROL);
  expect(pathOf(page.url())).toBe("/account");
  expect(statusOf(page.url())).toBe("identity_conflict");
  await expect(page.getByTestId("display-name")).toHaveText("Alice Example");
  expect(await accountId(context)).toBe(idA);
  expect(await sessionValue(context)).toBe(before);
  // The refused sign-in had a validated token, so the provider created a session for Carol; core revoked exactly that
  // one (best effort): none of hers was added and none of her earlier sessions was touched.
  await expect.poll(async () => changes(carolBefore, await admin.sessions(carolId))).toEqual({ added: 0, removed: 0 });
});

test("[E-CB-2] two overlapping starts by one user: the first signs in, the second (a new provider session) rotates", async ({ page, context }) => {
  await startSignIn(page);
  const second = await context.newPage();
  await startSignIn(second);
  await completeAtEmulator(page, ALICE);
  expect(statusOf(page.url())).toBe("signed_in");
  const s1 = await sessionValue(context);
  await completeAtEmulator(second, ALICE);
  expect(statusOf(second.url())).toBe("session_rotated");
  const s2 = await sessionValue(context);
  expect(s2).not.toBe(s1);
  const d = await driver();
  expect(await coreSession(d, CORE, s1, ORIGIN_A)).toBe(401);
  expect(await coreSession(d, CORE, s2, ORIGIN_A)).toBe(200);
});

test("[E-CB-3] a replayed callback URL changes nothing", async ({ page, context }) => {
  await startSignIn(page);
  const callback = await captureCallback(page, ALICE);
  await page.goto(callback);
  expect(statusOf(page.url())).toBe("signed_in");
  const before = await sessionValue(context);
  await page.goto(callback);
  expect(pathOf(page.url())).toBe("/sign-in");
  expect(statusOf(page.url())).toBe("expired");
  expect(await sessionValue(context)).toBe(before);
  expect((await account(context)).status).toBe(200);
});

test("[E-CB-4] a callback opened in another browser is not consumed; the original browser completes", async ({ page, context, browser }, testInfo) => {
  await startSignIn(page);
  const callback = await captureCallback(page, ALICE);
  const other = await open(browser, testInfo);
  await other.page.goto(callback);
  expect(statusOf(other.page.url())).toBe("expired");
  expect(await sessionCookie(other.context)).toBeUndefined();
  await page.goto(callback);
  expect(statusOf(page.url())).toBe("signed_in");
  expect((await account(context)).status).toBe(200);
});

test("[E-CB-5] a wrong-binding callback leaves both transactions intact", async ({ page, browser }, testInfo) => {
  await startSignIn(page);
  const callback = await captureCallback(page, ALICE);
  const other = await open(browser, testInfo);
  await startSignIn(other.page); // the other browser's own pending transaction
  const probe = await other.context.newPage();
  await probe.goto(callback);
  expect(statusOf(probe.url())).toBe("expired");
  await page.goto(callback);
  expect(statusOf(page.url())).toBe("signed_in");
  await completeAtEmulator(other.page, BOB);
  expect(statusOf(other.page.url())).toBe("signed_in");
  expect((await account(other.context)).body.display_name).toBe("Bob Example");
});

test("[E-CB-6] login CSRF: an attacker's callback fed to a victim signs nobody in", async ({ page, context, browser }, testInfo) => {
  const attacker = await open(browser, testInfo);
  await startSignIn(attacker.page);
  const callback = await captureCallback(attacker.page, BOB);
  await page.goto(callback);
  expect(statusOf(page.url())).toBe("expired");
  expect((await account(context)).status).toBe(401);
  // Not consumed by the victim: it still completes in the attacker's own browser.
  await attacker.page.goto(callback);
  expect(statusOf(attacker.page.url())).toBe("signed_in");
});

test("[E-CB-7] [A-7] a canceled sign-in consumes its transaction and issues no session", async ({ page, context }) => {
  await startSignIn(page);
  const state = new URL(page.url()).searchParams.get("state");
  expect(state).toMatch(SECRET);
  observe(state);
  const description = canary();
  await page.goto(`${ORIGIN_A}/auth/callback?error=access_denied&error_description=${description}&state=${state}`);
  expect(pathOf(page.url())).toBe("/sign-in");
  expect(statusOf(page.url())).toBe("canceled");
  await expect(page.getByTestId("status-message")).toHaveText("Sign-in was canceled. Nothing was changed.");
  expect(await page.content()).not.toContain(description);
  expect(await sessionCookie(context)).toBeUndefined();
  await page.goto(`${ORIGIN_A}/auth/callback?code=${canary()}&state=${state}`);
  expect(statusOf(page.url())).toBe("expired");
  expect((await account(context)).status).toBe(401);
});

test("[E-CB-8] hostile return paths and forged Host or forwarding headers never change the destination", async ({ page, context }) => {
  await context.setExtraHTTPHeaders({ "X-Forwarded-Host": "evil.example", "X-Forwarded-Proto": "http", Forwarded: "host=evil.example;proto=http" });
  await page.goto(`${ORIGIN_A}/sign-in?return_to=https%3A%2F%2Fevil.example%2F`);
  expect(await page.locator('input[name="return_to"]').count()).toBe(0);
  // A crafted form with a hostile return path: the authorize redirect and the final landing stay allow-listed.
  const authorize: string[] = [];
  page.on("request", (r) => {
    if (r.url().startsWith(`${EMULATOR}/user_management/authorize?`)) authorize.push(r.url());
  });
  await page.evaluate(() => {
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = "return_to";
    input.value = "//evil.example/account";
    document.querySelector('form[action="/auth/login"]')!.appendChild(input);
  });
  await page.getByRole("button", { name: "Sign in" }).click();
  await completeAtEmulator(page, ALICE);
  expect(new URL(authorize[0]).searchParams.get("redirect_uri")).toBe(`${ORIGIN_A}/auth/callback`);
  expect(new URL(page.url()).origin).toBe(ORIGIN_A);
  expect(pathOf(page.url())).toBe("/");
  // Forged Host from a driver straight to web: the derived origin is refused and nothing reaches an endpoint.
  const d = await driver();
  const forged = await d.post(`${WEB_DIRECT}/auth/login`, {
    headers: { Host: "evil.example", Origin: "https://evil.example", "Content-Type": FORM },
    data: "return_to=%2Faccount",
    maxRedirects: 0,
  });
  expect(forged.status()).toBe(403);
  const cb = await d.get(`${WEB_DIRECT}/auth/callback?code=x&state=y`, { headers: { Host: "evil.example" }, maxRedirects: 0 });
  expect(cb.status()).toBe(303);
  expect(cb.headers()["location"]).toBe("/sign-in?status=failed");
});

test("[E-CB-10] a new provider session for the signed-in user rotates the session; old forms refuse until reload", async ({ page, context }) => {
  const pending = await context.newPage();
  await startSignIn(pending);
  await signIn(page, ALICE);
  const s1 = await sessionValue(context);
  const tab = await context.newPage();
  await tab.goto("/account");
  const t1 = await csrfOf(tab);
  const admin = await emulatorAdmin();
  const aliceId = await admin.userId(ALICE.email);
  const aliceBefore = await admin.sessions(aliceId);

  await completeAtEmulator(pending, ALICE);
  expect(statusOf(pending.url())).toBe("session_rotated");
  const s2 = await sessionValue(context);
  expect(s2).not.toBe(s1);
  const d = await driver();
  expect(await coreSession(d, CORE, s1, ORIGIN_A)).toBe(401);
  expect(await coreSession(d, CORE, s2, ORIGIN_A)).toBe(200);
  // The other tab's form still carries the old session's token: refused, nothing changes, until it reloads.
  await signOutButton(tab, "all").click();
  await expect(tab.getByRole("heading", { name: "Form expired" })).toBeVisible();
  expect((await account(context)).status).toBe(200);
  await tab.goto("/account");
  expect(await csrfOf(tab)).not.toBe(t1);
  // Rotation calls no provider revocation: the new provider session is added and none is removed.
  expect(changes(aliceBefore, await admin.sessions(aliceId))).toEqual({ added: 1, removed: 0 });
});

test("[E-PKCE] the victim's code injected into the attacker's transaction is rejected; nobody signs in", async ({ page, context, browser }, testInfo) => {
  const attacker = await open(browser, testInfo);
  await startSignIn(attacker.page);
  const attackerState = new URL(attacker.page.url()).searchParams.get("state")!;
  observe(attackerState);
  await startSignIn(page);
  const victimCallback = await captureCallback(page, ALICE);
  const victimCode = new URL(victimCallback).searchParams.get("code")!;
  await attacker.page.goto(`${ORIGIN_A}/auth/callback?code=${victimCode}&state=${attackerState}`);
  expect(statusOf(attacker.page.url())).toBe("failed");
  expect((await account(attacker.context)).status).toBe(401);
  expect((await account(context)).status).toBe(401);
  // The attacker's transaction was consumed by the rejected exchange.
  await attacker.page.goto(`${ORIGIN_A}/auth/callback?code=${victimCode}&state=${attackerState}`);
  expect(statusOf(attacker.page.url())).toBe("expired");
});

test("[E-CT] web form routes accept only form posts: JSON, text and multipart are refused unchanged", async ({ page, context }) => {
  await signIn(page, ALICE);
  await page.goto("/account");
  const token = await csrfOf(page);
  const boundary = "----identitytest";
  const multipart = `--${boundary}\r\nContent-Disposition: form-data; name="csrf_token"\r\n\r\n${token}\r\n--${boundary}--\r\n`;
  for (const path of ["/auth/logout", "/auth/logout-all", "/auth/provider-logout", "/auth/login"]) {
    for (const [body, type] of [
      [JSON.stringify({ csrf_token: token }), "application/json"],
      [`csrf_token=${token}`, "text/plain"],
      [multipart, `multipart/form-data; boundary=${boundary}`],
    ] as const) {
      const r = await browserPost(context, path, body, type);
      expect(r.status, `${path} ${type}`).toBe(415);
      expect(r.cacheControl).toBe("no-store");
    }
  }
  expect((await account(context)).status).toBe(200);
  const get = await context.newPage();
  const response = await get.goto("/auth/logout");
  expect(response?.status()).toBe(405);
  expect(response?.headers()["allow"]).toBe("POST");
  expect((await account(context)).status).toBe(200);
});

test("[E-CSRF] the rendered token works across reloads and tabs; missing, wrong, other-user, text and cross-site posts are refused", async ({ page, context, browser }, testInfo) => {
  await signIn(page, ALICE);
  await page.reload();
  const token = await csrfOf(page);
  const bob = await open(browser, testInfo);
  await signIn(bob.page, BOB);
  await bob.page.goto("/account");
  const bobToken = await csrfOf(bob.page);

  for (const [label, body] of [
    ["missing", ""],
    ["empty", "csrf_token="],
    ["mismatched", `csrf_token=${randomBytes(32).toString("base64url")}`],
    ["other user's", `csrf_token=${bobToken}`],
    ["repeated", `csrf_token=${token}&csrf_token=${token}`],
  ]) {
    const r = await browserPost(context, "/auth/logout-all", body, FORM);
    expect(r.status, label).toBe(label === "repeated" ? 400 : 403);
  }
  expect((await browserPost(context, "/auth/logout-all", `csrf_token=${token}`, "text/plain")).status).toBe(415);
  expect((await account(context)).status).toBe(200);
  expect((await account(bob.context)).status).toBe(200);

  // A cross-site page posting the real token is refused by origin (SameSite=Lax would still send the cookie).
  tracking.allowed.add(ATTACKER);
  const victim = await context.newPage();
  for (const attack of [`logout-all:token=${token}`, `logout:token=${token}`, "login", `provider-logout:text:token=${token}`]) {
    await victim.goto(`${ATTACKER}/#${encodeURIComponent(attack)}`);
    await victim.waitForURL((u) => u.origin === ORIGIN_A);
    await expect(victim.locator("h1")).toHaveText(attack.includes(":text") ? "Unsupported request" : "Request refused");
  }
  expect((await account(context)).status).toBe(200);

  // Two tabs: one signs out everywhere, the other's sign-out is then idempotent.
  const tab = await context.newPage();
  await tab.goto("/account");
  expect(await csrfOf(tab)).toBe(token);
  await signOutButton(tab, "all").click();
  await tab.waitForURL((u) => statusOf(u.href) === "signed_out_all");
  await signOutButton(page, "local").click();
  await page.waitForURL((u) => statusOf(u.href) === "signed_out");
  expect((await account(context)).status).toBe(401);
});

test("[E-LO-1] a repeated sign-out in another tab is idempotent and clears the cookie", async ({ page, context }) => {
  await signIn(page, ALICE);
  const tab = await context.newPage();
  await tab.goto("/account");
  await page.goto("/account");
  await signOutButton(page, "local").click();
  await page.waitForURL((u) => statusOf(u.href) === "signed_out");
  await expect(page.getByTestId("status-message")).toHaveText("You have signed out of this application.");
  expect(await sessionCookie(context)).toBeUndefined();
  await signOutButton(tab, "local").click();
  await tab.waitForURL((u) => statusOf(u.href) === "signed_out");
  expect(await sessionCookie(context)).toBeUndefined();
});

test("[E-LO-2] signing out everywhere without a live session is refused with 401", async ({ page, context }) => {
  await signIn(page, ALICE);
  const tab = await context.newPage();
  await tab.goto("/account");
  await page.goto("/account");
  await signOutButton(page, "local").click();
  await page.waitForURL((u) => statusOf(u.href) === "signed_out");
  const response = tab.waitForResponse((r) => r.url().endsWith("/auth/logout-all"));
  await signOutButton(tab, "all").click();
  expect((await response).status()).toBe(401);
  await expect(tab.getByRole("heading", { name: "Not signed in" })).toBeVisible();
});

test("[E-LO-3] a copied credential is dead after sign-out through web, core and the replica", async ({ page, context }) => {
  await signIn(page, ALICE);
  const value = await sessionValue(context);
  const d = await driver();
  expect(await coreSession(d, CORE, value, ORIGIN_A)).toBe(200);
  expect(await coreSession(d, CORE_REPLICA, value, ORIGIN_A)).toBe(200);
  await page.goto("/account");
  await signOutButton(page, "local").click();
  await page.waitForURL((u) => statusOf(u.href) === "signed_out");
  expect(await coreSession(d, CORE, value, ORIGIN_A)).toBe(401);
  expect(await coreSession(d, CORE_REPLICA, value, ORIGIN_A)).toBe(401);
  const viaWeb = await d.get(`${WEB_DIRECT}/api/account`, { headers: { Host: "org-a.localhost:543", Cookie: `__Host-sid=${value}` } });
  expect(viaWeb.status()).toBe(401);
});

test("[E-LO-4] signing out everywhere ends every session of that user only", async ({ page, context, browser }, testInfo) => {
  await signIn(page, ALICE);
  const second = await open(browser, testInfo);
  await signIn(second.page, ALICE);
  const bob = await open(browser, testInfo);
  await signIn(bob.page, BOB);
  await page.goto("/account");
  await signOutButton(page, "all").click();
  await page.waitForURL((u) => statusOf(u.href) === "signed_out_all");
  expect((await account(context)).status).toBe(401);
  expect((await account(second.context)).status).toBe(401);
  expect((await account(bob.context)).status).toBe(200);
});

// The logout reference cookie on ORIGIN, if any (its value is registered as an observed secret).
async function logoutCookie(context: BrowserContext, origin = ORIGIN_A): Promise<Cookie | undefined> {
  const cookie = (await context.cookies(origin)).find((c) => c.name === "__Host-logout");
  observe(cookie?.value);
  return cookie;
}

// Captures the logout references issued by POST /auth/provider-logout (from its Set-Cookie) and every request URL of the
// flow, so a test can prove the reference never travels in a URL (C7 addendum §2).
function watchProviderLogout(context: BrowserContext) {
  const refs: string[] = [];
  const urls: string[] = [];
  const hops: string[] = [];
  context.on("request", (r) => urls.push(r.url()));
  context.on("response", async (r) => {
    const u = new URL(r.url());
    if ([ORIGIN_A, ORIGIN_B].includes(u.origin)) hops.push(`${u.origin}${u.pathname} ${r.status()} ${r.headers()["location"] ?? "-"}`);
    if (u.pathname === "/auth/provider-logout" && r.request().method() === "POST") {
      for (const value of await r.headerValues("set-cookie")) {
        const m = /^__Host-logout=([A-Za-z0-9_-]{43}); Path=\/; Max-Age=600; HttpOnly; SameSite=Lax; Secure$/.exec(value);
        if (m) {
          observe(m[1]);
          refs.push(m[1]);
        }
      }
    }
  });
  return { refs, urls, hops };
}

// Plants a logout cookie. Only a canonical reference can be a secret, so only such a value is registered; a deliberately
// malformed test literal is not (it also appears in the repository's tests).
async function putLogoutCookie(context: BrowserContext, value: string, origin = ORIGIN_A) {
  if (SECRET.test(value)) observe(value);
  await context.addCookies([{ name: "__Host-logout", value, url: origin, secure: true, httpOnly: true, sameSite: "Lax" }]);
}

function message(page: Page) {
  return page.locator("main p").first();
}

test("[E-LO-6] [E-LO-7] a failed provider revocation never redirects to the provider; the logout reference is one-use and never claims success", async ({ page, context, browser }, testInfo) => {
  await signIn(page, ALICE);
  const admin = await emulatorAdmin();
  await admin.failOnce("POST", "/user_management/sessions/revoke", 500);
  const flow = watchProviderLogout(context);
  await page.goto("/account");
  await signOutButton(page, "provider").click();
  await page.waitForURL((u) => u.origin === ORIGIN_A && pathOf(u.href) === "/auth/logout-return");
  await expect(message(page)).toHaveText("Signed out of this application. The WorkOS session for this sign-in could not be revoked or confirmed.");
  expect(flow.urls.filter((u) => u.startsWith(`${EMULATOR}/user_management/sessions/logout`))).toEqual([]);
  expect(flow.refs).toHaveLength(1);
  expect(flow.urls.filter((u) => u.includes(flow.refs[0]))).toEqual([]);
  expect(await page.content()).not.toContain(flow.refs[0]);
  expect(await logoutCookie(context)).toBeUndefined();
  expect((await account(context)).status).toBe(401);
  // E-LO-7: a spent, a forged and a malformed reference never claim a provider outcome; the sign-in page hands over
  // with a fixed 307 and every visit clears the cookie.
  for (const [value, text] of [
    [flow.refs[0], "Signed out of this application."],
    [randomBytes(32).toString("base64url"), "Signed out of this application."],
    ["not-a-reference", "There is no sign-out outcome to show here."],
  ]) {
    await putLogoutCookie(context, value);
    const hops = flow.hops.length;
    await page.goto("/sign-in");
    expect(pathOf(page.url())).toBe("/auth/logout-return");
    expect(flow.hops.slice(hops)).toContain(`${ORIGIN_A}/sign-in 307 /auth/logout-return`);
    await expect(message(page)).toHaveText(text);
    expect(await logoutCookie(context)).toBeUndefined();
  }
  // Without the cookie the sign-in page stays itself, and the return route has nothing to report.
  await page.goto("/sign-in");
  expect(pathOf(page.url())).toBe("/sign-in");
  await page.goto("/auth/logout-return");
  await expect(message(page)).toHaveText("There is no sign-out outcome to show here.");
  // The return route only reads notification state: a live session survives a visit with a reference.
  const bob = await open(browser, testInfo);
  await signIn(bob.page, BOB);
  await putLogoutCookie(bob.context, randomBytes(32).toString("base64url"));
  await bob.page.goto("/auth/logout-return");
  expect((await account(bob.context)).status).toBe(200);
  // Duplicates (a raw header), other methods and caching, through web's own port.
  const d = await driver();
  const dup = await d.get(`${WEB_DIRECT}/auth/logout-return`, {
    headers: { Host: "org-a.localhost:543", Cookie: `__Host-logout=${flow.refs[0]}; __Host-logout=${flow.refs[0]}` },
    maxRedirects: 0,
  });
  expect(dup.status()).toBe(200);
  expect(await dup.text()).toContain("There is no sign-out outcome to show here.");
  expect(dup.headers()["cache-control"]).toBe("no-store");
  expect(dup.headersArray().filter((h) => h.name.toLowerCase() === "set-cookie").map((h) => h.value)).toEqual([
    "__Host-logout=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure",
  ]);
  for (const method of ["POST", "PUT", "PATCH", "DELETE"]) {
    const r = await d.fetch(`${WEB_DIRECT}/auth/logout-return`, { method, headers: { Host: "org-a.localhost:543" }, maxRedirects: 0 });
    expect(r.status(), method).toBe(405);
    expect(r.headers()["allow"]).toBe("GET");
  }
});

test("[E-LO-8] a confirmed provider revocation continues to the provider's logout and returns once, without a query, on both origins", async ({ page, context }) => {
  const admin = await emulatorAdmin();
  const aliceId = await admin.userId(ALICE.email);
  for (const origin of [ORIGIN_A, ORIGIN_B]) {
    await signIn(page, ALICE, origin);
    const before = await admin.sessions(aliceId);
    const flow = watchProviderLogout(context);
    await page.goto(`${origin}/account`);
    await signOutButton(page, "provider").click();
    await page.waitForURL((u) => u.origin === origin && pathOf(u.href) === "/auth/logout-return");
    const logoutRequests = flow.urls.filter((u) => u.startsWith(`${EMULATOR}/user_management/sessions/logout?`));
    expect(logoutRequests).toHaveLength(1);
    // The provider returns to exactly this origin's registered sign-in page, which hands over to the return route.
    expect(new URL(logoutRequests[0]).searchParams.get("return_to")).toBe(`${origin}/sign-in`);
    expect(flow.hops).toContain(`${origin}/sign-in 307 /auth/logout-return`);
    expect(flow.refs).toHaveLength(1);
    expect(flow.urls.filter((u) => u.includes(flow.refs[0]))).toEqual([]);
    expect(await page.content()).not.toContain(flow.refs[0]);
    await expect(message(page)).toHaveText("Signed out of this application. The WorkOS session for this sign-in was revoked.");
    expect(await logoutCookie(context, origin)).toBeUndefined();
    expect((await account(context, origin)).status).toBe(401);
    // Exactly this sign-in's provider session was revoked.
    expect(changes(before, await admin.sessions(aliceId))).toEqual({ added: 0, removed: 1 });
    // The outcome is shown once: the sign-in page is plain again.
    await page.goto(`${origin}/sign-in`);
    expect(pathOf(page.url())).toBe("/sign-in");
  }
});

test("[E-CACHE] alternating users and browsers only ever see their own account; identity responses are no-store", async ({ page, context, browser }, testInfo) => {
  const cacheControl: string[] = [];
  const watch = (c: BrowserContext) =>
    c.on("response", (r) => {
      const u = new URL(r.url());
      if ([ORIGIN_A, ORIGIN_B].includes(u.origin) && /^\/(sign-in|account|api\/account|auth\/)/.test(u.pathname)) {
        cacheControl.push(`${u.pathname} ${r.headers()["cache-control"] ?? "(none)"}`);
      }
    });
  watch(context);
  await signIn(page, ALICE);
  const bob = await open(browser, testInfo);
  watch(bob.context);
  await signIn(bob.page, BOB);
  for (let i = 0; i < 3; i++) {
    await page.goto("/account");
    await expect(page.getByTestId("display-name")).toHaveText("Alice Example");
    await bob.page.goto("/account");
    await expect(bob.page.getByTestId("display-name")).toHaveText("Bob Example");
    expect((await account(context)).body.display_name).toBe("Alice Example");
    expect((await account(bob.context)).body.display_name).toBe("Bob Example");
  }
  expect(cacheControl.length).toBeGreaterThan(10);
  expect(cacheControl.filter((line) => !/\bno-store\b/.test(line))).toEqual([]);
});

test("[E-A11Y] sign-in, pending, canceled, error, signed-out and account states: keyboard and axe in both widths", async ({ page, context }) => {
  const tags = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];
  const audit = async (p: Page, label: string) => {
    const results = await new AxeBuilder({ page: p }).withTags(tags).analyze();
    // G069: only the kit's white label on its #0075ff primary (where no tenant palette applies) is an OWNER-WAIVED
    // color-contrast node, recorded with its ratio; every other violation fails (support/contrast.ts expectAxeClean).
    expectAxeClean(results.violations, `axe (${label})`);
  };
  for (const width of [1280, 320]) {
    await page.setViewportSize({ width, height: width === 320 ? 640 : 800 });
    await page.goto("/sign-in?status=canceled");
    await audit(page, `${width} canceled`);
    await page.goto("/sign-in?status=signed_out");
    await audit(page, `${width} signed-out`);
    const width320 = await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth);
    expect(width320, `no horizontal scroll at ${width}`).toBe(true);
  }
  await page.setViewportSize({ width: 1280, height: 800 });
  // Keyboard: the sign-in button is reachable and submits. For the pending state, a later submit listener cancels
  // the navigation (the button's own listener runs first), so the busy state and the second-submit block stay visible.
  await page.goto("/sign-in");
  await audit(page, "sign-in");
  const button = page.getByRole("button", { name: "Sign in" });
  for (let i = 0; i < 20 && !(await button.evaluate((b) => b === document.activeElement)); i++) await page.keyboard.press("Tab");
  await expect(button).toBeFocused();
  await page.evaluate(() => {
    const w = window as unknown as { blocked: boolean[] };
    w.blocked = [];
    document.querySelector('form[action="/auth/login"]')!.addEventListener("submit", (e) => {
      w.blocked.push(e.defaultPrevented); // whether the button's own listener already blocked this submission
      e.preventDefault();
    });
  });
  await page.keyboard.press("Enter");
  const pending = page.getByRole("button", { name: "Signing in…" });
  await expect(pending).toHaveAttribute("aria-disabled", "true");
  await expect(pending).toBeFocused();
  await expect(page.getByRole("status").filter({ hasText: "Signing in…" })).toHaveCount(1);
  await audit(page, "pending");
  await page.keyboard.press("Enter");
  // The first submission went ahead; the button itself blocked the second.
  await expect.poll(() => page.evaluate(() => (window as unknown as { blocked: boolean[] }).blocked)).toEqual([false, true]);
  await expect(pending).toHaveAttribute("aria-disabled", "true");
  // Without the listener, the keyboard submission signs in.
  await page.goto("/sign-in");
  await page.getByRole("button", { name: "Sign in" }).focus();
  await page.keyboard.press("Enter");
  await completeAtEmulator(page, ALICE);
  await audit(page, "account");
  // An error page (refused form) is a complete, accessible document too.
  const error = await context.newPage();
  await error.goto(`${ORIGIN_A}/auth/logout`);
  await audit(error, "405 page");
});

test("[A-3] forged assurance labels in the query and request headers change nothing", async ({ page, context }) => {
  await context.setExtraHTTPHeaders({ "X-MFA-Verified": "true", "X-User-Id": "00000000-0000-0000-0000-000000000000", "X-Role": "admin" });
  expect((await account(context)).status).toBe(401);
  await signIn(page, ALICE);
  await page.goto("/account?mfa=verified&amr=mfa&status=signed_in");
  await expect(page.getByTestId("mfa-status")).toHaveText("Not verified by this application");
  expect((await account(context)).body.mfa_assurance).toBe("unknown");
  const d = await driver();
  const forged = await d.get(`${CORE}/internal/v1/auth/session`, { headers: { "X-User-Id": "00000000-0000-0000-0000-000000000000", "X-MFA-Verified": "true" } });
  expect(forged.status()).toBe(401);
});

test("[E-NET-2] extra pages and popups are captured too; an injected page request to core is detected", async ({ page, context }) => {
  await signIn(page, ALICE);
  const extra = await context.newPage();
  await extra.goto("/account");
  const popupPromise = context.waitForEvent("page");
  await page.evaluate(() => window.open("/sign-in", "_blank"));
  const popup = await popupPromise;
  await popup.waitForLoadState();
  expect(pathOf(popup.url())).toBe("/sign-in");
  expect([...tracking.origins]).toContain(ORIGIN_A);
  // Negative control: a page navigating to core is captured, and the allow-list would fail it.
  const probe = await context.newPage();
  await probe.goto(`${CORE}/healthz`).catch(() => undefined);
  expect([...tracking.origins]).toContain(CORE);
  expect([...tracking.origins].filter((o) => !tracking.allowed.has(o))).toEqual([CORE]);
  tracking.origins.delete(CORE); // the control is spent here; the after-test assertion stays strict
});

test("[E-PRIVACY] a hostile display name renders as text for its owner only; the email never reaches the app", async ({ page, context, browser }, testInfo) => {
  await signIn(page, HOSTILE);
  await page.goto("/account");
  const expected = `${HOSTILE.first_name} ${HOSTILE.last_name}`;
  await expect(page.getByTestId("display-name")).toHaveText(expected);
  expect(await page.locator("img[src='x']").count()).toBe(0);
  expect(await page.title()).not.toContain(SEED.canaries.script);
  const html = await page.content();
  expect(html).not.toContain(SEED.canaries.email);
  expect((await account(context)).body.display_name).toBe(expected);
  const alice = await open(browser, testInfo);
  await signIn(alice.page, ALICE);
  await alice.page.goto("/account");
  expect(await alice.page.content()).not.toContain(SEED.canaries.display);
  expect(JSON.stringify((await account(alice.context)).body)).not.toContain(SEED.canaries.display);
});

// Hands the harness one live and one revoked credential for its restart, outage and forged-header probes
// (B-IDENTITY-PROBES, E-LO-5). The file sits beside the report (readable by the harness's user; protected by the
// report's 0700 parent); the harness moves it into its 0700 probe inputs, registers the values as secrets and deletes
// it after use. The live session is Carol's: no later test (in either project) signs her out everywhere or disables
// her, so it is still live when the probes run.
test("[E-HANDOFF] a live and a revoked session for the harness probes", async ({ page, context }, testInfo) => {
  await signIn(page, BOB);
  const revokedValue = await sessionValue(context);
  await page.goto("/account");
  await signOutButton(page, "local").click();
  await page.waitForURL((u) => statusOf(u.href) === "signed_out");
  await signIn(page, CAROL);
  const live = await sessionValue(context);
  await page.goto("/account");
  const csrf = await csrfOf(page);
  const report = dirname(process.env.PLAYWRIGHT_JSON_OUTPUT_FILE ?? "/tmp/playwright/results.json");
  const file = `${report}/handoff-${testInfo.project.name}.json`;
  writeFileSync(file, JSON.stringify({ origin: ORIGIN_A, cookie: "__Host-sid", live, csrf, revoked: revokedValue }) + "\n", { mode: 0o644 });
});

// =======================================================================================================================
// G016 hostname resolution (Plan r4 §1, §1.7; E-DOM-1..5) and the G057 opaque platform address (Plan r5 §3.4, §5.2 T8;
// E-DOM-6), in the domain-platform and domain-custom projects only. The harness seeds Organization A (platform label
// org-a, friendly slug alpha-agency, the routing custom host portal.acme.test and the removed host old.acme.test), B
// (label org-b, slug bravo-agency) and C (label org-c, slug charlie-agency) with every sign-in user a member
// (infra/identity/mcp-seed.sql; G057 D-2: org-a/b/c are fixed synthetic test labels, the slugs are metadata and route
// nowhere). The edge
// passes every *.localhost and *.acme.test name to web; web's host gate and core decide what each one is. Chromium maps
// *.acme.test to the edge (playwright.config.ts). localhost and www.localhost are the public entry: G062 (Plan r5 §1.1,
// SUP-16) replaced the marketing redirect with the root landing, whose only action is Sign Up to the platform root's
// /signup (www canonicalizes to the root); the root signup flow itself is signup.spec.ts.

const ORIGIN_C = "https://org-c.localhost:543";
const PORTAL = "https://portal.acme.test:543";
const AUTH = "https://auth.localhost:543";
const ROOT = "https://localhost:543";
// OS002: the default product mark and the favicon (byte-identical files; infra/validation/lib.sh WEB_PUBLIC_PINS).
const PRODUCT_ASSETS = ["/brand/mark.png", "/icon.png"];
const PRODUCT_MARK_SHA256 = "8b8238b26a7974148b2ef47db42c4c5200537bfa583de5426ff8acbe9aebcacd";

// An extra page of the context answering an error status (the default page's S-ERRORS check stays on normal pages).
async function statusAt(context: BrowserContext, url: string): Promise<{ status: number; text: string }> {
  const p = await context.newPage();
  const r = await p.goto(url);
  const out = { status: r?.status() ?? 0, text: (await p.locator("body").textContent()) ?? "" };
  await p.close();
  return out;
}

// A sign-in start on HOST as the browser's form post arrives at web through the edge (the Node client cannot resolve
// the test names): the status and Location of web's answer.
async function loginStart(host: string): Promise<{ status: number; location: string }> {
  const d = await driver();
  const r = await d.fetch(`${WEB_DIRECT}/auth/login`, {
    method: "POST",
    headers: { Host: host, Origin: `https://${host}`, "Content-Type": FORM },
    data: "return_to=%2Faccount",
    maxRedirects: 0,
  });
  return { status: r.status(), location: r.headers()["location"] ?? "" };
}

test("[E-DOM-1] a third platform host signs in with its own callback and host-only cookie; the same account as on org-a @domain-platform", async ({ page, context }) => {
  tracking.allowed.add(ORIGIN_C);
  const authorize: string[] = [];
  page.on("request", (r) => {
    if (r.url().startsWith(`${EMULATOR}/user_management/authorize?`)) authorize.push(r.url());
  });
  await signIn(page, ALICE, ORIGIN_C);
  expect(new URL(page.url()).origin).toBe(ORIGIN_C);
  expect(pathOf(page.url())).toBe("/account");
  expect(statusOf(page.url())).toBe("signed_in");
  expect(authorize).toHaveLength(1);
  expect(new URL(authorize[0]).searchParams.get("redirect_uri")).toBe(`${ORIGIN_C}/auth/callback`);
  const sid = await sessionCookie(context, ORIGIN_C);
  expect(sid).toMatchObject({ domain: "org-c.localhost", path: "/", httpOnly: true, secure: true, sameSite: "Lax" });
  expect(await sessionCookie(context, ORIGIN_A)).toBeUndefined();
  const idC = await accountId(context, ORIGIN_C);
  await signIn(page, ALICE, ORIGIN_A);
  expect(await accountId(context, ORIGIN_A)).toBe(idC);
  expect((await sessionCookie(context, ORIGIN_A))?.value).not.toBe(sid?.value);
});

test("[E-DOM-2] a custom host signs in through the shared callback relay; host-only cookie, same account, local sign-out @domain-custom", async ({ page, context, browser }, testInfo) => {
  for (const o of [PORTAL, AUTH, ORIGIN_A]) tracking.allowed.add(o);
  const authorize: string[] = [];
  const hops: string[] = [];
  page.on("request", (r) => {
    if (r.url().startsWith(`${EMULATOR}/user_management/authorize?`)) authorize.push(r.url());
  });
  page.on("response", (r) => {
    const u = new URL(r.url());
    if ([AUTH, PORTAL].includes(u.origin) && u.pathname === "/auth/callback") {
      hops.push(`${u.origin} ${r.status()} ${(r.headers()["location"] ?? "-").split("?")[0]}`);
    }
  });
  await signIn(page, ALICE, PORTAL);
  expect(new URL(page.url()).origin).toBe(PORTAL);
  expect(pathOf(page.url())).toBe("/account");
  expect(statusOf(page.url())).toBe("signed_in");
  // The provider returns to the shared auth host, which relays the unchanged query to the custom host's callback.
  expect(authorize).toHaveLength(1);
  expect(new URL(authorize[0]).searchParams.get("redirect_uri")).toBe(`${AUTH}/auth/callback`);
  expect(hops).toEqual([`${AUTH} 303 ${PORTAL}/auth/callback`, `${PORTAL} 303 /account`]);
  const sid = await sessionCookie(context, PORTAL);
  expect(sid).toMatchObject({ domain: "portal.acme.test", path: "/", httpOnly: true, secure: true, sameSite: "Lax" });
  expect(sid?.value).toMatch(SECRET);
  expect((await context.cookies(AUTH)).map((c) => c.name)).toEqual([]);
  const idPortal = await accountId(context, PORTAL);
  // G016 A2: the portal session value replayed as org-a's cookie (same Organization A, another origin) is refused; the
  // same replay at the portal works.
  expect((await replayAt(browser, testInfo, sid!.value, ORIGIN_A)).status).toBe(401);
  const control = await replayAt(browser, testInfo, sid!.value, PORTAL);
  expect(control.status).toBe(200);
  expect(control.body.id).toBe(idPortal);
  await signIn(page, ALICE, ORIGIN_A);
  expect(await accountId(context, ORIGIN_A)).toBe(idPortal);
  // Local sign-out on the custom host ends that session only.
  await page.goto(`${PORTAL}/account`);
  await signOutButton(page, "local").click();
  await page.waitForURL((u) => u.origin === PORTAL && statusOf(u.href) === "signed_out");
  expect(await sessionCookie(context, PORTAL)).toBeUndefined();
  expect((await account(context, PORTAL)).status).toBe(401);
  expect(await accountId(context, ORIGIN_A)).toBe(idPortal);
  // The auth host serves only its fixed sign-in page; any other page there is the fixed 404.
  expect((await statusAt(context, `${AUTH}/account`)).status).toBe(404);
});

test("[E-DOM-3] an unknown platform-shaped or custom host is the fixed 404 and cannot start a sign-in @domain-platform", async ({ context }) => {
  const d = await driver();
  for (const origin of ["https://org-z.localhost:543", "https://unknown.acme.test:543"]) {
    tracking.allowed.add(origin);
    for (const path of ["/", "/sign-in", "/account", "/connections/mcp"]) {
      const r = await statusAt(context, `${origin}${path}`);
      expect(r.status, `${origin}${path}`).toBe(404);
      expect(r.text).toContain("There is nothing at this address.");
    }
    expect(await loginStart(new URL(origin).host)).toEqual({ status: 403, location: "" });
    // OS002: the default product mark and the favicon are no exception on an unknown host.
    for (const path of PRODUCT_ASSETS) {
      const asset = await d.get(`${WEB_DIRECT}${path}`, { headers: { Host: new URL(origin).host }, maxRedirects: 0 });
      expect(asset.status(), `${origin}${path}`).toBe(404);
    }
  }
});

test("[E-DOM-4] the platform domain and www serve the landing with one Sign Up to the root, no marketing hop; protocol paths stay exempt @domain-platform", async ({ page, context }) => {
  for (const origin of [ROOT, "https://www.localhost:543"]) {
    tracking.allowed.add(origin);
    const p = await context.newPage();
    const r = await p.goto(`${origin}/`);
    // Served in place: no redirect hop and no other origin.
    expect(r?.status()).toBe(200);
    expect(r?.request().redirectedFrom()).toBeNull();
    expect(p.url()).toBe(`${origin}/`);
    // The only authentication or acquisition action is Sign Up, an absolute link to the root's /signup.
    await expect(p.getByTestId("root-sign-up")).toHaveCount(1);
    expect(await p.getByTestId("root-sign-up").getAttribute("href")).toBe(`${ROOT}/signup`);
    await expect(p.getByRole("link", { name: /sign in|account/i })).toHaveCount(0);
    await expect(p.getByRole("button", { name: /sign in/i })).toHaveCount(0);
    await p.close();
    // Every other page of the public entry is the fixed 404 (no sign-in, account, directory or tenant page).
    for (const path of ["/account", "/sign-in", "/connections/mcp", "/groups"]) {
      const answer = await statusAt(context, `${origin}${path}`);
      expect(answer.status, `${origin}${path}`).toBe(404);
      expect(answer.text).toContain("There is nothing at this address.");
    }
    // The protocol paths stay exempt from the gate.
    const health = await context.newPage();
    expect((await health.goto(`${origin}/api/health`))?.status()).toBe(200);
    await health.close();
  }
  // The signup surface exists only on the root itself; www has no /signup.
  expect((await statusAt(context, `${ROOT}/signup`)).status).toBe(200);
  expect((await statusAt(context, "https://www.localhost:543/signup")).status).toBe(404);
  // OS002: on the root, www and the shared auth host the gate passes exactly a GET or HEAD of the default product mark and
  // the favicon (the pinned bytes, image/png); another method or another /brand/ file stays the fixed 404 there.
  const d = await driver();
  for (const host of ["localhost:543", "www.localhost:543", "auth.localhost:543"]) {
    for (const path of PRODUCT_ASSETS) {
      const got = await d.get(`${WEB_DIRECT}${path}`, { headers: { Host: host }, maxRedirects: 0 });
      expect(got.status(), `GET ${host}${path}`).toBe(200);
      expect(got.headers()["content-type"], `GET ${host}${path}`).toMatch(/^image\/png(;|$)/);
      expect(createHash("sha256").update(await got.body()).digest("hex"), `GET ${host}${path} bytes`).toBe(PRODUCT_MARK_SHA256);
      const head = await d.fetch(`${WEB_DIRECT}${path}`, { method: "HEAD", headers: { Host: host }, maxRedirects: 0 });
      expect(head.status(), `HEAD ${host}${path}`).toBe(200);
      expect(head.headers()["content-type"], `HEAD ${host}${path}`).toMatch(/^image\/png(;|$)/);
      const post = await d.fetch(`${WEB_DIRECT}${path}`, { method: "POST", headers: { Host: host }, maxRedirects: 0 });
      expect(post.status(), `POST ${host}${path}`).toBe(404);
    }
    const other = await d.get(`${WEB_DIRECT}/brand/other.png`, { headers: { Host: host }, maxRedirects: 0 });
    expect(other.status(), `GET ${host}/brand/other.png`).toBe(404);
  }
  // The header of the root and www landing and of the auth host's sign-in page shows the mark, loaded (S-ERRORS also
  // holds the default page free of a failed image or favicon load).
  for (const url of [`${ROOT}/`, "https://www.localhost:543/", `${AUTH}/sign-in`]) {
    tracking.allowed.add(new URL(url).origin);
    expect((await page.goto(url))?.status(), url).toBe(200);
    const mark = page.getByTestId("product-mark");
    await expect(mark, url).toHaveCount(1);
    await expect
      .poll(() => mark.evaluate((img) => ((img as HTMLImageElement).complete ? (img as HTMLImageElement).naturalWidth : 0)), { message: `${url}: mark loaded` })
      .toBeGreaterThan(0);
  }
});

test("[E-DOM-5] a removed custom host routes nowhere: pages are the fixed 404 and a sign-in cannot start @domain-custom", async ({ context }) => {
  const removed = "https://old.acme.test:543";
  for (const o of [removed, PORTAL]) tracking.allowed.add(o);
  for (const path of ["/", "/sign-in", "/account"]) {
    const r = await statusAt(context, `${removed}${path}`);
    expect(r.status, `${removed}${path}`).toBe(404);
    expect(r.text).toContain("There is nothing at this address.");
  }
  expect(await loginStart("old.acme.test:543")).toEqual({ status: 403, location: "" });
  // The gate's refusals stay no-store; a gated JSON route gets its 404 as JSON, a page the fixed HTML page.
  const d = await driver();
  const api = await d.get(`${WEB_DIRECT}/api/account`, { headers: { Host: "old.acme.test:543" } });
  expect(api.status()).toBe(404);
  expect(await api.json()).toEqual({ error: "not_found" });
  expect(api.headers()["cache-control"]).toBe("no-store");
  const pageAnswer = await d.get(`${WEB_DIRECT}/`, { headers: { Host: "old.acme.test:543" } });
  expect(pageAnswer.status()).toBe(404);
  expect(pageAnswer.headers()["content-type"]).toContain("text/html");
  expect(pageAnswer.headers()["cache-control"]).toBe("no-store");
  // The routing custom host of the same Organization still serves its sign-in page.
  expect((await statusAt(context, `${PORTAL}/sign-in`)).status).toBe(200);
});

test("[E-DOM-6] a friendly slug's host routes nowhere: the fixed 404 in place, no Location, no sign-in, nothing disclosed @domain-platform", async ({ context }) => {
  // G057 (Plan r5 §3.4, §5.2 T8; the amendment's "no slug-host redirect discloses the random address"): Organization A's
  // friendly slug alpha-agency is metadata; only its label host org-a routes. The slug host is no alias of it.
  const slugHost = "alpha-agency.localhost:543";
  const slugOrigin = `https://${slugHost}`;
  tracking.allowed.add(slugOrigin);
  for (const path of ["/", "/sign-in", "/account"]) {
    const r = await statusAt(context, `${slugOrigin}${path}`);
    expect(r.status, `${slugOrigin}${path}`).toBe(404);
    expect(r.text).toContain("There is nothing at this address.");
    expect(r.text).not.toContain("org-a");
  }
  // web's own answer on the slug host: the fixed no-store 404 with no Location (no redirect to the label host).
  const d = await driver();
  for (const path of ["/", "/sign-in", "/account"]) {
    const answer = await d.get(`${WEB_DIRECT}${path}`, { headers: { Host: slugHost }, maxRedirects: 0 });
    expect(answer.status(), `web ${path}`).toBe(404);
    expect(answer.headers()["location"], `no Location for ${path}`).toBeUndefined();
    expect(answer.headers()["cache-control"]).toBe("no-store");
    expect(await answer.text()).not.toContain("org-a");
  }
  expect(await loginStart(slugHost)).toEqual({ status: 403, location: "" });
  // core resolves the slug host to nothing and the label host to Organization A.
  const slug = await d.get(`${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(slugOrigin)}`);
  expect(slug.status()).toBe(404);
  expect(await slug.json()).toEqual({ error: "unknown_host" });
  const label = await d.get(`${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(ORIGIN_A)}`);
  expect(label.status()).toBe(200);
  const resolved = (await label.json()) as { kind: string; organization_id?: string };
  expect(resolved.kind).toBe("platform");
  expect(resolved.organization_id).toMatch(UUID);
});

// --- G017 Organization branding (Plan r2 §6.2 E-BRD-1..6) -------------------------------------------------------------
// The overlay seed (infra/identity/mcp-seed.sql) brands Organization A (label with markup characters, a complete palette,
// no logo) and makes carol its owner; Organizations B and C stay unbranded. G069 (Plan S7, §2.5): a tenant page carries
// the Organization's palette at the document root, in the head's <style data-tenant-branding> (html:root:not(.dark) for
// Light, html.dark:root for Dark); an unbranded host renders that element empty.

const SEED_LABEL = 'Acme <Works> & "Sons"';
// G069 S23 / D-M3a (i) (Codex P0 disposition, 2026-10-26): the seeded Light primary is #0a5a4d (was #0f6b5c, whose white
// label on the kit's 80 % hover measured about 4.27:1); infra/identity/mcp-seed.sql carries the same test input.
const SEED_COLOURS = {
  light: { primary: "rgb(10, 90, 77)", background: "rgb(241, 245, 244)" },
  dark: { primary: "rgb(63, 184, 163)", background: "rgb(11, 21, 20)" },
};
// The platform primary per appearance: the kit's own --vx-blue #0075ff, which its primary follows, in both appearances
// (G069 owner direction, exact kit: the app's former #004fb4 Light / #0062d6 Dark correction is removed; Core's platform
// defaults match, core/internal/branding/colors.go).
const PLATFORM_PRIMARY = { light: "rgb(0, 117, 255)", dark: "rgb(0, 117, 255)" };
const BRANDING_PAGE = "/organization/branding";
const LOGO_IMMUTABLE = "public, max-age=86400, immutable";

const schemeOf = (testInfo: TestInfo) => (testInfo.project.use.colorScheme === "dark" ? "dark" : "light");

// One semantic token as the document root resolves it in the current appearance (a probe element in the body inherits
// the root's variables; its computed colour).
async function tenantColour(page: Page, token: string): Promise<string> {
  return page.evaluate((t) => {
    const probe = document.createElement("span");
    probe.style.color = `var(--${t})`;
    document.body.appendChild(probe);
    const value = getComputedStyle(probe).color;
    probe.remove();
    return value;
  }, token);
}

// The text of the tenant stylesheet in the head (exactly one such element on every page).
async function tenantStylesheet(page: Page): Promise<string> {
  const style = page.locator("head > style[data-tenant-branding]");
  await expect(style).toHaveCount(1);
  return (await style.textContent()) ?? "";
}

type LogoAnswer = { status: number; type: string | null; cache: string | null; nosniff: string | null; etag: string | null; disposition: string | null; size: number };
// Fetched from the context's API page of the same origin, so an expected 404 never reaches the test page's S-ERRORS.
async function logoAt(page: Page, path: string): Promise<LogoAnswer> {
  const api = await apiPage(page.context(), new URL(page.url()).origin);
  return api.evaluate(async (p) => {
    const r = await fetch(p, { cache: "no-store" });
    const body = new Uint8Array(await r.arrayBuffer());
    const h = r.headers;
    return {
      status: r.status,
      type: h.get("content-type"),
      cache: h.get("cache-control"),
      nosniff: h.get("x-content-type-options"),
      etag: h.get("etag"),
      disposition: h.get("content-disposition"),
      size: body.byteLength,
    };
  }, path);
}

// Submits one management form on the branding page and waits for the page's answer (?status=...).
async function submitBranding(page: Page, form: string, button: string): Promise<URL> {
  await page.getByTestId(form).getByRole("button", { name: button }).click();
  await page.waitForURL((u) => u.pathname === BRANDING_PAGE && u.searchParams.has("status"));
  return new URL(page.url());
}

async function currentLogoVersion(page: Page): Promise<string> {
  const src = await page.getByTestId("branding-logo-preview").getAttribute("src");
  const v = new URL(src ?? "", page.url()).searchParams.get("v") ?? "";
  expect(v).toMatch(/^[0-9a-f]{12}$/);
  return v;
}

test("[E-BRD-1] the branded custom host renders its label escaped and its palette in Light and Dark @domain-custom", async ({ page }) => {
  tracking.allowed.add(PORTAL);
  const r = await page.goto(`${PORTAL}/`);
  expect(r?.status()).toBe(200);
  expect(r?.headers()["cache-control"]).toContain("no-store");
  const html = (await r?.text()) ?? "";
  expect(html).toContain("Acme &lt;Works&gt; &amp;");
  expect(html).not.toContain("<Works>");
  await expect(page.getByTestId("tenant-name")).toHaveText(SEED_LABEL);
  await expect(page.getByTestId("brand-label")).toHaveText(SEED_LABEL);
  await expect(page.getByTestId("brand-logo")).toHaveCount(0);
  await expect(page.getByTestId("entry-sign-in")).toBeVisible();
  await expect(page.getByTestId("platform-attribution")).toBeVisible();
  // Both appearances' saved values are in the first HTML, at the document root (Light block, then Dark block).
  const sheet = await tenantStylesheet(page);
  expect(sheet, "Light block carries the seeded primary").toMatch(/^html:root:not\(\.dark\)\{[^}]*--primary:#0a5a4d[;}]/);
  expect(sheet, "Dark block carries the seeded primary").toMatch(/html\.dark:root\{[^}]*--primary:#3fb8a3[;}]/);
  for (const scheme of ["light", "dark"] as const) {
    await page.emulateMedia({ colorScheme: scheme });
    // No stored choice: the theme provider follows the OS and re-classes <html>.
    await expect.poll(() => page.evaluate((s) => document.documentElement.classList.contains(s), scheme), { message: `<html> class follows OS ${scheme}` }).toBe(true);
    expect(await tenantColour(page, "primary"), `${scheme} primary`).toBe(SEED_COLOURS[scheme].primary);
    expect(await tenantColour(page, "background"), `${scheme} background`).toBe(SEED_COLOURS[scheme].background);
  }
});

test("[E-BRD-2] an unbranded platform host falls back to the product name and platform palette @domain-platform", async ({ page }, testInfo) => {
  tracking.allowed.add(ORIGIN_C);
  const r = await page.goto(`${ORIGIN_C}/`);
  expect(r?.status()).toBe(200);
  const name = (await page.getByTestId("tenant-name").textContent()) ?? "";
  expect(name.length).toBeGreaterThan(0);
  expect(name).not.toBe(SEED_LABEL);
  await expect(page.getByTestId("brand-label")).toHaveText(name);
  await expect(page.getByTestId("brand-logo")).toHaveCount(0);
  // No saved palette: the tenant stylesheet is empty, so the kit's defaults apply.
  expect(await tenantStylesheet(page), "no tenant values").toBe("");
  expect(await tenantColour(page, "primary")).toBe(PLATFORM_PRIMARY[schemeOf(testInfo)]);
  // OS002: the unbranded frame shows the default product mark before the product name; a tenant with a saved label
  // (Organization A) shows its label alone, without the mark.
  await expect(page.getByTestId("brand-label").getByTestId("product-mark")).toHaveCount(1);
  expect((await page.goto(`${ORIGIN_A}/`))?.status()).toBe(200);
  await expect(page.getByTestId("brand-label")).toHaveText(SEED_LABEL);
  await expect(page.getByTestId("product-mark")).toHaveCount(0);
});

test("[E-BRD-3] unknown hosts stay the fixed 404 for the branding routes; a stale logo version is a no-store 404 @domain-platform", async ({ context }) => {
  for (const origin of ["https://org-z.localhost:543", "https://unknown.acme.test:543"]) {
    tracking.allowed.add(origin);
    for (const path of [BRANDING_PAGE, "/branding/logo?v=0123456789ab"]) {
      const r = await statusAt(context, `${origin}${path}`);
      expect(r.status, `${origin}${path}`).toBe(404);
    }
  }
  const d = await driver();
  for (const host of ["org-c.localhost:543", "org-a.localhost:543", "old.acme.test:543"]) {
    const logo = await d.get(`${WEB_DIRECT}/branding/logo?v=0123456789ab`, { headers: { Host: host } });
    expect(logo.status(), host).toBe(404);
    expect(logo.headers()["cache-control"], host).toBe("no-store");
  }
  // Anonymous on a routing host: the management page asks to sign in first.
  tracking.allowed.add(ORIGIN_C);
  const anon = await statusAt(context, `${ORIGIN_C}${BRANDING_PAGE}`);
  expect(anon.status).toBe(200);
  expect(anon.text).toContain("Sign in to this workspace to manage its branding.");
});

test("[E-BRD-4] a signed-in member sees the member indicator; a non-manager gets 404 and no account link @domain-platform", async ({ page }) => {
  tracking.allowed.add(ORIGIN_C);
  await signIn(page, ALICE, ORIGIN_C);
  await page.goto(`${ORIGIN_C}/`);
  const indicator = page.getByTestId("context-indicator");
  await expect(indicator).toHaveAttribute("data-membership", "member");
  await expect(indicator).toContainText("member of");
  await expect(page.getByTestId("entry-sign-in")).toHaveCount(0);
  await page.goto(`${ORIGIN_C}/account`);
  await expect(page.getByTestId("display-name")).toBeVisible();
  await expect(page.getByTestId("account-branding-link")).toHaveCount(0);
  expect((await statusAt(page.context(), `${ORIGIN_C}${BRANDING_PAGE}`)).status).toBe(404);
});

test("[E-BRD-5] the owner manages the logo: PNG accepted, SVG refused, replacement changes v, stale v 404; contrast and version refusals @domain-platform", async ({ page, browser }, testInfo) => {
  await signIn(page, CAROL, ORIGIN_A);
  await page.goto(`${ORIGIN_A}/account`);
  await page.getByTestId("account-branding-link").click();
  await page.waitForURL((u) => u.pathname === BRANDING_PAGE);
  await expect(page.getByTestId("branding-logo-none")).toBeVisible();
  await expect(page.locator("#display_label")).toHaveValue(SEED_LABEL);
  // G068, retyped to the kit Input (G069 S22): the branding fields and the logo file input ease their focus state in as
  // every kit field does (with and without a reduced-motion preference) and settle on the kit's ring.
  await expectFocusMotion(page, page.locator("#display_label"), "workspace name");
  await expectFocusMotion(page, page.locator("#light-primary"), "light primary colour");
  await expectFocusMotion(page, page.locator("#logo-file"), "logo file");
  try {
    // An SVG is refused and changes nothing.
    await page.locator("#logo-file").setInputFiles("tests/fixtures/branding/logo-vector.svg");
    expect((await submitBranding(page, "branding-logo-upload", "Upload logo")).searchParams.get("status")).toBe("logo_invalid");
    await expect(page.getByTestId("branding-logo-none")).toBeVisible();
    // A PNG is accepted: normalized bytes behind a versioned, immutable URL that the public entry references.
    await page.locator("#logo-file").setInputFiles("tests/fixtures/branding/logo-teal.png");
    expect((await submitBranding(page, "branding-logo-upload", "Upload logo")).searchParams.get("status")).toBe("logo_set");
    const v1 = await currentLogoVersion(page);
    const first = await logoAt(page, `/branding/logo?v=${v1}`);
    expect(first).toMatchObject({ status: 200, type: "image/png", cache: LOGO_IMMUTABLE, nosniff: "nosniff" });
    expect(first.etag).toBeTruthy();
    expect(first.size).toBeGreaterThan(0);
    expect(first.size).not.toBe(readFileSync("tests/fixtures/branding/logo-teal.png").byteLength);
    await page.goto(`${ORIGIN_A}/`);
    expect(await page.getByTestId("brand-logo").getAttribute("src")).toBe(`/branding/logo?v=${v1}`);
    expect(await page.getByTestId("brand-logo").getAttribute("alt")).toBe(SEED_LABEL);
    // OS002: a tenant logo replaces the default product mark.
    await expect(page.getByTestId("product-mark")).toHaveCount(0);
    // A replacement changes v; the old v is refused at the origin (no-store) and the old object is listed as an upload.
    await page.goto(`${ORIGIN_A}${BRANDING_PAGE}`);
    await page.locator("#logo-file").setInputFiles("tests/fixtures/branding/logo-amber.png");
    expect((await submitBranding(page, "branding-logo-upload", "Upload logo")).searchParams.get("status")).toBe("logo_set");
    const v2 = await currentLogoVersion(page);
    expect(v2).not.toBe(v1);
    expect(await logoAt(page, `/branding/logo?v=${v1}`)).toMatchObject({ status: 404, cache: "no-store" });
    expect((await logoAt(page, `/branding/logo?v=${v2}`)).status).toBe(200);
    await expect(page.getByTestId("branding-upload")).toHaveCount(1);
    // The replacement is a legal extreme aspect ratio (2048×16): at 320px its preview stays inside its box and the
    // viewport, aspect ratio kept by containment, alt text unchanged, and the page does not scroll sideways.
    const viewport = page.viewportSize()!;
    await page.setViewportSize({ width: 320, height: 640 });
    const preview = page.getByTestId("branding-logo-preview");
    await expect(preview).toHaveAttribute("alt", /^Current logo of /);
    await expect.poll(() => preview.evaluate((img) => (img as HTMLImageElement).complete && (img as HTMLImageElement).naturalWidth)).toBe(2048);
    const fit = await preview.evaluate((img) => {
      const r = img.getBoundingClientRect();
      const b = img.parentElement!.getBoundingClientRect();
      return { natural: [(img as HTMLImageElement).naturalWidth, (img as HTMLImageElement).naturalHeight], objectFit: getComputedStyle(img).objectFit, imgLeft: r.left, imgRight: r.right, boxLeft: b.left, boxRight: b.right, scrollWidth: document.documentElement.scrollWidth, innerWidth };
    });
    expect(fit.natural).toEqual([2048, 16]);
    expect(fit.objectFit).toBe("contain");
    expect(fit.boxLeft, "preview box inside the viewport").toBeGreaterThanOrEqual(-0.5);
    expect(fit.boxRight, "preview box inside the viewport").toBeLessThanOrEqual(fit.innerWidth + 0.5);
    expect(fit.imgLeft, "image inside its box").toBeGreaterThanOrEqual(fit.boxLeft - 0.5);
    expect(fit.imgRight, "image inside its box").toBeLessThanOrEqual(fit.boxRight + 0.5);
    expect(fit.scrollWidth, "no horizontal page scroll at 320px").toBeLessThanOrEqual(fit.innerWidth);
    await page.setViewportSize(viewport);
    // An unreadable palette is refused with the failing pair; nothing changes. G069 kit-baseline floor (Core requires
    // min(threshold, the kit's own ratio) per pair): #0f6b5c on the seeded Light primary #0a5a4d is about 1.27:1, below
    // the T8 at-rest floor (the kit's 4.2097:1, OWNER-WAIVED below 4.5), and no earlier pair involves the label colour,
    // so T8:light is still the first failing pair.
    await page.locator('input[name="light.primary-foreground"]').fill("#0f6b5c");
    const refused = await submitBranding(page, "branding-update", "Save branding");
    expect(refused.searchParams.get("status")).toBe("contrast_failed");
    expect(refused.searchParams.get("field")).toBe("T8:light");
    await expect(page.getByText(/Failed check T8 in Light/)).toBeVisible();
    await expect(page.locator('input[name="light.primary-foreground"]')).toHaveValue("#ffffff");
    // A stale form (another tab loaded before a change) is refused as version_conflict.
    const other = await browser.newContext({ baseURL: ORIGIN_A, colorScheme: testInfo.project.use.colorScheme, storageState: await page.context().storageState() });
    track(other);
    tracking.contexts.push(other);
    const stale = await other.newPage();
    await stale.goto(`${ORIGIN_A}${BRANDING_PAGE}`);
    expect((await submitBranding(page, "branding-logo-remove", "Remove logo")).searchParams.get("status")).toBe("logo_removed");
    expect((await submitBranding(stale, "branding-update", "Save branding")).searchParams.get("status")).toBe("version_conflict");
    expect(await logoAt(page, `/branding/logo?v=${v2}`)).toMatchObject({ status: 404, cache: "no-store" });
    // A non-manager member gets the fixed 404 on the management page.
    const member = await open(browser, testInfo);
    await signIn(member.page, ALICE, ORIGIN_A);
    expect((await statusAt(member.context, `${ORIGIN_A}${BRANDING_PAGE}`)).status).toBe(404);
  } finally {
    // Restore the seeded visible state: no logo, no uploads (the label and colours never changed).
    await page.goto(`${ORIGIN_A}${BRANDING_PAGE}`);
    if ((await page.getByTestId("branding-logo-remove").count()) > 0) await submitBranding(page, "branding-logo-remove", "Remove logo");
    for (let i = 0; i < 4 && (await page.getByTestId("branding-upload").count()) > 0; i++) {
      await page.getByTestId("branding-upload").first().getByRole("button", { name: "Delete this upload" }).click();
      await page.waitForURL((u) => u.pathname === BRANDING_PAGE && u.searchParams.has("status"));
    }
  }
  await expect(page.getByTestId("branding-logo-none")).toBeVisible();
  await expect(page.getByTestId("branding-uploads-empty")).toBeVisible();
});

test("[E-BRD-6] interleaved hosts and viewers get exactly their own pages; HTML is never cacheable @domain-custom", async ({ page, browser }, testInfo) => {
  for (const o of [PORTAL, AUTH, ORIGIN_B]) tracking.allowed.add(o);
  const anonA = await page.goto(`${PORTAL}/`);
  expect(anonA?.headers()["cache-control"]).toContain("no-store");
  await expect(page.getByTestId("tenant-name")).toHaveText(SEED_LABEL);
  await expect(page.getByTestId("context-indicator")).toHaveCount(0);
  const b = await open(browser, testInfo);
  const anonB = await b.page.goto(`${ORIGIN_B}/`);
  expect(anonB?.headers()["cache-control"]).toContain("no-store");
  await expect(b.page.getByTestId("tenant-name")).not.toHaveText(SEED_LABEL);
  expect(await tenantStylesheet(b.page), "B: no tenant values").toBe("");
  expect(await tenantColour(b.page, "primary")).toBe(PLATFORM_PRIMARY[schemeOf(testInfo)]);
  const m = await open(browser, testInfo);
  await signIn(m.page, ALICE, PORTAL);
  const memberA = await m.page.goto(`${PORTAL}/`);
  expect(memberA?.headers()["cache-control"]).toContain("no-store");
  await expect(m.page.getByTestId("context-indicator")).toHaveAttribute("data-membership", "member");
  await expect(m.page.getByTestId("tenant-name")).toHaveText(SEED_LABEL);
  const again = await page.goto(`${PORTAL}/`);
  expect(again?.headers()["cache-control"]).toContain("no-store");
  await expect(page.getByTestId("context-indicator")).toHaveCount(0);
  await expect(page.getByTestId("entry-sign-in")).toBeVisible();
  expect(await tenantColour(page, "primary")).toBe(SEED_COLOURS.dark.primary);
  expect(await logoAt(page, "/branding/logo?v=0123456789ab")).toMatchObject({ status: 404, cache: "no-store" });
});
