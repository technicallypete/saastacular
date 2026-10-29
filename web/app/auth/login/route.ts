import { login } from "@/lib/identity-client";
import { deriveContext } from "@/lib/request-context";
import { AMBIGUOUS, badRequest, field, forbidden, methodNotAllowed, readCookie, readForm, seeOther, setCookie } from "@/lib/session";

// Sign-in start (Plan r5 §4.2, §7.1): a form POST from this site's own page. Core starts a one-use login
// transaction bound to this browser's login binding; the browser goes to the hosted authorization page.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const returnTo = field(form.fields, "return_to");
  if (returnTo === AMBIGUOUS) return badRequest();
  const result = await login(
    deriveContext(request.headers),
    { origin: form.origin, returnTo: returnTo === "/account" ? "/account" : undefined },
    readCookie(form.cookie, "session", form.profile),
    readCookie(form.cookie, "login", form.profile),
  );
  if (result.ok) {
    const { authorizationUrl, binding } = result.value;
    return seeOther(authorizationUrl, setCookie("login", binding.value, binding.maxAge, form.profile));
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
      // Unconfigured, unavailable, namespace mismatch, or core unreachable or out of contract.
      return seeOther("/sign-in?status=unavailable");
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
