import "server-only";

import { runtimeConfig } from "./config";
import { errorCategory, outboundHeaders, type RequestContext } from "./request-context";

// Server-only client for core's internal auth contract (/internal/v1/auth/*; Plan r5 §7.2). Each call sends
// exactly the request's diagnostic headers (outboundHeaders(ctx)), "Authorization: Session <credential>" when
// a credential is present, X-Login-Binding on login and callback only, and Content-Type: application/json with
// a JSON body of the documented fields. Nothing else from the browser is forwarded. Answers are validated and
// reduced to closed outcomes; core's error detail is never surfaced beyond its closed code.

export const CORE_TIMEOUT_MS = 10_000;
const MAX_RESPONSE_BYTES = 64 * 1024;
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const RFC3339 = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z$/;

const CORE_ERRORS = [
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
  "host_changed", // G016: the host no longer routes to the login's Organization
  "organization_not_member", // G016 D-2(a): no current membership of that Organization
  "unknown_host", // G016: the session route's origin routes nowhere
  "rate_limited", // G062: the root signup login start's bucket (Plan r5 §4.5 (i))
] as const;
const KNOWN_ERRORS: ReadonlySet<string> = new Set(CORE_ERRORS);

// A closed core error code, or the call's own failure: no answer in time, no connection, or an answer outside
// the contract.
export type CoreError = (typeof CORE_ERRORS)[number] | "timeout" | "unreachable" | "malformed";
export type CoreResult<T> = { ok: true; value: T } | { ok: false; error: CoreError };

export type IssuedSecret = { value: string; maxAge: number };
export type ReturnPath = "/" | "/account";
export type LoginStarted = { authorizationUrl: string; binding: IssuedSecret };
export type CallbackCompleted =
  | { outcome: "signed_in" | "session_rotated"; returnTo: ReturnPath; session: IssuedSecret }
  | { outcome: "already_signed_in"; returnTo: ReturnPath };
export type SessionView = {
  userId: string;
  displayName: string | null;
  mfaAssurance: "unknown";
  authTime: string;
  expiresAt: string;
  csrfToken: string;
  // G017 D-6: whether the user currently belongs to the origin's Organization (an indicator fact; it grants nothing).
  membership: "member" | "none";
};
export type LogoutScope = "current" | "all" | "provider";
export type LogoutCompleted =
  | { outcome: "local_revoked" }
  | { outcome: "provider_logout_redirected"; providerLogoutUrl: string; logoutRef: string }
  | { outcome: "provider_session_revoked" | "provider_revoke_failed"; logoutRef: string };
export type ProviderReturnOutcome = "local_revoked" | "provider_session_revoked" | "provider_revoke_failed";

type Call = {
  method: "GET" | "POST";
  path: string;
  credential?: string;
  binding?: string;
  body?: Readonly<Record<string, string>>;
};

type Answer = { kind: "answered"; status: number; body: unknown } | { kind: "failed"; error: "timeout" | "unreachable" | "malformed" };

type Json = Readonly<Record<string, unknown>>;

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

async function readBounded(response: Response): Promise<string | undefined> {
  if (response.body === null) return "";
  const reader = response.body.getReader();
  const chunks: Uint8Array[] = [];
  let size = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    size += value.byteLength;
    if (size > MAX_RESPONSE_BYTES) {
      await reader.cancel();
      return undefined;
    }
    chunks.push(value);
  }
  return new TextDecoder().decode(Buffer.concat(chunks));
}

async function call(ctx: RequestContext, c: Call): Promise<Answer> {
  const headers: Record<string, string> = { ...outboundHeaders(ctx) };
  if (c.credential !== undefined) headers.Authorization = `Session ${c.credential}`;
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
      signal: AbortSignal.timeout(CORE_TIMEOUT_MS),
    });
  } catch (error) {
    return { kind: "failed", error: errorCategory(error) === "TimeoutError" ? "timeout" : "unreachable" };
  }
  try {
    const text = await readBounded(response);
    if (text === undefined) return { kind: "failed", error: "malformed" };
    return { kind: "answered", status: response.status, body: JSON.parse(text) as unknown };
  } catch (error) {
    if (errorCategory(error) === "TimeoutError") return { kind: "failed", error: "timeout" };
    return { kind: "failed", error: "malformed" };
  }
}

