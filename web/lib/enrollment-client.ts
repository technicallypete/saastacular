import "server-only";

import { runtimeConfig } from "./config";
import { onboardingPage, type StatusLink } from "./onboarding-client";
import { deriveContext, errorCategory, outboundHeaders, type RequestContext } from "./request-context";
import {
  AMBIGUOUS,
  badRequest,
  clearCookie,
  field,
  forbidden,
  readCookie,
  readForm,
  seeOther,
  setJoinCookie,
  setCookie,
  type CookieProfile,
} from "./session";

// G058 (Plan r3 §1.3, §1.6, §5.2–§5.3; R13 route reconciliation): the server-only client of core's internal enrollment
// contract (/internal/v1/enrollment/*: ten method/path pairs, one function each) plus the enrollment variants of the two
// modified auth routes (the enrollment-intent login and the callback answers; lib/identity-client.ts is unchanged), and
// the form handlers of the /join and /organization/enrollment routes. Each call sends exactly the request's diagnostic
// headers, at most one Authorization header — "Enrollment <credential>" for the applicant's own routes, "Session
// <credential>" for the manager routes and the auth routes, none for the public route — X-Login-Binding on the auth
// routes only and, with a body, Content-Type: application/json with the documented fields only. Core alone resolves
// the Organization from its host registry and decides every read and write; an enrollment credential is never a
// session, and a historical "admitted" never derives authority (current membership comes only from core's guards on
// the ordinary sign-in that follows). Answers are validated against strict shapes with per-endpoint byte bounds;
// anything outside the contract is "malformed". The invitation token travels only in the one-use join cookie and core's
// JSON bodies: never in a URL, a Location header, a page title or a log.

export const ENROLLMENT_TIMEOUT_MS = 10_000;
const MAX_RESPONSE_BYTES = 64 * 1024; // every refusal and any answer without its own bound

export const MAX_REQUESTS_PAGE = 100; // tenancy MaxEnrollmentRequestsPage
export const REQUESTS_PAGE_SIZE = 50;
const MAX_DISPLAY_NAME = 200; // app.users display_name CHECK (code points)
const MAX_LABEL = 80; // the G017 public label (code points)

// Success bounds. An applicant's request row (ENR-17 minimal disclosure): a UUID (38), a status (10), two booleans,
// four RFC 3339 instants (88) and about 150 bytes of keys and syntax ≈ 300 ≤ 600. A manager's row adds the user id (50)
// and a display name of at most 200 code points escaped (1,202) = ≤ 1,900. A list: 100 rows, 99 commas and a 128-byte
// envelope.
export const REQUEST_ROW_MAX_BYTES = 600;
export const MANAGED_ROW_MAX_BYTES = 1_900;
const ENVELOPE_BYTES = 128;
export const REQUEST_LIST_MAX_BYTES = ENVELOPE_BYTES + (MAX_REQUESTS_PAGE - 1) + MAX_REQUESTS_PAGE * MANAGED_ROW_MAX_BYTES;
const STATUS_MAX_BYTES = ENVELOPE_BYTES + REQUEST_ROW_MAX_BYTES + 64;
const ROW_MAX_BYTES = ENVELOPE_BYTES + MANAGED_ROW_MAX_BYTES;
const SMALL_MAX_BYTES = 1_024;
const AUTH_MAX_BYTES = 16 * 1024;

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const INSTANT = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/;
const CODE = /^[a-z_]{1,64}$/;

// core/internal/enrollment/api's closed refusal codes.
const API_ERRORS = [
  "invalid_request",
  "ambiguous_parameter",
  "origin_rejected",
  "unauthenticated",
  "csrf_failed",
  "not_available",
  "forbidden",
  "not_found",
  "method_not_allowed",
  "conflict",
  "enrollment_unavailable",
  "rate_limited",
  "unknown_host",
  "enrollment_closed",
] as const;
// The auth routes' closed codes (identity's set) plus the two G058 callback refusals.
const AUTH_ERRORS = [
  "invalid_request",
  "ambiguous_parameter",
  "unsupported_media_type",
  "origin_rejected",
  "unauthenticated",
  "csrf_failed",
  "already_authenticated",
  "transaction_invalid",
  "transaction_expired",
  "login_canceled",
  "code_rejected",
  "identity_conflict",
  "user_disabled",
  "impersonation_refused",
  "stale_authentication",
  "provider_error",
  "provider_timeout",
  "identity_namespace_mismatch",
  "identity_unconfigured",
  "identity_unavailable",
  "host_changed",
  "organization_not_member",
  "unknown_host",
  "enrollment_closed",
  "invitation_mismatch",
] as const;
const API_KNOWN: ReadonlySet<string> = new Set(API_ERRORS);
const AUTH_KNOWN: ReadonlySet<string> = new Set(AUTH_ERRORS);
// Codes that may carry a closed field (core's refuseErr: invalid_request, forbidden and conflict).
const FIELD_ERRORS: ReadonlySet<string> = new Set(["invalid_request", "forbidden", "conflict"]);

export type EnrollmentError = (typeof API_ERRORS)[number] | (typeof AUTH_ERRORS)[number] | "timeout" | "unreachable" | "malformed";
export type EnrollmentResult<T> = { ok: true; value: T } | { ok: false; error: EnrollmentError; field?: string };

