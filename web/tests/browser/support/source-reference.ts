import type { PaletteName } from "./modes";

// Source reference expectations (G048 Plan §4.7; re-typed in G067, and in G069 to the unaltered kit, Plan §6.2): every
// value is typed from the source declaration it cites, never read from the application's CSS. "kit <path>" is the
// adopted kit's byte-identical copy under web/src (G069 S1; the skill's kit/src, lineage and hashes in its
// ADOPTION.md); "app <path>" is an application composition around it (S12–S14: the shell header, the page's props).
// Root 16px, `--radius` 0.625rem (web/src/styles/theme.css), so `rounded-2xl` = 18px and the dialog's
// min(--radius-4xl, 24px) = 24px. Every h1–h4 takes the heading face from the kit's base layer (the brand stylesheet
// in web/src/styles: `h1, h2, h3, h4 { font-family: var(--font-heading) }`) unless a utility sets another.
// Measured at 1280x800, deviceScaleFactor 1, default text size, fonts loaded, reduced motion. Lengths are CSS px
// (±1 tolerance); letter-spacing ±0.05px; weights, family lists, border widths, radius, shadows and text-transform
// exact. Colours name a role: a saved palette name (SOURCES) or that name at a kit opacity modifier (`<name>/<percent>`,
// e.g. the kit's text-foreground/80), compared with its alpha.

// Family stacks: heading and label from the kit's `--font-heading`/`--font-label`; sans and mono from the kit's
// `--font-sans`/`--font-mono` (web/src/styles/theme.css), the bundled Geist and Geist Mono.
export const STACKS = {
  display: ["Oswald", "Arial Narrow", "sans-serif"],
  label: ["Barlow Condensed", "Arial Narrow", "sans-serif"],
  sans: ["Geist", "ui-sans-serif", "system-ui", "sans-serif"],
  mono: ["Geist Mono", "ui-monospace", "SFMono-Regular", "Menlo", "monospace"],
} as const;

// A palette name, or a palette name at a kit opacity modifier (G069 Plan §6.2: the retired opaque text roles are the
// kit's foreground/80 and muted-foreground; input/50 is the kit field and Select trigger fill, bg-input/50).
export type Role = PaletteName | "foreground/80" | "foreground/85" | "muted/40" | "input/50";

export type Expectation = {
  // Human name and selector on the production page (fixture-only rows say so).
  name: string;
  selector: string;
  cite: string;
  // Computed-style expectations: px lengths (±1), lengthExact lengths (±0.05), exact strings.
  px?: Record<string, number>;
  lengthExact?: Record<string, number>;
  exact?: Record<string, string>;
  stack?: keyof typeof STACKS;
  color?: Role;
  background?: Role;
  // border-bottom-color role.
  border?: Role;
  // max-width in ch of the element's own font (compared with a probe of that font at that many ch).
  maxWidthCh?: number;
};

