import { ConnectionManager } from "@/components/account/connection-manager";
import { RULED_GRID_REFLOW } from "@/components/account/ruled-grid-reflow";
import { StatusMessage } from "@/components/account/status-message";
import { ButtonLink } from "@/components/controls/button-link";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { RuledGrid, Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text } from "@/components/vestive/typography";
import type { SessionView } from "@/lib/identity-client";
import type { RequestContext } from "@/lib/request-context";

// G063 (Plan r1 §5.2): the signed-in body of My Account, shared by /account and the test-only rendered fixture
// (/account-fixtures, D-5): the Session section (name, multi-factor status, session times, the retirement notice and the
// management links), the shared MCP connection manager and the Sign-out section with its three forms. It reads no
// cookies or headers: /account resolves the session, its notices and every flag, and passes them in. The connection
// manager makes the one extra core call (the list) and reduces its failures inside its own section, so Session and
// Sign-out always render.

const WHEN = new Intl.DateTimeFormat("en-GB", { dateStyle: "medium", timeStyle: "short", timeZone: "UTC" });

function when(iso: string) {
  return <time dateTime={iso}>{`${WHEN.format(new Date(iso))} UTC`}</time>;
}

type Param = string | string[] | undefined;

export type AccountSectionsProps = {
  value: SessionView;
  // A fixed account notice chosen by the page from its closed table, or undefined.
  notice: string | undefined;
  soleOwned: { count: number; truncated: boolean } | undefined;
  manages: boolean;
  // The connection manager's inputs: the request context for its list call, this request's session credential and
  // origin, and the raw mcp_cursor and mcp_status parameters (validated by the manager, never reflected).
  mcp: { ctx: RequestContext; credential: string; origin: string; cursor: Param; status: Param };
};

export function AccountSections({ value, notice, soleOwned, manages, mcp }: AccountSectionsProps) {
  return (
    <>
      <Section id="session" aria-labelledby="account-title">
        <Kicker>Account</Kicker>
        <Heading level={1} id="account-title" className="mt-6">
          Your account
        </Heading>
        <StatusMessage message={notice} className="mt-10" />
        <dl className="mt-10 grid gap-6 font-sans text-sm text-foreground/80">
          <div>
            <dt className="text-muted-foreground">Name</dt>
            <dd data-testid="display-name" className="m-0 [overflow-wrap:anywhere]">
              {value.displayName ?? "Not provided"}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Multi-factor status</dt>
            <dd data-testid="mfa-status" className="m-0">
              Not verified by this application
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Signed in</dt>
            <dd className="m-0">{when(value.authTime)}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Session expires</dt>
            <dd data-testid="session-expires" className="m-0">
              {when(value.expiresAt)}
            </dd>
          </div>
        </dl>
        {soleOwned === undefined ? null : (
          <Alert className="mt-10" data-testid="retirement-notice">
            <AlertDescription>
              {`You are the only owner of ${soleOwned.truncated ? "at least " : ""}${soleOwned.count} ${soleOwned.count === 1 ? "workspace or Group" : "workspaces or Groups"}. Before this account can be retired, transfer ownership in each workspace first.`}
            </AlertDescription>
          </Alert>
        )}
        {manages ? (
          <div className="mt-10 flex flex-wrap gap-6">
            <ButtonLink href="/organization/branding" data-testid="account-branding-link" variant="outline">
              Manage workspace branding
            </ButtonLink>
            {/* G060 (Plan r0 §1.6): the Group-administration setting has the same audience as the branding view
                (an Organization owner or admin); its own page decides what each sees. */}
            <ButtonLink href="/organization/group-authority" data-testid="account-group-authority-link" variant="outline">
              Group administration
            </ButtonLink>
            {/* G020 (Plan r2 §1.5; Q-S0-1 option (a)): the Organization's labels, for the same audience; its own page
                decides what each sees. */}
            <ButtonLink href="/organization/labels" data-testid="account-labels-link" variant="outline">
              Workspace labels
            </ButtonLink>
          </div>
        ) : null}
      </Section>
      <ConnectionManager page="account" ctx={mcp.ctx} credential={mcp.credential} origin={mcp.origin} cursor={mcp.cursor} status={mcp.status} />
      <Section id="sign-out" aria-labelledby="sign-out-title" tone="muted">
        <Heading level={2} id="sign-out-title">
          Sign out
        </Heading>
        <RuledGrid className="mt-12" style={RULED_GRID_REFLOW}>
          <form method="post" action="/auth/logout" className="flex flex-col items-start gap-6">
            <input type="hidden" name="csrf_token" value={value.csrfToken} />
            <Text>End this session in this browser.</Text>
            <PendingSubmit label="Sign out" pendingLabel="Signing out…" />
          </form>
          <form method="post" action="/auth/logout-all" className="flex flex-col items-start gap-6">
            <input type="hidden" name="csrf_token" value={value.csrfToken} />
            <Text>End every session of this account in this application, on every device.</Text>
            <PendingSubmit label="Sign out everywhere" pendingLabel="Signing out everywhere…" variant="secondary" />
          </form>
          <form method="post" action="/auth/provider-logout" className="flex flex-col items-start gap-6">
            <input type="hidden" name="csrf_token" value={value.csrfToken} />
            <Text>End this session and revoke the WorkOS session for this sign-in.</Text>
            <PendingSubmit label="Sign out and revoke WorkOS session" pendingLabel="Signing out…" variant="outline" block />
          </form>
        </RuledGrid>
      </Section>
    </>
  );
}
