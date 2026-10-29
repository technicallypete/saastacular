"use client";

import * as React from "react";

import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

// A full-width action whose label may wrap (the height follows the label), replacing the kit's fixed height and
// single line.
const BLOCK_CLASSES = "h-auto min-h-8 w-full whitespace-normal py-1.5 text-center leading-[1.2]";

type PendingSubmitProps = Omit<React.ComponentProps<typeof Button>, "type" | "children" | "className"> & {
  label: string;
  pendingLabel: string;
  block?: boolean;
  className?: string;
};

// The submit button of a native POST form with an accessible busy state, on the kit Button (G069 S12; the kit's
// angled call to action is never a submit, S24). Once its form submits, the button stays focusable but reports
// aria-disabled and its pending label, a polite status region announces the label, and a second submission is
// blocked until the page is left. A page restored from the back-forward cache starts idle again. Without script the
// form submits as plain HTML.
export function PendingSubmit({ label, pendingLabel, block = false, className, ...props }: PendingSubmitProps) {
  const ref = React.useRef<HTMLButtonElement>(null);
  const submitted = React.useRef(false);
  const [pending, setPending] = React.useState(false);

  React.useEffect(() => {
    const form = ref.current?.form;
    if (!form) return;
    const onSubmit = (event: SubmitEvent) => {
      if (event.submitter !== ref.current) return;
      if (submitted.current) {
        event.preventDefault();
        return;
      }
      submitted.current = true;
      setPending(true);
    };
    const onPageShow = (event: PageTransitionEvent) => {
      if (!event.persisted) return;
      submitted.current = false;
      setPending(false);
    };
    form.addEventListener("submit", onSubmit);
    window.addEventListener("pageshow", onPageShow);
    return () => {
      form.removeEventListener("submit", onSubmit);
      window.removeEventListener("pageshow", onPageShow);
    };
  }, []);

  return (
    <>
      <Button
        ref={ref}
        type="submit"
        aria-disabled={pending ? true : undefined}
        className={cn(block && BLOCK_CLASSES, className)}
        {...props}
      >
        {pending ? pendingLabel : label}
      </Button>
      <span role="status" className="sr-only">
        {pending ? pendingLabel : ""}
      </span>
    </>
  );
}
