package access

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/branding"
	"core/internal/objectstore"
	"core/internal/tenancy"
)

// G017 guarded branding operations (Plan r2 §1.3, §1.4, §5.1). Each authorizes and acts in one transaction through
// guard, like every facade operation, with organization.manage_branding (Organization owner or admin, owner decision
// D-1) on the Organization target: liveness, the Organization facts (FOR SHARE for writes), the context check and the
// decision, then the branding row FOR NO KEY UPDATE (lock order user -> session -> Organization -> branding row) and one
// version-guarded statement. The row is created at version 1 only by an update or a logo association; reset
// and logo removal never create it and never delete an object. Every statement lives in core/internal/tenancy
// (branding.go); the triggers own the version rule, the logo-reference check and the single organization.branding_changed
// audit row.

// The branding operations' logged names (and access.denied operations).
const (
	opGetBranding        = "branding.get"
	opUpdateBranding     = "branding.update"
	opResetBranding      = "branding.reset"
	opSetBrandingLogo    = "branding.set_logo"
	opRemoveBrandingLogo = "branding.remove_logo"
)

// The live states of the current logo reference: the stored object's own state, or absent once it is purged.
const (
	logoReady    = "ready"
	logoDeleting = "deleting"
	logoDeleted  = "deleted"
	logoAbsent   = "absent"
)

// BrandingLogo is the current logo reference with its live object state. SHA12 is the first 12 hex digits of the
// stored bytes' SHA-256 (the public cache revision), empty unless the object is ready.
type BrandingLogo struct {
	ObjectID    string
	ContentType string
	ByteSize    int64
	SHA12       string
	State       string
}

// BrandingView is the management view of the active Organization's branding. Version 0 means no row (the platform
// defaults apply); Label nil means no label; Logo nil means no logo reference. Uploads (GetBranding only) are the
// unreferenced logo uploads, at most 100 with their total, which the owner may delete through the guarded DeleteObject.
type BrandingView struct {
	Label        *string
	Light        map[string]string
	Dark         map[string]string
	Version      int64
	Logo         *BrandingLogo
	Uploads      []BrandingLogo
	UploadsTotal int
}

// BrandingUpdate is the full new label and colour configuration. ExpectedVersion 0 means "no row yet".
type BrandingUpdate struct {
	ExpectedVersion int64
	Label           *string
	Light, Dark     map[string]string
}

func brandingInput(field string) *failure {
	return &failure{code: CodeInvalidInput, field: field, reason: reasonInputInvalid}
}

// brandingValidation maps a closed branding validation error onto the facade's codes.
func brandingValidation(err error) *failure {
	if e, ok := branding.AsError(err); ok {
		if e.Code == branding.CodeContrastFailed {
			return &failure{code: CodeContrastFailed, field: e.Field, reason: reasonInputInvalid}
		}
		return brandingInput(e.Field)
	}
	return fail(CodeInternal, reasonStoreUnavailable)
}

// brandingFailure maps an error of a branding statement: a lost version race or a guard refusal is conflict, a
// constraint refusal invalid_input, a missing row conflict (the row the caller's version named is gone); a store
// failure is unavailable with the transaction's reason.
func brandingFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	var pg *pgconn.PgError
	switch {
	case errors.Is(err, tenancy.ErrBrandingAbsent), errors.Is(err, pgx.ErrNoRows):
		return fail(CodeConflict, reasonWriteConflict)
	case errors.As(err, &pg) && pg.ConstraintName == "organization_branding_guard":
		return fail(CodeConflict, reasonWriteConflict)
	case errors.As(err, &pg) && pg.Code == "23505":
		return fail(CodeConflict, reasonWriteConflict)
	case errors.As(err, &pg) && pg.Code == "23514" && pg.ConstraintName == "organization_branding_label_valid":
		return brandingInput("display_label")
	case errors.As(err, &pg) && pg.Code == "23514" && pg.ConstraintName == "organization_branding_light_valid":
		return brandingInput("light")
	case errors.As(err, &pg) && pg.Code == "23514" && pg.ConstraintName == "organization_branding_dark_valid":
		return brandingInput("dark")
	case errors.As(err, &pg) && commandError(pg) && len(pg.Code) == 5 && pg.Code[:2] != "40" && pg.Code[:2] != "08" &&
		pg.Code[:2] != "57":
		return fail(CodeInternal, reasonStoreUnavailable)
	}
	return fail(CodeUnavailable, failureReason(ctx, tx.cause))
}

