import { createHash, randomBytes } from "node:crypto";
import { readFileSync } from "node:fs";

import AxeBuilder from "@axe-core/playwright";
import { type Browser, type BrowserContext, type Locator, type Page, type TestInfo } from "@playwright/test";

import { expectAxeClean } from "./support/contrast";
import { chooseOption, SELECT_POPUP } from "./support/modes";
import { expect, test } from "./support/page-errors";
import { observe } from "./support/sanitize";

// G074 global account profiles (Plan r3.1 §3.4-§3.10, §5 E-PRO-1..6) in the disposable HTTPS identity environment
// (infra/compose/identity.yml), projects profile-light and profile-dark, on the identity seed (infra/identity/mcp-seed.sql):
// on Organization A carol is the owner, frank an admin, alice a plain member without a Group, bob a member and the owner
// of research, dave a member of research, dana the owner of design (no sign-in) and hostile a plain member; bianca is
// Organization B's and cyrus Organization C's owner (neither signs in). Both projects run against the same database one
// after the other, after the membership projects: every scenario that needs a relationship acts on a Group the test
// itself creates (its creator is its owner, G018) between users the earlier projects never relate (dave and alice), uses
// project-distinct names, and restores what it changed (explicit names, avatars, its Groups' memberships) in a finally
// block, so the seed state is unchanged for every later project. Expected refusals are read from extra pages, so the
// default page's S-ERRORS check covers every page the user sees succeed. Core decides every read and write; these tests
// observe the UI and the BFF routes.

const ORIGIN_A = "https://org-a.localhost:543";
const ORIGIN_B = "https://org-b.localhost:543";
const ATTACKER = "https://attacker.localhost:543";
const EMULATOR = "http://emulator:4100";
const FORM = "application/x-www-form-urlencoded";
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const LINK = /https:\/\/org-a\.localhost:543\/invitations\/([0-9a-f-]{36})#([A-Za-z0-9_-]{43})/;
const AXE_TAGS = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

// Seeded user ids (infra/identity/mcp-seed.sql).
const ALICE_ID = "0f16e000-0000-4000-8000-0000000000a1";
const BOB_ID = "0f16e000-0000-4000-8000-0000000000b1";
const DAVE_ID = "0f14d000-0000-4000-8000-00000000000d";
const DANA_ID = "0f19e000-0000-4000-8000-0000000000d0";
const BIANCA_ID = "0f19e000-0000-4000-8000-0000000000b0";
const CYRUS_ID = "0f19e000-0000-4000-8000-0000000000c0";

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
const DAVE_NAME = "Dave Example";

// The Profile section's closed texts (components/account/profile-section.tsx).
const PROVIDER_SOURCE = "Your sign-in provider";
const CHOSEN_SOURCE = "The name you chose";
const NO_AVATAR = "Not set; your initials are shown";
const AVATAR_SET = "Set";

function suffix(testInfo: TestInfo): string {
  return `${testInfo.project.name.endsWith("dark") ? "d" : "l"}${randomBytes(3).toString("hex")}`;
}

// --- avatar files ----------------------------------------------------------------------------------------------------
// The accepted fixtures are the G017 logo PNGs (96x32 and 2048x16, inside the 16-2048 px bound); the refused ones are
// derived here: an SVG, the teal PNG with an animation control chunk (valid CRC), the teal PNG with bytes after IEND,
// and 513 KiB of bytes (larger than the 512 KiB bound, refused before core).
type Upload = { name: string; mimeType: string; buffer: Buffer };
const TEAL_BYTES = readFileSync("tests/fixtures/branding/logo-teal.png");
const TEAL: Upload = { name: "teal.png", mimeType: "image/png", buffer: TEAL_BYTES };
const AMBER: Upload = { name: "amber.png", mimeType: "image/png", buffer: readFileSync("tests/fixtures/branding/logo-amber.png") };
const SVG: Upload = { name: "vector.svg", mimeType: "image/svg+xml", buffer: readFileSync("tests/fixtures/branding/logo-vector.svg") };

