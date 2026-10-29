package identity

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/audit"
)

// The core auth error vocabulary is closed: fixed lowercase codes with fixed statuses, and AsError never exposes an
// arbitrary error's text.
func TestClosedErrorVocabulary(t *testing.T) {
	codeRe := regexp.MustCompile(`^[a-z][a-z_]{2,40}$`)
	seen := map[string]bool{}
	for _, e := range []*Error{
		ErrInvalidRequest, ErrAmbiguousParameter, ErrUnsupportedMediaType, ErrOriginRejected, ErrUnauthenticated,
		ErrCSRFFailed, ErrAlreadyAuthenticated, ErrTransactionInvalid, ErrTransactionExpired, ErrLoginCanceled,
		ErrCodeRejected, ErrIdentityConflict, ErrUserDisabled, ErrImpersonationRefused, ErrStaleAuthentication,
		ErrMFAUnverified, ErrProviderError, ErrProviderTimeout, ErrNamespaceMismatch, ErrUnconfigured, ErrUnavailable,
	} {
		if !codeRe.MatchString(e.Code) || seen[e.Code] || e.Status < 400 || e.Status > 599 {
			t.Errorf("bad entry %+v", e)
		}
		seen[e.Code] = true
	}
	if AsError(errors.New("pq: password=secret")) != ErrUnavailable {
		t.Fatal("an arbitrary error must map to identity_unavailable")
	}
	if AsError(ErrCSRFFailed).Status != http.StatusForbidden {
		t.Fatal("AsError lost the status")
	}
}

// countingDB is a storeDB that records every use and never reaches a database.
type countingDB struct{ begins, queries int }

func (d *countingDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	d.begins++
	return nil, errors.New("countingDB: no database")
}

func (d *countingDB) QueryRow(context.Context, string, ...any) pgx.Row {
	d.queries++
	return nil
}

// DisableUser and EnableUser reject every user ID the context parser would refuse (anything but a nonzero, lowercase,
// hyphenated UUID) as invalid_request before any database work: no transaction and no statement (Duo finding on MR
// !23, G008 correction pass 1). A canonical ID reaches the database.
func TestUserStatusRejectsNonCanonicalIDsBeforeDatabaseWork(t *testing.T) {
	invalid := []string{
		"",
		"not-a-uuid",
		"00000000-0000-0000-0000-000000000000",
		"0F8FAD5B-D9CB-469F-A165-70867728950E",
		"0f8fad5b-d9cb-469f-a165-70867728950E",
		"0f8fad5bd9cb469fa16570867728950e",
		"{0f8fad5b-d9cb-469f-a165-70867728950e}",
		"urn:uuid:0f8fad5b-d9cb-469f-a165-70867728950e",
		" 0f8fad5b-d9cb-469f-a165-70867728950e",
		"0f8fad5b-d9cb-469f-a165-70867728950e ",
		"0f8fad5b-d9cb-469f-a165_70867728950e",
		"0f8fad5b-d9cb-469f-a165-70867728950g",
		"0f8fad5b-d9cb-469f-a165-70867728950",
	}
	now := time.Unix(1_800_000_000, 0)
	for _, id := range invalid {
		for name, op := range map[string]func(*Store) error{
			"DisableUser": func(s *Store) error { return s.DisableUser(context.Background(), id, now) },
			"EnableUser":  func(s *Store) error { return s.EnableUser(context.Background(), id, now) },
		} {
			db := &countingDB{}
			if err := op(newStore(db)); err != ErrInvalidRequest {
				t.Errorf("%s(%q) = %v, want invalid_request", name, id, err)
			}
			if db.begins != 0 || db.queries != 0 {
				t.Errorf("%s(%q): %d BeginTx, %d QueryRow before rejection", name, id, db.begins, db.queries)
			}
		}
	}
	// A canonical ID passes the guard and reaches the database (here a refused BeginTx: identity_unavailable).
	const valid = "0f8fad5b-d9cb-469f-a165-70867728950e"
	if !canonicalUserID(valid) {
		t.Fatal("canonical UUID rejected")
	}
	for name, op := range map[string]func(*Store) error{
		"DisableUser": func(s *Store) error { return s.DisableUser(context.Background(), valid, now) },
		"EnableUser":  func(s *Store) error { return s.EnableUser(context.Background(), valid, now) },
	} {
		db := &countingDB{}
		if err := op(newStore(db)); !errors.Is(err, ErrUnavailable) || db.begins != 1 {
			t.Errorf("%s(valid) = %v after %d BeginTx, want identity_unavailable after one BeginTx", name, err, db.begins)
		}
	}
}

// --- U-IDN-AUDIT (G011 Plan §9, r1 item 2): the bounded identity recorder and the revocation gap decision ------------

