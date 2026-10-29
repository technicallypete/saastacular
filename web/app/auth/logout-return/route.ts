import { providerLogoutReturn, type ProviderReturnOutcome } from "@/lib/identity-client";
import { deriveContext } from "@/lib/request-context";
import { clearCookie, cookieProfile, methodNotAllowed, readCookie, statusPage, type CookieProfile } from "@/lib/session";

// The notification step of a provider sign-out (Plan r5 §3.5; C7 addendum §2). The host-only logout cookie names the
// one-use outcome that core recorded; this route consumes it once through core, shows the closed-map message and
// clears the cookie in this same response. It only reads notification state: it never revokes or changes a session,
// and nothing from the request is reflected. A missing, malformed, duplicate, expired, forged or replayed reference
// never claims a provider outcome.
export const dynamic = "force-dynamic";

const PROVIDER_RETURN: Readonly<Record<ProviderReturnOutcome, string>> = {
  provider_session_revoked: "Signed out of this application. The WorkOS session for this sign-in was revoked.",
  provider_revoke_failed: "Signed out of this application. The WorkOS session for this sign-in could not be revoked or confirmed.",
  local_revoked: "Signed out of this application.",
};

const NOTHING = "There is no sign-out outcome to show here.";

function profileOrUndefined(): CookieProfile | undefined {
  try {
    return cookieProfile();
  } catch {
    return undefined;
  }
}

export async function GET(request: Request): Promise<Response> {
  const profile = profileOrUndefined();
  if (profile === undefined) return statusPage(503, "Sign-out status unavailable", "The sign-out outcome cannot be shown right now.");
  const cleared = clearCookie("logout", profile);
  const ref = readCookie(request.headers.get("cookie"), "logout", profile);
  if (ref === undefined) return statusPage(200, "Signed out", NOTHING, cleared);
  const result = await providerLogoutReturn(deriveContext(request.headers), ref);
  if (!result.ok) {
    return statusPage(503, "Sign-out status unavailable", "The sign-out outcome could not be read right now. Nothing was changed.", cleared);
  }
  return statusPage(200, "Signed out", PROVIDER_RETURN[result.value], cleared);
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
