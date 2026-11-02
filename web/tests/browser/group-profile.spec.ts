import { randomBytes } from "node:crypto";
import { readFileSync } from "node:fs";

import AxeBuilder from "@axe-core/playwright";
import { type Browser, type BrowserContext, type Locator, type Page, type TestInfo } from "@playwright/test";

import { expectAxeClean } from "./support/contrast";
import { chooseOption } from "./support/modes";
import { expect, test } from "./support/page-errors";
import { observe } from "./support/sanitize";

// G075 Group member profiles (Plan r2 §3.5, §3.7, §5 E-GMP-1..7, S-GMP-A11Y) in the disposable HTTPS identity environment
// (infra/compose/identity.yml), projects profile-light and profile-dark (after profile.spec.ts's G074 scenarios' file in
// the same projects), on the identity seed (infra/identity/mcp-seed.sql): on Organization A carol is the owner, frank an
// admin, alice a plain member without a Group, bob a member and the owner of research, dave a member of research and
// hostile a plain member; on Organization B dave is a member and a member of operations; bianca (B only) and cyrus (C only)
// never sign in. The seed's second admin of A (grace, also a research admin) has no sign-in principal, so the Group
// manager edits run as bob, research's owner (the Group-admin path is the facade's GMP-3). Both projects run against the
// same database one after the other: every label a scenario sets is cleared again, the Group-administration override is
// turned off, an Organization role changed for a scenario is changed back, and dave is removed again from the Group the
// scenario created (Groups cannot be deleted; they stay, like G074's), in finally blocks, so the seed state is unchanged
// for every later project. Labels carry a per-project suffix and versions are compared relative to what the scenario read
// first. Expected refusals are read from extra pages, so the default page's S-ERRORS check covers every page the user sees
// succeed. Core decides every card and every write; these tests observe the composite profile page, its label form route
// and the JSON route GET /api/profiles/<id>/groups.

const ORIGIN_A = "https://org-a.localhost:543";
const ORIGIN_B = "https://org-b.localhost:543";
const ATTACKER = "https://attacker.localhost:543";
const EMULATOR = "http://emulator:4100";
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const LINK = /https:\/\/org-a\.localhost:543\/invitations\/([0-9a-f-]{36})#([A-Za-z0-9_-]{43})/;
const AXE_TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

// Seeded rows (infra/identity/mcp-seed.sql).
const RESEARCH = "0f14a000-0000-4000-8000-0000000000a1";
const DESIGN = "0f14a000-0000-4000-8000-0000000000a2";
const OPERATIONS = "0f14b000-0000-4000-8000-0000000000b1";
const BOB_ID = "0f16e000-0000-4000-8000-0000000000b1";
const DAVE_ID = "0f14d000-0000-4000-8000-00000000000d";
const BIANCA_ID = "0f19e000-0000-4000-8000-0000000000b0";
const CYRUS_ID = "0f19e000-0000-4000-8000-0000000000c0";
const UNKNOWN_ID = "0f16e000-0000-4000-8000-0000000000ff";

type Credentials = { email: string; password: string };
type SeedUser = Credentials & { name: string };
const SEED = JSON.parse(readFileSync("/identity/browser/users.json", "utf8")) as { users: SeedUser[] };
function seeded(name: string): SeedUser {
  const found = SEED.users.find((u) => u.name === name);
  if (!found) throw new Error(`seed user ${name} missing`);
  return found;
}
const ALICE = seeded("alice");
const BOB = seeded("bob");
const CAROL = seeded("carol");
const DAVE = seeded("dave");
const FRANK = seeded("frank");
const HOSTILE = seeded("hostile");
const ALICE_NAME = "Alice Example";
const BOB_NAME = "Bob Example";
const DAVE_NAME = "Dave Example";

// The page's closed texts (components/profile/group-cards-section.tsx).
const SELECTED_CONTEXT = "Showing this person's card for one Group.";
const NO_CARDS = "No Groups to show.";
const NO_ROLE = "No role given";
const STATUS_TEXT = {
  label_saved: "The role was saved.",
  label_cleared: "The role was removed.",
  invalid_label: "That role was not saved. Use up to 100 characters without line breaks or control characters. Nothing was changed.",
  not_permitted: "You cannot change that role. Only the member and that Group's owners and admins can. Nothing was changed.",
  invalid_context:
    "That request named a Group context that is not valid, so it was refused. Nothing was changed. Open this profile again from a Group's members page.",
  stale_version:
    "This role changed since the page was loaded, for example in another tab. Review the current value below and try again. Nothing was changed.",
} as const;

// The exact JSON keys (lib/group-profile-client.ts CARDS_KEYS, CARD_KEYS), sorted.
const LIST_KEYS = ["cards", "narrow_required", "user_id"];
const CARD_KEYS = ["can_edit", "group_id", "group_name", "role_label", "version"];

function suffix(testInfo: TestInfo): string {
  return `${testInfo.project.name.endsWith("dark") ? "d" : "l"}${randomBytes(3).toString("hex")}`;
}

// --- page traffic and extra contexts ------------------------------------------------------------------------------------
type Tracking = { origins: Set<string>; allowed: Set<string>; contexts: BrowserContext[] };
let tracking: Tracking;

function track(context: BrowserContext) {
  context.on("request", (request) => {
    const url = request.url();
    if (/^https?:/.test(url)) tracking.origins.add(new URL(url).origin);
  });
}

test.beforeEach(async ({ context }) => {
  tracking = { origins: new Set(), allowed: new Set([ORIGIN_A, EMULATOR]), contexts: [] };
  track(context);
});

test.afterEach(async () => {
  for (const c of tracking.contexts) await c.close();
  const outside = [...tracking.origins].filter((o) => !tracking.allowed.has(o));
  expect(outside, "page requests outside the app origins under test and the emulator").toEqual([]);
});

async function open(browser: Browser, testInfo: TestInfo, origin = ORIGIN_A): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ baseURL: origin, colorScheme: testInfo.project.use.colorScheme, viewport: testInfo.project.use.viewport });
  track(context);
  tracking.contexts.push(context);
  return { context, page: await context.newPage() };
}

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

