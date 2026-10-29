import { brandingAnswer, BRANDING_PAGE, deleteBrandingUpload, readMutationFields } from "@/lib/branding";
import { deriveContext } from "@/lib/request-context";
import { badRequest, field, methodNotAllowed, readCookie, readForm, seeOther, statusPage } from "@/lib/session";

// Delete one unreferenced logo upload of this host's Organization (G017 D-3): core deletes only an object its current
// view lists as an upload, at the caller's version, and answers the truthful outcome (deleted or deletion pending).
export const dynamic = "force-dynamic";

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const m = readMutationFields((name) => field(form.fields, name));
  const objectId = field(form.fields, "object_id");
  if (m === undefined || typeof objectId !== "string" || !UUID.test(objectId)) return badRequest();
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
  const result = await deleteBrandingUpload(deriveContext(request.headers), { origin: form.origin, ...m }, objectId, credential);
  if (result.ok) return seeOther(`${BRANDING_PAGE}?status=${result.value === "deleted" ? "upload_deleted" : "deletion_pending"}`);
  return brandingAnswer(result, "upload_deleted", "The deletion");
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
