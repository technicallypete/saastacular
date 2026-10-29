import { randomBytes, randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";

import AxeBuilder from "@axe-core/playwright";
import { type Browser, type BrowserContext, type Locator, type Page, type TestInfo } from "@playwright/test";

import { expectAxeClean } from "./support/contrast";
import { expectRadioOptions } from "./support/form-controls";
import { chooseOption } from "./support/modes";
import { expect, test } from "./support/page-errors";
import { observe } from "./support/sanitize";

// G058 joining a workspace (Plan r3 §1.3, §1.5–§1.6, §6.2 E-ENR-1..5, §6.4) in the disposable HTTPS identity environment
// (infra/compose/identity.yml), projects enrollment-light and enrollment-dark. Every step goes through the real BFF routes
// to core's /internal/v1/enrollment and /internal/v1/auth routes, its facades and serializers; nothing is stubbed. The
// harness seeds the fixture first (infra/identity/mcp-seed.sql): Organization A is invite-only with its membership history
// closed (owner carol, so open is offered disabled), B is open, C is request access (owner carol, administrator frank,
// plain member alice); dave and bob hold no Organization C binding (asserted by the seed) and erin and grace
// (infra/identity/emulator-seed.yaml) have no account. Both projects run against the same database one after the other,
// so each uses its own subjects for the single-use steps: the light project's open applicant is erin, its invitation
// recipient grace and its request applicant dave; the dark project swaps erin and grace and uses bob. Whichever project
// runs first meets the first-time accounts; the other admits existing accounts of other workspaces through the same
// routes. E-ENR-4 runs before E-ENR-2 with the same applicant (a denial first; the later approval covers it). Pages that
// answer an error status are reached only from extra pages, so each test's default page stays under S-ERRORS. Core
// decides every read and action; these tests observe the UI and the cookies.

const ORIGIN_A = "https://org-a.localhost:543";
const ORIGIN_B = "https://org-b.localhost:543";
const ORIGIN_C = "https://org-c.localhost:543";
const EMULATOR = "http://emulator:4100";
const FORM = "application/x-www-form-urlencoded";
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const JOIN_LINK = /https:\/\/org-a\.localhost:543\/join\/invitations\/([0-9a-f-]{36})#([A-Za-z0-9_-]{43})/;
const AXE_TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];
const ENROLL_COOKIE = "__Host-enroll";
const JOIN_COOKIE = "__Host-join";
const SESSION_COOKIE = "__Host-sid";

type Credentials = { email: string; password: string };
type SeedUser = Credentials & { name: string; first_name: string; last_name: string };
const SEED = JSON.parse(readFileSync("/identity/browser/users.json", "utf8")) as { users: SeedUser[] };
function seeded(name: string): SeedUser {
  const found = SEED.users.find((u) => u.name === name);
  if (!found) throw new Error(`seed user ${name} missing`);
  return found;
}
const ALICE = seeded("alice");
const CAROL = seeded("carol");
const FRANK = seeded("frank");
const fullName = (u: SeedUser) => `${u.first_name} ${u.last_name}`;

// The single-use subjects of each project (see above).
function cast(testInfo: TestInfo): { open: SeedUser; invited: SeedUser; applicant: SeedUser } {
  return testInfo.project.name.endsWith("dark")
    ? { open: seeded("grace"), invited: seeded("erin"), applicant: seeded("bob") }
    : { open: seeded("erin"), invited: seeded("grace"), applicant: seeded("dave") };
}

function suffix(testInfo: TestInfo): string {
  return `${testInfo.project.name.endsWith("dark") ? "d" : "l"}${randomBytes(3).toString("hex")}`;
}

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
  const outside = [...tracking.origins].filter((o) => ![ORIGIN_A, ORIGIN_B, ORIGIN_C, EMULATOR].includes(o));
  expect(outside, "page requests outside the app origins under test and the emulator").toEqual([]);
});