function crc32(data: Buffer): number {
  let crc = 0xffffffff;
  for (const byte of data) {
    crc ^= byte;
    for (let k = 0; k < 8; k++) crc = crc & 1 ? (crc >>> 1) ^ 0xedb88320 : crc >>> 1;
  }
  return (crc ^ 0xffffffff) >>> 0;
}

function pngChunk(type: string, data: Buffer): Buffer {
  const length = Buffer.alloc(4);
  length.writeUInt32BE(data.length);
  const body = Buffer.concat([Buffer.from(type, "ascii"), data]);
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(body));
  return Buffer.concat([length, body, crc]);
}

// The signature (8 bytes) and the IHDR chunk (12 + 13 bytes) end at offset 33; acTL (one frame, endless) follows it.
const ANIMATED: Upload = (() => {
  const actl = Buffer.alloc(8);
  actl.writeUInt32BE(1, 0);
  actl.writeUInt32BE(0, 4);
  return { name: "animated.png", mimeType: "image/png", buffer: Buffer.concat([TEAL_BYTES.subarray(0, 33), pngChunk("acTL", actl), TEAL_BYTES.subarray(33)]) };
})();
const APPENDED: Upload = { name: "appended.png", mimeType: "image/png", buffer: Buffer.concat([TEAL_BYTES, Buffer.from("appended bytes")]) };
const OVERSIZED: Upload = { name: "large.png", mimeType: "image/png", buffer: Buffer.alloc(513 * 1024, 7) };

const sha256 = (b: Buffer | Uint8Array) => createHash("sha256").update(b).digest("hex");

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

// A same-origin request from an extra page of the context, so an expected error status never reaches S-ERRORS. The
// answer's headers are lower-cased; an image body is reported by its size and SHA-256 only.
type Fetched = { status: number; redirected: boolean; headers: Record<string, string>; text: string; size: number; sha256: string };
type Init = { method?: "GET" | "POST"; body?: string; contentType?: string; multipart?: [string, string | { bytes: number[]; type: string; name: string }][] };
async function fetchFrom(context: BrowserContext, path: string, init: Init = {}, origin = ORIGIN_A): Promise<Fetched> {
  const p = await context.newPage();
  await p.goto(`${origin}/api/health`);
  const r = await p.evaluate(
    async ({ path, init }) => {
      const request: RequestInit = { method: init.method ?? "GET", redirect: "manual", cache: "no-store" };
      if (init.multipart !== undefined) {
        const fd = new FormData();
        for (const [name, value] of init.multipart) {
          fd.append(name, typeof value === "string" ? value : new File([new Uint8Array(value.bytes)], value.name, { type: value.type }));
        }
        request.body = fd;
      } else if (init.body !== undefined) {
        request.body = init.body;
        request.headers = { "content-type": init.contentType ?? "application/x-www-form-urlencoded" };
      }
      const res = await fetch(path, request);
      const redirected = res.type === "opaqueredirect";
      const headers: Record<string, string> = {};
      res.headers.forEach((v, k) => {
        headers[k] = v;
      });
      const bytes = redirected ? new Uint8Array(0) : new Uint8Array(await res.arrayBuffer());
      const digest = Array.from(new Uint8Array(await crypto.subtle.digest("SHA-256", bytes)), (b) => b.toString(16).padStart(2, "0")).join("");
      const image = (headers["content-type"] ?? "").startsWith("image/");
      return { status: res.status, redirected, headers, text: image ? "" : new TextDecoder().decode(bytes), size: bytes.byteLength, sha256: digest };
    },
    { path, init },
  );
  await p.close();
  return r;
}

async function profileJSON(context: BrowserContext, id: string): Promise<{ status: number; body: Record<string, unknown>; cache: string }> {
  const r = await fetchFrom(context, `/api/profiles/${id}`);
  return { status: r.status, body: JSON.parse(r.text) as Record<string, unknown>, cache: r.headers["cache-control"] ?? "" };
}

async function apiAccountName(context: BrowserContext, origin: string): Promise<unknown> {
  const r = await fetchFrom(context, "/api/account", {}, origin);
  expect(r.status, `/api/account on ${origin}`).toBe(200);
  return (JSON.parse(r.text) as Record<string, unknown>).display_name;
}

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

