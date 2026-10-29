package tenancy

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Organization branding (G017 Plan r2 §1.4, §4, §5.1; migration 00013_organization_branding.sql). These are
// transaction-scoped model primitives for the guarded branding operations of core/internal/tenancy/access: they run in
// a transaction the caller owns (with the request context initialized first), enforce no capability and never log.
// Row-level security admits only the initialized Organization's row to its owner or admin, every statement names its
// Organization ($1), and the guard and capture triggers own the version rule, the logo-reference check and the audit
// row. Errors are returned unmapped (the caller maps constraint names, as for the G016 domain statements); a missing
// row is ErrBrandingAbsent.

// ErrBrandingAbsent reports that the Organization has no branding row (the platform defaults apply).
var ErrBrandingAbsent = errors.New("tenancy: branding absent")

// Branding is one row of app.organization_branding. Label is nil for "no label"; LogoObjectID is zero for "no logo";
// Light and Dark are the stored overrides (token -> lowercase #rrggbb).
type Branding struct {
	OrganizationID OrganizationID
	Label          *string
	LogoObjectID   ObjectID
	Light, Dark    map[string]string
	Version        int64
}

// BrandingWrite is the full new content of a branding row; the version and the actor are explicit arguments.
type BrandingWrite struct {
	Label        *string
	LogoObjectID ObjectID
	Light, Dark  map[string]string
}

// The branding statements. Every statement names the Organization ($1); none writes organization_id after insert or
// derives the version from anything but the locked row (the guard requires exactly OLD.version + 1).
const (
	brandingColumns   = `organization_id, display_label, logo_object_id, light, dark, version`
	sqlSelectBranding = `SELECT ` + brandingColumns + ` FROM app.organization_branding WHERE organization_id = $1`
	sqlLockBranding   = sqlSelectBranding + ` FOR NO KEY UPDATE`
	sqlInsertBranding = `INSERT INTO app.organization_branding (organization_id, display_label, logo_object_id, light, dark, version, updated_by) ` +
		`VALUES ($1, $2, $3, $4, $5, 1, $6) RETURNING ` + brandingColumns
	sqlUpdateBranding = `UPDATE app.organization_branding SET display_label = $3, logo_object_id = $4, light = $5, dark = $6, ` +
		`version = version + 1, updated_at = now(), updated_by = $7 WHERE organization_id = $1 AND version = $2 RETURNING ` + brandingColumns
	// sqlNormalizeLabel is the accepted label contract's NFC normalization, done by PostgreSQL's built-in normalize()
	// (UTF-8 database; no Go normalization dependency). It reads no relation.
	sqlNormalizeLabel = `SELECT normalize($1::text, NFC)`
)

// The logo statements on app.stored_objects (Plan r2 §1.4.2, §5.1): the association's FOR SHARE lock of the new object
// (serialized against the G009 delete's FOR UPDATE lock: whichever commits first decides), and the bounded listing of
// the Organization's unreferenced logo uploads (Organization objects of a logo type, ready or deleting, other than the
// current reference; at most 100 rows with the total count). Row-level security admits only the context's rows.
const (
	logoObjectColumns = `id, organization_id, group_id, creator_user_id, storage_key, state, byte_size, content_type, sha256, ` +
		`created_at, ready_at, deleting_at, deleted_at, deleted_by`
	sqlShareLogoObject = `SELECT ` + logoObjectColumns + ` FROM app.stored_objects WHERE organization_id = $1 AND id = $2 FOR SHARE`
	sqlListLogoUploads = `SELECT id, content_type, byte_size, state, sha256, count(*) OVER () FROM app.stored_objects ` +
		`WHERE organization_id = $1 AND group_id IS NULL AND content_type IN ('image/png', 'image/jpeg') ` +
		`AND state IN ('ready', 'deleting') AND id IS DISTINCT FROM $2 ORDER BY created_at, id LIMIT 100`
	// sqlPublicBranding calls the auth reader's SECURITY DEFINER projection (00013): the intentionally public branding of
	// one Organization with its logo joined only while ready. It names no protected relation itself.
	sqlPublicBranding = `SELECT revision, display_label, light, dark, logo_sha256, logo_content_type, logo_size, logo_key ` +
		`FROM app.public_branding($1)`
)

// MaxLogoUploadsListed bounds the unreferenced-upload listing.
const MaxLogoUploadsListed = 100

// LogoUpload is one unreferenced logo upload of the Organization (an Organization object of a logo type that is not
// the current reference).
type LogoUpload struct {
	ID          ObjectID
	ContentType string
	ByteSize    int64
	State       ObjectState
	SHA256      [32]byte
	HasSHA256   bool
}

// PublicBranding is the public projection of one Organization's branding. Present is false when it has no row. The
// logo fields are set only while the referenced object is ready; LogoKey and LogoSHA256 stay inside core.
type PublicBranding struct {
	Present         bool
	Revision        int64
	Label           *string
	Light, Dark     map[string]string
	HasLogo         bool
	LogoSHA256      [32]byte
	LogoContentType string
	LogoSize        int64
	LogoKey         string
}

// Querier is a single-statement reader (a pool), for the anonymous public projection, which has no subject and so no
// initialized request context.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func scanBranding(row pgx.Row) (Branding, error) {
	var b Branding
	var org, logo pgtype.UUID
	if err := row.Scan(&org, &b.Label, &logo, &b.Light, &b.Dark, &b.Version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Branding{}, ErrBrandingAbsent
		}
		return Branding{}, err
	}
	b.OrganizationID = OrganizationID(org.Bytes)
	if logo.Valid {
		b.LogoObjectID = ObjectID(logo.Bytes)
	}
	if b.Light == nil {
		b.Light = map[string]string{}
	}
	if b.Dark == nil {
		b.Dark = map[string]string{}
	}
	return b, nil
}