async function open(browser: Browser, testInfo: TestInfo, origin = ORIGIN_A): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ baseURL: origin, colorScheme: testInfo.project.use.colorScheme, viewport: testInfo.project.use.viewport });
  track(context);
  tracking.contexts.push(context);
  return { context, page: await context.newPage() };
}

// --- sign-in through the emulator's hosted page ------------------------------------------------------------------------
async function atEmulator(page: Page, who: Credentials) {
  await page.waitForURL((u) => u.origin === EMULATOR);
  await page.locator("#email").fill(who.email);
  await page.getByRole("button", { name: "Continue" }).click();
  const password = page.locator('input[name="password"]');
  await password.fill(who.password);
  await password.press("Enter");
}

// The ordinary sign-in from ORIGIN's sign-in page, back on /account.
async function signIn(page: Page, who: Credentials, origin: string) {
  await page.goto(`${origin}/sign-in?return_to=/account`);
  await page.getByRole("button", { name: "Sign in" }).click();
  await atEmulator(page, who);
  await page.waitForURL((u) => u.origin === origin && u.pathname === "/account");
}

// After a confirmed admission: the status page's Sign in button starts the ordinary fresh sign-in, which alone issues a
// session (back on the tenant entry).
async function freshSignIn(page: Page, who: Credentials, origin: string) {
  const button = page.getByTestId("join-sign-in");
  await expect(button).toBeVisible();
  await button.click();
  await atEmulator(page, who);
  await page.waitForURL((u) => u.origin === origin && u.pathname === "/" && u.searchParams.get("status") === "signed_in");
  await expect(page.getByText("You are signed in.")).toBeVisible();
}

// The enrollment-intent start from ORIGIN's /join page ("Sign in to join" or "Sign in to request access").
async function joinFromEntry(page: Page, who: Credentials, origin: string) {
  await page.goto(`${origin}/join`);
  await page.getByTestId("join-start").click();
  await atEmulator(page, who);
}

async function cookie(context: BrowserContext, origin: string, name: string) {
  return (await context.cookies(origin)).find((c) => c.name === name);
}

// The live enrollment credential: present, host-only, Secure and HttpOnly; never beside a session cookie.
async function expectEnrollmentOnly(context: BrowserContext, origin: string) {
  const enroll = await cookie(context, origin, ENROLL_COOKIE);
  expect(enroll, "the enrollment credential cookie").toBeDefined();
  expect(enroll).toMatchObject({ path: "/", httpOnly: true, secure: true });
  observe(enroll?.value);
  expect(await cookie(context, origin, SESSION_COOKIE), "no session beside the enrollment credential").toBeUndefined();
}

