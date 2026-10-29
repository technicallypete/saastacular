//go:build integration

package access

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"core/internal/tenancy"
)

// TestLabels (G020 Plan r2 §5.2 T4 LAB-1..12 and LAB-17, T5 LAB-13..16; AC2-AC5): the guarded label facade against the
// migrated database. Subjects (Plan §1 table): carol Organization owner, frank Organization admin, grace Organization
// admin and research admin (the independent dual role), bob research owner, dave research member, dana design owner,
// mia a plain Organization member, foreign the owner of Organization B. LAB-13..16 run under the test-only fixture
// widening (labels_fixture_test.go), installed for this test and restored at its end.

type labWorld struct {
	A, B                                                    tenancy.OrganizationID
	research, design, ops, gB                               tenancy.Group
	carol, frank, grace, bob, dave, dana, mia, foreign, own *itUser
}

func (e *itEnv) labWorld(t *testing.T) *labWorld {
	t.Helper()
	c := e.canaries
	w := &labWorld{}
	w.A, w.B = e.org(t, "Labels A "+c[0]), e.org(t, "Labels B "+c[1])
	for _, u := range []**itUser{&w.carol, &w.frank, &w.grace, &w.bob, &w.dave, &w.dana, &w.mia, &w.foreign, &w.own} {
		*u = e.newUser(t)
	}
	w.research = e.group(t, w.A, w.own, "Research "+c[2], "labels-research")
	w.design = e.group(t, w.A, w.own, "Design "+c[0], "labels-design")
	w.ops = e.group(t, w.A, w.own, "Operations "+c[1], "labels-ops")
	w.gB = e.group(t, w.B, w.foreign, "Group B "+c[2], "labels-b")
	e.grantOrg(t, w.A, w.carol, tenancy.OrgOwner)
	e.grantOrg(t, w.A, w.frank, tenancy.OrgAdmin)
	e.grantOrg(t, w.A, w.grace, tenancy.OrgAdmin)
	for _, u := range []*itUser{w.bob, w.dave, w.dana, w.mia} {
		e.grantOrg(t, w.A, u, tenancy.OrgMember)
	}
	e.grantGroup(t, w.A, w.research.ID, w.grace, tenancy.GroupAdmin)
	e.grantGroup(t, w.A, w.research.ID, w.bob, tenancy.GroupOwner)
	e.grantGroup(t, w.A, w.research.ID, w.dave, tenancy.GroupMember)
	e.grantGroup(t, w.A, w.design.ID, w.dana, tenancy.GroupOwner)
	e.grantOrg(t, w.B, w.foreign, tenancy.OrgOwner)
	for _, u := range []*itUser{w.carol, w.frank, w.grace, w.bob, w.dave, w.dana, w.mia} {
		e.bind(t, u, w.A)
	}
	e.bind(t, w.foreign, w.B)
	return w
}

// at resolves u's context in Organization A: Organization-wide when g is nil, otherwise with g selected.
func (w *labWorld) at(t *testing.T, e *itEnv, u *itUser, g *tenancy.Group) (context.Context, *TenantContext) {
	t.Helper()
	if g == nil {
		return e.as(t, u, w.A)
	}
	rctx, _ := beginAs(t, e.svc, context.Background(), u, w.A)
	return rctx, resolve(t, e.svc, rctx, w.A, g.ID)
}

var labOrgScope = Scope{Kind: tenancy.ScopeOrganization}

func labGroupScope(g tenancy.Group) Scope {
	id := g.ID
	return Scope{Kind: tenancy.ScopeGroup, Group: &id}
}

func labGroupTarget(g tenancy.Group) LabelTargetSelector {
	return LabelTargetSelector{Group: g.ID.String()}
}

func labFixtureTarget(id string) LabelTargetSelector {
	return LabelTargetSelector{Kind: string(labelTargetFixture), ID: id}
}

// wantLabel requires err to be exactly code with field ("" code: success).
func wantLabel(t *testing.T, what string, err error, code Code, field string) {
	t.Helper()
	if code == "" {
		if err != nil {
			t.Errorf("%s: %v (field %q), want success", what, err, labField(err))
		}
		return
	}
	if CodeOf(err) != code || labField(err) != field {
		t.Errorf("%s: %v (field %q), want %s %q", what, err, labField(err), code, field)
	}
}

func labField(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Field
	}
	return ""
}

// labCount runs one superuser count.
func (e *itEnv) labCount(t *testing.T, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("superuser count: %v (SQLSTATE %s)", err, pgState(err))
	}
	return n
}

const (
	labDefByID   = `SELECT count(*) FROM app.label_definitions WHERE id = $1::uuid`
	labAsgByDef  = `SELECT count(*) FROM app.label_assignments WHERE label_id = $1::uuid`
	labAsgByKey  = `SELECT count(*) FROM app.label_assignments WHERE label_id = $1::uuid AND target_type = $2 AND target_id = $3::uuid`
	labAudit     = `SELECT count(*) FROM audit.events WHERE action = $1 AND resource_type = 'label' AND resource_id = $2::uuid`
	labAuditGrp  = `SELECT count(*) FROM audit.events WHERE action = $1 AND resource_type = 'label' AND resource_id = $2::uuid AND group_id = $3::uuid`
	labAuditWhy  = `SELECT count(*) FROM audit.events WHERE action = $1 AND resource_type = 'label' AND resource_id = $2::uuid AND reason = $3`
	labAuditAny  = `SELECT count(*) FROM audit.events WHERE resource_id = $1::uuid`
	labNameInOrg = `SELECT count(*) FROM app.label_definitions WHERE organization_id = $1::text::uuid AND name = $2`
)

func labIDs(vs []LabelAssignmentView) []string {
	out := []string{}
	for _, v := range vs {
		out = append(out, v.LabelID.String())
	}
	return out
}

func labSorted(ids ...string) []string {
	out := slices.Clone(ids)
	slices.Sort(out)
	return out
}

func labPageHas(p LabelDefinitionPage, id tenancy.LabelDefinitionID) (LabelDefinitionView, bool) {
	for _, v := range p.Labels {
		if v.ID == id {
			return v, true
		}
	}
	return LabelDefinitionView{}, false
}

