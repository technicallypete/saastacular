import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

const config = [
  ...nextVitals,
  ...nextTs,
  { ignores: [".next/**", "node_modules/**", "next-env.d.ts"] },
  // G067: the adopted brand agent skill (design input, never web source) and its Claude discovery alias.
  { ignores: [".claude/skills/vestive-brand/**", ".agents/skills/vestive-brand/**"] },
  // G069 S18: unaltered kit source; the app adapts, never the kit. The copied kit tree under src/ stays linted and
  // typechecked; only the rules it reports are off there: no-img-element (the brand content and header modules' plain
  // <img>) and set-state-in-effect (the carousel's initial selection and the mobile hook's initial width).
  {
    files: ["src/**"],
    rules: {
      "@next/next/no-img-element": "off",
      "react-hooks/set-state-in-effect": "off",
    },
  },
];

export default config;
