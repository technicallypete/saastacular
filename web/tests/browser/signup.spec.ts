import { randomBytes } from "node:crypto";
import { readFileSync } from "node:fs";

import AxeBuilder from "@axe-core/playwright";
import {
  request as playwrightRequest,
  type APIRequestContext,
  type APIResponse,
  type Browser,
  type BrowserContext,
  type Cookie,
  type Locator,
  type Page,
  type Response,
  type Route,
  type TestInfo,
} from "@playwright/test";

import { expectAxeClean } from "./support/contrast";
import { expect, test } from "./support/page-errors";
import { observe } from "./support/sanitize";

// G062 root-host tenant signup (Plan r5 §1, §4.4, SUP-15 E-SUP-1..6; Codex Round 14 option A and Round 15 decision 2) in
// the disposable HTTPS identity environment (infra/compose/identity.yml), projects signup-light and signup-dark with the
// base https://localhost:543 (the platform root; PLATFORM_BASE_DOMAIN=localhost, PLATFORM_HOST_PORT=543). Every step goes
// through the real BFF routes to core's /internal/v1/auth and /internal/v1/signup routes and the pinned emulator; nothing
// is stubbed. A first-time creator is created per test through the emulator's API; the existing-identity case reuses the
// seeded org-a member alice. Each test uses its own random slug, so both projects share the database. G057 (Plan r5
// §3.3, §5.2 T8): the slug is metadata; a newly created tenant's host is the generated platform label core issues
// (<label>.localhost:543, adjective-noun-four-digit), learned from the create's 303 continuation and never built from
// the slug, and the typed slug's host stays the fixed 404 without a redirect. The label host is reachable at once
// through the edge's *.localhost pattern and Chromium's own *.localhost resolution, with no DNS or CNAME step
// (fixture-proven only; real wildcard ingress/TLS is G057 Part B).
//
// The continuation (Plan r5 §1.4): create (root POST) 303 → tenant /signup/continue?ref=R → root /signup/prove?n=N →
// tenant /signup/complete?p=P → emulator → tenant /auth/callback → /?status=tenant_created. Playwright never routes a
// redirect hop, so a test that must stop the chain holds the 303 of a routable request (a form POST or a page.goto):
// hold() replays that exact request from a driver to web's own port (as identity.spec.ts drives web directly; the Node
// client trusts no run certificate and resolves no *.localhost name, while the TLS edge adds nothing but the Host header
// web already reads), applies the answer's Set-Cookie to the browser context as the browser would have stored them, and
// shows the page a placeholder. The reference, nonce, proof, the three signup cookies, the login binding, CSRF tokens and
// the organization-name canaries are registered with observe(), so B-IDENTITY-PRIVACY scans every core, replica, web
// and proxy log and the browser evidence for them (AC9); a secret's shape is asserted without printing its value. Pages
// that answer an error status are opened in extra pages, so each test's default page stays under S-ERRORS.
//
// The pinned emulator runs in its interactive mode (--interactive-password): every authorize renders the hosted prompt,
// whether or not the browser signed in there before, so the continuation's final hop always reaches the prompt here and
// the creator signs in again before the pinned callback. Fixture-proven here vs G057-qualified (Plan r5 §5.6): the
// prompt path; a real provider's silent return while its session is live, real WorkOS behaviour, real wildcard DNS/TLS
// and production edge redaction are G057's.

const ROOT = "https://localhost:543";
const WWW = "https://www.localhost:543";
const ORIGIN_A = "https://org-a.localhost:543";
const ATTACKER = "https://attacker.localhost:543";
const EMULATOR = "http://emulator:4100";
const CORE = "http://core:8180";
// web's own port in the shared network namespace (the TLS edge proxies to it with the Host header as sent).
const WEB_DIRECT = "http://127.0.0.1:3000";
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const AXE_TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];
const SIGNUP_COOKIE = "__Host-signup";
const FORM_COOKIE = "__Host-signup-form";
const CHALLENGE_COOKIE = "__Host-sct";
const LOGIN_COOKIE = "__Host-login";
const SESSION_COOKIE = "__Host-sid";
const HOPS = new Set(["/signup/continue", "/signup/prove", "/signup/complete"]);
const READY = "Your organization is ready.";
const NOTICE = {
  continueFailed: "We could not take you to your organization automatically. Nothing was lost.",
  intentUsed: "Your organization was already created with different details. Nothing was changed.",
  retryUnavailable: "Your sign-up details are no longer available in this browser.",
  slugTaken: "That short name is already taken. Choose another one.",
  invalidSlug:
    "Use lowercase letters, digits and hyphens, starting and ending with a letter or digit, up to 63 characters. Some short names are reserved.",
  abandoned: "You stopped signing up. Nothing more was created.",
};

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
const BOB = seeded("bob");

// A registered canary (scanned for in every log and in the browser evidence).
function canary(): string {
  const value = `Cx${randomBytes(10).toString("hex")}`;
  observe(value);
  return value;
}

// One signup's subject: a fresh slug (unique per test and project) and an organization name carrying a registered
// canary, so the name never appearing in a log or URL is checked by the scanners and by urlsClean(). The tenant origin is
// empty until core's continuation names it (learnTenant); continues collects the create 303 Locations the browser saw.
type Intent = { name: string; mark: string; slug: string; tenant: string; continues: string[] };
function intent(testInfo: TestInfo): Intent {
  const slug = `s${testInfo.project.name.endsWith("dark") ? "d" : "l"}${randomBytes(4).toString("hex")}`;
  const mark = canary();
  return { name: `Works ${mark}`, mark, slug, tenant: "", continues: [] };
}

// The typed slug's host: metadata only since G057, it routes nowhere.
function slugOrigin(i: Intent): string {
  return `https://${i.slug}.localhost:543`;
}

// G057 (Plan r5 §3.3, §5.2 T8; D-5): the new tenant's origin is learned from core's continuation (a create's 303
// Location), never built from the slug. It is a generated platform label directly under the platform root on the root's
// port and never the typed slug's host; every later continuation of the same intent (a held or re-minted create,
// already_created) names that same stored origin.
const LABEL_ORIGIN = /^https:\/\/[a-z]+-[a-z]+-[0-9]{4}\.localhost:543$/;
function learnTenant(i: Intent, location: string) {
  const u = new URL(location);
  expect(u.pathname, "the create answers into the continuation").toBe("/signup/continue");
  expect(u.origin, "the continuation's origin is a generated label host").toMatch(LABEL_ORIGIN);
  expect(u.origin, "never the typed slug's host").not.toBe(slugOrigin(i));
  if (i.tenant === "") {
    i.tenant = u.origin;
    tracking.allowed.add(u.origin);
  } else {
    expect(u.origin, "a re-minted continuation names the stored origin").toBe(i.tenant);
  }
}

