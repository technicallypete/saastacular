//go:build integration

package identity_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"net/url"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
	"core/internal/identity/identitytest"
)

// Login start (Plan r5 §4.2): origin, existing session, configuration and namespace checks; the stored transaction.
func TestLoginStart(t *testing.T) {
	resetIdentity(t)
	// G062: the root origin is configured; every tenant-origin assertion below is unchanged by it.
	h := newHarness(t, withRoot(rootOrigin))
	ctx := context.Background()

	// G016: an unknown slug, the shared auth host (non-tenant) and an unrouted custom host are refused like any other.
	for _, origin := range []string{"", "https://attacker.localhost:543", "https://org-a.localhost", "http://org-a.localhost:543",
		authOrigin, "https://unrouted.example.com:543"} {
		if _, err := h.svc.Login(ctx, identity.LoginInput{Origin: origin}); err != identity.ErrOriginRejected {
			t.Errorf("origin %q: %v", origin, err)
		}
	}
	if n := count(t, `SELECT count(*)::int FROM app.login_transactions`); n != 0 {
		t.Fatalf("rejected starts stored %d transactions", n)
	}

	out, state, challenge := h.start(originA, "")
	if out.BindingMaxAge != 600 || len(out.Binding) != 43 || len(state) != 43 || len(challenge) != 43 {
		t.Fatalf("start output %+v state=%q challenge=%q", out, state, challenge)
	}
	u, _ := url.Parse(out.AuthorizationURL)
	if u.Query().Get("redirect_uri") != originA+"/auth/callback" || u.Query().Get("code_challenge_method") != "S256" || u.Query().Has("max_age") {
		t.Fatalf("authorization URL %q", out.AuthorizationURL)
	}
	// The binding is reused when valid; the stored transaction holds hashes, not the state or binding.
	again, _, _ := h.start(originA, out.Binding)
	if again.Binding != out.Binding {
		t.Fatal("a valid binding was not reused")
	}
	if fresh, _, _ := h.start(originA, "not-a-binding"); fresh.Binding == "not-a-binding" || len(fresh.Binding) != 43 {
		t.Fatal("an invalid binding was not replaced")
	}
	var hashed, leaks, ttlOK, ret int
	if err := setupConn(t).QueryRow(ctx, `
SELECT count(*) FILTER (WHERE state_hash = sha256(convert_to($1, 'UTF8')))::int,
       count(*) FILTER (WHERE state_hash = convert_to($1, 'UTF8') OR binding_hash = convert_to($2, 'UTF8')
                           OR pkce_verifier IN ($1, $2))::int,
       count(*) FILTER (WHERE expires_at = created_at + interval '10 minutes')::int,
       count(*) FILTER (WHERE return_path = '/account')::int
FROM app.login_transactions`, state, out.Binding).Scan(&hashed, &leaks, &ttlOK, &ret); err != nil {
		t.Fatal(err)
	}
	if hashed != 1 || leaks != 0 || ttlOK != 3 || ret != 3 {
		t.Fatalf("stored transactions: hashed=%d leaks=%d ttl=%d return=%d", hashed, leaks, ttlOK, ret)
	}

	// Freshness configured: max_age is sent.
	fresh := newHarness(t, withMaxAuthAge(10*time.Minute))
	resetIdentity(t)
	o, _, _ := fresh.start(originA, "")
	if u, _ := url.Parse(o.AuthorizationURL); u.Query().Get("max_age") != "600" {
		t.Fatalf("max_age missing: %q", o.AuthorizationURL)
	}

	// Already signed in: 409, no transaction.
	resetIdentity(t)
	cred := h.mustSignIn(originA, grant("user_01START", "session_start")).Credential
	if _, err := h.svc.Login(ctx, identity.LoginInput{Origin: originA, Credential: cred}); err != identity.ErrAlreadyAuthenticated {
		t.Fatalf("live session: %v", err)
	}
	if n := count(t, `SELECT count(*)::int FROM app.login_transactions`); n != 0 {
		t.Fatalf("already-signed-in start stored %d transactions", n)
	}
	// A dead credential is anonymous.
	if _, err := h.svc.Login(ctx, identity.LoginInput{Origin: originA, Credential: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}); err != nil {
		t.Fatalf("dead credential: %v", err)
	}

	// Unconfigured: provider operations are identity_unconfigured; session resolution still works.
	off := newHarness(t, unconfigured())
	if _, err := off.svc.Login(ctx, identity.LoginInput{Origin: originA}); err != identity.ErrUnconfigured {
		t.Fatalf("unconfigured login: %v", err)
	}
	if _, err := off.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: "code=x&state=y", Binding: out.Binding}); err != identity.ErrUnconfigured {
		t.Fatalf("unconfigured callback: %v", err)
	}
	if _, err := off.svc.Session(ctx, cred, originA); err != nil {
		t.Fatalf("unconfigured session resolution: %v", err)
	}
	// A real unconfigured deployment has no origins either: provider operations still answer identity_unconfigured
	// (web shows "unavailable"), never origin_rejected, and session resolution is unaffected.
	bare := newHarness(t, unconfigured(), func(c *identity.Config) { c.Origins = identity.NewOrigins(nil) })
	for _, origin := range []string{originA, "https://unlisted.localhost", ""} {
		if _, err := bare.svc.Login(ctx, identity.LoginInput{Origin: origin, Credential: cred}); err != identity.ErrUnconfigured {
			t.Fatalf("unconfigured login without origins (%q): %v", origin, err)
		}
		if _, err := bare.svc.Callback(ctx, identity.CallbackInput{Origin: origin, RawQuery: "code=x&state=y", Binding: out.Binding}); err != identity.ErrUnconfigured {
			t.Fatalf("unconfigured callback without origins (%q): %v", origin, err)
		}
	}
	if _, err := bare.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: "code=x&code=y"}); err != identity.ErrAmbiguousParameter {
		t.Fatalf("ambiguous query precedes configuration: %v", err)
	}
	if _, err := bare.svc.Session(ctx, cred, originA); err != nil {
		t.Fatalf("unconfigured session resolution without origins: %v", err)
	}
}

