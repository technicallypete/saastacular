import { headers } from "next/headers";

import { ConnectionManager } from "@/components/account/connection-manager";
import { StatusMessage } from "@/components/account/status-message";
import { shellBrand, tenantRequest, viewerSession } from "@/components/branding/tenant-frame";
import { ButtonLink } from "@/components/controls/button-link";
import { AppShell, type ShellNavItem } from "@/components/layout/app-shell";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { isListCursor, listConnections } from "@/lib/mcp-client";
import { deriveContext } from "@/lib/request-context";
import { cookieProfile, ownOrigin, readCookie } from "@/lib/session";

// MCP connection setup on this tenant host (G014 Plan r8 §3.2, §3.4): the signed-in member's own connections in
// this host's Organization (connection URL, status, bound client ID, created and activated times), a create form and a
// revoke form per connection, each carrying this session's CSRF token. Server-rendered and never cached. Anonymous
// visitors get a sign-in-first page; a visitor who is not a member of this host's Organization gets the generic "not
// available" page. The connection URL is a selector shown to its owner, never a credential. G063: this compatibility
// entrypoint renders the same shared connection manager as My Account (friendly names, rename, create with an
// optional name), keeping its whole-page sign-in-first, not-available and invalid-page branches.
export const dynamic = "force-dynamic";

const NAV: ShellNavItem[] = [
  { href: "#connections", label: "Connections" },
  { href: "#create", label: "New connection" },
];

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

// G017: the tenant frame (public branding and the membership indicator, both read with lookup request IDs because the
// page's own core call is the connections list).
async function Shell({ children, nav }: { children: React.ReactNode; nav: ShellNavItem[] }) {
  const { productName, platformBaseDomain } = runtimeConfig();
  const req = await tenantRequest();
  const brand = req.tenant === undefined ? undefined : shellBrand(req, await viewerSession(req, true), productName);
  return (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={nav} brand={brand}>
      {children}
    </AppShell>
  );
}

function SignInFirst() {
  return (
    <Shell nav={NAV.slice(0, 1)}>
      <Section id="connections" aria-labelledby="mcp-title">
        <Kicker>AI connections</Kicker>
        <Heading level={1} id="mcp-title" className="mt-6">
          Sign in first
        </Heading>
        <Text className="mt-10" data-testid="mcp-sign-in-first">
          Sign in to this workspace to create or manage connections for your AI client.
        </Text>
        <ButtonLink href="/sign-in" className="mt-10">Sign in</ButtonLink>
      </Section>
    </Shell>
  );
}

function NotAvailable({ message }: { message: string }) {
  return (
    <Shell nav={NAV.slice(0, 1)}>
      <Section id="connections" aria-labelledby="mcp-title">
        <Kicker>AI connections</Kicker>
        <Heading level={1} id="mcp-title" className="mt-6">
          Not available
        </Heading>
        <StatusMessage message={message} className="mt-10" />
      </Section>
    </Shell>
  );
}

// A page link that is malformed or not valid for this account and host (the cursor is opaque; no detail is shown).
function InvalidPage() {
  return (
    <Shell nav={NAV.slice(0, 1)}>
      <Section id="connections" aria-labelledby="mcp-title">
        <Kicker>AI connections</Kicker>
        <Heading level={1} id="mcp-title" className="mt-6">
          Page link no longer valid
        </Heading>
        <StatusMessage message="This page of connections is not valid for your account here." className="mt-10" />
        <ButtonLink href="/connections/mcp" data-testid="mcp-page-newest" className="mt-10">
          Show your newest connections
        </ButtonLink>
      </Section>
    </Shell>
  );
}

export default async function McpConnections({ searchParams }: { searchParams: SearchParams }) {
  const requestHeaders = await headers();
  let credential: string | undefined;
  let origin: string | undefined;
  try {
    const profile = cookieProfile();
    credential = readCookie(requestHeaders.get("cookie"), "session", profile);
    origin = ownOrigin(requestHeaders, profile);
  } catch {
    credential = undefined;
  }
  if (credential === undefined) return <SignInFirst />;
  if (origin === undefined) return <NotAvailable message="Connections are unavailable right now. Please try again later." />;
  const params = await searchParams;
  const cursorParam = params.cursor;
  if (cursorParam !== undefined && (typeof cursorParam !== "string" || !isListCursor(cursorParam))) return <InvalidPage />;
  const ctx = deriveContext(requestHeaders);
  const list = await listConnections(ctx, credential, origin, cursorParam);
  if (!list.ok && list.error === "unauthenticated") return <SignInFirst />;
  if (!list.ok && (list.error === "not_available" || list.error === "origin_rejected")) {
    return <NotAvailable message="Connections are not available for your account here." />;
  }
  if (!list.ok && list.error === "invalid_request" && cursorParam !== undefined) return <InvalidPage />;
  if (!list.ok) return <NotAvailable message="Connections are unavailable right now. Please try again later." />;
  // G063 (Plan r1 §5.1-5.2): the body is the shared connection manager, given this page's own list answer (one core
  // call) and its existing ids and cursor/status parameters.
  return (
    <Shell nav={NAV}>
      <ConnectionManager
        page="connections"
        ctx={ctx}
        credential={credential}
        origin={origin}
        cursor={cursorParam}
        status={params.status}
        preloaded={list}
      />
    </Shell>
  );
}
