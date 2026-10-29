package identity

import (
	"context"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
)

// LogoutInput is one core logout request (Plan r5 §4.4, §7.2).
type LogoutInput struct {
	Scope      string // current, all or provider
	CSRF       string
	Origin     string
	Credential string
}

// LogoutOutput is the committed logout result. ProviderLogoutURL is set only after a confirmed provider revocation;
// ReturnTo is the fixed local return of a provider sign-out.
type LogoutOutput struct {
	Outcome           string
	ProviderLogoutURL string
	ReturnTo          string
	// LogoutRef is the one-use reference to the recorded provider outcome (scope provider only). Web carries it in
	// a host-only cookie; it never appears in a URL (G004 C7 addendum §2).
	LogoutRef string
}

// providerReturnPath is the fixed local return of the provider sign-out: the sign-in page registered with the
// provider as the sign-out return for each origin, without a query.
const providerReturnPath = "/sign-in"

// Logout applies one sign-out scope. Rejection order after the HTTP layer: scope (400) → origin structure (403
// origin_rejected, before any session or routing lookup) → session bound to this origin (401 for all and provider;
// current is idempotent) → CSRF (403, live sessions only). Local revocation is always committed before any
// success is reported or any provider call is made; a commit failure is identity_unavailable, never success.
func (s *Service) Logout(ctx context.Context, in LogoutInput) (out LogoutOutput, err error) {
	defer func() {
		outcome := out.Outcome
		if err != nil {
			outcome = outcomeOf(err)
		}
		s.event(ctx, slog.LevelInfo, "auth.session.revoked", slog.String("outcome", outcome), slog.String("scope", scopeOf(in.Scope)))
	}()
	if in.Scope != ScopeCurrent && in.Scope != ScopeAll && in.Scope != ScopeProvider {
		return LogoutOutput{}, ErrInvalidRequest
	}
	// G016 LOW-1: a missing or malformed origin is origin_rejected for every scope, decided on its structure alone, so
	// no routing lookup precedes the local revocation; a well-formed origin that routes nowhere still revokes.
	if !s.logoutOrigin(in.Origin) {
		return LogoutOutput{}, ErrOriginRejected
	}
	// G016 security amendment (L-1, Plan amendment A2): the session resolves only at its own origin, and host routing
	// no longer gates revocation, so sign-out still revokes on a host that stopped routing. The tenant host is resolved
	// only afterwards, to pick a provider sign-out return origin (providerLogout).
	info, err := s.Session(ctx, in.Credential, in.Origin)
	if err == ErrUnauthenticated {
		if in.Scope == ScopeCurrent {
			return LogoutOutput{Outcome: OutcomeLocalRevoked}, nil // idempotent: nothing live to revoke
		}
		return LogoutOutput{}, ErrUnauthenticated
	}
	if err != nil {
		return LogoutOutput{}, err
	}
	csrf, ok := decodeSecret(in.CSRF)
	if !ok || !equalSecret(csrf, info.CSRFToken) {
		return LogoutOutput{}, ErrCSRFFailed
	}
	now := s.now()
	switch in.Scope {
	case ScopeCurrent:
		if err := s.store.revokeSession(ctx, info.UserID, info.SessionID, revokeLogout, now); err != nil {
			return LogoutOutput{}, err
		}
		return LogoutOutput{Outcome: OutcomeLocalRevoked}, nil
	case ScopeAll:
		if _, err := s.store.revokeAll(ctx, info.UserID, revokeLogoutAll, now); err != nil {
			return LogoutOutput{}, err
		}
		return LogoutOutput{Outcome: OutcomeLocalRevoked}, nil
	}
	return s.providerLogout(ctx, info, in.Origin)
}

// maxOriginLen bounds an origin (as Login and tenantOrigin do).
const maxOriginLen = 300

// logoutOrigin reports whether origin is structurally one a session can be bound to: an exact configured non-tenant
// origin (IDENTITY_PUBLIC_ORIGINS, validated with the cookie profile's scheme rules at startup), or a canonical https
// origin (tenant hosts resolve as https only): lowercase, a host and an optional canonical port, no userinfo, path,
// query or fragment, at most maxOriginLen bytes. It is pure: whether the host routes is never asked here.
func (s *Service) logoutOrigin(origin string) bool {
	if origin == "" || len(origin) > maxOriginLen {
		return false
	}
	if s.cfg.Origins.Allowed(origin) {
		return true
	}
	// An empty port ("https://host:") round-trips through url.Parse with Port() == "", so the trailing colon is
	// rejected explicitly: it is not the canonical form of any origin.
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.Hostname() == "" || strings.HasSuffix(u.Host, ":") ||
		u.User != nil || u.Opaque != "" || u.Path != "" || u.RawPath != "" || u.RawQuery != "" || u.ForceQuery ||
		u.Fragment != "" || u.String() != origin || strings.ToLower(origin) != origin {
		return false
	}
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err != nil || n < 1 || n > 65535 || strconv.Itoa(n) != p {
			return false
		}
	}
	return true
}

