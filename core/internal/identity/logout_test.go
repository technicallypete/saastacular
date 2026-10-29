//go:build integration

package identity_test

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"core/internal/identity"
)

// Plan r5 §4.4 / §3.5: local sign-out, sign-out everywhere and provider sign-out outcomes, with state assertions.
func TestLocalLogoutAndCSRF(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	a := h.mustSignIn(originA, grant("user_01LOA", "session_loa")).Credential
	b := h.mustSignIn(originA, grant("user_01LOB", "session_lob")).Credential
	aInfo, bInfo := h.session(a), h.session(b)

	// Rejections change nothing: wrong origin, missing, malformed, other user's or stale CSRF token, unknown scope. G016
	// security amendment: at another origin the credential is not a session there (bound origin), so current is the
	// idempotent local outcome and all/provider are unauthenticated; nothing is revoked.
	for name, in := range map[string]identity.LogoutInput{
		"origin":                   {Scope: identity.ScopeCurrent, CSRF: csrfOf(aInfo), Origin: "https://attacker.localhost:543", Credential: a},
		"origin all":               {Scope: identity.ScopeAll, CSRF: csrfOf(aInfo), Origin: "https://attacker.localhost:543", Credential: a},
		"origin provider":          {Scope: identity.ScopeProvider, CSRF: csrfOf(aInfo), Origin: originB, Credential: a},
		"missing csrf":             {Scope: identity.ScopeCurrent, Origin: originA, Credential: a},
		"malformed csrf":           {Scope: identity.ScopeCurrent, CSRF: "x", Origin: originA, Credential: a},
		"cross-user csrf":          {Scope: identity.ScopeCurrent, CSRF: csrfOf(bInfo), Origin: originA, Credential: a},
		"all with cross-user csrf": {Scope: identity.ScopeAll, CSRF: csrfOf(bInfo), Origin: originA, Credential: a},
		"unknown scope":            {Scope: "everything", CSRF: csrfOf(aInfo), Origin: originA, Credential: a},
	} {
		_, err := h.svc.Logout(ctx, in)
		var want error = identity.ErrCSRFFailed
		switch name {
		case "origin":
			want = nil
		case "origin all", "origin provider":
			want = identity.ErrUnauthenticated
		case "unknown scope":
			want = identity.ErrInvalidRequest
		}
		if err != want {
			t.Errorf("%s: %v", name, err)
		}
	}
	// G016 LOW-1: a missing or malformed origin is origin_rejected for every scope before any session or routing
	// lookup; nothing is revoked and the provider is never called.
	for _, origin := range []string{"", "null", "not an origin", "https://", "https://u@org-a.localhost:543",
		"https://org-a.localhost:543/", "https://org-a.localhost:543/sign-in", "https://org-a.localhost:543?x=1",
		"https://org-a.localhost:543?", "https://org-a.localhost:543#f", "https://ORG-A.localhost:543",
		"https://org-a.localhost:", "https://org-a.localhost:0543", "https://org-a.localhost:99999", "http://org-a.localhost:543",
		"ftp://org-a.localhost:543", "//org-a.localhost:543", "https://" + strings.Repeat("a", 300) + ".localhost"} {
		for _, scope := range []string{identity.ScopeCurrent, identity.ScopeAll, identity.ScopeProvider} {
			if _, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: scope, CSRF: csrfOf(aInfo), Origin: origin, Credential: a}); err != identity.ErrOriginRejected {
				t.Errorf("origin %.40q scope %s: %v, want origin_rejected", origin, scope, err)
			}
		}
	}
	if got := h.fake.Revoked(); len(got) != 0 {
		t.Fatalf("a rejected logout called the provider: %v", got)
	}
	if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE revoked_at IS NOT NULL`); n != 0 {
		t.Fatalf("a rejected logout revoked %d sessions", n)
	}
	if _, err := h.svc.Session(ctx, a, originA); err != nil {
		t.Fatalf("a rejected logout revoked the session: %v", err)
	}

	// The token stays valid across reloads (repeated resolution) and is per session.
	if again := h.session(a); csrfOf(again) != csrfOf(aInfo) {
		t.Fatal("CSRF token changed within a session")
	}

	// Local sign-out commits revocation; the copied credential is denied; repeating is idempotent.
	out, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: csrfOf(aInfo), Origin: originA, Credential: a})
	if err != nil || out.Outcome != identity.OutcomeLocalRevoked {
		t.Fatalf("logout: %+v %v", out, err)
	}
	if _, err := h.svc.Session(ctx, a, originA); err != identity.ErrUnauthenticated {
		t.Fatalf("copied credential after logout: %v", err)
	}
	if out, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, Origin: originA, Credential: a}); err != nil || out.Outcome != identity.OutcomeLocalRevoked {
		t.Fatalf("repeated logout: %+v %v", out, err)
	}
	if _, err := h.svc.Session(ctx, b, originA); err != nil {
		t.Fatalf("another user's session affected: %v", err)
	}
	if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE revoke_reason = 'logout'`); n != 1 {
		t.Fatalf("revocations %d", n)
	}
	// E-LO-2: sign out everywhere with a dead session is 401 and revokes nothing.
	if _, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: csrfOf(aInfo), Origin: originA, Credential: a}); err != identity.ErrUnauthenticated {
		t.Fatalf("logout-all with a dead session: %v", err)
	}
}

