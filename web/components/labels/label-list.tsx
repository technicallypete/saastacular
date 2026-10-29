import { PendingSubmit } from "@/components/controls/pending-submit";
import { ScrollRegion } from "@/components/controls/scroll-region";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Text, TextLink } from "@/components/vestive/typography";
import { VISIBILITY_LABELS, type Label, type LabelScope } from "@/lib/labels-client";

import { LabelField, NameInput, VisibilityChoice } from "./label-form";

// G020 (Plan r2 §1.5, D-7): the label definitions of one owner scope. Managers (core decides label.organization.manage or
// label.group.manage) get a table with per-row rename, visibility (Group labels only) and delete forms; a Group member
// gets the read-only list of the definitions core shares with members (target_members). Names render as escaped text
// and carry no colour. Every submission is re-decided by core. Never imports server configuration or the request context.
// G069 S12: a kit Table in a labelled scroll region; its body cells wrap (the kit's cells keep one line).

type ManageProps = { scope: LabelScope; labels: Label[]; action: `/${string}`; csrfToken: string; next?: `/${string}` };

function Hidden({ csrfToken, action, labelId }: { csrfToken: string; action: string; labelId: string }) {
  return (
    <>
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <input type="hidden" name="action" value={action} />
      <input type="hidden" name="label_id" value={labelId} />
    </>
  );
}

function More({ next }: { next?: `/${string}` }) {
  return next === undefined ? null : (
    <div className="mt-6">
      <TextLink href={next} data-testid="labels-next">
        More labels
      </TextLink>
    </div>
  );
}

export function LabelList({ scope, labels, action, csrfToken, next }: ManageProps) {
  if (labels.length === 0) {
    return (
      <Text className="mt-8" data-testid="label-list">
        No labels yet.
      </Text>
    );
  }
  return (
    <>
      <ScrollRegion label="Labels" className="mt-8">
        <Table data-testid="label-list">
          <TableHeader>
            <TableRow>
              <TableHead scope="col">Name</TableHead>
              <TableHead scope="col">Visible to</TableHead>
              <TableHead scope="col">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {labels.map((l) => (
              <TableRow key={l.id} data-testid="label-row">
                <TableCell className="whitespace-normal [overflow-wrap:anywhere]" data-testid="label-name">
                  {l.name}
                </TableCell>
                <TableCell className="whitespace-normal" data-testid="label-visibility">
                  {VISIBILITY_LABELS[scope][l.visibility]}
                </TableCell>
                <TableCell className="whitespace-normal">
                  <div className="grid gap-6">
                    <form method="post" action={action} className="grid gap-3" data-testid="label-rename-form">
                      <Hidden csrfToken={csrfToken} action="rename" labelId={l.id} />
                      {scope === "group" ? <input type="hidden" name="visibility" value={l.visibility} /> : null}
                      <LabelField id={`label-rename-${l.id}`} label={`New name for ${l.name}`}>
                        <NameInput id={`label-rename-${l.id}`} defaultValue={l.name} />
                      </LabelField>
                      <PendingSubmit label="Rename" pendingLabel="Renaming…" size="sm" variant="outline" />
                    </form>
                    {scope === "group" ? (
                      <form method="post" action={action} className="grid gap-3" data-testid="label-visibility-form">
                        <Hidden csrfToken={csrfToken} action="visibility" labelId={l.id} />
                        <input type="hidden" name="name" value={l.name} />
                        <VisibilityChoice id={`label-visibility-${l.id}`} legend={`Who can see ${l.name}`} current={l.visibility} />
                        <PendingSubmit label="Change visibility" pendingLabel="Changing…" size="sm" variant="outline" />
                      </form>
                    ) : null}
                    <form method="post" action={action} data-testid="label-delete-form">
                      <Hidden csrfToken={csrfToken} action="delete" labelId={l.id} />
                      <PendingSubmit label="Delete" pendingLabel="Deleting…" size="sm" variant="outline" aria-label={`Delete ${l.name}`} />
                    </form>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </ScrollRegion>
      <More next={next} />
    </>
  );
}

// A Group member's view: the names only, no controls.
export function ReadOnlyLabelList({ labels, next }: { labels: Label[]; next?: `/${string}` }) {
  return (
    <div className="mt-8" data-testid="labels-read-only">
      {labels.length === 0 ? (
        <Text>No labels of this Group are shared with its members.</Text>
      ) : (
        <ul className="flex flex-wrap gap-3">
          {/* The kit Badge rendered as the list item; layout only: a long name wraps instead of the badge's one line. */}
          {labels.map((l) => (
            <Badge
              key={l.id}
              variant="outline"
              render={<li />}
              data-testid="label-row"
              className="h-auto max-w-full whitespace-normal [overflow-wrap:anywhere]"
            >
              {l.name}
            </Badge>
          ))}
        </ul>
      )}
      <More next={next} />
    </div>
  );
}
