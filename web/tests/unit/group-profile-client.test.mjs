// Unit tests for lib/group-profile-client.ts (G075 Plan r2 §3.4, §3.5, C8, C10): the selected-Group context selector's
// grammar (absent is Organization-only; exactly one canonical lowercase UUID selects that Group; an empty, uppercase,
// malformed or repeated value is refused and never dropped to Organization-only), the exact outbound requests of the two
// calls (diagnostic headers, the session credential, context_group to core as a query parameter on GET and a body field
// on POST, JSON with the documented fields only), the strict reading of core's answers (exactly {group_id, group_name,
// role_label, can_edit, version} per card and {user_id, cards, narrow_required} per list; no permission role, member,
// count or Business Profile is ever accepted) and the fixed mapping of core's refusals to the label form's closed
// statuses. fetch is replaced by a recorder, so nothing leaves the process. The resolve hook maps the extensionless lib
// sibling imports to .ts, as profile-client.test.mjs does.
import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import { afterEach, beforeEach, describe, it, mock } from "node:test";

registerHooks({
  resolve(specifier, context, nextResolve) {
    if (/^\.\/[a-z-]+$/.test(specifier) && /\/lib\/[a-z-]+\.ts$/.test(context.parentURL ?? "")) {
      return nextResolve(`${specifier}.ts`, context);
    }
    return nextResolve(specifier, context);
  },
});

const gp = await import("../../lib/group-profile-client.ts");
const { deriveContext, outboundHeaders } = await import("../../lib/request-context.ts");

const CORE = "http://core.internal:8180";
const ORIGIN = "https://org-a.localhost:543";
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const CTX = deriveContext(new Headers({ "x-request-id": "req-gmp-0001", traceparent: `00-${"c".repeat(32)}-${"d".repeat(16)}-01` }));
const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };
const ID = (n) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
const SUBJECT = ID(2);
// The Group ids carry hex letters so that their uppercase form differs from the canonical lowercase one.
const GROUP = "0000000a-0000-4000-8000-00000000003b";
const OTHER_GROUP = "0000000a-0000-4000-8000-00000000003c";
const M = { origin: ORIGIN, csrfToken: CSRF, expectedVersion: 3, contextGroup: undefined };
// One card as core answers it and as the client reads it.
const CARD = { group_id: GROUP, group_name: "Research", role_label: "Lab lead", can_edit: true, version: 3 };
const READ = { groupId: GROUP, groupName: "Research", roleLabel: "Lab lead", canEdit: true, version: 3 };
const BLANK = { group_id: OTHER_GROUP, group_name: "Design", role_label: null, can_edit: false, version: 0 };
const LIST = { user_id: SUBJECT, cards: [CARD, BLANK], narrow_required: false };

function without(o, key) {
  const copy = { ...o };
  delete copy[key];
  return copy;
}

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), init });
    return reply(String(url), init);
  });
});

afterEach(() => {
  mock.restoreAll();
  for (const [key, value] of Object.entries(saved)) {
    if (value === undefined) delete process.env[key];
    else process.env[key] = value;
  }
});

const answer = (status, body) => () =>
  new Response(typeof body === "string" ? body : JSON.stringify(body), { status, headers: { "content-type": "application/json" } });

function sent(i = 0) {
  const c = calls[i];
  const raw = c.init.body;
  return { url: c.url, method: c.init.method, headers: c.init.headers, body: raw === undefined ? undefined : JSON.parse(raw), cache: c.init.cache, redirect: c.init.redirect };
}