export const MODES = ["invite_only", "request_access", "open"] as const;
export type EnrollmentMode = (typeof MODES)[number];
export type PublicEnrollment = { mode: EnrollmentMode; label: string | null };
export type RequestStatus = "pending" | "approved" | "denied" | "revoked";
export type OwnRequest = {
  id: string;
  status: RequestStatus;
  expired: boolean;
  consumed: boolean;
  createdAt: string;
  expiresAt: string;
  decidedAt: string | null;
  revokedAt: string | null;
};
export type ManagedRequest = OwnRequest & { userId: string; displayName: string | null };
export type RequestAction = "approve" | "deny" | "revoke";

// The administrators' row view (Codex preliminary correction 2). A valid pending request offers Approve and Deny. An
// expired pending request stays pending until an administrator decides (one_pending), so it keeps the existing Deny,
// with copy saying so; approval is not offered and never renews validity. An unconsumed, unexpired approval offers
// Revoke; an expired approval stays non-consumable and offers nothing. The actions are hints: core re-decides.
export function requestRowView(r: ManagedRequest): { state: string; actions: RequestAction[] } {
  if (r.consumed) return { state: "Joined", actions: [] };
  if (r.status === "pending") return r.expired ? { state: "Expired, deny to resolve", actions: ["deny"] } : { state: "Pending", actions: ["approve", "deny"] };
  if (r.status === "approved") return r.expired ? { state: "Expired", actions: [] } : { state: "Approved", actions: ["revoke"] };
  return { state: r.status === "denied" ? "Denied" : "Approval withdrawn", actions: [] };
}
// admitted is the credential's historical admission proof, never membership; member is core's current re-read.
export type ApplicantStatus = "admitted" | "none" | RequestStatus;
export type StatusView = { status: ApplicantStatus; request: OwnRequest | null; member: boolean; csrfToken: string };
export type Admitted = { outcome: "admitted" | "already_member" };
export type Policy = { mode: EnrollmentMode; openReady: boolean; historyBasis: string | null };
export type RequestPage = { requests: ManagedRequest[]; next: string | null };
export type Mutation = { origin: string; csrfToken: string };
export type IssuedSecret = { value: string; maxAge: number };
export type Decision = "approve" | "deny";
// The 202 statuses of an enrollment callback: pending, or a known refusal with nothing written.
export const CALLBACK_STATUSES = [
  "pending",
  "enrollment_closed",
  "enrollment_restricted",
  "recipient_ineligible",
  "already_consumed",
  "not_found",
  "enrollment_expired",
  "host_changed",
  "denied",
  "approval_revoked",
] as const;
export type CallbackStatus = (typeof CALLBACK_STATUSES)[number];
export type CallbackReturn = "/" | "/account" | "/join/status";
export type CallbackOutcome =
  | { kind: "session"; outcome: "signed_in" | "session_rotated"; returnTo: CallbackReturn; session: IssuedSecret }
  | { kind: "already_signed_in"; returnTo: CallbackReturn }
  | { kind: "admitted"; outcome: "admitted" | "already_member" }
  | { kind: "enrollment"; status: CallbackStatus; credential: IssuedSecret };

type Json = Readonly<Record<string, unknown>>;
type Auth = { scheme: "Session" | "Enrollment"; credential: string };
type Call = { method: "GET" | "POST"; path: string; auth?: Auth; binding?: string; body?: Json; successBytes: number };
type Answer = { kind: "answered"; status: number; body: unknown } | { kind: "failed"; error: "timeout" | "unreachable" | "malformed" };

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasExactly(o: Json, keys: readonly string[]): boolean {
  const own = Object.keys(o);
  return own.length === keys.length && keys.every((k) => Object.hasOwn(o, k));
}

export function isId(value: unknown): value is string {
  return typeof value === "string" && UUID.test(value);
}

export function isMode(value: unknown): value is EnrollmentMode {
  return typeof value === "string" && (MODES as readonly string[]).includes(value);
}

function instant(value: unknown): value is string {
  return typeof value === "string" && INSTANT.test(value) && !Number.isNaN(Date.parse(value));
}

function optionalInstant(value: unknown): value is string | null {
  return value === null || instant(value);
}

function textOf(value: unknown, max: number): value is string {
  return typeof value === "string" && value.length > 0 && Array.from(value).length <= max;
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
  if (c.auth !== undefined) headers.Authorization = `${c.auth.scheme} ${c.auth.credential}`;
  if (c.binding !== undefined) headers["X-Login-Binding"] = c.binding;
  if (c.body !== undefined) headers["Content-Type"] = "application/json";
  let response: Response;
  try {
    response = await fetch(new URL(c.path, runtimeConfig().coreInternalUrl), {
      method: c.method,
      headers,
      body: c.body === undefined ? undefined : JSON.stringify(c.body),
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(ENROLLMENT_TIMEOUT_MS),
    });
  } catch (error) {
    return { kind: "failed", error: errorCategory(error) === "TimeoutError" ? "timeout" : "unreachable" };
  }
  try {
    const ok = response.status >= 200 && response.status < 300;
    const body = await readBounded(response, ok ? c.successBytes : MAX_RESPONSE_BYTES);
    if (body === undefined) return { kind: "failed", error: "malformed" };
    return { kind: "answered", status: response.status, body: JSON.parse(body) as unknown };
  } catch (error) {
    if (errorCategory(error) === "TimeoutError") return { kind: "failed", error: "timeout" };
    return { kind: "failed", error: "malformed" };
  }
}

