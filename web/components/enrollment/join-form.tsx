import { PendingSubmit } from "@/components/controls/pending-submit";
import { TokenPrefill } from "@/components/membership/token-prefill";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

// G058 (Plan r3 §1.6): the enrollment sign-in start. Open and request-access workspaces post the plain form (request
// access also asks core to file the request on arrival); an invitation landing adds the invitation id and the code
// field, which the link's fragment fills in the browser (a fragment never reaches a server, so the code travels only in
// this form's POST body, then the one-use join cookie and core's JSON body).

export function JoinForm({ label, request = false, testId }: { label: string; request?: boolean; testId?: string }) {
  return (
    <form method="post" action="/join/start" className="mt-10">
      {request ? <input type="hidden" name="request" value="1" /> : null}
      <PendingSubmit label={label} pendingLabel="Signing in…" data-testid={testId} />
    </form>
  );
}

export function InvitationJoinForm({ invitationId }: { invitationId: string }) {
  return (
    <form method="post" action="/join/start" className="mt-10 grid max-w-xl gap-6" data-testid="join-invitation-form">
      <input type="hidden" name="invitation_id" value={invitationId} />
      <Field className="min-w-0">
        <FieldLabel htmlFor="invitation-token">Invitation code</FieldLabel>
        <FieldDescription id="invitation-token-hint">
          Opening the invitation link fills this in. Otherwise paste the code from the link (the part after #).
        </FieldDescription>
        <Input id="invitation-token" name="invitation_token" required autoComplete="off" spellCheck={false} aria-describedby="invitation-token-hint" />
      </Field>
      <TokenPrefill fieldId="invitation-token" />
      <div>
        <PendingSubmit label="Sign in to accept" pendingLabel="Signing in…" data-testid="join-invitation-submit" />
      </div>
    </form>
  );
}