// --- the Profile section ------------------------------------------------------------------------------------------------

const section = (page: Page) => page.locator("#profile");

async function expectProfile(page: Page, want: { name: string; source: string; avatar: string }) {
  await expect(page.getByTestId("profile-effective-name")).toHaveText(want.name);
  await expect(page.getByTestId("profile-name-source")).toHaveText(want.source);
  await expect(page.getByTestId("profile-avatar-state")).toHaveText(want.avatar);
}

// Waits for the profile form's 303 back to /account and returns its closed profile_status.
async function submitted(page: Page, button: string): Promise<string | null> {
  await section(page).getByRole("button", { name: button, exact: true }).click();
  await page.waitForURL((u) => u.pathname === "/account" && u.searchParams.has("profile_status"));
  return new URL(page.url()).searchParams.get("profile_status");
}

async function saveName(page: Page, value: string): Promise<string | null> {
  await page.goto("/account");
  await page.getByRole("textbox", { name: "Name", exact: true }).fill(value);
  return submitted(page, "Save name");
}

async function uploadAvatar(page: Page, file: Upload): Promise<string | null> {
  await page.goto("/account");
  await page.locator("#profile-avatar-file").setInputFiles(file);
  return submitted(page, "Upload avatar");
}

async function removeAvatar(page: Page): Promise<string | null> {
  await page.goto("/account");
  return submitted(page, "Remove avatar");
}

// The current avatar's versioned URL, once the image has loaded (the kit Avatar renders its image only after loading).
async function avatarSrc(page: Page): Promise<string> {
  const image = page.getByTestId("profile-avatar-image");
  await expect(image).toBeVisible();
  const src = (await image.getAttribute("src")) ?? "";
  expect(src).toMatch(/^\/profiles\/[0-9a-f-]{36}\/avatar\?v=[1-9][0-9]*$/);
  return src;
}
const versionOf = (src: string) => Number(new URL(src, ORIGIN_A).searchParams.get("v"));

// Restores the seed state of the signed-in user's profile: no explicit name, no avatar (best effort in finally blocks).
async function restoreSeed(page: Page) {
  await page.goto("/account");
  if ((await page.getByTestId("profile-name-source").textContent())?.trim() === CHOSEN_SOURCE) await saveName(page, "");
  if ((await page.getByTestId("profile-avatar-remove-form").count()) > 0) await removeAvatar(page);
  await page.goto("/account");
  await expect(page.getByTestId("profile-name-source")).toHaveText(PROVIDER_SOURCE);
  await expect(page.getByTestId("profile-avatar-state")).toHaveText(NO_AVATAR);
}

// --- Groups and invitations (the G018 New Group and G019 invitation flows) ---------------------------------------------

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

// The candidate names of the invitation form's Workspace member picker (the kit Select's options), then closed.
async function pickerNames(page: Page): Promise<string[]> {
  const popup = page.locator(SELECT_POPUP);
  await page.getByTestId("invitation-form").getByLabel("Workspace member").click();
  await expect(popup).toBeVisible();
  const names = (await popup.getByRole("option").allTextContents()).map((t) => t.trim());
  await page.keyboard.press("Escape");
  await expect(popup).toHaveCount(0);
  return names;
}

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

// --- E-PRO-1..6 -----------------------------------------------------------------------------------------------------------

