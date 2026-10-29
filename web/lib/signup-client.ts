import "server-only";

import { runtimeConfig } from "./config";
import { CALLBACK_STATUSES, type CallbackOutcome, type CallbackStatus } from "./enrollment-client";
import { isHttpsOrigin, lookupContext, resolveHost } from "./host";
import { errorCategory, outboundHeaders, type RequestContext } from "./request-context";
import { cookieProfile, statusPage } from "./session";

// G062 (Plan r5 §1.1, §1.4, §4.2, §4.4): the server-only client of core's platform root signup contract
// (/internal/v1/signup/*: seven method/path pairs, one function each), the signup variants of the auth callback answer
// (the root's 202 signup credential and the tenant's pinned tenant_created session), the platform root resolution of the
// root and tenant signup routes, and the closed notices of the signup pages. Each call sends exactly the request's
// diagnostic headers; the root routes add "Authorization: Signup <credential>", the tenant continuation hops add only
// X-Login-Binding (never a credential), and a body is JSON with the route's documented members only. Core alone resolves
// hosts, credentials and continuation secrets and decides every write; web never derives the root or a tenant origin
// from input. Answers are validated against strict shapes and bounds; anything outside the contract is "malformed".
// The continuation secrets (ref, n, p, the challenge) and the submitted name and slug are never logged; the secrets
// travel only in the 303 Location core's own answer names and in the host-only challenge cookie.

export const SIGNUP_TIMEOUT_MS = 10_000;
const MAX_RESPONSE_BYTES = 16 * 1024;
const SECRET = /^[A-Za-z0-9_-]{43}$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const LABEL = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;
// G057 (Plan r5 §3.2, §3.3; D-1, D-5; friendly-address amendment 2026-10-29): the generated platform label core issues for
// every new Organization, <adjective>-<noun>-<four digits> (for example gusty-flowerpot-2643), or the 26 lowercase RFC 4648
// base32 characters of a label issued before that amendment, which keeps resolving. The friendly slug is metadata: it
// never forms an origin (a host equal to the Organization's own slug is refused even with this shape).
const PLATFORM_LABEL = /^(?:[a-z2-7]{26}|[a-z]{3,10}-[a-z]{3,10}-[0-9]{4})$/;

// core/internal/signup/api's closed refusal codes (Plan r5 §4.2).
const SIGNUP_ERRORS = [
  "invalid_request",
  "origin_rejected",
  "unauthenticated",
  "csrf_failed",
  "user_disabled",
  "conflict",
  "continuation_invalid",
  "signup_context_mismatch",
  "rate_limited",
  "not_found",
  "method_not_allowed",
  "signup_unconfigured",
  "signup_unavailable",
] as const;
const SIGNUP_KNOWN: ReadonlySet<string> = new Set(SIGNUP_ERRORS);
// The closed fields: invalid_request names the field (name | slug), conflict the conflict (slug_taken | intent_used).
const FIELDS: Readonly<Record<string, ReadonlySet<string>>> = {
  invalid_request: new Set(["name", "slug"]),
  conflict: new Set(["slug_taken", "intent_used"]),
};
// The auth callback's closed codes: identity's set, the two G058 callback refusals and the two G062 additions.
const CALLBACK_ERRORS = [
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
  "signup_user_mismatch",
  "rate_limited",
] as const;
const CALLBACK_KNOWN: ReadonlySet<string> = new Set(CALLBACK_ERRORS);

export type SignupError = (typeof SIGNUP_ERRORS)[number] | (typeof CALLBACK_ERRORS)[number] | "timeout" | "unreachable" | "malformed";
export type SignupResult<T> = { ok: true; value: T } | { ok: false; error: SignupError; field?: string };
export type IssuedSecret = { value: string; maxAge: number };
export type SignupStatus = {
  status: "pending" | "created" | "completed";
  csrfToken: string;
  organization: { slug: string; targetOrigin: string } | null;
};
export type Suggestion = { suggested: string | null; available: boolean };
export type Created = { outcome: "created" | "already_created"; organizationId: string; slug: string; continueUrl: string };
export type Challenged = { challenge: IssuedSecret; proveUrl: string };
export type Mutation = { origin: string; csrfToken: string };
// The callback answer: G058's four kinds (lib/enrollment-client.ts), the pinned tenant callback's session outcome
// tenant_created, and the root's signup credential.
export type CallbackReturn = "/" | "/account" | "/join/status";
export type AuthCallbackOutcome =
  | { kind: "session"; outcome: "signed_in" | "session_rotated" | "tenant_created"; returnTo: CallbackReturn; session: IssuedSecret }
  | Exclude<CallbackOutcome, { kind: "session" }>
  | { kind: "signup"; credential: IssuedSecret };

