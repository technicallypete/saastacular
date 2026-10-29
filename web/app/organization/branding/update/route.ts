import { brandingAnswer, BRANDING_PAGE, readBrandingInput, readMutationFields, updateBranding } from "@/lib/branding";
import { deriveContext } from "@/lib/request-context";
import { badRequest, field, methodNotAllowed, readCookie, readForm, seeOther, statusPage } from "@/lib/session";

// Update this host's Organization branding: the label and the per-appearance colour overrides (G017 Plan r2 §5.3–5.4).
// The existing form-route order (method, form encoding, exact Origin, bounded body), then core checks the session, the
// CSRF token, the manage_branding capability, the expected version and the effective contrast.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const m = readMutationFields((name) => field(form.fields, name));
  const input = readBrandingInput(form.fields);
  if (m === undefined || (!input.ok && input.field === undefined)) return badRequest();
  if (!input.ok) return seeOther(`${BRANDING_PAGE}?status=invalid_input&field=${input.field}`);
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
  const result = await updateBranding(deriveContext(request.headers), { origin: form.origin, ...m }, input.value, credential);
  return brandingAnswer(result, "updated", "The branding change");
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
