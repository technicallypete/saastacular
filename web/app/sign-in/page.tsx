import { headers } from "next/headers";
import { redirect } from "next/navigation";

import { StatusMessage } from "@/components/account/status-message";
import { shellBrand, tenantRequest, viewerSession } from "@/components/branding/tenant-frame";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { AppShell, type ShellNavItem } from "@/components/layout/app-shell";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { cookieProfile, hasCookie, type CookieProfile } from "@/lib/session";

// Sign in (Plan r5 §5, §7.1; C7 addendum §2). Server-rendered and never cached. Query values are never reflected: a
// status selects fixed text from a closed map, and no query ever carries a provider outcome. The provider's hosted
// logout returns here; when the logout cookie is present this page hands over, with a fixed 307 GET, to
// /auth/logout-return, whose response consumes the recorded outcome once and clears the cookie (a Server Component
// cannot set cookies).
export const dynamic = "force-dynamic";

const NAV: ShellNavItem[] = [{ href: "#sign-in", label: "Sign in" }];

const STATUS: Readonly<Record<string, string>> = {
  signed_out: "You have signed out of this application.",
  signed_out_all: "You have signed out of this application everywhere.",
  canceled: "Sign-in was canceled. Nothing was changed.",
  expired: "That sign-in attempt expired or was already used. Please start again.",
  stale: "That sign-in was not recent enough. Please sign in again.",
  unavailable: "Sign-in is unavailable right now. Please try again later.",
  failed: "Sign-in could not be completed. Please try again.",
  // G016: the host stopped routing to the sign-in's Organization, or the account has no membership of it.
  host_changed: "This address no longer belongs to the workspace you started signing in to. Nothing was changed.",
  not_member: "Your account is not a member of this workspace. Ask an administrator to add you, then sign in again.",
  // G058: after a confirmed admission the ordinary sign-in decides current membership.
  admitted: "Your membership was confirmed. Sign in to open the workspace.",
};

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function single(value: string | string[] | undefined): string | undefined {
  return typeof value === "string" ? value : undefined;
}

function message(params: Record<string, string | string[] | undefined>): string | undefined {
  const status = single(params.status);
  return status !== undefined && Object.hasOwn(STATUS, status) ? STATUS[status] : undefined;
}

function profileOrUndefined(): CookieProfile | undefined {
  try {
    return cookieProfile();
  } catch {
    return undefined;
  }
}

export default async function SignIn({ searchParams }: { searchParams: SearchParams }) {
  const profile = profileOrUndefined();
  if (profile !== undefined && hasCookie((await headers()).get("cookie"), "logout", profile)) {
    redirect("/auth/logout-return");
  }
  const { productName, platformBaseDomain } = runtimeConfig();
  const params = await searchParams;
  const returnToAccount = single(params.return_to) === "/account";
  const notice = message(params);
  // G017: the tenant frame on a platform or custom host (the shared auth host and the loopback profile keep the generic
  // frame); the sign-in action below stays the existing one. G069 (Plan r1 S24): this page is served on tenant hosts, so
  // its submit is the kit Button (the tenant's primary pair applies), never the kit's fixed-white call to action.
  const req = await tenantRequest();
  const brand = req.tenant === undefined ? undefined : shellBrand(req, await viewerSession(req), productName);
  return (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={NAV} brand={brand && { ...brand, entry: true }}>
      <Section id="sign-in" aria-labelledby="sign-in-title">
        <Kicker>Account</Kicker>
        <Heading level={1} id="sign-in-title" className="mt-6">
          Sign in
        </Heading>
        <StatusMessage message={notice} className="mt-10" />
        <Text className="mt-10">
          Sign in with your {productName} identity. You continue on the identity provider&apos;s page and come back here when you are done.
        </Text>
        <form method="post" action="/auth/login" className="mt-10">
          {returnToAccount ? <input type="hidden" name="return_to" value="/account" /> : null}
          <PendingSubmit size="lg" label="Sign in" pendingLabel="Signing in…" />
        </form>
      </Section>
    </AppShell>
  );
}
