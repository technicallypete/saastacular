import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { dirname } from "node:path";

import { request as playwrightRequest, type APIRequestContext, type Browser, type BrowserContext, type Cookie, type Page, type TestInfo } from "@playwright/test";

import { expect, test } from "./support/page-errors";
import { observe } from "./support/sanitize";

// G071 default-tenant root alias (Plan r4 §1.1 S4–S6, §5.1 B1 E-DT-1..6) in the disposable HTTPS identity environment
// (infra/compose/identity.yml), project default-tenant-light with the base https://localhost:543 (the platform root;
// PLATFORM_BASE_DOMAIN=localhost, PLATFORM_HOST_PORT=543). The coordinator's private G071 stack restarts core with one
// configuration per phase (IDENTITY_TEST_DEFAULT_TENANT_SLUG, IDENTITY_TEST_PLATFORM_ROOT_PAGE; the overlay only passes
// the real core variables through, no test bypass) and names it in DEFAULT_TENANT_PHASE; each scenario runs only in its
// own phase. The project is not in IDENTITY.md's identity:browser list, so the shared harness never runs it.
//   phase 1  DEFAULT_TENANT_SLUG=alpha-agency, PLATFORM_ROOT_PAGE=landing: E-DT-1, E-DT-2
//   phase 2  DEFAULT_TENANT_SLUG=bravo-agency, PLATFORM_ROOT_PAGE=signup_redirect: E-DT-3 (reads E-DT-1's handoff)
//   phase 4  DEFAULT_TENANT_SLUG unset, PLATFORM_ROOT_PAGE="" (the product default signup_redirect): E-DT-5
//   phase 5  DEFAULT_TENANT_SLUG=nope-nobody (unknown), PLATFORM_ROOT_PAGE=bogus (invalid): E-DT-6
// (E-DT-4 is signup.spec.ts E-SUP-1, reused unchanged in phases 0 and 3.) Every step goes through the real BFF, core and
// the pinned emulator; nothing is stubbed. The seed (infra/identity/mcp-seed.sql) makes alpha-agency Organization A
// (label org-a, invite-only, branded with SEED_LABEL) and alice a plain member of it, and bravo-agency Organization B
// (label org-b, open enrollment, no public label). Every page request of every browser context is captured and must
// stay on the app origins under test and the emulator (as identity.spec.ts E-NET-1). Pages that answer an error status
// are opened in extra pages, so each test's default page stays under S-ERRORS. Answers whose exact status, Location or
// headers matter (a 307 is followed by the browser) are read by a driver straight from web's own port with the Host
// header as sent (signup.spec.ts, identity.spec.ts: the TLS edge adds nothing but that header), and core's
// hosts/resolve answer is read directly from core (signup.spec.ts organizationOf).

const ROOT = "https://localhost:543";
const ROOT_HOST = "localhost:543";
const WWW = "https://www.localhost:543";
const ORIGIN_A = "https://org-a.localhost:543";
const ORIGIN_B = "https://org-b.localhost:543";
const AUTH = "https://auth.localhost:543";
const NOBODY = "https://nobody.localhost:543";
const EMULATOR = "http://emulator:4100";
const CORE = "http://core:8180";
// web's own port in the shared network namespace (the TLS edge proxies to it with the Host header as sent).
const WEB_DIRECT = "http://127.0.0.1:3000";
const NOT_FOUND = "There is nothing at this address.";
const LOGIN_COOKIE = "__Host-login";
const SIGNUP_COOKIE = "__Host-signup";
// Every /signup path (the root signup surface, the tenant continuation hops and an unknown one): all of them are the
// fixed 404 on the aliased root (Plan r4 S6).
const SIGNUP_PATHS = [
  "/signup",
  "/signup/details",
  "/signup/start",
  "/signup/create",
  "/signup/abandon",
  "/signup/prove",
  "/signup/continue",
  "/signup/complete",
  "/signup/x",
];
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const SESSION_COOKIE = "__Host-sid";
// Organization A's seeded public label (mcp-seed.sql organization_branding; rendered escaped).
const SEED_LABEL = 'Acme <Works> & "Sons"';

