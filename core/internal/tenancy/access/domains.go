package access

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/domains"
	"core/internal/domains/db"
	"core/internal/hostname"
	"core/internal/tenancy"
)

// G016 guarded domain operations (Plan r4 §1.4, §1.5, §5.0.1). Each authorizes and acts in one transaction through
// guard, like every facade operation: liveness, the Organization facts (FOR SHARE for writes), the context check and the
// decision, then the domain row FOR NO KEY UPDATE (lock order user -> session -> Organization -> domain row) and one
// version-guarded statement. The Service holds no resolver, checker or TLS adapter: every DNS/TLS step runs in
// domains/api between two facade calls (BeginDomainCheck, then RecordDomainCheck), so no network call happens under a
// facade lock. Domain operations and the slug need organization.manage_domains (Organization owner or admin); the slug
// write passes the owner/admin organizations_update_slug policy, whose guard confines an admin to the slug.

// domainHooks are inert ordering seams: nil in every production construction; set only by this package's tests.
type domainHooks struct {
	claimLocked, checkLocked, removeLocked func()
	generation                             func(attempt int, drawn int64) int64 // replaces the drawn claim fence
}

func (s *Service) hook(f func(*domainHooks) func()) {
	if s.domainHooks != nil {
		if h := f(s.domainHooks); h != nil {
			h()
		}
	}
}

// The domain operations' logged names (and access.denied operations).
const (
	opListDomains      = "domain.list"
	opClaimDomain      = "domain.claim"
	opRotateChallenge  = "domain.rotate"
	opVerifyDomain     = "domain.verify"
	opRecheckDomain    = "domain.recheck"
	opRemoveDomain     = "domain.remove"
	opRetryWithdrawal  = "domain.retry_withdrawal"
	opSetSlug          = "organization.set_slug"
	challengeBytes     = 32
	maxChallengeTTL    = 30 * 24 * time.Hour
	maxProofFreshness  = 90 * 24 * time.Hour
	maxWithdrawalRetry = 20
)

var challengeEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// DomainView is a domain row as the setup surface sees it. The challenge value appears only in DomainClaim.
type DomainView struct {
	ID               string
	Hostname         string
	Generation       int64
	State            domains.State
	VerifiedUntil    *time.Time
	TLSState         domains.TLSState
	LastCheckAt      *time.Time
	LastCheckOutcome string
	CheckFailures    int32
	CreatedAt        time.Time
	RemovedAt        *time.Time
}

// DomainClaim is a claim's view plus the DNS records the customer publishes: a TXT record named ChallengeName with
// ChallengeValue (a public record value, not a credential; never logged) and a CNAME to the platform's target.
type DomainClaim struct {
	DomainView
	ChallengeName      string
	ChallengeValue     string
	ChallengeExpiresAt time.Time
}

func viewOf(r domains.Record) DomainView {
	v := DomainView{ID: r.ID, Hostname: r.Hostname, Generation: r.Generation, State: r.State, VerifiedUntil: r.VerifiedUntil,
		TLSState: r.TLSState, LastCheckAt: r.LastCheckAt, CheckFailures: r.CheckFailures, CreatedAt: r.CreatedAt, RemovedAt: r.RemovedAt}
	if r.LastCheckOutcome != nil {
		v.LastCheckOutcome = *r.LastCheckOutcome
	}
	return v
}

func recordOf(r db.GetDomainRow) domains.Record {
	return domains.Record{ID: r.ID, OrganizationID: r.OrganizationID, Hostname: r.Hostname, Generation: r.Generation,
		State: domains.State(r.State), ChallengeValue: r.ChallengeValue, ChallengeExpiresAt: r.ChallengeExpiresAt,
		VerifiedAt: r.VerifiedAt, VerifiedUntil: r.VerifiedUntil, LastCheckAt: r.LastCheckAt, LastCheckOutcome: r.LastCheckOutcome,
		CheckFailures: r.CheckFailures, TLSState: domains.TLSState(r.TlsState), TLSCheckedAt: r.TlsCheckedAt,
		TLSNotAfter: r.TlsNotAfter, TLSReference: r.TlsReference, RemovedAt: r.RemovedAt, CreatedAt: r.CreatedAt, Version: r.Version}
}

