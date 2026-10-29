//go:build integration

// G016 B-DOMAINS-IT (Plan r4 §1.4, §1.5, §5.5, §6.2): TestDomains DOM-1..10 against the access integration environment
// (real PostgreSQL at version 11, the attested runtime role, the real guarded facade). DOM-1 the operation x role matrix
// with zero effects (and the slug policy/guard cases of deviation #1), DOM-2 concurrent duplicate claims and the claim
// fence (deviation #3), DOM-3 concurrent slug updates and UUID stability, DOM-4 verify/remove races, late and stale
// results, the replacement claim and the explicit withdrawal retry, DOM-5 reassignment needs fresh proof, DOM-6
// freshness and TLS expiry, DOM-7 the resolver per state, DOM-8 the exact audit rows per operation, DOM-9 the
// restore/configuration mismatch, DOM-10 the bounded footprint observations E-DOM-F1..F4. Fixture time travel (a past
// deadline the guard refuses to write) uses the attested superuser with triggers disabled for one statement. G057
// (version 21): platform hosts are platform labels (DOM-7, DOM-9 resolve label hosts); DOM-11 a slug change keeps the
// label, its routing and the live sessions, and no runtime context writes the label; DOM-12 the label host resolves to
// the exact UUID under the runtime role while the slug, unknown, reserved and suffix-confusion hosts are unknown_host.
package access

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/domains"
	"core/internal/hostname"
	"core/internal/tenancy"
)

const (
	domTTL   = time.Hour
	domLabel = "_platform-challenge"
	domFresh = 24 * time.Hour
)

type domWorld struct {
	org, other                            tenancy.OrganizationID
	owner, admin, member, gowner, foreign *itUser
}

func (e *itEnv) domWorld(t *testing.T) domWorld {
	t.Helper()
	w := domWorld{owner: e.newUser(t), admin: e.newUser(t), member: e.newUser(t), gowner: e.newUser(t), foreign: e.newUser(t)}
	w.org = e.org(t, "Domains "+e.canaries[0])
	w.other = e.org(t, "Domains Other")
	e.grantOrg(t, w.org, w.owner, tenancy.OrgOwner)
	e.grantOrg(t, w.org, w.admin, tenancy.OrgAdmin)
	e.grantOrg(t, w.org, w.member, tenancy.OrgMember)
	e.grantOrg(t, w.org, w.gowner, tenancy.OrgMember)
	g := e.group(t, w.org, w.owner, "Domains Group", fmt.Sprintf("dom-g-%d", itSeq.Add(1)))
	e.grantGroup(t, w.org, g.ID, w.gowner, tenancy.GroupOwner)
	e.grantOrg(t, w.other, w.foreign, tenancy.OrgOwner)
	// G016 A2: one prepared credential per (user, Organization); foreign at org is membership-negative (admitted only
	// for the issuance, see bind).
	for _, u := range []*itUser{w.owner, w.admin, w.member, w.gowner, w.foreign} {
		e.bind(t, u, w.org)
	}
	e.bind(t, w.foreign, w.other)
	return w
}

// as resolves u's Organization context in org (no Group) in a request scope of u's credential for org.
func (e *itEnv) as(t *testing.T, u *itUser, org tenancy.OrganizationID) (context.Context, *TenantContext) {
	t.Helper()
	rctx, _ := beginAs(t, e.svc, context.Background(), u, org)
	return rctx, resolve(t, e.svc, rctx, org, tenancy.GroupID{})
}

// domProofDNS is fixture DNS for the proof function: the listed TXT and CNAME answers, NXDOMAIN otherwise (no live
// lookup).
type domProofDNS struct {
	txt   map[string][]string
	cname map[string]string
}

func (d domProofDNS) LookupTXT(_ context.Context, name string) ([]string, error) {
	if v, ok := d.txt[name]; ok {
		return v, nil
	}
	return nil, domains.ErrNXDomain
}

func (d domProofDNS) LookupCNAME(_ context.Context, name string) (string, error) {
	if v, ok := d.cname[name]; ok {
		return v, nil
	}
	return "", domains.ErrNXDomain
}

