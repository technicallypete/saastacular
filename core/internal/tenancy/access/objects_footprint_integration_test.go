//go:build integration

// E-OBJ-1..4 (G009 Plan r1 §10, PG-003 r0 assessment; recorded in B-OBJ-IT): TestObjectsFootprint collects bounded
// footprint evidence on a populated fixture (about 200 objects in two Organizations): EXPLAIN summaries and policy
// reader calls of every storage statement (E-OBJ-1, E-OBJ-3), the locked-row inventory of the writes (E-OBJ-2), and the
// separation of transactions and transfers (E-OBJ-4). Statements are replayed in rolled-back runtime transactions,
// with reader calls counted as differences of pg_stat_get_xact_function_calls inside the replaying transaction (the
// accepted G008 measurement). Prompt-escalation conditions print E-OBJ-TRIGGER lines for review: growth with bindings,
// a statement above its hypothesized reader bound, or a runtime backend idle in transaction during a transfer. A PASS
// means the collection succeeded and the asserted safety facts held; it does not accept a cost.
package access

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/objectstore"
	"core/internal/tenancy"
)

// objReaderBounds are the Plan r1 §10 hypotheses per statement kind: rls_org_role and rls_group_role calls.
var objReaderBounds = map[string][2]int64{"select": {2, 1}, "insert": {2, 1}, "lock": {4, 3}, "update": {6, 5}}

func objStatementKind(sql string) string {
	switch {
	case strings.HasPrefix(sql, "INSERT INTO app.stored_objects"):
		return "insert"
	case strings.HasPrefix(sql, "UPDATE app.stored_objects"):
		return "update"
	case strings.Contains(sql, "FROM app.stored_objects") && strings.HasSuffix(sql, "FOR UPDATE"):
		return "lock"
	case strings.Contains(sql, "FROM app.stored_objects"):
		return "select"
	}
	return ""
}

func objTrigger(fields map[string]any) {
	fields["evidence"] = "E-OBJ-TRIGGER"
	line, _ := json.Marshal(fields)
	fmt.Println("E-OBJ-TRIGGER " + string(line))
}

// objReplay replays one storage transaction's captured statements (context statement first) under EXPLAIN in a
// rolled-back runtime transaction and returns the per-statement reader calls of the stored-object statements.
func (e *objEnv) objReplay(t *testing.T, pool *pgxpool.Pool, subject, op string, stmts []capturedStmt) map[string][2]int64 {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, stmts[0].sql, stmts[0].args...); err != nil {
		t.Fatalf("replay context: %v", err)
	}
	calls := func() (int64, int64) {
		var org, grp int64
		if err := tx.QueryRow(ctx, `SELECT coalesce(pg_stat_get_xact_function_calls('app.rls_org_role()'::regprocedure), 0),
  coalesce(pg_stat_get_xact_function_calls('app.rls_group_role(uuid)'::regprocedure), 0)`).Scan(&org, &grp); err != nil {
			t.Fatal(err)
		}
		return org, grp
	}
	out := map[string][2]int64{}
	for i, st := range stmts[1:] {
		args := st.args
		kind := objStatementKind(st.sql)
		if kind == "insert" { // the row exists for real: replay the insert's plan under a fresh identity
			var id [16]byte
			_, _ = rand.Read(id[:])
			id[6], id[8] = (id[6]&0x0f)|0x40, (id[8]&0x3f)|0x80
			args = append([]any{}, args...)
			args[1] = pgtype.UUID{Bytes: id, Valid: true}
		}
		o0, g0 := calls()
		var raw []byte
		if err := tx.QueryRow(ctx, "EXPLAIN (ANALYZE, VERBOSE, BUFFERS, FORMAT JSON) "+st.sql, args...).Scan(&raw); err != nil {
			t.Fatalf("explain %s %s: %v", op, st.step, err)
		}
		o1, g1 := calls()
		var doc []map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil || len(doc) != 1 {
			t.Fatalf("explain: %v", err)
		}
		plan, _ := doc[0]["Plan"].(map[string]any)
		line, _ := json.Marshal(map[string]any{"evidence": "E-OBJ-1", "subject": subject, "operation": op, "index": i + 1, "step": st.step,
			"kind": kind, "plan": summarize(plan), "scanned_rows": scannedRows(plan), "rls_org_role": o1 - o0, "rls_group_role": g1 - g0,
			"execution_ms": doc[0]["Execution Time"]})
		fmt.Println("E-OBJ-1 " + string(line))
		if kind == "" {
			continue
		}
		out[fmt.Sprintf("%d:%s", i+1, kind)] = [2]int64{o1 - o0, g1 - g0}
		if b := objReaderBounds[kind]; o1-o0 > b[0] || g1-g0 > b[1] {
			objTrigger(map[string]any{"subject": subject, "operation": op, "kind": kind, "reason": "statement above its hypothesized reader bound",
				"rls_org_role": o1 - o0, "rls_group_role": g1 - g0, "bound": b})
		}
		if sum := summarize(plan); strings.Contains(fmt.Sprint(sum), "Seq Scan") && strings.Contains(fmt.Sprint(sum), "stored_objects") {
			objTrigger(map[string]any{"subject": subject, "operation": op, "kind": kind, "reason": "sequential scan of stored_objects (observation on a small fixture)"})
		}
	}
	return out
}

