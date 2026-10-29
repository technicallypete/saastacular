import { deriveContext } from "@/lib/request-context";
import { complete, hopSecret, refusedRoot, rootOfTenantRequest } from "@/lib/signup-client";
import { clearCookie, cookieProfile, methodNotAllowed, ownOrigin, readCookie, seeOther, setCookie, type CookieProfile } from "@/lib/session";

// G062 (Plan r5 §1.4 step 4, §4.4): GET /signup/complete?p=<P> on the new tenant's platform host (the host gate passes it
// only there). The browser's host-only tenant challenge cookie (C) and the proof P must belong to the same continuation
// generation for this exact origin and its Organization; core then starts the pinned login (expected user, current
// membership) and answers the ordinary login start. The challenge cookie is cleared on every answer; on success the
// fresh login binding is set and the browser goes to the identity provider (a silent return while its session is live;
// otherwise its prompt). Any refusal sends the browser to the root's recovery page
// (/signup/details?status=continue_failed); the root origin is configuration confirmed by core, never a query parameter.
// No credential is read or forwarded. Every answer is a no-store, no-referrer 303; P and C are never logged.
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
  const clearChallenge = clearCookie("challenge", profile);
  const failed = seeOther(`${root.root}/signup/details?status=continue_failed`, clearChallenge);
  const cookie = request.headers.get("cookie");
  const challengeValue = readCookie(cookie, "challenge", profile);
  const proof = hopSecret(new URL(request.url).search.slice(1), "p");
  if (challengeValue === undefined || proof === undefined) return failed;
  const result = await complete(ctx, { origin, challenge: challengeValue, proof }, readCookie(cookie, "login", profile));
  if (result.ok) {
    const { authorizationUrl, binding } = result.value;
    return seeOther(authorizationUrl, clearChallenge, setCookie("login", binding.value, binding.maxAge, profile));
  }
  if (result.error === "rate_limited") return seeOther(`${root.root}/signup/details?status=rate_limited`, clearChallenge);
  return failed;
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
