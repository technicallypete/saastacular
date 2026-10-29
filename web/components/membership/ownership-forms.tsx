import { FormSelect } from "@/components/controls/form-select";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { Heading, Text, TextLink } from "@/components/vestive/typography";
import { ROLE_LABELS, ROLES, type Member, type ScopeKind, type SelfMember } from "@/lib/membership-client";

import { memberName, roleOptions } from "./member-table";

// G069 r3 (owner finding): both dropdowns are the kit Select through FormSelect, at full field width with the app's
// solid field boundary. The recipient defaults to the first eligible admin, as the native options did.

// G019 (Plan r1 §1.6; r2 addendum §7): the caller's own card (role, Leave, the truthful last-owner notice) shown to every
// member, and the owner's Transfer form fed by core's owner-candidate page. Hints only; core re-decides each action.

type Base = { kind: ScopeKind; action: `/${string}`; csrfToken: string };

export function SelfCard({ kind, action, csrfToken, self, where }: Base & { self: SelfMember; where: string }) {
  return (
    <div className="mt-10 grid gap-6 rounded-2xl border border-border p-6" data-testid="self-card">
      <Heading level={2} id="self-title">
        Your membership
      </Heading>
      <Text>{`Your role in ${where}: ${ROLE_LABELS[self.role]}`}</Text>
      {self.lastOwner ? (
        <Text data-testid="last-owner-notice">
          {`You are the last owner of ${where}, so you cannot leave yet. Make another member an owner or transfer ownership first.`}
        </Text>
      ) : self.canLeave ? (
        <form method="post" action={action} className="grid gap-4">
          <input type="hidden" name="csrf_token" value={csrfToken} />
          <input type="hidden" name="action" value="leave" />
          <Text variant="meta">{`Leaving removes your access to ${where}. You need a new invitation to return.`}</Text>
          <div>
            <PendingSubmit label="Leave" pendingLabel="Leaving…" variant="outline" />
          </div>
        </form>
      ) : (
        <Text>{`Leaving ${where} is not available for your membership here.`}</Text>
      )}
      {kind === "organization" ? null : <Text variant="meta">Your role in the workspace itself is unchanged by Group changes.</Text>}
    </div>
  );
}

// candidates is one bounded page of core's eligible admins; candidatesNext links to the following page (core's cursor)
// and laterPage back to the first, so every eligible admin is reachable without loading the whole list.
type TransferProps = Base & { candidates: Member[]; candidatesNext?: `/${string}`; laterPage?: `/${string}` };

export function TransferForm({ kind, action, csrfToken, candidates, candidatesNext, laterPage }: TransferProps) {
  const eligible = candidates.filter((c) => c.canTransfer && !c.self);
  const paged = candidatesNext !== undefined || laterPage !== undefined;
  return (
    <div className="mt-12 grid gap-6" data-testid="transfer-form">
      <Heading level={2} id="transfer-title">
        Transfer ownership
      </Heading>
      {eligible.length === 0 ? (
        <Text>
          {paged ? "No admin on this page can receive ownership." : "No member can receive ownership right now. Only an admin can become an owner."}
        </Text>
      ) : (
        <form method="post" action={action} aria-labelledby="transfer-title" className="grid max-w-xl gap-6">
          <input type="hidden" name="csrf_token" value={csrfToken} />
          <input type="hidden" name="action" value="transfer" />
          <div className="grid min-w-0 gap-2">
            <label htmlFor="transfer-recipient" className="font-sans text-sm font-medium text-foreground">
              New owner
            </label>
            <FormSelect id="transfer-recipient" name="user_id" required options={eligible.map((c) => ({ value: c.userId, label: memberName(c) }))} />
          </div>
          <div className="grid min-w-0 gap-2">
            <label htmlFor="transfer-former" className="font-sans text-sm font-medium text-foreground">
              Your role afterwards
            </label>
            <FormSelect
              id="transfer-former"
              name="former_role"
              defaultValue={ROLES[kind].admin}
              required
              options={[...roleOptions(kind), { value: "none", label: "Leave after the transfer" }]}
            />
          </div>
          <div>
            <PendingSubmit label="Transfer ownership" pendingLabel="Transferring…" variant="outline" />
          </div>
        </form>
      )}
      {paged ? (
        <div className="flex flex-wrap gap-x-8 gap-y-3" data-testid="transfer-pages">
          {laterPage === undefined ? null : <TextLink href={laterPage}>First eligible admins</TextLink>}
          {candidatesNext === undefined ? null : <TextLink href={candidatesNext}>More eligible admins</TextLink>}
        </div>
      ) : null}
    </div>
  );
}
