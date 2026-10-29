import "server-only";

import { runtimeConfig } from "./config";
import { errorCategory, outboundHeaders, type RequestContext } from "./request-context";

// Server-only client for core's internal MCP connection setup contract (/internal/v1/mcp/connections; G014 Plan r8
// §3.2). Each call sends exactly the request's diagnostic headers (outboundHeaders(ctx)), "Authorization: Session
// <credential>" and, on mutations, Content-Type: application/json with a JSON body of the documented fields. The
// setup origin is this request's own origin; core resolves the Organization only from its trusted host registry and
// checks the session, the CSRF token and membership itself. Answers are validated and reduced to closed outcomes.

export const MCP_TIMEOUT_MS = 10_000;
const MAX_RESPONSE_BYTES = 64 * 1024;
// The list-success bound (addendum 01 A1): the largest serialized 100-row page from core's JSON encoder (HTML escaping
// on, sorted keys, trailing newline). Per row: a 2048-byte client ID escaped to at most 6 bytes per byte (12,288), a
// resource URL of at most 310 bytes (https:// + a 253-character host + :65535 + /mcp/c/ + UUID) and 171 bytes of fixed
// syntax with 20-byte RFC 3339 UTC timestamps. Page: 100 rows, 99 commas, a 222-byte envelope (43-character CSRF
// token, 128-character cursor) and the newline. Every other answer (create, revoke, rename, any error) keeps 64 KiB.
// G063 (Plan r1 §5.3): each row also carries "display_name": 15 bytes of key, 1 separator and at most 482 bytes of
// value (80 code points each escaped by Go's encoder to a six-byte \u003c form, plus two quotes; a null is 4, non-BMP
// is 4 bytes per code point, and control characters never leave core), so 498 more per row.
const LIST_ITEM_MAX_BYTES = 12_288 + 310 + 171 + 498;
export const MCP_LIST_MAX_RESPONSE_BYTES = 222 + 1 + 99 + 100 * LIST_ITEM_MAX_BYTES;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const RFC3339 = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z$/;
const RESOURCE = /^https:\/\/[a-z0-9.-]+(?::[0-9]{1,5})?\/mcp\/c\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const MAX_CONNECTIONS = 100;
const CSRF = /^[A-Za-z0-9_-]{43}$/;
const CURSOR = /^[A-Za-z0-9_-]{1,128}$/;

const MCP_ERRORS = [
  "invalid_request",
  "ambiguous_parameter",
  "unsupported_media_type",
  "origin_rejected",
  "unauthenticated",
  "csrf_failed",
  "not_available",
  "not_found",
  "mcp_unavailable",
  "invalid_name", // G063 (D-2): a friendly name outside the rule; bounded and non-echoing
] as const;
const KNOWN_ERRORS: ReadonlySet<string> = new Set(MCP_ERRORS);

// A closed setup error code, or the call's own failure: no answer in time, no connection, or an answer outside the
// contract (including an unmounted route: MCP unconfigured).
export type McpError = (typeof MCP_ERRORS)[number] | "timeout" | "unreachable" | "malformed";
export type McpResult<T> = { ok: true; value: T } | { ok: false; error: McpError };

export type ConnectionView = {
  connectionId: string;
  url: string;
  host: string; // G016: the connection's own host, derived from its URL (an Organization's rows may live on several hosts)
  state: "pending" | "active";
  clientId: string | null;
  createdAt: string;
  activatedAt: string | null;
  // G063: the owner's friendly name, presentation metadata only (never identity, client binding or authority); null for
  // an unnamed or legacy connection. Kept separate from clientId, the actually bound client.
  displayName: string | null;
};

type Call = {
  method: "GET" | "POST";
  path: string;
  credential: string;
  body?: Readonly<Record<string, string>>;
  successBytes?: number; // the bound of a 200 answer when it is larger than MAX_RESPONSE_BYTES (the list only)
};
type Answer = { kind: "answered"; status: number; body: unknown } | { kind: "failed"; error: "timeout" | "unreachable" | "malformed" };
type Json = Readonly<Record<string, unknown>>;

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

async function readBounded(response: Response, maxBytes: number): Promise<string | undefined> {
  if (response.body === null) return "";
  const reader = response.body.getReader();
  const chunks: Uint8Array[] = [];
  let size = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.byteLength;
    if (size > maxBytes) {
      await reader.cancel();
      return undefined;
    }
    chunks.push(value);
  }
  return new TextDecoder().decode(Buffer.concat(chunks));
}

