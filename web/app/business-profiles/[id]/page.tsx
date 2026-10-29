import { notFound } from "next/navigation";

import { StatusMessage } from "@/components/account/status-message";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { ProfileFields, VALUE_OF } from "@/components/onboarding/profile-fields";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { FIELD_LABELS, fieldError, getProfile, isId, PROFILE_FIELDS } from "@/lib/onboarding-client";

// G018 (Plan r5 §1.5–§1.6): the one shared Business Profile page. Core's can_update display hint chooses the editor or
// the read-only view; every update is re-decided by core on fresh facts. Saving replaces the whole record and the last
// save wins, which the page discloses. The page shows the profile only: never its Groups, sibling Groups or people.
// Unknown and foreign ids are the fixed 404. Server-rendered and never cached.
export const dynamic = "force-dynamic";

const STATUS: Readonly<Record<string, string>> = {
  updated: "The Business Profile was saved. Every Group that uses it shows these details now.",
};

export default async function BusinessProfilePage({ params, searchParams }: { params: Promise<{ id: string }>; searchParams: SearchParams }) {
  const { id } = await params;
  if (!isId(id)) notFound();
  const r = await onboardingRequest({
    config: runtimeConfig(),
    area: "business-profiles",
    kicker: "Business Profile",
    nav: [{ href: "#business-profile", label: "Business Profile" }],
  });
  if (!r.ready) return r.page;
  const read = await getProfile(r.lookup(), r.credential, r.origin, id);
  if (!read.ok) return r.refused(read.error);
  const { profile, canUpdate } = read.value;
  const query = await searchParams;
  const status = param(query.status);
  const refusal = canUpdate ? fieldError(param(query.error)) : undefined;
  const message =
    refusal !== undefined
      ? "A value was not in the accepted form, so nothing was changed."
      : status !== undefined && Object.hasOwn(STATUS, status)
        ? STATUS[status]
        : undefined;

  return r.frame(
    <Section id="business-profile" aria-labelledby="business-profile-title">
      <Kicker>Business Profile</Kicker>
      <Heading level={1} id="business-profile-title" className="mt-6 [overflow-wrap:anywhere]">
        {profile.displayName}
      </Heading>
      <StatusMessage message={message} className="mt-10" />
      {canUpdate ? (
        <>
          <Text className="mt-10" data-testid="last-save-wins">
            This profile is shared by every Group that uses it. Saving replaces all of its details; if someone else saves
            after you loaded this page, the last save is kept.
          </Text>
          <form method="post" action={`/business-profiles/${profile.id}/update`} className="mt-12 grid max-w-[48rem] gap-8" data-testid="profile-edit-form">
            <input type="hidden" name="csrf_token" value={r.csrfToken} />
            <ProfileFields values={profile} error={refusal} />
            <div>
              <PendingSubmit label="Save" pendingLabel="Saving the Business Profile…" />
            </div>
          </form>
        </>
      ) : (
        <>
          <Text className="mt-10">You can view this Business Profile. Changing it is not available for your role.</Text>
          <dl className="mt-12 grid max-w-[48rem] gap-6" data-testid="profile-view">
            {PROFILE_FIELDS.map((name) => {
              const value = VALUE_OF[name](profile);
              return (
                <div key={name} className="grid gap-1 border-b border-border pb-4">
                  <dt className="font-sans text-sm font-medium text-foreground/80">{FIELD_LABELS[name]}</dt>
                  <dd className="m-0 font-sans text-sm whitespace-pre-wrap text-foreground [overflow-wrap:anywhere]">
                    {value ?? <span className="text-muted-foreground">Not provided</span>}
                  </dd>
                </div>
              );
            })}
          </dl>
        </>
      )}
      <div className="mt-12">
        <TextLink href="/business-profiles">All Business Profiles</TextLink>
      </div>
    </Section>,
  );
}
