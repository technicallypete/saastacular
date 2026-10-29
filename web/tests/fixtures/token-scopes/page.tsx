// Test-only route. The browser-test image copies this file to app/token-fixtures/page.tsx in its
// fixture build stage; it is never part of the production web image (asserted by S-NOFIXTURE).
import { Plus } from "lucide-react";

import { BrandPreview } from "@/components/branding/brand-preview";
import { TenantStyle } from "@/components/branding/tenant-style";
import { ButtonLink } from "@/components/controls/button-link";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { RadioChoice } from "@/components/controls/radio-choice";
import { ScrollRegion } from "@/components/controls/scroll-region";
import { BlueprintBackdrop } from "@/components/layout/blueprint-backdrop";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { NativeSelect, NativeSelectOption } from "@/components/ui/native-select";
import { RadioGroup } from "@/components/ui/radio-group";
import { Toaster } from "@/components/ui/sonner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { BulletItem, BulletList } from "@/components/vestive/content";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";

import { FIXTURE_PREVIEWS, FIXTURE_TENANTS, fixtureBranded } from "@/tests/fixtures/token-scopes/fixtures";
import { FixtureOverlays } from "@/tests/fixtures/token-scopes/overlays";

export const dynamic = "force-dynamic";

// Every kit Button variant the application uses, bare: the kit's own classes, focus state included (G069 owner kit-fidelity
// direction, 2026-10-26: the app's destructive focus border is withdrawn).
const VARIANTS = ["default", "outline", "secondary", "ghost", "destructive"] as const;
const SIZES = ["sm", "default", "lg"] as const;
// Font probes: every bundled face and weight the shell uses, plus Geist 500 and 700.
const FONT_PROBES = [
  { id: "oswald-400", className: "font-heading font-normal" },
  { id: "oswald-500", className: "font-heading font-medium" },
  { id: "oswald-600", className: "font-heading font-semibold" },
  { id: "oswald-700", className: "font-heading font-bold" },
  { id: "barlow-400", className: "font-label font-normal" },
  { id: "barlow-600", className: "font-label font-semibold" },
  { id: "geist-400", className: "font-sans font-normal" },
  { id: "geist-500", className: "font-sans font-medium" },
  { id: "geist-700", className: "font-sans font-bold" },
  { id: "geist-mono-400", className: "font-mono font-normal" },
] as const;

// The same content on every surface (the root and each preview): text, a link, each kit Button variant enabled and
// disabled, the sizes, a link styled as a button, a typography sampler, font probes, an invalid field, a card panel and
// a blueprint panel. Nothing here portals, so a preview may render it.
function SurfaceContent({ label }: { label: string }) {
  return (
    <>
      <p data-testid="body-text">Body text in this scope.</p>
      <p data-testid="muted-text" className="text-muted-foreground">
        Muted text in this scope.
      </p>
      <p>
        <a data-testid="link" href="#top" className="text-gold underline underline-offset-4">
          A link in this scope
        </a>
      </p>
      <div className="flex flex-wrap gap-4">
        {VARIANTS.map((variant) => (
          <Button key={variant} variant={variant} data-testid={`button-${variant}`}>
            {variant} action
          </Button>
        ))}
      </div>
      <div className="flex flex-wrap gap-4">
        {VARIANTS.map((variant) => (
          <Button key={variant} variant={variant} disabled data-testid={`button-${variant}-disabled`}>
            {variant} action
          </Button>
        ))}
      </div>
      <div className="flex flex-wrap items-center gap-4">
        {SIZES.map((size) => (
          <Button key={size} size={size} data-testid={`button-size-${size}`}>
            {size} size
          </Button>
        ))}
        <Button size="icon" aria-label="Add" data-testid="button-size-icon">
          <Plus />
        </Button>
        <ButtonLink href="#top" variant="secondary" data-testid="button-as-link">
          Link as button
        </ButtonLink>
      </div>
      <div data-testid="typography-sampler" className="grid gap-3">
        <Kicker data-testid="sample-kicker">Kicker label</Kicker>
        <Heading level={1} as="div" data-testid="sample-h1">
          Heading one
        </Heading>
        <Heading level={2} as="div" data-testid="sample-h2">
          Heading two
        </Heading>
        <Heading level={3} as="div" data-testid="sample-h3">
          Heading three
        </Heading>
        <Heading level={2} cta as="div" data-testid="sample-cta">
          Call to action heading
        </Heading>
        <Text variant="lead" data-testid="sample-lead">
          Lead text in this scope.
        </Text>
        <Text variant="body" data-testid="sample-body">
          Body text in this scope.
        </Text>
        <Text variant="meta" data-testid="sample-meta">
          Meta text in this scope.
        </Text>
        <p>
          <TextLink href="#top" data-testid="sample-text-link">
            A text link
          </TextLink>
        </p>
        <BulletList data-testid="sample-bullets">
          <BulletItem>First bullet item</BulletItem>
          <BulletItem>Second bullet item</BulletItem>
        </BulletList>
      </div>
      <div data-testid="font-probes" className="grid gap-1">
        {FONT_PROBES.map((probe) => (
          <p key={probe.id} data-font-probe={probe.id} className={probe.className}>
            Font probe {probe.id}
          </p>
        ))}
      </div>
      <div className="max-w-[360px]">
        <Input aria-invalid="true" aria-label={`Invalid field in surface ${label}`} data-testid="input-invalid" />
      </div>
      <Section tone="card" contained={false} className="p-6" data-testid="raised-panel">
        <Text>Raised panel in this scope.</Text>
      </Section>
      {/* 320px wide where there is room, narrower in a narrow surface: a fixed width with a percentage maximum would
          contribute the full 320px as this grid item's minimum size, and the surfaces are grid items themselves, so
          the page would scroll sideways at 320px. */}
      <div data-testid="blueprint-panel" className="relative isolate h-[160px] w-full max-w-[320px] overflow-hidden">
        <BlueprintBackdrop />
      </div>
    </>
  );
}

