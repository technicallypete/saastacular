// Unit tests for the G075 Group member label form and the card JSON route (Plan r2 §3.5, §3.7, C6, C10): the closed
// group_status table, the 303 back to /profiles/<id>#groups with one closed status and the request's own validated
// selector (never the submitted label, token or any other request value), the label form's fields (csrf_token,
// expected_version in the grammar, exactly one role_label where "" clears), the form-route order (method, encoding, exact
// Origin, bounded body, the selector's grammar, the fields, the session) before core, the selector passed through to core
// as context_group and refused as invalid_context when empty, malformed or repeated, the fixed refusal pages that never
// claim a change, and GET /api/profiles/[id]/groups: exactly {user_id, cards, narrow_required} (400 for a bad ?context=,
// 401/404/503, no-store). fetch is a recorder; nothing leaves the process. The resolve hook maps lib siblings and the
// routes' @/lib aliases to .ts, as profile-forms.test.mjs does.
import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import { afterEach, beforeEach, describe, it, mock } from "node:test";

const LIB = new URL("../../lib/", import.meta.url);
registerHooks({
  resolve(specifier, context, nextResolve) {
    if (/^\.\/[a-z-]+$/.test(specifier) && /\/lib\/[a-z-]+\.ts$/.test(context.parentURL ?? "")) {
      return nextResolve(`${specifier}.ts`, context);
    }
    const alias = /^@\/lib\/([a-z-]+)$/.exec(specifier);
    if (alias !== null && /\/app\/.+\/route\.ts$/.test(context.parentURL ?? "")) {
      return nextResolve(new URL(`${alias[1]}.ts`, LIB).href, context);
    }
    return nextResolve(specifier, context);
  },
});

const gp = await import("../../lib/group-profile-client.ts");
const { cookieName } = await import("../../lib/session.ts");
const labelRoute = await import("../../app/profiles/[id]/groups/[group]/label/route.ts");
const cardsRoute = await import("../../app/api/profiles/[id]/groups/route.ts");

const CORE = "http://core.internal:8180";
const HOST = "org-a.localhost:543";
const ORIGIN = `https://${HOST}`;
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const ENV = {
  PRODUCT_NAME: "Test Product",
  PLATFORM_BASE_DOMAIN: "example.test",
  CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test",
  CORE_INTERNAL_URL: CORE,
  IDENTITY_COOKIE_PROFILE: "https",
};
const ID = (n) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
const SUBJECT = ID(2);
// The Group ids carry hex letters so that their uppercase form differs from the canonical lowercase one.
const GROUP = "0000000a-0000-4000-8000-00000000003b";
const OTHER_GROUP = "0000000a-0000-4000-8000-00000000003c";
const CARD = { group_id: GROUP, group_name: "Research", role_label: "Lab lead", can_edit: true, version: 4 };
const COOKIE = () => `${cookieName("session", "https")}=${SECRET}`;
const LABEL_PATH = `/profiles/${SUBJECT}/groups/${GROUP}/label`;
// The page a label form returns to: group_status first, then the selector, at the Groups section.
const back = (status, context) => `/profiles/${SUBJECT}?group_status=${status}${context === undefined ? "" : `&context=${context}`}#groups`;

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    const raw = init.body;
    calls.push({ url: String(url), method: init.method, headers: init.headers, body: raw === undefined ? undefined : JSON.parse(raw) });
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

const json = (status, body) => () => Response.json(body, { status });
const form = (fields) => new URLSearchParams(fields).toString();
function post(path, body, headers = {}) {
  return new Request(`${ORIGIN}${path}`, {
    method: "POST",
    headers: { host: HOST, origin: ORIGIN, "content-type": "application/x-www-form-urlencoded", cookie: COOKIE(), ...headers },
    body,
  });
}
function get(path, headers = {}) {
  return new Request(`${ORIGIN}${path}`, { method: "GET", headers: { host: HOST, cookie: COOKIE(), ...headers } });
}
const at = (res) => [res.status, res.headers.get("location"), res.headers.get("cache-control")];
const label = (fields, headers, id = SUBJECT, group = GROUP) =>
  labelRoute.POST(post(`/profiles/${id}/groups/${group}/label`, typeof fields === "string" ? fields : form(fields), headers), {
    params: Promise.resolve({ id, group }),
  });
const GOOD = { csrf_token: CSRF, expected_version: "4", role_label: "Lab lead" };

