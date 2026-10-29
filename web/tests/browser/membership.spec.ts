import { randomBytes, randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";

import AxeBuilder from "@axe-core/playwright";
import { type Browser, type BrowserContext, type Locator, type Page, type TestInfo } from "@playwright/test";

import { computedColor, cssToRgba, distance, hex, rgb } from "./support/color";
import { box, expectAxeClean } from "./support/contrast";
import { expectRadioOptions } from "./support/form-controls";
import { chooseOption, SELECT_POPUP } from "./support/modes";
import { expect, test } from "./support/page-errors";
import { observe } from "./support/sanitize";

// G019 invitations, membership and ownership (Plan r1 §1.6, §6.3 E-MEM-1..7; r2 addendum §7) in the disposable HTTPS
// identity environment (infra/compose/identity.yml), projects membership-light and membership-dark. The harness seeds the
// fixture first (infra/identity/mcp-seed.sql): carol owns Organization A, frank is its admin, alice a plain member of A,
// bob owns research (A), dave is a member of research; the G019 rows give every Organization and Group an owner (design
// is owned by a seeded user who never signs in) and add a second admin of A. Both projects run against the same database
// one after the other, so every scenario that changes membership or ownership acts on a Group the test itself creates
// (its creator is its owner, G018) with a per-run suffix, and the Organization-level scenario only issues and revokes
// its own invitation. Pages that answer an error status are reached only from extra pages, so the default page's
// S-ERRORS check covers every page the user sees succeed. Core decides every read and action; these tests observe the UI.
// G060 (Plan r0 §5.2 T11; r1 C7) adds E-MEM-8 (the owner-only Group-administration setting) and E-MEM-9 (override
// administration of a Group the test creates, and its revocation); each ends with the override off (the default).
// The G060 owner amendment (Option B) adds E-MEM-10: one generic invalid-invitation presentation, recipient success kept.

const ORIGIN_A = "https://org-a.localhost:543";
const EMULATOR = "http://emulator:4100";
const FORM = "application/x-www-form-urlencoded";
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
// G058 scope amendment04: an Organization invitation's link opens the enrollment entry /join/invitations/<id>; a Group
// invitation's link keeps /invitations/<id>. Each caller asserts its own path.
const LINK = /https:\/\/org-a\.localhost:543\/(?:join\/)?invitations\/([0-9a-f-]{36})#([A-Za-z0-9_-]{43})/;
const GROUP_LINK = /^https:\/\/org-a\.localhost:543\/invitations\//;
const ORGANIZATION_LINK = /^https:\/\/org-a\.localhost:543\/join\/invitations\//;
const AXE_TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

// Seeded rows (infra/identity/mcp-seed.sql).
const RESEARCH = "0f14a000-0000-4000-8000-0000000000a1";
const DESIGN = "0f14a000-0000-4000-8000-0000000000a2";

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
const ALICE_NAME = "Alice Example";

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
  const outside = [...tracking.origins].filter((o) => ![ORIGIN_A, EMULATOR].includes(o));
  expect(outside, "page requests outside the app origin under test and the emulator").toEqual([]);
});

async function open(browser: Browser, testInfo: TestInfo): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ baseURL: ORIGIN_A, colorScheme: testInfo.project.use.colorScheme, viewport: testInfo.project.use.viewport });
  track(context);
  tracking.contexts.push(context);
  return { context, page: await context.newPage() };
}

async function signIn(page: Page, who: Credentials) {
  await page.goto(`${ORIGIN_A}/sign-in?return_to=/account`);
  await page.getByRole("button", { name: "Sign in" }).click();
  await page.waitForURL((u) => u.origin === EMULATOR);
  await page.locator("#email").fill(who.email);
  await page.getByRole("button", { name: "Continue" }).click();
  const password = page.locator('input[name="password"]');
  await password.fill(who.password);
  await password.press("Enter");
  await page.waitForURL((u) => u.origin === ORIGIN_A && u.pathname === "/account");
}

