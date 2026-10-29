package tenancy

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// MEM-8 (unit): the single address normalization. Local part byte-exact (case and Unicode included); domain NFC then
// lowercase; trim of surrounding ASCII whitespace only; refusals for structure, whitespace/control characters and the
// 64/254-byte bounds.
func TestNormalizeAddress(t *testing.T) {
	const decomposedE = "é" // e + COMBINING ACUTE ACCENT
	const composedE = "é"    // é
	accept := []struct{ in, want string }{
		{"user@example.com", "user@example.com"},
		{"  user@example.com\t", "user@example.com"},
		{"User.Name+tag@Example.COM", "User.Name+tag@example.com"},
		{"MiXeD@EXAMPLE.org", "MiXeD@example.org"},
		{"J" + decomposedE + "r@example.com", "J" + decomposedE + "r@example.com"},   // local part kept byte-exact
		{"user@caf" + decomposedE + ".example", "user@caf" + composedE + ".example"}, // domain composed
		{"user@CAF" + "É" + ".EXAMPLE", "user@caf" + composedE + ".example"},         // domain composed and lowercased
		{"a@b", "a@b"},
		{strings.Repeat("l", 64) + "@example.com", strings.Repeat("l", 64) + "@example.com"},
		{"x@" + strings.Repeat("d", 252), "x@" + strings.Repeat("d", 252)},
	}
	for _, c := range accept {
		got, err := NormalizeAddress(c.in)
		if err != nil || got != c.want {
			t.Errorf("NormalizeAddress(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	refuse := []string{
		"", "   ", "user", "@example.com", "user@", "a@b@c", "us er@example.com", "user@exa mple.com",
		"user x@example.com", "user@example.com\x00", "us\ter@example.com", "user@ex ample.com",
		"user\x7f@example.com", string([]byte{0xff, '@', 'a'}),
		strings.Repeat("l", 65) + "@example.com",
		"x@" + strings.Repeat("d", 253),
	}
	for _, in := range refuse {
		if got, err := NormalizeAddress(in); err == nil {
			t.Errorf("NormalizeAddress(%q) = %q, want invalid_address", in, got)
		} else {
			var te *Error
			if !errors.As(err, &te) || te.Code != CodeInvalidAddress {
				t.Errorf("NormalizeAddress(%q) error = %v, want invalid_address", in, err)
			}
		}
	}
	// A local-part case variant is a different address; a domain case variant is the same after normalization.
	a, _ := NormalizeAddress("Alice@Example.com")
	b, _ := NormalizeAddress("alice@example.com")
	c, _ := NormalizeAddress("Alice@EXAMPLE.COM")
	if a == b || a != c {
		t.Errorf("case handling: %q %q %q", a, b, c)
	}
}

// Token shape and digest: 43-character base64url tokens from 32 random bytes; the tagged SHA-256 digest; distinct
// tokens per call; the digest never equals the untagged hash.
func TestInvitationToken(t *testing.T) {
	tok, dig, err := NewInvitationToken()
	if err != nil || len(tok) != 43 || !ValidInvitationToken(tok) || len(dig) != 32 {
		t.Fatalf("token %q (%d) digest %d err %v", tok, len(tok), len(dig), err)
	}
	if !bytes.Equal(dig, InvitationDigest(tok)) || !DigestEqual(dig, InvitationDigest(tok)) {
		t.Fatal("digest is not deterministic")
	}
	plain := sha256.Sum256([]byte(tok))
	if bytes.Equal(dig, plain[:]) {
		t.Fatal("digest is not domain-separated")
	}
	tok2, dig2, _ := NewInvitationToken()
	if tok == tok2 || DigestEqual(dig, dig2) {
		t.Fatal("tokens repeat")
	}
	for _, bad := range []string{"", tok[:42], tok + "A", strings.Repeat("+", 43), strings.Repeat("A", 42) + "="} {
		if ValidInvitationToken(bad) {
			t.Errorf("ValidInvitationToken(%q) = true", bad)
		}
	}
	if DigestEqual(dig, dig[:31]) {
		t.Error("DigestEqual accepts a length mismatch")
	}
}

// Derived status and the invitable role set (never an owner role).
func TestInvitationStatusAndRoles(t *testing.T) {
	now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	at := now.Add(-time.Minute)
	cases := []struct {
		inv  Invitation
		want string
	}{
		{Invitation{ExpiresAt: now.Add(time.Hour)}, InvitationPending},
		{Invitation{ExpiresAt: now}, InvitationExpired},
		{Invitation{ExpiresAt: now.Add(time.Hour), RevokedAt: &at}, InvitationRevoked},
		{Invitation{ExpiresAt: now.Add(-time.Hour), ConsumedAt: &at}, InvitationConsumed},
	}
	for i, c := range cases {
		if got := c.inv.Status(now); got != c.want {
			t.Errorf("case %d: status %q, want %q", i, got, c.want)
		}
	}
	for _, r := range []string{"org_admin", "org_member"} {
		if !InvitationRoleValid(ScopeOrganization, r) || InvitationRoleValid(ScopeGroup, r) {
			t.Errorf("role %s", r)
		}
	}
	for _, r := range []string{"group_admin", "group_member"} {
		if !InvitationRoleValid(ScopeGroup, r) || InvitationRoleValid(ScopeOrganization, r) {
			t.Errorf("role %s", r)
		}
	}
	for _, r := range []string{"org_owner", "group_owner", "", "admin"} {
		if InvitationRoleValid(ScopeOrganization, r) || InvitationRoleValid(ScopeGroup, r) {
			t.Errorf("owner or unknown role %q is invitable", r)
		}
	}
	if InvitationMaxLifetime != 30*24*time.Hour || InvitationMinLifetime != time.Hour || MaxPendingInvitations != 100 {
		t.Error("lifetime or bound constants changed")
	}
}

// The invitation writers and readers validate before any statement: a nil transaction is never touched by a rejected
// call (TestServiceRejectsBeforeSQL extension, Plan r1 §3).
func TestInvitationRejectsBeforeSQL(t *testing.T) {
	ctx := t.Context()
	org, _ := ParseOrganizationID("0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	u, _ := ParseUserID("1f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	g, _ := ParseGroupID("2f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	id, _ := ParseInvitationID("4f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	token, _, err := NewInvitationToken()
	if err != nil {
		t.Fatal(err)
	}
	var zu UserID
	var zi InvitationID
	orgSpec := InvitationSpec{ID: id, OrganizationID: org, Role: string(OrgMember), Issuer: u, IntendedAddress: "a@example.com"}
	v, _ := ParseUserID("3f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	grpSpec := InvitationSpec{ID: id, OrganizationID: org, Group: g, Role: string(GroupMember), Issuer: u, IntendedUser: v}
	with := func(sp InvitationSpec, f func(*InvitationSpec)) InvitationSpec { f(&sp); return sp }
	ev := &OrganizationClaim{Issuer: "https://issuer.example", Subject: "user_1", Address: "a@example.com"}
	for _, c := range []struct {
		name string
		err  error
		want Code
	}{
		{"insert zero id", third(InsertInvitationTx(ctx, nil, with(orgSpec, func(s *InvitationSpec) { s.ID = zi }))), CodeInvalidID},
		{"insert zero org", third(InsertInvitationTx(ctx, nil, with(orgSpec, func(s *InvitationSpec) { s.OrganizationID = OrganizationID{} }))), CodeInvalidID},
		{"insert zero issuer", third(InsertInvitationTx(ctx, nil, with(grpSpec, func(s *InvitationSpec) { s.Issuer = zu }))), CodeInvalidID},
		{"insert owner role (org)", third(InsertInvitationTx(ctx, nil, with(orgSpec, func(s *InvitationSpec) { s.Role = string(OrgOwner) }))), CodeInvalidRole},
		{"insert owner role (group)", third(InsertInvitationTx(ctx, nil, with(grpSpec, func(s *InvitationSpec) { s.Role = string(GroupOwner) }))), CodeInvalidRole},
		{"insert group role on org scope", third(InsertInvitationTx(ctx, nil, with(orgSpec, func(s *InvitationSpec) { s.Role = string(GroupAdmin) }))), CodeInvalidRole},
		{"insert org with user", third(InsertInvitationTx(ctx, nil, with(orgSpec, func(s *InvitationSpec) { s.IntendedUser = u }))), CodeInvalidID},
		{"insert org bad address", third(InsertInvitationTx(ctx, nil, with(orgSpec, func(s *InvitationSpec) { s.IntendedAddress = "no-at" }))), CodeInvalidAddress},
		{"insert org empty address", third(InsertInvitationTx(ctx, nil, with(orgSpec, func(s *InvitationSpec) { s.IntendedAddress = "" }))), CodeInvalidAddress},
		{"insert group without user", third(InsertInvitationTx(ctx, nil, with(grpSpec, func(s *InvitationSpec) { s.IntendedUser = zu }))), CodeInvalidID},
		{"insert group with address", third(InsertInvitationTx(ctx, nil, with(grpSpec, func(s *InvitationSpec) { s.IntendedAddress = "a@example.com" }))), CodeInvalidAddress},
		{"insert group naming its issuer", third(InsertInvitationTx(ctx, nil, with(grpSpec, func(s *InvitationSpec) { s.IntendedUser = u }))), CodeSelfTarget},
		{"insert lifetime under 1h", third(InsertInvitationTx(ctx, nil, with(grpSpec, func(s *InvitationSpec) { s.Lifetime = time.Hour - time.Second }))), CodeInvalidLifetime},
		{"insert lifetime 1h less than 1s", third(InsertInvitationTx(ctx, nil, with(grpSpec, func(s *InvitationSpec) { s.Lifetime = time.Hour - time.Millisecond }))), CodeInvalidLifetime},
		{"insert lifetime over 30d", third(InsertInvitationTx(ctx, nil, with(grpSpec, func(s *InvitationSpec) { s.Lifetime = InvitationMaxLifetime + time.Second }))), CodeInvalidLifetime},
		{"insert negative lifetime", third(InsertInvitationTx(ctx, nil, with(grpSpec, func(s *InvitationSpec) { s.Lifetime = -time.Hour }))), CodeInvalidLifetime},
		{"revoke zero org", second(RevokeInvitationTx(ctx, nil, OrganizationID{}, id)), CodeInvalidID},
		{"revoke zero id", second(RevokeInvitationTx(ctx, nil, org, zi)), CodeInvalidID},
		{"claim zero id", second(ClaimInvitationTx(ctx, nil, zi, token, nil)), CodeInvalidID},
		{"claim empty token", second(ClaimInvitationTx(ctx, nil, id, "", nil)), CodeInvalidToken},
		{"claim short token", second(ClaimInvitationTx(ctx, nil, id, token[:42], ev)), CodeInvalidToken},
		{"claim padded token", second(ClaimInvitationTx(ctx, nil, id, token[:42]+"=", nil)), CodeInvalidToken},
		{"claim org empty issuer", second(ClaimInvitationTx(ctx, nil, id, token, &OrganizationClaim{Subject: "s", Address: "a@example.com"})), CodeInvalidID},
		{"claim org empty subject", second(ClaimInvitationTx(ctx, nil, id, token, &OrganizationClaim{Issuer: "i", Address: "a@example.com"})), CodeInvalidID},
		{"claim org bad address", second(ClaimInvitationTx(ctx, nil, id, token, &OrganizationClaim{Issuer: "i", Subject: "s", Address: "a@@b"})), CodeInvalidAddress},
		{"page zero org", second(InvitationsPage(ctx, nil, OrganizationID{}, GroupID{}, zi, 10)), CodeInvalidID},
		{"page limit 0", second(InvitationsPage(ctx, nil, org, g, zi, 0)), CodeInvalidID},
		{"page limit 101", second(InvitationsPage(ctx, nil, org, GroupID{}, zi, MaxMembersPage+1)), CodeInvalidID},
		{"mine zero org", second(MyInvitationsPage(ctx, nil, OrganizationID{}, u, zi, 10)), CodeInvalidID},
		{"mine zero recipient", second(MyInvitationsPage(ctx, nil, org, zu, zi, 10)), CodeInvalidID},
		{"mine limit 101", second(MyInvitationsPage(ctx, nil, org, u, zi, MaxMembersPage+1)), CodeInvalidID},
		{"pending zero org", second(PendingInvitationCount(ctx, nil, OrganizationID{}, g)), CodeInvalidID},
	} {
		if got := CodeOf(c.err); got != c.want {
			t.Errorf("%s: %v, want %s", c.name, c.err, c.want)
		}
	}
}

// The spec's derived values: scope from Group, the normalized address for the Organization form, the default and
// whole-second lifetimes.
func TestInvitationSpecValidate(t *testing.T) {
	org, _ := ParseOrganizationID("0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	u, _ := ParseUserID("1f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	g, _ := ParseGroupID("2f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	id, _ := ParseInvitationID("4f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	scope, addr, secs, err := InvitationSpec{ID: id, OrganizationID: org, Role: string(OrgAdmin), Issuer: u,
		IntendedAddress: "  Ab@EXAMPLE.Com "}.validate()
	if err != nil || scope != ScopeOrganization || addr == nil || *addr != "Ab@example.com" || secs != int64(InvitationDefaultLifetime/time.Second) {
		t.Errorf("organization spec = %s %v %d %v", scope, addr, secs, err)
	}
	v, _ := ParseUserID("3f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	scope, addr, secs, err = InvitationSpec{ID: id, OrganizationID: org, Group: g, Role: string(GroupAdmin), Issuer: u, IntendedUser: v,
		Lifetime: time.Hour + 1500*time.Millisecond}.validate()
	if err != nil || scope != ScopeGroup || addr != nil || secs != 3601 {
		t.Errorf("group spec = %s %v %d %v", scope, addr, secs, err)
	}
	if _, _, secs, err := (InvitationSpec{ID: id, OrganizationID: org, Group: g, Role: string(GroupMember), Issuer: u, IntendedUser: v,
		Lifetime: InvitationMaxLifetime}).validate(); err != nil || secs != 30*24*3600 {
		t.Errorf("maximum lifetime = %d %v", secs, err)
	}
	// G060 (r1 C3): a Group invitation naming its issuer is refused before any statement, for every Group role.
	for _, role := range []GroupRole{GroupAdmin, GroupMember} {
		if _, _, _, err := (InvitationSpec{ID: id, OrganizationID: org, Group: g, Role: string(role), Issuer: u, IntendedUser: u}).validate(); CodeOf(err) != CodeSelfTarget {
			t.Errorf("self-targeted %s group spec: %v, want self_target", role, err)
		}
	}
	if s, id := scopeOf(org, GroupID{}); s != ScopeOrganization || id.Bytes != org || !id.Valid {
		t.Errorf("scopeOf(org) = %s %v", s, id)
	}
	if s, id := scopeOf(org, g); s != ScopeGroup || id.Bytes != g {
		t.Errorf("scopeOf(group) = %s %v", s, id)
	}
}

// claimRow is a canned claim-body row.
type claimRow struct {
	outcome                    string
	org, scope, group, binding pgtype.UUID
	scopeType, role            pgtype.Text
	err                        error
}

func (r claimRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*string) = r.outcome
	*dest[1].(*pgtype.UUID), *dest[2].(*pgtype.Text), *dest[3].(*pgtype.UUID) = r.org, r.scopeType, r.scope
	*dest[4].(*pgtype.UUID), *dest[5].(*pgtype.Text), *dest[6].(*pgtype.UUID) = r.group, r.role, r.binding
	return nil
}

// The claim bodies' outcomes map onto ClaimOutcome (joined, upgraded, already_member) or the closed codes (not_found ->
// invitation_not_found, consumed -> invitation_not_pending); anything else, or a success row without its scope or
// binding, is internal.
func TestScanClaim(t *testing.T) {
	org, _ := ParseOrganizationID("0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	g, _ := ParseGroupID("2f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	b, _ := ParseBindingID("3f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	ok := claimRow{org: uuidArg(org), scope: uuidArg(g), group: uuidArg(g), binding: uuidArg(b),
		scopeType: pgtype.Text{String: "group", Valid: true}, role: pgtype.Text{String: "group_member", Valid: true}}
	for _, o := range []string{ClaimJoined, ClaimUpgraded, ClaimAlreadyMember} {
		r := ok
		r.outcome = o
		out, err := scanClaim(r)
		if err != nil || out.Outcome != o || out.OrganizationID != org || out.ScopeType != ScopeGroup || out.ScopeID != g ||
			out.GroupID == nil || *out.GroupID != g || out.Role != "group_member" || out.BindingID != b {
			t.Errorf("%s: %+v %v", o, out, err)
		}
	}
	orgRow := claimRow{outcome: ClaimJoined, org: uuidArg(org), scope: uuidArg(org), binding: uuidArg(b),
		scopeType: pgtype.Text{String: "organization", Valid: true}, role: pgtype.Text{String: "org_member", Valid: true}}
	if out, err := scanClaim(orgRow); err != nil || out.GroupID != nil || out.ScopeType != ScopeOrganization {
		t.Errorf("organization claim: %+v %v", out, err)
	}
	noBinding := ok
	noBinding.outcome, noBinding.binding = ClaimJoined, pgtype.UUID{}
	for _, c := range []struct {
		name string
		row  claimRow
		want Code
	}{
		{"not_found", claimRow{outcome: "not_found"}, CodeInvitationNotFound},
		{"consumed", claimRow{outcome: "consumed", org: uuidArg(org), scope: uuidArg(g), group: uuidArg(g),
			scopeType: pgtype.Text{String: "group", Valid: true}, role: pgtype.Text{String: "group_member", Valid: true}}, CodeInvitationNotPending},
		{"unknown outcome", claimRow{outcome: "restored"}, CodeInternal},
		{"empty outcome", claimRow{}, CodeInternal},
		{"success without binding", noBinding, CodeInternal},
	} {
		if _, err := scanClaim(c.row); CodeOf(err) != c.want {
			t.Errorf("%s: %v, want %s", c.name, err, c.want)
		}
	}
	// A missing row (never produced by the set-returning bodies) is the closed not-found; a store error keeps its class.
	if CodeOf(classify(one(claimRow{err: pgx.ErrNoRows}.Scan(), CodeInvitationNotFound))) != CodeInvitationNotFound {
		t.Error("no rows is not invitation_not_found")
	}
}
