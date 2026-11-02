//go:build integration

// G074 B-PROFILE-IT, the account avatar lifecycle (Plan r3.1 §3.8, §5; lane L2): TestProfileAvatar against the access
// integration environment with the private MinIO of setupObjects (the scoped credential as infra/minio/init.sh
// provisions it), real PostgreSQL at version 23 and the G009 seams (hookDB.setCommit/setFail, objSpy.set,
// objectHooks.stage/crash). EP-2 (A6, A7) first: PRO-AV-4a/4b (fail-closed reconcile of an unknown publish COMMIT),
// PRO-AV-7 (persistent provider Delete fault), PRO-AV-8 (exhausted 20-row sweep cap), PRO-AV-9 (crashed upload and the
// lease), PRO-AV-11 (NULL-reference gate: the statements, concurrent first-ever Sets), PRO-AV-12 (cleanup failure and an
// abandoned pending row after removal) and PRO-AV-10 (I-CUR), then the write-ahead upload marker (review P1): PRO-AV-13
// (a writer paused past the lease gets no Put) and PRO-AV-14 (an issued Put with an unproven outcome, a live-context
// unknown and a deadline cut-off, each with a late remote commit, is settled only by its observed bytes).
// Every case uses its own users; their avatar rows, profile
// rows and provider keys are removed at the end of each case (the superuser fixture and the scoped credential).
package access

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/objectstore"
	"core/internal/tenancy"
)

// --- avatar fixtures ---------------------------------------------------------------------------------------------------

// avWorld is one Organization of the avatar cases; its members are created per case.
type avWorld struct {
	e   *objEnv
	org tenancy.OrganizationID
}

func (e *objEnv) avatarWorld(t *testing.T, tag string) *avWorld {
	t.Helper()
	return &avWorld{e: e, org: e.org(t, "Avatars "+tag+" "+e.canaries[0])}
}

// member signs in a new user, grants it org_member of the world's Organization and prepares its credential there; its
// avatar rows, profile row and provider keys are removed when the case ends.
func (w *avWorld) member(t *testing.T) *itUser {
	t.Helper()
	return w.memberAs(t, tenancy.OrgMember)
}

// memberAs is member with Organization role r.
func (w *avWorld) memberAs(t *testing.T, r tenancy.OrganizationRole) *itUser {
	t.Helper()
	u := w.e.newUser(t)
	w.e.grantOrg(t, w.org, u, r)
	w.e.bind(t, u, w.org)
	t.Cleanup(func() { w.e.avatarCleanup(t, u.id) })
	return u
}

// avSvc is a storage service over its own hookDB and provider spy (faultService) that counts Put calls (in all and per
// key) and Delete calls, can make every provider Put return putErr without reaching the provider, and can make every
// provider Delete fail.
type avSvc struct {
	*faultSvc
	puts       atomic.Int64
	deletes    atomic.Int64
	failDelete atomic.Bool
	putErr     atomic.Pointer[error]
	mu         sync.Mutex
	putKeys    map[string]int
}

func (w *avWorld) service(cfg ObjectConfig) *avSvc {
	a := &avSvc{faultSvc: w.e.faultService(cfg), putKeys: map[string]int{}}
	a.spy.set(func(op string, key objectstore.Key) error {
		switch op {
		case "put":
			a.puts.Add(1)
			a.mu.Lock()
			a.putKeys[key.String()]++
			a.mu.Unlock()
			if p := a.putErr.Load(); p != nil {
				return *p
			}
		case "delete":
			a.deletes.Add(1)
			if a.failDelete.Load() {
				return objectstore.ErrUnavailable
			}
		}
		return nil
	})
	return a
}

