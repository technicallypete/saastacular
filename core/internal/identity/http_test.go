package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// U-HTTP (Plan r5 §8.2, §7.2): content types, body limit, rejection order, forged headers ignored, no CORS, no-store.

type fakeSvc struct {
	calls     int
	lastCred  string
	lastBind  string
	loginErr  error
	sessErr   error
	logoutErr error
	logoutOut *LogoutOutput
	hostErr   error // G016: SessionHost
	relayErr  error // G016: Relay
	lastState string
	// G016 security amendment: the origin of the last Session call, and a binding Organization override ("" = the
	// SessionHost Organization).
	lastOrigin string
	sessOrg    string
	// G017 D-6: the Membership answer, its error and the (user, Organization) it was asked for.
	notMember  bool
	memberErr  error
	memberArgs string
	// G058: the last login and callback inputs, and a scripted callback answer.
	lastLogin    LoginInput
	lastCallback CallbackInput
	cbOut        *CallbackOutput
	cbErr        error
}

func (f *fakeSvc) Membership(_ context.Context, user, org string) (bool, error) {
	f.calls++
	f.memberArgs = user + "@" + org
	if f.memberErr != nil {
		return false, f.memberErr
	}
	return !f.notMember, nil
}

func (f *fakeSvc) Login(_ context.Context, in LoginInput) (LoginOutput, error) {
	f.calls++
	f.lastCred, f.lastBind, f.lastLogin = in.Credential, in.Binding, in
	if f.loginErr != nil {
		return LoginOutput{}, f.loginErr
	}
	return LoginOutput{AuthorizationURL: "https://api.workos.com/user_management/authorize?x=1", Binding: "b", BindingMaxAge: 600}, nil
}

func (f *fakeSvc) Callback(_ context.Context, in CallbackInput) (CallbackOutput, error) {
	f.calls++
	f.lastCred, f.lastBind, f.lastCallback = in.Credential, in.Binding, in
	if f.cbErr != nil {
		return CallbackOutput{}, f.cbErr
	}
	if f.cbOut != nil {
		return *f.cbOut, nil
	}
	return CallbackOutput{Outcome: OutcomeSignedIn, Credential: "cred", MaxAge: 100, ReturnTo: "/account"}, nil
}

func (f *fakeSvc) Session(_ context.Context, cred, origin string) (SessionInfo, error) {
	f.calls++
	f.lastCred, f.lastOrigin = cred, origin
	if f.sessErr != nil {
		return SessionInfo{}, f.sessErr
	}
	org := "22222222-2222-2222-2222-222222222222"
	if f.sessOrg != "" {
		org = f.sessOrg
	}
	return SessionInfo{SessionID: "sid", UserID: "11111111-1111-1111-1111-111111111111", DisplayName: "Ada",
		MFAAssurance: MFAAssuranceUnknown, AuthTime: time.Unix(1, 0), ExpiresAt: time.Unix(2, 0), CSRFToken: make([]byte, 32),
		Origin: origin, OrganizationID: org}, nil
}

func (f *fakeSvc) Logout(_ context.Context, in LogoutInput) (LogoutOutput, error) {
	f.calls++
	f.lastCred = in.Credential
	if f.logoutErr != nil {
		return LogoutOutput{}, f.logoutErr
	}
	if f.logoutOut != nil {
		return *f.logoutOut, nil
	}
	return LogoutOutput{Outcome: OutcomeLocalRevoked}, nil
}

func (f *fakeSvc) ProviderLogoutReturn(context.Context, string) (string, error) {
	f.calls++
	return OutcomeLocalRevoked, nil
}

func (f *fakeSvc) SessionHost(_ context.Context, origin string) (HostResolution, error) {
	f.calls++
	if f.hostErr != nil {
		return HostResolution{}, f.hostErr
	}
	return HostResolution{Kind: HostPlatform, OrganizationID: "22222222-2222-2222-2222-222222222222", Routing: true}, nil
}

func (f *fakeSvc) Relay(_ context.Context, state string) (string, error) {
	f.calls++
	f.lastState = state
	if f.relayErr != nil {
		return "", f.relayErr
	}
	return "https://portal.example.com/auth/callback", nil
}

