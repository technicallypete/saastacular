// Package workos is the WorkOS AuthKit adapter behind identity.Provider (ADR-003; Plan r5 §3). It is the only
// package that imports workos-go, golang-jwt and keyfunc/jwkset. Provider tokens, claims and SDK types never leave
// it: callers receive an identity.VerifiedIdentity or a closed identity.Error.
package workos

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	sdk "github.com/workos/workos-go/v10"

	"core/internal/identity"
)

// Call deadlines (Plan r5 §3.6): authenticate 8 s; JWKS, revoke and the rest 5 s.
const (
	authenticateTimeout = 8 * time.Second
	defaultTimeout      = 5 * time.Second
	maxDisplayName      = 200
)

// Config is the adapter configuration, taken from the validated core identity configuration.
type Config struct {
	APIKey     string
	ClientID   string
	Issuer     string
	BaseURL    string // https://api.workos.com, or the emulator with APP_ENV=test
	Algorithms []string
	Now        func() time.Time
	Transport  http.RoundTripper // optional; tests only
}

// Adapter implements identity.Provider for WorkOS AuthKit.
type Adapter struct {
	cfg           Config
	client        *sdk.Client
	tokens        *tokenValidator
	exchangeLimit time.Duration
	callLimit     time.Duration
}

// New returns an adapter. Every endpoint is built from the configured base URL and client ID only; no token,
// response, Host or forwarding header can choose an endpoint.
func New(cfg Config) (*Adapter, error) {
	if cfg.APIKey == "" || cfg.ClientID == "" || cfg.Issuer == "" || cfg.BaseURL == "" || len(cfg.Algorithms) == 0 {
		return nil, errors.New("workos: incomplete configuration")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
	api := boundedClient(cfg.Transport, authenticateTimeout, authenticateMaxBdy)
	client := sdk.NewClient(cfg.APIKey,
		sdk.WithBaseURL(cfg.BaseURL), sdk.WithClientID(cfg.ClientID), sdk.WithHTTPClient(api), sdk.WithMaxRetries(0))
	jwksURL := cfg.BaseURL + "/sso/jwks/" + url.PathEscape(cfg.ClientID)
	keys := newKeySource(jwksURL, boundedClient(cfg.Transport, jwksFetchDeadline, jwksMaxBody), cfg.Now)
	return &Adapter{cfg: cfg, client: client, exchangeLimit: authenticateTimeout, callLimit: defaultTimeout,
		tokens: &tokenValidator{issuer: cfg.Issuer, clientID: cfg.ClientID, algorithms: cfg.Algorithms, keys: keys, now: cfg.Now}}, nil
}

// AuthorizationURL builds the hosted AuthKit authorization URL. It sends exactly client_id, redirect_uri,
// response_type=code, provider=authkit, state, an S256 code_challenge and, only when freshness is configured,
// max_age; never organization_id, connection_id, screen_hint, login_hint or prompt.
func (a *Adapter) AuthorizationURL(_ context.Context, r identity.AuthRequest) (string, error) {
	if r.State == "" || r.CodeChallenge == "" || r.RedirectURI == "" {
		return "", identity.ErrInvalidRequest
	}
	q := url.Values{}
	q.Set("client_id", a.cfg.ClientID)
	q.Set("redirect_uri", r.RedirectURI)
	q.Set("response_type", "code")
	q.Set("provider", "authkit")
	q.Set("state", r.State)
	q.Set("code_challenge", r.CodeChallenge)
	q.Set("code_challenge_method", "S256")
	if r.MaxAge > 0 {
		q.Set("max_age", strconv.Itoa(r.MaxAge))
	}
	return a.cfg.BaseURL + "/user_management/authorize?" + q.Encode(), nil
}

// Exchange redeems the code with the transaction's PKCE verifier (and the API key as client secret) and validates the
// returned access token (Plan r5 §3.3). Refresh and OAuth tokens are discarded unread; authentication method (except to
// refuse impersonation) and vendor organization data are dropped here. G058 (D-15(i); readiness correction 01 §2): after
// the identity is validated, the address projection of this same response is handed to the caller's sealed minter
// (in.Minter; a zero minter records nothing) and never returned in VerifiedIdentity.
func (a *Adapter) Exchange(ctx context.Context, in identity.Exchange) (identity.VerifiedIdentity, error) {
	vi, ev, err := a.exchange(ctx, in)
	if err != nil {
		return identity.VerifiedIdentity{}, err
	}
	in.Minter.Mint(vi, ev.address, ev.verified, ev.observedAt)
	return vi, nil
}

// projection is the adapter's address observation of one authenticate response: resp.User.Email, the pinned SDK's
// non-pointer resp.User.EmailVerified (models.go:7099/7101, v10.5.0: an absent email_verified decodes to false, so
// absence fails closed) and observedAt, the adapter clock read after the response's access token and impersonation
// checks passed. It is the observation of this exchange: never claims.AuthTime (the provider's authentication instant)
// and never a provider verification timestamp (the pinned SDK has none).
type projection struct {
	address    string
	verified   bool
	observedAt time.Time
}

// exchange is Exchange without the minting step: the validated identity and the projection of the same response.
func (a *Adapter) exchange(ctx context.Context, in identity.Exchange) (identity.VerifiedIdentity, projection, error) {
	if in.Code == "" || in.CodeVerifier == "" {
		return identity.VerifiedIdentity{}, projection{}, identity.ErrCodeRejected
	}
	ctx, cancel := context.WithTimeout(ctx, a.exchangeLimit)
	defer cancel()
	verifier := in.CodeVerifier
	resp, err := a.client.UserManagement().AuthenticateWithCode(ctx, &sdk.UserManagementAuthenticateWithCodeParams{
		Code: in.Code, CodeVerifier: &verifier,
	})
	if err != nil {
		return identity.VerifiedIdentity{}, projection{}, mapExchangeError(ctx, err)
	}
	vi, err := a.verify(ctx, resp)
	if err != nil {
		return identity.VerifiedIdentity{}, projection{}, err
	}
	// verify refused a nil resp.User; the observation clock is read only now, after every check of this exchange passed.
	return vi, projection{address: resp.User.Email, verified: resp.User.EmailVerified, observedAt: a.cfg.Now().UTC()}, nil
}

// verify applies the §3.3 contract to an authenticate response obtained by core over the configured TLS endpoint:
// the access-token checks, then the impersonation refusal. It returns only the validated identity.
func (a *Adapter) verify(ctx context.Context, resp *sdk.AuthenticateResponse) (identity.VerifiedIdentity, error) {
	if resp == nil || resp.User == nil || resp.User.ID == "" || resp.AccessToken == "" {
		return identity.VerifiedIdentity{}, identity.ErrProviderError
	}
	claims, err := a.tokens.validate(ctx, resp.AccessToken, resp.User.ID)
	switch {
	case errors.Is(err, errKeysUnavailable):
		return identity.VerifiedIdentity{}, identity.ErrUnavailable
	case errors.Is(err, errImpersonation):
		return identity.VerifiedIdentity{}, identity.ErrImpersonationRefused
	case err != nil:
		return identity.VerifiedIdentity{}, identity.ErrProviderError
	}
	if resp.Impersonator != nil ||
		(resp.AuthenticationMethod != nil && string(*resp.AuthenticationMethod) == "Impersonation") {
		return identity.VerifiedIdentity{}, identity.ErrImpersonationRefused
	}
	return identity.VerifiedIdentity{
		Issuer:            claims.Issuer,
		ClientID:          claims.ClientID,
		Subject:           claims.Subject,
		ProviderSessionID: claims.SID,
		DisplayName:       displayName(resp.User.FirstName, resp.User.LastName),
		AuthTime:          claims.AuthTime.Time.UTC(),
	}, nil
}

// RevokeProviderSession revokes exactly one WorkOS session. Success proves only that this session was revoked.
func (a *Adapter) RevokeProviderSession(ctx context.Context, sid string) error {
	if sid == "" {
		return identity.ErrInvalidRequest
	}
	ctx, cancel := context.WithTimeout(ctx, a.callLimit)
	defer cancel()
	if err := a.client.UserManagement().RevokeSession(ctx, &sdk.UserManagementRevokeSessionParams{SessionID: sid}); err != nil {
		if timedOut(ctx, err) {
			return identity.ErrProviderTimeout
		}
		return identity.ErrProviderError
	}
	return nil
}

// ProviderLogoutURL returns the hosted logout URL for one session with an allow-listed return destination.
func (a *Adapter) ProviderLogoutURL(sid, returnTo string) (string, error) {
	if sid == "" || returnTo == "" {
		return "", identity.ErrInvalidRequest
	}
	q := url.Values{}
	q.Set("session_id", sid)
	q.Set("return_to", returnTo)
	return a.cfg.BaseURL + "/user_management/sessions/logout?" + q.Encode(), nil
}

// mapExchangeError maps an authenticate failure onto the closed vocabulary without reading any provider body:
// a provider rejection of the code/verifier is code_rejected, a timeout provider_timeout, anything else provider_error.
func mapExchangeError(ctx context.Context, err error) error {
	if timedOut(ctx, err) {
		return identity.ErrProviderTimeout
	}
	// Every typed SDK API error (authentication, not found, unprocessable, rate limit, server) unwraps to *APIError.
	var apiErr *sdk.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 &&
		apiErr.StatusCode != http.StatusTooManyRequests {
		return identity.ErrCodeRejected
	}
	return identity.ErrProviderError
}

func timedOut(ctx context.Context, err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
		return true
	}
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// displayName is the trimmed "first last" name, without control characters, truncated to 200 characters; empty when
// both are absent. It is untrusted text, never a key.
func displayName(first, last *string) string {
	var parts []string
	for _, p := range []*string{first, last} {
		if p == nil {
			continue
		}
		if s := strings.TrimSpace(strings.Map(func(r rune) rune {
			if unicode.IsControl(r) || r == utf8.RuneError {
				return -1
			}
			return r
		}, *p)); s != "" {
			parts = append(parts, s)
		}
	}
	name := strings.Join(parts, " ")
	if utf8.RuneCountInString(name) > maxDisplayName {
		name = strings.TrimSpace(string([]rune(name)[:maxDisplayName]))
	}
	return name
}
