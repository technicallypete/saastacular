import { randomBytes } from "node:crypto";
import { readFileSync } from "node:fs";

import AxeBuilder from "@axe-core/playwright";
import { type Browser, type BrowserContext, type Locator, type Page, type TestInfo } from "@playwright/test";

import { expectAxeClean } from "./support/contrast";
import { expectRadioOptions } from "./support/form-controls";
import { chooseOption } from "./support/modes";
import { expect, test } from "./support/page-errors";
import { observe } from "./support/sanitize";

// G020 generic labels and visibility (Plan r2 §1.5, §5.2 T12 E-LAB-1..6, D-7, D-8, D-10) in the disposable HTTPS identity
// environment (infra/compose/identity.yml), projects labels-light and labels-dark. The harness seeds the fixture first
// (infra/identity/mcp-seed.sql): carol owns Organization A, frank is its admin, alice a plain member of A, bob owns
// research (A), dave is a member of research. Both projects run against the same database one after the other, so every
// label a scenario creates carries a per-run suffix and is deleted again (in a finally block where a failure could leave
// it behind), the Group-administration override always ends off (the default), and the scenario that removes a Group role
// acts on a Group the test itself creates (its creator is its owner, G018), never on the seeded research. Pages that answer
// an error status are reached only from extra pages, so the default page's S-ERRORS check covers every page the user sees
// succeed. Core decides every read and action; these tests observe the UI and the absence of label markup. A forged write
// is refused as core maps it (the same-facts read rule, docs/foundation/AUTHORIZATION.md "Guarded label operations"): the
// fixed 403 "Not permitted" page when the subject can read the object whose capability denies (bob, an Organization
// member, writing an Organization label), the fixed 404 "Not found" page when the definition or the Group is invisible to
// the subject (alice, not in research, assigning an Organization label she cannot read; frank, Organization staff without
// a research role, creating a research label: the override is never consulted for label.group.*). Neither page names a label.

const ORIGIN_A = "https://org-a.localhost:543";
const EMULATOR = "http://emulator:4100";
const FORM = "application/x-www-form-urlencoded";
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const LINK = /https:\/\/org-a\.localhost:543\/invitations\/([0-9a-f-]{36})#([A-Za-z0-9_-]{43})/;
const AXE_TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

// Seeded rows (infra/identity/mcp-seed.sql).
const RESEARCH = "0f14a000-0000-4000-8000-0000000000a1";

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
const DAVE_NAME = "Dave Example";

// The pages and their fixed texts (lib/labels-client.ts).
const ORG_LABELS = "/organization/labels";
const groupPage = (id: string) => `/groups/${id}`;
const groupLabels = (id: string) => `/groups/${id}/labels`;
const STATUS = {
  created: "The label was created.",
  renamed: "The label was renamed.",
  visibility_changed: "The label's visibility was changed.",
  deleted: "The label and every assignment of it were deleted.",
  assigned: "The label was assigned to this Group.",
  unassigned: "The label was removed from this Group.",
} as const;
const NO_TARGETS = "Group labels cannot be applied to anything yet: no Group-owned object type exists in this foundation";
const NOT_ADMINISTERED = "Labels of this Group are managed by its owners and admins";

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
  for (let i = 0; i < 120 && !(await target.evaluate((t) => t === document.activeElement).catch(() => false)); i++) await page.keyboard.press("Tab");
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

const notice = (page: Page, message: string) => expect(page.getByTestId("labels-status")).toHaveText(message);

// --- the Group-administration override (G060), always restored to off -------------------------------------------------
const GROUP_AUTHORITY = "/organization/group-authority";
async function setOverride(owner: Page, state: "on" | "off") {
  await owner.goto(GROUP_AUTHORITY);
  const form = owner.getByTestId("group-authority-form");
  await form.getByTestId(`group-authority-${state}`).check();
  await form.getByRole("button", { name: "Save" }).click();
  await owner.waitForURL((u) => u.pathname === GROUP_AUTHORITY && u.searchParams.get("status") === (state === "on" ? "enabled" : "disabled"));
  await expect(owner.getByTestId(`group-authority-${state}`)).toBeChecked();
}

// --- label steps ---------------------------------------------------------------------------------------------------------

// The row of one label by its exact name (names in a scenario are never prefixes of each other).
function labelRow(page: Page, name: string): Locator {
  return page.getByTestId("label-row").filter({ has: page.getByTestId("label-name").getByText(name, { exact: true }) });
}

