import { notFound } from "next/navigation";

import { StatusMessage } from "@/components/account/status-message";
import { rootPage, SignupShell } from "@/components/signup/landing";
import { SignupForm } from "@/components/signup/signup-form";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import { noticeOf, SIGNUP_NOTICES } from "@/lib/signup-client";
import { readCookie } from "@/lib/session";

// G062 (Plan r5 §1.2, §4.4): the platform root's signup start, served only at the root itself (the fixed 404 on
// www.<domain>, tenant hosts and the loopback profile). One action starts the platform signup login; a browser that still
// holds a signup credential can return to its details. Server-rendered and never cached; ?status= selects fixed text
// only.
export const dynamic = "force-dynamic";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

export default async function SignupPage({ searchParams }: { searchParams: SearchParams }) {
  const r = await rootPage();
  if (r === undefined) notFound();
  const raw = (await searchParams).status;
  const notice = noticeOf(SIGNUP_NOTICES, typeof raw === "string" ? raw : undefined);
  const resumable = readCookie(r.cookie, "signup", r.profile) !== undefined;
  const { productName, platformBaseDomain } = runtimeConfig();
  return (
    <SignupShell productName={productName} platformBaseDomain={platformBaseDomain} nav={[{ href: "#signup", label: "Sign up" }]}>
      <Section id="signup" aria-labelledby="signup-title">
        <Kicker>Sign up</Kicker>
        <Heading level={1} id="signup-title" className="mt-6">
          Create your organization
        </Heading>
        <StatusMessage message={notice} className="mt-10" />
        <Text className="mt-10 [overflow-wrap:anywhere]">
          Sign in or create your {productName} identity on the identity provider&apos;s page. You then name your organization
          and choose its address, and continue there as its owner.
        </Text>
        <SignupForm />
        {resumable ? (
          <div className="mt-6">
            <TextLink href="/signup/details" data-testid="signup-resume">
              Continue your sign-up
            </TextLink>
          </div>
        ) : null}
      </Section>
    </SignupShell>
  );
}