describe("group profile client: the context selector", () => {
  it("is Organization-only without a value and that Group with exactly one canonical lowercase UUID", () => {
    assert.equal(gp.CONTEXT_PARAM, "context");
    assert.equal(gp.CONTEXT_FIELD, "context_group");
    assert.deepEqual(gp.readContextValues([]), { ok: true, group: undefined });
    assert.deepEqual(gp.readContextValues([GROUP]), { ok: true, group: GROUP });
    assert.deepEqual(gp.readContextParam(undefined), { ok: true, group: undefined });
    assert.deepEqual(gp.readContextParam(GROUP), { ok: true, group: GROUP });
    assert.deepEqual(gp.readContextParam([GROUP]), { ok: true, group: GROUP });
  });

  it("refuses an empty, uppercase, malformed or repeated value instead of dropping it to Organization-only", () => {
    for (const values of [
      [""],
      [GROUP.toUpperCase()],
      [`{${GROUP}}`],
      [` ${GROUP}`],
      [`${GROUP} `],
      [GROUP.replaceAll("-", "")],
      [GROUP.slice(0, 35)],
      ["not-a-uuid"],
      ["research"],
      [GROUP, GROUP],
      [GROUP, OTHER_GROUP],
      [GROUP, ""],
    ]) {
      assert.deepEqual(gp.readContextValues(values), { ok: false }, JSON.stringify(values));
    }
    // A page's searchParams carry a repeated parameter as an array: every non-canonical form is refused there too.
    for (const value of ["", GROUP.toUpperCase(), "not-a-uuid", [GROUP, GROUP], [GROUP, OTHER_GROUP], [""]]) {
      assert.deepEqual(gp.readContextParam(value), { ok: false }, JSON.stringify(value));
    }
    assert.equal(gp.isGroupId(GROUP), true);
    for (const bad of [GROUP.toUpperCase(), "", undefined, null, 7, `${GROUP}x`]) assert.equal(gp.isGroupId(bad), false, String(bad));
  });
});

