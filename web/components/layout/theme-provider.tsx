"use client";

import { ThemeProvider as NextThemesProvider, useTheme } from "next-themes";
import * as React from "react";

import { DEFAULT_APPEARANCE, parseAppearanceMode } from "@/lib/appearance";

// The provider's inline script gives up without writing a class when the browser refuses storage access (its whole
// resolution sits in one try/catch), and the kit's palette would paint Light until hydration whatever the OS says.
// This fallback, served right after that script, resolves System from the OS exactly as the provider does, only when
// no class was written; it reads no storage, resolves no stored preference and sets nothing else, so the provider
// stays the one appearance manager (after hydration it applies the same resolution).
const SYSTEM_FALLBACK_SCRIPT =
  "(function(){try{var c=document.documentElement.classList;if(!c.contains('light')&&!c.contains('dark'))c.add(window.matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light')}catch(e){}})();";

// The provider validates nothing: a stored value that is not light, dark or system (another writer, an older format)
// is applied verbatim as the <html> class, by its script before the first paint and by its effect after hydration,
// and the document would keep no palette class at all. This guard, rendered inside the provider, resets such a
// preference to System through the provider's own setter (which rewrites the stored value), so the provider stays the
// one manager and no app code touches storage. A layout effect: its update is flushed in the same task as the
// provider's own effects, so the stray class is removed and the System class applied before the browser paints. The
// first paint is already right: the fallback script above resolves the OS whenever no palette class was written.
function StoredPreferenceGuard() {
  const { theme, setTheme } = useTheme();
  const stray = React.useRef<string | null>(null);
  React.useLayoutEffect(() => {
    if (theme === undefined) return;
    if (parseAppearanceMode(theme) !== theme) {
      stray.current = theme;
      setTheme(DEFAULT_APPEARANCE);
      return;
    }
    if (stray.current !== null) {
      try {
        document.documentElement.classList.remove(stray.current);
      } catch {
        // A value the class list refuses (whitespace) was never added as a class.
      }
      stray.current = null;
    }
  }, [theme, setTheme]);
  return null;
}

// G069 (Plan r1 S5): the one appearance manager, and the only app module that names next-themes. It keeps the user's
// Light, Dark or System preference under the storage key "appearance-mode" (lib/appearance.ts
// APPEARANCE_STORAGE_KEY, so every preference stored before G069 survives), puts the resolved light or dark class on
// <html> through its classic inline theme script before the first paint, and follows the OS live for System. The
// kit's theme.css sets color-scheme itself, so the provider does not. The kit's Toaster reads the same state, so
// toasts follow the user's choice. A stored value outside the three modes is reset to System (the guard above).
// Render it once, as the first child of <body> (app/layout.tsx).
export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return (
    <NextThemesProvider
      attribute="class"
      defaultTheme="system"
      enableSystem
      storageKey="appearance-mode"
      enableColorScheme={false}
    >
      <script dangerouslySetInnerHTML={{ __html: SYSTEM_FALLBACK_SCRIPT }} />
      <StoredPreferenceGuard />
      {children}
    </NextThemesProvider>
  );
}

// The appearance control reads and sets the preference through this module.
export { useTheme };