async function expectNoCredentials(context: BrowserContext, origin: string) {
  expect(await cookie(context, origin, ENROLL_COOKIE), "the enrollment cookie is cleared").toBeUndefined();
  expect(await cookie(context, origin, JOIN_COOKIE), "the join cookie is cleared").toBeUndefined();
  expect(await cookie(context, origin, SESSION_COOKIE), "no session before the ordinary sign-in").toBeUndefined();
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

// Same-origin requests from an extra page of the context (the browser adds Origin and the cookies itself), so an
// expected error status never reaches the test page's S-ERRORS.
type Answer = { status: number; redirected: boolean; text: string };
async function fromPage(context: BrowserContext, origin: string, path: string, body?: string): Promise<Answer> {
  const p = await context.newPage();
  await p.goto(`${origin}/api/health`);
  const r = await p.evaluate(
    async ({ path, body, type }) => {
      const init: RequestInit = body === undefined ? { redirect: "manual", cache: "no-store" } : { method: "POST", body, headers: { "content-type": type }, redirect: "manual", cache: "no-store" };
      const res = await fetch(path, init);
      const redirected = res.type === "opaqueredirect";
      return { status: res.status, redirected, text: redirected ? "" : await res.text() };
    },
    { path, body, type: FORM },
  );
  await p.close();
  return r;
}

// Submits a same-origin form POST of the page itself (a real browser navigation with its Origin and cookies): the
// existing applicant routes in states where the status page offers no button for them.
async function postForm(page: Page, action: string, fields: Record<string, string>) {
  await page.evaluate(
    ({ action, fields }) => {
      const form = document.createElement("form");
      form.method = "post";
      form.action = action;
      for (const [name, value] of Object.entries(fields)) {
        const input = document.createElement("input");
        input.type = "hidden";
        input.name = name;
        input.value = value;
        form.append(input);
      }
      document.body.append(form);
      form.submit();
    },
    { action, fields },
  );
}

// The applicant's own CSRF token from the status page (the enrollment credential's token, never the session's).
async function applicantCsrf(page: Page): Promise<string> {
  const token = await page.locator('form[action="/join/abandon"] input[name="csrf_token"]').inputValue();
  expect(token).toMatch(SECRET);
  observe(token);
  return token;
}

const statusUrl = (status: string) => (u: URL) => u.pathname === "/join/status" && u.searchParams.get("status") === status;

// The status page never shows tenant data: no Group, Business Profile or member name, no Workspace areas navigation.
async function expectNoTenantData(page: Page) {
  const text = await page.locator("main").innerText();
  for (const s of ["Planning", "Operations", "Research", "Design Studio", "Alice Example", "Carol Example", "Frank Example", "Cyrus Owner"]) expect(text, s).not.toContain(s);
  await expect(page.getByRole("navigation", { name: "Workspace areas" })).toHaveCount(0);
}

async function tabTo(page: Page, target: Locator, label: string) {
  for (let i = 0; i < 80 && !(await target.evaluate((t) => t === document.activeElement).catch(() => false)); i++) await page.keyboard.press("Tab");
  await expect(target, `keyboard focus on ${label}`).toBeFocused();
}

async function audit(page: Page, label: string) {
  for (const width of [1280, 320]) {
    await page.setViewportSize({ width, height: width === 320 ? 640 : 800 });
    const results = await new AxeBuilder({ page }).withTags(AXE_TAGS).analyze();
    // G069: the kit's white label on its #0075ff primary (the unbranded hosts B and C) is an OWNER-WAIVED color-contrast
    // node, recorded with its ratio; every other violation fails (support/contrast.ts expectAxeClean).
    expectAxeClean(results.violations, `axe (${width} ${label})`);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), `no horizontal scroll at ${width} (${label})`).toBe(true);
  }
  await page.setViewportSize({ width: 1280, height: 800 });
}

// --- the enrollment page (owner policy, administrators' requests) -------------------------------------------------------
const ENROLLMENT = "/organization/enrollment";
const requestRows = (page: Page) => page.getByTestId("enrollment-request-row");
const pendingRow = (page: Page) => requestRows(page).filter({ has: page.getByTestId("enrollment-approve") });
const approvedRow = (page: Page) => requestRows(page).filter({ has: page.getByTestId("enrollment-revoke") });
const notice = (page: Page, message: string) => expect(page.getByText(message, { exact: true })).toBeVisible();

async function decide(page: Page, origin: string, action: "approve" | "deny" | "revoke", status: string, message: string) {
  await page.goto(`${origin}${ENROLLMENT}`);
  const row = action === "revoke" ? approvedRow(page) : pendingRow(page);
  await expect(row).toHaveCount(1);
  await row.getByTestId(`enrollment-${action}`).click();
  await page.waitForURL((u) => u.pathname === ENROLLMENT && u.searchParams.get("status") === status);
  await notice(page, message);
}

async function setMode(page: Page, origin: string, mode: "invite_only" | "request_access") {
  await page.goto(`${origin}${ENROLLMENT}`);
  await page.getByTestId(`enrollment-mode-${mode}`).check();
  await page.getByTestId("enrollment-policy-save").click();
  await page.waitForURL((u) => u.pathname === ENROLLMENT && u.searchParams.get("status") === "policy_changed");
  await notice(page, "The enrollment policy was changed.");
  await expect(page.getByTestId(`enrollment-mode-${mode}`)).toBeChecked();
}

