import { StatusMessage } from "@/components/account/status-message";
import { shellBrand, tenantRequest, viewerSession } from "@/components/branding/tenant-frame";
import { ButtonLink } from "@/components/controls/button-link";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { AppShell, type ShellNavItem } from "@/components/layout/app-shell";
import { BlueprintBackdrop } from "@/components/layout/blueprint-backdrop";
import { Landing, SignupShell } from "@/components/signup/landing";
import { BulletItem, BulletList } from "@/components/vestive/content";
import { RuledGrid, Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { publicEnrollment } from "@/lib/enrollment-client";
import { lookupContext } from "@/lib/host";
import { publicEntryOf } from "@/lib/signup-client";

export const dynamic = "force-dynamic";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

// G062 (Plan r5 §1.1, §4.4): the public entry landing's in-page navigation.
const LANDING_NAV: ShellNavItem[] = [{ href: "#landing", label: "Sign up" }];
// G062 (Plan r5 §1.4 step 4, §4.4): the closed notice of the pinned continuation's landing on the new tenant.
const TENANT_CREATED = "Your organization is ready.";

const NAV: ShellNavItem[] = [
  { href: "#overview", label: "Overview" },
  { href: "#getting-around", label: "Getting around" },
  { href: "#about", label: "About" },
];

const GETTING_AROUND = [
  { title: "Skip to content", body: "The first Tab on the page reveals a link straight to the main content." },
  { title: "Keyboard help", body: "The help dialog explains moving through the page and closing dialogs." },
  { title: "Appearance", body: "Choose Light, Dark or System. System follows your device setting." },
];

const ABOUT = ["Configured product identity", "Keyboard-operable structure", "Light, dark and system appearance"];

// G069 (Plan r1 S14): the kit's RuledGrid keeps a fixed column count and scrolls sideways below its minimum width; its
// style prop (spread last) reflows the cells instead, so the grid fits a 320px viewport.
const REFLOW = { gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 16rem), 1fr))", minWidth: 0 };

// The kit's text link draws no focus indicator of its own: the shell's solid 2px outline.
const LINK_FOCUS = "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring focus-visible:outline-solid";

// G017 (Plan r2 §1.5): on a tenant host (platform or custom) the logged-out / is the tenant entry: the brand mark, the
// workspace name (public label or the configured product name), one neutral Sign in action, the appearance control,
// keyboard help and the platform attribution; nothing else (no marketing, signup or member list; G058: one Join link
// only when the public enrollment mode is open or request access). A signed-in
// viewer gets the Account link in place of Sign in. G062: ?status=tenant_created (the pinned continuation's landing)
// shows one closed notice. On the https public entry hosts the landing replaces the marketing redirect; on every other
// host (the loopback profile) the generic page below is unchanged.
const TENANT_NAV: ShellNavItem[] = [{ href: "#entry", label: "Workspace" }];

export default async function Home({ searchParams }: { searchParams: SearchParams }) {
  const { productName, platformBaseDomain } = runtimeConfig();
  const req = await tenantRequest();
  if (req.tenant !== undefined) {
    const created = (await searchParams).status === "tenant_created";
    const viewer = await viewerSession(req);
    const brand = shellBrand(req, viewer, productName);
    const name = req.tenant.label ?? productName;
    // Only a confirmed session (member or non-member) is signed in; an unavailable session read claims nothing.
    const signedIn = brand !== undefined && (brand.indicator.kind === "member" || brand.indicator.kind === "none");
    // G058: a logged-out viewer also sees one Join link when the workspace's public mode admits joining (open or request
    // access); invite-only workspaces and any failed read show none.
    let joinable = false;
    if (!signedIn && req.origin !== undefined) {
      const pub = await publicEnrollment(lookupContext(req.ctx), req.origin);
      joinable = pub.ok && pub.value.mode !== "invite_only";
    }
    return (
      <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={TENANT_NAV} brand={brand && { ...brand, entry: true }}>
        <Section id="entry" aria-labelledby="entry-title">
          <Kicker>Workspace</Kicker>
          <Heading level={1} id="entry-title" data-testid="tenant-name" className="mt-6 [overflow-wrap:anywhere]">
            {name}
          </Heading>
          <StatusMessage message={created ? TENANT_CREATED : undefined} className="mt-10" />
          <Text variant="lead" className="mt-10 [overflow-wrap:anywhere]">
            {signedIn ? "You are signed in." : `Sign in to continue to ${name}.`}
          </Text>
          {signedIn ? (
            <ButtonLink href="/account" className="mt-10">Account</ButtonLink>
          ) : (
            // G018 M1 (D1): the sign-in page's own guarded login form, so the entry starts sign-in directly. /sign-in
            // stays for its closed notices and the provider's logout return.
            <form method="post" action="/auth/login" className="mt-10">
              <PendingSubmit label="Sign in" pendingLabel="Signing in…" data-testid="entry-sign-in" />
            </form>
          )}
          {joinable ? (
            <div className="mt-6">
              <TextLink href="/join" data-testid="entry-join" className={LINK_FOCUS}>
                Join this workspace
              </TextLink>
            </div>
          ) : null}
        </Section>
      </AppShell>
    );
  }
  // G062: on the https profile a public entry host (the root or www.<domain>) serves the landing; the loopback profile and
  // every other host keep the generic page below unchanged.
  const entry = await publicEntryOf(req.ctx, req.origin);
  if (entry !== undefined) {
    return (
      <SignupShell productName={productName} platformBaseDomain={platformBaseDomain} nav={LANDING_NAV}>
        <Landing productName={productName} signupOrigin={entry.signupOrigin} />
      </SignupShell>
    );
  }
  return (
    <AppShell productName={productName} platformBaseDomain={platformBaseDomain} nav={NAV}>
      <Section id="overview" aria-labelledby="overview-title" containerClassName="isolate">
        <BlueprintBackdrop />
        <Kicker>Platform foundation</Kicker>
        <Heading level={1} id="overview-title" data-testid="product-name" className="mt-6 [overflow-wrap:anywhere] lg:mr-[calc(46%+1.5rem)]">
          {productName}
        </Heading>
        <Text variant="lead" className="mt-10 [overflow-wrap:anywhere] lg:mr-[calc(46%+1.5rem)]">
          The generic application shell for {productName}, served for{" "}
          <span data-testid="platform-domain">{platformBaseDomain}</span>. Features appear here as they are built.
        </Text>
      </Section>

      <Section id="getting-around" aria-labelledby="getting-around-title" tone="muted">
        <Heading level={2} id="getting-around-title">
          Getting around
        </Heading>
        <RuledGrid className="mt-12" style={REFLOW}>
          {GETTING_AROUND.map((cell, index) => (
            <div key={cell.title}>
              <Text variant="meta">{String(index + 1).padStart(2, "0")}</Text>
              <Heading level={3} className="mt-3.5 text-gold">
                {cell.title}
              </Heading>
              <Text className="mt-3.5">{cell.body}</Text>
            </div>
          ))}
        </RuledGrid>
      </Section>

      <Section id="about" aria-labelledby="about-title">
        <Heading level={2} id="about-title">
          About
        </Heading>
        <Text className="mt-6">This shell is the shared frame for the platform&apos;s pages. It holds no product features yet.</Text>
        <BulletList className="mt-6">
          {ABOUT.map((item) => (
            <BulletItem key={item}>{item}</BulletItem>
          ))}
        </BulletList>
      </Section>
    </AppShell>
  );
}