func manageBranding(s *Service, ctx context.Context, tc *TenantContext, op string, write bool, pre func() *failure,
	act func(tx *trackedTx, org tenancy.OrganizationID) *failure) error {
	return s.guard(ctx, tc, op, capOrganizationManageBranding, write, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, f Facts) *failure { return act(tx, f.Organization.ID) })
}

// brandingView builds the view of a row (or of no row), reading the current logo's live state.
func brandingView(ctx context.Context, tx *trackedTx, b tenancy.Branding, present bool) (BrandingView, *failure) {
	v := BrandingView{Light: map[string]string{}, Dark: map[string]string{}}
	if !present {
		return v, nil
	}
	v.Label, v.Light, v.Dark, v.Version = b.Label, b.Light, b.Dark, b.Version
	if zero(b.LogoObjectID) {
		return v, nil
	}
	logo := &BrandingLogo{ObjectID: b.LogoObjectID.String(), State: logoAbsent}
	o, err := tenancy.ReadObjectTx(ctx, tx, b.OrganizationID, b.LogoObjectID)
	switch {
	case tenancy.CodeOf(err) == tenancy.CodeObjectNotFound:
	case err != nil:
		return BrandingView{}, fail(CodeUnavailable, failureReason(ctx, tx.cause))
	default:
		logo.ContentType, logo.ByteSize, logo.State = o.ContentType, o.ByteSize, string(o.State)
		if o.State == tenancy.ObjectReady && o.HasSHA256 {
			logo.SHA12 = hex.EncodeToString(o.SHA256[:6])
		}
	}
	v.Logo = logo
	return v, nil
}

// lockBranding locks the Organization's row and checks the caller's expected version: 0 must find no row, any other
// value must equal the locked version; a mismatch is conflict with zero effects.
func lockBranding(ctx context.Context, tx *trackedTx, org tenancy.OrganizationID, expected int64) (tenancy.Branding, bool, *failure) {
	b, err := tenancy.LockBrandingTx(ctx, tx, org)
	present := err == nil
	if err != nil && !errors.Is(err, tenancy.ErrBrandingAbsent) {
		return tenancy.Branding{}, false, brandingFailure(ctx, tx, err)
	}
	if (present && b.Version != expected) || (!present && expected != 0) {
		return tenancy.Branding{}, false, fail(CodeConflict, reasonWriteConflict)
	}
	return b, present, nil
}

func versionInput(v int64) func() *failure {
	return func() *failure {
		if v < 0 {
			return brandingInput("expected_version")
		}
		return nil
	}
}

// GetBranding returns the active Organization's branding view (version 0 and empty overrides when it has none).
func (s *Service) GetBranding(ctx context.Context, tc *TenantContext) (BrandingView, error) {
	var out BrandingView
	err := manageBranding(s, ctx, tc, opGetBranding, false, nil, func(tx *trackedTx, org tenancy.OrganizationID) *failure {
		b, err := tenancy.ReadBrandingTx(ctx, tx, org)
		if err != nil && !errors.Is(err, tenancy.ErrBrandingAbsent) {
			return brandingFailure(ctx, tx, err)
		}
		v, f := brandingView(ctx, tx, b, err == nil)
		if f != nil {
			return f
		}
		ups, total, lerr := tenancy.ListLogoUploadsTx(ctx, tx, org, b.LogoObjectID)
		if lerr != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		v.Uploads, v.UploadsTotal = make([]BrandingLogo, 0, len(ups)), total
		for _, u := range ups {
			l := BrandingLogo{ObjectID: u.ID.String(), ContentType: u.ContentType, ByteSize: u.ByteSize, State: string(u.State)}
			if u.State == tenancy.ObjectReady && u.HasSHA256 {
				l.SHA12 = hex.EncodeToString(u.SHA256[:6])
			}
			v.Uploads = append(v.Uploads, l)
		}
		out = v
		return nil
	})
	return out, err
}

