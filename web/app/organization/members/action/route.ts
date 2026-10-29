import { memberActionRoute } from "@/lib/membership-client";
import { methodNotAllowed } from "@/lib/session";

// G019 (Plan r1 §1.6): the Organization members form route. One closed action field (invite, revoke, change_role,
// remove, promote, demote, transfer, leave) maps to the matching core route; core decides every action.
export const dynamic = "force-dynamic";

export function POST(request: Request): Promise<Response> {
  return memberActionRoute(request, { kind: "organization" }, "/organization/members");
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
