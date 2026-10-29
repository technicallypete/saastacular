// Unit tests for the G062 signup routes and the modified callback (Plan r5 §1.4, §4.4; SUP-14; Codex Round 14 option A):
// /signup/start (the signup-intent login on the root only), /signup/create (the suggestion, the create and the identical
// "Try again" from the host-only form restoration cookie: its flags and lifetime, tamper/stale refusal by core's intent
// contract, and no submitted value in any Location or log), /signup/abandon (POST revokes; GET only clears a dead
// credential's cookies), the three continuation hops (/signup/continue and /signup/complete on the tenant,
// /signup/prove on the root: core's URLs, the 120 s challenge cookie, recovery redirects to the configured root) and the
// callback (the root's signup credential, www 404, the tenant's tenant_created). fetch is a recorder; nothing leaves the
// process.
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

const { setSignupFormCookie } = await import("../../lib/session.ts");
const startRoute = await import("../../app/signup/start/route.ts");
const createRoute = await import("../../app/signup/create/route.ts");
const abandonRoute = await import("../../app/signup/abandon/route.ts");
const continueRoute = await import("../../app/signup/continue/route.ts");
const proveRoute = await import("../../app/signup/prove/route.ts");
const completeRoute = await import("../../app/signup/complete/route.ts");
const callbackRoute = await import("../../app/auth/callback/route.ts");

const CORE = "http://core.internal:8180";
const ROOT_HOST = "example.test:543";
const ROOT = `https://${ROOT_HOST}`;
const WWW_HOST = "www.example.test:543";
// G057 (Plan r5 §3.3): the new tenant's host is its generated platform label (adjective-noun-four-digit, or a 26-character label issued earlier); the slug "acme"
// is metadata and its host routes nowhere.
const TENANT_HOST = "q3vk7m2xa5rt6pwz4bnhyc2dfe.example.test:543";
const TENANT = `https://${TENANT_HOST}`;
const SLUG_HOST = "acme.example.test:543";
const CRED = "S".repeat(43);
const CSRF = "C".repeat(43);
const BINDING = "B".repeat(43);
const REF = "R".repeat(43);
const NONCE = "N".repeat(43);
const PROOF = "P".repeat(43);
const CHALLENGE = "K".repeat(43);
const ORG = "0b8f6c1e-2a4d-4c7b-9e1f-3a5b7c9d1e2f";
const NAME = "Acme Canary Works";
const SLUG = "acme";
const ENV = {
  PRODUCT_NAME: "Test Product",
  PLATFORM_BASE_DOMAIN: "example.test",
  CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test",
  CORE_INTERNAL_URL: CORE,
  IDENTITY_COOKIE_PROFILE: "https",
  APP_ENV: "test",
};
const SIGNUP = `__Host-signup=${CRED}`;
const RECORD = setSignupFormCookie({ name: NAME, slug: SLUG }, "https").split(";")[0];
const CONTINUE = `${TENANT}/signup/continue?ref=${REF}`;

let saved;
let calls;
let routes;
let logged;

// Core's answers by path; hosts/resolve answers by the asked origin (the root and www are the public entry of ROOT, the
// tenant a platform host).
function resolveAnswer(url) {
  const origin = new URL(url).searchParams.get("origin");
  if (origin === ROOT || origin === `https://${WWW_HOST}`) return Response.json({ kind: "public_entry", signup_origin: ROOT, root_page: "signup_redirect" });
  if (origin === TENANT) return Response.json({ kind: "platform", organization_id: ORG });
  return Response.json({ error: "unknown_host" }, { status: 404 });
}

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
    if (u.startsWith(`${CORE}/internal/v1/hosts/resolve?`)) return resolveAnswer(u);
    const path = new URL(u).pathname;
    const answer = routes[path];
    return answer === undefined ? Response.json({ error: "signup_unavailable" }, { status: 503 }) : answer();
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
function post(path, body, cookie, host = ROOT_HOST) {
  return new Request(`https://${host}${path}`, {
    method: "POST",
    headers: { host, origin: `https://${host}`, "content-type": "application/x-www-form-urlencoded", ...(cookie === undefined ? {} : { cookie }) },
    body,
  });
}
function get(path, cookie, host = ROOT_HOST) {
  return new Request(`https://${host}${path}`, { headers: { host, ...(cookie === undefined ? {} : { cookie }) } });
}
const cookies = (res) => res.headers.getSetCookie();
const named = (res, name) => cookies(res).filter((c) => c.startsWith(`${name}=`));
const core = (path) => calls.filter((c) => c.url === `${CORE}${path}`);
// Nothing submitted (the name or the slug as a value) appears in a Location header or a log line.
function noSubmittedValueOut(res) {
  const location = res.headers.get("location") ?? "";
  assert.ok(!location.includes(encodeURIComponent(NAME)) && !location.includes("Canary"), location);
  assert.ok(!/[?&](name|slug)=/.test(location), location);
  assert.ok(!logged.some((line) => line.includes("Canary") || line.includes(CRED)), logged.join("\n"));
}
const created = (outcome = "created") => () => Response.json({ outcome, organization: { id: ORG, slug: SLUG }, continue: CONTINUE }, { status: 201 });

