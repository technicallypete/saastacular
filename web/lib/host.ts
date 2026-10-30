import "server-only";

import { randomBytes } from "node:crypto";

import { runtimeConfig } from "./config";
import { outboundHeaders, type RequestContext } from "./request-context";

// G016 (Plan r4 §1.6, §1.7, §5.3): web's tenant-host gate and the shared callback relay. Core alone decides what a host
// is; this module asks it (GET /internal/v1/hosts/resolve?origin=, GET /internal/v1/auth/relay?state=), sends exactly
// the request's diagnostic headers and no credential, and reduces the answer to a closed result. A store failure is
// "unavailable", never a clean unknown host. Nothing here is cached: a removed or reassigned host stops routing on the
// next request. G062 (Plan r5 §1.1, §4.4): the public entry answer names the platform root origin (signup_origin) and
// the gate's path sets for the root, www and tenant hosts live here. G071 (Plan r4 S16, S17): the public entry answer
// also names core's root presentation (root_page), and the root's redirect decision lives here.

// The context of a host lookup made before the request's own core call (G016 A1.5 L-2): core keeps one completion
// record per request ID, so the lookup carries a freshly generated request ID with the same trace, caller span and
// flags, while the request's own core call keeps ctx's (inbound or generated) request ID. The narrow diagnostic
// exception of OPERATIONS.md "Propagation allow-list"; used only by the host gate and the /auth/callback route.
export function lookupContext(ctx: RequestContext): RequestContext {
  return { ...ctx, requestId: randomBytes(16).toString("hex"), requestIdSource: "generated" };
}

export const HOST_TIMEOUT_MS = 5_000;
const MAX_RESPONSE_BYTES = 4 * 1024;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const LABEL = "[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?";
const ORIGIN = new RegExp(`^https://${LABEL}(?:\\.${LABEL})*(?::[0-9]{1,5})?$`);
// The longest provider callback query web forwards (the callback route's own bound).
export const MAX_RELAY_QUERY = 8 * 1024;
const MAX_STATE = 512;

export type HostResolution =
  // G071 (Plan r4 S5, S6): rootAlias is true only when core aliases the exact platform root to its configured default
  // tenant (the answer's root_alias member); every other platform host has rootAlias false.
  | { kind: "platform"; organizationId: string; rootAlias: boolean }
  | { kind: "custom"; organizationId: string; callbackOrigin: string | null }
  | { kind: "auth" }
  | { kind: "reserved" }
  // G062 (Plan r5 §1.1, §4.4): the platform domain and www.<domain> are the public entry; signupOrigin is the exact
  // configured platform root origin (www canonicalizes to it). There is no redirect location any more. G071 (Plan r4
  // S16, S17): rootPage is core's effective root presentation (the answer's root_page member); only the exact root
  // applies it, www ignores it.
  | { kind: "public_entry"; signupOrigin: string; rootPage: RootPage };
// G071 (Plan r4 S15, S16): the multi-tenant root presentation core selected at startup (PLATFORM_ROOT_PAGE, effective
// value): a temporary redirect of the exact root / to /signup, or the existing landing.
export type RootPage = "signup_redirect" | "landing";
export type HostError = "unknown_host" | "unavailable" | "malformed";
export type HostResult = { ok: true; value: HostResolution } | { ok: false; error: HostError };
export type RelayResult = { ok: true; callback: string } | { ok: false; error: "relay_refused" | "unavailable" | "malformed" };

type Json = Readonly<Record<string, unknown>>;
type Answer = { status: number; body: unknown } | undefined;

