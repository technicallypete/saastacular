import { notFound } from "next/navigation";
import type { ReactNode } from "react";

import { StatusMessage } from "@/components/account/status-message";
import { shellBrand, tenantRequest } from "@/components/branding/tenant-frame";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { AppShell, type ShellNavItem, type WorkspaceArea } from "@/components/layout/app-shell";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text } from "@/components/vestive/typography";
import { lookupContext } from "@/lib/host";
import { session } from "@/lib/identity-client";
import type { OnboardingError } from "@/lib/onboarding-client";

// G018 (Plan r5 §1.6): the shared request and frame of the onboarding pages, following the branding page. Onboarding
// exists on a tenant host only (anything else is the fixed 404). The page's own core call is the session read (its CSRF
// token feeds the forms; core's list answers carry none); every onboarding read that follows carries a lookup request
// ID. Core alone resolves the Organization and decides every read; a refusal renders a fixed page, never a guess.

export type Frame = (children: ReactNode) => ReactNode;

export type OnboardingRequest =
  | { ready: false; page: ReactNode }
  | {
      ready: true;
      frame: Frame;
      // A fresh lookup request ID for each onboarding call (core keeps one completion record per request ID).
      lookup: () => ReturnType<typeof lookupContext>;
      credential: string;
      origin: string;
      csrfToken: string;
      // The fixed page of a refused or failed onboarding read (not_found is the fixed 404).
      refused: (error: OnboardingError) => ReactNode;
    };

// config: the page's runtime configuration (components never read server configuration themselves).
type Copy = { area: WorkspaceArea; kicker: string; nav: ShellNavItem[]; config: { productName: string; platformBaseDomain: string } };

function Notice({ kicker, title, children }: { kicker: string; title: string; children: ReactNode }) {
  return (
    <Section id="onboarding" aria-labelledby="onboarding-title">
      <Kicker>{kicker}</Kicker>
      <Heading level={1} id="onboarding-title" className="mt-6">
        {title}
      </Heading>
      {children}
    </Section>
  );
}

export async function onboardingRequest({ area, kicker, nav, config }: Copy): Promise<OnboardingRequest> {
  const { productName, platformBaseDomain } = config;
  const req = await tenantRequest();
  if (req.tenant === undefined) notFound();
  const frameFor = (viewer: Parameters<typeof shellBrand>[1]): Frame =>
    function frame(children: ReactNode) {
      return (
        <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={nav} brand={shellBrand(req, viewer, productName)} area={area}>
          {children}
        </AppShell>
      );
    };
  const signInFirst = frameFor(undefined)(
    <Notice kicker={kicker} title="Sign in first">
      <Text className="mt-10" data-testid="onboarding-sign-in-first">
        Sign in to this workspace to see its Groups and Business Profiles.
      </Text>
      {/* G018 M1 (D4): the existing guarded login form, as on the sign-in page and the tenant entry. */}
      <form method="post" action="/auth/login" className="mt-10">
        <PendingSubmit label="Sign in" pendingLabel="Signing in…" />
      </form>
    </Notice>,
  );
  if (req.credential === undefined || req.origin === undefined) return { ready: false, page: signInFirst };
  const viewer = await session(req.ctx, req.credential, req.origin);
  if (!viewer.ok && viewer.error === "unauthenticated") return { ready: false, page: signInFirst };
  if (!viewer.ok && viewer.error === "unknown_host") notFound();
  const frame = frameFor(viewer);
  const message = (title: string, text: string) =>
    frame(
      <Notice kicker={kicker} title={title}>
        <StatusMessage message={text} className="mt-10" />
      </Notice>,
    );
  const unavailable = () => message("Not available", "This is unavailable right now. Please try again later.");
  if (!viewer.ok) return { ready: false, page: unavailable() };
  const refused = (error: OnboardingError): ReactNode => {
    if (error === "unauthenticated") return signInFirst;
    if (error === "not_found" || error === "origin_rejected") notFound();
    if (error === "not_available" || error === "forbidden") {
      return message("Not available", "This is not available for your account in this workspace.");
    }
    return unavailable();
  };
  return {
    ready: true,
    frame,
    lookup: () => lookupContext(req.ctx),
    credential: req.credential,
    origin: req.origin,
    csrfToken: viewer.value.csrfToken,
    refused,
  };
}

// One query parameter as a string, or undefined when absent or repeated.
export function param(value: string | string[] | undefined): string | undefined {
  return typeof value === "string" ? value : undefined;
}

export type SearchParams = Promise<Record<string, string | string[] | undefined>>;
