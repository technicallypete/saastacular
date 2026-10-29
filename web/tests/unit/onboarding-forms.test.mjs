// Unit tests for the onboarding form helpers of lib/onboarding-client.ts and the three onboarding form routes (G018
// Plan r5 §1.4, §1.6): the profile form budget pinned from the exported maxima, a max-field all-astral submission
// (accepted and forwarded intact) and a PROFILE_FORM_BYTES + 1 submission (400, nothing forwarded), the closed field-code
// messages, and each route's closed outcomes (redirects with closed codes, honest conflict and unknown-state pages that
// never claim success), and (G018 M1) the inline slug_taken redirect, the advisory GET /groups/address-check route and
// the client hint logic with its stale-answer guards. fetch is replaced by a recorder, so nothing leaves the process. Resolve hooks map the lib
// modules' extensionless sibling imports and the routes' "@/lib/" alias to the .ts sources.
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { registerHooks } from "node:module";
import { afterEach, beforeEach, describe, it, mock } from "node:test";

import ts from "typescript";

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
  // G018 M1: the advisory client component (TSX) is transpiled with the project's TypeScript to test its hint logic.
  load(url, context, nextLoad) {
    if (/\/components\/onboarding\/address-check\.tsx$/.test(url)) {
      const out = ts.transpileModule(readFileSync(new URL(url), "utf8"), {
        compilerOptions: { jsx: ts.JsxEmit.ReactJSX, module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2022 },
      });
      return { format: "module", source: out.outputText, shortCircuit: true };
    }
    return nextLoad(url, context);
  },
});

const onb = await import("../../lib/onboarding-client.ts");
const { cookieName } = await import("../../lib/session.ts");
const createRoute = await import("../../app/groups/create/route.ts");
const checkRoute = await import("../../app/groups/address-check/route.ts");
const advisory = await import("../../components/onboarding/address-check.tsx");
const completeRoute = await import("../../app/groups/[id]/onboarding/complete/route.ts");
const updateRoute = await import("../../app/business-profiles/[id]/update/route.ts");

const CORE = "http://core.internal:8180";
const HOST = "org-a.localhost:543";
const ORIGIN = `https://${HOST}`;
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const ENV = {
  APP_ENV: "test",
  IDENTITY_COOKIE_PROFILE: "https",
  PRODUCT_NAME: "Test Product",
  PLATFORM_BASE_DOMAIN: "example.test",
  CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test",
  CORE_INTERNAL_URL: CORE,
};
const ID = (n) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
// G060 (Plan r0 §4.2): the caller's own Group-owner hints, as core answers its creator.
const GROUP = { id: ID(1), name: "Field <Team>", slug: "field-team", business_profile_id: null, complete: false, self_role: "group_owner", can_manage: true, management_source: "group_role" };
const DONE = { ...GROUP, business_profile_id: ID(9), complete: true };
const ASTRAL = "\u{1F600}";

let saved;
let calls;
let replies;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  replies = [];
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), init });
    const next = replies.shift();
    return next === undefined ? Response.json({ error: "onboarding_unavailable" }, { status: 503 }) : next();
  });
});

afterEach(() => {
  mock.restoreAll();
  for (const [key, value] of Object.entries(saved)) {
    if (value === undefined) delete process.env[key];
    else process.env[key] = value;
  }
});

function post(path, body, headers = {}) {
  return new Request(`${ORIGIN}${path}`, {
    method: "POST",
    headers: {
      host: HOST,
      origin: ORIGIN,
      "content-type": "application/x-www-form-urlencoded",
      cookie: `${cookieName("session", "https")}=${SECRET}`,
      ...headers,
    },
    body,
  });
}

const params = (id) => ({ params: Promise.resolve({ id }) });
const form = (fields) => new URLSearchParams(fields).toString();
const MAX_ASTRAL = Object.fromEntries(onb.PROFILE_FIELDS.map((name) => [name, ASTRAL.repeat(onb.PROFILE_MAXIMA[name])]));

