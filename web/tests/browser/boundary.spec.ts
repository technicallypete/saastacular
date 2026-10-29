import { createHash } from "node:crypto";
import { writeFile } from "node:fs/promises";
import { runInNewContext } from "node:vm";

import { ThemeProvider } from "next-themes";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";

import { APPEARANCE_STORAGE_KEY, DEFAULT_APPEARANCE } from "@/lib/appearance";

import { expect, test } from "./support/page-errors";
import { chooseAppearance, waitForHydration } from "./support/modes";

const INTERNAL = /core:8180|intelligence:8000|postgres:5432|redis:6379|minio:9000|CORE_INTERNAL_URL|\b(DB|REDIS|OBJECT_STORE|MINIO|INTELLIGENCE)_[A-Z]|fonts\.googleapis|fonts\.gstatic/;

test("[S-BOUNDARY] the browser only talks to its own origin and receives no internal configuration", async ({ page, baseURL }, testInfo) => {
  const origin = new URL(baseURL!).origin;
  const requests: string[] = [];
  const bodies: string[] = [];
  const types = new Set<string>();
  page.on("request", (req) => requests.push(req.url()));
  page.on("response", async (res) => {
    const type = res.request().resourceType();
    if (["document", "script", "stylesheet", "fetch", "xhr"].includes(type) && res.status() < 400) {
      try {
        bodies.push(`### ${res.url()}\n${await res.text()}`);
        types.add(type);
      } catch {
        // Redirects and aborted responses have no body.
      }
    }
  });
  await page.goto("/");
  await waitForHydration(page);
  await chooseAppearance(page, "Dark");
  await chooseAppearance(page, "System");
  await page.getByRole("link", { name: "About" }).click();
  await page.waitForLoadState("networkidle");

  const foreign = requests.filter((u) => !u.startsWith("data:") && !u.startsWith("blob:") && new URL(u).origin !== origin);
  expect(foreign, "requests to other origins (internal services, fonts, CDNs)").toEqual([]);
  expect(requests.length).toBeGreaterThan(0);
  expect([...types].sort(), "document, script and stylesheet bodies were captured").toEqual(expect.arrayContaining(["document", "script", "stylesheet"]));
  const leaks = bodies.filter((b) => INTERNAL.test(b)).map((b) => b.split("\n")[0]);
  expect(leaks, "internal service names or configuration in delivered HTML/JS/CSS").toEqual([]);
  // Saved as a plain file in the output directory, so the harness secret scan reads the bodies.
  const saved = testInfo.outputPath("responses.txt");
  await writeFile(saved, bodies.join("\n\n"));
  await testInfo.attach("responses.txt", { path: saved, contentType: "text/plain" });
  await testInfo.attach("requests.txt", { body: requests.join("\n"), contentType: "text/plain" });
});

test("[S-HEALTH] liveness and readiness answer with no-store", async ({ request }) => {
  const health = await request.get("/api/health");
  expect(health.status()).toBe(200);
  expect(health.headers()["cache-control"]).toContain("no-store");
  expect(await health.json()).toEqual({ status: "alive" });
  const ready = await request.get("/api/ready");
  expect(ready.status()).toBe(200);
  expect(ready.headers()["cache-control"]).toContain("no-store");
  expect(await ready.json()).toEqual({ status: "ready", checks: { core: "ok" } });
});

test("[S-NOFIXTURE] the test-only fixture route does not exist in the production image", async ({ request }) => {
  expect((await request.get("/token-fixtures")).status()).toBe(404);
});

test("[S-FAILCLOSED] missing PRODUCT_NAME fails closed without leaking configuration", async ({ request }) => {
  const res = await request.get("http://127.0.0.1:3002/");
  expect(res.status()).toBe(500);
  const body = await res.text();
  expect(body).not.toContain("Fixture Product");
  expect(body).not.toMatch(/PRODUCT_NAME|PLATFORM_BASE_DOMAIN|CORE_INTERNAL_URL|fixture\.invalid/);
  expect((await request.get("http://127.0.0.1:3002/api/health")).status()).toBe(200);
});

// The inline script next-themes itself renders for the provider props G069 S5 fixes (attribute class, System by default
// and followed, the unchanged storage key, no color-scheme style), produced by the library, never copied. The served
// text is `(<function>)(<arguments>)`: the arguments are built from the props at render time; the function's source text
// is whatever the production bundler emitted, so it is compared by behaviour below and recorded, not required byte-equal.
function libraryThemeScript(): string {
  const html = renderToStaticMarkup(
    createElement(ThemeProvider, { attribute: "class", defaultTheme: DEFAULT_APPEARANCE, enableSystem: true, storageKey: APPEARANCE_STORAGE_KEY, enableColorScheme: false }, null),
  );
  const m = html.match(/^<script\b[^>]*>([\s\S]*?)<\/script>/);
  if (!m) throw new Error(`next-themes rendered no leading script: ${html.slice(0, 120)}`);
  return m[1];
}

