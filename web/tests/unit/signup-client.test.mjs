// Unit tests for lib/signup-client.ts (G062 Plan r5 §4.2, §4.4; SUP-14): the exact outbound request of every core signup
// call (headers: "Signup <credential>" on the root routes only, X-Login-Binding and no credential on the tenant hops; the
// documented body members only), the strict reading of core's answers and closed refusals, the continuation URL rules
// (core's URLs are accepted only for the configured root or one of its platform-subdomain tenants), the auth callback's
// G062 answers (the root's 202 signup credential and the tenant's tenant_created) next to G058's, the root resolution of
// the root and tenant routes, and the hop query rule. G057 (Plan r5 §3.3, §5.2 T7; D-5): a tenant origin is a generated
// platform label under the root, validated structurally (other port, extra label, slug-equal host and http refused);
// web never builds an origin from the slug. fetch is a recorder; nothing leaves the process.
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

const client = await import("../../lib/signup-client.ts");
const { deriveContext } = await import("../../lib/request-context.ts");

const CORE = "http://core.internal:8180";
const ROOT = "https://example.test:543";
// Platform labels issued before the 2026-10-29 amendment (26 characters of [a-z2-7]; they keep resolving); the slug "acme" never routes.
const LABEL_A = "q3vk7m2xa5rt6pwz4bnhyc2dfe";
const LABEL_B = "abcdefghijklmnopqrstuvwxyz";
const TENANT = `https://${LABEL_A}.example.test:543`;
// A friendly generated label (<adjective>-<noun>-<four digits>, 2026-10-29 amendment); LABEL_A/LABEL_B are labels issued before it.
const FRIENDLY = "gusty-flowerpot-2643";
const CRED = "S".repeat(43);
const CSRF = "C".repeat(43);
const BINDING = "B".repeat(43);
const REF = "R".repeat(43);
const NONCE = "N".repeat(43);
const PROOF = "P".repeat(43);
const CHALLENGE = "K".repeat(43);
const ORG = "0b8f6c1e-2a4d-4c7b-9e1f-3a5b7c9d1e2f";
const CTX = deriveContext(new Headers({ "x-request-id": "req-signup-1", traceparent: `00-${"a".repeat(32)}-${"b".repeat(16)}-01` }));
const ENV = {
  PRODUCT_NAME: "Test Product",
  PLATFORM_BASE_DOMAIN: "example.test",
  CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test",
  CORE_INTERNAL_URL: CORE,
  IDENTITY_COOKIE_PROFILE: "https",
  APP_ENV: "test",
};

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "signup_unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), init, body: init.body === undefined ? undefined : JSON.parse(init.body) });
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

const M = { origin: ROOT, csrfToken: CSRF };

function assertTransport(call, method) {
  assert.equal(call.init.method, method);
  assert.equal(call.init.cache, "no-store");
  assert.equal(call.init.redirect, "error");
  assert.ok(call.init.signal instanceof AbortSignal);
  assert.equal(call.init.headers["X-Request-Id"], "req-signup-1");
}

