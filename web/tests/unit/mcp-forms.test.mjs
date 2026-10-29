// Unit tests for the MCP connection form routes (G063 Plan r1 §5.4, §6; r1 C4): /connections/mcp/create (an optional
// friendly name), the new /connections/mcp/rename and /connections/mcp/revoke (the closed return_to). Raw request bodies
// prove the endpoint-level form negatives: a malformed name encoding (invalid UTF-8, an encoded surrogate) or a name
// outside the rule redirects to the page's closed invalid_name status without any call to core; a repeated name is
// refused; legacy fields keep their existing decoding and ambiguity rules; no submitted name appears in a Location
// header, a status page or a log. fetch is a recorder; nothing leaves the process. Code points are built with
// String.fromCodePoint so no escape sequence in this file can be rewritten by tooling.
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

const createRoute = await import("../../app/connections/mcp/create/route.ts");
const renameRoute = await import("../../app/connections/mcp/rename/route.ts");
const revokeRoute = await import("../../app/connections/mcp/revoke/route.ts");

const CORE = "http://core.internal:8180";
const HOST = "acme.example.test:543";
const ORIGIN = `https://${HOST}`;
const CRED = "S".repeat(43);
const CSRF = "C".repeat(43);
const ID = "7f000000-0000-4000-8000-000000000001";
const COOKIE = `__Host-sid=${CRED}`;
const ENV = {
  PRODUCT_NAME: "Test Product",
  PLATFORM_BASE_DOMAIN: "example.test",
  CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test",
  CORE_INTERNAL_URL: CORE,
  IDENTITY_COOKIE_PROFILE: "https",
  APP_ENV: "test",
};
const cp = (...points) => String.fromCodePoint(...points);
const EMOJI80 = cp(0x1f600).repeat(80);
const PADDED = cp(0x3000).repeat(8) + EMOJI80 + cp(0xa0, 0x2000, 0x20);
const CANARY = "Canary Laptop";

let saved;
let calls;
let routes;
let logged;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  routes = {};
  logged = [];
  for (const level of ["log", "info", "warn", "error", "debug"]) mock.method(console, level, (...args) => logged.push(args.map(String).join(" ")));
  mock.method(globalThis, "fetch", async (url, init) => {
    const u = String(url);
    calls.push({ url: u, headers: init.headers, body: init.body === undefined ? undefined : JSON.parse(init.body) });
    const answer = routes[new URL(u).pathname];
    return answer === undefined ? Response.json({ error: "mcp_unavailable" }, { status: 503 }) : answer();
  });
});

afterEach(() => {
  mock.restoreAll();
  for (const [key, value] of Object.entries(saved)) {
    if (value === undefined) delete process.env[key];
    else process.env[key] = value;
  }
});

// A form POST with a raw body (a string is sent as its UTF-8 bytes, a Uint8Array as is); cookie: null sends none.
function post(path, body, { cookie = COOKIE, origin = ORIGIN } = {}) {
  return new Request(`${ORIGIN}${path}`, {
    method: "POST",
    headers: { host: HOST, origin, "content-type": "application/x-www-form-urlencoded", ...(cookie === null ? {} : { cookie }) },
    body: typeof body === "string" ? new TextEncoder().encode(body) : body,
  });
}
const enc = (fields) => new URLSearchParams(fields).toString();
const RENAME = "/internal/v1/mcp/connections/rename";
const CREATE = "/internal/v1/mcp/connections";
const REVOKE = "/internal/v1/mcp/connections/revoke";
const core = (path) => calls.filter((c) => c.url === `${CORE}${path}`);

// Nothing submitted as a name appears in the Location, the body or a log line.
async function noNameOut(res, ...names) {
  const location = res.headers.get("location") ?? "";
  const text = await res.clone().text();
  for (const n of [CANARY, ...names]) {
    assert.ok(!location.includes(n) && !location.includes(encodeURIComponent(n)), location);
    assert.ok(!text.includes(n), text);
    assert.ok(!logged.some((line) => line.includes(n)), logged.join("\n"));
  }
  assert.ok(!/[?&]display_name=/.test(location), location);
}

