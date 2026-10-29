package access

import (
	"bytes"
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
	"core/internal/tenancy"
)

// G019 ordinary retirement (Plan r1 §1.4 "Retirement"; C3, C6). Eligibility is a self read in the IDENTITY-shape
// context: the definer reader app.rls_sole_owner_scopes returns only the subject's own sole-owned scope identifiers,
// never another tenant's members. It is a preflight, not a promise: each departure re-locks and rechecks.

// ScopeRef identifies one scope by identifiers only.
type ScopeRef = tenancy.ScopeRef

// MaxRetirementBlockers bounds the scopes a RetirementBlockers carries; Count is always the full number.
const MaxRetirementBlockers = 100

// RetirementBlockers are the subject's sole-owned scopes: retirement is refused while Count is not zero.
type RetirementBlockers struct {
	Count     int
	Scopes    []ScopeRef
	Truncated bool
}

// RetirementEligibility reads the session subject's retirement blockers (no capability; IDENTITY shape, no tenant
// context): the credential is resolved at its bound origin as BeginRequest does, then, in one read transaction in the
// subject's IDENTITY-shape context, the session liveness is rechecked and the sole-owned scopes are read.
func (s *Service) RetirementEligibility(ctx context.Context, credential, origin string) (RetirementBlockers, error) {
	var out RetirementBlockers
	f := func() *failure {
		info, err := s.id.Session(ctx, credential, origin)
		switch {
		case errors.Is(err, identity.ErrUnauthenticated):
			return fail(CodeUnauthenticated, reasonSessionInvalid)
		case err != nil:
			return fail(CodeUnavailable, failureReason(ctx, nil))
		}
		subject, err := tenancy.ParseUserID(info.UserID)
		if err != nil || zero(subject) || info.SessionID == "" {
			return fail(CodeUnavailable, reasonStoreUnavailable)
		}
		return s.inContextTx(ctx, false, dbContext{user: subject}, cleanupUnbounded, func(tx *trackedTx) *failure {
			if err := s.id.CheckLiveSession(ctx, tx, info.SessionID, info.UserID); err != nil {
				return sessionFailure(ctx, tx, err)
			}
			scopes, err := tenancy.SoleOwnerScopes(ctx, tx)
			if err != nil {
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			}
			out = blockers(scopes)
			return nil
		})
	}()
	s.logDecision(ctx, opRetirementEligibility, noCapability, outcomeAllow, reasonAllowed, f)
	if f != nil {
		return RetirementBlockers{}, f.err()
	}
	return out, nil
}

// retirementFinalizer is identity's trusted, route-less retirement finalization (identity.Service.FinalizeRetirement,
// Result F-01): the subject's users row locked FOR UPDATE, the check run once on that transaction, then the user disabled
// and its sessions revoked in the same transaction — or nothing written when the check refuses. Reached through the
// Service's identity boundary; the emergency identity.DisableUser is not used by ordinary retirement.
type retirementFinalizer interface {
	FinalizeRetirement(ctx context.Context, userID string, check identity.RetirementCheck) error
}

// errRetirementRemaining is the final check's refusal: under the finalization lock the subject still holds a
// membership or solely owns a scope. identity returns it unchanged and writes nothing.
var errRetirementRemaining = errors.New("access: retirement remaining")

// errRetirementOrder reports a non-ascending own-Organization page (never produced by the keyset reader).
var errRetirementOrder = errors.New("access: own-Organization pages out of order")

// remainingReader reads the subject's sole-owned scopes and every own-Organization page on the given transaction.
type remainingReader func(ctx context.Context, tx pgx.Tx) ([]ScopeRef, []tenancy.OrganizationID, error)

// readRemaining is the production remainingReader: the existing self-only definer readers, every page (no cap).
func readRemaining(ctx context.Context, tx pgx.Tx) ([]ScopeRef, []tenancy.OrganizationID, error) {
	scopes, err := tenancy.SoleOwnerScopes(ctx, tx)
	if err != nil {
		return nil, nil, err
	}
	orgs, err := ownOrganizations(ctx, tx)
	if err != nil {
		return nil, nil, err
	}
	return scopes, orgs, nil
}

// ownOrganizations reads every page of the context subject's own Organizations in ascending keyset order; a
// non-ascending page fails closed (errRetirementOrder) instead of looping. The result is never nil.
func ownOrganizations(ctx context.Context, tx pgx.Tx) ([]tenancy.OrganizationID, error) {
	orgs := []tenancy.OrganizationID{}
	var after tenancy.OrganizationID
	for {
		page, err := tenancy.OwnOrganizationsPage(ctx, tx, after, retirementPage)
		if err != nil {
			return nil, err
		}
		for _, o := range page {
			if !organizationAfter(o, after) {
				return nil, errRetirementOrder
			}
			after = o
		}
		orgs = append(orgs, page...)
		if len(page) < retirementPage {
			return orgs, nil
		}
	}
}