function isObject(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

// The paths the gate never applies to: sign-in and its relay (/auth/*), the MCP transport (core's own host check),
// well-known metadata and the probes.
export function isGateExempt(pathname: string): boolean {
  return (
    pathname.startsWith("/auth/") ||
    pathname.startsWith("/mcp/") ||
    pathname.startsWith("/.well-known/") ||
    pathname === "/api/health" ||
    pathname === "/api/ready"
  );
}

// Whether value is an exact https origin (scheme, lower-case host labels and an optional port; no path, query,
// userinfo or fragment).
export function isHttpsOrigin(value: unknown): value is string {
  return typeof value === "string" && value.length <= 300 && ORIGIN.test(value);
}

// G071 (Plan r4 S16): exactly one of the two presentation names (no case folding: an enum, not a hostname).
function isRootPage(value: unknown): value is RootPage {
  return value === "signup_redirect" || value === "landing";
}

// The closed reading of core's hosts/resolve answer (exported for the unit tests). G062: a public entry carries exactly
// signup_origin, an exact https origin; the retired location member (or any other member) is malformed. G071 (Plan r4
// S5, S6, §4): a platform answer may carry exactly root_alias: true (core's aliased platform root); root_alias false,
// any other value, or the member on any other kind is malformed. G071 (Plan r4 S16, §4): a public entry also carries
// exactly root_page, one of "signup_redirect" or "landing"; a missing or other value, or the member on any other kind,
// is malformed.
export function readResolution(status: number, body: unknown): HostResult {
  if (status === 404 && isObject(body) && body.error === "unknown_host" && Object.keys(body).length === 1) {
    return { ok: false, error: "unknown_host" };
  }
  if (status === 503) return { ok: false, error: "unavailable" };
  if (status !== 200 || !isObject(body)) return { ok: false, error: "malformed" };
  const { kind, organization_id: org, callback_origin: cb, signup_origin: so, root_alias: alias, root_page: page } = body;
  const keys = Object.keys(body).length;
  switch (kind) {
    case "platform":
      if (typeof org === "string" && UUID.test(org) && (alias === undefined ? keys === 2 : alias === true && keys === 3)) {
        return { ok: true, value: { kind, organizationId: org, rootAlias: alias === true } };
      }
      break;
    case "custom":
      if (typeof org === "string" && UUID.test(org) && (cb === undefined ? keys === 2 : isHttpsOrigin(cb) && keys === 3)) {
        return { ok: true, value: { kind, organizationId: org, callbackOrigin: typeof cb === "string" ? cb : null } };
      }
      break;
    case "auth":
    case "reserved":
      if (keys === 1) return { ok: true, value: { kind } };
      break;
    case "public_entry":
      if (isHttpsOrigin(so) && isRootPage(page) && keys === 3) return { ok: true, value: { kind, signupOrigin: so, rootPage: page } };
      break;
  }
  return { ok: false, error: "malformed" };
}

// G071 (Plan r4 S17; Spec §1 "Multi-tenant root presentation"): whether the gate answers this public entry request with
// the temporary redirect to /signup. Only a GET or HEAD of exactly / on the exact platform root (root: the request's own
// origin equals core's signup_origin) in the signup_redirect presentation redirects; landing, www.<domain>, /signup
// itself, every other path and every other method keep the existing handling. The query is never part of the decision.
export function rootRedirects(method: string, pathname: string, root: boolean, rootPage: RootPage): boolean {
  return root && pathname === "/" && (method === "GET" || method === "HEAD") && rootPage === "signup_redirect";
}

// The redirect itself (S17): 307 with the fixed same-origin Location <signup_origin>/signup (built from core's own
// signup_origin answer, never from the request's Host, Origin or URL; nothing of the request's query is read or
// forwarded) and Cache-Control: no-store; no body and no cookie. The Location is absolute because the proxy adapter of
// the pinned Next.js (16.4.1) rejects a relative one with ERR_INVALID_URL (a 500).
export function rootSignupRedirect(signupOrigin: string): Response {
  return new Response(null, { status: 307, headers: { Location: new URL("/signup", signupOrigin).toString(), "Cache-Control": "no-store" } });
}

// G062 (Plan r5 §1.1, §4.4): the paths the gate passes on a public entry host. The root origin (the request's own origin
// equals signup_origin) serves / and the root signup surface (/signup, /signup/details, /signup/start, /signup/create,
// /signup/abandon, /signup/prove); www.<domain> serves / only. Both serve the shell's public static files. The tenant
// continuation hops (/signup/continue, /signup/complete) and every other path are the fixed 404 there. /auth/* and the
// probes are exempt before this runs.
const ROOT_SIGNUP_PATHS: ReadonlySet<string> = new Set([
  "/signup",
  "/signup/details",
  "/signup/start",
  "/signup/create",
  "/signup/abandon",
  "/signup/prove",
]);
// The tenant-host continuation hops (G062 Plan r5 §4.4): served on a routing platform host only.
const TENANT_SIGNUP_PATHS: ReadonlySet<string> = new Set(["/signup/continue", "/signup/complete"]);

// The shell's own public static files (the self-hosted fonts and the blueprint mask of globals.css), which the landing
// and signup pages load from their own origin.
const PUBLIC_ASSET_PREFIXES = ["/fonts/", "/blueprint/"] as const;

export function publicEntryAllows(pathname: string, root: boolean): boolean {
  if (pathname === "/" || PUBLIC_ASSET_PREFIXES.some((prefix) => pathname.startsWith(prefix))) return true;
  return root && ROOT_SIGNUP_PATHS.has(pathname);
}

// OS002: the default product mark (the header's /brand/mark.png) and the favicon (/icon.png, web/app/icon.png) are fixed
// same-origin static files that the root, www and auth host pages reference; nothing a request carries selects the file.
// The gate passes exactly a GET or HEAD of exactly those two paths on the public entry and auth hosts, after the host has
// resolved; any other method, path, spelling or file (the rest of /brand/ included) keeps the fixed 404 there.
const PRODUCT_ASSET_PATHS: ReadonlySet<string> = new Set(["/brand/mark.png", "/icon.png"]);

export function productAssetAllows(method: string, pathname: string): boolean {
  return (method === "GET" || method === "HEAD") && PRODUCT_ASSET_PATHS.has(pathname);
}

// OS002: the auth host's fixed /sign-in page references four of the shell's self-hosted font files (globals.css), which the
// gate had not served there (a 404 for each, present before OS002). The gate passes exactly a GET or HEAD of exactly those four
// shipped files on the auth host, after the host has resolved; /fonts/ is no prefix exemption there, so any other font file,
// spelling or method keeps the fixed 404.
const AUTH_FONT_PATHS: ReadonlySet<string> = new Set([
  "/fonts/geist/geist-wght.ttf",
  "/fonts/oswald/oswald-wght.ttf",
  "/fonts/barlow-condensed/barlow-condensed-400.ttf",
  "/fonts/barlow-condensed/barlow-condensed-600.ttf",
]);

export function authAssetAllows(method: string, pathname: string): boolean {
  return (method === "GET" || method === "HEAD") && AUTH_FONT_PATHS.has(pathname);
}

// Whether the path is /signup or below it.
export function isSignupPath(pathname: string): boolean {
  return pathname === "/signup" || pathname.startsWith("/signup/");
}

// Whether a /signup path is served on a routing tenant host of this kind: only the two continuation hops, and only on a
// platform host (core refuses a continuation on a custom host as continuation_invalid; the gate does not pass it).
// G071 (Plan r4 S6; Spec ambiguity (d)): on the aliased platform root no /signup path is served, the continuation hops
// included, so every one of them is the fixed 404 there.
export function tenantAllowsSignup(pathname: string, kind: "platform" | "custom", rootAlias = false): boolean {
  return !rootAlias && kind === "platform" && TENANT_SIGNUP_PATHS.has(pathname);
}

// The relay's state value: exactly one non-empty state parameter of at most 512 characters, read without
// re-serializing the query; undefined otherwise.
export function relayState(rawQuery: string): string | undefined {
  if (rawQuery.length > MAX_RELAY_QUERY) return undefined;
  const values = new URLSearchParams(rawQuery).getAll("state");
  return values.length === 1 && values[0] !== "" && values[0].length <= MAX_STATE ? values[0] : undefined;
}

// The relay destination: core's callback (exactly https://<custom host>/auth/callback) with the provider's raw query
// appended byte-identical; undefined when the callback is outside that shape (exported for the unit tests).
export function relayLocation(callback: string, rawQuery: string): string | undefined {
  if (!callback.endsWith("/auth/callback") || !ORIGIN.test(callback.slice(0, -"/auth/callback".length))) return undefined;
  return rawQuery === "" ? callback : `${callback}?${rawQuery}`;
}

async function get(ctx: RequestContext, path: string): Promise<Answer> {
  let response: Response;
  try {
    response = await fetch(new URL(path, runtimeConfig().coreInternalUrl), {
      method: "GET",
      headers: { ...outboundHeaders(ctx) },
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(HOST_TIMEOUT_MS),
    });
  } catch {
    return undefined; // no answer in time or no connection: unavailable, never a clean unknown host
  }
  try {
    const text = await response.text();
    if (text.length > MAX_RESPONSE_BYTES) return { status: response.status, body: undefined };
    return { status: response.status, body: text === "" ? undefined : (JSON.parse(text) as unknown) };
  } catch {
    return { status: response.status, body: undefined };
  }
}

// GET /internal/v1/hosts/resolve?origin=<origin>: the routing decision of this request's own origin.
export async function resolveHost(ctx: RequestContext, origin: string): Promise<HostResult> {
  const answer = await get(ctx, `/internal/v1/hosts/resolve?origin=${encodeURIComponent(origin)}`);
  if (answer === undefined) return { ok: false, error: "unavailable" };
  return readResolution(answer.status, answer.body);
}

// GET /internal/v1/auth/relay?state=<state>: core's relay decision for one provider callback on the auth host.
export async function relayCallback(ctx: RequestContext, state: string): Promise<RelayResult> {
  const answer = await get(ctx, `/internal/v1/auth/relay?state=${encodeURIComponent(state)}`);
  if (answer === undefined || answer.status >= 500) return { ok: false, error: "unavailable" };
  const body = answer.body;
  if (answer.status === 404 && isObject(body) && body.error === "relay_refused") return { ok: false, error: "relay_refused" };
  if (answer.status === 200 && isObject(body) && Object.keys(body).length === 1 && typeof body.callback === "string") {
    return { ok: true, callback: body.callback };
  }
  return { ok: false, error: "malformed" };
}
