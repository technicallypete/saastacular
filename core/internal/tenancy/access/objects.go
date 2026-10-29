package access

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
	"strings"
	"time"

	"core/internal/audit"
	"core/internal/objectstore"
	"core/internal/tenancy"
)

// The guarded storage operations (G009 Plan r1 §4.1-§4.5; docs/foundation/OBJECT-STORAGE.md). Each runs the binding
// check, then, in a short storage transaction (storageTx): session liveness, row-level-security scope discovery for an
// object selector, bounded exact-scope facts, the context revalidation and the closed storage decision, all before any
// provider call or disclosure. Provider I/O happens only between transactions; no connection is held during a
// transfer. Exactly one storage.operation record is written per call.

// PutObjectRequest is an upload: the owning Group (zero: an Organization object of the active Organization), the
// content type (empty: application/octet-stream), the exact size (required; -1 or less is refused) and the body.
// PutObject owns Body from entry: it closes it exactly once on every path and never reads it after returning. Close
// must be safe concurrently with Read and must make a blocked Read return promptly (io.Pipe readers, net.Conn-backed
// bodies and files conform).
type PutObjectRequest struct {
	Group       tenancy.GroupID
	ContentType string
	Size        int64
	Body        io.ReadCloser
}

// SignedGetURLRequest selects an object and exactly one TTL form: DefaultTTL (TTLSeconds 0) or TTLSeconds 1..max.
type SignedGetURLRequest struct {
	Object     tenancy.ObjectID
	TTLSeconds int64
	DefaultTTL bool
}

// ObjectInfo is a ready object's metadata. Group is zero for an Organization object.
type ObjectInfo struct {
	ID           tenancy.ObjectID
	Organization tenancy.OrganizationID
	Group        tenancy.GroupID
	ContentType  string
	Size         int64
	SHA256       [32]byte
	CreatedAt    time.Time
	ReadyAt      time.Time
}

// DeleteOutcome is the result of an authorized delete.
type DeleteOutcome string

// The delete outcomes: deleted (the row is a tombstone and the provider bytes are confirmed absent) or deletion_pending
// (the object is logically deleted; physical removal is not confirmed, and a repeat delete completes it).
const (
	DeleteOutcomeDeleted         DeleteOutcome = "deleted"
	DeleteOutcomeDeletionPending DeleteOutcome = "deletion_pending"
)

// DeleteResult is an authorized delete's outcome.
type DeleteResult struct {
	Outcome DeleteOutcome
}

// objectHooks are inert test seams (nil in production, U-OBJ-SEAM): stage is called at each Put and Delete stage
// boundary; crash, when it returns true, abandons the call right there as a simulated process interruption (nothing
// further runs, not even compensation); sourceRead runs on the transport goroutine after each upload source Read has
// returned and before its bytes are digested (the C1 cancellation window).
type objectHooks struct {
	stage      func(op, stage string)
	crash      func(op, stage string) bool
	sourceRead func()
}

func (s *Service) at(op, stage string) bool {
	if s.hooks == nil {
		return false
	}
	if s.hooks.stage != nil {
		s.hooks.stage(op, stage)
	}
	return s.hooks.crash != nil && s.hooks.crash(op, stage)
}

// errSimulatedCrash is the failure of a call a test seam interrupted.
var errSimulatedCrash = fail(CodeUnavailable, reasonCanceled)

// storageTx is every storage transaction: inContextTx with the context statement first, a fixed decision budget for
// BEGIN through COMMIT and a fixed cleanup budget for the deferred rollback (Plan r1 §4.3a).
func (s *Service) storageTx(ctx context.Context, write bool, dbc dbContext, fn func(ctx context.Context, tx *trackedTx) *failure) *failure {
	ctx, cancel := context.WithTimeout(ctx, storageDecisionBudget)
	defer cancel()
	// fn receives the budgeted context: every statement of the transaction runs under the decision budget.
	return s.inContextTx(ctx, write, dbc, storageCleanupBudget, func(tx *trackedTx) *failure { return fn(ctx, tx) })
}

func dbcOf(tc *TenantContext) dbContext {
	return dbContext{user: tc.subject, organization: tc.organization, group: tc.group}
}

func objectInfo(o tenancy.StoredObject) ObjectInfo {
	return ObjectInfo{ID: o.ID, Organization: o.OrganizationID, Group: o.GroupID, ContentType: o.ContentType, Size: o.ByteSize,
		SHA256: o.SHA256, CreatedAt: o.CreatedAt, ReadyAt: o.ReadyAt}
}