test("[E-PRO-1] a self edit on host A is the name on host B, in the shell and in /api/account; clearing returns the seed name", async ({ page, browser }, testInfo) => {
  test.slow();
  tracking.allowed.add(ORIGIN_B);
  const chosen = `Alice Profile ${suffix(testInfo)}`;
  await signIn(page, ALICE);
  try {
    await expectProfile(page, { name: ALICE_NAME, source: PROVIDER_SOURCE, avatar: NO_AVATAR });
    await expect(page.getByTestId("profile-avatar-fallback")).toHaveText("AE");
    await expect(page.getByRole("textbox", { name: "Name", exact: true })).toHaveValue("");
    expect(await saveName(page, chosen)).toBe("profile_saved");
    await expect(page.getByTestId("profile-status")).toHaveText("Your name was saved. People you work with see it on newly loaded pages.");
    await expectProfile(page, { name: chosen, source: CHOSEN_SOURCE, avatar: NO_AVATAR });
    await expect(page.getByRole("textbox", { name: "Name", exact: true })).toHaveValue(chosen);
    // E1: the session's presentation name on the account page, in /api/account and in the shell indicator.
    await expect(page.getByTestId("display-name")).toHaveText(chosen);
    expect(await apiAccountName(page.context(), ORIGIN_A)).toBe(chosen);
    await page.goto("/");
    await expect(page.getByTestId("context-indicator")).toContainText(`Signed in as ${chosen}`);
    // Host B reads the same global row.
    const b = await signedIn(browser, testInfo, ALICE, ORIGIN_B);
    await expectProfile(b.page, { name: chosen, source: CHOSEN_SOURCE, avatar: NO_AVATAR });
    await expect(b.page.getByTestId("display-name")).toHaveText(chosen);
    expect(await apiAccountName(b.context, ORIGIN_B)).toBe(chosen);
    await b.page.goto(`${ORIGIN_B}/`);
    await expect(b.page.getByTestId("context-indicator")).toContainText(`Signed in as ${chosen}`);
    // Clearing on host B returns the current provider (seed) name everywhere.
    expect(await saveName(b.page, "")).toBe("profile_saved");
    await expectProfile(b.page, { name: ALICE_NAME, source: PROVIDER_SOURCE, avatar: NO_AVATAR });
    await page.goto("/account");
    await expectProfile(page, { name: ALICE_NAME, source: PROVIDER_SOURCE, avatar: NO_AVATAR });
    await expect(page.getByTestId("display-name")).toHaveText(ALICE_NAME);
    expect(await apiAccountName(page.context(), ORIGIN_A)).toBe(ALICE_NAME);
  } finally {
    await restoreSeed(page);
  }
});

test("[E-PRO-2] avatar set, replace and remove; SVG, animated, trailing-byte and oversized files are refused and store nothing", async ({ page }) => {
  test.slow();
  await signIn(page, ALICE);
  try {
    for (const [file, status, text] of [
      [SVG, "invalid_image", "That file is not an accepted avatar"],
      [ANIMATED, "invalid_image", "That file is not an accepted avatar"],
      [APPENDED, "invalid_image", "That file is not an accepted avatar"],
      [OVERSIZED, "payload_too_large", "That file is larger than 512 KiB. Nothing was changed."],
    ] as const) {
      expect(await uploadAvatar(page, file), file.name).toBe(status);
      await expect(page.getByTestId("profile-status"), file.name).toContainText(text);
      await expectProfile(page, { name: ALICE_NAME, source: PROVIDER_SOURCE, avatar: NO_AVATAR });
      await expect(page.getByTestId("profile-avatar-image")).toHaveCount(0);
      await expect(page.getByTestId("profile-avatar-remove-form")).toHaveCount(0);
    }
    // A PNG is accepted: checked, re-encoded (metadata stripped) and served at a versioned, session-authorized URL.
    expect(await uploadAvatar(page, TEAL)).toBe("avatar_set");
    await expect(page.getByTestId("profile-status")).toHaveText("Your new avatar was checked, stored and is now shown.");
    await expectProfile(page, { name: ALICE_NAME, source: PROVIDER_SOURCE, avatar: AVATAR_SET });
    const first = await avatarSrc(page);
    expect(first.startsWith(`/profiles/${ALICE_ID}/avatar?v=`)).toBe(true);
    await expect(page.getByTestId("profile-avatar-fallback")).toHaveCount(0);
    const one = await fetchFrom(page.context(), first);
    expect(one.status).toBe(200);
    expect(one.headers["content-type"]).toBe("image/png");
    expect(one.headers["cache-control"]).toBe("private, max-age=300");
    expect(one.headers["x-content-type-options"]).toBe("nosniff");
    expect(one.headers["content-disposition"]).toBe("inline");
    expect(one.headers.etag).toBe(`"v${versionOf(first)}"`);
    expect(one.size).toBeGreaterThan(0);
    expect(one.sha256, "the stored bytes are the re-encoded image, not the upload").not.toBe(sha256(TEAL_BYTES));
    // A replacement advances the version; the old URL is a no-store 404.
    expect(await uploadAvatar(page, AMBER)).toBe("avatar_set");
    const second = await avatarSrc(page);
    expect(versionOf(second)).toBe(versionOf(first) + 1);
    const stale = await fetchFrom(page.context(), first);
    expect([stale.status, stale.headers["cache-control"]]).toEqual([404, "no-store"]);
    expect((await fetchFrom(page.context(), second)).status).toBe(200);
    // Removal shows the initials again; the last URL is a no-store 404.
    expect(await removeAvatar(page)).toBe("avatar_removed");
    await expect(page.getByTestId("profile-status")).toHaveText("Your avatar was removed. Your initials are shown instead.");
    await expectProfile(page, { name: ALICE_NAME, source: PROVIDER_SOURCE, avatar: NO_AVATAR });
    await expect(page.getByTestId("profile-avatar-fallback")).toHaveText("AE");
    await expect(page.getByTestId("profile-avatar-remove-form")).toHaveCount(0);
    const removed = await fetchFrom(page.context(), second);
    expect([removed.status, removed.headers["cache-control"]]).toEqual([404, "no-store"]);
  } finally {
    await restoreSeed(page);
  }
});

