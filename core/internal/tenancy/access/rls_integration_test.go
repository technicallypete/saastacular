//go:build integration

// B-RLS-IT (G008 Plan r3 §8.3): TestRLSIsolation (RLS-ATTEST, RLS-CTX, RLS-ORG, RLS-SCOPE against the independent
// oracle of rls_expected_test.go, RLS-AUTH with the E5 raw-statement controls (a)-(e), RLS-IMMUT, RLS-PROJ), TestRLSPool
// (P1-P10 through the facade on a one-connection pool) and TestRLSFootprint (E-RLS-1..3), against a disposable
// PostgreSQL migrated to version 6 (setupIT). Identity's own pool cases (I1-I5), E5 (f) and I-MAP-CONC are in
// core/internal/identity/rls_pool_integration_test.go.
//
// Every isolation assertion runs on a direct, attested runtime login (rlsRuntime); fixtures and setup-side hashes use
// the attested superuser (e.setup), which an assertion must never use (RLS-ATTEST). Raw-SQL cases set the four
// transaction-local settings themselves (the trust limit: they keep the actor fixed and substitute tenant/scope/row
// IDs). Failure messages print labels, codes and hashes only.
//
// The oracle glue: rlsExpectedFixture is created exactly (rlsOracleWorld); each rlsCase runs in its own rolled-back
// runtime transaction. Visible sets are compared exactly, with every non-fixture row counted as an extra. "ok" requires
// exactly one updated row whose new value is read back; "no_update" requires no change through the Plan-owned mechanism
// of that row (rlsMechanism: privilege 42501 for a column runtime cannot update, WITH CHECK 42501 for a row in the
// policy's existing-row set, zero rows otherwise); any other error, a connection failure or a missing fixture fails.
// A disagreement between the oracle and the database is reported as ORACLE-DISAGREEMENT and never resolved here.
// Oracle r3 (RLS-SCOPE-INSERT, G018) and r4 (RLS-SCOPE-MUTATE, G019: role_bindings UPDATE/DELETE, readback, the
// named owner-guard protections through rlsStmtC/rlsGuardState and the policy mechanism through r4Mechanism) follow
// the same glue. G020 (oracle G020): the label rows are fixture rows, every RLS-SCOPE case also compares the two label
// relations, RLS-ORG adds frank's fixed-actor label counts and joins, and RLS-LABEL (last) runs the label reads, the raw
// label writes and the Plan §5.2 T3 isolated foreign-key cases, each of the latter in its own committed runtime
// transaction under a savepoint followed by superuser zero-row/zero-audit counts (labelCommit).
package access

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/tenancy"
)

// --- connections and context ---------------------------------------------------------------------------------------------

// rlsRuntime returns a direct runtime login, attested before any assertion (R2).
func (e *itEnv) rlsRuntime(t *testing.T) *pgx.Conn {
	t.Helper()
	c := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
	if err := e.rlsAttest(c); err != nil {
		t.Fatal(err)
	}
	return c
}

// rlsAttest refuses any connection that is not the direct runtime login with no privileged attribute or membership.
func (e *itEnv) rlsAttest(c interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}) error {
	var got string
	err := c.QueryRow(context.Background(), `
SELECT format('%s|%s|%s|%s|%s|%s|%s', session_user, current_user, r.rolsuper, r.rolbypassrls, r.rolcreaterole, r.rolcreatedb,
  (SELECT count(*) FROM unnest($1::text[]) x WHERE pg_has_role(r.oid, x, 'MEMBER')))
FROM pg_roles r WHERE r.rolname = current_user`, []string{e.cfg.OwnerRole(), e.cfg.MigratorRole(), e.cfg.Database + "_retention",
		e.cfg.Database + "_policy_reader", e.cfg.Database + "_auth_reader"}).Scan(&got)
	rt := e.cfg.RuntimeRole()
	if err != nil || got != rt+"|"+rt+"|f|f|f|f|0" {
		return fmt.Errorf("not an attested runtime connection: %q (%v)", got, err)
	}
	return nil
}

// rlsSet sets the context settings in tx; a nil pointer leaves that setting uninitialized.
func rlsSet(ctx context.Context, tx pgx.Tx, user, org, group, binding *string) error {
	for _, s := range []struct {
		name string
		v    *string
	}{{"app.user_id", user}, {"app.organization_id", org}, {"app.group_id", group}, {"app.login_binding", binding}} {
		if s.v == nil {
			continue
		}
		if _, err := tx.Exec(ctx, `SELECT set_config($1, $2, true)`, s.name, *s.v); err != nil {
			return err
		}
	}
	return nil
}

func ptr(s string) *string { return &s }

const zeroUUID = "00000000-0000-0000-0000-000000000000"

// rlsTx runs fn in a runtime transaction with the given settings and always rolls it back.
func rlsTx(t *testing.T, c *pgx.Conn, user, org, group, binding *string, fn func(tx pgx.Tx)) {
	t.Helper()
	ctx := context.Background()
	tx, err := c.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := rlsSet(ctx, tx, user, org, group, binding); err != nil {
		t.Fatal(err)
	}
	fn(tx)
}

// rlsStmt runs one statement in a savepoint of tx and returns its SQLSTATE ("" on success) and affected/returned rows.
// A 42P17 (infinite recursion in policy) anywhere fails the test (AC7).
func rlsStmt(t *testing.T, tx pgx.Tx, sql string, args ...any) (string, int64) {
	t.Helper()
	st, _, n := rlsStmtC(t, tx, sql, args...)
	return st, n
}

// rlsStmtC is rlsStmt with the server's constraint name of a failure (r4: the owner guard names its protection; "" for
// a policy refusal or success).
func rlsStmtC(t *testing.T, tx pgx.Tx, sql string, args ...any) (string, string, int64) {
	t.Helper()
	ctx := context.Background()
	sp, err := tx.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sp.Rollback(ctx) }()
	var n int64
	if strings.HasPrefix(strings.TrimSpace(sql), "SELECT") {
		rows, err := sp.Query(ctx, sql, args...)
		if err == nil {
			for rows.Next() {
				n++
			}
			rows.Close()
			err = rows.Err()
		}
		if s := pgState(err); s != "" || err != nil {
			if s == "42P17" {
				t.Errorf("42P17 policy recursion: %s", sql)
			}
			return stateOrUnknown(err), pgConstraint(err), n
		}
		return "", "", n
	}
	tag, err := sp.Exec(ctx, sql, args...)
	if err != nil {
		if pgState(err) == "42P17" {
			t.Errorf("42P17 policy recursion: %s", sql)
		}
		return stateOrUnknown(err), pgConstraint(err), 0
	}
	return "", "", tag.RowsAffected()
}

// pgConstraint returns the constraint name a server error carries ("" otherwise).
func pgConstraint(err error) string {
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return pg.ConstraintName
	}
	return ""
}

// rlsGuardState is the Plan-owned SQLSTATE class of each named owner-guard protection (00015, Plan §4.4): an authority
// or context refusal is a privilege error, a cardinality or eligibility refusal an integrity error. The oracle names
// only the protection; the driver requires that name together with its class.
var rlsGuardState = map[string]string{"owner_authority": "42501", "context_required": "42501",
	"recipient_ineligible": "23000", "last_owner": "23000", "dependent_bindings": "23000",
	// G060 (00017 guard step (3b)): an own Group role change without that Group's owner role is a privilege refusal.
	"self_escalation": "42501",
	// G060 amendment01 (guard step (3) cascade): an administrative departure of a Group-owner subject.
	"owner_subject": "42501"}

func pgState(err error) string {
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return pg.Code
	}
	return ""
}

// stateOrUnknown returns the SQLSTATE, or "error" for a non-server failure (never an accepted outcome).
func stateOrUnknown(err error) string {
	if s := pgState(err); s != "" {
		return s
	}
	if err != nil {
		return "error"
	}
	return ""
}

// --- the oracle world ----------------------------------------------------------------------------------------------------

type rlsOracleWorld struct {
	id      map[string]string    // label -> UUID (orgs, groups, users, profiles, associations)
	binding map[string]string    // binding label -> "subject|scope_id"
	byKey   map[string]string    // "subject|scope_id" -> binding label
	assoc   map[string][2]string // association label -> group, profile UUID
	labels  map[string]string    // UUID -> label
	fx      rlsFixture
}

func rlsUUID(t *testing.T) string {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	b[6], b[8] = b[6]&0x0f|0x40, b[8]&0x3f|0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// oracleWorld creates exactly rlsExpectedFixture's rows (setup connection), with fresh UUIDs.
func (e *itEnv) oracleWorld(t *testing.T) *rlsOracleWorld {
	t.Helper()
	ctx := context.Background()
	w := &rlsOracleWorld{id: map[string]string{}, binding: map[string]string{}, byKey: map[string]string{}, assoc: map[string][2]string{},
		labels: map[string]string{}, fx: rlsExpectedFixture()}
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := e.setup.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("oracle fixture: %v", err)
		}
	}
	tag := strings.ReplaceAll(rlsUUID(t)[:8], "-", "")
	for _, o := range w.fx.Orgs {
		w.id[o.Label] = rlsUUID(t)
		exec(`INSERT INTO app.organizations (id, name, group_admins_edit_business_profiles) VALUES ($1, $2, $3)`, w.id[o.Label], "Oracle "+o.Label+" "+tag, o.GroupAdminsEditProfiles)
	}
	for _, g := range w.fx.Groups {
		w.id[g.Label] = rlsUUID(t)
		exec(`INSERT INTO app.groups (id, organization_id, name, slug) VALUES ($1, $2, $3, $4)`, w.id[g.Label], w.id[g.Org], "Group "+g.Label,
			"oracle-"+strings.ToLower(g.Label)+"-"+tag)
	}
	for _, u := range w.fx.Users {
		w.id[u.Label] = rlsUUID(t)
		if u.Status == "disabled" {
			exec(`INSERT INTO app.users (id, display_name, status, disabled_at) VALUES ($1, $2, 'disabled', now())`, w.id[u.Label], "Oracle "+u.Label)
		} else {
			exec(`INSERT INTO app.users (id, display_name) VALUES ($1, $2)`, w.id[u.Label], "Oracle "+u.Label)
		}
	}
	for _, b := range w.fx.Bindings {
		scope, group := w.id[b.Org], any(nil)
		if b.ScopeType == "group" {
			scope, group = w.id[b.Group], w.id[b.Group]
		}
		exec(`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1, $2, $3, $4, $5, $6, $7)`, w.id[b.Org], w.id[b.User], b.ScopeType, scope, group, b.Role, b.Source)
		k := w.id[b.User] + "|" + scope
		w.binding[b.Label], w.byKey[k] = k, b.Label
	}
	for _, p := range w.fx.Profiles {
		w.id[p.Label] = rlsUUID(t)
		exec(`INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES ($1, $2, $3)`, w.id[p.Label], w.id[p.Org], "Profile "+p.Label)
	}
	for _, a := range w.fx.Associations {
		w.id[a.Label] = rlsUUID(t)
		exec(`INSERT INTO app.group_business_profiles (id, organization_id, group_id, business_profile_id) VALUES ($1, $2, $3, $4)`,
			w.id[a.Label], w.id[a.Org], w.id[a.Group], w.id[a.Profile])
		w.assoc[a.Label] = [2]string{w.id[a.Group], w.id[a.Profile]}
	}
	// G020: label definitions (Organization-owned: owner facts = the Organization, no group_id; Group-owned: the Group)
	// and their production-shape assignments (Organization owner facts, target_type 'group', target_id = target Group).
	defOrg := map[string]string{}
	for _, d := range w.fx.LabelDefinitions {
		w.id[d.Label], defOrg[d.Label] = rlsUUID(t), d.Org
		typ, owner, group := "organization", w.id[d.Org], any(nil)
		if d.Group != "" {
			typ, owner, group = "group", w.id[d.Group], w.id[d.Group]
		}
		exec(`INSERT INTO app.label_definitions (id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility)
VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7)`, w.id[d.Label], w.id[d.Org], typ, owner, group, "Oracle label "+d.Label+" "+tag, d.Visibility)
	}
	for _, a := range w.fx.LabelAssignments {
		w.id[a.Label] = rlsUUID(t)
		exec(`INSERT INTO app.label_assignments (id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id, target_group_id)
VALUES ($1::uuid, $2::uuid, $3::uuid, 'organization', $2::uuid, 'group', $4::uuid, $4::uuid)`, w.id[a.Label], w.id[defOrg[a.Def]], w.id[a.Def], w.id[a.Group])
	}
	for l, id := range w.id {
		w.labels[id] = l
	}
	return w
}

