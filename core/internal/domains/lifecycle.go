package domains

import "time"

// Check is the kind of an ownership check: the first verification of a pending claim (DNS only: under E1 no
// certificate can exist before the issuance predicate holds) or a recheck of a verified, active or suspended row (DNS
// and TLS readiness).
type Check string

// The checks.
const (
	CheckVerify  Check = "verify"
	CheckRecheck Check = "recheck"
)

// Observation is what one check observed outside every transaction: the DNS proof outcome (ok, mismatch, nxdomain or
// transient) and, for a recheck whose proof is ok, the TLS adapter's status for the row's DomainRef.
type Observation struct {
	Proof Outcome
	TLS   TLSResult
}

// Update is the guarded column change one observation produces (Plan r4 §1.4 "exact allowed updates"). Nil pointer
// fields are left unchanged; the facade adds version + 1.
type Update struct {
	State            State
	VerifiedAt       *time.Time
	VerifiedUntil    *time.Time
	LastCheckAt      time.Time
	LastCheckOutcome Outcome
	CheckFailures    int32
	TLS              *TLSUpdate
}

// TLSUpdate is a change of the TLS columns.
type TLSUpdate struct {
	State     TLSState
	CheckedAt time.Time
	NotAfter  *time.Time
	Reference *string
}

// Begin reports whether a check of kind c may start on r at now: a verify only on a pending row whose challenge has
// not expired; a recheck only on a verified, active or suspended row. Removed rows are excluded from both (the only
// operation on a removed row is the explicit withdrawal retry).
func Begin(r Record, c Check, now time.Time) error {
	switch {
	case c == CheckVerify && r.State == StatePending:
		if !now.Before(r.ChallengeExpiresAt) {
			return ErrChallengeExpired
		}
		return nil
	case c == CheckRecheck && (r.State == StateVerified || r.State == StateActive || r.State == StateSuspended):
		return nil
	}
	return ErrStateInvalid
}

// Apply computes the update of one observation on r at now with the configured proof freshness. It is the pure form of
// the lifecycle table; the database guard enforces the same edges. A transient outcome (DNS, or an unknown TLS
// status) changes neither the state nor any deadline. A confirmed proof loss suspends a verified or active row and
// sets verified_until = now (issuance stops at that commit), also on an already suspended row. A TLS-only failure
// suspends an active row and leaves the proof untouched.
func Apply(r Record, c Check, obs Observation, now time.Time, freshness time.Duration) (Update, error) {
	if err := Begin(r, c, now); err != nil {
		return Update{}, err
	}
	u := Update{State: r.State, LastCheckAt: now, LastCheckOutcome: obs.Proof, CheckFailures: r.CheckFailures + 1}
	fresh := now.Add(freshness)
	switch obs.Proof {
	case OutcomeOK:
	case OutcomeMismatch, OutcomeNXDomain:
		if r.State != StatePending {
			u.State, u.VerifiedUntil = StateSuspended, &now
		}
		return u, nil
	default:
		u.LastCheckOutcome = OutcomeTransient
		return u, nil
	}
	if c == CheckVerify {
		u.State, u.VerifiedAt, u.VerifiedUntil, u.CheckFailures = StateVerified, &now, &fresh, 0
		return u, nil
	}
	t := obs.TLS
	if t.State == TLSReady && !t.NotAfter.After(now) {
		t.State = TLSExpired
	}
	tu := &TLSUpdate{State: t.State, CheckedAt: now}
	if !t.NotAfter.IsZero() {
		na := t.NotAfter
		tu.NotAfter = &na
	}
	if t.Reference != "" {
		ref := t.Reference
		tu.Reference = &ref
	}
	switch t.State {
	case TLSReady:
		u.State, u.VerifiedUntil, u.CheckFailures, u.TLS = StateActive, &fresh, 0, tu
	case TLSNone, TLSRequested, TLSPending:
		tu.NotAfter = nil // a stale stored not_after is harmless once the state is no longer ready
		u.State, u.VerifiedUntil, u.CheckFailures, u.TLS = StateVerified, &fresh, 0, tu
	case TLSError, TLSExpired:
		u.LastCheckOutcome, u.TLS = OutcomeTLSError, tu
		switch r.State {
		case StateActive:
			u.State = StateSuspended // proof untouched: issuance continues while fresh
		case StateVerified:
			u.VerifiedUntil = &fresh
		}
	default: // unknown or anything unexpected: transient
		u.LastCheckOutcome = OutcomeTransient
	}
	return u, nil
}

// RemovalOutcome maps a post-commit Remove (or an explicit retry) result onto the removed row's cleanup outcome.
func RemovalOutcome(res TLSResult) (TLSState, Outcome) {
	switch res.State {
	case TLSWithdrawn:
		return TLSWithdrawn, OutcomeWithdrawn
	case TLSError:
		return TLSError, OutcomeWithdrawError
	}
	return TLSUnknown, OutcomeWithdrawUnknown
}

// RetryAllowed reports whether an explicit withdrawal retry may start on r (Plan r4 §1.5): the row is removed, its
// withdrawal is not complete and fewer than max retries were reserved.
func RetryAllowed(r Record, max int) (bool, error) {
	if r.State != StateRemoved || (r.TLSState != TLSWithdrawing && r.TLSState != TLSUnknown && r.TLSState != TLSError) {
		return false, ErrStateInvalid
	}
	return int(r.CheckFailures) < max, nil
}