// Same-origin requests from an extra page of the context, so an expected error status never reaches S-ERRORS.
type Answer = { status: number; redirected: boolean; text: string };
async function fromPage(context: BrowserContext, path: string, body?: string): Promise<Answer> {
  const p = await context.newPage();
  await p.goto(`${ORIGIN_A}/api/health`);
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

async function tabTo(page: Page, target: Locator, label: string) {
  for (let i = 0; i < 80 && !(await target.evaluate((t) => t === document.activeElement).catch(() => false)); i++) await page.keyboard.press("Tab");
  await expect(target, `keyboard focus on ${label}`).toBeFocused();
}

// The highlighted option of an open kit Select list: the kit item marks it (data-highlighted) and the list moves focus
// to it.
async function highlightedOption(popup: Locator): Promise<number> {
  return popup.getByRole("option").evaluateAll((els) => els.findIndex((el) => el.hasAttribute("data-highlighted") || el === document.activeElement));
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

// --- membership steps ---------------------------------------------------------------------------------------------------

// Creates a Group through G018's New Group flow (the creator becomes its owner) and returns its id; onboarding is not
// needed for membership.
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
// The Role cell only (the row's role selects also contain every role label).
const roleCell = (page: Page, name: string) => memberRow(page, name).getByRole("cell").nth(1);
const status = (page: Page, message: string) => expect(page.getByText(message, { exact: true })).toBeVisible();

// Reads the one-time link from the invitation result page; the token is registered with the report sanitizer.
async function issuedLink(page: Page): Promise<{ id: string; link: string; token: string }> {
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Invitation created");
  const text = await page.locator("main").innerText();
  const m = LINK.exec(text);
  expect(m, "the one-time invitation link").not.toBeNull();
  const [link, id, token] = m as RegExpExecArray;
  expect(token).toMatch(SECRET);
  observe(token);
  expect(await page.title()).not.toContain(token);
  expect(new URL(page.url()).hash).toBe("");
  return { id, link, token };
}

// The Group owner invites a co-member of the Organization by candidate and returns the one-time link.
async function inviteToGroup(page: Page, groupId: string, name: string, role: "Admin" | "Member"): Promise<string> {
  await page.goto(membersOf(groupId));
  const form = page.getByTestId("invitation-form");
  await chooseOption(form.getByLabel("Workspace member"), name);
  await chooseOption(form.getByLabel("Role"), role);
  await form.getByRole("button", { name: "Create invitation" }).click();
  const { link } = await issuedLink(page);
  expect(link, "a Group invitation keeps the member acceptance page").toMatch(GROUP_LINK);
  return link;
}

// The recipient opens the link (the fragment prefills the code and is removed) and accepts; returns the outcome status.
async function accept(page: Page, link: string, expected: "joined" | "already_member", groupId: string) {
  await page.goto(link);
  const code = page.getByLabel("Invitation code");
  await expect(code).not.toHaveValue("");
  await expect.poll(() => new URL(page.url()).hash).toBe("");
  await page.getByRole("button", { name: "Accept invitation" }).click();
  await page.waitForURL((u) => u.pathname === `/groups/${groupId}` && u.searchParams.get("status") === expected);
}

// A fresh Group owned by frank (default page) with alice as a Group admin who accepted through her own context.
async function groupWithAdmin(page: Page, browser: Browser, testInfo: TestInfo, label: string): Promise<{ groupId: string; alice: Page }> {
  await signIn(page, FRANK);
  const groupId = await createOwnGroup(page, `${label} ${suffix(testInfo)}`);
  const link = await inviteToGroup(page, groupId, ALICE_NAME, "Admin");
  const alice = await open(browser, testInfo);
  await signIn(alice.page, ALICE);
  await accept(alice.page, link, "joined", groupId);
  await page.goto(membersOf(groupId));
  await expect(roleCell(page, ALICE_NAME)).toHaveText("Admin");
  return { groupId, alice: alice.page };
}

// --- E-MEM-1..7 ---------------------------------------------------------------------------------------------------------

test("[E-MEM-1] the Organization owner invites by address, sees the link once, and revokes the pending invitation", async ({ page }, testInfo) => {
  const address = `invitee-${suffix(testInfo)}@example.test`;
  await signIn(page, CAROL);
  await page.goto("/organization/members");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Workspace members");
  const form = page.getByTestId("invitation-form");
  await form.getByLabel("Email address").fill(address);
  await chooseOption(form.getByLabel("Role"), "Member");
  await form.getByRole("button", { name: "Create invitation" }).click();
  const { link, token } = await issuedLink(page);
  expect(link, "an Organization invitation opens the enrollment entry").toMatch(ORGANIZATION_LINK);
  expect(await page.locator("main").innerText()).toContain("it is shown only once");
  await page.getByRole("link", { name: "Back to the members page" }).click();
  await page.waitForURL((u) => u.pathname === "/organization/members");
  // The members page never shows the token again; the pending row names the address.
  expect(await page.content()).not.toContain(token);
  const row = page.getByTestId("invitation-row").filter({ hasText: address });
  await expect(row).toContainText("pending", { ignoreCase: true });
  await row.getByRole("button", { name: `Revoke the invitation for ${address}` }).click();
  await page.waitForURL((u) => u.pathname === "/organization/members" && u.searchParams.get("status") === "revoked");
  await status(page, "The invitation was revoked.");
  await expect(page.getByTestId("invitation-row").filter({ hasText: address })).not.toContainText("pending", { ignoreCase: true });
});

test("[E-MEM-2] a Group owner invites a co-member who accepts through the link and appears in the Group; a second visit changes nothing", async ({ page, browser }, testInfo) => {
  await signIn(page, FRANK);
  const groupId = await createOwnGroup(page, `Invite ${suffix(testInfo)}`);
  const link = await inviteToGroup(page, groupId, ALICE_NAME, "Member");
  const alice = await open(browser, testInfo);
  await signIn(alice.page, ALICE);
  await alice.page.goto("/invitations");
  await expect(alice.page.getByTestId("my-invitations")).toContainText("Group invitation");
  // A second tab opens the same link while the invitation is still pending: its accept form is the stale duplicate
  // submission below (core answers a same-recipient replay already_member while the binding is current).
  const stale = await alice.context.newPage();
  await stale.goto(link);
  await expect(stale.getByLabel("Invitation code")).not.toHaveValue("");
  await expect.poll(() => new URL(stale.url()).hash).toBe("");
  await accept(alice.page, link, "joined", groupId);
  await expect(alice.page.getByText("You joined this Group.")).toBeVisible();
  await page.goto(membersOf(groupId));
  await expect(roleCell(page, ALICE_NAME)).toHaveText("Member");
  // Revisiting the consumed invitation states it and offers no acceptance form.
  await alice.page.goto(link);
  await expect(alice.page.getByText("This invitation can no longer be accepted (it has expired, was revoked or was already used).")).toBeVisible();
  await expect(alice.page.getByLabel("Invitation code")).toHaveCount(0);
  await expect(alice.page.getByRole("button", { name: "Accept invitation" })).toHaveCount(0);
  // The stale pending form submits the duplicate: nothing changes and the closed already_member status is shown.
  await stale.getByRole("button", { name: "Accept invitation" }).click();
  await stale.waitForURL((u) => u.pathname === `/groups/${groupId}` && u.searchParams.get("status") === "already_member");
  await expect(stale.getByText("You already have this role or a higher one in this Group, so nothing was changed.")).toBeVisible();
  await stale.close();
  await page.reload();
  await expect(roleCell(page, ALICE_NAME)).toHaveText("Member");
});

test("[E-MEM-3] the owner promotes an admin, demotes the new owner and transfers ownership", async ({ page, browser }, testInfo) => {
  const { groupId, alice } = await groupWithAdmin(page, browser, testInfo, "Owners");
  await memberRow(page, ALICE_NAME).getByRole("button", { name: `Make ${ALICE_NAME} an owner` }).click();
  await page.waitForURL((u) => u.pathname === membersOf(groupId) && u.searchParams.get("status") === "promoted");
  await status(page, "The member is now an owner.");
  await expect(roleCell(page, ALICE_NAME)).toHaveText("Owner");
  const row = memberRow(page, ALICE_NAME);
  await chooseOption(row.getByLabel(`Role for ${ALICE_NAME} after stepping down as owner`), "Admin");
  await row.getByRole("button", { name: "Remove owner role" }).click();
  await page.waitForURL((u) => u.searchParams.get("status") === "demoted");
  await status(page, "The owner's role was changed.");
  await expect(roleCell(page, ALICE_NAME)).toHaveText("Admin");
  // Transfer: the eligible admin is offered on the owner's paged picker; frank keeps the Member role.
  const transfer = page.getByTestId("transfer-form");
  await chooseOption(transfer.getByLabel("New owner"), ALICE_NAME);
  await chooseOption(transfer.getByLabel("Your role afterwards"), "Member");
  await transfer.getByRole("button", { name: "Transfer ownership" }).click();
  await page.waitForURL((u) => u.searchParams.get("status") === "transferred");
  await status(page, "Ownership was transferred.");
  await expect(page.getByTestId("self-card")).toContainText("Your role in this Group: Member");
  // The new owner sees the transfer from her own members page (frank, now a plain Member, no longer sees the list).
  await alice.goto(membersOf(groupId));
  await expect(alice.getByTestId("self-card")).toContainText("Your role in this Group: Owner");
  // G060 (Plan r0 §1.6, AC2; supersedes the G019 oracle here): frank's Group role is now Member and his Organization
  // admin role confers no Group management while the override is off (the default). He keeps his self card with Leave,
  // and sees no member list, no invitation form and no owner control.
  await expect(page.getByTestId("transfer-form")).toHaveCount(0);
  await expect(page.getByRole("button", { name: /^Make .+ an owner$/ })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Remove owner role" })).toHaveCount(0);
  await expect(page.getByTestId("member-table")).toHaveCount(0);
  await expect(page.getByTestId("invitation-form")).toHaveCount(0);
  await expect(page.getByTestId("self-card").getByRole("button", { name: "Leave" })).toBeVisible();
  await expect(page.getByTestId("override-admin-member-notice")).toHaveCount(0);
  await expect(page.getByText("Only the administrators of this Group can see and manage its members.")).toBeVisible();
  await page.goto("/organization/members");
  await expect(page.getByTestId("member-table")).toBeVisible();
  await expect(page.getByTestId("self-card")).toContainText("Admin");
});

test("[E-MEM-4] the last owner cannot leave and is told why; after promoting another owner, leaving succeeds", async ({ page, browser }, testInfo) => {
  await signIn(page, FRANK);
  const groupId = await createOwnGroup(page, `Leave ${suffix(testInfo)}`);
  await page.goto(membersOf(groupId));
  const card = page.getByTestId("self-card");
  await expect(card.getByTestId("last-owner-notice")).toHaveText(
    "You are the last owner of this Group, so you cannot leave yet. Make another member an owner or transfer ownership first.",
  );
  await expect(card.getByRole("button", { name: "Leave" })).toHaveCount(0);
  const link = await inviteToGroup(page, groupId, ALICE_NAME, "Admin");
  const alice = await open(browser, testInfo);
  await signIn(alice.page, ALICE);
  await accept(alice.page, link, "joined", groupId);
  await page.goto(membersOf(groupId));
  await memberRow(page, ALICE_NAME).getByRole("button", { name: `Make ${ALICE_NAME} an owner` }).click();
  await page.waitForURL((u) => u.searchParams.get("status") === "promoted");
  await expect(card.getByTestId("last-owner-notice")).toHaveCount(0);
  await card.getByRole("button", { name: "Leave" }).click();
  await page.waitForURL((u) => u.pathname === membersOf(groupId) && u.searchParams.get("status") === "left");
  await expect(page.getByText("You have left this Group.")).toBeVisible();
});

test("[E-MEM-5] an admin sees no owner controls, an Organization owner none on a Group it does not own, and a stale action changes nothing", async ({ page, context, browser }, testInfo) => {
  // frank (signed in by groupWithAdmin) administers Organization A without owning it.
  const { groupId, alice } = await groupWithAdmin(page, browser, testInfo, "Stale");
  await page.goto("/organization/members");
  await expect(page.getByTestId("member-table")).toBeVisible();
  await expect(page.getByTestId("transfer-form")).toHaveCount(0);
  await expect(page.getByRole("button", { name: /^Make .+ an owner$/ })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "Remove owner role" })).toHaveCount(0);
  await expect(memberRow(page, "Carol Example")).toContainText("None");
  // G060 (Plan r0 D-10, AC1/AC2; supersedes the G019 404 oracle for this case only): carol owns Organization A and frank
  // administers it, and neither holds a binding in design. design is readable metadata for both, so its members page
  // answers 200 with the truthful notice while the override is off (the default): no member list, no invitation form,
  // no self card, no Leave and no owner control.
  const carol = await open(browser, testInfo);
  await signIn(carol.page, CAROL);
  for (const [who, label] of [
    [carol.page, "carol (Organization owner) on design"],
    [page, "frank (Organization admin) on design"],
  ] as const) {
    await who.goto(membersOf(DESIGN));
    await expect(who.getByRole("heading", { level: 1 }), label).toHaveText("Members of Design");
    await expect(who.getByTestId("members-not-administered"), label).toBeVisible();
    for (const id of ["member-table", "invitation-form", "self-card", "transfer-form", "override-admin-notice"]) {
      await expect(who.getByTestId(id), `${label}: ${id}`).toHaveCount(0);
    }
    await expect(who.getByRole("button", { name: "Leave" }), label).toHaveCount(0);
  }
  // The restrictions stay the fixed 404 with no member list, self card or owner control (a 500 or unavailable page would
  // fail here): alice, a plain member of Organization A without a design binding; an absent Group; Organization B's
  // Group from Organization A's host, for carol and for frank.
  const OPERATIONS = "0f14b000-0000-4000-8000-0000000000b1"; // Organization B (infra/identity/mcp-seed.sql)
  const denied: [BrowserContext, string, string][] = [
    [alice.context(), DESIGN, "alice (plain member, no binding) on design"],
    [carol.context, "0f14a000-0000-4000-8000-0000000000ff", "an absent Group"],
    [carol.context, OPERATIONS, "carol on Organization B's Group"],
    [context, OPERATIONS, "frank on Organization B's Group"],
  ];
  for (const [who, id, label] of denied) {
    const answer = await fromPage(who, membersOf(id));
    expect(answer.status, label).toBe(404);
    expect(answer.redirected, label).toBe(false);
    for (const control of ['data-testid="member-table"', 'data-testid="self-card"', 'data-testid="transfer-form"', 'data-testid="invitation-form"', "Remove owner role", 'an owner"']) {
      expect(answer.text, `${label}: ${control}`).not.toContain(control);
    }
  }
  // Stale action: a second tab still shows alice after the first removed her; its change answers a fixed page (an
  // extra page, so its error status stays outside S-ERRORS).
  await page.goto(membersOf(groupId));
  const stale = await context.newPage();
  await stale.goto(membersOf(groupId));
  await memberRow(page, ALICE_NAME).getByRole("button", { name: `Remove ${ALICE_NAME}` }).click();
  await page.waitForURL((u) => u.searchParams.get("status") === "removed");
  await status(page, "The member was removed.");
  await expect(memberRow(page, ALICE_NAME)).toHaveCount(0);
  const row = memberRow(stale, ALICE_NAME);
  await chooseOption(row.getByLabel(`New role for ${ALICE_NAME}`), "Member");
  await row.getByRole("button", { name: "Change role" }).click();
  await expect(stale.getByRole("heading", { level: 1 })).not.toHaveText("Workspace members");
  await expect(stale.locator("main")).toContainText("Nothing was changed.");
  await expect(stale.getByRole("link", { name: "Back to the members page" })).toBeVisible();
  await stale.close();
  await page.reload();
  await expect(memberRow(page, ALICE_NAME)).toHaveCount(0);
});

test("[E-MEM-6] keyboard-only Group invitation and acceptance; axe at 1280 and 320 on every membership page", async ({ page, browser }, testInfo) => {
  const frank = await open(browser, testInfo);
  await signIn(frank.page, FRANK);
  const groupId = await createOwnGroup(frank.page, `Keys ${suffix(testInfo)}`);
  await frank.page.goto(membersOf(groupId));
  await audit(frank.page, "Group members (owner)");
  const form = frank.page.getByTestId("invitation-form");
  const recipient = form.getByLabel("Workspace member");
  await tabTo(frank.page, recipient, "Workspace member");
  // The kit Select by keyboard (G069 r3; the browser's list and its type-ahead are gone): ArrowDown opens the list with
  // the chosen candidate highlighted, the arrow keys move the highlight one option per press, Enter chooses and closes,
  // and focus returns to the trigger, which shows the chosen name.
  const popup = frank.page.locator(SELECT_POPUP);
  await frank.page.keyboard.press("ArrowDown");
  await expect(popup).toBeVisible();
  const names = (await popup.getByRole("option").allTextContents()).map((t) => t.trim());
  const want = names.indexOf(ALICE_NAME);
  expect(want, `${ALICE_NAME} among the candidates ${JSON.stringify(names)}`).toBeGreaterThanOrEqual(0);
  await expect.poll(() => highlightedOption(popup), { message: "an option is highlighted once the list opens" }).toBeGreaterThanOrEqual(0);
  for (let at = await highlightedOption(popup); at !== want; at = await highlightedOption(popup)) {
    const next = want > at ? at + 1 : at - 1;
    await frank.page.keyboard.press(want > at ? "ArrowDown" : "ArrowUp");
    await expect.poll(() => highlightedOption(popup), { message: `the highlight moves to ${names[next]}` }).toBe(next);
  }
  await frank.page.keyboard.press("Enter");
  await expect(popup).toBeHidden();
  await expect(recipient).toHaveText(ALICE_NAME, { useInnerText: true });
  await expect(recipient).toBeFocused();
  await tabTo(frank.page, form.getByRole("button", { name: "Create invitation" }), "Create invitation");
  await frank.page.keyboard.press("Enter");
  const { link } = await issuedLink(frank.page);
  await audit(frank.page, "invitation created");
  // alice (the default page, under S-ERRORS) accepts by keyboard.
  await signIn(page, ALICE);
  await page.goto("/invitations");
  await audit(page, "your invitations");
  await page.goto(link);
  await expect(page.getByLabel("Invitation code")).not.toHaveValue("");
  await audit(page, "invitation");
  await tabTo(page, page.getByRole("button", { name: "Accept invitation" }), "Accept invitation");
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.pathname === `/groups/${groupId}` && u.searchParams.get("status") === "joined");
  await page.goto("/organization/members");
  await audit(page, "workspace members (member)");
  await page.goto(membersOf(groupId));
  await audit(page, "Group members (member)");
});