// newChallenge returns a fresh 32-byte random challenge value, lowercase base32 without padding (52 characters).
func newChallenge() (string, error) {
	b := make([]byte, challengeBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.ToLower(challengeEncoding.EncodeToString(b)), nil
}

// newGeneration returns an opaque claim fence drawn uniformly from 1..MaxInt64 by the CSPRNG (deviation #3 closure: it
// is compared only for equality with the row id and carries no history). A randomness failure fails closed.
func newGeneration() (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64)) // [0, MaxInt64)
	if err != nil {
		return 0, err
	}
	return n.Int64() + 1, nil
}

func newDomainID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	var id tenancy.OrganizationID
	copy(id[:], b)
	return id.String(), nil
}

func domainInput(field string) *failure {
	return &failure{code: CodeInvalidInput, field: field, reason: reasonInputInvalid}
}

// domainFailure maps an error of a domain statement: a lost version race or a guard refusal is domain_state_invalid, a
// proven-hostname collision (a lost promotion: G016 security amendment, organization_domains_hostname_owned)
// host_taken, a slug collision conflict, a constraint refusal invalid_input, a missing row
// not_found; a store failure is unavailable with the transaction's reason.
func domainFailure(ctx context.Context, tx *trackedTx, err error, missing Code) *failure {
	var pg *pgconn.PgError
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return fail(missing, reasonTargetNotFound)
	case errors.As(err, &pg) && pg.ConstraintName == "organization_domains_guard":
		return fail(CodeDomainStateInvalid, reasonWriteConflict)
	case errors.As(err, &pg) && pg.Code == "23505" && pg.ConstraintName == "organization_domains_hostname_owned":
		return fail(CodeHostTaken, reasonWriteConflict)
	case errors.As(err, &pg) && pg.Code == "23505":
		return fail(CodeConflict, reasonWriteConflict)
	case errors.As(err, &pg) && (pg.Code == "23514" || pg.Code == "22P02"):
		return domainInput(string(tenancy.CodeInvalidSlug))
	case errors.As(err, &pg) && commandError(pg) && len(pg.Code) == 5 && pg.Code[:2] != "40" && pg.Code[:2] != "08" &&
		pg.Code[:2] != "57":
		return fail(CodeInternal, reasonStoreUnavailable)
	}
	return fail(CodeUnavailable, failureReason(ctx, tx.cause))
}

func manageDomains(s *Service, ctx context.Context, tc *TenantContext, op string, write bool, pre func() *failure,
	act func(tx *trackedTx, q *db.Queries, org string) *failure) error {
	return s.guard(ctx, tc, op, capOrganizationManageDomains, write, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, f Facts) *failure { return act(tx, db.New(), f.Organization.ID.String()) })
}

// lockDomain reads the row FOR NO KEY UPDATE in the context Organization (not_found otherwise).
func lockDomain(ctx context.Context, tx *trackedTx, q *db.Queries, org, id string) (domains.Record, *failure) {
	r, err := q.GetDomainForUpdate(ctx, tx, db.GetDomainForUpdateParams{ID: id, OrganizationID: org})
	if err != nil {
		return domains.Record{}, domainFailure(ctx, tx, err, CodeNotFound)
	}
	return recordOf(db.GetDomainRow(r)), nil
}

func domainID(id string) func() *failure {
	return func() *failure {
		if _, err := tenancy.ParseOrganizationID(id); err != nil {
			return &failure{code: CodeInvalidInput, field: string(tenancy.CodeInvalidID), reason: reasonInputInvalid}
		}
		return nil
	}
}

