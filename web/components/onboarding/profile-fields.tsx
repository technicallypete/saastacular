import { FieldGroup } from "@/components/ui/field";
import { FIELD_LABELS, PROFILE_FIELDS, type ProfileField, type ProfileView } from "@/lib/onboarding-client";

import { FormField } from "./form-field";

const HINTS: Partial<Record<ProfileField, string>> = {
  display_name: "The name people in this workspace see for this business.",
  description: "What the business does. Tabs and line breaks are kept.",
  website: "A full address such as https://example.com.",
};

// The stored value of each field of core's record (null when not provided).
export const VALUE_OF: Readonly<Record<ProfileField, (p: ProfileView) => string | null>> = {
  display_name: (p) => p.displayName,
  legal_name: (p) => p.legalName,
  description: (p) => p.description,
  industry: (p) => p.industry,
  website: (p) => p.website,
};

// G018 (Plan r5 §1.6): the five Business Profile fields of the shared editor and of onboarding's "Create a new Business
// Profile" form. values are core's stored record (the editor) or absent (a new profile); a refused submission is never
// echoed back. error is the field and fixed message of core's field code.
export function ProfileFields({ values, error }: { values?: ProfileView; error?: { field: string; message: string } }) {
  return (
    <FieldGroup>
      {PROFILE_FIELDS.map((name) => (
        <FormField
          key={name}
          name={name}
          label={FIELD_LABELS[name]}
          required={name === "display_name"}
          hint={HINTS[name]}
          error={error?.field === name ? error.message : undefined}
          multiline={name === "description"}
          type={name === "website" ? "url" : "text"}
          defaultValue={values === undefined ? undefined : (VALUE_OF[name](values) ?? "")}
        />
      ))}
    </FieldGroup>
  );
}
