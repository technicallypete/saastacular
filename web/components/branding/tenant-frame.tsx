import { headers } from "next/headers";
import { cache } from "react";

import type { ShellBrand } from "@/components/layout/app-shell";
import { indicatorState, resolveBranding, type Brand } from "@/lib/branding";
import { lookupContext } from "@/lib/host";
import { session, type CoreResult, type SessionView } from "@/lib/identity-client";
import { deriveContext, type RequestContext } from "@/lib/request-context";
import { cookieProfile, ownOrigin, readCookie } from "@/lib/session";

// G017 (Plan r0 §1.5): the request facts a page needs for the tenant frame. tenant is the public branding of the
// Organization this request's origin routes to (undefined on the loopback profile or a host core does not route as a
// tenant, where the generic frame stays byte-for-byte the same); a branding store failure or malformed answer keeps the
// tenant frame with the platform defaults and shows no tenant data. The branding read precedes the page's own core call
// (the session read), so it carries a lookup request ID (lookupContext, as the host gate's lookup does).
export type TenantRequest = {
  ctx: RequestContext;
  origin: string | undefined;
  credential: string | undefined;
  tenant: Brand | undefined;
};

const PLATFORM_DEFAULT: Brand = { revision: 0, label: null, logo: null, tokens: {} };

// G069 (Plan r1 S7, §2.4): read once per request. The root layout renders the tenant stylesheet from the same facts the
// page's frame uses (components/branding/tenant-style.tsx), so both share this one answer and one branding read.
export const tenantRequest = cache(async (): Promise<TenantRequest> => {
  const requestHeaders = await headers();
  const ctx = deriveContext(requestHeaders);
  let origin: string | undefined;
  let credential: string | undefined;
  let loopback = true;
  try {
    const profile = cookieProfile();
    loopback = profile === "loopback-http";
    origin = ownOrigin(requestHeaders, profile);
    credential = readCookie(requestHeaders.get("cookie"), "session", profile);
  } catch {
    return { ctx, origin: undefined, credential: undefined, tenant: undefined };
  }
  if (loopback || origin === undefined) return { ctx, origin, credential, tenant: undefined };
  const brand = await resolveBranding(lookupContext(ctx), origin);
  if (brand.ok) return { ctx, origin, credential, tenant: brand.value };
  return { ctx, origin, credential, tenant: brand.error === "unknown_host" ? undefined : PLATFORM_DEFAULT };
});

// The session answer for this origin, or undefined when the browser holds no session cookie. A page whose own core call
// is another one (the MCP connections list) reads the session for its indicator with a lookup request ID.
export async function viewerSession(req: TenantRequest, lookup = false): Promise<CoreResult<SessionView> | undefined> {
  if (req.credential === undefined || req.origin === undefined) return undefined;
  return session(lookup ? lookupContext(req.ctx) : req.ctx, req.credential, req.origin);
}

// The shell's tenant frame: the public label and logo plus the indicator for this viewer. The colour overrides are not
// part of it: the root layout's tenant stylesheet already renders them for the whole document.
export function shellBrand(req: TenantRequest, viewer: CoreResult<SessionView> | undefined, productName: string): ShellBrand | undefined {
  if (req.tenant === undefined) return undefined;
  const { label, logo } = req.tenant;
  return { label, logo, indicator: indicatorState(viewer, label ?? productName) };
}