// A-1..A-4, A-6 (Plan r5 §8.2): every session's MFA assurance is unknown, whatever the method, labels or policy.
func TestAssuranceIsAlwaysUnknown(t *testing.T) {
	resetIdentity(t)
	// G062: the SUP-1 block below needs the root origin configured; the tenant-origin assertions are unchanged by it.
	h := newHarness(t, withRoot(rootOrigin))
	ctx := context.Background()
	variants := []identitytest.Grant{
		grant("user_01MFA", "session_a1"),                             // A-1: no MFA metadata
		{UserID: "user_01MFA", SID: "session_a2", Method: "Password"}, // A-2
		{UserID: "user_01MFA", SID: "session_a3", Claims: map[string]any{"amr": []string{"mfa", "otp"}, "acr": "mfa", "mfa": true}}, // A-3
		{UserID: "user_01MFA", SID: "session_a4", Method: "MagicAuth"},                                                              // A-4 policy drift
	}
	for i, g := range variants {
		out, state, challenge := h.start(originA, "")
		code := h.code(challenge, g)
		raw := query(code, state)
		if i == 2 {
			raw += "&mfa=verified&amr=mfa" // forged browser parameters are ignored
		}
		cb, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: raw, Binding: out.Binding})
		if err != nil {
			t.Fatalf("variant %d: %v", i, err)
		}
		if info := h.session(cb.Credential); info.MFAAssurance != identity.MFAAssuranceUnknown {
			t.Fatalf("variant %d: assurance %q", i, info.MFAAssurance)
		} else if identity.RequireVerifiedMFA(info) != identity.ErrMFAUnverified {
			t.Fatalf("variant %d: verified-MFA requirement not denied", i)
		}
	}
	if n := count(t, `SELECT count(DISTINCT mfa_assurance)::int FROM app.sessions`); n != 1 {
		t.Fatalf("assurance values: %d", n)
	}
	// A-6: the runtime role cannot store any other value.
	rt := runtimeConn(t)
	if _, err := rt.Exec(ctx, `UPDATE app.sessions SET mfa_assurance = 'verified'`); err == nil {
		t.Fatal("runtime stored mfa_assurance='verified'")
	}

	// G062 SUP-1 (Plan r5 §1.2, §4.2): the signup intent starts only on the exact configured root origin; www, a tenant,
	// the auth host, a custom host and any scheme/port/path variant are origin_rejected; combined with the enrollment
	// facts it is invalid_request; an unconfigured root refuses it; the root without the intent is not a tenant host.
	// Nothing is stored for a refusal.
	signupRows := func() int {
		return count(t, `SELECT count(*)::int FROM app.login_transactions WHERE signup_intent`)
	}
	noRoot := h.cfg
	noRoot.RootOrigin = ""
	unrooted := identity.NewService(identity.NewStore(it.pool), noRoot)
	for name, c := range map[string]struct {
		svc *identity.Service
		in  identity.LoginInput
		err error
	}{
		"www":                 {h.svc, identity.LoginInput{Origin: "https://www.localhost:543", SignupIntent: true}, identity.ErrOriginRejected},
		"tenant":              {h.svc, identity.LoginInput{Origin: originA, SignupIntent: true}, identity.ErrOriginRejected},
		"auth host":           {h.svc, identity.LoginInput{Origin: authOrigin, SignupIntent: true}, identity.ErrOriginRejected},
		"custom host":         {h.svc, identity.LoginInput{Origin: "https://custom.example:543", SignupIntent: true}, identity.ErrOriginRejected},
		"no port":             {h.svc, identity.LoginInput{Origin: "https://localhost", SignupIntent: true}, identity.ErrOriginRejected},
		"http":                {h.svc, identity.LoginInput{Origin: "http://localhost:543", SignupIntent: true}, identity.ErrOriginRejected},
		"other port":          {h.svc, identity.LoginInput{Origin: "https://localhost:544", SignupIntent: true}, identity.ErrOriginRejected},
		"trailing slash":      {h.svc, identity.LoginInput{Origin: rootOrigin + "/", SignupIntent: true}, identity.ErrOriginRejected},
		"enrollment facts":    {h.svc, identity.LoginInput{Origin: rootOrigin, SignupIntent: true, EnrollmentIntent: true}, identity.ErrInvalidRequest},
		"unconfigured root":   {unrooted, identity.LoginInput{Origin: rootOrigin, SignupIntent: true}, identity.ErrOriginRejected},
		"root without intent": {h.svc, identity.LoginInput{Origin: rootOrigin}, identity.ErrOriginRejected},
	} {
		if _, err := c.svc.Login(ctx, c.in); err != c.err {
			t.Errorf("signup start %s: %v, want %v", name, err, c.err)
		}
	}
	if n := signupRows(); n != 0 {
		t.Fatalf("refused signup starts stored %d transactions", n)
	}
	so, err := h.svc.Login(ctx, identity.LoginInput{Origin: rootOrigin, SignupIntent: true, ReturnTo: "/account"})
	if err != nil {
		t.Fatalf("root signup start: %v", err)
	}
	su, _ := url.Parse(so.AuthorizationURL)
	if su.Query().Get("redirect_uri") != rootOrigin+"/auth/callback" {
		t.Errorf("signup redirect_uri %q, want the root's own callback (no relay)", su.Query().Get("redirect_uri"))
	}
	var shape string
	if err := setupConn(t).QueryRow(ctx, `SELECT format('%s|%s|%s|%s|%s', origin, organization_id IS NULL, return_path, signup_handoff,
  enrollment_intent) FROM app.login_transactions WHERE signup_intent AND state_hash = sha256(convert_to($1, 'UTF8'))`,
		su.Query().Get("state")).Scan(&shape); err != nil || shape != rootOrigin+"|t|/signup/details|f|f" {
		t.Errorf("stored signup transaction %q %v, want the root origin, no Organization, /signup/details", shape, err)
	}
}

// --- G062 platform root signup (Plan r5 §1.2-§1.4, §5.2 SUP-2/SUP-8/SUP-10) -------------------------------------------

// rootOrigin is the test platform root (PLATFORM_BASE_DOMAIN localhost, port 543).
const rootOrigin = "https://localhost:543"

func withRoot(root string) option { return func(c *identity.Config) { c.RootOrigin = root } }

// signupStart begins a root signup-intent login and returns the output with the provider-visible state and challenge.
func (h *harness) signupStart(svc *identity.Service, binding string) (identity.LoginOutput, string, string) {
	h.t.Helper()
	out, err := svc.Login(context.Background(), identity.LoginInput{Origin: rootOrigin, Binding: binding, SignupIntent: true})
	if err != nil {
		h.t.Fatalf("signup start: %v", err)
	}
	u, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		h.t.Fatal(err)
	}
	return out, u.Query().Get("state"), u.Query().Get("code_challenge")
}

// signupResolves reports whether the runtime role's bootstrap reader app.auth_resolve_signup finds the credential digest
// now (unrevoked, unexpired, active user).
func signupResolves(t *testing.T, hash []byte) bool {
	t.Helper()
	var n int
	if err := runtimeConn(t).QueryRow(context.Background(), `SELECT count(*)::int FROM app.auth_resolve_signup($1, now())`, hash).Scan(&n); err != nil {
		t.Fatalf("auth_resolve_signup: %v", err)
	}
	return n == 1
}

// signupReason returns the signup row's revoke_reason ("" while unrevoked; setup side).
func signupReason(t *testing.T, hash []byte) string {
	t.Helper()
	var reason *string
	if err := setupConn(t).QueryRow(context.Background(), `SELECT revoke_reason FROM app.signup_sessions WHERE credential_hash = $1`, hash).Scan(&reason); err != nil {
		t.Fatalf("signup row: %v", err)
	}
	if reason == nil {
		return ""
	}
	return *reason
}

// seedSignup inserts one pending signup credential of user minted from provider session sid at the root (setup side)
// and returns its credential hash.
func seedSignup(t *testing.T, user, sid string) []byte {
	t.Helper()
	secret, csrf := make([]byte, 32), make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(csrf); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(secret)
	if _, err := setupConn(t).Exec(context.Background(), `INSERT INTO app.signup_sessions
  (id, user_id, origin, credential_hash, csrf_token, provider_session_id, expires_at)
VALUES (gen_random_uuid(), $1::text::uuid, $2, $3, $4, $5, now() + interval '20 minutes')`, user, rootOrigin, sum[:], csrf, sid); err != nil {
		t.Fatalf("seed signup: %v", err)
	}
	return sum[:]
}

