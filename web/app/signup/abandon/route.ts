import { deriveContext } from "@/lib/request-context";
import { abandonSignup, formExpired, refusedRoot, rootOfRootRequest, signupStatus } from "@/lib/signup-client";
import {
  badRequest,
  clearCookie,
  cookieProfile,
  field,
  forbidden,
  methodNotAllowed,
  ownOrigin,
  readCookie,
  readForm,
  seeOther,
  type CookieProfile,
} from "@/lib/session";

// G062 (Plan r5 §4.4; Codex Round 14 option A): /signup/abandon on the platform root.
// POST (the form layer, the root check, the signup credential's CSRF token): core revokes this browser's own signup
// credential and its current continuation generation; the signup credential and form restoration cookies are cleared
// and the browser returns to /signup. An Organization already created is unaffected.
// GET (no form, no write): the cleanup hop of a dead credential. A Server Component cannot set cookies, so the details
// page sends a browser whose credential core reports unauthenticated here; this route asks core's status read again and
// clears both cookies only when that read is unauthenticated (or no credential cookie is present). A live credential is
// sent back to /signup/details untouched, so a forged link cannot end a live signup.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  if (typeof csrfToken !== "string" || csrfToken === "") return badRequest();
  const ctx = deriveContext(request.headers);
  const root = await rootOfRootRequest(ctx, form.origin);
  if (!root.ok) return refusedRoot(root.error);
  const credential = readCookie(form.cookie, "signup", form.profile);
  if (credential === undefined) return cleared("/signup?status=expired", form.profile);
  const result = await abandonSignup(ctx, credential, { origin: form.origin, csrfToken });
  if (result.ok) return cleared("/signup?status=abandoned", form.profile);
  switch (result.error) {
    case "unauthenticated":
      return cleared("/signup?status=expired", form.profile);
    case "csrf_failed":
      return formExpired();
    case "origin_rejected":
      return forbidden();
    case "rate_limited":
      return seeOther("/signup/details?status=rate_limited");
    case "invalid_request":
      return badRequest();
    default:
      return seeOther("/signup/details?status=unavailable");
  }
}

export async function GET(request: Request): Promise<Response> {
  let profile: CookieProfile;
  try {
    profile = cookieProfile();
  } catch {
    return refusedRoot("unavailable");
  }
  const origin = ownOrigin(request.headers, profile);
  if (origin === undefined || new URL(request.url).search !== "") return refusedRoot("not_found");
  const ctx = deriveContext(request.headers);
  const root = await rootOfRootRequest(ctx, origin);
  if (!root.ok) return refusedRoot(root.error);
  const credential = readCookie(request.headers.get("cookie"), "signup", profile);
  if (credential === undefined) return cleared("/signup", profile);
  const view = await signupStatus(ctx, credential, origin);
  if (!view.ok && view.error === "unauthenticated") return cleared("/signup?status=expired", profile);
  return seeOther("/signup/details");
}

function cleared(location: string, profile: CookieProfile): Response {
  return seeOther(location, clearCookie("signup", profile), clearCookie("signup_form", profile));
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
