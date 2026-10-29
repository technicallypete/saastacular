import { groupLabelActionRoute, isId, labelsRefusal } from "@/lib/labels-client";
import { methodNotAllowed } from "@/lib/session";

// G020 (Plan r2 §1.5, §4.2): POST /groups/<uuid>/labels/action — one Group's own label definitions (create, rename,
// visibility, delete; core decides label.group.manage on the exact Group) and the assignment of the Organization's labels
// to it from the Group page's Labels section (assign, unassign; core decides label.organization.manage and group.read).
// A non-canonical id is the fixed not-found page.
export const dynamic = "force-dynamic";

export async function POST(request: Request, { params }: { params: Promise<{ id: string }> }): Promise<Response> {
  const { id } = await params;
  if (!isId(id)) return labelsRefusal("not_found", undefined, { href: "/groups", label: "Your Groups" });
  return groupLabelActionRoute(request, id);
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
