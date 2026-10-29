// Unit tests for the G020 label forms and routes (Plan r2 §1.5, §4.2, T11): the closed action fields, the form-layer order
// (method, encoding, exact Origin, session) before core, the 303 to each page with a closed status (including
// organization_context_required: nothing deleted), the fixed refusal pages ("Not permitted" on forbidden, the closed
// conflict sentences) and the unknown-state page that never claims success. fetch is a recorder; nothing leaves the
// process. The resolve hook maps lib siblings and the routes' @/lib aliases to .ts, as group-authority-forms.test.mjs does.
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

const lab = await import("../../lib/labels-client.ts");
const { cookieName } = await import("../../lib/session.ts");
const orgRoute = await import("../../app/organization/labels/action/route.ts");
const groupRoute = await import("../../app/groups/[id]/labels/action/route.ts");

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
const G = ID(77);
const ORG1 = { id: ID(1), scope: "organization", group_id: null, name: "Priority", visibility: "owner_scope" };
const GRP1 = { id: ID(3), scope: "group", group_id: G, name: "Shared", visibility: "target_members" };

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "labels_unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), method: init.method, body: init.body === undefined ? undefined : JSON.parse(init.body) });
    return reply();
  });
});

afterEach(() => {
  mock.restoreAll();
  for (const [key, value] of Object.entries(saved)) {
    if (value === undefined) delete process.env[key];
    else process.env[key] = value;
  }
});

const form = (fields) => new URLSearchParams(fields).toString();
function post(path, body, headers = {}) {
  return new Request(`${ORIGIN}${path}`, {
    method: "POST",
    headers: { host: HOST, origin: ORIGIN, "content-type": "application/x-www-form-urlencoded", cookie: `${cookieName("session", "https")}=${SECRET}`, ...headers },
    body,
  });
}
const ORG_PATH = "/organization/labels/action";
const orgPost = (fields, headers) => orgRoute.POST(post(ORG_PATH, typeof fields === "string" ? fields : form(fields), headers));
const groupPost = (fields, headers, id = G) =>
  groupRoute.POST(post(`/groups/${id}/labels/action`, typeof fields === "string" ? fields : form(fields), headers), { params: Promise.resolve({ id }) });

describe("readLabelAction", () => {
  const read = (kind, fields) => lab.readLabelAction(kind, new URLSearchParams(fields));

  it("reads each action's closed fields; an Organization label's visibility is fixed and never submitted", () => {
    assert.deepEqual(read("organization", { csrf_token: CSRF, action: "create", name: "A" }), { action: "create", name: "A", visibility: "owner_scope" });
    assert.deepEqual(read("organization", { action: "rename", label_id: ID(1), name: "B" }), { action: "rename", labelId: ID(1), name: "B", visibility: "owner_scope" });
    assert.deepEqual(read("organization", { action: "delete", label_id: ID(1) }), { action: "delete", labelId: ID(1) });
    assert.deepEqual(read("group", { action: "create", name: "A", visibility: "target_members" }), { action: "create", name: "A", visibility: "target_members" });
    assert.deepEqual(read("group", { action: "visibility", label_id: ID(3), name: "A", visibility: "owner_scope" }), {
      action: "visibility",
      labelId: ID(3),
      name: "A",
      visibility: "owner_scope",
    });
    assert.deepEqual(read("assignment", { action: "assign", label_id: ID(1) }), { action: "assign", labelId: ID(1) });
    assert.deepEqual(read("assignment", { action: "unassign", label_id: ID(1) }), { action: "unassign", labelId: ID(1) });
  });

  it("refuses unknown, repeated, missing or out-of-scope fields and actions", () => {
    for (const [kind, fields] of [
      ["organization", { action: "create", name: "A", visibility: "owner_scope" }],
      ["organization", { action: "visibility", label_id: ID(1), name: "A", visibility: "owner_scope" }],
      ["organization", { action: "assign", label_id: ID(1) }],
      ["organization", { action: "create" }],
      ["organization", { action: "create", name: "A", extra: "1" }],
      ["organization", [["action", "create"], ["name", "A"], ["name", "B"]]],
      ["organization", { action: "delete", label_id: "1" }],
      ["group", { action: "create", name: "A" }],
      ["group", { action: "create", name: "A", visibility: "everyone" }],
      ["group", { action: "rename", label_id: ID(3), name: "A" }],
      ["group", { action: "assign", label_id: ID(1) }],
      ["assignment", { action: "assign", label_id: ID(1), group_id: G }],
      ["assignment", { action: "create", name: "A" }],
      ["assignment", {}],
      ["assignment", { action: "toString", label_id: ID(1) }],
    ]) {
      assert.equal(read(kind, fields), undefined, `${kind} ${JSON.stringify(fields)}`);
    }
  });

  it("has a page message for each closed status, including the Organization-wide context refusal", () => {
    assert.deepEqual(Object.keys(lab.LABEL_STATUS).sort(), ["created", "deleted", "organization_context_required", "renamed", "visibility_changed"]);
    assert.equal(lab.LABEL_STATUS.organization_context_required, "Deleting an Organization label requires the Organization-wide context");
    assert.deepEqual(Object.keys(lab.GROUP_LABEL_STATUS).sort(), ["assigned", "unassigned"]);
    for (const f of ["invalid_name", "name_conflict", "scope_full", "already_assigned", "not_assigned", "visibility_not_allowed"]) {
      assert.equal(typeof lab.labelRefusalSentence(f), "string", f);
    }
    assert.equal(lab.labelRefusalSentence("toString"), undefined);
  });
});

