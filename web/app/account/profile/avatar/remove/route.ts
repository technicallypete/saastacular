import { readMutationFields } from "@/lib/branding";
import { profileAnswer } from "@/lib/profile";
import { removeAvatar } from "@/lib/profile-client";
import { deriveContext } from "@/lib/request-context";
import { badRequest, field, methodNotAllowed, readCookie, readForm, statusPage } from "@/lib/session";

// G074 (Plan r3.1 §3.7, §3.8 Remove): withdraw the signed-in user's own avatar (initials are shown instead). The existing
// form-route order, then core decides; the withdrawal is durable when core answers, whatever its storage cleanup state.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const m = readMutationFields((name) => field(form.fields, name));
  if (m === undefined) return badRequest();
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
  const result = await removeAvatar(deriveContext(request.headers), credential, { origin: form.origin, ...m });
  return profileAnswer(result, "avatar_removed");
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