// Collects the 303 Locations of this context's creates that answer into a continuation (a held create is fulfilled with
// a placeholder and seen here not at all; its Location goes to learnTenant directly). Only recorded here: the
// assertions run in the test's own flow (tenantOf).
function watchCreates(context: BrowserContext, i: Intent) {
  context.on("response", (response: Response) => {
    const u = new URL(response.url());
    if (u.origin !== ROOT || u.pathname !== "/signup/create" || response.status() !== 303) return;
    const location = response.headers()["location"] ?? "";
    if (location.startsWith("https://")) i.continues.push(location);
  });
}

// The tenant origin once core's continuation has named it (and every create continuation seen since agrees with it).
async function tenantOf(i: Intent): Promise<string> {
  await expect.poll(() => i.tenant !== "" || i.continues.length > 0, { message: "a create answered into the continuation" }).toBe(true);
  for (const location of i.continues.splice(0)) learnTenant(i, location);
  return i.tenant;
}

// --- page traffic (every request URL is kept), extra contexts and drivers ---------------------------------------------
type Tracking = { urls: string[]; allowed: Set<string>; contexts: BrowserContext[]; drivers: APIRequestContext[] };
let tracking: Tracking;

function track(context: BrowserContext) {
  context.on("request", (request) => {
    const url = request.url();
    if (/^https?:/.test(url)) tracking.urls.push(url);
  });
}

test.beforeEach(async ({ context }) => {
  tracking = { urls: [], allowed: new Set([ROOT, WWW, EMULATOR]), contexts: [], drivers: [] };
  track(context);
});

test.afterEach(async () => {
  for (const c of tracking.contexts) await c.close();
  for (const d of tracking.drivers) await d.dispose();
  const outside = [...new Set(tracking.urls.map((u) => new URL(u).origin))].filter((o) => !tracking.allowed.has(o));
  expect(outside, "page requests outside the app origins under test and the emulator").toEqual([]);
});

async function open(browser: Browser, testInfo: TestInfo): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ baseURL: ROOT, colorScheme: testInfo.project.use.colorScheme, viewport: testInfo.project.use.viewport });
  track(context);
  tracking.contexts.push(context);
  return { context, page: await context.newPage() };
}

async function driver(): Promise<APIRequestContext> {
  const d = await playwrightRequest.newContext();
  tracking.drivers.push(d);
  return d;
}

// No organization name or slug in the path or query of any request URL, and no request carried a name or slug parameter.
// The tenant's own origin (for example the emulator's redirect_uri https://<label>.localhost:543/auth/callback, which
// core builds from the resolved host) is removed before the check. G057: the slug is no host label any more, so it may
// not appear at all, not even as a host inside a URL's path or query.
function urlsClean(i: Intent) {
  for (const raw of tracking.urls) {
    const u = new URL(raw);
    let rest = `${u.pathname}${u.search}`;
    try {
      rest = decodeURIComponent(rest);
    } catch {
      // keep the raw form
    }
    if (i.tenant !== "") rest = rest.split(i.tenant).join("");
    expect(rest, raw).not.toContain(i.slug);
    expect(rest, raw).not.toContain(i.mark);
    expect(u.searchParams.has("name") || u.searchParams.has("slug"), raw).toBe(false);
  }
}

// --- the emulator ------------------------------------------------------------------------------------------------------
async function emulatorAdmin() {
  const d = await driver();
  const headers = { Authorization: `Bearer ${SEED.emulator_api_key}` };
  return {
    async create(user: Credentials & { first_name: string; last_name: string }): Promise<string> {
      const r = await d.post(`${EMULATOR}/user_management/users`, { headers, data: { ...user, email_verified: true } });
      expect(r.status(), "emulator user creation").toBe(201);
      return ((await r.json()) as { id: string }).id;
    },
  };
}

// A first-time creator: a new emulator user with no account in the application.
async function freshUser(): Promise<Credentials> {
  const tag = randomBytes(6).toString("hex");
  const who = { email: `signup-${tag}@identity.test`, password: `Pw-${randomBytes(18).toString("base64url")}` };
  observe(who.password);
  await (await emulatorAdmin()).create({ ...who, first_name: "Signup", last_name: `Creator ${tag}` });
  return who;
}

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

// --- core lookups (driver; never a page) -------------------------------------------------------------------------------
// The Organization a host resolves to: the label's one platform mapping (unique), so an unchanged ID across a retry shows
// that no second Organization exists for this intent (the database-level counts are TestSignupProvisioning's).
async function organizationOf(origin: string): Promise<string> {
  const d = await driver();
  const r = await d.get(`${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(origin)}`);
  expect(r.status(), `hosts/resolve ${origin}`).toBe(200);
  const body = (await r.json()) as { kind: string; organization_id?: string };
  expect(body.kind).toBe("platform");
  expect(body.organization_id).toMatch(UUID);
  return body.organization_id!;
}

// --- cookies -----------------------------------------------------------------------------------------------------------
// The cookies the browser would send with a request to url, by RFC 6265 domain matching: a host-only cookie (Chromium
// reports its domain as the bare host) only to exactly that host; a domain cookie (a leading dot) to that domain and its
// subdomains. context.cookies(url) alone is wider: it matches every stored domain as a suffix, so the root's host-only
// localhost cookies are listed for https://<label>.localhost although Chromium never sends them there (the second run
// listed the root's __Host-signup and __Host-login under the tenant URL, and cookieAt picked the root's binding as the
// tenant's). Every assertion about where a cookie lives reads through this filter, and hold() builds the replayed Cookie
// header from it. context.cookies(url) still applies its own path and Secure checks.
async function cookiesSentTo(context: BrowserContext, url: string): Promise<Cookie[]> {
  const host = new URL(url).hostname;
  return (await context.cookies(url)).filter((c) => (c.domain.startsWith(".") ? `.${host}`.endsWith(c.domain) : c.domain === host));
}

async function cookieAt(context: BrowserContext, origin: string, name: string): Promise<Cookie | undefined> {
  const found = (await cookiesSentTo(context, origin)).find((c) => c.name === name);
  observe(found?.value);
  return found;
}