const argumentsOf = (script: string) => script.slice(script.lastIndexOf(")(") + 1);

// Runs a served theme script against a minimal document: the classes it leaves on <html> for a stored value (null: none
// stored) and an OS appearance.
function runThemeScript(script: string, stored: string | null, os: "light" | "dark"): { classes: string[]; attributes: Record<string, string>; colorScheme: string } {
  const classes = new Set<string>();
  const attributes: Record<string, string> = {};
  const style = { colorScheme: "" };
  const el = {
    classList: { add: (...c: string[]) => c.forEach((x) => classes.add(x)), remove: (...c: string[]) => c.forEach((x) => classes.delete(x)) },
    setAttribute: (n: string, v: string) => (attributes[n] = v),
    style,
  };
  const matchMedia = (q: string) => ({ matches: q === "(prefers-color-scheme: dark)" && os === "dark" });
  const localStorage = { getItem: (k: string) => (k === APPEARANCE_STORAGE_KEY ? stored : null) };
  const sandbox: Record<string, unknown> = { document: { documentElement: el }, localStorage, matchMedia };
  sandbox.window = sandbox;
  runInNewContext(script, sandbox, { timeout: 1000 });
  return { classes: [...classes].sort(), attributes, colorScheme: style.colorScheme };
}

test("[S-HEAD-ORDER] the theme init is one classic inline script, first in <body> before any content, from next-themes", async ({ request }, testInfo) => {
  const html = await (await request.get("/")).text();
  const headEnd = html.indexOf("</head>");
  const bodyStart = html.search(/<body[\s>]/);
  expect(headEnd).toBeGreaterThan(0);
  expect(bodyStart).toBeGreaterThan(headEnd);
  // G069 S5: the retired appearance-init script is gone; exactly one inline script writes the stored theme.
  expect(html.includes('id="appearance-init"'), "no appearance-init script").toBe(false);
  const scripts = [...html.matchAll(/<script\b([^>]*)>([\s\S]*?)<\/script>/g)].filter((m) => m[2].includes(`"${APPEARANCE_STORAGE_KEY}"`));
  expect(scripts, "exactly one inline theme script").toHaveLength(1);
  const [, attrs, content] = scripts[0];
  expect(attrs, "classic and synchronous").not.toMatch(/\b(src|async|defer|type|nonce)\b/);
  const at = scripts[0].index!;
  // The provider is the body's first child, so its script precedes every element below it.
  expect(at, "inside <body>").toBeGreaterThan(bodyStart);
  for (const tag of ["<header", "<main", "<h1", "<nav"]) {
    const first = html.indexOf(tag, bodyStart);
    if (first >= 0) expect(at, `before the first ${tag}`).toBeLessThan(first);
  }
  const library = libraryThemeScript();
  expect(argumentsOf(content), "served with next-themes' arguments for the S5 props").toBe(argumentsOf(library));
  // Behaviour equals the library's own script: the stored preference (System: the OS) as the only appearance class,
  // no attribute and no color-scheme style.
  const behaviour: Record<string, unknown> = {};
  for (const stored of [null, "light", "dark", "system"]) {
    for (const os of ["light", "dark"] as const) {
      const served = runThemeScript(content, stored, os);
      const own = runThemeScript(library, stored, os);
      behaviour[`${stored ?? "none"}/${os}`] = served;
      expect(served, `stored ${stored}, OS ${os}: the served script behaves as next-themes' own`).toEqual(own);
      expect(served.classes, `stored ${stored}, OS ${os}: the resolved class`).toEqual([stored === "light" || stored === "dark" ? stored : os]);
      expect([served.attributes, served.colorScheme], `stored ${stored}, OS ${os}: nothing else written`).toEqual([{}, ""]);
    }
  }
  const sha256 = createHash("sha256").update(content).digest("hex");
  // The attachment name is kept so the harness's documented-hash check (docs/foundation/WEB-SHELL.md) keeps reading it.
  await testInfo.attach("appearance-init.sha256", { body: sha256, contentType: "text/plain" });
  await testInfo.attach("theme-script.json", { body: JSON.stringify({ identicalToLibrary: content === library, sha256, behaviour }, null, 2), contentType: "application/json" });
  // Served <html> carries exactly lang="en"; next-themes writes the class in the browser, and React serializes no
  // hydration-warning attribute.
  const htmlTag = html.match(/<html([^>]*)>/)![1].trim();
  expect(htmlTag).toBe('lang="en"');
});
