import {
  codePoints,
  createGroup,
  fieldError,
  isSlug,
  MAX_NAME,
  outcomeUnknown,
  refusalPage,
  unknownOutcomePage,
  type StatusLink,
} from "@/lib/onboarding-client";
import { deriveContext } from "@/lib/request-context";
import { badRequest, field, methodNotAllowed, readCookie, readForm, seeOther } from "@/lib/session";

// Create a Group owned by the caller (G018 Plan r5 §1.4, §1.6). The existing form-route order (method, form encoding,
// exact Origin, bounded body: the 16 KiB default), then core checks the session, the CSRF token, the current
// Organization membership and the create_group capability, validates the name and address, and inserts the Group with
// its owner binding. A taken address returns to the form with the honest inline conflict (never "created"); an outcome
// core may have committed is the unknown-state page, which never retries.
export const dynamic = "force-dynamic";

const BACK: StatusLink = { href: "/groups/new", label: "Back to the new Group form" };

// The submitted name as a query parameter for the form to re-check with core (an empty or over-long name is dropped).
function keepName(name: string): string {
  return name !== "" && codePoints(name) <= MAX_NAME ? `&name=${encodeURIComponent(name)}` : "";
}

export async function POST(request: Request): Promise<Response> {
  const form = await readForm(request);
  if (!form.ok) return form.response;
  const csrfToken = field(form.fields, "csrf_token");
  const name = field(form.fields, "name");
  const rawSlug = field(form.fields, "slug");
  if (typeof csrfToken !== "string" || csrfToken === "" || typeof name !== "string" || typeof rawSlug === "symbol") return badRequest();
  const slug = rawSlug === undefined || rawSlug === "" ? null : rawSlug;
  const credential = readCookie(form.cookie, "session", form.profile);
  if (credential === undefined) return refusalPage("unauthenticated", BACK);
  const ctx = deriveContext(request.headers);
  const result = await createGroup(ctx, credential, { origin: form.origin, csrfToken }, { name, slug });
  if (result.ok) return seeOther(`/groups/${result.value.id}/onboarding`);
  if (result.error === "invalid_request") {
    const refused = fieldError(result.field);
    if (refused?.field === "name") return seeOther(`/groups/new?error=${result.field}`);
    // slug_taken is web-local (core's conflict below); core never sends it as a field code.
    if (refused?.field === "slug" && result.field !== "slug_taken") {
      // The name returns for core to re-check with a new suggestion; the refused address is not echoed.
      return seeOther(`/groups/new?error=${result.field}${keepName(name)}`);
    }
    return refusalPage("invalid_request", BACK);
  }
  if (result.error === "conflict") {
    // G018 M1 (D3): core's conflict is definitive (refused before COMMIT), so it returns to the form as the web-local
    // closed code slug_taken with the name and, only when grammar-valid, the refused address (core validated it before
    // answering conflict). The create page makes the fresh authorized holder read. An outcome that may have committed
    // stays the separate unknown-state page below.
    const keepSlug = slug !== null && isSlug(slug) ? `&slug=${slug}` : "";
    return seeOther(`/groups/new?error=slug_taken${keepName(name)}${keepSlug}`);
  }
  if (outcomeUnknown(result.error)) return unknownOutcomePage("group");
  return refusalPage(result.error, BACK);
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
