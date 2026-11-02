import { getProfile, isProfileId } from "@/lib/profile-client";
import { deriveContext } from "@/lib/request-context";
import { cookieProfile, jsonResponse, ownOrigin, readCookie, type CookieProfile } from "@/lib/session";

// G074 (Plan r3.1 §3.4, §3.5, §3.7): another user's profile as the signed-in viewer may see it now, exactly
// {user_id, name, avatar: {version, content_type} | null}. Core decides with the session's own Organization and the
// relationship predicate (self, an owner or admin of a workspace the subject belongs to, or a shared Group); unknown,
// foreign, unrelated and revoked subjects (and a malformed id) are one 404. 401 for no live session at this request's own
// origin, 503 when core cannot answer. Never an email, provider identity, session, Organization list, object id or key.
export const dynamic = "force-dynamic";

const UNAUTHENTICATED = { error: "unauthenticated" } as const;
const NOT_FOUND = { error: "not_found" } as const;
const UNAVAILABLE = { error: "unavailable" } as const;

export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }): Promise<Response> {
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
  const { id } = await params;
  if (!isProfileId(id)) return jsonResponse(404, NOT_FOUND);
  const result = await getProfile(deriveContext(request.headers), credential, origin, id);
  if (result.ok) {
    const p = result.value;
    return jsonResponse(200, {
      user_id: p.userId,
      name: p.name,
      avatar: p.avatar === null ? null : { version: p.avatar.version, content_type: p.avatar.contentType },
    });
  }
  switch (result.error) {
    case "unauthenticated":
      return jsonResponse(401, UNAUTHENTICATED);
    case "not_found":
    case "forbidden":
    case "not_available":
    case "origin_rejected":
    case "unknown_host":
      return jsonResponse(404, NOT_FOUND);
    default:
      return jsonResponse(503, UNAVAILABLE);
  }
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
