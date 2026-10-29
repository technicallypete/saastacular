package access

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/identity"
	"core/internal/tenancy"
)

// U-MEM (G019 Plan r1 §6.2, the unit half of MEM-2..MEM-14): the pure parts of the membership, ownership, invitation and
// retirement operations. The database halves (the guard, the claim bodies, RLS) are the CT-MEM-1/MEM integration rows.

func TestMembershipScope(t *testing.T) {
	g := g1
	zeroG := tenancy.GroupID{}
	for name, c := range map[string]struct {
		sc    Scope
		valid bool
	}{
		"organization":            {Scope{Kind: tenancy.ScopeOrganization}, true},
		"organization with Group": {Scope{Kind: tenancy.ScopeOrganization, Group: &g}, false},
		"group":                   {Scope{Kind: tenancy.ScopeGroup, Group: &g}, true},
		"group without Group":     {Scope{Kind: tenancy.ScopeGroup}, false},
		"group with zero Group":   {Scope{Kind: tenancy.ScopeGroup, Group: &zeroG}, false},
		"unknown kind":            {Scope{Kind: "tenant"}, false},
		"zero":                    {Scope{}, false},
	} {
		if got := c.sc.valid(); got != c.valid {
			t.Errorf("%s: valid = %t, want %t", name, got, c.valid)
		}
		if !c.valid {
			if f := scopePre(c.sc, nil)(); f == nil || f.code != CodeInvalidInput || f.field != fieldScope {
				t.Errorf("%s: scopePre = %+v, want invalid_input scope", name, f)
			}
		}
	}
	org, grp := Scope{Kind: tenancy.ScopeOrganization}, Scope{Kind: tenancy.ScopeGroup, Group: &g}
	for _, c := range []struct {
		sc                     Scope
		members, owners, leave Capability
		owner, admin           string
		target                 Target
		scopeID                [16]byte
	}{
		{org, capOrganizationManageMembers, capOrganizationManageOwners, capOrganizationLeave, "org_owner", "org_admin",
			Target{Kind: kindOrganization}, orgA},
		{grp, capGroupManageMembers, capGroupManageOwners, capGroupLeave, "group_owner", "group_admin",
			Target{Kind: kindGroup, Group: g1}, g1},
	} {
		if c.sc.manageMembers() != c.members || c.sc.manageOwners() != c.owners || c.sc.leave() != c.leave {
			t.Errorf("%s: capabilities %s/%s/%s", c.sc.Kind, c.sc.manageMembers(), c.sc.manageOwners(), c.sc.leave())
		}
		if c.sc.ownerRole() != c.owner || c.sc.adminRole() != c.admin {
			t.Errorf("%s: roles %s/%s", c.sc.Kind, c.sc.ownerRole(), c.sc.adminRole())
		}
		if c.sc.target() != c.target || c.sc.scopeID(orgA) != c.scopeID {
			t.Errorf("%s: target %+v scope %x", c.sc.Kind, c.sc.target(), c.sc.scopeID(orgA))
		}
		for _, k := range []Capability{c.members, c.owners, c.leave} {
			if kind, _ := capabilityKind(k); kind != c.target.Kind {
				t.Errorf("%s: kind %s, want %s", k, kind, c.target.Kind)
			}
		}
	}
	// Ordinary roles are admin/member of the scope only: never an owner role, never the other scope's roles.
	for _, c := range []struct {
		sc   Scope
		role string
		ok   bool
	}{
		{org, "org_admin", true}, {org, "org_member", true}, {org, "org_owner", false}, {org, "group_admin", false},
		{grp, "group_admin", true}, {grp, "group_member", true}, {grp, "group_owner", false}, {grp, "org_member", false},
		{grp, "", false}, {org, roleNone, false},
	} {
		if got := ordinaryRole(c.sc, c.role); got != c.ok {
			t.Errorf("ordinaryRole(%s, %q) = %t", c.sc.Kind, c.role, got)
		}
	}
}

