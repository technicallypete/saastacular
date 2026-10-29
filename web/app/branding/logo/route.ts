import { fetchLogo } from "@/lib/branding";
import { deriveContext } from "@/lib/request-context";
import { cookieProfile, methodNotAllowed, ownOrigin } from "@/lib/session";

// G017 (Plan r0 §1.5, r2 §5.5): GET /branding/logo?v=<sha12> on a tenant host (after the host gate) serves the current
// logo of the Organization this host routes to, proxied from core for this request's own origin. No credential is
// read or sent. Only the reviewed headers leave: the logo type, its length, the bounded 24 h public cache contract,
// the ETag of its version, nosniff and inline. Any refusal is a no-store 404 (core unavailable: a no-store 503).
export const dynamic = "force-dynamic";

const SHA12 = /^[0-9a-f]{12}$/;

function refused(status: 404 | 503): Response {
  return new Response(null, { status, headers: { "Cache-Control": "no-store", "X-Content-Type-Options": "nosniff" } });
}

export async function GET(request: Request): Promise<Response> {
  const versions = new URL(request.url).searchParams.getAll("v");
  if (versions.length !== 1 || !SHA12.test(versions[0])) return refused(404);
  let origin: string | undefined;
  try {
    const profile = cookieProfile();
    origin = profile === "loopback-http" ? undefined : ownOrigin(request.headers, profile);
  } catch {
    return refused(503);
  }
  if (origin === undefined) return refused(404);
  const logo = await fetchLogo(deriveContext(request.headers), origin, versions[0]);
  if (!logo.ok) return refused(logo.error === "not_found" ? 404 : 503);
  return new Response(Buffer.from(logo.bytes), {
    status: 200,
    headers: {
      "Content-Type": logo.contentType,
      "Content-Length": String(logo.bytes.byteLength),
      "Cache-Control": "public, max-age=86400, immutable",
      ETag: `"${versions[0]}"`,
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
