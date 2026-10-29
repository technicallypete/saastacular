import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

/** Container — 1360px column with the fluid Vestive gutter. */
export function Container({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div data-slot="container" className={cn("mx-auto max-w-[var(--vx-container)] px-[var(--vx-gutter)]", className)} {...props} />
}

export const sectionVariants = cva("relative border-b", {
  variants: { tone: { default: "bg-background", muted: "bg-muted/40", card: "bg-card" } },
  defaultVariants: { tone: "default" },
})

export interface SectionProps extends React.HTMLAttributes<HTMLElement>, VariantProps<typeof sectionVariants> {
  contained?: boolean
  containerClassName?: string
}

/** Section — a full-width marketing band closed by a rule; children sit in a Container with 88px vertical padding. */
export function Section({ tone, contained = true, className, containerClassName, children, ...props }: SectionProps) {
  return (
    <section data-slot="section" className={cn(sectionVariants({ tone }), className)} {...props}>
      {contained ? <Container className={cn("relative py-[var(--vx-section-y)]", containerClassName)}>{children}</Container> : children}
    </section>
  )
}

export interface RuledGridProps extends React.HTMLAttributes<HTMLDivElement> {
  columns?: number
  /** Below this width the grid scrolls sideways instead of reflowing. */
  minWidth?: number
}

/** RuledGrid — the visible grid: hairline rules between cells, column count kept at every width. */
export function RuledGrid({ columns = 3, minWidth = 860, className, style, children, ...props }: RuledGridProps) {
  return (
    <div data-slot="ruled-grid" className="-mt-[14px] overflow-x-auto pt-[14px]">
      <div
        className={cn("grid gap-px overflow-visible rounded-2xl border bg-border [&>*]:bg-background [&>*]:px-8 [&>*]:pt-8 [&>*]:pb-9 [&>*:first-child]:rounded-l-2xl [&>*:last-child]:rounded-r-2xl", className)}
        style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`, minWidth, ...style }}
        {...props}
      >
        {children}
      </div>
    </div>
  )
}
