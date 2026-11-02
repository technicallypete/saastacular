//go:build integration

// G075 B-GROUP-PROFILE-IT, second and third selectors (Plan r2 §3.4, §3.6, §5, §8 L1), against the access integration
// environment (setupIT: the private PostgreSQL of the migrations suite, no MinIO).
//
// TestGroupProfileOrdering runs the label write against the races the Spec names, with the G007 barriers of
// TestLabelsOrdering (a pause before COMMIT and an observed lock wait, no sleeps):
//
//   - O-GMP-1 two concurrent first edits of one membership's label, different and identical values: serialized on the
//     Group row (both writes lock it FOR NO KEY UPDATE first); one row, one audit row, the second edit a version
//     conflict (different value) or an idempotent success (identical value);
//   - O-GMP-2 a label edit racing the member's removal in both orders: edit first → the removal waits and then cascades
//     the committed label away; removal first → the edit waits and observes not_found; never an orphan row, never a
//     label event after membership.removed;
//   - O-GMP-3 an admin's edit racing the admin's own demotion to member: edit first → both commit and the demoted
//     admin's next edit is forbidden; demotion first → the edit waits and is forbidden with zero effects.
//
// TestGroupMemberProfileMatrix holds the GMP-* cases EP-1's TestGroupMemberProfile (group_profile_integration_test.go,
// kept unchanged as proven) does not: GMP-1 self set, replace and clear with the version chain, GMP-5 the label grammar
// at the facade, GMP-6 stale and identical writes, GMP-8 the audit-fault rollback, GMP-10 the denials attributed to the
// two operations, GMP-11 the rest of the facade matrix (S1-S4, N1-N7, the selected-Group rows, live membership and
// Organization-role revocation) and GMP-12 the no-authority proof (W4). It lives in this file rather than
// group_profile_integration_test.go (a recorded file-placement deviation of lane L1b); select both with
// -run '^TestGroupMemberProfile(Matrix)?$'. Failure messages print codes, counts and hashes only.
package access

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"core/internal/tenancy"
)

// gmpDenials counts the access.denied rows of actor for op with reason (any scope or resource).
func (e *itEnv) gmpDenials(t *testing.T, actor *itUser, op, reason string) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*)::int FROM audit.events WHERE action = 'access.denied'
  AND actor_user_id = $1::text::uuid AND operation = $2 AND ($3::text = '' OR reason = $3::text)`, actor.id.String(), op, reason).Scan(&n); err != nil {
		t.Fatalf("denials: %v", err)
	}
	return n
}

// gmpLateLabels counts u's label events in Group g recorded after u's last membership.removed in g (superuser read).
func (e *itEnv) gmpLateLabels(t *testing.T, u *itUser, g tenancy.GroupID) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*)::int FROM audit.events l WHERE l.action = 'group_member_profile.updated'
  AND l.subject_user_id = $1::text::uuid AND l.group_id = $2::text::uuid AND l.occurred_at > (SELECT max(r.occurred_at) FROM audit.events r
  WHERE r.action = 'membership.removed' AND r.subject_user_id = $1::text::uuid AND r.group_id = $2::text::uuid)`,
		u.id.String(), g.String()).Scan(&n); err != nil {
		t.Fatalf("late label events: %v", err)
	}
	return n
}