// Creates a label on a definitions page through its create form and returns its id (from the row's delete form).
async function createLabel(page: Page, path: string, name: string, visibility?: "Group administrators" | "Group members"): Promise<string> {
  await page.goto(path);
  const form = page.getByTestId("label-create-form");
  await form.getByLabel(/^Label name/).fill(name);
  if (visibility !== undefined) await form.getByRole("radio", { name: visibility }).check();
  await form.getByRole("button", { name: "Create label" }).click();
  await page.waitForURL((u) => u.pathname === path && u.searchParams.get("status") === "created");
  await notice(page, STATUS.created);
  const id = await labelRow(page, name).getByTestId("label-delete-form").locator('input[name="label_id"]').inputValue();
  expect(id).toMatch(UUID);
  return id;
}

// Deletes the named label if the page still lists it (cleanup after a failure as well as the normal path).
async function deleteIfPresent(page: Page, path: string, name: string) {
  await page.goto(path);
  const row = labelRow(page, name);
  if ((await row.count()) === 0) return;
  await row.getByRole("button", { name: `Delete ${name}` }).click();
  await page.waitForURL((u) => u.pathname === path && u.searchParams.get("status") === "deleted");
}

// The Group page of a reader who must see no Labels section and no label markup at all (Spec §4: not serialized and
// hidden with CSS): the rendered page and its raw server answer both lack the names and the section's markers.
async function noLabelMarkup(page: Page, groupId: string, names: string[], who: string) {
  await page.goto(groupPage(groupId));
  await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
  await expect(page.getByTestId("group-labels"), who).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "Labels" }), who).toHaveCount(0);
  const html = await page.content();
  const raw = await fromPage(page.context(), groupPage(groupId));
  expect(raw.status, who).toBe(200);
  for (const text of [html, raw.text]) {
    for (const name of names) expect(text, `${who}: ${name}`).not.toContain(name);
    expect(text, who).not.toContain("group-label-assign-form");
    expect(text, who).not.toContain('data-testid="group-label"');
  }
}

// --- Group membership steps (a Group the test creates; G018/G019 flows) ------------------------------------------------

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

// The Group owner invites a co-member as Member; the recipient accepts through the one-time link.
async function addMember(owner: Page, member: Page, groupId: string, name: string) {
  await owner.goto(`/groups/${groupId}/members`);
  const form = owner.getByTestId("invitation-form");
  await chooseOption(form.getByLabel("Workspace member"), name);
  await chooseOption(form.getByLabel("Role"), "Member");
  await form.getByRole("button", { name: "Create invitation" }).click();
  await expect(owner.getByRole("heading", { level: 1 })).toHaveText("Invitation created");
  const m = LINK.exec(await owner.locator("main").innerText());
  expect(m, "the one-time invitation link").not.toBeNull();
  const [link, , token] = m as RegExpExecArray;
  observe(token);
  await member.goto(link);
  await expect(member.getByLabel("Invitation code")).not.toHaveValue("");
  await member.getByRole("button", { name: "Accept invitation" }).click();
  await member.waitForURL((u) => u.pathname === `/groups/${groupId}` && u.searchParams.get("status") === "joined");
}

// --- E-LAB-1..6 ---------------------------------------------------------------------------------------------------------

