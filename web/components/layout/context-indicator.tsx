import { ButtonLink } from "@/components/controls/button-link";
import { Text } from "@/components/vestive/typography";
import type { Indicator } from "@/lib/branding";

// G017 (Plan r2 §1.5, D-6): the tenant header's current-Organization indicator, rendered from core facts only
// (indicatorState in lib/branding.ts): anonymous shows the one neutral Sign in action; a signed-in member or non-member
// sees who they are and whether they belong to this workspace, with the Account link (sign-out lives there). No
// Organization listing, Group picker or tenant switch. An unavailable session shows no identity, only the Account link.
export function ContextIndicator({ indicator }: { indicator: Indicator }) {
  if (indicator.kind === "anonymous") {
    return (
      <ButtonLink href="/sign-in" data-testid="context-sign-in" size="sm">
        Sign in
      </ButtonLink>
    );
  }
  return (
    <div className="flex min-w-0 flex-wrap items-center gap-4">
      {indicator.kind === "unavailable" ? null : (
        <Text variant="meta" data-testid="context-indicator" data-membership={indicator.kind} className="min-w-0 [overflow-wrap:anywhere]">
          {indicator.text}
        </Text>
      )}
      <ButtonLink href="/account" size="sm">Account</ButtonLink>
    </div>
  );
}
