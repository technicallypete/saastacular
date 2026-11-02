package access

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/audit"
	"core/internal/branding"
	"core/internal/objectstore"
	"core/internal/tenancy"
)

// G074 guarded account avatar operations (Plan r3.1 §3.5, §3.8). An avatar is an identity-owned object: its bytes live
// under avatars/<user>/<avatar> in the owned store (core's existing scoped credential), its metadata in
// app.account_avatars with the G009 lifecycle, and the current reference in the caller's own app.account_profiles row.
// The storage code paths and their meaning are G009's (objects.go): the owned upload source, the exact streaming upload,
// verification by Head, compensation confirmed by Head, storage transactions with fixed budgets, and provider I/O only
// between transactions. Three safeguards are specific to avatars:
//
//   - Reconcile-then-gate (C11): every Set first reconciles the caller's own non-current rows (S0b: at most 20, oldest
//     first; pending rows younger than the lease are never touched) and then, under the profile row lock in the
//     allocation transaction, allocates only when no non-current, non-tombstoned row remains; otherwise it allocates
//     nothing and answers a repairable HOLD (unavailable with field avatar_cleanup_pending or avatar_upload_in_progress
//     and a retry delay, RetryAfterOf). The selection and the gate are null-safe (IS DISTINCT FROM) against a NULL
//     reference.
//   - Fail-closed reconcile (C12): after an unknown publish COMMIT, compensation runs only when a successful locked read
//     proves the new row is not current (pending or deleting); a failed, unavailable or unknown read retains the row and
//     its bytes and answers the HOLD commit_outcome_unknown, which the caller's next own operation settles (I-FIN).
//   - I-CUR: a withdrawal (replacement or removal) moves or clears the reference first and withdraws the old row in the
//     same transaction (the guard refuses a current avatar leaving ready).
//
// Reads of another user's avatar are decided now, on every read, by the policy reader's projection
// (app.rls_account_profile): no signed URL is ever issued for an avatar key. Each call writes exactly one final
// profile.storage record, plus one record per reconcile, cleanup or unknown-outcome step; never a key, URL, size, name or
// digest.

// The avatar operations' logged names (and access.denied operations, 00023).
const (
	opProfileAvatarGet    = "profile.avatar.get"
	opProfileAvatarSet    = "profile.avatar.set"
	opProfileAvatarRemove = "profile.avatar.remove"
)

// The avatar stages beside G009's (t1, provider, verify, t2, reconcile, compensate): the S0b reconcile sweep and the
// cleanup of a withdrawn row (S5/S6, R2/R3). They name the objectHooks seam stages and the records' stage.
const (
	stageAvatarSweep   = "sweep"
	stageAvatarCleanup = "cleanup"
)

// The closed HOLD fields of SetAccountAvatar (Error.Field with CodeUnavailable; each carries a retry delay), and the
// invalid_input field of a body whose length is not the declared size.
const (
	fieldAvatarCleanupPending   = "avatar_cleanup_pending"
	fieldAvatarUploadInProgress = "avatar_upload_in_progress"
	fieldCommitOutcomeUnknown   = "commit_outcome_unknown"
	fieldAvatarSize             = "avatar_size"
)

// The avatar record reasons beside the storage reasons.
const (
	reasonAvatarCleanupPending   Reason = fieldAvatarCleanupPending
	reasonAvatarUploadInProgress Reason = fieldAvatarUploadInProgress
	reasonReconcileUnknown       Reason = "reconcile_unknown"
	reasonDeletionPending        Reason = "deletion_pending"
	reasonReconciled             Reason = "reconciled"
)

// avatarHoldRetry is the retry delay of a cleanup or unknown-outcome HOLD; avatarLeaseGrace is the margin of the lease.
const (
	avatarHoldRetry  = 30 * time.Second
	avatarLeaseGrace = 5 * time.Second
)

// eventProfileStorage is the operational record of the avatar operations: operation, outcome, closed reason, stage, the
// actor and the avatar UUID, and the row count of a reconcile step.
const eventProfileStorage = "profile.storage"

// avatarHold is the error of a repairable HOLD: the facade error (unavailable with its closed field) and the retry delay.
type avatarHold struct {
	err        *Error
	retryAfter time.Duration
}

func (h *avatarHold) Error() string { return h.err.Error() }

func (h *avatarHold) Unwrap() error { return h.err }

// RetryAfterOf returns the retry delay of a repairable avatar HOLD (SetAccountAvatar's unavailable with field
// avatar_cleanup_pending, avatar_upload_in_progress or commit_outcome_unknown), or false for any other error.
func RetryAfterOf(err error) (time.Duration, bool) {
	var h *avatarHold
	if errors.As(err, &h) && h.retryAfter > 0 {
		return h.retryAfter, true
	}
	return 0, false
}

// avatarError is the returned error of a failure: a HOLD carries its retry delay.
func avatarError(f *failure, hold time.Duration) error {
	e := &Error{Code: f.code, Field: f.field, Object: f.object, ObjectState: f.objectState}
	if hold > 0 {
		return &avatarHold{err: e, retryAfter: hold}
	}
	return e
}

// avatarRecord collects one avatar call's record fields.
type avatarRecord struct {
	op     string
	bound  bool // the binding check passed: the attribution below is trusted
	actor  tenancy.UserID
	object tenancy.AvatarID
	stage  string
	reason Reason // the success reason when not "allowed" (reconciled_ready)
}

func (s *Service) logAvatarAttrs(level slog.Level, r *avatarRecord, outcome, stage string, reason Reason, count int) {
	attrs := []slog.Attr{slog.String("operation", r.op), slog.String("outcome", outcome), slog.String("reason", string(reason))}
	if stage != "" {
		attrs = append(attrs, slog.String("stage", stage))
	}
	if r.bound {
		attrs = append(attrs, slog.String("actor_id", r.actor.String()))
		if !zero(r.object) {
			attrs = append(attrs, slog.String("avatar_id", r.object.String()))
		}
	}
	if count > 0 {
		attrs = append(attrs, slog.Int("count", count))
	}
	s.logger.LogAttrs(context.Background(), level, eventProfileStorage, attrs...)
}

