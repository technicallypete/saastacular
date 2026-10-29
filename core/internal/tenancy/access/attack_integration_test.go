//go:build integration

// B-OBJ-IT second invocation (G010 Plan r1 §3.3, §3.4): TestIsolationAttacks, the foundation isolation attack suite.
// Chains A-E compose the accepted single-boundary guarantees across the session/context boundary, the guarded facade,
// raw runtime SQL and the actual provider, against real PostgreSQL at version 7 (the attested runtime login), real G004
// sessions through the fake WorkOS and the private MinIO of the migrations suite (the scoped credential). It reuses the
// accepted access integration environment (setupObjects) and its mechanics unchanged; no existing test file is edited.
//
// The independent oracle (attack_expected_test.go, written without this file or the policy code) holds every literal
// expectation. Each step below is explicit code that runs the oracle's listed operations in order, on one world
// (oracle AMB-01), and looks each literal up by its expectation ID, checking that the ID belongs to the step and names
// the target the code acts on. Every expectation must be checked exactly once. The glue prints only
// ISO-ORACLE {"cases":K,"disagreements":D} and ISO-ORACLE-DISAGREEMENT <step>/<ID> lines, never a UUID, URL or payload.
// A disagreement is reported, never resolved by editing the oracle.
//
// Denial evidence (Plan r1 §1): a pre-provider denial counts only with the oracle's exact external code, its paired
// authorized control checked earlier and in agreement, a zero provider-call delta, unchanged Head presence and an
// unchanged stored-object row hash, and an unchanged owner-SQL hash of the tenancy relations. The staged denials (E1,
// E3) assert their accepted staged effects and a working authorized control instead. A missing prerequisite is
// t.Fatal; B-OBJ-IT's it_count fails on any SKIP or FAIL, so infrastructure failure never passes.
//
// Required environment: setupObjects'.
package access

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/identity"
	"core/internal/objectstore"
	"core/internal/tenancy"
)

// --- oracle bookkeeping ------------------------------------------------------------------------------------------------

// atkObs is one observation in the oracle's vocabulary. An empty field (or provider -1) was not measured.
type atkObs struct {
	code     string // "ok" | access code | "committed"
	result   string // delete outcome
	visible  []string
	sql      bool   // visible is meaningful (an empty set is a valid observation)
	row      string // object state | "absent" | "updated" | "unchanged"
	bytes    string // "present" | "absent"
	provider int64  // provider-call delta; -1 when not measured
	url      string // "200" | "non2xx"
	tx       string // "begun" | "not_begun"
	creator  string
	affected string
	backend  string
	settings string
	reader   string // r2: app.rls_profile_eligible result
}

func atkNew() atkObs { return atkObs{provider: -1} }

// atkOracle tracks every expectation: each must be checked exactly once, and a paired positive must have been checked
// and agreed before its denial.
type atkOracle struct {
	mu     sync.Mutex
	want   map[string]atkExpectation
	order  []string
	used   map[string]bool
	agreed map[string]bool
	n, dis int
}

func newAtkOracle(t *testing.T) *atkOracle {
	t.Helper()
	o := &atkOracle{want: map[string]atkExpectation{}, used: map[string]bool{}, agreed: map[string]bool{}}
	for _, w := range atkExpected() {
		if _, dup := o.want[w.ID]; dup {
			t.Fatalf("oracle: duplicate expectation %s", w.ID)
		}
		o.want[w.ID] = w
		o.order = append(o.order, w.ID)
	}
	if len(o.want) == 0 {
		t.Fatal("oracle: no expectations")
	}
	return o
}

// expect returns the expectation id after checking that it belongs to step and names target.
func (o *atkOracle) expect(t *testing.T, step, id, target string) atkExpectation {
	t.Helper()
	w, ok := o.want[id]
	switch {
	case !ok:
		t.Fatalf("glue: no oracle expectation %s", id)
	case w.Step != step:
		t.Fatalf("glue: expectation %s belongs to step %s, not %s", id, w.Step, step)
	case target != "" && w.Target != target:
		t.Fatalf("glue: expectation %s targets %s; the step acts on %s", id, w.Target, target)
	}
	return w
}

// check compares an observation with expectation id on every field the oracle asserts.
func (o *atkOracle) check(t *testing.T, step, id, target string, ob atkObs) bool {
	t.Helper()
	w := o.expect(t, step, id, target)
	o.mu.Lock()
	if o.used[id] {
		o.mu.Unlock()
		t.Fatalf("glue: expectation %s checked twice", id)
	}
	o.used[id] = true
	o.n++
	pairOK := w.Pair == "" || o.agreed[w.Pair]
	o.mu.Unlock()
	var diffs []string
	cmp := func(field, want, got string) {
		if want == "" {
			return
		}
		if got == "" {
			t.Fatalf("glue: %s asserts %s, which the step did not measure", id, field)
		}
		if want != got {
			diffs = append(diffs, fmt.Sprintf("%s %s, oracle %s", field, got, want))
		}
	}
	cmp("code", w.Code, ob.code)
	cmp("result", w.Result, ob.result)
	if w.Visible != nil {
		if !ob.sql {
			t.Fatalf("glue: %s asserts a visible set, which the step did not measure", id)
		}
		want := slices.Clone(w.Visible)
		sort.Strings(want)
		if !slices.Equal(want, ob.visible) {
			diffs = append(diffs, fmt.Sprintf("visible %v, oracle %v", ob.visible, want))
		}
	}
	cmp("row", w.Effect.Row, ob.row)
	cmp("bytes", w.Effect.Bytes, ob.bytes)
	if w.Effect.Provider != "" {
		got := ""
		switch {
		case ob.provider == 0:
			got = atkCallsZero
		case ob.provider > 0:
			got = atkCallsNonzero
		}
		cmp("provider", w.Effect.Provider, got)
	}
	cmp("url", w.Effect.URL, ob.url)
	cmp("tx", w.Effect.Tx, ob.tx)
	cmp("creator", w.Effect.Creator, ob.creator)
	cmp("affected rows", w.Effect.AffectedRows, ob.affected)
	cmp("backend", w.Effect.Backend, ob.backend)
	cmp("settings", w.Effect.Settings, ob.settings)
	cmp("reader", w.Effect.Reader, ob.reader)
	if !pairOK {
		t.Errorf("%s: its paired positive %s was not checked in agreement earlier (Plan §1 rule 2)", id, w.Pair)
	}
	if len(diffs) > 0 {
		o.mu.Lock()
		o.dis++
		o.mu.Unlock()
		fmt.Printf("ISO-ORACLE-DISAGREEMENT %s/%s\n", step, id)
		t.Errorf("ISO-ORACLE-DISAGREEMENT %s/%s: %s", step, id, strings.Join(diffs, "; "))
		return false
	}
	o.mu.Lock()
	o.agreed[id] = true
	o.mu.Unlock()
	return true
}

// finish prints the summary line and fails on any expectation the glue did not check.
func (o *atkOracle) finish(t *testing.T) {
	t.Helper()
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, id := range o.order {
		if !o.used[id] {
			t.Errorf("glue: oracle expectation %s was not checked", id)
		}
	}
	fmt.Printf("ISO-ORACLE {\"cases\":%d,\"disagreements\":%d}\n", o.n, o.dis)
}

// atkCitations returns every Inherited citation of the oracle: expectations, actors and inherited-only topics.
func atkCitations() []string {
	var out []string
	for _, w := range atkExpected() {
		out = append(out, w.Inherited...)
	}
	f := atkExpectedFixture()
	for _, a := range f.Actors {
		out = append(out, a.Inherited...)
	}
	for _, i := range f.InheritedOnly {
		out = append(out, i.Cites...)
	}
	return out
}

// --- the attack world --------------------------------------------------------------------------------------------------

// atkWorld instantiates the oracle fixture with fresh UUIDs and distinct names through the existing builders (the
// privileged e.model as superuser, real G004 sign-ins) and guarded puts with the exact creators.
type atkWorld struct {
	e     *objEnv
	o     *atkOracle
	org   map[string]tenancy.OrganizationID
	grp   map[string]tenancy.GroupID
	prof  map[string]tenancy.BusinessProfileID
	usr   map[string]*itUser
	obj   map[string]tenancy.ObjectID
	objOf map[string]tenancy.OrganizationID // object key -> its Organization
	data  map[string][]byte                 // object key -> its bytes
	label map[string]string                 // UUID, "subject|scope" or "group|profile" -> fixture key
	lab   map[string]string                 // r5 (G020): label definition key -> UUID
	seq   int
}

func (x *atkWorld) addObject(key string, org tenancy.OrganizationID, id tenancy.ObjectID, data []byte) {
	x.obj[key], x.objOf[key], x.data[key] = id, org, data
	x.label[id.String()] = key
}