describe("profile form budget", () => {
  it("is derived from the exported maxima and covers every accepted value", () => {
    const total = onb.PROFILE_FIELDS.reduce((sum, name) => sum + onb.PROFILE_MAXIMA[name], 0);
    assert.equal(total, 6_813);
    assert.equal(total * onb.FORM_BYTES_PER_CODE_POINT + onb.FORM_OVERHEAD_BYTES, 82_780);
    assert.ok(total * onb.FORM_BYTES_PER_CODE_POINT + onb.FORM_OVERHEAD_BYTES < onb.PROFILE_FORM_BYTES);
    assert.equal(onb.PROFILE_FORM_BYTES, 96 * 1024);
    // An astral code point costs exactly the per-code-point maximum in the form encoding.
    assert.equal(new URLSearchParams({ a: ASTRAL }).toString().length - 2, onb.FORM_BYTES_PER_CODE_POINT);
  });
  it("accepts a max-field all-astral submission and forwards it intact", async () => {
    const body = form({ csrf_token: CSRF, ...MAX_ASTRAL });
    assert.ok(Buffer.byteLength(body) > 16 * 1024 && Buffer.byteLength(body) <= onb.PROFILE_FORM_BYTES);
    replies.push(() => Response.json({ business_profile: { id: ID(9), ...MAX_ASTRAL } }));
    const res = await updateRoute.POST(post(`/business-profiles/${ID(9)}/update`, body), params(ID(9)));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), `/business-profiles/${ID(9)}?status=updated`);
    assert.equal(calls.length, 1);
    assert.deepEqual(JSON.parse(calls[0].init.body), { origin: ORIGIN, csrf_token: CSRF, profile: MAX_ASTRAL });

    calls = [];
    replies.push(() => Response.json({ group: DONE }));
    const res2 = await completeRoute.POST(post(`/groups/${ID(1)}/onboarding/complete`, form({ csrf_token: CSRF, mode: "new", ...MAX_ASTRAL })), params(ID(1)));
    assert.equal(res2.status, 303);
    assert.equal(res2.headers.get("location"), `/groups/${ID(1)}?status=completed`);
    assert.deepEqual(JSON.parse(calls[0].init.body), { origin: ORIGIN, csrf_token: CSRF, mode: "new", profile: MAX_ASTRAL });
  });
  it("refuses a PROFILE_FORM_BYTES + 1 submission with 400 and forwards nothing", async () => {
    const prefix = `csrf_token=${CSRF}&display_name=`;
    const body = prefix + "x".repeat(onb.PROFILE_FORM_BYTES + 1 - prefix.length);
    assert.equal(Buffer.byteLength(body), onb.PROFILE_FORM_BYTES + 1);
    const res = await updateRoute.POST(post(`/business-profiles/${ID(9)}/update`, body), params(ID(9)));
    assert.equal(res.status, 400);
    const res2 = await completeRoute.POST(post(`/groups/${ID(1)}/onboarding/complete`, `mode=new&${body}`), params(ID(1)));
    assert.equal(res2.status, 400);
    assert.equal(calls.length, 0);
  });
  it("keeps the 16 KiB default for the Group create form", async () => {
    const res = await createRoute.POST(post("/groups/create", `csrf_token=${CSRF}&name=${"x".repeat(16 * 1024)}`));
    assert.equal(res.status, 400);
    assert.equal(calls.length, 0);
  });
});

describe("form helpers", () => {
  it("reads the five profile fields: CRLF undone in the description, empty optional fields null, repeats refused", () => {
    const fields = new URLSearchParams({ display_name: "Acme", legal_name: "", description: "a\r\nb\r\n", industry: "Tools" });
    const get = (name) => (fields.getAll(name).length > 1 ? Symbol("ambiguous") : (fields.get(name) ?? undefined));
    assert.deepEqual(onb.readProfileInput(get), { display_name: "Acme", legal_name: null, description: "a\nb\n", industry: "Tools", website: null });
    fields.append("website", "https://a.example");
    fields.append("website", "https://b.example");
    assert.equal(onb.readProfileInput(get), undefined);
    assert.deepEqual(onb.readProfileInput(() => undefined), { display_name: "", legal_name: null, description: null, industry: null, website: null });
  });
  it("maps exactly the closed field codes to their field and message", () => {
    for (const [code, fieldName] of [
      ["invalid_name", "name"],
      ["invalid_slug", "slug"],
      ["slug_reserved", "slug"],
      ["slug_taken", "slug"],
      ["invalid_display_name", "display_name"],
      ["invalid_legal_name", "legal_name"],
      ["invalid_description", "description"],
      ["invalid_industry", "industry"],
      ["invalid_website", "website"],
    ]) {
      assert.equal(onb.fieldError(code)?.field, fieldName, code);
      assert.equal(typeof onb.fieldError(code)?.message, "string");
    }
    for (const code of ["__proto__", "constructor", "toString", "slug", "", undefined, 1, "<b>"]) assert.equal(onb.fieldError(code), undefined);
    assert.equal(onb.fieldErrorId("website"), "website-error");
  });
});

