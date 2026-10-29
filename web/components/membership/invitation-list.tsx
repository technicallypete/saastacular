import { PendingSubmit } from "@/components/controls/pending-submit";
import { ScrollRegion } from "@/components/controls/scroll-region";
import { Item, ItemActions, ItemContent, ItemDescription, ItemGroup, ItemTitle } from "@/components/ui/item";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Text, TextLink } from "@/components/vestive/typography";
import { ROLE_LABELS, type Invitation, type InvitationStatus } from "@/lib/membership-client";

// G019 (Plan r1 §1.6): invitation lists. The administrator view of one scope (the address only where core supplies it to
// the issuing scope's administrators; Revoke for a pending invitation) and the signed-in member's own pending Group
// invitations (no address; each links to its acceptance page). Never imports server configuration or request context.
// G069 S12: a kit Table in a labelled scroll region; its body cells wrap (the kit's cells keep one line).

const STATUS_LABELS: Readonly<Record<InvitationStatus, string>> = { pending: "Pending", expired: "Expired", revoked: "Revoked", consumed: "Accepted" };

const WHEN = new Intl.DateTimeFormat("en-GB", { dateStyle: "medium", timeStyle: "short", timeZone: "UTC" });

export function Expires({ iso }: { iso: string }) {
  return <time dateTime={iso}>{`${WHEN.format(new Date(iso))} UTC`}</time>;
}

function recipient(i: Invitation): string {
  if (i.intendedAddress !== null) return i.intendedAddress;
  return i.intendedUserId !== null ? "A current workspace member" : "Not shown";
}

export function InvitationList({ invitations, action, csrfToken }: { invitations: Invitation[]; action: `/${string}`; csrfToken: string }) {
  if (invitations.length === 0) return <Text className="mt-8">No invitations.</Text>;
  return (
    <ScrollRegion label="Invitations" className="mt-8">
      <Table data-testid="invitation-table">
        <TableHeader>
          <TableRow>
            <TableHead scope="col">Recipient</TableHead>
            <TableHead scope="col">Role</TableHead>
            <TableHead scope="col">Expires</TableHead>
            <TableHead scope="col">Status</TableHead>
            <TableHead scope="col">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {invitations.map((i) => (
            <TableRow key={i.id} data-testid="invitation-row">
              <TableCell className="whitespace-normal [overflow-wrap:anywhere]">{recipient(i)}</TableCell>
              <TableCell className="whitespace-normal">{ROLE_LABELS[i.role]}</TableCell>
              <TableCell className="whitespace-normal">
                <Expires iso={i.expiresAt} />
              </TableCell>
              <TableCell className="whitespace-normal">{STATUS_LABELS[i.status]}</TableCell>
              <TableCell className="whitespace-normal">
                {i.status === "pending" ? (
                  <form method="post" action={action}>
                    <input type="hidden" name="csrf_token" value={csrfToken} />
                    <input type="hidden" name="action" value="revoke" />
                    <input type="hidden" name="invitation_id" value={i.id} />
                    <PendingSubmit label="Revoke" pendingLabel="Revoking…" size="sm" variant="outline" aria-label={`Revoke the invitation for ${recipient(i)}`} />
                  </form>
                ) : (
                  <span className="text-muted-foreground">None</span>
                )}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </ScrollRegion>
  );
}

export function MyInvitations({ invitations }: { invitations: Invitation[] }) {
  if (invitations.length === 0) return <Text className="mt-10">You have no pending Group invitations in this workspace.</Text>;
  return (
    <ItemGroup className="mt-10" data-testid="my-invitations">
      {invitations.map((i) => (
        <Item key={i.id} variant="outline" role="listitem">
          {/* Layout only: the text wraps instead of the kit's line clamps, so a long sender name stays readable. */}
          <ItemContent className="min-w-0">
            <ItemTitle className="line-clamp-none">{`Group invitation · ${ROLE_LABELS[i.role]}`}</ItemTitle>
            <ItemDescription className="line-clamp-none [overflow-wrap:anywhere]">
              {i.issuerDisplayName === null ? "Sender not shown" : `From ${i.issuerDisplayName}`} · Expires <Expires iso={i.expiresAt} />
            </ItemDescription>
          </ItemContent>
          <ItemActions>
            <TextLink href={`/invitations/${i.id}`}>Open invitation</TextLink>
          </ItemActions>
        </Item>
      ))}
    </ItemGroup>
  );
}
