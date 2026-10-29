//go:build integration

package access

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/observability"
	"core/internal/tenancy"
)

// TestEnrollmentAdmission is the G058 slice-1a admission entry point (Plan r3 §6.2 ENR-7/8/11/23 and the open kind;
// readiness correction-01 §1 rows 3 and 6). Each subtest builds its own environment (TestMembership's convention); a
// row selector is '^TestEnrollmentAdmission$/^ENR-7_'. The enrollment credential is a fixture row of
// app.enrollment_sessions inserted through the attested setup connection (identity's callback branch is not part of
// this slice's access tests); the address evidence is identity's AttestVerifiedAddress (memAttest).
func TestEnrollmentAdmission(t *testing.T) {
	for _, c := range []struct {
		name string
		run  func(*testing.T)
	}{
		{"ENR-7 first time invitation claim and lift", enr7FirstTimeInvitation},
		{"ENR-23 stale invitation rolled back", enr23StaleInvitation},
		{"ENR-8 replay and wrong token", enr8ReplayAndWrongToken},
		{"ENR-open restricted and closed", enrOpenRestrictedAndClosed},
		{"ENR-11 fault rolls back claim", enr11FaultRollsBackClaim},
		// G058 round 1a-5 (round-4 Result designs; Codex round-4 correction).
		{"ENR-HOST host guard on the admission transaction", enrHostGuard},
		{"ENR-AUD known refusals recorded without an Organization", enrAuditRefusals},
		{"ENR-LOST unknown COMMIT reconciles to already_member", enrLostCommit},
		{"ENR-REMOVED stale or replaced admitted binding", enrRemovedBinding},
		{"ENR-INTENT callback T1 and T2 through the bound Admitter", enrCallbackIntent},
		{"O-ENR-8 claim versus departure interleavings", oEnr8ClaimDeparture},
	} {
		t.Run(c.name, c.run)
	}
}

// enrHandle inserts a fresh pending enrollment credential of user for org (digest = sha256 of 32 random bytes, a fresh
// provider session id, expiring in 20 minutes) and returns its resolved handle.
func (e *itEnv) enrHandle(t *testing.T, user tenancy.UserID, org tenancy.OrganizationID) identity.EnrollmentHandle {
	t.Helper()
	var cred, csrf [32]byte
	if _, err := rand.Read(cred[:]); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(csrf[:]); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(cred[:])
	origin := itOriginOf(org)
	if _, err := e.setup.Exec(context.Background(), `INSERT INTO app.enrollment_sessions (id, user_id, organization_id, origin,
  credential_hash, csrf_token, provider_session_id, expires_at) VALUES (gen_random_uuid(), $1::text::uuid, $2::text::uuid, $3,
  $4, $5, $6, now() + interval '20 minutes')`, user.String(), org.String(), origin, digest[:], csrf[:],
		fmt.Sprintf("session_01ENR%08d", itSeq.Add(1))); err != nil {
		t.Fatalf("enrollment session fixture: %v", err)
	}
	return identity.EnrollmentHandle{UserID: user.String(), OrganizationID: org.String(), Origin: origin, Digest: digest}
}

// enrAdmitted reads the credential's T2 proof: admitted_at set and admitted_binding_id (empty when NULL).
func (e *itEnv) enrAdmitted(t *testing.T, h identity.EnrollmentHandle) (bool, string) {
	t.Helper()
	var ok bool
	var b string
	if err := e.setup.QueryRow(context.Background(), `SELECT admitted_at IS NOT NULL, coalesce(admitted_binding_id::text, '')
  FROM app.enrollment_sessions WHERE credential_hash = $1`, h.Digest[:]).Scan(&ok, &b); err != nil {
		t.Fatalf("enrollment session row: %v", err)
	}
	return ok, b
}

// enrRestrictions counts user's unlifted restrictions in org and those lifted through an invitation.
func (e *itEnv) enrRestrictions(t *testing.T, org tenancy.OrganizationID, user tenancy.UserID) (unlifted, byInvitation int) {
	t.Helper()
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) FILTER (WHERE lifted_at IS NULL),
  count(*) FILTER (WHERE lifted_via = 'invitation') FROM app.enrollment_restrictions
  WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid`, org.String(), user.String()).Scan(&unlifted, &byInvitation); err != nil {
		t.Fatalf("restrictions: %v", err)
	}
	return unlifted, byInvitation
}

// enrDenials counts user's access.denied records of enrollment.admit without an Organization with reason.
func (e *itEnv) enrDenials(t *testing.T, user tenancy.UserID, reason string) int {
	t.Helper()
	return e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND organization_id IS NULL
  AND actor_user_id = $1::text::uuid AND operation = 'enrollment.admit' AND reason = $2`, user.String(), reason)
}

// enrDepart grants user org_member in org (fixture) and removes it through G019's RemoveMember as owner, so 00016's
// departure trigger records a departed restriction stamped now.
func (e *itEnv) enrDepart(t *testing.T, owner *itUser, org tenancy.OrganizationID, user tenancy.UserID) {
	t.Helper()
	e.grantOrg(t, org, &itUser{id: user}, tenancy.OrgMember)
	ctx, tc := e.as(t, owner, org)
	if err := e.svc.RemoveMember(ctx, tc, memOrgScope, user); err != nil {
		t.Fatalf("remove member: %v", err)
	}
}