func mux(svc authService) *http.ServeMux {
	h := Handler{svc: svc}
	m := http.NewServeMux()
	m.HandleFunc(RouteLogin, h.Login)
	m.HandleFunc(RouteCallback, h.Callback)
	m.HandleFunc(RouteSession, h.Session)
	m.HandleFunc(RouteLogout, h.Logout)
	m.HandleFunc(RouteProviderReturn, h.ProviderReturn)
	m.HandleFunc(RouteRelay, h.Relay)
	return m
}

func do(t *testing.T, m http.Handler, method, path, ctype, body string, hdr map[string][]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	for k, vs := range hdr {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)
	return rec
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var b map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &b)
	return b["error"]
}

func TestMutationsAcceptJSONOnly(t *testing.T) {
	for _, path := range []string{"/internal/v1/auth/login", "/internal/v1/auth/callback", "/internal/v1/auth/logout", "/internal/v1/auth/provider-logout/return"} {
		for _, ct := range []string{"", "application/x-www-form-urlencoded", "text/plain", "multipart/form-data; boundary=x", "application/json; charset=latin1", "application/jsonx"} {
			svc := &fakeSvc{}
			rec := do(t, mux(svc), http.MethodPost, path, ct, `{"origin":"https://org-a.localhost:543"}`, nil)
			if rec.Code != http.StatusUnsupportedMediaType || errorCode(t, rec) != "unsupported_media_type" || svc.calls != 0 {
				t.Errorf("%s %q: %d %q calls=%d", path, ct, rec.Code, rec.Body.String(), svc.calls)
			}
		}
		svc := &fakeSvc{}
		rec := do(t, mux(svc), http.MethodPost, path, "application/json", `{"origin":"x"}`, map[string][]string{"Content-Type": {"application/json"}})
		if rec.Code != http.StatusUnsupportedMediaType || svc.calls != 0 {
			t.Errorf("%s with a repeated Content-Type: %d", path, rec.Code)
		}
	}
}

func TestBodyRules(t *testing.T) {
	big := `{"origin":"` + strings.Repeat("a", maxBody) + `"}`
	for name, body := range map[string]string{
		"over 16 KiB":   big,
		"unknown field": `{"origin":"https://org-a.localhost:543","user_id":"x"}`,
		"two objects":   `{"origin":"a"}{"origin":"b"}`,
		"not an object": `["origin"]`,
		"malformed":     `{"origin":`,
		"empty":         ``,
	} {
		svc := &fakeSvc{}
		rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/login", "application/json", body, nil)
		if rec.Code != http.StatusBadRequest || errorCode(t, rec) != "invalid_request" || svc.calls != 0 {
			t.Errorf("%s: %d %q", name, rec.Code, rec.Body.String())
		}
	}
	svc := &fakeSvc{}
	if rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/login", "application/json; charset=UTF-8", `{"origin":"https://org-a.localhost:543"}`, nil); rec.Code != http.StatusOK {
		t.Fatalf("valid JSON: %d %s", rec.Code, rec.Body.String())
	}
}

func TestCredentialAndBindingHeaders(t *testing.T) {
	svc := &fakeSvc{}
	m := mux(svc)
	rec := do(t, m, http.MethodPost, "/internal/v1/auth/login", "application/json", `{"origin":"o"}`,
		map[string][]string{"Authorization": {"Session cred1"}, "X-Login-Binding": {"bind1"},
			"X-User-Id": {"22222222-2222-2222-2222-222222222222"}, "X-Role": {"admin"}, "X-MFA-Verified": {"true"},
			"X-Forwarded-Host": {"evil.example"}})
	if rec.Code != http.StatusOK || svc.lastCred != "cred1" || svc.lastBind != "bind1" {
		t.Fatalf("headers: %d cred=%q bind=%q", rec.Code, svc.lastCred, svc.lastBind)
	}
	for name, hdr := range map[string]map[string][]string{
		"repeated Authorization":   {"Authorization": {"Session a", "Session b"}},
		"repeated X-Login-Binding": {"X-Login-Binding": {"a", "b"}},
	} {
		svc := &fakeSvc{}
		rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/login", "application/json", `{"origin":"o"}`, hdr)
		if rec.Code != http.StatusBadRequest || errorCode(t, rec) != "ambiguous_parameter" || svc.calls != 0 {
			t.Errorf("%s: %d", name, rec.Code)
		}
	}
	for _, auth := range []string{"Bearer cred", "session cred", "Session", "Session a b", "Basic dTpw"} {
		svc := &fakeSvc{}
		do(t, mux(svc), http.MethodPost, "/internal/v1/auth/login", "application/json", `{"origin":"o"}`, map[string][]string{"Authorization": {auth}})
		if svc.lastCred != "" {
			t.Errorf("Authorization %q yielded credential %q", auth, svc.lastCred)
		}
	}
}