type Json = Readonly<Record<string, unknown>>;
// scheme is "Signup" for the signup routes; the auth callback forwards the browser's session as "Session".
type Call = { method: "GET" | "POST"; path: string; credential?: string; scheme?: "Signup" | "Session"; binding?: string; body?: Json };
type Answer = { kind: "answered"; status: number; body: unknown } | { kind: "failed"; error: "timeout" | "unreachable" | "malformed" };

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasExactly(o: Json, keys: readonly string[]): boolean {
  const own = Object.keys(o);
  return own.length === keys.length && keys.every((k) => Object.hasOwn(o, k));
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
  if (c.credential !== undefined) headers.Authorization = `${c.scheme ?? "Signup"} ${c.credential}`;
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
      signal: AbortSignal.timeout(SIGNUP_TIMEOUT_MS),
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

// A refusal must be exactly {"error": <closed code>} or, for invalid_request and conflict, {"error", "field"} with a
// closed field.
function failure(answer: Answer, known: ReadonlySet<string> = SIGNUP_KNOWN): { ok: false; error: SignupError; field?: string } {
  if (answer.kind === "failed") return { ok: false, error: answer.error };
  const body = answer.body;
  if (answer.status >= 400 && isObject(body) && typeof body.error === "string" && known.has(body.error)) {
    const error = body.error as SignupError;
    if (hasExactly(body, ["error"])) return { ok: false, error };
    const fields = known === SIGNUP_KNOWN ? FIELDS[error] : undefined;
    if (fields !== undefined && hasExactly(body, ["error", "field"]) && typeof body.field === "string" && fields.has(body.field)) {
      return { ok: false, error, field: body.field };
    }
  }
  return { ok: false, error: "malformed" };
}

function okAnswer(answer: Answer, status = 200): Json | undefined {
  return answer.kind === "answered" && answer.status === status && isObject(answer.body) ? answer.body : undefined;
}

function issuedSecret(value: unknown, maxAge: number): IssuedSecret | undefined {
  if (!isObject(value) || !hasExactly(value, ["value", "max_age"])) return undefined;
  const { value: secret, max_age: age } = value;
  if (typeof secret !== "string" || !SECRET.test(secret)) return undefined;
  if (typeof age !== "number" || !Number.isInteger(age) || age < 1 || age > maxAge) return undefined;
  return { value: secret, maxAge: age };
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

// --- platform root and tenant origins -----------------------------------------------------------------------------------

// The root's host[:port] (the origin without its scheme).
function hostOf(origin: string): string {
  return origin.slice("https://".length);
}

// G057 (Plan r5 §3.3, D-5): the structural check of a tenant origin that core answers (status.target_origin, the continue
// and complete URLs) or that a continuation hop arrives on. It holds only for https://<label>.<root host>[:<root port>]:
// an exact https origin (no userinfo, path, query or fragment), the root's own port, exactly one label more than the root
// and that label a generated platform label. A host equal to a known friendly slug of the Organization is refused even
// when it has the label's shape, so web never accepts or forwards a slug-derived origin. This only closes the shape of
// core's answer; core alone resolves hosts and decides which Organization an origin names.
export function isTenantOriginUnder(root: string, origin: string, ...slugs: string[]): boolean {
  if (!isHttpsOrigin(root) || !isHttpsOrigin(origin)) return false;
  const suffix = `.${hostOf(root)}`;
  const host = hostOf(origin);
  if (!host.endsWith(suffix)) return false;
  const label = host.slice(0, -suffix.length);
  return PLATFORM_LABEL.test(label) && !slugs.includes(label);
}

// A continuation URL answered by core: exactly <origin><path>?<param>=<43-character secret>, with origin accepted by
// originOk. Returns the URL unchanged (never re-serialized) or undefined.
export function continuationUrl(value: unknown, path: string, param: string, originOk: (origin: string) => boolean): string | undefined {
  if (typeof value !== "string" || value.length > 400) return undefined;
  const marker = `${path}?${param}=`;
  const at = value.indexOf(marker);
  if (at <= 0) return undefined;
  const origin = value.slice(0, at);
  const secret = value.slice(at + marker.length);
  return isHttpsOrigin(origin) && originOk(origin) && SECRET.test(secret) ? value : undefined;
}

export type RootResolution = { ok: true; root: string } | { ok: false; error: "not_found" | "unavailable" };

// The root signup routes (Plan r5 §4.4): served only on the https profile at the platform root itself, i.e. a public
// entry host whose own origin is core's signup_origin (www.<domain> is not the root). The lookup precedes the route's
// own core call, so it carries its own request ID.
export async function rootOfRootRequest(ctx: RequestContext, origin: string): Promise<RootResolution> {
  if (profileOf() !== "https") return { ok: false, error: "not_found" };
  const host = await resolveHost(lookupContext(ctx), origin);
  if (!host.ok) return { ok: false, error: host.error === "unavailable" ? "unavailable" : "not_found" };
  return host.value.kind === "public_entry" && host.value.signupOrigin === origin ? { ok: true, root: origin } : { ok: false, error: "not_found" };
}

// The tenant continuation hops (Plan r5 §4.4): the root origin for their redirects is configuration, never a query
// parameter. The candidate is https://<PLATFORM_BASE_DOMAIN> with the request's own port (the platform host port), and it
// is the root only when core resolves it as the public entry whose signup_origin is exactly that origin.
export async function rootOfTenantRequest(ctx: RequestContext, origin: string): Promise<RootResolution> {
  if (profileOf() !== "https" || !isHttpsOrigin(origin)) return { ok: false, error: "not_found" };
  const port = /:[0-9]{1,5}$/.exec(origin)?.[0] ?? "";
  const candidate = `https://${runtimeConfig().platformBaseDomain.trim().toLowerCase()}${port}`;
  if (!isHttpsOrigin(candidate) || !isTenantOriginUnder(candidate, origin)) return { ok: false, error: "not_found" };
  const host = await resolveHost(lookupContext(ctx), candidate);
  if (!host.ok) return { ok: false, error: host.error === "unavailable" ? "unavailable" : "not_found" };
  return host.value.kind === "public_entry" && host.value.signupOrigin === candidate ? { ok: true, root: candidate } : { ok: false, error: "not_found" };
}

// The public entry facts of a page request on the https profile: whether this origin is the root itself and the root
// origin the landing links Sign Up to; undefined on the loopback profile, for any other host or on a lookup failure.
export async function publicEntryOf(ctx: RequestContext, origin: string | undefined): Promise<{ root: boolean; signupOrigin: string } | undefined> {
  if (origin === undefined || profileOf() !== "https") return undefined;
  const host = await resolveHost(lookupContext(ctx), origin);
  if (!host.ok || host.value.kind !== "public_entry") return undefined;
  return { root: origin === host.value.signupOrigin, signupOrigin: host.value.signupOrigin };
}

function profileOf(): "https" | "loopback-http" | undefined {
  try {
    return cookieProfile();
  } catch {
    return undefined;
  }
}

// --- the seven signup routes ----------------------------------------------------------------------------------------------

// GET /internal/v1/signup/status?origin= (own signup credential): the reconciliation read of the root forms. A dead
// credential is unauthenticated (the caller clears the signup cookies).
export async function signupStatus(ctx: RequestContext, credential: string, origin: string): Promise<SignupResult<SignupStatus>> {
  const answer = await call(ctx, { method: "GET", path: `/internal/v1/signup/status?origin=${encodeURIComponent(origin)}`, credential });
  const b = okAnswer(answer);
  if (b === undefined) return failure(answer);
  const { status, csrf_token: csrf, organization: org } = b;
  if (status !== "pending" && status !== "created" && status !== "completed") return { ok: false, error: "malformed" };
  if (typeof csrf !== "string" || !SECRET.test(csrf)) return { ok: false, error: "malformed" };
  if (status === "pending") {
    return hasExactly(b, ["status", "csrf_token"]) ? { ok: true, value: { status, csrfToken: csrf, organization: null } } : { ok: false, error: "malformed" };
  }
  if (!hasExactly(b, ["status", "csrf_token", "organization"]) || !isObject(org) || !hasExactly(org, ["slug", "target_origin"])) return { ok: false, error: "malformed" };
  const { slug, target_origin: target } = org;
  if (typeof slug !== "string" || !LABEL.test(slug) || typeof target !== "string" || !isTenantOriginUnder(origin, target, slug)) return { ok: false, error: "malformed" };
  return { ok: true, value: { status, csrfToken: csrf, organization: { slug, targetOrigin: target } } };
}

// POST /internal/v1/signup/slug {origin, csrf_token, name}: the advisory suggestion (nothing is reserved).
export async function suggestSlug(ctx: RequestContext, credential: string, m: Mutation, name: string): Promise<SignupResult<Suggestion>> {
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/signup/slug", credential, body: { origin: m.origin, csrf_token: m.csrfToken, name } });
  const b = okAnswer(answer);
  if (b === undefined) return failure(answer);
  const { suggested, available } = b;
  if (!hasExactly(b, ["suggested", "available"]) || typeof available !== "boolean" || !(suggested === null || (typeof suggested === "string" && LABEL.test(suggested)))) {
    return { ok: false, error: "malformed" };
  }
  return { ok: true, value: { suggested, available } };
}

// POST /internal/v1/signup/organizations {origin, csrf_token, name, slug}: constrained provisioning. 201 created or
// already_created (an identical intent re-mints the continuation); the continue URL must be exactly a generated label
// origin of this root (G057 D-5: never the slug's host; on already_created core answers the stored origin) +
// /signup/continue?ref=<secret>.
export async function createOrganization(ctx: RequestContext, credential: string, m: Mutation, name: string, slug: string): Promise<SignupResult<Created>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/signup/organizations",
    credential,
    body: { origin: m.origin, csrf_token: m.csrfToken, name, slug },
  });
  const b = okAnswer(answer, 201);
  if (b === undefined) return failure(answer);
  const { outcome, organization: org } = b;
  if (!hasExactly(b, ["outcome", "organization", "continue"]) || (outcome !== "created" && outcome !== "already_created")) return { ok: false, error: "malformed" };
  if (!isObject(org) || !hasExactly(org, ["id", "slug"]) || typeof org.id !== "string" || !UUID.test(org.id) || typeof org.slug !== "string" || !LABEL.test(org.slug)) {
    return { ok: false, error: "malformed" };
  }
  const createdSlug = org.slug;
  const continueUrl = continuationUrl(b.continue, "/signup/continue", "ref", (o) => isTenantOriginUnder(m.origin, o, createdSlug, slug));
  if (continueUrl === undefined) return { ok: false, error: "malformed" };
  return { ok: true, value: { outcome, organizationId: org.id, slug: createdSlug, continueUrl } };
}

