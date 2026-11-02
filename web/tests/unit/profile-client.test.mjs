// Unit tests for lib/profile-client.ts (G074 Plan r3.1 §3.5, §3.7): the exact outbound requests of the six calls
// (diagnostic headers, the session credential, JSON with the documented fields only, the raw avatar bytes with the CSRF
// token, expected version and origin as headers), the strict reading of core's answers (exact key sets: no email,
// provider identity, Organization list, object id or key is ever accepted; UUIDs; the explicit-name and avatar-version
// invariants), the closed refusal reading with its one field or reason, the three avatar HOLD reasons, and the avatar
// byte proxy's checks (type, length, ETag, bound). fetch is replaced by a recorder, so nothing leaves the process. A
// resolve hook scoped to the lib directory maps the extensionless sibling imports to .ts, as membership-client.test.mjs
// does.
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

const pc = await import("../../lib/profile-client.ts");
const { deriveContext, outboundHeaders } = await import("../../lib/request-context.ts");

const CORE = "http://core.internal:8180";
const ORIGIN = "https://org-a.localhost:543";
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const M = { origin: ORIGIN, csrfToken: CSRF, expectedVersion: 3 };
const CTX = deriveContext(new Headers({ "x-request-id": "req-profile-1", traceparent: `00-${"a".repeat(32)}-${"b".repeat(16)}-01` }));
const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };
const ID = (n) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
const SELF = ID(1);
const OTHER = ID(2);
const AVATAR = { version: 2, content_type: "image/png" };
// The own view {user_id, name, effective_name, avatar, version} and another user's view {user_id, name, avatar}.
const OWN = { user_id: SELF, name: "Ada Override", effective_name: "Ada Override", avatar: AVATAR, version: 4 };
const OWN_DEFAULT = { user_id: SELF, name: null, effective_name: "Ada Provider", avatar: null, version: 0 };
const VISIBLE = { user_id: OTHER, name: "Bob Example", avatar: AVATAR };
// A copy of o without key.
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

const answer = (status, body, headers = {}) => () =>
  new Response(typeof body === "string" ? body : JSON.stringify(body), { status, headers: { "content-type": "application/json", ...headers } });

function sent(i = 0) {
  const c = calls[i];
  const raw = c.init.body;
  const body = raw === undefined ? undefined : typeof raw === "string" ? JSON.parse(raw) : new Uint8Array(raw);
  return { url: c.url, method: c.init.method, headers: c.init.headers, body, cache: c.init.cache, redirect: c.init.redirect };
}

describe("profile client: requests", () => {
  it("reads the own profile with the origin query, the read headers, no body, no-store and no redirects", async () => {
    reply = answer(200, OWN);
    assert.deepEqual(await pc.getOwnProfile(CTX, SECRET, ORIGIN), {
      ok: true,
      value: { userId: SELF, name: "Ada Override", effectiveName: "Ada Override", avatar: { version: 2, contentType: "image/png" }, version: 4 },
    });
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/profiles/me?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(s.method, "GET");
    assert.deepEqual(s.headers, { ...outboundHeaders(CTX), Authorization: `Session ${SECRET}` });
    assert.equal(s.body, undefined);
    assert.equal(s.cache, "no-store");
    assert.equal(s.redirect, "error");
  });

  it("sets and clears the explicit name with exactly origin, csrf_token, expected_version and name", async () => {
    reply = answer(200, OWN);
    await pc.updateProfileName(CTX, SECRET, M, "  Ada Override ");
    reply = answer(200, OWN_DEFAULT);
    const cleared = await pc.updateProfileName(CTX, SECRET, M, "");
    assert.deepEqual(cleared, { ok: true, value: { userId: SELF, name: null, effectiveName: "Ada Provider", avatar: null, version: 0 } });
    for (const [i, name] of [[0, "  Ada Override "], [1, ""]]) {
      const s = sent(i);
      assert.equal(s.url, `${CORE}/internal/v1/profiles/me`);
      assert.equal(s.method, "POST");
      // The name is sent as submitted (core trims and validates it); "" clears the override.
      assert.deepEqual(s.body, { origin: ORIGIN, csrf_token: CSRF, expected_version: 3, name });
      assert.deepEqual(s.headers, { ...outboundHeaders(CTX), Authorization: `Session ${SECRET}`, "Content-Type": "application/json" });
    }
  });

  it("uploads the raw avatar bytes with the CSRF token, expected version and origin as headers only", async () => {
    reply = answer(200, OWN);
    const bytes = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 1, 2, 3]);
    assert.equal((await pc.setAvatar(CTX, SECRET, M, bytes)).ok, true);
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/profiles/me/avatar?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(s.method, "POST");
    assert.deepEqual(s.body, bytes);
    assert.deepEqual(s.headers, {
      ...outboundHeaders(CTX),
      Authorization: `Session ${SECRET}`,
      "Content-Type": "application/octet-stream",
      "X-CSRF-Token": CSRF,
      "X-Expected-Version": "3",
      "X-Origin": ORIGIN,
    });
  });

  it("refuses an avatar above 512 KiB locally, without calling core", async () => {
    assert.equal(pc.MAX_AVATAR_BYTES, 524_288);
    assert.deepEqual(await pc.setAvatar(CTX, SECRET, M, new Uint8Array(pc.MAX_AVATAR_BYTES + 1)), { ok: false, error: "payload_too_large" });
    assert.equal(calls.length, 0);
  });

  it("removes the avatar with exactly origin, csrf_token and expected_version", async () => {
    reply = answer(200, { ...OWN, avatar: null, version: 5 });
    const r = await pc.removeAvatar(CTX, SECRET, M);
    assert.equal(r.ok, true);
    assert.equal(r.value.avatar, null);
    assert.equal(sent().url, `${CORE}/internal/v1/profiles/me/avatar/remove`);
    assert.deepEqual(sent().body, { origin: ORIGIN, csrf_token: CSRF, expected_version: 3 });
  });

  it("reads another user's profile by its canonical id; a malformed id is not_found without calling core", async () => {
    reply = answer(200, VISIBLE);
    assert.deepEqual(await pc.getProfile(CTX, SECRET, ORIGIN, OTHER), {
      ok: true,
      value: { userId: OTHER, name: "Bob Example", avatar: { version: 2, contentType: "image/png" } },
    });
    assert.equal(sent().url, `${CORE}/internal/v1/profiles/${OTHER}?origin=${encodeURIComponent(ORIGIN)}`);
    for (const bad of ["x", `A${OTHER.slice(1)}`,`${OTHER}/avatar`, "../me"]) {
      assert.deepEqual(await pc.getProfile(CTX, SECRET, ORIGIN, bad), { ok: false, error: "not_found" }, bad);
    }
    assert.equal(calls.length, 1);
    // An answer about another subject than the one asked for is outside the contract.
    reply = answer(200, { ...VISIBLE, user_id: ID(9) });
    assert.deepEqual(await pc.getProfile(CTX, SECRET, ORIGIN, OTHER), { ok: false, error: "malformed" });
  });
});

