import { manageActionRoute } from "@/lib/enrollment-client";
import { methodNotAllowed } from "@/lib/session";

// G058 (Plan r3 §1.6; selected-design addendum §A.5): POST /organization/enrollment/action — the policy change (owner)
// and the request decisions and approval revocation (administrators); core decides every authority.
export const dynamic = "force-dynamic";

export function POST(request: Request): Promise<Response> {
  return manageActionRoute(request);
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
