package workos

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"core/internal/identity"
	"core/internal/identity/identitytest"
)

func newAdapter(t *testing.T) (*Adapter, *identitytest.Fake) {
	t.Helper()
	f := identitytest.New(t, testClientID)
	a, err := New(Config{APIKey: f.APIKey, ClientID: testClientID, Issuer: f.Issuer, BaseURL: f.URL(), Algorithms: []string{"RS256"}})
	if err != nil {
		t.Fatal(err)
	}
	return a, f
}

const verifier = "dGhpcyBpcyBhIHRlc3QgdmVyaWZpZXIgdmFsdWUgMDAwMDAw"

func TestAuthorizationURLSendsExactlyThePlannedParameters(t *testing.T) {
	a, f := newAdapter(t)
	for _, maxAge := range []int{0, 600} {
		u, err := a.AuthorizationURL(context.Background(), identity.AuthRequest{
			State: "st", CodeChallenge: identitytest.Challenge(verifier), RedirectURI: "https://org-a.localhost:543/auth/callback", MaxAge: maxAge})
		if err != nil {
			t.Fatal(err)
		}
		p, err := url.Parse(u)
		if err != nil {
			t.Fatal(err)
		}
		if p.Scheme+"://"+p.Host != f.URL() || p.Path != "/user_management/authorize" {
			t.Fatalf("endpoint %q", u)
		}
		q := p.Query()
		var keys []string
		for k := range q {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		want := []string{"client_id", "code_challenge", "code_challenge_method", "provider", "redirect_uri", "response_type", "state"}
		if maxAge > 0 {
			want = []string{"client_id", "code_challenge", "code_challenge_method", "max_age", "provider", "redirect_uri", "response_type", "state"}
		}
		if !slices.Equal(keys, want) || q.Get("code_challenge_method") != "S256" || q.Get("provider") != "authkit" ||
			q.Get("response_type") != "code" || q.Get("client_id") != testClientID {
			t.Fatalf("parameters %v", q)
		}
	}
}

func TestProviderLogoutURL(t *testing.T) {
	a, f := newAdapter(t)
	u, err := a.ProviderLogoutURL("session_1", "https://org-a.localhost:543/sign-in?logout_ref=r")
	if err != nil {
		t.Fatal(err)
	}
	p, _ := url.Parse(u)
	if p.Scheme+"://"+p.Host != f.URL() || p.Path != "/user_management/sessions/logout" ||
		p.Query().Get("session_id") != "session_1" || p.Query().Get("return_to") != "https://org-a.localhost:543/sign-in?logout_ref=r" {
		t.Fatalf("logout URL %q", u)
	}
}

func TestExchangeValidIdentity(t *testing.T) {
	a, f := newAdapter(t)
	code := f.IssueCode(identitytest.Challenge(verifier), identitytest.Grant{
		UserID: "user_01VALID", SID: "session_01S", FirstName: "  Ada\x00 ", LastName: "<img src=x onerror=alert(1)>", Email: "canary-email@example.test",
		AuthTime: time.Now().Add(-time.Minute), Method: "Password"})
	vi, err := a.Exchange(context.Background(), identity.Exchange{Code: code, CodeVerifier: verifier})
	if err != nil {
		t.Fatal(err)
	}
	if vi.Issuer != f.Issuer || vi.ClientID != testClientID || vi.Subject != "user_01VALID" || vi.ProviderSessionID != "session_01S" ||
		vi.DisplayName != "Ada <img src=x onerror=alert(1)>" || vi.AuthTime.IsZero() {
		t.Fatalf("identity %+v", vi)
	}
	// U-AUTHZ: the validated identity has no email, method, assurance, organization, role or permission field.
	typ := reflect.TypeOf(vi)
	var fields []string
	for i := 0; i < typ.NumField(); i++ {
		fields = append(fields, typ.Field(i).Name)
	}
	if strings.Join(fields, ",") != "Issuer,ClientID,Subject,ProviderSessionID,DisplayName,AuthTime" {
		t.Fatalf("VerifiedIdentity fields = %v", fields)
	}
}

func TestExchangeRejections(t *testing.T) {
	a, f := newAdapter(t)
	grant := identitytest.Grant{UserID: "user_01R", SID: "session_01R"}

	// PKCE: a code redeemed with a different verifier is rejected by the provider (fixture), never mapped.
	code := f.IssueCode(identitytest.Challenge(verifier), grant)
	if _, err := a.Exchange(context.Background(), identity.Exchange{Code: code, CodeVerifier: verifier + "x"}); err != identity.ErrCodeRejected {
		t.Fatalf("wrong verifier: %v", err)
	}
	// One-use: the code is spent even though the first attempt failed.
	if _, err := a.Exchange(context.Background(), identity.Exchange{Code: code, CodeVerifier: verifier}); err != identity.ErrCodeRejected {
		t.Fatalf("reused code: %v", err)
	}
	if _, err := a.Exchange(context.Background(), identity.Exchange{Code: "code_random", CodeVerifier: verifier}); err != identity.ErrCodeRejected {
		t.Fatalf("random code: %v", err)
	}
	if _, err := a.Exchange(context.Background(), identity.Exchange{Code: "", CodeVerifier: verifier}); err != identity.ErrCodeRejected {
		t.Fatalf("empty code: %v", err)
	}

	for name, g := range map[string]identitytest.Grant{
		"impersonator in response":      {UserID: "user_01I", SID: "s1", Impersonator: true},
		"impersonation method":          {UserID: "user_01I", SID: "s1", Method: "Impersonation"},
		"act claim in the access token": {UserID: "user_01I", SID: "s1", Claims: map[string]any{"act": map[string]any{"sub": "admin"}}},
	} {
		c := f.IssueCode(identitytest.Challenge(verifier), g)
		if _, err := a.Exchange(context.Background(), identity.Exchange{Code: c, CodeVerifier: verifier}); err != identity.ErrImpersonationRefused {
			t.Errorf("%s: %v", name, err)
		}
	}

	c := f.IssueCode(identitytest.Challenge(verifier), identitytest.Grant{UserID: "user_01T", SID: "s1", Claims: map[string]any{"iss": "https://evil.example"}})
	if _, err := a.Exchange(context.Background(), identity.Exchange{Code: c, CodeVerifier: verifier}); err != identity.ErrProviderError {
		t.Fatalf("invalid token: %v", err)
	}
}

// TestExchangeAddressProjection (G058 D-15(i), scope amendment 02): the real adapter's address projection of one
// authenticate response decoded by the pinned SDK. email_verified true, false and absent (non-pointer bool: absent
// decodes to false, fail closed) are carried exactly; the address is the response's user email; observedAt is the
// adapter clock of this exchange, never the token's auth_time; the validated identity still has no email. A refused
// exchange yields no projection, and the exported Exchange with a zero minter returns the same identity.
func TestExchangeAddressProjection(t *testing.T) {
	obs := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return obs }
	f := identitytest.New(t, testClientID)
	f.SetNow(clock)
	a, err := New(Config{APIKey: f.APIKey, ClientID: testClientID, Issuer: f.Issuer, BaseURL: f.URL(), Algorithms: []string{"RS256"}, Now: clock})
	if err != nil {
		t.Fatal(err)
	}
	authTime := obs.Add(-4 * time.Minute)
	ctx := context.Background()
	for _, c := range []struct {
		name     string
		grant    identitytest.Grant
		verified bool
	}{
		{"verified", identitytest.Grant{UserID: "user_01PV", SID: "session_01PV", Email: "hana@identity.test", AuthTime: authTime}, true},
		{"false", identitytest.Grant{UserID: "user_01PF", SID: "session_01PF", Email: "hana@identity.test", AuthTime: authTime,
			EmailUnverified: true}, false},
		{"absent", identitytest.Grant{UserID: "user_01PA", SID: "session_01PA", Email: "hana@identity.test", AuthTime: authTime,
			OmitEmailVerified: true}, false},
	} {
		code := f.IssueCode(identitytest.Challenge(verifier), c.grant)
		vi, p, err := a.exchange(ctx, identity.Exchange{Code: code, CodeVerifier: verifier})
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if vi.Subject != c.grant.UserID || vi.Issuer != f.Issuer || !vi.AuthTime.Equal(authTime) {
			t.Errorf("%s identity %+v", c.name, vi)
		}
		if p.address != "hana@identity.test" || p.verified != c.verified {
			t.Errorf("%s projection address %q verified %v, want verified %v", c.name, p.address, p.verified, c.verified)
		}
		if !p.observedAt.Equal(obs) || p.observedAt.Equal(vi.AuthTime) || p.observedAt.Location() != time.UTC {
			t.Errorf("%s observedAt %v, want the adapter clock %v (auth_time %v)", c.name, p.observedAt, obs, vi.AuthTime)
		}
	}

	// A refused exchange (wrong verifier) yields no projection.
	code := f.IssueCode(identitytest.Challenge(verifier), identitytest.Grant{UserID: "user_01PR", SID: "s1", Email: "hana@identity.test"})
	if _, p, err := a.exchange(ctx, identity.Exchange{Code: code, CodeVerifier: verifier + "x"}); err != identity.ErrCodeRejected || p != (projection{}) {
		t.Fatalf("refused exchange: %v %+v", err, p)
	}
	// Impersonation is refused before any projection.
	code = f.IssueCode(identitytest.Challenge(verifier), identitytest.Grant{UserID: "user_01PI", SID: "s1", Email: "hana@identity.test", Impersonator: true})
	if _, p, err := a.exchange(ctx, identity.Exchange{Code: code, CodeVerifier: verifier}); err != identity.ErrImpersonationRefused || p != (projection{}) {
		t.Fatalf("impersonation: %v %+v", err, p)
	}
	// The exported Exchange with a zero minter (every non-enrollment caller) returns the identity and mints nothing.
	code = f.IssueCode(identitytest.Challenge(verifier), identitytest.Grant{UserID: "user_01PZ", SID: "s1", Email: "hana@identity.test"})
	if vi, err := a.Exchange(ctx, identity.Exchange{Code: code, CodeVerifier: verifier}); err != nil || vi.Subject != "user_01PZ" {
		t.Fatalf("zero-minter exchange: %+v %v", vi, err)
	}
}