// The phases of the private stack that run this project (phases 0 and 3 run signup-light's E-SUP-1 instead).
const PROJECT_PHASES = ["1", "2", "4", "5"];
const PHASE = process.env.DEFAULT_TENANT_PHASE ?? "";

// Runs the calling test only in the named phase of the private stack.
function onlyInPhase(phase: string) {
  test.skip(PHASE !== phase, `B1 phase ${phase} only (this run is phase ${PHASE || "unset"})`);
}

type Credentials = { email: string; password: string };
type SeedUser = Credentials & { name: string; first_name: string; last_name: string };
type Seed = { users: SeedUser[]; emulator_api_key: string };
const SEED = JSON.parse(readFileSync("/identity/browser/users.json", "utf8")) as Seed;
function seeded(name: string): SeedUser {
  const found = SEED.users.find((u) => u.name === name);
  if (!found) throw new Error(`seed user ${name} missing`);
  return found;
}
const ALICE = seeded("alice");

// --- page traffic (E-NET-1) ---------------------------------------------------------------------------------------------
type Tracking = { origins: Set<string>; allowed: Set<string>; contexts: BrowserContext[]; drivers: APIRequestContext[] };
let tracking: Tracking;

function track(context: BrowserContext) {
  context.on("request", (request) => {
    const url = request.url();
    if (/^https?:/.test(url)) tracking.origins.add(new URL(url).origin);
  });
}

test.beforeEach(async ({ context }) => {
  tracking = { origins: new Set(), allowed: new Set([ROOT, EMULATOR]), contexts: [], drivers: [] };
  track(context);
});

test.afterEach(async () => {
  for (const c of tracking.contexts) await c.close();
  for (const d of tracking.drivers) await d.dispose();
  const outside = [...tracking.origins].filter((o) => !tracking.allowed.has(o));
  expect(outside, "E-NET-1: page requests outside the app origins under test and the emulator").toEqual([]);
});

// A fresh browser profile of this project (same pinned browser), tracked and closed like any other.
async function open(browser: Browser, testInfo: TestInfo): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ baseURL: ROOT, colorScheme: testInfo.project.use.colorScheme, viewport: testInfo.project.use.viewport });
  track(context);
  tracking.contexts.push(context);
  return { context, page: await context.newPage() };
}

// A Node request client for web's own port and core (never a page; it resolves no *.localhost name and trusts no run
// certificate, so it talks plain HTTP inside the shared network namespace only).
async function driver(): Promise<APIRequestContext> {
  const d = await playwrightRequest.newContext();
  tracking.drivers.push(d);
  return d;
}

// --- sign-in through the emulator's hosted page (identity.spec.ts shapes) ----------------------------------------------
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

// Opens ORIGIN's sign-in page (returning to /account), submits it, completes the hosted sign-in and waits to be back on
// an application origin.
async function signIn(page: Page, who: Credentials, origin: string) {
  await page.goto(`${origin}/sign-in?return_to=/account`);
  await page.getByRole("button", { name: "Sign in" }).click();
  await enterCredentials(page, who);
  await page.waitForURL((u) => u.origin !== EMULATOR);
}

// --- observations -------------------------------------------------------------------------------------------------------
// The cookies the browser would send with a request to url, by RFC 6265 domain matching (signup.spec.ts cookiesSentTo):
// context.cookies(url) alone also lists the root's host-only localhost cookies under https://<label>.localhost.
async function cookiesSentTo(context: BrowserContext, url: string): Promise<Cookie[]> {
  const host = new URL(url).hostname;
  return (await context.cookies(url)).filter((c) => (c.domain.startsWith(".") ? `.${host}`.endsWith(c.domain) : c.domain === host));
}

async function sessionCookie(context: BrowserContext, origin: string): Promise<Cookie | undefined> {
  const cookie = (await cookiesSentTo(context, origin)).find((c) => c.name === SESSION_COOKIE);
  observe(cookie?.value);
  return cookie;
}

