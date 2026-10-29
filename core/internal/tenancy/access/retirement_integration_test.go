//go:build integration

package access

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/tenancy"
)

// MEM-14 (G019 Plan r1 §6.2 as extended by Result addendum A1-01, D-A1): the trusted RetireUser and its self-only
// Organization enumeration app.rls_own_organizations. Pending (NOT EXECUTED) while the native test hold applies.

// ownOrganizations reads every page of app.rls_own_organizations in one runtime transaction in the given context (a zero
// user: no context statement at all).
func (e *itEnv) ownOrganizations(t *testing.T, user tenancy.UserID, org tenancy.OrganizationID, limit int) []tenancy.OrganizationID {
	t.Helper()
	ctx := context.Background()
	tx, err := e.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if !zero(user) {
		if err := tenancy.InitContext(ctx, tx, user, org, tenancy.GroupID{}); err != nil {
			t.Fatalf("context: %v", err)
		}
	}
	var out []tenancy.OrganizationID
	var after tenancy.OrganizationID
	for {
		page, err := tenancy.OwnOrganizationsPage(ctx, tx, after, limit)
		if err != nil {
			t.Fatalf("page: %v", err)
		}
		out = append(out, page...)
		if len(page) < limit {
			return out
		}
		after = page[len(page)-1]
	}
}