test("[E-LAB-1] the Organization owner creates, renames and deletes an Organization label by keyboard; an admin sees it, a member gets the fixed 404", async ({ page, browser }, testInfo) => {
  const sfx = suffix(testInfo);
  const first = `Alpha ${sfx}`;
  const renamed = `Bravo ${sfx}`;
  await signIn(page, CAROL);
  try {
    // The owner reaches the page from her account and from the members page.
    await page.getByTestId("account-labels-link").click();
    await page.waitForURL((u) => u.pathname === ORG_LABELS);
    await expect(page.getByRole("heading", { level: 1 })).toHaveText("Workspace labels");
    await expect(page.getByTestId("label-visibility-fixed")).toContainText("Organization administrators");
    await page.goto("/organization/members");
    await page.getByTestId("members-labels-link").click();
    await page.waitForURL((u) => u.pathname === ORG_LABELS);

    // Keyboard only: create, rename, delete.
    const name = page.getByTestId("label-create-form").getByLabel(/^Label name/);
    await tabTo(page, name, "Label name");
    await page.keyboard.type(first);
    await page.keyboard.press("Enter");
    await page.waitForURL((u) => u.searchParams.get("status") === "created");
    await notice(page, STATUS.created);
    await expect(labelRow(page, first).getByTestId("label-visibility")).toHaveText("Organization administrators");
    const rename = page.getByLabel(`New name for ${first}`);
    await tabTo(page, rename, "New name");
    await page.keyboard.press("ControlOrMeta+A");
    await page.keyboard.type(renamed);
    await page.keyboard.press("Enter");
    await page.waitForURL((u) => u.searchParams.get("status") === "renamed");
    await notice(page, STATUS.renamed);
    await expect(labelRow(page, renamed)).toHaveCount(1);
    await expect(labelRow(page, first)).toHaveCount(0);

    // frank, an Organization admin: the same definitions.
    const frank = await open(browser, testInfo);
    await signIn(frank.page, FRANK);
    await expect(frank.page.getByTestId("account-labels-link")).toBeVisible();
    await frank.page.goto(ORG_LABELS);
    await expect(labelRow(frank.page, renamed)).toHaveCount(1);

    // alice, a plain member: no link and the fixed 404 without the name.
    const alice = await open(browser, testInfo);
    await signIn(alice.page, ALICE);
    await expect(alice.page.getByTestId("account-labels-link")).toHaveCount(0);
    await alice.page.goto("/organization/members");
    await expect(alice.page.getByTestId("members-labels-link")).toHaveCount(0);
    const read = await fromPage(alice.context, ORG_LABELS);
    expect(read.status).toBe(404);
    expect(read.text).not.toContain(renamed);
    expect(read.text).not.toContain("label-list");

    const del = labelRow(page, renamed).getByRole("button", { name: `Delete ${renamed}` });
    await tabTo(page, del, "Delete");
    await page.keyboard.press("Enter");
    await page.waitForURL((u) => u.searchParams.get("status") === "deleted");
    await notice(page, STATUS.deleted);
    await expect(labelRow(page, renamed)).toHaveCount(0);
  } finally {
    await deleteIfPresent(page, ORG_LABELS, first);
    await deleteIfPresent(page, ORG_LABELS, renamed);
  }
});

test("[E-LAB-2] the Organization owner assigns and removes an Organization label on a Group page; the Group's owner and member see no Labels section and no label markup", async ({ page, browser }, testInfo) => {
  const name = `Tier ${suffix(testInfo)}`;
  await signIn(page, CAROL);
  try {
    await createLabel(page, ORG_LABELS, name);
    await page.goto(groupPage(RESEARCH));
    const section = page.getByTestId("group-labels");
    await expect(section.getByRole("heading", { level: 2 })).toHaveText("Labels");
    await chooseOption(section.getByTestId("group-label-assign-form").getByLabel("Label to assign"), name);
    await section.getByRole("button", { name: "Assign label" }).click();
    await page.waitForURL((u) => u.pathname === groupPage(RESEARCH) && u.searchParams.get("status") === "assigned");
    await notice(page, STATUS.assigned);
    await expect(page.getByTestId("group-label").filter({ hasText: name })).toHaveCount(1);

    // bob (research owner, a plain Organization member) and dave (research member): no section, no markup, no name.
    const bob = await open(browser, testInfo);
    await signIn(bob.page, BOB);
    await noLabelMarkup(bob.page, RESEARCH, [name], "bob");
    const dave = await open(browser, testInfo);
    await signIn(dave.page, DAVE);
    await noLabelMarkup(dave.page, RESEARCH, [name], "dave");

    await page.goto(groupPage(RESEARCH));
    await page.getByRole("button", { name: `Remove the label ${name}` }).click();
    await page.waitForURL((u) => u.searchParams.get("status") === "unassigned");
    await notice(page, STATUS.unassigned);
    await expect(page.getByTestId("group-label").filter({ hasText: name })).toHaveCount(0);
  } finally {
    // Deleting the definition removes any assignment left by a failure above.
    await deleteIfPresent(page, ORG_LABELS, name);
  }
});

