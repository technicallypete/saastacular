import { StatusMessage } from "@/components/account/status-message";
import { LabelCreateForm } from "@/components/labels/label-form";
import { LabelList } from "@/components/labels/label-list";
import { asOnboardingError } from "@/components/membership/members-view";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { isId, LABEL_STATUS, labelsPageError, listLabels, ORGANIZATION_LABELS_PATH } from "@/lib/labels-client";
import { ROLES, selfMembership } from "@/lib/membership-client";

// G020 (Plan r2 §1.5, D-7): the Organization's own label definitions, for its owners and admins (core decides every read
// and write: label.organization.manage). The caller's own Organization role chooses the view, as on the
// Group-administration page: an owner or admin gets the list with rename and delete and the one create form (name only;
// the visibility is fixed to Organization administrators and stated); anyone else gets the fixed 404. Deleting an
// Organization label under a selected Group is refused by core (organization_context_required, a closed status; nothing
// is deleted). Tenant hosts only; server-rendered and never cached.
export const dynamic = "force-dynamic";

export default async function OrganizationLabelsPage({ searchParams }: { searchParams: SearchParams }) {
  const r = await onboardingRequest({ config: runtimeConfig(), area: "members", kicker: "Labels", nav: [{ href: "#labels", label: "Labels" }] });
  if (!r.ready) return r.page;
  const search = await searchParams;
  const after = param(search.after);
  const self = await selfMembership(r.lookup(), r.credential, r.origin, { kind: "organization" });
  if (!self.ok) return r.refused(asOnboardingError(self.error));
  if (self.value.role !== ROLES.organization.owner && self.value.role !== ROLES.organization.admin) return r.refused("not_found");
  const list = await listLabels(r.lookup(), r.credential, r.origin, "organization", undefined, isId(after) ? after : undefined);
  if (!list.ok) return r.refused(list.error === "forbidden" ? "not_found" : labelsPageError(list.error));
  const status = param(search.status);
  const notice = status !== undefined && Object.hasOwn(LABEL_STATUS, status) ? LABEL_STATUS[status] : undefined;
  const next = list.value.next === null ? undefined : (`${ORGANIZATION_LABELS_PATH}?after=${list.value.next}` as const);
  return r.frame(
    <Section id="labels" aria-labelledby="labels-title">
      <Kicker>Labels</Kicker>
      <Heading level={1} id="labels-title" className="mt-6">
        Workspace labels
      </Heading>
      <Text className="mt-10">
        Organization labels mark Groups for this workspace&apos;s owners and admins. They are visible to Organization administrators
        only and are never shown to Group members.
      </Text>
      {notice === undefined ? null : (
        <div data-testid="labels-status" className="mt-10">
          <StatusMessage message={notice} />
        </div>
      )}
      <LabelList scope="organization" labels={list.value.labels} action={`${ORGANIZATION_LABELS_PATH}/action`} csrfToken={r.csrfToken} next={next} />
      <LabelCreateForm scope="organization" action={`${ORGANIZATION_LABELS_PATH}/action`} csrfToken={r.csrfToken} />
      <div className="mt-12 flex flex-wrap gap-x-9 gap-y-3">
        <TextLink href="/organization/members">Members</TextLink>
        <TextLink href="/groups">Groups</TextLink>
      </div>
    </Section>,
  );
}