describe("group profile client: requests", () => {
  it("lists the cards with the origin query, the read headers, no body, no-store, no redirects and no selector when absent", async () => {
    reply = answer(200, LIST);
    assert.deepEqual(await gp.listGroupCards(CTX, SECRET, ORIGIN, SUBJECT), {
      ok: true,
      value: { userId: SUBJECT, cards: [READ, { groupId: OTHER_GROUP, groupName: "Design", roleLabel: null, canEdit: false, version: 0 }], narrowRequired: false },
    });
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/profiles/${SUBJECT}/groups?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(new URL(s.url).searchParams.has("context_group"), false);
    assert.equal(s.method, "GET");
    assert.deepEqual(s.headers, { ...outboundHeaders(CTX), Authorization: `Session ${SECRET}` });
    assert.equal(s.body, undefined);
    assert.equal(s.cache, "no-store");
    assert.equal(s.redirect, "error");
  });

  it("sends a selected Group to core as exactly one context_group query value", async () => {
    reply = answer(200, { user_id: SUBJECT, cards: [CARD], narrow_required: false });
    assert.deepEqual(await gp.listGroupCards(CTX, SECRET, ORIGIN, SUBJECT, GROUP), { ok: true, value: { userId: SUBJECT, cards: [READ], narrowRequired: false } });
    const url = new URL(sent().url);
    assert.equal(url.pathname, `/internal/v1/profiles/${SUBJECT}/groups`);
    assert.deepEqual([...url.searchParams.keys()].sort(), ["context_group", "origin"]);
    assert.deepEqual(url.searchParams.getAll("context_group"), [GROUP]);
    assert.deepEqual(url.searchParams.getAll("origin"), [ORIGIN]);
  });

  it("refuses a malformed selector (invalid_request, field context_group) and a malformed subject (not_found) without calling core", async () => {
    for (const bad of ["", GROUP.toUpperCase(), "not-a-uuid"]) {
      assert.deepEqual(await gp.listGroupCards(CTX, SECRET, ORIGIN, SUBJECT, bad), { ok: false, error: "invalid_request", field: "context_group" }, bad);
    }
    for (const bad of ["x", `A${SUBJECT.slice(1)}`, `${SUBJECT}/groups`, "../me"]) {
      assert.deepEqual(await gp.listGroupCards(CTX, SECRET, ORIGIN, bad), { ok: false, error: "not_found" }, bad);
    }
    assert.equal(calls.length, 0);
  });

  it("sets the label with exactly origin, csrf_token, expected_version and role_label, and context_group only when selected", async () => {
    reply = answer(200, CARD);
    assert.deepEqual(await gp.setGroupLabel(CTX, SECRET, M, SUBJECT, GROUP, "  Lab lead "), { ok: true, value: READ });
    reply = answer(200, { ...CARD, role_label: null, version: 4 });
    assert.deepEqual(await gp.setGroupLabel(CTX, SECRET, { ...M, contextGroup: GROUP }, SUBJECT, GROUP, ""), {
      ok: true,
      value: { ...READ, roleLabel: null, version: 4 },
    });
    for (const [i, body] of [
      [0, { origin: ORIGIN, csrf_token: CSRF, expected_version: 3, role_label: "  Lab lead " }],
      [1, { origin: ORIGIN, csrf_token: CSRF, expected_version: 3, role_label: "", context_group: GROUP }],
    ]) {
      const s = sent(i);
      // The edited Group is the path's; the selector travels separately in the body.
      assert.equal(s.url, `${CORE}/internal/v1/profiles/${SUBJECT}/groups/${GROUP}/label`);
      assert.equal(s.method, "POST");
      assert.deepEqual(s.body, body);
      assert.deepEqual(s.headers, { ...outboundHeaders(CTX), Authorization: `Session ${SECRET}`, "Content-Type": "application/json" });
      assert.equal(s.cache, "no-store");
      assert.equal(s.redirect, "error");
    }
  });

  it("keeps the edited Group and the selector apart: another selected Group is sent as given, never substituted", async () => {
    reply = answer(200, CARD);
    await gp.setGroupLabel(CTX, SECRET, { ...M, contextGroup: OTHER_GROUP }, SUBJECT, GROUP, "x");
    assert.equal(sent().url, `${CORE}/internal/v1/profiles/${SUBJECT}/groups/${GROUP}/label`);
    assert.equal(sent().body.context_group, OTHER_GROUP);
  });

  it("refuses malformed ids (not_found) and a malformed selector (invalid_request) locally", async () => {
    for (const [subject, group] of [["x", GROUP], [SUBJECT, "x"], [SUBJECT, GROUP.toUpperCase()]]) {
      assert.deepEqual(await gp.setGroupLabel(CTX, SECRET, M, subject, group, "x"), { ok: false, error: "not_found" }, `${subject} ${group}`);
    }
    assert.deepEqual(await gp.setGroupLabel(CTX, SECRET, { ...M, contextGroup: "" }, SUBJECT, GROUP, "x"), { ok: false, error: "invalid_request", field: "context_group" });
    assert.equal(calls.length, 0);
  });
});