// A host-only, Path=/, HttpOnly, SameSite=Lax, Secure cookie on exactly this origin's host with a lifetime of at most
// maxAge seconds (a session cookie never: each signup cookie carries Max-Age). The domain check is the positive control
// of cookiesSentTo's premise: a domain cookie (.localhost) would be selected for the root too and fail here.
function expectHostOnly(c: Cookie | undefined, origin: string, maxAge: number) {
  expect(c, "cookie present").toBeDefined();
  expect(c!.domain).toBe(new URL(origin).hostname);
  expect(c!.path).toBe("/");
  expect(c!.httpOnly).toBe(true);
  expect(c!.secure).toBe(true);
  expect(c!.sameSite).toBe("Lax");
  expect(c!.expires).toBeGreaterThan(Date.now() / 1000);
  expect(c!.expires).toBeLessThanOrEqual(Date.now() / 1000 + maxAge + 5);
}

// The root signup cookies are never sent to the tenant host, and the tenant cookies never to the root; and no signup,
// login or session cookie on the root and tenant hosts is a domain cookie (one would reach every subdomain, which the
// two per-host checks could not tell from a host-only cookie of the same name). The unfiltered URL selection is the
// right reader here: it keeps every cookie whose domain covers either host, dotted or not (so a .localhost cookie set
// by either host is listed for both).
const OWN_COOKIES = new Set([SIGNUP_COOKIE, FORM_COOKIE, CHALLENGE_COOKIE, LOGIN_COOKIE, SESSION_COOKIE]);
async function expectCookiesWhereExpected(context: BrowserContext, tenant: string) {
  const onTenant = (await cookiesSentTo(context, tenant)).map((c) => c.name);
  expect(onTenant).not.toContain(SIGNUP_COOKIE);
  expect(onTenant).not.toContain(FORM_COOKIE);
  const onRoot = (await cookiesSentTo(context, ROOT)).map((c) => c.name);
  expect(onRoot).not.toContain(SESSION_COOKIE);
  expect(onRoot).not.toContain(CHALLENGE_COOKIE);
  for (const c of await context.cookies([ROOT, tenant])) {
    if (OWN_COOKIES.has(c.name)) expect(c.domain, `${c.name} on ${c.domain} is host-only`).not.toMatch(/^\./);
  }
}

// The form restoration record for a name and slug, exactly as web writes it (v1.<base64url name>.<slug>).
function formRecord(name: string, slug: string): string {
  return `v1.${Buffer.from(name, "utf8").toString("base64url")}.${slug}`;
}

// --- the continuation hops ---------------------------------------------------------------------------------------------
// Every continuation hop answered in this context: a 303 with no-store and no-referrer; the secrets in its request URL
// and Location are registered (they are never logged).
type Hop = { path: string; status: number; cacheControl: string; referrerPolicy: string; location: string };
function watchHops(context: BrowserContext): Hop[] {
  const hops: Hop[] = [];
  context.on("response", (response: Response) => {
    const u = new URL(response.url());
    if (!HOPS.has(u.pathname)) return;
    const h = response.headers();
    const location = h["location"] ?? "";
    for (const p of ["ref", "n", "p"]) {
      observe(u.searchParams.get(p));
      if (location.startsWith("https://")) observe(new URL(location).searchParams.get(p));
    }
    hops.push({ path: u.pathname, status: response.status(), cacheControl: h["cache-control"] ?? "", referrerPolicy: h["referrer-policy"] ?? "", location });
  });
  return hops;
}

function expectHopsSealed(hops: Hop[], tenant: string) {
  expect(hops.map((h) => h.path)).toEqual(["/signup/continue", "/signup/prove", "/signup/complete"]);
  for (const h of hops) {
    expect(h.status, h.path).toBe(303);
    expect(h.cacheControl, h.path).toContain("no-store");
    expect(h.referrerPolicy, h.path).toBe("no-referrer");
  }
  expect(new URL(hops[0].location).origin).toBe(ROOT);
  expect(new URL(hops[0].location).pathname).toBe("/signup/prove");
  expect(new URL(hops[1].location).origin).toBe(tenant);
  expect(new URL(hops[1].location).pathname).toBe("/signup/complete");
  expect(new URL(hops[2].location).origin).toBe(EMULATOR);
}

// A 43-character secret, asserted without printing the value (a failed assertion's message is saved in the report).
function expectSecret(value: string | null | undefined, label: string) {
  observe(value);
  expect(typeof value === "string" && SECRET.test(value), `${label} is a 43-character secret`).toBe(true);
}

// Holds the 303 of the next request matching `match` (a form POST or a page.goto, never a redirect hop) without following
// it; returns its Location and leaves a placeholder page. The held request is replayed from a driver straight to web's
// own port with the browser's method, Host, headers, cookies (the context's store for that URL) and body; its answer's
// Set-Cookie headers are applied to the browser context (applySetCookies) and the Location's continuation secret is
// registered. A transport failure is reported by its first line only (Playwright's call log would print the headers).
async function hold(page: Page, match: (u: URL, method: string) => boolean, act: () => Promise<unknown>): Promise<string> {
  let location = "";
  const context = page.context();
  const handler = async (route: Route) => {
    const req = route.request();
    const u = new URL(req.url());
    if (location !== "" || !match(u, req.method())) return route.continue();
    const headers: Record<string, string> = { ...req.headers(), host: u.host };
    for (const name of ["content-length", "accept-encoding", "cookie"]) delete headers[name];
    const jar = await cookiesSentTo(context, req.url());
    if (jar.length > 0) headers.cookie = jar.map((c) => `${c.name}=${c.value}`).join("; ");
    const d = await driver();
    let response: APIResponse;
    try {
      response = await d.fetch(`${WEB_DIRECT}${u.pathname}${u.search}`, { method: req.method(), headers, data: req.postDataBuffer() ?? undefined, maxRedirects: 0 });
    } catch (error) {
      throw new Error(`held ${req.method()} ${u.pathname}: ${String((error as Error).message).split("\n")[0]}`);
    }
    expect(response.status(), `held ${req.method()} ${u.pathname}`).toBe(303);
    location = response.headers()["location"] ?? "";
    await applySetCookies(context, u.hostname, response.headersArray());
    return route.fulfill({ status: 200, contentType: "text/plain", headers: { "cache-control": "no-store" }, body: "redirect held" });
  };
  await page.route("**/*", handler);
  await act();
  await expect.poll(() => location).not.toBe("");
  await page.unroute("**/*", handler);
  if (location.startsWith("https://")) for (const p of ["ref", "n", "p"]) observe(new URL(location).searchParams.get(p));
  return location;
}

