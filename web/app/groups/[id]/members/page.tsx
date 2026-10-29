import { notFound } from "next/navigation";

import { membersView } from "@/components/membership/members-view";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { isId } from "@/lib/membership-client";
import { getGroup } from "@/lib/onboarding-client";

// G019 (Plan r1 §1.6): the Group members page, the Organization page's twin for one Group. The Group's name comes from
// core's Group read; after the caller leaves, that read is refused and the closed "left" status is shown instead.
// Unknown and foreign ids are the fixed 404. Server-rendered and never cached. G060 (Plan r0 §1.6; r1 C7): the same
// read supplies core's self_role, can_manage and management_source hints, which choose the view (members-view.tsx).
export const dynamic = "force-dynamic";

export default async function GroupMembersPage({ params, searchParams }: { params: Promise<{ id: string }>; searchParams: SearchParams }) {
  const { id } = await params;
  if (!isId(id)) notFound();
  const r = await onboardingRequest({ config: runtimeConfig(), area: "groups", kicker: "Group members", nav: [{ href: "#members", label: "Members" }] });
  if (!r.ready) return r.page;
  const search = await searchParams;
  const g = await getGroup(r.lookup(), r.credential, r.origin, id);
  if (!g.ok && param(search.status) !== "left") return r.refused(g.error);
  return membersView({
    r,
    scope: { kind: "group", groupId: id },
    title: g.ok ? `Members of ${g.value.name}` : "Group members",
    where: "this Group",
    path: `/groups/${id}/members`,
    search,
    back: g.ok ? <TextLink href={`/groups/${id}`}>Back to the Group</TextLink> : <TextLink href="/groups">All Groups</TextLink>,
    authority: g.ok ? { selfRole: g.value.selfRole, canManage: g.value.canManage, managementSource: g.value.managementSource } : undefined,
  });
}
