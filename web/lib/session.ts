import "server-only";

// The browser side of an application session (Plan r5 §4.2–4.4, §5, §7.1; C7 addendum §2). This module alone handles
// the session, login-binding and logout-reference cookies: their names and fixed attributes, and reading exactly one
// well-formed value from the request's Cookie header. It also holds the browser-form checks (method, form encoding, exact
// Origin) and the minimal no-store responses of the form layer. Values are opaque here; core issues, hashes,
// resolves and revokes them. It reads the raw Cookie header rather than next/headers' cookie store, which keeps
// only one of several same-named cookies: a repeated name is ambiguous and yields no value.

export type CookieProfile = "https" | "loopback-http";
// G058 (Plan r3 §1.6 cookie contract): enrollment is the applicant's enrollment credential (never a session; its own
// name, never mixed with the session cookie) and join the one-use invitation intent `<invitation id>.<token>`, written
// and read only through setJoinCookie/readJoinCookie.
// G062 (Plan r5 §1.4, §4.4; Codex Round 14 option A): signup is the root signup credential (host-only on the platform root,
// never a session), challenge the tenant continuation challenge C (host-only on the new tenant host, 120 s) and
// signup_form the root form-restoration record of the exact submitted Organization name and slug (written and read only
// through setSignupFormCookie/readSignupFormCookie; never authentication, never in a URL or a log).
export type CookieKind = "session" | "login" | "logout" | "enrollment" | "join" | "signup" | "challenge" | "signup_form";

type Env = Readonly<Record<string, string | undefined>>;

const NAMES: Readonly<Record<CookieProfile, Readonly<Record<CookieKind, string>>>> = {
  https: {
    session: "__Host-sid",
    login: "__Host-login",
    logout: "__Host-logout",
    enrollment: "__Host-enroll",
    join: "__Host-join",
    signup: "__Host-signup",
    challenge: "__Host-sct",
    signup_form: "__Host-signup-form",
  },
  "loopback-http": {
    session: "sid",
    login: "login",
    logout: "logout",
    enrollment: "enroll",
    join: "join",
    signup: "signup",
    challenge: "sct",
    signup_form: "signup-form",
  },
};

// The logout-reference cookie never outlives core's ten-minute reference: it only names the server-recorded
// one-use outcome of a provider sign-out (never identity, session authority or a provider result).
export const LOGOUT_COOKIE_MAX_AGE = 600;
// G058: the enrollment credential's absolute lifetime (core EnrollmentSessionLifetime, 30 min) bounds its cookie, and the
// invitation intent lives no longer than the login transaction it accompanies (JoinCookieLifetime, 10 min).
export const ENROLLMENT_COOKIE_MAX_AGE = 1800;
export const JOIN_COOKIE_MAX_AGE = 600;
// <invitation uuid (36)>.<token (43)>: 80 characters.
const JOIN = /^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\.([A-Za-z0-9_-]{43})$/;
// G062: the root signup credential's absolute lifetime (the signup row's 30 minutes) bounds its cookie and the form
// restoration cookie; the tenant challenge lives exactly as long as its continuation generation (120 s).
export const SIGNUP_COOKIE_MAX_AGE = 1800;
export const CHALLENGE_COOKIE_MAX_AGE = 120;
// The form restoration record: "v1.<base64url of the UTF-8 name>.<slug>" with the name 1-800 bytes (the 200-code-point
// name bound in at most four bytes each; whitespace a browser submits around it is kept exactly) and the slug empty or one
// host label (lower-case letters, digits and inner hyphens, at most 63). At most 1135 characters.
export const MAX_FORM_NAME_BYTES = 800;
const FORM_SLUG = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)?$/;
const FORM_RECORD = /^v1\.([A-Za-z0-9_-]{2,1067})\.([a-z0-9-]{0,63})$/;

// Core's canonical encoding of a 256-bit value: 43 unpadded base64url characters.
const SECRET = /^[A-Za-z0-9_-]{43}$/;
// The longest configurable absolute session lifetime (24 h).
const MAX_COOKIE_AGE = 86_400;
const LABEL = "[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?";
const HOST = new RegExp(`^${LABEL}(?:\\.${LABEL})*(?::[0-9]{1,5})?$`);
const LOOPBACK_HOSTS: ReadonlySet<string> = new Set(["127.0.0.1", "localhost"]);
const MAX_FORM_BYTES = 16 * 1024;
const FORM_TYPE = "application/x-www-form-urlencoded";