func TestGroupProfileOrdering(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	zeroG := tenancy.GroupID{}
	seq := 0
	next := func() int { seq++; return seq }
	org := e.org(t, "Group Profile Ordering "+e.canaries[0])
	creator := e.newUser(t)
	// member returns a new Organization member of org with the given Group roles, bound to org (before any window).
	member := func(t *testing.T, groups map[tenancy.GroupID]tenancy.GroupRole) *itUser {
		t.Helper()
		u := e.newUser(t)
		e.grantOrg(t, org, u, tenancy.OrgMember)
		for g, r := range groups {
			e.grantGroup(t, org, g, u, r)
		}
		e.bind(t, u, org)
		return u
	}
	newGroup := func(t *testing.T) tenancy.GroupID {
		t.Helper()
		return e.group(t, org, creator, "GMP Ordering Group", "gmp-ordering-"+itH(next())[:8]).ID
	}
	// as resolves u's Organization-only context in org in a fresh request scope of a separate service (its own hookDB).
	type actorCtx struct {
		svc  *Service
		hook *hookDB
		rctx context.Context
		tc   *TenantContext
	}
	as := func(t *testing.T, u *itUser) actorCtx {
		t.Helper()
		svc, h := e.service(e.id)
		rctx, _ := beginAs(t, svc, ctx, u, org)
		return actorCtx{svc: svc, hook: h, rctx: rctx, tc: resolve(t, svc, rctx, org, zeroG)}
	}
	// serialize runs first paused before its COMMIT, starts second, waits until second waits for a lock, then releases
	// first; it returns both errors and whether first finished before second.
	serialize := func(t *testing.T, hFirst *hookDB, first, second func() error) (error, error, bool) {
		t.Helper()
		p := newPause()
		hFirst.on("before", "COMMIT", p.hook)
		a := start(first)
		p.wait(t, "first")
		b := start(second)
		e.waitWaiters(t, 1)
		close(p.release)
		ea, eb := a.wait(t, "first"), b.wait(t, "second")
		hFirst.reset()
		return ea, eb, a.at < b.at
	}
	setLabel := func(c actorCtx, g tenancy.GroupID, subject *itUser, expected int32, label string) func() error {
		return func() error {
			_, err := c.svc.SetGroupMemberLabel(c.rctx, c.tc, g, subject.id, expected, &label)
			return err
		}
	}

	t.Run("O-GMP-1 two concurrent first edits: one row, one audit row; the second conflicts or is idempotent", func(t *testing.T) {
		g := newGroup(t)
		owner := member(t, map[tenancy.GroupID]tenancy.GroupRole{g: tenancy.GroupOwner})
		admin := member(t, map[tenancy.GroupID]tenancy.GroupRole{g: tenancy.GroupAdmin})
		for _, c := range []struct {
			what, first, second, row string
			conflict                 bool
		}{
			{"different values", "Left role", "Right role", "Left role|1", true},
			{"identical values", "Same role", "Same role", "Same role|1", false},
		} {
			subject := member(t, map[tenancy.GroupID]tenancy.GroupRole{g: tenancy.GroupMember})
			oc, ac := as(t, owner), as(t, admin)
			e1, e2, inOrder := serialize(t, oc.hook, setLabel(oc, g, subject, 0, c.first), setLabel(ac, g, subject, 0, c.second))
			if e1 != nil || !inOrder {
				t.Errorf("%s: the first edit %v (finished first %t), want it committed first", c.what, e1, inOrder)
			}
			if c.conflict {
				gmpWant(t, c.what+": the second edit", e2, CodeConflict, fieldVersionConflict)
			} else if e2 != nil {
				t.Errorf("%s: the second edit %v, want the idempotent success", c.what, e2)
			}
			if got := e.gmpRow(t, e.gmpBinding(t, subject, g)); got != c.row {
				t.Errorf("%s: label row %s, want one row at version 1", c.what, itH(got))
			}
			if l, _ := e.gmpEvents(t, subject, g); l != 1 {
				t.Errorf("%s: %d label events, want 1", c.what, l)
			}
			ordering(t, map[string]any{"case": "O-GMP-1", "values": c.what})
		}
	})

	t.Run("O-GMP-2 a label edit racing the member's removal in both orders: no orphan, no label event after the removal", func(t *testing.T) {
		g := newGroup(t)
		owner := member(t, map[tenancy.GroupID]tenancy.GroupRole{g: tenancy.GroupOwner})
		admin := member(t, map[tenancy.GroupID]tenancy.GroupRole{g: tenancy.GroupAdmin})
		scope := Scope{Kind: tenancy.ScopeGroup, Group: &g}
		for _, order := range []string{"edit first", "removal first"} {
			subject := member(t, map[tenancy.GroupID]tenancy.GroupRole{g: tenancy.GroupMember})
			binding := e.gmpBinding(t, subject, g)
			ac, oc := as(t, admin), as(t, owner)
			edit := setLabel(ac, g, subject, 0, "Racing role")
			remove := func() error { return oc.svc.RemoveMember(oc.rctx, oc.tc, scope, subject.id) }
			wantLabels := 0
			switch order {
			case "edit first":
				ee, er, inOrder := serialize(t, ac.hook, edit, remove)
				if ee != nil || er != nil || !inOrder {
					t.Errorf("%s: edit %v, removal %v (edit finished first %t), want both, the edit first", order, ee, er, inOrder)
				}
				wantLabels = 1
			case "removal first":
				er, ee, _ := serialize(t, oc.hook, remove, edit)
				if er != nil {
					t.Errorf("%s: removal %v", order, er)
				}
				gmpWant(t, order+": the edit after the removal", ee, CodeNotFound, "")
			}
			if e.gmpBinding(t, subject, g) != "" || e.gmpRow(t, binding) != "" || e.gmpRows(t, subject) != 0 {
				t.Errorf("%s: the membership or its label survived (an orphan row)", order)
			}
			if l, r := e.gmpEvents(t, subject, g); l != wantLabels || r != 1 {
				t.Errorf("%s: %d label events (want %d), %d membership.removed rows (want 1)", order, l, wantLabels, r)
			}
			if n := e.gmpLateLabels(t, subject, g); n != 0 {
				t.Errorf("%s: %d label events after membership.removed, want 0", order, n)
			}
			ordering(t, map[string]any{"case": "O-GMP-2", "order": order})
		}
	})

	t.Run("O-GMP-3 an admin's edit racing the admin's own demotion: the demoted admin's edit refuses", func(t *testing.T) {
		g := newGroup(t)
		owner := member(t, map[tenancy.GroupID]tenancy.GroupRole{g: tenancy.GroupOwner})
		scope := Scope{Kind: tenancy.ScopeGroup, Group: &g}
		for _, order := range []string{"edit first", "demotion first"} {
			admin := member(t, map[tenancy.GroupID]tenancy.GroupRole{g: tenancy.GroupAdmin})
			subject := member(t, map[tenancy.GroupID]tenancy.GroupRole{g: tenancy.GroupMember})
			binding := e.gmpBinding(t, subject, g)
			ac, oc := as(t, admin), as(t, owner)
			edit := setLabel(ac, g, subject, 0, "Admin role")
			demote := func() error {
				_, err := oc.svc.ChangeMemberRole(oc.rctx, oc.tc, scope, admin.id, string(tenancy.GroupMember))
				return err
			}
			switch order {
			case "edit first":
				ee, ed, inOrder := serialize(t, ac.hook, edit, demote)
				if ee != nil || ed != nil || !inOrder {
					t.Errorf("%s: edit %v, demotion %v (edit finished first %t), want both, the edit first", order, ee, ed, inOrder)
				}
				if got := e.gmpRow(t, binding); got != "Admin role|1" {
					t.Errorf("%s: label row %s, want the edit at version 1", order, itH(got))
				}
				// The demoted admin's next edit, from the same context, is forbidden (the Group stays readable to a member).
				_, err := ac.svc.SetGroupMemberLabel(ac.rctx, ac.tc, g, subject.id, 1, proStr("After the demotion"))
				gmpWant(t, order+": the demoted admin's next edit", err, CodeForbidden, "")
				if got := e.gmpRow(t, binding); got != "Admin role|1" {
					t.Errorf("%s: label row %s after the refused edit, want version 1 unchanged", order, itH(got))
				}
				if l, _ := e.gmpEvents(t, subject, g); l != 1 {
					t.Errorf("%s: %d label events, want 1", order, l)
				}
			case "demotion first":
				ed, ee, _ := serialize(t, oc.hook, demote, edit)
				if ed != nil {
					t.Errorf("%s: demotion %v", order, ed)
				}
				gmpWant(t, order+": the edit after the demotion", ee, CodeForbidden, "")
				if got := e.gmpRow(t, binding); got != "" {
					t.Errorf("%s: a label row %s after the refused edit, want none", order, itH(got))
				}
				if l, _ := e.gmpEvents(t, subject, g); l != 0 {
					t.Errorf("%s: %d label events, want 0", order, l)
				}
			}
			ordering(t, map[string]any{"case": "O-GMP-3", "order": order})
		}
	})
}