// One signed-in extra browser profile.
async function signedIn(browser: Browser, testInfo: TestInfo, who: Credentials, origin = ORIGIN_A): Promise<{ context: BrowserContext; page: Page }> {
  const s = await open(browser, testInfo, origin);
  await signIn(s.page, who, origin);
  return s;
}

// A same-origin request from an extra page of the context, so an expected error status never reaches S-ERRORS. A form
// POST carries the form encoding (the browser adds its Origin); follow makes the request follow a 303 and report the page
// it landed on (url, without the fragment). Headers are lower-cased.
type Fetched = { status: number; url: string; redirected: boolean; headers: Record<string, string>; text: string };
type Init = { fields?: Record<string, string> | [string, string][]; follow?: boolean };
async function fetchFrom(context: BrowserContext, path: string, init: Init = {}, origin = ORIGIN_A): Promise<Fetched> {
  const p = await context.newPage();
  await p.goto(`${origin}/api/health`);
  const body = init.fields === undefined ? undefined : new URLSearchParams(init.fields).toString();
  const r = await p.evaluate(
    async ({ path, body, follow }) => {
      const request: RequestInit = { method: body === undefined ? "GET" : "POST", redirect: follow ? "follow" : "manual", cache: "no-store" };
      if (body !== undefined) {
        request.body = body;
        request.headers = { "content-type": "application/x-www-form-urlencoded" };
      }
      const res = await fetch(path, request);
      const opaque = res.type === "opaqueredirect";
      const headers: Record<string, string> = {};
      res.headers.forEach((v, k) => {
        headers[k] = v;
      });
      return { status: res.status, url: res.url, redirected: opaque || res.redirected, headers, text: opaque ? "" : await res.text() };
    },
    { path, body, follow: init.follow === true },
  );
  await p.close();
  return r;
}

type Card = { group_id: string; group_name: string; role_label: string | null; can_edit: boolean; version: number };
type Cards = { status: number; cache: string; text: string; body: Record<string, unknown>; cards: Card[] };
// GET /api/profiles/<id>/groups (query: "" or "?context=..."), with the strict key check of a 200 answer.
async function cardsJSON(context: BrowserContext, subject: string, query = "", origin = ORIGIN_A): Promise<Cards> {
  const r = await fetchFrom(context, `/api/profiles/${subject}/groups${query}`, {}, origin);
  const body = JSON.parse(r.text) as Record<string, unknown>;
  const cards = r.status === 200 ? (body.cards as Card[]) : [];
  if (r.status === 200) {
    expect(Object.keys(body).sort(), `list keys (${subject}${query})`).toEqual(LIST_KEYS);
    expect(body.user_id).toBe(subject);
    for (const c of cards) expect(Object.keys(c).sort(), `card keys (${subject}${query})`).toEqual(CARD_KEYS);
  }
  return { status: r.status, cache: r.headers["cache-control"] ?? "", text: r.text, body, cards };
}
const idsOf = (cards: Card[]) => cards.map((c) => c.group_id).sort();
const cardIn = (cards: Card[], group: string) => cards.find((c) => c.group_id === group);

async function tabTo(page: Page, target: Locator, label: string) {
  for (let i = 0; i < 80 && !(await target.evaluate((t) => t === document.activeElement).catch(() => false)); i++) await page.keyboard.press("Tab");
  await expect(target, `keyboard focus on ${label}`).toBeFocused();
}

async function audit(page: Page, label: string) {
  for (const width of [1280, 320]) {
    await page.setViewportSize({ width, height: width === 320 ? 640 : 800 });
    const results = await new AxeBuilder({ page }).withTags(AXE_TAGS).analyze();
    // G069: only the kit's white label on its #0075ff primary (where no tenant palette applies) is an OWNER-WAIVED
    // color-contrast node, recorded with its ratio; every other violation fails (support/contrast.ts expectAxeClean).
    expectAxeClean(results.violations, `axe (${width} ${label})`);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), `no horizontal scroll at ${width} (${label})`).toBe(true);
  }
  await page.setViewportSize({ width: 1280, height: 800 });
}

// The CSRF token of the current page's forms (registered with the report sanitizer).
async function csrfOf(page: Page): Promise<string> {
  const value = await page.locator('input[name="csrf_token"]').first().inputValue();
  expect(value).toMatch(SECRET);
  observe(value);
  return value;
}

// --- the composite profile page -----------------------------------------------------------------------------------------

const profilePath = (subject: string, context?: string) => `/profiles/${subject}${context === undefined ? "" : `?context=${context}`}`;
const labelPath = (subject: string, group: string) => `/profiles/${subject}/groups/${group}/label`;
const cardOf = (page: Page, group: string) => page.locator(`[data-testid="group-card"][data-group-id="${group}"]`);

async function pageCardIds(page: Page): Promise<string[]> {
  return (await page.getByTestId("group-card").evaluateAll((els) => els.map((e) => e.getAttribute("data-group-id") ?? ""))).sort();
}

// Opens the subject's page (optionally in a selected-Group context), fills the card's label form and saves it; returns
// the closed group_status of the 303 back to the page (which keeps the selector).
async function saveLabel(page: Page, subject: string, group: string, value: string, context?: string): Promise<string | null> {
  await page.goto(profilePath(subject, context));
  const card = cardOf(page, group);
  await card.getByRole("textbox").fill(value);
  await card.getByRole("button", { name: "Save role", exact: true }).click();
  await page.waitForURL((u) => u.pathname === `/profiles/${subject}` && u.searchParams.has("group_status"));
  const url = new URL(page.url());
  expect(url.searchParams.get("context") ?? undefined, "the 303 keeps the page's selector").toBe(context);
  return url.searchParams.get("group_status");
}

async function expectStatus(page: Page, status: keyof typeof STATUS_TEXT) {
  const testId = status === "stale_version" ? "group-label-alert" : "group-label-status";
  await expect(page.getByTestId(testId)).toHaveText(STATUS_TEXT[status]);
}

// Clears the signed-in member's (or manager's) label of subject in group when one is set (best effort in finally blocks).
async function clearLabel(page: Page, subject: string, group: string, origin = ORIGIN_A) {
  const now = await cardsJSON(page.context(), subject, "", origin);
  const card = cardIn(now.cards, group);
  if (card === undefined || card.role_label === null || !card.can_edit) return;
  // saveLabel navigates relative to the page's own baseURL, which is that origin (the default page or an extra context).
  expect(await saveLabel(page, subject, group, "")).toBe("label_cleared");
}