// ListDomains lists the active Organization's domain rows (removed ones included), oldest first, at most 500.
func (s *Service) ListDomains(ctx context.Context, tc *TenantContext) ([]DomainView, error) {
	var out []DomainView
	err := manageDomains(s, ctx, tc, opListDomains, false, nil, func(tx *trackedTx, q *db.Queries, org string) *failure {
		rows, err := q.ListDomains(ctx, tx, org)
		if err != nil {
			return domainFailure(ctx, tx, err, CodeNotFound)
		}
		out = make([]DomainView, 0, len(rows))
		for _, r := range rows {
			out = append(out, viewOf(recordOf(db.GetDomainRow(r))))
		}
		return nil
	})
	return out, err
}

// ClaimDomain claims a normalized custom hostname for the active Organization as a pending row with a fresh challenge
// that expires after ttl. The caller (domains/api) has normalized the name and refused reserved and special-use names;
// the facade re-checks the shape. G016 security amendment (M-1): pending claims of one hostname coexist across
// Organizations and reserve nothing; a hostname already proven (verified, active or suspended) by anyone is host_taken
// with zero effects, and exclusivity is decided at proof time by the partial unique index on proven rows (a lost
// promotion is host_taken in RecordDomainCheck).
func (s *Service) ClaimDomain(ctx context.Context, tc *TenantContext, host string, ttl time.Duration, challengeLabel string) (DomainClaim, error) {
	var out DomainClaim
	pre := func() *failure {
		if !hostname.ValidName(host) || !strings.Contains(host, ".") || len(host) < 3 {
			return domainInput("host_invalid")
		}
		if ttl <= 0 || ttl > maxChallengeTTL {
			return domainInput("challenge_ttl_invalid")
		}
		return nil
	}
	err := manageDomains(s, ctx, tc, opClaimDomain, true, pre, func(tx *trackedTx, q *db.Queries, org string) *failure {
		value, err := newChallenge()
		if err != nil {
			return fail(CodeInternal, reasonStoreUnavailable)
		}
		id, err := newDomainID()
		if err != nil {
			return fail(CodeInternal, reasonStoreUnavailable)
		}
		s.hook(func(h *domainHooks) func() { return h.claimLocked })
		proven, err := q.HostnameProven(ctx, tx, host)
		if err != nil {
			return domainFailure(ctx, tx, err, CodeNotFound)
		}
		if proven {
			return fail(CodeHostTaken, reasonWriteConflict)
		}
		// One bounded regeneration on the fence's own (hostname, generation) collision, which the statement absorbs with
		// ON CONFLICT DO NOTHING (no aborted transaction); a second collision fails closed. Nothing else is masked.
		var r db.InsertDomainRow
		for attempt := 0; ; attempt++ {
			gen, err := newGeneration()
			if err != nil {
				return fail(CodeInternal, reasonStoreUnavailable)
			}
			if s.domainHooks != nil && s.domainHooks.generation != nil {
				gen = s.domainHooks.generation(attempt, gen)
			}
			r, err = q.InsertDomain(ctx, tx, db.InsertDomainParams{ID: id, OrganizationID: org, Hostname: host, Generation: gen,
				ChallengeValue: value, ChallengeExpiresAt: time.Now().Add(ttl), CreatedBy: tc.subject.String()})
			if errors.Is(err, pgx.ErrNoRows) && attempt == 0 {
				continue
			}
			if errors.Is(err, pgx.ErrNoRows) {
				return fail(CodeInternal, reasonStoreUnavailable)
			}
			if err != nil {
				return domainFailure(ctx, tx, err, CodeNotFound)
			}
			break
		}
		rec := recordOf(db.GetDomainRow(r))
		out = DomainClaim{DomainView: viewOf(rec), ChallengeName: challengeLabel + "." + host, ChallengeValue: rec.ChallengeValue,
			ChallengeExpiresAt: rec.ChallengeExpiresAt}
		return nil
	})
	return out, err
}

