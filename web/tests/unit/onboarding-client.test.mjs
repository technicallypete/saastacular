// Unit tests for lib/onboarding-client.ts (G018 Plan r5 §1.5, §5.3–§5.4): the exact outbound requests of the eight
// calls (diagnostic headers, the session credential, JSON with the documented fields only), the strict reading of core's
// answers (exact keys, UUIDs, slug grammar, derived completion, ascending pages, per-endpoint byte bounds) and the closed
// refusal reading. fetch is replaced by a recorder, so nothing leaves the process. The lib modules import their siblings
// without extensions, as the Next bundler resolves them; a resolve hook scoped to the lib directory maps them to .ts.
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

const onb = await import("../../lib/onboarding-client.ts");
const { deriveContext, outboundHeaders } = await import("../../lib/request-context.ts");

const CORE = "http://core.internal:8180";
const ORIGIN = "https://org-a.localhost:543";
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const M = { origin: ORIGIN, csrfToken: CSRF };
const CTX = deriveContext(new Headers({ "x-request-id": "req-onboarding-1", traceparent: `00-${"a".repeat(32)}-${"b".repeat(16)}-01` }));
const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };

const ID = (n) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
// G060 (Plan r0 §4.2, D-9): every Group answer carries self_role, can_manage and management_source.
const NONE = { self_role: null, can_manage: false, management_source: null };
const G1 = { id: ID(1), name: "Field <Team>", slug: "field-team", business_profile_id: null, complete: false, ...NONE };
const G2 = { id: ID(2), name: "Ops", slug: "ops", business_profile_id: ID(9), complete: true, self_role: "group_owner", can_manage: true, management_source: "group_role" };
const S3 = { id: ID(3), display_name: "Acme", industry: null, can_select: true, can_update: false };
const P9 = { id: ID(9), display_name: "Acme", legal_name: null, description: "Line\nbreak", industry: "Tools", website: "https://acme.example" };

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "onboarding_unavailable" }, { status: 503 });
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

function only() {
  assert.equal(calls.length, 1);
  return calls[0];
}

function headersOf(init) {
  return Object.fromEntries(Object.entries(init.headers).sort());
}

const READ_HEADERS = { Authorization: `Session ${SECRET}`, ...outboundHeaders(CTX) };
const UPPER = "00000000-0000-4000-8000-00000000000A";
const WRITE_HEADERS = { ...READ_HEADERS, "Content-Type": "application/json" };