// TestCallbackSignup (SUP-2): the root signup callback (T1) maps the exact (issuer, subject) — a new principal gets a new
// user and mapping, an existing one keeps its UUID and bindings, never merged by e-mail — and issues one signup
// credential and no session, with auth.sign_in signup_authenticated; cancel, expiry, replay, a foreign binding, a forged
// state, a disabled user and a changed root configuration create nothing.
func TestCallbackSignup(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t, withRoot(rootOrigin))
	ctx := context.Background()
	rows := func() int { return count(t, `SELECT count(*)::int FROM app.signup_sessions`) }
	sessions := func() int { return count(t, `SELECT count(*)::int FROM app.sessions`) }
	userOf := func(subject string) string {
		t.Helper()
		var uid string
		if err := setupConn(t).QueryRow(ctx, `SELECT user_id::text FROM app.identity_principals WHERE issuer = $1 AND subject = $2`,
			h.fake.Issuer, subject).Scan(&uid); err != nil {
			t.Fatalf("mapping of %s: %v", subject, err)
		}
		return uid
	}
	callback := func(svc *identity.Service, out identity.LoginOutput, raw string) (identity.CallbackOutput, error) {
		return svc.Callback(ctx, identity.CallbackInput{Origin: rootOrigin, RawQuery: raw, Binding: out.Binding})
	}

	t.Run("new principal", func(t *testing.T) {
		out, state, challenge := h.signupStart(h.svc, "")
		code := h.fake.IssueCode(challenge, identitytest.Grant{UserID: "user_01SIGNEW", SID: "session_signew", FirstName: "Synthetic", LastName: "User", Email: "shared@example.test"})
		cb, err := callback(h.svc, out, query(code, state))
		if err != nil || cb.Outcome != identity.OutcomeSignup || cb.ReturnTo != "/signup/details" || cb.Credential != "" || cb.Signup == nil ||
			cb.Signup.Credential == "" || cb.MaxAge != 1800 {
			t.Fatalf("signup callback: %+v %v", cb, err)
		}
		uid := userOf("user_01SIGNEW")
		info, err := h.svc.SignupCredential(ctx, cb.Signup.Credential, rootOrigin)
		if err != nil || info.Handle.UserID() != uid || info.OrganizationID != "" || info.TargetOrigin != "" {
			t.Fatalf("signup credential: %+v %v", info, err)
		}
		var sid string
		if err := setupConn(t).QueryRow(ctx, `SELECT provider_session_id FROM app.signup_sessions WHERE user_id = $1::text::uuid AND origin = $2`,
			uid, rootOrigin).Scan(&sid); err != nil || sid != "session_signew" {
			t.Errorf("signup row: %q %v", sid, err)
		}
		if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE user_id = $1::text::uuid`, uid); n != 0 {
			t.Errorf("the signup callback issued %d sessions", n)
		}
		if n := count(t, `SELECT count(*)::int FROM app.role_bindings WHERE subject_user_id = $1::text::uuid`, uid); n != 0 {
			t.Errorf("the signup callback granted %d bindings", n)
		}
		if n := count(t, `SELECT count(*)::int FROM audit.events WHERE action = 'auth.sign_in' AND outcome = 'success'
  AND reason = 'signup_authenticated' AND (actor_user_id = $1::text::uuid OR subject_user_id = $1::text::uuid)`, uid); n != 1 {
			t.Errorf("auth.sign_in signup_authenticated rows: %d, want 1", n)
		}
		// The signup credential is not a session credential.
		if _, err := h.svc.Session(ctx, cb.Signup.Credential, rootOrigin); err != identity.ErrUnauthenticated {
			t.Errorf("signup credential as a session: %v", err)
		}
		// Replay of the consumed transaction: nothing more.
		before := rows()
		if _, err := callback(h.svc, out, query(code, state)); err != identity.ErrTransactionInvalid || rows() != before {
			t.Errorf("replay: %v (%d -> %d rows)", err, before, rows())
		}
		// No e-mail linking: another subject with the same e-mail is another user.
		out2, state2, challenge2 := h.signupStart(h.svc, "")
		code2 := h.fake.IssueCode(challenge2, identitytest.Grant{UserID: "user_01SIGTWIN", SID: "session_sigtwin", FirstName: "Synthetic", LastName: "User", Email: "shared@example.test"})
		if _, err := callback(h.svc, out2, query(code2, state2)); err != nil {
			t.Fatal(err)
		}
		if userOf("user_01SIGTWIN") == uid {
			t.Error("two subjects with one e-mail were merged")
		}
	})

	t.Run("existing principal keeps its user and bindings", func(t *testing.T) {
		uid := h.provision("user_01SIGOLD")
		var bindings string
		const bindingState = `SELECT coalesce(md5(string_agg(b::text, '|' ORDER BY b.id)), '-') FROM app.role_bindings b WHERE subject_user_id = $1::text::uuid`
		if err := setupConn(t).QueryRow(ctx, bindingState, uid).Scan(&bindings); err != nil {
			t.Fatal(err)
		}
		out, state, challenge := h.signupStart(h.svc, "")
		cb, err := callback(h.svc, out, query(h.fake.IssueCode(challenge, grant("user_01SIGOLD", "session_sigold")), state))
		if err != nil || cb.Signup == nil {
			t.Fatalf("existing principal: %+v %v", cb, err)
		}
		info, err := h.svc.SignupCredential(ctx, cb.Signup.Credential, rootOrigin)
		if err != nil || info.Handle.UserID() != uid {
			t.Fatalf("existing principal credential: %+v %v (want user %s)", info, err, uid)
		}
		var after string
		if err := setupConn(t).QueryRow(ctx, bindingState, uid).Scan(&after); err != nil || after != bindings {
			t.Errorf("the signup callback changed the existing user's bindings (%v)", err)
		}
		if n := count(t, `SELECT count(*)::int FROM app.identity_principals WHERE user_id = $1::text::uuid`, uid); n != 1 {
			t.Errorf("principals of the existing user: %d", n)
		}
	})

	t.Run("refusals create nothing", func(t *testing.T) {
		users0, rows0, sessions0 := count(t, `SELECT count(*)::int FROM app.users`), rows(), sessions()
		// cancel: consumed, nothing created
		out, state, _ := h.signupStart(h.svc, "")
		if _, err := callback(h.svc, out, "error=access_denied&state="+state); err != identity.ErrLoginCanceled {
			t.Errorf("cancel: %v", err)
		}
		if _, err := callback(h.svc, out, "error=access_denied&state="+state); err != identity.ErrTransactionInvalid {
			t.Errorf("cancel did not consume: %v", err)
		}
		// a foreign binding and a forged state consume nothing
		out2, state2, challenge2 := h.signupStart(h.svc, "")
		other, _, _ := h.signupStart(h.svc, "")
		code2 := h.fake.IssueCode(challenge2, grant("user_01SIGREF", "session_sigref"))
		if _, err := callback(h.svc, other, query(code2, state2)); err != identity.ErrTransactionInvalid {
			t.Errorf("foreign binding: %v", err)
		}
		if _, err := callback(h.svc, out2, query(code2, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")); err != identity.ErrTransactionInvalid {
			t.Errorf("forged state: %v", err)
		}
		// expiry: deleted and unusable
		out3, state3, challenge3 := h.signupStart(h.svc, "")
		code3 := h.fake.IssueCode(challenge3, grant("user_01SIGLATE", "session_siglate"))
		h.clk.Advance(10*time.Minute + time.Second)
		if _, err := callback(h.svc, out3, query(code3, state3)); err != identity.ErrTransactionExpired {
			t.Errorf("expired: %v", err)
		}
		if got := []int{count(t, `SELECT count(*)::int FROM app.users`), rows(), sessions()}; !slices.Equal(got, []int{users0, rows0, sessions0}) {
			t.Fatalf("refused signup callbacks changed users/rows/sessions: %v -> %v", []int{users0, rows0, sessions0}, got)
		}
		// a changed root configuration at the callback: consumed, signup_origin_rejected, nothing created
		out4, state4, challenge4 := h.signupStart(h.svc, "")
		moved := h.cfg
		moved.RootOrigin = "https://other.localhost:543"
		code4 := h.fake.IssueCode(challenge4, grant("user_01SIGMOVE", "session_sigmove"))
		if _, err := callback(identity.NewService(identity.NewStore(it.pool), moved), out4, query(code4, state4)); err != identity.ErrSignupOriginRejected {
			t.Errorf("changed root: %v", err)
		}
		if _, err := callback(h.svc, out4, query(code4, state4)); err != identity.ErrTransactionInvalid {
			t.Errorf("changed root did not consume: %v", err)
		}
		// a disabled existing user: user_disabled, no signup row
		uid := h.provision("user_01SIGDIS")
		if err := h.svc.DisableUser(ctx, uid); err != nil {
			t.Fatal(err)
		}
		out5, state5, challenge5 := h.signupStart(h.svc, "")
		if _, err := callback(h.svc, out5, query(h.fake.IssueCode(challenge5, grant("user_01SIGDIS", "session_sigdis")), state5)); err != identity.ErrUserDisabled {
			t.Errorf("disabled user: %v", err)
		}
		if n := count(t, `SELECT count(*)::int FROM app.signup_sessions WHERE user_id = $1::text::uuid`, uid); n != 0 {
			t.Errorf("a disabled user got %d signup rows", n)
		}
		if got := rows(); got != rows0 {
			t.Errorf("signup rows %d -> %d after the refusals", rows0, got)
		}
	})
}

// pinnedFixture is one provisioned signup row of a provisioned user whose continuation is ready for completion (R
// consumed, C and N stored, N consumed, P stored) for a fresh slugged Organization the user belongs to.
type pinnedFixture struct {
	user, org, slug, target string
	challenge, proof        []byte
}

func newPinnedFixture(t *testing.T, h *harness, subject string) pinnedFixture {
	t.Helper()
	ctx := context.Background()
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	f := pinnedFixture{slug: fmt.Sprintf("pin-%x", b)}
	f.target = "https://" + f.slug + ".localhost:543"
	setup := setupConn(t)
	// G057: the target host is the Organization's platform label (f.slug, a synthetic fixture label); its friendly slug is
	// distinct metadata that never routes.
	if err := setup.QueryRow(ctx, `INSERT INTO app.organizations (id, name, slug, platform_label) VALUES (gen_random_uuid(), 'Pinned', $1, $2)
  RETURNING id::text`, "meta-"+f.slug, f.slug).Scan(&f.org); err != nil {
		t.Fatalf("pinned Organization: %v", err)
	}
	f.user = h.provision(subject, f.org)
	c, p := sha256.Sum256([]byte("challenge/"+f.slug)), sha256.Sum256([]byte("proof/"+f.slug))
	f.challenge, f.proof = c[:], p[:]
	cred, csrf, intent, ref, nonce := sha256.Sum256([]byte("cred/"+f.slug)), sha256.Sum256([]byte("csrf/"+f.slug)),
		sha256.Sum256([]byte("intent/"+f.slug)), sha256.Sum256([]byte("ref/"+f.slug)), sha256.Sum256([]byte("nonce/"+f.slug))
	now := h.clk.Now()
	if _, err := setup.Exec(ctx, `INSERT INTO app.signup_sessions (id, user_id, origin, credential_hash, csrf_token, provider_session_id,
  expires_at, intent_hash, organization_id, provisioned_at, target_origin, continuation_generation, continuation_expires_at, ref_hash,
  ref_consumed_at, challenge_hash, nonce_hash, nonce_consumed_at, proof_hash)
VALUES (gen_random_uuid(), $1::text::uuid, $2, $3, $4, 'session_pinned', now() + interval '25 minutes', $5, $6::text::uuid, $7, $8, 1,
  $7::timestamptz + interval '110 seconds', $9, $7, $10, $11, $7, $12)`, f.user, rootOrigin, cred[:], csrf[:], intent[:], f.org, now, f.target,
		ref[:], f.challenge, nonce[:], f.proof); err != nil {
		t.Fatalf("pinned signup row: %v", err)
	}
	return f
}

// TestCallbackPinned (SUP-8, identity part; Plan r5 §1.4 step 4): the pinned tenant login completes only for its expected
// user — with a new provider session after interaction (the not-live path as the callback observes it) the outcome is
// tenant_created with return_to "/" and a session bound to the tenant origin and Organization; another user is
// signup_user_mismatch (consumed, no session, no mapping change); a cancel is login_canceled with the tenant intact; a
// disabled user is user_disabled; a membership removed before the callback is organization_not_member; a replay is
// transaction_invalid.
func TestCallbackPinned(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t, withRoot(rootOrigin))
	ctx := context.Background()
	pinned := func(f pinnedFixture) identity.LoginOutput {
		t.Helper()
		out, err := h.svc.LoginPinned(ctx, identity.PinnedLoginInput{Origin: f.target, Challenge: f.challenge, Proof: f.proof})
		if err != nil {
			t.Fatalf("pinned login: %v", err)
		}
		return out
	}
	finish := func(f pinnedFixture, out identity.LoginOutput, g identitytest.Grant) (identity.CallbackOutput, error) {
		u, err := url.Parse(out.AuthorizationURL)
		if err != nil {
			t.Fatal(err)
		}
		return h.svc.Callback(ctx, identity.CallbackInput{Origin: f.target, Binding: out.Binding,
			RawQuery: query(h.fake.IssueCode(u.Query().Get("code_challenge"), g), u.Query().Get("state"))})
	}
	stateOf := func(out identity.LoginOutput) string {
		u, _ := url.Parse(out.AuthorizationURL)
		return u.Query().Get("state")
	}

	t.Run("expected user after interaction", func(t *testing.T) {
		f := newPinnedFixture(t, h, "user_01PINOK")
		out := pinned(f)
		var shape string
		if err := setupConn(t).QueryRow(ctx, `SELECT format('%s|%s|%s|%s|%s', signup_handoff, expected_user_id, organization_id, return_path, signup_intent)
FROM app.login_transactions WHERE state_hash = sha256(convert_to($1, 'UTF8'))`, stateOf(out)).Scan(&shape); err != nil ||
			shape != "t|"+f.user+"|"+f.org+"|/|f" {
			t.Fatalf("pinned row %q %v", shape, err)
		}
		// The provider session was not live: the user authenticated again (a new provider session, a fresh auth_time).
		g := grant("user_01PINOK", "session_pin_new")
		g.AuthTime = h.clk.Now()
		cb, err := finish(f, out, g)
		if err != nil || cb.Outcome != identity.OutcomeTenantCreated || cb.ReturnTo != "/" || cb.Credential == "" {
			t.Fatalf("pinned callback: %+v %v", cb, err)
		}
		info := h.sessionAt(cb.Credential, f.target)
		if info.UserID != f.user || info.OrganizationID != f.org || info.ProviderSessionID != "session_pin_new" {
			t.Errorf("pinned session %+v", info)
		}
		if _, err := h.svc.Session(ctx, cb.Credential, originA); err != identity.ErrUnauthenticated {
			t.Errorf("the pinned session resolved at another tenant: %v", err)
		}
		// auth.sign_in is an identity event (no Organization): attributed to the signed-in user.
		if n := count(t, `SELECT count(*)::int FROM audit.events WHERE action = 'auth.sign_in' AND outcome = 'success' AND reason = 'tenant_created'
  AND (actor_user_id = $1::text::uuid OR subject_user_id = $1::text::uuid)`, f.user); n != 1 {
			t.Errorf("auth.sign_in tenant_created rows: %d, want 1", n)
		}
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: f.target, Binding: out.Binding, RawQuery: "code=x&state=" + stateOf(out)}); err != identity.ErrTransactionInvalid {
			t.Errorf("replay: %v", err)
		}
	})

	t.Run("another user is signup_user_mismatch", func(t *testing.T) {
		f := newPinnedFixture(t, h, "user_01PINOWN")
		other := h.provision("user_01PINOTHER", f.org) // an existing member of the same Organization
		var before string
		const otherState = `SELECT coalesce(md5(string_agg(x, '|' ORDER BY x)), '-') FROM (
  SELECT b::text AS x FROM app.role_bindings b WHERE subject_user_id = $1::text::uuid
  UNION ALL SELECT p::text FROM app.identity_principals p WHERE user_id = $1::text::uuid) s`
		if err := setupConn(t).QueryRow(ctx, otherState, other).Scan(&before); err != nil {
			t.Fatal(err)
		}
		mismatches := func() int {
			return count(t, `SELECT count(*)::int FROM audit.events WHERE action = 'auth.sign_in' AND reason = 'signup_user_mismatch'`)
		}
		m0 := mismatches()
		// an existing member of the same Organization, then an unknown principal (each at its own pinned login)
		f2 := newPinnedFixture(t, h, "user_01PINOWN2")
		for _, c := range []struct {
			f pinnedFixture
			g identitytest.Grant
		}{{f, grant("user_01PINOTHER", "session_pin_other")}, {f2, grant("user_01PINFRESH", "session_pin_fresh")}} {
			o := pinned(c.f)
			cb, err := finish(c.f, o, c.g)
			if err != identity.ErrSignupUserMismatch || cb.Credential != "" {
				t.Errorf("%s at the pinned callback: %+v %v", c.g.UserID, cb, err)
			}
			if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: c.f.target, Binding: o.Binding, RawQuery: "code=x&state=" + stateOf(o)}); err != identity.ErrTransactionInvalid {
				t.Errorf("%s: the mismatch did not consume: %v", c.g.UserID, err)
			}
		}
		var after string
		if err := setupConn(t).QueryRow(ctx, otherState, other).Scan(&after); err != nil || after != before {
			t.Errorf("the mismatch changed the other user's bindings or mapping (%v)", err)
		}
		if n := count(t, `SELECT count(*)::int FROM app.identity_principals WHERE subject = 'user_01PINFRESH'`); n != 0 {
			t.Errorf("the mismatch mapped an unknown principal (%d)", n)
		}
		if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE user_id = $1::text::uuid`, other); n != 0 {
			t.Errorf("the mismatch issued a session to the other user (%d)", n)
		}
		if n := mismatches() - m0; n != 2 {
			t.Errorf("new auth.sign_in signup_user_mismatch rows: %d, want 2", n)
		}
	})

	t.Run("cancel keeps the tenant", func(t *testing.T) {
		f := newPinnedFixture(t, h, "user_01PINCAN")
		out := pinned(f)
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: f.target, Binding: out.Binding,
			RawQuery: "error=access_denied&state=" + stateOf(out)}); err != identity.ErrLoginCanceled {
			t.Fatalf("cancel: %v", err)
		}
		if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE user_id = $1::text::uuid`, f.user); n != 0 {
			t.Errorf("cancel issued %d sessions", n)
		}
		if n := count(t, `SELECT count(*)::int FROM app.role_bindings WHERE organization_id = $1::text::uuid AND subject_user_id = $2::text::uuid`,
			f.org, f.user); n != 1 {
			t.Errorf("cancel changed the membership (%d)", n)
		}
		// a later ordinary sign-in on the tenant host succeeds
		if cb, err := h.signIn(f.target, "", grant("user_01PINCAN", "session_pin_later")); err != nil || cb.Outcome != identity.OutcomeSignedIn {
			t.Errorf("ordinary sign-in after the cancel: %+v %v", cb, err)
		}
	})

	t.Run("disabled user and removed membership", func(t *testing.T) {
		f := newPinnedFixture(t, h, "user_01PINDIS")
		out := pinned(f)
		if err := h.svc.DisableUser(ctx, f.user); err != nil {
			t.Fatal(err)
		}
		if _, err := finish(f, out, grant("user_01PINDIS", "session_pin_dis")); err != identity.ErrUserDisabled {
			t.Errorf("disabled user: %v", err)
		}
		g := newPinnedFixture(t, h, "user_01PINREM")
		outG := pinned(g)
		// the membership is removed without its capture (superuser, triggers off): the fixture only needs the binding gone
		tx, err := setupConn(t).Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM app.role_bindings WHERE organization_id = $1::text::uuid`, g.org); err != nil {
			t.Fatalf("remove membership: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err := finish(g, outG, grant("user_01PINREM", "session_pin_rem")); err != identity.ErrNotMember {
			t.Errorf("removed membership: %v", err)
		}
		if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE user_id IN ($1::text::uuid, $2::text::uuid)`, f.user, g.user); n != 0 {
			t.Errorf("refused pinned callbacks issued %d sessions", n)
		}
	})
}

// A-7, A-8, U-PARAM, I-TXN: cancellation, invalid identity, duplicate parameters, replay, expiry, wrong binding or
// origin, parallel isolation — with the recorded transaction effect, not only the error.
func TestCallbackParametersAndTransactions(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	txns := func() int { return count(t, `SELECT count(*)::int FROM app.login_transactions`) }
	sessions := func() int { return count(t, `SELECT count(*)::int FROM app.sessions`) }

	// U-PARAM: duplicates are ambiguous and consume nothing.
	out, state, challenge := h.start(originA, "")
	code := h.code(challenge, grant("user_01P", "session_p"))
	for _, raw := range []string{
		query(code, state) + "&code=" + code, query(code, state) + "&state=" + state,
		query(code, state) + "&error=access_denied&error=x", query(code, state) + "&error_description=a&error_description=b",
		"code=a;state=b",
	} {
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: raw, Binding: out.Binding}); err != identity.ErrAmbiguousParameter {
			t.Errorf("%q: %v", raw, err)
		}
	}
	if txns() != 1 {
		t.Fatal("ambiguous callbacks consumed the transaction")
	}

	// Wrong or missing binding, wrong origin, unknown state: not consumed.
	other, _, _ := h.start(originA, "")
	for name, in := range map[string]identity.CallbackInput{
		"other browser binding": {Origin: originA, RawQuery: query(code, state), Binding: other.Binding},
		"no binding":            {Origin: originA, RawQuery: query(code, state)},
		"other origin":          {Origin: originB, RawQuery: query(code, state), Binding: out.Binding},
		"unknown state":         {Origin: originA, RawQuery: query(code, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), Binding: out.Binding},
		"no state":              {Origin: originA, RawQuery: "code=" + code, Binding: out.Binding},
	} {
		if _, err := h.svc.Callback(ctx, in); err != identity.ErrTransactionInvalid {
			t.Errorf("%s: %v", name, err)
		}
	}
	if txns() != 2 || sessions() != 0 {
		t.Fatalf("transactions %d sessions %d after non-consuming failures", txns(), sessions())
	}
	// The original transaction still completes, once.
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(code, state), Binding: out.Binding}); err != nil {
		t.Fatalf("valid callback after failures: %v", err)
	}
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(code, state), Binding: out.Binding}); err != identity.ErrTransactionInvalid {
		t.Fatalf("replay: %v", err)
	}

	// A-7: cancellation consumes and issues no session.
	before := sessions()
	o2, s2, _ := h.start(originA, "")
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: "error=access_denied&error_description=canceled&state=" + s2, Binding: o2.Binding}); err != identity.ErrLoginCanceled {
		t.Fatalf("cancel: %v", err)
	}
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: "error=access_denied&state=" + s2, Binding: o2.Binding}); err != identity.ErrTransactionInvalid {
		t.Fatalf("cancel did not consume: %v", err)
	}
	o3, s3, _ := h.start(originA, "")
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: "error=server_error&state=" + s3, Binding: o3.Binding}); err != identity.ErrProviderError {
		t.Fatalf("provider error: %v", err)
	}

	// A-8: an invalid signed identity is rejected after consumption.
	o4, s4, c4 := h.start(originA, "")
	bad := h.code(c4, identitytest.Grant{UserID: "user_01BAD", SID: "s", Claims: map[string]any{"client_id": "client_01OTHER"}})
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(bad, s4), Binding: o4.Binding}); err != identity.ErrProviderError {
		t.Fatalf("invalid identity: %v", err)
	}
	// PKCE at L1: the provider rejects a code bound to another transaction's challenge (code injection).
	o5, s5, _ := h.start(originA, "")
	_, _, victimChallenge := h.start(originA, "")
	injected := h.code(victimChallenge, grant("user_01VICTIM", "session_victim"))
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(injected, s5), Binding: o5.Binding}); err != identity.ErrCodeRejected {
		t.Fatalf("injected code: %v", err)
	}
	if sessions() != before {
		t.Fatalf("failed callbacks created sessions: %d -> %d", before, sessions())
	}

	// Expiry: an expired transaction is deleted and unusable.
	o6, s6, c6 := h.start(originA, "")
	late := h.code(c6, grant("user_01LATE", "session_late"))
	h.clk.Advance(10*time.Minute + time.Second)
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(late, s6), Binding: o6.Binding}); err != identity.ErrTransactionExpired {
		t.Fatalf("expired: %v", err)
	}
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(late, s6), Binding: o6.Binding}); err != identity.ErrTransactionInvalid {
		t.Fatalf("expired transaction still present: %v", err)
	}

	// Parallel transactions stay isolated: completing one leaves the other usable.
	pa, sa, ca := h.start(originA, "")
	pb, sb, cb := h.start(originA, pa.Binding)
	codeB := h.code(cb, grant("user_01ISO", "session_iso_b"))
	codeA := h.code(ca, grant("user_01ISO", "session_iso_a"))
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(codeB, sb), Binding: pb.Binding}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(codeA, sa), Binding: pa.Binding}); err != nil {
		t.Fatalf("second parallel transaction: %v", err)
	}
}

// Continuity (Plan r5 §4.2 table; E-CB-9, E-CB-10, E-CB-1 at L1).
func TestCallbackContinuity(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	a := h.mustSignIn(originA, grant("user_01CONTA", "session_S"))
	aInfo := h.session(a.Credential)

	// Same provider session (SSO returned S): A unchanged, already_signed_in, no revoke.
	same, err := h.signIn(originA, a.Credential, grant("user_01CONTA", "session_S"))
	if err != nil || same.Outcome != identity.OutcomeAlreadySignedIn || same.Credential != "" {
		t.Fatalf("same sid: %+v %v", same, err)
	}
	if _, err := h.svc.Session(ctx, a.Credential, originA); err != nil || len(h.fake.Revoked()) != 0 {
		t.Fatalf("A changed or a revoke happened: %v %v", err, h.fake.Revoked())
	}

	// Additional provider session S′ for A: old session superseded, new session on S′, no revoke.
	rot, err := h.signIn(originA, a.Credential, grant("user_01CONTA", "session_S2"))
	if err != nil || rot.Outcome != identity.OutcomeSessionRotated || rot.Credential == "" {
		t.Fatalf("new sid: %+v %v", rot, err)
	}
	if _, err := h.svc.Session(ctx, a.Credential, originA); err != identity.ErrUnauthenticated {
		t.Fatalf("old session survived rotation: %v", err)
	}
	newInfo := h.session(rot.Credential)
	if newInfo.UserID != aInfo.UserID || newInfo.ProviderSessionID != "session_S2" || len(h.fake.Revoked()) != 0 {
		t.Fatalf("rotation: %+v revoked %v", newInfo, h.fake.Revoked())
	}
	if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE revoke_reason = 'superseded'`); n != 1 {
		t.Fatalf("superseded count %d", n)
	}

	// A different principal B while A is live: A kept, identity_conflict, S_B revoked once, and nothing of B's changes.
	// G016 D-2(a): B's account is provisioned (an operator's act) before B can sign in at all, so the conflict evidence
	// is that B's user, principal and memberships are identical afterwards and B holds no session.
	h.provision("user_01CONTB")
	const bState = `SELECT coalesce(md5(string_agg(x, '|' ORDER BY x)), '-') FROM (
  SELECT u::text AS x FROM app.users u JOIN app.identity_principals p ON p.user_id = u.id WHERE p.subject = 'user_01CONTB'
  UNION ALL SELECT p::text FROM app.identity_principals p WHERE p.subject = 'user_01CONTB'
  UNION ALL SELECT b::text FROM app.role_bindings b JOIN app.identity_principals p ON p.user_id = b.subject_user_id
    WHERE p.subject = 'user_01CONTB') s`
	var bBefore, bAfter string
	if err := setupConn(t).QueryRow(ctx, bState).Scan(&bBefore); err != nil {
		t.Fatal(err)
	}
	if _, err := h.signIn(originA, rot.Credential, grant("user_01CONTB", "session_SB")); err != identity.ErrIdentityConflict {
		t.Fatalf("conflict: %v", err)
	}
	if _, err := h.svc.Session(ctx, rot.Credential, originA); err != nil {
		t.Fatalf("A lost on conflict: %v", err)
	}
	if err := setupConn(t).QueryRow(ctx, bState).Scan(&bAfter); err != nil || bAfter != bBefore {
		t.Fatalf("conflict changed B's account: %v", err)
	}
	if n := count(t, `SELECT count(*)::int FROM app.sessions s JOIN app.identity_principals p ON p.user_id = s.user_id
WHERE p.subject = 'user_01CONTB'`); n != 0 {
		t.Fatal("conflict created a session for B")
	}
	if got := h.fake.Revoked(); !slices.Equal(got, []string{"session_SB"}) {
		t.Fatalf("revoked %v", got)
	}
	// A dead credential is anonymous: B signs in normally.
	if out, err := h.signIn(originA, a.Credential, grant("user_01CONTB", "session_SB2")); err != nil || out.Outcome != identity.OutcomeSignedIn {
		t.Fatalf("dead credential: %+v %v", out, err)
	}
}