// Browser-side calls run as same-origin fetches from a helper page of the context (Chromium resolves *.localhost and
// trusts the pinned key; the Node request client does neither), carrying exactly the browser's cookies.
async function accountAt(context: BrowserContext, origin: string): Promise<{ status: number; body: Record<string, unknown> }> {
  const p = await context.newPage();
  await p.goto(`${origin}/api/health`);
  const r = await p.evaluate(async () => {
    const res = await fetch("/api/account", { cache: "no-store" });
    return { status: res.status, cacheControl: res.headers.get("cache-control"), body: (await res.json()) as Record<string, unknown> };
  });
  await p.close();
  expect(r.cacheControl).toBe("no-store");
  return { status: r.status, body: r.body };
}

// A page's answer in an extra page of the context (error statuses stay off the test's default page).
async function statusAt(context: BrowserContext, url: string): Promise<{ status: number; url: string; text: string }> {
  const p = await context.newPage();
  const r = await p.goto(url);
  const out = { status: r?.status() ?? 0, url: p.url(), text: (await p.locator("body").textContent()) ?? "" };
  await p.close();
  return out;
}

// web's own answer for one request with this Host header, never following a redirect: the status, Location,
// Cache-Control, whether any Set-Cookie was sent, and the body (empty for HEAD). A POST carries a small form body.
type Direct = { status: number; location: string | undefined; cacheControl: string | undefined; setCookie: boolean; text: string };
async function direct(method: "GET" | "HEAD" | "POST", path: string, host: string): Promise<Direct> {
  const d = await driver();
  const r = await d.fetch(`${WEB_DIRECT}${path}`, { method, headers: { Host: host }, maxRedirects: 0, ...(method === "POST" ? { form: { probe: "1" } } : {}) });
  const h = r.headers();
  return {
    status: r.status(),
    location: h["location"],
    cacheControl: h["cache-control"],
    setCookie: r.headersArray().some((x) => x.name.toLowerCase() === "set-cookie"),
    text: method === "HEAD" ? "" : await r.text(),
  };
}

// core's hosts/resolve answer for an origin (the gate's own lookup, read directly; no credential).
async function coreResolve(origin: string): Promise<{ status: number; body: Record<string, unknown> }> {
  const d = await driver();
  const r = await d.get(`${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(origin)}`);
  return { status: r.status(), body: (await r.json()) as Record<string, unknown> };
}

// The Organization a tenant label host resolves to (a plain platform answer).
async function organizationOf(origin: string): Promise<string> {
  const r = await coreResolve(origin);
  expect(r.status, `hosts/resolve ${origin}`).toBe(200);
  expect(r.body).toEqual({ kind: "platform", organization_id: expect.stringMatching(UUID) });
  return r.body.organization_id as string;
}

// core's answer for the aliased root: exactly the platform kind, the pinned Organization and root_alias (Plan r4 S5).
async function expectRootAliasedTo(organizationId: string) {
  const r = await coreResolve(ROOT);
  expect(r.status, "hosts/resolve of the aliased root").toBe(200);
  expect(r.body).toEqual({ kind: "platform", organization_id: organizationId, root_alias: true });
}

// core's answer for a public entry host: exactly the kind, the root's signup_origin and the effective presentation
// (Plan r4 S16).
async function expectPublicEntry(origin: string, rootPage: "signup_redirect" | "landing") {
  const r = await coreResolve(origin);
  expect(r.status, `hosts/resolve ${origin}`).toBe(200);
  expect(r.body).toEqual({ kind: "public_entry", signup_origin: ROOT, root_page: rootPage });
}

// The landing of a public entry host served in place: one Sign Up to the root's /signup, no redirect hop.
async function expectLanding(page: Page, origin: string) {
  const r = await page.goto(`${origin}/`);
  expect(r?.status(), `${origin}/`).toBe(200);
  expect(r?.request().redirectedFrom()).toBeNull();
  expect(page.url()).toBe(`${origin}/`);
  await expect(page.getByTestId("root-sign-up")).toHaveCount(1);
  await expect(page.getByTestId("root-sign-up")).toHaveAttribute("href", `${ROOT}/signup`);
  await expect(page.getByTestId("tenant-name")).toHaveCount(0);
}