// TestMembershipCapabilities pins the four G019 capabilities directly (the oracle's tables restate them): Group
// ownership is the exact Group's owner only, with no Organization-owner override; leaving needs a binding to leave.
func TestMembershipCapabilities(t *testing.T) {
	owner := groupBinding(g1, tenancy.GroupOwner)
	admin := groupBinding(g1, tenancy.GroupAdmin)
	member := groupBinding(g1, tenancy.GroupMember)
	wantDecision(t, "org_owner manages owners", capOrganizationManageOwners, orgOnly, factsA(tenancy.OrgOwner), targetOrgA, true, reasonAllowed)
	wantDecision(t, "org_admin does not", capOrganizationManageOwners, orgOnly, factsA(tenancy.OrgAdmin), targetOrgA, false, reasonRoleNotPermitted)
	wantDecision(t, "org_owner in selected context", capOrganizationManageOwners, selG1, factsA(tenancy.OrgOwner), targetOrgA, true, reasonAllowed)
	for _, r := range []tenancy.OrganizationRole{tenancy.OrgOwner, tenancy.OrgAdmin} {
		wantDecision(t, string(r)+" has no Group ownership override", capGroupManageOwners, orgOnly, factsA(r), targetG(g1), false, reasonRoleNotPermitted)
		wantDecision(t, string(r)+" without a Group binding cannot leave it", capGroupLeave, orgOnly, factsA(r), targetG(g1), false, reasonRoleNotPermitted)
		wantDecision(t, string(r)+" leaves the Organization", capOrganizationLeave, orgOnly, factsA(r), targetOrgA, true, reasonAllowed)
	}
	wantDecision(t, "org_member leaves the Organization", capOrganizationLeave, orgOnly, factsA(tenancy.OrgMember), targetOrgA, true, reasonAllowed)
	wantDecision(t, "group_owner manages own owners", capGroupManageOwners, orgOnly, factsA(tenancy.OrgMember, owner), targetG(g1), true, reasonAllowed)
	wantDecision(t, "group_admin does not", capGroupManageOwners, orgOnly, factsA(tenancy.OrgMember, admin), targetG(g1), false, reasonRoleNotPermitted)
	wantDecision(t, "group_owner of g1 not on g3", capGroupManageOwners, orgOnly, factsA(tenancy.OrgMember, owner), targetG(g3), false, reasonRoleNotPermitted)
	wantDecision(t, "org_owner with g1 ownership", capGroupManageOwners, orgOnly, factsA(tenancy.OrgOwner, owner), targetG(g1), true, reasonAllowed)
	wantDecision(t, "group_owner outside the selected Group", capGroupManageOwners, selG3, factsA(tenancy.OrgMember, owner, groupBinding(g3, tenancy.GroupMember)),
		targetG(g1), false, reasonTargetOutsideContext)
	for name, b := range map[string]tenancy.BindingFact{"owner": owner, "admin": admin, "member": member} {
		wantDecision(t, "group "+name+" leaves", capGroupLeave, orgOnly, factsA(tenancy.OrgMember, b), targetG(g1), true, reasonAllowed)
	}
	wantDecision(t, "Group role without the Organization binding", capGroupLeave, orgOnly, factsA("", member), targetG(g1), false, reasonOrganizationNotPermitted)
}

func TestMembershipPageInput(t *testing.T) {
	for _, c := range []struct {
		in, want int
		ok       bool
	}{{0, tenancy.MaxMembersPage, true}, {1, 1, true}, {100, 100, true}, {101, 0, false}, {-1, 0, false}} {
		got, f := pageLimit(c.in)
		if (f == nil) != c.ok || got != c.want || (f != nil && f.field != fieldLimit) {
			t.Errorf("pageLimit(%d) = %d, %+v", c.in, got, f)
		}
	}
	if id, f := cursor("", tenancy.ParseUserID); f != nil || !zero(id) {
		t.Errorf("empty cursor = %v, %+v", id, f)
	}
	for _, bad := range []string{"x", "00000000-0000-0000-0000-000000000000", " " + g1.String()} {
		if _, f := cursor(bad, tenancy.ParseUserID); f == nil || f.code != CodeInvalidInput || f.field != fieldCursor {
			t.Errorf("cursor(%q) = %+v, want invalid_input after", bad, f)
		}
	}
	u := uid[tenancy.UserID](0x33)
	if id, f := cursor(u.String(), tenancy.ParseUserID); f != nil || id != u {
		t.Errorf("cursor(%s) = %v, %+v", u, id, f)
	}
}

