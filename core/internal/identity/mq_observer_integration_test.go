//go:build integration

// The least-privilege observer against a real database (G057 Plan r5 §3.7, b2): the same mqBaseline / mqNewSessions /
// mqWaitNewSession / mqBindTenant path the workos_manual steps call, executed as <db>_auth_reader with row-level
// security in force, plus the negative privilege probes that prove the grant boundary the design respects. Fixtures
// are written by the attested superuser setup connection (integration_test.go); the assertions read only through the
// observer helpers, except where a probe deliberately names an ungranted column to show it is refused.
package identity_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// mqProbe runs one statement in the observer's transaction shape (BEGIN READ ONLY; SET LOCAL ROLE reader) without the
// registry guard and returns its SQLSTATE ("" when it succeeded). It is the negative-probe path, never mqRead.
func mqProbe(t *testing.T, pool *pgxpool.Pool, reader, stmt string, args ...any) string {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE "+pgx.Identifier{reader}.Sanitize()); err != nil {
		t.Fatalf("SET LOCAL ROLE %s: %v", reader, err)
	}
	rows, err := tx.Query(ctx, stmt, args...)
	if err == nil {
		rows.Close()
		err = rows.Err()
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	if err != nil {
		t.Fatalf("probe %q: %v", stmt, err)
	}
	return ""
}

// mqFixtures writes synthetic users, Organizations and sessions through the superuser setup connection and removes
// its Organizations afterwards (resetIdentity empties the identity tables; the fixture Organizations would otherwise
// keep their labels).
type mqFixtures struct {
	t    *testing.T
	conn *pgx.Conn
	orgs []string
	n    int
}

func newMQFixtures(t *testing.T) *mqFixtures {
	t.Helper()
	f := &mqFixtures{t: t, conn: setupConn(t)}
	t.Cleanup(func() {
		ctx := context.Background()
		for _, id := range f.orgs {
			_, _ = f.conn.Exec(ctx, `DELETE FROM app.sessions WHERE organization_id = $1::text::uuid`, id)
			_, _ = f.conn.Exec(ctx, `DELETE FROM app.organizations WHERE id = $1::text::uuid`, id)
		}
	})
	return f
}

func (f *mqFixtures) user() string {
	f.t.Helper()
	var id string
	if err := f.conn.QueryRow(context.Background(), `INSERT INTO app.users (display_name) VALUES ('Synthetic Observer User') RETURNING id::text`).Scan(&id); err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *mqFixtures) org(label string) string {
	f.t.Helper()
	var id string
	if err := f.conn.QueryRow(context.Background(), `INSERT INTO app.organizations (name, platform_label) VALUES ('Org ' || $1, $1) RETURNING id::text`,
		label).Scan(&id); err != nil {
		f.t.Fatal(err)
	}
	f.orgs = append(f.orgs, id)
	return id
}

// session inserts one browser session bound to origin and org with the given auth_time (created_at and renewed_at
// equal to it), the shape of identity/store.go's issuance INSERT.
func (f *mqFixtures) session(user, origin, org string, at time.Time) string {
	f.t.Helper()
	f.n++
	tag := fmt.Sprintf("mq-observer-%d-%d", time.Now().UnixNano(), f.n)
	var id string
	if err := f.conn.QueryRow(context.Background(), `
INSERT INTO app.sessions (user_id, kind, credential_hash, csrf_token, auth_time, created_at, renewed_at, idle_expires_at,
                          absolute_expires_at, origin, organization_id)
VALUES ($1::text::uuid, 'browser', sha256(convert_to($2 || '-cred', 'UTF8')), sha256(convert_to($2 || '-csrf', 'UTF8')), $3::timestamptz,
        $3::timestamptz, $3::timestamptz, $3::timestamptz + interval '30 minutes', $3::timestamptz + interval '8 hours', $4, $5::text::uuid)
RETURNING id::text`, user, tag, at, origin, org).Scan(&id); err != nil {
		f.t.Fatal(err)
	}
	return id
}