// The tenant entry of a host served in place (no redirect hop); returns the rendered workspace name.
async function expectTenantEntry(page: Page, origin: string): Promise<string> {
  const r = await page.goto(`${origin}/`);
  expect(r?.status(), `${origin}/`).toBe(200);
  expect(r?.request().redirectedFrom()).toBeNull();
  expect(page.url()).toBe(`${origin}/`);
  await expect(page.getByTestId("tenant-name")).toHaveCount(1);
  await expect(page.getByTestId("root-sign-up")).toHaveCount(0);
  return ((await page.getByTestId("tenant-name").textContent()) ?? "").trim();
}

// The root's redirect for the product default and the fallback (Plan r4 S17), read from web's own answer: GET and HEAD
// of exactly / answer 307 with exactly Location <root origin>/signup and no-store, no cookie and no body; a request query is never
// forwarded; POST / and /signup itself are not redirected.
async function expectRootRedirect() {
  for (const method of ["GET", "HEAD"] as const) {
    const r = await direct(method, "/", ROOT_HOST);
    expect(r.status, `${method} /`).toBe(307);
    expect(r.location, `${method} / Location`).toBe(`${ROOT}/signup`);
    expect(r.cacheControl, `${method} / Cache-Control`).toBe("no-store");
    expect(r.setCookie, `${method} / sets no cookie`).toBe(false);
    // The gate's response has no body; Next.js' proxy adapter answers a GET redirect with the target URL as plain text
    // (observed in run b1-4), so the only content allowed is exactly the Location itself, never a page.
    expect([ "", `${ROOT}/signup` ], `${method} / body is empty or exactly the Location`).toContain(r.text);
    if (method === "HEAD") expect(r.text, "HEAD / has no body").toBe("");
  }
  for (const query of ["?utm=1&state=x", "?x=1&state=y"]) {
    const r = await direct("GET", `/${query}`, ROOT_HOST);
    expect(r.status, `GET /${query}`).toBe(307);
    expect(r.location, `GET /${query} forwards nothing`).toBe(`${ROOT}/signup`);
  }
  const post = await direct("POST", "/", ROOT_HOST);
  expect(post.status, "POST / is not redirected").not.toBe(307);
  expect(post.location, "POST / has no Location").toBeUndefined();
  const signup = await direct("GET", "/signup", ROOT_HOST);
  expect(signup.status, "GET /signup (no loop)").toBe(200);
  expect(signup.location).toBeUndefined();
}

// Following the root's redirect in the browser: exactly one 307 hop from / to /signup (no loop), the signup start page,
// and nothing started: no request to the provider and no login, signup or session cookie on the root.
async function expectBrowserFollowsToSignup(page: Page, path: string) {
  const r = await page.goto(`${ROOT}${path}`);
  expect(page.url(), `${path} lands on /signup with nothing forwarded`).toBe(`${ROOT}/signup`);
  expect(r?.status()).toBe(200);
  const hop = r?.request().redirectedFrom();
  expect(hop?.url()).toBe(`${ROOT}${path}`);
  expect((await hop?.response())?.status()).toBe(307);
  expect(hop?.redirectedFrom()).toBeNull();
  await expect(page.getByTestId("signup-start")).toHaveCount(1);
  // The calling scenarios contact the provider nowhere else, so any emulator request would be this redirect's.
  expect(tracking.origins.has(EMULATOR), "no provider request").toBe(false);
  const names = (await cookiesSentTo(page.context(), ROOT)).map((c) => c.name);
  for (const name of [LOGIN_COOKIE, SIGNUP_COOKIE, SESSION_COOKIE]) expect(names, `no ${name} on the root`).not.toContain(name);
}

// The phase-1 root session E-DT-1 handed over (beside the report, as identity.spec.ts E-HANDOFF). The coordinator makes
// the phase-1 file available in this run's report directory; its secret is registered with observe().
function readHandoff(testInfo: TestInfo): { origin: string; cookie: string; live: string } {
  const report = dirname(process.env.PLAYWRIGHT_JSON_OUTPUT_FILE ?? "/tmp/playwright/results.json");
  const file = `${report}/handoff-${testInfo.project.name}.json`;
  expect(existsSync(file), `the phase-1 handoff ${file} (written by E-DT-1) is present`).toBe(true);
  const handoff = JSON.parse(readFileSync(file, "utf8")) as { origin: string; cookie: string; live: string };
  observe(handoff.live);
  expect(handoff.origin).toBe(ROOT);
  expect(handoff.cookie).toBe(SESSION_COOKIE);
  expect(SECRET.test(handoff.live), "the handed-over session is a 43-character secret").toBe(true);
  return handoff;
}