func TestExchangeProviderFailuresAndBounds(t *testing.T) {
	a, f := newAdapter(t)
	a.exchangeLimit = 300 * time.Millisecond

	f.Delay(time.Second, 0)
	start := time.Now()
	if _, err := a.Exchange(context.Background(), identity.Exchange{Code: "code_x", CodeVerifier: verifier}); err != identity.ErrProviderTimeout {
		t.Fatalf("slow authenticate: %v", err)
	}
	if time.Since(start) > 900*time.Millisecond {
		t.Fatal("authenticate deadline not enforced")
	}
	f.Delay(0, 0)

	f.ServeAuthenticate([]byte(`{"access_token":"` + strings.Repeat("a", authenticateMaxBdy) + `"}`))
	if _, err := a.Exchange(context.Background(), identity.Exchange{Code: "code_x", CodeVerifier: verifier}); err != identity.ErrProviderError {
		t.Fatalf("oversize authenticate body: %v", err)
	}
	f.ServeAuthenticate([]byte(`{"user":{"id":"user_01Q"},"access_token":""}`))
	if _, err := a.Exchange(context.Background(), identity.Exchange{Code: "code_x", CodeVerifier: verifier}); err != identity.ErrProviderError {
		t.Fatalf("missing token: %v", err)
	}
	f.ServeAuthenticate(nil)

	f.FailJWKS(http.StatusServiceUnavailable)
	c := f.IssueCode(identitytest.Challenge(verifier), identitytest.Grant{UserID: "user_01J", SID: "s1"})
	if _, err := a.Exchange(context.Background(), identity.Exchange{Code: c, CodeVerifier: verifier}); err != identity.ErrUnavailable {
		t.Fatalf("JWKS outage: %v", err)
	}
}

