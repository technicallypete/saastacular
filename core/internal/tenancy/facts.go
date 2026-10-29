package tenancy

import (
	"bytes"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"core/internal/observability"
)

// Transaction-scoped reads and writes for the G007 guarded facade (core/internal/tenancy/access; docs/foundation/
// AUTHORIZATION.md). They run in a transaction the caller owns, which is the only way a policy decision and its write
// share one database unit. Like the Service methods they are model primitives for trusted in-process callers: they
// carry no actor and enforce no capability, they never log, and the production import boundary (only the access
// package may import tenancy) is what keeps them from untrusted callers.

// OrganizationPolicy is the Organization policy: the G007 flag (00005_organization_policy.sql, NOT NULL default true) and
// the G060 Group-management override (00017_group_authority.sql, NOT NULL default false). Both are NOT NULL columns, so
// neither has a missing or malformed state; a read failure is a classified error, never an allow.
type OrganizationPolicy struct {
	GroupAdminsEditBusinessProfiles bool
	// GroupManagementOverride lets current Organization owners and admins administer ordinary Group metadata, members and
	// invitations without a Group binding (G060 Spec R2). It never confers Group content, ownership or membership.
	GroupManagementOverride bool
}

// BindingFact is one current membership binding of the subject, as read by ReadFacts. GroupID is zero for an
// Organization-scope binding. Role and Source are the stored values, unvalidated; classifying them is the evaluator's.
type BindingFact struct {
	OrganizationID OrganizationID
	GroupID        GroupID
	Role           string
	Source         string
}

// AssociationFact is one Group-BusinessProfile association of a read Group.
type AssociationFact struct {
	GroupID           GroupID
	BusinessProfileID BusinessProfileID
}

// Facts are the raw authorization facts of one subject in one Organization, read by ReadFacts in one transaction. They
// are facts, not decisions, and are never cached: every guarded operation reads its own.
type Facts struct {
	Organization         Organization // zero ID when the Organization does not exist
	Policy               OrganizationPolicy
	OrganizationBindings []BindingFact // the subject's Organization-scope bindings in the Organization
	GroupBindings        []BindingFact // the subject's Group bindings of the candidate Groups
	Groups               []Group       // the read Group rows of the Organization (locked in ascending id order)
	Associations         []AssociationFact
	Profiles             []BusinessProfile // the target profile, or every profile of the Organization
	// G019: the Subject fact — the locked scope bindings of the subject (and, with SubjectAndActor, of the actor) in the
	// fact scope, ordered by binding id; rows the context cannot see are absent. OwnerCount is app.rls_owner_count of the
	// fact scope, read after the scope locks (0 when the context cannot see the scope).
	SubjectBindings []Binding
	OwnerCount      int
}

// FactScope returns the scope a FactQuery's Subject and OwnerCount facts address: the TargetGroup's Group scope when
// a target is set, otherwise the Organization scope.
func (q FactQuery) FactScope() (ScopeType, [16]byte) {
	if !zero(q.TargetGroup) {
		return ScopeGroup, q.TargetGroup
	}
	return ScopeOrganization, q.Organization
}

// FactQuery selects what ReadFacts reads. Groups nil means every Group in which the subject holds a binding in the
// Organization; a non-nil empty Groups reads no Group fact. TargetGroup, when set, is read first and, with Lock,
// FOR NO KEY UPDATE (the row a Group write updates); it is a candidate too. Associations are read when Profile is set,
// AllProfiles is true or Associations is true (G018: a Group's onboarding state without any profile read). Lock takes row locks (a READ COMMITTED read-write transaction); without it every statement is
// a plain read (a REPEATABLE READ snapshot). ExclusiveOrganization locks the Organization row FOR NO KEY UPDATE instead
// of FOR SHARE (the policy write).
type FactQuery struct {
	Organization          OrganizationID
	User                  UserID
	Groups                []GroupID
	TargetGroup           GroupID
	Profile               BusinessProfileID
	AllProfiles           bool
	Associations          bool
	Lock                  bool
	ExclusiveOrganization bool
	// G019 (Plan r1 §1.4 "Guard shape"). Subject, with Lock, reads and locks FOR NO KEY UPDATE the target subject's
	// binding in the fact scope (FactScope) after the Organization, own-binding and Group locks; SubjectAndActor adds the
	// actor's (User's) binding of that scope to the same statement (a transfer). Subject requires Lock: a snapshot query
	// naming a subject is refused with invalid_id before any statement. OwnerCount reads the fact scope's owner count
	// after the scope locks (one definer call).
	Subject         *UserID
	SubjectAndActor bool
	OwnerCount      bool
}