func TestGroupMemberProfileMatrix(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	w := e.gmpWorld(t)
	gid := w.research.ID
	research := Scope{Kind: tenancy.ScopeGroup, Group: &gid}
	set := func(c context.Context, tc *TenantContext, g tenancy.GroupID, subject *itUser, expected int32, label *string) (GroupMemberCard, error) {
		t.Helper()
		return e.svc.SetGroupMemberLabel(c, tc, g, subject.id, expected, label)
	}
	// names labels the fixture Groups in cards output; an unexpected Group is "?".
	names := map[tenancy.GroupID]string{w.research.ID: "research", w.design.ID: "design"}
	// cards renders a viewer's cards of subject as sorted "group:label:edit:version" entries (plus "narrow" when the
	// overflow signal is set); a refusal is "error:<code>".
	cards := func(c context.Context, tc *TenantContext, subject *itUser) string {
		t.Helper()
		cs, err := e.svc.ListGroupMemberCards(c, tc, subject.id)
		if err != nil {
			return "error:" + string(CodeOf(err))
		}
		var out []string
		for _, card := range cs.Cards {
			n, ok := names[card.GroupID]
			if !ok {
				n = "?"
			}
			out = append(out, n+":"+gmpCard(card))
		}
		slices.Sort(out)
		if cs.NarrowRequired {
			out = append(out, "narrow")
		}
		return strings.Join(out, ",")
	}
	// current is subject's card for research as viewer reads it (the test fails without exactly one).
	current := func(c context.Context, tc *TenantContext, subject *itUser) GroupMemberCard {
		t.Helper()
		cs, err := e.svc.ListGroupMemberCards(c, tc, subject.id)
		if err != nil || len(cs.Cards) != 1 || cs.Cards[0].GroupID != gid {
			t.Fatalf("the research card: %d cards, %v", len(cs.Cards), err)
		}
		return cs.Cards[0]
	}
	notFound := "error:" + string(CodeNotFound)

	t.Run("GMP-1 self set, replace and clear with the version chain; a clear without a row writes nothing", func(t *testing.T) {
		dctx, dtc := e.as(t, w.dave, w.A)
		binding := e.gmpBinding(t, w.dave, gid)
		if card, err := set(dctx, dtc, gid, w.dave, 0, nil); err != nil || gmpCard(card) != "-:true:0" {
			t.Fatalf("clear without a label row: %s, %v", gmpCard(card), err)
		}
		if got := e.gmpRow(t, binding); got != "" {
			t.Errorf("a clear without a label row wrote %s", itH(got))
		}
		for _, s := range []struct {
			what     string
			expected int32
			label    *string
			card     string
			row      string
			events   int
		}{
			{"set", 0, proStr("First role"), "First role:true:1", "First role|1", 1},
			{"replace", 1, proStr("Second role"), "Second role:true:2", "Second role|2", 2},
			{"clear with an empty value", 2, proStr(""), "-:true:3", "-|3", 3},
			{"clear again (no-op)", 3, nil, "-:true:3", "-|3", 3},
			{"a blank value on a cleared label (no-op)", 3, proStr(" \u3000 "), "-:true:3", "-|3", 3},
			{"set after the clear", 3, proStr("Third role"), "Third role:true:4", "Third role|4", 4},
		} {
			card, err := set(dctx, dtc, gid, w.dave, s.expected, s.label)
			if err != nil || gmpCard(card) != s.card || card.GroupID != gid || card.GroupName != w.research.Name {
				t.Fatalf("%s: %s, %v", s.what, gmpCard(card), err)
			}
			if got := e.gmpRow(t, binding); got != s.row {
				t.Errorf("%s: label row %s", s.what, itH(got))
			}
			if l, _ := e.gmpEvents(t, w.dave, gid); l != s.events {
				t.Errorf("%s: %d label events, want %d", s.what, l, s.events)
			}
			// Each write is what the next read shows.
			if got := cards(dctx, dtc, w.dave); got != "research:"+s.card {
				t.Errorf("%s: the next read shows %s", s.what, itH(got))
			}
		}
	})

	t.Run("GMP-5 the label grammar at the facade: trimming, the code-point bound, controls and invalid UTF-8", func(t *testing.T) {
		gctx, gtc := e.as(t, w.grace, w.A)
		binding := e.gmpBinding(t, w.grace, gid)
		version := int32(0)
		for _, c := range []struct{ what, label, stored string }{
			{"Unicode white space trimmed", "\u00a0\u2003 Lab manager \u3000", "Lab manager"},
			{"100 code points", strings.Repeat("\u00e9", 100), strings.Repeat("\u00e9", 100)},
			{"100 supplementary code points", strings.Repeat("\U0001F52C", 100), strings.Repeat("\U0001F52C", 100)},
			{"markup kept as text", `<b>bold</b> & "quotes"`, `<b>bold</b> & "quotes"`},
		} {
			card, err := set(gctx, gtc, gid, w.grace, version, proStr(c.label))
			if err != nil || card.RoleLabel == nil || *card.RoleLabel != c.stored || card.Version != version+1 {
				t.Fatalf("%s: version %d, stored as given %t, %v", c.what, card.Version, card.RoleLabel != nil && *card.RoleLabel == c.stored, err)
			}
			version = card.Version
		}
		before := e.gmpRow(t, binding)
		labels, _ := e.gmpEvents(t, w.grace, gid)
		for _, c := range []struct{ what, label string }{
			{"101 supplementary code points", strings.Repeat("\U0001F52C", 101)},
			{"a tab", "a\tb"},
			{"a next-line control", "a\u0085b"},
			{"a paragraph separator", "a\u2029b"},
			{"invalid UTF-8", "a\xffb"},
		} {
			_, err := set(gctx, gtc, gid, w.grace, version, proStr(c.label))
			gmpWant(t, c.what, err, CodeInvalidInput, fieldRoleLabel)
		}
		if got := e.gmpRow(t, binding); got != before {
			t.Errorf("label row changed by a refused call: %s, want %s", itH(got), itH(before))
		}
		if l, _ := e.gmpEvents(t, w.grace, gid); l != labels {
			t.Errorf("label events %d after the refusals, want %d", l, labels)
		}
	})

	t.Run("GMP-6 a different value on a stale or future version conflicts with zero effects; an identical one is a silent success", func(t *testing.T) {
		bctx, btc := e.as(t, w.bob, w.A)
		binding := e.gmpBinding(t, w.dave, gid)
		cur := current(bctx, btc, w.dave)
		if cur.RoleLabel == nil || cur.Version < 2 || !cur.CanEdit {
			t.Fatalf("the research card before the stale writes: version %d, label set %t, edit %t", cur.Version, cur.RoleLabel != nil, cur.CanEdit)
		}
		before := e.gmpRow(t, binding)
		labels, _ := e.gmpEvents(t, w.dave, gid)
		for _, expected := range []int32{cur.Version - 1, cur.Version + 1} {
			_, err := set(bctx, btc, gid, w.dave, expected, proStr("Stale role"))
			gmpWant(t, fmt.Sprintf("a different value at version %d (current %d)", expected, cur.Version), err, CodeConflict, fieldVersionConflict)
		}
		if card, err := set(bctx, btc, gid, w.dave, cur.Version-1, cur.RoleLabel); err != nil || gmpCard(card) != gmpCard(cur) {
			t.Errorf("the identical value on a stale version: %s, %v", gmpCard(card), err)
		}
		if got := e.gmpRow(t, binding); got != before {
			t.Errorf("label row %s after the stale writes, want %s", itH(got), itH(before))
		}
		if l, _ := e.gmpEvents(t, w.dave, gid); l != labels {
			t.Errorf("label events %d after the stale writes, want %d", l, labels)
		}
	})

	t.Run("GMP-8 an audit fault refuses the label write with neither the change nor an event", func(t *testing.T) {
		dctx, dtc := e.as(t, w.dave, w.A)
		binding := e.gmpBinding(t, w.dave, gid)
		cur := current(dctx, dtc, w.dave)
		before := e.gmpRow(t, binding)
		labels, _ := e.gmpEvents(t, w.dave, gid)
		name := fmt.Sprintf("gmp_audit_fault_%d", itSeq.Add(1))
		if _, err := e.setup.Exec(ctx, `ALTER TABLE audit.events ADD CONSTRAINT `+name+` CHECK (action <> 'group_member_profile.updated') NOT VALID`); err != nil {
			t.Fatal(err)
		}
		planted := true
		remove := func() {
			if !planted {
				return
			}
			planted = false
			if _, err := e.setup.Exec(ctx, `ALTER TABLE audit.events DROP CONSTRAINT `+name); err != nil {
				t.Errorf("remove the audit fault: %v", err)
			}
		}
		t.Cleanup(remove)
		_, ferr := set(dctx, dtc, gid, w.dave, cur.Version, proStr("Faulted role"))
		remove()
		if ferr == nil {
			t.Error("the label write under a group member profile audit fault succeeded")
		}
		if got := e.gmpRow(t, binding); got != before {
			t.Errorf("label row %s after the faulted write, want %s", itH(got), itH(before))
		}
		if l, _ := e.gmpEvents(t, w.dave, gid); l != labels {
			t.Errorf("label events %d after the faulted write, want %d", l, labels)
		}
		// The paired positive once the fault is gone: the write commits with its one event.
		if card, err := set(dctx, dtc, gid, w.dave, cur.Version, proStr("After the fault")); err != nil || card.Version != cur.Version+1 {
			t.Fatalf("the write after the fault: version %d, %v", card.Version, err)
		}
		if l, _ := e.gmpEvents(t, w.dave, gid); l != labels+1 {
			t.Errorf("label events %d after the paired write, want %d", l, labels+1)
		}
	})

	t.Run("GMP-10 denials are attributed to the two operations", func(t *testing.T) {
		actx, atc := e.as(t, w.alice, w.A)
		dctx, dtc := e.as(t, w.dave, w.A)
		// A subject the caller may not read: not_found, recorded as profile.groups.list target_not_found in the
		// Organization, without the subject.
		before := e.proDenials(t, w.alice, opProfileGroupsList, "target_not_found", &w.A)
		_, err := e.svc.ListGroupMemberCards(actx, atc, w.dave.id)
		gmpWant(t, "an unrelated subject's cards", err, CodeNotFound, "")
		if got := e.proDenials(t, w.alice, opProfileGroupsList, "target_not_found", &w.A); got != before+1 {
			t.Errorf("profile.groups.list target_not_found denials %d -> %d, want one more", before, got)
		}
		// The label write: a peer on a readable Group (forbidden, role_not_permitted) and a non-member (not_found).
		for _, c := range []struct {
			what   string
			actor  *itUser
			ctx    context.Context
			tc     *TenantContext
			target *itUser
			code   Code
			reason string
		}{
			{"a peer edits the Group owner's label", w.dave, dctx, dtc, w.bob, CodeForbidden, "role_not_permitted"},
			{"a non-member edits a member's label", w.alice, actx, atc, w.dave, CodeNotFound, ""},
		} {
			b := e.gmpDenials(t, c.actor, opProfileGroupLabelUpdate, c.reason)
			_, err := set(c.ctx, c.tc, gid, c.target, 0, proStr("Denied role"))
			gmpWant(t, c.what, err, c.code, "")
			if got := e.gmpDenials(t, c.actor, opProfileGroupLabelUpdate, c.reason); got != b+1 {
				t.Errorf("%s: profile.group_label.update denials %d -> %d, want one more", c.what, b, got)
			}
		}
	})

	t.Run("GMP-11 the facade matrix: S1-S4, N1-N7, the selected-Group rows and live revocation", func(t *testing.T) {
		c := e.canaries
		bctx, btc := e.as(t, w.bob, w.A)
		cur := current(bctx, btc, w.dave)
		labelled, err := set(bctx, btc, gid, w.dave, cur.Version, proStr("Matrix role"))
		if err != nil {
			t.Fatalf("the Group owner labels the member: %v", err)
		}
		v := labelled.Version
		// dave's research card with and without the viewer's edit hint.
		withEdit, readOnly := fmt.Sprintf("research:Matrix role:true:%d", v), fmt.Sprintf("research:Matrix role:false:%d", v)
		// Group Z (alice owns it; dave joins): the sibling of research. dana: a member of design only, which shares one
		// Business Profile with research. bianca: a member of Organization B only.
		z := e.group(t, w.A, w.alice, "GMP Z "+c[0], "gmp-z").ID
		names[z] = "z"
		e.grantGroup(t, w.A, z, w.dave, tenancy.GroupMember)
		shared := e.profile(t, w.A, "GMP Shared "+c[1])
		e.link(t, w.A, gid, shared.ID)
		e.link(t, w.A, w.design.ID, shared.ID)
		dana, bianca := e.newUser(t), e.newUser(t)
		e.grantOrg(t, w.A, dana, tenancy.OrgMember)
		e.grantGroup(t, w.A, w.design.ID, dana, tenancy.GroupMember)
		e.bind(t, dana, w.A)
		e.grantOrg(t, w.B, bianca, tenancy.OrgMember)

		dctx, dtc := e.as(t, w.dave, w.A)
		drctx, drtc := e.proIn(t, w.dave, w.A, gid)
		dzctx, dztc := e.proIn(t, w.dave, w.A, z)
		brctx, brtc := e.proIn(t, w.bob, w.A, gid)
		gctx, gtc := e.as(t, w.grace, w.A)
		actx, atc := e.as(t, w.alice, w.A)
		azctx, aztc := e.proIn(t, w.alice, w.A, z)
		hctx, htc := e.as(t, w.hostile, w.A)
		hbctx, hbtc := e.as(t, w.hostile, w.B)
		nctx, ntc := e.as(t, dana, w.A)
		for _, r := range []struct {
			what    string
			ctx     context.Context
			tc      *TenantContext
			subject *itUser
			want    string
		}{
			{"S1 self", dctx, dtc, w.dave, withEdit + ",z:-:true:0"},
			{"S1 self, research selected", drctx, drtc, w.dave, withEdit},
			{"S1 self, Z selected", dzctx, dztc, w.dave, "z:-:true:0"},
			{"S2 the research owner reads the member (no Z card)", bctx, btc, w.dave, withEdit},
			{"S2 the member reads the research owner read-only", dctx, dtc, w.bob, "research:-:false:0"},
			{"S3 the research admin (also an Organization admin: the Z card read-only) reads the member", gctx, gtc, w.dave, withEdit + ",z:-:false:0"},
			{"N2 sibling: Z's owner sees only the Z card", actx, atc, w.dave, "z:-:true:0"},
			{"N3 Z's owner from Z's members page", azctx, aztc, w.dave, "z:-:true:0"},
			{"N3 the research owner from research's members page", brctx, brtc, w.dave, withEdit},
			{"N3 Z selected: a visible subject without a Z card", dzctx, dztc, w.bob, ""},
			{"N1 no shared Group", hctx, htc, w.bob, notFound},
			{"N1 Z's owner and the research owner share nothing", actx, atc, w.bob, notFound},
			{"N6 a shared Business Profile confers nothing", nctx, ntc, w.dave, notFound},
			{"N6 the reverse", dctx, dtc, dana, notFound},
			{"N7 a subject of Organization B only", hctx, htc, bianca, notFound},
			{"N7 the same from a plain member", actx, atc, bianca, notFound},
			{"N7 a foreign viewer on host B", hbctx, hbtc, w.dave, notFound},
		} {
			if got := cards(r.ctx, r.tc, r.subject); got != r.want {
				t.Errorf("%s: cards %s (hash), want %s (hash)", r.what, itH(got), itH(r.want))
			}
		}
		// S4: the Organization owner and admin read every card of the subject read-only, under both override settings.
		for _, override := range []bool{false, true} {
			off := func() {}
			if override {
				off = e.withOverride(t, w.carol, w.A)
			}
			for _, u := range []*itUser{w.carol, w.frank} {
				uctx, utc := e.as(t, u, w.A)
				if got := cards(uctx, utc, w.dave); got != readOnly+",z:-:false:0" {
					t.Errorf("S4 oversight by %s (override %t): cards %s (hash)", itH(u.id), override, itH(got))
				}
			}
			off()
		}
		// N3 selected-other denial: research's owner cannot select Z (no binding there); a target outside the selected
		// Group is not_found.
		rctx, _ := beginAs(t, e.svc, ctx, w.bob, w.A)
		_, err = e.svc.Resolve(rctx, ResolveRequest{OrganizationID: w.A.String(), GroupID: z.String()})
		gmpWant(t, "the research owner selects Z", err, CodeContextDenied, "")
		_, err = set(dzctx, dztc, gid, w.dave, v, proStr("Outside the selected Group"))
		gmpWant(t, "a research write with Z selected", err, CodeNotFound, "")

		// N4 live membership revocation and readmission: dave labels himself in Z; Z's owner removes him (the next read
		// in the same request scope is not_found and the label row is gone) and readmits him (the card starts empty).
		if _, err := set(dctx, dtc, z, w.dave, 0, proStr("Z role")); err != nil {
			t.Fatalf("the member labels himself in Z: %v", err)
		}
		zb := e.gmpBinding(t, w.dave, z)
		if got := cards(actx, atc, w.dave); got != "z:Z role:true:1" {
			t.Errorf("N4 before the removal: cards %s (hash)", itH(got))
		}
		zScope := Scope{Kind: tenancy.ScopeGroup, Group: &z}
		if err := e.svc.RemoveMember(actx, atc, zScope, w.dave.id); err != nil {
			t.Fatalf("Z's owner removes the member: %v", err)
		}
		if got := cards(actx, atc, w.dave); got != notFound {
			t.Errorf("N4 after the removal: cards %s (hash), want not_found", itH(got))
		}
		if e.gmpBinding(t, w.dave, z) != "" || e.gmpRow(t, zb) != "" {
			t.Error("N4: the Z membership or its label survived the removal")
		}
		e.grantGroup(t, w.A, z, w.dave, tenancy.GroupMember)
		if got := cards(actx, atc, w.dave); got != "z:-:true:0" {
			t.Errorf("N4 after the readmission: cards %s (hash), want an empty Z card", itH(got))
		}

		// N5 Organization-role revocation: the owner demotes the Organization admin to member (through the facade), whose
		// next read in the same request scope is not_found.
		fctx, ftc := e.as(t, w.frank, w.A)
		if got := cards(fctx, ftc, w.dave); got == notFound {
			t.Fatal("N5: the Organization admin cannot read the subject before the demotion")
		}
		cctx, ctc := e.as(t, w.carol, w.A)
		if _, err := e.svc.ChangeMemberRole(cctx, ctc, Scope{Kind: tenancy.ScopeOrganization}, w.frank.id, string(tenancy.OrgMember)); err != nil {
			t.Fatalf("the owner demotes the Organization admin: %v", err)
		}
		if got := cards(fctx, ftc, w.dave); got != notFound {
			t.Errorf("N5 after the demotion: cards %s (hash), want not_found", itH(got))
		}
	})

	t.Run("GMP-12 a label confers no authority (W4)", func(t *testing.T) {
		bctx, btc := e.as(t, w.bob, w.A)
		dctx, dtc := e.as(t, w.dave, w.A)
		cur := current(bctx, btc, w.dave)
		if _, err := set(bctx, btc, gid, w.dave, cur.Version, proStr("Owner")); err != nil {
			t.Fatalf("label the member Owner: %v", err)
		}
		// The permission role is unchanged.
		var role string
		if err := e.setup.QueryRow(ctx, `SELECT role FROM app.role_bindings WHERE id = $1::text::uuid`, e.gmpBinding(t, w.dave, gid)).Scan(&role); err != nil ||
			role != string(tenancy.GroupMember) {
			t.Errorf("the labelled member's Group role: %s (%v), want group_member", role, err)
		}
		// Group administration and another member's label stay refused to the labelled member.
		gmpWant(t, "the labelled member removes the Group admin", e.svc.RemoveMember(dctx, dtc, research, w.grace.id), CodeForbidden, "")
		_, err := set(dctx, dtc, gid, w.bob, 0, proStr("Not mine"))
		gmpWant(t, "the labelled member labels the Group owner", err, CodeForbidden, "")
		// No database function other than the guard and the card projection reads the table, and no other table's
		// policy names it (no capability, policy or eligibility reads a label).
		var readers string
		var policies int
		if err := e.setup.QueryRow(ctx, `SELECT coalesce((SELECT string_agg(p.proname::text, ',' ORDER BY p.proname) FROM pg_proc p
  WHERE position('group_member_profiles' IN p.prosrc) > 0 AND p.proname NOT IN ('group_member_profiles_guard', 'rls_group_member_cards')), ''),
  (SELECT count(*) FROM pg_policy p WHERE p.polrelid <> 'app.group_member_profiles'::regclass
    AND position('group_member_profiles' IN coalesce(pg_get_expr(p.polqual, p.polrelid), '') || coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')) > 0)::int`).Scan(&readers, &policies); err != nil {
			t.Fatal(err)
		}
		if readers != "" || policies != 0 {
			t.Errorf("label readers outside the projection: functions %q, %d policies of other tables", readers, policies)
		}
	})
}