// =======================================================================================================================

test("[E-DT-1] the aliased root is Organization A's tenant entry; a member signs in there with a host-only session; account, groups and join pages work", async ({ page, context }, testInfo) => {
  onlyInPhase("1");
  // The root / is the tenant entry in place: no redirect hop, no landing, the tenant's branded name and one Sign in.
  const r = await page.goto(`${ROOT}/`);
  expect(r?.status()).toBe(200);
  expect(r?.request().redirectedFrom()).toBeNull();
  expect(page.url()).toBe(`${ROOT}/`);
  await expect(page.getByTestId("tenant-name")).toHaveText(SEED_LABEL);
  await expect(page.getByTestId("root-sign-up")).toHaveCount(0);
  await expect(page.getByTestId("entry-sign-in")).toHaveCount(1);
  // Organization A is invite-only: the entry offers no Join link.
  await expect(page.getByTestId("entry-join")).toHaveCount(0);

  // A seeded member signs in on the root itself: the provider returns directly to the root's own callback (a platform
  // tenant host, no relay) and the session cookie is host-only on localhost, never sent to a tenant label host.
  const authorize: string[] = [];
  page.on("request", (req) => {
    if (req.url().startsWith(`${EMULATOR}/user_management/authorize?`)) authorize.push(req.url());
  });
  await signIn(page, ALICE, ROOT);
  expect(new URL(page.url()).origin).toBe(ROOT);
  expect(new URL(page.url()).pathname).toBe("/account");
  expect(new URL(page.url()).searchParams.get("status")).toBe("signed_in");
  expect(authorize).toHaveLength(1);
  expect(new URL(authorize[0]).searchParams.get("redirect_uri")).toBe(`${ROOT}/auth/callback`);
  const sid = await sessionCookie(context, ROOT);
  expect(sid).toMatchObject({ domain: "localhost", path: "/", httpOnly: true, secure: true, sameSite: "Lax" });
  expect(sid?.value).toMatch(SECRET);
  expect(await sessionCookie(context, ORIGIN_A)).toBeUndefined();
  const me = await accountAt(context, ROOT);
  expect(me.status).toBe(200);
  expect(me.body.id).toMatch(UUID);

  // The tenant pages answer on the root as on any tenant host.
  const accountPage = await page.goto(`${ROOT}/account`);
  expect(accountPage?.status()).toBe(200);
  expect(page.url()).toBe(`${ROOT}/account`);
  const groups = await page.goto(`${ROOT}/groups/new`);
  expect(groups?.status()).toBe(200);
  expect(page.url()).toBe(`${ROOT}/groups/new`);
  await expect(page.getByTestId("group-name-form")).toHaveCount(1);
  const join = await page.goto(`${ROOT}/join`);
  expect(join?.status()).toBe(200);
  expect(page.url()).toBe(`${ROOT}/join`);
  await expect(page.getByTestId("join-invite-only")).toHaveCount(1);
  // Signed in, the entry shows the Account action in place of Sign in.
  await page.goto(`${ROOT}/`);
  await expect(page.getByTestId("tenant-name")).toHaveText(SEED_LABEL);
  await expect(page.getByTestId("entry-sign-in")).toHaveCount(0);

  // Hands the phase-1 root session to the later phases (E-DT-3 replays it after the A→B restart), as identity.spec.ts
  // E-HANDOFF does: beside the report, readable by the coordinator's user, protected by the report's 0700 parent.
  const report = dirname(process.env.PLAYWRIGHT_JSON_OUTPUT_FILE ?? "/tmp/playwright/results.json");
  const file = `${report}/handoff-${testInfo.project.name}.json`;
  writeFileSync(file, JSON.stringify({ origin: ROOT, cookie: SESSION_COOKIE, live: sid!.value }) + "\n", { mode: 0o644 });
});