// enrSetMode sets org's enrollment mode as fixture setup (attested superuser, triggers disabled for the statement as
// memExpire; the owner's guarded policy write is a later slice).
func (e *itEnv) enrSetMode(t *testing.T, org tenancy.OrganizationID, mode string) {
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
	if _, err := tt.Exec(ctx, `UPDATE app.organizations SET enrollment_mode = $2 WHERE id = $1::text::uuid`, org.String(), mode); err != nil {
		t.Fatalf("set mode: %v", err)
	}
	if err := tt.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

// enrWorld is one Organization with a bound owner.
func (e *itEnv) enrWorld(t *testing.T, name string) (*itUser, tenancy.OrganizationID) {
	t.Helper()
	o := e.newUser(t)
	a := e.org(t, name)
	e.grantOrg(t, a, o, tenancy.OrgOwner)
	e.bind(t, o, a) // G016 A2: the issuer's credential for A, setup only
	return o, a
}

// enr7FirstTimeInvitation (ENR-7): a first-time principal (mapped, no session) with an owner-issued org_member
// invitation and identity-attested evidence is admitted in one transaction: the invitation is consumed by and intended
// for the user, exactly one org_member binding exists, the credential's admitted_at/admitted_binding_id name it, the
// invitation.consumed + membership.granted pair is recorded and no access.denied is.
func enr7FirstTimeInvitation(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, a := e.enrWorld(t, "ENR7 A")
	sub, user := e.memPrincipal(t, "N")
	addr := memAddr("enr7")
	iss := e.memIssueOrg(t, o, a, addr)
	h := e.enrHandle(t, user, a)
	out, err := e.svc.AdmitInvitationAtCallback(ctx, e.memAttest(t, sub, addr), h, iss.Invitation.ID, iss.Token)
	if err != nil || out.Outcome != string(tenancy.AdmitAdmitted) || zero(out.BindingID) {
		t.Fatalf("admission: %v, %+v", err, out)
	}
	by, intended, pending := e.memInvitation(t, iss.Invitation.ID)
	if by != user.String() || intended != user.String() || pending {
		t.Errorf("invitation: consumed by %q, intended %q, pending %v", by, intended, pending)
	}
	if n := e.memOrgBindings(t, a, user); n != 1 {
		t.Errorf("bindings: %d, want 1", n)
	}
	if role := e.memRole(t, a.String(), user); role != string(tenancy.OrgMember) {
		t.Errorf("role: %s", role)
	}
	if ok, b := e.enrAdmitted(t, h); !ok || b != out.BindingID.String() {
		t.Errorf("credential proof: admitted %v, binding %q, want %s", ok, b, out.BindingID)
	}
	if ev := e.memConsumeEvents(t, a, iss.Invitation.ID, out.BindingID); !slices.Equal(ev, []string{"invitation.consumed", "membership.granted"}) {
		t.Errorf("audit: %v", ev)
	}
	if n := e.memDenials(t, user); n != 0 {
		t.Errorf("access.denied: %d, want 0", n)
	}
}

// enr23StaleInvitation (ENR-23 stale invitation, O-ENR-8 departure-first order): an invitation issued before the user's
// departure (grant, then G019 removal: a departed restriction newer than the invitation) cannot lift it: the lift RAISE
// rolls the claim back (enrollment_restricted; invitation still pending, no binding, restriction unlifted, credential not
// admitted) and one access.denied without an Organization (enrollment_restricted, enrollment.admit) is recorded. A fresh
// re-invite, newer than the restriction, is admitted and lifts it (lifted_via invitation).
func enr23StaleInvitation(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, a := e.enrWorld(t, "ENR23 A")
	sub, user := e.memPrincipal(t, "S")
	addr := memAddr("enr23")
	stale := e.memIssueOrg(t, o, a, addr)
	e.enrDepart(t, o, a, user)
	if unlifted, _ := e.enrRestrictions(t, a, user); unlifted != 1 {
		t.Fatalf("departure restrictions: %d, want 1", unlifted)
	}
	h := e.enrHandle(t, user, a)
	ev := e.memAttest(t, sub, addr)
	_, err := e.svc.AdmitInvitationAtCallback(ctx, ev, h, stale.Invitation.ID, stale.Token)
	memWantCode(t, "stale invitation", err, CodeConflict, fieldEnrollmentRestricted)
	if _, _, pending := e.memInvitation(t, stale.Invitation.ID); !pending {
		t.Error("stale invitation consumed: the claim was not rolled back")
	}
	if consumed, _, b := e.memInvitationRow(t, stale.Invitation.ID); consumed || b != "" {
		t.Errorf("stale invitation row: consumed %v, binding %q", consumed, b)
	}
	if n := e.memOrgBindings(t, a, user); n != 0 {
		t.Errorf("bindings: %d, want 0", n)
	}
	if unlifted, lifted := e.enrRestrictions(t, a, user); unlifted != 1 || lifted != 0 {
		t.Errorf("restrictions after refusal: unlifted %d, lifted %d", unlifted, lifted)
	}
	if ok, _ := e.enrAdmitted(t, h); ok {
		t.Error("credential admitted by a refused admission")
	}
	if n, all := e.enrDenials(t, user, "enrollment_restricted"), e.memDenials(t, user); n != 1 || all != 1 {
		t.Errorf("denials: enrollment_restricted %d, all %d, want 1 and 1", n, all)
	}

	fresh := e.memIssueOrg(t, o, a, addr)
	out, err := e.svc.AdmitInvitationAtCallback(ctx, ev, h, fresh.Invitation.ID, fresh.Token)
	if err != nil || out.Outcome != string(tenancy.AdmitAdmitted) {
		t.Fatalf("fresh invitation: %v, %+v", err, out)
	}
	if n := e.memOrgBindings(t, a, user); n != 1 {
		t.Errorf("bindings after re-invite: %d, want 1", n)
	}
	if unlifted, lifted := e.enrRestrictions(t, a, user); unlifted != 0 || lifted != 1 {
		t.Errorf("restrictions after re-invite: unlifted %d, lifted by invitation %d", unlifted, lifted)
	}
	if _, _, pending := e.memInvitation(t, stale.Invitation.ID); !pending {
		t.Error("the stale invitation was consumed by the re-invite admission")
	}
	if ok, b := e.enrAdmitted(t, h); !ok || b != out.BindingID.String() {
		t.Errorf("credential proof: admitted %v, binding %q", ok, b)
	}
}

// enr8ReplayAndWrongToken (ENR-8): a wrong token is the closed not_found with nothing written (invitation pending, no
// binding, credential not admitted) and one access.denied without an Organization (target_not_found); evidence of
// another user is refused before any statement; after a success, a replay of the same admission admits nothing again
// (already_member, still one binding and one audit pair).
func enr8ReplayAndWrongToken(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, a := e.enrWorld(t, "ENR8 A")
	sub, user := e.memPrincipal(t, "W")
	addr := memAddr("enr8")
	iss := e.memIssueOrg(t, o, a, addr)
	h := e.enrHandle(t, user, a)
	ev := e.memAttest(t, sub, addr)

	wrong := []byte(iss.Token)
	if wrong[0] == 'A' {
		wrong[0] = 'B'
	} else {
		wrong[0] = 'A'
	}
	_, err := e.svc.AdmitInvitationAtCallback(ctx, ev, h, iss.Invitation.ID, string(wrong))
	memWantCode(t, "wrong token", err, CodeNotFound, "")
	if _, _, pending := e.memInvitation(t, iss.Invitation.ID); !pending {
		t.Error("wrong token consumed the invitation")
	}
	if n := e.memOrgBindings(t, a, user); n != 0 {
		t.Errorf("wrong token bindings: %d", n)
	}
	if ok, _ := e.enrAdmitted(t, h); ok {
		t.Error("wrong token admitted the credential")
	}
	if n := e.enrDenials(t, user, "target_not_found"); n != 1 {
		t.Errorf("wrong token denials: %d, want 1", n)
	}

	// Evidence of another canonical user is refused before any statement (no denial recorded for either user).
	osub, other := e.memPrincipal(t, "Y")
	oaddr := memAddr("enr8other")
	_, err = e.svc.AdmitInvitationAtCallback(ctx, e.memAttest(t, osub, oaddr), h, iss.Invitation.ID, iss.Token)
	memWantCode(t, "foreign evidence", err, CodeInvalidInput, fieldEvidence)
	if _, _, pending := e.memInvitation(t, iss.Invitation.ID); !pending {
		t.Error("foreign evidence consumed the invitation")
	}
	if n, m := e.memDenials(t, user), e.memDenials(t, other); n != 1 || m != 0 {
		t.Errorf("foreign evidence denials: user %d, other %d", n, m)
	}

	out, err := e.svc.AdmitInvitationAtCallback(ctx, ev, h, iss.Invitation.ID, iss.Token)
	if err != nil || out.Outcome != string(tenancy.AdmitAdmitted) {
		t.Fatalf("admission: %v, %+v", err, out)
	}
	again, err := e.svc.AdmitInvitationAtCallback(ctx, ev, h, iss.Invitation.ID, iss.Token)
	if err != nil || again.Outcome != string(tenancy.AdmitAlreadyMember) {
		t.Errorf("replay: %v, %+v, want already_member without a write", err, again)
	}
	if n := e.memOrgBindings(t, a, user); n != 1 {
		t.Errorf("replay bindings: %d, want 1", n)
	}
	if ev := e.memConsumeEvents(t, a, iss.Invitation.ID, out.BindingID); !slices.Equal(ev, []string{"invitation.consumed", "membership.granted"}) {
		t.Errorf("replay audit: %v", ev)
	}
}

// enrOpenRestrictedAndClosed (open kind): an invite_only Organization refuses open admission (enrollment_closed, zero
// writes, one access.denied enrollment_closed); once the Organization is open the user is admitted (one binding, the
// credential's proof names it); a user with an unlifted departed restriction is refused (enrollment_restricted, zero
// writes, the restriction stays, one access.denied enrollment_restricted).
func enrOpenRestrictedAndClosed(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, a := e.enrWorld(t, "ENROPEN A")
	_, user := e.memPrincipal(t, "O")
	h := e.enrHandle(t, user, a)

	_, err := e.svc.AdmitOpen(ctx, h)
	memWantCode(t, "invite_only", err, CodeConflict, fieldEnrollmentClosed)
	if n := e.memOrgBindings(t, a, user); n != 0 {
		t.Errorf("closed bindings: %d", n)
	}
	if ok, _ := e.enrAdmitted(t, h); ok {
		t.Error("closed admission admitted the credential")
	}
	if n, all := e.enrDenials(t, user, "enrollment_closed"), e.memDenials(t, user); n != 1 || all != 1 {
		t.Errorf("closed denials: enrollment_closed %d, all %d", n, all)
	}

	e.enrSetMode(t, a, "open")
	out, err := e.svc.AdmitOpen(ctx, h)
	if err != nil || out.Outcome != string(tenancy.AdmitAdmitted) || zero(out.BindingID) {
		t.Fatalf("open admission: %v, %+v", err, out)
	}
	if n := e.memOrgBindings(t, a, user); n != 1 {
		t.Errorf("open bindings: %d, want 1", n)
	}
	if ok, b := e.enrAdmitted(t, h); !ok || b != out.BindingID.String() {
		t.Errorf("open credential proof: admitted %v, binding %q", ok, b)
	}

	_, departed := e.memPrincipal(t, "R")
	e.enrDepart(t, o, a, departed)
	dh := e.enrHandle(t, departed, a)
	_, err = e.svc.AdmitOpen(ctx, dh)
	memWantCode(t, "restricted", err, CodeConflict, fieldEnrollmentRestricted)
	if n := e.memOrgBindings(t, a, departed); n != 0 {
		t.Errorf("restricted bindings: %d", n)
	}
	if unlifted, _ := e.enrRestrictions(t, a, departed); unlifted != 1 {
		t.Errorf("restricted: unlifted %d, want 1", unlifted)
	}
	if ok, _ := e.enrAdmitted(t, dh); ok {
		t.Error("restricted admission admitted the credential")
	}
	if n, all := e.enrDenials(t, departed, "enrollment_restricted"), e.memDenials(t, departed); n != 1 || all != 1 {
		t.Errorf("restricted denials: enrollment_restricted %d, all %d", n, all)
	}

	// ENR-22g (r3 correction 1): a subject whose only membership of A is the Organization binding (no Group binding)
	// leaves through G019's LeaveScope as itself: the call commits, the binding is gone and the departure trigger wrote
	// exactly one unlifted, trigger-recorded departed row for (A, subject).
	lv := e.newUser(t)
	e.grantOrg(t, a, lv, tenancy.OrgMember)
	e.bind(t, lv, a)
	if n := e.memCount(t, `SELECT count(*) FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid`,
		a.String(), lv.id.String()); n != 1 {
		t.Fatalf("leaver bindings of A: %d, want only the Organization binding", n)
	}
	lctx, ltc := e.as(t, lv, a)
	if err := e.svc.LeaveScope(lctx, ltc, memOrgScope); err != nil {
		t.Fatalf("self-leave of the last Organization membership: %v", err)
	}
	if n := e.memOrgBindings(t, a, lv.id); n != 0 {
		t.Errorf("leaver bindings after LeaveScope: %d, want 0", n)
	}
	var rows string
	if err := e.setup.QueryRow(ctx, `SELECT coalesce(string_agg(format('%s|%s|%s', kind, recorded_via, lifted_at IS NULL), ','), '')
  FROM app.enrollment_restrictions WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid`, a.String(), lv.id.String()).Scan(&rows); err != nil ||
		rows != "departed|trigger|t" {
		t.Errorf("self-leave ledger rows: %q %v, want exactly one unlifted trigger-recorded departed row", rows, err)
	}
}

// enrFaultDB is the runtime pool with one injected statement failure: the enrollment_admit call fails as a connection
// error (errHookInjected) after the claim in the same transaction already ran (test-only seam; no production change).
type enrFaultDB struct{ pool *pgxpool.Pool }

func (d enrFaultDB) BeginTx(ctx context.Context, o pgx.TxOptions) (pgx.Tx, error) {
	tx, err := d.pool.BeginTx(ctx, o)
	if err != nil {
		return nil, err
	}
	return enrFaultTx{Tx: tx}, nil
}

type enrFaultTx struct{ pgx.Tx }

func (t enrFaultTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if strings.Contains(sql, "app.enrollment_admit(") {
		return errRow{errHookInjected}
	}
	return t.Tx.QueryRow(ctx, sql, args...)
}

// enr11FaultRollsBackClaim (ENR-11 T2 abort): a failure of the enrollment_admit statement after the claim rolls the
// whole transaction back: unavailable, the invitation still pending, no binding, no invitation.consumed or
// membership.granted, the credential not admitted and no access.denied (a store failure is not a denial).
func enr11FaultRollsBackClaim(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, a := e.enrWorld(t, "ENR11 A")
	sub, user := e.memPrincipal(t, "F")
	addr := memAddr("enr11")
	iss := e.memIssueOrg(t, o, a, addr)
	h := e.enrHandle(t, user, a)
	svc := NewService(enrFaultDB{pool: e.pool}, e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil))
	_, err := svc.AdmitInvitationAtCallback(ctx, e.memAttest(t, sub, addr), h, iss.Invitation.ID, iss.Token)
	memWantCode(t, "admit fault", err, CodeUnavailable, "")
	if _, _, pending := e.memInvitation(t, iss.Invitation.ID); !pending {
		t.Error("the claim survived the admission fault")
	}
	if n := e.memOrgBindings(t, a, user); n != 0 {
		t.Errorf("fault bindings: %d", n)
	}
	if n := e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid
  AND action IN ('invitation.consumed', 'membership.granted') AND (resource_id::text = $2 OR subject_user_id = $3::text::uuid)`,
		a.String(), iss.Invitation.ID.String(), user.String()); n != 0 {
		t.Errorf("fault audit pair: %d rows", n)
	}
	if ok, _ := e.enrAdmitted(t, h); ok {
		t.Error("fault admitted the credential")
	}
	if n := e.memDenials(t, user); n != 0 {
		t.Errorf("fault denials: %d, want 0", n)
	}
	// The same admission through the unfaulted runtime pool then succeeds (nothing was left half-written).
	out, err := e.svc.AdmitInvitationAtCallback(ctx, e.memAttest(t, sub, addr), h, iss.Invitation.ID, iss.Token)
	if err != nil || out.Outcome != string(tenancy.AdmitAdmitted) {
		t.Errorf("retry after fault: %v, %+v", err, out)
	}
}

// --- G058 round 1a-5 rows ---------------------------------------------------------------------------------------------

// enrHostSpy is a HostResolver answering a configured resolution; each call first reads app.user_id on the querier it
// was handed, so a row can prove the guard ran on the admission transaction itself (the IDENTITY shape of the user).
type enrHostSpy struct {
	mu    sync.Mutex
	res   identity.HostResolution
	err   error
	users []string
}

func (s *enrHostSpy) set(res identity.HostResolution, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.res, s.err, s.users = res, err, nil
}

func (s *enrHostSpy) seen() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.users)
}

func (s *enrHostSpy) ResolveOrigin(ctx context.Context, q identity.Querier, _ string) (identity.HostResolution, error) {
	var u string
	if err := q.QueryRow(ctx, `SELECT coalesce(current_setting('app.user_id', true), '')`).Scan(&u); err != nil {
		return identity.HostResolution{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users = append(s.users, u)
	return s.res, s.err
}

// enrAdmitStatements reports whether the hook captured the claim or the admission body statement.
func (e *itEnv) enrAdmitStatements() bool {
	for _, s := range e.hook.capturedStmts() {
		if strings.Contains(s.sql, "app.invitation_claim_organization(") || strings.Contains(s.sql, "app.enrollment_admit(") {
			return true
		}
	}
	return false
}

func (e *itEnv) enrCapture() {
	e.hook.reset()
	e.hook.mu.Lock()
	e.hook.capture = true
	e.hook.mu.Unlock()
}

// enrHostGuard (ENR-HOST; Codex round-2 host decision): the admission's host guard runs on the admission transaction
// (the resolver's querier carries the user's IDENTITY context) before any claim or admission statement. An origin that
// routes nowhere or to another Organization is host_changed: the invitation stays pending, no binding, the credential is
// not admitted, no claim/admit statement ran, and one access.denied (host_changed, no Organization) is recorded per
// refusal. A resolver failure is unavailable with no denial; a routing origin admits.
func enrHostGuard(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, a := e.enrWorld(t, "ENRHOST A")
	other := e.org(t, "ENRHOST B")
	sub, user := e.memPrincipal(t, "H")
	addr := memAddr("enrhost")
	iss := e.memIssueOrg(t, o, a, addr)
	ev := e.memAttest(t, sub, addr)
	spy := &enrHostSpy{}
	sid := identity.NewService(identity.NewStore(e.pool), identity.Config{Hosts: spy, Now: e.clk.Now})
	svc := NewService(e.hook, sid, observability.NewLogger(e.logs, slog.LevelInfo, nil))
	refusals := 0
	for _, c := range []struct {
		name string
		res  identity.HostResolution
	}{
		{"non-routing", identity.HostResolution{}},
		{"reassigned", identity.HostResolution{Kind: identity.HostPlatform, OrganizationID: other.String(), Routing: true}},
	} {
		h := e.enrHandle(t, user, a)
		spy.set(c.res, nil)
		e.enrCapture()
		_, err := svc.AdmitInvitationAtCallback(ctx, ev, h, iss.Invitation.ID, iss.Token)
		memWantCode(t, c.name, err, CodeConflict, fieldHostChanged)
		refusals++
		if got := spy.seen(); len(got) != 1 || got[0] != user.String() {
			t.Errorf("%s: resolver contexts %v, want one call on the admission transaction of %s", c.name, got, user)
		}
		if e.enrAdmitStatements() {
			t.Errorf("%s: a claim or admission statement ran after the host refusal", c.name)
		}
		if _, _, pending := e.memInvitation(t, iss.Invitation.ID); !pending {
			t.Errorf("%s: invitation consumed", c.name)
		}
		if n := e.memOrgBindings(t, a, user); n != 0 {
			t.Errorf("%s: bindings %d", c.name, n)
		}
		if ok, _ := e.enrAdmitted(t, h); ok {
			t.Errorf("%s: credential admitted", c.name)
		}
		if n, all := e.enrDenials(t, user, "host_changed"), e.memDenials(t, user); n != refusals || all != refusals {
			t.Errorf("%s: denials host_changed %d, all %d, want %d", c.name, n, all, refusals)
		}
	}
	h := e.enrHandle(t, user, a)
	spy.set(identity.HostResolution{}, errors.New("resolver unavailable"))
	e.enrCapture()
	_, err := svc.AdmitInvitationAtCallback(ctx, ev, h, iss.Invitation.ID, iss.Token)
	memWantCode(t, "resolver failure", err, CodeUnavailable, "")
	if e.enrAdmitStatements() || e.memDenials(t, user) != refusals || e.memOrgBindings(t, a, user) != 0 {
		t.Errorf("resolver failure: statements %v, denials %d, bindings %d", e.enrAdmitStatements(), e.memDenials(t, user), e.memOrgBindings(t, a, user))
	}
	spy.set(identity.HostResolution{Kind: identity.HostPlatform, OrganizationID: a.String(), Routing: true}, nil)
	e.hook.reset()
	out, err := svc.AdmitInvitationAtCallback(ctx, ev, h, iss.Invitation.ID, iss.Token)
	if err != nil || out.Outcome != string(tenancy.AdmitAdmitted) || e.memOrgBindings(t, a, user) != 1 {
		t.Errorf("routing origin: %v, %+v, bindings %d", err, out, e.memOrgBindings(t, a, user))
	}
	if got := spy.seen(); len(got) != 1 || got[0] != user.String() {
		t.Errorf("routing origin: resolver contexts %v", got)
	}
}

// enrAdmitDenialsWithOrg counts user's enrollment.admit access.denied records that carry an Organization (always 0).
func (e *itEnv) enrAdmitDenialsWithOrg(t *testing.T, user tenancy.UserID) int {
	t.Helper()
	return e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND operation = 'enrollment.admit'
  AND actor_user_id = $1::text::uuid AND (organization_id IS NOT NULL OR group_id IS NOT NULL OR resource_id IS NOT NULL)`, user.String())
}

// enrResolves reports whether the runtime role's bootstrap reader resolves the credential digest now.
func (e *itEnv) enrResolves(t *testing.T, digest []byte) bool {
	t.Helper()
	var n int
	if err := e.pool.QueryRow(context.Background(), `SELECT count(*)::int FROM app.auth_resolve_enrollment($1, now())`, digest).Scan(&n); err != nil {
		t.Fatalf("auth_resolve_enrollment: %v", err)
	}
	return n == 1
}

