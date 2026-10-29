import { PendingSubmit } from "@/components/controls/pending-submit";
import { RadioChoice } from "@/components/controls/radio-choice";
import { FieldLegend, FieldSet } from "@/components/ui/field";
import { RadioGroup } from "@/components/ui/radio-group";
import { Text } from "@/components/vestive/typography";
import { GROUP_AUTHORITY_PATH, type GroupAuthority } from "@/lib/membership-client";

// G060 (Plan r0 §1.6, D-11; Spec R2, R4): the Organization's Group-administration setting, patterned on the enrollment
// policy form (a native radio fieldset, the scope text referenced by aria-describedby, PendingSubmit). The owner changes
// it (core decides: organization.manage_policy, owner only); an admin sees the effective state read-only. The scope text
// states exactly what the setting allows and what it never allows.
// G069 S12: the options are a kit RadioGroup inside the kit FieldSet, which keeps the group's name and description.

export const GROUP_AUTHORITY_LABELS: Readonly<Record<"off" | "on", string>> = {
  off: "Off: each Group is administered only by its own owners and admins",
  on: "On: Organization owners and admins may also administer every Group",
};

export const GROUP_AUTHORITY_SCOPE: readonly string[] = [
  "This setting applies to every owner and admin of this workspace alike.",
  "When it is on, they may manage a Group's members, invitations, name, address, onboarding and Business Profile without joining the Group.",
  "It never gives access to a Group's content or files, never makes anyone a Group owner and never lets them add themselves to a Group or raise their own role.",
  "People who administer a Group this way can grant access to other people. Turn it on only if you trust every owner and admin with that.",
  "Turning it off takes effect immediately. Group roles that people actually hold are kept.",
];

function Scope() {
  return (
    <div id="group-authority-scope" className="grid gap-3.5" data-testid="group-authority-scope">
      {GROUP_AUTHORITY_SCOPE.map((s) => (
        <Text key={s} variant="meta">
          {s}
        </Text>
      ))}
    </div>
  );
}

export function GroupAuthorityForm({ authority, canChange, csrfToken }: { authority: GroupAuthority; canChange: boolean; csrfToken: string }) {
  const current = authority.overrideEnabled ? "on" : "off";
  if (!canChange) {
    return (
      <div className="mt-10 grid max-w-xl gap-6" data-testid="group-authority-readonly">
        <Text data-testid="group-authority-state">{`Current setting: ${GROUP_AUTHORITY_LABELS[current]}.`}</Text>
        <Text variant="meta">Only an owner of this workspace can change it.</Text>
        <Scope />
      </div>
    );
  }
  return (
    <form method="post" action={`${GROUP_AUTHORITY_PATH}/action`} className="mt-10 grid max-w-xl gap-6" data-testid="group-authority-form">
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <FieldSet aria-describedby="group-authority-scope">
        <FieldLegend id="group-authority-legend">Group administration by Organization staff</FieldLegend>
        <RadioGroup name="enabled" defaultValue={current} aria-labelledby="group-authority-legend">
          {(["off", "on"] as const).map((v) => (
            <RadioChoice key={v} value={v} data-testid={`group-authority-${v}`}>
              {GROUP_AUTHORITY_LABELS[v]}
            </RadioChoice>
          ))}
        </RadioGroup>
      </FieldSet>
      <Scope />
      <div>
        <PendingSubmit label="Save" pendingLabel="Saving…" data-testid="group-authority-save" />
      </div>
    </form>
  );
}
