package tenancy

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// G020 label primitives (Plan r2 §1.2, §1.3, §4.1; 00020_labels.sql). Like the other *Tx functions they run in the
// caller's transaction, carry no actor and authorize nothing: core/internal/tenancy/access decides the label capability
// and the target first, and row-level security repeats the owner-scope rule. Go identifiers say LabelDefinition and
// LabelAssignment (never a bare Label, which would read as the branding label of branding.go); the SQL relations are
// app.label_definitions and app.label_assignments. Every statement names its columns: none selects *, and none names a
// column that exists only in a test database's widened schema.

// LabelDefinitionID is the canonical UUID of a label definition (core-generated; keyed insert without RETURNING).
type LabelDefinitionID [16]byte

func (id LabelDefinitionID) String() string { return formatUUID(id) }

// ParseLabelDefinitionID parses a label definition ID; see ParseOrganizationID.
func ParseLabelDefinitionID(s string) (LabelDefinitionID, error) {
	return parseUUID[LabelDefinitionID](s)
}

// LabelAssignmentID is the canonical UUID of a label assignment (core-generated; keyed insert without RETURNING).
type LabelAssignmentID [16]byte

func (id LabelAssignmentID) String() string { return formatUUID(id) }

// LabelVisibility is who besides the owner scope's managers reads a definition: owner_scope (managers only) or
// target_members (also the members of a Group-owned definition's Group; Organization-owned definitions never).
type LabelVisibility string

// The two visibilities (label_definitions_visibility_valid).
const (
	LabelVisibilityOwnerScope    LabelVisibility = "owner_scope"
	LabelVisibilityTargetMembers LabelVisibility = "target_members"
)

// ParseLabelVisibility parses a visibility: exactly one of the two values, otherwise invalid_visibility.
func ParseLabelVisibility(s string) (LabelVisibility, error) {
	switch v := LabelVisibility(s); v {
	case LabelVisibilityOwnerScope, LabelVisibilityTargetMembers:
		return v, nil
	}
	return "", newError(CodeInvalidVisibility)
}

// Label bounds (Plan r2 §1.2, D-5; §4.1): the name length in code points, the definitions of one owner scope
// (label_definitions_bound) and the definition page sizes.
const (
	MaxLabelNameLength          = 80
	MaxLabelDefinitionsPerScope = 200
	DefaultLabelPage            = 50
	MaxLabelPage                = 100
)

// LabelDefinition is one row of app.label_definitions as the runtime role reads it (never the generated name key).
// GroupID is zero for an Organization-owned definition and the owning Group for a Group-owned one; OwnerScopeID is the
// Organization or the owning Group (label_definitions_scope_consistent).
type LabelDefinition struct {
	ID             LabelDefinitionID
	OrganizationID OrganizationID
	OwnerScopeType ScopeType
	OwnerScopeID   [16]byte
	GroupID        GroupID
	Name           string
	Visibility     LabelVisibility
	CreatedAt      time.Time
}

// LabelDefinitionSpec is a new definition: Group zero for the Organization scope, else the owning Group.
type LabelDefinitionSpec struct {
	ID             LabelDefinitionID
	OrganizationID OrganizationID
	Group          GroupID
	Name           string
	Visibility     LabelVisibility
}

// LabelAssignment is one row of app.label_assignments; LabelName is the definition's name, set by
// LabelAssignmentsByTarget only. The owner facts are copied from the definition (the total composite foreign key).
type LabelAssignment struct {
	ID             LabelAssignmentID
	OrganizationID OrganizationID
	LabelID        LabelDefinitionID
	OwnerScopeType ScopeType
	OwnerScopeID   [16]byte
	TargetType     string
	TargetID       [16]byte
	TargetGroupID  GroupID
	CreatedAt      time.Time
	LabelName      string
}

// LabelAssignmentSpec is a new assignment of Definition to the target (TargetType, TargetID) owned by TargetGroup (the
// Group itself for the group target). The owner facts come from Definition, never from the caller.
type LabelAssignmentSpec struct {
	ID          LabelAssignmentID
	Definition  LabelDefinition
	TargetType  string
	TargetID    [16]byte
	TargetGroup GroupID
}

