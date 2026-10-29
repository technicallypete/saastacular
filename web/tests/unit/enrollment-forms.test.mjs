// Unit tests for the G058 enrollment form routes and the modified callback (Plan r3 §1.6, ENR-20; selected-design
// addendum §A.5): /join/start (login binding + one-use join cookie in one 303; the token never in a Location), the
// applicant routes (enrollment credential only; a confirmed admission clears the enrollment cookie and never sets a
// session; an unknown outcome is never a claimed success), the management action (closed actions; enrollment_not_ready's
// generic copy) and the callback's cookie contract (202 sets only the enrollment cookie; the join cookie is forwarded in
// the body and cleared on every outcome). fetch is a recorder; nothing leaves the process.
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

const { cookieName } = await import("../../lib/session.ts");
const startRoute = await import("../../app/join/start/route.ts");
const requestRoute = await import("../../app/join/request/route.ts");
const admitRoute = await import("../../app/join/admit/route.ts");
const abandonRoute = await import("../../app/join/abandon/route.ts");
const actionRoute = await import("../../app/organization/enrollment/action/route.ts");
const callbackRoute = await import("../../app/auth/callback/route.ts");

const CORE = "http://core.internal:8180";
const HOST = "org-c.localhost:543";
const ORIGIN = `https://${HOST}`;
const SECRET = "S".repeat(43);
const ENROLL = "E".repeat(43);
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

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "enrollment_unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), headers: init.headers, body: init.body === undefined ? undefined : JSON.parse(init.body) });
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
function post(path, body, cookie) {
  return new Request(`${ORIGIN}${path}`, {
    method: "POST",
    headers: { host: HOST, origin: ORIGIN, "content-type": "application/x-www-form-urlencoded", ...(cookie === undefined ? {} : { cookie }) },
    body,
  });
}
const cookies = (res) => res.headers.getSetCookie();
const enrollCookie = `${cookieName("enrollment", "https")}=${ENROLL}`;

describe("POST /join/start", () => {
  it("sets the login binding and the one-use join cookie in one 303; the token is never in the Location", async () => {
    reply = () => Response.json({ authorization_url: "https://auth.example.test/authorize?s=1", binding: { value: SECRET, max_age: 600 } });
    const res = await startRoute.POST(post("/join/start", form({ invitation_id: ID(5), invitation_token: TOKEN })));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "https://auth.example.test/authorize?s=1");
    assert.ok(!res.headers.get("location").includes(TOKEN));
    const set = cookies(res);
    assert.ok(set.some((c) => c.startsWith(`__Host-login=${SECRET};`)));
    assert.ok(set.some((c) => c === `__Host-join=${ID(5)}.${TOKEN}; Path=/; Max-Age=600; HttpOnly; SameSite=Lax; Secure`));
    assert.ok(set.some((c) => c.startsWith("__Host-enroll=; Path=/; Max-Age=0")));
    assert.deepEqual(calls[0].body, { origin: ORIGIN, intent: "enroll", invitation_id: ID(5) });
  });

  it("request flag; no invitation clears any join cookie; half an invitation or a repeated field is refused before core", async () => {
    reply = () => Response.json({ authorization_url: "https://auth.example.test/a", binding: { value: SECRET, max_age: 600 } });
    const res = await startRoute.POST(post("/join/start", form({ request: "1" })));
    assert.deepEqual(calls[0].body, { origin: ORIGIN, intent: "enroll", request: true });
    assert.ok(cookies(res).some((c) => c.startsWith("__Host-join=; Path=/; Max-Age=0")));
    for (const body of [form({ invitation_id: ID(5) }), form({ invitation_token: TOKEN }), form([["request", "1"], ["request", "1"]]), form({ request: "yes" })]) {
      const r = await startRoute.POST(post("/join/start", body));
      assert.equal(r.status, 400, body);
    }
    assert.equal(calls.length, 1);
    assert.equal((await startRoute.GET()).status, 405);
  });
});

