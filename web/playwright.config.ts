import { readFileSync } from "node:fs";

import { defineConfig, devices, type PlaywrightTestConfig } from "@playwright/test";

// Browser validation (docs/foundation/WEB-SHELL.md). Runs in the pinned Playwright container,
// sharing the web service's network namespace (infra/compose/browser.yml):
//   shell-*      the real running web service at WEB_BASE_URL
//   tokens-*     the test-only fixture build, served from this container on 127.0.0.1:3001
//   appearance   Light/Dark/System behaviour and first paint against the real service
//   dev-hydration  React's development hydration checks against `next dev` on 127.0.0.1:3003
//   S-LONG-IDENTITY (shell-*) the fixture build with a 64-character product name on 127.0.0.1:3004
// With IDENTITY_E2E=1 (docs/foundation/IDENTITY.md "identity:browser"; G004) only the identity projects run,
// against the separate disposable HTTPS identity environment (infra/compose/identity.yml):
//   identity-*   https://org-a.localhost:543 and https://org-b.localhost:543 through the TLS test proxy and the
//                emulator; Chromium trusts exactly the run's leaf key (its SPKI pin from the run's volume), never
//                any certificate error in general. No fixture server starts, and no inherited project runs there.
//   mcp-*        (G014) the MCP connection setup page on the same environment and pin, after the identity projects;
//                the harness seeds the two registry Organizations first (infra/identity/mcp-seed.sql)
//   onboarding-*, membership-*, enrollment-*  (G018, G019, G058) on the same environment, pin and seed
//   signup-*     (G062) root-host tenant signup from https://localhost:543 on the same environment, pin and seed
//   labels-*     (G020) generic labels and their visibility on the tenant hosts, same environment, pin and seed
//   default-tenant-light  (G071) the default-tenant root alias on https://localhost:543, same environment, pin and seed;
//                run only by the coordinator's private G071 stack with DEFAULT_TENANT_PHASE naming the core
//                configuration under test, never by the documented identity:browser line
// With FEATURE_E2E=1 (the harness's feature-web suite; G051) only the feature projects run, over the registered
// feature areas' specs in tests/browser/features/<area>/, which no inherited project collects:
//   feature-*        every feature spec except *.stack.spec.ts, against the fixture build on 127.0.0.1:3001, the only
//                    fixture server started
//   feature-stack-*  with FEATURE_STACK=1 as well: the *.stack.spec.ts specs against the running stack at WEB_BASE_URL
//                    (no further server)
// PLAYWRIGHT_WORKERS (an integer from 1 to 4, default 1) sets the worker count; the documented commands keep 1 and the
// harness sets it for its own runs (G051 L6).
// Chromium only; no retries, so failures stay visible.
const WEB_BASE_URL = process.env.WEB_BASE_URL;
if (!WEB_BASE_URL) throw new Error("WEB_BASE_URL is required");
const IDENTITY_E2E = process.env.IDENTITY_E2E === "1";
const FEATURE_E2E = process.env.FEATURE_E2E === "1";
const FEATURE_STACK = process.env.FEATURE_STACK === "1";
if (FEATURE_E2E && IDENTITY_E2E) throw new Error("FEATURE_E2E and IDENTITY_E2E are exclusive");
if (FEATURE_STACK && !FEATURE_E2E) throw new Error("FEATURE_STACK requires FEATURE_E2E");

function workerCount(): number {
  const value = process.env.PLAYWRIGHT_WORKERS;
  if (value === undefined) return 1;
  if (!/^[1-4]$/.test(value)) throw new Error("PLAYWRIGHT_WORKERS must be an integer from 1 to 4");
  return Number(value);
}

const FIXTURE_ENV = {
  PRODUCT_NAME: "Fixture Product",
  PLATFORM_BASE_DOMAIN: "fixture.invalid",
  CUSTOM_DOMAIN_CNAME_TARGET: "custom.fixture.invalid",
  CORE_INTERNAL_URL: "http://127.0.0.1:9",
  APP_ENV: "test",
  NEXT_TELEMETRY_DISABLED: "1",
};

// A configured identity at 64 characters, with one long unbroken token (S-LONG-IDENTITY).
const LONG_PRODUCT_NAME = "Configured Identity Unbrokenproductnamewithoutanyspacesforwrap Q";

const chromium = { ...devices["Desktop Chrome"], viewport: { width: 1280, height: 800 }, deviceScaleFactor: 1 };

const SHELL_SPECS = /(shell|overlay|accessibility|boundary|design|account)\.spec\.ts/;
// Feature-area specs (G051): never collected by an inherited project.
const FEATURE_SPECS = /tests\/browser\/features\//;

