package tenancy

import (
	"strings"
	"testing"
)

// U-ID: canonical hyphenated UUIDs only, case-insensitive and normalized to lowercase; everything else is invalid_id.
func TestParseIDs(t *testing.T) {
	const upper = "0F1E2D3C-4B5A-4978-8695-A4B3C2D1E0F9"
	id, err := ParseOrganizationID(upper)
	if err != nil {
		t.Fatalf("parse %s: %v", upper, err)
	}
	if got := id.String(); got != strings.ToLower(upper) {
		t.Fatalf("String() = %s", got)
	}
	g, err := ParseGroupID(strings.ToLower(upper))
	if err != nil || [16]byte(g) != [16]byte(id) {
		t.Fatalf("group parse: %v", err)
	}
	if _, err := ParseUserID(upper); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseBindingID(upper); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		"", "00000000-0000-0000-0000-000000000000", "{0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9}",
		"urn:uuid:0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9", "0f1e2d3c4b5a49788695a4b3c2d1e0f9",
		"0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f", "0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9a", "0f1e2d3c_4b5a_4978_8695_a4b3c2d1e0f9",
		"0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0fg", " 0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f", "0f1e2d3c-4b5a-4978-8695+a4b3c2d1e0f9",
		"0f1e2d3c-4b5a-4978-869-5a4b3c2d1e0f9",
	} {
		for name, parse := range map[string]func(string) error{
			"organization": func(s string) error { _, err := ParseOrganizationID(s); return err },
			"group":        func(s string) error { _, err := ParseGroupID(s); return err },
			"binding":      func(s string) error { _, err := ParseBindingID(s); return err },
			"user":         func(s string) error { _, err := ParseUserID(s); return err },
		} {
			if err := parse(bad); CodeOf(err) != CodeInvalidID {
				t.Errorf("%s %q: %v, want invalid_id", name, bad, err)
			}
		}
	}
}

// U-ROLE: exactly three Organization and three Group roles; scope types and the source are fixed.
func TestRoleVocabulary(t *testing.T) {
	for _, r := range []OrganizationRole{OrgOwner, OrgAdmin, OrgMember} {
		if !r.Valid() {
			t.Errorf("%s invalid", r)
		}
	}
	for _, r := range []GroupRole{GroupOwner, GroupAdmin, GroupMember} {
		if !r.Valid() {
			t.Errorf("%s invalid", r)
		}
	}
	for _, r := range []string{"", "owner", "admin", "member", "group_owner", "Org_Owner", "org_owner ", "be_owner", "be_admin",
		"business_profile_owner", "business_profile_admin", "advisor", "principal", "associate", "Owner", "Admin"} {
		if OrganizationRole(r).Valid() {
			t.Errorf("organization role %q accepted", r)
		}
	}
	for _, r := range []string{"", "owner", "admin", "member", "org_owner", "Group_Owner", "be_owner", "be_admin",
		"business_profile_member", "advisor", "principal", "associate"} {
		if GroupRole(r).Valid() {
			t.Errorf("group role %q accepted", r)
		}
	}
	if ScopeOrganization != "organization" || ScopeGroup != "group" || SourceMembership != "membership" {
		t.Fatal("scope or source codes changed")
	}
}

