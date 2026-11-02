package access

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/audit"
	"core/internal/tenancy"
)

// G075 guarded Group member profile operations (Plan r2 §3.4, §3.6, §3.9; C4, C6, C8). A Group member profile is one
// optional descriptive role label for exactly one current Group membership; it has no access-control effect. Every
// operation is one guarded call in the session's context: liveness, the facts and the decision, then the statements of
// core/internal/tenancy (group_profiles.go) under row-level security. The cards a viewer sees are decided by the
// database's projection (app.rls_group_member_cards) on every call, behind G074's relationship predicate: an unknown,
// foreign, unrelated or no-longer-related subject is the same not_found as GetAccountProfile. A label is written by the
// member itself (group_member_profile.manage_self: any own role in that Group) or by the exact Group's own owner or
// admin (group_member_profile.manage); an Organization role and the G060 override never write one. The triggers own the
// binding check, the version rules and the single group_member_profile.updated audit row of a changed label; a write
// that changes nothing writes nothing.

// The Group member profile operations' logged names (and access.denied operations, 00024).
const (
	opProfileGroupsList       = "profile.groups.list"
	opProfileGroupLabelUpdate = "profile.group_label.update"
)

// fieldRoleLabel is the invalid_input field of a role label outside the label contract.
const fieldRoleLabel = "role_label"

// GroupMemberCard is one Group member card: the Group, its name, the member's role label in it (nil: none), whether the
// caller may edit the label (a hint; every write re-decides) and the label's optimistic version (0: no label row yet).
// It never carries a permission role, another member, a count, an email or a Business Profile.
type GroupMemberCard struct {
	GroupID   tenancy.GroupID
	GroupName string
	RoleLabel *string
	CanEdit   bool
	Version   int32
}

// GroupMemberCards is the complete list of the subject's cards the caller may see now, in Group id order. When more than
// the bound would be shown in Organization-only context, Cards is empty and NarrowRequired is set (no total, no cursor,
// no partial list); a selected-Group context shows at most one card and never sets it.
type GroupMemberCards struct {
	Subject        tenancy.UserID
	Cards          []GroupMemberCard
	NarrowRequired bool
}

// groupProfileFailure maps an error of a Group member profile statement: a lost version race or a guard refusal is the
// version conflict, a binding refusal (the membership ended under the write) not_found, the label CHECK invalid_input
// role_label; a store failure is unavailable with the transaction's reason.
func groupProfileFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	var pg *pgconn.PgError
	switch {
	case errors.Is(err, tenancy.ErrGroupMemberLabelAbsent), errors.Is(err, pgx.ErrNoRows):
		return conflict(fieldVersionConflict)
	case errors.As(err, &pg) && pg.ConstraintName == "group_member_profiles_guard":
		return conflict(fieldVersionConflict)
	case errors.As(err, &pg) && pg.ConstraintName == "group_member_profiles_binding":
		return notFound()
	case errors.As(err, &pg) && pg.Code == "23514" && pg.ConstraintName == "group_member_profiles_label_valid":
		return invalid(fieldRoleLabel)
	case errors.As(err, &pg) && commandError(pg) && len(pg.Code) == 5 && pg.Code[:2] != "40" && pg.Code[:2] != "08" &&
		pg.Code[:2] != "57":
		return fail(CodeInternal, reasonStoreUnavailable)
	case tenancy.CodeOf(err) != "":
		return fail(CodeInternal, reasonFactInconsistent)
	}
	return fail(CodeUnavailable, failureReason(ctx, tx.cause))
}

// ListGroupMemberCards returns the subject's Group member cards as the caller may see them now
// (group_member_profile.read): nothing unless G074's relationship predicate admits the subject (not_found, the same
// identity as GetAccountProfile, recorded as a target_not_found denial without the subject); then one card per current
// Group membership of the subject in the active Organization that the caller may see (self, a Group the caller is also a
// member of, or any of them for a current Organization owner/admin under either override setting), restricted to the
// selected Group in selected-Group context. A visible subject with no card is an empty list.
func (s *Service) ListGroupMemberCards(ctx context.Context, tc *TenantContext, subject tenancy.UserID) (GroupMemberCards, error) {
	var out GroupMemberCards
	err := s.guard(ctx, tc, opProfileGroupsList, capGroupMemberProfileRead, false, Target{Kind: kindOrganization}, organizationQuery,
		targetID(subject), func(tx *trackedTx, f Facts) *failure {
			if _, err := tenancy.VisibleProfileTx(ctx, tx, subject); err != nil {
				if errors.Is(err, tenancy.ErrProfileAbsent) {
					return markDenied(notFound(), tc.selection(), f, "", audit.UUID{}, tenancy.GroupID{})
				}
				return groupProfileFailure(ctx, tx, err)
			}
			cards, err := tenancy.GroupMemberCardsTx(ctx, tx, f.Organization.ID, subject)
			if err != nil {
				return groupProfileFailure(ctx, tx, err)
			}
			out = GroupMemberCards{Subject: subject, Cards: []GroupMemberCard{}}
			if len(cards) > tenancy.GroupMemberCardsMax {
				// The overflow signal is computed after every authorization and context filter: only authorized cards
				// count, and no number is disclosed.
				out.NarrowRequired = true
				return nil
			}
			for _, c := range cards {
				if c.GroupName == "" {
					// Every Group a card names is readable by the viewer the projection admitted.
					return fail(CodeInternal, reasonFactInconsistent)
				}
				out.Cards = append(out.Cards, GroupMemberCard{GroupID: c.GroupID, GroupName: c.GroupName, RoleLabel: c.RoleLabel,
					CanEdit: c.CanEdit, Version: c.Version})
			}
			return nil
		})
	if err != nil {
		return GroupMemberCards{}, err
	}
	return out, nil
}

