//go:build integration

// B-POLICY-ORDERING (G007 Plan r1 §6.4, AC8): TestAccessOrdering, the authorization-to-write ordering of every guarded
// write against each conflicting accepted writer, deterministically and without sleeps: the hookDB pauses a guarded
// operation or a tenancy writer at a statement or before COMMIT, a pgx.QueryTracer on a dedicated identity pool pauses
// an identity writer inside its transaction, and lock waits are proven by waitWaiters (pg_locks joined to
// pg_stat_activity). Orderings:
//
//   - A (guard first): the guard pauses after its facts (before its write), the writer blocks on a lock the guard holds,
//     the guard commits, then the writer commits. Both changes persist and the guard finished first.
//   - B (writer first): the guard pauses after BEGIN and before its first statement, the writer commits, then the guard
//     denies with its pinned code and its target is unchanged.
//   - C (writer in flight): the writer pauses before COMMIT holding its row lock, the guard blocks, the writer commits,
//     the guard re-checks and denies; its target is unchanged.
//
// O-15 records the one possible wait cycle (sign-in rotation with a changed display name; Plan r1 §4.4, Q15) and
// E-PG-3a/b characterize the lock queue (PG-001 H-4). G019 adds O-MEM-1..3(d) and O-MEM-DUP (Plan r1 §6.2; r2 addendum §4:
// a facade cycle ends in exactly one class-40 victim answered unavailable; r3 C2). Each case prints one "ORDERING " JSON
// line.
//
// Required environment: as TestAccessPolicy (TENANCY_LOG_FILE receives this test's access.decision records).
package access

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/tenancy"
)

// --- synchronization -------------------------------------------------------------------------------------------------

// pause blocks the first caller of hook until release is closed; reached is closed when it blocks.
type pause struct {
	reached, release chan struct{}
	once             sync.Once
}

func newPause() *pause { return &pause{reached: make(chan struct{}), release: make(chan struct{})} }

func (p *pause) hook() {
	p.once.Do(func() {
		close(p.reached)
		<-p.release
	})
}

func (p *pause) wait(t *testing.T, what string) {
	t.Helper()
	select {
	case <-p.reached:
	case <-time.After(15 * time.Second):
		t.Fatalf("%s never reached its pause", what)
	}
}

var finishSeq atomic.Int64

// actor runs one call in its own goroutine and records its error and finishing order.
type actor struct {
	done chan struct{}
	err  error
	at   int64
}

func start(fn func() error) *actor {
	a := &actor{done: make(chan struct{})}
	go func() {
		defer close(a.done)
		a.err = fn()
		a.at = finishSeq.Add(1)
	}()
	return a
}

func (a *actor) wait(t *testing.T, what string) error {
	t.Helper()
	select {
	case <-a.done:
	case <-time.After(30 * time.Second):
		t.Fatalf("%s did not finish", what)
	}
	return a.err
}

func (a *actor) finished() bool {
	select {
	case <-a.done:
		return true
	default:
		return false
	}
}

type tracerMark struct{}

// pauseTracer pauses, once armed, the first statement whose SQL contains match, after it completed successfully and
// while its transaction stays open.
type pauseTracer struct {
	mu    sync.Mutex
	match string
	p     *pause
}

func (pt *pauseTracer) arm(match string) *pause {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.match, pt.p = match, newPause()
	return pt.p
}

func (pt *pauseTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if pt.p != nil && pt.match != "" && strings.Contains(data.SQL, pt.match) {
		return context.WithValue(ctx, tracerMark{}, pt.p)
	}
	return ctx
}

func (pt *pauseTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	p, ok := ctx.Value(tracerMark{}).(*pause)
	if !ok || data.Err != nil {
		return
	}
	pt.mu.Lock()
	if pt.p == p {
		pt.match = ""
	}
	pt.mu.Unlock()
	p.hook()
}

func ordering(t *testing.T, fields map[string]any) {
	t.Helper()
	line, _ := json.Marshal(fields)
	fmt.Println("ORDERING " + string(line))
}

func isDeadlock(err error) bool {
	var pg *pgconn.PgError
	return errors.As(err, &pg) && pg.Code == "40P01"
}

// --- the ordering test -------------------------------------------------------------------------------------------------