// recDB is a storeDB whose transactions are recTx values; it records every BeginTx context.
type recDB struct {
	mu     sync.Mutex
	begins []context.Context
	tx     *recTx
	begin  error
}

func (d *recDB) BeginTx(ctx context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.begins = append(d.begins, ctx)
	if d.begin != nil {
		return nil, d.begin
	}
	return d.tx, nil
}

func (d *recDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return recRow{err: errors.New("recDB: no reader")}
}

// recTx records the contexts of its statements. noRow makes the context statement report prior context; execErr fails
// the append; blockRollback blocks Rollback until its context ends.
type recTx struct {
	pgx.Tx
	mu            sync.Mutex
	stmtCtx       []context.Context
	canceled      int // statements (and the rollback) whose context was already done when called
	execs         int
	conns         int
	commits       int
	rollbackCtx   context.Context
	rollbackTook  time.Duration
	noRow         bool
	execErr       error
	blockRollback bool
}

type recRow struct{ err error }

func (r recRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for _, d := range dest {
		if p, ok := d.(*string); ok {
			*p = ""
		}
	}
	return nil
}

func (t *recTx) QueryRow(ctx context.Context, _ string, _ ...any) pgx.Row {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.note(ctx)
	if t.noRow {
		return recRow{err: pgx.ErrNoRows}
	}
	return recRow{}
}

func (t *recTx) note(ctx context.Context) {
	t.stmtCtx = append(t.stmtCtx, ctx)
	if ctx.Err() != nil {
		t.canceled++
	}
}

func (t *recTx) Exec(ctx context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.note(ctx)
	t.execs++
	return pgconn.CommandTag{}, t.execErr
}

func (t *recTx) Commit(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.note(ctx)
	t.commits++
	return nil
}

func (t *recTx) Rollback(ctx context.Context) error {
	start := time.Now()
	t.mu.Lock()
	t.rollbackCtx = ctx
	if ctx.Err() != nil {
		t.canceled++
	}
	block := t.blockRollback
	t.mu.Unlock()
	if block {
		<-ctx.Done()
	}
	t.mu.Lock()
	t.rollbackTook = time.Since(start)
	t.mu.Unlock()
	return nil
}

func (t *recTx) Conn() *pgx.Conn {
	t.mu.Lock()
	t.conns++
	t.mu.Unlock()
	return nil
}

func auditEvent() audit.Event {
	subject, _ := audit.ParseUUID("0f8fad5b-d9cb-469f-a165-70867728950e")
	return audit.Event{Action: audit.AuthSignIn, Outcome: audit.Denied, Reason: audit.ReasonUserDisabled, Subject: subject}
}

