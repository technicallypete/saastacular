import { NextResponse, type NextRequest } from "next/server";

import { isGateExempt, isSignupPath, lookupContext, publicEntryAllows, resolveHost, rootRedirects, rootSignupRedirect, tenantAllowsSignup } from "@/lib/host";
import { deriveContext } from "@/lib/request-context";
import { cookieProfile, jsonResponse, ownOrigin, statusPage } from "@/lib/session";

// G016 tenant-host gate (Plan r4 §1.6, §5.3): every page and form request is served only on a host core resolves
// right now. A platform or custom host that routes continues; the shared auth host serves only its fixed /sign-in page
// (sign-in itself runs on /auth/*); anything else is the fixed 404 page and a store failure is 503 (never a clean unknown
// host). Core is asked on every request (no cache), without a credential. /auth/*, /mcp/*, /.well-known/* and the probes
// are exempt (their own routes decide). The loopback-http cookie profile (plain-HTTP loopback development, refused in
// production) has no tenant hosts and is not gated. A gated JSON route (/api/*) gets its refusal as JSON
// ({"error":"not_found"} or {"error":"unavailable"}), as the route itself answers; pages get the fixed HTML status pages.
//
// G062 (Plan r5 §1.1, §4.4; supersedes G016's marketing redirect): a public entry host is served, never redirected. The
// platform root (the request's own origin equals core's signup_origin) serves / and the root signup surface; www.<domain>
// serves / only (its landing links Sign Up to the root origin). On a tenant host /signup/* is the fixed 404 except the
// two continuation hops on a platform host.
//
// G071 (Plan r4 S6): when core aliases the exact platform root to its configured default tenant, the root answers as a
// platform host (root_alias) and passes like any tenant host, except that every /signup path there, the continuation
// hops included, is the fixed 404. No redirect.
//
// G071 (Plan r4 S16, S17): on the exact platform root (not www) a GET or HEAD of exactly / answers 307 with the fixed
// same-origin Location <signup_origin>/signup (core's own answer, not the request's URL) and Cache-Control: no-store when core's root presentation is signup_redirect; the query is
// never read or forwarded, and no cookie, core call or sign-in is involved (the host lookup above is the only call and
// carries no credential). The landing presentation, www, /signup itself, other paths and other methods are unchanged.

export async function proxy(request: NextRequest): Promise<Response> {
  const pathname = request.nextUrl.pathname;
  if (isGateExempt(pathname)) return NextResponse.next();
  const api = pathname.startsWith("/api/");
  let origin: string | undefined;
  try {
    const profile = cookieProfile();
    if (profile === "loopback-http") return NextResponse.next();
    origin = ownOrigin(request.headers, profile);
  } catch {
    return unavailable(api);
  }
  if (origin === undefined) return notFound(api);
  // L-2: the gate's lookup precedes the page's or form's own core call, so it carries its own generated request ID.
  const host = await resolveHost(lookupContext(deriveContext(request.headers)), origin);
  if (!host.ok) return host.error === "unknown_host" ? notFound(api) : unavailable(api);
  switch (host.value.kind) {
    case "platform":
    case "custom": {
      const rootAlias = host.value.kind === "platform" && host.value.rootAlias;
      return !isSignupPath(pathname) || tenantAllowsSignup(pathname, host.value.kind, rootAlias) ? NextResponse.next() : notFound(api);
    }
    case "auth":
      return pathname === "/sign-in" ? NextResponse.next() : notFound(api);
    case "public_entry": {
      const root = origin === host.value.signupOrigin;
      if (rootRedirects(request.method, pathname, root, host.value.rootPage)) return rootSignupRedirect(host.value.signupOrigin);
      return publicEntryAllows(pathname, root) ? NextResponse.next() : notFound(api);
    }
    default:
      return notFound(api);
  }
}

function notFound(api: boolean): Response {
  return api ? jsonResponse(404, { error: "not_found" }) : statusPage(404, "Not found", "There is nothing at this address.");
}

function unavailable(api: boolean): Response {
  return api
    ? jsonResponse(503, { error: "unavailable" })
    : statusPage(503, "Unavailable", "This service is not available right now. Please try again shortly.");
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
