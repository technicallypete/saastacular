//go:build integration

package access

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/identity"
	"core/internal/observability"
	"core/internal/tenancy"
)

// MEM-10/MEM-11 (G019 Plan r1 §6.2; r2 addendum §1–§2, D-15): the bounded Organization-invitation consumer. In G019 only
// the owner role could execute app.invitation_claim_organization, so the contract is tested on an owner-role pool; G058
// D-15(i) grants the runtime EXECUTE (00016), and MEM-10 asserts the runtime pool's valid claim and its no-write replay
// (already_member) at the current schema. Evidence is built by the test through identity's
// AttestVerifiedAddress (a test file may name it; the pin covers non-test files). Pending (NOT EXECUTED) while the
// native test hold applies; DB evidence after Q1.

// memSource is the fixture evidence kind recorded by AttestVerifiedAddress.
const memSource = "test_fixture"

// memOwnerService is an access Service whose pool connects as the attested setup login and assumes <db>_owner on every
// connection (after the setup pool's own connect hook), so the consumer runs as the owner role (D-15).
func (e *itEnv) memOwnerService(t *testing.T) *Service {
	t.Helper()
	return NewService(e.memOwnerPool(t), e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil))
}

// memOwnerHooked is memOwnerService over a counting hookDB (TestAccessFootprint E-PG-MEM).
func (e *itEnv) memOwnerHooked(t *testing.T) (*Service, *hookDB) {
	t.Helper()
	h := &hookDB{pool: e.memOwnerPool(t)}
	return NewService(h, e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil)), h
}

// memOwnerPool is the owner-role pool of memOwnerService.
func (e *itEnv) memOwnerPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pc := e.setupDB.Config().Copy()
	prev := pc.AfterConnect
	setRole := "SET ROLE " + pgx.Identifier{e.cfg.OwnerRole()}.Sanitize()
	pc.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		if prev != nil {
			if err := prev(ctx, c); err != nil {
				return err
			}
		}
		_, err := c.Exec(ctx, setRole)
		return err
	}
	pc.MaxConns = 4
	p, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

// memPrincipal maps a new provider principal to a new user, as an operator or a first sign-in would, without a session
// (a first-time recipient).
func (e *itEnv) memPrincipal(t *testing.T, tag string) (string, tenancy.UserID) {
	t.Helper()
	sub := fmt.Sprintf("user_01MEM%s%08d", tag, itSeq.Add(1))
	var uid string
	if err := e.setupDB.QueryRow(context.Background(), `WITH u AS (INSERT INTO app.users (display_name) VALUES ('First time') RETURNING id),
  p AS (INSERT INTO app.identity_principals (issuer, subject, user_id) SELECT $1, $2, id FROM u)
SELECT id::text FROM u`, e.fake.Issuer, sub).Scan(&uid); err != nil {
		t.Fatalf("map principal: %v", err)
	}
	id, err := tenancy.ParseUserID(uid)
	if err != nil {
		t.Fatal(err)
	}
	return sub, id
}

// memAttest builds verified-address evidence for the principal (issuer, sub).
func (e *itEnv) memAttest(t *testing.T, sub, addr string) identity.VerifiedAddress {
	t.Helper()
	ev, err := e.id.AttestVerifiedAddress(context.Background(), e.fake.Issuer, sub, addr, e.clk.Now(), memSource)
	if err != nil {
		t.Fatalf("attest %s: %v", addr, err)
	}
	return ev
}

// memAddr is a fresh fixture address with the given local-part prefix.
func memAddr(prefix string) string {
	return fmt.Sprintf("%s.%d@example.test", prefix, itSeq.Add(1))
}

// memIssueOrg issues an org_member invitation of org to addr as issuer (a guarded administrator operation).
func (e *itEnv) memIssueOrg(t *testing.T, issuer *itUser, org tenancy.OrganizationID, addr string) IssuedInvitation {
	t.Helper()
	ctx, tc := e.as(t, issuer, org)
	iss, err := e.svc.IssueInvitation(ctx, tc, Scope{Kind: tenancy.ScopeOrganization}, string(tenancy.OrgMember), Recipient{Address: &addr}, 0)
	if err != nil {
		t.Fatalf("issue to %s: %v", addr, err)
	}
	return iss
}

// memInvitation reads an invitation's consumer, intended user and pending state (attested superuser read).
func (e *itEnv) memInvitation(t *testing.T, id tenancy.InvitationID) (consumedBy, intended string, pending bool) {
	t.Helper()
	if err := e.setup.QueryRow(context.Background(), `SELECT coalesce(consumed_by_user_id::text, ''), coalesce(intended_user_id::text, ''),
  consumed_at IS NULL AND revoked_at IS NULL FROM app.invitations WHERE id = $1::text::uuid`, id.String()).Scan(&consumedBy, &intended, &pending); err != nil {
		t.Fatalf("invitation row: %v", err)
	}
	return consumedBy, intended, pending
}