// --- Groups, invitations and roles (the G018 New Group, G019 invitation and member flows) -------------------------------

async function createOwnGroup(page: Page, name: string): Promise<string> {
  await page.goto("/groups/new");
  await page.getByRole("textbox", { name: "Group name" }).fill(name);
  await page.getByRole("button", { name: "Continue" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Choose its address");
  await page.getByRole("button", { name: "Create Group" }).click();
  await page.waitForURL((u) => /^\/groups\/[0-9a-f-]{36}\/onboarding$/.test(u.pathname));
  const id = new URL(page.url()).pathname.split("/")[2];
  expect(id).toMatch(UUID);
  return id;
}

const membersOf = (groupId: string) => `/groups/${groupId}/members`;
const memberRow = (page: Page, name: string) => page.getByTestId("member-row").filter({ hasText: name });
const roleCell = (page: Page, name: string) => memberRow(page, name).getByRole("cell").nth(1);

// The Group owner invites a co-member as Member by the name its picker shows and returns the one-time link.
async function inviteToGroup(page: Page, groupId: string, name: string): Promise<string> {
  await page.goto(membersOf(groupId));
  const form = page.getByTestId("invitation-form");
  await chooseOption(form.getByLabel("Workspace member"), name);
  await chooseOption(form.getByLabel("Role"), "Member");
  await form.getByRole("button", { name: "Create invitation" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Invitation created");
  const m = LINK.exec(await page.locator("main").innerText());
  expect(m, "the one-time invitation link").not.toBeNull();
  const [link, , token] = m as RegExpExecArray;
  expect(token).toMatch(SECRET);
  observe(token);
  return link;
}

async function accept(page: Page, link: string, groupId: string) {
  await page.goto(link);
  await expect(page.getByLabel("Invitation code")).not.toHaveValue("");
  await expect.poll(() => new URL(page.url()).hash).toBe("");
  await page.getByRole("button", { name: "Accept invitation" }).click();
  await page.waitForURL((u) => u.pathname === `/groups/${groupId}` && u.searchParams.get("status") === "joined");
}

async function removeFromGroup(page: Page, groupId: string, name: string) {
  await page.goto(membersOf(groupId));
  await memberRow(page, name).getByRole("button", { name: `Remove ${name}` }).click();
  await page.waitForURL((u) => u.pathname === membersOf(groupId) && u.searchParams.get("status") === "removed");
  await expect(memberRow(page, name)).toHaveCount(0);
}

// carol (the Organization owner) changes an Organization member's role on the workspace members page.
async function setOrganizationRole(owner: Page, name: string, role: "Admin" | "Member") {
  await owner.goto("/organization/members");
  const row = memberRow(owner, name);
  await chooseOption(row.getByLabel(`New role for ${name}`), role);
  await row.locator("form", { has: owner.getByRole("button", { name: "Change role" }) }).getByRole("button", { name: "Change role" }).click();
  await owner.waitForURL((u) => u.pathname === "/organization/members" && u.searchParams.get("status") === "role_changed");
  await expect(roleCell(owner, name)).toHaveText(role);
}

const GROUP_AUTHORITY = "/organization/group-authority";
const OVERRIDE_ON = "Group administration by Organization staff is now on.";
const OVERRIDE_OFF = "Group administration by Organization staff is now off. Organization owners and admins keep only the Group roles they actually hold.";

// carol sets the G060 Group-administration override through the settings form and sees core's state.
async function setOverride(owner: Page, state: "on" | "off") {
  await owner.goto(GROUP_AUTHORITY);
  const form = owner.getByTestId("group-authority-form");
  await form.getByTestId(`group-authority-${state}`).check();
  await form.getByRole("button", { name: "Save" }).click();
  await owner.waitForURL((u) => u.pathname === GROUP_AUTHORITY && u.searchParams.get("status") === (state === "on" ? "enabled" : "disabled"));
  await expect(owner.getByText(state === "on" ? OVERRIDE_ON : OVERRIDE_OFF, { exact: true })).toBeVisible();
  await expect(owner.getByTestId(`group-authority-${state}`)).toBeChecked();
}

// A forged label write with the writer's own valid CSRF token, following the form route's answer.
async function forgeLabel(context: BrowserContext, csrf: string, subject: string, group: string, version: number, label: string, extra: [string, string][] = []): Promise<Fetched> {
  const fields: [string, string][] = [["csrf_token", csrf], ["expected_version", String(version)], ["role_label", label], ...extra];
  return fetchFrom(context, labelPath(subject, group), { fields, follow: true });
}
const statusOf = (r: Fetched) => new URL(r.url).searchParams.get("group_status");

// --- E-GMP-1..7, S-GMP-A11Y -----------------------------------------------------------------------------------------------

test("[E-GMP-1] a member's own labels are tenant-local: set, replace and clear on host A; host B shows its own card and label; the header is one global identity", async ({ page, browser }, testInfo) => {
  test.slow();
  tracking.allowed.add(ORIGIN_B);
  const sfx = suffix(testInfo);
  const first = `Lab lead ${sfx}`;
  const second = `Principal investigator ${sfx}`;
  const onB = `Shift coordinator ${sfx}`;
  await signIn(page, DAVE);
  const b = await signedIn(browser, testInfo, DAVE, ORIGIN_B);
  try {
    // dave on host A: his own composite page with an editable research card; design (not his) and B's operations absent.
    await page.goto(profilePath(DAVE_ID));
    await expect(page.getByTestId("member-profile-name")).toHaveText(DAVE_NAME);
    await expect(page.getByTestId("group-cards-context")).toHaveCount(0);
    const startA = await cardsJSON(page.context(), DAVE_ID);
    expect([startA.status, startA.cache, startA.body.narrow_required]).toEqual([200, "no-store", false]);
    const research0 = cardIn(startA.cards, RESEARCH);
    expect(research0, "dave's research card on host A").toBeDefined();
    expect(startA.cards.every((c) => c.can_edit), "every own card is editable by its member").toBe(true);
    expect(idsOf(startA.cards)).not.toContain(DESIGN);
    expect(idsOf(startA.cards)).not.toContain(OPERATIONS);
    expect(await pageCardIds(page)).toEqual(idsOf(startA.cards));
    await expect(cardOf(page, RESEARCH).getByTestId("group-card-name")).toHaveText("Research");
    await expect(cardOf(page, RESEARCH).getByRole("textbox", { name: "Role in Research" })).toBeVisible();
    // W1: set, replace, an identical repeat (no write: the version stays), each visible on the next load and in the JSON.
    expect(await saveLabel(page, DAVE_ID, RESEARCH, first)).toBe("label_saved");
    await expectStatus(page, "label_saved");
    await expect(cardOf(page, RESEARCH).getByTestId("group-card-label")).toHaveText(first);
    const set = cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, RESEARCH);
    expect([set?.role_label, set?.version]).toEqual([first, (research0?.version ?? 0) + 1]);
    expect(await saveLabel(page, DAVE_ID, RESEARCH, second)).toBe("label_saved");
    await page.reload();
    await expect(cardOf(page, RESEARCH).getByTestId("group-card-label")).toHaveText(second);
    await expect(cardOf(page, RESEARCH).getByRole("textbox")).toHaveValue(second);
    const replaced = cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, RESEARCH);
    expect([replaced?.role_label, replaced?.version]).toEqual([second, (set?.version ?? 0) + 1]);
    expect(await saveLabel(page, DAVE_ID, RESEARCH, second)).toBe("label_saved");
    expect(cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, RESEARCH)?.version, "an identical save writes nothing").toBe(replaced?.version);

    // Host B: the same global header, B's own operations card without a label, nothing of A.
    await b.page.goto(profilePath(DAVE_ID));
    await expect(b.page.getByTestId("member-profile-name")).toHaveText(DAVE_NAME);
    const startB = await cardsJSON(b.context, DAVE_ID, "", ORIGIN_B);
    expect(startB.status).toBe(200);
    expect(cardIn(startB.cards, OPERATIONS), "dave's operations card on host B").toMatchObject({ role_label: null, can_edit: true });
    expect(idsOf(startB.cards)).not.toContain(RESEARCH);
    expect(startB.text).not.toContain(second);
    await expect(cardOf(b.page, OPERATIONS).getByTestId("group-card-label")).toHaveText(NO_ROLE);
    await expect(cardOf(b.page, RESEARCH)).toHaveCount(0);
    expect(await saveLabel(b.page, DAVE_ID, OPERATIONS, onB)).toBe("label_saved");
    await expect(cardOf(b.page, OPERATIONS).getByTestId("group-card-label")).toHaveText(onB);

    // Back on host A: research keeps its own label; no operations card, no B label anywhere.
    const backA = await cardsJSON(page.context(), DAVE_ID);
    expect(cardIn(backA.cards, RESEARCH)?.role_label).toBe(second);
    expect(idsOf(backA.cards)).not.toContain(OPERATIONS);
    expect(backA.text).not.toContain(onB);
    await page.goto(profilePath(DAVE_ID));
    expect(await page.content()).not.toContain(onB);

    // Clearing ("") removes the label: "No role given" and role_label null; the B label is untouched.
    expect(await saveLabel(page, DAVE_ID, RESEARCH, "")).toBe("label_cleared");
    await expectStatus(page, "label_cleared");
    await expect(cardOf(page, RESEARCH).getByTestId("group-card-label")).toHaveText(NO_ROLE);
    expect(cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, RESEARCH)?.role_label).toBeNull();
    expect(cardIn((await cardsJSON(b.context, DAVE_ID, "", ORIGIN_B)).cards, OPERATIONS)?.role_label).toBe(onB);
  } finally {
    await clearLabel(b.page, DAVE_ID, OPERATIONS, ORIGIN_B);
    await clearLabel(page, DAVE_ID, RESEARCH);
  }
});