// retirementPage is the page size of the trusted-retirement enumeration (app.rls_own_organizations admits 1..100).
const retirementPage = tenancy.MaxMembersPage

// RetirementProgress is RetireUser's honest, resumable progress (Plan r1 §1.4 "Retirement", C3; Result addendum
// A1-01). Nothing is ever undone: a re-run enumerates the remaining memberships again and continues.
type RetirementProgress struct {
	// Blockers are the preflight's sole-owned scopes; a non-zero Count refuses before any write.
	Blockers RetirementBlockers
	// Departed are the Organizations this call departed, in ascending id.
	Departed []tenancy.OrganizationID
	// AlreadyDeparted were enumerated but held no membership of the subject when their departure ran (a concurrent
	// removal); this call did not depart them.
	AlreadyDeparted []tenancy.OrganizationID
	// Remaining were not departed when the call stopped: the refusing or failing Organization first, then the rest.
	Remaining []tenancy.OrganizationID
	// Disabled reports that the user was disabled, which happens only after every departure succeeded.
	Disabled bool
}

// RetireUser is the trusted ordinary retirement (no route, no capability; an operator path, never a tenant
// administrator's): (1) one IDENTITY-shape read transaction reads the subject's sole-owned scopes and every page of its
// own Organizations (app.rls_own_organizations, ascending keyset pages; no cap); blockers refuse (conflict,
// retirement_blocked) before any write; (2) for each Organization in ascending id one guarded departure transaction in
// the subject's own Organization-only context rechecks the locked facts and departs the subject's Group bindings, then
// its Organization binding (the 00015 guard re-enforces owner authority, dependent bindings and the last owner); the
// first refusal or failure stops the call with the user still enabled; (3) only after every departure, identity's
// finalization (Result F-01) locks the user row FOR UPDATE, rereads the sole-owned scopes and every own-Organization
// page on that transaction, and disables the user (sessions revoked) only when none remains; a grant or rejoin that
// committed after the enumeration refuses (conflict, retirement_blocked) with the user still enabled, Remaining = the
// complete locked list and Blockers = the locked scopes. It is not the security-suspension shortcut: no owner is
// disabled with ownership retained through this path. A failed or unconfirmed finalization (unknown COMMIT) reports
// unavailable with Disabled false and leaves every departure in place; a re-run enumerates again and finalizes. One
// tenancy.operation record (retirement.retire) is written per call.
func (s *Service) RetireUser(ctx context.Context, user tenancy.UserID) (RetirementProgress, error) {
	p := RetirementProgress{Blockers: blockers(nil), Departed: []tenancy.OrganizationID{},
		AlreadyDeparted: []tenancy.OrganizationID{}, Remaining: []tenancy.OrganizationID{}}
	f := s.retire(ctx, user, &p)
	logRetirement(ctx, s.logger, f)
	if f != nil {
		return p, f.err()
	}
	return p, nil
}

// retire is RetireUser's body; p carries the progress of every outcome.
func (s *Service) retire(ctx context.Context, user tenancy.UserID, p *RetirementProgress) *failure {
	if f := targetID(user)(); f != nil {
		return f
	}
	// The finalization is checked before any write: a Service whose identity boundary cannot finalize never departs.
	fin, ok := s.id.(retirementFinalizer)
	if !ok {
		return fail(CodeUnavailable, reasonStoreUnavailable)
	}
	orgs, f := s.retirementPreflight(ctx, user, p)
	if f != nil {
		return f
	}
	for i, org := range orgs {
		gone, f := s.retireDepart(ctx, user, org)
		if f != nil {
			p.Remaining = append(p.Remaining, orgs[i:]...)
			return f
		}
		if gone {
			p.AlreadyDeparted = append(p.AlreadyDeparted, org)
		} else {
			p.Departed = append(p.Departed, org)
		}
	}
	return finalizeRetirement(ctx, fin, user, readRemaining, p)
}

