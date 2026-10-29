//go:build integration

package access

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/tenancy"
)

// G020 Plan r2 §1.4 "Test-only fixture adapter and schema widening" (T5, AC4): the extension TEMPLATE proof of the
// closed label target registry. Test-only evidence, never production coverage: the fixture kind is registered through
// the nil-in-production labelHooks seam by the suites that install it (TestLabels, TestRLSIsolation) and restored at
// their teardown; the widening of app.label_assignments exists only in the disposable integration database, applied by
// the attested superuser of itSetup and removed again, never in a core/migrations file. No production file names the
// fixture (TestPackageBoundary, TestLabelTargetsClosed); core statements never name the per-type key column.

// The production storage fragments of app.label_assignments, copied verbatim from the exported pins of
// core/internal/migrate/labels_integration_test.go (LabelsWantTargetShape, LabelsWantAssignmentsSelect,
// LabelsWantAssignmentColumns; TestLabelsCatalog CT-LAB-2), which a test of this package cannot import.
// checkLabelFixturePins proves the copies byte-identical to that source text, so they cannot drift.
const (
	labFxWantTargetShape = `CHECK (((owner_scope_type = 'organization'::text) AND (owner_scope_id = organization_id) ` +
		`AND (target_type = 'group'::text) AND (target_group_id = target_id)))`
	labFxWantAssignmentsSelect = `((organization_id = ( SELECT app.ctx_org() AS ctx_org)) AND (( SELECT app.rls_org_role() AS rls_org_role) IS NOT NULL) ` +
		`AND ((( SELECT app.ctx_group() AS ctx_group) IS NULL) OR (target_group_id = ( SELECT app.ctx_group() AS ctx_group))) ` +
		`AND (EXISTS ( SELECT 1 FROM app.label_definitions d WHERE ((d.id = label_assignments.label_id) ` +
		`AND (d.organization_id = label_assignments.organization_id)))) AND (target_type = 'group'::text))`
	labFxWantAssignmentColumns = "id,organization_id,label_id,owner_scope_type,owner_scope_id,target_type,target_id,target_group_id,created_at"
)

// labFxPinSource is the file that owns the production pins (relative to this package directory).
const labFxPinSource = "../../migrate/labels_integration_test.go"

// labelTargetFixture is the test-only Group-owned target kind (the target_type value of its assignment rows).
const labelTargetFixture labelTargetKind = "fixture"

// fixtureTargetAdapter is the registry adapter of the fixture kind (contract item 1). resolve reads
// app.label_fixture_targets in the operation's own transaction (under the invoker's row-level security, like every
// runtime read); readable mirrors the fixture table's SELECT policy on the operation's facts: a manager (owner/admin) of
// the owning Group, or any role of it when the target admits members (members_may_read). The members_may_read fact of
// the target the operation resolved is kept from resolve for readable (the operations of these suites run one at a
// time; readable answers false for any Group other than the one last resolved).
type fixtureTargetAdapter struct {
	mu   sync.Mutex
	last fixtureResolved
}

type fixtureResolved struct {
	ok      bool
	owning  tenancy.GroupID
	mayRead bool
}

func (*fixtureTargetAdapter) admits(owner tenancy.ScopeType) bool { return owner == tenancy.ScopeGroup }

func (a *fixtureTargetAdapter) resolve(ctx context.Context, tx *trackedTx, _ Facts, org tenancy.OrganizationID, id [16]byte) (tenancy.GroupID, bool, error) {
	a.mu.Lock()
	a.last = fixtureResolved{}
	a.mu.Unlock()
	var g string
	var may bool
	err := tx.QueryRow(ctx, `SELECT group_id::text, members_may_read FROM app.label_fixture_targets
WHERE organization_id = $1::text::uuid AND id = $2::text::uuid`, org.String(), tenancy.GroupID(id).String()).Scan(&g, &may)
	if errors.Is(err, pgx.ErrNoRows) {
		return tenancy.GroupID{}, false, nil
	}
	if err != nil {
		return tenancy.GroupID{}, false, err
	}
	owning, err := tenancy.ParseGroupID(g)
	if err != nil {
		return tenancy.GroupID{}, false, err
	}
	a.mu.Lock()
	a.last = fixtureResolved{ok: true, owning: owning, mayRead: may}
	a.mu.Unlock()
	return owning, true, nil
}

