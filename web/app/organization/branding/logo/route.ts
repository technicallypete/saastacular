import {
  brandingAnswer,
  BRANDING_PAGE,
  isMultipartForm,
  MAX_UPLOAD_FORM_BYTES,
  parseLogoUpload,
  readBoundedBytes,
  setBrandingLogo,
} from "@/lib/branding";
import { deriveContext } from "@/lib/request-context";
import { badRequest, cookieProfile, forbidden, methodNotAllowed, ownOrigin, readCookie, seeOther, statusPage, type CookieProfile } from "@/lib/session";

// Upload a new logo for this host's Organization (G017 Plan r2 §5.3): the browser's multipart form is reduced to the raw
// image bytes (at most 512 KiB) and sent to core with the CSRF token and expected version as headers. The form-route
// order of the other forms, with the multipart encoding: exactly one multipart/form-data Content-Type (415), an Origin
// exactly equal to this request's own origin (403), then a bounded body (too large returns to the page). Core validates,
// normalizes and associates the image and alone decides.
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
  const body = await readBoundedBytes(request, MAX_UPLOAD_FORM_BYTES);
  if (body === undefined) return seeOther(`${BRANDING_PAGE}?status=payload_too_large`);
  const upload = await parseLogoUpload(contentType, body);
  if (!upload.ok) return upload.error === "too_large" ? seeOther(`${BRANDING_PAGE}?status=payload_too_large`) : badRequest();
  const credential = readCookie(request.headers.get("cookie"), "session", profile);
  if (credential === undefined) return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
  const result = await setBrandingLogo(deriveContext(request.headers), { origin, ...upload.fields }, upload.bytes, credential);
  return brandingAnswer(result, "logo_set", "The logo change");
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