test("[E-DT-2] the aliased root serves no /signup path, GET or POST; www, tenant, auth and unknown hosts keep their own answers", async ({ page, context }) => {
  onlyInPhase("1");
  for (const o of [WWW, ORIGIN_A, AUTH, NOBODY]) tracking.allowed.add(o);
  // core: the root is Organization A's own platform host with root_alias; www stays the public entry with the selected
  // presentation (landing in this phase); the auth host and an unknown label are unchanged.
  const orgA = await organizationOf(ORIGIN_A);
  await expectRootAliasedTo(orgA);
  await expectPublicEntry(WWW, "landing");
  expect(await coreResolve(AUTH)).toEqual({ status: 200, body: { kind: "auth" } });
  expect(await coreResolve(NOBODY)).toEqual({ status: 404, body: { error: "unknown_host" } });

  // Every /signup path on the root is the fixed 404, in the browser and as web's own answer, for GET and POST alike; no
  // redirect and no Location.
  for (const path of SIGNUP_PATHS) {
    const answer = await statusAt(context, `${ROOT}${path}`);
    expect(answer.status, `GET ${path}`).toBe(404);
    expect(answer.url, `GET ${path} served in place`).toBe(`${ROOT}${path}`);
    expect(answer.text).toContain(NOT_FOUND);
    for (const method of ["GET", "POST"] as const) {
      const r = await direct(method, path, ROOT_HOST);
      expect(r.status, `${method} ${path}`).toBe(404);
      expect(r.location, `${method} ${path} has no Location`).toBeUndefined();
      expect(r.cacheControl, `${method} ${path}`).toBe("no-store");
      expect(r.text).toContain(NOT_FOUND);
    }
  }
  // Without a session the root's account API answers as on any tenant host (the route's own 401, not the gate's 404).
  const anonymous = await accountAt(context, ROOT);
  expect(anonymous.status).toBe(401);
  expect(anonymous.body).toEqual({ error: "unauthenticated" });

  // The other hosts are unchanged: www keeps its landing (Sign Up then reaches the root's 404 while the alias is active),
  // org-a its own tenant entry, the auth host only its sign-in page, and an unknown label the fixed 404.
  await expectLanding(page, WWW);
  expect(await expectTenantEntry(page, ORIGIN_A)).toBe(SEED_LABEL);
  // The auth host's sign-in page is read on its own page: the gate passes only /sign-in there, so the page's four font
  // files are the fixed 404 (an existing property of that host, unchanged here) and their console errors must not reach
  // the shared page's S-ERRORS guard.
  const auth = await context.newPage();
  const authSignIn = await auth.goto(`${AUTH}/sign-in`);
  expect(authSignIn?.status()).toBe(200);
  expect(auth.url()).toBe(`${AUTH}/sign-in`);
  await auth.close();
  for (const url of [`${AUTH}/`, `${NOBODY}/`, `${NOBODY}/sign-in`]) {
    const answer = await statusAt(context, url);
    expect(answer.status, url).toBe(404);
    expect(answer.text).toContain(NOT_FOUND);
  }
});

