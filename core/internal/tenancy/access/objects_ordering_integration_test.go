//go:build integration

// B-OBJ-ORDERING (G009 Plan r1 §4.4 linearization, §5.2 OOR-1..8): deterministic, sleep-free interleavings (except the
// one bounded TTL wait) of the storage operations with conflicting writers, using the stage seams of the storage
// service, the hookDB statement and COMMIT hooks and the accepted waitWaiters synchronization. Conflicting tenancy
// writers are trusted model primitives on the attested setup pool (G008 pattern). Every case inspects the database and
// the provider afterwards.
package access

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"core/internal/tenancy"
)

// objOrderWorld: Organization A with Groups g1 and g2; an owner, an admin, a g1 admin, two g1 members.
type objOrderWorld struct {
	org                           tenancy.OrganizationID
	g1, g2                        tenancy.GroupID
	owner, admin, g1admin, m1, m2 *itUser
}

func (e *objEnv) orderWorld(t *testing.T, tag string) *objOrderWorld {
	t.Helper()
	w := &objOrderWorld{org: e.org(t, "Objects Ordering "+tag+" "+e.canaries[0])}
	creator := e.newUser(t)
	w.g1 = e.group(t, w.org, creator, "Ordering G1 "+tag, "ordering-g1-"+tag).ID
	w.g2 = e.group(t, w.org, creator, "Ordering G2 "+tag, "ordering-g2-"+tag).ID
	w.owner, w.admin, w.g1admin, w.m1, w.m2 = e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t)
	e.grantOrg(t, w.org, w.owner, tenancy.OrgOwner)
	e.grantOrg(t, w.org, w.admin, tenancy.OrgAdmin)
	for _, u := range []*itUser{w.g1admin, w.m1, w.m2} {
		e.grantOrg(t, w.org, u, tenancy.OrgMember)
	}
	e.grantGroup(t, w.org, w.g1, w.g1admin, tenancy.GroupAdmin)
	e.grantGroup(t, w.org, w.g1, w.m1, tenancy.GroupMember)
	e.grantGroup(t, w.org, w.g1, w.m2, tenancy.GroupMember)
	// G016 A2: each user's credential for org, prepared before any ordering window.
	for _, u := range []*itUser{w.owner, w.admin, w.g1admin, w.m1, w.m2} {
		e.bind(t, u, w.org)
	}
	return w
}

// stageService is a storage service over its own hookDB with a stage hook.
func (e *objEnv) stageService(stage func(op, stage string)) (*Service, *hookDB) {
	h := &hookDB{pool: e.pool}
	s := e.objService(h, e.spy, e.ocfg)
	s.hooks = &objectHooks{stage: stage}
	return s, h
}