func (e *objEnv) atkWorld(t *testing.T, o *atkOracle) *atkWorld {
	t.Helper()
	ctx := context.Background()
	f := atkExpectedFixture()
	x := &atkWorld{e: e, o: o, org: map[string]tenancy.OrganizationID{}, grp: map[string]tenancy.GroupID{},
		prof: map[string]tenancy.BusinessProfileID{}, usr: map[string]*itUser{}, obj: map[string]tenancy.ObjectID{},
		objOf: map[string]tenancy.OrganizationID{}, data: map[string][]byte{}, label: map[string]string{}}
	c := e.canaries
	for i, fo := range f.Organizations {
		id := e.org(t, "Attack "+fo.Key+" "+c[i%len(c)])
		e.setFlag(t, id, fo.GroupAdminsEditBusinessProfiles)
		x.org[fo.Key], x.label[id.String()] = id, fo.Key
	}
	for _, fa := range f.Actors {
		if fa.Session != "real" {
			t.Fatalf("fixture actor %s: session %q", fa.Key, fa.Session)
		}
		u := e.newUser(t)
		x.usr[fa.Key], x.label[u.id.String()] = u, fa.Key
	}
	// A Group's creator receives group_owner (tenancy.CreateGroup). The creator is a fixture-only global user whose
	// binding is removed at once, so exactly the oracle's bindings hold.
	creator := e.newUser(t)
	for i, fg := range f.Groups {
		org := x.org[fg.Org]
		gr := e.group(t, org, creator, "Attack "+fg.Key+" "+c[(i+1)%len(c)], fg.Slug)
		if err := e.model.RemoveGroupMembership(ctx, org, gr.ID, creator.id); err != nil {
			t.Fatalf("remove the creator binding of %s: %v", fg.Key, err)
		}
		x.grp[fg.Key], x.label[gr.ID.String()] = gr.ID, fg.Key
	}
	for i, fp := range f.Profiles {
		pr := e.profile(t, x.org[fp.Org], "Attack "+fp.Key+" "+c[(i+2)%len(c)])
		x.prof[fp.Key], x.label[pr.ID.String()] = pr.ID, fp.Key
	}
	groupOrg := map[string]string{}
	for _, fg := range f.Groups {
		groupOrg[fg.Key] = fg.Org
	}
	for _, fl := range f.Links {
		e.link(t, x.org[groupOrg[fl.Group]], x.grp[fl.Group], x.prof[fl.Profile])
		x.label[x.grp[fl.Group].String()+"|"+x.prof[fl.Profile].String()] = fl.Key
	}
	bindings := map[string]atkBinding{}
	for _, b := range f.Bindings {
		bindings[b.Key] = b
	}
	for _, fa := range f.Actors {
		u := x.usr[fa.Key]
		for _, bk := range fa.Bindings {
			b, ok := bindings[bk]
			if !ok || b.Actor != fa.Key {
				t.Fatalf("fixture binding %s of %s", bk, fa.Key)
			}
			org := x.org[b.Org]
			if b.Group == "" {
				e.grantOrg(t, org, u, tenancy.OrganizationRole(b.Role))
				x.label[u.id.String()+"|"+org.String()] = b.Key
				continue
			}
			e.grantGroup(t, org, x.grp[b.Group], u, tenancy.GroupRole(b.Role))
			x.label[u.id.String()+"|"+x.grp[b.Group].String()] = b.Key
		}
	}
	// r5 (G020): the label rows, inserted by the attested superuser (fixtures only). An Organization-owned definition
	// carries the Organization as its owner facts; an assignment is the production shape (Organization owner facts,
	// target_type 'group', target_id = target_group_id) and is keyed "<definition UUID>|<target Group UUID>".
	x.lab = map[string]string{}
	labelOrg := map[string]string{}
	for i, fl := range f.Labels {
		id := rlsUUID(t)
		typ, owner, group := "organization", x.org[fl.Org].String(), any(nil)
		if fl.Group != "" {
			typ, owner, group = "group", x.grp[fl.Group].String(), x.grp[fl.Group].String()
		}
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.label_definitions (id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility)
VALUES ($1::text::uuid, $2::text::uuid, $3, $4::text::uuid, $5::text::uuid, $6, $7)`, id, x.org[fl.Org].String(), typ, owner, group,
			"Attack "+fl.Key+" "+c[i%len(c)], fl.Visibility); err != nil {
			t.Fatalf("fixture label %s: %v", fl.Key, err)
		}
		x.lab[fl.Key], x.label[id], labelOrg[fl.Key] = id, fl.Key, fl.Org
	}
	for _, fa := range f.LabelAssignments {
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.label_assignments (id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id, target_group_id)
VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid, 'organization', $2::text::uuid, 'group', $4::text::uuid, $4::text::uuid)`,
			rlsUUID(t), x.org[labelOrg[fa.Label]].String(), x.lab[fa.Label], x.grp[fa.Group].String()); err != nil {
			t.Fatalf("fixture label assignment %s: %v", fa.Key, err)
		}
		x.label[x.lab[fa.Label]+"|"+x.grp[fa.Group].String()] = fa.Key
	}
	// G016 A2: one prepared credential per (actor, Organization) at that Organization's origin, for every Organization the
	// actor holds a binding in (a Group-only actor is admitted only for the issuance); an actor without bindings gets an
	// orgA credential and stays membership-negative.
	for _, fa := range f.Actors {
		var orgs []tenancy.OrganizationID
		for _, bk := range fa.Bindings {
			if o := x.org[bindings[bk].Org]; !slices.Contains(orgs, o) {
				orgs = append(orgs, o)
			}
		}
		if len(orgs) == 0 {
			orgs = append(orgs, x.org["orgA"])
		}
		e.bind(t, x.usr[fa.Key], orgs...)
	}
	for i, fob := range f.Objects {
		if fob.Origin != "seed" {
			continue
		}
		org := x.org[fob.Org]
		data := payload(256+i*7, "attack seed "+fob.Key+" "+c[0])
		info, err := e.put(t, e.osvc, x.usr[fob.Creator], org, x.grp[fob.Group], data)
		if err != nil {
			t.Fatalf("fixture object %s: %v", fob.Key, err)
		}
		x.addObject(fob.Key, org, info.ID, data)
	}
	return x
}

// stepObject returns the oracle's org and group of an object created by a step (Origin "step:<ID>").
func atkStepObject(t *testing.T, key, step string) atkObject {
	t.Helper()
	for _, o := range atkExpectedFixture().Objects {
		if o.Key == key {
			if o.Origin != "step:"+step {
				t.Fatalf("glue: object %s has origin %s, not step:%s", key, o.Origin, step)
			}
			return o
		}
	}
	t.Fatalf("glue: unknown fixture object %s", key)
	return atkObject{}
}

func (x *atkWorld) keyOfObject(t *testing.T, key string) objFixture {
	t.Helper()
	id, ok := x.obj[key]
	if !ok {
		t.Fatalf("glue: unknown object %s", key)
	}
	return objFixture{id: id, org: x.objOf[key]}
}

// objState measures one world object: its row state and provider presence.
func (x *atkWorld) objState(t *testing.T, key string) (row, bytes string) {
	t.Helper()
	k := x.keyOfObject(t, key)
	var st string
	if err := x.e.setup.QueryRow(context.Background(), `SELECT coalesce((SELECT state FROM app.stored_objects WHERE id = $1::text::uuid), 'absent')`,
		k.id.String()).Scan(&st); err != nil {
		t.Fatal(err)
	}
	bytes = atkBytesAbsent
	if x.e.present(t, k.key(t)) {
		bytes = atkBytesPresent
	}
	return st, bytes
}

// creatorOf returns the fixture key of an object's stored creator.
func (x *atkWorld) creatorOf(t *testing.T, key string) string {
	t.Helper()
	var c string
	if err := x.e.setup.QueryRow(context.Background(), `SELECT creator_user_id::text FROM app.stored_objects WHERE id = $1::text::uuid`,
		x.obj[key].String()).Scan(&c); err != nil {
		t.Fatal(err)
	}
	if l, ok := x.label[c]; ok {
		return l
	}
	return "?unknown"
}

