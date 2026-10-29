package access

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"core/internal/tenancy"
)

// T1/T2 (G020 Plan r2 §5.2) at unit level: the label capabilities' kinds and read capabilities, the not_found versus
// forbidden mapping on the same facts, the audit resource of a label denial and the label failure mapping
// (TestLabelPolicy); the closed target registry and its refusals before any SQL (TestLabelTargetsClosed); the label
// facade's input refusals before any transaction (TestLabelInputs) and its decision helpers on fixed facts
// (TestLabelDecisions; G020 Plan r2 §4.1). The literal matrix cells are matrix_test.go's.

func TestLabelPolicy(t *testing.T) {
	t.Run("kinds and read capabilities", func(t *testing.T) {
		for c, want := range map[Capability]struct {
			kind TargetKind
			read Capability
		}{
			capLabelOrganizationManage: {kindOrganization, capOrganizationRead},
			capLabelGroupManage:        {kindGroup, capLabelGroupRead},
			capLabelGroupRead:          {kindGroup, capLabelGroupRead},
		} {
			if k, ok := capabilityKind(c); !ok || k != want.kind {
				t.Errorf("capabilityKind(%s) = %s, %t; want %s", c, k, ok, want.kind)
			}
			if r := readCapability(c); r != want.read {
				t.Errorf("readCapability(%s) = %s, want %s", c, r, want.read)
			}
			if req, err := RequireCapability(string(c)); err != nil || req.cap != c {
				t.Errorf("RequireCapability(%s) = %+v, %v", c, req, err)
			}
		}
		// The existing capabilities keep their read capability of the kind.
		for c, want := range map[Capability]Capability{capGroupUpdate: capGroupRead, capOrganizationManageMembers: capOrganizationRead,
			capProfileUpdate: capProfileRead} {
			if r := readCapability(c); r != want {
				t.Errorf("readCapability(%s) = %s, want %s", c, r, want)
			}
		}
		if len(allCapabilities) != 22 {
			t.Errorf("%d capabilities, want 22", len(allCapabilities))
		}
	})
	t.Run("not_found versus forbidden on the same facts", func(t *testing.T) {
		mapped := func(what string, c Capability, sel contextSelection, f Facts, tg Target, want Code) {
			t.Helper()
			tg.Organization = orgA
			ok, r := decide(c, sel, f, tg)
			if ok {
				t.Fatalf("%s: %s allowed", what, c)
			}
			if got := denial(c, sel, f, tg, r); got.code != want {
				t.Errorf("%s: %s denied as %s, want %s", what, c, got.code, want)
			}
		}
		member := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupMember))
		// A member of g1 reads g1's labels: managing them is forbidden.
		mapped("group_member manages own Group labels", capLabelGroupManage, orgOnly, member, targetG(g1), CodeForbidden)
		// No role in g2: the labels are invisible (not_found), whatever group.read says.
		mapped("group_member manages other Group labels", capLabelGroupManage, orgOnly, member, targetG(g2), CodeNotFound)
		mapped("group_member reads other Group labels", capLabelGroupRead, orgOnly, member, targetG(g2), CodeNotFound)
		for _, override := range []bool{false, true} {
			frank := factsA(tenancy.OrgAdmin)
			frank.Policy.GroupManagementOverride = override
			// An Organization admin reads g1's metadata (group.read) but not its labels: not_found, override on or off.
			if ok, _ := decide(capGroupRead, orgOnly, frank, targetG(g1)); !ok {
				t.Fatal("fixture: org_admin must read Group metadata")
			}
			mapped("org_admin manages Group labels", capLabelGroupManage, orgOnly, frank, targetG(g1), CodeNotFound)
			mapped("org_admin reads Group labels", capLabelGroupRead, orgOnly, frank, targetG(g1), CodeNotFound)
		}
		// The Organization is readable by every member: the Organization labels are forbidden to a non-admin.
		mapped("org_member Organization labels", capLabelOrganizationManage, orgOnly, factsA(tenancy.OrgMember), targetOrgA, CodeForbidden)
		mapped("group_owner Organization labels", capLabelOrganizationManage, orgOnly, factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner)), targetOrgA, CodeForbidden)
		// Unknown, foreign and outside-context Groups are not_found.
		owner := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner))
		mapped("foreign Group", capLabelGroupManage, orgOnly, owner, targetG(gB), CodeNotFound)
		mapped("outside context", capLabelGroupRead, selG3, factsA(tenancy.OrgAdmin, groupBinding(g3, tenancy.GroupMember)), targetG(g1), CodeNotFound)
	})
	t.Run("audit resource of a label denial", func(t *testing.T) {
		label := tenancy.LabelDefinitionID{0x42}
		res, id, g := targetScope(Target{Kind: kindGroup, Organization: orgA, Group: g1, Label: label})
		if res != "label" || id != [16]byte(label) || g != g1 {
			t.Errorf("Group-owned label scope = %s %v %v", res, id, g)
		}
		res, id, g = targetScope(Target{Kind: kindOrganization, Organization: orgA, Label: label})
		if res != "label" || id != [16]byte(label) || !zero(g) {
			t.Errorf("Organization-owned label scope = %s %v %v", res, id, g)
		}
		if res, _, _ := targetScope(Target{Kind: kindProfile, Profile: p1, Label: label}); res != "" {
			t.Errorf("label on a profile target = %s", res)
		}
		// Without a label the existing resources are unchanged.
		if res, _, g := targetScope(Target{Kind: kindGroup, Group: g1}); res != "group" || g != g1 {
			t.Errorf("Group scope = %s %v", res, g)
		}
	})
	t.Run("label failure mapping", func(t *testing.T) {
		tx := &trackedTx{}
		for code, want := range map[tenancy.Code]struct {
			code  Code
			field string
		}{
			tenancy.CodeNameConflict:         {CodeConflict, "name_conflict"},
			tenancy.CodeScopeFull:            {CodeConflict, "scope_full"},
			tenancy.CodeAlreadyAssigned:      {CodeConflict, "already_assigned"},
			tenancy.CodeNotAssigned:          {CodeConflict, "not_assigned"},
			tenancy.CodeLabelNotFound:        {CodeNotFound, ""},
			tenancy.CodeTargetUnsupported:    {CodeInvalidInput, "target_unsupported"},
			tenancy.CodeVisibilityNotAllowed: {CodeInvalidInput, "visibility_not_allowed"},
			tenancy.CodeInvalidVisibility:    {CodeInvalidInput, "invalid_visibility"},
			tenancy.CodeInvalidScope:         {CodeInvalidInput, "invalid_scope"},
			tenancy.CodeInvalidName:          {CodeInvalidInput, "invalid_name"},
			tenancy.CodeGroupNotFound:        {CodeNotFound, ""},
			tenancy.CodeUnavailable:          {CodeUnavailable, ""},
			tenancy.CodeInternal:             {CodeInternal, ""},
		} {
			f := labelFailure(context.Background(), tx, &tenancy.Error{Code: code})
			if f.code != want.code || f.field != want.field {
				t.Errorf("%s -> %s/%q, want %s/%q", code, f.code, f.field, want.code, want.field)
			}
		}
		if f := labelFailure(context.Background(), tx, errors.New("x")); f.code != CodeInternal {
			t.Errorf("a foreign error -> %s", f.code)
		}
		for _, field := range []string{fieldNameConflict, fieldScopeFull, fieldAlreadyAssigned, fieldNotAssigned, fieldTargetUnsupported,
			fieldVisibilityNotAllowed, fieldTargetAmbiguous, fieldOrganizationContextRequired} {
			if f := labelInputFailure(field); f.code != CodeInvalidInput || f.field != field || f.reason != reasonInputInvalid {
				t.Errorf("labelInputFailure(%s) = %+v", field, f)
			}
		}
	})
}

