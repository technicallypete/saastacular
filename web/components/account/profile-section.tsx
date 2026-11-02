import { RULED_GRID_REFLOW } from "@/components/account/ruled-grid-reflow";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { RuledGrid, Section } from "@/components/vestive/layout";
import { Heading, Text } from "@/components/vestive/typography";
import type { OwnProfile, ProfileResult } from "@/lib/profile-client";

// G074 (Plan r3.1 §3.10): the Profile section of My Account — the signed-in user's own global presentation name and
// avatar, with three native POST forms (name, avatar upload, avatar removal) carrying this session's CSRF token and the
// profile's optimistic version. Server-rendered from the page's one own-profile read (or the fixture's scripted value);
// no cookie access and no client component of its own beyond the shared pending-submit control. Outcome notices come
// from a closed table keyed by the page's profile_status parameter; nothing read from the request is reflected. Also
// exports the avatar thumbnail used by the members tables.

type Param = string | string[] | undefined;

// The closed outcomes of the profile form routes (lib/profile.ts PROFILE_STATUSES).
const STATUS: Readonly<Record<string, string>> = {
  profile_saved: "Your name was saved. People you work with see it on newly loaded pages.",
  avatar_set: "Your new avatar was checked, stored and is now shown.",
  avatar_removed: "Your avatar was removed. Your initials are shown instead.",
  stale_version:
    "Your profile changed since this page was loaded, for example in another tab. Review the current values below and try again. Nothing was changed.",
  invalid_name: "That name was not saved. Use up to 200 characters without line breaks or control characters. Nothing was changed.",
  invalid_image:
    "That file is not an accepted avatar: use a PNG or JPEG from 16 to 2048 pixels on each side. Animated images, SVG and other formats are refused. Nothing was changed.",
  payload_too_large: "That file is larger than 512 KiB. Nothing was changed.",
  storage_unavailable: "The change could not be confirmed right now. Your current profile is shown below; check it before trying again.",
  avatar_hold:
    "Your avatar could not be replaced yet because an earlier upload or removal is still being cleaned up. Try again in a moment.",
};

// The initials shown when there is no avatar to show: the first letter of the first and last words of the name, or "?"
// when there is no name. Code points, never UTF-16 halves.
export function initials(name: string | null): string {
  const words = (name ?? "").trim().split(/\s+/).filter((w) => w !== "");
  if (words.length === 0) return "?";
  const first = Array.from(words[0])[0] ?? "";
  const last = words.length > 1 ? (Array.from(words[words.length - 1])[0] ?? "") : "";
  return `${first}${last}`.toLocaleUpperCase();
}

// The versioned, session-authorized avatar route of one user (app/profiles/[id]/avatar).
export function avatarSrc(userId: string, version: number): string {
  return `/profiles/${userId}/avatar?v=${version}`;
}

// A members-table thumbnail: the avatar only when core reported a version for this viewer (the relationship predicate
// admitted it and the subject has an avatar), otherwise initials. Decorative: the adjacent name is the accessible
// identity, so the image has empty alternative text and the whole mark is hidden from assistive technology.
export function MemberAvatar({ userId, name, version }: { userId: string; name: string | null; version: number | null }) {
  return (
    <Avatar aria-hidden="true" data-testid="member-avatar">
      {version !== null ? <AvatarImage src={avatarSrc(userId, version)} alt="" loading="lazy" data-testid="member-avatar-image" /> : null}
      <AvatarFallback>{initials(name)}</AvatarFallback>
    </Avatar>
  );
}

function Notice({ status }: { status: Param }) {
  if (typeof status !== "string" || !Object.hasOwn(STATUS, status)) return null;
  // A stale version is an alert (the kit Alert's own role); every other outcome is a polite status notice.
  if (status === "stale_version") {
    return (
      <Alert data-testid="profile-alert" className="mt-10">
        <AlertDescription>{STATUS[status]}</AlertDescription>
      </Alert>
    );
  }
  return (
    <Alert role="status" data-testid="profile-status" className="mt-10">
      <AlertDescription>{STATUS[status]}</AlertDescription>
    </Alert>
  );
}

function Hidden({ csrfToken, version }: { csrfToken: string; version: number }) {
  return (
    <>
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <input type="hidden" name="expected_version" value={version} />
    </>
  );
}

export type ProfileSectionProps = {
  // The page's own-profile read (GET /internal/v1/profiles/me) or the fixture's scripted answer.
  result: ProfileResult<OwnProfile>;
  csrfToken: string;
  // The raw profile_status parameter (validated here, never reflected).
  status: Param;
};