// A non-2xx answer must be exactly {"error": <closed code>}.
function failure(answer: Answer): { ok: false; error: CoreError } {
  if (answer.kind === "failed") return { ok: false, error: answer.error };
  const body = answer.body;
  if (answer.status >= 400 && isObject(body) && Object.keys(body).length === 1 && typeof body.error === "string" && KNOWN_ERRORS.has(body.error)) {
    return { ok: false, error: body.error as CoreError };
  }
  return { ok: false, error: "malformed" };
}

function success(answer: Answer): Json | undefined {
  return answer.kind === "answered" && answer.status === 200 && isObject(answer.body) ? answer.body : undefined;
}

function issuedSecret(value: unknown): IssuedSecret | undefined {
  if (!isObject(value)) return undefined;
  const { value: secret, max_age: maxAge } = value;
  if (typeof secret !== "string" || !SECRET.test(secret)) return undefined;
  if (typeof maxAge !== "number" || !Number.isInteger(maxAge) || maxAge < 1 || maxAge > 86_400) return undefined;
  return { value: secret, maxAge };
}

function returnPath(value: unknown): ReturnPath | undefined {
  return value === "/" || value === "/account" ? value : undefined;
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

// A one-use logout reference, exactly as core encodes it (43 unpadded base64url characters). It travels only in its
// own answer field and the host-only logout cookie, never in a URL (C7 addendum §2).
function logoutRefOf(value: unknown): string | undefined {
  return typeof value === "string" && /^[A-Za-z0-9_-]{43}$/.test(value) ? value : undefined;
}

// POST /internal/v1/auth/login: start a login for this origin, reusing the browser's binding when present. G062 (Plan r5
// §1.2, §4.2): intent "signup" starts the platform signup login, which core accepts only at the exact root origin
// (origin_rejected elsewhere) and whose return path core fixes (/signup/details); it carries no return_to.
export async function login(
  ctx: RequestContext,
  input: { origin: string; returnTo?: ReturnPath; intent?: "signup" },
  credential?: string,
  binding?: string,
): Promise<CoreResult<LoginStarted>> {
  const body: Record<string, string> = { origin: input.origin };
  if (input.intent === "signup") {
    if (input.returnTo !== undefined) return { ok: false, error: "invalid_request" };
    body.intent = "signup";
  }
  if (input.returnTo !== undefined) body.return_to = input.returnTo;
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/auth/login", credential, binding, body });
  const ok = success(answer);
  if (ok === undefined) return failure(answer);
  const authorizationUrl = absoluteHttpUrl(ok.authorization_url);
  const issued = issuedSecret(ok.binding);
  if (authorizationUrl === undefined || issued === undefined) return { ok: false, error: "malformed" };
  return { ok: true, value: { authorizationUrl, binding: issued } };
}

// POST /internal/v1/auth/callback: complete the login named by the callback query for this browser's binding.
export async function callback(
  ctx: RequestContext,
  input: { origin: string; rawQuery: string },
  binding?: string,
  credential?: string,
): Promise<CoreResult<CallbackCompleted>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/auth/callback",
    credential,
    binding,
    body: { origin: input.origin, raw_query: input.rawQuery },
  });
  const ok = success(answer);
  if (ok === undefined) return failure(answer);
  const returnTo = returnPath(ok.return_to);
  if (returnTo === undefined) return { ok: false, error: "malformed" };
  if (ok.outcome === "already_signed_in") return { ok: true, value: { outcome: "already_signed_in", returnTo } };
  const issued = issuedSecret(ok.session);
  if ((ok.outcome === "signed_in" || ok.outcome === "session_rotated") && issued !== undefined) {
    return { ok: true, value: { outcome: ok.outcome, returnTo, session: issued } };
  }
  return { ok: false, error: "malformed" };
}