// enrAuditRefusals (ENR-AUD, CT-ENR-1_audit runtime half): through the composed Admitter (AdmitAtCallback) each known
// refusal is a status with a nil error and exactly one access.denied enrollment.admit without an Organization, Group or
// resource: a revoked credential on an open Organization is enrollment_expired (session_invalid); a disabled user
// (DisableUser, then a fresh credential) is recipient_ineligible and the credential never bootstraps; an invitation whose
// issuer lost authority (G019 removal of the issuing administrator) is not_found (target_not_found), still pending.
func enrAuditRefusals(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, a := e.enrWorld(t, "ENRAUD A")
	e.enrSetMode(t, a, "open")
	open := identity.AdmitSelector{Kind: identity.AdmitKindOpen}

	_, ru := e.memPrincipal(t, "V")
	rh := e.enrHandle(t, ru, a)
	if _, err := e.setup.Exec(ctx, `UPDATE app.enrollment_sessions SET revoked_at = now(), revoke_reason = 'logout'
  WHERE credential_hash = $1`, rh.Digest[:]); err != nil {
		t.Fatalf("revoke fixture: %v", err)
	}
	out, err := e.svc.AdmitAtCallback(ctx, rh, open)
	if err != nil || out.Status != identity.AdmitEnrollmentExpired {
		t.Errorf("revoked credential: %+v %v", out, err)
	}
	if n, all := e.enrDenials(t, ru, "session_invalid"), e.memDenials(t, ru); n != 1 || all != 1 || e.memOrgBindings(t, a, ru) != 0 {
		t.Errorf("revoked credential: session_invalid %d, all %d, bindings %d", n, all, e.memOrgBindings(t, a, ru))
	}

	_, du := e.memPrincipal(t, "D")
	if err := e.id.DisableUser(ctx, du.String()); err != nil {
		t.Fatalf("disable: %v", err)
	}
	dh := e.enrHandle(t, du, a)
	out, err = e.svc.AdmitAtCallback(ctx, dh, open)
	if err != nil || out.Status != identity.AdmitRecipientIneligible {
		t.Errorf("disabled user: %+v %v", out, err)
	}
	if n, all := e.enrDenials(t, du, "recipient_ineligible"), e.memDenials(t, du); n != 1 || all != 1 {
		t.Errorf("disabled user: recipient_ineligible %d, all %d", n, all)
	}
	if ok, _ := e.enrAdmitted(t, dh); ok || e.memOrgBindings(t, a, du) != 0 || e.enrResolves(t, dh.Digest[:]) {
		t.Errorf("disabled user: admitted %v, bindings %d, bootstrap resolves %v", ok, e.memOrgBindings(t, a, du), e.enrResolves(t, dh.Digest[:]))
	}

	adm := e.newUser(t)
	e.grantOrg(t, a, adm, tenancy.OrgAdmin)
	e.bind(t, adm, a)
	sub, iu := e.memPrincipal(t, "I")
	addr := memAddr("enraud")
	iss := e.memIssueOrg(t, adm, a, addr)
	octx, otc := e.as(t, o, a)
	if err := e.svc.RemoveMember(octx, otc, memOrgScope, adm.id); err != nil {
		t.Fatalf("remove issuer: %v", err)
	}
	ih := e.enrHandle(t, iu, a)
	va := e.memAttest(t, sub, addr)
	ref := identity.InvitationRef(iss.Invitation.ID)
	out, err = e.svc.AdmitAtCallback(ctx, ih, identity.AdmitSelector{Kind: identity.AdmitKindInvitation, Invitation: &ref,
		InvitationToken: iss.Token, Address: &va})
	if err != nil || out.Status != identity.AdmitNotFound {
		t.Errorf("issuer authority lost: %+v %v", out, err)
	}
	if _, _, pending := e.memInvitation(t, iss.Invitation.ID); !pending || e.memOrgBindings(t, a, iu) != 0 {
		t.Errorf("issuer authority lost: pending %v, bindings %d", pending, e.memOrgBindings(t, a, iu))
	}
	if n, all := e.enrDenials(t, iu, "target_not_found"), e.memDenials(t, iu); n != 1 || all != 1 {
		t.Errorf("issuer authority lost: target_not_found %d, all %d", n, all)
	}
	for _, u := range []tenancy.UserID{ru, du, iu} {
		if n := e.enrAdmitDenialsWithOrg(t, u); n != 0 {
			t.Errorf("%s: %d enrollment.admit denials carry an Organization, Group or resource", u, n)
		}
	}
}

