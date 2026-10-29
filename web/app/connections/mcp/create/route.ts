import { createConnection, normalizeDisplayNameInput, returnTarget, statusLocation, strictFormField } from "@/lib/mcp-client";
import { deriveContext } from "@/lib/request-context";
import { AMBIGUOUS, badRequest, field, forbidden, methodNotAllowed, readCookie, readForm, seeOther, statusPage } from "@/lib/session";

// Create a pending MCP connection of the signed-in member on this tenant host (G014 Plan r8 §3.4). The existing
// form-route order: method, form encoding, exact Origin (readForm), then core checks the session, the CSRF token, the
// trusted host registry and membership. G063 (Plan r1 §5.4): an optional friendly name, read strictly from the raw body
// (request.clone(), so readForm and every other field keep their existing decoding): a malformed or ambiguous name never
// reaches core, a name outside the rule redirects with the closed invalid_name status, and a name that normalizes to
// empty is "no name" (the unnamed-create request is unchanged). The closed return_to selects the page to return to
// (omitted: the compatibility page). The raw name is never placed in a redirect, a status page or a log.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const copy = request.clone();
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  const target = returnTarget(form.fields);
  if (csrfToken === AMBIGUOUS || target === undefined) return badRequest();
  const name = await strictFormField(copy, "display_name");
  if (name.kind === "ambiguous") return badRequest();
  if (name.kind === "malformed") return seeOther(statusLocation(target, "invalid_name"));
  let displayName: string | undefined;
  if (name.kind === "value") {
    const normalized = normalizeDisplayNameInput(name.value);
    if (!normalized.ok) return seeOther(statusLocation(target, "invalid_name"));
    if (normalized.value !== "") displayName = normalized.value;
  }
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) {
    return statusPage(401, "Not signed in", "Sign in first, then create the connection again. Nothing was changed.");
  }
  const result = await createConnection(deriveContext(request.headers), { origin: form.origin, csrfToken: csrfToken ?? "", displayName }, credential);
  if (result.ok) return seeOther(statusLocation(target, "created"));
  switch (result.error) {
    case "invalid_name":
      return seeOther(statusLocation(target, "invalid_name"));
    case "csrf_failed":
      return statusPage(403, "Form expired", "This form is no longer valid. Reload the connections page and try again. Nothing was changed.");
    case "unauthenticated":
      return statusPage(401, "Not signed in", "Sign in first, then create the connection again. Nothing was changed.");
    case "not_available":
      return statusPage(403, "Not available", "Connections are not available for your account here. Nothing was changed.");
    case "origin_rejected":
      return forbidden();
    case "invalid_request":
    case "ambiguous_parameter":
    case "unsupported_media_type":
      return badRequest();
    default:
      return statusPage(503, "Connection not created", "The connection could not be created. Please try again shortly.");
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