// The root surface only: the portaled kit components, a table, a radio on the muted band and the native form (the kit
// fields bare, as the kit ships them, G069 owner kit-fidelity direction 2026-10-26: no app boundary or colour classes;
// the bare kit NativeSelect as the kit field proof; a kit RadioGroup with one disabled item; the pending submit in the
// sign-in shape). Under the platform palette only, the one call to action with its S15 focus classes.
function RootContent({ open, platform }: { open: boolean; platform: boolean }) {
  return (
    <>
      <FixtureOverlays open={open} />
      <ScrollRegion label="Fixture table">
        <Table data-testid="fixture-table">
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Role</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>Fixture member</TableCell>
              <TableCell>Member</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </ScrollRegion>
      <div data-testid="muted-panel" className="bg-muted p-4">
        <RadioGroup name="muted-choice" aria-label="Muted choice">
          <RadioChoice value="muted" data-testid="radio-muted">
            Muted option
          </RadioChoice>
        </RadioGroup>
      </div>
      <form method="post" action="/token-fixtures/submit" aria-label="Fixture form" data-testid="fixture-form" className="grid max-w-[360px] gap-4">
        <div className="grid gap-2">
          <Label htmlFor="fixture-title">Title</Label>
          <Input id="fixture-title" name="title" data-testid="input-rest" />
        </div>
        <Input name="disabled-field" defaultValue="locked" disabled aria-label="Disabled field" data-testid="input-disabled" />
        <Textarea name="notes" aria-label="Notes" data-testid="field-textarea" />
        <NativeSelect name="choice-select" aria-label="Native choice" defaultValue="one" data-testid="field-select" className="w-full">
          <NativeSelectOption value="one">One</NativeSelectOption>
          <NativeSelectOption value="two">Two</NativeSelectOption>
        </NativeSelect>
        <RadioGroup name="choice" defaultValue="one" required aria-label="Fixture choice" data-testid="radio-group">
          <RadioChoice value="one" data-testid="radio-one">
            One
          </RadioChoice>
          <RadioChoice value="two" data-testid="radio-two">
            Two
          </RadioChoice>
          <RadioChoice value="three" disabled data-testid="radio-three">
            Three
          </RadioChoice>
        </RadioGroup>
        <PendingSubmit label="Submit fixture" pendingLabel="Submitting fixture" size="lg" data-testid="form-submit" />
      </form>
      {platform ? (
        <p>
          <ButtonLink
            href="#top"
            variant="cta"
            className="focus-visible:outline-offset-[-3px] focus-visible:outline-white"
            data-testid="cta-link"
          >
            Start the fixture
          </ButtonLink>
        </p>
      ) : null}
    </>
  );
}

// ?tenant=<label> selects the document root's palette (default branded; platform = no tenant values), rendered by the
// app's own TenantStyle before any surface; ?open=root opens the root dialog at mount. The root surface has no ground
// of its own (the body's shows); the two previews each carry one appearance and their own values.
export default async function TokenFixtures({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const params = await searchParams;
  const label = typeof params.tenant === "string" ? params.tenant : "branded";
  const tenant = FIXTURE_TENANTS.find((t) => t.label === label) ?? { label: "branded", tokens: fixtureBranded };
  const open = params.open === "root";
  return (
    <>
      <TenantStyle tokens={tenant.tokens} />
      <main data-fixture-tenant={tenant.label} className="grid gap-6 p-6 md:grid-cols-2">
        <h1 className="sr-only">Token fixtures</h1>
        <div data-fixture-surface="root" className="border-2 border-border p-6 md:col-span-2">
          <section aria-label="Root surface" className="grid gap-4">
            <SurfaceContent label="root" />
            <RootContent open={open} platform={tenant.label === "platform"} />
          </section>
        </div>
        {FIXTURE_PREVIEWS.map((preview) => (
          <BrandPreview
            key={preview.label}
            label={preview.label}
            appearance={preview.appearance}
            tokens={preview.tokens}
            data-fixture-surface={preview.label}
            className="border-2 border-border p-6"
          >
            <section aria-label={`Preview ${preview.label}`} className="grid gap-4">
              <SurfaceContent label={preview.label} />
            </section>
          </BrandPreview>
        ))}
      </main>
      <Toaster />
    </>
  );
}
