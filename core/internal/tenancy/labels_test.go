package tenancy

import (
	"go/ast"
	"go/token"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"
)

// T2 (G020 Plan r2 §5.2): the label name grammar (ValidateLabelName: the ValidateName rule with the 80-code-point bound,
// D-5) and the uniqueness key the generated name_key column computes.

// sep returns the line (U+2028) or paragraph (U+2029) separator, built from its code point so this file holds no escape.
func sep(cp rune) string { return string(cp) }

func TestLabelNameBounds(t *testing.T) {
	for _, c := range []struct {
		in, want string
	}{
		{"a", "a"},
		{strings.Repeat("a", 80), strings.Repeat("a", 80)},
		{strings.Repeat("é", 80), strings.Repeat("é", 80)}, // 80 code points, 160 bytes
		{strings.Repeat("界", 80), strings.Repeat("界", 80)},
		{"  " + strings.Repeat("b", 80) + "\t", strings.Repeat("b", 80)}, // the bound applies after trimming
	} {
		got, err := ValidateLabelName(c.in)
		if err != nil || got != c.want {
			t.Errorf("ValidateLabelName(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	for _, in := range []string{"", " ", "\t\n", strings.Repeat("a", 81), strings.Repeat("é", 81), strings.Repeat("界", 81),
		strings.Repeat("a", 201)} {
		if _, err := ValidateLabelName(in); CodeOf(err) != CodeInvalidName {
			t.Errorf("ValidateLabelName(%d code points) = %v; want invalid_name", len([]rune(in)), err)
		}
	}
	if MaxLabelNameLength != 80 || MaxLabelDefinitionsPerScope != 200 || DefaultLabelPage != 50 || MaxLabelPage != 100 {
		t.Errorf("label bounds %d/%d/%d/%d, want 80/200/50/100", MaxLabelNameLength, MaxLabelDefinitionsPerScope, DefaultLabelPage, MaxLabelPage)
	}
}

func TestLabelNameTrim(t *testing.T) {
	for in, want := range map[string]string{
		" Urgent ":                 "Urgent",
		"\tUrgent\n":               "Urgent",
		string(rune(0x3000)) + "x": "x", // ideographic space is Unicode whitespace
		"two  words":               "two  words",
		"<b>kept</b>":              "<b>kept</b>", // markup is kept verbatim, never interpreted
	} {
		if got, err := ValidateLabelName(in); err != nil || got != want {
			t.Errorf("ValidateLabelName(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}

// TestLabelNameTrimSet pins the trimmed set (round 15, Codex source review): strings.TrimSpace removes exactly the 25
// Unicode White_Space code points (unicode.IsSpace) at a boundary, never JavaScript's trim() set, which differs in
// U+FEFF (trimmed there, kept and counted here) and U+0085 (trimmed here; a control inside). The web response predicate
// (lib/labels-client.ts isLabelName) mirrors this set, so every name core stores is read back.
func TestLabelNameTrimSet(t *testing.T) {
	whiteSpace := []rune{0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x20, 0x85, 0xa0, 0x1680, 0x2000, 0x2001, 0x2002, 0x2003, 0x2004,
		0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200a, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000}
	n := 0
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if unicode.IsSpace(r) {
			n++
			if !slices.Contains(whiteSpace, r) {
				t.Errorf("unicode.IsSpace(%U) is true but %U is not in the pinned set", r, r)
			}
		}
	}
	if n != len(whiteSpace) {
		t.Errorf("unicode.IsSpace holds for %d code points, want %d", n, len(whiteSpace))
	}
	for _, r := range whiteSpace {
		in := string(r) + "x" + string(r)
		got, err := ValidateLabelName(in)
		switch {
		case r == 0x2028 || r == 0x2029:
			if CodeOf(err) != CodeInvalidName {
				t.Errorf("ValidateLabelName(%q) = %q, %v; want invalid_name (separators are refused before trimming)", in, got, err)
			}
		case err != nil || got != "x":
			t.Errorf("ValidateLabelName(%q) = %q, %v; want %q (boundary %U trimmed)", in, got, err, "x", r)
		}
	}
	bom := string(rune(0xfeff))
	for _, in := range []string{bom + "x" + bom, bom, strings.Repeat(bom, 39) + "x" + strings.Repeat(bom, 40)} {
		if got, err := ValidateLabelName(in); err != nil || got != in {
			t.Errorf("ValidateLabelName(%q) = %q, %v; want it unchanged (U+FEFF is not White_Space)", in, got, err)
		}
	}
	if _, err := ValidateLabelName(strings.Repeat(bom, 40) + "x" + strings.Repeat(bom, 40)); CodeOf(err) != CodeInvalidName {
		t.Errorf("ValidateLabelName(81 code points with U+FEFF) = %v; want invalid_name (U+FEFF is counted)", err)
	}
	if got, err := ValidateLabelName("x" + string(rune(0x85))); err != nil || got != "x" {
		t.Errorf("ValidateLabelName(x U+0085) = %q, %v; want %q (U+0085 is White_Space at a boundary)", got, err, "x")
	}
}

func TestLabelNameControlCharacters(t *testing.T) {
	bad := []string{"a\x00b", "a\x01b", "a\x1fb", "a\x7fb", "a\u0080b", "a\u0085b", "a\u009fb", "a\nb", "a\tb", "\xff", "a\xc3",
		"a" + sep(0x2028) + "b", "a" + sep(0x2029) + "b", sep(0x2028) + "a", "a" + sep(0x2029)}
	for _, in := range bad {
		if _, err := ValidateLabelName(in); CodeOf(err) != CodeInvalidName {
			t.Errorf("ValidateLabelName(%q) = %v; want invalid_name", in, err)
		}
	}
	// U+00A0 and U+FEFF are not control characters: kept inside a name.
	for _, in := range []string{"a\u00a0b", "a\ufeffb"} {
		if got, err := ValidateLabelName(in); err != nil || got != in {
			t.Errorf("ValidateLabelName(%q) = %q, %v", in, got, err)
		}
	}
}

func TestLabelNameKeyCollisions(t *testing.T) {
	collide := [][2]string{{"Urgent", "urgent"}, {"URGENT", "urgent"}, {"Q4-Review", "q4-review"}, {"ABC xyz", "abc XYZ"}}
	for _, c := range collide {
		if labelNameKey(c[0]) != labelNameKey(c[1]) {
			t.Errorf("%q and %q must collide (ASCII case only)", c[0], c[1])
		}
	}
	distinct := [][2]string{{"Émigré", "émigré"}, {"straße", "STRASSE"}, {"İ", "i"}, {"Ω", "ω"}, {"cafe", "café"},
		{"e\u0301", "é"}, {"urgent", "urgent "}}
	for _, c := range distinct {
		if labelNameKey(c[0]) == labelNameKey(c[1]) {
			t.Errorf("%q and %q must not collide (non-ASCII is byte-exact, never normalized)", c[0], c[1])
		}
	}
	if labelNameKey("Émigré") != "Émigré" || labelNameKey("MiXeD") != "mixed" {
		t.Error("labelNameKey folds more than ASCII letters")
	}
}

func TestLabelNameSpec(t *testing.T) {
	for _, v := range []string{"owner_scope", "target_members"} {
		if got, err := ParseLabelVisibility(v); err != nil || string(got) != v {
			t.Errorf("ParseLabelVisibility(%q) = %q, %v", v, got, err)
		}
	}
	for _, v := range []string{"", "Owner_Scope", "owner_scope ", "members", "public"} {
		if _, err := ParseLabelVisibility(v); CodeOf(err) != CodeInvalidVisibility {
			t.Errorf("ParseLabelVisibility(%q) = %v; want invalid_visibility", v, err)
		}
	}
	org, grp, id := OrganizationID{1}, GroupID{2}, LabelDefinitionID{3}
	for name, c := range map[string]struct {
		sp   LabelDefinitionSpec
		want Code
	}{
		"zero id":                     {LabelDefinitionSpec{OrganizationID: org, Name: "a", Visibility: LabelVisibilityOwnerScope}, CodeInvalidID},
		"zero Organization":           {LabelDefinitionSpec{ID: id, Name: "a", Visibility: LabelVisibilityOwnerScope}, CodeInvalidID},
		"bad name":                    {LabelDefinitionSpec{ID: id, OrganizationID: org, Name: " ", Visibility: LabelVisibilityOwnerScope}, CodeInvalidName},
		"bad visibility":              {LabelDefinitionSpec{ID: id, OrganizationID: org, Name: "a", Visibility: "x"}, CodeInvalidVisibility},
		"Organization target_members": {LabelDefinitionSpec{ID: id, OrganizationID: org, Name: "a", Visibility: LabelVisibilityTargetMembers}, CodeVisibilityNotAllowed},
		"Organization owner_scope":    {LabelDefinitionSpec{ID: id, OrganizationID: org, Name: " a ", Visibility: LabelVisibilityOwnerScope}, ""},
		"Group target_members":        {LabelDefinitionSpec{ID: id, OrganizationID: org, Group: grp, Name: "a", Visibility: LabelVisibilityTargetMembers}, ""},
		"Group owner_scope":           {LabelDefinitionSpec{ID: id, OrganizationID: org, Group: grp, Name: "a", Visibility: LabelVisibilityOwnerScope}, ""},
	} {
		name1, scope, scopeID, group, err := c.sp.validate()
		if CodeOf(err) != c.want {
			t.Errorf("%s: %v, want %q", name, err, c.want)
			continue
		}
		if err != nil {
			continue
		}
		if name1 != "a" {
			t.Errorf("%s: stored name %q", name, name1)
		}
		if zero(c.sp.Group) && (scope != ScopeOrganization || scopeID != [16]byte(org) || group.Valid) {
			t.Errorf("%s: Organization scope facts %s %v %v", name, scope, scopeID, group)
		}
		if !zero(c.sp.Group) && (scope != ScopeGroup || scopeID != [16]byte(grp) || !group.Valid || group.Bytes != [16]byte(grp)) {
			t.Errorf("%s: Group scope facts %s %v %v", name, scope, scopeID, group)
		}
	}
}

// labelStatements are the label statements of labels.go (Plan r2 §1.2-§1.4, §4.1): the definition insert, read, two
// locks, page, update and two-statement delete, and the assignment read by target, insert and delete.
var labelStatements = []string{
	"sqlDeleteLabelAssignment", "sqlDeleteLabelDefinition", "sqlDeleteLabelDefinitionAssignments", "sqlInsertLabelAssignment",
	"sqlInsertLabelDefinition", "sqlLabelAssignmentsByTarget", "sqlLabelDefinitionsPage", "sqlLockLabelDefinition",
	"sqlSelectLabelDefinition", "sqlShareLabelDefinition", "sqlUpdateLabelDefinition",
}

// TestLabelStatements pins the label statements: names; only the two label relations; explicit columns (never *,
// never the generated name_key, never a column outside production storage such as a test-only generated target key);
// the Organization as organization_id = $1; the INSERT column lists equal to the runtime column grants of 00020; UPDATE
// sets name and visibility of definitions only; the two locks are the plain read plus FOR NO KEY UPDATE or FOR SHARE.
func TestLabelStatements(t *testing.T) {
	_, files := productionFiles(t)
	stmts := constStrings(t, files["labels.go"])
	var names []string
	rel := regexp.MustCompile(`app\.[a-z_]+`)
	for name, sql := range stmts {
		if !strings.HasPrefix(name, "sql") {
			continue
		}
		names = append(names, name)
		for _, r := range rel.FindAllString(sql, -1) {
			if r != "app.label_definitions" && r != "app.label_assignments" {
				t.Errorf("%s names %s", name, r)
			}
		}
		if regexp.MustCompile(`\*|\bname_key\b|_fixture|\btarget_[a-z]+_id\b`).MatchString(strings.ReplaceAll(sql, "target_group_id", "")) {
			t.Errorf("%s selects *, names name_key or names a non-production column: %s", name, sql)
		}
		if !strings.Contains(sql, "organization_id = $1") && !strings.Contains(sql, "VALUES ($2, $1, ") {
			t.Errorf("%s lacks the Organization predicate: %s", name, sql)
		}
		if strings.HasPrefix(sql, "UPDATE ") && !strings.HasPrefix(sql, "UPDATE app.label_definitions SET name = $3, visibility = $4 WHERE ") {
			t.Errorf("%s updates other columns or relations: %s", name, sql)
		}
		if regexp.MustCompile(`\bFOR (UPDATE|NO KEY UPDATE|SHARE|KEY SHARE)\b`).MatchString(sql) &&
			name != "sqlLockLabelDefinition" && name != "sqlShareLabelDefinition" {
			t.Errorf("%s takes an unpinned row lock: %s", name, sql)
		}
	}
	slices.Sort(names)
	if strings.Join(names, ",") != strings.Join(labelStatements, ",") {
		t.Fatalf("label statements:\n got %v\nwant %v", names, labelStatements)
	}
	sel := stmts["sqlSelectLabelDefinition"]
	if sel != "SELECT id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility, created_at "+
		"FROM app.label_definitions WHERE organization_id = $1 AND id = $2" {
		t.Errorf("sqlSelectLabelDefinition: %s", sel)
	}
	if stmts["sqlLockLabelDefinition"] != sel+" FOR NO KEY UPDATE" || stmts["sqlShareLabelDefinition"] != sel+" FOR SHARE" {
		t.Error("the definition locks are not the plain read FOR NO KEY UPDATE / FOR SHARE")
	}
	for name, want := range map[string]string{
		"sqlInsertLabelDefinition":            "INSERT INTO app.label_definitions (id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility) VALUES ($2, $1, $3, $4, $5, $6, $7)",
		"sqlInsertLabelAssignment":            "INSERT INTO app.label_assignments (id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id, target_group_id) VALUES ($2, $1, $3, $4, $5, $6, $7, $8)",
		"sqlDeleteLabelDefinitionAssignments": "DELETE FROM app.label_assignments WHERE organization_id = $1 AND label_id = $2",
		"sqlDeleteLabelDefinition":            "DELETE FROM app.label_definitions WHERE organization_id = $1 AND id = $2",
		"sqlDeleteLabelAssignment":            "DELETE FROM app.label_assignments WHERE organization_id = $1 AND label_id = $2 AND target_type = $3 AND target_id = $4",
	} {
		if stmts[name] != want {
			t.Errorf("%s differs from its pinned text:\n got %s\nwant %s", name, stmts[name], want)
		}
	}
	if !strings.HasSuffix(stmts["sqlLabelDefinitionsPage"], "AND id > $5 ORDER BY id LIMIT $6") ||
		!strings.Contains(stmts["sqlLabelDefinitionsPage"], "owner_scope_type = $2 AND owner_scope_id = $3") {
		t.Errorf("sqlLabelDefinitionsPage is not the owner-scope keyset page: %s", stmts["sqlLabelDefinitionsPage"])
	}
	if !strings.Contains(stmts["sqlLabelAssignmentsByTarget"], "WHERE a.organization_id = $1 AND a.target_type = $2 AND a.target_id = $3") {
		t.Errorf("sqlLabelAssignmentsByTarget is not the target read: %s", stmts["sqlLabelAssignmentsByTarget"])
	}
	// No other production file names the label relations.
	for name, f := range files {
		if name == "labels.go" {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING && strings.Contains(lit.Value, "app.label_") {
				t.Errorf("%s names a label relation: %s", name, lit.Value)
			}
			return true
		})
	}
}