export const SOURCE_REFERENCE: Expectation[] = [
  {
    name: "container",
    selector: "#overview [data-slot=container]",
    cite: "kit brand-layer layout.tsx:Container mx-auto max-w-[var(--vx-container)] px-[var(--vx-gutter)]; Section py-[var(--vx-section-y)] (web/src/styles/theme.css 1360px, clamp(20px,4vw,40px), 88px)",
    px: { "max-width": 1360, "padding-left": 40, "padding-right": 40, "padding-top": 88, "padding-bottom": 88 },
  },
  {
    name: "overview section",
    selector: "#overview",
    cite: "kit brand-layer layout.tsx:sectionVariants relative border-b (1px hairline, base layer border-border); tone default bg-background",
    exact: { "border-bottom-width": "1px", "border-bottom-style": "solid", position: "relative" },
    background: "background",
    border: "border",
  },
  {
    name: "getting-around section",
    selector: "#getting-around",
    cite: "kit brand-layer layout.tsx:sectionVariants relative border-b; tone muted bg-muted/40 (G069: the page's tone=\"muted\")",
    exact: { "border-bottom-width": "1px", "border-bottom-style": "solid", position: "relative" },
    background: "muted/40",
    border: "border",
  },
  {
    name: "about section",
    selector: "#about",
    cite: "kit brand-layer layout.tsx:sectionVariants relative border-b; tone default bg-background",
    exact: { "border-bottom-width": "1px", "border-bottom-style": "solid", position: "relative" },
    background: "background",
    border: "border",
  },
  {
    name: "kicker",
    selector: "#overview [data-slot=kicker]",
    cite: "kit brand-layer typography.tsx:Kicker font-label text-base font-semibold uppercase leading-tight tracking-[2.5px] text-gold (saved accent -> --vx-gold, Plan §2.5)",
    px: { "font-size": 16, "line-height": 20 },
    lengthExact: { "letter-spacing": 2.5 },
    exact: { "font-weight": "600", "text-transform": "uppercase" },
    stack: "label",
    color: "accent",
  },
  {
    name: "H1",
    selector: "#overview h1",
    cite: "kit brand-layer typography.tsx:headingVariants m-0 font-heading uppercase text-foreground; level 1 text-[clamp(2.6rem,5.6vw,5rem)] font-bold leading-none tracking-[-1px]",
    px: { "font-size": 71.68, "line-height": 71.68 },
    lengthExact: { "letter-spacing": -1 },
    exact: { "font-weight": "700", "text-transform": "uppercase" },
    stack: "display",
    color: "foreground",
  },
  {
    name: "H2",
    selector: "#getting-around h2",
    cite: "kit brand-layer typography.tsx:headingVariants font-heading uppercase text-foreground; level 2 text-[clamp(1.9rem,3.4vw,2.7rem)] font-semibold leading-[1.1] tracking-[-.5px] (43.2px at 1280)",
    px: { "font-size": 43.2, "line-height": 47.52 },
    lengthExact: { "letter-spacing": -0.5 },
    exact: { "font-weight": "600", "text-transform": "uppercase" },
    stack: "display",
    color: "foreground",
  },
  {
    name: "H3",
    selector: "[data-slot=ruled-grid] h3",
    cite: "kit brand-layer typography.tsx:headingVariants font-heading uppercase; level 3 text-[1.35rem] font-medium leading-[1.2] (no tracking); app/page.tsx mt-3.5 text-gold",
    px: { "font-size": 21.6, "line-height": 25.92, "margin-top": 14 },
    exact: { "font-weight": "500", "text-transform": "uppercase", "letter-spacing": "normal" },
    stack: "display",
    color: "accent",
  },
  {
    name: "lead",
    selector: "#overview h1 + [data-slot=text]",
    cite: "kit brand-layer typography.tsx:textVariants lead max-w-[52ch] text-[1.19rem] leading-[1.62] text-foreground/80 (base layer body font-sans)",
    px: { "font-size": 19.04, "line-height": 30.8448 },
    exact: { "font-weight": "400" },
    stack: "sans",
    color: "foreground/80",
    maxWidthCh: 52,
  },
  {
    name: "cell body",
    selector: "[data-slot=ruled-grid] h3 + [data-slot=text]",
    cite: "kit brand-layer typography.tsx:textVariants body max-w-[62ch] text-[1.02rem] leading-[1.65] text-muted-foreground",
    px: { "font-size": 16.32, "line-height": 26.928 },
    exact: { "font-weight": "400" },
    stack: "sans",
    color: "muted-foreground",
    maxWidthCh: 62,
  },
  {
    name: "About body",
    selector: "#about h2 + [data-slot=text]",
    cite: "kit brand-layer typography.tsx:textVariants body max-w-[62ch] text-[1.02rem] leading-[1.65] text-muted-foreground",
    px: { "font-size": 16.32, "line-height": 26.928 },
    exact: { "font-weight": "400" },
    stack: "sans",
    color: "muted-foreground",
    maxWidthCh: 62,
  },
  {
    name: "footer meta",
    selector: "footer [data-slot=text]",
    cite: "kit brand-layer typography.tsx:textVariants meta font-label text-[1.08rem] leading-[1.35] text-muted-foreground (app app-shell.tsx footer)",
    px: { "font-size": 17.28, "line-height": 23.328 },
    exact: { "font-weight": "400" },
    stack: "label",
    color: "muted-foreground",
  },
  {
    // G069 Plan S12: the grid's index numbers render as the kit's meta text (its index variant, muted-foreground/80
    // at about 3.8:1, is not used).
    name: "index (meta)",
    selector: "[data-slot=ruled-grid] [data-slot=text]:first-child",
    cite: "kit brand-layer typography.tsx:textVariants meta font-label text-[1.08rem] leading-[1.35] text-muted-foreground (app/page.tsx variant=\"meta\")",
    px: { "font-size": 17.28, "line-height": 23.328 },
    exact: { "font-weight": "400", "letter-spacing": "normal" },
    stack: "label",
    color: "muted-foreground",
  },
  {
    name: "ruled grid",
    selector: "[data-slot=ruled-grid] > div",
    cite: "kit brand-layer layout.tsx:RuledGrid grid gap-px overflow-visible rounded-2xl border bg-border (app/page.tsx S14 style reflows the columns)",
    px: { "column-gap": 1, "row-gap": 1 },
    exact: {
      "border-top-width": "1px",
      "border-bottom-width": "1px",
      "border-top-style": "solid",
      "border-bottom-style": "solid",
      "border-top-left-radius": "18px",
    },
    background: "border",
    border: "border",
  },
  {
    name: "ruled cell",
    selector: "[data-slot=ruled-grid] > div > div",
    cite: "kit brand-layer layout.tsx:RuledGrid [&>*]:bg-background [&>*]:px-8 [&>*]:pt-8 [&>*]:pb-9",
    px: { "padding-top": 32, "padding-right": 32, "padding-bottom": 36, "padding-left": 32 },
    background: "background",
  },
  {
    name: "bullet list",
    selector: "#about [data-slot=bullet-list]",
    cite: "kit brand-layer content.tsx:BulletList m-0 grid list-none gap-3 p-0",
    px: { "row-gap": 12 },
  },
  {
    name: "bullet item",
    selector: "#about [data-slot=bullet-item]",
    cite: "kit brand-layer content.tsx:BulletItem flex gap-3.5 text-sm leading-snug text-foreground/85",
    px: { "column-gap": 14, "font-size": 14, "line-height": 19.25 },
    exact: { "font-weight": "400" },
    stack: "sans",
    color: "foreground/85",
  },
  {
    name: "bullet marker",
    selector: "#about [data-slot=bullet-item] > span[aria-hidden]",
    cite: "kit brand-layer content.tsx:BulletItem marker mt-[9px] h-[2px] w-[14px] rounded-full bg-gold (a gold fill, no border)",
    px: { width: 14, height: 2, "margin-top": 9 },
    exact: { "border-top-width": "0px" },
    background: "accent",
  },
  {
    name: "header",
    selector: "[data-slot=site-header]",
    cite: "app web/components/layout/app-shell.tsx header relative border-b border-border bg-background (S13: the app's static header on the kit Container)",
    exact: { "border-bottom-width": "1px", "border-bottom-style": "solid", position: "relative" },
    background: "background",
    border: "border",
  },
  {
    name: "header container",
    selector: "[data-slot=site-header] > [data-slot=container]",
    cite: "app web/components/layout/app-shell.tsx Container flex min-h-[72px] flex-wrap items-center justify-between gap-x-10 gap-y-4 py-3",
    px: { "min-height": 72, "padding-top": 12, "padding-bottom": 12, "column-gap": 40, "row-gap": 16 },
  },
  {
    name: "nav",
    selector: "[data-slot=site-header] nav",
    cite: "app web/components/layout/app-shell.tsx nav flex flex-wrap gap-x-8 gap-y-2 font-sans text-sm font-medium",
    px: { "column-gap": 32, "row-gap": 8 },
  },
  {
    name: "nav link",
    selector: "[data-slot=site-header] nav a",
    cite: "app web/components/layout/app-shell.tsx nav a inline-flex min-h-8 items-center text-foreground/80 (text-sm: 14px / 20px)",
    px: { "font-size": 14, "line-height": 20, "min-height": 32 },
    exact: { "font-weight": "500", "text-transform": "none", "letter-spacing": "normal" },
    stack: "sans",
    color: "foreground/80",
  },
  {
    // The everyday Button at size sm (the angled call to action is the kit CtaButton's alone, S24).
    name: "help trigger",
    selector: "[data-slot=site-header] button[aria-haspopup=dialog]",
    cite: "kit web/src/components/ui/button.tsx:buttonVariants rounded-2xl border border-transparent bg-clip-padding text-sm font-medium; size sm h-7 px-3; default bg-primary text-primary-foreground",
    px: { "padding-top": 0, "padding-bottom": 0, "padding-left": 12, "padding-right": 12, "font-size": 14, "line-height": 20, height: 28 },
    exact: {
      "font-weight": "500",
      "text-transform": "none",
      "letter-spacing": "normal",
      "clip-path": "none",
      "box-shadow": "none",
      "border-top-left-radius": "18px",
      "border-bottom-right-radius": "18px",
      "border-top-width": "1px",
      "border-top-style": "solid",
    },
    stack: "sans",
    color: "primary-foreground",
    background: "primary",
  },
  {
    // The kit's ring-1 ring-foreground/5 and shadow-xl replace a border (S-DESIGN, S-PALETTE measure them).
    name: "dialog content",
    selector: "[data-slot=dialog-content]",
    cite: "kit web/src/components/ui/dialog.tsx:DialogContent bg-popover text-popover-foreground grid max-w-[calc(100%-2rem)] gap-6 rounded-[min(var(--radius-4xl),24px)] p-6 text-sm shadow-xl ring-1 sm:max-w-md w-full",
    px: { width: 448, "padding-top": 24, "padding-right": 24, "padding-bottom": 24, "padding-left": 24, "row-gap": 24 },
    exact: {
      "border-top-left-radius": "24px",
      "border-bottom-right-radius": "24px",
      "border-top-width": "0px",
      "border-right-width": "0px",
      "border-bottom-width": "0px",
      "border-left-width": "0px",
    },
    background: "popover",
  },
  {
    // Base UI's Dialog.Title is an <h2>, so the kit's base layer gives it the heading face; the kit sets no case.
    name: "dialog title",
    selector: "[data-slot=dialog-title]",
    cite: "kit web/src/components/ui/dialog.tsx:DialogTitle text-base leading-none font-medium (an h2: base layer font-heading; inherits text-popover-foreground)",
    px: { "font-size": 16, "line-height": 16 },
    exact: { "font-weight": "500", "text-transform": "none", "letter-spacing": "normal" },
    stack: "display",
    color: "popover-foreground",
  },
  {
    name: "dialog description",
    selector: "[data-slot=dialog-description]",
    cite: "kit web/src/components/ui/dialog.tsx:DialogDescription text-muted-foreground text-sm",
    px: { "font-size": 14, "line-height": 20 },
    exact: { "font-weight": "400" },
    stack: "sans",
    color: "muted-foreground",
  },
  {
    // G069: the kit's DialogHeader has no right padding; that the Close never covers the title or description is
    // measured geometrically (S-FIDELITY at 1280 and 320; the text-spacing modal check).
    name: "dialog header",
    selector: "[data-slot=dialog-header]",
    cite: "kit web/src/components/ui/dialog.tsx:DialogHeader gap-1.5 flex flex-col",
    px: { "row-gap": 6, "padding-right": 0 },
  },
  {
    // G069 owner kit-fidelity direction (2026-10-26): the kit DialogContent's own corner Close (showCloseButton), whose
    // ghost Button sets no text colour, so the cross inherits the popup's popover-foreground on the bg-secondary fill.
    name: "dialog close",
    selector: "[data-slot=dialog-content] > [data-slot=dialog-close]",
    cite: "kit web/src/components/ui/dialog.tsx:DialogContent showCloseButton: DialogPrimitive.Close render Button variant=\"ghost\" size=\"icon-sm\" absolute top-4 right-4 bg-secondary (kit button.tsx icon-sm size-7, rounded-2xl; no text colour: inherits text-popover-foreground)",
    px: { width: 28, height: 28, top: 16, right: 16 },
    exact: { "border-top-left-radius": "18px" },
    color: "popover-foreground",
    background: "secondary",
  },
  {
    name: "dialog close icon",
    selector: "[data-slot=dialog-content] > [data-slot=dialog-close] > svg",
    cite: "kit web/src/components/ui/dialog.tsx:DialogContent XIcon; kit button.tsx [&_svg:not([class*='size-'])]:size-4",
    px: { width: 16, height: 16 },
  },
  {
    name: "dialog close stroke (first)",
    selector: "[data-slot=dialog-content] > [data-slot=dialog-close] > svg > path:first-of-type",
    cite: "kit web/src/components/ui/dialog.tsx:DialogContent XIcon (lucide x: two round-capped 2-unit strokes in currentColor, the inherited popover-foreground)",
    exact: { "stroke-width": "2px", "stroke-linecap": "round" },
    color: "popover-foreground",
  },
  {
    name: "dialog close stroke (second)",
    selector: "[data-slot=dialog-content] > [data-slot=dialog-close] > svg > path:last-of-type",
    cite: "kit web/src/components/ui/dialog.tsx:DialogContent XIcon (lucide x: two round-capped 2-unit strokes in currentColor, the inherited popover-foreground)",
    exact: { "stroke-width": "2px", "stroke-linecap": "round" },
    color: "popover-foreground",
  },
  {
    // G069 S16: the kit's theme-fixed scrim with its blur is accepted (backdrop-blur-sm, Tailwind's --blur-sm 8px).
    name: "overlay",
    selector: "[data-slot=dialog-overlay]",
    cite: "kit web/src/components/ui/dialog.tsx:DialogOverlay bg-black/30 supports-backdrop-filter:backdrop-blur-sm fixed inset-0 isolate z-50",
    exact: { "backdrop-filter": "blur(8px)" },
  },
];

