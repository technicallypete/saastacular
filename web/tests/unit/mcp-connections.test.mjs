// Unit tests for lib/mcp-client.ts (G014): the exact outbound request of every MCP setup call (headers, method, body,
// cache, redirect and timeout) and the validation that reduces core's answers to closed outcomes. fetch is replaced by
// a recorder, so nothing leaves the process. Run by `npm run test:unit` in the web check image with networking
// disabled. A resolve hook scoped to the module maps its extensionless sibling imports to the .ts sources.
import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import { afterEach, beforeEach, describe, it, mock } from "node:test";

registerHooks({
  resolve(specifier, context, nextResolve) {
    if ((specifier === "./config" || specifier === "./request-context") && context.parentURL?.endsWith("/lib/mcp-client.ts")) {
      return nextResolve(`${specifier}.ts`, context);
    }
    return nextResolve(specifier, context);
  },
});

const client = await import("../../lib/mcp-client.ts");
const { deriveContext } = await import("../../lib/request-context.ts");

const CORE = "http://core.internal:8180";
const SECRET = "S".repeat(43);
const CSRF = "C".repeat(43);
const ID = "7f000000-0000-4000-8000-000000000001";
const ORIGIN = "https://org-a.localhost:543";
const URL_A = `${ORIGIN}/mcp/c/${ID}`;
const CTX = deriveContext(new Headers({ "x-request-id": "req-mcp-1", traceparent: `00-${"a".repeat(32)}-${"b".repeat(16)}-01` }));
const ENV = { PRODUCT_NAME: "Test Product", PLATFORM_BASE_DOMAIN: "example.test", CUSTOM_DOMAIN_CNAME_TARGET: "edge.example.test", CORE_INTERNAL_URL: CORE };

let saved;
let calls;
let reply;

