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

// Global account profiles (G074 Plan r3.1 §3.5, §3.9; migration 00023_account_profiles.sql). These are
// transaction-scoped model primitives for the guarded profile operations of core/internal/tenancy/access: they run in a
// transaction the caller owns (with the request context initialized first), enforce no capability and never log.
// Row-level security admits the caller's own profile and avatar rows for writes, and other users' profile rows only
// through the relationship predicate; every statement names its user ($1), and the guard and capture triggers own the
// version rules, the avatar-reference check and the audit row. Errors are returned unmapped (the caller maps constraint
// names, as for the branding statements); a missing or invisible row is ErrProfileAbsent. Avatar statements are the
// avatar facade's (Plan r3.1 §8 L2).

// ErrProfileAbsent reports that the user has no profile row (or, for VisibleProfileTx, that the caller may not read it:
// unknown, foreign and unrelated subjects are indistinguishable).
var ErrProfileAbsent = errors.New("tenancy: profile absent")

// ErrProfileNameInvalid reports a name outside the profile name contract.
var ErrProfileNameInvalid = errors.New("tenancy: invalid profile name")

// ProfileNameMax bounds an explicit profile name, in code points.
const ProfileNameMax = 200

// Profile is one user's profile as a statement returned it. ProviderName is the provider-seeded app.users.display_name
// (own reads only); Name the explicit override (nil: none) or, for VisibleProfileTx, the effective name; Version the
// optimistic row version (0: no row); AvatarVersion and AvatarType describe the current avatar while it is ready
// (AvatarType "" otherwise).
type Profile struct {
	UserID        UserID
	ProviderName  string
	Name          *string
	Version       int32
	AvatarVersion int32
	AvatarType    string
}

// The profile name statements. The guard requires version = OLD.version + 1 and sets updated_at; the capture records
// profile.updated (name_changed) for a changed name only. A first row is inserted with its name at version 1 (the empty
// first row is never written: clearing a name that was never set writes nothing).
const (
	sqlReadOwnProfile = `SELECT u.display_name, p.name, coalesce(p.version, 0), coalesce(p.avatar_version, 0), a.content_type ` +
		`FROM app.users u LEFT JOIN app.account_profiles p ON p.user_id = u.id ` +
		`LEFT JOIN app.account_avatars a ON a.id = p.avatar_object_id AND a.user_id = u.id AND a.state = 'ready' WHERE u.id = $1`
	sqlLockOwnProfile    = `SELECT name, version FROM app.account_profiles WHERE user_id = $1 FOR NO KEY UPDATE`
	sqlInsertProfileName = `INSERT INTO app.account_profiles (user_id, name) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING RETURNING version`
	sqlUpdateProfileName = `UPDATE app.account_profiles SET name = $3, version = version + 1 WHERE user_id = $1 AND version = $2 RETURNING version`
	// sqlVisibleProfile calls the policy reader's SECURITY DEFINER projection (00023): zero rows unless the relationship
	// predicate admits the subject in the caller's current context. It names no protected relation itself.
	sqlVisibleProfile = `SELECT user_id, name, avatar_version, avatar_type FROM app.rls_account_profile($1)`
)

// NormalizeProfileName applies the profile name contract (the G063 grammar with a 200 code point bound): valid UTF-8,
// no control, line-separator or paragraph-separator character, trimmed of Unicode white space. nil, or a value that is
// empty after trimming, is nil (the override is cleared and the provider name applies).
func NormalizeProfileName(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	if !utf8.ValidString(*raw) {
		return nil, ErrProfileNameInvalid
	}
	for _, r := range *raw {
		if unicode.IsControl(r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return nil, ErrProfileNameInvalid
		}
	}
	name := strings.TrimFunc(*raw, unicode.IsSpace)
	if name == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(name) > ProfileNameMax {
		return nil, ErrProfileNameInvalid
	}
	return &name, nil
}