// visible reads the nine relations (G020: and the two label relations) in tx and maps fixture rows to labels; extras
// counts rows that are not fixture rows.
func (w *rlsOracleWorld) visible(t *testing.T, tx pgx.Tx) (rlsVisible, map[string]int, error) {
	t.Helper()
	ctx := context.Background()
	var v rlsVisible
	extras := map[string]int{}
	q := func(rel, sql string, fn func(cols []string) (string, bool)) ([]string, error) {
		sp, err := tx.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer func() { _ = sp.Rollback(ctx) }()
		rows, err := sp.Query(ctx, sql)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			vals, err := rows.Values()
			if err != nil {
				return nil, err
			}
			cols := make([]string, len(vals))
			for i, x := range vals {
				cols[i] = fmt.Sprint(x)
				if u, ok := x.([16]byte); ok {
					cols[i] = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
				}
			}
			if l, ok := fn(cols); ok {
				out = append(out, l)
			} else {
				extras[rel]++
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		sort.Strings(out)
		return out, nil
	}
	label := func(id string) (string, bool) { l, ok := w.labels[id]; return l, ok }
	var err error
	steps := []struct {
		dst  *[]string
		rel  string
		sql  string
		conv func([]string) (string, bool)
	}{
		{&v.RoleBindings, "role_bindings", `SELECT subject_user_id, scope_id FROM app.role_bindings`, func(c []string) (string, bool) { l, ok := w.byKey[c[0]+"|"+c[1]]; return l, ok }},
		{&v.OrgMemberships, "organization_memberships", `SELECT user_id, organization_id FROM app.organization_memberships`, func(c []string) (string, bool) { l, ok := w.byKey[c[0]+"|"+c[1]]; return l, ok }},
		{&v.GroupMemberships, "group_memberships", `SELECT user_id, group_id FROM app.group_memberships`, func(c []string) (string, bool) { l, ok := w.byKey[c[0]+"|"+c[1]]; return l, ok }},
		{&v.ProfileAccessPaths, "business_profile_access_paths", `SELECT user_id, group_id, business_profile_id FROM app.business_profile_access_paths`,
			func(c []string) (string, bool) {
				b, ok := w.byKey[c[0]+"|"+c[1]]
				if !ok {
					return "", false
				}
				for l, a := range w.assoc {
					if a[0] == c[1] && a[1] == c[2] {
						return l + "/" + b, true
					}
				}
				return "", false
			}},
		{&v.Users, "users", `SELECT id FROM app.users`, func(c []string) (string, bool) { return label(c[0]) }},
		{&v.Groups, "groups", `SELECT id FROM app.groups`, func(c []string) (string, bool) { return label(c[0]) }},
		{&v.Associations, "group_business_profiles", `SELECT id FROM app.group_business_profiles`, func(c []string) (string, bool) { return label(c[0]) }},
		{&v.Profiles, "business_profiles", `SELECT id FROM app.business_profiles`, func(c []string) (string, bool) { return label(c[0]) }},
		{&v.Organizations, "organizations", `SELECT id FROM app.organizations`, func(c []string) (string, bool) { return label(c[0]) }},
		{&v.LabelDefinitions, "label_definitions", `SELECT id FROM app.label_definitions`, func(c []string) (string, bool) { return label(c[0]) }},
		{&v.LabelAssignments, "label_assignments", `SELECT id FROM app.label_assignments`, func(c []string) (string, bool) { return label(c[0]) }},
	}
	for _, s := range steps {
		if *s.dst, err = q(s.rel, s.sql, s.conv); err != nil {
			return rlsVisible{}, nil, fmt.Errorf("%s: %w", s.rel, err)
		}
	}
	return v, extras, nil
}

// rlsUpdatable are the columns the runtime role may UPDATE at version 6 (Plan §3.1): any other column is a privilege
// 42501 whatever the row.
var rlsUpdatable = map[string]bool{"groups.name": true, "groups.slug": true, "business_profiles.display_name": true,
	"business_profiles.legal_name": true, "business_profiles.description": true, "business_profiles.industry": true,
	"business_profiles.website": true, "organizations.group_admins_edit_business_profiles": true, "role_bindings.role": true,
	"users.display_name": true, "users.updated_at": true, "users.status": true, "users.disabled_at": true}

// rlsMechanism is the Plan-owned mechanism of a refused raw UPDATE (§4.2, §4.4): "privilege" (42501 before any row is
// read) for a column runtime cannot update; "check" (WITH CHECK 42501) for a row in the UPDATE policy's existing-row set
// (groups, profiles and Organizations: the read set; role_bindings: the actor's own in-context bindings; users: self);
// "zero" otherwise.
func (w *rlsOracleWorld) rlsMechanism(c rlsCase, u rlsUpdate) string {
	if !rlsUpdatable[u.Table+"."+u.Column] {
		return "privilege"
	}
	in := func(set []string) bool { return slices.Contains(set, u.Row) }
	switch u.Table {
	case "groups":
		return map[bool]string{true: "check", false: "zero"}[c.Visible != nil && in(c.Visible.Groups)]
	case "business_profiles":
		return map[bool]string{true: "check", false: "zero"}[c.Visible != nil && in(c.Visible.Profiles)]
	case "organizations":
		return map[bool]string{true: "check", false: "zero"}[c.Visible != nil && in(c.Visible.Organizations)]
	case "users":
		return map[bool]string{true: "check", false: "zero"}[u.Row == c.Context.Actor]
	case "role_bindings":
		for _, b := range w.fx.Bindings {
			if b.Label != u.Row {
				continue
			}
			own := b.User == c.Context.Actor && b.Org == c.Context.Org && (b.ScopeType == "organization" || c.Context.Group == "" || b.Group == c.Context.Group)
			return map[bool]string{true: "check", false: "zero"}[own && c.Visible != nil && in(c.Visible.RoleBindings)]
		}
	}
	return "zero"
}

// r4Mechanism is the Plan-owned mechanism (Plan §4.4; 00015 role_bindings_update/_delete) of an oracle r4 denial that
// names no guard protection: "zero" when the management policy hides the row from the actor (the row is outside the
// active Organization, the actor holds no membership there, or the row is neither the actor's own in-context binding
// nor an Organization-scope row of an Organization manager nor an in-context Group row of an Organization or that
// Group's manager); "check" when the row is in that USING set but the managerial WITH CHECK refuses the new row (an
// UPDATE by a non-manager of the row's scope); "admitted" when the policy alone would admit it (a driver/oracle
// inconsistency, never an accepted outcome).
func (w *rlsOracleWorld) r4Mechanism(m rlsMutation) string {
	c := m.Context
	var row rlsBinding
	orgRole := ""
	groupRole := map[string]string{}
	for _, b := range w.fx.Bindings {
		if b.Label == m.Row {
			row = b
		}
		if b.User == c.Actor && b.Org == c.Org {
			if b.ScopeType == "organization" {
				orgRole = b.Role
			} else {
				groupRole[b.Group] = b.Role
			}
		}
	}
	mgr := orgRole == "org_owner" || orgRole == "org_admin"
	gmgr := groupRole[row.Group] == "group_owner" || groupRole[row.Group] == "group_admin"
	inScope := row.ScopeType == "organization" || c.Group == "" || row.Group == c.Group
	own := row.User == c.Actor && inScope
	// G060 amendment01 A1.3: the Group branch admits an Organization manager only with the override (depth zero).
	managed := (row.ScopeType == "organization" && mgr) || (row.ScopeType == "group" && inScope && (gmgr || (mgr && c.Override)))
	if row.Org != c.Org || orgRole == "" || !(own || managed) {
		return "zero"
	}
	if m.Op == "update" && !managed {
		return "check"
	}
	return "admitted"
}

// updateSQL builds the single-row UPDATE of one attempt with a new, constraint-respecting value.
func (w *rlsOracleWorld) updateSQL(t *testing.T, u rlsUpdate, n int) (string, []any) {
	t.Helper()
	id := w.id[u.Row]
	if u.Table == "role_bindings" {
		k := strings.SplitN(w.binding[u.Row], "|", 2)
		if len(k) != 2 {
			t.Fatalf("oracle update names an unknown binding %s", u.Row)
		}
		var role, scope string
		for _, b := range w.fx.Bindings {
			if b.Label == u.Row {
				scope = b.ScopeType
				role = map[string]string{"org_owner": "org_admin", "org_admin": "org_member", "org_member": "org_admin",
					"group_owner": "group_admin", "group_admin": "group_member", "group_member": "group_admin"}[b.Role]
			}
		}
		set := map[string]string{"role": `role = '` + role + `'`, "subject_user_id": `subject_user_id = '` + w.id["outsider"] + `'`,
			"scope_type": `scope_type = '` + map[string]string{"organization": "group", "group": "organization"}[scope] + `'`,
			"group_id":   `group_id = '` + w.id["A2"] + `'`, "source": `source = 'other'`, "organization_id": `organization_id = '` + w.id["B"] + `'`}[u.Column]
		if set == "" {
			t.Fatalf("no new value for role_bindings.%s", u.Column)
		}
		return `UPDATE app.role_bindings SET ` + set + ` WHERE subject_user_id = $1 AND scope_id = $2`, []any{k[0], k[1]}
	}
	if id == "" {
		t.Fatalf("oracle update names an unknown row %s", u.Row)
	}
	val := map[string]string{"name": fmt.Sprintf("'Oracle Name %d'", n), "slug": fmt.Sprintf("'oracle-slug-%d-%s'", n, id[:8]),
		"display_name": fmt.Sprintf("'Oracle Display %d'", n), "description": fmt.Sprintf("'Oracle Description %d'", n),
		"legal_name": fmt.Sprintf("'Oracle Legal %d'", n), "industry": fmt.Sprintf("'Oracle Industry %d'", n),
		"website": "'https://oracle.example.invalid'", "group_admins_edit_business_profiles": "NOT group_admins_edit_business_profiles",
		"organization_id": "'" + w.id["B"] + "'", "group_id": "'" + w.id["A4"] + "'", "business_profile_id": "'" + w.id["pAUnlinked"] + "'",
		"status": "'disabled', disabled_at = now()", "display": "'x'"}[u.Column]
	if val == "" {
		t.Fatalf("no new value for %s.%s", u.Table, u.Column)
	}
	return `UPDATE app.` + u.Table + ` SET ` + u.Column + ` = ` + val + ` WHERE id = $1`, []any{id}
}

// applyOverride is the G060 driver toggle: for a context with Override, the Organization's owner (ownerA for A, ownerB
// for B) sets group_management_override true inside the case's rolled-back transaction (the owner-only guard admits it;
// its organization.policy_changed capture rolls back with the case), then the case's own settings are restored.
func (w *rlsOracleWorld) applyOverride(t *testing.T, tx pgx.Tx, c rlsContext, user, org, group *string) bool {
	t.Helper()
	if !c.Override {
		return true
	}
	ctx := context.Background()
	owner := map[string]string{"A": "ownerA", "B": "ownerB"}[c.Org]
	if owner == "" {
		t.Fatalf("override toggle for an unmapped Organization %q", c.Org)
	}
	if err := rlsSet(ctx, tx, ptr(w.id[owner]), ptr(w.id[c.Org]), ptr(""), nil); err != nil {
		t.Fatal(err)
	}
	tag, err := tx.Exec(ctx, `UPDATE app.organizations SET group_management_override = true WHERE id = $1::uuid`, w.id[c.Org])
	if err != nil || tag.RowsAffected() != 1 {
		t.Errorf("override toggle by %s: %v (SQLSTATE %q) rows %d", owner, err, pgState(err), tag.RowsAffected())
		return false
	}
	if err := rlsSet(ctx, tx, user, org, group, nil); err != nil {
		t.Fatal(err)
	}
	return true
}

// shapeSettings maps a case's context and shape onto settings (nil = never initialized) and the Plan-owned rejection
// mechanism of an invalid shape: "22P02", "22023", "zero" (no row, including self) or "" (a valid shape).
func (w *rlsOracleWorld) shapeSettings(t *testing.T, c rlsContext) (user, org, group *string, mech string) {
	t.Helper()
	lbl := func(l string) *string {
		if l == "" {
			return ptr("")
		}
		id, ok := w.id[l]
		if !ok {
			t.Fatalf("oracle context names an unknown label %s", l)
		}
		return ptr(id)
	}
	user, org, group = lbl(c.Actor), lbl(c.Org), lbl(c.Group)
	switch c.Shape {
	case "valid":
		return user, org, group, ""
	case "no-context":
		return nil, nil, nil, "zero"
	case "missing-actor":
		return nil, org, group, "22023"
	case "empty-actor":
		return ptr(""), org, group, "22023"
	case "zero-actor":
		if c.Org == "" {
			return ptr(zeroUUID), org, group, "22023"
		}
		return ptr(zeroUUID), org, group, "22023"
	case "malformed-actor":
		return ptr("not-a-uuid"), org, group, "22P02"
	case "zero-org":
		return user, ptr(zeroUUID), ptr(""), "22023"
	case "malformed-org":
		return user, ptr("not-a-uuid"), ptr(""), "22P02"
	case "group-without-org":
		return user, ptr(""), group, "22023"
	case "zero-group":
		return user, org, ptr(zeroUUID), "22023"
	case "malformed-group":
		return user, org, ptr("not-a-uuid"), "22P02"
	case "group-of-other-org":
		return user, org, group, "data"
	}
	t.Fatalf("unmapped oracle shape %q", c.Shape)
	return nil, nil, nil, ""
}

func eqSet(a, b []string) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	sort.Strings(x)
	sort.Strings(y)
	return slices.Equal(x, y)
}

// --- TestRLSIsolation ----------------------------------------------------------------------------------------------------