// Applies a held answer's Set-Cookie headers to the context exactly as the browser would store them for host: every
// signup cookie is host-only (no Domain), Path=/, HttpOnly, SameSite=Lax, Secure with a Max-Age (0 clears it). Any
// other shape is a defect of the answering route.
async function applySetCookies(context: BrowserContext, host: string, headers: { name: string; value: string }[]) {
  for (const h of headers) {
    if (h.name.toLowerCase() !== "set-cookie") continue;
    const [pair, ...attributes] = h.value.split(";").map((s) => s.trim());
    const at = pair.indexOf("=");
    expect(at, "a Set-Cookie value starts with name=").toBeGreaterThan(0);
    const name = pair.slice(0, at);
    const value = pair.slice(at + 1);
    observe(value);
    const flags = new Map(
      attributes.map((a) => {
        const eq = a.indexOf("=");
        return eq < 0 ? [a.toLowerCase(), ""] : [a.slice(0, eq).toLowerCase(), a.slice(eq + 1)];
      }),
    );
    expect([...flags.keys()].sort(), `attributes of ${name}`).toEqual(["httponly", "max-age", "path", "samesite", "secure"]);
    expect(flags.get("path"), `Path of ${name}`).toBe("/");
    expect(flags.get("samesite"), `SameSite of ${name}`).toBe("Lax");
    const maxAge = Number(flags.get("max-age"));
    expect(Number.isInteger(maxAge) && maxAge >= 0, `Max-Age of ${name}`).toBe(true);
    await context.clearCookies({ name, domain: host });
    if (maxAge === 0) continue;
    await context.addCookies([{ name, value, domain: host, path: "/", expires: Math.floor(Date.now() / 1000) + maxAge, httpOnly: true, secure: true, sameSite: "Lax" }]);
  }
}

const createPost = (u: URL, method: string) => method === "POST" && u.origin === ROOT && u.pathname === "/signup/create";
const navigationTo = (path: string) => (u: URL, method: string) => method === "GET" && u.pathname === path;

// --- the root signup steps ---------------------------------------------------------------------------------------------
// From the root's /signup through the emulator's hosted page to /signup/details with the signup credential.
async function signupLogin(page: Page, who: Credentials) {
  await page.goto(`${ROOT}/signup`);
  await page.getByTestId("signup-start").click();
  await enterCredentials(page, who);
  await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup/details");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Name your organization");
  expectSecret((await cookieAt(page.context(), ROOT, SIGNUP_COOKIE))?.value, "the signup credential");
  for (const v of await page.locator('input[name="csrf_token"]').evaluateAll((els) => els.map((e) => (e as HTMLInputElement).value))) {
    expectSecret(v, "a rendered CSRF token");
  }
}

// Step 1 (the name; core suggests the short name) and step 2 (the short name, edited to the intent's slug). Stops before
// Create; from here on this context's create continuations are collected for the intent (watchCreates).
async function nameAndAddress(page: Page, i: Intent) {
  watchCreates(page.context(), i);
  const nameForm = page.getByTestId("signup-name-form");
  await nameForm.locator("#field-name").fill(i.name);
  await page.getByTestId("signup-name-submit").click();
  await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup/details" && u.search === "");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Choose its short name");
  const details = page.getByTestId("signup-details-form");
  await expect(details.locator("#field-name")).toHaveValue(i.name);
  await expect(details.locator("#field-slug")).not.toHaveValue("");
  await details.locator("#field-slug").fill(i.slug);
}

async function landedCreated(page: Page, i: Intent) {
  const tenant = await tenantOf(i);
  await page.waitForURL((u) => u.origin === tenant && u.pathname === "/" && u.searchParams.get("status") === "tenant_created");
  await expect(page.getByTestId("status-message")).toHaveText(READY);
}

// The automatic continuation from its first hop to the new tenant: continue → prove → complete → the emulator's hosted
// prompt (its interactive mode prompts on every authorize) → the pinned callback → the tenant entry with the closed
// notice. The creator signs in again as the same user; no application page has a button on this path.
async function continueAtPrompt(page: Page, who: Credentials, i: Intent) {
  await enterCredentials(page, who);
  await landedCreated(page, i);
}

// The recovery view of an already provisioned signup on the root (Codex Round 15 decision 2: a GET never re-mints).
async function expectRecoveryView(page: Page, i: Intent, retry: boolean, notice?: string) {
  expect(new URL(page.url()).origin).toBe(ROOT);
  expect(new URL(page.url()).pathname).toBe("/signup/details");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Your organization is ready");
  await expect(page.getByText(`Your organization was created at ${i.tenant}.`)).toBeVisible();
  await expect(page.getByTestId("signup-retry")).toHaveCount(retry ? 1 : 0);
  await expect(page.getByTestId("signup-tenant-sign-in")).toHaveAttribute("href", `${i.tenant}/sign-in`);
  await expect(page.getByTestId("signup-abandon")).toHaveCount(1);
  if (notice !== undefined) await expect(page.getByTestId("status-message")).toHaveText(notice);
}

// Counts this page's create/re-mint requests and its requests to the tenant while `act` runs.
async function mutationsDuring(page: Page, tenant: string, act: () => Promise<unknown>): Promise<{ creates: number; tenant: number }> {
  const seen = { creates: 0, tenant: 0 };
  const listener = (r: { url(): string; method(): string }) => {
    const u = new URL(r.url());
    if (u.origin === ROOT && u.pathname === "/signup/create") seen.creates += 1;
    if (u.origin === tenant) seen.tenant += 1;
  };
  page.on("request", listener);
  await act();
  page.off("request", listener);
  return seen;
}

// core's per-credential bucket on every root signup route (Plan r5 §4.5; core/internal/signup/api/http.go
// credentialRate/credentialBurst): 5 tokens per second, burst 10, spent by each status read of the details page, the
// slug suggestion, every create (a refused one too), every prove and the abandon. Playwright drives a whole scenario in
// one or two seconds, far faster than a person, so a test that spends more than the burst on one credential lets the
// bucket refill before its last chain instead of meeting 429 rate_limited there (the second run: the captured-secrets
// test's fifteenth call, the re-mint's prove, in both projects). Nothing is retried or relaxed; the budget is core's.
const CREDENTIAL_RATE = 5;
const CREDENTIAL_BURST = 10;
async function credentialBudgetRefilled(page: Page) {
  await page.waitForTimeout((CREDENTIAL_BURST / CREDENTIAL_RATE) * 1000);
}