describe("root routes", () => {
  it("status: GET with the Signup credential only; pending and provisioned answers", async () => {
    reply = () => Response.json({ status: "pending", csrf_token: CSRF });
    assert.deepEqual(await client.signupStatus(CTX, CRED, ROOT), { ok: true, value: { status: "pending", csrfToken: CSRF, organization: null } });
    const { url, init, body } = calls[0];
    assert.equal(url, `${CORE}/internal/v1/signup/status?origin=${encodeURIComponent(ROOT)}`);
    assertTransport(calls[0], "GET");
    assert.equal(init.headers.Authorization, `Signup ${CRED}`);
    assert.equal(init.headers["X-Login-Binding"], undefined);
    assert.equal(init.headers["Content-Type"], undefined);
    assert.equal(body, undefined);
    reply = () => Response.json({ status: "created", csrf_token: CSRF, organization: { slug: "acme", target_origin: TENANT } });
    assert.deepEqual(await client.signupStatus(CTX, CRED, ROOT), {
      ok: true,
      value: { status: "created", csrfToken: CSRF, organization: { slug: "acme", targetOrigin: TENANT } },
    });
  });
  it("status: anything outside the contract is malformed; refusals are closed", async () => {
    for (const bad of [
      { status: "pending" },
      { status: "pending", csrf_token: "short" },
      { status: "pending", csrf_token: CSRF, organization: { slug: "acme", target_origin: TENANT } },
      { status: "created", csrf_token: CSRF },
      { status: "created", csrf_token: CSRF, organization: { slug: "acme", target_origin: "https://other.example.test:543" } },
      { status: "created", csrf_token: CSRF, organization: { slug: "acme", target_origin: "https://acme.example.test" } },
      // G057: the slug's own host (the pre-G057 origin), the slug-equal label host, another port, an extra label, http.
      { status: "created", csrf_token: CSRF, organization: { slug: "acme", target_origin: "https://acme.example.test:543" } },
      { status: "created", csrf_token: CSRF, organization: { slug: LABEL_B, target_origin: `https://${LABEL_B}.example.test:543` } },
      { status: "created", csrf_token: CSRF, organization: { slug: "acme", target_origin: `https://${LABEL_A}.example.test:544` } },
      { status: "created", csrf_token: CSRF, organization: { slug: "acme", target_origin: `https://x.${LABEL_A}.example.test:543` } },
      { status: "created", csrf_token: CSRF, organization: { slug: "acme", target_origin: `http://${LABEL_A}.example.test:543` } },
      { status: "created", csrf_token: CSRF, organization: { slug: "Acme", target_origin: TENANT } },
      { status: "completed", csrf_token: CSRF, organization: { slug: "acme", target_origin: TENANT, name: "Acme" } },
      { status: "done", csrf_token: CSRF },
    ]) {
      reply = () => Response.json(bad);
      assert.deepEqual(await client.signupStatus(CTX, CRED, ROOT), { ok: false, error: "malformed" }, JSON.stringify(bad));
    }
    reply = () => Response.json({ error: "unauthenticated" }, { status: 401 });
    assert.deepEqual(await client.signupStatus(CTX, CRED, ROOT), { ok: false, error: "unauthenticated" });
    reply = () => Response.json({ error: "session_expired" }, { status: 401 });
    assert.deepEqual(await client.signupStatus(CTX, CRED, ROOT), { ok: false, error: "malformed" });
    reply = () => {
      throw new TypeError("fetch failed");
    };
    assert.deepEqual(await client.signupStatus(CTX, CRED, ROOT), { ok: false, error: "unreachable" });
  });
  it("slug: the exact body; the advisory answer; a closed field refusal", async () => {
    reply = () => Response.json({ suggested: "acme", available: true });
    assert.deepEqual(await client.suggestSlug(CTX, CRED, M, " Acme "), { ok: true, value: { suggested: "acme", available: true } });
    assertTransport(calls[0], "POST");
    assert.equal(calls[0].url, `${CORE}/internal/v1/signup/slug`);
    assert.deepEqual(calls[0].body, { origin: ROOT, csrf_token: CSRF, name: " Acme " });
    assert.equal(calls[0].init.headers.Authorization, `Signup ${CRED}`);
    assert.equal(calls[0].init.headers["Content-Type"], "application/json");
    reply = () => Response.json({ suggested: null, available: false });
    assert.deepEqual(await client.suggestSlug(CTX, CRED, M, "!!"), { ok: true, value: { suggested: null, available: false } });
    for (const bad of [{ suggested: "Acme", available: true }, { suggested: "acme" }, { suggested: "acme", available: "yes" }, { suggested: "acme", available: true, x: 1 }]) {
      reply = () => Response.json(bad);
      assert.deepEqual(await client.suggestSlug(CTX, CRED, M, "Acme"), { ok: false, error: "malformed" }, JSON.stringify(bad));
    }
    reply = () => Response.json({ error: "invalid_request", field: "name" }, { status: 400 });
    assert.deepEqual(await client.suggestSlug(CTX, CRED, M, ""), { ok: false, error: "invalid_request", field: "name" });
    reply = () => Response.json({ error: "invalid_request", field: "owner" }, { status: 400 });
    assert.deepEqual(await client.suggestSlug(CTX, CRED, M, ""), { ok: false, error: "malformed" });
  });
  it("organizations: 201 with core's continuation on a generated label origin of this root, never the slug's host", async () => {
    const cont = `${TENANT}/signup/continue?ref=${REF}`;
    reply = () => Response.json({ outcome: "created", organization: { id: ORG, slug: "acme" }, continue: cont }, { status: 201 });
    assert.deepEqual(await client.createOrganization(CTX, CRED, M, "Acme", "acme"), {
      ok: true,
      value: { outcome: "created", organizationId: ORG, slug: "acme", continueUrl: cont },
    });
    assert.equal(calls[0].url, `${CORE}/internal/v1/signup/organizations`);
    assert.deepEqual(calls[0].body, { origin: ROOT, csrf_token: CSRF, name: "Acme", slug: "acme" });
    assert.equal(calls[0].init.headers.Authorization, `Signup ${CRED}`);
    reply = () => Response.json({ outcome: "already_created", organization: { id: ORG, slug: "acme" }, continue: cont }, { status: 201 });
    assert.equal((await client.createOrganization(CTX, CRED, M, "Acme", "acme")).value.outcome, "already_created");
    // Core decides which label the Organization has (on already_created, the stored origin): any generated label origin
    // of this root is structurally acceptable.
    const other = `https://${LABEL_B}.example.test:543/signup/continue?ref=${REF}`;
    reply = () => Response.json({ outcome: "already_created", organization: { id: ORG, slug: "acme" }, continue: other }, { status: 201 });
    assert.equal((await client.createOrganization(CTX, CRED, M, "Acme", "acme")).value.continueUrl, other);
    for (const c of [
      `https://other.example.test:543/signup/continue?ref=${REF}`,
      `https://acme.example.test/signup/continue?ref=${REF}`,
      `http://acme.example.test:543/signup/continue?ref=${REF}`,
      // G057 (T7): the slug's own host, another port, an extra label, http, a label of the wrong shape.
      `https://acme.example.test:543/signup/continue?ref=${REF}`,
      `https://${LABEL_A}.example.test:544/signup/continue?ref=${REF}`,
      `https://${LABEL_A}.example.test/signup/continue?ref=${REF}`,
      `https://x.${LABEL_A}.example.test:543/signup/continue?ref=${REF}`,
      `https://${LABEL_A}.x.example.test:543/signup/continue?ref=${REF}`,
      `http://${LABEL_A}.example.test:543/signup/continue?ref=${REF}`,
      `https://${LABEL_A.slice(1)}.example.test:543/signup/continue?ref=${REF}`,
      `https://${LABEL_A}0.example.test:543/signup/continue?ref=${REF}`,
      `https://user@${LABEL_A}.example.test:543/signup/continue?ref=${REF}`,
      `${TENANT}/signup/complete?ref=${REF}`,
      `${TENANT}/signup/continue?ref=${REF}&x=1`,
      `${TENANT}/signup/continue?ref=${REF.slice(1)}`,
      `${TENANT}/signup/continue?n=${REF}`,
      `https://evil.test/${TENANT}/signup/continue?ref=${REF}`,
      `${ROOT}/signup/continue?ref=${REF}`,
    ]) {
      reply = () => Response.json({ outcome: "created", organization: { id: ORG, slug: "acme" }, continue: c }, { status: 201 });
      assert.deepEqual(await client.createOrganization(CTX, CRED, M, "Acme", "acme"), { ok: false, error: "malformed" }, c);
    }
    // G057 (T7): a slug-equal host is refused even when the slug has a label's shape (the created or the submitted slug).
    const slugHost = `https://${LABEL_B}.example.test:543/signup/continue?ref=${REF}`;
    reply = () => Response.json({ outcome: "created", organization: { id: ORG, slug: LABEL_B }, continue: slugHost }, { status: 201 });
    assert.deepEqual(await client.createOrganization(CTX, CRED, M, "Acme", LABEL_B), { ok: false, error: "malformed" }, "created slug's host");
    reply = () => Response.json({ outcome: "created", organization: { id: ORG, slug: "acme" }, continue: cont }, { status: 201 });
    assert.deepEqual(await client.createOrganization(CTX, CRED, M, "Acme", LABEL_A), { ok: false, error: "malformed" }, "submitted slug's host");
    reply = () => Response.json({ outcome: "created", organization: { id: ORG, slug: "Acme" }, continue: cont }, { status: 201 });
    assert.deepEqual(await client.createOrganization(CTX, CRED, M, "Acme", "acme"), { ok: false, error: "malformed" }, "a slug outside the grammar");
    reply = () => Response.json({ outcome: "created", organization: { id: ORG, slug: "acme" }, continue: cont });
    assert.deepEqual(await client.createOrganization(CTX, CRED, M, "Acme", "acme"), { ok: false, error: "malformed" }, "200 is not 201");
    reply = () => Response.json({ outcome: "created", organization: { id: "x", slug: "acme" }, continue: cont }, { status: 201 });
    assert.deepEqual(await client.createOrganization(CTX, CRED, M, "Acme", "acme"), { ok: false, error: "malformed" });
  });
  it("organizations: closed refusals with their closed fields", async () => {
    for (const [status, body, expected] of [
      [409, { error: "conflict", field: "slug_taken" }, { ok: false, error: "conflict", field: "slug_taken" }],
      [409, { error: "conflict", field: "intent_used" }, { ok: false, error: "conflict", field: "intent_used" }],
      [400, { error: "invalid_request", field: "slug" }, { ok: false, error: "invalid_request", field: "slug" }],
      [400, { error: "invalid_request" }, { ok: false, error: "invalid_request" }],
      [403, { error: "csrf_failed" }, { ok: false, error: "csrf_failed" }],
      [403, { error: "user_disabled" }, { ok: false, error: "user_disabled" }],
      [429, { error: "rate_limited" }, { ok: false, error: "rate_limited" }],
      [503, { error: "signup_unavailable" }, { ok: false, error: "signup_unavailable" }],
      [409, { error: "conflict", field: "name_taken" }, { ok: false, error: "malformed" }],
      [409, { error: "conflict", field: "slug_taken", slug: "acme" }, { ok: false, error: "malformed" }],
      [403, { error: "csrf_failed", field: "x" }, { ok: false, error: "malformed" }],
    ]) {
      reply = () => Response.json(body, { status });
      assert.deepEqual(await client.createOrganization(CTX, CRED, M, "Acme", "acme"), expected, JSON.stringify(body));
    }
  });
  it("prove: the Signup credential and {origin, nonce} without a CSRF member; core's complete URL for a tenant of this root", async () => {
    const done = `${TENANT}/signup/complete?p=${PROOF}`;
    reply = () => Response.json({ complete: done });
    assert.deepEqual(await client.prove(CTX, CRED, ROOT, NONCE), { ok: true, value: done });
    assert.equal(calls[0].url, `${CORE}/internal/v1/signup/prove`);
    assert.deepEqual(calls[0].body, { origin: ROOT, nonce: NONCE });
    assert.equal(calls[0].init.headers.Authorization, `Signup ${CRED}`);
    for (const c of [
      `https://a.b.example.test:543/signup/complete?p=${PROOF}`,
      `https://example.test:543/signup/complete?p=${PROOF}`,
      `https://acme.example.test:544/signup/complete?p=${PROOF}`,
      `https://acme.other.test:543/signup/complete?p=${PROOF}`,
      // G057 (T7): a slug-shaped host, another port, an extra label, a foreign root, http.
      `https://acme.example.test:543/signup/complete?p=${PROOF}`,
      `https://${LABEL_A}.example.test:544/signup/complete?p=${PROOF}`,
      `https://x.${LABEL_A}.example.test:543/signup/complete?p=${PROOF}`,
      `https://${LABEL_A}.other.test:543/signup/complete?p=${PROOF}`,
      `http://${LABEL_A}.example.test:543/signup/complete?p=${PROOF}`,
      `${TENANT}/signup/continue?p=${PROOF}`,
      `${TENANT}/signup/complete?p=${PROOF}#x`,
    ]) {
      reply = () => Response.json({ complete: c });
      assert.deepEqual(await client.prove(CTX, CRED, ROOT, NONCE), { ok: false, error: "malformed" }, c);
    }
    reply = () => Response.json({ error: "signup_context_mismatch" }, { status: 403 });
    assert.deepEqual(await client.prove(CTX, CRED, ROOT, NONCE), { ok: false, error: "signup_context_mismatch" });
  });
  it("abandon: {origin, csrf_token}; an empty object answer", async () => {
    reply = () => Response.json({});
    assert.deepEqual(await client.abandonSignup(CTX, CRED, M), { ok: true, value: null });
    assert.equal(calls[0].url, `${CORE}/internal/v1/signup/abandon`);
    assert.deepEqual(calls[0].body, { origin: ROOT, csrf_token: CSRF });
    reply = () => Response.json({ ok: true });
    assert.deepEqual(await client.abandonSignup(CTX, CRED, M), { ok: false, error: "malformed" });
  });
});