func TestRLSIsolation(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	rt := e.rlsRuntime(t)
	w := e.oracleWorld(t)

	t.Run("RLS-ATTEST the runtime login is attested; a privileged assertion connection is refused", func(t *testing.T) {
		if err := e.rlsAttest(rt); err != nil {
			t.Fatal(err)
		}
		if err := e.rlsAttest(e.setup); err == nil {
			t.Fatal("the superuser setup connection passed the runtime attestation")
		}
		own := itConnect(t, e.cfg, e.cfg.MigratorRole(), e.cfg.Password)
		if _, err := own.Exec(ctx, "SET ROLE "+pgx.Identifier{e.cfg.OwnerRole()}.Sanitize()); err != nil {
			t.Fatal(err)
		}
		if err := e.rlsAttest(own); err == nil {
			t.Fatal("migrator SET ROLE owner passed the runtime attestation")
		}
	})

	t.Run("RLS-SCOPE the independent oracle (C1, AC5, AC8 identity sets)", func(t *testing.T) {
		disagree := 0
		report := func(c rlsCase, what string) {
			disagree++
			line, _ := json.Marshal(map[string]string{"case": c.Name, "finding": what})
			fmt.Println("ORACLE-DISAGREEMENT " + string(line))
			t.Errorf("ORACLE-DISAGREEMENT %s: %s", c.Name, what)
		}
		before := e.state(t, tenancy.OrganizationID(mustUUID(t, w.id["A"])), tenancy.OrganizationID(mustUUID(t, w.id["B"])))
		for _, c := range rlsExpectedCases() {
			user, org, group, mech := w.shapeSettings(t, c.Context)
			if c.Reject != "" {
				// Every read returns no row or fails with the Plan's SQLSTATE for that shape; no write is attempted.
				rlsTx(t, rt, user, org, group, nil, func(tx pgx.Tx) {
					for _, rel := range []string{"role_bindings", "organization_memberships", "group_memberships", "business_profile_access_paths",
						"users", "groups", "group_business_profiles", "business_profiles", "organizations", "label_definitions", "label_assignments"} {
						st, n := rlsStmt(t, tx, `SELECT 1 FROM app.`+rel)
						switch {
						case st == "" && n == 0:
							if mech == "22P02" || mech == "22023" {
								t.Errorf("%s: %s returned no row without the Plan's %s", c.Name, rel, mech)
							}
						case st == "" && n > 0:
							report(c, fmt.Sprintf("%s returned %d rows for a rejected shape", rel, n))
						case st != mech:
							t.Errorf("%s: %s failed with %s, want %s", c.Name, rel, st, mech)
						}
					}
				})
				continue
			}
			rlsTx(t, rt, user, org, group, nil, func(tx pgx.Tx) {
				if !w.applyOverride(t, tx, c.Context, user, org, group) {
					return
				}
				got, extras, err := w.visible(t, tx)
				if err != nil {
					t.Errorf("%s: read: %v (SQLSTATE %s)", c.Name, err, pgState(err))
					return
				}
				for rel, n := range extras {
					t.Errorf("%s: %d non-fixture rows of %s visible", c.Name, n, rel)
				}
				for _, cmp := range []struct {
					rel       string
					got, want []string
				}{
					{"RoleBindings", got.RoleBindings, c.Visible.RoleBindings}, {"OrgMemberships", got.OrgMemberships, c.Visible.OrgMemberships},
					{"GroupMemberships", got.GroupMemberships, c.Visible.GroupMemberships},
					{"ProfileAccessPaths", got.ProfileAccessPaths, c.Visible.ProfileAccessPaths}, {"Users", got.Users, c.Visible.Users},
					{"Groups", got.Groups, c.Visible.Groups}, {"Associations", got.Associations, c.Visible.Associations},
					{"Profiles", got.Profiles, c.Visible.Profiles}, {"Organizations", got.Organizations, c.Visible.Organizations},
					{"LabelDefinitions", got.LabelDefinitions, c.Visible.LabelDefinitions}, // G020
					{"LabelAssignments", got.LabelAssignments, c.Visible.LabelAssignments},
				} {
					if !eqSet(cmp.got, cmp.want) {
						report(c, fmt.Sprintf("%s: database %v, oracle %v", cmp.rel, cmp.got, cmp.want))
					}
				}
				// The users self row is not a pinned column set beyond the minimum: assert the minimum columns.
				for i, u := range c.Updates {
					sql, args := w.updateSQL(t, u, i)
					st, cn, n := rlsStmtC(t, tx, sql, args...)
					switch u.Expect {
					case "ok":
						if st != "" || n != 1 {
							report(c, fmt.Sprintf("%s.%s on %s: SQLSTATE %q rows %d, oracle ok", u.Table, u.Column, u.Row, st, n))
						}
					case "no_update":
						want := w.rlsMechanism(c, u)
						switch {
						case st == "" && n == 1:
							report(c, fmt.Sprintf("%s.%s on %s was updated, oracle no_update", u.Table, u.Column, u.Row))
						case u.Guard != "" && (st != rlsGuardState[u.Guard] || cn != u.Guard):
							// r4: the oracle names the owner-guard protection; an unnamed error of the same class is not it.
							t.Errorf("%s: %s.%s on %s: SQLSTATE %q constraint %q rows %d, want %s %s", c.Name, u.Table, u.Column, u.Row, st, cn, n, rlsGuardState[u.Guard], u.Guard)
						case u.Guard != "":
						case want == "zero" && !(st == "" && n == 0):
							t.Errorf("%s: %s.%s on %s: SQLSTATE %q rows %d, want zero rows (Plan mechanism)", c.Name, u.Table, u.Column, u.Row, st, n)
						case (want == "privilege" || want == "check") && st != "42501":
							t.Errorf("%s: %s.%s on %s: SQLSTATE %q rows %d, want 42501 (%s)", c.Name, u.Table, u.Column, u.Row, st, n, want)
						}
					default:
						t.Fatalf("%s: unknown oracle outcome %q", c.Name, u.Expect)
					}
				}
				if mech == "data" {
					t.Logf("%s: data-dependent shape", c.Name)
				}
			})
		}
		if after := e.state(t, tenancy.OrganizationID(mustUUID(t, w.id["A"])), tenancy.OrganizationID(mustUUID(t, w.id["B"]))); after != before {
			t.Fatal("the oracle fixture changed")
		}
		fmt.Printf("RLS-SCOPE {\"cases\":%d,\"disagreements\":%d}\n", len(rlsExpectedCases()), disagree)
	})

	t.Run("RLS-SCOPE last-path removal and second path (AC5)", func(t *testing.T) {
		// gOwnerA1 reaches pA12 only through A1 (A3's path is pA3). Removing its A1 binding removes pA12 with no profile
		// write; restoring it restores exactly that path.
		g, o, a1 := w.id["gOwnerA1"], w.id["A"], w.id["A1"]
		see := func() []string {
			var out []string
			rlsTx(t, rt, &g, &o, ptr(""), nil, func(tx pgx.Tx) {
				v, _, err := w.visible(t, tx)
				if err != nil {
					t.Fatal(err)
				}
				out = v.Profiles
			})
			return out
		}
		if got := see(); !eqSet(got, []string{"pA12", "pA3"}) {
			t.Fatalf("before: %v", got)
		}
		profileHash := e.part(t, "business_profiles", tenancy.OrganizationID(mustUUID(t, o)))
		if _, err := e.setup.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1 AND scope_id = $2`, g, a1); err != nil {
			t.Fatal(err)
		}
		if got := see(); !eqSet(got, []string{"pA3"}) {
			t.Errorf("after last-path removal: %v, want pA3 only", got)
		}
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1, $2, 'group', $3, $3, 'group_owner', 'membership')`, o, g, a1); err != nil {
			t.Fatal(err)
		}
		if got := see(); !eqSet(got, []string{"pA12", "pA3"}) {
			t.Errorf("after restoring the A1 path: %v", got)
		}
		if e.part(t, "business_profiles", tenancy.OrganizationID(mustUUID(t, o))) != profileHash {
			t.Error("a profile row changed")
		}
	})

	t.Run("RLS-SCOPE-INSERT oracle r3 inserts and the association DELETE refusal (G018)", func(t *testing.T) {
		before := e.state(t, tenancy.OrganizationID(mustUUID(t, w.id["A"])), tenancy.OrganizationID(mustUUID(t, w.id["B"])))
		for i, c := range rlsExpectedInserts() {
			user, org, group, _ := w.shapeSettings(t, c.Context)
			rlsTx(t, rt, user, org, group, nil, func(tx pgx.Tx) {
				if !w.applyOverride(t, tx, c.Context, user, org, group) {
					return
				}
				var sql string
				var args []any
				switch {
				case c.Op == "delete":
					sql, args = `DELETE FROM app.group_business_profiles WHERE id = $1::uuid`, []any{w.id[c.Row]}
				case c.Table == "groups":
					sql, args = `INSERT INTO app.groups (id, organization_id, name, slug) VALUES ($1::uuid, $2::uuid, $3, $4)`,
						[]any{rlsUUID(t), w.id[c.Org], fmt.Sprintf("Oracle r3 %d", i), fmt.Sprintf("oracle-r3-%d", i)}
				case c.Table == "role_bindings":
					sql = `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1::uuid, $2::uuid, 'group', $3::uuid, $3::uuid, 'group_owner', 'membership')`
					args = []any{w.id[c.Org], w.id[c.Subject], w.id[c.Group]}
				case c.Table == "business_profiles":
					sql, args = `INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES ($1::uuid, $2::uuid, $3)`,
						[]any{rlsUUID(t), w.id[c.Org], fmt.Sprintf("Oracle r3 %d", i)}
				case c.Table == "group_business_profiles":
					p := w.id[c.Profile]
					if c.Profile == "new" {
						p = rlsUUID(t)
						if _, err := tx.Exec(context.Background(), `INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES ($1::uuid, $2::uuid, 'Oracle r3 new')`,
							p, w.id[c.Org]); err != nil {
							t.Errorf("%s: the same-transaction profile insert failed: %v", c.Name, err)
							return
						}
					}
					sql, args = `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ($1::uuid, $2::uuid, $3::uuid)`,
						[]any{w.id[c.Org], w.id[c.Group], p}
				default:
					t.Fatalf("%s: unmapped table %s", c.Name, c.Table)
				}
				st, n := rlsStmt(t, tx, sql, args...)
				switch c.Expect {
				case "ok":
					if st != "" || n != 1 {
						t.Errorf("ORACLE-DISAGREEMENT %s: SQLSTATE %q rows %d, oracle ok", c.Name, st, n)
					}
				case "denied":
					if st != "42501" || n != 0 {
						t.Errorf("ORACLE-DISAGREEMENT %s: SQLSTATE %q rows %d, oracle denied (42501)", c.Name, st, n)
					}
				default:
					t.Fatalf("%s: unknown oracle outcome %q", c.Name, c.Expect)
				}
			})
		}
		if after := e.state(t, tenancy.OrganizationID(mustUUID(t, w.id["A"])), tenancy.OrganizationID(mustUUID(t, w.id["B"]))); after != before {
			t.Error("an oracle r3 attempt changed state (every attempt is rolled back)")
		}
	})

	t.Run("RLS-SCOPE-MUTATE oracle r4 role_bindings UPDATE and DELETE (G019)", func(t *testing.T) {
		// Each rlsMutation runs in its own rolled-back runtime transaction: the Pre steps (admitted, run directly in the
		// transaction), then the attempt. "ok" requires exactly one row, a success, and the readback of the target role (or
		// no row after DELETE) as the same actor; "denied" requires no change through the Plan-owned mechanism: the named
		// guard protection with its class (rlsGuardState), or the policy itself (r4Mechanism: a hidden row affects zero
		// rows, a row in the actor's USING set fails the managerial WITH CHECK with 42501 and no constraint).
		before := e.state(t, tenancy.OrganizationID(mustUUID(t, w.id["A"])), tenancy.OrganizationID(mustUUID(t, w.id["B"])))
		key := func(label string) (string, string) {
			k := strings.SplitN(w.binding[label], "|", 2)
			if len(k) != 2 {
				t.Fatalf("oracle r4 names an unknown binding %s", label)
			}
			return k[0], k[1]
		}
		seen := map[string]bool{}
		counts := map[string]int{}
		for _, m := range rlsExpectedMutations() {
			if seen[m.Name] {
				t.Fatalf("duplicate oracle r4 case name %q", m.Name)
			}
			seen[m.Name] = true
			user, org, group, _ := w.shapeSettings(t, m.Context)
			rlsTx(t, rt, user, org, group, nil, func(tx pgx.Tx) {
				ctx := context.Background()
				if !w.applyOverride(t, tx, m.Context, user, org, group) {
					return
				}
				// G060 amendment01: the dependent rows a departure removes exist (and are the actor's to see) before it.
				for _, a := range m.Also {
					as, aid := key(a)
					var n int
					if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.role_bindings WHERE subject_user_id = $1 AND scope_id = $2`, as, aid).Scan(&n); err != nil || n != 1 {
						t.Errorf("%s: dependent row %s before the attempt: %d (%v)", m.Name, a, n, err)
						return
					}
				}
				for _, p := range m.Pre {
					s, id := key(p.Row)
					tag, err := tx.Exec(ctx, `UPDATE app.role_bindings SET role = $3 WHERE subject_user_id = $1 AND scope_id = $2`, s, id, p.Role)
					if err != nil || tag.RowsAffected() != 1 {
						t.Errorf("ORACLE-DISAGREEMENT %s: pre step %s -> %s: %v (SQLSTATE %q, constraint %q), oracle ok", m.Name, p.Row, p.Role, err, pgState(err), pgConstraint(err))
						return
					}
				}
				s, id := key(m.Row)
				var sql string
				args := []any{s, id}
				switch m.Op {
				case "update":
					sql, args = `UPDATE app.role_bindings SET role = $3 WHERE subject_user_id = $1 AND scope_id = $2`, append(args, m.Role)
				case "delete":
					sql = `DELETE FROM app.role_bindings WHERE subject_user_id = $1 AND scope_id = $2`
				default:
					t.Fatalf("%s: unknown oracle r4 op %q", m.Name, m.Op)
				}
				switch m.Expect {
				case "ok":
					counts["ok"]++
					tag, err := tx.Exec(ctx, sql, args...)
					if err != nil || tag.RowsAffected() != 1 {
						t.Errorf("ORACLE-DISAGREEMENT %s: %v (SQLSTATE %q, constraint %q) rows %d, oracle ok", m.Name, err, pgState(err), pgConstraint(err), tag.RowsAffected())
						return
					}
					// Readback as the same actor in the same transaction: the target role, or no row after DELETE. It runs in
					// the actor's Organization-only context (the selected Group cleared): a self-demotion can make the
					// selected Group unselectable, and the r2 context-validity rule then hides every tenant row including
					// the actor's own binding ('adminA selected A4 demotes itself': org_member cannot select A4), which says
					// nothing about the admitted write. Own bindings stay visible in Organization-only context to every
					// current member.
					if m.Context.Group != "" {
						if err := rlsSet(ctx, tx, nil, nil, ptr(""), nil); err != nil {
							t.Fatal(err)
						}
					}
					var roles []string
					rows, err := tx.Query(ctx, `SELECT role FROM app.role_bindings WHERE subject_user_id = $1 AND scope_id = $2`, s, id)
					if err == nil {
						for rows.Next() {
							var r string
							if err = rows.Scan(&r); err != nil {
								break
							}
							roles = append(roles, r)
						}
						rows.Close()
					}
					if err != nil {
						t.Errorf("%s: readback: %v (SQLSTATE %q)", m.Name, err, pgState(err))
						return
					}
					want := []string{m.Role}
					if m.Op == "delete" {
						want = nil
					}
					if !eqSet(roles, want) {
						t.Errorf("ORACLE-DISAGREEMENT %s: readback %v, oracle %v", m.Name, roles, want)
					}
					// G060 amendment01: the same statement removed the subject's Group bindings (the definer dependent count
					// answers for the subject itself whatever its remaining membership; every Also row is that subject's).
					if len(m.Also) > 0 {
						var dep int
						if err := tx.QueryRow(ctx, `SELECT app.rls_dependent_group_bindings($1::uuid)`, s).Scan(&dep); err != nil || dep != 0 {
							t.Errorf("ORACLE-DISAGREEMENT %s: %d dependent Group binding(s) remain (%v), oracle also removed %v", m.Name, dep, err, m.Also)
						}
					}
				case "denied":
					st, cn, n := rlsStmtC(t, tx, sql, args...)
					switch {
					case st == "" && n > 0:
						t.Errorf("ORACLE-DISAGREEMENT %s: %d row(s) changed, oracle denied (%s)", m.Name, n, m.Guard)
					case m.Guard != "":
						counts["guard:"+m.Guard]++
						if st != rlsGuardState[m.Guard] || cn != m.Guard {
							t.Errorf("ORACLE-DISAGREEMENT %s: SQLSTATE %q constraint %q rows %d, oracle denied by %s (%s)", m.Name, st, cn, n, m.Guard, rlsGuardState[m.Guard])
						}
					default:
						mech := w.r4Mechanism(m)
						counts["policy:"+mech]++
						switch {
						case mech == "zero" && !(st == "" && n == 0):
							t.Errorf("%s: SQLSTATE %q constraint %q rows %d, want zero rows (Plan mechanism: hidden row)", m.Name, st, cn, n)
						case mech == "check" && !(st == "42501" && cn == ""):
							t.Errorf("%s: SQLSTATE %q constraint %q rows %d, want 42501 without a constraint (Plan mechanism: WITH CHECK)", m.Name, st, cn, n)
						case mech == "admitted":
							t.Errorf("%s: the oracle expects a policy refusal but the Plan's policy admits the shape (driver model); observed SQLSTATE %q constraint %q rows %d", m.Name, st, cn, n)
						}
					}
				default:
					t.Fatalf("%s: unknown oracle outcome %q", m.Name, m.Expect)
				}
			})
		}
		if after := e.state(t, tenancy.OrganizationID(mustUUID(t, w.id["A"])), tenancy.OrganizationID(mustUUID(t, w.id["B"]))); after != before {
			t.Error("an oracle r4 attempt changed state (every attempt is rolled back)")
		}
		keys := make([]string, 0, len(counts))
		for k := range counts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%q:%d", k, counts[k]))
		}
		fmt.Printf("RLS-SCOPE-MUTATE {\"cases\":%d,%s}\n", len(rlsExpectedMutations()), strings.Join(parts, ","))
	})

	t.Run("RLS-CTX every field varied independently (C3, AC3)", func(t *testing.T) {
		u, o, g := w.id["ownerA"], w.id["A"], w.id["A1"]
		foreignUser, foreignOrg, foreignGroup := w.id["outsider"], w.id["B"], w.id["B1"]
		hexb := hex.EncodeToString(make([]byte, 32))
		type variant struct {
			name             string
			user, org, group *string
			binding          *string
			want             string // "22P02", "22023", "none" (no tenant row, self users row only), "self-none" (no row), "ok"
		}
		vs := []variant{
			{"user malformed", ptr("not-a-uuid"), &o, &g, ptr(""), "22P02"},
			{"user zero", ptr(zeroUUID), &o, &g, ptr(""), "22023"},
			{"user foreign (trust limit: reads as that subject)", &foreignUser, &o, &g, ptr(""), "none"},
			{"user absent", ptr(""), &o, &g, ptr(""), "22023"},
			{"org malformed", &u, ptr("not-a-uuid"), ptr(""), ptr(""), "22P02"},
			{"org zero", &u, ptr(zeroUUID), ptr(""), ptr(""), "22023"},
			{"org foreign", &u, &foreignOrg, ptr(""), ptr(""), "none"},
			{"org absent (identity shape)", &u, ptr(""), ptr(""), ptr(""), "none"},
			{"group malformed", &u, &o, ptr("not-a-uuid"), ptr(""), "22P02"},
			{"group zero", &u, &o, ptr(zeroUUID), ptr(""), "22023"},
			{"group foreign (inconsistent)", &u, &o, &foreignGroup, ptr(""), "none"},
			{"group absent (Organization-only)", &u, &o, ptr(""), ptr(""), "ok"},
			{"group without Organization", &u, ptr(""), &g, ptr(""), "22023"},
			{"binding malformed", &u, &o, &g, ptr("zz"), "22023"},
			{"binding with a user", &u, &o, &g, &hexb, "22023"},
			{"binding uppercase", ptr(""), ptr(""), ptr(""), ptr(strings.ToUpper(hex.EncodeToString(sha256.New().Sum(nil)))), "22023"},
			{"valid selected Group", &u, &o, &g, ptr(""), "ok"},
		}
		reads := []string{"role_bindings", "organization_memberships", "group_memberships", "business_profile_access_paths", "users",
			"groups", "group_business_profiles", "business_profiles", "organizations", "sessions", "identity_principals", "login_transactions"}
		writes := []string{
			`UPDATE app.groups SET name = 'Ctx Write' WHERE id = '` + g + `'`,
			`UPDATE app.groups SET slug = 'ctx-write' WHERE id = '` + g + `'`,
			`UPDATE app.business_profiles SET display_name = 'Ctx Write' WHERE id = '` + w.id["pA12"] + `'`,
			`UPDATE app.organizations SET group_admins_edit_business_profiles = false WHERE id = '` + o + `'`,
			`UPDATE app.users SET display_name = 'Ctx Write' WHERE id = '` + u + `'`,
		}
		// One pending login transaction, so that a login_transactions read has a candidate row to evaluate (with none,
		// every shape correctly yields zero rows).
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, created_at, expires_at)
VALUES (sha256(convert_to($1, 'UTF8')), sha256(convert_to($1 || 'b', 'UTF8')), repeat('v', 43), 'c', 'https://o', '/', now(), now() + interval '10 minutes')`, rlsUUID(t)); err != nil {
			t.Fatal(err)
		}
		// And a session and a principal of the fixed actor, so that every read has a candidate row.
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at)
VALUES ($1::uuid, sha256(convert_to($1::text || 'ctx', 'UTF8')), sha256(convert_to($1::text || 'csrf', 'UTF8')), now(), now() + interval '1 hour', now() + interval '2 hours')`, u); err != nil {
			t.Fatal(err)
		}
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('https://ctx.invalid', 'user_ctx' || left(md5($1::text), 8), $1::uuid)`, u); err != nil {
			t.Fatal(err)
		}
		before := e.state(t, tenancy.OrganizationID(mustUUID(t, o)))
		for _, v := range vs {
			rlsTx(t, rt, v.user, v.org, v.group, v.binding, func(tx pgx.Tx) {
				for _, rel := range reads {
					st, n := rlsStmt(t, tx, `SELECT 1 FROM app.`+rel)
					switch v.want {
					case "22P02", "22023":
						if st != v.want {
							t.Errorf("%s: read %s: SQLSTATE %q rows %d, want %s", v.name, rel, st, n, v.want)
						}
					case "none":
						tenant := rel != "users" && rel != "sessions" && rel != "identity_principals" && rel != "login_transactions"
						if st != "" || (tenant && n != 0) {
							t.Errorf("%s: read %s: SQLSTATE %q rows %d, want no tenant row", v.name, rel, st, n)
						}
					case "ok":
						if st != "" {
							t.Errorf("%s: read %s: SQLSTATE %q", v.name, rel, st)
						}
					}
				}
				for _, s := range writes {
					st, n := rlsStmt(t, tx, s)
					switch v.want {
					case "22P02", "22023":
						if st != v.want {
							t.Errorf("%s: %s: SQLSTATE %q rows %d, want %s", v.name, s, st, n, v.want)
						}
					case "none":
						if (st != "" && st != "42501") || (st == "" && n != 0 && !strings.Contains(s, "app.users")) {
							t.Errorf("%s: %s: SQLSTATE %q rows %d, want no tenant mutation", v.name, s, st, n)
						}
					}
				}
			})
		}
		if e.state(t, tenancy.OrganizationID(mustUUID(t, o))) != before {
			t.Fatal("rows changed (every write was rolled back)")
		}
		// set_config(..., true) outside a transaction block is ineffective: the next statement has no context.
		if _, err := rt.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`, u, o); err != nil {
			t.Fatal(err)
		}
		var n int
		if err := rt.QueryRow(ctx, `SELECT count(*) FROM app.organizations`).Scan(&n); err != nil || n != 0 {
			t.Errorf("autocommit set_config(..., true) gave %d Organization rows (%v)", n, err)
		}
	})

	t.Run("RLS-ORG fixed adminA without application filters (AC4)", func(t *testing.T) {
		u, o, b := w.id["adminA"], w.id["A"], w.id["B"]
		rlsTx(t, rt, &u, &o, ptr(""), nil, func(tx pgx.Tx) {
			for _, c := range []struct {
				sql  string
				want int64
			}{
				{`SELECT count(*) FROM app.groups`, 1},
				{`SELECT g.id FROM app.groups g JOIN app.group_business_profiles a ON a.group_id = g.id JOIN app.business_profiles p ON p.id = a.business_profile_id`, 3},
				// G060 (override off): the 8 Organization-scope bindings and adminA's own A1 member binding (all 15 with it on).
				{`SELECT rb.scope_id FROM app.role_bindings rb JOIN app.users u ON u.id = rb.subject_user_id`, 9},
				{`SELECT id FROM app.groups WHERE organization_id IN (SELECT id FROM app.organizations)`, 4},
				{`SELECT id FROM app.groups WHERE organization_id = '` + b + `'`, 0},
				{`SELECT id FROM app.business_profiles WHERE organization_id = '` + b + `'`, 0},
				// G020 (frank, override off): the two Organization-owned definitions and their two assignments joined to the
				// Groups; no Group-owned definition (lA3Own/lA3Tm/lA4Own/lA4Tm are present) and nothing of B.
				{`SELECT id FROM app.label_definitions`, 2},
				{`SELECT a.id FROM app.label_assignments a JOIN app.label_definitions d ON d.id = a.label_id JOIN app.groups g ON g.id = a.target_group_id`, 2},
				{`SELECT id FROM app.label_definitions WHERE owner_scope_type = 'group'`, 0},
				{`SELECT id FROM app.label_definitions WHERE organization_id = '` + b + `'`, 0},
				{`SELECT id FROM app.label_assignments WHERE organization_id = '` + b + `'`, 0},
			} {
				st, n := rlsStmt(t, tx, c.sql)
				if st != "" || n != c.want {
					t.Errorf("%s: SQLSTATE %q rows %d, want %d", c.sql, st, n, c.want)
				}
			}
			var groups int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.groups`).Scan(&groups); err != nil || groups != 4 {
				t.Errorf("aggregate over groups: %d (%v), want the 4 Groups of A", groups, err)
			}
			for _, c := range []struct {
				sql, state string
				rows       int64
			}{
				{`UPDATE app.groups SET name = name WHERE organization_id = '` + b + `' RETURNING id`, "", 0},
				{`UPDATE app.business_profiles SET display_name = display_name WHERE organization_id = '` + b + `' RETURNING id`, "", 0},
				// G060 (override off): an Organization admin with only a member binding of A1 renames no Group (WITH CHECK).
				{`UPDATE app.groups SET name = 'Org Positive' WHERE id = '` + w.id["A1"] + `' RETURNING id`, "42501", 0},
				// G018 version 14: group.create admits a Group insert in the own Organization (Organization-only context; the
				// bootstrap trigger writes the creator's owner binding); a Group of another Organization stays refused.
				{`INSERT INTO app.groups (organization_id, name, slug) VALUES ('` + o + `', 'X', 'org-insert')`, "", 1},
				{`INSERT INTO app.groups (organization_id, name, slug) VALUES ('` + b + `', 'X', 'org-insert')`, "42501", 0},
				{`DELETE FROM app.groups WHERE organization_id = '` + b + `'`, "42501", 0},
			} {
				st, n := rlsStmt(t, tx, c.sql)
				if st != c.state || n != c.rows {
					t.Errorf("%s: SQLSTATE %q rows %d, want %q %d", c.sql, st, n, c.state, c.rows)
				}
			}
			// G060: the same fixed actor with the override on (set by ownerA in this rolled-back transaction): every Group
			// roster row of A and the Group rename; still nothing of B.
			if !w.applyOverride(t, tx, rlsContext{Actor: "adminA", Org: "A", Shape: "valid", Override: true}, &u, &o, ptr("")) {
				return
			}
			for _, c := range []struct {
				sql, state string
				rows       int64
			}{
				// The 15 fixture bindings of A (every rlsStmt attempt above ran in a rolled-back savepoint).
				{`SELECT rb.scope_id FROM app.role_bindings rb JOIN app.users u ON u.id = rb.subject_user_id`, "", 15},
				{`UPDATE app.groups SET name = 'Org Positive' WHERE id = '` + w.id["A1"] + `' RETURNING id`, "", 1},
				{`UPDATE app.groups SET name = name WHERE organization_id = '` + b + `' RETURNING id`, "", 0},
				// G020: the override confers nothing on Group-owned labels.
				{`SELECT id FROM app.label_definitions WHERE owner_scope_type = 'group'`, "", 0},
				{`SELECT id FROM app.label_definitions`, "", 2},
			} {
				st, n := rlsStmt(t, tx, c.sql)
				if st != c.state || n != c.rows {
					t.Errorf("override on: %s: SQLSTATE %q rows %d, want %q %d", c.sql, st, n, c.state, c.rows)
				}
			}
		})
	})

	t.Run("RLS-AUTH fixed-actor authentication privacy and the E5 raw controls (C2, AC8)", func(t *testing.T) {
		x, y := e.newUser(t), e.newUser(t)
		xs, ys := x.id.String(), y.id.String()
		// Pending login transactions of three browsers, two of them expired.
		bx, by, bz := sha256.Sum256([]byte(rlsUUID(t))), sha256.Sum256([]byte(rlsUUID(t))), sha256.Sum256([]byte(rlsUUID(t)))
		ins := func(b [32]byte, expired bool) {
			created := "now()"
			if expired {
				created = "now() - interval '20 minutes'"
			}
			if _, err := e.setup.Exec(ctx, `INSERT INTO app.login_transactions (state_hash, binding_hash, pkce_verifier, client_id, origin, return_path, created_at, expires_at)
VALUES (sha256(convert_to($1, 'UTF8')), $2, repeat('v', 43), 'c', 'https://o', '/', `+created+`, `+created+` + interval '10 minutes')`, rlsUUID(t), b[:]); err != nil {
				t.Fatal(err)
			}
		}
		ins(bx, false)
		ins(by, true)
		ins(bz, true)
		hx, hy := hex.EncodeToString(bx[:]), hex.EncodeToString(by[:])
		count := func(q string, args ...any) int64 {
			var n int64
			if err := e.setup.QueryRow(ctx, q, args...).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		xSessions := count(`SELECT count(*) FROM app.sessions WHERE user_id = $1`, xs)
		rlsTx(t, rt, &xs, ptr(""), ptr(""), ptr(""), func(tx pgx.Tx) {
			for _, c := range []struct {
				sql  string
				want int64
			}{
				{`SELECT 1 FROM app.sessions`, xSessions},
				{`SELECT 1 FROM app.sessions s JOIN app.users u ON u.id = s.user_id`, xSessions},
				{`SELECT 1 FROM app.identity_principals`, 1},
				{`SELECT 1 FROM app.identity_principals p JOIN app.users u ON u.id = p.user_id`, 1},
				{`SELECT 1 FROM app.login_transactions`, 0},
				{`SELECT 1 FROM app.users`, 1},
				{`SELECT 1 FROM app.sessions WHERE user_id = '` + ys + `'`, 0},
			} {
				if st, n := rlsStmt(t, tx, c.sql); st != "" || n != c.want {
					t.Errorf("IDENTITY %s: SQLSTATE %q rows %d, want %d", c.sql, st, n, c.want)
				}
			}
			var agg int64
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.sessions`).Scan(&agg); err != nil || agg != xSessions {
				t.Errorf("count(*) on sessions: %d (%v)", agg, err)
			}
			for _, c := range []struct{ sql, state string }{
				{`UPDATE app.sessions SET revoked_at = NULL, revoke_reason = NULL WHERE user_id = '` + ys + `'`, ""},
				{`UPDATE app.sessions SET idle_expires_at = absolute_expires_at WHERE user_id = '` + ys + `'`, ""},
				{`UPDATE app.users SET display_name = 'Taken' WHERE id = '` + ys + `'`, ""},
				{`DELETE FROM app.login_transactions WHERE binding_hash = decode('` + hx + `', 'hex')`, ""},
				{`DELETE FROM app.sessions WHERE user_id = '` + ys + `'`, "42501"},
				{`INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at)
VALUES ('` + ys + `', sha256('x'), sha256('y'), now(), now() + interval '1 hour', now() + interval '2 hours')`, "42501"},
				{`INSERT INTO app.identity_principals (issuer, subject, user_id) VALUES ('i', 'user_01FORGED', '` + ys + `')`, "42501"},
			} {
				if st, n := rlsStmt(t, tx, c.sql); st != c.state || n != 0 {
					t.Errorf("IDENTITY %s: SQLSTATE %q rows %d, want %q 0", c.sql, st, n, c.state)
				}
			}
			// Own session writes in IDENTITY shape are the pinned self-scope writes.
			if st, n := rlsStmt(t, tx, `UPDATE app.sessions SET renewed_at = renewed_at WHERE user_id = '`+xs+`'`); st != "" || n != xSessions {
				t.Errorf("own session renewal write: %q %d", st, n)
			}
		})
		// TENANT shape: own identity and session rows readable and lockable, never writable.
		org := w.id["A"]
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source)
VALUES ($1, $2, 'organization', $1, 'org_member', 'membership')`, org, xs); err != nil {
			t.Fatal(err)
		}
		rlsTx(t, rt, &xs, &org, ptr(""), ptr(""), func(tx pgx.Tx) {
			if st, n := rlsStmt(t, tx, `SELECT 1 FROM app.sessions WHERE user_id = '`+xs+`' FOR SHARE`); st != "" || n != xSessions {
				t.Errorf("TENANT own session lock: %q %d", st, n)
			}
			for _, s := range []string{
				`UPDATE app.sessions SET renewed_at = renewed_at WHERE user_id = '` + xs + `'`,
				`UPDATE app.users SET display_name = 'Tenant' WHERE id = '` + xs + `'`,
				`UPDATE app.users SET status = 'active', disabled_at = NULL WHERE id = '` + xs + `'`,
				`INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at)
VALUES ('` + xs + `', sha256('t'), sha256('u'), now(), now() + interval '1 hour', now() + interval '2 hours')`,
			} {
				if st, _ := rlsStmt(t, tx, s); st != "42501" {
					t.Errorf("TENANT %s: SQLSTATE %q, want 42501 (WITH CHECK)", s, st)
				}
			}
		})
		// LOGIN shape: own binding only.
		rlsTx(t, rt, ptr(""), ptr(""), ptr(""), &hx, func(tx pgx.Tx) {
			if st, n := rlsStmt(t, tx, `SELECT 1 FROM app.login_transactions`); st != "" || n != 1 {
				t.Errorf("LOGIN own binding: %q %d", st, n)
			}
			if st, n := rlsStmt(t, tx, `SELECT 1 FROM app.sessions`); st != "" || n != 0 {
				t.Errorf("LOGIN sessions: %q %d", st, n)
			}
			if st, n := rlsStmt(t, tx, `DELETE FROM app.login_transactions WHERE binding_hash = decode('`+hy+`', 'hex')`); st != "" || n != 0 {
				t.Errorf("LOGIN delete of another binding: %q %d", st, n)
			}
		})
		// G058 ENR-21d: one live and one expired enrollment credential of the subject (committed by the superuser, removed
		// after the EMPTY-shape checks).
		const enrLive, enrExpired = "0f58e000-0000-4000-8000-0000000021d1", "0f58e000-0000-4000-8000-0000000021d2"
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.enrollment_sessions (id, user_id, organization_id, origin, credential_hash, csrf_token,
  provider_session_id, created_at, expires_at) VALUES
  ($1, $3, $4, 'https://enr-21d.example.test', sha256('enr-21d-live'), sha256('enr-21d-csrf-live'), 'enr-21d-ps', now(), now() + interval '20 minutes'),
  ($2, $3, $4, 'https://enr-21d.example.test', sha256('enr-21d-expired'), sha256('enr-21d-csrf-expired'), 'enr-21d-ps', now() - interval '1 hour',
   now() - interval '50 minutes')`, enrLive, enrExpired, xs, org); err != nil {
			t.Fatalf("enrollment credential fixture: %v", err)
		}
		t.Cleanup(func() {
			_, _ = e.setup.Exec(context.Background(), `DELETE FROM app.enrollment_sessions WHERE id IN ($1, $2)`, enrLive, enrExpired)
		})
		var enrExpiredAll int64
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.enrollment_sessions WHERE expires_at <= now()`).Scan(&enrExpiredAll); err != nil ||
			enrExpiredAll < 1 {
			t.Fatalf("expired enrollment credentials: %d %v", enrExpiredAll, err)
		}
		// EMPTY shape: nothing readable. Of app.enrollment_sessions the only access is the expiry sweep's DELETE
		// eligibility (enrollment_sessions_expire: expires_at <= now() with NULL ctx_user() and NULL ctx_login_binding()):
		// identity's unqualified sweep statement (store.go sqlExpireEnrollmentSessions) removes exactly the expired rows; a
		// DELETE naming a row by a column also needs the SELECT policy (user_id = ctx_user(), NULL here), so it reaches no
		// row, live or expired.
		rlsTx(t, rt, ptr(""), ptr(""), ptr(""), ptr(""), func(tx pgx.Tx) {
			for _, rel := range []string{"sessions", "identity_principals", "login_transactions", "users", "enrollment_sessions"} {
				if st, n := rlsStmt(t, tx, `SELECT 1 FROM app.`+rel); st != "" || n != 0 {
					t.Errorf("EMPTY %s: %q %d", rel, st, n)
				}
			}
			if st, n := rlsStmt(t, tx, `DELETE FROM app.enrollment_sessions WHERE id = $1`, enrLive); st != "" || n != 0 {
				t.Errorf("EMPTY delete of a live enrollment credential: %q %d, want 0 rows", st, n)
			}
			if st, n := rlsStmt(t, tx, `DELETE FROM app.enrollment_sessions WHERE id = $1`, enrExpired); st != "" || n != 0 {
				t.Errorf("EMPTY delete of an expired enrollment credential by id: %q %d, want 0 rows (SELECT policy)", st, n)
			}
			if st, n := rlsStmt(t, tx, `DELETE FROM app.enrollment_sessions`); st != "" || n != enrExpiredAll {
				t.Errorf("EMPTY expiry sweep: %q %d, want exactly the %d expired rows", st, n, enrExpiredAll)
			}
		})
		var enrLeft int
		if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM app.enrollment_sessions WHERE id IN ($1, $2)`, enrLive, enrExpired).Scan(&enrLeft); err != nil ||
			enrLeft != 2 {
			t.Errorf("enrollment credentials after the rolled-back EMPTY transaction: %d %v, want 2", enrLeft, err)
		}
		if _, err := e.setup.Exec(ctx, `DELETE FROM app.enrollment_sessions WHERE id IN ($1, $2)`, enrLive, enrExpired); err != nil {
			t.Fatal(err)
		}
		// E1-E3: exact rows for exact keys, nothing for wrong keys, identical under every context and a poisoned setting.
		var credHash []byte
		if err := e.setup.QueryRow(ctx, `SELECT credential_hash FROM app.sessions WHERE user_id = $1 AND revoked_at IS NULL LIMIT 1`, xs).Scan(&credHash); err != nil {
			t.Fatal(err)
		}
		e1 := func(tx pgx.Tx, h []byte) string {
			var s string
			if err := tx.QueryRow(ctx, `SELECT coalesce(string_agg(user_id::text, ','), '') FROM app.auth_resolve_session($1, now(), $2)`, h, itOrigin).Scan(&s); err != nil {
				t.Fatalf("E1: %v", err)
			}
			return s
		}
		for _, ctxs := range [][4]*string{{ptr(""), ptr(""), ptr(""), ptr("")}, {&ys, ptr(""), ptr(""), ptr("")}, {&xs, &org, ptr(""), ptr("")}, {ptr(""), ptr(""), ptr(""), &hy}} {
			rlsTx(t, rt, ctxs[0], ctxs[1], ctxs[2], ctxs[3], func(tx pgx.Tx) {
				if got := e1(tx, credHash); got != xs {
					t.Errorf("E1 exact key: %s", itH(got))
				}
				if got := e1(tx, make([]byte, 32)); got != "" {
					t.Error("E1 wrong key returned a row")
				}
				var subject string
				if err := tx.QueryRow(ctx, `SELECT coalesce(app.auth_principal_subject(issuer, subject)::text, '') FROM (SELECT 'none' AS issuer, 'user_none' AS subject) s`).Scan(&subject); err != nil || subject != "" {
					t.Errorf("E2 unknown principal: %q %v", subject, err)
				}
				var refs int
				if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.auth_logout_ref($1, now())`, make([]byte, 32)).Scan(&refs); err != nil || refs != 0 {
					t.Errorf("E3 wrong key: %d %v", refs, err)
				}
			})
		}
		// E5 raw-statement controls (a)-(e) (the production path, (f), is in the identity package).
		e5 := func(user, org, group, binding *string, sql string) (string, int64) {
			var st string
			var n int64
			rlsTx(t, rt, user, org, group, binding, func(tx pgx.Tx) { st, n = rlsStmt(t, tx, sql) })
			return st, n
		}
		expired := count(`SELECT count(*) FROM app.login_transactions WHERE expires_at <= now()`)
		if expired < 2 {
			t.Fatalf("fixture: %d expired transactions", expired)
		}
		const cleanup = `DELETE FROM app.login_transactions`
		if st, n := e5(ptr(""), ptr(""), ptr(""), ptr(""), cleanup); st != "" || n != expired {
			t.Errorf("(a) EMPTY cleanup: %q %d rows, want exactly the %d expired rows", st, n, expired)
		}
		for _, c := range []struct {
			name             string
			user, org, group *string
			binding          *string
			state            string
		}{
			{"malformed user", ptr("x"), ptr(""), ptr(""), ptr(""), "22P02"},
			{"malformed organization", &xs, ptr("x"), ptr(""), ptr(""), "22P02"},
			{"malformed group", &xs, &org, ptr("x"), ptr(""), "22P02"},
			{"malformed binding", ptr(""), ptr(""), ptr(""), ptr("x"), "22023"},
			{"zero user", ptr(zeroUUID), ptr(""), ptr(""), ptr(""), "22023"},
			{"Group without Organization", &xs, ptr(""), &org, ptr(""), "22023"},
			{"Organization without user", ptr(""), &org, ptr(""), ptr(""), "22023"},
			{"binding with user", &xs, ptr(""), ptr(""), &hy, "22023"},
		} {
			if st, n := e5(c.user, c.org, c.group, c.binding, cleanup); st != c.state || n != 0 {
				t.Errorf("(b) %s: %q %d, want %s", c.name, st, n, c.state)
			}
		}
		for _, c := range []struct {
			name             string
			user, org, group *string
		}{{"IDENTITY", &xs, ptr(""), ptr("")}, {"TENANT-ORG", &xs, &org, ptr("")}, {"TENANT-GROUP", &xs, &org, ptr(w.id["A1"])}} {
			if st, n := e5(c.user, c.org, c.group, ptr(""), cleanup); st != "" || n != 0 {
				t.Errorf("(c) %s cleanup: %q %d, want nothing deleted", c.name, st, n)
			}
		}
		if st, n := e5(ptr(""), ptr(""), ptr(""), &hx, cleanup); st != "" || n != 1 {
			t.Errorf("(c) LOGIN b cleanup: %q %d, want binding b's own (unexpired) row: the reason production runs init(EMPTY) first", st, n)
		}
		if st, n := e5(ptr(""), ptr(""), ptr(""), ptr(""), cleanup+` WHERE expires_at <= now()`); st != "" || n != 0 {
			t.Errorf("(d) EMPTY cleanup with a WHERE clause: %q %d, want nothing deleted", st, n)
		}
		before := count(`SELECT count(*) FROM app.login_transactions`)
		if st, n := e5(ptr(""), ptr(""), ptr(""), &hx, `DELETE FROM app.login_transactions WHERE binding_hash = decode('`+hx+`', 'hex') RETURNING state_hash`); st != "" || n != 1 {
			t.Errorf("(e) LOGIN consume: %q %d, want one row", st, n)
		}
		if st, n := e5(ptr(""), ptr(""), ptr(""), &hy, `DELETE FROM app.login_transactions WHERE binding_hash = decode('`+hx+`', 'hex') RETURNING state_hash`); st != "" || n != 0 {
			t.Errorf("(e) consume under another binding: %q %d", st, n)
		}
		if count(`SELECT count(*) FROM app.login_transactions`) != before {
			t.Error("(a)-(e) changed rows (every control was rolled back)")
		}
	})

	t.Run("RLS-IMMUT lock positives and mutation negatives (AC6)", func(t *testing.T) {
		u, o, g := w.id["gOwnerA1"], w.id["A"], w.id["A1"]
		if _, err := e.setup.Exec(ctx, `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at, absolute_expires_at)
VALUES ($1::uuid, sha256(convert_to($1::text, 'UTF8')), sha256(convert_to($1::text || 'c', 'UTF8')), now(), now() + interval '1 hour', now() + interval '2 hours')
ON CONFLICT DO NOTHING`, u); err != nil {
			t.Fatal(err)
		}
		var session string
		if err := e.setup.QueryRow(ctx, `SELECT id::text FROM app.sessions WHERE user_id = $1 LIMIT 1`, u).Scan(&session); err != nil {
			t.Fatal(err)
		}
		for _, mode := range []struct{ name, group string }{{"Organization-only", ""}, {"selected A1", g}} {
			gp := mode.group
			rlsTx(t, rt, &u, &o, &gp, ptr(""), func(tx pgx.Tx) {
				for _, c := range []struct {
					name string
					sql  string
					want int64
				}{
					{"sqlLockLiveUser", `SELECT id FROM app.users WHERE id = '` + u + `' AND status = 'active' FOR SHARE`, 1},
					{"sqlLockLiveSession", `SELECT id FROM app.sessions WHERE id = '` + session + `' AND user_id = '` + u + `' AND revoked_at IS NULL AND idle_expires_at > now() AND absolute_expires_at > now() FOR SHARE`, 1},
					{"sqlFactOrganizationShare", `SELECT id FROM app.organizations WHERE id = '` + o + `' FOR SHARE`, 1},
					{"sqlFactOrganizationExclusive", `SELECT id FROM app.organizations WHERE id = '` + o + `' FOR NO KEY UPDATE`, 1},
					{"sqlFactOrganizationBindingLock", `SELECT organization_id, role FROM app.role_bindings WHERE organization_id = '` + o + `' AND subject_user_id = '` + u + `' AND scope_type = 'organization' AND source = 'membership' FOR SHARE`, 1},
					{"sqlFactGroupBindingsLock", `SELECT organization_id, group_id, role FROM app.role_bindings WHERE organization_id = '` + o + `' AND subject_user_id = '` + u + `' AND scope_type = 'group' AND source = 'membership' AND group_id = ANY (ARRAY['` + g + `']::uuid[]) ORDER BY group_id FOR SHARE`, 1},
					{"sqlFactAllGroupBindingsLock", `SELECT organization_id, group_id, role FROM app.role_bindings WHERE organization_id = '` + o + `' AND subject_user_id = '` + u + `' AND scope_type = 'group' AND source = 'membership' ORDER BY group_id FOR SHARE`, map[string]int64{"": 2, g: 1}[gp]},
					{"sqlFactTargetGroupLock", `SELECT id FROM app.groups WHERE organization_id = '` + o + `' AND id = '` + g + `' FOR NO KEY UPDATE`, 1},
					{"sqlFactGroupsLock", `SELECT id FROM app.groups WHERE organization_id = '` + o + `' AND id = ANY (ARRAY['` + g + `']::uuid[]) ORDER BY id FOR SHARE`, 1},
				} {
					if st, n := rlsStmt(t, tx, c.sql); st != "" || n != c.want {
						t.Errorf("%s %s: SQLSTATE %q rows %d, want %d", mode.name, c.name, st, n, c.want)
					}
				}
				// r4 (G019, version 15): gOwnerA1's own group_admin self-demotion is admitted by the UPDATE policy and refused by
				// the owner guard as the named last_owner protection (sole owner of A1; oracle r4 'gOwnerA1 selected A1 cannot
				// demote itself as the sole owner of A1'); every other negative keeps its version-6 privilege refusal.
				for _, c := range []struct{ sql, state, constraint string }{
					{`UPDATE app.role_bindings SET role = 'org_admin' WHERE subject_user_id = '` + u + `' AND scope_type = 'organization'`, "42501", ""},
					{`UPDATE app.role_bindings SET role = 'group_admin' WHERE subject_user_id = '` + u + `' AND scope_id = '` + g + `'`, "23000", "last_owner"},
					{`UPDATE app.users SET status = 'disabled', disabled_at = now() WHERE id = '` + u + `'`, "42501", ""},
					{`UPDATE app.sessions SET revoked_at = now(), revoke_reason = 'logout' WHERE id = '` + session + `'`, "42501", ""},
					{`UPDATE app.role_bindings SET subject_user_id = subject_user_id`, "42501", ""},
					{`UPDATE app.role_bindings SET source = source`, "42501", ""},
					{`UPDATE app.role_bindings SET scope_id = scope_id`, "42501", ""},
					{`UPDATE app.groups SET organization_id = organization_id`, "42501", ""},
					{`UPDATE app.business_profiles SET organization_id = organization_id`, "42501", ""},
					{`UPDATE app.group_business_profiles SET business_profile_id = business_profile_id`, "42501", ""},
					{`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, role, source) VALUES ('` + o + `', '` + u + `', 'organization', '` + o + `', 'org_owner', 'membership')`, "42501", ""},
					{`DELETE FROM app.group_business_profiles`, "42501", ""},
					{`UPDATE app.sessions SET user_id = user_id`, "42501", ""},
				} {
					if st, cn, _ := rlsStmtC(t, tx, c.sql); st != c.state || cn != c.constraint {
						t.Errorf("%s %s: SQLSTATE %q constraint %q, want %s %q", mode.name, c.sql, st, cn, c.state, c.constraint)
					}
				}
			})
		}
		// Constraint controls through the setup connection, rolled back: the key-immutability triggers and composite foreign
		// keys hold for every role, independently of row-level security.
		for _, c := range []struct{ sql, state string }{
			{`UPDATE app.groups SET organization_id = '` + w.id["B"] + `' WHERE id = '` + g + `'`, "23000"},
			{`UPDATE app.role_bindings SET subject_user_id = '` + w.id["outsider"] + `' WHERE subject_user_id = '` + u + `' AND scope_id = '` + g + `'`, "23000"},
			{`INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ('` + o + `', '` + w.id["A4"] + `', '` + w.id["pB1"] + `')`, "23503"},
			{`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES ('` + o + `', '` + u + `', 'group', '` + w.id["B1"] + `', '` + w.id["B1"] + `', 'group_owner', 'membership')`, "23503"},
		} {
			tx, err := e.setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			_, err = tx.Exec(ctx, c.sql)
			_ = tx.Rollback(ctx)
			if pgState(err) != c.state {
				t.Errorf("setup constraint control %s: %v, want %s", c.sql, err, c.state)
			}
		}
	})

	t.Run("RLS-PROJ projections and helpers (AC7)", func(t *testing.T) {
		for _, actor := range []string{"ownerA", "gOwnerA1", "adminA2", "memberA", "crossUser"} {
			u, o := w.id[actor], w.id["A"]
			rlsTx(t, rt, &u, &o, ptr(""), ptr(""), func(tx pgx.Tx) {
				for _, c := range []struct{ view, base string }{
					{`SELECT user_id, organization_id, role FROM app.organization_memberships`,
						`SELECT subject_user_id, organization_id, role FROM app.role_bindings WHERE scope_type = 'organization' AND source = 'membership'`},
					{`SELECT user_id, organization_id, group_id, role FROM app.group_memberships`,
						`SELECT subject_user_id, organization_id, group_id, role FROM app.role_bindings WHERE scope_type = 'group' AND source = 'membership'`},
					{`SELECT user_id, organization_id, business_profile_id, group_id, group_role FROM app.business_profile_access_paths`,
						`SELECT rb.subject_user_id, gbp.organization_id, gbp.business_profile_id, gbp.group_id, rb.role FROM app.group_business_profiles gbp
JOIN app.role_bindings rb ON rb.organization_id = gbp.organization_id AND rb.group_id = gbp.group_id WHERE rb.scope_type = 'group' AND rb.source = 'membership'`},
				} {
					var diff int
					if err := tx.QueryRow(ctx, `SELECT count(*) FROM ((`+c.view+` EXCEPT ALL `+c.base+`) UNION ALL (`+c.base+` EXCEPT ALL `+c.view+`)) d`).Scan(&diff); err != nil || diff != 0 {
						t.Errorf("%s: view and base differ by %d rows (%v)", actor, diff, err)
					}
				}
				for _, s := range []struct{ sql, states string }{
					{`INSERT INTO app.organization_memberships (user_id, organization_id, role) VALUES ('` + u + `', '` + o + `', 'org_owner')`, "42501"},
					{`UPDATE app.group_memberships SET role = 'group_owner'`, "42501"},
					{`DELETE FROM app.business_profile_access_paths`, "42501,55000"},
					{`SELECT 1 FROM app.organization_memberships FOR SHARE`, "42501"},
					{`SELECT 1 FROM app.business_profile_access_paths FOR SHARE`, "42501"},
				} {
					if st, _ := rlsStmt(t, tx, s.sql); !strings.Contains(s.states, st) || st == "" {
						t.Errorf("%s %s: SQLSTATE %q, want one of %s", actor, s.sql, st, s.states)
					}
				}
				// The readers return only the caller's own facts.
				var role, other *string
				if err := tx.QueryRow(ctx, `SELECT app.rls_org_role(), app.rls_group_role($1)`, w.id["A2"]).Scan(&role, &other); err != nil {
					t.Fatal(err)
				}
				wantRole := map[string]string{"ownerA": "org_owner", "gOwnerA1": "org_member", "adminA2": "org_member", "memberA": "org_member", "crossUser": "org_member"}[actor]
				if role == nil || *role != wantRole {
					t.Errorf("%s: rls_org_role = %v, want %s", actor, role, wantRole)
				}
				wantA2 := map[string]string{"adminA2": "group_admin"}[actor]
				if (other == nil) != (wantA2 == "") || (other != nil && *other != wantA2) {
					t.Errorf("%s: rls_group_role(A2) = %v, want %q (own role only)", actor, other, wantA2)
				}
			})
		}
	})

	// G020 oracle (rls_expected_test.go "oracle G020"): the label reads of the fixture-only subject Grace and the other
	// label contexts, the raw label writes, and the isolated foreign-key cases with superuser zero-row/zero-audit counts.
	// It runs last: Grace's bindings are added here and removed before it ends, and the restored state is verified.
	t.Run("RLS-LABEL oracle G020 label reads, raw writes and isolated foreign-key cases (T3, AC3, AC5)", func(t *testing.T) {
		orgA, orgB := tenancy.OrganizationID(mustUUID(t, w.id["A"])), tenancy.OrganizationID(mustUUID(t, w.id["B"]))
		before, labelsBefore := e.state(t, orgA, orgB), e.labelState(t, w.id["A"], w.id["B"])
		users, bindings := rlsExpectedLabelSubjects()
		for _, u := range users {
			w.id[u.Label] = rlsUUID(t)
			w.labels[w.id[u.Label]] = u.Label
			if _, err := e.setup.Exec(ctx, `INSERT INTO app.users (id, display_name) VALUES ($1, $2)`, w.id[u.Label], "Oracle "+u.Label); err != nil {
				t.Fatal(err)
			}
		}
		defer func() {
			// Group-scope bindings first, then the Organization binding (no departure cascade is relied on).
			for _, scope := range []string{"group", "organization"} {
				for _, u := range users {
					if _, err := e.setup.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1 AND scope_type = $2`, w.id[u.Label], scope); err != nil {
						t.Errorf("remove the fixture-only %s bindings of %s: %v", scope, u.Label, err)
					}
				}
			}
			if e.state(t, orgA, orgB) != before {
				t.Error("the fixture-only label subjects' bindings were not removed exactly (the oracle fixture changed)")
			}
			if e.labelState(t, w.id["A"], w.id["B"]) != labelsBefore {
				t.Error("the label fixture changed (every RLS-LABEL write is rolled back or restored)")
			}
		}()
		for _, b := range bindings {
			scope, group := w.id[b.Org], any(nil)
			if b.ScopeType == "group" {
				scope, group = w.id[b.Group], w.id[b.Group]
			}
			if _, err := e.setup.Exec(ctx, `INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source)
VALUES ($1, $2, $3, $4, $5, $6, $7)`, w.id[b.Org], w.id[b.User], b.ScopeType, scope, group, b.Role, b.Source); err != nil {
				t.Fatal(err)
			}
		}

		t.Run("reads", func(t *testing.T) {
			seen := map[string]bool{}
			for _, c := range rlsExpectedLabelReads() {
				if seen[c.Name] {
					t.Fatalf("duplicate oracle G020 read %q", c.Name)
				}
				seen[c.Name] = true
				user, org, group, _ := w.shapeSettings(t, c.Context)
				rlsTx(t, rt, user, org, group, nil, func(tx pgx.Tx) {
					if !w.applyOverride(t, tx, c.Context, user, org, group) {
						return
					}
					got, extras, err := w.visible(t, tx)
					if err != nil {
						t.Errorf("%s: read: %v (SQLSTATE %s)", c.Name, err, pgState(err))
						return
					}
					for _, rel := range []string{"label_definitions", "label_assignments"} {
						if extras[rel] > 0 {
							t.Errorf("%s: %d non-fixture rows of %s visible", c.Name, extras[rel], rel)
						}
					}
					if !eqSet(got.LabelDefinitions, c.Definitions) {
						t.Errorf("ORACLE-DISAGREEMENT %s: label_definitions: database %v, oracle %v", c.Name, got.LabelDefinitions, c.Definitions)
					}
					if !eqSet(got.LabelAssignments, c.Assignments) {
						t.Errorf("ORACLE-DISAGREEMENT %s: label_assignments: database %v, oracle %v", c.Name, got.LabelAssignments, c.Assignments)
					}
				})
			}
			fmt.Printf("RLS-LABEL-READ {\"cases\":%d}\n", len(seen))
		})

		// labelSQL builds one raw write; for an assignment INSERT it also returns the attempted label_id and target.
		labelSQL := func(t *testing.T, c rlsLabelWrite, i int, attempt string) (string, []any, string, string) {
			t.Helper()
			ref := func(l string) string {
				if l == "unknown" {
					return rlsUUID(t)
				}
				id, ok := w.id[l]
				if !ok {
					t.Fatalf("%s: unknown oracle label %q", c.Name, l)
				}
				return id
			}
			switch c.Op + " " + c.Table {
			case "insert label_definitions":
				typ, owner, group := "organization", ref(c.Org), any(nil)
				if c.Owner != "" {
					typ, owner, group = "group", ref(c.Owner), ref(c.Owner)
				}
				vis := c.Visibility
				if vis == "" {
					vis = "owner_scope"
				}
				return `INSERT INTO app.label_definitions (id, organization_id, owner_scope_type, owner_scope_id, group_id, name, visibility)
VALUES ($1::uuid, $2::uuid, $3, $4::uuid, $5::uuid, $6, $7)`, []any{attempt, ref(c.Org), typ, owner, group, fmt.Sprintf("Oracle G020 %d", i), vis}, "", ""
			case "insert label_assignments":
				typ, owner := "organization", ref(c.Org)
				if c.Owner != "" {
					typ, owner = "group", ref(c.Owner)
				}
				def, target := ref(c.Def), ref(c.Target)
				return `INSERT INTO app.label_assignments (id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id, target_group_id)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid, 'group', $6::uuid, $6::uuid)`, []any{attempt, ref(c.Org), def, typ, owner, target}, def, target
			case "update label_definitions", "update label_assignments":
				var val string
				switch c.Column {
				case "name":
					val = fmt.Sprintf("'Oracle G020 rename %d'", i)
				case "visibility":
					val = "CASE visibility WHEN 'owner_scope' THEN 'target_members' ELSE 'owner_scope' END"
				case "organization_id":
					val = "'" + w.id["B"] + "'"
				case "group_id", "target_id":
					val = "'" + w.id["A4"] + "'"
				default:
					t.Fatalf("%s: no new value for %s.%s", c.Name, c.Table, c.Column)
				}
				return `UPDATE app.` + c.Table + ` SET ` + c.Column + ` = ` + val + ` WHERE id = $1::uuid`, []any{ref(c.Row)}, "", ""
			case "delete label_definitions", "delete label_assignments":
				return `DELETE FROM app.` + c.Table + ` WHERE id = $1::uuid`, []any{ref(c.Row)}, "", ""
			}
			t.Fatalf("%s: unmapped label write %s %s", c.Name, c.Op, c.Table)
			return "", nil, "", ""
		}
		// judge compares one outcome with the oracle: "ok" exactly one row; "zero" no error and no row; "denied" the pinned
		// SQLSTATE with exactly the named constraint ("" for a privilege or policy refusal) and no row.
		judge := func(t *testing.T, c rlsLabelWrite, st, cn string, n int64) bool {
			t.Helper()
			ok := false
			switch c.Expect {
			case "ok":
				ok = st == "" && n == 1
			case "zero":
				ok = st == "" && n == 0
			case "denied":
				if c.State == "" {
					t.Fatalf("%s: a denied case pins no SQLSTATE", c.Name)
				}
				ok = st == c.State && cn == c.Constraint && n == 0
			default:
				t.Fatalf("%s: unknown oracle outcome %q", c.Name, c.Expect)
			}
			if !ok {
				t.Errorf("ORACLE-DISAGREEMENT %s: SQLSTATE %q constraint %q rows %d, oracle %s %s %s", c.Name, st, cn, n, c.Expect, c.State, c.Constraint)
			}
			return ok
		}
		cases := rlsExpectedLabelWrites()
		seen := map[string]bool{}
		for _, c := range cases {
			if seen[c.Name] {
				t.Fatalf("duplicate oracle G020 write %q", c.Name)
			}
			seen[c.Name] = true
		}

		t.Run("writes", func(t *testing.T) {
			n := 0
			for i, c := range cases {
				if c.Commit {
					continue
				}
				n++
				user, org, group, _ := w.shapeSettings(t, c.Context)
				rlsTx(t, rt, user, org, group, nil, func(tx pgx.Tx) {
					if !w.applyOverride(t, tx, c.Context, user, org, group) {
						return
					}
					if c.Op == "delete_with_assignments" {
						// The facade's two-statement delete (Plan §1.3) in one savepoint: the definition's assignments, then it.
						want := 0
						for _, a := range w.fx.LabelAssignments {
							if a.Def == c.Row {
								want++
							}
						}
						sp, err := tx.Begin(ctx)
						if err != nil {
							t.Fatal(err)
						}
						defer func() { _ = sp.Rollback(ctx) }()
						tag, err := sp.Exec(ctx, `DELETE FROM app.label_assignments WHERE label_id = $1::uuid`, w.id[c.Row])
						if err != nil || tag.RowsAffected() != int64(want) {
							t.Errorf("ORACLE-DISAGREEMENT %s: assignment delete: %v (SQLSTATE %q) rows %d, want %d", c.Name, err, pgState(err), tag.RowsAffected(), want)
							return
						}
						tag, err = sp.Exec(ctx, `DELETE FROM app.label_definitions WHERE id = $1::uuid`, w.id[c.Row])
						judge(t, c, stateOrUnknown(err), pgConstraint(err), tag.RowsAffected())
						return
					}
					sql, args, _, _ := labelSQL(t, c, i, rlsUUID(t))
					st, cn, rows := rlsStmtC(t, tx, sql, args...)
					judge(t, c, st, cn, rows)
				})
			}
			fmt.Printf("RLS-LABEL-WRITE {\"cases\":%d}\n", n)
		})

		t.Run("isolated", func(t *testing.T) {
			count := func(t *testing.T, sql string, args ...any) int {
				t.Helper()
				var n int
				if err := e.setup.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
					t.Fatalf("superuser count: %v (SQLSTATE %s)", err, pgState(err))
				}
				return n
			}
			const (
				byID    = `SELECT count(*) FROM app.label_assignments WHERE id = $1::uuid`
				byKey   = `SELECT count(*) FROM app.label_assignments WHERE label_id = $1::uuid AND target_type = 'group' AND target_id = $2::uuid`
				byDef   = `SELECT count(*) FROM app.label_assignments WHERE label_id = $1::uuid`
				defByID = `SELECT count(*) FROM app.label_definitions WHERE id = $1::uuid`
				audGrp  = `SELECT count(*) FROM audit.events WHERE action = $1 AND resource_type = 'label' AND resource_id = $2::uuid AND group_id = $3::uuid`
				audAny  = `SELECT count(*) FROM audit.events WHERE action = $1 AND resource_type = 'label' AND resource_id = $2::uuid`
			)
			n := 0
			for i, c := range cases {
				if !c.Commit {
					continue
				}
				n++
				if c.Context.Override || c.Context.Shape != "valid" {
					t.Fatalf("%s: an isolated case runs in a valid shape without the override toggle (it commits)", c.Name)
				}
				user, org, group, _ := w.shapeSettings(t, c.Context)
				attempt := rlsUUID(t)
				switch {
				case c.Op == "insert" && c.Table == "label_assignments":
					sql, args, def, target := labelSQL(t, c, i, attempt)
					k0, a0 := count(t, byKey, def, target), count(t, audGrp, "label.assigned", def, target)
					st, cn, rows := labelCommit(t, rt, user, org, group, sql, args...)
					judge(t, c, st, cn, rows)
					want := 0
					if c.Expect == "ok" {
						want = 1
					}
					if got := count(t, byID, attempt); got != want {
						t.Errorf("ORACLE-DISAGREEMENT %s: %d committed row(s) for the attempted key, oracle %d", c.Name, got, want)
					}
					if got := count(t, byKey, def, target) - k0; got != want {
						t.Errorf("ORACLE-DISAGREEMENT %s: %d new (label, target) row(s), oracle %d", c.Name, got, want)
					}
					if got := count(t, audGrp, "label.assigned", def, target) - a0; got != want {
						t.Errorf("ORACLE-DISAGREEMENT %s: %d new label.assigned audit row(s) for the definition and Group, oracle %d", c.Name, got, want)
					}
					if c.Expect == "ok" {
						// Restore: the same subject removes the row it created (committed; its label.unassigned is recorded).
						if st, cn, rows := labelCommit(t, rt, user, org, group, `DELETE FROM app.label_assignments WHERE id = $1::uuid`, attempt); st != "" || rows != 1 {
							t.Errorf("%s: restoring delete: SQLSTATE %q constraint %q rows %d", c.Name, st, cn, rows)
						}
						if got := count(t, byID, attempt); got != 0 {
							t.Errorf("%s: the paired positive row was not removed", c.Name)
						}
					}
				case c.Op == "delete" && c.Table == "label_definitions" && c.Expect == "denied":
					row := w.id[c.Row]
					d0, as0 := count(t, defByID, row), count(t, byDef, row)
					del0, un0 := count(t, audAny, "label.deleted", row), count(t, audAny, "label.unassigned", row)
					sql, args, _, _ := labelSQL(t, c, i, attempt)
					st, cn, rows := labelCommit(t, rt, user, org, group, sql, args...)
					judge(t, c, st, cn, rows)
					if got := count(t, defByID, row); got != 1 || d0 != 1 {
						t.Errorf("ORACLE-DISAGREEMENT %s: the definition has %d row(s) (before %d), oracle 1", c.Name, got, d0)
					}
					if got := count(t, byDef, row); got != as0 || as0 == 0 {
						t.Errorf("ORACLE-DISAGREEMENT %s: %d assignment(s) (before %d), oracle unchanged and nonzero", c.Name, got, as0)
					}
					if got := count(t, audAny, "label.deleted", row) - del0; got != 0 {
						t.Errorf("ORACLE-DISAGREEMENT %s: %d new label.deleted audit row(s), oracle 0", c.Name, got)
					}
					if got := count(t, audAny, "label.unassigned", row) - un0; got != 0 {
						t.Errorf("ORACLE-DISAGREEMENT %s: %d new label.unassigned audit row(s), oracle 0", c.Name, got)
					}
				default:
					t.Fatalf("%s: unmapped isolated case %s %s %s", c.Name, c.Op, c.Table, c.Expect)
				}
			}
			fmt.Printf("RLS-LABEL-ISOLATED {\"cases\":%d}\n", n)
		})
	})

	t.Run("RLS-LABEL-FIXTURE T5 (a)-(g) direct SQL under the test-only fixture widening (AC4, AC5)", func(t *testing.T) {
		e.rlsLabelFixture(t, rt, w)
	})
}

// labelCommit runs one statement under its own savepoint of a committed runtime transaction (G020 T3 isolated protocol):
// a refused statement rolls back to the savepoint and the transaction then commits nothing of it; an admitted statement
// is kept and committed. It returns the SQLSTATE ("" on success), the constraint name of a failure and the row count.
func labelCommit(t *testing.T, c *pgx.Conn, user, org, group *string, sql string, args ...any) (string, string, int64) {
	t.Helper()
	ctx := context.Background()
	tx, err := c.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := rlsSet(ctx, tx, user, org, group, nil); err != nil {
		t.Fatal(err)
	}
	sp, err := tx.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	st, cn, n := "", "", int64(0)
	tag, err := sp.Exec(ctx, sql, args...)
	if err != nil {
		_ = sp.Rollback(ctx)
		st, cn = stateOrUnknown(err), pgConstraint(err)
		if st == "42P17" {
			t.Errorf("42P17 policy recursion: %s", sql)
		}
	} else {
		n = tag.RowsAffected()
		if err := sp.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("isolated case commit: %v (SQLSTATE %s)", err, pgState(err))
	}
	return st, cn, n
}

// rlsLabelFixture is T5's direct-SQL half (G020 Plan r2 §5.2 T5 (a)-(g), §1.4 r2 storage paragraph; AC4, AC5): the
// test-only fixture widening is installed for this subtest only (labels_fixture_test.go) over the oracle world, whose
// subjects play carol (ownerA), bob (crossUser, A3 owner), dave (gOwnerA1, A3 member) and Dana (ownerA4, A4 owner); A3 is
// research and A4 design. Every refused raw INSERT runs under its own savepoint of a committed runtime transaction
// (labelCommit) and is followed by superuser counts proving zero rows for the attempted key and zero new label.assigned
// records; refusal order RLS 42501 -> CHECK 23514 -> unique 23505 -> FK 23503. Superuser statements that must be refused
// run in a rolled-back transaction. The teardown restores the production storage and is byte-compared (g).
func (e *itEnv) rlsLabelFixture(t *testing.T, rt *pgx.Conn, w *rlsOracleWorld) {
	ctx := context.Background()
	orgA := tenancy.OrganizationID(mustUUID(t, w.id["A"]))
	A, A3, A4 := w.id["A"], w.id["A3"], w.id["A4"]
	research, design := tenancy.GroupID(mustUUID(t, A3)), tenancy.GroupID(mustUUID(t, A4))
	carol, bob, dave, dana := w.id["ownerA"], w.id["crossUser"], w.id["gOwnerA1"], w.id["ownerA4"]
	stateBefore, labelsBefore := e.state(t, orgA), e.labelState(t, w.id["A"], w.id["B"])

	adapter, teardown := e.installLabelFixture(t)
	fxHidden := e.fixtureTarget(t, orgA, research, false) // research-owned, managers only
	fxOpen := e.fixtureTarget(t, orgA, research, true)    // research-owned, members may read
	fxD := e.fixtureTarget(t, orgA, design, true)         // design-owned

	const (
		fxInsert = `INSERT INTO app.label_assignments (id, organization_id, label_id, owner_scope_type, owner_scope_id, target_type, target_id, target_group_id)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid, $6, $7::uuid, $8::uuid)`
		byID = `SELECT count(*) FROM app.label_assignments WHERE id = $1::uuid`
	)
	type fxCase struct {
		name, actor, def, ownerType, owner, targetType, target, targetGroup string
		state, constraint                                                   string // "" "" = admitted (one row, one label.assigned)
	}
	run := func(t *testing.T, c fxCase) {
		t.Helper()
		attempt := rlsUUID(t)
		a0 := e.labCount(t, labAuditGrp, "label.assigned", c.def, c.targetGroup)
		st, cn, n := labelCommit(t, rt, ptr(c.actor), ptr(A), ptr(""), fxInsert, attempt, A, c.def, c.ownerType, c.owner, c.targetType, c.target, c.targetGroup)
		want := int64(0)
		if c.state == "" {
			want = 1
		}
		if st != c.state || cn != c.constraint || n != want {
			t.Errorf("ORACLE-DISAGREEMENT %s: SQLSTATE %q constraint %q rows %d, want %q %q %d", c.name, st, cn, n, c.state, c.constraint, want)
		}
		if got := e.labCount(t, byID, attempt); int64(got) != want {
			t.Errorf("ORACLE-DISAGREEMENT %s: %d committed row(s) for the attempted key, want %d", c.name, got, want)
		}
		if got := e.labCount(t, labAuditGrp, "label.assigned", c.def, c.targetGroup) - a0; int64(got) != want {
			t.Errorf("ORACLE-DISAGREEMENT %s: %d new label.assigned record(s), want %d", c.name, got, want)
		}
		if want == 1 {
			if st, cn, rows := labelCommit(t, rt, ptr(c.actor), ptr(A), ptr(""), `DELETE FROM app.label_assignments WHERE id = $1::uuid`, attempt); st != "" || rows != 1 {
				t.Errorf("%s: restoring delete: SQLSTATE %q constraint %q rows %d", c.name, st, cn, rows)
			}
		}
	}
	super := func(t *testing.T, sql string, args ...any) (string, string) {
		t.Helper()
		tx, err := e.setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		_, err = tx.Exec(ctx, sql, args...)
		return stateOrUnknown(err), pgConstraint(err)
	}
	// resolveAs runs the fixture adapter's resolve (and readable on the read facts) in a runtime transaction of user.
	resolveAs := func(t *testing.T, user, fx string) (owning tenancy.GroupID, ok, readable bool) {
		t.Helper()
		rlsTx(t, rt, ptr(user), ptr(A), ptr(""), nil, func(tx pgx.Tx) {
			f, err := tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: orgA, User: tenancy.UserID(mustUUID(t, user))})
			if err != nil {
				t.Fatalf("read facts: %v", err)
			}
			g, found, err := adapter.resolve(ctx, &trackedTx{Tx: tx}, f, orgA, mustUUID(t, fx))
			if err != nil {
				t.Fatalf("resolve: %v (SQLSTATE %s)", err, pgState(err))
			}
			owning, ok = g, found
			readable = found && adapter.readable(contextSelection{Organization: orgA}, f, g)
		})
		return owning, ok, readable
	}

	// Two superuser fixture assignments of research's target_members definition (exact copied owner facts).
	asHidden, asOpen := rlsUUID(t), rlsUUID(t)
	for _, a := range []struct{ id, fx string }{{asHidden, fxHidden}, {asOpen, fxOpen}} {
		if _, err := e.setup.Exec(ctx, fxInsert, a.id, A, w.id["lA3Tm"], "group", A3, "fixture", a.fx, A3); err != nil {
			t.Fatalf("fixture assignment: %v (SQLSTATE %s)", err, pgState(err))
		}
	}

	t.Run("(a) unreadable target: a member reads no row; core readable agrees with the raw SELECT", func(t *testing.T) {
		for _, s := range []struct {
			name, user   string
			hidden, open int64
		}{{"bob", bob, 1, 1}, {"dave", dave, 0, 1}, {"carol", carol, 0, 0}, {"dana", dana, 0, 0}} {
			for _, a := range []struct {
				id, fx string
				want   int64
			}{{asHidden, fxHidden, s.hidden}, {asOpen, fxOpen, s.open}} {
				var n int64
				rlsTx(t, rt, ptr(s.user), ptr(A), ptr(""), nil, func(tx pgx.Tx) {
					var st string
					st, n = rlsStmt(t, tx, `SELECT id FROM app.label_assignments WHERE id = $1::uuid`, a.id)
					if st != "" {
						t.Errorf("%s raw SELECT: SQLSTATE %s", s.name, st)
					}
				})
				if n != a.want {
					t.Errorf("ORACLE-DISAGREEMENT %s reads %d row(s) of the assignment on %s, want %d", s.name, n, a.fx, a.want)
				}
				owning, ok, readable := resolveAs(t, s.user, a.fx)
				if ok && owning != research {
					t.Errorf("%s: resolve returned another owning Group", s.name)
				}
				if readable != (n == 1) {
					t.Errorf("ORACLE-DISAGREEMENT parity %s: core readable %t, raw SELECT %d row(s)", s.name, readable, n)
				}
			}
		}
	})

	t.Run("(b) nonexistent target: the fixture key refuses; an assigned target cannot be deleted; paired positive", func(t *testing.T) {
		run(t, fxCase{name: "bob unknown fixture target", actor: bob, def: w.id["lA3Own"], ownerType: "group", owner: A3, targetType: "fixture",
			target: rlsUUID(t), targetGroup: A3, state: "23503", constraint: "label_assignments_fixture_target_fkey"})
		run(t, fxCase{name: "bob live research fixture target (paired positive)", actor: bob, def: w.id["lA3Own"], ownerType: "group", owner: A3,
			targetType: "fixture", target: fxOpen, targetGroup: A3})
		if st, cn := super(t, `DELETE FROM app.label_fixture_targets WHERE id = $1::uuid`, fxHidden); st != "23503" || cn != "label_assignments_fixture_target_fkey" {
			t.Errorf("ORACLE-DISAGREEMENT superuser delete of an assigned fixture target: SQLSTATE %q constraint %q, want 23503 label_assignments_fixture_target_fkey", st, cn)
		}
		if e.labCount(t, `SELECT count(*) FROM app.label_fixture_targets WHERE id = $1::uuid`, fxHidden) != 1 {
			t.Error("the assigned fixture target was removed")
		}
	})

	t.Run("(d) wrong Group under direct SQL: a dual-Group subject is refused by the fixture key; CHECK and policy variants", func(t *testing.T) {
		danaU := &itUser{id: tenancy.UserID(mustUUID(t, dana))}
		e.grantGroup(t, orgA, research, danaU, tenancy.GroupAdmin)
		defer func() {
			if _, err := e.setup.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1::uuid AND scope_type = 'group' AND scope_id = $2::uuid`,
				dana, A3); err != nil {
				t.Errorf("remove Dana's per-case research role: %v", err)
			}
		}()
		run(t, fxCase{name: "Dana (design owner, research admin) design target with research facts", actor: dana, def: w.id["lA3Own"],
			ownerType: "group", owner: A3, targetType: "fixture", target: fxD, targetGroup: A3, state: "23503", constraint: "label_assignments_fixture_target_fkey"})
		run(t, fxCase{name: "Dana design target with target_group_id design", actor: dana, def: w.id["lA3Own"], ownerType: "group", owner: A3,
			targetType: "fixture", target: fxD, targetGroup: A4, state: "23514", constraint: "label_assignments_target_shape"})
		run(t, fxCase{name: "bob (research only) design target with target_group_id design", actor: bob, def: w.id["lA3Own"], ownerType: "group",
			owner: A3, targetType: "fixture", target: fxD, targetGroup: A4, state: "42501", constraint: ""})
		if st, cn := super(t, `UPDATE app.label_fixture_targets SET group_id = $2::uuid WHERE id = $1::uuid`, fxHidden, A4); st != "23503" ||
			cn != "label_assignments_fixture_target_fkey" {
			t.Errorf("ORACLE-DISAGREEMENT superuser move of an assigned fixture target: SQLSTATE %q constraint %q, want 23503 label_assignments_fixture_target_fkey", st, cn)
		}
		// The core half of the same placement (LAB-13 asserts the facade's not_found): resolve answers the actual owner.
		if owning, ok, _ := resolveAs(t, dana, fxD); !ok || owning != design || owning == research {
			t.Errorf("Dana's resolve of the design target: ok %t, design %t", ok, owning == design)
		}
		if _, ok, _ := resolveAs(t, bob, fxD); ok {
			t.Error("bob (no design role) resolved the design target")
		}
	})

	t.Run("(e) production rows coexist with fixture rows; the production branches still refuse", func(t *testing.T) {
		attempt := rlsUUID(t)
		st, cn, n := labelCommit(t, rt, ptr(carol), ptr(A), ptr(""), fxInsert, attempt, A, w.id["lOrgA2"], "organization", A, "group", A3, A3)
		if st != "" || n != 1 {
			t.Fatalf("ORACLE-DISAGREEMENT carol production-shape insert: SQLSTATE %q constraint %q rows %d", st, cn, n)
		}
		if e.labCount(t, `SELECT count(*) FROM app.label_assignments WHERE id = $1::uuid AND target_fixture_id IS NULL`, attempt) != 1 {
			t.Error("the production-shape row carries a fixture key")
		}
		rlsTx(t, rt, ptr(carol), ptr(A), ptr(""), nil, func(tx pgx.Tx) {
			if st, n := rlsStmt(t, tx, `SELECT id FROM app.label_assignments WHERE id = $1::uuid`, attempt); st != "" || n != 1 {
				t.Errorf("carol reads her production-shape row: SQLSTATE %q rows %d", st, n)
			}
		})
		if st, cn, rows := labelCommit(t, rt, ptr(carol), ptr(A), ptr(""), `DELETE FROM app.label_assignments WHERE id = $1::uuid`, attempt); st != "" || rows != 1 {
			t.Errorf("restoring delete: SQLSTATE %q constraint %q rows %d", st, cn, rows)
		}
		unknown := rlsUUID(t)
		run(t, fxCase{name: "carol unknown Group target", actor: carol, def: w.id["lOrgA2"], ownerType: "organization", owner: A, targetType: "group",
			target: unknown, targetGroup: unknown, state: "23503", constraint: "label_assignments_target_group_fkey"})
		// Organization owner facts with the fixture kind: the INSERT policy's Organization branch requires target_type
		// 'group', so the runtime's first refusal is RLS; the CHECK is observed by the superuser (no RLS).
		run(t, fxCase{name: "carol fixture kind with Organization owner facts", actor: carol, def: w.id["lOrgA2"], ownerType: "organization", owner: A,
			targetType: "fixture", target: fxOpen, targetGroup: A3, state: "42501", constraint: ""})
		if st, cn := super(t, fxInsert, rlsUUID(t), A, w.id["lOrgA2"], "organization", A, "fixture", fxOpen, A3); st != "23514" || cn != "label_assignments_target_shape" {
			t.Errorf("ORACLE-DISAGREEMENT superuser fixture kind with Organization owner facts: SQLSTATE %q constraint %q, want 23514 label_assignments_target_shape", st, cn)
		}
		if got := e.labCount(t, `SELECT count(*) FROM app.label_fixture_targets t JOIN app.groups g ON g.id = t.id`); got != 0 {
			t.Errorf("%d fixture target(s) stand in for a Group", got)
		}
	})

	t.Run("(f) isolated FK, Organization->Group direction: an Organization-owned id with Group facts", func(t *testing.T) {
		run(t, fxCase{name: "bob Organization-owned definition with research facts on a research fixture target", actor: bob, def: w.id["lOrgA"],
			ownerType: "group", owner: A3, targetType: "fixture", target: fxOpen, targetGroup: A3, state: "23503", constraint: "label_assignments_label_fkey"})
	})

	t.Run("(g) teardown: the fixture objects are gone and the production bodies and columns are restored byte for byte", func(t *testing.T) {
		tag, err := e.setup.Exec(ctx, `DELETE FROM app.label_assignments WHERE id IN ($1::uuid, $2::uuid)`, asHidden, asOpen)
		if err != nil || tag.RowsAffected() != 2 {
			t.Fatalf("remove the fixture assignments: %v rows %d", err, tag.RowsAffected())
		}
		teardown(t) // asserts the restored CHECK/SELECT bodies, column list and the absence of every fixture object
		if e.labelState(t, w.id["A"], w.id["B"]) != labelsBefore {
			t.Error("the label fixture changed")
		}
		if e.state(t, orgA) != stateBefore {
			t.Error("Organization A's rows changed (Dana's per-case research role was not removed exactly)")
		}
	})
}

