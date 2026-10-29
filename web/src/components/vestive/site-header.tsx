import * as React from "react"
import { cn } from "@/lib/utils"
import { Container } from "./layout"

export interface SiteHeaderProps extends React.HTMLAttributes<HTMLElement> {
  /** The Vestive icon or wordmark URL. */
  logoSrc: string
  /** Optional wordmark text set beside the icon. */
  wordmark?: string
  links?: { label: string; href: string }[]
  /** Usually <CtaButton size="sm">Book a free assessment</CtaButton>. */
  action?: React.ReactNode
}

const DEFAULT_LINKS = [
  { label: "Plans & Pricing", href: "#" },
  { label: "Methodology", href: "#" },
  { label: "Insights", href: "#" },
  { label: "About", href: "#" },
]

/** SiteHeader — sticky marketing bar: icon + wordmark, Oswald nav, angled CTA. */
export function SiteHeader({ logoSrc, wordmark = "Vestive", links = DEFAULT_LINKS, action, className, ...props }: SiteHeaderProps) {
  return (
    <header data-slot="site-header" className={cn("sticky top-0 z-50 border-b bg-background/95 backdrop-blur-none", className)} {...props}>
      <Container className="flex min-h-[72px] flex-wrap items-center justify-between gap-x-10 gap-y-4 py-3">
        <a href="#" className="flex items-center gap-3 no-underline">
          <img src={logoSrc} width={36} height={36} alt="" className="size-9" />
          {wordmark && <span className="font-heading text-xl font-bold uppercase tracking-[1px] text-foreground">{wordmark}</span>}
        </a>
        <nav className="flex flex-wrap gap-8 font-heading text-[.92rem] uppercase tracking-[1.2px]">
          {links.map((l) => (
            <a key={l.label} href={l.href} className="text-foreground/85 no-underline transition-colors hover:text-gold">{l.label}</a>
          ))}
        </nav>
        {action}
      </Container>
    </header>
  )
}