func TestAccessOrdering(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	var zeroG tenancy.GroupID
	org := e.org(t, "Ordering Organization "+e.canaries[0])
	owner := e.newUser(t)
	e.grantOrg(t, org, owner, tenancy.OrgOwner)
	e.bind(t, owner, org) // G016 A2: owner's credential for org (setup only)
	creator := e.newUser(t)
	tracer := &pauseTracer{}
	idTraced := e.identity(t, e.clk, e.runtimePool(t, tracer))
	slugSeq := 0
	newGroup := func(t *testing.T) tenancy.Group {
		slugSeq++
		return e.group(t, org, creator, fmt.Sprintf("Ordering Group %d", slugSeq), fmt.Sprintf("ordering-%d", slugSeq))
	}
	newProfile := func(t *testing.T) tenancy.BusinessProfile {
		slugSeq++
		return e.profile(t, org, fmt.Sprintf("Ordering Profile %d %s", slugSeq, e.canaries[1]))
	}
	// member returns a new subject: org_member (or the given Organization role) plus the given Group roles.
	member := func(t *testing.T, orgRole tenancy.OrganizationRole, groups map[tenancy.GroupID]tenancy.GroupRole) *itUser {
		u := e.newUser(t)
		e.grantOrg(t, org, u, orgRole)
		for g, r := range groups {
			e.grantGroup(t, org, g, u, r)
		}
		e.bind(t, u, org) // G016 A2: u's credential for org, prepared before any ordering window
		return u
	}
	rowOf := func(t *testing.T, rel string, id [16]byte) string {
		var s string
		if err := e.setup.QueryRow(ctx, `SELECT coalesce(md5(r::text), '-') FROM app.`+rel+` r WHERE r.id = $1::uuid`, fmt.Sprintf("%x", id[:])).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	profileRow := func(t *testing.T, p tenancy.BusinessProfileID) func() string {
		return func() string { return rowOf(t, "business_profiles", p) }
	}
	groupRow := func(t *testing.T, g tenancy.GroupID) func() string {
		return func() string { return rowOf(t, "groups", g) }
	}
	orgRow := func(t *testing.T) func() string { return func() string { return rowOf(t, "organizations", org) } }
	// G060: the Organization owner sets the Group-management override through the facade (the cases that need an
	// Organization owner/admin's Group or profile administration enable it and disable it afterwards).
	setOverride := func(t *testing.T, on bool) {
		t.Helper()
		osvc, _ := e.service(e.id)
		octx, _ := beginAs(t, osvc, ctx, owner, org)
		otc := resolve(t, osvc, octx, org, zeroG)
		if v, err := osvc.SetGroupManagementOverride(octx, otc, on); err != nil || v.OverrideEnabled != on {
			t.Fatalf("owner sets the Group-management override to %t: %+v %v", on, v, err)
		}
	}

	// prepared is one guard and its writer, built fresh for each ordering.
	type prepared struct {
		guard  func() error
		writer func() error
		armC   func() *pause // pauses the writer before its COMMIT (nil: single statement, C equals A or B)
		target func() string
		step   string // G018: the statement the guard pauses before in order A (default W)
	}
	type caseDef struct {
		id       string
		orders   string
		wantDeny Code
		setup    func(t *testing.T, gsvc *Service) prepared
	}
	// guardOn returns the guarded call of subject u through gsvc in a fresh request scope (resolved before any ordering)
	// of u's credential for org (G016 A2).
	guardOn := func(t *testing.T, gsvc *Service, u *itUser, sel tenancy.GroupID, call func(ctx context.Context, tc *TenantContext) error) func() error {
		rctx, _ := beginAs(t, gsvc, ctx, u, org)
		tc := resolve(t, gsvc, rctx, org, sel)
		return func() error { return call(rctx, tc) }
	}
	updateProfile := func(svc *Service, p tenancy.BusinessProfile) func(context.Context, *TenantContext) error {
		return func(c context.Context, tc *TenantContext) error {
			_, err := svc.UpdateBusinessProfile(c, tc, p.ID, tenancy.BusinessContext{DisplayName: p.Context.DisplayName + " Guarded"})
			return err
		}
	}
	renameGroup := func(svc *Service, g tenancy.Group) func(context.Context, *TenantContext) error {
		return func(c context.Context, tc *TenantContext) error {
			_, err := svc.RenameGroup(c, tc, g.ID, g.Name+" Guarded")
			return err
		}
	}
	setPolicy := func(svc *Service, v bool) func(context.Context, *TenantContext) error {
		return func(c context.Context, tc *TenantContext) error {
			_, err := svc.SetOrganizationPolicy(c, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: v})
			return err
		}
	}
	// G018 onboarding completions and their target: the Group's associations and the Organization's profile count.
	completeNew := func(svc *Service, g tenancy.Group) func(context.Context, *TenantContext) error {
		return func(c context.Context, tc *TenantContext) error {
			_, err := svc.CompleteGroupOnboardingWithNewProfile(c, tc, g.ID, tenancy.BusinessContext{DisplayName: g.Name + " Profile"})
			return err
		}
	}
	completeExisting := func(svc *Service, g tenancy.Group, p tenancy.BusinessProfileID) func(context.Context, *TenantContext) error {
		return func(c context.Context, tc *TenantContext) error {
			_, err := svc.CompleteGroupOnboardingWithProfile(c, tc, g.ID, p)
			return err
		}
	}
	onbTarget := func(t *testing.T, g tenancy.GroupID) func() string {
		return func() string {
			var a, n int
			if err := e.setup.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.group_business_profiles WHERE group_id = $1::text::uuid),
  (SELECT count(*) FROM app.business_profiles WHERE organization_id = $2::text::uuid)`, g.String(), org.String()).Scan(&a, &n); err != nil {
				t.Fatal(err)
			}
			return fmt.Sprintf("%d/%d", a, n)
		}
	}
	// writerModel returns the accepted tenancy model over a hookDB, so that a writer can pause before its COMMIT.
	// Conflicting tenancy writers are trusted model primitives: from version 6 they run on the attested setup pool
	// (G008 RC-8), taking the same row locks; the guarded operations stay on the runtime pool.
	writerModel := func() (*tenancy.Service, func() *pause) {
		h := &hookDB{pool: e.setupDB}
		return tenancy.NewService(h, nil), func() *pause {
			p := newPause()
			h.on("before", "COMMIT", p.hook)
			return p
		}
	}
	// G019 (Plan r1 §6.2 O-MEM): a Group owned by u (the creator's trigger-written owner binding, G018), the group-scope
	// ownership transfer to rcp with u keeping the Group member role, and the target the transfer changes and no
	// conflicting writer does: the Group's owner subjects.
	ownedGroup := func(t *testing.T, u *itUser) tenancy.Group {
		slugSeq++
		return e.group(t, org, u, fmt.Sprintf("Ordering Group %d", slugSeq), fmt.Sprintf("ordering-%d", slugSeq))
	}
	groupScope := func(g tenancy.GroupID) Scope { return Scope{Kind: tenancy.ScopeGroup, Group: &g} }
	transfer := func(svc *Service, g tenancy.GroupID, rcp tenancy.UserID) func(context.Context, *TenantContext) error {
		return func(c context.Context, tc *TenantContext) error {
			return svc.TransferOwnership(c, tc, groupScope(g), rcp, string(tenancy.GroupMember))
		}
	}
	ownersOf := func(t *testing.T, g tenancy.GroupID) func() string {
		return func() string {
			var s string
			if err := e.setup.QueryRow(ctx, `SELECT coalesce(string_agg(subject_user_id::text, ',' ORDER BY subject_user_id), '') FROM app.role_bindings
WHERE scope_type = 'group' AND scope_id = $1::text::uuid AND role = 'group_owner'`, g.String()).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
	}

	cases := []caseDef{
		{"O-1 logout (current) versus RenameGroup", "AB", CodeUnauthenticated, func(t *testing.T, gsvc *Service) prepared {
			g := newGroup(t)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			cred := u.in(t, org) // G016 A2: the current-session logout revokes the guard's own credential
			return prepared{guard: guardOn(t, gsvc, u, zeroG, renameGroup(gsvc, g)),
				writer: func() error {
					_, err := e.id.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: e.csrf(t, cred), Origin: itOriginOf(org), Credential: cred})
					return err
				}, target: groupRow(t, g.ID)}
		}},
		{"O-2 logout-all versus UpdateBusinessProfile", "AB", CodeUnauthenticated, func(t *testing.T, gsvc *Service) prepared {
			// G060 (r0 A4): the Organization admin's profile path needs the Group-management override.
			setOverride(t, true)
			t.Cleanup(func() { setOverride(t, false) })
			p := newProfile(t)
			u := member(t, tenancy.OrgAdmin, nil)
			other := e.relogin(t, u)
			return prepared{guard: guardOn(t, gsvc, u, zeroG, updateProfile(gsvc, p)),
				writer: func() error {
					_, err := e.id.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: e.csrf(t, other), Origin: itOrigin, Credential: other})
					return err
				}, target: profileRow(t, p.ID)}
		}},
		{"O-3 DisableUser versus RenameGroup", "ABC", CodeUnauthenticated, func(t *testing.T, gsvc *Service) prepared {
			g := newGroup(t)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupOwner})
			return prepared{guard: guardOn(t, gsvc, u, zeroG, renameGroup(gsvc, g)),
				writer: func() error { return idTraced.DisableUser(ctx, u.id.String()) },
				armC:   func() *pause { return tracer.arm("revoke_reason = 'disabled'") }, target: groupRow(t, g.ID)}
		}},
		{"O-4 sign-in rotation versus RenameGroup", "AB", CodeUnauthenticated, func(t *testing.T, gsvc *Service) prepared {
			g := newGroup(t)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			// G016 A2: the rotation supersedes the guard's own credential at org's origin; u is a member of org (D-2(a)), so no
			// admission is needed.
			cred := u.in(t, org)
			return prepared{guard: guardOn(t, gsvc, u, zeroG, renameGroup(gsvc, g)),
				writer: func() error { _, err := e.rotate(idTraced, u, "", cred, itOriginOf(org)); return err }, target: groupRow(t, g.ID)}
		}},
		{"O-6 RemoveOrganizationMembership versus UpdateBusinessProfile (Organization path)", "ABC", CodeContextDenied, func(t *testing.T, gsvc *Service) prepared {
			// G060 (r0 A4): the Organization path exists only with the Group-management override.
			setOverride(t, true)
			t.Cleanup(func() { setOverride(t, false) })
			p := newProfile(t)
			u := member(t, tenancy.OrgAdmin, nil)
			wm, armC := writerModel()
			return prepared{guard: guardOn(t, gsvc, u, zeroG, updateProfile(gsvc, p)),
				writer: func() error { return wm.RemoveOrganizationMembership(ctx, org, u.id) }, armC: armC, target: profileRow(t, p.ID)}
		}},
		{"O-7 ChangeOrganizationRole owner to member versus SetOrganizationPolicy", "ABC", CodeForbidden, func(t *testing.T, gsvc *Service) prepared {
			u := member(t, tenancy.OrgOwner, nil)
			wm, armC := writerModel()
			e.setFlag(t, org, true)
			t.Cleanup(func() { e.setFlag(t, org, true) })
			return prepared{guard: guardOn(t, gsvc, u, zeroG, setPolicy(gsvc, false)),
				writer: func() error { _, err := wm.ChangeOrganizationRole(ctx, org, u.id, tenancy.OrgMember); return err }, armC: armC, target: orgRow(t)}
		}},
		{"O-8 RemoveGroupMembership versus UpdateBusinessProfile (Group path)", "ABC", CodeNotFound, func(t *testing.T, gsvc *Service) prepared {
			g, p := newGroup(t), newProfile(t)
			e.link(t, org, g.ID, p.ID)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			wm, armC := writerModel()
			return prepared{guard: guardOn(t, gsvc, u, zeroG, updateProfile(gsvc, p)),
				writer: func() error { return wm.RemoveGroupMembership(ctx, org, g.ID, u.id) }, armC: armC, target: profileRow(t, p.ID)}
		}},
		{"O-9 ChangeGroupRole admin to member versus RenameGroup", "ABC", CodeForbidden, func(t *testing.T, gsvc *Service) prepared {
			g := newGroup(t)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			wm, armC := writerModel()
			return prepared{guard: guardOn(t, gsvc, u, zeroG, renameGroup(gsvc, g)),
				writer: func() error { _, err := wm.ChangeGroupRole(ctx, org, g.ID, u.id, tenancy.GroupMember); return err }, armC: armC,
				target: groupRow(t, g.ID)}
		}},
		{"O-10 UnlinkGroupBusinessProfile versus UpdateBusinessProfile (Group path)", "ABC", CodeNotFound, func(t *testing.T, gsvc *Service) prepared {
			g, p := newGroup(t), newProfile(t)
			e.link(t, org, g.ID, p.ID)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupOwner})
			wm, armC := writerModel()
			return prepared{guard: guardOn(t, gsvc, u, zeroG, updateProfile(gsvc, p)),
				writer: func() error { return wm.UnlinkGroupBusinessProfile(ctx, org, g.ID, p.ID) }, armC: armC, target: profileRow(t, p.ID)}
		}},
		{"O-11 SetOrganizationPolicy(false) versus UpdateBusinessProfile (Group path)", "ABC", CodeForbidden, func(t *testing.T, gsvc *Service) prepared {
			g, p := newGroup(t), newProfile(t)
			e.link(t, org, g.ID, p.ID)
			e.setFlag(t, org, true)
			t.Cleanup(func() { e.setFlag(t, org, true) })
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			psvc, phook := e.service(e.id)
			pol := guardOn(t, psvc, owner, zeroG, setPolicy(psvc, false))
			return prepared{guard: guardOn(t, gsvc, u, zeroG, updateProfile(gsvc, p)), writer: pol,
				armC: func() *pause { p := newPause(); phook.on("before", "COMMIT", p.hook); return p }, target: profileRow(t, p.ID)}
		}},
		// G018 O-ONB-1..3 (ONB-7): pre-commit revocations against the onboarding completions on the hookDB seams. Order A
		// pauses the guard before its first insert (S-P1, or S-P2 for an existing profile).
		{"O-ONB-1 RemoveGroupMembership versus CompleteGroupOnboardingWithNewProfile", "ABC", CodeNotFound, func(t *testing.T, gsvc *Service) prepared {
			g := newGroup(t)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupOwner})
			wm, armC := writerModel()
			return prepared{guard: guardOn(t, gsvc, u, zeroG, completeNew(gsvc, g)), step: "IP",
				writer: func() error { return wm.RemoveGroupMembership(ctx, org, g.ID, u.id) }, armC: armC, target: onbTarget(t, g.ID)}
		}},
		{"O-ONB-2 ChangeGroupRole admin to member versus CompleteGroupOnboardingWithNewProfile", "ABC", CodeForbidden, func(t *testing.T, gsvc *Service) prepared {
			g := newGroup(t)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			wm, armC := writerModel()
			return prepared{guard: guardOn(t, gsvc, u, zeroG, completeNew(gsvc, g)), step: "IP",
				writer: func() error { _, err := wm.ChangeGroupRole(ctx, org, g.ID, u.id, tenancy.GroupMember); return err }, armC: armC,
				target: onbTarget(t, g.ID)}
		}},
		{"O-ONB-3 DisableUser versus CompleteGroupOnboardingWithProfile", "ABC", CodeUnauthenticated, func(t *testing.T, gsvc *Service) prepared {
			g, ga, p := newGroup(t), newGroup(t), newProfile(t)
			e.link(t, org, ga.ID, p.ID)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupOwner, ga.ID: tenancy.GroupOwner})
			return prepared{guard: guardOn(t, gsvc, u, zeroG, completeExisting(gsvc, g, p.ID)), step: "IA",
				writer: func() error { return idTraced.DisableUser(ctx, u.id.String()) },
				armC:   func() *pause { return tracer.arm("revoke_reason = 'disabled'") }, target: onbTarget(t, g.ID)}
		}},
		{"O-ONB-3s logout-all versus CompleteGroupOnboardingWithNewProfile", "AB", CodeUnauthenticated, func(t *testing.T, gsvc *Service) prepared {
			g := newGroup(t)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupOwner})
			other := e.relogin(t, u)
			return prepared{guard: guardOn(t, gsvc, u, zeroG, completeNew(gsvc, g)), step: "IP",
				writer: func() error {
					_, err := e.id.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: e.csrf(t, other), Origin: itOrigin, Credential: other})
					return err
				}, target: onbTarget(t, g.ID)}
		}},
		{"O-ONB-3p removal of the last qualifying path versus CompleteGroupOnboardingWithProfile", "ABC", CodeNotFound, func(t *testing.T, gsvc *Service) prepared {
			g, ga, p := newGroup(t), newGroup(t), newProfile(t)
			e.link(t, org, ga.ID, p.ID)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupOwner, ga.ID: tenancy.GroupOwner})
			wm, armC := writerModel()
			return prepared{guard: guardOn(t, gsvc, u, zeroG, completeExisting(gsvc, g, p.ID)), step: "IA",
				writer: func() error { return wm.RemoveGroupMembership(ctx, org, ga.ID, u.id) }, armC: armC, target: onbTarget(t, g.ID)}
		}},
		{"O-12 removal of the selected-Group binding versus UpdateBusinessProfile in selected context", "AB", CodeContextDenied, func(t *testing.T, gsvc *Service) prepared {
			g, p := newGroup(t), newProfile(t)
			e.link(t, org, g.ID, p.ID)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			return prepared{guard: guardOn(t, gsvc, u, g.ID, updateProfile(gsvc, p)),
				writer: func() error { return e.model.RemoveGroupMembership(ctx, org, g.ID, u.id) }, target: profileRow(t, p.ID)}
		}},
		// G019 O-MEM-2 (AC3): the ownership transfer racing the recipient's disable and removal and the actor's session
		// revocation. A revocation committed before the guard (B) or in flight while the guard waits behind it (C) denies
		// the transfer with the owners unchanged; the actor's logout-all also queues behind the guard's session share (A).
		// The recipient's disable and removal take no lock the guard holds before its write, so order A does not apply to
		// them (O-MEM-3(c) covers the promotion committing first).
		{"O-MEM-2a DisableUser of the recipient versus TransferOwnership", "BC", CodeConflict, func(t *testing.T, gsvc *Service) prepared {
			actor := member(t, tenancy.OrgMember, nil)
			g := ownedGroup(t, actor)
			rcp := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			return prepared{guard: guardOn(t, gsvc, actor, zeroG, transfer(gsvc, g.ID, rcp.id)),
				writer: func() error { return idTraced.DisableUser(ctx, rcp.id.String()) },
				armC:   func() *pause { return tracer.arm("revoke_reason = 'disabled'") }, target: ownersOf(t, g.ID)}
		}},
		{"O-MEM-2b removal of the recipient's Group binding versus TransferOwnership", "BC", CodeNotFound, func(t *testing.T, gsvc *Service) prepared {
			actor := member(t, tenancy.OrgMember, nil)
			g := ownedGroup(t, actor)
			rcp := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			wm, armC := writerModel()
			return prepared{guard: guardOn(t, gsvc, actor, zeroG, transfer(gsvc, g.ID, rcp.id)),
				writer: func() error { return wm.RemoveGroupMembership(ctx, org, g.ID, rcp.id) }, armC: armC, target: ownersOf(t, g.ID)}
		}},
		{"O-MEM-2c logout-all of the actor versus TransferOwnership", "AB", CodeUnauthenticated, func(t *testing.T, gsvc *Service) prepared {
			actor := member(t, tenancy.OrgMember, nil)
			g := ownedGroup(t, actor)
			rcp := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			other := e.relogin(t, actor)
			return prepared{guard: guardOn(t, gsvc, actor, zeroG, transfer(gsvc, g.ID, rcp.id)),
				writer: func() error {
					_, err := e.id.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: e.csrf(t, other), Origin: itOrigin, Credential: other})
					return err
				}, target: ownersOf(t, g.ID)}
		}},
	}

	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			for _, order := range c.orders {
				gsvc, ghook := e.service(e.id)
				pr := c.setup(t, gsvc)
				before := pr.target()
				switch order {
				case 'A':
					p := newPause()
					step := pr.step
					if step == "" {
						step = "W"
					}
					ghook.on("before", step, p.hook)
					g := start(pr.guard)
					p.wait(t, "guard")
					w := start(pr.writer)
					e.waitWaiters(t, 1)
					close(p.release)
					gerr, werr := g.wait(t, "guard"), w.wait(t, "writer")
					if gerr != nil || werr != nil || g.at > w.at {
						t.Errorf("A: guard %v (finished %d), writer %v (finished %d)", gerr, g.at, werr, w.at)
					}
					if pr.target() == before {
						t.Error("A: the guarded write did not persist")
					}
				case 'B':
					p := newPause()
					ghook.on("before", "BEGIN", p.hook)
					g := start(pr.guard)
					p.wait(t, "guard")
					if err := pr.writer(); err != nil {
						t.Fatalf("B: writer: %v", err)
					}
					close(p.release)
					wantCode(t, "B: guard after the writer", g.wait(t, "guard"), c.wantDeny)
					if pr.target() != before {
						t.Error("B: the target changed")
					}
				case 'C':
					wp := pr.armC()
					w := start(pr.writer)
					wp.wait(t, "writer")
					g := start(pr.guard)
					e.waitWaiters(t, 1)
					close(wp.release)
					if err := w.wait(t, "writer"); err != nil {
						t.Fatalf("C: writer: %v", err)
					}
					wantCode(t, "C: guard behind the in-flight writer", g.wait(t, "guard"), c.wantDeny)
					if pr.target() != before {
						t.Error("C: the target changed")
					}
				}
				ordering(t, map[string]any{"case": c.id[:strings.Index(c.id, " ")], "order": string(order), "guard_reason": e.lastReason(t)})
			}
		})
	}

	t.Run("O-5 expiry at the lock point", func(t *testing.T) {
		for _, expired := range []bool{false, true} {
			clk := newITClock()
			clk.t = e.clk.Now()
			xsvc, xhook := e.service(e.identity(t, clk, e.pool))
			g := newGroup(t)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			guard := guardOn(t, xsvc, u, zeroG, renameGroup(xsvc, g))
			before := groupRow(t, g.ID)()
			// The clock moves while the guard is paused after BEGIN: the session predicate is evaluated once, at the lock.
			p := newPause()
			xhook.on("before", "BEGIN", p.hook)
			a := start(guard)
			p.wait(t, "guard")
			if expired {
				clk.Advance(itIdle + time.Second)
			} else {
				clk.Advance(itIdle - time.Minute)
			}
			close(p.release)
			err := a.wait(t, "guard")
			if expired {
				wantCode(t, "expired at the lock point", err, CodeUnauthenticated)
				if groupRow(t, g.ID)() != before {
					t.Error("expired guard wrote")
				}
			} else if err != nil {
				t.Errorf("unexpired at the lock point: %v", err)
			}
			ordering(t, map[string]any{"case": "O-5", "expired_at_lock_point": expired, "outcome": map[bool]string{true: "allow", false: string(CodeOf(err))}[err == nil]})
		}
	})

	t.Run("O-13 unlink of g1 while the g2 path holds", func(t *testing.T) {
		for _, order := range "AB" {
			gx, gy, p := newGroup(t), newGroup(t), newProfile(t)
			e.link(t, org, gx.ID, p.ID)
			e.link(t, org, gy.ID, p.ID)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{gx.ID: tenancy.GroupAdmin, gy.ID: tenancy.GroupAdmin})
			gsvc, ghook := e.service(e.id)
			guard := guardOn(t, gsvc, u, zeroG, updateProfile(gsvc, p))
			unlink := func() error { return e.model.UnlinkGroupBusinessProfile(ctx, org, gx.ID, p.ID) }
			step := map[rune]string{'A': "W", 'B': "BEGIN"}[order]
			pp := newPause()
			ghook.on("before", step, pp.hook)
			g := start(guard)
			pp.wait(t, "guard")
			if order == 'A' {
				w := start(unlink)
				e.waitWaiters(t, 1)
				close(pp.release)
				if err := w.wait(t, "unlink"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := unlink(); err != nil {
					t.Fatal(err)
				}
				close(pp.release)
			}
			if err := g.wait(t, "guard"); err != nil {
				t.Errorf("%c: the g2 path must still allow: %v", order, err)
			}
			ordering(t, map[string]any{"case": "O-13", "order": string(order), "outcome": "allow"})
		}
	})

	t.Run("O-14 guarded writers in both orders without a deadlock", func(t *testing.T) {
		g, p := newGroup(t), newProfile(t)
		e.link(t, org, g.ID, p.ID)
		u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
		pairs := []struct {
			name   string
			first  func(*Service) func() error
			second func(*Service) func() error
		}{
			{"rename then profile update", func(s *Service) func() error { return guardOn(t, s, u, zeroG, renameGroup(s, g)) },
				func(s *Service) func() error { return guardOn(t, s, u, zeroG, updateProfile(s, p)) }},
			{"profile update then rename", func(s *Service) func() error { return guardOn(t, s, u, zeroG, updateProfile(s, p)) },
				func(s *Service) func() error { return guardOn(t, s, u, zeroG, renameGroup(s, g)) }},
			{"policy then profile update", func(s *Service) func() error { return guardOn(t, s, owner, zeroG, setPolicy(s, true)) },
				func(s *Service) func() error { return guardOn(t, s, u, zeroG, updateProfile(s, p)) }},
			{"profile update then policy", func(s *Service) func() error { return guardOn(t, s, u, zeroG, updateProfile(s, p)) },
				func(s *Service) func() error { return guardOn(t, s, owner, zeroG, setPolicy(s, true)) }},
		}
		for _, pr := range pairs {
			s1, h1 := e.service(e.id)
			s2, _ := e.service(e.id)
			first, second := pr.first(s1), pr.second(s2)
			pp := newPause()
			h1.on("before", "W", pp.hook)
			a := start(first)
			pp.wait(t, pr.name)
			b := start(second)
			e.waitWaiters(t, 1)
			close(pp.release)
			ea, eb := a.wait(t, "first"), b.wait(t, "second")
			if ea != nil || eb != nil || isDeadlock(ea) || isDeadlock(eb) || a.at > b.at {
				t.Errorf("%s: %v / %v", pr.name, ea, eb)
			}
			ordering(t, map[string]any{"case": "O-14", "pair": pr.name, "outcome": "both allowed, no deadlock"})
		}
	})

	t.Run("O-15 sign-in rotation of the guard's own session with a changed display name", func(t *testing.T) {
		g := newGroup(t)
		u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
		gsvc, _ := e.service(e.id)
		guard := guardOn(t, gsvc, u, zeroG, renameGroup(gsvc, g))
		before := groupRow(t, g.ID)()
		// G016 A2, as O-4: the rotation supersedes the guard's own credential at org's origin (u is a member of org).
		cred, origin := u.in(t, org), itOriginOf(org)
		rp := tracer.arm("revoke_reason = 'superseded'")
		var newCred string
		rot := start(func() error {
			var err error
			newCred, err = e.rotate(idTraced, u, "Rotated Display Name", cred, origin)
			return err
		})
		rp.wait(t, "rotation")
		ga := start(guard)
		e.waitWaiters(t, 1) // the guard holds U FOR SHARE and waits for S
		close(rp.release)   // the rotation's display-name UPDATE now waits for the guard's U share: a cycle
		gerr, rerr := ga.wait(t, "guard"), rot.wait(t, "rotation")
		var victim string
		switch {
		case gerr != nil && rerr == nil:
			victim = "guard"
			wantCode(t, "aborted guard", gerr, CodeUnavailable)
			if r := e.lastReason(t); r != string(reasonAborted) {
				t.Errorf("aborted guard reason %s", r)
			}
			if groupRow(t, g.ID)() != before {
				t.Error("the aborted guard wrote")
			}
			if _, err := e.id.Session(ctx, newCred, origin); err != nil {
				t.Errorf("the rotation did not commit: %v", err)
			}
		case rerr != nil && gerr == nil:
			victim = "rotation"
			if !errors.Is(rerr, identity.ErrUnavailable) {
				t.Errorf("aborted rotation: %v", rerr)
			}
			if _, err := e.id.Session(ctx, cred, origin); err != nil {
				t.Errorf("the old session must stay live: %v", err)
			}
			if groupRow(t, g.ID)() == before {
				t.Error("the guard did not commit")
			}
		default:
			t.Fatalf("want exactly one deadlock victim: guard %v, rotation %v", gerr, rerr)
		}
		ordering(t, map[string]any{"case": "O-15", "victim": victim, "outcome": "one definitive abort, no unauthorized write"})
	})

	t.Run("E-PG-3a a policy write queued behind a guard's Organization share", func(t *testing.T) {
		e.setFlag(t, org, true)
		defer e.setFlag(t, org, true)
		g1, g2, p1, p2 := newGroup(t), newGroup(t), newProfile(t), newProfile(t)
		e.link(t, org, g1.ID, p1.ID)
		e.link(t, org, g2.ID, p2.ID)
		m1 := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g1.ID: tenancy.GroupAdmin})
		m2 := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g2.ID: tenancy.GroupAdmin})
		s1, h1 := e.service(e.id)
		sp, _ := e.service(e.id)
		s2, _ := e.service(e.id)
		t1, pol, t2 := guardOn(t, s1, m1, zeroG, updateProfile(s1, p1)), guardOn(t, sp, owner, zeroG, setPolicy(sp, false)),
			guardOn(t, s2, m2, zeroG, updateProfile(s2, p2))
		pp := newPause()
		h1.on("before", "W", pp.hook)
		a1 := start(t1)
		pp.wait(t, "T1")
		ap := start(pol)
		e.waitWaiters(t, 1) // the policy write waits for T1's Organization share
		a2 := start(t2)
		obs := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		deadline := time.Now().Add(15 * time.Second)
		for !a2.finished() && e.waiters(t, obs) < 2 && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond) // waitWaiters' polling: T2 either completes or becomes a second waiter
		}
		passed := a2.finished()
		close(pp.release)
		err1, errP, err2 := a1.wait(t, "T1"), ap.wait(t, "policy"), a2.wait(t, "T2")
		if err1 != nil || errP != nil || a1.at > ap.at {
			t.Fatalf("T1 %v, policy %v", err1, errP)
		}
		switch {
		case err2 == nil && a2.at < ap.at:
		case CodeOf(err2) == CodeForbidden && a2.at > ap.at:
		default:
			t.Errorf("T2 neither preceded the policy commit nor observed false: %v (T2 %d, policy %d)", err2, a2.at, ap.at)
		}
		ordering(t, map[string]any{"case": "E-PG-3a", "t2_passed_queued_writer": passed, "t2_outcome": map[bool]string{true: "allow", false: string(CodeOf(err2))}[err2 == nil],
			"t2_before_policy_commit": a2.at < ap.at})
	})

	t.Run("E-PG-3b a removal queued behind the subject's own guard", func(t *testing.T) {
		// G060 (r0 A4): an Organization admin's profile write is admitted only with the Group-management override, so the
		// owner enables it for this characterization (the queued-writer shape is unchanged) and disables it afterwards.
		setOverride(t, true)
		defer setOverride(t, false)
		p := newProfile(t)
		m := member(t, tenancy.OrgAdmin, nil)
		s1, h1 := e.service(e.id)
		s2, _ := e.service(e.id)
		t1, t2 := guardOn(t, s1, m, zeroG, updateProfile(s1, p)), guardOn(t, s2, m, zeroG, updateProfile(s2, p))
		pp := newPause()
		h1.on("before", "W", pp.hook)
		a1 := start(t1)
		pp.wait(t, "T1")
		ar := start(func() error { return e.model.RemoveOrganizationMembership(ctx, org, m.id) })
		e.waitWaiters(t, 1) // the removal waits for T1's binding share
		a2 := start(t2)
		obs := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		deadline := time.Now().Add(15 * time.Second)
		for !a2.finished() && e.waiters(t, obs) < 2 && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}
		passed := a2.finished()
		close(pp.release)
		err1, errR, err2 := a1.wait(t, "T1"), ar.wait(t, "removal"), a2.wait(t, "T2")
		if err1 != nil || errR != nil || a1.at > ar.at {
			t.Fatalf("T1 %v, removal %v", err1, errR)
		}
		switch {
		case err2 == nil && a2.at < ar.at:
		case CodeOf(err2) == CodeContextDenied && a2.at > ar.at:
		default:
			t.Errorf("T2 neither preceded the removal nor observed it: %v (T2 %d, removal %d)", err2, a2.at, ar.at)
		}
		lctx, _ := beginAs(t, s1, ctx, m, org) // G016 A2: m's credential for org (the membership removal does not end it)
		_, err := s1.Resolve(lctx, ResolveRequest{OrganizationID: org.String()})
		wantCode(t, "a new guard after the removal", err, CodeContextDenied)
		ordering(t, map[string]any{"case": "E-PG-3b", "t2_passed_queued_writer": passed, "t2_outcome": map[bool]string{true: "allow", false: string(CodeOf(err2))}[err2 == nil],
			"t2_before_removal_commit": a2.at < ar.at})
	})
	t.Run("O-ONB-D combined-facts lock order in both interleavings without a deadlock", func(t *testing.T) {
		for _, firstA := range []bool{true, false} {
			gs := []tenancy.Group{newGroup(t), newGroup(t), newGroup(t)}
			sort.Slice(gs, func(i, j int) bool { return bytes.Compare(gs[i].ID[:], gs[j].ID[:]) < 0 })
			g1, g2, g3 := gs[0], gs[1], gs[2]
			p3 := newProfile(t)
			e.link(t, org, g3.ID, p3.ID)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g1.ID: tenancy.GroupOwner, g2.ID: tenancy.GroupOwner, g3.ID: tenancy.GroupOwner})
			sA, hA := e.service(e.id)
			sB, hB := e.service(e.id)
			callA := guardOn(t, sA, u, zeroG, completeExisting(sA, g1, p3.ID))
			callB := guardOn(t, sB, u, zeroG, completeExisting(sB, g2, p3.ID))
			first, second, hFirst := callA, callB, hA
			if !firstA {
				first, second, hFirst = callB, callA, hB
			}
			pp := newPause()
			hFirst.on("before", "IA", pp.hook)
			a := start(first)
			pp.wait(t, "first completion")
			b := start(second)
			e.waitWaiters(t, 1)
			close(pp.release)
			ea, eb := a.wait(t, "first"), b.wait(t, "second")
			if ea != nil || eb != nil || isDeadlock(ea) || isDeadlock(eb) || a.at > b.at {
				t.Errorf("A first=%v: %v / %v", firstA, ea, eb)
			}
			for _, g := range []tenancy.GroupID{g1.ID, g2.ID, g3.ID} {
				var n int
				if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.group_business_profiles WHERE group_id = $1::text::uuid AND business_profile_id = $2::text::uuid`,
					g.String(), p3.ID.String()).Scan(&n); err != nil || n != 1 {
					t.Errorf("A first=%v: Group %s has %d associations to p3 (%v), want 1", firstA, g, n, err)
				}
			}
			if l := hA.lockedRows(); len(l) == 0 {
				t.Errorf("A first=%v: completion A took no row locks", firstA)
			}
			ordering(t, map[string]any{"case": "O-ONB-D", "first": map[bool]string{true: "A", false: "B"}[firstA], "outcome": "both allowed, no deadlock"})
		}
	})

	t.Run("O-ONB-E an UPDATE-refreshed or equal-timestamp profile is not eligible after a revocation", func(t *testing.T) {
		e.setFlag(t, org, true)
		rt := e.rlsRuntime(t)
		type fx struct {
			gA, gB tenancy.Group
			p      tenancy.BusinessProfile
			u      *itUser
		}
		fixture := func() fx {
			f := fx{gA: newGroup(t), gB: newGroup(t), p: newProfile(t)}
			e.link(t, org, f.gA.ID, f.p.ID)
			f.u = member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{f.gA.ID: tenancy.GroupOwner, f.gB.ID: tenancy.GroupOwner})
			return f
		}
		begin := func(f fx) pgx.Tx {
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if err := rlsSet(ctx, tx, ptr(f.u.id.String()), ptr(org.String()), ptr(""), nil); err != nil {
				t.Fatal(err)
			}
			return tx
		}
		reader := func(tx pgx.Tx, p tenancy.BusinessProfileID) bool {
			var ok bool
			if err := tx.QueryRow(ctx, `SELECT app.rls_profile_eligible($1::text::uuid)`, p.String()).Scan(&ok); err != nil {
				t.Fatal(err)
			}
			return ok
		}
		update := func(tx pgx.Tx, f fx) {
			tag, err := tx.Exec(ctx, `UPDATE app.business_profiles SET description = $2 WHERE id = $1::text::uuid`, f.p.ID.String(), "Refreshed "+e.canaries[1])
			if err != nil || tag.RowsAffected() != 1 {
				t.Fatalf("T1 profile UPDATE: %v rows %d, want one row", err, tag.RowsAffected())
			}
		}
		revoke := func(f fx) {
			w := start(func() error { return e.model.RemoveGroupMembership(ctx, org, f.gA.ID, f.u.id) })
			if err := w.wait(t, "T2 revocation (must not block on T1)"); err != nil {
				t.Fatal(err)
			}
		}
		link := func(tx pgx.Tx, g tenancy.GroupID, p tenancy.BusinessProfileID) (string, int64) {
			return rlsStmt(t, tx, `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid)`,
				org.String(), g.String(), p.String())
		}
		assocs := func(g tenancy.GroupID) int {
			var n int
			if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.group_business_profiles WHERE group_id = $1::text::uuid`, g.String()).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}

		// (a) UPDATE then revocation: 42501.
		f := fixture()
		tx := begin(f)
		update(tx, f)
		if !reader(tx, f.p.ID) {
			t.Error("(a) before the revocation the gA path must make p eligible")
		}
		revoke(f)
		if reader(tx, f.p.ID) {
			t.Error("(a) an UPDATE-refreshed profile is eligible after the revocation")
		}
		if st, n := link(tx, f.gB.ID, f.p.ID); st != "42501" || n != 0 {
			t.Errorf("(a) association after the revocation: SQLSTATE %q rows %d, want 42501", st, n)
		}
		_ = tx.Rollback(ctx)
		if assocs(f.gB.ID) != 0 {
			t.Error("(a) gB gained an association")
		}
		ordering(t, map[string]any{"case": "O-ONB-E", "branch": "update-then-revocation", "outcome": "42501"})

		// (b) Mirror positive: a profile inserted by T1 itself attaches after the identical revocation.
		f = fixture()
		tx = begin(f)
		update(tx, f)
		revoke(f)
		pN := rlsUUID(t)
		if _, err := tx.Exec(ctx, `INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES ($1::text::uuid, $2::text::uuid, 'Mirror Positive')`, pN, org.String()); err != nil {
			t.Fatalf("(b) S-P1: %v", err)
		}
		if st, n := link(tx, f.gB.ID, tenancy.BusinessProfileID(mustUUID(t, pN))); st != "" || n != 1 {
			t.Errorf("(b) the new-profile association: SQLSTATE %q rows %d, want one row", st, n)
		}
		_ = tx.Rollback(ctx)
		ordering(t, map[string]any{"case": "O-ONB-E", "branch": "mirror-new-profile", "outcome": "admitted"})

		// (c) Equal timestamp: created_at set to T1's own transaction timestamp by the owner, then the (a) sequence.
		f = fixture()
		tx = begin(f)
		var ts time.Time
		if err := tx.QueryRow(ctx, `SELECT transaction_timestamp()`).Scan(&ts); err != nil {
			t.Fatal(err)
		}
		if _, err := e.setup.Exec(ctx, `UPDATE app.business_profiles SET created_at = $2 WHERE id = $1::text::uuid`, f.p.ID.String(), ts); err != nil {
			t.Fatal(err)
		}
		update(tx, f)
		var equal bool
		if err := tx.QueryRow(ctx, `SELECT created_at = transaction_timestamp() FROM app.business_profiles WHERE id = $1::text::uuid`, f.p.ID.String()).Scan(&equal); err != nil || !equal {
			t.Fatalf("(c) the fixture timestamp does not equal T1's (%v %v)", equal, err)
		}
		revoke(f)
		if reader(tx, f.p.ID) {
			t.Error("(c) an equal-timestamp refreshed profile is eligible after the revocation")
		}
		if st, n := link(tx, f.gB.ID, f.p.ID); st != "42501" || n != 0 {
			t.Errorf("(c) association after the revocation: SQLSTATE %q rows %d, want 42501", st, n)
		}
		_ = tx.Rollback(ctx)
		ordering(t, map[string]any{"case": "O-ONB-E", "branch": "equal-timestamp", "outcome": "42501"})
	})

	// G018 availability addendum: a completion's lifetime is bounded while it waits. (1) At the pre-decision target
	// Group lock an earlier caller deadline is kept. (2) In the act callback's write itself (no caller deadline: only
	// the decision budget can end it; a callback still holding the caller's context would wait for the holder), for
	// both completions. Each ends as the closed unavailable failure, rolls back (no profile, no association, no open
	// transaction of the runtime role), and the conflicting writer and a later completion of the same Group progress.
	t.Run("O-ONB-T completion lifetimes are bounded at the Group lock and in the act write", func(t *testing.T) {
		// G060 (r0 A2): the Organization owner completes another creator's Group only with the Group-management override.
		setOverride(t, true)
		defer setOverride(t, false)
		obs := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		gsvc, _ := e.service(e.id)
		openTx := func() int {
			var n int
			if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM pg_stat_activity WHERE usename = $1 AND xact_start IS NOT NULL AND pid <> pg_backend_pid()`,
				e.cfg.RuntimeRole()).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		// holdBeforeCommit runs a conflicting model writer that pauses before its COMMIT, holding its row locks.
		holdBeforeCommit := func(what string, w func(m *tenancy.Service) error) (*actor, *pause) {
			m, armC := writerModel()
			pc := armC()
			a := start(func() error { return w(m) })
			pc.wait(t, what)
			return a, pc
		}
		// settled polls n until it is zero (a connection closed by a canceled statement leaves its backend briefly).
		settled := func(n func() int) int {
			v := n()
			for end := time.Now().Add(3 * time.Second); v != 0 && time.Now().Before(end); v = n() {
				time.Sleep(20 * time.Millisecond)
			}
			return v
		}
		// bounded runs call and fails the subtest unless it ends as unavailable within [lo, hi]; waitLock first observes
		// its lock wait.
		bounded := func(what string, call func() error, lo, hi time.Duration, waitLock bool) {
			t.Helper()
			began := time.Now()
			a := start(call)
			if waitLock {
				e.waitWaiters(t, 1)
			}
			select {
			case <-a.done:
			case <-time.After(hi + 5*time.Second):
				t.Errorf("%s: still waiting after %v (the wait is not bounded)", what, hi+5*time.Second)
				return
			}
			el := time.Since(began)
			wantCode(t, what, a.err, CodeUnavailable)
			if el < lo || el > hi {
				t.Errorf("%s: ended after %v, want within [%v, %v]", what, el, lo, hi)
			}
			if n := settled(func() int { return e.waiters(t, obs) }); n != 0 {
				t.Errorf("%s: %d lock waiters left", what, n)
			}
			if n := settled(openTx); n != 0 {
				t.Errorf("%s: %d runtime-role transactions left open", what, n)
			}
		}

		// (1) The target Group row held by an uncommitted model rename; the caller's 1 s deadline ends the wait.
		g := newGroup(t)
		before := onbTarget(t, g.ID)()
		hw, hp := holdBeforeCommit("Group-lock holder", func(m *tenancy.Service) error {
			_, err := m.RenameGroup(ctx, org, g.ID, g.Name+" Held")
			return err
		})
		early := guardOn(t, gsvc, owner, zeroG, func(c context.Context, tc *TenantContext) error {
			dc, cancel := context.WithTimeout(c, time.Second)
			defer cancel()
			return completeNew(gsvc, g)(dc, tc)
		})
		bounded("caller deadline at the Group lock", early, time.Second, completionDecisionBudget-time.Second, false)
		close(hp.release)
		if err := hw.wait(t, "Group-lock holder"); err != nil {
			t.Fatalf("the Group-lock holder could not commit after the bounded completion: %v", err)
		}
		if got := onbTarget(t, g.ID)(); got != before {
			t.Errorf("(1) after the bounded completion: %s, want %s", got, before)
		}
		if err := guardOn(t, gsvc, owner, zeroG, completeNew(gsvc, g))(); err != nil {
			t.Errorf("(1) a later completion of the same Group: %v", err)
		}
		ordering(t, map[string]any{"case": "O-ONB-T", "branch": "caller-deadline-at-group-lock", "outcome": "unavailable, then allowed"})

		// (2) The act write: an uncommitted association of the same Group (a model link paused before COMMIT) holds
		// the association key; the completion's locked facts do not see it, so the decision allows and the act
		// callback's association insert waits on the key until the decision budget ends it.
		pSel := newProfile(t)
		for _, mode := range []string{"new", "existing"} {
			g := newGroup(t)
			pHeld := newProfile(t)
			before := onbTarget(t, g.ID)()
			lw, lp := holdBeforeCommit("association holder", func(m *tenancy.Service) error {
				_, err := m.LinkGroupBusinessProfile(ctx, org, g.ID, pHeld.ID)
				return err
			})
			call := completeNew(gsvc, g)
			if mode == "existing" {
				call = completeExisting(gsvc, g, pSel.ID)
			}
			bounded("act write ("+mode+")", guardOn(t, gsvc, owner, zeroG, call), completionDecisionBudget,
				completionDecisionBudget+completionCleanupBudget, true)
			if got := onbTarget(t, g.ID)(); got != before {
				t.Errorf("(2 %s) with the holder still open: %s, want %s (no profile, no association)", mode, got, before)
			}
			close(lp.release)
			if err := lw.wait(t, "association holder"); err != nil {
				t.Fatalf("(2 %s) the association holder could not commit after the bounded completion: %v", mode, err)
			}
			var held int
			if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.group_business_profiles WHERE group_id = $1::text::uuid AND business_profile_id = $2::text::uuid`,
				g.ID.String(), pHeld.ID.String()).Scan(&held); err != nil || held != 1 {
				t.Errorf("(2 %s) the holder's association: %d (%v), want 1", mode, held, err)
			}
			ordering(t, map[string]any{"case": "O-ONB-T", "branch": "decision-budget-in-act-write", "mode": mode, "outcome": "unavailable, holder committed"})
		}
	})

	// --- G019 O-MEM (Plan r1 §6.2 as amended by the r2 addendum §4 and the r3 addendum C2) ---------------------------
	//
	// Expectations are the Plan's accepted outcomes, not a measured schedule: where a wait-for cycle is possible (the r2
	// §4 facade-versus-facade and facade-versus-direct cycles) either the two serialize or exactly one side aborts with a
	// class-40 error, which the facade answers unavailable (reason aborted, never retried); every branch asserts the
	// last-owner invariant and zero partial writes. The second actor may block behind the first or, where it takes no lock
	// the paused first holds, finish before the release; blockedOrDone records which happened.
	blockedOrDone := func(t *testing.T, a *actor) string {
		t.Helper()
		obs := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			if a.finished() {
				return "finished"
			}
			if e.waiters(t, obs) >= 1 {
				return "waiting"
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal("the second actor neither waited for a lock nor finished")
		return ""
	}
	fieldOf := func(err error) string {
		var ae *Error
		if errors.As(err, &ae) {
			return ae.Field
		}
		return ""
	}
	outcome := func(err error) string {
		if err == nil {
			return "allowed"
		}
		if f := fieldOf(err); f != "" {
			return string(CodeOf(err)) + ":" + f
		}
		return string(CodeOf(err))
	}
	// lostCall checks the losing facade call of a race: either a definitive refusal with one of the pinned (code, field)
	// pairs and its one deny record, or a class-40 abort attributed to that call: unavailable with no field and, among
	// the op's access.decision records written since the byte offset from (taken just before the race), exactly one
	// failed record, reason aborted. A generic unavailable (store failure, cancellation, commit_outcome_unknown) or an
	// earlier aborted record does not qualify. winnerLogged says whether the winner is a facade call of the same op (its
	// one allow record is then expected in the window too). A pinned pair with no field (roleRefused, the policy guard's
	// generic forbidden) also requires that one deny record to carry reason role_not_permitted. Only the closed
	// outcome/reason vocabulary is read.
	lostCall := func(t *testing.T, what string, lerr error, from int, op string, winnerLogged bool, refusals ...[2]string) string {
		t.Helper()
		count := map[string]int{}
		var failedReason, denyReason string
		for _, r := range e.recordsSince(t, from) {
			if r["operation"] != op {
				continue
			}
			o, _ := r["outcome"].(string)
			count[o]++
			if o == outcomeFailed {
				failedReason, _ = r["reason"].(string)
			}
			if o == outcomeDeny {
				denyReason, _ = r["reason"].(string)
			}
		}
		wantAllow := map[bool]int{true: 1, false: 0}[winnerLogged]
		if count[outcomeAllow] != wantAllow || count[outcomeDeny]+count[outcomeFailed] != 1 {
			t.Errorf("%s: %s records since the race began %v, want %d allow and exactly one record of the losing call", what, op, count, wantAllow)
		}
		if CodeOf(lerr) == CodeUnavailable {
			if f := fieldOf(lerr); f != "" || count[outcomeFailed] != 1 || failedReason != string(reasonAborted) {
				t.Errorf("%s: unavailable without attributable class-40 evidence (field %q, failed records %d, reason %q)",
					what, f, count[outcomeFailed], failedReason)
			}
			return "aborted"
		}
		for _, p := range refusals {
			if string(CodeOf(lerr)) == p[0] && fieldOf(lerr) == p[1] {
				if count[outcomeDeny] != 1 {
					t.Errorf("%s: refusal %s:%s without exactly one deny record (%v)", what, p[0], p[1], count)
				}
				if p[1] == "" && denyReason != string(reasonRoleNotPermitted) {
					t.Errorf("%s: refusal %s with no field whose deny record has reason %q, want %s", what, p[0], denyReason, reasonRoleNotPermitted)
				}
				return "refused " + p[0] + ":" + p[1]
			}
		}
		t.Errorf("%s: the losing call answered %v, want one of %v or an attributable class-40 abort", what, lerr, refusals)
		return "unexpected " + outcome(lerr)
	}
	lastOwner := [2]string{string(CodeConflict), string(tenancy.CodeLastOwner)}
	ownerAuthority := [2]string{string(CodeForbidden), string(tenancy.CodeOwnerAuthority)}
	// roleRefused is the policy guard's serialized refusal (Plan r2 §4(c)): once the winner has demoted the losing actor
	// to admin, the loser's DemoteOwner guard re-reads manage_owners and denies generically (forbidden, no field, reason
	// role_not_permitted) before any body owner check can run. Accepted for O-MEM-3b/3d only.
	roleRefused := [2]string{string(CodeForbidden), ""}
	// twoOwners returns a fresh Group owned by a (creator) and b, both plain Organization members with prepared credentials.
	twoOwners := func(t *testing.T) (tenancy.Group, *itUser, *itUser) {
		a := member(t, tenancy.OrgMember, nil)
		g := ownedGroup(t, a)
		b := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupOwner})
		return g, a, b
	}
	roleOf := func(t *testing.T, g tenancy.GroupID, u *itUser) string {
		var r string
		if err := e.setup.QueryRow(ctx, `SELECT coalesce((SELECT role FROM app.role_bindings WHERE scope_type = 'group' AND scope_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid), 'none')`, g.String(), u.id.String()).Scan(&r); err != nil {
			t.Fatal(err)
		}
		return r
	}
	// race runs first (paused before its step on hFirst) and second, and returns both errors and the seam observed.
	race := func(t *testing.T, hFirst *hookDB, step string, first, second func() error) (error, error, string) {
		t.Helper()
		pp := newPause()
		hFirst.on("before", step, pp.hook)
		a := start(first)
		pp.wait(t, "first")
		b := start(second)
		seam := blockedOrDone(t, b)
		close(pp.release)
		return a.wait(t, "first"), b.wait(t, "second"), seam
	}

	t.Run("O-MEM-1 two last owners leave concurrently at the lock seams", func(t *testing.T) {
		for _, firstA := range []bool{true, false} {
			g, a, b := twoOwners(t)
			sA, hA := e.service(e.id)
			sB, hB := e.service(e.id)
			leaveA := guardOn(t, sA, a, zeroG, func(c context.Context, tc *TenantContext) error { return sA.LeaveScope(c, tc, groupScope(g.ID)) })
			leaveB := guardOn(t, sB, b, zeroG, func(c context.Context, tc *TenantContext) error { return sB.LeaveScope(c, tc, groupScope(g.ID)) })
			first, second, hFirst, fu, su := leaveA, leaveB, hA, a, b
			if !firstA {
				first, second, hFirst, fu, su = leaveB, leaveA, hB, b, a
			}
			// The first pauses before its DELETE, after its facts; the second departs meanwhile.
			mark := len(e.logs.String())
			ef, es, seam := race(t, hFirst, "D", first, second)
			var winner, loser *itUser
			var lerr error
			switch {
			case ef == nil && es != nil:
				winner, loser, lerr = fu, su, es
			case es == nil && ef != nil:
				winner, loser, lerr = su, fu, ef
			default:
				t.Fatalf("first=%v: exactly one departure must succeed: %v / %v", firstA, ef, es)
			}
			lost := lostCall(t, fmt.Sprintf("first=%v", firstA), lerr, mark, opMembershipLeave, true, lastOwner)
			if got := ownersOf(t, g.ID)(); got != loser.id.String() {
				t.Errorf("first=%v: owners %q, want only the refused leaver %s (zero partial writes)", firstA, got, loser.id)
			}
			if r := roleOf(t, g.ID, winner); r != "none" {
				t.Errorf("first=%v: the departed owner still holds %s", firstA, r)
			}
			ordering(t, map[string]any{"case": "O-MEM-1", "first": map[bool]string{true: "A", false: "B"}[firstA], "second": seam,
				"outcome": "one left; other " + lost})
		}
	})

	t.Run("O-MEM-3a promotion and an ownership change on another Group of the same actor share one order", func(t *testing.T) {
		for _, firstA := range []bool{true, false} {
			actor := member(t, tenancy.OrgMember, nil)
			g1, g2 := ownedGroup(t, actor), ownedGroup(t, actor)
			r1 := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g1.ID: tenancy.GroupAdmin})
			r2 := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g2.ID: tenancy.GroupAdmin})
			sA, hA := e.service(e.id)
			sB, hB := e.service(e.id)
			promote := guardOn(t, sA, actor, zeroG, func(c context.Context, tc *TenantContext) error {
				_, err := sA.PromoteToOwner(c, tc, groupScope(g1.ID), r1.id)
				return err
			})
			handOver := guardOn(t, sB, actor, zeroG, transfer(sB, g2.ID, r2.id))
			first, second, hFirst := promote, handOver, hA
			if !firstA {
				first, second, hFirst = handOver, promote, hB
			}
			ef, es, seam := race(t, hFirst, "W", first, second)
			if ef != nil || es != nil || isDeadlock(ef) || isDeadlock(es) {
				t.Errorf("first=%v: %v / %v, want both allowed without a deadlock", firstA, ef, es)
			}
			if got := ownersOf(t, g1.ID)(); !strings.Contains(got, r1.id.String()) || !strings.Contains(got, actor.id.String()) {
				t.Errorf("first=%v: g1 owners %q", firstA, got)
			}
			if got := ownersOf(t, g2.ID)(); got != r2.id.String() || roleOf(t, g2.ID, actor) != string(tenancy.GroupMember) {
				t.Errorf("first=%v: g2 owners %q, actor %s", firstA, got, roleOf(t, g2.ID, actor))
			}
			ordering(t, map[string]any{"case": "O-MEM-3a", "first": map[bool]string{true: "promote", false: "transfer"}[firstA], "second": seam,
				"outcome": "both allowed, no deadlock"})
		}
	})

	t.Run("O-MEM-3b a direct runtime demotion racing a facade demotion on the same Group", func(t *testing.T) {
		for _, facadeFirst := range []bool{true, false} {
			g, a, b := twoOwners(t)
			svc, hook := e.service(e.id)
			demoteB := guardOn(t, svc, a, zeroG, func(c context.Context, tc *TenantContext) error {
				_, err := svc.DemoteOwner(c, tc, groupScope(g.ID), b.id, string(tenancy.GroupAdmin))
				return err
			})
			var bindingA string
			if err := e.setup.QueryRow(ctx, `SELECT id::text FROM app.role_bindings WHERE scope_type = 'group' AND scope_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid`, g.ID.String(), a.id.String()).Scan(&bindingA); err != nil {
				t.Fatal(err)
			}
			// The direct statement: b, a current owner, demotes a through the runtime role in b's own context (no facade).
			rt := e.rlsRuntime(t)
			direct := func(hold func()) error {
				tx, err := rt.Begin(ctx)
				if err != nil {
					return err
				}
				defer func() { _ = tx.Rollback(ctx) }()
				if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`,
					b.id.String(), org.String()); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `UPDATE app.role_bindings SET role = 'group_admin' WHERE id = $1::uuid`, bindingA); err != nil {
					return err
				}
				hold()
				return tx.Commit(ctx)
			}
			var ef, ed error
			var seam string
			mark := len(e.logs.String())
			if facadeFirst {
				ef, ed, seam = race(t, hook, "W", demoteB, func() error { return direct(func() {}) })
			} else {
				dp := newPause()
				d := start(func() error { return direct(dp.hook) })
				dp.wait(t, "direct demotion")
				f := start(demoteB)
				seam = blockedOrDone(t, f)
				close(dp.release)
				ed, ef = d.wait(t, "direct"), f.wait(t, "facade")
			}
			// Exactly one demotion commits (never both, never neither), and both exact binding states follow from which.
			var demoted, kept *itUser
			facadeOutcome := "allowed"
			var pg *pgconn.PgError
			switch {
			case ef == nil && ed != nil:
				demoted, kept = b, a
				// The direct statement's own SQLSTATE attributes it: a class-40 victim or the guard's definitive refusal.
				if !errors.As(ed, &pg) || (pg.Code != "40P01" && pg.Code != "40001" && pg.Code != "42501" && pg.Code != "23000") {
					t.Errorf("facadeFirst=%v: direct statement failed with %v", facadeFirst, ed)
				}
			case ed == nil && ef != nil:
				demoted, kept = a, b
				// The facade loser is a's guard denial (roleRefused) once the direct statement has demoted a, the
				// recount's last_owner, owner_authority, or an attributable class-40 abort.
				facadeOutcome = lostCall(t, fmt.Sprintf("facadeFirst=%v", facadeFirst), ef, mark, opOwnershipDemote, false, ownerAuthority, lastOwner, roleRefused)
			default:
				t.Fatalf("facadeFirst=%v: exactly one demotion must commit: facade %v / direct %v", facadeFirst, ef, ed)
			}
			if r := roleOf(t, g.ID, demoted); r != string(tenancy.GroupAdmin) {
				t.Errorf("facadeFirst=%v: the demoted owner holds %s, want %s", facadeFirst, r, tenancy.GroupAdmin)
			}
			if r := roleOf(t, g.ID, kept); r != string(tenancy.GroupOwner) {
				t.Errorf("facadeFirst=%v: the remaining owner holds %s, want %s (zero partial writes)", facadeFirst, r, tenancy.GroupOwner)
			}
			if owners := ownersOf(t, g.ID)(); owners != kept.id.String() {
				t.Errorf("facadeFirst=%v: owners %q, want only %s", facadeFirst, owners, kept.id)
			}
			ordering(t, map[string]any{"case": "O-MEM-3b", "first": map[bool]string{true: "facade", false: "direct"}[facadeFirst], "second": seam,
				"facade": facadeOutcome, "direct_deadlock": isDeadlock(ed), "direct_failed": ed != nil})
		}
	})

	t.Run("O-MEM-3c promotion racing DisableUser of the recipient", func(t *testing.T) {
		for _, branch := range []string{"disable-committed-first", "promotion-committed-first", "disable-in-flight"} {
			actor := member(t, tenancy.OrgMember, nil)
			g := ownedGroup(t, actor)
			rcp := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			svc, hook := e.service(e.id)
			promote := guardOn(t, svc, actor, zeroG, func(c context.Context, tc *TenantContext) error {
				_, err := svc.PromoteToOwner(c, tc, groupScope(g.ID), rcp.id)
				return err
			})
			disable := func() error { return idTraced.DisableUser(ctx, rcp.id.String()) }
			var perr error
			switch branch {
			case "disable-committed-first":
				if err := disable(); err != nil {
					t.Fatalf("%s: disable: %v", branch, err)
				}
				perr = promote()
				wantCode(t, branch, perr, CodeConflict)
			case "promotion-committed-first":
				// The promotion's write has taken the recipient's enabled-user share; DisableUser queues behind it.
				pp := newPause()
				hook.on("after", "W", pp.hook)
				p := start(promote)
				pp.wait(t, "promotion")
				d := start(disable)
				e.waitWaiters(t, 1)
				close(pp.release)
				perr = p.wait(t, "promotion")
				if err := d.wait(t, "disable"); err != nil || perr != nil || p.at > d.at {
					t.Errorf("%s: promotion %v, disable %v", branch, perr, err)
				}
			case "disable-in-flight":
				wp := tracer.arm("revoke_reason = 'disabled'")
				d := start(disable)
				wp.wait(t, "disable")
				p := start(promote)
				e.waitWaiters(t, 1)
				close(wp.release)
				if err := d.wait(t, "disable"); err != nil {
					t.Fatalf("%s: disable: %v", branch, err)
				}
				perr = p.wait(t, "promotion")
				wantCode(t, branch, perr, CodeConflict)
			}
			var status string
			if err := e.setup.QueryRow(ctx, `SELECT status FROM app.users WHERE id = $1::text::uuid`, rcp.id.String()).Scan(&status); err != nil || status != "disabled" {
				t.Errorf("%s: recipient status %q (%v)", branch, status, err)
			}
			want := string(tenancy.GroupAdmin)
			if branch == "promotion-committed-first" {
				want = string(tenancy.GroupOwner) // the owner is then disabled with ownership retained (C2)
			}
			if r := roleOf(t, g.ID, rcp); r != want {
				t.Errorf("%s: recipient role %s, want %s", branch, r, want)
			}
			ordering(t, map[string]any{"case": "O-MEM-3c", "branch": branch, "outcome": outcome(perr)})
		}
	})

	t.Run("O-MEM-3d two owners of one Group demote each other concurrently", func(t *testing.T) {
		for _, firstA := range []bool{true, false} {
			g, a, b := twoOwners(t)
			sA, hA := e.service(e.id)
			sB, hB := e.service(e.id)
			demote := func(s *Service, actor, subject *itUser) func() error {
				return guardOn(t, s, actor, zeroG, func(c context.Context, tc *TenantContext) error {
					_, err := s.DemoteOwner(c, tc, groupScope(g.ID), subject.id, string(tenancy.GroupAdmin))
					return err
				})
			}
			aDemotesB, bDemotesA := demote(sA, a, b), demote(sB, b, a)
			first, second, hFirst := aDemotesB, bDemotesA, hA
			if !firstA {
				first, second, hFirst = bDemotesA, aDemotesB, hB
			}
			mark := len(e.logs.String())
			ef, es, seam := race(t, hFirst, "W", first, second)
			var lerr error
			switch {
			case ef == nil && es != nil:
				lerr = es
			case es == nil && ef != nil:
				lerr = ef
			default:
				t.Fatalf("first=%v: exactly one demotion must commit: %v / %v", firstA, ef, es)
			}
			// Serialized: the later actor is no longer an owner, so its guard denies (roleRefused: forbidden, no field,
			// reason role_not_permitted), or it answers forbidden owner_authority, or the recount refuses (conflict
			// last_owner); a cycle: exactly one class-40 victim, answered unavailable with its aborted record.
			lost := lostCall(t, fmt.Sprintf("first=%v", firstA), lerr, mark, opOwnershipDemote, true, ownerAuthority, lastOwner, roleRefused)
			if owners := ownersOf(t, g.ID)(); owners == "" || strings.Contains(owners, ",") {
				t.Errorf("first=%v: owners %q, want exactly one (zero partial writes)", firstA, owners)
			}
			ordering(t, map[string]any{"case": "O-MEM-3d", "first": map[bool]string{true: "A", false: "B"}[firstA], "second": seam,
				"outcome": "one demoted; other " + lost})
		}
	})

	t.Run("O-MEM-DUP duplicate claims of one Group invitation by its recipient serialize", func(t *testing.T) {
		actor := member(t, tenancy.OrgMember, nil)
		g := ownedGroup(t, actor)
		rcp := member(t, tenancy.OrgMember, nil)
		isvc, _ := e.service(e.id)
		ictx, _ := beginAs(t, isvc, ctx, actor, org)
		itc := resolve(t, isvc, ictx, org, zeroG)
		iss, err := isvc.IssueInvitation(ictx, itc, groupScope(g.ID), string(tenancy.GroupMember), Recipient{UserID: &rcp.id}, 0)
		if err != nil {
			t.Fatalf("issue: %v", err)
		}
		s1, h1 := e.service(e.id)
		s2, _ := e.service(e.id)
		var o1, o2 AcceptOutcome
		claim1 := guardOn(t, s1, rcp, zeroG, func(c context.Context, tc *TenantContext) error {
			var err error
			o1, err = s1.AcceptGroupInvitation(c, tc, iss.Invitation.ID, iss.Token)
			return err
		})
		claim2 := guardOn(t, s2, rcp, zeroG, func(c context.Context, tc *TenantContext) error {
			var err error
			o2, err = s2.AcceptGroupInvitation(c, tc, iss.Invitation.ID, iss.Token)
			return err
		})
		// The first claim holds the invitation row (FOR NO KEY UPDATE inside the claim body) until its COMMIT; the
		// duplicate queues on that row and then reads the committed consumption.
		pp := newPause()
		h1.on("before", "COMMIT", pp.hook)
		a := start(claim1)
		pp.wait(t, "first claim")
		b := start(claim2)
		e.waitWaiters(t, 1)
		close(pp.release)
		e1, e2 := a.wait(t, "first claim"), b.wait(t, "duplicate claim")
		if e1 != nil || e2 != nil || o1.Outcome != "joined" || o2.Outcome != "already_member" || a.at > b.at {
			t.Errorf("claims: %q %v / %q %v, want joined then already_member", o1.Outcome, e1, o2.Outcome, e2)
		}
		var n int
		var consumedBy string
		if err := e.setup.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.role_bindings WHERE scope_type = 'group' AND scope_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid), (SELECT coalesce(consumed_by_user_id::text, '') FROM app.invitations WHERE id = $3::text::uuid)`,
			g.ID.String(), rcp.id.String(), iss.Invitation.ID.String()).Scan(&n, &consumedBy); err != nil || n != 1 || consumedBy != rcp.id.String() {
			t.Errorf("after the duplicate claim: %d bindings, consumed by %q (%v)", n, consumedBy, err)
		}
		ordering(t, map[string]any{"case": "O-MEM-DUP", "outcome": o1.Outcome + " then " + o2.Outcome})
	})

	// G019 writer25 (Codex writer23 review item 2, MEM-12/13): a Group acceptance racing the recipient's administrative
	// Organization removal, each order with the first actor paused before its COMMIT. The claim holds the recipient's
	// Organization binding FOR SHARE and the removal locks it FOR UPDATE, so the two serialize on that row: accept
	// first, the removal then departs the new Group binding together with the Organization binding in one transaction;
	// removal first, the acceptance then finds no Organization membership and is refused without consuming. Neither
	// order leaves a Group binding without an Organization binding (the invariant asserted in both branches).
	t.Run("O-MEM-RM a Group acceptance versus the recipient's Organization removal in both orders", func(t *testing.T) {
		actor := member(t, tenancy.OrgMember, nil)
		g := ownedGroup(t, actor)
		isvc, _ := e.service(e.id)
		ictx, _ := beginAs(t, isvc, ctx, actor, org)
		itc := resolve(t, isvc, ictx, org, zeroG)
		bindingsOf := func(t *testing.T, u *itUser) (all, dormant int) {
			t.Helper()
			if err := e.setup.QueryRow(ctx, `SELECT count(*), count(*) FILTER (WHERE b.scope_type = 'group' AND NOT EXISTS (SELECT 1 FROM app.role_bindings o
  WHERE o.organization_id = b.organization_id AND o.subject_user_id = b.subject_user_id AND o.scope_type = 'organization'))
  FROM app.role_bindings b WHERE b.organization_id = $1::text::uuid AND b.subject_user_id = $2::text::uuid`, org.String(), u.id.String()).Scan(&all, &dormant); err != nil {
				t.Fatal(err)
			}
			return
		}
		consumedBy := func(t *testing.T, id tenancy.InvitationID) string {
			t.Helper()
			var s string
			if err := e.setup.QueryRow(ctx, `SELECT coalesce(consumed_by_user_id::text, '') FROM app.invitations WHERE id = $1::text::uuid`, id.String()).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
		for _, acceptFirst := range []bool{true, false} {
			name := map[bool]string{true: "accept_first", false: "removal_first"}[acceptFirst]
			t.Run(name, func(t *testing.T) {
				rcp := member(t, tenancy.OrgMember, nil)
				iss, err := isvc.IssueInvitation(ictx, itc, groupScope(g.ID), string(tenancy.GroupMember), Recipient{UserID: &rcp.id}, 0)
				if err != nil {
					t.Fatalf("issue: %v", err)
				}
				s1, h1 := e.service(e.id)
				s2, h2 := e.service(e.id)
				var out AcceptOutcome
				accept := guardOn(t, s1, rcp, zeroG, func(c context.Context, tc *TenantContext) error {
					var err error
					out, err = s1.AcceptGroupInvitation(c, tc, iss.Invitation.ID, iss.Token)
					return err
				})
				remove := guardOn(t, s2, owner, zeroG, func(c context.Context, tc *TenantContext) error {
					return s2.RemoveMember(c, tc, Scope{Kind: tenancy.ScopeOrganization}, rcp.id)
				})
				var aerr, rerr error
				var seam string
				if acceptFirst {
					aerr, rerr, seam = race(t, h1, "COMMIT", accept, remove)
				} else {
					rerr, aerr, seam = race(t, h2, "COMMIT", remove, accept)
				}
				all, dormant := bindingsOf(t, rcp)
				by := consumedBy(t, iss.Invitation.ID)
				if dormant != 0 {
					t.Errorf("%s: %d Group bindings without an Organization binding", name, dormant)
				}
				if rerr != nil || all != 0 {
					t.Errorf("%s: removal %v, %d bindings remain", name, rerr, all)
				}
				if acceptFirst {
					if aerr != nil || out.Outcome != tenancy.ClaimJoined || by != rcp.id.String() {
						t.Errorf("%s: acceptance %v (%q), consumed by %q, want joined by the recipient before the removal", name, aerr, out.Outcome, by)
					}
				} else if c := CodeOf(aerr); (c != CodeContextDenied && c != CodeNotFound) || out.Outcome != "" || by != "" {
					t.Errorf("%s: acceptance %v (%q), consumed by %q, want a context_denied/not_found refusal with nothing consumed", name, aerr, out.Outcome, by)
				}
				ordering(t, map[string]any{"case": "O-MEM-RM", "first": map[bool]string{true: "accept", false: "remove"}[acceptFirst], "second": seam,
					"accept": outcome(aerr), "remove": outcome(rerr), "consumed": by != ""})
			})
		}
	})

	// G060 S2 amendment01 A3 and the attempt-2 release (A1.3 correction): the guard's Organization-departure cascade
	// against concurrent Group-side writers and the claim writer.
	orgBindings := func(t *testing.T, u *itUser) int {
		t.Helper()
		var n int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid`,
			org.String(), u.id.String()).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	removedOf := func(t *testing.T, u *itUser) int {
		t.Helper()
		var n int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = 'membership.removed'
  AND subject_user_id = $2::text::uuid`, org.String(), u.id.String()).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	consumedBy := func(t *testing.T, id tenancy.InvitationID) string {
		t.Helper()
		var s string
		if err := e.setup.QueryRow(ctx, `SELECT coalesce(consumed_by_user_id::text, '') FROM app.invitations WHERE id = $1::text::uuid`, id.String()).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	ownerSubjectRefusal := [2]string{string(CodeForbidden), fieldOwnerSubject}

	t.Run("O-MEM-4a Organization removal of a Group admin versus its promotion to Group owner", func(t *testing.T) {
		for _, promoteFirst := range []bool{true, false} {
			gowner := member(t, tenancy.OrgMember, nil)
			g := ownedGroup(t, gowner)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			sP, hP := e.service(e.id)
			sR, hR := e.service(e.id)
			promote := guardOn(t, sP, gowner, zeroG, func(c context.Context, tc *TenantContext) error {
				_, err := sP.PromoteToOwner(c, tc, groupScope(g.ID), u.id)
				return err
			})
			remove := guardOn(t, sR, owner, zeroG, func(c context.Context, tc *TenantContext) error {
				return sR.RemoveMember(c, tc, Scope{Kind: tenancy.ScopeOrganization}, u.id)
			})
			mark := len(e.logs.String())
			var perr, rerr error
			var seam, lost string
			if promoteFirst {
				perr, rerr, seam = race(t, hP, "COMMIT", promote, remove)
				if perr != nil {
					t.Fatalf("promotion first: %v", perr)
				}
				// The removal refuses with the Group-owner protection owner_subject: the cascade locks each Group binding
				// FOR NO KEY UPDATE before reading its role (attempt3, SR2-3), so it reads the committed owner row whether
				// the promotion committed before the cascade or while the cascade waited on that row; the statement rolls
				// back. (owner_authority, the earlier unlocked-read outcome, is no longer accepted.)
				lost = lostCall(t, "promotion first", rerr, mark, opMembershipRemove, false, ownerSubjectRefusal)
				if n, r := orgBindings(t, u), roleOf(t, g.ID, u); n != 2 || r != string(tenancy.GroupOwner) {
					t.Errorf("promotion first: %d bindings, Group role %s, want both intact with the promotion", n, r)
				}
			} else {
				rerr, perr, seam = race(t, hR, "COMMIT", remove, promote)
				if rerr != nil {
					t.Fatalf("removal first: %v", rerr)
				}
				if CodeOf(perr) != CodeNotFound {
					t.Errorf("removal first: promotion %v, want not_found (the Group binding departed with the Organization binding)", perr)
				}
				lost = outcome(perr)
				if n := orgBindings(t, u); n != 0 {
					t.Errorf("removal first: %d bindings remain", n)
				}
			}
			ordering(t, map[string]any{"case": "O-MEM-4a", "first": map[bool]string{true: "promote", false: "remove"}[promoteFirst], "second": seam,
				"outcome": lost})
		}
	})

	t.Run("O-MEM-4b Organization removal of a Group admin versus its own Group claim", func(t *testing.T) {
		for _, claimFirst := range []bool{true, false} {
			gowner := member(t, tenancy.OrgMember, nil)
			g, g2 := ownedGroup(t, gowner), ownedGroup(t, gowner)
			u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
			isvc, _ := e.service(e.id)
			ictx, _ := beginAs(t, isvc, ctx, gowner, org)
			itc := resolve(t, isvc, ictx, org, zeroG)
			iss, err := isvc.IssueInvitation(ictx, itc, groupScope(g2.ID), string(tenancy.GroupMember), Recipient{UserID: &u.id}, 0)
			if err != nil {
				t.Fatalf("issue: %v", err)
			}
			sA, hA := e.service(e.id)
			sR, hR := e.service(e.id)
			var out AcceptOutcome
			accept := guardOn(t, sA, u, zeroG, func(c context.Context, tc *TenantContext) error {
				var err error
				out, err = sA.AcceptGroupInvitation(c, tc, iss.Invitation.ID, iss.Token)
				return err
			})
			remove := guardOn(t, sR, owner, zeroG, func(c context.Context, tc *TenantContext) error {
				return sR.RemoveMember(c, tc, Scope{Kind: tenancy.ScopeOrganization}, u.id)
			})
			var aerr, rerr error
			var seam string
			if claimFirst {
				aerr, rerr, seam = race(t, hA, "COMMIT", accept, remove)
			} else {
				rerr, aerr, seam = race(t, hR, "COMMIT", remove, accept)
			}
			if rerr != nil || orgBindings(t, u) != 0 {
				t.Errorf("claimFirst=%v: removal %v, %d bindings remain (no dormant Group binding)", claimFirst, rerr, orgBindings(t, u))
			}
			by := consumedBy(t, iss.Invitation.ID)
			if claimFirst {
				// Both Group bindings (g and the claimed g2) depart with the Organization binding: three removed rows.
				if aerr != nil || out.Outcome != tenancy.ClaimJoined || by != u.id.String() || removedOf(t, u) != 3 {
					t.Errorf("claim first: claim %v (%q), consumed by %q, %d removed rows, want joined then three removed rows", aerr, out.Outcome, by, removedOf(t, u))
				}
			} else {
				c := CodeOf(aerr)
				if !(c == CodeContextDenied || c == CodeNotFound || (c == CodeConflict && fieldOf(aerr) == string(tenancy.CodeRecipientIneligible))) ||
					out.Outcome != "" || by != "" || removedOf(t, u) != 2 {
					t.Errorf("removal first: claim %v (%q), consumed by %q, %d removed rows, want a refusal with nothing consumed and two removed rows", aerr, out.Outcome, by, removedOf(t, u))
				}
			}
			ordering(t, map[string]any{"case": "O-MEM-4b", "first": map[bool]string{true: "claim", false: "remove"}[claimFirst], "second": seam,
				"claim": outcome(aerr), "remove": outcome(rerr), "consumed": by != ""})
		}
	})

	t.Run("O-MEM-4c raw Organization departure versus a Group claim: one bounded class-40 abort, nothing partial", func(t *testing.T) {
		// Codex A1.3 correction: the raw departure's lock order differs from the facade's. The raw DELETE locks u's
		// Organization binding (and, in the guard's cascade, its Group binding) and then, in the G058 departure trigger,
		// requests the Organization row FOR NO KEY UPDATE; a facade claim of u holds that row FOR SHARE (fact step O) and
		// next requests u's Organization binding FOR SHARE (step OB). The barriers force exactly that interleaving: one
		// transaction is a class-40 victim (40P01, or 40001), the other commits, nothing partial survives, the facade
		// answers once (no internal retry) and its step order is unchanged.
		gowner := member(t, tenancy.OrgMember, nil)
		g, g2 := ownedGroup(t, gowner), ownedGroup(t, gowner)
		u := member(t, tenancy.OrgMember, map[tenancy.GroupID]tenancy.GroupRole{g.ID: tenancy.GroupAdmin})
		isvc, _ := e.service(e.id)
		ictx, _ := beginAs(t, isvc, ctx, gowner, org)
		itc := resolve(t, isvc, ictx, org, zeroG)
		iss, err := isvc.IssueInvitation(ictx, itc, groupScope(g2.ID), string(tenancy.GroupMember), Recipient{UserID: &u.id}, 0)
		if err != nil {
			t.Fatalf("issue: %v", err)
		}
		var orgBinding string
		if err := e.setup.QueryRow(ctx, `SELECT id::text FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid
  AND scope_type = 'organization'`, org.String(), u.id.String()).Scan(&orgBinding); err != nil {
			t.Fatal(err)
		}
		s2, h2 := e.service(e.id)
		var out AcceptOutcome
		accept := guardOn(t, s2, u, zeroG, func(c context.Context, tc *TenantContext) error {
			var err error
			out, err = s2.AcceptGroupInvitation(c, tc, iss.Invitation.ID, iss.Token)
			return err
		})
		h2.reset()
		pp := newPause()
		h2.on("before", "OB", pp.hook)
		removedBefore := removedOf(t, u)
		mark := len(e.logs.String())
		a := start(accept)
		pp.wait(t, "the claim holding the Organization row FOR SHARE")
		rt := e.rlsRuntime(t)
		uid, oid := u.id.String(), org.String()
		raw := start(func() error {
			tx, err := rt.Begin(ctx)
			if err != nil {
				return err
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if err := rlsSet(ctx, tx, &uid, &oid, ptr(""), nil); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `DELETE FROM app.role_bindings WHERE id = $1::uuid`, orgBinding); err != nil {
				return err
			}
			return tx.Commit(ctx)
		})
		if seam := blockedOrDone(t, raw); seam != "waiting" {
			t.Fatalf("the raw departure did not wait behind the claim's Organization share lock: %s (%v)", seam, raw.wait(t, "raw"))
		}
		close(pp.release)
		aerr, rerr := a.wait(t, "claim"), raw.wait(t, "raw departure")
		victim := ""
		switch {
		case rerr != nil && aerr == nil:
			victim = "raw"
			if s := pgState(rerr); s != "40P01" && s != "40001" {
				t.Errorf("raw departure failed with %v (SQLSTATE %q), want a class-40 abort", rerr, s)
			}
			// The raw transaction rolled back entirely; the claim committed: Organization, g and g2 bindings, no removal.
			if n, r := orgBindings(t, u), removedOf(t, u); n != 3 || r != removedBefore || out.Outcome != tenancy.ClaimJoined ||
				consumedBy(t, iss.Invitation.ID) != u.id.String() {
				t.Errorf("raw victim: %d bindings, removed %d->%d, claim %q, want 3 bindings, no removed row, joined", n, removedBefore, r, out.Outcome)
			}
		case aerr != nil && rerr == nil:
			victim = "claim"
			lostCall(t, "claim victim", aerr, mark, opInvitationAccept, false)
			// The claim rolled back entirely; the raw departure committed: no binding, g and the Organization removed.
			if n, r := orgBindings(t, u), removedOf(t, u); n != 0 || r != removedBefore+2 || consumedBy(t, iss.Invitation.ID) != "" {
				t.Errorf("claim victim: %d bindings, removed %d->%d, invitation consumed by %q, want 0, +2, unconsumed", n, removedBefore, r,
					consumedBy(t, iss.Invitation.ID))
			}
		default:
			t.Fatalf("exactly one transaction must be a class-40 victim: claim %v / raw %v", aerr, rerr)
		}
		// No internal retry: exactly one access.decision record of the claim since the race began.
		if recs := e.memOpRecords(t, mark, opInvitationAccept); len(recs) != 1 {
			t.Errorf("claim access.decision records %v, want exactly one (no internal retry)", recs)
		}
		// The facade's step order is unchanged: the Organization row (O) before the own Organization binding (OB).
		_, stmts := h2.counts()
		io, iob := slices.Index(stmts, "O"), slices.Index(stmts, "OB")
		if iob < 0 || (io >= 0 && io > iob) {
			t.Errorf("claim steps %v: the Organization binding lock must follow the Organization row", stmts)
		}
		ordering(t, map[string]any{"case": "O-MEM-4c", "victim": victim, "raw": stateOrUnknown(rerr), "claim": outcome(aerr),
			"outcome": "one bounded class-40 abort, no partial binding or audit, no retry"})
	})
}

// rotate signs subject u in again through id at origin (G016 A2: the origin live was issued at) while live is its live
// credential: an additional provider session supersedes the live one (sign-in rotation). first, when set, changes the
// display name. It returns the new credential or the callback's error (never a test failure: a deadlock victim is an
// expected outcome of O-15).
func (e *itEnv) rotate(id *identity.Service, u *itUser, first, live, origin string) (string, error) {
	ctx := context.Background()
	out, err := id.Login(ctx, identity.LoginInput{Origin: origin, ReturnTo: "/account"})
	if err != nil {
		return "", err
	}
	au, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		return "", err
	}
	if first == "" {
		first = "Again"
	}
	code := e.fake.IssueCode(au.Query().Get("code_challenge"), identitytest.Grant{UserID: u.sub, FirstName: first, LastName: "Fixture",
		SID: fmt.Sprintf("session_01ROT%08d", itSeq.Add(1))})
	cb, err := id.Callback(ctx, identity.CallbackInput{Origin: origin, RawQuery: "code=" + url.QueryEscape(code) + "&state=" +
		url.QueryEscape(au.Query().Get("state")), Binding: out.Binding, Credential: live})
	if err != nil {
		return "", err
	}
	return cb.Credential, nil
}