// probeTargetAdapter is a test-only kind registered through labelHooks (a Group-owned target owned by g1).
type probeTargetAdapter struct{}

func (probeTargetAdapter) admits(owner tenancy.ScopeType) bool { return owner == tenancy.ScopeGroup }
func (probeTargetAdapter) resolve(context.Context, *trackedTx, Facts, tenancy.OrganizationID, [16]byte) (tenancy.GroupID, bool, error) {
	return g1, true, nil
}
func (probeTargetAdapter) readable(contextSelection, Facts, tenancy.GroupID) bool { return false }

// failed asserts an invalid_input refusal with field.
func failed(t *testing.T, what string, f *failure, field string) {
	t.Helper()
	if f == nil || f.code != CodeInvalidInput || f.field != field || f.reason != reasonInputInvalid {
		t.Errorf("%s: %+v, want invalid_input/%s", what, f, field)
	}
}

func TestLabelInputs(t *testing.T) {
	gid := g1
	t.Run("scope", func(t *testing.T) {
		more := func() *failure { return nil }
		for name, sc := range map[string]Scope{
			"empty kind":                 {},
			"unknown kind":               {Kind: "user"},
			"organization with a group":  {Kind: tenancy.ScopeOrganization, Group: &gid},
			"group without a group":      {Kind: tenancy.ScopeGroup},
			"group with a zero group":    {Kind: tenancy.ScopeGroup, Group: &tenancy.GroupID{}},
			"case variant of the kind":   {Kind: "Organization"},
			"case variant of group kind": {Kind: "Group", Group: &gid},
		} {
			failed(t, name, labelScopePre(sc, more)(), "invalid_scope")
		}
		ran := 0
		for _, sc := range []Scope{{Kind: tenancy.ScopeOrganization}, {Kind: tenancy.ScopeGroup, Group: &gid}} {
			if f := labelScopePre(sc, func() *failure { ran++; return nil })(); f != nil {
				t.Errorf("%+v: %+v", sc, f)
			}
		}
		if ran != 2 {
			t.Errorf("the further checks ran %d times, want 2", ran)
		}
	})
	t.Run("definition id", func(t *testing.T) {
		for _, s := range []string{"", "x", "00000000-0000-0000-0000-000000000000", "{" + g1.String() + "}",
			g1.String() + " "} {
			_, f := labelID(s)
			failed(t, s, f, "invalid_id")
		}
		if id, f := labelID(g1.String()); f != nil || id != tenancy.LabelDefinitionID(g1) {
			t.Errorf("canonical id: %v, %+v", id, f)
		}
	})
	t.Run("page", func(t *testing.T) {
		for _, l := range []int{-1, 101, 1000} {
			_, _, f := labelPageInput("", l)
			failed(t, "limit", f, "invalid_page_size")
		}
		for _, c := range []string{"x", "00000000-0000-0000-0000-000000000000"} {
			_, _, f := labelPageInput(c, 10)
			failed(t, "cursor "+c, f, "invalid_cursor")
		}
		// The size is checked before the cursor.
		_, _, f := labelPageInput("x", 101)
		failed(t, "both", f, "invalid_page_size")
		for in, want := range map[int]int{0: tenancy.DefaultLabelPage, 1: 1, 100: tenancy.MaxLabelPage} {
			if from, got, f := labelPageInput("", in); f != nil || got != want || !zero(from) {
				t.Errorf("limit %d -> %d, %v, %+v; want %d", in, got, from, f, want)
			}
		}
		if from, _, f := labelPageInput(g1.String(), 5); f != nil || from != tenancy.LabelDefinitionID(g1) {
			t.Errorf("cursor: %v, %+v", from, f)
		}
	})
	t.Run("definition name and visibility", func(t *testing.T) {
		for name, c := range map[string]struct {
			scope            tenancy.ScopeType
			label, vis, want string
		}{
			"unknown visibility":                 {tenancy.ScopeGroup, "Urgent", "public", "invalid_visibility"},
			"empty visibility":                   {tenancy.ScopeGroup, "Urgent", "", "invalid_visibility"},
			"case variant visibility":            {tenancy.ScopeGroup, "Urgent", "Owner_scope", "invalid_visibility"},
			"target_members in the Organization": {tenancy.ScopeOrganization, "Urgent", "target_members", "visibility_not_allowed"},
			"empty name":                         {tenancy.ScopeGroup, "", "owner_scope", "invalid_name"},
			"blank name":                         {tenancy.ScopeGroup, "   ", "owner_scope", "invalid_name"},
			"long name":                          {tenancy.ScopeGroup, strings.Repeat("a", tenancy.MaxLabelNameLength+1), "owner_scope", "invalid_name"},
			"control character":                  {tenancy.ScopeOrganization, "a\x07b", "owner_scope", "invalid_name"},
		} {
			_, _, f := labelDefinitionInput(c.scope, c.label, c.vis)
			failed(t, name, f, c.want)
		}
		// visibility_not_allowed is decided before the name: nothing about the name is reported for it.
		_, _, f := labelDefinitionInput(tenancy.ScopeOrganization, "", "target_members")
		failed(t, "Organization target_members with a bad name", f, "visibility_not_allowed")
		for _, c := range []struct {
			scope tenancy.ScopeType
			vis   tenancy.LabelVisibility
		}{{tenancy.ScopeOrganization, tenancy.LabelVisibilityOwnerScope}, {tenancy.ScopeGroup, tenancy.LabelVisibilityOwnerScope},
			{tenancy.ScopeGroup, tenancy.LabelVisibilityTargetMembers}} {
			name, vis, f := labelDefinitionInput(c.scope, "  Urgent  ", string(c.vis))
			if f != nil || name != "Urgent" || vis != c.vis {
				t.Errorf("%s/%s: %q, %s, %+v", c.scope, c.vis, name, vis, f)
			}
		}
		long := strings.Repeat("é", tenancy.MaxLabelNameLength)
		if name, _, f := labelDefinitionInput(tenancy.ScopeGroup, long, "owner_scope"); f != nil || name != long {
			t.Errorf("80 code points refused: %+v", f)
		}
	})
	t.Run("target selector", func(t *testing.T) {
		g := g1.String()
		for name, c := range map[string]struct {
			sel  LabelTargetSelector
			want string
		}{
			"none":              {LabelTargetSelector{}, "target_unsupported"},
			"group and kind":    {LabelTargetSelector{Group: g, Kind: "group"}, "target_ambiguous"},
			"group and id":      {LabelTargetSelector{Group: g, ID: g}, "target_ambiguous"},
			"unregistered kind": {LabelTargetSelector{Kind: "business_profile", ID: g}, "target_unsupported"},
			"malformed group":   {LabelTargetSelector{Group: "x"}, "invalid_id"},
			"zero group":        {LabelTargetSelector{Group: "00000000-0000-0000-0000-000000000000"}, "invalid_id"},
		} {
			_, _, f := labelTargetInput(c.sel)
			failed(t, name, f, c.want)
		}
		for _, sel := range []LabelTargetSelector{{Group: g}, {Kind: "group", ID: g}} {
			tg, a, f := labelTargetInput(sel)
			if f != nil || tg.Kind != labelTargetGroup || tg.ID != [16]byte(g1) || a != (groupTargetAdapter{}) {
				t.Errorf("%+v -> %+v, %T, %+v", sel, tg, a, f)
			}
		}
		// A test-registered kind is admitted only as a Group-owned target; one admitting no Group-owned definition is
		// target_unsupported before any SQL.
		labelHooks = &labelTargetHooks{targets: map[labelTargetKind]labelTargetAdapter{"probe": probeTargetAdapter{},
			"orgprobe": groupTargetAdapter{}}}
		defer func() { labelHooks = nil }()
		if tg, a, f := labelTargetInput(LabelTargetSelector{Kind: "probe", ID: g}); f != nil || tg.Kind != "probe" || a != (probeTargetAdapter{}) {
			t.Errorf("probe: %+v, %T, %+v", tg, a, f)
		}
		_, _, f := labelTargetInput(LabelTargetSelector{Kind: "orgprobe", ID: g})
		failed(t, "kind without Group-owned admission", f, "target_unsupported")
	})
	t.Run("target fact query", func(t *testing.T) {
		tc := &TenantContext{organization: orgA}
		g := g1.String()
		// Both selector forms of a group target read (and, for a write, lock) the target Group.
		for _, sel := range []LabelTargetSelector{{Group: g}, {Kind: "group", ID: g}} {
			if q := labelTargetQuery(sel, true)(tc); !slices.Equal(q.Groups, []tenancy.GroupID{g1}) || q.TargetGroup != g1 {
				t.Errorf("%+v write: %+v", sel, q)
			}
			if q := labelTargetQuery(sel, false)(tc); !slices.Equal(q.Groups, []tenancy.GroupID{g1}) || !zero(q.TargetGroup) {
				t.Errorf("%+v read: %+v", sel, q)
			}
		}
		// Another selected Group: nothing is locked exclusively (the context gate denies).
		other := &TenantContext{organization: orgA, group: g3}
		if q := labelTargetQuery(LabelTargetSelector{Group: g}, true)(other); !slices.Equal(q.Groups, []tenancy.GroupID{g1, g3}) || !zero(q.TargetGroup) {
			t.Errorf("other selected Group: %+v", q)
		}
		// Another kind, an ambiguous or a malformed selector: the Organization facts only.
		for _, sel := range []LabelTargetSelector{{Kind: "probe", ID: g}, {Group: g, Kind: "group"}, {Group: g, ID: g}, {Group: "x"},
			{Group: "00000000-0000-0000-0000-000000000000"}, {Kind: "group", ID: "x"}, {}} {
			if q := labelTargetQuery(sel, true)(tc); len(q.Groups) != 0 || !zero(q.TargetGroup) || q.ExclusiveOrganization {
				t.Errorf("%+v: %+v", sel, q)
			}
		}
	})
	t.Run("views and pages", func(t *testing.T) {
		org := tenancy.LabelDefinition{ID: uid[tenancy.LabelDefinitionID](0x21), OrganizationID: orgA, OwnerScopeType: tenancy.ScopeOrganization,
			OwnerScopeID: orgA, Name: "Org", Visibility: tenancy.LabelVisibilityOwnerScope}
		grp := tenancy.LabelDefinition{ID: uid[tenancy.LabelDefinitionID](0x22), OrganizationID: orgA, OwnerScopeType: tenancy.ScopeGroup,
			OwnerScopeID: g1, GroupID: g1, Name: "Grp", Visibility: tenancy.LabelVisibilityTargetMembers}
		if v := labelDefinitionView(org); v.Group != nil || v.ID != org.ID || v.Scope != tenancy.ScopeOrganization || v.Name != "Org" ||
			v.Visibility != tenancy.LabelVisibilityOwnerScope {
			t.Errorf("Organization view %+v", v)
		}
		if v := labelDefinitionView(grp); v.Group == nil || *v.Group != g1 || v.Scope != tenancy.ScopeGroup {
			t.Errorf("Group view %+v", v)
		}
		if p := labelDefinitionPage(nil, 2); p.Labels == nil || len(p.Labels) != 0 || p.Next != nil {
			t.Errorf("empty page %+v", p)
		}
		if p := labelDefinitionPage([]tenancy.LabelDefinition{org, grp}, 2); len(p.Labels) != 2 || p.Next != nil {
			t.Errorf("full last page %+v", p)
		}
		third := grp
		third.ID = uid[tenancy.LabelDefinitionID](0x23)
		if p := labelDefinitionPage([]tenancy.LabelDefinition{org, grp, third}, 2); len(p.Labels) != 2 || p.Next == nil || *p.Next != grp.ID {
			t.Errorf("limit+1 read %+v", p)
		}
	})
	t.Run("assignment owner facts", func(t *testing.T) {
		orgRow := tenancy.LabelAssignment{OrganizationID: orgA, OwnerScopeType: tenancy.ScopeOrganization, OwnerScopeID: orgA, TargetGroupID: g1}
		grpRow := tenancy.LabelAssignment{OrganizationID: orgA, OwnerScopeType: tenancy.ScopeGroup, OwnerScopeID: g1, TargetGroupID: g1}
		if !labelAssignmentOwned(orgRow, tenancy.ScopeOrganization, g1) || labelAssignmentOwned(grpRow, tenancy.ScopeOrganization, g1) {
			t.Error("a group target admits exactly Organization-owned rows")
		}
		if !labelAssignmentOwned(grpRow, tenancy.ScopeGroup, g1) || labelAssignmentOwned(grpRow, tenancy.ScopeGroup, g2) ||
			labelAssignmentOwned(orgRow, tenancy.ScopeGroup, g1) || labelAssignmentOwned(grpRow, "", g1) {
			t.Error("a Group-owned kind admits exactly the owning Group's rows")
		}
	})
}

