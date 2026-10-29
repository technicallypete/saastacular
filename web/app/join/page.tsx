import { notFound } from "next/navigation";

import { StatusMessage } from "@/components/account/status-message";
import { shellBrand, tenantRequest, viewerSession } from "@/components/branding/tenant-frame";
import { JoinForm } from "@/components/enrollment/join-form";
import { AppShell, type ShellNavItem } from "@/components/layout/app-shell";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { joinNotice, publicEnrollment } from "@/lib/enrollment-client";
import { lookupContext } from "@/lib/host";

// G058 (Plan r3 §1.6; Spec R5): the public entry for joining a workspace on a tenant host. It shows only the current
// mode's intentional entry: "Sign in to join" (open), "Sign in to request access" (request access) or that membership
// is by invitation (no form). It never lists or reveals anything else. Server-rendered and never cached; a ?status=
// selects fixed text from a closed map and is never reflected.
export const dynamic = "force-dynamic";

const NAV: ShellNavItem[] = [{ href: "#join", label: "Join" }];

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

export default async function JoinPage({ searchParams }: { searchParams: SearchParams }) {
  const { productName, platformBaseDomain } = runtimeConfig();
  const req = await tenantRequest();
  if (req.tenant === undefined || req.origin === undefined) notFound();
  const status = (await searchParams).status;
  const notice = joinNotice(typeof status === "string" ? status : undefined);
  const viewer = await viewerSession(req);
  const pub = await publicEnrollment(lookupContext(req.ctx), req.origin);
  if (!pub.ok && pub.error === "unknown_host") notFound();
  const name = req.tenant.label ?? productName;
  let body;
  if (!pub.ok) {
    body = <StatusMessage message="Joining is unavailable right now. Please try again later." className="mt-10" />;
  } else if (pub.value.mode === "open") {
    body = (
      <>
        <Text className="mt-10">{`Anyone who signs in can join ${name}.`}</Text>
        <JoinForm label="Sign in to join" testId="join-start" />
      </>
    );
  } else if (pub.value.mode === "request_access") {
    body = (
      <>
        <Text className="mt-10">{`Sign in to ask to join ${name}. An administrator decides each request.`}</Text>
        <JoinForm label="Sign in to request access" request testId="join-start" />
      </>
    );
  } else {
    body = (
      <Text className="mt-10" data-testid="join-invite-only">
        Membership is by invitation. Open the invitation link you received to join.
      </Text>
    );
  }
  return (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={NAV} brand={shellBrand(req, viewer, productName)}>
      <Section id="join" aria-labelledby="join-title">
        <Kicker>Join</Kicker>
        <Heading level={1} id="join-title" className="mt-6 [overflow-wrap:anywhere]">
          {`Join ${name}`}
        </Heading>
        {notice === undefined ? null : <StatusMessage message={notice} className="mt-10" />}
        {body}
      </Section>
    </AppShell>
  );
}
