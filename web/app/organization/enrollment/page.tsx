import { StatusMessage } from "@/components/account/status-message";
import { PolicyForm } from "@/components/enrollment/policy-form";
import { RequestList } from "@/components/enrollment/request-list";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { getPolicy, isId, listRequests, MANAGE_NOTICES } from "@/lib/enrollment-client";
import { ROLES, selfMembership } from "@/lib/membership-client";

// G058 (Plan r3 §1.5–§1.6; selected-design addendum §A.5): the workspace's enrollment page. Every member reads the mode
// (organization.read); the owner changes it (the form is a hint: core decides organization.manage_policy);
// administrators see one page of access requests with Approve/Deny and Revoke on approvals (organization.manage_members;
// a member without it sees no list). Tenant hosts only; server-rendered and never cached.
export const dynamic = "force-dynamic";

export default async function OrganizationEnrollmentPage({ searchParams }: { searchParams: SearchParams }) {
  const r = await onboardingRequest({ config: runtimeConfig(), area: "members", kicker: "Enrollment", nav: [{ href: "#enrollment", label: "Enrollment" }] });
  if (!r.ready) return r.page;
  const search = await searchParams;
  const status = param(search.status);
  const after = param(search.after);
  const policy = await getPolicy(r.lookup(), r.credential, r.origin);
  if (!policy.ok) {
    const e = policy.error;
    return r.refused(e === "not_available" || e === "forbidden" || e === "unauthenticated" || e === "not_found" ? e : "onboarding_unavailable");
  }
  const self = await selfMembership(r.lookup(), r.credential, r.origin, { kind: "organization" });
  const owner = self.ok && self.value.role === ROLES.organization.owner;
  const manager = self.ok && self.value.canManage;
  const requests = manager ? await listRequests(r.lookup(), r.credential, r.origin, isId(after) ? { after } : {}) : undefined;
  const notice = status !== undefined && Object.hasOwn(MANAGE_NOTICES, status) ? MANAGE_NOTICES[status] : undefined;
  return r.frame(
    <Section id="enrollment" aria-labelledby="enrollment-title">
      <Kicker>Enrollment</Kicker>
      <Heading level={1} id="enrollment-title" className="mt-6">
        Joining this workspace
      </Heading>
      {notice === undefined ? null : <StatusMessage message={notice} className="mt-10" />}
      <PolicyForm policy={policy.value} canChange={owner} csrfToken={r.csrfToken} />
      {requests === undefined ? null : (
        <div className="mt-12">
          <Heading level={2}>Access requests</Heading>
          {requests.ok ? (
            <>
              <RequestList requests={requests.value.requests} csrfToken={r.csrfToken} />
              {requests.value.next === null ? null : (
                <div className="mt-6">
                  <TextLink href={`/organization/enrollment?after=${requests.value.next}`}>Next page</TextLink>
                </div>
              )}
            </>
          ) : (
            <Text className="mt-10">The access requests cannot be shown right now.</Text>
          )}
        </div>
      )}
      <div className="mt-12">
        <TextLink href="/organization/members">Members</TextLink>
      </div>
    </Section>,
  );
}