func TestLabelDecisions(t *testing.T) {
	label := uid[tenancy.LabelDefinitionID](0x21)
	ctx := context.Background()
	groupTarget := labelTarget{Kind: labelTargetGroup, ID: [16]byte(g1)}
	t.Run("labelDecide records the evaluated capability", func(t *testing.T) {
		var ev evaluation
		tg := Target{Kind: kindGroup, Organization: orgA, Group: g1, Label: label}
		if f := labelDecide(orgOnly, factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupAdmin)), capLabelGroupManage, tg, &ev); f != nil ||
			ev.capability != capLabelGroupManage || ev.reason != reasonAllowed {
			t.Errorf("Group admin: %+v, %+v", f, ev)
		}
		ev = evaluation{}
		// A Group member: forbidden (readable through label.group.read); the Organization admin: not_found.
		if f := labelDecide(orgOnly, factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupMember)), capLabelGroupManage, tg, &ev); f == nil ||
			f.code != CodeForbidden || ev.capability != capLabelGroupManage || f.denied == nil {
			t.Errorf("Group member: %+v, %+v", f, ev)
		}
		if f := labelDecide(orgOnly, factsA(tenancy.OrgAdmin), capLabelGroupManage, tg, &ev); f == nil || f.code != CodeNotFound {
			t.Errorf("Organization admin: %+v", f)
		}
	})
	t.Run("group target: label.organization.manage then group.read", func(t *testing.T) {
		var s *Service
		a := groupTargetAdapter{}
		var ev evaluation
		g, f := s.labelGroupTarget(ctx, nil, orgOnly, factsA(tenancy.OrgAdmin), groupTarget, a, label, &ev)
		if f != nil || g != g1 || ev.capability != capLabelOrganizationManage {
			t.Errorf("Organization admin: %v, %+v, %+v", g, f, ev)
		}
		ev = evaluation{}
		owner := factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupOwner))
		if _, f := s.labelGroupTarget(ctx, nil, orgOnly, owner, groupTarget, a, label, &ev); f == nil || f.code != CodeForbidden ||
			ev.capability != capLabelOrganizationManage {
			t.Errorf("Group owner without an Organization admin role: %+v, %+v", f, ev)
		}
		ev = evaluation{}
		absent := labelTarget{Kind: labelTargetGroup, ID: [16]byte(gB)}
		if _, f := s.labelGroupTarget(ctx, nil, orgOnly, factsA(tenancy.OrgAdmin), absent, a, label, &ev); f == nil || f.code != CodeNotFound {
			t.Errorf("foreign Group: %+v", f)
		}
		ev = evaluation{}
		if _, f := s.labelGroupTarget(ctx, nil, selG3, factsA(tenancy.OrgAdmin), groupTarget, a, label, &ev); f == nil || f.code != CodeNotFound ||
			ev.capability != capGroupRead {
			t.Errorf("target outside the selected Group: %+v, %+v", f, ev)
		}
	})
	t.Run("labelResolve", func(t *testing.T) {
		if g, f := labelResolve(ctx, nil, factsA(tenancy.OrgAdmin), groupTarget, groupTargetAdapter{}); f != nil || g != g1 {
			t.Errorf("g1: %v, %+v", g, f)
		}
		gone := labelTarget{Kind: labelTargetGroup, ID: [16]byte(gB)}
		if _, f := labelResolve(ctx, nil, factsA(tenancy.OrgAdmin), gone, groupTargetAdapter{}); f == nil || f.code != CodeNotFound {
			t.Errorf("gB: %+v", f)
		}
		if g, f := labelResolve(ctx, nil, factsA(tenancy.OrgAdmin), labelTarget{Kind: "probe"}, probeTargetAdapter{}); f != nil || g != g1 {
			t.Errorf("probe: %v, %+v", g, f)
		}
	})
}

