import type { CSSProperties } from "react";

// G069 S14 (D4): the style prop the app passes to the unaltered kit RuledGrid. The kit keeps its column count at every
// width and scrolls sideways below its minWidth; these props (the kit spreads `style` last) let the cells reflow
// instead, one column at 320px, so the page never scrolls horizontally (S-RESPONSIVE).
export const RULED_GRID_REFLOW: CSSProperties = {
  gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 16rem), 1fr))",
  minWidth: 0,
};