// --- G018 tail (onboarding.spec.ts helpers, reused): the admitted member creates and onboards its own Group ------------
async function ownGroup(page: Page, origin: string, name: string, profile: string): Promise<string> {
  await page.goto(`${origin}/groups`);
  await page.getByRole("link", { name: "New Group", exact: true }).click();
  await page.waitForURL((u) => u.pathname === "/groups/new");
  await page.getByRole("textbox", { name: "Group name" }).fill(name);
  await page.getByRole("button", { name: "Continue" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Choose its address");
  await page.getByRole("button", { name: "Create Group" }).click();
  await page.waitForURL((u) => /^\/groups\/[0-9a-f-]{36}\/onboarding$/.test(u.pathname));
  const id = new URL(page.url()).pathname.split("/")[2];
  expect(id).toMatch(UUID);
  const form = page.getByTestId("profile-new-form");
  await form.getByRole("textbox", { name: "Display name" }).fill(profile);
  await form.getByRole("textbox", { name: "Industry" }).fill("Manufacturing");
  await form.getByRole("button", { name: "Create and attach" }).click();
  await page.waitForURL((u) => u.pathname === `/groups/${id}` && u.searchParams.get("status") === "completed");
  await expect(page.getByTestId("group-state")).toHaveText("Complete");
  return id;
}

// --- E-ENR-1..5 ---------------------------------------------------------------------------------------------------------

test("[E-ENR-1] open signup on B by keyboard: the join entry, admission at the callback, the ordinary fresh sign-in and an own onboarded Group; no owner controls", async ({ page, context }, testInfo) => {
  test.slow();
  const who = cast(testInfo).open;
  // The logged-out tenant entry offers Join (open mode); keyboard only on the application's pages.
  await page.goto(`${ORIGIN_B}/`);
  const join = page.getByTestId("entry-join");
  await tabTo(page, join, "Join this workspace");
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.origin === ORIGIN_B && u.pathname === "/join");
  await expect(page.getByText(/^Anyone who signs in can join /)).toBeVisible();
  const start = page.getByTestId("join-start");
  await expect(start).toHaveText("Sign in to join");
  await tabTo(page, start, "Sign in to join");
  await page.keyboard.press("Enter");
  await atEmulator(page, who);
  // The callback admitted at once (open): a confirmed admission, no enrollment credential, no session.
  await page.waitForURL((u) => u.origin === ORIGIN_B && statusUrl("admitted")(u));
  await notice(page, "Your membership was confirmed. Sign in to open the workspace.");
  await expectNoCredentials(context, ORIGIN_B);
  await expectNoTenantData(page);
  const signInButton = page.getByTestId("join-sign-in");
  await tabTo(page, signInButton, "Sign in");
  await page.keyboard.press("Enter");
  await atEmulator(page, who);
  await page.waitForURL((u) => u.origin === ORIGIN_B && u.pathname === "/" && u.searchParams.get("status") === "signed_in");
  await expect(page.getByText("You are signed in.")).toBeVisible();
  expect(await cookie(context, ORIGIN_B, SESSION_COOKIE), "the ordinary sign-in issued the session").toBeDefined();
  // The ordinary member continues into G018's own Group with a new Business Profile.
  const sfx = suffix(testInfo);
  await ownGroup(page, ORIGIN_B, `Joined ${sfx}`, `Joined Co ${sfx}`);
  // A plain member: no owner or invitation controls, a read-only enrollment mode and no request list.
  await page.goto(`${ORIGIN_B}/organization/members`);
  await expect(page.getByTestId("self-card")).toContainText("Member");
  await expect(page.getByTestId("transfer-form")).toHaveCount(0);
  await expect(page.getByTestId("invitation-form")).toHaveCount(0);
  await expect(page.getByRole("button", { name: /^Make .+ an owner$/ })).toHaveCount(0);
  await page.goto(`${ORIGIN_B}${ENROLLMENT}`);
  await expect(page.getByTestId("enrollment-policy-readonly")).toContainText("Current mode: Open: anyone who signs in can join.");
  await expect(page.getByTestId("enrollment-policy-form")).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "Access requests" })).toHaveCount(0);
});