// Freshness, impersonation and disabled users (Plan r5 §3.3, §4.5).
func TestCallbackRefusals(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t, withMaxAuthAge(5*time.Minute))
	ctx := context.Background()
	stale := grant("user_01STALE", "session_stale")
	stale.AuthTime = h.clk.Now().Add(-6 * time.Minute)
	if _, err := h.signIn(originA, "", stale); err != identity.ErrStaleAuthentication {
		t.Fatalf("stale auth_time: %v", err)
	}
	fresh := grant("user_01STALE", "session_fresh")
	fresh.AuthTime = h.clk.Now().Add(-4 * time.Minute)
	if _, err := h.signIn(originA, "", fresh); err != nil {
		t.Fatalf("fresh auth_time: %v", err)
	}
	imp := grant("user_01IMP", "session_imp")
	imp.Impersonator = true
	if _, err := h.signIn(originA, "", imp); err != identity.ErrImpersonationRefused {
		t.Fatalf("impersonation: %v", err)
	}

	resetIdentity(t)
	plain := newHarness(t)
	info := plain.session(plain.mustSignIn(originA, grant("user_01OFF", "session_off1")).Credential)
	if err := plain.svc.DisableUser(ctx, info.UserID); err != nil {
		t.Fatal(err)
	}
	if _, err := plain.signIn(originA, "", grant("user_01OFF", "session_off2")); err != identity.ErrUserDisabled {
		t.Fatalf("disabled user: %v", err)
	}
	if n := count(t, `SELECT count(*)::int FROM app.users WHERE status = 'disabled'`); n != 1 {
		t.Fatal("sign-in changed the disabled status")
	}
}

