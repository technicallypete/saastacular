//go:build integration

// TestAccessFootprint (G007 Plan r1 §6.7; PG-001, docs/PERFORMANCE.md): bounded characterization of the guarded
// operations' query and lock footprint, run in the B-POLICY-IT step. It is not a benchmark and sets no target.
//
//   - E-PG-1 (recorded, no pass/fail threshold): every fact statement and write of the guarded profile and Group writes,
//     captured as executed, is replayed as the runtime role under EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) in a
//     transaction that is rolled back, at the s1 and s50 characterization points and in both context modes. Each summary
//     (node types, relation and index names, actual rows, loops, rows removed, shared buffers, planning and execution
//     times) is one JSON line on stdout prefixed "E-PG-1 "; conditions, constants and parameters are dropped. A line
//     prefixed "E-PG-1-TRIGGER " records the Plan's escalation trigger (s50: an Organization-wide membership path, or a
//     sequential scan in statements 4-7).
//   - E-PG-2 (asserted): the rows each locking statement returned, through the counting hookDB, equal Plan r1 §4.7.1 for
//     every guarded write in both context modes for s1, s50 and the X admin; no other statement carries a locking
//     clause. G018 (PG-009): "E-PG-2 onboarding writes" records the create and both completion writes and the two list
//     reads (no locks; the same statement count for s1 and s50).
//   - E-PG-MEM (G019, recorded with bounded assertions): one line prefixed "E-PG-MEM " per membership, ownership and
//     invitation operation (incl. role change, demotion, leave, revoke, retirement eligibility and Group removal) —
//     statement steps, claim-function calls and locked rows by step.
//
// Dataset (owner SQL in its own Organizations X and Y, then ANALYZE as the owner; the runtime role cannot ANALYZE): X has
// 200 Groups, 400 Organization members and 4,000 Group bindings; s1 holds 1 Group binding in X and s50 holds 50, and each
// also holds 50 in Y; the X admin holds 50 Group bindings in X. The subjects' sessions are real G004 sessions. These
// points only distinguish subject-scoped from Organization-wide access; they are not representative cardinalities.
package access

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/tenancy"
)

// footprintWorld is the E-PG dataset.
type footprintWorld struct {
	X, Y                  tenancy.OrganizationID
	groupsX               []tenancy.GroupID // by id order
	s1, s50, admin, owner *itUser
	s1Group, s50Group     tenancy.GroupID // a Group of the subject linked to its target profile
	s1Profile, s50Profile tenancy.BusinessProfileID
	adminGroup            tenancy.GroupID
	adminProfile          tenancy.BusinessProfileID
	ng                    map[*itUser]int // the subject's Group bindings in X
	mOrg, bOrg            int             // Organization-scope bindings of X; all Group bindings of X
}