func TestSessionRoute(t *testing.T) {
	svc := &fakeSvc{}
	rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/session", "", "", nil)
	if rec.Code != http.StatusUnauthorized || errorCode(t, rec) != "unauthenticated" || svc.calls != 0 {
		t.Fatalf("anonymous: %d", rec.Code)
	}
	// G016 security amendment: ?origin= is required (absent is invalid_request, repeated is ambiguous, both before any
	// service call); the routing tenant host is answered with the session, which must be bound to exactly that origin
	// and to the Organization it resolves to now (otherwise unauthenticated); a host that routes nowhere is unknown_host.
	auth := map[string][]string{"Authorization": {"Session c"}, "X-User-Id": {"x"}}
	if rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/session", "", "", auth); rec.Code != http.StatusBadRequest ||
		errorCode(t, rec) != "invalid_request" || svc.calls != 0 {
		t.Fatalf("absent origin: %d %s", rec.Code, errorCode(t, rec))
	}
	if rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/session?origin=a&origin=b", "", "", auth); rec.Code != http.StatusBadRequest ||
		errorCode(t, rec) != "ambiguous_parameter" || svc.calls != 0 {
		t.Fatalf("repeated origin: %d %s", rec.Code, errorCode(t, rec))
	}
	rec = do(t, mux(svc), http.MethodGet, "/internal/v1/auth/session?origin=https%3A%2F%2Forg-a.localhost%3A543", "", "", auth)
	var b map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("live: %d %v", rec.Code, err)
	}
	if b["mfa_assurance"] != "unknown" || b["user_id"] != "11111111-1111-1111-1111-111111111111" || b["csrf_token"] == "" || len(b) != 8 ||
		svc.lastOrigin != "https://org-a.localhost:543" {
		t.Fatalf("session body %v (origin %q)", b, svc.lastOrigin)
	}
	// G017 D-6 (BRD-13 at the HTTP level): the closed membership answer for the origin's Organization, asked for the
	// session's user only after the binding check; "none" when the binding is gone; a failed read is unavailable, never
	// a guessed state; a session bound to another Organization never reaches the membership read.
	if b["membership"] != "member" || svc.memberArgs != "11111111-1111-1111-1111-111111111111@22222222-2222-2222-2222-222222222222" {
		t.Fatalf("membership %v (asked %q)", b["membership"], svc.memberArgs)
	}
	svc.notMember = true
	rec = do(t, mux(svc), http.MethodGet, "/internal/v1/auth/session?origin=https%3A%2F%2Forg-a.localhost%3A543", "", "", auth)
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || rec.Code != http.StatusOK || b["membership"] != "none" || len(b) != 8 {
		t.Fatalf("revoked membership: %d %v", rec.Code, b)
	}
	svc.notMember, svc.memberErr = false, ErrUnavailable
	if rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/session?origin=https%3A%2F%2Forg-a.localhost%3A543", "", "", auth); rec.Code != http.StatusServiceUnavailable ||
		errorCode(t, rec) != "identity_unavailable" {
		t.Fatalf("membership outage: %d %s", rec.Code, errorCode(t, rec))
	}
	svc.memberErr, svc.memberArgs = nil, ""
	if h, _ := b["host"].(map[string]any); h["kind"] != "platform" || h["organization_id"] != "22222222-2222-2222-2222-222222222222" || len(h) != 2 {
		t.Fatalf("host %v", b["host"])
	}
	svc.sessOrg = "33333333-3333-3333-3333-333333333333" // bound to another Organization (a reassigned host)
	if rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/session?origin=https%3A%2F%2Forg-a.localhost%3A543", "", "", auth); rec.Code != http.StatusUnauthorized ||
		errorCode(t, rec) != "unauthenticated" || svc.memberArgs != "" {
		t.Fatalf("other Organization: %d %s (membership asked %q)", rec.Code, errorCode(t, rec), svc.memberArgs)
	}
	svc.sessOrg = ""
	calls := svc.calls
	svc.hostErr = ErrUnknownHost
	if rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/session?origin=x", "", "", auth); rec.Code != http.StatusNotFound ||
		errorCode(t, rec) != "unknown_host" || svc.calls != calls+1 {
		t.Fatalf("unknown host: %d %s", rec.Code, errorCode(t, rec))
	}
	svc.hostErr = nil
	svc.sessErr = ErrUnavailable
	if rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/session?origin=x", "", "", auth); rec.Code != http.StatusServiceUnavailable || errorCode(t, rec) != "identity_unavailable" {
		t.Fatalf("store outage: %d", rec.Code)
	}
}