describe("Organization labels action route", () => {
  it("creates with owner_scope and answers 303 to the page with the created status", async () => {
    reply = () => Response.json(ORG1);
    const res = await orgPost({ csrf_token: CSRF, action: "create", name: "Priority" });
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "/organization/labels?status=created");
    assert.deepEqual(calls, [
      { url: `${CORE}/internal/v1/labels`, method: "POST", body: { origin: ORIGIN, csrf_token: CSRF, scope: "organization", name: "Priority", visibility: "owner_scope" } },
    ]);
  });

  it("renames and deletes, each with its closed status", async () => {
    reply = () => Response.json({ ...ORG1, name: "Urgent" });
    const renamed = await orgPost({ csrf_token: CSRF, action: "rename", label_id: ID(1), name: "Urgent" });
    assert.equal(renamed.headers.get("location"), "/organization/labels?status=renamed");
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, name: "Urgent", visibility: "owner_scope" });
    reply = () => Response.json({ deleted: true, assignments_removed: 3 });
    const deleted = await orgPost({ csrf_token: CSRF, action: "delete", label_id: ID(1) });
    assert.equal(deleted.headers.get("location"), "/organization/labels?status=deleted");
    assert.deepEqual(calls[1], { url: `${CORE}/internal/v1/labels/${ID(1)}/delete`, method: "POST", body: { origin: ORIGIN, csrf_token: CSRF } });
  });

  // Round 15: core keeps a boundary U+FEFF (not Unicode White_Space), so its create and rename answers carry it; the
  // route must answer the closed status, not the unknown-outcome page. The name is sent as submitted: web never trims.
  it("answers the closed status when core's stored name begins or ends with U+FEFF", async () => {
    const bom = String.fromCodePoint(0xfeff);
    const name = `${bom}Urgent${bom}`;
    reply = () => Response.json({ ...ORG1, name });
    const created = await orgPost({ csrf_token: CSRF, action: "create", name });
    assert.equal(created.status, 303);
    assert.equal(created.headers.get("location"), "/organization/labels?status=created");
    const renamed = await orgPost({ csrf_token: CSRF, action: "rename", label_id: ID(1), name });
    assert.equal(renamed.status, 303);
    assert.equal(renamed.headers.get("location"), "/organization/labels?status=renamed");
    assert.deepEqual(calls.map((c) => c.body.name), [name, name]);
  });

  it("shows core's Organization-wide context refusal as the closed status: nothing deleted, never a success status", async () => {
    reply = () => Response.json({ error: "invalid_request", field: "organization_context_required" }, { status: 400 });
    const res = await orgPost({ csrf_token: CSRF, action: "delete", label_id: ID(1) });
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "/organization/labels?status=organization_context_required");
  });

  it("shows a refusal as a fixed page: Not permitted on forbidden, the closed conflict sentence, nothing changed", async () => {
    reply = () => Response.json({ error: "forbidden" }, { status: 403 });
    const forbidden = await orgPost({ csrf_token: CSRF, action: "create", name: "Priority" });
    assert.equal(forbidden.status, 403);
    assert.equal(forbidden.headers.get("location"), null);
    assert.equal(forbidden.headers.get("cache-control"), "no-store");
    assert.match(await forbidden.text(), /Not permitted.*Nothing was changed/s);
    reply = () => Response.json({ error: "conflict", field: "name_conflict" }, { status: 409 });
    const conflict = await orgPost({ csrf_token: CSRF, action: "create", name: "Priority" });
    assert.equal(conflict.status, 409);
    assert.match(await conflict.text(), /A label with this name already exists here.*Nothing was changed/s);
    reply = () => Response.json({ error: "not_found" }, { status: 404 });
    assert.equal((await orgPost({ csrf_token: CSRF, action: "delete", label_id: ID(9) })).status, 404);
  });

  it("shows an outcome core may have committed as unknown, never as success", async () => {
    for (const r of [() => Response.json({ error: "labels_unavailable" }, { status: 503 }), () => new Response("{", { status: 200 }), () => Response.json({ ...ORG1, visibility: "target_members" })]) {
      reply = r;
      const res = await orgPost({ csrf_token: CSRF, action: "create", name: "Priority" });
      assert.equal(res.status, 503);
      assert.match(await res.text(), /may or may not have completed/);
    }
  });

  it("refuses before core when the form, origin or session is wrong, and answers 405 to other methods", async () => {
    for (const [body, headers, status] of [
      [form({ csrf_token: CSRF, action: "create", name: "A", visibility: "target_members" }), {}, 400],
      [form({ csrf_token: CSRF, action: "assign", label_id: ID(1) }), {}, 400],
      [form({ action: "create", name: "A" }), {}, 400],
      [form({ csrf_token: "", action: "create", name: "A" }), {}, 400],
      [`${form({ csrf_token: CSRF, action: "create", name: "A" })}&csrf_token=${CSRF}`, {}, 400],
      [form({ csrf_token: CSRF, action: "create", name: "A" }), { origin: "https://evil.test" }, 403],
      [form({ csrf_token: CSRF, action: "create", name: "A" }), { "content-type": "application/json" }, 415],
      [form({ csrf_token: CSRF, action: "create", name: "A" }), { cookie: "" }, 401],
    ]) {
      const res = await orgPost(body, headers);
      assert.equal(res.status, status, `${body} ${JSON.stringify(headers)}`);
    }
    assert.equal(calls.length, 0);
    for (const m of ["GET", "PUT", "PATCH", "DELETE", "OPTIONS"]) assert.equal(orgRoute[m]().status, 405, m);
  });
});