func TestMemberHints(t *testing.T) {
	g := g1
	grp := Scope{Kind: tenancy.ScopeGroup, Group: &g}
	actor, other := uid[tenancy.UserID](0x01), uid[tenancy.UserID](0x02)
	row := func(u tenancy.UserID, role string, disabled bool) tenancy.Member {
		return tenancy.Member{UserID: u, Role: role, Disabled: disabled}
	}
	for name, c := range map[string]struct {
		r                   tenancy.Member
		owners, actorOwner  bool
		count               int
		change, remove      bool
		promote, demote, tr bool
	}{
		"admin, owner actor":      {row(other, "group_admin", false), true, true, 1, true, true, true, false, true},
		"disabled admin":          {row(other, "group_admin", true), true, true, 1, true, true, false, false, false},
		"member":                  {row(other, "group_member", false), true, true, 1, true, true, false, false, false},
		"other owner, two owners": {row(other, "group_owner", false), true, true, 2, false, false, false, true, false},
		"other owner, last owner": {row(other, "group_owner", false), true, true, 1, false, false, false, false, false},
		"self":                    {row(actor, "group_owner", false), true, true, 2, false, false, false, true, false},
		"admin, admin actor":      {row(other, "group_admin", false), false, false, 1, true, true, false, false, false},
		"admin, Organization MGR": {row(other, "group_admin", false), false, false, 0, true, true, false, false, false},
	} {
		m := memberHints(grp, actor, c.r, c.owners, c.actorOwner, c.count)
		if m.CanChangeRole != c.change || m.CanRemove != c.remove || m.CanPromote != c.promote || m.CanDemote != c.demote ||
			m.CanTransfer != c.tr || m.Self != (c.r.UserID == actor) || m.Disabled != c.r.Disabled || m.Role != c.r.Role {
			t.Errorf("%s: hints %+v", name, m)
		}
	}
}

func TestSubjectBindingAndContext(t *testing.T) {
	g := g1
	org, grp := Scope{Kind: tenancy.ScopeOrganization}, Scope{Kind: tenancy.ScopeGroup, Group: &g}
	u, v := uid[tenancy.UserID](0x01), uid[tenancy.UserID](0x02)
	ob := tenancy.Binding{ID: uid[tenancy.BindingID](0x10), OrganizationID: orgA, SubjectUserID: u, ScopeType: tenancy.ScopeOrganization, ScopeID: orgA, Role: "org_admin"}
	gb := tenancy.Binding{ID: uid[tenancy.BindingID](0x11), OrganizationID: orgA, SubjectUserID: u, ScopeType: tenancy.ScopeGroup, ScopeID: g1, GroupID: &g, Role: "group_member"}
	vb := tenancy.Binding{ID: uid[tenancy.BindingID](0x12), OrganizationID: orgA, SubjectUserID: v, ScopeType: tenancy.ScopeOrganization, ScopeID: orgA, Role: "org_owner"}
	f := Facts{Organization: tenancy.Organization{ID: orgA}, SubjectBindings: []tenancy.Binding{ob, gb, vb}}
	if b, ok := subjectBinding(f, org, u); !ok || b.ID != ob.ID {
		t.Errorf("organization binding of u = %+v, %t", b, ok)
	}
	if b, ok := subjectBinding(f, grp, u); !ok || b.ID != gb.ID {
		t.Errorf("group binding of u = %+v, %t", b, ok)
	}
	if b, ok := subjectBinding(f, org, v); !ok || b.ID != vb.ID {
		t.Errorf("organization binding of v = %+v, %t", b, ok)
	}
	if _, ok := subjectBinding(f, grp, v); ok {
		t.Error("v has no binding of g1")
	}
	g3p := g3
	if _, ok := subjectBinding(f, Scope{Kind: tenancy.ScopeGroup, Group: &g3p}, u); ok {
		t.Error("u has no binding of g3")
	}
	if f := organizationOnly(&TenantContext{organization: orgA, group: g1}, org); f == nil || f.code != CodeInvalidInput || f.field != fieldContext {
		t.Errorf("selected-Group Organization departure = %+v, want invalid_input context", f)
	}
	if f := organizationOnly(&TenantContext{organization: orgA, group: g1}, grp); f != nil {
		t.Errorf("Group departure in selected context = %+v", f)
	}
	if f := organizationOnly(&TenantContext{organization: orgA}, org); f != nil {
		t.Errorf("Organization-only departure = %+v", f)
	}
}