test("[E-PRO-3] the API privacy matrix: self, Organization staff and shared-Group peers read; everyone else, foreign and anonymous get one 404", async ({ page, browser }, testInfo) => {
  test.slow();
  await signIn(page, ALICE);
  const bob = await signedIn(browser, testInfo, BOB);
  const restore = [bob.page];
  try {
    // A5/A6 need avatars: bob's and dave's, each uploaded through the owner's own Profile section.
    expect(await uploadAvatar(bob.page, TEAL)).toBe("avatar_set");
    const bobAvatar = await avatarSrc(bob.page);
    const frank = await signedIn(browser, testInfo, FRANK);
    const carol = await signedIn(browser, testInfo, CAROL);
    const dave = await signedIn(browser, testInfo, DAVE);
    restore.push(dave.page);
    expect(await uploadAvatar(dave.page, TEAL)).toBe("avatar_set");
    const daveAvatar = await avatarSrc(dave.page);
    const hostile = await signedIn(browser, testInfo, HOSTILE);
    const KEYS = ["avatar", "name", "user_id"];
    // P1 self, P2 Organization owner and admin, P3 shared research Group in both directions: exactly the three keys.
    for (const [who, label, id, name] of [
      [page.context(), "alice -> alice", ALICE_ID, ALICE_NAME],
      [frank.context, "frank (admin) -> alice", ALICE_ID, ALICE_NAME],
      [carol.context, "carol (owner) -> dave", DAVE_ID, DAVE_NAME],
      [dave.context, "dave -> bob (research)", BOB_ID, "Bob Example"],
      [bob.context, "bob -> dave (research)", DAVE_ID, DAVE_NAME],
    ] as const) {
      const r = await profileJSON(who, id);
      expect(r.status, label).toBe(200);
      expect(r.cache, label).toBe("no-store");
      expect(Object.keys(r.body).sort(), label).toEqual(KEYS);
      expect([r.body.user_id, r.body.name], label).toEqual([id, name]);
    }
    const viaDave = await profileJSON(dave.context, BOB_ID);
    expect(viaDave.body.avatar).toEqual({ version: versionOf(bobAvatar), content_type: "image/png" });
    // N1 same tenant without a shared Group or staff role, N2 foreign tenants: the same 404 body.
    const refused: string[] = [];
    for (const [who, label, id] of [
      [page.context(), "alice -> dave", DAVE_ID],
      [dave.context, "dave -> dana", DANA_ID],
      [hostile.context, "hostile -> bob", BOB_ID],
      [hostile.context, "hostile -> bianca (B only)", BIANCA_ID],
      [page.context(), "alice -> cyrus (C only)", CYRUS_ID],
      [page.context(), "alice -> an unknown user", "0f16e000-0000-4000-8000-0000000000ff"],
    ] as const) {
      const r = await fetchFrom(who, `/api/profiles/${id}`);
      expect([r.status, r.headers["cache-control"]], label).toEqual([404, "no-store"]);
      refused.push(r.text);
    }
    expect(new Set(refused).size, "one identical refusal body").toBe(1);
    expect(JSON.parse(refused[0])).toEqual({ error: "not_found" });
    // A5: a shared-Group peer reads bob's avatar (private, nosniff). A6: whoever the profile read refuses gets the same
    // no-store 404 for the avatar: hostile (foreign) for bob's, and alice, who shares no Group with dave (N1 above), for
    // dave's. Earlier membership specs leave alice in a Group of bob's, so her access to his avatar is judged by her own
    // profile read of him: the two answers must agree.
    const peer = await fetchFrom(dave.context, bobAvatar);
    expect(peer.status).toBe(200);
    expect([peer.headers["content-type"], peer.headers["cache-control"], peer.headers["x-content-type-options"]]).toEqual(["image/png", "private, max-age=300", "nosniff"]);
    for (const [who, label, src] of [[page.context(), "alice -> dave", daveAvatar], [hostile.context, "hostile -> bob", bobAvatar]] as const) {
      const r = await fetchFrom(who, src);
      expect([r.status, r.headers["cache-control"], r.size], label).toEqual([404, "no-store", 0]);
    }
    const aliceProfile = await profileJSON(page.context(), BOB_ID);
    const aliceImage = await fetchFrom(page.context(), bobAvatar);
    expect(aliceImage.status, "alice's avatar read agrees with her profile read of bob").toBe(aliceProfile.status);
    expect(aliceImage.headers["cache-control"]).toBe(aliceProfile.status === 200 ? "private, max-age=300" : "no-store");
    // N4: anonymous requests.
    const anon = await open(browser, testInfo);
    const a1 = await fetchFrom(anon.context, `/api/profiles/${ALICE_ID}`);
    expect([a1.status, a1.headers["cache-control"], JSON.parse(a1.text)]).toEqual([401, "no-store", { error: "unauthenticated" }]);
    const a2 = await fetchFrom(anon.context, bobAvatar);
    expect([a2.status, a2.headers["cache-control"]]).toEqual([404, "no-store"]);
    // N5: no address, provider subject, storage key or object id in the pages and answers that carry profile data.
    await page.goto("/account");
    await frank.page.goto("/organization/members");
    await expect(frank.page.getByTestId("member-table")).toBeVisible();
    const texts = [await page.content(), await frank.page.content(), JSON.stringify(viaDave.body)];
    for (const text of texts) {
      // The address leak is the JSON key (plain or escaped in the page's flight data); the members page's own invitation
      // form legitimately carries an input of type="email".
      for (const leak of ["identity.test", "user_01", "avatars/", "session_id", '"email":', '\\"email\\":', '"organizations"', '"object_id"']) {
        expect(text, `no ${leak}`).not.toContain(leak);
      }
    }
  } finally {
    for (const p of restore) await restoreSeed(p);
  }
});

