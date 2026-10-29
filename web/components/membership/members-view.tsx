import type { ReactNode } from "react";

import { StatusMessage } from "@/components/account/status-message";
import { param, type OnboardingRequest } from "@/components/onboarding/onboarding-frame";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import {
  ACTION_STATUS,
  invitationCandidates,
  isId,
  listInvitations,
  listMembers,
  OWNER_CANDIDATE_CURSOR,
  OWNER_CANDIDATE_PAGE,
  ownerCandidatesNext,
  ROLES,
  selfMembership,
  type MemberScope,
  type MembershipError,
  type MembershipResult,
  type SelfMember,
} from "@/lib/membership-client";
import type { GroupView, OnboardingError } from "@/lib/onboarding-client";

import { InvitationForm } from "./invitation-form";
import { InvitationList } from "./invitation-list";
import { MemberTable } from "./member-table";
import { SelfCard, TransferForm } from "./ownership-forms";

// G019 (Plan r1 §1.6; r2 addendum §7): the shared body of the Organization and Group members pages. It reads the
// caller's own membership first and renders the self card for every member; only when core's can_manage hint is true
// does it read the administrative lists (members, pending invitations, owner candidates, Group candidates), so a plain
// member never issues a denied request and never sees the directory. Every read and action is decided by core.
// G060 (Plan r0 §1.6, D-10; r1 C7): on a Group the page first takes core's hints from the Group read (`authority`). The
// own-membership read and its self card, Leave and last-owner notice exist only for an actual Group role (self_role);
// an Organization owner or admin without one reads the Group's metadata but is not a member: with the owner-set
// override they administer its members through it (an explicit notice, no self card, no Leave, no ownership control,
// which core's per-row hints already deny), and without it the page says the members are not administered by them
// (no table, no form). An actual member who also administers through the override is told so in one line. Absent,
// foreign and unreadable Groups keep the fixed 404 (the page's Group read).

type Ready = Extract<OnboardingRequest, { ready: true }>;
type Search = Record<string, string | string[] | undefined>;
// The G060 display hints of the Group read (onboarding-client GroupView).
export type GroupAuthorityHints = Pick<GroupView, "selfRole" | "canManage" | "managementSource">;

export const OVERRIDE_ADMIN_NOTICE =
  "You administer this Group's members through the Organization setting Group administration by Organization staff. You are not a member of this Group.";
export const OVERRIDE_MEMBER_NOTICE = "You also administer this Group's members through the Organization setting.";
export const NOT_ADMINISTERED_NOTICE =
  "You can see this Group as an owner or admin of this workspace, but you are not a member of it. Only this Group's own owners and admins can see and manage its members.";

// The frame's fixed refusal pages take the onboarding codes; membership's own unavailable code is the same case.
export function asOnboardingError(e: MembershipError): OnboardingError {
  return e === "membership_unavailable" ? "onboarding_unavailable" : e;
}

function cursor(search: Search, name: string): string | undefined {
  const v = param(search[name]);
  return isId(v) ? v : undefined;
}

type Props = {
  r: Ready;
  scope: MemberScope;
  title: string;
  where: string;
  path: `/${string}`;
  search: Search;
  back?: ReactNode;
  // A Group's hints from the page's Group read (G060); absent on the Organization page.
  authority?: GroupAuthorityHints;
  // Links shown below the page to an Organization owner or admin only (G060: the Group-administration setting).
  staffLinks?: ReactNode;
};

