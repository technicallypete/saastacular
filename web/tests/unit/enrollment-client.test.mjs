// Unit tests for the G058 enrollment client (Plan r3 §1.6, §5.2–§5.3; R13 route reconciliation): the ten route functions
// (path, method, the one Authorization scheme, exact body fields, strict answer shapes, 200 for a new and an existing
// request), the enrollment variants of the auth login and callback (202 credential {value, max_age} only, never mixed
// with a session), and the enrollment and join cookies. fetch is a recorder; nothing leaves the process.
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

const enr = await import("../../lib/enrollment-client.ts");
const session = await import("../../lib/session.ts");

const CORE = "http://core.internal:8180";
const ORIGIN = "https://org-c.localhost:543";
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const TOKEN = "T".repeat(43);
const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };
const ID = (n) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
const { deriveContext } = await import("../../lib/request-context.ts");
const CTX = deriveContext(new Headers());
const M = { origin: ORIGIN, csrfToken: CSRF };
const OWN = {
  id: ID(1),
  status: "pending",
  expired: false,
  consumed: false,
  created_at: "2026-10-18T10:00:00Z",
  expires_at: "2026-11-17T10:00:00Z",
  decided_at: null,
  revoked_at: null,
};
const MANAGED = { ...OWN, user_id: ID(9), display_name: null };

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "enrollment_unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), method: init.method, headers: init.headers, body: init.body === undefined ? undefined : JSON.parse(init.body) });
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

const answer = (body, status = 200) => () => Response.json(body, { status });