func TestMembershipFailure(t *testing.T) {
	tx := &trackedTx{}
	ctx := context.Background()
	for code, want := range map[tenancy.Code]struct {
		code  Code
		field string
	}{
		tenancy.CodeLastOwner:            {CodeConflict, "last_owner"},
		tenancy.CodeDependentBindings:    {CodeConflict, "dependent_bindings"},
		tenancy.CodeRecipientIneligible:  {CodeConflict, "recipient_ineligible"},
		tenancy.CodeIssuerAuthorityLost:  {CodeConflict, "issuer_authority_lost"},
		tenancy.CodeInvitationLimit:      {CodeConflict, "invitation_limit"},
		tenancy.CodeInvitationNotPending: {CodeConflict, "invitation_not_pending"},
		tenancy.CodeMembershipExists:     {CodeConflict, "membership_exists"},
		tenancy.CodeOwnerAuthority:       {CodeForbidden, "owner_authority"},
		tenancy.CodeSelfEscalation:       {CodeForbidden, "self_escalation"},
		tenancy.CodeOwnerSubject:         {CodeForbidden, "owner_subject"}, // G060 amendment01: the guard's departure cascade
		tenancy.CodeContextRequired:      {CodeInvalidInput, "context"},
		tenancy.CodeMembershipNotFound:   {CodeNotFound, ""},
		tenancy.CodeInvitationNotFound:   {CodeNotFound, ""},
		tenancy.CodeUserNotFound:         {CodeNotFound, ""},
		tenancy.CodeInvalidRole:          {CodeInvalidInput, "invalid_role"},
		tenancy.CodeUnavailable:          {CodeUnavailable, ""},
		tenancy.CodeInternal:             {CodeInternal, ""},
	} {
		f := membershipFailure(ctx, tx, &tenancy.Error{Code: code})
		if f == nil || f.code != want.code || f.field != want.field {
			t.Errorf("%s: %+v, want %s/%s", code, f, want.code, want.field)
		}
	}
	if f := membershipFailure(ctx, tx, errors.New("x")); f == nil || f.code != CodeInternal {
		t.Errorf("foreign error: %+v", f)
	}
}

// claimRowTx is a commitTx (U-COMMIT) whose one claim statement returns the consumed status row, or raises raise.
type claimRowTx struct {
	*commitTx
	raise error
}

func (t *claimRowTx) QueryRow(context.Context, string, ...any) pgx.Row {
	return claimRow{raise: t.raise}
}

type claimRow struct{ raise error }

func (r claimRow) Scan(dest ...any) error {
	if r.raise != nil {
		return r.raise
	}
	*dest[0].(*string) = "consumed"
	return nil
}

type claimDB struct {
	raise                      error
	begins, commits, rollbacks atomic.Int64
}

func (d *claimDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	d.begins.Add(1)
	return &claimRowTx{commitTx: &commitTx{commits: &d.commits, rollbacks: &d.rollbacks}, raise: d.raise}, nil
}

