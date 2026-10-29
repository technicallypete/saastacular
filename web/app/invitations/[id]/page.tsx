import { StatusMessage } from "@/components/account/status-message";
import { AcceptForm } from "@/components/membership/accept-form";
import { asOnboardingError } from "@/components/membership/members-view";
import { onboardingRequest } from "@/components/onboarding/onboarding-frame";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { INVALID_INVITATION } from "@/lib/enrollment-client";
import { getInvitation, isId } from "@/lib/membership-client";

// G019 (Plan r1 §1.6): one invitation's acceptance page, showing what core lets this recipient read. A pending Group
// invitation offers the accept form; an Organization invitation has no web acceptance in G019 (G058 owns it); any
// other state is stated, never accepted. Server-rendered and never cached.
// G060 owner amendment (Option B): a malformed id (never sent to core) and core's closed not_found (nonexistent,
// foreign and wrong-recipient ids are one answer) render the same fixed INVALID_INVITATION sentence in this page's frame
// instead of the fixed 404, so the page never says whether an invitation exists or whom it was for. Sign-in, the
// context checks of the shared frame, other refusals and infrastructure failures keep their existing handling.
export const dynamic = "force-dynamic";

export default async function InvitationPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const r = await onboardingRequest({ config: runtimeConfig(), area: "members", kicker: "Invitation", nav: [{ href: "#invitation", label: "Invitation" }] });
  if (!r.ready) return r.page;
  const inv = isId(id) ? await getInvitation(r.lookup(), r.credential, r.origin, id) : ({ ok: false, error: "not_found" } as const);
  if (!inv.ok && inv.error !== "not_found") return r.refused(asOnboardingError(inv.error));
  let body;
  if (!inv.ok) {
    body = <StatusMessage message={INVALID_INVITATION} className="mt-10" />;
  } else if (inv.value.scope === "organization") {
    body = <StatusMessage message="Accepting a workspace invitation on the web is not available yet. Keep your invitation link." className="mt-10" />;
  } else if (inv.value.status !== "pending") {
    body = <StatusMessage message="This invitation can no longer be accepted (it has expired, was revoked or was already used)." className="mt-10" />;
  } else {
    body = <AcceptForm invitation={inv.value} csrfToken={r.csrfToken} />;
  }
  return r.frame(
    <Section id="invitation" aria-labelledby="invitation-title">
      <Kicker>Invitation</Kicker>
      <Heading level={1} id="invitation-title" className="mt-6">
        Invitation
      </Heading>
      {body}
      <div className="mt-12">
        <TextLink href="/invitations">Your invitations</TextLink>
      </div>
    </Section>,
  );
}
