// Unit tests for the G060 Group-administration setting form and route (Plan r0 §1.6, §4.2, D-3, D-11): the closed
// choice, the form-layer order (method, encoding, exact Origin, session) before core, the 303 to the page with the
// state core reported, the fixed refusal page for a non-owner ("Not permitted", nothing changed) and the unknown-state
// page that never claims success. fetch is a recorder; nothing leaves the process. The resolve hook maps lib siblings
// and the route's @/lib aliases to .ts, as membership-forms.test.mjs does.
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
const route = await import("../../app/organization/group-authority/action/route.ts");

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

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "membership_unavailable" }, { status: 503 });
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
const PATH = "/organization/group-authority/action";
function post(body, headers = {}) {
  return new Request(`${ORIGIN}${PATH}`, {
    method: "POST",
    headers: { host: HOST, origin: ORIGIN, "content-type": "application/x-www-form-urlencoded", cookie: `${cookieName("session", "https")}=${SECRET}`, ...headers },
    body,
  });
}
const getter = (fields) => (name) => {
  const all = new URLSearchParams(fields).getAll(name);
  return all.length > 1 ? Symbol("ambiguous") : all[0];
};

describe("readGroupAuthorityForm", () => {
  it("accepts exactly on or off once", () => {
    assert.equal(mem.readGroupAuthorityForm(getter({ enabled: "on" })), true);
    assert.equal(mem.readGroupAuthorityForm(getter({ enabled: "off" })), false);
    for (const fields of [{}, { enabled: "" }, { enabled: "true" }, { enabled: "On" }, [["enabled", "on"], ["enabled", "off"]]]) {
      assert.equal(mem.readGroupAuthorityForm(getter(fields)), undefined, JSON.stringify(fields));
    }
  });

  it("has a page message for each closed status and the refusal sentences of the G060 fields", () => {
    assert.deepEqual(Object.keys(mem.GROUP_AUTHORITY_STATUS).sort(), ["disabled", "enabled"]);
    for (const f of ["self_target", "self_escalation"]) assert.equal(typeof mem.refusalSentence(f), "string", f);
  });
});

describe("Group-authority action route", () => {
  it("posts the owner's choice to core and answers 303 to the page with the state core reported", async () => {
    reply = () => Response.json({ override_enabled: true, can_change: true });
    const res = await route.POST(post(form({ csrf_token: CSRF, enabled: "on" })));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "/organization/group-authority?status=enabled");
    assert.deepEqual(calls, [{ url: `${CORE}/internal/v1/organization/group-authority`, method: "POST", body: { origin: ORIGIN, csrf_token: CSRF, enabled: true } }]);
    calls = [];
    reply = () => Response.json({ override_enabled: false, can_change: true });
    const off = await route.POST(post(form({ csrf_token: CSRF, enabled: "off" })));
    assert.equal(off.headers.get("location"), "/organization/group-authority?status=disabled");
    assert.equal(calls[0].body.enabled, false);
  });

  it("shows a non-owner's refusal as Not permitted with nothing changed and no status redirect", async () => {
    reply = () => Response.json({ error: "forbidden" }, { status: 403 });
    const res = await route.POST(post(form({ csrf_token: CSRF, enabled: "off" })));
    assert.equal(res.status, 403);
    assert.equal(res.headers.get("location"), null);
    assert.equal(res.headers.get("cache-control"), "no-store");
    assert.match(await res.text(), /Not permitted.*Nothing was changed/s);
  });

  it("shows an outcome core may have committed as unknown, never as success", async () => {
    for (const r of [() => Response.json({ error: "membership_unavailable" }, { status: 503 }), () => new Response("{", { status: 200 }), () => Response.json({ override_enabled: false, can_change: true })]) {
      reply = r;
      const res = await route.POST(post(form({ csrf_token: CSRF, enabled: "on" })));
      assert.equal(res.status, 503);
      assert.match(await res.text(), /may or may not have completed/);
    }
  });

  it("refuses before core when the form, origin or session is wrong, and answers 405 to other methods", async () => {
    for (const [body, headers, status] of [
      [form({ csrf_token: CSRF, enabled: "maybe" }), {}, 400],
      [form({ csrf_token: CSRF }), {}, 400],
      [form({ enabled: "on" }), {}, 400],
      [`${form({ csrf_token: CSRF, enabled: "on" })}&enabled=off`, {}, 400],
      [form({ csrf_token: CSRF, enabled: "on" }), { origin: "https://evil.test" }, 403],
      [form({ csrf_token: CSRF, enabled: "on" }), { cookie: "" }, 401],
    ]) {
      const res = await route.POST(post(body, headers));
      assert.equal(res.status, status, `${body} ${JSON.stringify(headers)}`);
    }
    assert.equal(calls.length, 0);
    for (const m of ["GET", "PUT", "PATCH", "DELETE", "OPTIONS"]) assert.equal(route[m]().status, 405, m);
  });
});