describe("POST /connections/mcp/rename", () => {
  it("renames through core and returns to the closed page status; the raw name never leaves the request", async () => {
    routes[RENAME] = () => Response.json({ outcome: "renamed" });
    const res = await renameRoute.POST(post("/connections/mcp/rename", enc({ csrf_token: CSRF, connection_id: ID, display_name: ` ${CANARY} ` })));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "/connections/mcp?status=renamed");
    assert.equal(res.headers.get("cache-control"), "no-store");
    const sent = core(RENAME);
    assert.equal(sent.length, 1);
    assert.equal(sent[0].headers.Authorization, `Session ${CRED}`);
    assert.deepEqual(sent[0].body, { origin: ORIGIN, connection_id: ID, csrf_token: CSRF, display_name: CANARY });
    await noNameOut(res);
    const acct = await renameRoute.POST(post("/connections/mcp/rename", enc({ csrf_token: CSRF, connection_id: ID, display_name: "x", return_to: "account" })));
    assert.equal(acct.headers.get("location"), "/account?mcp_status=renamed#mcp-connections");
  });
  it("an empty or whitespace-only name is a clear; the padded 80-non-BMP vector is sent as the 80 emoji", async () => {
    routes[RENAME] = () => Response.json({ outcome: "renamed" });
    for (const [value, want] of [["", ""], [cp(0x3000, 0x20, 0xa0), ""], [PADDED, EMOJI80]]) {
      calls = [];
      const res = await renameRoute.POST(post("/connections/mcp/rename", enc({ csrf_token: CSRF, connection_id: ID, display_name: value })));
      assert.equal(res.status, 303);
      assert.equal(core(RENAME)[0].body.display_name, want);
    }
  });
  it("malformed name bytes, control characters and 81 code points redirect to invalid_name with no core call", async () => {
    const base = `csrf_token=${CSRF}&connection_id=${ID}`;
    const bad = [
      `${base}&display_name=%FF%FE`,
      `${base}&display_name=%ED%A0%80`,
      `${base}&display_name=%C0%AF`,
      `${base}&display_name=Lap%0Atop`,
      `${base}&display_name=Laptop%0D`,
      `${base}&display_name=%C2%85Laptop`,
      `${base}&display_name=Laptop%E2%80%A8`,
      `${base}&${enc({ display_name: "a".repeat(81) })}`,
      `${base}&${enc({ display_name: EMOJI80 + cp(0x1f600) })}`,
    ];
    for (const body of bad) {
      const res = await renameRoute.POST(post("/connections/mcp/rename", body));
      assert.equal(res.status, 303, body);
      assert.equal(res.headers.get("location"), "/connections/mcp?status=invalid_name", body);
      await noNameOut(res, "Laptop");
    }
    const raw = Uint8Array.from([...new TextEncoder().encode(`${base}&return_to=account&display_name=`), 0xff, 0xfe]);
    const res = await renameRoute.POST(post("/connections/mcp/rename", raw));
    assert.equal(res.headers.get("location"), "/account?mcp_status=invalid_name#mcp-connections");
    assert.equal(calls.length, 0);
  });
  it("a repeated or absent name, a bad return_to and the legacy field rules are refused before core", async () => {
    const base = `csrf_token=${CSRF}&connection_id=${ID}`;
    for (const body of [
      `${base}&display_name=a&display_name=b`,
      `${base}&display_name=a&display%5Fname=a`,
      base,
      `${base}&display_name=a&return_to=evil`,
      `${base}&display_name=a&return_to=account&return_to=account`,
      `csrf_token=${CSRF}&csrf_token=${CSRF}&connection_id=${ID}&display_name=a`,
      `csrf_token=${CSRF}&display_name=a`,
      `csrf_token=${CSRF}&connection_id=${ID.toUpperCase()}&display_name=a`,
      `csrf_token=${CSRF}&connection_id=${ID}&connection_id=${ID}&display_name=a`,
    ]) {
      const res = await renameRoute.POST(post("/connections/mcp/rename", body));
      assert.equal(res.status, 400, body);
    }
    assert.equal(calls.length, 0);
  });
  it("the existing form layer: foreign Origin 403, wrong media type 415, no session 401, every other method 405", async () => {
    const body = enc({ csrf_token: CSRF, connection_id: ID, display_name: CANARY });
    assert.equal((await renameRoute.POST(post("/connections/mcp/rename", body, { origin: "https://evil.example" }))).status, 403);
    const text = new Request(`${ORIGIN}/connections/mcp/rename`, { method: "POST", headers: { host: HOST, origin: ORIGIN, "content-type": "text/plain", cookie: COOKIE }, body });
    assert.equal((await renameRoute.POST(text)).status, 415);
    const anon = await renameRoute.POST(post("/connections/mcp/rename", body, { cookie: null }));
    assert.equal(anon.status, 401);
    assert.match(await anon.text(), /Not signed in/);
    assert.equal(calls.length, 0);
    for (const m of ["GET", "PUT", "PATCH", "DELETE", "OPTIONS"]) {
      const res = renameRoute[m]();
      assert.equal(res.status, 405, m);
      assert.equal(res.headers.get("allow"), "POST");
    }
  });
  it("maps core's answers: not_found 404, unauthenticated 401, invalid_name 303, csrf_failed 403, not_available 403, unavailable 503", async () => {
    const body = enc({ csrf_token: CSRF, connection_id: ID, display_name: CANARY, return_to: "account" });
    const cases = [
      ["not_found", 404, 404, /not one of your active connections here/],
      ["unauthenticated", 401, 401, /Not signed in/],
      ["csrf_failed", 403, 403, /Form expired/],
      ["not_available", 403, 403, /not available for your account here/],
      ["mcp_unavailable", 503, 503, /could not be saved/],
    ];
    for (const [code, coreStatus, status, text] of cases) {
      routes[RENAME] = () => Response.json({ error: code }, { status: coreStatus });
      const res = await renameRoute.POST(post("/connections/mcp/rename", body));
      assert.equal(res.status, status, code);
      assert.match(await res.clone().text(), text);
      await noNameOut(res);
    }
    routes[RENAME] = () => Response.json({ error: "invalid_name" }, { status: 400 });
    const res = await renameRoute.POST(post("/connections/mcp/rename", body));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "/account?mcp_status=invalid_name#mcp-connections");
  });
});