test("[E-MEM-7] membership navigation leaves the member's MCP connections unchanged", async ({ page }) => {
  await signIn(page, DAVE);
  const connections = async () => {
    await page.goto("/connections/mcp");
    await expect(page.getByRole("heading", { level: 1 })).toHaveText("Your AI client connections");
    return page.getByTestId("mcp-connection").evaluateAll((rows) => rows.map((r) => `${r.getAttribute("data-state")} ${r.textContent}`));
  };
  const before = await connections();
  expect(before.length).toBeGreaterThan(0);
  await page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link", { name: "Members" }).click();
  await page.waitForURL((u) => u.pathname === "/organization/members");
  await expect(page.getByTestId("self-card")).toBeVisible();
  await page.goto("/invitations");
  await page.goto(`/groups/${RESEARCH}`);
  await page.getByRole("main").getByRole("link", { name: "Members", exact: true }).click();
  await page.waitForURL((u) => u.pathname === membersOf(RESEARCH));
  expect(await connections()).toEqual(before);
});

// --- G069 r3 (owner finding) S-SELECT-KIT: every form dropdown is the kit Select ----------------------------------------

// The members page's role dropdown stands for every form dropdown (components/controls/form-select over the kit Select,
// never the browser's option list): a combobox named by the row's visible label whose closed trigger shows the chosen
// label in the kit's own field treatment (G069 owner direction, exact kit: a transparent 1px border on the kit's
// bg-input/50 fill, superseding the app's solid input boundary of S22); its popup is the kit's select-content in the
// document body, its items
// inside the kit's padded group and inset >= 4px from every popup edge, the chosen item marked with the kit's check
// indicator, ground and ink the palette's popover pair of this project's appearance (the light and dark projects both
// run it); the kit's hidden input posts name=value with the chosen value, and the form's submission changes the role.
test("[S-SELECT-KIT] the role dropdown is the kit Select: labelled trigger, inset items, check mark, popover palette, posted value", async ({ page, browser }, testInfo) => {
  const { groupId } = await groupWithAdmin(page, browser, testInfo, "Kit select");
  const row = memberRow(page, ALICE_NAME);
  const form = row.locator("form", { has: page.getByRole("button", { name: "Change role" }) });
  const trigger = row.getByLabel(`New role for ${ALICE_NAME}`);
  const token = (name: string) => page.evaluate((n) => getComputedStyle(document.documentElement).getPropertyValue(n).trim(), name);
  const near = async (got: string, want: string, what: string) => {
    const g = rgb(await cssToRgba(page, got));
    const w = rgb(await cssToRgba(page, want));
    expect(distance(g, w), `${what}: ${hex(g)} vs ${hex(w)}`).toBeLessThanOrEqual(1);
  };
  expect(await row.locator("select").count(), "no native select in the row").toBe(0);
  await expect(trigger, "the row's label names the kit trigger, a combobox").toHaveAttribute("role", "combobox");
  await expect(trigger).toHaveAttribute("data-slot", "select-trigger");
  await expect(trigger, "the closed trigger shows the chosen label").toHaveText("Member", { useInnerText: true });
  // The kit trigger exactly: its 1px border transparent (border-transparent) and its fill the input token at 50 % mixed
  // in oklab with transparent (bg-input/50), as the browser computes that expression beside the trigger.
  const closed = await trigger.evaluate((el) => {
    const probe = document.createElement("span");
    el.parentElement!.append(probe);
    probe.style.backgroundColor = "color-mix(in oklab, var(--input) 50%, transparent)";
    const kitFill = getComputedStyle(probe).backgroundColor;
    probe.remove();
    const c = getComputedStyle(el);
    return { border: c.borderTopColor, width: c.borderTopWidth, fill: c.backgroundColor, kitFill };
  });
  expect((await cssToRgba(page, closed.border))[3], `closed trigger: the kit's transparent border (${closed.border})`).toBe(0);
  expect(closed.width, "closed trigger: the kit's 1px border").toBe("1px");
  const [fill, kitFill] = [await cssToRgba(page, closed.fill), await cssToRgba(page, closed.kitFill)];
  expect(distance(rgb(fill), rgb(kitFill)), `closed trigger: the kit's bg-input/50 fill (${closed.fill} vs ${closed.kitFill})`).toBeLessThanOrEqual(1);
  expect(Math.abs(fill[3] - kitFill[3]), `closed trigger: the fill's alpha (${closed.fill} vs ${closed.kitFill})`).toBeLessThanOrEqual(1);
  // The hidden input the kit posts: name=role with the chosen role's value.
  const posted = () => form.evaluate((f) => String(new FormData(f as HTMLFormElement).get("role")));
  expect(await posted(), "the default choice").toBe("group_member");
  const popup = page.locator(SELECT_POPUP);
  await expect(popup).toHaveCount(0);
  await trigger.click();
  await expect(popup).toBeVisible();
  await expect(popup).toHaveCount(1);
  expect(await popup.evaluate((el) => el.closest("form") === null), "the popup is portalled out of the form").toBe(true);
  const group = popup.locator("[data-slot=select-group]");
  await expect(group).toHaveCount(1);
  const items = group.locator("[data-slot=select-item]");
  await expect(items).toHaveText(["Admin", "Member"]);
  const selected = popup.getByRole("option", { name: "Member", exact: true });
  await expect(selected).toHaveAttribute("aria-selected", "true");
  await expect(selected.locator("svg.lucide-check"), "the chosen item carries the kit's check indicator").toHaveCount(1);
  await expect(popup.getByRole("option", { name: "Admin", exact: true }).locator("svg.lucide-check")).toHaveCount(0);
  // The items are inset from every popup edge by the kit group's padding (p-1: 4px), so none touches the edge. The kit
  // popup's open motion (fade and zoom-in-95, 100ms) scales its box, so it is measured once that has settled.
  await popup.evaluate((e) => Promise.allSettled(e.getAnimations({ subtree: true }).map((a) => a.finished)));
  const pb = await box(popup);
  const first = await box(items.first());
  const last = await box(items.last());
  expect(first.y - pb.y, "first item inset from the top edge").toBeGreaterThanOrEqual(4);
  expect(first.x - pb.x, "first item inset from the left edge").toBeGreaterThanOrEqual(4);
  expect(pb.x + pb.width - (last.x + last.width), "last item inset from the right edge").toBeGreaterThanOrEqual(4);
  expect(pb.y + pb.height - (last.y + last.height), "last item inset from the bottom edge").toBeGreaterThanOrEqual(4);
  // The popover palette: the popup's ground and ink, and an item at rest (the first, never the highlighted chosen one).
  const surface = await popup.evaluate((el) => ({ background: getComputedStyle(el).backgroundColor, color: getComputedStyle(el).color }));
  await near(surface.background, await token("--popover"), "popup ground (popover)");
  await near(surface.color, await token("--popover-foreground"), "popup ink (popover-foreground)");
  await near(await computedColor(items.first(), "color"), await token("--popover-foreground"), "item ink (popover-foreground)");
  // A pointer choice closes the list; the trigger and the hidden input follow it.
  await popup.getByRole("option", { name: "Admin", exact: true }).click();
  await expect(popup).toBeHidden();
  await expect(trigger).toHaveText("Admin", { useInnerText: true });
  expect(await posted(), "the chosen value").toBe("group_admin");
  // Back to Member through the shared helper; the form posts it and core changes alice's role.
  await chooseOption(trigger, "Member");
  expect(await posted()).toBe("group_member");
  await form.getByRole("button", { name: "Change role" }).click();
  await page.waitForURL((u) => u.pathname === membersOf(groupId) && u.searchParams.get("status") === "role_changed");
  await status(page, "The member's role was changed.");
  await expect(roleCell(page, ALICE_NAME)).toHaveText("Member");
});