// ReadOwnProfileTx reads the user's own profile with the provider name and the current ready avatar (Version 0 when
// the user has no profile row). ErrProfileAbsent means the user row itself is not readable in this context.
func ReadOwnProfileTx(ctx context.Context, tx pgx.Tx, user UserID) (Profile, error) {
	if zero(user) {
		return Profile{}, newError(CodeInvalidID)
	}
	p := Profile{UserID: user}
	var provider, avatarType *string
	if err := tx.QueryRow(ctx, sqlReadOwnProfile, uuidArg(user)).Scan(&provider, &p.Name, &p.Version, &p.AvatarVersion,
		&avatarType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, ErrProfileAbsent
		}
		return Profile{}, err
	}
	// app.users.display_name is nullable (a provider without a name): no provider name is "".
	if provider != nil {
		p.ProviderName = *provider
	}
	if avatarType != nil {
		p.AvatarType = *avatarType
	}
	return p, nil
}

// LockOwnProfileTx reads and locks the user's own profile row FOR NO KEY UPDATE (ErrProfileAbsent when there is none;
// an absent row locks nothing, and a concurrent first insert loses on the primary key).
func LockOwnProfileTx(ctx context.Context, tx pgx.Tx, user UserID) (Profile, error) {
	if zero(user) {
		return Profile{}, newError(CodeInvalidID)
	}
	p := Profile{UserID: user}
	if err := tx.QueryRow(ctx, sqlLockOwnProfile, uuidArg(user)).Scan(&p.Name, &p.Version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, ErrProfileAbsent
		}
		return Profile{}, err
	}
	return p, nil
}

// InsertProfileNameTx creates the user's profile row at version 1 with name and returns its version. A row that
// already exists (a concurrent first edit won the primary key) is ErrProfileAbsent: nothing was written.
func InsertProfileNameTx(ctx context.Context, tx pgx.Tx, user UserID, name string) (int32, error) {
	if zero(user) {
		return 0, newError(CodeInvalidID)
	}
	var v int32
	if err := tx.QueryRow(ctx, sqlInsertProfileName, uuidArg(user), name).Scan(&v); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrProfileAbsent
		}
		return 0, err
	}
	return v, nil
}

// UpdateProfileNameTx sets (or, with nil, clears) the explicit name when the row is still at version, advancing it by
// one, and returns the new version. A row at another version (or none) is ErrProfileAbsent.
func UpdateProfileNameTx(ctx context.Context, tx pgx.Tx, user UserID, version int32, name *string) (int32, error) {
	if zero(user) {
		return 0, newError(CodeInvalidID)
	}
	var v int32
	if err := tx.QueryRow(ctx, sqlUpdateProfileName, uuidArg(user), version, name).Scan(&v); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrProfileAbsent
		}
		return 0, err
	}
	return v, nil
}

// VisibleProfileTx reads subject's profile as the caller may see it now (the effective name in Name; the avatar fields
// only while its current avatar is ready). Every subject the caller may not read, unknown, foreign or unrelated, is
// ErrProfileAbsent.
func VisibleProfileTx(ctx context.Context, tx pgx.Tx, subject UserID) (Profile, error) {
	if zero(subject) {
		return Profile{}, newError(CodeInvalidID)
	}
	var id pgtype.UUID
	var name, avatarType *string
	var avatarVersion *int32
	if err := tx.QueryRow(ctx, sqlVisibleProfile, uuidArg(subject)).Scan(&id, &name, &avatarVersion, &avatarType); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, ErrProfileAbsent
		}
		return Profile{}, err
	}
	if !id.Valid || UserID(id.Bytes) != subject {
		return Profile{}, ErrProfileAbsent
	}
	// The effective name is NULL only for a user without an explicit or a provider name: "".
	if name == nil {
		name = new(string)
	}
	p := Profile{UserID: subject, Name: name}
	if avatarVersion != nil && avatarType != nil {
		p.AvatarVersion, p.AvatarType = *avatarVersion, *avatarType
	}
	return p, nil
}