// I-NS-1..3 (Plan r5 §3.6): one bound provider namespace; a changed client or issuer blocks provider operations without
// minting, issuing or consuming anything, and existing sessions stay valid.
func TestNamespaceBinding(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	cred := h.mustSignIn(originA, grant("user_01NS", "session_ns")).Credential
	out, _, _ := h.start(originA, "")
	txnsBefore := count(t, `SELECT count(*)::int FROM app.login_transactions`)

	for name, opts := range map[string][]option{
		"same issuer, changed client (I-NS-1)": {withIssuer(h.fake.Issuer), withClient("client_01OTHERENV0000000000000")},
		"changed issuer, same client (I-NS-2)": {withIssuer("https://api.workos.com/other"), withClient(itClient)},
	} {
		mis := newHarness(t, opts...)
		if _, err := mis.svc.Login(ctx, identity.LoginInput{Origin: originA}); err != identity.ErrNamespaceMismatch {
			t.Errorf("%s login: %v", name, err)
		}
		if _, err := mis.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: "code=x&state=y", Binding: out.Binding}); err != identity.ErrNamespaceMismatch {
			t.Errorf("%s callback: %v", name, err)
		}
		if _, err := mis.svc.Session(ctx, cred, originA); err != nil {
			t.Errorf("%s: existing session invalidated: %v", name, err)
		}
	}
	if n := count(t, `SELECT count(*)::int FROM app.login_transactions`); n != txnsBefore {
		t.Fatalf("mismatch consumed or created transactions: %d -> %d", txnsBefore, n)
	}
	if n := count(t, `SELECT count(*)::int FROM app.identity_provider_binding`); n != 1 {
		t.Fatalf("binding rows %d", n)
	}

	// I-NS-3: the first-ever binding under 8 concurrent first sign-ins is exactly one row.
	resetIdentity(t)
	race := newHarness(t)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = race.signIn(originA, "", grant(fmt.Sprintf("user_01NSR%d", i), fmt.Sprintf("session_nsr%d", i)))
		}()
	}
	wg.Wait()
	if n := count(t, `SELECT count(*)::int FROM app.identity_provider_binding`); n != 1 {
		t.Fatalf("binding rows after the race: %d", n)
	}
	if n := count(t, `SELECT count(*)::int FROM app.users`); n != 8 {
		t.Fatalf("users after the race: %d", n)
	}
}