// --- G060 E-MEM-8..9: the owner-only Group-administration setting and override administration -------------------------

const GROUP_AUTHORITY = "/organization/group-authority";
const OVERRIDE_ON = "Group administration by Organization staff is now on.";
const OVERRIDE_OFF = "Group administration by Organization staff is now off. Organization owners and admins keep only the Group roles they actually hold.";

// The CSRF token of the current page's forms (registered with the report sanitizer).
async function csrfOf(page: Page): Promise<string> {
  const value = await page.locator('input[name="csrf_token"]').first().inputValue();
  expect(value).toMatch(SECRET);
  observe(value);
  return value;
}

// carol (the Organization owner, signed in on OWNER) sets the override through the settings form and sees core's state.
async function setOverride(owner: Page, state: "on" | "off") {
  await owner.goto(GROUP_AUTHORITY);
  const form = owner.getByTestId("group-authority-form");
  await form.getByTestId(`group-authority-${state}`).check();
  await form.getByRole("button", { name: "Save" }).click();
  await owner.waitForURL((u) => u.pathname === GROUP_AUTHORITY && u.searchParams.get("status") === (state === "on" ? "enabled" : "disabled"));
  await status(owner, state === "on" ? OVERRIDE_ON : OVERRIDE_OFF);
  await expect(owner.getByTestId(`group-authority-${state}`)).toBeChecked();
}

