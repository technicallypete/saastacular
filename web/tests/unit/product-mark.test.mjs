// Unit tests for lib/product-mark.ts (OS002 Plan r1 F4, decision D1 = B): the default product mark is a file
// convention. productMarkAt(root) answers the fixed same-origin URL only while public/brand/mark.png under root is a
// regular file inside root, and null otherwise: a missing root or file, a directory at that path, or a symbolic link
// that resolves outside root (no path escape). Each case builds its own temporary directory, removed afterwards; the
// module imports only node:fs and node:path (server-only resolves to its empty react-server entry under
// `npm run test:unit`), so no resolve hook is needed. Run by `npm run test:unit` in the web check image with networking
// disabled.
import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { after, describe, it } from "node:test";

const { PRODUCT_MARK_URL, productMarkAt } = await import("../../lib/product-mark.ts");

const scratch = mkdtempSync(join(tmpdir(), "product-mark-"));
after(() => rmSync(scratch, { recursive: true, force: true }));

let n = 0;
// A fresh server root under the scratch directory; with `mark`, public/brand/ is created first.
function root({ mark } = {}) {
  const dir = join(scratch, `root-${(n += 1)}`);
  mkdirSync(dir);
  if (mark !== undefined) {
    mkdirSync(join(dir, "public", "brand"), { recursive: true });
    mark(join(dir, "public", "brand", "mark.png"));
  }
  return dir;
}

const PNG = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);

describe("productMarkAt", () => {
  it("answers the fixed same-origin URL", () => {
    assert.equal(PRODUCT_MARK_URL, "/brand/mark.png");
  });

  it("answers the URL when public/brand/mark.png is a file under the root", () => {
    assert.equal(productMarkAt(root({ mark: (p) => writeFileSync(p, PNG) })), "/brand/mark.png");
  });

  it("follows a symbolic link that stays inside the root", () => {
    const dir = root({ mark: () => {} });
    writeFileSync(join(dir, "public", "shared.png"), PNG);
    symlinkSync(join(dir, "public", "shared.png"), join(dir, "public", "brand", "mark.png"));
    assert.equal(productMarkAt(dir), "/brand/mark.png");
  });

  it("answers null when the file is absent", () => {
    assert.equal(productMarkAt(root()), null);
    assert.equal(productMarkAt(root({ mark: () => {} })), null);
  });

  it("answers null when the root does not exist", () => {
    assert.equal(productMarkAt(join(scratch, "no-such-root")), null);
  });

  it("answers null when the path is a directory, not a file", () => {
    assert.equal(productMarkAt(root({ mark: (p) => mkdirSync(p) })), null);
  });

  it("answers null for a symbolic link that escapes the root", () => {
    const outside = join(scratch, `outside-${(n += 1)}.png`);
    writeFileSync(outside, PNG);
    assert.equal(productMarkAt(root({ mark: (p) => symlinkSync(outside, p) })), null);
  });

  it("answers null for a dangling symbolic link", () => {
    assert.equal(productMarkAt(root({ mark: (p) => symlinkSync(join(scratch, "missing.png"), p) })), null);
  });

  it("reads only the fixed path under the given root, never a parent directory's mark", () => {
    const parent = root();
    mkdirSync(join(parent, "public", "brand"), { recursive: true });
    writeFileSync(join(parent, "public", "brand", "mark.png"), PNG);
    const child = join(parent, "child");
    mkdirSync(child);
    assert.equal(productMarkAt(child), null);
    assert.equal(productMarkAt(join(child, "..")), "/brand/mark.png");
  });
});
