// Unit tests for lib/host.ts (G016): the gate's exempt paths, the closed reading of core's hosts/resolve answer, the
// relay's state and destination rules, the G062 public entry answer (signup_origin, no location) and gate path sets, and
// the exact outbound requests (no credential, no cache, no redirect), the G071 aliased platform root (root_alias on
// the platform answer only; no /signup path there) and the G071 root presentation (root_page on the public entry answer
// only; the root's 307 to /signup for GET/HEAD of exactly / in signup_redirect). fetch is
// replaced by a recorder, so nothing leaves the process. Run by `npm run test:unit` in the web check image with
// networking disabled. A resolve hook scoped to the module maps its extensionless sibling imports to the .ts sources.
import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import { afterEach, beforeEach, describe, it, mock } from "node:test";

registerHooks({
  resolve(specifier, context, nextResolve) {
    if ((specifier === "./config" || specifier === "./request-context") && context.parentURL?.endsWith("/lib/host.ts")) {
      return nextResolve(`${specifier}.ts`, context);
    }
    return nextResolve(specifier, context);
  },
});

const host = await import("../../lib/host.ts");
const { deriveContext } = await import("../../lib/request-context.ts");

const CORE = "http://core.internal:8180";
const ORG = "0b8f6c1e-2a4d-4c7b-9e1f-3a5b7c9d1e2f";
const CTX = deriveContext(new Headers({ "x-request-id": "req-host-1", traceparent: `00-${"a".repeat(32)}-${"b".repeat(16)}-01` }));
const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };
let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "unknown_host" }, { status: 404 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), init });
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

describe("gate exemptions", () => {
  it("exempts exactly sign-in, the MCP transport, well-known metadata and the probes", () => {
    for (const p of ["/auth/callback", "/auth/login", "/mcp/c/x", "/.well-known/oauth-protected-resource", "/api/health", "/api/ready"]) {
      assert.equal(host.isGateExempt(p), true, p);
    }
    for (const p of ["/", "/account", "/sign-in", "/connections/mcp", "/api/account", "/auth", "/mcp", "/api/healthz", "/authx/y"]) {
      assert.equal(host.isGateExempt(p), false, p);
    }
  });
});