func (e *itEnv) footprintWorld(t *testing.T) *footprintWorld {
	t.Helper()
	ctx := context.Background()
	fw := &footprintWorld{ng: map[*itUser]int{}}
	fw.X, fw.Y = e.org(t, "Footprint X"), e.org(t, "Footprint Y")
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := e.setup.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	x, y := fw.X.String(), fw.Y.String()
	exec(`INSERT INTO app.groups (organization_id, name, slug) SELECT $1::text::uuid, 'Footprint X ' || i, 'fx-' || i FROM generate_series(1, 200) i`, x)
	exec(`INSERT INTO app.groups (organization_id, name, slug) SELECT $1::text::uuid, 'Footprint Y ' || i, 'fy-' || i FROM generate_series(1, 60) i`, y)
	// 396 filler members (with s1, s50, the admin and the owner: 400), each holding nine or ten Group bindings: 3,899 in
	// all (with the subjects' 101: 4,000).
	exec(`WITH u AS (INSERT INTO app.users (display_name) SELECT 'Footprint filler' FROM generate_series(1, 396) RETURNING id)
INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
SELECT $1::text::uuid, id, 'organization', $1::text::uuid, NULL, 'org_member', 'membership' FROM u`, x)
	exec(`WITH m AS (SELECT subject_user_id AS u, row_number() OVER (ORDER BY subject_user_id) - 1 AS i FROM app.role_bindings
           WHERE organization_id = $1::text::uuid AND scope_type = 'organization'),
     g AS (SELECT id, row_number() OVER (ORDER BY id) - 1 AS j FROM app.groups WHERE organization_id = $1::text::uuid)
INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
SELECT $1::text::uuid, m.u, 'group', g.id, g.id, 'group_member', 'membership'
FROM generate_series(0, 3898) k JOIN m ON m.i = k % 396 JOIN g ON g.j = (k / 396 + m.i) % 200`, x)
	// 100 profiles of X, profile i linked to Groups 2i and 2i+1 (id order).
	exec(`WITH p AS (INSERT INTO app.business_profiles (organization_id, display_name) SELECT $1::text::uuid, 'Footprint Profile ' || i
                 FROM generate_series(0, 99) i RETURNING id),
     pn AS (SELECT id, row_number() OVER (ORDER BY id) - 1 AS i FROM p),
     g AS (SELECT id, row_number() OVER (ORDER BY id) - 1 AS j FROM app.groups WHERE organization_id = $1::text::uuid)
INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id)
SELECT $1::text::uuid, g.id, pn.id FROM g JOIN pn ON pn.i = g.j / 2`, x)
	rows, err := e.setup.Query(ctx, `SELECT id::text FROM app.groups WHERE organization_id = $1::text::uuid ORDER BY id`, x)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil || len(ids) != 200 {
		t.Fatalf("X groups: %d %v", len(ids), err)
	}
	for _, s := range ids {
		g, _ := tenancy.ParseGroupID(s)
		fw.groupsX = append(fw.groupsX, g)
	}
	profileOf := func(g tenancy.GroupID) tenancy.BusinessProfileID {
		t.Helper()
		var s string
		if err := e.setup.QueryRow(ctx, `SELECT business_profile_id::text FROM app.group_business_profiles WHERE organization_id = $1::text::uuid
AND group_id = $2::text::uuid`, x, g.String()).Scan(&s); err != nil {
			t.Fatal(err)
		}
		p, _ := tenancy.ParseBusinessProfileID(s)
		return p
	}
	// The subjects: real G004 sessions; bindings through the accepted model.
	fw.s1, fw.s50, fw.admin, fw.owner = e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t)
	e.grantOrg(t, fw.X, fw.s1, tenancy.OrgMember)
	e.grantOrg(t, fw.X, fw.s50, tenancy.OrgMember)
	e.grantOrg(t, fw.X, fw.admin, tenancy.OrgAdmin)
	e.grantOrg(t, fw.X, fw.owner, tenancy.OrgOwner)
	e.grantGroup(t, fw.X, fw.groupsX[0], fw.s1, tenancy.GroupAdmin)
	for i := range 50 {
		e.grantGroup(t, fw.X, fw.groupsX[i], fw.s50, tenancy.GroupAdmin)
		e.grantGroup(t, fw.X, fw.groupsX[100+i], fw.admin, tenancy.GroupMember)
	}
	fw.ng[fw.s1], fw.ng[fw.s50], fw.ng[fw.admin], fw.ng[fw.owner] = 1, 50, 50, 0
	fw.s1Group, fw.s50Group, fw.adminGroup = fw.groupsX[0], fw.groupsX[49], fw.groupsX[100]
	fw.s1Profile, fw.s50Profile, fw.adminProfile = profileOf(fw.s1Group), profileOf(fw.s50Group), profileOf(fw.adminGroup)
	for _, u := range []*itUser{fw.s1, fw.s50, fw.admin, fw.owner} { // G016 A2: one prepared credential per user at X
		e.bind(t, u, fw.X)
	}
	for _, u := range []*itUser{fw.s1, fw.s50} {
		e.grantOrg(t, fw.Y, u, tenancy.OrgMember)
		exec(`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
SELECT $1::text::uuid, $2::text::uuid, 'group', id, id, 'group_member', 'membership'
FROM (SELECT id FROM app.groups WHERE organization_id = $1::text::uuid ORDER BY id LIMIT 50) g`, y, u.id.String())
	}
	exec(`ANALYZE app.organizations, app.groups, app.role_bindings, app.business_profiles, app.group_business_profiles, app.users, app.sessions`)
	if err := e.setup.QueryRow(ctx, `SELECT count(*) FILTER (WHERE scope_type = 'organization'), count(*) FILTER (WHERE scope_type = 'group')
FROM app.role_bindings WHERE organization_id = $1::text::uuid`, x).Scan(&fw.mOrg, &fw.bOrg); err != nil {
		t.Fatal(err)
	}
	if fw.mOrg != 400 || fw.bOrg != 4000 {
		t.Fatalf("dataset: %d Organization members and %d Group bindings in X, want 400 and 4000", fw.mOrg, fw.bOrg)
	}
	return fw
}

