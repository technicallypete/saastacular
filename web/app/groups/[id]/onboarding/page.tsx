import { notFound } from "next/navigation";

import { StatusMessage } from "@/components/account/status-message";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { RadioChoice } from "@/components/controls/radio-choice";
import type { ShellNavItem } from "@/components/layout/app-shell";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { ProfileFields } from "@/components/onboarding/profile-fields";
import { StateBadge } from "@/components/onboarding/state-badge";
import { Card, CardContent } from "@/components/ui/card";
import { FieldLegend, FieldSet } from "@/components/ui/field";
import { RadioGroup } from "@/components/ui/radio-group";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { fieldError, getGroup, isId, listProfiles, ONBOARDING_NOT_ADMINISTERED, PAGE_SIZE, selectableProfiles } from "@/lib/onboarding-client";

// G018 (Plan r5 §1.6): onboard a Group with exactly one Business Profile through two explicit choices: create a new
// profile (the five fields) or select one of "My Business Profiles" (one keyset page of core's selectable summaries,
// native radios). Both post to the completion route, where core re-decides the choice on fresh facts. A completed Group
// shows its state instead of the forms. Server-rendered and never cached.
// G060 (r1 C2, C6; correction02 §P): a readable Group the viewer does not manage (core's can_manage hint) shows its
// metadata, state and a notice, never the forms (a forged completion is refused by core: group.complete_onboarding).
// The picker offers only the rows core marks can_select; the cursor and the First/Next links stay the unfiltered page's.
// G069 S12: the picker's options are a kit RadioGroup (one submitted business_profile_id, required) inside the kit
// FieldSet, on a kit Card.
export const dynamic = "force-dynamic";

const NAV = [
  { href: "#new-profile", label: "Create new" },
  { href: "#existing-profile", label: "Select existing" },
] as const;
// G060: the read-only view renders only its #new-profile section (metadata and the notice, no forms), so its Primary
// navigation offers only that section, under a label that promises no form.
const READ_ONLY_NAV = [{ href: "#new-profile", label: "Group onboarding" }] as const;

export default async function GroupOnboarding({ params, searchParams }: { params: Promise<{ id: string }>; searchParams: SearchParams }) {
  const { id } = await params;
  if (!isId(id)) notFound();
  // The frame renders this array when a view calls r.frame, so a view that renders fewer sections narrows it first.
  const nav: ShellNavItem[] = [...NAV];
  const r = await onboardingRequest({ config: runtimeConfig(), area: "groups", kicker: "Group onboarding", nav });
  if (!r.ready) return r.page;
  const g = await getGroup(r.lookup(), r.credential, r.origin, id);
  if (!g.ok) return r.refused(g.error);
  const group = g.value;
  const query = await searchParams;
  const action = `/groups/${group.id}/onboarding/complete`;

  if (group.complete) {
    return r.frame(
      <Section id="new-profile" aria-labelledby="onboarding-title">
        <Kicker>Group onboarding</Kicker>
        <Heading level={1} id="onboarding-title" className="mt-6 [overflow-wrap:anywhere]">
          {group.name}
        </Heading>
        <StatusMessage message="This Group's onboarding is already complete: it has a Business Profile." className="mt-10" />
        <div className="mt-10">
          <TextLink href={`/groups/${group.id}`}>Open the Group</TextLink>
        </div>
      </Section>,
    );
  }

  if (!group.canManage) {
    nav.splice(0, nav.length, ...READ_ONLY_NAV);
    return r.frame(
      <Section id="new-profile" aria-labelledby="onboarding-title">
        <Kicker>Group onboarding</Kicker>
        <Heading level={1} id="onboarding-title" className="mt-6 [overflow-wrap:anywhere]">
          {group.name}
        </Heading>
        <div className="mt-8 flex flex-wrap items-center gap-6">
          <StateBadge complete={group.complete} />
          <Text variant="meta" className="[overflow-wrap:anywhere]">
            {group.slug}
          </Text>
        </div>
        <Text className="mt-10" data-testid="onboarding-not-administered">
          {ONBOARDING_NOT_ADMINISTERED}
        </Text>
        <div className="mt-10">
          <TextLink href={`/groups/${group.id}`}>Open the Group</TextLink>
        </div>
      </Section>,
    );
  }

  const after = param(query.after);
  const page = await listProfiles(r.lookup(), r.credential, r.origin, { after: isId(after) ? after : undefined, limit: PAGE_SIZE });
  if (!page.ok) return r.refused(page.error);
  const { profiles, next } = selectableProfiles(page.value);
  const refusal = fieldError(param(query.error));
  const first = !isId(after);

  return r.frame(
    <>
      <Section id="new-profile" aria-labelledby="onboarding-title">
        <Kicker>Group onboarding</Kicker>
        <Heading level={1} id="onboarding-title" className="mt-6 [overflow-wrap:anywhere]">
          Onboard {group.name}
        </Heading>
        <Text className="mt-10">
          A Group is onboarded with one Business Profile. Create a new one below, or select one you can already use.
        </Text>
        <StatusMessage message={refusal === undefined ? undefined : "A value was not in the accepted form, so nothing was changed."} className="mt-10" />
        <Heading level={2} id="new-profile-title" className="mt-12">
          Create a new Business Profile
        </Heading>
        <form method="post" action={action} aria-labelledby="new-profile-title" className="mt-10 grid max-w-[48rem] gap-8" data-testid="profile-new-form">
          <input type="hidden" name="csrf_token" value={r.csrfToken} />
          <input type="hidden" name="mode" value="new" />
          <ProfileFields error={refusal} />
          <div>
            <PendingSubmit label="Create and attach" pendingLabel="Creating the Business Profile…" />
          </div>
        </form>
      </Section>
      <Section id="existing-profile" aria-labelledby="existing-profile-title" tone="muted">
        <Heading level={2} id="existing-profile-title">
          Select an existing Business Profile
        </Heading>
        {profiles.length === 0 ? (
          <StatusMessage
            message={first && next === null ? "No Business Profile you can attach. Create a new one above." : "There are no selectable Business Profiles on this page."}
            className="mt-10"
          />
        ) : (
          <form method="post" action={action} className="mt-10 grid max-w-[48rem] gap-8" data-testid="profile-existing-form">
            <input type="hidden" name="csrf_token" value={r.csrfToken} />
            <input type="hidden" name="mode" value="existing" />
            <Card className="min-w-0">
              <CardContent>
                <FieldSet className="min-w-0">
                  <FieldLegend id="existing-profile-legend" variant="label">
                    My Business Profiles
                  </FieldLegend>
                  <RadioGroup name="business_profile_id" required aria-labelledby="existing-profile-legend">
                    {profiles.map((p) => (
                      <RadioChoice key={p.id} value={p.id}>
                        {p.displayName}
                        {p.industry === null ? null : <span className="text-muted-foreground"> · {p.industry}</span>}
                      </RadioChoice>
                    ))}
                  </RadioGroup>
                </FieldSet>
              </CardContent>
            </Card>
            <div>
              <PendingSubmit label="Attach selected" pendingLabel="Attaching the Business Profile…" />
            </div>
          </form>
        )}
        <div className="mt-10 flex flex-wrap gap-8">
          {first ? null : <TextLink href={`/groups/${group.id}/onboarding#existing-profile`}>First page</TextLink>}
          {next === null ? null : <TextLink href={`/groups/${group.id}/onboarding?after=${next}#existing-profile`}>Next page</TextLink>}
        </div>
      </Section>
    </>,
  );
}