describe("Group create route", () => {
  it("creates and continues to onboarding, sending the name and the slug", async () => {
    replies.push(() => Response.json({ group: GROUP }, { status: 201 }));
    const res = await createRoute.POST(post("/groups/create", form({ csrf_token: CSRF, name: "Field <Team>", slug: "field-team" })));
    assert.equal(res.status, 303);
    assert.equal(res.headers.get("location"), `/groups/${ID(1)}/onboarding`);
    assert.deepEqual(JSON.parse(calls[0].init.body), { origin: ORIGIN, csrf_token: CSRF, name: "Field <Team>", slug: "field-team" });
  });
  it("returns a refused name or address as its closed code, never the refused address", async () => {
    replies.push(() => Response.json({ error: "invalid_request", field: "invalid_name" }, { status: 400 }));
    let res = await createRoute.POST(post("/groups/create", form({ csrf_token: CSRF, name: "x", slug: "ok" })));
    assert.equal(res.headers.get("location"), "/groups/new?error=invalid_name");
    replies.push(() => Response.json({ error: "invalid_request", field: "slug_reserved" }, { status: 400 }));
    res = await createRoute.POST(post("/groups/create", form({ csrf_token: CSRF, name: "Ops & Co", slug: "admin" })));
    assert.equal(res.headers.get("location"), "/groups/new?error=slug_reserved&name=Ops%20%26%20Co");
    replies.push(() => Response.json({ error: "invalid_request", field: "unknown_code" }, { status: 400 }));
    res = await createRoute.POST(post("/groups/create", form({ csrf_token: CSRF, name: "Ops", slug: "ops" })));
    assert.equal(res.status, 400);
  });
  it("returns core's definitive conflict to the form as slug_taken (PRG), echoing only a grammar-valid address", async () => {
    // G018 M1 (D3): one core call, no re-read here (the create page makes the fresh authorized holder read).
    for (const [slug, location] of [
      ["field-team", "/groups/new?error=slug_taken&name=Ops%20%26%20Co&slug=field-team"],
      ["", "/groups/new?error=slug_taken&name=Ops%20%26%20Co"],
      ["Not A Slug", "/groups/new?error=slug_taken&name=Ops%20%26%20Co"],
    ]) {
      calls = [];
      replies.push(() => Response.json({ error: "conflict", field: "slug" }, { status: 409 }));
      const res = await createRoute.POST(post("/groups/create", form({ csrf_token: CSRF, name: "Ops & Co", slug })));
      assert.equal(res.status, 303, slug);
      assert.equal(res.headers.get("location"), location);
      assert.equal(res.headers.get("cache-control"), "no-store");
      assert.equal(calls.length, 1);
    }
    // A conflict for an over-long name keeps no name (the form re-checks a name with core only when it may be valid).
    replies.push(() => Response.json({ error: "conflict", field: "slug" }, { status: 409 }));
    const res = await createRoute.POST(post("/groups/create", form({ csrf_token: CSRF, name: "n".repeat(onb.MAX_NAME + 1), slug: "ops" })));
    assert.equal(res.headers.get("location"), "/groups/new?error=slug_taken&slug=ops");
    // slug_taken is web-local: core sending it as a field code is not a known refusal.
    replies.push(() => Response.json({ error: "invalid_request", field: "slug_taken" }, { status: 400 }));
    const forged = await createRoute.POST(post("/groups/create", form({ csrf_token: CSRF, name: "Ops", slug: "ops" })));
    assert.equal(forged.status, 400);
  });
  it("keeps the inline conflict copy honest and the advisory copy free of availability claims", () => {
    assert.equal(onb.fieldError("slug_taken")?.field, "slug");
    assert.match(onb.fieldError("slug_taken").message, /already in use/);
    assert.equal(onb.CREATE_CONFLICT.summary, "Nothing was created by this request.");
    for (const text of [onb.fieldError("slug_taken").message, ...Object.values(onb.CREATE_CONFLICT), ...Object.values(onb.ADDRESS_ADVISORY)]) {
      assert.doesNotMatch(text, /was created\.|successfully|recovered|availab/i, text);
    }
    assert.deepEqual(Object.keys(onb.ADDRESS_ADVISORY).sort(), ["in_use", "invalid", "not_seen"]);
    assert.equal(onb.ADDRESS_ADVISORY.invalid, onb.fieldError("invalid_slug").message);
  });
  it("shows the unknown-state page when the outcome may have committed, without retrying", async () => {
    for (const reply of [
      () => Response.json({ error: "onboarding_unavailable" }, { status: 503 }),
      () => new Response("{", { status: 201 }),
      () => {
        throw new TypeError("fetch failed");
      },
    ]) {
      calls = [];
      replies.push(reply);
      const res = await createRoute.POST(post("/groups/create", form({ csrf_token: CSRF, name: "Ops", slug: "ops" })));
      assert.equal(res.status, 503);
      assert.match(await res.text(), /may or may not have completed/);
      assert.equal(calls.length, 1);
    }
  });
  it("refuses a request without a CSRF field, with a repeated field or without a session before calling core", async () => {
    for (const [body, headers, status] of [
      [form({ name: "Ops" }), {}, 400],
      ["csrf_token=a&csrf_token=b&name=Ops", {}, 400],
      [`csrf_token=${CSRF}&name=Ops&name=Two`, {}, 400],
      [form({ csrf_token: CSRF, name: "Ops" }), { cookie: "" }, 401],
      [form({ csrf_token: CSRF, name: "Ops" }), { origin: "https://evil.example" }, 403],
      [form({ csrf_token: CSRF, name: "Ops" }), { "content-type": "application/json" }, 415],
    ]) {
      const res = await createRoute.POST(post("/groups/create", body, headers));
      assert.equal(res.status, status, body);
    }
    assert.equal(calls.length, 0);
    assert.equal(createRoute.GET().status, 405);
    assert.equal(createRoute.GET().headers.get("allow"), "POST");
  });
});