// logAvatar writes the final record of one call: allow (info), deny (warn), failed (error: unavailable, internal or
// missing bytes under a ready row).
func (s *Service) logAvatar(r *avatarRecord, f *failure) {
	level, outcome, reason := slog.LevelInfo, outcomeAllow, reasonAllowed
	if r.reason != "" {
		reason = r.reason
	}
	if f != nil {
		level, outcome, reason = slog.LevelWarn, outcomeDeny, f.reason
		if f.code == CodeUnavailable || f.code == CodeInternal || f.reason == reasonObjectBytesMissing {
			level, outcome = slog.LevelError, outcomeFailed
		}
	}
	s.logAvatarAttrs(level, r, outcome, r.stage, reason, 0)
}

// logAvatarStep writes one intermediate step record (a reconcile, cleanup, compensation or unknown-outcome step): allow
// for a completed step, failed otherwise.
func (s *Service) logAvatarStep(r *avatarRecord, stage string, ok bool, reason Reason, count int) {
	if ok {
		s.logAvatarAttrs(slog.LevelInfo, r, outcomeAllow, stage, reason, count)
		return
	}
	s.logAvatarAttrs(slog.LevelError, r, outcomeFailed, stage, reason, count)
}

// beginAvatar is the common start of an avatar call: the binding check (attribution is recorded only after it passes)
// and the configured store.
func (s *Service) beginAvatar(ctx context.Context, tc *TenantContext, rec *avatarRecord) *failure {
	if _, f := s.bind(ctx, tc); f != nil {
		return f
	}
	rec.bound, rec.actor = true, tc.subject
	if s.objects == nil {
		return fail(CodeUnavailable, reasonObjectsUnconfigured)
	}
	return nil
}

// avatarLease is the lease L of a pending row (Plan r3.1 §3.8): an in-flight upload cannot outlive its transfer
// deadline, the source reap grace and the margin, so an older pending row is abandoned.
func (s *Service) avatarLease() time.Duration {
	return s.ocfg.TransferTimeout + sourceReapGrace + avatarLeaseGrace
}

// avatarAuthorize is the guarded part of an avatar transaction: the session liveness (locked for writes), the
// Organization facts (FOR SHARE for writes), the context check and the decision on c for the active Organization.
func (s *Service) avatarAuthorize(ctx context.Context, tx *trackedTx, tc *TenantContext, c Capability, write bool) (Facts, *failure) {
	live := s.id.CheckLiveSession
	if write {
		live = s.id.LockLiveSession
	}
	if err := live(ctx, tx, tc.session, tc.scope.subjectText); err != nil {
		return Facts{}, sessionFailure(ctx, tx, err)
	}
	q := organizationQuery(tc)
	q.Organization, q.User, q.Lock = tc.organization, tc.subject, write
	facts, err := tenancy.ReadFacts(ctx, tx, q)
	if err != nil {
		return Facts{}, fail(CodeUnavailable, failureReason(ctx, tx.cause))
	}
	sel := tc.selection()
	if r := contextCheck(sel, facts); r != "" {
		return Facts{}, markDenied(fail(CodeContextDenied, r), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
	}
	t := Target{Kind: kindOrganization, Organization: tc.organization}
	if ok, r := decide(c, sel, facts, t); !ok {
		res, id, g := targetScope(t)
		return Facts{}, markDenied(denial(c, sel, facts, t, r), sel, facts, res, id, g)
	}
	return facts, nil
}

// avatarLive is the guarded part of a cleanup transaction (the reconcile sweep, a compensation tombstone, a withdrawn
// row's tombstone): the session liveness only. The rows are the caller's own (row-level security and the guard's actor
// rule), and their cleanup confers nothing.
func (s *Service) avatarLive(ctx context.Context, tx *trackedTx, tc *TenantContext) *failure {
	if err := s.id.LockLiveSession(ctx, tx, tc.session, tc.scope.subjectText); err != nil {
		return sessionFailure(ctx, tx, err)
	}
	return nil
}

// avatarFailure maps an error of an avatar statement: a refused statement (a guard, policy or constraint refusal) is
// internal, never a policy outcome; a store failure is unavailable with the transaction's reason.
func avatarFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && commandError(pg) && len(pg.Code) == 5 && pg.Code[:2] != "40" && pg.Code[:2] != "08" && pg.Code[:2] != "57" {
		return fail(CodeInternal, reasonStoreUnavailable)
	}
	return fail(CodeUnavailable, failureReason(ctx, tx.cause))
}

// avatarValidation maps a closed image refusal of branding.NormalizeLogo onto the avatar fields (logo_* -> avatar_*).
func avatarValidation(err error) *failure {
	e, ok := branding.AsError(err)
	if !ok {
		return fail(CodeInternal, reasonStoreUnavailable)
	}
	switch e.Field {
	case branding.FieldLogoTooLarge:
		return invalid("avatar_too_large")
	case branding.FieldLogoType:
		return invalid("avatar_type")
	case branding.FieldLogoDimensions:
		return invalid("avatar_dimensions")
	case branding.FieldNormalizedTooLarge:
		return invalid(branding.FieldNormalizedTooLarge)
	}
	return invalid("avatar_malformed")
}

// lockAvatarProfile locks the caller's own profile row; with create it first inserts the empty first row when none
// exists (INSERT-or-lock), so that every allocation serializes on that lock. present is false only without create and
// without a row.
func lockAvatarProfile(ctx context.Context, tx *trackedTx, user tenancy.UserID, create bool) (tenancy.AvatarProfile, bool, *failure) {
	if create {
		if err := tenancy.InsertEmptyProfileTx(ctx, tx, user); err != nil {
			return tenancy.AvatarProfile{}, false, avatarFailure(ctx, tx, err)
		}
	}
	p, err := tenancy.LockAvatarProfileTx(ctx, tx, user)
	switch {
	case errors.Is(err, tenancy.ErrProfileAbsent):
		return tenancy.AvatarProfile{}, false, nil
	case err != nil:
		return tenancy.AvatarProfile{}, false, avatarFailure(ctx, tx, err)
	}
	return p, true, nil
}