// enrLostCommit (ENR-LOST): an open admission whose COMMIT succeeds but whose reply is lost is unavailable after one
// transaction (no retry); the committed effect is exactly one binding named by the credential's proof, one
// membership.granted and no denial; the same admission again is already_member with that binding and writes nothing.
func enrLostCommit(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	_, a := e.enrWorld(t, "ENRLOST A")
	e.enrSetMode(t, a, "open")
	_, user := e.memPrincipal(t, "L")
	h := e.enrHandle(t, user, a)
	granted := func() int {
		return e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'membership.granted' AND subject_user_id = $1::text::uuid`, user.String())
	}
	e.hook.reset()
	e.hook.setCommit(commitLostReply)
	_, err := e.svc.AdmitOpen(ctx, h)
	memWantCode(t, "lost commit reply", err, CodeUnavailable, "")
	if begins, _ := e.hook.counts(); begins != 1 {
		t.Errorf("lost commit reply: %d transactions, want 1", begins)
	}
	e.hook.reset()
	ok, proof := e.enrAdmitted(t, h)
	var bid string
	if err := e.setup.QueryRow(ctx, `SELECT coalesce(string_agg(id::text, ','), '') FROM app.role_bindings WHERE organization_id = $1::text::uuid
  AND subject_user_id = $2::text::uuid AND scope_type = 'organization' AND source = 'membership'`, a.String(), user.String()).Scan(&bid); err != nil {
		t.Fatal(err)
	}
	if !ok || proof == "" || proof != bid || granted() != 1 || e.memDenials(t, user) != 0 {
		t.Errorf("committed effect: admitted %v, proof %q, bindings %q, granted %d, denials %d", ok, proof, bid, granted(), e.memDenials(t, user))
	}
	again, err := e.svc.AdmitOpen(ctx, h)
	if err != nil || again.Outcome != string(tenancy.AdmitAlreadyMember) || again.BindingID.String() != bid {
		t.Errorf("reconcile: %v, %+v, want already_member %s", err, again, bid)
	}
	if e.memOrgBindings(t, a, user) != 1 || granted() != 1 || e.memDenials(t, user) != 0 {
		t.Errorf("reconcile wrote: bindings %d, granted %d, denials %d", e.memOrgBindings(t, a, user), granted(), e.memDenials(t, user))
	}
}

// enrRemovedBinding (ENR-REMOVED; Codex round-3 decision 3): an admitted credential answers already_member only while
// its admitted binding is the user's current Organization membership. After a G019 removal the same credential is
// enrollment_restricted (no write, the proof still names the old binding, one denial); a fresh invitation admits the
// user again (restriction lifted, new binding); the old credential stays enrollment_restricted (its binding was
// replaced) while the new credential answers already_member with the new binding.
func enrRemovedBinding(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, a := e.enrWorld(t, "ENRRM A")
	e.enrSetMode(t, a, "open")
	sub, user := e.memPrincipal(t, "M")
	h := e.enrHandle(t, user, a)
	first, err := e.svc.AdmitOpen(ctx, h)
	if err != nil || first.Outcome != string(tenancy.AdmitAdmitted) {
		t.Fatalf("open admission: %v, %+v", err, first)
	}
	octx, otc := e.as(t, o, a)
	if err := e.svc.RemoveMember(octx, otc, memOrgScope, user); err != nil {
		t.Fatalf("remove: %v", err)
	}
	_, err = e.svc.AdmitOpen(ctx, h)
	memWantCode(t, "removed binding", err, CodeConflict, fieldEnrollmentRestricted)
	var member bool
	if err := e.setup.QueryRow(ctx, `SELECT app.is_organization_member($1::text::uuid, $2::text::uuid)`, user.String(), a.String()).Scan(&member); err != nil {
		t.Fatal(err)
	}
	if _, proof := e.enrAdmitted(t, h); member || e.memOrgBindings(t, a, user) != 0 || proof != first.BindingID.String() {
		t.Errorf("removed binding: member %v, bindings %d, proof %q (want %s)", member, e.memOrgBindings(t, a, user), proof, first.BindingID)
	}
	if n := e.enrDenials(t, user, "enrollment_restricted"); n != 1 {
		t.Errorf("removed binding denials: %d, want 1", n)
	}
	addr := memAddr("enrrm")
	iss := e.memIssueOrg(t, o, a, addr)
	h2 := e.enrHandle(t, user, a)
	second, err := e.svc.AdmitInvitationAtCallback(ctx, e.memAttest(t, sub, addr), h2, iss.Invitation.ID, iss.Token)
	if err != nil || second.Outcome != string(tenancy.AdmitAdmitted) || second.BindingID == first.BindingID {
		t.Fatalf("re-invite: %v, %+v", err, second)
	}
	if unlifted, lifted := e.enrRestrictions(t, a, user); unlifted != 0 || lifted != 1 {
		t.Errorf("re-invite restrictions: unlifted %d, lifted %d", unlifted, lifted)
	}
	_, err = e.svc.AdmitOpen(ctx, h)
	memWantCode(t, "replaced binding", err, CodeConflict, fieldEnrollmentRestricted)
	again, err := e.svc.AdmitOpen(ctx, h2)
	if err != nil || again.Outcome != string(tenancy.AdmitAlreadyMember) || again.BindingID != second.BindingID {
		t.Errorf("new credential: %v, %+v, want already_member %s", err, again, second.BindingID)
	}
	if _, proof := e.enrAdmitted(t, h); e.memOrgBindings(t, a, user) != 1 || proof != first.BindingID.String() {
		t.Errorf("after replacement: bindings %d, old proof %q", e.memOrgBindings(t, a, user), proof)
	}
	if n := e.enrDenials(t, user, "enrollment_restricted"); n != 2 {
		t.Errorf("restricted denials: %d, want 2", n)
	}
}

// enrCallback runs one enrollment-intent login on org's tenant origin (stored invitation loginInv) and its callback for
// provider principal sub with a verified address (forwarded invitation cbInv with token; live forwarded credential).
func (e *itEnv) enrCallback(t *testing.T, org tenancy.OrganizationID, sub, addr string, loginInv, cbInv *identity.InvitationRef,
	token, live string) (identity.CallbackOutput, error) {
	t.Helper()
	return e.enrCallbackWith(t, org, sub, addr, loginInv, cbInv, token, live, false)
}

// enrCallbackWith is enrCallback with the login's request_on_arrival flag.
func (e *itEnv) enrCallbackWith(t *testing.T, org tenancy.OrganizationID, sub, addr string, loginInv, cbInv *identity.InvitationRef,
	token, live string, request bool) (identity.CallbackOutput, error) {
	t.Helper()
	ctx := context.Background()
	origin := itOriginOf(org)
	out, err := e.id.Login(ctx, identity.LoginInput{Origin: origin, EnrollmentIntent: true, Invitation: loginInv, Request: request})
	if err != nil {
		t.Fatalf("enrollment login: %v", err)
	}
	au, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	code := e.fake.IssueCode(au.Query().Get("code_challenge"), identitytest.Grant{UserID: sub, FirstName: "First", LastName: "Time",
		Email: addr, SID: fmt.Sprintf("session_01ENI%08d", itSeq.Add(1))})
	return e.id.Callback(ctx, identity.CallbackInput{Origin: origin, RawQuery: "code=" + url.QueryEscape(code) + "&state=" +
		url.QueryEscape(au.Query().Get("state")), Binding: out.Binding, Credential: live, Invitation: cbInv, InvitationToken: token})
}

// enrStatus asserts an enrollment-status answer with status and, when withCredential, an enrollment credential (and no
// session); it returns the credential's digest (nil without one).
func enrStatus(t *testing.T, what string, out identity.CallbackOutput, err error, status string, withCredential bool) []byte {
	t.Helper()
	if err != nil || out.Outcome != identity.OutcomeEnrollmentStatus || out.Enrollment == nil || out.Enrollment.Status != status ||
		out.Credential != "" || (out.Enrollment.Credential != "") != withCredential {
		t.Errorf("%s: %+v %v, want enrollment_status %s (credential %v)", what, out, err, status, withCredential)
		return nil
	}
	if !withCredential {
		return nil
	}
	raw, derr := base64.RawURLEncoding.DecodeString(out.Enrollment.Credential)
	if derr != nil {
		t.Fatalf("%s: credential encoding: %v", what, derr)
	}
	sum := sha256.Sum256(raw)
	return sum[:]
}

// enrCallbackIntent (ENR-INTENT; Codex round-4 correction and disabled-subject decision; the T1 Login/callback identity
// rows): identity's callback composed with this Service as the bound Admitter, end to end over the fake provider.
// T1 refusals: an invite_only Organization without an invitation is enrollment_closed (one access.denied
// enrollment.start without an Organization, no auth.sign_in, no principal created); a forwarded invitation other than the
// stored one is invitation_mismatch (one enrollment.start denial attributed to the mapped subject). Stale invitation
// intent on an OPEN Organization (expired, revoked, consumed by another user) is adjudicated by T2's claim and answers
// the claim's closed not_found with the enrollment credential, never an open admission (no binding). Without an
// invitation the open Organization admits (admitted, next sign_in, no credential); that member then signs in the
// ordinary way; a live session of another user is identity_conflict. A disabled existing user gets no membership and no
// session: T2 refuses recipient_ineligible and the T1 credential never bootstraps (auth_resolve_enrollment).
func enrCallbackIntent(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	if err := e.id.BindAdmitter(e.svc); err != nil {
		t.Fatalf("bind admitter: %v", err)
	}
	o, a := e.enrWorld(t, "ENRINT A")
	signIns := func() int { return e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'auth.sign_in'`) }
	starts := func(reason string) int {
		return e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND operation = 'enrollment.start'
  AND reason = $1 AND organization_id IS NULL AND group_id IS NULL AND resource_id IS NULL`, reason)
	}

	csub := fmt.Sprintf("user_01ENC%08d", itSeq.Add(1))
	// The closed denial is unattributed (no mapped subject), so it is counted as a delta over the shared database.
	before, closedBefore := signIns(), starts("enrollment_closed")
	_, err := e.enrCallback(t, a, csub, memAddr("enrc"), nil, nil, "", "")
	if !errors.Is(err, identity.ErrEnrollmentClosed) {
		t.Errorf("closed: %v", err)
	}
	if n := e.memCount(t, `SELECT count(*) FROM app.identity_principals WHERE subject = $1`, csub); n != 0 ||
		starts("enrollment_closed")-closedBefore != 1 || signIns() != before {
		t.Errorf("closed: principals %d, start denials delta %d, auth.sign_in delta %d", n, starts("enrollment_closed")-closedBefore, signIns()-before)
	}

	msub, mu := e.memPrincipal(t, "X")
	refX, refY := identity.InvitationRef{0x58, 1}, identity.InvitationRef{0x58, 2}
	tok := strings.Repeat("A", 43)
	before = signIns()
	_, err = e.enrCallback(t, a, msub, memAddr("enrx"), &refX, &refY, tok, "")
	if !errors.Is(err, identity.ErrInvitationMismatch) {
		t.Errorf("mismatch: %v", err)
	}
	if n := e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND operation = 'enrollment.start'
  AND reason = 'invitation_mismatch' AND actor_user_id = $1::text::uuid AND organization_id IS NULL`, mu.String()); n != 1 || signIns() != before {
		t.Errorf("mismatch: attributed start denials %d, auth.sign_in delta %d", n, signIns()-before)
	}

	e.enrSetMode(t, a, "open")
	stale := func(t *testing.T, name, addr string, sub string, user tenancy.UserID, iss IssuedInvitation) {
		t.Helper()
		ref := identity.InvitationRef(iss.Invitation.ID)
		out, err := e.enrCallback(t, a, sub, addr, &ref, &ref, iss.Token, "")
		digest := enrStatus(t, name, out, err, identity.AdmitNotFound, true)
		if n := e.memOrgBindings(t, a, user); n != 0 {
			t.Errorf("%s: stale invitation intent admitted (%d bindings)", name, n)
		}
		if digest != nil {
			var admitted bool
			if err := e.setup.QueryRow(ctx, `SELECT admitted_at IS NOT NULL FROM app.enrollment_sessions WHERE credential_hash = $1`, digest).Scan(&admitted); err != nil || admitted {
				t.Errorf("%s: credential row %v, admitted %v", name, err, admitted)
			}
		}
		if n, all := e.enrDenials(t, user, "target_not_found"), e.memDenials(t, user); n != 1 || all != 1 {
			t.Errorf("%s: target_not_found %d, all %d", name, n, all)
		}
	}
	{
		sub, user := e.memPrincipal(t, "E")
		addr := memAddr("enrexp")
		iss := e.memIssueOrg(t, o, a, addr)
		e.memExpire(t, iss.Invitation.ID)
		stale(t, "expired invitation, open Organization", addr, sub, user, iss)
	}
	{
		sub, user := e.memPrincipal(t, "R")
		addr := memAddr("enrrev")
		iss := e.memIssueOrg(t, o, a, addr)
		octx, otc := e.as(t, o, a)
		if err := e.svc.RevokeInvitation(octx, otc, iss.Invitation.ID); err != nil {
			t.Fatalf("revoke: %v", err)
		}
		stale(t, "revoked invitation, open Organization", addr, sub, user, iss)
	}
	{
		sub, user := e.memPrincipal(t, "C")
		ysub, yu := e.memPrincipal(t, "Y")
		addr := memAddr("enrcon")
		iss := e.memIssueOrg(t, o, a, addr)
		if out, err := e.svc.AdmitInvitationAtCallback(ctx, e.memAttest(t, ysub, addr), e.enrHandle(t, yu, a), iss.Invitation.ID, iss.Token); err != nil ||
			out.Outcome != string(tenancy.AdmitAdmitted) {
			t.Fatalf("consume by another user: %v, %+v", err, out)
		}
		stale(t, "invitation consumed by another user, open Organization", addr, sub, user, iss)
	}

	psub, pu := e.memPrincipal(t, "P")
	paddr := memAddr("enrpos")
	out, err := e.enrCallback(t, a, psub, paddr, nil, nil, "", "")
	enrStatus(t, "open positive", out, err, identity.AdmitAdmitted, false)
	if out.Enrollment != nil && out.Enrollment.Next != identity.NextSignIn {
		t.Errorf("open positive: next %q", out.Enrollment.Next)
	}
	if n := e.memOrgBindings(t, a, pu); n != 1 {
		t.Errorf("open positive: bindings %d, want 1", n)
	}
	member, err := e.enrCallback(t, a, psub, paddr, nil, nil, "", "")
	if err != nil || member.Outcome != identity.OutcomeSignedIn || member.Credential == "" || member.Enrollment != nil {
		t.Fatalf("member ordinary sign-in: %+v %v", member, err)
	}
	if _, err := e.enrCallback(t, a, fmt.Sprintf("user_01ENK%08d", itSeq.Add(1)), memAddr("enrk"), nil, nil, "", member.Credential); !errors.Is(err, identity.ErrIdentityConflict) {
		t.Errorf("live session of another user: %v", err)
	}

	dsub, du := e.memPrincipal(t, "Q")
	if err := e.id.DisableUser(ctx, du.String()); err != nil {
		t.Fatalf("disable: %v", err)
	}
	out, err = e.enrCallback(t, a, dsub, memAddr("enrdis"), nil, nil, "", "")
	digest := enrStatus(t, "disabled user", out, err, identity.AdmitRecipientIneligible, true)
	if digest != nil && e.enrResolves(t, digest) {
		t.Error("disabled user: the T1 credential bootstraps")
	}
	if n := e.memOrgBindings(t, a, du); n != 0 {
		t.Errorf("disabled user: bindings %d", n)
	}
	if n := e.memCount(t, `SELECT count(*) FROM app.sessions WHERE user_id = $1::text::uuid`, du.String()); n != 0 {
		t.Errorf("disabled user: %d sessions", n)
	}
	if n := e.enrDenials(t, du, "recipient_ineligible"); n != 1 {
		t.Errorf("disabled user: recipient_ineligible denials %d", n)
	}

	// G058 HTTP slice (request_on_arrival; Codex request-c2 disposition 3; Plan r3 §1.3 step 3): on a request_access
	// Organization an enrollment login with request_on_arrival submits the user's own access request in T2 (status pending
	// with the enrollment credential; one pending row and one applicant-attributed enrollment.requested event; never a
	// membership); a repeat login answers the same pending request idempotently; a login without the flag writes no
	// request. The returned credential resolves only at its own origin (EnrollmentCredential), never as a session, and no
	// longer after abandonment. An approval followed by the applicant's own admission is ENR-9 (TestEnrollmentRequests).
	e.enrSetMode(t, a, "request_access")
	requests := func(user tenancy.UserID) int {
		return e.memCount(t, `SELECT count(*) FROM app.enrollment_requests WHERE organization_id = $1::text::uuid
  AND user_id = $2::text::uuid AND status = 'pending'`, a.String(), user.String())
	}
	requested := func(user tenancy.UserID) int {
		return e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'enrollment.requested' AND actor_user_id = $1::text::uuid
  AND organization_id = $2::text::uuid`, user.String(), a.String())
	}
	rsub, ru := e.memPrincipal(t, "RQ")
	raddr := memAddr("enrrq")
	out, err = e.enrCallbackWith(t, a, rsub, raddr, nil, nil, "", "", true)
	digest = enrStatus(t, "request on arrival", out, err, identity.EnrollmentPending, true)
	if requests(ru) != 1 || requested(ru) != 1 || e.memOrgBindings(t, a, ru) != 0 {
		t.Errorf("request on arrival: pending %d, requested events %d, bindings %d", requests(ru), requested(ru), e.memOrgBindings(t, a, ru))
	}
	if digest != nil {
		var admitted bool
		if err := e.setup.QueryRow(ctx, `SELECT admitted_at IS NOT NULL FROM app.enrollment_sessions WHERE credential_hash = $1`, digest).Scan(&admitted); err != nil || admitted {
			t.Errorf("request on arrival: credential row %v, admitted %v", err, admitted)
		}
	}
	if out.Enrollment == nil || out.Enrollment.Credential == "" {
		t.Fatalf("request on arrival: no enrollment credential: %+v", out)
	}
	origin := itOriginOf(a)
	info, err := e.id.EnrollmentCredential(ctx, out.Enrollment.Credential, origin)
	if err != nil || info.Handle.UserID != ru.String() || info.Handle.OrganizationID != a.String() || info.Handle.Origin != origin ||
		len(info.CSRFToken) != 32 || !bytes.Equal(info.Handle.Digest[:], digest) {
		t.Errorf("credential resolution: %+v %v", info.Handle, err)
	}
	if _, err := e.id.EnrollmentCredential(ctx, out.Enrollment.Credential, origin+"0"); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Errorf("credential at another origin: %v", err)
	}
	if _, err := e.id.Session(ctx, out.Enrollment.Credential, origin); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Errorf("enrollment credential as a session: %v", err)
	}
	again, err := e.enrCallbackWith(t, a, rsub, raddr, nil, nil, "", "", true)
	enrStatus(t, "repeat request on arrival", again, err, identity.EnrollmentPending, true)
	if requests(ru) != 1 || requested(ru) != 1 {
		t.Errorf("repeat request on arrival: pending %d, requested events %d", requests(ru), requested(ru))
	}
	if err := e.svc.AbandonEnrollment(ctx, info.Handle); err != nil {
		t.Fatalf("abandon: %v", err)
	}
	if _, err := e.id.EnrollmentCredential(ctx, out.Enrollment.Credential, origin); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Errorf("abandoned credential resolves: %v", err)
	}
	nsub, nu := e.memPrincipal(t, "NR")
	out, err = e.enrCallbackWith(t, a, nsub, memAddr("enrnr"), nil, nil, "", "", false)
	enrStatus(t, "no request on arrival", out, err, identity.EnrollmentPending, true)
	if requests(nu) != 0 || requested(nu) != 0 {
		t.Errorf("no request on arrival: pending %d, requested events %d", requests(nu), requested(nu))
	}
}

// oEnr8ClaimDeparture (O-ENR-8; Plan r3 §1.4 chronology under the Organization lock). It is a TestEnrollmentAdmission
// row (TestAccessOrdering's subtest count is pinned by B-POLICY-ORDERING) using that file's barriers. claim_first: an
// admitted claim followed by the user's removal records a departure newer than the admission and the open path is then
// refused. departure_first is ENR-23 (cited, not repeated). claim_in_flight: the claim paused after its statement while
// the recipient's removal runs: the removal either finds no committed binding (not_found; the admission commits, no
// restriction) or serializes after the admission (binding removed, one departed restriction); never a half state.
// departure_in_flight: an existing member's upgrade invitation (issued before the departure) claimed while the member's
// removal is paused after its DELETE holding the locks: the claim waits (an ungranted lock observed in pg_locks while the
// admission has not returned); after the departure commits the lift refuses (enrollment_restricted, or a class-40 victim
// whose decision reason is aborted) with a full rollback (invitation pending, no binding, not admitted).
func oEnr8ClaimDeparture(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, a := e.enrWorld(t, "OENR8 A")
	e.enrSetMode(t, a, "open")
	var zeroG tenancy.GroupID

	t.Run("claim_first", func(t *testing.T) {
		sub, user := e.memPrincipal(t, "8A")
		addr := memAddr("oenr8a")
		iss := e.memIssueOrg(t, o, a, addr)
		out, err := e.svc.AdmitInvitationAtCallback(ctx, e.memAttest(t, sub, addr), e.enrHandle(t, user, a), iss.Invitation.ID, iss.Token)
		if err != nil || out.Outcome != string(tenancy.AdmitAdmitted) {
			t.Fatalf("claim: %v, %+v", err, out)
		}
		octx, otc := e.as(t, o, a)
		if err := e.svc.RemoveMember(octx, otc, memOrgScope, user); err != nil {
			t.Fatalf("remove: %v", err)
		}
		_, err = e.svc.AdmitOpen(ctx, e.enrHandle(t, user, a))
		memWantCode(t, "open after departure", err, CodeConflict, fieldEnrollmentRestricted)
		if unlifted, _ := e.enrRestrictions(t, a, user); unlifted != 1 || e.memOrgBindings(t, a, user) != 0 {
			t.Errorf("claim_first: unlifted %d, bindings %d", unlifted, e.memOrgBindings(t, a, user))
		}
	})

	t.Run("claim_in_flight", func(t *testing.T) {
		tracer := &pauseTracer{}
		svc := NewService(e.runtimePool(t, tracer), e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil))
		sub, user := e.memPrincipal(t, "8B")
		addr := memAddr("oenr8b")
		iss := e.memIssueOrg(t, o, a, addr)
		h := e.enrHandle(t, user, a)
		ev := e.memAttest(t, sub, addr)
		p := tracer.arm("app.invitation_claim_organization(")
		adm := start(func() error {
			_, err := svc.AdmitInvitationAtCallback(ctx, ev, h, iss.Invitation.ID, iss.Token)
			return err
		})
		p.wait(t, "claim")
		octx, otc := e.as(t, o, a)
		rem := start(func() error { return e.svc.RemoveMember(octx, otc, memOrgScope, user) })
		seam := "removal_blocked"
		select {
		case <-rem.done:
			seam = "removal_done"
		case <-time.After(2 * time.Second):
		}
		close(p.release)
		aerr, rerr := adm.wait(t, "admission"), rem.wait(t, "removal")
		if aerr != nil {
			t.Errorf("claim_in_flight: admission %v", aerr)
		}
		unlifted, _ := e.enrRestrictions(t, a, user)
		n := e.memOrgBindings(t, a, user)
		switch {
		case CodeOf(rerr) == CodeNotFound && n == 1 && unlifted == 0:
		case rerr == nil && n == 0 && unlifted == 1 && seam == "removal_blocked":
		default:
			t.Errorf("claim_in_flight (%s): removal %v, bindings %d, unlifted %d", seam, rerr, n, unlifted)
		}
		ordering(t, map[string]any{"case": "O-ENR-8", "order": "claim_in_flight", "seam": seam, "removal": fmt.Sprint(CodeOf(rerr)), "bindings": n})
	})

	t.Run("departure_in_flight", func(t *testing.T) {
		sub, user := e.memPrincipal(t, "8C")
		e.grantOrg(t, a, &itUser{id: user}, tenancy.OrgMember)
		addr := memAddr("oenr8c")
		octx, otc := e.as(t, o, a)
		iss, err := e.svc.IssueInvitation(octx, otc, memOrgScope, string(tenancy.OrgAdmin), Recipient{Address: &addr}, 0)
		if err != nil {
			t.Fatalf("issue upgrade: %v", err)
		}
		s2, h2 := e.service(e.id)
		dp := newPause()
		h2.on("after", "D", dp.hook)
		rctx, _ := beginAs(t, s2, ctx, o, a)
		rtc := resolve(t, s2, rctx, a, zeroG)
		dep := start(func() error { return s2.RemoveMember(rctx, rtc, memOrgScope, user) })
		dp.wait(t, "departure")
		h := e.enrHandle(t, user, a)
		ev := e.memAttest(t, sub, addr)
		from := len(e.logs.String())
		adm := start(func() error {
			_, err := e.svc.AdmitInvitationAtCallback(ctx, ev, h, iss.Invitation.ID, iss.Token)
			return err
		})
		// The seam is observed in PostgreSQL, not inferred from elapsed time: another backend of this database holds an
		// ungranted lock (pg_locks, the existing waitWaiters observer) while the admission has not returned and the
		// paused departure still holds its locks.
		e.waitWaiters(t, 1)
		seam := "claim_lock_wait"
		select {
		case <-adm.done:
			seam = "claim_done"
		default:
		}
		close(dp.release)
		derr, aerr := dep.wait(t, "departure"), adm.wait(t, "admission")
		if derr != nil {
			t.Errorf("departure_in_flight: departure %v", derr)
		}
		// The admission's own access.decision record (filtered by operation: the concurrent departure writes to the same
		// log, so the last record is not necessarily the admission's).
		recs := e.memOpRecords(t, from, opEnrollmentAdmit)
		reason := ""
		if len(recs) == 1 {
			reason, _ = recs[0]["reason"].(string)
		} else {
			t.Errorf("departure_in_flight: %d enrollment.admit decision records, want 1", len(recs))
		}
		var ae *Error
		restricted := errors.As(aerr, &ae) && ae.Code == CodeConflict && ae.Field == fieldEnrollmentRestricted &&
			reason == string(reasonEnrollmentRestricted)
		// A class-40 victim counts only with the attributed reason aborted (failureReason: SQLSTATE class 40, or a
		// COMMIT the server answered with ROLLBACK); a transport failure (store_unavailable), a cancellation (canceled)
		// or an unknown COMMIT outcome (commit_outcome_unknown) is unavailable too but never satisfies this branch.
		victim := CodeOf(aerr) == CodeUnavailable && reason == string(reasonAborted)
		if seam != "claim_lock_wait" || !(restricted || victim) {
			t.Errorf("departure_in_flight (%s): admission %v, reason %q, want a lock-waiting claim refused enrollment_restricted (or an aborted class-40 victim)", seam, aerr, reason)
		}
		if _, _, pending := e.memInvitation(t, iss.Invitation.ID); !pending || e.memOrgBindings(t, a, user) != 0 {
			t.Errorf("departure_in_flight: pending %v, bindings %d", pending, e.memOrgBindings(t, a, user))
		}
		if ok, _ := e.enrAdmitted(t, h); ok {
			t.Error("departure_in_flight: credential admitted")
		}
		ordering(t, map[string]any{"case": "O-ENR-8", "order": "departure_in_flight", "seam": seam, "admission": fmt.Sprint(CodeOf(aerr)), "reason": reason})
	})
}

// TestEnrollmentRequests is the G058 request-slice entry point (Plan r3 §6.2 ENR-9, ENR-22/23 request paths, the
// request authority/expiry/rollback rows and ENR-10 O-ENR-6/7 plus the revocation-versus-consumption and one_pending
// races; Codex round-8 dispositions and request-c1 rulings). Each subtest builds its own environment; a row selector is
// '^TestEnrollmentRequests$/^ENR-9_'. Applicants are mapped principals without a session; their credentials are
// app.enrollment_sessions fixture rows (enrHandle). Every race runs real concurrent transactions on PostgreSQL: the
// first actor is held open at a pinned statement or before its COMMIT, and the second actor's lock wait is observed in
// pg_locks before the first is released.
func TestEnrollmentRequests(t *testing.T) {
	for _, c := range []struct {
		name string
		run  func(*testing.T)
	}{
		{"ENR-9 request approve admit and reconcile", enr9ApproveAdmit},
		{"ENR-9D denial restricts and a new approval lifts", enr9DenyRelift},
		{"ENR-9R revocation and a newer denial refuse a stale approval", enr9RevokeStale},
		{"ENR-RQA request authority and closed set", enrRequestAuthority},
		{"ENR-RQE expired pending stays pending and unconsumable", enrRequestExpiry},
		{"ENR-RQB decision revocation and request roll back with their audit row", enrRequestRollback},
		{"ENR-RQL lost and aborted COMMIT of a request", enrRequestCommitOutcomes},
		{"O-ENR-RQ concurrent first requests resolve to one pending row", oEnrOnePendingRace},
		{"O-ENR-7 two consumers of one approval", oEnr7TwoConsumers},
		{"O-ENR-RV revocation versus consumption in both orders", oEnrRevokeConsume},
		{"O-ENR-6 approval versus removal in both orders", oEnr6ApprovalRemoval},
	} {
		t.Run(c.name, c.run)
	}
}

// reqWorld is one request_access Organization with a bound owner and a bound org_admin.
func (e *itEnv) reqWorld(t *testing.T, name string) (owner, admin *itUser, org tenancy.OrganizationID) {
	t.Helper()
	owner, org = e.enrWorld(t, name)
	admin = e.newUser(t)
	e.grantOrg(t, org, admin, tenancy.OrgAdmin)
	e.bind(t, admin, org)
	e.enrSetMode(t, org, "request_access")
	return owner, admin, org
}

// reqApplicant maps a first-time applicant and returns it with one fresh credential for org.
func (e *itEnv) reqApplicant(t *testing.T, org tenancy.OrganizationID, tag string) (tenancy.UserID, identity.EnrollmentHandle) {
	t.Helper()
	_, user := e.memPrincipal(t, tag)
	return user, e.enrHandle(t, user, org)
}

// reqRequest submits h's request and requires a pending answer.
func (e *itEnv) reqRequest(t *testing.T, svc *Service, h identity.EnrollmentHandle) EnrollmentRequestView {
	t.Helper()
	v, err := svc.RequestEnrollment(context.Background(), h)
	if err != nil || v.Status != tenancy.RequestPending {
		t.Fatalf("request: %+v %v", v, err)
	}
	return v
}

// reqDecide decides id as u (through e.svc) and requires the decided state.
func (e *itEnv) reqDecide(t *testing.T, u *itUser, org tenancy.OrganizationID, id tenancy.EnrollmentRequestID, decision string) EnrollmentRequestView {
	t.Helper()
	ctx, tc := e.as(t, u, org)
	v, err := e.svc.DecideEnrollmentRequest(ctx, tc, id, decision)
	want := map[string]string{DecisionApprove: tenancy.RequestApproved, DecisionDeny: tenancy.RequestDenied}[decision]
	if err != nil || v.Status != want || v.ID != id {
		t.Fatalf("decide %s: %+v %v", decision, v, err)
	}
	return v
}

// reqRow reads a stored request: status, consumed, the decider and revoker (empty when NULL) and authority_seq.
func (e *itEnv) reqRow(t *testing.T, id tenancy.EnrollmentRequestID) (status string, consumed bool, decider, revoker string, seq int64) {
	t.Helper()
	if err := e.setup.QueryRow(context.Background(), `SELECT status, consumed_at IS NOT NULL, coalesce(decided_by_user_id::text, ''),
  coalesce(revoked_by_user_id::text, ''), authority_seq FROM app.enrollment_requests WHERE id = $1::text::uuid`, id.String()).Scan(
		&status, &consumed, &decider, &revoker, &seq); err != nil {
		t.Fatalf("request row: %v", err)
	}
	return status, consumed, decider, revoker, seq
}

// reqCount counts user's requests of org in status ("" = all).
func (e *itEnv) reqCount(t *testing.T, org tenancy.OrganizationID, user tenancy.UserID, status string) int {
	t.Helper()
	return e.memCount(t, `SELECT count(*) FROM app.enrollment_requests WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid
  AND ($3 = '' OR status = $3)`, org.String(), user.String(), status)
}

// reqEvents counts the audit rows of action naming request id.
func (e *itEnv) reqEvents(t *testing.T, id tenancy.EnrollmentRequestID, action string) int {
	t.Helper()
	return e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = $1 AND resource_type = 'enrollment_request'
  AND resource_id = $2::text::uuid`, action, id.String())
}