describe("tenant hops", () => {
  it("challenge: no credential, the login binding when present, {origin, ref}; the prove URL must be the root's", async () => {
    const proveUrl = `${ROOT}/signup/prove?n=${NONCE}`;
    reply = () => Response.json({ challenge: { value: CHALLENGE, max_age: 120 }, prove: proveUrl });
    assert.deepEqual(await client.challenge(CTX, { origin: TENANT, ref: REF, root: ROOT }, BINDING), {
      ok: true,
      value: { challenge: { value: CHALLENGE, maxAge: 120 }, proveUrl },
    });
    assert.equal(calls[0].url, `${CORE}/internal/v1/signup/challenge`);
    assert.deepEqual(calls[0].body, { origin: TENANT, ref: REF });
    assert.equal(calls[0].init.headers.Authorization, undefined);
    assert.equal(calls[0].init.headers["X-Login-Binding"], BINDING);
    await client.challenge(CTX, { origin: TENANT, ref: REF, root: ROOT });
    assert.equal(calls[1].init.headers["X-Login-Binding"], undefined);
    for (const bad of [
      { challenge: { value: CHALLENGE, max_age: 121 }, prove: proveUrl },
      { challenge: { value: "short", max_age: 120 }, prove: proveUrl },
      { challenge: { value: CHALLENGE, max_age: 120 }, prove: `https://www.example.test:543/signup/prove?n=${NONCE}` },
      { challenge: { value: CHALLENGE, max_age: 120 }, prove: `${TENANT}/signup/prove?n=${NONCE}` },
      { challenge: { value: CHALLENGE, max_age: 120 }, prove: `${ROOT}/signup/prove?n=${NONCE}&ref=${REF}` },
      { challenge: { value: CHALLENGE, max_age: 120 } },
    ]) {
      reply = () => Response.json(bad);
      assert.deepEqual(await client.challenge(CTX, { origin: TENANT, ref: REF, root: ROOT }), { ok: false, error: "malformed" }, JSON.stringify(bad));
    }
    reply = () => Response.json({ error: "continuation_invalid" }, { status: 404 });
    assert.deepEqual(await client.challenge(CTX, { origin: TENANT, ref: REF, root: ROOT }), { ok: false, error: "continuation_invalid" });
  });
  it("complete: no credential, the binding, {origin, challenge, proof}; the ordinary login answer", async () => {
    reply = () => Response.json({ authorization_url: "https://auth.example.test/authorize?x=1", binding: { value: BINDING, max_age: 600 } });
    assert.deepEqual(await client.complete(CTX, { origin: TENANT, challenge: CHALLENGE, proof: PROOF }, BINDING), {
      ok: true,
      value: { authorizationUrl: "https://auth.example.test/authorize?x=1", binding: { value: BINDING, maxAge: 600 } },
    });
    assert.equal(calls[0].url, `${CORE}/internal/v1/signup/complete`);
    assert.deepEqual(calls[0].body, { origin: TENANT, challenge: CHALLENGE, proof: PROOF });
    assert.equal(calls[0].init.headers.Authorization, undefined);
    reply = () => Response.json({ authorization_url: "javascript:alert(1)", binding: { value: BINDING, max_age: 600 } });
    assert.deepEqual(await client.complete(CTX, { origin: TENANT, challenge: CHALLENGE, proof: PROOF }), { ok: false, error: "malformed" });
    reply = () => Response.json({ error: "continuation_invalid" }, { status: 404 });
    assert.deepEqual(await client.complete(CTX, { origin: TENANT, challenge: CHALLENGE, proof: PROOF }), { ok: false, error: "continuation_invalid" });
  });
});

