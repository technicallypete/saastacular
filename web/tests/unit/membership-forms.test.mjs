// Unit tests for the membership form helpers and routes (G019 Plan r1 §1.6): the closed action field and its shape per
// scope, the fixed refusal sentences, and the three form routes — the form-layer order, invite's direct no-store result
// page (the one-time token on that page only, never in a Location header or the title), the 303 statuses of the other
// actions, the accept route's body-only token and its single not_found page. fetch is a recorder; nothing leaves the
// process. The resolve hook maps lib siblings and the routes' @/lib aliases to .ts, as onboarding-forms.test.mjs does.
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

const mem = await import("../../lib/membership-client.ts");
const { cookieName } = await import("../../lib/session.ts");
const orgRoute = await import("../../app/organization/members/action/route.ts");
const groupRoute = await import("../../app/groups/[id]/members/action/route.ts");
const acceptRoute = await import("../../app/invitations/[id]/accept/route.ts");

const CORE = "http://core.internal:8180";
const HOST = "org-a.localhost:543";
const ORIGIN = `https://${HOST}`;
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const TOKEN = "T".repeat(43);
const ENV = {
  PRODUCT_NAME: "Test Product",
  PLATFORM_BASE_DOMAIN: "example.test",
  CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test",
  CORE_INTERNAL_URL: CORE,
  IDENTITY_COOKIE_PROFILE: "https",
};
const ID = (n) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
const INV = {
  id: ID(10),
  scope: "organization",
  group_id: null,
  role: "org_member",
  issuer_display_name: null,
  intended_user_id: null,
  expires_at: "2026-10-25T12:00:00Z",
  status: "pending",
  consumed_by: null,
  intended_address: "new@example.test",
};

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "membership_unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), body: init.body === undefined ? undefined : JSON.parse(init.body) });
    return reply(String(url));
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
const params = (id) => ({ params: Promise.resolve({ id }) });
function post(path, body, headers = {}) {
  return new Request(`${ORIGIN}${path}`, {
    method: "POST",
    headers: { host: HOST, origin: ORIGIN, "content-type": "application/x-www-form-urlencoded", cookie: `${cookieName("session", "https")}=${SECRET}`, ...headers },
    body,
  });
}
const getter = (fields) => (name) => {
  const all = new URLSearchParams(fields).getAll(name);
  return all.length > 1 ? Symbol("ambiguous") : all[0];
};

describe("readMemberAction", () => {
  it("accepts each closed action with its exact shape", () => {
    assert.deepEqual(mem.readMemberAction("organization", getter({ action: "invite", role: "org_admin", intended_address: "a@b.test", expires_in_hours: "24" })), {
      action: "invite",
      role: "org_admin",
      recipient: { address: "a@b.test" },
      expiresInHours: 24,
    });
    assert.deepEqual(mem.readMemberAction("group", getter({ action: "invite", role: "group_member", intended_user_id: ID(5), expires_in_hours: "720" })), {
      action: "invite",
      role: "group_member",
      recipient: { userId: ID(5) },
      expiresInHours: 720,
    });
    assert.deepEqual(mem.readMemberAction("organization", getter({ action: "revoke", invitation_id: ID(10) })), { action: "revoke", invitationId: ID(10) });
    assert.deepEqual(mem.readMemberAction("organization", getter({ action: "change_role", user_id: ID(1), role: "org_member" })), { action: "change_role", userId: ID(1), role: "org_member" });
    assert.deepEqual(mem.readMemberAction("group", getter({ action: "demote", user_id: ID(1), role: "group_admin" })), { action: "demote", userId: ID(1), role: "group_admin" });
    assert.deepEqual(mem.readMemberAction("organization", getter({ action: "remove", user_id: ID(1) })), { action: "remove", userId: ID(1) });
    assert.deepEqual(mem.readMemberAction("organization", getter({ action: "promote", user_id: ID(1) })), { action: "promote", userId: ID(1) });
    assert.deepEqual(mem.readMemberAction("organization", getter({ action: "transfer", user_id: ID(1), former_role: "none" })), { action: "transfer", recipientUserId: ID(1), formerRole: "none" });
    assert.deepEqual(mem.readMemberAction("organization", getter({ action: "leave" })), { action: "leave" });
  });

  it("refuses unknown actions, other-scope or owner roles, open lifetimes, malformed ids and repeated fields", () => {
    for (const [kind, fields] of [
      ["organization", { action: "retire" }],
      ["organization", { action: "change_role", user_id: ID(1), role: "org_owner" }],
      ["organization", { action: "change_role", user_id: ID(1), role: "group_admin" }],
      ["organization", { action: "invite", role: "org_member", intended_address: "a@b.test", expires_in_hours: "5" }],
      ["organization", { action: "invite", role: "org_member", intended_address: "", expires_in_hours: "24" }],
      ["group", { action: "invite", role: "group_member", intended_address: "a@b.test", expires_in_hours: "24" }],
      ["organization", { action: "remove", user_id: "not-an-id" }],
      ["organization", { action: "transfer", user_id: ID(1), former_role: "org_owner" }],
      ["organization", [["action", "leave"], ["action", "leave"]]],
    ]) {
      assert.equal(mem.readMemberAction(kind, getter(fields)), undefined, JSON.stringify(fields));
    }
  });

  it("has a status for every non-invite action and a sentence for the closed conflict fields", () => {
    for (const a of mem.MEMBER_ACTIONS.filter((x) => x !== "invite")) assert.ok(Object.hasOwn(mem.ACTION_STATUS, mem.STATUS_OF[a]), a);
    for (const f of ["last_owner", "recipient_ineligible", "invitation_limit", "context", "dependent_bindings", "owner_subject"]) assert.equal(typeof mem.refusalSentence(f), "string", f);
    assert.equal(mem.refusalSentence("unknown_field"), undefined);
  });

  it("labels every role and never the raw code", () => {
    for (const kind of ["organization", "group"]) for (const role of Object.values(mem.ROLES[kind])) assert.ok(!mem.ROLE_LABELS[role].includes("_"), role);
  });
});