// ReadFacts performs the ordered fact reads of the authorization contract in tx: the Organization row and its policy
// (FOR SHARE, or FOR NO KEY UPDATE when exclusive); the subject's Organization binding; the subject's Group bindings of
// the candidate Groups in group_id order; the candidate Groups in ascending id order, the target FOR NO KEY UPDATE and
// the others FOR SHARE (G018: others below the target id, the target, others above it); the associations of those Groups in a new statement; and the target or all profiles. In lock
// mode every relied-on mutable fact is row-locked by its own statement until the transaction ends. Every statement names
// the Organization. An absent row is an absent fact, never an error; a store failure is a classified error (the
// transaction is then unusable).
func ReadFacts(ctx context.Context, tx pgx.Tx, q FactQuery) (Facts, error) {
	var f Facts
	if zero(q.Organization) || zero(q.User) || (q.Subject != nil && (!q.Lock || zero(*q.Subject))) {
		return Facts{}, newError(CodeInvalidID)
	}
	org, user := uuidArg(q.Organization), uuidArg(q.User)

	orgSQL := sqlFactOrganization
	switch {
	case q.Lock && q.ExclusiveOrganization:
		orgSQL = sqlFactOrganizationExclusive
	case q.Lock:
		orgSQL = sqlFactOrganizationShare
	}
	var id pgtype.UUID
	err := tx.QueryRow(ctx, orgSQL, org).Scan(&id, &f.Organization.Name, &f.Organization.CreatedAt, &f.Policy.GroupAdminsEditBusinessProfiles,
		&f.Policy.GroupManagementOverride)
	if err != nil && !isNoRows(err) {
		return Facts{}, classify(err)
	}
	f.Organization.ID = OrganizationID(id.Bytes)

	bindingSQL := sqlFactOrganizationBinding
	if q.Lock {
		bindingSQL = sqlFactOrganizationBindingLock
	}
	rows, err := tx.Query(ctx, bindingSQL, org, user)
	if f.OrganizationBindings, err = collect(rows, err, scanOrganizationBindingFact); err != nil {
		return Facts{}, classify(err)
	}

	candidates := q.Groups
	if candidates != nil && !zero(q.TargetGroup) && !containsGroup(candidates, q.TargetGroup) {
		candidates = append(append([]GroupID(nil), candidates...), q.TargetGroup)
	}
	if candidates == nil || len(candidates) > 0 {
		var sql string
		var args []any
		switch {
		case candidates == nil && q.Lock:
			sql, args = sqlFactAllGroupBindingsLock, []any{org, user}
		case candidates == nil:
			sql, args = sqlFactAllGroupBindings, []any{org, user}
		case q.Lock:
			sql, args = sqlFactGroupBindingsLock, []any{org, user, groupArgs(candidates)}
		default:
			sql, args = sqlFactGroupBindings, []any{org, user, groupArgs(candidates)}
		}
		rows, err := tx.Query(ctx, sql, args...)
		if f.GroupBindings, err = collect(rows, err, scanGroupBindingFact); err != nil {
			return Facts{}, classify(err)
		}
		if candidates == nil {
			for _, b := range f.GroupBindings {
				if !containsGroup(candidates, b.GroupID) {
					candidates = append(candidates, b.GroupID)
				}
			}
			if candidates == nil {
				candidates = []GroupID{}
			}
		}
	}

	// Group row locks are taken in ascending Group id, each with its own strength (G018 Plan r5 §1.4): with Lock and a
	// target, the other candidates below the target id are locked FOR SHARE first, then the target FOR NO KEY UPDATE,
	// then the others above it FOR SHARE. Two transactions therefore never wait on each other's Group rows in a cycle.
	// Without a target (one ascending statement) or without Lock (plain reads) the order is as before.
	readOthers := func(others []GroupID) error {
		if len(others) == 0 {
			return nil
		}
		rows, err := tx.Query(ctx, pick(q.Lock, sqlFactGroupsLock, sqlFactGroups), org, groupArgs(others))
		gs, err := collect(rows, err, scanGroup)
		if err != nil {
			return err
		}
		f.Groups = append(f.Groups, gs...)
		return nil
	}
	if zero(q.TargetGroup) {
		if err := readOthers(candidates); err != nil {
			return Facts{}, classify(err)
		}
	} else {
		var below, above []GroupID
		for _, c := range candidates {
			switch {
			case c == q.TargetGroup:
			case q.Lock && bytes.Compare(c[:], q.TargetGroup[:]) < 0:
				below = append(below, c)
			default:
				above = append(above, c)
			}
		}
		if err := readOthers(below); err != nil {
			return Facts{}, classify(err)
		}
		g, err := scanGroup(tx.QueryRow(ctx, pick(q.Lock, sqlFactTargetGroupLock, sqlSelectGroup), org, uuidArg(q.TargetGroup)))
		switch {
		case err == nil:
			f.Groups = append(f.Groups, g)
		case !isNoRows(err):
			return Facts{}, classify(err)
		}
		if err := readOthers(above); err != nil {
			return Facts{}, classify(err)
		}
	}

	// G019: the Subject fact after every scope lock, then the owner count of the same scope.
	scope, scopeID := q.FactScope()
	if q.Subject != nil {
		subjects := []UserID{*q.Subject}
		if q.SubjectAndActor && q.User != *q.Subject {
			subjects = append(subjects, q.User)
		}
		users := make([]pgtype.UUID, len(subjects))
		for i, u := range subjects {
			users[i] = uuidArg(u)
		}
		rows, err := tx.Query(ctx, sqlSubjectBindingsLock, org, users, string(scope), uuidArg(scopeID))
		if f.SubjectBindings, err = collect(rows, err, scanFactBinding); err != nil {
			return Facts{}, classify(err)
		}
	}
	if q.OwnerCount {
		n, err := ownerCount(ctx, tx, scope, scopeID)
		if err != nil {
			return Facts{}, classify(err)
		}
		f.OwnerCount = n
	}

	if (!zero(q.Profile) || q.AllProfiles || q.Associations) && len(f.Groups) > 0 {
		ids := make([]GroupID, len(f.Groups))
		for i, g := range f.Groups {
			ids[i] = g.ID
		}
		rows, err := tx.Query(ctx, sqlFactAssociations, org, groupArgs(ids))
		if f.Associations, err = collect(rows, err, scanAssociationFact); err != nil {
			return Facts{}, classify(err)
		}
	}

	switch {
	case q.AllProfiles:
		rows, err := tx.Query(ctx, sqlFactAllProfiles, org)
		if f.Profiles, err = collect(rows, err, scanBusinessProfile); err != nil {
			return Facts{}, classify(err)
		}
	case !zero(q.Profile):
		p, err := scanBusinessProfile(tx.QueryRow(ctx, sqlSelectBusinessProfile, org, uuidArg(q.Profile)))
		switch {
		case err == nil:
			f.Profiles = []BusinessProfile{p}
		case !isNoRows(err):
			return Facts{}, classify(err)
		}
	}
	return f, nil
}