// A refusal must be {"error": <closed code>} or, for invalid_request, forbidden and conflict, {"error", "field"}.
function failure(answer: Answer, known: ReadonlySet<string> = API_KNOWN): { ok: false; error: EnrollmentError; field?: string } {
  if (answer.kind === "failed") return { ok: false, error: answer.error };
  const body = answer.body;
  if (answer.status >= 400 && isObject(body) && typeof body.error === "string" && known.has(body.error)) {
    const error = body.error as EnrollmentError;
    if (hasExactly(body, ["error"])) return { ok: false, error };
    if (known === API_KNOWN && FIELD_ERRORS.has(error) && hasExactly(body, ["error", "field"]) && typeof body.field === "string" && CODE.test(body.field)) {
      return { ok: false, error, field: body.field };
    }
  }
  return { ok: false, error: "malformed" };
}

function ok200(answer: Answer): Json | undefined {
  return answer.kind === "answered" && answer.status === 200 && isObject(answer.body) ? answer.body : undefined;
}

const OWN_KEYS = ["id", "status", "expired", "consumed", "created_at", "expires_at", "decided_at", "revoked_at"];
const REQUEST_STATUSES: ReadonlySet<string> = new Set(["pending", "approved", "denied", "revoked"]);

function ownFields(v: Json): OwnRequest | undefined {
  const { id, status, expired, consumed, created_at: createdAt, expires_at: expiresAt, decided_at: decidedAt, revoked_at: revokedAt } = v;
  if (!isId(id) || typeof status !== "string" || !REQUEST_STATUSES.has(status) || typeof expired !== "boolean" || typeof consumed !== "boolean") return undefined;
  if (!instant(createdAt) || !instant(expiresAt) || !optionalInstant(decidedAt) || !optionalInstant(revokedAt)) return undefined;
  // A pending request is undecided and unconsumed; only a revoked approval carries its revocation time.
  if (status === "pending" && (decidedAt !== null || consumed)) return undefined;
  if ((status === "revoked") !== (revokedAt !== null)) return undefined;
  return { id, status: status as RequestStatus, expired, consumed, createdAt, expiresAt, decidedAt, revokedAt };
}

function ownRequest(value: unknown): OwnRequest | undefined {
  return isObject(value) && hasExactly(value, OWN_KEYS) ? ownFields(value) : undefined;
}

function managedRequest(value: unknown): ManagedRequest | undefined {
  if (!isObject(value) || !hasExactly(value, [...OWN_KEYS, "user_id", "display_name"])) return undefined;
  const own = ownFields(value);
  const { user_id: userId, display_name: displayName } = value;
  if (own === undefined || !isId(userId) || !(displayName === null || textOf(displayName, MAX_DISPLAY_NAME))) return undefined;
  return { ...own, userId, displayName };
}

function wrappedRequest(answer: Answer): EnrollmentResult<ManagedRequest> {
  const b = ok200(answer);
  if (b === undefined) return failure(answer);
  const v = hasExactly(b, ["request"]) ? managedRequest(b.request) : undefined;
  return v === undefined ? { ok: false, error: "malformed" } : { ok: true, value: v };
}

function policyOf(b: Json): Policy | undefined {
  if (!hasExactly(b, ["mode", "open_ready", "history_basis"])) return undefined;
  const { mode, open_ready: openReady, history_basis: basis } = b;
  if (!isMode(mode) || typeof openReady !== "boolean" || !(basis === null || (typeof basis === "string" && CODE.test(basis)))) return undefined;
  // The CHECK makes open imply ready.
  if (mode === "open" && !openReady) return undefined;
  return { mode, openReady, historyBasis: basis };
}

function query(origin: string): string {
  return `origin=${encodeURIComponent(origin)}`;
}

// --- the ten enrollment routes -----------------------------------------------------------------------------------------

