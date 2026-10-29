import { labelActionRoute } from "@/lib/labels-client";
import { methodNotAllowed } from "@/lib/session";

// G020 (Plan r2 §1.5, §4.2): POST /organization/labels/action — the Organization's own label definitions (create, rename,
// delete); core decides the authority (label.organization.manage: anyone else is "Not permitted"), the CSRF token, the
// name rules and the Organization-wide context of a deletion.
export const dynamic = "force-dynamic";

export function POST(request: Request): Promise<Response> {
  return labelActionRoute(request);
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