test("[E-MEM-8] only the Organization owner changes Group administration (by keyboard); an admin reads it, a member gets the fixed 404 and a forged change is refused", async ({ page, browser }, testInfo) => {
  const carol = await open(browser, testInfo);
  await signIn(carol.page, CAROL);
  try {
    // The owner reaches the page from her account and from the members page; the default is off.
    await expect(carol.page.getByTestId("account-group-authority-link")).toBeVisible();
    await carol.page.goto("/organization/members");
    await carol.page.getByTestId("members-group-authority-link").click();
    await carol.page.waitForURL((u) => u.pathname === GROUP_AUTHORITY);
    await expect(carol.page.getByRole("heading", { level: 1 })).toHaveText("Group administration");
    await expect(carol.page.getByTestId("group-authority-off")).toBeChecked();
    const fieldset = carol.page.getByRole("group", { name: "Group administration by Organization staff" });
    expect(await fieldset.getAttribute("aria-describedby")).toBe("group-authority-scope");
    await expect(carol.page.getByTestId("group-authority-scope")).toContainText("never gives access to a Group's content or files");
    await expect(carol.page.getByTestId("group-authority-scope")).toContainText("can grant access to other people");
    await audit(carol.page, "Group administration (owner)");
    // G068/G069: the kit radio treatment, label alignment, form value and focus state, in this appearance and at 320px
    // (the On option wraps).
    await expectRadioOptions(carol.page, testInfo, fieldset, "Group administration", "enabled", true);
    // Keyboard only: focus the selected Off radio, arrow to On, then Save.
    await tabTo(carol.page, carol.page.getByTestId("group-authority-off"), "Off");
    await carol.page.keyboard.press("ArrowDown");
    await expect(carol.page.getByTestId("group-authority-on")).toBeChecked();
    await tabTo(carol.page, carol.page.getByRole("button", { name: "Save" }), "Save");
    await carol.page.keyboard.press("Enter");
    await carol.page.waitForURL((u) => u.searchParams.get("status") === "enabled");
    await status(carol.page, OVERRIDE_ON);
    await expect(carol.page.getByTestId("group-authority-on")).toBeChecked();

    // frank, an Organization admin (the default page, under S-ERRORS): the effective state read-only, no form.
    await signIn(page, FRANK);
    await expect(page.getByTestId("account-group-authority-link")).toBeVisible();
    await page.goto(GROUP_AUTHORITY);
    await expect(page.getByTestId("group-authority-readonly")).toBeVisible();
    await expect(page.getByTestId("group-authority-state")).toHaveText("Current setting: On: Organization owners and admins may also administer every Group.");
    await expect(page.getByTestId("group-authority-form")).toHaveCount(0);
    await expect(page.getByRole("radio")).toHaveCount(0);
    await audit(page, "Group administration (admin)");
    // A forged change by the admin with his own valid CSRF token: "Not permitted", nothing changed.
    await page.goto("/organization/members");
    const csrf = await csrfOf(page);
    const forged = await fromPage(page.context(), `${GROUP_AUTHORITY}/action`, new URLSearchParams({ csrf_token: csrf, enabled: "off" }).toString());
    expect(forged.status).toBe(403);
    expect(forged.text).toContain("Not permitted");
    expect(forged.text).toContain("Nothing was changed.");
    await carol.page.reload();
    await expect(carol.page.getByTestId("group-authority-on")).toBeChecked();

    // alice, a plain member: no link and the fixed 404.
    const alice = await open(browser, testInfo);
    await signIn(alice.page, ALICE);
    await expect(alice.page.getByTestId("account-group-authority-link")).toHaveCount(0);
    await alice.page.goto("/organization/members");
    await expect(alice.page.getByTestId("members-group-authority-link")).toHaveCount(0);
    const memberRead = await fromPage(alice.context, GROUP_AUTHORITY);
    expect(memberRead.status).toBe(404);
    expect(memberRead.text).not.toContain("group-authority-form");
  } finally {
    // The default (off) for every later scenario and project, even after a failure above.
    await setOverride(carol.page, "off");
  }
});

