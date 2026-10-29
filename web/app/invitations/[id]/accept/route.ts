import { acceptInvitation, isId, membershipRefusal, membershipUnknown, outcomeUnknown } from "@/lib/membership-client";
import type { StatusLink } from "@/lib/onboarding-client";
import { deriveContext } from "@/lib/request-context";
import { badRequest, field, methodNotAllowed, readCookie, readForm, seeOther } from "@/lib/session";

// G019 (Plan r1 §1.6): accept a Group invitation. The existing form-route order (method, form encoding, exact Origin,
// the 16 KiB body budget); the token arrives in the form body only and is forwarded in core's JSON body, never in a URL,
// a Location header or a log. Core answers one closed not_found for every unknown, expired, revoked, foreign,
// wrong-token or wrong-recipient case. Success is a 303 to the Group with the closed outcome.
export const dynamic = "force-dynamic";

const INVITATIONS: StatusLink = { href: "/invitations", label: "Your invitations" };

export async function POST(request: Request, { params }: { params: Promise<{ id: string }> }): Promise<Response> {
  const { id } = await params;
  if (!isId(id)) return membershipRefusal("not_found", undefined, INVITATIONS);
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  const token = field(form.fields, "token");
  if (typeof csrfToken !== "string" || csrfToken === "" || typeof token !== "string") return badRequest();
  const back: StatusLink = { href: `/invitations/${id}`, label: "Back to the invitation" };
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return membershipRefusal("unauthenticated", undefined, back);
  const r = await acceptInvitation(deriveContext(request.headers), credential, { origin: form.origin, csrfToken }, id, token.trim());
  if (r.ok) return seeOther(r.value.groupId === null ? `/invitations?status=${r.value.outcome}` : `/groups/${r.value.groupId}?status=${r.value.outcome}`);
  if (r.error === "not_found") {
    return membershipRefusal("not_found", undefined, INVITATIONS);
  }
  return outcomeUnknown(r.error) ? membershipUnknown(INVITATIONS) : membershipRefusal(r.error, r.field, back);
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
