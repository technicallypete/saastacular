//go:build integration

// B-BUSINESS-PROFILE-IT (G006 Plan r2 §5): the BusinessProfile model against the disposable PostgreSQL of the "migrations"
// suite, through a pool logged in as the actual runtime role; owner-level cases use the migrator with SET ROLE
// <db>_owner. It reuses the G005 helpers of integration_test.go (setup, the fixtures, faultDB, waitBlocked, h) and adds
// the statement-recording recordDB. Assertions are scoped to this test's own rows, compare row-set hashes around every
// rejected or failed operation and name the intended SQLSTATE and constraint or error code. Fails, never skips, when
// settings are absent.
//
// Required environment: as TestTenancyModel (TENANCY_LOG_FILE receives this test's operation log).
package tenancy_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/observability"
	"core/internal/tenancy"
)

// --- recordDB: the runtime pool, recording every statement a Service sends ------------------------------------------

type recordDB struct {
	pool  *pgxpool.Pool
	mu    sync.Mutex
	stmts []string
}

func (r *recordDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	tx, err := r.pool.BeginTx(ctx, o)
	if err != nil {
		return nil, err
	}
	return &recordTx{Tx: tx, r: r}, nil
}

type recordTx struct {
	pgx.Tx
	r *recordDB
}

func (t *recordTx) note(sql string) {
	t.r.mu.Lock()
	t.r.stmts = append(t.r.stmts, sql)
	t.r.mu.Unlock()
}

func (t *recordTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	t.note(sql)
	return t.Tx.QueryRow(ctx, sql, args...)
}

func (t *recordTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	t.note(sql)
	return t.Tx.Query(ctx, sql, args...)
}

func (t *recordTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	t.note(sql)
	return t.Tx.Exec(ctx, sql, args...)
}

// writes returns the recorded write statements (INSERT, UPDATE, DELETE) and resets the record.
func (r *recordDB) writes() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []string
	for _, s := range r.stmts {
		if strings.HasPrefix(s, "INSERT") || strings.HasPrefix(s, "UPDATE") || strings.HasPrefix(s, "DELETE") {
			out = append(out, s)
		}
	}
	r.stmts = nil
	return out
}

// all returns every recorded statement and resets the record.
func (r *recordDB) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := r.stmts
	r.stmts = nil
	return out
}

// --- fixtures (Spec R8) ---------------------------------------------------------------------------------------------

// world is the Plan r2 §5.2 fixture: Organizations A and B; users U, V, W, X, Y, Z; Groups A/g1 (created by U), A/g2
// (X), A/g3 (Y), A/g4 (X, never associated) and B/h1 (Z, g1's slug); profiles A/p1 and A/p2 and B/q1 with one display
// name; bindings U group_admin@g3, U org_member@A, U group_member@h1, V group_member@g2, W org_owner@A; links g1->p1,
// g2->p1, g3->p2, h1->q1.
type world struct {
	A, B               tenancy.OrganizationID
	U, V, W, X, Y, Z   tenancy.UserID
	R                  tenancy.UserID // creates the scratch Groups of the race and helper cases
	g1, g2, g3, g4, h1 tenancy.Group
	p1, p2, q1         tenancy.BusinessProfile
	l1, l2, l3, l4     tenancy.GroupBusinessProfile
}

const sharedName = "Acme Shared Name"

func (e *env) profile(t *testing.T, org tenancy.OrganizationID, name string) tenancy.BusinessProfile {
	t.Helper()
	p, err := e.svc.CreateBusinessProfile(context.Background(), org, tenancy.BusinessContext{DisplayName: name})
	if err != nil {
		t.Fatalf("create business profile: %v", err)
	}
	return p
}

func (e *env) link(t *testing.T, org tenancy.OrganizationID, g tenancy.GroupID, p tenancy.BusinessProfileID) tenancy.GroupBusinessProfile {
	t.Helper()
	a, err := e.svc.LinkGroupBusinessProfile(context.Background(), org, g, p)
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	return a
}

func (e *env) world(t *testing.T) *world {
	t.Helper()
	ctx := context.Background()
	w := &world{}
	w.A, w.B = e.org(t, "Profile Fixture A").ID, e.org(t, "Profile Fixture B").ID
	w.U, w.V, w.W, w.X, w.Y, w.Z = e.user(t, false), e.user(t, false), e.user(t, false), e.user(t, false), e.user(t, false), e.user(t, false)
	w.R = e.user(t, false)
	w.g1, _ = e.group(t, w.A, w.U, "Acme Revenue", "acme-revenue")
	w.g2, _ = e.group(t, w.A, w.X, "Acme Operations", "acme-operations")
	w.g3, _ = e.group(t, w.A, w.Y, "Acme Field", "acme-field")
	w.g4, _ = e.group(t, w.A, w.X, "Acme Foundation", "acme-foundation")
	w.h1, _ = e.group(t, w.B, w.Z, "Acme Revenue", "acme-revenue")
	for _, b := range []func() error{
		func() error { _, err := e.svc.GrantGroupRole(ctx, w.A, w.g3.ID, w.U, tenancy.GroupAdmin); return err },
		func() error { _, err := e.svc.GrantOrganizationRole(ctx, w.A, w.U, tenancy.OrgMember); return err },
		func() error { _, err := e.svc.GrantGroupRole(ctx, w.B, w.h1.ID, w.U, tenancy.GroupMember); return err },
		func() error { _, err := e.svc.GrantGroupRole(ctx, w.A, w.g2.ID, w.V, tenancy.GroupMember); return err },
		func() error { _, err := e.svc.GrantOrganizationRole(ctx, w.A, w.W, tenancy.OrgOwner); return err },
	} {
		if err := b(); err != nil {
			t.Fatalf("fixture binding: %v", err)
		}
	}
	w.p1, w.p2, w.q1 = e.profile(t, w.A, sharedName), e.profile(t, w.A, sharedName), e.profile(t, w.B, sharedName)
	w.l1, w.l2 = e.link(t, w.A, w.g1.ID, w.p1.ID), e.link(t, w.A, w.g2.ID, w.p1.ID)
	w.l3, w.l4 = e.link(t, w.A, w.g3.ID, w.p2.ID), e.link(t, w.B, w.h1.ID, w.q1.ID)
	return w
}