// POST /internal/v1/signup/challenge {origin, ref} (tenant origin; X-Login-Binding; no credential): continuation step 2.
// The prove URL must be exactly the root origin + /signup/prove?n=<secret>.
export async function challenge(ctx: RequestContext, input: { origin: string; ref: string; root: string }, binding?: string): Promise<SignupResult<Challenged>> {
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/signup/challenge", binding, body: { origin: input.origin, ref: input.ref } });
  const b = okAnswer(answer);
  if (b === undefined) return failure(answer);
  const issued = hasExactly(b, ["challenge", "prove"]) ? issuedSecret(b.challenge, 120) : undefined;
  const proveUrl = continuationUrl(b.prove, "/signup/prove", "n", (o) => o === input.root);
  if (issued === undefined || proveUrl === undefined) return { ok: false, error: "malformed" };
  return { ok: true, value: { challenge: issued, proveUrl } };
}

// POST /internal/v1/signup/prove {origin, nonce} (own signup credential; no CSRF member): continuation step 3. The
// complete URL must be exactly a generated label origin of this root + /signup/complete?p=<secret>.
export async function prove(ctx: RequestContext, credential: string, origin: string, nonce: string): Promise<SignupResult<string>> {
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/signup/prove", credential, body: { origin, nonce } });
  const b = okAnswer(answer);
  if (b === undefined) return failure(answer);
  const url = hasExactly(b, ["complete"]) ? continuationUrl(b.complete, "/signup/complete", "p", (o) => isTenantOriginUnder(origin, o)) : undefined;
  return url === undefined ? { ok: false, error: "malformed" } : { ok: true, value: url };
}

