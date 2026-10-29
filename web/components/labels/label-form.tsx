import type { ReactNode } from "react";

import { PendingSubmit } from "@/components/controls/pending-submit";
import { RadioChoice } from "@/components/controls/radio-choice";
import { Field, FieldLabel, FieldLegend, FieldSet } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { RadioGroup } from "@/components/ui/radio-group";
import { Heading, Text } from "@/components/vestive/typography";
import { MAX_LABEL_NAME, VISIBILITIES, VISIBILITY_LABELS, type LabelScope, type LabelVisibility } from "@/lib/labels-client";

// G020 (Plan r2 §1.5, D-7): the label definition forms, patterned on the Group-administration form (a native radio
// fieldset, PendingSubmit, the hidden CSRF token). An Organization label's visibility is fixed (Organization
// administrators) and stated, never submitted; a Group label's visibility is a radio choice between Group administrators
// (owner_scope) and Group members (target_members). Core validates the name and decides every write. Never imports server
// configuration or the request context.

// G069: one labelled control, the kit Field with its FieldLabel.
export function LabelField({ id, label, children }: { id: string; label: string; children: ReactNode }) {
  return (
    <Field className="min-w-0">
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
      {children}
    </Field>
  );
}

export function NameInput({ id, defaultValue }: { id: string; defaultValue?: string }) {
  // No maxLength: it counts UTF-16 units, not code points; core enforces the 1–80 code point rule and the refusal says so.
  return <Input id={id} name="name" required autoComplete="off" spellCheck={false} defaultValue={defaultValue} />;
}

// The Group visibility choice; the legend names the decision and each option says who reads the label. G069 S12: the
// options are a kit RadioGroup inside the kit FieldSet, which keeps the group's name.
export function VisibilityChoice({ id, legend, current }: { id: string; legend: string; current: LabelVisibility }) {
  return (
    <FieldSet>
      <FieldLegend id={`${id}-legend`}>{legend}</FieldLegend>
      <RadioGroup name="visibility" defaultValue={current} aria-labelledby={`${id}-legend`}>
        {VISIBILITIES.map((v) => (
          <RadioChoice key={v} value={v} data-testid={`${id}-${v}`}>
            {VISIBILITY_LABELS.group[v]}
          </RadioChoice>
        ))}
      </RadioGroup>
    </FieldSet>
  );
}

export function LabelCreateForm({ scope, action, csrfToken }: { scope: LabelScope; action: `/${string}`; csrfToken: string }) {
  return (
    <div className="mt-12 grid gap-6">
      <Heading level={2} id="label-create-title">
        New label
      </Heading>
      <form method="post" action={action} aria-labelledby="label-create-title" className="grid max-w-xl gap-6" data-testid="label-create-form">
        <input type="hidden" name="csrf_token" value={csrfToken} />
        <input type="hidden" name="action" value="create" />
        <LabelField id="label-create-name" label={`Label name (1 to ${MAX_LABEL_NAME} characters)`}>
          <NameInput id="label-create-name" />
        </LabelField>
        {scope === "organization" ? (
          <Text variant="meta" data-testid="label-visibility-fixed">
            {`Visible to: ${VISIBILITY_LABELS.organization.owner_scope}. Organization labels are never shown to Group members.`}
          </Text>
        ) : (
          <VisibilityChoice id="label-create-visibility" legend="Who can see this label" current="owner_scope" />
        )}
        <div>
          <PendingSubmit label="Create label" pendingLabel="Creating…" />
        </div>
      </form>
    </div>
  );
}
