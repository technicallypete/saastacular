// Package identitytest is a test-only fake of the WorkOS endpoints core uses (authenticate, client JWKS, session
// revoke, and G014's Connect key set at /oauth2/jwks) plus a token signer (G004 access tokens and G014 Connect access
// tokens). It is imported only from _test.go files; B-IDENTITY-GUARDS asserts that no production
// package or binary contains it. Fixture success is never evidence of real provider behaviour (Plan r5 §8.1).
package identitytest

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Grant describes the identity a code redeems to.
type Grant struct {
	UserID       string
	FirstName    string
	LastName     string
	Email        string
	SID          string
	AuthTime     time.Time
	Claims       map[string]any // extra or overriding access-token claims (nil value deletes a claim)
	Impersonator bool
	Method       string // authentication_method in the response
	// G058: the user's email_verified in the authenticate response is true unless EmailUnverified (false) or
	// OmitEmailVerified (the field is absent). Fixture behaviour, never provider proof.
	EmailUnverified   bool
	OmitEmailVerified bool
}

type code struct {
	challenge string
	grant     Grant
	used      bool
}

// Fake is an httptest server implementing the fake WorkOS API.
type Fake struct {
	Server   *httptest.Server
	ClientID string
	APIKey   string
	Issuer   string

	mu            sync.Mutex
	keys          []*signingKey // index 0 signs; all are published
	codes         map[string]*code
	revoked       []string
	jwksHits      int
	authHits      int
	revokeStatus  int
	jwksStatus    int
	jwksBody      []byte        // non-nil: served instead of the real set
	authBody      []byte        // non-nil: served instead of the authenticate response
	authDelay     time.Duration // added before answering authenticate
	jwksDelay     time.Duration
	jwksRedirect  string
	now           func() time.Time
	unpublishNext bool
}

type signingKey struct {
	kid string
	key *rsa.PrivateKey
}

// New starts a fake for clientID. The issuer is the fake's own base URL unless overridden by SetIssuer.
func New(t testing.TB, clientID string) *Fake {
	t.Helper()
	f := &Fake{ClientID: clientID, APIKey: "sk_test_fake_" + randomHex(8), codes: map[string]*code{}, now: time.Now,
		revokeStatus: http.StatusOK, jwksStatus: http.StatusOK}
	f.keys = []*signingKey{newKey(t)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /sso/jwks/{client}", f.handleJWKS)
	mux.HandleFunc("POST /user_management/authenticate", f.handleAuthenticate)
	mux.HandleFunc("POST /user_management/sessions/revoke", f.handleRevoke)
	// G014: the Connect issuer mode publishes the same signing keys at the AuthKit key-set path.
	mux.HandleFunc("GET /oauth2/jwks", f.handleConnectJWKS)
	f.Server = httptest.NewServer(mux)
	f.Issuer = f.Server.URL
	t.Cleanup(f.Server.Close)
	return f
}

func newKey(t testing.TB) *signingKey {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return &signingKey{kid: "kid_" + randomHex(6), key: k}
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// URL is the fake's base URL (the adapter's IDENTITY_WORKOS_API_BASE_URL in tests).
func (f *Fake) URL() string { return f.Server.URL }

// SetNow sets the fake's clock used for iat/exp of minted tokens.
func (f *Fake) SetNow(now func() time.Time) { f.mu.Lock(); f.now = now; f.mu.Unlock() }

// Rotate adds a new signing key and makes it current; the previous key stays published.
func (f *Fake) Rotate(t testing.TB) string {
	t.Helper()
	k := newKey(t)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.keys = append([]*signingKey{k}, f.keys...)
	return k.kid
}

// RotateUnpublished adds a new current signing key that the JWKS does not publish until Publish is called.
func (f *Fake) RotateUnpublished(t testing.TB) string {
	kid := f.Rotate(t)
	f.mu.Lock()
	f.unpublishNext = true
	f.mu.Unlock()
	return kid
}

// Publish makes every key visible in the JWKS.
func (f *Fake) Publish() { f.mu.Lock(); f.unpublishNext = false; f.mu.Unlock() }

// IssueCode registers a one-use code redeemable with the verifier whose S256 challenge is challenge.
func (f *Fake) IssueCode(challenge string, g Grant) string {
	c := "code_" + randomHex(16)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.codes[c] = &code{challenge: challenge, grant: g}
	return c
}

// Revoked returns the session IDs revoked so far.
func (f *Fake) Revoked() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.revoked...)
}

// JWKSHits and AuthHits count requests.
func (f *Fake) JWKSHits() int { f.mu.Lock(); defer f.mu.Unlock(); return f.jwksHits }
func (f *Fake) AuthHits() int { f.mu.Lock(); defer f.mu.Unlock(); return f.authHits }

// FailRevoke makes revoke answer status (0 restores 200).
func (f *Fake) FailRevoke(status int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if status == 0 {
		status = http.StatusOK
	}
	f.revokeStatus = status
}

