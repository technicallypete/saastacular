import { deriveContext } from "@/lib/request-context";
import { challenge, hopSecret, refusedRoot, rootOfTenantRequest } from "@/lib/signup-client";
import { cookieProfile, methodNotAllowed, ownOrigin, readCookie, seeOther, setCookie, type CookieProfile } from "@/lib/session";

// G062 (Plan r5 §1.4 step 2, §4.4): GET /signup/continue?ref=<R> on the new tenant's platform host (the host gate passes
// it only there). The root origin for every redirect is configuration confirmed by core (rootOfTenantRequest), never a
// query parameter. Core consumes R once for this exact origin and its Organization and issues the challenge C and the
// nonce N: C becomes the host-only tenant challenge cookie (120 s) in the same 303 that sends the browser to the root's
// /signup/prove?n=<N> (core's own URL). Any refusal sends the browser to the root's recovery page
// (/signup/details?status=continue_failed); nothing is set. No credential is read or forwarded; only the browser's login
// binding. Every answer is a no-store, no-referrer 303; R, N and C are never logged.
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
  const root = await rootOfTenantRequest(ctx, origin);
  if (!root.ok) return refusedRoot(root.error);
  const failed = `${root.root}/signup/details?status=continue_failed`;
  const ref = hopSecret(new URL(request.url).search.slice(1), "ref");
  if (ref === undefined) return seeOther(failed);
  const result = await challenge(ctx, { origin, ref, root: root.root }, readCookie(request.headers.get("cookie"), "login", profile));
  if (result.ok) {
    const { challenge: c, proveUrl } = result.value;
    return seeOther(proveUrl, setCookie("challenge", c.value, c.maxAge, profile));
  }
  if (result.error === "rate_limited") return seeOther(`${root.root}/signup/details?status=rate_limited`);
  return seeOther(failed);
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
