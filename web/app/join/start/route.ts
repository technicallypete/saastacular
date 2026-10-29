import { joinStartRoute } from "@/lib/enrollment-client";
import { methodNotAllowed } from "@/lib/session";

// G058 (Plan r3 §1.6): POST /join/start — the enrollment-intent sign-in start (optionally one invitation intent and a
// request on arrival); the handler and its checks live in lib/enrollment-client.ts.
export const dynamic = "force-dynamic";

export function POST(request: Request): Promise<Response> {
  return joinStartRoute(request);
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
