import { RULED_GRID_REFLOW } from "@/components/account/ruled-grid-reflow";
import { StatusMessage } from "@/components/account/status-message";
import { ButtonLink } from "@/components/controls/button-link";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { RuledGrid, Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text } from "@/components/vestive/typography";
import { isListCursor, listConnections, type ConnectionList, type ConnectionView, type McpResult, type ReturnTarget } from "@/lib/mcp-client";
import type { RequestContext } from "@/lib/request-context";

// G063 (Plan r1 §5.1): the one shared, server-rendered MCP connection manager, embedded in My Account (/account, through
// AccountSections) and in the compatibility page (/connections/mcp). It lists the signed-in member's own live
// connections in this host's Organization (newest first, 100 per page, opaque cursor), with each row's friendly name
// (escaped React text; "Unnamed connection" when none) kept separate from the actually bound client, a rename form and
// a revoke form per row, and a create form with an optional name. Every form carries this session's CSRF token and the
// closed return_to of its page. It makes at most one core call (the list) and reduces every failure to a closed state
// rendered inside its own section, so a failed list never breaks the rest of the embedding page. No cookie access, no client
// component of its own and nothing read from the request is reflected: status notices come from a closed table.

export type ManagerPage = ReturnTarget;

// Page-specific anchors, headings and links: the compatibility page keeps its existing ids and query parameters
// (cursor, status); My Account uses its own (mcp_cursor, mcp_status), disjoint from its account notices.
const PAGES: Readonly<Record<ManagerPage, { list: string; listTitle: string; create: string; createTitle: string; newest: string; older: (cursor: string) => string }>> = {
  connections: {
    list: "connections",
    listTitle: "mcp-title",
    create: "create",
    createTitle: "create-title",
    newest: "/connections/mcp",
    older: (cursor) => `/connections/mcp?cursor=${encodeURIComponent(cursor)}`,
  },
  account: {
    list: "mcp-connections",
    listTitle: "mcp-connections-title",
    create: "mcp-create",
    createTitle: "mcp-create-title",
    newest: "/account#mcp-connections",
    older: (cursor) => `/account?mcp_cursor=${encodeURIComponent(cursor)}#mcp-connections`,
  },
};

// The closed setup outcomes a form route may redirect with (its page's status parameter); an unknown or repeated value
// shows nothing.
const STATUS: Readonly<Record<string, string>> = {
  created: "A new connection was created. Paste its URL into your AI client soon to finish setup; an unused connection expires.",
  revoked: "The connection was revoked. It can never be used again.",
  renamed: "The connection's name was saved.",
  invalid_name: "That name was not saved. Use up to 80 characters without line breaks or control characters. Nothing was changed.",
};

const WHEN = new Intl.DateTimeFormat("en-GB", { dateStyle: "medium", timeStyle: "short", timeZone: "UTC" });

function when(iso: string) {
  return <time dateTime={iso}>{`${WHEN.format(new Date(iso))} UTC`}</time>;
}

type Param = string | string[] | undefined;

export type ConnectionManagerProps = {
  page: ManagerPage;
  ctx: RequestContext;
  credential: string;
  origin: string;
  // The page's raw cursor and status parameters (cursor/status on the compatibility page, mcp_cursor/mcp_status on My
  // Account); validated here, never reflected.
  cursor: Param;
  status: Param;
  // The compatibility page's own list answer (it decides its whole-page branches from it), so the page still makes one
  // core call; My Account omits it and the component makes the call.
  preloaded?: McpResult<ConnectionList>;
};

// A closed message in a polite status region inside the section (distinct test ids, so the page's own status message
// stays unique): the kit Alert with the status role, as StatusMessage.
function SectionNotice({ testId, message }: { testId: string; message: string }) {
  return (
    <Alert role="status" data-testid={testId} className="mt-10">
      <AlertDescription>{message}</AlertDescription>
    </Alert>
  );
}

function Pages({ page, paged, nextCursor }: { page: ManagerPage; paged: boolean; nextCursor: string | null }) {
  if (!paged && nextCursor === null) return null;
  const p = PAGES[page];
  return (
    <nav aria-label="Connection pages" className="mt-12 flex flex-wrap gap-6">
      {paged ? (
        <ButtonLink href={p.newest} data-testid="mcp-page-newest" variant="outline">
          Newest connections
        </ButtonLink>
      ) : null}
      {nextCursor !== null ? (
        <ButtonLink href={p.older(nextCursor)} data-testid="mcp-page-older" variant="outline">
          Older connections
        </ButtonLink>
      ) : null}
    </nav>
  );
}