test("[E-MEM-9] with the override on an Organization admin administers a Group without joining it; turning it off removes only that path", async ({ page, browser }, testInfo) => {
  const sfx = suffix(testInfo);
  // A Group of bob's (a plain Organization member), in which frank and carol hold no binding.
  const bob = await open(browser, testInfo);
  await signIn(bob.page, BOB);
  const groupId = await createOwnGroup(bob.page, `Override ${sfx}`);
  const carol = await open(browser, testInfo);
  await signIn(carol.page, CAROL);
  await signIn(page, FRANK);
  try {
    // Off (the default): metadata only.
    await page.goto(`/groups/${groupId}`);
    await expect(page.getByRole("main").getByRole("link", { name: "Members", exact: true })).toHaveCount(0);
    await page.goto(membersOf(groupId));
    await expect(page.getByTestId("members-not-administered")).toBeVisible();

    await setOverride(carol.page, "on");
    // On: frank administers the Group's members through the setting; he is told he is not a member and gets no self
    // card, no Leave and no ownership control.
    await page.goto(`/groups/${groupId}`);
    await page.getByRole("main").getByRole("link", { name: "Members", exact: true }).click();
    await page.waitForURL((u) => u.pathname === membersOf(groupId));
    await expect(page.getByTestId("override-admin-notice")).toHaveText(
      "You administer this Group's members through the Organization setting Group administration by Organization staff. You are not a member of this Group.",
    );
    await expect(page.getByTestId("member-table")).toBeVisible();
    await expect(page.getByTestId("invitation-form")).toBeVisible();
    for (const id of ["self-card", "transfer-form", "members-not-administered", "override-admin-member-notice"]) {
      await expect(page.getByTestId(id), id).toHaveCount(0);
    }
    await expect(page.getByRole("button", { name: "Leave" })).toHaveCount(0);
    await expect(page.getByRole("button", { name: /^Make .+ an owner$/ })).toHaveCount(0);
    await expect(page.getByRole("button", { name: "Remove owner role" })).toHaveCount(0);
    await audit(page, "Group members (override administrator)");
    // frank invites alice; alice joins through her own link.
    const link = await inviteToGroup(page, groupId, ALICE_NAME, "Member");
    const alice = await open(browser, testInfo);
    await signIn(alice.page, ALICE);
    await accept(alice.page, link, "joined", groupId);
    await page.goto(membersOf(groupId));
    await expect(roleCell(page, ALICE_NAME)).toHaveText("Member");

    // An actual member who also administers through the setting: bob invites frank as Member; frank's own card and
    // Leave come from his real role, with the one-line notice beside the administration.
    const frankLink = await inviteToGroup(bob.page, groupId, "Frank Example", "Member");
    await accept(page, frankLink, "joined", groupId);
    await page.goto(membersOf(groupId));
    await expect(page.getByTestId("self-card")).toContainText("Your role in this Group: Member");
    await expect(page.getByTestId("self-card").getByRole("button", { name: "Leave" })).toBeVisible();
    await expect(page.getByTestId("override-admin-member-notice")).toHaveText("You also administer this Group's members through the Organization setting.");
    await expect(page.getByTestId("member-table")).toBeVisible();
    await expect(page.getByTestId("override-admin-notice")).toHaveCount(0);
    await expect(page.getByTestId("transfer-form")).toHaveCount(0);

    // Off again: frank's next request has no administration path; his genuine membership and alice's survive.
    await setOverride(carol.page, "off");
    await page.reload();
    await expect(page.getByTestId("self-card")).toContainText("Your role in this Group: Member");
    await expect(page.getByTestId("member-table")).toHaveCount(0);
    await expect(page.getByTestId("invitation-form")).toHaveCount(0);
    await expect(page.getByTestId("override-admin-member-notice")).toHaveCount(0);
    await page.goto(membersOf(DESIGN));
    await expect(page.getByTestId("members-not-administered")).toBeVisible();
    await alice.page.goto(membersOf(groupId));
    await expect(alice.page.getByTestId("self-card")).toContainText("Your role in this Group: Member");
  } finally {
    await setOverride(carol.page, "off");
  }
});