// The skip link while focused (the kit TextLink with the shell's gutter, top padding and focus outline).
export const SKIP_LINK: Expectation = {
  name: "skip link (focused)",
  selector: 'a[href="#main"]',
  cite: "kit brand-layer typography.tsx:TextLink inline-block border-b-2 border-gold pb-1 font-heading text-base uppercase tracking-[1.5px] text-gold (app app-shell.tsx px-[var(--vx-gutter)] pt-3)",
  px: { "font-size": 16, "padding-bottom": 4 },
  lengthExact: { "letter-spacing": 1.5 },
  exact: { "text-transform": "uppercase", "font-weight": "400", "border-bottom-width": "2px", "border-bottom-style": "solid" },
  stack: "display",
  color: "accent",
  border: "accent",
};

// Button sizes on the fixture page (the kit's base-rhea Button, web/src/components/ui/button.tsx size): sm h-7 28px,
// default h-8 32px, lg h-9 36px, no vertical padding, px-3 / px-3 / px-4, text-sm (14px, line-height 20px).
export const BUTTON_SIZES: Record<"sm" | "default" | "lg", { height: number; py: number; px: number; fontSize: number; lineHeight: number }> = {
  sm: { height: 28, py: 0, px: 12, fontSize: 14, lineHeight: 20 },
  default: { height: 32, py: 0, px: 12, fontSize: 14, lineHeight: 20 },
  lg: { height: 36, py: 0, px: 16, fontSize: 14, lineHeight: 20 },
};
// The icon size: size-8 (32px).
export const BUTTON_ICON = { width: 32, height: 32 };
// The source angle: 15% cuts (web/src/styles/theme.css --vx-angle), drawn only by the kit CtaButton (its vx-angle
// utility, clip-path: var(--vx-angle)), which renders only under the platform palette (S24).
export const SOURCE_ANGLE = "polygon(15% 0%, 100% 0%, 100% 35%, 85% 100%, 0% 100%, 0% 65%)";

// Responsive expectations (§4.7). 320: the gutter's clamp minimum, the H1 at 2.6rem, the H2 at 1.9rem (the kit's clamp
// minimum), the dialog at w-full max-w-[calc(100%-2rem)] (288px) with p-6 (240px content). 640x400: 4vw gutter.
export const RESPONSIVE = {
  narrow: { width: 320, height: 640, gutter: 20, h1: 41.6, h2: 30.4, dialogWidth: 288, dialogContent: 240 },
  zoomed: { width: 640, height: 400, gutter: 25.6 },
} as const;