export function ProfileSection({ result, csrfToken, status }: ProfileSectionProps) {
  return (
    <Section id="profile" aria-labelledby="profile-title" tone="muted">
      <Heading level={2} id="profile-title">
        Your profile
      </Heading>
      <Text className="mt-10">
        Your name and avatar are the same in every workspace you belong to. In each workspace they are shown to its owners
        and admins and to members of Groups you share there. Your email address and sign-in details are never shown.
      </Text>
      {!result.ok ? (
        <Alert role="status" data-testid="profile-unavailable" className="mt-10">
          <AlertDescription>Your profile is unavailable right now. Your account and sign-out still work. Please try again later.</AlertDescription>
        </Alert>
      ) : (
        <ProfileForms profile={result.value} csrfToken={csrfToken} status={status} />
      )}
    </Section>
  );
}

function ProfileForms({ profile, csrfToken, status }: { profile: OwnProfile; csrfToken: string; status: Param }) {
  const shown = profile.effectiveName;
  return (
    <>
      <Notice status={status} />
      <div className="mt-12 flex min-w-0 flex-wrap items-center gap-6" data-testid="profile-current">
        <Avatar className="size-24" data-testid="profile-avatar">
          {profile.avatar !== null ? (
            <AvatarImage src={avatarSrc(profile.userId, profile.avatar.version)} alt="Your current avatar" data-testid="profile-avatar-image" />
          ) : null}
          <AvatarFallback aria-hidden="true" className="text-2xl" data-testid="profile-avatar-fallback">
            {initials(shown)}
          </AvatarFallback>
        </Avatar>
        <dl className="m-0 grid min-w-0 gap-4 font-sans text-sm text-foreground/80">
          <div>
            <dt className="text-muted-foreground">Shown as</dt>
            <dd data-testid="profile-effective-name" className="m-0 [overflow-wrap:anywhere]">
              {shown ?? "Name not provided"}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Name source</dt>
            <dd data-testid="profile-name-source" className="m-0">
              {profile.name !== null ? "The name you chose" : "Your sign-in provider"}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Avatar</dt>
            <dd data-testid="profile-avatar-state" className="m-0">
              {profile.avatar !== null ? "Set" : "Not set; your initials are shown"}
            </dd>
          </div>
        </dl>
      </div>
      <RuledGrid className="mt-12" style={RULED_GRID_REFLOW}>
        <form method="post" action="/account/profile/update" className="flex min-w-0 flex-col items-start gap-6" data-testid="profile-name-form">
          <Hidden csrfToken={csrfToken} version={profile.version} />
          <Field className="min-w-0">
            <FieldLabel htmlFor="profile-name">Name</FieldLabel>
            <Input
              id="profile-name"
              name="name"
              type="text"
              defaultValue={profile.name ?? ""}
              maxLength={400} // UTF-16 units: core's authoritative limit is 200 code points (up to 2 units each)
              autoComplete="off"
              spellCheck={false}
              aria-describedby="profile-name-help"
            />
            <FieldDescription id="profile-name-help">
              Shown to people you work with; leave empty to use the name from your sign-in provider
            </FieldDescription>
          </Field>
          <PendingSubmit label="Save name" pendingLabel="Saving…" />
        </form>
        <form
          method="post"
          action="/account/profile/avatar"
          encType="multipart/form-data"
          className="flex min-w-0 flex-col items-start gap-6"
          data-testid="profile-avatar-form"
        >
          <Hidden csrfToken={csrfToken} version={profile.version} />
          <Field className="min-w-0">
            <FieldLabel htmlFor="profile-avatar-file">New avatar (PNG or JPEG, at most 512 KiB)</FieldLabel>
            {/* The kit Input (its file-button styles), as the G017 logo form. */}
            <Input id="profile-avatar-file" name="avatar" type="file" accept="image/png,image/jpeg" required aria-describedby="profile-avatar-help" />
            <FieldDescription id="profile-avatar-help">
              From 16 to 2048 pixels on each side. The image is checked and re-encoded before it is shown; animated images,
              SVG and other formats are refused.
            </FieldDescription>
          </Field>
          <PendingSubmit label="Upload avatar" pendingLabel="Uploading…" />
        </form>
        {profile.avatar !== null ? (
          <form method="post" action="/account/profile/avatar/remove" className="flex min-w-0 flex-col items-start gap-6" data-testid="profile-avatar-remove-form">
            <Hidden csrfToken={csrfToken} version={profile.version} />
            <Text>Stop showing your avatar. Your initials are shown instead.</Text>
            <PendingSubmit label="Remove avatar" pendingLabel="Removing…" variant="outline" />
          </form>
        ) : null}
      </RuledGrid>
    </>
  );
}
