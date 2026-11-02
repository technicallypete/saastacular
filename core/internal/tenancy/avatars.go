package tenancy

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Account avatar objects (G074 Plan r3.1 §3.1, §3.8; migration 00023_account_profiles.sql). These are
// transaction-scoped model primitives for the guarded avatar operations of core/internal/tenancy/access
// (profile_avatar.go): they run in a transaction the caller owns (with the request context initialized first), enforce
// no capability and never log. Every statement names its user ($1); row-level security admits only the caller's own
// avatar rows and own profile row for writes, and the guard triggers own the lifecycle transitions, the timestamps, the
// avatar-reference check, the version rules and I-CUR (a current avatar never becomes deleting). The reconcile
// selection and the allocation gate use the null-safe IS DISTINCT FROM against the current reference, so a profile
// without a current avatar (NULL, before the first avatar or after removal) counts every unresolved row. Errors are
// returned unmapped; a missing, invisible or no longer matching avatar row is ErrAvatarAbsent, a missing profile row (or
// a stale version) ErrProfileAbsent.

// AvatarID is the canonical UUID of an account avatar object, distinct from every other ID type. The zero value is
// never valid.
type AvatarID [16]byte

func (id AvatarID) String() string { return formatUUID(id) }

// ErrAvatarAbsent reports that the avatar row is missing, invisible, or not in the state the statement requires.
var ErrAvatarAbsent = errors.New("tenancy: avatar absent")

// MaxAvatarSweep bounds one reconcile selection (Plan r3.1 §3.8 S0b): at most 20 rows, oldest first.
const MaxAvatarSweep = 20

// Avatar is one row of app.account_avatars. State reuses the four object lifecycle states; SHA256 is set from ready on;
// ReadyAt is zero for a row that never became ready (the guard preserves it across withdrawal). UploadStarted is the
// write-ahead marker of the row's one provider Put (upload_started_at): true from before the Put was issued, so a row
// that never became ready but is marked has a provider outcome that is unproven until its bytes are observed and deleted.
type Avatar struct {
	ID            AvatarID
	UserID        UserID
	StorageKey    string
	State         ObjectState
	ByteSize      int64
	ContentType   string
	SHA256        [32]byte
	HasSHA256     bool
	CreatedAt     time.Time
	ReadyAt       time.Time
	UploadStarted bool
}

// AvatarProfile is the caller's own profile row as the avatar operations see it: the optimistic version, the current
// avatar reference (zero: none), the avatar version and whether no explicit name is set.
type AvatarProfile struct {
	Version       int32
	AvatarID      AvatarID
	AvatarVersion int32
	NoName        bool
}

// Empty reports a first row that never changed: version 1, no name, no avatar ever set (the row an allocation creates
// to serialize on). Its content equals having no row at all.
func (p AvatarProfile) Empty() bool {
	return p.Version == 1 && p.NoName && zero(p.AvatarID) && p.AvatarVersion == 0
}

// AvatarGate is the allocation gate's count of the caller's unresolved rows: every row that is not the current avatar and
// not yet a tombstone (Outstanding), how many of them are pending rows younger than the lease (Young), and the lease
// remainder of the youngest pending row (zero when none).
type AvatarGate struct {
	Outstanding int64
	Young       int64
	RetryAfter  time.Duration
}

// VisibleAvatar is the current ready avatar of a subject the caller may read now: its version, row, content type, size
// and digest.
type VisibleAvatar struct {
	AvatarVersion int32
	ID            AvatarID
	ContentType   string
	ByteSize      int64
	SHA256        [32]byte
}

// AvatarStorageKey is the provider key of an avatar, derived exactly as the database's generated column derives it.
func AvatarStorageKey(user UserID, id AvatarID) string {
	return "avatars/" + user.String() + "/" + id.String()
}