// G016 (Plan r4 §1.7.2): no-referrer as well, so neither a fixed page rendered at a code/state-bearing callback URL nor
// a redirected navigation sends that URL as Referer.
const NO_STORE: Readonly<Record<string, string>> = {
  "Cache-Control": "no-store",
  "X-Content-Type-Options": "nosniff",
  "Referrer-Policy": "no-referrer",
};

// IDENTITY_COOKIE_PROFILE: "https" (__Host- cookies, Secure, https origins) or "loopback-http" (plain-HTTP
// loopback development). Unset means https in production and loopback-http elsewhere; production refuses
// loopback-http. Any other value is a configuration error.
export function cookieProfile(env: Env = process.env): CookieProfile {
  const production = env.APP_ENV === "production";
  const value = (env.IDENTITY_COOKIE_PROFILE ?? "").trim();
  if (value === "") return production ? "https" : "loopback-http";
  if (value === "https") return "https";
  if (value === "loopback-http" && !production) return "loopback-http";
  throw new Error("IDENTITY_COOKIE_PROFILE is invalid");
}

export function cookieName(kind: CookieKind, profile: CookieProfile): string {
  return NAMES[profile][kind];
}

function serialize(name: string, value: string, maxAge: number, profile: CookieProfile): string {
  const parts = [`${name}=${value}`, "Path=/", `Max-Age=${maxAge}`, "HttpOnly", "SameSite=Lax"];
  if (profile === "https") parts.push("Secure");
  return parts.join("; ");
}

// A Set-Cookie value with the fixed attributes: HttpOnly, Path=/, no Domain, SameSite=Lax, Secure for https.
const LIMITS: Partial<Readonly<Record<CookieKind, number>>> = {
  logout: LOGOUT_COOKIE_MAX_AGE,
  enrollment: ENROLLMENT_COOKIE_MAX_AGE,
  signup: SIGNUP_COOKIE_MAX_AGE,
  challenge: CHALLENGE_COOKIE_MAX_AGE,
};

export function setCookie(kind: CookieKind, value: string, maxAge: number, profile: CookieProfile): string {
  if (kind === "join" || kind === "signup_form") throw new TypeError(`the ${kind} cookie has its own serializer`);
  if (!SECRET.test(value)) throw new TypeError("cookie value is not a credential");
  const limit = LIMITS[kind] ?? MAX_COOKIE_AGE;
  if (!Number.isInteger(maxAge) || maxAge < 1 || maxAge > limit) throw new RangeError("cookie lifetime out of range");
  return serialize(cookieName(kind, profile), value, maxAge, profile);
}

export function clearCookie(kind: CookieKind, profile: CookieProfile): string {
  return serialize(cookieName(kind, profile), "", 0, profile);
}

// Every value of the named cookie in the header, in order.
function valuesOf(header: string, name: string): string[] {
  const values: string[] = [];
  for (const pair of header.split(";")) {
    const at = pair.indexOf("=");
    if (at > 0 && pair.slice(0, at).trim() === name) values.push(pair.slice(at + 1).trim());
  }
  return values;
}

// Exactly one well-formed value of the cookie, or undefined when absent, repeated or malformed. The join cookie never
// has a credential's shape, so it never reads here (readJoinCookie reads it).
export function readCookie(header: string | null, kind: CookieKind, profile: CookieProfile): string | undefined {
  if (header === null) return undefined;
  const values = valuesOf(header, cookieName(kind, profile));
  return values.length === 1 && SECRET.test(values[0]) ? values[0] : undefined;
}

// G058 (Plan r3 §1.6): the one-use invitation intent set by POST /join/start in the same answer as the login binding.
// Its value is exactly `<invitation id>.<token>`; anything else is refused. Path=/ (the only legal __Host- path), no
// Domain, HttpOnly, SameSite=Lax, Secure on https; 600 s.
export function setJoinCookie(invitationId: string, token: string, profile: CookieProfile): string {
  const value = `${invitationId}.${token}`;
  if (!JOIN.test(value)) throw new TypeError("join cookie value is not an invitation intent");
  return serialize(cookieName("join", profile), value, JOIN_COOKIE_MAX_AGE, profile);
}