// planSummary is the privacy-filtered summary of one plan node: no condition, filter, output or parameter.
type planSummary struct {
	Node           string        `json:"node"`
	Relation       string        `json:"relation,omitempty"`
	Index          string        `json:"index,omitempty"`
	ActualRows     float64       `json:"actual_rows"`
	Loops          float64       `json:"loops"`
	RemovedFilter  float64       `json:"removed_by_filter,omitempty"`
	RemovedRecheck float64       `json:"removed_by_recheck,omitempty"`
	SharedHit      float64       `json:"shared_hit"`
	SharedRead     float64       `json:"shared_read"`
	Plans          []planSummary `json:"plans,omitempty"`
}

func summarize(n map[string]any) planSummary {
	num := func(k string) float64 { v, _ := n[k].(float64); return v }
	str := func(k string) string { v, _ := n[k].(string); return v }
	s := planSummary{Node: str("Node Type"), Relation: str("Relation Name"), Index: str("Index Name"), ActualRows: num("Actual Rows"),
		Loops: num("Actual Loops"), RemovedFilter: num("Rows Removed by Filter"), RemovedRecheck: num("Rows Removed by Index Recheck"),
		SharedHit: num("Shared Hit Blocks"), SharedRead: num("Shared Read Blocks")}
	if kids, ok := n["Plans"].([]any); ok {
		for _, k := range kids {
			if m, ok := k.(map[string]any); ok {
				s.Plans = append(s.Plans, summarize(m))
			}
		}
	}
	return s
}

// examined is the number of role_bindings index or heap entries the plan's scan nodes examined, and whether any node
// is a sequential scan.
func examined(s planSummary) (n float64, seq bool) {
	if s.Node == "Seq Scan" {
		seq = true
	}
	switch s.Node {
	case "Index Scan", "Index Only Scan", "Bitmap Index Scan", "Seq Scan":
		if s.Relation == "role_bindings" || strings.HasPrefix(s.Index, "role_bindings_") {
			n += (s.ActualRows + s.RemovedFilter + s.RemovedRecheck) * max(s.Loops, 1)
		}
	}
	for _, k := range s.Plans {
		kn, ks := examined(k)
		n, seq = n+kn, seq || ks
	}
	return n, seq
}

