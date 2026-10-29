import { StatusMessage } from "@/components/account/status-message";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { AddressCheck } from "@/components/onboarding/address-check";
import { FormField } from "@/components/onboarding/form-field";
import { onboardingRequest, param, type SearchParams } from "@/components/onboarding/onboarding-frame";
import { Button } from "@/components/ui/button";
import { FieldDescription } from "@/components/ui/field";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker, Text, TextLink } from "@/components/vestive/typography";
import { runtimeConfig } from "@/lib/config";
import {
  ADDRESS_ADVISORY,
  codePoints,
  CREATE_CONFLICT,
  FIELD_LABELS,
  fieldError,
  isSlug,
  listGroups,
  MAX_NAME,
  MAX_SLUG,
  suggestSlug,
  type GroupView,
} from "@/lib/onboarding-client";

// G018 (Plan r5 §1.4, §1.6): create a Group in two steps on one page. Step 1 takes the name (a GET form); step 2 shows
// core's suggested, editable address (slug) for that name and posts both to /groups/create. A name is echoed only after
// core accepted it for a suggestion; a refused value is never echoed. Server-rendered and never cached. G018 M1: core's
// definitive conflict on the address returns here as slug_taken with the refused (grammar-valid) address, which a fresh
// authorized read resolves to a holder only when the viewer can read it; the slug field carries the advisory hint.
export const dynamic = "force-dynamic";

const INVALID_NAME = fieldError("invalid_name")?.message;
const ADVISORY_ID = "slug-advisory";

export default async function NewGroup({ searchParams }: { searchParams: SearchParams }) {
  const r = await onboardingRequest({ config: runtimeConfig(), area: "groups", kicker: "New Group", nav: [{ href: "#new-group", label: "New Group" }] });
  if (!r.ready) return r.page;
  const params = await searchParams;
  // A create refused by core for a field returns here with its closed code (never the submitted value).
  const refusal = fieldError(param(params.error));
  const rawName = param(params.name);
  let nameError = refusal?.field === "name" ? refusal.message : undefined;
  let step2: { name: string; slug: string } | undefined;
  if (nameError === undefined && rawName !== undefined && rawName !== "") {
    if (codePoints(rawName) > MAX_NAME) {
      nameError = INVALID_NAME;
    } else {
      const suggestion = await suggestSlug(r.lookup(), r.credential, { origin: r.origin, csrfToken: r.csrfToken }, rawName);
      if (suggestion.ok) step2 = { name: rawName, slug: suggestion.value };
      else if (suggestion.error === "invalid_request" && fieldError(suggestion.field)?.field === "name") nameError = fieldError(suggestion.field)?.message;
      else return r.refused(suggestion.error);
    }
  }
  const slugError = step2 !== undefined && refusal?.field === "slug" ? refusal.message : undefined;
  // slug_taken (core's conflict): the refused address is shown again only when grammar-valid, and its holder only when a
  // fresh authorized read returns it (a failed or empty read shows none).
  const taken = step2 !== undefined && param(params.error) === "slug_taken";
  const echoed = param(params.slug);
  const takenSlug = taken && isSlug(echoed) ? echoed : undefined;
  let holder: GroupView | undefined;
  if (takenSlug !== undefined) {
    const read = await listGroups(r.lookup(), r.credential, r.origin, { slug: takenSlug, limit: 1 });
    holder = read.ok ? read.value.groups[0] : undefined;
  }
  const summary = taken
    ? CREATE_CONFLICT.summary
    : nameError !== undefined || slugError !== undefined
      ? "A value was not in the accepted form, so nothing was created."
      : undefined;

  return r.frame(
    <Section id="new-group" aria-labelledby="new-group-title">
      <Kicker>New Group</Kicker>
      <Heading level={1} id="new-group-title" className="mt-6">
        {step2 === undefined ? "Name your Group" : "Choose its address"}
      </Heading>
      <Text className="mt-10">
        {step2 === undefined
          ? "Step 1 of 2. You will own the new Group. Next, you choose its address, then attach a Business Profile."
          : "Step 2 of 2. The address is suggested from the name; you can change it. It is unique in this workspace."}
      </Text>
      <StatusMessage message={summary} className="mt-10" />
      {step2 === undefined ? (
        <form method="get" action="/groups/new" className="mt-12 grid max-w-[40rem] gap-8" data-testid="group-name-form">
          <FormField name="name" label={FIELD_LABELS.name} required error={nameError} />
          <div>
            <Button type="submit">Continue</Button>
          </div>
        </form>
      ) : (
        <form method="post" action="/groups/create" className="mt-12 grid max-w-[40rem] gap-8" data-testid="group-create-form">
          <input type="hidden" name="csrf_token" value={r.csrfToken} />
          <FormField name="name" label={FIELD_LABELS.name} required defaultValue={step2.name} />
          <AddressCheck copy={ADDRESS_ADVISORY} maxSlug={MAX_SLUG} advisoryId={ADVISORY_ID}>
            <FormField
              name="slug"
              label={FIELD_LABELS.slug}
              required
              hint="Lowercase letters, digits and hyphens, up to 63 characters."
              error={slugError}
              defaultValue={takenSlug ?? step2.slug}
              spellCheck={false}
              describedBy={ADVISORY_ID}
            >
              {holder === undefined ? null : (
                <FieldDescription data-testid="address-holder" className="[overflow-wrap:anywhere]">
                  {CREATE_CONFLICT.holder} {holder.name}. <TextLink href={`/groups/${holder.id}`}>Open {holder.name}</TextLink>{" "}
                  {CREATE_CONFLICT.interrupted}
                </FieldDescription>
              )}
            </FormField>
          </AddressCheck>
          <div className="flex flex-wrap items-center gap-8">
            <PendingSubmit label="Create Group" pendingLabel="Creating the Group…" />
            <TextLink href="/groups/new">Change the name</TextLink>
          </div>
        </form>
      )}
    </Section>,
  );
}