test("[E-LAB-3] a Group owner manages the Group's own labels; a member sees only shared ones; Organization staff (override on or off) see only the notice; a non-member gets the fixed 404", async ({ page, browser }, testInfo) => {
  const sfx = suffix(testInfo);
  const shared = `Shared ${sfx}`;
  const admins = `Admins ${sfx}`;
  await signIn(page, BOB);
  const carol = await open(browser, testInfo);
  await signIn(carol.page, CAROL);
  try {
    await page.goto(groupPage(RESEARCH));
    await page.getByTestId("group-labels-link").click();
    await page.waitForURL((u) => u.pathname === groupLabels(RESEARCH));
    await expect(page.getByTestId("labels-no-targets")).toHaveText(NO_TARGETS);
    await createLabel(page, groupLabels(RESEARCH), shared, "Group members");
    await createLabel(page, groupLabels(RESEARCH), admins, "Group administrators");
    await expect(labelRow(page, shared).getByTestId("label-visibility")).toHaveText("Group members");
    await expect(labelRow(page, admins).getByTestId("label-visibility")).toHaveText("Group administrators");

    // dave (member): the read-only list of the shared label only.
    const dave = await open(browser, testInfo);
    await signIn(dave.page, DAVE);
    await dave.page.goto(groupPage(RESEARCH));
    await dave.page.getByTestId("group-labels-link").click();
    await dave.page.waitForURL((u) => u.pathname === groupLabels(RESEARCH));
    await expect(dave.page.getByTestId("labels-read-only").getByTestId("label-row")).toHaveText([shared]);
    await expect(dave.page.getByTestId("label-create-form")).toHaveCount(0);
    await expect(dave.page.getByTestId("labels-no-targets")).toHaveCount(0);
    expect(await dave.page.content()).not.toContain(admins);

    // frank (Organization admin) with the override on: the notice, never the definitions; no link on the Group page.
    await setOverride(carol.page, "on");
    const frank = await open(browser, testInfo);
    await signIn(frank.page, FRANK);
    await frank.page.goto(groupPage(RESEARCH));
    await expect(frank.page.getByTestId("group-labels-link")).toHaveCount(0);
    await frank.page.goto(groupLabels(RESEARCH));
    await expect(frank.page.getByTestId("labels-not-administered")).toHaveText(NOT_ADMINISTERED);
    await expect(frank.page.getByTestId("label-list")).toHaveCount(0);
    for (const n of [shared, admins]) expect(await frank.page.content()).not.toContain(n);

    // carol (Organization owner) with the override off: the same notice.
    await setOverride(carol.page, "off");
    await carol.page.goto(groupLabels(RESEARCH));
    await expect(carol.page.getByTestId("labels-not-administered")).toHaveText(NOT_ADMINISTERED);
    for (const n of [shared, admins]) expect(await carol.page.content()).not.toContain(n);

    // alice (not in research): the fixed 404 without any name.
    const alice = await open(browser, testInfo);
    await signIn(alice.page, ALICE);
    const read = await fromPage(alice.context, groupLabels(RESEARCH));
    expect(read.status).toBe(404);
    for (const n of [shared, admins]) expect(read.text).not.toContain(n);
  } finally {
    await setOverride(carol.page, "off");
    await deleteIfPresent(page, groupLabels(RESEARCH), shared);
    await deleteIfPresent(page, groupLabels(RESEARCH), admins);
  }
});

test("[E-LAB-4] narrowing a label's visibility empties the member's next load; removing the member's Group role turns the page into the fixed 404", async ({ page, browser }, testInfo) => {
  const sfx = suffix(testInfo);
  const name = `Visible ${sfx}`;
  // A Group of bob's with dave as a member, so the seeded research keeps its members for every later scenario.
  await signIn(page, BOB);
  const groupId = await createOwnGroup(page, `Labels ${sfx}`);
  const dave = await open(browser, testInfo);
  await signIn(dave.page, DAVE);
  await addMember(page, dave.page, groupId, DAVE_NAME);
  const carol = await open(browser, testInfo);
  await signIn(carol.page, CAROL);
  try {
    await createLabel(page, groupLabels(groupId), name, "Group members");
    await dave.page.goto(groupLabels(groupId));
    await expect(dave.page.getByTestId("labels-read-only").getByTestId("label-row")).toHaveText([name]);

    // bob narrows it to Group administrators: dave's next load lists nothing.
    await page.goto(groupLabels(groupId));
    const form = labelRow(page, name).getByTestId("label-visibility-form");
    // G068/G069: the kit radio treatment, label alignment, form value and focus state of the visibility choice. The kit
    // FieldSet is found by its legend's name, since every kit Field also renders an (unnamed) group.
    await expectRadioOptions(page, testInfo, form.getByRole("group", { name: `Who can see ${name}`, exact: true }), "Label visibility", "visibility");
    await form.getByRole("radio", { name: "Group administrators" }).check();
    await form.getByRole("button", { name: "Change visibility" }).click();
    await page.waitForURL((u) => u.searchParams.get("status") === "visibility_changed");
    await notice(page, STATUS.visibility_changed);
    await dave.page.reload();
    await expect(dave.page.getByTestId("labels-read-only")).toBeVisible();
    await expect(dave.page.getByTestId("labels-read-only").getByTestId("label-row")).toHaveCount(0);
    expect(await dave.page.content()).not.toContain(name);

    // carol removes dave's Group role (through the Group-administration override): his next request is the fixed 404.
    await setOverride(carol.page, "on");
    await carol.page.goto(`/groups/${groupId}/members`);
    await carol.page.getByRole("button", { name: `Remove ${DAVE_NAME}` }).click();
    await carol.page.waitForURL((u) => u.searchParams.get("status") === "removed");
    await setOverride(carol.page, "off");
    const read = await fromPage(dave.context, groupLabels(groupId));
    expect(read.status).toBe(404);
    expect(read.text).not.toContain(name);
  } finally {
    await setOverride(carol.page, "off");
    await deleteIfPresent(page, groupLabels(groupId), name);
  }
});

