import { notFound } from "next/navigation";

import { StatusMessage } from "@/components/account/status-message";
import { ButtonLink } from "@/components/controls/button-link";
import { GroupLabelsSection } from "@/components/labels/group-labels-section";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { StateBadge } from "@/components/onboarding/state-badge";
import { Card, CardContent, CardDescription, CardHeader } from "@/components/ui/card";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { GROUP_LABEL_STATUS, groupLabelsPath, labelOutcomeUnknown, listGroupLabels, listLabels, MAX_LABELS_PER_SCOPE, MAX_LABEL_PAGE, type Label } from "@/lib/labels-client";
import { selfMembership } from "@/lib/membership-client";
import { getGroup, getProfile, isId, ONBOARDING_NOT_ADMINISTERED } from "@/lib/onboarding-client";

// G018 (Plan r5 §1.6): one Group's onboarding view. Incomplete: the resume action. Complete: a summary of its Business
// Profile and the link to the shared profile page (Edit or View per core's can_update display hint; every update is
// re-decided by core). Unknown and foreign ids are the fixed 404. Server-rendered and never cached.
// G060 (Plan r0 §1.6; r1 C6): the Group is readable metadata for Organization owners and admins, so the page offers
// only what core's hints allow: Resume onboarding when can_manage (otherwise a notice), and the Members link for an
// actual Group role or a manager. Reading a Group never opens its content.
// G020 (Plan r2 §1.5, D-7): a Labels section rendered only when core answers the Group's Organization-label assignments
// (Organization owners and admins); a forbidden or not_found answer renders nothing at all (no heading, no hidden markup)
// and an unavailable one the unavailable notice without the section. The Group's own labels page is linked only for an
// actual Group role (self_role), never from the override hint.
export const dynamic = "force-dynamic";

const UNAVAILABLE = "This is unavailable right now. Please try again later.";

// Every Organization-owned definition (at most 200, in pages of 100) not yet assigned to this Group, or undefined when
// they cannot be listed right now.
async function unassignedLabels(r: { lookup: () => Parameters<typeof listLabels>[0]; credential: string; origin: string }, assigned: ReadonlySet<string>): Promise<Label[] | undefined> {
  const out: Label[] = [];
  let after: string | undefined;
  for (let i = 0; i <= MAX_LABELS_PER_SCOPE / MAX_LABEL_PAGE; i++) {
    const page = await listLabels(r.lookup(), r.credential, r.origin, "organization", undefined, after);
    if (!page.ok) return undefined;
    out.push(...page.value.labels.filter((l) => !assigned.has(l.id)));
    if (page.value.next === null) return out;
    after = page.value.next;
  }
  return out;
}

// The closed outcome statuses of the onboarding completion route. Both claim completion, so the page shows them only when
// core reports the Group complete: the URL marker alone never presents an incomplete Group as complete (Spec R2/AC5).
const STATUS: Readonly<Record<string, string>> = {
  completed: "Onboarding is complete: the Business Profile is attached to this Group.",
  already_complete: "This Group already has a Business Profile, so nothing was changed. Its current state is shown below.",
};

// G019 (Plan r1 §1.6): the closed outcomes of an invitation acceptance (POST /invitations/<id>/accept).
const JOINED: Readonly<Record<string, string>> = {
  joined: "You joined this Group.",
  upgraded: "Your role in this Group was raised by the invitation.",
  already_member: "You already have this role or a higher one in this Group, so nothing was changed.",
  consumed: "This invitation was already used, so nothing was changed.",
};

export default async function GroupPage({ params, searchParams }: { params: Promise<{ id: string }>; searchParams: SearchParams }) {
  const { id } = await params;
  if (!isId(id)) notFound();
  const r = await onboardingRequest({ config: runtimeConfig(), area: "groups", kicker: "Group", nav: [{ href: "#group", label: "Group" }] });
  if (!r.ready) return r.page;
  const g = await getGroup(r.lookup(), r.credential, r.origin, id);
  if (!g.ok) return r.refused(g.error);
  const group = g.value;
  const status = param((await searchParams).status);
  let message = group.complete && status !== undefined && Object.hasOwn(STATUS, status) ? STATUS[status] : undefined;
  // G019: an invitation acceptance outcome is shown only when core confirms the caller's membership of this Group.
  if (status !== undefined && Object.hasOwn(JOINED, status) && (await selfMembership(r.lookup(), r.credential, r.origin, { kind: "group", groupId: group.id })).ok) {
    message = JOINED[status];
  }
  const profile = group.businessProfileId === null ? undefined : await getProfile(r.lookup(), r.credential, r.origin, group.businessProfileId);
  const labels = await listGroupLabels(r.lookup(), r.credential, r.origin, group.id);
  const unassigned = labels.ok ? await unassignedLabels(r, new Set(labels.value.map((l) => l.labelId))) : undefined;
  const labelNotice = labels.ok && status !== undefined && Object.hasOwn(GROUP_LABEL_STATUS, status) ? GROUP_LABEL_STATUS[status] : undefined;

  return r.frame(
    <Section id="group" aria-labelledby="group-title">
      <Kicker>Group</Kicker>
      <Heading level={1} id="group-title" className="mt-6 [overflow-wrap:anywhere]">
        {group.name}
      </Heading>
      <div className="mt-8 flex flex-wrap items-center gap-6">
        <StateBadge complete={group.complete} />
        <Text variant="meta" className="[overflow-wrap:anywhere]">
          {group.slug}
        </Text>
      </div>
      <StatusMessage message={message} className="mt-10" />
      {!group.complete && !group.canManage ? (
        <Text className="mt-10" data-testid="onboarding-not-administered">
          {ONBOARDING_NOT_ADMINISTERED}
        </Text>
      ) : !group.complete ? (
        <>
          <Text className="mt-10">Attach a Business Profile to finish onboarding this Group: create a new one or select an existing one.</Text>
          <ButtonLink href={`/groups/${group.id}/onboarding`} className="mt-10">Resume onboarding</ButtonLink>
        </>
      ) : profile === undefined || !profile.ok ? (
        <StatusMessage message="The Business Profile of this Group cannot be shown right now." className="mt-10" />
      ) : (
        <Card className="mt-12" data-testid="group-profile">
          <CardHeader>
            <Heading level={2} className="[overflow-wrap:anywhere]">
              {profile.value.profile.displayName}
            </Heading>
            {profile.value.profile.industry === null ? null : (
              <CardDescription className="[overflow-wrap:anywhere]">{profile.value.profile.industry}</CardDescription>
            )}
          </CardHeader>
          <CardContent>
            <TextLink href={`/business-profiles/${profile.value.profile.id}`}>
              {profile.value.canUpdate ? "Edit Business Profile" : "View Business Profile"}
            </TextLink>
          </CardContent>
        </Card>
      )}
      {labels.ok ? (
        <GroupLabelsSection groupId={group.id} labels={labels.value} unassigned={unassigned} csrfToken={r.csrfToken} notice={labelNotice} />
      ) : labelOutcomeUnknown(labels.error) ? (
        <StatusMessage message={UNAVAILABLE} className="mt-10" />
      ) : null}
      <div className="mt-12 flex flex-wrap gap-x-9 gap-y-3">
        {group.canManage || group.selfRole !== null ? <TextLink href={`/groups/${group.id}/members`}>Members</TextLink> : null}
        {group.selfRole !== null ? (
          <TextLink href={groupLabelsPath(group.id)} data-testid="group-labels-link">
            Labels
          </TextLink>
        ) : null}
        <TextLink href="/groups">All Groups</TextLink>
      </div>
    </Section>,
  );
}
