package tenancy

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Group member profiles (G075 Plan r2 §3.1-§3.4; migration 00024_group_member_profiles.sql). These are
// transaction-scoped model primitives for the guarded Group member profile operations of core/internal/tenancy/access:
// they run in a transaction the caller owns (with the request context initialized first), enforce no capability and
// never log. Row-level security admits the two writer paths (the member's own row; any row of a Group the caller owns
// or administers) for the label statements, and every other read goes through the policy reader's card projection
// (app.rls_group_member_cards); every statement names its Organization ($1), and the guard and capture triggers own the
// binding check, the version rules and the audit row. Errors are returned unmapped (the caller maps constraint names); a
// missing or invisible row is ErrGroupMemberLabelAbsent.

// ErrGroupMemberLabelAbsent reports that the membership has no label row (or that the caller may not write it).
var ErrGroupMemberLabelAbsent = errors.New("tenancy: group member label absent")

// ErrRoleLabelInvalid reports a role label outside the label contract.
var ErrRoleLabelInvalid = errors.New("tenancy: invalid role label")

// RoleLabelMax bounds a role label, in code points (Plan r2 C5, Q1).
const RoleLabelMax = 100

// GroupMemberCardsMax is the largest card list a caller returns; the projection emits one more row at most, and a 201st
// row is the overflow signal (Plan r2 C8).
const GroupMemberCardsMax = 200

// GroupMemberCard is one card as the projection returned it: the Group, its name as the caller's own groups_select
// policy shows it, the label (nil: none), whether the caller may edit it and the label row's version (0: no row).
type GroupMemberCard struct {
	GroupID   GroupID
	GroupName string
	RoleLabel *string
	CanEdit   bool
	Version   int32
}

// GroupMemberLabel is a membership's label row as the writer reads it (Version 0: no row).
type GroupMemberLabel struct {
	BindingID BindingID
	RoleLabel *string
	Version   int32
}

// The Group member profile statements. The card read joins the Group name under the caller's own groups_select policy
// (every Group a card names is visible to the viewer the projection admitted: its own Group, a shared Group, or any
// Group for an Organization owner/admin); a card whose Group row the caller cannot read has no name and is refused by
// the facade. The guard requires version = OLD.version + 1 and a changed label and sets updated_at; the capture records
// group_member_profile.updated for a changed label only. A first row is inserted with its label at version 1 (a clear of
// a label that was never set writes nothing).
const (
	sqlGroupMemberCards = `SELECT c.group_id, g.name, c.role_label, c.can_edit, c.version FROM app.rls_group_member_cards($2) c ` +
		`LEFT JOIN app.groups g ON g.organization_id = $1 AND g.id = c.group_id ORDER BY c.group_id`
	sqlLockGroupMemberLabel   = `SELECT role_label, version FROM app.group_member_profiles WHERE organization_id = $1 AND binding_id = $2 FOR NO KEY UPDATE`
	sqlInsertGroupMemberLabel = `INSERT INTO app.group_member_profiles (organization_id, binding_id, group_id, user_id, role_label) ` +
		`VALUES ($1, $2, $3, $4, $5) ON CONFLICT (binding_id) DO NOTHING RETURNING version`
	sqlUpdateGroupMemberLabel = `UPDATE app.group_member_profiles SET role_label = $4, version = version + 1 ` +
		`WHERE organization_id = $1 AND binding_id = $2 AND version = $3 RETURNING version`
)

// NormalizeRoleLabel applies the role label contract (the G063/G074 grammar with a 100 code point bound): valid UTF-8,
// no control, line-separator or paragraph-separator character, trimmed of Unicode white space. nil, or a value that is
// empty after trimming, is nil (the label is cleared).
func NormalizeRoleLabel(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	if !utf8.ValidString(*raw) {
		return nil, ErrRoleLabelInvalid
	}
	for _, r := range *raw {
		if unicode.IsControl(r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return nil, ErrRoleLabelInvalid
		}
	}
	label := strings.TrimFunc(*raw, unicode.IsSpace)
	if label == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(label) > RoleLabelMax {
		return nil, ErrRoleLabelInvalid
	}
	return &label, nil
}

// GroupMemberCardsTx reads subject's cards in org as the caller may see them now, in Group id order: at most
// GroupMemberCardsMax+1 rows (the last one only as the overflow signal). A subject the relationship predicate refuses
// has no card.
func GroupMemberCardsTx(ctx context.Context, tx pgx.Tx, org OrganizationID, subject UserID) ([]GroupMemberCard, error) {
	if zero(org) || zero(subject) {
		return nil, newError(CodeInvalidID)
	}
	rows, err := tx.Query(ctx, sqlGroupMemberCards, uuidArg(org), uuidArg(subject))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []GroupMemberCard{}
	for rows.Next() {
		var g pgtype.UUID
		var name, label *string
		var c GroupMemberCard
		if err := rows.Scan(&g, &name, &label, &c.CanEdit, &c.Version); err != nil {
			return nil, err
		}
		if !g.Valid {
			return nil, newError(CodeInvalidID)
		}
		c.GroupID, c.RoleLabel = GroupID(g.Bytes), label
		if name != nil {
			c.GroupName = *name
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// LockGroupMemberLabelTx reads and locks a membership's label row FOR NO KEY UPDATE (ErrGroupMemberLabelAbsent when
// there is none; an absent row locks nothing, and a concurrent first insert loses on the primary key).
func LockGroupMemberLabelTx(ctx context.Context, tx pgx.Tx, org OrganizationID, binding BindingID) (GroupMemberLabel, error) {
	if zero(org) || zero(binding) {
		return GroupMemberLabel{}, newError(CodeInvalidID)
	}
	l := GroupMemberLabel{BindingID: binding}
	if err := tx.QueryRow(ctx, sqlLockGroupMemberLabel, uuidArg(org), uuidArg(binding)).Scan(&l.RoleLabel, &l.Version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GroupMemberLabel{}, ErrGroupMemberLabelAbsent
		}
		return GroupMemberLabel{}, err
	}
	return l, nil
}

// InsertGroupMemberLabelTx creates the label row of binding (the membership of subject in Group g of org) at version 1
// with label and returns its version. A row that already exists (a concurrent first edit won the primary key) is
// ErrGroupMemberLabelAbsent: nothing was written.
func InsertGroupMemberLabelTx(ctx context.Context, tx pgx.Tx, org OrganizationID, binding BindingID, g GroupID, subject UserID,
	label string) (int32, error) {
	if zero(org) || zero(binding) || zero(g) || zero(subject) {
		return 0, newError(CodeInvalidID)
	}
	var v int32
	if err := tx.QueryRow(ctx, sqlInsertGroupMemberLabel, uuidArg(org), uuidArg(binding), uuidArg(g), uuidArg(subject), label).Scan(&v); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrGroupMemberLabelAbsent
		}
		return 0, err
	}
	return v, nil
}

// UpdateGroupMemberLabelTx sets (or, with nil, clears) a membership's label when its row is still at version, advancing
// it by one, and returns the new version. A row at another version (or none) is ErrGroupMemberLabelAbsent.
func UpdateGroupMemberLabelTx(ctx context.Context, tx pgx.Tx, org OrganizationID, binding BindingID, version int32,
	label *string) (int32, error) {
	if zero(org) || zero(binding) {
		return 0, newError(CodeInvalidID)
	}
	var v int32
	if err := tx.QueryRow(ctx, sqlUpdateGroupMemberLabel, uuidArg(org), uuidArg(binding), version, label).Scan(&v); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrGroupMemberLabelAbsent
		}
		return 0, err
	}
	return v, nil
}
