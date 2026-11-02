import { fetchAvatar, isAvatarVersion, isProfileId } from "@/lib/profile-client";
import { deriveContext } from "@/lib/request-context";
import { cookieProfile, methodNotAllowed, ownOrigin, readCookie } from "@/lib/session";

// G074 (Plan r3.1 §3.4, §3.7, §3.8 delivery): GET /profiles/<user id>/avatar?v=<avatar version> serves a user's current
// avatar, proxied from core for this request's own origin with the browser's own session credential: core re-decides
// the viewer's session and relationship to the subject on every read (no signed URL, nothing cached server-side). Only
// the reviewed headers leave: the image type, its length, the bounded private 300 s cache contract, the ETag of its
// version, nosniff and inline. No session, a refusal of any kind (an unknown, foreign, unrelated or revoked subject, a
// stale version, an ended session) and a malformed request are one no-store 404; core unavailable is a no-store 503.
export const dynamic = "force-dynamic";

function refused(status: 404 | 503): Response {
  return new Response(null, { status, headers: { "Cache-Control": "no-store", "X-Content-Type-Options": "nosniff" } });
}

export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }): Promise<Response> {
  const { id } = await params;
  const versions = new URL(request.url).searchParams.getAll("v");
  if (!isProfileId(id) || versions.length !== 1 || !isAvatarVersion(versions[0])) return refused(404);
  let origin: string | undefined;
  let credential: string | undefined;
  try {
    const profile = cookieProfile();
    origin = ownOrigin(request.headers, profile);
    credential = readCookie(request.headers.get("cookie"), "session", profile);
  } catch {
    return refused(503);
  }
  if (origin === undefined || credential === undefined) return refused(404);
  const avatar = await fetchAvatar(deriveContext(request.headers), credential, origin, id, versions[0]);
  if (!avatar.ok) return refused(avatar.error === "not_found" ? 404 : 503);
  return new Response(Buffer.from(avatar.bytes), {
    status: 200,
    headers: {
      "Content-Type": avatar.contentType,
      "Content-Length": String(avatar.bytes.byteLength),
      "Cache-Control": "private, max-age=300",
      ETag: `"v${versions[0]}"`,
      "X-Content-Type-Options": "nosniff",
      "Content-Disposition": "inline",
    },
  });
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