// UpdateBranding replaces the label and both appearances' overrides when the row is still at ExpectedVersion (0: no
// row yet, which creates it at version 1), keeping the logo reference. The overrides' grammar and every contrast pair
// are checked before any transaction; the label is trimmed, checked for control characters, NFC-normalized by the
// database and then bounded (1-80 code points) inside the guarded transaction. Any refusal commits nothing.
func (s *Service) UpdateBranding(ctx context.Context, tc *TenantContext, in BrandingUpdate) (BrandingView, error) {
	var out BrandingView
	var label *string
	var light, dark branding.Overrides
	pre := func() *failure {
		if f := versionInput(in.ExpectedVersion)(); f != nil {
			return f
		}
		var err error
		if label, err = branding.PrepareLabel(in.Label); err != nil {
			return brandingValidation(err)
		}
		if light, dark, err = branding.ValidateColors(in.Light, in.Dark); err != nil {
			return brandingValidation(err)
		}
		return nil
	}
	err := manageBranding(s, ctx, tc, opUpdateBranding, true, pre, func(tx *trackedTx, org tenancy.OrganizationID) *failure {
		stored := label
		if label != nil {
			n, err := tenancy.NormalizeLabelTx(ctx, tx, *label)
			if err != nil {
				return brandingFailure(ctx, tx, err)
			}
			if stored, err = branding.ValidateLabel(&n); err != nil {
				return brandingValidation(err)
			}
		}
		cur, present, f := lockBranding(ctx, tx, org, in.ExpectedVersion)
		if f != nil {
			return f
		}
		w := tenancy.BrandingWrite{Label: stored, LogoObjectID: cur.LogoObjectID, Light: light, Dark: dark}
		var b tenancy.Branding
		var err error
		if present {
			b, err = tenancy.UpdateBrandingTx(ctx, tx, org, cur.Version, tc.subject, w)
		} else {
			b, err = tenancy.InsertBrandingTx(ctx, tx, org, tc.subject, w)
		}
		if err != nil {
			return brandingFailure(ctx, tx, err)
		}
		out, f = brandingView(ctx, tx, b, true)
		return f
	})
	return out, err
}

// ResetBranding returns the Organization to the platform defaults (no label, no overrides, no logo reference) when the
// row is still at expected, advancing its version. It deletes no stored object (a former logo stays listed until its
// owner deletes it through the guarded DeleteObject). With no row (expected 0) it is a successful no-op.
func (s *Service) ResetBranding(ctx context.Context, tc *TenantContext, expected int64) (BrandingView, error) {
	return s.clearBranding(ctx, tc, opResetBranding, expected, func(cur tenancy.Branding) (tenancy.BrandingWrite, bool) {
		w := tenancy.BrandingWrite{}
		return w, cur.Label != nil || !zero(cur.LogoObjectID) || len(cur.Light) > 0 || len(cur.Dark) > 0
	})
}

// RemoveBrandingLogo clears the logo reference when the row is still at expected, advancing its version; the label and
// overrides are kept and no stored object is deleted. With no row (expected 0) or no reference it is a successful
// no-op at the current version.
func (s *Service) RemoveBrandingLogo(ctx context.Context, tc *TenantContext, expected int64) (BrandingView, error) {
	return s.clearBranding(ctx, tc, opRemoveBrandingLogo, expected, func(cur tenancy.Branding) (tenancy.BrandingWrite, bool) {
		return tenancy.BrandingWrite{Label: cur.Label, Light: cur.Light, Dark: cur.Dark}, !zero(cur.LogoObjectID)
	})
}