// avatarVersionOK reports whether the caller's expected version still names the profile: no row is version 0, and an
// empty first row (created at version 1 by an earlier allocation, nothing ever set) is version 0 as well as 1.
func avatarVersionOK(p tenancy.AvatarProfile, present bool, expected int32) bool {
	switch {
	case !present:
		return expected == 0
	case p.Version == expected:
		return true
	}
	return expected == 0 && p.Empty()
}

func versionConflict() *failure {
	return &failure{code: CodeConflict, field: fieldVersionConflict, reason: reasonWriteConflict}
}

// avatarKey derives the provider key of an own avatar in Go and asserts it equals the database-generated key.
func avatarKey(a tenancy.Avatar) (objectstore.Key, bool) {
	k, err := objectstore.NewAvatarKey(a.UserID.String(), a.ID.String())
	if err != nil || a.StorageKey != tenancy.AvatarStorageKey(a.UserID, a.ID) || a.StorageKey != k.String() {
		return objectstore.Key{}, false
	}
	return k, true
}

func avatarKeyOf(user tenancy.UserID, id tenancy.AvatarID) (objectstore.Key, bool) {
	k, err := objectstore.NewAvatarKey(user.String(), id.String())
	return k, err == nil
}

func avatarBinding(user tenancy.UserID, id tenancy.AvatarID) objectstore.Binding {
	return objectstore.Binding{Object: id.String(), User: user.String()}
}

// removeAvatarBytes deletes the exact key and confirms its absence with a Head, on a context detached from the caller's
// cancellation, bounded by the provider control timeout per call.
func (s *Service) removeAvatarBytes(ctx context.Context, key objectstore.Key) bool {
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

// providerRefused reports a Put the provider answered and refused (denied or invalid, or rejected before any request):
// no object exists for it, unlike the unknown outcome of every other failure.
func providerRefused(err error) bool {
	return errors.Is(err, objectstore.ErrDenied) || errors.Is(err, objectstore.ErrInvalid)
}

// markAvatarUpload is S1b: it re-authorizes and sets the write-ahead upload marker of the pending row, in its own
// transaction, before the provider Put. A row that is no longer pending (swept) or already marked is
// object_not_ready, with no provider call.
func (s *Service) markAvatarUpload(ctx context.Context, tc *TenantContext, id tenancy.AvatarID) *failure {
	return s.storageTx(ctx, true, dbcOf(tc), func(ctx context.Context, tx *trackedTx) *failure {
		if _, f := s.avatarAuthorize(ctx, tx, tc, capAccountProfileManageSelf, true); f != nil {
			return f
		}
		if _, err := tenancy.MarkAvatarUploadStartedTx(ctx, tx, tc.subject, id); err != nil {
			if errors.Is(err, tenancy.ErrAvatarAbsent) {
				return fail(CodeUnavailable, reasonObjectNotReady)
			}
			return avatarFailure(ctx, tx, err)
		}
		return nil
	})
}

// releaseRefusedUpload withdraws the marked row of a Put that cannot have written anything (the provider refused it, or
// no request was issued): the marker is cleared with the move to deleting, so a compensation that then fails leaves an
// ordinary withdrawn row that the next sweep settles by deleting and confirming absence. Without it a failed deletion
// would leave a marked row whose absent object never settles it. Detached from the caller's cancellation; a failure
// leaves the row marked and is reported by the compensation that follows.
func (s *Service) releaseRefusedUpload(ctx context.Context, tc *TenantContext, id tenancy.AvatarID) {
	_ = s.storageTx(context.WithoutCancel(ctx), true, dbcOf(tc), func(ctx context.Context, tx *trackedTx) *failure {
		if f := s.avatarLive(ctx, tx, tc); f != nil {
			return f
		}
		if _, _, f := lockAvatarProfile(ctx, tx, tc.subject, false); f != nil {
			return f
		}
		if _, err := tenancy.MarkAvatarUploadRefusedTx(ctx, tx, tc.subject, id); err != nil && !errors.Is(err, tenancy.ErrAvatarAbsent) {
			return avatarFailure(ctx, tx, err)
		}
		return nil
	})
}

// settleObserved settles the row of a Put whose outcome may be unproven (it is marked and never became ready): the
// bytes are deleted, and their absence confirmed, only after a Head observed them. The one request of the row (no
// automatic retry, one generated key) has then completed, so it cannot write again; an absent object proves nothing
// (the request may still commit later), so the row stays unresolved and keeps blocking allocation, and a failed Head
// is unknown.
func (s *Service) settleObserved(ctx context.Context, key objectstore.Key) bool {
	hctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), providerControlTimeout)
	_, err := s.objects.Metadata(hctx, key)
	cancel()
	if err != nil {
		return false
	}
	return s.removeAvatarBytes(ctx, key)
}

// tombstoneAvatars marks the caller's rows deleted after their bytes were confirmed absent, in one cleanup transaction
// detached from the caller's cancellation: a pending (or non-current ready) row passes through deleting, a deleting row
// becomes deleted, a deleted row is left alone. The guard refuses the current avatar, which fails the transaction.
func (s *Service) tombstoneAvatars(ctx context.Context, tc *TenantContext, ids []tenancy.AvatarID) bool {
	if len(ids) == 0 {
		return true
	}
	f := s.storageTx(context.WithoutCancel(ctx), true, dbcOf(tc), func(ctx context.Context, tx *trackedTx) *failure {
		if f := s.avatarLive(ctx, tx, tc); f != nil {
			return f
		}
		if _, _, f := lockAvatarProfile(ctx, tx, tc.subject, false); f != nil {
			return f
		}
		for _, id := range ids {
			if _, err := tenancy.MarkAvatarDeletingTx(ctx, tx, tc.subject, id); err != nil && !errors.Is(err, tenancy.ErrAvatarAbsent) {
				return avatarFailure(ctx, tx, err)
			}
			if _, err := tenancy.MarkAvatarDeletedTx(ctx, tx, tc.subject, id); err != nil && !errors.Is(err, tenancy.ErrAvatarAbsent) {
				return avatarFailure(ctx, tx, err)
			}
		}
		return nil
	})
	return f == nil
}

