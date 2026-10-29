import * as React from "react";

import { cn } from "@/lib/utils";

type ScrollRegionProps = {
  /** The region's accessible name, for example "Members". */
  label: string;
  className?: string;
  children: React.ReactNode;
};

// G069 S12: a focusable, labelled horizontal scroll region around a kit Table, so a genuine data table may scroll
// at 320px and keyboard users can reach and scroll it. The kit Table's own container is made non-scrolling here, so
// this region, the focusable element, is the one that scrolls. Focus shows a solid 2px ring outline (the kit's call
// to action treatment), since the region has no border for the kit's field ring.
export function ScrollRegion({ label, className, children }: ScrollRegionProps) {
  return (
    <div
      role="region"
      aria-label={label}
      tabIndex={0}
      className={cn(
        "relative w-full overflow-x-auto focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring focus-visible:outline-solid [&>[data-slot=table-container]]:overflow-visible",
        className
      )}
    >
      {children}
    </div>
  );
}