// clearBranding runs reset or logo removal: next derives the new content from the locked row and reports whether
// anything changes (an unchanged row is not rewritten, so no version advance and no audit row).
func (s *Service) clearBranding(ctx context.Context, tc *TenantContext, op string, expected int64,
	next func(tenancy.Branding) (tenancy.BrandingWrite, bool)) (BrandingView, error) {
	var out BrandingView
	err := manageBranding(s, ctx, tc, op, true, versionInput(expected), func(tx *trackedTx, org tenancy.OrganizationID) *failure {
		cur, present, f := lockBranding(ctx, tx, org, expected)
		if f != nil {
			return f
		}
		w, changed := tenancy.BrandingWrite{}, false
		if present {
			w, changed = next(cur)
		}
		if !changed {
			out, f = brandingView(ctx, tx, cur, present)
			return f
		}
		b, err := tenancy.UpdateBrandingTx(ctx, tx, org, cur.Version, tc.subject, w)
		if err != nil {
			return brandingFailure(ctx, tx, err)
		}
		out, f = brandingView(ctx, tx, b, true)
		return f
	})
	return out, err
}

// --- logo upload (Plan r2 §1.4.1, §1.4.2) ------------------------------------------------------------------------------

// The SetBrandingLogo stages of the inert objectHooks seam (operation branding.set_logo): after the upload (P), inside
// the association before its COMMIT is sent (A), before the report read of an unknown association outcome (R) and
// before the compensating delete (C). A crash at P abandons the call there (the upload stays listed, unpublished).
const (
	stageLogoUploaded   = "P"
	stageLogoAssociate  = "A"
	stageLogoReport     = "R"
	stageLogoCompensate = "C"
)

// The closed logo outcome fields (Error.Field) of SetBrandingLogo after its upload, and the unread live state.
const (
	fieldPublicationUnknown  = "publication_outcome_unknown"
	fieldCompensated         = "compensated"
	fieldCompensationPending = "compensation_pending"
	fieldCompensationFailed  = "compensation_failed"
	logoUnread               = "unread"
)

