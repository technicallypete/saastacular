import * as React from "react"
import { cn } from "@/lib/utils"
import { Ribbon } from "./ribbon"

/** BulletList — items marked by a 14×2px gold rule, never a glyph. */
export function BulletList({ className, ...props }: React.HTMLAttributes<HTMLUListElement>) {
  return <ul data-slot="bullet-list" className={cn("m-0 grid list-none gap-3 p-0", className)} {...props} />
}
export function BulletItem({ className, children, ...props }: React.LiHTMLAttributes<HTMLLIElement>) {
  return (
    <li data-slot="bullet-item" className={cn("flex gap-3.5 text-sm leading-snug text-foreground/85", className)} {...props}>
      <span aria-hidden="true" className="mt-[9px] h-[2px] w-[14px] flex-[0_0_14px] rounded-full bg-gold" />
      <span>{children}</span>
    </li>
  )
}

export interface PlanCardProps extends Omit<React.HTMLAttributes<HTMLDivElement>, "title"> {
  verb: string
  name: string
  price: string
  period?: string
  terms?: string
  badge?: string
  features?: React.ReactNode[]
  action?: React.ReactNode
  featured?: boolean
}

/** PlanCard — one Interlock™ plan: gold verb, Oswald name and price, gold-ruled features, CTA. Featured squares its top corners so the ribbon sits flush across the full width. */
export function PlanCard({ verb, name, price, period = "/ month", terms, badge, features = [], action, featured, className, ...props }: PlanCardProps) {
  return (
    <div
      data-slot="plan-card"
      data-featured={featured ? "" : undefined}
      className={cn("relative flex flex-col rounded-2xl bg-card p-7 text-card-foreground ring-1 ring-foreground/10", featured && "rounded-t-none ring-2 ring-blue", className)}
      {...props}
    >
      {featured && <Ribbon placement="edge" className="-left-[2px] w-[calc(100%+4px)]" />}
      <div className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
        <div className="font-heading text-[1.15rem] font-bold uppercase tracking-[1.5px] text-gold">{verb}</div>
        {badge && <div className="font-label text-base uppercase tracking-[1.5px] text-orange">{badge}</div>}
      </div>
      <h3 className="m-0 mt-4 font-heading text-[1.9rem] font-semibold uppercase leading-[1.15]">{name}</h3>
      <div className="mt-5">
        <span className="font-heading text-[3rem] font-bold leading-none">{price}</span>{" "}
        <span className="font-label text-[1.08rem] text-muted-foreground">{period}</span>
      </div>
      {terms && <div className="mt-2 font-label text-[1.08rem] text-muted-foreground">{terms}</div>}
      {features.length > 0 && (
        <BulletList className="mt-6 mb-8">
          {features.map((f, i) => <BulletItem key={i}>{f}</BulletItem>)}
        </BulletList>
      )}
      {action && <div className="mt-auto">{action}</div>}
    </div>
  )
}

/** Stat — a gold Oswald figure over a Barlow label. Real numbers only, or placeholders flagged to the client. */
export function Stat({ value, label, className, ...props }: React.HTMLAttributes<HTMLDivElement> & { value: string; label: string }) {
  return (
    <div data-slot="stat" className={cn("flex flex-col gap-2", className)} {...props}>
      <div className="font-heading text-[3rem] font-bold leading-none text-gold">{value}</div>
      <div className="max-w-[28ch] font-label text-[1.08rem] leading-snug text-muted-foreground">{label}</div>
    </div>
  )
}

/** Quote — a testimonial in large Barlow, attributed by role, industry and headcount (no names). */
export function Quote({ role, context, className, children, ...props }: React.HTMLAttributes<HTMLElement> & { role: string; context?: string }) {
  return (
    <figure data-slot="quote" className={cn("m-0", className)} {...props}>
      <blockquote className="m-0 max-w-[34ch] font-label text-[2rem] leading-[1.2] text-foreground">{children}</blockquote>
      <figcaption className="mt-5 flex flex-wrap gap-x-3 font-label text-[1.08rem] text-muted-foreground">
        <span className="uppercase tracking-[1.5px] text-gold">{role}</span>
        {context && <span>{context}</span>}
      </figcaption>
    </figure>
  )
}

const STAR = "M12,17.27L18.18,21L16.54,15.36L21,10.27L14.81,9.62L12,2L9.19,9.62L3,10.27L7.45,15.36L5.82,21L12,17.27Z"

/** Rating — five stars in star gold with a source label ("5.0 on Clutch"). */
export function Rating({ value = 5, label, className, ...props }: React.HTMLAttributes<HTMLDivElement> & { value?: number; label?: string }) {
  return (
    <div data-slot="rating" className={cn("inline-flex items-center gap-3", className)} {...props}>
      <span className="inline-flex gap-0.5" role="img" aria-label={`${value} out of 5`}>
        {[0, 1, 2, 3, 4].map((i) => (
          <svg key={i} viewBox="0 0 24 24" width="20" height="20" aria-hidden="true" className={i < value ? "fill-star" : "fill-muted"}>
            <path d={STAR} />
          </svg>
        ))}
      </span>
      {label && <span className="font-label text-[1.08rem] text-muted-foreground">{label}</span>}
    </div>
  )
}

/** ImageFrame — black-and-white documentary photo in a rounded, ringed frame; a hatched placeholder until real photos exist. */
export function ImageFrame({ src, alt = "", caption, ratio = "4 / 3", className, ...props }: React.HTMLAttributes<HTMLElement> & { src?: string; alt?: string; caption?: string; ratio?: string }) {
  return (
    <figure data-slot="image-frame" className={cn("m-0 overflow-hidden rounded-2xl bg-card ring-1 ring-foreground/10", className)} {...props}>
      <div className={cn("relative w-full", !src && "vx-hatch")} style={{ aspectRatio: ratio }}>
        {src && <img src={src} alt={alt} className="absolute inset-0 size-full object-cover grayscale" />}
      </div>
      {caption && <figcaption className="border-t px-4 py-3 font-mono text-[.72rem] tracking-[1px] text-muted-foreground">{caption}</figcaption>}
    </figure>
  )
}
