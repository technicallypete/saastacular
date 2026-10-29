import { MutationFields } from "@/components/branding/branding-form";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Item, ItemActions, ItemContent, ItemGroup, ItemTitle } from "@/components/ui/item";
import { Text } from "@/components/vestive/typography";
import type { BrandingView, ManagedLogo } from "@/lib/branding";

// G017 (Plan r2 §1.4, §5.4): the logo part of the branding management page — the current logo (shown through this host's
// own versioned logo route), the upload form (PNG or JPEG up to 512 KiB; core validates and re-encodes it), the remove
// form, and the uploads that are not the current logo, each with a guarded delete form. Upload state is core's; an
// object whose publication outcome was uncertain appears here as an upload until it is deleted.

const STATE: Readonly<Record<string, string>> = {
  pending: "Upload in progress or incomplete",
  ready: "Stored",
  deleting: "Being deleted",
  deleted: "Deleted",
  absent: "No longer stored",
};

function size(bytes: number): string {
  return bytes < 1024 ? `${bytes} bytes` : `${(bytes / 1024).toFixed(1)} KiB`;
}

function kind(logo: ManagedLogo): string {
  return logo.contentType === "image/png" ? "PNG" : logo.contentType === "image/jpeg" ? "JPEG" : "Unknown type";
}

function Current({ view, label }: { view: BrandingView; label: string }) {
  const logo = view.logo;
  if (logo === null) {
    return (
      <Text data-testid="branding-logo-none">No logo is set. The workspace name is shown instead.</Text>
    );
  }
  return (
    <div className="flex flex-col items-start gap-4" data-testid="branding-logo-current" data-state={logo.state}>
      {logo.version !== null && logo.state === "ready" ? (
        <span className="flex h-16 max-w-full min-w-0 items-center rounded-2xl border border-border bg-background p-2">
          {/* eslint-disable-next-line @next/next/no-img-element -- this host's own versioned logo route, sized by its box */}
          <img src={`/branding/logo?v=${logo.version}`} alt={`Current logo of ${label}`} className="h-full w-auto max-w-full min-w-0 object-contain" data-testid="branding-logo-preview" />
        </span>
      ) : null}
      <Text>
        {`${kind(logo)} · ${size(logo.byteSize)} · ${STATE[logo.state] ?? "Unknown state"}`}
      </Text>
    </div>
  );
}

export function LogoForms({ view, csrfToken, label }: { view: BrandingView; csrfToken: string; label: string }) {
  return (
    <div className="flex flex-col gap-12">
      <Current view={view} label={label} />
      <form
        method="post"
        action="/organization/branding/logo"
        encType="multipart/form-data"
        className="flex flex-col items-start gap-6"
        data-testid="branding-logo-upload"
      >
        <MutationFields csrfToken={csrfToken} version={view.version} />
        <Field>
          <FieldLabel htmlFor="logo-file">New logo</FieldLabel>
          {/* G069 S12: the kit Input (its file-button styles). */}
          <Input id="logo-file" name="logo" type="file" accept="image/png,image/jpeg" required aria-describedby="logo-help" />
          <FieldDescription id="logo-help">
            PNG or JPEG, at most 512 KiB, from 16 to 2048 pixels on each side. The image is checked and re-encoded before
            it is used; animated images, SVG and other formats are refused.
          </FieldDescription>
        </Field>
        <PendingSubmit label="Upload logo" pendingLabel="Uploading…" />
      </form>
      {view.logo !== null ? (
        <form method="post" action="/organization/branding/logo/remove" className="flex flex-col items-start gap-6" data-testid="branding-logo-remove">
          <MutationFields csrfToken={csrfToken} version={view.version} />
          <Text>Stop showing the current logo. The image stays listed below until you delete it.</Text>
          <PendingSubmit label="Remove logo" pendingLabel="Removing…" variant="outline" />
        </form>
      ) : null}
      <div className="flex flex-col gap-6">
        <Text className="font-medium">Other uploads</Text>
        {view.uploads.length === 0 ? (
          <Text data-testid="branding-uploads-empty">There are no other uploads.</Text>
        ) : (
          <ItemGroup aria-label="Other uploads">
            {view.uploads.map((u) => (
              <Item key={u.objectId} variant="outline" role="listitem" data-testid="branding-upload" data-state={u.state}>
                <ItemContent>
                  {/* Layout only: the whole state stays visible at 320px instead of the kit's one-line clamp. */}
                  <ItemTitle className="line-clamp-none">{`${kind(u)} · ${size(u.byteSize)} · ${STATE[u.state] ?? "Unknown state"}`}</ItemTitle>
                </ItemContent>
                <ItemActions>
                  <form method="post" action="/organization/branding/objects/delete">
                    <MutationFields csrfToken={csrfToken} version={view.version} />
                    <input type="hidden" name="object_id" value={u.objectId} />
                    <PendingSubmit label="Delete this upload" pendingLabel="Deleting…" variant="outline" />
                  </form>
                </ItemActions>
              </Item>
            ))}
          </ItemGroup>
        )}
        {view.uploadsTotal > view.uploads.length ? (
          <Text>{`Showing ${view.uploads.length} of ${view.uploadsTotal} uploads.`}</Text>
        ) : null}
      </div>
    </div>
  );
}
