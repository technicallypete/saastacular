import { lookupContext } from "@/lib/host";
import { session } from "@/lib/identity-client";
import { isSlug, listGroups, MAX_SLUG } from "@/lib/onboarding-client";
import { deriveContext } from "@/lib/request-context";
import { cookieProfile, jsonResponse, ownOrigin, readCookie, type CookieProfile } from "@/lib/session";

// G018 M1 (D2, Option B): the advisory address check of the New Group form, an authenticated same-origin GET. It adds
// no core authority: a grammar-valid slug goes to core's existing guarded exact-slug Group list (session bound to this
// origin's Organization, group.read per row under RLS); any other value is "invalid" after core confirms the session.
// The answer is closed: {state, slug} with state in_use (a readable Group holds the slug), not_seen (none the caller can
// read: a filtered absence, never availability) or invalid, and slug the request's own value (at most 63 characters)
// so the browser applies only the answer to what it asked; otherwise a fixed error. Never a name, id or count; no CORS
// headers; no-store, nosniff, no-referrer (jsonResponse). The create's unique key remains the authoritative check.
export const dynamic = "force-dynamic";

const UNAUTHENTICATED = { error: "unauthenticated" } as const;
const INVALID_REQUEST = { error: "invalid_request" } as const;
const NOT_AVAILABLE = { error: "not_available" } as const;
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
  // Exactly one slug parameter of a bounded length and nothing else: no ambiguous or unexpected parameters.
  const query = new URL(request.url).searchParams;
  const keys = [...query.keys()];
  const slug = query.get("slug");
  if (keys.length !== 1 || keys[0] !== "slug" || slug === null || slug === "" || slug.length > MAX_SLUG) return jsonResponse(400, INVALID_REQUEST);
  const ctx = deriveContext(request.headers);
  if (!isSlug(slug)) {
    // No Group lookup for a value outside the grammar; core still resolves the session for this origin.
    const viewer = await session(ctx, credential, origin);
    if (viewer.ok) return viewer.value.membership === "member" ? jsonResponse(200, { state: "invalid", slug }) : jsonResponse(403, NOT_AVAILABLE);
    return viewer.error === "unauthenticated" ? jsonResponse(401, UNAUTHENTICATED) : jsonResponse(503, UNAVAILABLE);
  }
  const found = await listGroups(lookupContext(ctx), credential, origin, { slug, limit: 1 });
  if (found.ok) return jsonResponse(200, { state: found.value.groups.length === 1 ? "in_use" : "not_seen", slug });
  switch (found.error) {
    case "unauthenticated":
      return jsonResponse(401, UNAUTHENTICATED);
    case "not_available":
    case "forbidden":
    case "origin_rejected":
    case "not_found":
      return jsonResponse(403, NOT_AVAILABLE);
    default:
      // An upstream failure or an answer outside the contract is never "not_seen".
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
