// Unit tests for lib/membership-client.ts (G019 Plan r1 §5.3–§5.4, r2 addendum §7): the exact outbound requests of the
// sixteen calls (diagnostic headers, the session credential, JSON with the documented fields only, the token in the body
// only), the strict reading of core's answers (exact keys, UUIDs, scope-consistent roles, ascending pages, the last-owner
// hint rule, per-endpoint byte bounds, the eligibility count without ids) and the closed refusal reading. fetch is
// replaced by a recorder, so nothing leaves the process. A resolve hook scoped to the lib directory maps the extensionless
// sibling imports to .ts, as onboarding-client.test.mjs does.
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

const mem = await import("../../lib/membership-client.ts");
const { deriveContext, outboundHeaders } = await import("../../lib/request-context.ts");

const CORE = "http://core.internal:8180";
const ORIGIN = "https://org-a.localhost:543";
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const TOKEN = "T".repeat(43);
const M = { origin: ORIGIN, csrfToken: CSRF };
const CTX = deriveContext(new Headers({ "x-request-id": "req-membership-1", traceparent: `00-${"a".repeat(32)}-${"b".repeat(16)}-01` }));
const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };
const ORG = { kind: "organization" };
const ID = (n) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
const GROUP = { kind: "group", groupId: ID(77) };
const flags = { self: false, disabled: false, can_change_role: true, can_remove: true, can_promote: false, can_demote: false, can_transfer: false };
const M1 = { user_id: ID(1), display_name: "Ada <Admin>", role: "org_admin", ...flags };
const M2 = { user_id: ID(2), display_name: null, role: "org_member", ...flags };
const SELF = { user_id: ID(3), display_name: "Me", role: "org_owner", self: true, can_leave: true, last_owner: false, can_manage: true };
const INV = {
  id: ID(10),
  scope: "organization",
  group_id: null,
  role: "org_member",
  issuer_display_name: "Ada",
  intended_user_id: null,
  expires_at: "2026-10-25T12:00:00Z",
  status: "pending",
  consumed_by: null,
  intended_address: "new@example.test",
};
const GINV = { ...INV, id: ID(11), scope: "group", group_id: ID(77), role: "group_member", intended_user_id: ID(5), intended_address: undefined };
delete GINV.intended_address;

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "membership_unavailable" }, { status: 503 });
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

const answer = (status, body) => () => new Response(typeof body === "string" ? body : JSON.stringify(body), { status, headers: { "content-type": "application/json" } });

function sent(i = 0) {
  const c = calls[i];
  return { url: c.url, method: c.init.method, headers: c.init.headers, body: c.init.body === undefined ? undefined : JSON.parse(c.init.body) };
}

