import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

// The angled CTA: the same silhouette as the container behind the V in the Vestive icon.
// Reserve it for the one call to action a view is for ("Book a free assessment"); everyday actions use <Button>.
export const ctaButtonVariants = cva(
  "vx-angle relative inline-flex shrink-0 cursor-pointer items-center justify-center gap-2 whitespace-nowrap border-0 font-heading font-medium uppercase leading-none tracking-[1.5px] no-underline transition-[filter] duration-150 hover:brightness-110 active:brightness-90 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:size-4 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "bg-blue text-white",
        green: "bg-green text-white",
        lime: "bg-lime text-ink",
      },
      size: {
        sm: "px-7 py-3.5 text-[.85rem]",
        default: "px-10 py-[18px] text-[.95rem]",
        lg: "px-12 py-[22px] text-base",
      },
      block: { true: "w-full whitespace-normal text-center leading-[1.2]", false: "" },
    },
    defaultVariants: { variant: "default", size: "default", block: false },
  }
)

export interface CtaButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof ctaButtonVariants> {}

/** CtaButton — the angled call-to-action. For links, put ctaButtonVariants() on an <a>. */
export function CtaButton({ className, variant, size, block, ...props }: CtaButtonProps) {
  return <button data-slot="cta-button" className={cn(ctaButtonVariants({ variant, size, block }), className)} {...props} />
}
