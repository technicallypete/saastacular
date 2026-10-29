import { callbackEnrollment } from "@/lib/enrollment-client";
import { lookupContext, relayCallback, relayLocation, relayState, resolveHost } from "@/lib/host";
import { deriveContext } from "@/lib/request-context";
import { authCallback } from "@/lib/signup-client";
import {
  clearCookie,
  cookieProfile,
  hasCookie,
  methodNotAllowed,
  ownOrigin,
  readCookie,
  readJoinCookie,
  seeOther,
  setCookie,
  statusPage,
  type CookieProfile,
} from "@/lib/session";

// Sign-in completion (Plan r5 §4.2, §7.1): the hosted page's redirect back. The raw query goes to core
// unparsed, with this browser's login binding and any live session; core validates, consumes the one-use
// transaction, exchanges the code and maps the identity. The browser sees only closed statuses.
//
// G016 B2 (Plan r4 §1.7 leg 2, §1.7.2): on the shared auth host this route is a relay with no cookie, no session and no
// transaction read. Core decides from the state value alone (GET /internal/v1/auth/relay) whether it names a routing
// custom host; only then the browser is sent 303 to that host's /auth/callback with the provider's raw query appended
// byte-identical (never re-serialized). Anything else is the fixed 404 page. Every answer is no-store and no-referrer.
//
// G062 (Plan r5 §1.2, §1.4, §4.2, §4.4): on the platform root this route completes the signup login (the signup
// credential, never a session; www.<domain> is 404), and on a tenant host it also accepts the pinned continuation
// callback's tenant_created outcome. Every completion is read through lib/signup-client.ts authCallback, which validates
// G058's answers exactly as lib/enrollment-client.ts completeCallback does plus the two G062 answers.
export const dynamic = "force-dynamic";

// Longer than any provider callback query; core bounds its own body at 16 KiB.
const MAX_QUERY = 8 * 1024;

const SIGN_IN_STATUS: Readonly<Record<string, string>> = {
  login_canceled: "canceled",
  transaction_invalid: "expired",
  transaction_expired: "expired",
  stale_authentication: "stale",
  identity_unconfigured: "unavailable",
  identity_unavailable: "unavailable",
  identity_namespace_mismatch: "unavailable",
  provider_timeout: "unavailable",
  timeout: "unavailable",
  unreachable: "unavailable",
};

export async function GET(request: Request): Promise<Response> {
  let profile: CookieProfile;
  try {
    profile = cookieProfile();
  } catch {
    return seeOther("/sign-in?status=unavailable");
  }
  const origin = ownOrigin(request.headers, profile);
  if (origin === undefined) return seeOther("/sign-in?status=failed");
  const rawQuery = new URL(request.url).search.slice(1);
  const oversized = rawQuery.length > MAX_QUERY;
  const ctx = deriveContext(request.headers);
  if (profile === "https") {
    // The host lookup is this request's second core call; core keeps one completion record per request ID, so the
    // lookup carries its own generated request ID (same trace and parent span), and the callback or relay keeps ctx's.
    // The lookup sends the origin only (never the query), so it precedes the size check: the host class decides whether
    // an answer may touch this host's cookies at all.
    const host = await resolveHost(lookupContext(ctx), origin);
    // The shared auth host never reads or writes a cookie, on any answer (an oversized query included).
    if (host.ok && host.value.kind === "auth") return oversized ? ambiguous() : relay(ctx, rawQuery);
    // An unknown host or a resolver failure is unclassified: no cookie is read or written (it may be the auth host).
    if (!host.ok) return seeOther(`/sign-in?status=${host.error === "unknown_host" ? "failed" : "unavailable"}`);
    // G062 (Plan r5 §1.2, §4.4): the platform root itself (a public entry whose own origin is core's signup_origin) is
    // the signup login's callback; www.<domain> is the fixed 404 page. Both are classified and are not the auth host, so
    // a present one-use join intent is still cleared; nothing on the root ever forwards it or a session.
    if (host.value.kind === "public_entry") {
      if (origin !== host.value.signupOrigin) return clearedJoin(relayRefused(), request, profile);
      if (oversized) return clearedJoin(ambiguous(), request, profile);
      const binding = readCookie(request.headers.get("cookie"), "login", profile);
      return clearedJoin(await signupCompleted(ctx, origin, rawQuery, binding, profile), request, profile);
    }
    // Only a routing tenant host continues to core: a classified non-routing host (reserved) is refused here and
    // nothing reaches the callback endpoint (core would refuse it as origin_rejected). It is classified and is not the
    // auth host, so a present one-use join intent is still cleared.
    if (host.value.kind !== "platform" && host.value.kind !== "custom") return clearedJoin(seeOther("/sign-in?status=failed"), request, profile);
  }
  // G058 (Plan r3 §1.6): from here the host is classified and is not the shared auth host. The one-use invitation intent
  // is forwarded in the callback body when exactly one well-formed join cookie is present, and the join cookie is cleared
  // on every outcome once the browser holds one, the early oversized-query refusal included (a duplicate or malformed
  // cookie reads as absent; core then refuses an invitation login as invitation_mismatch).
  if (oversized) return clearedJoin(ambiguous(), request, profile);
  const cookie = request.headers.get("cookie");
  const done = await completed(
    ctx,
    { origin, rawQuery, invitation: readJoinCookie(cookie, profile) },
    readCookie(cookie, "login", profile),
    readCookie(cookie, "session", profile),
    profile,
  );
  return clearedJoin(done, request, profile);
}

