import { StatusMessage } from "@/components/account/status-message";
import { FormSelect } from "@/components/controls/form-select";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { Heading, Text, TextLink } from "@/components/vestive/typography";
import { ORGANIZATION_LABELS_PATH, type GroupLabel, type Label } from "@/lib/labels-client";

import { LabelField } from "./label-form";

// G020 (Plan r2 §1.5, D-7): the Labels section of a Group page, for Organization owners and admins only. The page
// renders it only when core answers the Group's assignments (listGroupLabels ok); for anyone else there is no section,
// no heading and no hidden markup. Chips are StateBadge-style spans (text only, no colour); the assign form offers the
// Organization's definitions not yet assigned (the kit Select through FormSelect, G069 r3, the first offered by
// default); each chip has its own remove form. Core decides every read and write. Never imports server configuration
// or the request context.

type Props = {
  groupId: string;
  labels: GroupLabel[];
  // The Organization's definitions not yet assigned to this Group; undefined when they cannot be listed right now.
  unassigned: Label[] | undefined;
  csrfToken: string;
  // A fixed message chosen by the page from GROUP_LABEL_STATUS, or undefined.
  notice: string | undefined;
};

function Hidden({ csrfToken, action, labelId }: { csrfToken: string; action: string; labelId?: string }) {
  return (
    <>
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <input type="hidden" name="action" value={action} />
      {labelId === undefined ? null : <input type="hidden" name="label_id" value={labelId} />}
    </>
  );
}

export function GroupLabelsSection({ groupId, labels, unassigned, csrfToken, notice }: Props) {
  const action: `/${string}` = `/groups/${groupId}/labels/action`;
  return (
    <section className="mt-12 grid gap-6 rounded-2xl border border-border p-6" aria-labelledby="group-labels-title" data-testid="group-labels">
      <Heading level={2} id="group-labels-title">
        Labels
      </Heading>
      <Text variant="meta">Organization labels are visible to Organization owners and admins only, never to this Group&apos;s members.</Text>
      {notice === undefined ? null : (
        <div data-testid="labels-status">
          <StatusMessage message={notice} />
        </div>
      )}
      {labels.length === 0 ? (
        <Text data-testid="group-labels-empty">No labels are assigned to this Group.</Text>
      ) : (
        <ul className="flex flex-wrap gap-4">
          {labels.map((l) => (
            <li key={l.labelId} className="flex flex-wrap items-center gap-2">
              <span
                data-testid="group-label"
                className="inline-flex items-center rounded-2xl border border-border px-2 py-0.5 font-sans text-xs font-medium text-foreground [overflow-wrap:anywhere]"
              >
                {l.name}
              </span>
              <form method="post" action={action}>
                <Hidden csrfToken={csrfToken} action="unassign" labelId={l.labelId} />
                <PendingSubmit label="Remove" pendingLabel="Removing…" size="sm" variant="outline" aria-label={`Remove the label ${l.name}`} />
              </form>
            </li>
          ))}
        </ul>
      )}
      {unassigned === undefined ? (
        <StatusMessage message="The labels of this workspace cannot be listed right now." />
      ) : unassigned.length === 0 ? (
        <Text>
          Every label of this workspace is already assigned to this Group, or there are none yet.{" "}
          <TextLink href={ORGANIZATION_LABELS_PATH}>Workspace labels</TextLink>
        </Text>
      ) : (
        <form method="post" action={action} className="grid max-w-xl gap-6" data-testid="group-label-assign-form">
          <Hidden csrfToken={csrfToken} action="assign" />
          <LabelField id="group-label-assign" label="Label to assign">
            <FormSelect id="group-label-assign" name="label_id" required options={unassigned.map((l) => ({ value: l.id, label: l.name }))} />
          </LabelField>
          <div>
            <PendingSubmit label="Assign label" pendingLabel="Assigning…" />
          </div>
        </form>
      )}
    </section>
  );
}
