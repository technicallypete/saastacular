package tenancy

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// BusinessProfileID and AssociationID are canonical UUIDs (G006), distinct from every other ID type. The zero value is
// never a valid ID.
type (
	BusinessProfileID [16]byte
	AssociationID     [16]byte
)

func (id BusinessProfileID) String() string { return formatUUID(id) }
func (id AssociationID) String() string     { return formatUUID(id) }

// ParseBusinessProfileID parses a BusinessProfile ID; see ParseOrganizationID.
func ParseBusinessProfileID(s string) (BusinessProfileID, error) {
	return parseUUID[BusinessProfileID](s)
}

// ParseAssociationID parses a Group-BusinessProfile association ID; see ParseOrganizationID.
func ParseAssociationID(s string) (AssociationID, error) { return parseUUID[AssociationID](s) }

// BusinessProfile is the reusable business context of one Organization (migration 00004_business_profiles.sql). It
// belongs to its Organization for its whole life; its UUID is its identity and its context is editable. It carries no
// role, owner or grant: access to it is derived from Group membership bindings (BusinessProfileAccessPath).
type BusinessProfile struct {
	ID             BusinessProfileID
	OrganizationID OrganizationID
	Context        BusinessContext
	CreatedAt      time.Time
}

// CreateBusinessProfile creates a BusinessProfile in an Organization with a database-generated UUID. It links no Group,
// grants nothing and consults no provider; equal names never merge.
func (s *Service) CreateBusinessProfile(ctx context.Context, org OrganizationID, c BusinessContext) (BusinessProfile, error) {
	var out BusinessProfile
	err := s.mutate(ctx, opBusinessProfileCreate, "", func() error {
		if zero(org) {
			return newError(CodeInvalidID)
		}
		v, err := ValidateBusinessContext(c)
		if err != nil {
			return err
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
				return err
			}
			out, err = scanBusinessProfile(tx.QueryRow(ctx, sqlInsertBusinessProfile, uuidArg(org), v.DisplayName, v.LegalName,
				v.Description, v.Industry, v.Website))
			return err
		})
	})
	if err != nil {
		return BusinessProfile{}, err
	}
	return out, nil
}

// GetBusinessProfile reads a BusinessProfile of an Organization; an unknown and a foreign profile give the same
// business_profile_not_found.
func (s *Service) GetBusinessProfile(ctx context.Context, org OrganizationID, p BusinessProfileID) (BusinessProfile, error) {
	if zero(org) || zero(p) {
		return BusinessProfile{}, newError(CodeInvalidID)
	}
	var out BusinessProfile
	err := s.query(ctx, func(tx pgx.Tx) error {
		var err error
		out, err = scanBusinessProfile(tx.QueryRow(ctx, sqlSelectBusinessProfile, uuidArg(org), uuidArg(p)))
		return one(err, CodeBusinessProfileNotFound)
	})
	if err != nil {
		return BusinessProfile{}, err
	}
	return out, nil
}

// UpdateBusinessProfile replaces the whole context of a BusinessProfile in one statement (an omitted optional field is
// cleared); the UUID, the Organization, creation time and every association and binding are unchanged. Concurrent
// updates resolve as last committed wins; every Group associated with the profile reads the same updated context.
func (s *Service) UpdateBusinessProfile(ctx context.Context, org OrganizationID, p BusinessProfileID, c BusinessContext) (BusinessProfile, error) {
	var out BusinessProfile
	err := s.mutate(ctx, opBusinessProfileUpdate, "", func() error {
		if zero(org) || zero(p) {
			return newError(CodeInvalidID)
		}
		v, err := ValidateBusinessContext(c)
		if err != nil {
			return err
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
				return err
			}
			out, err = updateBusinessProfileIn(ctx, tx, org, p, v)
			return err
		})
	})
	if err != nil {
		return BusinessProfile{}, err
	}
	return out, nil
}
