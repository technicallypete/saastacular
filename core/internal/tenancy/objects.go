package tenancy

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Stored-object metadata (G009; migration 00007_stored_objects.sql, docs/foundation/OBJECT-STORAGE.md). These are
// transaction-scoped model primitives for the guarded storage operations of core/internal/tenancy/access: they run in
// a transaction the caller owns (with the request context initialized first), carry no actor beyond the explicit
// creator argument, enforce no capability and never log. Row-level security admits only rows of the initialized
// context, and every statement names its Organization.

// ObjectID is the canonical UUID of a stored object, distinct from every other ID type. The zero value is never valid.
type ObjectID [16]byte

func (id ObjectID) String() string { return formatUUID(id) }

// ParseObjectID parses a stored-object ID; see ParseOrganizationID.
func ParseObjectID(s string) (ObjectID, error) { return parseUUID[ObjectID](s) }

// ObjectState is the lifecycle state of a stored object.
type ObjectState string

// The four lifecycle states (00007): pending (inserted, bytes not yet published), ready (published, readable),
// deleting (logically deleted, physical removal not yet confirmed) and deleted (tombstone).
const (
	ObjectPending  ObjectState = "pending"
	ObjectReady    ObjectState = "ready"
	ObjectDeleting ObjectState = "deleting"
	ObjectDeleted  ObjectState = "deleted"
)

// MaxObjectBytes is the hard cap of an object's size, also enforced by the database (stored_objects_size_valid).
const MaxObjectBytes int64 = 1073741824

// DefaultContentType is the content type of an upload that names none.
const DefaultContentType = "application/octet-stream"

// contentTypeRe is the database pattern of stored_objects_content_type_valid: a lowercase type/subtype token pair
// without parameters.
var contentTypeRe = regexp.MustCompile(`^[a-z0-9][a-z0-9!#$&^_.+-]{0,62}/[a-z0-9][a-z0-9!#$&^_.+-]{0,62}$`)

// ValidateContentType returns the stored form of a content type: lowercased, application/octet-stream when empty, at
// most 127 bytes and matching the database pattern (so no parameters, whitespace or control characters). Anything
// else is invalid_content_type.
func ValidateContentType(s string) (string, error) {
	if s == "" {
		return DefaultContentType, nil
	}
	v := strings.ToLower(s)
	if len(v) > 127 || !contentTypeRe.MatchString(v) {
		return "", newError(CodeInvalidContentType)
	}
	return v, nil
}

// StoredObject is one row of app.stored_objects. GroupID is zero for an Organization-scoped object; SHA256 is set only
// from ready on; DeletedBy is zero until logical deletion.
type StoredObject struct {
	ID             ObjectID
	OrganizationID OrganizationID
	GroupID        GroupID
	CreatorUserID  UserID
	StorageKey     string
	State          ObjectState
	ByteSize       int64
	ContentType    string
	SHA256         [32]byte
	HasSHA256      bool
	CreatedAt      time.Time
	ReadyAt        time.Time
	DeletingAt     time.Time
	DeletedAt      time.Time
	DeletedBy      UserID
}

// ObjectStorageKey is the provider key of an object, derived exactly as the database's generated column derives it.
func ObjectStorageKey(org OrganizationID, id ObjectID) string {
	return "objects/" + org.String() + "/" + id.String()
}

// The stored-object statements. Every statement names the Organization ($1); none writes the generated key, the
// timestamps or deleted_by (the database and the 00007 trigger set them).
const (
	objectColumns = `id, organization_id, group_id, creator_user_id, storage_key, state, byte_size, content_type, sha256, ` +
		`created_at, ready_at, deleting_at, deleted_at, deleted_by`
	sqlInsertObject = `INSERT INTO app.stored_objects (organization_id, id, group_id, creator_user_id, byte_size, content_type) ` +
		`VALUES ($1, $2, $3, $4, $5, $6) RETURNING ` + objectColumns
	sqlSelectObject       = `SELECT ` + objectColumns + ` FROM app.stored_objects WHERE organization_id = $1 AND id = $2`
	sqlLockObject         = sqlSelectObject + ` FOR UPDATE`
	sqlFinalizeObject     = `UPDATE app.stored_objects SET state = 'ready', sha256 = $3 WHERE organization_id = $1 AND id = $2 AND state = 'pending' RETURNING ` + objectColumns
	sqlMarkObjectDeleting = `UPDATE app.stored_objects SET state = 'deleting' WHERE organization_id = $1 AND id = $2 AND state IN ('pending', 'ready') RETURNING ` + objectColumns
	sqlMarkObjectDeleted  = `UPDATE app.stored_objects SET state = 'deleted' WHERE organization_id = $1 AND id = $2 AND state = 'deleting' RETURNING ` + objectColumns
)

