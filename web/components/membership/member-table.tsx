import { MemberAvatar } from "@/components/account/profile-section";
import { FormSelect, type FormSelectOption } from "@/components/controls/form-select";
import { PendingSubmit } from "@/components/controls/pending-submit";
import { ScrollRegion } from "@/components/controls/scroll-region";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Text, TextLink } from "@/components/vestive/typography";
import { ROLE_LABELS, ROLES, type Member, type ScopeKind } from "@/lib/membership-client";

// G019 (Plan r1 §1.6): the members of the current scope (display names and roles of this scope only) with the per-row
// controls core's snapshot hints allow. The hints decide only which forms are shown; every submission is re-decided by
// core. Never imports server configuration or the request context (P-15).
// G069 S12: a kit Table in a labelled scroll region; its body cells wrap (the kit's cells keep one line). G069 r3
// (owner finding): every role dropdown is the kit Select through FormSelect.
// G074 (Plan r3.1 §3.9, §3.10): each name leads with the member's avatar thumbnail, an image only when core reported an
// avatar version for this viewer (its RLS-gated profile join admitted the subject), initials otherwise. The thumbnail
// sits inside the Name cell, so the table keeps its three columns.
// G075 (Plan r2 §3.7, C9, C10; Q7): each name links to the member's composite profile page /profiles/<id>. On a Group's
// members page (contextGroup: that Group's id, validated by the page) the link carries ?context=<Group id>, so the
// profile opens in that Group's selected context; on the Organization members page it carries no selector
// (Organization-only). The link adds no data: the target page re-decides every read. No label column (Q3).

type Props = { kind: ScopeKind; members: Member[]; action: `/${string}`; csrfToken: string; contextGroup?: string };

// The composite profile page of a member, in the table's selected-Group context when it has one.
export function memberProfileHref(userId: string, contextGroup: string | undefined): string {
  return contextGroup === undefined ? `/profiles/${userId}` : `/profiles/${userId}?context=${encodeURIComponent(contextGroup)}`;
}

export function memberName(m: { displayName: string | null }): string {
  return m.displayName ?? "Name not provided";
}

// The scope's two ordinary roles as dropdown options, admin first (the order the native options had).
export function roleOptions(kind: ScopeKind): FormSelectOption[] {
  return [ROLES[kind].admin, ROLES[kind].member].map((role) => ({ value: role, label: ROLE_LABELS[role] }));
}

function Hidden({ csrfToken, action, userId }: { csrfToken: string; action: string; userId: string }) {
  return (
    <>
      <input type="hidden" name="csrf_token" value={csrfToken} />
      <input type="hidden" name="action" value={action} />
      <input type="hidden" name="user_id" value={userId} />
    </>
  );
}

function RoleChoice({ kind, label, id }: { kind: ScopeKind; label: string; id: string }) {
  return (
    <div className="grid min-w-0 gap-2">
      <label htmlFor={id} className="font-sans text-sm font-medium text-foreground">
        {label}
      </label>
      <FormSelect id={id} name="role" defaultValue={ROLES[kind].member} required options={roleOptions(kind)} />
    </div>
  );
}

export function MemberTable({ kind, members, action, csrfToken, contextGroup }: Props) {
  if (members.length === 0) return <Text className="mt-8">No members to show.</Text>;
  return (
    <ScrollRegion label="Members" className="mt-8">
      <Table data-testid="member-table">
        <TableHeader>
          <TableRow>
            <TableHead scope="col">Name</TableHead>
            <TableHead scope="col">Role</TableHead>
            <TableHead scope="col">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {members.map((m) => {
            const name = memberName(m);
            const none = !m.canChangeRole && !m.canRemove && !m.canPromote && !m.canDemote;
            return (
              <TableRow key={m.userId} data-testid="member-row">
                <TableCell className="whitespace-normal [overflow-wrap:anywhere]">
                  <div className="flex min-w-0 items-center gap-3">
                    <MemberAvatar userId={m.userId} name={m.displayName} version={m.avatarVersion} />
                    <div className="min-w-0">
                      <TextLink
                        href={memberProfileHref(m.userId, contextGroup)}
                        className="max-w-full border-b pb-0 font-sans text-sm normal-case tracking-normal [overflow-wrap:anywhere]"
                        data-testid="member-profile-link"
                      >
                        {name}
                      </TextLink>
                      {m.self ? <span className="text-muted-foreground"> (you)</span> : null}
                      {m.disabled ? <span className="block text-muted-foreground">Account disabled</span> : null}
                    </div>
                  </div>
                </TableCell>
                <TableCell className="whitespace-normal">{ROLE_LABELS[m.role]}</TableCell>
                <TableCell className="whitespace-normal">
                  <div className="grid gap-6">
                    {m.canChangeRole ? (
                      <form method="post" action={action} className="grid gap-3">
                        <Hidden csrfToken={csrfToken} action="change_role" userId={m.userId} />
                        <RoleChoice kind={kind} id={`role-${m.userId}`} label={`New role for ${name}`} />
                        <PendingSubmit label="Change role" pendingLabel="Changing…" size="sm" variant="outline" />
                      </form>
                    ) : null}
                    {m.canDemote ? (
                      <form method="post" action={action} className="grid gap-3">
                        <Hidden csrfToken={csrfToken} action="demote" userId={m.userId} />
                        <RoleChoice kind={kind} id={`demote-${m.userId}`} label={`Role for ${name} after stepping down as owner`} />
                        <PendingSubmit label="Remove owner role" pendingLabel="Changing…" size="sm" variant="outline" />
                      </form>
                    ) : null}
                    {m.canPromote ? (
                      <form method="post" action={action}>
                        <Hidden csrfToken={csrfToken} action="promote" userId={m.userId} />
                        <PendingSubmit label="Make owner" pendingLabel="Changing…" size="sm" variant="outline" aria-label={`Make ${name} an owner`} />
                      </form>
                    ) : null}
                    {m.canRemove ? (
                      <form method="post" action={action}>
                        <Hidden csrfToken={csrfToken} action="remove" userId={m.userId} />
                        <PendingSubmit label="Remove" pendingLabel="Removing…" size="sm" variant="outline" aria-label={`Remove ${name}`} />
                      </form>
                    ) : null}
                    {none ? <span className="text-muted-foreground">None</span> : null}
                  </div>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </ScrollRegion>
  );
}
