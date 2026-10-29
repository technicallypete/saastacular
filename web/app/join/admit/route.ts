import { applicantRoute } from "@/lib/enrollment-client";
import { methodNotAllowed } from "@/lib/session";

// G058 (Plan r3 §1.6): POST /join/admit — the applicant's own admission (open, or an approved request). A confirmed
// admission clears the enrollment cookie and never sets a session: the status page then offers the ordinary sign-in.
export const dynamic = "force-dynamic";

export function POST(request: Request): Promise<Response> {
  return applicantRoute(request, "admit");
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
