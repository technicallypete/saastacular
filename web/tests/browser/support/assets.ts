// Independent pins (G048; Geist and Geist Mono G067): the SHA-256 of every bundled public asset, recorded at intake
// against the upstream git blob SHA-1 and size (docs/foundation/WEB-SHELL.md), never computed from the served files.
export const FONT_SHA256 = {
  "/fonts/oswald/oswald-wght.ttf": "5b38c246e255a12f5712d640d56bcced0472466fc68983d2d0410ec0457c2817",
  "/fonts/barlow-condensed/barlow-condensed-400.ttf": "583cec5da3b84bc4dc7c9c72e2a565c94d34e431518b19d7e250b7830ad5f996",
  "/fonts/barlow-condensed/barlow-condensed-600.ttf": "7b619d14bc2327509a9ef32b0890f709626f7ecc9ff61191c2a4314c5499d2d9",
  "/fonts/geist/geist-wght.ttf": "73894e0448cae90a92b6c2f8732b7bb9acb7b94c418bff559dad4a18e1de9659",
  "/fonts/geist-mono/geist-mono-wght.ttf": "d00e590b8eb3a59acc329b2d044fd143ae935090b7da33199ebee27cc7de8196",
} as const;

export const NOTICE_SHA256 = {
  "/fonts/oswald/OFL.txt": "0fd731a904b729a4e02eaf5e8ebd06783edd9abe400e8882760160230675b652",
  "/fonts/barlow-condensed/OFL.txt": "186d750eb496a4c17a76385f82be6aea2ac1cf2de074a811d63786cf374ea73f",
  "/fonts/geist/OFL.txt": "1781d2806a07d91c4edf4740b88449fab7d0eadad53f7c351b94cd4d4eb8c00f",
  "/fonts/geist-mono/OFL.txt": "1781d2806a07d91c4edf4740b88449fab7d0eadad53f7c351b94cd4d4eb8c00f",
} as const;

export const BLUEPRINT_PATH = "/blueprint/overview.svg";
export const BLUEPRINT_SHA256 = "942a2a242594773f6be7e638d6e07f9f723a0ff6b6c6e8334427702e52f98c65";

export type FontPath = keyof typeof FONT_SHA256;
export const FONT_PATHS = Object.keys(FONT_SHA256) as FontPath[];

// The binaries the application shell's pages request (G069): their text is the kit's Oswald headings, Barlow Condensed
// kicker and meta lines (400 and 600) and Geist; no shell text is mono since the kit's index text variant is not used
// (Plan S12: the four index sites render as meta), so the browser never requests Geist Mono there. Geist Mono's bytes,
// metadata and rendering stay covered on the fixture page, which renders every bundled face (tokens.spec.ts S-FONTS,
// probe geist-mono-400) and is gated on all of FONT_PATHS.
export const SHELL_FONT_PATHS = FONT_PATHS.filter((p) => p !== "/fonts/geist-mono/geist-mono-wght.ttf");

// What each binary must declare about itself (name table ID 1 / 16 family, fvar axes, OS/2 weight).
export const FONT_FACTS: Record<FontPath, { family: string; axes?: Record<string, [number, number]>; weightClass?: number }> = {
  "/fonts/oswald/oswald-wght.ttf": { family: "Oswald", axes: { wght: [200, 700] } },
  "/fonts/barlow-condensed/barlow-condensed-400.ttf": { family: "Barlow Condensed", weightClass: 400 },
  "/fonts/barlow-condensed/barlow-condensed-600.ttf": { family: "Barlow Condensed", weightClass: 600 },
  "/fonts/geist/geist-wght.ttf": { family: "Geist", axes: { wght: [100, 900] } },
  "/fonts/geist-mono/geist-mono-wght.ttf": { family: "Geist Mono", axes: { wght: [100, 900] } },
};
