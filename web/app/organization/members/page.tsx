import type { ReactNode } from "react";

import { membersView } from "@/components/membership/members-view";
import { onboardingRequest, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";

// G019 (Plan r1 §1.6; r2 addendum §7): the Organization members page. Every member sees their own membership card
// (role, Leave, the last-owner notice); administrators also see the members, pending invitations, the Invite form and,
// for an owner, Transfer ownership. Tenant hosts only; server-rendered and never cached.
export const dynamic = "force-dynamic";

export default async function OrganizationMembersPage({ searchParams }: { searchParams: SearchParams }) {
  const r = await onboardingRequest({ config: runtimeConfig(), area: "members", kicker: "Members", nav: [{ href: "#members", label: "Members" }] });
  if (!r.ready) return r.page;
  // G058 (Plan r3 §3): one Enrollment link below the members view (core decides what the enrollment page shows).
  const frame = (children: ReactNode) =>
    r.frame(
      <>
        {children}
        <div className="mx-auto w-full max-w-6xl px-4 py-10">
          <TextLink href="/organization/enrollment" data-testid="members-enrollment-link">
            Enrollment
          </TextLink>
        </div>
      </>,
    );
  return membersView({
    r: { ...r, frame },
    scope: { kind: "organization" },
    title: "Workspace members",
    where: "this workspace",
    path: "/organization/members",
    search: await searchParams,
    // G060 (Plan r0 §1.6): the Group-administration setting, linked for Organization owners and admins only.
    // G020 (Plan r2 §1.5): the Organization's labels, for the same audience.
    staffLinks: (
      <div className="flex flex-wrap gap-x-9 gap-y-3">
        <TextLink href="/organization/group-authority" data-testid="members-group-authority-link">
          Group administration
        </TextLink>
        <TextLink href="/organization/labels" data-testid="members-labels-link">
          Labels
        </TextLink>
      </div>
    ),
  });
}
