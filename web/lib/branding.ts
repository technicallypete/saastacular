import "server-only";

import { runtimeConfig } from "./config";
import { HOST_TIMEOUT_MS } from "./host";
import type { SessionView } from "./identity-client";
import { outboundHeaders, type RequestContext } from "./request-context";
import { AMBIGUOUS, badRequest, field, forbidden, seeOther, statusPage } from "./session";
import type { SemanticColorToken, TokenOverrides, TokenValues } from "./tokens";

// G017 (Plan r2 §1.5, §5.3–5.5): web's reading of Organization branding. Core alone decides what a host is and what its
// public branding is; this module asks it (GET /internal/v1/branding/public?origin=, GET /internal/v1/branding/logo),
// sends exactly the request's diagnostic headers and, on the public reads, no credential, and reduces every answer to a
// closed result. Defence in depth: every colour is re-checked against the grammar and the 15 overridable names, the
// label against its bounds; a malformed answer is "no branding" (never rendered). The management calls (owner/admin,
// core-authorized) carry the browser's own session credential only. Nothing here is cached.

// The overridable semantic colour tokens: lib/tokens.ts SEMANTIC_COLOR_TOKENS (core branding.Tokens).
export const BRANDING_TOKENS = [
  "background",
  "foreground",
  "primary",
  "primary-foreground",
  "accent",
  "muted-foreground",
  "border",
  "popover",
  "popover-foreground",
  "muted",
  "secondary",
  "secondary-foreground",
  "destructive",
  "input",
  "ring",
] as const satisfies readonly SemanticColorToken[];
const OVERRIDABLE: ReadonlySet<string> = new Set(BRANDING_TOKENS);
const COLOUR = /^#[0-9a-f]{6}$/;
const SHA12 = /^[0-9a-f]{12}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const CONTROL = /[\u0000-\u001f\u007f-\u009f]/;
const LOGO_TYPES: ReadonlySet<string> = new Set(["image/png", "image/jpeg"]);
export const MAX_LABEL = 80;
export const MAX_LOGO_BYTES = 524_288;
const MAX_JSON_BYTES = 16 * 1024;

export type Brand = {
  revision: number;
  label: string | null;
  logo: { version: string } | null;
  tokens: TokenOverrides;
};
export type BrandResult = { ok: true; value: Brand } | { ok: false; error: "unknown_host" | "unavailable" | "malformed" };

type Json = Readonly<Record<string, unknown>>;
type Answer = { status: number; body: unknown } | undefined;

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function onlyKeys(o: Json, allowed: readonly string[]): boolean {
  return Object.keys(o).every((k) => allowed.includes(k));
}

// A label as core stores it: 1–80 code points, no control characters, no surrounding whitespace.
export function validLabel(value: unknown): value is string {
  if (typeof value !== "string" || value === "" || value.trim() !== value || CONTROL.test(value)) return false;
  return [...value].length <= MAX_LABEL;
}

// One appearance's overrides: only the 15 names, only lowercase #rrggbb; undefined when anything is outside that.
export function readTokens(value: unknown): TokenValues | undefined {
  if (!isObject(value)) return undefined;
  const out: TokenValues = {};
  for (const [name, colour] of Object.entries(value)) {
    if (!OVERRIDABLE.has(name) || typeof colour !== "string" || !COLOUR.test(colour)) return undefined;
    out[name as SemanticColorToken] = colour;
  }
  return out;
}

// The closed reading of core's public projection answer (exported for the unit tests).
export function readPublicBranding(status: number, body: unknown): BrandResult {
  if (status === 404 && isObject(body) && body.error === "unknown_host" && Object.keys(body).length === 1) {
    return { ok: false, error: "unknown_host" };
  }
  if (status === 503) return { ok: false, error: "unavailable" };
  if (status !== 200 || !isObject(body) || !onlyKeys(body, ["revision", "display_label", "light", "dark", "logo"])) {
    return { ok: false, error: "malformed" };
  }
  const { revision, display_label: label, light, dark, logo } = body;
  const l = readTokens(light);
  const d = readTokens(dark);
  if (typeof revision !== "number" || !Number.isSafeInteger(revision) || revision < 0 || l === undefined || d === undefined) {
    return { ok: false, error: "malformed" };
  }
  if (label !== undefined && !validLabel(label)) return { ok: false, error: "malformed" };
  let mark: Brand["logo"] = null;
  if (logo !== undefined) {
    if (!isObject(logo) || Object.keys(logo).length !== 2 || typeof logo.v !== "string" || !SHA12.test(logo.v) ||
      typeof logo.content_type !== "string" || !LOGO_TYPES.has(logo.content_type)) {
      return { ok: false, error: "malformed" };
    }
    mark = { version: logo.v };
  }
  return { ok: true, value: { revision, label: label ?? null, logo: mark, tokens: { light: l, dark: d } } };
}

