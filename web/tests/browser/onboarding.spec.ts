import { randomBytes } from "node:crypto";
import { readFileSync } from "node:fs";

import AxeBuilder from "@axe-core/playwright";
import { type Browser, type BrowserContext, type Locator, type Page, type TestInfo } from "@playwright/test";

import { expectAxeClean } from "./support/contrast";
import { expectRadioOptions } from "./support/form-controls";
import { expect, test } from "./support/page-errors";
import { observe } from "./support/sanitize";

// G018 Group creation and create-or-select Business Profile onboarding (Plan r5 §1.6, §6.2 E-ONB-1..7) in the
// disposable HTTPS identity environment (infra/compose/identity.yml), projects onboarding-light and onboarding-dark. The
// harness seeds the fixture first (infra/identity/mcp-seed.sql): frank is Organization A's admin, alice a plain member
// of A (owner D-1, superseding: every Organization role creates its own Group), bob the owner of research (A, onboarded
// with "Research Labs"), dave a member of research and, in Organization C, a Group-role-only subject without any
// Organization binding there (his tenant sign-in on C is refused, G016 D-2(a)). design (A) is onboarded with "Design
// Studio"; Organization B holds "Operations Partner". Both projects run against the same database one after the other,
// so every Group a test creates carries a per-run suffix. Pages that answer an error status are reached only through
// same-origin fetches from extra pages, so the default page's S-ERRORS check covers every page the user sees. Every
// member tenant page carries the Workspace areas navigation (G018 decision (b)); the landmark and navigation checks
// below include it. G060 (r1 C6, correction02 C02-4) adds E-ONB-8: carol (Organization A's owner) and frank read an
// incomplete Group of alice's without onboarding it unless carol turns the Group-administration override on; the test
// always ends with the override off.

const ORIGIN_A = "https://org-a.localhost:543";
const ORIGIN_C = "https://org-c.localhost:543";
const EMULATOR = "http://emulator:4100";
const FORM = "application/x-www-form-urlencoded";
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const SLUG = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
const AXE_TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

// Seeded rows (infra/identity/mcp-seed.sql).
const RESEARCH = "0f14a000-0000-4000-8000-0000000000a1";
const DESIGN = "0f14a000-0000-4000-8000-0000000000a2";
const RESEARCH_LABS = "0f18b000-0000-4000-8000-0000000000a1";
const DESIGN_STUDIO = "0f18b000-0000-4000-8000-0000000000a2";
const ABSENT = "7f000000-0000-4000-8000-0000000000a9";

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

// A per-test suffix: the project's appearance letter and random hex (slug grammar: lowercase, digits, hyphens).
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
  const outside = [...tracking.origins].filter((o) => ![ORIGIN_A, ORIGIN_C, EMULATOR].includes(o));
  expect(outside, "page requests outside the app origins under test and the emulator").toEqual([]);
});

async function open(browser: Browser, testInfo: TestInfo): Promise<{ context: BrowserContext; page: Page }> {
  const context = await browser.newContext({ baseURL: ORIGIN_A, colorScheme: testInfo.project.use.colorScheme, viewport: testInfo.project.use.viewport });
  track(context);
  tracking.contexts.push(context);
  return { context, page: await context.newPage() };
}

// --- sign-in through the emulator's hosted page ------------------------------------------------------------------------
// from "/" starts at the tenant entry's own Sign in (G018 M1 D1: the guarded login form, no intermediate page).
async function credentials(page: Page, who: Credentials, origin: string, from: "/sign-in" | "/" = "/sign-in") {
  if (from === "/") {
    await page.goto(`${origin}/`);
    await page.getByTestId("entry-sign-in").click();
  } else {
    await page.goto(`${origin}/sign-in?return_to=/account`);
    await page.getByRole("button", { name: "Sign in" }).click();
  }
  await page.waitForURL((u) => u.origin === EMULATOR);
  await page.locator("#email").fill(who.email);
  await page.getByRole("button", { name: "Continue" }).click();
  const password = page.locator('input[name="password"]');
  await password.fill(who.password);
  await password.press("Enter");
}

async function signIn(page: Page, who: Credentials, origin = ORIGIN_A) {
  await credentials(page, who, origin);
  await page.waitForURL((u) => u.origin === origin && u.pathname === "/account");
}

