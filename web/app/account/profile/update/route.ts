import { profileAnswer, readNameForm } from "@/lib/profile";
import { updateProfileName } from "@/lib/profile-client";
import { deriveContext } from "@/lib/request-context";
import { badRequest, methodNotAllowed, readCookie, readForm, statusPage } from "@/lib/session";

// G074 (Plan r3.1 §3.7, §3.10): set or clear the signed-in user's own explicit name (an empty name returns to the sign-in
// provider's name). The existing form-route order (method, form encoding, exact Origin, bounded body), then core checks
// the session, the CSRF token, liveness, the expected version and the name, and writes only the caller's own row.
export const dynamic = "force-dynamic";

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const input = readNameForm(form.fields);
  if (input === undefined) return badRequest();
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
  const result = await updateProfileName(deriveContext(request.headers), credential, { origin: form.origin, ...input.m }, input.name);
  return profileAnswer(result, "profile_saved");
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