// ValidateLabelName returns the stored form of a label name: the ValidateName rule (valid UTF-8, Unicode whitespace
// trimmed, no C0, DEL or C1 control character) with the tighter bound of 1-MaxLabelNameLength code points, and no line
// or paragraph separator (U+2028, U+2029) anywhere, before or after trimming (the G063 name class). Anything else is
// invalid_name. The database rule (label_definitions_name_valid: the same length bound, ASCII-space trim and the same
// character class) is never stricter than this rule.
func ValidateLabelName(s string) (string, error) {
	for _, r := range s {
		if r == 0x2028 || r == 0x2029 {
			return "", newError(CodeInvalidName)
		}
	}
	t, err := ValidateName(s)
	if err != nil {
		return "", err
	}
	if utf8.RuneCountInString(t) > MaxLabelNameLength {
		return "", newError(CodeInvalidName)
	}
	return t, nil
}

// labelNameKey is the uniqueness key of a stored name, as the generated column name_key = lower(name COLLATE "C")
// computes it: ASCII letters fold to lower case, every other byte is kept exactly (validated, never normalized).
func labelNameKey(name string) string {
	b := []byte(name)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

// validate checks a definition spec before any statement: nonzero ids, the name rule, a known visibility, and
// owner_scope for the Organization scope (visibility_not_allowed).
func (sp LabelDefinitionSpec) validate() (string, ScopeType, [16]byte, pgtype.UUID, error) {
	if zero(sp.OrganizationID) || zero(sp.ID) {
		return "", "", [16]byte{}, pgtype.UUID{}, newError(CodeInvalidID)
	}
	name, err := ValidateLabelName(sp.Name)
	if err != nil {
		return "", "", [16]byte{}, pgtype.UUID{}, err
	}
	if _, err := ParseLabelVisibility(string(sp.Visibility)); err != nil {
		return "", "", [16]byte{}, pgtype.UUID{}, err
	}
	if zero(sp.Group) {
		if sp.Visibility != LabelVisibilityOwnerScope {
			return "", "", [16]byte{}, pgtype.UUID{}, newError(CodeVisibilityNotAllowed)
		}
		return name, ScopeOrganization, [16]byte(sp.OrganizationID), pgtype.UUID{}, nil
	}
	return name, ScopeGroup, [16]byte(sp.Group), uuidArg(sp.Group), nil
}

// labelKeyCollisions: a collision of a core-generated id (either primary key, or the composite owner key that carries
// the definition id) is a store failure (unavailable), never a conflict the caller could act on.
var labelKeyCollisions = map[string]bool{
	"label_definitions_pkey": true, "label_definitions_owner_key": true, "label_assignments_pkey": true,
}

func classifyLabelKeyed(err error) *Error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" && labelKeyCollisions[pg.ConstraintName] {
		return newError(CodeUnavailable)
	}
	return classify(err)
}

