import { applicantRoute } from "@/lib/enrollment-client";
import { methodNotAllowed } from "@/lib/session";

// G058 (Plan r3 §1.6): POST /join/request — the applicant's own access request with the enrollment credential.
export const dynamic = "force-dynamic";

export function POST(request: Request): Promise<Response> {
  return applicantRoute(request, "request");
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
