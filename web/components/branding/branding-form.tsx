import { PendingSubmit } from "@/components/controls/pending-submit";
import { Card, CardContent } from "@/components/ui/card";
import { Field, FieldDescription, FieldGroup, FieldLabel, FieldLegend, FieldSet } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Text } from "@/components/vestive/typography";
import { BRANDING_TOKENS, MAX_LABEL, type BrandingView } from "@/lib/branding";

// G017 (Plan r2 §5.4): the label and colour form and the reset form of the branding management page. Native POST forms
// carrying this session's CSRF token and the view's version (a stale edit is refused by core as version_conflict). Each
// colour is optional: blank inherits the platform default for that appearance. Values are core's validated #rrggbb.

// Plain-language names of the 15 overridable semantic tokens.
export const TOKEN_LABELS: Readonly<Record<(typeof BRANDING_TOKENS)[number], string>> = {
  background: "Page background",
  foreground: "Body text",
  primary: "Primary action",
  "primary-foreground": "Primary action text",
  accent: "Accent text and links",
  "muted-foreground": "Secondary text",
  border: "Borders and dividers",
  popover: "Dialog surface",
  "popover-foreground": "Dialog text",
  muted: "Muted surface",
  secondary: "Secondary action",
  "secondary-foreground": "Secondary action text",
  destructive: "Destructive action",
  input: "Input borders",
  ring: "Focus ring",
};

function Hidden({ csrfToken, version }: { csrfToken: string; version: number }) {
  return (
    <>
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <input type="hidden" name="expected_version" value={String(version)} />
    </>
  );
}

function Colours({ mode, values }: { mode: "light" | "dark"; values: BrandingView["light"] }) {
  const title = mode === "light" ? "Light appearance" : "Dark appearance";
  return (
    <Card className="min-w-0">
      <CardContent>
        <FieldSet className="min-w-0" data-testid={`branding-${mode}`}>
          <FieldLegend>{title}</FieldLegend>
          <FieldGroup>
            {BRANDING_TOKENS.map((name) => {
              const id = `${mode}-${name}`;
              const value = values[name];
              return (
                <Field key={name}>
                  <FieldLabel htmlFor={id}>{TOKEN_LABELS[name]}</FieldLabel>
                  <div className="flex items-center gap-3">
                    <span
                      aria-hidden="true"
                      className="size-8 shrink-0 rounded-2xl border border-input"
                      style={value === undefined ? undefined : { backgroundColor: value }}
                    />
                    <Input
                      id={id}
                      name={`${mode}.${name}`}
                      type="text"
                      defaultValue={value ?? ""}
                      pattern="#[0-9a-fA-F]{6}"
                      maxLength={7}
                      autoComplete="off"
                      spellCheck={false}
                      aria-describedby={value === undefined ? `${id}-default colour-help` : "colour-help"}
                    />
                  </div>
                  {/* G069 S22: a colour saved blank is named in visible text instead of a placeholder. */}
                  {value === undefined ? <FieldDescription id={`${id}-default`}>Platform default</FieldDescription> : null}
                </Field>
              );
            })}
          </FieldGroup>
        </FieldSet>
      </CardContent>
    </Card>
  );
}

export function BrandingForm({ view, csrfToken }: { view: BrandingView; csrfToken: string }) {
  return (
    <form method="post" action="/organization/branding/update" className="flex flex-col gap-10" data-testid="branding-update">
      <Hidden csrfToken={csrfToken} version={view.version} />
      <Field>
        <FieldLabel htmlFor="display_label">Workspace name</FieldLabel>
        <Input
          id="display_label"
          name="display_label"
          type="text"
          defaultValue={view.label ?? ""}
          maxLength={MAX_LABEL}
          autoComplete="off"
          aria-describedby="label-help"
        />
        <FieldDescription id="label-help">
          Shown in this workspace&apos;s header and sign-in page. Leave blank to use the product name.
        </FieldDescription>
      </Field>
      <Text id="colour-help">
        Colours use the #rrggbb form. Leave a colour blank to keep the platform default for that appearance. Every
        combination must keep text and controls readable; a change that does not is refused with the pair that failed.
      </Text>
      <div className="grid gap-8 lg:grid-cols-2">
        <Colours mode="light" values={view.light} />
        <Colours mode="dark" values={view.dark} />
      </div>
      <div>
        <PendingSubmit label="Save branding" pendingLabel="Saving…" />
      </div>
    </form>
  );
}

export function ResetForm({ view, csrfToken }: { view: BrandingView; csrfToken: string }) {
  return (
    <form method="post" action="/organization/branding/reset" className="flex flex-col items-start gap-6" data-testid="branding-reset">
      <Hidden csrfToken={csrfToken} version={view.version} />
      <Text>
        Return the workspace name, every colour and the logo to the platform defaults. Uploaded images are kept and stay
        listed until you delete them.
      </Text>
      {/* The nowrap label is wider than the 320px content column; it wraps inside the column instead (E-BRD-5). */}
      <PendingSubmit
        label="Reset to platform defaults"
        pendingLabel="Resetting…"
        variant="destructive"
        className="max-w-full whitespace-normal text-center leading-[1.2]"
      />
    </form>
  );
}

export { Hidden as MutationFields };