// Appends the join cookie's clearing when the browser holds one. Called only on a classified host that is not the shared
// auth host.
function clearedJoin(res: Response, request: Request, profile: CookieProfile): Response {
  if (hasCookie(request.headers.get("cookie"), "join", profile)) res.headers.append("Set-Cookie", clearCookie("join", profile));
  return res;
}

async function completed(
  ctx: ReturnType<typeof deriveContext>,
  input: Parameters<typeof authCallback>[1],
  binding: string | undefined,
  credential: string | undefined,
  profile: CookieProfile,
): Promise<Response> {
  const result = await authCallback(ctx, input, binding, credential);
  if (result.ok) {
    const done = result.value;
    if (done.kind === "admitted" || done.kind === "enrollment") return callbackEnrollment(done, profile);
    if (done.kind === "already_signed_in") return seeOther("/account?status=already_signed_in");
    // A signup credential is answered only for the root's own origin; a tenant host never stores one.
    if (done.kind === "signup") return seeOther("/sign-in?status=failed");
    // G062 (Plan r5 §1.4 step 4, §4.2): the pinned continuation callback signs the creator in to the new tenant and lands
    // on its entry with the closed tenant_created notice (core keeps return_to "/").
    if (done.outcome === "tenant_created") {
      return seeOther("/?status=tenant_created", setCookie("session", done.session.value, done.session.maxAge, profile));
    }
    // An enrollment login that the ordinary sign-in completed (a current member) returns to the tenant entry.
    const returnTo = done.returnTo === "/join/status" ? "/" : done.returnTo;
    return seeOther(`${returnTo}?status=${done.outcome}`, setCookie("session", done.session.value, done.session.maxAge, profile));
  }
  if (result.error === "identity_conflict") return seeOther("/account?status=identity_conflict");
  if (result.error === "ambiguous_parameter") return ambiguous();
  if (result.error === "organization_not_member") return seeOther("/sign-in?status=not_member");
  if (result.error === "host_changed") return seeOther("/sign-in?status=host_changed");
  if (result.error === "enrollment_closed") return seeOther("/join?status=closed");
  if (result.error === "invitation_mismatch") return seeOther("/join?status=invitation_mismatch");
  return seeOther(`/sign-in?status=${SIGN_IN_STATUS[result.error] ?? "failed"}`);
}

// The root signup statuses of a refused signup callback (Plan r5 §4.4: failures 303 /signup?status=<closed>).
const SIGNUP_STATUS: Readonly<Record<string, string>> = {
  login_canceled: "canceled",
  transaction_invalid: "expired",
  transaction_expired: "expired",
  user_disabled: "user_disabled",
  rate_limited: "rate_limited",
  identity_unconfigured: "unavailable",
  identity_unavailable: "unavailable",
  identity_namespace_mismatch: "unavailable",
  provider_timeout: "unavailable",
  timeout: "unavailable",
  unreachable: "unavailable",
};

// G062 (Plan r5 §1.2, §4.2, §4.4): the root signup login's completion. Core answers 202 with the signup credential only
// (never a session), which becomes the host-only root signup cookie, and the browser continues to /signup/details. Any
// other answer creates nothing here: a refusal goes to /signup with its closed status, and an answer that is not the
// signup credential (a root never issues a session) is the generic failure without any cookie.
async function signupCompleted(
  ctx: ReturnType<typeof deriveContext>,
  origin: string,
  rawQuery: string,
  binding: string | undefined,
  profile: CookieProfile,
): Promise<Response> {
  const result = await authCallback(ctx, { origin, rawQuery }, binding);
  if (result.ok) {
    if (result.value.kind !== "signup") return seeOther("/signup?status=failed");
    const { value, maxAge } = result.value.credential;
    return seeOther("/signup/details", setCookie("signup", value, maxAge, profile));
  }
  if (result.error === "ambiguous_parameter") return ambiguous();
  return seeOther(`/signup?status=${SIGNUP_STATUS[result.error] ?? "failed"}`);
}

// The relay leg: nothing is read from or set on this host's cookies, and nothing beyond core's closed decision is used.
async function relay(ctx: ReturnType<typeof deriveContext>, rawQuery: string): Promise<Response> {
  const state = relayState(rawQuery);
  if (state === undefined) return relayRefused();
  const decision = await relayCallback(ctx, state);
  if (!decision.ok) {
    return decision.error === "relay_refused" ? relayRefused() : statusPage(503, "Unavailable", "Sign-in is unavailable right now. Please try again later.");
  }
  const location = relayLocation(decision.callback, rawQuery);
  return location === undefined ? relayRefused() : seeOther(location);
}

function relayRefused(): Response {
  return statusPage(404, "Not found", "There is nothing at this address.");
}

function ambiguous(): Response {
  return statusPage(400, "Sign-in not completed", "The sign-in response could not be read. Nothing was changed. Please start again.");
}

export function POST(): Response {
  return methodNotAllowed("GET");
}

export function PUT(): Response {
  return methodNotAllowed("GET");
}

export function PATCH(): Response {
  return methodNotAllowed("GET");
}

export function DELETE(): Response {
  return methodNotAllowed("GET");
}

export function OPTIONS(): Response {
  return methodNotAllowed("GET");
}
