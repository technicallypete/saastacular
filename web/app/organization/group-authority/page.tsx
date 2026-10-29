import { StatusMessage } from "@/components/account/status-message";
import { GroupAuthorityForm } from "@/components/membership/group-authority-form";
import { asOnboardingError } from "@/components/membership/members-view";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { GROUP_AUTHORITY_STATUS, groupAuthority, ROLES, selfMembership } from "@/lib/membership-client";

// G060 (Plan r0 §1.6, D-11; Spec R2, R4): the Organization's Group-administration setting. Core's read gives the current
// state and the owner hint; the caller's own Organization role chooses the view: an owner gets the form (core still
// decides the write: organization.manage_policy, owner only), an admin the read-only effective state, anyone else the
// fixed 404 (the page is Organization administration, not member content). Tenant hosts only; server-rendered and never
// cached.
export const dynamic = "force-dynamic";

export default async function GroupAuthorityPage({ searchParams }: { searchParams: SearchParams }) {
  const r = await onboardingRequest({ config: runtimeConfig(), area: "members", kicker: "Group administration", nav: [{ href: "#group-authority", label: "Group administration" }] });
  if (!r.ready) return r.page;
  const [authority, self] = await Promise.all([
    groupAuthority(r.lookup(), r.credential, r.origin),
    selfMembership(r.lookup(), r.credential, r.origin, { kind: "organization" }),
  ]);
  if (!self.ok) return r.refused(asOnboardingError(self.error));
  const owner = self.value.role === ROLES.organization.owner;
  if (!owner && self.value.role !== ROLES.organization.admin) return r.refused("not_found");
  if (!authority.ok) return r.refused(asOnboardingError(authority.error));
  const status = param((await searchParams).status);
  const notice = status !== undefined && Object.hasOwn(GROUP_AUTHORITY_STATUS, status) ? GROUP_AUTHORITY_STATUS[status] : undefined;
  return r.frame(
    <Section id="group-authority" aria-labelledby="group-authority-title">
      <Kicker>Group administration</Kicker>
      <Heading level={1} id="group-authority-title" className="mt-6">
        Group administration
      </Heading>
      <Text className="mt-10">
        Owners and admins of this workspace can always see every Group and its Business Profile. By default only a Group&apos;s own
        owners and admins manage it.
      </Text>
      <StatusMessage message={notice} className="mt-10" />
      <GroupAuthorityForm authority={authority.value} canChange={owner && authority.value.canChange} csrfToken={r.csrfToken} />
      <div className="mt-12 flex flex-wrap gap-x-9 gap-y-3">
        <TextLink href="/organization/members">Members</TextLink>
        <TextLink href="/groups">Groups</TextLink>
      </div>
    </Section>,
  );
}
