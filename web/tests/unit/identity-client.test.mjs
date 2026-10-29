// Unit tests for lib/identity-client.ts: the exact outbound request of every core auth call (headers,
// method, body, cache, redirect and timeout), and the validation that reduces core's answers to closed
// outcomes. fetch is replaced by a recorder, so nothing leaves the process. Run by `npm run test:unit` in the
// web check image with networking disabled. The module imports its siblings without file extensions, as the
// Next bundler resolves them; a resolve hook scoped to that module maps them to the .ts sources.
import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import { afterEach, beforeEach, describe, it, mock } from "node:test";

registerHooks({
  resolve(specifier, context, nextResolve) {
    if ((specifier === "./config" || specifier === "./request-context") && context.parentURL?.endsWith("/lib/identity-client.ts")) {
      return nextResolve(`${specifier}.ts`, context);
    }
    return nextResolve(specifier, context);
  },
});

const client = await import("../../lib/identity-client.ts");
const { deriveContext } = await import("../../lib/request-context.ts");

const CORE = "http://core.internal:8180";
const SECRET = "S".repeat(43);
const BINDING = "B".repeat(43);
const CSRF = "C".repeat(43);
const UUID = "0b8f6c1e-2a4d-4c7b-9e1f-3a5b7c9d1e2f";
const ORIGIN = "https://org-a.localhost:543";
const CTX = deriveContext(new Headers({ "x-request-id": "req-identity-1", traceparent: `00-${"a".repeat(32)}-${"b".repeat(16)}-01` }));

const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };
let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "identity_unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), init });
    return reply(url, init);
  });
});
afterEach(() => {
  mock.restoreAll();
  for (const [key, value] of Object.entries(saved)) {
    if (value === undefined) delete process.env[key];
    else process.env[key] = value;
  }
});

function only() {
  assert.equal(calls.length, 1);
  return calls[0];
}

function headerNames(init) {
  return Object.keys(init.headers).map((name) => name.toLowerCase()).sort();
}

const DIAGNOSTIC = { "X-Request-Id": "req-identity-1" };