describe("label form helpers", () => {
  it("has exactly the seven closed statuses and returns to the Groups section with one of them and the validated selector", () => {
    assert.deepEqual(gp.GROUP_STATUSES, ["label_saved", "label_cleared", "stale_version", "invalid_label", "invalid_context", "not_permitted", "unavailable"]);
    for (const status of gp.GROUP_STATUSES) {
      assert.deepEqual(at(gp.groupBack(SUBJECT, undefined, status)), [303, back(status), "no-store"], status);
      assert.deepEqual(at(gp.groupBack(SUBJECT, GROUP, status)), [303, back(status, GROUP), "no-store"], status);
    }
    assert.equal(gp.profilePagePath(SUBJECT), `/profiles/${SUBJECT}`);
    assert.equal(gp.profilePagePath(SUBJECT, GROUP), `/profiles/${SUBJECT}?context=${GROUP}`);
    assert.equal(gp.labelFormAction(SUBJECT, GROUP), LABEL_PATH);
  });

  it("reads the label form: one role_label (empty clears), the CSRF token and a version in the grammar", () => {
    const read = (fields) => gp.readLabelForm(new URLSearchParams(fields));
    assert.deepEqual(read(GOOD), { m: { csrfToken: CSRF, expectedVersion: 4 }, label: "Lab lead" });
    assert.deepEqual(read({ csrf_token: CSRF, expected_version: "0", role_label: "" }), { m: { csrfToken: CSRF, expectedVersion: 0 }, label: "" });
    // The label is passed as submitted: core trims and validates it.
    assert.deepEqual(read({ ...GOOD, role_label: "  <b>x</b> " }).label, "  <b>x</b> ");
    for (const fields of [
      { csrf_token: CSRF, expected_version: "4" },
      [["csrf_token", CSRF], ["expected_version", "4"], ["role_label", "A"], ["role_label", "B"]],
      [["csrf_token", CSRF], ["csrf_token", CSRF], ["expected_version", "4"], ["role_label", "A"]],
      { csrf_token: CSRF, expected_version: "-1", role_label: "A" },
      { csrf_token: CSRF, expected_version: "04", role_label: "A" },
      { csrf_token: CSRF, expected_version: "1.5", role_label: "A" },
      { csrf_token: CSRF, role_label: "A" },
    ]) {
      assert.equal(read(fields), undefined, JSON.stringify(fields));
    }
  });

  it("answers success and the recoverable refusals with 303 and one closed status; core's invalid selector is not carried back", () => {
    const answer = (result, context) => at(gp.groupLabelAnswer(result, SUBJECT, context));
    assert.deepEqual(answer({ ok: true, value: { roleLabel: "Lab lead" } }, GROUP), [303, back("label_saved", GROUP), "no-store"]);
    assert.deepEqual(answer({ ok: true, value: { roleLabel: null } }, undefined), [303, back("label_cleared"), "no-store"]);
    for (const [failure, status] of [
      [{ ok: false, error: "version_conflict" }, "stale_version"],
      [{ ok: false, error: "invalid_request", field: "role_label" }, "invalid_label"],
      [{ ok: false, error: "invalid_input", field: "role_label" }, "invalid_label"],
      [{ ok: false, error: "forbidden" }, "not_permitted"],
      [{ ok: false, error: "unavailable" }, "unavailable"],
      [{ ok: false, error: "timeout" }, "unavailable"],
      [{ ok: false, error: "malformed" }, "unavailable"],
    ]) {
      assert.deepEqual(answer(failure, GROUP), [303, back(status, GROUP), "no-store"], failure.error);
    }
    assert.deepEqual(answer({ ok: false, error: "invalid_request", field: "context_group" }, GROUP), [303, back("invalid_context"), "no-store"]);
  });

  it("answers the session, CSRF, origin and visibility refusals with fixed pages that never claim a change", async () => {
    for (const [error, status, text] of [
      ["csrf_failed", 403, "Form expired"],
      ["unauthenticated", 401, "Not signed in"],
      ["origin_rejected", 403, "Request refused"],
      ["not_available", 403, "Not available"],
      ["not_found", 404, "Not found"],
      ["ambiguous_parameter", 400, "Request not understood"],
    ]) {
      const res = gp.groupLabelAnswer({ ok: false, error }, SUBJECT, GROUP);
      assert.equal(res.status, status, error);
      assert.equal(res.headers.get("location"), null, error);
      assert.equal(res.headers.get("cache-control"), "no-store", error);
      const body = await res.text();
      assert.ok(body.includes(text), error);
      assert.equal(body.includes("was saved"), false, error);
    }
    // invalid_request about another field (a malformed version reached core) is the fixed bad-request page.
    assert.equal(gp.groupLabelAnswer({ ok: false, error: "invalid_request", field: "expected_version" }, SUBJECT, GROUP).status, 400);
  });
});