// SetBrandingLogo reads the body (exactly size bytes, at most 512 KiB), checks authorization and the expected version
// with a guarded read, and only then decodes and normalizes the image, before anything is stored (a refused, stale or
// invalid request uploads nothing, and an unauthorized one is never decoded), uploads the normalized bytes through the unchanged G009 PutObject as an Organization object (P), and then
// associates it in one guarded, version-bound write that locks the object FOR SHARE and re-reads it ready (A). A
// definitive abort of A (known rolled back) deletes the upload through the guarded DeleteObject (C) and reports the
// result in the error's closed field. An unknown A outcome (lost COMMIT reply) is never compensated: one guarded read
// (R) returns the view when the row references the object, and otherwise unavailable with publication_outcome_unknown,
// the object and its actual live state; the object is retained and GetBranding shows the true state afterwards.
func (s *Service) SetBrandingLogo(ctx context.Context, tc *TenantContext, expected int64, body io.Reader, size int64) (BrandingView, error) {
	var raw []byte
	// V: only the cheap bounded read and size checks run before authorization; decoding is after it (D-1).
	pre := func() *failure {
		switch {
		case expected < 0:
			return brandingInput("expected_version")
		case body == nil || size < 0:
			return brandingInput("logo_size")
		case size > branding.MaxLogoBytes:
			return brandingInput(branding.FieldLogoTooLarge)
		}
		buf, err := io.ReadAll(io.LimitReader(body, branding.MaxLogoBytes+1))
		switch {
		case err != nil:
			return fail(CodeUnavailable, failureReason(ctx, err))
		case len(buf) > branding.MaxLogoBytes:
			return brandingInput(branding.FieldLogoTooLarge)
		case int64(len(buf)) != size:
			return brandingInput("logo_size")
		}
		raw = buf
		return nil
	}
	// The pre-check: V, then authorization and the expected version, before any upload.
	err := manageBranding(s, ctx, tc, opSetBrandingLogo, false, pre, func(tx *trackedTx, org tenancy.OrganizationID) *failure {
		b, err := tenancy.ReadBrandingTx(ctx, tx, org)
		present := err == nil
		if err != nil && !errors.Is(err, tenancy.ErrBrandingAbsent) {
			return brandingFailure(ctx, tx, err)
		}
		if (present && b.Version != expected) || (!present && expected != 0) {
			return fail(CodeConflict, reasonWriteConflict)
		}
		return nil
	})
	if err != nil {
		return BrandingView{}, err
	}
	// Decode and normalize only for an authorized caller with the expected version, outside any transaction and
	// before anything is stored; a refusal keeps the closed validation mapping and is logged like the pre-check.
	logo, err := branding.NormalizeLogo(raw)
	if err != nil {
		vf := brandingValidation(err)
		s.logGuarded(ctx, opSetBrandingLogo, capOrganizationManageBranding, vf)
		return BrandingView{}, vf.err()
	}
	// P: the unchanged G009 upload of the normalized bytes (its own transactions, audit, compensation and recovery).
	info, err := s.PutObject(ctx, tc, PutObjectRequest{ContentType: logo.ContentType, Size: int64(len(logo.Data)),
		Body: io.NopCloser(bytes.NewReader(logo.Data))})
	if err != nil {
		return BrandingView{}, err
	}
	if s.at(opSetBrandingLogo, stageLogoUploaded) {
		return BrandingView{}, errSimulatedCrash.err()
	}
	digest := sha256.Sum256(logo.Data)
	// A: the version-bound association.
	var out BrandingView
	f := s.guarded(ctx, tc, capOrganizationManageBranding, true, Target{Kind: kindOrganization}, organizationQuery, nil,
		func(tx *trackedTx, facts Facts) *failure {
			org := facts.Organization.ID
			cur, present, f := lockBranding(ctx, tx, org, expected)
			if f != nil {
				return f
			}
			o, err := tenancy.ShareLogoObjectTx(ctx, tx, org, info.ID)
			switch {
			case tenancy.CodeOf(err) == tenancy.CodeObjectNotFound:
				return fail(CodeConflict, reasonWriteConflict)
			case err != nil:
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			case o.State != tenancy.ObjectReady || !zero(o.GroupID) || o.ContentType != logo.ContentType ||
				o.ByteSize != int64(len(logo.Data)) || !o.HasSHA256 || o.SHA256 != digest:
				return fail(CodeConflict, reasonWriteConflict) // deleted meanwhile, or not the validated upload
			}
			w := tenancy.BrandingWrite{Label: cur.Label, LogoObjectID: info.ID, Light: cur.Light, Dark: cur.Dark}
			var b tenancy.Branding
			if present {
				b, err = tenancy.UpdateBrandingTx(ctx, tx, org, cur.Version, tc.subject, w)
			} else {
				b, err = tenancy.InsertBrandingTx(ctx, tx, org, tc.subject, w)
			}
			if err != nil {
				return brandingFailure(ctx, tx, err)
			}
			if out, f = brandingView(ctx, tx, b, true); f != nil {
				return f
			}
			s.at(opSetBrandingLogo, stageLogoAssociate)
			return nil
		})
	s.recordGuarded(ctx, tc, opSetBrandingLogo, f)
	s.logGuarded(ctx, opSetBrandingLogo, capOrganizationManageBranding, f)
	switch {
	case f == nil:
		return out, nil
	case f.reason == reasonCommitOutcomeUnknown:
		return s.reportLogo(ctx, tc, info.ID)
	}
	return BrandingView{}, s.compensateLogo(ctx, tc, info.ID, f)
}

// logoRead is one guarded snapshot read of the branding row and the object's live state (R, and the state after a
// failed compensation). ok is false when the read itself failed.
func (s *Service) logoRead(ctx context.Context, tc *TenantContext, id tenancy.ObjectID) (view BrandingView, referenced bool, state string, ok bool) {
	state = logoUnread
	err := manageBranding(s, ctx, tc, opSetBrandingLogo, false, nil, func(tx *trackedTx, org tenancy.OrganizationID) *failure {
		b, err := tenancy.ReadBrandingTx(ctx, tx, org)
		if err != nil && !errors.Is(err, tenancy.ErrBrandingAbsent) {
			return brandingFailure(ctx, tx, err)
		}
		o, oerr := tenancy.ReadObjectTx(ctx, tx, org, id)
		switch {
		case tenancy.CodeOf(oerr) == tenancy.CodeObjectNotFound:
			state = logoAbsent
		case oerr != nil:
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		default:
			state = string(o.State)
		}
		if err == nil && b.LogoObjectID == id {
			referenced = true
			var f *failure
			view, f = brandingView(ctx, tx, b, true)
			return f
		}
		return nil
	})
	if err != nil {
		return BrandingView{}, false, logoUnread, false
	}
	return view, referenced, state, true
}