describe("applicant routes", () => {
  it("request: the enrollment credential only, 303 to the status notice", async () => {
    reply = () => Response.json({ request: OWN });
    const res = await requestRoute.POST(post("/join/request", form({ csrf_token: CSRF }), enrollCookie));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "/join/status?status=requested");
    assert.equal(calls[0].headers.Authorization, `Enrollment ${ENROLL}`);
    assert.deepEqual(calls[0].body, { origin: ORIGIN, csrf_token: CSRF });
  });

  it("admit: a confirmed admission clears the enrollment cookie and sets no session", async () => {
    reply = () => Response.json({ outcome: "admitted", next: "sign_in" });
    const res = await admitRoute.POST(post("/join/admit", form({ csrf_token: CSRF, kind: "request", request_id: ID(1) }), enrollCookie));
    assert.equal(res.headers.get("location"), "/join/status?status=admitted");
    const set = cookies(res);
    assert.deepEqual(set, ["__Host-enroll=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure"]);
    assert.ok(!set.some((c) => c.startsWith("__Host-sid=")));
  });

  it("refusals map to closed notices; an unknown outcome never claims success; a dead credential is cleared", async () => {
    const cases = [
      [() => Response.json({ error: "conflict", field: "approval_revoked" }, { status: 409 }), "/join/status?status=approval_revoked"],
      [() => Response.json({ error: "enrollment_closed" }, { status: 403 }), "/join/status?status=closed"],
      [() => Response.json({ error: "enrollment_unavailable" }, { status: 503 }), "/join/status?status=unknown"],
      [() => new Response("oops", { status: 200 }), "/join/status?status=unknown"],
      [() => Response.json({ error: "unauthenticated" }, { status: 401 }), "/join/status?status=expired"],
    ];
    for (const [r, location] of cases) {
      reply = r;
      const res = await admitRoute.POST(post("/join/admit", form({ csrf_token: CSRF, kind: "open" }), enrollCookie));
      assert.equal(res.headers.get("location"), location);
    }
    const noCookie = await requestRoute.POST(post("/join/request", form({ csrf_token: CSRF })));
    assert.equal(noCookie.headers.get("location"), "/join/status?status=expired");
    assert.equal((await admitRoute.POST(post("/join/admit", form({ csrf_token: CSRF, kind: "member" }), enrollCookie))).status, 400);
    assert.equal((await requestRoute.POST(post("/join/request", form({ csrf_token: CSRF, kind: "open" }), enrollCookie))).status, 400);
  });

  it("abandon clears the enrollment cookie and returns to /join", async () => {
    reply = () => Response.json({});
    const res = await abandonRoute.POST(post("/join/abandon", form({ csrf_token: CSRF }), enrollCookie));
    assert.equal(res.headers.get("location"), "/join?status=abandoned");
    assert.ok(cookies(res).some((c) => c.startsWith("__Host-enroll=; Path=/; Max-Age=0")));
  });
});

describe("POST /organization/enrollment/action", () => {
  const sid = `${cookieName("session", "https")}=${SECRET}`;
  it("each closed action reaches its route with the session scheme", async () => {
    reply = (url) =>
      url.endsWith("/policy")
        ? Response.json({ mode: "request_access", open_ready: true, history_basis: "created_after_ledger" })
        : Response.json({ request: { ...OWN, user_id: ID(9), display_name: null, status: "approved", decided_at: "2026-10-18T11:00:00Z" } });
    assert.equal((await actionRoute.POST(post("/organization/enrollment/action", form({ csrf_token: CSRF, action: "policy", mode: "request_access" }), sid))).headers.get("location"), "/organization/enrollment?status=policy_changed");
    assert.equal((await actionRoute.POST(post("/organization/enrollment/action", form({ csrf_token: CSRF, action: "approve", request_id: ID(1) }), sid))).headers.get("location"), "/organization/enrollment?status=approved");
    assert.deepEqual(calls[1].body, { origin: ORIGIN, csrf_token: CSRF, decision: "approve" });
    assert.equal(calls[1].headers.Authorization, `Session ${SECRET}`);
  });

  it("enrollment_not_ready shows the generic copy; malformed shapes are refused before core", async () => {
    reply = () => Response.json({ error: "conflict", field: "enrollment_not_ready" }, { status: 409 });
    const res = await actionRoute.POST(post("/organization/enrollment/action", form({ csrf_token: CSRF, action: "policy", mode: "open" }), sid));
    assert.equal(res.status, 409);
    assert.match(await res.text(), /Open signup is unavailable until the platform operator has reviewed membership history/);
    const before = calls.length;
    for (const fields of [{ csrf_token: CSRF, action: "approve" }, { csrf_token: CSRF, action: "grant", request_id: ID(1) }, { action: "policy", mode: "open" }]) {
      assert.equal((await actionRoute.POST(post("/organization/enrollment/action", form(fields), sid))).status, 400, JSON.stringify(fields));
    }
    assert.equal(calls.length, before);
  });
});

