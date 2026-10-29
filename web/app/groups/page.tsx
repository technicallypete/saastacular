import { StatusMessage } from "@/components/account/status-message";
import { ButtonLink } from "@/components/controls/button-link";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { StateBadge } from "@/components/onboarding/state-badge";
import { Item, ItemActions, ItemContent, ItemDescription, ItemGroup } from "@/components/ui/item";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { isId, listGroups, ONBOARDING_NOT_ADMINISTERED, PAGE_SIZE } from "@/lib/onboarding-client";

// G018 (Plan r5 §1.6): the Groups of this host's Organization that core lets the viewer read, one keyset page at a time
// in core's order, each with its derived onboarding state and its next action. Server-rendered and never cached.
// G060 (r1 C6): Organization owners and admins read every Group's metadata; "Resume onboarding" appears only on a Group
// core says the viewer manages (can_manage), any other incomplete Group shows the notice and an Open link.
export const dynamic = "force-dynamic";

export default async function Groups({ searchParams }: { searchParams: SearchParams }) {
  const r = await onboardingRequest({ config: runtimeConfig(), area: "groups", kicker: "Groups", nav: [{ href: "#groups", label: "Groups" }] });
  if (!r.ready) return r.page;
  const params = await searchParams;
  const after = param(params.after);
  const page = await listGroups(r.lookup(), r.credential, r.origin, { after: isId(after) ? after : undefined, limit: PAGE_SIZE });
  if (!page.ok) return r.refused(page.error);
  const { groups, next } = page.value;
  const first = !isId(after);

  return r.frame(
    <Section id="groups" aria-labelledby="groups-title">
      <Kicker>Groups</Kicker>
      <Heading level={1} id="groups-title" className="mt-6">
        Your Groups
      </Heading>
      <Text className="mt-10">
        A Group is where people work together. Each Group is onboarded with one Business Profile: a new one, or one you can
        already select.
      </Text>
      <ButtonLink href="/groups/new" className="mt-10">New Group</ButtonLink>
      {groups.length === 0 ? (
        <StatusMessage message={first ? "There are no Groups to show yet." : "There are no more Groups."} className="mt-12" />
      ) : (
        <ItemGroup className="mt-12" data-testid="group-list">
          {groups.map((g) => (
            <Item key={g.id} variant="outline" role="listitem">
              {/* Layout only: names, addresses and the notice wrap instead of the kit's line clamps. */}
              <ItemContent className="min-w-0">
                <Heading level={2} className="[overflow-wrap:anywhere]">
                  {g.name}
                </Heading>
                <ItemDescription className="line-clamp-none [overflow-wrap:anywhere]">{g.slug}</ItemDescription>
                {g.complete || g.canManage ? null : (
                  <ItemDescription className="line-clamp-none" data-testid="onboarding-not-administered">
                    {ONBOARDING_NOT_ADMINISTERED}
                  </ItemDescription>
                )}
              </ItemContent>
              <ItemActions className="flex-wrap">
                <StateBadge complete={g.complete} />
                {g.complete || !g.canManage ? (
                  <TextLink href={`/groups/${g.id}`} aria-label={`Open ${g.name}`}>
                    Open
                  </TextLink>
                ) : (
                  <TextLink href={`/groups/${g.id}/onboarding`} aria-label={`Resume onboarding for ${g.name}`}>
                    Resume onboarding
                  </TextLink>
                )}
              </ItemActions>
            </Item>
          ))}
        </ItemGroup>
      )}
      <div className="mt-10 flex flex-wrap gap-8">
        {first ? null : <TextLink href="/groups">First page</TextLink>}
        {next === null ? null : <TextLink href={`/groups?after=${next}`}>Next page</TextLink>}
      </div>
    </Section>,
  );
}
