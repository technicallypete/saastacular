import { PendingSubmit } from "@/components/controls/pending-submit";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Text } from "@/components/vestive/typography";
import { ROLE_LABELS, type Invitation } from "@/lib/membership-client";

import { Expires } from "./invitation-list";
import { TokenPrefill } from "./token-prefill";

// G019 (Plan r1 §1.6): the Group invitation acceptance form. Shows the invitation core lets this recipient read, a code
// field (prefilled from the link's fragment in the browser, otherwise pasted) and the hidden CSRF token; the POST body
// carries the code to the accept route, which forwards it in core's JSON body.

export function AcceptForm({ invitation, csrfToken }: { invitation: Invitation; csrfToken: string }) {
  return (
    <div className="mt-10 grid gap-6" data-testid="accept-form">
      <Text>{`You are invited to join a Group in this workspace as ${ROLE_LABELS[invitation.role]}.`}</Text>
      <Text variant="meta">
        {invitation.issuerDisplayName === null ? "Sender not shown" : `From ${invitation.issuerDisplayName}`} · Expires <Expires iso={invitation.expiresAt} />
      </Text>
      <form method="post" action={`/invitations/${invitation.id}/accept`} className="grid max-w-xl gap-6">
        <input type="hidden" name="csrf_token" value={csrfToken} />
        <Field className="min-w-0">
          <FieldLabel htmlFor="invitation-token">Invitation code</FieldLabel>
          <FieldDescription id="invitation-token-hint">
            Opening the invitation link fills this in. Otherwise paste the code from the link (the part after #).
          </FieldDescription>
          <Input id="invitation-token" name="token" required autoComplete="off" spellCheck={false} aria-describedby="invitation-token-hint" />
        </Field>
        <TokenPrefill fieldId="invitation-token" />
        <div>
          <PendingSubmit label="Accept invitation" pendingLabel="Accepting…" />
        </div>
      </form>
    </div>
  );
}