// reportLogo is R: a snapshot can confirm a commit but never prove non-commit, so anything but a row referencing the
// object is publication_outcome_unknown with the object retained.
func (s *Service) reportLogo(ctx context.Context, tc *TenantContext, id tenancy.ObjectID) (BrandingView, error) {
	s.at(opSetBrandingLogo, stageLogoReport)
	view, referenced, state, _ := s.logoRead(ctx, tc, id)
	if referenced {
		return view, nil
	}
	return BrandingView{}, (&failure{code: CodeUnavailable, field: fieldPublicationUnknown, object: id.String(), objectState: state}).err()
}

// compensateLogo is C after a definitive abort of A: the guarded G009 delete of the upload, reported as G009 returned
// it; when the delete fails, the object's actual live state is read and reported (no retry loop).
func (s *Service) compensateLogo(ctx context.Context, tc *TenantContext, id tenancy.ObjectID, af *failure) error {
	out := &failure{code: af.code, object: id.String()}
	var res DeleteResult
	err := errSimulatedCrash.err()
	if !s.at(opSetBrandingLogo, stageLogoCompensate) {
		res, err = s.DeleteObject(ctx, tc, id)
	}
	switch {
	case err == nil && res.Outcome == DeleteOutcomeDeleted:
		out.field, out.objectState = fieldCompensated, logoDeleted
	case err == nil:
		out.field, out.objectState = fieldCompensationPending, logoDeleting
	default:
		_, _, state, _ := s.logoRead(ctx, tc, id)
		out.field, out.objectState = fieldCompensationFailed, state
	}
	return out.err()
}

// DeleteBrandingUpload deletes one unreferenced logo upload through the guarded G009 DeleteObject (D-3 as disposed):
// object must be listed by the current view as an upload (never the current logo, never another object), and the view
// must still be at expected (a stale view is conflict). DeleteObject then decides on its own rules (Organization object:
// owner or admin) and reports deleted or deletion_pending truthfully.
func (s *Service) DeleteBrandingUpload(ctx context.Context, tc *TenantContext, expected int64, object string) (DeleteResult, error) {
	id, err := tenancy.ParseObjectID(object)
	if err != nil {
		return DeleteResult{}, brandingInput("object_id").err()
	}
	v, err := s.GetBranding(ctx, tc)
	switch {
	case err != nil:
		return DeleteResult{}, err
	case v.Version != expected:
		return DeleteResult{}, fail(CodeConflict, reasonWriteConflict).err()
	}
	for _, u := range v.Uploads {
		if u.ObjectID == id.String() {
			return s.DeleteObject(ctx, tc, id)
		}
	}
	return DeleteResult{}, fail(CodeNotFound, reasonTargetNotFound).err()
}

// --- public projection (Plan r2 §1.5, §5.3, §5.5) -----------------------------------------------------------------------

// PublicLogo is the public logo reference: V is the first 12 hex digits of the stored bytes' SHA-256 (the cache
// revision) and the server-detected content type. No object ID, key or size leaves core's public routes.
type PublicLogo struct {
	V           string
	ContentType string
}

// PublicBrandingView is the intentionally public branding of one Organization: its revision (0: no row, platform
// defaults), label, overrides and current ready logo.
type PublicBrandingView struct {
	Revision    int64
	Label       *string
	Light, Dark map[string]string
	Logo        *PublicLogo
}

