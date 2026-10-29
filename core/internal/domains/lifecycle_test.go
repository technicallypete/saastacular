package domains

import (
	"errors"
	"testing"
	"time"
)

// TestDomainStateMachine mirrors the lifecycle table of Plan r4 §1.4 on the pure Apply/Begin functions; the database
// guard enforces the same edges (TestDomainsCatalog, TestDomains).
func TestDomainStateMachine(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	fresh := 720 * time.Hour
	past, future := now.Add(-time.Hour), now.Add(240*time.Hour)
	notAfter := now.Add(60 * 24 * time.Hour)
	row := func(s State, tls TLSState, vu *time.Time) Record {
		return Record{ID: "d", Generation: 1, Hostname: "portal.example.com", State: s, TLSState: tls, VerifiedUntil: vu,
			ChallengeExpiresAt: now.Add(time.Hour), CheckFailures: 2}
	}
	ok := Observation{Proof: OutcomeOK}
	tlsObs := func(s TLSState, na time.Time) Observation {
		return Observation{Proof: OutcomeOK, TLS: TLSResult{State: s, NotAfter: na}}
	}
	cases := []struct {
		name      string
		r         Record
		c         Check
		obs       Observation
		state     State
		outcome   Outcome
		failures  int32
		vuChanged bool
		vuNow     bool
		tls       TLSState
		err       error
	}{
		{"pending verify ok", row(StatePending, TLSNone, nil), CheckVerify, ok, StateVerified, OutcomeOK, 0, true, false, "", nil},
		{"pending verify mismatch stays pending", row(StatePending, TLSNone, nil), CheckVerify, Observation{Proof: OutcomeMismatch}, StatePending, OutcomeMismatch, 3, false, false, "", nil},
		{"pending verify nxdomain stays pending", row(StatePending, TLSNone, nil), CheckVerify, Observation{Proof: OutcomeNXDomain}, StatePending, OutcomeNXDomain, 3, false, false, "", nil},
		{"pending verify transient", row(StatePending, TLSNone, nil), CheckVerify, Observation{Proof: OutcomeTransient}, StatePending, OutcomeTransient, 3, false, false, "", nil},
		{"verified + tls ready -> active", row(StateVerified, TLSNone, &future), CheckRecheck, tlsObs(TLSReady, notAfter), StateActive, OutcomeOK, 0, true, false, TLSReady, nil},
		{"verified + tls pending stays verified", row(StateVerified, TLSNone, &future), CheckRecheck, tlsObs(TLSPending, time.Time{}), StateVerified, OutcomeOK, 0, true, false, TLSPending, nil},
		{"verified + tls ready but expired", row(StateVerified, TLSNone, &future), CheckRecheck, tlsObs(TLSReady, past), StateVerified, OutcomeTLSError, 3, true, false, TLSExpired, nil},
		{"active + proof loss -> suspended now", row(StateActive, TLSReady, &future), CheckRecheck, Observation{Proof: OutcomeMismatch}, StateSuspended, OutcomeMismatch, 3, true, true, "", nil},
		{"verified + nxdomain -> suspended now", row(StateVerified, TLSNone, &future), CheckRecheck, Observation{Proof: OutcomeNXDomain}, StateSuspended, OutcomeNXDomain, 3, true, true, "", nil},
		{"tls-suspended + proof loss stays suspended, now", row(StateSuspended, TLSError, &future), CheckRecheck, Observation{Proof: OutcomeMismatch}, StateSuspended, OutcomeMismatch, 3, true, true, "", nil},
		{"proof-lost suspended + loss again", row(StateSuspended, TLSError, &past), CheckRecheck, Observation{Proof: OutcomeNXDomain}, StateSuspended, OutcomeNXDomain, 3, true, true, "", nil},
		{"active + tls error -> suspended, proof untouched", row(StateActive, TLSReady, &future), CheckRecheck, tlsObs(TLSError, time.Time{}), StateSuspended, OutcomeTLSError, 3, false, false, TLSError, nil},
		{"active + tls expired -> suspended", row(StateActive, TLSReady, &future), CheckRecheck, tlsObs(TLSExpired, past), StateSuspended, OutcomeTLSError, 3, false, false, TLSExpired, nil},
		{"active + tls unknown is transient", row(StateActive, TLSReady, &future), CheckRecheck, tlsObs(TLSUnknown, time.Time{}), StateActive, OutcomeTransient, 3, false, false, "", nil},
		{"active + dns transient unchanged", row(StateActive, TLSReady, &future), CheckRecheck, Observation{Proof: OutcomeTransient}, StateActive, OutcomeTransient, 3, false, false, "", nil},
		{"suspended + ok + ready -> active", row(StateSuspended, TLSError, &past), CheckRecheck, tlsObs(TLSReady, notAfter), StateActive, OutcomeOK, 0, true, false, TLSReady, nil},
		{"suspended + ok + pending -> verified", row(StateSuspended, TLSError, &past), CheckRecheck, tlsObs(TLSPending, time.Time{}), StateVerified, OutcomeOK, 0, true, false, TLSPending, nil},
		{"active + ok + pending -> verified", row(StateActive, TLSReady, &future), CheckRecheck, tlsObs(TLSPending, time.Time{}), StateVerified, OutcomeOK, 0, true, false, TLSPending, nil},
		{"recheck of pending refused", row(StatePending, TLSNone, nil), CheckRecheck, ok, "", "", 0, false, false, "", ErrStateInvalid},
		{"verify of verified refused", row(StateVerified, TLSNone, &future), CheckVerify, ok, "", "", 0, false, false, "", ErrStateInvalid},
		{"recheck of removed refused", row(StateRemoved, TLSWithdrawn, nil), CheckRecheck, ok, "", "", 0, false, false, "", ErrStateInvalid},
		{"verify of removed refused", row(StateRemoved, TLSWithdrawn, nil), CheckVerify, ok, "", "", 0, false, false, "", ErrStateInvalid},
	}
	for _, c := range cases {
		u, err := Apply(c.r, c.c, c.obs, now, fresh)
		if !errors.Is(err, c.err) {
			t.Errorf("%s: err %v, want %v", c.name, err, c.err)
			continue
		}
		if err != nil {
			continue
		}
		if u.State != c.state || u.LastCheckOutcome != c.outcome || u.CheckFailures != c.failures || !u.LastCheckAt.Equal(now) {
			t.Errorf("%s: %+v", c.name, u)
		}
		if (u.VerifiedUntil != nil) != c.vuChanged {
			t.Errorf("%s: verified_until changed = %v", c.name, u.VerifiedUntil != nil)
		}
		if u.VerifiedUntil != nil {
			if c.vuNow && !u.VerifiedUntil.Equal(now) {
				t.Errorf("%s: confirmed loss must set verified_until = now, got %v", c.name, u.VerifiedUntil)
			}
			if !c.vuNow && !u.VerifiedUntil.Equal(now.Add(fresh)) {
				t.Errorf("%s: success must renew verified_until forward, got %v", c.name, u.VerifiedUntil)
			}
		}
		if c.tls == "" && u.TLS != nil || c.tls != "" && (u.TLS == nil || u.TLS.State != c.tls) {
			t.Errorf("%s: tls update %+v, want %q", c.name, u.TLS, c.tls)
		}
		if c.state == StateActive && c.r.State != StateActive && (u.TLS == nil || u.TLS.NotAfter == nil) {
			t.Errorf("%s: entering active needs a stored not_after", c.name)
		}
		if u.VerifiedAt != nil && !(c.r.State == StatePending && c.state == StateVerified) {
			t.Errorf("%s: verified_at set outside the first verification", c.name)
		}
	}

	// A pending claim whose challenge expired cannot verify (it must be rotated first).
	r := row(StatePending, TLSNone, nil)
	r.ChallengeExpiresAt = now
	if _, err := Apply(r, CheckVerify, ok, now, fresh); !errors.Is(err, ErrChallengeExpired) {
		t.Errorf("expired challenge: %v", err)
	}

	// Removal outcomes and the explicit withdrawal retry precondition.
	for _, c := range []struct {
		in      TLSState
		state   TLSState
		outcome Outcome
	}{{TLSWithdrawn, TLSWithdrawn, OutcomeWithdrawn}, {TLSError, TLSError, OutcomeWithdrawError}, {TLSUnknown, TLSUnknown, OutcomeWithdrawUnknown}, {TLSPending, TLSUnknown, OutcomeWithdrawUnknown}} {
		if s, o := RemovalOutcome(TLSResult{State: c.in}); s != c.state || o != c.outcome {
			t.Errorf("RemovalOutcome(%s) = %s %s", c.in, s, o)
		}
	}
	removed := row(StateRemoved, TLSUnknown, nil)
	if ok, err := RetryAllowed(removed, 5); !ok || err != nil {
		t.Errorf("retry of unknown cleanup: %v %v", ok, err)
	}
	removed.CheckFailures = 5
	if ok, err := RetryAllowed(removed, 5); ok || err != nil {
		t.Errorf("retry at the bound must be exhausted: %v %v", ok, err)
	}
	for _, r := range []Record{row(StateRemoved, TLSWithdrawn, nil), row(StateActive, TLSReady, &future)} {
		if _, err := RetryAllowed(r, 5); !errors.Is(err, ErrStateInvalid) {
			t.Errorf("retry on %s/%s: %v", r.State, r.TLSState, err)
		}
	}
	if (Record{ID: "x", Generation: 3, Hostname: "h.example.com", TLSReference: ptr("res-1")}).Ref() != (DomainRef{ID: "x", Generation: 3, Hostname: "h.example.com", Reference: "res-1"}) {
		t.Error("Ref must carry id, generation, hostname and the stored reference")
	}
}

func ptr[T any](v T) *T { return &v }