// RotateChallenge issues a new challenge value and expiry for a pending claim (pending only; the value is the ongoing
// proof once verified and is never rotated after verification).
func (s *Service) RotateChallenge(ctx context.Context, tc *TenantContext, id string, ttl time.Duration, challengeLabel string) (DomainClaim, error) {
	var out DomainClaim
	pre := func() *failure {
		if f := domainID(id)(); f != nil {
			return f
		}
		if ttl <= 0 || ttl > maxChallengeTTL {
			return domainInput("challenge_ttl_invalid")
		}
		return nil
	}
	err := manageDomains(s, ctx, tc, opRotateChallenge, true, pre, func(tx *trackedTx, q *db.Queries, org string) *failure {
		cur, f := lockDomain(ctx, tx, q, org, id)
		if f != nil {
			return f
		}
		if cur.State != domains.StatePending {
			return fail(CodeDomainStateInvalid, reasonWriteConflict)
		}
		value, err := newChallenge()
		if err != nil {
			return fail(CodeInternal, reasonStoreUnavailable)
		}
		r, err := q.RotateChallenge(ctx, tx, db.RotateChallengeParams{ChallengeValue: value,
			ChallengeExpiresAt: time.Now().Add(ttl), ID: id, Version: cur.Version})
		if err != nil {
			return domainFailure(ctx, tx, err, CodeDomainStateInvalid)
		}
		rec := recordOf(db.GetDomainRow(r))
		out = DomainClaim{DomainView: viewOf(rec), ChallengeName: challengeLabel + "." + rec.Hostname, ChallengeValue: rec.ChallengeValue,
			ChallengeExpiresAt: rec.ChallengeExpiresAt}
		return nil
	})
	return out, err
}

func checkOp(c domains.Check) string {
	if c == domains.CheckVerify {
		return opVerifyDomain
	}
	return opRecheckDomain
}

// BeginDomainCheck is the read step of a verify or recheck: it authorizes, reads the row and refuses a check that does
// not apply to its state (domain_state_invalid; an expired pending challenge too). The returned record (with its
// version, generation and challenge) feeds the network step outside every transaction.
func (s *Service) BeginDomainCheck(ctx context.Context, tc *TenantContext, id string, c domains.Check) (domains.Record, error) {
	var out domains.Record
	err := manageDomains(s, ctx, tc, checkOp(c), false, domainID(id), func(tx *trackedTx, q *db.Queries, org string) *failure {
		r, err := q.GetDomain(ctx, tx, db.GetDomainParams{ID: id, OrganizationID: org})
		if err != nil {
			return domainFailure(ctx, tx, err, CodeNotFound)
		}
		out = recordOf(r)
		if domains.Begin(out, c, time.Now()) != nil {
			return fail(CodeDomainStateInvalid, reasonWriteConflict)
		}
		return nil
	})
	return out, err
}

