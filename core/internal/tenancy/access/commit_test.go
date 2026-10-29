package access

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// U-COMMIT (G007 attempt 2, C2; Codex review 1bed0ddb §C2): a failed COMMIT reports a definitive abort only when one
// is established (COMMIT unsent, a ROLLBACK reply, or an ordinary command-severity ERROR); a session-ending FATAL, a
// PANIC, a missing or unrecognized severity and a lost reply are commit_outcome_unknown. Every case is the public code
// unavailable with no driver text, and the transaction is attempted exactly once.

// commitCanary is driver text that must never reach an error or an access.decision record.
const commitCanary = "Cx7f3acommitdrivertext"

// unsentErr is a driver error that proves the COMMIT was never sent (pgconn.SafeToRetry).
type unsentErr struct{}

func (unsentErr) Error() string {
	return "commit not sent " + commitCanary
}

func (unsentErr) SafeToRetry() bool {
	return true
}

func pgCommitErr(unlocalized, localized, code string) *pgconn.PgError {
	return &pgconn.PgError{SeverityUnlocalized: unlocalized, Severity: localized, Code: code, Message: "server text " + commitCanary}
}

type commitCase struct {
	name     string
	canceled bool
	err      error
	want     Reason
}

func commitCases() []commitCase {
	return []commitCase{
		{"ordinary ERROR rejecting COMMIT (serialization failure)", false, pgCommitErr("ERROR", "ERROR", "40001"), reasonAborted},
		{"ordinary ERROR rejecting COMMIT (deferred constraint)", false, pgCommitErr("ERROR", "ERROR", "23505"), reasonAborted},
		{"wrapped ordinary ERROR (deadlock)", false, fmt.Errorf("commit: %w", pgCommitErr("ERROR", "ERROR", "40P01")), reasonAborted},
		{"nonlocalized ERROR with a translated severity", false, pgCommitErr("ERROR", "FEHLER", "40001"), reasonAborted},
		{"localized ERROR when no nonlocalized severity was sent", false, pgCommitErr("", "ERROR", "40001"), reasonAborted},
		{"explicit ROLLBACK reply", false, pgx.ErrTxCommitRollback, reasonAborted},
		{"wrapped explicit ROLLBACK reply", false, fmt.Errorf("commit: %w", pgx.ErrTxCommitRollback), reasonAborted},
		{"FATAL 57P01 administrator shutdown", false, pgCommitErr("FATAL", "FATAL", "57P01"), reasonCommitOutcomeUnknown},
		{"wrapped FATAL 57P01", false, fmt.Errorf("commit: %w", pgCommitErr("FATAL", "FATAL", "57P01")), reasonCommitOutcomeUnknown},
		{"nonlocalized FATAL despite a localized ERROR", false, pgCommitErr("FATAL", "ERROR", "57P01"), reasonCommitOutcomeUnknown},
		{"PANIC 57P02 crash shutdown", false, pgCommitErr("PANIC", "PANIC", "57P02"), reasonCommitOutcomeUnknown},
		{"wrapped PANIC 57P02", false, fmt.Errorf("commit: %w", pgCommitErr("PANIC", "PANIC", "57P02")), reasonCommitOutcomeUnknown},
		{"no severity sent", false, pgCommitErr("", "", "40001"), reasonCommitOutcomeUnknown},
		{"unrecognized severity", false, pgCommitErr("WARNING", "WARNING", "40001"), reasonCommitOutcomeUnknown},
		{"translated severity only", false, pgCommitErr("", "FEHLER", "40001"), reasonCommitOutcomeUnknown},
		{"lowercase severity", false, pgCommitErr("error", "error", "40001"), reasonCommitOutcomeUnknown},
		{"lost reply: unexpected EOF", false, io.ErrUnexpectedEOF, reasonCommitOutcomeUnknown},
		{"lost reply: EOF", false, io.EOF, reasonCommitOutcomeUnknown},
		{"cancellation during COMMIT", true, context.Canceled, reasonCommitOutcomeUnknown},
		{"COMMIT unsent", false, unsentErr{}, reasonStoreUnavailable},
		{"COMMIT unsent after cancellation", true, unsentErr{}, reasonCanceled},
	}
}

func commitCtx(canceled bool) context.Context {
	if !canceled {
		return context.Background()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// TestCommitClassification: the classifier alone, for every case.
func TestCommitClassification(t *testing.T) {
	for _, c := range commitCases() {
		t.Run(c.name, func(t *testing.T) {
			f := commitFailure(commitCtx(c.canceled), c.err)
			if f == nil {
				t.Fatal("no failure")
			}
			if f.code != CodeUnavailable || f.field != "" || f.reason != c.want {
				t.Errorf("code %q field %q reason %q, want unavailable, no field, %q", f.code, f.field, f.reason, c.want)
			}
			err := f.err()
			if err.Error() != "access: unavailable" || CodeOf(err) != CodeUnavailable || strings.Contains(err.Error(), commitCanary) {
				t.Errorf("error %q", err)
			}
		})
	}
}

// commitTx is a transaction whose COMMIT fails with err; inTx calls only Commit and Rollback on it.
type commitTx struct {
	pgx.Tx
	err                error
	commits, rollbacks *atomic.Int64
}

func (t *commitTx) Commit(context.Context) error {
	t.commits.Add(1)
	return t.err
}

func (t *commitTx) Rollback(context.Context) error {
	t.rollbacks.Add(1)
	return nil
}

type commitDB struct {
	err                        error
	begins, commits, rollbacks atomic.Int64
}

func (d *commitDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	d.begins.Add(1)
	return &commitTx{err: d.err, commits: &d.commits, rollbacks: &d.rollbacks}, nil
}

// TestCommitFailureTransaction: through inTx, a failed COMMIT is attempted once (one BeginTx, one Commit, no retry),
// the deferred rollback still runs, and the access.decision record carries only the bounded reason.
func TestCommitFailureTransaction(t *testing.T) {
	for _, c := range commitCases() {
		t.Run(c.name, func(t *testing.T) {
			db := &commitDB{err: c.err}
			var logs bytes.Buffer
			s := newService(db, nil, slog.New(slog.NewJSONHandler(&logs, nil)))
			ctx := commitCtx(c.canceled)
			f := s.inTx(ctx, true, cleanupUnbounded, func(*trackedTx) *failure { return nil })
			if f == nil || f.code != CodeUnavailable || f.reason != c.want {
				t.Fatalf("failure %+v, want unavailable with reason %q", f, c.want)
			}
			if b, cm, rb := db.begins.Load(), db.commits.Load(), db.rollbacks.Load(); b != 1 || cm != 1 || rb != 1 {
				t.Errorf("BeginTx %d, Commit %d, Rollback %d: want exactly one attempt and the deferred rollback", b, cm, rb)
			}
			s.logGuarded(ctx, opRenameGroup, capGroupUpdate, f)
			if strings.Contains(logs.String(), commitCanary) {
				t.Fatalf("driver text in the record: %s", logs.String())
			}
			var rec map[string]any
			if err := json.Unmarshal(logs.Bytes(), &rec); err != nil {
				t.Fatal(err)
			}
			for k := range rec {
				switch k {
				case "time", "level", "msg", "operation", "capability", "outcome", "reason":
				default:
					t.Errorf("unexpected record field %q", k)
				}
			}
			if rec["msg"] != eventDecision || rec["level"] != "ERROR" || rec["operation"] != opRenameGroup ||
				rec["capability"] != string(capGroupUpdate) || rec["outcome"] != outcomeFailed || rec["reason"] != string(c.want) {
				t.Errorf("record %v", rec)
			}
		})
	}
}