// renew is the store.go renewal statement shape: renewed_at and idle_expires_at only; the id never changes.
func (f *mqFixtures) renew(id string, now time.Time) {
	f.t.Helper()
	tag, err := f.conn.Exec(context.Background(), `UPDATE app.sessions SET idle_expires_at = LEAST($3::timestamptz, absolute_expires_at), renewed_at = $2
WHERE id = $1::text::uuid AND revoked_at IS NULL AND idle_expires_at > $2`, id, now, now.Add(30*time.Minute))
	if err != nil || tag.RowsAffected() != 1 {
		f.t.Fatalf("renew: %v (%d rows)", err, tag.RowsAffected())
	}
}

func (f *mqFixtures) revoke(id string, now time.Time) {
	f.t.Helper()
	tag, err := f.conn.Exec(context.Background(), `UPDATE app.sessions SET revoked_at = $2, revoke_reason = 'logout' WHERE id = $1::text::uuid AND revoked_at IS NULL`, id, now)
	if err != nil || tag.RowsAffected() != 1 {
		f.t.Fatalf("revoke: %v (%d rows)", err, tag.RowsAffected())
	}
}

func mqIDs(rows []mqSessionState) string {
	s := ""
	for _, r := range rows {
		s += r.ID + ";"
	}
	return s
}