// E-ATTR at L1 (AC2): the exact (issuer, subject) is the key; names and emails are attributes. Equal emails never link;
// a rename keeps the UUID and updates the display name.
func TestMappingIsByIssuerAndSubject(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	g1 := identitytest.Grant{UserID: "user_01ATTRA", SID: "s1", FirstName: "Same", LastName: "Name", Email: "same@example.test"}
	g2 := identitytest.Grant{UserID: "user_01ATTRB", SID: "s2", FirstName: "Same", LastName: "Name", Email: "same@example.test"}
	u1 := h.session(h.mustSignIn(originA, g1).Credential)
	u2 := h.sessionAt(h.mustSignIn(originB, g2).Credential, originB)
	if u1.UserID == u2.UserID {
		t.Fatal("equal email linked two principals")
	}
	g1.FirstName, g1.Email, g1.SID = "Renamed", "new@example.test", "s3"
	renamed := h.sessionAt(h.mustSignIn(originB, g1).Credential, originB) // another origin, same principal
	if renamed.UserID != u1.UserID || renamed.DisplayName != "Renamed Name" {
		t.Fatalf("rename: %+v", renamed)
	}
	if n := count(t, `SELECT count(*)::int FROM app.users WHERE display_name IS NOT NULL AND display_name LIKE '%@%'`); n != 0 {
		t.Fatal("an email reached the users table")
	}
}

