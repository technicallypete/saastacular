import { logout } from "@/lib/identity-client";
import { deriveContext } from "@/lib/request-context";
import { AMBIGUOUS, badRequest, clearCookie, field, forbidden, methodNotAllowed, readCookie, readForm, seeOther, statusPage } from "@/lib/session";

// Sign out of this application in this browser (Plan r5 §4.4). Core commits the revocation before success is
// shown; without a live session the request is idempotent. A commit failure never claims success.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  if (csrfToken === AMBIGUOUS) return badRequest();
  const cleared = clearCookie("session", form.profile);
  const result = await logout(
    deriveContext(request.headers),
    { scope: "current", csrfToken: csrfToken ?? "", origin: form.origin },
    readCookie(form.cookie, "session", form.profile),
  );
  if (result.ok) return seeOther("/sign-in?status=signed_out", cleared);
  switch (result.error) {
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
