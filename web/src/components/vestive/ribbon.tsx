import * as React from "react"
import { cn } from "@/lib/utils"

export interface RibbonProps extends React.HTMLAttributes<HTMLSpanElement> {
  /** hero: 75% under the H1 · cta: full width under a closing heading · edge: rides the top of a featured card. */
  placement?: "hero" | "cta" | "edge"
}

/** Ribbon — the animated swash from the logo's top line. A rare accent: hero, closing CTA, one featured card. */
export function Ribbon({ placement = "hero", className, ...props }: RibbonProps) {
  if (placement === "edge") return <span aria-hidden="true" data-slot="ribbon" className={cn("vx-ribbon-edge", className)} {...props} />
  return (
    <span
      aria-hidden="true"
      data-slot="ribbon"
      className={cn("vx-ribbon", placement === "hero" ? "mt-[26px] h-[30px] w-3/4 max-w-[640px]" : "mt-[18px] h-[26px] w-full", className)}
      {...props}
    />
  )
}