describe("membership client: requests", () => {
  it("lists Organization members with the exact query, headers and no body", async () => {
    reply = answer(200, { members: [M1, M2], next: null });
    const r = await mem.listMembers(CTX, SECRET, ORIGIN, ORG, { after: ID(0) });
    assert.equal(r.ok, true);
    assert.deepEqual(r.value.members.map((m) => [m.userId, m.displayName, m.role]), [[ID(1), "Ada <Admin>", "org_admin"], [ID(2), null, "org_member"]]);
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/members?origin=${encodeURIComponent(ORIGIN)}&limit=50&after=${ID(0)}&scope=organization`);
    assert.equal(s.method, "GET");
    assert.deepEqual(s.headers, { ...outboundHeaders(CTX), Authorization: `Session ${SECRET}` });
    assert.equal(s.body, undefined);
  });

  it("asks for owner candidates and Group scope explicitly", async () => {
    reply = answer(200, { members: [], next: null });
    await mem.listMembers(CTX, SECRET, ORIGIN, GROUP, { ownerCandidates: true, limit: 100 });
    assert.equal(sent().url, `${CORE}/internal/v1/members?origin=${encodeURIComponent(ORIGIN)}&limit=100&scope=group&group_id=${ID(77)}&role=owner_candidates`);
  });

  it("pages 101 eligible admins: a bounded first page, core's cursor in the next link, the 101st on the second page", async () => {
    const admin = (n) => ({ ...M1, user_id: ID(n), display_name: `Admin ${n}`, can_promote: false, can_transfer: true });
    const all = Array.from({ length: 101 }, (_, i) => admin(i + 1));
    // core answers one bounded page per request; the client never asks for more than one page at a time.
    reply = (url) => {
      const after = new URL(url).searchParams.get("after");
      const start = after === null ? 0 : all.findIndex((m) => m.user_id === after) + 1;
      const page = all.slice(start, start + mem.OWNER_CANDIDATE_PAGE);
      const next = start + mem.OWNER_CANDIDATE_PAGE < all.length ? page[page.length - 1].user_id : null;
      return Response.json({ members: page, next });
    };
    assert.equal(mem.OWNER_CANDIDATE_PAGE, mem.MAX_PAGE);
    const first = await mem.listMembers(CTX, SECRET, ORIGIN, ORG, { ownerCandidates: true, limit: mem.OWNER_CANDIDATE_PAGE });
    assert.equal(first.ok, true);
    assert.equal(first.value.members.length, 100);
    assert.equal(first.value.next, ID(100));
    assert.equal(calls.length, 1);
    assert.equal(sent().url, `${CORE}/internal/v1/members?origin=${encodeURIComponent(ORIGIN)}&limit=100&scope=organization&role=owner_candidates`);
    // The next-page link carries only core's cursor and returns to the Transfer form.
    const link = mem.ownerCandidatesNext("/organization/members", first.value.next);
    assert.equal(link, `/organization/members?${mem.OWNER_CANDIDATE_CURSOR}=${ID(100)}#transfer-title`);
    assert.equal(mem.ownerCandidatesNext("/organization/members", null), undefined);
    assert.equal(mem.ownerCandidatesNext("/organization/members", "not-a-cursor"), undefined);
    // The members view forwards that cursor through listMembers; the second page holds the 101st admin and ends.
    const cursor = new URL(link, ORIGIN).searchParams.get(mem.OWNER_CANDIDATE_CURSOR);
    const second = await mem.listMembers(CTX, SECRET, ORIGIN, ORG, { ownerCandidates: true, limit: mem.OWNER_CANDIDATE_PAGE, after: cursor });
    assert.equal(sent(1).url, `${CORE}/internal/v1/members?origin=${encodeURIComponent(ORIGIN)}&limit=100&after=${ID(100)}&scope=organization&role=owner_candidates`);
    assert.deepEqual(second.value.members.map((m) => [m.userId, m.canTransfer]), [[ID(101), true]]);
    assert.equal(second.value.next, null);
    assert.equal(calls.length, 2);
  });

  it("refuses a malformed cursor, limit or Group id without calling core", async () => {
    for (const [scope, page] of [[ORG, { after: "x" }], [ORG, { limit: 0 }], [ORG, { limit: 101 }], [{ kind: "group", groupId: "nope" }, {}]]) {
      assert.deepEqual(await mem.listMembers(CTX, SECRET, ORIGIN, scope, page), { ok: false, error: "invalid_request" });
    }
    assert.equal(calls.length, 0);
  });

  it("issues an Organization invitation by address and a Group one by user id with documented fields only", async () => {
    reply = answer(201, { invitation: INV, token: TOKEN });
    const r = await mem.issueInvitation(CTX, SECRET, M, ORG, { role: "org_member", recipient: { address: "new@example.test" }, expiresInHours: 168 });
    assert.equal(r.ok, true);
    assert.equal(r.value.token, TOKEN);
    assert.deepEqual(sent().body, { origin: ORIGIN, csrf_token: CSRF, scope: "organization", role: "org_member", intended_address: "new@example.test", expires_in_hours: 168 });
    assert.equal(sent().headers["Content-Type"], "application/json");
    reply = answer(201, { invitation: GINV, token: TOKEN });
    await mem.issueInvitation(CTX, SECRET, M, GROUP, { role: "group_member", recipient: { userId: ID(5) } });
    assert.deepEqual(sent(1).body, { origin: ORIGIN, csrf_token: CSRF, scope: "group", group_id: ID(77), role: "group_member", intended_user_id: ID(5) });
  });

  it("sends the acceptance token in the JSON body only and refuses a malformed token locally", async () => {
    reply = answer(200, { outcome: "joined", group_id: ID(77) });
    const r = await mem.acceptInvitation(CTX, SECRET, M, ID(11), TOKEN);
    assert.deepEqual(r, { ok: true, value: { outcome: "joined", groupId: ID(77) } });
    assert.equal(sent().url, `${CORE}/internal/v1/invitations/${ID(11)}/accept`);
    assert.ok(!sent().url.includes(TOKEN));
    assert.deepEqual(sent().body, { origin: ORIGIN, csrf_token: CSRF, token: TOKEN });
    assert.deepEqual(await mem.acceptInvitation(CTX, SECRET, M, ID(11), "short"), { ok: false, error: "invalid_request", field: "invalid_token" });
    assert.deepEqual(await mem.acceptInvitation(CTX, SECRET, M, "x", TOKEN), { ok: false, error: "not_found" });
    assert.equal(calls.length, 1);
  });

  it("maps each member mutation to its path and body", async () => {
    reply = answer(200, { member: M1 });
    await mem.changeRole(CTX, SECRET, M, ORG, ID(1), "org_admin");
    await mem.promote(CTX, SECRET, M, ORG, ID(1));
    await mem.demote(CTX, SECRET, M, GROUP, ID(1), "group_member"); // answer of the other scope: malformed, path still pinned
    reply = answer(200, {});
    await mem.removeMember(CTX, SECRET, M, ORG, ID(1));
    await mem.transfer(CTX, SECRET, M, ORG, ID(1), "none");
    await mem.leave(CTX, SECRET, M, GROUP);
    await mem.revokeInvitation(CTX, SECRET, M, ID(10));
    const got = calls.map((c, i) => [c.url.slice(CORE.length), sent(i).body]);
    const base = { origin: ORIGIN, csrf_token: CSRF };
    assert.deepEqual(got, [
      [`/internal/v1/members/${ID(1)}/role`, { ...base, scope: "organization", role: "org_admin" }],
      [`/internal/v1/members/${ID(1)}/promote`, { ...base, scope: "organization" }],
      [`/internal/v1/members/${ID(1)}/demote`, { ...base, scope: "group", group_id: ID(77), role: "group_member" }],
      [`/internal/v1/members/${ID(1)}/remove`, { ...base, scope: "organization" }],
      ["/internal/v1/ownership/transfer", { ...base, scope: "organization", recipient_user_id: ID(1), former_role: "none" }],
      ["/internal/v1/members/leave", { ...base, scope: "group", group_id: ID(77) }],
      [`/internal/v1/invitations/${ID(10)}/revoke`, base],
    ]);
  });
});