describe("POST /connections/mcp/create", () => {
  it("without a name (absent or empty) the request to core is the unchanged unnamed create", async () => {
    routes[CREATE] = () => Response.json({ connection: { connection_id: ID, url: `${ORIGIN}/mcp/c/${ID}`, state: "pending" } }, { status: 201 });
    for (const body of [`csrf_token=${CSRF}`, `csrf_token=${CSRF}&display_name=`, `csrf_token=${CSRF}&display_name=+++`]) {
      calls = [];
      const res = await createRoute.POST(post("/connections/mcp/create", body));
      assert.equal(res.status, 303, body);
      assert.equal(res.headers.get("location"), "/connections/mcp?status=created");
      assert.deepEqual(core(CREATE)[0].body, { origin: ORIGIN, csrf_token: CSRF });
    }
  });
  it("a name is normalized and sent; the padded vector is sent as the 80 emoji; return_to=account returns to My Account", async () => {
    routes[CREATE] = () => Response.json({ connection: { connection_id: ID, url: `${ORIGIN}/mcp/c/${ID}`, state: "pending" } }, { status: 201 });
    let res = await createRoute.POST(post("/connections/mcp/create", enc({ csrf_token: CSRF, display_name: `${cp(0xa0)}${CANARY}`, return_to: "account" })));
    assert.equal(res.headers.get("location"), "/account?mcp_status=created#mcp-connections");
    assert.deepEqual(core(CREATE)[0].body, { origin: ORIGIN, csrf_token: CSRF, display_name: CANARY });
    await noNameOut(res);
    calls = [];
    res = await createRoute.POST(post("/connections/mcp/create", enc({ csrf_token: CSRF, display_name: PADDED })));
    assert.equal(core(CREATE)[0].body.display_name, EMOJI80);
  });
  it("malformed or invalid names redirect to invalid_name with no core call; a repeated name or bad return_to is 400", async () => {
    for (const body of [`csrf_token=${CSRF}&display_name=%FF%FE`, `csrf_token=${CSRF}&display_name=%ED%A0%80`, `csrf_token=${CSRF}&display_name=a%09b`,
      `csrf_token=${CSRF}&${enc({ display_name: "a".repeat(81) })}`]) {
      const res = await createRoute.POST(post("/connections/mcp/create", body));
      assert.equal(res.status, 303, body);
      assert.equal(res.headers.get("location"), "/connections/mcp?status=invalid_name");
    }
    for (const body of [`csrf_token=${CSRF}&display_name=a&display_name=a`, `csrf_token=${CSRF}&return_to=x`, `csrf_token=${CSRF}&csrf_token=${CSRF}`]) {
      assert.equal((await createRoute.POST(post("/connections/mcp/create", body))).status, 400, body);
    }
    assert.equal(calls.length, 0);
    routes[CREATE] = () => Response.json({ error: "invalid_name" }, { status: 400 });
    const res = await createRoute.POST(post("/connections/mcp/create", enc({ csrf_token: CSRF, display_name: "x" })));
    assert.equal(res.headers.get("location"), "/connections/mcp?status=invalid_name");
  });
});

describe("POST /connections/mcp/revoke", () => {
  it("keeps its contract and adds only the closed return_to", async () => {
    routes[REVOKE] = () => Response.json({ outcome: "revoked" });
    let res = await revokeRoute.POST(post("/connections/mcp/revoke", `csrf_token=${CSRF}&connection_id=${ID}`));
    assert.equal(res.headers.get("location"), "/connections/mcp?status=revoked");
    assert.deepEqual(core(REVOKE)[0].body, { origin: ORIGIN, connection_id: ID, csrf_token: CSRF });
    res = await revokeRoute.POST(post("/connections/mcp/revoke", `csrf_token=${CSRF}&connection_id=${ID}&return_to=account`));
    assert.equal(res.headers.get("location"), "/account?mcp_status=revoked#mcp-connections");
    calls = [];
    for (const body of [`csrf_token=${CSRF}&connection_id=${ID}&return_to=elsewhere`, `csrf_token=${CSRF}&connection_id=${ID}&return_to=account&return_to=connections`]) {
      assert.equal((await revokeRoute.POST(post("/connections/mcp/revoke", body))).status, 400, body);
    }
    assert.equal(calls.length, 0);
  });
});
