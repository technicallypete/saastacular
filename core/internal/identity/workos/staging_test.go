//go:build workos_staging

// L3: the real WorkOS staging API (Plan r5 §8.2 S-JWKS, S-MFA, S-ATTR, S-TAMP, S-CODE, S-REV, S-CLEAN). It runs in
// the protected identity-staging CI job (B-IDENTITY-STAGING) and, as CAP-3, on the operator workstation through the
// IDENTITY.md "identity:staging-local" block. It never runs in an ordinary profile: the tag and the credentials are
// only present there, and missing credentials fail (UNMET), never skip.
//
// Only synthetic per-run users created by this test are touched; each is positively owned (created here, recorded,
// deleted in cleanup). Secrets (API key, passwords, TOTP secrets, tokens, codes) stay in process memory and are never
// logged; the sanitized result file records only booleans, shapes and fingerprints.
//
// Environment:
//
//	IDENTITY_STAGING_WORKOS_API_KEY, IDENTITY_STAGING_WORKOS_CLIENT_ID      required
//	IDENTITY_STAGING_WORKOS_ISSUER                                         required, unless
//	IDENTITY_STAGING_ESTABLISH_ISSUER=1 (CAP-3): the exact issuer is taken from a real token obtained here over TLS
//	                                       and written to IDENTITY_STAGING_ISSUER_OUT (0600)
//	IDENTITY_STAGING_RESULT_FILE           optional sanitized JSON result
//	IDENTITY_STAGING_L3_EMAIL_DOMAIN       optional; default example.org (a reserved domain; no mail is ever sent:
//	                                       users are created verified and never asked to verify). Not example.com:
//	                                       WorkOS staging environments ship a test organization whose SSO
//	                                       connection owns example.com, so its users answer sso_required.
//	DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD, APP_ENV  for the mapping assertions
//
// TestStagingDomainQualification (G016 P-1 preparation; the actual evidence belongs to G057) is a separate top-level
// test with its own inputs, documented above it. The selector '^TestStagingContract$' never runs it, and
// TestStagingContract never reads its inputs. TestStagingDomainObservationFixtures is offline (httptest only).
package workos

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sdk "github.com/workos/workos-go/v10"

	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/migrate"
	"core/migrations"
)

const stgBase = "https://api.workos.com"

type stgResult map[string]any

func stgEnv(t *testing.T, name string, required bool) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" && required {
		t.Fatalf("UNMET: %s is required for the L3 staging checks", name)
	}
	return v
}

func stgRandom(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func fingerprint(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:8])
}

// totp is RFC 6238 (SHA-1, 30 s, 6 digits), the parameters WorkOS documents for TOTP factors.
func totp(t *testing.T, secret string, at time.Time) string {
	t.Helper()
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimRight(secret, "=")))
	if err != nil {
		t.Fatal("TOTP secret is not base32")
	}
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(at.Unix()/30))
	m := hmac.New(sha1.New, key)
	m.Write(msg[:])
	h := m.Sum(nil)
	off := h[len(h)-1] & 0x0f
	v := (binary.BigEndian.Uint32(h[off:off+4]) & 0x7fffffff) % 1000000
	return fmt.Sprintf("%06d", v)
}

// stgUser is one positively owned per-run synthetic user.
type stgUser struct {
	id, email, password, totpSecret, factorID string
	enrollChallenge                           string // the challenge returned with the enrollment (first sign-in)
	lastStep                                  int64  // the TOTP time step last used (codes are single-use per step)
}

// nextCode waits, if needed, for a TOTP time step this user has not used yet, then returns its code.
func (r *stgRun) nextCode(u *stgUser) string {
	for {
		now := time.Now()
		if step := now.Unix() / 30; step != u.lastStep {
			u.lastStep = step
			return totp(r.t, u.totpSecret, now)
		}
		time.Sleep(time.Until(time.Unix((now.Unix()/30+1)*30, 0)) + 500*time.Millisecond)
	}
}

type stgRun struct {
	t      *testing.T
	client *sdk.Client
	a      *Adapter
	tag    string
	domain string
	owned  []string
	res    stgResult
}

func (r *stgRun) createUser(label, first, email string) *stgUser {
	r.t.Helper()
	ctx := context.Background()
	if email == "" {
		email = "g004-l3-" + r.tag + "-" + label + "@" + r.domain
	}
	u := &stgUser{email: email, password: "L3-" + stgRandom(16) + "-Aa1!"}
	verified := true
	created, err := r.client.UserManagement().Create(ctx, &sdk.UserManagementCreateParams{
		Email: u.email, FirstName: &first, LastName: &label, EmailVerified: &verified,
		Password: sdk.UserManagementPasswordPlaintext{Password: &u.password},
	})
	if err != nil {
		r.t.Fatalf("create per-run user %s: %v", label, classify(err))
	}
	u.id = created.ID
	r.owned = append(r.owned, u.id)
	issuer := "g004-l3"
	enr, err := r.client.MultiFactorAuth().CreateUserAuthFactor(ctx, u.id, &sdk.MultiFactorAuthCreateUserAuthFactorParams{Type: "totp", TOTPIssuer: &issuer, TOTPUser: &u.email})
	if err != nil || enr.AuthenticationFactor == nil || enr.AuthenticationFactor.TOTP == nil || enr.AuthenticationFactor.TOTP.Secret == "" {
		r.t.Fatalf("enroll TOTP for %s: %v", label, classify(err))
	}
	u.totpSecret, u.factorID = enr.AuthenticationFactor.TOTP.Secret, enr.AuthenticationFactor.ID
	if enr.AuthenticationChallenge != nil {
		u.enrollChallenge = enr.AuthenticationChallenge.ID
	}
	return u
}

// classify reduces a provider error to its class (status and code) without bodies or values.
func classify(err error) string {
	if err == nil {
		return "nil"
	}
	var api *sdk.APIError
	if errors.As(err, &api) {
		return fmt.Sprintf("status %d code %s%s", api.StatusCode, api.Code, api.ErrorCode)
	}
	return "network or decode error"
}

