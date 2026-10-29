package identity

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TestFinalizeRetirementRefusesBeforeSQL (G019 F-01, MEM-14 unit half): a non-canonical user ID or a nil check is
// invalid_request before any database work (the store has no pool here, so any statement would panic); a Service
// without a store is unavailable for valid input.
func TestFinalizeRetirementRefusesBeforeSQL(t *testing.T) {
	ctx := context.Background()
	ok := func(context.Context, pgx.Tx) error { return nil }
	valid := "8c6f0a3e-1d2b-4c5a-9e7f-0a1b2c3d4e5f"
	st := newStore(nil)
	for name, c := range map[string]struct {
		user  string
		check RetirementCheck
	}{
		"empty":     {"", ok},
		"zero":      {zeroUUID, ok},
		"uppercase": {"8C6F0A3E-1D2B-4C5A-9E7F-0A1B2C3D4E5F", ok},
		"nil check": {valid, nil},
	} {
		if err := st.FinalizeRetirement(ctx, c.user, time.Now(), c.check); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("store %s: %v, want invalid_request", name, err)
		}
		if err := (&Service{}).FinalizeRetirement(ctx, c.user, c.check); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("service %s: %v, want invalid_request", name, err)
		}
	}
	if err := (&Service{}).FinalizeRetirement(ctx, valid, ok); !errors.Is(err, ErrUnavailable) {
		t.Errorf("service without store: %v, want unavailable", err)
	}
}

// readTx is a pgx.Tx whose reads succeed without a database; any other method would panic (nil embedded Tx), so the
// guarded view is shown never to forward them.
type readTx struct {
	pgx.Tx
	queries int
}

func (r *readTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	r.queries++
	return nil, nil
}

func (r *readTx) QueryRow(context.Context, string, ...any) pgx.Row {
	r.queries++
	return refusedRow{}
}

// TestRetirementCheckView (refinement d): while the check runs only Query and QueryRow reach the transaction; every
// other method is refused and marks a misuse; after the check returned every method is refused (retention).
func TestRetirementCheckView(t *testing.T) {
	ctx := context.Background()
	inner := &readTx{}
	v := &checkTx{tx: inner, open: true}
	if _, err := v.Query(ctx, "SELECT 1"); err != nil {
		t.Fatalf("open Query: %v", err)
	}
	_ = v.QueryRow(ctx, "SELECT 1")
	if inner.queries != 2 {
		t.Fatalf("open reads forwarded %d, want 2", inner.queries)
	}
	if v.seal() {
		t.Fatal("reads alone marked a misuse")
	}

	refused := map[string]func(c *checkTx) error{
		"Exec":     func(c *checkTx) error { _, err := c.Exec(ctx, "UPDATE x"); return err },
		"Begin":    func(c *checkTx) error { _, err := c.Begin(ctx); return err },
		"Commit":   func(c *checkTx) error { return c.Commit(ctx) },
		"Rollback": func(c *checkTx) error { return c.Rollback(ctx) },
		"CopyFrom": func(c *checkTx) error { _, err := c.CopyFrom(ctx, pgx.Identifier{"x"}, nil, nil); return err },
		"Prepare":  func(c *checkTx) error { _, err := c.Prepare(ctx, "p", "SELECT 1"); return err },
		"SendBatch": func(c *checkTx) error {
			return c.SendBatch(ctx, &pgx.Batch{}).Close()
		},
		"LargeObjects": func(c *checkTx) error { _ = c.LargeObjects(); return errCheckMisuse },
		"Conn": func(c *checkTx) error {
			if c.Conn() != nil {
				return nil
			}
			return errCheckMisuse
		},
	}
	for name, call := range refused {
		c := &checkTx{tx: &readTx{}, open: true}
		if err := call(c); !errors.Is(err, errCheckMisuse) {
			t.Errorf("%s: %v, want refused", name, err)
		}
		if !c.seal() {
			t.Errorf("%s did not mark a misuse", name)
		}
	}

	// Retention: a read after the check returned is refused, never forwarded, and marks a misuse.
	inner = &readTx{}
	v = &checkTx{tx: inner, open: true}
	_ = v.seal()
	if _, err := v.Query(ctx, "SELECT 1"); !errors.Is(err, errCheckMisuse) {
		t.Errorf("sealed Query: %v", err)
	}
	if err := v.QueryRow(ctx, "SELECT 1").Scan(); !errors.Is(err, errCheckMisuse) {
		t.Errorf("sealed QueryRow: %v", err)
	}
	if inner.queries != 0 || !v.seal() {
		t.Errorf("sealed reads forwarded %d, misuse not marked", inner.queries)
	}
}

// sqlDB is a storeDB whose one transaction is t.
type sqlDB struct{ t *sqlTx }