// InitContext initializes the transaction-local request context of tx (G008; docs/foundation/DATABASE-ISOLATION.md):
// the user, the active Organization and the selected Group (zero: Organization-only, explicitly cleared), with no login
// binding, in one statement that must be the first statement of the transaction. Row-level security then admits only
// rows of that context. A zero Organization is accepted only with a zero Group: the IDENTITY shape (the user alone),
// used for the Organization-less audit records of G011. The same statement carries the request and trace correlation
// of ctx as the diagnostic audit correlation (G011; docs/foundation/AUDIT.md). It returns invalid_id for a zero user,
// or a Group without an Organization (nothing is sent), context_present when the connection already carries a context
// value (the caller must discard the connection: a session-level value would otherwise survive the transaction), or
// the classified store failure.
func InitContext(ctx context.Context, tx pgx.Tx, user UserID, org OrganizationID, group GroupID) error {
	if zero(user) || (zero(org) && !zero(group)) {
		return newError(CodeInvalidID)
	}
	o, g := "", ""
	if !zero(org) {
		o = org.String()
	}
	if !zero(group) {
		g = group.String()
	}
	var a, b, c, d, e, f, r string
	err := tx.QueryRow(ctx, sqlInitContext, user.String(), o, g, "", observability.RequestID(ctx), observability.TraceID(ctx)).
		Scan(&a, &b, &c, &d, &e, &f, &r)
	if isNoRows(err) {
		return newError(CodeContextPresent)
	}
	if e := classify(err); e != nil {
		return e
	}
	return nil
}

func isNoRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

func pick(lock bool, locking, plain string) string {
	if lock {
		return locking
	}
	return plain
}

func containsGroup(gs []GroupID, g GroupID) bool {
	for _, c := range gs {
		if c == g {
			return true
		}
	}
	return false
}

func groupArgs(gs []GroupID) []pgtype.UUID {
	out := make([]pgtype.UUID, len(gs))
	for i, g := range gs {
		out[i] = uuidArg(g)
	}
	return out
}

func scanOrganizationBindingFact(row pgx.Row) (BindingFact, error) {
	var b BindingFact
	var org pgtype.UUID
	if err := row.Scan(&org, &b.Role); err != nil {
		return BindingFact{}, err
	}
	// The projection exposes membership bindings only (00003_tenancy.sql).
	b.OrganizationID, b.Source = OrganizationID(org.Bytes), SourceMembership
	return b, nil
}

func scanGroupBindingFact(row pgx.Row) (BindingFact, error) {
	var b BindingFact
	var org, group pgtype.UUID
	if err := row.Scan(&org, &group, &b.Role); err != nil {
		return BindingFact{}, err
	}
	b.OrganizationID, b.GroupID, b.Source = OrganizationID(org.Bytes), GroupID(group.Bytes), SourceMembership
	return b, nil
}

func scanAssociationFact(row pgx.Row) (AssociationFact, error) {
	var a AssociationFact
	var group, profile pgtype.UUID
	if err := row.Scan(&group, &profile); err != nil {
		return AssociationFact{}, err
	}
	a.GroupID, a.BusinessProfileID = GroupID(group.Bytes), BusinessProfileID(profile.Bytes)
	return a, nil
}

// --- transaction-scoped writes ------------------------------------------------------------------------------------

// RenameGroupTx is RenameGroup's validation and statement inside the caller's transaction: the display name only.
func RenameGroupTx(ctx context.Context, tx pgx.Tx, org OrganizationID, g GroupID, name string) (Group, error) {
	if zero(org) || zero(g) {
		return Group{}, newError(CodeInvalidID)
	}
	n, err := ValidateName(name)
	if err != nil {
		return Group{}, err
	}
	out, err := renameGroupIn(ctx, tx, org, g, n)
	if e := classify(err); e != nil {
		return Group{}, e
	}
	return out, nil
}

