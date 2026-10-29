import { PendingSubmit } from "@/components/controls/pending-submit";
import { ScrollRegion } from "@/components/controls/scroll-region";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Text } from "@/components/vestive/typography";
import { requestRowView, type ManagedRequest } from "@/lib/enrollment-client";

// G058 (Plan r3 §1.5–§1.6): the administrators' request list (one keyset page). Valid pending requests offer Approve and
// Deny; an expired pending request keeps Deny ("Expired, deny to resolve"); an unconsumed, unexpired approval offers
// Revoke (requestRowView). The applicant is shown by display name when core supplies one, otherwise as "Applicant" (no
// address or profile read is widened). The buttons are hints: core re-decides authority and state on every action.
// G069 S12: a kit Table in a labelled scroll region; its body cells wrap (the kit's cells keep one line).

const LABEL = { approve: "Approve", deny: "Deny", revoke: "Revoke approval" } as const;

function Action({ id, action, label, csrfToken }: { id: string; action: "approve" | "deny" | "revoke"; label: string; csrfToken: string }) {
  return (
    <form method="post" action="/organization/enrollment/action">
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <input type="hidden" name="action" value={action} />
      <input type="hidden" name="request_id" value={id} />
      <PendingSubmit label={label} pendingLabel="Saving…" variant={action === "approve" ? "default" : "outline"} size="sm" data-testid={`enrollment-${action}`} />
    </form>
  );
}

export function RequestList({ requests, csrfToken }: { requests: ManagedRequest[]; csrfToken: string }) {
  if (requests.length === 0) {
    return (
      <Text className="mt-10" data-testid="enrollment-requests-empty">
        There are no access requests.
      </Text>
    );
  }
  return (
    <ScrollRegion label="Access requests" className="mt-10">
      <Table data-testid="enrollment-requests">
        <TableHeader>
          <TableRow>
            <TableHead>Applicant</TableHead>
            <TableHead>State</TableHead>
            <TableHead>Requested</TableHead>
            <TableHead>Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {requests.map((r) => {
            const view = requestRowView(r);
            return (
              <TableRow key={r.id} data-testid="enrollment-request-row">
                <TableCell className="whitespace-normal [overflow-wrap:anywhere]">{r.displayName ?? "Applicant"}</TableCell>
                <TableCell className="whitespace-normal">{view.state}</TableCell>
                <TableCell className="whitespace-normal">{r.createdAt.replace("T", " ").replace("Z", " UTC")}</TableCell>
                <TableCell className="whitespace-normal">
                  <div className="flex flex-wrap gap-3">
                    {view.actions.map((action) => (
                      <Action key={action} id={r.id} action={action} label={LABEL[action]} csrfToken={csrfToken} />
                    ))}
                  </div>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </ScrollRegion>
  );
}
