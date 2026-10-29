package workos

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"core/internal/identity"
	"core/internal/identity/identitytest"
)

// U-MCP-CONNECT-TOKEN (G014 Plan r8 §10.1): the Connect access-token contract of §3.3 against the fixture key set, over
// the network key set and the test trust anchor (C10). Fixture success is never evidence of real provider behaviour; the
// accepted QP and QP-B records hold the provider facts.

const (
	connectClient   = "client_01CONNECTTEST000000000"
	connectSubject  = "user_01CONNECTSUBJECT"
	connectResource = "https://org-a.localhost:543/mcp/c/7f000000-0000-4000-8000-000000000001"
)

type connectFixture struct {
	fake *identitytest.Fake
	now  time.Time
	v    *ConnectValidator
}

func newConnectFixture(t *testing.T, static bool) *connectFixture {
	t.Helper()
	f := identitytest.New(t, connectClient)
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	f.SetNow(func() time.Time { return now })
	cfg := ConnectConfig{Issuer: f.ConnectIssuer(), Algorithms: []string{"RS256"}, Now: func() time.Time { return now }}
	if static {
		cfg.JWKS = f.ConnectJWKS()
	}
	v, err := NewConnectValidator(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return &connectFixture{fake: f, now: now, v: v}
}

func (cf *connectFixture) token(over map[string]any) string {
	return cf.fake.SignConnect(connectSubject, connectClient, connectResource, over)
}

func TestConnectValid(t *testing.T) {
	for _, static := range []bool{false, true} {
		cf := newConnectFixture(t, static)
		c, err := cf.v.Validate(context.Background(), cf.token(nil), connectResource)
		if err != nil {
			t.Fatalf("static=%v: %v", static, err)
		}
		if c.Subject != connectSubject || c.ClientID != connectClient || c.Audience != connectResource ||
			!c.Expiry.Equal(cf.now.Add(time.Hour)) {
			t.Fatalf("claims: %+v", c)
		}
		// A one-element audience array is the same single audience.
		if _, err := cf.v.Validate(context.Background(), cf.token(map[string]any{"aud": []string{connectResource}}), connectResource); err != nil {
			t.Fatalf("static=%v: one-element audience array: %v", static, err)
		}
	}
}

func TestConnectRejections(t *testing.T) {
	cf := newConnectFixture(t, false)
	now := cf.now
	attacker := identitytest.NewAttackerKey(t)
	kid := cf.fake.CurrentKID()
	for name, c := range map[string]struct {
		token    string
		resource string
		want     error
	}{
		"alg none":                 {identitytest.SignWith(attacker, "none", kid, cf.fake.ConnectClaims(connectSubject, connectClient, connectResource)), connectResource, identity.ErrUnauthenticated},
		"HS256 key confusion":      {identitytest.SignWith(attacker, "HS256", kid, cf.fake.ConnectClaims(connectSubject, connectClient, connectResource)), connectResource, identity.ErrUnauthenticated},
		"attacker key, reused kid": {identitytest.SignWith(attacker, "RS256", kid, cf.fake.ConnectClaims(connectSubject, connectClient, connectResource)), connectResource, identity.ErrUnauthenticated},
		"not allow-listed alg":     {identitytest.SignWith(attacker, "RS384", kid, cf.fake.ConnectClaims(connectSubject, connectClient, connectResource)), connectResource, identity.ErrUnauthenticated},
		"tampered payload":         {identitytest.Tamper(cf.token(nil)), connectResource, identity.ErrUnauthenticated},
		"other issuer":             {cf.token(map[string]any{"iss": "https://other.authkit.example"}), connectResource, identity.ErrUnauthenticated},
		"issuer with slash":        {cf.token(map[string]any{"iss": cf.fake.ConnectIssuer() + "/"}), connectResource, identity.ErrUnauthenticated},
		"other audience":           {cf.token(nil), connectResource + "x", identity.ErrUnauthenticated},
		"audience of another host": {cf.token(map[string]any{"aud": "https://org-b.localhost:543/mcp/c/7f000000-0000-4000-8000-000000000001"}), connectResource, identity.ErrUnauthenticated},
		"two audiences":            {cf.token(map[string]any{"aud": []string{connectResource, "https://org-b.localhost:543/mcp/c/x"}}), connectResource, identity.ErrUnauthenticated},
		"no audience (session)":    {cf.token(map[string]any{"aud": nil}), connectResource, identity.ErrUnauthenticated},
		"ID token shape":           {cf.token(map[string]any{"aud": connectClient}), connectResource, identity.ErrUnauthenticated},
		"missing sub":              {cf.token(map[string]any{"sub": nil}), connectResource, identity.ErrUnauthenticated},
		"non-user sub":             {cf.token(map[string]any{"sub": "client_01X"}), connectResource, identity.ErrUnauthenticated},
		"missing client_id":        {cf.token(map[string]any{"client_id": nil}), connectResource, identity.ErrUnauthenticated},
		"client_id with space":     {cf.token(map[string]any{"client_id": "client 01"}), connectResource, identity.ErrUnauthenticated},
		"expired":                  {cf.token(map[string]any{"exp": now.Add(-31 * time.Second).Unix()}), connectResource, identity.ErrUnauthenticated},
		"missing exp":              {cf.token(map[string]any{"exp": nil}), connectResource, identity.ErrUnauthenticated},
		"missing iat":              {cf.token(map[string]any{"iat": nil}), connectResource, identity.ErrUnauthenticated},
		"iat ahead":                {cf.token(map[string]any{"iat": now.Add(31 * time.Second).Unix()}), connectResource, identity.ErrUnauthenticated},
		"nbf ahead":                {cf.token(map[string]any{"nbf": now.Add(31 * time.Second).Unix()}), connectResource, identity.ErrUnauthenticated},
		"impersonation (act)":      {cf.token(map[string]any{"act": map[string]any{"sub": "user_01ADMIN"}}), connectResource, identity.ErrImpersonationRefused},
		"oversized":                {strings.Repeat("a", connectMaxToken+1), connectResource, identity.ErrUnauthenticated},
		"empty resource":           {cf.token(nil), "", identity.ErrUnauthenticated},
	} {
		if _, err := cf.v.Validate(context.Background(), c.token, c.resource); !errors.Is(err, c.want) {
			t.Errorf("%s: %v, want %v", name, err, c.want)
		}
	}
	// Within leeway is accepted.
	if _, err := cf.v.Validate(context.Background(), cf.token(map[string]any{"exp": now.Add(-29 * time.Second).Unix(),
		"iat": now.Add(29 * time.Second).Unix(), "nbf": now.Add(29 * time.Second).Unix()}), connectResource); err != nil {
		t.Errorf("within leeway: %v", err)
	}
	// A G004 AuthKit access token (no resource audience) is refused.
	g4 := cf.fake.Sign(cf.fake.ClaimsFor(identitytest.Grant{UserID: connectSubject, SID: "session_01X"}))
	if _, err := cf.v.Validate(context.Background(), g4, connectResource); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Errorf("G004 access token: %v", err)
	}
	// An act claim of JSON null is not impersonation.
	if _, err := cf.v.Validate(context.Background(), cf.token(map[string]any{"act": nil}), connectResource); err != nil {
		t.Errorf("act absent: %v", err)
	}
}

