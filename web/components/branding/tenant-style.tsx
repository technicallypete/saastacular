import { tenantCss } from "@/lib/tenant-css";
import type { TokenOverrides } from "@/lib/tokens";

// G069 (Plan r1 S7, §2.5): the tenant theme host. The root layout renders it in <head> from the request's re-checked
// Organization branding (components/branding/tenant-frame.tsx), so the first HTML byte already carries the tenant's
// values in both appearances, and every overlay the kit portals into document.body inherits them from its first
// frame. The element holds only the stylesheet text built by lib/tenant-css.ts from saved names and #rrggbb colours;
// a loopback or unknown host, no branding, or a value that fails the re-check renders it empty (the kit's defaults).
// A different host is a different request and a different element; nothing repairs styles after render.
export function TenantStyle({ tokens }: { tokens?: TokenOverrides }) {
  return <style data-tenant-branding="">{tenantCss(tokens)}</style>;
}