// POST /internal/v1/signup/complete {origin, challenge, proof} (tenant origin; X-Login-Binding; no credential):
// continuation step 4, answered with the ordinary login start of the pinned login.
export async function complete(
  ctx: RequestContext,
  input: { origin: string; challenge: string; proof: string },
  binding?: string,
): Promise<SignupResult<{ authorizationUrl: string; binding: IssuedSecret }>> {
  const answer = await call(ctx, {
    method: "POST",
    path: "/internal/v1/signup/complete",
    binding,
    body: { origin: input.origin, challenge: input.challenge, proof: input.proof },
  });
  const b = okAnswer(answer);
  if (b === undefined) return failure(answer);
  const authorizationUrl = absoluteHttpUrl(b.authorization_url);
  const issued = issuedSecret(b.binding, 86_400);
  if (!hasExactly(b, ["authorization_url", "binding"]) || authorizationUrl === undefined || issued === undefined) return { ok: false, error: "malformed" };
  return { ok: true, value: { authorizationUrl, binding: issued } };
}

// POST /internal/v1/signup/abandon {origin, csrf_token}: revokes the caller's own credential and its current generation.
export async function abandonSignup(ctx: RequestContext, credential: string, m: Mutation): Promise<SignupResult<null>> {
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/signup/abandon", credential, body: { origin: m.origin, csrf_token: m.csrfToken } });
  const b = okAnswer(answer);
  if (b === undefined) return failure(answer);
  return hasExactly(b, []) ? { ok: true, value: null } : { ok: false, error: "malformed" };
}