// G016 B2 (Plan r4 §1.7 leg 2, §1.7.2): the relay route takes exactly one state, consumes nothing and answers the
// callback URL or the fixed relay_refused 404; no credential is read.
func TestRelayRoute(t *testing.T) {
	svc := &fakeSvc{}
	for _, path := range []string{"/internal/v1/auth/relay", "/internal/v1/auth/relay?state=a&state=b", "/internal/v1/auth/relay?x=1"} {
		rec := do(t, mux(svc), http.MethodGet, path, "", "", nil)
		if rec.Code != http.StatusNotFound || errorCode(t, rec) != "relay_refused" || svc.calls != 0 || rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s: %d %s", path, rec.Code, errorCode(t, rec))
		}
	}
	rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/relay?state=abc.def", "", "", nil)
	var b map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || rec.Code != http.StatusOK || len(b) != 1 ||
		b["callback"] != "https://portal.example.com/auth/callback" || svc.lastState != "abc.def" {
		t.Fatalf("relayed: %d %v", rec.Code, b)
	}
	svc.relayErr = ErrRelayRefused
	if rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/relay?state=abc.def", "", "", nil); rec.Code != http.StatusNotFound || errorCode(t, rec) != "relay_refused" {
		t.Fatalf("refused: %d", rec.Code)
	}
	svc.relayErr = ErrUnavailable
	if rec := do(t, mux(svc), http.MethodGet, "/internal/v1/auth/relay?state=abc.def", "", "", nil); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("unavailable: %d", rec.Code)
	}
	if rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/relay?state=abc.def", "application/json", "{}", nil); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST relay: %d", rec.Code)
	}
}

func TestResponsesAreNoStoreWithoutCORS(t *testing.T) {
	svc := &fakeSvc{}
	for _, c := range []struct{ method, path, ct, body string }{
		{http.MethodPost, "/internal/v1/auth/login", "application/json", `{"origin":"o"}`},
		{http.MethodPost, "/internal/v1/auth/callback", "application/json", `{"origin":"o","raw_query":"code=x&state=y"}`},
		{http.MethodGet, "/internal/v1/auth/session", "", ""},
		{http.MethodPost, "/internal/v1/auth/logout", "application/json", `{"scope":"current","csrf_token":"","origin":"o"}`},
		{http.MethodPost, "/internal/v1/auth/provider-logout/return", "application/json", `{"logout_ref":"x"}`},
		{http.MethodPost, "/internal/v1/auth/login", "text/plain", `x`},
		{http.MethodGet, "/internal/v1/auth/relay?state=s", "", ""},
	} {
		rec := do(t, mux(svc), c.method, c.path, c.ct, c.body, map[string][]string{"Origin": {"https://evil.example"}, "Authorization": {"Session c"}})
		h := rec.Header()
		if h.Get("Cache-Control") != "no-store" || h.Get("Content-Type") != "application/json" {
			t.Errorf("%s %s: headers %v", c.method, c.path, h)
		}
		for k := range h {
			if strings.HasPrefix(strings.ToLower(k), "access-control-") {
				t.Errorf("%s %s: CORS header %s", c.method, c.path, k)
			}
		}
	}
}