beforeEach(() => {
  saved = Object.fromEntries(Object.keys(ENV).map((key) => [key, process.env[key]]));
  Object.assign(process.env, ENV);
  calls = [];
  reply = () => Response.json({ error: "mcp_unavailable" }, { status: 503 });
  mock.method(globalThis, "fetch", async (url, init) => {
    calls.push({ url: String(url), init });
    return reply(url, init);
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

function headerNames(init) {
  return Object.keys(init.headers).map((name) => name.toLowerCase()).sort();
}

// G063: every listed row carries display_name (null for an unnamed or legacy connection).
const PENDING = { connection_id: ID, url: URL_A, state: "pending", client_id: null, created_at: "2026-10-15T12:00:00Z", activated_at: null, display_name: null };
const ACTIVE = { ...PENDING, state: "active", client_id: "client_01ABC", activated_at: "2026-10-15T12:05:00Z" };
const PAGE = Array.from({ length: 100 }, () => ACTIVE);
const NEXT = "czF8MTc2MDUyOTYwMDAwMDAwMH" + "x".repeat(20);

// The largest list answer core can serialize (addendum 01 A1), written as Go's encoder writes it: 100 active rows with
// a 2048-character client ID of "<" (escaped to six bytes each), a 310-byte URL on a 253-character host with port
// 65535, 20-byte timestamps, a 43-character CSRF token, a 128-character cursor and the trailing newline. G063: plus an
// 80-code-point display_name of "<" (each escaped to six bytes, between created_at and state in Go's sorted keys).
const LT6 = String.fromCharCode(92) + "u003c";
const WIDE_ORIGIN = `https://${`${"a".repeat(63)}.`.repeat(3)}${"a".repeat(61)}:65535`;
function widestList() {
  const row = `{"activated_at":"9999-12-31T23:59:59Z","client_id":"${"\\u003c".repeat(2048)}","connection_id":"${ID}",` +
    `"created_at":"9999-12-31T23:59:59Z","display_name":"${LT6.repeat(80)}","state":"active","url":"${WIDE_ORIGIN}/mcp/c/${ID}"}`;
  return `{"connections":[${Array.from({ length: 100 }, () => row).join(",")}],"csrf_token":"${CSRF}","next_cursor":"${"n".repeat(128)}"}\n`;
}

describe("outbound requests", () => {
  it("list sends exactly the diagnostic headers and the credential, with the origin as the only query", async () => {
    reply = () => Response.json({ connections: [PENDING, ACTIVE], csrf_token: CSRF, next_cursor: null });
    const result = await client.listConnections(CTX, SECRET, ORIGIN);
    assert.equal(result.ok, true);
    assert.equal(result.value.csrfToken, CSRF);
    assert.equal(result.value.connections.length, 2);
    assert.deepEqual(result.value.connections[1], { connectionId: ID, url: URL_A, host: "org-a.localhost:543", state: "active", clientId: "client_01ABC",
      createdAt: "2026-10-15T12:00:00Z", activatedAt: "2026-10-15T12:05:00Z", displayName: null });
    const { url, init } = only();
    assert.equal(url, `${CORE}/internal/v1/mcp/connections?origin=${encodeURIComponent(ORIGIN)}`);
    assert.equal(init.method, "GET");
    assert.deepEqual(headerNames(init), ["authorization", "traceparent", "x-request-id"]);
    assert.equal(init.headers.Authorization, `Session ${SECRET}`);
    assert.equal(init.body, undefined);
    assert.equal(init.cache, "no-store");
    assert.equal(init.redirect, "error");
    assert.ok(init.signal instanceof AbortSignal);
    assert.equal(client.MCP_TIMEOUT_MS, 10_000);
  });
  it("list sends a next-page cursor as the second query parameter; a cursor of another shape never reaches core", async () => {
    reply = () => Response.json({ connections: PAGE, csrf_token: CSRF, next_cursor: NEXT });
    const result = await client.listConnections(CTX, SECRET, ORIGIN, "abc_-123");
    assert.equal(result.ok, true);
    assert.equal(result.value.connections.length, 100);
    assert.equal(result.value.nextCursor, NEXT);
    const { url, init } = only();
    assert.equal(url, `${CORE}/internal/v1/mcp/connections?origin=${encodeURIComponent(ORIGIN)}&cursor=abc_-123`);
    assert.deepEqual(headerNames(init), ["authorization", "traceparent", "x-request-id"]);
    for (const bad of ["", "a+b", "a/b", "a=", "a&origin=x", "A".repeat(129)]) {
      assert.deepEqual(await client.listConnections(CTX, SECRET, ORIGIN, bad), { ok: false, error: "invalid_request" });
    }
    assert.equal(calls.length, 1);
    assert.equal(client.isListCursor(NEXT), true);
    assert.equal(client.isListCursor("a=b"), false);
  });
  it("G016: lists the caller's rows of every host of the Organization, each with its own host derived from its URL", async () => {
    const id2 = "7f000000-0000-4000-8000-000000000002";
    const id3 = "7f000000-0000-4000-8000-000000000003";
    const alias = { ...PENDING, connection_id: id2, url: `https://portal.acme.example/mcp/c/${id2}` };
    const ported = { ...ACTIVE, connection_id: id3, url: `https://org-a-old.localhost:8443/mcp/c/${id3}` };
    reply = () => Response.json({ connections: [PENDING, alias, ported], csrf_token: CSRF, next_cursor: null });
    const result = await client.listConnections(CTX, SECRET, ORIGIN);
    assert.equal(result.ok, true);
    assert.deepEqual(result.value.connections.map((c) => [c.connectionId, c.host, c.url]), [
      [ID, "org-a.localhost:543", URL_A],
      [id2, "portal.acme.example", `https://portal.acme.example/mcp/c/${id2}`],
      [id3, "org-a-old.localhost:8443", `https://org-a-old.localhost:8443/mcp/c/${id3}`],
    ]);
    assert.equal(only().url, `${CORE}/internal/v1/mcp/connections?origin=${encodeURIComponent(ORIGIN)}`); // no per-host lookup
  });
  it("accepts the largest list answer core can send and refuses one byte more; other answers keep 64 KiB", async () => {
    const text = widestList();
    assert.equal(Buffer.byteLength(text), client.MCP_LIST_MAX_RESPONSE_BYTES);
    assert.equal(client.MCP_LIST_MAX_RESPONSE_BYTES, 1_327_022); // G063: 1_277_222 + 100 x 498 (Plan r1 §5.3)
    reply = () => new Response(text, { headers: { "content-type": "application/json" } });
    const result = await client.listConnections(CTX, SECRET, WIDE_ORIGIN);
    assert.equal(result.ok, true);
    assert.equal(result.value.connections.length, 100);
    assert.equal(result.value.connections[0].clientId, "<".repeat(2048));
    assert.equal(result.value.connections[0].displayName, "<".repeat(80));
    reply = () => new Response(`${text} `, { headers: { "content-type": "application/json" } });
    assert.deepEqual(await client.listConnections(CTX, SECRET, WIDE_ORIGIN), { ok: false, error: "malformed" });
    const overError = JSON.stringify({ error: "mcp_unavailable" }) + " ".repeat(64 * 1024);
    reply = () => new Response(overError, { status: 503 });
    assert.deepEqual(await client.listConnections(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    reply = () => new Response(overError, { status: 503 });
    assert.deepEqual(await client.createConnection(CTX, { origin: ORIGIN, csrfToken: CSRF }, SECRET), { ok: false, error: "malformed" });
    const overCreate = JSON.stringify({ connection: { connection_id: ID, url: URL_A, state: "pending" } }) + " ".repeat(64 * 1024);
    reply = () => new Response(overCreate, { status: 201 });
    assert.deepEqual(await client.createConnection(CTX, { origin: ORIGIN, csrfToken: CSRF }, SECRET), { ok: false, error: "malformed" });
    reply = () => new Response(JSON.stringify({ outcome: "revoked" }) + " ".repeat(64 * 1024));
    assert.deepEqual(await client.revokeConnection(CTX, { origin: ORIGIN, connectionId: ID, csrfToken: CSRF }, SECRET), { ok: false, error: "malformed" });
    // Within 64 KiB the same padded answers are accepted: the refusals above come from the byte bound alone.
    reply = () => new Response(JSON.stringify({ outcome: "revoked" }) + " ".repeat(1024));
    assert.deepEqual(await client.revokeConnection(CTX, { origin: ORIGIN, connectionId: ID, csrfToken: CSRF }, SECRET), { ok: true, value: "revoked" });
  });
  it("create sends the origin and the CSRF token as JSON", async () => {
    reply = () => Response.json({ connection: { connection_id: ID, url: URL_A, state: "pending" } }, { status: 201 });
    const result = await client.createConnection(CTX, { origin: ORIGIN, csrfToken: CSRF }, SECRET);
    assert.deepEqual(result, { ok: true, value: { connectionId: ID, url: URL_A } });
    const { url, init } = only();
    assert.equal(url, `${CORE}/internal/v1/mcp/connections`);
    assert.equal(init.method, "POST");
    assert.deepEqual(headerNames(init), ["authorization", "content-type", "traceparent", "x-request-id"]);
    assert.deepEqual(JSON.parse(init.body), { origin: ORIGIN, csrf_token: CSRF });
  });
  it("revoke sends the origin, the connection and the CSRF token as JSON", async () => {
    reply = () => Response.json({ outcome: "revoked" });
    const result = await client.revokeConnection(CTX, { origin: ORIGIN, connectionId: ID, csrfToken: CSRF }, SECRET);
    assert.deepEqual(result, { ok: true, value: "revoked" });
    const { url, init } = only();
    assert.equal(url, `${CORE}/internal/v1/mcp/connections/revoke`);
    assert.deepEqual(JSON.parse(init.body), { origin: ORIGIN, connection_id: ID, csrf_token: CSRF });
  });
});

describe("answers reduce to closed outcomes", () => {
  it("passes through every closed error code", async () => {
    for (const code of ["origin_rejected", "unauthenticated", "csrf_failed", "not_available", "not_found", "mcp_unavailable"]) {
      reply = () => Response.json({ error: code }, { status: 403 });
      assert.deepEqual(await client.listConnections(CTX, SECRET, ORIGIN), { ok: false, error: code });
    }
  });
  it("refuses everything outside the contract as malformed", async () => {
    const bad = [
      () => new Response("404 page not found", { status: 404 }), // MCP unconfigured: no route
      () => Response.json({ error: "something_else" }, { status: 400 }),
      () => Response.json({ error: "unauthenticated", detail: "x" }, { status: 401 }),
      // G016: a row of another host is accepted only as a complete connection URL of the row's own id.
      () => Response.json({ connections: [{ ...PENDING, url: "https://org-b.localhost:543/mcp/c/7f000000-0000-4000-8000-000000000002" }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [{ ...PENDING, url: "http://org-b.localhost:543/mcp/c/" + ID }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [{ ...PENDING, url: "https://org-b.localhost:543/x/mcp/c/" + ID }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [{ ...PENDING, url: "https://Org-B.localhost/mcp/c/" + ID }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [{ ...PENDING, url: "https://ops" + "@org-b.localhost/mcp/c/" + ID }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [{ ...PENDING, url: "https://org-b.localhost/mcp/c/" + ID + "/" }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [{ ...PENDING, url: `${ORIGIN}/mcp/c/7f000000-0000-4000-8000-000000000002` }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [{ ...PENDING, state: "revoked" }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [{ ...PENDING, client_id: "client_01X" }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [{ ...ACTIVE, activated_at: null }], csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [PENDING], csrf_token: CSRF, next_cursor: null, extra: true }),
      () => Response.json({ connections: Array.from({ length: 101 }, () => PENDING), csrf_token: CSRF, next_cursor: null }),
      () => Response.json({ connections: [PENDING], next_cursor: null }),
      () => Response.json({ connections: [PENDING], csrf_token: "short", next_cursor: null }),
      () => new Response("x".repeat(70_000)),
      // The next-page cursor: always present, null or an opaque string of the cursor shape on a full page only.
      () => Response.json({ connections: [PENDING], csrf_token: CSRF }),
      () => Response.json({ connections: PAGE, csrf_token: CSRF, next_cursor: 5 }),
      () => Response.json({ connections: PAGE, csrf_token: CSRF, next_cursor: "a+b/c=" }),
      () => Response.json({ connections: PAGE, csrf_token: CSRF, next_cursor: "" }),
      () => Response.json({ connections: PAGE, csrf_token: CSRF, next_cursor: "A".repeat(129) }),
      () => Response.json({ connections: PAGE.slice(1), csrf_token: CSRF, next_cursor: NEXT }),
      () => Response.json({ connections: [...PAGE, PENDING], csrf_token: CSRF, next_cursor: null }),
    ];
    for (const r of bad) {
      reply = r;
      assert.deepEqual(await client.listConnections(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    }
    // Create keeps the exact requesting origin (G016 widens only the listing).
    reply = () => Response.json({ connection: { connection_id: ID, url: `https://portal.acme.example/mcp/c/${ID}`, state: "pending" } }, { status: 201 });
    assert.deepEqual(await client.createConnection(CTX, { origin: ORIGIN, csrfToken: CSRF }, SECRET), { ok: false, error: "malformed" });
    reply = () => Response.json({ connection: { connection_id: ID, url: URL_A, state: "active" } }, { status: 201 });
    assert.deepEqual(await client.createConnection(CTX, { origin: ORIGIN, csrfToken: CSRF }, SECRET), { ok: false, error: "malformed" });
    reply = () => Response.json({ connection: { connection_id: ID, url: URL_A, state: "pending" } }, { status: 200 });
    assert.deepEqual(await client.createConnection(CTX, { origin: ORIGIN, csrfToken: CSRF }, SECRET), { ok: false, error: "malformed" });
    reply = () => Response.json({ outcome: "ok" });
    assert.deepEqual(await client.revokeConnection(CTX, { origin: ORIGIN, connectionId: ID, csrfToken: CSRF }, SECRET), { ok: false, error: "malformed" });
  });
  it("maps transport failures", async () => {
    reply = () => {
      throw new DOMException("timed out", "TimeoutError");
    };
    assert.deepEqual(await client.listConnections(CTX, SECRET, ORIGIN), { ok: false, error: "timeout" });
    reply = () => {
      throw new TypeError("fetch failed");
    };
    assert.deepEqual(await client.listConnections(CTX, SECRET, ORIGIN), { ok: false, error: "unreachable" });
  });
  it("recognizes connection identifiers", () => {
    assert.equal(client.isConnectionId(ID), true);
    for (const v of ["", ID.toUpperCase(), `${ID}x`, "00000000-0000-0000-0000-00000000000"]) assert.equal(client.isConnectionId(v), false);
  });
});

// --- G063: friendly names (Plan r1 §4.1, §5.3, §6; D-6) ----------------------------------------------------------------
// The shared vector table: the same cases as core's TestNormalizeMCPDisplayName (core/internal/identity/mcp_name_test.go).
// Code points are built with String.fromCodePoint so no escape sequence in this file can be rewritten by tooling. Go's
// two byte-level vectors (invalid UTF-8, an encoded surrogate) cannot exist as JavaScript strings; their counterparts
// here are the lone surrogates, and the byte-level refusal is the strictFormField table below (and core's
// TestSetupNameDecoding).
const cp = (...points) => String.fromCodePoint(...points);
const EMOJI80 = cp(0x1f600).repeat(80);
const NAME_VECTORS = [
  { label: "80 BMP", in: "a".repeat(80), want: "a".repeat(80) },
  { label: "80 non-BMP", in: EMOJI80, want: EMOJI80 },
  { label: "padded non-BMP", in: cp(0x3000).repeat(8) + EMOJI80 + cp(0xa0, 0x2000, 0x20), want: EMOJI80 },
  { label: "81 BMP", in: "a".repeat(81), bad: true },
  { label: "81 non-BMP", in: EMOJI80 + cp(0x1f600), bad: true },
  { label: "empty", in: "", want: "" },
  { label: "spaces only", in: "    ", want: "" },
  { label: "Unicode whitespace only", in: cp(0xa0, 0x3000, 0x20, 0x2000, 0x200a, 0x202f, 0x205f, 0x1680), want: "" },
  { label: "outer NBSP", in: `${cp(0xa0)}Laptop${cp(0xa0)}`, want: "Laptop" },
  { label: "outer ideographic space", in: `${cp(0x3000)}Laptop${cp(0x3000)}`, want: "Laptop" },
  { label: "outer en quad and hair space", in: `${cp(0x2000)}Laptop${cp(0x200a)}`, want: "Laptop" },
  { label: "outer narrow NBSP and medium math space", in: `${cp(0x202f)}Laptop${cp(0x205f)}`, want: "Laptop" },
  { label: "outer ogham space", in: `${cp(0x1680)}Laptop `, want: "Laptop" },
  { label: "inner spaces kept", in: `Work  laptop${cp(0xa0)}two`, want: `Work  laptop${cp(0xa0)}two` },
  { label: "BOM kept", in: `${cp(0xfeff)}x`, want: `${cp(0xfeff)}x` },
  { label: "BOM counted", in: cp(0xfeff) + "a".repeat(80), bad: true },
  { label: "BOM not trimmed", in: ` ${cp(0xfeff)} `, want: cp(0xfeff) },
  { label: "markup kept", in: "<script>alert(1)</script>", want: "<script>alert(1)</script>" },
  { label: "duplicate-friendly plain", in: "Claude Desktop", want: "Claude Desktop" },
  { label: "inner newline", in: `a${cp(0x0a)}b`, bad: true },
  { label: "only newline", in: cp(0x0a), bad: true },
  { label: "trailing newline", in: `Laptop${cp(0x0a)}`, bad: true },
  { label: "leading tab", in: `${cp(0x09)}Laptop`, bad: true },
  { label: "leading vertical tab", in: `${cp(0x0b)}Laptop`, bad: true },
  { label: "trailing form feed", in: `Laptop${cp(0x0c)}`, bad: true },
  { label: "trailing carriage return", in: `Laptop${cp(0x0d)}`, bad: true },
  { label: "NUL", in: `Lap${cp(0x00)}top`, bad: true },
  { label: "leading NEL", in: `${cp(0x85)}Laptop`, bad: true },
  { label: "trailing line separator", in: `Laptop${cp(0x2028)}`, bad: true },
  { label: "paragraph separator only", in: cp(0x2029), bad: true },
  { label: "DEL", in: `Laptop${cp(0x7f)}`, bad: true },
  { label: "C1 APC", in: `${cp(0x9f)}Laptop`, bad: true },
  { label: "lone high surrogate", in: `Lap${String.fromCharCode(0xd800)}top`, bad: true },
  { label: "lone low surrogate", in: String.fromCharCode(0xdc00), bad: true },
];

describe("G063 friendly names", () => {
  it("normalizeDisplayNameInput agrees with core on the shared vector table", () => {
    for (const v of NAME_VECTORS) {
      const got = client.normalizeDisplayNameInput(v.in);
      if (v.bad) assert.deepEqual(got, { ok: false }, v.label);
      else assert.deepEqual(got, { ok: true, value: v.want }, v.label);
    }
    assert.equal(client.DISPLAY_NAME_MAX, 80);
    // The padded vector's stored value is exactly 80 code points (160 UTF-16 units): the bound is code points.
    const padded = client.normalizeDisplayNameInput(NAME_VECTORS[2].in);
    assert.equal(Array.from(padded.value).length, 80);
    assert.equal(padded.value.length, 160);
  });
  it("listed names: null for legacy rows, any normalized name as text; anything else is malformed", async () => {
    const named = (display_name) => Response.json({ connections: [{ ...PENDING, display_name }], csrf_token: CSRF, next_cursor: null });
    for (const name of ["Laptop", "<b>bold</b> & co", EMOJI80, `${cp(0xfeff)}x`, "Claude Desktop"]) {
      reply = () => named(name);
      const result = await client.listConnections(CTX, SECRET, ORIGIN);
      assert.equal(result.ok, true, name);
      assert.equal(result.value.connections[0].displayName, name);
    }
    reply = () => named(null);
    assert.equal((await client.listConnections(CTX, SECRET, ORIGIN)).value.connections[0].displayName, null);
    const withoutName = { ...PENDING };
    delete withoutName.display_name;
    const bad = [
      () => Response.json({ connections: [withoutName], csrf_token: CSRF, next_cursor: null }), // the key is always present
      () => named(""),
      () => named(" Laptop"),
      () => named(`Laptop${cp(0xa0)}`),
      () => named("a".repeat(81)),
      () => named(EMOJI80 + cp(0x1f600)),
      () => named(`a${cp(0x0a)}b`),
      () => named(`a${cp(0x85)}b`),
      () => named(`a${cp(0x2028)}b`),
      () => named(5),
      () => named(["Laptop"]),
      () => named({ name: "Laptop" }),
    ];
    for (const r of bad) {
      reply = r;
      assert.deepEqual(await client.listConnections(CTX, SECRET, ORIGIN), { ok: false, error: "malformed" });
    }
  });
  it("create sends display_name only when a name is given", async () => {
    reply = () => Response.json({ connection: { connection_id: ID, url: URL_A, state: "pending" } }, { status: 201 });
    assert.equal((await client.createConnection(CTX, { origin: ORIGIN, csrfToken: CSRF, displayName: "Laptop" }, SECRET)).ok, true);
    assert.deepEqual(JSON.parse(calls[0].init.body), { origin: ORIGIN, csrf_token: CSRF, display_name: "Laptop" });
    assert.equal((await client.createConnection(CTX, { origin: ORIGIN, csrfToken: CSRF }, SECRET)).ok, true);
    assert.deepEqual(JSON.parse(calls[1].init.body), { origin: ORIGIN, csrf_token: CSRF });
    reply = () => Response.json({ error: "invalid_name" }, { status: 400 });
    assert.deepEqual(await client.createConnection(CTX, { origin: ORIGIN, csrfToken: CSRF, displayName: "x" }, SECRET), { ok: false, error: "invalid_name" });
  });
  it("rename sends the origin, the connection, the CSRF token and the name; the answer is exactly renamed", async () => {
    reply = () => Response.json({ outcome: "renamed" });
    const result = await client.renameConnection(CTX, { origin: ORIGIN, connectionId: ID, csrfToken: CSRF, displayName: "" }, SECRET);
    assert.deepEqual(result, { ok: true, value: "renamed" });
    const { url, init } = only();
    assert.equal(url, `${CORE}/internal/v1/mcp/connections/rename`);
    assert.equal(init.method, "POST");
    assert.deepEqual(headerNames(init), ["authorization", "content-type", "traceparent", "x-request-id"]);
    assert.deepEqual(JSON.parse(init.body), { origin: ORIGIN, connection_id: ID, csrf_token: CSRF, display_name: "" });
    for (const code of ["invalid_name", "not_found", "not_available", "csrf_failed", "unauthenticated", "invalid_request", "mcp_unavailable"]) {
      reply = () => Response.json({ error: code }, { status: code === "not_found" ? 404 : 400 });
      assert.deepEqual(await client.renameConnection(CTX, { origin: ORIGIN, connectionId: ID, csrfToken: CSRF, displayName: "x" }, SECRET), { ok: false, error: code });
    }
    for (const r of [() => Response.json({ outcome: "revoked" }), () => Response.json({ outcome: "renamed", name: "x" }), () => Response.json({ outcome: "renamed" }, { status: 201 }),
      () => new Response(JSON.stringify({ outcome: "renamed" }) + " ".repeat(64 * 1024))]) {
      reply = r;
      assert.deepEqual(await client.renameConnection(CTX, { origin: ORIGIN, connectionId: ID, csrfToken: CSRF, displayName: "x" }, SECRET), { ok: false, error: "malformed" });
    }
  });
});

// strictFormField: the name field read from the raw body bytes of a form request (Plan r1 §5.3, r1 C4).
function rawRequest(body) {
  return new Request("https://org-a.localhost:543/connections/mcp/rename", {
    method: "POST",
    headers: { "content-type": "application/x-www-form-urlencoded" },
    body: typeof body === "string" ? new TextEncoder().encode(body) : body,
  });
}

describe("G063 strictFormField", () => {
  it("decodes the one field strictly from raw bytes; malformed encodings are refused, never replaced", async () => {
    const field = (body) => client.strictFormField(rawRequest(body), "display_name");
    assert.deepEqual(await field("csrf_token=x&display_name=Work+laptop"), { kind: "value", value: "Work laptop" });
    assert.deepEqual(await field("display_name=%F0%9F%98%80&x=1"), { kind: "value", value: cp(0x1f600) });
    assert.deepEqual(await field("display_name=a%2Bb%26c%3Dd"), { kind: "value", value: "a+b&c=d" });
    assert.deepEqual(await field("display_name=100%"), { kind: "value", value: "100%" }); // a literal % stays, as URLSearchParams
    assert.deepEqual(await field("display_name=%EF%BB%BFx"), { kind: "value", value: `${cp(0xfeff)}x` }); // a leading BOM is kept
    assert.deepEqual(await field("display_name="), { kind: "value", value: "" });
    assert.deepEqual(await field("display_name"), { kind: "value", value: "" });
    assert.deepEqual(await field("csrf_token=x"), { kind: "absent" });
    assert.deepEqual(await field(""), { kind: "absent" });
    assert.deepEqual(await field("display_names=x&xdisplay_name=y"), { kind: "absent" });
    assert.deepEqual(await field("display_name=a&display_name=b"), { kind: "ambiguous" });
    assert.deepEqual(await field("display_name=a&display%5Fname=b"), { kind: "ambiguous" }); // an encoded key is the same key
    for (const bad of ["display_name=%FF", "display_name=%FF%FE", "display_name=%C0%AF", "display_name=%ED%A0%80", "display_name=%ED%B0%80", "display_name=a%E2%82"]) {
      assert.deepEqual(await field(bad), { kind: "malformed" }, bad);
    }
    // A raw (unencoded) invalid byte in the body is refused the same way.
    assert.deepEqual(await field(Uint8Array.from([...new TextEncoder().encode("display_name="), 0xff])), { kind: "malformed" });
    // Over the form budget (which readForm has already refused) is malformed, never a truncated value.
    assert.deepEqual(await client.strictFormField(rawRequest(`display_name=${"a".repeat(100)}`), "display_name", 64), { kind: "malformed" });
    // The padded 80-non-BMP vector survives the byte layer intact (the route normalizes it afterwards).
    const padded = NAME_VECTORS[2].in;
    assert.deepEqual(await field(new URLSearchParams({ display_name: padded }).toString()), { kind: "value", value: padded });
  });
  it("return_to is a closed selector; status locations never carry a submitted value", () => {
    const t = (q) => client.returnTarget(new URLSearchParams(q));
    assert.equal(t(""), "connections");
    assert.equal(t("return_to=account"), "account");
    assert.equal(t("return_to=connections"), "connections");
    for (const bad of ["return_to=", "return_to=/account", "return_to=https://evil.example/", "return_to=account&return_to=account", "return_to=Account"]) {
      assert.equal(t(bad), undefined, bad);
    }
    assert.equal(client.statusLocation("account", "renamed"), "/account?mcp_status=renamed#mcp-connections");
    assert.equal(client.statusLocation("connections", "invalid_name"), "/connections/mcp?status=invalid_name");
  });
});
