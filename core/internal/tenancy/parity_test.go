package tenancy

import (
	"io/fs"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"core/migrations"
)

// U-PARITY: the Go vocabulary and bounds equal the literals of the embedded 00003_tenancy.sql.
func TestMigrationParity(t *testing.T) {
	b, err := fs.ReadFile(migrations.FS, "00003_tenancy.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, ok := strings.Cut(string(b), "-- +goose Down")
	if !ok {
		t.Fatal("no down section")
	}
	sql := regexp.MustCompile(`--[^\n]*`).ReplaceAllString(up, " ")
	quoted := regexp.MustCompile(`'([^']*)'`)
	list := func(s string) []string {
		var out []string
		for _, m := range quoted.FindAllStringSubmatch(s, -1) {
			out = append(out, m[1])
		}
		return out
	}

	m := regexp.MustCompile(`(?s)CONSTRAINT groups_slug_reserved CHECK \(slug NOT IN \(([^)]*)\)\)`).FindStringSubmatch(sql)
	if m == nil {
		t.Fatal("groups_slug_reserved not found")
	}
	reserved := list(m[1])
	slices.Sort(reserved)
	if strings.Join(reserved, ",") != strings.Join(ReservedSlugs(), ",") || len(reserved) != 24 {
		t.Errorf("reserved set: SQL %v, Go %v", reserved, ReservedSlugs())
	}

	m = regexp.MustCompile(`(?s)CONSTRAINT role_bindings_scope_consistent CHECK \((.*?)\)\s*\);`).FindStringSubmatch(sql)
	if m == nil {
		t.Fatal("role_bindings_scope_consistent not found")
	}
	branches := strings.Split(m[1], " OR ")
	if len(branches) != 2 {
		t.Fatalf("scope_consistent has %d branches", len(branches))
	}
	roleIn := regexp.MustCompile(`role IN \(([^)]*)\)`)
	scopeIs := regexp.MustCompile(`scope_type = '([a-z_]+)'`)
	sourceIs := regexp.MustCompile(`source = '([a-z_]+)'`)
	want := []struct {
		scope ScopeType
		roles []string
	}{
		{ScopeOrganization, []string{string(OrgOwner), string(OrgAdmin), string(OrgMember)}},
		{ScopeGroup, []string{string(GroupOwner), string(GroupAdmin), string(GroupMember)}},
	}
	for i, br := range branches {
		r, sc, so := roleIn.FindStringSubmatch(br), scopeIs.FindStringSubmatch(br), sourceIs.FindStringSubmatch(br)
		if r == nil || sc == nil || so == nil {
			t.Fatalf("branch %d incomplete: %s", i, br)
		}
		if sc[1] != string(want[i].scope) || strings.Join(list(r[1]), ",") != strings.Join(want[i].roles, ",") || so[1] != SourceMembership {
			t.Errorf("branch %d: scope %s roles %v source %s", i, sc[1], list(r[1]), so[1])
		}
	}

	for _, tbl := range []string{"organizations", "groups"} {
		m := regexp.MustCompile(`CONSTRAINT ` + tbl + `_name_valid CHECK \(char_length\(name\) BETWEEN 1 AND ([0-9]+)`).FindStringSubmatch(sql)
		if m == nil || m[1] != strconv.Itoa(MaxNameLength) {
			t.Errorf("%s name bound: %v", tbl, m)
		}
	}
	m = regexp.MustCompile(`CONSTRAINT groups_slug_format CHECK \(char_length\(slug\) BETWEEN 1 AND ([0-9]+)`).FindStringSubmatch(sql)
	if m == nil || m[1] != strconv.Itoa(MaxSlugLength) {
		t.Errorf("slug bound: %v", m)
	}
	if !strings.Contains(sql, `slug ~ '^[0-9a-z]+(-[0-9a-z]+)*$'`) {
		t.Error("slug grammar differs")
	}
	// The views project exactly the membership source of each scope.
	for _, v := range []string{
		`CREATE VIEW app.organization_memberships WITH (security_invoker = true) AS
  SELECT subject_user_id AS user_id, organization_id, role
  FROM app.role_bindings WHERE scope_type = 'organization' AND source = 'membership';`,
		`CREATE VIEW app.group_memberships WITH (security_invoker = true) AS
  SELECT subject_user_id AS user_id, organization_id, group_id, role
  FROM app.role_bindings WHERE scope_type = 'group' AND source = 'membership';`,
	} {
		if !strings.Contains(sql, v) {
			t.Errorf("view differs: %s", v)
		}
	}

	// G057 (Plan r5 §3.5): the platform label of 00021_platform_labels.sql repeats the 00011 slug floor exactly (grammar,
	// bound and reserved list), so a platform label is never a host name the slug floor refuses.
	var floors []string
	for _, c := range []struct{ file, col string }{
		{"00011_organization_domains.sql", "slug"}, {"00021_platform_labels.sql", "platform_label"},
	} {
		b, err := fs.ReadFile(migrations.FS, c.file)
		if err != nil {
			t.Fatal(err)
		}
		up, _, ok := strings.Cut(string(b), "-- +goose Down")
		if !ok {
			t.Fatalf("%s: no down section", c.file)
		}
		src := regexp.MustCompile(`--[^\n]*`).ReplaceAllString(up, " ")
		if !strings.Contains(src, `CHECK (char_length(`+c.col+`) BETWEEN 1 AND 63 AND `+c.col+` ~ '^[0-9a-z]+(-[0-9a-z]+)*$')`) {
			t.Errorf("%s: organizations.%s grammar or bound differs", c.file, c.col)
		}
		m := regexp.MustCompile(`(?s)CONSTRAINT organizations_` + c.col + `_reserved CHECK \(` + c.col + ` NOT IN \(([^)]*)\)\)`).FindStringSubmatch(src)
		if m == nil {
			t.Fatalf("%s: organizations_%s_reserved not found", c.file, c.col)
		}
		r := list(m[1])
		slices.Sort(r)
		if len(r) != 35 || len(slices.Compact(slices.Clone(r))) != 35 {
			t.Errorf("%s: %d reserved labels (unique %v), want 35", c.file, len(r), r)
		}
		floors = append(floors, strings.Join(r, ","))
	}
	if floors[0] != floors[1] {
		t.Errorf("reserved floors differ:\n00011 slug %s\n00021 platform_label %s", floors[0], floors[1])
	}
}

// U-PARITY (G006): the Go context bounds and the website prefix rule equal the CHECK literals of the embedded
// 00004_business_profiles.sql, and its access-path view is exactly the pinned definition (00003 parity is unchanged).
func TestBusinessProfileMigrationParity(t *testing.T) {
	b, err := fs.ReadFile(migrations.FS, "00004_business_profiles.sql")
	if err != nil {
		t.Fatal(err)
	}
	up, _, ok := strings.Cut(string(b), "-- +goose Down")
	if !ok {
		t.Fatal("no down section")
	}
	sql := regexp.MustCompile(`--[^\n]*`).ReplaceAllString(up, " ")
	for col, want := range map[string]int{"display_name": MaxDisplayNameLength, "legal_name": MaxLegalNameLength,
		"description": MaxDescriptionLength, "industry": MaxIndustryLength, "website": MaxWebsiteLength} {
		m := regexp.MustCompile(`CONSTRAINT business_profiles_` + col + `_valid CHECK \((?:` + col + ` IS NULL OR \()?char_length\(` + col +
			`\) BETWEEN 1 AND ([0-9]+)`).FindStringSubmatch(sql)
		if m == nil || m[1] != strconv.Itoa(want) {
			t.Errorf("%s bound: SQL %v, Go %d", col, m, want)
		}
	}
	if !strings.Contains(sql, `(website COLLATE "C") ~ '^https?://[\u0021-\u005b\u005d-\u007e]+$'`) {
		t.Error("website CHECK differs from printable ASCII without space or backslash after http:// or https://")
	}
	for _, c := range []string{"display_name", "legal_name", "industry"} {
		if !strings.Contains(sql, `(`+c+` COLLATE "C") !~ '[\u0001-\u001f\u007f-\u009f]'`) {
			t.Errorf("%s control-character rule differs", c)
		}
	}
	if !strings.Contains(sql, `(description COLLATE "C") !~ '[\u0001-\u0008\u000b-\u001f\u007f-\u009f]'`) {
		t.Error("description control-character rule differs (TAB and LF allowed)")
	}
	view := `CREATE VIEW app.business_profile_access_paths WITH (security_invoker = true) AS
  SELECT rb.subject_user_id AS user_id, gbp.organization_id, gbp.business_profile_id, gbp.group_id, rb.role AS group_role
  FROM app.group_business_profiles gbp
  JOIN app.role_bindings rb ON rb.organization_id = gbp.organization_id AND rb.group_id = gbp.group_id
  WHERE rb.scope_type = 'group' AND rb.source = 'membership';`
	if !strings.Contains(sql, view) {
		t.Error("access-path view differs from the pinned definition")
	}
	if strings.Contains(sql, "role_bindings ADD") || strings.Contains(sql, "ALTER TABLE") {
		t.Error("00004 alters a table")
	}
}