export async function membersView({ r, scope, title, where, path, search, back, authority, staffLinks }: Props): Promise<ReactNode> {
  const status = param(search.status);
  // G060: no own-membership read without an actual Group role (an override is never a self membership).
  const nonMember = scope.kind === "group" && authority !== undefined && authority.selfRole === null;
  const self: MembershipResult<SelfMember> | undefined = nonMember ? undefined : await selfMembership(r.lookup(), r.credential, r.origin, scope);
  const heading = (
    <>
      <Kicker>Members</Kicker>
      <Heading level={1} id="members-title" className="mt-6 [overflow-wrap:anywhere]">
        {title}
      </Heading>
    </>
  );
  if (self !== undefined && !self.ok) {
    // After a successful leave the caller is no longer a member: the closed "left" status is shown instead of a refusal.
    if (status === "left" && (self.error === "not_found" || self.error === "not_available" || self.error === "forbidden")) {
      return r.frame(
        <Section id="members" aria-labelledby="members-title">
          {heading}
          <StatusMessage message={`You have left ${where}.`} className="mt-10" />
        </Section>,
      );
    }
    return r.refused(asOnboardingError(self.error));
  }
  const member = self?.ok ? self.value : undefined;
  const administers = member !== undefined ? member.canManage : authority?.canManage === true;
  if (member === undefined && !administers) {
    // G060 (D-10): a readable Group (metadata) whose members this caller does not administer: a truthful notice, no
    // table, no form and no self card; after a Leave the closed "left" status is shown instead.
    return r.frame(
      <Section id="members" aria-labelledby="members-title">
        {heading}
        {status === "left" ? (
          <StatusMessage message={`You have left ${where}.`} className="mt-10" />
        ) : (
          <Text className="mt-10" data-testid="members-not-administered">
            {NOT_ADMINISTERED_NOTICE}
          </Text>
        )}
        {back === undefined ? null : <div className="mt-12">{back}</div>}
      </Section>,
    );
  }
  const viaOverride = scope.kind === "group" && authority?.managementSource === "organization_override";
  const action: `/${string}` = `${path}/action`;
  const message = status !== undefined && Object.hasOwn(ACTION_STATUS, status) ? ACTION_STATUS[status] : undefined;
  let admin: ReactNode = null;
  if (administers) {
    // Ownership actions are own-only in core: only an actual owner reads the owner candidates.
    const owner = member !== undefined && member.role === ROLES[scope.kind].owner;
    const [members, invitations, owners, candidates] = await Promise.all([
      listMembers(r.lookup(), r.credential, r.origin, scope, { after: cursor(search, "after") }),
      listInvitations(r.lookup(), r.credential, r.origin, scope, { after: cursor(search, "invitations_after") }),
      owner
        ? listMembers(r.lookup(), r.credential, r.origin, scope, {
            ownerCandidates: true,
            limit: OWNER_CANDIDATE_PAGE,
            after: cursor(search, OWNER_CANDIDATE_CURSOR),
          })
        : undefined,
      scope.kind === "group" ? invitationCandidates(r.lookup(), r.credential, r.origin, scope.groupId, { after: cursor(search, "candidates_after") }) : undefined,
    ]);
    const cand = candidates !== undefined && candidates.ok ? candidates.value : undefined;
    const unavailable = (what: string) => <StatusMessage message={`The ${what} cannot be shown right now.`} className="mt-8" />;
    admin = (
      <>
        <Heading level={2} id="member-list-title" className="mt-12">
          Members
        </Heading>
        {members.ok ? (
          <>
            <MemberTable kind={scope.kind} members={members.value.members} action={action} csrfToken={r.csrfToken} />
            {members.value.next === null ? null : (
              <div className="mt-6">
                <TextLink href={`${path}?after=${members.value.next}`}>Next members</TextLink>
              </div>
            )}
          </>
        ) : (
          unavailable("member list")
        )}
        <Heading level={2} id="invitation-list-title" className="mt-12">
          Invitations
        </Heading>
        {invitations.ok ? (
          <>
            <InvitationList invitations={invitations.value.invitations} action={action} csrfToken={r.csrfToken} />
            {invitations.value.next === null ? null : (
              <div className="mt-6">
                <TextLink href={`${path}?invitations_after=${invitations.value.next}`}>Next invitations</TextLink>
              </div>
            )}
          </>
        ) : (
          unavailable("invitation list")
        )}
        {candidates !== undefined && !candidates.ok ? (
          unavailable("list of workspace members to invite")
        ) : (
          <InvitationForm
            kind={scope.kind}
            action={action}
            csrfToken={r.csrfToken}
            candidates={cand?.candidates}
            candidatesNext={cand === undefined || cand.next === null ? undefined : `${path}?candidates_after=${cand.next}`}
          />
        )}
        {owners === undefined ? null : owners.ok ? (
          <TransferForm
            kind={scope.kind}
            action={action}
            csrfToken={r.csrfToken}
            candidates={owners.value.members}
            candidatesNext={ownerCandidatesNext(path, owners.value.next)}
            laterPage={cursor(search, OWNER_CANDIDATE_CURSOR) !== undefined ? `${path}#transfer-title` : undefined}
          />
        ) : (
          unavailable("owner candidates")
        )}
      </>
    );
  }
  const staff = member !== undefined && scope.kind === "organization" && (member.role === ROLES.organization.owner || member.role === ROLES.organization.admin);
  return r.frame(
    <Section id="members" aria-labelledby="members-title">
      {heading}
      <StatusMessage message={message} className="mt-10" />
      {member === undefined ? (
        <Text className="mt-10" data-testid="override-admin-notice">
          {OVERRIDE_ADMIN_NOTICE}
        </Text>
      ) : (
        <SelfCard kind={scope.kind} action={action} csrfToken={r.csrfToken} self={member} where={where} />
      )}
      {member !== undefined && administers && viaOverride ? (
        <Text className="mt-10" data-testid="override-admin-member-notice">
          {OVERRIDE_MEMBER_NOTICE}
        </Text>
      ) : null}
      {admin}
      {administers ? null : <Text className="mt-10">Only the administrators of {where} can see and manage its members.</Text>}
      {staff && staffLinks !== undefined ? <div className="mt-12">{staffLinks}</div> : null}
      {back === undefined ? null : <div className="mt-12">{back}</div>}
    </Section>,
  );
}