// --- ENR-19 (G058 Plan r3 §6.2, D-18; readiness correction 01 §3, D-11): pending enrollment credentials are revoked at
// the revoking path's actual scope -----------------------------------------------------------------------------------

// seedEnrollment inserts one pending enrollment credential of user in org-a, minted from provider session sid (setup
// side, the attested superuser), and returns its credential hash.
func seedEnrollment(t *testing.T, user, sid string) []byte {
	t.Helper()
	secret, csrf := make([]byte, 32), make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	if _, err := rand.Read(csrf); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(secret)
	if _, err := setupConn(t).Exec(context.Background(), `INSERT INTO app.enrollment_sessions
  (id, user_id, organization_id, origin, credential_hash, csrf_token, provider_session_id, expires_at)
VALUES (gen_random_uuid(), $1::text::uuid, $2::text::uuid, $3, $4, $5, $6, now() + interval '20 minutes')`,
		user, orgAID, originA, sum[:], csrf, sid); err != nil {
		t.Fatalf("seed enrollment: %v", err)
	}
	return sum[:]
}

// enrollResolves reports whether the runtime role's bootstrap reader app.auth_resolve_enrollment finds the credential
// now (unrevoked, unexpired, active user).
func enrollResolves(t *testing.T, hash []byte) bool {
	t.Helper()
	var n int
	if err := runtimeConn(t).QueryRow(context.Background(), `SELECT count(*)::int FROM app.auth_resolve_enrollment($1, now())`, hash).Scan(&n); err != nil {
		t.Fatalf("auth_resolve_enrollment: %v", err)
	}
	return n == 1
}

// enrollReason returns the credential's revoke_reason ("" while unrevoked; setup side).
func enrollReason(t *testing.T, hash []byte) string {
	t.Helper()
	var reason *string
	if err := setupConn(t).QueryRow(context.Background(), `SELECT revoke_reason FROM app.enrollment_sessions WHERE credential_hash = $1`, hash).Scan(&reason); err != nil {
		t.Fatalf("enrollment row: %v", err)
	}
	if reason == nil {
		return ""
	}
	return *reason
}

// enrollFixture is one user with two live local sessions (sid1 on org-a, sid2 on org-b) and one pending enrollment
// credential minted from each provider session.
type enrollFixture struct {
	h          *harness
	user       string
	c1, c2     string
	i1         identity.SessionInfo
	sid1, sid2 string
	e1, e2     []byte
	// G062 SUP-10: one pending root signup credential minted from each provider session.
	s1, s2 []byte
}

func newEnrollFixture(t *testing.T, subject string) enrollFixture {
	t.Helper()
	resetIdentity(t)
	h := newHarness(t)
	f := enrollFixture{h: h, sid1: "session_" + subject + "_1", sid2: "session_" + subject + "_2"}
	f.c1 = h.mustSignIn(originA, grant(subject, f.sid1)).Credential
	f.c2 = h.mustSignIn(originB, grant(subject, f.sid2)).Credential
	f.i1 = h.session(f.c1)
	if f.i1.ProviderSessionID != f.sid1 || h.sessionAt(f.c2, originB).UserID != f.i1.UserID {
		t.Fatalf("fixture sessions: %+v", f.i1)
	}
	f.user = f.i1.UserID
	f.e1, f.e2 = seedEnrollment(t, f.user, f.sid1), seedEnrollment(t, f.user, f.sid2)
	if !enrollResolves(t, f.e1) || !enrollResolves(t, f.e2) {
		t.Fatal("fixture enrollment credentials do not resolve")
	}
	f.s1, f.s2 = seedSignup(t, f.user, f.sid1), seedSignup(t, f.user, f.sid2)
	if !signupResolves(t, f.s1) || !signupResolves(t, f.s2) {
		t.Fatal("fixture signup credentials do not resolve")
	}
	return f
}

// signupBoundOnly asserts that only the sid1 signup credential is revoked, with reason, and the sid2 one still resolves.
func (f enrollFixture) signupBoundOnly(t *testing.T, reason string) {
	t.Helper()
	if got := signupReason(t, f.s1); got != reason || signupResolves(t, f.s1) {
		t.Errorf("sid1 signup credential: reason %q (want %q), resolves %v", got, reason, signupResolves(t, f.s1))
	}
	if got := signupReason(t, f.s2); got != "" || !signupResolves(t, f.s2) {
		t.Errorf("sid2 signup credential touched by a bound path: reason %q, resolves %v", got, signupResolves(t, f.s2))
	}
}

// signupBoth asserts that both signup credentials are revoked with reason and neither resolves.
func (f enrollFixture) signupBoth(t *testing.T, reason string) {
	t.Helper()
	for i, s := range [][]byte{f.s1, f.s2} {
		if got := signupReason(t, s); got != reason || signupResolves(t, s) {
			t.Errorf("signup credential %d: reason %q (want %q), resolves %v", i+1, got, reason, signupResolves(t, s))
		}
	}
}

// boundOnly asserts that only the sid1 credential is revoked, with reason, and the sid2 credential still resolves.
func (f enrollFixture) boundOnly(t *testing.T, reason string) {
	t.Helper()
	if got := enrollReason(t, f.e1); got != reason || enrollResolves(t, f.e1) {
		t.Errorf("sid1 credential: reason %q (want %q), resolves %v", got, reason, enrollResolves(t, f.e1))
	}
	if got := enrollReason(t, f.e2); got != "" || !enrollResolves(t, f.e2) {
		t.Errorf("sid2 credential touched by a bound path: reason %q, resolves %v", got, enrollResolves(t, f.e2))
	}
}

