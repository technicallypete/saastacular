// User appearance preference: Light, Dark or System (the default). It is a browser display preference only, never
// in a cookie, an account or an Organization, and it carries no authorization meaning. G069 (Plan r1 S5): the theme
// provider (components/layout/theme-provider.tsx) keeps it under APPEARANCE_STORAGE_KEY and resolves it to the light
// or dark class on <html>; System follows the OS preference live. These constants are its vocabulary for the
// appearance control and the browser specs.
export const APPEARANCE_MODES = ["light", "dark", "system"] as const;
export type AppearanceMode = (typeof APPEARANCE_MODES)[number];

export const APPEARANCE_STORAGE_KEY = "appearance-mode";
export const DEFAULT_APPEARANCE: AppearanceMode = "system";

export function parseAppearanceMode(value: unknown): AppearanceMode {
  return value === "light" || value === "dark" || value === "system" ? value : DEFAULT_APPEARANCE;
}
