import {
  completeOnboarding,
  fieldError,
  isId,
  outcomeUnknown,
  PROFILE_FORM_BYTES,
  readProfileInput,
  refusalPage,
  unknownOutcomePage,
  type Completion,
  type StatusLink,
} from "@/lib/onboarding-client";
import { deriveContext } from "@/lib/request-context";
import { badRequest, field, methodNotAllowed, readCookie, readForm, seeOther } from "@/lib/session";

// Complete a Group's onboarding with a new or an existing Business Profile (G018 Plan r5 §1.4, §1.6). The existing
// form-route order with the profile form budget (PROFILE_FORM_BYTES), then core checks the session, the CSRF token, the
// Group capability and, for an existing profile, that the caller may select it, and attaches exactly one profile. A Group
// that already has one is conflict: the Group page shows its current state. An outcome core may have committed is the
// unknown-state page, which never retries.
export const dynamic = "force-dynamic";

type Context = { params: Promise<{ id: string }> };

export async function POST(request: Request, { params }: Context): Promise<Response> {
  const { id } = await params;
  const form = await readForm(request, { maxBytes: PROFILE_FORM_BYTES });
  if (!form.ok) return form.response;
  if (!isId(id)) return refusalPage("not_found", { href: "/groups", label: "Your Groups" });
  const back: StatusLink = { href: `/groups/${id}`, label: "Back to the Group" };
  const get = (name: string) => field(form.fields, name);
  const csrfToken = get("csrf_token");
  const mode = get("mode");
  if (typeof csrfToken !== "string" || csrfToken === "") return badRequest();
  let choice: Completion;
  if (mode === "new") {
    const profile = readProfileInput(get);
    if (profile === undefined || get("business_profile_id") !== undefined) return badRequest();
    choice = { mode: "new", profile };
  } else if (mode === "existing") {
    const businessProfileId = get("business_profile_id");
    if (!isId(businessProfileId)) return badRequest();
    choice = { mode: "existing", businessProfileId };
  } else {
    return badRequest();
  }
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return refusalPage("unauthenticated", back);
  const result = await completeOnboarding(deriveContext(request.headers), credential, { origin: form.origin, csrfToken }, id, choice);
  if (result.ok) return seeOther(`/groups/${id}?status=completed`);
  if (result.error === "conflict") return seeOther(`/groups/${id}?status=already_complete`);
  const refused = fieldError(result.field);
  if (result.error === "invalid_request" && choice.mode === "new" && refused !== undefined && refused.field !== "name" && refused.field !== "slug") {
    return seeOther(`/groups/${id}/onboarding?error=${result.field}`);
  }
  if (outcomeUnknown(result.error)) return unknownOutcomePage("completion", id);
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