func scanObject(row pgx.Row) (StoredObject, error) {
	var o StoredObject
	var id, org, group, creator, deletedBy pgtype.UUID
	var state string
	var sha []byte
	var readyAt, deletingAt, deletedAt pgtype.Timestamptz
	if err := row.Scan(&id, &org, &group, &creator, &o.StorageKey, &state, &o.ByteSize, &o.ContentType, &sha, &o.CreatedAt,
		&readyAt, &deletingAt, &deletedAt, &deletedBy); err != nil {
		return StoredObject{}, err
	}
	o.ID, o.OrganizationID, o.CreatorUserID, o.State = ObjectID(id.Bytes), OrganizationID(org.Bytes), UserID(creator.Bytes), ObjectState(state)
	if group.Valid {
		o.GroupID = GroupID(group.Bytes)
	}
	if deletedBy.Valid {
		o.DeletedBy = UserID(deletedBy.Bytes)
	}
	if sha != nil {
		if len(sha) != 32 {
			return StoredObject{}, newError(CodeInternal)
		}
		copy(o.SHA256[:], sha)
		o.HasSHA256 = true
	}
	o.ReadyAt, o.DeletingAt, o.DeletedAt = readyAt.Time, deletingAt.Time, deletedAt.Time
	return o, nil
}

func groupOrNull(g GroupID) pgtype.UUID {
	if zero(g) {
		return pgtype.UUID{}
	}
	return uuidArg(g)
}

// InsertObjectTx inserts a pending object row inside the caller's transaction: id is generated by the caller, group
// zero means an Organization object, creator is the trusted subject. Row-level security admits the row only within the
// initialized context's upload scope (with the subject as creator); a refusal is unavailable or internal.
func InsertObjectTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id ObjectID, group GroupID, creator UserID, size int64,
	contentType string) (StoredObject, error) {
	if zero(org) || zero(id) || zero(creator) {
		return StoredObject{}, newError(CodeInvalidID)
	}
	if size < 0 || size > MaxObjectBytes {
		return StoredObject{}, newError(CodeObjectTooLarge)
	}
	ct, err := ValidateContentType(contentType)
	if err != nil {
		return StoredObject{}, err
	}
	out, err := scanObject(tx.QueryRow(ctx, sqlInsertObject, uuidArg(org), uuidArg(id), groupOrNull(group), uuidArg(creator), size, ct))
	if e := classify(err); e != nil {
		return StoredObject{}, e
	}
	return out, nil
}

// ReadObjectTx reads an object of the Organization (a plain read: row-level security admits it only within the
// context's read scope). A missing or invisible row is object_not_found.
func ReadObjectTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id ObjectID) (StoredObject, error) {
	return objectRow(ctx, tx, sqlSelectObject, org, id)
}

// LockObjectTx reads and locks an object row FOR UPDATE (row-level security also requires the context's lifecycle
// scope). A missing, invisible or unlockable row is object_not_found.
func LockObjectTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id ObjectID) (StoredObject, error) {
	return objectRow(ctx, tx, sqlLockObject, org, id)
}

func objectRow(ctx context.Context, tx pgx.Tx, sql string, org OrganizationID, id ObjectID) (StoredObject, error) {
	if zero(org) || zero(id) {
		return StoredObject{}, newError(CodeInvalidID)
	}
	out, err := scanObject(tx.QueryRow(ctx, sql, uuidArg(org), uuidArg(id)))
	if e := classify(one(err, CodeObjectNotFound)); e != nil {
		return StoredObject{}, e
	}
	return out, nil
}

// FinalizeObjectTx publishes a pending object with its digest (pending -> ready; the trigger admits only the creator).
// A row that is not pending (or not updatable in the context) is object_not_found.
func FinalizeObjectTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id ObjectID, sha [32]byte) (StoredObject, error) {
	return objectUpdate(ctx, tx, sqlFinalizeObject, org, id, sha[:])
}

// MarkObjectDeletingTx logically deletes a pending or ready object (the trigger records the actor and the time). A row
// in another state is object_not_found.
func MarkObjectDeletingTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id ObjectID) (StoredObject, error) {
	return objectUpdate(ctx, tx, sqlMarkObjectDeleting, org, id)
}

// MarkObjectDeletedTx tombstones a deleting object after its provider bytes are confirmed absent. A row in another
// state is object_not_found.
func MarkObjectDeletedTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id ObjectID) (StoredObject, error) {
	return objectUpdate(ctx, tx, sqlMarkObjectDeleted, org, id)
}

func objectUpdate(ctx context.Context, tx pgx.Tx, sql string, org OrganizationID, id ObjectID, extra ...any) (StoredObject, error) {
	if zero(org) || zero(id) {
		return StoredObject{}, newError(CodeInvalidID)
	}
	args := append([]any{uuidArg(org), uuidArg(id)}, extra...)
	out, err := scanObject(tx.QueryRow(ctx, sql, args...))
	if e := classify(one(err, CodeObjectNotFound)); e != nil {
		return StoredObject{}, e
	}
	return out, nil
}