async function readBounded(response: Response, limit: number): Promise<Uint8Array | undefined> {
  if (response.body === null) return new Uint8Array();
  const reader = response.body.getReader();
  const chunks: Uint8Array[] = [];
  let size = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.byteLength;
    if (size > limit) {
      await reader.cancel();
      return undefined;
    }
    chunks.push(value);
  }
  return new Uint8Array(Buffer.concat(chunks));
}

type Call = {
  method: "GET" | "POST";
  path: string;
  credential?: string;
  json?: Readonly<Record<string, unknown>>;
  raw?: { bytes: Uint8Array; csrfToken: string; expectedVersion: number };
};

async function call(ctx: RequestContext, c: Call): Promise<Answer> {
  const headers: Record<string, string> = { ...outboundHeaders(ctx) };
  if (c.credential !== undefined) headers.Authorization = `Session ${c.credential}`;
  let body: BodyInit | undefined;
  if (c.json !== undefined) {
    headers["Content-Type"] = "application/json";
    body = JSON.stringify(c.json);
  } else if (c.raw !== undefined) {
    headers["Content-Type"] = "application/octet-stream";
    headers["X-CSRF-Token"] = c.raw.csrfToken;
    headers["X-Expected-Version"] = String(c.raw.expectedVersion);
    body = Buffer.from(c.raw.bytes);
  }
  let response: Response;
  try {
    response = await fetch(new URL(c.path, runtimeConfig().coreInternalUrl), {
      method: c.method,
      headers,
      body,
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(c.raw === undefined ? HOST_TIMEOUT_MS : 30_000),
    });
  } catch {
    return undefined; // no answer in time or no connection: unavailable, never a clean unknown host
  }
  try {
    const bytes = await readBounded(response, MAX_JSON_BYTES);
    if (bytes === undefined || bytes.byteLength === 0) return { status: response.status, body: undefined };
    return { status: response.status, body: JSON.parse(new TextDecoder().decode(bytes)) as unknown };
  } catch {
    return { status: response.status, body: undefined };
  }
}