// compensateAvatar is C for a new avatar that is authoritatively not current: delete the exact key, confirm absence,
// then tombstone its row (pending -> deleting -> deleted, or deleting -> deleted). A failure leaves inaccessible,
// attributable bytes under a non-ready row, reconciled by the next Set after the lease (and blocking allocation until
// then). It writes one compensate step record.
func (s *Service) compensateAvatar(ctx context.Context, tc *TenantContext, rec *avatarRecord, op string, id tenancy.AvatarID) bool {
	ok := false
	if !s.at(op, stageCompensate) {
		if key, valid := avatarKeyOf(tc.subject, id); valid && s.removeAvatarBytes(ctx, key) {
			ok = s.tombstoneAvatars(ctx, tc, []tenancy.AvatarID{id})
		}
	}
	r := reasonCompensated
	if !ok {
		r = reasonCompensationFailed
	}
	s.logAvatarStep(rec, stageCompensate, ok, r, 1)
	return ok
}

func compensationReason(ok bool) Reason {
	if ok {
		return reasonCompensated
	}
	return reasonCompensationFailed
}

// sweepAvatars is the reconcile (S0b, and the repair of R2/R3): tx A locks the profile row (with create, inserting the
// empty first row) and selects the caller's own non-current, non-tombstoned rows except pending rows younger than the
// lease (at most 20, oldest first), marking each deleting; between transactions each row's bytes are deleted and their
// absence confirmed, stopping at the first failure; tx B tombstones the confirmed rows. Young pending rows are never
// touched. A row that carries the upload marker and never became ready has an unproven provider outcome: it is settled
// (deleted and tombstoned) only after its bytes were observed, and otherwise stays deleting, held and blocking
// allocation. It returns false when the simulated-crash seam abandoned the call; every other outcome is left to the gate
// (a row that could not be cleaned stays deleting, never deliverable) and written as one sweep step record when any row
// was selected or the selection failed.
func (s *Service) sweepAvatars(ctx context.Context, tc *TenantContext, rec *avatarRecord, op string, create bool) bool {
	if s.at(op, stageAvatarSweep) {
		return false
	}
	lease := s.avatarLease()
	var rows []tenancy.Avatar
	f := s.storageTx(ctx, true, dbcOf(tc), func(ctx context.Context, tx *trackedTx) *failure {
		rows = nil
		if f := s.avatarLive(ctx, tx, tc); f != nil {
			return f
		}
		_, present, f := lockAvatarProfile(ctx, tx, tc.subject, create)
		if f != nil || !present {
			return f // no profile row: no avatar row can exist (an allocation always creates the row first)
		}
		stale, err := tenancy.StaleAvatarsTx(ctx, tx, tc.subject, lease)
		if err != nil {
			return avatarFailure(ctx, tx, err)
		}
		for _, a := range stale {
			if a.State != tenancy.ObjectDeleting {
				if a, err = tenancy.MarkAvatarDeletingTx(ctx, tx, tc.subject, a.ID); err != nil {
					return avatarFailure(ctx, tx, err)
				}
			}
			rows = append(rows, a)
		}
		return nil
	})
	if f != nil {
		s.logAvatarStep(rec, stageAvatarSweep, false, f.reason, 0)
		return true
	}
	if len(rows) == 0 {
		return true
	}
	var done []tenancy.AvatarID
	failed := false
	for _, a := range rows {
		key, ok := avatarKey(a)
		if !ok {
			failed = true
			break
		}
		if a.UploadStarted && a.ReadyAt.IsZero() {
			// A marked row that never became ready: its Put outcome is unproven, so only observed bytes settle it. An
			// absent object leaves the row deleting (held, still blocking allocation) for the next own operation.
			if !s.settleObserved(ctx, key) {
				continue
			}
		} else if !s.removeAvatarBytes(ctx, key) {
			failed = true
			break
		}
		done = append(done, a.ID)
	}
	tombstoned := s.tombstoneAvatars(ctx, tc, done)
	switch {
	case !tombstoned:
		s.logAvatarStep(rec, stageAvatarSweep, false, reasonStoreUnavailable, len(rows))
	case len(done) < len(rows):
		r := reasonDeletionPending
		if failed {
			r = reasonProviderUnavailable
		}
		s.logAvatarStep(rec, stageAvatarSweep, false, r, len(rows)-len(done))
	default:
		s.logAvatarStep(rec, stageAvatarSweep, true, reasonReconciled, len(done))
	}
	return true
}

// cleanupWithdrawn is S5/S6 (and R2/R3 of one withdrawn row): delete the old key, confirm absence, tombstone the row.
// A failure leaves the row deleting (its bytes are unreachable: delivery needs the current reference and ready) and is
// written as one cleanup step record; the next own avatar operation repairs it first. It returns false when the
// simulated-crash seam abandoned the call.
func (s *Service) cleanupWithdrawn(ctx context.Context, tc *TenantContext, rec *avatarRecord, op string, old tenancy.AvatarID) bool {
	if zero(old) {
		return true
	}
	if s.at(op, stageAvatarCleanup) {
		return false
	}
	key, ok := avatarKeyOf(tc.subject, old)
	if ok && s.removeAvatarBytes(ctx, key) && s.tombstoneAvatars(ctx, tc, []tenancy.AvatarID{old}) {
		s.logAvatarStep(rec, stageAvatarCleanup, true, reasonAllowed, 1)
		return true
	}
	s.logAvatarStep(rec, stageAvatarCleanup, false, reasonDeletionPending, 1)
	return true
}