test("[E-PRO-4] a revoked relationship: a peer's profile and avatar read until the shared Group membership is removed, then 404", async ({ page, browser }, testInfo) => {
  test.slow();
  const dave = await signedIn(browser, testInfo, DAVE);
  await signIn(page, ALICE);
  try {
    expect(await uploadAvatar(dave.page, TEAL)).toBe("avatar_set");
    const daveAvatar = await avatarSrc(dave.page);
    // Unrelated first.
    expect((await profileJSON(page.context(), DAVE_ID)).status).toBe(404);
    const groupId = await createOwnGroup(dave.page, `Profiles ${suffix(testInfo)}`);
    await accept(page, await inviteToGroup(dave.page, groupId, ALICE_NAME), groupId);
    const shared = await profileJSON(page.context(), DAVE_ID);
    expect([shared.status, shared.body.name, shared.body.avatar]).toEqual([200, DAVE_NAME, { version: versionOf(daveAvatar), content_type: "image/png" }]);
    expect((await fetchFrom(page.context(), daveAvatar)).status).toBe(200);
    // The owner removes alice: her next reads are refused, the avatar URL on a fresh (uncached) fetch included.
    await removeFromGroup(dave.page, groupId, ALICE_NAME);
    const after = await fetchFrom(page.context(), `/api/profiles/${DAVE_ID}`);
    expect([after.status, JSON.parse(after.text)]).toEqual([404, { error: "not_found" }]);
    const image = await fetchFrom(page.context(), daveAvatar);
    expect([image.status, image.headers["cache-control"], image.size]).toEqual([404, "no-store", 0]);
  } finally {
    await restoreSeed(dave.page);
  }
});