// bpRows hashes the profiles and associations of the given Organizations (owner connection).
func (e *env) bpRows(t *testing.T, orgs ...tenancy.OrganizationID) string {
	t.Helper()
	ids := make([]string, len(orgs))
	for i, o := range orgs {
		ids[i] = o.String()
	}
	var s string
	if err := e.own.QueryRow(context.Background(), `
SELECT coalesce((SELECT md5(string_agg(p::text, E'\n' ORDER BY p.id)) FROM app.business_profiles p WHERE p.organization_id::text = ANY ($1::text[])), '-') || '|' ||
       coalesce((SELECT md5(string_agg(a::text, E'\n' ORDER BY a.id)) FROM app.group_business_profiles a WHERE a.organization_id::text = ANY ($1::text[])), '-')`,
		ids).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

// all hashes every fixture row of the given Organizations: Organizations, Groups, bindings, profiles and associations.
func (e *env) all(t *testing.T, orgs ...tenancy.OrganizationID) string {
	t.Helper()
	return e.rows(t, orgs...) + "|" + e.bpRows(t, orgs...)
}

func pathKey(p tenancy.BusinessProfileAccessPath) string {
	return p.UserID.String() + " " + p.BusinessProfileID.String() + " " + p.GroupID.String() + " " + string(p.GroupRole)
}

func pathKeys(ps []tenancy.BusinessProfileAccessPath) []string {
	out := []string{}
	for _, p := range ps {
		out = append(out, pathKey(p))
	}
	return out
}

func key(u tenancy.UserID, p tenancy.BusinessProfileID, g tenancy.GroupID, r tenancy.GroupRole) string {
	return u.String() + " " + p.String() + " " + g.String() + " " + string(r)
}

// cc returns a string of one code point (keeps control characters out of this source file).
func cc(r rune) string { return string(r) }

func sp(s string) *string { return &s }

// --- the model test -----------------------------------------------------------------------------------------------------

func TestBusinessProfileModel(t *testing.T) {
	e := setup(t)
	ctx := context.Background()
	svc := e.svc
	rec := &recordDB{pool: e.pool}
	rsvc := tenancy.NewService(rec, nil)

	t.Run("runtime identity", func(t *testing.T) {
		var cur, sess string
		var super, bypass bool
		var memberships int
		err := e.pool.QueryRow(ctx, `SELECT current_user::text, session_user::text, r.rolsuper, r.rolbypassrls,
  (SELECT count(*) FROM pg_auth_members m WHERE m.member = r.oid)::int FROM pg_roles r WHERE r.rolname = current_user`).Scan(&cur, &sess, &super, &bypass, &memberships)
		if err != nil {
			t.Fatal(err)
		}
		if cur != e.cfg.RuntimeRole() || sess != cur || super || bypass || memberships != 0 {
			t.Fatalf("model pool: current=%s session=%s super=%v bypassrls=%v memberships=%d", cur, sess, super, bypass, memberships)
		}
	})

	w := e.world(t)

	t.Run("IT-BP-ACC-1 exact access paths of the fixture", func(t *testing.T) {
		// Read first, before any test changes the fixture.
		want := map[tenancy.BusinessProfileID][]string{
			w.p1.ID: {key(w.U, w.p1.ID, w.g1.ID, tenancy.GroupOwner), key(w.V, w.p1.ID, w.g2.ID, tenancy.GroupMember), key(w.X, w.p1.ID, w.g2.ID, tenancy.GroupOwner)},
			w.p2.ID: {key(w.U, w.p2.ID, w.g3.ID, tenancy.GroupAdmin), key(w.Y, w.p2.ID, w.g3.ID, tenancy.GroupOwner)},
		}
		for p, keys := range want {
			got, err := svc.ListBusinessProfileAccessPaths(ctx, w.A, p)
			if err != nil {
				t.Fatal(err)
			}
			// Ordered by user, then Group.
			slices.SortFunc(keys, func(a, b string) int { return strings.Compare(a[:36]+a[74:110], b[:36]+b[74:110]) })
			if strings.Join(pathKeys(got), "\n") != strings.Join(keys, "\n") {
				t.Errorf("paths of %s:\n got %v\nwant %v", h(p), pathKeys(got), keys)
			}
		}
		q, err := svc.ListBusinessProfileAccessPaths(ctx, w.B, w.q1.ID)
		if err != nil {
			t.Fatal(err)
		}
		wantQ := []string{key(w.U, w.q1.ID, w.h1.ID, tenancy.GroupMember), key(w.Z, w.q1.ID, w.h1.ID, tenancy.GroupOwner)}
		slices.Sort(wantQ)
		if strings.Join(pathKeys(q), "\n") != strings.Join(wantQ, "\n") {
			t.Errorf("paths of q1: %v", pathKeys(q))
		}
		// One global user reaches two profiles in A, and only q1 in B; W (org_owner only) has none.
		ua, err := svc.ListUserBusinessProfileAccessPaths(ctx, w.A, w.U)
		if err != nil {
			t.Fatal(err)
		}
		wantU := []string{key(w.U, w.p1.ID, w.g1.ID, tenancy.GroupOwner), key(w.U, w.p2.ID, w.g3.ID, tenancy.GroupAdmin)}
		slices.Sort(wantU)
		if strings.Join(pathKeys(ua), "\n") != strings.Join(wantU, "\n") {
			t.Errorf("U in A: %v", pathKeys(ua))
		}
		if ub, err := svc.ListUserBusinessProfileAccessPaths(ctx, w.B, w.U); err != nil || strings.Join(pathKeys(ub), ",") != key(w.U, w.q1.ID, w.h1.ID, tenancy.GroupMember) {
			t.Errorf("U in B: %v %v", pathKeys(ub), err)
		}
		if ww, err := svc.ListUserBusinessProfileAccessPaths(ctx, w.A, w.W); err != nil || len(ww) != 0 {
			t.Errorf("W (org_owner only): %v %v", pathKeys(ww), err)
		}
		// g4 (X is its owner) has no association, so it adds no path.
		if xa, _ := svc.ListUserBusinessProfileAccessPaths(ctx, w.A, w.X); len(xa) != 1 {
			t.Errorf("X in A: %v", pathKeys(xa))
		}
	})

	t.Run("IT-BP-ID-1 identities", func(t *testing.T) {
		ids := map[[16]byte]bool{}
		for _, p := range []tenancy.BusinessProfile{w.p1, w.p2, w.q1} {
			if !isV4(p.ID) || ids[p.ID] {
				t.Fatalf("profile ID %s is not a fresh v4 UUID", h(p.ID))
			}
			ids[p.ID] = true
			if p.Context.DisplayName != sharedName || p.CreatedAt.IsZero() {
				t.Fatalf("profile %s: name %s", h(p.ID), h(p.Context.DisplayName))
			}
		}
		if w.p1.OrganizationID != w.A || w.p2.OrganizationID != w.A || w.q1.OrganizationID != w.B {
			t.Fatal("profile Organizations differ")
		}
		for _, l := range []tenancy.GroupBusinessProfile{w.l1, w.l2, w.l3, w.l4} {
			if !isV4(l.ID) || ids[l.ID] {
				t.Fatalf("association ID %s is not a fresh v4 UUID", h(l.ID))
			}
			ids[l.ID] = true
		}
		if n := e.count(t, `SELECT count(*) FROM app.business_profiles WHERE organization_id = $1::text::uuid AND display_name = $2`, w.A.String(), sharedName); n != 2 {
			t.Fatalf("equal names in A: %d rows", n)
		}
	})

	t.Run("IT-BP-ID-2 updates keep identity", func(t *testing.T) {
		before := e.rows(t, w.A, w.B)
		links := e.count(t, `SELECT count(*) FROM app.group_business_profiles WHERE business_profile_id = $1::text::uuid`, w.p1.ID.String())
		up, err := svc.UpdateBusinessProfile(ctx, w.A, w.p1.ID, tenancy.BusinessContext{DisplayName: "Acme Renamed", Industry: sp("Logistics")})
		if err != nil {
			t.Fatal(err)
		}
		if up.ID != w.p1.ID || up.OrganizationID != w.A || !up.CreatedAt.Equal(w.p1.CreatedAt) || up.Context.DisplayName != "Acme Renamed" {
			t.Fatalf("updated profile: %s %s", h(up.ID), h(up.Context.DisplayName))
		}
		if e.rows(t, w.A, w.B) != before || e.count(t, `SELECT count(*) FROM app.group_business_profiles WHERE business_profile_id = $1::text::uuid`, w.p1.ID.String()) != links {
			t.Fatal("Organizations, Groups, bindings or links changed by an update")
		}
		if _, err := svc.UpdateBusinessProfile(ctx, w.A, w.p1.ID, tenancy.BusinessContext{DisplayName: sharedName}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("IT-BP-ID-3 runtime cannot change or supply keys", func(t *testing.T) {
		before := e.all(t, w.A, w.B)
		p, l, g := w.p1.ID.String(), w.l1.ID.String(), w.g4.ID.String()
		for what, stmt := range map[string]string{
			"profile id":           `UPDATE app.business_profiles SET id = gen_random_uuid() WHERE id = '` + p + `'`,
			"profile Organization": `UPDATE app.business_profiles SET organization_id = '` + w.B.String() + `' WHERE id = '` + p + `'`,
			"association group":    `UPDATE app.group_business_profiles SET group_id = '` + g + `' WHERE id = '` + l + `'`,
			"association profile":  `UPDATE app.group_business_profiles SET business_profile_id = '` + w.p2.ID.String() + `' WHERE id = '` + l + `'`,
			"explicit profile id":  `INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES (gen_random_uuid(), '` + w.A.String() + `', 'x')`,
			"explicit created_at":  `INSERT INTO app.business_profiles (organization_id, display_name, created_at) VALUES ('` + w.A.String() + `', 'x', now())`,
			"explicit association id": `INSERT INTO app.group_business_profiles (id, organization_id, group_id, business_profile_id) VALUES (gen_random_uuid(), '` +
				w.A.String() + `', '` + g + `', '` + p + `')`,
			"delete profile": `DELETE FROM app.business_profiles WHERE id = '` + p + `'`,
		} {
			expectSQL(t, e.rt, "runtime "+what, "42501", "", stmt)
		}
		if e.all(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-BP-ID-4 owner key changes hit the immutability triggers", func(t *testing.T) {
		p, l := w.p1.ID.String(), w.l1.ID.String()
		expectSQL(t, e.own, "owner profile id", "23000", "business_profiles_keys_immutable", `UPDATE app.business_profiles SET id = gen_random_uuid() WHERE id = '`+p+`'`)
		expectSQL(t, e.own, "owner profile Organization", "23000", "business_profiles_keys_immutable",
			`UPDATE app.business_profiles SET organization_id = '`+w.B.String()+`' WHERE id = '`+p+`'`)
		for _, col := range []string{"id", "organization_id", "group_id", "business_profile_id"} {
			expectSQL(t, e.own, "owner association "+col, "23000", "group_business_profiles_keys_immutable",
				`UPDATE app.group_business_profiles SET `+col+` = gen_random_uuid() WHERE id = '`+l+`'`)
		}
	})

	t.Run("IT-BP-ID-5 unknown Organization and zero IDs", func(t *testing.T) {
		unknown, _ := tenancy.ParseOrganizationID("0f8fad5b-d9cb-469f-a165-70867728950e")
		before := e.all(t, w.A, w.B)
		_, err := svc.CreateBusinessProfile(ctx, unknown, tenancy.BusinessContext{DisplayName: "x"})
		wantCode(t, "unknown Organization", err, tenancy.CodeOrganizationNotFound)
		expectSQL(t, e.rt, "runtime unknown Organization", "23503", "business_profiles_organization_fkey",
			`INSERT INTO app.business_profiles (organization_id, display_name) VALUES ('0f8fad5b-d9cb-469f-a165-70867728950e', 'x')`)
		var zp tenancy.BusinessProfileID
		var zo tenancy.OrganizationID
		_, err = svc.GetBusinessProfile(ctx, w.A, zp)
		wantCode(t, "zero profile", err, tenancy.CodeInvalidID)
		_, err = svc.CreateBusinessProfile(ctx, zo, tenancy.BusinessContext{DisplayName: "x"})
		wantCode(t, "zero Organization", err, tenancy.CodeInvalidID)
		expectSQL(t, e.own, "owner zero id", "23514", "business_profiles_id_nonzero",
			`INSERT INTO app.business_profiles (id, organization_id, display_name) VALUES ('00000000-0000-0000-0000-000000000000', '`+w.A.String()+`', 'x')`)
		expectSQL(t, e.own, "owner zero association id", "23514", "group_business_profiles_id_nonzero",
			`INSERT INTO app.group_business_profiles (id, organization_id, group_id, business_profile_id) VALUES ('00000000-0000-0000-0000-000000000000', '`+
				w.A.String()+`', '`+w.g4.ID.String()+`', '`+w.p1.ID.String()+`')`)
		if e.all(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-BP-CTX-1 the service field matrix", func(t *testing.T) {
		long := func(n int) string { return strings.Repeat("é", n) }
		before := e.all(t, w.A)
		for name, c := range map[string]struct {
			in   tenancy.BusinessContext
			want tenancy.Code
		}{
			"display 256":      {tenancy.BusinessContext{DisplayName: long(256)}, tenancy.CodeInvalidDisplayName},
			"display NUL":      {tenancy.BusinessContext{DisplayName: "a" + cc(0) + "b"}, tenancy.CodeInvalidDisplayName},
			"display TAB":      {tenancy.BusinessContext{DisplayName: "a\tb"}, tenancy.CodeInvalidDisplayName},
			"display DEL":      {tenancy.BusinessContext{DisplayName: "a" + cc(0x7f)}, tenancy.CodeInvalidDisplayName},
			"display C1":       {tenancy.BusinessContext{DisplayName: "a" + cc(0x9f)}, tenancy.CodeInvalidDisplayName},
			"display NEL":      {tenancy.BusinessContext{DisplayName: "a" + cc(0x85) + "b"}, tenancy.CodeInvalidDisplayName},
			"display spaces":   {tenancy.BusinessContext{DisplayName: "   "}, tenancy.CodeInvalidDisplayName},
			"legal CR":         {tenancy.BusinessContext{DisplayName: "A", LegalName: sp("a\rb")}, tenancy.CodeInvalidLegalName},
			"description 4001": {tenancy.BusinessContext{DisplayName: "A", Description: sp(long(4001))}, tenancy.CodeInvalidDescription},
			"industry LF":      {tenancy.BusinessContext{DisplayName: "A", Industry: sp("a\nb")}, tenancy.CodeInvalidIndustry},
			"website scheme":   {tenancy.BusinessContext{DisplayName: "A", Website: sp("javascript:alert(1)")}, tenancy.CodeInvalidWebsite},
			"invalid UTF-8":    {tenancy.BusinessContext{DisplayName: string([]byte{0x61, 0xff})}, tenancy.CodeInvalidDisplayName},
		} {
			_, err := svc.CreateBusinessProfile(ctx, w.A, c.in)
			wantCode(t, "create "+name, err, c.want)
			_, err = svc.UpdateBusinessProfile(ctx, w.A, w.p2.ID, c.in)
			wantCode(t, "update "+name, err, c.want)
		}
		if e.all(t, w.A) != before {
			t.Fatal("rows changed by rejected inputs")
		}
		// Valid values round-trip byte-for-byte after trimming; empty and whitespace-only optional fields are absent.
		in := tenancy.BusinessContext{DisplayName: long(255), LegalName: sp(" Acme Holdings Ltd "), Description: sp("Line one\n\tLine two"),
			Industry: sp(""), Website: sp("https://example.com/path?q=1")}
		p, err := svc.CreateBusinessProfile(ctx, w.A, in)
		if err != nil {
			t.Fatal(err)
		}
		got, err := svc.GetBusinessProfile(ctx, w.A, p.ID)
		if err != nil {
			t.Fatal(err)
		}
		c := got.Context
		if c.DisplayName != long(255) || c.LegalName == nil || *c.LegalName != "Acme Holdings Ltd" || c.Description == nil ||
			*c.Description != "Line one\n\tLine two" || c.Industry != nil || c.Website == nil || *c.Website != "https://example.com/path?q=1" {
			t.Fatalf("round trip: %s", h(c))
		}
	})

	t.Run("IT-BP-CTX-2 runtime CHECK negatives", func(t *testing.T) {
		before := e.all(t, w.A)
		a, p := w.A.String(), w.p2.ID.String()
		for _, c := range []struct{ what, constraint, ins, upd string }{
			{"empty display name", "business_profiles_display_name_valid", `'', NULL, NULL`, `display_name = ''`},
			{"long display name", "business_profiles_display_name_valid", `repeat('n', 256), NULL, NULL`, `display_name = repeat('n', 256)`},
			{"display control", "business_profiles_display_name_valid", `'a' || chr(1), NULL, NULL`, `display_name = 'a' || chr(1)`},
			{"display padding", "business_profiles_display_name_valid", `' a', NULL, NULL`, `display_name = ' a'`},
			{"empty legal name", "business_profiles_legal_name_valid", `'x', '', NULL`, `legal_name = ''`},
			{"description CR", "business_profiles_description_valid", `'x', NULL, 'a' || chr(13) || 'b'`, `description = 'a' || chr(13) || 'b'`},
			{"website javascript", "business_profiles_website_valid", `'x', NULL, NULL`, `website = 'javascript:alert(1)'`},
			{"website ftp", "business_profiles_website_valid", `'x', NULL, NULL`, `website = 'ftp://example.com'`},
			{"website backslash", "business_profiles_website_valid", `'x', NULL, NULL`, `website = 'https://e.com' || chr(92) || 'x'`},
			{"empty industry", "business_profiles_industry_valid", `'x', NULL, NULL`, `industry = ''`},
			{"display DEL", "business_profiles_display_name_valid", `'a' || chr(127), NULL, NULL`, `display_name = 'a' || chr(127)`},
			{"display C1", "business_profiles_display_name_valid", `'a' || chr(133) || 'b', NULL, NULL`, `display_name = 'a' || chr(133) || 'b'`},
			{"website DEL", "business_profiles_website_valid", `'x', NULL, NULL`, `website = 'https://e.com/' || chr(127)`},
			{"website C1", "business_profiles_website_valid", `'x', NULL, NULL`, `website = 'https://e.com/' || chr(133)`},
			{"website space", "business_profiles_website_valid", `'x', NULL, NULL`, `website = 'https://e.com/a b'`},
			{"website non-ASCII", "business_profiles_website_valid", `'x', NULL, NULL`, `website = 'https://e.com/' || chr(252)`},
		} {
			if !strings.HasPrefix(c.what, "website") && !strings.HasPrefix(c.what, "empty industry") {
				expectSQL(t, e.rt, "insert "+c.what, "23514", c.constraint,
					`INSERT INTO app.business_profiles (organization_id, display_name, legal_name, description) VALUES ('`+a+`', `+c.ins+`)`)
			}
			expectSQL(t, e.rt, "update "+c.what, "23514", c.constraint, `UPDATE app.business_profiles SET `+c.upd+` WHERE id = '`+p+`'`)
		}
		expectSQL(t, e.rt, "insert website ftp", "23514", "business_profiles_website_valid",
			`INSERT INTO app.business_profiles (organization_id, display_name, website) VALUES ('`+a+`', 'x', 'ftp://example.com')`)
		// Go-valid values never hit a CHECK: the U-URL valid websites and the Go bounds, validated by the service rules,
		// then inserted verbatim by the runtime role in one rolled-back transaction.
		long := func(n int) string { return strings.Repeat("é", n) }
		valid := []tenancy.BusinessContext{
			{DisplayName: long(255), LegalName: sp(long(255)), Description: sp(long(3990) + "\ta\nb" + long(6)), Industry: sp(long(255))},
			{DisplayName: "a" + cc(0x202e) + "b" + cc(0x200d), Description: sp("x" + cc(0xa0) + "y")},
		}
		for _, site := range []string{"https://example.com", "http://example.com:8080/p?q=1#f", "https://xn--bcher-kva.example/",
			"https://192.0.2.1/", "https://[2001:db8::1]/", "https://example.com/" + strings.Repeat("a", 2048-len("https://example.com/")),
			"https://example.com.", "http://a-b.example:1/", "https://example.com:65535", "https://e.com/~!$&'()*+,;=:@[]{}|^`<>\"/?#"} {
			valid = append(valid, tenancy.BusinessContext{DisplayName: "x", Website: sp(site)})
		}
		tx, err := e.rt.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		for i, c := range valid {
			v, err := tenancy.ValidateBusinessContext(c)
			if err != nil {
				t.Fatalf("valid %d is not Go-valid: %v", i, err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO app.business_profiles (organization_id, display_name, legal_name, description, industry, website) VALUES ($1::text::uuid, $2, $3, $4, $5, $6)`,
				a, v.DisplayName, v.LegalName, v.Description, v.Industry, v.Website); err != nil {
				t.Fatalf("valid %d hit the database: %v", i, err)
			}
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		if e.all(t, w.A) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-BP-CTX-3 one shared profile, one update", func(t *testing.T) {
		before := e.rows(t, w.A, w.B)
		other := e.bpRows(t, w.B)
		p2 := e.count(t, `SELECT count(*) FROM app.business_profiles WHERE id = $1::text::uuid AND display_name = $2`, w.p2.ID.String(), sharedName)
		_, a, err := svc.GetGroupBusinessProfile(ctx, w.A, w.g1.ID)
		_, b, err2 := svc.GetGroupBusinessProfile(ctx, w.A, w.g2.ID)
		if err != nil || err2 != nil || a.ID != w.p1.ID || b.ID != w.p1.ID {
			t.Fatalf("g1/g2 profiles before: %v %v", err, err2)
		}
		nc := tenancy.BusinessContext{DisplayName: "Acme Shared Update", Description: sp("Updated once")}
		if _, err := svc.UpdateBusinessProfile(ctx, w.A, w.p1.ID, nc); err != nil {
			t.Fatal(err)
		}
		for _, g := range []tenancy.GroupID{w.g1.ID, w.g2.ID} {
			_, p, err := svc.GetGroupBusinessProfile(ctx, w.A, g)
			if err != nil || p.ID != w.p1.ID || p.Context.DisplayName != "Acme Shared Update" || p.Context.Description == nil || *p.Context.Description != "Updated once" {
				t.Fatalf("Group %s reads %s %v", h(g), h(p.Context), err)
			}
		}
		if e.rows(t, w.A, w.B) != before || e.bpRows(t, w.B) != other ||
			e.count(t, `SELECT count(*) FROM app.business_profiles WHERE id = $1::text::uuid AND display_name = $2`, w.p2.ID.String(), sharedName) != p2 ||
			e.count(t, `SELECT count(*) FROM app.business_profiles WHERE id = $1::text::uuid`, w.p1.ID.String()) != 1 {
			t.Fatal("an unrelated row changed, or p1 is not exactly one row")
		}
		// Concurrent full replacements: the stored row always equals exactly one writer's complete context, and readers
		// through either Group never see a mix.
		var wg sync.WaitGroup
		stop := make(chan struct{})
		var mixed atomic.Int64
		for _, g := range []tenancy.GroupID{w.g1.ID, w.g2.ID} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-stop:
						return
					default:
					}
					if _, p, err := svc.GetGroupBusinessProfile(ctx, w.A, g); err == nil {
						d := p.Context.DisplayName
						if p.Context.Description == nil || *p.Context.Description != "writer "+strings.TrimPrefix(d, "Concurrent ") {
							if strings.HasPrefix(d, "Concurrent ") {
								mixed.Add(1)
							}
						}
					}
				}
			}()
		}
		var writers sync.WaitGroup
		for i := range 16 {
			writers.Add(1)
			go func() {
				defer writers.Done()
				n := fmt.Sprint(i)
				if _, err := svc.UpdateBusinessProfile(ctx, w.A, w.p1.ID, tenancy.BusinessContext{DisplayName: "Concurrent " + n, Description: sp("writer " + n)}); err != nil {
					t.Errorf("writer %s: %v", n, err)
				}
			}()
		}
		writers.Wait()
		close(stop)
		wg.Wait()
		final, err := svc.GetBusinessProfile(ctx, w.A, w.p1.ID)
		if err != nil || final.Context.Description == nil || *final.Context.Description != "writer "+strings.TrimPrefix(final.Context.DisplayName, "Concurrent ") {
			t.Fatalf("final context is not one writer's: %s %v", h(final.Context), err)
		}
		if mixed.Load() != 0 {
			t.Fatalf("%d reads saw a mixed context", mixed.Load())
		}
		if _, err := svc.UpdateBusinessProfile(ctx, w.A, w.p1.ID, tenancy.BusinessContext{DisplayName: sharedName}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("IT-BP-CTX-4 inert values; nothing is fetched", func(t *testing.T) {
		var hits atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hits.Add(1) }))
		defer srv.Close()
		relations := e.count(t, `SELECT count(*) FROM pg_class WHERE relnamespace = 'app'::regnamespace`)
		for _, v := range []string{`'); DROP TABLE app.groups; --`, `<script>alert(1)</script>`, `%s%n{{.}}${x}`, "<img src=x onerror=1>"} {
			in := tenancy.BusinessContext{DisplayName: v, LegalName: sp(v), Description: sp(v), Industry: sp(v), Website: sp(srv.URL + "/inert")}
			p, err := svc.CreateBusinessProfile(ctx, w.A, in)
			if err != nil {
				t.Fatalf("create %s: %v", h(v), err)
			}
			got, err := svc.GetBusinessProfile(ctx, w.A, p.ID)
			if err != nil || got.Context.DisplayName != v || *got.Context.LegalName != v || *got.Context.Description != v || *got.Context.Industry != v ||
				*got.Context.Website != srv.URL+"/inert" {
				t.Fatalf("round trip of %s: %v", h(v), err)
			}
			if _, err := svc.UpdateBusinessProfile(ctx, w.A, p.ID, in); err != nil {
				t.Fatal(err)
			}
			if _, err := svc.ListGroupsForBusinessProfile(ctx, w.A, p.ID); err != nil {
				t.Fatal(err)
			}
		}
		if n := e.count(t, `SELECT count(*) FROM pg_class WHERE relnamespace = 'app'::regnamespace`); n != relations {
			t.Fatalf("app relations %d, want %d", n, relations)
		}
		if hits.Load() != 0 {
			t.Fatalf("the website server received %d requests", hits.Load())
		}
	})

	t.Run("IT-BP-LINK-1 shared and separate associations", func(t *testing.T) {
		if w.l1.ID == w.l2.ID || w.l1.ID == w.l3.ID || w.l2.ID == w.l3.ID {
			t.Fatal("association IDs repeat")
		}
		if w.l1.GroupID != w.g1.ID || w.l1.BusinessProfileID != w.p1.ID || w.l1.OrganizationID != w.A || w.l3.BusinessProfileID != w.p2.ID {
			t.Fatal("association columns differ")
		}
		got, err := svc.ListGroupsForBusinessProfile(ctx, w.A, w.p1.ID)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{w.g1.ID.String(), w.g2.ID.String()}
		slices.Sort(want)
		var gotIDs []string
		for _, a := range got {
			gotIDs = append(gotIDs, a.GroupID.String())
		}
		if strings.Join(gotIDs, ",") != strings.Join(want, ",") {
			t.Fatalf("Groups of p1: %d", len(got))
		}
		if _, p, err := svc.GetGroupBusinessProfile(ctx, w.A, w.g3.ID); err != nil || p.ID != w.p2.ID {
			t.Fatalf("g3's profile: %v", err)
		}
		if n := e.count(t, `SELECT count(*) FROM app.business_profiles WHERE id = $1::text::uuid`, w.p1.ID.String()); n != 1 {
			t.Fatalf("p1 rows: %d", n)
		}
		_, _, err = svc.GetGroupBusinessProfile(ctx, w.A, w.g4.ID)
		wantCode(t, "unassociated g4", err, tenancy.CodeAssociationNotFound)
	})

	t.Run("IT-BP-LINK-2 a Group has at most one profile", func(t *testing.T) {
		before := e.bpRows(t, w.A)
		_ = rec.all()
		_, err := rsvc.LinkGroupBusinessProfile(ctx, w.A, w.g1.ID, w.p2.ID)
		wantCode(t, "second profile", err, tenancy.CodeAssociationExists)
		_, err = rsvc.LinkGroupBusinessProfile(ctx, w.A, w.g1.ID, w.p1.ID)
		wantCode(t, "duplicate", err, tenancy.CodeAssociationExists)
		if ws := rec.writes(); len(ws) != 0 {
			t.Fatalf("%d writes reached the database", len(ws))
		}
		expectSQL(t, e.rt, "runtime second profile", "23505", "group_business_profiles_group_key",
			`INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ('`+w.A.String()+`', '`+w.g1.ID.String()+`', '`+w.p2.ID.String()+`')`)
		if e.bpRows(t, w.A) != before {
			t.Fatal("associations changed")
		}
		if a, _, err := svc.GetGroupBusinessProfile(ctx, w.A, w.g1.ID); err != nil || a.ID != w.l1.ID || a.BusinessProfileID != w.p1.ID {
			t.Fatal("the original association changed")
		}
	})

	t.Run("IT-BP-LINK-3 concurrent links of one Group", func(t *testing.T) {
		for round := range 20 {
			g, _ := e.group(t, w.A, w.R, "Race Group", fmt.Sprintf("race-group-%d", round))
			var ok, exists atomic.Int64
			var winner atomic.Value
			var wg sync.WaitGroup
			for i := range 32 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					p := w.p1.ID
					if i%2 == 1 {
						p = w.p2.ID
					}
					a, err := svc.LinkGroupBusinessProfile(ctx, w.A, g.ID, p)
					switch tenancy.CodeOf(err) {
					case "":
						ok.Add(1)
						winner.Store(a.BusinessProfileID)
					case tenancy.CodeAssociationExists:
						exists.Add(1)
					default:
						t.Errorf("round %d link: %v", round, err)
					}
				}()
			}
			wg.Wait()
			var stored string
			if err := e.own.QueryRow(ctx, `SELECT string_agg(business_profile_id::text, ',') FROM app.group_business_profiles WHERE group_id = $1::text::uuid`, g.ID.String()).Scan(&stored); err != nil {
				t.Fatal(err)
			}
			if ok.Load() != 1 || exists.Load() != 31 || winner.Load() == nil || stored != winner.Load().(tenancy.BusinessProfileID).String() {
				t.Fatalf("round %d: %d ok, %d exists, stored %s", round, ok.Load(), exists.Load(), h(stored))
			}
		}
	})

	t.Run("IT-BP-LINK-4 deterministic interleaving", func(t *testing.T) {
		for _, commit := range []bool{true, false} {
			g, _ := e.group(t, w.A, w.R, "Interleave Group", fmt.Sprintf("interleave-%v", commit))
			tx, err := e.runtimeConn(t).Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid)`,
				w.A.String(), g.ID.String(), w.p1.ID.String()); err != nil {
				t.Fatal(err)
			}
			res := make(chan error, 1)
			go func() { _, err := svc.LinkGroupBusinessProfile(ctx, w.A, g.ID, w.p2.ID); res <- err }()
			e.waitBlocked(t)
			if commit {
				err = tx.Commit(ctx)
			} else {
				err = tx.Rollback(ctx)
			}
			if err != nil {
				t.Fatal(err)
			}
			err = <-res
			want := tenancy.Code("")
			if commit {
				want = tenancy.CodeAssociationExists
			}
			if tenancy.CodeOf(err) != want {
				t.Fatalf("commit=%v: %v, want %q", commit, err, want)
			}
			wantProfile := w.p2.ID
			if commit {
				wantProfile = w.p1.ID
			}
			if a, _, err := svc.GetGroupBusinessProfile(ctx, w.A, g.ID); err != nil || a.BusinessProfileID != wantProfile {
				t.Fatalf("commit=%v: stored association %v", commit, err)
			}
		}
	})

	t.Run("IT-BP-LINK-5 unlink versus relink", func(t *testing.T) {
		check := func(label string, g tenancy.GroupID, unlinkErr, linkErr error) {
			t.Helper()
			if unlinkErr != nil {
				t.Fatalf("%s: unlink %v", label, unlinkErr)
			}
			var stored string
			if err := e.own.QueryRow(ctx, `SELECT coalesce(string_agg(business_profile_id::text, ','), '') FROM app.group_business_profiles WHERE group_id = $1::text::uuid`, g.String()).Scan(&stored); err != nil {
				t.Fatal(err)
			}
			switch {
			case linkErr == nil && stored == w.p2.ID.String():
			case tenancy.CodeOf(linkErr) == tenancy.CodeAssociationExists && stored == "":
			default:
				t.Fatalf("%s: link %v, stored %s", label, linkErr, h(stored))
			}
		}
		// Both deterministic orders.
		g, _ := e.group(t, w.A, w.R, "Relink Group", "relink-a")
		e.link(t, w.A, g.ID, w.p1.ID)
		uerr := svc.UnlinkGroupBusinessProfile(ctx, w.A, g.ID, w.p1.ID)
		_, lerr := svc.LinkGroupBusinessProfile(ctx, w.A, g.ID, w.p2.ID)
		check("unlink then link", g.ID, uerr, lerr)
		g, _ = e.group(t, w.A, w.R, "Relink Group", "relink-b")
		e.link(t, w.A, g.ID, w.p1.ID)
		_, lerr = svc.LinkGroupBusinessProfile(ctx, w.A, g.ID, w.p2.ID)
		uerr = svc.UnlinkGroupBusinessProfile(ctx, w.A, g.ID, w.p1.ID)
		check("link then unlink", g.ID, uerr, lerr)
		for round := range 50 {
			g, _ := e.group(t, w.A, w.R, "Relink Group", fmt.Sprintf("relink-r%d", round))
			e.link(t, w.A, g.ID, w.p1.ID)
			var wg sync.WaitGroup
			var uerr, lerr error
			first := rand.IntN(2)
			wg.Add(2)
			go func() {
				defer wg.Done()
				if first == 1 {
					time.Sleep(time.Duration(rand.IntN(500)) * time.Microsecond)
				}
				uerr = svc.UnlinkGroupBusinessProfile(ctx, w.A, g.ID, w.p1.ID)
			}()
			go func() {
				defer wg.Done()
				if first == 0 {
					time.Sleep(time.Duration(rand.IntN(500)) * time.Microsecond)
				}
				_, lerr = svc.LinkGroupBusinessProfile(ctx, w.A, g.ID, w.p2.ID)
			}()
			wg.Wait()
			check(fmt.Sprintf("round %d", round), g.ID, uerr, lerr)
		}
	})
	t.Run("IT-BP-XORG-1 Organization-scoped prevalidation", func(t *testing.T) {
		unknownG, _ := tenancy.ParseGroupID("0f8fad5b-d9cb-469f-a165-70867728950e")
		unknownP, _ := tenancy.ParseBusinessProfileID("0f8fad5b-d9cb-469f-a165-70867728950e")
		before := e.all(t, w.A, w.B)
		_ = rec.all()
		pairs := []struct {
			label            string
			foreign, unknown func() error
			want             tenancy.Code
		}{
			{"link foreign Group", func() error { _, err := rsvc.LinkGroupBusinessProfile(ctx, w.A, w.h1.ID, w.p1.ID); return err },
				func() error { _, err := rsvc.LinkGroupBusinessProfile(ctx, w.A, unknownG, w.p1.ID); return err }, tenancy.CodeGroupNotFound},
			{"link foreign profile", func() error { _, err := rsvc.LinkGroupBusinessProfile(ctx, w.A, w.g4.ID, w.q1.ID); return err },
				func() error { _, err := rsvc.LinkGroupBusinessProfile(ctx, w.A, w.g4.ID, unknownP); return err }, tenancy.CodeBusinessProfileNotFound},
			{"unlink foreign Group", func() error { return rsvc.UnlinkGroupBusinessProfile(ctx, w.A, w.h1.ID, w.q1.ID) },
				func() error { return rsvc.UnlinkGroupBusinessProfile(ctx, w.A, unknownG, w.q1.ID) }, tenancy.CodeGroupNotFound},
			{"unlink foreign profile", func() error { return rsvc.UnlinkGroupBusinessProfile(ctx, w.A, w.g4.ID, w.q1.ID) },
				func() error { return rsvc.UnlinkGroupBusinessProfile(ctx, w.A, w.g4.ID, unknownP) }, tenancy.CodeBusinessProfileNotFound},
			{"update foreign profile", func() error {
				_, err := rsvc.UpdateBusinessProfile(ctx, w.A, w.q1.ID, tenancy.BusinessContext{DisplayName: "x"})
				return err
			},
				func() error {
					_, err := rsvc.UpdateBusinessProfile(ctx, w.A, unknownP, tenancy.BusinessContext{DisplayName: "x"})
					return err
				}, tenancy.CodeBusinessProfileNotFound},
		}
		for _, c := range pairs {
			ef, eu := c.foreign(), c.unknown()
			wantCode(t, c.label, ef, c.want)
			if ef == nil || eu == nil || ef.Error() != eu.Error() {
				t.Errorf("%s: foreign and unknown errors differ", c.label)
			}
		}
		for _, s := range rec.writes() {
			if !strings.HasPrefix(s, "UPDATE app.business_profiles") {
				t.Errorf("a write reached the database: %s", h(s))
			}
		}
		if e.all(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-BP-XORG-2 runtime INSERT matrix", func(t *testing.T) {
		ga, _ := e.group(t, w.A, w.R, "Matrix A", "matrix-a")
		gb, _ := e.group(t, w.B, w.R, "Matrix B", "matrix-b")
		const unknown = "0f8fad5b-d9cb-469f-a165-70867728950e"
		orgs := map[string]string{"A": w.A.String(), "B": w.B.String(), "u": unknown}
		groups := map[string]string{"A": ga.ID.String(), "B": gb.ID.String(), "u": unknown}
		profiles := map[string]string{"A": w.p2.ID.String(), "B": w.q1.ID.String(), "u": unknown}
		fks := map[string]bool{"group_business_profiles_organization_fkey": true, "group_business_profiles_group_fkey": true,
			"group_business_profiles_business_profile_fkey": true}
		before := e.bpRows(t, w.A, w.B)
		inserted := 0
		for _, o := range []string{"A", "B", "u"} {
			for _, g := range []string{"A", "B", "u"} {
				for _, p := range []string{"A", "B", "u"} {
					tx, err := e.rt.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					_, err = tx.Exec(ctx, `INSERT INTO app.group_business_profiles (organization_id, group_id, business_profile_id) VALUES ($1::text::uuid, $2::text::uuid, $3::text::uuid)`,
						orgs[o], groups[g], profiles[p])
					_ = tx.Rollback(ctx)
					code, con := sqlErr(err)
					same := o == g && g == p && o != "u"
					switch {
					case same && err == nil:
						inserted++
					case !same && code == "23503" && fks[con]:
					default:
						t.Errorf("(%s,%s,%s): %s %s", o, g, p, code, con)
					}
				}
			}
		}
		if inserted != 2 || e.bpRows(t, w.A, w.B) != before {
			t.Fatalf("inserted %d (rolled back), rows changed: %v", inserted, e.bpRows(t, w.A, w.B) != before)
		}
	})

	t.Run("IT-BP-XORG-3 no retarget or reparent", func(t *testing.T) {
		before := e.all(t, w.A, w.B)
		l, p := w.l1.ID.String(), w.p1.ID.String()
		for _, col := range []string{"organization_id", "group_id", "business_profile_id"} {
			expectSQL(t, e.rt, "runtime "+col, "42501", "", `UPDATE app.group_business_profiles SET `+col+` = '`+w.h1.ID.String()+`' WHERE id = '`+l+`'`)
		}
		expectSQL(t, e.rt, "runtime profile Organization", "42501", "", `UPDATE app.business_profiles SET organization_id = '`+w.B.String()+`' WHERE id = '`+p+`'`)
		expectSQL(t, e.own, "owner moves a link to B", "23000", "group_business_profiles_keys_immutable",
			`UPDATE app.group_business_profiles SET organization_id = '`+w.B.String()+`', group_id = '`+w.h1.ID.String()+`', business_profile_id = '`+w.q1.ID.String()+`' WHERE id = '`+l+`'`)
		if e.all(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-BP-XORG-4 reads and writes stay in their Organization", func(t *testing.T) {
		unknownP, _ := tenancy.ParseBusinessProfileID("0f8fad5b-d9cb-469f-a165-70867728950e")
		before := e.all(t, w.B)
		for _, c := range []struct {
			label            string
			foreign, unknown func() error
		}{
			{"get", func() error { _, err := svc.GetBusinessProfile(ctx, w.A, w.q1.ID); return err },
				func() error { _, err := svc.GetBusinessProfile(ctx, w.A, unknownP); return err }},
			{"list Groups", func() error { _, err := svc.ListGroupsForBusinessProfile(ctx, w.A, w.q1.ID); return err },
				func() error { _, err := svc.ListGroupsForBusinessProfile(ctx, w.A, unknownP); return err }},
			{"access paths", func() error { _, err := svc.ListBusinessProfileAccessPaths(ctx, w.A, w.q1.ID); return err },
				func() error { _, err := svc.ListBusinessProfileAccessPaths(ctx, w.A, unknownP); return err }},
			{"update", func() error {
				_, err := svc.UpdateBusinessProfile(ctx, w.A, w.q1.ID, tenancy.BusinessContext{DisplayName: "x"})
				return err
			},
				func() error {
					_, err := svc.UpdateBusinessProfile(ctx, w.A, unknownP, tenancy.BusinessContext{DisplayName: "x"})
					return err
				}},
		} {
			ef, eu := c.foreign(), c.unknown()
			wantCode(t, c.label, ef, tenancy.CodeBusinessProfileNotFound)
			if ef == nil || eu == nil || ef.Error() != eu.Error() {
				t.Errorf("%s: foreign and unknown errors differ", c.label)
			}
		}
		_, _, err := svc.GetGroupBusinessProfile(ctx, w.A, w.h1.ID)
		wantCode(t, "foreign Group association", err, tenancy.CodeGroupNotFound)
		if e.all(t, w.B) != before {
			t.Fatal("B rows changed")
		}
	})

	t.Run("IT-BP-UNLINK-1 unlink removes only its association", func(t *testing.T) {
		bindings := e.rows(t, w.A, w.B)
		profiles := e.count(t, `SELECT count(*) FROM app.business_profiles WHERE organization_id = $1::text::uuid`, w.A.String())
		if err := svc.UnlinkGroupBusinessProfile(ctx, w.A, w.g1.ID, w.p1.ID); err != nil {
			t.Fatal(err)
		}
		_, _, err := svc.GetGroupBusinessProfile(ctx, w.A, w.g1.ID)
		wantCode(t, "g1 after unlink", err, tenancy.CodeAssociationNotFound)
		if a, _, err := svc.GetGroupBusinessProfile(ctx, w.A, w.g2.ID); err != nil || a.ID != w.l2.ID {
			t.Fatal("g2's association changed")
		}
		if e.rows(t, w.A, w.B) != bindings || e.count(t, `SELECT count(*) FROM app.business_profiles WHERE organization_id = $1::text::uuid`, w.A.String()) != profiles {
			t.Fatal("profiles, Groups or bindings changed by an unlink")
		}
		before := e.bpRows(t, w.A)
		wantCode(t, "unlink the wrong profile", svc.UnlinkGroupBusinessProfile(ctx, w.A, w.g2.ID, w.p2.ID), tenancy.CodeAssociationNotFound)
		wantCode(t, "unlink again", svc.UnlinkGroupBusinessProfile(ctx, w.A, w.g1.ID, w.p1.ID), tenancy.CodeAssociationNotFound)
		if e.bpRows(t, w.A) != before {
			t.Fatal("a failed unlink changed rows")
		}
		w.l1 = e.link(t, w.A, w.g1.ID, w.p1.ID)
	})

	t.Run("IT-BP-UNLINK-2 concurrent unlinks", func(t *testing.T) {
		g, _ := e.group(t, w.A, w.R, "Unlink Race", "unlink-race")
		e.link(t, w.A, g.ID, w.p2.ID)
		var ok, missing atomic.Int64
		var wg sync.WaitGroup
		for range 16 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				switch err := svc.UnlinkGroupBusinessProfile(ctx, w.A, g.ID, w.p2.ID); tenancy.CodeOf(err) {
				case "":
					ok.Add(1)
				case tenancy.CodeAssociationNotFound:
					missing.Add(1)
				default:
					t.Errorf("unlink: %v", err)
				}
			}()
		}
		wg.Wait()
		if ok.Load() != 1 || missing.Load() != 15 {
			t.Fatalf("%d unlinked, %d not found", ok.Load(), missing.Load())
		}
	})

	t.Run("IT-BP-ATOM-1 rejected inputs change nothing", func(t *testing.T) {
		before := e.all(t, w.A, w.B)
		var zero tenancy.GroupID
		for name, c := range map[string]struct {
			call func() error
			want tenancy.Code
		}{
			"bad context": {func() error { _, err := svc.CreateBusinessProfile(ctx, w.A, tenancy.BusinessContext{}); return err }, tenancy.CodeInvalidDisplayName},
			"bad website": {func() error {
				_, err := svc.UpdateBusinessProfile(ctx, w.A, w.p1.ID, tenancy.BusinessContext{DisplayName: "x", Website: sp("ftp://x")})
				return err
			}, tenancy.CodeInvalidWebsite},
			"zero Group": {func() error { _, err := svc.LinkGroupBusinessProfile(ctx, w.A, zero, w.p1.ID); return err }, tenancy.CodeInvalidID},
			"conflict":   {func() error { _, err := svc.LinkGroupBusinessProfile(ctx, w.A, w.g3.ID, w.p1.ID); return err }, tenancy.CodeAssociationExists},
			"not linked": {func() error { return svc.UnlinkGroupBusinessProfile(ctx, w.A, w.g4.ID, w.p1.ID) }, tenancy.CodeAssociationNotFound},
			"foreign":    {func() error { _, err := svc.LinkGroupBusinessProfile(ctx, w.B, w.h1.ID, w.p1.ID); return err }, tenancy.CodeBusinessProfileNotFound},
		} {
			wantCode(t, name, c.call(), c.want)
		}
		if e.all(t, w.A, w.B) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-BP-ATOM-2 injected store failures", func(t *testing.T) {
		u := e.user(t, false)
		before := e.all(t, w.A)
		closed, err := pgxpool.New(ctx, e.pool.Config().ConnString())
		if err != nil {
			t.Fatal(err)
		}
		closed.Close()
		for name, db := range map[string]func(prefix string) *faultDB{
			"commit failure":    func(string) *faultDB { return &faultDB{pool: e.pool, failCommit: true} },
			"statement failure": func(p string) *faultDB { return &faultDB{pool: e.pool, failOn: p} },
			"closed pool":       func(string) *faultDB { return &faultDB{pool: e.pool, closed: closed} },
		} {
			_, err := tenancy.NewService(db("INSERT INTO app.business_profiles"), nil).CreateBusinessProfile(ctx, w.A, tenancy.BusinessContext{DisplayName: "Fault"})
			wantCode(t, name+" create", err, tenancy.CodeUnavailable)
			_, err = tenancy.NewService(db("UPDATE app.business_profiles"), nil).UpdateBusinessProfile(ctx, w.A, w.p1.ID, tenancy.BusinessContext{DisplayName: "Fault"})
			wantCode(t, name+" update", err, tenancy.CodeUnavailable)
			_, err = tenancy.NewService(db("INSERT INTO app.group_business_profiles"), nil).LinkGroupBusinessProfile(ctx, w.A, w.g4.ID, w.p1.ID)
			wantCode(t, name+" link", err, tenancy.CodeUnavailable)
			err = tenancy.NewService(db("DELETE FROM app.group_business_profiles"), nil).UnlinkGroupBusinessProfile(ctx, w.A, w.g2.ID, w.p1.ID)
			wantCode(t, name+" unlink", err, tenancy.CodeUnavailable)
			// The G005 Group role operations at version 4.
			_, err = tenancy.NewService(db("INSERT INTO app.role_bindings"), nil).GrantGroupRole(ctx, w.A, w.g4.ID, u, tenancy.GroupMember)
			wantCode(t, name+" grant", err, tenancy.CodeUnavailable)
			_, err = tenancy.NewService(db("UPDATE app.role_bindings"), nil).ChangeGroupRole(ctx, w.A, w.g2.ID, w.V, tenancy.GroupAdmin)
			wantCode(t, name+" change", err, tenancy.CodeUnavailable)
			err = tenancy.NewService(db("DELETE FROM app.role_bindings"), nil).RemoveGroupMembership(ctx, w.A, w.g2.ID, w.V)
			wantCode(t, name+" remove", err, tenancy.CodeUnavailable)
		}
		// Cancellation mid-transaction: every mutation starts with an Organization precheck, after which the context is
		// canceled, so the mutation statement itself never runs.
		for name, op := range map[string]func(context.Context, *tenancy.Service) error{
			"create": func(c context.Context, s *tenancy.Service) error {
				_, err := s.CreateBusinessProfile(c, w.A, tenancy.BusinessContext{DisplayName: "Fault"})
				return err
			},
			"update": func(c context.Context, s *tenancy.Service) error {
				_, err := s.UpdateBusinessProfile(c, w.A, w.p1.ID, tenancy.BusinessContext{DisplayName: "Fault"})
				return err
			},
			"link": func(c context.Context, s *tenancy.Service) error {
				_, err := s.LinkGroupBusinessProfile(c, w.A, w.g4.ID, w.p1.ID)
				return err
			},
			"unlink": func(c context.Context, s *tenancy.Service) error {
				return s.UnlinkGroupBusinessProfile(c, w.A, w.g2.ID, w.p1.ID)
			},
			"grant": func(c context.Context, s *tenancy.Service) error {
				_, err := s.GrantGroupRole(c, w.A, w.g4.ID, u, tenancy.GroupMember)
				return err
			},
			"change": func(c context.Context, s *tenancy.Service) error {
				_, err := s.ChangeGroupRole(c, w.A, w.g2.ID, w.V, tenancy.GroupAdmin)
				return err
			},
			"remove": func(c context.Context, s *tenancy.Service) error {
				return s.RemoveGroupMembership(c, w.A, w.g2.ID, w.V)
			},
		} {
			cctx, cancel := context.WithCancel(ctx)
			err := op(cctx, tenancy.NewService(&faultDB{pool: e.pool, after: "SELECT EXISTS", cancel: cancel}, nil))
			cancel()
			wantCode(t, name+" cancellation mid-transaction", err, tenancy.CodeUnavailable)
		}
		if e.all(t, w.A) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-BP-REF-1 RESTRICT on referenced rows", func(t *testing.T) {
		before := e.all(t, w.A)
		expectSQL(t, e.own, "delete a referenced profile", "23503", "group_business_profiles_business_profile_fkey",
			`DELETE FROM app.business_profiles WHERE id = '`+w.p1.ID.String()+`'`)
		// A Group without bindings but with an association.
		var bare string
		if err := e.own.QueryRow(ctx, `INSERT INTO app.groups (organization_id, name, slug) VALUES ($1::text::uuid, 'Bare Group', 'bare-group') RETURNING id::text`, w.A.String()).Scan(&bare); err != nil {
			t.Fatal(err)
		}
		bareID, _ := tenancy.ParseGroupID(bare)
		e.link(t, w.A, bareID, w.p2.ID)
		expectSQL(t, e.own, "delete a referenced Group", "23503", "group_business_profiles_group_fkey", `DELETE FROM app.groups WHERE id = '`+bare+`'`)
		org := e.org(t, "Referenced Organization").ID
		e.profile(t, org, "Only Profile")
		expectSQL(t, e.own, "delete a referenced Organization", "23503", "business_profiles_organization_fkey",
			`DELETE FROM app.organizations WHERE id = '`+org.String()+`'`)
		if err := svc.UnlinkGroupBusinessProfile(ctx, w.A, bareID, w.p2.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := e.own.Exec(ctx, `DELETE FROM app.groups WHERE id = $1::text::uuid`, bare); err != nil {
			t.Fatalf("delete the bare Group after unlink: %v", err)
		}
		if e.all(t, w.A) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-BP-REF-2 profile delete races link", func(t *testing.T) {
		race := func(label string, deleteFirst bool) {
			p := e.profile(t, w.A, "Race Profile")
			g, _ := e.group(t, w.A, w.R, "Race Link", "race-link-"+strings.ReplaceAll(label, " ", "-"))
			var derr, lerr error
			var wg sync.WaitGroup
			conn := e.ownerConn(t)
			wg.Add(2)
			go func() {
				defer wg.Done()
				if !deleteFirst {
					time.Sleep(time.Duration(rand.IntN(400)) * time.Microsecond)
				}
				_, derr = conn.Exec(ctx, `DELETE FROM app.business_profiles WHERE id = $1::text::uuid`, p.ID.String())
			}()
			go func() {
				defer wg.Done()
				if deleteFirst {
					time.Sleep(time.Duration(rand.IntN(400)) * time.Microsecond)
				}
				_, lerr = svc.LinkGroupBusinessProfile(ctx, w.A, g.ID, p.ID)
			}()
			wg.Wait()
			profiles := e.count(t, `SELECT count(*) FROM app.business_profiles WHERE id = $1::text::uuid`, p.ID.String())
			links := e.count(t, `SELECT count(*) FROM app.group_business_profiles WHERE business_profile_id = $1::text::uuid`, p.ID.String())
			dcode, dcon := sqlErr(derr)
			switch {
			case derr == nil && tenancy.CodeOf(lerr) == tenancy.CodeBusinessProfileNotFound && profiles == 0 && links == 0:
			case lerr == nil && dcode == "23503" && dcon == "group_business_profiles_business_profile_fkey" && profiles == 1 && links == 1:
			default:
				t.Fatalf("%s: delete %s %s, link %v, profiles %d, links %d", label, dcode, dcon, lerr, profiles, links)
			}
		}
		// Deterministic: the link committed first; the profile deleted first.
		p := e.profile(t, w.A, "Linked First")
		g, _ := e.group(t, w.A, w.R, "Linked First", "linked-first")
		e.link(t, w.A, g.ID, p.ID)
		expectSQL(t, e.own, "delete after a committed link", "23503", "group_business_profiles_business_profile_fkey",
			`DELETE FROM app.business_profiles WHERE id = '`+p.ID.String()+`'`)
		p = e.profile(t, w.A, "Deleted First")
		if _, err := e.own.Exec(ctx, `DELETE FROM app.business_profiles WHERE id = $1::text::uuid`, p.ID.String()); err != nil {
			t.Fatal(err)
		}
		_, err := svc.LinkGroupBusinessProfile(ctx, w.A, w.g4.ID, p.ID)
		wantCode(t, "link after a committed delete", err, tenancy.CodeBusinessProfileNotFound)
		for round := range 50 {
			race(fmt.Sprintf("round %d", round), round%2 == 0)
		}
	})

	t.Run("IT-BP-ACC-2 Group role changes change only the derived facts", func(t *testing.T) {
		profiles := e.bpRows(t, w.A)
		paths := func() []string {
			ps, err := svc.ListUserBusinessProfileAccessPaths(ctx, w.A, w.U)
			if err != nil {
				t.Fatal(err)
			}
			var out []string
			for _, p := range ps {
				if p.BusinessProfileID == w.p1.ID {
					out = append(out, p.GroupID.String()+" "+string(p.GroupRole))
				}
			}
			slices.Sort(out)
			return out
		}
		g1, g2 := w.g1.ID.String(), w.g2.ID.String()
		steps := []struct {
			label string
			do    func() error
			want  []string
		}{
			{"grant U group_admin@g2", func() error { _, err := svc.GrantGroupRole(ctx, w.A, w.g2.ID, w.U, tenancy.GroupAdmin); return err },
				[]string{g1 + " group_owner", g2 + " group_admin"}},
			{"remove U@g1", func() error { return svc.RemoveGroupMembership(ctx, w.A, w.g1.ID, w.U) }, []string{g2 + " group_admin"}},
			{"downgrade U@g2", func() error { _, err := svc.ChangeGroupRole(ctx, w.A, w.g2.ID, w.U, tenancy.GroupMember); return err },
				[]string{g2 + " group_member"}},
			{"remove U@g2", func() error { return svc.RemoveGroupMembership(ctx, w.A, w.g2.ID, w.U) }, nil},
		}
		slices.Sort(steps[0].want)
		if got := paths(); strings.Join(got, ",") != g1+" group_owner" {
			t.Fatalf("U's p1 paths before: %v", got)
		}
		for _, s := range steps {
			_ = rec.all()
			if err := s.do(); err != nil {
				t.Fatalf("%s: %v", s.label, err)
			}
			if got := paths(); strings.Join(got, ",") != strings.Join(s.want, ",") {
				t.Fatalf("after %s: %v, want %v", s.label, got, s.want)
			}
			if e.bpRows(t, w.A) != profiles {
				t.Fatalf("%s changed a profile or an association", s.label)
			}
		}
		// Restore U's ownership of g1 for the later cases.
		if _, err := svc.GrantGroupRole(ctx, w.A, w.g1.ID, w.U, tenancy.GroupOwner); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("IT-BP-ACC-3 unlink and relink change the facts only", func(t *testing.T) {
		g3, p2 := w.g3.ID, w.p2.ID
		has := func(u tenancy.UserID) bool {
			ps, err := svc.ListBusinessProfileAccessPaths(ctx, w.A, p2)
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range ps {
				if p.UserID == u && p.GroupID == g3 {
					return true
				}
			}
			return false
		}
		bindings := e.rows(t, w.A)
		profile := e.count(t, `SELECT count(*) FROM app.business_profiles WHERE id = $1::text::uuid AND display_name = $2`, p2.String(), sharedName)
		if !has(w.U) || !has(w.Y) {
			t.Fatal("U and Y lack their p2 paths")
		}
		if err := svc.UnlinkGroupBusinessProfile(ctx, w.A, g3, p2); err != nil {
			t.Fatal(err)
		}
		if has(w.U) || has(w.Y) {
			t.Fatal("paths survive the unlink")
		}
		w.l3 = e.link(t, w.A, g3, p2)
		if !has(w.U) || !has(w.Y) {
			t.Fatal("paths do not reappear after the relink")
		}
		if e.rows(t, w.A) != bindings || e.count(t, `SELECT count(*) FROM app.business_profiles WHERE id = $1::text::uuid AND display_name = $2`, p2.String(), sharedName) != profile {
			t.Fatal("bindings or the profile changed")
		}
	})

	t.Run("IT-BP-ACC-4 Organization authority is not a profile path", func(t *testing.T) {
		memberships := func() string {
			return fmt.Sprint(e.count(t, `SELECT count(*) FROM app.organization_memberships WHERE organization_id = $1::text::uuid`, w.A.String()), "/",
				e.count(t, `SELECT count(*) FROM app.group_memberships WHERE organization_id = $1::text::uuid`, w.A.String())) + "|" + e.rows(t, w.A)
		}
		before := memberships()
		if ps, err := svc.ListUserBusinessProfileAccessPaths(ctx, w.A, w.W); err != nil || len(ps) != 0 {
			t.Fatalf("W: %d paths %v", len(ps), err)
		}
		p := e.profile(t, w.A, "Unattached Profile")
		if ps, err := svc.ListBusinessProfileAccessPaths(ctx, w.A, p.ID); err != nil || len(ps) != 0 {
			t.Fatalf("unattached profile: %d paths %v", len(ps), err)
		}
		if _, err := svc.UpdateBusinessProfile(ctx, w.A, p.ID, tenancy.BusinessContext{DisplayName: "Unattached Renamed"}); err != nil {
			t.Fatal(err)
		}
		e.link(t, w.A, w.g4.ID, p.ID)
		if err := svc.UnlinkGroupBusinessProfile(ctx, w.A, w.g4.ID, p.ID); err != nil {
			t.Fatal(err)
		}
		if memberships() != before {
			t.Fatal("membership counts or hashes changed by profile operations")
		}
	})

	t.Run("IT-BP-ACC-5 cross-tenant and sibling isolation", func(t *testing.T) {
		unknownP, _ := tenancy.ParseBusinessProfileID("0f8fad5b-d9cb-469f-a165-70867728950e")
		_, ef := svc.ListBusinessProfileAccessPaths(ctx, w.A, w.q1.ID)
		_, eu := svc.ListBusinessProfileAccessPaths(ctx, w.A, unknownP)
		wantCode(t, "foreign profile paths", ef, tenancy.CodeBusinessProfileNotFound)
		if ef == nil || eu == nil || ef.Error() != eu.Error() {
			t.Fatal("foreign and unknown errors differ")
		}
		if ps, err := svc.ListUserBusinessProfileAccessPaths(ctx, w.A, w.Z); err != nil || len(ps) != 0 {
			t.Fatalf("Z in A: %d %v", len(ps), err)
		}
		_, err := svc.GetGroupMembership(ctx, w.A, w.g1.ID, w.V)
		wantCode(t, "V in g1", err, tenancy.CodeMembershipNotFound)
		for _, p := range func() []tenancy.BusinessProfileAccessPath {
			ps, _ := svc.ListUserBusinessProfileAccessPaths(ctx, w.A, w.V)
			return ps
		}() {
			if p.GroupID == w.g1.ID {
				t.Fatal("V reaches p1 through g1")
			}
		}
	})

	t.Run("IT-BP-ACC-6 the access paths are read-only", func(t *testing.T) {
		// A join view is not updatable: PostgreSQL's rewriter refuses the write (55000) before any privilege check, and
		// CT-BP-3 proves that no write privilege exists either.
		before := e.all(t, w.A)
		for what, stmt := range map[string]string{
			"INSERT": `INSERT INTO app.business_profile_access_paths (user_id) VALUES (gen_random_uuid())`,
			"UPDATE": `UPDATE app.business_profile_access_paths SET group_role = 'group_owner'`,
			"DELETE": `DELETE FROM app.business_profile_access_paths`,
		} {
			expectSQL(t, e.rt, "runtime "+what, "55000", "", stmt)
		}
		if e.all(t, w.A) != before {
			t.Fatal("rows changed")
		}
	})

	t.Run("IT-BP-ACC-7 readers see committed bindings only", func(t *testing.T) {
		has := func() bool {
			ps, err := svc.ListUserBusinessProfileAccessPaths(ctx, w.A, w.U)
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range ps {
				if p.GroupID == w.g1.ID {
					return true
				}
			}
			return false
		}
		tx, err := e.runtimeConn(t).Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid AND scope_type = 'group' AND scope_id = $3::text::uuid`,
			w.A.String(), w.U.String(), w.g1.ID.String()); err != nil {
			t.Fatal(err)
		}
		if !has() {
			t.Fatal("an uncommitted removal is visible")
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if has() {
			t.Fatal("a committed removal is not visible on the next read")
		}
		if _, err := svc.GrantGroupRole(ctx, w.A, w.g1.ID, w.U, tenancy.GroupOwner); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("IT-BP-INDEP-1 no cross-writes", func(t *testing.T) {
		u := e.user(t, false)
		bindings := func() string {
			var s string
			if err := e.own.QueryRow(ctx, `SELECT coalesce(md5(string_agg(b::text, E'\n' ORDER BY b.id)), '-') FROM app.role_bindings b WHERE b.organization_id = $1::text::uuid`, w.A.String()).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
		_ = rec.all()
		b0 := bindings()
		p, err := rsvc.CreateBusinessProfile(ctx, w.A, tenancy.BusinessContext{DisplayName: "Independent"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := rsvc.UpdateBusinessProfile(ctx, w.A, p.ID, tenancy.BusinessContext{DisplayName: "Independent 2"}); err != nil {
			t.Fatal(err)
		}
		if _, err := rsvc.LinkGroupBusinessProfile(ctx, w.A, w.g4.ID, p.ID); err != nil {
			t.Fatal(err)
		}
		if err := rsvc.UnlinkGroupBusinessProfile(ctx, w.A, w.g4.ID, p.ID); err != nil {
			t.Fatal(err)
		}
		for _, s := range rec.writes() {
			if !strings.Contains(s, "app.business_profiles") && !strings.Contains(s, "app.group_business_profiles") {
				t.Errorf("a G006 operation wrote outside its tables: %s", h(s))
			}
		}
		if bindings() != b0 {
			t.Fatal("G006 operations changed role_bindings")
		}
		profiles := e.bpRows(t, w.A)
		if _, err := rsvc.GrantGroupRole(ctx, w.A, w.g1.ID, u, tenancy.GroupMember); err != nil {
			t.Fatal(err)
		}
		if _, err := rsvc.ChangeGroupRole(ctx, w.A, w.g1.ID, u, tenancy.GroupAdmin); err != nil {
			t.Fatal(err)
		}
		if err := rsvc.RemoveGroupMembership(ctx, w.A, w.g1.ID, u); err != nil {
			t.Fatal(err)
		}
		for _, s := range rec.all() {
			if strings.Contains(s, "business_profile") {
				t.Errorf("a G005 operation touched a G006 relation: %s", h(s))
			}
		}
		if e.bpRows(t, w.A) != profiles {
			t.Fatal("G005 role operations changed profiles or associations")
		}
	})

	t.Run("IT-BP-INDEP-2 the creator's path comes from the existing binding", func(t *testing.T) {
		creator := e.user(t, false)
		g, owner := e.group(t, w.A, creator, "Creator Group", "creator-group")
		n := e.count(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid`, w.A.String())
		e.link(t, w.A, g.ID, w.p2.ID)
		if e.count(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid`, w.A.String()) != n {
			t.Fatal("linking changed the binding count")
		}
		ps, err := svc.ListUserBusinessProfileAccessPaths(ctx, w.A, creator)
		if err != nil || len(ps) != 1 || ps[0].GroupID != g.ID || ps[0].GroupRole != tenancy.GroupOwner || ps[0].BusinessProfileID != w.p2.ID ||
			owner.Role != string(tenancy.GroupOwner) {
			t.Fatalf("creator paths: %v %v", pathKeys(ps), err)
		}
	})

	t.Run("IT-BP-LOG-1 operation records", func(t *testing.T) {
		start := len(e.logs.String())
		c0, c1, c2 := e.canaries[0], e.canaries[1], e.canaries[2]
		g, _ := e.group(t, w.A, w.R, "Log Group", "log-group")
		var created []string // every UUID created inside the handler
		handler := observability.Middleware(slog.New(slog.DiscardHandler), http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			p, err := svc.CreateBusinessProfile(ctx, w.A, tenancy.BusinessContext{DisplayName: c0, LegalName: sp(c1),
				Description: sp(c1), Industry: sp(c2), Website: sp("https://" + c0 + ".example/" + c2)}) // ok: a canary in every field
			if err != nil {
				t.Errorf("create: %v", err)
			}
			created = append(created, p.ID.String())
			_, _ = svc.CreateBusinessProfile(ctx, w.A, tenancy.BusinessContext{DisplayName: c1 + cc(0)})                // rejected: invalid_display_name
			_, _ = svc.UpdateBusinessProfile(ctx, w.A, p.ID, tenancy.BusinessContext{DisplayName: c2, Website: sp(c2)}) // rejected: invalid_website
			if _, err := svc.UpdateBusinessProfile(ctx, w.A, p.ID, tenancy.BusinessContext{DisplayName: c2, LegalName: sp(c0),
				Description: sp(c2), Industry: sp(c1), Website: sp("http://" + c1 + ".example")}); err != nil { // ok
				t.Errorf("update: %v", err)
			}
			a, err := svc.LinkGroupBusinessProfile(ctx, w.A, g.ID, p.ID) // ok
			if err != nil {
				t.Errorf("link: %v", err)
			}
			created = append(created, a.ID.String())
			// The five reads write no record.
			_, _ = svc.GetBusinessProfile(ctx, w.A, p.ID)
			_, _, _ = svc.GetGroupBusinessProfile(ctx, w.A, g.ID)
			_, _ = svc.ListGroupsForBusinessProfile(ctx, w.A, p.ID)
			_, _ = svc.ListBusinessProfileAccessPaths(ctx, w.A, p.ID)
			_, _ = svc.ListUserBusinessProfileAccessPaths(ctx, w.A, w.U)
			_, _ = svc.LinkGroupBusinessProfile(ctx, w.A, g.ID, w.p1.ID) // rejected: association_exists
			_ = svc.UnlinkGroupBusinessProfile(ctx, w.A, g.ID, w.p1.ID)  // rejected: association_not_found
			_ = svc.UnlinkGroupBusinessProfile(ctx, w.A, g.ID, p.ID)     // ok
			failing := tenancy.NewService(&faultDB{pool: e.pool, failCommit: true}, observability.NewLogger(e.logs, slog.LevelInfo, nil))
			_, _ = failing.CreateBusinessProfile(ctx, w.A, tenancy.BusinessContext{DisplayName: c0}) // failed: unavailable
			_, _ = failing.LinkGroupBusinessProfile(ctx, w.A, g.ID, p.ID)                            // failed: unavailable
			rw.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/probe", nil)
		req.Header.Set("X-Request-Id", "req-profile-0001")
		req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
		handler.ServeHTTP(httptest.NewRecorder(), req)
		type lrec struct {
			Level, Event, Operation, Outcome string
			ScopeType                        string `json:"scope_type"`
			ErrorCode                        string `json:"error_code"`
			RequestID                        string `json:"request_id"`
			TraceID                          string `json:"trace_id"`
			Service                          string
		}
		var got []lrec
		for _, l := range strings.Split(strings.TrimSpace(e.logs.String()[start:]), "\n") {
			var r lrec
			if err := json.Unmarshal([]byte(l), &r); err != nil {
				t.Fatalf("log line is not JSON: %v", err)
			}
			if r.Operation == "group.create" {
				continue // the fixture Group above
			}
			got = append(got, r)
		}
		const rid, tid = "req-profile-0001", "0af7651916cd43dd8448eb211c80319c"
		want := []lrec{
			{"info", "tenancy.operation", "business_profile.create", "ok", "", "", rid, tid, "core"},
			{"warn", "tenancy.operation", "business_profile.create", "rejected", "", "invalid_display_name", rid, tid, "core"},
			{"warn", "tenancy.operation", "business_profile.update", "rejected", "", "invalid_website", rid, tid, "core"},
			{"info", "tenancy.operation", "business_profile.update", "ok", "", "", rid, tid, "core"},
			{"info", "tenancy.operation", "association.link", "ok", "", "", rid, tid, "core"},
			{"warn", "tenancy.operation", "association.link", "rejected", "", "association_exists", rid, tid, "core"},
			{"warn", "tenancy.operation", "association.unlink", "rejected", "", "association_not_found", rid, tid, "core"},
			{"info", "tenancy.operation", "association.unlink", "ok", "", "", rid, tid, "core"},
			{"error", "tenancy.operation", "business_profile.create", "failed", "", "unavailable", rid, tid, "core"},
			{"error", "tenancy.operation", "association.link", "failed", "", "unavailable", rid, tid, "core"},
		}
		if !slices.Equal(got, want) {
			t.Fatalf("records:\n got %+v\nwant %+v", got, want)
		}
		logs := e.logs.String()
		for _, c := range e.canaries {
			if strings.Contains(logs, c) {
				t.Fatal("a canary reached the operation log")
			}
		}
		if len(created) != 2 {
			t.Fatalf("created IDs: %d", len(created))
		}
		for _, id := range append(created, g.ID.String(), w.p1.ID.String(), w.A.String(), w.U.String()) {
			if strings.Contains(logs, id) {
				t.Fatal("a UUID reached the operation log")
			}
		}
	})
}
