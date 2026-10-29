// G017 (Plan r0 §1.5–1.6): the tenant header's identity. The Organization's current logo, served by this host's own
// /branding/logo route at its current version, with the label (or the configured product name) as its text
// alternative, so a broken image still names the workspace; otherwise the label or product name as text. The box is
// bounded (max-h-10, w-auto) so the mark never pushes the header past a 320 px viewport.
export function BrandMark({ name, logoVersion }: { name: string; logoVersion: string | null }) {
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
      {name}
    </p>
  );
}
