import { headers } from "next/headers";
import { notFound } from "next/navigation";

import { StatusMessage } from "@/components/account/status-message";
import { shellBrand, tenantRequest, viewerSession } from "@/components/branding/tenant-frame";
import { SignInAgain, StatusCard } from "@/components/enrollment/status-card";
import { AppShell, type ShellNavItem } from "@/components/layout/app-shell";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { enrollmentStatus, joinNotice, publicEnrollment } from "@/lib/enrollment-client";
import { lookupContext } from "@/lib/host";
import { cookieProfile, readCookie } from "@/lib/session";

// G058 (Plan r3 §1.3 steps 3–6, §1.6): the applicant's own joining state, read from core with the enrollment credential
// (the reconciliation read) on every request. It shows no tenant data and never claims membership: a confirmed
// admission (the ?status=admitted notice or core's historical "admitted") leads to the ordinary sign-in, whose guards
// decide. Without a live enrollment credential the page says so (a dead credential's cookie is cleared by the next
// /join form, which starts again). Server-rendered and never cached; ?status= selects fixed text only.
export const dynamic = "force-dynamic";

const NAV: ShellNavItem[] = [{ href: "#join-status", label: "Joining" }];

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

export default async function JoinStatusPage({ searchParams }: { searchParams: SearchParams }) {
  const { productName, platformBaseDomain } = runtimeConfig();
  const req = await tenantRequest();
  if (req.tenant === undefined || req.origin === undefined) notFound();
  const raw = (await searchParams).status;
  const status = typeof raw === "string" ? raw : undefined;
  const notice = joinNotice(status);
  let credential: string | undefined;
  try {
    credential = readCookie((await headers()).get("cookie"), "enrollment", cookieProfile());
  } catch {
    credential = undefined;
  }
  const viewer = await viewerSession(req);
  let body;
  if (credential === undefined) {
    // No live enrollment credential: after a confirmed admission the next step is the ordinary sign-in; otherwise start again.
    body =
      status === "admitted" || status === "already_member" ? (
        <SignInAgain />
      ) : (
        <div className="mt-10">
          <TextLink href="/join">Start joining</TextLink>
        </div>
      );
  } else {
    const view = await enrollmentStatus(lookupContext(req.ctx), credential, req.origin);
    if (view.ok) {
      const pub = view.value.status === "none" ? await publicEnrollment(lookupContext(req.ctx), req.origin) : undefined;
      body = <StatusCard view={view.value} mode={pub?.ok ? pub.value.mode : undefined} />;
    } else if (view.error === "unauthenticated") {
      body = (
        <>
          <StatusMessage message="Your joining session ended or expired." className="mt-10" />
          <div className="mt-10">
            <TextLink href="/join">Start again</TextLink>
          </div>
        </>
      );
    } else if (view.error === "unknown_host") {
      notFound();
    } else {
      body = <StatusMessage message="Your joining state cannot be shown right now. Please try again later." className="mt-10" />;
    }
  }
  return (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={NAV} brand={shellBrand(req, viewer, productName)}>
      <Section id="join-status" aria-labelledby="join-status-title">
        <Kicker>Joining</Kicker>
        <Heading level={1} id="join-status-title" className="mt-6">
          Your request to join
        </Heading>
        {notice === undefined ? null : <StatusMessage message={notice} className="mt-10" />}
        {body}
      </Section>
    </AppShell>
  );
}