// The invitation intent of exactly one well-formed join cookie; a duplicate or malformed cookie reads as absent.
export function readJoinCookie(header: string | null, profile: CookieProfile): { invitationId: string; token: string } | undefined {
  if (header === null) return undefined;
  const values = valuesOf(header, cookieName("join", profile));
  const m = values.length === 1 ? JOIN.exec(values[0]) : null;
  return m === null ? undefined : { invitationId: m[1], token: m[2] };
}

// G062 (Codex Round 14 option A): the root signup form's restoration record, so /signup/details can restore the exact
// submitted Organization name and slug and its "Try again" can repeat the identical create payload. It is form data only:
// never authentication or authority (the create route still requires the signup credential and its CSRF token, and
// core's intent digest decides; a changed value is refused as intent_used), never in a URL or a log, and never holding a
// credential, token, continuation secret or identity datum. Host-only on the root (Path=/, no Domain), HttpOnly,
// SameSite=Lax, Secure on https, at most the signup credential's 30 minutes. Undefined when a value cannot be recorded
// within the bounds (the name 1-800 UTF-8 bytes without NUL; the slug empty or one host label).
export type SignupForm = { name: string; slug: string };

function formNameOk(name: string): boolean {
  const bytes = Buffer.byteLength(name, "utf8");
  return name.length > 0 && bytes <= MAX_FORM_NAME_BYTES && !name.includes("\u0000") && name.isWellFormed();
}

export function setSignupFormCookie(form: SignupForm, profile: CookieProfile): string | undefined {
  if (!formNameOk(form.name) || !FORM_SLUG.test(form.slug)) return undefined;
  const value = `v1.${Buffer.from(form.name, "utf8").toString("base64url")}.${form.slug}`;
  return serialize(cookieName("signup_form", profile), value, SIGNUP_COOKIE_MAX_AGE, profile);
}

// The record of exactly one well-formed signup form cookie: canonical base64url of valid UTF-8 within the name bound and
// a slug that is empty or one host label. A duplicate, oversized or malformed cookie reads as absent.
export function readSignupFormCookie(header: string | null, profile: CookieProfile): SignupForm | undefined {
  if (header === null) return undefined;
  const values = valuesOf(header, cookieName("signup_form", profile));
  const m = values.length === 1 ? FORM_RECORD.exec(values[0]) : null;
  if (m === null || !FORM_SLUG.test(m[2])) return undefined;
  const bytes = Buffer.from(m[1], "base64url");
  if (bytes.toString("base64url") !== m[1]) return undefined;
  let name: string;
  try {
    name = new TextDecoder("utf-8", { fatal: true, ignoreBOM: true }).decode(bytes);
  } catch {
    return undefined;
  }
  return formNameOk(name) ? { name, slug: m[2] } : undefined;
}

// Whether the Cookie header names the cookie at all, whatever its value or count (presence only, never a value).
export function hasCookie(header: string | null, kind: CookieKind, profile: CookieProfile): boolean {
  if (header === null) return false;
  const name = cookieName(kind, profile);
  return header.split(";").some((pair) => {
    const at = pair.indexOf("=");
    return at > 0 && pair.slice(0, at).trim() === name;
  });
}

// This request's own origin from its Host header and the profile's scheme. Forwarding headers are never read.
// loopback-http serves only loopback hosts. Undefined when the host is absent or malformed.
export function ownOrigin(headers: Headers, profile: CookieProfile): string | undefined {
  const host = headers.get("host")?.toLowerCase();
  if (host === undefined || host.length > 261 || !HOST.test(host)) return undefined;
  if (profile === "loopback-http") {
    return LOOPBACK_HOSTS.has(host.replace(/:[0-9]+$/, "")) ? `http://${host}` : undefined;
  }
  return `https://${host}`;
}

