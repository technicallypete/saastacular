// G017 (Plan r0 §1.5–1.6): the tenant header's identity. The Organization's current logo, served by this host's own
// /branding/logo route at its current version, with the label (or the configured product name) as its text
// alternative, so a broken image still names the workspace; otherwise the label or product name as text. The box is
// bounded (max-h-10, w-auto) so the mark never pushes the header past a 320 px viewport.
// OS002 (Plan r1 S6, F4): mark is the default product mark's URL (lib/product-mark.ts) when the text is the configured
// product name and the mark file exists, else null; the shell decides (components/layout/app-shell.tsx).
export function BrandMark({ name, logoVersion, mark = null }: { name: string; logoVersion: string | null; mark?: string | null }) {
  if (logoVersion !== null) {
    return (
      // A plain img: the bytes are this host's own versioned, already-normalized logo; the Next image optimizer would add
      // a second fetch path and cache outside the reviewed contract (Plan r2 §5.5).
      // eslint-disable-next-line @next/next/no-img-element
      <img
        src={`/branding/logo?v=${logoVersion}`}
        alt={name}
        height={40}
        data-testid="brand-logo"
        className="block h-10 max-h-10 w-auto max-w-full object-contain"
      />
    );
  }
  return (
    <p data-testid="brand-label" className="max-w-full min-w-0 font-heading text-lg font-semibold tracking-[1.2px] uppercase [overflow-wrap:anywhere]">
      {mark === null ? null : <ProductMarkImage src={mark} />}
      {name}
    </p>
  );
}

// OS002 (Plan r1 S6, F4): the default product mark, decorative (alt="": the product name beside it is the text), inside
// the identity paragraph before the name, so the header's identity is still exactly one paragraph whose text is the
// name. One line box high (28px, the text-lg line height) and top-aligned in it, so the header keeps its height and
// 320 px reflow. A plain img: a fixed same-origin static file, no optimizer route.
export function ProductMarkImage({ src }: { src: string }) {
  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img src={src} alt="" width={28} height={28} data-testid="product-mark" className="mr-3 inline-block size-7 align-top" />
  );
}
