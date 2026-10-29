import { notFound, redirect } from "next/navigation";
import type { ReactNode } from "react";

import { StatusMessage } from "@/components/account/status-message";
import { AbandonForm, DetailsForm, NameForm, RetryForm } from "@/components/signup/details-form";
import { rootPage, SignupShell } from "@/components/signup/landing";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { DETAILS_NOTICES, fieldErrorOf, noticeOf, recoveryView, signupStatus } from "@/lib/signup-client";
import { readCookie, readSignupFormCookie } from "@/lib/session";

// G062 (Plan r5 §1.4, §4.4; Codex Round 14 option A): the root signup details, served only at the platform root. Core's
// status read (with the browser's own signup credential) decides what the page shows on every request:
// - no signup cookie: to /signup; a dead credential: to GET /signup/abandon, which clears the dead cookies (a Server
//   Component cannot) and continues to /signup;
// - pending: the name, then the editable suggested short name (the slug; values only from the browser's form restoration
//   cookie). G057 (Plan r5 §3.3): the slug is metadata and no web address is previewed from it;
// - already provisioned (with or without a ?status=, for example continue_failed): the closed "Your organization is
//   ready" text with core's target_origin (the generated label origin, structurally validated by signupStatus), one
//   user-triggered "Try again" POST (identical recorded payload; only when the restoration record names the created
//   slug) and the manual fallback of signing in at the new tenant. The happy path never renders this state: the first
//   create's POST answers straight into the continuation.
// Rendering never mutates (G062 bounded security correction to Plan r5 §4.4, Codex Round 15 decision 2): the only core
// call here is the status read; no create, re-mint or abandon follows from a GET, so a cross-site top-level navigation
// carrying the SameSite=Lax root cookie changes nothing.
// Server-rendered and never cached; ?status= selects fixed text only and nothing submitted is ever in a URL.
export const dynamic = "force-dynamic";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

const NAV = [{ href: "#signup-details" as const, label: "Your organization" }];

export default async function SignupDetailsPage({ searchParams }: { searchParams: SearchParams }) {
  const r = await rootPage();
  if (r === undefined) notFound();
  const raw = (await searchParams).status;
  const status = typeof raw === "string" ? raw : undefined;
  const credential = readCookie(r.cookie, "signup", r.profile);
  if (credential === undefined) redirect("/signup");
  const view = await signupStatus(r.ctx, credential, r.origin);
  if (!view.ok && view.error === "unauthenticated") redirect("/signup/abandon");
  const form = readSignupFormCookie(r.cookie, r.profile);
  let title = "Name your organization";
  let body: ReactNode;
  if (!view.ok) {
    body = (
      <StatusMessage
        message={view.error === "rate_limited" ? DETAILS_NOTICES.rate_limited : "Your sign-up cannot be shown right now. Please try again later."}
        className="mt-10"
      />
    );
  } else if (view.value.status === "pending") {
    const csrfToken = view.value.csrfToken;
    const refusal = fieldErrorOf(status);
    const notice = refusal === undefined ? noticeOf(DETAILS_NOTICES, status) : undefined;
    if (form === undefined) {
      body = (
        <>
          <StatusMessage message={notice} className="mt-10" />
          <Text className="mt-10">Step 1 of 2. You will own the new organization. Next, you choose its short name.</Text>
          <NameForm csrfToken={csrfToken} error={refusal?.field === "name" ? refusal.message : undefined} />
          <AbandonForm csrfToken={csrfToken} />
        </>
      );
    } else {
      title = "Choose its short name";
      body = (
        <>
          <StatusMessage message={notice} className="mt-10" />
          <Text className="mt-10">{"Step 2 of 2. We'll suggest a unique short name based on your organization name. You can change it."}</Text>
          <DetailsForm
            csrfToken={csrfToken}
            name={form.name}
            slug={form.slug}
            nameError={refusal?.field === "name" ? refusal.message : undefined}
            slugError={refusal?.field === "slug" ? refusal.message : undefined}
          />
          <AbandonForm csrfToken={csrfToken} />
        </>
      );
    }
  } else {
    const { csrfToken, organization } = view.value;
    // Bounded security correction (Codex Round 15 decision 2): a GET never re-mints. The recovery view only offers the
    // user-triggered "Try again" POST when the restoration record names the created address (core's intent digest still
    // decides whether it is the identical payload).
    const { retryable, notice } = recoveryView(status, form, organization);
    title = "Your organization is ready";
    body = (
      <>
        <StatusMessage message={notice} className="mt-10" />
        <Text className="mt-10 [overflow-wrap:anywhere]">
          {organization === null ? "Your organization was created." : `Your organization was created at ${organization.targetOrigin}.`}
        </Text>
        {retryable ? <RetryForm csrfToken={csrfToken} /> : null}
        {organization === null ? null : (
          <div className="mt-6">
            <TextLink href={`${organization.targetOrigin}/sign-in`} data-testid="signup-tenant-sign-in">
              Sign in at your organization instead
            </TextLink>
          </div>
        )}
        <AbandonForm csrfToken={csrfToken} />
      </>
    );
  }
  const { productName, platformBaseDomain } = runtimeConfig();
  return (
    <SignupShell productName={productName} platformBaseDomain={platformBaseDomain} nav={NAV}>
      <Section id="signup-details" aria-labelledby="signup-details-title">
        <Kicker>Sign up</Kicker>
        <Heading level={1} id="signup-details-title" className="mt-6">
          {title}
        </Heading>
        {body}
      </Section>
    </SignupShell>
  );
}