describe("POST /signup/start", () => {
  it("starts the signup-intent login at the root only; login binding set, old signup cookies cleared", async () => {
    routes["/internal/v1/auth/login"] = () => Response.json({ authorization_url: "https://auth.example.test/authorize?s=1", binding: { value: BINDING, max_age: 600 } });
    const res = await startRoute.POST(post("/signup/start", "", `${SIGNUP}; ${RECORD}`));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "https://auth.example.test/authorize?s=1");
    assert.equal(res.headers.get("cache-control"), "no-store");
    assert.equal(res.headers.get("referrer-policy"), "no-referrer");
    const login = core("/internal/v1/auth/login");
    assert.equal(login.length, 1);
    assert.deepEqual(login[0].body, { origin: ROOT, intent: "signup" });
    assert.equal(login[0].headers.Authorization, undefined);
    assert.ok(named(res, "__Host-login")[0].startsWith(`__Host-login=${BINDING}; Path=/; Max-Age=600; HttpOnly; SameSite=Lax; Secure`));
    assert.equal(named(res, "__Host-signup")[0], "__Host-signup=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure");
    assert.equal(named(res, "__Host-signup-form")[0], "__Host-signup-form=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure");
  });
  it("www, a tenant host and a foreign Origin never reach core's login; closed refusals", async () => {
    for (const host of [WWW_HOST, TENANT_HOST]) {
      const res = await startRoute.POST(post("/signup/start", "", undefined, host));
      assert.equal(res.status, 404, host);
    }
    const foreign = new Request(`${ROOT}/signup/start`, { method: "POST", headers: { host: ROOT_HOST, origin: TENANT, "content-type": "application/x-www-form-urlencoded" } });
    assert.equal((await startRoute.POST(foreign)).status, 403);
    assert.equal(core("/internal/v1/auth/login").length, 0);
    routes["/internal/v1/auth/login"] = () => Response.json({ error: "rate_limited" }, { status: 429 });
    assert.equal((await startRoute.POST(post("/signup/start", ""))).headers.get("location"), "/signup?status=rate_limited");
    routes["/internal/v1/auth/login"] = () => Response.json({ error: "identity_unavailable" }, { status: 503 });
    assert.equal((await startRoute.POST(post("/signup/start", ""))).headers.get("location"), "/signup?status=unavailable");
    assert.equal(startRoute.GET().status, 405);
  });
});

