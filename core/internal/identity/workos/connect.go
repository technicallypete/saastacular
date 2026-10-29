package workos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"core/internal/identity"
)

// G014: the WorkOS AuthKit Connect access-token validator of the MCP resource server (Plan r8 §3.3; ADR-005). Core is
// the resource server only: it never issues, stores or refreshes MCP bearer values. A Connect access token authenticates
// a subject and nothing else; the caller maps the subject through the existing G004 principal and applies every binding
// and live-authority check itself. Provider JWT libraries stay in this package (B-IDENTITY-GUARDS): callers receive a
// provider-free ConnectClaims or a closed identity error.

// connectMaxToken bounds the accepted token size (Plan r1 §3.6: 16 KiB token and internal bodies).
const connectMaxToken = 16 << 10

// clientIDMax bounds the client identifier (a client ID metadata document's ID is an https URL).
const clientIDMax = 2048

// ConnectConfig configures the validator from the validated core MCP configuration.
type ConnectConfig struct {
	Issuer     string   // exact MCP_CONNECT_ISSUER; the key set is <issuer>/oauth2/jwks
	Algorithms []string // allow-listed asymmetric JWS algorithms (default RS256)
	// JWKS is the test trust anchor (MCP_CONNECT_JWKS_FILE contents; C10). The configuration loader accepts it only with
	// APP_ENV=test; when set, no network key set is used.
	JWKS      []byte
	Now       func() time.Time
	Transport http.RoundTripper // optional; tests only
}

// ConnectClaims is the provider-free result of a successful validation.
type ConnectClaims struct {
	Subject  string
	ClientID string
	Audience string
	Expiry   time.Time
}

// connectKeys is a key source: the bounded, refreshing network key set (jwks.go) or the static test trust anchor.
type connectKeys interface {
	keyfunc(ctx context.Context) jwt.Keyfunc
}

// ConnectValidator validates Connect access tokens.
type ConnectValidator struct {
	issuer     string
	algorithms []string
	keys       connectKeys
	now        func() time.Time
}

// connectClaims are the only claims read. sid, org_id, auth_time and scope are never decoded into authority.
type connectClaims struct {
	jwt.RegisteredClaims
	ClientID string          `json:"client_id"`
	Act      json.RawMessage `json:"act"`
}

// NewConnectValidator returns a validator, or an error for an incomplete configuration or an unusable test key set.
func NewConnectValidator(cfg ConnectConfig) (*ConnectValidator, error) {
	if cfg.Issuer == "" || len(cfg.Algorithms) == 0 {
		return nil, errors.New("workos: incomplete connect configuration")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	v := &ConnectValidator{issuer: cfg.Issuer, algorithms: cfg.Algorithms, now: cfg.Now}
	if cfg.JWKS != nil {
		keys, err := newStaticKeys(cfg.JWKS)
		if err != nil {
			return nil, err
		}
		v.keys = keys
		return v, nil
	}
	u, err := url.Parse(strings.TrimSuffix(cfg.Issuer, "/") + "/oauth2/jwks")
	if err != nil || u.Host == "" {
		return nil, errors.New("workos: invalid connect issuer")
	}
	v.keys = newKeySource(u.String(), boundedClient(cfg.Transport, jwksFetchDeadline, jwksMaxBody), cfg.Now)
	return v, nil
}

// staticKeys is the test trust anchor: a fixed key set, never refreshed; an unknown kid is an unknown key (401), never
// an outage.
type staticKeys struct{ kf keyfunc.Keyfunc }

func newStaticKeys(body []byte) (*staticKeys, error) {
	var set jwkset.JWKSMarshal
	if err := json.Unmarshal(body, &set); err != nil || len(set.Keys) == 0 {
		return nil, errors.New("workos: invalid test key set")
	}
	store, err := set.ToStorage()
	if err != nil {
		return nil, errors.New("workos: invalid test key set")
	}
	kf, err := keyfunc.New(keyfunc.Options{Ctx: context.Background(), Storage: store})
	if err != nil {
		return nil, errors.New("workos: invalid test key set")
	}
	return &staticKeys{kf: kf}, nil
}

func (s *staticKeys) keyfunc(ctx context.Context) jwt.Keyfunc {
	return func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok || kid == "" || len(kid) > 256 {
			return nil, errUnknownKey
		}
		key, err := s.kf.KeyfuncCtx(ctx)(t)
		if err != nil {
			return nil, errUnknownKey
		}
		return key, nil
	}
}

// Validate verifies raw against the Connect contract (Plan r8 §3.3): an allow-listed asymmetric signature from the key
// set, the exact issuer, exactly one audience equal string-for-string to resource, exp in the future and iat/nbf not
// ahead (30 s leeway), a WorkOS user subject and a client_id. An act (impersonation) claim is refused. ID tokens (client
// audience), AuthKit session tokens (G004 issuer) and every other value fail the issuer or audience checks.
//
// It returns identity.ErrUnavailable when no trusted key set can be obtained (503, never anonymous),
// identity.ErrImpersonationRefused for an act claim, and identity.ErrUnauthenticated for every other failure.
func (v *ConnectValidator) Validate(ctx context.Context, raw, resource string) (ConnectClaims, error) {
	if raw == "" || len(raw) > connectMaxToken || resource == "" {
		return ConnectClaims{}, identity.ErrUnauthenticated
	}
	var c connectClaims
	parser := jwt.NewParser(
		jwt.WithValidMethods(v.algorithms),
		jwt.WithIssuer(v.issuer),
		jwt.WithLeeway(leeway),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(v.now),
		jwt.WithStrictDecoding(),
	)
	if _, err := parser.ParseWithClaims(raw, &c, v.keys.keyfunc(ctx)); err != nil {
		if errors.Is(err, errKeysUnavailable) {
			return ConnectClaims{}, identity.ErrUnavailable
		}
		return ConnectClaims{}, identity.ErrUnauthenticated
	}
	switch {
	case c.Issuer != v.issuer:
		return ConnectClaims{}, identity.ErrUnauthenticated
	case len(c.Audience) != 1 || c.Audience[0] != resource:
		return ConnectClaims{}, identity.ErrUnauthenticated
	case !subjectRe.MatchString(c.Subject):
		return ConnectClaims{}, identity.ErrUnauthenticated
	case c.ClientID == "" || len(c.ClientID) > clientIDMax || strings.ContainsAny(c.ClientID, " \t\r\n"):
		return ConnectClaims{}, identity.ErrUnauthenticated
	case c.IssuedAt == nil || c.ExpiresAt == nil:
		return ConnectClaims{}, identity.ErrUnauthenticated
	}
	if len(c.Act) > 0 && string(c.Act) != "null" {
		return ConnectClaims{}, identity.ErrImpersonationRefused
	}
	return ConnectClaims{Subject: c.Subject, ClientID: c.ClientID, Audience: c.Audience[0], Expiry: c.ExpiresAt.Time.UTC()}, nil
}