test("[E-LAB-5] forged label writes are refused (Not permitted, or the fixed 404 for an invisible label or Group) and change nothing", async ({ page, browser }, testInfo) => {
  const sfx = suffix(testInfo);
  const orgLabel = `Forge target ${sfx}`;
  const forgedOrg = `Forged org ${sfx}`;
  const forgedGroup = `Forged group ${sfx}`;
  await signIn(page, CAROL);
  try {
    const labelId = await createLabel(page, ORG_LABELS, orgLabel);

    // alice (plain member, not in research) assigns the Organization label to research with her own valid CSRF token:
    // the definition is invisible to her (row-level security admits Organization-owned definitions to Organization
    // owners/admins only), so core answers not_found before any decision and the route renders the fixed 404 page.
    const alice = await open(browser, testInfo);
    await signIn(alice.page, ALICE);
    const aliceCsrf = await csrfOf(alice.page);
    const assign = await fromPage(alice.context, `${groupLabels(RESEARCH)}/action`, new URLSearchParams({ csrf_token: aliceCsrf, action: "assign", label_id: labelId }).toString());
    expect(assign.status).toBe(404);
    expect(assign.text).toContain("Not found");
    expect(assign.text).toContain("Nothing was changed.");
    expect(assign.text).not.toContain(orgLabel);

    // bob (research owner, plain Organization member) creates an Organization label.
    const bob = await open(browser, testInfo);
    await signIn(bob.page, BOB);
    const bobCsrf = await csrfOf(bob.page);
    const create = await fromPage(bob.context, `${ORG_LABELS}/action`, new URLSearchParams({ csrf_token: bobCsrf, action: "create", name: forgedOrg }).toString());
    expect(create.status).toBe(403);
    expect(create.text).toContain("Not permitted");

    // frank (Organization admin) with the override on creates a research label: label.group.manage is denied and its
    // read capability label.group.read (an actual research role) too, so core answers not_found and the route renders
    // the fixed 404 page (the same answer as E-LAB-3's page for Organization staff without a Group role).
    await setOverride(page, "on");
    const frank = await open(browser, testInfo);
    await signIn(frank.page, FRANK);
    const frankCsrf = await csrfOf(frank.page);
    const group = await fromPage(
      frank.context,
      `${groupLabels(RESEARCH)}/action`,
      new URLSearchParams({ csrf_token: frankCsrf, action: "create", name: forgedGroup, visibility: "owner_scope" }).toString(),
    );
    expect(group.status).toBe(404);
    expect(group.text).toContain("Not found");
    expect(group.text).toContain("Nothing was changed.");
    expect(group.text).not.toContain(forgedGroup);
    await setOverride(page, "off");

    // Nothing changed: no assignment, no Organization label, no research label.
    await page.goto(groupPage(RESEARCH));
    await expect(page.getByTestId("group-label").filter({ hasText: orgLabel })).toHaveCount(0);
    await page.goto(ORG_LABELS);
    await expect(labelRow(page, forgedOrg)).toHaveCount(0);
    await bob.page.goto(groupLabels(RESEARCH));
    await expect(labelRow(bob.page, forgedGroup)).toHaveCount(0);
  } finally {
    await setOverride(page, "off");
    await deleteIfPresent(page, ORG_LABELS, orgLabel);
  }
});

