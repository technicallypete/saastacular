package workos

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"core/internal/identity/identitytest"
)

// U-TOK (Plan r5 §8.2): the access-token contract of §3.3 against a fixture JWKS. Fixture success is never evidence of
// real provider behaviour; CAP-3 and S-TAMP establish that.

const testClientID = "client_01TOKENTEST0000000000000"

type tokenFixture struct {
	fake *identitytest.Fake
	v    *tokenValidator
	now  time.Time
}

func newTokenFixture(t *testing.T) *tokenFixture {
	t.Helper()
	f := identitytest.New(t, testClientID)
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	f.SetNow(func() time.Time { return now })
	clock := func() time.Time { return now }
	keys := newKeySource(f.URL()+"/sso/jwks/"+testClientID, boundedClient(nil, time.Second, jwksMaxBody), clock)
	return &tokenFixture{fake: f, now: now, v: &tokenValidator{
		issuer: f.Issuer, clientID: testClientID, algorithms: []string{"RS256"}, keys: keys, now: clock}}
}

func (tf *tokenFixture) grant() identitytest.Grant {
	return identitytest.Grant{UserID: "user_01ABC", SID: "session_01XYZ", AuthTime: tf.now.Add(-time.Minute)}
}

func (tf *tokenFixture) claims(over map[string]any) map[string]any {
	g := tf.grant()
	g.Claims = over
	return tf.fake.ClaimsFor(g)
}

func TestTokenValid(t *testing.T) {
	tf := newTokenFixture(t)
	c, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC")
	if err != nil {
		t.Fatal(err)
	}
	if c.Issuer != tf.fake.Issuer || c.Subject != "user_01ABC" || c.SID != "session_01XYZ" || c.ClientID != testClientID ||
		!c.AuthTime.Time.Equal(tf.now.Add(-time.Minute).Truncate(time.Second)) {
		t.Fatalf("claims: %+v", c)
	}
}

func TestTokenRejections(t *testing.T) {
	tf := newTokenFixture(t)
	now := tf.now
	attacker := identitytest.NewAttackerKey(t)
	cases := []struct {
		name  string
		token func() string
		user  string
		want  error
	}{
		{"tampered payload", func() string { return identitytest.Tamper(tf.fake.Sign(tf.claims(nil))) }, "user_01ABC", errTokenInvalid},
		{"alg none", func() string { return identitytest.SignWith(attacker, "none", tf.fake.CurrentKID(), tf.claims(nil)) }, "user_01ABC", errTokenInvalid},
		{"HS256 key confusion", func() string { return identitytest.SignWith(attacker, "HS256", tf.fake.CurrentKID(), tf.claims(nil)) }, "user_01ABC", errTokenInvalid},
		{"unlisted algorithm", func() string { return identitytest.SignWith(attacker, "RS512", tf.fake.CurrentKID(), tf.claims(nil)) }, "user_01ABC", errTokenInvalid},
		{"attacker key reusing the trusted kid", func() string { return identitytest.SignWith(attacker, "RS256", tf.fake.CurrentKID(), tf.claims(nil)) }, "user_01ABC", errTokenInvalid},
		{"wrong issuer", func() string { return tf.fake.Sign(tf.claims(map[string]any{"iss": "https://evil.example"})) }, "user_01ABC", errTokenInvalid},
		{"issuer trailing-slash variant", func() string { return tf.fake.Sign(tf.claims(map[string]any{"iss": tf.fake.Issuer + "/"})) }, "user_01ABC", errTokenInvalid},
		{"wrong client_id", func() string { return tf.fake.Sign(tf.claims(map[string]any{"client_id": "client_01OTHER"})) }, "user_01ABC", errTokenInvalid},
		{"missing client_id", func() string { return tf.fake.Sign(tf.claims(map[string]any{"client_id": nil})) }, "user_01ABC", errTokenInvalid},
		{"sub differs from user.id", func() string { return tf.fake.Sign(tf.claims(nil)) }, "user_01OTHER", errTokenInvalid},
		{"sub without user_ prefix", func() string { return tf.fake.Sign(tf.claims(map[string]any{"sub": "01ABC"})) }, "01ABC", errTokenInvalid},
		{"missing sid", func() string { return tf.fake.Sign(tf.claims(map[string]any{"sid": nil})) }, "user_01ABC", errTokenInvalid},
		{"missing auth_time", func() string { return tf.fake.Sign(tf.claims(map[string]any{"auth_time": nil})) }, "user_01ABC", errTokenInvalid},
		{"forged future auth_time", func() string {
			return tf.fake.Sign(tf.claims(map[string]any{"auth_time": now.Add(2 * time.Minute).Unix()}))
		}, "user_01ABC", errTokenInvalid},
		{"zero auth_time", func() string { return tf.fake.Sign(tf.claims(map[string]any{"auth_time": 0})) }, "user_01ABC", errTokenInvalid},
		{"expired beyond leeway", func() string {
			return tf.fake.Sign(tf.claims(map[string]any{"exp": now.Add(-31 * time.Second).Unix()}))
		}, "user_01ABC", errTokenInvalid},
		{"missing exp", func() string { return tf.fake.Sign(tf.claims(map[string]any{"exp": nil})) }, "user_01ABC", errTokenInvalid},
		{"future iat beyond leeway", func() string { return tf.fake.Sign(tf.claims(map[string]any{"iat": now.Add(31 * time.Second).Unix()})) }, "user_01ABC", errTokenInvalid},
		{"missing iat", func() string { return tf.fake.Sign(tf.claims(map[string]any{"iat": nil})) }, "user_01ABC", errTokenInvalid},
		{"future nbf", func() string { return tf.fake.Sign(tf.claims(map[string]any{"nbf": now.Add(time.Minute).Unix()})) }, "user_01ABC", errTokenInvalid},
		{"impersonation act claim", func() string { return tf.fake.Sign(tf.claims(map[string]any{"act": map[string]any{"sub": "admin"}})) }, "user_01ABC", errImpersonation},
		{"refresh token offered as access token", func() string { return "rt_0123456789abcdef0123456789abcdef" }, "user_01ABC", errTokenInvalid},
		{"empty token", func() string { return "" }, "user_01ABC", errTokenInvalid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := tf.v.validate(context.Background(), c.token(), c.user); !errors.Is(err, c.want) {
				t.Fatalf("got %v, want %v", err, c.want)
			}
		})
	}
}

// Leeway boundaries: 30 s either side is accepted, beyond is not (covered above).
func TestTokenLeewayBoundaries(t *testing.T) {
	tf := newTokenFixture(t)
	for name, over := range map[string]map[string]any{
		"exp 29 s ago":         {"exp": tf.now.Add(-29 * time.Second).Unix()},
		"iat 29 s ahead":       {"iat": tf.now.Add(29 * time.Second).Unix()},
		"auth_time 29 s ahead": {"auth_time": tf.now.Add(29 * time.Second).Unix()},
	} {
		if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(over)), "user_01ABC"); err != nil {
			t.Errorf("%s rejected: %v", name, err)
		}
	}
}

// A JWKS outage before any trusted set exists is an availability failure, never a signature failure.
func TestTokenKeyOutageIsUnavailable(t *testing.T) {
	tf := newTokenFixture(t)
	tf.fake.FailJWKS(http.StatusServiceUnavailable)
	if _, err := tf.v.validate(context.Background(), tf.fake.Sign(tf.claims(nil)), "user_01ABC"); !errors.Is(err, errKeysUnavailable) {
		t.Fatalf("got %v, want errKeysUnavailable", err)
	}
}