describe("GET /auth/callback (enrollment answers)", () => {
  function callbackRequest(cookie) {
    return new Request(`${ORIGIN}/auth/callback?code=c&state=s`, { headers: { host: HOST, cookie } });
  }
  it("202 sets only the enrollment cookie; the join intent is forwarded in the body and the join cookie cleared", async () => {
    reply = (url) =>
      url.endsWith("/internal/v1/auth/callback")
        ? Response.json({ enrollment: { status: "pending", credential: { value: ENROLL, max_age: 1800 } } }, { status: 202 })
        : Response.json({ kind: "platform", organization_id: ID(7) });
    const res = await callbackRoute.GET(callbackRequest(`__Host-login=${SECRET}; __Host-join=${ID(5)}.${TOKEN}`));
    const cb = calls.find((c) => c.url.endsWith("/internal/v1/auth/callback"));
    assert.ok(calls[0].url.startsWith(`${CORE}/internal/v1/hosts/resolve?`));
    assert.notEqual(cb, undefined);
    assert.deepEqual(cb.body, { origin: ORIGIN, raw_query: "code=c&state=s", invitation_id: ID(5), invitation_token: TOKEN });
    assert.equal(res.headers.get("location"), "/join/status?status=pending");
    const set = cookies(res);
    assert.ok(set.some((c) => c === `__Host-enroll=${ENROLL}; Path=/; Max-Age=1800; HttpOnly; SameSite=Lax; Secure`));
    assert.ok(set.some((c) => c.startsWith("__Host-join=; Path=/; Max-Age=0")));
    assert.ok(!set.some((c) => c.startsWith("__Host-sid=")));
  });

  // Codex preliminary correction 1: early refusals on a classified non-auth host clear a present one-use join intent;
  // the shared auth host and an unclassified host never read or write a cookie.
  const JOINED = `__Host-login=${SECRET}; __Host-join=${ID(5)}.${TOKEN}`;
  const PAD = `code=c&state=s&pad=${"p".repeat(8 * 1024)}`;
  const resolving = (answer) => (url) =>
    url.includes("/internal/v1/hosts/resolve?") ? answer() : Response.json({ error: "relay_refused" }, { status: 404 });
  const callbackAt = (query, cookie) => new Request(`${ORIGIN}/auth/callback?${query}`, { headers: { host: HOST, cookie } });
  const noTokenOut = (res) => {
    for (const v of [res.headers.get("location") ?? "", ...cookies(res)]) assert.ok(!v.includes(TOKEN) && !v.includes(SECRET));
  };

  it("an oversized query on a tenant host clears the join intent and reaches no callback", async () => {
    reply = resolving(() => Response.json({ kind: "platform", organization_id: ID(7) }));
    const res = await callbackRoute.GET(callbackAt(PAD, JOINED));
    assert.equal(res.status, 400);
    assert.deepEqual(cookies(res).map((c) => c.split(";")[0]), ["__Host-join="]);
    assert.ok(cookies(res)[0].includes("Max-Age=0"));
    assert.ok(!calls.some((c) => c.url.endsWith("/internal/v1/auth/callback")));
    assert.ok(!calls.some((c) => c.url.includes("pad=")));
    noTokenOut(res);
  });

  it("a classified non-routing host (reserved) clears the join intent and reaches no callback", async () => {
    reply = resolving(() => Response.json({ kind: "reserved" }));
    const res = await callbackRoute.GET(callbackAt("code=c&state=s", JOINED));
    assert.equal(res.headers.get("location"), "/sign-in?status=failed");
    assert.deepEqual(cookies(res).map((c) => c.split(";")[0]), ["__Host-join="]);
    assert.ok(!calls.some((c) => c.url.endsWith("/internal/v1/auth/callback")));
    noTokenOut(res);
  });

  // G062 (Plan r5 §1.2, §4.4; scope amendment 04): a public entry that is not the canonical root (its own origin differs
  // from core's signup_origin) is the fixed 404 page; the join intent is still cleared and no callback is reached.
  it("a non-root public entry is the fixed 404, clears the join intent and reaches no callback", async () => {
    reply = resolving(() => Response.json({ kind: "public_entry", signup_origin: "https://example.test", root_page: "signup_redirect" }));
    const res = await callbackRoute.GET(callbackAt("code=c&state=s", JOINED));
    assert.equal(res.status, 404);
    assert.equal(res.headers.get("location"), null);
    assert.deepEqual(cookies(res).map((c) => c.split(";")[0]), ["__Host-join="]);
    assert.ok(cookies(res)[0].includes("Max-Age=0"));
    assert.ok(!calls.some((c) => c.url.endsWith("/internal/v1/auth/callback")));
    noTokenOut(res);
  });

  it("the shared auth host and an unclassified host set no cookie, oversized query included", async () => {
    const cases = [
      [() => Response.json({ kind: "auth" }), "code=c&state=s"],
      [() => Response.json({ kind: "auth" }), PAD],
      [() => Response.json({ error: "unknown_host" }, { status: 404 }), "code=c&state=s"],
      [() => Response.json({ error: "unknown_host" }, { status: 404 }), PAD],
      [() => Response.json({ error: "unavailable" }, { status: 503 }), PAD],
    ];
    for (const [answer, query] of cases) {
      calls = [];
      reply = resolving(answer);
      const res = await callbackRoute.GET(callbackAt(query, JOINED));
      assert.deepEqual(cookies(res), []);
      assert.ok(!calls.some((c) => c.url.endsWith("/internal/v1/auth/callback")));
      noTokenOut(res);
    }
  });
});