async function statusAt(context: BrowserContext, url: string): Promise<{ status: number; url: string; text: string }> {
  const p = await context.newPage();
  const r = await p.goto(url);
  const out = { status: r?.status() ?? 0, url: p.url(), text: (await p.locator("body").textContent()) ?? "" };
  await p.close();
  return out;
}

// G057 (Plan r5 §3.3; the amendment's "no slug-host redirect discloses the random address"): after creation the typed
// slug's host is the fixed 404 in place (no redirect hop, no Location in web's own answer) and names neither the label
// nor the tenant origin; core resolves it to nothing, while the label host resolves to the created Organization.
async function expectSlugHostUnknown(context: BrowserContext, i: Intent) {
  const origin = slugOrigin(i);
  const label = new URL(i.tenant).hostname.split(".")[0];
  tracking.allowed.add(origin);
  for (const path of ["/", "/sign-in"]) {
    const answer = await statusAt(context, `${origin}${path}`);
    expect(answer.status, `${origin}${path}`).toBe(404);
    expect(answer.url, "served in place, no redirect").toBe(`${origin}${path}`);
    expect(answer.text).toContain("There is nothing at this address.");
    expect(answer.text).not.toContain(label);
  }
  const d = await driver();
  for (const path of ["/", "/sign-in", "/signup/continue"]) {
    const direct = await d.get(`${WEB_DIRECT}${path}`, { headers: { Host: new URL(origin).host }, maxRedirects: 0 });
    expect(direct.status(), `web ${path} on the slug host`).toBe(404);
    expect(direct.headers()["location"], `no Location for ${path}`).toBeUndefined();
    expect(await direct.text()).not.toContain(label);
  }
  const r = await d.get(`${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(origin)}`);
  expect(r.status(), "the slug host resolves to nothing").toBe(404);
  expect(await r.json()).toEqual({ error: "unknown_host" });
  expect(await organizationOf(i.tenant)).toMatch(UUID);
}

async function tabTo(page: Page, target: Locator, label: string) {
  for (let i = 0; i < 80 && !(await target.evaluate((t) => t === document.activeElement).catch(() => false)); i++) await page.keyboard.press("Tab");
  await expect(target, `keyboard focus on ${label}`).toBeFocused();
}

async function audit(page: Page, label: string) {
  for (const width of [1280, 320]) {
    await page.setViewportSize({ width, height: width === 320 ? 640 : 800 });
    const results = await new AxeBuilder({ page }).withTags(AXE_TAGS).analyze();
    // G069: the kit's white label on its #0075ff primary (the platform root has no tenant palette) and, in Light, an
    // invalid kit Field's input text (text-destructive on the kit fill, the field-error states) are OWNER-WAIVED
    // color-contrast nodes, recorded with their ratios; every other violation fails (support/contrast.ts expectAxeClean).
    expectAxeClean(results.violations, `axe (${width} ${label})`);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), `no horizontal scroll at ${width} (${label})`).toBe(true);
  }
  await page.setViewportSize({ width: 1280, height: 800 });
}

// =======================================================================================================================

test("[E-SUP-1] the root and www show only Sign Up to the root's /signup; no redirect; other pages are the fixed 404", async ({ page, context }) => {
  for (const origin of [ROOT, WWW]) {
    const r = await page.goto(`${origin}/`);
    expect(r?.status()).toBe(200);
    expect(r?.request().redirectedFrom()).toBeNull();
    expect(page.url()).toBe(`${origin}/`);
    await expect(page.getByTestId("root-sign-up")).toHaveCount(1);
    await expect(page.getByTestId("root-sign-up")).toHaveAttribute("href", `${ROOT}/signup`);
    await expect(page.getByRole("link", { name: /sign in|account/i })).toHaveCount(0);
    await expect(page.getByRole("button", { name: /sign in/i })).toHaveCount(0);
    for (const path of ["/account", "/sign-in", "/groups", "/connections/mcp"]) {
      const answer = await statusAt(context, `${origin}${path}`);
      expect(answer.status, `${origin}${path}`).toBe(404);
      expect(answer.text).toContain("There is nothing at this address.");
    }
    expect((await statusAt(context, `${origin}/api/health`)).status).toBe(200);
  }
  expect((await statusAt(context, `${WWW}/signup`)).status).toBe(404);
  expect((await statusAt(context, `${WWW}/signup/details`)).status).toBe(404);
  // Sign Up from www lands on the root's own /signup (canonicalized), which offers exactly one start action.
  await page.goto(`${WWW}/`);
  await page.getByTestId("root-sign-up").click();
  await page.waitForURL(`${ROOT}/signup`);
  await expect(page.getByTestId("signup-start")).toHaveCount(1);
  // Without a signup credential, the details page sends the browser back to /signup.
  await page.goto(`${ROOT}/signup/details`);
  await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup");
});

test("[E-SUP-2] a first-time identity creates a tenant and continues into it automatically; cookies stay on their hosts", async ({ page, context }, testInfo) => {
  const who = await freshUser();
  const i = intent(testInfo);
  const hops = watchHops(context);
  await signupLogin(page, who);
  expectHostOnly(await cookieAt(context, ROOT, SIGNUP_COOKIE), ROOT, 1800);
  await nameAndAddress(page, i);
  await page.getByTestId("signup-create").click();
  // The happy path has no button: create → continue → prove → complete → the provider's prompt → the pinned callback →
  // the tenant entry with the closed notice.
  await continueAtPrompt(page, who, i);
  expectHopsSealed(hops, i.tenant);
  expect(await organizationOf(i.tenant)).toMatch(UUID);
  // G057: the tenant is its generated label host; the typed slug's host routes nowhere and discloses nothing.
  await expectSlugHostUnknown(context, i);
  // The session exists only on the new tenant host; the root keeps its own signup cookies (r4-2) and never a session.
  expectHostOnly(await cookieAt(context, i.tenant, SESSION_COOKIE), i.tenant, 24 * 3600);
  expect((await cookiesSentTo(context, ROOT)).map((c) => c.name)).not.toContain(SESSION_COOKIE);
  expectHostOnly(await cookieAt(context, ROOT, SIGNUP_COOKIE), ROOT, 1800);
  // The form restoration record (Codex Round 14 option A): exactly the submitted values, host-only on the root, ≤ 30 min.
  const record = await cookieAt(context, ROOT, FORM_COOKIE);
  expectHostOnly(record, ROOT, 1800);
  expect(record!.value).toBe(formRecord(i.name, i.slug));
  await expectCookiesWhereExpected(context, i.tenant);
  expect((await cookiesSentTo(context, i.tenant)).map((c) => c.name)).not.toContain(CHALLENGE_COOKIE);
  urlsClean(i);
});

