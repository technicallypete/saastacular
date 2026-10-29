//go:build workos_manual

// Manual hosted qualification observer (G004 Plan r5 §8.6 and §12.2 CAP-4; docs/foundation/IDENTITY.md
// "identity:manual"). It runs on the operator's workstation against the local identity overlay pointed at WorkOS
// staging, and never drives a browser. For each step it prints the instruction, waits for the operator's attestation
// in MQ_INBOX and observes the database for the expected transition (at most five minutes per step). It writes only
// sanitized results to MQ_RESULT_FILE: booleans, closed outcomes, timestamps and hashes, never an identifier,
// credential, code, token, name or address.
//
// Environment:
//   - database: DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD, DB_RUNTIME_PASSWORD
//   - provider (CAP-4a only): IDENTITY_TEST_CLIENT_ID, IDENTITY_TEST_ISSUER, IDENTITY_TEST_API_KEY
//   - IDENTITY_TEST_MAX_AUTH_AGE: the freshness bound the stack runs with (MQ-5)
//   - MQ_INBOX: a 0700 directory; the operator's attestations arrive as files named after the step, and CAP-4a's code
//     as cap4a.code, which is read once and deleted
//   - MQ_STEPS: comma-separated steps to run, in order (CAP-4a, MQ-0 … MQ-7, and MQ-6B for the second origin)
//
// The freshness and single-sign-on verdicts are the pure rules of mq_eval_test.go (G004 C7 addendum §1, §3), which
// run untagged in every core check.
//   - MQ_RESULT_FILE, MQ_TREE_SHA256, MQ_CA_SHA256: the result path and the hashes recorded at MQ-0
//
// G057 (Plan r5 §3.7, §4.3; D-14/D-16): the real-host steps Q-A1, Q-A2, Q-2, Q-4, Q-5, Q-6 and Q-7 observe the
// candidate stack on the qualification host. They read as <db>_auth_reader through the mq_observer_test.go helpers
// (migrator login, READ ONLY transactions, the operational statement registry) with row-level security in force and
// no bypass; DB_RUNTIME_PASSWORD is not required for them, and the emulator-era steps above (which need the runtime
// role and a database without row-level security) cannot be mixed into the same run. Their environment adds:
//   - MQ_ORIGINS: P=https://B[:port],T=https://T (the platform root and the custom host; A/B are bound by discovery
//     and are refused here); PLATFORM_HOST_PORT when the stack publishes a non-default port
//   - APP_ENV, DATABASE_NAME, DB_HOST, DB_PORT, DB_MIGRATOR_PASSWORD as for every migrate.LoadConfig caller
//
// The observer is started after the Q-P1 privacy proof and before Q-A's first real flow, so its first baseline
// precedes every real session; starting it opens no browser and sends no request to any host.
package identity_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/identity"
	"core/internal/identity/workos"
	"core/internal/migrate"
)

const (
	mqOriginA   = "https://org-a.localhost:543"
	mqOriginB   = "https://org-b.localhost:543"
	mqStepLimit = 5 * time.Minute
	mqPoll      = 2 * time.Second
)

type mqStep struct {
	ID           string         `json:"id"`
	Outcome      string         `json:"outcome"` // pass, fail or stopped
	At           string         `json:"at"`
	Attestation  string         `json:"attestation,omitempty"`
	Observations map[string]any `json:"observations"`
	Note         string         `json:"note,omitempty"`
}

type mqResult struct {
	Tree     string   `json:"tree_sha256"`
	Config   string   `json:"config_sha256"`
	CA       string   `json:"ca_sha256"`
	Started  string   `json:"started_at"`
	Steps    []mqStep `json:"steps"`
	Finished string   `json:"finished_at"`
}

type mqSession struct {
	id, userID, sid            string
	authTime, created          time.Time
	revoked                    bool
	reason, providerRevocation string
	assurance                  string
}

type observer struct {
	t      *testing.T
	pool   *pgxpool.Pool // emulator-era steps only: the runtime role
	svc    *identity.Service
	inbox  string
	result mqResult
	// Internal references between steps; never written out.
	userA                 string
	firstAuth, lastAuth   time.Time
	sessionsBeforeDisable []string
	// G057: the auth-reader observation path (migrator login, SET LOCAL ROLE reader) and the roles bound by discovery.
	ar      *pgxpool.Pool
	reader  string
	origins mqOrigins
	tenants map[string]mqTenant
	userT   string // the user of Q-2's sign-in on T; never written out
}

func mqEnv(t *testing.T, name string) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		t.Fatalf("%s is required", name)
	}
	return v
}

