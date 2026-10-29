package tenancy

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// DB begins transactions; *pgxpool.Pool satisfies it. Callers supply a pool logged in as the runtime role. The
// package reads no environment, credential or configuration.
type DB interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
}

// Service holds the model primitives. Every operation is one transaction at READ COMMITTED (read-write for mutations,
// read-only for reads) and returns success only after Commit returns nil; any error rolls back. There are no retries:
// a transient failure returns unavailable and the caller may retry. Concurrent role changes on one binding are
// serialized by its row lock and the last committed change wins; there is no optimistic-concurrency claim.
type Service struct {
	db     DB
	logger *slog.Logger
}

// NewService returns a Service over db. A nil logger discards the operation records.
func NewService(db DB, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Service{db: db, logger: logger}
}

func zero[T ~[16]byte](id T) bool { return id == T{} }

// inTx runs fn in one transaction. Begin and commit failures are unavailable; fn's errors are classified. The
// deferred rollback (a no-op after a successful commit) also covers a panic in fn.
func (s *Service) inTx(ctx context.Context, mode pgx.TxAccessMode, fn func(pgx.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: mode})
	if err != nil {
		return unavailable(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if err := fn(tx); err != nil {
		return classify(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return unavailable(err)
	}
	return nil
}

// mutate runs one logged mutation: validation and the transaction inside fn, then exactly one operation record.
func (s *Service) mutate(ctx context.Context, op string, scope ScopeType, fn func() error) error {
	var e *Error
	if err := fn(); err != nil {
		e = classify(err)
	}
	s.logOperation(ctx, op, scope, e)
	if e != nil {
		return e
	}
	return nil
}

// query runs one unlogged read-only transaction.
func (s *Service) query(ctx context.Context, fn func(pgx.Tx) error) error {
	return s.inTx(ctx, pgx.ReadOnly, fn)
}

// --- Organizations ---------------------------------------------------------------------------------------------

// CreateOrganization (ownerless) lives in fixture.go and is fixture-only (G019 Plan r1 §1.4, C2); the only supported
// creation shape is CreateOrganizationWithOwner (membership.go): the tenant and its org_owner binding in one transaction.

// GetOrganization reads one Organization by UUID.
func (s *Service) GetOrganization(ctx context.Context, org OrganizationID) (Organization, error) {
	if zero(org) {
		return Organization{}, newError(CodeInvalidID)
	}
	var out Organization
	err := s.query(ctx, func(tx pgx.Tx) error {
		var err error
		out, err = scanOrganization(tx.QueryRow(ctx, sqlSelectOrganization, uuidArg(org)))
		return one(err, CodeOrganizationNotFound)
	})
	if err != nil {
		return Organization{}, err
	}
	return out, nil
}

// RenameOrganization changes the display name only; the UUID, Groups and bindings are unchanged.
func (s *Service) RenameOrganization(ctx context.Context, org OrganizationID, name string) (Organization, error) {
	var out Organization
	err := s.mutate(ctx, opOrganizationRename, "", func() error {
		if zero(org) {
			return newError(CodeInvalidID)
		}
		n, err := ValidateName(name)
		if err != nil {
			return err
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			out, err = scanOrganization(tx.QueryRow(ctx, sqlRenameOrganization, uuidArg(org), n))
			return one(err, CodeOrganizationNotFound)
		})
	})
	if err != nil {
		return Organization{}, err
	}
	return out, nil
}

// --- Groups (every call names its owning Organization) ---------------------------------------------------------

// SuggestGroupSlug returns an advisory slug for a Group name in an Organization: the name's base if it is neither
// reserved nor used there, else the first free base-2 ... base-99. It reserves nothing; creation still decides
// collisions through the unique constraint. A name without a base gives slug_required; an exhausted search gives
// slug_unavailable.
func (s *Service) SuggestGroupSlug(ctx context.Context, org OrganizationID, name string) (string, error) {
	if zero(org) {
		return "", newError(CodeInvalidID)
	}
	n, err := ValidateName(name)
	if err != nil {
		return "", err
	}
	base := suggestBase(n)
	var out string
	err = s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
			return err
		}
		if base == "" {
			return newError(CodeSlugRequired)
		}
		candidates := slugCandidates(base)
		rows, err := tx.Query(ctx, sqlSlugsInUse, uuidArg(org), candidates)
		used, err := collect(rows, err, func(r pgx.Row) (string, error) {
			var v string
			return v, r.Scan(&v)
		})
		if err != nil {
			return err
		}
		taken := map[string]bool{}
		for _, v := range used {
			taken[v] = true
		}
		for _, c := range candidates {
			if !reservedSlugs[c] && !taken[c] {
				out = c
				return nil
			}
		}
		return newError(CodeSlugUnavailable)
	})
	if err != nil {
		return "", err
	}
	return out, nil
}