func TestMethodsAreExact(t *testing.T) {
	svc := &fakeSvc{}
	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/internal/v1/auth/login"}, {http.MethodGet, "/internal/v1/auth/logout"},
		{http.MethodPost, "/internal/v1/auth/session"}, {http.MethodPut, "/internal/v1/auth/callback"},
		{http.MethodDelete, "/internal/v1/auth/logout"},
	} {
		rec := do(t, mux(svc), c.method, c.path, "application/json", `{}`, nil)
		if rec.Code != http.StatusMethodNotAllowed || svc.calls != 0 {
			t.Errorf("%s %s: %d", c.method, c.path, rec.Code)
		}
	}
}

// TestLoginEnrollmentIntent (G058 ENR-H, modified login route; Plan r3 §1.3 step 1, §5.2): intent "enroll" with an
// optional canonical invitation id and request flag reaches Login as the enrollment facts; no intent is the ordinary
// login; an unknown intent or a malformed, non-canonical or zero invitation id is invalid_request before any service
// call; facts without the intent reach Login, which refuses them (TestEnrollmentCallbackInputs). The answer is unchanged.
func TestLoginEnrollmentIntent(t *testing.T) {
	const inv = "0f58e000-0000-4000-8000-00000000002a"
	svc := &fakeSvc{}
	rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/login", "application/json",
		`{"origin":"https://org-a.localhost:543","intent":"enroll","invitation_id":"`+inv+`","request":true}`, nil)
	var b map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || rec.Code != http.StatusOK || len(b) != 2 || b["authorization_url"] == nil {
		t.Fatalf("enrollment login: %d %v", rec.Code, b)
	}
	l := svc.lastLogin
	if !l.EnrollmentIntent || !l.Request || l.Invitation == nil || invitationString(*l.Invitation) != inv || l.ReturnTo != "" {
		t.Fatalf("enrollment facts: %+v", l)
	}
	svc = &fakeSvc{}
	if rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/login", "application/json", `{"origin":"o","return_to":"/account"}`, nil); rec.Code != http.StatusOK ||
		svc.lastLogin.EnrollmentIntent || svc.lastLogin.Invitation != nil || svc.lastLogin.Request {
		t.Fatalf("ordinary login: %d %+v", rec.Code, svc.lastLogin)
	}
	svc = &fakeSvc{}
	if rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/login", "application/json", `{"origin":"o","request":true}`, nil); rec.Code != http.StatusOK ||
		svc.lastLogin.EnrollmentIntent || !svc.lastLogin.Request {
		t.Fatalf("facts without intent reach Login: %d %+v", rec.Code, svc.lastLogin)
	}
	for name, body := range map[string]string{
		"unknown intent":       `{"origin":"o","intent":"join"}`,
		"uppercase id":         `{"origin":"o","intent":"enroll","invitation_id":"0F58E000-0000-4000-8000-00000000002A"}`,
		"braced id":            `{"origin":"o","intent":"enroll","invitation_id":"{` + inv + `}"}`,
		"compact id":           `{"origin":"o","intent":"enroll","invitation_id":"0f58e000000040008000000000002a00"}`,
		"zero id":              `{"origin":"o","intent":"enroll","invitation_id":"00000000-0000-0000-0000-000000000000"}`,
		"non-hex id":           `{"origin":"o","intent":"enroll","invitation_id":"0f58e000-0000-4000-8000-00000000002g"}`,
		"misplaced hyphens id": `{"origin":"o","intent":"enroll","invitation_id":"0f58e00-00000-4000-8000-00000000002a"}`,
		"request not a bool":   `{"origin":"o","intent":"enroll","request":"yes"}`,
	} {
		svc := &fakeSvc{}
		if rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/login", "application/json", body, nil); rec.Code != http.StatusBadRequest ||
			errorCode(t, rec) != "invalid_request" || svc.calls != 0 {
			t.Errorf("%s: %d %s (calls %d)", name, rec.Code, errorCode(t, rec), svc.calls)
		}
	}
}

