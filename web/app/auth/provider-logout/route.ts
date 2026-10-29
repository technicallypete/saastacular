import { logout } from "@/lib/identity-client";
import { deriveContext } from "@/lib/request-context";
import {
  AMBIGUOUS,
  LOGOUT_COOKIE_MAX_AGE,
  badRequest,
  clearCookie,
  field,
  forbidden,
  methodNotAllowed,
  readCookie,
  readForm,
  seeOther,
  setCookie,
  statusPage,
} from "@/lib/session";

// Sign out and revoke the WorkOS session for this sign-in (Plan r5 §3.5, §4.4; C7 addendum §2). Core commits the
// local revocation, revokes exactly this provider session and records the outcome under a one-use reference, which
// this route keeps in the host-only logout cookie (never in a URL). Only a confirmed revocation continues to the
// provider's hosted logout, which returns to this origin's registered sign-in page; otherwise the browser goes
// straight to /auth/logout-return, which shows the recorded outcome once and clears the cookie.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  if (csrfToken === AMBIGUOUS) return badRequest();
  const cleared = clearCookie("session", form.profile);
  const result = await logout(
    deriveContext(request.headers),
    { scope: "provider", csrfToken: csrfToken ?? "", origin: form.origin },
    readCookie(form.cookie, "session", form.profile),
  );
  if (result.ok) {
    const done = result.value;
    switch (done.outcome) {
      case "provider_logout_redirected":
        return seeOther(done.providerLogoutUrl, cleared, setCookie("logout", done.logoutRef, LOGOUT_COOKIE_MAX_AGE, form.profile));
      case "provider_session_revoked":
      case "provider_revoke_failed":
        return seeOther("/auth/logout-return", cleared, setCookie("logout", done.logoutRef, LOGOUT_COOKIE_MAX_AGE, form.profile));
      default:
        return seeOther("/sign-in?status=signed_out", cleared);
    }
  }
  switch (result.error) {
    case "unauthenticated":
      return statusPage(401, "Not signed in", "You are not signed in, so there was nothing to sign out of. Nothing was changed.", cleared);
    case "csrf_failed":
      return statusPage(403, "Form expired", "This form is no longer valid. Reload the account page and try again. Nothing was changed.");
    case "origin_rejected":
      return forbidden();
    case "invalid_request":
    case "ambiguous_parameter":
    case "unsupported_media_type":
      return badRequest();
    default:
      return statusPage(503, "Sign-out failed", "Sign-out could not be confirmed. Please try again shortly.", cleared);
  }
}

export function GET(): Response {
  return methodNotAllowed("POST");
}

export function PUT(): Response {
  return methodNotAllowed("POST");
}

export function PATCH(): Response {
  return methodNotAllowed("POST");
}

export function DELETE(): Response {
  return methodNotAllowed("POST");
}

export function OPTIONS(): Response {
  return methodNotAllowed("POST");
}
