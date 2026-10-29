import { login } from "@/lib/identity-client";
import { deriveContext } from "@/lib/request-context";
import { refusedRoot, rootOfRootRequest } from "@/lib/signup-client";
import { badRequest, clearCookie, forbidden, methodNotAllowed, readCookie, readForm, seeOther, setCookie } from "@/lib/session";

// G062 (Plan r5 §1.2, §4.4): POST /signup/start on the platform root. The existing form-route order (method, form
// encoding, exact Origin, 16 KiB), then the root check (the fixed 404 anywhere but the root itself), then core starts the
// platform signup login (intent "signup", exact root origin only) bound to this browser's login binding. The answer is
// the 303 to the hosted page with the login binding; any previous signup credential and form restoration cookies are
// cleared (a new signup login issues a new credential). No session is read or forwarded on the root.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const ctx = deriveContext(request.headers);
  const root = await rootOfRootRequest(ctx, form.origin);
  if (!root.ok) return refusedRoot(root.error);
  const result = await login(ctx, { origin: form.origin, intent: "signup" }, undefined, readCookie(form.cookie, "login", form.profile));
  if (result.ok) {
    const { authorizationUrl, binding } = result.value;
    return seeOther(
      authorizationUrl,
      setCookie("login", binding.value, binding.maxAge, form.profile),
      clearCookie("signup", form.profile),
      clearCookie("signup_form", form.profile),
    );
  }
  switch (result.error) {
    case "rate_limited":
      return seeOther("/signup?status=rate_limited");
    case "origin_rejected":
      return forbidden();
    case "invalid_request":
    case "ambiguous_parameter":
    case "unsupported_media_type":
      return badRequest();
    default:
      return seeOther("/signup?status=unavailable");
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