func TestManualQualification(t *testing.T) {
	ctx := context.Background()
	cfg, err := migrate.LoadConfig(os.Getenv)
	if err != nil {
		t.Fatalf("database configuration: %v", err)
	}
	o := &observer{t: t, inbox: mqEnv(t, "MQ_INBOX"), tenants: map[string]mqTenant{}}
	legacy := map[string]func() mqStep{
		"CAP-4a": o.cap4a, "MQ-0": o.mq0, "MQ-1": o.mq1, "MQ-2": o.mq2, "MQ-3": o.mq3,
		"MQ-4": o.mq4, "MQ-5": o.mq5, "MQ-6": o.mq6, "MQ-6B": o.mq6B, "MQ-7": o.mq7,
	}
	g057 := map[string]func() mqStep{
		"Q-A1": o.qA1, "Q-A2": o.qA2, "Q-2": o.q2, "Q-4": o.q4, "Q-5": o.q5, "Q-6": o.q6, "Q-7": o.q7,
	}
	var steps []string
	wantLegacy, wantG057 := false, false
	for _, id := range strings.Split(mqEnv(t, "MQ_STEPS"), ",") {
		id = strings.TrimSpace(id)
		_, isLegacy := legacy[id]
		_, isG057 := g057[id]
		if !isLegacy && !isG057 {
			t.Fatalf("unknown step %s", id)
		}
		wantLegacy, wantG057 = wantLegacy || isLegacy, wantG057 || isG057
		steps = append(steps, id)
	}
	if wantLegacy && wantG057 {
		t.Fatal("MQ_STEPS mixes emulator-era steps (runtime role, row-level security off) with G057 steps (auth reader, row-level security on); run them separately")
	}
	if wantLegacy {
		pc, err := pgxpool.ParseConfig("")
		if err != nil {
			t.Fatal(err)
		}
		pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = cfg.Host, uint16(cfg.Port), cfg.Database
		pc.ConnConfig.User, pc.ConnConfig.Password = cfg.RuntimeRole(), mqEnv(t, "DB_RUNTIME_PASSWORD")
		pool, err := pgxpool.NewWithConfig(ctx, pc)
		if err != nil {
			t.Fatal(err)
		}
		defer pool.Close()
		// G008 RC-12: the emulator-era observer reads other users' users and sessions rows through raw runtime SQL.
		// From version 6 row-level security correctly hides them, so it would record empty evidence; it fails closed
		// instead. The G057 steps read as the auth reader (below) and need no such gate.
		var protected int
		if err := pool.QueryRow(ctx, `SELECT count(*)::int FROM pg_class WHERE oid IN ('app.users'::regclass, 'app.sessions'::regclass)
  AND relrowsecurity`).Scan(&protected); err != nil || protected != 0 {
			t.Fatalf("manual qualification observer: app.users/app.sessions have row-level security (%d, %v); the observer's raw "+
				"runtime reads would record empty evidence and must be reconciled first (G008 RC-12)", protected, err)
		}
		o.pool = pool
		o.svc = identity.NewService(identity.NewStore(pool), identity.Config{
			Origins: identity.NewOrigins([]string{mqOriginA, mqOriginB}), Idle: 30 * time.Minute, Max: 8 * time.Hour,
			Now: time.Now, Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		})
	}
	if wantG057 {
		// Preconditions first (fail closed), then the origins: nothing below opens a browser or contacts a host.
		o.reader = cfg.Database + "_auth_reader"
		ar, err := mqMigratorPool(ctx, cfg)
		if err != nil {
			t.Fatalf("observer login: %v", err)
		}
		defer ar.Close()
		o.ar = ar
		if err := mqPreflight(ctx, ar, o.reader); err != nil {
			t.Fatalf("observer preconditions: %v", err)
		}
		if o.origins, err = mqParseOrigins(mqEnv(t, "MQ_ORIGINS"), strings.TrimSpace(os.Getenv("PLATFORM_HOST_PORT"))); err != nil {
			t.Fatal(err)
		}
	}
	o.result = mqResult{
		Tree: os.Getenv("MQ_TREE_SHA256"), CA: os.Getenv("MQ_CA_SHA256"), Started: time.Now().UTC().Format(time.RFC3339),
		Config: mqConfigHash(),
	}
	defer o.write()
	for _, id := range steps {
		run, ok := legacy[id]
		if !ok {
			run = g057[id]
		}
		s := run()
		s.ID, s.At = id, time.Now().UTC().Format(time.RFC3339)
		o.result.Steps = append(o.result.Steps, s)
		fmt.Printf("== %s: %s\n", id, s.Outcome)
		if s.Outcome == "stopped" {
			o.write()
			t.Fatalf("%s: capability stop: %s", id, s.Note)
		}
		if s.Outcome != "pass" {
			t.Errorf("%s: %s", id, s.Note)
		}
	}
}