// memOrgBindings counts user's Organization membership bindings in org.
func (e *itEnv) memOrgBindings(t *testing.T, org tenancy.OrganizationID, user tenancy.UserID) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid AND scope_type = 'organization' AND source = 'membership'`, org.String(), user.String()).Scan(&n); err != nil {
		t.Fatalf("bindings: %v", err)
	}
	return n
}

// memConsumeEvents lists the invitation.consumed and membership.granted actions of org recorded for user's binding
// and the invitation, in order.
func (e *itEnv) memConsumeEvents(t *testing.T, org tenancy.OrganizationID, inv tenancy.InvitationID, binding tenancy.BindingID) []string {
	t.Helper()
	rows, err := e.setup.Query(context.Background(), `SELECT action FROM audit.events WHERE organization_id = $1::text::uuid
  AND action IN ('invitation.consumed', 'membership.granted') AND resource_id::text IN ($2, $3) ORDER BY occurred_at, id`,
		org.String(), inv.String(), binding.String())
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
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	slices.Sort(out)
	return out
}

// memDenials counts access.denied records without an Organization attributed to user (Q7).
func (e *itEnv) memDenials(t *testing.T, user tenancy.UserID) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) FROM audit.events WHERE action = 'access.denied'
  AND organization_id IS NULL AND actor_user_id = $1::text::uuid`, user.String()).Scan(&n); err != nil {
		t.Fatalf("denials: %v", err)
	}
	return n
}

// memWantCode asserts err is an access Error with code and (when non-empty) field.
func memWantCode(t *testing.T, what string, err error, code Code, field string) {
	t.Helper()
	var ae *Error
	if !errors.As(err, &ae) || ae.Code != code || (field != "" && ae.Field != field) {
		t.Errorf("%s: %v, want %s %s", what, err, code, field)
	}
}

// mem10ConsumeOrganizationInvitation: a first-time user (a mapped principal without a session) and an existing user
// of Organization B both join A through address evidence; the claim sets intended_user_id and consumes; the binding
// and its audit pair are recorded; since G058 D-15(i) the runtime pool's valid claim succeeds the same way and its replay
// is not re-executed (already_member with the claimed binding, no further write).
func mem10ConsumeOrganizationInvitation(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	owner := e.memOwnerService(t)
	o := e.newUser(t)
	a, b := e.org(t, "MEM10 A"), e.org(t, "MEM10 B")
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.bind(t, o, a) // G016 A2: the issuer's credential for A, setup only

	check := func(what string, user tenancy.UserID, iss IssuedInvitation, out AcceptOutcome, err error) {
		t.Helper()
		if err != nil || out.Outcome != tenancy.ClaimJoined || out.Binding == nil || out.Binding.OrganizationID != a ||
			out.Binding.SubjectUserID != user || out.Binding.Role != string(tenancy.OrgMember) {
			t.Fatalf("%s: %v, outcome %+v", what, err, out)
		}
		by, intended, pending := e.memInvitation(t, iss.Invitation.ID)
		if by != user.String() || intended != user.String() || pending {
			t.Errorf("%s invitation: consumed by %q, intended %q, pending %v", what, by, intended, pending)
		}
		if n := e.memOrgBindings(t, a, user); n != 1 {
			t.Errorf("%s bindings: %d, want 1", what, n)
		}
		if ev := e.memConsumeEvents(t, a, iss.Invitation.ID, out.Binding.ID); !slices.Equal(ev, []string{"invitation.consumed", "membership.granted"}) {
			t.Errorf("%s audit: %v", what, ev)
		}
	}

	// First-time user: the principal is mapped, then the evidence is attested and consumed (no session).
	sub, first := e.memPrincipal(t, "F")
	addr := memAddr("first")
	iss := e.memIssueOrg(t, o, a, addr)
	out, err := owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, sub, addr), iss.Invitation.ID, iss.Token)
	check("first-time", first, iss, out, err)

	// An existing user of B joins A; B's membership is untouched.
	u := e.newUser(t)
	e.grantOrg(t, b, u, tenancy.OrgMember)
	addr = memAddr("existing")
	iss = e.memIssueOrg(t, o, a, addr)
	out, err = owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, u.sub, addr), iss.Invitation.ID, iss.Token)
	check("existing", u.id, iss, out, err)
	if n := e.memOrgBindings(t, b, u.id); n != 1 {
		t.Errorf("B binding: %d, want 1", n)
	}

	// G058 D-15(i) (owner-selected; 00016 grants the runtime EXECUTE): a valid first-time recipient's claim through the
	// current runtime pool succeeds atomically — the same consumption, intended user, single binding and audit pair as the
	// owner-role path. Every invalid principal/context/token/address/expiry/replay/authority case stays refused (MEM-11,
	// the CT-MEM-1 claim-level cases and the runtime replay below).
	sub, rt := e.memPrincipal(t, "R")
	addr = memAddr("runtime")
	iss = e.memIssueOrg(t, o, a, addr)
	rtEv := e.memAttest(t, sub, addr)
	out, err = e.svc.ConsumeOrganizationInvitation(ctx, rtEv, iss.Invitation.ID, iss.Token)
	check("runtime pool", rt, iss, out, err)
	bid := out.Binding.ID
	// A replay of the consumed invitation by the same recipient through the runtime pool is not re-executed and writes
	// nothing more: G019's claim body (00015 "replay by the same recipient") returns already_member with the current
	// qualifying binding — the one the claim created — and never restores or adds a binding.
	out, err = e.svc.ConsumeOrganizationInvitation(ctx, rtEv, iss.Invitation.ID, iss.Token)
	if err != nil || out.Outcome != tenancy.ClaimAlreadyMember || out.Binding == nil || out.Binding.ID != bid {
		t.Errorf("runtime replay of a consumed invitation: %v, outcome %+v, want already_member with the claimed binding", err, out)
	}
	if n := e.memOrgBindings(t, a, rt); n != 1 {
		t.Errorf("runtime replay bindings: %d, want 1", n)
	}
	if by, intended, pending := e.memInvitation(t, iss.Invitation.ID); by != rt.String() || intended != rt.String() || pending {
		t.Errorf("runtime replay invitation: consumed by %q, intended %q, pending %v", by, intended, pending)
	}
	if ev := e.memConsumeEvents(t, a, iss.Invitation.ID, bid); !slices.Equal(ev, []string{"invitation.consumed", "membership.granted"}) {
		t.Errorf("runtime replay audit: %v", ev)
	}
}

