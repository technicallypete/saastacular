import { StatusMessage } from "@/components/account/status-message";
import { MyInvitations } from "@/components/membership/invitation-list";
import { asOnboardingError } from "@/components/membership/members-view";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { isId, myInvitations } from "@/lib/membership-client";

// G019 (Plan r1 §1.6): the signed-in member's own pending Group invitations in this workspace (role, sender when core
// supplies a name, expiry; keyset pages), each linking to its acceptance page. Server-rendered and never cached.
export const dynamic = "force-dynamic";

// Closed outcomes of an acceptance that core answered without a Group to open (shown as the request's outcome only).
const STATUS: Readonly<Record<string, string>> = {
  already_member: "You already have this role or a higher one, so nothing was changed.",
  consumed: "This invitation was already used, so nothing was changed.",
};

export default async function InvitationsPage({ searchParams }: { searchParams: SearchParams }) {
  const r = await onboardingRequest({ config: runtimeConfig(), area: "members", kicker: "Invitations", nav: [{ href: "#invitations", label: "Invitations" }] });
  if (!r.ready) return r.page;
  const search = await searchParams;
  const after = param(search.after);
  const page = await myInvitations(r.lookup(), r.credential, r.origin, { after: isId(after) ? after : undefined });
  if (!page.ok) return r.refused(asOnboardingError(page.error));
  const status = param(search.status);
  return r.frame(
    <Section id="invitations" aria-labelledby="invitations-title">
      <Kicker>Invitations</Kicker>
      <Heading level={1} id="invitations-title" className="mt-6">
        Your invitations
      </Heading>
      <StatusMessage message={status !== undefined && Object.hasOwn(STATUS, status) ? STATUS[status] : undefined} className="mt-10" />
      <MyInvitations invitations={page.value.invitations} />
      {page.value.next === null ? null : (
        <div className="mt-8">
          <TextLink href={`/invitations?after=${page.value.next}`}>Next invitations</TextLink>
        </div>
      )}
    </Section>,
  );
}