// mqConfigHash identifies the provider configuration under test without recording any of it.
func mqConfigHash() string {
	h := sha256.New()
	for _, k := range []string{"IDENTITY_TEST_CLIENT_ID", "IDENTITY_TEST_ISSUER", "IDENTITY_TEST_MAX_AUTH_AGE"} {
		fmt.Fprintf(h, "%s=%s\n", k, strings.TrimSpace(os.Getenv(k)))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (o *observer) write() {
	o.result.Finished = time.Now().UTC().Format(time.RFC3339)
	b, err := json.MarshalIndent(o.result, "", "  ")
	if err != nil {
		o.t.Error(err)
		return
	}
	if p := os.Getenv("MQ_RESULT_FILE"); p != "" {
		if err := os.WriteFile(p, append(b, '\n'), 0o600); err != nil {
			o.t.Error(err)
		}
	}
}

// say prints an instruction for the operator.
func (o *observer) say(step, text string) {
	fmt.Printf("\n[%s] %s\n    When done, attest by writing one word into %s\n", step, text, filepath.Join(o.inbox, step))
}

// attest waits for the operator's one-word attestation for a step (one of allowed).
func (o *observer) attest(step string, allowed ...string) (string, error) {
	p := filepath.Join(o.inbox, step)
	deadline := time.Now().Add(mqStepLimit)
	for time.Now().Before(deadline) {
		b, err := os.ReadFile(p)
		if err == nil {
			_ = os.Remove(p)
			a := strings.TrimSpace(string(b))
			for _, w := range allowed {
				if a == w {
					return a, nil
				}
			}
			return "", fmt.Errorf("attestation %q is not one of %v", a, allowed)
		}
		time.Sleep(mqPoll)
	}
	return "", errors.New("no attestation within five minutes")
}

func (o *observer) sessions(since time.Time) []mqSession {
	rows, err := o.pool.Query(context.Background(), `
SELECT id::text, user_id::text, coalesce(provider_session_id, ''), auth_time, created_at, revoked_at IS NOT NULL,
       coalesce(revoke_reason, ''), coalesce(provider_revocation, ''), mfa_assurance
FROM app.sessions WHERE created_at >= $1 ORDER BY created_at`, since)
	if err != nil {
		o.t.Fatalf("sessions: %v", err)
	}
	defer rows.Close()
	var out []mqSession
	for rows.Next() {
		var s mqSession
		if err := rows.Scan(&s.id, &s.userID, &s.sid, &s.authTime, &s.created, &s.revoked, &s.reason, &s.providerRevocation, &s.assurance); err != nil {
			o.t.Fatalf("sessions: %v", err)
		}
		out = append(out, s)
	}
	return out
}

func (o *observer) session(id string) mqSession {
	var s mqSession
	err := o.pool.QueryRow(context.Background(), `
SELECT id::text, user_id::text, coalesce(provider_session_id, ''), auth_time, created_at, revoked_at IS NOT NULL,
       coalesce(revoke_reason, ''), coalesce(provider_revocation, ''), mfa_assurance
FROM app.sessions WHERE id = $1`, id).Scan(&s.id, &s.userID, &s.sid, &s.authTime, &s.created, &s.revoked, &s.reason, &s.providerRevocation, &s.assurance)
	if err != nil {
		o.t.Fatalf("session: %v", err)
	}
	return s
}

// latestAuth is the latest signed authentication time over all of a user's sessions, live or revoked.
func (o *observer) latestAuth(userID string) time.Time {
	var t *time.Time
	if err := o.pool.QueryRow(context.Background(), `SELECT max(auth_time) FROM app.sessions WHERE user_id = $1`, userID).Scan(&t); err != nil {
		o.t.Fatalf("latest auth_time: %v", err)
	}
	if t == nil {
		return time.Time{}
	}
	return *t
}

// returned reports whether a session's provider sign-out reference was consumed (the return reached the application).
func (o *observer) returned(id string) bool {
	return o.count(`SELECT count(*)::int FROM app.sessions WHERE id = $1 AND provider_logout_returned_at IS NOT NULL`, id) == 1
}

func (o *observer) count(q string, args ...any) int {
	var n int
	if err := o.pool.QueryRow(context.Background(), q, args...).Scan(&n); err != nil {
		o.t.Fatalf("count: %v", err)
	}
	return n
}

// waitSession waits for the first session created at or after since that satisfies ok.
func (o *observer) waitSession(since time.Time, ok func(mqSession) bool) (mqSession, bool) {
	deadline := time.Now().Add(mqStepLimit)
	for time.Now().Before(deadline) {
		for _, s := range o.sessions(since) {
			if ok(s) {
				return s, true
			}
		}
		time.Sleep(mqPoll)
	}
	return mqSession{}, false
}

func fail(note string, obs map[string]any) mqStep {
	return mqStep{Outcome: "fail", Note: note, Observations: obs}
}
func pass(obs map[string]any) mqStep { return mqStep{Outcome: "pass", Observations: obs} }

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// CAP-4a: does WorkOS reject a real code exchanged with a different PKCE verifier while the client secret is sent?
func (o *observer) cap4a() mqStep {
	t := o.t
	a, err := workos.New(workos.Config{
		APIKey: mqEnv(t, "IDENTITY_TEST_API_KEY"), ClientID: mqEnv(t, "IDENTITY_TEST_CLIENT_ID"), Issuer: mqEnv(t, "IDENTITY_TEST_ISSUER"),
		BaseURL: "https://api.workos.com", Algorithms: []string{"RS256"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	verifier := randomToken()
	u, err := a.AuthorizationURL(ctx, identity.AuthRequest{State: randomToken(), CodeChallenge: challenge(verifier), RedirectURI: mqOriginA + "/auth/callback"})
	if err != nil {
		t.Fatal(err)
	}
	urlFile := filepath.Join(o.inbox, "cap4a-authorize-url.txt")
	if err := os.WriteFile(urlFile, []byte(u+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	since := time.Now()
	o.say("CAP-4a", "Open the URL in "+urlFile+" in the dedicated browser profile and sign in as user A (password and TOTP). "+
		"The browser returns to the application, which shows that the sign-in expired (it never started this sign-in). "+
		"Copy the code parameter of the /auth/callback URL from the browser's history or developer tools into "+
		filepath.Join(o.inbox, "cap4a.code")+" (do not paste it anywhere else), then attest 'pasted'.")
	att, err := o.attest("CAP-4a", "pasted")
	_ = os.Remove(urlFile)
	if err != nil {
		return fail(err.Error(), nil)
	}
	b, err := os.ReadFile(filepath.Join(o.inbox, "cap4a.code"))
	_ = os.Remove(filepath.Join(o.inbox, "cap4a.code"))
	code := strings.TrimSpace(string(b))
	if err != nil || code == "" || len(code) > 512 {
		return fail("no code in cap4a.code", nil)
	}
	obs := map[string]any{"application_session_created": len(o.sessions(since)) > 0}
	// The exchange with a different verifier must be refused.
	got, err := a.Exchange(ctx, identity.Exchange{Code: code, CodeVerifier: randomToken()})
	if err == nil {
		revokeErr := a.RevokeProviderSession(ctx, got.ProviderSessionID)
		obs["wrong_verifier"] = "tokens_returned"
		obs["obtained_session_revoked"] = revokeErr == nil
		return mqStep{Outcome: "stopped", Attestation: att, Observations: obs,
			Note: "CAPABILITY GAP: WorkOS returned tokens for a code exchanged with a different PKCE verifier"}
	}
	obs["wrong_verifier"] = string(identity.AsError(err).Code)
	if !errors.Is(err, identity.ErrCodeRejected) {
		return fail("the wrong-verifier exchange failed for another reason than a provider rejection", obs)
	}
	// Positive control: the same code with its own verifier shows the code itself was valid; any session obtained is
	// revoked at once. A refusal here means the provider spent the code on the rejected attempt (inconclusive control).
	got, err = a.Exchange(ctx, identity.Exchange{Code: code, CodeVerifier: verifier})
	switch {
	case err == nil:
		obs["positive_control"] = "code_valid_with_own_verifier"
		obs["control_session_revoked"] = a.RevokeProviderSession(ctx, got.ProviderSessionID) == nil
	case errors.Is(err, identity.ErrCodeRejected):
		obs["positive_control"] = "code_spent_by_rejected_attempt"
	default:
		obs["positive_control"] = string(identity.AsError(err).Code)
	}
	obs["application_session_created"] = len(o.sessions(since)) > 0
	if obs["application_session_created"] == true {
		return fail("the application created a session from a sign-in it never started", obs)
	}
	s := pass(obs)
	s.Attestation = att
	return s
}

// MQ-0: preflight. The binding row is absent or matches the configured namespace.
func (o *observer) mq0() mqStep {
	var issuer, client string
	err := o.pool.QueryRow(context.Background(), `SELECT issuer, client_id FROM app.identity_provider_binding WHERE provider = 'workos'`).Scan(&issuer, &client)
	obs := map[string]any{"tree_recorded": o.result.Tree != "", "ca_recorded": o.result.CA != ""}
	switch {
	case err != nil && strings.Contains(err.Error(), "no rows"):
		obs["binding"] = "absent"
	case err != nil:
		return fail("binding: "+err.Error(), obs)
	case issuer == os.Getenv("IDENTITY_TEST_ISSUER") && client == os.Getenv("IDENTITY_TEST_CLIENT_ID"):
		obs["binding"] = "matches_configuration"
	default:
		obs["binding"] = "mismatch"
		return fail("the binding row does not match the configured namespace", obs)
	}
	return pass(obs)
}

// MQ-1: hosted sign-in with password and TOTP; the session's assurance stays unknown and auth_time is recent.
func (o *observer) mq1() mqStep {
	since := time.Now()
	o.say("MQ-1", "In the dedicated browser profile open "+mqOriginA+"/sign-in and sign in as user A with password and the TOTP challenge; attest 'signed-in'.")
	att, err := o.attest("MQ-1", "signed-in")
	if err != nil {
		return fail(err.Error(), nil)
	}
	s, ok := o.waitSession(since, func(mqSession) bool { return true })
	if !ok {
		return fail("no session", nil)
	}
	o.userA, o.firstAuth, o.lastAuth = s.userID, s.authTime, s.authTime
	obs := map[string]any{
		"session_created": true, "mfa_assurance_unknown": s.assurance == "unknown",
		"auth_time_recent": time.Since(s.authTime) < 10*time.Minute, "provider_session_recorded": s.sid != "",
	}
	if !obs["mfa_assurance_unknown"].(bool) || !obs["auth_time_recent"].(bool) {
		return fail("assurance or auth_time", obs)
	}
	st := pass(obs)
	st.Attestation = att
	return st
}

// MQ-2: a sign-in started as user B and canceled at TOTP creates no session.
func (o *observer) mq2() mqStep {
	since := time.Now()
	users := o.count(`SELECT count(*)::int FROM app.users`)
	o.say("MQ-2", "In a private window of the dedicated profile, start a sign-in on "+mqOriginA+" as user B and cancel at the TOTP step (close or leave the page if there is no cancel control); attest 'returned-canceled' if the application showed 'Sign-in was canceled', else 'abandoned'.")
	att, err := o.attest("MQ-2", "returned-canceled", "abandoned")
	if err != nil {
		return fail(err.Error(), nil)
	}
	obs := map[string]any{"sessions_created": len(o.sessions(since)), "users_created": o.count(`SELECT count(*)::int FROM app.users`) - users}
	if obs["sessions_created"].(int) != 0 || obs["users_created"].(int) != 0 {
		return fail("a session or user was created", obs)
	}
	st := pass(obs)
	st.Attestation = att
	return st
}

// MQ-3: local sign-out, then a sign-in without a prompt. Single sign-on reuse is qualified only by mqSSOReuse: no
// prompt, the same user, a new session after the local revocation and the retained signed authentication time. A
// prompted sign-in is recorded as an ordinary sign-in and fails the step (C7 addendum §3).
func (o *observer) mq3() mqStep {
	since := time.Now()
	prior := o.latestAuth(o.userA)
	o.say("MQ-3", "On "+mqOriginA+"/account choose 'Sign out', then sign in again; attest 'no-prompt' if WorkOS did not ask for credentials, else 'prompt'.")
	att, err := o.attest("MQ-3", "no-prompt", "prompt")
	if err != nil {
		return fail(err.Error(), nil)
	}
	s, ok := o.waitSession(since, func(s mqSession) bool { return s.userID == o.userA })
	if !ok {
		return fail("no new session for user A", nil)
	}
	var revokedAt *time.Time
	if err := o.pool.QueryRow(context.Background(), `SELECT max(revoked_at) FROM app.sessions WHERE user_id = $1 AND revoke_reason = 'logout' AND revoked_at >= $2`, o.userA, since).Scan(&revokedAt); err != nil {
		o.t.Fatalf("local revocation: %v", err)
	}
	in := mqSSOInput{Attestation: att, SameUser: s.userID == o.userA, NewCreated: s.created, NewAuth: s.authTime, PriorLatestAuth: prior}
	if revokedAt != nil {
		in.RevokedAt = *revokedAt
	}
	qualified, why := mqSSOReuse(in)
	obs := map[string]any{
		"same_user": in.SameUser, "local_revocation_recorded": revokedAt != nil, "auth_time_retained": s.authTime.Equal(prior),
		"sso_reuse_qualified": qualified,
	}
	o.lastAuth = o.latestAuth(o.userA)
	st := pass(obs)
	if !qualified {
		st = fail("single sign-on reuse not shown: "+why, obs)
	}
	st.Attestation = att
	return st
}

// MQ-4: sign-in on the second origin maps to the same user.
func (o *observer) mq4() mqStep {
	since := time.Now()
	o.say("MQ-4", "Open "+mqOriginB+"/sign-in and sign in (as user A); attest 'signed-in'.")
	att, err := o.attest("MQ-4", "signed-in")
	if err != nil {
		return fail(err.Error(), nil)
	}
	s, ok := o.waitSession(since, func(mqSession) bool { return true })
	if !ok {
		return fail("no session", nil)
	}
	obs := map[string]any{"same_user": s.userID == o.userA}
	o.lastAuth = o.latestAuth(o.userA)
	if !obs["same_user"].(bool) {
		return fail("a different user", obs)
	}
	st := pass(obs)
	st.Attestation = att
	return st
}

// MQ-5: after the freshness bound, a sign-in prompts and auth_time advances (C7 addendum §1). The baseline is the
// latest signed authentication over all of user A's sessions, re-read before "ready": a sign-in during the wait moves
// it and extends the wait (counted, never ignored). Waiting is bounded: a baseline that keeps moving ends the step as
// incomplete, never as a freshness claim. The verdict is mqFreshness.
func (o *observer) mq5() mqStep {
	maxAge, err := time.ParseDuration(mqEnv(o.t, "IDENTITY_TEST_MAX_AUTH_AGE"))
	if err != nil {
		o.t.Fatalf("IDENTITY_TEST_MAX_AUTH_AGE: %v", err)
	}
	o.say("MQ-5", "On "+mqOriginA+"/account choose 'Sign out'. Then do not sign in anywhere until the observer prints 'ready'.")
	baseline := o.latestAuth(o.userA)
	moved := 0
	giveUp := time.Now().Add(maxAge + mqFreshMargin + mqStepLimit)
	for {
		ready := mqReadyAt(baseline, maxAge)
		if ready.After(giveUp) {
			return fail("incomplete: the authentication baseline kept moving during the wait", map[string]any{"baseline_moved": moved})
		}
		if wait := time.Until(ready); wait > 0 {
			time.Sleep(wait)
		}
		latest := o.latestAuth(o.userA)
		if !latest.After(baseline) {
			break
		}
		baseline, moved = latest, moved+1
		fmt.Println("    a sign-in during the wait moved the baseline; waiting longer")
	}
	readyAt := time.Now()
	fmt.Println("    ready: sign in on " + mqOriginA + " now; attest 'prompt' if WorkOS asked for credentials, else 'no-prompt'.")
	att, err := o.attest("MQ-5", "prompt", "no-prompt")
	if err != nil {
		return fail(err.Error(), map[string]any{"baseline": baseline.UTC().Format(time.RFC3339), "baseline_moved": moved})
	}
	s, ok := o.waitSession(readyAt, func(s mqSession) bool { return s.userID == o.userA })
	if !ok {
		return fail("no new session", nil)
	}
	observed := time.Now()
	fresh, why := mqFreshness(mqFreshInput{Attestation: att, SameUser: s.userID == o.userA, Baseline: baseline, NewAuth: s.authTime, ObservedAt: observed, MaxAge: maxAge})
	obs := map[string]any{
		"baseline": baseline.UTC().Format(time.RFC3339), "bound": (maxAge + mqFreshMargin).String(),
		"ready_at": readyAt.UTC().Format(time.RFC3339), "new_auth_time": s.authTime.UTC().Format(time.RFC3339),
		"baseline_moved": moved, "auth_time_advanced": s.authTime.After(baseline),
		"within_bound": observed.Sub(s.authTime) <= maxAge+mqFreshMargin,
	}
	o.lastAuth = o.latestAuth(o.userA)
	if !fresh {
		return fail("freshness not shown: "+why, obs)
	}
	st := pass(obs)
	st.Attestation = att
	return st
}

// MQ-6 (org-a) and MQ-6B (org-b): sign out and revoke the WorkOS session on that origin (C7 addendum §2, §4). The
// step first establishes the session being revoked with an actual sign-in on that origin, then requires: the local
// revocation and the confirmed provider revocation of exactly that session; the return reaching the application (its
// one-use reference consumed); the operator's sanitized observations of the destination (that origin's
// /auth/logout-return) and of the logout cookie being gone; and a next sign-in that asks for credentials with an
// advanced signed authentication time (CAP-4b). A provider return to another origin fails the step.
func (o *observer) mq6() mqStep  { return o.providerSignOut("MQ-6", mqOriginA) }
func (o *observer) mq6B() mqStep { return o.providerSignOut("MQ-6B", mqOriginB) }

func (o *observer) providerSignOut(step, origin string) mqStep {
	since := time.Now()
	o.say(step+"-pre", "Sign in on "+origin+"/sign-in (as user A) so that this browser holds a session on that origin; attest 'signed-in'.")
	if _, err := o.attest(step+"-pre", "signed-in"); err != nil {
		return fail(err.Error(), nil)
	}
	cur, ok := o.waitSession(since, func(s mqSession) bool { return s.userID == o.userA && !s.revoked })
	if !ok {
		return fail("no new session on "+origin+" to revoke", nil)
	}
	o.say(step, "On "+origin+"/account choose 'Sign out and revoke WorkOS session'; attest 'revoked-message' if the page then says the WorkOS session was revoked, 'failed-message' if it says it could not be revoked or confirmed, else 'other'.")
	att, err := o.attest(step, "revoked-message", "failed-message", "other")
	if err != nil {
		return fail(err.Error(), nil)
	}
	o.say(step+"-dest", "Look at the address bar of that page: attest 'same-origin-return' if it is exactly "+origin+"/auth/logout-return, else 'other'. Then, in DevTools > Storage > Cookies for "+origin+", attest in "+step+"-cookie 'cookie-absent' if no __Host-logout cookie is listed, else 'cookie-present'.")
	dest, err := o.attest(step+"-dest", "same-origin-return", "other")
	if err != nil {
		return fail(err.Error(), nil)
	}
	cookie, err := o.attest(step+"-cookie", "cookie-absent", "cookie-present")
	if err != nil {
		return fail(err.Error(), nil)
	}
	s := o.session(cur.id)
	obs := map[string]any{
		"local_revoked": s.revoked && s.reason == "provider_logout", "provider_revocation_confirmed": s.providerRevocation == "confirmed",
		"return_consumed": o.returned(cur.id), "same_origin_return": dest == "same-origin-return", "cookie_cleared": cookie == "cookie-absent",
	}
	before := o.latestAuth(o.userA)
	next := time.Now()
	o.say(step+"b", "Now sign in again on "+origin+"; attest 'prompt' if WorkOS asked for credentials, else 'no-prompt'.")
	att2, err := o.attest(step+"b", "prompt", "no-prompt")
	if err != nil {
		return fail(err.Error(), obs)
	}
	ns, ok := o.waitSession(next, func(s mqSession) bool { return s.userID == o.userA })
	obs["next_sign_in_prompted"] = att2 == "prompt"
	obs["auth_time_advanced"] = ok && ns.authTime.After(before)
	o.lastAuth = o.latestAuth(o.userA)
	st := mqStep{Outcome: "pass", Attestation: att + "," + dest + "," + cookie + "," + att2, Observations: obs}
	switch {
	case obs["local_revoked"] != true || obs["provider_revocation_confirmed"] != true:
		st.Outcome, st.Note = "fail", "the session was not revoked locally and confirmed at the provider"
	case att != "revoked-message" || obs["return_consumed"] != true || obs["same_origin_return"] != true || obs["cookie_cleared"] != true:
		st.Outcome, st.Note = "fail", "the provider return did not reach this origin's outcome page with the reference consumed and the cookie cleared"
	case att2 != "prompt" || obs["auth_time_advanced"] != true:
		st.Outcome, st.Note = "stopped", "CAPABILITY GAP: the provider sign-out sequence did not end the hosted session (§3.5)"
	}
	return st
}

// MQ-7: disable via the trusted call, sign in (refused), re-enable, sign in (new session); old sessions stay revoked.
func (o *observer) mq7() mqStep {
	ctx := context.Background()
	if err := o.svc.DisableUser(ctx, o.userA); err != nil {
		o.t.Fatalf("disable: %v", err)
	}
	disabledAt := time.Now()
	o.say("MQ-7", "User A is now disabled. Sign in on "+mqOriginA+"; attest 'refused' if the application did not sign you in.")
	att, err := o.attest("MQ-7", "refused")
	if err != nil {
		return fail(err.Error(), nil)
	}
	obs := map[string]any{
		"no_session_while_disabled": len(o.sessions(disabledAt)) == 0,
		"all_sessions_revoked":      o.count(`SELECT count(*)::int FROM app.sessions WHERE user_id = $1 AND revoked_at IS NULL AND created_at < $2`, o.userA, disabledAt) == 0,
	}
	if err := o.svc.EnableUser(ctx, o.userA); err != nil {
		o.t.Fatalf("enable: %v", err)
	}
	enabledAt := time.Now()
	o.say("MQ-7b", "User A is enabled again. Sign in on "+mqOriginA+"; attest 'signed-in'.")
	att2, err := o.attest("MQ-7b", "signed-in")
	if err != nil {
		return fail(err.Error(), obs)
	}
	_, ok := o.waitSession(enabledAt, func(s mqSession) bool { return s.userID == o.userA })
	obs["new_session_after_enable"] = ok
	obs["old_sessions_still_revoked"] = o.count(`SELECT count(*)::int FROM app.sessions WHERE user_id = $1 AND revoked_at IS NULL AND created_at < $2`, o.userA, enabledAt) == 0
	st := mqStep{Outcome: "pass", Attestation: att + "," + att2, Observations: obs}
	for _, k := range []string{"no_session_while_disabled", "all_sessions_revoked", "new_session_after_enable", "old_sessions_still_revoked"} {
		if obs[k] != true {
			st.Outcome, st.Note = "fail", k
		}
	}
	return st
}

// G057 steps begin (Plan r5 §3.7/§4.3). Every database read below goes through the mq_observer_test.go helpers:
// mqRead as <db>_auth_reader over the mqOperationalSQL registry (baseline set difference on granted columns, the
// allowed-column revocation probe, the Organization discovery). None of them references the emulator-era runtime
// bodies above (pinned by TestMQStepsReadThroughTheRegistry). Steps name origins by role; the result file records
// roles, booleans, counts and timestamps only: hosts are printed to the console for the operator and recorded
// sanitized in the Result, never in the result file.

// baselines takes a step's session baseline S and Organization-ID baseline O before the instruction is printed.
func (o *observer) baselines() (map[string]mqSessionState, map[string]bool) {
	ctx := context.Background()
	s, err := mqBaseline(ctx, o.ar, o.reader)
	if err != nil {
		o.t.Fatalf("session baseline: %v", err)
	}
	orgs, err := mqOrganizations(ctx, o.ar, o.reader)
	if err != nil {
		o.t.Fatalf("organization baseline: %v", err)
	}
	return s, mqOrgIDs(orgs)
}

func (o *observer) organizations() []mqOrg {
	orgs, err := mqOrganizations(context.Background(), o.ar, o.reader)
	if err != nil {
		o.t.Fatalf("organizations: %v", err)
	}
	return orgs
}

// waitNew polls for the first new session since baseline that satisfies pred (at most mqStepLimit).
func (o *observer) waitNew(baseline map[string]mqSessionState, pred func(mqSessionState) bool) (mqSessionState, mqDelta, bool) {
	s, d, ok, err := mqWaitNewSession(context.Background(), o.ar, o.reader, baseline, pred, mqStepLimit, mqPoll)
	if err != nil {
		o.t.Fatalf("new sessions: %v", err)
	}
	return s, d, ok
}

// waitRevoked polls until the session shows revoked_at IS NOT NULL (at most mqStepLimit); it reports the provider
// return flag observed with the revocation.
func (o *observer) waitRevoked(id string) (returned, ok bool) {
	returned, ok, err := mqWaitRevoked(context.Background(), o.ar, o.reader, id, mqStepLimit, mqPoll)
	if err != nil {
		o.t.Fatalf("revocation: %v", err)
	}
	return returned, ok
}

func (o *observer) latestAuthAR(userID string) time.Time {
	t, err := mqLatestAuth(context.Background(), o.ar, o.reader, userID)
	if err != nil {
		o.t.Fatalf("latest auth_time: %v", err)
	}
	return t
}

func (o *observer) bound() []mqTenant {
	var out []mqTenant
	for _, role := range []string{"A", "B", "C"} {
		if t, ok := o.tenants[role]; ok {
			out = append(out, t)
		}
	}
	return out
}

func (o *observer) tenant(role string) mqTenant {
	t, ok := o.tenants[role]
	if !ok {
		o.t.Fatalf("role %s is not bound (run the step that binds it first)", role)
	}
	return t
}

func (o *observer) onT(s mqSessionState) bool {
	return s.Kind == "browser" && s.Origin == o.origins.T.Origin
}

func mqHost(origin string) string { return strings.TrimPrefix(origin, "https://") }

// Q-A1/Q-A2 (and the Q-4 binding of C): a live /signup on the platform root creates an Organization whose tenant is
// reached by its generated label host; the continuation's completion is one new browser session. The baselines are
// taken before the instruction, the role is bound from the discovered rows, and the operator attests that the
// address bar shows exactly the bound host and that the typed slug's host answers 404.
func (o *observer) qA1() mqStep { return o.signupBind("Q-A1", "A", "the first") }
func (o *observer) qA2() mqStep { return o.signupBind("Q-A2", "B", "a second") }

func (o *observer) signupBind(step, role, ordinal string) mqStep {
	s0, o0 := o.baselines()
	o.say(step, "In the dedicated browser profile open "+o.origins.P.Origin+"/signup and create "+ordinal+" Organization (a fresh name and slug); "+
		"complete the continuation until the tenant page shows status=tenant_created; attest 'tenant-created'.")
	att, err := o.attest(step, "tenant-created")
	if err != nil {
		return fail(err.Error(), nil)
	}
	_, d, ok := o.waitNew(s0, func(s mqSessionState) bool { return s.Kind == "browser" })
	obs := map[string]any{"role": role, "new_sessions": len(d.New), "renewed_only": d.Renewed}
	if !ok {
		return fail("no new browser session after the continuation", obs)
	}
	tenant, err := mqBindTenant(role, d.New, o.organizations(), o0, o.origins.P, o.bound())
	obs["bound"] = err == nil
	if err != nil {
		return fail("tenant binding: "+err.Error(), obs)
	}
	o.tenants[role] = tenant
	obs["organization_new_since_baseline"] = true
	obs["label_host_under_platform_root"] = true
	o.say(step+"-address", "Look at the address bar of the page showing status=tenant_created: attest 'address-matches' if its host is exactly "+
		mqHost(tenant.Origin)+", else 'address-differs'. Then open https://<the slug you typed>."+o.origins.P.hostPort()+"/ in a new tab and attest in "+
		step+"-slug 'slug-host-404' if it answers 404 with no redirect, else 'other'.")
	addr, err := o.attest(step+"-address", "address-matches", "address-differs")
	if err != nil {
		return fail(err.Error(), obs)
	}
	slug, err := o.attest(step+"-slug", "slug-host-404", "other")
	if err != nil {
		return fail(err.Error(), obs)
	}
	obs["address_matches"] = addr == "address-matches"
	obs["slug_host_404"] = slug == "slug-host-404"
	st := pass(obs)
	switch {
	case addr != "address-matches":
		st = fail("the address bar host differs from the bound label host", obs)
	case slug != "slug-host-404":
		st = fail("the typed slug's host did not answer the fixed 404", obs)
	}
	st.Attestation = att + "," + addr + "," + slug
	return st
}

// Q-2: the first sign-in on T (claimed by A's owner at X-6, after Q-P2): a new browser session bound to A's
// Organization and to the exact origin T; __Host-sid on T and nothing on the shared auth host.
func (o *observer) q2() mqStep {
	a := o.tenant("A")
	s0, _ := o.baselines()
	o.say("Q-2", "Open "+o.origins.T.Origin+"/sign-in and sign in with password and TOTP; attest 'signed-in' when /account shows status=signed_in.")
	att, err := o.attest("Q-2", "signed-in")
	if err != nil {
		return fail(err.Error(), nil)
	}
	s, d, ok := o.waitNew(s0, o.onT)
	obs := map[string]any{"origin_role": "T", "new_sessions": len(d.New), "renewed_only": d.Renewed}
	if !ok {
		return fail("no new browser session on T", obs)
	}
	o.userT = s.UserID
	obs["bound_to_organization_a"] = s.OrganizationID == a.OrganizationID
	obs["mfa_assurance_unknown"] = s.Assurance == "unknown"
	obs["auth_time_recent"] = time.Since(s.AuthTime) < 10*time.Minute
	o.say("Q-2-cookie", "In DevTools > Storage > Cookies: attest 'sid-on-t-only' if __Host-sid is listed for "+mqHost(o.origins.T.Origin)+
		" and no cookie at all is listed for the shared auth host, else 'other'.")
	cookie, err := o.attest("Q-2-cookie", "sid-on-t-only", "other")
	if err != nil {
		return fail(err.Error(), obs)
	}
	obs["cookie_on_t_only"] = cookie == "sid-on-t-only"
	st := pass(obs)
	for _, k := range []string{"bound_to_organization_a", "mfa_assurance_unknown", "auth_time_recent", "cookie_on_t_only"} {
		if obs[k] != true {
			st = fail(k, obs)
		}
	}
	st.Attestation = att + "," + cookie
	return st
}

// Q-4: a third live /signup creates C after the provider record of X-4 with no further provider or DNS change; the
// same person completes the code-bearing callback into C's label host (same global user as Q-2 when Q-2 ran).
func (o *observer) q4() mqStep {
	st := o.signupBind("Q-4", "C", "a third")
	if st.Outcome != "pass" {
		return st
	}
	c := o.tenant("C")
	switch {
	case o.userT == "":
		st.Observations["same_user_as_q2"] = "not_observed"
	case c.UserID == o.userT:
		st.Observations["same_user_as_q2"] = true
	default:
		st.Observations["same_user_as_q2"] = false
		st = fail("C's sign-in maps to another global user than Q-2's", st.Observations)
	}
	return st
}

// freshOnT establishes the session a sign-out step revokes: a fresh sign-in on T attested by the operator and found
// as a new unrevoked browser session on that origin.
func (o *observer) freshOnT(step, run string) (mqSessionState, mqStep, bool) {
	s0, _ := o.baselines()
	o.say(step+"-pre", run+": sign in on "+o.origins.T.Origin+"/sign-in (password and TOTP) so that this browser holds a fresh session on the custom host; attest 'signed-in'.")
	if _, err := o.attest(step+"-pre", "signed-in"); err != nil {
		return mqSessionState{}, fail(err.Error(), nil), false
	}
	cur, d, ok := o.waitNew(s0, func(s mqSessionState) bool { return o.onT(s) && !s.Revoked })
	if !ok {
		return mqSessionState{}, fail("no new session on T to sign out", map[string]any{"new_sessions": len(d.New), "renewed_only": d.Renewed}), false
	}
	return cur, mqStep{}, true
}

// Q-5 (run 1): local sign-out only. The specific prior session shows revoked_at set and no provider return; no
// provider revocation is claimed.
func (o *observer) q5() mqStep {
	cur, st, ok := o.freshOnT("Q-5", "Run 1")
	if !ok {
		return st
	}
	o.say("Q-5", "On "+o.origins.T.Origin+"/account choose 'Sign out' (local sign-out only); attest 'signed-out' when the page shows status=signed_out. "+
		"Then, in DevTools > Storage > Cookies for "+mqHost(o.origins.T.Origin)+", attest in Q-5-cookie 'sid-absent' if no __Host-sid cookie is listed, else 'sid-present'.")
	att, err := o.attest("Q-5", "signed-out")
	if err != nil {
		return fail(err.Error(), nil)
	}
	cookie, err := o.attest("Q-5-cookie", "sid-absent", "sid-present")
	if err != nil {
		return fail(err.Error(), nil)
	}
	returned, revoked := o.waitRevoked(cur.ID)
	obs := map[string]any{
		"local_revoked":            revoked,
		"provider_logout_returned": returned,
		"provider_logout_claimed":  false,
		"cookie_cleared":           cookie == "sid-absent",
		"same_user_as_q2":          o.userT == "" || cur.UserID == o.userT,
	}
	st = pass(obs)
	switch {
	case !revoked:
		st = fail("the session was not revoked locally", obs)
	case returned:
		st = fail("a provider return was recorded for a local-only sign-out", obs)
	case cookie != "sid-absent":
		st = fail("the session cookie was still listed", obs)
	}
	st.Attestation = att + "," + cookie
	return st
}

// Q-6 (run 2): sign out and revoke the WorkOS session. Provider revocation is observed as the §3.7 triple: (a) the
// prior session's revoked_at and provider_logout_returned_at both set under the auth reader, (b) core's structured
// auth.provider.revocation/provider_session_revoked event for the window (read from the stack's logs into the Result,
// outside this observer), (c) the operator's 'revoked-message' attestation. The hosted logout returns to the shared
// auth host's /sign-in; the next sign-in on T prompts for credentials with an advanced auth_time, and its session is
// new after the revocation by the re-baseline (no timestamp comparison).
func (o *observer) q6() mqStep {
	cur, st, ok := o.freshOnT("Q-6", "Run 2")
	if !ok {
		return st
	}
	before := o.latestAuthAR(cur.UserID)
	o.say("Q-6", "On "+o.origins.T.Origin+"/account choose 'Sign out and revoke WorkOS session'; attest 'revoked-message' if the page then says the WorkOS session was revoked, "+
		"'failed-message' if it says it could not be revoked or confirmed, else 'other'.")
	att, err := o.attest("Q-6", "revoked-message", "failed-message", "other")
	if err != nil {
		return fail(err.Error(), nil)
	}
	o.say("Q-6-dest", "Look at the address bar after the hosted logout: attest 'auth-sign-in-return' if it is exactly the shared auth host's /sign-in page, else 'other'.")
	dest, err := o.attest("Q-6-dest", "auth-sign-in-return", "other")
	if err != nil {
		return fail(err.Error(), nil)
	}
	returned, revoked := o.waitRevoked(cur.ID)
	obs := map[string]any{
		"local_revoked":                 revoked,
		"provider_logout_returned":      returned,
		"revoked_message":               att == "revoked-message",
		"auth_sign_in_return":           dest == "auth-sign-in-return",
		"core_event_recorded_in_result": "auth.provider.revocation/provider_session_revoked",
	}
	// Ordering without timestamps: the re-baseline is taken once the prior session shows revoked, so any id absent
	// from it was inserted after the revocation.
	s1, _ := o.baselines()
	o.say("Q-6b", "Within ten minutes open "+o.origins.T.Origin+"/account (the recorded sign-out outcome shows), then sign in again on "+
		o.origins.T.Origin+"/sign-in; attest 'prompt' if WorkOS asked for credentials, else 'no-prompt'.")
	att2, err := o.attest("Q-6b", "prompt", "no-prompt")
	if err != nil {
		return fail(err.Error(), obs)
	}
	ns, _, ok := o.waitNew(s1, func(s mqSessionState) bool { return o.onT(s) && s.UserID == cur.UserID })
	obs["new_session_after_revocation"] = ok
	obs["next_sign_in_prompted"] = att2 == "prompt"
	obs["auth_time_advanced"] = ok && ns.AuthTime.After(before)
	reuse, _ := mqSSOReuseObserved(mqSSOObservation{Attestation: att2, SameUser: ok && ns.UserID == cur.UserID, NewAfterRevocation: ok,
		NewAuth: ns.AuthTime, PriorLatestAuth: before})
	obs["sso_reuse_qualified"] = reuse
	st = mqStep{Outcome: "pass", Attestation: att + "," + dest + "," + att2, Observations: obs}
	switch {
	case !revoked || !returned:
		st.Outcome, st.Note = "fail", "the session was not revoked locally with the provider return consumed"
	case att != "revoked-message" || dest != "auth-sign-in-return":
		st.Outcome, st.Note = "fail", "the provider sign-out did not show the revoked outcome and return to the shared auth host's sign-in page"
	case !ok:
		st.Outcome, st.Note = "fail", "no new session on T after the revocation"
	case att2 != "prompt" || obs["auth_time_advanced"] != true:
		st.Outcome, st.Note = "stopped", "CAPABILITY GAP: the provider sign-out sequence did not end the hosted session (§3.5)"
	}
	return st
}

// Q-7: signed in on T, the owner removes the domain; T answers the fixed 404 and refuses a login start, A's label
// host is unaffected (A's bound session stays live) and no session is created on T afterwards.
func (o *observer) q7() mqStep {
	a := o.tenant("A")
	s0, _ := o.baselines()
	o.say("Q-7", "Signed in on "+o.origins.T.Origin+", remove the custom domain with the operation the agent supplied; then reload "+o.origins.T.Origin+
		"/ and open "+o.origins.T.Origin+"/api/account and "+o.origins.T.Origin+"/sign-in: attest 'removed-404' if the first two answer 404 and the sign-in start is refused, else 'other'.")
	att, err := o.attest("Q-7", "removed-404", "other")
	if err != nil {
		return fail(err.Error(), nil)
	}
	o.say("Q-7-a", "Open "+a.Origin+"/account in the same browser: attest 'a-unaffected' if Organization A's tenant still shows the signed-in account, else 'other'.")
	att2, err := o.attest("Q-7-a", "a-unaffected", "other")
	if err != nil {
		return fail(err.Error(), nil)
	}
	aRevoked, _, err := mqRevocation(context.Background(), o.ar, o.reader, a.SessionID)
	if err != nil {
		o.t.Fatalf("A's session: %v", err)
	}
	d, err := mqNewSessions(context.Background(), o.ar, o.reader, s0)
	if err != nil {
		o.t.Fatalf("new sessions: %v", err)
	}
	onT := 0
	for _, s := range d.New {
		if s.Origin == o.origins.T.Origin {
			onT++
		}
	}
	obs := map[string]any{
		"removed_404":                     att == "removed-404",
		"a_unaffected":                    att2 == "a-unaffected",
		"a_session_still_live":            !aRevoked,
		"new_sessions_on_t_after_removal": onT,
	}
	st := pass(obs)
	switch {
	case att != "removed-404":
		st = fail("the removed host did not answer the fixed refusals", obs)
	case att2 != "a-unaffected" || aRevoked:
		st = fail("Organization A's label host was affected by the removal", obs)
	case onT != 0:
		st = fail("a session was created on the removed host", obs)
	}
	st.Attestation = att + "," + att2
	return st
}

// G057 steps end