describe("the ten enrollment routes", () => {
  it("public: no credential, exactly {mode, label}", async () => {
    reply = answer({ mode: "request_access", label: null });
    assert.deepEqual(await enr.publicEnrollment(CTX, ORIGIN), { ok: true, value: { mode: "request_access", label: null } });
    assert.equal(calls[0].url, `${CORE}/internal/v1/enrollment/public?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(calls[0].method, "GET");
    assert.equal(calls[0].headers.Authorization, undefined);
    reply = answer({ mode: "request_access", label: null, extra: 1 });
    assert.deepEqual(await enr.publicEnrollment(CTX, ORIGIN), { ok: false, error: "malformed" });
    reply = answer({ mode: "everyone", label: null });
    assert.deepEqual(await enr.publicEnrollment(CTX, ORIGIN), { ok: false, error: "malformed" });
    reply = answer({ error: "unknown_host" }, 404);
    assert.deepEqual(await enr.publicEnrollment(CTX, ORIGIN), { ok: false, error: "unknown_host" });
  });

  it("status: the enrollment scheme; status, request, current member and own csrf_token; never a session", async () => {
    reply = answer({ status: "pending", request: OWN, member: false, csrf_token: CSRF });
    const r = await enr.enrollmentStatus(CTX, SECRET, ORIGIN);
    assert.equal(r.ok, true);
    assert.equal(r.value.member, false);
    assert.equal(r.value.csrfToken, CSRF);
    assert.equal(r.value.request.id, ID(1));
    assert.equal(calls[0].headers.Authorization, `Enrollment ${SECRET}`);
    // historical admitted with a removed membership stays member:false
    reply = answer({ status: "admitted", request: null, member: false, csrf_token: CSRF });
    assert.deepEqual((await enr.enrollmentStatus(CTX, SECRET, ORIGIN)).value.status, "admitted");
    for (const bad of [
      { status: "none", request: OWN, member: false, csrf_token: CSRF },
      { status: "approved", request: OWN, member: false, csrf_token: CSRF },
      { status: "pending", request: OWN, member: "no", csrf_token: CSRF },
      { status: "pending", request: OWN, member: false },
      { status: "pending", request: { ...OWN, user_id: ID(9) }, member: false, csrf_token: CSRF },
      { status: "member", request: null, member: true, csrf_token: CSRF },
    ]) {
      reply = answer(bad);
      assert.deepEqual(await enr.enrollmentStatus(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" }, JSON.stringify(bad));
    }
    reply = answer({ error: "unauthenticated" }, 401);
    assert.deepEqual(await enr.enrollmentStatus(CTX, SECRET, ORIGIN), { ok: false, error: "unauthenticated" });
  });

  it("request: 200 for new and existing; body origin and csrf_token only", async () => {
    reply = answer({ request: OWN });
    assert.deepEqual((await enr.requestEnrollment(CTX, SECRET, M)).value.id, ID(1));
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF });
    assert.equal(calls[0].url, `${CORE}/internal/v1/enrollment/request`);
    reply = answer({ request: OWN }, 201);
    assert.deepEqual(await enr.requestEnrollment(CTX, SECRET, M), { ok: false, error: "malformed" });
    reply = answer({ error: "enrollment_closed" }, 403);
    assert.deepEqual(await enr.requestEnrollment(CTX, SECRET, M), { ok: false, error: "enrollment_closed" });
    reply = answer({ error: "conflict", field: "already_member" }, 409);
    assert.deepEqual(await enr.requestEnrollment(CTX, SECRET, M), { ok: false, error: "conflict", field: "already_member" });
  });

  it("admit: open and request kinds; outcome with next sign_in and no session", async () => {
    reply = answer({ outcome: "admitted", next: "sign_in" });
    assert.deepEqual(await enr.admit(CTX, SECRET, M, "open"), { ok: true, value: { outcome: "admitted" } });
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, kind: "open" });
    assert.deepEqual(await enr.admit(CTX, SECRET, M, "request", ID(1)), { ok: true, value: { outcome: "admitted" } });
    assert.deepEqual(calls[1].body, { origin: ORIGIN, csrf_token: CSRF, kind: "request", request_id: ID(1) });
    assert.deepEqual(await enr.admit(CTX, SECRET, M, "request", "x"), { ok: false, error: "invalid_request", field: "request_id" });
    reply = answer({ outcome: "admitted", next: "sign_in", session: { value: SECRET, max_age: 60 } });
    assert.deepEqual(await enr.admit(CTX, SECRET, M, "open"), { ok: false, error: "malformed" });
    reply = answer({ error: "conflict", field: "approval_revoked" }, 409);
    assert.deepEqual(await enr.admit(CTX, SECRET, M, "open"), { ok: false, error: "conflict", field: "approval_revoked" });
  });

  it("abandon: exactly {}", async () => {
    reply = answer({});
    assert.deepEqual(await enr.abandon(CTX, SECRET, M), { ok: true, value: null });
    reply = answer({ ok: true });
    assert.deepEqual(await enr.abandon(CTX, SECRET, M), { ok: false, error: "malformed" });
  });

  it("policy GET/POST: the session scheme; open implies ready; enrollment_not_ready conflict", async () => {
    reply = answer({ mode: "invite_only", open_ready: false, history_basis: null });
    assert.deepEqual(await enr.getPolicy(CTX, SECRET, ORIGIN), { ok: true, value: { mode: "invite_only", openReady: false, historyBasis: null } });
    assert.equal(calls[0].headers.Authorization, `Session ${SECRET}`);
    reply = answer({ mode: "open", open_ready: false, history_basis: null });
    assert.deepEqual(await enr.getPolicy(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    reply = answer({ mode: "request_access", open_ready: true, history_basis: "created_after_ledger" });
    assert.equal((await enr.updatePolicy(CTX, SECRET, M, "request_access")).ok, true);
    assert.deepEqual(calls[2].body, { origin: ORIGIN, csrf_token: CSRF, mode: "request_access" });
    reply = answer({ error: "conflict", field: "enrollment_not_ready" }, 409);
    assert.deepEqual(await enr.updatePolicy(CTX, SECRET, M, "open"), { ok: false, error: "conflict", field: "enrollment_not_ready" });
    assert.deepEqual(await enr.updatePolicy(CTX, SECRET, M, "everyone"), { ok: false, error: "invalid_request", field: "invalid_enrollment_mode" });
  });

  it("requests: keyset page, ascending ids, bounded limit", async () => {
    reply = answer({ requests: [MANAGED, { ...MANAGED, id: ID(2) }], next: ID(2) });
    const r = await enr.listRequests(CTX, SECRET, ORIGIN, { after: ID(0), status: "pending" });
    assert.equal(r.value.requests.length, 2);
    assert.equal(calls[0].url, `${CORE}/internal/v1/enrollment/requests?origin=${encodeURIComponent(ORIGIN)}&limit=50&after=${ID(0)}&status=pending`);
    reply = answer({ requests: [{ ...MANAGED, id: ID(2) }, MANAGED], next: null });
    assert.deepEqual(await enr.listRequests(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    reply = answer({ requests: [OWN], next: null });
    assert.deepEqual(await enr.listRequests(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    assert.deepEqual(await enr.listRequests(CTX, SECRET, ORIGIN, { limit: 101 }), { ok: false, error: "invalid_request" });
    reply = answer({ error: "forbidden" }, 403);
    assert.deepEqual(await enr.listRequests(CTX, SECRET, ORIGIN), { ok: false, error: "forbidden" });
  });

  it("decide and revoke: path id, body fields, the manager row", async () => {
    reply = answer({ request: { ...MANAGED, status: "approved", decided_at: "2026-10-18T11:00:00Z" } });
    assert.equal((await enr.decideRequest(CTX, SECRET, M, ID(1), "approve")).value.status, "approved");
    assert.equal(calls[0].url, `${CORE}/internal/v1/enrollment/requests/${ID(1)}/decide`);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, decision: "approve" });
    reply = answer({ request: { ...MANAGED, status: "revoked", decided_at: "2026-10-18T11:00:00Z", revoked_at: "2026-10-18T12:00:00Z" } });
    assert.equal((await enr.revokeRequest(CTX, SECRET, M, ID(1))).value.status, "revoked");
    assert.deepEqual(calls[1].body, { origin: ORIGIN, csrf_token: CSRF });
    assert.deepEqual(await enr.revokeRequest(CTX, SECRET, M, "x"), { ok: false, error: "not_found" });
    reply = answer({ error: "conflict", field: "already_consumed" }, 409);
    assert.deepEqual(await enr.revokeRequest(CTX, SECRET, M, ID(1)), { ok: false, error: "conflict", field: "already_consumed" });
  });
});

describe("the enrollment auth variants", () => {
  it("startEnrollment sends intent enroll, the invitation id and the request flag", async () => {
    reply = answer({ authorization_url: "https://auth.example.test/authorize?x=1", binding: { value: SECRET, max_age: 600 } });
    const r = await enr.startEnrollment(CTX, { origin: ORIGIN, invitationId: ID(5), request: true }, undefined, SECRET);
    assert.equal(r.ok, true);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, intent: "enroll", invitation_id: ID(5), request: true });
    assert.equal(calls[0].headers["X-Login-Binding"], SECRET);
    assert.deepEqual(await enr.startEnrollment(CTX, { origin: ORIGIN, invitationId: "x" }), { ok: false, error: "invalid_request" });
  });

  it("completeCallback: 202 credential {value, max_age ≤ 1800}; 200 admitted with no session; ordinary answers", async () => {
    reply = answer({ enrollment: { status: "pending", credential: { value: SECRET, max_age: 1800 } } }, 202);
    assert.deepEqual(await enr.completeCallback(CTX, { origin: ORIGIN, rawQuery: "code=a&state=b", invitation: { invitationId: ID(5), token: TOKEN } }, SECRET), {
      ok: true,
      value: { kind: "enrollment", status: "pending", credential: { value: SECRET, maxAge: 1800 } },
    });
    assert.deepEqual(calls[0].body, { origin: ORIGIN, raw_query: "code=a&state=b", invitation_id: ID(5), invitation_token: TOKEN });
    for (const bad of [
      { enrollment: { status: "pending", credential: { value: SECRET, max_age: 1801 } } },
      { enrollment: { status: "pending", credential: SECRET } },
      { enrollment: { status: "member", credential: { value: SECRET, max_age: 60 } } },
      { enrollment: { status: "pending", credential: { value: SECRET, max_age: 60 } }, session: { value: SECRET, max_age: 60 } },
    ]) {
      reply = answer(bad, 202);
      assert.deepEqual(await enr.completeCallback(CTX, { origin: ORIGIN, rawQuery: "q" }), { ok: false, error: "malformed" }, JSON.stringify(bad));
    }
    reply = answer({ outcome: "admitted", next: "sign_in" });
    assert.deepEqual(await enr.completeCallback(CTX, { origin: ORIGIN, rawQuery: "q" }), { ok: true, value: { kind: "admitted", outcome: "admitted" } });
    reply = answer({ outcome: "admitted", next: "sign_in", session: { value: SECRET, max_age: 60 } });
    assert.deepEqual(await enr.completeCallback(CTX, { origin: ORIGIN, rawQuery: "q" }), { ok: false, error: "malformed" });
    reply = answer({ outcome: "signed_in", return_to: "/join/status", session: { value: SECRET, max_age: 3600 } });
    assert.equal((await enr.completeCallback(CTX, { origin: ORIGIN, rawQuery: "q" })).value.kind, "session");
    reply = answer({ error: "invitation_mismatch" }, 409);
    assert.deepEqual(await enr.completeCallback(CTX, { origin: ORIGIN, rawQuery: "q" }), { ok: false, error: "invitation_mismatch" });
  });

  it("G060 Option B: invitation_mismatch shows the fixed generic invalid-invitation notice", () => {
    const expected = "Invalid invitation link. Switch to a different account, or ask for a new invitation link.";
    assert.equal(enr.INVALID_INVITATION, expected);
    assert.equal(enr.joinNotice("invitation_mismatch"), expected);
    assert.equal(enr.JOIN_NOTICES.invitation_mismatch, expected);
    // Generic: no account, email or existence wording; the other notices are unchanged.
    assert.doesNotMatch(expected, /@|email|did not match|opened last|expired|revoked|exist/i);
    assert.equal(enr.joinNotice("not_found"), "That invitation or request is not available. Ask for a new invitation.");
    assert.equal(enr.joinNotice("unavailable"), "Joining is unavailable right now. Please try again later.");
  });
});

describe("enrollment and join cookies", () => {
  it("names, lifetime bounds and the separate join serializer", () => {
    assert.equal(session.cookieName("enrollment", "https"), "__Host-enroll");
    assert.equal(session.cookieName("join", "loopback-http"), "join");
    assert.match(session.setCookie("enrollment", SECRET, 1800, "https"), /^__Host-enroll=S{43}; Path=\/; Max-Age=1800; HttpOnly; SameSite=Lax; Secure$/);
    assert.throws(() => session.setCookie("enrollment", SECRET, 1801, "https"), RangeError);
    assert.throws(() => session.setCookie("join", SECRET, 60, "https"), TypeError);
    const join = session.setJoinCookie(ID(5), TOKEN, "https");
    assert.equal(join, `__Host-join=${ID(5)}.${TOKEN}; Path=/; Max-Age=600; HttpOnly; SameSite=Lax; Secure`);
    assert.throws(() => session.setJoinCookie("x", TOKEN, "https"), TypeError);
  });

  it("reads exactly one well-formed join cookie; duplicates and malformed values are absent", () => {
    const v = `${ID(5)}.${TOKEN}`;
    assert.deepEqual(session.readJoinCookie(`a=b; __Host-join=${v}`, "https"), { invitationId: ID(5), token: TOKEN });
    assert.equal(session.readJoinCookie(`__Host-join=${v}; __Host-join=${v}`, "https"), undefined);
    assert.equal(session.readJoinCookie(`__Host-join=${ID(5)}.short`, "https"), undefined);
    assert.equal(session.readCookie(`__Host-join=${v}`, "join", "https"), undefined);
    assert.equal(session.readCookie(`__Host-enroll=${SECRET}`, "session", "https"), undefined);
    assert.equal(session.readCookie(`__Host-enroll=${SECRET}`, "enrollment", "https"), SECRET);
  });
});

describe("the administrators' request row view", () => {
  const row = (fields) => ({ ...OWN, userId: ID(9), displayName: null, createdAt: OWN.created_at, ...fields });
  it("an expired pending request keeps Deny only, with truthful copy; approval is not offered", () => {
    assert.deepEqual(enr.requestRowView(row({ status: "pending", expired: true })), { state: "Expired, deny to resolve", actions: ["deny"] });
    assert.deepEqual(enr.requestRowView(row({ status: "pending", expired: false })), { state: "Pending", actions: ["approve", "deny"] });
  });
  it("an expired approval stays non-consumable and offers nothing; a valid unconsumed approval offers Revoke", () => {
    assert.deepEqual(enr.requestRowView(row({ status: "approved", expired: true })), { state: "Expired", actions: [] });
    assert.deepEqual(enr.requestRowView(row({ status: "approved", expired: false })), { state: "Approved", actions: ["revoke"] });
    assert.deepEqual(enr.requestRowView(row({ status: "approved", consumed: true })), { state: "Joined", actions: [] });
    assert.deepEqual(enr.requestRowView(row({ status: "denied", expired: true })), { state: "Denied", actions: [] });
    assert.deepEqual(enr.requestRowView(row({ status: "revoked" })), { state: "Approval withdrawn", actions: [] });
  });
});