// --- set -------------------------------------------------------------------------------------------------------------

// SetAccountAvatar replaces the caller's avatar (account_profile.manage_self) when its profile is still at expected (0:
// no row yet). It reads the body (exactly size bytes, at most 512 KiB), checks liveness, the capability and the expected
// version with a guarded read, and only then normalizes the image with the unchanged branding.NormalizeLogo (PNG or JPEG
// by content, 16..2048 px, at most 512 KiB in and out, metadata stripped); bytes identical to the current avatar are a
// successful no-op with nothing stored. It then reconciles the caller's own unresolved rows (S0b), allocates under the
// reconcile-then-gate rule (S1), marks the row's upload as started (S1b, the write-ahead marker that fences a swept row),
// uploads the normalized bytes (S2), verifies them (S3), publishes them and withdraws the
// old avatar in one version-bound transaction (S4) and removes the old bytes (S5/S6). Failures: invalid_input
// (avatar_too_large, avatar_type, avatar_dimensions, avatar_malformed, normalized_too_large, avatar_size), conflict
// version_conflict, and the HOLDs avatar_cleanup_pending, avatar_upload_in_progress and commit_outcome_unknown
// (unavailable with RetryAfterOf); other unavailable reasons are unchanged.
func (s *Service) SetAccountAvatar(ctx context.Context, tc *TenantContext, expected int32, body io.Reader, size int64) (AccountProfileView, error) {
	rec := &avatarRecord{op: opProfileAvatarSet}
	var hold time.Duration
	v, f := s.setAvatar(ctx, tc, expected, body, size, rec, &hold)
	s.recordGuarded(ctx, tc, rec.op, f)
	s.logAvatar(rec, f)
	if f != nil {
		return AccountProfileView{}, avatarError(f, hold)
	}
	return v, nil
}