func TestAccessFootprint(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	fw := e.footprintWorld(t)
	var serverVersion string
	if err := e.pool.QueryRow(ctx, `SELECT current_setting('server_version')`).Scan(&serverVersion); err != nil {
		t.Fatal(err)
	}
	// Environment label; the PostgreSQL image reference is in the batch's images.txt.
	env := map[string]any{"server_version": serverVersion, "goos": runtime.GOOS, "goarch": runtime.GOARCH, "cpus": runtime.NumCPU(),
		"runner": "harness check container"}

	type scenario struct {
		subject string
		u       *itUser
		group   tenancy.GroupID
		profile tenancy.BusinessProfileID
	}
	scenarios := []scenario{{"s1", fw.s1, fw.s1Group, fw.s1Profile}, {"s50", fw.s50, fw.s50Group, fw.s50Profile},
		{"x_admin", fw.admin, fw.adminGroup, fw.adminProfile}}

	t.Run("E-PG-1 plan summaries", func(t *testing.T) {
		svc, hook := e.service(e.id)
		for _, sc := range scenarios[:2] {
			rctx, _ := beginAs(t, svc, ctx, sc.u, fw.X)
			p, err := e.model.GetBusinessProfile(ctx, fw.X, sc.profile)
			if err != nil {
				t.Fatal(err)
			}
			g, err := e.model.GetGroup(ctx, fw.X, sc.group)
			if err != nil {
				t.Fatal(err)
			}
			for _, mode := range []string{"org-only", "selected"} {
				sel := tenancy.GroupID{}
				if mode == "selected" {
					sel = sc.group
				}
				tc := resolve(t, svc, rctx, fw.X, sel)
				ops := map[string]func() error{
					opUpdateBusinessProfile: func() error { _, err := svc.UpdateBusinessProfile(rctx, tc, sc.profile, p.Context); return err },
					opRenameGroup:           func() error { _, err := svc.RenameGroup(rctx, tc, sc.group, g.Name); return err },
				}
				for op, call := range ops {
					hook.reset()
					hook.capture = true
					if err := call(); err != nil {
						t.Fatalf("%s %s %s: %v", sc.subject, mode, op, err)
					}
					stmts := hook.capturedStmts()
					if len(stmts) == 0 || stmts[0].step != "C" {
						t.Fatalf("%s %s %s: the context statement is not the first statement", sc.subject, mode, op)
					}
					for _, st := range stmts[1:] {
						e.explain(t, env, fw, sc.subject, mode, op, st, stmts[0])
					}
				}
			}
		}
		hook.reset()
	})

	t.Run("E-PG-2 locked-row inventory", func(t *testing.T) {
		svc, hook := e.service(e.id)
		for _, sc := range scenarios {
			rctx, _ := beginAs(t, svc, ctx, sc.u, fw.X)
			p, _ := e.model.GetBusinessProfile(ctx, fw.X, sc.profile)
			g, _ := e.model.GetGroup(ctx, fw.X, sc.group)
			for _, mode := range []string{"org-only", "selected"} {
				sel := tenancy.GroupID{}
				if mode == "selected" {
					sel = sc.group
				}
				tc := resolve(t, svc, rctx, fw.X, sel)
				type write struct {
					op   string
					call func() error
					want map[string]int // locked rows by step
				}
				gbSel := 0
				if mode == "selected" {
					gbSel = 1 // every scenario subject holds a binding in its scenario Group
				}
				ng := fw.ng[sc.u]
				writes := []write{
					{opUpdateBusinessProfile, func() error { _, err := svc.UpdateBusinessProfile(rctx, tc, sc.profile, p.Context); return err },
						map[string]int{"U": 1, "S": 1, "O": 1, "OB": 1, "GB": map[string]int{"org-only": ng, "selected": gbSel}[mode],
							"G": map[string]int{"org-only": ng, "selected": 1}[mode]}},
					{opRenameGroup, func() error { _, err := svc.RenameGroup(rctx, tc, sc.group, g.Name); return err },
						map[string]int{"U": 1, "S": 1, "O": 1, "OB": 1, "GB": 1, "Gt": 1}},
					{opChangeGroupSlug, func() error { _, err := svc.ChangeGroupSlug(rctx, tc, sc.group, g.Slug); return err },
						map[string]int{"U": 1, "S": 1, "O": 1, "OB": 1, "GB": 1, "Gt": 1}},
					{opSetOrganizationPolicy, func() error {
						_, err := svc.SetOrganizationPolicy(rctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: true})
						return err
					}, map[string]int{"U": 1, "S": 1, "O": 1, "OB": 1, "GB": gbSel, "G": map[string]int{"org-only": 0, "selected": 1}[mode]}},
				}
				for _, wr := range writes {
					hook.reset()
					err := wr.call()
					got := map[string]int{}
					total, want := 0, 0
					for _, l := range hook.lockedRows() {
						if l.step == "?" || l.step == "A" || l.step == "P" || l.step == "W" {
							t.Errorf("%s %s %s: an unplanned statement (%s) carries a locking clause", sc.subject, mode, wr.op, l.step)
						}
						if l.step == "O" && (wr.op == opSetOrganizationPolicy) != (l.mode == "NO KEY UPDATE") {
							t.Errorf("%s %s %s: Organization row lock mode %s", sc.subject, mode, wr.op, l.mode)
						}
						if l.step == "Gt" && l.mode != "NO KEY UPDATE" {
							t.Errorf("%s %s %s: target Group lock mode %s", sc.subject, mode, wr.op, l.mode)
						}
						got[l.step] += l.rows
						total += l.rows
					}
					for k, v := range wr.want {
						if v != 0 && got[k] != v {
							t.Errorf("%s %s %s: %s locked %d rows, want %d", sc.subject, mode, wr.op, k, got[k], v)
						}
						want += v
					}
					if total != want {
						t.Errorf("%s %s %s: %d rows locked, want %d (%v)", sc.subject, mode, wr.op, total, want, got)
					}
					if wr.op == opUpdateBusinessProfile && mode == "org-only" && total != 4+2*ng {
						t.Errorf("%s: Organization-only profile write locked %d rows, want 4 + 2*%d", sc.subject, total, ng)
					}
					line, _ := json.Marshal(map[string]any{"evidence": "E-PG-2", "subject": sc.subject, "mode": mode, "operation": wr.op,
						"outcome": map[bool]string{true: "allow", false: string(CodeOf(err))}[err == nil], "locked_rows": got, "total": total})
					fmt.Println("E-PG-2 " + string(line))
				}
			}
		}
		hook.reset()
	})

	// G018 (Plan r5 §1.5, PG-009): one E-PG observation per onboarding write in Organization-only context (the only
	// context CreateGroup admits) for s1, s50 and the X admin — the statement steps in order and the locked rows by
	// step. Asserted: each write succeeds; it issues its keyed inserts exactly once (create: IG; new-profile
	// completion: IP and IA; existing-profile completion: IA); no unrecognized statement and no insert carries a locking
	// clause; and s1 and s50 issue the same number of statements for every write, so a subject's Group-binding count
	// changes locked rows but never the round trips. It runs last, so the Groups it creates touch no other row.
	t.Run("E-PG-2 onboarding writes (G018)", func(t *testing.T) {
		svc, hook := e.service(e.id)
		// G060: the X admin attaches its existing profile through the Organization path, which needs the Group-management
		// override (explicit setup through the fixture service; turned off when this subtest ends). s1 and s50 are
		// Organization members acting through their own Group roles, so the setting does not apply to them.
		e.withOverride(t, fw.owner, fw.X)
		inserts := map[string]map[string]int{
			opCreateGroup:                {"IG": 1, "IP": 0, "IA": 0},
			opCompleteOnboardingNew:      {"IG": 0, "IP": 1, "IA": 1},
			opCompleteOnboardingExisting: {"IG": 0, "IP": 0, "IA": 1},
		}
		stmtCount := map[string]map[string]int{}
		for _, sc := range scenarios {
			rctx, _ := beginAs(t, svc, ctx, sc.u, fw.X)
			tc := resolve(t, svc, rctx, fw.X, tenancy.GroupID{})
			var gNew, gExisting tenancy.GroupID
			run := func(op string, call func() error) {
				t.Helper()
				hook.reset()
				if err := call(); err != nil {
					t.Fatalf("%s %s: %v", sc.subject, op, err)
				}
				_, steps := hook.counts()
				n := map[string]int{}
				for _, s := range steps {
					n[s]++
				}
				for k, v := range inserts[op] {
					if n[k] != v {
						t.Errorf("%s %s: %d %s statements, want %d", sc.subject, op, n[k], k, v)
					}
				}
				locked := map[string]int{}
				for _, l := range hook.lockedRows() {
					if l.step == "?" || l.step == "IG" || l.step == "IP" || l.step == "IA" || l.step == "W" || op == opListGroups || op == opListBusinessProfiles {
						t.Errorf("%s %s: an unplanned statement (%s) carries a locking clause", sc.subject, op, l.step)
					}
					locked[l.step] += l.rows
				}
				if stmtCount[op] == nil {
					stmtCount[op] = map[string]int{}
				}
				stmtCount[op][sc.subject] = len(steps)
				line, _ := json.Marshal(map[string]any{"evidence": "E-PG-2", "subject": sc.subject, "mode": "org-only", "operation": op,
					"statements": len(steps), "steps": steps, "locked_rows": locked})
				fmt.Println("E-PG-2 " + string(line))
			}
			for i, target := range []*tenancy.GroupID{&gNew, &gExisting} {
				run(opCreateGroup, func() error {
					out, err := svc.CreateGroup(rctx, tc, fmt.Sprintf("Footprint onboarding %s %d", sc.subject, i), fmt.Sprintf("fo-%s-%d", strings.ReplaceAll(sc.subject, "_", "-"), i))
					*target = out.Group.ID
					return err
				})
			}
			run(opCompleteOnboardingNew, func() error {
				_, err := svc.CompleteGroupOnboardingWithNewProfile(rctx, tc, gNew, tenancy.BusinessContext{DisplayName: "Footprint onboarding " + sc.subject})
				return err
			})
			run(opCompleteOnboardingExisting, func() error {
				_, err := svc.CompleteGroupOnboardingWithProfile(rctx, tc, gExisting, sc.profile)
				return err
			})
			// §1.5: ListGroups = facts + 1 page + 1 associations statement; ListBusinessProfiles = facts + 1 page statement.
			run(opListGroups, func() error {
				_, err := svc.ListGroups(rctx, tc, "", 100, "")
				return err
			})
			run(opListBusinessProfiles, func() error {
				_, err := svc.ListBusinessProfiles(rctx, tc, "", 100)
				return err
			})
		}
		for op, by := range stmtCount {
			d := by["s1"] - by["s50"]
			if op == opCompleteOnboardingExisting && (d == 1 || d == -1) {
				continue // §1.5: one extra lock statement when the candidate Groups straddle the target (id order)
			}
			if d != 0 {
				t.Errorf("%s: s1 issued %d statements and s50 %d; the round trips depend on the subject's Group bindings", op, by["s1"], by["s50"])
			}
		}
		hook.reset()
	})

	// G019 (Plan r1 §1.5, E-PG-MEM): one E-PG observation per membership, ownership and invitation operation in
	// Organization-only context — the member and invitation lists, issue (Organization and Group), get, candidates,
	// promotion, removal of a member holding three dependent Group bindings, an ordinary role change, an owner demotion, an
	// invitation revocation, a Group member's removal, an Organization departure, the retirement eligibility read (each on
	// its own prepared subject), the recipient's own list and Group accept,
	// the Organization consume (owner-role pool, D-15) and, last, the ownership transfer — recording the statement steps
	// in order (an unrecognized statement as "?" with its normalized prefix), the claim-function calls and the locked rows
	// by step. Asserted: each operation succeeds; each consumer issues exactly one claim-function call and no other
	// operation issues one; no list or get statement carries a locking clause. It is an observation, not a benchmark. It
	// runs last with its own users, so the E-PG-1/E-PG-2 dataset and subjects are unchanged for them.
	t.Run("E-PG-MEM membership operations (G019)", func(t *testing.T) {
		svc, hook := e.service(e.id)
		osvc, ohook := e.memOwnerHooked(t)
		orgScope := Scope{Kind: tenancy.ScopeOrganization}
		g0 := fw.groupsX[0]
		mo, adm1, adm2, dep, acc := e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t)
		e.grantOrg(t, fw.X, mo, tenancy.OrgOwner)
		e.grantOrg(t, fw.X, adm1, tenancy.OrgAdmin)
		e.grantOrg(t, fw.X, adm2, tenancy.OrgAdmin)
		e.grantOrg(t, fw.X, dep, tenancy.OrgMember)
		e.grantOrg(t, fw.X, acc, tenancy.OrgMember)
		for _, g := range fw.groupsX[:3] {
			e.grantGroup(t, fw.X, g, dep, tenancy.GroupMember)
		}
		// G016 A2: the two acting subjects need a prepared credential at X before beginAs (u.in).
		e.bind(t, mo, fw.X)
		e.bind(t, acc, fw.X)
		// G060: the owner mo holds no Group role, so its Group-scoped operations on g0 (such as the Group invitation) are
		// Organization-path Group management that needs the override. It is set
		// through fw.owner, which stays an owner of X (mo transfers its ownership last), and is turned off when this
		// subtest ends.
		e.withOverride(t, fw.owner, fw.X)
		observe := func(h *hookDB, subject, op string, call func() error) {
			t.Helper()
			h.reset()
			h.capture = true
			if err := call(); err != nil {
				t.Fatalf("%s %s: %v", subject, op, err)
			}
			// Reads carry no locking clause: the lists, the invitation read and the retirement eligibility read.
			read := strings.HasPrefix(op, "list") || op == "invitation.get" || op == "retirement.eligibility"
			steps, claims := []string{}, 0
			for _, st := range h.capturedStmts() {
				s := st.step
				if s == "?" {
					s = "?" + memStmtPrefix(st.sql)
				}
				if strings.Contains(st.sql, "app.invitation_claim_") {
					claims++
				}
				steps = append(steps, s)
			}
			consumer := op == "invitation.accept" || op == "invitation.consume"
			if (consumer && claims != 1) || (!consumer && claims != 0) {
				t.Errorf("%s %s: %d claim-function calls", subject, op, claims)
			}
			locked := map[string]int{}
			for _, l := range h.lockedRows() {
				if read {
					t.Errorf("%s %s: a read statement (%s) carries a locking clause", subject, op, l.step)
				}
				locked[l.step] += l.rows
			}
			line, _ := json.Marshal(map[string]any{"evidence": "E-PG-MEM", "subject": subject, "mode": "org-only", "operation": op,
				"statements": len(steps), "steps": steps, "claim_calls": claims, "locked_rows": locked})
			fmt.Println("E-PG-MEM " + string(line))
			h.reset()
		}

		rctx, _ := beginAs(t, svc, ctx, mo, fw.X)
		tc := resolve(t, svc, rctx, fw.X, tenancy.GroupID{})
		observe(hook, "owner", "list.members", func() error { _, err := svc.ListMembers(rctx, tc, orgScope, "", 100, ""); return err })
		addr := memAddr("footprint")
		var oiss, giss IssuedInvitation
		observe(hook, "owner", "invitation.issue", func() error {
			var err error
			oiss, err = svc.IssueInvitation(rctx, tc, orgScope, string(tenancy.OrgMember), Recipient{Address: &addr}, 0)
			return err
		})
		observe(hook, "owner", "invitation.issue_group", func() error {
			var err error
			giss, err = svc.IssueInvitation(rctx, tc, Scope{Kind: tenancy.ScopeGroup, Group: &g0}, string(tenancy.GroupMember), Recipient{UserID: &acc.id}, 0)
			return err
		})
		observe(hook, "owner", "list.invitations", func() error { _, err := svc.ListInvitations(rctx, tc, orgScope, "", 100); return err })
		observe(hook, "owner", "invitation.get", func() error { _, err := svc.GetInvitation(rctx, tc, oiss.Invitation.ID); return err })
		observe(hook, "owner", "list.candidates", func() error { _, err := svc.ListInvitationCandidates(rctx, tc, g0, "", 100); return err })
		observe(hook, "owner", "membership.promote", func() error { _, err := svc.PromoteToOwner(rctx, tc, orgScope, adm1.id); return err })
		observe(hook, "owner", "membership.remove_with_dependents", func() error { return svc.RemoveMember(rctx, tc, orgScope, dep.id) })

		// The remaining per-operation observations, each on its own prepared subject (fixture grants before observing;
		// the acting leaver and retiree have credentials prepared at X, G016 A2): an ordinary role change, an owner
		// demotion, an invitation revocation, a Group member's removal, an Organization departure and the retirement
		// eligibility read. Recorded only; the bounded claim and read-lock assertions of observe apply unchanged.
		chg, dem, gm, lv, re := e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t)
		e.grantOrg(t, fw.X, chg, tenancy.OrgMember)
		e.grantOrg(t, fw.X, dem, tenancy.OrgOwner)
		e.grantOrg(t, fw.X, gm, tenancy.OrgMember)
		e.grantGroup(t, fw.X, g0, gm, tenancy.GroupMember)
		e.grantOrg(t, fw.X, lv, tenancy.OrgMember)
		e.grantOrg(t, fw.X, re, tenancy.OrgMember)
		e.bind(t, lv, fw.X)
		e.bind(t, re, fw.X)
		observe(hook, "owner", "membership.change_role", func() error {
			_, err := svc.ChangeMemberRole(rctx, tc, orgScope, chg.id, string(tenancy.OrgAdmin))
			return err
		})
		observe(hook, "owner", "ownership.demote", func() error {
			_, err := svc.DemoteOwner(rctx, tc, orgScope, dem.id, string(tenancy.OrgAdmin))
			return err
		})
		raddr := memAddr("footprint-revoke")
		riss, err := svc.IssueInvitation(rctx, tc, orgScope, string(tenancy.OrgMember), Recipient{Address: &raddr}, 0)
		if err != nil {
			t.Fatalf("issue the invitation to revoke: %v", err)
		}
		observe(hook, "owner", "invitation.revoke", func() error { return svc.RevokeInvitation(rctx, tc, riss.Invitation.ID) })
		observe(hook, "owner", "membership.remove_group", func() error {
			return svc.RemoveMember(rctx, tc, Scope{Kind: tenancy.ScopeGroup, Group: &g0}, gm.id)
		})
		lctx, _ := beginAs(t, svc, ctx, lv, fw.X)
		ltc := resolve(t, svc, lctx, fw.X, tenancy.GroupID{})
		observe(hook, "leaver", "membership.leave", func() error { return svc.LeaveScope(lctx, ltc, orgScope) })
		observe(hook, "retiree", "retirement.eligibility", func() error {
			_, err := svc.RetirementEligibility(ctx, re.in(t, fw.X), itOriginOf(fw.X))
			return err
		})

		actx, _ := beginAs(t, svc, ctx, acc, fw.X)
		atc := resolve(t, svc, actx, fw.X, tenancy.GroupID{})
		observe(hook, "recipient", "list.mine", func() error { _, err := svc.ListMyInvitations(actx, atc, "", 100); return err })
		observe(hook, "recipient", "invitation.accept", func() error {
			_, err := svc.AcceptGroupInvitation(actx, atc, giss.Invitation.ID, giss.Token)
			return err
		})
		sub, _ := e.memPrincipal(t, "FP")
		ev := e.memAttest(t, sub, addr)
		observe(ohook, "first-time", "invitation.consume", func() error {
			_, err := osvc.ConsumeOrganizationInvitation(ctx, ev, oiss.Invitation.ID, oiss.Token)
			return err
		})
		observe(hook, "owner", "ownership.transfer", func() error {
			return svc.TransferOwnership(rctx, tc, orgScope, adm2.id, string(tenancy.OrgAdmin))
		})
	})
}