// tenantRow hashes one Group or profile row (owner SQL), for the "updated"/"unchanged" effects.
func (x *atkWorld) tenantRow(t *testing.T, key string) string {
	t.Helper()
	q, id := "", ""
	if g, ok := x.grp[key]; ok {
		q, id = `SELECT coalesce((SELECT md5(r::text) FROM app.groups r WHERE r.id = $1::text::uuid), '-')`, g.String()
	} else if p, ok := x.prof[key]; ok {
		q, id = `SELECT coalesce((SELECT md5(r::text) FROM app.business_profiles r WHERE r.id = $1::text::uuid), '-')`, p.String()
	} else {
		t.Fatalf("glue: %s is not a Group or profile", key)
	}
	var s string
	if err := x.e.setup.QueryRow(context.Background(), q, id).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func (x *atkWorld) orgs() []tenancy.OrganizationID {
	out := make([]tenancy.OrganizationID, 0, len(x.org))
	for _, id := range x.org {
		out = append(out, id)
	}
	return out
}

// allObjects lists every stored object of the world's Organizations (owner SQL).
func (x *atkWorld) allObjects(t *testing.T) map[tenancy.ObjectID]string {
	t.Helper()
	out := map[tenancy.ObjectID]string{}
	for _, org := range x.orgs() {
		for id, st := range x.e.objectsOf(t, org) {
			out[id] = st
		}
	}
	return out
}

// atkSnap is the before-state of an operation: stored-object rows, tenancy rows and the provider-call counter.
type atkSnap struct {
	rows, state string
	calls       int64
	objects     map[tenancy.ObjectID]string
}

func (x *atkWorld) snap(t *testing.T) atkSnap {
	t.Helper()
	orgs := x.orgs()
	return atkSnap{rows: x.e.objectRows(t, orgs...), state: x.e.state(t, orgs...), calls: x.e.spy.calls.Load(), objects: x.allObjects(t)}
}

// unchanged asserts the §1 pre-provider denial evidence: zero provider calls, unchanged stored-object rows and
// unchanged tenancy rows. It returns the provider-call delta.
func (x *atkWorld) unchanged(t *testing.T, what string, s atkSnap) int64 {
	t.Helper()
	now := x.snap(t)
	d := now.calls - s.calls
	if d != 0 {
		t.Errorf("%s: %d provider calls, want 0", what, d)
	}
	if now.rows != s.rows {
		t.Errorf("%s: the stored-object rows changed", what)
	}
	if now.state != s.state {
		t.Errorf("%s: the tenancy rows changed", what)
	}
	return d
}

// newRow is the "absent" effect of a Put: no row was created since s, or the state of the one created.
func (x *atkWorld) newRow(t *testing.T, s atkSnap) (string, tenancy.ObjectID) {
	t.Helper()
	created := newObjects(s.objects, x.allObjects(t))
	switch len(created) {
	case 0:
		return atkRowAbsent, tenancy.ObjectID{}
	case 1:
		for id, st := range created {
			return st, id
		}
	}
	t.Fatalf("%d rows created by one Put", len(created))
	return "", tenancy.ObjectID{}
}

func atkCode(err error) string {
	if err == nil {
		return atkOK
	}
	if c := CodeOf(err); c != "" {
		return string(c)
	}
	return "error"
}

func atkCommit(err error) string {
	if err != nil {
		return "event failed"
	}
	return atkCommitted
}

func atkBody(data []byte) io.ReadCloser { return io.NopCloser(bytes.NewReader(data)) }

// putReq is a Put of fresh bytes into group (zero: an Organization object).
func (x *atkWorld) putReq(group tenancy.GroupID, tag string) (PutObjectRequest, []byte) {
	x.seq++
	data := payload(300+x.seq, fmt.Sprintf("attack put %s %d %s", tag, x.seq, x.e.canaries[1]))
	return PutObjectRequest{Group: group, ContentType: "application/octet-stream", Size: int64(len(data)), Body: atkBody(data)}, data
}

// begins counts the transactions a hookDB began.
func atkBegins(h *hookDB) int { n, _ := h.counts(); return n }

func atkTx(before, after int) string {
	if after > before {
		return atkTxBegun
	}
	return atkTxNotBegun
}

// --- one operation -----------------------------------------------------------------------------------------------------

// atkOp runs one guarded operation as expectation id and checks it. kind selects the measured effects: an object
// target (row state, bytes), a Put (the new row, its creator), a Group or profile target ("updated"/"unchanged") or
// none. h is the service's hookDB (Tx). A denial (any code but ok) also asserts the §1 pre-provider evidence unless
// staged is set. It returns the operation's error and, for a successful Put, the new object's ID.
type atkKind int

const (
	atkKindNone atkKind = iota
	atkKindObject
	atkKindPut
	atkKindTenant
)

type atkRun struct {
	step, id, target string
	kind             atkKind
	h                *hookDB
	staged           bool                              // a staged operation: the zero-I/O rule does not apply
	fn               func() (result string, err error) // the operation; result is a delete outcome
	extra            func(ob *atkObs, err error)       // further measurements (URL, backend)
	putKey           string                            // the object key a Put creates
	putOrg           tenancy.OrganizationID            // its Organization
	putData          func() []byte                     // its bytes
	control          bool                              // an authorized control: a failure stops the step
}

func (x *atkWorld) run(t *testing.T, r atkRun) (tenancy.ObjectID, error) {
	t.Helper()
	x.o.expect(t, r.step, r.id, r.target)
	s := x.snap(t)
	var rowB, bytesB, tenB string
	switch r.kind {
	case atkKindObject:
		rowB, bytesB = x.objState(t, r.target)
	case atkKindTenant:
		tenB = x.tenantRow(t, r.target)
	}
	var b0 int
	if r.h != nil {
		b0 = atkBegins(r.h)
	}
	result, err := r.fn()
	ob := atkNew()
	ob.code, ob.result = atkCode(err), result
	ob.provider = x.e.spy.calls.Load() - s.calls
	if r.h != nil {
		ob.tx = atkTx(b0, atkBegins(r.h))
	}
	var created tenancy.ObjectID
	switch r.kind {
	case atkKindObject:
		ob.row, ob.bytes = x.objState(t, r.target)
	case atkKindTenant:
		ob.row = atkRowUnchanged
		if x.tenantRow(t, r.target) != tenB {
			ob.row = atkRowUpdated
		}
	case atkKindPut:
		ob.row, created = x.newRow(t, s)
		// A denied Put has no object key: a row it created is reported through the row effect and unchanged().
		if !zero(created) && r.putKey != "" {
			x.addObject(r.putKey, r.putOrg, created, r.putData())
			_, ob.bytes = x.objState(t, r.putKey)
			ob.creator = x.creatorOf(t, r.putKey)
		}
	}
	if r.extra != nil {
		r.extra(&ob, err)
	}
	if err != nil && !r.staged {
		x.unchanged(t, r.step+"/"+r.id, s)
		if r.kind == atkKindObject && (ob.row != rowB || ob.bytes != bytesB) {
			t.Errorf("%s/%s: object state %s/%s, before %s/%s", r.step, r.id, ob.row, ob.bytes, rowB, bytesB)
		}
	}
	ok := x.o.check(t, r.step, r.id, r.target, ob)
	if r.control && (err != nil || !ok) {
		t.Fatalf("%s/%s: the authorized control failed (%s); later denials would not count", r.step, r.id, ob.code)
	}
	return created, err
}

// event commits one accepted lifecycle writer as expectation id ("committed").
func (x *atkWorld) event(t *testing.T, step, id, target string, err error) {
	t.Helper()
	ob := atkNew()
	ob.code = atkCommit(err)
	if !x.o.check(t, step, id, target, ob) {
		t.Fatalf("%s/%s: the event did not commit", step, id)
	}
}

// --- raw runtime SQL ---------------------------------------------------------------------------------------------------

// atkShape runs fn in a rolled-back transaction on the attested direct runtime login, with the production context
// statement (tenancy.InitContext) of the given subject, Organization and optional Group as its first statement.
func atkShape(t *testing.T, conn *pgx.Conn, user tenancy.UserID, org tenancy.OrganizationID, group tenancy.GroupID, fn func(tx pgx.Tx)) {
	t.Helper()
	ctx := context.Background()
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := tenancy.InitContext(ctx, tx, user, org, group); err != nil {
		t.Fatalf("context statement: %v", err)
	}
	fn(tx)
}

// G018 (oracle r2): the raw runtime writes of B2/C3 — an association, and a direct group_owner binding (organization,
// subject, Group) outside the creation statement.
const (
	atkInsertLink  = `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid)`
	atkInsertOwner = `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'group', $3::text::uuid, $3::text::uuid, 'group_owner', 'membership')`
)

// G019 (oracle r3): the raw runtime writes of F1/F2 — an Organization-scope binding INSERT, a binding role UPDATE or
// DELETE by (subject, scope) and the invitation column probes (AMB-27, AMB-28).
const (
	atkInsertOrgBinding = `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_member', 'membership')`
	atkUpdateRole    = `UPDATE app.role_bindings SET role = $3 WHERE subject_user_id = $1::text::uuid AND scope_id = $2::text::uuid`
	atkDeleteBinding = `DELETE FROM app.role_bindings WHERE subject_user_id = $1::text::uuid AND scope_id = $2::text::uuid`
	atkInvDigest     = `SELECT token_digest FROM app.invitations WHERE id = $1::text::uuid`
	atkInvConsume    = `UPDATE app.invitations SET consumed_at = now(), consumed_by_user_id = $2::text::uuid WHERE id = $1::text::uuid`
	atkInvRole       = `UPDATE app.invitations SET role = 'org_owner' WHERE id = $1::text::uuid`
	atkInvExpiry     = `UPDATE app.invitations SET expires_at = now() + interval '365 days' WHERE id = $1::text::uuid`
	atkInvDigestSet  = `UPDATE app.invitations SET token_digest = '\x00'::bytea WHERE id = $1::text::uuid`
	atkInvRevoke     = `UPDATE app.invitations SET revoked_at = now() WHERE id = $1::text::uuid`
	atkOrgClaim      = `SELECT app.invitation_claim_organization($1::text::uuid, $2::bytea, $3, $4, $5)`
)

// membershipDigest is one superuser-read digest of every role_bindings and invitations row of the world's Organizations
// (F1/F2 post-state: a failed attack preserves membership and invitation state, SPEC-R3 "Attribution and atomicity").
func (x *atkWorld) membershipDigest(t *testing.T) string {
	t.Helper()
	orgs := []string{}
	for _, o := range x.orgs() {
		orgs = append(orgs, o.String())
	}
	var s string
	if err := x.e.setup.QueryRow(context.Background(), `SELECT md5(coalesce((SELECT string_agg(r::text, ',' ORDER BY r.id) FROM app.role_bindings r
WHERE r.organization_id::text = ANY($1::text[])), '') || '/' || coalesce((SELECT string_agg(i::text, ',' ORDER BY i.id) FROM app.invitations i
WHERE i.organization_id::text = ANY($1::text[])), ''))`, orgs).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

// atkRelations are the visible-set queries: each row becomes the fixture key of its UUID, subject|scope or
// group|profile pair (r5: or label|target Group pair); a row outside the world becomes "?unknown".
var atkRelations = map[string]string{
	"organizations":                 `SELECT id::text FROM app.organizations`,
	"groups":                        `SELECT id::text FROM app.groups`,
	"business_profiles":             `SELECT id::text FROM app.business_profiles`,
	"group_business_profiles":       `SELECT group_id::text || '|' || business_profile_id::text FROM app.group_business_profiles`,
	"role_bindings":                 `SELECT subject_user_id::text || '|' || scope_id::text FROM app.role_bindings`,
	"organization_memberships":      `SELECT user_id::text || '|' || organization_id::text FROM app.organization_memberships`,
	"group_memberships":             `SELECT user_id::text || '|' || group_id::text FROM app.group_memberships`,
	"business_profile_access_paths": `SELECT user_id::text || '|' || group_id::text || '|' || business_profile_id::text FROM app.business_profile_access_paths`,
	"stored_objects":                `SELECT id::text FROM app.stored_objects`,
	"users":                         `SELECT id::text FROM app.users`,
	"label_definitions":             `SELECT id::text FROM app.label_definitions`,
	"label_assignments":             `SELECT label_id::text || '|' || target_id::text FROM app.label_assignments`,
	"label_assignment_joins":        `SELECT d.id::text || '|' || g.id::text FROM app.label_assignments a JOIN app.label_definitions d ON d.id = a.label_id JOIN app.groups g ON g.id = a.target_group_id`,
}

// visible checks expectation id: the sorted fixture keys of the rows of its Table visible in tx.
func (x *atkWorld) visible(t *testing.T, tx pgx.Tx, step, id string) {
	t.Helper()
	ctx := context.Background()
	w := x.o.expect(t, step, id, "")
	q, ok := atkRelations[w.Table]
	if !ok {
		t.Fatalf("glue: no query for table %s (%s)", w.Table, id)
	}
	sp, err := tx.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sp.Rollback(ctx) }()
	rows, err := sp.Query(ctx, q)
	if err != nil {
		t.Fatalf("visible %s: %s", w.Table, pgState(err))
	}
	defer rows.Close()
	keys := []string{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		if l, ok := x.label[v]; ok {
			keys = append(keys, l)
		} else {
			keys = append(keys, "?unknown")
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("visible %s: %s", w.Table, pgState(err))
	}
	sort.Strings(keys)
	ob := atkNew()
	ob.visible, ob.sql = keys, true
	x.o.check(t, step, id, "", ob)
}

// --- signed URL and backend probes -------------------------------------------------------------------------------------

// atkFetch GETs a signed URL and returns "200" (exact bytes) or "non2xx". Neither the URL nor the body is kept or
// printed.
func atkFetch(t *testing.T, u SignedURL, want []byte) string {
	t.Helper()
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.URL())
	if err != nil {
		t.Fatal("fetch failed") // no error text: a transport error can carry the object path
	}
	body, rerr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	switch {
	case resp.StatusCode < 200 || resp.StatusCode > 299:
		return atkURLNon2xx
	case resp.StatusCode != 200:
		return fmt.Sprint(resp.StatusCode)
	case rerr != nil || sha256.Sum256(body) != sha256.Sum256(want):
		return "200 with other bytes"
	}
	return atkURL200
}

// atkBackend returns the backend PID and the four context settings of the one connection of pool, by a context-free
// probe (no context statement runs).
func atkBackend(t *testing.T, pool *pgxpool.Pool) (int, string) {
	t.Helper()
	var pid int
	var s string
	if err := pool.QueryRow(context.Background(), `SELECT pg_backend_pid(), concat_ws('|', coalesce(current_setting('app.user_id', true), ''),
  coalesce(current_setting('app.organization_id', true), ''), coalesce(current_setting('app.group_id', true), ''),
  coalesce(current_setting('app.login_binding', true), ''))`).Scan(&pid, &s); err != nil {
		t.Fatal(err)
	}
	return pid, s
}

// --- the suite ---------------------------------------------------------------------------------------------------------

func TestIsolationAttacks(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()
	zg := tenancy.GroupID{}

	t.Run("ISO-ATTEST runtime login and scoped credential", func(t *testing.T) {
		conn, err := e.pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if err := e.rlsAttest(conn); err != nil {
			t.Fatal(err)
		}
		conn.Release()
		_ = e.rlsRuntime(t) // attested before use
		if err := e.rlsAttest(e.setup); err == nil {
			t.Fatal("the privileged setup connection passed the runtime attestation (the privileged-assertion negative)")
		}
		for _, k := range []string{"MINIO_ROOT_USER", "MINIO_ROOT_PASSWORD"} {
			if os.Getenv(k) != "" {
				t.Fatalf("%s is present in the test environment", k)
			}
		}
		if e.scfg.AccessKey == "" || e.scfg.SecretKey == "" {
			t.Fatal("no scoped credential")
		}
	})

	o := newAtkOracle(t)
	x := e.atkWorld(t, o)
	g, p, u, ob := x.grp, x.prof, x.usr, x.obj
	A, B := x.org["orgA"], x.org["orgB"]
	svc, h := e.osvc, e.hook
	rt := e.rlsRuntime(t)

	// put is a guarded Put through s as expectation id, creating object key (from the oracle's Origin).
	put := func(t *testing.T, s *Service, hh *hookDB, rctx context.Context, tc *TenantContext, step, id, target, key, tag string, control, staged bool) tenancy.ObjectID {
		t.Helper()
		fo := atkStepObject(t, key, step)
		req, data := x.putReq(x.grp[fo.Group], tag)
		created, _ := x.run(t, atkRun{step: step, id: id, target: target, kind: atkKindPut, h: hh, putKey: key, putOrg: x.org[fo.Org],
			putData: func() []byte { return data }, control: control, staged: staged,
			fn: func() (string, error) { _, err := s.PutObject(rctx, tc, req); return "", err }})
		return created
	}
	// putDenied is a denied Put into group (no object created).
	putDenied := func(t *testing.T, s *Service, hh *hookDB, rctx context.Context, tc *TenantContext, step, id, target string, group tenancy.GroupID) {
		t.Helper()
		x.run(t, atkRun{step: step, id: id, target: target, kind: atkKindPut, h: hh,
			fn: func() (string, error) {
				req, _ := x.putReq(group, step+id)
				_, err := s.PutObject(rctx, tc, req)
				return "", err
			}})
	}
	sign := func(t *testing.T, rctx context.Context, tc *TenantContext, step, id, key string, fetch bool) {
		t.Helper()
		var url SignedURL
		x.run(t, atkRun{step: step, id: id, target: key, kind: atkKindObject, h: h, control: fetch,
			fn: func() (string, error) {
				var err error
				url, err = svc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: ob[key], DefaultTTL: true})
				return "", err
			},
			extra: func(o *atkObs, err error) {
				if fetch && err == nil {
					o.url = atkFetch(t, url, x.data[key])
				}
			}})
	}
	del := func(t *testing.T, s *Service, hh *hookDB, rctx context.Context, tc *TenantContext, step, id, key string, control bool) {
		t.Helper()
		x.run(t, atkRun{step: step, id: id, target: key, kind: atkKindObject, h: hh, control: control,
			fn: func() (string, error) {
				res, err := s.DeleteObject(rctx, tc, ob[key])
				return string(res.Outcome), err
			}})
	}
	facade := func(t *testing.T, rctx context.Context, step, id, target string, control bool, fn func() error) {
		t.Helper()
		x.run(t, atkRun{step: step, id: id, target: target, kind: atkKindTenant, h: h, control: control,
			fn: func() (string, error) { return "", fn() }})
	}
	// G018 (oracle r2): create is a guarded CreateGroup as expectation id (an authorized control); the new Group gets
	// fixture key key and the creator's trigger-written owner binding the key rb:<actor>:<key>.
	create := func(t *testing.T, rctx context.Context, tc *TenantContext, step, id, key, actor string) {
		t.Helper()
		var gid tenancy.GroupID
		x.run(t, atkRun{step: step, id: id, target: key, kind: atkKindNone, h: h, control: true,
			fn: func() (string, error) {
				out, err := svc.CreateGroup(rctx, tc, "Attack "+key+" "+e.canaries[0], "attack-"+strings.ToLower(key))
				gid = out.Group.ID
				return "", err
			}})
		x.grp[key], x.label[gid.String()] = gid, key
		x.label[u[actor].id.String()+"|"+gid.String()] = "rb:" + actor + ":" + key
	}
	// sqlWrite runs one raw runtime INSERT or DELETE in a savepoint of the rolled-back shape transaction as expectation
	// id: the affected row count, or "sqlstate <code>" for a refusal (AMB-25).
	sqlWrite := func(t *testing.T, tx pgx.Tx, step, id, target, sql string, args ...any) {
		t.Helper()
		x.o.expect(t, step, id, target)
		st, n := rlsStmt(t, tx, sql, args...)
		r := atkNew()
		r.affected = fmt.Sprint(n)
		if st != "" {
			r.affected = "sqlstate " + st
		}
		x.o.check(t, step, id, target, r)
	}
	// reader calls app.rls_profile_eligible(profile) in a savepoint of the shape transaction as expectation id.
	reader := func(t *testing.T, tx pgx.Tx, step, id, profile string) {
		t.Helper()
		x.o.expect(t, step, id, profile)
		sp, err := tx.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = sp.Rollback(ctx) }()
		var ok bool
		if err := sp.QueryRow(ctx, `SELECT app.rls_profile_eligible($1::text::uuid)`, p[profile].String()).Scan(&ok); err != nil {
			t.Fatalf("%s/%s: reader: %s", step, id, pgState(err))
		}
		r := atkNew()
		r.reader = fmt.Sprint(ok)
		x.o.check(t, step, id, profile, r)
	}

	// --- A: tenant/subject substitution (AC2) ---
	t.Run("A", func(t *testing.T) {
		// G016 A2: a browser request scope is confined to its session's Organization, so multi's legitimate A and B contexts
		// come from two explicitly bound request scopes (its A and B credentials); the A scope never resolves B.
		rctx, end := beginAs(t, svc, ctx, u["multi"], A)
		rctxB, _ := beginAs(t, svc, ctx, u["multi"], B)
		ctxA := resolve(t, svc, rctx, A, g["g3"])
		ctxB := resolve(t, svc, rctxB, B, zg)
		if _, err := svc.Resolve(rctx, ResolveRequest{OrganizationID: B.String()}); CodeOf(err) != CodeContextDenied {
			t.Fatalf("A scope resolves B: %v, want context_denied", err)
		}

		t.Run("A1 one scope, two tenant contexts, three surfaces", func(t *testing.T) {
			put(t, svc, h, rctx, ctxA, "A1", "A1-01", "g3", "a1G3", "a1-g3", true, false)
			put(t, svc, h, rctxB, ctxB, "A1", "A1-02", "orgB", "a1BOrg", "a1-orgB", true, false)
			sign(t, rctxB, ctxB, "A1", "A1-03", "a1BOrg", true)
			facade(t, rctxB, "A1", "A1-04", "g1", false, func() error { _, err := svc.GetGroup(rctxB, ctxB, g["g1"]); return err })
			sign(t, rctxB, ctxB, "A1", "A1-05", "oG1Admin", false)
			del(t, svc, h, rctxB, ctxB, "A1", "A1-06", "oG1Admin", false)
			putDenied(t, svc, h, rctxB, ctxB, "A1", "A1-07", "g1", g["g1"])
			facade(t, rctx, "A1", "A1-08", "pB", false, func() error { _, err := svc.GetBusinessProfile(rctx, ctxA, p["pB"]); return err })
			sign(t, rctx, ctxA, "A1", "A1-09", "oBOrg", false)
			atkShape(t, rt, u["multi"].id, A, g["g3"], func(tx pgx.Tx) {
				x.visible(t, tx, "A1", "A1-10")
				x.visible(t, tx, "A1", "A1-11")
			})
			atkShape(t, rt, u["multi"].id, B, zg, func(tx pgx.Tx) {
				x.visible(t, tx, "A1", "A1-12")
				x.visible(t, tx, "A1", "A1-13")
				x.o.expect(t, "A1", "A1-14", "oG1Admin")
				st, n := rlsStmt(t, tx, `UPDATE app.stored_objects SET state = 'deleting' WHERE id = $1::text::uuid`, ob["oG1Admin"].String())
				r := atkNew()
				r.affected = fmt.Sprint(n)
				if st != "" {
					r.affected = "sqlstate " + st
				}
				r.row, r.bytes = x.objState(t, "oG1Admin")
				x.o.check(t, "A1", "A1-14", "oG1Admin", r)
			})
		})

		t.Run("A2 binding faults on storage writes", func(t *testing.T) {
			rctx2, _ := beginAs(t, svc, ctx, u["multi"], A) // a second request scope of the same credential
			rctxO, _ := beginAs(t, svc, ctx, u["outsider"], A)
			end() // ctxA's request scope ends (B4)
			putDenied(t, svc, h, rctx, ctxA, "A2", "A2-01", "g3", g["g3"])
			putDenied(t, svc, h, rctx2, ctxA, "A2", "A2-02", "g3", g["g3"])
			putDenied(t, svc, h, rctxO, ctxA, "A2", "A2-03", "g3", g["g3"])
			del(t, svc, h, rctx2, ctxA, "A2", "A2-04", "a1G3", false)
		})
	})

	// --- B: shared-profile escalation (AC3) ---
	t.Run("B", func(t *testing.T) {
		t.Run("B1 the shared profile gives no path to the sibling Group", func(t *testing.T) {
			rctx, end := beginAs(t, svc, ctx, u["gAdmin"], A)
			defer end()
			tc := resolve(t, svc, rctx, A, zg)
			facade(t, rctx, "B1", "B1-01", "p1", true, func() error { _, err := svc.GetBusinessProfile(rctx, tc, p["p1"]); return err })
			facade(t, rctx, "B1", "B1-02", "p1", true, func() error {
				_, err := svc.UpdateBusinessProfile(rctx, tc, p["p1"], tenancy.BusinessContext{DisplayName: "Attack p1 Updated " + e.canaries[2]})
				return err
			})
			facade(t, rctx, "B1", "B1-03", "g2", false, func() error { _, err := svc.GetGroup(rctx, tc, g["g2"]); return err })
			facade(t, rctx, "B1", "B1-04", "g2", false, func() error { _, err := svc.RenameGroup(rctx, tc, g["g2"], "Attack Renamed"); return err })
			sign(t, rctx, tc, "B1", "B1-05", "oG2Dual", false)
			del(t, svc, h, rctx, tc, "B1", "B1-06", "oG2Dual", false)
			putDenied(t, svc, h, rctx, tc, "B1", "B1-07", "g2", g["g2"])
			atkShape(t, rt, u["gAdmin"].id, A, zg, func(tx pgx.Tx) {
				for _, id := range []string{"B1-08", "B1-09", "B1-10", "B1-11"} {
					x.visible(t, tx, "B1", id)
				}
			})
		})

		// G018 (oracle r2): the new Group is no profile path, and the creator's owner binding cannot be forged.
		t.Run("B2 a new Group gives no path to a profile", func(t *testing.T) {
			rctx, end := beginAs(t, svc, ctx, u["gMember"], A)
			defer end()
			tc := resolve(t, svc, rctx, A, zg)
			create(t, rctx, tc, "B2", "B2-01", "gB2", "gMember")
			facade(t, rctx, "B2", "B2-02", "gB2", true, func() error { _, err := svc.GetGroupOnboarding(rctx, tc, g["gB2"]); return err })
			for _, c := range []struct{ id, profile string }{{"B2-03", "p1"}, {"B2-04", "p2"}, {"B2-05", "p3"}, {"B2-06", "pB"}} {
				facade(t, rctx, "B2", c.id, "gB2", false, func() error {
					_, err := svc.CompleteGroupOnboardingWithProfile(rctx, tc, g["gB2"], p[c.profile])
					return err
				})
			}
			atkShape(t, rt, u["gMember"].id, A, zg, func(tx pgx.Tx) {
				sqlWrite(t, tx, "B2", "B2-07", "gB2", atkInsertLink, A.String(), g["gB2"].String(), p["p1"].String())
				sqlWrite(t, tx, "B2", "B2-08", "g2", atkInsertOwner, A.String(), u["gMember"].id.String(), g["g2"].String())
				sqlWrite(t, tx, "B2", "B2-09", "gB2", atkInsertOwner, A.String(), u["gMember2"].id.String(), g["gB2"].String())
				x.visible(t, tx, "B2", "B2-10")
				x.visible(t, tx, "B2", "B2-11")
			})
		})
	})

	// --- C: uploader privilege bleed (AC4) ---
	t.Run("C", func(t *testing.T) {
		t.Run("C1 live uploads into the own-versus-other delete decision", func(t *testing.T) {
			scope := func(actor string) (context.Context, *TenantContext) {
				rctx, _ := beginAs(t, svc, ctx, u[actor], A)
				return rctx, resolve(t, svc, rctx, A, g["g1"]) // oracle AMB-13: selected g1
			}
			mctx, mtc := scope("gMember")
			m2ctx, m2tc := scope("gMember2")
			actx, atc := scope("gAdmin")
			put(t, svc, h, mctx, mtc, "C1", "C1-01", "g1", "c1Member", "c1-gMember", true, false)
			put(t, svc, h, m2ctx, m2tc, "C1", "C1-02", "g1", "c1Member2", "c1-gMember2", true, false)
			del(t, svc, h, mctx, mtc, "C1", "C1-03", "c1Member", true)
			del(t, svc, h, mctx, mtc, "C1", "C1-04", "c1Member2", false)
			del(t, svc, h, actx, atc, "C1", "C1-05", "c1Member2", true)
		})

		t.Run("C2 a retained selected-Group context after the Organization membership is removed", func(t *testing.T) {
			rctx, end := beginAs(t, svc, ctx, u["gMember3"], A)
			defer end()
			tc := resolve(t, svc, rctx, A, g["g1"])
			put(t, svc, h, rctx, tc, "C2", "C2-01", "g1", "c2Member3", "c2", true, false)
			x.event(t, "C2", "C2-02", "rb:gMember3:orgA", e.model.RemoveOrganizationMembership(ctx, A, u["gMember3"].id))
			del(t, svc, h, rctx, tc, "C2", "C2-03", "c2Member3", false)
			sign(t, rctx, tc, "C2", "C2-04", "c2Member3", false)
			atkShape(t, rt, u["gMember3"].id, A, g["g1"], func(tx pgx.Tx) {
				for i := 5; i <= 14; i++ {
					x.visible(t, tx, "C2", fmt.Sprintf("C2-%02d", i))
				}
			})
		})

		// G018 (oracle r2): the C2 shape against onboarding completion and the eligibility reader (AMB-26).
		t.Run("C3 a retained context reaches neither completion nor the eligibility reader", func(t *testing.T) {
			rctx, end := beginAs(t, svc, ctx, u["gOwner"], A)
			defer end()
			tc := resolve(t, svc, rctx, A, zg)
			create(t, rctx, tc, "C3", "C3-01", "gC3", "gOwner")
			facade(t, rctx, "C3", "C3-02", "gC3", true, func() error { _, err := svc.GetGroupOnboarding(rctx, tc, g["gC3"]); return err })
			atkShape(t, rt, u["gOwner"].id, A, g["g1"], func(tx pgx.Tx) { reader(t, tx, "C3", "C3-03", "p1") })
			atkShape(t, rt, u["gOwner"].id, A, zg, func(tx pgx.Tx) {
				sqlWrite(t, tx, "C3", "C3-04", "gC3", atkInsertLink, A.String(), g["gC3"].String(), p["p1"].String())
				sqlWrite(t, tx, "C3", "C3-05", "l:g1-p1", `DELETE FROM app.group_business_profiles WHERE group_id = $1::text::uuid AND business_profile_id = $2::text::uuid`,
					g["g1"].String(), p["p1"].String())
			})
			x.event(t, "C3", "C3-06", "rb:gOwner:orgA", e.model.RemoveOrganizationMembership(ctx, A, u["gOwner"].id))
			facade(t, rctx, "C3", "C3-07", "gC3", false, func() error {
				_, err := svc.CompleteGroupOnboardingWithProfile(rctx, tc, g["gC3"], p["p1"])
				return err
			})
			facade(t, rctx, "C3", "C3-08", "gC3", false, func() error {
				_, err := svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, g["gC3"], tenancy.BusinessContext{DisplayName: "Attack C3 " + e.canaries[2]})
				return err
			})
			atkShape(t, rt, u["gOwner"].id, A, g["g1"], func(tx pgx.Tx) {
				reader(t, tx, "C3", "C3-09", "p1")
				x.visible(t, tx, "C3", "C3-11")
			})
			atkShape(t, rt, u["gOwner"].id, A, zg, func(tx pgx.Tx) {
				sqlWrite(t, tx, "C3", "C3-10", "gC3", atkInsertLink, A.String(), g["gC3"].String(), p["p1"].String())
			})
		})
	})

	// --- D: current facts across reuse (AC5) ---
	t.Run("D", func(t *testing.T) {
		t.Run("D1 one retained scope observes profile and storage authority together", func(t *testing.T) {
			rctx, end := beginAs(t, svc, ctx, u["dual"], A)
			defer end()
			ctxO := resolve(t, svc, rctx, A, zg)
			ctxG := resolve(t, svc, rctx, A, g["g1"])
			x.event(t, "D1", "D1-01", "l:g2-p1", e.model.UnlinkGroupBusinessProfile(ctx, A, g["g2"], p["p1"]))
			facade(t, rctx, "D1", "D1-02", "p1", true, func() error {
				_, err := svc.UpdateBusinessProfile(rctx, ctxO, p["p1"], tenancy.BusinessContext{DisplayName: "Attack p1 Through g1 " + e.canaries[0]})
				return err
			})
			x.run(t, atkRun{step: "D1", id: "D1-03", target: "oG2Dual", kind: atkKindObject, h: h, control: true,
				fn: func() (string, error) { _, err := svc.ObjectMetadata(rctx, ctxO, ob["oG2Dual"]); return "", err }})
			sign(t, rctx, ctxO, "D1", "D1-04", "oG2Dual", true)
			_, cerr := e.model.ChangeGroupRole(ctx, A, g["g1"], u["dual"].id, tenancy.GroupMember)
			x.event(t, "D1", "D1-05", "rb:dual:g1", cerr)
			facade(t, rctx, "D1", "D1-06", "p1", false, func() error {
				_, err := svc.UpdateBusinessProfile(rctx, ctxG, p["p1"], tenancy.BusinessContext{DisplayName: "Attack Denied"})
				return err
			})
			facade(t, rctx, "D1", "D1-07", "p1", true, func() error { _, err := svc.GetBusinessProfile(rctx, ctxG, p["p1"]); return err })
			del(t, svc, h, rctx, ctxG, "D1", "D1-08", "oG1Admin", false)
			del(t, svc, h, rctx, ctxG, "D1", "D1-09", "oG1Dual", true)
			x.event(t, "D1", "D1-10", "rb:dual:g1", e.model.RemoveGroupMembership(ctx, A, g["g1"], u["dual"].id))
			facade(t, rctx, "D1", "D1-11", "g1", false, func() error { _, err := svc.GetGroup(rctx, ctxG, g["g1"]); return err })
			facade(t, rctx, "D1", "D1-12", "p1", false, func() error { _, err := svc.GetBusinessProfile(rctx, ctxG, p["p1"]); return err })
			sign(t, rctx, ctxG, "D1", "D1-13", "oG1Admin", false)
			del(t, svc, h, rctx, ctxG, "D1", "D1-14", "oG1Admin", false)
			putDenied(t, svc, h, rctx, ctxG, "D1", "D1-15", "g1", g["g1"])
			facade(t, rctx, "D1", "D1-16", "g1", false, func() error { _, err := svc.GetGroup(rctx, ctxO, g["g1"]); return err })
			facade(t, rctx, "D1", "D1-17", "p1", false, func() error { _, err := svc.GetBusinessProfile(rctx, ctxO, p["p1"]); return err })
		})

		t.Run("D2 one backend serves storage and facade operations", func(t *testing.T) {
			pool := e.singlePool(t)
			if err := e.rlsAttest(pool); err != nil {
				t.Fatal(err)
			}
			hh := &hookDB{pool: pool}
			one := e.objService(hh, e.spy, e.ocfg)
			rctx, end := beginAs(t, one, ctx, u["multi"], A)
			defer end()
			ctxA3 := resolve(t, one, rctx, A, g["g3"])
			rctxB, endB := beginAs(t, one, ctx, u["multi"], B) // G016 A2: B is a separate bound request scope
			defer endB()
			ctxB := resolve(t, one, rctxB, B, zg)
			ref, _ := atkBackend(t, pool)
			backend := func(o *atkObs) {
				pid, s := atkBackend(t, pool)
				o.backend, o.settings = "same as D2-01", "empty"
				if pid != ref {
					o.backend = "different"
				}
				if s != "|||" {
					o.settings = "not empty"
				}
			}
			before := func() string { // settings when the operation begins
				_, s := atkBackend(t, pool)
				if s != "|||" {
					return "not empty"
				}
				return "empty"
			}
			fo := atkStepObject(t, "d2G3", "D2")
			req, data := x.putReq(x.grp[fo.Group], "d2")
			x.run(t, atkRun{step: "D2", id: "D2-01", target: "g3", kind: atkKindPut, h: hh, putKey: "d2G3", putOrg: x.org[fo.Org],
				putData: func() []byte { return data }, control: true,
				fn: func() (string, error) { _, err := one.PutObject(rctx, ctxA3, req); return "", err },
				extra: func(o *atkObs, _ error) {
					backend(o)
					if o.backend == "same as D2-01" { // the one pooled connection served the Put: the reference PID
						o.backend = "observed (reference PID)"
					}
				}})
			pr := atkNew()
			backend(&pr)
			x.o.check(t, "D2", "D2-02", "backend", pr)
			s0 := before()
			x.run(t, atkRun{step: "D2", id: "D2-03", target: "orgB", kind: atkKindNone, h: hh, control: true,
				fn: func() (string, error) {
					org, err := one.GetOrganization(rctxB, ctxB)
					if err == nil && org.ID != B {
						return "", fmt.Errorf("access: inexact organization")
					}
					return "", err
				},
				extra: func(o *atkObs, _ error) {
					backend(o)
					if s0 != "empty" {
						o.settings = s0
					}
				}})
			hh.mu.Lock()
			hh.failOn = "O" // the organization facts statement of the Put's T1
			hh.mu.Unlock()
			x.run(t, atkRun{step: "D2", id: "D2-04", target: "orgB", kind: atkKindPut, h: hh,
				fn: func() (string, error) {
					req, _ := x.putReq(zg, "d2-fail")
					_, err := one.PutObject(rctxB, ctxB, req)
					hh.mu.Lock()
					hh.failOn = ""
					hh.mu.Unlock()
					return "", err
				},
				extra: func(o *atkObs, _ error) { backend(o); o.settings = "" }})
			s0 = before()
			x.run(t, atkRun{step: "D2", id: "D2-05", target: "g3", kind: atkKindNone, h: hh, control: true,
				fn: func() (string, error) {
					gr, err := one.GetGroup(rctx, ctxA3, g["g3"])
					if err == nil && gr.ID != g["g3"] {
						return "", fmt.Errorf("access: inexact group")
					}
					return "", err
				},
				extra: func(o *atkObs, _ error) {
					backend(o)
					if s0 != "empty" {
						o.settings = s0
					}
				}})
		})
	})

	// --- E: staged storage and the bearer residual (AC6) ---
	t.Run("E", func(t *testing.T) {
		t.Run("E1 an Organization role downgrade between the landed bytes and the publish", func(t *testing.T) {
			var mu sync.Mutex
			armed := false
			fired := false
			staged, sh := e.stageService(func(op, stage string) {
				mu.Lock()
				fire := armed && op == opObjectPut && stage == stageT2
				if fire {
					armed, fired = false, true
				}
				mu.Unlock()
				if fire {
					_, err := e.model.ChangeOrganizationRole(ctx, A, u["aAdmin"].id, tenancy.OrgMember)
					x.event(t, "E1", "E1-02", "rb:aAdmin:orgA", err)
				}
			})
			stagedPut := func(actor, id, key string) {
				rctx, end := beginAs(t, staged, ctx, u[actor], A)
				defer end()
				tc := resolve(t, staged, rctx, A, zg)
				// The control must succeed; the denied Put is the staged denial (§1: staged effects, not zero I/O).
				put(t, staged, sh, rctx, tc, "E1", id, "orgA", key, "e1-"+actor, id == "E1-01", id == "E1-03")
			}
			// The working control first (oracle AMB-20): the identical staged path, no role change.
			stagedPut("admin2", "E1-01", "e1Admin2Org")
			var ops []string
			e.spy.set(func(op string, _ objectstore.Key) error { mu.Lock(); ops = append(ops, op); mu.Unlock(); return nil })
			mu.Lock()
			armed = true
			mu.Unlock()
			stagedPut("aAdmin", "E1-03", "e1AdminOrg")
			e.spy.set(nil)
			if !fired {
				t.Fatal("E1: the Put did not reach its publish stage")
			}
			if !slices.Equal(ops, []string{"put", "head", "delete", "head"}) {
				t.Errorf("E1: provider calls %v, want put, head (verify), delete, head (compensation)", ops)
			}
			// Never ready, never readable; the control is readable.
			rctx, end := beginAs(t, svc, ctx, u["aOwner"], A)
			defer end()
			tc := resolve(t, svc, rctx, A, zg)
			if id, ok := ob["e1AdminOrg"]; ok {
				if _, err := svc.ObjectMetadata(rctx, tc, id); CodeOf(err) != CodeNotFound {
					t.Errorf("E1: the denied staged object is readable: %v", err)
				}
			}
			if _, err := svc.ObjectMetadata(rctx, tc, ob["e1Admin2Org"]); err != nil {
				t.Errorf("E1: the working control is not readable: %v", err)
			}
		})

		t.Run("E2 an issued URL across the logical and the physical delete", func(t *testing.T) {
			// G060 (Plan r0 A9/A10): g1's admin, not the Organization owner, holds Group-object authority (AMB-13 note).
			rctx, end := beginAs(t, svc, ctx, u["gAdmin"], A)
			defer end()
			tc := resolve(t, svc, rctx, A, zg)
			var url SignedURL
			x.run(t, atkRun{step: "E2", id: "E2-01", target: "oG1E2", kind: atkKindObject, h: h, control: true,
				fn: func() (string, error) {
					var err error
					url, err = svc.SignedGetURL(rctx, tc, SignedGetURLRequest{Object: ob["oG1E2"], DefaultTTL: true})
					return "", err
				},
				extra: func(o *atkObs, err error) {
					if err == nil {
						o.url = atkFetch(t, url, x.data["oG1E2"])
					}
				}})
			paused := false
			staged, sh := e.stageService(func(op, stage string) {
				if op != opObjectDelete || stage != stageProvider || paused {
					return
				}
				paused = true // D1 committed (deleting); the provider bytes are not yet removed
				r := atkNew()
				r.url = atkFetch(t, url, x.data["oG1E2"])
				r.row, r.bytes = x.objState(t, "oG1E2")
				x.o.check(t, "E2", "E2-02", "oG1E2", r)
				sign(t, rctx, tc, "E2", "E2-03", "oG1E2", false)
				r = atkNew()
				r.row, r.bytes = x.objState(t, "oG1E2")
				x.o.check(t, "E2", "E2-04", "oG1E2", r)
			})
			drctx, dend := beginAs(t, staged, ctx, u["gAdmin"], A)
			defer dend()
			dtc := resolve(t, staged, drctx, A, zg)
			x.run(t, atkRun{step: "E2", id: "E2-05", target: "oG1E2", kind: atkKindObject, h: sh, staged: true,
				fn: func() (string, error) {
					res, err := staged.DeleteObject(drctx, dtc, ob["oG1E2"])
					return string(res.Outcome), err
				}})
			if !paused {
				t.Fatal("E2: the Delete did not reach its provider stage")
			}
			if !time.Now().Before(url.ExpiresAt) {
				t.Fatal("E2: the URL expired before the post-delete fetch; it would not show the physical removal")
			}
			r := atkNew()
			r.url = atkFetch(t, url, x.data["oG1E2"])
			r.row, r.bytes = x.objState(t, "oG1E2")
			x.o.check(t, "E2", "E2-06", "oG1E2", r)
		})

		t.Run("E3 a Group removal between the byte removal and the tombstone", func(t *testing.T) {
			if row, b := x.objState(t, "oG1MemberE3"); row != atkRowReady || b != atkBytesPresent {
				t.Fatalf("E3: the pre-delete object is %s/%s, want a live ready object (not C1's tombstone)", row, b)
			}
			fired := false
			var callsAtRevocation int64
			staged, sh := e.stageService(func(op, stage string) {
				if op != opObjectDelete || stage != stageT2 || fired {
					return
				}
				fired = true
				x.event(t, "E3", "E3-01", "rb:gMember:g1", e.model.RemoveGroupMembership(ctx, A, g["g1"], u["gMember"].id))
				callsAtRevocation = e.spy.calls.Load() // D2 delete and D3 Head are done; nothing may follow
			})
			rctx, end := beginAs(t, staged, ctx, u["gMember"], A)
			defer end()
			tc := resolve(t, staged, rctx, A, g["g1"]) // oracle AMB-13: selected g1
			x.run(t, atkRun{step: "E3", id: "E3-02", target: "oG1MemberE3", kind: atkKindObject, h: sh, staged: true,
				fn: func() (string, error) {
					res, err := staged.DeleteObject(rctx, tc, ob["oG1MemberE3"])
					// Codex attempt-1 ruling 1: no provider call follows the revocation.
					if fired && e.spy.calls.Load() != callsAtRevocation {
						t.Errorf("E3-02: %d provider calls after the revocation, want 0", e.spy.calls.Load()-callsAtRevocation)
					}
					return string(res.Outcome), err
				}})
			if !fired {
				t.Fatal("E3: the Delete did not reach its tombstone stage")
			}
			arctx, aend := beginAs(t, svc, ctx, u["gAdmin"], A) // G060: g1's admin repeats the delete (AMB-13 note)
			defer aend()
			atc := resolve(t, svc, arctx, A, zg)
			del(t, svc, h, arctx, atc, "E3", "E3-03", "oG1MemberE3", false)
		})
	})

	// --- F: direct binding writes, invitation tampering and consumer forgery (G019, oracle r3; AMB-27..29) ---
	t.Run("F", func(t *testing.T) {
		// sqlGuard is sqlWrite with the server's constraint name of a guard refusal appended (AMB-28); a SELECT counts rows.
		sqlGuard := func(t *testing.T, tx pgx.Tx, step, id, target, sql string, args ...any) {
			t.Helper()
			x.o.expect(t, step, id, target)
			st, cons, n := rlsStmtC(t, tx, sql, args...)
			r := atkNew()
			r.affected = fmt.Sprint(n)
			if st != "" {
				r.affected = "sqlstate " + st
				if cons != "" {
					r.affected += " " + cons
				}
			}
			x.o.check(t, step, id, target, r)
		}
		// op is a guarded facade or consumer operation as expectation id (Code, and the consumption outcome as Result).
		op := func(t *testing.T, step, id, target string, control bool, fn func() (string, error)) {
			t.Helper()
			x.run(t, atkRun{step: step, id: id, target: target, kind: atkKindNone, control: control, fn: fn})
		}
		ids := func(actor string) string { return u[actor].id.String() }

		t.Run("F1 direct role_bindings writes", func(t *testing.T) {
			rctx, end := beginAs(t, svc, ctx, u["aMember"], A)
			defer end()
			tc := resolve(t, svc, rctx, A, zg)
			create(t, rctx, tc, "F1", "F1-01", "gF", "aMember")
			before := x.membershipDigest(t)
			atkShape(t, rt, u["aOwner"].id, A, zg, func(tx pgx.Tx) {
				sqlGuard(t, tx, "F1", "F1-02", "rb:admin2:orgA", atkUpdateRole, ids("admin2"), A.String(), "org_member")
				sqlGuard(t, tx, "F1", "F1-03", "rb:outsider:orgA", atkInsertOrgBinding, A.String(), ids("outsider"))
				sqlGuard(t, tx, "F1", "F1-05", "rb:aMember:gF", atkUpdateRole, ids("aMember"), g["gF"].String(), "group_admin")
				sqlGuard(t, tx, "F1", "F1-06", "rb:aMember:orgA", atkUpdateRole, ids("aMember"), A.String(), "org_owner")
				sqlGuard(t, tx, "F1", "F1-07", "rb:aOwner:orgA", atkUpdateRole, ids("aOwner"), A.String(), "org_admin")
				sqlGuard(t, tx, "F1", "F1-08", "rb:aOwner:orgA", atkDeleteBinding, ids("aOwner"), A.String())
			})
			atkShape(t, rt, u["admin2"].id, A, zg, func(tx pgx.Tx) {
				sqlGuard(t, tx, "F1", "F1-04", "rb:aOwner:orgA", atkUpdateRole, ids("aOwner"), A.String(), "org_admin")
				sqlGuard(t, tx, "F1", "F1-13", "rb:admin2:orgA", atkUpdateRole, ids("admin2"), A.String(), "org_owner")
			})
			atkShape(t, rt, u["aMember"].id, A, zg, func(tx pgx.Tx) {
				sqlGuard(t, tx, "F1", "F1-09", "rb:aMember:orgA", atkDeleteBinding, ids("aMember"), A.String())
			})
			atkShape(t, rt, u["aMember"].id, A, g["gF"], func(tx pgx.Tx) {
				sqlGuard(t, tx, "F1", "F1-10", "rb:aMember:orgA", atkDeleteBinding, ids("aMember"), A.String())
			})
			atkShape(t, rt, u["dual"].id, A, g["g1"], func(tx pgx.Tx) {
				sqlGuard(t, tx, "F1", "F1-11", "rb:dual:g2", atkUpdateRole, ids("dual"), g["g2"].String(), "group_member")
			})
			atkShape(t, rt, u["multi"].id, B, zg, func(tx pgx.Tx) {
				sqlGuard(t, tx, "F1", "F1-12", "rb:aMember:orgA", atkUpdateRole, ids("aMember"), A.String(), "org_admin")
			})
			if after := x.membershipDigest(t); after != before {
				t.Errorf("F1: a refused direct binding write changed membership or invitation state")
			}
		})

		t.Run("F2 invitation tampering, unreadable digest and consumer forgery", func(t *testing.T) {
			octx, oend := beginAs(t, svc, ctx, u["aOwner"], A)
			defer oend()
			otc := resolve(t, svc, octx, A, zg)
			mctx, mend := beginAs(t, svc, ctx, u["aMember"], A)
			defer mend()
			mtc := resolve(t, svc, mctx, A, zg)
			a2ctx, a2end := beginAs(t, svc, ctx, u["admin2"], A)
			defer a2end()
			a2tc := resolve(t, svc, a2ctx, A, zg)
			g2ctx, g2end := beginAs(t, svc, ctx, u["gMember2"], A)
			defer g2end()
			g2tc := resolve(t, svc, g2ctx, A, zg)
			dctx, dend := beginAs(t, svc, ctx, u["dual"], A)
			defer dend()
			dtc := resolve(t, svc, dctx, A, zg)
			gF := g["gF"]
			addr, addr2 := memAddr("attack-f2"), memAddr("attack-f2-owner-role")
			var invOrg, invGrp IssuedInvitation
			op(t, "F2", "F2-01", "invOrg", true, func() (string, error) {
				var err error
				invOrg, err = svc.IssueInvitation(octx, otc, Scope{Kind: tenancy.ScopeOrganization}, string(tenancy.OrgMember), Recipient{Address: &addr}, 0)
				return "", err
			})
			x.label[invOrg.Invitation.ID.String()] = "invOrg"
			op(t, "F2", "F2-02", "invGrp", true, func() (string, error) {
				var err error
				invGrp, err = svc.IssueInvitation(mctx, mtc, Scope{Kind: tenancy.ScopeGroup, Group: &gF}, string(tenancy.GroupMember), Recipient{UserID: &u["gMember2"].id}, 0)
				return "", err
			})
			x.label[invGrp.Invitation.ID.String()] = "invGrp"
			invO, invG := invOrg.Invitation.ID.String(), invGrp.Invitation.ID.String()
			op(t, "F2", "F2-03", "invOwnerRole", false, func() (string, error) {
				_, err := svc.IssueInvitation(a2ctx, a2tc, Scope{Kind: tenancy.ScopeOrganization}, string(tenancy.OrgOwner), Recipient{Address: &addr2}, 0)
				return "", err
			})
			before := x.membershipDigest(t)
			atkShape(t, rt, u["aOwner"].id, A, zg, func(tx pgx.Tx) {
				sqlGuard(t, tx, "F2", "F2-04", "invOrg", atkInvDigest, invO)
				sqlGuard(t, tx, "F2", "F2-07", "invOrg", atkInvRole, invO)
				sqlGuard(t, tx, "F2", "F2-08", "invOrg", atkInvExpiry, invO)
				sqlGuard(t, tx, "F2", "F2-09", "invOrg", atkInvDigestSet, invO)
				sqlGuard(t, tx, "F2", "F2-10", "invOrg", atkInvRevoke, invO)
			})
			atkShape(t, rt, u["gMember2"].id, A, zg, func(tx pgx.Tx) {
				sqlGuard(t, tx, "F2", "F2-05", "invGrp", atkInvDigest, invG)
				sqlGuard(t, tx, "F2", "F2-06", "invGrp", atkInvConsume, invG, ids("gMember2"))
				sqlGuard(t, tx, "F2", "F2-11", "invGrp", atkInvRevoke, invG)
			})
			atkShape(t, rt, u["aMember"].id, A, zg, func(tx pgx.Tx) {
				sqlGuard(t, tx, "F2", "F2-12", "invOrg", atkOrgClaim, invO, []byte(invOrg.Token), e.fake.Issuer, "user_01ATTACKF2CLAIM", addr)
			})
			// The wrong token is another invitation's token (well-formed, never this row's; Plan MEM-11 "token of another
			// invitation"); the non-recipient presents the right one.
			op(t, "F2", "F2-13", "invGrp", false, func() (string, error) {
				_, err := svc.AcceptGroupInvitation(g2ctx, g2tc, invGrp.Invitation.ID, invOrg.Token)
				return "", err
			})
			op(t, "F2", "F2-14", "invGrp", false, func() (string, error) {
				_, err := svc.AcceptGroupInvitation(dctx, dtc, invGrp.Invitation.ID, invGrp.Token)
				return "", err
			})
			// The synthetic Organization consumer (the owner service with attested evidence, MEM-10's shape).
			consumer := e.memOwnerService(t)
			sub1, _ := e.memPrincipal(t, "ATKF2A")
			sub2, _ := e.memPrincipal(t, "ATKF2B")
			consume := func(ev identity.VerifiedAddress, id tenancy.InvitationID, token string) func() (string, error) {
				return func() (string, error) {
					out, err := consumer.ConsumeOrganizationInvitation(ctx, ev, id, token)
					return out.Outcome, err
				}
			}
			op(t, "F2", "F2-15", "invOrg", false, consume(identity.VerifiedAddress{}, invOrg.Invitation.ID, invOrg.Token))
			op(t, "F2", "F2-16", "invOrg", false, consume(e.memAttest(t, sub1, addr), invOrg.Invitation.ID, invGrp.Token))
			op(t, "F2", "F2-17", "invOrg", false, consume(e.memAttest(t, sub1, memAddr("attack-f2-other")), invOrg.Invitation.ID, invOrg.Token))
			op(t, "F2", "F2-18", "invGrp", false, consume(e.memAttest(t, sub1, addr), invGrp.Invitation.ID, invGrp.Token))
			if after := x.membershipDigest(t); after != before {
				t.Errorf("F2: a refused invitation attack changed membership or invitation state")
			}
			accept := func() (string, error) {
				out, err := svc.AcceptGroupInvitation(g2ctx, g2tc, invGrp.Invitation.ID, invGrp.Token)
				return out.Outcome, err
			}
			op(t, "F2", "F2-19", "invGrp", true, accept)
			x.label[ids("gMember2")+"|"+gF.String()] = "rb:gMember2:gF"
			op(t, "F2", "F2-20", "invGrp", true, accept)
			x.event(t, "F2", "F2-21", "rb:gMember2:gF", e.model.RemoveGroupMembership(ctx, A, gF, u["gMember2"].id))
			op(t, "F2", "F2-22", "invGrp", true, accept)
			atkShape(t, rt, u["gMember2"].id, A, zg, func(tx pgx.Tx) { x.visible(t, tx, "F2", "F2-23") })
			op(t, "F2", "F2-24", "invOrg", true, consume(e.memAttest(t, sub1, addr), invOrg.Invitation.ID, invOrg.Token))
			op(t, "F2", "F2-25", "invOrg", false, consume(e.memAttest(t, sub2, addr), invOrg.Invitation.ID, invOrg.Token))
		})
	})

	// --- G: generic labels (r5, G020): cross-tenant and cross-Group label reads, counts, joins and refused writes ---
	t.Run("G", func(t *testing.T) {
		// sqlLabel is a raw runtime statement in a savepoint of the rolled-back shape transaction as expectation id: the
		// row count (a SELECT counts returned rows), or "sqlstate <code>[ <constraint>]" for a refusal.
		sqlLabel := func(t *testing.T, tx pgx.Tx, id, target, sql string, args ...any) {
			t.Helper()
			x.o.expect(t, "G1", id, target)
			st, cons, n := rlsStmtC(t, tx, sql, args...)
			r := atkNew()
			r.affected = fmt.Sprint(n)
			if st != "" {
				r.affected = "sqlstate " + st
				if cons != "" {
					r.affected += " " + cons
				}
			}
			x.o.check(t, "G1", id, target, r)
		}
		const insertAssignment = `INSERT INTO app.label_assignments (id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id, target_group_id)
VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid, 'organization', $2::text::uuid, 'group', $4::text::uuid, $4::text::uuid)`

		t.Run("G1 label rows of other tenants and other Groups stay invisible with the subject's own rows present", func(t *testing.T) {
			orgs := []string{}
			for _, o := range x.orgs() {
				orgs = append(orgs, o.String())
			}
			before := e.labelState(t, orgs...)
			atkShape(t, rt, u["aOwner"].id, A, zg, func(tx pgx.Tx) {
				x.visible(t, tx, "G1", "G1-01")
				x.visible(t, tx, "G1", "G1-02")
				x.visible(t, tx, "G1", "G1-03")
				sqlLabel(t, tx, "G1-04", "", `SELECT 1 FROM app.label_definitions WHERE owner_scope_type = 'group'`)
			})
			atkShape(t, rt, u["aOwner"].id, A, g["g1"], func(tx pgx.Tx) {
				x.visible(t, tx, "G1", "G1-05")
			})
			atkShape(t, rt, u["admin2"].id, A, zg, func(tx pgx.Tx) {
				x.visible(t, tx, "G1", "G1-06")
			})
			// gAdmin, not gOwner: C3-06 removed gOwner's orgA membership (AMB-24: gOwner is exercised only before C3), so its
			// raw context reads nothing; gAdmin holds org_member and group_admin of g1 through G.
			atkShape(t, rt, u["gAdmin"].id, A, zg, func(tx pgx.Tx) {
				x.visible(t, tx, "G1", "G1-07")
				x.visible(t, tx, "G1", "G1-08")
				sqlLabel(t, tx, "G1-09", "", `SELECT 1 FROM app.label_definitions d JOIN app.label_assignments a ON a.label_id = d.id`)
				sqlLabel(t, tx, "G1-10", "as:lA:g1", `DELETE FROM app.label_assignments WHERE target_group_id = $1::text::uuid`, g["g1"].String())
			})
			atkShape(t, rt, u["gMember2"].id, A, zg, func(tx pgx.Tx) {
				x.visible(t, tx, "G1", "G1-11")
				sqlLabel(t, tx, "G1-12", "lG1tm", `UPDATE app.label_definitions SET name = 'Attack G1 rename' WHERE id = $1::text::uuid`, x.lab["lG1tm"])
			})
			atkShape(t, rt, u["multi"].id, A, g["g3"], func(tx pgx.Tx) {
				x.visible(t, tx, "G1", "G1-13")
				x.visible(t, tx, "G1", "G1-14")
				sqlLabel(t, tx, "G1-15", "lA", `DELETE FROM app.label_definitions WHERE id = $1::text::uuid`, x.lab["lA"])
			})
			atkShape(t, rt, u["multi"].id, B, zg, func(tx pgx.Tx) {
				x.visible(t, tx, "G1", "G1-16")
				x.visible(t, tx, "G1", "G1-17")
				sqlLabel(t, tx, "G1-18", "", `SELECT 1 FROM app.label_definitions WHERE organization_id = $1::text::uuid`, A.String())
				sqlLabel(t, tx, "G1-19", "", `SELECT 1 FROM app.label_assignments WHERE label_id = $1::text::uuid`, x.lab["lA"])
				sqlLabel(t, tx, "G1-20", "g3", insertAssignment, rlsUUID(t), B.String(), x.lab["lB"], g["g3"].String())
				sqlLabel(t, tx, "G1-21", "g1", insertAssignment, rlsUUID(t), A.String(), x.lab["lA"], g["g1"].String())
			})
			if after := e.labelState(t, orgs...); after != before {
				t.Errorf("G1: a refused or rolled-back label attack changed label state")
			}
		})
	})

	t.Run("ISO-CITATIONS every objects: and rls: citation resolves", func(t *testing.T) {
		objIDs, rlsNames := map[string]bool{}, map[string]bool{}
		for _, c := range objectsExpectedCases() {
			objIDs[c.ID] = true
			if i := strings.Index(c.ID, "/"); i > 0 { // a read row expands into ID/metadata, ID/get and ID/sign
				objIDs[c.ID[:i]] = true
			}
		}
		for _, c := range rlsExpectedCases() {
			rlsNames[c.Name] = true
		}
		insNames := map[string]bool{} // G018: oracle r2 cites the rls oracle r3 insert rows as rlsins:<name>
		for _, c := range rlsExpectedInserts() {
			insNames[c.Name] = true
		}
		mutNames := map[string]bool{} // G019: oracle r3 cites the rls oracle r4 mutation rows as rlsmut:<name>
		for _, c := range rlsExpectedMutations() {
			mutNames[c.Name] = true
		}
		n := 0
		for _, c := range atkCitations() {
			switch {
			case strings.HasPrefix(c, "rlsmut:"):
				n++
				if !mutNames[strings.TrimPrefix(c, "rlsmut:")] {
					t.Errorf("citation %q does not resolve in rlsExpectedMutations()", c)
				}
			case strings.HasPrefix(c, "objects:"):
				n++
				if !objIDs[strings.TrimPrefix(c, "objects:")] {
					t.Errorf("citation %s does not resolve in objectsExpectedCases()", c)
				}
			case strings.HasPrefix(c, "rls:"):
				n++
				if !rlsNames[strings.TrimPrefix(c, "rls:")] {
					t.Errorf("citation %q does not resolve in rlsExpectedCases()", c)
				}
			case strings.HasPrefix(c, "rlsins:"):
				n++
				if !insNames[strings.TrimPrefix(c, "rlsins:")] {
					t.Errorf("citation %q does not resolve in rlsExpectedInserts()", c)
				}
			}
		}
		if n == 0 {
			t.Error("no objects:/rls: citation found in the oracle")
		}
	})

	o.finish(t)
}