test("[E-GMP-2] the API privacy matrix: self, shared-Group peers and Organization staff read only their cards; everyone else, foreign and anonymous get one 404", async ({ page, browser }, testInfo) => {
  test.slow();
  await signIn(page, DAVE);
  const bob = await signedIn(browser, testInfo, BOB);
  const carol = await signedIn(browser, testInfo, CAROL);
  const frank = await signedIn(browser, testInfo, FRANK);
  const alice = await signedIn(browser, testInfo, ALICE);
  const hostile = await signedIn(browser, testInfo, HOSTILE);
  // S1 self: every own Group of A, each editable.
  const own = await cardsJSON(page.context(), DAVE_ID);
  expect([own.status, own.cache, own.body.narrow_required]).toEqual([200, "no-store", false]);
  const daveGroups = idsOf(own.cards);
  expect(daveGroups).toContain(RESEARCH);
  expect(daveGroups).not.toContain(DESIGN);
  expect(own.cards.every((c) => c.can_edit)).toBe(true);
  // S2 shared Group, both directions: only Groups both belong to; bob (research owner) may edit dave's research label,
  // dave (a member) may not edit bob's.
  const bobToDave = await cardsJSON(bob.context, DAVE_ID);
  expect(bobToDave.status).toBe(200);
  expect(cardIn(bobToDave.cards, RESEARCH)?.can_edit, "bob (research owner) -> dave").toBe(true);
  for (const c of bobToDave.cards) expect(daveGroups, `bob sees dave's card for ${c.group_name} only through a shared Group`).toContain(c.group_id);
  const daveToBob = await cardsJSON(page.context(), BOB_ID);
  expect(daveToBob.status).toBe(200);
  expect(cardIn(daveToBob.cards, RESEARCH)?.can_edit, "dave (research member) -> bob").toBe(false);
  for (const c of daveToBob.cards) expect(daveGroups, `dave sees bob's card for ${c.group_name} only through a shared Group`).toContain(c.group_id);
  // S4 Organization oversight (P2, override off): the owner and an admin see every one of dave's Groups, read-only on
  // research (no Group role there).
  for (const [who, label] of [[carol.context, "carol (owner)"], [frank.context, "frank (admin)"]] as const) {
    const r = await cardsJSON(who, DAVE_ID);
    expect(r.status, label).toBe(200);
    expect(idsOf(r.cards), `${label}: every Group of dave's in A`).toEqual(daveGroups);
    expect(cardIn(r.cards, RESEARCH)?.can_edit, `${label}: research read-only`).toBe(false);
  }
  // N1 same tenant without a relationship, N7 foreign and unknown subjects: one 404, the body of /api/profiles/<id>'s.
  const n1 = await fetchFrom(alice.context, `/api/profiles/${DAVE_ID}`);
  expect(n1.status).toBe(404);
  const refused: string[] = [n1.text];
  for (const [who, label, id] of [
    [alice.context, "alice -> dave", DAVE_ID],
    [hostile.context, "hostile -> bob", BOB_ID],
    [hostile.context, "hostile -> bianca (B only)", BIANCA_ID],
    [alice.context, "alice -> cyrus (C only)", CYRUS_ID],
    [alice.context, "alice -> an unknown user", UNKNOWN_ID],
  ] as const) {
    const r = await cardsJSON(who, id);
    expect([r.status, r.cache], label).toEqual([404, "no-store"]);
    refused.push(r.text);
  }
  expect(new Set(refused).size, "one identical refusal body").toBe(1);
  expect(JSON.parse(refused[0])).toEqual({ error: "not_found" });
  // The pages: the same refusal is the fixed 404; a visible subject renders exactly the cards of the JSON.
  expect((await fetchFrom(alice.context, profilePath(DAVE_ID))).status).toBe(404);
  await bob.page.goto(profilePath(DAVE_ID));
  expect(await pageCardIds(bob.page)).toEqual(idsOf(bobToDave.cards));
  await expect(cardOf(bob.page, RESEARCH).getByTestId("group-label-form")).toHaveCount(1);
  await carol.page.goto(profilePath(DAVE_ID));
  expect(await pageCardIds(carol.page)).toEqual(daveGroups);
  await expect(cardOf(carol.page, RESEARCH).getByTestId("group-label-form")).toHaveCount(0);
  // Anonymous: 401 no-store.
  const anon = await open(browser, testInfo);
  const a = await fetchFrom(anon.context, `/api/profiles/${DAVE_ID}/groups`);
  expect([a.status, a.headers["cache-control"], JSON.parse(a.text)]).toEqual([401, "no-store", { error: "unauthenticated" }]);
  // No address, provider subject, storage key, Business Profile, member list or count in the page and the answers.
  await page.goto(profilePath(DAVE_ID));
  for (const text of [await page.content(), await bob.page.content(), own.text, bobToDave.text, daveToBob.text]) {
    // (Research Labs is the Business Profile research is onboarded with; a card never carries it.)
    for (const leak of ["identity.test", "user_01", "avatars/", '"email":', '\\"email\\":', "Research Labs", "member_count"]) {
      expect(text, `no ${leak}`).not.toContain(leak);
    }
  }
});

