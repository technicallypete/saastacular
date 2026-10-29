package tenancy

import (
	"bytes"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// G018 onboarding primitives (Plan r5 §4.7 and §5.1): the transaction-scoped writes of guarded Group creation and
// onboarding completion. Like the other *Tx functions they run in the caller's transaction, carry no actor and
// authorize nothing; core/internal/tenancy/access is their only caller. Each executes exactly the statement sequence
// that migration 00014's grants and INSERT policies admit for the runtime role (TestOnboardingCatalog CT-ONB-1).

// keyCollisions are the primary keys of the keyed inserts: a collision of a caller-generated UUID v4 is a store
// failure (unavailable), never a slug or association conflict.
var keyCollisions = map[string]bool{"groups_pkey": true, "business_profiles_pkey": true}

func classifyKeyed(err error) *Error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" && keyCollisions[pg.ConstraintName] {
		return newError(CodeUnavailable)
	}
	return classify(err)
}

// CreateGroupTx validates and creates Group id in Organization org inside the caller's transaction (S-G1/S-G2): one
// keyed INSERT without RETURNING, whose AFTER INSERT trigger (00014 groups_bootstrap_owner) writes the context user's
// group_owner binding in the same statement, then the read of the new row, which the creator can now see. An empty
// slug is replaced by the deterministic suggestion base of the name (no suffix search, as CreateGroup). A taken slug is
// slug_conflict; a row the read cannot see is unavailable.
func CreateGroupTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id GroupID, name, slug string) (Group, error) {
	if zero(org) || zero(id) {
		return Group{}, newError(CodeInvalidID)
	}
	n, err := ValidateName(name)
	if err != nil {
		return Group{}, err
	}
	if slug == "" {
		if slug = suggestBase(n); slug == "" {
			return Group{}, newError(CodeSlugRequired)
		}
	}
	if err := ValidateSlug(slug); err != nil {
		return Group{}, err
	}
	if _, err := tx.Exec(ctx, sqlInsertGroupKeyed, uuidArg(id), uuidArg(org), n, slug); err != nil {
		return Group{}, classifyKeyed(err)
	}
	g, err := scanGroup(tx.QueryRow(ctx, sqlSelectGroup, uuidArg(org), uuidArg(id)))
	switch {
	case isNoRows(err):
		return Group{}, newError(CodeUnavailable)
	case err != nil:
		return Group{}, classify(err)
	}
	return g, nil
}

// CreateBusinessProfileTx validates c and creates profile id in Organization org inside the caller's transaction (S-P1):
// one keyed INSERT without RETURNING (a Group-derived creator cannot read an unlinked profile). The insertion provenance
// columns are assigned by their defaults. It returns the stored form of c.
func CreateBusinessProfileTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id BusinessProfileID, c BusinessContext) (BusinessContext, error) {
	if zero(org) || zero(id) {
		return BusinessContext{}, newError(CodeInvalidID)
	}
	v, err := ValidateBusinessContext(c)
	if err != nil {
		return BusinessContext{}, err
	}
	if _, err := tx.Exec(ctx, sqlInsertBusinessProfileKeyed, uuidArg(id), uuidArg(org), v.DisplayName, v.LegalName,
		v.Description, v.Industry, v.Website); err != nil {
		return BusinessContext{}, classifyKeyed(err)
	}
	return v, nil
}

// LinkGroupBusinessProfileTx associates profile p with Group g inside the caller's transaction (S-P2, the association
// statement verbatim). A Group that already has a profile is association_exists (the unique key is the backstop of the
// caller's locked check).
func LinkGroupBusinessProfileTx(ctx context.Context, tx pgx.Tx, org OrganizationID, g GroupID, p BusinessProfileID) (GroupBusinessProfile, error) {
	if zero(org) || zero(g) || zero(p) {
		return GroupBusinessProfile{}, newError(CodeInvalidID)
	}
	out, err := scanAssociation(tx.QueryRow(ctx, sqlInsertAssociation, uuidArg(org), uuidArg(g), uuidArg(p)))
	if e := classify(err); e != nil {
		return GroupBusinessProfile{}, e
	}
	return out, nil
}

// MaxOnboardingPage bounds one onboarding list page (Plan r5 §1.5: limit 1..100).
const MaxOnboardingPage = 100

