import { createHash } from "node:crypto";
import { appendFileSync } from "node:fs";
import { dirname } from "node:path";

// Evidence hygiene for the identity browser tests (G004 Plan r5 §8.3). Bearer and protocol values the tests observe
// (session and login-binding cookies, codes, states, logout references, CSRF tokens) are recorded once in the run's
// observed-secrets file beside the JSON report, which the harness moves into its out-of-tree secret registry before
// scanning; saved evidence carries only counted SHA-256-prefix placeholders, never the values themselves.

const REPORT_DIR = dirname(process.env.PLAYWRIGHT_JSON_OUTPUT_FILE ?? "/tmp/playwright/results.json");
const OBSERVED = `${REPORT_DIR}/secrets-observed.txt`;
const recorded = new Set<string>();

// Records a bearer or protocol value for the secret scan. Empty values are ignored. The file is readable by the
// harness's (different) user; its protection is the report directory's 0700 parent, as for every report file.
export function observe(value: string | undefined | null): void {
  if (!value || recorded.has(value)) return;
  recorded.add(value);
  appendFileSync(OBSERVED, `${value}\n`, { mode: 0o644 });
}

export function placeholder(kind: string, value: string): string {
  return `[${kind}:sha256:${createHash("sha256").update(value).digest("hex").slice(0, 12)}]`;
}

export type Sanitized = { text: string; replaced: number };

// Replaces every CSRF hidden-field value with a placeholder and counts the replacements, so saved account HTML stays
// useful display evidence. The caller asserts the count equals the number of CSRF-bearing forms.
export function sanitizeHtml(html: string): Sanitized {
  let replaced = 0;
  const text = html.replace(/(<input[^>]*\bname="csrf_token"[^>]*\bvalue=")([^"]*)(")/g, (_m, a: string, v: string, b: string) => {
    replaced += 1;
    observe(v);
    return `${a}${placeholder("csrf", v)}${b}`;
  });
  return { text, replaced };
}

const SECRET_PARAMS = new Set(["code", "state", "code_challenge", "logout_ref", "session_id"]);

// Replaces the values of protocol and bearer query parameters with counted placeholders.
export function sanitizeUrl(url: string): Sanitized {
  let replaced = 0;
  let parsed: URL;
  try {
    parsed = new URL(url);
  } catch {
    return { text: "[unparsable-url]", replaced: 0 };
  }
  const params = new URLSearchParams();
  for (const [key, value] of parsed.searchParams) {
    if (SECRET_PARAMS.has(key) && value !== "") {
      observe(value);
      params.append(key, placeholder(key, value));
      replaced += 1;
    } else {
      params.append(key, value);
    }
  }
  const query = params.toString();
  return { text: `${parsed.origin}${parsed.pathname}${query ? `?${decodeURIComponent(query)}` : ""}`, replaced };
}