// The avatar statements. None writes the generated key, created_at or a lifecycle timestamp (the database and the guard
// set them); none deletes.
const (
	avatarColumns = `id, user_id, storage_key, state, byte_size, content_type, sha256, created_at, ready_at, upload_started_at`
	// sqlInsertEmptyProfile creates the caller's empty first profile row (version 1, no name, no avatar) when none exists,
	// so that every avatar allocation serializes on the profile row lock; the capture records nothing for it.
	sqlInsertEmptyProfile = `INSERT INTO app.account_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`
	sqlReadAvatarProfile  = `SELECT version, avatar_object_id, avatar_version, name IS NULL FROM app.account_profiles WHERE user_id = $1`
	sqlLockAvatarProfile  = sqlReadAvatarProfile + ` FOR NO KEY UPDATE`
	// sqlSetAvatarReference moves (or, with NULL, clears) the current reference when the row is still at $2; the guard
	// requires the +1 advances and a ready avatar of the same user, and the capture records avatar_set or avatar_removed.
	sqlSetAvatarReference = `UPDATE app.account_profiles SET avatar_object_id = $3, avatar_version = avatar_version + 1, ` +
		`version = version + 1 WHERE user_id = $1 AND version = $2 RETURNING version, avatar_object_id, avatar_version, name IS NULL`
	sqlInsertAvatar = `INSERT INTO app.account_avatars (user_id, id, byte_size, content_type) VALUES ($1, $2, $3, $4) ` +
		`RETURNING ` + avatarColumns
	sqlSelectAvatar       = `SELECT ` + avatarColumns + ` FROM app.account_avatars WHERE user_id = $1 AND id = $2`
	sqlFinalizeAvatar     = `UPDATE app.account_avatars SET state = 'ready', sha256 = $3 WHERE user_id = $1 AND id = $2 AND state = 'pending' RETURNING ` + avatarColumns
	sqlMarkAvatarDeleting = `UPDATE app.account_avatars SET state = 'deleting' WHERE user_id = $1 AND id = $2 AND state IN ('pending', 'ready') RETURNING ` + avatarColumns
	sqlMarkAvatarDeleted  = `UPDATE app.account_avatars SET state = 'deleted' WHERE user_id = $1 AND id = $2 AND state = 'deleting' RETURNING ` + avatarColumns
	// sqlMarkAvatarUploadStarted is the write-ahead marker: a still-pending, still-unmarked row gets its upload marker (the
	// guard stamps upload_started_at; the value written here is replaced). A row a sweep already moved on matches nothing.
	sqlMarkAvatarUploadStarted = `UPDATE app.account_avatars SET upload_started_at = now() WHERE user_id = $1 AND id = $2 AND state = 'pending' ` +
		`AND upload_started_at IS NULL RETURNING ` + avatarColumns
	// sqlMarkAvatarUploadRefused withdraws a marked pending row whose provider Put was definitively refused (or never
	// issued): no object exists for it, so the marker is cleared with the move to deleting and the row settles like any
	// unmarked withdrawn row (the guard allows this one clearing).
	sqlMarkAvatarUploadRefused = `UPDATE app.account_avatars SET state = 'deleting', upload_started_at = NULL WHERE user_id = $1 AND id = $2 ` +
		`AND state = 'pending' AND upload_started_at IS NOT NULL RETURNING ` + avatarColumns
	// sqlStaleAvatars is the S0b reconcile selection: the caller's non-current, non-tombstoned rows except pending rows
	// younger than the lease ($2 seconds), oldest first, at most 20.
	sqlStaleAvatars = `SELECT ` + avatarColumns + ` FROM app.account_avatars WHERE user_id = $1 AND state <> 'deleted' ` +
		`AND id IS DISTINCT FROM (SELECT p.avatar_object_id FROM app.account_profiles p WHERE p.user_id = $1) ` +
		`AND (state <> 'pending' OR created_at < now() - make_interval(secs => $2)) ORDER BY created_at, id LIMIT 20`
	// sqlAvatarGate is the S1 allocation gate: every non-current, non-tombstoned row of the caller, the young pending rows
	// among them and the lease remainder of the youngest pending row (whole seconds, rounded up).
	sqlAvatarGate = `SELECT count(*), count(*) FILTER (WHERE state = 'pending' AND created_at >= now() - make_interval(secs => $2)), ` +
		`coalesce(ceil(extract(epoch FROM max(created_at) FILTER (WHERE state = 'pending') + make_interval(secs => $2) - now())), 0)::bigint ` +
		`FROM app.account_avatars WHERE user_id = $1 AND state <> 'deleted' ` +
		`AND id IS DISTINCT FROM (SELECT p.avatar_object_id FROM app.account_profiles p WHERE p.user_id = $1)`
	// sqlVisibleAvatar calls the policy reader's SECURITY DEFINER projection (00023): zero rows unless the relationship
	// predicate admits the subject now; the avatar columns are NULL unless the current avatar is ready.
	sqlVisibleAvatar = `SELECT user_id, avatar_version, avatar_object_id, avatar_type, avatar_size, avatar_sha256 FROM app.rls_account_profile($1)`
)