func (s *Service) setAvatar(ctx context.Context, tc *TenantContext, expected int32, body io.Reader, size int64, rec *avatarRecord,
	hold *time.Duration) (AccountProfileView, *failure) {
	const op = opProfileAvatarSet
	if f := s.beginAvatar(ctx, tc, rec); f != nil {
		return AccountProfileView{}, f
	}
	// S0: the bounded read and size checks, then the guarded self read; decoding only after it.
	switch {
	case expected < 0:
		return AccountProfileView{}, invalid("expected_version")
	case body == nil || size < 0:
		return AccountProfileView{}, invalid(fieldAvatarSize)
	case size > branding.MaxLogoBytes:
		return AccountProfileView{}, invalid("avatar_too_large")
	}
	raw, err := io.ReadAll(io.LimitReader(body, branding.MaxLogoBytes+1))
	switch {
	case err != nil:
		return AccountProfileView{}, fail(CodeUnavailable, failureReason(ctx, err))
	case len(raw) > branding.MaxLogoBytes:
		return AccountProfileView{}, invalid("avatar_too_large")
	case int64(len(raw)) != size:
		return AccountProfileView{}, invalid(fieldAvatarSize)
	}
	dbc := dbcOf(tc)
	var pre AccountProfileView
	var curSum [32]byte
	hasCur := false
	f := s.storageTx(ctx, false, dbc, func(ctx context.Context, tx *trackedTx) *failure {
		if _, f := s.avatarAuthorize(ctx, tx, tc, capAccountProfileManageSelf, false); f != nil {
			return f
		}
		p, err := tenancy.ReadAvatarProfileTx(ctx, tx, tc.subject)
		present := err == nil
		if err != nil && !errors.Is(err, tenancy.ErrProfileAbsent) {
			return avatarFailure(ctx, tx, err)
		}
		if !avatarVersionOK(p, present, expected) {
			return versionConflict()
		}
		if present && !zero(p.AvatarID) {
			a, err := tenancy.ReadAvatarTx(ctx, tx, tc.subject, p.AvatarID)
			switch {
			case err == nil && a.State == tenancy.ObjectReady && a.HasSHA256:
				curSum, hasCur = a.SHA256, true
			case err != nil && !errors.Is(err, tenancy.ErrAvatarAbsent):
				return avatarFailure(ctx, tx, err)
			}
		}
		var f *failure
		pre, f = readOwn(ctx, tx, tc.subject)
		return f
	})
	if f != nil {
		return AccountProfileView{}, f
	}
	logo, err := branding.NormalizeLogo(raw)
	if err != nil {
		return AccountProfileView{}, avatarValidation(err)
	}
	sum := sha256.Sum256(logo.Data)
	if hasCur && sum == curSum {
		return pre, nil // identical bytes: nothing stored
	}
	n := int64(len(logo.Data))

	// S0b: reconcile the caller's own unresolved rows before allocating.
	if !s.sweepAvatars(ctx, tc, rec, op, true) {
		return AccountProfileView{}, errSimulatedCrash
	}

	// S1 (T1) and the gate: under the profile row lock, allocate only when no unresolved row remains.
	id, err := newObjectID()
	if err != nil {
		return AccountProfileView{}, fail(CodeInternal, reasonStoreUnavailable)
	}
	aid := tenancy.AvatarID(id)
	if s.at(op, stageT1) {
		return AccountProfileView{}, errSimulatedCrash
	}
	rec.stage = stageT1
	lease := s.avatarLease()
	var row tenancy.Avatar
	f = s.storageTx(ctx, true, dbc, func(ctx context.Context, tx *trackedTx) *failure {
		*hold = 0
		if _, f := s.avatarAuthorize(ctx, tx, tc, capAccountProfileManageSelf, true); f != nil {
			return f
		}
		p, present, f := lockAvatarProfile(ctx, tx, tc.subject, true)
		if f != nil {
			return f
		}
		if !avatarVersionOK(p, present, expected) {
			return versionConflict()
		}
		g, err := tenancy.AvatarGateTx(ctx, tx, tc.subject, lease)
		if err != nil {
			return avatarFailure(ctx, tx, err)
		}
		if g.Outstanding > 0 {
			if g.Young == g.Outstanding {
				*hold = g.RetryAfter
				if *hold < time.Second || *hold > lease {
					*hold = lease
				}
				return &failure{code: CodeUnavailable, field: fieldAvatarUploadInProgress, reason: reasonAvatarUploadInProgress}
			}
			*hold = avatarHoldRetry
			return &failure{code: CodeUnavailable, field: fieldAvatarCleanupPending, reason: reasonAvatarCleanupPending}
		}
		if row, err = tenancy.InsertAvatarTx(ctx, tx, tc.subject, aid, n, logo.ContentType); err != nil {
			return avatarFailure(ctx, tx, err)
		}
		if row.ID != aid || row.State != tenancy.ObjectPending {
			return fail(CodeInternal, reasonFactInconsistent)
		}
		return nil
	})
	if f != nil {
		return AccountProfileView{}, f // an unknown COMMIT is unavailable with no provider call (a pending row has no bytes)
	}
	rec.object = aid
	key, ok := avatarKey(row)
	if !ok {
		return AccountProfileView{}, fail(CodeInternal, reasonBindingMismatch)
	}
	binding := avatarBinding(tc.subject, aid)
	if s.at(op, stageProvider) {
		return AccountProfileView{}, errSimulatedCrash
	}

	// S1b: the write-ahead marker, durable before the provider is called. The marker update matches only a row that is
	// still pending, so a row a concurrent sweep already moved on gets no Put at all (the fence of Plan r3.1 §3.8): a
	// pending row without the marker can never have a Put in flight, a marked row's outcome is unproven until observed.
	if f := s.markAvatarUpload(ctx, tc, aid); f != nil {
		return AccountProfileView{}, f
	}

	// S2 (P): stream exactly the normalized bytes under the transfer deadline.
	rec.stage = stageProvider
	src := newOwnedSource(io.NopCloser(bytes.NewReader(logo.Data)))
	tctx, release := transferContext(ctx, src, s.ocfg.TransferTimeout)
	var verdict uploadVerdict
	var got [32]byte
	perr := tctx.Err()
	issued := perr == nil // no request is issued under an already ended context
	if issued {
		verdict, got, perr = s.upload(tctx, src, key, n, logo.ContentType, binding)
	}
	release()
	_ = src.Close()
	if verdict != verdictExact || perr != nil || got != sum {
		if issued && perr != nil && !providerRefused(perr) {
			// The issued Put's outcome is unproven: a transport failure, a deadline, a cancellation or any other
			// non-definitive answer does not stop the provider from committing the request later. No delete and no
			// tombstone now (either could precede the late bytes and leave them under a tombstoned row the gate
			// ignores): the marked row stays, blocks allocation, and is settled only when its bytes are observed and
			// deleted (sweepAvatars). Fail closed.
			s.logAvatarStep(rec, stageCompensate, false, reasonCompensationFailed, 1)
			return AccountProfileView{}, fail(CodeUnavailable, reasonCompensationFailed)
		}
		// Settled: nothing was issued, the provider refused the Put, or the Put returned and its bytes were observed.
		if !issued || perr != nil {
			s.releaseRefusedUpload(ctx, tc, aid) // no object can exist: an unsettled deletion must not leave a marked row
		}
		return AccountProfileView{}, fail(CodeUnavailable, compensationReason(s.compensateAvatar(ctx, tc, rec, op, aid)))
	}
	if s.at(op, stageVerify) {
		return AccountProfileView{}, errSimulatedCrash
	}

	// S3 (V): the stored length and binding.
	rec.stage = stageVerify
	vctx, cancelV := context.WithTimeout(ctx, providerControlTimeout)
	pinfo, herr := s.objects.Metadata(vctx, key)
	cancelV()
	switch {
	case herr != nil:
		s.compensateAvatar(ctx, tc, rec, op, aid)
		return AccountProfileView{}, fail(CodeUnavailable, reasonProviderUnavailable)
	case pinfo.Binding != binding:
		s.compensateAvatar(ctx, tc, rec, op, aid)
		return AccountProfileView{}, fail(CodeInternal, reasonBindingMismatch)
	case pinfo.Size != n:
		s.compensateAvatar(ctx, tc, rec, op, aid)
		return AccountProfileView{}, fail(CodeUnavailable, reasonIntegrityMismatch)
	}
	if s.at(op, stageT2) {
		return AccountProfileView{}, errSimulatedCrash
	}

	// S4 (T2): re-authorize, re-check the version, publish, move the reference, then withdraw the old avatar (I-CUR).
	rec.stage = stageT2
	var view AccountProfileView
	var old tenancy.AvatarID
	f = s.storageTx(ctx, true, dbc, func(ctx context.Context, tx *trackedTx) *failure {
		old = tenancy.AvatarID{}
		if _, f := s.avatarAuthorize(ctx, tx, tc, capAccountProfileManageSelf, true); f != nil {
			return f
		}
		p, present, f := lockAvatarProfile(ctx, tx, tc.subject, false)
		if f != nil {
			return f
		}
		if !present || !avatarVersionOK(p, present, expected) {
			return versionConflict()
		}
		if _, err := tenancy.FinalizeAvatarTx(ctx, tx, tc.subject, aid, sum); err != nil {
			if errors.Is(err, tenancy.ErrAvatarAbsent) {
				return fail(CodeUnavailable, reasonObjectNotReady) // withdrawn meanwhile: never published
			}
			return avatarFailure(ctx, tx, err)
		}
		if _, err := tenancy.SetAvatarReferenceTx(ctx, tx, tc.subject, p.Version, &aid); err != nil {
			if errors.Is(err, tenancy.ErrProfileAbsent) {
				return versionConflict()
			}
			return avatarFailure(ctx, tx, err)
		}
		if !zero(p.AvatarID) {
			if _, err := tenancy.MarkAvatarDeletingTx(ctx, tx, tc.subject, p.AvatarID); err != nil {
				return avatarFailure(ctx, tx, err)
			}
			old = p.AvatarID
		}
		var f2 *failure
		view, f2 = readOwn(ctx, tx, tc.subject)
		return f2
	})
	switch {
	case f == nil:
		rec.stage = ""
		if !s.cleanupWithdrawn(ctx, tc, rec, op, old) {
			return AccountProfileView{}, errSimulatedCrash
		}
		return view, nil
	case f.reason != reasonCommitOutcomeUnknown:
		// Definitive: the new row is not published. A T2 policy denial is not a new access.denied (S0 and S1 allowed).
		f.denied = nil
		s.compensateAvatar(ctx, tc, rec, op, aid)
		return AccountProfileView{}, f
	}
	if s.at(op, stageReconcile) {
		return AccountProfileView{}, errSimulatedCrash
	}

	// R: one locked read, detached from the caller's cancellation; compensate only on an authoritative not-current row.
	rec.stage = stageReconcile
	var cur tenancy.Avatar
	var ref tenancy.AvatarProfile
	rf := s.storageTx(context.WithoutCancel(ctx), true, dbc, func(ctx context.Context, tx *trackedTx) *failure {
		p, present, f := lockAvatarProfile(ctx, tx, tc.subject, false)
		if f != nil {
			return f
		}
		if !present {
			return fail(CodeUnavailable, reasonFactInconsistent)
		}
		a, err := tenancy.ReadAvatarTx(ctx, tx, tc.subject, aid)
		if err != nil {
			if errors.Is(err, tenancy.ErrAvatarAbsent) {
				return fail(CodeUnavailable, reasonFactInconsistent) // never proof of non-commit
			}
			return avatarFailure(ctx, tx, err)
		}
		v, f := readOwn(ctx, tx, tc.subject)
		if f != nil {
			return f
		}
		ref, cur, view = p, a, v
		return nil
	})
	switch {
	case rf != nil:
		// The read failed: the row may be current. No provider call; the row and its bytes stay as recorded.
		s.logAvatarStep(rec, stageReconcile, false, reasonReconcileUnknown, 1)
		s.record(ctx, dbContext{user: tc.subject}, audit.Event{Action: audit.ObjectUploadFailed, Outcome: audit.Unknown,
			Reason: audit.ReasonNotCompensated, ResourceType: audit.ResourceStoredObject, ResourceID: audit.UUID(aid)})
		*hold = avatarHoldRetry
		return AccountProfileView{}, &failure{code: CodeUnavailable, field: fieldCommitOutcomeUnknown, reason: reasonCommitOutcomeUnknown}
	case ref.AvatarID == aid && cur.State == tenancy.ObjectReady && cur.HasSHA256 && cur.SHA256 == sum:
		rec.stage, rec.reason = "", reasonReconciledReady
		if !s.cleanupWithdrawn(ctx, tc, rec, op, old) {
			return AccountProfileView{}, errSimulatedCrash
		}
		return view, nil
	case ref.AvatarID != aid && (cur.State == tenancy.ObjectPending || cur.State == tenancy.ObjectDeleting):
		// Authoritatively not current (an in-flight T2 would have held the lock): compensate.
		return AccountProfileView{}, fail(CodeUnavailable, compensationReason(s.compensateAvatar(ctx, tc, rec, op, aid)))
	}
	// Any other state proves nothing: retain, as for a failed read.
	s.logAvatarStep(rec, stageReconcile, false, reasonReconcileUnknown, 1)
	*hold = avatarHoldRetry
	return AccountProfileView{}, &failure{code: CodeUnavailable, field: fieldCommitOutcomeUnknown, reason: reasonCommitOutcomeUnknown}
}