// The label statements (Plan r2 §1.2-§1.4, §4.1), pinned by TestLabelStatements and counted with store.go's by
// TestSQLBoundary. $1 is always the Organization. The INSERT column lists are exactly the runtime role's column grants
// (00020); UPDATE sets name and visibility only; assignments are never updated.
const (
	labelDefinitionColumns = `id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility, created_at`
	labelAssignmentColumns = `a.id, a.organization_id, a.label_id, a.owner_scope_type, a.owner_scope_id, a.target_type, a.target_id, ` +
		`a.target_group_id, a.created_at, d.name`

	// One keyed INSERT without RETURNING (the id is core-generated). The BEFORE INSERT trigger
	// app.label_definitions_bound re-locks the owner scope row and counts (scope_full).
	sqlInsertLabelDefinition = `INSERT INTO app.label_definitions (id, organization_id, owner_scope_type, owner_scope_id, group_id, ` +
		`name, visibility) VALUES ($2, $1, $3, $4, $5, $6, $7)`
	// The definition read by id, and its two locking forms taken after the facts (AUTHORIZATION.md step 7): FOR NO KEY
	// UPDATE for update and delete, FOR SHARE for assign and unassign (O-LAB-1).
	sqlSelectLabelDefinition = `SELECT ` + labelDefinitionColumns + ` FROM app.label_definitions WHERE organization_id = $1 AND id = $2`
	sqlLockLabelDefinition   = `SELECT ` + labelDefinitionColumns + ` FROM app.label_definitions WHERE organization_id = $1 AND id = $2 ` +
		`FOR NO KEY UPDATE`
	sqlShareLabelDefinition = `SELECT ` + labelDefinitionColumns + ` FROM app.label_definitions WHERE organization_id = $1 AND id = $2 ` +
		`FOR SHARE`
	// The keyset page of one owner scope (cursor = definition id, limit+1 rows), optionally restricted to one visibility
	// (the target_members subset of a Group reader; row-level security applies the same rule).
	sqlLabelDefinitionsPage = `SELECT ` + labelDefinitionColumns + ` FROM app.label_definitions ` +
		`WHERE organization_id = $1 AND owner_scope_type = $2 AND owner_scope_id = $3 AND ($4::text IS NULL OR visibility = $4) ` +
		`AND id > $5 ORDER BY id LIMIT $6`
	// The absolute set of name and visibility (a no-op set records nothing: the capture trigger's WHEN clauses).
	sqlUpdateLabelDefinition = `UPDATE app.label_definitions SET name = $3, visibility = $4 WHERE organization_id = $1 AND id = $2 ` +
		`RETURNING ` + labelDefinitionColumns
	// The two-statement definition delete (§1.3): its assignments, then the definition.
	sqlDeleteLabelDefinitionAssignments = `DELETE FROM app.label_assignments WHERE organization_id = $1 AND label_id = $2`
	sqlDeleteLabelDefinition            = `DELETE FROM app.label_definitions WHERE organization_id = $1 AND id = $2`
	// The labels on one target with their names (label_assignments_target_idx); row-level security on both relations
	// admits only readers of the definition.
	sqlLabelAssignmentsByTarget = `SELECT ` + labelAssignmentColumns + ` FROM app.label_assignments a ` +
		`JOIN app.label_definitions d ON d.organization_id = a.organization_id AND d.id = a.label_id ` +
		`WHERE a.organization_id = $1 AND a.target_type = $2 AND a.target_id = $3 ORDER BY a.label_id`
	// One keyed INSERT without RETURNING; the owner facts are the definition's (label_assignments_label_fkey).
	sqlInsertLabelAssignment = `INSERT INTO app.label_assignments (id, organization_id, label_id, owner_scope_type, owner_scope_id, ` +
		`target_type, target_id, target_group_id) VALUES ($2, $1, $3, $4, $5, $6, $7, $8)`
	sqlDeleteLabelAssignment = `DELETE FROM app.label_assignments WHERE organization_id = $1 AND label_id = $2 AND target_type = $3 ` +
		`AND target_id = $4`
)

// InsertLabelDefinitionTx validates sp and creates the definition inside the caller's transaction, after the facade
// has locked the owner scope row (the Organization or the owning Group FOR NO KEY UPDATE): one keyed INSERT without
// RETURNING, then the read of the new row, which its manager can see. A taken name key is name_conflict, a full scope
// scope_full; a row the read cannot see is unavailable.
func InsertLabelDefinitionTx(ctx context.Context, tx pgx.Tx, sp LabelDefinitionSpec) (LabelDefinition, error) {
	name, scope, scopeID, group, err := sp.validate()
	if err != nil {
		return LabelDefinition{}, err
	}
	if _, err := tx.Exec(ctx, sqlInsertLabelDefinition, uuidArg(sp.OrganizationID), uuidArg(sp.ID), string(scope),
		pgtype.UUID{Bytes: scopeID, Valid: true}, group, name, string(sp.Visibility)); err != nil {
		return LabelDefinition{}, classifyLabelKeyed(err)
	}
	d, err := scanLabelDefinition(tx.QueryRow(ctx, sqlSelectLabelDefinition, uuidArg(sp.OrganizationID), uuidArg(sp.ID)))
	switch {
	case isNoRows(err):
		return LabelDefinition{}, newError(CodeUnavailable)
	case err != nil:
		return LabelDefinition{}, classify(err)
	}
	return d, nil
}

// LabelDefinitionTx reads definition id of org inside the caller's transaction, as row-level security admits it. A row
// the context cannot see is label_not_found.
func LabelDefinitionTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id LabelDefinitionID) (LabelDefinition, error) {
	return readLabelDefinition(ctx, tx, sqlSelectLabelDefinition, org, id)
}

// LockLabelDefinitionTx reads and locks definition id of org: FOR NO KEY UPDATE when exclusive (update, delete), FOR
// SHARE otherwise (assign, unassign). A row the context cannot see, or one deleted before the lock was granted, is
// label_not_found.
func LockLabelDefinitionTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id LabelDefinitionID, exclusive bool) (LabelDefinition, error) {
	if exclusive {
		return readLabelDefinition(ctx, tx, sqlLockLabelDefinition, org, id)
	}
	return readLabelDefinition(ctx, tx, sqlShareLabelDefinition, org, id)
}