test("[E-LAB-6] axe WCAG 2.2 AA at 1280 and 320 on the three label pages", async ({ page, browser }, testInfo) => {
  const sfx = suffix(testInfo);
  const orgLabel = `Audit org ${sfx}`;
  const groupLabel = `Audit group ${sfx}`;
  await signIn(page, CAROL);
  const bob = await open(browser, testInfo);
  await signIn(bob.page, BOB);
  try {
    await createLabel(page, ORG_LABELS, orgLabel);
    await audit(page, "Organization labels");
    await page.goto(groupPage(RESEARCH));
    await chooseOption(page.getByTestId("group-label-assign-form").getByLabel("Label to assign"), orgLabel);
    await page.getByRole("button", { name: "Assign label" }).click();
    await page.waitForURL((u) => u.searchParams.get("status") === "assigned");
    await audit(page, "Group page with its Labels section");
    await createLabel(bob.page, groupLabels(RESEARCH), groupLabel, "Group members");
    await audit(bob.page, "Group labels (manager)");
    const dave = await open(browser, testInfo);
    await signIn(dave.page, DAVE);
    await dave.page.goto(groupLabels(RESEARCH));
    await audit(dave.page, "Group labels (member)");
    await page.goto(groupLabels(RESEARCH));
    await audit(page, "Group labels (not administered)");
  } finally {
    await deleteIfPresent(page, ORG_LABELS, orgLabel);
    await deleteIfPresent(bob.page, groupLabels(RESEARCH), groupLabel);
  }
});

// --- landmarks and navigation of the Organization labels page (S-LANDMARKS and S-NAV with the Members area) -------------

test("[S-LANDMARKS] Organization labels: one banner, Primary and Workspace areas navigation, main and contentinfo; heading order", async ({ page }, testInfo) => {
  await signIn(page, CAROL);
  await page.goto(ORG_LABELS);
  await expect(page.getByRole("banner")).toHaveCount(1);
  await expect(page.getByRole("navigation", { name: "Primary" })).toHaveCount(1);
  await expect(page.getByRole("navigation", { name: "Workspace areas" })).toHaveCount(1);
  await expect(page.getByRole("main")).toHaveCount(1);
  await expect(page.getByRole("contentinfo")).toHaveCount(1);
  const areas = page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link");
  await expect(areas).toHaveText(["Groups", "Business Profiles", "Members"]);
  const current = await areas.evaluateAll((as) => as.filter((a) => a.getAttribute("aria-current") === "page").map((a) => a.getAttribute("href")));
  expect(current).toEqual(["/organization/members"]);
  const levels = await page.locator("h1, h2, h3, h4, h5, h6").evaluateAll((hs) => hs.map((h) => Number(h.tagName[1])));
  expect(levels.filter((l) => l === 1)).toHaveLength(1);
  expect(levels[0]).toBe(1);
  levels.forEach((level, i) => {
    if (i > 0) expect(level - levels[i - 1], `heading order ${levels.join(",")}`).toBeLessThanOrEqual(1);
  });
  // G067 (AC2, Plan §7.1 row 6): visual evidence of the successor styling on an app surface, full page at the
  // project's 1280px viewport and at 320px, in the project's appearance. Evidence only; no assertion depends on it.
  await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
  await testInfo.attach(`labels-org-1280-${testInfo.project.name}.png`, { body: await page.screenshot({ fullPage: true, animations: "disabled" }), contentType: "image/png" });
  await page.setViewportSize({ width: 320, height: 640 });
  await testInfo.attach(`labels-org-320-${testInfo.project.name}.png`, { body: await page.screenshot({ fullPage: true, animations: "disabled" }), contentType: "image/png" });
});

test("[S-NAV] Organization labels: Primary links reach their sections by keyboard", async ({ page }) => {
  await signIn(page, CAROL);
  await page.goto(ORG_LABELS);
  const links = page.getByRole("navigation", { name: "Primary" }).getByRole("link");
  const count = await links.count();
  expect(count).toBeGreaterThan(0);
  for (let i = 0; i < count; i++) {
    const href = await links.nth(i).getAttribute("href");
    expect(href).toMatch(/^#[a-z-]+$/);
    await links.nth(i).focus();
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(new RegExp(`${href}$`));
    await expect(page.locator(`${href} h1, ${href} h2`).first(), href ?? "").toBeInViewport();
  }
});