// G018 M1 (D2): the advisory address check route and the client hint logic.
function get(path, headers = {}) {
  return new Request(`${ORIGIN}${path}`, { headers: { host: HOST, cookie: `${cookieName("session", "https")}=${SECRET}`, ...headers } });
}
const SESSION = {
  user_id: ID(5),
  display_name: null,
  mfa_assurance: "unknown",
  auth_time: "2026-10-17T10:00:00Z",
  expires_at: "2026-10-18T10:00:00Z",
  csrf_token: CSRF,
  membership: "member",
};

describe("address check route", () => {
  it("answers exactly {state, slug}: in_use for a readable holder, not_seen for none, from one exact-slug lookup", async () => {
    for (const [groups, state] of [
      [[GROUP], "in_use"],
      [[], "not_seen"],
    ]) {
      calls = [];
      replies.push(() => Response.json({ groups, next: null }));
      const res = await checkRoute.GET(get("/groups/address-check?slug=field-team"));
      assert.equal(res.status, 200);
      assert.deepEqual(await res.json(), { state, slug: "field-team" });
      assert.equal(res.headers.get("cache-control"), "no-store");
      assert.equal(res.headers.get("x-content-type-options"), "nosniff");
      assert.equal(res.headers.get("referrer-policy"), "no-referrer");
      assert.equal(res.headers.get("access-control-allow-origin"), null);
      assert.equal(calls.length, 1);
      assert.match(calls[0].url, /\/internal\/v1\/groups\?origin=[^&]+&limit=1&slug=field-team$/);
      assert.equal(calls[0].init.headers.Authorization, `Session ${SECRET}`);
      assert.match(calls[0].init.headers["X-Request-Id"], /^[0-9a-f]{32}$/);
    }
  });
  it("answers invalid without a Group lookup once core confirms the session; a non-member is refused", async () => {
    replies.push(() => Response.json(SESSION));
    let res = await checkRoute.GET(get("/groups/address-check?slug=Not_A_Slug"));
    assert.equal(res.status, 200);
    assert.deepEqual(await res.json(), { state: "invalid", slug: "Not_A_Slug" });
    assert.equal(calls.length, 1);
    assert.match(calls[0].url, /\/internal\/v1\/auth\/session\?origin=/);
    replies.push(() => Response.json({ ...SESSION, membership: "none" }));
    res = await checkRoute.GET(get("/groups/address-check?slug=-x"));
    assert.equal(res.status, 403);
    assert.deepEqual(await res.json(), { error: "not_available" });
  });
  it("refuses no cookie, a missing, empty, over-long, repeated or extra parameter and other methods before calling core", async () => {
    for (const [path, headers, status] of [
      ["/groups/address-check?slug=ops", { cookie: "" }, 401],
      ["/groups/address-check?slug=ops", { host: "bad host" }, 401],
      ["/groups/address-check", {}, 400],
      ["/groups/address-check?slug=", {}, 400],
      [`/groups/address-check?slug=${"a".repeat(onb.MAX_SLUG + 1)}`, {}, 400],
      ["/groups/address-check?slug=ops&slug=dev", {}, 400],
      ["/groups/address-check?slug=ops&limit=5", {}, 400],
      ["/groups/address-check?name=ops", {}, 400],
    ]) {
      const res = await checkRoute.GET(get(path, headers));
      assert.equal(res.status, status, path);
      assert.equal(Object.keys(await res.json()).join(), "error", path);
    }
    assert.equal(calls.length, 0);
    for (const method of ["POST", "PUT", "PATCH", "DELETE", "OPTIONS"]) {
      const res = checkRoute[method]();
      assert.equal(res.status, 405);
      assert.equal(res.headers.get("allow"), "GET");
    }
  });
  it("maps a stale or foreign session to 401 and an upstream failure to 503, never not_seen", async () => {
    for (const [reply, status, error] of [
      [() => Response.json({ error: "unauthenticated" }, { status: 401 }), 401, "unauthenticated"],
      [() => Response.json({ error: "not_available" }, { status: 403 }), 403, "not_available"],
      [() => Response.json({ error: "onboarding_unavailable" }, { status: 503 }), 503, "unavailable"],
      [() => Response.json({ groups: [{ ...GROUP, slug: "other" }], next: null }), 503, "unavailable"],
      [() => new Response("{", { status: 200 }), 503, "unavailable"],
      [
        () => {
          throw new TypeError("fetch failed");
        },
        503,
        "unavailable",
      ],
    ]) {
      replies.push(reply);
      const res = await checkRoute.GET(get("/groups/address-check?slug=field-team"));
      assert.equal(res.status, status);
      const body = await res.json();
      assert.deepEqual(body, { error }, JSON.stringify(body));
    }
    // A stale session on the invalid path is 401 too.
    replies.push(() => Response.json({ error: "unauthenticated" }, { status: 401 }));
    const res = await checkRoute.GET(get("/groups/address-check?slug=Bad"));
    assert.equal(res.status, 401);
  });
});