// finalizeRetirement runs identity's finalization with read as the check and applies its outcome to p: a final
// refusal is conflict/retirement_blocked with the locked scopes and the complete locked Organization list; any other
// error, or a success without exactly one check call, is unavailable with Disabled false (never claimed unconfirmed).
// The check only reads: it keeps no reference to the transaction after it returns and leaves its context untouched.
func finalizeRetirement(ctx context.Context, fin retirementFinalizer, user tenancy.UserID, read remainingReader, p *RetirementProgress) *failure {
	var scopes []ScopeRef
	var orgs []tenancy.OrganizationID
	calls := 0
	err := fin.FinalizeRetirement(ctx, user.String(), func(ctx context.Context, tx pgx.Tx) error {
		calls++
		sc, og, err := read(ctx, tx)
		if err != nil {
			return err
		}
		scopes, orgs = sc, og
		if len(sc) > 0 || len(og) > 0 {
			return errRetirementRemaining
		}
		return nil
	})
	switch {
	case errors.Is(err, errRetirementRemaining) && calls == 1:
		p.Blockers = blockers(scopes)
		p.Remaining = append(p.Remaining, orgs...)
		return &failure{code: CodeConflict, field: fieldRetirementBlocked, reason: reasonWriteConflict}
	case errors.Is(err, errRetirementOrder):
		return fail(CodeInternal, reasonFactInconsistent)
	case err != nil || calls != 1:
		return fail(CodeUnavailable, failureReason(ctx, nil))
	}
	p.Disabled = true
	return nil
}

// logRetirement writes the call's one retirement.retire record: no code on success, the refusal field for a conflict
// (retirement_blocked or the departure's tenancy code), otherwise the closed access code.
func logRetirement(ctx context.Context, logger *slog.Logger, f *failure) {
	var code tenancy.Code
	if f != nil {
		code = tenancy.Code(f.field)
		if code == "" || f.code == CodeUnavailable || f.code == CodeInternal {
			code = tenancy.Code(f.code)
		}
	}
	tenancy.LogRetirement(ctx, logger, code)
}

// retirementPreflight reads, in one IDENTITY-shape READ ONLY snapshot of the subject, its sole-owned scopes and every
// page of its own Organizations. A non-ascending page (never produced by the keyset reader) fails closed instead of
// looping. Blockers refuse with every enumerated Organization remaining.
func (s *Service) retirementPreflight(ctx context.Context, user tenancy.UserID, p *RetirementProgress) ([]tenancy.OrganizationID, *failure) {
	var orgs []tenancy.OrganizationID
	f := s.inContextTx(ctx, false, dbContext{user: user}, cleanupUnbounded, func(tx *trackedTx) *failure {
		scopes, err := tenancy.SoleOwnerScopes(ctx, tx)
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		p.Blockers = blockers(scopes)
		var err2 error
		if orgs, err2 = ownOrganizations(ctx, tx); err2 != nil {
			if errors.Is(err2, errRetirementOrder) {
				return fail(CodeInternal, reasonFactInconsistent)
			}
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		return nil
	})
	if f != nil {
		return nil, f
	}
	if p.Blockers.Count > 0 {
		p.Remaining = append(p.Remaining, orgs...)
		return nil, &failure{code: CodeConflict, field: fieldRetirementBlocked, reason: reasonWriteConflict}
	}
	return orgs, nil
}

// organizationAfter reports o > after in the reader's uuid order (byte order); the zero cursor precedes every id.
func organizationAfter(o, after tenancy.OrganizationID) bool {
	return bytes.Compare(o[:], after[:]) > 0
}

// retireDepart departs one Organization in the subject's own Organization-only context (the LeaveScope sequence,
// without a session or capability: the trusted caller established the context). gone reports a membership that no
// longer exists, which is not a departure by this call.
func (s *Service) retireDepart(ctx context.Context, user tenancy.UserID, org tenancy.OrganizationID) (gone bool, f *failure) {
	sc := Scope{Kind: tenancy.ScopeOrganization}
	q := tenancy.FactQuery{Organization: org, User: user, Lock: true, Groups: []tenancy.GroupID{}, ExclusiveOrganization: true,
		Subject: &user, OwnerCount: true}
	f = s.inContextTx(ctx, true, dbContext{user: user, organization: org}, cleanupUnbounded, func(tx *trackedTx) *failure {
		gone = false
		facts, err := tenancy.ReadFacts(ctx, tx, q)
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		b, ok := subjectBinding(facts, sc, user)
		if zero(facts.Organization.ID) || !ok {
			gone = true
			return nil
		}
		if b.Role == sc.ownerRole() && facts.OwnerCount <= 1 {
			return conflict(string(tenancy.CodeLastOwner))
		}
		// The guard removes the subject's Group bindings inside this DELETE (00017 amendment01).
		if err := tenancy.RemoveBindingTx(ctx, tx, b); err != nil {
			return membershipFailure(ctx, tx, err)
		}
		return nil
	})
	return gone, f
}

// blockers bounds a sole-owned scope list.
func blockers(scopes []ScopeRef) RetirementBlockers {
	b := RetirementBlockers{Count: len(scopes), Scopes: scopes}
	if len(scopes) > MaxRetirementBlockers {
		b.Scopes, b.Truncated = scopes[:MaxRetirementBlockers], true
	}
	if b.Scopes == nil {
		b.Scopes = []ScopeRef{}
	}
	return b
}