func TestLabelTargetsClosed(t *testing.T) {
	t.Run("the production registry is exactly group", func(t *testing.T) {
		var kinds []string
		for k := range labelTargets {
			kinds = append(kinds, string(k))
		}
		if !slices.Equal(kinds, []string{"group"}) || labelHooks != nil {
			t.Fatalf("production label targets %v, hooks %v; want [group] and nil", kinds, labelHooks)
		}
		if _, ok := labelTargets[labelTargetGroup].(groupTargetAdapter); !ok {
			t.Fatal("group is not the groupTargetAdapter")
		}
	})
	gid := g1.String()
	t.Run("selectors are parsed before any SQL", func(t *testing.T) {
		// No transaction exists in this test: every refusal below happens without storage.
		for name, c := range map[string]struct {
			sel   labelSelector
			code  Code
			field string
		}{
			"no target":          {labelSelector{}, CodeInvalidInput, "target_unsupported"},
			"unknown kind":       {labelSelector{kind: "business_profile", id: gid}, CodeInvalidInput, "target_unsupported"},
			"unregistered kind":  {labelSelector{kind: "probe", id: gid}, CodeInvalidInput, "target_unsupported"},
			"case variant":       {labelSelector{kind: "Group", id: gid}, CodeInvalidInput, "target_unsupported"},
			"padded kind":        {labelSelector{kind: "group ", id: gid}, CodeInvalidInput, "target_unsupported"},
			"id without kind":    {labelSelector{id: gid}, CodeInvalidInput, "target_unsupported"},
			"group and kind":     {labelSelector{group: gid, kind: "group"}, CodeInvalidInput, "target_ambiguous"},
			"group and id":       {labelSelector{group: gid, id: gid}, CodeInvalidInput, "target_ambiguous"},
			"group, kind and id": {labelSelector{group: gid, kind: "probe", id: gid}, CodeInvalidInput, "target_ambiguous"},
			"malformed group":    {labelSelector{group: "not-a-uuid"}, CodeInvalidInput, "invalid_id"},
			"zero group":         {labelSelector{group: "00000000-0000-0000-0000-000000000000"}, CodeInvalidInput, "invalid_id"},
			"malformed kind id":  {labelSelector{kind: "group", id: "x"}, CodeInvalidInput, "invalid_id"},
		} {
			if _, f := parseLabelTarget(c.sel); f == nil || f.code != c.code || f.field != c.field || f.reason != reasonInputInvalid {
				t.Errorf("%s: %+v, want %s/%s", name, f, c.code, c.field)
			}
		}
		for _, sel := range []labelSelector{{group: gid}, {kind: "group", id: gid}} {
			got, f := parseLabelTarget(sel)
			if f != nil || got.Kind != labelTargetGroup || got.ID != [16]byte(g1) || !zero(got.Group) {
				t.Errorf("%+v -> %+v, %+v", sel, got, f)
			}
		}
	})
	t.Run("owner-scope admission before any SQL", func(t *testing.T) {
		tg := labelTarget{Kind: labelTargetGroup, ID: [16]byte(g1)}
		if a, f := admitLabelTarget(tenancy.ScopeOrganization, tg); f != nil || a == nil {
			t.Errorf("Organization-owned definition on a Group: %+v", f)
		}
		// A Group-owned definition targeting a Group is refused: production storage holds no Group-owned assignment.
		if _, f := admitLabelTarget(tenancy.ScopeGroup, tg); f == nil || f.code != CodeInvalidInput || f.field != "target_unsupported" {
			t.Errorf("Group-owned definition on a Group: %+v", f)
		}
		for _, owner := range []tenancy.ScopeType{"", "Organization", "user"} {
			if _, f := admitLabelTarget(owner, tg); f == nil || f.field != "target_unsupported" {
				t.Errorf("owner scope %q: %+v", owner, f)
			}
		}
		if _, f := admitLabelTarget(tenancy.ScopeGroup, labelTarget{Kind: "probe", ID: [16]byte(g1)}); f == nil || f.field != "target_unsupported" {
			t.Errorf("unregistered kind admitted: %+v", f)
		}
	})
	t.Run("the group adapter resolves from the operation's facts", func(t *testing.T) {
		a := groupTargetAdapter{}
		f := factsA(tenancy.OrgAdmin)
		if g, ok, err := a.resolve(context.Background(), nil, f, orgA, [16]byte(g1)); err != nil || !ok || g != g1 {
			t.Errorf("resolve(g1) = %v, %t, %v", g, ok, err)
		}
		for _, id := range [][16]byte{[16]byte(gB), [16]byte(uid[tenancy.GroupID](0x7e)), {}} {
			if _, ok, err := a.resolve(context.Background(), nil, f, orgA, id); err != nil || ok {
				t.Errorf("resolve(%v) = %t, %v; want absent", id, ok, err)
			}
		}
		if _, ok, _ := a.resolve(context.Background(), nil, f, orgB, [16]byte(g1)); ok {
			t.Error("resolve found an A Group in Organization B")
		}
		if !a.readable(orgOnly, f, g1) {
			t.Error("an Organization admin reads the target Group (group.read)")
		}
		if a.readable(orgOnly, factsA(tenancy.OrgMember), g1) || a.readable(selG3, factsA(tenancy.OrgAdmin), g1) {
			t.Error("readable without group.read on the target Group")
		}
		if !a.readable(orgOnly, factsA(tenancy.OrgMember, groupBinding(g1, tenancy.GroupMember)), g1) {
			t.Error("a member of the Group reads it")
		}
	})
	t.Run("a test-registered kind never replaces group", func(t *testing.T) {
		labelHooks = &labelTargetHooks{targets: map[labelTargetKind]labelTargetAdapter{"probe": probeTargetAdapter{}, labelTargetGroup: probeTargetAdapter{}}}
		defer func() { labelHooks = nil }()
		if got, f := parseLabelTarget(labelSelector{kind: "probe", id: gid}); f != nil || got.Kind != "probe" {
			t.Errorf("registered probe: %+v, %+v", got, f)
		}
		if a, _ := labelTargetAdapterFor(labelTargetGroup); a != (groupTargetAdapter{}) {
			t.Errorf("group adapter replaced by %T", a)
		}
		if _, f := admitLabelTarget(tenancy.ScopeGroup, labelTarget{Kind: "probe"}); f != nil {
			t.Errorf("probe on a Group-owned definition: %+v", f)
		}
		if _, f := admitLabelTarget(tenancy.ScopeOrganization, labelTarget{Kind: "probe"}); f == nil {
			t.Error("probe on an Organization-owned definition admitted")
		}
	})
}