func scanAvatar(row pgx.Row) (Avatar, error) {
	var a Avatar
	var id, user pgtype.UUID
	var state string
	var sha []byte
	var readyAt, uploadStarted pgtype.Timestamptz
	if err := row.Scan(&id, &user, &a.StorageKey, &state, &a.ByteSize, &a.ContentType, &sha, &a.CreatedAt, &readyAt, &uploadStarted); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Avatar{}, ErrAvatarAbsent
		}
		return Avatar{}, err
	}
	a.ID, a.UserID, a.State, a.ReadyAt = AvatarID(id.Bytes), UserID(user.Bytes), ObjectState(state), readyAt.Time
	a.UploadStarted = uploadStarted.Valid
	if sha != nil {
		if len(sha) != 32 {
			return Avatar{}, newError(CodeInternal)
		}
		copy(a.SHA256[:], sha)
		a.HasSHA256 = true
	}
	if a.StorageKey != AvatarStorageKey(a.UserID, a.ID) {
		return Avatar{}, newError(CodeInternal)
	}
	return a, nil
}

func scanAvatarProfile(row pgx.Row) (AvatarProfile, error) {
	var p AvatarProfile
	var ref pgtype.UUID
	if err := row.Scan(&p.Version, &ref, &p.AvatarVersion, &p.NoName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AvatarProfile{}, ErrProfileAbsent
		}
		return AvatarProfile{}, err
	}
	if ref.Valid {
		p.AvatarID = AvatarID(ref.Bytes)
	}
	return p, nil
}

// InsertEmptyProfileTx creates the user's empty first profile row when it has none (a no-op otherwise).
func InsertEmptyProfileTx(ctx context.Context, tx pgx.Tx, user UserID) error {
	if zero(user) {
		return newError(CodeInvalidID)
	}
	_, err := tx.Exec(ctx, sqlInsertEmptyProfile, uuidArg(user))
	return err
}

// ReadAvatarProfileTx reads the user's own profile row without a lock (ErrProfileAbsent when there is none).
func ReadAvatarProfileTx(ctx context.Context, tx pgx.Tx, user UserID) (AvatarProfile, error) {
	if zero(user) {
		return AvatarProfile{}, newError(CodeInvalidID)
	}
	return scanAvatarProfile(tx.QueryRow(ctx, sqlReadAvatarProfile, uuidArg(user)))
}

// LockAvatarProfileTx reads and locks the user's own profile row FOR NO KEY UPDATE (ErrProfileAbsent when there is
// none; an absent row locks nothing).
func LockAvatarProfileTx(ctx context.Context, tx pgx.Tx, user UserID) (AvatarProfile, error) {
	if zero(user) {
		return AvatarProfile{}, newError(CodeInvalidID)
	}
	return scanAvatarProfile(tx.QueryRow(ctx, sqlLockAvatarProfile, uuidArg(user)))
}

// SetAvatarReferenceTx sets the current avatar reference to avatar (nil clears it) when the row is still at version,
// and returns the row after the change. A row at another version (or none) is ErrProfileAbsent.
func SetAvatarReferenceTx(ctx context.Context, tx pgx.Tx, user UserID, version int32, avatar *AvatarID) (AvatarProfile, error) {
	if zero(user) || (avatar != nil && zero(*avatar)) {
		return AvatarProfile{}, newError(CodeInvalidID)
	}
	ref := pgtype.UUID{}
	if avatar != nil {
		ref = uuidArg(*avatar)
	}
	return scanAvatarProfile(tx.QueryRow(ctx, sqlSetAvatarReference, uuidArg(user), version, ref))
}

// InsertAvatarTx inserts a pending avatar row of the user (id generated by the caller, so that a lost reply is
// reconcilable).
func InsertAvatarTx(ctx context.Context, tx pgx.Tx, user UserID, id AvatarID, size int64, contentType string) (Avatar, error) {
	if zero(user) || zero(id) {
		return Avatar{}, newError(CodeInvalidID)
	}
	return scanAvatar(tx.QueryRow(ctx, sqlInsertAvatar, uuidArg(user), uuidArg(id), size, contentType))
}

// ReadAvatarTx reads one avatar row of the user.
func ReadAvatarTx(ctx context.Context, tx pgx.Tx, user UserID, id AvatarID) (Avatar, error) {
	return avatarRow(ctx, tx, sqlSelectAvatar, user, id)
}

// FinalizeAvatarTx publishes a pending avatar with its digest (pending -> ready). A row that is not pending is
// ErrAvatarAbsent.
func FinalizeAvatarTx(ctx context.Context, tx pgx.Tx, user UserID, id AvatarID, sha [32]byte) (Avatar, error) {
	return avatarRow(ctx, tx, sqlFinalizeAvatar, user, id, sha[:])
}