// mem11ConsumeOrganizationNegatives: every refusal is not_found or conflict with zero rows written: expired,
// revoked, consumed by another user, wrong token, another invitation's token, Group/Organization token crossover,
// address mismatch (a local-part case variant is a different address; a domain case variant matches), disabled
// recipient, and an issuer demoted after issue (issuer_authority_lost, still pending, the administrators' revoke works).
// The Group-claim recipient whose Organization binding was removed after issue is refused by the facade guard before
// the claim, so its SQL-level recipient_ineligible branch stays with the CT-MEM-1/claim-level cases.
func mem11ConsumeOrganizationNegatives(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	owner := e.memOwnerService(t)
	o, admin := e.newUser(t), e.newUser(t)
	a := e.org(t, "MEM11 A")
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.grantOrg(t, a, admin, tenancy.OrgAdmin)
	e.bind(t, o, a) // G016 A2: the issuers' credentials for A, setup only
	e.bind(t, admin, a)
	octx, otc := e.as(t, o, a)

	refused := func(what string, user tenancy.UserID, iss IssuedInvitation, err error, code Code, field string) {
		t.Helper()
		memWantCode(t, what, err, code, field)
		if n := e.memOrgBindings(t, a, user); n != 0 {
			t.Errorf("%s: %d bindings written", what, n)
		}
	}

	t.Run("expired", func(t *testing.T) {
		sub, user := e.memPrincipal(t, "E")
		addr := memAddr("expired")
		iss := e.memIssueOrg(t, o, a, addr)
		// Fixture-only time travel as the attested superuser with triggers disabled for this statement (as timeTravel):
		// invitations_keys_immutable refuses any expires_at change by design; this simulates elapsed time.
		tt, err := e.setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tt.Rollback(ctx) }()
		if _, err := tt.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
			t.Fatal(err)
		}
		if _, err := tt.Exec(ctx, `UPDATE app.invitations SET created_at = now() - interval '2 hours',
  expires_at = now() - interval '1 hour' WHERE id = $1::text::uuid`, iss.Invitation.ID.String()); err != nil {
			t.Fatalf("time travel: %v", err)
		}
		if err := tt.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		before := e.memDenials(t, user)
		_, err = owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, sub, addr), iss.Invitation.ID, iss.Token)
		refused("expired", user, iss, err, CodeNotFound, "")
		if after := e.memDenials(t, user); after != before+1 {
			t.Errorf("expired denial records: %d -> %d, want one access.denied without an Organization", before, after)
		}
	})

	t.Run("revoked", func(t *testing.T) {
		sub, user := e.memPrincipal(t, "V")
		addr := memAddr("revoked")
		iss := e.memIssueOrg(t, o, a, addr)
		if err := e.svc.RevokeInvitation(octx, otc, iss.Invitation.ID); err != nil {
			t.Fatal(err)
		}
		_, err := owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, sub, addr), iss.Invitation.ID, iss.Token)
		refused("revoked", user, iss, err, CodeNotFound, "")
	})

	t.Run("consumed_by_another", func(t *testing.T) {
		addr := memAddr("shared")
		iss := e.memIssueOrg(t, o, a, addr)
		sub1, _ := e.memPrincipal(t, "C")
		if _, err := owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, sub1, addr), iss.Invitation.ID, iss.Token); err != nil {
			t.Fatalf("first consumer: %v", err)
		}
		sub2, other := e.memPrincipal(t, "D")
		_, err := owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, sub2, addr), iss.Invitation.ID, iss.Token)
		refused("consumed by another", other, iss, err, CodeNotFound, "")
	})

	t.Run("wrong_and_foreign_token", func(t *testing.T) {
		sub, user := e.memPrincipal(t, "T")
		addr := memAddr("token")
		iss, foreign := e.memIssueOrg(t, o, a, addr), e.memIssueOrg(t, o, a, memAddr("foreign"))
		ev := e.memAttest(t, sub, addr)
		wrong := []byte(iss.Token)
		if wrong[0] == 'A' {
			wrong[0] = 'B'
		} else {
			wrong[0] = 'A'
		}
		_, err := owner.ConsumeOrganizationInvitation(ctx, ev, iss.Invitation.ID, string(wrong))
		refused("wrong token", user, iss, err, CodeNotFound, "")
		_, err = owner.ConsumeOrganizationInvitation(ctx, ev, iss.Invitation.ID, foreign.Token)
		refused("another invitation's token", user, iss, err, CodeNotFound, "")
		if _, _, pending := e.memInvitation(t, iss.Invitation.ID); !pending {
			t.Error("wrong tokens consumed the invitation")
		}
	})

	t.Run("token_kind_crossover", func(t *testing.T) {
		m := e.newUser(t)
		e.grantOrg(t, a, m, tenancy.OrgMember)
		e.bind(t, m, a) // G016 A2, setup only
		g := e.group(t, a, o, "MEM11 Group", fmt.Sprintf("mem11-%d", itSeq.Add(1)))
		gid := g.ID
		giss, err := e.svc.IssueInvitation(octx, otc, Scope{Kind: tenancy.ScopeGroup, Group: &gid}, string(tenancy.GroupMember),
			Recipient{UserID: &m.id}, 0)
		if err != nil {
			t.Fatalf("group issue: %v", err)
		}
		addr := memAddr("crossover")
		_, err = owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, m.sub, addr), giss.Invitation.ID, giss.Token)
		memWantCode(t, "group token to the Organization consumer", err, CodeNotFound, "")
		oiss := e.memIssueOrg(t, o, a, addr)
		mctx, mtc := e.as(t, m, a)
		_, err = e.svc.AcceptGroupInvitation(mctx, mtc, oiss.Invitation.ID, oiss.Token)
		memWantCode(t, "Organization token to the Group consumer", err, CodeNotFound, "")
		for _, id := range []tenancy.InvitationID{giss.Invitation.ID, oiss.Invitation.ID} {
			if _, _, pending := e.memInvitation(t, id); !pending {
				t.Errorf("crossover consumed %s", id)
			}
		}
	})

	t.Run("address_case", func(t *testing.T) {
		n := itSeq.Add(1)
		addr := fmt.Sprintf("Ada.%d@example.test", n)
		iss := e.memIssueOrg(t, o, a, addr)
		sub, user := e.memPrincipal(t, "A")
		_, err := owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, sub, fmt.Sprintf("ada.%d@example.test", n)), iss.Invitation.ID, iss.Token)
		refused("local-part case variant", user, iss, err, CodeNotFound, "")
		out, err := owner.ConsumeOrganizationInvitation(ctx, e.memAttest(t, sub, fmt.Sprintf("Ada.%d@EXAMPLE.TEST", n)), iss.Invitation.ID, iss.Token)
		if err != nil || out.Outcome != tenancy.ClaimJoined {
			t.Errorf("domain case variant: %v, outcome %+v", err, out)
		}
	})

	t.Run("disabled_recipient", func(t *testing.T) {
		sub, user := e.memPrincipal(t, "X")
		addr := memAddr("disabled")
		iss := e.memIssueOrg(t, o, a, addr)
		ev := e.memAttest(t, sub, addr)
		if err := e.id.DisableUser(ctx, user.String()); err != nil {
			t.Fatal(err)
		}
		_, err := owner.ConsumeOrganizationInvitation(ctx, ev, iss.Invitation.ID, iss.Token)
		refused("disabled recipient", user, iss, err, CodeConflict, string(tenancy.CodeRecipientIneligible))
		if _, _, pending := e.memInvitation(t, iss.Invitation.ID); !pending {
			t.Error("disabled recipient consumed the invitation")
		}
	})

	t.Run("issuer_demoted", func(t *testing.T) {
		sub, user := e.memPrincipal(t, "I")
		addr := memAddr("demoted")
		iss := e.memIssueOrg(t, admin, a, addr)
		if _, err := e.setup.Exec(ctx, `UPDATE app.role_bindings SET role = 'org_member' WHERE organization_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid AND scope_type = 'organization'`, a.String(), admin.id.String()); err != nil {
			t.Fatal(err)
		}
		ev := e.memAttest(t, sub, addr)
		_, err := owner.ConsumeOrganizationInvitation(ctx, ev, iss.Invitation.ID, iss.Token)
		refused("issuer demoted", user, iss, err, CodeConflict, string(tenancy.CodeIssuerAuthorityLost))
		if _, _, pending := e.memInvitation(t, iss.Invitation.ID); !pending {
			t.Fatal("issuer demotion revoked or consumed the invitation")
		}
		if err := e.svc.RevokeInvitation(octx, otc, iss.Invitation.ID); err != nil {
			t.Fatalf("administrators' revoke: %v", err)
		}
		_, err = owner.ConsumeOrganizationInvitation(ctx, ev, iss.Invitation.ID, iss.Token)
		refused("after revoke", user, iss, err, CodeNotFound, "")
	})
}

