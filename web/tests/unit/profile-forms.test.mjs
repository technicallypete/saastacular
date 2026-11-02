// Unit tests for the G074 profile forms and routes (Plan r3.1 §3.7, §3.10, §5 A7): the closed profile_status table and
// its mapping of every refusal (the three avatar HOLD reasons are one avatar_hold; no reason, Retry-After or query value
// is reflected), the name form and the multipart avatar form reduced to raw bytes, the form-route order (method,
// encoding, exact Origin, bounded body, session) before core, the 303 back to /account#profile, the fixed refusal pages,
// and the two read routes: /api/profiles/[id] answers exactly {user_id, name, avatar} (401/404/503, no-store) and
// /profiles/[id]/avatar proxies only the reviewed headers (404/503 no-store, never a cached refusal). fetch is a
// recorder; nothing leaves the process. The resolve hook maps lib siblings and the routes' @/lib aliases to .ts, as
// labels-forms.test.mjs does.
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

const pf = await import("../../lib/profile.ts");
const { cookieName } = await import("../../lib/session.ts");
const updateRoute = await import("../../app/account/profile/update/route.ts");
const avatarRoute = await import("../../app/account/profile/avatar/route.ts");
const removeRoute = await import("../../app/account/profile/avatar/remove/route.ts");
const apiRoute = await import("../../app/api/profiles/[id]/route.ts");
const imageRoute = await import("../../app/profiles/[id]/avatar/route.ts");

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
const SELF = ID(1);
const OTHER = ID(2);
const OWN = { user_id: SELF, name: "Ada Override", effective_name: "Ada Override", avatar: { version: 2, content_type: "image/png" }, version: 4 };
const COOKIE = () => `${cookieName("session", "https")}=${SECRET}`;
const back = (status) => `/account?profile_status=${status}#profile`;

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
    calls.push({
      url: String(url),
      method: init.method,
      headers: init.headers,
      body: raw === undefined ? undefined : typeof raw === "string" ? JSON.parse(raw) : new Uint8Array(raw),
    });
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

const json = (status, body, headers = {}) => () => Response.json(body, { status, headers });
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
async function multipart(parts) {
  const fd = new FormData();
  for (const [name, value] of parts) fd.append(name, value);
  const r = new Request("http://form.test/", { method: "POST", body: fd });
  return { type: r.headers.get("content-type"), body: new Uint8Array(await r.arrayBuffer()) };
}
const at = (res) => [res.status, res.headers.get("location"), res.headers.get("cache-control")];

