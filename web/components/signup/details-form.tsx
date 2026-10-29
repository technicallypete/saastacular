import { PendingSubmit } from "@/components/controls/pending-submit";
import { FormField } from "@/components/onboarding/form-field";
import { Button } from "@/components/ui/button";

// G062 (Plan r5 §4.4; Codex Round 14 option A): the root /signup/details forms. Every form posts its signup credential's
// CSRF token (from core's status read) with the route's own members; the routes read the signup credential from its
// host-only cookie and core decides. Values shown come only from the browser's own form restoration cookie (the exact
// name and slug it submitted, or core's suggested slug for that name), never from a URL. G057 (Plan r5 §3.3): the slug is
// the organization's short name, metadata only; its web address is a label core generates at creation, so no address is
// previewed or built here.

const NAME_LABEL = "Organization name";
const SLUG_LABEL = "Short name";
const SLUG_HINT =
  "Your organization's web address is generated when it is created; this short name is not its web address. Lowercase letters, digits and hyphens, up to 63 characters.";

// Step 1: the Organization name. Posting it without a slug asks core for a suggested, editable short name.
export function NameForm({ csrfToken, error, defaultName }: { csrfToken: string; error?: string; defaultName?: string }) {
  return (
    <form method="post" action="/signup/create" className="mt-12 grid max-w-[40rem] gap-8" data-testid="signup-name-form">
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <FormField name="name" label={NAME_LABEL} required error={error} defaultValue={defaultName} autoComplete="organization" />
      <div>
        <PendingSubmit label="Continue" pendingLabel="Continuing…" data-testid="signup-name-submit" />
      </div>
    </form>
  );
}

// Step 2: the name and the editable short name (suggested, or exactly as submitted last). Creating posts both.
export function DetailsForm({
  csrfToken,
  name,
  slug,
  nameError,
  slugError,
}: {
  csrfToken: string;
  name: string;
  slug: string;
  nameError?: string;
  slugError?: string;
}) {
  return (
    <form method="post" action="/signup/create" className="mt-12 grid max-w-[40rem] gap-8" data-testid="signup-details-form">
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <FormField name="name" label={NAME_LABEL} required error={nameError} defaultValue={name} autoComplete="organization" />
      <FormField
        name="slug"
        label={SLUG_LABEL}
        required
        hint={SLUG_HINT}
        error={slugError}
        defaultValue={slug}
        spellCheck={false}
      />
      <div>
        <PendingSubmit label="Create organization" pendingLabel="Creating your organization…" data-testid="signup-create" />
      </div>
    </form>
  );
}

// The failure-path "Try again" (Plan r5 §1.4 recovery (b)): the route repeats the identical create payload from the form
// restoration cookie, so core answers already_created and mints a fresh continuation.
export function RetryForm({ csrfToken }: { csrfToken: string }) {
  return (
    <form method="post" action="/signup/create" className="mt-10" data-testid="signup-retry-form">
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <input type="hidden" name="retry" value="1" />
      <PendingSubmit label="Try again" pendingLabel="Trying again…" data-testid="signup-retry" />
    </form>
  );
}

// Stop signing up: core revokes this browser's signup credential (an Organization already created stays).
export function AbandonForm({ csrfToken }: { csrfToken: string }) {
  return (
    <form method="post" action="/signup/abandon" className="mt-10" data-testid="signup-abandon-form">
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <Button type="submit" variant="outline" data-testid="signup-abandon">
        Stop signing up
      </Button>
    </form>
  );
}