func (d *sqlDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) { return d.t, nil }
func (d *sqlDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return recRow{err: errors.New("sqlDB: no reader")}
}

// sqlTx is a recTx that also records the text of every statement in order; an Exec whose text contains failAt fails
// (after it is counted), as a database error would.
type sqlTx struct {
	*recTx
	stmts  []string
	failAt string
}

func (t *sqlTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	t.stmts = append(t.stmts, strings.TrimSpace(sql))
	return t.recTx.QueryRow(ctx, sql, args...)
}

func (t *sqlTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	t.stmts = append(t.stmts, strings.TrimSpace(sql))
	tag, err := t.recTx.Exec(ctx, sql, args...)
	if err == nil && t.failAt != "" && strings.Contains(sql, t.failAt) {
		err = errors.New("injected exec failure")
	}
	return tag, err
}

// TestFinalizeRetirementRevokesEnrollmentInOrder (G058 D-11, unit half; G062 SUP-10): the finalization's writes are
// exactly users → sessions → enrollment_sessions (the shared revokeEnrollmentSessionsAll statement, reason disabled) →
// signup_sessions (the shared revokeSignupSessionsAll statement, as DisableUser), and the revocation-gap read stays the
// last statement before the one COMMIT; any failed write, the added enrollment and signup UPDATEs included, is
// identity_unavailable with no COMMIT (the deferred rollback discards every write); a refusing check executes nothing.
func TestFinalizeRetirementRevokesEnrollmentInOrder(t *testing.T) {
	ctx := context.Background()
	user := "8c6f0a3e-1d2b-4c5a-9e7f-0a1b2c3d4e5f"
	now := time.Now()
	ok := func(context.Context, pgx.Tx) error { return nil }

	tx := &sqlTx{recTx: &recTx{}}
	if err := newStore(&sqlDB{t: tx}).FinalizeRetirement(ctx, user, now, ok); err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if tx.execs != 4 || tx.commits != 1 {
		t.Fatalf("execs %d commits %d, want 4 and 1", tx.execs, tx.commits)
	}
	var writes []string
	for _, s := range tx.stmts {
		if strings.HasPrefix(s, "UPDATE ") {
			writes = append(writes, s)
		}
	}
	want := []string{"UPDATE app.users SET status = 'disabled'", "UPDATE app.sessions SET revoked_at", sqlRevokeEnrollmentAll, sqlRevokeSignupAll}
	if len(writes) != len(want) {
		t.Fatalf("writes %q, want %d", writes, len(want))
	}
	for i, w := range want {
		if !strings.HasPrefix(writes[i], w) {
			t.Errorf("write %d is %q, want prefix %q", i, writes[i], w)
		}
	}
	if n := len(tx.stmts); n < 2 || !strings.Contains(tx.stmts[n-1], "audit.revocation_gap") || tx.stmts[n-2] != sqlRevokeSignupAll {
		t.Errorf("statement order %q: want the signup UPDATE, then the gap read last before COMMIT", tx.stmts)
	}

	for name, failAt := range map[string]string{
		"signup UPDATE":     "UPDATE app.signup_sessions",
		"enrollment UPDATE": "UPDATE app.enrollment_sessions",
		"sessions UPDATE":   "UPDATE app.sessions",
		"users UPDATE":      "UPDATE app.users",
	} {
		tx := &sqlTx{recTx: &recTx{}, failAt: failAt}
		if err := newStore(&sqlDB{t: tx}).FinalizeRetirement(ctx, user, now, ok); !errors.Is(err, ErrUnavailable) {
			t.Errorf("%s failing: %v, want identity_unavailable", name, err)
		}
		if tx.commits != 0 {
			t.Errorf("%s failing: %d commits, want 0", name, tx.commits)
		}
	}
	// The recorder's own execErr (every Exec fails): unavailable at the first write, never committed.
	rtx := &sqlTx{recTx: &recTx{execErr: errors.New("exec refused")}}
	if err := newStore(&sqlDB{t: rtx}).FinalizeRetirement(ctx, user, now, ok); !errors.Is(err, ErrUnavailable) || rtx.commits != 0 || rtx.execs != 1 {
		t.Errorf("execErr: %v after %d execs and %d commits, want identity_unavailable after 1 exec and 0 commits", err, rtx.execs, rtx.commits)
	}

	refusal := errors.New("refused by check")
	ftx := &sqlTx{recTx: &recTx{}}
	if err := newStore(&sqlDB{t: ftx}).FinalizeRetirement(ctx, user, now, func(context.Context, pgx.Tx) error { return refusal }); err != refusal {
		t.Errorf("refusing check: %v, want the refusal unchanged", err)
	}
	if ftx.execs != 0 || ftx.commits != 0 {
		t.Errorf("refusing check: execs %d commits %d, want 0 and 0", ftx.execs, ftx.commits)
	}
}