test("[E-GMP-3] a sibling Group shows only its own card; the selected-Group context from a members page narrows and refuses what it cannot select; revocation and readmission", async ({ page, browser }, testInfo) => {
  test.slow();
  const sfx = suffix(testInfo);
  const zLabel = `Z role ${sfx}`;
  await signIn(page, ALICE);
  const dave = await signedIn(browser, testInfo, DAVE);
  const bob = await signedIn(browser, testInfo, BOB);
  // alice (plain member, no Group) creates Z and admits dave: a Group sibling to research for him.
  const z = await createOwnGroup(page, `Cards Z ${sfx}`);
  let daveInZ = false;
  try {
    await accept(dave.page, await inviteToGroup(page, z, DAVE_NAME), z);
    daveInZ = true;
    // N2 sibling: alice sees dave's Z card only, never research; bob sees research only, never Z.
    const aliceView = await cardsJSON(page.context(), DAVE_ID);
    expect(aliceView.status).toBe(200);
    expect(idsOf(aliceView.cards), "alice -> dave: only the Group they share").toEqual([z]);
    expect(cardIn(aliceView.cards, z)?.can_edit, "alice owns Z").toBe(true);
    const bobView = await cardsJSON(bob.context, DAVE_ID);
    expect(idsOf(bobView.cards)).toContain(RESEARCH);
    expect(idsOf(bobView.cards)).not.toContain(z);

    // N3 entry from the members pages: a Group page's name link carries that Group as ?context=.
    await page.goto(membersOf(z));
    const zLink = memberRow(page, DAVE_NAME).getByTestId("member-profile-link");
    await expect(zLink).toHaveAttribute("href", `/profiles/${DAVE_ID}?context=${z}`);
    await zLink.click();
    await page.waitForURL((u) => u.pathname === `/profiles/${DAVE_ID}` && u.searchParams.get("context") === z);
    await expect(page.getByTestId("group-cards-context")).toHaveText(SELECTED_CONTEXT);
    expect(await pageCardIds(page)).toEqual([z]);
    await bob.page.goto(membersOf(RESEARCH));
    const researchLink = memberRow(bob.page, DAVE_NAME).getByTestId("member-profile-link");
    await expect(researchLink).toHaveAttribute("href", `/profiles/${DAVE_ID}?context=${RESEARCH}`);
    await researchLink.click();
    await bob.page.waitForURL((u) => u.pathname === `/profiles/${DAVE_ID}` && u.searchParams.get("context") === RESEARCH);
    expect(await pageCardIds(bob.page)).toEqual([RESEARCH]);
    const bobInResearch = await cardsJSON(bob.context, DAVE_ID, `?context=${RESEARCH}`);
    expect([bobInResearch.status, idsOf(bobInResearch.cards), bobInResearch.body.narrow_required]).toEqual([200, [RESEARCH], false]);

    // dave in Z's context: bob is visible (research) but holds no card in Z, so the list is empty; dave's own is Z only.
    const daveToBobZ = await cardsJSON(dave.context, BOB_ID, `?context=${z}`);
    expect([daveToBobZ.status, daveToBobZ.cards, daveToBobZ.body.narrow_required]).toEqual([200, [], false]);
    await dave.page.goto(profilePath(BOB_ID, z));
    await expect(dave.page.getByTestId("member-profile-name")).toHaveText(BOB_NAME);
    await expect(dave.page.getByTestId("group-cards-empty")).toHaveText(NO_CARDS);
    expect(idsOf((await cardsJSON(dave.context, DAVE_ID, `?context=${z}`)).cards)).toEqual([z]);

    // Selected-other denial: Z is not selectable for bob, so the page is the fixed 404 and the JSON the N1 body.
    const n1 = await fetchFrom(bob.context, `/api/profiles/${UNKNOWN_ID}`);
    const bobZ = await cardsJSON(bob.context, DAVE_ID, `?context=${z}`);
    expect([bobZ.status, bobZ.cache, bobZ.text]).toEqual([404, "no-store", n1.text]);
    expect((await fetchFrom(bob.context, profilePath(DAVE_ID, z))).status).toBe(404);
    // bob edits dave's research label with Z as the selector: the target is outside a context he cannot select (404).
    await bob.page.goto(profilePath(DAVE_ID));
    const bobCsrf = await csrfOf(bob.page);
    const before = cardIn((await cardsJSON(bob.context, DAVE_ID)).cards, RESEARCH);
    const mismatched = await forgeLabel(bob.context, bobCsrf, DAVE_ID, RESEARCH, before?.version ?? 0, `Forged ${sfx}`, [["context_group", z]]);
    expect(mismatched.status).toBe(404);
    expect(mismatched.text).toContain("Not found");
    expect(mismatched.text).not.toContain(`Forged ${sfx}`);
    expect(cardIn((await cardsJSON(bob.context, DAVE_ID)).cards, RESEARCH)).toEqual(before);

    // A malformed, empty or repeated selector: the fixed 404 page and a 400 from the JSON route, never Organization-only.
    for (const query of ["?context=not-a-uuid", "?context=", `?context=${z.toUpperCase()}`, `?context=${z}&context=${z}`]) {
      expect((await fetchFrom(dave.context, `/profiles/${DAVE_ID}${query}`)).status, `page ${query}`).toBe(404);
      const r = await cardsJSON(dave.context, DAVE_ID, query);
      expect([r.status, r.cache, JSON.parse(r.text)], `JSON ${query}`).toEqual([400, "no-store", { error: "invalid_request" }]);
    }
    // ... and in the label form: invalid_context, no write, the notice and no card at all.
    await dave.page.goto(profilePath(DAVE_ID));
    const daveCsrf = await csrfOf(dave.page);
    const invalid = await forgeLabel(dave.context, daveCsrf, DAVE_ID, z, 0, `Bad context ${sfx}`, [["context_group", "not-a-uuid"]]);
    expect([invalid.status, statusOf(invalid), new URL(invalid.url).searchParams.has("context")]).toEqual([200, "invalid_context", false]);
    await dave.page.goto(`${profilePath(DAVE_ID)}?group_status=invalid_context`);
    await expectStatus(dave.page, "invalid_context");
    await expect(dave.page.getByTestId("group-card")).toHaveCount(0);
    await expect(dave.page.getByTestId("group-cards-empty")).toHaveCount(0);
    expect(cardIn((await cardsJSON(dave.context, DAVE_ID)).cards, z)?.role_label).toBeNull();

    // N4 revocation: dave labels himself in Z's context; alice removes him; her next reads are refused and his Z card is
    // gone with its label.
    expect(await saveLabel(dave.page, DAVE_ID, z, zLabel, z)).toBe("label_saved");
    expect(cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, z)?.role_label).toBe(zLabel);
    await removeFromGroup(page, z, DAVE_NAME);
    daveInZ = false;
    const revoked = await cardsJSON(page.context(), DAVE_ID);
    expect([revoked.status, JSON.parse(revoked.text)]).toEqual([404, { error: "not_found" }]);
    expect((await fetchFrom(page.context(), profilePath(DAVE_ID))).status).toBe(404);
    expect(idsOf((await cardsJSON(dave.context, DAVE_ID)).cards)).not.toContain(z);
    // Readmission: a new membership starts without a label (no revival).
    await accept(dave.page, await inviteToGroup(page, z, DAVE_NAME), z);
    daveInZ = true;
    const readmitted = cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, z);
    expect([readmitted?.role_label, readmitted?.version]).toEqual([null, 0]);
  } finally {
    if (daveInZ) await removeFromGroup(page, z, DAVE_NAME);
  }
});