// reqEventActor returns the actor of the single audit row of action naming request id ("" unless exactly one).
func (e *itEnv) reqEventActor(t *testing.T, id tenancy.EnrollmentRequestID, action string) string {
	t.Helper()
	var actor string
	var n int
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*), coalesce(max(actor_user_id::text), '') FROM audit.events
  WHERE action = $1 AND resource_type = 'enrollment_request' AND resource_id = $2::text::uuid AND reason IS NULL
    AND subject_user_id = (SELECT user_id FROM app.enrollment_requests WHERE id = $2::text::uuid)`, action, id.String()).Scan(&n, &actor); err != nil {
		t.Fatalf("event actor: %v", err)
	}
	if n != 1 {
		return ""
	}
	return actor
}

// reqLifted counts user's restrictions in org lifted through an approved request, and the unlifted ones.
func (e *itEnv) reqLifted(t *testing.T, org tenancy.OrganizationID, user tenancy.UserID) (byRequest, unlifted int) {
	t.Helper()
	if err := e.setup.QueryRow(context.Background(), `SELECT count(*) FILTER (WHERE lifted_via = 'approved_request'),
  count(*) FILTER (WHERE lifted_at IS NULL) FROM app.enrollment_restrictions WHERE organization_id = $1::text::uuid
  AND user_id = $2::text::uuid`, org.String(), user.String()).Scan(&byRequest, &unlifted); err != nil {
		t.Fatalf("lifted: %v", err)
	}
	return byRequest, unlifted
}

// enrWaitedRelations names the relations of the tuple locks held by backends that wait for a lock ("" when the waits
// are on transaction ids only, as an INSERT waiting for a conflicting speculative insertion): PostgreSQL takes the
// tuple lock before waiting for the row's locker, so it names the row the waiter queues on.
func (e *itEnv) enrWaitedRelations(t *testing.T) string {
	t.Helper()
	obs := itConnect(t, e.cfg, e.cfg.RuntimeRole(), e.pw)
	var rels string
	if err := obs.QueryRow(context.Background(), `SELECT coalesce(string_agg(DISTINCT t.relation::regclass::text, ','), '')
FROM pg_locks w JOIN pg_locks t ON t.pid = w.pid AND t.locktype = 'tuple' JOIN pg_stat_activity a ON a.pid = w.pid
WHERE NOT w.granted AND a.datname = current_database() AND w.pid <> pg_backend_pid()`).Scan(&rels); err != nil {
		t.Fatal(err)
	}
	return rels
}

// enr9ApproveAdmit (ENR-9 approval path, ENR-17 disclosure): a first request is pending and idempotent per subject (a
// second credential gets the same row; one enrollment.requested by the applicant); status reads it; an org_admin
// approves (decider = the acting admin, one enrollment.decided with reason NULL, authority_seq 0 without
// restrictions) and a second decision is not_pending; the applicant's admission is admitted with one binding and the
// consumption; status is admitted; another credential's admission reconciles to already_member; revoking the consumed
// approval is already_consumed. The views never carry a display name.
func enr9ApproveAdmit(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	_, adm, a := e.reqWorld(t, "ENR9 A")
	user, h := e.reqApplicant(t, a, "R9")
	v := e.reqRequest(t, e.svc, h)
	if v.UserID != user || v.DisplayName != nil || v.Expired || v.Consumed || v.DecidedAt != nil {
		t.Errorf("pending view: %+v", v)
	}
	h2 := e.enrHandle(t, user, a)
	if v2 := e.reqRequest(t, e.svc, h2); v2.ID != v.ID || !v2.CreatedAt.Equal(v.CreatedAt) || !v2.ExpiresAt.Equal(v.ExpiresAt) {
		t.Errorf("second credential: %+v, want the same pending row %s", v2, v.ID)
	}
	if n := e.reqCount(t, a, user, ""); n != 1 || e.reqEvents(t, v.ID, "enrollment.requested") != 1 ||
		e.reqEventActor(t, v.ID, "enrollment.requested") != user.String() {
		t.Errorf("requests %d, requested events %d, actor %q", n, e.reqEvents(t, v.ID, "enrollment.requested"), e.reqEventActor(t, v.ID, "enrollment.requested"))
	}
	if st, err := e.svc.EnrollmentStatus(ctx, h); err != nil || st.Status != tenancy.RequestPending || st.Request == nil || st.Request.ID != v.ID {
		t.Errorf("pending status: %+v %v", st, err)
	}
	d := e.reqDecide(t, adm, a, v.ID, DecisionApprove)
	if d.DecidedAt == nil || d.DisplayName != nil {
		t.Errorf("approved view: %+v", d)
	}
	if status, _, decider, _, seq := e.reqRow(t, v.ID); status != "approved" || decider != adm.id.String() || seq != 0 ||
		e.reqEventActor(t, v.ID, "enrollment.decided") != adm.id.String() {
		t.Errorf("approved row: %s decider %q seq %d, decided actor %q", status, decider, seq, e.reqEventActor(t, v.ID, "enrollment.decided"))
	}
	actx, atc := e.as(t, adm, a)
	_, err := e.svc.DecideEnrollmentRequest(actx, atc, v.ID, DecisionDeny)
	memWantCode(t, "second decision", err, CodeConflict, "not_pending")
	out, err := e.svc.AdmitRequest(ctx, h, v.ID)
	if err != nil || out.Outcome != string(tenancy.AdmitAdmitted) || zero(out.BindingID) {
		t.Fatalf("admit: %+v %v", out, err)
	}
	if _, consumed, _, _, _ := e.reqRow(t, v.ID); !consumed || e.memOrgBindings(t, a, user) != 1 {
		t.Errorf("consumed %v, bindings %d", consumed, e.memOrgBindings(t, a, user))
	}
	if ok, b := e.enrAdmitted(t, h); !ok || b != out.BindingID.String() {
		t.Errorf("credential proof: %v %q", ok, b)
	}
	if st, err := e.svc.EnrollmentStatus(ctx, h); err != nil || st.Status != EnrollmentStatusAdmitted || st.Request == nil || !st.Request.Consumed {
		t.Errorf("admitted status: %+v %v", st, err)
	}
	again, err := e.svc.AdmitRequest(ctx, h2, v.ID)
	if err != nil || again.Outcome != string(tenancy.AdmitAlreadyMember) || again.BindingID != out.BindingID {
		t.Errorf("second credential admission: %+v %v", again, err)
	}
	_, err = e.svc.RevokeEnrollmentApproval(actx, atc, v.ID)
	memWantCode(t, "revoke consumed", err, CodeConflict, "already_consumed")
	if e.memOrgBindings(t, a, user) != 1 || e.reqEvents(t, v.ID, "enrollment.revoked") != 0 || e.reqEvents(t, v.ID, "enrollment.decided") != 1 ||
		e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'membership.granted' AND subject_user_id = $1::text::uuid`, user.String()) != 1 ||
		e.memDenials(t, user) != 0 {
		t.Error("the consumed approval's effects changed after the reconciliation and the refused revocation")
	}
	// A current member's new request is the unrecorded already_member conflict.
	_, err = e.svc.RequestEnrollment(ctx, e.enrHandle(t, user, a))
	memWantCode(t, "member request", err, CodeConflict, fieldAlreadyMember)
	if e.memDenials(t, user) != 0 || e.reqCount(t, a, user, "") != 1 {
		t.Errorf("member request wrote: denials %d, requests %d", e.memDenials(t, user), e.reqCount(t, a, user, ""))
	}
}

// enr9DenyRelift (ENR-9 denial path, ENR-22 denial ledger row): a denial records one denied restriction by the denier
// and one enrollment.decided; the request admission answers the unrecorded status conflict denied; open admission on an
// open Organization is enrollment_restricted after any elapsed time; a new request is allowed (new id) and its approval
// admits, lifting the denial (lifted_via approved_request).
func enr9DenyRelift(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, _, a := e.reqWorld(t, "ENR9D A")
	user, h := e.reqApplicant(t, a, "R9D")
	v := e.reqRequest(t, e.svc, h)
	e.reqDecide(t, o, a, v.ID, DecisionDeny)
	if n := e.memCount(t, `SELECT count(*) FROM app.enrollment_restrictions WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid
  AND kind = 'denied' AND recorded_via = 'trigger' AND created_by_user_id = $3::text::uuid AND source_request_id = $4::text::uuid AND lifted_at IS NULL`,
		a.String(), user.String(), o.id.String(), v.ID.String()); n != 1 || e.reqEventActor(t, v.ID, "enrollment.decided") != o.id.String() {
		t.Errorf("denied ledger rows %d, decided actor %q", n, e.reqEventActor(t, v.ID, "enrollment.decided"))
	}
	_, err := e.svc.AdmitRequest(ctx, h, v.ID)
	memWantCode(t, "admit denied", err, CodeConflict, fieldDenied)
	if e.memOrgBindings(t, a, user) != 0 || e.memDenials(t, user) != 0 {
		t.Errorf("denied admission: bindings %d, denials %d", e.memOrgBindings(t, a, user), e.memDenials(t, user))
	}
	if st, err := e.svc.EnrollmentStatus(ctx, h); err != nil || st.Status != tenancy.RequestDenied {
		t.Errorf("denied status: %+v %v", st, err)
	}
	e.enrSetMode(t, a, "open")
	_, err = e.svc.AdmitOpen(ctx, h)
	memWantCode(t, "open after denial", err, CodeConflict, fieldEnrollmentRestricted)
	e.enrSetMode(t, a, "request_access")
	v2 := e.reqRequest(t, e.svc, h)
	if v2.ID == v.ID {
		t.Fatal("a request after a denial reused the denied row")
	}
	e.reqDecide(t, o, a, v2.ID, DecisionApprove)
	if _, _, _, _, seq := e.reqRow(t, v2.ID); seq < 1 {
		t.Errorf("approval after a denial: authority_seq %d does not cover the denial", seq)
	}
	if out, err := e.svc.AdmitRequest(ctx, h, v2.ID); err != nil || out.Outcome != string(tenancy.AdmitAdmitted) {
		t.Fatalf("admit after re-approval: %+v %v", out, err)
	}
	if byRequest, unlifted := e.reqLifted(t, a, user); byRequest != 1 || unlifted != 0 {
		t.Errorf("lift: by request %d, unlifted %d", byRequest, unlifted)
	}
}

