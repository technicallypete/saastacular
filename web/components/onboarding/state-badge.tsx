import { Badge } from "@/components/ui/badge";

// G018 (Plan r5 §1.6, D-2): a Group's onboarding state, derived by core from whether a Business Profile is attached.
// The text carries the state; the variant is decoration only (G069: the kit Badge, filled when complete, outlined
// when incomplete).
export function StateBadge({ complete }: { complete: boolean }) {
  return (
    <Badge variant={complete ? "default" : "outline"} data-testid="group-state">
      {complete ? "Complete" : "Incomplete"}
    </Badge>
  );
}
