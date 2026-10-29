import { notFound } from "next/navigation";

import { shellBrand, tenantRequest, viewerSession } from "@/components/branding/tenant-frame";
import { InvitationJoinForm } from "@/components/enrollment/join-form";
import { AppShell, type ShellNavItem } from "@/components/layout/app-shell";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { isId } from "@/lib/enrollment-client";

// G058 (Plan r3 §1.6, D-3r1): the workspace-invitation landing for first-time and known recipients. Before sign-in it
// shows nothing about the invitation beyond the workspace name; the code comes from the link's fragment in the browser
// and is posted to /join/start, which keeps it in the one-use join cookie for the callback. Core decides the invitation
// after sign-in. Tenant hosts only; server-rendered and never cached.
export const dynamic = "force-dynamic";

const NAV: ShellNavItem[] = [{ href: "#join-invitation", label: "Invitation" }];

export default async function JoinInvitationPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  if (!isId(id)) notFound();
  const { productName, platformBaseDomain } = runtimeConfig();
  const req = await tenantRequest();
  if (req.tenant === undefined) notFound();
  const viewer = await viewerSession(req);
  const name = req.tenant.label ?? productName;
  return (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={NAV} brand={shellBrand(req, viewer, productName)}>
      <Section id="join-invitation" aria-labelledby="join-invitation-title">
        <Kicker>Invitation</Kicker>
        <Heading level={1} id="join-invitation-title" className="mt-6 [overflow-wrap:anywhere]">
          {`Join ${name}`}
        </Heading>
        <Text className="mt-10">Sign in to accept your invitation. Your account is created on first sign-in when the invitation allows it.</Text>
        <InvitationJoinForm invitationId={id} />
      </Section>
    </AppShell>
  );
}