// enr9RevokeStale (ENR-9 revocation and r2 rows, ENR-23 stale approval): a revocation by the admin is idempotent (one
// enrollment.revoked, one revoked restriction by the revoker) and the admission is approval_revoked (one recorded
// denial); an approval followed by a newer denial of another request is enrollment_restricted with zero writes (the
// approval stays approved and unconsumed); after the admin revokes it and approves a third request, the admission lifts
// every older row.
func enr9RevokeStale(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, adm, a := e.reqWorld(t, "ENR9R A")
	user, h := e.reqApplicant(t, a, "R9R")
	v1 := e.reqRequest(t, e.svc, h)
	e.reqDecide(t, o, a, v1.ID, DecisionApprove)
	actx, atc := e.as(t, adm, a)
	r1, err := e.svc.RevokeEnrollmentApproval(actx, atc, v1.ID)
	if err != nil || r1.Status != tenancy.RequestRevoked || r1.RevokedAt == nil {
		t.Fatalf("revoke: %+v %v", r1, err)
	}
	r1b, err := e.svc.RevokeEnrollmentApproval(actx, atc, v1.ID)
	if err != nil || r1b.Status != tenancy.RequestRevoked || !r1b.RevokedAt.Equal(*r1.RevokedAt) {
		t.Errorf("idempotent revoke: %+v %v", r1b, err)
	}
	if _, _, decider, revoker, _ := e.reqRow(t, v1.ID); decider != o.id.String() || revoker != adm.id.String() ||
		e.reqEvents(t, v1.ID, "enrollment.revoked") != 1 || e.reqEventActor(t, v1.ID, "enrollment.revoked") != adm.id.String() ||
		e.memCount(t, `SELECT count(*) FROM app.enrollment_restrictions WHERE source_request_id = $1::text::uuid AND kind = 'revoked'
  AND created_by_user_id = $2::text::uuid`, v1.ID.String(), adm.id.String()) != 1 {
		t.Errorf("revocation records: decider %q revoker %q events %d", decider, revoker, e.reqEvents(t, v1.ID, "enrollment.revoked"))
	}
	_, err = e.svc.AdmitRequest(ctx, h, v1.ID)
	memWantCode(t, "admit revoked", err, CodeConflict, fieldApprovalRevoked)
	if e.enrDenials(t, user, "approval_revoked") != 1 || e.memOrgBindings(t, a, user) != 0 {
		t.Errorf("revoked admission: denials %d, bindings %d", e.enrDenials(t, user, "approval_revoked"), e.memOrgBindings(t, a, user))
	}
	// A pending request cannot be revoked.
	v2 := e.reqRequest(t, e.svc, h)
	_, err = e.svc.RevokeEnrollmentApproval(actx, atc, v2.ID)
	memWantCode(t, "revoke pending", err, CodeConflict, "not_approved")
	e.reqDecide(t, o, a, v2.ID, DecisionApprove)
	v3 := e.reqRequest(t, e.svc, h)
	e.reqDecide(t, o, a, v3.ID, DecisionDeny)
	_, err = e.svc.AdmitRequest(ctx, h, v2.ID)
	memWantCode(t, "stale approval", err, CodeConflict, fieldEnrollmentRestricted)
	if status, consumed, _, _, _ := e.reqRow(t, v2.ID); status != "approved" || consumed || e.memOrgBindings(t, a, user) != 0 {
		t.Errorf("stale approval wrote: %s consumed %v bindings %d", status, consumed, e.memOrgBindings(t, a, user))
	}
	if _, err := e.svc.RevokeEnrollmentApproval(actx, atc, v2.ID); err != nil {
		t.Fatalf("revoke stale approval: %v", err)
	}
	v4 := e.reqRequest(t, e.svc, h)
	e.reqDecide(t, adm, a, v4.ID, DecisionApprove)
	if out, err := e.svc.AdmitRequest(ctx, h, v4.ID); err != nil || out.Outcome != string(tenancy.AdmitAdmitted) {
		t.Fatalf("admit fresh approval: %+v %v", out, err)
	}
	// Three restrictions (v1 revoked, v3 denied, v2 revoked), all covered by v4's watermark and lifted by its decider.
	if byRequest, unlifted := e.reqLifted(t, a, user); byRequest != 3 || unlifted != 0 {
		t.Errorf("lift: by request %d, unlifted %d", byRequest, unlifted)
	}
	if n := e.memCount(t, `SELECT count(*) FROM app.enrollment_restrictions WHERE organization_id = $1::text::uuid AND user_id = $2::text::uuid
  AND lifted_by_user_id = $3::text::uuid`, a.String(), user.String(), adm.id.String()); n != 3 {
		t.Errorf("lifted by the decider: %d", n)
	}
}

// enrRequestAuthority (Plan r3 §1.5 authority rows; Codex request-c1 issued-authority distinction): an org_member
// cannot list, decide or revoke (forbidden, recorded); another Organization's manager cannot see the request
// (not_found); an approval whose decider was demoted is approval_revoked at admission; the applicant's request is
// enrollment_closed outside request_access (recorded on enrollment.request); the manager list pages by id with a status
// filter, never with a display name; malformed list input is invalid_input before any statement.
func enrRequestAuthority(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, adm, a := e.reqWorld(t, "ENRRQA A")
	m := e.newUser(t)
	e.grantOrg(t, a, m, tenancy.OrgMember)
	e.bind(t, m, a)
	ob, b := e.enrWorld(t, "ENRRQA B")
	user, h := e.reqApplicant(t, a, "RQA")
	v := e.reqRequest(t, e.svc, h)
	mctx, mtc := e.as(t, m, a)
	_, err := e.svc.DecideEnrollmentRequest(mctx, mtc, v.ID, DecisionApprove)
	memWantCode(t, "member decides", err, CodeForbidden, "")
	_, err = e.svc.RevokeEnrollmentApproval(mctx, mtc, v.ID)
	memWantCode(t, "member revokes", err, CodeForbidden, "")
	_, err = e.svc.ListEnrollmentRequests(mctx, mtc, "", 0, "")
	memWantCode(t, "member lists", err, CodeForbidden, "")
	if n := e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND actor_user_id = $1::text::uuid
  AND organization_id = $2::text::uuid AND operation LIKE 'enrollment.requests.%'`, m.id.String(), a.String()); n != 3 {
		t.Errorf("member denials recorded: %d, want 3", n)
	}
	bctx, btc := e.as(t, ob, b)
	_, err = e.svc.DecideEnrollmentRequest(bctx, btc, v.ID, DecisionApprove)
	memWantCode(t, "foreign manager decides", err, CodeNotFound, "")
	if status, _, _, _, _ := e.reqRow(t, v.ID); status != "pending" {
		t.Fatalf("refused decisions changed the request: %s", status)
	}
	actx, atc := e.as(t, adm, a)
	_, err = e.svc.DecideEnrollmentRequest(actx, atc, v.ID, "approved")
	memWantCode(t, "unknown decision", err, CodeInvalidInput, fieldDecision)
	_, err = e.svc.DecideEnrollmentRequest(actx, atc, tenancy.EnrollmentRequestID{}, DecisionApprove)
	memWantCode(t, "zero request", err, CodeInvalidInput, "")
	// The list: two more applicants (pending), the first request approved; page size 2 gives a cursor.
	_, h2 := e.reqApplicant(t, a, "RQA2")
	_, h3 := e.reqApplicant(t, a, "RQA3")
	e.reqRequest(t, e.svc, h2)
	e.reqRequest(t, e.svc, h3)
	e.reqDecide(t, adm, a, v.ID, DecisionApprove)
	page, err := e.svc.ListEnrollmentRequests(actx, atc, "", 2, "")
	if err != nil || len(page.Requests) != 2 || page.Next == nil {
		t.Fatalf("first page: %+v %v", page, err)
	}
	rest, err := e.svc.ListEnrollmentRequests(actx, atc, page.Next.String(), 2, "")
	if err != nil || len(rest.Requests) != 1 || rest.Next != nil {
		t.Fatalf("second page: %+v %v", rest, err)
	}
	all := append(page.Requests, rest.Requests...)
	for i, r := range all {
		if r.DisplayName != nil || (i > 0 && r.ID.String() <= all[i-1].ID.String()) {
			t.Errorf("list row %d: %+v (order or display name)", i, r)
		}
	}
	pending, err := e.svc.ListEnrollmentRequests(actx, atc, "", 0, tenancy.RequestPending)
	if err != nil || len(pending.Requests) != 2 {
		t.Errorf("pending filter: %+v %v", pending, err)
	}
	for _, c := range []struct {
		after, status string
		limit         int
		field         string
	}{{"", "", 101, fieldLimit}, {"", "expired", 0, fieldStatus}, {"not-a-uuid", "", 0, fieldCursor}} {
		_, err := e.svc.ListEnrollmentRequests(actx, atc, c.after, c.limit, c.status)
		memWantCode(t, "list input "+c.field, err, CodeInvalidInput, c.field)
	}
	// The decider's demotion (G019's guarded role change by the owner) after the approval: the approval is no longer
	// current authority (role, not status).
	octx, otc := e.as(t, o, a)
	if _, err := e.svc.ChangeMemberRole(octx, otc, memOrgScope, adm.id, string(tenancy.OrgMember)); err != nil {
		t.Fatalf("demote decider: %v", err)
	}
	_, err = e.svc.AdmitRequest(ctx, h, v.ID)
	memWantCode(t, "demoted decider", err, CodeConflict, fieldApprovalRevoked)
	if e.memOrgBindings(t, a, user) != 0 {
		t.Error("demoted decider's approval admitted")
	}
	// The applicant's request outside request_access: enrollment_closed, recorded on enrollment.request.
	for _, mode := range []string{"invite_only", "open"} {
		e.enrSetMode(t, b, mode)
		_, hb := e.reqApplicant(t, b, "RQAB")
		u, _ := tenancy.ParseUserID(hb.UserID)
		_, err := e.svc.RequestEnrollment(ctx, hb)
		memWantCode(t, "request in "+mode, err, CodeConflict, fieldEnrollmentClosed)
		if n := e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND organization_id IS NULL
  AND actor_user_id = $1::text::uuid AND operation = 'enrollment.request' AND reason = 'enrollment_closed'`, u.String()); n != 1 {
			t.Errorf("%s: request denials %d, want 1", mode, n)
		}
	}
}

// enrRequestExpiry (Codex round-8 disposition 3): an expired pending request keeps the one pending slot: a repeat
// request returns the same row unrenewed (Expired, same expires_at), status shows it expired, an administrator may still
// approve it but the approval is never consumable (approval_revoked, no binding), and only a decision frees the slot.
func enrRequestExpiry(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, _, a := e.reqWorld(t, "ENRRQE A")
	user, h := e.reqApplicant(t, a, "RQE")
	v := e.reqRequest(t, e.svc, h)
	tt, err := e.setup.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tt.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
		t.Fatal(err)
	}
	if _, err := tt.Exec(ctx, `UPDATE app.enrollment_requests SET created_at = now() - interval '2 days', expires_at = now() - interval '1 day'
  WHERE id = $1::text::uuid`, v.ID.String()); err != nil {
		t.Fatalf("expire fixture: %v", err)
	}
	if err := tt.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	again := e.reqRequest(t, e.svc, e.enrHandle(t, user, a))
	if again.ID != v.ID || !again.Expired || again.ExpiresAt.After(time.Now()) || e.reqCount(t, a, user, "pending") != 1 {
		t.Errorf("repeat of an expired pending request: %+v (pending rows %d)", again, e.reqCount(t, a, user, "pending"))
	}
	if st, err := e.svc.EnrollmentStatus(ctx, h); err != nil || st.Status != tenancy.RequestPending || st.Request == nil || !st.Request.Expired {
		t.Errorf("expired status: %+v %v", st, err)
	}
	_, err = e.svc.AdmitRequest(ctx, h, v.ID)
	memWantCode(t, "admit expired pending", err, CodeConflict, fieldPending)
	d := e.reqDecide(t, o, a, v.ID, DecisionApprove)
	if !d.Expired || !d.ExpiresAt.Equal(again.ExpiresAt) {
		t.Errorf("approval renewed the lifetime: %+v", d)
	}
	_, err = e.svc.AdmitRequest(ctx, h, v.ID)
	memWantCode(t, "admit expired approval", err, CodeConflict, fieldApprovalRevoked)
	if status, consumed, _, _, _ := e.reqRow(t, v.ID); status != "approved" || consumed || e.memOrgBindings(t, a, user) != 0 {
		t.Errorf("expired approval: %s consumed %v bindings %d", status, consumed, e.memOrgBindings(t, a, user))
	}
	if v2 := e.reqRequest(t, e.svc, h); v2.ID == v.ID || v2.Expired {
		t.Errorf("request after the decision: %+v", v2)
	}
}

// enrRequestRollback (ENR-22 request paths, atomic append): with a planted audit CHECK refusing exactly the one
// capture, the denial, the revocation and the applicant's request each fail as a whole: the request keeps its prior
// state, no restriction row and no event exist; after the plant is removed each succeeds.
func enrRequestRollback(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, _, a := e.reqWorld(t, "ENRRQB A")
	plant := func(action string) func() {
		t.Helper()
		if _, err := e.setup.Exec(ctx, `ALTER TABLE audit.events ADD CONSTRAINT g058_c2_plant CHECK (action <> '`+action+`') NOT VALID`); err != nil {
			t.Fatalf("plant: %v", err)
		}
		return func() {
			if _, err := e.setup.Exec(ctx, `ALTER TABLE audit.events DROP CONSTRAINT IF EXISTS g058_c2_plant`); err != nil {
				t.Fatalf("unplant: %v", err)
			}
		}
	}
	user, h := e.reqApplicant(t, a, "RQB")
	unplant := plant("enrollment.requested")
	_, err := e.svc.RequestEnrollment(ctx, h)
	unplant()
	if CodeOf(err) != CodeInternal || e.reqCount(t, a, user, "") != 0 {
		t.Errorf("planted request: %v, requests %d", err, e.reqCount(t, a, user, ""))
	}
	v := e.reqRequest(t, e.svc, h)
	octx, otc := e.as(t, o, a)
	unplant = plant("enrollment.decided")
	_, err = e.svc.DecideEnrollmentRequest(octx, otc, v.ID, DecisionDeny)
	unplant()
	restrictions := func() int {
		return e.memCount(t, `SELECT count(*) FROM app.enrollment_restrictions WHERE source_request_id = $1::text::uuid`, v.ID.String())
	}
	if status, _, _, _, _ := e.reqRow(t, v.ID); CodeOf(err) != CodeInternal || status != "pending" || restrictions() != 0 ||
		e.reqEvents(t, v.ID, "enrollment.decided") != 0 {
		t.Errorf("planted denial: %v, status %s, restrictions %d", err, status, restrictions())
	}
	e.reqDecide(t, o, a, v.ID, DecisionApprove)
	unplant = plant("enrollment.revoked")
	_, err = e.svc.RevokeEnrollmentApproval(octx, otc, v.ID)
	unplant()
	if status, _, _, revoker, _ := e.reqRow(t, v.ID); CodeOf(err) != CodeInternal || status != "approved" || revoker != "" || restrictions() != 0 {
		t.Errorf("planted revocation: %v, status %s, revoker %q, restrictions %d", err, status, revoker, restrictions())
	}
	if _, err := e.svc.RevokeEnrollmentApproval(octx, otc, v.ID); err != nil || restrictions() != 1 || e.reqEvents(t, v.ID, "enrollment.revoked") != 1 {
		t.Errorf("revocation after unplant: %v, restrictions %d", err, restrictions())
	}
}