describe("outbound requests", () => {
  it("lists Groups with exactly the origin, limit, cursor and slug query, no body and the read headers", async () => {
    reply = () => Response.json({ groups: [], next: null });
    await onb.listGroups(CTX, SECRET, ORIGIN, { after: ID(5), limit: 50, slug: "ops" });
    const c = only();
    assert.equal(c.url, `${CORE}/internal/v1/groups?origin=${encodeURIComponent(ORIGIN)}&limit=50&after=${ID(5)}&slug=ops`);
    assert.equal(c.init.method, "GET");
    assert.equal(c.init.body, undefined);
    assert.deepEqual(headersOf(c.init), Object.fromEntries(Object.entries(READ_HEADERS).sort()));
    assert.equal(c.init.redirect, "error");
    assert.equal(c.init.cache, "no-store");
  });
  it("refuses a malformed cursor, slug or page size without calling core", async () => {
    for (const page of [{ after: "x" }, { after: UPPER }, { slug: "Ops" }, { slug: "-a" }, { limit: 0 }, { limit: 101 }, { limit: 1.5 }]) {
      assert.deepEqual(await onb.listGroups(CTX, SECRET, ORIGIN, page), { ok: false, error: "invalid_request" }, JSON.stringify(page));
      assert.deepEqual(await onb.listProfiles(CTX, SECRET, ORIGIN, page), { ok: false, error: "invalid_request" }, JSON.stringify(page));
    }
    assert.equal(calls.length, 0);
  });
  it("sends the documented JSON fields of each mutation and nothing else", async () => {
    const input = { display_name: "Acme", legal_name: null, description: null, industry: null, website: null };
    const cases = [
      [() => onb.suggestSlug(CTX, SECRET, M, "Field Team"), "POST", "/internal/v1/groups/slug-suggestion", { origin: ORIGIN, csrf_token: CSRF, name: "Field Team" }],
      [() => onb.createGroup(CTX, SECRET, M, { name: "Ops", slug: null }), "POST", "/internal/v1/groups", { origin: ORIGIN, csrf_token: CSRF, name: "Ops" }],
      [() => onb.createGroup(CTX, SECRET, M, { name: "Ops", slug: "ops" }), "POST", "/internal/v1/groups", { origin: ORIGIN, csrf_token: CSRF, name: "Ops", slug: "ops" }],
      [
        () => onb.completeOnboarding(CTX, SECRET, M, ID(1), { mode: "new", profile: input }),
        "POST",
        `/internal/v1/groups/${ID(1)}/business-profile`,
        { origin: ORIGIN, csrf_token: CSRF, mode: "new", profile: input },
      ],
      [
        () => onb.completeOnboarding(CTX, SECRET, M, ID(1), { mode: "existing", businessProfileId: ID(9) }),
        "POST",
        `/internal/v1/groups/${ID(1)}/business-profile`,
        { origin: ORIGIN, csrf_token: CSRF, mode: "existing", business_profile_id: ID(9) },
      ],
      [() => onb.updateProfile(CTX, SECRET, M, ID(9), input), "PUT", `/internal/v1/business-profiles/${ID(9)}`, { origin: ORIGIN, csrf_token: CSRF, profile: input }],
    ];
    for (const [run, method, path, body] of cases) {
      calls = [];
      await run();
      const c = only();
      assert.equal(c.url, `${CORE}${path}`);
      assert.equal(c.init.method, method);
      assert.deepEqual(JSON.parse(c.init.body), body);
      assert.deepEqual(headersOf(c.init), Object.fromEntries(Object.entries(WRITE_HEADERS).sort()));
    }
  });
  it("reads a Group or profile by id with the origin query, and refuses a non-UUID id without calling core", async () => {
    await onb.getGroup(CTX, SECRET, ORIGIN, ID(1));
    await onb.getProfile(CTX, SECRET, ORIGIN, ID(9));
    assert.deepEqual(
      calls.map((c) => [c.init.method, c.url]),
      [
        ["GET", `${CORE}/internal/v1/groups/${ID(1)}?origin=${encodeURIComponent(ORIGIN)}`],
        ["GET", `${CORE}/internal/v1/business-profiles/${ID(9)}?origin=${encodeURIComponent(ORIGIN)}`],
      ],
    );
    calls = [];
    for (const bad of ["../x", `${ID(1)}/x`, UPPER, ""]) {
      assert.deepEqual(await onb.getGroup(CTX, SECRET, ORIGIN, bad), { ok: false, error: "not_found" });
      assert.deepEqual(await onb.getProfile(CTX, SECRET, ORIGIN, bad), { ok: false, error: "not_found" });
      assert.deepEqual(await onb.updateProfile(CTX, SECRET, M, bad, P9), { ok: false, error: "not_found" });
      assert.deepEqual(await onb.completeOnboarding(CTX, SECRET, M, bad, { mode: "existing", businessProfileId: ID(9) }), { ok: false, error: "not_found" });
    }
    assert.equal(calls.length, 0);
  });
});