// TestCallbackEnrollment (G058 ENR-H, modified callback route; Plan r3 §1.3 steps 2-4, §5.2; Codex request-c2
// disposition 2): the BFF-forwarded invitation intent reaches Callback (a malformed id is invalid_request before the
// service); a confirmed admission is 200 {outcome, next: sign_in} with no session and no credential (status admitted is
// historical; the ordinary sign-in decides); a pending or refused admission is 202 {enrollment: {status, credential:
// {value, max_age}}} with the credential only in its own field and never a session; the closed refusals keep their
// statuses (403 enrollment_closed, 409 invitation_mismatch and identity_conflict, 503 identity_unavailable for an
// unknown outcome); the ordinary answer is unchanged.
func TestCallbackEnrollment(t *testing.T) {
	const inv = "0f58e000-0000-4000-8000-00000000002a"
	tok := strings.Repeat("A", 43)
	cred := strings.Repeat("C", 43)
	svc := &fakeSvc{cbOut: &CallbackOutput{Outcome: OutcomeEnrollmentStatus, ReturnTo: "/join/status",
		Enrollment: &EnrollmentOutcome{Status: AdmitAdmitted, Next: NextSignIn}}}
	rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/callback", "application/json",
		`{"origin":"https://org-a.localhost:543","raw_query":"code=c&state=s","invitation_id":"`+inv+`","invitation_token":"`+tok+`"}`, nil)
	var b map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || rec.Code != http.StatusOK || len(b) != 2 || b["outcome"] != "admitted" ||
		b["next"] != "sign_in" {
		t.Fatalf("admitted: %d %v", rec.Code, b)
	}
	if c := svc.lastCallback; c.Invitation == nil || invitationString(*c.Invitation) != inv || c.InvitationToken != tok {
		t.Fatalf("forwarded intent: %+v", c)
	}
	for _, status := range []string{EnrollmentPending, AdmitEnrollmentRestricted, AdmitNotFound} {
		svc := &fakeSvc{cbOut: &CallbackOutput{Outcome: OutcomeEnrollmentStatus, ReturnTo: "/join/status", MaxAge: 1800,
			Enrollment: &EnrollmentOutcome{Status: status, Credential: cred}}}
		rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/callback", "application/json", `{"origin":"o","raw_query":"code=c&state=s"}`, nil)
		var b map[string]map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || rec.Code != http.StatusAccepted || len(b) != 1 {
			t.Fatalf("%s: %d %s", status, rec.Code, rec.Body.String())
		}
		e := b["enrollment"]
		c, _ := e["credential"].(map[string]any)
		if len(e) != 2 || e["status"] != status || len(c) != 2 || c["value"] != cred || c["max_age"] != float64(1800) {
			t.Errorf("%s: enrollment answer %v", status, e)
		}
		if strings.Count(rec.Body.String(), cred) != 1 || strings.Contains(rec.Body.String(), "session") {
			t.Errorf("%s: credential outside its field or a session in %s", status, rec.Body.String())
		}
	}
	for err, code := range map[error]int{ErrEnrollmentClosed: 403, ErrInvitationMismatch: 409, ErrIdentityConflict: 409, ErrUnavailable: 503} {
		svc := &fakeSvc{cbErr: err}
		if rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/callback", "application/json", `{"origin":"o","raw_query":"x"}`, nil); rec.Code != code ||
			errorCode(t, rec) != AsError(err).Code {
			t.Errorf("%v: %d %s", err, rec.Code, errorCode(t, rec))
		}
	}
	for name, body := range map[string]string{
		"malformed id": `{"origin":"o","raw_query":"x","invitation_id":"nope","invitation_token":"` + tok + `"}`,
		"zero id":      `{"origin":"o","raw_query":"x","invitation_id":"00000000-0000-0000-0000-000000000000","invitation_token":"` + tok + `"}`,
		"address":      `{"origin":"o","raw_query":"x","address":"a@example.invalid","email_verified":true}`,
	} {
		svc := &fakeSvc{}
		if rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/callback", "application/json", body, nil); rec.Code != http.StatusBadRequest ||
			errorCode(t, rec) != "invalid_request" || svc.calls != 0 {
			t.Errorf("%s: %d %s (calls %d)", name, rec.Code, errorCode(t, rec), svc.calls)
		}
	}
	svc = &fakeSvc{}
	rec = do(t, mux(svc), http.MethodPost, "/internal/v1/auth/callback", "application/json", `{"origin":"o","raw_query":"x"}`, nil)
	b = nil
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || rec.Code != http.StatusOK || b["outcome"] != "signed_in" || b["return_to"] != "/account" ||
		svc.lastCallback.Invitation != nil || svc.lastCallback.InvitationToken != "" {
		t.Fatalf("ordinary callback: %d %v", rec.Code, b)
	}
}

