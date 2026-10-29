import { isConnectionId, normalizeDisplayNameInput, renameConnection, returnTarget, statusLocation, strictFormField } from "@/lib/mcp-client";
import { deriveContext } from "@/lib/request-context";
import { AMBIGUOUS, badRequest, field, forbidden, methodNotAllowed, readCookie, readForm, seeOther, statusPage } from "@/lib/session";

// Rename (or, with an empty name, clear the name of) one own live MCP connection on this tenant host (G063 Plan r1
// §4.3, §5.4). The existing form-route order: method, form encoding, exact Origin and the 16 KiB bound (readForm), the
// fields, the session cookie, then core checks the session, the CSRF token, current membership and that the target is
// one of the caller's live connections here. The name field alone is read strictly from the raw body (request.clone()):
// malformed (invalid UTF-8, encoded surrogates) or outside the rule redirects with the closed invalid_name status and no
// core call; repeated or absent is refused. The name is a label only: it changes nothing else about the connection, and
// it is never placed in a redirect, a status page or a log.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const copy = request.clone();
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  const connectionId = field(form.fields, "connection_id");
  const target = returnTarget(form.fields);
  if (csrfToken === AMBIGUOUS || connectionId === AMBIGUOUS || connectionId === undefined || !isConnectionId(connectionId) || target === undefined) {
    return badRequest();
  }
  const name = await strictFormField(copy, "display_name");
  if (name.kind === "ambiguous" || name.kind === "absent") return badRequest();
  if (name.kind === "malformed") return seeOther(statusLocation(target, "invalid_name"));
  const normalized = normalizeDisplayNameInput(name.value);
  if (!normalized.ok) return seeOther(statusLocation(target, "invalid_name"));
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) {
    return statusPage(401, "Not signed in", "Sign in first, then rename the connection again. Nothing was changed.");
  }
  const result = await renameConnection(
    deriveContext(request.headers),
    { origin: form.origin, connectionId, csrfToken: csrfToken ?? "", displayName: normalized.value },
    credential,
  );
  if (result.ok) return seeOther(statusLocation(target, "renamed"));
  switch (result.error) {
    case "invalid_name":
      return seeOther(statusLocation(target, "invalid_name"));
    case "csrf_failed":
      return statusPage(403, "Form expired", "This form is no longer valid. Reload the connections page and try again. Nothing was changed.");
    case "unauthenticated":
      return statusPage(401, "Not signed in", "Sign in first, then rename the connection again. Nothing was changed.");
    case "not_found":
      return statusPage(404, "Connection not found", "That connection is not one of your active connections here (it may already be revoked). Nothing was changed.");
    case "not_available":
      return statusPage(403, "Not available", "Connections are not available for your account here. Nothing was changed.");
    case "origin_rejected":
      return forbidden();
    case "invalid_request":
    case "ambiguous_parameter":
    case "unsupported_media_type":
      return badRequest();
    default:
      return statusPage(503, "Name not saved", "The name could not be saved. Please try again shortly.");
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