// publicRead reads the public projection of an Organization the caller resolved from a trusted host. An anonymous
// read has no subject, so it cannot open the facade's context-initialized transaction: it is one statement on the
// pool calling the auth reader's definer projection (like G016's host resolver), never a protected relation directly.
func (s *Service) publicRead(ctx context.Context, organization string) (tenancy.PublicBranding, *failure) {
	org, err := tenancy.ParseOrganizationID(organization)
	if err != nil {
		return tenancy.PublicBranding{}, fail(CodeNotFound, reasonTargetNotFound)
	}
	q, ok := s.db.(tenancy.Querier)
	if !ok {
		return tenancy.PublicBranding{}, fail(CodeUnavailable, reasonStoreUnavailable)
	}
	p, err := tenancy.ReadPublicBranding(ctx, q, org)
	if err != nil {
		return tenancy.PublicBranding{}, fail(CodeUnavailable, failureReason(ctx, err))
	}
	return p, nil
}

// PublicBranding returns the public projection of organization, which the caller resolved from the request's host
// through the trusted G016 resolver. It is never authorized by a credential and never logged here.
func (s *Service) PublicBranding(ctx context.Context, organization string) (PublicBrandingView, error) {
	p, f := s.publicRead(ctx, organization)
	if f != nil {
		return PublicBrandingView{}, f.err()
	}
	v := PublicBrandingView{Revision: p.Revision, Label: p.Label, Light: p.Light, Dark: p.Dark}
	if p.HasLogo {
		v.Logo = &PublicLogo{V: hex.EncodeToString(p.LogoSHA256[:6]), ContentType: p.LogoContentType}
	}
	return v, nil
}

// PublicBrandingLogo returns the bytes of organization's current ready logo when v is its current revision; anything
// else (no logo, a stale or foreign v, an object no longer ready) is not_found. The bytes are read from the owned store
// by the projection's key, checked against the key's binding, length and SHA-256, and returned whole (at most 512 KiB).
func (s *Service) PublicBrandingLogo(ctx context.Context, organization, v string) (string, []byte, error) {
	p, f := s.publicRead(ctx, organization)
	switch {
	case f != nil:
		return "", nil, f.err()
	case !p.HasLogo || len(v) != 12 || hex.EncodeToString(p.LogoSHA256[:6]) != v:
		return "", nil, fail(CodeNotFound, reasonTargetNotFound).err()
	case s.objects == nil:
		return "", nil, fail(CodeUnavailable, reasonObjectsUnconfigured).err()
	case p.LogoSize < 0 || p.LogoSize > branding.MaxLogoBytes:
		return "", nil, fail(CodeInternal, reasonBindingMismatch).err()
	}
	key, err := objectstore.ParseKey(p.LogoKey)
	if err != nil {
		return "", nil, fail(CodeInternal, reasonBindingMismatch).err()
	}
	tctx, cancel := context.WithTimeout(ctx, s.ocfg.TransferTimeout)
	defer cancel()
	pinfo, rc, err := s.objects.Get(tctx, key)
	switch {
	case errors.Is(err, objectstore.ErrNotFound):
		return "", nil, fail(CodeNotFound, reasonObjectBytesMissing).err()
	case err != nil:
		return "", nil, fail(CodeUnavailable, reasonProviderUnavailable).err()
	}
	defer func() { _ = rc.Close() }()
	want, kerr := objectstore.NewKey(organization, pinfo.Binding.Object)
	if kerr != nil || want != key || pinfo.Binding.Organization != organization || pinfo.Binding.Group != "" || pinfo.Size != p.LogoSize {
		return "", nil, fail(CodeInternal, reasonBindingMismatch).err()
	}
	data, err := io.ReadAll(io.LimitReader(rc, p.LogoSize+1))
	switch {
	case err != nil:
		return "", nil, fail(CodeUnavailable, reasonProviderUnavailable).err()
	case int64(len(data)) != p.LogoSize || sha256.Sum256(data) != p.LogoSHA256:
		return "", nil, fail(CodeInternal, reasonBindingMismatch).err()
	}
	return p.LogoContentType, data, nil
}