test("[E-SUP-3] an existing identity creates a second tenant; the same account, and its other memberships unchanged", async ({ page }, testInfo) => {
  tracking.allowed.add(ORIGIN_A);
  // alice's existing account and membership on org-a, through the ordinary tenant sign-in.
  await page.goto(`${ORIGIN_A}/sign-in?return_to=/account`);
  await page.getByRole("button", { name: "Sign in" }).click();
  await enterCredentials(page, ALICE);
  await page.waitForURL((u) => u.origin === ORIGIN_A && u.pathname === "/account");
  const before = await accountAt(page, ORIGIN_A);
  expect(before.membership).toBe("member");
  // Root signup with the same identity: the signup login's own hosted prompt (the emulator prompts on every authorize;
  // alice signs in again), then the details page with a signup credential.
  const i = intent(testInfo);
  await signupLogin(page, ALICE);
  await nameAndAddress(page, i);
  await page.getByTestId("signup-create").click();
  await continueAtPrompt(page, ALICE, i);
  const created = await accountAt(page, i.tenant);
  expect(created.userId, "the same account (mapping by issuer and subject, never a new user)").toBe(before.userId);
  expect(created.membership).toBe("member");
  const after = await accountAt(page, ORIGIN_A);
  expect(after).toEqual(before);
  expect(await organizationOf(i.tenant)).not.toBe(await organizationOf(ORIGIN_A));
  urlsClean(i);
});

// The signed-in account on ORIGIN: /api/account's user ID (a same-origin fetch from a helper page of this page's context;
// Chromium resolves *.localhost and trusts the pinned key, the Node client does neither) and the tenant frame's
// membership indicator on its entry page (G017 D-6: "member" or "none"; an indicator fact that grants nothing).
async function accountAt(page: Page, origin: string): Promise<{ userId: string; membership: string | null }> {
  const p = await page.context().newPage();
  await p.goto(`${origin}/api/health`);
  const r = await p.evaluate(async () => {
    const res = await fetch("/api/account", { cache: "no-store" });
    return { status: res.status, body: (await res.json()) as Record<string, unknown> };
  });
  expect(r.status, `/api/account on ${origin}`).toBe(200);
  expect(r.body.id).toMatch(UUID);
  await p.goto(`${origin}/`);
  const membership = await p.getByTestId("context-indicator").getAttribute("data-membership");
  await p.close();
  return { userId: r.body.id as string, membership };
}

test("[E-SUP-4] address refusals are field errors; abandon clears the signup; keyboard and axe on every signup page", async ({ page, context }, testInfo) => {
  test.slow();
  await page.goto(`${ROOT}/`);
  await audit(page, "root landing");
  await page.goto(`${WWW}/`);
  await audit(page, "www landing");
  await page.goto(`${ROOT}/signup`);
  await audit(page, "signup start");
  // Keyboard only: the landing's Sign Up, then the start action, reach the hosted page.
  await page.goto(`${ROOT}/`);
  await tabTo(page, page.getByTestId("root-sign-up"), "Sign Up");
  await page.keyboard.press("Enter");
  await page.waitForURL(`${ROOT}/signup`);
  await tabTo(page, page.getByTestId("signup-start"), "Continue to sign up");
  await page.keyboard.press("Enter");
  const who = await freshUser();
  await enterCredentials(page, who);
  await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup/details");
  await audit(page, "details step 1");
  const i = intent(testInfo);
  await nameAndAddress(page, i);
  await audit(page, "details step 2");
  // Taken (a seeded Organization's slug: Organization A's friendly slug alpha-agency, whose platform label is org-a,
  // infra/identity/mcp-seed.sql, G057 D-2), reserved and malformed short names: closed field errors, values restored from
  // the browser's own record, nothing created and no value in any URL.
  for (const [slug, message] of [
    ["alpha-agency", NOTICE.slugTaken],
    ["www", NOTICE.invalidSlug],
    ["auth", NOTICE.invalidSlug],
    ["Bad_Slug", NOTICE.invalidSlug],
  ] as const) {
    const details = page.getByTestId("signup-details-form");
    await details.locator("#field-slug").fill(slug);
    await tabTo(page, page.getByTestId("signup-create"), "Create organization");
    await page.keyboard.press("Enter");
    await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup/details" && u.searchParams.has("status"));
    await expect(page.getByTestId("signup-details-form").getByTestId("field-error")).toHaveText(message);
    await expect(page.getByTestId("signup-details-form").locator("#field-name")).toHaveValue(i.name);
    expect(page.url()).not.toContain(encodeURIComponent(i.name));
    await audit(page, `field error ${slug}`);
  }
  // Abandon: core revokes the credential; both root signup cookies are cleared; nothing was created.
  await page.getByTestId("signup-abandon").click();
  await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup" && u.searchParams.get("status") === "abandoned");
  await expect(page.getByTestId("status-message")).toHaveText(NOTICE.abandoned);
  const names = (await cookiesSentTo(context, ROOT)).map((c) => c.name);
  expect(names).not.toContain(SIGNUP_COOKIE);
  expect(names).not.toContain(FORM_COOKIE);
  // Nothing was created: no create answered into a continuation (every one was refused), so no tenant origin was ever
  // named; the typed slug's host resolves to nothing either way.
  expect(i.continues, "an abandoned signup created no tenant").toEqual([]);
  expect(i.tenant).toBe("");
  const d = await driver();
  const r = await d.get(`${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(slugOrigin(i))}`);
  expect(r.status(), "the slug host resolves to nothing").toBe(404);
  urlsClean(i);
});

