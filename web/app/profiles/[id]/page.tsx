import { notFound } from "next/navigation";

import type { ShellNavItem } from "@/components/layout/app-shell";
import { onboardingRequest, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { GroupCardsSection } from "@/components/profile/group-cards-section";
import { ProfileHeader } from "@/components/profile/profile-header";
import { runtimeConfig } from "@/lib/config";
import { listGroupCards, readContextParam } from "@/lib/group-profile-client";
import type { OnboardingError } from "@/lib/onboarding-client";
import { getProfile, isProfileId, type ProfileError } from "@/lib/profile-client";

// G075 (Plan r2 §3.5, §3.7, C8, C9, C10): the composite profile page of one user in this workspace. The header is G074's
// global name and avatar as core's relationship predicate shows them to the viewer now; the "Groups" section lists the
// subject's Group member cards the viewer may see now (core's projection decides every card), with an inline label form
// on each card core marks editable. ?context=<Group id> opens the page in that Group's selected context (from a Group's
// members page): core's Resolve decides selectability and the selector only narrows the cards; without it the context is
// Organization-only. An empty, malformed or repeated ?context= is the fixed 404 (never dropped to Organization-only), and
// the validated selector is carried by every label form. An unknown, foreign, unrelated or revoked subject and an
// unselectable context are the same fixed 404. Tenant hosts only; server-rendered and never cached.
export const dynamic = "force-dynamic";

const NAV: ShellNavItem[] = [
  { href: "#profile", label: "Profile" },
  { href: "#groups", label: "Groups" },
];

// The frame's fixed pages for a refused profile read: no session is sign-in-first; every refusal of the subject or the
// context is the one fixed 404 (never a distinguishable forbidden); anything else is the unavailable page.
function frameError(error: ProfileError): OnboardingError {
  switch (error) {
    case "unauthenticated":
      return "unauthenticated";
    case "not_found":
    case "forbidden":
    case "not_available":
    case "origin_rejected":
    case "unknown_host":
    case "invalid_request":
      return "not_found";
    default:
      return "onboarding_unavailable";
  }
}

export default async function MemberProfilePage({ params, searchParams }: { params: Promise<{ id: string }>; searchParams: SearchParams }) {
  const { id } = await params;
  if (!isProfileId(id)) notFound();
  const search = await searchParams;
  const selector = readContextParam(search.context);
  if (!selector.ok) notFound();
  const r = await onboardingRequest({ config: runtimeConfig(), area: "members", kicker: "Profile", nav: NAV });
  if (!r.ready) return r.page;
  const [header, cards] = await Promise.all([
    getProfile(r.lookup(), r.credential, r.origin, id),
    listGroupCards(r.lookup(), r.credential, r.origin, id, selector.group),
  ]);
  if (!header.ok) return r.refused(frameError(header.error));
  // A refused card read (an unselectable context, or a subject no longer visible) is the page's refusal; a card read core
  // could not answer leaves the header and shows the unavailable notice in the Groups section.
  if (!cards.ok && frameError(cards.error) !== "onboarding_unavailable") return r.refused(frameError(cards.error));
  return r.frame(
    <>
      <ProfileHeader profile={header.value} />
      <GroupCardsSection
        userId={id}
        cards={cards.ok ? cards.value : undefined}
        contextGroup={selector.group}
        csrfToken={r.csrfToken}
        status={search.group_status}
      />
    </>,
  );
}