function Connection({ c, csrfToken, page }: { c: ConnectionView; csrfToken: string; page: ManagerPage }) {
  const nameId = `mcp-name-input-${c.connectionId}`;
  const hintId = `mcp-name-hint-${c.connectionId}`;
  return (
    <li data-testid="mcp-connection" data-state={c.state} className="flex flex-col gap-6">
      {/* The friendly name is presentation metadata only, shown as text exactly as stored (casing kept). */}
      <Heading level={3} as={page === "connections" ? "h2" : "h3"} data-testid="mcp-name" className="normal-case [overflow-wrap:anywhere]">
        {c.displayName ?? "Unnamed connection"}
      </Heading>
      <dl className="grid gap-4 font-sans text-sm text-foreground/80">
        <div>
          <dt className="text-muted-foreground">Connection URL</dt>
          <dd data-testid="mcp-url" className="m-0 font-mono text-[0.8125rem] [overflow-wrap:anywhere]">
            {c.url}
          </dd>
        </div>
        <div>
          {/* G016: a workspace's connections may live on several of its addresses; each row names its own. */}
          <dt className="text-muted-foreground">Address</dt>
          <dd data-testid="mcp-host" className="m-0 font-mono text-[0.8125rem] [overflow-wrap:anywhere]">
            {c.host}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Status</dt>
          <dd data-testid="mcp-state" className="m-0">
            {c.state === "active" ? "Active" : "Waiting for your AI client"}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Bound client</dt>
          <dd data-testid="mcp-client" className="m-0 [overflow-wrap:anywhere]">
            {c.clientId ?? "Not yet bound"}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Created</dt>
          <dd className="m-0">{when(c.createdAt)}</dd>
        </div>
        <div>
          <dt className="text-muted-foreground">Activated</dt>
          <dd className="m-0">{c.activatedAt === null ? "Not yet" : when(c.activatedAt)}</dd>
        </div>
      </dl>
      <form method="post" action="/connections/mcp/rename" className="grid min-w-0 max-w-xl gap-2" data-testid="mcp-rename-form">
        <input type="hidden" name="csrf_token" value={csrfToken} />
        <input type="hidden" name="connection_id" value={c.connectionId} />
        <input type="hidden" name="return_to" value={page} />
        <Field>
          <FieldLabel htmlFor={nameId}>Name</FieldLabel>
          <FieldDescription id={hintId}>Up to 80 characters; leave empty to remove the name.</FieldDescription>
          {/* No maxLength: core validates code points, and a UTF-16 maxLength would refuse valid names. */}
          <Input
            id={nameId}
            name="display_name"
            type="text"
            defaultValue={c.displayName ?? ""}
            autoComplete="off"
            spellCheck={false}
            aria-describedby={hintId}
          />
        </Field>
        <div className="mt-4">
          <PendingSubmit label="Save name" pendingLabel="Saving…" variant="secondary" />
        </div>
      </form>
      <form method="post" action="/connections/mcp/revoke">
        <input type="hidden" name="csrf_token" value={csrfToken} />
        <input type="hidden" name="connection_id" value={c.connectionId} />
        <input type="hidden" name="return_to" value={page} />
        <PendingSubmit label="Revoke this connection" pendingLabel="Revoking…" variant="outline" />
      </form>
    </li>
  );
}

type View =
  | { kind: "ok"; list: ConnectionList }
  | { kind: "invalid_page" }
  | { kind: "not_available" }
  | { kind: "unavailable" };

// Reduces the list answer to the component's closed states: a cursor of the wrong shape (or one core refuses) is an
// invalid page; a non-member or a refused origin is not available; every other failure (mcp_unavailable, timeout,
// unreachable, malformed — an unconfigured MCP reduces to malformed — or an ended session) is unavailable.
function reduce(result: McpResult<ConnectionList> | undefined, cursorGiven: boolean): View {
  if (result === undefined) return { kind: "invalid_page" };
  if (result.ok) return { kind: "ok", list: result.value };
  if (result.error === "invalid_request" && cursorGiven) return { kind: "invalid_page" };
  if (result.error === "not_available" || result.error === "origin_rejected") return { kind: "not_available" };
  return { kind: "unavailable" };
}