describe("resolution answers", () => {
  it("reads every kind core answers", () => {
    assert.deepEqual(host.readResolution(200, { kind: "platform", organization_id: ORG }), { ok: true, value: { kind: "platform", organizationId: ORG, rootAlias: false } });
    assert.deepEqual(host.readResolution(200, { kind: "custom", organization_id: ORG, callback_origin: "https://auth.example.test:543" }), {
      ok: true,
      value: { kind: "custom", organizationId: ORG, callbackOrigin: "https://auth.example.test:543" },
    });
    assert.deepEqual(host.readResolution(200, { kind: "custom", organization_id: ORG }), { ok: true, value: { kind: "custom", organizationId: ORG, callbackOrigin: null } });
    assert.deepEqual(host.readResolution(200, { kind: "auth" }), { ok: true, value: { kind: "auth" } });
    assert.deepEqual(host.readResolution(200, { kind: "reserved" }), { ok: true, value: { kind: "reserved" } });
    // G062: a public entry names the exact platform root origin and no redirect location. G071: and core's root
    // presentation (root_page).
    assert.deepEqual(host.readResolution(200, { kind: "public_entry", signup_origin: "https://example.test:543", root_page: "signup_redirect" }), {
      ok: true,
      value: { kind: "public_entry", signupOrigin: "https://example.test:543", rootPage: "signup_redirect" },
    });
    assert.deepEqual(host.readResolution(200, { kind: "public_entry", signup_origin: "https://example.test", root_page: "landing" }), {
      ok: true,
      value: { kind: "public_entry", signupOrigin: "https://example.test", rootPage: "landing" },
    });
  });
  it("an unknown host is unknown_host; a store failure is unavailable, never a clean unknown host", () => {
    assert.deepEqual(host.readResolution(404, { error: "unknown_host" }), { ok: false, error: "unknown_host" });
    assert.deepEqual(host.readResolution(503, { error: "domains_unavailable" }), { ok: false, error: "unavailable" });
    assert.deepEqual(host.readResolution(503, undefined), { ok: false, error: "unavailable" });
  });
  it("anything outside the contract is malformed", () => {
    for (const [status, body] of [
      [200, { kind: "platform" }],
      [200, { kind: "platform", organization_id: "not-a-uuid" }],
      [200, { kind: "platform", organization_id: ORG, callback_origin: "https://a.example.test" }],
      [200, { kind: "custom", organization_id: ORG, callback_origin: "http://auth.example.test" }],
      [200, { kind: "custom", organization_id: ORG, callback_origin: "https://auth.example.test/path" }],
      [200, { kind: "auth", organization_id: ORG }],
      // G062: the retired location member, a missing, empty, non-https, path-bearing or non-origin signup_origin, and any
      // extra member are malformed (each beside a valid G071 root_page, so only the named defect is under test).
      [200, { kind: "public_entry" }],
      [200, { kind: "public_entry", root_page: "landing" }],
      [200, { kind: "public_entry", location: "https://www.example.test/", root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: "https://example.test", location: "https://www.example.test/", root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: "", root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: "javascript:alert(1)", root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: "http://example.test", root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: "https://example.test/", root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: "https://example.test/signup", root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: "https://user@example.test", root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: "https://Example.test", root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: 1, root_page: "landing" }],
      [200, { kind: "public_entry", signup_origin: "https://example.test", root_page: "landing", organization_id: ORG }],
      [200, { kind: "platform", organization_id: ORG, signup_origin: "https://example.test" }],
      [200, { kind: "reserved", location: "https://www.example.test/" }],
      [200, { kind: "tenant", organization_id: ORG }],
      [200, []],
      [404, { error: "unknown_host", extra: 1 }],
      [404, { error: "not_found" }],
      [400, { error: "ambiguous_parameter" }],
    ]) {
      assert.deepEqual(host.readResolution(status, body), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
  });
});

describe("G062 public entry and signup paths", () => {
  it("the root serves /, the shell's static files and the root signup surface; www serves / and the static files only", () => {
    for (const p of ["/", "/fonts/geist/geist-wght.ttf", "/blueprint/overview.svg"]) {
      assert.equal(host.publicEntryAllows(p, true), true, p);
      assert.equal(host.publicEntryAllows(p, false), true, p);
    }
    for (const p of ["/signup", "/signup/details", "/signup/start", "/signup/create", "/signup/abandon", "/signup/prove"]) {
      assert.equal(host.publicEntryAllows(p, true), true, p);
      assert.equal(host.publicEntryAllows(p, false), false, `www ${p}`);
    }
    // The tenant continuation hops, tenant pages, sign-in, the account and every other path are the fixed 404 on both.
    for (const p of ["/signup/continue", "/signup/complete", "/signup/", "/signup/x", "/signupx", "/sign-in", "/account", "/groups", "/join", "/api/account", "/fontsx/a", "//signup"]) {
      assert.equal(host.publicEntryAllows(p, true), false, p);
      assert.equal(host.publicEntryAllows(p, false), false, p);
    }
  });
  it("a tenant host serves only the two continuation hops under /signup, and only on a platform host", () => {
    for (const p of ["/signup", "/signup/", "/signup/details", "/signup/start", "/signup/create", "/signup/abandon", "/signup/prove", "/signup/x"]) {
      assert.equal(host.isSignupPath(p), true, p);
      assert.equal(host.tenantAllowsSignup(p, "platform"), false, p);
      assert.equal(host.tenantAllowsSignup(p, "custom"), false, p);
    }
    for (const p of ["/signup/continue", "/signup/complete"]) {
      assert.equal(host.tenantAllowsSignup(p, "platform"), true, p);
      assert.equal(host.tenantAllowsSignup(p, "custom"), false, `custom ${p}`);
    }
    for (const p of ["/", "/signupx", "/account", "/sign-in", "/join"]) assert.equal(host.isSignupPath(p), false, p);
  });
  it("isHttpsOrigin accepts only an exact lower-case https origin", () => {
    for (const o of ["https://example.test", "https://example.test:543", "https://org-a.example.test:543"]) assert.equal(host.isHttpsOrigin(o), true, o);
    for (const o of ["http://example.test", "https://example.test/", "https://Example.test", "https://user@example.test", "https://example.test?x", "example.test", "", 1, null]) {
      assert.equal(host.isHttpsOrigin(o), false, String(o));
    }
  });
});

// OS002: the default product mark and the favicon on the public entry and auth hosts. productAssetAllows is separate from
// publicEntryAllows (whose path set is unchanged) and passes exactly a GET or HEAD of the two exact paths.
describe("OS002 product mark and favicon paths", () => {
  const ASSETS = ["/brand/mark.png", "/icon.png"];
  it("passes GET and HEAD of exactly /brand/mark.png and /icon.png", () => {
    for (const p of ASSETS) {
      assert.equal(host.productAssetAllows("GET", p), true, `GET ${p}`);
      assert.equal(host.productAssetAllows("HEAD", p), true, `HEAD ${p}`);
    }
  });
  it("refuses every other method, including a lower-case get", () => {
    for (const p of ASSETS) {
      for (const m of ["POST", "PUT", "PATCH", "DELETE", "OPTIONS", "TRACE", "get", "head", "Get", ""]) {
        assert.equal(host.productAssetAllows(m, p), false, `${m} ${p}`);
      }
    }
  });
  it("refuses every other path and spelling (no prefix, case folding, trailing slash or normalization)", () => {
    const others = [
      "/brand/",
      "/brand",
      "/brand/mark.png/",
      "/brand/mark.PNG",
      "/Brand/mark.png",
      "/brand/other.png",
      "/brand/mark.png.bak",
      "/brand//mark.png",
      "/icon.png/x",
      "/icon.png/",
      "/icon.ico",
      "/ICON.png",
      "//icon.png",
      "/x/../icon.png",
      "/brand/../icon.png",
      "/icon.png?x",
      "icon.png",
      "/",
      "",
    ];
    for (const p of others) {
      assert.equal(host.productAssetAllows("GET", p), false, `GET ${p}`);
      assert.equal(host.productAssetAllows("HEAD", p), false, `HEAD ${p}`);
    }
  });
  it("leaves the public entry path set unchanged: the two files are not publicEntryAllows paths", () => {
    for (const p of ASSETS) {
      assert.equal(host.publicEntryAllows(p, true), false, p);
      assert.equal(host.publicEntryAllows(p, false), false, p);
    }
  });
});

// G071 (Plan r4 S5, S6, §4; §5.1 U5): core answers the exact platform root aliased to its configured default tenant as
// {kind: "platform", organization_id, root_alias: true}; the member is absent from every other answer.
describe("G071 aliased platform root", () => {
  it("reads root_alias: true on a platform answer only", () => {
    assert.deepEqual(host.readResolution(200, { kind: "platform", organization_id: ORG, root_alias: true }), {
      ok: true,
      value: { kind: "platform", organizationId: ORG, rootAlias: true },
    });
    // Without the member a platform host is not the aliased root (the inactive alias is byte-identical to today).
    assert.deepEqual(host.readResolution(200, { kind: "platform", organization_id: ORG }), {
      ok: true,
      value: { kind: "platform", organizationId: ORG, rootAlias: false },
    });
  });
  it("root_alias false, any other value, an extra member or the member on another kind is malformed", () => {
    for (const body of [
      { kind: "platform", organization_id: ORG, root_alias: false },
      { kind: "platform", organization_id: ORG, root_alias: 1 },
      { kind: "platform", organization_id: ORG, root_alias: "true" },
      { kind: "platform", organization_id: ORG, root_alias: null },
      { kind: "platform", organization_id: ORG, root_alias: {} },
      { kind: "platform", root_alias: true },
      { kind: "platform", organization_id: "not-a-uuid", root_alias: true },
      { kind: "platform", organization_id: ORG, root_alias: true, signup_origin: "https://example.test" },
      { kind: "platform", organization_id: ORG, root_alias: true, callback_origin: "https://auth.example.test" },
      { kind: "custom", organization_id: ORG, root_alias: true },
      { kind: "custom", organization_id: ORG, callback_origin: "https://auth.example.test:543", root_alias: true },
      { kind: "public_entry", signup_origin: "https://example.test", root_alias: true },
      { kind: "public_entry", signup_origin: "https://example.test", root_page: "landing", root_alias: true },
      { kind: "auth", root_alias: true },
      { kind: "reserved", root_alias: true },
    ]) {
      assert.deepEqual(host.readResolution(200, body), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
  });
  it("the aliased root serves no /signup path, the continuation hops included", () => {
    for (const p of [
      "/signup",
      "/signup/",
      "/signup/details",
      "/signup/start",
      "/signup/create",
      "/signup/abandon",
      "/signup/prove",
      "/signup/continue",
      "/signup/complete",
      "/signup/x",
    ]) {
      assert.equal(host.isSignupPath(p), true, p);
      assert.equal(host.tenantAllowsSignup(p, "platform", true), false, p);
    }
    // Every other platform host keeps the two continuation hops (the default is no alias).
    for (const p of ["/signup/continue", "/signup/complete"]) {
      assert.equal(host.tenantAllowsSignup(p, "platform"), true, p);
      assert.equal(host.tenantAllowsSignup(p, "platform", false), true, p);
    }
    // Tenant pages are not /signup paths, so the gate passes them on the aliased root as on any tenant host.
    for (const p of ["/", "/join", "/groups/new", "/sign-in", "/account", "/signupx"]) assert.equal(host.isSignupPath(p), false, p);
  });
  it("resolve reads the aliased root's answer through the same outbound request", async () => {
    reply = () => Response.json({ kind: "platform", organization_id: ORG, root_alias: true });
    assert.deepEqual(await host.resolveHost(CTX, "https://example.test:543"), { ok: true, value: { kind: "platform", organizationId: ORG, rootAlias: true } });
    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, `${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent("https://example.test:543")}`);
    reply = () => Response.json({ kind: "platform", organization_id: ORG, root_alias: false });
    assert.deepEqual(await host.resolveHost(CTX, "https://example.test:543"), { ok: false, error: "malformed" });
  });
});

// G071 (Plan r4 S15–S17, §4; §5.1 U5): every public entry answer carries core's effective root presentation, root_page
// "signup_redirect" or "landing"; the gate redirects only a GET or HEAD of exactly / on the exact root in
// signup_redirect, with the fixed same-origin Location <signup_origin>/signup, no-store, no body and no cookie.
describe("G071 root presentation", () => {
  const ROOT_ORIGIN = "https://example.test:543";
  it("reads root_page on a public entry answer, for the root and www alike", () => {
    for (const page of ["signup_redirect", "landing"]) {
      assert.deepEqual(host.readResolution(200, { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: page }), {
        ok: true,
        value: { kind: "public_entry", signupOrigin: ROOT_ORIGIN, rootPage: page },
      });
    }
  });
  it("a missing or malformed root_page, an extra member beside it, or the member on any other kind is malformed", () => {
    for (const body of [
      // Missing: the G062 two-member answer is no longer the contract.
      { kind: "public_entry", signup_origin: ROOT_ORIGIN },
      // Exact names only: no case folding, no other spelling, no whitespace, no other type.
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "Landing" },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "SIGNUP_REDIRECT" },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "signup-redirect" },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "redirect" },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: " landing" },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "landing " },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "" },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "/signup" },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: null },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: 1 },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: true },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: ["landing"] },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: {} },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: { mode: "landing" } },
      // An extra member beside a valid root_page.
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "landing", location: "/signup" },
      { kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "signup_redirect", redirect: "/signup" },
      // A valid root_page with a malformed signup_origin.
      { kind: "public_entry", signup_origin: "https://example.test/", root_page: "signup_redirect" },
      // The member on every other kind.
      { kind: "platform", organization_id: ORG, root_page: "landing" },
      { kind: "platform", organization_id: ORG, root_page: "signup_redirect" },
      { kind: "platform", organization_id: ORG, root_alias: true, root_page: "signup_redirect" },
      { kind: "custom", organization_id: ORG, root_page: "landing" },
      { kind: "custom", organization_id: ORG, callback_origin: "https://auth.example.test:543", root_page: "landing" },
      { kind: "auth", root_page: "landing" },
      { kind: "reserved", root_page: "signup_redirect" },
    ]) {
      assert.deepEqual(host.readResolution(200, body), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
  });
  it("resolve reads root_page through the same outbound request", async () => {
    reply = () => Response.json({ kind: "public_entry", signup_origin: ROOT_ORIGIN, root_page: "signup_redirect" });
    assert.deepEqual(await host.resolveHost(CTX, ROOT_ORIGIN), { ok: true, value: { kind: "public_entry", signupOrigin: ROOT_ORIGIN, rootPage: "signup_redirect" } });
    assert.equal(calls.length, 1);
    assert.equal(calls[0].url, `${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(ROOT_ORIGIN)}`);
    reply = () => Response.json({ kind: "public_entry", signup_origin: ROOT_ORIGIN });
    assert.deepEqual(await host.resolveHost(CTX, ROOT_ORIGIN), { ok: false, error: "malformed" });
  });
  it("redirects a GET or HEAD of exactly / on the root in signup_redirect, a request query included", () => {
    for (const method of ["GET", "HEAD"]) assert.equal(host.rootRedirects(method, "/", true, "signup_redirect"), true, method);
    // The gate decides on the pathname alone: a query leaves the pathname /, so the request redirects, and the Location
    // stays the fixed <signup_origin>/signup (nothing of the query is forwarded).
    for (const url of ["https://example.test:543/?x=1&state=y", "https://example.test:543/?utm=1&state=x", "https://example.test:543/?"]) {
      const pathname = new URL(url).pathname;
      assert.equal(pathname, "/", url);
      assert.equal(host.rootRedirects("GET", pathname, true, "signup_redirect"), true, url);
      assert.equal(host.rootSignupRedirect(ROOT_ORIGIN).headers.get("location"), `${ROOT_ORIGIN}/signup`, url);
    }
  });
  it("never redirects another method, another path, www or the landing presentation", () => {
    for (const method of ["POST", "PUT", "PATCH", "DELETE", "OPTIONS"]) {
      assert.equal(host.rootRedirects(method, "/", true, "signup_redirect"), false, method);
    }
    for (const p of ["/signup", "/signup/details", "/account", "/sign-in", "/join", "//", "/index", "/fonts/geist/geist-wght.ttf", "/api/health", ""]) {
      assert.equal(host.rootRedirects("GET", p, true, "signup_redirect"), false, p);
      assert.equal(host.rootRedirects("HEAD", p, true, "signup_redirect"), false, `HEAD ${p}`);
    }
    // www.<domain> (its own origin is not signup_origin) keeps its landing in either presentation.
    for (const page of ["signup_redirect", "landing"]) {
      for (const method of ["GET", "HEAD"]) assert.equal(host.rootRedirects(method, "/", false, page), false, `www ${method} ${page}`);
    }
    // The landing presentation keeps today's root page.
    for (const method of ["GET", "HEAD", "POST"]) assert.equal(host.rootRedirects(method, "/", true, "landing"), false, `landing ${method}`);
  });
  it("the redirect is a 307 with the fixed same-origin Location <signup_origin>/signup, no-store, no body and no cookie", async () => {
    const r = host.rootSignupRedirect(ROOT_ORIGIN);
    assert.equal(r.status, 307);
    assert.equal(r.headers.get("location"), "https://example.test:543/signup");
    // The Location follows the origin core named (a port, no path or query of its own), never anything else.
    assert.equal(host.rootSignupRedirect("https://example.test").headers.get("location"), "https://example.test/signup");
    assert.equal(r.headers.get("cache-control"), "no-store");
    assert.equal(r.headers.get("set-cookie"), null);
    assert.deepEqual([...r.headers.keys()].sort(), ["cache-control", "location"]);
    assert.equal(r.body, null);
    assert.equal(await r.text(), "");
    // Each answer is its own response (no shared, already-consumed object).
    assert.notEqual(host.rootSignupRedirect(ROOT_ORIGIN), r);
  });
});

