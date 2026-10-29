package identity_test

import (
	"testing"
	"time"
)

// The pure evaluation rules of the manual qualification observer (manual_qualification_test.go, build tag
// workos_manual), kept untagged so that their unit tests run in every core check (G004 C7 addendum §1 and §3). The
// observer passes them only timestamps, booleans and closed attestation words.

// mqFreshMargin is the slack the freshness test adds to the configured maximum authentication age.
const mqFreshMargin = 30 * time.Second

// mqLatest is the latest of the given signed authentication times (zero when there is none).
func mqLatest(times ...time.Time) time.Time {
	var latest time.Time
	for _, t := range times {
		if t.After(latest) {
			latest = t
		}
	}
	return latest
}

// mqReadyAt is the earliest moment a freshness test may ask for a sign-in: the latest observed signed authentication
// plus the bound and the margin. A sign-in that is not prompted before then may legitimately reuse it.
func mqReadyAt(latest time.Time, maxAge time.Duration) time.Time {
	return latest.Add(maxAge + mqFreshMargin)
}

// mqFreshInput is one freshness observation (MQ-5).
type mqFreshInput struct {
	Attestation string    // the operator's word: "prompt" or "no-prompt"
	SameUser    bool      // the new session belongs to the qualified user
	Baseline    time.Time // the latest signed authentication when the test became ready
	NewAuth     time.Time // the new session's signed auth_time
	ObservedAt  time.Time // when the observer saw the new session
	MaxAge      time.Duration
}

// mqFreshness passes only a prompted sign-in whose new signed authentication is later than the baseline and still
// within the bound when observed.
func mqFreshness(in mqFreshInput) (bool, string) {
	switch {
	case in.Attestation != "prompt":
		return false, "no credential prompt was attested"
	case !in.SameUser:
		return false, "the new session belongs to another user"
	case !in.NewAuth.After(in.Baseline):
		return false, "the signed authentication time did not advance past the baseline"
	case in.ObservedAt.Sub(in.NewAuth) > in.MaxAge+mqFreshMargin:
		return false, "the new signed authentication is older than the bound"
	}
	return true, ""
}

// mqSSOInput is one reuse observation (MQ-3, or linked evidence).
type mqSSOInput struct {
	Attestation     string    // "no-prompt" or "prompt"
	SameUser        bool      // the new session maps to the same canonical user
	RevokedAt       time.Time // the local revocation of the previous session
	NewCreated      time.Time // the new application session
	NewAuth         time.Time // its signed auth_time
	PriorLatestAuth time.Time // the latest signed authentication before the new session
}

// mqSSOReuse qualifies single sign-on reuse only for an unprompted sign-in that maps to the same user, creates a new
// application session after the local revocation and keeps the prior signed authentication time. A prompted sign-in
// is an ordinary sign-in and never counts as reuse.
func mqSSOReuse(in mqSSOInput) (bool, string) {
	switch {
	case in.Attestation != "no-prompt":
		return false, "a credential prompt was attested: an ordinary sign-in, not reuse"
	case !in.SameUser:
		return false, "the new session belongs to another user"
	case in.RevokedAt.IsZero() || !in.NewCreated.After(in.RevokedAt):
		return false, "no new application session after the local revocation"
	case !in.NewAuth.Equal(in.PriorLatestAuth):
		return false, "the signed authentication time was not retained"
	}
	return true, ""
}

