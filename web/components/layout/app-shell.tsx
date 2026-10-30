import type { ReactNode } from "react";

import { ButtonLink } from "@/components/controls/button-link";
import { PLATFORM_ENTRY } from "@/components/signup/landing";
import { Container } from "@/components/vestive/layout";
import { Text, TextLink } from "@/components/vestive/typography";
import type { Indicator } from "@/lib/branding";
import { productMark } from "@/lib/product-mark";

import { AppearanceControl } from "./appearance-control";
import { BrandMark, ProductMarkImage } from "./brand-mark";
import { ContextIndicator } from "./context-indicator";
import { ShellHelpDialog } from "./shell-help-dialog";

export type ShellNavItem = { href: `#${string}`; label: string };

// G018 (Plan r5 §1.6): the workspace areas of a tenant host, offered to a signed-in member only. The links are
// navigation, never authority: each area's page asks core, which decides every read and write. G019 adds Members (the
// Organization members page; a plain member sees only their own membership card there).
export type WorkspaceArea = "groups" | "business-profiles" | "members";
const WORKSPACE_AREAS: readonly { area: WorkspaceArea; href: `/${string}`; label: string }[] = [
  { area: "groups", href: "/groups", label: "Groups" },
  { area: "business-profiles", href: "/business-profiles", label: "Business Profiles" },
  { area: "members", href: "/organization/members", label: "Members" },
];

// G017: the tenant frame of a platform or custom host (Plan r0 §1.5). label and logo come from core's public branding
// projection (null = platform default), indicator the session's membership fact. G069 (Plan r1 S7): the Organization's
// colour overrides are not the frame's; the root layout's tenant stylesheet renders them for the whole document.
export type ShellBrand = {
  label: string | null;
  logo: { version: string } | null;
  indicator: Indicator;
  // The tenant entry page carries the one Sign in action in its body, so the header omits its own.
  entry?: boolean;
};

type AppShellProps = {
  productName: string;
  platformBaseDomain: string;
  nav: ShellNavItem[];
  brand?: ShellBrand;
  // The workspace area this page belongs to (marked as the current page in the areas navigation).
  area?: WorkspaceArea;
  children: ReactNode;
};

// The shell's solid 2px focus outline on its links (the kit's links draw no focus indicator of their own).
const FOCUS_OUTLINE = "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring focus-visible:outline-solid";

// The generic application frame: skip link, ruled header with configured identity, primary
// navigation, appearance choice, keyboard help, the Account link, main content and footer. No domain
// or tenant surface. G017: with brand (a tenant host), the header shows the brand mark and the current-Organization
// indicator, and the footer the platform attribution. Without brand the frame is unchanged. G018: for a member of the
// host's Organization, a "Workspace areas" navigation follows the header; the Primary navigation is unchanged.
// G069 (Plan r1 S13): the header is the app's own composition on the kit's Container, Button and typography (the kit's
// SiteHeader is a sticky marketing bar with a logo URL and default links). It is not sticky, so it never covers
// focused content.
// OS002 (Plan r1 S6, F4): the default product mark sits before the identity text only where that text is the configured
// product name: the generic frame and a tenant frame whose label and logo are both null. A tenant frame with a label or
// a logo renders exactly as before. The public entry frame (PLATFORM_ENTRY: the platform root and www.<domain>) keeps
// text only, because the host gate serves no /brand/ file on those hosts (lib/host.ts PUBLIC_ASSET_PREFIXES). The mark
// shows only while its file exists (lib/product-mark.ts); deleting it returns every frame to text only. The img sits
// inside the identity paragraph, so that paragraph (and its text) stays the container's direct child.
function markFor(brand: ShellBrand | undefined): string | null {
  const productIdentity = brand === undefined || (brand !== PLATFORM_ENTRY && brand.label === null && brand.logo === null);
  return productIdentity ? productMark() : null;
}

export function AppShell({ productName, platformBaseDomain, nav, brand, area, children }: AppShellProps) {
  const name = brand?.label ?? productName;
  const mark = markFor(brand);
  return (
    <div className="flex min-h-screen flex-col">
      <TextLink
        href="#main"
        className={`fixed top-0 left-0 z-50 -translate-y-full bg-background px-[var(--vx-gutter)] pt-3 focus:translate-y-0 ${FOCUS_OUTLINE}`}
      >
        Skip to main content
      </TextLink>
      <header data-slot="site-header" className="relative border-b border-border bg-background">
        <Container className="flex min-h-[72px] flex-wrap items-center justify-between gap-x-10 gap-y-4 py-3">
          {brand === undefined ? (
            <p className="max-w-full min-w-0 font-heading text-lg font-semibold tracking-[1.2px] uppercase [overflow-wrap:anywhere]">
              {mark === null ? null : <ProductMarkImage src={mark} />}
              {productName}
            </p>
          ) : (
            <BrandMark name={name} logoVersion={brand.logo?.version ?? null} mark={mark} />
          )}
          <nav aria-label="Primary" className="flex flex-wrap gap-x-8 gap-y-2 font-sans text-sm font-medium">
            {nav.map((l) => (
              <a
                key={l.href}
                href={l.href}
                className={`inline-flex min-h-8 items-center text-foreground/80 no-underline transition-colors hover:text-gold ${FOCUS_OUTLINE}`}
              >
                {l.label}
              </a>
            ))}
          </nav>
          <div className="flex flex-wrap items-center gap-4">
            <AppearanceControl />
            <ShellHelpDialog />
            {brand === undefined ? (
              <ButtonLink href="/account" size="sm">Account</ButtonLink>
            ) : (
              brand.entry === true && brand.indicator.kind === "anonymous" ? null : <ContextIndicator indicator={brand.indicator} />
            )}
          </div>
        </Container>
      </header>
      {brand?.indicator.kind === "member" ? (
        <nav aria-label="Workspace areas" data-testid="workspace-areas" className="border-b border-border">
          <Container className="flex flex-wrap gap-x-9 gap-y-3 py-3 font-sans text-sm font-medium">
            {WORKSPACE_AREAS.map((a) => (
              <a
                key={a.href}
                href={a.href}
                aria-current={a.area === area ? "page" : undefined}
                className={`inline-flex min-h-8 items-center border-b-2 border-transparent text-foreground no-underline transition-colors hover:text-gold aria-[current=page]:border-gold aria-[current=page]:text-gold ${FOCUS_OUTLINE}`}
              >
                {a.label}
              </a>
            ))}
          </Container>
        </nav>
      ) : null}
      <main id="main" tabIndex={-1} className="flex-1 outline-none">
        {children}
      </main>
      <footer className="border-t border-border">
        <Container className="py-6">
          {brand === undefined ? (
            <Text variant="meta">{platformBaseDomain}</Text>
          ) : (
            <Text variant="meta" data-testid="platform-attribution" className="[overflow-wrap:anywhere]">
              {`${productName} · ${platformBaseDomain}`}
            </Text>
          )}
        </Container>
      </footer>
    </div>
  );
}