// GET /internal/v1/enrollment/public?origin= (no credential): the mode and public label of the Organization the origin
// routes to; a host that routes nowhere is unknown_host.
export async function publicEnrollment(ctx: RequestContext, origin: string): Promise<EnrollmentResult<PublicEnrollment>> {
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/enrollment/public?${query(origin)}`, successBytes: SMALL_MAX_BYTES });
  const b = ok200(answer);
  if (b === undefined) return failure(answer);
  const { mode, label } = b;
  if (!hasExactly(b, ["mode", "label"]) || !isMode(mode) || !(label === null || textOf(label, MAX_LABEL))) return { ok: false, error: "malformed" };
  return { ok: true, value: { mode, label } };
}

const APPLICANT_STATUSES: ReadonlySet<string> = new Set(["admitted", "none", ...REQUEST_STATUSES]);

// GET /internal/v1/enrollment/status?origin= (own enrollment credential): the reconciliation read.
export async function enrollmentStatus(ctx: RequestContext, credential: string, origin: string): Promise<EnrollmentResult<StatusView>> {
  const answer = await call(ctx, {
    method: "GET",
    path: `/internal/v1/enrollment/status?${query(origin)}`,
    auth: { scheme: "Enrollment", credential },
    successBytes: STATUS_MAX_BYTES,
  });
  const b = ok200(answer);
  if (b === undefined) return failure(answer);
  if (!hasExactly(b, ["status", "request", "member", "csrf_token"])) return { ok: false, error: "malformed" };
  const { status, member, csrf_token: csrf } = b;
  const request = b.request === null ? null : ownRequest(b.request);
  if (typeof status !== "string" || !APPLICANT_STATUSES.has(status) || request === undefined || typeof member !== "boolean" || typeof csrf !== "string" || !SECRET.test(csrf)) {
    return { ok: false, error: "malformed" };
  }
  // none means no request; a request state names the latest request's own state.
  if ((status === "none" && request !== null) || (REQUEST_STATUSES.has(status) && request?.status !== status)) return { ok: false, error: "malformed" };
  return { ok: true, value: { status: status as ApplicantStatus, request, member, csrfToken: csrf } };
}

// POST /internal/v1/enrollment/request: the applicant's own access request; 200 for a new and an existing pending request.
export async function requestEnrollment(ctx: RequestContext, credential: string, m: Mutation): Promise<EnrollmentResult<OwnRequest>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/enrollment/request",
    auth: { scheme: "Enrollment", credential },
    body: { origin: m.origin, csrf_token: m.csrfToken },
    successBytes: STATUS_MAX_BYTES,
  });
  const b = ok200(answer);
  if (b === undefined) return failure(answer);
  const v = hasExactly(b, ["request"]) ? ownRequest(b.request) : undefined;
  return v === undefined ? { ok: false, error: "malformed" } : { ok: true, value: v };
}

// POST /internal/v1/enrollment/admit: the applicant's own admission (open, or an approved request). 200 never carries a
// session: the caller clears the enrollment cookie and the browser starts the ordinary sign-in.
export async function admit(ctx: RequestContext, credential: string, m: Mutation, kind: "open" | "request", requestId?: string): Promise<EnrollmentResult<Admitted>> {
  const body: Record<string, string> = { origin: m.origin, csrf_token: m.csrfToken, kind };
  if (kind === "request") {
    if (!isId(requestId)) return { ok: false, error: "invalid_request", field: "request_id" };
    body.request_id = requestId;
  }
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/enrollment/admit", auth: { scheme: "Enrollment", credential }, body, successBytes: SMALL_MAX_BYTES });
  const b = ok200(answer);
  if (b === undefined) return failure(answer);
  if (!hasExactly(b, ["outcome", "next"]) || (b.outcome !== "admitted" && b.outcome !== "already_member") || b.next !== "sign_in") return { ok: false, error: "malformed" };
  return { ok: true, value: { outcome: b.outcome } };
}

// POST /internal/v1/enrollment/abandon: revokes the applicant's own credential (idempotent).
export async function abandon(ctx: RequestContext, credential: string, m: Mutation): Promise<EnrollmentResult<null>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/enrollment/abandon",
    auth: { scheme: "Enrollment", credential },
    body: { origin: m.origin, csrf_token: m.csrfToken },
    successBytes: SMALL_MAX_BYTES,
  });
  const b = ok200(answer);
  if (b === undefined) return failure(answer);
  return hasExactly(b, []) ? { ok: true, value: null } : { ok: false, error: "malformed" };
}

// GET /internal/v1/enrollment/policy?origin= (session; organization.read).
export async function getPolicy(ctx: RequestContext, credential: string, origin: string): Promise<EnrollmentResult<Policy>> {
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/enrollment/policy?${query(origin)}`, auth: { scheme: "Session", credential }, successBytes: SMALL_MAX_BYTES });
  const b = ok200(answer);
  if (b === undefined) return failure(answer);
  const p = policyOf(b);
  return p === undefined ? { ok: false, error: "malformed" } : { ok: true, value: p };
}

// POST /internal/v1/enrollment/policy (session; organization.manage_policy): 409 conflict enrollment_not_ready for open on
// an unresolved membership history.
export async function updatePolicy(ctx: RequestContext, credential: string, m: Mutation, mode: EnrollmentMode): Promise<EnrollmentResult<Policy>> {
  if (!isMode(mode)) return { ok: false, error: "invalid_request", field: "invalid_enrollment_mode" };
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/enrollment/policy",
    auth: { scheme: "Session", credential },
    body: { origin: m.origin, csrf_token: m.csrfToken, mode },
    successBytes: SMALL_MAX_BYTES,
  });
  const b = ok200(answer);
  if (b === undefined) return failure(answer);
  const p = policyOf(b);
  return p === undefined ? { ok: false, error: "malformed" } : { ok: true, value: p };
}

export type RequestsQuery = { after?: string; limit?: number; status?: RequestStatus };