// --- the auth callback ---------------------------------------------------------------------------------------------------

function callbackReturn(value: unknown): CallbackReturn | undefined {
  return value === "/" || value === "/account" || value === "/join/status" ? value : undefined;
}

// POST /internal/v1/auth/callback for every completion on a classified host that is not the shared auth host (only core
// knows what the stored login carried). The answers of lib/enrollment-client.ts completeCallback (ordinary and
// enrollment), plus G062's: the root signup login's 202 {"signup": {"credential": {value, max_age ≤ 1800}, "next":
// "/signup/details"}} (the signup credential only, never a session) and the pinned tenant callback's ordinary session
// answer with outcome tenant_created (return_to "/"). signup_user_mismatch and rate_limited are closed refusals.
export async function authCallback(
  ctx: RequestContext,
  input: { origin: string; rawQuery: string; invitation?: { invitationId: string; token: string } },
  binding?: string,
  sessionCredential?: string,
): Promise<SignupResult<AuthCallbackOutcome>> {
  const body: Record<string, string> = { origin: input.origin, raw_query: input.rawQuery };
  if (input.invitation !== undefined) {
    body.invitation_id = input.invitation.invitationId;
    body.invitation_token = input.invitation.token;
  }
  // The callback carries the browser's live session, if any, as "Session <credential>" (never the signup scheme).
  const answer = await call(ctx, { method: "POST", path: "/internal/v1/auth/callback", credential: sessionCredential, scheme: "Session", binding, body });
  const accepted = okAnswer(answer, 202);
  if (accepted !== undefined) {
    if (hasExactly(accepted, ["signup"])) {
      const s = accepted.signup;
      const credential = isObject(s) && hasExactly(s, ["credential", "next"]) && s.next === "/signup/details" ? issuedSecret(s.credential, 1_800) : undefined;
      return credential === undefined ? { ok: false, error: "malformed" } : { ok: true, value: { kind: "signup", credential } };
    }
    const e = hasExactly(accepted, ["enrollment"]) ? accepted.enrollment : undefined;
    if (isObject(e) && hasExactly(e, ["status", "credential"]) && (CALLBACK_STATUSES as readonly unknown[]).includes(e.status)) {
      const issued = issuedSecret(e.credential, 1_800);
      if (issued !== undefined) return { ok: true, value: { kind: "enrollment", status: e.status as CallbackStatus, credential: issued } };
    }
    return { ok: false, error: "malformed" };
  }
  const b = okAnswer(answer);
  if (b === undefined) return failure(answer, CALLBACK_KNOWN);
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
  const sessionOutcome = b.outcome === "signed_in" || b.outcome === "session_rotated" || (b.outcome === "tenant_created" && returnTo === "/");
  if (sessionOutcome && issued !== undefined && hasExactly(b, ["outcome", "return_to", "session"])) {
    return { ok: true, value: { kind: "session", outcome: b.outcome as "signed_in" | "session_rotated" | "tenant_created", returnTo, session: issued } };
  }
  return { ok: false, error: "malformed" };
}

// --- route helpers -------------------------------------------------------------------------------------------------------