// memStmtPrefix is the first 64 bytes of a statement with its whitespace collapsed (an E-PG-MEM step label for a
// statement the authorization-contract patterns do not name); it carries SQL text only, never arguments.
func memStmtPrefix(sql string) string {
	s := strings.Join(strings.Fields(sql), " ")
	if len(s) > 64 {
		s = s[:64]
	}
	return s
}

// explain replays one captured statement under EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) as the runtime role in a
// read-write READ COMMITTED transaction that is rolled back, prints its summary and evaluates the escalation trigger.
// From version 6 the operation's captured context statement runs first in that transaction, so the replay sees the
// same row-level security context as the operation.
func (e *itEnv) explain(t *testing.T, env map[string]any, fw *footprintWorld, subject, mode, op string, st, ctxStmt capturedStmt) {
	t.Helper()
	ctx := context.Background()
	tx, err := e.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, ctxStmt.sql, ctxStmt.args...); err != nil {
		t.Fatalf("explain %s: context: %v", st.step, err)
	}
	var raw []byte
	if err := tx.QueryRow(ctx, "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) "+st.sql, st.args...).Scan(&raw); err != nil {
		t.Fatalf("explain %s: %v", st.step, err)
	}
	var doc []map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil || len(doc) != 1 {
		t.Fatalf("explain %s: %v", st.step, err)
	}
	plan, _ := doc[0]["Plan"].(map[string]any)
	sum := summarize(plan)
	n, seq := examined(sum)
	out := map[string]any{"evidence": "E-PG-1", "subject": subject, "mode": mode, "operation": op, "step": st.step,
		"cardinality": map[string]int{"m_org": fw.mOrg, "b_org": fw.bOrg, "groups_x": len(fw.groupsX), "n_g": map[string]int{"s1": 1, "s50": 50}[subject],
			"N_g": map[string]int{"s1": 51, "s50": 100}[subject]},
		"locking": lockingClause.MatchString(st.sql), "plan": sum, "role_bindings_examined": n, "seq_scan": seq,
		"planning_ms": doc[0]["Planning Time"], "execution_ms": doc[0]["Execution Time"], "environment": env}
	line, _ := json.Marshal(out)
	fmt.Println("E-PG-1 " + string(line))
	if subject != "s50" {
		return
	}
	var why []string
	switch st.step {
	case "OB":
		if n >= float64(fw.mOrg) {
			why = append(why, "statement 4 examines at least M_org entries")
		}
	case "GB":
		if n >= float64(fw.bOrg) {
			why = append(why, "statement 5 examines at least B_org entries")
		}
	}
	if seq && (st.step == "OB" || st.step == "GB" || st.step == "Gt" || st.step == "G" || st.step == "A") {
		why = append(why, "a sequential scan in statements 4-7")
	}
	if len(why) > 0 {
		// Codex review 432d5f17 §2 accepted the sequential scans of statements 6b (groups) and 7 (associations) on the small
		// fixture tables as a recorded PG-001 observation; anything else is new and returns to Codex.
		known := (st.step == "G" || st.step == "A") && len(why) == 1 && why[0] == "a sequential scan in statements 4-7"
		line, _ := json.Marshal(map[string]any{"evidence": "E-PG-1-TRIGGER", "subject": subject, "mode": mode, "operation": op, "step": st.step,
			"reasons": why, "known": known})
		fmt.Println("E-PG-1-TRIGGER " + string(line))
		t.Logf("E-PG-1 escalation trigger (recorded, Plan r1 §6.7): %s %s %s: %v", mode, op, st.step, why)
	}
}