// ChangeGroupSlugTx is ChangeGroupSlug's validation and statement inside the caller's transaction; setting the current
// slug is a successful no-op and a collision is slug_conflict.
func ChangeGroupSlugTx(ctx context.Context, tx pgx.Tx, org OrganizationID, g GroupID, slug string) (Group, error) {
	if zero(org) || zero(g) {
		return Group{}, newError(CodeInvalidID)
	}
	if err := ValidateSlug(slug); err != nil {
		return Group{}, err
	}
	out, err := changeGroupSlugIn(ctx, tx, org, g, slug)
	if e := classify(err); e != nil {
		return Group{}, e
	}
	return out, nil
}

// UpdateBusinessProfileTx is UpdateBusinessProfile's validation and full-replace statement inside the caller's
// transaction.
func UpdateBusinessProfileTx(ctx context.Context, tx pgx.Tx, org OrganizationID, p BusinessProfileID, c BusinessContext) (BusinessProfile, error) {
	if zero(org) || zero(p) {
		return BusinessProfile{}, newError(CodeInvalidID)
	}
	v, err := ValidateBusinessContext(c)
	if err != nil {
		return BusinessProfile{}, err
	}
	out, err := updateBusinessProfileIn(ctx, tx, org, p, v)
	if e := classify(err); e != nil {
		return BusinessProfile{}, e
	}
	return out, nil
}

// SetOrganizationPolicyTx assigns the Organization's policy inside the caller's transaction (an absolute set: setting
// the current value is a successful no-op) and returns the stored policy.
func SetOrganizationPolicyTx(ctx context.Context, tx pgx.Tx, org OrganizationID, p OrganizationPolicy) (OrganizationPolicy, error) {
	if zero(org) {
		return OrganizationPolicy{}, newError(CodeInvalidID)
	}
	var out OrganizationPolicy
	err := tx.QueryRow(ctx, sqlSetOrganizationPolicy, uuidArg(org), p.GroupAdminsEditBusinessProfiles).Scan(&out.GroupAdminsEditBusinessProfiles)
	if e := classify(one(err, CodeOrganizationNotFound)); e != nil {
		return OrganizationPolicy{}, e
	}
	return out, nil
}

// SetGroupManagementOverrideTx assigns the Organization's Group-management override inside the caller's transaction (an
// absolute set: setting the current value is a successful no-op) and returns the stored value (G060). The database guard
// app.organizations_group_override_guard refuses a change by anyone but the current Organization owner
// (owner_authority); the capture trigger records organization.policy_changed atomically.
func SetGroupManagementOverrideTx(ctx context.Context, tx pgx.Tx, org OrganizationID, enabled bool) (bool, error) {
	if zero(org) {
		return false, newError(CodeInvalidID)
	}
	var out bool
	err := tx.QueryRow(ctx, sqlSetGroupManagementOverride, uuidArg(org), enabled).Scan(&out)
	if e := classify(one(err, CodeOrganizationNotFound)); e != nil {
		return false, e
	}
	return out, nil
}

// The statement bodies shared by the Service methods and the transaction-scoped writes.

func renameGroupIn(ctx context.Context, tx pgx.Tx, org OrganizationID, g GroupID, name string) (Group, error) {
	out, err := scanGroup(tx.QueryRow(ctx, sqlRenameGroup, uuidArg(org), uuidArg(g), name))
	return out, one(err, CodeGroupNotFound)
}

func changeGroupSlugIn(ctx context.Context, tx pgx.Tx, org OrganizationID, g GroupID, slug string) (Group, error) {
	out, err := scanGroup(tx.QueryRow(ctx, sqlChangeSlug, uuidArg(org), uuidArg(g), slug))
	return out, one(err, CodeGroupNotFound)
}

func updateBusinessProfileIn(ctx context.Context, tx pgx.Tx, org OrganizationID, p BusinessProfileID, v BusinessContext) (BusinessProfile, error) {
	out, err := scanBusinessProfile(tx.QueryRow(ctx, sqlUpdateBusinessProfile, uuidArg(org), uuidArg(p), v.DisplayName, v.LegalName,
		v.Description, v.Industry, v.Website))
	return out, one(err, CodeBusinessProfileNotFound)
}
