package access

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/audit"
	"core/internal/tenancy"
)

// G074 guarded account profile operations (Plan r3.1 §3.4, §3.5, §3.9). A profile is global (one row per canonical user
// UUID) and only its user edits it; every operation is one guarded call in the session's Organization context: liveness,
// the Organization facts and the decision on account_profile.read or account_profile.manage_self (every Organization
// role), then the statements of core/internal/tenancy (profiles.go) under row-level security. Reads of another user's
// profile are decided by the database's relationship predicate (app.rls_account_profile) on every call: an unknown,
// foreign, unrelated or no-longer-related subject is the same not_found. The triggers own the version rules and the
// single profile.updated audit row of a changed name; a write that changes nothing writes nothing. The avatar operations
// are profile_avatar.go's (Plan r3.1 §8 L2).

// The profile operations' logged names (and access.denied operations, 00023). The avatar operations log
// profile.avatar.get, profile.avatar.set and profile.avatar.remove.
const (
	opProfileGet    = "profile.get"
	opProfileUpdate = "profile.update"
)

// fieldVersionConflict is the conflict field of a profile write whose expected version is stale.
const fieldVersionConflict = "version_conflict"

// AccountAvatar is a current, ready avatar's public reference: its version (the cache and URL key) and content type.
type AccountAvatar struct {
	Version     int32
	ContentType string
}

// AccountProfileView is the caller's own profile: Name is the explicit override (nil: none, the provider name
// applies), EffectiveName the presentation name, Avatar the current ready avatar (nil: none) and Version the optimistic
// row version (0: no profile row yet).
type AccountProfileView struct {
	UserID        tenancy.UserID
	Name          *string
	EffectiveName string
	Avatar        *AccountAvatar
	Version       int32
}

// AccountProfile is another user's profile as a related viewer sees it now: the effective name and the current ready
// avatar (nil: none). It never carries an email, provider identity, status, session, Organization list, object id or key.
type AccountProfile struct {
	UserID tenancy.UserID
	Name   string
	Avatar *AccountAvatar
}

func avatarRef(p tenancy.Profile) *AccountAvatar {
	if p.AvatarType == "" || p.AvatarVersion < 1 {
		return nil
	}
	return &AccountAvatar{Version: p.AvatarVersion, ContentType: p.AvatarType}
}

// ownView builds the caller's view from its own read.
func ownView(p tenancy.Profile) AccountProfileView {
	v := AccountProfileView{UserID: p.UserID, Name: p.Name, EffectiveName: p.ProviderName, Avatar: avatarRef(p), Version: p.Version}
	if p.Name != nil {
		v.EffectiveName = *p.Name
	}
	return v
}

// profileFailure maps an error of a profile statement: a lost version race or a guard refusal is the version conflict,
// the name CHECK invalid_input name; a store failure is unavailable with the transaction's reason.
func profileFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	var pg *pgconn.PgError
	switch {
	case errors.Is(err, tenancy.ErrProfileAbsent), errors.Is(err, pgx.ErrNoRows):
		return &failure{code: CodeConflict, field: fieldVersionConflict, reason: reasonWriteConflict}
	case errors.As(err, &pg) && pg.ConstraintName == "account_profiles_guard":
		return &failure{code: CodeConflict, field: fieldVersionConflict, reason: reasonWriteConflict}
	case errors.As(err, &pg) && pg.Code == "23514" && pg.ConstraintName == "account_profiles_name_valid":
		return invalid("name")
	case errors.As(err, &pg) && commandError(pg) && len(pg.Code) == 5 && pg.Code[:2] != "40" && pg.Code[:2] != "08" &&
		pg.Code[:2] != "57":
		return fail(CodeInternal, reasonStoreUnavailable)
	}
	return fail(CodeUnavailable, failureReason(ctx, tx.cause))
}