test("[E-ENR-3] invite-only A: no join form; the owner's Organization invitation link opens /join/invitations and the recipient joins through the ordinary fresh sign-in", async ({ page, browser }, testInfo) => {
  test.slow();
  const who = cast(testInfo).invited;
  const recipient = await open(browser, testInfo);
  // Invite-only: the entry has no Join link and /join no form.
  await recipient.page.goto(`${ORIGIN_A}/`);
  await expect(recipient.page.getByTestId("entry-join")).toHaveCount(0);
  await recipient.page.goto(`${ORIGIN_A}/join`);
  await expect(recipient.page.getByTestId("join-invite-only")).toBeVisible();
  await expect(recipient.page.getByTestId("join-start")).toHaveCount(0);
  // carol (owner of A) invites the recipient's address; the one-time link is the enrollment entry (amendment04).
  await signIn(page, CAROL, ORIGIN_A);
  await page.goto("/organization/members");
  const form = page.getByTestId("invitation-form");
  await form.getByLabel("Email address").fill(who.email);
  await chooseOption(form.getByLabel("Role"), "Member");
  await form.getByRole("button", { name: "Create invitation" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Invitation created");
  const m = JOIN_LINK.exec(await page.locator("main").innerText());
  expect(m, "the Organization invitation link on the enrollment entry").not.toBeNull();
  const [link, id, token] = m as RegExpExecArray;
  expect(token).toMatch(SECRET);
  observe(token);
  expect(await page.title()).not.toContain(token);
  // The recipient's traffic never carries the token in a URL or a Location header.
  const leaks: string[] = [];
  recipient.context.on("request", (r) => {
    if (r.url().includes(token)) leaks.push(`request ${new URL(r.url()).pathname}`);
  });
  recipient.context.on("response", (r) => {
    if ((r.headers()["location"] ?? "").includes(token)) leaks.push(`location of ${new URL(r.url()).pathname}`);
  });
  await recipient.page.goto(link);
  await expect(recipient.page.getByRole("heading", { level: 1 })).toHaveText(/^Join /);
  await expect(recipient.page.getByLabel("Invitation code")).toHaveValue(token);
  await expect.poll(() => new URL(recipient.page.url()).hash).toBe("");
  expect(new URL(recipient.page.url()).pathname).toBe(`/join/invitations/${id}`);
  await recipient.page.getByTestId("join-invitation-submit").click();
  await recipient.page.waitForURL((u) => u.origin === EMULATOR);
  const intent = await cookie(recipient.context, ORIGIN_A, JOIN_COOKIE);
  expect(intent, "the one-use join cookie carries the intent to the callback").toBeDefined();
  expect(intent).toMatchObject({ path: "/", httpOnly: true, secure: true });
  observe(intent?.value);
  await atEmulator(recipient.page, who);
  await recipient.page.waitForURL((u) => u.origin === ORIGIN_A && statusUrl("admitted")(u));
  await expectNoCredentials(recipient.context, ORIGIN_A);
  await freshSignIn(recipient.page, who, ORIGIN_A);
  await recipient.page.goto("/organization/members");
  await expect(recipient.page.getByTestId("self-card")).toContainText("Member");
  expect(leaks, "the invitation token in a URL or Location").toEqual([]);
  // carol sees the new member and the invitation is no longer pending.
  await page.goto("/organization/members");
  await expect(page.getByTestId("member-row").filter({ hasText: fullName(who) })).toHaveCount(1);
  await expect(page.getByTestId("invitation-row").filter({ hasText: who.email })).not.toContainText("pending", { ignoreCase: true });
});

test("[E-ENR-4] a request on C closed after it started and then denied: truthful status pages without tenant data", async ({ page, context, browser }, testInfo) => {
  test.slow();
  const who = cast(testInfo).applicant;
  // The applicant (default page) asks to join C; the callback files the request and sets only the enrollment credential.
  await joinFromEntry(page, who, ORIGIN_C);
  await page.waitForURL((u) => u.origin === ORIGIN_C && statusUrl("pending")(u));
  await expect(page.getByTestId("join-state-pending")).toContainText("Your request is waiting for an administrator's decision. Nothing has been granted yet.");
  await expectEnrollmentOnly(context, ORIGIN_C);
  expect((await fromPage(context, ORIGIN_C, "/api/account")).status, "the enrollment credential is not a session").toBe(401);
  await expectNoTenantData(page);
  await audit(page, "status pending");
  // carol (owner of C) closes C after the applicant started.
  const carol = await open(browser, testInfo, ORIGIN_C);
  await signIn(carol.page, CAROL, ORIGIN_C);
  await carol.page.goto(`${ORIGIN_C}${ENROLLMENT}`);
  await expect(carol.page.getByTestId("enrollment-mode-request_access")).toBeChecked();
  await expect(carol.page.getByTestId("enrollment-mode-open")).toBeEnabled();
  await setMode(carol.page, ORIGIN_C, "invite_only");
  // The applicant's next step is refused as closed; the page still shows core's current state.
  await page.goto(`${ORIGIN_C}/join/status`);
  await postForm(page, "/join/request", { csrf_token: await applicantCsrf(page) });
  await page.waitForURL((u) => u.origin === ORIGIN_C && statusUrl("closed")(u));
  await notice(page, "This workspace is not accepting new members this way right now. Nothing was changed.");
  await expect(page.getByTestId("join-state-pending")).toBeVisible();
  await expectNoTenantData(page);
  await setMode(carol.page, ORIGIN_C, "request_access");
  // frank (administrator of C) denies the request.
  const frank = await open(browser, testInfo, ORIGIN_C);
  await signIn(frank.page, FRANK, ORIGIN_C);
  await decide(frank.page, ORIGIN_C, "deny", "denied", "The request was denied.");
  await page.goto(`${ORIGIN_C}/join/status`);
  await expect(page.getByTestId("join-state-denied")).toContainText("Your request was not approved.");
  await expectNoTenantData(page);
  await audit(page, "status denied");
});

test("[E-ENR-2] a request on C: approved, withdrawn, re-requested, approved and admitted through the ordinary fresh sign-in; a member and an administrator are refused the wrong authority", async ({ page, context, browser }, testInfo) => {
  test.slow();
  const who = cast(testInfo).applicant;
  await joinFromEntry(page, who, ORIGIN_C);
  await page.waitForURL((u) => u.origin === ORIGIN_C && statusUrl("pending")(u));
  await expect(page.getByTestId("join-state-pending")).toBeVisible();
  await expectEnrollmentOnly(context, ORIGIN_C);
  // POST /join/request on the existing pending request: the same pending request (idempotent).
  await postForm(page, "/join/request", { csrf_token: await applicantCsrf(page) });
  await page.waitForURL((u) => u.origin === ORIGIN_C && statusUrl("requested")(u));
  await notice(page, "Your request was sent. An administrator of this workspace will review it.");
  await expect(page.getByTestId("join-state-pending")).toBeVisible();
  // frank (administrator) reads the policy and the request list.
  const frank = await open(browser, testInfo, ORIGIN_C);
  await signIn(frank.page, FRANK, ORIGIN_C);
  await frank.page.goto(`${ORIGIN_C}${ENROLLMENT}`);
  await expect(frank.page.getByTestId("enrollment-policy-readonly")).toContainText("Current mode: People can request access; an administrator decides.");
  await expect(frank.page.getByRole("heading", { name: "Access requests" })).toBeVisible();
  await expect(pendingRow(frank.page)).toHaveCount(1);
  const requestId = await pendingRow(frank.page).locator('input[name="request_id"]').first().inputValue();
  expect(requestId).toMatch(UUID);
  // Wrong authority, through the BFF to core: the plain member alice may not decide; the administrator may not change
  // the policy (owner only). Nothing changes.
  const alice = await open(browser, testInfo, ORIGIN_C);
  await signIn(alice.page, ALICE, ORIGIN_C);
  const aliceCsrf = await csrfOf(alice.page);
  const memberDecide = await fromPage(alice.context, ORIGIN_C, ENROLLMENT + "/action", new URLSearchParams({ csrf_token: aliceCsrf, action: "approve", request_id: requestId }).toString());
  expect(memberDecide.status, "a plain member's decision").toBe(403);
  expect(memberDecide.text).toContain("Not permitted");
  await alice.page.goto(`${ORIGIN_C}${ENROLLMENT}`);
  await expect(alice.page.getByTestId("enrollment-policy-readonly")).toBeVisible();
  await expect(alice.page.getByRole("heading", { name: "Access requests" })).toHaveCount(0);
  const frankCsrf = await csrfOf(frank.page);
  const adminPolicy = await fromPage(frank.context, ORIGIN_C, ENROLLMENT + "/action", new URLSearchParams({ csrf_token: frankCsrf, action: "policy", mode: "invite_only" }).toString());
  expect(adminPolicy.status, "an administrator's policy change").toBe(403);
  expect(adminPolicy.text).toContain("Not permitted");
  await frank.page.reload();
  await expect(frank.page.getByTestId("enrollment-policy-readonly")).toContainText("People can request access");
  await expect(pendingRow(frank.page)).toHaveCount(1);
  // Approve, then withdraw the approval before the applicant returns.
  await decide(frank.page, ORIGIN_C, "approve", "approved", "The request was approved. The person joins when they return and confirm.");
  await expect(approvedRow(frank.page)).toHaveCount(1);
  await decide(frank.page, ORIGIN_C, "revoke", "revoked", "The approval was withdrawn.");
  await expect(approvedRow(frank.page)).toHaveCount(0);
  await page.goto(`${ORIGIN_C}/join/status`);
  await expect(page.getByTestId("join-state-revoked")).toContainText("The approval of your request was withdrawn. Nothing was granted.");
  await expectNoTenantData(page);
  // The applicant stops (the credential is revoked and its cookie cleared), then starts again with a new request.
  await page.getByTestId("join-abandon").click();
  await page.waitForURL((u) => u.origin === ORIGIN_C && u.pathname === "/join" && u.searchParams.get("status") === "abandoned");
  await notice(page, "You stopped joining. A request you sent stays with the administrators.");
  expect(await cookie(context, ORIGIN_C, ENROLL_COOKIE), "abandonment clears the enrollment cookie").toBeUndefined();
  await joinFromEntry(page, who, ORIGIN_C);
  await page.waitForURL((u) => u.origin === ORIGIN_C && statusUrl("pending")(u));
  await expect(page.getByTestId("join-state-pending")).toBeVisible();
  await decide(frank.page, ORIGIN_C, "approve", "approved", "The request was approved. The person joins when they return and confirm.");
  // The applicant returns, confirms (POST /join/admit kind=request) and signs in the ordinary way.
  await page.goto(`${ORIGIN_C}/join/status`);
  await expect(page.getByTestId("join-state-approved")).toBeVisible();
  await audit(page, "status approved");
  await page.getByTestId("join-admit-request").click();
  await page.waitForURL((u) => u.origin === ORIGIN_C && statusUrl("admitted")(u));
  await notice(page, "Your membership was confirmed. Sign in to open the workspace.");
  await expectNoCredentials(context, ORIGIN_C);
  await freshSignIn(page, who, ORIGIN_C);
  await page.goto(`${ORIGIN_C}/organization/members`);
  await expect(page.getByTestId("self-card")).toContainText("Member");
  await frank.page.goto(`${ORIGIN_C}${ENROLLMENT}`);
  await expect(requestRows(frank.page).filter({ hasText: "Joined" })).not.toHaveCount(0);
  await expect(pendingRow(frank.page)).toHaveCount(0);
});

test("[E-ENR-5] keyboard-only entries and owner policy change; axe at 1280 and 320 on every enrollment page", async ({ page, browser }, testInfo) => {
  test.slow();
  // Logged-out entries of the three modes, the status page without a credential and an invitation landing.
  for (const [origin, joinable] of [[ORIGIN_A, false], [ORIGIN_B, true], [ORIGIN_C, true]] as const) {
    await page.goto(`${origin}/`);
    await expect(page.getByTestId("entry-join")).toHaveCount(joinable ? 1 : 0);
    await page.goto(`${origin}/join`);
    await expect(page.getByTestId(joinable ? "join-start" : "join-invite-only")).toBeVisible();
    await audit(page, `join ${origin}`);
  }
  await page.goto(`${ORIGIN_B}/join/status`);
  await expect(page.getByRole("link", { name: "Start joining" })).toBeVisible();
  await audit(page, "status without a credential");
  await page.goto(`${ORIGIN_A}/join/invitations/${randomUUID()}`);
  await expect(page.getByLabel("Invitation code")).toHaveValue("");
  await audit(page, "invitation landing");
  // Keyboard: C's entry Join link, then its request start, reaches the hosted sign-in.
  await page.goto(`${ORIGIN_C}/`);
  await tabTo(page, page.getByTestId("entry-join"), "Join this workspace");
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.origin === ORIGIN_C && u.pathname === "/join");
  await tabTo(page, page.getByTestId("join-start"), "Sign in to request access");
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.origin === EMULATOR);
  // carol, owner of A (history closed): open is disabled with the generic sentence; mode changes by keyboard only.
  const carol = await open(browser, testInfo, ORIGIN_A);
  await signIn(carol.page, CAROL, ORIGIN_A);
  await carol.page.goto("/organization/members");
  await tabTo(carol.page, carol.page.getByTestId("members-enrollment-link"), "Enrollment");
  await carol.page.keyboard.press("Enter");
  await carol.page.waitForURL((u) => u.pathname === ENROLLMENT);
  const modeRadio = (m: string) => carol.page.getByTestId(`enrollment-mode-${m}`);
  await expect(modeRadio("open")).toBeDisabled();
  await expect(carol.page.getByTestId("enrollment-open-not-ready")).toHaveText("Open signup is unavailable until the platform operator has reviewed membership history");
  await expect(carol.page.getByRole("group", { name: "Who can join" })).toHaveAttribute("aria-describedby", "enrollment-open-hint");
  await audit(carol.page, "owner enrollment page");
  // G068/G069: the kit radio treatment, label alignment, form value and focus state, the disabled open option included.
  await expectRadioOptions(carol.page, testInfo, carol.page.getByRole("group", { name: "Who can join" }), "Who can join", "mode");
  for (const [from, key, to] of [["invite_only", "ArrowDown", "request_access"], ["request_access", "ArrowUp", "invite_only"]] as const) {
    await expect(modeRadio(from)).toBeChecked();
    await tabTo(carol.page, modeRadio(from), from);
    await carol.page.keyboard.press(key);
    await expect(modeRadio(to)).toBeChecked();
    await tabTo(carol.page, carol.page.getByTestId("enrollment-policy-save"), "Save");
    await carol.page.keyboard.press("Enter");
    await carol.page.waitForURL((u) => u.pathname === ENROLLMENT && u.searchParams.get("status") === "policy_changed");
    await expect(modeRadio(to)).toBeChecked();
  }
  // The administrator's request page and a member's read-only page on C.
  const frank = await open(browser, testInfo, ORIGIN_C);
  await signIn(frank.page, FRANK, ORIGIN_C);
  await frank.page.goto(`${ORIGIN_C}${ENROLLMENT}`);
  await expect(frank.page.getByRole("heading", { name: "Access requests" })).toBeVisible();
  await audit(frank.page, "administrator enrollment page");
  const alice = await open(browser, testInfo, ORIGIN_C);
  await signIn(alice.page, ALICE, ORIGIN_C);
  await alice.page.goto(`${ORIGIN_C}${ENROLLMENT}`);
  await expect(alice.page.getByTestId("enrollment-policy-readonly")).toBeVisible();
  await audit(alice.page, "member enrollment page");
});
