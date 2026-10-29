import { PendingSubmit } from "@/components/controls/pending-submit";

// G062 (Plan r5 §4.4): the root signup start. The plain form posts to /signup/start, whose route starts the platform
// signup login (intent "signup") and sends the browser to the identity provider's hosted page; nothing is entered here.
export function SignupForm() {
  return (
    <form method="post" action="/signup/start" className="mt-10" data-testid="signup-start-form">
      <PendingSubmit label="Continue to sign up" pendingLabel="Continuing…" data-testid="signup-start" />
    </form>
  );
}