func TestIdentityAuditRecorder(t *testing.T) {
	const subject = "0f8fad5b-d9cb-469f-a165-70867728950e"

	t.Run("statement budget and detached caller cancellation", func(t *testing.T) {
		db := &recDB{tx: &recTx{}}
		parent, cancel := context.WithCancel(context.Background())
		cancel() // the caller's request already ended: recording still runs
		start := time.Now()
		if err := newStore(db).recordAuditEvents(parent, subject, auditEvent()); err != nil {
			t.Fatalf("recordAuditEvents: %v", err)
		}
		if len(db.begins) != 1 || db.tx.execs != 1 || db.tx.commits != 1 {
			t.Fatalf("begins %d, execs %d, commits %d: want one transaction with one append", len(db.begins), db.tx.execs, db.tx.commits)
		}
		if db.tx.canceled != 0 {
			t.Errorf("%d statements ran on a canceled context: caller cancellation must not cancel recording", db.tx.canceled)
		}
		for i, ctx := range append([]context.Context{db.begins[0]}, db.tx.stmtCtx...) {
			d, ok := ctx.Deadline()
			if !ok || d.Sub(start) > auditRecordBudget+time.Second || d.Sub(start) < auditRecordBudget-time.Second {
				t.Errorf("statement %d deadline %v after start, want auditRecordBudget (%v)", i, d.Sub(start), auditRecordBudget)
			}
		}
		rc := db.tx.rollbackCtx
		if rc == nil {
			t.Fatal("no deferred rollback")
		}
		if d, ok := rc.Deadline(); !ok || d.Sub(start) > auditCleanupBudget+auditRecordBudget+time.Second {
			t.Errorf("rollback context has no auditCleanupBudget deadline (%v, %v)", d, ok)
		}
	})

	t.Run("blocked rollback bounded by the cleanup budget", func(t *testing.T) {
		db := &recDB{tx: &recTx{execErr: errors.New("append refused"), blockRollback: true}}
		if err := newStore(db).recordAuditEvents(context.Background(), subject, auditEvent()); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("recordAuditEvents with a refused append = %v, want identity_unavailable", err)
		}
		if db.tx.commits != 0 {
			t.Error("a failed append must not commit")
		}
		took := db.tx.rollbackTook
		if took < auditCleanupBudget-200*time.Millisecond || took > auditCleanupBudget+time.Second {
			t.Errorf("blocked rollback returned after %v, want about auditCleanupBudget (%v)", took, auditCleanupBudget)
		}
		if err := db.tx.rollbackCtx.Err(); err == nil {
			t.Error("the rollback context must end with its budget")
		}
	})

	t.Run("prior context discards and appends nothing", func(t *testing.T) {
		db := &recDB{tx: &recTx{noRow: true}}
		if err := newStore(db).recordAuditEvents(context.Background(), "", auditEvent()); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("prior context = %v, want identity_unavailable", err)
		}
		if db.tx.execs != 0 || db.tx.commits != 0 || db.tx.conns != 1 {
			t.Errorf("execs %d, commits %d, discards %d: want no append, no commit, one discard", db.tx.execs, db.tx.commits, db.tx.conns)
		}
	})

	t.Run("revocation gap decision", func(t *testing.T) {
		const session = "1f8fad5b-d9cb-469f-a165-70867728950e"
		for _, c := range []struct {
			name      string
			gap       []string
			err       error
			appendErr error
			want      error
			begins    int
			diag      string
		}{
			{"no marker: nothing", nil, nil, nil, nil, 0, ""},
			{"no marker, failed commit: the failure", nil, ErrUnavailable, nil, ErrUnavailable, 0, ""},
			{"marker, committed: G6 recorded", []string{audit.GapSessionRevoked}, nil, nil, nil, 1, `"independent":"recorded"`},
			{"marker, committed: G6 failed", []string{audit.GapSessionRevoked}, nil, errors.New("refused"), nil, 1, `"independent":"failed"`},
			{"marker, commit failed or unknown: not attempted", []string{audit.GapSessionRevoked}, ErrUnavailable, nil, ErrUnavailable, 0,
				`"independent":"not_attempted"`},
		} {
			t.Run(c.name, func(t *testing.T) {
				var buf bytes.Buffer
				db := &recDB{tx: &recTx{execErr: c.appendErr}}
				s := newStore(db)
				s.logger = slog.New(slog.NewJSONHandler(&buf, nil))
				err := s.afterRevocation(context.Background(), subject, revokeLogout, audit.ResourceSession, session, c.gap, c.err)
				if !errors.Is(err, c.want) || (c.want == nil && err != nil) {
					t.Errorf("afterRevocation = %v, want %v", err, c.want)
				}
				if len(db.begins) != c.begins {
					t.Errorf("%d independent transactions, want %d", len(db.begins), c.begins)
				}
				out := buf.String()
				if c.diag == "" && out != "" {
					t.Errorf("unexpected diagnostic %s", out)
				}
				if c.diag != "" && (!strings.Contains(out, `"msg":"audit.gap"`) || !strings.Contains(out, c.diag) ||
					!strings.Contains(out, `"operation":"logout"`) || strings.Count(out, "\n") != 1) {
					t.Errorf("diagnostic %s, want one audit.gap with %s", out, c.diag)
				}
			})
		}
	})

	t.Run("sign-in failure classification", func(t *testing.T) {
		for err, want := range map[error][2]string{
			ErrTransactionExpired: {"denied", "transaction_expired"}, ErrLoginCanceled: {"denied", "login_canceled"},
			ErrCodeRejected: {"denied", "code_rejected"}, ErrImpersonationRefused: {"denied", "impersonation_refused"},
			ErrMFAUnverified: {"denied", "mfa_unverified"}, ErrStaleAuthentication: {"denied", "stale_authentication"},
			ErrIdentityConflict: {"denied", "identity_conflict"}, ErrUserDisabled: {"denied", "user_disabled"},
			ErrProviderError: {"failed", "provider_error"}, ErrProviderTimeout: {"failed", "provider_timeout"},
			ErrNamespaceMismatch: {"failed", "identity_namespace_mismatch"}, ErrUnavailable: {"failed", "identity_unavailable"},
			errors.New("anything"): {"failed", "identity_unavailable"},
		} {
			e := signInFailure(err, subject)
			if string(e.Outcome) != want[0] || string(e.Reason) != want[1] || e.Action != audit.AuthSignIn {
				t.Errorf("%v: %s/%s, want %s/%s", err, e.Outcome, e.Reason, want[0], want[1])
			}
			if attributed := e.Subject != (audit.UUID{}); attributed != (want[1] == "user_disabled" || want[1] == "identity_conflict") {
				t.Errorf("%v: subject attributed = %v", err, attributed)
			}
		}
	})
}