// --- remove ----------------------------------------------------------------------------------------------------------

// RemoveAccountAvatar withdraws the caller's current avatar (account_profile.manage_self) when its profile is still at
// expected: R1 clears the reference (avatar_version and version +1, profile.updated avatar_removed) and then marks the
// old row deleting in the same transaction (I-CUR; the withdrawal is durable at this COMMIT); R2/R3 delete the bytes,
// confirm their absence and tombstone the row, together with every other own unresolved row (the reconcile). No current
// avatar is a successful no-op at the current version (it still repairs own unresolved rows). A failed R2/R3 is recorded
// as deletion_pending; the returned view shows no avatar regardless, and the next own avatar operation completes it.
func (s *Service) RemoveAccountAvatar(ctx context.Context, tc *TenantContext, expected int32) (AccountProfileView, error) {
	rec := &avatarRecord{op: opProfileAvatarRemove}
	v, f := s.removeAvatar(ctx, tc, expected, rec)
	s.recordGuarded(ctx, tc, rec.op, f)
	s.logAvatar(rec, f)
	if f != nil {
		return AccountProfileView{}, avatarError(f, 0)
	}
	return v, nil
}

func (s *Service) removeAvatar(ctx context.Context, tc *TenantContext, expected int32, rec *avatarRecord) (AccountProfileView, *failure) {
	const op = opProfileAvatarRemove
	if f := s.beginAvatar(ctx, tc, rec); f != nil {
		return AccountProfileView{}, f
	}
	if expected < 0 {
		return AccountProfileView{}, invalid("expected_version")
	}
	if s.at(op, stageT1) {
		return AccountProfileView{}, errSimulatedCrash
	}
	rec.stage = stageT1
	var view AccountProfileView
	var withdrawn tenancy.AvatarID
	present := false
	f := s.storageTx(ctx, true, dbcOf(tc), func(ctx context.Context, tx *trackedTx) *failure {
		withdrawn = tenancy.AvatarID{}
		if _, f := s.avatarAuthorize(ctx, tx, tc, capAccountProfileManageSelf, true); f != nil {
			return f
		}
		p, ok, f := lockAvatarProfile(ctx, tx, tc.subject, false)
		if f != nil {
			return f
		}
		present = ok
		if ok && !zero(p.AvatarID) {
			if !avatarVersionOK(p, ok, expected) {
				return versionConflict()
			}
			if _, err := tenancy.SetAvatarReferenceTx(ctx, tx, tc.subject, p.Version, nil); err != nil {
				if errors.Is(err, tenancy.ErrProfileAbsent) {
					return versionConflict()
				}
				return avatarFailure(ctx, tx, err)
			}
			if _, err := tenancy.MarkAvatarDeletingTx(ctx, tx, tc.subject, p.AvatarID); err != nil {
				return avatarFailure(ctx, tx, err)
			}
			withdrawn = p.AvatarID
		}
		var f2 *failure
		view, f2 = readOwn(ctx, tx, tc.subject)
		return f2
	})
	if f != nil {
		return AccountProfileView{}, f // an unknown COMMIT may have withdrawn it; a repeat remove completes it
	}
	rec.object = withdrawn
	rec.stage = ""
	if !present {
		return view, nil // no profile row: no avatar row can exist
	}
	if s.at(op, stageAvatarCleanup) {
		return AccountProfileView{}, errSimulatedCrash
	}
	// R2/R3 for the withdrawn row and every other own unresolved row (young pending rows excepted).
	if !s.sweepAvatars(ctx, tc, rec, op, false) {
		return AccountProfileView{}, errSimulatedCrash
	}
	return view, nil
}