// TestClaimFailureCause (G019 review; the claimFailure finding): both consumers' claimFailure, through inTx
// and the real tenancy.ClaimInvitationTx. The consumed status row (a successful Scan: local invitation_not_pending,
// tx.cause nil) is the consumed outcome on a healthy transaction (the database consumed replay is F2-22/F2-23). A raised
// 23000 invitations_keys_immutable maps to the same code but is tracked as tx.cause: it is the conflict
// invitation_not_pending (membershipFailure), with the deferred rollback and no COMMIT attempt on the aborted transaction.
func TestClaimFailureCause(t *testing.T) {
	token, _, err := tenancy.NewInvitationToken()
	if err != nil {
		t.Fatal(err)
	}
	raised := &pgconn.PgError{Severity: "ERROR", SeverityUnlocalized: "ERROR", Code: "23000", ConstraintName: "invitations_keys_immutable"}
	for _, c := range []struct {
		name     string
		raise    error
		consumed bool
		commits  int64
	}{
		{"consumed status row, nil cause", nil, true, 1},
		{"raised invitations_keys_immutable, tracked cause", raised, false, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			db := &claimDB{raise: c.raise}
			s := newService(db, nil, nil)
			ctx := context.Background()
			var consumed bool
			var cause error
			f := s.inTx(ctx, true, cleanupUnbounded, func(tx *trackedTx) *failure {
				_, err := tenancy.ClaimInvitationTx(ctx, tx, uid[tenancy.InvitationID](0x40), token, nil)
				if tenancy.CodeOf(err) != tenancy.CodeInvitationNotPending {
					t.Errorf("claim error %v, want invitation_not_pending", err)
				}
				var fl *failure
				consumed, fl = claimFailure(ctx, tx, err)
				cause = tx.cause
				return fl
			})
			if consumed != c.consumed || (cause == nil) != (c.raise == nil) {
				t.Errorf("consumed %v with cause %v, want consumed %v", consumed, cause, c.consumed)
			}
			if c.consumed && f != nil {
				t.Errorf("consumed replay failure %+v, want none", f)
			}
			if !c.consumed && (f == nil || f.code != CodeConflict || f.field != string(tenancy.CodeInvitationNotPending)) {
				t.Errorf("raised claim error: %+v, want conflict invitation_not_pending", f)
			}
			if b, cm, rb := db.begins.Load(), db.commits.Load(), db.rollbacks.Load(); b != 1 || cm != c.commits || rb != 1 {
				t.Errorf("BeginTx %d, Commit %d, Rollback %d: want 1, %d, 1", b, cm, rb, c.commits)
			}
		})
	}
}

func TestInvitationView(t *testing.T) {
	addr, g, u := "a@example.com", g1, uid[tenancy.UserID](0x05)
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	inv := tenancy.Invitation{ID: uid[tenancy.InvitationID](0x20), OrganizationID: orgA, ScopeType: tenancy.ScopeGroup, ScopeID: g1,
		GroupID: &g, Role: "group_member", IntendedUserID: &u, IntendedAddress: &addr, ExpiresAt: now.Add(time.Hour)}
	admin, recipient := invitationView(inv, true, now), invitationView(inv, false, now)
	if admin.IntendedAddress == nil || *admin.IntendedAddress != addr || recipient.IntendedAddress != nil {
		t.Errorf("address: admin %v, recipient %v", admin.IntendedAddress, recipient.IntendedAddress)
	}
	if admin.Status != tenancy.InvitationPending || admin.Scope.Kind != tenancy.ScopeGroup || admin.Scope.Group == nil || *admin.Scope.Group != g1 {
		t.Errorf("view %+v", admin)
	}
	// D-A3: no authorized reader supplies the issuer's name; it stays nil (never guessed) in both views.
	if admin.IssuerDisplayName != nil || recipient.IssuerDisplayName != nil {
		t.Errorf("issuer name: admin %v, recipient %v", admin.IssuerDisplayName, recipient.IssuerDisplayName)
	}
	if v := invitationView(inv, true, now.Add(2*time.Hour)); v.Status != tenancy.InvitationExpired {
		t.Errorf("expired status %q", v.Status)
	}
	rows := []tenancy.Invitation{inv, inv, inv}
	for i := range rows {
		rows[i].ID = uid[tenancy.InvitationID](byte(0x30 + i))
	}
	p := invitationPage(rows, 2, false)
	if len(p.Invitations) != 2 || p.Next == nil || *p.Next != rows[1].ID || p.Invitations[0].IntendedAddress != nil {
		t.Errorf("page %+v", p)
	}
	if p := invitationPage(rows[:2], 2, true); p.Next != nil || len(p.Invitations) != 2 {
		t.Errorf("last page %+v", p)
	}
}