func sameName(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// readOwn reads the caller's own view inside the guarded transaction.
func readOwn(ctx context.Context, tx *trackedTx, user tenancy.UserID) (AccountProfileView, *failure) {
	p, err := tenancy.ReadOwnProfileTx(ctx, tx, user)
	if err != nil {
		if errors.Is(err, tenancy.ErrProfileAbsent) {
			// The live subject's own users row is always readable; its absence is an inconsistency, never not_found.
			return AccountProfileView{}, fail(CodeInternal, reasonFactInconsistent)
		}
		return AccountProfileView{}, profileFailure(ctx, tx, err)
	}
	return ownView(p), nil
}

// GetOwnAccountProfile returns the caller's own profile (account_profile.manage_self): Version 0 and the provider name
// when it has no profile row.
func (s *Service) GetOwnAccountProfile(ctx context.Context, tc *TenantContext) (AccountProfileView, error) {
	var out AccountProfileView
	err := s.guard(ctx, tc, opProfileGet, capAccountProfileManageSelf, false, Target{Kind: kindOrganization}, organizationQuery, nil,
		func(tx *trackedTx, _ Facts) *failure {
			var f *failure
			out, f = readOwn(ctx, tx, tc.subject)
			return f
		})
	if err != nil {
		return AccountProfileView{}, err
	}
	return out, nil
}

// UpdateAccountProfileName sets the caller's explicit presentation name (account_profile.manage_self) when its profile
// is still at expected (0: no row yet); nil, or a value empty after trimming, clears it, so the current provider name
// applies again (resolved P3). The name contract is checked before any transaction. Setting the current value is a
// successful no-op with no statement and no audit row, whatever expected says (an identical concurrent or retried edit
// is idempotent); a different value on a stale expected is conflict version_conflict with zero effects. The first edit
// inserts the row with its name at version 1; a concurrent first edit loses on the primary key and is compared again.
func (s *Service) UpdateAccountProfileName(ctx context.Context, tc *TenantContext, expected int32, name *string) (AccountProfileView, error) {
	var out AccountProfileView
	var want *string
	pre := func() *failure {
		if expected < 0 {
			return invalid("expected_version")
		}
		var err error
		if want, err = tenancy.NormalizeProfileName(name); err != nil {
			return invalid("name")
		}
		return nil
	}
	err := s.guard(ctx, tc, opProfileUpdate, capAccountProfileManageSelf, true, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, _ Facts) *failure {
			cur, err := tenancy.LockOwnProfileTx(ctx, tx, tc.subject)
			present := err == nil
			if err != nil && !errors.Is(err, tenancy.ErrProfileAbsent) {
				return profileFailure(ctx, tx, err)
			}
			switch {
			case sameName(cur.Name, want):
				// No statement: the value is already current.
			case cur.Version != expected:
				return &failure{code: CodeConflict, field: fieldVersionConflict, reason: reasonWriteConflict}
			case present:
				if _, err := tenancy.UpdateProfileNameTx(ctx, tx, tc.subject, cur.Version, want); err != nil {
					return profileFailure(ctx, tx, err)
				}
			default:
				// want is not nil here: an absent row's name is nil, and an equal value took the first branch.
				if _, err := tenancy.InsertProfileNameTx(ctx, tx, tc.subject, *want); err != nil {
					if !errors.Is(err, tenancy.ErrProfileAbsent) {
						return profileFailure(ctx, tx, err)
					}
					// A concurrent first edit committed its row: only an identical value is still a success.
					if cur, err = tenancy.LockOwnProfileTx(ctx, tx, tc.subject); err != nil {
						return profileFailure(ctx, tx, err)
					}
					if !sameName(cur.Name, want) {
						return &failure{code: CodeConflict, field: fieldVersionConflict, reason: reasonWriteConflict}
					}
				}
			}
			var f *failure
			out, f = readOwn(ctx, tx, tc.subject)
			return f
		})
	if err != nil {
		return AccountProfileView{}, err
	}
	return out, nil
}

// GetAccountProfile reads another user's (or the caller's own) profile as the caller may see it now
// (account_profile.read): the caller itself, the current owner/admin of the active Organization when the subject holds
// a binding in it, or a current member of a Group of the active Organization the subject is also a member of. Every
// other subject, unknown, foreign, unrelated or no longer related, is not_found with an identical outcome, recorded as a
// target_not_found denial without the subject.
func (s *Service) GetAccountProfile(ctx context.Context, tc *TenantContext, subject tenancy.UserID) (AccountProfile, error) {
	var out AccountProfile
	err := s.guard(ctx, tc, opProfileGet, capAccountProfileRead, false, Target{Kind: kindOrganization}, organizationQuery, targetID(subject),
		func(tx *trackedTx, f Facts) *failure {
			p, err := tenancy.VisibleProfileTx(ctx, tx, subject)
			if errors.Is(err, tenancy.ErrProfileAbsent) {
				return markDenied(notFound(), tc.selection(), f, "", audit.UUID{}, tenancy.GroupID{})
			}
			if err != nil {
				return profileFailure(ctx, tx, err)
			}
			out = AccountProfile{UserID: p.UserID, Avatar: avatarRef(p)}
			if p.Name != nil {
				out.Name = *p.Name
			}
			return nil
		})
	if err != nil {
		return AccountProfile{}, err
	}
	return out, nil
}