// signIn performs password then TOTP authentication. It returns whether the password step demanded MFA and the
// authenticate response of the TOTP step.
func (r *stgRun) signIn(u *stgUser) (bool, *sdk.AuthenticateResponse) {
	r.t.Helper()
	ctx := context.Background()
	resp, err := r.client.UserManagement().AuthenticateWithPassword(ctx, &sdk.UserManagementAuthenticateWithPasswordParams{Email: u.email, Password: u.password})
	if err == nil {
		if resp != nil && resp.AccessToken != "" {
			return false, resp // MFA was not required: tokens without a second factor
		}
		r.t.Fatal("password step returned neither tokens nor an MFA challenge")
	}
	// The MFA-challenge error class: mfa_enrollment (the API-enrolled factor is not yet verified; first sign-in) or
	// mfa_challenge (later sign-ins). Both carry a pending authentication token and no tokens.
	var raw string
	var challenge *sdk.MFAChallengeError
	var enrollment *sdk.MFAEnrollmentError
	switch {
	case errors.As(err, &challenge):
		raw = challenge.APIError.RawBody
	case errors.As(err, &enrollment):
		raw = enrollment.APIError.RawBody
	default:
		r.t.Fatalf("password step: %s (Radar, rate limit or configuration: STOP and report)", classify(err))
	}
	var body struct {
		Pending string `json:"pending_authentication_token"`
	}
	if json.Unmarshal([]byte(raw), &body) != nil || body.Pending == "" {
		r.t.Fatal("MFA challenge without a pending authentication token")
	}
	// First sign-in (unverified factor): the enrollment's own challenge. Later sign-ins: a new challenge on the
	// verified factor.
	challengeID := u.enrollChallenge
	if enrollment == nil || challengeID == "" {
		ch, err := r.client.MultiFactorAuth().ChallengeFactor(ctx, u.factorID, &sdk.MultiFactorAuthChallengeFactorParams{})
		if err != nil {
			r.t.Fatalf("challenge factor: %s", classify(err))
		}
		challengeID = ch.ID
	}
	u.enrollChallenge = ""
	resp, err = r.client.UserManagement().AuthenticateWithTOTP(ctx, &sdk.UserManagementAuthenticateWithTOTPParams{
		Code: r.nextCode(u), PendingAuthenticationToken: body.Pending, AuthenticationChallengeID: challengeID})
	if err != nil || resp == nil || resp.AccessToken == "" {
		r.t.Fatalf("TOTP step: %s", classify(err))
	}
	return true, resp
}

// verifiedProvider returns one already validated real identity from Exchange, so the production service maps it
// through its normal callback path (binding, consumption, mapping, session).
type verifiedProvider struct{ vi identity.VerifiedIdentity }

func (p verifiedProvider) AuthorizationURL(_ context.Context, r identity.AuthRequest) (string, error) {
	return stgBase + "/user_management/authorize?state=" + url.QueryEscape(r.State), nil
}
func (p verifiedProvider) Exchange(context.Context, identity.Exchange) (identity.VerifiedIdentity, error) {
	return p.vi, nil
}
func (verifiedProvider) RevokeProviderSession(context.Context, string) error { return nil }
func (verifiedProvider) ProviderLogoutURL(string, string) (string, error)    { return "", nil }

func stgPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	cfg, err := migrate.LoadConfig(os.Getenv)
	if err != nil || os.Getenv("DB_RUNTIME_PASSWORD") == "" {
		t.Fatal("UNMET: database settings for the mapping assertions are required")
	}
	r, err := migrate.Open(ctx, cfg, migrations.FS)
	if err != nil {
		t.Fatalf("open migrator: %v", err)
	}
	if err := r.Up(ctx, io.Discard); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	_ = r.Close()
	cc, _ := pgx.ParseConfig("")
	cc.Host, cc.Port, cc.Database, cc.User, cc.Password = cfg.Host, uint16(cfg.Port), cfg.Database, cfg.MigratorRole(), cfg.Password
	cc.TLSConfig, cc.Fallbacks = nil, nil
	own, err := pgx.ConnectConfig(ctx, cc)
	if err != nil {
		t.Fatal(err)
	}
	defer own.Close(ctx)
	if _, err := own.Exec(ctx, "SET ROLE "+pgx.Identifier{cfg.OwnerRole()}.Sanitize()); err != nil {
		t.Fatal(err)
	}
	if _, err := own.Exec(ctx, `TRUNCATE app.login_transactions, app.sessions, app.identity_principals, app.users, app.identity_provider_binding CASCADE`); err != nil {
		t.Fatal(err)
	}
	pc, _ := pgxpool.ParseConfig("")
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = cfg.Host, uint16(cfg.Port), cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = cfg.RuntimeRole(), os.Getenv("DB_RUNTIME_PASSWORD")
	p, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

// mapIdentity runs the production login and callback path for an already validated identity and returns the
// resolved session.
func mapIdentity(t *testing.T, pool *pgxpool.Pool, issuer, client string, vi identity.VerifiedIdentity) identity.SessionInfo {
	t.Helper()
	ctx := context.Background()
	origin := "https://org-a.localhost:543"
	svc := identity.NewService(identity.NewStore(pool), identity.Config{Provider: verifiedProvider{vi: vi}, Issuer: issuer, ClientID: client,
		Origins: identity.NewOrigins([]string{origin}), Idle: 30 * time.Minute, Max: 8 * time.Hour})
	out, err := svc.Login(ctx, identity.LoginInput{Origin: origin})
	if err != nil {
		t.Fatalf("login start: %v", err)
	}
	u, _ := url.Parse(out.AuthorizationURL)
	cb, err := svc.Callback(ctx, identity.CallbackInput{Origin: origin, RawQuery: "code=l3&state=" + url.QueryEscape(u.Query().Get("state")), Binding: out.Binding})
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	info, err := svc.Session(ctx, cb.Credential, origin)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	return info
}