describe("POST /signup/create", () => {
  it("name only: core's suggestion; the record holds the name and the suggested address; nothing in the Location", async () => {
    routes["/internal/v1/signup/slug"] = () => Response.json({ suggested: SLUG, available: true });
    const res = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, name: NAME }), SIGNUP));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), "/signup/details");
    const slug = core("/internal/v1/signup/slug")[0];
    assert.equal(slug.headers.Authorization, `Signup ${CRED}`);
    assert.deepEqual(slug.body, { origin: ROOT, csrf_token: CSRF, name: NAME });
    assert.deepEqual(named(res, "__Host-signup-form"), [setSignupFormCookie({ name: NAME, slug: SLUG }, "https")]);
    noSubmittedValueOut(res);
  });
  it("create: 201 keeps the signup cookie, records exactly the submitted values (flags, ≤ 30 min) and 303s into core's continuation", async () => {
    routes["/internal/v1/signup/organizations"] = created();
    const res = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, name: NAME, slug: SLUG }), SIGNUP));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), CONTINUE);
    assert.deepEqual(core("/internal/v1/signup/organizations")[0].body, { origin: ROOT, csrf_token: CSRF, name: NAME, slug: SLUG });
    assert.deepEqual(named(res, "__Host-signup"), []);
    const [record] = named(res, "__Host-signup-form");
    assert.match(record, /^__Host-signup-form=v1\.[A-Za-z0-9_-]+\.acme; Path=\/; Max-Age=1800; HttpOnly; SameSite=Lax; Secure$/);
    assert.doesNotMatch(record, /domain|expires/i);
    assert.ok(!record.includes(CRED) && !record.includes(CSRF) && !record.includes(REF));
    noSubmittedValueOut(res);
  });
  it("continue_failed recovery: Try again repeats the identical recorded payload and continues (already_created); the record is not rewritten", async () => {
    routes["/internal/v1/signup/organizations"] = created("already_created");
    const res = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, retry: "1" }), `${SIGNUP}; ${RECORD}`));
    assert.equal(res.headers.get("location"), CONTINUE);
    assert.deepEqual(core("/internal/v1/signup/organizations")[0].body, { origin: ROOT, csrf_token: CSRF, name: NAME, slug: SLUG });
    assert.deepEqual(cookies(res), []);
    noSubmittedValueOut(res);
  });
  it("a tampered or changed record is refused by core's intent contract and nothing is rewritten; a stale or missing record re-mints nothing", async () => {
    routes["/internal/v1/signup/organizations"] = () => Response.json({ error: "conflict", field: "intent_used" }, { status: 409 });
    const tampered = setSignupFormCookie({ name: "Another Canary", slug: SLUG }, "https").split(";")[0];
    const res = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, retry: "1" }), `${SIGNUP}; ${tampered}`));
    assert.equal(res.headers.get("location"), "/signup/details?status=intent_used");
    assert.deepEqual(cookies(res), []);
    // A changed submission after provisioning keeps the existing record too.
    const changed = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, name: "Changed Canary", slug: SLUG }), `${SIGNUP}; ${RECORD}`));
    assert.equal(changed.headers.get("location"), "/signup/details?status=intent_used");
    assert.deepEqual(named(changed, "__Host-signup-form"), []);
    const signupCalls = () => calls.filter((c) => c.url.includes("/internal/v1/signup/")).length;
    const before = signupCalls();
    for (const cookie of [SIGNUP, `${SIGNUP}; __Host-signup-form=v1.QWNtZR.acme`, `${SIGNUP}; ${RECORD}; ${RECORD}`, `${SIGNUP}; __Host-signup-form=v1.QWNtZQ.`]) {
      const r = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, retry: "1" }), cookie));
      assert.equal(r.headers.get("location"), "/signup/details?status=retry_unavailable", cookie);
    }
    assert.equal(signupCalls(), before, "a stale or missing record never reaches core");
    noSubmittedValueOut(res);
  });
  it("field refusals and an unknown outcome restore the form from the record; slug_taken keeps the submitted values", async () => {
    for (const [answer, location] of [
      [() => Response.json({ error: "conflict", field: "slug_taken" }, { status: 409 }), "/signup/details?status=slug_taken"],
      [() => Response.json({ error: "invalid_request", field: "slug" }, { status: 400 }), "/signup/details?status=invalid_slug"],
      [() => Response.json({ error: "invalid_request", field: "name" }, { status: 400 }), "/signup/details?status=invalid_name"],
      [() => Response.json({ error: "rate_limited" }, { status: 429 }), "/signup/details?status=rate_limited"],
      [() => Response.json({ error: "signup_unavailable" }, { status: 503 }), "/signup/details?status=unknown"],
      [() => new Response("not json", { status: 502 }), "/signup/details?status=unknown"],
    ]) {
      routes["/internal/v1/signup/organizations"] = answer;
      const res = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, name: NAME, slug: SLUG }), SIGNUP));
      assert.equal(res.headers.get("location"), location);
      assert.deepEqual(named(res, "__Host-signup-form"), [setSignupFormCookie({ name: NAME, slug: SLUG }, "https")], location);
      noSubmittedValueOut(res);
    }
  });
  it("a dead or missing credential clears both signup cookies; a stale form, malformed fields and other hosts never reach core", async () => {
    routes["/internal/v1/signup/organizations"] = () => Response.json({ error: "unauthenticated" }, { status: 401 });
    const dead = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, name: NAME, slug: SLUG }), `${SIGNUP}; ${RECORD}`));
    assert.equal(dead.headers.get("location"), "/signup?status=expired");
    assert.deepEqual(cookies(dead).map((c) => c.split(";")[0]).sort(), ["__Host-signup-form=", "__Host-signup="]);
    routes["/internal/v1/signup/organizations"] = () => Response.json({ error: "csrf_failed" }, { status: 403 });
    const stale = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, name: NAME, slug: SLUG }), SIGNUP));
    assert.equal(stale.status, 403);
    assert.deepEqual(cookies(stale), []);
    const before = core("/internal/v1/signup/organizations").length;
    const none = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, name: NAME, slug: SLUG })));
    assert.equal(none.headers.get("location"), "/signup?status=expired");
    for (const fields of [
      { name: NAME, slug: SLUG },
      { csrf_token: "", name: NAME, slug: SLUG },
      { csrf_token: CSRF },
      { csrf_token: CSRF, retry: "1", name: NAME },
      { csrf_token: CSRF, retry: "yes" },
    ]) {
      assert.equal((await createRoute.POST(post("/signup/create", form(fields), SIGNUP))).status, 400, JSON.stringify(fields));
    }
    assert.equal((await createRoute.POST(post("/signup/create", form([["csrf_token", CSRF], ["name", NAME], ["name", NAME]]), SIGNUP))).status, 400);
    for (const host of [WWW_HOST, TENANT_HOST]) {
      assert.equal((await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, name: NAME, slug: SLUG }), SIGNUP, host))).status, 404, host);
    }
    assert.equal(core("/internal/v1/signup/organizations").length, before);
    assert.equal(createRoute.GET().status, 405);
  });
});

