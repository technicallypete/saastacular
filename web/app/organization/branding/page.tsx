import { notFound } from "next/navigation";

import { RULED_GRID_REFLOW } from "@/components/account/ruled-grid-reflow";
import { StatusMessage } from "@/components/account/status-message";
import { BrandingForm, ResetForm } from "@/components/branding/branding-form";
import { LogoForms } from "@/components/branding/logo-forms";
import { shellBrand, tenantRequest } from "@/components/branding/tenant-frame";
import { ButtonLink } from "@/components/controls/button-link";
import { AppShell, type ShellNavItem } from "@/components/layout/app-shell";
import { RuledGrid, Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text } from "@/components/vestive/typography";
import { getBranding, PAIR_FIELD } from "@/lib/branding";
import { runtimeConfig } from "@/lib/config";
import { lookupContext } from "@/lib/host";
import { session } from "@/lib/identity-client";

// Organization branding management on this tenant host (G017 Plan r2 §5.4): the workspace name, the per-appearance colour
// overrides, the logo and its uploads, and reset — native POST forms carrying this session's CSRF token and the view's
// version, following connections/mcp. Server-rendered and never cached. Core alone authorizes (manage_branding: owner or
// admin of this host's Organization); anyone else, and any host core does not route as a tenant, gets the fixed 404.
// Anonymous visitors get a sign-in-first page. The page's own core call is the session read; the public branding read
// (tenant frame) and the management read carry lookup request IDs.
export const dynamic = "force-dynamic";

const NAV: ShellNavItem[] = [
  { href: "#branding", label: "Name and colours" },
  { href: "#logo", label: "Logo" },
  { href: "#reset", label: "Reset" },
];

// The closed status map: each form's outcome, by core's closed code (never a value).
const STATUS: Readonly<Record<string, string>> = {
  updated: "The branding was saved. New pages of this workspace use it now.",
  reset: "The branding was reset to the platform defaults. Uploaded images are kept below.",
  logo_set: "The new logo was checked, stored and is now shown.",
  logo_removed: "The logo is no longer shown. The image is listed under other uploads.",
  upload_deleted: "The upload was deleted.",
  deletion_pending: "The upload is being deleted; it disappears from the list once its storage is removed.",
  upload_not_found: "That upload is not one of this workspace's other uploads (it may already be deleted). Nothing was changed.",
  version_conflict: "The branding changed since this page was loaded. Review the current values below and try again. Nothing was changed.",
  contrast_failed: "Those colours would make text or controls hard to read, so nothing was changed.",
  invalid_input: "A value was not in the accepted form, so nothing was changed.",
  logo_invalid: "That file is not an accepted logo: use a PNG or JPEG from 16 to 2048 pixels on each side. Nothing was changed.",
  payload_too_large: "That file is larger than 512 KiB. Nothing was changed.",
  publication_outcome_unknown:
    "The logo upload's outcome could not be confirmed. The current state is shown below: an upload that did not become the logo is listed and can be deleted.",
};

// Plain-language descriptions of core's contrast pairs (core/internal/branding/colors.go, palette delta r3), in core's
// evaluation order: T1-T10, T12, N2-N4 (there is no T11).
const PAIRS: Readonly<Record<string, string>> = {
  T1: "text on the page background",
  T2: "body text, including softened text, on the page background",
  T3: "text on muted surfaces and on hovered outline and ghost buttons",
  T4: "secondary text on the page background and muted surfaces",
  T5: "accent text on the page background, muted surfaces and dialogs",
  T6: "dialog text, including softened text, on the dialog surface",
  T7: "text on the border colour",
  T8: "primary button text on the primary colour, at rest and when hovered",
  T9: "secondary button text on the secondary colour, at rest and when hovered",
  T10: "the destructive colour as text: on its own tint in destructive buttons, at rest and when hovered, and as error text on the page background",
  T12: "text in input fields on the input fill, on the page and in dialogs",
  N2: "the focus ring against the page background, muted surfaces and dialogs",
  N3: "input borders against the page background and dialogs",
  N4: "secondary text against the dialog surface",
};

