package identity

import (
	"net/http"
)

// RequireVerifiedMFA is the owned guard for any request or test profile that requires verified MFA (Spec R7,
// ADR-003 clarification M2). No reviewed source of session-specific MFA proof exists in G004, so it denies every
// session with mfa_unverified, whatever the session, provider configuration or claims. A future proof source needs
// a reviewed migration (sessions.mfa_assurance accepts only 'unknown') and Plan before this can change.
func RequireVerifiedMFA(SessionInfo) error { return ErrMFAUnverified }

// RequireVerifiedMFAHandler wraps a handler that requires verified MFA. It resolves the caller's session with
// resolve; an unauthenticated caller gets 401, every authenticated caller gets 403 mfa_unverified, and next is never
// reached. No G004 route uses it; its contract is tested directly.
func RequireVerifiedMFAHandler(resolve func(*http.Request) (SessionInfo, error), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info, err := resolve(r)
		if err != nil {
			writeError(w, AsError(err))
			return
		}
		if err := RequireVerifiedMFA(info); err != nil {
			writeError(w, AsError(err))
			return
		}
		next.ServeHTTP(w, r)
	})
}