func TestMQObserverReadsUnderAuthReader(t *testing.T) {
	resetIdentity(t)
	seedTenants(t)
	ctx := context.Background()
	reader := it.cfg.Database + "_auth_reader"
	pool, err := mqMigratorPool(ctx, it.cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := mqPreflight(ctx, pool, reader); err != nil {
		t.Fatalf("preflight: %v", err)
	}
	fx := newMQFixtures(t)
	p := mqOrigin{Role: "P", Origin: "https://localhost:543", Host: "localhost", Port: "543"}
	now := time.Now().UTC().Truncate(time.Microsecond)
	u1, u2 := fx.user(), fx.user()
	s1 := fx.session(u1, originA, orgAID, now.Add(-10*time.Minute))
	s2 := fx.session(u2, originB, orgBID, now.Add(-9*time.Minute))
	s3 := fx.session(u1, originB, orgBID, now.Add(-8*time.Minute))

	// The baseline reads every user's rows under the reader while FORCE row-level security is on (the preflight).
	base, err := mqBaseline(ctx, pool, reader)
	if err != nil {
		t.Fatalf("baseline: %v", err)
	}
	for _, c := range []struct{ id, user, origin, org string }{{s1, u1, originA, orgAID}, {s2, u2, originB, orgBID}, {s3, u1, originB, orgBID}} {
		got, ok := base[c.id]
		if !ok || got.UserID != c.user || got.Origin != c.origin || got.OrganizationID != c.org || got.Kind != "browser" || got.Revoked ||
			got.ProviderLogoutReturned || got.Assurance != "unknown" || got.AuthTime.IsZero() || got.RenewedAt.IsZero() {
			t.Fatalf("baseline row %s: %+v (present %v)", c.id, got, ok)
		}
	}

	t.Run("fresh_only", func(t *testing.T) {
		fx.renew(s1, now)
		fx.revoke(s2, now)
		s4 := fx.session(u1, originA, orgAID, now)
		d, err := mqNewSessions(ctx, pool, reader, base)
		if err != nil {
			t.Fatal(err)
		}
		if len(d.New) != 1 || d.New[0].ID != s4 || d.Renewed != 1 {
			t.Fatalf("new %s renewed %d, want exactly %s and 1", mqIDs(d.New), d.Renewed, s4)
		}
		found, _, ok, err := mqWaitNewSession(ctx, pool, reader, base, func(s mqSessionState) bool { return s.UserID == u1 }, 5*time.Second, 50*time.Millisecond)
		if err != nil || !ok || found.ID != s4 {
			t.Fatalf("wait: %+v %v %v", found, ok, err)
		}
		// The renewed and the revoked baseline rows are never offered as new, whatever the predicate.
		if _, _, ok, err := mqWaitNewSession(ctx, pool, reader, base, func(s mqSessionState) bool { return s.ID == s1 || s.ID == s2 }, 200*time.Millisecond, 50*time.Millisecond); err != nil || ok {
			t.Fatalf("a baseline row was offered as new (%v %v)", ok, err)
		}
	})

	t.Run("after_revocation", func(t *testing.T) {
		// A session inserted before the revocation of the prior id is in the re-baseline taken after that revocation;
		// only a session inserted after the re-baseline is new (Q-6's ordering, without any timestamp).
		before := fx.session(u2, originB, orgBID, now.Add(time.Second))
		revoked, returned, err := mqRevocation(ctx, pool, reader, s3)
		if err != nil || revoked || returned {
			t.Fatalf("prior session before the revocation: %v %v %v", revoked, returned, err)
		}
		fx.revoke(s3, now.Add(2*time.Second))
		returned, ok, err := mqWaitRevoked(ctx, pool, reader, s3, 5*time.Second, 50*time.Millisecond)
		if err != nil || !ok || returned {
			t.Fatalf("revocation wait: ok %v returned %v %v", ok, returned, err)
		}
		re, err := mqBaseline(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := re[before]; !ok {
			t.Fatal("the pre-revocation session is missing from the re-baseline")
		}
		after := fx.session(u2, originB, orgBID, now.Add(3*time.Second))
		d, err := mqNewSessions(ctx, pool, reader, re)
		if err != nil {
			t.Fatal(err)
		}
		if len(d.New) != 1 || d.New[0].ID != after {
			t.Fatalf("after the re-baseline: %s, want exactly %s", mqIDs(d.New), after)
		}
		latest, err := mqLatestAuth(ctx, pool, reader, u2)
		if err != nil || !latest.Equal(now.Add(3*time.Second)) {
			t.Fatalf("latest auth_time %v %v", latest, err)
		}
	})

	t.Run("org_discovery", func(t *testing.T) {
		orgs0, err := mqOrganizations(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		o0 := mqOrgIDs(orgs0)
		label := fmt.Sprintf("mq-c-%d", time.Now().UnixNano()%1000000)
		orgC := fx.org(label)
		orgs, err := mqOrganizations(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		discovered := ""
		for _, o := range orgs {
			if !o0[o.ID] && o.PlatformLabel == label {
				discovered = o.ID
			}
		}
		if discovered != orgC {
			t.Fatalf("discovered %q, want the new Organization %s", discovered, orgC)
		}
		// The Q-4 match: the new session on C's label origin binds C by the same mechanism as A and B.
		s0, err := mqBaseline(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		fx.session(u1, p.labelOrigin(label), orgC, now.Add(4*time.Second))
		d, err := mqNewSessions(ctx, pool, reader, s0)
		if err != nil {
			t.Fatal(err)
		}
		c, err := mqBindTenant("C", d.New, orgs, o0, p, nil)
		if err != nil || c.OrganizationID != orgC || c.Label != label || c.UserID != u1 {
			t.Fatalf("binding C: %+v %v", c, err)
		}
	})

	t.Run("qa_binding", func(t *testing.T) {
		// A session on a label host that exists before S0 is never offered for binding.
		preLabel := fmt.Sprintf("mq-pre-%d", time.Now().UnixNano()%1000000)
		preOrg := fx.org(preLabel)
		pre := fx.session(u1, p.labelOrigin(preLabel), preOrg, now.Add(5*time.Second))
		s0, err := mqBaseline(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		orgs0, err := mqOrganizations(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		o0 := mqOrgIDs(orgs0)
		if !o0[preOrg] {
			t.Fatal("the pre-existing Organization is missing from O0")
		}
		la := fmt.Sprintf("mq-la-%d", time.Now().UnixNano()%1000000)
		orgA := fx.org(la)
		sa := fx.session(u1, p.labelOrigin(la), orgA, now.Add(6*time.Second))
		found, d, ok, err := mqWaitNewSession(ctx, pool, reader, s0, func(s mqSessionState) bool { return s.Kind == "browser" }, 5*time.Second, 50*time.Millisecond)
		if err != nil || !ok || found.ID != sa {
			t.Fatalf("Q-A1 discovery: %+v %v %v", found, ok, err)
		}
		for _, s := range d.New {
			if s.ID == pre {
				t.Fatal("the pre-S0 session was offered for binding")
			}
		}
		orgs, err := mqOrganizations(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		a, err := mqBindTenant("A", d.New, orgs, o0, p, nil)
		if err != nil || a.OrganizationID != orgA || a.Label != la || a.SessionID != sa || a.Origin != p.labelOrigin(la) {
			t.Fatalf("binding A: %+v %v", a, err)
		}
		// Q-A2 re-takes both baselines after the A binding and binds B, which must differ from A.
		s0, err = mqBaseline(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		orgs0, err = mqOrganizations(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		o0 = mqOrgIDs(orgs0)
		lb := fmt.Sprintf("mq-lb-%d", time.Now().UnixNano()%1000000)
		orgB := fx.org(lb)
		sb := fx.session(u2, p.labelOrigin(lb), orgB, now.Add(7*time.Second))
		d, err = mqNewSessions(ctx, pool, reader, s0)
		if err != nil {
			t.Fatal(err)
		}
		orgs, err = mqOrganizations(ctx, pool, reader)
		if err != nil {
			t.Fatal(err)
		}
		b, err := mqBindTenant("B", d.New, orgs, o0, p, []mqTenant{a})
		if err != nil || b.OrganizationID != orgB || b.SessionID != sb || b.Origin == a.Origin {
			t.Fatalf("binding B: %+v %v", b, err)
		}
		// A's session offered again as B is refused (bound origin and Organization).
		if _, err := mqBindTenant("B", []mqSessionState{found}, orgs, map[string]bool{}, p, []mqTenant{a}); err == nil {
			t.Fatal("A's session bound as B")
		}
	})

	t.Run("negative_probes", func(t *testing.T) {
		// The grant boundary the design respects (D-14/D-16): the three ungranted columns are refused as 42501, which
		// also proves the legacy created_at look-back could not have run under the reader.
		for _, stmt := range []string{
			`SELECT created_at FROM app.sessions`,
			`SELECT provider_revocation FROM app.sessions`,
			`SELECT revoke_reason FROM app.sessions`,
		} {
			if code := mqProbe(t, pool, reader, stmt); code != "42501" {
				t.Errorf("%s: %q, want 42501", stmt, code)
			}
		}
		// credential_hash is granted to the reader (00006): it is excluded by the registry, not by the grant.
		if code := mqProbe(t, pool, reader, `SELECT credential_hash FROM app.sessions WHERE false`); code != "" {
			t.Errorf("credential_hash under the reader: %q (granted; excluded by TestMQSessionColumnsAreSanitized)", code)
		}
		if code := mqProbe(t, pool, reader, `UPDATE app.sessions SET renewed_at = now() WHERE false`); code != "25006" {
			t.Errorf("UPDATE in the observer's transaction: %q, want 25006 (read_only_sql_transaction)", code)
		}
		// Row-level security is in force, not bypassed: the same select as the runtime role without a context sees
		// nothing, while the reader's policy admitted every row above.
		rows, err := runtimeConn(t).Query(ctx, mqSQLSessions)
		if err != nil {
			t.Fatalf("runtime select: %v", err)
		}
		n := 0
		for rows.Next() {
			n++
		}
		rows.Close()
		if err := rows.Err(); err != nil || n != 0 {
			t.Fatalf("runtime without context: %d rows, %v (want 0)", n, err)
		}
		// mqRead refuses a non-registry statement before any SQL runs.
		err = mqRead(ctx, pool, reader, func(ctx context.Context, q mqTx) error {
			_, err := q.Query(ctx, `SELECT created_at FROM app.sessions`)
			return err
		})
		if !errors.Is(err, errMQNotOperational) {
			t.Errorf("non-registry statement through mqRead: %v", err)
		}
		// The registry's own statements run under the reader: the live proof that every member names granted columns.
		for name, stmt := range mqOperationalSQL {
			args := []any{}
			switch stmt {
			case mqSQLRevocation:
				args = append(args, s1)
			case mqSQLLatestAuth:
				args = append(args, u1)
			case mqSQLRoleFlags, mqSQLReaderPolicies:
				args = append(args, reader)
			}
			if code := mqProbe(t, pool, reader, stmt, args...); code != "" {
				t.Errorf("registry %s under the reader: %s", name, code)
			}
		}
	})
}
