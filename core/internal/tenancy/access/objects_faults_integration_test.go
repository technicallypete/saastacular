//go:build integration

// B-OBJ-FAULTS (G009 Plan r1 §4.4, §5.2 FT/FP/FV/FD/FG/FX): every Put and Delete transition under injected and real
// faults against real PostgreSQL and MinIO, with the database and provider state inspected afterwards. Injected faults
// come from the hookDB (statement failure, COMMIT modes, step hooks), the provider spy and the storage stage seams;
// real faults are real MinIO objects, real blocked sources, a real pg_terminate_backend and a real frozen TCP path.
package access

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/objectstore"
	"core/internal/tenancy"
)

func (h *hookDB) setCommit(m commitMode) {
	h.mu.Lock()
	h.commit = m
	h.mu.Unlock()
}

func (h *hookDB) setFail(step string) {
	h.mu.Lock()
	h.failOn = step
	h.mu.Unlock()
}

// faultCase prepares a storage service with its own hookDB and spy over the real store, and the stage seams.
type faultSvc struct {
	svc  *Service
	h    *hookDB
	spy  *objSpy
	hook *objectHooks
}

func (e *objEnv) faultService(cfg ObjectConfig) *faultSvc {
	f := &faultSvc{h: &hookDB{pool: e.pool}, spy: &objSpy{inner: e.store}, hook: &objectHooks{}}
	f.svc = e.objService(f.h, f.spy, cfg)
	f.svc.hooks = f.hook
	return f
}

// putAs uploads through f as u (G016 A2: in a request scope of u's credential for org).
func (f *faultSvc) putAs(t *testing.T, u *itUser, org tenancy.OrganizationID, group tenancy.GroupID, size int64, body io.ReadCloser) (ObjectInfo, error) {
	t.Helper()
	rctx, end := beginAs(t, f.svc, context.Background(), u, org)
	defer end()
	tc := resolve(t, f.svc, rctx, org, tenancy.GroupID{})
	f.h.reset()
	return f.svc.PutObject(rctx, tc, PutObjectRequest{Group: group, Size: size, Body: body})
}

func (f *faultSvc) deleteAs(t *testing.T, u *itUser, org tenancy.OrganizationID, id tenancy.ObjectID) (DeleteResult, error) {
	t.Helper()
	rctx, end := beginAs(t, f.svc, context.Background(), u, org) // G016 A2
	defer end()
	tc := resolve(t, f.svc, rctx, org, tenancy.GroupID{})
	return f.svc.DeleteObject(rctx, tc, id)
}

// one returns the single new object of org since before (zero when none) and its state.
func (e *objEnv) one(t *testing.T, org tenancy.OrganizationID, before map[tenancy.ObjectID]string) (tenancy.ObjectID, string) {
	t.Helper()
	created := newObjects(before, e.objectsOf(t, org))
	if len(created) > 1 {
		t.Fatalf("%d new objects", len(created))
	}
	for id, st := range created {
		return id, st
	}
	return tenancy.ObjectID{}, ""
}

// closeCounter counts Close calls of an upload source.
type closeCounter struct {
	io.Reader
	n      atomic.Int64
	closer func() error
}

func (c *closeCounter) Close() error {
	c.n.Add(1)
	if c.closer != nil {
		return c.closer()
	}
	return nil
}

// repairDelete is the documented repair: an authorized Delete of a pending, ready or deleting object.
func (e *objEnv) repairDelete(t *testing.T, w *objOrderWorld, id tenancy.ObjectID) {
	t.Helper()
	res, err := e.faultService(e.ocfg).deleteAs(t, w.admin, w.org, id)
	if err != nil || res.Outcome != DeleteOutcomeDeleted || e.present(t, e.keyOf(t, w.org, id).key(t)) {
		t.Errorf("repair delete: %+v %v", res, err)
	}
}

