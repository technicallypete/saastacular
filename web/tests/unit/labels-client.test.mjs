// Unit tests for lib/labels-client.ts (G020 Plan r2 §1.5, §4.2, §4.4, T11): the exact outbound requests of the seven
// calls (diagnostic headers, the session credential, JSON with the documented fields only), the strict reading of core's
// answers (the closed key sets LABEL_KEYS and GROUP_LABEL_KEYS, UUIDs, owner-scope-consistent rows, ascending pages, the
// name rule, per-endpoint byte bounds), the closed refusal reading, and the CopyIsExact pin: the transport helpers are
// membership-client.ts's, byte for byte apart from two declared renames. fetch is replaced by a recorder, so nothing
// leaves the process. A resolve hook scoped to the lib directory maps the extensionless sibling imports to .ts.
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
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

const lab = await import("../../lib/labels-client.ts");
const { deriveContext, outboundHeaders } = await import("../../lib/request-context.ts");

const CORE = "http://core.internal:8180";
const ORIGIN = "https://org-a.localhost:543";
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const M = { origin: ORIGIN, csrfToken: CSRF };
const CTX = deriveContext(new Headers({ "x-request-id": "req-labels-1", traceparent: `00-${"a".repeat(32)}-${"b".repeat(16)}-01` }));
const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };
const ID = (n) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
const G = ID(77);
const ORG1 = { id: ID(1), scope: "organization", group_id: null, name: "Priority <A>", visibility: "owner_scope" };
const ORG2 = { id: ID(2), scope: "organization", group_id: null, name: "Zürich", visibility: "owner_scope" };
const GRP1 = { id: ID(3), scope: "group", group_id: G, name: "Shared", visibility: "target_members" };
const GRP2 = { id: ID(4), scope: "group", group_id: G, name: "Admins only", visibility: "owner_scope" };
const READ = () => ({ ...outboundHeaders(CTX), Authorization: `Session ${SECRET}` });
const WRITE = () => ({ ...READ(), "Content-Type": "application/json" });

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "labels_unavailable" }, { status: 503 });
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

const answer = (status, body) => () => new Response(typeof body === "string" ? body : JSON.stringify(body), { status, headers: { "content-type": "application/json" } });

function sent(i = 0) {
  const c = calls[i];
  return { url: c.url, method: c.init.method, headers: c.init.headers, body: c.init.body === undefined ? undefined : JSON.parse(c.init.body) };
}

// The JSON Go's encoder writes for a value: encoding/json escapes "<", ">" and "&" as six-byte sequences (built here by
// concatenation so this file holds no escape sequence itself).
function goJSON(value) {
  const esc = (c) => "\\" + "u00" + c.charCodeAt(0).toString(16).padStart(2, "0");
  return JSON.stringify(value).replace(/[<>&]/g, esc);
}

describe("labels client: CopyIsExact", () => {
  // Each helper's source from "function <name>(" (optionally exported or async) to the first line that is exactly "}".
  function helper(source, name) {
    const m = new RegExp(`^(?:export )?(?:async )?function ${name}\\([\\s\\S]*?^\\}$`, "m").exec(source);
    assert.ok(m, `${name} present`);
    return m[0];
  }
  const RENAMES = [
    ["LABELS_TIMEOUT_MS", "MEMBERSHIP_TIMEOUT_MS"],
    ["LabelsError", "MembershipError"],
  ];

  it("keeps isObject, hasExactly, isId, readBounded, call, failure and ascending identical to membership-client.ts", () => {
    const labels = readFileSync(new URL("../../lib/labels-client.ts", import.meta.url), "utf8");
    const membership = readFileSync(new URL("../../lib/membership-client.ts", import.meta.url), "utf8");
    for (const name of ["isObject", "hasExactly", "isId", "readBounded", "call", "failure", "ascending"]) {
      let copy = helper(labels, name);
      for (const [from, to] of RENAMES) copy = copy.replaceAll(from, to);
      assert.equal(copy, helper(membership, name), name);
    }
  });

  it("keeps the shared constants: timeout, refusal bound, UUID and field-code patterns, closed refusal codes", () => {
    const labels = readFileSync(new URL("../../lib/labels-client.ts", import.meta.url), "utf8");
    const membership = readFileSync(new URL("../../lib/membership-client.ts", import.meta.url), "utf8");
    for (const line of [
      "const MAX_RESPONSE_BYTES = 64 * 1024; // every refusal and any answer without its own bound",
      "const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;",
      "const FIELD_CODE = /^[a-z_]{1,64}$/;",
      'const FIELD_ERRORS: ReadonlySet<string> = new Set(["invalid_request", "forbidden", "conflict"]);',
    ]) {
      assert.ok(labels.includes(line) && membership.includes(line), line);
    }
    assert.equal(lab.LABELS_TIMEOUT_MS, 10_000);
    const codes = (source, list) => new RegExp(`const ${list} = \\[([^\\]]*)\\]`).exec(source)[1].replace(/\s+/g, "");
    assert.equal(codes(labels, "LABELS_ERRORS").replace('"labels_unavailable"', '"membership_unavailable"'), codes(membership, "MEMBERSHIP_ERRORS"));
  });
});