describe("label form route", () => {
  it("sets the label with the documented fields and the selector, then returns 303 with label_saved and the same selector", async () => {
    reply = json(200, CARD);
    const res = await label({ ...GOOD, context_group: GROUP });
    assert.deepEqual(at(res), [303, back("label_saved", GROUP), "no-store"]);
    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, `${CORE}/internal/v1/profiles/${SUBJECT}/groups/${GROUP}/label`);
    assert.equal(calls[0].method, "POST");
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, expected_version: 4, role_label: "Lab lead", context_group: GROUP });
    assert.equal(calls[0].headers.Authorization, `Session ${SECRET}`);
    assert.equal(calls[0].headers["Content-Type"], "application/json");
  });

  it("sends no selector without one and returns label_cleared for an empty label", async () => {
    reply = json(200, { ...CARD, role_label: null, version: 5 });
    const res = await label({ ...GOOD, role_label: "" });
    assert.deepEqual(at(res), [303, back("label_cleared"), "no-store"]);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, expected_version: 4, role_label: "" });
  });

  it("keeps the edited Group (the path) apart from the selector (the hidden field)", async () => {
    reply = json(200, CARD);
    await label({ ...GOOD, context_group: OTHER_GROUP });
    assert.equal(calls[0].url, `${CORE}/internal/v1/profiles/${SUBJECT}/groups/${GROUP}/label`);
    assert.equal(calls[0].body.context_group, OTHER_GROUP);
  });

  it("refuses an empty, malformed or repeated selector as invalid_context without the selector and without calling core", async () => {
    for (const fields of [
      { ...GOOD, context_group: "" },
      { ...GOOD, context_group: GROUP.toUpperCase() },
      { ...GOOD, context_group: "not-a-uuid" },
      [...Object.entries(GOOD), ["context_group", GROUP], ["context_group", GROUP]],
      [...Object.entries(GOOD), ["context_group", GROUP], ["context_group", OTHER_GROUP]],
    ]) {
      assert.deepEqual(at(await label(fields)), [303, back("invalid_context"), "no-store"], JSON.stringify(fields));
    }
    assert.equal(calls.length, 0);
  });

  it("maps core's answers: stale_version, invalid_label, not_permitted and unavailable keep the selector; invalid_context drops it", async () => {
    for (const [status, body, want, context] of [
      [409, { error: "version_conflict" }, "stale_version", GROUP],
      [400, { error: "invalid_input", field: "role_label" }, "invalid_label", GROUP],
      [400, { error: "invalid_request", field: "role_label" }, "invalid_label", GROUP],
      [403, { error: "forbidden" }, "not_permitted", GROUP],
      [503, { error: "unavailable" }, "unavailable", GROUP],
      [400, { error: "invalid_request", field: "context_group" }, "invalid_context", undefined],
    ]) {
      reply = json(status, body);
      assert.deepEqual(at(await label({ ...GOOD, context_group: GROUP })), [303, back(want, context), "no-store"], want);
    }
  });

  it("answers core's not_found (an invisible subject, an unreadable or other Group, an unselectable selector) with the fixed 404", async () => {
    reply = json(404, { error: "not_found" });
    const res = await label({ ...GOOD, context_group: GROUP });
    assert.equal(res.status, 404);
    assert.equal(res.headers.get("location"), null);
    assert.ok((await res.text()).includes("Not found"));
    reply = json(403, { error: "csrf_failed" });
    const csrf = await label({ ...GOOD, csrf_token: "forged" });
    assert.equal(csrf.status, 403);
    const text = await csrf.text();
    assert.ok(text.includes("Form expired"));
    assert.ok(text.includes("Nothing was changed."));
  });

  it("never reflects the submitted label, token or selector value in the answer", async () => {
    const marker = `<script>alert("x")</script>`;
    for (const [status, body] of [
      [400, { error: "invalid_input", field: "role_label" }],
      [403, { error: "forbidden" }],
      [404, { error: "not_found" }],
      [403, { error: "csrf_failed" }],
    ]) {
      reply = json(status, body);
      const res = await label({ csrf_token: CSRF, expected_version: "4", role_label: marker, context_group: GROUP });
      const location = res.headers.get("location") ?? "";
      const text = res.status === 303 ? "" : await res.text();
      for (const leak of ["script", "alert", CSRF, "Lab lead"]) {
        assert.equal(location.includes(leak), false, `${leak} in ${location}`);
        assert.equal(text.includes(leak), false, `${leak} in the ${res.status} page`);
      }
    }
    // A malformed selector is never echoed either.
    const res = await label({ ...GOOD, context_group: "evil<value>" });
    assert.equal(res.headers.get("location"), back("invalid_context"));
  });

  it("refuses before core: another method, another encoding, a foreign Origin, malformed ids, a malformed form, no session", async () => {
    for (const method of ["GET", "PUT", "PATCH", "DELETE", "OPTIONS"]) {
      const res = labelRoute[method]();
      assert.equal(res.status, 405, method);
    }
    const fields = form(GOOD);
    assert.equal((await label(fields, { "content-type": "text/plain" })).status, 415);
    assert.equal((await label(fields, { origin: "https://attacker.localhost:543" })).status, 403);
    assert.equal((await label(fields, {}, "x")).status, 404);
    assert.equal((await label(fields, {}, SUBJECT, GROUP.toUpperCase())).status, 404);
    assert.equal((await label({ csrf_token: CSRF, expected_version: "4" })).status, 400);
    assert.equal((await label({ csrf_token: CSRF, role_label: "x" })).status, 400);
    assert.equal((await label([...Object.entries(GOOD), ["role_label", "again"]])).status, 400);
    assert.equal((await label(fields, { cookie: "" })).status, 401);
    assert.equal(calls.length, 0);
  });
});