// segments splits captured statements into transactions at each context statement.
func segments(stmts []capturedStmt) [][]capturedStmt {
	var out [][]capturedStmt
	for _, s := range stmts {
		if s.step == "C" {
			out = append(out, nil)
		}
		if len(out) > 0 {
			out[len(out)-1] = append(out[len(out)-1], s)
		}
	}
	return out
}

func TestObjectsFootprint(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()
	if _, err := e.setup.Exec(ctx, `ALTER DATABASE `+pgx.Identifier{e.cfg.Database}.Sanitize()+` SET track_functions = 'all'`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = e.setup.Exec(ctx, `ALTER DATABASE `+pgx.Identifier{e.cfg.Database}.Sanitize()+` RESET track_functions`)
	}()
	pool := e.runtimePool(t, nil)

	// The populated fixture: two Organizations with about 100 objects each (rows only; bytes for the measured ones).
	orgX, orgY := e.org(t, "Footprint X "+e.canaries[0]), e.org(t, "Footprint Y "+e.canaries[1])
	creator := e.newUser(t)
	g := e.group(t, orgX, creator, "Footprint G", "footprint-g").ID
	gy := e.group(t, orgY, creator, "Footprint GY", "footprint-gy").ID
	s1, s50 := e.newUser(t), e.newUser(t)
	for _, u := range []*itUser{s1, s50} {
		e.grantOrg(t, orgX, u, tenancy.OrgMember)
		e.grantGroup(t, orgX, g, u, tenancy.GroupMember)
	}
	for i := 0; i < 49; i++ {
		gi := e.group(t, orgX, creator, fmt.Sprintf("Footprint %02d", i), fmt.Sprintf("footprint-%02d", i))
		e.grantGroup(t, orgX, gi.ID, s50, tenancy.GroupMember)
	}
	e.bind(t, s1, orgX) // G016 A2: each subject's credential for orgX, before any measured window
	e.bind(t, s50, orgX)
	e.grantOrg(t, orgY, creator, tenancy.OrgOwner)
	var seeded []objFixture
	for i := 0; i < 200; i++ {
		org, grp := orgX, g
		if i%2 == 1 {
			org, grp = orgY, gy
		}
		if i%5 == 0 {
			grp = tenancy.GroupID{}
		}
		var id tenancy.ObjectID
		_, _ = rand.Read(id[:])
		id[6], id[8] = (id[6]&0x0f)|0x40, (id[8]&0x3f)|0x80
		f := objFixture{id: id, org: org, group: grp, creator: creator.id, state: tenancy.ObjectPending, ct: "application/octet-stream",
			data: []byte{}}
		e.seed(t, f)
		seeded = append(seeded, f)
	}
	defer func() {
		for _, f := range seeded {
			e.unseed(t, f.org, f.id)
		}
	}()

	results := map[string]map[string]map[string][2]int64{}
	for _, sc := range []struct {
		name string
		u    *itUser
	}{{"s1", s1}, {"s50", s50}} {
		h := &hookDB{pool: pool}
		svc := e.objService(h, e.spy, e.ocfg)
		rctx, end := beginAs(t, svc, ctx, sc.u, orgX) // G016 A2
		tc := resolve(t, svc, rctx, orgX, tenancy.GroupID{})
		results[sc.name] = map[string]map[string][2]int64{}
		data := payload(2048, "footprint"+sc.name)
		h.reset()
		h.capture = true
		info, err := svc.PutObject(rctx, tc, PutObjectRequest{Group: g, Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
		if err != nil {
			t.Fatalf("%s put: %v", sc.name, err)
		}
		stmts, locked := h.capturedStmts(), h.lockedRows()
		segs := segments(stmts)
		if len(segs) != 2 {
			t.Fatalf("%s put: %d transactions, want T1 and T2", sc.name, len(segs))
		}
		for i, seg := range segs {
			results[sc.name][fmt.Sprintf("put.t%d", i+1)] = e.objReplay(t, pool, sc.name, fmt.Sprintf("object.put T%d", i+1), seg)
		}
		inventory := map[string]int{}
		for _, l := range locked {
			inventory[l.step+":"+l.mode] += l.rows
		}
		line, _ := json.Marshal(map[string]any{"evidence": "E-OBJ-2", "subject": sc.name, "operation": "object.put", "locked_rows": inventory})
		fmt.Println("E-OBJ-2 " + string(line))
		// T1 and T2 each lock the Organization, the Organization binding, the Group binding and the Group FOR SHARE
		// (beside the accepted session/user liveness locks), and T2 the object row FOR UPDATE; no exclusive
		// Organization or Group lock and no other object lock.
		want := map[string]int{"O:SHARE": 2, "OB:SHARE": 2, "GB:SHARE": 2, "G:SHARE": 2, "?:UPDATE": 1}
		for k, v := range want {
			if inventory[k] != v {
				t.Errorf("%s put locked rows %v, want %v", sc.name, inventory, want)
				break
			}
		}
		for k := range inventory {
			if strings.HasPrefix(k, "?:") && k != "?:UPDATE" || strings.HasPrefix(k, "O:NO KEY") || strings.HasPrefix(k, "Gt:") {
				t.Errorf("%s put took an unexpected lock %s", sc.name, k)
			}
		}
		for _, op := range []struct {
			name string
			call func() error
		}{
			{"object.metadata", func() error { _, err := svc.ObjectMetadata(rctx, tc, info.ID); return err }},
			{"object.delete", func() error { _, err := svc.DeleteObject(rctx, tc, info.ID); return err }},
		} {
			h.reset()
			h.capture = true
			if err := op.call(); err != nil {
				t.Fatalf("%s %s: %v", sc.name, op.name, err)
			}
			stmts, locked := h.capturedStmts(), h.lockedRows()
			segs := segments(stmts)
			results[sc.name][op.name] = e.objReplay(t, pool, sc.name, op.name, segs[0])
			inv := map[string]int{}
			for _, l := range locked {
				inv[l.step+":"+l.mode] += l.rows
			}
			line, _ := json.Marshal(map[string]any{"evidence": "E-OBJ-2", "subject": sc.name, "operation": op.name, "locked_rows": inv,
				"transactions": len(segs)})
			fmt.Println("E-OBJ-2 " + string(line))
		}
		e.unseed(t, orgX, info.ID)
		end()
	}
	// E-OBJ-3: reader calls are equal for 1 and 50 Group bindings.
	a, _ := json.Marshal(results["s1"])
	b, _ := json.Marshal(results["s50"])
	line, _ := json.Marshal(map[string]any{"evidence": "E-OBJ-3", "s1": results["s1"], "s50": results["s50"], "equal": string(a) == string(b)})
	fmt.Println("E-OBJ-3 " + string(line))
	if string(a) != string(b) {
		objTrigger(map[string]any{"reason": "reader calls grow with the subject's binding count", "s1": results["s1"], "s50": results["s50"]})
		t.Errorf("reader calls differ between s1 and s50")
	}

	// E-OBJ-4: while a transfer is blocked (a stub store), no runtime backend is idle in transaction.
	blocked, release := make(chan struct{}), make(chan struct{})
	stub := &objSpy{inner: e.store}
	stub.putOverride = func(ctx context.Context, key objectstore.Key, body io.Reader, size int64, ct string, b objectstore.Binding) error {
		close(blocked)
		<-release
		return e.store.Put(ctx, key, body, size, ct, b)
	}
	svc := e.objService(&hookDB{pool: pool}, stub, e.ocfg)
	rctx, end := beginAs(t, svc, ctx, s1, orgX) // G016 A2
	defer end()
	tc := resolve(t, svc, rctx, orgX, tenancy.GroupID{})
	data := payload(1000, "footprint-transfer")
	done := make(chan ObjectInfo, 1)
	go func() {
		info, err := svc.PutObject(rctx, tc, PutObjectRequest{Group: g, Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
		if err != nil {
			t.Errorf("transfer put: %v", err)
		}
		done <- info
	}()
	select {
	case <-blocked:
	case <-time.After(30 * time.Second):
		t.Fatal("the transfer never started")
	}
	var idle int
	if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND usename = $1
  AND state LIKE 'idle in transaction%'`, e.cfg.RuntimeRole()).Scan(&idle); err != nil {
		t.Fatal(err)
	}
	close(release)
	info := <-done
	line, _ = json.Marshal(map[string]any{"evidence": "E-OBJ-4", "runtime_idle_in_transaction_during_transfer": idle})
	fmt.Println("E-OBJ-4 " + string(line))
	if idle != 0 {
		objTrigger(map[string]any{"reason": "a runtime backend sat idle in transaction during a transfer", "count": idle})
		t.Errorf("%d runtime backends idle in transaction during the transfer", idle)
	}
	if !zero(info.ID) {
		e.unseed(t, orgX, info.ID)
	}
}