describe("outbound requests", () => {
  it("login sends exactly the diagnostic headers, the credential, the binding and a JSON body", async () => {
    reply = () => Response.json({ authorization_url: "https://api.workos.test/user_management/authorize?x=1", binding: { value: BINDING, max_age: 600 } });
    const result = await client.login(CTX, { origin: ORIGIN, returnTo: "/account" }, SECRET, BINDING);
    assert.equal(result.ok, true);
    const { url, init } = only();
    assert.equal(url, `${CORE}/internal/v1/auth/login`);
    assert.equal(init.method, "POST");
    assert.deepEqual(headerNames(init), ["authorization", "content-type", "traceparent", "x-login-binding", "x-request-id"]);
    assert.equal(init.headers["X-Request-Id"], DIAGNOSTIC["X-Request-Id"]);
    assert.match(init.headers.traceparent, new RegExp(`^00-${"a".repeat(32)}-[0-9a-f]{16}-01$`));
    assert.equal(init.headers.Authorization, `Session ${SECRET}`);
    assert.equal(init.headers["X-Login-Binding"], BINDING);
    assert.equal(init.headers["Content-Type"], "application/json");
    assert.deepEqual(JSON.parse(init.body), { origin: ORIGIN, return_to: "/account" });
    assert.equal(init.cache, "no-store");
    assert.equal(init.redirect, "error");
    assert.ok(init.signal instanceof AbortSignal);
    assert.equal(client.CORE_TIMEOUT_MS, 10_000);
  });
  it("login omits absent credential, binding and return path", async () => {
    await client.login(CTX, { origin: ORIGIN });
    const { init } = only();
    assert.deepEqual(headerNames(init), ["content-type", "traceparent", "x-request-id"]);
    assert.deepEqual(JSON.parse(init.body), { origin: ORIGIN });
  });
  it("callback sends the binding and credential with the raw query", async () => {
    await client.callback(CTX, { origin: ORIGIN, rawQuery: "code=c&state=s" }, BINDING, SECRET);
    const { url, init } = only();
    assert.equal(url, `${CORE}/internal/v1/auth/callback`);
    assert.deepEqual(headerNames(init), ["authorization", "content-type", "traceparent", "x-login-binding", "x-request-id"]);
    assert.deepEqual(JSON.parse(init.body), { origin: ORIGIN, raw_query: "code=c&state=s" });
  });
  it("session is a GET with the credential and its origin, and no body, binding or content type", async () => {
    await client.session(CTX, SECRET, ORIGIN);
    const { url, init } = only();
    // G016 A2: the origin is required; core answers only for a session bound to that origin's Organization.
    assert.equal(url, `${CORE}/internal/v1/auth/session?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(init.method, "GET");
    assert.equal(init.body, undefined);
    assert.deepEqual(headerNames(init), ["authorization", "traceparent", "x-request-id"]);
  });
  it("logout never sends a binding", async () => {
    await client.logout(CTX, { scope: "provider", csrfToken: CSRF, origin: ORIGIN }, SECRET);
    const { url, init } = only();
    assert.equal(url, `${CORE}/internal/v1/auth/logout`);
    assert.deepEqual(headerNames(init), ["authorization", "content-type", "traceparent", "x-request-id"]);
    assert.deepEqual(JSON.parse(init.body), { scope: "provider", csrf_token: CSRF, origin: ORIGIN });
  });
  it("the provider-logout return sends neither credential nor binding", async () => {
    await client.providerLogoutReturn(CTX, SECRET);
    const { url, init } = only();
    assert.equal(url, `${CORE}/internal/v1/auth/provider-logout/return`);
    assert.deepEqual(headerNames(init), ["content-type", "traceparent", "x-request-id"]);
    assert.deepEqual(JSON.parse(init.body), { logout_ref: SECRET });
  });
});

describe("answers", () => {
  it("maps closed core errors and refuses anything else as malformed", async () => {
    reply = () => Response.json({ error: "csrf_failed" }, { status: 403 });
    assert.deepEqual(await client.logout(CTX, { scope: "current", csrfToken: "", origin: ORIGIN }), { ok: false, error: "csrf_failed" });
    for (const [status, body] of [
      [403, { error: "not_a_code" }],
      [403, { error: "csrf_failed", detail: "x" }],
      [500, "Internal Server Error"],
      [200, []],
      [200, "text"],
      [302, { error: "csrf_failed" }],
    ]) {
      reply = () => (typeof body === "string" ? new Response(body, { status }) : Response.json(body, { status }));
      assert.deepEqual(await client.session(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" }, `${status} ${JSON.stringify(body)}`);
    }
  });
  it("distinguishes no answer in time from no connection, and bounds the body", async () => {
    reply = () => {
      throw new DOMException("The operation timed out.", "TimeoutError");
    };
    assert.deepEqual(await client.session(CTX, SECRET, ORIGIN), { ok: false, error: "timeout" });
    reply = () => {
      throw new TypeError("fetch failed");
    };
    assert.deepEqual(await client.session(CTX, SECRET, ORIGIN), { ok: false, error: "unreachable" });
    reply = () => new Response(`{"error":"${"x".repeat(70 * 1024)}"}`, { status: 503 });
    assert.deepEqual(await client.session(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
  });
  it("validates a login answer", async () => {
    reply = () => Response.json({ authorization_url: "https://api.workos.test/user_management/authorize?client_id=c", binding: { value: BINDING, max_age: 600 } });
    assert.deepEqual(await client.login(CTX, { origin: ORIGIN }), {
      ok: true,
      value: { authorizationUrl: "https://api.workos.test/user_management/authorize?client_id=c", binding: { value: BINDING, maxAge: 600 } },
    });
    for (const body of [
      { authorization_url: "javascript:alert(1)", binding: { value: BINDING, max_age: 600 } },
      { authorization_url: "/relative", binding: { value: BINDING, max_age: 600 } },
      { authorization_url: "https://x.test/", binding: { value: "short", max_age: 600 } },
      { authorization_url: "https://x.test/", binding: { value: BINDING, max_age: 0 } },
      { authorization_url: "https://x.test/" },
    ]) {
      reply = () => Response.json(body);
      assert.deepEqual(await client.login(CTX, { origin: ORIGIN }), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
    reply = () => Response.json({ error: "already_authenticated" }, { status: 409 });
    assert.deepEqual(await client.login(CTX, { origin: ORIGIN }, SECRET), { ok: false, error: "already_authenticated" });
  });
  it("validates a callback answer", async () => {
    reply = () => Response.json({ outcome: "signed_in", return_to: "/account", session: { value: SECRET, max_age: 28800 } });
    assert.deepEqual(await client.callback(CTX, { origin: ORIGIN, rawQuery: "" }, BINDING), {
      ok: true,
      value: { outcome: "signed_in", returnTo: "/account", session: { value: SECRET, maxAge: 28800 } },
    });
    reply = () => Response.json({ outcome: "already_signed_in", return_to: "/" });
    assert.deepEqual(await client.callback(CTX, { origin: ORIGIN, rawQuery: "" }, BINDING, SECRET), {
      ok: true,
      value: { outcome: "already_signed_in", returnTo: "/" },
    });
    for (const body of [
      { outcome: "signed_in", return_to: "/account" },
      { outcome: "signed_in", return_to: "//evil.example", session: { value: SECRET, max_age: 60 } },
      { outcome: "signed_in", return_to: "/elsewhere", session: { value: SECRET, max_age: 60 } },
      { outcome: "verified", return_to: "/", session: { value: SECRET, max_age: 60 } },
    ]) {
      reply = () => Response.json(body);
      assert.deepEqual(await client.callback(CTX, { origin: ORIGIN, rawQuery: "" }, BINDING), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
  });
  it("validates a session answer and never accepts an assurance other than unknown", async () => {
    const good = {
      user_id: UUID,
      display_name: null,
      mfa_assurance: "unknown",
      auth_time: "2026-10-07T12:00:00Z",
      expires_at: "2026-10-07T12:30:00Z",
      csrf_token: CSRF,
      membership: "member",
    };
    reply = () => Response.json(good);
    assert.deepEqual(await client.session(CTX, SECRET, ORIGIN), {
      ok: true,
      value: { userId: UUID, displayName: null, mfaAssurance: "unknown", authTime: good.auth_time, expiresAt: good.expires_at, csrfToken: CSRF, membership: "member" },
    });
    // G017 D-6: the closed membership answer; "none" is a valid session (signed in, not a member of this Organization).
    reply = () => Response.json({ ...good, membership: "none" });
    const none = await client.session(CTX, SECRET, ORIGIN);
    assert.equal(none.ok && none.value.membership, "none");
    for (const change of [
      { membership: undefined },
      { membership: "owner" },
      { membership: true },
      { mfa_assurance: "verified" },
      { mfa_assurance: undefined },
      { user_id: "user_01" },
      { display_name: 7 },
      { auth_time: "yesterday" },
      { expires_at: "2026-10-07 12:30:00" },
      { csrf_token: "short" },
    ]) {
      reply = () => Response.json({ ...good, ...change });
      assert.deepEqual(await client.session(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" }, JSON.stringify(change));
    }
  });
  it("validates logout answers: the reference only in its own field, the fixed /sign-in return (C7 addendum §2)", async () => {
    const LOGOUT_URL = "https://api.workos.test/user_management/sessions/logout?session_id=s&return_to=https%3A%2F%2Forg-a.localhost%3A543%2Fsign-in";
    const cases = [
      [{ outcome: "local_revoked" }, { ok: true, value: { outcome: "local_revoked" } }],
      [
        { outcome: "provider_logout_redirected", provider_logout_url: LOGOUT_URL, return_to: "/sign-in", logout_ref: SECRET },
        { ok: true, value: { outcome: "provider_logout_redirected", providerLogoutUrl: LOGOUT_URL, logoutRef: SECRET } },
      ],
      [{ outcome: "provider_revoke_failed", return_to: "/sign-in", logout_ref: SECRET }, { ok: true, value: { outcome: "provider_revoke_failed", logoutRef: SECRET } }],
      [{ outcome: "provider_session_revoked", return_to: "/sign-in", logout_ref: SECRET }, { ok: true, value: { outcome: "provider_session_revoked", logoutRef: SECRET } }],
      // The reference is required, canonical and never part of a URL; the return is exactly /sign-in.
      [{ outcome: "provider_logout_redirected", provider_logout_url: LOGOUT_URL, return_to: "/sign-in" }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_logout_redirected", provider_logout_url: `${LOGOUT_URL}&r=${SECRET}`, return_to: "/sign-in", logout_ref: SECRET }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_logout_redirected", provider_logout_url: LOGOUT_URL, return_to: "/sign-in", logout_ref: "A".repeat(42) }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_logout_redirected", provider_logout_url: LOGOUT_URL, return_to: "/account", logout_ref: SECRET }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_logout_redirected", provider_logout_url: "javascript:alert(1)", return_to: "/sign-in", logout_ref: SECRET }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_revoke_failed", return_to: `/sign-in?logout_ref=${SECRET}` }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_revoke_failed", return_to: `/sign-in?logout_ref=${SECRET}`, logout_ref: SECRET }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_revoke_failed", return_to: "https://evil.example/sign-in", logout_ref: SECRET }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_revoke_failed", return_to: "/sign-in", logout_ref: [SECRET] }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_session_revoked", return_to: "/sign-in", logout_ref: SECRET, provider_logout_url: LOGOUT_URL }, { ok: false, error: "malformed" }],
      [{ outcome: "provider_logout_redirected" }, { ok: false, error: "malformed" }],
      [{ outcome: "signed_out" }, { ok: false, error: "malformed" }],
    ];
    for (const [body, want] of cases) {
      reply = () => Response.json(body);
      assert.deepEqual(await client.logout(CTX, { scope: "provider", csrfToken: CSRF, origin: ORIGIN }, SECRET), want, JSON.stringify(body));
    }
  });
  it("validates provider-logout return outcomes", async () => {
    for (const outcome of ["local_revoked", "provider_session_revoked", "provider_revoke_failed"]) {
      reply = () => Response.json({ outcome });
      assert.deepEqual(await client.providerLogoutReturn(CTX, SECRET), { ok: true, value: outcome });
    }
    reply = () => Response.json({ outcome: "provider_logout_returned" });
    assert.deepEqual(await client.providerLogoutReturn(CTX, SECRET), { ok: false, error: "malformed" });
  });
});