// providerLogout is the fixed §3.5 sequence: commit the local revocation, revoke exactly this provider session (5 s),
// record the outcome with a one-use reference, and only when the revocation is confirmed and a safe return origin is
// resolved redirect to the provider's logout URL. A failed or timed-out revocation never redirects to the provider.
// G016: the return origin is the requesting tenant origin, or the shared auth host for a custom host (Plan r4 §1.7);
// it is resolved only after the revocations (L-1). A host that no longer routes, or a resolution failure, suppresses
// the hosted redirect while the outcome still reports the actual provider result.
func (s *Service) providerLogout(ctx context.Context, info SessionInfo, requestOrigin string) (LogoutOutput, error) {
	now := s.now()
	if err := s.store.revokeSession(ctx, info.UserID, info.SessionID, revokeProviderLogout, now); err != nil {
		return LogoutOutput{}, err
	}
	outcome := providerRevocationFailed
	if s.Configured() && info.ProviderSessionID != "" {
		rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), revokeDeadline)
		if err := s.cfg.Provider.RevokeProviderSession(rctx, info.ProviderSessionID); err == nil {
			outcome = providerRevocationOK
		}
		cancel()
	}
	ref, err := newSecret()
	if err != nil {
		return LogoutOutput{}, ErrUnavailable
	}
	if err := s.store.recordProviderLogout(ctx, info.UserID, info.SessionID, outcome, hashSecret(ref), now.Add(logoutRefTTL)); err != nil {
		return LogoutOutput{}, err
	}
	logoutRef := encodeSecret(ref)
	if outcome != providerRevocationOK {
		s.event(ctx, slog.LevelWarn, "auth.provider.revocation", slog.String("outcome", OutcomeProviderFailed), slog.String("scope", ScopeProvider))
		return LogoutOutput{Outcome: OutcomeProviderFailed, ReturnTo: providerReturnPath, LogoutRef: logoutRef}, nil
	}
	s.event(ctx, slog.LevelInfo, "auth.provider.revocation", slog.String("outcome", OutcomeProviderRevoked), slog.String("scope", ScopeProvider))
	tn, err := s.tenantOrigin(ctx, s.store.db, requestOrigin)
	if err != nil {
		// ErrOriginRejected (the host no longer routes: its sign-in page would be the gate's 404) or ErrUnavailable: no
		// safe return origin, so no hosted sign-out redirect; the local and provider revocations already happened.
		return LogoutOutput{Outcome: OutcomeProviderRevoked, ReturnTo: providerReturnPath, LogoutRef: logoutRef}, nil
	}
	origin := requestOrigin
	if tn.Kind == HostCustom && s.cfg.AuthOrigin != "" {
		origin = s.cfg.AuthOrigin
	}
	// The provider returns to exactly the registered sign-in page of the return origin; the reference stays out of
	// every URL.
	logoutURL, err := s.cfg.Provider.ProviderLogoutURL(info.ProviderSessionID, origin+providerReturnPath)
	if err != nil {
		return LogoutOutput{Outcome: OutcomeProviderRevoked, ReturnTo: providerReturnPath, LogoutRef: logoutRef}, nil
	}
	s.event(ctx, slog.LevelInfo, "auth.provider.revocation", slog.String("outcome", OutcomeProviderRedirect), slog.String("scope", ScopeProvider))
	return LogoutOutput{Outcome: OutcomeProviderRedirect, ProviderLogoutURL: logoutURL, ReturnTo: providerReturnPath, LogoutRef: logoutRef}, nil
}

// ProviderLogoutReturn consumes a logout reference once and returns the server-recorded provider outcome. Unknown,
// expired or reused references return the generic local outcome with no provider claim.
func (s *Service) ProviderLogoutReturn(ctx context.Context, ref string) (string, error) {
	raw, ok := decodeSecret(ref)
	if !ok {
		return OutcomeLocalRevoked, nil
	}
	recorded, found, err := s.store.consumeLogoutRef(ctx, hashSecret(raw), s.now())
	if err != nil {
		return "", err
	}
	if !found {
		return OutcomeLocalRevoked, nil
	}
	s.event(ctx, slog.LevelInfo, "auth.provider.revocation", slog.String("outcome", OutcomeProviderReturned), slog.String("scope", ScopeProvider))
	if recorded == providerRevocationOK {
		return OutcomeProviderRevoked, nil
	}
	return OutcomeProviderFailed, nil
}

func scopeOf(s string) string {
	switch s {
	case ScopeCurrent, ScopeAll, ScopeProvider:
		return s
	}
	return "invalid"
}