// rawOwnOrganizations calls the reader directly with an out-of-range limit in u's IDENTITY-shape context.
func (e *itEnv) rawOwnOrganizations(t *testing.T, user tenancy.UserID, limit int) int {
	t.Helper()
	ctx := context.Background()
	tx, err := e.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := tenancy.InitContext(ctx, tx, user, tenancy.OrganizationID{}, tenancy.GroupID{}); err != nil {
		t.Fatalf("context: %v", err)
	}
	var n int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.rls_own_organizations(NULL, $1)`, int32(limit)).Scan(&n); err != nil {
		t.Fatalf("raw reader: %v", err)
	}
	return n
}

func (e *itEnv) userStatus(t *testing.T, u *itUser) string {
	t.Helper()
	var st string
	if err := e.setup.QueryRow(context.Background(), `SELECT status FROM app.users WHERE id = $1`, u.id.String()).Scan(&st); err != nil {
		t.Fatalf("user status: %v", err)
	}
	return st
}

func sortedOrgs(orgs []tenancy.OrganizationID) []tenancy.OrganizationID {
	out := slices.Clone(orgs)
	slices.SortFunc(out, func(a, b tenancy.OrganizationID) int { return slices.Compare(a[:], b[:]) })
	return out
}

// mem14OwnOrganizationsReader: the reader answers only the context subject's own Organizations, only in the
// IDENTITY shape, only for limits 1..100, in ascending keyset pages; never another subject's memberships.
func mem14OwnOrganizationsReader(t *testing.T) {
	e := setupIT(t)
	u, v := e.newUser(t), e.newUser(t)
	a, b, c := e.org(t, "Retire A"), e.org(t, "Retire B"), e.org(t, "Foreign C")
	e.grantOrg(t, a, u, tenancy.OrgMember)
	e.grantOrg(t, b, u, tenancy.OrgAdmin)
	e.grantOrg(t, c, v, tenancy.OrgMember)
	e.grantOrg(t, a, v, tenancy.OrgOwner)

	if got, want := e.ownOrganizations(t, u.id, tenancy.OrganizationID{}, 1), sortedOrgs([]tenancy.OrganizationID{a, b}); !slices.Equal(got, want) {
		t.Errorf("self (pages of 1): %d orgs, want exactly A and B ascending", len(got))
	}
	if got := e.ownOrganizations(t, v.id, tenancy.OrganizationID{}, 100); slices.Contains(got, b) || len(got) != 2 {
		t.Errorf("foreign subject sees %d orgs (B included: %t), want only its own two", len(got), slices.Contains(got, b))
	}
	if got := e.ownOrganizations(t, tenancy.UserID{}, tenancy.OrganizationID{}, 100); len(got) != 0 {
		t.Errorf("no context: %d rows, want none", len(got))
	}
	if got := e.ownOrganizations(t, u.id, a, 100); len(got) != 0 {
		t.Errorf("Organization context: %d rows, want none (IDENTITY shape only)", len(got))
	}
	for _, limit := range []int{0, 101} {
		if n := e.rawOwnOrganizations(t, u.id, limit); n != 0 {
			t.Errorf("limit %d: %d rows, want none", limit, n)
		}
	}
}

// mem14RetireUserPages: a subject in more Organizations than one page holds is departed from every one of them
// (ascending, no silent cap), then disabled.
func mem14RetireUserPages(t *testing.T) {
	e := setupIT(t)
	u := e.newUser(t)
	var orgs []tenancy.OrganizationID
	for i := range retirementPage + 3 {
		o := e.org(t, fmt.Sprintf("Retire page %03d", i))
		e.grantOrg(t, o, u, tenancy.OrgMember)
		orgs = append(orgs, o)
	}
	p, err := e.svc.RetireUser(context.Background(), u.id)
	if err != nil {
		t.Fatalf("retire: %v", err)
	}
	if !slices.Equal(p.Departed, sortedOrgs(orgs)) || len(p.Remaining) != 0 || len(p.AlreadyDeparted) != 0 || !p.Disabled {
		t.Errorf("progress: %d departed, %d remaining, %d already departed, disabled %t; want all %d departed ascending",
			len(p.Departed), len(p.Remaining), len(p.AlreadyDeparted), p.Disabled, len(orgs))
	}
	if got := e.ownOrganizations(t, u.id, tenancy.OrganizationID{}, 100); len(got) != 0 {
		t.Errorf("%d memberships remain", len(got))
	}
	if st := e.userStatus(t, u); st != "disabled" {
		t.Errorf("status %s, want disabled", st)
	}
}

// mem14RetireUserStopsAndResumes: a sole-owner refusal preflights before any write; a blocker injected after the
// first departure stops the call with Departed=[first], Remaining=[second] and the user still enabled; after
// succession a re-run completes from the remaining list and disables the user.
func mem14RetireUserStopsAndResumes(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	u, heir := e.newUser(t), e.newUser(t)
	x, y := e.org(t, "Retire X"), e.org(t, "Retire Y")
	first, second := sortedOrgs([]tenancy.OrganizationID{x, y})[0], sortedOrgs([]tenancy.OrganizationID{x, y})[1]
	e.grantOrg(t, first, u, tenancy.OrgMember)
	e.grantOrg(t, second, u, tenancy.OrgOwner)

	// Sole owner of second: refused before any write; every Organization remains; still enabled.
	p, err := e.svc.RetireUser(ctx, u.id)
	var ae *Error
	if !errors.As(err, &ae) || ae.Code != CodeConflict || ae.Field != fieldRetirementBlocked || p.Blockers.Count != 1 ||
		len(p.Departed) != 0 || !slices.Equal(p.Remaining, []tenancy.OrganizationID{first, second}) {
		t.Fatalf("blocked: %v, progress %+v", err, p)
	}
	if st := e.userStatus(t, u); st != "active" {
		t.Fatalf("blocked status %s, want active", st)
	}

	// Succession: heir becomes a co-owner, so the preflight passes; then a Group of second solely owned by u is created
	// after the first departure's transaction began, so second's departure is refused by the guard (last_owner).
	e.grantOrg(t, second, heir, tenancy.OrgOwner)
	var g tenancy.Group
	begins := 0
	e.hook.on("before", "BEGIN", func() {
		if begins++; begins == 2 { // 1: preflight, 2: first departure
			g = e.group(t, second, u, "Late blocker", "late-blocker")
		}
	})
	p, err = e.svc.RetireUser(ctx, u.id)
	e.hook.on("before", "BEGIN", nil)
	if !errors.As(err, &ae) || ae.Code != CodeConflict || ae.Field != string(tenancy.CodeLastOwner) ||
		!slices.Equal(p.Departed, []tenancy.OrganizationID{first}) || !slices.Equal(p.Remaining, []tenancy.OrganizationID{second}) || p.Disabled {
		t.Fatalf("stopped: %v, progress %+v", err, p)
	}
	if st := e.userStatus(t, u); st != "active" {
		t.Fatalf("stopped status %s, want active", st)
	}

	// Group succession, then the resumed call departs only the remaining Organization and disables.
	e.grantGroup(t, second, g.ID, heir, tenancy.GroupOwner)
	p, err = e.svc.RetireUser(ctx, u.id)
	if err != nil || !slices.Equal(p.Departed, []tenancy.OrganizationID{second}) || len(p.Remaining) != 0 || !p.Disabled {
		t.Fatalf("resumed: %v, progress %+v", err, p)
	}
	if st := e.userStatus(t, u); st != "disabled" {
		t.Errorf("resumed status %s, want disabled", st)
	}
}

// G019 F-01 (MEM-14 extensions): identity's finalization rereads the remaining memberships and sole-owned scopes under
// the subject's users row FOR UPDATE. Late grants are injected deterministically through the access pool's BEGIN hook
// of the last departure (committed on the setup pool before the finalization starts on identity's pool).

// mem14OnLastDeparture runs fn (once) when the last departure's transaction begins: begin 1 is the preflight, begins
// 2..n+1 are the n departures of u's current Organizations.
func (e *itEnv) mem14OnLastDeparture(t *testing.T, u *itUser, fn func()) {
	t.Helper()
	last, begins := 1+len(e.ownOrganizations(t, u.id, tenancy.OrganizationID{}, 100)), 0
	e.hook.on("before", "BEGIN", func() {
		if begins++; begins == last {
			fn()
		}
	})
}

// mem14WantBlocked asserts a final refusal: conflict/retirement_blocked, no disable, the given complete Remaining and
// blocker count, the user still active.
func (e *itEnv) mem14WantBlocked(t *testing.T, u *itUser, p RetirementProgress, err error, remaining []tenancy.OrganizationID, blockers int) {
	t.Helper()
	var ae *Error
	if !errors.As(err, &ae) || ae.Code != CodeConflict || ae.Field != fieldRetirementBlocked || p.Disabled ||
		!slices.Equal(p.Remaining, sortedOrgs(remaining)) || p.Blockers.Count != blockers {
		t.Fatalf("final refusal: %v, progress %+v", err, p)
	}
	if st := e.userStatus(t, u); st != "active" {
		t.Fatalf("status %s after a final refusal, want active", st)
	}
}

// mem14FinalizeNewMembership: a membership of a new Organization C committed after the enumeration refuses the
// finalization (Remaining = [C], the departures kept, still enabled, sessions intact); a re-run departs C and disables.
func mem14FinalizeNewMembership(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	u := e.newUser(t)
	a, c := e.org(t, "Final A"), e.org(t, "Final C")
	e.grantOrg(t, a, u, tenancy.OrgMember)
	e.mem14OnLastDeparture(t, u, func() { e.grantOrg(t, c, u, tenancy.OrgMember) })
	p, err := e.svc.RetireUser(ctx, u.id)
	e.hook.on("before", "BEGIN", nil)
	e.mem14WantBlocked(t, u, p, err, []tenancy.OrganizationID{c}, 0)
	if !slices.Contains(p.Departed, a) {
		t.Errorf("departed %v, want A kept departed", p.Departed)
	}
	if _, err := e.id.Session(ctx, u.cred, itOrigin); err != nil {
		t.Errorf("session after a final refusal: %v, want intact", err)
	}
	p, err = e.svc.RetireUser(ctx, u.id)
	if err != nil || !slices.Equal(p.Departed, []tenancy.OrganizationID{c}) || !p.Disabled {
		t.Fatalf("re-run: %v, progress %+v", err, p)
	}
}

// mem14FinalizeRejoin: readmission to an already departed Organization refuses the finalization with Remaining = [it];
// a re-run completes.
func mem14FinalizeRejoin(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	u := e.newUser(t)
	x, y := e.org(t, "Rejoin X"), e.org(t, "Rejoin Y")
	e.grantOrg(t, x, u, tenancy.OrgMember)
	e.grantOrg(t, y, u, tenancy.OrgMember)
	first := e.ownOrganizations(t, u.id, tenancy.OrganizationID{}, 100)[0] // departed and committed before the last begins
	e.mem14OnLastDeparture(t, u, func() { e.grantOrg(t, first, u, tenancy.OrgMember) })
	p, err := e.svc.RetireUser(ctx, u.id)
	e.hook.on("before", "BEGIN", nil)
	e.mem14WantBlocked(t, u, p, err, []tenancy.OrganizationID{first}, 0)
	if p, err = e.svc.RetireUser(ctx, u.id); err != nil || !p.Disabled {
		t.Fatalf("re-run: %v, progress %+v", err, p)
	}
}

// mem14FinalizeSoleOwner: a sole ownership committed after the enumeration refuses the finalization with the locked
// blocker and Remaining = [B]; a re-run refuses at the preflight with nothing departed; after a co-owner is added the
// re-run completes.
func mem14FinalizeSoleOwner(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	u, heir := e.newUser(t), e.newUser(t)
	a, b := e.org(t, "Owner A"), e.org(t, "Owner B")
	e.grantOrg(t, a, u, tenancy.OrgMember)
	e.mem14OnLastDeparture(t, u, func() { e.grantOrg(t, b, u, tenancy.OrgOwner) })
	p, err := e.svc.RetireUser(ctx, u.id)
	e.hook.on("before", "BEGIN", nil)
	e.mem14WantBlocked(t, u, p, err, []tenancy.OrganizationID{b}, 1)
	p, err = e.svc.RetireUser(ctx, u.id)
	var ae *Error
	if !errors.As(err, &ae) || ae.Field != fieldRetirementBlocked || p.Blockers.Count != 1 || len(p.Departed) != 0 || p.Disabled {
		t.Fatalf("preflight refusal: %v, progress %+v", err, p)
	}
	e.grantOrg(t, b, heir, tenancy.OrgOwner)
	if p, err = e.svc.RetireUser(ctx, u.id); err != nil || !slices.Equal(p.Departed, []tenancy.OrganizationID{b}) || !p.Disabled {
		t.Fatalf("after succession: %v, progress %+v", err, p)
	}
}

// mem14FinalizeWaits: both lock orders. (1) An uncommitted binding insert of u (its FK check holds FOR KEY SHARE on u's
// users row) makes the finalization wait; once it commits the finalization refuses with Remaining = [D]. (2) The
// finalization holds u's row FOR UPDATE and is held at its sessions UPDATE by a session-row lock; an Organization claim
// by u then waits on the users row, and after the finalization commits the claim fails recipient_ineligible with no
// binding written.
func mem14FinalizeWaits(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	owner := e.memOwnerService(t)
	u, admin := e.newUser(t), e.newUser(t)
	d := e.org(t, "Wait D")

	x, err := e.setup.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = x.Rollback(ctx) }()
	if _, err := x.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::text::uuid, $2::text::uuid, 'organization', $1::text::uuid, NULL, 'org_member', 'membership')`, d.String(), u.id.String()); err != nil {
		t.Fatal(err)
	}
	type result struct {
		p   RetirementProgress
		err error
	}
	done := make(chan result, 1)
	go func() { p, err := e.svc.RetireUser(ctx, u.id); done <- result{p, err} }()
	e.waitWaiters(t, 1)
	if err := x.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	r := <-done
	e.mem14WantBlocked(t, u, r.p, r.err, []tenancy.OrganizationID{d}, 0)

	// (2) Depart D first so the finalization proceeds to its writes.
	e.grantOrg(t, d, admin, tenancy.OrgOwner)
	e.bind(t, admin, d) // G016 A2: the issuer's credential for D, setup only (no lock held)
	addr := memAddr("waiter")
	iss := e.memIssueOrg(t, admin, d, addr)
	ev := e.memAttest(t, u.sub, addr)
	y, err := e.setup.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = y.Rollback(ctx) }()
	if _, err := y.Exec(ctx, `SELECT 1 FROM app.sessions WHERE user_id = $1::text::uuid FOR SHARE`, u.id.String()); err != nil {
		t.Fatal(err)
	}
	go func() { p, err := e.svc.RetireUser(ctx, u.id); done <- result{p, err} }()
	e.waitWaiters(t, 1)
	claimed := make(chan error, 1)
	go func() {
		_, err := owner.ConsumeOrganizationInvitation(ctx, ev, iss.Invitation.ID, iss.Token)
		claimed <- err
	}()
	e.waitWaiters(t, 2)
	if err := y.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if r := <-done; r.err != nil || !r.p.Disabled {
		t.Fatalf("finalization: %v, progress %+v", r.err, r.p)
	}
	memWantCode(t, "claim after the disable", <-claimed, CodeConflict, string(tenancy.CodeRecipientIneligible))
	if n := e.memOrgBindings(t, d, u.id); n != 0 {
		t.Errorf("claim after the disable wrote %d bindings", n)
	}
}