describe("address advisory hint logic", () => {
  const answer = (body, init = {}) => new Response(typeof body === "string" ? body : JSON.stringify(body), { status: 200, headers: { "content-type": "application/json" }, ...init });
  // A fetcher whose answers the test releases by hand (it ignores aborts, like a slow server).
  function manual() {
    const pending = [];
    const fetcher = (url, init) => new Promise((resolve) => pending.push({ url, init, resolve }));
    return { pending, fetcher };
  }

  it("drops a stale answer after an edit and a newer check, and applies only the exact current answer", async () => {
    const shown = [];
    let field = "first-slug";
    const { pending, fetcher } = manual();
    const checker = advisory.addressChecker((s) => shown.push(s), { maxSlug: 63, current: () => field, fetcher });
    const first = checker.check("first-slug");
    // The person edits the field and leaves it again before the first answer arrives.
    field = "second-slug";
    checker.edited();
    const second = checker.check("second-slug");
    assert.equal(pending.length, 2);
    assert.equal(pending[0].init.signal.aborted, true, "the edit aborts the request in flight");
    // The stale first answer arrives anyway: dropped (its generation is old).
    pending[0].resolve(answer({ state: "in_use", slug: "first-slug" }));
    await first;
    assert.ok(!shown.includes("in_use"), "stale in_use never shown");
    pending[1].resolve(answer({ state: "not_seen", slug: "second-slug" }));
    await second;
    assert.equal(shown.at(-1), "not_seen");
    assert.equal(pending[1].url, "/groups/address-check?slug=second-slug");
    assert.equal(pending[1].init.method, "GET");
    assert.equal(pending[1].init.cache, "no-store");
    assert.equal(pending[1].init.credentials, "same-origin");
    assert.equal(pending[1].init.redirect, "error");
    // Leaving the field again without an edit asks nothing; an edit clears the hint.
    await checker.check("second-slug");
    assert.equal(pending.length, 2);
    checker.edited();
    assert.equal(shown.at(-1), undefined);
  });
  it("applies an answer only when it echoes the exact value the field still holds", async () => {
    const shown = [];
    let field = "ops";
    const { pending, fetcher } = manual();
    const checker = advisory.addressChecker((s) => shown.push(s), { maxSlug: 63, current: () => field, fetcher });
    let run = checker.check("ops");
    pending[0].resolve(answer({ state: "in_use", slug: "dev" }));
    await run;
    assert.equal(shown.at(-1), undefined, "a foreign echo is dropped");
    // The value changed without an input event (e.g. autofill): the answer for the old value is dropped.
    run = checker.check("ops ");
    field = "ops";
    pending[1].resolve(answer({ state: "not_seen", slug: "ops " }));
    await run;
    assert.equal(shown.at(-1), undefined, "the field no longer holds the asked value");
    assert.equal(pending[1].url, "/groups/address-check?slug=ops%20", "the exact, untrimmed value is asked about");
  });
  it("clears the hint on a refusal, an unexpected shape, an oversized answer or a timeout; over-long values ask nothing", async () => {
    for (const res of [
      answer({ error: "unauthenticated" }, { status: 401 }),
      answer({ state: "available", slug: "ops" }),
      answer({ state: "in_use", slug: "ops", name: "x" }),
      answer(["in_use"]),
      answer("{"),
      answer({ state: "in_use", slug: "ops" }, { headers: { "content-type": "text/html" } }),
      answer(`{"state":"in_use","slug":"ops"}${" ".repeat(advisory.ADDRESS_CHECK_MAX_BYTES)}`),
    ]) {
      const shown = [];
      const checker = advisory.addressChecker((s) => shown.push(s), { maxSlug: 63, current: () => "ops", fetcher: async () => res });
      await checker.check("ops");
      assert.ok(shown.length >= 1 && shown.every((s) => s === undefined), JSON.stringify(shown));
    }
    const shown = [];
    let aborted = false;
    const hanging = (url, init) =>
      new Promise((_, reject) =>
        init.signal.addEventListener("abort", () => {
          aborted = true;
          reject(new DOMException("aborted", "AbortError"));
        }),
      );
    let checker = advisory.addressChecker((s) => shown.push(s), { maxSlug: 63, current: () => "ops", fetcher: hanging, timeoutMs: 20 });
    await checker.check("ops");
    assert.ok(aborted, "the request has a finite lifetime");
    assert.ok(shown.length >= 1 && shown.every((s) => s === undefined), JSON.stringify(shown));
    let asked = 0;
    const long = "a".repeat(64);
    const out = [];
    checker = advisory.addressChecker((s) => out.push(s), { maxSlug: 63, current: () => long, fetcher: async () => {
        asked += 1;
        return answer({});
      },
    });
    await checker.check(long);
    await checker.check("");
    assert.equal(asked, 0);
    assert.deepEqual(out, [undefined, "invalid", undefined]);
  });
});

