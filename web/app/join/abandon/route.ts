import { applicantRoute } from "@/lib/enrollment-client";
import { methodNotAllowed } from "@/lib/session";

// G058 (Plan r3 §1.6): POST /join/abandon — revokes the applicant's own enrollment credential and clears its cookie.
export const dynamic = "force-dynamic";

export function POST(request: Request): Promise<Response> {
  return applicantRoute(request, "abandon");
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