describe("auth callback", () => {
  const input = { origin: ROOT, rawQuery: "code=c&state=s" };
  it("the root's 202 signup credential only (≤ 1800 s, next /signup/details); never a session", async () => {
    reply = () => Response.json({ signup: { credential: { value: CRED, max_age: 1800 }, next: "/signup/details" } }, { status: 202 });
    assert.deepEqual(await client.authCallback(CTX, input, BINDING), { ok: true, value: { kind: "signup", credential: { value: CRED, maxAge: 1800 } } });
    assert.equal(calls[0].url, `${CORE}/internal/v1/auth/callback`);
    assert.deepEqual(calls[0].body, { origin: ROOT, raw_query: "code=c&state=s" });
    assert.equal(calls[0].init.headers.Authorization, undefined);
    assert.equal(calls[0].init.headers["X-Login-Binding"], BINDING);
    for (const bad of [
      { signup: { credential: { value: CRED, max_age: 1801 }, next: "/signup/details" } },
      { signup: { credential: { value: CRED, max_age: 1800 }, next: "/" } },
      { signup: { credential: { value: CRED, max_age: 1800 } } },
      { signup: { credential: { value: CRED, max_age: 1800 }, next: "/signup/details" }, session: { value: CRED, max_age: 60 } },
    ]) {
      reply = () => Response.json(bad, { status: 202 });
      assert.deepEqual(await client.authCallback(CTX, input), { ok: false, error: "malformed" }, JSON.stringify(bad));
    }
  });
  it("the pinned tenant callback's tenant_created session (return_to / only), G058's answers, and the closed refusals", async () => {
    const t = { origin: TENANT, rawQuery: "code=c&state=s" };
    reply = () => Response.json({ outcome: "tenant_created", return_to: "/", session: { value: CRED, max_age: 28800 } });
    assert.deepEqual(await client.authCallback(CTX, t, BINDING, "X".repeat(43)), {
      ok: true,
      value: { kind: "session", outcome: "tenant_created", returnTo: "/", session: { value: CRED, maxAge: 28800 } },
    });
    assert.equal(calls[0].init.headers.Authorization, `Session ${"X".repeat(43)}`);
    reply = () => Response.json({ outcome: "tenant_created", return_to: "/account", session: { value: CRED, max_age: 28800 } });
    assert.deepEqual(await client.authCallback(CTX, t), { ok: false, error: "malformed" });
    reply = () => Response.json({ outcome: "signed_in", return_to: "/account", session: { value: CRED, max_age: 28800 } });
    assert.equal((await client.authCallback(CTX, t)).value.outcome, "signed_in");
    reply = () => Response.json({ outcome: "already_signed_in", return_to: "/" });
    assert.deepEqual(await client.authCallback(CTX, t), { ok: true, value: { kind: "already_signed_in", returnTo: "/" } });
    reply = () => Response.json({ outcome: "admitted", next: "sign_in" });
    assert.deepEqual(await client.authCallback(CTX, t), { ok: true, value: { kind: "admitted", outcome: "admitted" } });
    reply = () => Response.json({ enrollment: { status: "pending", credential: { value: CRED, max_age: 1800 } } }, { status: 202 });
    assert.deepEqual(await client.authCallback(CTX, { ...t, invitation: { invitationId: ORG, token: "T".repeat(43) } }), {
      ok: true,
      value: { kind: "enrollment", status: "pending", credential: { value: CRED, maxAge: 1800 } },
    });
    assert.deepEqual(calls.at(-1).body, { origin: TENANT, raw_query: "code=c&state=s", invitation_id: ORG, invitation_token: "T".repeat(43) });
    for (const code of ["signup_user_mismatch", "rate_limited", "login_canceled", "invitation_mismatch", "host_changed"]) {
      reply = () => Response.json({ error: code }, { status: 403 });
      assert.deepEqual(await client.authCallback(CTX, t), { ok: false, error: code });
    }
    reply = () => Response.json({ error: "continuation_invalid" }, { status: 404 });
    assert.deepEqual(await client.authCallback(CTX, t), { ok: false, error: "malformed" });
  });
});