async function call(ctx: RequestContext, c: Call): Promise<Answer> {
  const headers: Record<string, string> = { ...outboundHeaders(ctx) };
  headers.Authorization = `Session ${c.credential}`;
  if (c.body !== undefined) headers["Content-Type"] = "application/json";
  let response: Response;
  try {
    response = await fetch(new URL(c.path, runtimeConfig().coreInternalUrl), {
      method: c.method,
      headers,
      body: c.body === undefined ? undefined : JSON.stringify(c.body),
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(MCP_TIMEOUT_MS),
    });
  } catch (error) {
    return { kind: "failed", error: errorCategory(error) === "TimeoutError" ? "timeout" : "unreachable" };
  }
  try {
    const text = await readBounded(response, response.status === 200 && c.successBytes !== undefined ? c.successBytes : MAX_RESPONSE_BYTES);
    if (text === undefined) return { kind: "failed", error: "malformed" };
    return { kind: "answered", status: response.status, body: JSON.parse(text) as unknown };
  } catch (error) {
    if (errorCategory(error) === "TimeoutError") return { kind: "failed", error: "timeout" };
    return { kind: "failed", error: "malformed" };
  }
}

// A non-2xx answer must be exactly {"error": <closed code>}.
function failure(answer: Answer): { ok: false; error: McpError } {
  if (answer.kind === "failed") return { ok: false, error: answer.error };
  const body = answer.body;
  if (answer.status >= 400 && isObject(body) && Object.keys(body).length === 1 && typeof body.error === "string" && KNOWN_ERRORS.has(body.error)) {
    return { ok: false, error: body.error as McpError };
  }
  return { ok: false, error: "malformed" };
}

// A listed connection answer, valid only when its URL is a complete connection URL of its own id. G016 (Plan r4 §1.6):
// the listing on one host of an Organization shows the caller's rows of every host of that Organization, each with its
// own host, so the URL's host may differ from the requesting origin; core alone decides which rows belong.
function connection(value: unknown): ConnectionView | undefined {
  if (!isObject(value)) return undefined;
  const { connection_id: id, url, state, client_id: clientId, created_at: createdAt, activated_at: activatedAt, display_name: displayName } = value;
  if (typeof id !== "string" || !UUID.test(id) || typeof url !== "string" || !RESOURCE.test(url) || !url.endsWith(`/mcp/c/${id}`)) {
    return undefined;
  }
  const host = url.slice("https://".length, url.length - `/mcp/c/${id}`.length);
  if (state !== "pending" && state !== "active") return undefined;
  if (!(clientId === null || (typeof clientId === "string" && clientId.length > 0 && clientId.length <= 2048))) return undefined;
  if (typeof createdAt !== "string" || !RFC3339.test(createdAt)) return undefined;
  if (!(activatedAt === null || (typeof activatedAt === "string" && RFC3339.test(activatedAt)))) return undefined;
  // A pending connection has neither a bound client nor an activation time; an active one has both.
  if (state === "active" ? clientId === null || activatedAt === null : clientId !== null || activatedAt !== null) return undefined;
  // G063: display_name is always present, null or an already-normalized name of 1-80 code points.
  if (!Object.hasOwn(value, "display_name") || !(displayName === null || isDisplayName(displayName))) return undefined;
  return { connectionId: id, url, host, state, clientId, createdAt, activatedAt, displayName };
}

export type ConnectionList = { connections: ConnectionView[]; csrfToken: string; nextCursor: string | null };