// putsFor counts the provider Put calls the spy saw for u's avatar key of row id.
func (a *avSvc) putsFor(t *testing.T, user tenancy.UserID, id string) int {
	t.Helper()
	k, err := objectstore.NewAvatarKey(user.String(), id)
	if err != nil {
		t.Fatal(err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.putKeys[k.String()]
}

// lateCommit is a provider Put whose remote commit lands after the Set returned: the spy's next Put reads the exact body
// and answers without storing anything; the real Put of the same bytes, key and binding runs only once land releases it.
type lateCommit struct {
	release   chan struct{}
	committed chan struct{}
	once      sync.Once
	used      atomic.Bool // the armed Put was called
	err       error       // the real Put's result, read after committed is closed
}

// armLateCommit arms the next provider Put of a as a late commit. Without waitCtx the Put answers at once with
// objectstore.ErrOutcomeUnknown on a live context; with waitCtx it blocks until its context ends (the transfer deadline)
// and answers with the context's error. Later Puts reach the provider unchanged. A commit still held when the case ends
// is landed before the users' keys are removed.
func (a *avSvc) armLateCommit(t *testing.T, e *objEnv, waitCtx bool) *lateCommit {
	t.Helper()
	l := &lateCommit{release: make(chan struct{}), committed: make(chan struct{})}
	a.spy.putOverride = func(ctx context.Context, key objectstore.Key, body io.Reader, size int64, ct string, b objectstore.Binding) error {
		if l.used.Swap(true) {
			return e.store.Put(ctx, key, body, size, ct, b)
		}
		data, err := io.ReadAll(body)
		if err != nil {
			l.err = err
			close(l.committed)
			return err
		}
		go func() {
			defer close(l.committed)
			select {
			case <-l.release:
			case <-time.After(5 * time.Minute):
				l.err = errors.New("the late commit was never released")
				return
			}
			l.err = e.store.Put(context.Background(), key, bytes.NewReader(data), size, ct, b)
		}()
		if waitCtx {
			<-ctx.Done()
			return ctx.Err()
		}
		return objectstore.ErrOutcomeUnknown
	}
	t.Cleanup(func() {
		if l.used.Load() {
			if err := l.wait(); err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
	})
	return l
}

// land releases the held remote commit and waits until it landed.
func (l *lateCommit) land(t *testing.T) {
	t.Helper()
	if !l.used.Load() {
		t.Fatal("the late-commit Put was never called")
	}
	if err := l.wait(); err != nil {
		t.Fatal(err)
	}
}

func (l *lateCommit) wait() error {
	l.once.Do(func() { close(l.release) })
	select {
	case <-l.committed:
	case <-time.After(30 * time.Second):
		return errors.New("the late commit did not land")
	}
	if l.err != nil {
		return fmt.Errorf("the late commit: %w", l.err)
	}
	return nil
}

// shortLease is the storage configuration of the lease cases: a one-second transfer deadline, so the lease L is
// 1 s + sourceReapGrace + avatarLeaseGrace (11 s).
func (e *objEnv) shortLease() ObjectConfig {
	c := e.ocfg
	c.TransferTimeout = time.Second
	return c
}

// as opens u's request scope at the world's origin and resolves the Organization-only context.
func (a *avSvc) as(t *testing.T, w *avWorld, u *itUser) (context.Context, *TenantContext) {
	t.Helper()
	rctx, _ := beginAs(t, a.svc, context.Background(), u, w.org)
	return rctx, resolve(t, a.svc, rctx, w.org, tenancy.GroupID{})
}

// own reads u's own profile view.
func (a *avSvc) own(t *testing.T, w *avWorld, u *itUser) AccountProfileView {
	t.Helper()
	rctx, tc := a.as(t, w, u)
	v, err := a.svc.GetOwnAccountProfile(rctx, tc)
	if err != nil {
		t.Fatalf("own profile: %v", err)
	}
	return v
}

// set uploads data as u's avatar at u's current version (read first; the hookDB modes are cleared before the call).
func (a *avSvc) set(t *testing.T, w *avWorld, u *itUser, data []byte) (AccountProfileView, error) {
	t.Helper()
	a.h.reset()
	v := a.own(t, w, u)
	rctx, tc := a.as(t, w, u)
	return a.svc.SetAccountAvatar(rctx, tc, v.Version, bytes.NewReader(data), int64(len(data)))
}

// remove withdraws u's avatar at u's current version.
func (a *avSvc) remove(t *testing.T, w *avWorld, u *itUser) (AccountProfileView, error) {
	t.Helper()
	a.h.reset()
	v := a.own(t, w, u)
	rctx, tc := a.as(t, w, u)
	return a.svc.RemoveAccountAvatar(rctx, tc, v.Version)
}

// avatarPNG is a 32x32 PNG whose pixels depend on seed (distinct seeds normalize to distinct bytes).
func avatarPNG(t *testing.T, seed uint8) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.NRGBA{R: seed, G: uint8(x * 8), B: uint8(y * 8), A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// avRow is one avatar row as the superuser fixture reads it.
type avRow struct {
	id    string
	state string
	ready bool // ready_at is set (a published avatar)
}

func (e *objEnv) avatarRows(t *testing.T, user tenancy.UserID) []avRow {
	t.Helper()
	rows, err := e.setup.Query(context.Background(), `SELECT id::text, state, ready_at IS NOT NULL FROM app.account_avatars
WHERE user_id = $1::text::uuid ORDER BY created_at, id`, user.String())
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []avRow
	for rows.Next() {
		var r avRow
		if err := rows.Scan(&r.id, &r.state, &r.ready); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// avatarStates counts u's rows per state.
func (e *objEnv) avatarStates(t *testing.T, user tenancy.UserID) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, r := range e.avatarRows(t, user) {
		out[r.state]++
	}
	return out
}

// avatarRef is u's current avatar reference ("-" for none or no profile row).
func (e *objEnv) avatarRef(t *testing.T, user tenancy.UserID) string {
	t.Helper()
	var s string
	if err := e.setup.QueryRow(context.Background(), `SELECT coalesce((SELECT avatar_object_id::text FROM app.account_profiles
WHERE user_id = $1::text::uuid), '-')`, user.String()).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func (e *objEnv) avatarPresent(t *testing.T, user tenancy.UserID, id string) bool {
	t.Helper()
	k, err := objectstore.NewAvatarKey(user.String(), id)
	if err != nil {
		t.Fatal(err)
	}
	return e.present(t, k)
}

// avatarMarked reports whether avatar row id carries the write-ahead upload marker (upload_started_at).
func (e *objEnv) avatarMarked(t *testing.T, id string) bool {
	t.Helper()
	var marked bool
	if err := e.setup.QueryRow(context.Background(), `SELECT upload_started_at IS NOT NULL FROM app.account_avatars
WHERE id = $1::text::uuid`, id).Scan(&marked); err != nil {
		t.Fatal(err)
	}
	return marked
}

// seedAvatar inserts one avatar row as the superuser fixture (no UPDATE guard on INSERT): state pending or deleting,
// created age ago; no bytes.
func (e *objEnv) seedAvatar(t *testing.T, user tenancy.UserID, state string, age time.Duration) string {
	t.Helper()
	var id string
	deletingAt := "NULL"
	if state == "deleting" {
		deletingAt = "now()"
	}
	if err := e.setup.QueryRow(context.Background(), `INSERT INTO app.account_avatars (id, user_id, state, byte_size, content_type,
  created_at, deleting_at) VALUES (gen_random_uuid(), $1::text::uuid, $2, 100, 'image/png', now() - make_interval(secs => $3), `+
		deletingAt+`) RETURNING id::text`, user.String(), state, age.Seconds()).Scan(&id); err != nil {
		t.Fatalf("seed avatar: %v", err)
	}
	return id
}

// avatarCleanup removes the user's avatar keys (scoped credential), then its profile row and avatar rows (superuser).
func (e *objEnv) avatarCleanup(t *testing.T, user tenancy.UserID) {
	t.Helper()
	ctx := context.Background()
	for _, r := range e.avatarRows(t, user) {
		if k, err := objectstore.NewAvatarKey(user.String(), r.id); err == nil {
			_ = e.store.Delete(ctx, k)
		}
	}
	if _, err := e.setup.Exec(ctx, `DELETE FROM app.account_profiles WHERE user_id = $1::text::uuid`, user.String()); err != nil {
		t.Errorf("cleanup profile: %v", err)
	}
	if _, err := e.setup.Exec(ctx, `DELETE FROM app.account_avatars WHERE user_id = $1::text::uuid`, user.String()); err != nil {
		t.Errorf("cleanup avatars: %v", err)
	}
}

// notCompensated counts the object.upload_failed not_compensated audit rows of avatar id.
func (e *objEnv) notCompensated(t *testing.T, id string) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) FROM audit.events WHERE action = 'object.upload_failed'
  AND reason = 'not_compensated' AND resource_id = $1::text::uuid`, id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// avatarRecords returns the profile.storage records written since mark (a length of the shared log).
func (e *objEnv) avatarRecords(t *testing.T, mark int) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(e.logs.String()[mark:], "\n") {
		if !strings.Contains(line, `"profile.storage"`) {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatal(err)
		}
		out = append(out, rec)
	}
	return out
}

func (e *objEnv) logMark() int { return len(e.logs.String()) }

// hasRecord reports whether recs hold a record with the stage and reason ("" matches any).
func hasRecord(recs []map[string]any, stage, reason string) bool {
	for _, r := range recs {
		if (stage == "" || r["stage"] == stage) && (reason == "" || r["reason"] == reason) {
			return true
		}
	}
	return false
}

func fieldOf(err error) string {
	var ae *Error
	if errors.As(err, &ae) {
		return ae.Field
	}
	return ""
}

// wantHold asserts a repairable HOLD: unavailable with field and a retry delay in (0, max].
func wantHold(t *testing.T, what string, err error, field string, max time.Duration) time.Duration {
	t.Helper()
	d, ok := RetryAfterOf(err)
	if CodeOf(err) != CodeUnavailable || fieldOf(err) != field || !ok || d <= 0 || d > max {
		t.Errorf("%s: %v (field %q, retry %v %t), want unavailable %s with a retry delay in (0, %v]", what, err, fieldOf(err), d, ok, field, max)
	}
	return d
}

// --- the cases ---------------------------------------------------------------------------------------------------------

func TestProfileAvatar(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()
	w := e.avatarWorld(t, "lifecycle")
	lease := e.shortLease().TransferTimeout + sourceReapGrace + avatarLeaseGrace

	// --- EP-2: A7, the fail-closed reconcile (C12) ---
	for _, c := range []struct {
		name      string
		mode      commitMode
		committed bool
	}{
		{"PRO-AV-4a lost COMMIT reply, the commit landed, the locked read fails: retained, settled as current", commitLostReply, true},
		{"PRO-AV-4b lost COMMIT reply, the commit aborted, the locked read fails: retained, swept after the lease", commitLostNoCommit, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			a := w.service(e.shortLease())
			u := w.member(t)
			data, other := avatarPNG(t, 41), avatarPNG(t, 42)
			a.hook.stage = func(op, stage string) {
				if op != opProfileAvatarSet {
					return
				}
				switch stage {
				case stageT2:
					a.h.setCommit(c.mode)
				case stageReconcile:
					// The locked read fails (its first statement after the context statement); later statements succeed.
					a.h.setCommit(commitNormal)
					a.h.setFail("?")
					a.h.on("before", "?", func() {
						a.h.setFail("")
						a.h.on("before", "?", nil)
					})
				}
			}
			mark := e.logMark()
			_, err := a.set(t, w, u, data)
			a.hook.stage = nil
			wantHold(t, "unknown publish", err, fieldCommitOutcomeUnknown, avatarHoldRetry)
			if a.deletes.Load() != 0 {
				t.Errorf("%d provider Delete calls after an unknown COMMIT with a failed locked read", a.deletes.Load())
			}
			rows := e.avatarRows(t, u.id)
			if len(rows) != 1 || !e.avatarPresent(t, u.id, rows[0].id) {
				t.Fatalf("rows %+v: want one row with its bytes retained", rows)
			}
			id := rows[0].id
			wantState, wantRef := "pending", "-"
			if c.committed {
				wantState, wantRef = "ready", id
			}
			if rows[0].state != wantState || e.avatarRef(t, u.id) != wantRef {
				t.Errorf("row %+v, reference %s; want %s, %s", rows[0], e.avatarRef(t, u.id), wantState, wantRef)
			}
			recs := e.avatarRecords(t, mark)
			if !hasRecord(recs, stageReconcile, string(reasonReconcileUnknown)) || e.notCompensated(t, id) != 1 {
				t.Errorf("fault history: records %v, not_compensated rows %d", recs, e.notCompensated(t, id))
			}

			// The next own operation settles the actual outcome.
			if c.committed {
				v := a.own(t, w, u)
				if v.Avatar == nil || v.Avatar.Version != 1 {
					t.Errorf("own view after a committed unknown publish: %+v", v)
				}
				puts := a.puts.Load()
				v, err = a.set(t, w, u, data)
				if err != nil || v.Avatar == nil || a.puts.Load() != puts || e.avatarRef(t, u.id) != id {
					t.Errorf("same bytes again: %+v %v, puts %d -> %d", v, err, puts, a.puts.Load())
				}
				return
			}
			puts := a.puts.Load()
			_, err = a.set(t, w, u, other)
			d := wantHold(t, "inside the lease", err, fieldAvatarUploadInProgress, lease)
			if a.puts.Load() != puts || len(e.avatarRows(t, u.id)) != 1 {
				t.Errorf("a held Set allocated: puts %d -> %d, rows %+v", puts, a.puts.Load(), e.avatarRows(t, u.id))
			}
			time.Sleep(d + time.Second)
			mark = e.logMark()
			v, err := a.set(t, w, u, other)
			if err != nil || v.Avatar == nil {
				t.Fatalf("after the lease: %+v %v", v, err)
			}
			rows = e.avatarRows(t, u.id)
			if len(rows) != 2 || rows[0].id != id || rows[0].state != "deleted" || e.avatarPresent(t, u.id, id) ||
				rows[1].state != "ready" || e.avatarRef(t, u.id) != rows[1].id {
				t.Errorf("after the lease: rows %+v, reference %s", rows, e.avatarRef(t, u.id))
			}
			if !hasRecord(e.avatarRecords(t, mark), stageAvatarSweep, string(reasonReconciled)) {
				t.Errorf("no sweep record: %v", e.avatarRecords(t, mark))
			}
		})
	}

	// --- EP-2: A6, reconcile-then-gate (C11) ---
	t.Run("PRO-AV-7 a persistent provider Delete fault blocks every allocation until it clears", func(t *testing.T) {
		a := w.service(e.ocfg)
		u := w.member(t)
		if _, err := a.set(t, w, u, avatarPNG(t, 70)); err != nil {
			t.Fatalf("initial avatar: %v", err)
		}
		first := e.avatarRef(t, u.id)
		a.failDelete.Store(true)
		v, err := a.set(t, w, u, avatarPNG(t, 71)) // Set #1: publishes; S5 fails, the old row stays deleting
		if err != nil || v.Avatar == nil {
			t.Fatalf("Set #1: %+v %v", v, err)
		}
		second := e.avatarRef(t, u.id)
		if st := e.avatarStates(t, u.id); st["deleting"] != 1 || st["ready"] != 1 || !e.avatarPresent(t, u.id, first) {
			t.Fatalf("after Set #1: %v", st)
		}
		for i := 2; i <= 4; i++ {
			puts, rows := a.puts.Load(), len(e.avatarRows(t, u.id))
			_, err := a.set(t, w, u, avatarPNG(t, uint8(70+i)))
			if d := wantHold(t, "Set while cleanup is pending", err, fieldAvatarCleanupPending, avatarHoldRetry); d != avatarHoldRetry {
				t.Errorf("Set #%d: retry %v, want %v", i, d, avatarHoldRetry)
			}
			if a.puts.Load() != puts || len(e.avatarRows(t, u.id)) != rows || e.avatarRef(t, u.id) != second {
				t.Errorf("Set #%d allocated: puts %d -> %d, rows %d -> %d", i, puts, a.puts.Load(), rows, len(e.avatarRows(t, u.id)))
			}
		}
		a.failDelete.Store(false)
		v, err = a.set(t, w, u, avatarPNG(t, 75)) // Set #5: reconciles, then publishes and withdraws Set #1's avatar
		if err != nil || v.Avatar == nil {
			t.Fatalf("Set #5: %+v %v", v, err)
		}
		cur := e.avatarRef(t, u.id)
		for _, r := range e.avatarRows(t, u.id) {
			switch {
			case r.id == cur:
				if r.state != "ready" || !e.avatarPresent(t, u.id, r.id) {
					t.Errorf("current %+v", r)
				}
			case r.state != "deleted" || e.avatarPresent(t, u.id, r.id):
				t.Errorf("withdrawn %+v (bytes present %t)", r, e.avatarPresent(t, u.id, r.id))
			}
		}
		if st := e.avatarStates(t, u.id); st["deleted"] != 2 || st["ready"] != 1 || len(st) != 2 {
			t.Errorf("after Set #5: %v", st)
		}
	})

	t.Run("PRO-AV-8 21 stale rows: the sweep cap holds after 20, the retry clears the 21st and publishes", func(t *testing.T) {
		a := w.service(e.ocfg)
		u := w.member(t)
		for i := 0; i <= tenancy.MaxAvatarSweep; i++ {
			e.seedAvatar(t, u.id, "deleting", time.Hour-time.Duration(i)*time.Second)
		}
		_, err := a.set(t, w, u, avatarPNG(t, 80))
		wantHold(t, "Set over the sweep cap", err, fieldAvatarCleanupPending, avatarHoldRetry)
		if st := e.avatarStates(t, u.id); st["deleted"] != 20 || st["deleting"] != 1 || len(st) != 2 || a.puts.Load() != 0 {
			t.Fatalf("after the first Set: %v, puts %d", st, a.puts.Load())
		}
		v, err := a.set(t, w, u, avatarPNG(t, 80))
		if err != nil || v.Avatar == nil {
			t.Fatalf("retry: %+v %v", v, err)
		}
		if st := e.avatarStates(t, u.id); st["deleted"] != 21 || st["ready"] != 1 || len(st) != 2 || a.puts.Load() != 1 {
			t.Errorf("after the retry: %v, puts %d", st, a.puts.Load())
		}
	})

	t.Run("PRO-AV-9 a crash after T1 leaves a pending row: held inside the lease, reconciled after it", func(t *testing.T) {
		a := w.service(e.shortLease())
		u := w.member(t)
		var crashed atomic.Bool
		a.hook.crash = func(op, stage string) bool {
			return op == opProfileAvatarSet && stage == stageProvider && !crashed.Swap(true)
		}
		if _, err := a.set(t, w, u, avatarPNG(t, 90)); CodeOf(err) != CodeUnavailable {
			t.Fatalf("crash: %v", err)
		}
		rows := e.avatarRows(t, u.id)
		if len(rows) != 1 || rows[0].state != "pending" || a.puts.Load() != 0 {
			t.Fatalf("after the crash: %+v, puts %d", rows, a.puts.Load())
		}
		_, err := a.set(t, w, u, avatarPNG(t, 91))
		d := wantHold(t, "Set inside the lease", err, fieldAvatarUploadInProgress, lease)
		if len(e.avatarRows(t, u.id)) != 1 || a.puts.Load() != 0 {
			t.Errorf("a held Set allocated: rows %+v, puts %d", e.avatarRows(t, u.id), a.puts.Load())
		}
		time.Sleep(d + time.Second)
		v, err := a.set(t, w, u, avatarPNG(t, 91))
		if err != nil || v.Avatar == nil {
			t.Fatalf("after the lease: %+v %v", v, err)
		}
		rows = e.avatarRows(t, u.id)
		if len(rows) != 2 || rows[0].state != "deleted" || rows[0].ready || rows[1].state != "ready" || e.avatarRef(t, u.id) != rows[1].id {
			t.Errorf("after the lease: %+v", rows)
		}
	})

	t.Run("PRO-AV-11 a NULL reference counts every unresolved row", func(t *testing.T) {
		t.Run("(a) the reconcile selection and the gate with no current avatar", func(t *testing.T) {
			u := w.member(t)
			if _, err := e.setup.Exec(ctx, `INSERT INTO app.account_profiles (user_id) VALUES ($1::text::uuid)`, u.id.String()); err != nil {
				t.Fatal(err)
			}
			stale := e.seedAvatar(t, u.id, "pending", time.Hour)
			deleting := e.seedAvatar(t, u.id, "deleting", time.Minute)
			// An ordinary inequality against the NULL reference counts nothing: the erratum's failure mode.
			var plain int
			if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.account_avatars a WHERE a.user_id = $1::text::uuid AND a.state <> 'deleted'
  AND a.id <> (SELECT p.avatar_object_id FROM app.account_profiles p WHERE p.user_id = $1::text::uuid)`, u.id.String()).Scan(&plain); err != nil {
				t.Fatal(err)
			}
			if plain != 0 {
				t.Fatalf("plain inequality counted %d rows; the probe is vacuous", plain)
			}
			tx, err := e.pool.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if err := tenancy.InitContext(ctx, tx, u.id, w.org, tenancy.GroupID{}); err != nil {
				t.Fatal(err)
			}
			rows, err := tenancy.StaleAvatarsTx(ctx, tx, u.id, lease)
			if err != nil {
				t.Fatal(err)
			}
			got := map[string]tenancy.ObjectState{}
			for _, r := range rows {
				got[r.ID.String()] = r.State
			}
			if len(got) != 2 || got[stale] != tenancy.ObjectPending || got[deleting] != tenancy.ObjectDeleting {
				t.Errorf("reconcile selection: %v", got)
			}
			g, err := tenancy.AvatarGateTx(ctx, tx, u.id, lease)
			if err != nil || g.Outstanding != 2 || g.Young != 0 {
				t.Errorf("gate: %+v %v", g, err)
			}
		})
		t.Run("(b) two concurrent first-ever Sets: one Put, one row, the second holds; its retry replaces", func(t *testing.T) {
			a, b := w.service(e.ocfg), w.service(e.ocfg)
			u := w.member(t)
			var nested atomic.Bool
			var second error
			// The second tab's Set runs while the first is between its allocation (T1) and its upload (P).
			a.hook.stage = func(op, stage string) {
				if op == opProfileAvatarSet && stage == stageProvider && !nested.Swap(true) {
					rctx, tc := b.as(t, w, u)
					_, second = b.svc.SetAccountAvatar(rctx, tc, 0, bytes.NewReader(avatarPNG(t, 112)), int64(len(avatarPNG(t, 112))))
				}
			}
			rctx, tc := a.as(t, w, u)
			data := avatarPNG(t, 111)
			v, err := a.svc.SetAccountAvatar(rctx, tc, 0, bytes.NewReader(data), int64(len(data)))
			a.hook.stage = nil
			if err != nil || v.Avatar == nil || !nested.Load() {
				t.Fatalf("first Set: %+v %v (second ran %t)", v, err, nested.Load())
			}
			wantHold(t, "second first-ever Set", second, fieldAvatarUploadInProgress, e.ocfg.TransferTimeout+sourceReapGrace+avatarLeaseGrace)
			if a.puts.Load()+b.puts.Load() != 1 || len(e.avatarRows(t, u.id)) != 1 {
				t.Errorf("puts %d+%d, rows %+v", a.puts.Load(), b.puts.Load(), e.avatarRows(t, u.id))
			}
			first := e.avatarRef(t, u.id)
			v, err = b.set(t, w, u, avatarPNG(t, 112)) // PRO-AV-6: the retry after the first published is a replacement
			if err != nil || v.Avatar == nil || e.avatarRef(t, u.id) == first {
				t.Errorf("retry: %+v %v", v, err)
			}
			if st := e.avatarStates(t, u.id); st["ready"] != 1 || st["deleted"] != 1 || len(st) != 2 || e.avatarPresent(t, u.id, first) {
				t.Errorf("after the retry: %v", st)
			}
		})
	})

	t.Run("PRO-AV-12 after removal: cleanup failure and an abandoned pending row block allocation until resolved", func(t *testing.T) {
		a := w.service(e.shortLease())
		u := w.member(t)
		if _, err := a.set(t, w, u, avatarPNG(t, 120)); err != nil {
			t.Fatalf("Set #1: %v", err)
		}
		x := e.avatarRef(t, u.id)
		a.failDelete.Store(true)
		v, err := a.remove(t, w, u)
		if err != nil || v.Avatar != nil || e.avatarRef(t, u.id) != "-" {
			t.Fatalf("remove: %+v %v, reference %s", v, err, e.avatarRef(t, u.id))
		}
		if st := e.avatarStates(t, u.id); st["deleting"] != 1 || len(st) != 1 || !e.avatarPresent(t, u.id, x) {
			t.Fatalf("after the failed removal cleanup: %v", st)
		}
		check := func(what, field string, max time.Duration) time.Duration {
			t.Helper()
			puts, rows := a.puts.Load(), len(e.avatarRows(t, u.id))
			_, err := a.set(t, w, u, avatarPNG(t, 121))
			d := wantHold(t, what, err, field, max)
			if a.puts.Load() != puts || len(e.avatarRows(t, u.id)) != rows || e.avatarRef(t, u.id) != "-" {
				t.Errorf("%s allocated: puts %d -> %d, rows %d -> %d", what, puts, a.puts.Load(), rows, len(e.avatarRows(t, u.id)))
			}
			return d
		}
		check("Set with the withdrawn row unresolved", fieldAvatarCleanupPending, avatarHoldRetry)
		check("repeated Set", fieldAvatarCleanupPending, avatarHoldRetry)

		// The fault clears; the next Set reconciles the withdrawn row, allocates and crashes after T1 (abandoned pending).
		a.failDelete.Store(false)
		var crashed atomic.Bool
		a.hook.crash = func(op, stage string) bool {
			return op == opProfileAvatarSet && stage == stageProvider && !crashed.Swap(true)
		}
		if _, err := a.set(t, w, u, avatarPNG(t, 122)); CodeOf(err) != CodeUnavailable {
			t.Fatalf("crash: %v", err)
		}
		if st := e.avatarStates(t, u.id); st["deleted"] != 1 || st["pending"] != 1 || len(st) != 2 || e.avatarPresent(t, u.id, x) {
			t.Fatalf("after the crash: %v", st)
		}
		d := check("Set inside the lease of the abandoned row", fieldAvatarUploadInProgress, lease)
		time.Sleep(d + time.Second)
		a.failDelete.Store(true)
		check("Set after the lease while its cleanup fails", fieldAvatarCleanupPending, avatarHoldRetry)
		if st := e.avatarStates(t, u.id); st["deleting"] != 1 || st["deleted"] != 1 || len(st) != 2 {
			t.Fatalf("after the failed sweep: %v", st)
		}
		a.failDelete.Store(false)
		v, err = a.set(t, w, u, avatarPNG(t, 123))
		if err != nil || v.Avatar == nil {
			t.Fatalf("resolved: %+v %v", v, err)
		}
		if st := e.avatarStates(t, u.id); st["deleted"] != 2 || st["ready"] != 1 || len(st) != 2 {
			t.Errorf("after the resolving Set: %v", st)
		}
	})

	t.Run("PRO-AV-10 I-CUR: the current avatar never leaves ready until its reference is cleared", func(t *testing.T) {
		a := w.service(e.ocfg)
		u := w.member(t)
		if _, err := a.set(t, w, u, avatarPNG(t, 100)); err != nil {
			t.Fatalf("avatar: %v", err)
		}
		cur := e.avatarRef(t, u.id)
		tx, err := e.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := tenancy.InitContext(ctx, tx, u.id, w.org, tenancy.GroupID{}); err != nil {
			t.Fatal(err)
		}
		withdraw := `UPDATE app.account_avatars SET state = 'deleting' WHERE id = $1::text::uuid`
		if _, err := tx.Exec(ctx, `SAVEPOINT icur`); err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, withdraw, cur)
		var pg *pgconn.PgError
		if !errors.As(err, &pg) || pg.ConstraintName != "account_avatars_current" {
			t.Fatalf("withdrawal of the current avatar: %v, want account_avatars_current", err)
		}
		if _, err := tx.Exec(ctx, `ROLLBACK TO SAVEPOINT icur`); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE app.account_profiles SET avatar_object_id = NULL, avatar_version = avatar_version + 1,
  version = version + 1 WHERE user_id = $1::text::uuid`, u.id.String()); err != nil {
			t.Fatalf("clear the reference: %v", err)
		}
		if tag, err := tx.Exec(ctx, withdraw, cur); err != nil || tag.RowsAffected() != 1 {
			t.Errorf("withdrawal after the reference moved: %v %v", tag, err)
		}
		// Rolled back: the case leaves the avatar current.
	})

	// --- the fence of Plan r3.1 §3.8 (review P1): the write-ahead upload marker ---
	t.Run("PRO-AV-13 a writer paused after its allocation beyond the lease gets no Put and leaves no orphan bytes", func(t *testing.T) {
		a := w.service(e.shortLease())
		u := w.member(t)
		cycle := func(n int) {
			t.Helper()
			first, second := avatarPNG(t, uint8(130+2*n)), avatarPNG(t, uint8(131+2*n))
			var armed atomic.Bool
			paused, resume, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
			released := false
			release := func() {
				if !released {
					released = true
					close(resume)
				}
			}
			defer release()
			// The first Set stops after its allocation COMMIT, before its write-ahead marker (the provider-stage hook).
			a.hook.stage = func(op, stage string) {
				if op == opProfileAvatarSet && stage == stageProvider && armed.Swap(false) {
					close(paused)
					select {
					case <-resume:
					case <-time.After(time.Minute):
					}
				}
			}
			armed.Store(true)
			a.h.reset()
			v0 := a.own(t, w, u)
			rctx, tc := a.as(t, w, u)
			var errA error
			go func() {
				defer close(done)
				_, errA = a.svc.SetAccountAvatar(rctx, tc, v0.Version, bytes.NewReader(first), int64(len(first)))
			}()
			select {
			case <-paused:
			case <-done:
				t.Fatalf("cycle %d: the first Set returned before its pause: %v", n, errA)
			case <-time.After(30 * time.Second):
				t.Fatalf("cycle %d: the first Set never reached its pause", n)
			}
			fenced := ""
			for _, r := range e.avatarRows(t, u.id) {
				if r.state == "pending" {
					if fenced != "" {
						t.Fatalf("cycle %d: two pending rows while the first Set is paused: %+v", n, e.avatarRows(t, u.id))
					}
					fenced = r.id
				}
			}
			if fenced == "" {
				t.Fatalf("cycle %d: no pending row while the first Set is paused: %+v", n, e.avatarRows(t, u.id))
			}
			if e.avatarMarked(t, fenced) {
				t.Fatalf("cycle %d: the paused row already carries the upload marker", n)
			}

			// Past the lease of the paused allocation: a second Set sweeps the aged unmarked row (no Put can be in flight for
			// it, so no object exists), tombstones it, then publishes.
			time.Sleep(lease + time.Second)
			mark := e.logMark()
			v, err := a.set(t, w, u, second)
			if err != nil || v.Avatar == nil {
				t.Fatalf("cycle %d: the second Set: %+v %v", n, v, err)
			}
			if !hasRecord(e.avatarRecords(t, mark), stageAvatarSweep, string(reasonReconciled)) {
				t.Errorf("cycle %d: no sweep record: %v", n, e.avatarRecords(t, mark))
			}
			for _, r := range e.avatarRows(t, u.id) {
				if r.id == fenced && r.state != "deleted" {
					t.Fatalf("cycle %d: the second Set did not tombstone the aged unmarked row: %+v", n, r)
				}
			}
			cur, version := e.avatarRef(t, u.id), v.Version

			// Any provider Delete would now fail; the first writer resumes and its marker update finds no pending row.
			a.failDelete.Store(true)
			puts, deletes := a.puts.Load(), a.deletes.Load()
			mark = e.logMark()
			release()
			select {
			case <-done:
			case <-time.After(time.Minute):
				t.Fatalf("cycle %d: the resumed first Set did not return", n)
			}
			a.hook.stage = nil
			a.failDelete.Store(false)
			if CodeOf(errA) != CodeUnavailable || fieldOf(errA) != "" {
				t.Errorf("cycle %d: the fenced Set: %v, want unavailable", n, errA)
			}
			if _, ok := RetryAfterOf(errA); ok {
				t.Errorf("cycle %d: the fenced Set is not a HOLD: %v", n, errA)
			}
			if a.puts.Load() != puts || a.putsFor(t, u.id, fenced) != 0 {
				t.Errorf("cycle %d: the fenced Set made %d provider Put calls (%d for its key)", n, a.puts.Load()-puts, a.putsFor(t, u.id, fenced))
			}
			if a.deletes.Load() != deletes {
				t.Errorf("cycle %d: the fenced Set made %d provider Delete calls", n, a.deletes.Load()-deletes)
			}
			if !hasRecord(e.avatarRecords(t, mark), "", string(reasonObjectNotReady)) {
				t.Errorf("cycle %d: no object_not_ready record for the fenced Set: %v", n, e.avatarRecords(t, mark))
			}
			if pv := a.own(t, w, u); pv.Version != version || e.avatarRef(t, u.id) != cur {
				t.Errorf("cycle %d: the fenced Set changed the profile: version %d -> %d, reference changed %t", n, version, pv.Version,
					e.avatarRef(t, u.id) != cur)
			}
			stored := 0
			for _, r := range e.avatarRows(t, u.id) {
				present := e.avatarPresent(t, u.id, r.id)
				if present {
					stored++
				}
				switch {
				case r.id == fenced:
					if r.state != "deleted" || r.ready || present {
						t.Errorf("cycle %d: the fenced row %+v (bytes present %t)", n, r, present)
					}
				case r.id == cur:
					if r.state != "ready" || !present {
						t.Errorf("cycle %d: the current row %+v (bytes present %t)", n, r, present)
					}
				case r.state != "deleted" || present:
					t.Errorf("cycle %d: a withdrawn row %+v (bytes present %t)", n, r, present)
				}
			}
			if stored != 1 {
				t.Errorf("cycle %d: %d objects under the user's avatar keys, want exactly the current one", n, stored)
			}
		}
		cycle(0)
		cycle(1)
		// The gate is not blocked: an immediate Set (no lease wait) publishes.
		v, err := a.set(t, w, u, avatarPNG(t, 139))
		if err != nil || v.Avatar == nil {
			t.Fatalf("Set after the fenced writers: %+v %v", v, err)
		}
		if st := e.avatarStates(t, u.id); st["deleted"] != 4 || st["ready"] != 1 || len(st) != 2 {
			t.Errorf("after the fenced writers: %v", st)
		}
	})

	t.Run("PRO-AV-14 an issued Put with an unproven outcome is never settled by absence, only by its observed bytes", func(t *testing.T) {
		// unproven asserts what an issued Put with an unproven outcome leaves: unavailable compensation_failed (not a HOLD),
		// one Put, no Delete and no tombstone, one pending row carrying the upload marker, no object yet and no reference.
		unproven := func(t *testing.T, a *avSvc, u *itUser, err error, mark int) string {
			t.Helper()
			if CodeOf(err) != CodeUnavailable || fieldOf(err) != "" {
				t.Fatalf("unproven Put: %v, want unavailable", err)
			}
			if _, ok := RetryAfterOf(err); ok {
				t.Errorf("an unproven Put is not a HOLD: %v", err)
			}
			if recs := e.avatarRecords(t, mark); !hasRecord(recs, stageCompensate, string(reasonCompensationFailed)) ||
				hasRecord(recs, "", string(reasonCompensated)) {
				t.Errorf("records of the unproven Put: %v", recs)
			}
			rows := e.avatarRows(t, u.id)
			if len(rows) != 1 || rows[0].state != "pending" || rows[0].ready || a.deletes.Load() != 0 || a.puts.Load() != 1 {
				t.Fatalf("after the unproven Put: rows %+v, deletes %d, puts %d", rows, a.deletes.Load(), a.puts.Load())
			}
			id := rows[0].id
			if !e.avatarMarked(t, id) || e.avatarPresent(t, u.id, id) || e.avatarRef(t, u.id) != "-" {
				t.Errorf("the unproven row: marked %t, bytes present %t, reference %s", e.avatarMarked(t, id), e.avatarPresent(t, u.id, id),
					e.avatarRef(t, u.id))
			}
			return id
		}
		// heldYoung: inside the lease a Set holds (avatar_upload_in_progress), allocates nothing and deletes nothing.
		heldYoung := func(t *testing.T, a *avSvc, u *itUser, seed uint8) time.Duration {
			t.Helper()
			_, err := a.set(t, w, u, avatarPNG(t, seed))
			d := wantHold(t, "Set inside the lease of the unproven row", err, fieldAvatarUploadInProgress, lease)
			if a.puts.Load() != 1 || a.deletes.Load() != 0 || len(e.avatarRows(t, u.id)) != 1 {
				t.Errorf("a held Set acted: puts %d, deletes %d, rows %+v", a.puts.Load(), a.deletes.Load(), e.avatarRows(t, u.id))
			}
			return d
		}
		// heldAged: after the lease a Set finds no object for the marked row. Absence proves nothing (the request may still
		// commit), so the row is moved to deleting and held (avatar_cleanup_pending), never tombstoned; the Set allocates
		// nothing and deletes nothing, and the row stays the user's one unresolved row.
		heldAged := func(t *testing.T, a *avSvc, u *itUser, id string, seed uint8) {
			t.Helper()
			mark := e.logMark()
			_, err := a.set(t, w, u, avatarPNG(t, seed))
			wantHold(t, "Set after the lease with no observed bytes", err, fieldAvatarCleanupPending, avatarHoldRetry)
			rows := e.avatarRows(t, u.id)
			if len(rows) != 1 || rows[0].id != id || rows[0].state != "deleting" || a.puts.Load() != 1 || a.deletes.Load() != 0 {
				t.Errorf("an absent object settled the unproven row: rows %+v, puts %d, deletes %d", rows, a.puts.Load(), a.deletes.Load())
			}
			if !hasRecord(e.avatarRecords(t, mark), stageAvatarSweep, string(reasonDeletionPending)) {
				t.Errorf("no deletion_pending sweep record: %v", e.avatarRecords(t, mark))
			}
		}
		// recovered: the next Set observes the late bytes, deletes them, confirms their absence, tombstones the row and
		// publishes its own avatar.
		recovered := func(t *testing.T, a *avSvc, u *itUser, id string, seed uint8) {
			t.Helper()
			mark, deletes := e.logMark(), a.deletes.Load()
			v, err := a.set(t, w, u, avatarPNG(t, seed))
			if err != nil || v.Avatar == nil {
				t.Fatalf("the Set after the late bytes landed: %+v %v", v, err)
			}
			rows := e.avatarRows(t, u.id)
			if len(rows) != 2 || rows[0].id != id || rows[0].state != "deleted" || rows[0].ready || e.avatarPresent(t, u.id, id) ||
				rows[1].state != "ready" || !e.avatarPresent(t, u.id, rows[1].id) || e.avatarRef(t, u.id) != rows[1].id {
				t.Errorf("after the observed settlement: rows %+v, reference %s", rows, e.avatarRef(t, u.id))
			}
			if a.deletes.Load() != deletes+1 || a.puts.Load() != 2 || !hasRecord(e.avatarRecords(t, mark), stageAvatarSweep, string(reasonReconciled)) {
				t.Errorf("no observed settlement: deletes %d -> %d, puts %d, records %v", deletes, a.deletes.Load(), a.puts.Load(),
					e.avatarRecords(t, mark))
			}
		}

		t.Run("(a) a live-context unknown outcome whose remote commit lands late, with a Delete fault on the observed bytes", func(t *testing.T) {
			a := w.service(e.shortLease())
			u := w.member(t)
			late := a.armLateCommit(t, e, false)
			mark := e.logMark()
			_, err := a.set(t, w, u, avatarPNG(t, 140))
			id := unproven(t, a, u, err, mark)
			time.Sleep(heldYoung(t, a, u, 141) + time.Second)
			heldAged(t, a, u, id, 141)
			heldAged(t, a, u, id, 141) // repeated: still held, still no tombstone, no further allocation

			// The remote commit lands under the held row while every provider Delete fails: the bytes are observed but stay
			// accounted for under the deleting row, and allocation stays held.
			late.land(t)
			if !e.avatarPresent(t, u.id, id) {
				t.Fatal("the late commit stored nothing under the held row")
			}
			a.failDelete.Store(true)
			mark = e.logMark()
			_, err = a.set(t, w, u, avatarPNG(t, 141))
			wantHold(t, "Set while the observed bytes cannot be deleted", err, fieldAvatarCleanupPending, avatarHoldRetry)
			rows := e.avatarRows(t, u.id)
			if len(rows) != 1 || rows[0].id != id || rows[0].state != "deleting" || !e.avatarPresent(t, u.id, id) || a.puts.Load() != 1 ||
				a.deletes.Load() != 1 {
				t.Errorf("observed bytes with a failing Delete: rows %+v, bytes present %t, puts %d, deletes %d", rows,
					e.avatarPresent(t, u.id, id), a.puts.Load(), a.deletes.Load())
			}
			if !hasRecord(e.avatarRecords(t, mark), stageAvatarSweep, string(reasonDeletionPending)) {
				t.Errorf("no deletion_pending sweep record: %v", e.avatarRecords(t, mark))
			}
			a.failDelete.Store(false)
			recovered(t, a, u, id, 142)
		})

		t.Run("(b) a Put cut off by its transfer deadline whose remote commit lands late", func(t *testing.T) {
			a := w.service(e.shortLease())
			u := w.member(t)
			late := a.armLateCommit(t, e, true)
			mark := e.logMark()
			_, err := a.set(t, w, u, avatarPNG(t, 145))
			id := unproven(t, a, u, err, mark)
			time.Sleep(heldYoung(t, a, u, 146) + time.Second)
			heldAged(t, a, u, id, 146)
			late.land(t)
			if !e.avatarPresent(t, u.id, id) {
				t.Fatal("the late commit stored nothing under the held row")
			}
			recovered(t, a, u, id, 146)
		})
	})

	// --- the rest of lane L2: lifecycle, refusals, fault seams, delivery ---
	t.Run("PRO-6 set, identical bytes, replace and remove with the provider state after each step", func(t *testing.T) {
		a := w.service(e.ocfg)
		u := w.member(t)
		mark := e.logMark()
		v, err := a.set(t, w, u, avatarPNG(t, 60))
		if err != nil || v.Avatar == nil || v.Avatar.Version != 1 || v.Avatar.ContentType != "image/png" {
			t.Fatalf("set: %+v %v", v, err)
		}
		first := e.avatarRef(t, u.id)
		if !e.avatarPresent(t, u.id, first) || e.avatarStates(t, u.id)["ready"] != 1 {
			t.Fatalf("after set: %v", e.avatarStates(t, u.id))
		}
		for _, r := range e.avatarRecords(t, mark) {
			for _, k := range []string{"key", "storage_key", "url", "size", "byte_size", "name", "sha256", "organization_id"} {
				if _, ok := r[k]; ok {
					t.Errorf("profile.storage record carries %s: %v", k, r)
				}
			}
		}
		puts := a.puts.Load()
		if v2, err := a.set(t, w, u, avatarPNG(t, 60)); err != nil || v2.Version != v.Version || a.puts.Load() != puts {
			t.Errorf("identical bytes: %+v %v, puts %d -> %d", v2, err, puts, a.puts.Load())
		}
		v, err = a.set(t, w, u, avatarPNG(t, 61))
		if err != nil || v.Avatar == nil || v.Avatar.Version != 2 {
			t.Fatalf("replace: %+v %v", v, err)
		}
		second := e.avatarRef(t, u.id)
		if second == first || !e.avatarPresent(t, u.id, second) || e.avatarPresent(t, u.id, first) {
			t.Errorf("after replace: first present %t, second present %t", e.avatarPresent(t, u.id, first), e.avatarPresent(t, u.id, second))
		}
		if rows := e.avatarRows(t, u.id); len(rows) != 2 || rows[0].state != "deleted" || !rows[0].ready || rows[1].state != "ready" {
			t.Errorf("after replace: %+v (the withdrawn row keeps ready_at)", rows)
		}
		v, err = a.remove(t, w, u)
		if err != nil || v.Avatar != nil || e.avatarRef(t, u.id) != "-" || e.avatarPresent(t, u.id, second) {
			t.Fatalf("remove: %+v %v", v, err)
		}
		if st := e.avatarStates(t, u.id); st["deleted"] != 2 || len(st) != 1 {
			t.Errorf("after remove: %v", st)
		}
		version := v.Version
		if v, err = a.remove(t, w, u); err != nil || v.Version != version {
			t.Errorf("remove twice: %+v %v", v, err)
		}
	})

	t.Run("PRO-7 normalization refusals store nothing and call no provider", func(t *testing.T) {
		a := w.service(e.ocfg)
		u := w.member(t)
		good := avatarPNG(t, 70)
		tiny := func() []byte {
			img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
			var buf bytes.Buffer
			if err := png.Encode(&buf, img); err != nil {
				t.Fatal(err)
			}
			return buf.Bytes()
		}()
		for _, c := range []struct {
			name  string
			body  []byte
			size  int64
			field string
		}{
			{"SVG", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"/>`), -1, "avatar_type"},
			{"GIF", append([]byte("GIF89a"), bytes.Repeat([]byte{0}, 64)...), -1, "avatar_type"},
			{"appended bytes", append(append([]byte{}, good...), "trailing"...), -1, "avatar_malformed"},
			{"8x8", tiny, -1, "avatar_dimensions"},
			{"over 512 KiB", bytes.Repeat([]byte{1}, 524289), -1, "avatar_too_large"},
			{"length differs from the declared size", good, int64(len(good)) + 1, fieldAvatarSize},
		} {
			size := c.size
			if size < 0 {
				size = int64(len(c.body))
			}
			rctx, tc := a.as(t, w, u)
			_, err := a.svc.SetAccountAvatar(rctx, tc, 0, bytes.NewReader(c.body), size)
			if CodeOf(err) != CodeInvalidInput || fieldOf(err) != c.field {
				t.Errorf("%s: %v (field %q), want invalid_input %s", c.name, err, fieldOf(err), c.field)
			}
		}
		var profiles int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.account_profiles WHERE user_id = $1::text::uuid`, u.id.String()).Scan(&profiles); err != nil {
			t.Fatal(err)
		}
		if len(e.avatarRows(t, u.id)) != 0 || profiles != 0 || a.spy.calls.Load() != 0 {
			t.Errorf("a refusal stored something: rows %d, profiles %d, provider calls %d", len(e.avatarRows(t, u.id)), profiles, a.spy.calls.Load())
		}
	})

	t.Run("PRO-AV-1 a T1 statement failure allocates nothing and calls no provider Put", func(t *testing.T) {
		a := w.service(e.ocfg)
		u := w.member(t)
		a.hook.stage = func(op, stage string) {
			if op == opProfileAvatarSet && stage == stageT1 {
				a.h.setFail("?") // the empty-row insert, the first unclassified statement of T1
			}
		}
		_, err := a.set(t, w, u, avatarPNG(t, 10))
		a.hook.stage = nil
		if CodeOf(err) != CodeUnavailable || fieldOf(err) != "" || len(e.avatarRows(t, u.id)) != 0 || a.puts.Load() != 0 {
			t.Errorf("T1 failure: %v, rows %+v, puts %d", err, e.avatarRows(t, u.id), a.puts.Load())
		}
	})

	t.Run("PRO-AV-2 a refused provider Put is compensated at once, an unproven one is held, a failed compensation blocks allocation", func(t *testing.T) {
		a := w.service(e.ocfg)
		u := w.member(t)
		// A definitive refusal (denied): no object can exist for it, so the row is compensated at once (deleted, absence
		// confirmed, tombstoned) and the very next Set publishes.
		a.putErr.Store(&objectstore.ErrDenied)
		mark := e.logMark()
		_, err := a.set(t, w, u, avatarPNG(t, 20))
		a.putErr.Store(nil)
		rows := e.avatarRows(t, u.id)
		if CodeOf(err) != CodeUnavailable || fieldOf(err) != "" || len(rows) != 1 || rows[0].state != "deleted" || rows[0].ready ||
			e.avatarPresent(t, u.id, rows[0].id) || a.puts.Load() != 1 {
			t.Fatalf("compensated Put refusal: %v, rows %+v, puts %d", err, rows, a.puts.Load())
		}
		if _, ok := RetryAfterOf(err); ok {
			t.Errorf("a compensated failure is not a HOLD: %v", err)
		}
		if !hasRecord(e.avatarRecords(t, mark), stageCompensate, string(reasonCompensated)) {
			t.Errorf("no compensated step record: %v", e.avatarRecords(t, mark))
		}
		v, err := a.set(t, w, u, avatarPNG(t, 21))
		if err != nil || v.Avatar == nil {
			t.Fatalf("the Set right after the compensated refusal: %+v %v", v, err)
		}
		cur := e.avatarRef(t, u.id)

		// A refusal (invalid) whose compensation fails: the pending row stays and blocks allocation inside its lease.
		a.putErr.Store(&objectstore.ErrInvalid)
		a.failDelete.Store(true)
		mark, deletes := e.logMark(), a.deletes.Load()
		_, err = a.set(t, w, u, avatarPNG(t, 22))
		a.putErr.Store(nil)
		a.failDelete.Store(false)
		if CodeOf(err) != CodeUnavailable {
			t.Fatalf("Put refusal with a failed compensation: %v", err)
		}
		if st := e.avatarStates(t, u.id); st["deleting"] != 1 || st["deleted"] != 1 || st["ready"] != 1 || len(st) != 3 ||
			a.deletes.Load() == deletes || e.avatarRef(t, u.id) != cur {
			t.Fatalf("after the failed compensation: %v, deletes %d -> %d", st, deletes, a.deletes.Load())
		}
		for _, r := range e.avatarRows(t, u.id) {
			if r.state == "deleting" && e.avatarMarked(t, r.id) {
				t.Errorf("a refused Put's row kept its upload marker: %+v", r)
			}
		}
		if !hasRecord(e.avatarRecords(t, mark), stageCompensate, string(reasonCompensationFailed)) {
			t.Errorf("no compensation_failed step record: %v", e.avatarRecords(t, mark))
		}
		// The withdrawn row carries no marker (no object can exist for the refused Put), so it is an ordinary withdrawn
		// row: the next Set deletes and confirms its absence, tombstones it, and publishes, with no lease wait.
		v, err = a.set(t, w, u, avatarPNG(t, 23))
		if err != nil || v.Avatar == nil {
			t.Fatalf("the Set after the refused Put's failed compensation: %+v %v", v, err)
		}
		if st := e.avatarStates(t, u.id); st["deleting"] != 0 || st["pending"] != 0 || st["ready"] != 1 {
			t.Errorf("after the repairing Set: %v", st)
		}

		// A transport failure (unavailable) of a live call is no refusal: its outcome is unproven, so nothing is deleted and
		// the marked row stays pending (PRO-AV-14 proves its settlement by observed bytes only).
		u2 := w.member(t)
		a.putErr.Store(&objectstore.ErrUnavailable)
		mark, deletes = e.logMark(), a.deletes.Load()
		_, err = a.set(t, w, u2, avatarPNG(t, 24))
		a.putErr.Store(nil)
		rows = e.avatarRows(t, u2.id)
		if CodeOf(err) != CodeUnavailable || fieldOf(err) != "" || len(rows) != 1 || rows[0].state != "pending" ||
			!e.avatarMarked(t, rows[0].id) || e.avatarPresent(t, u2.id, rows[0].id) || a.deletes.Load() != deletes {
			t.Errorf("unproven Put: %v, rows %+v, deletes %d -> %d", err, rows, deletes, a.deletes.Load())
		}
		if _, ok := RetryAfterOf(err); ok {
			t.Errorf("an unproven Put is not a HOLD: %v", err)
		}
		if recs := e.avatarRecords(t, mark); !hasRecord(recs, stageCompensate, string(reasonCompensationFailed)) ||
			hasRecord(recs, "", string(reasonCompensated)) {
			t.Errorf("records of the unproven Put: %v", recs)
		}
	})

	t.Run("PRO-AV-3 a verification binding mismatch is internal, compensated and never published", func(t *testing.T) {
		a := w.service(e.ocfg)
		u, other := w.member(t), w.member(t)
		a.spy.putOverride = func(ctx context.Context, key objectstore.Key, body io.Reader, size int64, ct string, b objectstore.Binding) error {
			return e.store.Put(ctx, key, body, size, ct, objectstore.Binding{Object: b.Object, User: other.id.String()})
		}
		_, err := a.set(t, w, u, avatarPNG(t, 30))
		rows := e.avatarRows(t, u.id)
		if CodeOf(err) != CodeInternal || len(rows) != 1 || rows[0].state != "deleted" || e.avatarPresent(t, u.id, rows[0].id) ||
			e.avatarRef(t, u.id) != "-" {
			t.Errorf("binding mismatch: %v, rows %+v", err, rows)
		}
	})

	t.Run("PRO-AV-5 a stale version at T2 is version_conflict with the new bytes compensated", func(t *testing.T) {
		a, b := w.service(e.ocfg), w.service(e.ocfg)
		u := w.member(t)
		var bumped atomic.Bool
		a.hook.stage = func(op, stage string) {
			if op == opProfileAvatarSet && stage == stageT2 && !bumped.Swap(true) {
				rctx, tc := b.as(t, w, u)
				cur, err := b.svc.GetOwnAccountProfile(rctx, tc)
				if err == nil {
					_, err = b.svc.UpdateAccountProfileName(rctx, tc, cur.Version, ptrString("Concurrent Name"))
				}
				if err != nil {
					t.Errorf("concurrent name edit: %v", err)
				}
			}
		}
		_, err := a.set(t, w, u, avatarPNG(t, 50))
		a.hook.stage = nil
		rows := e.avatarRows(t, u.id)
		if CodeOf(err) != CodeConflict || fieldOf(err) != fieldVersionConflict || len(rows) != 1 || rows[0].state != "deleted" ||
			e.avatarPresent(t, u.id, rows[0].id) || e.avatarRef(t, u.id) != "-" {
			t.Errorf("stale T2: %v, rows %+v", err, rows)
		}
	})

	t.Run("PRO-AV-G delivery decides now on every read, verifies before any byte and never discloses", func(t *testing.T) {
		a := w.service(e.ocfg)
		creator := w.member(t)
		owner, peer, outsider, admin := w.member(t), w.member(t), w.member(t), w.memberAs(t, tenancy.OrgAdmin)
		g := e.group(t, w.org, creator, "Avatars delivery", "avatars-delivery").ID
		e.grantGroup(t, w.org, g, owner, tenancy.GroupMember)
		e.grantGroup(t, w.org, g, peer, tenancy.GroupMember)
		data := avatarPNG(t, 150)
		v, err := a.set(t, w, owner, data)
		if err != nil || v.Avatar == nil {
			t.Fatalf("set: %+v %v", v, err)
		}
		ver := v.Avatar.Version
		id := e.avatarRef(t, owner.id)
		var want []byte
		for _, viewer := range []*itUser{owner, peer, admin} {
			ref, got, err := a.get(t, w, viewer, owner.id, ver)
			if err != nil || ref.Version != ver || ref.ContentType != "image/png" || len(got) == 0 {
				t.Errorf("%s reads the avatar: %+v %d bytes %v", viewer.sub, ref, len(got), err)
			}
			if want == nil {
				want = got
			} else if !bytes.Equal(got, want) {
				t.Error("viewers read different bytes")
			}
		}
		gets := func() int64 { return a.spy.calls.Load() }
		before := gets()
		for _, c := range []struct {
			name    string
			viewer  *itUser
			subject tenancy.UserID
			version int32
		}{
			{"an unrelated member of the Organization", outsider, owner.id, ver},
			{"an unknown subject", peer, unknownUserID(t), ver},
			{"a stale version", peer, owner.id, ver + 1},
			{"version zero", peer, owner.id, 0},
		} {
			if _, got, err := a.get(t, w, c.viewer, c.subject, c.version); CodeOf(err) != CodeNotFound || got != nil {
				t.Errorf("%s: %v (%d bytes), want not_found", c.name, err, len(got))
			}
		}
		if gets() != before {
			t.Errorf("a refused read called the provider (%d calls)", gets()-before)
		}

		// Integrity: same-length foreign bytes under the key are internal, never returned; missing bytes are not_found.
		key, err := objectstore.NewAvatarKey(owner.id.String(), id)
		if err != nil {
			t.Fatal(err)
		}
		corrupt := append([]byte{}, want...)
		corrupt[len(corrupt)-1] ^= 0xff
		if err := e.store.Put(ctx, key, bytes.NewReader(corrupt), int64(len(corrupt)), "image/png",
			objectstore.Binding{Object: id, User: owner.id.String()}); err != nil {
			t.Fatal(err)
		}
		if _, got, err := a.get(t, w, peer, owner.id, ver); CodeOf(err) != CodeInternal || got != nil {
			t.Errorf("corrupt bytes: %v (%d bytes), want internal", err, len(got))
		}
		if err := e.store.Delete(ctx, key); err != nil {
			t.Fatal(err)
		}
		mark := e.logMark()
		if _, got, err := a.get(t, w, peer, owner.id, ver); CodeOf(err) != CodeNotFound || got != nil {
			t.Errorf("missing bytes: %v, want not_found", err)
		}
		if !hasRecord(e.avatarRecords(t, mark), "", string(reasonObjectBytesMissing)) {
			t.Errorf("no object_bytes_missing record: %v", e.avatarRecords(t, mark))
		}
		if err := e.store.Put(ctx, key, bytes.NewReader(want), int64(len(want)), "image/png",
			objectstore.Binding{Object: id, User: owner.id.String()}); err != nil {
			t.Fatal(err)
		}

		// Revocation: once the shared Group binding ends, the peer's next read is not_found; replacement retires the old v.
		if err := e.model.RemoveGroupMembership(ctx, w.org, g, peer.id); err != nil {
			t.Fatal(err)
		}
		if _, _, err := a.get(t, w, peer, owner.id, ver); CodeOf(err) != CodeNotFound {
			t.Errorf("after revocation: %v, want not_found", err)
		}
		v, err = a.set(t, w, owner, avatarPNG(t, 151))
		if err != nil || v.Avatar == nil || v.Avatar.Version != ver+1 {
			t.Fatalf("replace: %+v %v", v, err)
		}
		if _, _, err := a.get(t, w, admin, owner.id, ver); CodeOf(err) != CodeNotFound {
			t.Errorf("the replaced version: %v, want not_found", err)
		}
		if _, got, err := a.get(t, w, admin, owner.id, ver+1); err != nil || len(got) == 0 {
			t.Errorf("the current version: %v", err)
		}
	})
}

// get reads subject's avatar at version as viewer u.
func (a *avSvc) get(t *testing.T, w *avWorld, u *itUser, subject tenancy.UserID, version int32) (AccountAvatar, []byte, error) {
	t.Helper()
	rctx, tc := a.as(t, w, u)
	return a.svc.GetAccountAvatar(rctx, tc, subject, version)
}

func ptrString(s string) *string { return &s }

// unknownUserID is a random version-4 UUID that names no user.
func unknownUserID(t *testing.T) tenancy.UserID {
	t.Helper()
	id, err := newObjectID()
	if err != nil {
		t.Fatal(err)
	}
	return tenancy.UserID(id)
}