test("[E-SUP-5] a GET of the recovery page never re-mints: a held continuation stays valid after direct and cross-site GETs", async ({ page, context }, testInfo) => {
  tracking.allowed.add(ATTACKER);
  const who = await freshUser();
  const i = intent(testInfo);
  await signupLogin(page, who);
  await nameAndAddress(page, i);
  // The create's 303 is held: the Organization exists and generation 1's reference R is unused.
  const continuation = await hold(page, createPost, () => page.getByTestId("signup-create").click());
  learnTenant(i, continuation);
  expect(new URL(continuation).origin).toBe(i.tenant);
  expect(new URL(continuation).pathname).toBe("/signup/continue");
  expectSecret(new URL(continuation).searchParams.get("ref"), "the held reference");
  const org = await organizationOf(i.tenant);
  // Direct GET: the recovery view, with Try again (the record names the created address), no create and no tenant hop.
  const direct = await mutationsDuring(page, i.tenant, async () => {
    await page.goto(`${ROOT}/signup/details`);
    await expectRecoveryView(page, i, true);
  });
  expect(direct).toEqual({ creates: 0, tenant: 0 });
  // Cross-site top-level navigation from the attack origin (it carries the SameSite=Lax root cookies), in an extra page of
  // the same browser: the same view, nothing mutated.
  const xs = await context.newPage();
  const crossSite = await mutationsDuring(xs, i.tenant, async () => {
    await xs.goto(`${ATTACKER}/`);
    await xs.evaluate((url) => {
      window.location.href = url;
    }, `${ROOT}/signup/details?status=continue_failed`);
    await xs.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup/details");
    await expectRecoveryView(xs, i, true, NOTICE.continueFailed);
  });
  await xs.close();
  expect(crossSite).toEqual({ creates: 0, tenant: 0 });
  // No generation change: the held reference R of generation 1 still completes the automatic continuation.
  const hops = watchHops(context);
  await page.goto(continuation);
  await continueAtPrompt(page, who, i);
  expectHopsSealed(hops, i.tenant);
  expect(await organizationOf(i.tenant)).toBe(org);
  urlsClean(i);
});

test("[E-SUP-5] captured continuation secrets cannot complete elsewhere; Try again re-mints and recovers with one Organization", async ({ page, context, browser }, testInfo) => {
  test.slow();
  const who = await freshUser();
  const i = intent(testInfo);
  await signupLogin(page, who);
  await nameAndAddress(page, i);
  const r1 = await hold(page, createPost, () => page.getByTestId("signup-create").click());
  learnTenant(i, r1);
  const org = await organizationOf(i.tenant);
  // Captured R opened in a second browser: R is spent there (the only consumption); that browser gets its own challenge
  // cookie (host-only on the tenant, ≤ 120 s) and nonce, but cannot prove without the original root signup credential.
  const other = await open(browser, testInfo);
  const prove2 = await hold(other.page, navigationTo("/signup/continue"), () => other.page.goto(r1));
  expect(new URL(prove2).origin).toBe(ROOT);
  expect(new URL(prove2).pathname).toBe("/signup/prove");
  expectHostOnly(await cookieAt(other.context, i.tenant, CHALLENGE_COOKIE), i.tenant, 120);
  await other.page.goto(prove2);
  await other.page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup");
  expect(await cookieAt(other.context, i.tenant, SESSION_COOKIE)).toBeUndefined();
  expect(await cookieAt(other.context, ROOT, SIGNUP_COOKIE)).toBeUndefined();
  // Captured N opened in the original browser: it proves with its own credential (a legitimate match) and carries P to
  // the tenant, where it lacks the other browser's challenge cookie: refused, nothing consumed, recovery page.
  const complete1 = await hold(page, navigationTo("/signup/prove"), () => page.goto(prove2));
  expect(new URL(complete1).origin).toBe(i.tenant);
  expect(new URL(complete1).pathname).toBe("/signup/complete");
  await page.goto(complete1);
  await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup/details" && u.searchParams.get("status") === "continue_failed");
  await expectRecoveryView(page, i, true, NOTICE.continueFailed);
  expect(await cookieAt(other.context, i.tenant, SESSION_COOKIE)).toBeUndefined();
  // The original browser's own continuation finds R spent: the same recovery page.
  await page.goto(r1);
  await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup/details" && u.searchParams.get("status") === "continue_failed");
  // A tampered restoration record (another name, the same address): Try again is offered (the address matches) but core's
  // intent digest refuses it (intent_used); nothing is created or rewritten.
  const original = (await cookieAt(context, ROOT, FORM_COOKIE))!.value;
  expect(original).toBe(formRecord(i.name, i.slug));
  const tampered = formRecord(`Changed ${canary()}`, i.slug);
  await context.addCookies([{ name: FORM_COOKIE, value: tampered, url: ROOT, secure: true, httpOnly: true, sameSite: "Lax" }]);
  await page.goto(`${ROOT}/signup/details?status=continue_failed`);
  await page.getByTestId("signup-retry").click();
  await page.waitForURL((u) => u.origin === ROOT && u.pathname === "/signup/details" && u.searchParams.get("status") === "intent_used");
  await expect(page.getByTestId("status-message")).toHaveText(NOTICE.intentUsed);
  expect((await cookieAt(context, ROOT, FORM_COOKIE))!.value, "a retry never rewrites the record").toBe(tampered);
  expect(await organizationOf(i.tenant)).toBe(org);
  // A stale record (another address, or none at all as after its expiry): no Try again and no create request.
  for (const value of [formRecord(i.name, `${i.slug}x`), undefined]) {
    await context.clearCookies({ name: FORM_COOKIE });
    if (value !== undefined) await context.addCookies([{ name: FORM_COOKIE, value, url: ROOT, secure: true, httpOnly: true, sameSite: "Lax" }]);
    const seen = await mutationsDuring(page, i.tenant, async () => {
      await page.goto(`${ROOT}/signup/details?status=continue_failed`);
      await expectRecoveryView(page, i, false, NOTICE.continueFailed);
    });
    expect(seen).toEqual({ creates: 0, tenant: 0 });
  }
  // The exact record restored: Try again repeats the identical payload (already_created, generation 2) and the automatic
  // continuation completes; every secret of generation 1 is dead and exactly one Organization exists for the intent.
  // Thirteen calls have spent this credential's bucket (nine status reads, the suggestion, two creates, one prove); the
  // chain below needs three more and the replays after it another three, so the budget refills first.
  await credentialBudgetRefilled(page);
  await context.addCookies([{ name: FORM_COOKIE, value: original, url: ROOT, secure: true, httpOnly: true, sameSite: "Lax" }]);
  await page.goto(`${ROOT}/signup/details?status=continue_failed`);
  const hops = watchHops(context);
  await page.getByTestId("signup-retry").click();
  await continueAtPrompt(page, who, i);
  expectHopsSealed(hops, i.tenant);
  expect(await organizationOf(i.tenant)).toBe(org);
  await expectCookiesWhereExpected(context, i.tenant);
  // Replays after use: R and N of generation 1 refused; the captured P of generation 1 in the other browser refused.
  for (const replay of [r1, prove2]) {
    const answer = await statusAt(context, replay);
    expect(new URL(answer.url).origin, replay).toBe(ROOT);
    expect(new URL(answer.url).pathname).toBe("/signup/details");
  }
  await other.page.goto(complete1);
  await other.page.waitForURL((u) => u.origin === ROOT && (u.pathname === "/signup" || u.pathname === "/signup/details"));
  expect(await cookieAt(other.context, i.tenant, SESSION_COOKIE)).toBeUndefined();
  urlsClean(i);
});