// GET /internal/v1/enrollment/requests (session; organization.manage_members): one keyset page by id.
export async function listRequests(ctx: RequestContext, credential: string, origin: string, page: RequestsQuery = {}): Promise<EnrollmentResult<RequestPage>> {
  const limit = page.limit ?? REQUESTS_PAGE_SIZE;
  if ((page.after !== undefined && !isId(page.after)) || !Number.isInteger(limit) || limit < 1 || limit > MAX_REQUESTS_PAGE) return { ok: false, error: "invalid_request" };
  if (page.status !== undefined && !REQUEST_STATUSES.has(page.status)) return { ok: false, error: "invalid_request" };
  let q = `${query(origin)}&limit=${limit}`;
  if (page.after !== undefined) q += `&after=${page.after}`;
  if (page.status !== undefined) q += `&status=${page.status}`;
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/enrollment/requests?${q}`, auth: { scheme: "Session", credential }, successBytes: REQUEST_LIST_MAX_BYTES });
  const b = ok200(answer);
  if (b === undefined) return failure(answer);
  if (!hasExactly(b, ["requests", "next"]) || !Array.isArray(b.requests) || b.requests.length > limit) return { ok: false, error: "malformed" };
  const rows: ManagedRequest[] = [];
  for (const item of b.requests as unknown[]) {
    const r = managedRequest(item);
    if (r === undefined || (rows.length > 0 && r.id <= rows[rows.length - 1].id)) return { ok: false, error: "malformed" };
    rows.push(r);
  }
  const next = b.next;
  if (!(next === null || (isId(next) && (rows.length === 0 || next >= rows[rows.length - 1].id)))) return { ok: false, error: "malformed" };
  return { ok: true, value: { requests: rows, next } };
}

// POST /internal/v1/enrollment/requests/{id}/decide (session; organization.manage_members).
export async function decideRequest(ctx: RequestContext, credential: string, m: Mutation, id: string, decision: Decision): Promise<EnrollmentResult<ManagedRequest>> {
  if (!isId(id)) return { ok: false, error: "not_found" };
  if (decision !== "approve" && decision !== "deny") return { ok: false, error: "invalid_request" };
  const answer = await call(ctx, {
    method: "POST",
    path: `/internal/v1/enrollment/requests/${id}/decide`,
    auth: { scheme: "Session", credential },
    body: { origin: m.origin, csrf_token: m.csrfToken, decision },
    successBytes: ROW_MAX_BYTES,
  });
  return wrappedRequest(answer);
}

// POST /internal/v1/enrollment/requests/{id}/revoke (session; organization.manage_members): withdraws an unconsumed approval.
export async function revokeRequest(ctx: RequestContext, credential: string, m: Mutation, id: string): Promise<EnrollmentResult<ManagedRequest>> {
  if (!isId(id)) return { ok: false, error: "not_found" };
  const answer = await call(ctx, {
    method: "POST",
    path: `/internal/v1/enrollment/requests/${id}/revoke`,
    auth: { scheme: "Session", credential },
    body: { origin: m.origin, csrf_token: m.csrfToken },
    successBytes: ROW_MAX_BYTES,
  });
  return wrappedRequest(answer);
}

// --- the enrollment variants of the two modified auth routes -----------------------------------------------------------

function issuedSecret(value: unknown, max: number): IssuedSecret | undefined {
  if (!isObject(value) || !hasExactly(value, ["value", "max_age"])) return undefined;
  const { value: secret, max_age: maxAge } = value;
  if (typeof secret !== "string" || !SECRET.test(secret)) return undefined;
  if (typeof maxAge !== "number" || !Number.isInteger(maxAge) || maxAge < 1 || maxAge > max) return undefined;
  return { value: secret, maxAge };
}

function absoluteHttpUrl(value: unknown): string | undefined {
  if (typeof value !== "string" || value.length > 8192) return undefined;
  try {
    const url = new URL(value);
    return url.protocol === "https:" || url.protocol === "http:" ? url.href : undefined;
  } catch {
    return undefined;
  }
}

export type EnrollmentLogin = { origin: string; invitationId?: string; request?: boolean };

// POST /internal/v1/auth/login with intent "enroll" (optionally one invitation id and a request on arrival). Core fixes
// the return path (/join/status); the answer is the ordinary login answer.
export async function startEnrollment(
  ctx: RequestContext,
  input: EnrollmentLogin,
  credential?: string,
  binding?: string,
): Promise<EnrollmentResult<{ authorizationUrl: string; binding: IssuedSecret }>> {
  const body: Record<string, unknown> = { origin: input.origin, intent: "enroll" };
  if (input.invitationId !== undefined) {
    if (!isId(input.invitationId)) return { ok: false, error: "invalid_request" };
    body.invitation_id = input.invitationId;
  }
  if (input.request === true) body.request = true;
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/auth/login",
    auth: credential === undefined ? undefined : { scheme: "Session", credential },
    binding,
    body,
    successBytes: AUTH_MAX_BYTES,
  });
  const b = ok200(answer);
  if (b === undefined) return failure(answer, AUTH_KNOWN);
  const authorizationUrl = absoluteHttpUrl(b.authorization_url);
  const issued = issuedSecret(b.binding, 86_400);
  if (!hasExactly(b, ["authorization_url", "binding"]) || authorizationUrl === undefined || issued === undefined) return { ok: false, error: "malformed" };
  return { ok: true, value: { authorizationUrl, binding: issued } };
}

function callbackReturn(value: unknown): CallbackReturn | undefined {
  return value === "/" || value === "/account" || value === "/join/status" ? value : undefined;
}

// POST /internal/v1/auth/callback for every completion (only core knows whether the stored login carried the enrollment
// intent): the ordinary 200 session answers (as lib/identity-client.ts validates them, plus the enrollment login's fixed
// /join/status return path), the enrollment 200 {outcome, next: "sign_in"} (a confirmed admission: no credential, no
// session) and 202 {enrollment: {status, credential: {value, max_age}}} (the enrollment credential only, ≤ 1800 s).
export async function completeCallback(
  ctx: RequestContext,
  input: { origin: string; rawQuery: string; invitation?: { invitationId: string; token: string } },
  binding?: string,
  credential?: string,
): Promise<EnrollmentResult<CallbackOutcome>> {
  const body: Record<string, string> = { origin: input.origin, raw_query: input.rawQuery };
  if (input.invitation !== undefined) {
    body.invitation_id = input.invitation.invitationId;
    body.invitation_token = input.invitation.token;
  }
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/auth/callback",
    auth: credential === undefined ? undefined : { scheme: "Session", credential },
    binding,
    body,
    successBytes: AUTH_MAX_BYTES,
  });
  if (answer.kind === "answered" && answer.status === 202 && isObject(answer.body)) {
    const e = hasExactly(answer.body, ["enrollment"]) ? answer.body.enrollment : undefined;
    if (isObject(e) && hasExactly(e, ["status", "credential"]) && (CALLBACK_STATUSES as readonly unknown[]).includes(e.status)) {
      const issued = issuedSecret(e.credential, 1_800);
      if (issued !== undefined) return { ok: true, value: { kind: "enrollment", status: e.status as CallbackStatus, credential: issued } };
    }
    return { ok: false, error: "malformed" };
  }
  const b = ok200(answer);
  if (b === undefined) return failure(answer, AUTH_KNOWN);
  if (Object.hasOwn(b, "next")) {
    if (hasExactly(b, ["outcome", "next"]) && b.next === "sign_in" && (b.outcome === "admitted" || b.outcome === "already_member")) {
      return { ok: true, value: { kind: "admitted", outcome: b.outcome } };
    }
    return { ok: false, error: "malformed" };
  }
  const returnTo = callbackReturn(b.return_to);
  if (returnTo === undefined) return { ok: false, error: "malformed" };
  if (b.outcome === "already_signed_in" && hasExactly(b, ["outcome", "return_to"])) return { ok: true, value: { kind: "already_signed_in", returnTo } };
  const issued = issuedSecret(b.session, 86_400);
  if ((b.outcome === "signed_in" || b.outcome === "session_rotated") && issued !== undefined && hasExactly(b, ["outcome", "return_to", "session"])) {
    return { ok: true, value: { kind: "session", outcome: b.outcome, returnTo, session: issued } };
  }
  return { ok: false, error: "malformed" };
}

// --- closed notices ------------------------------------------------------------------------------------------------------

// G060 owner amendment (Option B): the one fixed sentence for an invitation link that does not work for the signed-in
// account, shared by the Organization callback's invitation_mismatch notice and the Group invitation page (malformed,
// nonexistent, foreign and wrong-recipient ids alike). Its advice is generic: it never names the intended account or
// email and never says whether the invitation exists.
export const INVALID_INVITATION = "Invalid invitation link. Switch to a different account, or ask for a new invitation link.";

// The closed ?status= codes of the /join pages and the sentence each shows; never a submitted value. They describe the
// last step only: the status page always shows core's current state below them.
export const JOIN_NOTICES: Readonly<Record<string, string>> = {
  requested: "Your request was sent. An administrator of this workspace will review it.",
  pending: "You are signed in for joining. Nothing has been granted yet.",
  admitted: "Your membership was confirmed. Sign in to open the workspace.",
  already_member: "Your account is already a member of this workspace. Sign in to open it.",
  denied: "Your request was not approved.",
  approval_revoked: "The approval of your request was withdrawn. Nothing was granted.",
  enrollment_restricted: "Joining this workspace this way is not available for your account. Ask an administrator.",
  closed: "This workspace is not accepting new members this way right now. Nothing was changed.",
  recipient_ineligible: "Your account cannot join this workspace this way.",
  already_consumed: "That invitation was already used.",
  not_found: "That invitation or request is not available. Ask for a new invitation.",
  expired: "Your joining session ended or expired. Start again.",
  host_changed: "This address no longer belongs to the workspace you started joining. Nothing was changed.",
  not_pending: "That request is no longer pending.",
  invitation_mismatch: INVALID_INVITATION,
  abandoned: "You stopped joining. A request you sent stays with the administrators.",
  unknown: "Your last step may or may not have completed. The current state is shown below; check it before trying again.",
  rate_limited: "Too many attempts. Wait a moment and try again.",
  unavailable: "Joining is unavailable right now. Please try again later.",
};

// A callback 202 status or an applicant conflict field as a notice code.
function noticeOf(code: string | undefined): string {
  if (code === "enrollment_closed") return "closed";
  if (code === "enrollment_expired") return "expired";
  return code !== undefined && Object.hasOwn(JOIN_NOTICES, code) ? code : "unknown";
}

export function joinNotice(value: string | undefined): string | undefined {
  return value !== undefined && Object.hasOwn(JOIN_NOTICES, value) ? JOIN_NOTICES[value] : undefined;
}

// An outcome core may or may not have committed.
export function outcomeUnknown(error: EnrollmentError): boolean {
  return error === "timeout" || error === "unreachable" || error === "malformed" || error === "enrollment_unavailable";
}

// The callback's location and cookies for an enrollment answer (the cookie contract): 202 sets the enrollment credential
// (never the session cookie) and shows its status; a confirmed admission clears the enrollment cookie and asks for the
// ordinary sign-in.
export function callbackEnrollment(outcome: Extract<CallbackOutcome, { kind: "admitted" | "enrollment" }>, profile: CookieProfile): Response {
  if (outcome.kind === "admitted") return seeOther(`/join/status?status=${outcome.outcome}`, clearCookie("enrollment", profile));
  return seeOther(`/join/status?status=${noticeOf(outcome.status)}`, setCookie("enrollment", outcome.credential.value, outcome.credential.maxAge, profile));
}

// --- form handlers -------------------------------------------------------------------------------------------------------

const BACK_TO_JOIN: StatusLink = { href: "/join", label: "Back to joining" };

// POST /join/start: the existing form-route order (method, form encoding, exact Origin, 16 KiB), then an optional
// invitation intent (both id and token, or neither) and an optional request flag; core starts the enrollment-intent login.
// The answer is the 303 to the hosted page with the login binding, the one-use join cookie for an invitation (any older
// one cleared otherwise) and any previous enrollment cookie cleared (a new login creates a new credential).
export async function joinStartRoute(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const id = field(form.fields, "invitation_id");
  const token = field(form.fields, "invitation_token");
  const req = field(form.fields, "request");
  if (id === AMBIGUOUS || token === AMBIGUOUS || req === AMBIGUOUS || (req !== undefined && req !== "1")) return badRequest();
  const hasInvitation = id !== undefined || token !== undefined;
  if (hasInvitation && (!isId(id) || typeof token !== "string" || !SECRET.test(token))) {
    return onboardingPage(400, "Invitation not read", ["Open the invitation link again, or paste the code exactly as you received it. Nothing was changed."], [BACK_TO_JOIN]);
  }
  const result = await startEnrollment(
    deriveContext(request.headers),
    { origin: form.origin, invitationId: hasInvitation ? id : undefined, request: req === "1" },
    readCookie(form.cookie, "session", form.profile),
    readCookie(form.cookie, "login", form.profile),
  );
  if (result.ok) {
    const { authorizationUrl, binding } = result.value;
    const join = hasInvitation ? setJoinCookie(id as string, token as string, form.profile) : clearCookie("join", form.profile);
    return seeOther(authorizationUrl, setCookie("login", binding.value, binding.maxAge, form.profile), join, clearCookie("enrollment", form.profile));
  }
  switch (result.error) {
    case "already_authenticated":
      return seeOther("/account?status=already_signed_in");
    case "origin_rejected":
      return forbidden();
    case "invalid_request":
    case "ambiguous_parameter":
    case "unsupported_media_type":
      return badRequest();
    default:
      return seeOther("/join?status=unavailable");
  }
}

export type ApplicantOperation = "request" | "admit" | "abandon";

// POST /join/request, /join/admit and /join/abandon: the form layer, the hidden CSRF token, the enrollment cookie, then
// core decides. Every answer is a 303: to the status page with a closed notice (core's state is re-read there; an outcome
// core may have committed is "unknown", never a claimed success), or after abandonment to /join. A dead credential and
// a confirmed admission clear the enrollment cookie; nothing here ever sets a session.
export async function applicantRoute(request: Request, op: ApplicantOperation): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const get = (name: string) => field(form.fields, name);
  const csrfToken = get("csrf_token");
  const kind = get("kind");
  const requestId = get("request_id");
  if (typeof csrfToken !== "string" || csrfToken === "" || kind === AMBIGUOUS || requestId === AMBIGUOUS) return badRequest();
  if (op === "admit" ? kind !== "open" && kind !== "request" : kind !== undefined || requestId !== undefined) return badRequest();
  const credential = readCookie(form.cookie, "enrollment", form.profile);
  const cleared = clearCookie("enrollment", form.profile);
  if (credential === undefined) return seeOther("/join/status?status=expired", cleared);
  const ctx = deriveContext(request.headers);
  const m: Mutation = { origin: form.origin, csrfToken };
  let r: EnrollmentResult<unknown>;
  if (op === "request") r = await requestEnrollment(ctx, credential, m);
  else if (op === "abandon") r = await abandon(ctx, credential, m);
  else r = await admit(ctx, credential, m, kind as "open" | "request", typeof requestId === "string" ? requestId : undefined);
  if (r.ok) {
    if (op === "request") return seeOther("/join/status?status=requested");
    if (op === "abandon") return seeOther("/join?status=abandoned", cleared);
    return seeOther(`/join/status?status=${(r.value as Admitted).outcome}`, cleared);
  }
  switch (r.error) {
    case "unauthenticated":
      return seeOther("/join/status?status=expired", cleared);
    case "csrf_failed":
      return onboardingPage(403, "Form expired", ["This form is no longer valid. Reload the page and try again. Nothing was changed."], [{ href: "/join/status", label: "Back to your status" }]);
    case "origin_rejected":
      return forbidden();
    case "enrollment_closed":
      return seeOther("/join/status?status=closed");
    case "rate_limited":
      return seeOther("/join/status?status=rate_limited");
    case "not_found":
      return seeOther("/join/status?status=not_found");
    case "conflict":
      return seeOther(`/join/status?status=${noticeOf(r.field)}`);
    case "invalid_request":
    case "ambiguous_parameter":
      return badRequest();
    default:
      return seeOther("/join/status?status=unknown");
  }
}

// The closed ?status= codes of the enrollment management page.
export const MANAGE_NOTICES: Readonly<Record<string, string>> = {
  policy_changed: "The enrollment policy was changed.",
  approved: "The request was approved. The person joins when they return and confirm.",
  denied: "The request was denied.",
  revoked: "The approval was withdrawn.",
};

// The closed conflict/refusal fields of the management actions and the fixed sentence each shows.
const MANAGE_REFUSALS: Readonly<Record<string, string>> = {
  enrollment_not_ready: "Open signup is unavailable until the platform operator has reviewed membership history",
  not_pending: "That request is no longer pending.",
  already_consumed: "That approval was already used to join; remove the member on the members page instead.",
  not_approved: "That request is not an approval that can be withdrawn.",
  invalid_enrollment_mode: "Choose one of the offered enrollment modes.",
};

export function manageRefusalSentence(field: string | undefined): string | undefined {
  return field !== undefined && Object.hasOwn(MANAGE_REFUSALS, field) ? MANAGE_REFUSALS[field] : undefined;
}

const MANAGE_PATH = "/organization/enrollment";
const BACK_TO_MANAGE: StatusLink = { href: MANAGE_PATH, label: "Back to enrollment" };

function manageRefusal(error: EnrollmentError, field: string | undefined): Response {
  const sentence = manageRefusalSentence(field);
  switch (error) {
    case "conflict":
      return onboardingPage(409, "Not changed", [sentence ?? "The current state does not allow this change.", "Nothing was changed."], [BACK_TO_MANAGE]);
    case "unauthenticated":
      return onboardingPage(401, "Not signed in", ["Sign in first, then try again. Nothing was changed."], [{ href: "/sign-in", label: "Sign in" }]);
    case "csrf_failed":
      return onboardingPage(403, "Form expired", ["This form is no longer valid. Reload the page and try again. Nothing was changed."], [BACK_TO_MANAGE]);
    case "origin_rejected":
      return onboardingPage(403, "Request refused", ["This request did not come from this site's own page. Nothing was changed."], [BACK_TO_MANAGE]);
    case "forbidden":
    case "not_available":
      return onboardingPage(403, "Not permitted", [sentence ?? "Your role does not permit this change.", "Nothing was changed."], [BACK_TO_MANAGE]);
    case "not_found":
      return onboardingPage(404, "Not found", ["That request does not exist or is not available to you. Nothing was changed."], [BACK_TO_MANAGE]);
    case "invalid_request":
      return onboardingPage(400, "Not changed", [sentence ?? "This request could not be read.", "Nothing was changed."], [BACK_TO_MANAGE]);
    default:
      return onboardingPage(400, "Request not understood", ["This request could not be read. Nothing was changed."], [BACK_TO_MANAGE]);
  }
}

// POST /organization/enrollment/action: the form layer, the closed action field (policy | approve | deny | revoke) and its
// shape, the session, then core decides (authority, CSRF, current state). Success is a 303 with a closed status; an
// outcome core may have committed is the unknown-state page, which never claims success and never retries.
export async function manageActionRoute(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const get = (name: string) => field(form.fields, name);
  const csrfToken = get("csrf_token");
  const action = get("action");
  const mode = get("mode");
  const id = get("request_id");
  if (typeof csrfToken !== "string" || csrfToken === "" || mode === AMBIGUOUS || id === AMBIGUOUS) return badRequest();
  const shapeOk =
    action === "policy" ? typeof mode === "string" && id === undefined : (action === "approve" || action === "deny" || action === "revoke") && mode === undefined && isId(id);
  if (!shapeOk) return badRequest();
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return manageRefusal("unauthenticated", undefined);
  const ctx = deriveContext(request.headers);
  const m: Mutation = { origin: form.origin, csrfToken };
  let r: EnrollmentResult<unknown>;
  let status: string;
  if (action === "policy") {
    if (!isMode(mode)) return manageRefusal("invalid_request", "invalid_enrollment_mode");
    r = await updatePolicy(ctx, credential, m, mode);
    status = "policy_changed";
  } else if (action === "revoke") {
    r = await revokeRequest(ctx, credential, m, id as string);
    status = "revoked";
  } else {
    r = await decideRequest(ctx, credential, m, id as string, action === "approve" ? "approve" : "deny");
    status = action === "approve" ? "approved" : "denied";
  }
  if (r.ok) return seeOther(`${MANAGE_PATH}?status=${status}`);
  if (outcomeUnknown(r.error)) {
    return onboardingPage(503, "Outcome unknown", ["Your request may or may not have completed. Check the current state before trying again."], [BACK_TO_MANAGE]);
  }
  return manageRefusal(r.error, r.field);
}