func (a *fixtureTargetAdapter) readable(sel contextSelection, f Facts, owning tenancy.GroupID) bool {
	// app.rls_group_role answers NULL for a Group other than the selected one.
	if !zero(sel.Group) && sel.Group != owning {
		return false
	}
	a.mu.Lock()
	r := a.last
	a.mu.Unlock()
	if !r.ok || r.owning != owning {
		return false
	}
	t := Target{Kind: kindGroup, Organization: sel.Organization, Group: owning}
	if ok, _ := decide(capLabelGroupManage, sel, f, t); ok {
		return true
	}
	ok, _ := decide(capLabelGroupRead, sel, f, t)
	return ok && r.mayRead
}

// labFxNormalize is the catalog test's whitespace normalization (strings.Fields joined by one space).
func labFxNormalize(s string) string { return strings.Join(strings.Fields(s), " ") }

// labFxConst evaluates a string constant expression of literals joined by +.
func labFxConst(e ast.Expr) (string, bool) {
	switch x := e.(type) {
	case *ast.BasicLit:
		if x.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(x.Value)
		return s, err == nil
	case *ast.ParenExpr:
		return labFxConst(x.X)
	case *ast.BinaryExpr:
		if x.Op != token.ADD {
			return "", false
		}
		a, ok1 := labFxConst(x.X)
		b, ok2 := labFxConst(x.Y)
		return a + b, ok1 && ok2
	}
	return "", false
}

// checkLabelFixturePins reads the migrate package's source text and asserts that this file's three copies equal its
// exported pins byte for byte.
func checkLabelFixturePins(t *testing.T) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), labFxPinSource, nil, 0)
	if err != nil {
		t.Fatalf("read the production label pins: %v", err)
	}
	got := map[string]string{}
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.CONST {
			continue
		}
		for _, s := range g.Specs {
			vs := s.(*ast.ValueSpec)
			for i, n := range vs.Names {
				if strings.HasPrefix(n.Name, "LabelsWant") && i < len(vs.Values) {
					v, ok := labFxConst(vs.Values[i])
					if !ok {
						t.Fatalf("%s: not a string constant of literals", n.Name)
					}
					got[n.Name] = v
				}
			}
		}
	}
	for name, copied := range map[string]string{"LabelsWantTargetShape": labFxWantTargetShape,
		"LabelsWantAssignmentsSelect": labFxWantAssignmentsSelect, "LabelsWantAssignmentColumns": labFxWantAssignmentColumns} {
		v, ok := got[name]
		if !ok {
			t.Fatalf("%s is not declared in %s", name, labFxPinSource)
		}
		if v != copied {
			t.Fatalf("the copied %s drifted from %s:\n copy %s\n pin  %s", name, labFxPinSource, copied, v)
		}
	}
	if len(got) != 3 {
		t.Fatalf("%d LabelsWant* pins in %s, want exactly the three copied", len(got), labFxPinSource)
	}
}

// labFxCatalog is the observed storage shape of app.label_assignments: the target CHECK, the SELECT policy expression
// (raw and normalized, deparsed with search_path = pg_catalog), its roles and command, the column list in attnum order,
// and the fixture objects present.
type labFxCatalog struct {
	shape, selectRaw, selectNorm, selectMeta, columns string
	fixtureObjects                                    string
}

