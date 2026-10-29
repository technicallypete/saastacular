import { deriveContext } from "@/lib/request-context";
import { hopSecret, prove, refusedRoot, rootOfRootRequest } from "@/lib/signup-client";
import { clearCookie, cookieProfile, methodNotAllowed, ownOrigin, readCookie, seeOther, type CookieProfile } from "@/lib/session";

// G062 (Plan r5 §1.4 step 3, §4.4): GET /signup/prove?n=<N> on the platform root (the host gate passes it only there).
// The browser's original host-only root signup credential spends its own row's nonce: core answers the tenant's
// /signup/complete?p=<P> URL, and the browser is sent there. Without a signup cookie (for example a captured N opened in
// another browser) the browser goes to /signup; a dead credential clears both root signup cookies; any other refusal
// (signup_context_mismatch: nothing consumed) sends it to the recovery page /signup/details?status=continue_failed.
// No CSRF token: this GET-initiated hop's only effect is spending the presenting credential's own current nonce (Plan r5
// §4.2, D-15). Every answer is a no-store, no-referrer 303; N and P are never logged.
export const dynamic = "force-dynamic";

export async function GET(request: Request): Promise<Response> {
  let profile: CookieProfile;
  try {
    profile = cookieProfile();
  } catch {
    return refusedRoot("unavailable");
  }
  const origin = ownOrigin(request.headers, profile);
  if (origin === undefined) return refusedRoot("not_found");
  const ctx = deriveContext(request.headers);
  const root = await rootOfRootRequest(ctx, origin);
  if (!root.ok) return refusedRoot(root.error);
  const credential = readCookie(request.headers.get("cookie"), "signup", profile);
  if (credential === undefined) return seeOther("/signup?status=expired");
  const nonce = hopSecret(new URL(request.url).search.slice(1), "n");
  if (nonce === undefined) return seeOther("/signup/details?status=continue_failed");
  const result = await prove(ctx, credential, origin, nonce);
  if (result.ok) return seeOther(result.value);
  switch (result.error) {
    case "unauthenticated":
      return seeOther("/signup?status=expired", clearCookie("signup", profile), clearCookie("signup_form", profile));
    case "rate_limited":
      return seeOther("/signup/details?status=rate_limited");
    default:
      return seeOther("/signup/details?status=continue_failed");
  }
}

export function POST(): Response {
  return methodNotAllowed("GET");
}

export function PUT(): Response {
  return methodNotAllowed("GET");
}

export function PATCH(): Response {
  return methodNotAllowed("GET");
}

export function DELETE(): Response {
  return methodNotAllowed("GET");
}

export function OPTIONS(): Response {
  return methodNotAllowed("GET");
}
