import { CONTEXT_PARAM, listGroupCards, readContextValues } from "@/lib/group-profile-client";
import { isProfileId } from "@/lib/profile-client";
import { deriveContext } from "@/lib/request-context";
import { cookieProfile, jsonResponse, ownOrigin, readCookie, type CookieProfile } from "@/lib/session";

// G075 (Plan r2 §3.4, §3.5, §3.9, C8, C10): another user's (or the viewer's own) Group member cards as the signed-in
// viewer may see them now, exactly {user_id, cards: [{group_id, group_name, role_label, can_edit, version}],
// narrow_required}. ?context=<Group id> selects one Group (core's Resolve decides selectability; the selector only
// narrows); without it the context is Organization-only. An empty, malformed or repeated ?context= is 400 invalid_request
// and is never dropped to Organization-only. An unknown, foreign, unrelated or revoked subject (and a malformed id) and an
// unselectable context are one 404 with the body of /api/profiles/<id>'s refusal. 401 for no live session at this
// request's own origin, 503 when core cannot answer. Never another member, a count, an email or a Business Profile.
export const dynamic = "force-dynamic";

const UNAUTHENTICATED = { error: "unauthenticated" } as const;
const INVALID_REQUEST = { error: "invalid_request" } as const;
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
  const selector = readContextValues(new URL(request.url).searchParams.getAll(CONTEXT_PARAM));
  if (!selector.ok) return jsonResponse(400, INVALID_REQUEST);
  const result = await listGroupCards(deriveContext(request.headers), credential, origin, id, selector.group);
  if (result.ok) {
    const v = result.value;
    return jsonResponse(200, {
      user_id: v.userId,
      cards: v.cards.map((c) => ({ group_id: c.groupId, group_name: c.groupName, role_label: c.roleLabel, can_edit: c.canEdit, version: c.version })),
      narrow_required: v.narrowRequired,
    });
  }
  switch (result.error) {
    case "unauthenticated":
      return jsonResponse(401, UNAUTHENTICATED);
    case "invalid_request":
      return jsonResponse(400, INVALID_REQUEST);
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