// --- G060 owner amendment (Option B) E-MEM-10: one generic invalid-invitation presentation ------------------------------

const INVALID_INVITATION = "Invalid invitation link. Switch to a different account, or ask for a new invitation link.";

test("[E-MEM-10] a wrong account, a nonexistent and a malformed invitation id show one generic invalid-link message; the intended recipient still accepts", async ({ page, browser }, testInfo) => {
  await signIn(page, FRANK);
  const groupId = await createOwnGroup(page, `Invalid link ${suffix(testInfo)}`);
  const link = await inviteToGroup(page, groupId, ALICE_NAME, "Member");
  // dave: a plain member of Organization A who is neither the recipient nor an administrator of frank's Group.
  const dave = await open(browser, testInfo);
  await signIn(dave.page, DAVE);
  const seen: string[] = [];
  for (const [label, target] of [
    ["wrong account", link],
    ["nonexistent id", `/invitations/${randomUUID()}`],
    ["malformed id", "/invitations/not-an-invitation"],
  ] as const) {
    const answer = await dave.page.goto(target);
    expect(answer?.status(), label).toBe(200);
    await expect(dave.page.getByRole("heading", { level: 1 }), label).toHaveText("Invitation");
    await expect(dave.page.getByTestId("status-message"), label).toHaveText(INVALID_INVITATION);
    await expect(dave.page.getByLabel("Invitation code"), label).toHaveCount(0);
    await expect(dave.page.getByRole("button", { name: "Accept invitation" }), label).toHaveCount(0);
    const main = await dave.page.getByRole("main").innerText();
    for (const secret of [ALICE_NAME, ALICE.email, "alice"]) expect(main.toLowerCase(), `${label} discloses no recipient`).not.toContain(secret.toLowerCase());
    seen.push(main);
  }
  expect(new Set(seen).size, "one identical presentation for the three cases").toBe(1);
  await audit(dave.page, "invalid invitation");
  // The intended recipient accepts through the same link.
  const alice = await open(browser, testInfo);
  await signIn(alice.page, ALICE);
  await accept(alice.page, link, "joined", groupId);
  await expect(alice.page.getByText("You joined this Group.")).toBeVisible();
  await page.goto(membersOf(groupId));
  await expect(roleCell(page, ALICE_NAME)).toHaveText("Member");
});

