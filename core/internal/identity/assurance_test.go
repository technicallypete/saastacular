package identity

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// A-5 (Plan r5 §8.2): with no reviewed proof source, a verified-MFA requirement denies every session, including a
// valid one, whatever labels it carries; the handler wrapper never reaches the protected handler.
func TestRequireVerifiedMFADeniesEverySession(t *testing.T) {
	for _, s := range []SessionInfo{
		{},
		{SessionID: "s", UserID: "u", MFAAssurance: MFAAssuranceUnknown, AuthTime: time.Now(), ExpiresAt: time.Now().Add(time.Hour)},
		{SessionID: "s", UserID: "u", MFAAssurance: "verified"},
		{SessionID: "s", UserID: "u", MFAAssurance: "mfa"},
	} {
		if err := RequireVerifiedMFA(s); err != ErrMFAUnverified {
			t.Errorf("RequireVerifiedMFA(%+v) = %v", s, err)
		}
	}
}

func TestRequireVerifiedMFAHandler(t *testing.T) {
	reached := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached = true })
	live := func(*http.Request) (SessionInfo, error) {
		return SessionInfo{SessionID: "s", UserID: "u", MFAAssurance: MFAAssuranceUnknown}, nil
	}
	anonymous := func(*http.Request) (SessionInfo, error) { return SessionInfo{}, ErrUnauthenticated }

	for _, c := range []struct {
		resolve func(*http.Request) (SessionInfo, error)
		status  int
		body    string
	}{
		{live, http.StatusForbidden, `{"error":"mfa_unverified"}`},
		{anonymous, http.StatusUnauthorized, `{"error":"unauthenticated"}`},
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/requires-mfa", nil)
		req.Header.Set("X-MFA-Verified", "true") // forged labels are never read
		req.Header.Set("Amr", "mfa")
		RequireVerifiedMFAHandler(c.resolve, next).ServeHTTP(rec, req)
		if rec.Code != c.status || rec.Body.String() != c.body+"\n" || rec.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("status %d body %q", rec.Code, rec.Body.String())
		}
	}
	if reached {
		t.Fatal("protected handler reached")
	}
}