// enrRequestCommitOutcomes (request-c1 ruling 2: class-40 abort versus unknown COMMIT): a request whose COMMIT reply is
// lost is unavailable with commit_outcome_unknown after one transaction (never aborted, never retried); the committed
// row is the one pending request and a repeat returns it (one row, one enrollment.requested). A COMMIT the server
// answers with a serialization failure is unavailable with aborted, nothing is committed, and a repeat inserts the
// request once.
func enrRequestCommitOutcomes(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	_, _, a := e.reqWorld(t, "ENRRQL A")
	reason := func(from int) string {
		recs := e.memOpRecords(t, from, opEnrollmentRequest)
		if len(recs) != 1 {
			t.Fatalf("%d enrollment.request records, want 1", len(recs))
		}
		r, _ := recs[0]["reason"].(string)
		return r
	}
	user, h := e.reqApplicant(t, a, "RQL")
	e.hook.reset()
	e.hook.setCommit(commitLostReply)
	from := len(e.logs.String())
	_, err := e.svc.RequestEnrollment(ctx, h)
	memWantCode(t, "lost reply", err, CodeUnavailable, "")
	if begins, _ := e.hook.counts(); begins != 1 || reason(from) != string(reasonCommitOutcomeUnknown) {
		t.Errorf("lost reply: %d transactions, reason %q", begins, reason(from))
	}
	e.hook.reset()
	if e.reqCount(t, a, user, "pending") != 1 {
		t.Fatalf("lost reply: the COMMIT did not land (%d pending)", e.reqCount(t, a, user, "pending"))
	}
	v := e.reqRequest(t, e.svc, h)
	if e.reqCount(t, a, user, "") != 1 || e.reqEvents(t, v.ID, "enrollment.requested") != 1 {
		t.Errorf("repeat after lost reply: requests %d, events %d", e.reqCount(t, a, user, ""), e.reqEvents(t, v.ID, "enrollment.requested"))
	}
	user2, h2 := e.reqApplicant(t, a, "RQL2")
	e.hook.setCommit(commitDefinitive)
	from = len(e.logs.String())
	_, err = e.svc.RequestEnrollment(ctx, h2)
	memWantCode(t, "serialization failure at COMMIT", err, CodeUnavailable, "")
	if reason(from) != string(reasonAborted) || e.reqCount(t, a, user2, "") != 0 {
		t.Errorf("aborted COMMIT: reason %q, requests %d", reason(from), e.reqCount(t, a, user2, ""))
	}
	e.hook.reset()
	v2 := e.reqRequest(t, e.svc, h2)
	if e.reqCount(t, a, user2, "") != 1 || e.reqEvents(t, v2.ID, "enrollment.requested") != 1 {
		t.Errorf("repeat after abort: requests %d", e.reqCount(t, a, user2, ""))
	}
}

// oEnrOnePendingRace (request-c1 ruling 2, the actual two-credential race): two credentials of one subject request
// concurrently. The first request's transaction is held open right after its enrollment_request statement inserted the
// pending row; the second, on another connection, reaches the INSERT and waits on the one_pending unique index (a lock
// wait observed in pg_locks while it has not returned). After the first commits, the second's ON CONFLICT DO NOTHING
// reads the committed row: both answers name the same request, exactly one pending row and one enrollment.requested
// exist, and neither call failed (no 23505, no 40001 on this path).
func oEnrOnePendingRace(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	_, _, a := e.reqWorld(t, "OENRRQ A")
	user, h1 := e.reqApplicant(t, a, "ORQ")
	h2 := e.enrHandle(t, user, a)
	tracer := &pauseTracer{}
	svc1 := NewService(e.runtimePool(t, tracer), e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil))
	var v1, v2 EnrollmentRequestView
	p := tracer.arm("app.enrollment_request(")
	first := start(func() error {
		var err error
		v1, err = svc1.RequestEnrollment(ctx, h1)
		return err
	})
	p.wait(t, "first request")
	second := start(func() error {
		var err error
		v2, err = e.svc.RequestEnrollment(ctx, h2)
		return err
	})
	e.waitWaiters(t, 1)
	seam := "second_waits_on_one_pending"
	if second.finished() {
		seam = "second_done"
	}
	rels := e.enrWaitedRelations(t)
	close(p.release)
	e1, e2 := first.wait(t, "first request"), second.wait(t, "second request")
	if e1 != nil || e2 != nil || seam != "second_waits_on_one_pending" || v1.ID != v2.ID || zero(v1.ID) ||
		v1.Status != tenancy.RequestPending || v2.Status != tenancy.RequestPending {
		t.Errorf("race (%s): %+v %v / %+v %v, want one shared pending request", seam, v1, e1, v2, e2)
	}
	if n := e.reqCount(t, a, user, ""); n != 1 || e.reqEvents(t, v1.ID, "enrollment.requested") != 1 {
		t.Errorf("race wrote %d requests, %d requested events", n, e.reqEvents(t, v1.ID, "enrollment.requested"))
	}
	ordering(t, map[string]any{"case": "O-ENR-RQ", "seam": seam, "tuple_lock_relations": rels, "same_request": v1.ID == v2.ID})
}

// oEnr7TwoConsumers (O-ENR-7): two credentials of the applicant consume one approval concurrently. The first admission
// is held open after its enrollment_admit statement (holding the request row FOR UPDATE); the second queues on that row
// (observed lock wait; tuple lock on app.enrollment_requests) and then reads the consumption: admitted then
// already_member with the same binding, exactly one binding and one membership.granted.
func oEnr7TwoConsumers(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, _, a := e.reqWorld(t, "OENR7 A")
	user, h1 := e.reqApplicant(t, a, "O7")
	h2 := e.enrHandle(t, user, a)
	v := e.reqRequest(t, e.svc, h1)
	e.reqDecide(t, o, a, v.ID, DecisionApprove)
	tracer := &pauseTracer{}
	svc1 := NewService(e.runtimePool(t, tracer), e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil))
	var o1, o2 EnrollmentAdmission
	p := tracer.arm("app.enrollment_admit(")
	first := start(func() error {
		var err error
		o1, err = svc1.AdmitRequest(ctx, h1, v.ID)
		return err
	})
	p.wait(t, "first consumer")
	second := start(func() error {
		var err error
		o2, err = e.svc.AdmitRequest(ctx, h2, v.ID)
		return err
	})
	e.waitWaiters(t, 1)
	rels := e.enrWaitedRelations(t)
	waited := !second.finished()
	close(p.release)
	e1, e2 := first.wait(t, "first consumer"), second.wait(t, "second consumer")
	if e1 != nil || e2 != nil || !waited || o1.Outcome != string(tenancy.AdmitAdmitted) || o2.Outcome != string(tenancy.AdmitAlreadyMember) ||
		o1.BindingID != o2.BindingID || !strings.Contains(rels, "enrollment_requests") {
		t.Errorf("consumers (waited %v, %q): %+v %v / %+v %v", waited, rels, o1, e1, o2, e2)
	}
	if e.memOrgBindings(t, a, user) != 1 || e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'membership.granted'
  AND subject_user_id = $1::text::uuid`, user.String()) != 1 {
		t.Errorf("two consumers: bindings %d", e.memOrgBindings(t, a, user))
	}
	ordering(t, map[string]any{"case": "O-ENR-7", "tuple_lock_relations": rels, "outcome": o1.Outcome + " then " + o2.Outcome})
}

// oEnrRevokeConsume (ENR-10 revocation versus consumption, Plan r3 §1.4 decision/revocation order): consumption first —
// the admission is held open after enrollment_admit (Organization FOR SHARE, request row FOR UPDATE); the revocation's
// guarded facts queue on the Organization row (observed wait, tuple lock on app.organizations) before touching the
// request; after the admission commits the revocation answers already_consumed with nothing written. Revocation first —
// the revocation is held before its COMMIT (Organization FOR NO KEY UPDATE, request row updated); the admission queues
// on the Organization row and then reads the revoked approval: approval_revoked, no binding. One winner in each order,
// no partial write, no deadlock.
func oEnrRevokeConsume(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, _, a := e.reqWorld(t, "OENRRV A")
	var zeroG tenancy.GroupID

	t.Run("consumption_first", func(t *testing.T) {
		user, h := e.reqApplicant(t, a, "ORV1")
		v := e.reqRequest(t, e.svc, h)
		e.reqDecide(t, o, a, v.ID, DecisionApprove)
		tracer := &pauseTracer{}
		svc1 := NewService(e.runtimePool(t, tracer), e.id, observability.NewLogger(e.logs, slog.LevelInfo, nil))
		octx, otc := e.as(t, o, a)
		p := tracer.arm("app.enrollment_admit(")
		var out EnrollmentAdmission
		adm := start(func() error {
			var err error
			out, err = svc1.AdmitRequest(ctx, h, v.ID)
			return err
		})
		p.wait(t, "admission")
		rev := start(func() error { _, err := e.svc.RevokeEnrollmentApproval(octx, otc, v.ID); return err })
		e.waitWaiters(t, 1)
		rels := e.enrWaitedRelations(t)
		waited := !rev.finished()
		close(p.release)
		aerr, rerr := adm.wait(t, "admission"), rev.wait(t, "revocation")
		if aerr != nil || out.Outcome != string(tenancy.AdmitAdmitted) || !waited || !strings.Contains(rels, "organizations") {
			t.Errorf("consumption_first (waited %v, %q): admission %+v %v", waited, rels, out, aerr)
		}
		memWantCode(t, "consumption_first revocation", rerr, CodeConflict, "already_consumed")
		if status, consumed, _, revoker, _ := e.reqRow(t, v.ID); status != "approved" || !consumed || revoker != "" ||
			e.memOrgBindings(t, a, user) != 1 || e.reqEvents(t, v.ID, "enrollment.revoked") != 0 {
			t.Errorf("consumption_first: %s consumed %v revoker %q bindings %d", status, consumed, revoker, e.memOrgBindings(t, a, user))
		}
		ordering(t, map[string]any{"case": "O-ENR-RV", "order": "consumption_first", "tuple_lock_relations": rels, "revocation": fmt.Sprint(CodeOf(rerr))})
	})

	t.Run("revocation_first", func(t *testing.T) {
		user, h := e.reqApplicant(t, a, "ORV2")
		v := e.reqRequest(t, e.svc, h)
		e.reqDecide(t, o, a, v.ID, DecisionApprove)
		s2, h2 := e.service(e.id)
		rctx, _ := beginAs(t, s2, ctx, o, a)
		rtc := resolve(t, s2, rctx, a, zeroG)
		pp := newPause()
		h2.on("before", "COMMIT", pp.hook)
		rev := start(func() error { _, err := s2.RevokeEnrollmentApproval(rctx, rtc, v.ID); return err })
		pp.wait(t, "revocation")
		var out EnrollmentAdmission
		adm := start(func() error {
			var err error
			out, err = e.svc.AdmitRequest(ctx, h, v.ID)
			return err
		})
		e.waitWaiters(t, 1)
		rels := e.enrWaitedRelations(t)
		waited := !adm.finished()
		close(pp.release)
		rerr, aerr := rev.wait(t, "revocation"), adm.wait(t, "admission")
		if rerr != nil || !waited || !strings.Contains(rels, "organizations") {
			t.Errorf("revocation_first (waited %v, %q): revocation %v", waited, rels, rerr)
		}
		memWantCode(t, "revocation_first admission", aerr, CodeConflict, fieldApprovalRevoked)
		if status, consumed, _, _, _ := e.reqRow(t, v.ID); status != "revoked" || consumed || e.memOrgBindings(t, a, user) != 0 {
			t.Errorf("revocation_first: %s consumed %v bindings %d (out %+v)", status, consumed, e.memOrgBindings(t, a, user), out)
		}
		ordering(t, map[string]any{"case": "O-ENR-RV", "order": "revocation_first", "tuple_lock_relations": rels, "admission": fmt.Sprint(CodeOf(aerr))})
	})
}

// oEnr6ApprovalRemoval (O-ENR-6, Plan r3 §1.4 reinstatement authority): the applicant (pending request) is also a
// current member and is removed concurrently with the approval. Removal first — the removal is held before its COMMIT
// (its departure body holds the Organization FOR NO KEY UPDATE); the approval's guarded facts queue on the Organization
// row (observed wait) and its guard then derives authority_seq over the committed departed row: the admission is
// admitted and lifts it. Approval first — the approval is held before its COMMIT; the removal queues on the
// Organization row; its departed row is newer than the approval's watermark: the admission is enrollment_restricted
// with zero writes.
func oEnr6ApprovalRemoval(t *testing.T) {
	e := setupIT(t)
	ctx := context.Background()
	o, _, a := e.reqWorld(t, "OENR6 A")
	var zeroG tenancy.GroupID

	t.Run("removal_first", func(t *testing.T) {
		user, h := e.reqApplicant(t, a, "O6A")
		v := e.reqRequest(t, e.svc, h)
		e.grantOrg(t, a, &itUser{id: user}, tenancy.OrgMember)
		s2, h2 := e.service(e.id)
		rctx, _ := beginAs(t, s2, ctx, o, a)
		rtc := resolve(t, s2, rctx, a, zeroG)
		pp := newPause()
		h2.on("before", "COMMIT", pp.hook)
		rem := start(func() error { return s2.RemoveMember(rctx, rtc, memOrgScope, user) })
		pp.wait(t, "removal")
		octx, otc := e.as(t, o, a)
		dec := start(func() error { _, err := e.svc.DecideEnrollmentRequest(octx, otc, v.ID, DecisionApprove); return err })
		e.waitWaiters(t, 1)
		rels := e.enrWaitedRelations(t)
		waited := !dec.finished()
		close(pp.release)
		rerr, derr := rem.wait(t, "removal"), dec.wait(t, "approval")
		if rerr != nil || derr != nil || !waited || !strings.Contains(rels, "organizations") {
			t.Fatalf("removal_first (waited %v, %q): removal %v, approval %v", waited, rels, rerr, derr)
		}
		var departed int64
		if err := e.setup.QueryRow(ctx, `SELECT seq FROM app.enrollment_restrictions WHERE organization_id = $1::text::uuid
  AND user_id = $2::text::uuid AND kind = 'departed'`, a.String(), user.String()).Scan(&departed); err != nil {
			t.Fatal(err)
		}
		if _, _, _, _, seq := e.reqRow(t, v.ID); seq != departed {
			t.Errorf("removal_first: authority_seq %d, departed seq %d", seq, departed)
		}
		out, err := e.svc.AdmitRequest(ctx, h, v.ID)
		if err != nil || out.Outcome != string(tenancy.AdmitAdmitted) {
			t.Errorf("removal_first admission: %+v %v", out, err)
		}
		if byRequest, unlifted := e.reqLifted(t, a, user); byRequest != 1 || unlifted != 0 {
			t.Errorf("removal_first lift: by request %d, unlifted %d", byRequest, unlifted)
		}
		ordering(t, map[string]any{"case": "O-ENR-6", "order": "removal_first", "tuple_lock_relations": rels, "admission": out.Outcome})
	})

	t.Run("approval_first", func(t *testing.T) {
		user, h := e.reqApplicant(t, a, "O6B")
		v := e.reqRequest(t, e.svc, h)
		e.grantOrg(t, a, &itUser{id: user}, tenancy.OrgMember)
		s2, h2 := e.service(e.id)
		dctx, _ := beginAs(t, s2, ctx, o, a)
		dtc := resolve(t, s2, dctx, a, zeroG)
		pp := newPause()
		h2.on("before", "COMMIT", pp.hook)
		dec := start(func() error { _, err := s2.DecideEnrollmentRequest(dctx, dtc, v.ID, DecisionApprove); return err })
		pp.wait(t, "approval")
		octx, otc := e.as(t, o, a)
		rem := start(func() error { return e.svc.RemoveMember(octx, otc, memOrgScope, user) })
		e.waitWaiters(t, 1)
		rels := e.enrWaitedRelations(t)
		waited := !rem.finished()
		close(pp.release)
		derr, rerr := dec.wait(t, "approval"), rem.wait(t, "removal")
		if derr != nil || rerr != nil || !waited || !strings.Contains(rels, "organizations") {
			t.Fatalf("approval_first (waited %v, %q): approval %v, removal %v", waited, rels, derr, rerr)
		}
		_, err := e.svc.AdmitRequest(ctx, h, v.ID)
		memWantCode(t, "approval_first admission", err, CodeConflict, fieldEnrollmentRestricted)
		if status, consumed, _, _, _ := e.reqRow(t, v.ID); status != "approved" || consumed || e.memOrgBindings(t, a, user) != 0 {
			t.Errorf("approval_first: %s consumed %v bindings %d", status, consumed, e.memOrgBindings(t, a, user))
		}
		if _, unlifted := e.reqLifted(t, a, user); unlifted != 1 {
			t.Errorf("approval_first: unlifted %d, want the newer departed row", unlifted)
		}
		ordering(t, map[string]any{"case": "O-ENR-6", "order": "approval_first", "tuple_lock_relations": rels})
	})
}

// TestEnrollment is the G058 enrollment policy entry point (Plan r3 §6.2 ENR-1 and ENR-3, the facade parts;
// selected-design addendum §A.5/§A.6). Each subtest builds its own environment; a row selector is
// '^TestEnrollment$/^ENR-3_'. The legacy Organization is closed through the operator setter as the attested superuser
// (the audited procedure itself, not a fixture write).
func TestEnrollment(t *testing.T) {
	for _, c := range []struct {
		name string
		run  func(*testing.T)
	}{
		{"ENR-1 closed default and ready by construction", enr1ClosedDefault},
		{"ENR-3 policy update authority and readiness", enr3PolicyAuthority},
	} {
		t.Run(c.name, c.run)
	}
}

// enrMode reads org's stored enrollment mode and readiness (setup connection).
func (e *itEnv) enrMode(t *testing.T, org tenancy.OrganizationID) (mode string, ready bool) {
	t.Helper()
	if err := e.setup.QueryRow(context.Background(), `SELECT enrollment_mode, enrollment_history_resolved_at IS NOT NULL
  FROM app.organizations WHERE id = $1::text::uuid`, org.String()).Scan(&mode, &ready); err != nil {
		t.Fatalf("organization mode: %v", err)
	}
	return mode, ready
}

// enrPolicyEvents counts org's enrollment.policy_changed rows (actor, Organization and resource as the capture writes).
func (e *itEnv) enrPolicyEvents(t *testing.T, org tenancy.OrganizationID, actor tenancy.UserID) int {
	t.Helper()
	return e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'enrollment.policy_changed' AND outcome = 'success'
  AND organization_id = $1::text::uuid AND resource_type = 'organization' AND resource_id = $1::text::uuid AND reason IS NULL
  AND actor_kind = 'user' AND actor_user_id = $2::text::uuid`, org.String(), actor.String())
}

