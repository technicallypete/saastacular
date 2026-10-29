import { StatusMessage } from "@/components/account/status-message";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { Item, ItemActions, ItemContent, ItemDescription, ItemGroup } from "@/components/ui/item";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { isId, listProfiles, PAGE_SIZE } from "@/lib/onboarding-client";

// G018 (Plan r5 §1.5–§1.6): the Business Profiles index, as summaries, one keyset page at a time (a page may hold fewer
// rows than its size while a next page exists). Each opens the one shared profile page. Server-rendered and never cached.
// G060 (r1 C2): the listing is every profile core lets the viewer read in this host's Organization (metadata); core's
// can_update hint marks each row editable or read-only, and the shared page re-reads its own hint.
export const dynamic = "force-dynamic";

export default async function BusinessProfiles({ searchParams }: { searchParams: SearchParams }) {
  const r = await onboardingRequest({
    config: runtimeConfig(),
    area: "business-profiles",
    kicker: "Business Profiles",
    nav: [{ href: "#business-profiles", label: "Business Profiles" }],
  });
  if (!r.ready) return r.page;
  const after = param((await searchParams).after);
  const page = await listProfiles(r.lookup(), r.credential, r.origin, { after: isId(after) ? after : undefined, limit: PAGE_SIZE });
  if (!page.ok) return r.refused(page.error);
  const { profiles, next } = page.value;
  const first = !isId(after);

  return r.frame(
    <Section id="business-profiles" aria-labelledby="business-profiles-title">
      <Kicker>Business Profiles</Kicker>
      <Heading level={1} id="business-profiles-title" className="mt-6">
        Business Profiles
      </Heading>
      <Text className="mt-10">
        Business Profiles you can read in this Organization. A profile is shared: every Group that uses it shows the same
        details. Those marked read-only you can view but not change.
      </Text>
      {profiles.length === 0 ? (
        <StatusMessage
          message={first && next === null ? "There are no Business Profiles to show yet. Onboarding a Group creates one." : "There are no Business Profiles on this page."}
          className="mt-12"
        />
      ) : (
        <ItemGroup className="mt-12" data-testid="profile-list">
          {profiles.map((p) => (
            <Item key={p.id} variant="outline" role="listitem">
              {/* Layout only: names and industries wrap instead of the kit's line clamps. */}
              <ItemContent className="min-w-0">
                <Heading level={2} className="[overflow-wrap:anywhere]">
                  {p.displayName}
                </Heading>
                {p.industry === null ? null : (
                  <ItemDescription className="line-clamp-none [overflow-wrap:anywhere]">{p.industry}</ItemDescription>
                )}
                {p.canUpdate ? null : <ItemDescription data-testid="profile-read-only">Read-only</ItemDescription>}
              </ItemContent>
              <ItemActions>
                <TextLink href={`/business-profiles/${p.id}`} aria-label={`Open ${p.displayName}`}>
                  Open
                </TextLink>
              </ItemActions>
            </Item>
          ))}
        </ItemGroup>
      )}
      <div className="mt-10 flex flex-wrap gap-8">
        {first ? null : <TextLink href="/business-profiles">First page</TextLink>}
        {next === null ? null : <TextLink href={`/business-profiles?after=${next}`}>Next page</TextLink>}
      </div>
    </Section>,
  );
}
