import * as React from "react";
import type { VariantProps } from "class-variance-authority";

import { buttonVariants } from "@/components/ui/button";
import { ctaButtonVariants } from "@/components/vestive/cta-button";
import { cn } from "@/lib/utils";

type ButtonLinkVariants =
  | {
      variant?: VariantProps<typeof buttonVariants>["variant"];
      size?: VariantProps<typeof buttonVariants>["size"];
    }
  | {
      /** The angled call to action (G069 S24): only the root-host landing's Sign Up link uses it. */
      variant: "cta";
      size?: VariantProps<typeof ctaButtonVariants>["size"];
    };

export type ButtonLinkProps = React.ComponentProps<"a"> & ButtonLinkVariants;

// G069 S12/S24: a link styled as a button, a native <a> carrying the kit's button classes and data-slot="button".
// Links keep their own semantics and are never rendered as buttons (Base UI's documented pattern for links). The
// kit's angled call to action renders only where the platform palette is guaranteed: its fixed white label cannot
// follow a tenant's primary-foreground. That site passes its own focus classes (S15) as className. This is the only
// app module that imports the call-to-action classes.
export function ButtonLink({ className, variant, size, ...props }: ButtonLinkProps) {
  const classes =
    variant === "cta" ? cn(ctaButtonVariants({ size }), className) : cn(buttonVariants({ variant, size, className }));
  return <a data-slot="button" className={classes} {...props} />;
}