export async function ConnectionManager({ page, ctx, credential, origin, cursor: cursorParam, status, preloaded }: ConnectionManagerProps) {
  const p = PAGES[page];
  const cursorValid = cursorParam === undefined || (typeof cursorParam === "string" && isListCursor(cursorParam));
  const cursor = typeof cursorParam === "string" && cursorValid ? cursorParam : undefined;
  const result = !cursorValid ? undefined : (preloaded ?? (await listConnections(ctx, credential, origin, cursor)));
  const view = reduce(result, cursorParam !== undefined);
  const notice = view.kind === "ok" && typeof status === "string" && Object.hasOwn(STATUS, status) ? STATUS[status] : undefined;
  const top = page === "connections";

  return (
    <>
      <Section id={p.list} aria-labelledby={p.listTitle}>
        {top ? <Kicker>AI connections</Kicker> : null}
        <Heading level={top ? 1 : 2} id={p.listTitle} className={top ? "mt-6" : undefined}>
          Your AI client connections
        </Heading>
        <Text className="mt-10">
          Each connection URL links one AI client to this workspace with your own current permissions. Group selection
          happens inside a connection. A connection&apos;s name is only a label for you; it never changes what the
          connection can do. Revoking a connection ends it on this side; a new connection gets a new URL.
        </Text>
        {view.kind === "unavailable" ? (
          <SectionNotice testId="mcp-unavailable" message="Connections are unavailable right now. Your account and sign-out still work. Please try again later." />
        ) : view.kind === "not_available" ? (
          <SectionNotice testId="mcp-not-available" message="Connections are not available for your account here." />
        ) : view.kind === "invalid_page" ? (
          <>
            <SectionNotice testId="mcp-invalid-page" message="This page of connections is not valid for your account here." />
            <ButtonLink href={p.newest} data-testid="mcp-page-newest" className="mt-10">
              Show your newest connections
            </ButtonLink>
          </>
        ) : (
          <>
            <StatusMessage message={notice} className="mt-10" />
            {view.list.connections.length === 0 ? (
              <Text className="mt-10" data-testid="mcp-empty">
                {cursor !== undefined ? "No older connections." : "You have no connections here yet."}
              </Text>
            ) : (
              <ul className="mt-10 flex list-none flex-col gap-12 p-0" aria-label="Your connections">
                {view.list.connections.map((c) => (
                  <Connection key={c.connectionId} c={c} csrfToken={view.list.csrfToken} page={page} />
                ))}
              </ul>
            )}
            <Pages page={page} paged={cursor !== undefined} nextCursor={view.list.nextCursor} />
          </>
        )}
      </Section>
      {view.kind === "ok" ? (
        <Section id={p.create} aria-labelledby={p.createTitle} tone="muted">
          <Heading level={top ? 2 : 3} as={top ? "h2" : "h3"} id={p.createTitle}>
            New connection
          </Heading>
          <RuledGrid className="mt-12" style={RULED_GRID_REFLOW}>
            <form method="post" action="/connections/mcp/create" className="flex flex-col items-start gap-6">
              <input type="hidden" name="csrf_token" value={view.list.csrfToken} />
              <input type="hidden" name="return_to" value={page} />
              <Text>Create a connection URL for an AI client, then finish setup in the client soon: an unused connection expires.</Text>
              <Field className="min-w-0">
                <FieldLabel htmlFor={`${p.create}-name`}>
                  Name<span className="text-muted-foreground"> (optional)</span>
                </FieldLabel>
                <FieldDescription id={`${p.create}-name-hint`}>
                  A label for you, such as the AI client or device. Up to 80 characters.
                </FieldDescription>
                <Input
                  id={`${p.create}-name`}
                  name="display_name"
                  type="text"
                  autoComplete="off"
                  spellCheck={false}
                  aria-describedby={`${p.create}-name-hint`}
                />
              </Field>
              <PendingSubmit label="Create connection" pendingLabel="Creating…" />
            </form>
          </RuledGrid>
        </Section>
      ) : null}
    </>
  );
}