func objectOrNull(id ObjectID) pgtype.UUID {
	if zero(id) {
		return pgtype.UUID{}
	}
	return uuidArg(id)
}

func overridesArg(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}

// ReadBrandingTx reads the Organization's branding row (ErrBrandingAbsent when there is none).
func ReadBrandingTx(ctx context.Context, tx pgx.Tx, org OrganizationID) (Branding, error) {
	if zero(org) {
		return Branding{}, newError(CodeInvalidID)
	}
	return scanBranding(tx.QueryRow(ctx, sqlSelectBranding, uuidArg(org)))
}

// LockBrandingTx reads and locks the Organization's branding row FOR NO KEY UPDATE (ErrBrandingAbsent when there is
// none; an absent row locks nothing, and a concurrent first insert loses on the primary key).
func LockBrandingTx(ctx context.Context, tx pgx.Tx, org OrganizationID) (Branding, error) {
	if zero(org) {
		return Branding{}, newError(CodeInvalidID)
	}
	return scanBranding(tx.QueryRow(ctx, sqlLockBranding, uuidArg(org)))
}

// InsertBrandingTx creates the Organization's branding row at version 1, attributed to actor.
func InsertBrandingTx(ctx context.Context, tx pgx.Tx, org OrganizationID, actor UserID, w BrandingWrite) (Branding, error) {
	if zero(org) || zero(actor) {
		return Branding{}, newError(CodeInvalidID)
	}
	return scanBranding(tx.QueryRow(ctx, sqlInsertBranding, uuidArg(org), w.Label, objectOrNull(w.LogoObjectID),
		overridesArg(w.Light), overridesArg(w.Dark), uuidArg(actor)))
}

// UpdateBrandingTx replaces the row's content when its version is still version, advancing it by one. A row at another
// version (or none) is ErrBrandingAbsent.
func UpdateBrandingTx(ctx context.Context, tx pgx.Tx, org OrganizationID, version int64, actor UserID, w BrandingWrite) (Branding, error) {
	if zero(org) || zero(actor) {
		return Branding{}, newError(CodeInvalidID)
	}
	return scanBranding(tx.QueryRow(ctx, sqlUpdateBranding, uuidArg(org), version, w.Label, objectOrNull(w.LogoObjectID),
		overridesArg(w.Light), overridesArg(w.Dark), uuidArg(actor)))
}

// NormalizeLabelTx returns the NFC normalization of s (the caller has refused invalid UTF-8 and control characters).
func NormalizeLabelTx(ctx context.Context, tx pgx.Tx, s string) (string, error) {
	var out string
	if err := tx.QueryRow(ctx, sqlNormalizeLabel, s).Scan(&out); err != nil {
		return "", err
	}
	return out, nil
}

// ShareLogoObjectTx reads and locks the Organization object id FOR SHARE (row-level security admits only the context's
// rows). A missing or invisible row is object_not_found.
func ShareLogoObjectTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id ObjectID) (StoredObject, error) {
	return objectRow(ctx, tx, sqlShareLogoObject, org, id)
}

// ListLogoUploadsTx lists at most MaxLogoUploadsListed unreferenced logo uploads of the Organization, oldest first,
// with the total number of such objects. current is the current reference (zero: none).
func ListLogoUploadsTx(ctx context.Context, tx pgx.Tx, org OrganizationID, current ObjectID) ([]LogoUpload, int, error) {
	if zero(org) {
		return nil, 0, newError(CodeInvalidID)
	}
	rows, err := tx.Query(ctx, sqlListLogoUploads, uuidArg(org), objectOrNull(current))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []LogoUpload
	total := 0
	for rows.Next() {
		var u LogoUpload
		var id pgtype.UUID
		var state string
		var sha []byte
		var n int64
		if err := rows.Scan(&id, &u.ContentType, &u.ByteSize, &state, &sha, &n); err != nil {
			return nil, 0, err
		}
		u.ID, u.State, total = ObjectID(id.Bytes), ObjectState(state), int(n)
		if len(sha) == 32 {
			copy(u.SHA256[:], sha)
			u.HasSHA256 = true
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ReadPublicBranding reads the public projection of org through the auth reader's definer function, as one statement
// on q (no row: Present false).
func ReadPublicBranding(ctx context.Context, q Querier, org OrganizationID) (PublicBranding, error) {
	if zero(org) {
		return PublicBranding{}, newError(CodeInvalidID)
	}
	var p PublicBranding
	var sha []byte
	var ct, key *string
	var size *int64
	err := q.QueryRow(ctx, sqlPublicBranding, uuidArg(org)).Scan(&p.Revision, &p.Label, &p.Light, &p.Dark, &sha, &ct, &size, &key)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicBranding{Light: map[string]string{}, Dark: map[string]string{}}, nil
	}
	if err != nil {
		return PublicBranding{}, err
	}
	p.Present = true
	if p.Light == nil {
		p.Light = map[string]string{}
	}
	if p.Dark == nil {
		p.Dark = map[string]string{}
	}
	if len(sha) == 32 && ct != nil && size != nil && key != nil {
		copy(p.LogoSHA256[:], sha)
		p.HasLogo, p.LogoContentType, p.LogoSize, p.LogoKey = true, *ct, *size, *key
	}
	return p, nil
}