describe("member action routes", () => {
  it("answers invite with a direct no-store page carrying the token once and no Location; the link follows the invitation's scope", async () => {
    reply = () => Response.json({ invitation: INV, token: TOKEN }, { status: 201 });
    const res = await orgRoute.POST(post("/organization/members/action", form({ csrf_token: CSRF, action: "invite", role: "org_member", intended_address: "new@example.test", expires_in_hours: "168" })));
    assert.equal(res.status, 200);
    assert.equal(res.headers.get("cache-control"), "no-store");
    assert.equal(res.headers.get("location"), null);
    const html = await res.text();
    assert.equal(html.split(TOKEN).length - 1, 1);
    // G058 scope amendment04: an Organization invitation's link opens the enrollment entry (selected by the invitation's
    // scope), with the token in the fragment only; the obsolete "not available yet" copy is gone.
    assert.ok(html.includes(`${ORIGIN}/join/invitations/${ID(10)}#${TOKEN}`));
    assert.ok(!html.includes(`${ORIGIN}/invitations/${ID(10)}`));
    assert.ok(!html.includes("not available yet"));
    assert.ok(html.includes("signs in (or creates an account with the invited address) and joins this workspace"));
    assert.ok(!/<title>[^<]*T{43}/.test(html));
    assert.equal(calls[0].url, `${CORE}/internal/v1/invitations`);
    // A Group invitation keeps the existing member acceptance page /invitations/<id>, the same single fragment token,
    // no Location and no token in the title.
    calls = [];
    const group = { ...INV, id: ID(12), scope: "group", group_id: ID(77), role: "group_member", intended_user_id: ID(5) };
    delete group.intended_address;
    reply = () => Response.json({ invitation: group, token: TOKEN }, { status: 201 });
    const g = await groupRoute.POST(post(`/groups/${ID(77)}/members/action`, form({ csrf_token: CSRF, action: "invite", role: "group_member", intended_user_id: ID(5), expires_in_hours: "168" })), params(ID(77)));
    assert.equal(g.status, 200);
    assert.equal(g.headers.get("cache-control"), "no-store");
    assert.equal(g.headers.get("location"), null);
    const ghtml = await g.text();
    assert.equal(ghtml.split(TOKEN).length - 1, 1);
    assert.ok(ghtml.includes(`${ORIGIN}/invitations/${ID(12)}#${TOKEN}`));
    assert.ok(!ghtml.includes("/join/invitations/"));
    assert.ok(!ghtml.includes("joins this workspace"));
    assert.ok(!/<title>[^<]*T{43}/.test(ghtml));
    assert.equal(calls[0].url, `${CORE}/internal/v1/invitations`);
    assert.equal(calls[0].body.scope, "group");
  });

  it("answers the other actions with a 303 to the members page and a closed status", async () => {
    reply = () => Response.json({}, { status: 200 });
    const res = await groupRoute.POST(post(`/groups/${ID(77)}/members/action`, form({ csrf_token: CSRF, action: "leave" })), params(ID(77)));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), `/groups/${ID(77)}/members?status=left`);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, scope: "group", group_id: ID(77) });
  });

  it("posts a Transfer to a recipient chosen on the second page of eligible admins and leaves the decision to core", async () => {
    // The 101st eligible admin is only offered on the owners_after page; its id posts like any other and core decides.
    reply = () => Response.json({}, { status: 200 });
    const res = await orgRoute.POST(post("/organization/members/action", form({ csrf_token: CSRF, action: "transfer", user_id: ID(101), former_role: "org_admin" })));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "/organization/members?status=transferred");
    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, `${CORE}/internal/v1/ownership/transfer`);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, scope: "organization", recipient_user_id: ID(101), former_role: "org_admin" });
    // Core's refusal of a recipient that is no longer eligible is shown as the closed conflict, not as success.
    calls = [];
    reply = () => Response.json({ error: "conflict", field: "recipient_ineligible" }, { status: 409 });
    const refused = await groupRoute.POST(post(`/groups/${ID(77)}/members/action`, form({ csrf_token: CSRF, action: "transfer", user_id: ID(101), former_role: "none" })), params(ID(77)));
    assert.notEqual(refused.headers.get("location"), `/groups/${ID(77)}/members?status=transferred`);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, scope: "group", group_id: ID(77), recipient_user_id: ID(101), former_role: "none" });
  });

  it("shows a conflict as an unchanged state, an unknown outcome as unknown, and refuses before core when the form is wrong", async () => {
    reply = () => Response.json({ error: "conflict", field: "last_owner" }, { status: 409 });
    const conflict = await orgRoute.POST(post("/organization/members/action", form({ csrf_token: CSRF, action: "leave" })));
    assert.equal(conflict.status, 409);
    assert.ok((await conflict.text()).includes("Nothing was changed"));
    reply = () => new Response("{", { status: 200 });
    assert.equal((await orgRoute.POST(post("/organization/members/action", form({ csrf_token: CSRF, action: "leave" })))).status, 503);
    const before = calls.length;
    assert.equal((await orgRoute.POST(post("/organization/members/action", form({ csrf_token: CSRF, action: "retire" })))).status, 400);
    assert.equal((await orgRoute.POST(post("/organization/members/action", form({ csrf_token: CSRF, action: "leave" }), { origin: "https://evil.test" }))).status, 403);
    assert.equal((await orgRoute.POST(post("/organization/members/action", form({ csrf_token: CSRF, action: "leave" }), { cookie: "" }))).status, 401);
    assert.equal((await groupRoute.POST(post("/groups/x/members/action", form({ csrf_token: CSRF, action: "leave" })), params("x"))).status, 404);
    assert.equal(calls.length, before);
    assert.equal(orgRoute.GET().status, 405);
  });
});

describe("accept route", () => {
  it("forwards the token in the body only and redirects to the Group with the outcome", async () => {
    reply = () => Response.json({ outcome: "joined", group_id: ID(77) }, { status: 200 });
    const res = await acceptRoute.POST(post(`/invitations/${ID(11)}/accept`, form({ csrf_token: CSRF, token: TOKEN })), params(ID(11)));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), `/groups/${ID(77)}?status=joined`);
    assert.ok(!calls[0].url.includes(TOKEN));
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, token: TOKEN });
  });

  it("answers core's single not_found with one fixed page", async () => {
    reply = () => Response.json({ error: "not_found" }, { status: 404 });
    const res = await acceptRoute.POST(post(`/invitations/${ID(11)}/accept`, form({ csrf_token: CSRF, token: TOKEN })), params(ID(11)));
    assert.equal(res.status, 404);
    assert.ok(!(await res.text()).includes(TOKEN));
  });
});