// labAll pages through one scope with limit 1 and returns every definition id in page order (ascending ids).
func (e *itEnv) labAll(t *testing.T, c context.Context, tc *TenantContext, sc Scope) ([]string, error) {
	t.Helper()
	var ids []string
	after := ""
	for i := 0; i < 1000; i++ {
		p, err := e.svc.ListLabelDefinitions(c, tc, sc, after, 1)
		if err != nil {
			return ids, err
		}
		if len(p.Labels) > 1 {
			t.Errorf("a page of limit 1 holds %d definitions", len(p.Labels))
		}
		for _, v := range p.Labels {
			ids = append(ids, v.ID.String())
		}
		if p.Next == nil {
			return ids, nil
		}
		after = p.Next.String()
	}
	t.Fatal("definition paging does not terminate")
	return nil, nil
}

func TestLabels(t *testing.T) {
	e := setupIT(t)
	w := e.labWorld(t)
	c := e.canaries
	svc := e.svc
	sel := func(t *testing.T, u *itUser, g *tenancy.Group) (context.Context, *TenantContext) {
		t.Helper()
		return w.at(t, e, u, g)
	}
	create := func(t *testing.T, u *itUser, sc Scope, name, vis string) LabelDefinitionView {
		t.Helper()
		cx, tc := sel(t, u, nil)
		v, err := svc.CreateLabelDefinition(cx, tc, sc, name, vis)
		if err != nil {
			t.Fatalf("create %q: %v (field %q)", name, err, labField(err))
		}
		return v
	}
	assign := func(t *testing.T, u *itUser, id tenancy.LabelDefinitionID, tg LabelTargetSelector) {
		t.Helper()
		cx, tc := sel(t, u, nil)
		if err := svc.AssignLabel(cx, tc, id.String(), tg); err != nil {
			t.Fatalf("assign: %v (field %q)", err, labField(err))
		}
	}

	t.Run("LAB-1 Organization-scope lifecycle: create, list and page, rename, no-op, delete (carol)", func(t *testing.T) {
		cx, tc := sel(t, w.carol, nil)
		v, err := svc.CreateLabelDefinition(cx, tc, labOrgScope, "Priority "+c[0], "owner_scope")
		wantLabel(t, "create", err, "", "")
		if v.Scope != tenancy.ScopeOrganization || v.Group != nil || v.Name != "Priority "+c[0] || v.Visibility != tenancy.LabelVisibilityOwnerScope ||
			zero(v.ID) {
			t.Fatalf("created view %+v", v)
		}
		id := v.ID.String()
		if e.labCount(t, labDefByID, id) != 1 || e.labCount(t, labAudit, "label.created", id) != 1 {
			t.Error("the created definition or its one label.created record is missing")
		}
		v2 := create(t, w.carol, labOrgScope, "Escalated "+c[1], "owner_scope")
		p, err := svc.ListLabelDefinitions(cx, tc, labOrgScope, "", 0)
		wantLabel(t, "list", err, "", "")
		if got, ok := labPageHas(p, v.ID); !ok || !labSame(got, v) {
			t.Errorf("list: %+v, want %+v", got, v)
		}
		ids, err := e.labAll(t, cx, tc, labOrgScope)
		wantLabel(t, "page through", err, "", "")
		if !slices.IsSorted(ids) || !slices.Contains(ids, id) || !slices.Contains(ids, v2.ID.String()) {
			t.Errorf("paged ids %v: want ascending and both definitions", ids)
		}
		u, err := svc.UpdateLabelDefinition(cx, tc, id, "Priority renamed "+c[0], "owner_scope")
		wantLabel(t, "rename", err, "", "")
		if u.ID != v.ID || u.Name != "Priority renamed "+c[0] || u.Scope != tenancy.ScopeOrganization {
			t.Errorf("renamed view %+v", u)
		}
		if e.labCount(t, labAuditWhy, "label.updated", id, "renamed") != 1 {
			t.Error("rename: want exactly one label.updated renamed record")
		}
		before := e.labCount(t, labAuditAny, id)
		if _, err := svc.UpdateLabelDefinition(cx, tc, id, "Priority renamed "+c[0], "owner_scope"); err != nil {
			t.Errorf("no-op update: %v", err)
		}
		if e.labCount(t, labAuditAny, id) != before {
			t.Error("a no-op update recorded an audit row")
		}
		d, err := svc.DeleteLabelDefinition(cx, tc, id)
		wantLabel(t, "delete", err, "", "")
		if d.AssignmentsRemoved != 0 || e.labCount(t, labDefByID, id) != 0 || e.labCount(t, labAudit, "label.deleted", id) != 1 {
			t.Errorf("delete: %+v; the definition must be gone with one label.deleted record", d)
		}
		p, err = svc.ListLabelDefinitions(cx, tc, labOrgScope, "", 0)
		if _, ok := labPageHas(p, v.ID); err != nil || ok {
			t.Errorf("list after delete: %v, present %t", err, ok)
		}
		if _, err := svc.DeleteLabelDefinition(cx, tc, v2.ID.String()); err != nil {
			t.Errorf("cleanup delete: %v", err)
		}
	})

	t.Run("LAB-2 Group-scope lifecycle: create, list, visibility change, delete (bob in research)", func(t *testing.T) {
		cx, tc := sel(t, w.bob, nil)
		v, err := svc.CreateLabelDefinition(cx, tc, labGroupScope(w.research), "Research queue "+c[1], "owner_scope")
		wantLabel(t, "create", err, "", "")
		if v.Scope != tenancy.ScopeGroup || v.Group == nil || *v.Group != w.research.ID || v.Visibility != tenancy.LabelVisibilityOwnerScope {
			t.Fatalf("created view %+v", v)
		}
		id := v.ID.String()
		if e.labCount(t, labAuditGrp, "label.created", id, w.research.ID.String()) != 1 {
			t.Error("want one label.created record with the owning Group")
		}
		u, err := svc.UpdateLabelDefinition(cx, tc, id, v.Name, "target_members")
		wantLabel(t, "visibility change", err, "", "")
		if u.Visibility != tenancy.LabelVisibilityTargetMembers || e.labCount(t, labAuditWhy, "label.updated", id, "visibility_changed") != 1 {
			t.Errorf("visibility change: %+v; want one label.updated visibility_changed record", u)
		}
		p, err := svc.ListLabelDefinitions(cx, tc, labGroupScope(w.research), "", 0)
		if got, ok := labPageHas(p, v.ID); err != nil || !ok || !labSame(got, u) {
			t.Errorf("list: %v %+v, want %+v", err, got, u)
		}
		// With research selected the Group scope is the same.
		scx, stc := sel(t, w.bob, &w.research)
		v2, err := svc.CreateLabelDefinition(scx, stc, labGroupScope(w.research), "Research selected "+c[2], "owner_scope")
		wantLabel(t, "create with research selected", err, "", "")
		d, err := svc.DeleteLabelDefinition(cx, tc, id)
		wantLabel(t, "delete", err, "", "")
		if d.AssignmentsRemoved != 0 || e.labCount(t, labDefByID, id) != 0 ||
			e.labCount(t, labAuditGrp, "label.deleted", id, w.research.ID.String()) != 1 {
			t.Errorf("delete: %+v", d)
		}
		if _, err := svc.DeleteLabelDefinition(scx, stc, v2.ID.String()); err != nil {
			t.Errorf("delete with research selected: %v", err)
		}
	})

	t.Run("LAB-3 name collisions: ASCII case-insensitive, non-ASCII exact, scoped", func(t *testing.T) {
		cx, tc := sel(t, w.carol, nil)
		a := create(t, w.carol, labOrgScope, "Alpha Label", "owner_scope")
		for _, n := range []string{"alpha label", "ALPHA LABEL", "Alpha Label"} {
			_, err := svc.CreateLabelDefinition(cx, tc, labOrgScope, n, "owner_scope")
			wantLabel(t, "create "+n, err, CodeConflict, fieldNameConflict)
		}
		if got := e.labCount(t, `SELECT count(*) FROM app.label_definitions WHERE organization_id = $1::text::uuid AND name_key = 'alpha label'`, w.A.String()); got != 1 {
			t.Errorf("%d rows under the key 'alpha label', want 1 (zero writes)", got)
		}
		e1 := create(t, w.carol, labOrgScope, "Été", "owner_scope")
		e2 := create(t, w.carol, labOrgScope, "été", "owner_scope") // a non-ASCII case difference is a different name
		_, err := svc.CreateLabelDefinition(cx, tc, labOrgScope, "Été", "owner_scope")
		wantLabel(t, "create Été twice", err, CodeConflict, fieldNameConflict)
		_, err = svc.UpdateLabelDefinition(cx, tc, e2.ID.String(), "ALPHA label", "owner_scope")
		wantLabel(t, "rename into a collision", err, CodeConflict, fieldNameConflict)
		g := create(t, w.bob, labGroupScope(w.research), "Alpha Label", "owner_scope") // another owner scope
		bcx, btc := sel(t, w.bob, nil)
		for _, d := range []struct {
			cx context.Context
			tc *TenantContext
			id tenancy.LabelDefinitionID
		}{{cx, tc, a.ID}, {cx, tc, e1.ID}, {cx, tc, e2.ID}, {bcx, btc, g.ID}} {
			if _, err := svc.DeleteLabelDefinition(d.cx, d.tc, d.id.String()); err != nil {
				t.Errorf("cleanup delete: %v", err)
			}
		}
	})

	t.Run("LAB-4 at most 200 definitions per owner scope: the 201st is scope_full", func(t *testing.T) {
		bulk := e.group(t, w.A, w.own, "Bulk "+c[0], "labels-bulk")
		e.grantGroup(t, w.A, bulk.ID, w.bob, tenancy.GroupOwner)
		defer func() {
			if _, err := e.setup.Exec(context.Background(), `DELETE FROM app.label_definitions WHERE group_id = $1::text::uuid`, bulk.ID.String()); err != nil {
				t.Errorf("remove the bulk definitions: %v", err)
			}
		}()
		if _, err := e.setup.Exec(context.Background(), `INSERT INTO app.label_definitions (id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility)
SELECT gen_random_uuid(), $1::text::uuid, 'group', $2::text::uuid, $2::text::uuid, 'Bulk ' || n, 'owner_scope' FROM generate_series(1, 199) n`,
			w.A.String(), bulk.ID.String()); err != nil {
			t.Fatalf("bulk fixture: %v", err)
		}
		cx, tc := sel(t, w.bob, nil)
		_, err := svc.CreateLabelDefinition(cx, tc, labGroupScope(bulk), "Bulk 200", "owner_scope")
		wantLabel(t, "the 200th definition", err, "", "")
		_, err = svc.CreateLabelDefinition(cx, tc, labGroupScope(bulk), "Bulk 201", "owner_scope")
		wantLabel(t, "the 201st definition", err, CodeConflict, fieldScopeFull)
		if got := e.labCount(t, `SELECT count(*) FROM app.label_definitions WHERE group_id = $1::text::uuid`, bulk.ID.String()); got != 200 {
			t.Errorf("%d definitions in the scope, want 200", got)
		}
	})

	t.Run("LAB-5 input refusals before any write: visibility_not_allowed, invalid visibility/name/id/cursor/page", func(t *testing.T) {
		cx, tc := sel(t, w.carol, nil)
		_, err := svc.CreateLabelDefinition(cx, tc, labOrgScope, "Members see "+c[0], "target_members")
		wantLabel(t, "Organization target_members", err, CodeInvalidInput, fieldVisibilityNotAllowed)
		if e.labCount(t, labNameInOrg, w.A.String(), "Members see "+c[0]) != 0 {
			t.Error("a refused Organization target_members definition was written")
		}
		o := create(t, w.carol, labOrgScope, "Owner only "+c[1], "owner_scope")
		_, err = svc.UpdateLabelDefinition(cx, tc, o.ID.String(), o.Name, "target_members")
		wantLabel(t, "Organization target_members update", err, CodeInvalidInput, fieldVisibilityNotAllowed)
		_, err = svc.CreateLabelDefinition(cx, tc, labOrgScope, "Public "+c[2], "public")
		wantLabel(t, "unknown visibility", err, CodeInvalidInput, string(tenancy.CodeInvalidVisibility))
		for _, n := range []string{"", "   ", strings.Repeat("x", tenancy.MaxLabelNameLength+1), "Tab\there", "a\u2028b"} {
			_, err = svc.CreateLabelDefinition(cx, tc, labOrgScope, n, "owner_scope")
			wantLabel(t, "invalid name", err, CodeInvalidInput, string(tenancy.CodeInvalidName))
		}
		if e.labCount(t, `SELECT count(*) FROM app.label_definitions WHERE organization_id = $1::text::uuid AND name LIKE 'Tab%'`, w.A.String()) != 0 {
			t.Error("a refused name was written")
		}
		// Plan r2 E18 / tenancy.ValidateName (TestLabelNameTrim): surrounding whitespace is not a refusal; the name is
		// stored in its trimmed form, which the database CHECK (name = btrim(name)) then admits.
		p, err := svc.CreateLabelDefinition(cx, tc, labOrgScope, "  Padded "+c[2]+"\t", "owner_scope")
		wantLabel(t, "a padded name", err, "", "")
		if err == nil {
			if p.Name != "Padded "+c[2] {
				t.Errorf("padded name stored as %q, want %q", p.Name, "Padded "+c[2])
			}
			if _, err := svc.DeleteLabelDefinition(cx, tc, p.ID.String()); err != nil {
				t.Errorf("cleanup delete: %v", err)
			}
		}
		_, err = svc.CreateLabelDefinition(cx, tc, labOrgScope, strings.Repeat("x", tenancy.MaxLabelNameLength), "owner_scope")
		wantLabel(t, "an 80-code-point name", err, "", "")
		_, err = svc.UpdateLabelDefinition(cx, tc, "not-a-label", "Name", "owner_scope")
		wantLabel(t, "malformed id (update)", err, CodeInvalidInput, string(tenancy.CodeInvalidID))
		_, err = svc.DeleteLabelDefinition(cx, tc, "not-a-label")
		wantLabel(t, "malformed id (delete)", err, CodeInvalidInput, string(tenancy.CodeInvalidID))
		_, err = svc.ListLabelDefinitions(cx, tc, labOrgScope, "not-a-cursor", 0)
		wantLabel(t, "malformed cursor", err, CodeInvalidInput, "invalid_cursor")
		for _, n := range []int{-1, tenancy.MaxLabelPage + 1} {
			_, err = svc.ListLabelDefinitions(cx, tc, labOrgScope, "", n)
			wantLabel(t, "page size", err, CodeInvalidInput, "invalid_page_size")
		}
		if _, err := e.setup.Exec(context.Background(), `DELETE FROM app.label_definitions WHERE organization_id = $1::text::uuid AND name = $2`,
			w.A.String(), strings.Repeat("x", tenancy.MaxLabelNameLength)); err != nil {
			t.Error(err)
		}
		if _, err := svc.DeleteLabelDefinition(cx, tc, o.ID.String()); err != nil {
			t.Errorf("cleanup delete: %v", err)
		}
	})

	t.Run("LAB-6..8 assign/unassign the real Group target; already_assigned; not_assigned", func(t *testing.T) {
		cx, tc := sel(t, w.carol, nil)
		l := create(t, w.carol, labOrgScope, "Assigned "+c[0], "owner_scope")
		id := l.ID.String()
		wantLabel(t, "LAB-6 assign research", svc.AssignLabel(cx, tc, id, labGroupTarget(w.research)), "", "")
		if e.labCount(t, labAsgByKey, id, "group", w.research.ID.String()) != 1 ||
			e.labCount(t, labAuditGrp, "label.assigned", id, w.research.ID.String()) != 1 {
			t.Error("LAB-6: want one assignment row and one label.assigned record with the target Group")
		}
		got, err := svc.ListLabelAssignments(cx, tc, labGroupTarget(w.research))
		if err != nil || !slices.Contains(labIDs(got), id) {
			t.Errorf("LAB-6 list: %v %v", got, err)
		}
		for _, a := range got {
			if a.LabelID == l.ID && a.Name != l.Name {
				t.Errorf("LAB-6 list name %q, want %q", a.Name, l.Name)
			}
		}
		wantLabel(t, "LAB-7 assign twice", svc.AssignLabel(cx, tc, id, labGroupTarget(w.research)), CodeConflict, fieldAlreadyAssigned)
		if e.labCount(t, labAsgByKey, id, "group", w.research.ID.String()) != 1 ||
			e.labCount(t, labAuditGrp, "label.assigned", id, w.research.ID.String()) != 1 {
			t.Error("LAB-7: a duplicate assign wrote a row or an audit record")
		}
		wantLabel(t, "LAB-6 unassign", svc.UnassignLabel(cx, tc, id, labGroupTarget(w.research)), "", "")
		if e.labCount(t, labAsgByKey, id, "group", w.research.ID.String()) != 0 ||
			e.labCount(t, labAuditGrp, "label.unassigned", id, w.research.ID.String()) != 1 {
			t.Error("LAB-6: want the row removed and one label.unassigned record")
		}
		got, err = svc.ListLabelAssignments(cx, tc, labGroupTarget(w.research))
		if err != nil || slices.Contains(labIDs(got), id) {
			t.Errorf("LAB-6 list after unassign: %v %v", got, err)
		}
		wantLabel(t, "LAB-8 unassign twice", svc.UnassignLabel(cx, tc, id, labGroupTarget(w.research)), CodeConflict, fieldNotAssigned)
		if e.labCount(t, labAudit, "label.unassigned", id) != 1 {
			t.Error("LAB-8: an absent unassign recorded an audit row")
		}
		if _, err := svc.DeleteLabelDefinition(cx, tc, id); err != nil {
			t.Errorf("cleanup delete: %v", err)
		}
	})

	t.Run("LAB-9 delete removes the definition's assignments in the same transaction, with the count", func(t *testing.T) {
		cx, tc := sel(t, w.carol, nil)
		l := create(t, w.carol, labOrgScope, "Three Groups "+c[1], "owner_scope")
		id := l.ID.String()
		for _, g := range []tenancy.Group{w.research, w.design, w.ops} {
			assign(t, w.carol, l.ID, labGroupTarget(g))
		}
		d, err := svc.DeleteLabelDefinition(cx, tc, id)
		wantLabel(t, "delete", err, "", "")
		if d.AssignmentsRemoved != 3 || e.labCount(t, labAsgByDef, id) != 0 || e.labCount(t, labDefByID, id) != 0 {
			t.Errorf("delete: %+v; want 3 removed, no assignment and no definition left", d)
		}
		if e.labCount(t, labAudit, "label.unassigned", id) != 3 || e.labCount(t, labAudit, "label.deleted", id) != 1 {
			t.Error("want three label.unassigned records and one label.deleted")
		}
	})

	t.Run("LAB-10 targets: foreign, unknown, cross-Group, unsupported, ambiguous, malformed", func(t *testing.T) {
		cx, tc := sel(t, w.carol, nil)
		l := create(t, w.carol, labOrgScope, "Targets "+c[2], "owner_scope")
		g := create(t, w.bob, labGroupScope(w.research), "Group targets "+c[2], "owner_scope")
		id := l.ID.String()
		wantLabel(t, "foreign Organization's Group", svc.AssignLabel(cx, tc, id, labGroupTarget(w.gB)), CodeNotFound, "")
		wantLabel(t, "unknown Group", svc.AssignLabel(cx, tc, id, LabelTargetSelector{Group: rlsUUID(t)}), CodeNotFound, "")
		rcx, rtc := sel(t, w.carol, &w.research)
		wantLabel(t, "cross-Group (design with research selected)", svc.AssignLabel(rcx, rtc, id, labGroupTarget(w.design)), CodeNotFound, "")
		_, err := svc.ListLabelAssignments(rcx, rtc, labGroupTarget(w.design))
		wantLabel(t, "cross-Group list", err, CodeNotFound, "")
		wantLabel(t, "unregistered kind", svc.AssignLabel(cx, tc, id, LabelTargetSelector{Kind: "document", ID: rlsUUID(t)}), CodeInvalidInput, fieldTargetUnsupported)
		wantLabel(t, "no target", svc.AssignLabel(cx, tc, id, LabelTargetSelector{}), CodeInvalidInput, fieldTargetUnsupported)
		bcx, btc := sel(t, w.bob, nil)
		wantLabel(t, "Group-owned definition on a Group", svc.AssignLabel(bcx, btc, g.ID.String(), labGroupTarget(w.research)),
			CodeInvalidInput, fieldTargetUnsupported)
		wantLabel(t, "ambiguous selector", svc.AssignLabel(cx, tc, id, LabelTargetSelector{Group: w.research.ID.String(), Kind: "group", ID: w.research.ID.String()}),
			CodeInvalidInput, fieldTargetAmbiguous)
		wantLabel(t, "malformed target id", svc.AssignLabel(cx, tc, id, LabelTargetSelector{Group: "not-a-group"}), CodeInvalidInput, string(tenancy.CodeInvalidID))
		wantLabel(t, "malformed definition id", svc.AssignLabel(cx, tc, "not-a-label", labGroupTarget(w.research)), CodeInvalidInput, string(tenancy.CodeInvalidID))
		wantLabel(t, "unknown definition", svc.AssignLabel(cx, tc, rlsUUID(t), labGroupTarget(w.research)), CodeNotFound, "")
		fcx, ftc := e.as(t, w.foreign, w.B)
		fl, err := svc.CreateLabelDefinition(fcx, ftc, labOrgScope, "Foreign "+c[0], "owner_scope")
		wantLabel(t, "foreign create", err, "", "")
		wantLabel(t, "foreign definition", svc.AssignLabel(cx, tc, fl.ID.String(), labGroupTarget(w.research)), CodeNotFound, "")
		_, err = svc.UpdateLabelDefinition(cx, tc, fl.ID.String(), "Taken over", "owner_scope")
		wantLabel(t, "foreign definition update", err, CodeNotFound, "")
		if e.labAuditOf(t, l.ID) != 1 || e.labCount(t, labAsgByDef, id) != 0 || e.labCount(t, labAsgByDef, fl.ID.String()) != 0 {
			t.Error("a refused target wrote a row or an audit record")
		}
		for _, d := range []struct {
			cx context.Context
			tc *TenantContext
			id tenancy.LabelDefinitionID
		}{{cx, tc, l.ID}, {bcx, btc, g.ID}, {fcx, ftc, fl.ID}} {
			if _, err := svc.DeleteLabelDefinition(d.cx, d.tc, d.id.String()); err != nil {
				t.Errorf("cleanup delete: %v", err)
			}
		}
	})

	t.Run("LAB-11 Organization-owned labels: Organization owner/admin only; Group roles and the override confer nothing", func(t *testing.T) {
		l := create(t, w.carol, labOrgScope, "Org private "+c[0], "owner_scope")
		assign(t, w.carol, l.ID, labGroupTarget(w.research))
		id := l.ID.String()
		for _, u := range []*itUser{w.carol, w.frank, w.grace} {
			cx, tc := sel(t, u, nil)
			p, err := svc.ListLabelDefinitions(cx, tc, labOrgScope, "", 0)
			if _, ok := labPageHas(p, l.ID); err != nil || !ok {
				t.Errorf("%s list Organization labels: %v, present %t", u.sub, err, ok)
			}
			got, err := svc.ListLabelAssignments(cx, tc, labGroupTarget(w.research))
			if err != nil || !slices.Contains(labIDs(got), id) {
				t.Errorf("%s list research assignments: %v %v", u.sub, got, err)
			}
		}
		// grace (Organization admin) writes Organization labels through the Organization path.
		gcx, gtc := sel(t, w.grace, nil)
		wantLabel(t, "grace assigns design", svc.AssignLabel(gcx, gtc, id, labGroupTarget(w.design)), "", "")
		wantLabel(t, "grace unassigns design", svc.UnassignLabel(gcx, gtc, id, labGroupTarget(w.design)), "", "")
		for _, u := range []*itUser{w.bob, w.dave, w.dana, w.mia} {
			cx, tc := sel(t, u, nil)
			_, err := svc.ListLabelDefinitions(cx, tc, labOrgScope, "", 0)
			wantLabel(t, u.sub+" list Organization labels", err, CodeForbidden, "")
			_, err = svc.CreateLabelDefinition(cx, tc, labOrgScope, "Not mine "+u.sub, "owner_scope")
			wantLabel(t, u.sub+" create an Organization label", err, CodeForbidden, "")
			_, err = svc.UpdateLabelDefinition(cx, tc, id, "Taken over", "owner_scope")
			wantLabel(t, u.sub+" update an Organization label", err, CodeNotFound, "")
			_, err = svc.DeleteLabelDefinition(cx, tc, id)
			wantLabel(t, u.sub+" delete an Organization label", err, CodeNotFound, "")
			wantLabel(t, u.sub+" assign an Organization label", svc.AssignLabel(cx, tc, id, labGroupTarget(w.design)), CodeNotFound, "")
		}
		for _, u := range []*itUser{w.bob, w.dave} {
			cx, tc := sel(t, u, nil)
			_, err := svc.ListLabelAssignments(cx, tc, labGroupTarget(w.research))
			wantLabel(t, u.sub+" list research's Organization-label assignments", err, CodeForbidden, "")
		}
		off := e.withOverride(t, w.carol, w.A)
		fcx, ftc := sel(t, w.frank, nil)
		p, err := svc.ListLabelDefinitions(fcx, ftc, labOrgScope, "", 0)
		if _, ok := labPageHas(p, l.ID); err != nil || !ok {
			t.Errorf("frank with the override on: %v, present %t (the override is irrelevant to Organization labels)", err, ok)
		}
		off()
		if e.labCount(t, labNameInOrg, w.A.String(), "Not mine "+w.bob.sub) != 0 || e.labCount(t, labDefByID, id) != 1 {
			t.Error("a refused Organization-label write took effect")
		}
		ccx, ctc := sel(t, w.carol, nil)
		if d, err := svc.DeleteLabelDefinition(ccx, ctc, id); err != nil || d.AssignmentsRemoved != 1 {
			t.Errorf("cleanup delete: %+v %v", d, err)
		}
	})

	t.Run("LAB-12 Group-owned labels: exact owning-Group roles only; members read target_members; staff and the override nothing", func(t *testing.T) {
		own := create(t, w.bob, labGroupScope(w.research), "Research private "+c[1], "owner_scope")
		tm := create(t, w.bob, labGroupScope(w.research), "Research shared "+c[1], "target_members")
		dz := create(t, w.dana, labGroupScope(w.design), "Design private "+c[1], "target_members")
		for _, u := range []*itUser{w.bob, w.grace} {
			cx, tc := sel(t, u, nil)
			p, err := svc.ListLabelDefinitions(cx, tc, labGroupScope(w.research), "", 0)
			_, ok1 := labPageHas(p, own.ID)
			_, ok2 := labPageHas(p, tm.ID)
			if err != nil || !ok1 || !ok2 {
				t.Errorf("%s (research manager) list: %v %t %t", u.sub, err, ok1, ok2)
			}
		}
		gcx, gtc := sel(t, w.grace, nil)
		gu, err := svc.UpdateLabelDefinition(gcx, gtc, own.ID.String(), "Research private renamed "+c[1], "owner_scope")
		wantLabel(t, "grace (research admin) renames", err, "", "")
		own = gu
		_, err = svc.ListLabelDefinitions(gcx, gtc, labGroupScope(w.design), "", 0)
		wantLabel(t, "grace lists design (no design role; Organization admin confers nothing)", err, CodeNotFound, "")
		_, err = svc.UpdateLabelDefinition(gcx, gtc, dz.ID.String(), "Taken over", "target_members")
		wantLabel(t, "grace updates design's label", err, CodeNotFound, "")
		dcx, dtc := sel(t, w.dave, nil)
		p, err := svc.ListLabelDefinitions(dcx, dtc, labGroupScope(w.research), "", 0)
		_, okOwn := labPageHas(p, own.ID)
		_, okTm := labPageHas(p, tm.ID)
		if err != nil || okOwn || !okTm {
			t.Errorf("dave (research member) list: %v; owner_scope present %t (want false), target_members present %t (want true)", err, okOwn, okTm)
		}
		_, err = svc.CreateLabelDefinition(dcx, dtc, labGroupScope(w.research), "Member made "+c[1], "owner_scope")
		wantLabel(t, "dave creates", err, CodeForbidden, "")
		_, err = svc.UpdateLabelDefinition(dcx, dtc, tm.ID.String(), "Member renamed", "target_members")
		wantLabel(t, "dave updates a visible target_members label", err, CodeForbidden, "")
		_, err = svc.UpdateLabelDefinition(dcx, dtc, own.ID.String(), "Member renamed", "owner_scope")
		wantLabel(t, "dave updates an owner_scope label", err, CodeNotFound, "")
		_, err = svc.DeleteLabelDefinition(dcx, dtc, tm.ID.String())
		wantLabel(t, "dave deletes a visible target_members label", err, CodeForbidden, "")
		check := func(t *testing.T, who string, cx context.Context, tc *TenantContext) {
			t.Helper()
			_, err := svc.ListLabelDefinitions(cx, tc, labGroupScope(w.research), "", 0)
			wantLabel(t, who+" lists research", err, CodeNotFound, "")
			_, err = svc.CreateLabelDefinition(cx, tc, labGroupScope(w.research), "Staff made "+c[1], "owner_scope")
			wantLabel(t, who+" creates in research", err, CodeNotFound, "")
			_, err = svc.UpdateLabelDefinition(cx, tc, tm.ID.String(), "Staff renamed", "target_members")
			wantLabel(t, who+" updates research's label", err, CodeNotFound, "")
			_, err = svc.DeleteLabelDefinition(cx, tc, own.ID.String())
			wantLabel(t, who+" deletes research's label", err, CodeNotFound, "")
		}
		for _, u := range []*itUser{w.carol, w.frank, w.dana, w.mia} {
			cx, tc := sel(t, u, nil)
			check(t, u.sub, cx, tc)
		}
		off := e.withOverride(t, w.carol, w.A)
		for _, u := range []*itUser{w.frank, w.carol} {
			cx, tc := sel(t, u, nil)
			check(t, u.sub+" with the override on", cx, tc)
		}
		off()
		if e.labCount(t, labNameInOrg, w.A.String(), "Staff made "+c[1]) != 0 || e.labCount(t, labNameInOrg, w.A.String(), "Member made "+c[1]) != 0 ||
			e.labCount(t, labDefByID, own.ID.String()) != 1 || e.labCount(t, labDefByID, tm.ID.String()) != 1 {
			t.Error("a refused Group-label write took effect")
		}
		bcx, btc := sel(t, w.bob, nil)
		for _, id := range []tenancy.LabelDefinitionID{own.ID, tm.ID} {
			if _, err := svc.DeleteLabelDefinition(bcx, btc, id.String()); err != nil {
				t.Errorf("cleanup delete: %v", err)
			}
		}
		ncx, ntc := sel(t, w.dana, nil)
		if _, err := svc.DeleteLabelDefinition(ncx, ntc, dz.ID.String()); err != nil {
			t.Errorf("cleanup delete: %v", err)
		}
	})

	// --- T5: the fixture kind under the test-only widening ----------------------------------------------------------
	_, teardown := e.installLabelFixture(t)
	fxR := e.fixtureTarget(t, w.A, w.research.ID, true)       // research-owned, members may read
	fxHidden := e.fixtureTarget(t, w.A, w.research.ID, false) // research-owned, managers only
	fxD := e.fixtureTarget(t, w.A, w.design.ID, true)         // design-owned
	rOwn := create(t, w.bob, labGroupScope(w.research), "Fixture private "+c[2], "owner_scope")
	rTm := create(t, w.bob, labGroupScope(w.research), "Fixture shared "+c[2], "target_members")
	orgL := create(t, w.carol, labOrgScope, "Fixture org "+c[2], "owner_scope")
	removeGroupRole := func(t *testing.T, u *itUser, g tenancy.GroupID) {
		t.Helper()
		if _, err := e.setup.Exec(context.Background(), `DELETE FROM app.role_bindings WHERE subject_user_id = $1::text::uuid AND scope_type = 'group'
  AND scope_id = $2::text::uuid`, u.id.String(), g.String()); err != nil {
			t.Errorf("remove the per-case Group role: %v", err)
		}
	}

	t.Run("LAB-13 fixture: Group-owned assignment by the owning Group's manager; a design-owned target is not_found", func(t *testing.T) {
		cx, tc := sel(t, w.bob, nil)
		id := rOwn.ID.String()
		wantLabel(t, "bob assigns a research target", svc.AssignLabel(cx, tc, id, labFixtureTarget(fxR)), "", "")
		if e.labCount(t, labAsgByKey, id, "fixture", fxR) != 1 || e.labCount(t, labAuditGrp, "label.assigned", id, w.research.ID.String()) != 1 {
			t.Error("want one fixture assignment row and one label.assigned record with the owning Group")
		}
		got, err := svc.ListLabelAssignments(cx, tc, labFixtureTarget(fxR))
		if err != nil || !slices.Equal(labIDs(got), []string{id}) {
			t.Errorf("bob lists the target: %v %v", labIDs(got), err)
		}
		wantLabel(t, "bob assigns twice", svc.AssignLabel(cx, tc, id, labFixtureTarget(fxR)), CodeConflict, fieldAlreadyAssigned)
		wantLabel(t, "bob assigns a design target", svc.AssignLabel(cx, tc, id, labFixtureTarget(fxD)), CodeNotFound, "")
		wantLabel(t, "bob assigns an unknown target", svc.AssignLabel(cx, tc, id, labFixtureTarget(rlsUUID(t))), CodeNotFound, "")
		// The same cross-Group placement by a subject who manages both Groups: resolve returns design, the definition is
		// research's, so core refuses it (T5 (d) is the SQL half).
		e.grantGroup(t, w.A, w.research.ID, w.dana, tenancy.GroupAdmin)
		ncx, ntc := sel(t, w.dana, nil)
		wantLabel(t, "dana (design owner, research admin) assigns research's label to a design target",
			svc.AssignLabel(ncx, ntc, id, labFixtureTarget(fxD)), CodeNotFound, "")
		removeGroupRole(t, w.dana, w.research.ID)
		if e.labCount(t, labAsgByKey, id, "fixture", fxD) != 0 {
			t.Error("a design-owned target received research's label")
		}
		ccx, ctc := sel(t, w.carol, nil)
		wantLabel(t, "an Organization-owned definition on a fixture target", svc.AssignLabel(ccx, ctc, orgL.ID.String(), labFixtureTarget(fxR)),
			CodeInvalidInput, fieldTargetUnsupported)
		wantLabel(t, "ambiguous fixture selector", svc.AssignLabel(cx, tc, id, LabelTargetSelector{Group: w.research.ID.String(), Kind: string(labelTargetFixture), ID: fxR}),
			CodeInvalidInput, fieldTargetAmbiguous)
		wantLabel(t, "bob unassigns", svc.UnassignLabel(cx, tc, id, labFixtureTarget(fxR)), "", "")
		wantLabel(t, "bob unassigns twice", svc.UnassignLabel(cx, tc, id, labFixtureTarget(fxR)), CodeConflict, fieldNotAssigned)
		if e.labCount(t, labAsgByDef, id) != 0 || e.labCount(t, labAuditGrp, "label.unassigned", id, w.research.ID.String()) != 1 {
			t.Error("want the row removed and one label.unassigned record")
		}
	})

	t.Run("LAB-14 fixture: dana, frank with the override and carol cannot assign or read research's Group-owned labels", func(t *testing.T) {
		id := rOwn.ID.String()
		assign(t, w.bob, rOwn.ID, labFixtureTarget(fxR))
		defer func() {
			cx, tc := sel(t, w.bob, nil)
			if err := svc.UnassignLabel(cx, tc, id, labFixtureTarget(fxR)); err != nil {
				t.Errorf("cleanup unassign: %v", err)
			}
		}()
		denied := func(t *testing.T, who string, u *itUser) {
			t.Helper()
			cx, tc := sel(t, u, nil)
			wantLabel(t, who+" assigns", svc.AssignLabel(cx, tc, rTm.ID.String(), labFixtureTarget(fxR)), CodeNotFound, "")
			wantLabel(t, who+" unassigns", svc.UnassignLabel(cx, tc, id, labFixtureTarget(fxR)), CodeNotFound, "")
			_, err := svc.ListLabelAssignments(cx, tc, labFixtureTarget(fxR))
			wantLabel(t, who+" lists the target", err, CodeNotFound, "")
		}
		denied(t, "dana (design owner)", w.dana)
		denied(t, "carol (Organization owner, override off)", w.carol)
		off := e.withOverride(t, w.carol, w.A)
		denied(t, "frank (Organization admin) with the override on", w.frank)
		off()
		if e.labCount(t, labAsgByDef, rTm.ID.String()) != 0 || e.labCount(t, labAsgByKey, id, "fixture", fxR) != 1 {
			t.Error("a refused fixture write took effect")
		}
	})

	t.Run("LAB-15 fixture: a member reads target_members assignments only on a target readable to members", func(t *testing.T) {
		assign(t, w.bob, rOwn.ID, labFixtureTarget(fxR))
		assign(t, w.bob, rTm.ID, labFixtureTarget(fxR))
		assign(t, w.bob, rTm.ID, labFixtureTarget(fxHidden))
		defer func() {
			cx, tc := sel(t, w.bob, nil)
			for _, a := range []struct {
				id tenancy.LabelDefinitionID
				fx string
			}{{rOwn.ID, fxR}, {rTm.ID, fxR}, {rTm.ID, fxHidden}} {
				if err := svc.UnassignLabel(cx, tc, a.id.String(), labFixtureTarget(a.fx)); err != nil {
					t.Errorf("cleanup unassign: %v", err)
				}
			}
		}()
		dcx, dtc := sel(t, w.dave, nil)
		got, err := svc.ListLabelAssignments(dcx, dtc, labFixtureTarget(fxR))
		if err != nil || !slices.Equal(labIDs(got), []string{rTm.ID.String()}) {
			t.Errorf("dave lists a members-readable target: %v %v, want only the target_members label", labIDs(got), err)
		}
		_, err = svc.ListLabelAssignments(dcx, dtc, labFixtureTarget(fxHidden))
		wantLabel(t, "dave lists a managers-only target (readable is false)", err, CodeNotFound, "")
		wantLabel(t, "dave unassigns (visible label, readable target, no manage)", svc.UnassignLabel(dcx, dtc, rTm.ID.String(), labFixtureTarget(fxR)),
			CodeForbidden, "")
		for _, u := range []*itUser{w.bob, w.grace} {
			cx, tc := sel(t, u, nil)
			got, err := svc.ListLabelAssignments(cx, tc, labFixtureTarget(fxR))
			if err != nil || !slices.Equal(labIDs(got), labSorted(rOwn.ID.String(), rTm.ID.String())) {
				t.Errorf("%s (research manager) lists the target: %v %v", u.sub, labIDs(got), err)
			}
			got, err = svc.ListLabelAssignments(cx, tc, labFixtureTarget(fxHidden))
			if err != nil || !slices.Equal(labIDs(got), []string{rTm.ID.String()}) {
				t.Errorf("%s lists the managers-only target: %v %v", u.sub, labIDs(got), err)
			}
		}
	})

	t.Run("LAB-16 fixture: the adapter reads the target in the operation's transaction (a moved unassigned target)", func(t *testing.T) {
		fxMove := e.fixtureTarget(t, w.A, w.research.ID, true)
		if _, err := e.setup.Exec(context.Background(), `UPDATE app.label_fixture_targets SET group_id = $2::text::uuid WHERE id = $1::uuid`,
			fxMove, w.design.ID.String()); err != nil {
			t.Fatalf("move the unassigned target: %v", err)
		}
		cx, tc := sel(t, w.bob, nil)
		wantLabel(t, "bob assigns research's label to the moved target", svc.AssignLabel(cx, tc, rOwn.ID.String(), labFixtureTarget(fxMove)), CodeNotFound, "")
		if e.labCount(t, labAsgByKey, rOwn.ID.String(), "fixture", fxMove) != 0 {
			t.Error("the moved target received research's label")
		}
	})

	t.Run("LAB-17 deletion context: an Organization label needs the Organization-wide context; Group labels per Group", func(t *testing.T) {
		l := create(t, w.carol, labOrgScope, "Two Groups "+c[0], "owner_scope")
		id := l.ID.String()
		assign(t, w.carol, l.ID, labGroupTarget(w.research))
		assign(t, w.carol, l.ID, labGroupTarget(w.design))
		audit0 := e.labCount(t, labAuditAny, id)
		rcx, rtc := sel(t, w.carol, &w.research)
		_, err := svc.DeleteLabelDefinition(rcx, rtc, id)
		wantLabel(t, "carol deletes with research selected", err, CodeInvalidInput, fieldOrganizationContextRequired)
		if e.labCount(t, labDefByID, id) != 1 || e.labCount(t, labAsgByDef, id) != 2 || e.labCount(t, labAuditAny, id) != audit0 {
			t.Error("the selected-Group deletion removed something or recorded an audit row")
		}
		ccx, ctc := sel(t, w.carol, nil)
		d, err := svc.DeleteLabelDefinition(ccx, ctc, id)
		wantLabel(t, "carol deletes Organization-wide", err, "", "")
		if d.AssignmentsRemoved != 2 || e.labCount(t, labDefByID, id) != 0 || e.labCount(t, labAsgByDef, id) != 0 {
			t.Errorf("Organization-wide delete: %+v", d)
		}
		e.labDeleteOrder(t, l.ID, 2)

		g := create(t, w.bob, labGroupScope(w.research), "Group delete "+c[0], "owner_scope")
		gid := g.ID.String()
		assign(t, w.bob, g.ID, labFixtureTarget(fxR))
		// bob selects design only with a design role (granted for this case and removed after).
		e.grantGroup(t, w.A, w.design.ID, w.bob, tenancy.GroupMember)
		dcx, dtc := sel(t, w.bob, &w.design)
		_, err = svc.DeleteLabelDefinition(dcx, dtc, gid)
		wantLabel(t, "bob deletes research's label with design selected", err, CodeNotFound, "")
		removeGroupRole(t, w.bob, w.design.ID)
		if e.labCount(t, labDefByID, gid) != 1 || e.labCount(t, labAsgByDef, gid) != 1 {
			t.Error("the design-selected deletion removed something")
		}
		bcx, btc := sel(t, w.bob, &w.research)
		d, err = svc.DeleteLabelDefinition(bcx, btc, gid)
		wantLabel(t, "bob deletes with research selected", err, "", "")
		if d.AssignmentsRemoved != 1 || e.labCount(t, labDefByID, gid) != 0 || e.labCount(t, labAsgByDef, gid) != 0 {
			t.Errorf("research-selected delete: %+v", d)
		}
		e.labDeleteOrder(t, g.ID, 1)
	})

	ccx, ctc := sel(t, w.carol, nil)
	if _, err := svc.DeleteLabelDefinition(ccx, ctc, orgL.ID.String()); err != nil {
		t.Errorf("cleanup delete: %v", err)
	}
	bcx, btc := sel(t, w.bob, nil)
	for _, id := range []tenancy.LabelDefinitionID{rOwn.ID, rTm.ID} {
		if _, err := svc.DeleteLabelDefinition(bcx, btc, id.String()); err != nil {
			t.Errorf("cleanup delete: %v", err)
		}
	}
	teardown(t)
}