// put uploads data through svc as u into group (zero: an Organization object) and returns the info and error, in a
// request scope of u's credential for org (G016 A2).
func (e *objEnv) put(t *testing.T, svc *Service, u *itUser, org tenancy.OrganizationID, group tenancy.GroupID, data []byte) (ObjectInfo, error) {
	t.Helper()
	rctx, end := beginAs(t, svc, context.Background(), u, org)
	defer end()
	tc := resolve(t, svc, rctx, org, tenancy.GroupID{})
	return svc.PutObject(rctx, tc, PutObjectRequest{Group: group, Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
}

// objectsOf lists the stored objects of an Organization as id:state (superuser fixture read).
func (e *objEnv) objectsOf(t *testing.T, org tenancy.OrganizationID) map[tenancy.ObjectID]string {
	t.Helper()
	rows, err := e.setup.Query(context.Background(), `SELECT id::text, state FROM app.stored_objects WHERE organization_id = $1::text::uuid`, org.String())
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[tenancy.ObjectID]string{}
	for rows.Next() {
		var id, st string
		if err := rows.Scan(&id, &st); err != nil {
			t.Fatal(err)
		}
		oid, _ := tenancy.ParseObjectID(id)
		out[oid] = st
	}
	return out
}

// newObjects returns the objects of org that are not in before.
func newObjects(before, after map[tenancy.ObjectID]string) map[tenancy.ObjectID]string {
	out := map[tenancy.ObjectID]string{}
	for id, st := range after {
		if _, ok := before[id]; !ok {
			out[id] = st
		}
	}
	return out
}

func (e *objEnv) keyOf(t *testing.T, org tenancy.OrganizationID, id tenancy.ObjectID) objFixture {
	t.Helper()
	return objFixture{id: id, org: org}
}

func (e *objEnv) cleanupOrg(t *testing.T, org tenancy.OrganizationID) {
	t.Helper()
	for id := range e.objectsOf(t, org) {
		e.unseed(t, org, id)
	}
}

func TestObjectsOrdering(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()

	t.Run("OOR-1 Group removal committed before Put T2: denied, compensated, never readable", func(t *testing.T) {
		w := e.orderWorld(t, "o1")
		defer e.cleanupOrg(t, w.org)
		before := e.objectsOf(t, w.org)
		svc, _ := e.stageService(func(op, stage string) {
			if op == opObjectPut && stage == stageT2 {
				if err := e.model.RemoveGroupMembership(ctx, w.org, w.g1, w.m1.id); err != nil {
					t.Errorf("remove: %v", err)
				}
			}
		})
		_, err := e.put(t, svc, w.m1, w.org, w.g1, payload(500, "oor1"))
		if CodeOf(err) != CodeNotFound {
			t.Fatalf("put after removal: %v, want not_found", err)
		}
		created := newObjects(before, e.objectsOf(t, w.org))
		if len(created) != 1 {
			t.Fatalf("created rows %v", created)
		}
		for id, st := range created {
			k := e.keyOf(t, w.org, id)
			if st != "pending" || e.present(t, k.key(t)) {
				t.Errorf("row %s, provider present %v: want pending and compensated", st, e.present(t, k.key(t)))
			}
			rctx, end := beginAs(t, e.osvc, ctx, w.g1admin, w.org) // G016 A2; G060 AC5: the g1 Group admin may read g1, so not_found is the lifecycle denial
			tc := resolve(t, e.osvc, rctx, w.org, tenancy.GroupID{})
			if _, err := e.osvc.ObjectMetadata(rctx, tc, id); CodeOf(err) != CodeNotFound {
				t.Errorf("the compensated object is readable: %v", err)
			}
			end()
		}
		ordering(t, map[string]any{"case": "OOR-1", "put": string(CodeOf(err)), "row": "pending", "provider": "absent"})
	})

	t.Run("OOR-2 removal racing T2 that holds the facts locks: the remover waits, then reads are denied", func(t *testing.T) {
		w := e.orderWorld(t, "o2")
		defer e.cleanupOrg(t, w.org)
		var h *hookDB
		p := newPause()
		svc, h := e.stageService(func(op, stage string) {
			if op == opObjectPut && stage == stageT2 {
				h.on("before", "COMMIT", p.hook)
			}
		})
		putter := start(func() error { _, err := e.put(t, svc, w.m1, w.org, w.g1, payload(600, "oor2")); return err })
		p.wait(t, "T2 commit")
		remover := start(func() error { return e.model.RemoveGroupMembership(ctx, w.org, w.g1, w.m1.id) })
		e.waitWaiters(t, 1)
		if remover.finished() {
			t.Fatal("the remover did not wait for T2's facts locks")
		}
		close(p.release)
		if err := putter.wait(t, "put"); err != nil {
			t.Fatalf("put serialized before the removal: %v", err)
		}
		if err := remover.wait(t, "remover"); err != nil {
			t.Fatal(err)
		}
		if putter.at > remover.at {
			t.Error("the removal finished before the publishing transaction")
		}
		var id tenancy.ObjectID
		for oid, st := range e.objectsOf(t, w.org) {
			if st != "ready" {
				t.Errorf("row %s", st)
			}
			id = oid
		}
		rctx, end := beginAs(t, e.osvc, ctx, w.m1, w.org) // G016 A2
		defer end()
		tc := resolve(t, e.osvc, rctx, w.org, tenancy.GroupID{})
		if _, err := e.osvc.ObjectMetadata(rctx, tc, id); CodeOf(err) != CodeNotFound {
			t.Errorf("read after the removal: %v, want not_found", err)
		}
		ordering(t, map[string]any{"case": "OOR-2", "order": "publish then removal", "next_read": "not_found"})
	})

	t.Run("OOR-3 Delete D1 versus Put T2, both orders", func(t *testing.T) {
		for _, order := range []string{"put-first", "delete-first"} {
			w := e.orderWorld(t, "o3"+order[:1])
			before := e.objectsOf(t, w.org)
			var h *hookDB
			pPut := newPause()
			pending, gate := make(chan struct{}), make(chan struct{})
			svc, h := e.stageService(func(op, stage string) {
				if op != opObjectPut {
					return
				}
				switch stage {
				case stageProvider:
					close(pending) // the pending row is committed: the deleter can find it
				case stageT2:
					if order == "put-first" {
						h.on("before", "COMMIT", pPut.hook)
					} else {
						<-gate // hold the publish until D1 holds the row lock
					}
				}
			})
			var info ObjectInfo
			putter := start(func() error {
				var err error
				info, err = e.put(t, svc, w.m1, w.org, w.g1, payload(700, "oor3"+order))
				return err
			})
			<-pending
			var id tenancy.ObjectID
			for oid := range newObjects(before, e.objectsOf(t, w.org)) {
				id = oid
			}
			pDel := newPause()
			var dh *hookDB
			dsvc, dh := e.stageService(func(op, stage string) {
				if op == opObjectDelete && stage == stageT1 && order == "delete-first" {
					dh.on("before", "COMMIT", pDel.hook) // D1's COMMIT, not Resolve's
				}
			})
			var res DeleteResult
			deleter := func() error {
				rctx, end := beginAs(t, dsvc, ctx, w.g1admin, w.org) // G016 A2; G060 AC5: the g1 Group admin deletes
				defer end()
				tc := resolve(t, dsvc, rctx, w.org, tenancy.GroupID{})
				var err error
				res, err = dsvc.DeleteObject(rctx, tc, id)
				return err
			}
			var del *actor
			if order == "put-first" {
				pPut.wait(t, "T2 commit")
				del = start(deleter)
				e.waitWaiters(t, 1) // D1's row lock waits for T2
				close(pPut.release)
			} else {
				del = start(deleter)
				pDel.wait(t, "D1 commit")
				close(gate)
				e.waitWaiters(t, 1) // T2's row lock waits for D1
				close(pDel.release)
			}
			putErr := putter.wait(t, "put")
			if err := del.wait(t, "delete"); err != nil || res.Outcome != DeleteOutcomeDeleted {
				t.Fatalf("%s: delete %+v %v", order, res, err)
			}
			k := e.keyOf(t, w.org, id)
			st := e.objectsOf(t, w.org)[id]
			if st != "deleted" || e.present(t, k.key(t)) {
				t.Errorf("%s: final row %s, provider present %v", order, st, e.present(t, k.key(t)))
			}
			switch order {
			case "put-first":
				if putErr != nil || info.ID != id {
					t.Errorf("put-first: the publish serialized first should succeed: %v", putErr)
				}
			case "delete-first":
				if CodeOf(putErr) != CodeNotFound {
					t.Errorf("delete-first: the put after logical deletion: %v, want not_found", putErr)
				}
			}
			ordering(t, map[string]any{"case": "OOR-3", "order": order, "put": string(CodeOf(putErr)), "final": st})
			e.cleanupOrg(t, w.org)
		}
	})

	t.Run("OOR-4 admin to member downgrade: own delete allowed, others' denied, reads kept", func(t *testing.T) {
		w := e.orderWorld(t, "o4")
		defer e.cleanupOrg(t, w.org)
		own, err := e.put(t, e.osvc, w.g1admin, w.org, w.g1, payload(100, "oor4a"))
		if err != nil {
			t.Fatal(err)
		}
		other, err := e.put(t, e.osvc, w.m2, w.org, w.g1, payload(100, "oor4b"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.model.ChangeGroupRole(ctx, w.org, w.g1, w.g1admin.id, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		rctx, end := beginAs(t, e.osvc, ctx, w.g1admin, w.org) // G016 A2
		defer end()
		tc := resolve(t, e.osvc, rctx, w.org, w.g1)
		if _, err := e.osvc.DeleteObject(rctx, tc, other.ID); CodeOf(err) != CodeForbidden {
			t.Errorf("delete another's object after the downgrade: %v, want forbidden", err)
		}
		if _, err := e.osvc.ObjectMetadata(rctx, tc, other.ID); err != nil {
			t.Errorf("read after the downgrade: %v", err)
		}
		if res, err := e.osvc.DeleteObject(rctx, tc, own.ID); err != nil || res.Outcome != DeleteOutcomeDeleted {
			t.Errorf("own delete after the downgrade: %+v %v", res, err)
		}
		ordering(t, map[string]any{"case": "OOR-4", "others": "forbidden", "own": "deleted", "read": "allow"})
	})

	t.Run("OOR-5 session revoke between T1 and T2", func(t *testing.T) {
		w := e.orderWorld(t, "o5")
		defer e.cleanupOrg(t, w.org)
		// G016 A2: the fresh session (formerly relogin) is issued at org's origin; the logout revokes exactly it.
		cred := e.signInAt(t, e.id, w.org, w.m1.sub, "Again", nil, "", "")
		before := e.objectsOf(t, w.org)
		svc, _ := e.stageService(func(op, stage string) {
			if op == opObjectPut && stage == stageT2 {
				e.logout(t, cred, "current")
			}
		})
		rctx, end, err := svc.BeginRequest(ctx, cred, itOriginOf(w.org))
		if err != nil {
			t.Fatalf("BeginRequest: %v", err)
		}
		defer end()
		tc := resolve(t, svc, rctx, w.org, tenancy.GroupID{})
		data := payload(800, "oor5")
		_, err = svc.PutObject(rctx, tc, PutObjectRequest{Group: w.g1, Size: int64(len(data)), Body: io.NopCloser(bytes.NewReader(data))})
		if CodeOf(err) != CodeUnauthenticated {
			t.Fatalf("put across a revocation: %v, want unauthenticated", err)
		}
		for id, st := range newObjects(before, e.objectsOf(t, w.org)) {
			if st != "pending" || e.present(t, e.keyOf(t, w.org, id).key(t)) {
				t.Errorf("row %s: want pending and compensated", st)
			}
		}
		ordering(t, map[string]any{"case": "OOR-5", "put": "unauthenticated", "row": "pending", "provider": "absent"})
	})

	t.Run("OOR-6 DisableUser during an in-flight Get: the transfer finishes, the next call is denied", func(t *testing.T) {
		w := e.orderWorld(t, "o6")
		defer e.cleanupOrg(t, w.org)
		data := payload(200000, "oor6")
		info, err := e.put(t, e.osvc, w.m1, w.org, w.g1, data)
		if err != nil {
			t.Fatal(err)
		}
		// G016 A2: the fresh session (formerly relogin) is issued at org's origin.
		cred := e.signInAt(t, e.id, w.org, w.m2.sub, "Again", nil, "", "")
		rctx, end, err := e.osvc.BeginRequest(ctx, cred, itOriginOf(w.org))
		if err != nil {
			t.Fatalf("BeginRequest: %v", err)
		}
		defer end()
		tc := resolve(t, e.osvc, rctx, w.org, tenancy.GroupID{})
		_, rc, err := e.osvc.GetObject(rctx, tc, info.ID)
		if err != nil {
			t.Fatal(err)
		}
		first := make([]byte, 1000)
		if _, err := io.ReadFull(rc, first); err != nil {
			t.Fatal(err)
		}
		if err := e.id.DisableUser(ctx, w.m2.id.String()); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = e.id.EnableUser(ctx, w.m2.id.String()) }()
		rest, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil || !bytes.Equal(append(first, rest...), data) {
			t.Errorf("the in-flight transfer did not finish: %v", err)
		}
		if _, err := e.osvc.ObjectMetadata(rctx, tc, info.ID); CodeOf(err) != CodeUnauthenticated {
			t.Errorf("the next call after disablement: %v, want unauthenticated", err)
		}
		ordering(t, map[string]any{"case": "OOR-6", "in_flight": "completed", "next": "unauthenticated"})
	})

	t.Run("OOR-7 sign, then revoke: the URL works until expiry, new issuance is denied", func(t *testing.T) {
		w := e.orderWorld(t, "o7")
		defer e.cleanupOrg(t, w.org)
		info, err := e.put(t, e.osvc, w.m1, w.org, w.g1, payload(300, "oor7"))
		if err != nil {
			t.Fatal(err)
		}
		rctx, end := beginAs(t, e.osvc, ctx, w.m2, w.org) // G016 A2
		defer end()
		tc := resolve(t, e.osvc, rctx, w.org, tenancy.GroupID{})
		u, err := e.osvc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: info.ID, TTLSeconds: 2})
		if err != nil {
			t.Fatal(err)
		}
		if err := e.model.RemoveGroupMembership(ctx, w.org, w.g1, w.m2.id); err != nil {
			t.Fatal(err)
		}
		client := &http.Client{Timeout: 10 * time.Second}
		get := func() int {
			resp, err := client.Get(u.URL())
			if err != nil {
				t.Fatalf("fetch: %s", objRedact(err.Error()))
			}
			_ = resp.Body.Close()
			return resp.StatusCode
		}
		if st := get(); st != 200 {
			t.Errorf("the issued URL after revocation: %d, want 200 (residual X1)", st)
		}
		if _, err := e.osvc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: info.ID, DefaultTTL: true}); CodeOf(err) != CodeNotFound {
			t.Errorf("new issuance after revocation: %v, want not_found", err)
		}
		time.Sleep(time.Until(u.ExpiresAt) + 1500*time.Millisecond)
		if st := get(); st != 403 {
			t.Errorf("the URL after expiry: %d, want 403", st)
		}
		ordering(t, map[string]any{"case": "OOR-7", "url_after_revoke": 200, "issuance": "not_found", "url_after_expiry": 403})
	})

	t.Run("OOR-8 delete committed while a Get stream is open: no new read; bounded residual", func(t *testing.T) {
		w := e.orderWorld(t, "o8")
		defer e.cleanupOrg(t, w.org)
		data := payload(300000, "oor8")
		info, err := e.put(t, e.osvc, w.m1, w.org, w.g1, data)
		if err != nil {
			t.Fatal(err)
		}
		rctx, end := beginAs(t, e.osvc, ctx, w.m2, w.org) // G016 A2
		defer end()
		tc := resolve(t, e.osvc, rctx, w.org, tenancy.GroupID{})
		_, rc, err := e.osvc.GetObject(rctx, tc, info.ID)
		if err != nil {
			t.Fatal(err)
		}
		first := make([]byte, 100)
		if _, err := io.ReadFull(rc, first); err != nil {
			t.Fatal(err)
		}
		arctx, aend := beginAs(t, e.osvc, ctx, w.g1admin, w.org) // G016 A2; G060 AC5: the g1 Group admin deletes
		defer aend()
		atc := resolve(t, e.osvc, arctx, w.org, tenancy.GroupID{})
		if res, err := e.osvc.DeleteObject(arctx, atc, info.ID); err != nil || res.Outcome != DeleteOutcomeDeleted {
			t.Fatalf("delete: %+v %v", res, err)
		}
		if _, err := e.osvc.ObjectMetadata(rctx, tc, info.ID); CodeOf(err) != CodeNotFound {
			t.Errorf("new read after deletion: %v", err)
		}
		start := time.Now()
		rest, rerr := io.ReadAll(rc)
		_ = rc.Close()
		outcome := "completed"
		switch {
		case rerr != nil:
			outcome = "failed:" + string(CodeOf(rerr))
		case !bytes.Equal(append(first, rest...), data):
			outcome = "differs"
			t.Error("the open stream returned other bytes without an error")
		}
		if time.Since(start) > e.ocfg.TransferTimeout {
			t.Error("the open stream outlived the transfer deadline")
		}
		ordering(t, map[string]any{"case": "OOR-8", "open_stream": outcome, "new_read": "not_found"})
	})
}
