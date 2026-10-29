import { brandingAnswer, readMutationFields, resetBranding } from "@/lib/branding";
import { deriveContext } from "@/lib/request-context";
import { badRequest, field, methodNotAllowed, readCookie, readForm, statusPage } from "@/lib/session";

// Reset this host's Organization branding to the platform defaults (label, colours and logo reference; uploaded
// objects are kept and stay listed). The existing form-route order, then core decides (G017 Plan r2 §5.3).
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const m = readMutationFields((name) => field(form.fields, name));
  if (m === undefined) return badRequest();
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
  const result = await resetBranding(deriveContext(request.headers), { origin: form.origin, ...m }, credential);
  return brandingAnswer(result, "reset", "The reset");
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
