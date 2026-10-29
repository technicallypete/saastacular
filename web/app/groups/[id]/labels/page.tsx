import { notFound } from "next/navigation";

import { StatusMessage } from "@/components/account/status-message";
import { LabelCreateForm } from "@/components/labels/label-form";
import { LabelList, ReadOnlyLabelList } from "@/components/labels/label-list";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { groupLabelsPath, isId, LABEL_STATUS, LABELS_NO_TARGETS, LABELS_NOT_ADMINISTERED, labelsPageError, listLabels } from "@/lib/labels-client";
import { getGroup } from "@/lib/onboarding-client";

// G020 (Plan r2 §1.5, D-7, D-8): one Group's own label definitions, which are Group content. Core's Group read gives the
// caller's actual Group role (self_role), which chooses the view; core decides every read and write on the exact Group:
// an owner or admin (label.group.manage) lists, creates (name and visibility), renames, changes visibility and deletes;
// a member (label.group.read) gets the read-only list of the definitions shared with members (target_members). A
// readable Group without a Group role (Organization owners and admins, with the Group-administration override on or off)
// gets a notice and no list: the definitions are never requested for them. Absent, foreign and unreadable Groups are the
// fixed 404. Server-rendered and never cached.
export const dynamic = "force-dynamic";

export default async function GroupLabelsPage({ params, searchParams }: { params: Promise<{ id: string }>; searchParams: SearchParams }) {
  const { id } = await params;
  if (!isId(id)) notFound();
  const r = await onboardingRequest({ config: runtimeConfig(), area: "groups", kicker: "Group labels", nav: [{ href: "#labels", label: "Labels" }] });
  if (!r.ready) return r.page;
  const g = await getGroup(r.lookup(), r.credential, r.origin, id);
  if (!g.ok) return r.refused(g.error);
  const group = g.value;
  const search = await searchParams;
  const path = groupLabelsPath(group.id);
  const heading = (
    <>
      <Kicker>Group labels</Kicker>
      <Heading level={1} id="labels-title" className="mt-6 [overflow-wrap:anywhere]">
        {`Labels of ${group.name}`}
      </Heading>
    </>
  );
  const links = (
    <div className="mt-12 flex flex-wrap gap-x-9 gap-y-3">
      <TextLink href={`/groups/${group.id}`}>Back to the Group</TextLink>
      <TextLink href="/groups">All Groups</TextLink>
    </div>
  );
  if (group.selfRole === null) {
    return r.frame(
      <Section id="labels" aria-labelledby="labels-title">
        {heading}
        <Text className="mt-10" data-testid="labels-not-administered">
          {LABELS_NOT_ADMINISTERED}
        </Text>
        {links}
      </Section>,
    );
  }
  const after = param(search.after);
  const list = await listLabels(r.lookup(), r.credential, r.origin, "group", group.id, isId(after) ? after : undefined);
  if (!list.ok) return r.refused(list.error === "forbidden" ? "not_found" : labelsPageError(list.error));
  const next = list.value.next === null ? undefined : (`${path}?after=${list.value.next}` as const);
  const manager = group.selfRole === "group_owner" || group.selfRole === "group_admin";
  const status = param(search.status);
  const notice = manager && status !== undefined && Object.hasOwn(LABEL_STATUS, status) ? LABEL_STATUS[status] : undefined;
  return r.frame(
    <Section id="labels" aria-labelledby="labels-title">
      {heading}
      {manager ? (
        <>
          <Text className="mt-10" data-testid="labels-no-targets">
            {LABELS_NO_TARGETS}
          </Text>
          {notice === undefined ? null : (
            <div data-testid="labels-status" className="mt-10">
              <StatusMessage message={notice} />
            </div>
          )}
          <LabelList scope="group" labels={list.value.labels} action={`${path}/action`} csrfToken={r.csrfToken} next={next} />
          <LabelCreateForm scope="group" action={`${path}/action`} csrfToken={r.csrfToken} />
        </>
      ) : (
        <ReadOnlyLabelList labels={list.value.labels} next={next} />
      )}
      {links}
    </Section>,
  );
}