// FailJWKS makes the JWKS endpoint answer status (0 restores 200).
func (f *Fake) FailJWKS(status int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if status == 0 {
		status = http.StatusOK
	}
	f.jwksStatus = status
}

// ServeJWKS replaces the JWKS body (nil restores the real set).
func (f *Fake) ServeJWKS(body []byte) { f.mu.Lock(); f.jwksBody = body; f.mu.Unlock() }

// RedirectJWKS makes the JWKS endpoint redirect to target ("" restores).
func (f *Fake) RedirectJWKS(target string) { f.mu.Lock(); f.jwksRedirect = target; f.mu.Unlock() }

// ServeAuthenticate replaces the authenticate body (nil restores).
func (f *Fake) ServeAuthenticate(body []byte) { f.mu.Lock(); f.authBody = body; f.mu.Unlock() }

// Delay delays authenticate and JWKS answers.
func (f *Fake) Delay(auth, jwks time.Duration) {
	f.mu.Lock()
	f.authDelay, f.jwksDelay = auth, jwks
	f.mu.Unlock()
}

// Challenge is the PKCE S256 challenge of verifier.
func Challenge(verifier string) string {
	s := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(s[:])
}

func (f *Fake) handleJWKS(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.jwksHits++
	status, body, delay, redirect := f.jwksStatus, f.jwksBody, f.jwksDelay, f.jwksRedirect
	keys := f.keys
	if f.unpublishNext && len(keys) > 1 {
		keys = keys[1:]
	}
	f.mu.Unlock()
	if delay > 0 {
		time.Sleep(delay)
	}
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}
	if r.PathValue("client") != f.ClientID {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	if body == nil {
		body = jwks(keys)
	}
	_, _ = w.Write(body)
}

// JWKSOf renders a key set for arbitrary keys (used to serve attacker keys under a reused kid).
func JWKSOf(kid string, pub *rsa.PublicKey) []byte {
	return jwks([]*signingKey{{kid: kid, key: &rsa.PrivateKey{PublicKey: *pub}}})
}

func jwks(keys []*signingKey) []byte {
	type jwk struct {
		Kty string `json:"kty"`
		Use string `json:"use"`
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	}
	set := struct {
		Keys []jwk `json:"keys"`
	}{}
	for _, k := range keys {
		set.Keys = append(set.Keys, jwk{Kty: "RSA", Use: "sig", Alg: "RS256", Kid: k.kid,
			N: base64.RawURLEncoding.EncodeToString(k.key.N.Bytes()),
			E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(k.key.E)).Bytes())})
	}
	b, _ := json.Marshal(set)
	return b
}

func (f *Fake) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		CodeVerifier string `json:"code_verifier"`
	}
	f.mu.Lock()
	f.authHits++
	delay, override := f.authDelay, f.authBody
	f.mu.Unlock()
	if delay > 0 {
		time.Sleep(delay)
	}
	if override != nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(override)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.GrantType != "authorization_code" ||
		in.ClientID != f.ClientID || in.ClientSecret != f.APIKey || r.Header.Get("Authorization") != "Bearer "+f.APIKey {
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client")
		return
	}
	f.mu.Lock()
	c, ok := f.codes[in.Code]
	if ok && !c.used {
		c.used = true
	} else {
		ok = false
	}
	f.mu.Unlock()
	if !ok || Challenge(in.CodeVerifier) != c.challenge {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant")
		return
	}
	g := c.grant
	token := f.Sign(f.claimsFor(g))
	user := map[string]any{"object": "user", "id": g.UserID, "first_name": nullIfEmpty(g.FirstName),
		"last_name": nullIfEmpty(g.LastName), "email": g.Email, "email_verified": !g.EmailUnverified,
		"created_at": "2026-10-07T00:00:00.000Z", "updated_at": "2026-10-07T00:00:00.000Z"}
	if g.OmitEmailVerified {
		delete(user, "email_verified")
	}
	resp := map[string]any{
		"user":            user,
		"access_token":    token,
		"refresh_token":   "rt_" + randomHex(16),
		"organization_id": "org_" + randomHex(6),
	}
	if g.Method != "" {
		resp["authentication_method"] = g.Method
	}
	if g.Impersonator {
		resp["impersonator"] = map[string]any{"email": "admin@example.test", "reason": "support"}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func writeOAuthError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "error_description": "fixture rejection"})
}