// C7 addendum §2: the provider sign-out answer carries the one-use reference only in its own logout_ref field, with
// the fixed local return; a local sign-out answer carries none.
func TestLogoutAnswerCarriesTheReferenceOnlyInItsField(t *testing.T) {
	ref := strings.Repeat("A", 43)
	svc := &fakeSvc{logoutOut: &LogoutOutput{Outcome: OutcomeProviderRedirect, ReturnTo: "/sign-in", LogoutRef: ref,
		ProviderLogoutURL: "https://api.workos.com/user_management/sessions/logout?session_id=s&return_to=https%3A%2F%2Forg-a.localhost%3A543%2Fsign-in"}}
	rec := do(t, mux(svc), http.MethodPost, "/internal/v1/auth/logout", "application/json",
		`{"scope":"provider","csrf_token":"t","origin":"https://org-a.localhost:543"}`, map[string][]string{"Authorization": {"Session " + ref}})
	var b map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if b["logout_ref"] != ref || b["return_to"] != "/sign-in" || b["outcome"] != OutcomeProviderRedirect ||
		strings.Contains(b["provider_logout_url"], ref) || strings.Contains(b["return_to"], "logout_ref") {
		t.Fatalf("answer %v", b)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("cache-control %q", rec.Header().Get("Cache-Control"))
	}
	rec = do(t, mux(&fakeSvc{}), http.MethodPost, "/internal/v1/auth/logout", "application/json",
		`{"scope":"current","csrf_token":"t","origin":"https://org-a.localhost:543"}`, map[string][]string{"Authorization": {"Session " + ref}})
	b = nil
	_ = json.Unmarshal(rec.Body.Bytes(), &b)
	if _, ok := b["logout_ref"]; ok {
		t.Fatalf("a local sign-out answer carries a reference: %v", b)
	}
}

// rootMux is mux with the G062 root origin and the NewHandler limiters (fresh per call).
func rootMux(svc authService, root string) *http.ServeMux {
	h := Handler{svc: svc, root: root, rootStart: newLimiter(rootStartRate, rootStartBurst),
		cbBinding: newLimiter(rootCallbackRate, rootCallbackBurst), cbGlobal: newLimiter(rootCallbackGlobal, rootCallbackGBurst)}
	m := http.NewServeMux()
	m.HandleFunc(RouteLogin, h.Login)
	m.HandleFunc(RouteCallback, h.Callback)
	return m
}