export function seeOther(location: string, ...cookies: string[]): Response {
  const headers = new Headers({ ...NO_STORE, Location: location });
  for (const cookie of cookies) headers.append("Set-Cookie", cookie);
  return new Response(null, { status: 303, headers });
}

function escapeHtml(text: string): string {
  return text.replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`);
}

// A minimal, self-contained no-store HTML page for a refused or failed form request. The text is fixed by the
// caller; nothing from the request is reflected.
export function statusPage(status: number, title: string, message: string, ...cookies: string[]): Response {
  const body =
    `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">` +
    `<title>${escapeHtml(title)}</title></head><body><main><h1>${escapeHtml(title)}</h1><p>${escapeHtml(message)}</p>` +
    `<p><a href="/account">Account</a> · <a href="/sign-in">Sign in</a></p></main></body></html>`;
  const headers = new Headers({ ...NO_STORE, "Content-Type": "text/html; charset=utf-8" });
  for (const cookie of cookies) headers.append("Set-Cookie", cookie);
  return new Response(body, { status, headers });
}

export function methodNotAllowed(allow: "GET" | "POST"): Response {
  const response = statusPage(405, "Method not allowed", "This address does not accept that kind of request.");
  response.headers.set("Allow", allow);
  return response;
}

export function forbidden(): Response {
  return statusPage(403, "Request refused", "This request did not come from this site's own page. Nothing was changed.");
}

export function badRequest(): Response {
  return statusPage(400, "Request not understood", "This request could not be read. Nothing was changed.");
}

// Exactly one Content-Type of the form encoding; repeated headers arrive comma-joined and are refused.
function isFormEncoded(value: string | null): boolean {
  if (value === null || value.includes(",")) return false;
  const [type, ...params] = value.split(";").map((part) => part.trim().toLowerCase());
  return type === FORM_TYPE && params.every((param) => param === "charset=utf-8");
}

// G018 (Plan r5 §1.6): maxBytes is the call's form budget; its absence keeps the 16 KiB default of every existing form.
async function readBounded(request: Request, maxBytes: number = MAX_FORM_BYTES): Promise<string | undefined> {
  if (request.body === null) return "";
  const reader = request.body.getReader();
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
  return new TextDecoder("utf-8", { fatal: false }).decode(Buffer.concat(chunks));
}

export type FormRequest =
  | { ok: true; profile: CookieProfile; origin: string; fields: URLSearchParams; cookie: string | null }
  | { ok: false; response: Response };

// The browser-form layer checks, in order (Plan r5 §4.4): exactly one application/x-www-form-urlencoded
// Content-Type (415), then an Origin header exactly equal to this request's own origin (403), then a body of at
// most 16 KiB (400). The route handler accepts only POST; other methods answer 405 before this runs. G018: options.maxBytes
// replaces the 16 KiB body budget for that call only (the two Business Profile forms; Plan r5 §1.6).
export async function readForm(request: Request, options?: { maxBytes?: number }): Promise<FormRequest> {
  let profile: CookieProfile;
  try {
    profile = cookieProfile();
  } catch {
    return { ok: false, response: statusPage(503, "Unavailable", "This service is not available right now. Nothing was changed.") };
  }
  if (!isFormEncoded(request.headers.get("content-type"))) {
    return { ok: false, response: statusPage(415, "Unsupported request", "Only this site's own forms can submit here. Nothing was changed.") };
  }
  const origin = ownOrigin(request.headers, profile);
  if (origin === undefined || request.headers.get("origin") !== origin) return { ok: false, response: forbidden() };
  const body = await readBounded(request, options?.maxBytes);
  if (body === undefined) return { ok: false, response: badRequest() };
  return { ok: true, profile, origin, fields: new URLSearchParams(body), cookie: request.headers.get("cookie") };
}

export const AMBIGUOUS = Symbol("ambiguous");

// One form field: its value, undefined when absent, or AMBIGUOUS when repeated.
export function field(fields: URLSearchParams, name: string): string | undefined | typeof AMBIGUOUS {
  const values = fields.getAll(name);
  if (values.length > 1) return AMBIGUOUS;
  return values[0];
}

export function jsonResponse(status: number, body: Readonly<Record<string, unknown>>): Response {
  return Response.json(body, { status, headers: NO_STORE });
}
