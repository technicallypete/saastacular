"use client";

import { Monitor, Moon, Sun, type LucideIcon } from "lucide-react";
import * as React from "react";

import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger } from "@/components/ui/select";
import { APPEARANCE_MODES, DEFAULT_APPEARANCE, parseAppearanceMode, type AppearanceMode } from "@/lib/appearance";

import { useTheme } from "./theme-provider";

const LABELS: Record<AppearanceMode, string> = { light: "Light", dark: "Dark", system: "System" };
const SYMBOLS: Record<AppearanceMode, LucideIcon> = { light: Sun, dark: Moon, system: Monitor };

const subscribeNever = () => () => {};

// False for the server render and the hydration pass, true afterwards (no effect and no state update).
function useHydrated(): boolean {
  return React.useSyncExternalStore(
    subscribeNever,
    () => true,
    () => false,
  );
}

// The kit's compound Select (owner preview finding, G069 r2: the control's own list, not the browser's). The closed
// trigger, the combobox named "Appearance", shows only the selected preference's symbol: Sun (Light), Moon (Dark) or
// Monitor (System, also while the OS resolves System to Light or Dark), with the kit's chevron. The selected
// preference's label (Light, Dark or System) is the combobox's text value for assistive technology, visually hidden
// (sr-only) so the shipped icon-only trigger is unchanged; data-appearance-value mirrors it for tests. Its popup portals into document.body
// (the kit's select-content, right-aligned under the trigger) and lists the modes with their symbols inside the kit's
// padded group. Enter, Space or an arrow key opens it, the arrow keys move, Enter chooses, Escape closes and returns
// focus. G069 (Plan r1 S5, S6): the preference lives in the theme provider. The server cannot know it, so until
// hydration the symbol follows the palette already resolved on <html> (Sun or Moon, by the kit's dark variant) and the
// value is the default; after hydration both follow the stored preference.
export function AppearanceControl() {
  const { theme, setTheme } = useTheme();
  const hydrated = useHydrated();
  const mode = hydrated ? parseAppearanceMode(theme) : DEFAULT_APPEARANCE;
  const icon = (option: AppearanceMode) => {
    if (hydrated) return option === mode ? "block size-4" : "hidden";
    if (option === "light") return "block size-4 dark:hidden";
    if (option === "dark") return "hidden size-4 dark:block";
    return "hidden";
  };
  return (
    <span data-appearance-control="" className="inline-flex">
      <Select value={mode} onValueChange={(value) => setTheme(parseAppearanceMode(value))}>
        <SelectTrigger aria-label="Appearance" data-appearance-value={mode}>
          <Sun data-appearance-icon="light" className={icon("light")} />
          <Moon data-appearance-icon="dark" className={icon("dark")} />
          <Monitor data-appearance-icon="system" className={icon("system")} />
          <span data-appearance-label="" className="sr-only">
            {LABELS[mode]}
          </span>
        </SelectTrigger>
        <SelectContent align="end" alignItemWithTrigger={false}>
          <SelectGroup>
            {APPEARANCE_MODES.map((option) => {
              const Icon = SYMBOLS[option];
              return (
                <SelectItem key={option} value={option}>
                  <Icon />
                  {LABELS[option]}
                </SelectItem>
              );
            })}
          </SelectGroup>
        </SelectContent>
      </Select>
    </span>
  );
}