// labelState hashes every label definition and assignment row of the given Organizations (superuser; G020).
func (e *itEnv) labelState(t *testing.T, orgs ...string) string {
	t.Helper()
	var s string
	if err := e.setup.QueryRow(context.Background(), `
SELECT coalesce((SELECT md5(string_agg(d::text, E'\n' ORDER BY d.id)) FROM app.label_definitions d WHERE d.organization_id::text = ANY ($1::text[])), '-') || '|' ||
       coalesce((SELECT md5(string_agg(a::text, E'\n' ORDER BY a.id)) FROM app.label_assignments a WHERE a.organization_id::text = ANY ($1::text[])), '-')`,
		orgs).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func mustUUID(t *testing.T, s string) [16]byte {
	t.Helper()
	id, err := tenancy.ParseOrganizationID(s)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// --- TestRLSPool ---------------------------------------------------------------------------------------------------------

// TestRLSPool: P1-P10 through the guarded facade on a runtime pool constrained to one observed backend. Context
// settings are read on that same connection after each operation; a discard is proven by a new backend PID.
func TestRLSPool(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = e.cfg.Host, uint16(e.cfg.Port), e.cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = e.cfg.RuntimeRole(), e.pw
	pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
	pc.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := e.rlsAttest(pool); err != nil {
		t.Fatal(err)
	}
	hook := &hookDB{pool: pool}
	svc := NewService(hook, e.id, nil)
	w := e.world(t)
	// G060: the owner holds no Group role, so its g3 renames (P3, P7, P8) are Organization-path writes that need the
	// Group-management override (explicit setup through the fixture service, before the pool's first acquisition).
	e.withOverride(t, w.owner, w.A)
	backend := func() int {
		var pid int
		if err := pool.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
			t.Fatal(err)
		}
		return pid
	}
	settings := func() string {
		var s string
		if err := pool.QueryRow(ctx, `SELECT concat_ws('|', coalesce(current_setting('app.user_id', true), ''), coalesce(current_setting('app.organization_id', true), ''),
  coalesce(current_setting('app.group_id', true), ''), coalesce(current_setting('app.login_binding', true), ''))`).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	record := func(c, outcome string) {
		fmt.Printf("RLS-POOL {\"case\":%q,\"outcome\":%q}\n", c, outcome)
	}
	rctxOwner, _ := beginAs(t, svc, ctx, w.owner, w.A)
	rctxForeign, _ := beginAs(t, svc, ctx, w.foreign, w.B)
	zeroG := tenancy.GroupID{}

	t.Run("P1 Organization A, context-free access, Organization B on one backend", func(t *testing.T) {
		pid := backend()
		tcA := resolve(t, svc, rctxOwner, w.A, zeroG)
		if _, err := svc.GetOrganization(rctxOwner, tcA); err != nil {
			t.Fatal(err)
		}
		if settings() != "|||" || backend() != pid {
			t.Fatalf("settings after A: %q", settings())
		}
		var n int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM app.organizations`).Scan(&n); err != nil || n != 0 {
			t.Fatalf("context-free access after A sees %d Organizations", n)
		}
		tcB := resolve(t, svc, rctxForeign, w.B, zeroG)
		o, err := svc.GetOrganization(rctxForeign, tcB)
		if err != nil || o.ID != w.B {
			t.Fatalf("B: %v", err)
		}
		if _, err := svc.GetOrganization(rctxForeign, resolveOrNil(svc, rctxForeign, w.A)); err == nil {
			t.Fatal("the B subject read Organization A")
		}
		if backend() != pid {
			t.Fatal("backend changed without a discard reason")
		}
		record("P1", "same backend, no carried authority")
	})
	t.Run("P2 Group-selected then Organization-only clears the Group", func(t *testing.T) {
		pid := backend()
		tcG := resolve(t, svc, rctxOwner, w.A, w.g1.ID)
		if _, err := svc.GetGroup(rctxOwner, tcG, w.g1.ID); err != nil {
			t.Fatal(err)
		}
		tcO := resolve(t, svc, rctxOwner, w.A, zeroG)
		if _, err := svc.GetGroup(rctxOwner, tcO, w.g2.ID); err != nil {
			t.Fatalf("Organization-only read of g2 after a g1 selection: %v", err)
		}
		if settings() != "|||" || backend() != pid {
			t.Fatal("not a clean reuse")
		}
		record("P2", "same backend, Group cleared")
	})
	t.Run("P3/P4/P5 commit, rollback (denial) and statement error leave a clean connection", func(t *testing.T) {
		pid := backend()
		tc := resolve(t, svc, rctxOwner, w.A, zeroG)
		if _, err := svc.RenameGroup(rctxOwner, tc, w.g3.ID, w.g3.Name); err != nil {
			t.Fatalf("commit: %v", err)
		}
		if settings() != "|||" || backend() != pid {
			t.Fatal("P3 not clean")
		}
		rctxMember, _ := beginAs(t, svc, ctx, w.member, w.A)
		tcm := resolve(t, svc, rctxMember, w.A, zeroG)
		if _, err := svc.RenameGroup(rctxMember, tcm, w.g1.ID, "Denied"); err == nil {
			t.Fatal("an Organization member renamed a Group")
		}
		if settings() != "|||" || backend() != pid {
			t.Fatal("P4 (rolled back denial) not clean")
		}
		hook.reset()
		hook.failOn = "O"
		_, err := svc.GetOrganization(rctxOwner, tc)
		hook.reset()
		wantCode(t, "statement error", err, CodeUnavailable)
		got := backend()
		if got == pid {
			if settings() != "|||" {
				t.Fatal("P5 reused with settings")
			}
			record("P5", "clean reuse")
		} else {
			record("P5", "discarded")
		}
	})
	t.Run("P6 a context-initialization failure does no protected work and discards", func(t *testing.T) {
		pool.Reset()
		c, err := pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := c.Exec(ctx, `SELECT set_config('app.user_id', $1, false)`, w.foreign.id.String()); err != nil {
			t.Fatal(err)
		}
		poisoned := int(c.Conn().PgConn().PID())
		// The unsafe control: a raw transaction on the poisoned connection without the boundary's guard sees the foreign
		// subject's self row, and a local set_config would revert to the leaked value.
		var leaked int
		if err := c.QueryRow(ctx, `SELECT count(*) FROM app.users`).Scan(&leaked); err != nil || leaked != 1 {
			t.Fatalf("unsafe control: the session-level setting did not leak (%d, %v)", leaked, err)
		}
		c.Release()
		hook.reset()
		hook.capture = true
		tc := resolve2(t, svc, rctxOwner, w.A)
		stmts := hook.capturedStmts()
		hook.reset()
		if tc != nil {
			t.Fatal("Resolve on a poisoned connection succeeded")
		}
		if len(stmts) != 1 || stmts[0].step != "C" {
			t.Fatalf("statements after the refused initialization: %d (protected work ran)", len(stmts))
		}
		if backend() == poisoned {
			t.Fatal("the poisoned connection was not discarded")
		}
		if settings() != "|||" {
			t.Fatal("the new connection carries settings")
		}
		record("P6/P9", "refused at init, discarded")
	})
	t.Run("P7 cancellation", func(t *testing.T) {
		pid := backend()
		tc := resolve(t, svc, rctxOwner, w.A, zeroG)
		cctx, cancel := context.WithCancel(rctxOwner)
		hook.reset()
		hook.on("after", "O", cancel)
		_, err := svc.RenameGroup(cctx, tc, w.g3.ID, "Canceled")
		hook.reset()
		cancel()
		wantCode(t, "cancellation", err, CodeUnavailable)
		if got := backend(); got != pid {
			record("P7", "discarded")
		} else if settings() != "|||" {
			t.Fatal("canceled connection reused with settings")
		} else {
			record("P7", "clean reuse")
		}
	})
	// P8 uses the preserved G007 synthetic hook: the COMMIT really completes and is acknowledged on a clean connection,
	// then the hook manufactures io.ErrUnexpectedEOF. It covers the conservative application-outcome classification and
	// clean settings on reuse, not the discard of a connection with real transport loss (identity I3/I5 and the poison
	// paths cover that).
	t.Run("P8 synthetic EOF after an acknowledged COMMIT: unavailable, outcome recorded, clean settings on reuse", func(t *testing.T) {
		tc := resolve(t, svc, rctxOwner, w.A, zeroG)
		hook.reset()
		hook.commit = commitLostReply
		_, err := svc.RenameGroup(rctxOwner, tc, w.g3.ID, "Lost Reply")
		hook.reset()
		wantCode(t, "lost reply", err, CodeUnavailable)
		var name string
		if err := e.setup.QueryRow(ctx, `SELECT name FROM app.groups WHERE id = $1`, w.g3.ID.String()).Scan(&name); err != nil {
			t.Fatal(err)
		}
		if settings() != "|||" {
			t.Fatal("reused with settings")
		}
		record("P8", "committed="+fmt.Sprint(name == "Lost Reply"))
	})
	t.Run("P10 release inside a transaction is destroyed by the pool", func(t *testing.T) {
		c, err := pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		pid := int(c.Conn().PgConn().PID())
		if _, err := c.Exec(ctx, `BEGIN; SELECT set_config('app.user_id', $1, true)`, w.owner.id.String()); err != nil {
			// multi-statement with a parameter is not allowed; fall back to two statements
			if _, err := c.Exec(ctx, `BEGIN`); err != nil {
				t.Fatal(err)
			}
			if _, err := c.Exec(ctx, `SELECT set_config('app.user_id', $1, true)`, w.owner.id.String()); err != nil {
				t.Fatal(err)
			}
		}
		c.Release()
		if backend() == pid {
			t.Fatal("a connection released inside a transaction was reused")
		}
		if settings() != "|||" {
			t.Fatal("settings survived")
		}
		record("P10", "destroyed on release")
	})
}

// resolveOrNil resolves an Organization context, or returns nil when it is denied.
func resolveOrNil(svc *Service, rctx context.Context, org tenancy.OrganizationID) *TenantContext {
	tc, err := svc.Resolve(rctx, ResolveRequest{OrganizationID: org.String()})
	if err != nil {
		return nil
	}
	return tc
}

func resolve2(t *testing.T, svc *Service, rctx context.Context, org tenancy.OrganizationID) *TenantContext {
	t.Helper()
	tc, err := svc.Resolve(rctx, ResolveRequest{OrganizationID: org.String()})
	if err == nil {
		return tc
	}
	wantCode(t, "poisoned resolve", err, CodeUnavailable)
	return nil
}

// --- TestRLSFootprint ----------------------------------------------------------------------------------------------------

// TestRLSFootprint (E-RLS-1..3, Plan r3 §6; PG-002): bounded characterization on TestAccessFootprint's dataset, no
// benchmark or target. E-RLS-1: EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) summaries of every statement of the guarded
// writes and reads in both context modes, replayed in their own context, plus F1 (E1) and an F5 write. E-RLS-2: the
// locked-row inventory of the guarded writes (E-PG-2 with no RLS-hidden row for these authorized subjects). E-RLS-3:
// statements per operation and reader/parser calls per operation and per statement, counted in the replaying
// transaction as differences of pg_stat_get_xact_function_calls (track_functions = 'all' set on the disposable database
// by the setup connection). On the verified PostgreSQL 17 implementation the function reads the backend's pending, not
// yet flushed, entry, which can include earlier transactions on the same connection, so only differences taken inside
// the replaying transaction are used (a test-measurement substitution, Codex review of escalation 03).
// Alert lines are printed as "E-RLS-TRIGGER" for review: measured observations, never a pass condition. A PASS of this
// test means the collection succeeded, not that the measured cost is accepted.
func TestRLSFootprint(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	if _, err := e.setup.Exec(ctx, `ALTER DATABASE `+pgx.Identifier{e.cfg.Database}.Sanitize()+` SET track_functions = 'all'`); err != nil {
		t.Fatal(err)
	}
	pool := e.runtimePool(t, nil) // new connections see track_functions = 'all'
	fw := e.footprintWorld(t)
	var serverVersion string
	if err := pool.QueryRow(ctx, `SELECT current_setting('server_version')`).Scan(&serverVersion); err != nil {
		t.Fatal(err)
	}
	env := map[string]any{"server_version": serverVersion, "runner": "harness check container", "dataset": "TestAccessFootprint"}
	h := &hookDB{pool: pool}
	svc := NewService(h, e.id, nil)
	type scenario struct {
		subject string
		u       *itUser
		group   tenancy.GroupID
		profile tenancy.BusinessProfileID
	}
	scenarios := []scenario{{"s1", fw.s1, fw.s1Group, fw.s1Profile}, {"s50", fw.s50, fw.s50Group, fw.s50Profile}}
	for _, sc := range scenarios {
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
			ops := []struct {
				op   string
				call func() error
			}{
				{opUpdateBusinessProfile, func() error { _, err := svc.UpdateBusinessProfile(rctx, tc, sc.profile, p.Context); return err }},
				{opRenameGroup, func() error { _, err := svc.RenameGroup(rctx, tc, sc.group, g.Name); return err }},
				{opChangeGroupSlug, func() error { _, err := svc.ChangeGroupSlug(rctx, tc, sc.group, g.Slug); return err }},
				{opGetBusinessProfile, func() error { _, err := svc.GetBusinessProfile(rctx, tc, sc.profile); return err }},
				{opGetGroup, func() error { _, err := svc.GetGroup(rctx, tc, sc.group); return err }},
				{opGetOrganization, func() error { _, err := svc.GetOrganization(rctx, tc); return err }},
			}
			for _, o := range ops {
				h.reset()
				h.capture = true
				if err := o.call(); err != nil {
					t.Fatalf("%s %s %s: %v", sc.subject, mode, o.op, err)
				}
				stmts := h.capturedStmts()
				locked := h.lockedRows()
				h.reset()
				if len(stmts) == 0 || stmts[0].step != "C" {
					t.Fatalf("%s %s %s: no leading context statement", sc.subject, mode, o.op)
				}
				calls := e.rlsReplay(t, pool, env, sc.subject, mode, o.op, stmts)
				lockedByStep := map[string]int{}
				for _, l := range locked {
					lockedByStep[l.step] += l.rows
				}
				line, _ := json.Marshal(map[string]any{"evidence": "E-RLS-2", "subject": sc.subject, "mode": mode, "operation": o.op, "locked_rows": lockedByStep})
				fmt.Println("E-RLS-2 " + string(line))
				line, _ = json.Marshal(map[string]any{"evidence": "E-RLS-3", "subject": sc.subject, "mode": mode, "operation": o.op,
					"statements": len(stmts), "function_calls": calls, "n_g": map[string]int{"s1": 1, "s50": 50}[sc.subject]})
				fmt.Println("E-RLS-3 " + string(line))
			}
		}
	}
	// F1 (E1) and an F5 write (a revocation in IDENTITY shape), replayed for their plans.
	var cred []byte
	if err := e.setup.QueryRow(ctx, `SELECT credential_hash FROM app.sessions WHERE user_id = $1 AND revoked_at IS NULL LIMIT 1`, fw.s1.id.String()).Scan(&cred); err != nil {
		t.Fatal(err)
	}
	ident := capturedStmt{step: "C", sql: `SELECT pg_catalog.set_config('app.user_id', $1, true)`, args: []any{fw.s1.id.String()}}
	e.rlsReplay(t, pool, env, "s1", "identity", "F1 Session (E1)", []capturedStmt{{step: "C", sql: `SELECT 1`}, {step: "E1",
		sql: `SELECT session_id, user_id FROM app.auth_resolve_session($1, now(), $2)`, args: []any{cred, itOrigin}}})
	e.rlsReplay(t, pool, env, "s1", "identity", "F5 logout revoke", []capturedStmt{ident, {step: "W",
		sql:  `UPDATE app.sessions SET revoked_at = now(), revoke_reason = 'logout' WHERE id = (SELECT id FROM app.sessions WHERE user_id = $1 LIMIT 1) AND revoked_at IS NULL`,
		args: []any{fw.s1.id.String()}}})
	if _, err := e.setup.Exec(ctx, `ALTER DATABASE `+pgx.Identifier{e.cfg.Database}.Sanitize()+` RESET track_functions`); err != nil {
		t.Fatal(err)
	}
}

// rlsReplay replays one operation's captured statements in one rolled-back runtime transaction: the context statement
// first, then each statement under EXPLAIN (ANALYZE, VERBOSE, BUFFERS, FORMAT JSON), printing one E-RLS-1 summary per
// statement with its reader and parser calls and the rows its scans examined, one E-RLS-SUBPLAN line per SubPlan, and
// E-RLS-TRIGGER alert lines: reader calls beyond twice the rows all of the statement's scans examined (a descriptive
// all-scan heuristic: the Plan's candidate-row term does not define this denominator, scans can repeat relations and
// nest policy work, and it is neither an approved replacement threshold nor a pass condition); a SubPlan that scans no
// relation yet runs more than once (a per-row parser or R expression); and a SubPlan with a sequential scan that ran
// (for review against PG-001's known small-fixture scans). It returns the replaying transaction's calls of each app
// function (differences, see TestRLSFootprint).
func (e *itEnv) rlsReplay(t *testing.T, pool *pgxpool.Pool, env map[string]any, subject, mode, op string, stmts []capturedStmt) map[string]int64 {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, stmts[0].sql, stmts[0].args...); err != nil {
		t.Fatalf("replay context: %v", err)
	}
	counts := func() map[string]int64 {
		rows, err := tx.Query(ctx, `SELECT p.proname::text, coalesce(pg_stat_get_xact_function_calls(p.oid), 0) FROM pg_proc p WHERE p.pronamespace = 'app'::regnamespace`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		m := map[string]int64{}
		for rows.Next() {
			var n string
			var c int64
			if err := rows.Scan(&n, &c); err != nil {
				t.Fatal(err)
			}
			m[n] = c
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return m
	}
	diff := func(after, before map[string]int64) map[string]int64 {
		d := map[string]int64{}
		for k, v := range after {
			if v-before[k] > 0 {
				d[k] = v - before[k]
			}
		}
		return d
	}
	trigger := func(fields map[string]any) {
		fields["evidence"], fields["subject"], fields["mode"], fields["operation"] = "E-RLS-TRIGGER", subject, mode, op
		line, _ := json.Marshal(fields)
		fmt.Println("E-RLS-TRIGGER " + string(line))
	}
	start := counts()
	for i, st := range stmts[1:] {
		pre := counts()
		var raw []byte
		if err := tx.QueryRow(ctx, "EXPLAIN (ANALYZE, VERBOSE, BUFFERS, FORMAT JSON) "+st.sql, st.args...).Scan(&raw); err != nil {
			t.Fatalf("explain %s %s: %v", op, st.step, err)
		}
		d := diff(counts(), pre)
		var doc []map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil || len(doc) != 1 {
			t.Fatalf("explain: %v", err)
		}
		plan, _ := doc[0]["Plan"].(map[string]any)
		sum := summarize(plan)
		subs := subPlanInventory(plan)
		readers := d["rls_org_role"] + d["rls_group_role"]
		parsers := d["ctx_user"] + d["ctx_org"] + d["ctx_group"] + d["ctx_login_binding"]
		scanned := scannedRows(plan)
		line, _ := json.Marshal(map[string]any{"evidence": "E-RLS-1", "subject": subject, "mode": mode, "operation": op, "index": i + 1,
			"step": st.step, "plan": sum, "subplans": len(subs), "reader_calls": readers, "parser_calls": parsers, "scanned_rows": scanned,
			"planning_ms": doc[0]["Planning Time"], "execution_ms": doc[0]["Execution Time"], "environment": env})
		fmt.Println("E-RLS-1 " + string(line))
		if float64(readers) > 2*max(scanned, 1) {
			trigger(map[string]any{"step": st.step, "index": i + 1, "reason": "reader calls exceed twice the rows all of the statement's scans examined (descriptive heuristic, not the Plan's candidate-row term; unresolved observation)",
				"reader_calls": readers, "scanned_rows": scanned})
		}
		for _, sp := range subs {
			line, _ := json.Marshal(map[string]any{"evidence": "E-RLS-SUBPLAN", "subject": subject, "mode": mode, "operation": op, "step": st.step,
				"index": i + 1, "subplan": sp})
			fmt.Println("E-RLS-SUBPLAN " + string(line))
			if len(sp.Relations) == 0 && sp.Loops > 1 {
				trigger(map[string]any{"step": st.step, "index": i + 1, "reason": "a relation-less SubPlan evaluated per row (an uncorrelated parser or R expression)",
					"subplan": sp.Name, "loops": sp.Loops})
			}
			if sp.SeqScan && sp.Loops > 0 {
				trigger(map[string]any{"step": st.step, "index": i + 1, "reason": "a policy SubPlan with a sequential scan that ran (review against PG-001's known small-fixture scans)",
					"subplan": sp.Name, "loops": sp.Loops, "relations": sp.Relations})
			}
		}
	}
	return diff(counts(), start)
}

// subPlan is one SubPlan node of a plan (an InitPlan, evaluated once, is not one): its name, how often it ran, the
// relations its subtree scans, whether any of those scans is sequential, and the rows they examined.
type subPlan struct {
	Name      string   `json:"name"`
	Loops     float64  `json:"loops"`
	Relations []string `json:"relations"`
	SeqScan   bool     `json:"seq_scan"`
	Examined  float64  `json:"examined_rows"`
}

// subPlanInventory lists the plan's SubPlan nodes, outermost first.
func subPlanInventory(n map[string]any) []subPlan {
	var out []subPlan
	if r, _ := n["Parent Relationship"].(string); r == "SubPlan" {
		sp := subPlan{Relations: []string{}}
		sp.Name, _ = n["Subplan Name"].(string)
		sp.Loops, _ = n["Actual Loops"].(float64)
		var walk func(m map[string]any)
		walk = func(m map[string]any) {
			if rel, _ := m["Relation Name"].(string); rel != "" && !slices.Contains(sp.Relations, rel) {
				sp.Relations = append(sp.Relations, rel)
			}
			if nt, _ := m["Node Type"].(string); nt == "Seq Scan" {
				sp.SeqScan = true
			}
			sp.Examined += nodeExamined(m)
			if kids, ok := m["Plans"].([]any); ok {
				for _, k := range kids {
					if km, ok := k.(map[string]any); ok {
						walk(km)
					}
				}
			}
		}
		walk(n)
		sort.Strings(sp.Relations)
		out = append(out, sp)
	}
	if kids, ok := n["Plans"].([]any); ok {
		for _, k := range kids {
			if m, ok := k.(map[string]any); ok {
				out = append(out, subPlanInventory(m)...)
			}
		}
	}
	return out
}

// nodeExamined is the number of rows one scan node examined (returned plus removed by filter or recheck, over all of
// its loops); zero for a node that is not a relation scan.
func nodeExamined(m map[string]any) float64 {
	switch nt, _ := m["Node Type"].(string); nt {
	case "Seq Scan", "Index Scan", "Index Only Scan", "Bitmap Heap Scan":
		num := func(k string) float64 { v, _ := m[k].(float64); return v }
		return (num("Actual Rows") + num("Rows Removed by Filter") + num("Rows Removed by Index Recheck")) * max(num("Actual Loops"), 1)
	}
	return 0
}

// scannedRows is the number of rows every relation scan of the plan examined, repeated relations and nested policy
// scans included: a descriptive figure, not an attribution of policy candidate rows.
func scannedRows(n map[string]any) float64 {
	c := nodeExamined(n)
	if kids, ok := n["Plans"].([]any); ok {
		for _, k := range kids {
			if m, ok := k.(map[string]any); ok {
				c += scannedRows(m)
			}
		}
	}
	return c
}

var _ = os.Getenv
var _ = time.Second