// Service validation happens before any SQL: a nil database is never touched by a rejected call.
func TestServiceRejectsBeforeSQL(t *testing.T) {
	s := NewService(nil, nil)
	ctx := t.Context()
	org, _ := ParseOrganizationID("0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	u, _ := ParseUserID("1f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	g, _ := ParseGroupID("2f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	var zo OrganizationID
	var zu UserID
	var zg GroupID
	cases := []struct {
		name string
		err  error
		want Code
	}{
		{"create org bad name", second(s.CreateOrganization(ctx, "  ")), CodeInvalidName},
		{"get org zero", second(s.GetOrganization(ctx, zo)), CodeInvalidID},
		{"rename org zero", second(s.RenameOrganization(ctx, zo, "A")), CodeInvalidID},
		{"rename org bad name", second(s.RenameOrganization(ctx, org, "a\x00b")), CodeInvalidName},
		{"suggest zero", second(s.SuggestGroupSlug(ctx, zo, "A")), CodeInvalidID},
		{"suggest bad name", second(s.SuggestGroupSlug(ctx, org, "")), CodeInvalidName},
		{"create group zero org", third(s.CreateGroup(ctx, zo, u, "A", "")), CodeInvalidID},
		{"create group zero creator", third(s.CreateGroup(ctx, org, zu, "A", "")), CodeInvalidID},
		{"create group bad name", third(s.CreateGroup(ctx, org, u, "\t", "")), CodeInvalidName},
		{"create group no base", third(s.CreateGroup(ctx, org, u, "東京チーム", "")), CodeSlugRequired},
		{"create group reserved base", third(s.CreateGroup(ctx, org, u, "API", "")), CodeSlugReserved},
		{"create group bad slug", third(s.CreateGroup(ctx, org, u, "Acme", "Acme")), CodeInvalidSlug},
		{"create group reserved slug", third(s.CreateGroup(ctx, org, u, "Acme", "settings")), CodeSlugReserved},
		{"get group zero", second(s.GetGroup(ctx, org, zg)), CodeInvalidID},
		{"get by slug bad", second(s.GetGroupBySlug(ctx, org, "a/b")), CodeInvalidSlug},
		{"list groups zero", second(s.ListGroups(ctx, zo)), CodeInvalidID},
		{"rename group bad name", second(s.RenameGroup(ctx, org, g, strings.Repeat("x", 201))), CodeInvalidName},
		{"change slug bad", second(s.ChangeGroupSlug(ctx, org, g, "acme--revenue")), CodeInvalidSlug},
		{"change slug reserved", second(s.ChangeGroupSlug(ctx, org, g, "mcp")), CodeSlugReserved},
		{"grant org bad role", second(s.GrantOrganizationRole(ctx, org, u, "group_owner")), CodeInvalidRole},
		{"grant org staffing role", second(s.GrantOrganizationRole(ctx, org, u, "advisor")), CodeInvalidRole},
		{"grant org zero user", second(s.GrantOrganizationRole(ctx, org, zu, OrgMember)), CodeInvalidID},
		{"change org bad role", second(s.ChangeOrganizationRole(ctx, org, u, "")), CodeInvalidRole},
		{"remove org zero", s.RemoveOrganizationMembership(ctx, zo, u), CodeInvalidID},
		{"grant group bad role", second(s.GrantGroupRole(ctx, org, g, u, "org_owner")), CodeInvalidRole},
		{"grant group profile role", second(s.GrantGroupRole(ctx, org, g, u, "business_profile_owner")), CodeInvalidRole},
		{"grant group zero group", second(s.GrantGroupRole(ctx, org, zg, u, GroupMember)), CodeInvalidID},
		{"change group bad role", second(s.ChangeGroupRole(ctx, org, g, u, "principal")), CodeInvalidRole},
		{"remove group zero user", s.RemoveGroupMembership(ctx, org, g, zu), CodeInvalidID},
		{"get org membership zero", second(s.GetOrganizationMembership(ctx, org, zu)), CodeInvalidID},
		{"list org memberships zero", second(s.ListOrganizationMemberships(ctx, zo)), CodeInvalidID},
		{"get group membership zero", second(s.GetGroupMembership(ctx, org, zg, u)), CodeInvalidID},
		{"list group memberships zero", second(s.ListGroupMemberships(ctx, org, zg)), CodeInvalidID},
	}
	for _, c := range cases {
		if got := CodeOf(c.err); got != c.want {
			t.Errorf("%s: %v, want %s", c.name, c.err, c.want)
		}
	}
}

func second[T any](_ T, err error) error        { return err }
func third[T, U any](_ T, _ U, err error) error { return err }