func TestConnectKeyRotationAndOutage(t *testing.T) {
	cf := newConnectFixture(t, false)
	ctx := context.Background()
	if _, err := cf.v.Validate(ctx, cf.token(nil), connectResource); err != nil {
		t.Fatal(err)
	}
	// A rotated key the key set publishes is fetched once on the unknown kid.
	cf.fake.Rotate(t)
	if _, err := cf.v.Validate(ctx, cf.token(nil), connectResource); err != nil {
		t.Fatalf("rotated, published key: %v", err)
	}
	// A reachable key set without the key: unauthenticated (the unknown-kid refresh is rate limited).
	cf.fake.RotateUnpublished(t)
	if _, err := cf.v.Validate(ctx, cf.token(nil), connectResource); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("unpublished key: %v", err)
	}

	// An unreachable key set with no cached key: unavailable (503), never unauthenticated.
	down := newConnectFixture(t, false)
	down.fake.FailJWKS(http.StatusServiceUnavailable)
	if _, err := down.v.Validate(ctx, down.token(nil), connectResource); !errors.Is(err, identity.ErrUnavailable) {
		t.Fatalf("key set down: %v", err)
	}
}

func TestConnectStaticTrustAnchor(t *testing.T) {
	cf := newConnectFixture(t, true)
	ctx := context.Background()
	// The static key set never refreshes: an unknown kid is unauthenticated, never an outage, and no fetch happens.
	hits := cf.fake.JWKSHits()
	cf.fake.Rotate(t)
	if _, err := cf.v.Validate(ctx, cf.token(nil), connectResource); !errors.Is(err, identity.ErrUnauthenticated) {
		t.Fatalf("unknown kid under the static anchor: %v", err)
	}
	if cf.fake.JWKSHits() != hits {
		t.Fatal("the static trust anchor fetched a key set")
	}
	for name, body := range map[string][]byte{"empty": []byte(`{"keys":[]}`), "not json": []byte("x"), "no keys": []byte(`{}`)} {
		if _, err := NewConnectValidator(ConnectConfig{Issuer: "http://issuer.test", Algorithms: []string{"RS256"}, JWKS: body}); err == nil {
			t.Errorf("%s test key set accepted", name)
		}
	}
	if _, err := NewConnectValidator(ConnectConfig{Algorithms: []string{"RS256"}}); err == nil {
		t.Error("missing issuer accepted")
	}
	if _, err := NewConnectValidator(ConnectConfig{Issuer: "https://x.authkit.app"}); err == nil {
		t.Error("missing algorithms accepted")
	}
}