func bindingOf(o tenancy.StoredObject) objectstore.Binding {
	b := objectstore.Binding{Object: o.ID.String(), Organization: o.OrganizationID.String()}
	if !zero(o.GroupID) {
		b.Group = o.GroupID.String()
	}
	return b
}

// keyOf derives the provider key in Go and asserts it equals the database-generated key of the row.
func keyOf(o tenancy.StoredObject) (objectstore.Key, bool) {
	k, err := objectstore.NewKey(o.OrganizationID.String(), o.ID.String())
	if err != nil || o.StorageKey != tenancy.ObjectStorageKey(o.OrganizationID, o.ID) || o.StorageKey != k.String() {
		return objectstore.Key{}, false
	}
	return k, true
}

// newObjectID returns a version-4 UUID from crypto/rand, generated before the insert so that a lost reply is
// reconcilable.
func newObjectID() (tenancy.ObjectID, error) {
	var id tenancy.ObjectID
	if _, err := rand.Read(id[:]); err != nil {
		return tenancy.ObjectID{}, err
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	if zero(id) {
		return tenancy.ObjectID{}, errors.New("access: zero object id")
	}
	return id, nil
}

// begin is the common start of a storage call: the binding check (attribution is recorded only after it passes) and
// the configured store.
func (s *Service) beginStorage(ctx context.Context, tc *TenantContext, rec *storageRecord) (*requestScope, *failure) {
	sc, f := s.bind(ctx, tc)
	if f != nil {
		return nil, f
	}
	rec.bound, rec.actor, rec.org = true, tc.subject, tc.organization
	if s.objects == nil {
		return nil, fail(CodeUnavailable, reasonObjectsUnconfigured)
	}
	return sc, nil
}

// --- reads -------------------------------------------------------------------------------------------------------------

// readObject is the read guard of Metadata, Get and Sign (Plan r1 §4.4, correction 3), in one REPEATABLE READ READ ONLY
// snapshot: the context statement, the live session, row-level-security scope discovery of the selector in the active
// Organization, the bounded facts of the discovered scope (the owning Group plus the selected Group; the
// Organization-only query when nothing was discovered), the context check, then the read decision on the discovered
// scope; the row must be ready. The row is returned only after the allow.
func (s *Service) readObject(ctx context.Context, tc *TenantContext, id tenancy.ObjectID, rec *storageRecord) (tenancy.StoredObject, *failure) {
	sc, f := s.beginStorage(ctx, tc, rec)
	if f != nil {
		return tenancy.StoredObject{}, f
	}
	if zero(id) {
		return tenancy.StoredObject{}, inputFailureCode(tenancy.CodeInvalidID)
	}
	rec.object = id
	sel := tc.selection()
	var out tenancy.StoredObject
	f = s.storageTx(ctx, false, dbcOf(tc), func(ctx context.Context, tx *trackedTx) *failure {
		if err := s.id.CheckLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
			return sessionFailure(ctx, tx, err)
		}
		row, err := tenancy.ReadObjectTx(ctx, tx, tc.organization, id)
		found := err == nil
		if err != nil && tenancy.CodeOf(err) != tenancy.CodeObjectNotFound {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		var owning tenancy.GroupID
		if found {
			owning = row.GroupID
		}
		facts, err := tenancy.ReadFacts(ctx, tx, objectFactQuery(tc, owning, false))
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if r := contextCheck(sel, facts); r != "" {
			return markDenied(fail(CodeContextDenied, r), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		if !found {
			return markDenied(fail(CodeNotFound, reasonTargetNotFound), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		t := objectTarget{group: row.GroupID, creator: row.CreatorUserID}
		if ok, r := decideObject(actionRead, sel, facts, t, tc.subject); !ok {
			return markDenied(objectDenial(actionRead, sel, facts, t, tc.subject, r), sel, facts, audit.ResourceStoredObject,
				audit.UUID(row.ID), row.GroupID)
		}
		if row.State != tenancy.ObjectReady || !row.HasSHA256 {
			return markDenied(fail(CodeNotFound, reasonObjectNotReady), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		out = row
		return nil
	})
	if f != nil {
		return tenancy.StoredObject{}, f
	}
	rec.group = out.GroupID
	return out, nil
}

// ObjectMetadata returns a ready object's metadata (organization.read or group.read on its owning Group). No provider
// call is made.
func (s *Service) ObjectMetadata(ctx context.Context, tc *TenantContext, id tenancy.ObjectID) (ObjectInfo, error) {
	rec := &storageRecord{op: opObjectMetadata}
	row, f := s.readObject(ctx, tc, id, rec)
	s.recordGuarded(ctx, tc, rec.op, f)
	s.logStorage(ctx, rec, f)
	if f != nil {
		return ObjectInfo{}, f.err()
	}
	return objectInfo(row), nil
}

// GetObject returns a ready object's metadata and its bytes. After the read decision it fetches the exact key under
// the transfer deadline and verifies the binding metadata and length before returning. The returned reader is bounded
// to the stored size and verifies the stored SHA-256 at EOF (a mismatch is an internal error there, after any
// delivered bytes); Close cancels the transfer. A missing provider object for a ready row is not_found.
func (s *Service) GetObject(ctx context.Context, tc *TenantContext, id tenancy.ObjectID) (ObjectInfo, io.ReadCloser, error) {
	rec := &storageRecord{op: opObjectGet}
	info, rc, f := s.getObject(ctx, tc, id, rec)
	s.recordGuarded(ctx, tc, rec.op, f)
	s.logStorage(ctx, rec, f)
	if f != nil {
		return ObjectInfo{}, nil, f.err()
	}
	return info, rc, nil
}

func (s *Service) getObject(ctx context.Context, tc *TenantContext, id tenancy.ObjectID, rec *storageRecord) (ObjectInfo, io.ReadCloser, *failure) {
	row, f := s.readObject(ctx, tc, id, rec)
	if f != nil {
		return ObjectInfo{}, nil, f
	}
	key, ok := keyOf(row)
	if !ok {
		return ObjectInfo{}, nil, fail(CodeInternal, reasonBindingMismatch)
	}
	tctx, cancel := context.WithTimeout(ctx, s.ocfg.TransferTimeout)
	pinfo, body, err := s.objects.Get(tctx, key)
	if err != nil {
		cancel()
		if errors.Is(err, objectstore.ErrNotFound) {
			return ObjectInfo{}, nil, fail(CodeNotFound, reasonObjectBytesMissing)
		}
		return ObjectInfo{}, nil, fail(CodeUnavailable, reasonProviderUnavailable)
	}
	if pinfo.Binding != bindingOf(row) || pinfo.Size != row.ByteSize {
		_ = body.Close()
		cancel()
		return ObjectInfo{}, nil, fail(CodeInternal, reasonBindingMismatch)
	}
	return objectInfo(row), &verifyingReader{body: body, cancel: cancel, remaining: row.ByteSize, want: row.SHA256, hash: sha256.New()}, nil
}

// SignedGetURL issues an offline presigned GET of a ready object's exact key over the trusted signing endpoint for the
// requested TTL (default 60 s, at most the configured maximum, never above 300 s), with the attachment, stored
// content-type and no-store response overrides. The URL is a bearer secret (SignedURL redacts it). An issued URL stays
// usable until it expires or the bytes are physically removed, even after revocation or logical deletion (X1).
func (s *Service) SignedGetURL(ctx context.Context, tc *TenantContext, req SignedGetURLRequest) (SignedURL, error) {
	rec := &storageRecord{op: opObjectSign}
	u, f := s.signedGetURL(ctx, tc, req, rec)
	s.recordGuarded(ctx, tc, rec.op, f)
	s.logStorage(ctx, rec, f)
	if f != nil {
		return SignedURL{}, f.err()
	}
	return u, nil
}

func (s *Service) signedGetURL(ctx context.Context, tc *TenantContext, req SignedGetURLRequest, rec *storageRecord) (SignedURL, *failure) {
	if _, f := s.beginStorage(ctx, tc, rec); f != nil {
		return SignedURL{}, f
	}
	if zero(req.Object) {
		return SignedURL{}, inputFailureCode(tenancy.CodeInvalidID)
	}
	ttl, ok := s.ocfg.signedTTL(req)
	if !ok {
		rec.object = req.Object
		return SignedURL{}, inputFailureCode(tenancy.CodeInvalidTTL)
	}
	row, f := s.readObject(ctx, tc, req.Object, rec)
	if f != nil {
		return SignedURL{}, f
	}
	key, ok := keyOf(row)
	if !ok {
		return SignedURL{}, fail(CodeInternal, reasonBindingMismatch)
	}
	pctx, cancel := context.WithTimeout(ctx, providerControlTimeout)
	defer cancel()
	issued := time.Now()
	u, err := s.objects.SignedGetURL(pctx, key, ttl, row.ContentType)
	if err != nil || !strings.HasPrefix(u, "http") {
		return SignedURL{}, fail(CodeUnavailable, reasonProviderUnavailable)
	}
	// M4 (G011): the issuance is durable before the URL is returned; no URL without its object.url_issued event. The
	// URL, signature and key are never stored.
	rec.stage = stageAudit
	dbc, e := objectEvent(tc, audit.ObjectURLIssued, audit.Success, "", row.ID, row.GroupID, true)
	e.TTLSeconds = int(ttl / time.Second)
	if f := s.auditTx(ctx, dbc, e); f != nil {
		return SignedURL{}, fail(CodeUnavailable, f.reason)
	}
	rec.stage = ""
	rec.ttl = int64(ttl / time.Second)
	return SignedURL{url: u, ExpiresAt: issued.Add(ttl), TTL: ttl}, nil
}

// --- put -------------------------------------------------------------------------------------------------------------

// PutObject uploads a new object (Plan r1 §4.4): an Organization object needs org_owner or org_admin; a Group object
// org_owner, org_admin or any current role in that exact Group. It inserts a pending row (T1), streams exactly Size
// bytes to the exact key outside any transaction (P2), verifies the stored length and binding (V), re-authorizes and
// publishes the row as ready with the local digest (T2). An unknown publish COMMIT is reconciled by locking the row
// (R). Compensation (C: provider delete, confirmed by Head) runs only when the row is authoritatively not ready
// (invariant I-FIN). Metadata never becomes ready unless the stream was exact, the provider reported success and V
// passed.
func (s *Service) PutObject(ctx context.Context, tc *TenantContext, req PutObjectRequest) (ObjectInfo, error) {
	rec := &storageRecord{op: opObjectPut}
	var src *ownedSource
	if req.Body != nil {
		src = newOwnedSource(req.Body)
		defer func() { _ = src.Close() }()
	}
	info, f := s.putObject(ctx, tc, req, src, rec)
	s.recordGuarded(ctx, tc, rec.op, f)
	s.logStorage(ctx, rec, f)
	if f != nil {
		return ObjectInfo{}, f.err()
	}
	return info, nil
}

func (s *Service) putObject(ctx context.Context, tc *TenantContext, req PutObjectRequest, src *ownedSource, rec *storageRecord) (ObjectInfo, *failure) {
	if _, f := s.beginStorage(ctx, tc, rec); f != nil {
		return ObjectInfo{}, f
	}
	rec.group = req.Group
	switch {
	case req.Size < 0:
		return ObjectInfo{}, inputFailureCode(tenancy.CodeObjectSizeRequired)
	case req.Size > s.ocfg.MaxObjectBytes:
		return ObjectInfo{}, inputFailureCode(tenancy.CodeObjectTooLarge)
	}
	ct, err := tenancy.ValidateContentType(req.ContentType)
	if err != nil {
		return ObjectInfo{}, inputFailureCode(tenancy.CodeInvalidContentType)
	}
	if src == nil {
		return ObjectInfo{}, inputFailureCode(tenancy.CodeMissingBody)
	}
	// The transfer deadline covers the body: cancellation or the deadline closes the source, unblocking a blocked Read.
	tctx, release := transferContext(ctx, src, s.ocfg.TransferTimeout)
	defer release()

	// P1: a zero-length upload must really be empty, before any database or provider work.
	if req.Size == 0 {
		switch probeEmpty(tctx, src) {
		case verdictLong:
			return ObjectInfo{}, inputFailureCode(tenancy.CodeObjectSizeMismatch)
		case verdictNotReleased:
			return ObjectInfo{}, fail(CodeUnavailable, reasonSourceNotReleased)
		case verdictSourceError:
			return ObjectInfo{}, fail(CodeUnavailable, reasonCanceled)
		}
	}

	// T1: authorize and insert the pending row.
	id, err := newObjectID()
	if err != nil {
		return ObjectInfo{}, fail(CodeInternal, reasonStoreUnavailable)
	}
	sel := tc.selection()
	dbc := dbcOf(tc)
	t := objectTarget{group: req.Group, creator: tc.subject}
	authorize := func(ctx context.Context, tx *trackedTx) *failure {
		if err := s.id.LockLiveSession(ctx, tx, tc.session, tc.scope.subjectText); err != nil {
			return sessionFailure(ctx, tx, err)
		}
		facts, err := tenancy.ReadFacts(ctx, tx, objectFactQuery(tc, req.Group, true))
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if r := contextCheck(sel, facts); r != "" {
			return markDenied(fail(CodeContextDenied, r), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		if ok, r := decideObject(actionPut, sel, facts, t, tc.subject); !ok {
			return markDenied(objectDenial(actionPut, sel, facts, t, tc.subject, r), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		return nil
	}
	// uploadFailed records object.upload_failed (G4) after a post-T1 failure and its compensation attempt.
	uploadFailed := func(o audit.Outcome, r audit.Reason, withOrg bool) {
		s.recordObject(ctx, tc, audit.ObjectUploadFailed, o, r, id, req.Group, withOrg)
	}
	if s.at(opObjectPut, stageT1) {
		return ObjectInfo{}, errSimulatedCrash
	}
	var row tenancy.StoredObject
	rec.stage = stageT1
	f := s.storageTx(ctx, true, dbc, func(ctx context.Context, tx *trackedTx) *failure {
		if f := authorize(ctx, tx); f != nil {
			return f
		}
		var err error
		if row, err = tenancy.InsertObjectTx(ctx, tx, tc.organization, id, req.Group, tc.subject, req.Size, ct); err != nil {
			return objectStoreFailure(ctx, tx, err)
		}
		if row.ID != id || row.State != tenancy.ObjectPending {
			return fail(CodeInternal, reasonFactInconsistent)
		}
		return nil
	})
	if f != nil {
		return ObjectInfo{}, f // an unknown COMMIT is unavailable with no provider call (a pending row has no bytes: X2)
	}
	rec.object = id
	key, ok := keyOf(row)
	if !ok {
		return ObjectInfo{}, fail(CodeInternal, reasonBindingMismatch)
	}
	binding := bindingOf(row)
	if s.at(opObjectPut, stageProvider) {
		return ObjectInfo{}, errSimulatedCrash
	}

	// P2: stream exactly Size bytes; then close the source and reap any in-flight Read.
	rec.stage = stageProvider
	verdict, sum, perr := s.upload(tctx, src, key, req.Size, ct, binding)
	switch {
	case verdict == verdictExact && perr == nil:
	case verdict == verdictShort || verdict == verdictLong:
		ok := s.compensate(ctx, key)
		r := reasonCompensated
		if !ok {
			r = reasonCompensationFailed
		}
		uploadFailed(audit.Failed, compensation(ok), true)
		return ObjectInfo{}, &failure{code: CodeInvalidInput, field: string(tenancy.CodeObjectSizeMismatch), reason: r}
	case verdict == verdictNotReleased:
		uploadFailed(audit.Failed, compensation(s.compensate(ctx, key)), true)
		return ObjectInfo{}, fail(CodeUnavailable, reasonSourceNotReleased)
	case verdict == verdictSourceError:
		uploadFailed(audit.Failed, compensation(s.compensate(ctx, key)), true)
		return ObjectInfo{}, fail(CodeUnavailable, reasonCanceled)
	default:
		uploadFailed(audit.Failed, compensation(s.compensate(ctx, key)), true)
		return ObjectInfo{}, fail(CodeUnavailable, reasonProviderOutcomeUnknown)
	}
	if s.at(opObjectPut, stageVerify) {
		return ObjectInfo{}, errSimulatedCrash
	}

	// V: the stored length and binding.
	rec.stage = stageVerify
	vctx, cancelV := context.WithTimeout(ctx, providerControlTimeout)
	pinfo, herr := s.objects.Metadata(vctx, key)
	cancelV()
	switch {
	case herr != nil:
		uploadFailed(audit.Failed, compensation(s.compensate(ctx, key)), true)
		return ObjectInfo{}, fail(CodeUnavailable, reasonProviderUnavailable)
	case pinfo.Binding != binding:
		uploadFailed(audit.Failed, compensation(s.compensate(ctx, key)), true)
		return ObjectInfo{}, fail(CodeInternal, reasonBindingMismatch)
	case pinfo.Size != req.Size:
		uploadFailed(audit.Failed, compensation(s.compensate(ctx, key)), true)
		return ObjectInfo{}, fail(CodeUnavailable, reasonIntegrityMismatch)
	}
	if s.at(opObjectPut, stageT2) {
		return ObjectInfo{}, errSimulatedCrash
	}

	// T2: re-authorize and publish.
	rec.stage = stageT2
	var final tenancy.StoredObject
	f = s.storageTx(ctx, true, dbc, func(ctx context.Context, tx *trackedTx) *failure {
		if f := authorize(ctx, tx); f != nil {
			return f
		}
		cur, err := tenancy.LockObjectTx(ctx, tx, tc.organization, id)
		if err != nil {
			return objectStoreFailure(ctx, tx, err)
		}
		if cur.State != tenancy.ObjectPending || cur.CreatorUserID != tc.subject {
			return fail(CodeNotFound, reasonObjectNotReady)
		}
		if final, err = tenancy.FinalizeObjectTx(ctx, tx, tc.organization, id, sum); err != nil {
			return objectStoreFailure(ctx, tx, err)
		}
		return nil
	})
	switch {
	case f == nil:
		rec.stage = ""
		return objectInfo(final), nil
	case f.reason != reasonCommitOutcomeUnknown:
		// Definitive: the row is not ready. A T2 policy denial is not a new access.denied (the T1 decision allowed the
		// upload); its upload_failed event carries no Organization, since the scope is no longer established.
		withOrg := f.denied == nil
		f.denied = nil
		uploadFailed(audit.Failed, compensation(s.compensate(ctx, key)), withOrg)
		return ObjectInfo{}, f
	}
	if s.at(opObjectPut, stageReconcile) {
		return ObjectInfo{}, errSimulatedCrash
	}

	// R: lock the row (waiting out an in-flight T2) and act only on its authoritative state.
	rec.stage = stageReconcile
	rctx := context.WithoutCancel(ctx)
	var cur tenancy.StoredObject
	rf := s.storageTx(rctx, true, dbc, func(ctx context.Context, tx *trackedTx) *failure {
		var err error
		if cur, err = tenancy.LockObjectTx(ctx, tx, tc.organization, id); err != nil {
			return objectStoreFailure(ctx, tx, err)
		}
		return nil
	})
	switch {
	case rf != nil:
		uploadFailed(audit.Unknown, audit.ReasonNotCompensated, true)
		return ObjectInfo{}, fail(CodeUnavailable, reasonCommitOutcomeUnknown) // the row may be ready: no compensation
	case cur.State == tenancy.ObjectReady && cur.HasSHA256 && cur.SHA256 == sum:
		rec.stage, rec.reason = "", reasonReconciledReady
		return objectInfo(cur), nil
	}
	uploadFailed(audit.Failed, compensation(s.compensate(ctx, key)), true)
	return ObjectInfo{}, fail(CodeUnavailable, reasonCommitOutcomeUnknown)
}

// upload is P2: the provider Put of exactly size bytes of src through the exact reader (an empty body for size 0) under
// tctx, whose cancellation or deadline closes src; after the provider call returns, on every path, src is closed and an
// in-flight Read is reaped for at most the reap grace. It returns the stream verdict (verdictNotReleased when a Read
// did not return), the digest of the released bytes and the provider error.
func (s *Service) upload(tctx context.Context, src *ownedSource, key objectstore.Key, size int64, ct string,
	b objectstore.Binding) (uploadVerdict, [32]byte, error) {
	var body io.Reader = strings.NewReader("")
	var xr *exactReader
	if size > 0 {
		xr = newExactReader(src, size)
		if s.hooks != nil {
			xr.afterRead = s.hooks.sourceRead
		}
		body = xr
	}
	perr := s.objects.Put(tctx, key, body, size, ct, b)
	_ = src.Close()
	released := src.waitReleased(sourceReapGrace)
	verdict, sum := verdictExact, sha256.Sum256(nil)
	if xr != nil {
		verdict, sum = xr.result() // the digest only with verdictExact, published after the final hash write (C1)
	}
	if !released {
		verdict = verdictNotReleased
	}
	return verdict, sum, perr
}

// transferContext is the transfer deadline of an upload: cancellation of ctx or the deadline closes src, which makes a
// conforming source's blocked Read return. The returned function releases the deadline and the close hook.
func transferContext(ctx context.Context, src *ownedSource, d time.Duration) (context.Context, func()) {
	tctx, cancel := context.WithTimeout(ctx, d)
	stop := context.AfterFunc(tctx, func() { _ = src.Close() })
	return tctx, func() { stop(); cancel() }
}

// compensate deletes the exact key and confirms its absence with a Head (C). It runs on a context detached from the
// caller's cancellation, bounded by the provider control timeout per call. Metadata is not changed. A failure leaves
// inaccessible, attributable bytes under a non-ready row; an authorized Delete repairs it (X3).
func (s *Service) compensate(ctx context.Context, key objectstore.Key) bool {
	if s.at(opObjectPut, stageCompensate) {
		return false
	}
	base := context.WithoutCancel(ctx)
	dctx, cancel := context.WithTimeout(base, providerControlTimeout)
	err := s.objects.Delete(dctx, key)
	cancel()
	if err != nil {
		return false
	}
	hctx, cancel := context.WithTimeout(base, providerControlTimeout)
	defer cancel()
	_, err = s.objects.Metadata(hctx, key)
	return errors.Is(err, objectstore.ErrNotFound)
}

// --- delete ----------------------------------------------------------------------------------------------------------

// DeleteObject deletes an object (Plan r1 §4.4): an Organization object needs org_owner or org_admin; a Group object
// org_owner, org_admin, the Group's owner or admin, or a group_member that created it. D1 decides on the object's
// immutable scope and creator and marks it deleting; D2/D3 remove the provider bytes and confirm their absence; D4
// tombstones the row. The result is deleted only when the row is deleted and absence is confirmed, otherwise
// deletion_pending (a repeat delete by an authorized deleter completes it). A failed provider delete is never reported
// as confirmed.
func (s *Service) DeleteObject(ctx context.Context, tc *TenantContext, id tenancy.ObjectID) (DeleteResult, error) {
	rec := &storageRecord{op: opObjectDelete}
	res, f := s.deleteObject(ctx, tc, id, rec)
	s.recordGuarded(ctx, tc, rec.op, f)
	s.logStorage(ctx, rec, f)
	if f != nil {
		return DeleteResult{}, f.err()
	}
	return res, nil
}

func (s *Service) deleteObject(ctx context.Context, tc *TenantContext, id tenancy.ObjectID, rec *storageRecord) (DeleteResult, *failure) {
	if _, f := s.beginStorage(ctx, tc, rec); f != nil {
		return DeleteResult{}, f
	}
	if zero(id) {
		return DeleteResult{}, inputFailureCode(tenancy.CodeInvalidID)
	}
	rec.object = id
	sel := tc.selection()
	dbc := dbcOf(tc)
	// live: the session liveness check, first in D1 and D4 after the context statement (Plan r1 §1 item 3: binding,
	// liveness, then row-level-security discovery, facts and the decision).
	live := func(ctx context.Context, tx *trackedTx) *failure {
		if err := s.id.LockLiveSession(ctx, tx, tc.session, tc.scope.subjectText); err != nil {
			return sessionFailure(ctx, tx, err)
		}
		return nil
	}
	// decide: the facts of the immutable scope, the context check and the delete decision, before the row lock (so the
	// lifecycle scope of row-level security never hides a readable row and turns forbidden into not_found).
	decide := func(ctx context.Context, tx *trackedTx, row tenancy.StoredObject) *failure {
		facts, err := tenancy.ReadFacts(ctx, tx, objectFactQuery(tc, row.GroupID, true))
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if r := contextCheck(sel, facts); r != "" {
			return markDenied(fail(CodeContextDenied, r), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		if zero(row.ID) {
			return markDenied(fail(CodeNotFound, reasonTargetNotFound), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		t := objectTarget{group: row.GroupID, creator: row.CreatorUserID}
		if ok, r := decideObject(actionDelete, sel, facts, t, tc.subject); !ok {
			return markDenied(objectDenial(actionDelete, sel, facts, t, tc.subject, r), sel, facts, audit.ResourceStoredObject,
				audit.UUID(row.ID), row.GroupID)
		}
		return nil
	}
	if s.at(opObjectDelete, stageT1) {
		return DeleteResult{}, errSimulatedCrash
	}

	// D1: liveness, discover, decide, lock, mark deleting.
	rec.stage = stageT1
	var row tenancy.StoredObject
	f := s.storageTx(ctx, true, dbc, func(ctx context.Context, tx *trackedTx) *failure {
		if f := live(ctx, tx); f != nil {
			return f
		}
		found, err := tenancy.ReadObjectTx(ctx, tx, tc.organization, id)
		if err != nil && tenancy.CodeOf(err) != tenancy.CodeObjectNotFound {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if f := decide(ctx, tx, found); f != nil {
			return f
		}
		if row, err = tenancy.LockObjectTx(ctx, tx, tc.organization, id); err != nil {
			return objectStoreFailure(ctx, tx, err)
		}
		if row.State == tenancy.ObjectPending || row.State == tenancy.ObjectReady {
			if row, err = tenancy.MarkObjectDeletingTx(ctx, tx, tc.organization, id); err != nil {
				return objectStoreFailure(ctx, tx, err)
			}
		}
		return nil
	})
	if f != nil {
		return DeleteResult{}, f // an unknown COMMIT may have marked it deleting; a retry completes it
	}
	rec.group = row.GroupID
	key, ok := keyOf(row)
	if !ok {
		return DeleteResult{}, fail(CodeInternal, reasonBindingMismatch)
	}
	// pending reports deletion_pending and records object.deletion_pending (G5): the earlier object.delete_started
	// event stays the durable record of the logical deletion; this one names the unconfirmed physical removal.
	pending := func(r Reason, o audit.Outcome, ar audit.Reason, withOrg bool) (DeleteResult, *failure) {
		rec.physical, rec.reason = physicalPending, r
		s.recordObject(ctx, tc, audit.ObjectDeletionPending, o, ar, id, row.GroupID, withOrg)
		return DeleteResult{Outcome: DeleteOutcomeDeletionPending}, nil
	}
	if s.at(opObjectDelete, stageProvider) {
		return DeleteResult{}, errSimulatedCrash
	}

	// D2/D3: remove the bytes and confirm their absence.
	rec.stage = stageProvider
	dctx, cancel := context.WithTimeout(ctx, providerControlTimeout)
	err := s.objects.Delete(dctx, key)
	cancel()
	if err != nil {
		return pending(reasonProviderUnavailable, audit.Failed, audit.ReasonProviderUnavailable, true)
	}
	if s.at(opObjectDelete, stageVerify) {
		return DeleteResult{}, errSimulatedCrash
	}
	rec.stage = stageVerify
	hctx, cancel := context.WithTimeout(ctx, providerControlTimeout)
	_, herr := s.objects.Metadata(hctx, key)
	cancel()
	if !errors.Is(herr, objectstore.ErrNotFound) {
		return pending(reasonProviderUnavailable, audit.Failed, audit.ReasonProviderUnavailable, true)
	}
	if s.at(opObjectDelete, stageT2) {
		return DeleteResult{}, errSimulatedCrash
	}

	// D4: tombstone a deleting row.
	if row.State != tenancy.ObjectDeleted {
		rec.stage = stageT2
		f = s.storageTx(ctx, true, dbc, func(ctx context.Context, tx *trackedTx) *failure {
			if f := live(ctx, tx); f != nil {
				return f
			}
			cur, err := tenancy.ReadObjectTx(ctx, tx, tc.organization, id)
			if err != nil {
				return objectStoreFailure(ctx, tx, err)
			}
			if f := decide(ctx, tx, cur); f != nil {
				return f
			}
			if cur, err = tenancy.LockObjectTx(ctx, tx, tc.organization, id); err != nil {
				return objectStoreFailure(ctx, tx, err)
			}
			if cur.State == tenancy.ObjectDeleting {
				if _, err = tenancy.MarkObjectDeletedTx(ctx, tx, tc.organization, id); err != nil {
					return objectStoreFailure(ctx, tx, err)
				}
			} else if cur.State != tenancy.ObjectDeleted {
				return fail(CodeInternal, reasonFactInconsistent)
			}
			return nil
		})
		if f != nil {
			// D4 definitive: tombstone_failed; unknown COMMIT: tombstone_unknown (object.deleted may also exist). A D4
			// policy denial records no access.denied and no Organization (the scope is no longer established).
			o, ar := audit.Failed, audit.ReasonTombstoneFailed
			if f.reason == reasonCommitOutcomeUnknown {
				o, ar = audit.Unknown, audit.ReasonTombstoneUnknown
			}
			return pending(f.reason, o, ar, f.denied == nil)
		}
	}
	rec.stage, rec.physical = "", physicalConfirmed
	return DeleteResult{Outcome: DeleteOutcomeDeleted}, nil
}