func TestStagingContract(t *testing.T) {
	ctx := context.Background()
	key := stgEnv(t, "IDENTITY_STAGING_WORKOS_API_KEY", true)
	client := stgEnv(t, "IDENTITY_STAGING_WORKOS_CLIENT_ID", true)
	establish := os.Getenv("IDENTITY_STAGING_ESTABLISH_ISSUER") == "1"
	issuer := stgEnv(t, "IDENTITY_STAGING_WORKOS_ISSUER", !establish)
	domain := stgEnv(t, "IDENTITY_STAGING_L3_EMAIL_DOMAIN", false)
	if domain == "" {
		domain = "example.org"
	}
	res := stgResult{"run": time.Now().UTC().Format(time.RFC3339), "client_fingerprint": "client_" + fingerprint(client), "mode": map[bool]string{true: "establish-issuer (CAP-3)", false: "configured-issuer"}[establish]}
	t.Cleanup(func() {
		if p := os.Getenv("IDENTITY_STAGING_RESULT_FILE"); p != "" {
			b, _ := json.MarshalIndent(res, "", "  ")
			_ = os.WriteFile(p, append(b, '\n'), 0o600)
		}
	})
	sdkClient := sdk.NewClient(key, sdk.WithBaseURL(stgBase), sdk.WithClientID(client),
		sdk.WithHTTPClient(boundedClient(nil, 15*time.Second, authenticateMaxBdy)), sdk.WithMaxRetries(0))
	r := &stgRun{t: t, client: sdkClient, tag: stgRandom(4), domain: domain, res: res}

	// S-CLEAN: delete every per-run user this test created, then confirm each is gone.
	t.Cleanup(func() {
		deleted := 0
		for _, id := range r.owned {
			if err := sdkClient.UserManagement().Delete(ctx, id); err == nil {
				deleted++
			}
			if _, err := sdkClient.UserManagement().Get(ctx, id); err == nil {
				t.Errorf("S-CLEAN: per-run user still present after deletion")
			}
		}
		res["s_clean"] = map[string]int{"created": len(r.owned), "deleted": deleted}
	})

	// S-JWKS: the client JWKS parses; the algorithm is pinned.
	keys := newKeySource(stgBase+"/sso/jwks/"+url.PathEscape(client), boundedClient(nil, jwksFetchDeadline, jwksMaxBody), time.Now)
	if err := keys.fetch(ctx); err != nil {
		t.Fatalf("S-JWKS: %v", err)
	}
	res["s_jwks"] = "pass"

	// S-MFA and CAP-3: password alone yields an MFA challenge and no tokens; TOTP yields a token that the production
	// validator accepts.
	alice := r.createUser("a", "Synthetic", "")
	mfa, resp := r.signIn(alice)
	if !mfa {
		res["s_mfa"] = "FAIL: tokens without MFA"
		t.Fatal("S-MFA: the tested configuration issued tokens without a second factor (MFA Required not enforced)")
	}
	res["s_mfa"] = map[string]any{"password_only_tokens": false, "challenge_required": true, "observed_at": time.Now().UTC().Format(time.RFC3339)}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	var unverified struct {
		Iss string `json:"iss"`
	}
	parts := strings.Split(resp.AccessToken, ".")
	if len(parts) != 3 {
		t.Fatal("CAP-3: access token is not a JWS")
	}
	hb, _ := base64.RawURLEncoding.DecodeString(parts[0])
	pb, _ := base64.RawURLEncoding.DecodeString(parts[1])
	_ = json.Unmarshal(hb, &header)
	_ = json.Unmarshal(pb, &unverified)
	if establish {
		// The issuer is taken only from a token core itself obtained from the configured TLS endpoint with the API key,
		// and only after that token passes signature, client, subject and time validation below.
		issuer = unverified.Iss
	}
	if issuer != unverified.Iss {
		t.Fatalf("CAP-3 V-1: the token issuer differs from IDENTITY_STAGING_WORKOS_ISSUER")
	}
	a, err := New(Config{APIKey: key, ClientID: client, Issuer: issuer, BaseURL: stgBase, Algorithms: []string{"RS256"}})
	if err != nil {
		t.Fatal(err)
	}
	r.a = a
	vi, err := a.verify(ctx, resp)
	if err != nil {
		t.Fatalf("CAP-3: the production validator rejected a real staging token: %v", err)
	}
	iu, _ := url.Parse(issuer)
	// The shape never carries a client identifier: each client_ path segment is replaced by its fingerprint.
	segs := strings.Split(iu.Path, "/")
	issuerClient := ""
	for i, s := range segs {
		if strings.HasPrefix(s, "client_") {
			issuerClient = s
			segs[i] = "<client_" + fingerprint(s) + ">"
		}
	}
	res["cap3"] = map[string]any{
		"alg": header.Alg, "kid_present": header.Kid != "",
		"issuer_shape":                      iu.Scheme + "://" + iu.Host + strings.Join(segs, "/"),
		"issuer_client_equals_configured":   issuerClient == client,
		"token_client_id_equals_configured": vi.ClientID == client,
		"issuer_has_trailing_slash":         strings.HasSuffix(issuer, "/"),
		"sub_is_user_id":                    vi.Subject == alice.id, "sub_prefix_user_": strings.HasPrefix(vi.Subject, "user_"),
		"sid_present": vi.ProviderSessionID != "", "auth_time_recent": time.Since(vi.AuthTime) < 5*time.Minute,
	}
	if header.Alg != "RS256" || header.Kid == "" || vi.Subject != alice.id {
		t.Fatalf("CAP-3: alg %q, kid present %v, sub matches %v", header.Alg, header.Kid != "", vi.Subject == alice.id)
	}
	if establish {
		if out := os.Getenv("IDENTITY_STAGING_ISSUER_OUT"); out != "" {
			if err := os.WriteFile(out, []byte(issuer+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}

	pool := stgPool(t)
	sess := mapIdentity(t, pool, issuer, client, vi)
	if sess.MFAAssurance != identity.MFAAssuranceUnknown || identity.RequireVerifiedMFA(sess) != identity.ErrMFAUnverified {
		t.Fatalf("S-MFA: mapped session assurance %q", sess.MFAAssurance)
	}
	res["s_mfa_mapping"] = map[string]any{"mfa_assurance": sess.MFAAssurance, "verified_mfa_requirement": "denied"}

	// S-TAMP: modified payload or signature, wrong expected client, shifted clock and wrong issuer are rejected.
	tamper := map[string]error{}
	_, tamper["payload"] = a.tokens.validate(ctx, identitytest.Tamper(resp.AccessToken), alice.id)
	sig := []byte(parts[2])
	sig[len(sig)/2] ^= 1
	_, tamper["signature"] = a.tokens.validate(ctx, parts[0]+"."+parts[1]+"."+string(sig), alice.id)
	wrongClient := &tokenValidator{issuer: issuer, clientID: client + "x", algorithms: []string{"RS256"}, keys: a.tokens.keys, now: time.Now}
	_, tamper["client"] = wrongClient.validate(ctx, resp.AccessToken, alice.id)
	late := &tokenValidator{issuer: issuer, clientID: client, algorithms: []string{"RS256"}, keys: a.tokens.keys, now: func() time.Time { return time.Now().Add(2 * time.Hour) }}
	_, tamper["clock"] = late.validate(ctx, resp.AccessToken, alice.id)
	wrongIssuer := &tokenValidator{issuer: issuer + "/", clientID: client, algorithms: []string{"RS256"}, keys: a.tokens.keys, now: time.Now}
	_, tamper["issuer"] = wrongIssuer.validate(ctx, resp.AccessToken, alice.id)
	for name, err := range tamper {
		if err == nil {
			t.Errorf("S-TAMP %s accepted", name)
		}
	}
	res["s_tamp"] = "pass"

	// S-CODE: a random code with a random verifier is code_rejected; no session.
	if _, err := a.Exchange(ctx, identity.Exchange{Code: "code_" + stgRandom(12), CodeVerifier: stgRandom(32)}); err != identity.ErrCodeRejected {
		t.Fatalf("S-CODE: %v", err)
	}
	res["s_code"] = "pass"

	// S-REV: RevokeSession(sid) is confirmed; a repeat maps to the documented error class.
	if err := a.RevokeProviderSession(ctx, vi.ProviderSessionID); err != nil {
		t.Fatalf("S-REV: %v", err)
	}
	repeat := a.RevokeProviderSession(ctx, vi.ProviderSessionID)
	repeatClass := "accepted (idempotent)"
	if repeat != nil {
		repeatClass = identity.AsError(repeat).Code
	}
	res["s_rev"] = map[string]any{"revoke": "confirmed", "repeat": repeatClass}

	// S-ATTR: change A's name and email and re-authenticate (same subject, same UUID, display updated); create B with
	// A's former email (different subject, different UUID). No email is requested or stored.
	former := alice.email
	renamed := "Renamed"
	newEmail := "g004-l3-" + r.tag + "-a2@" + r.domain
	verified := true // synthetic address: marked verified so no verification mail is ever sent
	if _, err := sdkClient.UserManagement().Update(ctx, alice.id, &sdk.UserManagementUpdateParams{FirstName: &renamed, Email: &newEmail, EmailVerified: &verified}); err != nil {
		t.Fatalf("S-ATTR update: %s", classify(err))
	}
	alice.email = newEmail
	_, resp2 := r.signIn(alice)
	vi2, err := a.verify(ctx, resp2)
	if err != nil {
		t.Fatalf("S-ATTR re-authentication: %v", err)
	}
	sess2 := mapIdentity(t, pool, issuer, client, vi2)
	if vi2.Subject != vi.Subject || sess2.UserID != sess.UserID || !strings.HasPrefix(sess2.DisplayName, "Renamed") {
		t.Fatalf("S-ATTR rename: same subject %v, same UUID %v", vi2.Subject == vi.Subject, sess2.UserID == sess.UserID)
	}
	bob := r.createUser("b", "Second", former) // a new principal created with A's former email
	_, respB := r.signIn(bob)
	viB, err := a.verify(ctx, respB)
	if err != nil {
		t.Fatalf("S-ATTR second principal: %v", err)
	}
	sessB := mapIdentity(t, pool, issuer, client, viB)
	if viB.Subject == vi.Subject || sessB.UserID == sess.UserID {
		t.Fatal("S-ATTR: the reassigned email reached the first principal's UUID")
	}
	res["s_attr"] = map[string]any{"rename_same_uuid": true, "reassigned_email_distinct_subject": true, "reassigned_email_distinct_uuid": true}
	res["result"] = "pass"
}

// TestStagingDomainQualification prepares the G016 P-1 provider evidence (Plan §6.2, §6.4; actual evidence is G057's):
// WorkOS's own acceptance of a wildcard-covered platform host and of the sign-out return rule as configured. Select it
// explicitly with -run '^TestStagingDomainQualification$'. When selected, every input is required (missing or
// malformed is UNMET, never a skip):
//
//	IDENTITY_STAGING_WORKOS_API_KEY, IDENTITY_STAGING_WORKOS_CLIENT_ID, IDENTITY_STAGING_WORKOS_ISSUER  (configured issuer)
//	IDENTITY_STAGING_PLATFORM_DOMAIN     the platform domain B (entries https://*.B/auth/callback and https://*.B/sign-in)
//	IDENTITY_STAGING_AUTH_ORIGIN         the shared auth origin A (entries A/auth/callback and A/sign-in)
//	IDENTITY_STAGING_AUTHKIT_ORIGIN      the verified origin of WorkOS's hosted sign-in: where the provider sends the
//	                                     browser for an accepted authorization request (not derivable from the others)
//	IDENTITY_STAGING_AUTHKIT_PATH_PREFIX the closed path allow-list of that first hop (G057 Plan r5 §3.6): the path, or
//	                                     path prefix at a segment boundary, established at EP-2 from the first verified
//	                                     observation; a hop to the right origin on another path is unknown, never accepted
//	IDENTITY_STAGING_DOMAIN_RESULT_FILE  optional sanitized JSON result
//
// Each observation is WorkOS's response to an unauthenticated browser request (no redirect followed, no cookie). The
// probe hosts are a per-run random label under B (no exact dashboard entry can name it, so only the wildcard can admit
// it) and a reserved .invalid name as the negative control (RFC 2606). An unknown response qualifies as neither
// acceptance nor refusal. It creates and deletes one per-run user and touches no database. URLs, Location values,
// codes, state, provider bodies and session IDs are never logged or recorded. The code-bearing redirect after an
// interactive hosted sign-in is outside this API-password harness (a browser step).
//
// Cleanup (G057 §3.6) is fail-closed: a provider 404 (or its explicit not-found class) after the deletion is confirmed
// absence; any other outcome is unknown or present, recorded under "recovery" with the per-run user id only, and the
// "result" member is written only after cleanup has finished.
func TestStagingDomainQualification(t *testing.T) {
	ctx := context.Background()
	key := stgEnv(t, "IDENTITY_STAGING_WORKOS_API_KEY", true)
	client := stgEnv(t, "IDENTITY_STAGING_WORKOS_CLIENT_ID", true)
	issuer := stgEnv(t, "IDENTITY_STAGING_WORKOS_ISSUER", true)
	in := stgDomainInputsFromEnv(t)
	domain := stgEnv(t, "IDENTITY_STAGING_L3_EMAIL_DOMAIN", false)
	if domain == "" {
		domain = "example.org"
	}
	res := stgResult{"run": time.Now().UTC().Format(time.RFC3339), "client_fingerprint": "client_" + fingerprint(client),
		"authkit_origin_fingerprint": "origin_" + fingerprint(in.authkit.Host), "authkit_path_prefix": in.authkitPath}
	t.Cleanup(func() {
		if p := os.Getenv("IDENTITY_STAGING_DOMAIN_RESULT_FILE"); p != "" {
			b, _ := json.MarshalIndent(res, "", "  ")
			_ = os.WriteFile(p, append(b, '\n'), 0o600)
		}
	})
	sdkClient := sdk.NewClient(key, sdk.WithBaseURL(stgBase), sdk.WithClientID(client),
		sdk.WithHTTPClient(boundedClient(nil, 15*time.Second, authenticateMaxBdy)), sdk.WithMaxRetries(0))
	a, err := New(Config{APIKey: key, ClientID: client, Issuer: issuer, BaseURL: stgBase, Algorithms: []string{"RS256"}})
	if err != nil {
		t.Fatal(err)
	}
	r := &stgRun{t: t, client: sdkClient, a: a, tag: stgRandom(4), domain: domain, res: res}
	passed := false
	t.Cleanup(func() { // as S-CLEAN: delete every per-run user created here, then confirm each is gone (fail closed)
		deleted := 0
		verdicts := map[string]int{}
		var recovery []string
		for _, id := range r.owned {
			if err := sdkClient.UserManagement().Delete(ctx, id); err == nil {
				deleted++
			}
			_, err := sdkClient.UserManagement().Get(ctx, id)
			v := stgAbsence(err)
			verdicts[v]++
			if v != stgAbsent {
				t.Errorf("clean: per-run user %s after deletion (the recovery reference carries its id)", v)
				recovery = append(recovery, id)
			}
		}
		res["clean"] = map[string]int{"created": len(r.owned), "deleted": deleted, "absent": verdicts[stgAbsent],
			"unknown": verdicts[stgUnknownAbsence], "present": verdicts[stgPresent]}
		if len(recovery) > 0 {
			res["recovery"] = map[string]any{"per_run_user_ids": recovery, "action": "delete and confirm absence before the next run"}
		}
		// "result" is written only now, after cleanup: a pass needs both subtests and every per-run user confirmed absent.
		switch {
		case passed && len(recovery) == 0:
			res["result"] = "pass"
		case passed:
			res["result"] = "fail: cleanup unknown"
		default:
			res["result"] = "fail"
		}
	})
	u := r.createUser("d", "Synthetic", "")

	callbackOK := t.Run("WildcardCallback", func(t *testing.T) { r.sub(t).domainCallback(ctx, in) })
	signOutOK := t.Run("SignOutReturn", func(t *testing.T) { r.sub(t).domainSignOut(ctx, in, u) })
	passed = callbackOK && signOutOK
}

// The cleanup verdict of one per-run user after its deletion (G057 §3.6): present (the provider still returns it),
// absent (404 or the provider's explicit not-found class) or unknown (any other error: fail closed, never claimed).
const (
	stgPresent        = "present"
	stgAbsent         = "absent"
	stgUnknownAbsence = "unknown"
)

func stgAbsence(getErr error) string {
	if getErr == nil {
		return stgPresent
	}
	var api *sdk.APIError
	if errors.As(getErr, &api) && (api.StatusCode == http.StatusNotFound || fmt.Sprint(api.Code) == "entity_not_found") {
		return stgAbsent
	}
	return stgUnknownAbsence
}

// sub returns the run state for a subtest, bound to the subtest's own T (Fatal never targets the parent).
func (r *stgRun) sub(t *testing.T) *stgRun {
	c := *r
	c.t = t
	return &c
}

type stgDomainInputs struct {
	base        string   // the platform domain B
	authOrigin  string   // the shared auth origin A
	authHost    string   // A's host name
	authkit     *url.URL // the verified hosted sign-in origin
	authkitPath string   // the closed path allow-list of the first hop (a path or segment-boundary prefix)
}

func stgDomainInputsFromEnv(t *testing.T) stgDomainInputs {
	t.Helper()
	var in stgDomainInputs
	in.base = stgEnv(t, "IDENTITY_STAGING_PLATFORM_DOMAIN", true)
	if !stgDomainName(in.base) {
		t.Fatal("UNMET: IDENTITY_STAGING_PLATFORM_DOMAIN is not a lowercase multi-label DNS name outside localhost")
	}
	in.authOrigin = stgEnv(t, "IDENTITY_STAGING_AUTH_ORIGIN", true)
	au, ok := stgOrigin(in.authOrigin)
	if !ok {
		t.Fatal("UNMET: IDENTITY_STAGING_AUTH_ORIGIN is not an exact lowercase https origin")
	}
	in.authHost = au.Hostname()
	if in.authkit, ok = stgOrigin(stgEnv(t, "IDENTITY_STAGING_AUTHKIT_ORIGIN", true)); !ok {
		t.Fatal("UNMET: IDENTITY_STAGING_AUTHKIT_ORIGIN is not an exact lowercase https origin")
	}
	if in.authkitPath = stgEnv(t, "IDENTITY_STAGING_AUTHKIT_PATH_PREFIX", true); !stgPathPrefix(in.authkitPath) {
		t.Fatal("UNMET: IDENTITY_STAGING_AUTHKIT_PATH_PREFIX is not an absolute path without query, fragment or whitespace")
	}
	return in
}

// stgPathPrefix reports whether s is usable as the closed first-hop path allow-list: an absolute path ("/" alone
// included) with no query, fragment, whitespace or empty segment, at most 200 bytes.
func stgPathPrefix(s string) bool {
	if s == "" || s[0] != '/' || len(s) > 200 || strings.ContainsAny(s, "?# \t\r\n") || strings.Contains(s, "//") {
		return false
	}
	return s == "/" || !strings.HasSuffix(s, "/")
}

// stgPathAllowed reports whether a hop's path is within the allow-list: the path itself, or a longer path below it at
// a segment boundary ("/" admits every path). An empty path is the root. Comparison is exact (case included).
// The path is the decoded, un-normalized hop path, so a browser would resolve a dot segment ("." or "..", literal or
// percent-encoded) or a backslash separator to a path outside the allow-list: such a path is never allowed.
func stgPathAllowed(path, prefix string) bool {
	if path == "" {
		path = "/"
	}
	if strings.Contains(path, `\`) {
		return false
	}
	for _, seg := range strings.Split(path, "/") {
		if seg == "." || seg == ".." {
			return false
		}
	}
	if prefix == "/" {
		return strings.HasPrefix(path, "/")
	}
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

// stgDomainName reports whether s is a lowercase multi-label DNS name outside localhost.
func stgDomainName(s string) bool {
	labels := strings.Split(s, ".")
	if len(labels) < 2 || len(s) > 200 || labels[len(labels)-1] == "localhost" {
		return false
	}
	for _, l := range labels {
		if l == "" || len(l) > 63 || l[0] == '-' || l[len(l)-1] == '-' {
			return false
		}
		for _, c := range l {
			if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '-' {
				return false
			}
		}
	}
	return true
}

// stgOrigin parses an exact lowercase https origin: a DNS host name, an optional port, and no userinfo, path, query
// or fragment.
func stgOrigin(s string) (*url.URL, bool) {
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Opaque != "" || strings.HasSuffix(u.Host, ":") ||
		!stgDomainName(u.Hostname()) || s != "https://"+u.Host {
		return nil, false
	}
	return u, true
}

// stgObservation is one provider response: its status, the raw Location value and its resolution (nil without one).
// Neither is ever logged.
type stgObservation struct {
	status int
	loc    string
	dest   *url.URL
}

func (o stgObservation) redirected() bool { return o.status >= 300 && o.status < 400 && o.dest != nil }

// toHost reports whether the response redirects the browser to host (whatever userinfo, port, path or query).
func (o stgObservation) toHost(host string) bool {
	return o.redirected() && strings.EqualFold(o.dest.Hostname(), host)
}

// stgObserve sends one GET without following redirects and reads the whole bounded body. Any transport or body-read
// failure (bound exceeded, including a chunked overflow; a timeout after the headers) is an error. Errors carry only a
// class: a transport error's text holds the URL, with its state, client and session identifiers.
func stgObserve(ctx context.Context, hc *http.Client, raw string) (stgObservation, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return stgObservation{}, errors.New("request: could not be built")
	}
	resp, err := hc.Do(req)
	if err != nil {
		return stgObservation{}, errors.New("request: " + stgErrClass(err))
	}
	_, err = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return stgObservation{}, errors.New("body: " + stgErrClass(err))
	}
	o := stgObservation{status: resp.StatusCode}
	switch locs := resp.Header.Values("Location"); len(locs) {
	case 0:
	case 1:
		u, err := req.URL.Parse(locs[0])
		if err != nil {
			return stgObservation{}, errors.New("response: unparseable Location")
		}
		o.loc, o.dest = locs[0], u
	default:
		return stgObservation{}, errors.New("response: more than one Location")
	}
	return o, nil
}

func stgErrClass(err error) string {
	var ne net.Error
	switch {
	case errors.Is(err, errBodyTooLarge):
		return "response bound exceeded"
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &ne) && ne.Timeout():
		return "timeout"
	default:
		return "network or read error"
	}
}

func (r *stgRun) observe(ctx context.Context, raw string) stgObservation {
	r.t.Helper()
	o, err := stgObserve(ctx, boundedClient(nil, 15*time.Second, 1<<20), raw)
	if err != nil {
		r.t.Fatalf("provider observation failed: %v", err)
	}
	return o
}

type stgVerdict string

const (
	stgAccepted stgVerdict = "accepted"
	stgRefused  stgVerdict = "refused"
	stgUnknown  stgVerdict = "unknown"
)

// stgAuthorizeVerdict classifies an authorization response for a redirect_uri on redirectHost. Accepted: a redirect to
// the verified hosted sign-in origin exactly (https, same host and port, no userinfo, no OAuth error parameter) on a
// path within the closed allow-list pathPrefix (G057 §3.6: origin and path shape, not origin only). Refused: a 4xx,
// or a redirect carrying an OAuth error that does not lead to redirectHost. Anything else (another destination or
// path, a redirect to redirectHost, 2xx, 5xx) is unknown and qualifies as neither.
func stgAuthorizeVerdict(o stgObservation, authkit *url.URL, pathPrefix, redirectHost string) stgVerdict {
	if o.status >= 400 && o.status < 500 {
		return stgRefused
	}
	if !o.redirected() || o.toHost(redirectHost) || o.dest.User != nil || o.dest.Scheme != "https" {
		return stgUnknown
	}
	if q := o.dest.Query(); q.Has("error") || q.Has("error_description") {
		return stgRefused
	}
	if strings.EqualFold(o.dest.Host, authkit.Host) && stgPathAllowed(o.dest.Path, pathPrefix) {
		return stgAccepted
	}
	return stgUnknown
}

// stgReturnExact: the provider redirects to exactly returnTo, byte for byte: no userinfo, query (not even an empty
// "?"), fragment, other path, case or port.
func stgReturnExact(o stgObservation, returnTo string) bool {
	return o.redirected() && o.loc == returnTo
}

// stgReturnNotHonored: the unregistered return is refused (4xx) or replaced by a redirect that does not lead to its
// host. Anything else (2xx, 5xx, a redirect to that host) fails.
func stgReturnNotHonored(o stgObservation, host string) bool {
	if o.status >= 400 && o.status < 500 {
		return true
	}
	return o.redirected() && !o.toHost(host)
}

// authorize asks the provider to start a hosted sign-in whose redirect_uri is origin's callback, built by the
// production adapter with a fresh state and S256 challenge.
func (r *stgRun) authorize(ctx context.Context, origin string) stgObservation {
	r.t.Helper()
	sum := sha256.Sum256([]byte(stgRandom(32)))
	raw, err := r.a.AuthorizationURL(ctx, identity.AuthRequest{State: stgRandom(16),
		CodeChallenge: base64.RawURLEncoding.EncodeToString(sum[:]), RedirectURI: identity.CallbackURL(origin)})
	if err != nil {
		r.t.Fatalf("authorization URL: %v", err)
	}
	return r.observe(ctx, raw)
}

// domainCallback (AC7, and AC8's shared callback at the authorization step): the provider accepts a wildcard-covered
// third host and A as redirect_uri and refuses an unregistered one.
func (r *stgRun) domainCallback(ctx context.Context, in stgDomainInputs) {
	t := r.t
	wildHost := "g016-" + stgRandom(6) + "." + in.base
	negHost := "g016-neg-" + stgRandom(6) + ".invalid"
	neg := r.authorize(ctx, "https://"+negHost)
	wild := r.authorize(ctx, "https://"+wildHost)
	auth := r.authorize(ctx, in.authOrigin)
	nv, wv, av := stgAuthorizeVerdict(neg, in.authkit, in.authkitPath, negHost), stgAuthorizeVerdict(wild, in.authkit, in.authkitPath, wildHost),
		stgAuthorizeVerdict(auth, in.authkit, in.authkitPath, in.authHost)
	r.res["wildcard_callback"] = map[string]any{
		"wildcard_host": "per-run random label under the configured platform domain", "negative_host": "per-run .invalid",
		"wildcard_status": wild.status, "wildcard_verdict": wv, "auth_host_status": auth.status, "auth_host_verdict": av,
		"negative_status": neg.status, "negative_verdict": nv,
		// The first hop's shape (host and path only, never query), so that an unknown verdict caused by a path outside
		// the allow-list can be reconciled at EP-2 without recording anything else.
		"wildcard_hop_path_within_allowlist":  wild.redirected() && stgPathAllowed(wild.dest.Path, in.authkitPath),
		"auth_host_hop_path_within_allowlist": auth.redirected() && stgPathAllowed(auth.dest.Path, in.authkitPath),
	}
	if nv != stgRefused {
		t.Errorf("negative control: an unregistered redirect_uri must be refused; verdict %s (status %d)", nv, neg.status)
	}
	if wv != stgAccepted {
		t.Errorf("the wildcard-covered host must be accepted; verdict %s (status %d)", wv, wild.status)
	}
	if av != stgAccepted {
		t.Errorf("the auth host must be accepted; verdict %s (status %d)", av, auth.status)
	}
}

// domainSignOut (AC7 and AC8's sign-out leg): the production order, each case on its own fresh session: revoke the
// provider session through the API, then send the browser to the hosted logout with origin/sign-in as return_to. The
// provider returns to exactly that page for a wildcard-covered host and for A, and does not honor an unregistered one.
func (r *stgRun) domainSignOut(ctx context.Context, in stgDomainInputs, u *stgUser) {
	t := r.t
	wildOrigin := "https://g016-" + stgRandom(6) + "." + in.base
	negHost := "g016-neg-" + stgRandom(6) + ".invalid"
	logout := func(returnTo string) stgObservation {
		t.Helper()
		_, resp := r.signIn(u)
		vi, err := r.a.verify(ctx, resp)
		if err != nil || vi.ProviderSessionID == "" {
			t.Fatalf("fresh session not established (sid present %v): %v", vi.ProviderSessionID != "", err)
		}
		if err := r.a.RevokeProviderSession(ctx, vi.ProviderSessionID); err != nil {
			t.Fatalf("revoke before the hosted logout: %v", err)
		}
		raw, err := r.a.ProviderLogoutURL(vi.ProviderSessionID, returnTo)
		if err != nil {
			t.Fatalf("logout URL: %v", err)
		}
		return r.observe(ctx, raw)
	}
	wildReturn, authReturn := identity.SignInURL(wildOrigin), identity.SignInURL(in.authOrigin)
	neg := logout(identity.SignInURL("https://" + negHost))
	wild := logout(wildReturn)
	auth := logout(authReturn)
	r.res["sign_out_return"] = map[string]any{
		"order": "api revoke, then hosted logout (production)", "fresh_session_per_case": true,
		"wildcard_status": wild.status, "wildcard_returned_exactly": stgReturnExact(wild, wildReturn),
		"auth_host_status": auth.status, "auth_host_returned_exactly": stgReturnExact(auth, authReturn),
		"negative_status": neg.status, "negative_not_honored": stgReturnNotHonored(neg, negHost),
	}
	if !stgReturnNotHonored(neg, negHost) {
		t.Errorf("negative control: an unregistered return_to must not be honored (status %d)", neg.status)
	}
	if !stgReturnExact(wild, wildReturn) {
		t.Errorf("the wildcard-covered host's /sign-in must be the exact return destination (status %d)", wild.status)
	}
	if !stgReturnExact(auth, authReturn) {
		t.Errorf("the auth host's /sign-in must be the exact return destination (status %d)", auth.status)
	}
}

// TestStagingDomainObservationFixtures is offline (no provider, no inputs): the observation and verdict rules of
// TestStagingDomainQualification against pure table cases and loopback httptest responses.
func TestStagingDomainObservationFixtures(t *testing.T) {
	obs := func(status int, loc string) stgObservation {
		o := stgObservation{status: status, loc: loc}
		if loc != "" {
			o.dest, _ = url.Parse(loc)
		}
		return o
	}
	authkit, ok := stgOrigin("https://signin.authkit.example")
	if !ok {
		t.Fatal("fixture origin rejected")
	}
	const redirectHost = "g016-neg-x.invalid"
	for _, c := range []struct {
		name string
		o    stgObservation
		want stgVerdict
	}{
		{"hosted sign-in origin", obs(302, "https://signin.authkit.example/sign-in?client_id=c"), stgAccepted},
		{"hosted sign-in origin, other case", obs(303, "https://SIGNIN.authkit.example/x"), stgAccepted},
		{"arbitrary https error page", obs(302, "https://errors.example/oops"), stgUnknown},
		{"provider's own host", obs(302, "https://api.workos.com/error"), stgUnknown},
		{"hosted origin, other port", obs(302, "https://signin.authkit.example:8443/x"), stgUnknown},
		{"hosted origin over http", obs(302, "http://signin.authkit.example/x"), stgUnknown},
		{"hosted origin with userinfo", obs(302, "https://u@signin.authkit.example/x"), stgUnknown},
		{"hosted origin with an OAuth error", obs(302, "https://signin.authkit.example/x?error=invalid_request"), stgRefused},
		{"error description only", obs(302, "https://errors.example/?error_description=bad"), stgRefused},
		{"redirect to the redirect_uri host", obs(302, "https://g016-neg-x.invalid/auth/callback?error=x"), stgUnknown},
		{"redirect to the redirect_uri host with code", obs(302, "https://g016-neg-x.invalid/auth/callback?code=c"), stgUnknown},
		{"4xx", obs(400, ""), stgRefused},
		{"redirect without Location", obs(302, ""), stgUnknown},
		{"2xx page", obs(200, ""), stgUnknown},
		{"5xx", obs(503, ""), stgUnknown},
	} {
		// The origin-only cases keep the open prefix "/"; the path-shape cases below pin a closed allow-list.
		if got := stgAuthorizeVerdict(c.o, authkit, "/", redirectHost); got != c.want {
			t.Errorf("authorize verdict %q: got %s, want %s", c.name, got, c.want)
		}
	}
	// G057 §3.6: origin and path shape, not origin only (over the closed allow-list "/sign-in", dot-segment traversal included).
	for _, c := range []struct {
		name string
		o    stgObservation
		want stgVerdict
	}{
		{"hosted origin on the allow-listed path", obs(302, "https://signin.authkit.example/sign-in?client_id=c"), stgAccepted},
		{"hosted origin below the allow-listed path", obs(302, "https://signin.authkit.example/sign-in/mfa?x=1"), stgAccepted},
		{"hosted origin on another path", obs(302, "https://signin.authkit.example/oauth2/authorize?client_id=c"), stgUnknown},
		{"hosted origin on a path sharing the prefix bytes", obs(302, "https://signin.authkit.example/sign-inx"), stgUnknown},
		{"hosted origin below the path through a dot segment", obs(302, "https://signin.authkit.example/sign-in/../oauth2/authorize"), stgUnknown},
		{"hosted origin below the path through an encoded dot segment", obs(302, "https://signin.authkit.example/sign-in/%2e%2e/oauth2/authorize"), stgUnknown},
		{"hosted origin below the path through an upper-case encoded dot segment", obs(302, "https://signin.authkit.example/sign-in/%2E%2E/x"), stgUnknown},
		{"hosted origin below the path through a trailing dot segment", obs(302, "https://signin.authkit.example/sign-in/.."), stgUnknown},
		{"hosted origin below the path through a current-directory segment", obs(302, "https://signin.authkit.example/sign-in/./mfa"), stgUnknown},
		{"hosted origin below the path through a backslash", obs(302, "https://signin.authkit.example/sign-in/..%5c..%5cx"), stgUnknown},
		{"dot segment to a sibling", obs(302, "https://signin.authkit.example/sign-in/../x"), stgUnknown},
		{"encoded dot segment to a sibling", obs(302, "https://signin.authkit.example/sign-in/%2e%2e/x"), stgUnknown},
		{"encoded backslash after a dot pair", obs(302, "https://signin.authkit.example/sign-in/..%5cx"), stgUnknown},
		{"segment that merely contains a dot", obs(302, "https://signin.authkit.example/sign-in/a.b"), stgAccepted},
	} {
		if got := stgAuthorizeVerdict(c.o, authkit, "/sign-in", redirectHost); got != c.want {
			t.Errorf("path-shape verdict %q: got %s, want %s", c.name, got, c.want)
		}
	}
	for s, ok := range map[string]bool{"/": true, "/sign-in": true, "/a/b": true, "": false, "sign-in": false, "/sign-in/": false,
		"/sign-in?x": false, "/sign-in#f": false, "/a//b": false, "/a b": false} {
		if stgPathPrefix(s) != ok {
			t.Errorf("path prefix %q: got %v", s, !ok)
		}
	}
	if !stgPathAllowed("", "/") || stgPathAllowed("", "/sign-in") || !stgPathAllowed("/sign-in", "/sign-in") || stgPathAllowed("/Sign-in", "/sign-in") {
		t.Error("path allow-list edge cases")
	}
	for _, p := range []string{"/sign-in/../x", "/sign-in/./x", "/sign-in/..", "/sign-in/.", "/..", "/../x", "/sign-in/x/../../y", `/sign-in\..\x`} {
		if stgPathAllowed(p, "/sign-in") || stgPathAllowed(p, "/") {
			t.Errorf("path %q: dot segments and backslashes must never be allowed", p)
		}
	}
	for _, p := range []string{"/sign-in/..x", "/sign-in/x.y", "/sign-in/a.b", "/sign-in/...", "/sign-in/.well-known", "/sign-in/a..b"} {
		if !stgPathAllowed(p, "/sign-in") || !stgPathAllowed(p, "/") {
			t.Errorf("path %q: a segment that merely contains dots stays allowed", p)
		}
	}
	// G057 §3.6: the three cleanup verdicts of a per-run user after its deletion (present, absent, unknown).
	for name, c := range map[string]struct {
		err  error
		want string
	}{
		"still returned":       {nil, stgPresent},
		"provider 404":         {&sdk.APIError{StatusCode: http.StatusNotFound}, stgAbsent},
		"other provider error": {&sdk.APIError{StatusCode: http.StatusInternalServerError}, stgUnknownAbsence},
		"network error":        {errors.New("network or decode error"), stgUnknownAbsence},
	} {
		if got := stgAbsence(c.err); got != c.want {
			t.Errorf("cleanup verdict %q: got %s, want %s", name, got, c.want)
		}
	}

	const ret = "https://org-c.example.test/sign-in"
	for _, c := range []struct {
		name  string
		o     stgObservation
		exact bool
	}{
		{"exact", obs(302, ret), true},
		{"query", obs(302, ret+"?x=1"), false},
		{"empty query", obs(302, ret+"?"), false},
		{"fragment", obs(302, ret+"#f"), false},
		{"empty fragment", obs(302, ret+"#"), false},
		{"userinfo", obs(302, "https://u@org-c.example.test/sign-in"), false},
		{"port", obs(302, "https://org-c.example.test:443/sign-in"), false},
		{"case", obs(302, "https://ORG-C.example.test/sign-in"), false},
		{"trailing slash", obs(302, ret+"/"), false},
		{"not a redirect", obs(200, ret), false},
	} {
		if got := stgReturnExact(c.o, ret); got != c.exact {
			t.Errorf("exact return %q: got %v", c.name, got)
		}
	}
	for _, c := range []struct {
		name       string
		o          stgObservation
		notHonored bool
	}{
		{"4xx", obs(400, ""), true},
		{"redirect elsewhere", obs(302, "https://auth.example.test/sign-in"), true},
		{"honored", obs(302, "https://g016-neg-x.invalid/sign-in"), false},
		{"honored with userinfo", obs(302, "https://auth.example.test@g016-neg-x.invalid/sign-in"), false},
		{"2xx page", obs(200, ""), false},
		{"5xx", obs(502, ""), false},
	} {
		if got := stgReturnNotHonored(c.o, redirectHost); got != c.notHonored {
			t.Errorf("negative return %q: got %v", c.name, got)
		}
	}

	if _, ok := stgOrigin("https://signin.authkit.example/path"); ok {
		t.Error("an origin with a path was accepted")
	}

	const marker = "fixture-secret-state"
	mux := http.NewServeMux()
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "https://signin.authkit.example/sign-in?x=1")
		w.WriteHeader(http.StatusFound)
	})
	mux.HandleFunc("/two-locations", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Location", "https://a.example/")
		w.Header().Add("Location", "https://b.example/")
		w.WriteHeader(http.StatusFound)
	})
	mux.HandleFunc("/chunked-overflow", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusFound) // no Content-Length: chunked
		for range 8 {
			_, _ = w.Write([]byte(strings.Repeat("x", 32)))
			w.(http.Flusher).Flush()
		}
	})
	mux.HandleFunc("/declared-overflow", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "200")
		w.WriteHeader(http.StatusFound)
		_, _ = w.Write([]byte(strings.Repeat("x", 200)))
	})
	mux.HandleFunc("/stall", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Location", "https://signin.authkit.example/")
		w.WriteHeader(http.StatusFound)
		_, _ = w.Write([]byte("x"))
		w.(http.Flusher).Flush()
		select { // the headers are out; the body never ends before the client's deadline
		case <-req.Context().Done():
		case <-time.After(5 * time.Second):
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx := context.Background()
	hc := boundedClient(nil, 5*time.Second, 64)

	o, err := stgObserve(ctx, hc, srv.URL+"/redirect?state="+marker)
	if err != nil || o.status != http.StatusFound || o.dest == nil || o.dest.Host != "signin.authkit.example" ||
		stgAuthorizeVerdict(o, authkit, "/sign-in", redirectHost) != stgAccepted || stgAuthorizeVerdict(o, authkit, "/other", redirectHost) != stgUnknown {
		t.Errorf("redirect fixture: err %v, status %d", err, o.status)
	}
	for _, c := range []struct {
		path, want string
		hc         *http.Client
	}{
		{"/two-locations", "response: more than one Location", hc},
		{"/chunked-overflow", "body: response bound exceeded", hc},
		{"/declared-overflow", "request: response bound exceeded", hc},
		{"/stall", "body: timeout", boundedClient(nil, 300*time.Millisecond, 64)},
	} {
		o, err := stgObserve(ctx, c.hc, srv.URL+c.path+"?state="+marker)
		if err == nil || err.Error() != c.want {
			t.Errorf("%s: got error %v (status %d), want %q", c.path, err, o.status, c.want)
		}
		if err != nil && (strings.Contains(err.Error(), marker) || strings.Contains(err.Error(), srv.URL)) {
			t.Errorf("%s: the error text carries the request URL", c.path)
		}
	}
}