// CreateGroup creates a Group in an Organization and, in the same transaction, the explicit creator's group_owner
// membership binding. Either both rows commit or neither does. It mints no Organization membership and consults no
// provider claim; the creator must be an existing global user supplied by a trusted caller. An empty slug uses the
// name's base without a suffix search, so a collision is reported deterministically.
func (s *Service) CreateGroup(ctx context.Context, org OrganizationID, creator UserID, name, slug string) (Group, Binding, error) {
	var g Group
	var b Binding
	err := s.mutate(ctx, opGroupCreate, "", func() error {
		if zero(org) || zero(creator) {
			return newError(CodeInvalidID)
		}
		n, err := ValidateName(name)
		if err != nil {
			return err
		}
		if slug == "" {
			if slug = suggestBase(n); slug == "" {
				return newError(CodeSlugRequired)
			}
		}
		if err := ValidateSlug(slug); err != nil {
			return err
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
				return err
			}
			if err := exists(ctx, tx, CodeUserNotFound, sqlUserExists, uuidArg(creator)); err != nil {
				return err
			}
			if g, err = scanGroup(tx.QueryRow(ctx, sqlInsertGroup, uuidArg(org), n, slug)); err != nil {
				return err
			}
			b, err = scanBinding(tx.QueryRow(ctx, sqlInsertBinding, uuidArg(org), uuidArg(creator),
				string(ScopeGroup), uuidArg(g.ID), uuidArg(g.ID), string(GroupOwner)))
			return err
		})
	})
	if err != nil {
		return Group{}, Binding{}, err
	}
	return g, b, nil
}

// GetGroup reads a Group of an Organization; an unknown and a foreign Group give the same group_not_found.
func (s *Service) GetGroup(ctx context.Context, org OrganizationID, g GroupID) (Group, error) {
	if zero(org) || zero(g) {
		return Group{}, newError(CodeInvalidID)
	}
	var out Group
	err := s.query(ctx, func(tx pgx.Tx) error {
		var err error
		out, err = scanGroup(tx.QueryRow(ctx, sqlSelectGroup, uuidArg(org), uuidArg(g)))
		return one(err, CodeGroupNotFound)
	})
	if err != nil {
		return Group{}, err
	}
	return out, nil
}

// GetGroupBySlug reads a Group by its slug within an explicit Organization. A slug is an alternate identifier, never
// authority; the same slug in another Organization is a different Group.
func (s *Service) GetGroupBySlug(ctx context.Context, org OrganizationID, slug string) (Group, error) {
	if zero(org) {
		return Group{}, newError(CodeInvalidID)
	}
	if !canonicalSlug(slug) {
		return Group{}, newError(CodeInvalidSlug)
	}
	var out Group
	err := s.query(ctx, func(tx pgx.Tx) error {
		var err error
		out, err = scanGroup(tx.QueryRow(ctx, sqlSelectBySlug, uuidArg(org), slug))
		return one(err, CodeGroupNotFound)
	})
	if err != nil {
		return Group{}, err
	}
	return out, nil
}

// ListGroups lists an Organization's Groups ordered by slug.
func (s *Service) ListGroups(ctx context.Context, org OrganizationID) ([]Group, error) {
	if zero(org) {
		return nil, newError(CodeInvalidID)
	}
	var out []Group
	err := s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, sqlListGroups, uuidArg(org))
		out, err = collect(rows, err, scanGroup)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RenameGroup changes the display name only; the stored slug is never rewritten by a rename.
