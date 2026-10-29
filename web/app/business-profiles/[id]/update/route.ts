import {
  fieldError,
  isId,
  outcomeUnknown,
  PROFILE_FORM_BYTES,
  readProfileInput,
  refusalPage,
  unknownOutcomePage,
  updateProfile,
  type StatusLink,
} from "@/lib/onboarding-client";
import { deriveContext } from "@/lib/request-context";
import { badRequest, field, methodNotAllowed, readCookie, readForm, seeOther } from "@/lib/session";

// Save the shared Business Profile (G018 Plan r5 §1.5–§1.6): replaces its editable details; the last save wins. The
// existing form-route order with the profile form budget (PROFILE_FORM_BYTES), then core checks the session, the CSRF
// token and the update decision on fresh facts (a member is forbidden: nothing changed). An outcome core may have
// committed is the unknown-state page, which never retries.
export const dynamic = "force-dynamic";

type Context = { params: Promise<{ id: string }> };

export async function POST(request: Request, { params }: Context): Promise<Response> {
  const { id } = await params;
  const form = await readForm(request, { maxBytes: PROFILE_FORM_BYTES });
  if (!form.ok) return form.response;
  if (!isId(id)) return refusalPage("not_found", { href: "/business-profiles", label: "Your Business Profiles" });
  const back: StatusLink = { href: `/business-profiles/${id}`, label: "Back to the Business Profile" };
  const get = (name: string) => field(form.fields, name);
  const csrfToken = get("csrf_token");
  const input = readProfileInput(get);
  if (typeof csrfToken !== "string" || csrfToken === "" || input === undefined) return badRequest();
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return refusalPage("unauthenticated", back);
  const result = await updateProfile(deriveContext(request.headers), credential, { origin: form.origin, csrfToken }, id, input);
  if (result.ok) return seeOther(`/business-profiles/${id}?status=updated`);
  const refused = fieldError(result.field);
  if (result.error === "invalid_request" && refused !== undefined && refused.field !== "name" && refused.field !== "slug") {
    return seeOther(`/business-profiles/${id}?error=${result.field}`);
  }
  if (outcomeUnknown(result.error)) return unknownOutcomePage("update", id);
  return refusalPage(result.error, back);
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
