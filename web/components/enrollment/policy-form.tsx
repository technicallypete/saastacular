import { PendingSubmit } from "@/components/controls/pending-submit";
import { RadioChoice } from "@/components/controls/radio-choice";
import { FieldDescription, FieldLegend, FieldSet } from "@/components/ui/field";
import { RadioGroup } from "@/components/ui/radio-group";
import { Text } from "@/components/vestive/typography";
import type { EnrollmentMode, Policy } from "@/lib/enrollment-client";

// G058 (Plan r3 §1.5–§1.6; selected-design addendum §A.5): the enrollment mode. The owner changes it (core decides:
// organization.manage_policy); everyone else sees it read-only. While the platform operator has not resolved the
// membership history, the open option is disabled with the one generic sentence (no legacy details, no user list).
// G069 S12: the options are a kit RadioGroup inside the kit FieldSet, which keeps the group's name and description.

export const MODE_LABELS: Readonly<Record<EnrollmentMode, string>> = {
  invite_only: "By invitation only",
  request_access: "People can request access; an administrator decides",
  open: "Open: anyone who signs in can join",
};

export const NOT_READY = "Open signup is unavailable until the platform operator has reviewed membership history";

export function PolicyForm({ policy, canChange, csrfToken }: { policy: Policy; canChange: boolean; csrfToken: string }) {
  if (!canChange) {
    return (
      <div className="mt-10" data-testid="enrollment-policy-readonly">
        <Text>{`Current mode: ${MODE_LABELS[policy.mode]}.`}</Text>
        <Text variant="meta" className="mt-3.5">
          Only an owner can change it.
        </Text>
      </div>
    );
  }
  return (
    <form method="post" action="/organization/enrollment/action" className="mt-10 grid max-w-xl gap-6" data-testid="enrollment-policy-form">
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <input type="hidden" name="action" value="policy" />
      <FieldSet aria-describedby={policy.openReady ? undefined : "enrollment-open-hint"}>
        <FieldLegend id="enrollment-mode-legend">Who can join</FieldLegend>
        <RadioGroup name="mode" defaultValue={policy.mode} aria-labelledby="enrollment-mode-legend">
          {(Object.keys(MODE_LABELS) as EnrollmentMode[]).map((mode) => {
            const disabled = mode === "open" && !policy.openReady;
            return (
              <RadioChoice key={mode} value={mode} disabled={disabled} data-testid={`enrollment-mode-${mode}`}>
                {MODE_LABELS[mode]}
              </RadioChoice>
            );
          })}
        </RadioGroup>
        {policy.openReady ? null : (
          <FieldDescription id="enrollment-open-hint" data-testid="enrollment-open-not-ready">
            {NOT_READY}
          </FieldDescription>
        )}
      </FieldSet>
      <div>
        <PendingSubmit label="Save" pendingLabel="Saving…" data-testid="enrollment-policy-save" />
      </div>
    </form>
  );
}