describe("profile client: strict answers", () => {
  it("reads the own view: an empty presentation name is null, the explicit name is the presentation name", () => {
    assert.deepEqual(pc.readOwnProfile({ ...OWN_DEFAULT, effective_name: "" }), { userId: SELF, name: null, effectiveName: null, avatar: null, version: 0 });
    assert.equal(pc.readOwnProfile({ ...OWN, effective_name: "Someone Else" }), undefined);
    assert.equal(pc.readOwnProfile({ ...OWN, version: 0 }), undefined, "a row with an avatar has been written at least once");
  });

  it("refuses an own view with an extra or missing key, a bad id, version, name or avatar", () => {
    for (const bad of [
      { ...OWN, email: "ada@example.test" },
      { ...OWN, organizations: [] },
      { ...OWN, avatar: { ...AVATAR, object_id: ID(7) } },
      { ...OWN, avatar: { ...AVATAR, key: `avatars/${SELF}/${ID(7)}` } },
      without(OWN, "version"),
      { ...OWN, user_id: `A${SELF.slice(1)}` },
      { ...OWN, version: -1 },
      { ...OWN, version: 1.5 },
      { ...OWN, version: 2 ** 31 },
      { ...OWN, name: "x".repeat(201), effective_name: "x".repeat(201) },
      { ...OWN, avatar: { version: 0, content_type: "image/png" } },
      { ...OWN, avatar: { version: 2, content_type: "image/svg+xml" } },
      { ...OWN, avatar: { version: "2", content_type: "image/png" } },
      [],
      null,
    ]) {
      assert.equal(pc.readOwnProfile(bad), undefined, JSON.stringify(bad));
    }
  });

  it("reads another user's view with exactly user_id, name and avatar, never an address or provider identity", () => {
    assert.deepEqual(pc.readVisibleProfile({ ...VISIBLE, avatar: null, name: "" }), { userId: OTHER, name: null, avatar: null });
    for (const bad of [
      { ...VISIBLE, email: "bob@example.test" },
      { ...VISIBLE, effective_name: "Bob Example" },
      { ...VISIBLE, version: 1 },
      { ...VISIBLE, provider_subject: "user_01" },
      without(VISIBLE, "avatar"),
      { ...VISIBLE, avatar: { ...AVATAR, sha256: "ab" } },
    ]) {
      assert.equal(pc.readVisibleProfile(bad), undefined, JSON.stringify(bad));
    }
  });

  it("refuses an answer larger than the 8 KiB bound and an unparsable body", async () => {
    reply = answer(200, `{"user_id":"${SELF}","pad":"${"x".repeat(8 * 1024)}"}`);
    assert.deepEqual(await pc.getOwnProfile(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    reply = answer(200, "not json");
    assert.deepEqual(await pc.getOwnProfile(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
  });

  it("reads closed refusals with at most one closed field or reason", () => {
    assert.deepEqual(pc.readRefusal(409, { error: "version_conflict" }), { ok: false, error: "version_conflict" });
    assert.deepEqual(pc.readRefusal(400, { error: "invalid_request", field: "name" }), { ok: false, error: "invalid_request", field: "name" });
    assert.deepEqual(pc.readRefusal(400, { error: "invalid_input", field: "avatar_type" }), { ok: false, error: "invalid_input", field: "avatar_type" });
    assert.deepEqual(pc.readRefusal(503, { error: "unavailable", reason: "avatar_cleanup_pending" }), { ok: false, error: "unavailable", reason: "avatar_cleanup_pending" });
    assert.deepEqual(pc.readRefusal(404, { error: "not_found" }), { ok: false, error: "not_found" });
    for (const [status, body, want] of [
      [404, { error: "not_found", field: "x", reason: "y" }, "malformed"],
      [400, { error: "invalid_request", field: "Bad Field" }, "malformed"],
      [400, { error: "boom" }, "malformed"],
      [400, { error: "invalid_input", field: "avatar_type", detail: "png" }, "malformed"],
      [200, { error: "not_found" }, "malformed"],
      [500, { error: "boom" }, "unavailable"],
      [502, undefined, "unavailable"],
    ]) {
      assert.deepEqual(pc.readRefusal(status, body), { ok: false, error: want }, JSON.stringify([status, body]));
    }
  });

  it("recognizes exactly the three repairable avatar HOLD reasons", () => {
    assert.deepEqual(pc.HOLD_REASONS, ["avatar_cleanup_pending", "avatar_upload_in_progress", "commit_outcome_unknown"]);
    for (const reason of pc.HOLD_REASONS) assert.equal(pc.isHold({ ok: false, error: "unavailable", reason }), true, reason);
    for (const f of [
      { ok: false, error: "unavailable" },
      { ok: false, error: "unavailable", reason: "store_unavailable" },
      { ok: false, error: "version_conflict", reason: "avatar_cleanup_pending" },
      { ok: false, error: "timeout" },
    ]) {
      assert.equal(pc.isHold(f), false, JSON.stringify(f));
    }
  });

  it("carries a HOLD's closed reason through the avatar upload, never the Retry-After value", async () => {
    reply = answer(503, { error: "unavailable", reason: "avatar_upload_in_progress" }, { "retry-after": "11" });
    const r = await pc.setAvatar(CTX, SECRET, M, new Uint8Array([1]));
    assert.deepEqual(r, { ok: false, error: "unavailable", reason: "avatar_upload_in_progress" });
    assert.equal(pc.isHold(r), true);
  });

  it("reports an unreachable core as unreachable, never as a refusal", async () => {
    reply = () => {
      throw new TypeError("fetch failed");
    };
    assert.deepEqual(await pc.getOwnProfile(CTX, SECRET, ORIGIN), { ok: false, error: "unreachable" });
  });
});

describe("profile client: avatar bytes", () => {
  const png = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
  const image = (headers, body = png, status = 200) => () => new Response(body, { status, headers });
  const good = { "content-type": "image/png", "content-length": String(png.byteLength), etag: '"v2"' };

  it("fetches the current version with the read headers and returns exactly the bytes", async () => {
    reply = image(good);
    const r = await pc.fetchAvatar(CTX, SECRET, ORIGIN, OTHER, "2");
    assert.equal(r.ok, true);
    assert.equal(r.contentType, "image/png");
    assert.deepEqual(r.bytes, png);
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/profiles/${OTHER}/avatar?origin=${encodeURIComponent(ORIGIN)}&v=2`);
    assert.deepEqual(s.headers, { ...outboundHeaders(CTX), Authorization: `Session ${SECRET}` });
    assert.equal(s.cache, "no-store");
  });

  it("refuses a malformed id or version locally (not_found) and every core refusal as not_found", async () => {
    for (const [id, v] of [["x", "2"], [OTHER, "0"], [OTHER, "02"], [OTHER, "-1"], [OTHER, "2147483648"], [OTHER, "1.5"]]) {
      assert.deepEqual(await pc.fetchAvatar(CTX, SECRET, ORIGIN, id, v), { ok: false, error: "not_found" }, `${id} ${v}`);
    }
    assert.equal(calls.length, 0);
    for (const status of [401, 403, 404]) {
      reply = answer(status, { error: "not_found" });
      assert.deepEqual(await pc.fetchAvatar(CTX, SECRET, ORIGIN, OTHER, "2"), { ok: false, error: "not_found" }, String(status));
    }
  });

  it("treats a wrong type, ETag or length, an oversized body and a server error as unavailable", async () => {
    for (const headers of [
      { ...good, "content-type": "image/svg+xml" },
      { ...good, etag: '"v3"' },
      { ...good, "content-length": "9" },
      { ...good, "content-length": String(pc.MAX_AVATAR_BYTES + 1) },
      without(good, "etag"),
    ]) {
      reply = image(headers);
      assert.deepEqual(await pc.fetchAvatar(CTX, SECRET, ORIGIN, OTHER, "2"), { ok: false, error: "unavailable" }, JSON.stringify(headers));
    }
    reply = answer(503, { error: "unavailable" });
    assert.deepEqual(await pc.fetchAvatar(CTX, SECRET, ORIGIN, OTHER, "2"), { ok: false, error: "unavailable" });
  });
});
