import "server-only";

import { realpathSync, statSync } from "node:fs";
import { isAbsolute, join, relative, resolve, sep } from "node:path";

// OS002 (Plan r1 F4, decision D1 = B): the default product mark is a file-replacement convention, not configuration.
// The header shows the mark beside the configured product name only while web/public/brand/mark.png exists; an adopter
// who renames the product replaces or deletes that file (and web/app/icon.png, the favicon), with no variable and no
// code change (docs/foundation/WEB-SHELL.md "Product mark and favicon"). The URL is fixed and same-origin; nothing a
// request carries reaches this module.
export const PRODUCT_MARK_URL = "/brand/mark.png";

// The mark's file under a server root: the public/ directory that the server serves at "/".
const MARK_SEGMENTS = ["public", "brand", "mark.png"] as const;

// productMarkAt(root): PRODUCT_MARK_URL when public/brand/mark.png under root is a regular file whose real path stays
// inside root's real path, else null (missing root or file, a directory or other non-file, or a symbolic link that
// resolves outside root). Any filesystem error is "no mark", so the header falls back to text only.
export function productMarkAt(root: string): string | null {
  try {
    const base = realpathSync(resolve(root));
    const file = realpathSync(join(base, ...MARK_SEGMENTS));
    const rel = relative(base, file);
    if (rel === "" || isAbsolute(rel) || rel.split(sep)[0] === "..") return null;
    return statSync(file).isFile() ? PRODUCT_MARK_URL : null;
  } catch {
    return null;
  }
}

// The server's own root is its working directory: the standalone server (node server.js) changes to its own directory,
// where the image build copies public/ beside server.js (infra/docker/web.Dockerfile), and `next dev`/`next start` run
// from the web project root, which holds public/.
export function productMark(): string | null {
  return productMarkAt(process.cwd());
}
