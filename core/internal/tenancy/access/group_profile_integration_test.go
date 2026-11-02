//go:build integration

// G075 B-GROUP-PROFILE-IT (Plan r2 §5, §8 L1; EP-1): TestGroupMemberProfile against the access integration environment
// (setupIT: the private PostgreSQL of the migrations suite, no MinIO). EP-1 slice: GMP-2 a Group owner's label edit of a
// member (the DTO field pins, the identical-repeat no-op and both readers' cards), GMP-3 a Group admin's edit (an
// Organization admin who is also the Group's admin) and a stale version, GMP-NEG the refusals (a peer, an Organization
// owner/admin without the Group role under both override settings, a non-member editor, a subject without a membership,
// a target outside the selected Group, invalid input before any transaction) and the raw runtime forgeries of the
// Organization, Group and user against the guard's binding check, GMP-7 the cascade through RemoveMember, LeaveScope,
// the Organization departure and retirement with readmission starting empty and no label event in the audit stream,
// GMP-8 the trigger-owned audit rows of admin edits (actor not the subject) and a self edit (actor = subject), never the
// text, GMP-11 the facade reads (one not_found identity for an unrelated or foreign subject, the Organization oversight
// cards under both override settings, the selected-Group narrowing) and GMP-13 the authorized-overflow outcome.
// Failure messages print codes, counts and hashes only.
package access

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/tenancy"
)

// gmpWorld is TestGroupMemberProfile's fixture (Plan r2 §5 "Fixed-actor worlds": the identity seed's roles on fresh
// users). Organization A: carol (org_owner), frank (org_admin), grace (org_admin and research group_admin), alice, bob,
// dave and hostile (org_member); Group research (bob group_owner, grace group_admin, dave group_member) and design (no
// member but its creator). Organization B: hostile (org_owner). own creates the Groups (each creator's owner binding) and
// holds no Organization binding; it is never a viewer or a subject.
type gmpWorld struct {
	A, B                       tenancy.OrganizationID
	research, design           tenancy.Group
	carol, frank, grace, alice *itUser
	bob, dave, hostile, own    *itUser
}

func (e *itEnv) gmpWorld(t *testing.T) *gmpWorld {
	t.Helper()
	c := e.canaries
	w := &gmpWorld{}
	w.A, w.B = e.org(t, "Group Profiles A "+c[0]), e.org(t, "Group Profiles B "+c[1])
	for _, u := range []**itUser{&w.carol, &w.frank, &w.grace, &w.alice, &w.bob, &w.dave, &w.hostile, &w.own} {
		*u = e.newUser(t)
	}
	w.research = e.group(t, w.A, w.own, "GMP Research "+c[1], "gmp-research")
	w.design = e.group(t, w.A, w.own, "GMP Design "+c[2], "gmp-design")
	e.grantOrg(t, w.A, w.carol, tenancy.OrgOwner)
	e.grantOrg(t, w.A, w.frank, tenancy.OrgAdmin)
	e.grantOrg(t, w.A, w.grace, tenancy.OrgAdmin)
	for _, u := range []*itUser{w.alice, w.bob, w.dave, w.hostile} {
		e.grantOrg(t, w.A, u, tenancy.OrgMember)
	}
	e.grantGroup(t, w.A, w.research.ID, w.bob, tenancy.GroupOwner)
	e.grantGroup(t, w.A, w.research.ID, w.grace, tenancy.GroupAdmin)
	e.grantGroup(t, w.A, w.research.ID, w.dave, tenancy.GroupMember)
	e.grantOrg(t, w.B, w.hostile, tenancy.OrgOwner)
	for _, u := range []*itUser{w.carol, w.frank, w.grace, w.alice, w.bob, w.dave, w.hostile} {
		e.bind(t, u, w.A)
	}
	e.bind(t, w.hostile, w.B)
	return w
}