// GET /internal/v1/mcp/connections?origin=<origin>[&cursor=<cursor>]: one page (newest first, at most 100) of the
// caller's own live connections in the Organization of origin, with the session's CSRF token for the page's forms (one
// core call per page) and the next page's opaque cursor (null on the last page). A cursor of the wrong shape is refused
// as invalid_request without calling core.
export async function listConnections(
  ctx: RequestContext,
  credential: string,
  origin: string,
  cursor?: string,
): Promise<McpResult<ConnectionList>> {
  if (cursor !== undefined && !CURSOR.test(cursor)) return { ok: false, error: "invalid_request" };
  const query = `origin=${encodeURIComponent(origin)}${cursor === undefined ? "" : `&cursor=${encodeURIComponent(cursor)}`}`;
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/mcp/connections?${query}`, credential, successBytes: MCP_LIST_MAX_RESPONSE_BYTES });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const body = answer.body;
  if (!isObject(body) || Object.keys(body).length !== 3 || !Array.isArray(body.connections) || body.connections.length > MAX_CONNECTIONS ||
    typeof body.csrf_token !== "string" || !CSRF.test(body.csrf_token)) {
    return { ok: false, error: "malformed" };
  }
  const next = body.next_cursor;
  if (!(next === null || (typeof next === "string" && CURSOR.test(next) && body.connections.length === MAX_CONNECTIONS))) {
    return { ok: false, error: "malformed" };
  }
  const out: ConnectionView[] = [];
  for (const item of body.connections) {
    const c = connection(item);
    if (c === undefined) return { ok: false, error: "malformed" };
    out.push(c);
  }
  return { ok: true, value: { connections: out, csrfToken: body.csrf_token, nextCursor: next } };
}

// POST /internal/v1/mcp/connections: a new pending connection of the caller (CSRF-checked by core). G063: display_name is
// sent only when a name is given (an omitted name keeps the unnamed-create request exactly as before).
export async function createConnection(
  ctx: RequestContext,
  input: { origin: string; csrfToken: string; displayName?: string },
  credential: string,
): Promise<McpResult<{ connectionId: string; url: string }>> {
  const request: Record<string, string> = { origin: input.origin, csrf_token: input.csrfToken };
  if (input.displayName !== undefined) request.display_name = input.displayName;
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/mcp/connections", credential, body: request });
  if (answer.kind !== "answered" || answer.status !== 201) return failure(answer);
  const body = answer.body;
  const c = isObject(body) && Object.keys(body).length === 1 && isObject(body.connection) ? body.connection : undefined;
  if (c === undefined || c.state !== "pending" || typeof c.connection_id !== "string" || !UUID.test(c.connection_id) ||
    typeof c.url !== "string" || !RESOURCE.test(c.url) || c.url !== `${input.origin}/mcp/c/${c.connection_id}`) {
    return { ok: false, error: "malformed" };
  }
  return { ok: true, value: { connectionId: c.connection_id, url: c.url } };
}

// POST /internal/v1/mcp/connections/revoke: terminate one own connection (CSRF-checked by core; exact ownership).
export async function revokeConnection(
  ctx: RequestContext,
  input: { origin: string; connectionId: string; csrfToken: string },
  credential: string,
): Promise<McpResult<"revoked">> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/mcp/connections/revoke",
    credential,
    body: { origin: input.origin, connection_id: input.connectionId, csrf_token: input.csrfToken },
  });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const body = answer.body;
  if (!isObject(body) || Object.keys(body).length !== 1 || body.outcome !== "revoked") return { ok: false, error: "malformed" };
  return { ok: true, value: "revoked" };
}

// POST /internal/v1/mcp/connections/rename (G063 Plan r1 §4.3): set, change or (with "") clear the friendly name of one
// own live connection through this host (CSRF-checked by core; current membership required). Core decides the name.
export async function renameConnection(
  ctx: RequestContext,
  input: { origin: string; connectionId: string; csrfToken: string; displayName: string },
  credential: string,
): Promise<McpResult<"renamed">> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/mcp/connections/rename",
    credential,
    body: { origin: input.origin, connection_id: input.connectionId, csrf_token: input.csrfToken, display_name: input.displayName },
  });
  if (answer.kind !== "answered" || answer.status !== 200) return failure(answer);
  const body = answer.body;
  if (!isObject(body) || Object.keys(body).length !== 1 || body.outcome !== "renamed") return { ok: false, error: "malformed" };
  return { ok: true, value: "renamed" };
}

// --- G063 friendly names (Plan r1 §4.1, §5.3, §5.4; D-6) -----------------------------------------------------------------
// The web mirror of core's identity.NormalizeMCPDisplayName, in the same order: refuse a string that is not well formed
// or holds any control (Cc, incl. tab, newline and U+0085), line separator (Zl) or paragraph separator (Zp) anywhere,
// before trimming; then trim the effective Unicode White_Space set (SP, U+00A0, U+1680, U+2000-U+200A, U+202F, U+205F,
// U+3000; U+FEFF is not White_Space and is kept and counted); then bound the trimmed value to 80 code points. An empty
// result means "no name" (a clear on rename). JavaScript's trim() is not used (it also trims U+FEFF), and UTF-16 units
// are not the limit. Core stays authoritative; this only avoids a core call for a name core would refuse.
export const DISPLAY_NAME_MAX = 80;
const NAME_REFUSED = /[\p{Cc}\p{Zl}\p{Zp}]/u;

function isTrimSpace(cp: number): boolean {
  return cp === 0x20 || cp === 0xa0 || cp === 0x1680 || (cp >= 0x2000 && cp <= 0x200a) || cp === 0x202f || cp === 0x205f || cp === 0x3000;
}

export function normalizeDisplayNameInput(raw: string): { ok: true; value: string } | { ok: false } {
  if (!raw.isWellFormed() || NAME_REFUSED.test(raw)) return { ok: false };
  const points = Array.from(raw);
  let start = 0;
  let end = points.length;
  while (start < end && isTrimSpace(points[start].codePointAt(0) ?? -1)) start++;
  while (end > start && isTrimSpace(points[end - 1].codePointAt(0) ?? -1)) end--;
  if (end - start > DISPLAY_NAME_MAX) return { ok: false };
  return { ok: true, value: points.slice(start, end).join("") };
}

// A listed name: a non-empty string already in its normalized form.
function isDisplayName(value: unknown): value is string {
  if (typeof value !== "string" || value.length === 0) return false;
  const n = normalizeDisplayNameInput(value);
  return n.ok && n.value === value;
}

export type StrictField = { kind: "absent" } | { kind: "ambiguous" } | { kind: "malformed" } | { kind: "value"; value: string };

function hexValue(b: number): number {
  if (b >= 0x30 && b <= 0x39) return b - 0x30;
  if (b >= 0x41 && b <= 0x46) return b - 0x37;
  if (b >= 0x61 && b <= 0x66) return b - 0x57;
  return -1;
}

// The form encoding's percent-decoding, on bytes: "+" is a space and %XX one byte; a "%" not followed by two hex digits
// stays literal (as URLSearchParams does).
function percentDecode(bytes: Uint8Array): Buffer {
  const out: number[] = [];
  for (let i = 0; i < bytes.length; i++) {
    const b = bytes[i];
    if (b === 0x2b) {
      out.push(0x20);
    } else if (b === 0x25 && i + 2 < bytes.length && hexValue(bytes[i + 1]) >= 0 && hexValue(bytes[i + 2]) >= 0) {
      out.push(hexValue(bytes[i + 1]) * 16 + hexValue(bytes[i + 2]));
      i += 2;
    } else {
      out.push(b);
    }
  }
  return Buffer.from(out);
}

// One field of an application/x-www-form-urlencoded body read strictly from its raw bytes (G063 Plan r1 §5.4, r1 C4):
// the form routes call it on request.clone() for the name field alone, so readForm and every other field keep their
// existing lossy decoding. Absent, ambiguous (more than one), malformed (invalid UTF-8, overlong forms and encoded
// surrogates are refused by a fatal decoder, never replaced; also a body over the form budget, which readForm has
// already refused) or the value. A leading U+FEFF is kept (ignoreBOM), as core keeps it.
export async function strictFormField(request: Request, name: string, maxBytes = 16_384): Promise<StrictField> {
  let body: Buffer;
  try {
    if (request.body === null) {
      body = Buffer.alloc(0);
    } else {
      const reader = request.body.getReader();
      const chunks: Uint8Array[] = [];
      let size = 0;
      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;
        size += value.byteLength;
        if (size > maxBytes) {
          await reader.cancel();
          return { kind: "malformed" };
        }
        chunks.push(value);
      }
      body = Buffer.concat(chunks);
    }
  } catch {
    return { kind: "malformed" };
  }
  const key = Buffer.from(name, "utf8");
  const values: Buffer[] = [];
  for (let start = 0; start <= body.length; ) {
    let end = body.indexOf(0x26, start);
    if (end === -1) end = body.length;
    if (end > start) {
      const pair = body.subarray(start, end);
      const eq = pair.indexOf(0x3d);
      if (percentDecode(eq === -1 ? pair : pair.subarray(0, eq)).equals(key)) {
        values.push(percentDecode(eq === -1 ? Buffer.alloc(0) : pair.subarray(eq + 1)));
      }
    }
    start = end + 1;
  }
  if (values.length === 0) return { kind: "absent" };
  if (values.length > 1) return { kind: "ambiguous" };
  try {
    return { kind: "value", value: new TextDecoder("utf-8", { fatal: true, ignoreBOM: true }).decode(values[0]) };
  } catch {
    return { kind: "malformed" };
  }
}

// The closed return destinations of the setup forms (G063 Plan r1 §5.4): an omitted return_to keeps the compatibility
// page; any other value or a repeated field is refused (undefined). The status codes are a closed set; the raw name is
// never placed in a redirect.
export type ReturnTarget = "account" | "connections";
export type SetupStatus = "created" | "revoked" | "renamed" | "invalid_name";

export function returnTarget(fields: URLSearchParams): ReturnTarget | undefined {
  const values = fields.getAll("return_to");
  if (values.length === 0) return "connections";
  const value = values.length === 1 ? values[0] : undefined;
  return value === "account" || value === "connections" ? value : undefined;
}

export function statusLocation(target: ReturnTarget, status: SetupStatus): string {
  return target === "account" ? `/account?mcp_status=${status}#mcp-connections` : `/connections/mcp?status=${status}`;
}

export function isConnectionId(value: string): boolean {
  return UUID.test(value);
}

export function isListCursor(value: string): boolean {
  return CURSOR.test(value);
}