// GET /internal/v1/auth/session?origin=: the caller's own session, including the CSRF token for server-rendered forms.
// G016: core answers only for a session bound to that origin's Organization while the host routes (A2: a session issued
// for another Organization is unauthenticated; a removed or reassigned host is unknown_host). The origin is required.
export async function session(ctx: RequestContext, credential: string, origin: string): Promise<CoreResult<SessionView>> {
  const path = `/internal/v1/auth/session?origin=${encodeURIComponent(origin)}`;
  const answer = await call(ctx, { method: "GET", path, credential });
  const ok = success(answer);
  if (ok === undefined) return failure(answer);
  const { user_id: userId, display_name: displayName, mfa_assurance: mfa, auth_time: authTime, expires_at: expiresAt, csrf_token: csrf, membership } = ok;
  if (
    (membership !== "member" && membership !== "none") ||
    typeof userId !== "string" ||
    !UUID.test(userId) ||
    !(displayName === null || (typeof displayName === "string" && displayName.length <= 400)) ||
    mfa !== "unknown" ||
    typeof authTime !== "string" ||
    !RFC3339.test(authTime) ||
    typeof expiresAt !== "string" ||
    !RFC3339.test(expiresAt) ||
    typeof csrf !== "string" ||
    !SECRET.test(csrf)
  ) {
    return { ok: false, error: "malformed" };
  }
  return { ok: true, value: { userId, displayName, mfaAssurance: "unknown", authTime, expiresAt, csrfToken: csrf, membership } };
}

// POST /internal/v1/auth/logout: one sign-out scope, CSRF-checked by core against the caller's own session.
export async function logout(
  ctx: RequestContext,
  input: { scope: LogoutScope; csrfToken: string; origin: string },
  credential?: string,
): Promise<CoreResult<LogoutCompleted>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/auth/logout",
    credential,
    body: { scope: input.scope, csrf_token: input.csrfToken, origin: input.origin },
  });
  const ok = success(answer);
  if (ok === undefined) return failure(answer);
  switch (ok.outcome) {
    case "local_revoked":
      return { ok: true, value: { outcome: "local_revoked" } };
    case "provider_logout_redirected": {
      const providerLogoutUrl = absoluteHttpUrl(ok.provider_logout_url);
      const logoutRef = logoutRefOf(ok.logout_ref);
      if (providerLogoutUrl !== undefined && logoutRef !== undefined && ok.return_to === "/sign-in" && !providerLogoutUrl.includes(logoutRef)) {
        return { ok: true, value: { outcome: ok.outcome, providerLogoutUrl, logoutRef } };
      }
      break;
    }
    case "provider_session_revoked":
    case "provider_revoke_failed": {
      const logoutRef = logoutRefOf(ok.logout_ref);
      if (logoutRef !== undefined && ok.return_to === "/sign-in" && ok.provider_logout_url === undefined) {
        return { ok: true, value: { outcome: ok.outcome, logoutRef } };
      }
      break;
    }
  }
  return { ok: false, error: "malformed" };
}

// POST /internal/v1/auth/provider-logout/return: consume a one-use logout reference once; unknown, expired or
// reused references come back as the generic local outcome.
export async function providerLogoutReturn(ctx: RequestContext, logoutRef: string): Promise<CoreResult<ProviderReturnOutcome>> {
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/auth/provider-logout/return", body: { logout_ref: logoutRef } });
  const ok = success(answer);
  if (ok === undefined) return failure(answer);
  const outcome = ok.outcome;
  if (outcome === "local_revoked" || outcome === "provider_session_revoked" || outcome === "provider_revoke_failed") {
    return { ok: true, value: outcome };
  }
  return { ok: false, error: "malformed" };
}