describe("profile form helpers", () => {
  it("has exactly the nine closed profile statuses and returns to the Profile section with one of them", () => {
    assert.deepEqual(pf.PROFILE_STATUSES, [
      "profile_saved",
      "avatar_set",
      "avatar_removed",
      "stale_version",
      "invalid_name",
      "invalid_image",
      "payload_too_large",
      "storage_unavailable",
      "avatar_hold",
    ]);
    for (const status of pf.PROFILE_STATUSES) assert.deepEqual(at(pf.profileBack(status)), [303, back(status), "no-store"]);
  });

  it("maps each refusal to its closed status, the three HOLD reasons to avatar_hold, the form-layer refusals to none", () => {
    const of = (error, extra = {}) => pf.profileStatusOf({ ok: false, error, ...extra });
    assert.equal(of("version_conflict"), "stale_version");
    assert.equal(of("invalid_request", { field: "name" }), "invalid_name");
    assert.equal(of("invalid_request", { field: "expected_version" }), undefined);
    assert.equal(of("invalid_input", { field: "avatar_too_large" }), "payload_too_large");
    for (const field of ["avatar_type", "avatar_dimensions", "avatar_malformed", "normalized_too_large"]) {
      assert.equal(of("invalid_input", { field }), "invalid_image", field);
    }
    assert.equal(of("invalid_input", { field: "avatar_size" }), undefined);
    assert.equal(of("payload_too_large"), "payload_too_large");
    for (const reason of ["avatar_cleanup_pending", "avatar_upload_in_progress", "commit_outcome_unknown"]) {
      assert.equal(of("unavailable", { reason }), "avatar_hold", reason);
    }
    assert.equal(of("unavailable"), "storage_unavailable");
    assert.equal(of("unavailable", { reason: "store_unavailable" }), "storage_unavailable");
    for (const error of ["timeout", "unreachable", "malformed"]) assert.equal(of(error), "storage_unavailable", error);
    for (const error of ["csrf_failed", "unauthenticated", "origin_rejected", "forbidden", "not_available", "not_found", "unknown_host", "ambiguous_parameter"]) {
      assert.equal(of(error), undefined, error);
    }
  });

  it("answers the session, CSRF and context refusals with fixed pages that never claim a change", async () => {
    for (const [error, status, text] of [
      ["csrf_failed", 403, "Form expired"],
      ["unauthenticated", 401, "Not signed in"],
      ["origin_rejected", 403, "Request refused"],
      ["forbidden", 403, "Not available"],
      ["not_found", 404, "Not found"],
      ["ambiguous_parameter", 400, "Request not understood"],
    ]) {
      const res = pf.profileAnswer({ ok: false, error }, "profile_saved");
      assert.equal(res.status, status, error);
      assert.equal(res.headers.get("cache-control"), "no-store");
      assert.ok((await res.text()).includes(text), error);
    }
    assert.deepEqual(at(pf.profileAnswer({ ok: true, value: {} }, "avatar_removed")), [303, back("avatar_removed"), "no-store"]);
  });

  it("reads the name form: one name (empty clears), the CSRF token and a version in the grammar", () => {
    const read = (fields) => pf.readNameForm(new URLSearchParams(fields));
    assert.deepEqual(read({ csrf_token: CSRF, expected_version: "4", name: " Ada " }), { m: { csrfToken: CSRF, expectedVersion: 4 }, name: " Ada " });
    assert.deepEqual(read({ csrf_token: CSRF, expected_version: "0", name: "" }), { m: { csrfToken: CSRF, expectedVersion: 0 }, name: "" });
    for (const fields of [
      { csrf_token: CSRF, expected_version: "4" },
      [["csrf_token", CSRF], ["expected_version", "4"], ["name", "A"], ["name", "B"]],
      { csrf_token: CSRF, expected_version: "-1", name: "A" },
      { csrf_token: CSRF, expected_version: "04", name: "A" },
      { csrf_token: CSRF, name: "A" },
    ]) {
      assert.equal(read(fields), undefined, JSON.stringify(fields));
    }
  });

  it("reduces the avatar form to the raw bytes, at most 512 KiB, with exactly the three fields", async () => {
    const png = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 1, 2, 3]);
    let m = await multipart([["csrf_token", CSRF], ["expected_version", "4"], ["avatar", new Blob([png], { type: "image/svg+xml" })]]);
    const ok = await pf.parseAvatarUpload(m.type, m.body);
    assert.equal(ok.ok, true);
    assert.deepEqual(ok.fields, { csrfToken: CSRF, expectedVersion: 4 });
    assert.deepEqual([...ok.bytes], [...png]);
    m = await multipart([["csrf_token", CSRF], ["expected_version", "4"], ["avatar", new Blob([new Uint8Array(524_289)])]]);
    assert.deepEqual(await pf.parseAvatarUpload(m.type, m.body), { ok: false, error: "too_large" });
    for (const parts of [
      [["csrf_token", CSRF], ["expected_version", "4"]],
      [["csrf_token", CSRF], ["expected_version", "4"], ["avatar", new Blob([])]],
      [["csrf_token", CSRF], ["expected_version", "4"], ["avatar", "text"]],
      [["csrf_token", CSRF], ["expected_version", "4"], ["avatar", new Blob([png])], ["avatar", new Blob([png])]],
      [["csrf_token", CSRF], ["expected_version", "4"], ["avatar", new Blob([png])], ["logo", new Blob([png])]],
      [["csrf_token", CSRF], ["avatar", new Blob([png])]],
    ]) {
      m = await multipart(parts);
      assert.deepEqual(await pf.parseAvatarUpload(m.type, m.body), { ok: false, error: "bad_request" });
    }
    assert.deepEqual(await pf.parseAvatarUpload("multipart/form-data; boundary=x", new Uint8Array([1, 2, 3])), { ok: false, error: "bad_request" });
  });
});