func TestObjectsFaults(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()
	w := e.orderWorld(t, "faults")
	defer e.cleanupOrg(t, w.org)
	// G060 AC5: Group object deletion needs an explicit Group role, so the documented repairs run as the g1 Group admin
	// (the objOrderWorld{admin, org} form audit_integration_test.go already uses for its Group admin).
	repair := &objOrderWorld{admin: w.g1admin, org: w.org}
	data := payload(4096, "faults")

	type outcome struct {
		code   Code
		field  string
		reason string
		row    string // "-" none, or the state
		bytes  bool
	}
	observe := func(t *testing.T, before map[tenancy.ObjectID]string, err error) (tenancy.ObjectID, outcome) {
		t.Helper()
		id, st := e.one(t, w.org, before)
		o := outcome{code: CodeOf(err), row: "-"}
		var ae *Error
		if errors.As(err, &ae) {
			o.field = ae.Field
		}
		if r, ok := e.lastStorage(t)["reason"].(string); ok {
			o.reason = r
		}
		if !zero(id) {
			o.row = st
			o.bytes = e.present(t, e.keyOf(t, w.org, id).key(t))
		}
		return id, o
	}
	expect := func(t *testing.T, name string, got, want outcome) {
		t.Helper()
		if got != want {
			t.Errorf("%s: %+v, want %+v", name, got, want)
		}
		ordering(t, map[string]any{"case": name, "code": string(got.code), "reason": got.reason, "row": got.row, "bytes": got.bytes})
	}

	// --- T1 ---
	t.Run("FT-STMT a T1 statement error writes nothing and calls no provider", func(t *testing.T) {
		f := e.faultService(e.ocfg)
		before := e.objectsOf(t, w.org)
		f.hook.stage = func(op, stage string) {
			if stage == stageT1 {
				f.h.setFail("?") // the insert (the first stored-object statement of T1)
			}
		}
		_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
		_, got := observe(t, before, err)
		expect(t, "FT-STMT", got, outcome{code: CodeUnavailable, reason: "store_unavailable", row: "-"})
		if f.spy.calls.Load() != 0 {
			t.Error("provider called after a T1 failure")
		}
	})
	t.Run("FT-UNKNOWN an acknowledged-then-unknown T1 COMMIT: unavailable, no provider call, inert pending row (X2)", func(t *testing.T) {
		f := e.faultService(e.ocfg)
		before := e.objectsOf(t, w.org)
		f.hook.stage = func(op, stage string) {
			if stage == stageT1 {
				f.h.setCommit(commitLostReply)
			}
		}
		_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
		id, got := observe(t, before, err)
		expect(t, "FT-UNKNOWN", got, outcome{code: CodeUnavailable, reason: "commit_outcome_unknown", row: "pending"})
		if f.spy.calls.Load() != 0 {
			t.Error("provider called after an unknown T1 COMMIT")
		}
		e.repairDelete(t, repair, id)
	})
	t.Run("FT-FATAL a FATAL before the T1 COMMIT: nothing written", func(t *testing.T) {
		f := e.faultService(e.ocfg)
		before := e.objectsOf(t, w.org)
		f.hook.stage = func(op, stage string) {
			if stage != stageT1 {
				return
			}
			f.h.on("before", "COMMIT", func() {
				f.h.on("before", "COMMIT", nil)
				if _, err := e.setup.Exec(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity
WHERE usename = $1 AND state = 'idle in transaction' AND pid <> pg_backend_pid()`, e.cfg.RuntimeRole()); err != nil {
					t.Errorf("terminate: %v", err)
				}
			})
		}
		_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
		_, got := observe(t, before, err)
		if got.code != CodeUnavailable || got.row != "-" || f.spy.calls.Load() != 0 {
			t.Errorf("FT-FATAL: %+v, provider calls %d", got, f.spy.calls.Load())
		}
		ordering(t, map[string]any{"case": "FT-FATAL", "code": string(got.code), "reason": got.reason, "row": got.row})
	})

	// --- P2 ---
	t.Run("FP-ERR a provider error before send: compensated, unavailable", func(t *testing.T) {
		f := e.faultService(e.ocfg)
		before := e.objectsOf(t, w.org)
		f.spy.set(func(op string, _ objectstore.Key) error {
			if op == "put" {
				return objectstore.ErrUnavailable
			}
			return nil
		})
		_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
		_, got := observe(t, before, err)
		expect(t, "FP-ERR", got, outcome{code: CodeUnavailable, reason: "provider_outcome_unknown", row: "pending"})
	})
	t.Run("FP-TIMEOUT the provider stored the bytes, then the reply was lost: compensated", func(t *testing.T) {
		f := e.faultService(e.ocfg)
		before := e.objectsOf(t, w.org)
		f.spy.putOverride = func(ctx context.Context, key objectstore.Key, body io.Reader, size int64, ct string, b objectstore.Binding) error {
			if err := e.store.Put(ctx, key, body, size, ct, b); err != nil {
				return err
			}
			return objectstore.ErrOutcomeUnknown
		}
		_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
		_, got := observe(t, before, err)
		expect(t, "FP-TIMEOUT", got, outcome{code: CodeUnavailable, reason: "provider_outcome_unknown", row: "pending"})
	})
	for _, c := range []struct {
		name string
		body []byte
	}{{"FP-SHORT", data[:len(data)-1]}, {"FP-LONG", append(append([]byte{}, data...), 'x')}} {
		t.Run(c.name+" a mismatched source: invalid_input object_size_mismatch, compensated", func(t *testing.T) {
			f := e.faultService(e.ocfg)
			before := e.objectsOf(t, w.org)
			body := &closeCounter{Reader: bytes.NewReader(c.body)}
			_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), body)
			_, got := observe(t, before, err)
			expect(t, c.name, got, outcome{code: CodeInvalidInput, field: "object_size_mismatch", reason: "compensated", row: "pending"})
			if body.n.Load() != 1 {
				t.Errorf("Body closed %d times", body.n.Load())
			}
		})
	}
	t.Run("FP-LONG-SUCCESS a provider that completes despite the overlong source is still compensated", func(t *testing.T) {
		f := e.faultService(e.ocfg)
		before := e.objectsOf(t, w.org)
		f.spy.putOverride = func(ctx context.Context, key objectstore.Key, body io.Reader, size int64, ct string, b objectstore.Binding) error {
			_, _ = io.Copy(io.Discard, body) // the exact reader errors on the extra byte
			return e.store.Put(ctx, key, bytes.NewReader(data), size, ct, b)
		}
		long := append(append([]byte{}, data...), 'x')
		_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(long)))
		_, got := observe(t, before, err)
		expect(t, "FP-LONG-SUCCESS", got, outcome{code: CodeInvalidInput, field: "object_size_mismatch", reason: "compensated", row: "pending"})
	})
	t.Run("FP-BLOCK1 and FP-BLOCKN blocked sources end within the transfer deadline", func(t *testing.T) {
		short := e.ocfg
		short.TransferTimeout = time.Second
		for _, name := range []string{"FP-BLOCK1", "FP-BLOCKN"} {
			f := e.faultService(short)
			before := e.objectsOf(t, w.org)
			base := runtime.NumGoroutine()
			pr, pw := io.Pipe()
			if name == "FP-BLOCKN" {
				go func() { _, _ = pw.Write(data) }() // all Size bytes, then neither EOF nor more
			}
			body := &closeCounter{Reader: pr, closer: pr.Close}
			startAt := time.Now()
			_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), body)
			elapsed := time.Since(startAt)
			_, got := observe(t, before, err)
			expect(t, name, got, outcome{code: CodeUnavailable, reason: "canceled", row: "pending"})
			if elapsed > short.TransferTimeout+sourceReapGrace+storageDecisionBudget || body.n.Load() != 1 {
				t.Errorf("%s: %v, closes %d", name, elapsed, body.n.Load())
			}
			_ = pw.Close()
			waitGoroutines(t, base+2)
		}
	})
	t.Run("FP-ZERO an empty upload succeeds; data for size 0 is refused before any database or provider work", func(t *testing.T) {
		f := e.faultService(e.ocfg)
		info, err := f.putAs(t, w.m1, w.org, w.g1, 0, io.NopCloser(strings.NewReader("")))
		if err != nil || info.Size != 0 {
			t.Fatalf("empty upload: %v", err)
		}
		if pi, err := e.store.Metadata(ctx, e.keyOf(t, w.org, info.ID).key(t)); err != nil || pi.Size != 0 {
			t.Errorf("stored length: %+v %v", pi, err)
		}
		f = e.faultService(e.ocfg)
		before := e.objectsOf(t, w.org)
		_, err = f.putAs(t, w.m1, w.org, w.g1, 0, io.NopCloser(strings.NewReader("x")))
		_, got := observe(t, before, err)
		if got.code != CodeInvalidInput || got.field != "object_size_mismatch" || got.row != "-" || f.spy.calls.Load() != 0 {
			t.Errorf("size 0 with data: %+v", got)
		}
		if begins, _ := f.h.counts(); begins != 0 {
			t.Errorf("size 0 with data opened %d transactions", begins)
		}
	})

	// --- V ---
	t.Run("FV verification mismatch and unavailability: never ready, compensated", func(t *testing.T) {
		f := e.faultService(e.ocfg)
		before := e.objectsOf(t, w.org)
		f.spy.putOverride = func(ctx context.Context, key objectstore.Key, body io.Reader, size int64, ct string, b objectstore.Binding) error {
			b.Group = ""
			return e.store.Put(ctx, key, body, size, ct, b)
		}
		_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
		_, got := observe(t, before, err)
		expect(t, "FV-BINDING", got, outcome{code: CodeInternal, reason: "binding_mismatch", row: "pending"})

		f = e.faultService(e.ocfg)
		before = e.objectsOf(t, w.org)
		var heads atomic.Int64
		f.spy.set(func(op string, _ objectstore.Key) error {
			if op == "head" && heads.Add(1) == 1 {
				return objectstore.ErrUnavailable
			}
			return nil
		})
		_, err = f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
		_, got = observe(t, before, err)
		expect(t, "FV-UNAVAILABLE", got, outcome{code: CodeUnavailable, reason: "provider_unavailable", row: "pending"})
	})

	// --- T2 and R ---
	t.Run("FT2 definitive abort, unknown-but-committed, unknown-and-reconcile-unavailable, unknown-not-committed", func(t *testing.T) {
		for _, c := range []struct {
			name      string
			mode      commitMode
			failR     bool
			want      outcome
			wantReady bool
		}{
			{"FT2-ABORT", commitDefinitive, false, outcome{code: CodeUnavailable, reason: "aborted", row: "pending"}, false},
			{"FT2-UNKNOWN-COMMITTED", commitLostReply, false, outcome{reason: "reconciled_ready", row: "ready", bytes: true}, true},
			{"FT2-UNKNOWN-RECONCILE-UNAVAILABLE", commitLostReply, true, outcome{code: CodeUnavailable, reason: "commit_outcome_unknown", row: "ready", bytes: true}, true},
			{"FT2-UNKNOWN-NOT-COMMITTED", commitLostNoCommit, false, outcome{code: CodeUnavailable, reason: "commit_outcome_unknown", row: "pending"}, false},
		} {
			f := e.faultService(e.ocfg)
			before := e.objectsOf(t, w.org)
			f.hook.stage = func(op, stage string) {
				switch stage {
				case stageT2:
					f.h.setCommit(c.mode)
				case stageReconcile:
					f.h.setCommit(commitNormal)
					if c.failR {
						f.h.setFail("?")
					}
				}
			}
			info, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
			id, got := observe(t, before, err)
			expect(t, c.name, got, c.want)
			if c.wantReady && err == nil && info.ID != id {
				t.Errorf("%s: info %+v", c.name, info)
			}
			e.repairDelete(t, repair, id)
		}
	})

	// --- cancellation and simulated interruption at each stage ---
	t.Run("FX-CANCEL cancellation at each stage never publishes", func(t *testing.T) {
		for _, stage := range []string{stageT1, stageProvider, stageVerify, stageT2} {
			f := e.faultService(e.ocfg)
			before := e.objectsOf(t, w.org)
			rctx, end := beginAs(t, f.svc, ctx, w.m1, w.org) // G016 A2
			tc := resolve(t, f.svc, rctx, w.org, tenancy.GroupID{})
			cctx, cancel := context.WithCancel(rctx)
			f.hook.stage = func(op, s string) {
				if s == stage {
					cancel()
				}
			}
			_, err := f.svc.PutObject(cctx, tc, PutObjectRequest{Group: w.g1, Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
			end()
			id, got := observe(t, before, err)
			if got.code != CodeUnavailable || got.row == "ready" || got.bytes {
				t.Errorf("cancel at %s: %+v", stage, got)
			}
			ordering(t, map[string]any{"case": "FX-CANCEL", "stage": stage, "code": string(got.code), "reason": got.reason, "row": got.row, "bytes": got.bytes})
			if !zero(id) {
				e.repairDelete(t, repair, id)
			}
		}
	})
	t.Run("FX-CRASH a simulated process interruption after each stage; an authorized Delete repairs", func(t *testing.T) {
		for _, stage := range []string{stageProvider, stageVerify, stageT2, stageCompensate} {
			f := e.faultService(e.ocfg)
			before := e.objectsOf(t, w.org)
			if stage == stageCompensate {
				f.spy.set(func(op string, _ objectstore.Key) error {
					if op == "head" {
						return objectstore.ErrUnavailable // V fails, so compensation runs (and is interrupted)
					}
					return nil
				})
			}
			f.hook.crash = func(op, s string) bool { return s == stage }
			_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
			id, got := observe(t, before, err)
			if got.row != "pending" || CodeOf(err) != CodeUnavailable {
				t.Errorf("crash after %s: %+v", stage, got)
			}
			ordering(t, map[string]any{"case": "FX-CRASH", "stage": stage, "row": got.row, "bytes": got.bytes})
			e.repairDelete(t, repair, id)
		}
	})
	t.Run("FC compensation failure leaves inaccessible bytes under a non-ready row (X3); Delete repairs", func(t *testing.T) {
		f := e.faultService(e.ocfg)
		before := e.objectsOf(t, w.org)
		f.spy.set(func(op string, _ objectstore.Key) error {
			if op == "delete" {
				return objectstore.ErrUnavailable
			}
			return nil
		})
		long := append(append([]byte{}, data...), 'x')
		f.spy.putOverride = func(ctx context.Context, key objectstore.Key, body io.Reader, size int64, ct string, b objectstore.Binding) error {
			_, _ = io.Copy(io.Discard, body)
			return e.store.Put(ctx, key, bytes.NewReader(data), size, ct, b)
		}
		_, err := f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(long)))
		id, got := observe(t, before, err)
		expect(t, "FC", got, outcome{code: CodeInvalidInput, field: "object_size_mismatch", reason: "compensation_failed", row: "pending", bytes: true})
		rctx, end := beginAs(t, e.osvc, ctx, w.g1admin, w.org) // G016 A2; G060 AC5: the g1 Group admin may read g1, so not_found is the lifecycle denial
		tc := resolve(t, e.osvc, rctx, w.org, tenancy.GroupID{})
		if _, err := e.osvc.ObjectMetadata(rctx, tc, id); CodeOf(err) != CodeNotFound {
			t.Errorf("the leftover bytes are accessible: %v", err)
		}
		end()
		e.repairDelete(t, repair, id)
	})

	// --- Delete ---
	t.Run("FD D1 unknown, D2/D3 failure and D4 failure are deletion_pending; a retry completes", func(t *testing.T) {
		for _, c := range []struct {
			name, failOp, failStep string
			mode                   commitMode
			wantErr                bool
		}{
			{"FD-D1-UNKNOWN", "", "", commitLostReply, true},
			{"FD-D2", "delete", "", commitNormal, false},
			{"FD-D3", "head", "", commitNormal, false},
			{"FD-D4", "", "W", commitNormal, false},
		} {
			info, err := e.put(t, e.osvc, w.m1, w.org, w.g1, payload(300, c.name))
			if err != nil {
				t.Fatal(err)
			}
			f := e.faultService(e.ocfg)
			f.hook.stage = func(op, stage string) {
				switch {
				case stage == stageT1 && c.mode != commitNormal:
					f.h.setCommit(c.mode)
				case stage == stageProvider:
					f.h.setCommit(commitNormal)
				case stage == stageT2 && c.failStep != "":
					f.h.setFail(c.failStep)
				}
			}
			if c.failOp != "" {
				op := c.failOp
				f.spy.set(func(o string, _ objectstore.Key) error {
					if o == op {
						return objectstore.ErrUnavailable
					}
					return nil
				})
			}
			res, err := f.deleteAs(t, w.g1admin, w.org, info.ID) // G060 AC5: the g1 Group admin deletes
			st := e.objectsOf(t, w.org)[info.ID]
			switch {
			case c.wantErr && (CodeOf(err) != CodeUnavailable || st != "deleting"):
				t.Errorf("%s: %v, row %s", c.name, err, st)
			case !c.wantErr && (err != nil || res.Outcome != DeleteOutcomeDeletionPending || st != "deleting" || e.lastStorage(t)["physical"] != "pending"):
				t.Errorf("%s: %+v %v, row %s", c.name, res, err, st)
			}
			ordering(t, map[string]any{"case": c.name, "result": string(res.Outcome), "code": string(CodeOf(err)), "row": st})
			retry, err := e.faultService(e.ocfg).deleteAs(t, w.g1admin, w.org, info.ID)
			if err != nil || retry.Outcome != DeleteOutcomeDeleted || e.objectsOf(t, w.org)[info.ID] != "deleted" ||
				e.present(t, e.keyOf(t, w.org, info.ID).key(t)) {
				t.Errorf("%s retry: %+v %v", c.name, retry, err)
			}
		}
	})

	// --- Get ---
	t.Run("FG missing bytes, provider unavailable and integrity mismatch", func(t *testing.T) {
		info, err := e.put(t, e.osvc, w.m1, w.org, w.g1, data)
		if err != nil {
			t.Fatal(err)
		}
		k := e.keyOf(t, w.org, info.ID).key(t)
		get := func(f *faultSvc) (io.ReadCloser, error) {
			rctx, end := beginAs(t, f.svc, ctx, w.m2, w.org) // G016 A2
			defer end()
			tc := resolve(t, f.svc, rctx, w.org, tenancy.GroupID{})
			_, rc, err := f.svc.GetObject(rctx, tc, info.ID)
			return rc, err
		}
		f := e.faultService(e.ocfg)
		f.spy.set(func(op string, _ objectstore.Key) error {
			if op == "get" {
				return objectstore.ErrUnavailable
			}
			return nil
		})
		if _, err := get(f); CodeOf(err) != CodeUnavailable || e.lastStorage(t)["reason"] != "provider_unavailable" {
			t.Errorf("FG-UNAVAILABLE: %v", err)
		}
		// Same length and binding, other content: delivered, then internal at EOF.
		other := payload(len(data), "other")
		if err := e.store.Put(ctx, k, bytes.NewReader(other), int64(len(other)), "application/octet-stream",
			objectstore.Binding{Object: info.ID.String(), Organization: w.org.String(), Group: w.g1.String()}); err != nil {
			t.Fatal(err)
		}
		rc, err := get(e.faultService(e.ocfg))
		if err != nil {
			t.Fatalf("FG-INTEGRITY: %v", err)
		}
		_, rerr := io.ReadAll(rc)
		_ = rc.Close()
		if CodeOf(rerr) != CodeInternal {
			t.Errorf("FG-INTEGRITY: %v, want internal at EOF", rerr)
		}
		// Bytes removed behind the service: not_found with the error-level reason object_bytes_missing.
		if err := e.store.Delete(ctx, k); err != nil {
			t.Fatal(err)
		}
		if _, err := get(e.faultService(e.ocfg)); CodeOf(err) != CodeNotFound {
			t.Errorf("FG-MISSING: %v", err)
		}
		rec := e.lastStorage(t)
		if rec["reason"] != "object_bytes_missing" || rec["level"] != "error" || rec["outcome"] != "failed" {
			t.Errorf("FG-MISSING record %v", rec)
		}
		ordering(t, map[string]any{"case": "FG", "unavailable": "unavailable", "integrity": "internal_at_eof", "missing": "not_found"})
	})
	t.Run("pending and deleting objects never serve new reads or signs", func(t *testing.T) {
		before := e.objectsOf(t, w.org)
		f := e.faultService(e.ocfg)
		f.hook.crash = func(op, s string) bool { return s == stageT2 }
		_, _ = f.putAs(t, w.m1, w.org, w.g1, int64(len(data)), io.NopCloser(bytes.NewReader(data)))
		id, st := e.one(t, w.org, before)
		if st != "pending" || !e.present(t, e.keyOf(t, w.org, id).key(t)) {
			t.Fatalf("setup: %s", st)
		}
		check := func(what string) {
			rctx, end := beginAs(t, e.osvc, ctx, w.g1admin, w.org) // G016 A2; G060 AC5: the g1 Group admin may read g1, so not_found is the lifecycle denial
			defer end()
			tc := resolve(t, e.osvc, rctx, w.org, tenancy.GroupID{})
			calls := e.spy.calls.Load()
			_, err1 := e.osvc.ObjectMetadata(rctx, tc, id)
			_, _, err2 := e.osvc.GetObject(rctx, tc, id)
			_, err3 := e.osvc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: id, DefaultTTL: true})
			if CodeOf(err1) != CodeNotFound || CodeOf(err2) != CodeNotFound || CodeOf(err3) != CodeNotFound || e.spy.calls.Load() != calls {
				t.Errorf("%s object served: %v %v %v", what, err1, err2, err3)
			}
		}
		check("pending")
		if _, err := e.setup.Exec(ctx, `UPDATE app.stored_objects SET state = 'ready', sha256 = decode(repeat('aa', 32), 'hex') WHERE id = $1::text::uuid`, id.String()); err == nil {
			t.Error("the superuser fixture bypassed the lifecycle guard")
		}
		e.repairDelete(t, repair, id)
	})

	t.Run("FX-STALL a frozen database path ends within the decision and cleanup budgets", func(t *testing.T) {
		fw := newFreezer(t, fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port))
		pc, err := pgxpool.ParseConfig("")
		if err != nil {
			t.Fatal(err)
		}
		host, port := fw.addr()
		pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = host, port, e.cfg.Database
		pc.ConnConfig.User, pc.ConnConfig.Password = e.cfg.RuntimeRole(), e.pw
		pc.ConnConfig.ConnectTimeout = 5 * time.Second
		pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
		pc.MaxConns = 1
		pool, err := pgxpool.NewWithConfig(ctx, pc)
		if err != nil {
			t.Fatal(err)
		}
		defer pool.Close()
		h := &hookDB{pool: pool}
		svc := e.objService(h, e.spy, e.ocfg)
		rctx, end := beginAs(t, svc, ctx, w.m1, w.org) // G016 A2
		defer end()
		tc := resolve(t, svc, rctx, w.org, tenancy.GroupID{})
		var pid uint32
		if err := pool.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
			t.Fatal(err)
		}
		before := e.objectsOf(t, w.org)
		base := runtime.NumGoroutine()
		h.on("after", "C", fw.freeze) // freeze after the storage write transaction's first statement
		startAt := time.Now()
		_, err = svc.PutObject(rctx, tc, PutObjectRequest{Group: w.g1, Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
		elapsed := time.Since(startAt)
		fw.thaw()
		h.on("after", "C", nil)
		if CodeOf(err) != CodeUnavailable || elapsed > storageDecisionBudget+storageCleanupBudget+5*time.Second {
			t.Errorf("frozen path: %v after %v", err, elapsed)
		}
		if id, st := e.one(t, w.org, before); !zero(id) {
			t.Errorf("T1 committed a row (%s) on the frozen path", st)
		}
		var pid2 uint32
		if err := pool.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid2); err != nil || pid2 == pid {
			t.Errorf("the stalled connection was reused (pid %d -> %d, %v)", pid, pid2, err)
		}
		fw.close()
		waitGoroutines(t, base)
		// An existing G007 guarded read on an unfrozen pool is unaffected.
		grctx, gend := beginAs(t, e.svc, ctx, w.owner, w.org) // G016 A2
		defer gend()
		if _, err := e.svc.GetOrganization(grctx, resolve(t, e.svc, grctx, w.org, tenancy.GroupID{})); err != nil {
			t.Errorf("G007 read: %v", err)
		}
		ordering(t, map[string]any{"case": "FX-STALL", "code": string(CodeOf(err)), "elapsed_ms": elapsed.Milliseconds()})
	})
}

// freezer is a test-local TCP forwarder (stdlib only) between a pool and PostgreSQL whose traffic can be frozen: while
// frozen, bytes are held in both directions.
type freezer struct {
	l      net.Listener
	target string
	mu     sync.Mutex
	frozen bool
	cond   *sync.Cond
	conns  []net.Conn
	wg     sync.WaitGroup
}

func newFreezer(t *testing.T, target string) *freezer {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &freezer{l: l, target: target}
	f.cond = sync.NewCond(&f.mu)
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			s, err := net.DialTimeout("tcp", target, 5*time.Second)
			if err != nil {
				_ = c.Close()
				continue
			}
			f.mu.Lock()
			f.conns = append(f.conns, c, s)
			f.mu.Unlock()
			f.wg.Add(2)
			go f.pipe(c, s)
			go f.pipe(s, c)
		}
	}()
	return f
}

func (f *freezer) addr() (string, uint16) {
	a := f.l.Addr().(*net.TCPAddr)
	return a.IP.String(), uint16(a.Port)
}

func (f *freezer) pipe(dst, src net.Conn) {
	defer f.wg.Done()
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			f.mu.Lock()
			for f.frozen {
				f.cond.Wait()
			}
			f.mu.Unlock()
			if _, werr := dst.Write(buf[:n]); werr != nil {
				_ = src.Close()
				return
			}
		}
		if err != nil {
			_ = dst.Close()
			return
		}
	}
}

func (f *freezer) freeze() {
	f.mu.Lock()
	f.frozen = true
	f.mu.Unlock()
}

func (f *freezer) thaw() {
	f.mu.Lock()
	f.frozen = false
	f.cond.Broadcast()
	f.mu.Unlock()
}

func (f *freezer) close() {
	_ = f.l.Close()
	f.thaw()
	f.mu.Lock()
	for _, c := range f.conns {
		_ = c.Close()
	}
	f.mu.Unlock()
	f.wg.Wait()
}