describe("answers", () => {
  it("reads a Group page with derived completion and keeps names as text", async () => {
    reply = () => Response.json({ groups: [G1, G2], next: ID(2) });
    assert.deepEqual(await onb.listGroups(CTX, SECRET, ORIGIN), {
      ok: true,
      value: {
        groups: [
          { id: ID(1), name: "Field <Team>", slug: "field-team", businessProfileId: null, complete: false, selfRole: null, canManage: false, managementSource: null },
          { id: ID(2), name: "Ops", slug: "ops", businessProfileId: ID(9), complete: true, selfRole: "group_owner", canManage: true, managementSource: "group_role" },
        ],
        next: ID(2),
      },
    });
  });
  it("reads the four G060 authority cases truthfully: nonmember and member, override off and on", async () => {
    // r1 C7: a nonmember reads metadata only (off) or administers through the override (on); an actual member keeps
    // the real role and, with the override on, administers through it in addition.
    const cases = [
      [{ self_role: null, can_manage: false, management_source: null }, { selfRole: null, canManage: false, managementSource: null }],
      [{ self_role: null, can_manage: true, management_source: "organization_override" }, { selfRole: null, canManage: true, managementSource: "organization_override" }],
      [{ self_role: "group_member", can_manage: false, management_source: null }, { selfRole: "group_member", canManage: false, managementSource: null }],
      [{ self_role: "group_member", can_manage: true, management_source: "organization_override" }, { selfRole: "group_member", canManage: true, managementSource: "organization_override" }],
      [{ self_role: "group_admin", can_manage: true, management_source: "group_role" }, { selfRole: "group_admin", canManage: true, managementSource: "group_role" }],
    ];
    for (const [hints, want] of cases) {
      reply = () => Response.json({ group: { ...G1, ...hints } });
      const r = await onb.getGroup(CTX, SECRET, ORIGIN, ID(1));
      assert.equal(r.ok, true, JSON.stringify(hints));
      assert.deepEqual({ selfRole: r.value.selfRole, canManage: r.value.canManage, managementSource: r.value.managementSource }, want);
    }
  });
  it("treats inconsistent or unknown authority hints as malformed", async () => {
    const bad = [
      { self_role: null, can_manage: true, management_source: null },
      { self_role: null, can_manage: false, management_source: "organization_override" },
      { self_role: null, can_manage: true, management_source: "group_role" },
      { self_role: "group_member", can_manage: true, management_source: "group_role" },
      { self_role: "group_owner", can_manage: true, management_source: "organization_override" },
      { self_role: "org_admin", can_manage: false, management_source: null },
      { self_role: "group_member", can_manage: "false", management_source: null },
      { self_role: null, can_manage: true, management_source: "organization" },
    ];
    for (const hints of bad) {
      reply = () => Response.json({ group: { ...G1, ...hints } });
      assert.deepEqual(await onb.getGroup(CTX, SECRET, ORIGIN, ID(1)), { ok: false, error: "malformed" }, JSON.stringify(hints));
    }
    const missing = Object.fromEntries(Object.entries(G1).filter(([k]) => k !== "self_role"));
    reply = () => Response.json({ group: missing });
    assert.deepEqual(await onb.getGroup(CTX, SECRET, ORIGIN, ID(1)), { ok: false, error: "malformed" });
  });
  it("treats any Group page outside the contract as malformed", async () => {
    const bodies = [
      { groups: [G1], next: null, csrf_token: CSRF },
      { groups: [G2, G1], next: null },
      { groups: [G1, G1], next: null },
      { groups: [G2], next: ID(1) },
      { groups: [{ ...G1, complete: true }], next: null },
      { groups: [{ ...G2, complete: false }], next: null },
      { groups: [{ ...G1, slug: "Field" }], next: null },
      { groups: [{ ...G1, name: "" }], next: null },
      { groups: [{ ...G1, name: "x".repeat(201) }], next: null },
      { groups: [{ ...G1, extra: 1 }], next: null },
      { groups: [{ ...G1, id: "1" }], next: null },
      { groups: Array.from({ length: 101 }, (_, i) => ({ ...G1, id: ID(i + 1), slug: `g${i}` })), next: null },
      { groups: {}, next: null },
      [],
    ];
    for (const body of bodies) {
      reply = () => Response.json(body);
      assert.deepEqual(await onb.listGroups(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" }, JSON.stringify(body).slice(0, 80));
    }
  });
  it("accepts a slug-filtered page only with at most one row of exactly that slug", async () => {
    reply = () => Response.json({ groups: [G2], next: null });
    assert.equal((await onb.listGroups(CTX, SECRET, ORIGIN, { slug: "ops" })).ok, true);
    assert.deepEqual(await onb.listGroups(CTX, SECRET, ORIGIN, { slug: "field-team" }), { ok: false, error: "malformed" });
    reply = () => Response.json({ groups: [G1, G2], next: null });
    assert.deepEqual(await onb.listGroups(CTX, SECRET, ORIGIN, { slug: "ops" }), { ok: false, error: "malformed" });
  });
  it("reads profile summaries with the can_select and can_update hints, including a short page whose next cursor is present", async () => {
    reply = () => Response.json({ business_profiles: [S3], next: ID(7) });
    assert.deepEqual(await onb.listProfiles(CTX, SECRET, ORIGIN), {
      ok: true,
      value: { profiles: [{ id: ID(3), displayName: "Acme", industry: null, canSelect: true, canUpdate: false }], next: ID(7) },
    });
    for (const body of [
      { business_profiles: [{ ...S3, legal_name: null }], next: null },
      { business_profiles: [{ ...S3, display_name: "" }], next: null },
      { business_profiles: [{ id: ID(3), display_name: "Acme", industry: null }], next: null },
      { business_profiles: [{ ...S3, can_select: "true" }], next: null },
      { business_profiles: [{ ...S3, can_update: null }], next: null },
      { business_profiles: [], next: null, csrf_token: CSRF },
    ]) {
      reply = () => Response.json(body);
      assert.deepEqual(await onb.listProfiles(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
  });
  it("filters the picker to can_select rows and never alters the cursor (correction02 §P)", async () => {
    const row = (n, canSelect) => ({ id: ID(n), displayName: `P${n}`, industry: null, canSelect, canUpdate: false });
    // A mixed page keeps next.
    assert.deepEqual(onb.selectableProfiles({ profiles: [row(1, true), row(2, false), row(3, true)], next: ID(3) }), { profiles: [row(1, true), row(3, true)], next: ID(3) });
    // A readable page with no selectable row: empty rows, next intact (the Next link still reaches the following page).
    assert.deepEqual(onb.selectableProfiles({ profiles: [row(4, false), row(5, false)], next: ID(5) }), { profiles: [], next: ID(5) });
    assert.deepEqual(onb.selectableProfiles({ profiles: [], next: null }), { profiles: [], next: null });
    // Through the client: the filtered page's next is the unfiltered page's.
    reply = () => Response.json({ business_profiles: [{ ...S3, can_select: false }, { ...S3, id: ID(4), can_select: true }], next: ID(4) });
    const page = await onb.listProfiles(CTX, SECRET, ORIGIN);
    assert.equal(page.ok, true);
    assert.deepEqual(onb.selectableProfiles(page.value), { profiles: [{ id: ID(4), displayName: "Acme", industry: null, canSelect: true, canUpdate: false }], next: ID(4) });
  });
  it("requires 201 for a create, the requested id for reads and a complete Group after completion", async () => {
    reply = () => Response.json({ group: G1 }, { status: 201 });
    assert.equal((await onb.createGroup(CTX, SECRET, M, { name: G1.name, slug: null })).ok, true);
    reply = () => Response.json({ group: G1 }, { status: 200 });
    assert.deepEqual(await onb.createGroup(CTX, SECRET, M, { name: G1.name, slug: null }), { ok: false, error: "malformed" });
    assert.deepEqual(await onb.getGroup(CTX, SECRET, ORIGIN, ID(2)), { ok: false, error: "malformed" });
    assert.deepEqual(await onb.completeOnboarding(CTX, SECRET, M, ID(1), { mode: "existing", businessProfileId: ID(9) }), { ok: false, error: "malformed" });
    reply = () => Response.json({ group: { ...G1, business_profile_id: ID(9), complete: true } });
    assert.equal((await onb.completeOnboarding(CTX, SECRET, M, ID(1), { mode: "existing", businessProfileId: ID(9) })).ok, true);
    // The create answer reports the creator as the Group's owner (00014 bootstrap) with the matching hints.
    reply = () => Response.json({ group: { ...G1, self_role: "group_owner", can_manage: true, management_source: "group_role" } }, { status: 201 });
    const created = await onb.createGroup(CTX, SECRET, M, { name: G1.name, slug: null });
    assert.equal(created.ok && created.value.selfRole, "group_owner");
    reply = () => Response.json({ slug: "field-team" });
    assert.deepEqual(await onb.suggestSlug(CTX, SECRET, M, "Field Team"), { ok: true, value: "field-team" });
    reply = () => Response.json({ slug: "Field Team" });
    assert.deepEqual(await onb.suggestSlug(CTX, SECRET, M, "Field Team"), { ok: false, error: "malformed" });
  });
  it("reads the shared profile and the can_update display hint exactly", async () => {
    reply = () => Response.json({ business_profile: P9, can_update: false });
    assert.deepEqual(await onb.getProfile(CTX, SECRET, ORIGIN, ID(9)), {
      ok: true,
      value: {
        profile: { id: ID(9), displayName: "Acme", legalName: null, description: "Line\nbreak", industry: "Tools", website: "https://acme.example" },
        canUpdate: false,
      },
    });
    for (const body of [
      { business_profile: P9, can_update: "true" },
      { business_profile: P9 },
      { business_profile: P9, can_update: true, groups: [] },
      { business_profile: { ...P9, id: ID(8) }, can_update: true },
      { business_profile: { ...P9, description: "x".repeat(4001) }, can_update: true },
      { business_profile: { ...P9, legal_name: "" }, can_update: true },
    ]) {
      reply = () => Response.json(body);
      assert.deepEqual(await onb.getProfile(CTX, SECRET, ORIGIN, ID(9)), { ok: false, error: "malformed" });
    }
  });
});

describe("refusals, bounds and transport", () => {
  it("reads closed refusals with a field only for invalid_request and conflict", async () => {
    const cases = [
      [400, { error: "invalid_request", field: "invalid_slug" }, { ok: false, error: "invalid_request", field: "invalid_slug" }],
      [409, { error: "conflict", field: "slug" }, { ok: false, error: "conflict", field: "slug" }],
      [409, { error: "conflict", field: "association_exists" }, { ok: false, error: "conflict", field: "association_exists" }],
      [403, { error: "forbidden" }, { ok: false, error: "forbidden" }],
      [401, { error: "unauthenticated" }, { ok: false, error: "unauthenticated" }],
      [503, { error: "onboarding_unavailable" }, { ok: false, error: "onboarding_unavailable" }],
      [403, { error: "forbidden", field: "x" }, { ok: false, error: "malformed" }],
      [400, { error: "invalid_request", field: "Bad Field" }, { ok: false, error: "malformed" }],
      [400, { error: "invalid_request", field: "x", name: "Ops" }, { ok: false, error: "malformed" }],
      [403, { error: "teapot" }, { ok: false, error: "malformed" }],
      [200, { error: "forbidden" }, { ok: false, error: "malformed" }],
    ];
    for (const [status, body, want] of cases) {
      reply = () => Response.json(body, { status });
      assert.deepEqual(await onb.createGroup(CTX, SECRET, M, { name: "Ops", slug: "ops" }), want, JSON.stringify(body));
    }
  });
  it("pins the per-endpoint success bounds to the documented arithmetic", () => {
    // Worst-case Go-encoded rows (≤ 6 bytes per code point): a Group row and a profile summary row. G060: the Group row's
    // three hints at their longest and the profile row's two hints.
    const groupHints = ',"self_role":"group_member"'.length + ',"can_manage":false'.length + ',"management_source":"organization_override"'.length;
    const groupRow = 2 * 36 + 6 * onb.MAX_NAME + onb.MAX_SLUG + 71 + groupHints;
    const profileRow = 36 + 2 * 6 * onb.PROFILE_MAXIMA.display_name + 45 + ',"can_select":false'.length + ',"can_update":false'.length;
    assert.equal(groupHints, 90);
    assert.ok(groupRow <= onb.GROUP_ROW_MAX_BYTES);
    assert.ok(profileRow <= onb.PROFILE_ROW_MAX_BYTES);
    assert.ok(onb.GROUP_LIST_MAX_BYTES < 160 * 1024);
    assert.ok(onb.PROFILE_LIST_MAX_BYTES < 330 * 1024);
    const single = 3 * 6 * 255 + 6 * onb.PROFILE_MAXIMA.description + 6 * onb.PROFILE_MAXIMA.website + 200;
    assert.ok(single < onb.PROFILE_MAX_BYTES);
    assert.equal(onb.MAX_PAGE, 100);
    assert.equal(onb.PAGE_SIZE, 50);
  });
  it("treats an answer over its bound as malformed, never truncated", async () => {
    const pad = (body, size) => {
      const json = JSON.stringify(body);
      return `${json}${" ".repeat(size - Buffer.byteLength(json))}`;
    };
    reply = () => new Response(pad({ groups: [G1], next: null }, onb.GROUP_LIST_MAX_BYTES), { status: 200 });
    assert.equal((await onb.listGroups(CTX, SECRET, ORIGIN)).ok, true);
    reply = () => new Response(pad({ groups: [G1], next: null }, onb.GROUP_LIST_MAX_BYTES + 1), { status: 200 });
    assert.deepEqual(await onb.listGroups(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    reply = () => new Response(pad({ group: G1 }, onb.GROUP_MAX_BYTES + 1), { status: 200 });
    assert.deepEqual(await onb.getGroup(CTX, SECRET, ORIGIN, ID(1)), { ok: false, error: "malformed" });
    reply = () => new Response(pad({ error: "forbidden" }, 64 * 1024 + 1), { status: 403 });
    assert.deepEqual(await onb.getGroup(CTX, SECRET, ORIGIN, ID(1)), { ok: false, error: "malformed" });
    reply = () => new Response("not json", { status: 200 });
    assert.deepEqual(await onb.getGroup(CTX, SECRET, ORIGIN, ID(1)), { ok: false, error: "malformed" });
  });
  it("maps a failed connection to unreachable and an elapsed deadline to timeout", async () => {
    reply = () => {
      throw new TypeError("fetch failed");
    };
    assert.deepEqual(await onb.getGroup(CTX, SECRET, ORIGIN, ID(1)), { ok: false, error: "unreachable" });
    reply = () => {
      throw new DOMException("deadline", "TimeoutError");
    };
    assert.deepEqual(await onb.createGroup(CTX, SECRET, M, { name: "Ops", slug: null }), { ok: false, error: "timeout" });
    assert.equal(onb.outcomeUnknown("timeout"), true);
    assert.equal(onb.outcomeUnknown("unreachable"), true);
    assert.equal(onb.outcomeUnknown("malformed"), true);
    assert.equal(onb.outcomeUnknown("onboarding_unavailable"), true);
    assert.equal(onb.outcomeUnknown("forbidden"), false);
    assert.equal(onb.outcomeUnknown("conflict"), false);
  });
});