test("[E-PRO-5] the invitation picker shows an edited name only to issuers who may read it, and never an avatar (R1)", async ({ page, browser }, testInfo) => {
  test.slow();
  const sfx = suffix(testInfo);
  const override = `Alice Override ${sfx}`;
  await signIn(page, ALICE);
  try {
    expect(await saveName(page, override)).toBe("profile_saved");
    expect(await uploadAvatar(page, TEAL)).toBe("avatar_set");
    const aliceAvatar = await avatarSrc(page);
    // dave (a plain member who owns his own Groups) shares no Group with alice: the candidate set is unchanged and she
    // is listed with her provider (seed) name; no avatar reference reaches the page; her profile is a 404 for him.
    const dave = await signedIn(browser, testInfo, DAVE);
    const x = await createOwnGroup(dave.page, `Picker X ${sfx}`);
    const y = await createOwnGroup(dave.page, `Picker Y ${sfx}`);
    await dave.page.goto(membersOf(y));
    let names = await pickerNames(dave.page);
    expect(names, "alice is a candidate with her seed name").toContain(ALICE_NAME);
    expect(names).not.toContain(override);
    expect(await dave.page.content()).not.toContain(`/profiles/${ALICE_ID}/avatar`);
    expect((await profileJSON(dave.context, ALICE_ID)).status).toBe(404);
    // An Organization admin may read her profile: frank's own new Group's picker and the workspace members table show
    // the edited name, and the table her thumbnail.
    const frank = await signedIn(browser, testInfo, FRANK);
    const f = await createOwnGroup(frank.page, `Picker F ${sfx}`);
    await frank.page.goto(membersOf(f));
    names = await pickerNames(frank.page);
    expect(names).toContain(override);
    expect(names).not.toContain(ALICE_NAME);
    await frank.page.goto("/organization/members");
    const row = memberRow(frank.page, override);
    await expect(row).toHaveCount(1);
    await expect(row.getByTestId("member-avatar-image")).toBeVisible();
    expect(await row.getByTestId("member-avatar-image").getAttribute("src")).toBe(aliceAvatar);
    // Once alice joins dave's Group X, dave's other picker and X's member table show her edited name and thumbnail.
    await accept(page, await inviteToGroup(dave.page, x, ALICE_NAME), x);
    await dave.page.goto(membersOf(x));
    await expect(memberRow(dave.page, override).getByTestId("member-avatar-image")).toBeVisible();
    await dave.page.goto(membersOf(y));
    names = await pickerNames(dave.page);
    expect(names).toContain(override);
    expect(names).not.toContain(ALICE_NAME);
    const visible = await profileJSON(dave.context, ALICE_ID);
    expect([visible.status, visible.body.name, visible.body.avatar]).toEqual([200, override, { version: versionOf(aliceAvatar), content_type: "image/png" }]);
    // dave removes her from X: the seed name again, and her profile is a 404 for him.
    await removeFromGroup(dave.page, x, override);
    await dave.page.goto(membersOf(y));
    names = await pickerNames(dave.page);
    expect(names).toContain(ALICE_NAME);
    expect(names).not.toContain(override);
    expect((await profileJSON(dave.context, ALICE_ID)).status).toBe(404);
  } finally {
    await restoreSeed(page);
  }
});