describe("origins, hop queries and root resolution", () => {
  // G057 (Plan r5 §3.3, §5.2 T7; D-5): the structural validator of core's target origin. Nothing builds an origin from a
  // slug any more (tenantOriginOf is gone); the validator accepts exactly one generated label under the root, same port.
  it("tenant origins are one generated platform label under the root, same port; never a slug's host", () => {
    assert.equal("tenantOriginOf" in client, false, "no slug-derived origin builder");
    assert.equal(client.isTenantOriginUnder(ROOT, TENANT), true);
    assert.equal(client.isTenantOriginUnder(ROOT, `https://${LABEL_B}.example.test:543`), true);
    assert.equal(client.isTenantOriginUnder(ROOT, TENANT, "acme"), true, "an unrelated slug does not matter");
    assert.equal(client.isTenantOriginUnder(ROOT, `https://${FRIENDLY}.example.test:543`), true, "a friendly generated label");
    assert.equal(client.isTenantOriginUnder(ROOT, `https://${FRIENDLY}.example.test:543`, "acme"), true);
    assert.equal(client.isTenantOriginUnder(ROOT, `https://${FRIENDLY}.example.test:543`, FRIENDLY), false, "a slug equal to the label never routes");
    for (const bad of ["gusty-flowerpot-264", "gusty-flowerpot-26430", "gusty-2643", "gusty-flower-pot-2643", "Gusty-flowerpot-2643", "gusty-flowerpot-abcd", "gu-flowerpot-2643", "-flowerpot-2643"]) {
      assert.equal(client.isTenantOriginUnder(ROOT, `https://${bad}.example.test:543`), false, bad);
    }
    assert.equal(client.isTenantOriginUnder("https://example.test", `https://${LABEL_A}.example.test`), true, "a root without a port");
    for (const [o, why] of [
      [`https://${LABEL_A}.example.test:544`, "other port"],
      [`https://${LABEL_A}.example.test`, "port dropped"],
      [`https://${LABEL_A}.example.test:5430`, "port extended"],
      [`https://x.${LABEL_A}.example.test:543`, "extra label"],
      [`https://${LABEL_A}.x.example.test:543`, "extra label below the label"],
      [`http://${LABEL_A}.example.test:543`, "http"],
      [ROOT, "the root itself"],
      ["https://acme.example.test:543", "a slug-shaped (pre-G057) host"],
      ["https://org-a.example.test:543", "a hyphenated label"],
      [`https://${LABEL_A.toUpperCase()}.example.test:543`, "upper case"],
      [`https://${LABEL_A.slice(1)}.example.test:543`, "25 characters"],
      [`https://${LABEL_A}a.example.test:543`, "27 characters"],
      [`https://${LABEL_A.slice(2)}01.example.test:543`, "digits outside base32"],
      [`https://${LABEL_A}example.test:543`, "suffix confusion"],
      [`https://${LABEL_A}.example.test.evil:543`, "foreign root"],
      [`https://user@${LABEL_A}.example.test:543`, "userinfo"],
      [`${TENANT}/`, "path"],
      [`${TENANT}?x=1`, "query"],
      [`${TENANT}#x`, "fragment"],
    ]) {
      assert.equal(client.isTenantOriginUnder(ROOT, o), false, why);
    }
    // A slug-equal host is refused even when the slug has a label's shape.
    assert.equal(client.isTenantOriginUnder(ROOT, `https://${LABEL_B}.example.test:543`, LABEL_B), false, "slug-equal host");
    assert.equal(client.isTenantOriginUnder(ROOT, TENANT, "acme", LABEL_A), false, "slug-equal host (second slug)");
    assert.equal(client.isTenantOriginUnder("http://example.test:543", `https://${LABEL_A}.example.test:543`), false, "an http root");
  });
  it("a hop query is exactly one named parameter holding a 43-character secret", () => {
    assert.equal(client.hopSecret(`ref=${REF}`, "ref"), REF);
    for (const q of ["", `ref=${REF}&ref=${REF}`, `ref=${REF}&x=1`, `n=${REF}`, `ref=${REF.slice(1)}`, `ref=${REF}=`, `ref=${"%41".repeat(43)}`, `ref=${REF}&${"p".repeat(300)}`]) {
      assert.equal(client.hopSecret(q, "ref"), undefined, q.slice(0, 60));
    }
  });
  it("the root routes resolve only at the root itself (never www, a tenant or the loopback profile)", async () => {
    reply = () => Response.json({ kind: "public_entry", signup_origin: ROOT, root_page: "signup_redirect" });
    assert.deepEqual(await client.rootOfRootRequest(CTX, ROOT), { ok: true, root: ROOT });
    assert.equal(calls[0].url, `${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(ROOT)}`);
    assert.notEqual(calls[0].init.headers["X-Request-Id"], "req-signup-1", "the lookup carries its own request id");
    assert.deepEqual(await client.rootOfRootRequest(CTX, "https://www.example.test:543"), { ok: false, error: "not_found" });
    reply = () => Response.json({ kind: "platform", organization_id: ORG });
    assert.deepEqual(await client.rootOfRootRequest(CTX, TENANT), { ok: false, error: "not_found" });
    reply = () => Response.json({ error: "x" }, { status: 503 });
    assert.deepEqual(await client.rootOfRootRequest(CTX, ROOT), { ok: false, error: "unavailable" });
    process.env.IDENTITY_COOKIE_PROFILE = "loopback-http";
    const before = calls.length;
    assert.deepEqual(await client.rootOfRootRequest(CTX, "http://localhost:3100"), { ok: false, error: "not_found" });
    assert.equal(calls.length, before);
  });
  it("the tenant hops derive the root from configuration and the request's port, confirmed by core", async () => {
    reply = () => Response.json({ kind: "public_entry", signup_origin: ROOT, root_page: "signup_redirect" });
    assert.deepEqual(await client.rootOfTenantRequest(CTX, TENANT), { ok: true, root: ROOT });
    assert.equal(calls[0].url, `${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(ROOT)}`);
    reply = () => Response.json({ kind: "public_entry", signup_origin: "https://example.test:544", root_page: "signup_redirect" });
    assert.deepEqual(await client.rootOfTenantRequest(CTX, TENANT), { ok: false, error: "not_found" });
    const before = calls.length;
    for (const o of [ROOT, "https://portal.acme.test:543", "https://a.b.example.test:543", "https://acme.example.test:543", "https://org-a.example.test:543", `https://x.${LABEL_A}.example.test:543`]) {
      assert.deepEqual(await client.rootOfTenantRequest(CTX, o), { ok: false, error: "not_found" }, o);
    }
    assert.equal(calls.length, before, "no lookup for an origin that is not a platform tenant of the configured root");
  });
  it("publicEntryOf reports the root and www for the landing, nothing elsewhere", async () => {
    reply = () => Response.json({ kind: "public_entry", signup_origin: ROOT, root_page: "signup_redirect" });
    assert.deepEqual(await client.publicEntryOf(CTX, ROOT), { root: true, signupOrigin: ROOT });
    assert.deepEqual(await client.publicEntryOf(CTX, "https://www.example.test:543"), { root: false, signupOrigin: ROOT });
    reply = () => Response.json({ kind: "platform", organization_id: ORG });
    assert.equal(await client.publicEntryOf(CTX, TENANT), undefined);
    assert.equal(await client.publicEntryOf(CTX, undefined), undefined);
  });
  it("refusal pages and notices are fixed text", async () => {
    assert.equal(client.refusedRoot("not_found").status, 404);
    assert.equal(client.refusedRoot("unavailable").status, 503);
    assert.equal(client.formExpired().status, 403);
    assert.equal(client.noticeOf(client.DETAILS_NOTICES, "continue_failed"), client.DETAILS_NOTICES.continue_failed);
    assert.equal(client.noticeOf(client.DETAILS_NOTICES, "<script>"), undefined);
    assert.equal(client.noticeOf(client.SIGNUP_NOTICES, "toString"), undefined);
    assert.deepEqual(client.fieldErrorOf("slug_taken"), { field: "slug", message: client.DETAILS_NOTICES.slug_taken });
    assert.equal(client.fieldErrorOf("continue_failed"), undefined);
  });
  // G062 bounded security correction to Plan r5 §4.4 (Codex Round 15 decision 2): the provisioned-state recovery view is
  // a pure decision; rendering it calls nothing, and Try again is offered only for a record naming the created address.
  it("recoveryView: a GET never calls core; Try again only for a record of the created address", async () => {
    const created = { slug: "acme" };
    const record = { name: "Acme Canary Works", slug: "acme" };
    assert.deepEqual(client.recoveryView(undefined, record, created), { retryable: true, notice: undefined });
    assert.deepEqual(client.recoveryView("continue_failed", record, created), { retryable: true, notice: client.DETAILS_NOTICES.continue_failed });
    assert.deepEqual(client.recoveryView("intent_used", record, created), { retryable: true, notice: client.DETAILS_NOTICES.intent_used });
    for (const [form, organization] of [
      [undefined, created],
      [{ name: "Acme Canary Works", slug: "other" }, created],
      [{ name: "Acme Canary Works", slug: "" }, created],
      [record, null],
    ]) {
      assert.deepEqual(client.recoveryView(undefined, form, organization), { retryable: false, notice: client.DETAILS_NOTICES.retry_unavailable });
      assert.equal(client.recoveryView("continue_failed", form, organization).retryable, false);
    }
    assert.equal(client.recoveryView("<script>", record, created).notice, undefined);
    assert.equal(calls.length, 0, "the recovery decision makes no core request");
  });
});
