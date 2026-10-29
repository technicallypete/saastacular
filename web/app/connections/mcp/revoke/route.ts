import { isConnectionId, returnTarget, revokeConnection, statusLocation } from "@/lib/mcp-client";
import { deriveContext } from "@/lib/request-context";
import { AMBIGUOUS, badRequest, field, forbidden, methodNotAllowed, readCookie, readForm, seeOther, statusPage } from "@/lib/session";

// Revoke one own MCP connection on this tenant host (G014 Plan r8 §4.6 M4): connection-local and terminal; core checks
// the session, the CSRF token and exact ownership (no membership needed). The existing form-route order applies. G063:
// the closed return_to selects the page to return to (omitted: the compatibility page; any other value or a repeat: 400).
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  const connectionId = field(form.fields, "connection_id");
  const target = returnTarget(form.fields);
  if (csrfToken === AMBIGUOUS || connectionId === AMBIGUOUS || connectionId === undefined || !isConnectionId(connectionId) || target === undefined) {
    return badRequest();
  }
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) {
    return statusPage(401, "Not signed in", "Sign in first, then revoke the connection again. Nothing was changed.");
  }
  const result = await revokeConnection(
    deriveContext(request.headers),
    { origin: form.origin, connectionId, csrfToken: csrfToken ?? "" },
    credential,
  );
  if (result.ok) return seeOther(statusLocation(target, "revoked"));
  switch (result.error) {
    case "csrf_failed":
      return statusPage(403, "Form expired", "This form is no longer valid. Reload the connections page and try again. Nothing was changed.");
    case "unauthenticated":
      return statusPage(401, "Not signed in", "Sign in first, then revoke the connection again. Nothing was changed.");
    case "not_found":
      return statusPage(404, "Connection not found", "That connection is not one of your active connections here (it may already be revoked). Nothing was changed.");
    case "origin_rejected":
      return forbidden();
    case "invalid_request":
    case "ambiguous_parameter":
    case "unsupported_media_type":
      return badRequest();
    default:
      return statusPage(503, "Revocation not confirmed", "The revocation could not be confirmed. Please try again shortly.");
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
