import { isId, memberActionRoute, membershipRefusal } from "@/lib/membership-client";
import { methodNotAllowed } from "@/lib/session";

// G019 (Plan r1 §1.6): the Group members form route, the Organization route's twin for one Group (a canonical id from
// the path; anything else is the fixed not-found page). Core decides every action, including that promote, demote and
// transfer belong to the Group's current owner.
export const dynamic = "force-dynamic";

export async function POST(request: Request, { params }: { params: Promise<{ id: string }> }): Promise<Response> {
  const { id } = await params;
  if (!isId(id)) return membershipRefusal("not_found", undefined, { href: "/groups", label: "Your Groups" });
  return memberActionRoute(request, { kind: "group", groupId: id }, `/groups/${id}/members`);
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