// I-ALL / E-LO-4: sign out everywhere revokes every session of the caller only.
func TestLogoutAll(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	u1a := h.mustSignIn(originA, grant("user_01ALL1", "session_all1a")).Credential
	u1b := h.mustSignIn(originB, grant("user_01ALL1", "session_all1b")).Credential
	u2 := h.mustSignIn(originA, grant("user_01ALL2", "session_all2")).Credential
	info := h.session(u1a)
	out, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: csrfOf(info), Origin: originA, Credential: u1a})
	if err != nil || out.Outcome != identity.OutcomeLocalRevoked {
		t.Fatalf("logout-all: %+v %v", out, err)
	}
	for _, c := range []string{u1a, u1b} {
		if _, err := h.svc.Session(ctx, c, originA); err != identity.ErrUnauthenticated {
			t.Fatalf("user 1 session survived: %v", err)
		}
	}
	if _, err := h.svc.Session(ctx, u2, originA); err != nil {
		t.Fatalf("user 2 affected: %v", err)
	}
	// A new sign-in afterwards is a new session, never a revival.
	again := h.mustSignIn(originA, grant("user_01ALL1", "session_all1c")).Credential
	if again == u1a || again == u1b {
		t.Fatal("credential reused")
	}
}

// Provider sign-out (Plan r5 §3.5; C7 addendum §2): commit local revocation, revoke exactly this sid, then redirect
// only when the revocation is confirmed. The return reference is one-use and server-recorded, travels only in its own
// field and never in a URL; the provider returns to exactly the requesting origin's registered sign-in page.
func TestProviderLogout(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	cred := h.mustSignIn(originA, grant("user_01PL", "session_pl1")).Credential
	info := h.session(cred)
	out, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: csrfOf(info), Origin: originA, Credential: cred})
	if err != nil || out.Outcome != identity.OutcomeProviderRedirect || out.ProviderLogoutURL == "" {
		t.Fatalf("provider logout: %+v %v", out, err)
	}
	if got := h.fake.Revoked(); !slices.Equal(got, []string{"session_pl1"}) {
		t.Fatalf("revoked %v", got)
	}
	if _, err := h.svc.Session(ctx, cred, originA); err != identity.ErrUnauthenticated {
		t.Fatal("local session survived provider logout")
	}
	u, _ := url.Parse(out.ProviderLogoutURL)
	if u.Path != "/user_management/sessions/logout" || u.Query().Get("session_id") != "session_pl1" ||
		u.Query().Get("return_to") != originA+"/sign-in" {
		t.Fatalf("logout URL %q", out.ProviderLogoutURL)
	}
	ref := out.LogoutRef
	if !refRe.MatchString(ref) || out.ReturnTo != "/sign-in" || strings.Contains(out.ProviderLogoutURL, ref) {
		t.Fatalf("reference or return: %+v", out)
	}
	if got, _ := h.svc.ProviderLogoutReturn(ctx, ref); got != identity.OutcomeProviderRevoked {
		t.Fatalf("return outcome %q", got)
	}
	if got, _ := h.svc.ProviderLogoutReturn(ctx, ref); got != identity.OutcomeLocalRevoked {
		t.Fatalf("reused reference claimed %q", got)
	}
	for _, forged := range []string{"", "x", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"} {
		if got, _ := h.svc.ProviderLogoutReturn(ctx, forged); got != identity.OutcomeLocalRevoked {
			t.Fatalf("forged reference %q claimed %q", forged, got)
		}
	}

	// E-LO-6: a failing provider revocation never redirects to the provider; local revocation stays committed.
	h.fake.FailRevoke(http.StatusInternalServerError)
	cred2 := h.mustSignIn(originA, grant("user_01PL", "session_pl2")).Credential
	info2 := h.session(cred2)
	out, err = h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: csrfOf(info2), Origin: originA, Credential: cred2})
	if err != nil || out.Outcome != identity.OutcomeProviderFailed || out.ProviderLogoutURL != "" || out.ReturnTo != "/sign-in" || !refRe.MatchString(out.LogoutRef) {
		t.Fatalf("failed revoke: %+v %v", out, err)
	}
	if _, err := h.svc.Session(ctx, cred2, originA); err != identity.ErrUnauthenticated {
		t.Fatal("local revocation lost when the provider failed")
	}
	if got, _ := h.svc.ProviderLogoutReturn(ctx, out.LogoutRef); got != identity.OutcomeProviderFailed {
		t.Fatalf("failed outcome %q", got)
	}
	// An expired reference returns the generic outcome.
	h.fake.FailRevoke(0)
	cred3 := h.mustSignIn(originA, grant("user_01PL", "session_pl3")).Credential
	info3 := h.session(cred3)
	out, _ = h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: csrfOf(info3), Origin: originA, Credential: cred3})
	h.clk.Advance(10*time.Minute + time.Second)
	if got, _ := h.svc.ProviderLogoutReturn(ctx, out.LogoutRef); got != identity.OutcomeLocalRevoked {
		t.Fatalf("expired reference claimed %q", got)
	}
	// Provider sign-out needs a live session.
	if _, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: csrfOf(info3), Origin: originA, Credential: cred3}); err != identity.ErrUnauthenticated {
		t.Fatalf("provider logout without a live session: %v", err)
	}
	// The second origin: the provider returns to that same origin's sign-in page (a host-only cookie on one origin
	// cannot serve a return to the other).
	cred4 := h.mustSignIn(originB, grant("user_01PL", "session_pl4")).Credential
	info4 := h.sessionAt(cred4, originB)
	out, err = h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: csrfOf(info4), Origin: originB, Credential: cred4})
	u4, _ := url.Parse(out.ProviderLogoutURL)
	if err != nil || out.Outcome != identity.OutcomeProviderRedirect || u4.Query().Get("return_to") != originB+"/sign-in" ||
		!refRe.MatchString(out.LogoutRef) || strings.Contains(out.ProviderLogoutURL, out.LogoutRef) {
		t.Fatalf("second-origin provider logout: %+v %v", out, err)
	}
	if got, _ := h.svc.ProviderLogoutReturn(ctx, out.LogoutRef); got != identity.OutcomeProviderRevoked {
		t.Fatalf("second-origin return outcome %q", got)
	}
	var stored int
	if err := setupConn(t).QueryRow(ctx, `SELECT count(*)::int FROM app.sessions WHERE provider_revocation IS NOT NULL AND revoke_reason = 'provider_logout'`).Scan(&stored); err != nil || stored != 4 {
		t.Fatalf("recorded provider outcomes %d (%v)", stored, err)
	}
}