// gmpBinding is u's current membership binding id in Group g ("" without one; attested superuser read).
func (e *itEnv) gmpBinding(t *testing.T, u *itUser, g tenancy.GroupID) string {
	t.Helper()
	var s string
	err := e.setup.QueryRow(context.Background(), `SELECT id::text FROM app.role_bindings WHERE subject_user_id = $1::text::uuid
  AND scope_type = 'group' AND scope_id = $2::text::uuid AND source = 'membership'`, u.id.String(), g.String()).Scan(&s)
	if errors.Is(err, pgx.ErrNoRows) {
		return ""
	}
	if err != nil {
		t.Fatalf("binding: %v", err)
	}
	return s
}

// gmpRow is the label row of a binding as "label|version" ("" without a row, "-" for a cleared label; superuser read).
func (e *itEnv) gmpRow(t *testing.T, binding string) string {
	t.Helper()
	var s string
	err := e.setup.QueryRow(context.Background(), `SELECT coalesce(role_label, '-') || '|' || version FROM app.group_member_profiles
WHERE binding_id = $1::text::uuid`, binding).Scan(&s)
	if errors.Is(err, pgx.ErrNoRows) {
		return ""
	}
	if err != nil {
		t.Fatalf("label row: %v", err)
	}
	return s
}

// gmpRows counts u's label rows in every Group (superuser read).
func (e *itEnv) gmpRows(t *testing.T, u *itUser) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*)::int FROM app.group_member_profiles WHERE user_id = $1::text::uuid`,
		u.id.String()).Scan(&n); err != nil {
		t.Fatalf("label rows: %v", err)
	}
	return n
}

// gmpEvents counts the group_member_profile.updated rows naming subject u, and the membership.removed rows of u's
// membership of Group g (superuser read).
func (e *itEnv) gmpEvents(t *testing.T, u *itUser, g tenancy.GroupID) (labels, removed int) {
	t.Helper()
	if err := e.setup.QueryRow(context.Background(), `SELECT
  count(*) FILTER (WHERE action = 'group_member_profile.updated')::int,
  count(*) FILTER (WHERE action = 'membership.removed' AND group_id = $2::text::uuid AND resource_type = 'role_binding')::int
