import { groupAuthorityActionRoute } from "@/lib/membership-client";
import { methodNotAllowed } from "@/lib/session";

// G060 (Plan r0 §1.6, §4.2, D-11): POST /organization/group-authority/action — the owner-only Group-administration
// setting; core decides the authority (a non-owner is "Not permitted"), the CSRF token and the current state.
export const dynamic = "force-dynamic";

export function POST(request: Request): Promise<Response> {
  return groupAuthorityActionRoute(request);
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