test("[E-GMP-4] Organization oversight is read-only under both override settings; an Organization role revoked is the next request's 404", async ({ page, browser }, testInfo) => {
  test.slow();
  const sfx = suffix(testInfo);
  await signIn(page, ALICE);
  const dave = await signedIn(browser, testInfo, DAVE);
  const carol = await signedIn(browser, testInfo, CAROL);
  const frank = await signedIn(browser, testInfo, FRANK);
  const frankCsrf = await csrfOf(frank.page);
  const daveGroups = idsOf((await cardsJSON(dave.context, DAVE_ID)).cards);
  const label0 = cardIn((await cardsJSON(dave.context, DAVE_ID)).cards, RESEARCH);
  let aliceAdmin = false;
  try {
    for (const state of ["off", "on"] as const) {
      if (state === "on") await setOverride(carol.page, "on");
      for (const [who, name] of [[frank, "frank (admin)"], [carol, "carol (owner)"]] as const) {
        const r = await cardsJSON(who.context, DAVE_ID);
        expect(r.status, `${name}, override ${state}`).toBe(200);
        expect(idsOf(r.cards), `${name}, override ${state}: every Group of dave's`).toEqual(daveGroups);
        expect(cardIn(r.cards, RESEARCH)?.can_edit, `${name}, override ${state}: no write path`).toBe(false);
        await who.page.goto(profilePath(DAVE_ID));
        await expect(cardOf(who.page, RESEARCH)).toHaveCount(1);
        await expect(cardOf(who.page, RESEARCH).getByTestId("group-label-form"), `${name}, override ${state}: read-only`).toHaveCount(0);
      }
      // frank's forged write with his own valid token: not permitted, nothing changed.
      const forged = await forgeLabel(frank.context, frankCsrf, DAVE_ID, RESEARCH, label0?.version ?? 0, `Staff ${sfx}`);
      expect([forged.status, statusOf(forged)], `frank's write, override ${state}`).toEqual([200, "not_permitted"]);
      expect(forged.text).not.toContain(`Staff ${sfx}`);
      expect(cardIn((await cardsJSON(dave.context, DAVE_ID)).cards, RESEARCH)).toEqual(label0);
    }
    await setOverride(carol.page, "off");

    // N5: alice (a plain member without a shared Group) is refused; as an Organization admin she reads dave's cards
    // read-only; demoted again, her next request is refused.
    expect((await cardsJSON(page.context(), DAVE_ID)).status).toBe(404);
    await setOrganizationRole(carol.page, ALICE_NAME, "Admin");
    aliceAdmin = true;
    const asAdmin = await cardsJSON(page.context(), DAVE_ID);
    expect(asAdmin.status).toBe(200);
    expect(idsOf(asAdmin.cards)).toEqual(daveGroups);
    expect(cardIn(asAdmin.cards, RESEARCH)?.can_edit).toBe(false);
    await page.goto(profilePath(DAVE_ID));
    expect(await pageCardIds(page)).toEqual(daveGroups);
    await setOrganizationRole(carol.page, ALICE_NAME, "Member");
    aliceAdmin = false;
    const demoted = await cardsJSON(page.context(), DAVE_ID);
    expect([demoted.status, JSON.parse(demoted.text)]).toEqual([404, { error: "not_found" }]);
    expect((await fetchFrom(page.context(), profilePath(DAVE_ID))).status).toBe(404);
  } finally {
    // The defaults for every later scenario and project, even after a failure above.
    await setOverride(carol.page, "off");
    if (aliceAdmin) await setOrganizationRole(carol.page, ALICE_NAME, "Member");
  }
});