// G062 bounded security correction to Plan r5 §4.4 (Codex Round 15 decision 2): a GET of /signup/details never re-mints.
// The page is a Server Component (TSX, not loadable by node --test), so its source is checked for the absence of every
// mutating core call, and the recovery decision it renders is lib/signup-client.ts recoveryView (signup-client.test.mjs).
describe("GET /signup/details never mutates (bounded security correction)", () => {
  const PAGE = new URL("../../app/signup/details/page.tsx", import.meta.url);
  it("the details page imports and calls no create, re-mint, slug, abandon or continuation function; it redirects only to /signup and /signup/abandon", async () => {
    const { readFile } = await import("node:fs/promises");
    const source = await readFile(PAGE, "utf8");
    for (const mutating of ["createOrganization", "suggestSlug", "abandonSignup", "challenge(", "prove(", "complete(", "fetch("]) {
      assert.ok(!source.includes(mutating), `details page must not use ${mutating}`);
    }
    const imported = /import \{([^}]*)\} from "@\/lib\/signup-client";/.exec(source);
    assert.notEqual(imported, null);
    assert.deepEqual(imported[1].split(",").map((s) => s.trim()).filter(Boolean).sort(), ["DETAILS_NOTICES", "fieldErrorOf", "noticeOf", "recoveryView", "signupStatus"]);
    assert.deepEqual([...source.matchAll(/redirect\(([^)]*)\)/g)].map((m) => m[1]).sort(), ['"/signup"', '"/signup/abandon"']);
  });
  it("a cross-site-style GET of the create route is refused with no core call; Try again is the only re-mint and is a POST", async () => {
    const before = calls.length;
    assert.equal(createRoute.GET().status, 405);
    assert.equal(calls.length, before);
    // The user-triggered Try again: the identical recorded payload, the original credential and the CSRF token.
    routes["/internal/v1/signup/organizations"] = created("already_created");
    const res = await createRoute.POST(post("/signup/create", form({ csrf_token: CSRF, retry: "1" }), `${SIGNUP}; ${RECORD}`));
    assert.equal(res.headers.get("location"), CONTINUE);
    const [call] = core("/internal/v1/signup/organizations");
    assert.equal(call.headers.Authorization, `Signup ${CRED}`);
    assert.deepEqual(call.body, { origin: ROOT, csrf_token: CSRF, name: NAME, slug: SLUG });
    // A foreign Origin (a cross-site form POST) never reaches core.
    const foreign = new Request(`${ROOT}/signup/create`, {
      method: "POST",
      headers: { host: ROOT_HOST, origin: "https://attacker.example.net", "content-type": "application/x-www-form-urlencoded", cookie: `${SIGNUP}; ${RECORD}` },
      body: form({ csrf_token: CSRF, retry: "1" }),
    });
    assert.equal((await createRoute.POST(foreign)).status, 403);
    assert.equal(core("/internal/v1/signup/organizations").length, 1);
    noSubmittedValueOut(res);
  });
});