// A sign-in that the membership rule refuses (G016 D-2(a)): back on ORIGIN's sign-in page with the not-member notice.
async function signInRefused(page: Page, who: Credentials, origin: string) {
  await credentials(page, who, origin);
  await page.waitForURL((u) => u.origin === origin && u.pathname === "/sign-in" && u.searchParams.get("status") === "not_member");
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
type Answer = { status: number; redirected: boolean; location: string; text: string };
async function fromPage(context: BrowserContext, origin: string, path: string, body?: string): Promise<Answer> {
  const p = await context.newPage();
  await p.goto(`${origin}/api/health`);
  const r = await p.evaluate(
    async ({ path, body, type }) => {
      const init: RequestInit = body === undefined ? { redirect: "manual", cache: "no-store" } : { method: "POST", body, headers: { "content-type": type }, redirect: "manual", cache: "no-store" };
      const res = await fetch(path, init);
      const redirected = res.type === "opaqueredirect";
      return { status: res.status, redirected, location: res.headers.get("location") ?? "", text: redirected ? "" : await res.text() };
    },
    { path, body, type: FORM },
  );
  await p.close();
  return r;
}

// --- onboarding steps ---------------------------------------------------------------------------------------------------
const nameField = (page: Page) => page.getByRole("textbox", { name: "Group name" });
const slugField = (page: Page) => page.getByRole("textbox", { name: "Group address (slug)" });
const advisory = (page: Page) => page.getByTestId("address-advisory");
// The advisory copy (lib/onboarding-client.ts ADDRESS_ADVISORY) and the inline conflict (G018 M1 D2, D3).
const IN_USE = "A Group you can see already uses this address. Choose another one.";
const NOT_SEEN = "No Group you can see uses this address. The final check happens when you create the Group.";
const NOTHING_CREATED = "Nothing was created by this request.";

// Fills the slug field and leaves it with the keyboard (the advisory check runs on blur).
async function blurSlug(page: Page, value: string) {
  await slugField(page).fill(value);
  await page.keyboard.press("Tab");
}

// Step 1 and step 2 of New Group, from the Groups page; returns the suggested address shown in step 2.
async function startGroup(page: Page, name: string, origin = ORIGIN_A): Promise<string> {
  await page.goto(`${origin}/groups`);
  await page.getByRole("link", { name: "New Group", exact: true }).click();
  await page.waitForURL((u) => u.pathname === "/groups/new");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Name your Group");
  await nameField(page).fill(name);
  await page.getByRole("button", { name: "Continue" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Choose its address");
  await expect(nameField(page)).toHaveValue(name);
  const suggested = await slugField(page).inputValue();
  expect(suggested).toMatch(SLUG);
  return suggested;
}

// Submits step 2 (optionally with an edited address) and returns the new Group's id from its onboarding page.
async function createGroup(page: Page, slug?: string): Promise<string> {
  if (slug !== undefined) await slugField(page).fill(slug);
  await page.getByRole("button", { name: "Create Group" }).click();
  await page.waitForURL((u) => /^\/groups\/[0-9a-f-]{36}\/onboarding$/.test(u.pathname));
  const id = new URL(page.url()).pathname.split("/")[2];
  expect(id).toMatch(UUID);
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(/^Onboard /);
  return id;
}

const newProfileForm = (page: Page) => page.getByTestId("profile-new-form");

// The create-new branch on the onboarding page; ends on the Group page with the completed status.
async function completeWithNewProfile(page: Page, groupId: string, displayName: string) {
  const form = newProfileForm(page);
  await form.getByRole("textbox", { name: "Display name" }).fill(displayName);
  await form.getByRole("textbox", { name: "Industry" }).fill("Manufacturing");
  await form.getByRole("textbox", { name: "Website" }).fill("https://example.com");
  await form.getByRole("button", { name: "Create and attach" }).click();
  await page.waitForURL((u) => u.pathname === `/groups/${groupId}` && u.searchParams.get("status") === "completed");
  await expect(page.getByText("Onboarding is complete: the Business Profile is attached to this Group.")).toBeVisible();
  await expect(page.getByTestId("group-state")).toHaveText("Complete");
  await expect(page.getByTestId("group-profile").getByRole("heading", { level: 2 })).toHaveText(displayName);
}

// The Group headings of the current Groups page.
async function groupNames(page: Page): Promise<string[]> {
  const list = page.getByTestId("group-list");
  if ((await list.count()) === 0) return [];
  return list.getByRole("heading", { level: 2 }).allTextContents();
}

// Presses Tab until TARGET has the focus (keyboard-only navigation).
async function tabTo(page: Page, target: Locator, label: string) {
  for (let i = 0; i < 60 && !(await target.evaluate((t) => t === document.activeElement).catch(() => false)); i++) await page.keyboard.press("Tab");
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

// --- E-ONB-1..7 --------------------------------------------------------------------------------------------------------

test("[E-ONB-1] an admin creates a Group with an edited suggested address and onboards it with a new Business Profile", async ({ page }, testInfo) => {
  const sfx = suffix(testInfo);
  // G018 M1 (D1): sign-in straight from the tenant entry's Sign in (the guarded login form), back on the entry.
  await credentials(page, FRANK, ORIGIN_A, "/");
  await page.waitForURL((u) => u.origin === ORIGIN_A && u.pathname === "/");
  await expect(page.getByText("You are signed in.")).toBeVisible();
  await page.goto("/groups");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Your Groups");
  // An Organization admin reads every Group of the Organization.
  expect(await groupNames(page)).toEqual(expect.arrayContaining(["Research", "Design"]));
  const suggested = await startGroup(page, `Launch ${sfx}`);
  expect(suggested).toContain(sfx);
  const slug = `e1-${sfx}-edited`;
  // G018 M1 (D2): the advisory hint on blur. A controlled stale answer: the check for "research" is held, the person
  // edits and leaves the field with another value, and the held in_use answer arriving afterwards is never shown.
  const fieldHints = slugField(page);
  expect(((await fieldHints.getAttribute("aria-describedby")) ?? "").split(" ")).toContain("slug-advisory");
  let release = () => {};
  const held = new Promise<void>((resolve) => (release = resolve));
  let heldSeen = false;
  await page.route(/\/groups\/address-check\?slug=research$/, async (route) => {
    heldSeen = true;
    await held;
    await route.continue().catch(() => undefined); // the browser has aborted it by now
  });
  await blurSlug(page, "research");
  await expect.poll(() => heldSeen).toBe(true);
  await blurSlug(page, slug);
  await expect(advisory(page)).toHaveText(NOT_SEEN);
  release();
  await page.unroute(/\/groups\/address-check\?slug=research$/);
  await page.waitForTimeout(500);
  await expect(advisory(page)).toHaveText(NOT_SEEN);
  // A readable holder: in_use, advisory only (the field is not marked invalid and the create stays available).
  await blurSlug(page, "research");
  await expect(advisory(page)).toHaveText(IN_USE);
  await expect(slugField(page)).not.toHaveAttribute("aria-invalid", "true");
  await expect(page.getByRole("button", { name: "Create Group" })).toBeEnabled();
  // Editing clears the hint at once.
  await slugField(page).press("End");
  await page.keyboard.type("x");
  await expect(advisory(page)).toHaveText("");
  const id = await createGroup(page, slug);
  // A refused field is associated with its input (no submitted value echoed); nothing is attached.
  const form = newProfileForm(page);
  await form.getByRole("textbox", { name: "Display name" }).fill(`Launch Co ${sfx}`);
  // A URL the browser accepts but core refuses (only https:// and http:// web addresses are valid).
  await form.getByRole("textbox", { name: "Website" }).fill("ftp://example.com");
  await form.getByRole("button", { name: "Create and attach" }).click();
  await page.waitForURL((u) => u.pathname === `/groups/${id}/onboarding` && u.searchParams.get("error") === "invalid_website");
  const website = newProfileForm(page).getByRole("textbox", { name: "Website" });
  await expect(website).toHaveAttribute("aria-invalid", "true");
  expect(((await website.getAttribute("aria-describedby")) ?? "").split(" ")).toContain("website-error");
  await expect(page.locator("#website-error")).toContainText("https://");
  await expect(website).toBeFocused();
  await expect(website).toHaveValue("");
  await expect(page.getByText("A value was not in the accepted form, so nothing was changed.")).toBeVisible();
  await completeWithNewProfile(page, id, `Launch Co ${sfx}`);
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(`Launch ${sfx}`);
  await expect(page.getByText(slug, { exact: true })).toBeVisible();
  await expect(page.getByRole("link", { name: "Edit Business Profile" })).toBeVisible();
  await page.goto("/groups");
  await expect(page.getByRole("link", { name: `Open Launch ${sfx}` })).toHaveAttribute("href", `/groups/${id}`);
});

test("[E-ONB-2] a Group owner resumes an incomplete Group and selects an existing Business Profile", async ({ page }, testInfo) => {
  const sfx = suffix(testInfo);
  await signIn(page, BOB);
  await startGroup(page, `Resume ${sfx}`);
  const id = await createGroup(page);
  // Leaving the onboarding page leaves the Group incomplete: no false completion anywhere.
  await page.goto(`/groups/${id}`);
  await expect(page.getByTestId("group-state")).toHaveText("Incomplete");
  await expect(page.getByTestId("group-profile")).toHaveCount(0);
  // A forged completion marker in the URL never presents the incomplete Group as complete: core's state decides.
  for (const status of ["completed", "already_complete"]) {
    await page.goto(`/groups/${id}?status=${status}`);
    await expect(page.getByTestId("group-state")).toHaveText("Incomplete");
    await expect(page.getByRole("link", { name: "Resume onboarding" })).toBeVisible();
    await expect(page.getByTestId("status-message")).toHaveCount(0);
    await expect(page.getByText(/Onboarding is complete|already has a Business Profile/)).toHaveCount(0);
    await expect(page.getByTestId("group-profile")).toHaveCount(0);
  }
  await page.goto("/groups");
  await page.getByRole("link", { name: `Resume onboarding for Resume ${sfx}` }).click();
  await page.waitForURL((u) => u.pathname === `/groups/${id}/onboarding`);
  // Only profiles bob may select: the one attached to research, which he owns; not design's, not Organization B's.
  const existing = page.getByTestId("profile-existing-form");
  await expect(existing.getByRole("radio", { name: /Research Labs/ })).toHaveCount(1);
  await expect(existing.getByRole("radio", { name: /Design Studio|Operations Partner/ })).toHaveCount(0);
  await existing.getByRole("radio", { name: /Research Labs/ }).check();
  // G068/G069: the kit radio treatment, label alignment, form value and focus state of the profile picker.
  await expectRadioOptions(page, testInfo, existing.getByRole("group", { name: "My Business Profiles" }), "My Business Profiles", "business_profile_id");
  await existing.getByRole("button", { name: "Attach selected" }).click();
  await page.waitForURL((u) => u.pathname === `/groups/${id}` && u.searchParams.get("status") === "completed");
  await expect(page.getByTestId("group-state")).toHaveText("Complete");
  await expect(page.getByTestId("status-message")).toHaveText("Onboarding is complete: the Business Profile is attached to this Group.");
  await expect(page.getByTestId("group-profile").getByRole("heading", { level: 2 })).toHaveText("Research Labs");
  await expect(page.getByRole("link", { name: "Edit Business Profile" })).toHaveAttribute("href", `/business-profiles/${RESEARCH_LABS}`);
  // A completed Group's onboarding page shows its state and no forms.
  await page.goto(`/groups/${id}/onboarding`);
  await expect(page.getByText("This Group's onboarding is already complete: it has a Business Profile.")).toBeVisible();
  await expect(page.locator("form")).toHaveCount(0);
});

test("[E-ONB-3] both entry points reach the same shared Business Profile; a member sees it read-only", async ({ page, browser }, testInfo) => {
  await signIn(page, BOB);
  // From the Group page.
  await page.goto(`/groups/${RESEARCH}`);
  await page.getByTestId("group-profile").getByRole("link", { name: "Edit Business Profile" }).click();
  await page.waitForURL((u) => u.pathname === `/business-profiles/${RESEARCH_LABS}`);
  const viaGroup = { heading: await page.getByRole("heading", { level: 1 }).textContent(), name: await page.getByRole("textbox", { name: "Display name" }).inputValue() };
  await expect(page.getByTestId("profile-edit-form")).toBeVisible();
  await expect(page.getByTestId("last-save-wins")).toBeVisible();
  // From the Business Profiles index: the same record. G060 (r1 C2): the index lists what bob may read — the profile of
  // the Group he owns, editable — and no other Group's or Organization's profile.
  await page.goto("/business-profiles");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Business Profiles");
  await expect(page.getByTestId("profile-list").getByRole("listitem").filter({ hasText: "Research Labs" }).getByTestId("profile-read-only")).toHaveCount(0);
  await expect(page.getByRole("link", { name: /^Open (Design Studio|Operations Partner)$/ })).toHaveCount(0);
  await page.getByRole("link", { name: "Open Research Labs" }).click();
  await page.waitForURL((u) => u.pathname === `/business-profiles/${RESEARCH_LABS}`);
  expect({ heading: await page.getByRole("heading", { level: 1 }).textContent(), name: await page.getByRole("textbox", { name: "Display name" }).inputValue() }).toEqual(viaGroup);
  expect(viaGroup.name).toBe("Research Labs");

  // dave, a member of research: the Group page offers View, the profile page is read-only, and a forged update is
  // refused with nothing changed. G060 (r1 C2, supersedes the G018 selectable-only list): his index lists the profile
  // he may read, marked read-only.
  const member = await open(browser, testInfo);
  await signIn(member.page, DAVE);
  await member.page.goto(`/groups/${RESEARCH}`);
  await member.page.getByTestId("group-profile").getByRole("link", { name: "View Business Profile" }).click();
  await member.page.waitForURL((u) => u.pathname === `/business-profiles/${RESEARCH_LABS}`);
  await expect(member.page.getByTestId("profile-view")).toContainText("Research Labs");
  await expect(member.page.getByTestId("profile-edit-form")).toHaveCount(0);
  await expect(member.page.getByRole("textbox")).toHaveCount(0);
  // dave's own CSRF token from a form page he may use (step 2 of New Group renders it; nothing is submitted there).
  await member.page.goto(`/groups/new?name=${encodeURIComponent("Token probe")}`);
  const csrf = await csrfOf(member.page);
  const body = new URLSearchParams({ csrf_token: csrf, display_name: "Changed by a member", legal_name: "", description: "", industry: "", website: "" }).toString();
  const forged = await fromPage(member.context, ORIGIN_A, `/business-profiles/${RESEARCH_LABS}/update`, body);
  expect(forged.status).toBe(403);
  expect(forged.text).toContain("Not permitted");
  await member.page.goto("/business-profiles");
  const daveRow = member.page.getByTestId("profile-list").getByRole("listitem").filter({ hasText: "Research Labs" });
  await expect(daveRow.getByRole("link", { name: "Open Research Labs" })).toHaveAttribute("href", `/business-profiles/${RESEARCH_LABS}`);
  await expect(daveRow.getByTestId("profile-read-only")).toHaveText("Read-only");
  await expect(member.page.getByRole("link", { name: /^Open (Design Studio|Operations Partner)$/ })).toHaveCount(0);
  await page.reload();
  await expect(page.getByRole("textbox", { name: "Display name" })).toHaveValue("Research Labs");
});

test("[E-ONB-4] a plain member creates and onboards its own Group and sees no other Group; CSRF and Group-role-only refusals create nothing", async ({ page, context, browser }, testInfo) => {
  const sfx = suffix(testInfo);
  // Owner D-1 (superseding): alice, a plain Organization member of A, sees New Group and onboards her own Group.
  await signIn(page, ALICE);
  await page.goto("/groups");
  await expect(page.getByRole("link", { name: "New Group", exact: true })).toBeVisible();
  const memberSlug = await startGroup(page, `Member ${sfx}`);
  // G018 M1 (D2, D3): research is not readable for alice, so its address is "not seen" (a filtered absence, never
  // availability), and core's conflict on submit is the inline refusal with nothing created and no holder named.
  await blurSlug(page, "research");
  await expect(advisory(page)).toHaveText(NOT_SEEN);
  await page.getByRole("button", { name: "Create Group" }).click();
  await page.waitForURL((u) => u.pathname === "/groups/new" && u.searchParams.get("error") === "slug_taken");
  await expect(slugField(page)).toHaveValue("research");
  await expect(slugField(page)).toHaveAttribute("aria-invalid", "true");
  await expect(page.getByTestId("field-error")).toHaveText(/already in use/);
  await expect(page.getByTestId("status-message")).toHaveText(NOTHING_CREATED);
  await expect(page.getByTestId("address-holder")).toHaveCount(0);
  await expect(page.getByText("Research", { exact: false })).toHaveCount(0);
  const id = await createGroup(page, memberSlug);
  await completeWithNewProfile(page, id, `Member Co ${sfx}`);
  await page.goto("/groups");
  const names = await groupNames(page);
  expect(names).toContain(`Member ${sfx}`);
  expect(names).not.toContain("Research");
  expect(names).not.toContain("Design");
  // Another member's Group, its profile and an absent id answer the same fixed 404 (no existence leak).
  for (const path of [`/groups/${RESEARCH}`, `/groups/${RESEARCH}/onboarding`, `/business-profiles/${RESEARCH_LABS}`, `/groups/${ABSENT}`, `/business-profiles/${ABSENT}`]) {
    const answer = await fromPage(context, ORIGIN_A, path);
    expect(answer.status, path).toBe(404);
    expect(answer.text, path).not.toContain("Research");
  }
  // A forged completion of another member's Group: refused, nothing attached.
  await page.goto(`/groups/new?name=${encodeURIComponent("Token probe")}`);
  const csrf = await csrfOf(page);
  const stolen = await fromPage(context, ORIGIN_A, `/groups/${DESIGN}/onboarding/complete`, new URLSearchParams({ csrf_token: csrf, mode: "existing", business_profile_id: DESIGN_STUDIO }).toString());
  expect([403, 404]).toContain(stolen.status);
  expect(stolen.text).toContain("Nothing was changed");
  // A create whose CSRF token is not the session's: the fixed refusal page, and no Group exists afterwards.
  const forged = await fromPage(context, ORIGIN_A, "/groups/create", new URLSearchParams({ csrf_token: "A".repeat(43), name: `Forged ${sfx}`, slug: `csrf-${sfx}` }).toString());
  expect(forged.status).toBe(403);
  expect(forged.text).toContain("Nothing was changed");
  await page.goto("/groups");
  expect(await groupNames(page)).not.toContain(`Forged ${sfx}`);

  // dave holds only a Group binding in Organization C (no Organization binding): his sign-in there is refused, the
  // Groups page offers no create action, and a direct create is the fixed status page with nothing created.
  const groupOnly = await open(browser, testInfo);
  await signInRefused(groupOnly.page, DAVE, ORIGIN_C);
  expect((await groupOnly.context.cookies(ORIGIN_C)).find((c) => c.name === "__Host-sid")).toBeUndefined();
  await groupOnly.page.goto(`${ORIGIN_C}/groups`);
  await expect(groupOnly.page.getByTestId("onboarding-sign-in-first")).toBeVisible();
  // G018 M1 (D4): the notice's Sign in is the guarded login form, not a link to the intermediate page.
  await expect(groupOnly.page.locator('#onboarding form[method="post"][action="/auth/login"]').getByRole("button", { name: "Sign in" })).toHaveCount(1);
  await expect(groupOnly.page.locator("#onboarding").getByRole("link", { name: "Sign in" })).toHaveCount(0);
  // The advisory check without a session for C: 401 with the fixed error only.
  const unsigned = await fromPage(groupOnly.context, ORIGIN_C, "/groups/address-check?slug=research");
  expect(unsigned.status).toBe(401);
  expect(JSON.parse(unsigned.text)).toEqual({ error: "unauthenticated" });
  await expect(groupOnly.page.getByRole("link", { name: "New Group", exact: true })).toHaveCount(0);
  await expect(groupOnly.page.getByRole("navigation", { name: "Workspace areas" })).toHaveCount(0);
  const probeSlug = `group-only-${sfx}`;
  const direct = await fromPage(groupOnly.context, ORIGIN_C, "/groups/create", new URLSearchParams({ csrf_token: "A".repeat(43), name: `Group only ${sfx}`, slug: probeSlug }).toString());
  expect(direct.status).toBe(401);
  expect(direct.text).toContain("Nothing was changed");
  // Nothing holds that address in C: a member of C creates a Group with it and is not shown the inline conflict.
  const memberC = await open(browser, testInfo);
  await signIn(memberC.page, ALICE, ORIGIN_C);
  await startGroup(memberC.page, `Group only ${sfx}`, ORIGIN_C);
  await createGroup(memberC.page, probeSlug);
});

test("[E-ONB-5] keyboard-only Group creation and onboarding; axe at 1280 and 320 on every onboarding page", async ({ page }, testInfo) => {
  const sfx = suffix(testInfo);
  await page.goto("/groups");
  await expect(page.getByTestId("onboarding-sign-in-first")).toBeVisible();
  await audit(page, "sign-in first");
  await signIn(page, FRANK);
  await page.goto("/groups");
  await audit(page, "Groups");
  await tabTo(page, page.getByRole("link", { name: "New Group", exact: true }), "New Group");
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.pathname === "/groups/new");
  await audit(page, "New Group step 1");
  await tabTo(page, nameField(page), "Group name");
  await page.keyboard.type(`Keyboard ${sfx}`);
  await page.keyboard.press("Enter");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Choose its address");
  await audit(page, "New Group step 2");
  await tabTo(page, page.getByRole("button", { name: "Create Group" }), "Create Group");
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => /^\/groups\/[0-9a-f-]{36}\/onboarding$/.test(u.pathname));
  const id = new URL(page.url()).pathname.split("/")[2];
  await audit(page, "onboarding");
  await tabTo(page, newProfileForm(page).getByRole("textbox", { name: "Display name" }), "Display name");
  await page.keyboard.type(`Keyboard Co ${sfx}`);
  await tabTo(page, newProfileForm(page).getByRole("button", { name: "Create and attach" }), "Create and attach");
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.pathname === `/groups/${id}` && u.searchParams.get("status") === "completed");
  await audit(page, "Group completed");
  await tabTo(page, page.getByRole("link", { name: "Edit Business Profile" }), "Edit Business Profile");
  await page.keyboard.press("Enter");
  await page.waitForURL((u) => u.pathname.startsWith("/business-profiles/"));
  await audit(page, "Business Profile editor");
  await page.goto("/business-profiles");
  await audit(page, "Business Profiles");
  // The field-error state.
  await startGroup(page, `Keyboard two ${sfx}`);
  const second = await createGroup(page);
  await newProfileForm(page).getByRole("textbox", { name: "Display name" }).fill(`Keyboard Two ${sfx}`);
  await newProfileForm(page).getByRole("textbox", { name: "Website" }).fill("ftp://example.com");
  await newProfileForm(page).getByRole("button", { name: "Create and attach" }).click();
  await page.waitForURL((u) => u.pathname === `/groups/${second}/onboarding` && u.searchParams.has("error"));
  await audit(page, "onboarding field error");
});

test("[E-ONB-6] a double submit creates one Group; the second answer is the inline authorized conflict", async ({ page, context, browser }, testInfo) => {
  const sfx = suffix(testInfo);
  await signIn(page, FRANK);
  const slug = await startGroup(page, `Double ${sfx}`);
  // The same step-2 form open twice (a second tab of the same session, here without script: the native form posts and
  // the conflict returns by post/redirect/get), then both submitted.
  const noScript = await browser.newContext({
    baseURL: ORIGIN_A,
    javaScriptEnabled: false,
    storageState: await context.storageState(),
    colorScheme: testInfo.project.use.colorScheme,
    viewport: testInfo.project.use.viewport,
  });
  track(noScript);
  tracking.contexts.push(noScript);
  const second = await noScript.newPage();
  await second.goto(page.url());
  await expect(slugField(second)).toHaveValue(slug);
  const id = await createGroup(page);
  await second.getByRole("button", { name: "Create Group" }).click();
  // G018 M1 (D3): back on step 2 with the closed field error, "Nothing was created by this request" and, from the fresh
  // authorized read, the readable holder with its Open link.
  await second.waitForURL((u) => u.pathname === "/groups/new" && u.searchParams.get("error") === "slug_taken");
  await expect(second.getByRole("heading", { level: 1 })).toHaveText("Choose its address");
  await expect(slugField(second)).toHaveValue(slug);
  await expect(slugField(second)).toHaveAttribute("aria-invalid", "true");
  const describedBy = ((await slugField(second).getAttribute("aria-describedby")) ?? "").split(" ");
  expect(describedBy).toEqual(expect.arrayContaining(["slug-error", "slug-hint", "slug-advisory"]));
  await expect(second.locator("#slug-error")).toHaveText(/already in use/);
  await expect(second.getByTestId("status-message")).toHaveText(NOTHING_CREATED);
  await expect(second.getByTestId("address-holder")).toContainText(`Double ${sfx}`);
  await expect(second.getByTestId("address-holder")).toContainText("If an earlier attempt was interrupted, this may be that Group.");
  await expect(second.getByRole("link", { name: `Open Double ${sfx}` })).toHaveAttribute("href", `/groups/${id}`);
  await expect(second.getByText(/was created\.|successfully/)).toHaveCount(0);
  await second.close();
  await page.goto("/groups");
  expect((await groupNames(page)).filter((n) => n === `Double ${sfx}`)).toHaveLength(1);
});

test("[E-ONB-7] onboarding navigation leaves the member's MCP connections unchanged", async ({ page }, testInfo) => {
  const sfx = suffix(testInfo);
  await signIn(page, DAVE);
  const connections = async () => {
    await page.goto("/connections/mcp");
    await expect(page.getByRole("heading", { level: 1 })).toHaveText("Your AI client connections");
    return page.getByTestId("mcp-connection").evaluateAll((rows) => rows.map((r) => `${r.getAttribute("data-state")} ${r.textContent}`));
  };
  const before = await connections();
  expect(before.length).toBeGreaterThan(0);
  await page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link", { name: "Groups" }).click();
  await page.waitForURL((u) => u.pathname === "/groups");
  expect(await groupNames(page)).toContain("Research");
  await page.goto(`/groups/${RESEARCH}`);
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Research");
  await page.getByTestId("group-profile").getByRole("link", { name: "View Business Profile" }).click();
  await page.waitForURL((u) => u.pathname === `/business-profiles/${RESEARCH_LABS}`);
  await page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link", { name: "Business Profiles" }).click();
  await page.waitForURL((u) => u.pathname === "/business-profiles");
  // Step 2 of New Group asks core for a suggested address only. dave creates no Group here: B-MCP-E2E, which runs
  // after this browser run, pins his selectable Groups of Organization A to research alone.
  await startGroup(page, `Context ${sfx}`);
  expect(await connections()).toEqual(before);
});

// --- G060 E-ONB-8: read-only onboarding for Organization staff ---------------------------------------------------------

const NOT_ADMINISTERED = "Onboarding is administered by this Group's owners and admins.";

// carol (the Organization owner, signed in on OWNER) sets the Group-administration override through its settings form
// (the steps E-MEM-8 scripts in membership.spec.ts; no cross-spec import).
async function setGroupAuthority(owner: Page, state: "on" | "off") {
  await owner.goto("/organization/group-authority");
  const form = owner.getByTestId("group-authority-form");
  await form.getByTestId(`group-authority-${state}`).check();
  await form.getByRole("button", { name: "Save" }).click();
  await owner.waitForURL((u) => u.pathname === "/organization/group-authority" && u.searchParams.get("status") === (state === "on" ? "enabled" : "disabled"));
  await expect(owner.getByTestId(`group-authority-${state}`)).toBeChecked();
}

// The page's Primary navigation is exactly these [label, anchor] links, and each reaches its existing section by keyboard.
async function primaryNavReaches(who: Page, expected: readonly (readonly [string, string])[], label: string) {
  const links = who.getByRole("navigation", { name: "Primary" }).getByRole("link");
  await expect(links, label).toHaveText(expected.map(([text]) => text));
  for (const [i, [, href]] of expected.entries()) {
    await expect(links.nth(i), label).toHaveAttribute("href", href);
    await expect(who.locator(href), `${label} ${href}`).toHaveCount(1);
    await links.nth(i).focus();
    await who.keyboard.press("Enter");
    await expect(who, `${label} ${href}`).toHaveURL(new RegExp(`${href}$`));
    await expect(who.locator(`${href} h1, ${href} h2`).first(), `${label} ${href}`).toBeInViewport();
  }
}

// The off-state view of an incomplete Group the viewer reads but does not manage.
async function readOnlyIncomplete(who: Page, groupId: string, name: string, label: string) {
  await who.goto(`/groups/${groupId}`);
  await expect(who.getByRole("heading", { level: 1 }), label).toHaveText(name);
  await expect(who.getByTestId("group-state"), label).toHaveText("Incomplete");
  await expect(who.getByTestId("onboarding-not-administered"), label).toHaveText(NOT_ADMINISTERED);
  await expect(who.getByRole("link", { name: "Resume onboarding" }), label).toHaveCount(0);
  await expect(who.getByRole("main").getByRole("link", { name: "Members", exact: true }), label).toHaveCount(0);
  await who.goto(`/groups/${groupId}/onboarding`);
  await expect(who.getByTestId("onboarding-not-administered"), label).toHaveText(NOT_ADMINISTERED);
  await expect(who.getByTestId("profile-new-form"), label).toHaveCount(0);
  await expect(who.getByTestId("profile-existing-form"), label).toHaveCount(0);
  await who.goto("/groups");
  const row = who.getByTestId("group-list").getByRole("listitem").filter({ hasText: name });
  await expect(row, label).toHaveCount(1);
  await expect(row.getByRole("link", { name: `Resume onboarding for ${name}` }), label).toHaveCount(0);
  await expect(row.getByRole("link", { name: `Open ${name}` }), label).toHaveAttribute("href", `/groups/${groupId}`);
}

test("[E-ONB-8] Organization staff read an incomplete Group's metadata without onboarding it; the override opens only ordinary onboarding", async ({ page, context, browser }, testInfo) => {
  const sfx = suffix(testInfo);
  const name = `Pending ${sfx}`;
  // A truly incomplete Group owned by alice only (a plain member; carol and frank hold no binding in it).
  const alice = await open(browser, testInfo);
  await signIn(alice.page, ALICE);
  await startGroup(alice.page, name);
  const groupId = await createGroup(alice.page);
  await expect(alice.page.getByTestId("profile-new-form")).toBeVisible();
  const carol = await open(browser, testInfo);
  await signIn(carol.page, CAROL);
  await signIn(page, FRANK);
  try {
    // Off (the default): metadata and state only, for the owner and the admin alike.
    await readOnlyIncomplete(carol.page, groupId, name, "carol, override off");
    await readOnlyIncomplete(page, groupId, name, "frank, override off");
    await page.goto(`/groups/${groupId}/onboarding`);
    // S-NAV for this view: its Primary navigation offers only the section it renders (no absent #existing-profile).
    await primaryNavReaches(page, [["Group onboarding", "#new-profile"]], "read-only onboarding");
    await audit(page, "onboarding not administered");
    // A forged completion with frank's own CSRF token: refused, nothing attached.
    await page.goto(`/groups/new?name=${encodeURIComponent("Token probe")}`);
    const csrf = await csrfOf(page);
    const forged = await fromPage(context, ORIGIN_A, `/groups/${groupId}/onboarding/complete`, new URLSearchParams({ csrf_token: csrf, mode: "existing", business_profile_id: DESIGN_STUDIO }).toString());
    expect(forged.status).toBe(403);
    expect(forged.text).toContain("Not permitted");
    await page.goto(`/groups/${groupId}`);
    await expect(page.getByTestId("group-state")).toHaveText("Incomplete");
    // A completed Group (design) is readable metadata: its profile is a read-only view for both.
    for (const [who, label] of [
      [carol.page, "carol"],
      [page, "frank"],
    ] as const) {
      await who.goto(`/groups/${DESIGN}`);
      await expect(who.getByTestId("group-profile").getByRole("link", { name: "View Business Profile" }), label).toHaveAttribute("href", `/business-profiles/${DESIGN_STUDIO}`);
      await who.goto("/business-profiles");
      const row = who.getByTestId("profile-list").getByRole("listitem").filter({ hasText: "Design Studio" });
      await expect(row.getByTestId("profile-read-only"), label).toHaveText("Read-only");
      await row.getByRole("link", { name: "Open Design Studio" }).click();
      await who.waitForURL((u) => u.pathname === `/business-profiles/${DESIGN_STUDIO}`);
      await expect(who.getByTestId("profile-view"), label).toContainText("Design Studio");
      await expect(who.getByTestId("profile-edit-form"), label).toHaveCount(0);
    }
    await page.goto("/business-profiles");
    await audit(page, "Business Profiles (read-only rows)");

    // On: the admin may onboard it (Resume, both forms, a picker of core's can_select rows); he does not complete it.
    await setGroupAuthority(carol.page, "on");
    await page.goto(`/groups/${groupId}`);
    await expect(page.getByRole("link", { name: "Resume onboarding" })).toBeVisible();
    await expect(page.getByTestId("onboarding-not-administered")).toHaveCount(0);
    await page.goto(`/groups/${groupId}/onboarding`);
    await expect(page.getByTestId("profile-new-form")).toBeVisible();
    const existing = page.getByTestId("profile-existing-form");
    await expect(existing.getByRole("radio", { name: /Design Studio/ })).toHaveCount(1);
    await expect(existing.getByRole("radio", { name: /Research Labs/ })).toHaveCount(1);
    await expect(existing.getByRole("radio", { name: /Operations Partner/ })).toHaveCount(0);
    await expect(existing.getByRole("button", { name: "Attach selected" })).toBeVisible();
    // The manager view keeps the full navigation, each anchor reaching its rendered section.
    await primaryNavReaches(
      page,
      [
        ["Create new", "#new-profile"],
        ["Select existing", "#existing-profile"],
      ],
      "manager onboarding",
    );
    await audit(page, "onboarding (override administrator)");

    // Off again: the admin's next request shows the notice and no form; the owner's own path is unchanged.
    await setGroupAuthority(carol.page, "off");
    await page.reload();
    await expect(page.getByTestId("onboarding-not-administered")).toHaveText(NOT_ADMINISTERED);
    await expect(page.getByTestId("profile-new-form")).toHaveCount(0);
    await expect(page.getByTestId("profile-existing-form")).toHaveCount(0);
    await alice.page.goto(`/groups/${groupId}`);
    await expect(alice.page.getByRole("link", { name: "Resume onboarding" })).toBeVisible();
    await alice.page.goto(`/groups/${groupId}/onboarding`);
    await expect(alice.page.getByTestId("profile-new-form")).toBeVisible();
  } finally {
    await setGroupAuthority(carol.page, "off");
  }
});

// --- member-page landmarks and navigation (S-LANDMARKS and S-NAV of the tenant shell with Workspace areas) ------------

const MEMBER_PAGES: readonly { path: string; area: "groups" | "business-profiles" | null }[] = [
  { path: "/groups", area: "groups" },
  { path: "/groups/new", area: "groups" },
  { path: `/groups/${RESEARCH}`, area: "groups" },
  { path: "/business-profiles", area: "business-profiles" },
  { path: `/business-profiles/${RESEARCH_LABS}`, area: "business-profiles" },
  { path: "/account", area: null },
];

test("[S-LANDMARKS] member tenant pages: one banner, Primary and Workspace areas navigation, main and contentinfo; heading order", async ({ page }) => {
  await page.goto("/groups");
  await expect(page.getByRole("navigation", { name: "Workspace areas" })).toHaveCount(0);
  await signIn(page, BOB);
  for (const { path, area } of MEMBER_PAGES) {
    await page.goto(path);
    await expect(page.getByRole("banner"), path).toHaveCount(1);
    await expect(page.getByRole("navigation", { name: "Primary" }), path).toHaveCount(1);
    await expect(page.getByRole("navigation", { name: "Workspace areas" }), path).toHaveCount(1);
    await expect(page.getByRole("main"), path).toHaveCount(1);
    await expect(page.getByRole("contentinfo"), path).toHaveCount(1);
    const areas = page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link");
    await expect(areas, path).toHaveText(["Groups", "Business Profiles", "Members"]); // G019: Members is the third area
    const current = await areas.evaluateAll((as) => as.filter((a) => a.getAttribute("aria-current") === "page").map((a) => a.getAttribute("href")));
    expect(current, path).toEqual(area === null ? [] : [`/${area}`]);
    const levels = await page.locator("h1, h2, h3, h4, h5, h6").evaluateAll((hs) => hs.map((h) => Number(h.tagName[1])));
    expect(levels.filter((l) => l === 1), path).toHaveLength(1);
    expect(levels[0], path).toBe(1);
    levels.forEach((level, i) => {
      if (i > 0) expect(level - levels[i - 1], `${path} heading order ${levels.join(",")}`).toBeLessThanOrEqual(1);
    });
  }
});

test("[S-NAV] member tenant pages: Primary links reach their sections and Workspace areas links their pages, by keyboard", async ({ page }) => {
  await signIn(page, BOB);
  for (const { path } of MEMBER_PAGES.filter((p) => p.area !== null)) {
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
  for (const [label, target] of [
    ["Business Profiles", "/business-profiles"],
    ["Groups", "/groups"],
  ] as const) {
    await page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link", { name: label }).focus();
    await page.keyboard.press("Enter");
    await page.waitForURL((u) => u.pathname === target);
    await expect(page.getByRole("navigation", { name: "Workspace areas" }).getByRole("link", { name: label })).toHaveAttribute("aria-current", "page");
  }
});
