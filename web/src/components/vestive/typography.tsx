import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

/** Kicker — the gold, wide-tracked Barlow label above a heading. */
export function Kicker({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div data-slot="kicker" className={cn("font-label text-base font-semibold uppercase leading-tight tracking-[2.5px] text-gold", className)} {...props} />
}

export const headingVariants = cva("m-0 font-heading uppercase text-foreground", {
  variants: {
    level: {
      1: "text-[clamp(2.6rem,5.6vw,5rem)] font-bold leading-none tracking-[-1px]",
      2: "text-[clamp(1.9rem,3.4vw,2.7rem)] font-semibold leading-[1.1] tracking-[-.5px]",
      3: "text-[1.35rem] font-medium leading-[1.2]",
    },
    cta: { true: "text-[clamp(2.1rem,4.4vw,3.4rem)] font-bold tracking-[-1px]", false: "" },
  },
  defaultVariants: { level: 2, cta: false },
})

export interface HeadingProps extends React.HTMLAttributes<HTMLHeadingElement>, VariantProps<typeof headingVariants> {
  as?: "h1" | "h2" | "h3" | "h4" | "div"
}

/** Heading — Oswald display heading, uppercase via CSS. Write sentence case in source. */
export function Heading({ level = 2, cta, as, className, ...props }: HeadingProps) {
  const Tag = (as ?? `h${level}`) as React.ElementType
  return <Tag data-slot="heading" className={cn(headingVariants({ level, cta }), className)} {...props} />
}

/** Emph — the one orange word a heading may carry. */
export function Emph({ className, ...props }: React.HTMLAttributes<HTMLSpanElement>) {
  return <span data-slot="emph" className={cn("text-orange", className)} {...props} />
}

export const textVariants = cva("m-0", {
  variants: {
    variant: {
      lead: "max-w-[52ch] text-[1.19rem] leading-[1.62] text-foreground/80",
      body: "max-w-[62ch] text-[1.02rem] leading-[1.65] text-muted-foreground",
      meta: "font-label text-[1.08rem] leading-[1.35] text-muted-foreground",
      index: "font-mono text-[.72rem] tracking-[1px] text-muted-foreground/80",
    },
  },
  defaultVariants: { variant: "body" },
})

export interface TextProps extends React.HTMLAttributes<HTMLParagraphElement>, VariantProps<typeof textVariants> {
  as?: "p" | "div" | "span"
}

/** Text — marketing copy styles: lead, body, meta (Barlow) and index (mono). */
export function Text({ variant, as = "p", className, ...props }: TextProps) {
  const Tag = as as React.ElementType
  return <Tag data-slot="text" className={cn(textVariants({ variant }), className)} {...props} />
}

/** TextLink — the gold, underlined Oswald link that sits beside a CTA. */
export function TextLink({ className, ...props }: React.AnchorHTMLAttributes<HTMLAnchorElement>) {
  return (
    <a
      data-slot="text-link"
      className={cn("inline-block border-b-2 border-gold pb-1 font-heading text-base uppercase tracking-[1.5px] text-gold no-underline transition-colors hover:border-foreground hover:text-foreground", className)}
      {...props}
    />
  )
}