// refRe is core's canonical encoding of a one-use reference: 43 unpadded base64url characters.
var refRe = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

// I-DISABLE: disable revokes every session and blocks sign-in; re-enable permits only new sessions.
func TestDisableAndEnable(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	c1 := h.mustSignIn(originA, grant("user_01DE", "session_de1")).Credential
	c2 := h.mustSignIn(originB, grant("user_01DE", "session_de2")).Credential
	info := h.session(c1)
	if err := h.svc.DisableUser(ctx, info.UserID); err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{c1, c2} {
		if _, err := h.svc.Session(ctx, c, originA); err != identity.ErrUnauthenticated {
			t.Fatalf("disabled user's session resolved: %v", err)
		}
	}
	if _, err := h.signIn(originA, "", grant("user_01DE", "session_de3")); err != identity.ErrUserDisabled {
		t.Fatalf("sign-in while disabled: %v", err)
	}
	if err := h.svc.EnableUser(ctx, info.UserID); err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{c1, c2} {
		if _, err := h.svc.Session(ctx, c, originA); err != identity.ErrUnauthenticated {
			t.Fatal("re-enable revived an old session")
		}
	}
	fresh := h.session(h.mustSignIn(originA, grant("user_01DE", "session_de4")).Credential)
	if fresh.UserID != info.UserID {
		t.Fatal("re-enabled user got a new UUID")
	}
	if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE revoke_reason = 'disabled'`); n != 2 {
		t.Fatalf("disabled revocations %d", n)
	}
	if err := h.svc.DisableUser(ctx, "00000000-0000-0000-0000-000000000000"); err != identity.ErrInvalidRequest {
		t.Fatalf("unknown user: %v", err)
	}
}