func readLabelDefinition(ctx context.Context, tx pgx.Tx, sql string, org OrganizationID, id LabelDefinitionID) (LabelDefinition, error) {
	if zero(org) || zero(id) {
		return LabelDefinition{}, newError(CodeInvalidID)
	}
	d, err := scanLabelDefinition(tx.QueryRow(ctx, sql, uuidArg(org), uuidArg(id)))
	if e := classify(one(err, CodeLabelNotFound)); e != nil {
		return LabelDefinition{}, e
	}
	return d, nil
}

// LabelDefinitionsPage reads at most limit+1 definitions of one owner scope of org (group zero: the Organization
// scope) ordered by id after after (zero: from the start), so the caller can tell whether a next page exists. A
// nonempty visibility restricts the page to that visibility. limit is 1..MaxLabelPage.
func LabelDefinitionsPage(ctx context.Context, tx pgx.Tx, org OrganizationID, group GroupID, visibility LabelVisibility,
	after LabelDefinitionID, limit int) ([]LabelDefinition, error) {
	if zero(org) || limit < 1 || limit > MaxLabelPage {
		return nil, newError(CodeInvalidID)
	}
	var vis any
	if visibility != "" {
		v, err := ParseLabelVisibility(string(visibility))
		if err != nil {
			return nil, err
		}
		vis = string(v)
	}
	scope, scopeID := scopeOf(org, group)
	rows, err := tx.Query(ctx, sqlLabelDefinitionsPage, uuidArg(org), string(scope), scopeID, vis, uuidArg(after), limit+1)
	out, err := collect(rows, err, scanLabelDefinition)
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

// UpdateLabelDefinitionTx sets the name and visibility of definition id of org absolutely and returns the row. Setting
// the current values succeeds without effect. A row the context cannot see or manage is label_not_found; target_members
// on an Organization-owned definition is visibility_not_allowed (the facade refuses it before this call).
func UpdateLabelDefinitionTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id LabelDefinitionID, name string,
	visibility LabelVisibility) (LabelDefinition, error) {
	if zero(org) || zero(id) {
		return LabelDefinition{}, newError(CodeInvalidID)
	}
	n, err := ValidateLabelName(name)
	if err != nil {
		return LabelDefinition{}, err
	}
	if _, err := ParseLabelVisibility(string(visibility)); err != nil {
		return LabelDefinition{}, err
	}
	d, err := scanLabelDefinition(tx.QueryRow(ctx, sqlUpdateLabelDefinition, uuidArg(org), uuidArg(id), n, string(visibility)))
	if e := classify(one(err, CodeLabelNotFound)); e != nil {
		return LabelDefinition{}, e
	}
	return d, nil
}

// DeleteLabelDefinitionAssignmentsTx deletes every assignment of definition label of org visible to the context and
// returns their number: the first statement of the definition delete, under the definition's FOR NO KEY UPDATE lock.
func DeleteLabelDefinitionAssignmentsTx(ctx context.Context, tx pgx.Tx, org OrganizationID, label LabelDefinitionID) (int64, error) {
	if zero(org) || zero(label) {
		return 0, newError(CodeInvalidID)
	}
	tag, err := tx.Exec(ctx, sqlDeleteLabelDefinitionAssignments, uuidArg(org), uuidArg(label))
	if err != nil {
		return 0, classify(err)
	}
	return tag.RowsAffected(), nil
}

// DeleteLabelDefinitionTx deletes definition id of org: the second statement of the definition delete. Zero rows (a
// row the context cannot see or manage) is label_not_found; a remaining assignment is refused by the foreign key.
func DeleteLabelDefinitionTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id LabelDefinitionID) error {
	if zero(org) || zero(id) {
		return newError(CodeInvalidID)
	}
	tag, err := tx.Exec(ctx, sqlDeleteLabelDefinition, uuidArg(org), uuidArg(id))
	if err != nil {
		return classify(err)
	}
	if tag.RowsAffected() == 0 {
		return newError(CodeLabelNotFound)
	}
	return nil
}