func mqClock(hms string) time.Time {
	t, err := time.Parse("15:04:05", hms)
	if err != nil {
		panic(err)
	}
	return time.Date(2026, 10, 8, t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
}

// The run-2 sequence (MQ-1 → MQ-3 → MQ-4, then MQ-5): readiness follows the latest authentication, never an earlier
// one, and a sign-in during the wait moves it.
func TestMQReadyFollowsTheLatestAuthentication(t *testing.T) {
	const maxAge = 2 * time.Minute
	mq1, mq3, mq4 := mqClock("17:35:00"), mqClock("17:41:19"), mqClock("17:44:05")
	ready := mqReadyAt(mqLatest(mq1, mq3, mq4), maxAge)
	if !ready.Equal(mqClock("17:46:35")) {
		t.Fatalf("ready %s, want 17:46:35", ready.Format(time.TimeOnly))
	}
	for _, auth := range []time.Time{mq1, mq3, mq4} {
		if !ready.After(auth.Add(maxAge)) {
			t.Fatalf("ready %s is not past %s plus the bound", ready.Format(time.TimeOnly), auth.Format(time.TimeOnly))
		}
	}
	// Run 2's defect: a baseline of MQ-1 alone made the observer ready at 17:37:30, before MQ-4's authentication had
	// aged past the bound.
	if stale := mqReadyAt(mq1, maxAge); !stale.Before(mq4.Add(maxAge)) {
		t.Fatal("the control no longer reproduces the stale baseline")
	}
	// A sign-in during the wait moves readiness.
	if moved := mqReadyAt(mqLatest(mq1, mq3, mq4, mqClock("17:45:00")), maxAge); !moved.Equal(mqClock("17:47:30")) {
		t.Fatalf("moved ready %s", moved.Format(time.TimeOnly))
	}
	if !mqLatest().IsZero() {
		t.Fatal("no authentication must give the zero time")
	}
}

func TestMQFreshnessVerdict(t *testing.T) {
	const maxAge = 2 * time.Minute
	good := mqFreshInput{Attestation: "prompt", SameUser: true, Baseline: mqClock("17:44:05"), NewAuth: mqClock("17:47:10"),
		ObservedAt: mqClock("17:47:12"), MaxAge: maxAge}
	if ok, note := mqFreshness(good); !ok {
		t.Fatalf("a prompted, advanced, fresh authentication failed: %s", note)
	}
	for name, change := range map[string]func(*mqFreshInput){
		"no prompt (run 2's MQ-5)": func(in *mqFreshInput) { in.Attestation = "no-prompt"; in.NewAuth = in.Baseline },
		"no prompt":                func(in *mqFreshInput) { in.Attestation = "no-prompt" },
		"another user":             func(in *mqFreshInput) { in.SameUser = false },
		"retained time":            func(in *mqFreshInput) { in.NewAuth = in.Baseline },
		"earlier time":             func(in *mqFreshInput) { in.NewAuth = in.Baseline.Add(-time.Second) },
		"stale when observed":      func(in *mqFreshInput) { in.ObservedAt = in.NewAuth.Add(maxAge + mqFreshMargin + time.Second) },
	} {
		in := good
		change(&in)
		if ok, _ := mqFreshness(in); ok {
			t.Errorf("%s passed", name)
		}
	}
}

func TestMQSSOReuseVerdict(t *testing.T) {
	// Run 2's linked evidence: org-a's session revoked locally at 17:45:30; the next org-a session at 17:45:34, not
	// prompted, same user, keeping the latest prior signed authentication (MQ-4's, 17:44:05).
	linked := mqSSOInput{Attestation: "no-prompt", SameUser: true, RevokedAt: mqClock("17:45:30"), NewCreated: mqClock("17:45:34"),
		NewAuth: mqClock("17:44:05"), PriorLatestAuth: mqClock("17:44:05")}
	if ok, note := mqSSOReuse(linked); !ok {
		t.Fatalf("the linked run-2 evidence does not qualify: %s", note)
	}
	for name, change := range map[string]func(*mqSSOInput){
		// Negative control: run 2's MQ-3 was prompted and must not count as reuse, whatever its times.
		"prompted with a retained time": func(in *mqSSOInput) { in.Attestation = "prompt" },
		"prompted with a new time":      func(in *mqSSOInput) { in.Attestation = "prompt"; in.NewAuth = mqClock("17:41:19") },
		"another user":                  func(in *mqSSOInput) { in.SameUser = false },
		"a new authentication":          func(in *mqSSOInput) { in.NewAuth = mqClock("17:45:33") },
		"no local revocation":           func(in *mqSSOInput) { in.RevokedAt = time.Time{} },
		"session before the revocation": func(in *mqSSOInput) { in.NewCreated = mqClock("17:45:29") },
	} {
		in := linked
		change(&in)
		if ok, _ := mqSSOReuse(in); ok {
			t.Errorf("%s qualified as reuse", name)
		}
	}
}