test("[E-GMP-5] the research owner edits a member's label by keyboard from the members page link, in research's context; staff see it read-only", async ({ page, browser }, testInfo) => {
  test.slow();
  const owners = `Owner-set role ${suffix(testInfo)}`;
  await signIn(page, BOB);
  const dave = await signedIn(browser, testInfo, DAVE);
  const frank = await signedIn(browser, testInfo, FRANK);
  try {
    await page.goto(membersOf(RESEARCH));
    const link = memberRow(page, DAVE_NAME).getByTestId("member-profile-link");
    await expect(link).toHaveAttribute("href", `/profiles/${DAVE_ID}?context=${RESEARCH}`);
    await tabTo(page, link, `${DAVE_NAME}'s profile link`);
    await page.keyboard.press("Enter");
    await page.waitForURL((u) => u.pathname === `/profiles/${DAVE_ID}` && u.searchParams.get("context") === RESEARCH);
    await expect(page.getByTestId("group-cards-context")).toHaveText(SELECTED_CONTEXT);
    expect(await pageCardIds(page)).toEqual([RESEARCH]);
    const input = cardOf(page, RESEARCH).getByRole("textbox", { name: "Role in Research" });
    await tabTo(page, input, "Role in Research");
    await page.keyboard.press("ControlOrMeta+A");
    await page.keyboard.type(owners);
    await tabTo(page, cardOf(page, RESEARCH).getByRole("button", { name: "Save role", exact: true }), "Save role");
    await page.keyboard.press("Enter");
    await page.waitForURL((u) => u.searchParams.get("group_status") === "label_saved" && u.searchParams.get("context") === RESEARCH);
    await expectStatus(page, "label_saved");
    await expect(cardOf(page, RESEARCH).getByTestId("group-card-label")).toHaveText(owners);
    // dave reads the label his Group's owner gave him; frank (admin, no research role) sees it without a form.
    await dave.page.goto(profilePath(DAVE_ID));
    await expect(cardOf(dave.page, RESEARCH).getByTestId("group-card-label")).toHaveText(owners);
    await frank.page.goto(profilePath(DAVE_ID));
    await expect(cardOf(frank.page, RESEARCH).getByTestId("group-card-label")).toHaveText(owners);
    await expect(cardOf(frank.page, RESEARCH).getByTestId("group-label-form")).toHaveCount(0);
  } finally {
    await clearLabel(page, DAVE_ID, RESEARCH);
  }
});

test("[E-GMP-6] labels render as text and keep their bounds: markup escaped, 100 code points accepted, 101 and controls refused, supplementary characters counted as code points", async ({ page }, testInfo) => {
  test.slow();
  const sfx = suffix(testInfo);
  const markup = `<b>bold</b> & "quotes" ${sfx}`;
  await signIn(page, DAVE);
  try {
    expect(await saveLabel(page, DAVE_ID, RESEARCH, markup)).toBe("label_saved");
    const shown = cardOf(page, RESEARCH).getByTestId("group-card-label");
    await expect(shown).toHaveText(markup);
    await expect(shown.locator("b")).toHaveCount(0);
    expect(cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, RESEARCH)?.role_label).toBe(markup);
    // 101 code points: refused with nothing changed.
    expect(await saveLabel(page, DAVE_ID, RESEARCH, "x".repeat(101))).toBe("invalid_label");
    await expectStatus(page, "invalid_label");
    expect(cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, RESEARCH)?.role_label).toBe(markup);
    // A control character (set on the input directly; a keyboard cannot type it): refused.
    await page.goto(profilePath(DAVE_ID));
    await cardOf(page, RESEARCH)
      .getByRole("textbox")
      .evaluate((el) => {
        (el as HTMLInputElement).value = "Bell\u0007role";
      });
    await cardOf(page, RESEARCH).getByRole("button", { name: "Save role", exact: true }).click();
    await page.waitForURL((u) => u.searchParams.get("group_status") === "invalid_label");
    expect(cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, RESEARCH)?.role_label).toBe(markup);
    // Exactly 100 code points of ASCII, then 100 supplementary characters (200 UTF-16 units, the input's maxLength).
    const hundred = "y".repeat(100);
    expect(await saveLabel(page, DAVE_ID, RESEARCH, hundred)).toBe("label_saved");
    expect(cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, RESEARCH)?.role_label).toBe(hundred);
    const wide = "\u{1D538}".repeat(100);
    expect(wide.length).toBe(200);
    await expect(cardOf(page, RESEARCH).getByRole("textbox")).toHaveAttribute("maxlength", "200");
    expect(await saveLabel(page, DAVE_ID, RESEARCH, wide)).toBe("label_saved");
    const stored = cardIn((await cardsJSON(page.context(), DAVE_ID)).cards, RESEARCH)?.role_label ?? "";
    expect([stored, Array.from(stored).length]).toEqual([wide, 100]);
    await expect(cardOf(page, RESEARCH).getByTestId("group-card-label")).toHaveText(wide);
  } finally {
    await clearLabel(page, DAVE_ID, RESEARCH);
  }
});