// SetGroupMemberLabel sets the subject's role label in Group g when its label row is still at expected (0: no row yet):
// group_member_profile.manage_self when the subject is the caller, group_member_profile.manage otherwise (the exact
// Group's own owner or admin; an Organization owner or admin without that role is forbidden under either override
// setting). nil, or a value empty after trimming, clears it. The label contract is checked before any transaction.
// Setting the current value is a successful no-op with no statement and no audit row, whatever expected says; a
// different value on a stale expected is conflict version_conflict with zero effects. The subject's membership binding
// is read and locked in the same transaction after the Group lock (the removal's lock order), so a membership revoked
// before COMMIT refuses: no current membership is not_found, like an unknown, foreign or (in selected-Group context)
// another Group. The first edit inserts the row at version 1; a concurrent first edit loses on the primary key and is
// compared again. It returns the subject's card for g.
func (s *Service) SetGroupMemberLabel(ctx context.Context, tc *TenantContext, g tenancy.GroupID, subject tenancy.UserID, expected int32,
	label *string) (GroupMemberCard, error) {
	var out GroupMemberCard
	var want *string
	pre := func() *failure {
		if f := targetID(g)(); f != nil {
			return f
		}
		if f := targetID(subject)(); f != nil {
			return f
		}
		if expected < 0 {
			return invalid("expected_version")
		}
		var err error
		if want, err = tenancy.NormalizeRoleLabel(label); err != nil {
			return invalid(fieldRoleLabel)
		}
		return nil
	}
	c := capGroupMemberProfileManage
	if tc != nil && subject == tc.subject {
		c = capGroupMemberProfileManageSelf
	}
	sc := Scope{Kind: tenancy.ScopeGroup, Group: &g}
	err := s.guard(ctx, tc, opProfileGroupLabelUpdate, c, true, Target{Kind: kindGroup, Group: g}, sc.query(true, &subject, false, false), pre,
		func(tx *trackedTx, f Facts) *failure {
			b, ok := subjectBinding(f, sc, subject)
			if !ok || b.GroupID == nil || *b.GroupID != g || b.Source != tenancy.SourceMembership {
				return markDenied(notFound(), tc.selection(), f, "", audit.UUID{}, tenancy.GroupID{})
			}
			org := f.Organization.ID
			cur, err := tenancy.LockGroupMemberLabelTx(ctx, tx, org, b.ID)
			present := err == nil
			if err != nil && !errors.Is(err, tenancy.ErrGroupMemberLabelAbsent) {
				return groupProfileFailure(ctx, tx, err)
			}
			switch {
			case sameName(cur.RoleLabel, want):
				// No statement: the value is already current.
			case cur.Version != expected:
				return conflict(fieldVersionConflict)
			case present:
				if _, err := tenancy.UpdateGroupMemberLabelTx(ctx, tx, org, b.ID, cur.Version, want); err != nil {
					return groupProfileFailure(ctx, tx, err)
				}
			default:
				// want is not nil here: an absent row's label is nil, and an equal value took the first branch.
				if _, err := tenancy.InsertGroupMemberLabelTx(ctx, tx, org, b.ID, g, subject, *want); err != nil {
					if !errors.Is(err, tenancy.ErrGroupMemberLabelAbsent) {
						return groupProfileFailure(ctx, tx, err)
					}
					// A concurrent first edit committed its row: only an identical value is still a success.
					if cur, err = tenancy.LockGroupMemberLabelTx(ctx, tx, org, b.ID); err != nil {
						return groupProfileFailure(ctx, tx, err)
					}
					if !sameName(cur.RoleLabel, want) {
						return conflict(fieldVersionConflict)
					}
				}
			}
			now, err := tenancy.LockGroupMemberLabelTx(ctx, tx, org, b.ID)
			if err != nil && !errors.Is(err, tenancy.ErrGroupMemberLabelAbsent) {
				return groupProfileFailure(ctx, tx, err)
			}
			out = GroupMemberCard{GroupID: g, GroupName: findGroup(f, g).Name, RoleLabel: now.RoleLabel, CanEdit: true, Version: now.Version}
			return nil
		})
	if err != nil {
		return GroupMemberCard{}, err
	}
	return out, nil
}