describe("labels client: requests", () => {
  it("lists Organization labels with the exact query, the read headers and no body", async () => {
    reply = answer(200, { labels: [ORG1, ORG2], next: null });
    const r = await lab.listLabels(CTX, SECRET, ORIGIN, "organization", undefined, ID(0));
    assert.deepEqual(r, {
      ok: true,
      value: {
        labels: [
          { id: ID(1), scope: "organization", groupId: null, name: "Priority <A>", visibility: "owner_scope" },
          { id: ID(2), scope: "organization", groupId: null, name: "Zürich", visibility: "owner_scope" },
        ],
        next: null,
      },
    });
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/labels?origin=${encodeURIComponent(ORIGIN)}&scope=organization&limit=100&after=${ID(0)}`);
    assert.equal(s.method, "GET");
    assert.deepEqual(s.headers, READ());
    assert.equal(s.body, undefined);
  });

  it("lists one Group's labels with scope and group_id; refuses a malformed scope or cursor before core", async () => {
    reply = answer(200, { labels: [GRP1, GRP2], next: ID(4) });
    const r = await lab.listLabels(CTX, SECRET, ORIGIN, "group", G);
    assert.equal(r.ok, true);
    assert.deepEqual(r.value.labels.map((l) => [l.id, l.groupId, l.visibility]), [[ID(3), G, "target_members"], [ID(4), G, "owner_scope"]]);
    assert.equal(r.value.next, ID(4));
    assert.equal(sent().url, `${CORE}/internal/v1/labels?origin=${encodeURIComponent(ORIGIN)}&scope=group&group_id=${G}&limit=100`);
    calls = [];
    for (const args of [["group", undefined], ["group", "research"], ["organization", G], ["organization", undefined, "x"]]) {
      assert.deepEqual(await lab.listLabels(CTX, SECRET, ORIGIN, ...args), { ok: false, error: "invalid_request" }, JSON.stringify(args));
    }
    assert.equal(calls.length, 0);
  });

  it("creates an Organization label with exactly origin, csrf_token, scope, name and owner_scope", async () => {
    reply = answer(200, ORG1);
    const r = await lab.createLabel(CTX, SECRET, M, "organization", undefined, "Priority <A>", "owner_scope");
    assert.equal(r.ok, true);
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/labels`);
    assert.equal(s.method, "POST");
    assert.deepEqual(s.headers, WRITE());
    assert.deepEqual(s.body, { origin: ORIGIN, csrf_token: CSRF, scope: "organization", name: "Priority <A>", visibility: "owner_scope" });
    assert.deepEqual(Object.keys(s.body), ["origin", "csrf_token", "scope", "name", "visibility"]);
  });

  it("creates a Group label with group_id and the chosen visibility; the answer must be that Group's row", async () => {
    reply = answer(200, GRP1);
    assert.equal((await lab.createLabel(CTX, SECRET, M, "group", G, "Shared", "target_members")).ok, true);
    assert.deepEqual(sent().body, { origin: ORIGIN, csrf_token: CSRF, scope: "group", group_id: G, name: "Shared", visibility: "target_members" });
    for (const row of [{ ...GRP1, group_id: ID(78) }, ORG1, GRP2]) {
      reply = answer(200, row);
      assert.deepEqual(await lab.createLabel(CTX, SECRET, M, "group", G, "Shared", "target_members"), { ok: false, error: "malformed" }, JSON.stringify(row));
    }
  });

  it("updates with exactly origin, csrf_token, name and visibility; the answer must carry the same id and visibility", async () => {
    reply = answer(200, { ...GRP1, name: "Renamed" });
    const r = await lab.updateLabel(CTX, SECRET, M, ID(3), "group", G, "Renamed", "target_members");
    assert.equal(r.ok, true);
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/labels/${ID(3)}/update`);
    assert.deepEqual(s.body, { origin: ORIGIN, csrf_token: CSRF, name: "Renamed", visibility: "target_members" });
    assert.deepEqual(s.headers, WRITE());
    reply = answer(200, { ...GRP1, id: ID(9) });
    assert.deepEqual(await lab.updateLabel(CTX, SECRET, M, ID(3), "group", G, "Shared", "target_members"), { ok: false, error: "malformed" });
    reply = answer(200, GRP1);
    assert.deepEqual(await lab.updateLabel(CTX, SECRET, M, ID(3), "group", G, "Shared", "owner_scope"), { ok: false, error: "malformed" });
  });

  it("deletes with exactly origin and csrf_token and reads the assignments removed", async () => {
    reply = answer(200, { deleted: true, assignments_removed: 2 });
    assert.deepEqual(await lab.deleteLabel(CTX, SECRET, M, ID(1)), { ok: true, value: { assignmentsRemoved: 2 } });
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/labels/${ID(1)}/delete`);
    assert.deepEqual(s.body, { origin: ORIGIN, csrf_token: CSRF });
    for (const body of [{ deleted: false, assignments_removed: 0 }, { deleted: true }, { deleted: true, assignments_removed: -1 }, { deleted: true, assignments_removed: 1.5 }, { deleted: true, assignments_removed: 0, id: ID(1) }]) {
      reply = answer(200, body);
      assert.deepEqual(await lab.deleteLabel(CTX, SECRET, M, ID(1)), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
  });

  it("lists a Group's assigned labels with the origin query and the read headers", async () => {
    reply = answer(200, { labels: [{ label_id: ID(2), name: "Zürich" }, { label_id: ID(1), name: "Priority <A>" }] });
    assert.deepEqual(await lab.listGroupLabels(CTX, SECRET, ORIGIN, G), {
      ok: true,
      value: [
        { labelId: ID(2), name: "Zürich" },
        { labelId: ID(1), name: "Priority <A>" },
      ],
    });
    const s = sent();
    assert.equal(s.url, `${CORE}/internal/v1/groups/${G}/labels?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(s.method, "GET");
    assert.deepEqual(s.headers, READ());
    assert.equal(s.body, undefined);
  });

  it("assigns and unassigns with exactly origin, csrf_token and label_id; the answer must report the requested state", async () => {
    reply = answer(200, { assigned: true });
    assert.deepEqual(await lab.assignLabel(CTX, SECRET, M, G, ID(1)), { ok: true, value: null });
    assert.equal(sent().url, `${CORE}/internal/v1/groups/${G}/labels/assign`);
    assert.deepEqual(sent().body, { origin: ORIGIN, csrf_token: CSRF, label_id: ID(1) });
    assert.deepEqual(sent().headers, WRITE());
    reply = answer(200, { assigned: false });
    assert.deepEqual(await lab.unassignLabel(CTX, SECRET, M, G, ID(1)), { ok: true, value: null });
    assert.equal(sent(1).url, `${CORE}/internal/v1/groups/${G}/labels/unassign`);
    assert.deepEqual(await lab.assignLabel(CTX, SECRET, M, G, ID(1)), { ok: false, error: "malformed" });
    reply = answer(200, { assigned: true });
    assert.deepEqual(await lab.unassignLabel(CTX, SECRET, M, G, ID(1)), { ok: false, error: "malformed" });
  });
});

describe("labels client: strict answers", () => {
  it("pins the closed DTO key sets", () => {
    assert.deepEqual([...lab.LABEL_KEYS], ["id", "scope", "group_id", "name", "visibility"]);
    assert.deepEqual([...lab.GROUP_LABEL_KEYS], ["label_id", "name"]);
  });

  it("refuses unsorted pages, extra or missing keys, rows of another owner scope and names outside the rule", async () => {
    for (const body of [
      { labels: [ORG2, ORG1], next: null },
      { labels: [{ ...ORG1, extra: 1 }], next: null },
      { labels: [{ id: ID(1), scope: "organization", group_id: null, name: "x" }], next: null },
      { labels: [{ ...ORG1, group_id: G }], next: null },
      { labels: [{ ...ORG1, visibility: "target_members" }], next: null },
      { labels: [{ ...ORG1, scope: "group" }], next: null },
      { labels: [{ ...ORG1, name: "x".repeat(81) }], next: null },
      { labels: [{ ...ORG1, name: "" }], next: null },
      { labels: [{ ...ORG1, name: " padded" }], next: null },
      { labels: [{ ...ORG1, name: "tab\there" }], next: null },
      { labels: [{ ...ORG1, name: `line${String.fromCodePoint(0x2028)}break` }], next: null },
      { labels: [ORG2], next: ID(1) },
      { labels: [ORG1] },
      { labels: Array.from({ length: 101 }, (_, i) => ({ ...ORG1, id: ID(i + 1) })), next: null },
    ]) {
      reply = answer(200, body);
      assert.deepEqual(await lab.listLabels(CTX, SECRET, ORIGIN, "organization"), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
    reply = answer(200, { labels: [{ ...ORG1, name: "😀".repeat(80) }], next: null });
    assert.equal((await lab.listLabels(CTX, SECRET, ORIGIN, "organization")).ok, true, "80 code points beyond the BMP are within the rule");
    for (const body of [{ labels: [{ ...GRP1, group_id: null }], next: null }, { labels: [{ ...GRP1, group_id: ID(78) }], next: null }, { labels: [{ ...GRP1, visibility: "everyone" }], next: null }]) {
      reply = answer(200, body);
      assert.deepEqual(await lab.listLabels(CTX, SECRET, ORIGIN, "group", G), { ok: false, error: "malformed" }, JSON.stringify(body));
    }
  });

  // Round 15 (Codex source review): core stores a name after Go's strings.TrimSpace, i.e. the Unicode White_Space set;
  // JavaScript's trim() differs in exactly U+FEFF (trimmed by JavaScript only, kept and counted by core) and U+0085
  // (trimmed by Go only; a control, refused anywhere by both). A stored name may therefore begin or end with U+FEFF, and
  // every response path must accept it, while a boundary White_Space code point (which core never stores) stays refused.
  it("accepts a stored name with a boundary U+FEFF on the create, rename and both list paths; refuses a boundary White_Space", async () => {
    const bom = String.fromCodePoint(0xfeff);
    const name = `${bom}Priority${bom}`;
    reply = answer(200, { ...ORG1, name });
    assert.deepEqual(await lab.createLabel(CTX, SECRET, M, "organization", undefined, name, "owner_scope"), {
      ok: true,
      value: { id: ID(1), scope: "organization", groupId: null, name, visibility: "owner_scope" },
    });
    reply = answer(200, { ...GRP1, name });
    assert.deepEqual(await lab.updateLabel(CTX, SECRET, M, ID(3), "group", G, name, "target_members"), {
      ok: true,
      value: { id: ID(3), scope: "group", groupId: G, name, visibility: "target_members" },
    });
    for (const stored of [name, bom, `${bom.repeat(39)}x${bom.repeat(40)}`]) {
      reply = answer(200, { labels: [{ ...ORG1, name: stored }], next: null });
      assert.deepEqual((await lab.listLabels(CTX, SECRET, ORIGIN, "organization")).value?.labels[0].name, stored, "U+FEFF counts as one of 80 code points");
      reply = answer(200, { labels: [{ label_id: ID(1), name: stored }] });
      assert.deepEqual(await lab.listGroupLabels(CTX, SECRET, ORIGIN, G), { ok: true, value: [{ labelId: ID(1), name: stored }] });
    }
    // The 17 Space_Separator members of White_Space (the other eight are controls or the separators refused anywhere):
    // core trims each at a boundary before storing, so a row carrying one is not a stored name. U+0085 is a control.
    const zs = [0x20, 0xa0, 0x1680, ...Array.from({ length: 11 }, (_, i) => 0x2000 + i), 0x202f, 0x205f, 0x3000];
    const edge = zs.flatMap((cp) => [`${String.fromCodePoint(cp)}x`, `x${String.fromCodePoint(cp)}`]);
    for (const bad of [...edge, `${String.fromCodePoint(0x85)}x`, `x${String.fromCodePoint(0x85)}`, `${bom.repeat(40)}x${bom.repeat(40)}`]) {
      reply = answer(200, { labels: [{ ...ORG1, name: bad }], next: null });
      assert.deepEqual(await lab.listLabels(CTX, SECRET, ORIGIN, "organization"), { ok: false, error: "malformed" }, JSON.stringify(bad));
      reply = answer(200, { labels: [{ label_id: ID(1), name: bad }] });
      assert.deepEqual(await lab.listGroupLabels(CTX, SECRET, ORIGIN, G), { ok: false, error: "malformed" }, JSON.stringify(bad));
    }
    reply = answer(200, { ...ORG1, name: " Priority" });
    assert.deepEqual(await lab.createLabel(CTX, SECRET, M, "organization", undefined, "Priority", "owner_scope"), { ok: false, error: "malformed" });
    reply = answer(200, { ...GRP1, name: `Priority${String.fromCodePoint(0x3000)}` });
    assert.deepEqual(await lab.updateLabel(CTX, SECRET, M, ID(3), "group", G, "Priority", "target_members"), { ok: false, error: "malformed" });
  });

  it("refuses duplicate, extra-key and over-many Group assignments", async () => {
    const a = { label_id: ID(1), name: "A" };
    for (const body of [
      { labels: [a, a] },
      { labels: [{ ...a, scope: "organization" }] },
      { labels: [a], next: null },
      { labels: Array.from({ length: 201 }, (_, i) => ({ label_id: ID(i + 1), name: "A" })) },
    ]) {
      reply = answer(200, body);
      assert.deepEqual(await lab.listGroupLabels(CTX, SECRET, ORIGIN, G), { ok: false, error: "malformed" }, JSON.stringify(body).slice(0, 80));
    }
  });

  it("budgets: the worst-case rows fit their bound and each endpoint refuses an answer above it", async () => {
    const worst = "<".repeat(lab.MAX_LABEL_NAME);
    const row = { id: ID(1), scope: "group", group_id: G, name: worst, visibility: "target_members" };
    assert.ok(Buffer.byteLength(goJSON(row)) <= lab.LABEL_ROW_MAX_BYTES, `${Buffer.byteLength(goJSON(row))} ≤ ${lab.LABEL_ROW_MAX_BYTES}`);
    assert.ok(Buffer.byteLength(goJSON({ label_id: ID(1), name: worst })) <= lab.GROUP_LABEL_ROW_MAX_BYTES);
    const page = { labels: Array.from({ length: 100 }, (_, i) => ({ ...row, id: ID(i + 1) })), next: ID(100) };
    assert.ok(Buffer.byteLength(goJSON(page)) + 1 <= lab.LABEL_PAGE_MAX_BYTES);
    const assigned = { labels: Array.from({ length: 200 }, (_, i) => ({ label_id: ID(i + 1), name: worst })) };
    assert.ok(Buffer.byteLength(goJSON(assigned)) + 1 <= lab.GROUP_LABELS_MAX_BYTES);
    // A full worst-case page is accepted; padding past the bound is refused.
    reply = () => new Response(goJSON(page) + "\n", { status: 200 });
    assert.equal((await lab.listLabels(CTX, SECRET, ORIGIN, "group", G)).ok, true);
    reply = answer(200, `{"labels":[],"next":null,"pad":"${"x".repeat(lab.LABEL_PAGE_MAX_BYTES)}"}`);
    assert.deepEqual(await lab.listLabels(CTX, SECRET, ORIGIN, "organization"), { ok: false, error: "malformed" });
    reply = answer(200, `{"labels":[],"pad":"${"x".repeat(lab.GROUP_LABELS_MAX_BYTES)}"}`);
    assert.deepEqual(await lab.listGroupLabels(CTX, SECRET, ORIGIN, G), { ok: false, error: "malformed" });
    reply = answer(200, `{"deleted":true,"assignments_removed":0,"pad":"${"x".repeat(200)}"}`);
    assert.deepEqual(await lab.deleteLabel(CTX, SECRET, M, ID(1)), { ok: false, error: "malformed" });
    reply = answer(200, { ...ORG1, name: worst.slice(0, 79) + "x".repeat(200) });
    assert.deepEqual(await lab.createLabel(CTX, SECRET, M, "organization", undefined, "x", "owner_scope"), { ok: false, error: "malformed" });
  });
});

describe("labels client: refusals", () => {
  it("reads closed codes, with a field only for invalid_request, forbidden and conflict", async () => {
    for (const [status, body, want] of [
      [403, { error: "forbidden" }, { ok: false, error: "forbidden" }],
      [404, { error: "not_found" }, { ok: false, error: "not_found" }],
      [409, { error: "conflict", field: "already_assigned" }, { ok: false, error: "conflict", field: "already_assigned" }],
      [400, { error: "invalid_request", field: "organization_context_required" }, { ok: false, error: "invalid_request", field: "organization_context_required" }],
      [503, { error: "labels_unavailable" }, { ok: false, error: "labels_unavailable" }],
      [404, { error: "not_found", field: "label" }, { ok: false, error: "malformed" }],
      [503, { error: "membership_unavailable" }, { ok: false, error: "malformed" }],
      [409, { error: "conflict", field: "Name Taken" }, { ok: false, error: "malformed" }],
      [200, { error: "forbidden" }, { ok: false, error: "malformed" }],
      [500, "not json", { ok: false, error: "malformed" }],
    ]) {
      reply = answer(status, body);
      assert.deepEqual(await lab.assignLabel(CTX, SECRET, M, G, ID(1)), want, JSON.stringify(body));
    }
  });

  it("reports a timeout and an unreachable core as such", async () => {
    mock.restoreAll();
    mock.method(globalThis, "fetch", async () => {
      throw new DOMException("timed out", "TimeoutError");
    });
    assert.deepEqual(await lab.listGroupLabels(CTX, SECRET, ORIGIN, G), { ok: false, error: "timeout" });
    mock.restoreAll();
    mock.method(globalThis, "fetch", async () => {
      throw new TypeError("fetch failed");
    });
    assert.deepEqual(await lab.listGroupLabels(CTX, SECRET, ORIGIN, G), { ok: false, error: "unreachable" });
  });

  it("maps the labels unavailable code onto the frame's unavailable page and keeps the rest", () => {
    assert.equal(lab.labelsPageError("labels_unavailable"), "onboarding_unavailable");
    assert.equal(lab.labelsPageError("not_found"), "not_found");
    for (const e of ["timeout", "unreachable", "malformed", "labels_unavailable"]) assert.equal(lab.labelOutcomeUnknown(e), true, e);
    for (const e of ["forbidden", "not_found", "conflict", "invalid_request"]) assert.equal(lab.labelOutcomeUnknown(e), false, e);
  });
});