// GroupsPage reads at most limit+1 Groups of org ordered by id after after (zero: from the start), so the caller can
// tell whether another page follows, then the associations of those Groups in a second statement (Plan r5 §1.5: one
// page and one associations statement, never one per Group). A non-empty slug instead reads the one Group with that
// exact slug (sqlSelectBySlug), returned only when it sorts after the cursor. Row-level security restricts both
// statements to the rows the transaction's context may read; the caller decides every row on its facts.
func GroupsPage(ctx context.Context, tx pgx.Tx, org OrganizationID, after GroupID, limit int, slug string) ([]Group, []AssociationFact, error) {
	if zero(org) || limit < 1 || limit > MaxOnboardingPage {
		return nil, nil, newError(CodeInvalidID)
	}
	var groups []Group
	if slug != "" {
		if err := ValidateSlug(slug); err != nil {
			return nil, nil, err
		}
		g, err := scanGroup(tx.QueryRow(ctx, sqlSelectBySlug, uuidArg(org), slug))
		switch {
		case err == nil:
			if bytes.Compare(g.ID[:], after[:]) > 0 {
				groups = []Group{g}
			}
		case !isNoRows(err):
			return nil, nil, classify(err)
		}
	} else {
		rows, err := tx.Query(ctx, sqlFactGroupsPage, uuidArg(org), afterArg(after), limit+1)
		if groups, err = collect(rows, err, scanGroup); err != nil {
			return nil, nil, classify(err)
		}
	}
	if len(groups) == 0 {
		return groups, nil, nil
	}
	ids := make([]GroupID, len(groups))
	for i, g := range groups {
		ids[i] = g.ID
	}
	rows, err := tx.Query(ctx, sqlFactAssociations, uuidArg(org), groupArgs(ids))
	assoc, err := collect(rows, err, scanAssociationFact)
	if err != nil {
		return nil, nil, classify(err)
	}
	return groups, assoc, nil
}

// ProfilesPage reads at most limit+1 Business Profiles of org ordered by id after after (zero: from the start), in one
// statement restricted by row-level security to the profiles the context may read. The caller decides each row
// (business_profile.select_for_onboarding) on its facts; the next page's cursor is the last scanned id.
func ProfilesPage(ctx context.Context, tx pgx.Tx, org OrganizationID, after BusinessProfileID, limit int) ([]BusinessProfile, error) {
	if zero(org) || limit < 1 || limit > MaxOnboardingPage {
		return nil, newError(CodeInvalidID)
	}
	rows, err := tx.Query(ctx, sqlFactProfilesPage, uuidArg(org), uuidArg(after), limit+1)
	out, err := collect(rows, err, scanBusinessProfile)
	if err != nil {
		return nil, classify(err)
	}
	return out, nil
}

// SuggestGroupSlugTx is the transaction-scoped form of SuggestGroupSlug for the guarded facade (Plan r5 §1.3): the
// name's base if it is neither reserved nor used by a Group the transaction's context can read, else the first free
// base-2 ... base-99. It is advisory and reserves nothing: row-level security hides the slugs of Groups the context
// cannot read, so creation's unique key still decides (conflict, field slug). A name without a base gives
// slug_required; an exhausted search gives slug_unavailable.
func SuggestGroupSlugTx(ctx context.Context, tx pgx.Tx, org OrganizationID, name string) (string, error) {
	if zero(org) {
		return "", newError(CodeInvalidID)
	}
	n, err := ValidateName(name)
	if err != nil {
		return "", err
	}
	base := suggestBase(n)
	if base == "" {
		return "", newError(CodeSlugRequired)
	}
	candidates := slugCandidates(base)
	rows, err := tx.Query(ctx, sqlSlugsInUse, uuidArg(org), candidates)
	used, err := collect(rows, err, func(r pgx.Row) (string, error) {
		var v string
		return v, r.Scan(&v)
	})
	if err != nil {
		return "", classify(err)
	}
	taken := map[string]bool{}
	for _, v := range used {
		taken[v] = true
	}
	for _, c := range candidates {
		if !reservedSlugs[c] && !taken[c] {
			return c, nil
		}
	}
	return "", newError(CodeSlugUnavailable)
}