// TestRootCallbackRateLimited (G062 Plan r5 §4.5 (i)-(ii), SUP-9): the root signup login start shares one global bucket
// (burst 20); a root-origin callback is limited per login binding (burst 5) and globally (burst 50), so churned binding
// headers exhaust the global bucket; tenant callbacks and ordinary logins are not limited here. Exhaustion is 429
// rate_limited before the service is called.
func TestRootCallbackRateLimited(t *testing.T) {
	const root = "https://localhost:543"
	login := func(m http.Handler, intent string) int {
		body := `{"origin":"` + root + `"` + intent + `}`
		return do(t, m, http.MethodPost, "/internal/v1/auth/login", "application/json", body, nil).Code
	}
	t.Run("signup start global bucket", func(t *testing.T) {
		svc := &fakeSvc{}
		m := rootMux(svc, root)
		for i := range rootStartBurst {
			if c := login(m, `,"intent":"signup"`); c != http.StatusOK {
				t.Fatalf("start %d: %d", i, c)
			}
		}
		if !svc.lastLogin.SignupIntent || svc.lastLogin.EnrollmentIntent {
			t.Fatalf("login input %+v", svc.lastLogin)
		}
		calls := svc.calls
		if c := login(m, `,"intent":"signup"`); c != http.StatusTooManyRequests || svc.calls != calls {
			t.Fatalf("exhausted start: %d (service calls %d -> %d)", c, calls, svc.calls)
		}
		if c := login(m, ""); c != http.StatusOK || svc.lastLogin.SignupIntent {
			t.Fatalf("an ordinary login was limited or marked signup: %d", c)
		}
	})
	cb := func(m http.Handler, origin, binding string) int {
		return do(t, m, http.MethodPost, "/internal/v1/auth/callback", "application/json",
			`{"origin":"`+origin+`","raw_query":"code=c&state=s"}`, map[string][]string{"X-Login-Binding": {binding}}).Code
	}
	t.Run("per binding", func(t *testing.T) {
		svc := &fakeSvc{}
		m := rootMux(svc, root)
		for i := range rootCallbackBurst {
			if c := cb(m, root, "binding-a"); c != http.StatusOK {
				t.Fatalf("callback %d: %d", i, c)
			}
		}
		calls := svc.calls
		if c := cb(m, root, "binding-a"); c != http.StatusTooManyRequests || svc.calls != calls {
			t.Fatalf("exhausted binding: %d", c)
		}
		if c := cb(m, root, "binding-b"); c != http.StatusOK {
			t.Fatalf("another binding: %d", c)
		}
		for i := range 2 * rootCallbackBurst {
			if c := cb(m, "https://org-a.localhost:543", "binding-a"); c != http.StatusOK {
				t.Fatalf("tenant callback %d limited: %d", i, c)
			}
		}
	})
	t.Run("churned bindings hit the global bucket", func(t *testing.T) {
		m := rootMux(&fakeSvc{}, root)
		ok := 0
		for i := range rootCallbackGBurst + 10 {
			if cb(m, root, "churn-"+string(rune('a'+i%26))+string(rune('a'+i/26))) == http.StatusOK {
				ok++
			}
		}
		if ok != rootCallbackGBurst {
			t.Fatalf("%d churned callbacks admitted, want exactly the global burst %d", ok, rootCallbackGBurst)
		}
	})
	t.Run("signup callback answer", func(t *testing.T) {
		svc := &fakeSvc{cbOut: &CallbackOutput{Outcome: OutcomeSignup, ReturnTo: "/signup/details", MaxAge: 1800,
			Signup: &SignupOutcome{Credential: "signup-cred"}}}
		rec := do(t, rootMux(svc, root), http.MethodPost, "/internal/v1/auth/callback", "application/json",
			`{"origin":"`+root+`","raw_query":"code=c&state=s"}`, map[string][]string{"X-Login-Binding": {"b"}})
		body := rec.Body.String()
		if rec.Code != http.StatusAccepted ||
			!strings.Contains(body, `"signup":{"credential":{"value":"signup-cred","max_age":1800},"next":"/signup/details"}`) ||
			strings.Contains(body, "session") {
			t.Fatalf("signup answer %d: %s", rec.Code, body)
		}
	})
	t.Run("signup intent refuses enrollment facts", func(t *testing.T) {
		m := rootMux(&fakeSvc{}, root)
		for _, extra := range []string{`,"intent":"signup","request":true`,
			`,"intent":"signup","invitation_id":"11111111-1111-4111-8111-111111111111"`, `,"intent":"other"`} {
			if c := login(m, extra); c != http.StatusBadRequest {
				t.Errorf("%s: %d, want 400", extra, c)
			}
		}
	})
}