const inheritedProjects: PlaywrightTestConfig["projects"] = [
  { name: "shell-light", testMatch: SHELL_SPECS, testIgnore: FEATURE_SPECS, use: { ...chromium, baseURL: WEB_BASE_URL, colorScheme: "light" } },
  { name: "shell-dark", testMatch: SHELL_SPECS, testIgnore: FEATURE_SPECS, use: { ...chromium, baseURL: WEB_BASE_URL, colorScheme: "dark" } },
  { name: "tokens-light", testMatch: /tokens\.spec\.ts/, testIgnore: FEATURE_SPECS, use: { ...chromium, baseURL: "http://127.0.0.1:3001", colorScheme: "light" } },
  { name: "tokens-dark", testMatch: /tokens\.spec\.ts/, testIgnore: FEATURE_SPECS, use: { ...chromium, baseURL: "http://127.0.0.1:3001", colorScheme: "dark" } },
  { name: "appearance", testMatch: /appearance\.spec\.ts/, testIgnore: FEATURE_SPECS, grepInvert: /@dev-hydration/, use: { ...chromium, baseURL: WEB_BASE_URL } },
  { name: "dev-hydration", testMatch: /appearance\.spec\.ts/, testIgnore: FEATURE_SPECS, grep: /@dev-hydration/, use: { ...chromium, baseURL: "http://127.0.0.1:3003" } },
];

// G051 feature projects: fixture-build specs, plus the stack specs with FEATURE_STACK=1.
const FEATURE_FIXTURE_SPECS = /tests\/browser\/features\/.+\.spec\.ts$/;
const FEATURE_STACK_SPECS = /tests\/browser\/features\/.+\.stack\.spec\.ts$/;
function featureProjects(): PlaywrightTestConfig["projects"] {
  const fixture = { ...chromium, baseURL: "http://127.0.0.1:3001" };
  const projects: NonNullable<PlaywrightTestConfig["projects"]> = [
    { name: "feature-light", testMatch: FEATURE_FIXTURE_SPECS, testIgnore: FEATURE_STACK_SPECS, use: { ...fixture, colorScheme: "light" } },
    { name: "feature-dark", testMatch: FEATURE_FIXTURE_SPECS, testIgnore: FEATURE_STACK_SPECS, use: { ...fixture, colorScheme: "dark" } },
  ];
  if (FEATURE_STACK) {
    projects.push(
      { name: "feature-stack-light", testMatch: FEATURE_STACK_SPECS, use: { ...chromium, baseURL: WEB_BASE_URL, colorScheme: "light" } },
      { name: "feature-stack-dark", testMatch: FEATURE_STACK_SPECS, use: { ...chromium, baseURL: WEB_BASE_URL, colorScheme: "dark" } },
    );
  }
  return projects;
}

// The run's leaf-key pin (identitysetup writes it into the identity volume, mounted at /identity).
function identityProjects(): PlaywrightTestConfig["projects"] {
  const pin = readFileSync("/identity/spki.txt", "utf8").trim();
  if (!/^[A-Za-z0-9+/]{43}=$/.test(pin)) throw new Error("/identity/spki.txt is not an SPKI pin");
  // G016: Chromium resolves localhost and *.localhost itself (G062: the platform root localhost and every newly created
  // <slug>.localhost tenant, with no DNS step); the overlay's custom hosts (*.acme.test) map explicitly to the same TLS
  // edge. G062 removed the marketing stand-in: the public entry serves its own landing (no redirect hop).
  const launchOptions = {
    args: [`--ignore-certificate-errors-spki-list=${pin}`, "--host-resolver-rules=MAP *.acme.test 127.0.0.1"],
  };
  const use = { ...chromium, baseURL: "https://org-a.localhost:543", launchOptions };
  // G016 E-DOM-1..5 (identity.spec.ts @domain-*): the third platform host org-c and the custom host portal.acme.test.
  const domains = /@domain-/;
  return [
    { name: "identity-light", testMatch: /identity\.spec\.ts/, grepInvert: domains, use: { ...use, colorScheme: "light" } },
    { name: "identity-dark", testMatch: /identity\.spec\.ts/, grepInvert: domains, use: { ...use, colorScheme: "dark" } },
    { name: "domain-platform", testMatch: /identity\.spec\.ts/, grep: /@domain-platform/, use: { ...use, baseURL: "https://org-c.localhost:543", colorScheme: "light" } },
    { name: "domain-custom", testMatch: /identity\.spec\.ts/, grep: /@domain-custom/, use: { ...use, baseURL: "https://portal.acme.test:543", colorScheme: "dark" } },
    { name: "mcp-light", testMatch: /mcp\.spec\.ts/, use: { ...use, colorScheme: "light" } },
    { name: "mcp-dark", testMatch: /mcp\.spec\.ts/, use: { ...use, colorScheme: "dark" } },
    // G018 (Plan r5 §6.2 E-ONB-1..7): Group creation and Business Profile onboarding on the tenant hosts.
    { name: "onboarding-light", testMatch: /onboarding\.spec\.ts/, use: { ...use, colorScheme: "light" } },
    { name: "onboarding-dark", testMatch: /onboarding\.spec\.ts/, use: { ...use, colorScheme: "dark" } },
    // G019 (Plan r1 §6.3 E-MEM-1..7): invitations, membership and ownership on the tenant hosts.
    { name: "membership-light", testMatch: /membership\.spec\.ts/, use: { ...use, colorScheme: "light" } },
    { name: "membership-dark", testMatch: /membership\.spec\.ts/, use: { ...use, colorScheme: "dark" } },
    // G058 (Plan r3 §6.2 E-ENR-1..5): joining a workspace (open signup, Organization invitation entry, access requests)
    // and the enrollment policy and request pages on the tenant hosts.
    { name: "enrollment-light", testMatch: /enrollment\.spec\.ts/, use: { ...use, colorScheme: "light" } },
    { name: "enrollment-dark", testMatch: /enrollment\.spec\.ts/, use: { ...use, colorScheme: "dark" } },
    // G062 (Plan r5 SUP-15 E-SUP-1..6): root-host tenant signup from the platform root https://localhost:543 (and
    // www.localhost:543) into the newly created <slug>.localhost:543 tenant.
    { name: "signup-light", testMatch: /signup\.spec\.ts/, use: { ...use, baseURL: "https://localhost:543", colorScheme: "light" } },
    { name: "signup-dark", testMatch: /signup\.spec\.ts/, use: { ...use, baseURL: "https://localhost:543", colorScheme: "dark" } },
    // G020 (Plan r2 §5.2 T12 E-LAB-1..6, D-10): generic labels and their visibility (Organization labels, a Group page's
    // Labels section, a Group's own labels) on the tenant hosts.
    { name: "labels-light", testMatch: /labels\.spec\.ts/, use: { ...use, colorScheme: "light" } },
    { name: "labels-dark", testMatch: /labels\.spec\.ts/, use: { ...use, colorScheme: "dark" } },
    // G071 (Plan r4 §5.1 B1 E-DT-1..6): the platform root https://localhost:543 aliased to a configured default tenant
    // (and its root presentation modes); DEFAULT_TENANT_PHASE selects the scenarios of the core configuration under
    // test. Not in the identity:browser project list, so the shared harness never runs it.
    { name: "default-tenant-light", testMatch: /default-tenant\.spec\.ts/, use: { ...use, baseURL: "https://localhost:543", colorScheme: "light" } },
  ];
}