// RecordDomainCheck applies one observation to the row read by BeginDomainCheck, re-authorizing in its own write
// transaction. The row must still carry the version that was read (a raced change is domain_state_invalid, zero
// effects), and a recheck's TLS result must name the locked row's exact DomainRef (else domain_state_invalid, zero
// effects); the update is the pure lifecycle table's (domains.Apply) and the guard enforces the same edges.
func (s *Service) RecordDomainCheck(ctx context.Context, tc *TenantContext, read domains.Record, c domains.Check, obs domains.Observation, freshness time.Duration) (DomainView, error) {
	var out DomainView
	pre := func() *failure {
		if f := domainID(read.ID)(); f != nil {
			return f
		}
		if freshness <= 0 || freshness > maxProofFreshness {
			return domainInput("freshness_invalid")
		}
		return nil
	}
	err := manageDomains(s, ctx, tc, checkOp(c), true, pre, func(tx *trackedTx, q *db.Queries, org string) *failure {
		cur, f := lockDomain(ctx, tx, q, org, read.ID)
		if f != nil {
			return f
		}
		s.hook(func(h *domainHooks) func() { return h.checkLocked })
		if cur.Version != read.Version || cur.Generation != read.Generation {
			return fail(CodeDomainStateInvalid, reasonWriteConflict)
		}
		// The TLS observation is fenced by the locked row, not only by the read snapshot: a recheck applies an adapter
		// result only when its DomainRef is exactly the locked row's (ID, generation, hostname and reference). A result
		// for another row, an older generation or with no ref at all is stale: zero effects, no successful check.
		if c == domains.CheckRecheck && obs.Proof == domains.OutcomeOK && obs.TLS.Ref != cur.Ref() {
			return fail(CodeDomainStateInvalid, reasonWriteConflict)
		}
		u, err := domains.Apply(cur, c, obs, time.Now(), freshness)
		if err != nil {
			return fail(CodeDomainStateInvalid, reasonWriteConflict)
		}
		p := db.RecordCheckParams{State: string(u.State), VerifiedAt: u.VerifiedAt, LastCheckAt: u.LastCheckAt,
			LastCheckOutcome: string(u.LastCheckOutcome), CheckFailures: u.CheckFailures, ID: cur.ID, Version: cur.Version}
		if u.VerifiedUntil != nil {
			if u.State == domains.StateSuspended {
				p.ProofLost = true // verified_until = now() in the statement: issuance stops at this commit
			} else {
				p.VerifiedUntil = u.VerifiedUntil
			}
		}
		if u.TLS != nil {
			st := string(u.TLS.State)
			p.TlsState, p.TlsCheckedAt, p.TlsNotAfter, p.TlsReference = &st, &u.TLS.CheckedAt, u.TLS.NotAfter, u.TLS.Reference
		}
		r, err := q.RecordCheck(ctx, tx, p)
		if err != nil {
			return domainFailure(ctx, tx, err, CodeDomainStateInvalid)
		}
		out = viewOf(recordOf(db.GetDomainRow(r)))
		return nil
	})
	return out, err
}

// RemoveDomain removes a non-removed row: it stops routing and tenant-host authentication at commit and starts the
// withdrawal (tls_state withdrawing). The returned removed record's DomainRef is what the caller passes to the TLS
// adapter's Remove after the commit, outside every lock.
func (s *Service) RemoveDomain(ctx context.Context, tc *TenantContext, id string) (domains.Record, error) {
	var out domains.Record
	err := manageDomains(s, ctx, tc, opRemoveDomain, true, domainID(id), func(tx *trackedTx, q *db.Queries, org string) *failure {
		cur, f := lockDomain(ctx, tx, q, org, id)
		if f != nil {
			return f
		}
		s.hook(func(h *domainHooks) func() { return h.removeLocked })
		if cur.State == domains.StateRemoved {
			return fail(CodeDomainStateInvalid, reasonWriteConflict)
		}
		r, err := q.RemoveDomain(ctx, tx, db.RemoveDomainParams{ID: id, Version: cur.Version})
		if err != nil {
			return domainFailure(ctx, tx, err, CodeDomainStateInvalid)
		}
		out = recordOf(db.GetDomainRow(r))
		return nil
	})
	return out, err
}

// RetryWithdrawal is the explicit, bounded withdrawal retry of a removed row whose cleanup outcome is unknown or error
// (Plan r4 §1.5): it reserves one attempt (check_failures + 1) under the row lock and returns the removed record for the
// caller's post-commit adapter Remove on that row's own DomainRef. A completed withdrawal or a live row is
// domain_state_invalid; reaching max is domain_retry_exhausted. Zero effects on refusal.
func (s *Service) RetryWithdrawal(ctx context.Context, tc *TenantContext, id string, max int) (domains.Record, error) {
	var out domains.Record
	pre := func() *failure {
		if f := domainID(id)(); f != nil {
			return f
		}
		if max < 1 || max > maxWithdrawalRetry {
			return domainInput("retry_max_invalid")
		}
		return nil
	}
	err := manageDomains(s, ctx, tc, opRetryWithdrawal, true, pre, func(tx *trackedTx, q *db.Queries, org string) *failure {
		cur, f := lockDomain(ctx, tx, q, org, id)
		if f != nil {
			return f
		}
		ok, err := domains.RetryAllowed(cur, max)
		switch {
		case err != nil:
			return fail(CodeDomainStateInvalid, reasonWriteConflict)
		case !ok:
			return fail(CodeDomainRetryExhausted, reasonWriteConflict)
		}
		r, err := q.ReserveWithdrawalRetry(ctx, tx, db.ReserveWithdrawalRetryParams{ID: id, Version: cur.Version})
		if err != nil {
			return domainFailure(ctx, tx, err, CodeDomainStateInvalid)
		}
		out = recordOf(db.GetDomainRow(r))
		return nil
	})
	return out, err
}

