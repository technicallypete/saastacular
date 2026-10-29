package tenancy

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// GroupBusinessProfile associates a Group with a BusinessProfile of the same Organization. A Group has at most one
// association; one profile may be associated with several Groups of its Organization.
type GroupBusinessProfile struct {
	ID                AssociationID
	OrganizationID    OrganizationID
	GroupID           GroupID
	BusinessProfileID BusinessProfileID
	CreatedAt         time.Time
}

// BusinessProfileAccessPath is a row of the read-only business_profile_access_paths projection: a current Group
// membership binding of a Group associated with the profile, with its raw Group role. It is a fact, not a capability;
// classifying roles into capabilities is G007's, enforcement G008's. A profile shared by several Groups appears once
// per Group path.
type BusinessProfileAccessPath struct {
	UserID            UserID
	OrganizationID    OrganizationID
	BusinessProfileID BusinessProfileID
	GroupID           GroupID
	GroupRole         GroupRole
}

// associationPrechecks runs the ordered, Organization-scoped prechecks of a Group and a profile.
func associationPrechecks(ctx context.Context, tx pgx.Tx, org OrganizationID, g GroupID, p BusinessProfileID) error {
	if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
		return err
	}
	if err := exists(ctx, tx, CodeGroupNotFound, sqlGroupExists, uuidArg(org), uuidArg(g)); err != nil {
		return err
	}
	return exists(ctx, tx, CodeBusinessProfileNotFound, sqlBusinessProfileExists, uuidArg(org), uuidArg(p))
}

// LinkGroupBusinessProfile associates a Group with a BusinessProfile of the same Organization. A Group that already
// has an association, with the same or another profile, gives association_exists and the existing association is
// unchanged; there is no replace operation. It writes no role binding.
func (s *Service) LinkGroupBusinessProfile(ctx context.Context, org OrganizationID, g GroupID, p BusinessProfileID) (GroupBusinessProfile, error) {
	var out GroupBusinessProfile
	err := s.mutate(ctx, opAssociationLink, "", func() error {
		if zero(org) || zero(g) || zero(p) {
			return newError(CodeInvalidID)
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			if err := associationPrechecks(ctx, tx, org, g, p); err != nil {
				return err
			}
			var linked bool
			if err := tx.QueryRow(ctx, sqlGroupHasAssociation, uuidArg(org), uuidArg(g)).Scan(&linked); err != nil {
				return err
			}
			if linked {
				return newError(CodeAssociationExists)
			}
			var err error
			out, err = scanAssociation(tx.QueryRow(ctx, sqlInsertAssociation, uuidArg(org), uuidArg(g), uuidArg(p)))
			return err
		})
	})
	if err != nil {
		return GroupBusinessProfile{}, err
	}
	return out, nil
}

// GetGroupBusinessProfile reads a Group's single association and its profile.
func (s *Service) GetGroupBusinessProfile(ctx context.Context, org OrganizationID, g GroupID) (GroupBusinessProfile, BusinessProfile, error) {
	if zero(org) || zero(g) {
		return GroupBusinessProfile{}, BusinessProfile{}, newError(CodeInvalidID)
	}
	var a GroupBusinessProfile
	var bp BusinessProfile
	err := s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeGroupNotFound, sqlGroupExists, uuidArg(org), uuidArg(g)); err != nil {
			return err
		}
		var err error
		a, bp, err = scanAssociationProfile(tx.QueryRow(ctx, sqlSelectGroupAssociation, uuidArg(org), uuidArg(g)))
		return one(err, CodeAssociationNotFound)
	})
	if err != nil {
		return GroupBusinessProfile{}, BusinessProfile{}, err
	}
	return a, bp, nil
}

// ListGroupsForBusinessProfile lists the associations of a profile, ordered by Group ID.
func (s *Service) ListGroupsForBusinessProfile(ctx context.Context, org OrganizationID, p BusinessProfileID) ([]GroupBusinessProfile, error) {
	if zero(org) || zero(p) {
		return nil, newError(CodeInvalidID)
	}
	var out []GroupBusinessProfile
	err := s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
			return err
		}
		if err := exists(ctx, tx, CodeBusinessProfileNotFound, sqlBusinessProfileExists, uuidArg(org), uuidArg(p)); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, sqlListProfileAssociations, uuidArg(org), uuidArg(p))
		out, err = collect(rows, err, scanAssociation)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UnlinkGroupBusinessProfile removes the association of a Group with the named profile. It is association_not_found
// when the Group is unassociated or associated with another profile (that association is unchanged). The profile, the
// Group, other associations and every binding are unchanged. A foundation primitive: completed-Group lifecycle rules
// (G018/G019) must guard any exposure of it.
//
// Since G007 its Group precheck locks the Group row FOR NO KEY UPDATE before the DELETE. The runtime role cannot
// row-lock associations, so the Group row is their lock surrogate: an unlink waits for every in-flight guarded write
// that relied on the association (which holds the Group row FOR SHARE), or commits before that write reads it
// (docs/foundation/AUTHORIZATION.md). Codes, their order and the result are unchanged.
func (s *Service) UnlinkGroupBusinessProfile(ctx context.Context, org OrganizationID, g GroupID, p BusinessProfileID) error {
	return s.mutate(ctx, opAssociationUnlink, "", func() error {
		if zero(org) || zero(g) || zero(p) {
			return newError(CodeInvalidID)
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
				return err
			}
			var locked pgtype.UUID
			if err := tx.QueryRow(ctx, sqlLockGroupForUnlink, uuidArg(org), uuidArg(g)).Scan(&locked); err != nil {
				return one(err, CodeGroupNotFound)
			}
			if err := exists(ctx, tx, CodeBusinessProfileNotFound, sqlBusinessProfileExists, uuidArg(org), uuidArg(p)); err != nil {
				return err
			}
			tag, err := tx.Exec(ctx, sqlDeleteAssociation, uuidArg(org), uuidArg(g), uuidArg(p))
			if err != nil {
				return err
			}
			if tag.RowsAffected() != 1 {
				return newError(CodeAssociationNotFound)
			}
			return nil
		})
	})
}

// ListBusinessProfileAccessPaths lists the derived access facts of a profile, ordered by user and Group.
func (s *Service) ListBusinessProfileAccessPaths(ctx context.Context, org OrganizationID, p BusinessProfileID) ([]BusinessProfileAccessPath, error) {
	if zero(org) || zero(p) {
		return nil, newError(CodeInvalidID)
	}
	var out []BusinessProfileAccessPath
	err := s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
			return err
		}
		if err := exists(ctx, tx, CodeBusinessProfileNotFound, sqlBusinessProfileExists, uuidArg(org), uuidArg(p)); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, sqlListProfileAccessPaths, uuidArg(org), uuidArg(p))
		out, err = collect(rows, err, scanAccessPath)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListUserBusinessProfileAccessPaths lists a user's derived access facts in an Organization, ordered by profile and
// Group. One global user may reach several profiles; an empty list is valid.
func (s *Service) ListUserBusinessProfileAccessPaths(ctx context.Context, org OrganizationID, u UserID) ([]BusinessProfileAccessPath, error) {
	if zero(org) || zero(u) {
		return nil, newError(CodeInvalidID)
	}
	var out []BusinessProfileAccessPath
	err := s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, sqlListUserAccessPaths, uuidArg(org), uuidArg(u))
		out, err = collect(rows, err, scanAccessPath)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