// --- delivery --------------------------------------------------------------------------------------------------------

// GetAccountAvatar returns the bytes of subject's current ready avatar when version is its current avatar version
// (account_profile.read), decided now on every read: the caller itself, the current owner/admin of the active
// Organization when the subject holds a binding in it, or a current member of a shared Group of the active Organization
// (app.rls_account_profile). Every other subject, an unknown or stale version and a subject without a current ready avatar
// are the same not_found, with no provider call. The bytes are fetched from the exact key under the transfer deadline,
// their binding (user-id = subject, object-id = the row, no Organization) and length verified, and the whole body (at most
// 512 KiB) read and its SHA-256 compared before anything is returned (a mismatch is internal, never truncated bytes).
// Missing bytes under a ready row are not_found. No signed URL is ever issued for an avatar.
func (s *Service) GetAccountAvatar(ctx context.Context, tc *TenantContext, subject tenancy.UserID, version int32) (AccountAvatar, []byte, error) {
	rec := &avatarRecord{op: opProfileAvatarGet}
	a, data, f := s.getAvatar(ctx, tc, subject, version, rec)
	s.recordGuarded(ctx, tc, rec.op, f)
	s.logAvatar(rec, f)
	if f != nil {
		return AccountAvatar{}, nil, avatarError(f, 0)
	}
	return a, data, nil
}

func (s *Service) getAvatar(ctx context.Context, tc *TenantContext, subject tenancy.UserID, version int32,
	rec *avatarRecord) (AccountAvatar, []byte, *failure) {
	if f := s.beginAvatar(ctx, tc, rec); f != nil {
		return AccountAvatar{}, nil, f
	}
	if zero(subject) {
		return AccountAvatar{}, nil, invalid(string(tenancy.CodeInvalidID))
	}
	var va tenancy.VisibleAvatar
	f := s.storageTx(ctx, false, dbcOf(tc), func(ctx context.Context, tx *trackedTx) *failure {
		facts, f := s.avatarAuthorize(ctx, tx, tc, capAccountProfileRead, false)
		if f != nil {
			return f
		}
		var err error
		va, err = tenancy.VisibleAvatarTx(ctx, tx, subject)
		switch {
		case errors.Is(err, tenancy.ErrProfileAbsent):
			return markDenied(notFound(), tc.selection(), facts, "", audit.UUID{}, tenancy.GroupID{})
		case errors.Is(err, tenancy.ErrAvatarAbsent):
			return fail(CodeNotFound, reasonObjectNotReady)
		case err != nil:
			return avatarFailure(ctx, tx, err)
		case version < 1 || va.AvatarVersion != version:
			return fail(CodeNotFound, reasonObjectNotReady)
		}
		return nil
	})
	if f != nil {
		return AccountAvatar{}, nil, f
	}
	rec.object = va.ID
	key, ok := avatarKeyOf(subject, va.ID)
	if !ok || va.ByteSize < 1 || va.ByteSize > branding.MaxLogoBytes {
		return AccountAvatar{}, nil, fail(CodeInternal, reasonBindingMismatch)
	}
	tctx, cancel := context.WithTimeout(ctx, s.ocfg.TransferTimeout)
	defer cancel()
	pinfo, rc, err := s.objects.Get(tctx, key)
	switch {
	case errors.Is(err, objectstore.ErrNotFound):
		return AccountAvatar{}, nil, fail(CodeNotFound, reasonObjectBytesMissing)
	case err != nil:
		return AccountAvatar{}, nil, fail(CodeUnavailable, reasonProviderUnavailable)
	}
	defer func() { _ = rc.Close() }()
	if pinfo.Binding != avatarBinding(subject, va.ID) || pinfo.Size != va.ByteSize {
		return AccountAvatar{}, nil, fail(CodeInternal, reasonBindingMismatch)
	}
	data, err := io.ReadAll(io.LimitReader(rc, va.ByteSize+1))
	switch {
	case err != nil:
		return AccountAvatar{}, nil, fail(CodeUnavailable, reasonProviderUnavailable)
	case int64(len(data)) != va.ByteSize || sha256.Sum256(data) != va.SHA256:
		return AccountAvatar{}, nil, fail(CodeInternal, reasonIntegrityMismatch)
	}
	return AccountAvatar{Version: va.AvatarVersion, ContentType: va.ContentType}, data, nil
}