test("[E-PRO-6] forged profile POSTs change nothing: a wrong CSRF token, a cross-site page and a stale version", async ({ page, context }) => {
  test.slow();
  tracking.allowed.add(ATTACKER);
  await signIn(page, ALICE);
  const chosen = `Alice Stale ${randomBytes(3).toString("hex")}`;
  try {
    const version = await page.locator('#profile input[name="expected_version"]').first().inputValue();
    // A wrong CSRF token, on the name form and the avatar form: the fixed refusal page, nothing changed.
    const forged = await fetchFrom(context, "/account/profile/update", {
      method: "POST",
      contentType: FORM,
      body: new URLSearchParams({ csrf_token: "forged", expected_version: version, name: "Mallory" }).toString(),
    });
    expect(forged.status).toBe(403);
    expect(forged.text).toContain("Form expired");
    expect(forged.text).toContain("Nothing was changed.");
    const forgedImage = await fetchFrom(context, "/account/profile/avatar", {
      method: "POST",
      multipart: [["csrf_token", "forged"], ["expected_version", version], ["avatar", { bytes: [...TEAL_BYTES], type: "image/png", name: "teal.png" }]],
    });
    expect(forgedImage.status).toBe(403);
    // A cross-site page posting to the profile form routes is refused by origin.
    const xs = await context.newPage();
    for (const route of ["/account/profile/update", "/account/profile/avatar/remove"]) {
      await xs.goto(`${ATTACKER}/`);
      await xs.evaluate(
        ({ action, version }) => {
          const form = document.createElement("form");
          form.method = "post";
          form.action = action;
          for (const [name, value] of [["csrf_token", "guessed"], ["expected_version", version], ["name", "Mallory"]]) {
            const input = document.createElement("input");
            input.type = "hidden";
            input.name = name;
            input.value = value;
            form.append(input);
          }
          document.body.append(form);
          form.submit();
        },
        { action: `${ORIGIN_A}${route}`, version },
      );
      await xs.waitForURL((u) => u.origin === ORIGIN_A);
      await expect(xs.locator("h1"), route).toHaveText("Request refused");
    }
    await xs.close();
    await page.goto("/account");
    await expectProfile(page, { name: ALICE_NAME, source: PROVIDER_SOURCE, avatar: NO_AVATAR });
    // A stale tab: the first tab saves, the second tab's older form is a version conflict with the alert.
    const stale = await context.newPage();
    await stale.goto("/account");
    expect(await saveName(page, chosen)).toBe("profile_saved");
    await stale.getByRole("textbox", { name: "Name", exact: true }).fill("Alice From An Old Tab");
    expect(await submitted(stale, "Save name")).toBe("stale_version");
    await expect(stale.getByTestId("profile-alert")).toContainText("Your profile changed since this page was loaded");
    await expectProfile(stale, { name: chosen, source: CHOSEN_SOURCE, avatar: NO_AVATAR });
    await stale.close();
  } finally {
    await restoreSeed(page);
  }
});

test("[S-PROFILE-A11Y] the Profile section by keyboard; axe at 1280 and 320 on /account and on a members table with a thumbnail", async ({ page, browser }, testInfo) => {
  test.slow();
  const chosen = `Alice Keys ${suffix(testInfo)}`;
  await signIn(page, ALICE);
  try {
    await page.goto("/account");
    await expect(page.getByRole("navigation", { name: "Primary" }).getByRole("link", { name: "Profile" })).toHaveAttribute("href", "#profile");
    await audit(page, "account Profile section (seed state)");
    const name = page.getByRole("textbox", { name: "Name", exact: true });
    await tabTo(page, name, "Name");
    await page.keyboard.type(chosen);
    await tabTo(page, section(page).getByRole("button", { name: "Save name", exact: true }), "Save name");
    await page.keyboard.press("Enter");
    await page.waitForURL((u) => u.pathname === "/account" && u.searchParams.get("profile_status") === "profile_saved");
    await expectProfile(page, { name: chosen, source: CHOSEN_SOURCE, avatar: NO_AVATAR });
    expect(await uploadAvatar(page, TEAL)).toBe("avatar_set");
    await avatarSrc(page);
    await audit(page, "account Profile section (name, avatar and status notice)");
    // The workspace members table as an admin: the thumbnail beside the edited name.
    const frank = await signedIn(browser, testInfo, FRANK);
    await frank.page.goto("/organization/members");
    await expect(memberRow(frank.page, chosen).getByTestId("member-avatar-image")).toBeVisible();
    await audit(frank.page, "workspace members with a thumbnail");
  } finally {
    await restoreSeed(page);
  }
});