// G057 (Plan r5 §3.3, §5.2 T7): the slug is the organization's short name, metadata only; the web address is a label core
// generates at creation. The form and page are TSX (not loadable by node --test), so their source carries the copy check.
describe("the details copy (G057): the slug is a short name, never the web address", () => {
  it("the field is the short name with the generated-address sentence; no address is previewed from the slug; the notices say short name", async () => {
    const { readFile } = await import("node:fs/promises");
    const formSource = await readFile(new URL("../../components/signup/details-form.tsx", import.meta.url), "utf8");
    assert.ok(formSource.includes('const SLUG_LABEL = "Short name";'), "the slug field is labelled Short name");
    assert.ok(
      formSource.includes("Your organization's web address is generated when it is created; this short name is not its web address."),
      "the hint says the web address is generated and the short name is not it",
    );
    assert.ok(!/rootHost|<address>|will be at/.test(formSource), "no web address is built or previewed from the slug");
    const pageSource = await readFile(new URL("../../app/signup/details/page.tsx", import.meta.url), "utf8");
    assert.ok(pageSource.includes('title = "Choose its short name";'));
    assert.ok(!/rootHost|choose its address|The address is suggested/.test(pageSource), "the details page names no slug-built address");
    // Owner-approved copy (G057 Plan r5 §3.3): step one keeps "Next, you choose its short name."; step two is revised.
    assert.ok(pageSource.includes("Next, you choose its short name."), "step one names the short name");
    assert.ok(
      pageSource.includes("Step 2 of 2. We'll suggest a unique short name based on your organization name. You can change it."),
      "step two is the owner-approved helper text",
    );
    assert.ok(!pageSource.includes("The short name is suggested from the name"), "the superseded step-two sentence is gone");
    const { DETAILS_NOTICES } = await import("../../lib/signup-client.ts");
    assert.equal(DETAILS_NOTICES.slug_taken, "That short name is already taken. Choose another one.");
    assert.ok(DETAILS_NOTICES.invalid_slug.endsWith(" Some short names are reserved."));
  });
});

