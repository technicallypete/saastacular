// The saved-branding colour names (docs/foundation/WEB-SHELL.md; core branding.Tokens). G069 (Plan r1 S7–S9): the
// palette itself is the adopted kit's, unaltered (src/styles/theme.css); a saved name sets the kit variables of the
// lib/tenant-css.ts fan-out.
export const SEMANTIC_COLOR_TOKENS = [
  "background",
  "foreground",
  "primary",
  "primary-foreground",
  "accent",
  "muted-foreground",
  "border",
  "popover",
  "popover-foreground",
  "muted",
  "secondary",
  "secondary-foreground",
  "destructive",
  "input",
  "ring",
] as const;
export type SemanticColorToken = (typeof SEMANTIC_COLOR_TOKENS)[number];

export type TokenValues = Partial<Record<SemanticColorToken, string>>;

// Saved overrides for each appearance. A missing value keeps the platform default, so an override that sets only
// Light values keeps the platform's Dark palette.
export type TokenOverrides = { light?: TokenValues; dark?: TokenValues };