describe("relay rules", () => {
  it("takes exactly one non-empty state of at most 512 characters", () => {
    assert.equal(host.relayState("code=c&state=abc.def"), "abc.def");
    assert.equal(host.relayState("error=access_denied&error_description=x&state=s1"), "s1");
    for (const q of ["", "code=c", "state=", "state=a&state=b", `state=${"s".repeat(513)}`, `state=s&pad=${"p".repeat(8 * 1024)}`]) {
      assert.equal(host.relayState(q), undefined, q.slice(0, 40));
    }
    assert.equal(host.relayState(`state=${"s".repeat(512)}`), "s".repeat(512));
  });
  it("appends the provider's raw query byte-identical to the custom host's callback", () => {
    const raw = "code=a%2Bb&state=x.y&error_description=A%20B&code=dup";
    assert.equal(host.relayLocation("https://portal.acme.example/auth/callback", raw), `https://portal.acme.example/auth/callback?${raw}`);
    assert.equal(host.relayLocation("https://portal.acme.example:543/auth/callback", "state=s"), "https://portal.acme.example:543/auth/callback?state=s");
  });
  it("refuses a destination outside https://<host>/auth/callback", () => {
    for (const cb of [
      "http://portal.acme.example/auth/callback",
      "https://portal.acme.example/auth/callback/x",
      "https://portal.acme.example/other",
      "https://user@portal.acme.example/auth/callback",
      "https://portal.acme.example/x/auth/callback",
      "//portal.acme.example/auth/callback",
      "/auth/callback",
    ]) {
      assert.equal(host.relayLocation(cb, "state=s"), undefined, cb);
    }
  });
});