// domRow is the full text of one domain row (attested superuser read).
func (e *itEnv) domRow(t *testing.T, id string) string {
	t.Helper()
	var s string
	if err := e.setup.QueryRow(context.Background(), `SELECT d::text FROM app.organization_domains d WHERE id = $1::text::uuid`, id).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func domHost() string { return fmt.Sprintf("portal-%d.acme-domains.example", itSeq.Add(1)) }

func (e *itEnv) claim(t *testing.T, ctx context.Context, tc *TenantContext, host string) DomainClaim {
	t.Helper()
	c, err := e.svc.ClaimDomain(ctx, tc, host, domTTL, domLabel)
	if err != nil {
		t.Fatalf("claim %s: %v", host, err)
	}
	return c
}

// check runs one verify or recheck through the facade with observation obs.
func (e *itEnv) check(t *testing.T, ctx context.Context, tc *TenantContext, id string, c domains.Check, obs domains.Observation) (DomainView, error) {
	t.Helper()
	r, err := e.svc.BeginDomainCheck(ctx, tc, id, c)
	if err != nil {
		return DomainView{}, err
	}
	if obs.TLS.State != "" && obs.TLS.Ref == (domains.DomainRef{}) {
		obs.TLS.Ref = r.Ref() // the status of the row's own DomainRef, echoed as every adapter does (domains/api check)
	}
	return e.svc.RecordDomainCheck(ctx, tc, r, c, obs, domFresh)
}

func okProof() domains.Observation { return domains.Observation{Proof: domains.OutcomeOK} }

func readyTLS(notAfter time.Time) domains.Observation {
	return domains.Observation{Proof: domains.OutcomeOK, TLS: domains.TLSResult{State: domains.TLSReady, NotAfter: notAfter}}
}

// activeDomain claims host in tc's Organization and drives it to active (proof fresh, TLS ready for 30 days).
func (e *itEnv) activeDomain(t *testing.T, ctx context.Context, tc *TenantContext, host string) DomainClaim {
	t.Helper()
	c := e.claim(t, ctx, tc, host)
	if _, err := e.check(t, ctx, tc, c.ID, domains.CheckVerify, okProof()); err != nil {
		t.Fatalf("verify %s: %v", host, err)
	}
	if v, err := e.check(t, ctx, tc, c.ID, domains.CheckRecheck, readyTLS(time.Now().Add(30*24*time.Hour))); err != nil || v.State != domains.StateActive {
		t.Fatalf("activate %s: %v %v", host, v.State, err)
	}
	return c
}

// hostState reads the custom-host predicates through the runtime role (the auth-reader function).
func (e *itEnv) hostState(t *testing.T, host string) string {
	t.Helper()
	var org *string
	var routing, issuance bool
	err := e.pool.QueryRow(context.Background(), `SELECT organization_id::text, routing, issuance FROM app.hostname_state('custom', $1)`, host).
		Scan(&org, &routing, &issuance)
	if errors.Is(err, pgx.ErrNoRows) {
		return "none"
	}
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%s routing=%v issuance=%v", *org, routing, issuance)
}

// setPlatformLabel gives org the platform label l (G057: the label, not the slug, is the platform host). A fixed
// synthetic label is legal in tests (production labels are generated by core). No runtime role may write the column, so
// the fixture uses the attested superuser.
func (e *itEnv) setPlatformLabel(t *testing.T, org tenancy.OrganizationID, l string) {
	t.Helper()
	if _, err := e.setup.Exec(context.Background(), `UPDATE app.organizations SET platform_label = $1 WHERE id = $2::text::uuid`, l, org.String()); err != nil {
		t.Fatalf("platform label %s: %v", l, err)
	}
}

func want(org tenancy.OrganizationID, routing, issuance bool) string {
	return fmt.Sprintf("%s routing=%v issuance=%v", org, routing, issuance)
}

// timeTravel applies a fixture-only assignment to one domain row as the attested superuser with triggers disabled for
// this statement (the guard refuses a past deadline by design; this simulates elapsed time, never a product path).
func (e *itEnv) timeTravel(t *testing.T, id, set string) {
	t.Helper()
	ctx := context.Background()
	tx, err := e.setup.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE app.organization_domains SET `+set+`, version = version + 1 WHERE id = $1::text::uuid`, id); err != nil {
		t.Fatalf("time travel: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

func (e *itEnv) domHash(t *testing.T, orgs ...tenancy.OrganizationID) string {
	t.Helper()
	var b strings.Builder
	for _, o := range orgs {
		b.WriteString(e.part(t, "organization_domains", o) + "/" + e.part(t, "organizations", o) + ";")
	}
	return b.String()
}

// domEvents lists the audit rows of one domain resource as sorted action/outcome/reason triples.
func (e *itEnv) domEvents(t *testing.T, id string) []string {
	t.Helper()
	rows, err := e.setup.Query(context.Background(), `SELECT action || '/' || outcome || '/' || coalesce(reason, '') FROM audit.events
WHERE resource_type = 'organization_domain' AND resource_id = $1::text::uuid`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}

func TestDomains(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	w := e.domWorld(t)
	octx, otc := e.as(t, w.owner, w.org)

	t.Run("DOM-1 operation x role matrix with zero effects", func(t *testing.T) {
		live := e.claim(t, octx, otc, domHost())
		removed := e.claim(t, octx, otc, domHost())
		rm, err := e.svc.RemoveDomain(octx, otc, removed.ID)
		if err != nil {
			t.Fatal(err)
		}
		ops := func(ctx context.Context, tc *TenantContext) map[string]func() error {
			return map[string]func() error{
				"list":  func() error { _, err := e.svc.ListDomains(ctx, tc); return err },
				"claim": func() error { _, err := e.svc.ClaimDomain(ctx, tc, domHost(), domTTL, domLabel); return err },
				"rotate": func() error {
					_, err := e.svc.RotateChallenge(ctx, tc, live.ID, domTTL, domLabel)
					return err
				},
				"verify": func() error { _, err := e.svc.BeginDomainCheck(ctx, tc, live.ID, domains.CheckVerify); return err },
				"remove": func() error { _, err := e.svc.RemoveDomain(ctx, tc, live.ID); return err },
				"retry":  func() error { _, err := e.svc.RetryWithdrawal(ctx, tc, rm.ID, 3); return err },
				"slug": func() error {
					_, err := e.svc.SetOrganizationSlug(ctx, tc, fmt.Sprintf("dom-%d", itSeq.Add(1)))
					return err
				},
				"cleanup": func() error {
					_, err := e.svc.RecordWithdrawal(ctx, tc, rm, domains.TLSResult{Ref: rm.Ref(), State: domains.TLSWithdrawn}, false)
					return err
				},
			}
		}
		before := e.domHash(t, w.org, w.other)
		for _, d := range []struct {
			name string
			u    *itUser
		}{{"member", w.member}, {"group owner", w.gowner}} {
			dctx, dtc := e.as(t, d.u, w.org)
			for op, f := range ops(dctx, dtc) {
				wantCode(t, d.name+" "+op, f(), CodeForbidden)
			}
		}
		// Cross-Organization owner: in its own context the target row is not found; nothing changes.
		fctx, ftc := e.as(t, w.foreign, w.other)
		wantCode(t, "foreign remove", func() error { _, err := e.svc.RemoveDomain(fctx, ftc, live.ID); return err }(), CodeNotFound)
		wantCode(t, "foreign verify", func() error { _, err := e.svc.BeginDomainCheck(fctx, ftc, live.ID, domains.CheckVerify); return err }(), CodeNotFound)
		if l, err := e.svc.ListDomains(fctx, ftc); err != nil || len(l) != 0 {
			t.Errorf("foreign list: %d rows (%v)", len(l), err)
		}
		// A non-member (and a global user with no membership at all) cannot resolve the Organization context.
		nobody := e.newUser(t)
		e.bind(t, nobody, w.org) // admitted only for the issuance: membership-negative again here
		for _, u := range []*itUser{w.foreign, nobody} {
			rctx, _ := beginAs(t, e.svc, ctx, u, w.org)
			if _, err := e.svc.Resolve(rctx, ResolveRequest{OrganizationID: w.org.String()}); err == nil {
				t.Error("a non-member resolved the Organization context")
			}
		}
		if got := e.domHash(t, w.org, w.other); got != before {
			t.Fatal("a denied operation changed domain or Organization rows")
		}
		// Owner and admin are allowed every operation.
		for _, u := range []*itUser{w.owner, w.admin} {
			uctx, utc := e.as(t, u, w.org)
			if _, err := e.svc.ListDomains(uctx, utc); err != nil {
				t.Errorf("list: %v", err)
			}
			c := e.claim(t, uctx, utc, domHost())
			if _, err := e.svc.RotateChallenge(uctx, utc, c.ID, domTTL, domLabel); err != nil {
				t.Errorf("rotate: %v", err)
			}
			if _, err := e.svc.RemoveDomain(uctx, utc, c.ID); err != nil {
				t.Errorf("remove: %v", err)
			}
		}
		// Deviation #1: owner and admin set the slug; an admin's combined slug + policy-column UPDATE in its own runtime
		// context is refused 42501 by the slug guard with zero effect; an admin slug-only UPDATE succeeds; the owner's
		// combined UPDATE stays allowed (organizations_update unchanged).
		for _, u := range []*itUser{w.owner, w.admin} {
			uctx, utc := e.as(t, u, w.org)
			s := fmt.Sprintf("dom-slug-%d", itSeq.Add(1))
			if got, err := e.svc.SetOrganizationSlug(uctx, utc, s); err != nil || got != s {
				t.Errorf("set slug: %q %v", got, err)
			}
		}
		direct := func(u *itUser, set string) error {
			tx, err := e.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`, u.id.String(), w.org.String()); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `UPDATE app.organizations SET `+set+` WHERE id = $1::text::uuid`, w.org.String()); err != nil {
				return err
			}
			return tx.Commit(ctx)
		}
		before = e.part(t, "organizations", w.org)
		err = direct(w.admin, fmt.Sprintf("slug = 'dom-combo-%d', group_admins_edit_business_profiles = NOT group_admins_edit_business_profiles", itSeq.Add(1)))
		var pg *pgconn.PgError
		if !errors.As(err, &pg) || pg.Code != "42501" {
			t.Errorf("admin combined slug+policy UPDATE: %v, want 42501", err)
		}
		if e.part(t, "organizations", w.org) != before {
			t.Error("the refused admin UPDATE changed the Organization row")
		}
		if err := direct(w.admin, fmt.Sprintf("slug = 'dom-only-%d'", itSeq.Add(1))); err != nil {
			t.Errorf("admin slug-only UPDATE: %v", err)
		}
		if err := direct(w.owner, fmt.Sprintf("slug = 'dom-own-%d', group_admins_edit_business_profiles = NOT group_admins_edit_business_profiles", itSeq.Add(1))); err != nil {
			t.Errorf("owner combined UPDATE: %v", err)
		}

		// Direct runtime-role SQL against the real policies (Codex domain/RLS evidence correction, 2026-10-16): every
		// statement runs in its own rolled-back transaction on the attested runtime pool under a fixed actor identity;
		// the paired owner/admin controls must succeed so a broken fixture cannot make a denial pass, and every denial
		// leaves the setup-side domain and Organization rows of both Organizations identical.
		if err := e.rlsAttest(e.pool); err != nil { // the executed runtime-role attestation of the raw-SQL pool
			t.Fatalf("runtime attestation: %v", err)
		}
		other := e.claim(t, fctx, ftc, domHost()) // a domain row of the other Organization (its owner's own claim)
		rt := func(u *itUser, ctxOrg string, stmt func(pgx.Tx) error) error {
			tx, err := e.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`, u.id.String(), ctxOrg); err != nil {
				t.Fatal(err)
			}
			return stmt(tx)
		}
		idsOf := func(row pgx.Row) string {
			var s *string
			if err := row.Scan(&s); err != nil {
				t.Fatal(err)
			}
			if s == nil {
				return ""
			}
			return *s
		}
		const idsQ = `SELECT string_agg(id::text, ',' ORDER BY id) FROM app.organization_domains`
		challenge := func(c string) string { return strings.Repeat(c, 52) }
		insertQ := `INSERT INTO app.organization_domains (id, organization_id, hostname, generation, challenge_value, challenge_expires_at,
  created_by) VALUES (gen_random_uuid(), $1::text::uuid, $2, 1, $3, now() + interval '1 hour', $4::text::uuid)`
		updateQ := `UPDATE app.organization_domains SET challenge_value = $1, challenge_expires_at = now() + interval '2 hours',
  version = version + 1 WHERE id = $2::text::uuid`
		before = e.domHash(t, w.org, w.other)
		// Paired authorized controls: owner and admin see exactly their Organization's rows, may insert a pending claim and
		// may rotate a pending row's challenge (allowed columns, version + 1).
		wantOrg := idsOf(e.setup.QueryRow(ctx, idsQ+` WHERE organization_id = $1::text::uuid`, w.org.String()))
		if wantOrg == "" || !strings.Contains(wantOrg, live.ID) {
			t.Fatalf("control fixture: the Organization has no domain rows (%q)", wantOrg)
		}
		for _, u := range []*itUser{w.owner, w.admin} {
			if err := rt(u, w.org.String(), func(tx pgx.Tx) error {
				if got := idsOf(tx.QueryRow(ctx, idsQ)); got != wantOrg {
					t.Errorf("control select: visible %q, want exactly %q", got, wantOrg)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if err := rt(u, w.org.String(), func(tx pgx.Tx) error {
				tag, err := tx.Exec(ctx, insertQ, w.org.String(), domHost(), challenge("a"), u.id.String())
				if err == nil && tag.RowsAffected() != 1 {
					t.Errorf("control insert: %d rows", tag.RowsAffected())
				}
				return err
			}); err != nil {
				t.Errorf("control insert: %v", err)
			}
			if err := rt(u, w.org.String(), func(tx pgx.Tx) error {
				tag, err := tx.Exec(ctx, updateQ, challenge("b"), live.ID)
				if err == nil && tag.RowsAffected() != 1 {
					t.Errorf("control update: %d rows", tag.RowsAffected())
				}
				return err
			}); err != nil {
				t.Errorf("control update: %v", err)
			}
		}
		// Denials: SELECT sees none of the target's rows, INSERT is refused by the policy (42501 row-level security) and
		// UPDATE affects zero rows, for each actor/context/target.
		for _, d := range []struct {
			name      string
			u         *itUser
			ctxOrg    string
			target    tenancy.OrganizationID
			targetRow string
		}{
			{"member", w.member, w.org.String(), w.org, live.ID},
			{"group owner", w.gowner, w.org.String(), w.org, live.ID},
			{"foreign owner in its own Organization", w.foreign, w.other.String(), w.org, live.ID},
			{"foreign owner with a substituted Organization context", w.foreign, w.org.String(), w.org, live.ID},
			{"admin with a substituted Organization context", w.admin, w.other.String(), w.other, other.ID},
			{"admin in its own context with a substituted target Organization and row", w.admin, w.org.String(), w.other, other.ID},
			{"admin without an Organization context", w.admin, "", w.org, live.ID},
		} {
			if err := rt(d.u, d.ctxOrg, func(tx pgx.Tx) error {
				if got := idsOf(tx.QueryRow(ctx, idsQ+` WHERE organization_id = $1::text::uuid`, d.target.String())); got != "" {
					t.Errorf("%s: select saw %q", d.name, got)
				}
				return nil
			}); err != nil {
				t.Fatalf("%s select: %v", d.name, err)
			}
			err := rt(d.u, d.ctxOrg, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, insertQ, d.target.String(), domHost(), challenge("c"), d.u.id.String())
				return err
			})
			var pg *pgconn.PgError
			if !errors.As(err, &pg) || pg.Code != "42501" || !strings.Contains(pg.Message, "row-level security policy") {
				t.Errorf("%s insert: %v, want the 42501 row-level security refusal", d.name, err)
			}
			if err := rt(d.u, d.ctxOrg, func(tx pgx.Tx) error {
				tag, err := tx.Exec(ctx, updateQ, challenge("d"), d.targetRow)
				if err == nil && tag.RowsAffected() != 0 {
					t.Errorf("%s update: %d rows, want 0", d.name, tag.RowsAffected())
				}
				return err
			}); err != nil {
				t.Errorf("%s update: %v, want zero rows", d.name, err)
			}
		}
		// Slug denials: the slug UPDATE by an ordinary member or a Group owner of the Organization is refused 42501 by the
		// unchanged G006 organizations_update WITH CHECK (a member's row passes its USING, so the refusal is an error,
		// not zero rows); a foreign actor (own or substituted context), an admin targeting another Organization (own or
		// substituted context) and an admin without an Organization context see no row and affect zero rows. The admin
		// slug-only success and combined-column refusal above are the controls.
		for _, d := range []struct {
			name    string
			u       *itUser
			ctxOrg  string
			target  tenancy.OrganizationID
			refused bool
		}{
			{"member", w.member, w.org.String(), w.org, true},
			{"group owner", w.gowner, w.org.String(), w.org, true},
			{"foreign owner in its own Organization", w.foreign, w.other.String(), w.org, false},
			{"foreign owner with a substituted Organization context", w.foreign, w.org.String(), w.org, false},
			{"admin targeting another Organization", w.admin, w.org.String(), w.other, false},
			{"admin with a substituted Organization context", w.admin, w.other.String(), w.other, false},
			{"admin without an Organization context", w.admin, "", w.org, false},
		} {
			var rows int64
			err := rt(d.u, d.ctxOrg, func(tx pgx.Tx) error {
				tag, err := tx.Exec(ctx, `UPDATE app.organizations SET slug = $1 WHERE id = $2::text::uuid`, fmt.Sprintf("dom-deny-%d", itSeq.Add(1)), d.target.String())
				rows = tag.RowsAffected()
				return err
			})
			var pg *pgconn.PgError
			switch {
			case d.refused && (!errors.As(err, &pg) || pg.Code != "42501" || !strings.Contains(pg.Message, "row-level security policy")):
				t.Errorf("slug %s: %v, want the 42501 row-level security refusal", d.name, err)
			case !d.refused && (err != nil || rows != 0):
				t.Errorf("slug %s: %d rows %v, want zero rows", d.name, rows, err)
			}
		}
		if got := e.domHash(t, w.org, w.other); got != before {
			t.Fatal("a direct runtime statement changed domain or Organization rows")
		}
		e.setFlag(t, w.org, false)
	})

	t.Run("DOM-2 concurrent duplicate claims and the claim fence", func(t *testing.T) {
		// G016 A2: exclusivity moved from claim time to proof time. A provable hostname (not a refused special-use
		// name), so each claimant's observation comes from the real proof function over fixture DNS answers.
		host := fmt.Sprintf("race-%d.acme-proof.net", itSeq.Add(1))
		type scope struct {
			ctx context.Context
			tc  *TenantContext
		}
		scopes := make([]scope, 8)
		for i := range scopes {
			u, org := w.owner, w.org
			if i%2 == 1 {
				u, org = w.foreign, w.other
			}
			scopes[i].ctx, scopes[i].tc = e.as(t, u, org)
		}
		var wg sync.WaitGroup
		start := make(chan struct{})
		errs := make([]error, 8)
		claims := make([]DomainClaim, 8)
		for i := range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				claims[i], errs[i] = e.svc.ClaimDomain(scopes[i].ctx, scopes[i].tc, host, domTTL, domLabel)
			}()
		}
		close(start)
		wg.Wait()
		for i, err := range errs {
			if err != nil {
				t.Fatalf("claim %d: %v", i, err)
			}
		}
		var live int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM app.organization_domains WHERE hostname = $1 AND state = 'pending'`, host).Scan(&live); err != nil {
			t.Fatal(err)
		}
		if live != 8 {
			t.Fatalf("concurrent claims: %d pending rows, want 8 (no claim-time exclusivity)", live)
		}
		// Each claimant proves its own challenge: valid TXT (its value) and CNAME answers, observed individually before the
		// race; another claimant's value does not prove it.
		pcfg := domains.ProofConfig{ChallengeLabel: domLabel, CNAMETarget: "edge.proof-fixture.net", BaseDomain: "proof-fixture.net", Timeout: time.Second}
		reads := make([]domains.Record, 8)
		obs := make([]domains.Observation, 8)
		rows := make([]string, 8)
		for i := range 8 {
			var err error
			if reads[i], err = e.svc.BeginDomainCheck(scopes[i].ctx, scopes[i].tc, claims[i].ID, domains.CheckVerify); err != nil {
				t.Fatalf("begin check %d: %v", i, err)
			}
			dns := domProofDNS{txt: map[string][]string{domLabel + "." + host: {reads[i].ChallengeValue}}, cname: map[string]string{host: pcfg.CNAMETarget}}
			if out := domains.Proof(ctx, dns, pcfg, host, reads[i].ChallengeValue); out != domains.OutcomeOK {
				t.Fatalf("claimant %d proof: %s", i, out)
			}
			other := domProofDNS{txt: map[string][]string{domLabel + "." + host: {reads[(i+1)%8].ChallengeValue}}, cname: dns.cname}
			if out := domains.Proof(ctx, other, pcfg, host, reads[i].ChallengeValue); out != domains.OutcomeMismatch {
				t.Fatalf("claimant %d proved by another claimant's value: %s", i, out)
			}
			obs[i] = domains.Observation{Proof: domains.Proof(ctx, dns, pcfg, host, reads[i].ChallengeValue)}
			rows[i] = e.domRow(t, claims[i].ID)
		}
		// Concurrent promotions: the proven-state unique index is the only arbiter.
		start = make(chan struct{})
		views := make([]DomainView, 8)
		for i := range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				views[i], errs[i] = e.svc.RecordDomainCheck(scopes[i].ctx, scopes[i].tc, reads[i], domains.CheckVerify, obs[i], domFresh)
			}()
		}
		close(start)
		wg.Wait()
		won, taken := -1, 0
		for i, err := range errs {
			switch {
			case err == nil && views[i].State == domains.StateVerified && won < 0:
				won = i
			case CodeOf(err) == CodeHostTaken:
				taken++
				if e.domRow(t, claims[i].ID) != rows[i] {
					t.Errorf("loser %d: row changed", i)
				}
				for _, ev := range e.domEvents(t, claims[i].ID) {
					if strings.HasPrefix(ev, "domain.checked/") {
						t.Errorf("loser %d: %s audit row", i, ev)
					}
				}
			default:
				t.Errorf("promotion %d: %v (state %s)", i, err, views[i].State)
			}
		}
		if won < 0 || taken != 7 {
			t.Fatalf("concurrent promotions: winner %d, %d host_taken; want one winner and 7", won, taken)
		}
		if got := e.domEvents(t, claims[won].ID); !slices.Contains(got, "domain.checked/success/ok") || !slices.Contains(got, "domain.state_changed/success/verified") {
			t.Errorf("winner audit rows: %v", got)
		}
		// A claim against the proven hostname is host_taken at claim time, with zero effects.
		loser := (won + 1) % 8
		pre := e.domHash(t, w.org, w.other)
		_, err := e.svc.ClaimDomain(scopes[loser].ctx, scopes[loser].tc, host, domTTL, domLabel)
		wantCode(t, "claim of a proven hostname", err, CodeHostTaken)
		if e.domHash(t, w.org, w.other) != pre {
			t.Fatal("the refused claim changed rows")
		}
		// Deviation #3: the claim generation is an opaque CSPRNG fence; the runtime has no generation sequence.
		var seqs int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM pg_class WHERE relname = 'hostname_generation'`).Scan(&seqs); err != nil || seqs != 0 {
			t.Fatalf("app.hostname_generation exists (%d, %v)", seqs, err)
		}
		// A fence collision with the removed row of the same hostname regenerates once; a second collision fails closed.
		h2 := domHost()
		first := e.claim(t, octx, otc, h2)
		gone, err := e.svc.RemoveDomain(octx, otc, first.ID)
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		e.svc.domainHooks = &domainHooks{generation: func(attempt int, drawn int64) int64 {
			calls++
			if attempt == 0 {
				return gone.Generation
			}
			return drawn
		}}
		second, err := e.svc.ClaimDomain(octx, otc, h2, domTTL, domLabel)
		e.svc.domainHooks = nil
		if err != nil || calls != 2 || second.Generation == gone.Generation {
			t.Fatalf("one fence collision: %v, %d draws, generation reused %v", err, calls, second.Generation == gone.Generation)
		}
		if _, err := e.svc.RemoveDomain(octx, otc, second.ID); err != nil {
			t.Fatal(err)
		}
		before := e.domHash(t, w.org)
		e.svc.domainHooks = &domainHooks{generation: func(int, int64) int64 { return gone.Generation }}
		_, err = e.svc.ClaimDomain(octx, otc, h2, domTTL, domLabel)
		e.svc.domainHooks = nil
		wantCode(t, "second fence collision", err, CodeInternal)
		if e.domHash(t, w.org) != before {
			t.Fatal("the failed claim changed rows")
		}
	})

	t.Run("DOM-3 concurrent slug updates and UUID stability", func(t *testing.T) {
		fctx, ftc := e.as(t, w.foreign, w.other)
		slug := fmt.Sprintf("dom-race-%d", itSeq.Add(1))
		var wg sync.WaitGroup
		errs := make([]error, 2)
		for i, c := range []struct {
			ctx context.Context
			tc  *TenantContext
		}{{octx, otc}, {fctx, ftc}} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, errs[i] = e.svc.SetOrganizationSlug(c.ctx, c.tc, slug)
			}()
		}
		wg.Wait()
		if !((errs[0] == nil && CodeOf(errs[1]) == CodeConflict) || (errs[1] == nil && CodeOf(errs[0]) == CodeConflict)) {
			t.Fatalf("concurrent slug updates: %v / %v, want one success and one conflict", errs[0], errs[1])
		}
		next := fmt.Sprintf("dom-next-%d", itSeq.Add(1))
		if _, err := e.svc.SetOrganizationSlug(octx, otc, next); err != nil {
			t.Fatal(err)
		}
		var id string
		if err := e.setup.QueryRow(ctx, `SELECT id::text FROM app.organizations WHERE slug = $1`, next).Scan(&id); err != nil || id != w.org.String() {
			t.Fatalf("slug %s belongs to %s (%v), want the unchanged Organization id %s", next, id, err, w.org)
		}
		var n int
		if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM app.hostname_state('platform', $1)`, slug).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if errs[0] == nil && n != 0 {
			t.Error("the replaced slug still routes")
		}
		if _, err := e.svc.SetOrganizationSlug(octx, otc, next); err != nil {
			t.Errorf("setting the current slug again: %v", err)
		}
		wantCode(t, "reserved slug", func() error {
			_, err := e.svc.SetOrganizationSlug(octx, otc, hostname.ReservedLabels()[0])
			return err
		}(), CodeInvalidInput)
	})

	t.Run("DOM-4 races, late and stale results, replacement claim and withdrawal retry", func(t *testing.T) {
		host := domHost()
		c := e.claim(t, octx, otc, host)
		read, err := e.svc.BeginDomainCheck(octx, otc, c.ID, domains.CheckVerify)
		if err != nil {
			t.Fatal(err)
		}
		// Verify vs remove: the remove commits while the check is in flight; the late verify is refused.
		rm, err := e.svc.RemoveDomain(octx, otc, c.ID)
		if err != nil {
			t.Fatal(err)
		}
		before := e.domHash(t, w.org)
		_, err = e.svc.RecordDomainCheck(octx, otc, read, domains.CheckVerify, okProof(), domFresh)
		wantCode(t, "late verify on a removed row", err, CodeDomainStateInvalid)
		if e.domHash(t, w.org) != before || e.hostState(t, host) != "none" {
			t.Fatal("the late verify changed the removed row or revived the host")
		}
		wantCode(t, "recheck of a removed row", func() error { _, err := e.svc.BeginDomainCheck(octx, otc, c.ID, domains.CheckRecheck); return err }(), CodeDomainStateInvalid)
		// The replacement claim: issuance and routing stay false until its own proof.
		repl := e.claim(t, octx, otc, host)
		if repl.Generation == rm.Generation {
			t.Fatal("the replacement claim reused the removed generation")
		}
		if got := e.hostState(t, host); got != "none" { // G016 A2: the reader is proven-state only
			t.Fatalf("pending replacement: %s", got)
		}
		// A late adapter result for the removed generation carrying the replacement's reference is stale.
		replRec, err := e.svc.BeginDomainCheck(octx, otc, repl.ID, domains.CheckVerify)
		if err != nil {
			t.Fatal(err)
		}
		before = e.domHash(t, w.org)
		_, err = e.svc.RecordWithdrawal(octx, otc, rm, domains.TLSResult{Ref: replRec.Ref(), State: domains.TLSWithdrawn}, false)
		wantCode(t, "stale result", err, CodeDomainStateInvalid)
		if e.domHash(t, w.org) != before {
			t.Fatal("the stale result changed rows")
		}
		// The removed row's cleanup: only the allowed columns may change (direct runtime UPDATE of a proof column refused).
		tx, err := e.pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`, w.owner.id.String(), w.org.String()); err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, `UPDATE app.organization_domains SET verified_until = now() + interval '1 day', version = version + 1 WHERE id = $1::text::uuid`, c.ID)
		_ = tx.Rollback(ctx)
		var pg *pgconn.PgError
		if !errors.As(err, &pg) || pg.Code != "23514" || pg.ConstraintName != "organization_domains_guard" {
			t.Errorf("removed-row proof update: %v, want the guard's 23514", err)
		}
		// Withdrawal unknown, then the explicit retry withdraws the old generation's own resource; the replacement is untouched.
		if v, err := e.svc.RecordWithdrawal(octx, otc, rm, domains.TLSResult{Ref: rm.Ref(), State: domains.TLSUnknown}, false); err != nil || v.TLSState != domains.TLSUnknown {
			t.Fatalf("unknown withdrawal: %v %v", v.TLSState, err)
		}
		var replRow string
		if err := e.setup.QueryRow(ctx, `SELECT md5(r::text) FROM app.organization_domains r WHERE id = $1::text::uuid`, repl.ID).Scan(&replRow); err != nil {
			t.Fatal(err)
		}
		reserved, err := e.svc.RetryWithdrawal(octx, otc, rm.ID, 3)
		if err != nil || reserved.Generation != rm.Generation || reserved.CheckFailures != 2 {
			t.Fatalf("retry: %+v %v", reserved, err)
		}
		if v, err := e.svc.RecordWithdrawal(octx, otc, reserved, domains.TLSResult{Ref: reserved.Ref(), State: domains.TLSWithdrawn}, true); err != nil || v.TLSState != domains.TLSWithdrawn {
			t.Fatalf("retried withdrawal: %v %v", v.TLSState, err)
		}
		var replAfter string
		if err := e.setup.QueryRow(ctx, `SELECT md5(r::text) FROM app.organization_domains r WHERE id = $1::text::uuid`, repl.ID).Scan(&replAfter); err != nil || replAfter != replRow {
			t.Fatal("the retry touched the replacement claim")
		}
		wantCode(t, "retry on a withdrawn row", func() error { _, err := e.svc.RetryWithdrawal(octx, otc, rm.ID, 3); return err }(), CodeDomainStateInvalid)
		wantCode(t, "retry on a live row", func() error { _, err := e.svc.RetryWithdrawal(octx, otc, repl.ID, 3); return err }(), CodeDomainStateInvalid)
		// The retry bound: a removed row with an error outcome and one counted failure, max 1.
		x := e.claim(t, octx, otc, domHost())
		xr, err := e.svc.RemoveDomain(octx, otc, x.ID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.svc.RecordWithdrawal(octx, otc, xr, domains.TLSResult{Ref: xr.Ref(), State: domains.TLSError}, false); err != nil {
			t.Fatal(err)
		}
		before = e.domHash(t, w.org)
		wantCode(t, "retry bound", func() error { _, err := e.svc.RetryWithdrawal(octx, otc, x.ID, 1); return err }(), CodeDomainRetryExhausted)
		if e.domHash(t, w.org) != before {
			t.Fatal("the exhausted retry changed rows")
		}
		if e.hostState(t, x.Hostname) != "none" || e.hostState(t, host) != "none" { // G016 A2: the pending replacement has no reader row
			t.Fatal("a removed row routes or issues")
		}
		// G016 correction: a recheck's TLS result applies only for the locked row's exact DomainRef. A result for another
		// row, an older generation, another hostname or reference, or with no ref is refused: zero effects and no
		// successful check event; the exact ref then applies (positive control).
		th := domHost()
		td := e.activeDomain(t, octx, otc, th)
		e.timeTravel(t, td.ID, `tls_not_after = now() - interval '1 second'`) // TLS expired: a ready result would restore routing
		other := e.activeDomain(t, octx, otc, domHost())
		otherRead, err := e.svc.BeginDomainCheck(octx, otc, other.ID, domains.CheckRecheck)
		if err != nil {
			t.Fatal(err)
		}
		tread, err := e.svc.BeginDomainCheck(octx, otc, td.ID, domains.CheckRecheck)
		if err != nil {
			t.Fatal(err)
		}
		staleGen, otherHost, otherRes := tread.Ref(), tread.Ref(), tread.Ref()
		staleGen.Generation--
		otherHost.Hostname = "elsewhere-" + th
		otherRes.Reference = "foreign-resource"
		successes := func() int {
			n := 0
			for _, ev := range e.domEvents(t, td.ID) {
				if strings.HasPrefix(ev, "domain.checked/success/") {
					n++
				}
			}
			return n
		}
		okBefore, hostBefore := successes(), e.hostState(t, th)
		for _, k := range []struct {
			name string
			ref  domains.DomainRef
		}{{"foreign row", otherRead.Ref()}, {"stale generation", staleGen}, {"foreign hostname", otherHost},
			{"foreign reference", otherRes}, {"empty ref", domains.DomainRef{}}} {
			obs := readyTLS(time.Now().Add(30 * 24 * time.Hour))
			obs.TLS.Ref = k.ref
			before := e.domHash(t, w.org)
			_, err := e.svc.RecordDomainCheck(octx, otc, tread, domains.CheckRecheck, obs, domFresh)
			wantCode(t, "TLS result for a "+k.name, err, CodeDomainStateInvalid)
			if e.domHash(t, w.org) != before || e.hostState(t, th) != hostBefore {
				t.Fatalf("the %s TLS result changed rows or readiness", k.name)
			}
		}
		if successes() != okBefore {
			t.Fatal("a refused TLS result recorded a successful check")
		}
		ok := readyTLS(time.Now().Add(30 * 24 * time.Hour))
		ok.TLS.Ref = tread.Ref()
		if v, err := e.svc.RecordDomainCheck(octx, otc, tread, domains.CheckRecheck, ok, domFresh); err != nil || v.State != domains.StateActive ||
			e.hostState(t, th) != want(w.org, true, true) || successes() != okBefore+1 {
			t.Fatalf("the exact ref: %+v %v", v, err)
		}
	})

	t.Run("DOM-5 reassignment needs fresh proof", func(t *testing.T) {
		host := domHost()
		a := e.activeDomain(t, octx, otc, host)
		if got := e.hostState(t, host); got != want(w.org, true, true) {
			t.Fatalf("active: %s", got)
		}
		if _, err := e.svc.RemoveDomain(octx, otc, a.ID); err != nil {
			t.Fatal(err)
		}
		fctx, ftc := e.as(t, w.foreign, w.other)
		b := e.claim(t, fctx, ftc, host)
		if b.ChallengeValue == a.ChallengeValue || b.Generation == a.Generation {
			t.Fatal("the reassigned claim reused the previous challenge or generation")
		}
		if got := e.hostState(t, host); got != "none" { // G016 A2: the reader is proven-state only
			t.Fatalf("reassigned, pending: %s", got)
		}
		if _, err := e.check(t, fctx, ftc, b.ID, domains.CheckVerify, okProof()); err != nil {
			t.Fatal(err)
		}
		if got := e.hostState(t, host); got != want(w.other, false, true) {
			t.Fatalf("reassigned, verified: %s", got)
		}
		if _, err := e.check(t, fctx, ftc, b.ID, domains.CheckRecheck, readyTLS(time.Now().Add(24*time.Hour))); err != nil {
			t.Fatal(err)
		}
		if got := e.hostState(t, host); got != want(w.other, true, true) {
			t.Fatalf("reassigned, active: %s", got)
		}
	})

	t.Run("DOM-6 freshness and TLS expiry fail closed", func(t *testing.T) {
		host := domHost()
		d := e.activeDomain(t, octx, otc, host)
		// Stored TLS expiry: routing false, issuance true while the proof is fresh.
		e.timeTravel(t, d.ID, `tls_not_after = now() - interval '1 second'`)
		if got := e.hostState(t, host); got != want(w.org, false, true) {
			t.Fatalf("TLS expired: %s", got)
		}
		e.timeTravel(t, d.ID, `tls_not_after = now() + interval '30 days'`)
		// Proof freshness expiry: routing and issuance false.
		e.timeTravel(t, d.ID, `verified_until = now() - interval '1 second'`)
		if got := e.hostState(t, host); got != want(w.org, false, false) {
			t.Fatalf("proof expired: %s", got)
		}
		// A successful recheck restores the deadline forward.
		if _, err := e.check(t, octx, otc, d.ID, domains.CheckRecheck, readyTLS(time.Now().Add(30*24*time.Hour))); err != nil {
			t.Fatal(err)
		}
		if got := e.hostState(t, host); got != want(w.org, true, true) {
			t.Fatalf("rechecked: %s", got)
		}
		// TLS-only suspension keeps issuance.
		v, err := e.check(t, octx, otc, d.ID, domains.CheckRecheck, domains.Observation{Proof: domains.OutcomeOK, TLS: domains.TLSResult{State: domains.TLSError}})
		if err != nil || v.State != domains.StateSuspended {
			t.Fatalf("TLS-only failure: %v %v", v.State, err)
		}
		if got := e.hostState(t, host); got != want(w.org, false, true) {
			t.Fatalf("TLS-suspended: %s", got)
		}
		// A TLS-suspended row whose recheck confirms proof loss: stays suspended, verified_until = now(), issuance false.
		prev := *v.VerifiedUntil
		v, err = e.check(t, octx, otc, d.ID, domains.CheckRecheck, domains.Observation{Proof: domains.OutcomeMismatch})
		if err != nil || v.State != domains.StateSuspended || v.VerifiedUntil == nil || !v.VerifiedUntil.Before(prev) ||
			time.Until(*v.VerifiedUntil) > time.Second {
			t.Fatalf("proof loss on a suspended row: %+v %v", v, err)
		}
		if got := e.hostState(t, host); got != want(w.org, false, false) {
			t.Fatalf("proof lost: %s", got)
		}
		// Confirmed proof loss on an active row stops issuance at that commit.
		h2 := domHost()
		d2 := e.activeDomain(t, octx, otc, h2)
		if _, err := e.check(t, octx, otc, d2.ID, domains.CheckRecheck, domains.Observation{Proof: domains.OutcomeNXDomain}); err != nil {
			t.Fatal(err)
		}
		if got := e.hostState(t, h2); got != want(w.org, false, false) {
			t.Fatalf("active proof lost: %s", got)
		}
		// A transient outcome changes neither state nor deadline.
		h3 := domHost()
		d3 := e.activeDomain(t, octx, otc, h3)
		v3, err := e.check(t, octx, otc, d3.ID, domains.CheckRecheck, domains.Observation{Proof: domains.OutcomeTransient})
		if err != nil || v3.State != domains.StateActive || e.hostState(t, h3) != want(w.org, true, true) {
			t.Fatalf("transient: %v %v", v3.State, err)
		}
	})

	t.Run("DOM-7 resolver per state through the auth reader", func(t *testing.T) {
		cfg := hostname.Config{BaseDomain: "localhost", CNAMETarget: "edge.localhost"}
		r := domains.NewResolver(e.pool, cfg)
		pend := e.claim(t, octx, otc, domHost())
		ver := e.claim(t, octx, otc, domHost())
		if _, err := e.check(t, octx, otc, ver.ID, domains.CheckVerify, okProof()); err != nil {
			t.Fatal(err)
		}
		act := e.activeDomain(t, octx, otc, domHost())
		sus := e.activeDomain(t, octx, otc, domHost())
		if _, err := e.check(t, octx, otc, sus.ID, domains.CheckRecheck, domains.Observation{Proof: domains.OutcomeOK, TLS: domains.TLSResult{State: domains.TLSExpired}}); err != nil {
			t.Fatal(err)
		}
		rem := e.activeDomain(t, octx, otc, domHost())
		if _, err := e.svc.RemoveDomain(octx, otc, rem.ID); err != nil {
			t.Fatal(err)
		}
		slug := fmt.Sprintf("dom-res-%d", itSeq.Add(1))
		if _, err := e.svc.SetOrganizationSlug(octx, otc, slug); err != nil {
			t.Fatal(err)
		}
		// G057: the platform label, not the slug, is the platform host.
		label := fmt.Sprintf("dom-lbl-%d", itSeq.Add(1))
		e.setPlatformLabel(t, w.org, label)
		// G016 A2: the reader is proven-state only. A pending + proven hostname: the other Organization's pending claim
		// first, then this Organization's verified one; only the proven row answers.
		mixed := domHost()
		fctx, ftc := e.as(t, w.foreign, w.other)
		e.claim(t, fctx, ftc, mixed)
		mix := e.claim(t, octx, otc, mixed)
		if _, err := e.check(t, octx, otc, mix.ID, domains.CheckVerify, okProof()); err != nil {
			t.Fatal(err)
		}
		if got := e.hostState(t, pend.Hostname); got != "none" {
			t.Errorf("pending-only hostname through the auth reader: %s, want no row", got)
		}
		if got := e.hostState(t, mixed); got != want(w.org, false, true) {
			t.Errorf("pending + proven hostname through the auth reader: %s", got)
		}
		for _, c := range []struct {
			host, want string
		}{
			{pend.Hostname, "custom  false false"},
			{mixed, "custom " + w.org.String() + " false true"},
			{ver.Hostname, "custom " + w.org.String() + " false true"},
			{act.Hostname, "custom " + w.org.String() + " true true"},
			{sus.Hostname, "custom " + w.org.String() + " false true"},
			{rem.Hostname, "custom  false false"},
			{"unknown-" + domHost(), "custom  false false"},
			{label + ".localhost", "platform " + w.org.String() + " true false"},
			{"nobody-" + label + ".localhost", "platform  false false"},
			{slug + ".localhost", "platform  false false"},
			{"localhost", "public_entry  false false"},
		} {
			res, err := r.Resolve(ctx, c.host)
			if err != nil {
				t.Fatalf("%s: %v", c.host, err)
			}
			if got := fmt.Sprintf("%s %s %v %v", res.Kind, res.OrganizationID, res.Routing, res.Issuance); got != c.want {
				t.Errorf("%s: %s, want %s", c.host, got, c.want)
			}
		}
	})

	t.Run("DOM-8 exact audit rows per operation", func(t *testing.T) {
		c := e.claim(t, octx, otc, domHost())
		steps := []struct {
			name string
			do   func()
			want []string
		}{
			{"claim", func() {}, []string{"domain.claimed/success/"}},
			{"verify", func() {
				if _, err := e.check(t, octx, otc, c.ID, domains.CheckVerify, okProof()); err != nil {
					t.Fatal(err)
				}
			}, []string{"domain.checked/success/ok", "domain.state_changed/success/verified"}},
			{"recheck transient", func() {
				if _, err := e.check(t, octx, otc, c.ID, domains.CheckRecheck, domains.Observation{Proof: domains.OutcomeTransient}); err != nil {
					t.Fatal(err)
				}
			}, []string{"domain.checked/failed/transient"}},
			{"activate", func() {
				if _, err := e.check(t, octx, otc, c.ID, domains.CheckRecheck, readyTLS(time.Now().Add(24*time.Hour))); err != nil {
					t.Fatal(err)
				}
			}, []string{"domain.checked/success/ok", "domain.state_changed/success/active"}},
			{"remove", func() {
				rm, err := e.svc.RemoveDomain(octx, otc, c.ID)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := e.svc.RecordWithdrawal(octx, otc, rm, domains.TLSResult{Ref: rm.Ref(), State: domains.TLSWithdrawn}, false); err != nil {
					t.Fatal(err)
				}
			}, []string{"domain.checked/success/withdrawn", "domain.state_changed/success/removed"}},
		}
		var acc []string
		for _, s := range steps {
			s.do()
			acc = append(acc, s.want...)
			slices.Sort(acc)
			if got := e.domEvents(t, c.ID); !slices.Equal(got, acc) {
				t.Errorf("after %s: audit rows %v, want %v", s.name, got, acc)
			}
		}
		// A denied operation records access.denied with its domain operation, no domain row event.
		mctx, mtc := e.as(t, w.member, w.org)
		var n0, n1 int
		count := `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND operation = 'domain.claim' AND organization_id = $1::text::uuid`
		if err := e.setup.QueryRow(ctx, count, w.org.String()).Scan(&n0); err != nil {
			t.Fatal(err)
		}
		_, err := e.svc.ClaimDomain(mctx, mtc, domHost(), domTTL, domLabel)
		wantCode(t, "member claim", err, CodeForbidden)
		if err := e.setup.QueryRow(ctx, count, w.org.String()).Scan(&n1); err != nil || n1 != n0+1 {
			t.Errorf("access.denied domain.claim rows %d -> %d (%v)", n0, n1, err)
		}
		// The slug change is captured with the Organization actions.
		var s0, s1 int
		sc := `SELECT count(*) FROM audit.events WHERE action = 'organization.slug_changed' AND organization_id = $1::text::uuid`
		_ = e.setup.QueryRow(ctx, sc, w.org.String()).Scan(&s0)
		if _, err := e.svc.SetOrganizationSlug(octx, otc, fmt.Sprintf("dom-aud-%d", itSeq.Add(1))); err != nil {
			t.Fatal(err)
		}
		if err := e.setup.QueryRow(ctx, sc, w.org.String()).Scan(&s1); err != nil || s1 != s0+1 {
			t.Errorf("organization.slug_changed rows %d -> %d (%v)", s0, s1, err)
		}
	})

	t.Run("DOM-9 restore and configuration mismatch", func(t *testing.T) {
		// G057: the platform host is the platform label (the slug no longer routes).
		label := fmt.Sprintf("dom-cfg-%d", itSeq.Add(1))
		e.setPlatformLabel(t, w.org, label)
		fctx, ftc := e.as(t, w.foreign, w.other)
		// Before the change, the other Organization holds an active custom host that looks like a label host of the new base.
		stale := "dom-stale-" + fmt.Sprint(itSeq.Add(1)) + ".renamed.example.net"
		e.activeDomain(t, fctx, ftc, stale)
		cust := e.activeDomain(t, octx, otc, domHost())
		before := e.domHash(t, w.org, w.other)
		oldCfg := hostname.Config{BaseDomain: "localhost", CNAMETarget: "edge.localhost"}
		newCfg := hostname.Config{BaseDomain: "renamed.example.net", CNAMETarget: "edge.renamed.example.net"}
		resolve := func(cfg hostname.Config, host string) domains.Resolution {
			res, err := domains.ResolveWith(ctx, e.pool, cfg, host)
			if err != nil {
				t.Fatalf("%s: %v", host, err)
			}
			return res
		}
		if r := resolve(oldCfg, label+".localhost"); !r.Routing || r.OrganizationID != w.org.String() {
			t.Fatalf("old base label host: %+v", r)
		}
		// After the configuration change, platform hosts follow the configuration.
		if r := resolve(newCfg, label+".renamed.example.net"); !r.Routing || r.OrganizationID != w.org.String() || r.Kind != hostname.KindPlatform {
			t.Errorf("new base label host: %+v", r)
		}
		if r := resolve(newCfg, label+".localhost"); r.Routing {
			t.Errorf("the old base label host still routes: %+v", r)
		}
		// Custom rows and their proof are unchanged and still route to their own Organization.
		if r := resolve(newCfg, cust.Hostname); !r.Routing || r.OrganizationID != w.org.String() {
			t.Errorf("custom host after the change: %+v", r)
		}
		// Stale state cannot route to another tenant: the other Organization's custom row under the new base is now a
		// platform name and resolves only through platform labels.
		if r := resolve(newCfg, stale); r.Kind != hostname.KindPlatform || r.OrganizationID == w.other.String() || r.Routing {
			t.Errorf("stale custom row under the new base: %+v", r)
		}
		if e.domHash(t, w.org, w.other) != before {
			t.Fatal("resolution changed rows")
		}
	})

	t.Run("DOM-11 a slug change keeps the platform label, its routing and the sessions", func(t *testing.T) {
		// G057 (Plan r5 §3.2, §5.2 T5): the slug edit route changes metadata only. Through SetOrganizationSlug the
		// platform label, the label host's resolution to the exact UUID and the Organization's live bound sessions stay
		// unchanged, the old and new slug hosts route nowhere, and no runtime context can write the label (no UPDATE
		// (platform_label) for any role: 42501 with zero effect).
		cfg := hostname.Config{BaseDomain: "localhost", Port: "543", CNAMETarget: "edge.localhost"}
		r := domains.NewResolver(e.pool, cfg)
		label := fmt.Sprintf("dom11lbl%d", itSeq.Add(1))
		e.setPlatformLabel(t, w.org, label)
		host := label + ".localhost:543"
		old, next := fmt.Sprintf("dom-slug11a-%d", itSeq.Add(1)), fmt.Sprintf("dom-slug11b-%d", itSeq.Add(1))
		if _, err := e.svc.SetOrganizationSlug(octx, otc, old); err != nil {
			t.Fatal(err)
		}
		address := func() (string, string, int) {
			t.Helper()
			var l, s string
			var live int
			if err := e.setup.QueryRow(ctx, `SELECT o.platform_label, o.slug, (SELECT count(*)::int FROM app.sessions s
  WHERE s.organization_id = o.id AND s.revoked_at IS NULL) FROM app.organizations o WHERE o.id = $1::text::uuid`, w.org.String()).Scan(&l, &s, &live); err != nil {
				t.Fatal(err)
			}
			return l, s, live
		}
		ownerCred := w.owner.in(t, w.org)
		before, err := e.id.Session(ctx, ownerCred, itOriginOf(w.org))
		if err != nil {
			t.Fatalf("owner session before the slug change: %v", err)
		}
		l0, s0, live0 := address()
		if l0 != label || s0 != old || live0 == 0 {
			t.Fatalf("before: label %q slug %q, %d live sessions", l0, s0, live0)
		}
		if got, err := e.svc.SetOrganizationSlug(octx, otc, next); err != nil || got != next {
			t.Fatalf("slug change: %q %v", got, err)
		}
		if l1, s1, live1 := address(); l1 != label || s1 != next || live1 != live0 {
			t.Errorf("after the slug change: label %q slug %q, %d live sessions; want label %q, slug %q, %d", l1, s1, live1, label, next, live0)
		}
		if got, hp, err := r.Lookup(ctx, host); err != nil || got != w.org.String() || hp != host {
			t.Errorf("label host after the slug change: %q %q %v, want %s", got, hp, err, w.org)
		}
		for _, s := range []string{old, next} {
			if _, _, err := r.Lookup(ctx, s+".localhost:543"); !errors.Is(err, domains.ErrUnknownHost) {
				t.Errorf("slug host %s: %v, want unknown_host", s, err)
			}
		}
		if after, err := e.id.Session(ctx, ownerCred, itOriginOf(w.org)); err != nil || after.SessionID != before.SessionID {
			t.Errorf("owner session after the slug change: %q %v, want %q", after.SessionID, err, before.SessionID)
		}
		// No runtime context writes the label, alone or with the slug: refused 42501, the row unchanged.
		direct := func(u *itUser, set string) error {
			tx, err := e.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`, u.id.String(), w.org.String()); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `UPDATE app.organizations SET `+set+` WHERE id = $1::text::uuid`, w.org.String()); err != nil {
				return err
			}
			return tx.Commit(ctx)
		}
		row := e.part(t, "organizations", w.org)
		for _, c := range []struct {
			u   *itUser
			set string
		}{
			{w.owner, fmt.Sprintf("platform_label = 'dom11mv%d'", itSeq.Add(1))},
			{w.admin, fmt.Sprintf("platform_label = 'dom11mv%d'", itSeq.Add(1))},
			{w.owner, fmt.Sprintf("slug = 'dom-slug11c-%d', platform_label = 'dom11mv%d'", itSeq.Add(1), itSeq.Add(1))},
		} {
			var pg *pgconn.PgError
			if err := direct(c.u, c.set); !errors.As(err, &pg) || pg.Code != "42501" {
				t.Errorf("runtime UPDATE %s: %v, want 42501", c.set, err)
			}
		}
		if e.part(t, "organizations", w.org) != row {
			t.Error("a refused label UPDATE changed the Organization row")
		}
	})

	t.Run("DOM-12 the platform label routes to the exact UUID; the slug host routes nowhere", func(t *testing.T) {
		// G057 (Plan r5 §3.1, §5.2 T5, EP-1): through the runtime pool (the attested runtime role) and the auth reader's
		// app.hostname_state, a label host resolves directly to its Organization's UUID. The friendly slug host, an unknown
		// label, a reserved label and suffix-confusion names resolve to no Organization: unknown_host, with nothing to
		// redirect to and nothing disclosed about the label.
		cfg := hostname.Config{BaseDomain: "localhost", Port: "543", CNAMETarget: "edge.localhost"}
		r := domains.NewResolver(e.pool, cfg)
		slug := fmt.Sprintf("dom-slug12-%d", itSeq.Add(1))
		if _, err := e.svc.SetOrganizationSlug(octx, otc, slug); err != nil {
			t.Fatal(err)
		}
		label, other := fmt.Sprintf("dom12lbl%d", itSeq.Add(1)), fmt.Sprintf("dom12oth%d", itSeq.Add(1))
		e.setPlatformLabel(t, w.org, label)
		e.setPlatformLabel(t, w.other, other)
		before := e.domHash(t, w.org, w.other)
		for host, org := range map[string]tenancy.OrganizationID{label + ".localhost:543": w.org, other + ".localhost:543": w.other} {
			got, hp, err := r.Lookup(ctx, host)
			if err != nil || got != org.String() || hp != host {
				t.Errorf("label host %s: %q %q %v, want %s", host, got, hp, err, org)
			}
		}
		for _, host := range []string{
			slug + ".localhost:543",                // the friendly slug: metadata only
			"nobody-" + label + ".localhost:543",   // an unknown label
			"admin.localhost:543",                  // a reserved label
			label + ".localhost.evil.example:543",  // suffix confusion: a custom name, never the label
			"x." + label + ".localhost:543",        // two labels under the platform domain
			label + "-" + other + ".localhost:543", // a concatenation of two real labels
		} {
			res, err := r.Resolve(ctx, host)
			if err != nil {
				t.Fatalf("%s: %v", host, err)
			}
			if res.Routing || res.OrganizationID != "" {
				t.Errorf("%s resolves to %+v, want no Organization", host, res)
			}
			if _, _, err := r.Lookup(ctx, host); !errors.Is(err, domains.ErrUnknownHost) {
				t.Errorf("%s lookup: %v, want unknown_host", host, err)
			}
		}
		// The auth reader's platform branch answers the label exactly once and never the slug.
		for key, want := range map[string]int{label: 1, slug: 0, strings.ToUpper(label): 0} {
			var n int
			if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM app.hostname_state('platform', $1)`, key).Scan(&n); err != nil || n != want {
				t.Errorf("hostname_state('platform', %q): %d rows (%v), want %d", key, n, err, want)
			}
		}
		if e.domHash(t, w.org, w.other) != before {
			t.Fatal("resolution changed rows")
		}
	})

	t.Run("DOM-10 footprint E-DOM-F1..F4", func(t *testing.T) {
		fp := func(name string, v map[string]any) {
			v["evidence"] = name
			b, _ := json.Marshal(v)
			fmt.Println(name + " " + string(b))
		}
		owner := e.newUser(t)
		org := e.org(t, "Domains Footprint")
		e.grantOrg(t, org, owner, tenancy.OrgOwner)
		e.bind(t, owner, org) // G016 A2: setup, before the measured statements
		fctx, ftc := e.as(t, owner, org)
		seed := func(n int) {
			for range n {
				e.claim(t, fctx, ftc, domHost())
			}
			if _, err := e.setup.Exec(ctx, `ANALYZE app.organization_domains; ANALYZE app.organizations`); err != nil {
				t.Fatal(err)
			}
		}
		count := func(f func() error) int {
			e.hook.reset()
			if err := f(); err != nil {
				t.Fatal(err)
			}
			_, s := e.hook.counts()
			return len(s)
		}
		measure := func() map[string]int {
			c := e.claim(t, fctx, ftc, domHost())
			out := map[string]int{}
			out["claim"] = count(func() error { _, err := e.svc.ClaimDomain(fctx, ftc, domHost(), domTTL, domLabel); return err })
			out["list"] = count(func() error { _, err := e.svc.ListDomains(fctx, ftc); return err })
			read, err := e.svc.BeginDomainCheck(fctx, ftc, c.ID, domains.CheckVerify)
			if err != nil {
				t.Fatal(err)
			}
			out["verify_write"] = count(func() error {
				_, err := e.svc.RecordDomainCheck(fctx, ftc, read, domains.CheckVerify, okProof(), domFresh)
				return err
			})
			var rm domains.Record
			out["remove"] = count(func() error { var err error; rm, err = e.svc.RemoveDomain(fctx, ftc, c.ID); return err })
			if _, err := e.svc.RecordWithdrawal(fctx, ftc, rm, domains.TLSResult{Ref: rm.Ref(), State: domains.TLSUnknown}, false); err != nil {
				t.Fatal(err)
			}
			out["retry"] = count(func() error { _, err := e.svc.RetryWithdrawal(fctx, ftc, rm.ID, 20); return err })
			out["set_slug"] = count(func() error {
				_, err := e.svc.SetOrganizationSlug(fctx, ftc, fmt.Sprintf("dom-fp-%d", itSeq.Add(1)))
				return err
			})
			return out
		}
		seed(10)
		at10 := measure()
		seed(30)
		at40 := measure()
		if fmt.Sprint(at10) != fmt.Sprint(at40) {
			fmt.Println("E-DOM-TRIGGER statement fan-out")
			t.Errorf("statement counts grew with the domain rows: %v -> %v", at10, at40)
		}
		// F1: the plans of the two probes of app.hostname_state (as the owner, the same predicates).
		host := domHost()
		e.activeDomain(t, fctx, ftc, host)
		plans := map[string][]string{}
		for name, q := range map[string][2]string{
			"custom": {`EXPLAIN (COSTS OFF) SELECT d.organization_id FROM app.organization_domains d WHERE d.hostname = $1 AND d.state <> 'removed'`, host},
			"slug":   {`EXPLAIN (COSTS OFF) SELECT o.id FROM app.organizations o WHERE o.slug = $1`, "dom-fp"},
		} {
			rows, err := e.setup.Query(ctx, q[0], q[1])
			if err != nil {
				t.Fatal(err)
			}
			for rows.Next() {
				var l string
				if err := rows.Scan(&l); err != nil {
					t.Fatal(err)
				}
				plans[name] = append(plans[name], strings.TrimSpace(l))
			}
			rows.Close()
		}
		fp("E-DOM-F1", map[string]any{"plans": plans})
		// F2: ResolveWith on a transaction is one statement; Resolve on the pool is one autocommit statement.
		cq := &countingQuerier{q: e.pool}
		tx, err := e.pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		cfg := hostname.Config{BaseDomain: "localhost", CNAMETarget: "edge.localhost"}
		txq := &countingQuerier{q: tx}
		if _, err := domains.ResolveWith(ctx, txq, cfg, host); err != nil {
			t.Fatal(err)
		}
		_ = tx.Rollback(ctx)
		if _, err := domains.NewResolver(cq, cfg).Resolve(ctx, host); err != nil {
			t.Fatal(err)
		}
		if txq.n != 1 || cq.n != 1 {
			fmt.Println("E-DOM-TRIGGER resolver statements")
			t.Errorf("resolver statements: tx %d, pool %d; want 1 and 1", txq.n, cq.n)
		}
		fp("E-DOM-F2", map[string]any{"resolve_with_tx_statements": txq.n, "resolve_pool_statements": cq.n, "resolve_pool_begins": 0})
		fp("E-DOM-F3", map[string]any{"statements_at_10": at10, "statements_at_40": at40})
		// F4: the relation locks held by the verify write (paused at its row lock) and by an issuance-shaped transaction.
		c := e.claim(t, fctx, ftc, domHost())
		read, err := e.svc.BeginDomainCheck(fctx, ftc, c.ID, domains.CheckVerify)
		if err != nil {
			t.Fatal(err)
		}
		entered, release := make(chan struct{}), make(chan struct{})
		e.svc.domainHooks = &domainHooks{checkLocked: func() { close(entered); <-release }}
		done := make(chan error, 1)
		go func() {
			_, err := e.svc.RecordDomainCheck(fctx, ftc, read, domains.CheckVerify, okProof(), domFresh)
			done <- err
		}()
		<-entered
		verifyLocks := e.domLocks(t)
		close(release)
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		e.svc.domainHooks = nil
		itx, err := e.pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var ig1, ig2 any
		if err := itx.QueryRow(ctx, `SELECT organization_id FROM app.hostname_state('custom', $1)`, host).Scan(&ig1); err != nil {
			t.Fatal(err)
		}
		if err := itx.QueryRow(ctx, `SELECT app.is_organization_member($1::uuid, $2::uuid)`, owner.id.String(), org.String()).Scan(&ig2); err != nil {
			t.Fatal(err)
		}
		issuanceLocks := e.domLocks(t)
		_ = itx.Rollback(ctx)
		for _, l := range strings.Split(issuanceLocks, ",") {
			if (strings.HasPrefix(l, "organization_domains:") || strings.HasPrefix(l, "role_bindings:")) && !strings.HasSuffix(l, ":AccessShareLock") {
				fmt.Println("E-DOM-TRIGGER issuance lock " + l)
				t.Errorf("issuance-shaped transaction holds %s", l)
			}
		}
		fp("E-DOM-F4", map[string]any{"verify_write_locks": verifyLocks, "issuance_locks": issuanceLocks})
	})
}

// domLocks lists the granted relation-level locks on schema app tables held by sessions other than the setup one.
func (e *itEnv) domLocks(t *testing.T) string {
	t.Helper()
	var s string
	if err := e.setup.QueryRow(context.Background(), `SELECT coalesce(string_agg(DISTINCT c.relname || ':' || l.mode, ',' ORDER BY c.relname || ':' || l.mode), '')
FROM pg_locks l JOIN pg_class c ON c.oid = l.relation
WHERE l.pid <> pg_backend_pid() AND c.relnamespace = 'app'::regnamespace AND c.relkind = 'r' AND l.granted`).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

// countingQuerier counts the statements a resolver sends.
type countingQuerier struct {
	q interface {
		QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	}
	n int
}

func (c *countingQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	c.n++
	return c.q.QueryRow(ctx, sql, args...)
}