func (f *Fake) handleRevoke(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SessionID string `json:"session_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	f.mu.Lock()
	status := f.revokeStatus
	if status == http.StatusOK && r.Header.Get("Authorization") == "Bearer "+f.APIKey {
		f.revoked = append(f.revoked, in.SessionID)
	}
	f.mu.Unlock()
	if r.Header.Get("Authorization") != "Bearer "+f.APIKey {
		status = http.StatusUnauthorized
	}
	w.WriteHeader(status)
}

// ClaimsFor returns the default access-token claims for a grant (exported for token tests).
func (f *Fake) ClaimsFor(g Grant) map[string]any { return f.claimsFor(g) }

func (f *Fake) claimsFor(g Grant) map[string]any {
	f.mu.Lock()
	now := f.now()
	f.mu.Unlock()
	authTime := g.AuthTime
	if authTime.IsZero() {
		authTime = now
	}
	c := map[string]any{
		"iss": f.Issuer, "sub": g.UserID, "client_id": f.ClientID, "sid": g.SID,
		"iat": now.Unix(), "exp": now.Add(5 * time.Minute).Unix(), "auth_time": authTime.Unix(),
		"org_id": "org_vendor", "role": "admin", "roles": []string{"admin"}, "permissions": []string{"all"},
		"entitlements": []string{"x"}, "feature_flags": []string{"y"}, "jti": "jti_" + randomHex(6),
	}
	for k, v := range g.Claims {
		if v == nil {
			delete(c, k)
		} else {
			c[k] = v
		}
	}
	return c
}

// Sign mints an RS256 token with the current key and its kid.
func (f *Fake) Sign(claims map[string]any) string {
	f.mu.Lock()
	k := f.keys[0]
	f.mu.Unlock()
	return SignWith(k.key, "RS256", k.kid, claims)
}

// CurrentKID returns the kid of the current signing key.
func (f *Fake) CurrentKID() string { f.mu.Lock(); defer f.mu.Unlock(); return f.keys[0].kid }

// SignWith mints a token with an arbitrary RSA key, header alg and kid. alg "none" produces an unsigned token and
// "HS256" an HMAC token keyed with the public modulus (key-confusion attempts).
func SignWith(key *rsa.PrivateKey, alg, kid string, claims map[string]any) string {
	header := map[string]any{"alg": alg, "typ": "JWT"}
	if kid != "" {
		header["kid"] = kid
	}
	hb, _ := json.Marshal(header)
	cb, _ := json.Marshal(claims)
	input := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(cb)
	switch alg {
	case "none":
		return input + "."
	case "HS256":
		return input + "." + base64.RawURLEncoding.EncodeToString(hmacSHA256(key.N.Bytes(), []byte(input)))
	}
	sum := sha256.Sum256([]byte(input))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		panic(err)
	}
	return input + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// Tamper flips one character of the payload segment, keeping the signature.
func Tamper(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || len(parts[1]) < 4 {
		return token
	}
	p := []byte(parts[1])
	if p[2] == 'A' {
		p[2] = 'B'
	} else {
		p[2] = 'A'
	}
	parts[1] = string(p)
	return strings.Join(parts, ".")
}

// NewAttackerKey returns a fresh RSA key not published by the fake.
func NewAttackerKey(t testing.TB) *rsa.PrivateKey { t.Helper(); return newKey(t).key }

func hmacSHA256(key, msg []byte) []byte {
	m := hmac.New(sha256.New, key)
	m.Write(msg)
	return m.Sum(nil)
}

// --- G014: Connect issuer mode -------------------------------------------------------------------------------------

// ConnectIssuer is the fake's Connect issuer (its base URL): the MCP_CONNECT_ISSUER of tests, whose key set is
// <issuer>/oauth2/jwks.
func (f *Fake) ConnectIssuer() string { return f.Server.URL }

func (f *Fake) handleConnectJWKS(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.jwksHits++
	status, body, delay := f.jwksStatus, f.jwksBody, f.jwksDelay
	keys := f.keys
	if f.unpublishNext && len(keys) > 1 {
		keys = keys[1:]
	}
	f.mu.Unlock()
	if delay > 0 {
		time.Sleep(delay)
	}
	w.Header().Set("Content-Type", "application/json")
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	if body == nil {
		body = jwks(keys)
	}
	_, _ = w.Write(body)
}

// ConnectJWKS renders the currently published key set (the test trust anchor of MCP_CONNECT_JWKS_FILE).
func (f *Fake) ConnectJWKS() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	keys := f.keys
	if f.unpublishNext && len(keys) > 1 {
		keys = keys[1:]
	}
	return jwks(keys)
}

// ConnectClaims returns default Connect access-token claims: the Connect issuer, the WorkOS user subject, one exact
// audience, the client, iat now and exp in one hour, plus ignored vendor claims (sid, org_id, scope, auth_time).
func (f *Fake) ConnectClaims(subject, client, audience string) map[string]any {
	f.mu.Lock()
	now := f.now()
	f.mu.Unlock()
	return map[string]any{
		"iss": f.ConnectIssuer(), "sub": subject, "aud": audience, "client_id": client,
		"iat": now.Unix(), "exp": now.Add(time.Hour).Unix(), "auth_time": now.Unix(),
		"sid": "session_" + randomHex(8), "org_id": "org_vendor", "scope": "openid offline_access", "jti": "jti_" + randomHex(6),
	}
}

// SignConnect mints a Connect access token with the current key for subject, client and audience; overrides replace
// claims (a nil value deletes one).
func (f *Fake) SignConnect(subject, client, audience string, overrides map[string]any) string {
	c := f.ConnectClaims(subject, client, audience)
	for k, v := range overrides {
		if v == nil {
			delete(c, k)
		} else {
			c[k] = v
		}
	}
	return f.Sign(c)
}
