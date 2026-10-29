//go:build integration

package identity_test

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"
	"testing"

	"core/internal/identity"
)

// TestCustomDomainRelay (B-IDENTITY-IT; G016 Plan r4 §1.7, §1.7.1, §6.2): the B2 shared callback relay and the custom-host
// consumption, with the pre-exchange checks and the final issuance guard, against the real app.hostname_state reader. The
// fixture custom host portal.example.com:543 is an injected row (no DNS); org-a/org-b are platform labels (G057 D-2).
const (
	customHost   = "portal.example.com"
	customOrigin = "https://portal.example.com:543"
)

// hookProvider is the fixture provider with an action run inside Exchange, after the code is redeemed and before the
// identity returns: a change committed there lands between the pre-exchange checks and the issuance transaction (cases A
// and B) with no production test seam.
type hookProvider struct {
	identity.Provider
	onExchange func()
}

func (p hookProvider) Exchange(ctx context.Context, in identity.Exchange) (identity.VerifiedIdentity, error) {
	vi, err := p.Provider.Exchange(ctx, in)
	if p.onExchange != nil {
		p.onExchange()
	}
	return vi, err
}

// seedCustom inserts a routing-ready (active, fresh proof, ready TLS) custom domain row for org created by user.
func seedCustom(t *testing.T, org, user string) string {
	t.Helper()
	var id string
	if err := setupConn(t).QueryRow(context.Background(), `INSERT INTO app.organization_domains
  (id, organization_id, hostname, generation, state, challenge_value, challenge_expires_at, verified_at, verified_until,
   tls_state, tls_checked_at, tls_not_after, created_by)
SELECT gen_random_uuid(), $1::text::uuid, $2, coalesce((SELECT max(generation) FROM app.organization_domains WHERE hostname = $2), 0) + 1,
       'active', repeat('a', 52), now() + interval '1 hour', now(), now() + interval '30 days', 'ready', now(),
       now() + interval '60 days', $3::text::uuid
RETURNING id::text`, org, customHost, user).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// removeCustom removes the live row (the lifecycle's removal shape: withdrawal starts).
func removeCustom(t *testing.T, id string) {
	t.Helper()
	if _, err := setupConn(t).Exec(context.Background(), `UPDATE app.organization_domains
SET state = 'removed', removed_at = now(), tls_state = 'withdrawing', version = version + 1 WHERE id = $1::text::uuid`, id); err != nil {
		t.Fatal(err)
	}
}

func b64(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }

// withAuth returns cfg with another shared relay origin ("" disables the relay).
func withAuth(cfg identity.Config, auth string) identity.Config {
	cfg.AuthOrigin = auth
	return cfg
}

func deniedSignIns(t *testing.T, reason string) int {
	t.Helper()
	return count(t, `SELECT count(*)::int FROM audit.events WHERE action = 'auth.sign_in' AND outcome = 'denied' AND reason = $1`, reason)
}

// relayStart begins a login on the custom host and returns the binding, the provider-visible state and challenge, after
// checking the authorization request names the shared relay callback.
func relayStart(t *testing.T, h *harness) (identity.LoginOutput, string, string) {
	t.Helper()
	out, state, challenge := h.start(customOrigin, "")
	u, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	if got := u.Query().Get("redirect_uri"); got != authOrigin+"/auth/callback" {
		t.Fatalf("custom-host redirect_uri %q, want the relay's", got)
	}
	return out, state, challenge
}

func TestCustomDomainRelay(t *testing.T) {
	resetIdentity(t)
	h := newHarness(t)
	ctx := context.Background()
	owner := h.provision("user_01RELAYOWNER")
	domainID := seedCustom(t, orgAID, owner)

	t.Run("relay answers only a routing-ready custom host", func(t *testing.T) {
		_, state, _ := relayStart(t, h)
		cb, err := h.svc.Relay(ctx, state)
		if err != nil || cb != customOrigin+"/auth/callback" {
			t.Fatalf("relay: %q %v", cb, err)
		}
		random := state[:strings.LastIndex(state, ".")]
		for name, st := range map[string]string{
			"untagged":       random,
			"platform host":  random + "." + b64(originA),
			"auth host":      random + "." + b64(authOrigin),
			"unknown custom": random + "." + b64("https://other.example.com:543"),
			"http":           random + "." + b64("http://portal.example.com:543"),
			"oversized":      random + "." + strings.Repeat("A", 600),
			"empty":          "",
		} {
			if cb, err := h.svc.Relay(ctx, st); err != identity.ErrRelayRefused || cb != "" {
				t.Errorf("%s: %q %v", name, cb, err)
			}
		}
		// The relay consumed nothing: the original transaction is still there.
		if n := count(t, `SELECT count(*)::int FROM app.login_transactions WHERE callback_origin = $1`, authOrigin); n < 1 {
			t.Fatal("relay consumed the transaction")
		}
		// An unconfigured relay (no auth origin) refuses everything, and a custom-host login has no callback then.
		off := identity.NewService(identity.NewStore(it.pool), withAuth(h.cfg, ""))
		if _, err := off.Relay(ctx, state); err != identity.ErrRelayRefused {
			t.Errorf("unconfigured relay: %v", err)
		}
		if _, err := off.Login(ctx, identity.LoginInput{Origin: customOrigin}); err != identity.ErrOriginRejected {
			t.Errorf("custom login without the relay: %v", err)
		}
		logs := h.logs.String()
		for _, want := range []string{`"msg":"auth.relay","outcome":"relayed"`, `"reason":"state_invalid"`, `"reason":"not_custom_host"`} {
			if !strings.Contains(logs, want) {
				t.Errorf("relay log lacks %s", want)
			}
		}
		if strings.Contains(logs, state) || strings.Contains(logs, b64(customOrigin)) || strings.Contains(logs, customHost) {
			t.Error("relay log carries the state, its tag or the hostname")
		}
	})

	t.Run("consumption on the custom host with its binding; replay, provider error and repeated code", func(t *testing.T) {
		out, state, challenge := relayStart(t, h)
		code := h.code(challenge, grant("user_01RELAYA", "session_relay_a"))
		raw := query(code, state)
		cb, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: raw, Binding: out.Binding})
		if err != nil || cb.Credential == "" {
			t.Fatalf("custom-host callback: %+v %v", cb, err)
		}
		host, err := h.svc.SessionHost(ctx, customOrigin)
		if err != nil || host.Kind != identity.HostCustom || host.OrganizationID != orgAID {
			t.Fatalf("session host: %+v %v", host, err)
		}
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: raw, Binding: out.Binding}); err != identity.ErrTransactionInvalid {
			t.Fatalf("replay: %v", err)
		}
		o2, s2, _ := relayStart(t, h)
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: "error=access_denied&error_description=x&state=" + url.QueryEscape(s2), Binding: o2.Binding}); err != identity.ErrLoginCanceled {
			t.Fatalf("relayed provider refusal: %v", err)
		}
		o3, s3, c3 := relayStart(t, h)
		c := h.code(c3, grant("user_01RELAYA", "session_relay_a2"))
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(c, s3) + "&code=" + c, Binding: o3.Binding}); err != identity.ErrAmbiguousParameter {
			t.Fatalf("repeated code: %v", err)
		}
	})

	t.Run("foreign browser fails the binding and consumes nothing", func(t *testing.T) {
		out, state, challenge := relayStart(t, h)
		other, _, _ := h.start(customOrigin, "")
		code := h.code(challenge, grant("user_01RELAYA", "session_relay_f"))
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: other.Binding}); err != identity.ErrTransactionInvalid {
			t.Fatalf("foreign binding: %v", err)
		}
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: out.Binding}); err != nil {
			t.Fatalf("the owner's browser after the foreign attempt: %v", err)
		}
	})

	t.Run("altered state or another origin fails the hash", func(t *testing.T) {
		out, state, challenge := relayStart(t, h)
		code := h.code(challenge, grant("user_01RELAYA", "session_relay_s"))
		altered := state[:strings.LastIndex(state, ".")] + "." + b64("https://other.example.com:543")
		for name, in := range map[string]identity.CallbackInput{
			"altered tag":     {Origin: customOrigin, RawQuery: query(code, altered), Binding: out.Binding},
			"untagged":        {Origin: customOrigin, RawQuery: query(code, state[:strings.LastIndex(state, ".")]), Binding: out.Binding},
			"platform origin": {Origin: originA, RawQuery: query(code, state), Binding: out.Binding},
		} {
			if _, err := h.svc.Callback(ctx, in); err != identity.ErrTransactionInvalid {
				t.Errorf("%s: %v", name, err)
			}
		}
	})

	t.Run("callback origin mismatch is host_changed after consumption", func(t *testing.T) {
		out, state, challenge := relayStart(t, h)
		code := h.code(challenge, grant("user_01RELAYA", "session_relay_m"))
		moved := identity.NewService(identity.NewStore(it.pool), withAuth(h.cfg, "https://auth2.localhost:543"))
		hits := h.fake.AuthHits()
		if _, err := moved.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: out.Binding}); err != identity.ErrHostChanged {
			t.Fatalf("mismatch: %v", err)
		}
		if h.fake.AuthHits() != hits {
			t.Fatal("an exchange happened")
		}
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: out.Binding}); err != identity.ErrTransactionInvalid {
			t.Fatalf("not consumed: %v", err)
		}
	})

	t.Run("membership denial creates nothing", func(t *testing.T) {
		h.provision("user_01RELAYB", orgBID) // a member of org-b only
		before := count(t, `SELECT count(*)::int FROM app.sessions`)
		denied := deniedSignIns(t, "organization_not_member")
		out, state, challenge := relayStart(t, h)
		code := h.fake.IssueCode(challenge, grant("user_01RELAYB", "session_relay_b"))
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: out.Binding}); err != identity.ErrNotMember {
			t.Fatalf("non-member: %v", err)
		}
		// An unmapped principal is never a member: no user, mapping or session is created.
		users := count(t, `SELECT count(*)::int FROM app.users`)
		o2, s2, c2 := h.start(originA, "")
		c := h.fake.IssueCode(c2, grant("user_01RELAYNEW", "session_relay_new"))
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(c, s2), Binding: o2.Binding}); err != identity.ErrNotMember {
			t.Fatalf("unmapped principal: %v", err)
		}
		if count(t, `SELECT count(*)::int FROM app.sessions`) != before || count(t, `SELECT count(*)::int FROM app.users`) != users ||
			count(t, `SELECT count(*)::int FROM app.identity_principals WHERE subject = 'user_01RELAYNEW'`) != 0 {
			t.Fatal("a refused sign-in created rows")
		}
		if got := deniedSignIns(t, "organization_not_member"); got != denied+2 {
			t.Fatalf("denied events %d -> %d", denied, got)
		}
	})

	// Case A (Plan r4 §1.7.1): the domain is removed while the provider exchange runs; the final guard in the issuance
	// transaction refuses (host_changed): no session, one denied event, the transaction consumed.
	t.Run("A removal during the exchange is host_changed at issuance", func(t *testing.T) {
		hp := hookProvider{Provider: h.cfg.Provider}
		cfg := h.cfg
		cfg.Provider = &hp
		svc := identity.NewService(identity.NewStore(it.pool), cfg)
		out, state, challenge := relayStart(t, h)
		code := h.code(challenge, grant("user_01RELAYA", "session_relay_ca"))
		hp.onExchange = func() { removeCustom(t, domainID) }
		sessions, denied := count(t, `SELECT count(*)::int FROM app.sessions`), deniedSignIns(t, "host_changed")
		if _, err := svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: out.Binding}); err != identity.ErrHostChanged {
			t.Fatalf("removal during exchange: %v", err)
		}
		if count(t, `SELECT count(*)::int FROM app.sessions`) != sessions || deniedSignIns(t, "host_changed") != denied+1 {
			t.Fatal("a session was issued or the denial was not recorded once")
		}
		if _, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: out.Binding}); err != identity.ErrTransactionInvalid {
			t.Fatalf("transaction not consumed: %v", err)
		}
		domainID = seedCustom(t, orgAID, owner) // a fresh claim generation for the next cases
	})

	// Case B: the membership is revoked while the exchange runs: organization_not_member at issuance.
	t.Run("B membership revocation during the exchange is organization_not_member", func(t *testing.T) {
		uid := h.provision("user_01RELAYREV")
		hp := hookProvider{Provider: h.cfg.Provider}
		cfg := h.cfg
		cfg.Provider = &hp
		svc := identity.NewService(identity.NewStore(it.pool), cfg)
		out, state, challenge := relayStart(t, h)
		code := h.fake.IssueCode(challenge, grant("user_01RELAYREV", "session_relay_cb"))
		hp.onExchange = func() {
			if _, err := setupConn(t).Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1::text::uuid AND organization_id = $2::text::uuid`, uid, orgAID); err != nil {
				t.Error(err)
			}
		}
		sessions, denied := count(t, `SELECT count(*)::int FROM app.sessions`), deniedSignIns(t, "organization_not_member")
		if _, err := svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: out.Binding}); err != identity.ErrNotMember {
			t.Fatalf("revocation during exchange: %v", err)
		}
		if count(t, `SELECT count(*)::int FROM app.sessions`) != sessions || deniedSignIns(t, "organization_not_member") != denied+1 {
			t.Fatal("a session was issued or the denial was not recorded once")
		}
	})

	// Case C: sign in, then remove the domain: the next session resolution for that origin is unknown_host.
	t.Run("C removal after sign-in makes the next session request unknown_host", func(t *testing.T) {
		out, state, challenge := relayStart(t, h)
		code := h.code(challenge, grant("user_01RELAYA", "session_relay_cc"))
		cb, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: out.Binding})
		if err != nil {
			t.Fatal(err)
		}
		removeCustom(t, domainID)
		if _, err := h.svc.SessionHost(ctx, customOrigin); err != identity.ErrUnknownHost {
			t.Fatalf("session host after removal: %v", err)
		}
		if _, err := h.svc.Session(ctx, cb.Credential, customOrigin); err != nil {
			t.Fatalf("the session itself is not revoked by G016 (D-4(a)): %v", err)
		}
		if _, err := h.svc.Login(ctx, identity.LoginInput{Origin: customOrigin}); err != identity.ErrOriginRejected {
			t.Fatalf("login on a removed host: %v", err)
		}
	})

	// Case D: the host is reassigned to org-b; a continuity sign-in (live session of the same user) on it is refused before
	// the supersession, so the live session stays untouched.
	t.Run("D continuity on a reassigned host leaves the live session untouched", func(t *testing.T) {
		domainID = seedCustom(t, orgAID, owner)
		out, state, challenge := relayStart(t, h)
		code := h.code(challenge, grant("user_01RELAYD", "session_relay_d1"))
		live, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(code, state), Binding: out.Binding})
		if err != nil {
			t.Fatal(err)
		}
		hp := hookProvider{Provider: h.cfg.Provider}
		cfg := h.cfg
		cfg.Provider = &hp
		svc := identity.NewService(identity.NewStore(it.pool), cfg)
		o2, s2, c2 := relayStart(t, h)
		c := h.code(c2, grant("user_01RELAYD", "session_relay_d2"))
		hp.onExchange = func() {
			removeCustom(t, domainID)
			seedCustom(t, orgBID, owner)
		}
		if _, err := svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(c, s2), Binding: o2.Binding, Credential: live.Credential}); err != identity.ErrHostChanged {
			t.Fatalf("continuity on a reassigned host: %v", err)
		}
		if _, err := h.svc.Session(ctx, live.Credential, customOrigin); err != nil {
			t.Fatalf("the live session was superseded: %v", err)
		}
		if n := count(t, `SELECT count(*)::int FROM app.sessions WHERE provider_session_id = 'session_relay_d2'`); n != 0 {
			t.Fatalf("%d sessions issued", n)
		}
		host, err := h.svc.SessionHost(ctx, customOrigin)
		if err != nil || host.OrganizationID != orgBID {
			t.Fatalf("reassigned host: %+v %v", host, err)
		}

		// G016 security amendment (Plan amendments A1/A2), on the same fixture: the portal session of case D is bound to
		// (customOrigin, org-a); the host now routes org-b.
		liveRow := func() int {
			t.Helper()
			return count(t, `SELECT count(*)::int FROM app.sessions WHERE provider_session_id = 'session_relay_d1' AND revoked_at IS NULL`)
		}

		// E: replayed at another tenant origin the portal credential is no session: Session is unauthenticated, a
		// continuity callback there issues a fresh session and leaves the portal row untouched, sign-out everywhere is
		// unauthenticated and revokes nothing.
		t.Run("E replay at another origin", func(t *testing.T) {
			if _, err := h.svc.Session(ctx, live.Credential, originA); err != identity.ErrUnauthenticated {
				t.Fatalf("portal credential at org-a: %v", err)
			}
			o, s, c := h.start(originA, "")
			cb, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: originA, RawQuery: query(h.code(c, grant("user_01RELAYD", "session_relay_e1")), s),
				Binding: o.Binding, Credential: live.Credential})
			if err != nil || cb.Credential == "" || cb.Credential == live.Credential {
				t.Fatalf("callback at org-a with the portal credential: %+v %v", cb.Outcome, err)
			}
			if info := h.sessionAt(cb.Credential, originA); info.Origin != originA || info.OrganizationID != orgAID {
				t.Fatalf("fresh org-a session binding: %q %q", info.Origin, info.OrganizationID)
			}
			if _, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: "x", Origin: originA, Credential: live.Credential}); err != identity.ErrUnauthenticated {
				t.Fatalf("logout all at org-a with the portal credential: %v", err)
			}
			if n := liveRow(); n != 1 {
				t.Fatalf("the portal session changed (%d live rows)", n)
			}
		})

		// F: a browser row without a binding (a pre-12 legacy row) is no session at any origin, and does not short-cut a
		// sign-in.
		t.Run("F missing binding", func(t *testing.T) {
			raw := make([]byte, 32)
			for i := range raw {
				raw[i] = byte(0xf0 + i%16)
			}
			legacy := base64.RawURLEncoding.EncodeToString(raw)
			if _, err := setupConn(t).Exec(ctx, `INSERT INTO app.sessions (user_id, credential_hash, csrf_token, auth_time, idle_expires_at,
  absolute_expires_at) VALUES ($1::text::uuid, sha256($2::bytea), sha256($2::bytea), now(), now() + interval '1 hour', now() + interval '1 day')`,
				owner, raw); err != nil {
				t.Fatal(err)
			}
			for _, origin := range []string{originA, originB, customOrigin} {
				if _, err := h.svc.Session(ctx, legacy, origin); err != identity.ErrUnauthenticated {
					t.Errorf("unbound row at %s: %v", origin, err)
				}
			}
			if _, err := h.svc.Login(ctx, identity.LoginInput{Origin: originA, Credential: legacy}); err != nil {
				t.Errorf("login with an unbound row: %v, want a new login", err)
			}
		})

		// G: on the reassigned host the old Organization's cookie (same origin) still resolves to its org-a binding, which
		// the origin's current Organization (org-b) refuses: Login starts a new sign-in instead of already_authenticated,
		// and the new sign-in issues an org-b-bound session, which is then already_authenticated there.
		var bCred string
		t.Run("G reassignment", func(t *testing.T) {
			info, err := h.svc.Session(ctx, live.Credential, customOrigin)
			if err != nil || info.OrganizationID != orgAID {
				t.Fatalf("old cookie binding on the reassigned host: %q %v", info.OrganizationID, err)
			}
			if _, err := h.svc.Login(ctx, identity.LoginInput{Origin: customOrigin, Credential: live.Credential}); err != nil {
				t.Fatalf("login on the reassigned host with the old cookie: %v, want a new login", err)
			}
			o, s, c := relayStart(t, h)
			cb, err := h.svc.Callback(ctx, identity.CallbackInput{Origin: customOrigin, RawQuery: query(h.code(c, grant("user_01RELAYG", "session_relay_g1")), s),
				Binding: o.Binding, Credential: live.Credential})
			if err != nil || cb.Credential == "" {
				t.Fatalf("sign-in on the reassigned host: %+v %v", cb.Outcome, err)
			}
			bCred = cb.Credential
			if info := h.sessionAt(bCred, customOrigin); info.OrganizationID != orgBID {
				t.Fatalf("new session binding %q, want org-b", info.OrganizationID)
			}
			if _, err := h.svc.Login(ctx, identity.LoginInput{Origin: customOrigin, Credential: bCred}); err != identity.ErrAlreadyAuthenticated {
				t.Fatalf("login with the org-b session: %v", err)
			}
			if n := liveRow(); n != 1 {
				t.Fatalf("the old portal session changed (%d live rows)", n)
			}
		})

		// L-1: after the host stops routing, sign-out still revokes: current is local_revoked, provider revokes locally and
		// at the provider and answers without a hosted sign-out redirect.
		t.Run("L-1 sign-out after removal", func(t *testing.T) {
			if bCred == "" {
				t.Skip("G did not issue the org-b session")
			}
			p := h.mustSignInRelay(t, "user_01RELAYG", "session_relay_l1")
			var activeID string
			if err := setupConn(t).QueryRow(ctx, `SELECT id::text FROM app.organization_domains WHERE hostname = $1 AND state = 'active'`, customHost).Scan(&activeID); err != nil {
				t.Fatal(err)
			}
			bInfo, pInfo := h.sessionAt(bCred, customOrigin), h.sessionAt(p, customOrigin)
			removeCustom(t, activeID)
			out, err := h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: csrfOf(bInfo), Origin: customOrigin, Credential: bCred})
			if err != nil || out.Outcome != identity.OutcomeLocalRevoked {
				t.Fatalf("current sign-out on a removed host: %+v %v", out, err)
			}
			if _, err := h.svc.Session(ctx, bCred, customOrigin); err != identity.ErrUnauthenticated {
				t.Fatalf("the session survived sign-out: %v", err)
			}
			out, err = h.svc.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: csrfOf(pInfo), Origin: customOrigin, Credential: p})
			if err != nil || out.Outcome != identity.OutcomeProviderRevoked || out.ProviderLogoutURL != "" || out.ReturnTo != "/sign-in" || out.LogoutRef == "" {
				t.Fatalf("provider sign-out on a removed host: %+v %v", out, err)
			}
			if _, err := h.svc.Session(ctx, p, customOrigin); err != identity.ErrUnauthenticated {
				t.Fatalf("the session survived provider sign-out: %v", err)
			}
		})
	})
}

// mustSignInRelay completes one relay sign-in of sub on the custom host and returns the credential.
func (h *harness) mustSignInRelay(t *testing.T, sub, sid string) string {
	t.Helper()
	o, s, c := relayStart(t, h)
	cb, err := h.svc.Callback(context.Background(), identity.CallbackInput{Origin: customOrigin, RawQuery: query(h.code(c, grant(sub, sid)), s),
		Binding: o.Binding})
	if err != nil || cb.Credential == "" {
		t.Fatalf("relay sign-in: %+v %v", cb.Outcome, err)
	}
	return cb.Credential
}