func TestRetirementBlockers(t *testing.T) {
	if b := blockers(nil); b.Count != 0 || b.Scopes == nil || len(b.Scopes) != 0 || b.Truncated {
		t.Errorf("none: %+v", b)
	}
	many := make([]ScopeRef, MaxRetirementBlockers+1)
	if b := blockers(many); b.Count != MaxRetirementBlockers+1 || len(b.Scopes) != MaxRetirementBlockers || !b.Truncated {
		t.Errorf("many: count %d, %d scopes, truncated %t", b.Count, len(b.Scopes), b.Truncated)
	}
}

// TestRetireUserRefusesBeforeSQL (MEM-14 unit half, Result addendum A1-01): a zero user is refused before any statement
// (no pool here, so any statement would panic), with empty, non-nil progress and the user not disabled.
func TestRetireUserRefusesBeforeSQL(t *testing.T) {
	s := newService(nil, nil, nil)
	p, err := s.RetireUser(context.Background(), tenancy.UserID{})
	var e *Error
	if !errors.As(err, &e) || e.Code != CodeInvalidInput {
		t.Errorf("zero user: %v", err)
	}
	if p.Disabled || p.Departed == nil || p.AlreadyDeparted == nil || p.Remaining == nil || p.Blockers.Scopes == nil ||
		len(p.Departed)+len(p.AlreadyDeparted)+len(p.Remaining)+p.Blockers.Count != 0 {
		t.Errorf("zero user progress: %+v", p)
	}
}

// TestOrganizationAfter: the enumeration's ascending guard uses the reader's uuid (byte) order; the zero cursor precedes
// every id and an equal or lower id is never "after".
func TestOrganizationAfter(t *testing.T) {
	lo, hi := uid[tenancy.OrganizationID](0x10), uid[tenancy.OrganizationID](0x20)
	for name, c := range map[string]struct {
		o, after tenancy.OrganizationID
		want     bool
	}{
		"zero cursor": {lo, tenancy.OrganizationID{}, true},
		"higher":      {hi, lo, true},
		"equal":       {lo, lo, false},
		"lower":       {lo, hi, false},
	} {
		if got := organizationAfter(c.o, c.after); got != c.want {
			t.Errorf("%s: %t, want %t", name, got, c.want)
		}
	}
	if retirementPage < 1 || retirementPage > 100 {
		t.Errorf("retirement page %d outside the reader's 1..100", retirementPage)
	}
}

// TestConsumeRefusesBeforeSQL: a zero VerifiedAddress, a malformed id or token is refused before any statement (the
// Service has no pool here, so any statement would panic).
func TestConsumeRefusesBeforeSQL(t *testing.T) {
	s := newService(nil, nil, nil)
	ctx := context.Background()
	_, err := s.ConsumeOrganizationInvitation(ctx, identity.VerifiedAddress{}, uid[tenancy.InvitationID](0x40), "token")
	var e *Error
	if !errors.As(err, &e) || e.Code != CodeInvalidInput || e.Field != fieldEvidence {
		t.Errorf("zero evidence: %v", err)
	}
	// MEM-11: evidence cannot be constructed outside identity, e.g. from JSON (unexported fields stay zero).
	var js identity.VerifiedAddress
	if err := json.Unmarshal([]byte(`{"userID":"8c6f0a3e-1d2b-4c5a-9e7f-0a1b2c3d4e5f","issuer":"i","subject":"s",`+
		`"address":"a@example.test","UserID":"x","Issuer":"i","Subject":"s","Address":"a@example.test"}`), &js); err != nil {
		t.Fatal(err)
	}
	_, err = s.ConsumeOrganizationInvitation(ctx, js, uid[tenancy.InvitationID](0x40), "token")
	if !js.IsZero() || !errors.As(err, &e) || e.Code != CodeInvalidInput || e.Field != fieldEvidence {
		t.Errorf("JSON-constructed evidence: zero %v, %v", js.IsZero(), err)
	}
	if f := tokenPre(tenancy.InvitationID{}, "")(); f == nil || f.code != CodeInvalidInput {
		t.Errorf("zero id: %+v", f)
	}
	if f := tokenPre(uid[tenancy.InvitationID](0x40), "short")(); f == nil || f.field != string(tenancy.CodeInvalidToken) {
		t.Errorf("malformed token: %+v", f)
	}
}