describe("profile name and avatar-removal routes", () => {
  it("saves the name with the documented fields only and returns 303 to /account#profile with profile_saved", async () => {
    reply = json(200, OWN);
    const res = await updateRoute.POST(post("/account/profile/update", form({ csrf_token: CSRF, expected_version: "4", name: "Ada Override" })));
    assert.deepEqual(at(res), [303, back("profile_saved"), "no-store"]);
    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, `${CORE}/internal/v1/profiles/me`);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, expected_version: 4, name: "Ada Override" });
    assert.equal(calls[0].headers.Authorization, `Session ${SECRET}`);
  });

  it("returns stale_version, invalid_name and storage_unavailable, and a fixed page for a failed CSRF check", async () => {
    for (const [status, body, want] of [
      [409, { error: "version_conflict" }, "stale_version"],
      [400, { error: "invalid_request", field: "name" }, "invalid_name"],
      [503, { error: "unavailable" }, "storage_unavailable"],
    ]) {
      reply = json(status, body);
      const res = await updateRoute.POST(post("/account/profile/update", form({ csrf_token: CSRF, expected_version: "4", name: "x" })));
      assert.deepEqual(at(res), [303, back(want), "no-store"], want);
    }
    reply = json(403, { error: "csrf_failed" });
    const res = await updateRoute.POST(post("/account/profile/update", form({ csrf_token: "forged", expected_version: "4", name: "x" })));
    assert.equal(res.status, 403);
    assert.ok((await res.text()).includes("Nothing was changed."));
  });

  it("refuses before core: another method, another encoding, a foreign Origin, no session, a malformed form", async () => {
    assert.equal(updateRoute.GET().status, 405);
    assert.equal(removeRoute.PUT().status, 405);
    const fields = form({ csrf_token: CSRF, expected_version: "4", name: "x" });
    assert.equal((await updateRoute.POST(post("/account/profile/update", fields, { "content-type": "text/plain" }))).status, 415);
    assert.equal((await updateRoute.POST(post("/account/profile/update", fields, { origin: "https://attacker.localhost:543" }))).status, 403);
    assert.equal((await updateRoute.POST(post("/account/profile/update", fields, { cookie: "" }))).status, 401);
    assert.equal((await updateRoute.POST(post("/account/profile/update", form({ csrf_token: CSRF, name: "x" })))).status, 400);
    assert.equal((await removeRoute.POST(post("/account/profile/avatar/remove", fields, { origin: "https://attacker.localhost:543" }))).status, 403);
    assert.equal(calls.length, 0);
  });

  it("removes the avatar with exactly origin, csrf_token and expected_version and returns avatar_removed", async () => {
    reply = json(200, { ...OWN, avatar: null, version: 5 });
    const res = await removeRoute.POST(post("/account/profile/avatar/remove", form({ csrf_token: CSRF, expected_version: "4" })));
    assert.deepEqual(at(res), [303, back("avatar_removed"), "no-store"]);
    assert.equal(calls[0].url, `${CORE}/internal/v1/profiles/me/avatar/remove`);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF, expected_version: 4 });
  });
});

describe("profile avatar upload route", () => {
  const png = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 9, 9, 9]);
  async function upload(parts, headers = {}) {
    const m = await multipart(parts);
    return avatarRoute.POST(
      new Request(`${ORIGIN}/account/profile/avatar`, {
        method: "POST",
        headers: { host: HOST, origin: ORIGIN, cookie: COOKIE(), "content-type": m.type, ...headers },
        body: m.body,
      }),
    );
  }
  const good = [["csrf_token", CSRF], ["expected_version", "4"], ["avatar", new Blob([png])]];

  it("sends only the raw image bytes with the CSRF token, version and origin headers and returns avatar_set", async () => {
    reply = json(200, OWN);
    assert.deepEqual(at(await upload(good)), [303, back("avatar_set"), "no-store"]);
    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, `${CORE}/internal/v1/profiles/me/avatar?origin=${encodeURIComponent(ORIGIN)}`);
    assert.deepEqual(calls[0].body, png);
    assert.equal(calls[0].headers["Content-Type"], "application/octet-stream");
    assert.equal(calls[0].headers["X-CSRF-Token"], CSRF);
    assert.equal(calls[0].headers["X-Expected-Version"], "4");
    assert.equal(calls[0].headers["X-Origin"], ORIGIN);
  });

  it("maps each HOLD to avatar_hold without reflecting its reason or Retry-After, and the image refusals to their statuses", async () => {
    for (const reason of ["avatar_cleanup_pending", "avatar_upload_in_progress", "commit_outcome_unknown"]) {
      reply = json(503, { error: "unavailable", reason }, { "retry-after": "30" });
      const res = await upload(good);
      assert.deepEqual(at(res), [303, back("avatar_hold"), "no-store"], reason);
      assert.equal(res.headers.get("retry-after"), null);
    }
    for (const [status, body, want] of [
      [400, { error: "invalid_input", field: "avatar_type" }, "invalid_image"],
      [400, { error: "invalid_input", field: "avatar_dimensions" }, "invalid_image"],
      [413, { error: "payload_too_large" }, "payload_too_large"],
      [409, { error: "version_conflict" }, "stale_version"],
      [503, { error: "unavailable", reason: "compensated" }, "storage_unavailable"],
    ]) {
      reply = json(status, body);
      assert.deepEqual(at(await upload(good)), [303, back(want), "no-store"], want);
    }
  });

  it("refuses before core: a non-multipart body, a foreign Origin, an oversized file, a malformed form, no session", async () => {
    const res415 = await avatarRoute.POST(
      new Request(`${ORIGIN}/account/profile/avatar`, {
        method: "POST",
        headers: { host: HOST, origin: ORIGIN, cookie: COOKIE(), "content-type": "application/x-www-form-urlencoded" },
        body: "a=b",
      }),
    );
    assert.equal(res415.status, 415);
    assert.equal((await upload(good, { origin: "https://attacker.localhost:543" })).status, 403);
    assert.deepEqual(at(await upload([["csrf_token", CSRF], ["expected_version", "4"], ["avatar", new Blob([new Uint8Array(524_289)])]])), [
      303,
      back("payload_too_large"),
      "no-store",
    ]);
    assert.equal((await upload([["csrf_token", CSRF], ["avatar", new Blob([png])]])).status, 400);
    assert.equal((await upload(good, { cookie: "" })).status, 401);
    assert.equal(avatarRoute.GET().status, 405);
    assert.equal(calls.length, 0);
  });
});