describe("Group labels action route", () => {
  it("manages the Group's own definitions and answers on the Group labels page", async () => {
    reply = () => Response.json(GRP1);
    const created = await groupPost({ csrf_token: CSRF, action: "create", name: "Shared", visibility: "target_members" });
    assert.equal(created.status, 303);
    assert.equal(created.headers.get("location"), `/groups/${G}/labels?status=created`);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, scope: "group", group_id: G, name: "Shared", visibility: "target_members" });
    reply = () => Response.json({ ...GRP1, visibility: "owner_scope" });
    const narrowed = await groupPost({ csrf_token: CSRF, action: "visibility", label_id: ID(3), name: "Shared", visibility: "owner_scope" });
    assert.equal(narrowed.headers.get("location"), `/groups/${G}/labels?status=visibility_changed`);
    assert.deepEqual(calls[1], { url: `${CORE}/internal/v1/labels/${ID(3)}/update`, method: "POST", body: { origin: ORIGIN, csrf_token: CSRF, name: "Shared", visibility: "owner_scope" } });
    reply = () => Response.json({ deleted: true, assignments_removed: 0 });
    assert.equal((await groupPost({ csrf_token: CSRF, action: "delete", label_id: ID(3) })).headers.get("location"), `/groups/${G}/labels?status=deleted`);
  });

  it("assigns and unassigns Organization labels and answers on the Group page", async () => {
    reply = () => Response.json({ assigned: true });
    const assigned = await groupPost({ csrf_token: CSRF, action: "assign", label_id: ID(1) });
    assert.equal(assigned.status, 303);
    assert.equal(assigned.headers.get("location"), `/groups/${G}?status=assigned`);
    assert.deepEqual(calls[0], { url: `${CORE}/internal/v1/groups/${G}/labels/assign`, method: "POST", body: { origin: ORIGIN, csrf_token: CSRF, label_id: ID(1) } });
    reply = () => Response.json({ assigned: false });
    assert.equal((await groupPost({ csrf_token: CSRF, action: "unassign", label_id: ID(1) })).headers.get("location"), `/groups/${G}?status=unassigned`);
    reply = () => Response.json({ error: "conflict", field: "not_assigned" }, { status: 409 });
    const again = await groupPost({ csrf_token: CSRF, action: "unassign", label_id: ID(1) });
    assert.equal(again.status, 409);
    const text = await again.text();
    assert.match(text, /This label is not assigned to this Group/);
    assert.match(text, new RegExp(`href="/groups/${G}"`));
  });

  it("shows a forged write as Not permitted (e.g. Organization staff without a Group role creating a Group label)", async () => {
    reply = () => Response.json({ error: "forbidden" }, { status: 403 });
    const res = await groupPost({ csrf_token: CSRF, action: "create", name: "Forged", visibility: "owner_scope" });
    assert.equal(res.status, 403);
    assert.match(await res.text(), /Not permitted/);
  });

  it("refuses a non-canonical Group id with the fixed not-found page and malformed forms before core", async () => {
    const res = await groupPost({ csrf_token: CSRF, action: "assign", label_id: ID(1) }, {}, "research");
    assert.equal(res.status, 404);
    for (const fields of [
      { csrf_token: CSRF, action: "assign", label_id: ID(1), name: "x" },
      { csrf_token: CSRF, action: "create", name: "A" },
      { csrf_token: CSRF, action: "rename", label_id: ID(3), name: "A" },
      { csrf_token: CSRF, action: "unassign" },
    ]) {
      assert.equal((await groupPost(fields)).status, 400, JSON.stringify(fields));
    }
    assert.equal(calls.length, 0);
    for (const m of ["GET", "PUT", "PATCH", "DELETE", "OPTIONS"]) assert.equal(groupRoute[m]().status, 405, m);
  });
});