describe("group profile client: strict answers", () => {
  it("reads a card with exactly its five keys; a label row is written at least once", () => {
    assert.deepEqual(gp.CARD_KEYS, ["group_id", "group_name", "role_label", "can_edit", "version"]);
    assert.deepEqual(gp.readGroupCard(CARD), READ);
    assert.deepEqual(gp.readGroupCard({ ...CARD, role_label: null, version: 0 }), { ...READ, roleLabel: null, version: 0 });
    assert.deepEqual(gp.readGroupCard({ ...CARD, role_label: null, version: 5 }), { ...READ, roleLabel: null, version: 5 });
    assert.deepEqual(gp.readGroupCard({ ...CARD, role_label: "𝔸".repeat(100) }), { ...READ, roleLabel: "𝔸".repeat(100) });
  });

  it("refuses a card with an extra or missing key, a permission role, a bad id, name, label, hint or version", () => {
    for (const bad of [
      { ...CARD, role: "group_admin" },
      { ...CARD, members: [] },
      { ...CARD, member_count: 3 },
      { ...CARD, business_profile: "Research Labs" },
      { ...CARD, user_id: SUBJECT },
      without(CARD, "version"),
      without(CARD, "can_edit"),
      without(CARD, "role_label"),
      { ...CARD, group_id: GROUP.toUpperCase() },
      { ...CARD, group_id: "research" },
      { ...CARD, group_name: "" },
      { ...CARD, group_name: "x".repeat(201) },
      { ...CARD, role_label: "" },
      { ...CARD, role_label: "x".repeat(101) },
      { ...CARD, role_label: 7 },
      { ...CARD, version: 0 },
      { ...CARD, version: -1 },
      { ...CARD, version: 1.5 },
      { ...CARD, version: 2 ** 31 },
      { ...CARD, version: "3" },
      { ...CARD, can_edit: "true" },
      { ...CARD, can_edit: 1 },
      [],
      null,
      "card",
    ]) {
      assert.equal(gp.readGroupCard(bad), undefined, JSON.stringify(bad));
    }
  });

  it("reads a list with exactly user_id, cards and narrow_required for the subject asked for", () => {
    assert.deepEqual(gp.CARDS_KEYS, ["user_id", "cards", "narrow_required"]);
    assert.equal(gp.MAX_CARDS, 200);
    assert.deepEqual(gp.readGroupCards({ user_id: SUBJECT, cards: [], narrow_required: false }, SUBJECT), { userId: SUBJECT, cards: [], narrowRequired: false });
    // Overflow (C8): no card, no number, only the signal.
    assert.deepEqual(gp.readGroupCards({ user_id: SUBJECT, cards: [], narrow_required: true }, SUBJECT), { userId: SUBJECT, cards: [], narrowRequired: true });
    const full = Array.from({ length: 200 }, (_, i) => ({ ...BLANK, group_id: ID(1000 + i) }));
    assert.equal(gp.readGroupCards({ user_id: SUBJECT, cards: full, narrow_required: false }, SUBJECT)?.cards.length, 200);
  });

  it("refuses a list for another subject, an extra key, a total, more than 200 cards, a repeated Group or a partial overflow", () => {
    const over = Array.from({ length: 201 }, (_, i) => ({ ...BLANK, group_id: ID(2000 + i) }));
    for (const bad of [
      { ...LIST, user_id: ID(9) },
      { ...LIST, total: 2 },
      { ...LIST, next: "cursor" },
      without(LIST, "narrow_required"),
      { ...LIST, narrow_required: "false" },
      { ...LIST, cards: {} },
      { ...LIST, cards: over },
      { ...LIST, cards: [CARD, CARD] },
      { ...LIST, cards: [CARD, { ...CARD, extra: 1 }] },
      { ...LIST, narrow_required: true },
      [],
      null,
    ]) {
      assert.equal(gp.readGroupCards(bad, SUBJECT), undefined, JSON.stringify(bad).slice(0, 120));
    }
  });

  it("in a selected-Group context accepts at most that Group's card and never the overflow signal", () => {
    assert.deepEqual(gp.readGroupCards({ user_id: SUBJECT, cards: [CARD], narrow_required: false }, SUBJECT, GROUP)?.cards, [READ]);
    assert.deepEqual(gp.readGroupCards({ user_id: SUBJECT, cards: [], narrow_required: false }, SUBJECT, GROUP)?.cards, []);
    for (const bad of [
      { user_id: SUBJECT, cards: [BLANK], narrow_required: false },
      { user_id: SUBJECT, cards: [CARD, BLANK], narrow_required: false },
      { user_id: SUBJECT, cards: [], narrow_required: true },
    ]) {
      assert.equal(gp.readGroupCards(bad, SUBJECT, GROUP), undefined, JSON.stringify(bad));
    }
  });

  it("treats an answer outside the contract as malformed: a wrong shape, another Group's card, an unparsable or oversized body", async () => {
    reply = answer(200, { ...LIST, total: 2 });
    assert.deepEqual(await gp.listGroupCards(CTX, SECRET, ORIGIN, SUBJECT), { ok: false, error: "malformed" });
    reply = answer(200, { user_id: SUBJECT, cards: [BLANK], narrow_required: false });
    assert.deepEqual(await gp.listGroupCards(CTX, SECRET, ORIGIN, SUBJECT, GROUP), { ok: false, error: "malformed" });
    reply = answer(200, { ...CARD, group_id: OTHER_GROUP });
    assert.deepEqual(await gp.setGroupLabel(CTX, SECRET, M, SUBJECT, GROUP, "x"), { ok: false, error: "malformed" });
    reply = answer(200, "not json");
    assert.deepEqual(await gp.setGroupLabel(CTX, SECRET, M, SUBJECT, GROUP, "x"), { ok: false, error: "malformed" });
    reply = answer(200, `{"group_id":"${GROUP}","pad":"${"x".repeat(gp.CARD_ANSWER_MAX_BYTES)}"}`);
    assert.deepEqual(await gp.setGroupLabel(CTX, SECRET, M, SUBJECT, GROUP, "x"), { ok: false, error: "malformed" });
  });
});

