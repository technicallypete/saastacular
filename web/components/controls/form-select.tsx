import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { cn } from "@/lib/utils";

// G069 (owner finding, r3): every form dropdown is the kit's compound Select, never the browser's option list. One
// closed choice of a native POST form: the trigger (the kit's combobox, carrying the site's `id` so the site's
// `<label htmlFor>` names it) shows the chosen option's label in the kit's own styling at
// full field width; the popup is the kit's select-content with the options inside the kit's padded group; the kit's
// hidden input posts `name=value` with the chosen value string. Like the native select it replaces, the first option
// is chosen unless the site passes `defaultValue`, so a required field is never submitted empty; the server action
// re-validates every value. No hooks: the site stays a server component and every prop serialises.

export type FormSelectOption = { value: string; label: string };

type Props = {
  id: string;
  name: string;
  options: FormSelectOption[];
  defaultValue?: string;
  required?: boolean;
  // Layout classes for the trigger (full width by default).
  className?: string;
};

export function FormSelect({ id, name, options, defaultValue = options[0]?.value, required = false, className }: Props) {
  return (
    <Select name={name} items={options} defaultValue={defaultValue} required={required}>
      <SelectTrigger id={id} className={cn("w-full", className)}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent alignItemWithTrigger={false}>
        <SelectGroup>
          {options.map((o) => (
            <SelectItem key={o.value} value={o.value}>
              {o.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}
