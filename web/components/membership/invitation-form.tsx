import type { ReactNode } from "react";

import { FormSelect } from "@/components/controls/form-select";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { Input } from "@/components/ui/input";
import { Heading, Text, TextLink } from "@/components/vestive/typography";
import { DEFAULT_LIFETIME_HOURS, LIFETIME_HOURS, ROLES, type Candidate, type ScopeKind } from "@/lib/membership-client";

import { memberName, roleOptions } from "./member-table";

// G019 (Plan r1 §1.6): the Invite form. An Organization invitation is addressed to an email address; a Group invitation
// to one of the current Organization members core lists as candidates (ids and display names only, D-13). Roles are the
// scope's admin and member roles; the expiry is one of three closed periods. Core validates and decides everything.

type Props = { kind: ScopeKind; action: `/${string}`; csrfToken: string; candidates?: Candidate[]; candidatesNext?: `/${string}` };

const LIFETIME_LABELS: Readonly<Record<(typeof LIFETIME_HOURS)[number], string>> = { 24: "1 day", 168: "7 days", 720: "30 days" };

// G069 r3 (owner finding): every dropdown is the kit Select through FormSelect, at full field width with the app's
// solid field boundary. The Group recipient defaults to the first candidate, as the native options did.
const LIFETIME_OPTIONS = LIFETIME_HOURS.map((h) => ({ value: String(h), label: LIFETIME_LABELS[h] }));

function Field({ id, label, children }: { id: string; label: string; children: ReactNode }) {
  return (
    <div className="grid min-w-0 gap-2">
      <label htmlFor={id} className="font-sans text-sm font-medium text-foreground">
        {label}
      </label>
      {children}
    </div>
  );
}

export function InvitationForm({ kind, action, csrfToken, candidates = [], candidatesNext }: Props) {
  const groupWithoutCandidates = kind === "group" && candidates.length === 0;
  return (
    <div className="mt-12 grid gap-6" data-testid="invitation-form">
      <Heading level={2} id="invite-title">
        Invite
      </Heading>
      {groupWithoutCandidates ? (
        <Text>Every current member of this workspace is already in this Group, or none can be listed right now.</Text>
      ) : (
        <form method="post" action={action} aria-labelledby="invite-title" className="grid max-w-xl gap-6">
          <input type="hidden" name="csrf_token" value={csrfToken} />
          <input type="hidden" name="action" value="invite" />
          {kind === "organization" ? (
            <Field id="invite-address" label="Email address">
              <Input id="invite-address" name="intended_address" type="email" required autoComplete="off" spellCheck={false} />
            </Field>
          ) : (
            <Field id="invite-recipient" label="Workspace member">
              <FormSelect id="invite-recipient" name="intended_user_id" required options={candidates.map((c) => ({ value: c.userId, label: memberName(c) }))} />
            </Field>
          )}
          <Field id="invite-role" label="Role">
            <FormSelect id="invite-role" name="role" defaultValue={ROLES[kind].member} required options={roleOptions(kind)} />
          </Field>
          <Field id="invite-expiry" label="Expires after">
            <FormSelect id="invite-expiry" name="expires_in_hours" defaultValue={String(DEFAULT_LIFETIME_HOURS)} required options={LIFETIME_OPTIONS} />
          </Field>
          <Text variant="meta">The next page shows the invitation link once. Copy it before leaving that page.</Text>
          <div>
            <PendingSubmit label="Create invitation" pendingLabel="Creating…" />
          </div>
        </form>
      )}
      {candidatesNext === undefined ? null : (
        <div>
          <TextLink href={candidatesNext}>More workspace members</TextLink>
        </div>
      )}
    </div>
  );
}
