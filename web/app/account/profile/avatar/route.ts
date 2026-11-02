import { isMultipartForm, readBoundedBytes } from "@/lib/branding";
import { MAX_AVATAR_FORM_BYTES, parseAvatarUpload, profileAnswer, profileBack } from "@/lib/profile";
import { setAvatar } from "@/lib/profile-client";
import { deriveContext } from "@/lib/request-context";
import { badRequest, cookieProfile, forbidden, methodNotAllowed, ownOrigin, readCookie, statusPage, type CookieProfile } from "@/lib/session";

// G074 (Plan r3.1 §3.7, §3.8 Set): upload a new avatar for the signed-in user. The browser's multipart form is reduced to
// the raw image bytes (at most 512 KiB) and sent to core with the CSRF token, expected version and origin. The G017 logo
// form-route order: exactly one multipart/form-data Content-Type (415), an Origin exactly equal to this request's own
// origin (403), then a bounded body (too large returns to the page). Core validates, normalizes and stores the image and
// alone decides; a repairable HOLD returns to the page as avatar_hold (never retried here).
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  let profile: CookieProfile;
  try {
    profile = cookieProfile();
  } catch {
    return statusPage(503, "Unavailable", "This service is not available right now. Nothing was changed.");
  }
  const contentType = request.headers.get("content-type");
  if (!isMultipartForm(contentType)) {
    return statusPage(415, "Unsupported request", "Only this site's own forms can submit here. Nothing was changed.");
  }
  const origin = ownOrigin(request.headers, profile);
  if (origin === undefined || request.headers.get("origin") !== origin) return forbidden();
  const body = await readBoundedBytes(request, MAX_AVATAR_FORM_BYTES);
  if (body === undefined) return profileBack("payload_too_large");
  const upload = await parseAvatarUpload(contentType, body);
  if (!upload.ok) return upload.error === "too_large" ? profileBack("payload_too_large") : badRequest();
  const credential = readCookie(request.headers.get("cookie"), "session", profile);
  if (credential === undefined) return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
  const result = await setAvatar(deriveContext(request.headers), credential, { origin, ...upload.fields }, upload.bytes);
  return profileAnswer(result, "avatar_set");
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