// RecordWithdrawal records the adapter's Remove outcome on the removed row it was called for: the same id, generation
// and version (anything else is a stale result: domain_state_invalid, nothing changes). Only the cleanup columns
// change; routing and issuance stay false whatever the outcome. reserved is true after RetryWithdrawal (the attempt is
// already counted); otherwise every non-withdrawn outcome counts one failure.
func (s *Service) RecordWithdrawal(ctx context.Context, tc *TenantContext, removed domains.Record, res domains.TLSResult, reserved bool) (DomainView, error) {
	var out DomainView
	op := opRemoveDomain
	if reserved {
		op = opRetryWithdrawal
	}
	err := manageDomains(s, ctx, tc, op, true, domainID(removed.ID), func(tx *trackedTx, q *db.Queries, org string) *failure {
		if res.Ref.ID != removed.ID || res.Ref.Generation != removed.Generation {
			return fail(CodeDomainStateInvalid, reasonWriteConflict)
		}
		cur, f := lockDomain(ctx, tx, q, org, removed.ID)
		if f != nil {
			return f
		}
		if cur.State != domains.StateRemoved || cur.Version != removed.Version || cur.Generation != removed.Generation {
			return fail(CodeDomainStateInvalid, reasonWriteConflict)
		}
		state, outcome := domains.RemovalOutcome(res)
		inc := int32(0)
		if !reserved && state != domains.TLSWithdrawn {
			inc = 1
		}
		r, err := q.RecordWithdrawal(ctx, tx, db.RecordWithdrawalParams{TlsState: string(state), LastCheckOutcome: string(outcome),
			FailureIncrement: inc, ID: cur.ID, Generation: cur.Generation, Version: cur.Version})
		if err != nil {
			return domainFailure(ctx, tx, err, CodeDomainStateInvalid)
		}
		out = viewOf(recordOf(db.GetDomainRow(r)))
		return nil
	})
	return out, err
}

// SetOrganizationSlug sets (or, with "", clears) the active Organization's platform hostname label: the Organization
// slug grammar, not in the reserved floor (hostname.ReservedLabels; the caller refuses configured labels first).
// Owner or admin (organization.manage_domains; the organizations_update_slug policy and the slug guard confine an admin
// to the slug), with the Organization row locked FOR NO KEY UPDATE up front; a collision is
// conflict. Setting the current value is a successful no-op. The slug selects a host; it never grants anything.
func (s *Service) SetOrganizationSlug(ctx context.Context, tc *TenantContext, slug string) (string, error) {
	var out string
	pre := func() *failure {
		if slug == "" {
			return nil
		}
		if !hostname.SlugGrammar(slug) {
			return domainInput(string(tenancy.CodeInvalidSlug))
		}
		for _, l := range hostname.ReservedLabels() {
			if l == slug {
				return domainInput(string(tenancy.CodeSlugReserved))
			}
		}
		return nil
	}
	err := s.guard(ctx, tc, opSetSlug, capOrganizationManageDomains, true, Target{Kind: kindOrganization}, policyWriteQuery, pre,
		func(tx *trackedTx, f Facts) *failure {
			var arg *string
			if slug != "" {
				arg = &slug
			}
			got, err := db.New().SetOrganizationSlug(ctx, tx, db.SetOrganizationSlugParams{Slug: arg, ID: f.Organization.ID.String()})
			if err != nil {
				return domainFailure(ctx, tx, err, CodeNotFound)
			}
			if got != nil {
				out = *got
			}
			return nil
		})
	return out, err
}
