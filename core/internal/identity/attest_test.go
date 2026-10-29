package identity

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/tenancy"
)

// attestDB answers the E2 principal reader from a fixed map and fails any transaction (AttestVerifiedAddress opens none).
type attestDB struct {
	principals map[[2]string]string
	fail       bool
	calls      int
}

func (d *attestDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	return nil, errors.New("unexpected transaction")
}

func (d *attestDB) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	d.calls++
	if sql != sqlAuthPrincipalSubject || len(args) != 2 {
		return attestRow{err: errors.New("unexpected statement")}
	}
	if d.fail {
		return attestRow{err: errors.New("connection refused")}
	}
	u, ok := d.principals[[2]string{args[0].(string), args[1].(string)}]
	if !ok {
		return attestRow{}
	}
	return attestRow{user: &u}
}

type attestRow struct {
	user *string
	err  error
}

func (r attestRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*(dest[0].(**string)) = r.user
	return nil
}

// TestVerifiedAddress (G019 Plan r2 addendum §2): the canonical user is the E2 resolution result (no user parameter),
// refusals happen before or instead of evidence, the zero value is recognisable, and the local normalization equals
// tenancy.NormalizeAddress on a shared table.
func TestVerifiedAddress(t *testing.T) {
	ctx := t.Context()
	const user = "7f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9"
	db := &attestDB{principals: map[[2]string]string{{"https://issuer.example", "user_01"}: user}}
	s := NewService(newStore(db), Config{})
	at := time.Date(2026, 10, 18, 12, 0, 0, 0, time.FixedZone("x", 3600))

	v, err := s.AttestVerifiedAddress(ctx, "https://issuer.example", "user_01", "  Alice.Smith@Example.COM ", at, "provider_email_verified")
	if err != nil {
		t.Fatal(err)
	}
	if v.IsZero() || v.UserID() != user || v.Issuer() != "https://issuer.example" || v.Subject() != "user_01" ||
		v.Address() != "Alice.Smith@example.com" || !v.verifiedAt.Equal(at) || v.verifiedAt.Location() != time.UTC || v.source != "provider_email_verified" {
		t.Fatalf("evidence = %+v", v)
	}
	if !(VerifiedAddress{}).IsZero() {
		t.Fatal("zero evidence not recognised")
	}

	calls := db.calls
	for _, c := range []struct {
		name                     string
		issuer, subject, address string
		at                       time.Time
		source                   string
		want                     *Error
	}{
		{"empty issuer", "", "user_01", "a@example.com", at, "s", ErrInvalidRequest},
		{"empty subject", "https://issuer.example", "", "a@example.com", at, "s", ErrInvalidRequest},
		{"empty source", "https://issuer.example", "user_01", "a@example.com", at, "", ErrInvalidRequest},
		{"zero time", "https://issuer.example", "user_01", "a@example.com", time.Time{}, "s", ErrInvalidRequest},
		{"bad address", "https://issuer.example", "user_01", "not-an-address", at, "s", ErrInvalidRequest},
		{"two ats", "https://issuer.example", "user_01", "a@b@example.com", at, "s", ErrInvalidRequest},
	} {
		if _, err := s.AttestVerifiedAddress(ctx, c.issuer, c.subject, c.address, c.at, c.source); !errors.Is(err, c.want) {
			t.Errorf("%s: %v, want %v", c.name, err, c.want)
		}
	}
	if db.calls != calls {
		t.Fatalf("input refusals reached the store (%d calls)", db.calls-calls)
	}
	if _, err := s.AttestVerifiedAddress(ctx, "https://issuer.example", "user_02", "a@example.com", at, "s"); !errors.Is(err, ErrUnauthenticated) {
		t.Errorf("unknown principal: %v", err)
	}
	if _, err := s.AttestVerifiedAddress(ctx, "https://other.example", "user_01", "a@example.com", at, "s"); !errors.Is(err, ErrUnauthenticated) {
		t.Errorf("issuer is part of the principal: %v", err)
	}
	db.fail = true
	if _, err := s.AttestVerifiedAddress(ctx, "https://issuer.example", "user_01", "a@example.com", at, "s"); !errors.Is(err, ErrUnavailable) {
		t.Errorf("store failure: %v", err)
	}

	// Cross-package pin: identity's local normalization equals tenancy.NormalizeAddress.
	for _, in := range []string{
		"a@example.com", " \tA.B+tag@Example.COM\n", "Ünï@Bücher.EXAMPLE", "x@éxample.com", "É@example.com",
		"@example.com", "a@", "a@@example.com", "a b@example.com", "a@exa\x00mple.com", "a b@example.com", "",
		strings.Repeat("l", 64) + "@example.com", strings.Repeat("l", 65) + "@example.com",
		"a@" + strings.Repeat("d", 251), "a@" + strings.Repeat("d", 252), "\xff@example.com",
	} {
		want, terr := tenancy.NormalizeAddress(in)
		got, ok := normalizeAddress(in)
		if ok != (terr == nil) || got != want {
			t.Errorf("normalizeAddress(%q) = %q %v; tenancy = %q %v", in, got, ok, want, terr)
		}
	}
}
