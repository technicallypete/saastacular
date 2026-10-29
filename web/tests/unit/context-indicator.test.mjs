// Unit tests for the current-Organization indicator (G017 Plan r2 §0.4 C-1, §6.2; AC7): indicatorState in
// lib/branding.ts is the closed state that components/layout/context-indicator.tsx renders verbatim (the component adds
// only the Sign in / Account links). On a stubbed SessionView: anonymous, member, signed-in non-member (a revoked or
// never-granted membership), the label or product-name fallback, and every failure other than unauthenticated showing
// no identity.
import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import { describe, it } from "node:test";

registerHooks({
  resolve(specifier, context, nextResolve) {
    if (/^\.\/[a-z-]+$/.test(specifier) && /\/lib\/[a-z-]+\.ts$/.test(context.parentURL ?? "")) {
      return nextResolve(`${specifier}.ts`, context);
    }
    return nextResolve(specifier, context);
  },
});

const { indicatorState } = await import("../../lib/branding.ts");

const view = (membership, displayName = "Ada Lovelace") => ({ ok: true, value: { displayName, membership } });

describe("indicatorState", () => {
  it("is anonymous without a session cookie or with an unauthenticated answer", () => {
    assert.deepEqual(indicatorState(undefined, "Acme"), { kind: "anonymous" });
    assert.deepEqual(indicatorState({ ok: false, error: "unauthenticated" }, "Acme"), { kind: "anonymous" });
  });
  it("names a member and the workspace", () => {
    assert.deepEqual(indicatorState(view("member"), "Acme <Tools>"), { kind: "member", text: "Signed in as Ada Lovelace · member of Acme <Tools>" });
    assert.deepEqual(indicatorState(view("member", null), "Test Product"), { kind: "member", text: "Signed in as your account · member of Test Product" });
  });
  it("says not a member for a signed-in user without a current membership (revoked or never granted)", () => {
    assert.deepEqual(indicatorState(view("none"), "Acme"), { kind: "none", text: "Signed in as Ada Lovelace · not a member of this workspace" });
  });
  it("shows no identity for any other failure", () => {
    for (const error of ["identity_unavailable", "timeout", "unreachable", "malformed", "unknown_host"]) {
      assert.deepEqual(indicatorState({ ok: false, error }, "Acme"), { kind: "unavailable" }, error);
    }
  });
});