// --- G019 writer25 membership-gap additions (Codex writer23 review, "bounded membership gaps required before merge") ---

// memExpire backdates invitation id to expired: fixture-only time travel as the attested superuser with triggers disabled
// for the statement (as mem11's expired case; invitations_keys_immutable refuses any expires_at change by design).
func (e *itEnv) memExpire(t *testing.T, id tenancy.InvitationID) {
	t.Helper()
	ctx := context.Background()
	tt, err := e.setup.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tt.Rollback(ctx) }()
	if _, err := tt.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}
	if _, err := tt.Exec(ctx, `UPDATE app.invitations SET created_at = now() - interval '2 hours',
  expires_at = now() - interval '1 hour' WHERE id = $1::text::uuid`, id.String()); err != nil {
		t.Fatalf("time travel: %v", err)
	}
	if err := tt.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

// memCount is one attested superuser count.
func (e *itEnv) memCount(t *testing.T, q string, args ...any) int {
	t.Helper()
	var n int
	if err := e.setup.QueryRow(context.Background(), q, args...).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// memRole reads user's role in the scope row scope (an Organization or Group id), or "none".
func (e *itEnv) memRole(t *testing.T, scope string, user tenancy.UserID) string {
	t.Helper()
	var r string
	if err := e.setup.QueryRow(context.Background(), `SELECT coalesce((SELECT role FROM app.role_bindings WHERE scope_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid), 'none')`, scope, user.String()).Scan(&r); err != nil {
		t.Fatalf("role: %v", err)
	}
	return r
}

// memInvitationRow reads an invitation's consumed_at/revoked_at presence and its consumed_binding_id (empty when NULL).
func (e *itEnv) memInvitationRow(t *testing.T, id tenancy.InvitationID) (consumed, revoked bool, consumedBinding string) {
	t.Helper()
	if err := e.setup.QueryRow(context.Background(), `SELECT consumed_at IS NOT NULL, revoked_at IS NOT NULL,
  coalesce(consumed_binding_id::text, '') FROM app.invitations WHERE id = $1::text::uuid`, id.String()).Scan(&consumed, &revoked, &consumedBinding); err != nil {
		t.Fatalf("invitation row: %v", err)
	}
	return consumed, revoked, consumedBinding
}

// memOpRecords returns the access.decision records of operation op written since byte offset from.
func (e *itEnv) memOpRecords(t *testing.T, from int, op string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, r := range e.recordsSince(t, from) {
		if r["operation"] == op {
			out = append(out, r)
		}
	}
	return out
}

// mem8CapRaceAndPaging (MEM-8; review item 1): one expired plus 99 pending Organization invitations, then two concurrent
// issues serialized on the Organization row (the first paused before its COMMIT, the second queued on the row): exactly
// one commits (pending 100; the expired row is not counted, or both would be refused), the loser is
// conflict/invitation_limit with no row and no invitation.issued capture; the 101st pending issue is refused the same
// way; the administrators' list pages the 101 rows at 50 per page: three pages, every row once, the final row reachable
// and Next nil on the last page.
func mem8CapRaceAndPaging(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o := e.newUser(t)
	a := e.org(t, "MEM8 A")
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.bind(t, o, a) // G016 A2: the issuer's credential for A, setup only
	octx, otc := e.as(t, o, a)
	orgScope := Scope{Kind: tenancy.ScopeOrganization}
	issue := func(svc *Service, c context.Context, tc *TenantContext) (IssuedInvitation, error) {
		addr := memAddr("cap")
		return svc.IssueInvitation(c, tc, orgScope, string(tenancy.OrgMember), Recipient{Address: &addr}, 0)
	}
	counts := func() (pending, rows, issued int) {
		pending = e.memCount(t, `SELECT count(*) FROM app.invitations WHERE organization_id = $1::text::uuid
  AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at > now()`, a.String())
		rows = e.memCount(t, `SELECT count(*) FROM app.invitations WHERE organization_id = $1::text::uuid`, a.String())
		issued = e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = 'invitation.issued'`, a.String())
		return
	}
	wantCounts := func(what string, pending, rows, issued int) {
		t.Helper()
		if p, r, n := counts(); p != pending || r != rows || n != issued {
			t.Errorf("%s: pending %d, rows %d, invitation.issued %d, want %d/%d/%d", what, p, r, n, pending, rows, issued)
		}
	}

	first, err := issue(e.svc, octx, otc)
	if err != nil {
		t.Fatalf("first issue: %v", err)
	}
	e.memExpire(t, first.Invitation.ID)
	for i := 1; i < tenancy.MaxPendingInvitations; i++ {
		if _, err := issue(e.svc, octx, otc); err != nil {
			t.Fatalf("pending issue %d: %v", i, err)
		}
	}
	wantCounts("before the race", tenancy.MaxPendingInvitations-1, tenancy.MaxPendingInvitations, tenancy.MaxPendingInvitations)

	// The race: s1 pauses before its COMMIT holding the Organization row at exclusive strength; s2 queues on that row,
	// then counts 100 pending and is refused.
	s1, h1 := e.service(e.id)
	s2, _ := e.service(e.id)
	c1, _ := beginAs(t, s1, ctx, o, a)
	tc1 := resolve(t, s1, c1, a, tenancy.GroupID{})
	c2, _ := beginAs(t, s2, ctx, o, a)
	tc2 := resolve(t, s2, c2, a, tenancy.GroupID{})
	var winner IssuedInvitation
	var e1, e2 error
	pp := newPause()
	h1.on("before", "COMMIT", pp.hook)
	w := start(func() error { winner, e1 = issue(s1, c1, tc1); return e1 })
	pp.wait(t, "first issue")
	l := start(func() error { _, e2 = issue(s2, c2, tc2); return e2 })
	e.waitWaiters(t, 1)
	close(pp.release)
	_ = w.wait(t, "first issue")
	_ = l.wait(t, "second issue")
	if e1 != nil || winner.Token == "" {
		t.Errorf("winner: %v", e1)
	}
	memWantCode(t, "loser", e2, CodeConflict, string(tenancy.CodeInvitationLimit))
	wantCounts("after the race", tenancy.MaxPendingInvitations, tenancy.MaxPendingInvitations+1, tenancy.MaxPendingInvitations+1)

	// The 101st pending issue is refused at the bound with nothing written.
	_, err = issue(e.svc, octx, otc)
	memWantCode(t, "101st pending issue", err, CodeConflict, string(tenancy.CodeInvitationLimit))
	wantCounts("after the refused 101st", tenancy.MaxPendingInvitations, tenancy.MaxPendingInvitations+1, tenancy.MaxPendingInvitations+1)

	// Paging across the 101 rows (all statuses) at 50 per page.
	seen := map[tenancy.InvitationID]bool{}
	after, pages := "", 0
	for {
		pg, err := e.svc.ListInvitations(octx, otc, orgScope, after, 50)
		if err != nil {
			t.Fatalf("page %d: %v", pages+1, err)
		}
		pages++
		for _, inv := range pg.Invitations {
			if seen[inv.ID] {
				t.Errorf("page %d repeats %s", pages, inv.ID)
			}
			seen[inv.ID] = true
		}
		if pg.Next == nil {
			if len(pg.Invitations) != 1 {
				t.Errorf("last page holds %d rows, want 1", len(pg.Invitations))
			}
			break
		}
		if len(pg.Invitations) != 50 || pages > 3 {
			t.Fatalf("page %d holds %d rows with a Next cursor", pages, len(pg.Invitations))
		}
		after = pg.Next.String()
	}
	if pages != 3 || len(seen) != tenancy.MaxPendingInvitations+1 || !seen[first.Invitation.ID] || !seen[winner.Invitation.ID] {
		t.Errorf("paging: %d pages, %d distinct rows (expired %v, winner %v), want 3 pages of 101 rows", pages, len(seen),
			seen[first.Invitation.ID], seen[winner.Invitation.ID])
	}
}

// mem9UpgradeVisibilityConsumedRevoke (MEM-9 and the consumed-intent revocation, SPEC-R3 row 6; review item 1): a Group
// invitation of group_admin to a current group_member of the Group.
//   - visibility: the issuing Group owner reads it; the intended recipient reads it and lists it among its own; an
//     unrelated Organization member and a foreign Organization's owner (its own context) are not_found and list nothing.
//     The Organization-invitation recipient has no identity before consumption, so its positive is MEM-10's claim; the
//     CT-MEM foreign/no-context rows and the digest ACL proof stay where they are.
//   - upgraded: acceptance upgrades the existing binding in place: the same binding ID now group_admin, outcome
//     upgraded, consumed_binding_id = that binding, one invitation.consumed and no membership.granted; a replay is
//     already_member on the same binding with no second effect.
//   - consumed_revoke: the issuer's revoke of the consumed invitation is conflict/invitation_not_pending; the row stays
//     consumed and unrevoked, the binding unchanged, and no invitation.revoked is recorded.
func mem9UpgradeVisibilityConsumedRevoke(t *testing.T) {
	e := setupIT(t)
	o, m, u2, f := e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t)
	a, b := e.org(t, "MEM9 A"), e.org(t, "MEM9 B")
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.grantOrg(t, a, m, tenancy.OrgMember)
	e.grantOrg(t, a, u2, tenancy.OrgMember)
	e.grantOrg(t, b, f, tenancy.OrgOwner)
	g := e.group(t, a, o, "MEM9 Group", fmt.Sprintf("mem9-%d", itSeq.Add(1)))
	e.grantGroup(t, a, g.ID, m, tenancy.GroupMember)
	for _, u := range []*itUser{o, m, u2} {
		e.bind(t, u, a) // G016 A2: credentials for A, setup only
	}
	e.bind(t, f, b)
	octx, otc := e.as(t, o, a)
	gid := g.ID
	gScope := Scope{Kind: tenancy.ScopeGroup, Group: &gid}
	iss, err := e.svc.IssueInvitation(octx, otc, gScope, string(tenancy.GroupAdmin), Recipient{UserID: &m.id}, 0)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	var bindingID string
	if err := e.setup.QueryRow(context.Background(), `SELECT id::text FROM app.role_bindings WHERE scope_type = 'group'
  AND scope_id = $1::text::uuid AND subject_user_id = $2::text::uuid`, g.ID.String(), m.id.String()).Scan(&bindingID); err != nil {
		t.Fatalf("existing binding: %v", err)
	}
	groupBindings := func() int {
		return e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE scope_type = 'group' AND scope_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid`, g.ID.String(), m.id.String())
	}
	events := func(action string) int {
		return e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = $2
  AND resource_id::text IN ($3, $4)`, a.String(), action, iss.Invitation.ID.String(), bindingID)
	}
	listed := func(c context.Context, tc *TenantContext) int {
		pg, err := e.svc.ListMyInvitations(c, tc, "", 0)
		if err != nil {
			t.Fatalf("my invitations: %v", err)
		}
		n := 0
		for _, inv := range pg.Invitations {
			if inv.ID == iss.Invitation.ID {
				n++
			}
		}
		return n
	}

	t.Run("visibility", func(t *testing.T) {
		got, err := e.svc.GetInvitation(octx, otc, iss.Invitation.ID)
		if err != nil || got.IntendedUserID == nil || *got.IntendedUserID != m.id || got.Status != "pending" || got.Role != string(tenancy.GroupAdmin) {
			t.Errorf("issuing Group owner: %v, %+v", err, got)
		}
		mctx, mtc := e.as(t, m, a)
		got, err = e.svc.GetInvitation(mctx, mtc, iss.Invitation.ID)
		if err != nil || got.IntendedUserID == nil || *got.IntendedUserID != m.id || got.IntendedAddress != nil || got.Status != "pending" {
			t.Errorf("intended recipient: %v, %+v", err, got)
		}
		if n := listed(mctx, mtc); n != 1 {
			t.Errorf("recipient's own list holds the invitation %d times, want 1", n)
		}
		u2ctx, u2tc := e.as(t, u2, a)
		_, err = e.svc.GetInvitation(u2ctx, u2tc, iss.Invitation.ID)
		memWantCode(t, "unrelated Organization member", err, CodeNotFound, "")
		if n := listed(u2ctx, u2tc); n != 0 {
			t.Errorf("unrelated member's own list holds the invitation %d times", n)
		}
		fctx, ftc := e.as(t, f, b)
		_, err = e.svc.GetInvitation(fctx, ftc, iss.Invitation.ID)
		memWantCode(t, "foreign Organization's owner", err, CodeNotFound, "")
		if n := listed(fctx, ftc); n != 0 {
			t.Errorf("foreign owner's own list holds the invitation %d times", n)
		}
		// An ordinary runtime SELECT in the IDENTITY shape (the known intended recipient initialized, NULL Organization and
		// Group) returns no invitation row; the row exists (setup read).
		if n := e.memCount(t, `SELECT count(*) FROM app.invitations WHERE id = $1::text::uuid AND intended_user_id = $2::text::uuid`,
			iss.Invitation.ID.String(), m.id.String()); n != 1 {
			t.Fatalf("setup read of the invitation: %d rows", n)
		}
		bg := context.Background()
		rtx, err := e.runtimePool(t, nil).Begin(bg)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rtx.Rollback(bg) }()
		if err := e.rlsAttest(rtx); err != nil {
			t.Fatal(err)
		}
		if err := tenancy.InitContext(bg, rtx, m.id, tenancy.OrganizationID{}, tenancy.GroupID{}); err != nil {
			t.Fatalf("IDENTITY-shape context: %v", err)
		}
		var shape string
		if err := rtx.QueryRow(bg, `SELECT concat_ws('|', current_setting('app.user_id', true), coalesce(nullif(current_setting('app.organization_id', true), ''), 'NULL'),
  coalesce(nullif(current_setting('app.group_id', true), ''), 'NULL'))`).Scan(&shape); err != nil || shape != m.id.String()+"|NULL|NULL" {
			t.Fatalf("IDENTITY-shape context settings %q (%v)", shape, err)
		}
		var rows int
		if err := rtx.QueryRow(bg, `SELECT count(*) FROM app.invitations WHERE id = $1::text::uuid`, iss.Invitation.ID.String()).Scan(&rows); err != nil || rows != 0 {
			t.Errorf("IDENTITY-shape runtime SELECT of the recipient's invitation: %d rows (%v), want 0", rows, err)
		}
		if consumed, revoked, _ := e.memInvitationRow(t, iss.Invitation.ID); consumed || revoked {
			t.Errorf("reads changed the row: consumed %v, revoked %v", consumed, revoked)
		}
	})

	t.Run("upgraded", func(t *testing.T) {
		// Deltas: the fixture's own grant of the existing binding already recorded one membership.granted for it.
		consumed0, granted0, changed0 := events("invitation.consumed"), events("membership.granted"), events("membership.role_changed")
		mctx, mtc := e.as(t, m, a)
		out, err := e.svc.AcceptGroupInvitation(mctx, mtc, iss.Invitation.ID, iss.Token)
		if err != nil || out.Outcome != tenancy.ClaimUpgraded || out.Binding == nil || out.Binding.ID.String() != bindingID ||
			out.Binding.Role != string(tenancy.GroupAdmin) || out.Binding.SubjectUserID != m.id {
			t.Fatalf("accept: %v, outcome %+v (existing binding %s)", err, out, bindingID)
		}
		consumed, revoked, cb := e.memInvitationRow(t, iss.Invitation.ID)
		if !consumed || revoked || cb != bindingID {
			t.Errorf("row after the upgrade: consumed %v, revoked %v, consumed_binding_id %q, want the existing binding %s", consumed, revoked, cb, bindingID)
		}
		if r, n := e.memRole(t, g.ID.String(), m.id), groupBindings(); r != string(tenancy.GroupAdmin) || n != 1 {
			t.Errorf("binding after the upgrade: role %s, %d rows", r, n)
		}
		if c, gr, rc := events("invitation.consumed")-consumed0, events("membership.granted")-granted0, events("membership.role_changed")-changed0; c != 1 || gr != 0 || rc != 1 {
			t.Errorf("audit after the upgrade: invitation.consumed +%d, membership.granted +%d, membership.role_changed +%d, want +1/+0/+1", c, gr, rc)
		}
		// Replay: the consumption is recorded once; the current qualifying binding answers already_member, unchanged.
		out, err = e.svc.AcceptGroupInvitation(mctx, mtc, iss.Invitation.ID, iss.Token)
		if err != nil || out.Outcome != tenancy.ClaimAlreadyMember || out.Binding == nil || out.Binding.ID.String() != bindingID ||
			out.Binding.Role != string(tenancy.GroupAdmin) {
			t.Errorf("replay: %v, outcome %+v", err, out)
		}
		if r, n := e.memRole(t, g.ID.String(), m.id), groupBindings(); r != string(tenancy.GroupAdmin) || n != 1 {
			t.Errorf("binding after the replay: role %s, %d rows", r, n)
		}
		if c, gr, rc := events("invitation.consumed")-consumed0, events("membership.granted")-granted0, events("membership.role_changed")-changed0; c != 1 || gr != 0 || rc != 1 {
			t.Errorf("audit after the replay: invitation.consumed +%d, membership.granted +%d, membership.role_changed +%d, want +1/+0/+1", c, gr, rc)
		}
	})

	t.Run("consumed_revoke", func(t *testing.T) {
		before := e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = 'invitation.revoked'`, a.String())
		err := e.svc.RevokeInvitation(octx, otc, iss.Invitation.ID)
		memWantCode(t, "revoke of the consumed invitation", err, CodeConflict, string(tenancy.CodeInvitationNotPending))
		consumed, revoked, cb := e.memInvitationRow(t, iss.Invitation.ID)
		if !consumed || revoked || cb != bindingID {
			t.Errorf("row after the refused revoke: consumed %v, revoked %v, consumed_binding_id %q", consumed, revoked, cb)
		}
		if r, n := e.memRole(t, g.ID.String(), m.id), groupBindings(); r != string(tenancy.GroupAdmin) || n != 1 {
			t.Errorf("binding after the refused revoke: role %s, %d rows", r, n)
		}
		if after := e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid AND action = 'invitation.revoked'`, a.String()); after != before {
			t.Errorf("invitation.revoked records %d -> %d after the refused revoke", before, after)
		}
	})
}
