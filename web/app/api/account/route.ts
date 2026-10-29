import { session } from "@/lib/identity-client";
import { deriveContext } from "@/lib/request-context";
import { cookieProfile, jsonResponse, ownOrigin, readCookie, type CookieProfile } from "@/lib/session";

// The signed-in user's own account (Plan r5 §7.1). It needs only the session store: 401 for no live session
// whether or not a provider is configured, 503 only when sessions cannot be resolved. It never carries the CSRF
// token or anything about another user. G016 A2: the session is read at this request's own origin, so a credential
// issued for another Organization's host is 401 here, as is a request without a usable origin.
export const dynamic = "force-dynamic";

const UNAUTHENTICATED = { error: "unauthenticated" } as const;
const UNAVAILABLE = { error: "unavailable" } as const;

export async function GET(request: Request): Promise<Response> {
  let profile: CookieProfile;
  try {
    profile = cookieProfile();
  } catch {
    return jsonResponse(503, UNAVAILABLE);
  }
  const credential = readCookie(request.headers.get("cookie"), "session", profile);
  if (credential === undefined) return jsonResponse(401, UNAUTHENTICATED);
  const origin = ownOrigin(request.headers, profile);
  if (origin === undefined) return jsonResponse(401, UNAUTHENTICATED);
  const result = await session(deriveContext(request.headers), credential, origin);
  if (result.ok) {
    const s = result.value;
    return jsonResponse(200, {
      id: s.userId,
      display_name: s.displayName,
      mfa_assurance: s.mfaAssurance,
      auth_time: s.authTime,
      session_expires_at: s.expiresAt,
    });
  }
  return result.error === "unauthenticated" ? jsonResponse(401, UNAUTHENTICATED) : jsonResponse(503, UNAVAILABLE);
}

function methodNotAllowed(): Response {
  const response = jsonResponse(405, { error: "method_not_allowed" });
  response.headers.set("Allow", "GET");
  return response;
}

export function POST(): Response {
  return methodNotAllowed();
}

export function PUT(): Response {
  return methodNotAllowed();
}

export function PATCH(): Response {
  return methodNotAllowed();
}

export function DELETE(): Response {
  return methodNotAllowed();
}

export function OPTIONS(): Response {
  return methodNotAllowed();
}
