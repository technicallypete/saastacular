import type * as React from "react";

import { previewVariables } from "@/lib/tenant-css";
import type { TokenValues } from "@/lib/tokens";
import { cn } from "@/lib/utils";

type BrandPreviewProps = Omit<React.ComponentProps<"div">, "style"> & {
  // A styling and test label only. Never an identifier, an authority or an Organization selector.
  label: string;
  // The one appearance this surface shows, whatever the page's appearance.
  appearance: "light" | "dark";
  // That appearance's saved branding values (re-checked; anything invalid previews the defaults).
  tokens?: TokenValues;
};

// G069 (Plan r1 S8): an isolated branding preview. The kit's palette rules match any element with the light or dark
// class, so the surface takes its appearance from that class and the saved values as inline custom properties
// (lib/tenant-css.ts fan-out) for its own subtree only; the document root and every other surface are unaffected.
// The app's dark variant (app/globals.css) excludes an element carrying data-brand-preview with the light class and its
// subtree, so a Light preview on a Dark page takes the Light utilities too; keep both on this one element.
// Rule: a preview renders no portaled component (dialog, sheet, popover, select, tooltip, menu, toast). Portals land in
// document.body, outside this element, with the live tenant's values.
export function BrandPreview({ label, appearance, tokens, className, children, ...props }: BrandPreviewProps) {
  return (
    <div
      data-brand-preview={label}
      className={cn(appearance, "bg-background text-foreground", className)}
      style={previewVariables(tokens) as React.CSSProperties}
      {...props}
    >
      {children}
    </div>
  );
}