describe("outbound requests", () => {
  it("resolve sends only the diagnostic headers, no credential, no cache and no redirect", async () => {
    reply = () => Response.json({ kind: "platform", organization_id: ORG });
    const result = await host.resolveHost(CTX, "https://org-a.example.test:543");
    assert.deepEqual(result, { ok: true, value: { kind: "platform", organizationId: ORG, rootAlias: false } });
    assert.equal(calls.length, 1);
    const { url, init } = calls[0];
    assert.equal(url, `${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent("https://org-a.example.test:543")}`);
    assert.equal(init.method, "GET");
    assert.equal(init.cache, "no-store");
    assert.equal(init.redirect, "error");
    assert.equal(init.body, undefined);
    assert.deepEqual(Object.keys(init.headers).sort(), ["X-Request-Id", "traceparent"].sort());
    assert.ok(init.signal instanceof AbortSignal);
  });
  it("no answer is unavailable", async () => {
    reply = () => {
      throw new TypeError("fetch failed");
    };
    assert.deepEqual(await host.resolveHost(CTX, "https://org-a.example.test:543"), { ok: false, error: "unavailable" });
    assert.deepEqual(await host.relayCallback(CTX, "s"), { ok: false, error: "unavailable" });
  });
  it("relay asks core with the state only and reads its closed answer", async () => {
    reply = () => Response.json({ callback: "https://portal.acme.example/auth/callback" });
    assert.deepEqual(await host.relayCallback(CTX, "abc.def+/="), { ok: true, callback: "https://portal.acme.example/auth/callback" });
    assert.equal(calls[0].url, `${CORE}/internal/v1/auth/relay?state=${encodeURIComponent("abc.def+/=")}`);
    assert.equal(calls[0].init.headers.Authorization, undefined);
    reply = () => Response.json({ error: "relay_refused" }, { status: 404 });
    assert.deepEqual(await host.relayCallback(CTX, "s"), { ok: false, error: "relay_refused" });
    reply = () => Response.json({ error: "identity_unavailable" }, { status: 503 });
    assert.deepEqual(await host.relayCallback(CTX, "s"), { ok: false, error: "unavailable" });
    reply = () => Response.json({ callback: 1 });
    assert.deepEqual(await host.relayCallback(CTX, "s"), { ok: false, error: "malformed" });
  });
  it("lookupContext keeps trace, span and flags and replaces the request id with a fresh generated one (L-2)", async () => {
    const lookup = host.lookupContext(CTX);
    assert.equal(CTX.requestId, "req-host-1");
    assert.equal(CTX.requestIdSource, "inbound");
    assert.match(lookup.requestId, /^[0-9a-f]{32}$/);
    assert.notEqual(lookup.requestId, CTX.requestId);
    assert.equal(lookup.requestIdSource, "generated");
    assert.deepEqual({ ...lookup, requestId: CTX.requestId, requestIdSource: CTX.requestIdSource }, CTX);
    assert.notEqual(host.lookupContext(CTX).requestId, lookup.requestId);
    reply = () => Response.json({ kind: "platform", organization_id: ORG });
    await host.resolveHost(lookup, "https://org-a.example.test:543");
    await host.resolveHost(CTX, "https://org-a.example.test:543");
    assert.equal(calls[0].init.headers["X-Request-Id"], lookup.requestId);
    assert.equal(calls[1].init.headers["X-Request-Id"], "req-host-1");
    assert.equal(calls[0].init.headers.traceparent, calls[1].init.headers.traceparent);
  });
});