describe("/signup/abandon", () => {
  it("POST revokes with the CSRF token and clears both cookies; a dead credential is cleared too", async () => {
    routes["/internal/v1/signup/abandon"] = () => Response.json({});
    const res = await abandonRoute.POST(post("/signup/abandon", form({ csrf_token: CSRF }), `${SIGNUP}; ${RECORD}`));
    assert.equal(res.headers.get("location"), "/signup?status=abandoned");
    assert.deepEqual(core("/internal/v1/signup/abandon")[0].body, { origin: ROOT, csrf_token: CSRF });
    assert.deepEqual(cookies(res).map((c) => c.split(";")[0]).sort(), ["__Host-signup-form=", "__Host-signup="]);
    routes["/internal/v1/signup/abandon"] = () => Response.json({ error: "unauthenticated" }, { status: 401 });
    const dead = await abandonRoute.POST(post("/signup/abandon", form({ csrf_token: CSRF }), SIGNUP));
    assert.equal(dead.headers.get("location"), "/signup?status=expired");
    assert.equal(cookies(dead).length, 2);
    routes["/internal/v1/signup/abandon"] = () => Response.json({ error: "csrf_failed" }, { status: 403 });
    assert.equal((await abandonRoute.POST(post("/signup/abandon", form({ csrf_token: CSRF }), SIGNUP))).status, 403);
    assert.equal((await abandonRoute.POST(post("/signup/abandon", form({}), SIGNUP))).status, 400);
  });
  it("GET clears the cookies only when core reports the credential dead; a live credential is untouched", async () => {
    routes["/internal/v1/signup/status"] = () => Response.json({ error: "unauthenticated" }, { status: 401 });
    const dead = await abandonRoute.GET(get("/signup/abandon", `${SIGNUP}; ${RECORD}`));
    assert.equal(dead.headers.get("location"), "/signup?status=expired");
    assert.equal(cookies(dead).length, 2);
    routes["/internal/v1/signup/status"] = () => Response.json({ status: "pending", csrf_token: CSRF });
    const live = await abandonRoute.GET(get("/signup/abandon", `${SIGNUP}; ${RECORD}`));
    assert.equal(live.headers.get("location"), "/signup/details");
    assert.deepEqual(cookies(live), []);
    assert.equal(core("/internal/v1/signup/abandon").length, 0, "a GET never revokes");
    const none = await abandonRoute.GET(get("/signup/abandon", RECORD));
    assert.equal(none.headers.get("location"), "/signup");
    assert.equal((await abandonRoute.GET(get("/signup/abandon?x=1", SIGNUP))).status, 404);
    assert.equal((await abandonRoute.GET(get("/signup/abandon", SIGNUP, WWW_HOST))).status, 404);
  });
});