// MarkAvatarUploadStartedTx records, durably and before the provider Put is issued, that the pending avatar's one Put may
// be in flight (the write-ahead marker). A row that is no longer pending or is already marked (a sweep moved it on) is
// ErrAvatarAbsent: the caller must then make no provider call.
func MarkAvatarUploadStartedTx(ctx context.Context, tx pgx.Tx, user UserID, id AvatarID) (Avatar, error) {
	return avatarRow(ctx, tx, sqlMarkAvatarUploadStarted, user, id)
}

// MarkAvatarUploadRefusedTx withdraws a marked pending avatar whose Put the provider definitively refused (or that was
// never issued): the row moves to deleting with its marker cleared, because no object can ever exist for it. A row in any
// other state, or never marked, is ErrAvatarAbsent.
func MarkAvatarUploadRefusedTx(ctx context.Context, tx pgx.Tx, user UserID, id AvatarID) (Avatar, error) {
	return avatarRow(ctx, tx, sqlMarkAvatarUploadRefused, user, id)
}

// MarkAvatarDeletingTx withdraws a pending or ready avatar (-> deleting; the guard refuses the current avatar). A row in
// another state is ErrAvatarAbsent.
func MarkAvatarDeletingTx(ctx context.Context, tx pgx.Tx, user UserID, id AvatarID) (Avatar, error) {
	return avatarRow(ctx, tx, sqlMarkAvatarDeleting, user, id)
}

// MarkAvatarDeletedTx tombstones a deleting avatar after its provider bytes are confirmed absent. A row in another
// state is ErrAvatarAbsent.
func MarkAvatarDeletedTx(ctx context.Context, tx pgx.Tx, user UserID, id AvatarID) (Avatar, error) {
	return avatarRow(ctx, tx, sqlMarkAvatarDeleted, user, id)
}

func avatarRow(ctx context.Context, tx pgx.Tx, sql string, user UserID, id AvatarID, extra ...any) (Avatar, error) {
	if zero(user) || zero(id) {
		return Avatar{}, newError(CodeInvalidID)
	}
	args := append([]any{uuidArg(user), uuidArg(id)}, extra...)
	return scanAvatar(tx.QueryRow(ctx, sql, args...))
}

// StaleAvatarsTx is the reconcile selection: the user's non-current, non-tombstoned rows except pending rows younger
// than lease, oldest first, at most MaxAvatarSweep.
func StaleAvatarsTx(ctx context.Context, tx pgx.Tx, user UserID, lease time.Duration) ([]Avatar, error) {
	if zero(user) {
		return nil, newError(CodeInvalidID)
	}
	rows, err := tx.Query(ctx, sqlStaleAvatars, uuidArg(user), lease.Seconds())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Avatar
	for rows.Next() {
		a, err := scanAvatar(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// AvatarGateTx counts the user's unresolved rows for the allocation gate (pending rows younger than lease are young).
func AvatarGateTx(ctx context.Context, tx pgx.Tx, user UserID, lease time.Duration) (AvatarGate, error) {
	if zero(user) {
		return AvatarGate{}, newError(CodeInvalidID)
	}
	var g AvatarGate
	var secs int64
	if err := tx.QueryRow(ctx, sqlAvatarGate, uuidArg(user), lease.Seconds()).Scan(&g.Outstanding, &g.Young, &secs); err != nil {
		return AvatarGate{}, err
	}
	if secs > 0 {
		g.RetryAfter = time.Duration(secs) * time.Second
	}
	return g, nil
}

// VisibleAvatarTx reads subject's current ready avatar as the caller may see it now. A subject the caller may not read
// (unknown, foreign, unrelated) is ErrProfileAbsent; a readable subject without a current ready avatar is ErrAvatarAbsent.
func VisibleAvatarTx(ctx context.Context, tx pgx.Tx, subject UserID) (VisibleAvatar, error) {
	if zero(subject) {
		return VisibleAvatar{}, newError(CodeInvalidID)
	}
	var id, ref pgtype.UUID
	var version *int32
	var ct *string
	var size *int64
	var sha []byte
	if err := tx.QueryRow(ctx, sqlVisibleAvatar, uuidArg(subject)).Scan(&id, &version, &ref, &ct, &size, &sha); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return VisibleAvatar{}, ErrProfileAbsent
		}
		return VisibleAvatar{}, err
	}
	if !id.Valid || UserID(id.Bytes) != subject {
		return VisibleAvatar{}, ErrProfileAbsent
	}
	if !ref.Valid || version == nil || ct == nil || size == nil || len(sha) != 32 {
		return VisibleAvatar{}, ErrAvatarAbsent
	}
	v := VisibleAvatar{AvatarVersion: *version, ID: AvatarID(ref.Bytes), ContentType: *ct, ByteSize: *size}
	copy(v.SHA256[:], sha)
	return v, nil
}