func (e *itEnv) labFxRead(t *testing.T) labFxCatalog {
	t.Helper()
	ctx := context.Background()
	tx, err := e.setup.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SET LOCAL search_path = pg_catalog`); err != nil {
		t.Fatal(err)
	}
	var c labFxCatalog
	if err := tx.QueryRow(ctx, `SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'app.label_assignments'::regclass
  AND conname = 'label_assignments_target_shape'`).Scan(&c.shape); err != nil {
		t.Fatalf("target-shape CHECK: %v", err)
	}
	if err := tx.QueryRow(ctx, `SELECT pg_get_expr(polqual, polrelid), format('%s|%s|%s', polcmd, polpermissive,
  (SELECT string_agg(r.rolname::text, ',' ORDER BY r.rolname::text COLLATE "C") FROM pg_roles r WHERE r.oid = ANY (polroles)))
FROM pg_policy WHERE polrelid = 'app.label_assignments'::regclass AND polname = 'label_assignments_select'`).Scan(&c.selectRaw, &c.selectMeta); err != nil {
		t.Fatalf("assignments SELECT policy: %v", err)
	}
	c.selectNorm = labFxNormalize(c.selectRaw)
	if err := tx.QueryRow(ctx, `SELECT string_agg(attname::text, ',' ORDER BY attnum) FROM pg_attribute
WHERE attrelid = 'app.label_assignments'::regclass AND attnum > 0 AND NOT attisdropped`).Scan(&c.columns); err != nil {
		t.Fatalf("assignment columns: %v", err)
	}
	if err := tx.QueryRow(ctx, `SELECT coalesce(string_agg(x, ',' ORDER BY x COLLATE "C"), '') FROM (
  SELECT 'relation:' || relname::text AS x FROM pg_class WHERE relnamespace = 'app'::regnamespace AND relname LIKE '%fixture%'
  UNION ALL SELECT 'constraint:' || conname::text FROM pg_constraint WHERE conrelid = 'app.label_assignments'::regclass AND conname LIKE '%fixture%'
  UNION ALL SELECT 'column:' || attname::text FROM pg_attribute WHERE attrelid = 'app.label_assignments'::regclass AND attnum > 0
    AND NOT attisdropped AND attname LIKE '%fixture%') s`).Scan(&c.fixtureObjects); err != nil {
		t.Fatalf("fixture objects: %v", err)
	}
	return c
}

// labFxProductionSuffix is the last conjunct of the production SELECT expression: the production target-visibility
// predicate the widening extends with the fixture disjunct.
const labFxProductionSuffix = "(target_type = 'group'::text))"

// installLabelFixture applies the test-only widening of Plan r2 §1.4 (contract items (2)-(4) for the fixture kind) as
// the attested superuser in one transaction, registers the fixture adapter through labelHooks, and returns the adapter
// and the teardown. The teardown (also registered with t.Cleanup; idempotent) removes any fixture-kind assignment row,
// drops the fixture foreign key and the generated key column, restores the production SELECT and CHECK bodies captured
// before the widening, then drops the fixture table (the widened policy depends on it), restores labelHooks, and
// byte-compares the restored fragments with the captured bodies and with the copied TestLabelsCatalog pins (T5 (g)).
func (e *itEnv) installLabelFixture(t *testing.T) (*fixtureTargetAdapter, func(t *testing.T)) {
	t.Helper()
	ctx := context.Background()
	checkLabelFixturePins(t)
	prod := e.labFxRead(t)
	// Precondition: the database holds exactly the production storage (never widen twice or a modified schema).
	if prod.shape != labFxWantTargetShape || prod.selectNorm != labFxWantAssignmentsSelect || prod.columns != labFxWantAssignmentColumns ||
		prod.fixtureObjects != "" {
		t.Fatalf("label storage is not the production shape before the fixture widening:\n%+v", prod)
	}
	rt, ow := pgx.Identifier{e.cfg.RuntimeRole()}.Sanitize(), pgx.Identifier{e.cfg.OwnerRole()}.Sanitize()
	install := []string{
		`CREATE TABLE app.label_fixture_targets (
  organization_id  uuid    NOT NULL,
  id               uuid    NOT NULL,
  group_id         uuid    NOT NULL,
  members_may_read boolean NOT NULL DEFAULT false,
  CONSTRAINT label_fixture_targets_pkey PRIMARY KEY (organization_id, id),
  CONSTRAINT label_fixture_targets_owner_key UNIQUE (organization_id, id, group_id),
  CONSTRAINT label_fixture_targets_group_fkey FOREIGN KEY (organization_id, group_id) REFERENCES app.groups (organization_id, id)
)`,
		`ALTER TABLE app.label_fixture_targets OWNER TO ` + ow,
		`ALTER TABLE app.label_fixture_targets ENABLE ROW LEVEL SECURITY`,
		`ALTER TABLE app.label_fixture_targets FORCE ROW LEVEL SECURITY`,
		`CREATE POLICY label_fixture_targets_select ON app.label_fixture_targets FOR SELECT TO ` + rt + `, ` + ow + ` USING (
  app.rls_group_role(group_id) IN ('group_owner', 'group_admin') OR (members_may_read AND app.rls_group_role(group_id) IS NOT NULL))`,
		`GRANT SELECT ON app.label_fixture_targets TO ` + rt,
		// (2) the target CHECK: the production disjunct OR the fixture disjunct.
		`ALTER TABLE app.label_assignments DROP CONSTRAINT label_assignments_target_shape`,
		`ALTER TABLE app.label_assignments ADD CONSTRAINT label_assignments_target_shape CHECK (
  (owner_scope_type = 'organization' AND owner_scope_id = organization_id AND target_type = 'group' AND target_group_id = target_id)
  OR (owner_scope_type = 'group' AND target_type = 'fixture' AND target_group_id = owner_scope_id))`,
		// (3) the NULL-skipping existence-and-ownership key.
		`ALTER TABLE app.label_assignments ADD COLUMN target_fixture_id uuid GENERATED ALWAYS AS (CASE WHEN target_type = 'fixture' THEN target_id END) STORED`,
		`ALTER TABLE app.label_assignments ADD CONSTRAINT label_assignments_fixture_target_fkey FOREIGN KEY (organization_id, target_fixture_id, target_group_id)
  REFERENCES app.label_fixture_targets (organization_id, id, group_id) ON UPDATE RESTRICT ON DELETE RESTRICT`,
		// (4) the SELECT policy: the production conjuncts with the production target predicate (target_type = 'group')
		// extended by the fixture disjunct, whose EXISTS carries the row's Organization and owning-Group facts and qualifies
		// the outer columns (label_assignments.*; Result round 1 deviation (2)).
		`DROP POLICY label_assignments_select ON app.label_assignments`,
		`CREATE POLICY label_assignments_select ON app.label_assignments FOR SELECT TO ` + rt + `, ` + ow + ` USING (
  organization_id = (SELECT app.ctx_org()) AND (SELECT app.rls_org_role()) IS NOT NULL
  AND ((SELECT app.ctx_group()) IS NULL OR target_group_id = (SELECT app.ctx_group()))
  AND EXISTS (SELECT 1 FROM app.label_definitions d WHERE d.id = label_assignments.label_id AND d.organization_id = label_assignments.organization_id)
  AND (target_type = 'group' OR (target_type = 'fixture' AND EXISTS (SELECT 1 FROM app.label_fixture_targets t
    WHERE t.organization_id = label_assignments.organization_id AND t.id = label_assignments.target_id
      AND t.group_id = label_assignments.target_group_id))))`,
	}
	e.labFxExec(t, "install the label fixture widening", install)

	a := &fixtureTargetAdapter{}
	prevHooks := labelHooks
	labelHooks = &labelTargetHooks{targets: map[labelTargetKind]labelTargetAdapter{labelTargetFixture: a}}

	var once sync.Once
	teardown := func(t *testing.T) {
		t.Helper()
		once.Do(func() {
			labelHooks = prevHooks
			var left int64
			tag, err := e.setup.Exec(ctx, `DELETE FROM app.label_assignments WHERE target_type <> 'group'`)
			if err != nil {
				t.Errorf("teardown: remove fixture-kind assignments: %v (SQLSTATE %s)", err, pgState(err))
			} else {
				left = tag.RowsAffected()
			}
			if left > 0 {
				t.Logf("teardown removed %d fixture-kind assignment row(s) left by the suite", left)
			}
			// The reverse of the install order: the widened SELECT policy's EXISTS references app.label_fixture_targets
			// (a pg_depend entry), so the production policy is restored before the fixture table is dropped (2BP01 otherwise).
			e.labFxExec(t, "restore the production label storage", []string{
				`ALTER TABLE app.label_assignments DROP CONSTRAINT label_assignments_fixture_target_fkey`,
				`ALTER TABLE app.label_assignments DROP COLUMN target_fixture_id`,
				`DROP POLICY label_assignments_select ON app.label_assignments`,
				`CREATE POLICY label_assignments_select ON app.label_assignments FOR SELECT TO ` + rt + `, ` + ow + ` USING (` + prod.selectRaw + `)`,
				`ALTER TABLE app.label_assignments DROP CONSTRAINT label_assignments_target_shape`,
				`ALTER TABLE app.label_assignments ADD CONSTRAINT label_assignments_target_shape ` + prod.shape,
				`DROP TABLE app.label_fixture_targets`,
			})
			got := e.labFxRead(t)
			if got.fixtureObjects != "" {
				t.Errorf("T5 (g): fixture objects remain after the teardown: %s", got.fixtureObjects)
			}
			for _, c := range []struct{ what, got, captured, pin string }{
				{"target-shape CHECK", got.shape, prod.shape, labFxWantTargetShape},
				{"assignments SELECT policy", got.selectNorm, prod.selectNorm, labFxWantAssignmentsSelect},
				{"assignment columns", got.columns, prod.columns, labFxWantAssignmentColumns},
				{"assignments SELECT command/permissive/roles", got.selectMeta, prod.selectMeta, prod.selectMeta},
			} {
				if c.got != c.captured || c.got != c.pin {
					t.Errorf("T5 (g): restored %s differs:\n got      %s\n captured %s\n pin      %s", c.what, c.got, c.captured, c.pin)
				}
			}
		})
	}
	t.Cleanup(func() { teardown(t) })

	// The widened storage (predicted deparse fragments; the production conjuncts are kept verbatim).
	w := e.labFxRead(t)
	inner := strings.TrimSuffix(strings.TrimPrefix(labFxWantTargetShape, "CHECK ("), ")")
	if !strings.Contains(w.shape, inner) || !strings.Contains(w.shape, "(target_type = 'fixture'::text)") ||
		!strings.Contains(w.shape, "(target_group_id = owner_scope_id)") {
		t.Errorf("widened target-shape CHECK: %s", w.shape)
	}
	if !strings.HasPrefix(w.selectNorm, strings.TrimSuffix(labFxWantAssignmentsSelect, labFxProductionSuffix)+"((target_type = 'group'::text) OR ") ||
		!strings.Contains(w.selectNorm, "app.label_fixture_targets t") ||
		!strings.Contains(w.selectNorm, "(t.group_id = label_assignments.target_group_id)") {
		t.Errorf("widened assignments SELECT policy: %s", w.selectNorm)
	}
	if w.selectMeta != prod.selectMeta {
		t.Errorf("widened SELECT policy command/roles %s, want the production %s", w.selectMeta, prod.selectMeta)
	}
	if w.columns != labFxWantAssignmentColumns+",target_fixture_id" {
		t.Errorf("widened assignment columns: %s", w.columns)
	}
	if w.fixtureObjects != "column:target_fixture_id,constraint:label_assignments_fixture_target_fkey,relation:label_fixture_targets,"+
		"relation:label_fixture_targets_owner_key,relation:label_fixture_targets_pkey" {
		t.Errorf("widened fixture objects: %s", w.fixtureObjects)
	}
	return a, teardown
}

// labFxExec runs DDL statements as the attested superuser in one transaction.
func (e *itEnv) labFxExec(t *testing.T, what string, stmts []string) {
	t.Helper()
	ctx := context.Background()
	tx, err := e.setup.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, s := range stmts {
		if _, err := tx.Exec(ctx, s); err != nil {
			t.Fatalf("%s: %v (SQLSTATE %s)\n%s", what, err, pgState(err), s)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("%s: commit: %v", what, err)
	}
}

// fixtureTarget creates one fixture target of Organization org owned by Group g (superuser) and returns its id.
func (e *itEnv) fixtureTarget(t *testing.T, org tenancy.OrganizationID, g tenancy.GroupID, membersMayRead bool) string {
	t.Helper()
	id := rlsUUID(t)
	if _, err := e.setup.Exec(context.Background(), `INSERT INTO app.label_fixture_targets (organization_id, id, group_id, members_may_read)
VALUES ($1::text::uuid, $2::uuid, $3::text::uuid, $4)`, org.String(), id, g.String(), membersMayRead); err != nil {
		t.Fatalf("fixture target: %v (SQLSTATE %s)", err, pgState(err))
	}
	return id
}