FROM audit.events WHERE subject_user_id = $1::text::uuid`, u.id.String(), g.String()).Scan(&labels, &removed); err != nil {
		t.Fatalf("audit rows: %v", err)
	}
	return labels, removed
}

// gmpWant asserts err's public code and (when field is not "") its field.
func gmpWant(t *testing.T, what string, err error, code Code, field string) {
	t.Helper()
	var ae *Error
	if CodeOf(err) != code || (field != "" && (!errors.As(err, &ae) || ae.Field != field)) {
		t.Errorf("%s: %v, want %s %s", what, err, code, field)
	}
}

// gmpCard renders a card without its Group name as "label:edit:version" ("-" for no label).
func gmpCard(c GroupMemberCard) string {
	label := "-"
	if c.RoleLabel != nil {
		label = *c.RoleLabel
	}
	return fmt.Sprintf("%s:%t:%d", label, c.CanEdit, c.Version)
}

func TestGroupMemberProfile(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	w := e.gmpWorld(t)
	gid := w.research.ID
	research := Scope{Kind: tenancy.ScopeGroup, Group: &gid}

	t.Run("GMP-2 a Group owner sets a member's label; the DTO fields, the no-op repeat and both readers' cards", func(t *testing.T) {
		// The DTOs carry exactly the reviewed fields (Plan r2 §3.4 visibility table): never a permission role, another
		// member, a count, an email or a Business Profile.
		for _, c := range []struct {
			v    any
			want string
		}{
			{GroupMemberCard{}, "GroupID,GroupName,RoleLabel,CanEdit,Version"},
			{GroupMemberCards{}, "Subject,Cards,NarrowRequired"},
		} {
			if got := proFields(c.v); got != c.want {
				t.Errorf("%T fields %s, want %s", c.v, got, c.want)
			}
		}
		bctx, btc := e.as(t, w.bob, w.A)
		card, err := e.svc.SetGroupMemberLabel(bctx, btc, gid, w.dave.id, 0, proStr("  Lab lead  "))
		if err != nil || card.GroupID != gid || card.GroupName != w.research.Name || gmpCard(card) != "Lab lead:true:1" {
			t.Fatalf("owner sets the label: %s, name matches %t, %v", gmpCard(card), card.GroupName == w.research.Name, err)
		}
		binding := e.gmpBinding(t, w.dave, gid)
		if got := e.gmpRow(t, binding); got != "Lab lead|1" {
			t.Fatalf("label row %s, want the trimmed label at version 1", itH(got))
		}
		labels, _ := e.gmpEvents(t, w.dave, gid)
		// An identical repeat, even on a stale version, writes nothing and records nothing.
		again, err := e.svc.SetGroupMemberLabel(bctx, btc, gid, w.dave.id, 0, proStr("Lab lead"))
		if err != nil || gmpCard(again) != "Lab lead:true:1" {
			t.Fatalf("identical repeat: %s, %v", gmpCard(again), err)
		}
		if l, _ := e.gmpEvents(t, w.dave, gid); l != labels || labels != 1 {
			t.Errorf("label events %d after the repeat (%d before), want 1", l, labels)
		}
		dctx, dtc := e.as(t, w.dave, w.A)
		for _, c := range []struct {
			what string
			ctx  context.Context
			tc   *TenantContext
		}{{"the member itself", dctx, dtc}, {"the Group owner", bctx, btc}} {
			cs, err := e.svc.ListGroupMemberCards(c.ctx, c.tc, w.dave.id)
			if err != nil || cs.Subject != w.dave.id || cs.NarrowRequired || len(cs.Cards) != 1 || cs.Cards[0].GroupID != gid ||
				gmpCard(cs.Cards[0]) != "Lab lead:true:1" {
				t.Errorf("%s reads %d cards (narrow %t), %v", c.what, len(cs.Cards), cs.NarrowRequired, err)
			}
		}
	})

	t.Run("GMP-3 a Group admin replaces the label; a stale version conflicts with zero effects", func(t *testing.T) {
		gctx, gtc := e.as(t, w.grace, w.A)
		card, err := e.svc.SetGroupMemberLabel(gctx, gtc, gid, w.dave.id, 1, proStr("Principal investigator"))
		if err != nil || gmpCard(card) != "Principal investigator:true:2" {
			t.Fatalf("Group admin replaces the label: %s, %v", gmpCard(card), err)
		}
		_, err = e.svc.SetGroupMemberLabel(gctx, gtc, gid, w.dave.id, 1, proStr("Something else"))
		gmpWant(t, "stale version", err, CodeConflict, fieldVersionConflict)
		if got := e.gmpRow(t, e.gmpBinding(t, w.dave, gid)); got != "Principal investigator|2" {
			t.Errorf("label row %s after the conflict, want version 2 unchanged", itH(got))
		}
	})

	t.Run("GMP-NEG refusals at the facade and forged rows against the guard's binding check", func(t *testing.T) {
		binding := e.gmpBinding(t, w.dave, gid)
		before := e.gmpRow(t, binding)
		labels, _ := e.gmpEvents(t, w.dave, gid)
		set := func(c context.Context, tc *TenantContext, subject tenancy.UserID) error {
			t.Helper()
			_, err := e.svc.SetGroupMemberLabel(c, tc, gid, subject, 2, proStr("Refused"))
			return err
		}
		fctx, ftc := e.as(t, w.frank, w.A)
		cctx, ctc := e.as(t, w.carol, w.A)
		actx, atc := e.as(t, w.alice, w.A)
		dctx, dtc := e.as(t, w.dave, w.A)
		bctx, btc := e.as(t, w.bob, w.A)
		for _, override := range []bool{false, true} {
			off := func() {}
			if override {
				off = e.withOverride(t, w.carol, w.A)
			}
			// An Organization owner or admin without the Group role: the Group is readable (group.read), the write is not.
			gmpWant(t, fmt.Sprintf("Organization admin (override %t)", override), set(fctx, ftc, w.dave.id), CodeForbidden, "")
			gmpWant(t, fmt.Sprintf("Organization owner (override %t)", override), set(cctx, ctc, w.dave.id), CodeForbidden, "")
			off()
		}
		gmpWant(t, "a peer edits the Group owner's label", set(dctx, dtc, w.bob.id), CodeForbidden, "")
		gmpWant(t, "a non-member editor", set(actx, atc, w.dave.id), CodeNotFound, "")
		gmpWant(t, "a subject without a membership", set(bctx, btc, w.alice.id), CodeNotFound, "")
		gctx, gtc := e.proIn(t, w.grace, w.A, w.design.ID)
		gmpWant(t, "a target outside the selected Group", set(gctx, gtc, w.dave.id), CodeNotFound, "")
		for _, c := range []struct {
			what     string
			expected int32
			label    string
			field    string
		}{
			{"101 code points", 2, strings.Repeat("\u00e9", 101), fieldRoleLabel},
			{"a control character", 2, "a\u0007b", fieldRoleLabel},
			{"a line separator", 2, "a\u2028b", fieldRoleLabel},
			{"a negative version", -1, "Fine", "expected_version"},
		} {
			_, err := e.svc.SetGroupMemberLabel(bctx, btc, gid, w.dave.id, c.expected, proStr(c.label))
			gmpWant(t, c.what, err, CodeInvalidInput, c.field)
		}
		if got := e.gmpRow(t, binding); got != before {
			t.Errorf("label row changed by a refused call: %s, want %s", itH(got), itH(before))
		}
		if l, _ := e.gmpEvents(t, w.dave, gid); l != labels {
			t.Errorf("label events %d after the refusals, want %d", l, labels)
		}

		// Raw runtime inserts as the Group owner with forged binding fields: the guard refuses each mismatch by name; the
		// unforged insert of another member's label (grace's, through the owner path) is admitted (rolled back).
		rt := e.rlsRuntime(t)
		insert := `INSERT INTO app.group_member_profiles (organization_id, binding_id, group_id, user_id, role_label) VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5)`
		rlsTx(t, rt, ptr(w.bob.id.String()), ptr(w.A.String()), ptr(""), nil, func(tx pgx.Tx) {
			for _, c := range []struct {
				what, state, constraint string
				args                    []any
				rows                    int64
			}{
				{"forged Organization", "23514", "group_member_profiles_binding", []any{w.B.String(), binding, gid.String(), w.dave.id.String(), "Forged"}, 0},
				{"forged Group", "23514", "group_member_profiles_binding", []any{w.A.String(), binding, w.design.ID.String(), w.dave.id.String(), "Forged"}, 0},
				{"forged user", "23514", "group_member_profiles_binding", []any{w.A.String(), binding, gid.String(), w.alice.id.String(), "Forged"}, 0},
				{"unforged, owner path", "", "", []any{w.A.String(), e.gmpBinding(t, w.grace, gid), gid.String(), w.grace.id.String(), "Co-lead"}, 1},
			} {
				st, cn, n := rlsStmtC(t, tx, insert, c.args...)
				if st != c.state || cn != c.constraint || n != c.rows {
					t.Errorf("raw %s: SQLSTATE %q constraint %q rows %d, want %q %q %d", c.what, st, cn, n, c.state, c.constraint, c.rows)
				}
			}
		})
		if got := e.gmpRows(t, w.grace); got != 0 {
			t.Errorf("grace holds %d label rows after the rolled-back raw insert, want 0", got)
		}
	})

	t.Run("GMP-7 the label goes with the membership on every removal path; readmission starts empty; no label event", func(t *testing.T) {
		bctx, btc := e.as(t, w.bob, w.A)
		cctx, ctc := e.as(t, w.carol, w.A)
		member := func() *itUser {
			t.Helper()
			u := e.newUser(t)
			e.grantOrg(t, w.A, u, tenancy.OrgMember)
			e.grantGroup(t, w.A, gid, u, tenancy.GroupMember)
			if _, err := e.svc.SetGroupMemberLabel(bctx, btc, gid, u.id, 0, proStr("Short-term role")); err != nil {
				t.Fatalf("label before removal: %v", err)
			}
			if n := e.gmpRows(t, u); n != 1 {
				t.Fatalf("%d label rows before removal, want 1", n)
			}
			return u
		}
		leaver, quitter, departer, retiree := member(), member(), member(), member()
		e.bind(t, quitter, w.A)
		paths := []struct {
			what   string
			u      *itUser
			remove func() error
		}{
			{"RemoveMember (the Group owner removes a member)", leaver, func() error { return e.svc.RemoveMember(bctx, btc, research, leaver.id) }},
			{"LeaveScope (the member leaves the Group)", quitter, func() error {
				qctx, qtc := e.as(t, quitter, w.A)
				return e.svc.LeaveScope(qctx, qtc, research)
			}},
			{"the Organization departure cascade (the owner removes the Organization member)", departer, func() error {
				return e.svc.RemoveMember(cctx, ctc, Scope{Kind: tenancy.ScopeOrganization}, departer.id)
			}},
			{"retirement", retiree, func() error {
				p, err := e.svc.RetireUser(ctx, retiree.id)
				if err == nil && !p.Disabled {
					return fmt.Errorf("retirement left the user enabled")
				}
				return err
			}},
		}
		for _, p := range paths {
			old := e.gmpBinding(t, p.u, gid)
			labels, removed := e.gmpEvents(t, p.u, gid)
			if err := p.remove(); err != nil {
				t.Errorf("%s: %v", p.what, err)
				continue
			}
			if e.gmpBinding(t, p.u, gid) != "" || e.gmpRow(t, old) != "" || e.gmpRows(t, p.u) != 0 {
				t.Errorf("%s: the membership or its label survived", p.what)
			}
			if l, r := e.gmpEvents(t, p.u, gid); l != labels || r != removed+1 {
				t.Errorf("%s: %d label events (want %d) and %d membership.removed rows (want %d)", p.what, l, labels, r, removed+1)
			}
		}
		// Readmission: a new binding without a label, visible as an empty card to the Organization owner and the member.
		old := e.gmpBinding(t, leaver, gid)
		labels, _ := e.gmpEvents(t, leaver, gid)
		e.grantGroup(t, w.A, gid, leaver, tenancy.GroupMember)
		now := e.gmpBinding(t, leaver, gid)
		if now == "" || now == old || e.gmpRow(t, now) != "" {
			t.Fatalf("readmission: a new binding %t, a label row %t", now != "" && now != old, e.gmpRow(t, now) != "")
		}
		cs, err := e.svc.ListGroupMemberCards(cctx, ctc, leaver.id)
		if err != nil || len(cs.Cards) != 1 || cs.Cards[0].GroupID != gid || gmpCard(cs.Cards[0]) != "-:false:0" {
			t.Errorf("readmitted member's card for the Organization owner: %d cards, %v", len(cs.Cards), err)
		}
		if l, _ := e.gmpEvents(t, leaver, gid); l != labels {
			t.Errorf("readmission recorded %d label events, want %d", l, labels)
		}
	})

	t.Run("GMP-8 trigger-owned audit rows: admin edits name the editor, a self edit the member; never the text", func(t *testing.T) {
		dctx, dtc := e.as(t, w.dave, w.A)
		if card, err := e.svc.SetGroupMemberLabel(dctx, dtc, gid, w.dave.id, 2, proStr("In my own words")); err != nil || gmpCard(card) != "In my own words:true:3" {
			t.Fatalf("self edit: %s, %v", gmpCard(card), err)
		}
		binding := e.gmpBinding(t, w.dave, gid)
		row := func(actor *itUser) int {
			t.Helper()
			var n int
			if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM audit.events WHERE action = 'group_member_profile.updated'
  AND outcome = 'success' AND reason = 'role_label_changed' AND actor_kind = 'user' AND actor_user_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid AND organization_id = $3::text::uuid AND group_id = $4::text::uuid
  AND resource_type = 'group_member_profile' AND resource_id = $5::text::uuid AND operation IS NULL AND role IS NULL
  AND previous_role IS NULL`, actor.id.String(), w.dave.id.String(), w.A.String(), gid.String(), binding).Scan(&n); err != nil {
				t.Fatalf("audit rows: %v", err)
			}
			return n
		}
		for _, c := range []struct {
			what  string
			actor *itUser
		}{{"the Group owner's set", w.bob}, {"the Group admin's replace", w.grace}, {"the member's own edit", w.dave}} {
			if n := row(c.actor); n != 1 {
				t.Errorf("%s: %d rows with the exact shape, want 1", c.what, n)
			}
		}
		if labels, _ := e.gmpEvents(t, w.dave, gid); labels != 3 {
			t.Errorf("%d label events for the member, want 3 (the no-op repeat and the refusals record nothing)", labels)
		}
		for _, text := range []string{"Lab lead", "Principal investigator", "In my own words"} {
			var n int
			if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM audit.events e WHERE position($1 in e::text) > 0`, text).Scan(&n); err != nil || n != 0 {
				t.Errorf("%d audit rows carry a label text (%v)", n, err)
			}
		}
	})

	t.Run("GMP-11 the facade reads: one not_found identity, the oversight cards under both override settings, selected context", func(t *testing.T) {
		actx, atc := e.as(t, w.alice, w.A)
		_, err := e.svc.ListGroupMemberCards(actx, atc, w.dave.id)
		gmpWant(t, "an unrelated viewer", err, CodeNotFound, "")
		_, perr := e.svc.GetAccountProfile(actx, atc, w.dave.id)
		gmpWant(t, "the same viewer's profile read", perr, CodeNotFound, "")
		hctx, htc := e.as(t, w.hostile, w.B)
		_, err = e.svc.ListGroupMemberCards(hctx, htc, w.dave.id)
		gmpWant(t, "a foreign viewer", err, CodeNotFound, "")
		cctx, ctc := e.as(t, w.carol, w.A)
		for _, override := range []bool{false, true} {
			off := func() {}
			if override {
				off = e.withOverride(t, w.carol, w.A)
			}
			cs, err := e.svc.ListGroupMemberCards(cctx, ctc, w.dave.id)
			if err != nil || cs.NarrowRequired || len(cs.Cards) != 1 || cs.Cards[0].GroupID != gid || gmpCard(cs.Cards[0]) != "In my own words:false:3" {
				t.Errorf("Organization owner oversight (override %t): %d cards, %v", override, len(cs.Cards), err)
			}
			off()
		}
		// Selected context: the selected Group only; another selectable Group shows no card of a visible subject.
		dctx, dtc := e.proIn(t, w.dave, w.A, gid)
		if cs, err := e.svc.ListGroupMemberCards(dctx, dtc, w.bob.id); err != nil || len(cs.Cards) != 1 || cs.Cards[0].GroupID != gid {
			t.Errorf("selected research, the member reads the owner: %d cards, %v", len(cs.Cards), err)
		}
		gctx, gtc := e.proIn(t, w.grace, w.A, w.design.ID)
		if cs, err := e.svc.ListGroupMemberCards(gctx, gtc, w.dave.id); err != nil || len(cs.Cards) != 0 || cs.NarrowRequired {
			t.Errorf("selected design, an Organization admin reads the member: %d cards (narrow %t), %v", len(cs.Cards), cs.NarrowRequired, err)
		}
	})

	t.Run("GMP-13 overflow: more than 200 authorized cards narrow, a selected Group shows one, unauthorized cards never count", func(t *testing.T) {
		o := e.org(t, "Group Profiles Overflow "+e.canaries[2])
		owner, peer, subject := e.newUser(t), e.newUser(t), e.newUser(t)
		e.grantOrg(t, o, owner, tenancy.OrgOwner)
		e.grantOrg(t, o, peer, tenancy.OrgMember)
		e.grantOrg(t, o, subject, tenancy.OrgMember)
		e.bind(t, owner, o)
		e.bind(t, peer, o)
		tag := strings.ReplaceAll(rlsUUID(t)[:8], "-", "")
		for _, s := range []struct {
			sql  string
			args []any
		}{
			{`INSERT INTO app.groups (id, organization_id, name, slug)
SELECT gen_random_uuid(), $1::text::uuid, 'Overflow ' || i, 'gmp-facade-' || $2::text || '-' || i FROM generate_series(1, 201) AS i`, []any{o.String(), tag}},
			{`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
SELECT g.organization_id, $2::text::uuid, 'group', g.id, g.id, 'group_member', 'membership' FROM app.groups g WHERE g.organization_id = $1::text::uuid`,
				[]any{o.String(), subject.id.String()}},
			{`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
SELECT g.organization_id, $2::text::uuid, 'group', g.id, g.id, 'group_member', 'membership' FROM app.groups g WHERE g.organization_id = $1::text::uuid
ORDER BY g.id LIMIT 1`, []any{o.String(), peer.id.String()}},
		} {
			if _, err := e.setup.Exec(ctx, s.sql, s.args...); err != nil {
				t.Fatalf("overflow fixture: %v", err)
			}
		}
		var firstText, middleText string
		if err := e.setup.QueryRow(ctx, `SELECT (SELECT id::text FROM app.groups WHERE organization_id = $1::text::uuid ORDER BY id LIMIT 1),
  (SELECT id::text FROM app.groups WHERE organization_id = $1::text::uuid ORDER BY id OFFSET 100 LIMIT 1)`, o.String()).Scan(&firstText, &middleText); err != nil {
			t.Fatalf("overflow fixture: %v", err)
		}
		first, err := tenancy.ParseGroupID(firstText)
		if err != nil {
			t.Fatal(err)
		}
		middle, err := tenancy.ParseGroupID(middleText)
		if err != nil {
			t.Fatal(err)
		}
		octx, otc := e.as(t, owner, o)
		if cs, err := e.svc.ListGroupMemberCards(octx, otc, subject.id); err != nil || !cs.NarrowRequired || len(cs.Cards) != 0 || cs.Cards == nil {
			t.Errorf("Organization owner, Organization-only: %d cards (narrow %t), %v; want none and narrow_required", len(cs.Cards), cs.NarrowRequired, err)
		}
		mctx, mtc := e.proIn(t, owner, o, middle)
		if cs, err := e.svc.ListGroupMemberCards(mctx, mtc, subject.id); err != nil || cs.NarrowRequired || len(cs.Cards) != 1 || cs.Cards[0].GroupID != middle {
			t.Errorf("Organization owner, selected Group: %d cards (narrow %t), %v; want exactly that card", len(cs.Cards), cs.NarrowRequired, err)
		}
		pctx, ptc := e.as(t, peer, o)
		if cs, err := e.svc.ListGroupMemberCards(pctx, ptc, subject.id); err != nil || cs.NarrowRequired || len(cs.Cards) != 1 || cs.Cards[0].GroupID != first {
			t.Errorf("a viewer sharing one Group: %d cards (narrow %t), %v; want exactly that card", len(cs.Cards), cs.NarrowRequired, err)
		}
	})
}