describe("completion and update routes", () => {
  it("selects an existing profile, and shows the current state on conflict", async () => {
    replies.push(() => Response.json({ error: "conflict", field: "association_exists" }, { status: 409 }));
    const res = await completeRoute.POST(post(`/groups/${ID(1)}/onboarding/complete`, form({ csrf_token: CSRF, mode: "existing", business_profile_id: ID(9) })), params(ID(1)));
    assert.equal(res.headers.get("location"), `/groups/${ID(1)}?status=already_complete`);
    assert.deepEqual(JSON.parse(calls[0].init.body), { origin: ORIGIN, csrf_token: CSRF, mode: "existing", business_profile_id: ID(9) });
  });
  it("refuses a malformed choice or id before calling core", async () => {
    for (const [id, body, status] of [
      [ID(1), form({ csrf_token: CSRF, mode: "existing", business_profile_id: "x" }), 400],
      [ID(1), form({ csrf_token: CSRF, mode: "other" }), 400],
      [ID(1), form({ csrf_token: CSRF }), 400],
      [ID(1), form({ csrf_token: CSRF, mode: "new", display_name: "A", business_profile_id: ID(9) }), 400],
      ["not-a-uuid", form({ csrf_token: CSRF, mode: "existing", business_profile_id: ID(9) }), 404],
    ]) {
      const res = await completeRoute.POST(post(`/groups/${id}/onboarding/complete`, body), params(id));
      assert.equal(res.status, status, body);
    }
    assert.equal(calls.length, 0);
  });
  it("returns a refused profile field as its closed code and maps other outcomes to fixed pages", async () => {
    replies.push(() => Response.json({ error: "invalid_request", field: "invalid_website" }, { status: 400 }));
    let res = await completeRoute.POST(post(`/groups/${ID(1)}/onboarding/complete`, form({ csrf_token: CSRF, mode: "new", display_name: "A", website: "ftp://x" })), params(ID(1)));
    assert.equal(res.headers.get("location"), `/groups/${ID(1)}/onboarding?error=invalid_website`);
    replies.push(() => Response.json({ error: "invalid_request", field: "invalid_display_name" }, { status: 400 }));
    res = await updateRoute.POST(post(`/business-profiles/${ID(9)}/update`, form({ csrf_token: CSRF, display_name: "" })), params(ID(9)));
    assert.equal(res.headers.get("location"), `/business-profiles/${ID(9)}?error=invalid_display_name`);
    replies.push(() => Response.json({ error: "forbidden" }, { status: 403 }));
    res = await updateRoute.POST(post(`/business-profiles/${ID(9)}/update`, form({ csrf_token: CSRF, display_name: "A" })), params(ID(9)));
    assert.equal(res.status, 403);
    assert.match(await res.text(), /Not permitted.*Nothing was changed/s);
    replies.push(() => Response.json({ error: "onboarding_unavailable" }, { status: 503 }));
    res = await updateRoute.POST(post(`/business-profiles/${ID(9)}/update`, form({ csrf_token: CSRF, display_name: "A" })), params(ID(9)));
    assert.equal(res.status, 503);
    assert.match(await res.text(), /may or may not have been saved/);
    replies.push(() => Response.json({ error: "not_found" }, { status: 404 }));
    res = await completeRoute.POST(post(`/groups/${ID(1)}/onboarding/complete`, form({ csrf_token: CSRF, mode: "existing", business_profile_id: ID(9) })), params(ID(1)));
    assert.equal(res.status, 404);
  });
  it("shows core's refusal of a completion by a viewer who reads but does not manage the Group as Not permitted (G060 r1 C6)", async () => {
    // The page offers no form without can_manage; a forged POST still reaches core, which decides
    // group.complete_onboarding and refuses: nothing is attached and no completion status is claimed.
    for (const body of [form({ csrf_token: CSRF, mode: "existing", business_profile_id: ID(9) }), form({ csrf_token: CSRF, mode: "new", display_name: "A" })]) {
      replies.push(() => Response.json({ error: "forbidden" }, { status: 403 }));
      const res = await completeRoute.POST(post(`/groups/${ID(1)}/onboarding/complete`, body), params(ID(1)));
      assert.equal(res.status, 403, body);
      assert.equal(res.headers.get("location"), null);
      assert.match(await res.text(), /Not permitted.*Nothing was changed/s);
    }
    assert.equal(calls.length, 2);
  });
});