// The fixed answers of a signup route outside its host (404) or with the root lookup unavailable (503).
export function refusedRoot(error: "not_found" | "unavailable"): Response {
  return error === "unavailable"
    ? statusPage(503, "Unavailable", "This service is not available right now. Nothing was changed.")
    : statusPage(404, "Not found", "There is nothing at this address.");
}

// A form whose CSRF token core refused (a reloaded or stale page).
export function formExpired(): Response {
  return statusPage(403, "Form expired", "This form is no longer valid. Reload the page and try again. Nothing was changed.");
}

// The one continuation secret of a hop's query: the raw query is exactly <name>=<43-character secret> (one parameter,
// no encoding, nothing else); undefined otherwise.
export function hopSecret(rawQuery: string, name: string): string | undefined {
  const prefix = `${name}=`;
  if (!rawQuery.startsWith(prefix)) return undefined;
  const value = rawQuery.slice(prefix.length);
  return SECRET.test(value) ? value : undefined;
}

// --- closed notices ------------------------------------------------------------------------------------------------------

// The closed ?status= codes of /signup and the sentence each shows; never a submitted value.
export const SIGNUP_NOTICES: Readonly<Record<string, string>> = {
  canceled: "Sign-up was canceled. Nothing was created.",
  expired: "Your sign-up session ended or expired. Start again.",
  abandoned: "You stopped signing up. Nothing more was created.",
  failed: "Sign-up could not be completed. Please try again.",
  user_disabled: "This account cannot create an organization.",
  rate_limited: "Too many attempts. Wait a moment and try again.",
  unavailable: "Sign-up is unavailable right now. Please try again later.",
};

// The closed ?status= codes of /signup/details and the sentence each shows; never a submitted value.
export const DETAILS_NOTICES: Readonly<Record<string, string>> = {
  slug_taken: "That short name is already taken. Choose another one.",
  invalid_name: "Enter a name of 1 to 200 characters without control characters.",
  invalid_slug: "Use lowercase letters, digits and hyphens, starting and ending with a letter or digit, up to 63 characters. Some short names are reserved.",
  intent_used: "Your organization was already created with different details. Nothing was changed.",
  continue_failed: "We could not take you to your organization automatically. Nothing was lost.",
  retry_unavailable: "Your sign-up details are no longer available in this browser.",
  rate_limited: "Too many attempts. Wait a moment and try again.",
  unknown: "Your last step may or may not have completed. The current state is shown below; check it before trying again.",
  unavailable: "Sign-up is unavailable right now. Please try again later.",
};

export function noticeOf(notices: Readonly<Record<string, string>>, value: string | undefined): string | undefined {
  return value !== undefined && Object.hasOwn(notices, value) ? notices[value] : undefined;
}

// The field error of a details refusal (never a submitted value).
export function fieldErrorOf(status: string | undefined): { field: "name" | "slug"; message: string } | undefined {
  if (status === "invalid_name") return { field: "name", message: DETAILS_NOTICES.invalid_name };
  if (status === "invalid_slug") return { field: "slug", message: DETAILS_NOTICES.invalid_slug };
  if (status === "slug_taken") return { field: "slug", message: DETAILS_NOTICES.slug_taken };
  return undefined;
}

// The recovery view of an already provisioned signup on /signup/details (G062 bounded security correction to Plan r5
// §4.4, Codex Round 15 decision 2). Rendering it is a GET and therefore never mutates: no create, re-mint or other core
// write follows from it, whatever the restoration record holds. "Try again" (the create route's user-triggered POST with
// the CSRF token and the recorded payload; core's intent digest decides) is offered only when the browser's restoration
// record names the created address; a missing, stale or mismatched record offers nothing and calls nothing.
export function recoveryView(
  status: string | undefined,
  form: { name: string; slug: string } | undefined,
  organization: { slug: string } | null,
): { retryable: boolean; notice: string | undefined } {
  const retryable = form !== undefined && organization !== null && form.slug !== "" && form.slug === organization.slug;
  const notice = noticeOf(DETAILS_NOTICES, status) ?? (retryable ? undefined : DETAILS_NOTICES.retry_unavailable);
  return { retryable, notice };
}

// An outcome core may or may not have committed.
export function outcomeUnknown(error: SignupError): boolean {
  return error === "timeout" || error === "unreachable" || error === "malformed" || error === "signup_unavailable";
}
