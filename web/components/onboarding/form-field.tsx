import type { ReactNode } from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { fieldErrorId } from "@/lib/onboarding-client";

type FormFieldProps = {
  name: string;
  label: string;
  required?: boolean;
  hint?: string;
  // The fixed message of core's field code for this field (never a submitted value); the field is marked invalid,
  // described by the message and focused.
  error?: string;
  multiline?: boolean;
  defaultValue?: string;
  type?: "text" | "url";
  autoComplete?: string;
  spellCheck?: boolean;
  // G018 M1: the id of a further description of the field (the slug field's advisory hint line), listed last.
  describedBy?: string;
  // G018 M1: closed content rendered after the error message (the create form's inline conflict holder line).
  children?: ReactNode;
};

// G018 (Plan r5 §1.6): one labelled field of an onboarding form: a native label, an optional hint and the error
// message associated by aria-describedby. No maxLength: core validates code points, and a UTF-16 maxLength would
// refuse valid astral values. G069: composed from the kit Field (FieldLabel, FieldDescription, FieldError), with
// data-invalid marking an invalid field as the kit expects.
export function FormField({ name, label, required = false, hint, error, multiline = false, defaultValue, type = "text", autoComplete, spellCheck, describedBy: extra, children }: FormFieldProps) {
  const id = `field-${name}`;
  const hintId = hint === undefined ? undefined : `${name}-hint`;
  const errorId = error === undefined ? undefined : fieldErrorId(name);
  const describedBy = [errorId, hintId, extra].filter((v) => v !== undefined).join(" ") || undefined;
  const common = {
    id,
    name,
    required,
    defaultValue,
    autoComplete: autoComplete ?? "off",
    spellCheck,
    "aria-describedby": describedBy,
    "aria-invalid": error === undefined ? undefined : true,
    autoFocus: error !== undefined,
  };
  return (
    <Field data-invalid={error === undefined ? undefined : true} className="min-w-0">
      <FieldLabel htmlFor={id}>
        {label}
        {required ? null : <span className="text-muted-foreground"> (optional)</span>}
      </FieldLabel>
      {hint === undefined ? null : <FieldDescription id={hintId}>{hint}</FieldDescription>}
      {/* The textarea sizes to its content (field-sizing), which ignores rows: min-h-44 keeps the eight-line starting
          height (rows still applies where content sizing is unsupported) and longer text grows the field. */}
      {multiline ? <Textarea rows={8} className="min-h-44" {...common} /> : <Input type={type} {...common} />}
      {error === undefined ? null : (
        <FieldError id={errorId} data-testid="field-error">
          {error}
        </FieldError>
      )}
      {children}
    </Field>
  );
}