// enrPolicyDenials counts actor's access.denied records of operation op in org.
func (e *itEnv) enrPolicyDenials(t *testing.T, org tenancy.OrganizationID, actor tenancy.UserID, op string) int {
	t.Helper()
	return e.memCount(t, `SELECT count(*) FROM audit.events WHERE action = 'access.denied' AND organization_id = $1::text::uuid
  AND actor_user_id = $2::text::uuid AND operation = $3`, org.String(), actor.String(), op)
}

// enrReadiness runs the operator setter as the attested superuser and returns its outcome.
func (e *itEnv) enrReadiness(t *testing.T, org tenancy.OrganizationID, basis any, ref string) string {
	t.Helper()
	var out string
	if err := e.setup.QueryRow(context.Background(), `SELECT app.enrollment_readiness_set($1::text::uuid, $2::text, $3)`,
		org.String(), basis, ref).Scan(&out); err != nil {
		t.Fatalf("enrollment_readiness_set: %v", err)
	}
	return out
}

// enr1ClosedDefault (ENR-1, facade part): a fresh Organization is invite_only and ready by construction
// (created_after_ledger); GetEnrollmentPolicy (organization.read) exposes exactly that to the owner and to a member,
// writes nothing and records no denial.
func enr1ClosedDefault(t *testing.T) {
	e := setupIT(t)
	o, a := e.enrWorld(t, "ENR1 A")
	m := e.newUser(t)
	e.grantOrg(t, a, m, tenancy.OrgMember)
	e.bind(t, m, a)
	want := tenancy.EnrollmentPolicy{Mode: tenancy.EnrollmentInviteOnly, OpenReady: true, HistoryBasis: "created_after_ledger"}
	for name, u := range map[string]*itUser{"owner": o, "member": m} {
		ctx, tc := e.as(t, u, a)
		if p, err := e.svc.GetEnrollmentPolicy(ctx, tc); err != nil || p != want {
			t.Errorf("%s policy: %+v %v, want %+v", name, p, err, want)
		}
	}
	if n := e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid
  AND action IN ('enrollment.policy_changed', 'enrollment.readiness_changed', 'access.denied')`, a.String()); n != 0 {
		t.Errorf("policy reads wrote %d audit rows", n)
	}
	// G058 HTTP slice (ENR-1 public half, ENR-H public projection): anonymous, no context, on the runtime pool — the
	// closed default is invite_only with no label; the mode follows the stored policy; an unknown Organization reads the
	// closed invite_only; nothing is recorded.
	ctx := context.Background()
	psvc := NewService(e.pool, e.id, nil) // the public read (single statements on the runtime pool, as the HTTP handler's)
	events := func() int { return e.memCount(t, `SELECT count(*) FROM audit.events`) }
	before := events()
	if p, err := psvc.PublicEnrollment(ctx, a.String()); err != nil || p.Mode != string(tenancy.EnrollmentInviteOnly) || p.Label != nil {
		t.Errorf("public closed default: %+v %v", p, err)
	}
	e.enrSetMode(t, a, "request_access")
	if p, err := psvc.PublicEnrollment(ctx, a.String()); err != nil || p.Mode != string(tenancy.EnrollmentRequestAccess) {
		t.Errorf("public request_access: %+v %v", p, err)
	}
	if p, err := psvc.PublicEnrollment(ctx, "0f58a000-0000-4000-8000-00000000ffff"); err != nil || p.Mode != string(tenancy.EnrollmentInviteOnly) {
		t.Errorf("public unknown Organization: %+v %v", p, err)
	}
	if n := events() - before; n != 0 {
		t.Errorf("public reads wrote %d audit rows", n)
	}
}

// enr3PolicyAuthority (ENR-3, facade part; addendum §A.6): organization.manage_policy is the owner's alone — an
// administrator and a member are forbidden (one access.denied each, mode unchanged, no policy_changed); an unknown mode
// is invalid_input before any write. The owner's request_access writes the mode with one enrollment.policy_changed in the
// same transaction, a repeat is a no-op without a row, open on a ready Organization succeeds. On a legacy (closed through
// the operator setter) Organization the owner's open is ErrEnrollmentNotReady with no UPDATE, no audit row and no denial,
// while request_access stays allowed; after the operator attests verified absence, open succeeds.
func enr3PolicyAuthority(t *testing.T) {
	e := setupIT(t)
	o, a := e.enrWorld(t, "ENR3 A")
	adm, mem := e.newUser(t), e.newUser(t)
	e.grantOrg(t, a, adm, tenancy.OrgAdmin)
	e.grantOrg(t, a, mem, tenancy.OrgMember)
	e.bind(t, adm, a)
	e.bind(t, mem, a)

	for name, u := range map[string]*itUser{"admin": adm, "member": mem} {
		ctx, tc := e.as(t, u, a)
		if _, err := e.svc.UpdateEnrollmentPolicy(ctx, tc, tenancy.EnrollmentRequestAccess); CodeOf(err) != CodeForbidden {
			t.Errorf("%s update: %v, want forbidden", name, err)
		}
		if n := e.enrPolicyDenials(t, a, u.id, opEnrollmentPolicyUpdate); n != 1 {
			t.Errorf("%s denials: %d, want 1", name, n)
		}
		if p, err := e.svc.GetEnrollmentPolicy(ctx, tc); err != nil || p.Mode != tenancy.EnrollmentInviteOnly {
			t.Errorf("%s read: %+v %v", name, p, err)
		}
	}
	ctx, tc := e.as(t, o, a)
	for _, bad := range []tenancy.EnrollmentMode{"", "closed", "OPEN"} {
		var ae *Error
		if _, err := e.svc.UpdateEnrollmentPolicy(ctx, tc, bad); !errors.As(err, &ae) || ae.Code != CodeInvalidInput ||
			ae.Field != string(tenancy.CodeInvalidEnrollmentMode) {
			t.Errorf("mode %q: %v, want invalid_input invalid_enrollment_mode", bad, err)
		}
	}
	if mode, _ := e.enrMode(t, a); mode != "invite_only" || e.enrPolicyEvents(t, a, o.id) != 0 {
		t.Fatalf("refused updates wrote: mode %s", mode)
	}

	// The owner on a ready Organization: request_access (one row), the same value again (no row), open (one row).
	if p, err := e.svc.UpdateEnrollmentPolicy(ctx, tc, tenancy.EnrollmentRequestAccess); err != nil ||
		p != (tenancy.EnrollmentPolicy{Mode: tenancy.EnrollmentRequestAccess, OpenReady: true, HistoryBasis: "created_after_ledger"}) {
		t.Fatalf("owner request_access: %+v %v", p, err)
	}
	if _, err := e.svc.UpdateEnrollmentPolicy(ctx, tc, tenancy.EnrollmentRequestAccess); err != nil {
		t.Fatalf("owner repeat: %v", err)
	}
	if n := e.enrPolicyEvents(t, a, o.id); n != 1 {
		t.Errorf("policy_changed after request_access and repeat: %d, want 1", n)
	}
	if p, err := e.svc.UpdateEnrollmentPolicy(ctx, tc, tenancy.EnrollmentOpen); err != nil || p.Mode != tenancy.EnrollmentOpen {
		t.Fatalf("owner open (ready): %+v %v", p, err)
	}
	if mode, _ := e.enrMode(t, a); mode != "open" || e.enrPolicyEvents(t, a, o.id) != 2 {
		t.Errorf("open: mode %s, events %d", mode, e.enrPolicyEvents(t, a, o.id))
	}

	// Legacy: the operator close flips open to invite_only (its own two rows, actor unknown); the owner sees not ready.
	if out := e.enrReadiness(t, a, nil, "enr3.close"); out != "closed" {
		t.Fatalf("operator close: %s", out)
	}
	if p, err := e.svc.GetEnrollmentPolicy(ctx, tc); err != nil || p != (tenancy.EnrollmentPolicy{Mode: tenancy.EnrollmentInviteOnly}) {
		t.Fatalf("legacy read: %+v %v", p, err)
	}
	before := e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid`, a.String())
	if _, err := e.svc.UpdateEnrollmentPolicy(ctx, tc, tenancy.EnrollmentOpen); !errors.Is(err, ErrEnrollmentNotReady) {
		t.Fatalf("legacy open: %v, want ErrEnrollmentNotReady", err)
	}
	if mode, ready := e.enrMode(t, a); mode != "invite_only" || ready {
		t.Errorf("legacy open wrote: mode %s, ready %v", mode, ready)
	}
	if after := e.memCount(t, `SELECT count(*) FROM audit.events WHERE organization_id = $1::text::uuid`, a.String()); after != before {
		t.Errorf("legacy open wrote %d audit rows", after-before)
	}
	if p, err := e.svc.UpdateEnrollmentPolicy(ctx, tc, tenancy.EnrollmentRequestAccess); err != nil ||
		p != (tenancy.EnrollmentPolicy{Mode: tenancy.EnrollmentRequestAccess}) {
		t.Fatalf("legacy request_access: %+v %v", p, err)
	}
	if n := e.enrPolicyEvents(t, a, o.id); n != 3 {
		t.Errorf("policy_changed after legacy request_access: %d, want 3", n)
	}
	if out := e.enrReadiness(t, a, "operator_verified_absent", "enr3.verify"); out != "resolved" {
		t.Fatalf("operator attestation: %s", out)
	}
	if p, err := e.svc.UpdateEnrollmentPolicy(ctx, tc, tenancy.EnrollmentOpen); err != nil ||
		p != (tenancy.EnrollmentPolicy{Mode: tenancy.EnrollmentOpen, OpenReady: true, HistoryBasis: "operator_verified_absent"}) {
		t.Fatalf("open after attestation: %+v %v", p, err)
	}
	if n := e.enrPolicyDenials(t, a, o.id, opEnrollmentPolicyUpdate); n != 0 {
		t.Errorf("owner denials: %d, want 0", n)
	}
}
