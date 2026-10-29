import { headers } from "next/headers";
import type { ReactNode } from "react";

import { ButtonLink } from "@/components/controls/button-link";
import { AppShell, type ShellBrand, type ShellNavItem } from "@/components/layout/app-shell";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text } from "@/components/vestive/typography";
import { deriveContext, type RequestContext } from "@/lib/request-context";
import { publicEntryOf } from "@/lib/signup-client";
import { cookieProfile, ownOrigin, type CookieProfile } from "@/lib/session";

// G062 (Plan r5 §1.1, §4.4): the platform root and www.<domain> frame. The generic frame's header carries an Account
// link, which would be a second (sign-in) action on the root; the entry frame with the platform defaults (no label, logo
// or overrides; anonymous) shows the configured product name, appearance and keyboard help only.
export const PLATFORM_ENTRY: ShellBrand = { label: null, logo: null, indicator: { kind: "anonymous" }, entry: true };

// The configured product name and platform domain come from the page (components never read server configuration).
export function SignupShell({
  productName,
  platformBaseDomain,
  nav,
  children,
}: {
  productName: string;
  platformBaseDomain: string;
  nav: ShellNavItem[];
  children: ReactNode;
}) {
  return (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={nav} brand={PLATFORM_ENTRY}>
      {children}
    </AppShell>
  );
}

// The request facts of a root signup page: only on the https profile at the platform root itself (a public entry whose
// own origin is core's signup_origin); undefined anywhere else (the page answers the fixed 404).
export type RootPage = { ctx: RequestContext; origin: string; profile: CookieProfile; cookie: string | null };

export async function rootPage(): Promise<RootPage | undefined> {
  const requestHeaders = await headers();
  const ctx = deriveContext(requestHeaders);
  let profile: CookieProfile;
  try {
    profile = cookieProfile();
  } catch {
    return undefined;
  }
  const origin = ownOrigin(requestHeaders, profile);
  const entry = await publicEntryOf(ctx, origin);
  if (entry === undefined || !entry.root || origin === undefined) return undefined;
  return { ctx, origin, profile, cookie: requestHeaders.get("cookie") };
}

// G062 (Plan r5 §1.1, §4.4; Spec R1): the platform root and www.<domain> landing. Its only authentication or acquisition
// action is Sign Up, an absolute link to the configured root origin's /signup (so www canonicalizes to the root). No
// sign-in, tenant picker, membership directory or lookup, and no tenant's landing. The product name is configuration.
export function Landing({ productName, signupOrigin }: { productName: string; signupOrigin: string }) {
  return (
    <Section id="landing" aria-labelledby="landing-title">
      <Kicker>Get started</Kicker>
      <Heading level={1} id="landing-title" data-testid="product-name" className="mt-6 [overflow-wrap:anywhere]">
        {productName}
      </Heading>
      <Text variant="lead" className="mt-10 [overflow-wrap:anywhere]">
        Create your organization and its own address. You will sign in to it there.
      </Text>
      {/* G069 S24/S15: the view's one call to action, the kit's angled CTA, only here where the platform palette is
          guaranteed. Its focus outline is inset and white on the blue fill, so the angled clip never cuts it. */}
      <ButtonLink
        href={`${signupOrigin}/signup`}
        variant="cta"
        data-testid="root-sign-up"
        className="mt-10 focus-visible:outline-offset-[-3px] focus-visible:outline-white"
      >
        Sign Up
      </ButtonLink>
    </Section>
  );
}