// Fixture build (same component source, plus the test-only /token-fixtures route); the feature projects' only server.
const fixtureServer3001 = {
  command: "node /fixture/server.js",
  url: "http://127.0.0.1:3001/api/health",
  env: { ...FIXTURE_ENV, PORT: "3001", HOSTNAME: "127.0.0.1" },
  reuseExistingServer: false,
  timeout: 60_000,
};

const fixtureServers: PlaywrightTestConfig["webServer"] = [
  fixtureServer3001,
  {
    // Same build with PRODUCT_NAME empty: the page must fail closed (S-FAILCLOSED).
    command: "node /fixture/server.js",
    url: "http://127.0.0.1:3002/api/health",
    env: { ...FIXTURE_ENV, PRODUCT_NAME: "", PORT: "3002", HOSTNAME: "127.0.0.1" },
    reuseExistingServer: false,
    timeout: 60_000,
  },
  {
    // Development server on the unmodified source: React reports any hydration mismatch here.
    command: "npx next dev --hostname 127.0.0.1 --port 3003",
    url: "http://127.0.0.1:3003/api/health",
    env: { ...FIXTURE_ENV, NODE_ENV: "development" },
    reuseExistingServer: false,
    timeout: 180_000,
  },
  {
    // Same build with a 64-character configured product name: reflow, wrapping and contrast (S-LONG-IDENTITY).
    command: "node /fixture/server.js",
    url: "http://127.0.0.1:3004/api/health",
    env: { ...FIXTURE_ENV, PRODUCT_NAME: LONG_PRODUCT_NAME, PORT: "3004", HOSTNAME: "127.0.0.1" },
    reuseExistingServer: false,
    timeout: 60_000,
  },
];

export default defineConfig({
  testDir: "./tests/browser",
  outputDir: process.env.PLAYWRIGHT_OUTPUT_DIR ?? "/tmp/playwright/out",
  forbidOnly: true,
  retries: 0,
  workers: workerCount(),
  timeout: 120_000,
  expect: { timeout: 10_000 },
  reporter: [["list"], ["json", { outputFile: process.env.PLAYWRIGHT_JSON_OUTPUT_FILE ?? "/tmp/playwright/results.json" }]],
  use: { trace: "off", screenshot: "only-on-failure", video: "off" },
  projects: IDENTITY_E2E ? identityProjects() : FEATURE_E2E ? featureProjects() : inheritedProjects,
  webServer: IDENTITY_E2E ? [] : FEATURE_E2E ? [fixtureServer3001] : fixtureServers,
});