describe("continuation hops", () => {
  it("continue (tenant): core's challenge with the binding and no credential; the 120 s challenge cookie and core's prove URL", async () => {
    routes["/internal/v1/signup/challenge"] = () => Response.json({ challenge: { value: CHALLENGE, max_age: 120 }, prove: `${ROOT}/signup/prove?n=${NONCE}` });
    const res = await continueRoute.GET(get(`/signup/continue?ref=${REF}`, `__Host-login=${BINDING}; ${SIGNUP}`, TENANT_HOST));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), `${ROOT}/signup/prove?n=${NONCE}`);
    assert.equal(res.headers.get("cache-control"), "no-store");
    assert.equal(res.headers.get("referrer-policy"), "no-referrer");
    assert.deepEqual(cookies(res), [`__Host-sct=${CHALLENGE}; Path=/; Max-Age=120; HttpOnly; SameSite=Lax; Secure`]);
    assert.ok(calls.some((c) => c.url === `${CORE}/internal/v1/hosts/resolve?origin=${encodeURIComponent(ROOT)}`), "the root is configuration confirmed by core");
    const ch = core("/internal/v1/signup/challenge")[0];
    assert.deepEqual(ch.body, { origin: TENANT, ref: REF });
    assert.equal(ch.headers.Authorization, undefined);
    assert.equal(ch.headers["X-Login-Binding"], BINDING);
  });
  it("continue: a refusal or a malformed ref goes to the root's recovery page with nothing set; never on the root", async () => {
    routes["/internal/v1/signup/challenge"] = () => Response.json({ error: "continuation_invalid" }, { status: 404 });
    const res = await continueRoute.GET(get(`/signup/continue?ref=${REF}`, undefined, TENANT_HOST));
    assert.equal(res.headers.get("location"), `${ROOT}/signup/details?status=continue_failed`);
    assert.deepEqual(cookies(res), []);
    const before = core("/internal/v1/signup/challenge").length;
    for (const q of ["", "?ref=short", `?ref=${REF}&ref=${REF}`, `?ref=${REF}&origin=https://evil.test`, `?n=${REF}`]) {
      const r = await continueRoute.GET(get(`/signup/continue${q}`, undefined, TENANT_HOST));
      assert.equal(r.headers.get("location"), `${ROOT}/signup/details?status=continue_failed`, q);
    }
    assert.equal(core("/internal/v1/signup/challenge").length, before);
    assert.equal((await continueRoute.GET(get(`/signup/continue?ref=${REF}`, undefined, ROOT_HOST))).status, 404);
    // G057: the slug's host is not a generated label host; the hop is the fixed 404 there with no core call at all.
    const quiet = calls.length;
    assert.equal((await continueRoute.GET(get(`/signup/continue?ref=${REF}`, `__Host-login=${BINDING}`, SLUG_HOST))).status, 404);
    assert.equal(calls.length, quiet, "no lookup or challenge for the slug host");
    assert.equal(continueRoute.POST().status, 405);
  });
  it("prove (root): the original signup credential spends the nonce; core's complete URL", async () => {
    routes["/internal/v1/signup/prove"] = () => Response.json({ complete: `${TENANT}/signup/complete?p=${PROOF}` });
    const res = await proveRoute.GET(get(`/signup/prove?n=${NONCE}`, SIGNUP));
    assert.equal(res.headers.get("location"), `${TENANT}/signup/complete?p=${PROOF}`);
    assert.deepEqual(cookies(res), []);
    const p = core("/internal/v1/signup/prove")[0];
    assert.deepEqual(p.body, { origin: ROOT, nonce: NONCE });
    assert.equal(p.headers.Authorization, `Signup ${CRED}`);
  });
  it("prove: no credential goes to /signup, a mismatch to the recovery page, a dead credential clears both cookies; never on a tenant", async () => {
    const before = core("/internal/v1/signup/prove").length;
    const none = await proveRoute.GET(get(`/signup/prove?n=${NONCE}`));
    assert.equal(none.headers.get("location"), "/signup?status=expired");
    assert.equal(core("/internal/v1/signup/prove").length, before);
    routes["/internal/v1/signup/prove"] = () => Response.json({ error: "signup_context_mismatch" }, { status: 403 });
    assert.equal((await proveRoute.GET(get(`/signup/prove?n=${NONCE}`, SIGNUP))).headers.get("location"), "/signup/details?status=continue_failed");
    routes["/internal/v1/signup/prove"] = () => Response.json({ error: "unauthenticated" }, { status: 401 });
    const dead = await proveRoute.GET(get(`/signup/prove?n=${NONCE}`, `${SIGNUP}; ${RECORD}`));
    assert.equal(dead.headers.get("location"), "/signup?status=expired");
    assert.equal(cookies(dead).length, 2);
    assert.equal((await proveRoute.GET(get(`/signup/prove?n=${NONCE}`, SIGNUP, TENANT_HOST))).status, 404);
    assert.equal((await proveRoute.GET(get(`/signup/prove?n=${NONCE}`, SIGNUP, WWW_HOST))).status, 404);
  });
  it("complete (tenant): the challenge cookie and P; challenge cleared, fresh binding set, 303 to the provider", async () => {
    routes["/internal/v1/signup/complete"] = () => Response.json({ authorization_url: "https://auth.example.test/authorize?p=1", binding: { value: BINDING, max_age: 600 } });
    const res = await completeRoute.GET(get(`/signup/complete?p=${PROOF}`, `__Host-sct=${CHALLENGE}`, TENANT_HOST));
    assert.equal(res.headers.get("location"), "https://auth.example.test/authorize?p=1");
    assert.deepEqual(cookies(res), [
      "__Host-sct=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure",
      `__Host-login=${BINDING}; Path=/; Max-Age=600; HttpOnly; SameSite=Lax; Secure`,
    ]);
    const c = core("/internal/v1/signup/complete")[0];
    assert.deepEqual(c.body, { origin: TENANT, challenge: CHALLENGE, proof: PROOF });
    assert.equal(c.headers.Authorization, undefined);
  });
  it("complete: a missing challenge cookie or a refusal clears it and goes to the root's recovery page", async () => {
    routes["/internal/v1/signup/complete"] = () => Response.json({ error: "continuation_invalid" }, { status: 404 });
    for (const cookie of [undefined, `__Host-sct=${CHALLENGE}`, `__Host-sct=${CHALLENGE}; __Host-sct=${CHALLENGE}`]) {
      const res = await completeRoute.GET(get(`/signup/complete?p=${PROOF}`, cookie, TENANT_HOST));
      assert.equal(res.headers.get("location"), `${ROOT}/signup/details?status=continue_failed`, String(cookie));
      assert.deepEqual(cookies(res), ["__Host-sct=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax; Secure"]);
    }
    assert.equal(core("/internal/v1/signup/complete").length, 1);
    assert.equal((await completeRoute.GET(get(`/signup/complete?p=${PROOF}`, `__Host-sct=${CHALLENGE}`, ROOT_HOST))).status, 404);
    const quiet = calls.length;
    assert.equal((await completeRoute.GET(get(`/signup/complete?p=${PROOF}`, `__Host-sct=${CHALLENGE}`, SLUG_HOST))).status, 404);
    assert.equal(calls.length, quiet, "no lookup or completion for the slug host");
  });
});