const INPUTS: Readonly<Record<string, string>> = {
  display_label: "The workspace name must be 1 to 80 characters with no control characters.",
  light: "A Light colour is not in the #rrggbb form.",
  dark: "A Dark colour is not in the #rrggbb form.",
};

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function notice(status: unknown, f: unknown): string | undefined {
  if (typeof status !== "string" || !Object.hasOwn(STATUS, status)) return undefined;
  if (status === "contrast_failed" && typeof f === "string" && PAIR_FIELD.test(f)) {
    const [pair, mode] = f.split(":");
    return `${STATUS[status]} Failed check ${pair} in ${mode === "light" ? "Light" : "Dark"}: ${PAIRS[pair]}.`;
  }
  if (status === "invalid_input" && typeof f === "string" && Object.hasOwn(INPUTS, f)) return `${STATUS[status]} ${INPUTS[f]}`;
  return STATUS[status];
}

export default async function OrganizationBranding({ searchParams }: { searchParams: SearchParams }) {
  const { productName, platformBaseDomain } = runtimeConfig();
  const req = await tenantRequest();
  if (req.tenant === undefined) notFound();
  const frame = (viewer: Parameters<typeof shellBrand>[1], nav: ShellNavItem[], children: React.ReactNode) => (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={nav} brand={shellBrand(req, viewer, productName)}>
      {children}
    </AppShell>
  );
  const signInFirst = frame(
    undefined,
    NAV.slice(0, 1),
    <Section id="branding" aria-labelledby="branding-title">
      <Kicker>Workspace branding</Kicker>
      <Heading level={1} id="branding-title" className="mt-6">
        Sign in first
      </Heading>
      <Text className="mt-10" data-testid="branding-sign-in-first">
        Sign in to this workspace to manage its branding.
      </Text>
      <ButtonLink href="/sign-in" className="mt-10">Sign in</ButtonLink>
    </Section>,
  );
  if (req.credential === undefined || req.origin === undefined) return signInFirst;
  const viewer = await session(req.ctx, req.credential, req.origin);
  if (!viewer.ok && viewer.error === "unauthenticated") return signInFirst;
  if (!viewer.ok && viewer.error === "unknown_host") notFound();
  const unavailable = (v: typeof viewer) =>
    frame(
      v,
      NAV.slice(0, 1),
      <Section id="branding" aria-labelledby="branding-title">
        <Kicker>Workspace branding</Kicker>
        <Heading level={1} id="branding-title" className="mt-6">
          Not available
        </Heading>
        <StatusMessage message="Branding is unavailable right now. Please try again later." className="mt-10" />
      </Section>,
    );
  if (!viewer.ok) return unavailable(viewer);
  const view = await getBranding(lookupContext(req.ctx), req.origin, req.credential);
  if (!view.ok && view.error === "unauthenticated") return signInFirst;
  if (!view.ok && ["forbidden", "not_available", "not_found", "unknown_host", "origin_rejected"].includes(view.error)) notFound();
  if (!view.ok) return unavailable(viewer);
  const params = await searchParams;
  const message = notice(params.status, params.field);
  const { csrfToken } = viewer.value;
  const label = view.value.label ?? productName;

  return frame(
    viewer,
    NAV,
    <>
      <Section id="branding" aria-labelledby="branding-title">
        <Kicker>Workspace branding</Kicker>
        <Heading level={1} id="branding-title" className="mt-6">
          Branding for {label}
        </Heading>
        <Text className="mt-10">
          The name, colours and logo shown on this workspace&apos;s pages, including its sign-in page, at every address of
          this workspace. Changes apply to newly loaded pages.
        </Text>
        <StatusMessage message={message} className="mt-10" />
        <div className="mt-12">
          <BrandingForm view={view.value} csrfToken={csrfToken} />
        </div>
      </Section>
      <Section id="logo" aria-labelledby="logo-title" tone="muted">
        <Heading level={2} id="logo-title">
          Logo
        </Heading>
        <RuledGrid className="mt-12" style={RULED_GRID_REFLOW}>
          <LogoForms view={view.value} csrfToken={csrfToken} label={label} />
        </RuledGrid>
      </Section>
      <Section id="reset" aria-labelledby="reset-title">
        <Heading level={2} id="reset-title">
          Reset
        </Heading>
        <div className="mt-12">
          <ResetForm view={view.value} csrfToken={csrfToken} />
        </div>
      </Section>
    </>,
  );
}