describe("group profile client: refusals", () => {
  it("reads core's closed refusals with their one closed field", async () => {
    for (const [status, body, want] of [
      [404, { error: "not_found" }, { ok: false, error: "not_found" }],
      [400, { error: "invalid_request", field: "context_group" }, { ok: false, error: "invalid_request", field: "context_group" }],
      [400, { error: "invalid_request", field: "role_label" }, { ok: false, error: "invalid_request", field: "role_label" }],
      [409, { error: "version_conflict" }, { ok: false, error: "version_conflict" }],
      [403, { error: "forbidden" }, { ok: false, error: "forbidden" }],
      [403, { error: "csrf_failed" }, { ok: false, error: "csrf_failed" }],
      [401, { error: "unauthenticated" }, { ok: false, error: "unauthenticated" }],
      [503, { error: "unavailable" }, { ok: false, error: "unavailable" }],
      [500, { error: "boom" }, { ok: false, error: "unavailable" }],
      [400, { error: "boom" }, { ok: false, error: "malformed" }],
      [404, { error: "not_found", field: "context_group", reason: "x" }, { ok: false, error: "malformed" }],
    ]) {
      reply = answer(status, body);
      assert.deepEqual(await gp.setGroupLabel(CTX, SECRET, M, SUBJECT, GROUP, "x"), want, JSON.stringify([status, body]));
      assert.deepEqual(await gp.listGroupCards(CTX, SECRET, ORIGIN, SUBJECT), want, JSON.stringify([status, body]));
    }
  });

  it("reports an unreachable core as unreachable, never as a refusal", async () => {
    reply = () => {
      throw new TypeError("fetch failed");
    };
    assert.deepEqual(await gp.listGroupCards(CTX, SECRET, ORIGIN, SUBJECT), { ok: false, error: "unreachable" });
    assert.deepEqual(await gp.setGroupLabel(CTX, SECRET, M, SUBJECT, GROUP, "x"), { ok: false, error: "unreachable" });
  });

  it("maps each refusal to the label form's fixed closed status, the form-layer refusals to none", () => {
    assert.deepEqual(gp.GROUP_STATUSES, ["label_saved", "label_cleared", "stale_version", "invalid_label", "invalid_context", "not_permitted", "unavailable"]);
    const of = (error, extra = {}) => gp.groupStatusOf({ ok: false, error, ...extra });
    assert.equal(of("version_conflict"), "stale_version");
    assert.equal(of("invalid_request", { field: "role_label" }), "invalid_label");
    assert.equal(of("invalid_input", { field: "role_label" }), "invalid_label");
    assert.equal(of("invalid_request", { field: "context_group" }), "invalid_context");
    assert.equal(of("invalid_input", { field: "context_group" }), "invalid_context");
    assert.equal(of("invalid_request", { field: "expected_version" }), undefined);
    assert.equal(of("invalid_request"), undefined);
    assert.equal(of("forbidden"), "not_permitted");
    for (const error of ["unavailable", "timeout", "unreachable", "malformed"]) assert.equal(of(error), "unavailable", error);
    for (const error of ["csrf_failed", "unauthenticated", "origin_rejected", "not_available", "not_found", "unknown_host", "ambiguous_parameter", "unsupported_media_type"]) {
      assert.equal(of(error), undefined, error);
    }
  });
});