describe("GET /auth/callback (G062 answers)", () => {
  it("root: the 202 signup credential becomes the 1800 s host-only signup cookie and the browser goes to /signup/details", async () => {
    routes["/internal/v1/auth/callback"] = () => Response.json({ signup: { credential: { value: CRED, max_age: 1800 }, next: "/signup/details" } }, { status: 202 });
    const res = await callbackRoute.GET(get("/auth/callback?code=c&state=s", `__Host-login=${BINDING}`));
    assert.equal(res.headers.get("location"), "/signup/details");
    assert.deepEqual(cookies(res), [`__Host-signup=${CRED}; Path=/; Max-Age=1800; HttpOnly; SameSite=Lax; Secure`]);
    const cb = core("/internal/v1/auth/callback")[0];
    assert.deepEqual(cb.body, { origin: ROOT, raw_query: "code=c&state=s" });
    assert.equal(cb.headers["X-Login-Binding"], BINDING);
    assert.equal(cb.headers.Authorization, undefined);
  });
  it("root: refusals go to /signup with a closed status; a session answer sets nothing; www is 404 without a callback", async () => {
    for (const [answer, location] of [
      [() => Response.json({ error: "login_canceled" }, { status: 400 }), "/signup?status=canceled"],
      [() => Response.json({ error: "transaction_expired" }, { status: 400 }), "/signup?status=expired"],
      [() => Response.json({ error: "rate_limited" }, { status: 429 }), "/signup?status=rate_limited"],
      [() => Response.json({ error: "code_rejected" }, { status: 400 }), "/signup?status=failed"],
      [() => Response.json({ outcome: "signed_in", return_to: "/", session: { value: CRED, max_age: 600 } }), "/signup?status=failed"],
    ]) {
      routes["/internal/v1/auth/callback"] = answer;
      const res = await callbackRoute.GET(get("/auth/callback?code=c&state=s"));
      assert.equal(res.headers.get("location"), location);
      assert.deepEqual(cookies(res), []);
    }
    const before = core("/internal/v1/auth/callback").length;
    const www = await callbackRoute.GET(get("/auth/callback?code=c&state=s", undefined, WWW_HOST));
    assert.equal(www.status, 404);
    assert.equal(core("/internal/v1/auth/callback").length, before);
  });
  it("tenant: the pinned callback's tenant_created signs in and lands on /?status=tenant_created; account substitution fails closed", async () => {
    routes["/internal/v1/auth/callback"] = () => Response.json({ outcome: "tenant_created", return_to: "/", session: { value: CRED, max_age: 28800 } });
    const res = await callbackRoute.GET(get("/auth/callback?code=c&state=s", `__Host-login=${BINDING}`, TENANT_HOST));
    assert.equal(res.headers.get("location"), "/?status=tenant_created");
    assert.deepEqual(cookies(res), [`__Host-sid=${CRED}; Path=/; Max-Age=28800; HttpOnly; SameSite=Lax; Secure`]);
    routes["/internal/v1/auth/callback"] = () => Response.json({ error: "signup_user_mismatch" }, { status: 403 });
    const mismatch = await callbackRoute.GET(get("/auth/callback?code=c&state=s", `__Host-login=${BINDING}`, TENANT_HOST));
    assert.equal(mismatch.headers.get("location"), "/sign-in?status=failed");
    assert.deepEqual(cookies(mismatch), []);
    routes["/internal/v1/auth/callback"] = () => Response.json({ signup: { credential: { value: CRED, max_age: 1800 }, next: "/signup/details" } }, { status: 202 });
    const stray = await callbackRoute.GET(get("/auth/callback?code=c&state=s", undefined, TENANT_HOST));
    assert.equal(stray.headers.get("location"), "/sign-in?status=failed");
    assert.deepEqual(cookies(stray), []);
  });
});