func (s *Service) RenameGroup(ctx context.Context, org OrganizationID, g GroupID, name string) (Group, error) {
	var out Group
	err := s.mutate(ctx, opGroupRename, "", func() error {
		if zero(org) || zero(g) {
			return newError(CodeInvalidID)
		}
		n, err := ValidateName(name)
		if err != nil {
			return err
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			out, err = renameGroupIn(ctx, tx, org, g, n)
			return err
		})
	})
	if err != nil {
		return Group{}, err
	}
	return out, nil
}

// ChangeGroupSlug changes only the slug to an explicit canonical value; setting the current value is a no-op success.
// A collision in the Organization is slug_conflict and changes nothing.
func (s *Service) ChangeGroupSlug(ctx context.Context, org OrganizationID, g GroupID, slug string) (Group, error) {
	var out Group
	err := s.mutate(ctx, opGroupChangeSlug, "", func() error {
		if zero(org) || zero(g) {
			return newError(CodeInvalidID)
		}
		if err := ValidateSlug(slug); err != nil {
			return err
		}
		return s.inTx(ctx, pgx.ReadWrite, func(tx pgx.Tx) error {
			var err error
			out, err = changeGroupSlugIn(ctx, tx, org, g, slug)
			return err
		})
	})
	if err != nil {
		return Group{}, err
	}
	return out, nil
}

// --- Memberships: reads go through the projections only ---------------------------------------------------------
//
// The unguarded Grant*/Change*/Remove* writers live in fixture.go and are fixture-only (G019 Plan r1 §1.4, C2).

// GetOrganizationMembership reads the subject's Organization membership from the projection.
func (s *Service) GetOrganizationMembership(ctx context.Context, org OrganizationID, u UserID) (OrganizationMembership, error) {
	if zero(org) || zero(u) {
		return OrganizationMembership{}, newError(CodeInvalidID)
	}
	var out OrganizationMembership
	err := s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
			return err
		}
		var err error
		out, err = scanOrgMembership(tx.QueryRow(ctx, sqlGetOrgMembership, uuidArg(org), uuidArg(u)))
		return one(err, CodeMembershipNotFound)
	})
	if err != nil {
		return OrganizationMembership{}, err
	}
	return out, nil
}

// ListOrganizationMemberships lists an Organization's memberships from the projection, ordered by user ID.
func (s *Service) ListOrganizationMemberships(ctx context.Context, org OrganizationID) ([]OrganizationMembership, error) {
	if zero(org) {
		return nil, newError(CodeInvalidID)
	}
	var out []OrganizationMembership
	err := s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeOrganizationNotFound, sqlOrganizationExists, uuidArg(org)); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, sqlListOrgMemberships, uuidArg(org))
		out, err = collect(rows, err, scanOrgMembership)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetGroupMembership reads the subject's membership of a Group of the Organization from the projection.
func (s *Service) GetGroupMembership(ctx context.Context, org OrganizationID, g GroupID, u UserID) (GroupMembership, error) {
	if zero(org) || zero(g) || zero(u) {
		return GroupMembership{}, newError(CodeInvalidID)
	}
	var out GroupMembership
	err := s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeGroupNotFound, sqlGroupExists, uuidArg(org), uuidArg(g)); err != nil {
			return err
		}
		var err error
		out, err = scanGroupMembership(tx.QueryRow(ctx, sqlGetGroupMembership, uuidArg(org), uuidArg(g), uuidArg(u)))
		return one(err, CodeMembershipNotFound)
	})
	if err != nil {
		return GroupMembership{}, err
	}
	return out, nil
}

// ListGroupMemberships lists a Group's memberships from the projection, ordered by user ID.
func (s *Service) ListGroupMemberships(ctx context.Context, org OrganizationID, g GroupID) ([]GroupMembership, error) {
	if zero(org) || zero(g) {
		return nil, newError(CodeInvalidID)
	}
	var out []GroupMembership
	err := s.query(ctx, func(tx pgx.Tx) error {
		if err := exists(ctx, tx, CodeGroupNotFound, sqlGroupExists, uuidArg(org), uuidArg(g)); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, sqlListGroupMemberships, uuidArg(org), uuidArg(g))
		out, err = collect(rows, err, scanGroupMembership)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