// --- members-page landmarks and navigation (S-LANDMARKS and S-NAV with the Members area) ---------------------------------

const MEMBERSHIP_PAGES: readonly { path: string; area: "groups" | "members" }[] = [
  { path: "/organization/members", area: "members" },
  { path: "/invitations", area: "members" },
  { path: membersOf(RESEARCH), area: "groups" },
];
const AREA_HREF = { groups: "/groups", members: "/organization/members" } as const;

test("[S-LANDMARKS] membership pages: one banner, Primary and Workspace areas navigation, main and contentinfo; heading order", async ({ page }) => {
  await signIn(page, BOB);
  for (const { path, area } of MEMBERSHIP_PAGES) {
    await page.goto(path);
    await expect(page.getByRole("banner"), path).toHaveCount(1);
    await expect(page.getByRole("navigation", { name: "Primary" }), path).toHaveCount(1);
    await expect(page.getByRole("navigation", { name: "Workspace areas" }), path).toHaveCount(1);
    await expect(page.getByRole("main"), path).toHaveCount(1);
    await expect(page.getByRole("contentinfo"), path).toHaveCount(1);
    const areas = page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link");
    await expect(areas, path).toHaveText(["Groups", "Business Profiles", "Members"]);
    const current = await areas.evaluateAll((as) => as.filter((a) => a.getAttribute("aria-current") === "page").map((a) => a.getAttribute("href")));
    expect(current, path).toEqual([AREA_HREF[area]]);
    const levels = await page.locator("h1, h2, h3, h4, h5, h6").evaluateAll((hs) => hs.map((h) => Number(h.tagName[1])));
    expect(levels.filter((l) => l === 1), path).toHaveLength(1);
    expect(levels[0], path).toBe(1);
    levels.forEach((level, i) => {
      if (i > 0) expect(level - levels[i - 1], `${path} heading order ${levels.join(",")}`).toBeLessThanOrEqual(1);
    });
  }
});

test("[S-NAV] membership pages: Primary links reach their sections and the Members area link its page, by keyboard", async ({ page }) => {
  await signIn(page, BOB);
  for (const { path } of MEMBERSHIP_PAGES) {
    await page.goto(path);
    const links = page.getByRole("navigation", { name: "Primary" }).getByRole("link");
    const count = await links.count();
    expect(count, path).toBeGreaterThan(0);
    for (let i = 0; i < count; i++) {
      const href = await links.nth(i).getAttribute("href");
      expect(href, path).toMatch(/^#[a-z-]+$/);
      await links.nth(i).focus();
      await page.keyboard.press("Enter");
      await expect(page).toHaveURL(new RegExp(`${href}$`));
      await expect(page.locator(`${href} h1, ${href} h2`).first(), `${path} ${href}`).toBeInViewport();
    }
  }
  await page.goto("/groups");
  const members = page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link", { name: "Members" });
  await members.focus();
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.pathname === "/organization/members");
  await expect(page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link", { name: "Members" })).toHaveAttribute("aria-current", "page");
});