// both asserts that both credentials are revoked with reason and neither resolves.
func (f enrollFixture) both(t *testing.T, reason string) {
	t.Helper()
	for i, e := range [][]byte{f.e1, f.e2} {
		if got := enrollReason(t, e); got != reason || enrollResolves(t, e) {
			t.Errorf("credential %d: reason %q (want %q), resolves %v", i+1, got, reason, enrollResolves(t, e))
		}
	}
}

func TestEnrollmentPendingCredentialRevocation(t *testing.T) {
	ctx := context.Background()

	t.Run("logout of the sid1 session revokes only the sid1 credential", func(t *testing.T) {
		f := newEnrollFixture(t, "user_01E19LO")
		if _, err := f.h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: csrfOf(f.i1), Origin: originA, Credential: f.c1}); err != nil {
			t.Fatal(err)
		}
		f.boundOnly(t, "logout")
		f.signupBoundOnly(t, "logout")
	})

	t.Run("provider logout of the sid1 session revokes only the sid1 credential", func(t *testing.T) {
		f := newEnrollFixture(t, "user_01E19PL")
		if _, err := f.h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: csrfOf(f.i1), Origin: originA, Credential: f.c1}); err != nil {
			t.Fatal(err)
		}
		f.boundOnly(t, "provider_logout")
		f.signupBoundOnly(t, "provider_logout")
	})

	t.Run("supersession of the sid1 session revokes only the sid1 credential", func(t *testing.T) {
		f := newEnrollFixture(t, "user_01E19SU")
		rot, err := f.h.signIn(originA, f.c1, grant("user_01E19SU", "session_user_01E19SU_3"))
		if err != nil || rot.Outcome != identity.OutcomeSessionRotated {
			t.Fatalf("rotation: %+v %v", rot, err)
		}
		f.boundOnly(t, "superseded")
		f.signupBoundOnly(t, "superseded")
	})

	t.Run("logout-all revokes both credentials", func(t *testing.T) {
		f := newEnrollFixture(t, "user_01E19LA")
		if _, err := f.h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: csrfOf(f.i1), Origin: originA, Credential: f.c1}); err != nil {
			t.Fatal(err)
		}
		f.both(t, "logout_all")
		f.signupBoth(t, "logout_all")
	})

	t.Run("disable revokes both credentials and re-enable revives neither", func(t *testing.T) {
		f := newEnrollFixture(t, "user_01E19DI")
		if err := f.h.svc.DisableUser(ctx, f.user); err != nil {
			t.Fatal(err)
		}
		f.both(t, "disabled")
		f.signupBoth(t, "disabled")
		if err := f.h.svc.EnableUser(ctx, f.user); err != nil {
			t.Fatal(err)
		}
		f.both(t, "disabled")
		f.signupBoth(t, "disabled")
	})

	t.Run("retire revokes both credentials and re-enable revives neither", func(t *testing.T) {
		f := newEnrollFixture(t, "user_01E19RE")
		if err := f.h.svc.FinalizeRetirement(ctx, f.user, func(context.Context, pgx.Tx) error { return nil }); err != nil {
			t.Fatal(err)
		}
		f.both(t, "disabled")
		f.signupBoth(t, "disabled")
		if err := f.h.svc.EnableUser(ctx, f.user); err != nil {
			t.Fatal(err)
		}
		// Not merely the disabled-user refusal: the user is active again and the old digests still never resolve, while a
		// credential minted after the re-enable does.
		f.both(t, "disabled")
		f.signupBoth(t, "disabled")
		if fresh := seedEnrollment(t, f.user, "session_user_01E19RE_3"); !enrollResolves(t, fresh) {
			t.Fatal("a credential minted after the re-enable does not resolve")
		}
		if fresh := seedSignup(t, f.user, "session_user_01E19RE_3"); !signupResolves(t, fresh) {
			t.Fatal("a signup credential minted after the re-enable does not resolve")
		}
	})

	t.Run("an admitted credential revoked after admission keeps admitted_at", func(t *testing.T) {
		f := newEnrollFixture(t, "user_01E19AD")
		setup := setupConn(t)
		if _, err := setup.Exec(ctx, `UPDATE app.enrollment_sessions SET admitted_at = now(), admitted_binding_id = (
  SELECT id FROM app.role_bindings WHERE organization_id = $2::text::uuid AND subject_user_id = $3::text::uuid AND source = 'membership')
WHERE credential_hash = $1`, f.e1, orgAID, f.user); err != nil {
			t.Fatalf("admit fixture: %v", err)
		}
		var before time.Time
		if err := setup.QueryRow(ctx, `SELECT admitted_at FROM app.enrollment_sessions WHERE credential_hash = $1`, f.e1).Scan(&before); err != nil {
			t.Fatal(err)
		}
		if _, err := f.h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: csrfOf(f.i1), Origin: originA, Credential: f.c1}); err != nil {
			t.Fatal(err)
		}
		f.both(t, "logout_all")
		var after *time.Time
		if err := setup.QueryRow(ctx, `SELECT admitted_at FROM app.enrollment_sessions WHERE credential_hash = $1`, f.e1).Scan(&after); err != nil {
			t.Fatal(err)
		}
		if after == nil || !after.Equal(before) {
			t.Fatalf("admitted_at %v -> %v", before, after)
		}
	})

	t.Run("login start removes an expired enrollment credential only", func(t *testing.T) {
		f := newEnrollFixture(t, "user_01E19EX")
		// ENR-21e: an enrollment credential is not a local session credential. Its digest resolves through
		// auth_resolve_enrollment (positive control) and through auth_resolve_session on its own origin to no row.
		var asSession int
		if err := runtimeConn(t).QueryRow(ctx, `SELECT count(*)::int FROM app.auth_resolve_session($1, now(), $2)`, f.e1, originA).Scan(&asSession); err != nil ||
			asSession != 0 || !enrollResolves(t, f.e1) {
			t.Fatalf("enrollment credential as a session: %d rows %v (want 0), resolves as enrollment %v (want true)", asSession, err, enrollResolves(t, f.e1))
		}
		if _, err := setupConn(t).Exec(ctx, `UPDATE app.enrollment_sessions SET created_at = now() - interval '1 hour',
  expires_at = now() - interval '50 minutes' WHERE credential_hash = $1`, f.e1); err != nil {
			t.Fatalf("expire fixture: %v", err)
		}
		// ENR-21b: the expired, unrevoked row (still present until the sweep below) resolves nothing.
		if enrollReason(t, f.e1) != "" || enrollResolves(t, f.e1) {
			t.Fatal("an expired, unrevoked enrollment credential resolves")
		}
		// G062 SUP-10: an expired signup credential is swept by the same login start; a live one stays.
		if _, err := setupConn(t).Exec(ctx, `UPDATE app.signup_sessions SET created_at = now() - interval '1 hour',
  expires_at = now() - interval '50 minutes' WHERE credential_hash = $1`, f.s1); err != nil {
			t.Fatalf("expire signup fixture: %v", err)
		}
		if signupResolves(t, f.s1) {
			t.Fatal("an expired, unrevoked signup credential resolves")
		}
		f.h.start(originA, "")
		if n := count(t, `SELECT count(*)::int FROM app.enrollment_sessions WHERE credential_hash = $1`, f.e1); n != 0 {
			t.Fatal("the expired enrollment credential survived login start")
		}
		if !enrollResolves(t, f.e2) {
			t.Fatal("login start removed or revoked a live enrollment credential")
		}
		if n := count(t, `SELECT count(*)::int FROM app.signup_sessions WHERE credential_hash = $1`, f.s1); n != 0 {
			t.Fatal("the expired signup credential survived login start")
		}
		if !signupResolves(t, f.s2) {
			t.Fatal("login start removed or revoked a live signup credential")
		}
	})
}