describe("Group member card JSON route", () => {
  const read = (id, query = "", headers) => cardsRoute.GET(get(`/api/profiles/${id}/groups${query}`, headers), { params: Promise.resolve({ id }) });

  it("answers exactly {user_id, cards, narrow_required} with the five card keys, no-store", async () => {
    reply = json(200, { user_id: SUBJECT, cards: [CARD], narrow_required: false });
    const res = await read(SUBJECT);
    assert.equal(res.status, 200);
    assert.equal(res.headers.get("cache-control"), "no-store");
    assert.deepEqual(await res.json(), { user_id: SUBJECT, cards: [CARD], narrow_required: false });
    assert.equal(calls[0].url, `${CORE}/internal/v1/profiles/${SUBJECT}/groups?origin=${encodeURIComponent(ORIGIN)}`);
  });

  it("passes ?context= to core as context_group and answers the overflow signal without a card or a number", async () => {
    reply = json(200, { user_id: SUBJECT, cards: [CARD], narrow_required: false });
    assert.equal((await read(SUBJECT, `?context=${GROUP}`)).status, 200);
    assert.deepEqual(new URL(calls[0].url).searchParams.getAll("context_group"), [GROUP]);
    reply = json(200, { user_id: SUBJECT, cards: [], narrow_required: true });
    assert.deepEqual(await (await read(SUBJECT)).json(), { user_id: SUBJECT, cards: [], narrow_required: true });
  });

  it("is 400 invalid_request for an empty, malformed or repeated ?context= without calling core", async () => {
    for (const query of ["?context=", `?context=${GROUP.toUpperCase()}`, "?context=not-a-uuid", `?context=${GROUP}&context=${GROUP}`, `?context=${GROUP}&context=${OTHER_GROUP}`]) {
      const res = await read(SUBJECT, query);
      assert.deepEqual([res.status, res.headers.get("cache-control")], [400, "no-store"], query);
      assert.deepEqual(await res.json(), { error: "invalid_request" }, query);
    }
    assert.equal(calls.length, 0);
  });

  it("is 401 without a session, 404 for a malformed id or any refusal of the subject or the context, 503 when core cannot answer", async () => {
    assert.deepEqual([(await read(SUBJECT, "", { cookie: "" })).status, calls.length], [401, 0]);
    assert.deepEqual([(await read("not-a-user")).status, calls.length], [404, 0]);
    for (const [status, body, want, wantBody] of [
      [404, { error: "not_found" }, 404, { error: "not_found" }],
      [403, { error: "forbidden" }, 404, { error: "not_found" }],
      [403, { error: "not_available" }, 404, { error: "not_found" }],
      [400, { error: "invalid_request", field: "context_group" }, 400, { error: "invalid_request" }],
      [401, { error: "unauthenticated" }, 401, { error: "unauthenticated" }],
      [503, { error: "unavailable" }, 503, { error: "unavailable" }],
      [200, { user_id: SUBJECT, cards: [{ ...CARD, role: "group_admin" }], narrow_required: false }, 503, { error: "unavailable" }],
      [200, { user_id: SUBJECT, cards: [], narrow_required: false, total: 0 }, 503, { error: "unavailable" }],
    ]) {
      reply = json(status, body);
      const res = await read(SUBJECT, `?context=${GROUP}`);
      assert.equal(res.status, want, JSON.stringify(body));
      assert.equal(res.headers.get("cache-control"), "no-store");
      assert.deepEqual(await res.json(), wantBody, JSON.stringify(body));
    }
    for (const method of ["POST", "PUT", "PATCH", "DELETE", "OPTIONS"]) assert.equal(cardsRoute[method]().status, 405, method);
  });
});