func TestRevokeProviderSession(t *testing.T) {
	a, f := newAdapter(t)
	if err := a.RevokeProviderSession(context.Background(), "session_01V"); err != nil {
		t.Fatal(err)
	}
	if got := f.Revoked(); !slices.Equal(got, []string{"session_01V"}) {
		t.Fatalf("revoked %v", got)
	}
	f.FailRevoke(http.StatusInternalServerError)
	if err := a.RevokeProviderSession(context.Background(), "session_01W"); !errors.Is(err, identity.ErrProviderError) {
		t.Fatalf("failing revoke: %v", err)
	}
	if err := a.RevokeProviderSession(context.Background(), ""); err != identity.ErrInvalidRequest {
		t.Fatalf("empty sid: %v", err)
	}
}

func TestDisplayName(t *testing.T) {
	s := func(v string) *string { return &v }
	long := strings.Repeat("é", 250)
	for _, c := range []struct {
		first, last *string
		want        string
	}{
		{nil, nil, ""},
		{s("  "), s("\t"), ""},
		{s("Ada"), nil, "Ada"},
		{nil, s("Lovelace"), "Lovelace"},
		{s("Ada\u0007"), s("Lovelace\n"), "Ada Lovelace"},
		{s(long), nil, strings.Repeat("é", 200)},
	} {
		if got := displayName(c.first, c.last); got != c.want {
			t.Errorf("displayName(%v, %v) = %q, want %q", c.first, c.last, got, c.want)
		}
	}
}