test("[E-DT-3] after the restart onto Organization B the root is B's entry without a redirect; the phase-1 session is no session there", async ({ page, browser }, testInfo) => {
  onlyInPhase("2");
  for (const o of [ORIGIN_A, ORIGIN_B, WWW]) tracking.allowed.add(o);
  // core: the root now resolves to Organization B (alias precedence over the signup_redirect presentation), never A;
  // www still answers the public entry with the selected presentation.
  const orgA = await organizationOf(ORIGIN_A);
  const orgB = await organizationOf(ORIGIN_B);
  expect(orgB).not.toBe(orgA);
  await expectRootAliasedTo(orgB);
  await expectPublicEntry(WWW, "signup_redirect");

  // The root / is B's tenant entry in place although signup_redirect is selected: no 307 in web's own answer and no
  // redirect hop in the browser; the same workspace name as B's own label host, and B's open enrollment shows Join.
  const plain = await direct("GET", "/", ROOT_HOST);
  expect(plain.status, "GET / on the aliased root").toBe(200);
  expect(plain.location).toBeUndefined();
  const nameB = await expectTenantEntry(page, ORIGIN_B);
  await expect(page.getByTestId("entry-join")).toHaveCount(1);
  const nameRoot = await expectTenantEntry(page, ROOT);
  expect(nameRoot).toBe(nameB);
  expect(nameRoot).not.toBe(SEED_LABEL);
  await expect(page.getByTestId("entry-sign-in")).toHaveCount(1);
  await expect(page.getByTestId("entry-join")).toHaveCount(1);
  expect((await statusAt(page.context(), `${ROOT}/signup`)).status, "/signup stays the fixed 404 on the aliased root").toBe(404);
  // Organization A's own host is unaffected.
  expect(await expectTenantEntry(page, ORIGIN_A)).toBe(SEED_LABEL);

  // The phase-1 root session (bound to the root origin and Organization A) replayed in a fresh profile is no session on
  // the root now resolving to B: the account API is unauthenticated and the account page sends the browser to sign in.
  const handoff = readHandoff(testInfo);
  const replay = await open(browser, testInfo);
  await replay.context.addCookies([{ name: handoff.cookie, value: handoff.live, url: ROOT, secure: true, httpOnly: true, sameSite: "Lax" }]);
  const account = await accountAt(replay.context, ROOT);
  expect(account.status).toBe(401);
  expect(account.body).toEqual({ error: "unauthenticated" });
  const accountPage = await replay.page.goto(`${ROOT}/account`);
  expect(accountPage?.status()).toBe(200);
  const landed = new URL(replay.page.url());
  expect(landed.origin).toBe(ROOT);
  expect(landed.pathname).toBe("/sign-in");
  expect(landed.searchParams.get("return_to")).toBe("/account");
});

test("[E-DT-5] the product default redirects GET and HEAD of exactly the root / to /signup with no-store and nothing forwarded; www and tenants keep their pages", async ({ page }) => {
  onlyInPhase("4");
  for (const o of [WWW, ORIGIN_A]) tracking.allowed.add(o);
  // core: no alias; the root and www answer the public entry with the product default presentation.
  await expectPublicEntry(ROOT, "signup_redirect");
  await expectPublicEntry(WWW, "signup_redirect");

  await expectRootRedirect();
  // www's / and a tenant's / are never redirected.
  for (const host of ["www.localhost:543", "org-a.localhost:543"]) {
    const r = await direct("GET", "/", host);
    expect(r.status, `GET / on ${host}`).toBe(200);
    expect(r.location, `GET / on ${host}`).toBeUndefined();
  }

  // In the browser: one 307 hop to /signup and the signup start page, for / and for / with a query (dropped); nothing
  // started on the way.
  await expectBrowserFollowsToSignup(page, "/");
  await expectBrowserFollowsToSignup(page, "/?utm=1&state=x");
  await expectLanding(page, WWW);
  expect(await expectTenantEntry(page, ORIGIN_A)).toBe(SEED_LABEL);
  // Following the redirect created nothing: core's answers are unchanged.
  await expectPublicEntry(ROOT, "signup_redirect");
});

test("[E-DT-6] an unknown slug and an invalid presentation fall back to the default: the root redirects to /signup and the signup flow is today's", async ({ page }) => {
  onlyInPhase("5");
  tracking.allowed.add(WWW);
  // core: the unknown slug pins nothing (the root stays the public entry) and the invalid value falls back to
  // signup_redirect.
  await expectPublicEntry(ROOT, "signup_redirect");
  await expectPublicEntry(WWW, "signup_redirect");
  await expectRootRedirect();
  await expectBrowserFollowsToSignup(page, "/");
  await expectLanding(page, WWW);
  // Without a signup credential the details page sends the browser back to /signup (E-SUP-1's last step).
  await page.goto(`${ROOT}/signup/details`);
  await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup");
  await expect(page.getByTestId("signup-start")).toHaveCount(1);
});

// The run names a phase this project serves; otherwise every scenario above would be skipped silently.
test("[E-DT-PHASE] the private stack names its core configuration phase", async () => {
  expect(PROJECT_PHASES, "DEFAULT_TENANT_PHASE must name a default-tenant-light phase").toContain(PHASE);
});
