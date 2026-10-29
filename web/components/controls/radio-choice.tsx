"use client";

import * as React from "react";

import { Label } from "@/components/ui/label";
import { RadioGroupItem } from "@/components/ui/radio-group";
import { cn } from "@/lib/utils";

type RadioChoiceProps = Omit<React.ComponentProps<typeof RadioGroupItem>, "children" | "className"> & {
  children: React.ReactNode;
  className?: string;
};

// G069 S12: one option of a kit RadioGroup, the successor of the G068 native radio option. Use it inside the kit's
// documented form API, one hidden input per group:
//   <RadioGroup name="mode" defaultValue={current} required aria-labelledby={legendId}>
//     <RadioChoice value="open" disabled={!ready} data-testid="mode-open">Open</RadioChoice>
//   </RadioGroup>
// The group owns the submitted name, the selected value and `required`; each choice takes its `value`, `disabled`
// and test hooks (they reach the radio element). The kit Label points at the item's id, so all of its text,
// including any description inside it, names and selects the option. The kit item and Label render unaltered (G069
// owner kit-fidelity delta: no colour or border classes); the app adds layout only: the circle sits centred on the
// first line and wrapped text stays beside it.
export function RadioChoice({ children, className, id, ...props }: RadioChoiceProps) {
  const generatedId = React.useId();
  const itemId = id ?? generatedId;
  return (
    <div className={cn("flex min-w-0 items-start gap-3 text-sm leading-snug", className)}>
      <RadioGroupItem id={itemId} className="mt-[calc((1lh_-_1rem)/2)]" {...props} />
      <Label
        htmlFor={itemId}
        className="block min-w-0 leading-snug [overflow-wrap:anywhere] peer-data-disabled:cursor-not-allowed"
      >
        {children}
      </Label>
    </div>
  );
}
