import { CONTEXT_FIELD, groupBack, groupLabelAnswer, isGroupId, readContextValues, readLabelForm, setGroupLabel } from "@/lib/group-profile-client";
import { isProfileId } from "@/lib/profile-client";
import { deriveContext } from "@/lib/request-context";
import { badRequest, methodNotAllowed, readCookie, readForm, statusPage } from "@/lib/session";

// G075 (Plan r2 §3.5, §3.7, C6, C10): set or clear one member's role label in one Group from the composite profile page
// (an empty value removes it). The existing form-route order (method, form encoding, exact Origin, bounded body), then
// the hidden context_group selector, if present, must be exactly one canonical Group id (an empty, malformed or repeated
// selector returns invalid_context and is never dropped to Organization-only), then core checks the session, the CSRF
// token, the selected context (Resolve), liveness, the writer (the member itself or the exact Group's own owner or admin),
// the expected version and the label. The edited Group is the path's; the selector is never substituted for it. Back to
// /profiles/<id> with one closed group_status and the same validated selector.
export const dynamic = "force-dynamic";

const NOT_FOUND = () => statusPage(404, "Not found", "There is nothing at this address.");

export async function POST(request: Request, { params }: { params: Promise<{ id: string; group: string }> }): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const { id, group } = await params;
  if (!isProfileId(id) || !isGroupId(group)) return NOT_FOUND();
  const selector = readContextValues(form.fields.getAll(CONTEXT_FIELD));
  if (!selector.ok) return groupBack(id, undefined, "invalid_context");
  const input = readLabelForm(form.fields);
  if (input === undefined) return badRequest();
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return statusPage(401, "Not signed in", "Sign in first, then try again. Nothing was changed.");
  const result = await setGroupLabel(
    deriveContext(request.headers),
    credential,
    { origin: form.origin, ...input.m, contextGroup: selector.group },
    id,
    group,
    input.label,
  );
  return groupLabelAnswer(result, id, selector.group);
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