// LabelAssignmentsByTarget reads the assignments on target (targetType, targetID) of org with their definition names,
// ordered by definition id, as row-level security admits them. The rows are bounded by MaxLabelDefinitionsPerScope per
// owner scope.
func LabelAssignmentsByTarget(ctx context.Context, tx pgx.Tx, org OrganizationID, targetType string, targetID [16]byte) ([]LabelAssignment, error) {
	if zero(org) || targetID == ([16]byte{}) || targetType == "" {
		return nil, newError(CodeInvalidID)
	}
	rows, err := tx.Query(ctx, sqlLabelAssignmentsByTarget, uuidArg(org), targetType, pgtype.UUID{Bytes: targetID, Valid: true})
	out, err := collect(rows, err, scanLabelAssignment)
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

// InsertLabelAssignmentTx assigns sp.Definition to the target inside the caller's transaction, after the facade has
// locked the target Group and the definition (FOR SHARE): one keyed INSERT without RETURNING. An existing assignment
// of the definition to the target is already_assigned; a shape production storage does not admit is target_unsupported.
func InsertLabelAssignmentTx(ctx context.Context, tx pgx.Tx, sp LabelAssignmentSpec) error {
	d := sp.Definition
	if zero(sp.ID) || zero(d.ID) || zero(d.OrganizationID) || d.OwnerScopeID == ([16]byte{}) || sp.TargetID == ([16]byte{}) ||
		zero(sp.TargetGroup) || sp.TargetType == "" {
		return newError(CodeInvalidID)
	}
	if d.OwnerScopeType != ScopeOrganization && d.OwnerScopeType != ScopeGroup {
		return newError(CodeInvalidScope)
	}
	if _, err := tx.Exec(ctx, sqlInsertLabelAssignment, uuidArg(d.OrganizationID), uuidArg(sp.ID), uuidArg(d.ID),
		string(d.OwnerScopeType), pgtype.UUID{Bytes: d.OwnerScopeID, Valid: true}, sp.TargetType,
		pgtype.UUID{Bytes: sp.TargetID, Valid: true}, uuidArg(sp.TargetGroup)); err != nil {
		return classifyLabelKeyed(err)
	}
	return nil
}

// DeleteLabelAssignmentTx removes the assignment of definition label to target (targetType, targetID) of org. Zero
// rows is not_assigned.
func DeleteLabelAssignmentTx(ctx context.Context, tx pgx.Tx, org OrganizationID, label LabelDefinitionID, targetType string,
	targetID [16]byte) error {
	if zero(org) || zero(label) || targetID == ([16]byte{}) || targetType == "" {
		return newError(CodeInvalidID)
	}
	tag, err := tx.Exec(ctx, sqlDeleteLabelAssignment, uuidArg(org), uuidArg(label), targetType, pgtype.UUID{Bytes: targetID, Valid: true})
	if err != nil {
		return classify(err)
	}
	if tag.RowsAffected() == 0 {
		return newError(CodeNotAssigned)
	}
	return nil
}

func scanLabelDefinition(row pgx.Row) (LabelDefinition, error) {
	var d LabelDefinition
	var id, org, scope, group pgtype.UUID
	var scopeType, visibility string
	if err := row.Scan(&id, &org, &scopeType, &scope, &group, &d.Name, &visibility, &d.CreatedAt); err != nil {
		return LabelDefinition{}, err
	}
	d.ID, d.OrganizationID, d.OwnerScopeType, d.OwnerScopeID = LabelDefinitionID(id.Bytes), OrganizationID(org.Bytes), ScopeType(scopeType), scope.Bytes
	d.Visibility = LabelVisibility(visibility)
	if group.Valid {
		d.GroupID = GroupID(group.Bytes)
	}
	return d, nil
}

func scanLabelAssignment(row pgx.Row) (LabelAssignment, error) {
	var a LabelAssignment
	var id, org, label, scope, target, group pgtype.UUID
	var scopeType string
	if err := row.Scan(&id, &org, &label, &scopeType, &scope, &a.TargetType, &target, &group, &a.CreatedAt, &a.LabelName); err != nil {
		return LabelAssignment{}, err
	}
	a.ID, a.OrganizationID, a.LabelID = LabelAssignmentID(id.Bytes), OrganizationID(org.Bytes), LabelDefinitionID(label.Bytes)
	a.OwnerScopeType, a.OwnerScopeID, a.TargetID, a.TargetGroupID = ScopeType(scopeType), scope.Bytes, target.Bytes, GroupID(group.Bytes)
	return a, nil
}
