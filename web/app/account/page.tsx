import { headers } from "next/headers";
import { notFound, redirect } from "next/navigation";

import { AccountSections } from "@/components/account/account-sections";
import { StatusMessage } from "@/components/account/status-message";
import { shellBrand, tenantRequest } from "@/components/branding/tenant-frame";
import { AppShell, type ShellNavItem } from "@/components/layout/app-shell";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker } from "@/components/vestive/typography";
import { getBranding } from "@/lib/branding";
import { runtimeConfig } from "@/lib/config";
import { lookupContext } from "@/lib/host";
import { session } from "@/lib/identity-client";
import { retirementEligibility } from "@/lib/membership-client";
import { deriveContext } from "@/lib/request-context";
import { cookieProfile, ownOrigin, readCookie } from "@/lib/session";

// The signed-in user's account (Plan r5 §5, §7.1): display name, multi-factor status, session expiry and the
// three sign-out forms, each carrying this session's CSRF token. Server-rendered and never cached. Anonymous
// visitors are sent to sign in. Multi-factor status is never claimed as verified (Plan r5 §3.4). G063 (Plan r1 §5.2):
// the signed-in body is the shared AccountSections composition, which embeds the shared MCP connection manager (its own
// mcp_cursor and mcp_status parameters, disjoint from this page's status notices).
export const dynamic = "force-dynamic";

const SIGN_IN = "/sign-in?return_to=/account";

const NAV: ShellNavItem[] = [
  { href: "#session", label: "Session" },
  { href: "#mcp-connections", label: "AI connections" },
  { href: "#sign-out", label: "Sign out" },
];

const STATUS: Readonly<Record<string, string>> = {
  signed_in: "You are signed in.",
  session_rotated: "You are signed in. Reload any other open tab of this site before using its forms.",
  already_signed_in: "You are already signed in.",
  identity_conflict:
    "A sign-in for a different account was refused because you are already signed in here. Sign out first to switch accounts.",
};

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

export default async function Account({ searchParams }: { searchParams: SearchParams }) {
  const { productName, platformBaseDomain } = runtimeConfig();
  const requestHeaders = await headers();
  let credential: string | undefined;
  let origin: string | undefined;
  try {
    const profile = cookieProfile();
    credential = readCookie(requestHeaders.get("cookie"), "session", profile);
    // G016: the session is answered for this request's own origin together with the host's current routing (one core
    // call); A2: under every cookie profile, so a session bound to another origin is never shown here.
    origin = ownOrigin(requestHeaders, profile);
  } catch {
    credential = undefined;
  }
  if (credential === undefined || origin === undefined) redirect(SIGN_IN);
  const result = await session(deriveContext(requestHeaders), credential, origin);
  if (!result.ok && result.error === "unauthenticated") redirect(SIGN_IN);
  if (!result.ok && result.error === "unknown_host") notFound();
  // G017: the tenant frame; the indicator reuses this page's own session answer (no second session read).
  const req = await tenantRequest();
  const brand = shellBrand(req, result, productName);
  // G017: the branding management link, only for a viewer core lets read the management view here (owner or admin);
  // a lookup request ID, as this page's own core call is the session read.
  const manages = req.tenant !== undefined && result.ok && (await getBranding(lookupContext(req.ctx), origin, credential)).ok;
  // G019 (Plan r1 §1.6, C6): the retirement eligibility notice — only the count of scopes this account solely owns, no
  // names, no ids and no retire action; a failed read shows no notice (the page claims nothing).
  const eligibility = result.ok ? await retirementEligibility(lookupContext(req.ctx), credential, origin) : undefined;
  const soleOwned = eligibility?.ok === true && eligibility.value.count > 0 ? eligibility.value : undefined;
  const params = await searchParams;
  const status = params.status;
  const notice = typeof status === "string" && Object.hasOwn(STATUS, status) ? STATUS[status] : undefined;

  return (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={result.ok ? NAV : NAV.slice(0, 1)} brand={brand}>
      {result.ok ? (
        // G063 (Plan r1 §5.2): the signed-in body, shared with the rendered fixture; the connection manager's one list
        // call uses a lookup request ID, as this page's own core call is the session read.
        <AccountSections
          value={result.value}
          notice={notice}
          soleOwned={soleOwned}
          manages={manages}
          mcp={{ ctx: lookupContext(req.ctx), credential, origin, cursor: params.mcp_cursor, status: params.mcp_status }}
        />
      ) : (
        <Section id="session" aria-labelledby="account-title">
          <Kicker>Account</Kicker>
          <Heading level={1} id="account-title" className="mt-6">
            Your account
          </Heading>
          <StatusMessage message="Your account details are unavailable right now. Please try again later." className="mt-10" />
        </Section>
      )}
    </AppShell>
  );
}