test("[E-GMP-7] forged label writes change nothing: a wrong CSRF token, a cross-site page, a peer, an unrelated member and a stale version", async ({ page, context, browser }, testInfo) => {
  test.slow();
  tracking.allowed.add(ATTACKER);
  const sfx = suffix(testInfo);
  await signIn(page, DAVE);
  try {
    const own = cardIn((await cardsJSON(context, DAVE_ID)).cards, RESEARCH);
    const version = own?.version ?? 0;
    // A wrong CSRF token: the fixed refusal page.
    const forged = await forgeLabel(context, "forged", DAVE_ID, RESEARCH, version, "Mallory");
    expect(forged.status).toBe(403);
    expect(forged.text).toContain("Form expired");
    expect(forged.text).toContain("Nothing was changed.");
    // A cross-site page posting to the label form route is refused by origin.
    const xs = await context.newPage();
    await xs.goto(`${ATTACKER}/`);
    await xs.evaluate(
      ({ action, version }) => {
        const form = document.createElement("form");
        form.method = "post";
        form.action = action;
        for (const [name, value] of [["csrf_token", "guessed"], ["expected_version", version], ["role_label", "Mallory"]]) {
          const input = document.createElement("input");
          input.type = "hidden";
          input.name = name;
          input.value = value;
          form.append(input);
        }
        document.body.append(form);
        form.submit();
      },
      { action: `${ORIGIN_A}${labelPath(DAVE_ID, RESEARCH)}`, version: String(version) },
    );
    await xs.waitForURL((u) => u.origin === ORIGIN_A);
    await expect(xs.locator("h1")).toHaveText("Request refused");
    await xs.close();
    // dave (a research member) on bob's research label with his own valid token: not permitted.
    await page.goto(profilePath(DAVE_ID));
    const daveCsrf = await csrfOf(page);
    const bobBefore = cardIn((await cardsJSON(context, BOB_ID)).cards, RESEARCH);
    const peer = await forgeLabel(context, daveCsrf, BOB_ID, RESEARCH, bobBefore?.version ?? 0, `Peer ${sfx}`);
    expect([peer.status, statusOf(peer)]).toEqual([200, "not_permitted"]);
    expect(cardIn((await cardsJSON(context, BOB_ID)).cards, RESEARCH)).toEqual(bobBefore);
    // hostile (a plain member of A without any research role) on dave's research label: the fixed 404.
    const hostile = await signedIn(browser, testInfo, HOSTILE);
    const unrelated = await forgeLabel(hostile.context, await csrfOf(hostile.page), DAVE_ID, RESEARCH, version, `Hostile ${sfx}`);
    expect(unrelated.status).toBe(404);
    expect(unrelated.text).toContain("Not found");
    expect(cardIn((await cardsJSON(context, DAVE_ID)).cards, RESEARCH)).toEqual(own);
    // A stale tab: the first saves, the second tab's older form is a version conflict with the alert and no change.
    const stale = await context.newPage();
    await stale.goto(profilePath(DAVE_ID));
    expect(await saveLabel(page, DAVE_ID, RESEARCH, `Fresh ${sfx}`)).toBe("label_saved");
    await cardOf(stale, RESEARCH).getByRole("textbox").fill(`Old tab ${sfx}`);
    await cardOf(stale, RESEARCH).getByRole("button", { name: "Save role", exact: true }).click();
    await stale.waitForURL((u) => u.searchParams.has("group_status"));
    expect(new URL(stale.url()).searchParams.get("group_status")).toBe("stale_version");
    await expectStatus(stale, "stale_version");
    await expect(cardOf(stale, RESEARCH).getByTestId("group-card-label")).toHaveText(`Fresh ${sfx}`);
    await stale.close();
  } finally {
    await clearLabel(page, DAVE_ID, RESEARCH);
  }
});

test("[S-GMP-A11Y] the composite profile page by keyboard; axe at 1280 and 320 with an editable card, as a read-only viewer and in a selected context", async ({ page }, testInfo) => {
  test.slow();
  const keys = `Keys ${suffix(testInfo)}`;
  await signIn(page, DAVE);
  try {
    await page.goto(profilePath(DAVE_ID));
    const nav = page.getByRole("navigation", { name: "Primary" });
    await expect(nav.getByRole("link", { name: "Profile" })).toHaveAttribute("href", "#profile");
    await expect(nav.getByRole("link", { name: "Groups" })).toHaveAttribute("href", "#groups");
    await expect(page.getByRole("heading", { level: 1, name: DAVE_NAME, exact: true })).toBeVisible();
    await expect(page.getByRole("heading", { level: 2, name: "Groups" })).toBeVisible();
    await audit(page, "own profile with an editable card");
    const input = cardOf(page, RESEARCH).getByRole("textbox", { name: "Role in Research" });
    await tabTo(page, input, "Role in Research");
    await page.keyboard.press("ControlOrMeta+A");
    await page.keyboard.type(keys);
    await tabTo(page, cardOf(page, RESEARCH).getByRole("button", { name: "Save role", exact: true }), "Save role");
    await page.keyboard.press("Enter");
    await page.waitForURL((u) => u.searchParams.get("group_status") === "label_saved");
    await expectStatus(page, "label_saved");
    await expect(cardOf(page, RESEARCH).getByTestId("group-card-label")).toHaveText(keys);
    await audit(page, "own profile with a status notice");
    // A read-only viewer: dave on bob's page (research shared, no write path).
    await page.goto(profilePath(BOB_ID));
    await expect(cardOf(page, RESEARCH)).toHaveCount(1);
    await expect(page.getByTestId("group-label-form")).toHaveCount(0);
    await audit(page, "read-only cards");
    // The selected-Group context.
    await page.goto(profilePath(DAVE_ID, RESEARCH));
    await expect(page.getByTestId("group-cards-context")).toHaveText(SELECTED_CONTEXT);
    await audit(page, "selected-Group context");
  } finally {
    await clearLabel(page, DAVE_ID, RESEARCH);
  }
});
