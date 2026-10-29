import { expect, test as base } from "@playwright/test";

// S-ERRORS: every test fails if the page reports an uncaught error or a console error, which
// includes React hydration mismatches. Nothing is filtered out.
export const test = base.extend<{ pageErrors: string[] }>({
  pageErrors: [
    async ({ page }, use, testInfo) => {
      const errors: string[] = [];
      page.on("pageerror", (error) => errors.push(`pageerror: ${error.message}`));
      page.on("console", (message) => {
        if (message.type() === "error") errors.push(`console.error: ${message.text()}`);
      });
      await use(errors);
      expect(errors, "uncaught page errors or console errors").toEqual([]);
      // Evidence for the harness's scenario matrix: this test ran under S-ERRORS and passed it.
      testInfo.annotations.push({ type: "S-ERRORS", description: "no pageerror or console.error" });
    },
    { auto: true },
  ],
});

export { expect };
