package workos

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"core/internal/identity/identitytest"
)

// U-JWKS (Plan r5 §8.2): rotation, unknown kid with rate-limited de-duplicated refresh, attacker key, oversize,
// timeout and redirect.

func TestJWKSRotationRefreshesOnceForUnknownKID(t *testing.T) {
	tf := newTokenFixture(t)
	if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); err != nil {
		t.Fatal(err)
	}
	hits := tf.fake.JWKSHits()
	tf.fake.Rotate(t) // the new key is published; tokens are now signed with it
	if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); err != nil {
		t.Fatalf("token from the rotated key: %v", err)
	}
	if got := tf.fake.JWKSHits() - hits; got != 1 {
		t.Fatalf("refreshes for one unknown kid = %d, want 1", got)
	}
	// The previous key stays valid while it is published.
	if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); err != nil {
		t.Fatal(err)
	}
}

func TestJWKSUnknownKIDIsRateLimitedAndFailsUntilTrustedRefresh(t *testing.T) {
	tf := newTokenFixture(t)
	if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); err != nil {
		t.Fatal(err)
	}
	tf.fake.RotateUnpublished(t)
	hits := tf.fake.JWKSHits()
	for i := 0; i < 5; i++ {
		if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); !errors.Is(err, errTokenInvalid) {
			t.Fatalf("attempt %d with an unpublished key: %v", i, err)
		}
	}
	if got := tf.fake.JWKSHits() - hits; got != 1 {
		t.Fatalf("refreshes within 30 s = %d, want exactly 1 (rate-limited, de-duplicated)", got)
	}
	// Once published, the key is used only after the next permitted refresh.
	tf.fake.Publish()
	tf.v.keys.limiter = rate.NewLimiter(rate.Inf, 1)
	if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); err != nil {
		t.Fatalf("after a trusted refresh: %v", err)
	}
}

func TestJWKSFailedRefreshKeepsTrustedSet(t *testing.T) {
	tf := newTokenFixture(t)
	if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); err != nil {
		t.Fatal(err)
	}
	tf.fake.FailJWKS(http.StatusInternalServerError)
	// A periodic refresh is due, fails, and the trusted set keeps working.
	tf.v.keys.loadedAt.Store(tf.now.Add(-2 * time.Hour).UnixNano())
	if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); err != nil {
		t.Fatalf("trusted set lost on a failed refresh: %v", err)
	}
	// An unknown kid during the outage is an availability failure, not a signature verdict.
	tf.fake.RotateUnpublished(t)
	if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); !errors.Is(err, errKeysUnavailable) {
		t.Fatalf("unknown kid during an outage: %v", err)
	}
}

func TestJWKSRefusesOversizeRedirectTimeoutAndEmpty(t *testing.T) {
	cases := map[string]func(f *identitytest.Fake){
		"oversize": func(f *identitytest.Fake) {
			f.ServeJWKS([]byte(`{"keys":[],"pad":"` + strings.Repeat("x", jwksMaxBody) + `"}`))
		},
		"redirect":  func(f *identitytest.Fake) { f.RedirectJWKS("https://attacker.example/jwks") },
		"timeout":   func(f *identitytest.Fake) { f.Delay(0, 2*time.Second) },
		"empty set": func(f *identitytest.Fake) { f.ServeJWKS([]byte(`{"keys":[]}`)) },
		"not json":  func(f *identitytest.Fake) { f.ServeJWKS([]byte(`<html>`)) },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			tf := newTokenFixture(t)
			tf.v.keys.timeout = 300 * time.Millisecond
			setup(tf.fake)
			start := time.Now()
			if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); !errors.Is(err, errKeysUnavailable) {
				t.Fatalf("got %v, want errKeysUnavailable", err)
			}
			if d := time.Since(start); d > time.Second {
				t.Fatalf("not bounded: %v", d)
			}
		})
	}
}

// A JWKS that publishes an attacker key under a new kid is only trusted because it came from the configured
// endpoint; a token whose kid names a trusted key but is signed by another key never verifies.
func TestJWKSAttackerKeyWithReusedKID(t *testing.T) {
	tf := newTokenFixture(t)
	attacker := identitytest.NewAttackerKey(t)
	tok := identitytest.SignWith(attacker, "RS256", tf.fake.CurrentKID(), tf.claims(nil))
	if _, err := tf.v.validate(context.Background(), tok, "user_01ABC"); !errors.Is(err, errTokenInvalid) {
		t.Fatalf("attacker-signed token with a trusted kid: %v", err)
	}
	tok = identitytest.SignWith(attacker, "RS256", "", tf.claims(nil))
	if _, err := tf.v.validate(context.Background(), tok, "user_01ABC"); !errors.Is(err, errTokenInvalid) {
		t.Fatalf("token without kid: %v", err)
	}
}