describe("membership client: strict answers", () => {
  it("refuses unsorted pages, extra keys, a role of the other scope and over-long names", async () => {
    for (const body of [
      { members: [M2, M1], next: null },
      { members: [{ ...M1, extra: 1 }], next: null },
      { members: [{ ...M1, role: "group_admin" }], next: null },
      { members: [{ ...M1, display_name: "x".repeat(201) }], next: null },
      { members: [M1], next: ID(0) },
      { members: [M1] },
    ]) {
      reply = answer(200, body);
      assert.deepEqual(await mem.listMembers(CTX, SECRET, ORIGIN, ORG), { ok: false, error: "malformed" });
    }
  });

  it("refuses an answer larger than the endpoint bound", async () => {
    reply = answer(200, `{"members":[],"next":null,"pad":"${"x".repeat(mem.MEMBER_LIST_MAX_BYTES)}"}`);
    assert.deepEqual(await mem.listMembers(CTX, SECRET, ORIGIN, ORG), { ok: false, error: "malformed" });
  });

  it("reads the self membership and refuses a last owner who could leave", async () => {
    reply = answer(200, { member: SELF });
    assert.deepEqual(await mem.selfMembership(CTX, SECRET, ORIGIN, ORG), {
      ok: true,
      value: { userId: ID(3), displayName: "Me", role: "org_owner", canLeave: true, lastOwner: false, canManage: true },
    });
    reply = answer(200, { member: { ...SELF, last_owner: true } });
    assert.deepEqual(await mem.selfMembership(CTX, SECRET, ORIGIN, ORG), { ok: false, error: "malformed" });
    reply = answer(200, { member: { ...SELF, role: "org_member", can_leave: false, last_owner: true } });
    assert.deepEqual(await mem.selfMembership(CTX, SECRET, ORIGIN, ORG), { ok: false, error: "malformed" });
  });

  it("keeps the address only where core sent it and refuses inconsistent invitations", async () => {
    reply = answer(200, { invitations: [INV], next: null });
    const r = await mem.listInvitations(CTX, SECRET, ORIGIN, ORG);
    assert.equal(r.value.invitations[0].intendedAddress, "new@example.test");
    for (const bad of [
      { ...INV, group_id: ID(77) },
      { ...INV, status: "consumed" },
      { ...INV, expires_at: "2026-10-25 12:00" },
      { ...INV, intended_user_id: ID(5) },
      { ...INV, role: "org_owner", scope: "group", group_id: ID(77) },
    ]) {
      reply = answer(200, { invitations: [bad], next: null });
      assert.deepEqual(await mem.listInvitations(CTX, SECRET, ORIGIN, ORG), { ok: false, error: "malformed" });
    }
    // The recipient's own list never carries an address and holds Group invitations only.
    reply = answer(200, { invitations: [INV], next: null });
    assert.deepEqual(await mem.myInvitations(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    reply = answer(200, { invitations: [GINV], next: null });
    assert.equal((await mem.myInvitations(CTX, SECRET, ORIGIN)).ok, true);
  });

  it("accepts both recipient forms only on a consumed Organization invitation whose intended user is its consumer", async () => {
    // G058 scope amendment05: core records the consumer of an address invitation as intended_user_id.
    const CONSUMED = { ...INV, status: "consumed", intended_user_id: ID(5), consumed_by: ID(5) };
    reply = answer(200, { invitations: [CONSUMED], next: null });
    const r = await mem.listInvitations(CTX, SECRET, ORIGIN, ORG);
    assert.equal(r.ok, true);
    assert.deepEqual(
      [r.value.invitations[0].status, r.value.invitations[0].intendedAddress, r.value.invitations[0].intendedUserId, r.value.invitations[0].consumedBy],
      ["consumed", "new@example.test", ID(5), ID(5)],
    );
    for (const bad of [
      { ...CONSUMED, scope: "group", group_id: ID(77), role: "group_member" },
      { ...CONSUMED, status: "pending", consumed_by: null },
      { ...CONSUMED, status: "revoked", consumed_by: null },
      { ...CONSUMED, status: "expired", consumed_by: null },
      { ...CONSUMED, consumed_by: ID(6) },
    ]) {
      reply = answer(200, { invitations: [bad], next: null });
      assert.deepEqual(await mem.listInvitations(CTX, SECRET, ORIGIN, bad.scope === "group" ? GROUP : ORG), { ok: false, error: "malformed" });
    }
  });

  it("refuses an issued invitation with a malformed token or of another scope", async () => {
    for (const body of [{ invitation: INV, token: "short" }, { invitation: GINV, token: TOKEN }, { invitation: INV }]) {
      reply = answer(201, body);
      assert.deepEqual(await mem.issueInvitation(CTX, SECRET, M, ORG, { role: "org_member", recipient: { address: "a@b.test" } }), { ok: false, error: "malformed" });
    }
  });

  it("refuses an owner role in every invitation answer of either scope (only the same scope's admin or member)", async () => {
    // Admin and member of the same scope are read.
    for (const [inv, role] of [[INV, "org_admin"], [GINV, "group_admin"]]) {
      reply = answer(200, { invitation: { ...inv, role } });
      const ok = await mem.getInvitation(CTX, SECRET, ORIGIN, inv.id);
      assert.equal(ok.ok, true);
      assert.equal(ok.value.role, role);
    }
    const ORG_OWNER = { ...INV, role: "org_owner" };
    const GROUP_OWNER = { ...GINV, role: "group_owner" };
    // Administrator lists of each scope.
    reply = answer(200, { invitations: [ORG_OWNER], next: null });
    assert.deepEqual(await mem.listInvitations(CTX, SECRET, ORIGIN, ORG), { ok: false, error: "malformed" });
    reply = answer(200, { invitations: [GROUP_OWNER], next: null });
    assert.deepEqual(await mem.listInvitations(CTX, SECRET, ORIGIN, GROUP), { ok: false, error: "malformed" });
    // The recipient's own list and the single invitation read.
    reply = answer(200, { invitations: [GROUP_OWNER], next: null });
    assert.deepEqual(await mem.myInvitations(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    for (const inv of [ORG_OWNER, GROUP_OWNER]) {
      reply = answer(200, { invitation: inv });
      assert.deepEqual(await mem.getInvitation(CTX, SECRET, ORIGIN, inv.id), { ok: false, error: "malformed" });
    }
    // The issuance answer of each scope.
    reply = answer(201, { invitation: ORG_OWNER, token: TOKEN });
    assert.deepEqual(await mem.issueInvitation(CTX, SECRET, M, ORG, { role: "org_member", recipient: { address: "new@example.test" } }), { ok: false, error: "malformed" });
    reply = answer(201, { invitation: GROUP_OWNER, token: TOKEN });
    assert.deepEqual(await mem.issueInvitation(CTX, SECRET, M, GROUP, { role: "group_member", recipient: { userId: ID(5) } }), { ok: false, error: "malformed" });
  });

  it("returns only the eligibility count and truncation flag", async () => {
    const scope = { organization_id: ID(1), scope_type: "group", scope_id: ID(77) };
    reply = answer(200, { count: 1, scopes: [scope], truncated: false });
    assert.deepEqual(await mem.retirementEligibility(CTX, SECRET, ORIGIN), { ok: true, value: { count: 1, truncated: false } });
    assert.equal(sent().url, `${CORE}/internal/v1/account/retirement-eligibility?origin=${encodeURIComponent(ORIGIN)}`);
    for (const body of [{ count: 2, scopes: [scope], truncated: false }, { count: 0, scopes: [scope], truncated: true }, { count: 1, scopes: [{ ...scope, name: "x" }], truncated: false }]) {
      reply = answer(200, body);
      assert.deepEqual(await mem.retirementEligibility(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    }
  });

  it("reads closed refusals with their field and nothing else", async () => {
    reply = answer(409, { error: "conflict", field: "last_owner" });
    assert.deepEqual(await mem.leave(CTX, SECRET, M, ORG), { ok: false, error: "conflict", field: "last_owner" });
    reply = answer(403, { error: "forbidden", field: "owner_subject" });
    assert.deepEqual(await mem.removeMember(CTX, SECRET, M, ORG, ID(1)), { ok: false, error: "forbidden", field: "owner_subject" });
    for (const [status, body] of [[404, { error: "not_found", field: "x" }], [500, { error: "boom" }], [409, { error: "conflict", field: "Bad Field" }]]) {
      reply = answer(status, body);
      assert.deepEqual(await mem.leave(CTX, SECRET, M, ORG), { ok: false, error: "malformed" });
    }
    reply = answer(503, { error: "membership_unavailable" });
    const r = await mem.leave(CTX, SECRET, M, ORG);
    assert.deepEqual(r, { ok: false, error: "membership_unavailable" });
    assert.equal(mem.outcomeUnknown(r.error), true);
  });
});

// G060 (Plan r0 §1.4, §4.2, D-3): the Organization's Group-authority read and owner-only write (the seventeenth and
// eighteenth method/path pairs).
describe("membership client: Group authority", () => {
  it("reads the override state and the owner hint with the origin query, the read headers and no body", async () => {
    reply = answer(200, { override_enabled: false, can_change: true });
    assert.deepEqual(await mem.groupAuthority(CTX, SECRET, ORIGIN), { ok: true, value: { overrideEnabled: false, canChange: true } });
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/organization/group-authority?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(s.method, "GET");
    assert.deepEqual(s.headers, { ...outboundHeaders(CTX), Authorization: `Session ${SECRET}` });
    assert.equal(s.body, undefined);
  });

  it("sets the override with exactly origin, csrf_token and enabled, and requires the answer to report that value", async () => {
    reply = answer(200, { override_enabled: true, can_change: true });
    assert.deepEqual(await mem.setGroupAuthority(CTX, SECRET, M, true), { ok: true, value: { overrideEnabled: true, canChange: true } });
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/organization/group-authority`);
    assert.equal(s.method, "POST");
    assert.deepEqual(s.body, { origin: ORIGIN, csrf_token: CSRF, enabled: true });
    assert.deepEqual(s.headers, { ...outboundHeaders(CTX), Authorization: `Session ${SECRET}`, "Content-Type": "application/json" });
    // An answer reporting the other value is not a success.
    reply = answer(200, { override_enabled: true, can_change: true });
    assert.deepEqual(await mem.setGroupAuthority(CTX, SECRET, M, false), { ok: false, error: "malformed" });
    // A non-owner is refused by core (forbidden), never treated as success.
    reply = answer(403, { error: "forbidden" });
    assert.deepEqual(await mem.setGroupAuthority(CTX, SECRET, M, false), { ok: false, error: "forbidden" });
  });

  it("treats any answer outside the contract as malformed", async () => {
    for (const body of [
      { override_enabled: "false", can_change: true },
      { override_enabled: false },
      { override_enabled: false, can_change: true, owner: true },
      { override_enabled: null, can_change: false },
      [],
    ]) {
      reply = answer(200, body);
      assert.deepEqual(await mem.groupAuthority(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
    reply = answer(200, `{"override_enabled":false,"can_change":false,"pad":"${"x".repeat(200)}"}`);
    assert.deepEqual(await mem.groupAuthority(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
  });
});