// GET /internal/v1/branding/public?origin=<origin>: the public projection of the Organization this origin routes to.
// The caller passes a lookup context (lookupContext in lib/host.ts) when the page makes its own core call as well.
export async function resolveBranding(ctx: RequestContext, origin: string): Promise<BrandResult> {
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/branding/public?origin=${encodeURIComponent(origin)}` });
  if (answer === undefined) return { ok: false, error: "unavailable" };
  return readPublicBranding(answer.status, answer.body);
}

export type LogoResult =
  | { ok: true; contentType: "image/png" | "image/jpeg"; bytes: Uint8Array }
  | { ok: false; error: "not_found" | "unavailable" };

// GET /internal/v1/branding/logo?origin=&v=: the current logo bytes of this origin's Organization, only for its current
// version. Anything outside the contract (type, length, ETag, size) is unavailable; core's refusals are not_found.
export async function fetchLogo(ctx: RequestContext, origin: string, version: string): Promise<LogoResult> {
  if (!SHA12.test(version)) return { ok: false, error: "not_found" };
  const path = `/internal/v1/branding/logo?origin=${encodeURIComponent(origin)}&v=${version}`;
  let response: Response;
  try {
    response = await fetch(new URL(path, runtimeConfig().coreInternalUrl), {
      method: "GET",
      headers: { ...outboundHeaders(ctx) },
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(HOST_TIMEOUT_MS),
    });
  } catch {
    return { ok: false, error: "unavailable" };
  }
  if (response.status === 404 || response.status === 400) {
    await response.body?.cancel();
    return { ok: false, error: "not_found" };
  }
  const type = response.headers.get("content-type") ?? "";
  const length = Number(response.headers.get("content-length"));
  if (response.status !== 200 || !LOGO_TYPES.has(type) || response.headers.get("etag") !== `"${version}"` ||
    !Number.isSafeInteger(length) || length < 1 || length > MAX_LOGO_BYTES) {
    await response.body?.cancel();
    return { ok: false, error: "unavailable" };
  }
  try {
    const bytes = await readBounded(response, MAX_LOGO_BYTES);
    if (bytes === undefined || bytes.byteLength !== length) return { ok: false, error: "unavailable" };
    return { ok: true, contentType: type as "image/png" | "image/jpeg", bytes };
  } catch {
    return { ok: false, error: "unavailable" };
  }
}

// --- Authenticated indicator (Plan r2 §1.5, D-6) ----------------------------------------------------------------------

export type Indicator =
  | { kind: "anonymous" }
  | { kind: "member"; text: string }
  | { kind: "none"; text: string }
  | { kind: "unavailable" };

// The header indicator, from core facts only: no session cookie or an unauthenticated answer is anonymous; a session
// answer is member or none for this origin's Organization; any other failure shows no identity at all. The workspace
// name is the public label, or the configured product name.
export function indicatorState(
  result: { ok: true; value: Pick<SessionView, "displayName" | "membership"> } | { ok: false; error: string } | undefined,
  workspace: string,
): Indicator {
  if (result === undefined || (!result.ok && result.error === "unauthenticated")) return { kind: "anonymous" };
  if (!result.ok) return { kind: "unavailable" };
  const who = result.value.displayName ?? "your account";
  return result.value.membership === "member"
    ? { kind: "member", text: `Signed in as ${who} · member of ${workspace}` }
    : { kind: "none", text: `Signed in as ${who} · not a member of this workspace` };
}

// --- Management (owner/admin; Plan r2 §5.3, §5.4) ---------------------------------------------------------------------

export type ManagedLogo = {
  objectId: string;
  contentType: "image/png" | "image/jpeg" | "";
  byteSize: number;
  state: string;
  version: string | null;
};
export type BrandingView = {
  version: number;
  label: string | null;
  light: TokenValues;
  dark: TokenValues;
  logo: ManagedLogo | null;
  uploads: ManagedLogo[];
  uploadsTotal: number;
};

const MANAGEMENT_ERRORS = [
  "invalid_request",
  "ambiguous_parameter",
  "unsupported_media_type",
  "origin_rejected",
  "unauthenticated",
  "csrf_failed",
  "not_available",
  "forbidden",
  "not_found",
  "unknown_host",
  "version_conflict",
  "contrast_failed",
  "logo_invalid",
  "payload_too_large",
  "publication_outcome_unknown",
  "branding_unavailable",
] as const;
export type ManagementError = (typeof MANAGEMENT_ERRORS)[number] | "unavailable" | "malformed";
const KNOWN: ReadonlySet<string> = new Set(MANAGEMENT_ERRORS);
export type ManagementResult<T> = { ok: true; value: T } | { ok: false; error: ManagementError; field?: string };
const LOGO_STATES: ReadonlySet<string> = new Set(["pending", "ready", "deleting", "deleted", "absent"]);

function readLogo(value: unknown): ManagedLogo | undefined {
  if (!isObject(value) || !onlyKeys(value, ["object_id", "content_type", "byte_size", "state", "v"])) return undefined;
  const { object_id: id, content_type: type, byte_size: size, state, v } = value;
  // A purged current logo is "absent": its row is gone, so core knows no type or size for it.
  const typeOk = typeof type === "string" && (LOGO_TYPES.has(type) || (state === "absent" && type === ""));
  if (typeof id !== "string" || !UUID.test(id) || !typeOk ||
    typeof size !== "number" || !Number.isSafeInteger(size) || size < 0 || typeof state !== "string" || !LOGO_STATES.has(state) ||
    (v !== undefined && (typeof v !== "string" || !SHA12.test(v)))) {
    return undefined;
  }
  return { objectId: id, contentType: type as ManagedLogo["contentType"], byteSize: size, state, version: typeof v === "string" ? v : null };
}

// The closed reading of a management answer {"branding": view} (exported for the unit tests).
export function readView(status: number, body: unknown): ManagementResult<BrandingView> {
  if (status !== 200) return refusal(status, body);
  const view = isObject(body) && Object.keys(body).length === 1 ? body.branding : undefined;
  if (!isObject(view)) return { ok: false, error: "malformed" };
  const { version, display_label: label, light, dark, logo, uploads, uploads_total: total } = view;
  const l = readTokens(light);
  const d = readTokens(dark);
  const current = logo === null ? null : readLogo(logo);
  const listed = Array.isArray(uploads) ? uploads.map(readLogo) : uploads === undefined ? [] : undefined;
  if (typeof version !== "number" || !Number.isSafeInteger(version) || version < 0 || l === undefined || d === undefined ||
    !(label === null || validLabel(label)) || current === undefined || listed === undefined || listed.some((u) => u === undefined) ||
    !(total === undefined || (typeof total === "number" && Number.isSafeInteger(total) && total >= 0))) {
    return { ok: false, error: "malformed" };
  }
  return {
    ok: true,
    value: { version, label, light: l, dark: d, logo: current, uploads: listed as ManagedLogo[], uploadsTotal: typeof total === "number" ? total : 0 },
  };
}

// A refusal: core's closed code, plus the closed field name of an invalid input or contrast refusal (a token or
// "<pair>:<appearance>" identifier, never a value) when it has that shape. No answer or a 5xx without a closed code is
// unavailable.
function refusal(status: number, body: unknown): { ok: false; error: ManagementError; field?: string } {
  if (status >= 400 && isObject(body) && typeof body.error === "string" && KNOWN.has(body.error)) {
    const field = typeof body.field === "string" && FIELD.test(body.field) ? body.field : undefined;
    return field === undefined ? { ok: false, error: body.error as ManagementError } : { ok: false, error: body.error as ManagementError, field };
  }
  return { ok: false, error: status >= 500 ? "unavailable" : "malformed" };
}
const FIELD = /^[A-Za-z0-9_:-]{1,64}$/;

async function managed<T>(ctx: RequestContext, c: Call, read: (status: number, body: unknown) => ManagementResult<T>): Promise<ManagementResult<T>> {
  const answer = await call(ctx, c);
  if (answer === undefined) return { ok: false, error: "unavailable" };
  return read(answer.status, answer.body);
}

export function getBranding(ctx: RequestContext, origin: string, credential: string): Promise<ManagementResult<BrandingView>> {
  return managed(ctx, { method: "GET", path: `/internal/v1/branding?origin=${encodeURIComponent(origin)}`, credential }, readView);
}

export type Mutation = { origin: string; csrfToken: string; expectedVersion: number };

export function updateBranding(
  ctx: RequestContext,
  m: Mutation,
  input: { label: string | null; light: TokenValues; dark: TokenValues },
  credential: string,
): Promise<ManagementResult<BrandingView>> {
  return managed(ctx, {
    method: "POST",
    path: "/internal/v1/branding",
    credential,
    json: { origin: m.origin, csrf_token: m.csrfToken, expected_version: m.expectedVersion, display_label: input.label, light: input.light, dark: input.dark },
  }, readView);
}

export function resetBranding(ctx: RequestContext, m: Mutation, credential: string): Promise<ManagementResult<BrandingView>> {
  return managed(ctx, {
    method: "POST",
    path: "/internal/v1/branding/reset",
    credential,
    json: { origin: m.origin, csrf_token: m.csrfToken, expected_version: m.expectedVersion },
  }, readView);
}

export function removeBrandingLogo(ctx: RequestContext, m: Mutation, credential: string): Promise<ManagementResult<BrandingView>> {
  return managed(ctx, {
    method: "POST",
    path: "/internal/v1/branding/logo/remove",
    credential,
    json: { origin: m.origin, csrf_token: m.csrfToken, expected_version: m.expectedVersion },
  }, readView);
}

export function setBrandingLogo(ctx: RequestContext, m: Mutation, bytes: Uint8Array, credential: string): Promise<ManagementResult<BrandingView>> {
  return managed(ctx, {
    method: "POST",
    path: `/internal/v1/branding/logo?origin=${encodeURIComponent(m.origin)}`,
    credential,
    raw: { bytes, csrfToken: m.csrfToken, expectedVersion: m.expectedVersion },
  }, readView);
}

export function deleteBrandingUpload(
  ctx: RequestContext,
  m: Mutation,
  objectId: string,
  credential: string,
): Promise<ManagementResult<"deleted" | "deletion_pending">> {
  return managed(ctx, {
    method: "POST",
    path: "/internal/v1/branding/objects/delete",
    credential,
    json: { origin: m.origin, csrf_token: m.csrfToken, expected_version: m.expectedVersion, object_id: objectId },
  }, (status, body) => {
    if (status !== 200) return refusal(status, body);
    if (isObject(body) && Object.keys(body).length === 1 && (body.outcome === "deleted" || body.outcome === "deletion_pending")) {
      return { ok: true, value: body.outcome };
    }
    return { ok: false, error: "malformed" };
  });
}

// --- Management forms (Plan r2 §5.4: native POST forms following connections/mcp) ------------------------------------

export const BRANDING_PAGE = "/organization/branding";
// The multipart logo form: the 512 KiB image plus at most 16 KiB of form framing; anything larger is never parsed.
export const MAX_UPLOAD_FORM_BYTES = MAX_LOGO_BYTES + 16 * 1024;
const VERSION = /^(0|[1-9][0-9]{0,14})$/;
const FORM_COLOUR = /^#[0-9a-fA-F]{6}$/;
// Core's closed contrast-refusal field "<pair id>:<appearance>" (core/internal/branding/colors.go T1–T10, T12, N2–N4).
export const PAIR_FIELD = /^(T1[02]|T[1-9]|N[2-4]):(light|dark)$/;
// Core's closed invalid-input fields of an update (branding.ValidateLabel / ValidateOverrides).
export const INPUT_FIELDS: readonly string[] = ["display_label", "light", "dark"];

export type MutationFields = { csrfToken: string; expectedVersion: number };

// The CSRF token and the expected version every management form carries. A missing token is sent as empty (core
// answers csrf_failed); a repeated field or a version outside the grammar is undefined (the route answers 400).
export function readMutationFields(get: (name: string) => string | undefined | typeof AMBIGUOUS): MutationFields | undefined {
  const csrf = get("csrf_token");
  const version = get("expected_version");
  if (csrf === AMBIGUOUS || typeof version !== "string" || !VERSION.test(version)) return undefined;
  return { csrfToken: csrf ?? "", expectedVersion: Number(version) };
}

export type BrandingInput = { label: string | null; light: TokenValues; dark: TokenValues };
export type InputResult = { ok: true; value: BrandingInput } | { ok: false; field?: "display_label" | "light" | "dark" };

// The label and colour form: display_label (blank = no label; trimmed and NFC-normalized, then held to core's bounds)
// and light.<name> / dark.<name> for the 15 names (blank = inherit the platform default; #rrggbb in either case, sent
// lowercase). A repeated field is a bad request; a value outside the grammar names its closed field. Core re-validates
// everything and alone decides contrast.
export function readBrandingInput(fields: URLSearchParams): InputResult {
  const rawLabel = field(fields, "display_label");
  if (rawLabel === AMBIGUOUS) return { ok: false };
  const trimmed = (rawLabel ?? "").trim().normalize("NFC");
  if (trimmed !== "" && !validLabel(trimmed)) return { ok: false, field: "display_label" };
  const out: BrandingInput = { label: trimmed === "" ? null : trimmed, light: {}, dark: {} };
  for (const mode of ["light", "dark"] as const) {
    for (const name of BRANDING_TOKENS) {
      const value = field(fields, `${mode}.${name}`);
      if (value === AMBIGUOUS) return { ok: false };
      const v = (value ?? "").trim();
      if (v === "") continue;
      if (!FORM_COLOUR.test(v)) return { ok: false, field: mode };
      out[mode][name] = v.toLowerCase();
    }
  }
  return { ok: true, value: out };
}

// Exactly one multipart/form-data Content-Type with a boundary; repeated headers arrive comma-joined and are refused.
export function isMultipartForm(value: string | null): value is string {
  if (value === null || value.includes(",")) return false;
  const [type, ...params] = value.split(";").map((part) => part.trim());
  return type.toLowerCase() === "multipart/form-data" && params.length === 1 && /^boundary=("?)[0-9A-Za-z'()+_.\/:=?-]{1,70}\1$/i.test(params[0]);
}

// The request body, or undefined once it exceeds limit bytes (a larger declared Content-Length is never read).
export async function readBoundedBytes(request: Request, limit: number): Promise<Uint8Array | undefined> {
  const declared = request.headers.get("content-length");
  if (declared !== null && (!/^[0-9]{1,12}$/.test(declared) || Number(declared) > limit)) return undefined;
  if (request.body === null) return new Uint8Array(0);
  const reader = request.body.getReader();
  const chunks: Uint8Array[] = [];
  let size = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.byteLength;
    if (size > limit) {
      await reader.cancel();
      return undefined;
    }
    chunks.push(value);
  }
  return new Uint8Array(Buffer.concat(chunks));
}

export type LogoUpload = { ok: true; fields: MutationFields; bytes: Uint8Array } | { ok: false; error: "bad_request" | "too_large" };

// The logo form's multipart body: exactly the fields csrf_token, expected_version and one non-empty logo file of at most
// 512 KiB, each at most once. The file's declared name and type are ignored: core validates the bytes and decides.
export async function parseLogoUpload(contentType: string, body: Uint8Array): Promise<LogoUpload> {
  let form: FormData;
  try {
    form = await new Response(Buffer.from(body), { headers: { "Content-Type": contentType } }).formData();
  } catch {
    return { ok: false, error: "bad_request" };
  }
  const names = [...form.keys()];
  if (names.some((n) => n !== "csrf_token" && n !== "expected_version" && n !== "logo")) return { ok: false, error: "bad_request" };
  const text = (name: string): string | undefined | typeof AMBIGUOUS => {
    const values = form.getAll(name);
    if (values.length > 1) return AMBIGUOUS;
    return typeof values[0] === "string" ? values[0] : values[0] === undefined ? undefined : AMBIGUOUS;
  };
  const fields = readMutationFields(text);
  const files = form.getAll("logo");
  const file = files.length === 1 && typeof files[0] !== "string" ? files[0] : undefined;
  if (fields === undefined || file === undefined || file.size === 0) return { ok: false, error: "bad_request" };
  if (file.size > MAX_LOGO_BYTES) return { ok: false, error: "too_large" };
  return { ok: true, fields, bytes: new Uint8Array(await file.arrayBuffer()) };
}

export type FormStatus = "updated" | "reset" | "logo_set" | "logo_removed" | "upload_deleted" | "deletion_pending";

// A management form's answer (POST-redirect-GET): success and the recoverable refusals return to the page with one
// closed status (and, for an input or contrast refusal, its closed field); the session, CSRF and origin refusals and
// an unconfirmed outcome are the form layer's fixed pages; a viewer who may not manage branding gets the fixed 404.
export function brandingAnswer(result: ManagementResult<unknown>, success: FormStatus, what: string): Response {
  if (result.ok) return seeOther(`${BRANDING_PAGE}?status=${success}`);
  const back = (status: string, f?: string) =>
    seeOther(`${BRANDING_PAGE}?status=${status}${f === undefined ? "" : `&field=${encodeURIComponent(f)}`}`);
  switch (result.error) {
    case "version_conflict":
    case "logo_invalid":
    case "payload_too_large":
    case "publication_outcome_unknown":
      return back(result.error);
    case "contrast_failed":
      return back(result.error, result.field !== undefined && PAIR_FIELD.test(result.field) ? result.field : undefined);
    case "invalid_request":
      return result.field !== undefined && INPUT_FIELDS.includes(result.field) ? back("invalid_input", result.field) : badRequest();
    case "not_found":
      return success === "upload_deleted" ? back("upload_not_found") : notHere();
    case "ambiguous_parameter":
    case "unsupported_media_type":
      return badRequest();
    case "csrf_failed":
      return statusPage(403, "Form expired", "This form is no longer valid. Reload the branding page and try again. Nothing was changed.");
    case "unauthenticated":
      return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
    case "origin_rejected":
      return forbidden();
    case "forbidden":
    case "not_available":
    case "unknown_host":
      return notHere();
    default:
      return statusPage(503, "Change not confirmed", `${what} could not be confirmed. Reload the branding page to see the current state.`);
  }
}

function notHere(): Response {
  return statusPage(404, "Not found", "There is nothing at this address.");
}
