// Test-only route (G063 Plan r1 §5.2, D-5). The browser-test image copies this file to app/account-fixtures/page.tsx in
// its fixture build stage; it is never part of the production web image (asserted by S-ACCOUNT-MCP-UNAVAILABLE).
// It renders the real My Account composition (AccountSections, which embeds the real ConnectionManager) with a scripted
// session. The fixture build's CORE_INTERNAL_URL (http://127.0.0.1:9, playwright.config.ts) makes the manager's real
// list call fail as unreachable, so the page exercises the production fallback: the connections section shows its
// closed unavailable state while Session and Sign-out stay usable.
import { headers } from "next/headers";

import { AccountSections } from "@/components/account/account-sections";
import { AppShell, type ShellNavItem } from "@/components/layout/app-shell";
import { runtimeConfig } from "@/lib/config";
import type { SessionView } from "@/lib/identity-client";
import { deriveContext } from "@/lib/request-context";

export const dynamic = "force-dynamic";

const NAV: ShellNavItem[] = [
  { href: "#session", label: "Session" },
  { href: "#mcp-connections", label: "AI connections" },
  { href: "#sign-out", label: "Sign out" },
];

// A scripted session: fixed values, a dummy CSRF token of the real shape; no credential exists anywhere.
const SESSION: SessionView = {
  userId: "00000000-0000-4000-8000-0000000000f1",
  displayName: "Fixture Account Holder",
  mfaAssurance: "unknown",
  authTime: "2026-10-21T08:00:00Z",
  expiresAt: "2026-10-21T20:00:00Z",
  csrfToken: "F".repeat(43),
  membership: "member",
};

export default async function AccountFixture() {
  const { productName, platformBaseDomain } = runtimeConfig();
  const ctx = deriveContext(await headers());
  return (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={NAV}>
      <AccountSections
        value={SESSION}
        notice={undefined}
        soleOwned={undefined}
        manages={false}
        mcp={{ ctx, credential: "fixture", origin: "http://127.0.0.1:3001", cursor: undefined, status: undefined }}
      />
    </AppShell>
  );
}