test("[E-SUP-5] the final hop is pinned to the creator: another account is refused; the creator recovers; a dropped create response continues", async ({ page, context, browser }, testInfo) => {
  test.slow();
  const who = await freshUser();
  const i = intent(testInfo);
  await signupLogin(page, who);
  await nameAndAddress(page, i);
  // A dropped create response: the browser never saw the 303. The details page shows the recovery view (no re-mint on
  // GET); Try again re-posts the identical payload and continues. Each hop is held to observe it.
  learnTenant(i, await hold(page, createPost, () => page.getByTestId("signup-create").click()));
  await page.goto(`${ROOT}/signup/details`);
  await expectRecoveryView(page, i, true);
  // The re-minted continuation (already_created) names the stored origin, not a new label (learnTenant asserts it).
  const r2 = await hold(page, createPost, () => page.getByTestId("signup-retry").click());
  learnTenant(i, r2);
  const prove = await hold(page, navigationTo("/signup/continue"), () => page.goto(r2));
  expectHostOnly(await cookieAt(context, i.tenant, CHALLENGE_COOKIE), i.tenant, 120);
  const complete = await hold(page, navigationTo("/signup/prove"), () => page.goto(prove));
  const authorize = await hold(page, navigationTo("/signup/complete"), () => page.goto(complete));
  expect(new URL(authorize).origin).toBe(EMULATOR);
  expect(new URL(authorize).searchParams.get("redirect_uri")).toBe(`${i.tenant}/auth/callback`);
  observe(new URL(authorize).searchParams.get("state"));
  expect(await cookieAt(context, i.tenant, CHALLENGE_COOKIE), "the challenge cookie is cleared at completion").toBeUndefined();
  const binding = await cookieAt(context, i.tenant, LOGIN_COOKIE);
  expectHostOnly(binding, i.tenant, 600);
  // Account substitution: another browser holding only a copy of the tenant login binding signs in as another seeded
  // user at the final hop: signup_user_mismatch (consumed); no session anywhere.
  const other = await open(browser, testInfo);
  await other.context.addCookies([{ name: LOGIN_COOKIE, value: binding!.value, url: i.tenant, secure: true, httpOnly: true, sameSite: "Lax" }]);
  await other.page.goto(authorize);
  await enterCredentials(other.page, BOB);
  await other.page.waitForURL((u) => u.origin === i.tenant && u.pathname === "/sign-in" && u.searchParams.get("status") === "failed");
  expect(await cookieAt(other.context, i.tenant, SESSION_COOKIE)).toBeUndefined();
  expect(await cookieAt(context, i.tenant, SESSION_COOKIE)).toBeUndefined();
  // The creator recovers on the root: the recovery view's Try again re-mints and completes into the same Organization.
  const org = await organizationOf(i.tenant);
  await page.goto(`${ROOT}/signup/details`);
  await expectRecoveryView(page, i, true);
  await page.getByTestId("signup-retry").click();
  await continueAtPrompt(page, who, i);
  expect(await organizationOf(i.tenant)).toBe(org);
  urlsClean(i);
});

test("[E-SUP-5] the hosted prompt on the way to the tenant: the same user signs in again; a cancel at the prompt leaves the tenant intact", async ({ page, browser }, testInfo) => {
  test.slow();
  // (a) The continuation never issues a session itself: its final hop sends the creator to the provider, whose prompt
  // appears (no Continue button anywhere in the application); signing in again as the same user (a new provider
  // session) is accepted. The emulator prompts on every authorize, so this is the fixture's only continuation path; a
  // real provider's silent return while its session is live is G057's.
  const who = await freshUser();
  const i = intent(testInfo);
  await signupLogin(page, who);
  await nameAndAddress(page, i);
  const r1 = await hold(page, createPost, () => page.getByTestId("signup-create").click());
  learnTenant(i, r1);
  await page.goto(r1);
  await continueAtPrompt(page, who, i);
  // (b) A second creator cancels at the prompt: the tenant already exists and routes, no session is issued, and a later
  // ordinary sign-in on the new tenant host succeeds (membership true).
  const second = await open(browser, testInfo);
  const who2 = await freshUser();
  const j = intent(testInfo);
  await signupLogin(second.page, who2);
  await nameAndAddress(second.page, j);
  const r2 = await hold(second.page, createPost, () => second.page.getByTestId("signup-create").click());
  learnTenant(j, r2);
  expect(j.tenant, "each Organization has its own label").not.toBe(i.tenant);
  await second.page.goto(r2);
  await atEmulator(second.page);
  const state = new URL(second.page.url()).searchParams.get("state");
  expectSecret(state, "the pinned login's state");
  await second.page.goto(`${j.tenant}/auth/callback?error=access_denied&state=${state}`);
  await second.page.waitForURL((u) => u.origin === j.tenant && u.pathname === "/sign-in" && u.searchParams.get("status") === "canceled");
  expect(await cookieAt(second.context, j.tenant, SESSION_COOKIE)).toBeUndefined();
  expect(await organizationOf(j.tenant)).toMatch(UUID);
  await second.page.goto(`${j.tenant}/sign-in?return_to=/account`);
  await second.page.getByRole("button", { name: "Sign in" }).click();
  await enterCredentials(second.page, who2);
  await second.page.waitForURL((u) => u.origin === j.tenant && u.pathname === "/account");
  expect((await accountAt(second.page, j.tenant)).membership).toBe("member");
  urlsClean(i);
  urlsClean(j);
});

test("[E-SUP-6] the new owner uses the tenant at once, without a Group, profile or billing step", async ({ page }, testInfo) => {
  const who = await freshUser();
  const i = intent(testInfo);
  await signupLogin(page, who);
  await nameAndAddress(page, i);
  await page.getByTestId("signup-create").click();
  await continueAtPrompt(page, who, i);
  // No onboarding redirect: the existing tenant pages load directly for the owner.
  for (const path of ["/groups", "/account", "/groups/new"]) {
    const r = await page.goto(`${i.tenant}${path}`);
    expect(r?.status(), path).toBe(200);
    expect(new URL(page.url()).pathname, path).toBe(path);
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
  }
  expect((await accountAt(page, i.tenant)).membership).toBe("member");
  urlsClean(i);
});