// labAuditOf counts the label.* audit rows of a label definition (the recorded denials, access.denied, are not counted).
func (e *itEnv) labAuditOf(t *testing.T, id tenancy.LabelDefinitionID) int {
	t.Helper()
	return e.labCount(t, `SELECT count(*) FROM audit.events WHERE resource_type = 'label' AND resource_id = $1::uuid AND action LIKE 'label.%'`, id.String())
}

// labSame compares two definition views field by field (the owning Group by value).
func labSame(a, b LabelDefinitionView) bool {
	return a.ID == b.ID && a.Scope == b.Scope && a.Name == b.Name && a.Visibility == b.Visibility &&
		(a.Group == nil) == (b.Group == nil) && (a.Group == nil || *a.Group == *b.Group)
}

// labDeleteOrder requires n label.unassigned records followed by one label.deleted record of a definition (the
// two-statement delete: assignments first, then the definition).
func (e *itEnv) labDeleteOrder(t *testing.T, id tenancy.LabelDefinitionID, n int) {
	t.Helper()
	rows, err := e.setup.Query(context.Background(), `SELECT action, occurred_at FROM audit.events WHERE resource_type = 'label' AND resource_id = $1::uuid
  AND action IN ('label.unassigned', 'label.deleted') ORDER BY occurred_at`, id.String())
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var un []time.Time
	var del []time.Time
	for rows.Next() {
		var a string
		var at time.Time
		if err := rows.Scan(&a, &at); err != nil {
			t.Fatal(err)
		}
		if a == "label.deleted" {
			del = append(del, at)
		} else {
			un = append(un, at)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(un) != n || len(del) != 1 {
		t.Errorf("audit: %d label.unassigned and %d label.deleted records, want %d and 1", len(un), len(del), n)
		return
	}
	for _, at := range un {
		if at.After(del[0]) {
			t.Error("audit: a label.unassigned record follows the label.deleted record")
		}
	}
}
