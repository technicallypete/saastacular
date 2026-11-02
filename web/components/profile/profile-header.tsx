import { avatarSrc, initials } from "@/components/account/profile-section";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Section } from "@/components/vestive/layout";
import { Heading, Kicker } from "@/components/vestive/typography";
import type { VisibleProfile } from "@/lib/profile-client";

// G075 (Plan r2 §3.7, C9): the header of the composite profile page — G074's global presentation name and avatar of the
// subject exactly as core's relationship predicate shows them to this viewer now (GET /internal/v1/profiles/{user_id}).
// The avatar is the kit Avatar at 96 px through G074's session-authorized avatar route, initials otherwise; it is
// decorative (the heading carries the name), so it is hidden from assistive technology. Never an email, a sign-in
// identity, an Organization role or a Business Profile.

export function ProfileHeader({ profile }: { profile: VisibleProfile }) {
  return (
    <Section id="profile" aria-labelledby="profile-title" data-testid="member-profile-header">
      <Kicker>Profile</Kicker>
      <div className="mt-6 flex min-w-0 flex-wrap items-center gap-6">
        <Avatar aria-hidden="true" className="size-24" data-testid="member-profile-avatar">
          {profile.avatar !== null ? (
            <AvatarImage src={avatarSrc(profile.userId, profile.avatar.version)} alt="" data-testid="member-profile-avatar-image" />
          ) : null}
          <AvatarFallback className="text-2xl" data-testid="member-profile-avatar-fallback">
            {initials(profile.name)}
          </AvatarFallback>
        </Avatar>
        <Heading level={1} id="profile-title" className="min-w-0 [overflow-wrap:anywhere]" data-testid="member-profile-name">
          {profile.name ?? "Name not provided"}
        </Heading>
      </div>
    </Section>
  );
}
