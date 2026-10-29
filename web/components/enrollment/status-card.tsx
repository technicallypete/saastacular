import { StatusMessage } from "@/components/account/status-message";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { Text } from "@/components/vestive/typography";
import type { EnrollmentMode, StatusView } from "@/lib/enrollment-client";

// G058 (Plan r3 §1.3 steps 3–4, §1.6): the applicant's own state from core's reconciliation read, and the one next step
// it allows. Nothing here shows tenant data, and nothing claims membership: "admitted" is the historical admission
// proof only, so the next step after it is always the ordinary sign-in, whose guards decide current membership; a
// current-membership hint (member) changes the wording, never the authority. Expired requests and approvals are stated
// as expired, never as pending or approved.

function Csrf({ token }: { token: string }) {
  return <input type="hidden" name="csrf_token" value={token} />;
}

export function SignInAgain({ label = "Sign in" }: { label?: string }) {
  return (
    <form method="post" action="/auth/login" className="mt-10">
      <PendingSubmit label={label} pendingLabel="Signing in…" data-testid="join-sign-in" />
    </form>
  );
}

function Abandon({ token }: { token: string }) {
  return (
    <form method="post" action="/join/abandon" className="mt-6">
      <Csrf token={token} />
      <PendingSubmit label="Stop joining" pendingLabel="Stopping…" variant="outline" data-testid="join-abandon" />
    </form>
  );
}

export function StatusCard({ view, mode }: { view: StatusView; mode: EnrollmentMode | undefined }) {
  const { status, request, member, csrfToken } = view;
  const say = (text: string) => <StatusMessage message={text} className="mt-10" />;
  if (status === "admitted") {
    return member ? (
      <div data-testid="join-state-admitted">
        {say("Your membership is confirmed. Sign in to open the workspace.")}
        <SignInAgain />
      </div>
    ) : (
      <div data-testid="join-state-admitted-not-member">
        {say("You joined earlier, but your account is not a current member of this workspace. Ask an administrator.")}
      </div>
    );
  }
  if (member) {
    return (
      <div data-testid="join-state-member">
        {say("Your account is already a member of this workspace. Sign in to open it.")}
        <SignInAgain />
      </div>
    );
  }
  if (status === "none") {
    if (mode === "request_access") {
      return (
        <div data-testid="join-state-none">
          <Text className="mt-10">You have not asked to join this workspace yet. An administrator decides each request.</Text>
          <form method="post" action="/join/request" className="mt-10">
            <Csrf token={csrfToken} />
            <PendingSubmit label="Request access" pendingLabel="Sending…" data-testid="join-request" />
          </form>
          <Abandon token={csrfToken} />
        </div>
      );
    }
    if (mode === "open") {
      return (
        <div data-testid="join-state-none">
          <Text className="mt-10">This workspace accepts new members. Confirm to join it.</Text>
          <form method="post" action="/join/admit" className="mt-10">
            <Csrf token={csrfToken} />
            <input type="hidden" name="kind" value="open" />
            <PendingSubmit label="Join this workspace" pendingLabel="Joining…" data-testid="join-admit-open" />
          </form>
          <Abandon token={csrfToken} />
        </div>
      );
    }
    return (
      <div data-testid="join-state-closed">
        {say("Membership of this workspace is by invitation. Ask an administrator for an invitation link.")}
        <Abandon token={csrfToken} />
      </div>
    );
  }
  if (request === null) return say("Your joining state could not be shown. Please try again later.");
  if (status === "pending") {
    return request.expired ? (
      <div data-testid="join-state-expired">
        {say("Your request expired before an administrator decided it.")}
        <Abandon token={csrfToken} />
      </div>
    ) : (
      <div data-testid="join-state-pending">
        {say("Your request is waiting for an administrator's decision. Nothing has been granted yet.")}
        <Abandon token={csrfToken} />
      </div>
    );
  }
  if (status === "approved") {
    if (request.expired || request.consumed) {
      return (
        <div data-testid="join-state-expired">
          {say(request.consumed ? "This approval was already used." : "This approval expired before you confirmed it. Ask an administrator again.")}
          <Abandon token={csrfToken} />
        </div>
      );
    }
    return (
      <div data-testid="join-state-approved">
        <Text className="mt-10">Your request was approved. Confirm to join; the approval is checked again when you do.</Text>
        <form method="post" action="/join/admit" className="mt-10">
          <Csrf token={csrfToken} />
          <input type="hidden" name="kind" value="request" />
          <input type="hidden" name="request_id" value={request.id} />
          <PendingSubmit label="Join this workspace" pendingLabel="Joining…" data-testid="join-admit-request" />
        </form>
        <Abandon token={csrfToken} />
      </div>
    );
  }
  if (status === "revoked") {
    return (
      <div data-testid="join-state-revoked">
        {say("The approval of your request was withdrawn. Nothing was granted.")}
        <Abandon token={csrfToken} />
      </div>
    );
  }
  return (
    <div data-testid="join-state-denied">
      {say("Your request was not approved.")}
      <Abandon token={csrfToken} />
    </div>
  );
}
