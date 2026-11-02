import { PendingSubmit } from "@/components/controls/pending-submit";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Section } from "@/components/vestive/layout";
import { Heading, Text, TextLink } from "@/components/vestive/typography";
import type { GroupCard, GroupCards } from "@/lib/group-profile-client";

// G075 (Plan r2 §3.7, C8, C9, C10): the "Groups" section of the composite profile page. It renders exactly the cards
// core returned (nothing is filtered or added in the browser): one kit Card per Group with the Group name (a link to the
// Group page; core re-decides on click), the member's role label as escaped text or "No role given", and, only when
// core's per-card can_edit hint is true, a native POST form carrying this session's CSRF token, the card's optimistic
// version and the page's validated context selector (when set). The hint decides only whether the form is shown; every
// write is re-decided by core. Overflow (narrow_required) and an empty list are fixed texts with no number; outcome
// notices come from a closed table keyed by the page's group_status parameter, and nothing read from the request is
// reflected. After an invalid_context refusal the cards are not shown at all (the notice only; the request that named the
// bad context is not answered with an Organization-only view). No Group permission role, member list, count, email or Business Profile is ever shown. Type-only imports
// from the server client: the form action (lib labelFormAction) and the selector field (CONTEXT_FIELD) are spelled here.

type Param = string | string[] | undefined;

// The closed outcomes of the label form route (lib/group-profile-client.ts GROUP_STATUSES).
const STATUS: Readonly<Record<string, string>> = {
  label_saved: "The role was saved.",
  label_cleared: "The role was removed.",
  stale_version:
    "This role changed since the page was loaded, for example in another tab. Review the current value below and try again. Nothing was changed.",
  invalid_label: "That role was not saved. Use up to 100 characters without line breaks or control characters. Nothing was changed.",
  invalid_context:
    "That request named a Group context that is not valid, so it was refused. Nothing was changed. Open this profile again from a Group's members page.",
  not_permitted: "You cannot change that role. Only the member and that Group's owners and admins can. Nothing was changed.",
  unavailable: "The change could not be confirmed right now. The current roles are shown below; check them before trying again.",
};

export const SELECTED_CONTEXT_TEXT = "Showing this person's card for one Group.";
export const NO_CARDS_TEXT = "No Groups to show.";
export const NARROW_REQUIRED_TEXT =
  "Too many Groups to show at once. Open this profile from a Group's members page to see that Group's card.";
export const CARDS_UNAVAILABLE_TEXT = "The Groups of this profile cannot be shown right now. Please try again later.";

function Notice({ status }: { status: Param }) {
  if (typeof status !== "string" || !Object.hasOwn(STATUS, status)) return null;
  // A stale version is an alert (the kit Alert's own role); every other outcome is a polite status notice.
  if (status === "stale_version") {
    return (
      <Alert data-testid="group-label-alert" className="mt-10">
        <AlertDescription>{STATUS[status]}</AlertDescription>
      </Alert>
    );
  }
  return (
    <Alert role="status" data-testid="group-label-status" className="mt-10">
      <AlertDescription>{STATUS[status]}</AlertDescription>
    </Alert>
  );
}

type CardProps = { userId: string; card: GroupCard; contextGroup: string | undefined; csrfToken: string };

function LabelForm({ userId, card, contextGroup, csrfToken }: CardProps) {
  const inputId = `group-label-${card.groupId}`;
  const helpId = `group-label-help-${card.groupId}`;
  return (
    <form method="post" action={`/profiles/${userId}/groups/${card.groupId}/label`} className="mt-6 flex min-w-0 flex-col items-start gap-6" data-testid="group-label-form">
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <input type="hidden" name="expected_version" value={card.version} />
      {contextGroup !== undefined ? <input type="hidden" name="context_group" value={contextGroup} /> : null}
      <Field className="min-w-0">
        <FieldLabel htmlFor={inputId} className="max-w-full [overflow-wrap:anywhere]">
          Role in {card.groupName}
        </FieldLabel>
        <Input
          id={inputId}
          name="role_label"
          type="text"
          defaultValue={card.roleLabel ?? ""}
          maxLength={200} // UTF-16 units: core's authoritative limit is 100 code points (up to 2 units each)
          autoComplete="off"
          spellCheck={false}
          aria-describedby={helpId}
          data-testid="group-label-input"
        />
        <FieldDescription id={helpId}>Leave empty to remove the role</FieldDescription>
      </Field>
      <PendingSubmit label="Save role" pendingLabel="Saving…" data-testid="group-label-save" />
    </form>
  );
}

function GroupMemberCard({ userId, card, contextGroup, csrfToken }: CardProps) {
  return (
    <li className="min-w-0">
      <Card data-testid="group-card" data-group-id={card.groupId}>
        <CardHeader>
          <Heading level={3} className="min-w-0 [overflow-wrap:anywhere]">
            <TextLink href={`/groups/${card.groupId}`} className="max-w-full [overflow-wrap:anywhere]" data-testid="group-card-name">
              {card.groupName}
            </TextLink>
          </Heading>
        </CardHeader>
        <CardContent>
          <Text className="[overflow-wrap:anywhere]" data-testid="group-card-label">
            {card.roleLabel ?? "No role given"}
          </Text>
          {card.canEdit ? <LabelForm userId={userId} card={card} contextGroup={contextGroup} csrfToken={csrfToken} /> : null}
        </CardContent>
      </Card>
    </li>
  );
}

export type GroupCardsSectionProps = {
  userId: string;
  // The page's card read, or undefined when core could not answer (the unavailable notice; nothing is claimed).
  cards: GroupCards | undefined;
  // The page's validated selected-Group context (undefined: Organization-only).
  contextGroup: string | undefined;
  csrfToken: string;
  // The raw group_status parameter (validated here, never reflected).
  status: Param;
};

export function GroupCardsSection({ userId, cards, contextGroup, csrfToken, status }: GroupCardsSectionProps) {
  return (
    <Section id="groups" aria-labelledby="groups-title" tone="muted" data-testid="group-cards-section">
      <Heading level={2} id="groups-title">
        Groups
      </Heading>
      {contextGroup !== undefined ? (
        <Text className="mt-10" data-testid="group-cards-context">
          {SELECTED_CONTEXT_TEXT}
        </Text>
      ) : null}
      <Notice status={status} />
      {status === "invalid_context" ? null : cards === undefined ? (
        <Alert role="status" data-testid="group-cards-unavailable" className="mt-10">
          <AlertDescription>{CARDS_UNAVAILABLE_TEXT}</AlertDescription>
        </Alert>
      ) : cards.narrowRequired ? (
        <Text className="mt-10" data-testid="group-cards-narrow">
          {NARROW_REQUIRED_TEXT}
        </Text>
      ) : cards.cards.length === 0 ? (
        <Text className="mt-10" data-testid="group-cards-empty">
          {NO_CARDS_TEXT}
        </Text>
      ) : (
        <ul aria-labelledby="groups-title" className="m-0 mt-10 grid list-none gap-6 p-0" data-testid="group-cards">
          {cards.cards.map((card) => (
            <GroupMemberCard key={card.groupId} userId={userId} card={card} contextGroup={contextGroup} csrfToken={csrfToken} />
          ))}
        </ul>
      )}
    </Section>
  );
}