describe("profile read routes", () => {
  const read = (id, headers) => apiRoute.GET(get(`/api/profiles/${id}`, headers), { params: Promise.resolve({ id }) });
  const image = (id, query, headers) => imageRoute.GET(get(`/profiles/${id}/avatar${query}`, headers), { params: Promise.resolve({ id }) });

  it("answers another user's profile as exactly {user_id, name, avatar}, no-store", async () => {
    reply = json(200, { user_id: OTHER, name: "Bob Example", avatar: { version: 2, content_type: "image/png" } });
    const res = await read(OTHER);
    assert.equal(res.status, 200);
    assert.equal(res.headers.get("cache-control"), "no-store");
    assert.deepEqual(await res.json(), { user_id: OTHER, name: "Bob Example", avatar: { version: 2, content_type: "image/png" } });
    assert.equal(calls[0].url, `${CORE}/internal/v1/profiles/${OTHER}?origin=${encodeURIComponent(ORIGIN)}`);
  });

  it("is 401 without a session, 404 for a malformed id or any core refusal, 503 when core cannot answer", async () => {
    assert.deepEqual([(await read(OTHER, { cookie: "" })).status, calls.length], [401, 0]);
    assert.deepEqual([(await read("not-a-user")).status, calls.length], [404, 0]);
    for (const [status, body, want] of [
      [404, { error: "not_found" }, 404],
      [403, { error: "forbidden" }, 404],
      [403, { error: "not_available" }, 404],
      [401, { error: "unauthenticated" }, 401],
      [503, { error: "unavailable" }, 503],
      [200, { user_id: OTHER, name: "Bob", avatar: null, email: "bob@example.test" }, 503],
    ]) {
      reply = json(status, body);
      const res = await read(OTHER);
      assert.equal(res.status, want, JSON.stringify(body));
      assert.equal(res.headers.get("cache-control"), "no-store");
      assert.equal(JSON.stringify(await res.json()).includes("bob@example.test"), false);
    }
    assert.equal(apiRoute.POST().status, 405);
  });

  it("proxies the current avatar with exactly the reviewed headers", async () => {
    const png = new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
    reply = () => new Response(png, { status: 200, headers: { "content-type": "image/png", "content-length": "8", etag: '"v2"', "set-cookie": "x=y" } });
    const res = await image(OTHER, "?v=2");
    assert.equal(res.status, 200);
    assert.deepEqual(Object.fromEntries(res.headers), {
      "cache-control": "private, max-age=300",
      "content-disposition": "inline",
      "content-length": "8",
      "content-type": "image/png",
      etag: '"v2"',
      "x-content-type-options": "nosniff",
    });
    assert.deepEqual(new Uint8Array(await res.arrayBuffer()), png);
    assert.equal(calls[0].url, `${CORE}/internal/v1/profiles/${OTHER}/avatar?origin=${encodeURIComponent(ORIGIN)}&v=2`);
    assert.equal(calls[0].headers.Authorization, `Session ${SECRET}`);
  });

  it("answers no session, a bad id or version and every core refusal with a no-store 404, an outage with a no-store 503", async () => {
    for (const [id, query, headers] of [
      [OTHER, "?v=2", { cookie: "" }],
      ["x", "?v=2", {}],
      [OTHER, "", {}],
      [OTHER, "?v=0", {}],
      [OTHER, "?v=2&v=3", {}],
    ]) {
      const res = await image(id, query, headers);
      assert.deepEqual([res.status, res.headers.get("cache-control")], [404, "no-store"], `${id}${query}`);
    }
    assert.equal(calls.length, 0);
    reply = json(404, { error: "not_found" });
    assert.deepEqual([(await image(OTHER, "?v=2")).status, calls.length], [404, 1]);
    reply = json(503, { error: "unavailable" });
    const res = await image(OTHER, "?v=2");
    assert.deepEqual([res.status, res.headers.get("cache-control")], [503, "no-store"]);
  });
});
