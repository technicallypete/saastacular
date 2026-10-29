package access

import (
	"context"
	"errors"
	"sync/atomic"

	"core/internal/identity"
	"core/internal/tenancy"
)

// requestScope is the trusted per-request binding (G007 Plan r1 §3.2): one authenticated session and its subject,
// created only by BeginRequest and identified by pointer identity, so two requests never share a scope, even with one
// credential. It holds no credential, CSRF token, display name, provider session, assurance, role or allow result, and
// it is a binding, not an authority snapshot: every operation rechecks the session, the user and the facts.
type requestScope struct {
	issuer      *Service
	session     string // the session UUID, as identity reports it
	subject     tenancy.UserID
	subjectText string // the subject UUID, as identity reports it
	ended       atomic.Bool
	// G014: set only by BeginVerifiedRequest: the identity-minted MCP connection binding of this call (its 'mcp'
	// session is session) and its immutable Organization. Nil for a browser scope (BeginRequest).
	conn *identity.MCPConnection
	// org is the Organization this scope is confined to: the MCP connection's (BeginVerifiedRequest) or, for a browser
	// scope (BeginRequest), the session's binding (G016 security amendment). Never zero in an issued scope; Resolve
	// refuses every other Organization.
	org tenancy.OrganizationID
}

// scopeKey is the unexported context key of the request scope: no other package can place a scope in a context.
type scopeKey struct{}

func scopeOf(ctx context.Context) *requestScope {
	sc, _ := ctx.Value(scopeKey{}).(*requestScope)
	return sc
}

// BeginRequest authenticates one request: it resolves the opaque credential through identity.Service.Session (the
// accepted G004 path, including its throttled renewal) and returns a derived context carrying a new request scope,
// and the end function, which the request owner defers. end is idempotent and safe for concurrent use; after it, the
// scope and every context resolved in it are refused. Derived contexts (deadline, cancellation, values) keep the
// scope. A context that already carries a scope is refused before identity is called: one request chain has one
// authenticated subject. G016 security amendment: origin is the request's exact tenant origin (server-derived by the
// caller); the session resolves only at the origin it is bound to, and the scope is confined to the session's
// Organization binding.
func (s *Service) BeginRequest(ctx context.Context, credential, origin string) (context.Context, func(), error) {
	sc, f := s.beginRequest(ctx, credential, origin)
	s.logContext(ctx, opBeginRequest, f)
	if f != nil {
		return ctx, func() {}, f.err()
	}
	return context.WithValue(ctx, scopeKey{}, sc), func() { sc.ended.Store(true) }, nil
}

func (s *Service) beginRequest(ctx context.Context, credential, origin string) (*requestScope, *failure) {
	if ctx.Value(scopeKey{}) != nil {
		return nil, fail(CodeContextDenied, reasonScopeNested)
	}
	info, err := s.id.Session(ctx, credential, origin)
	switch {
	case errors.Is(err, identity.ErrUnauthenticated):
		return nil, fail(CodeUnauthenticated, reasonSessionInvalid)
	case err != nil:
		return nil, fail(CodeUnavailable, failureReason(ctx, nil))
	}
	subject, err := tenancy.ParseUserID(info.UserID)
	if err != nil || info.SessionID == "" {
		return nil, fail(CodeUnavailable, reasonStoreUnavailable)
	}
	// The session's Organization binding confines the scope; an absent or malformed binding fails closed.
	org, err := tenancy.ParseOrganizationID(info.OrganizationID)
	if err != nil || zero(org) {
		return nil, fail(CodeUnavailable, reasonStoreUnavailable)
	}
	return &requestScope{issuer: s, session: info.SessionID, subject: subject, subjectText: info.UserID, org: org}, nil
}

// TenantContext is an immutable, opaque request-scoped security envelope: the authenticated subject, its session, the
// active Organization and the optional selected Group, bound to the exact request scope it was resolved in. Only
// Resolve produces a usable one; a caller-built value is refused. It stores no request ID or other correlation value,
// role, capability, assurance or allow result, and confers nothing: every guarded operation revalidates it.
type TenantContext struct {
	issuer       *Service
	scope        *requestScope
	subject      tenancy.UserID
	session      string
	organization tenancy.OrganizationID
	group        tenancy.GroupID
}

// Subject returns the authenticated global user.
func (tc *TenantContext) Subject() tenancy.UserID { return tc.subject }

// Organization returns the active Organization.
func (tc *TenantContext) Organization() tenancy.OrganizationID { return tc.organization }

// Group returns the selected Group and true, or false in Organization-only context.
func (tc *TenantContext) Group() (tenancy.GroupID, bool) { return tc.group, !zero(tc.group) }

func (tc *TenantContext) selection() contextSelection {
	return contextSelection{Organization: tc.organization, Group: tc.group}
}

// ResolveRequest carries the context selectors. They are selectors, never authority: canonical UUIDs of an
// Organization and, optionally (empty means Organization-only), a Group of it.
type ResolveRequest struct {
	OrganizationID string
	GroupID        string
}

// Resolve resolves a trusted context in the current request scope. In one REPEATABLE READ READ ONLY snapshot it
// checks that the scope's session is still live, that the Organization exists and the subject holds a current
// Organization membership in it, and, with a Group selector, that the Group belongs to it and is selectable (an
// Organization owner or admin, or a binding in that exact Group). An unknown or foreign Organization or Group, no
// membership and a Group binding without an Organization binding are the same context_denied. Nothing is written and
// there is no default Organization. Several contexts may be resolved in one scope.
//
// G011: a policy refusal (organization_not_permitted, group_not_selectable) is recorded as one durable access.denied
// event (operation access.resolve) after the snapshot ended and before the operational record; the Organization only
// when its gate passed, never the Group selector. A successful Resolve writes nothing.
func (s *Service) Resolve(ctx context.Context, req ResolveRequest) (*TenantContext, error) {
	tc, f := s.resolve(ctx, req)
	if f != nil && f.denied != nil {
		if sc := scopeOf(ctx); sc != nil {
			org, _ := tenancy.ParseOrganizationID(req.OrganizationID)
			s.recordDenial(ctx, sc.subject, org, opResolve, f)
		}
	}
	s.logContext(ctx, opResolve, f)
	if f != nil {
		return nil, f.err()
	}
	return tc, nil
}

func (s *Service) resolve(ctx context.Context, req ResolveRequest) (*TenantContext, *failure) {
	sc := scopeOf(ctx)
	if sc == nil || sc.issuer != s {
		return nil, fail(CodeContextDenied, reasonScopeMissing)
	}
	if sc.ended.Load() {
		return nil, fail(CodeContextDenied, reasonScopeEnded)
	}
	org, err := tenancy.ParseOrganizationID(req.OrganizationID)
	if err != nil {
		return nil, inputFailure(err, reasonSelectorInvalid)
	}
	// G016 security amendment: a scope resolves only its own Organization (the browser session's binding or the MCP
	// connection's); any other one, or a scope without a binding, is organization_not_permitted before the snapshot,
	// recorded as one access.denied event without the requested Organization.
	if zero(sc.org) || org != sc.org {
		return nil, &failure{code: CodeContextDenied, reason: reasonOrganizationNotPermitted, denied: &denialScope{}}
	}
	var group tenancy.GroupID
	if req.GroupID != "" {
		if group, err = tenancy.ParseGroupID(req.GroupID); err != nil {
			return nil, inputFailure(err, reasonSelectorInvalid)
		}
	}
	sel := contextSelection{Organization: org, Group: group}
	groups := []tenancy.GroupID{}
	if !zero(group) {
		groups = []tenancy.GroupID{group}
	}
	f := s.inContextTx(ctx, false, dbContext{user: sc.subject, organization: org, group: group}, cleanupUnbounded, func(tx *trackedTx) *failure {
		if err := s.id.CheckLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
			return sessionFailure(ctx, tx, err)
		}
		facts, err := tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: org, User: sc.subject, Groups: groups})
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if r := contextCheck(sel, facts); r != "" {
			return markDenied(fail(CodeContextDenied, r), sel, facts, "", [16]byte{}, tenancy.GroupID{})
		}
		return nil
	})
	if f != nil {
		return nil, f
	}
	return &TenantContext{issuer: s, scope: sc, subject: sc.subject, session: sc.session, organization: org, group: group}, nil
}

// bind is the binding check of every guarded operation (G007 Plan r1 §3.2, B1-B6). It is pure and runs before any
// transaction; the first failing condition decides the reason, and every failure is context_denied.
func (s *Service) bind(ctx context.Context, tc *TenantContext) (*requestScope, *failure) {
	switch {
	case tc == nil || tc.issuer == nil || tc.scope == nil: // B1: nil, zero-valued or caller-built
		return nil, fail(CodeContextDenied, reasonContextForged)
	case tc.issuer != s: // B2: issued by another Service
		return nil, fail(CodeContextDenied, reasonContextForged)
	}
	sc := scopeOf(ctx)
	switch {
	case sc == nil || sc.issuer != s: // B3: outside an authenticated request of this Service
		return nil, fail(CodeContextDenied, reasonScopeMissing)
	case sc.ended.Load(): // B4: after the request ended
		return nil, fail(CodeContextDenied, reasonScopeEnded)
	case tc.scope != sc: // B5: resolved in another request scope
		return nil, fail(CodeContextDenied, reasonScopeMismatch)
	case tc.subject != sc.subject || tc.session != sc.session: // B6: defence in depth; cannot fail when B5 passes
		return nil, fail(CodeContextDenied, reasonSubjectMismatch)
	}
	return sc, nil
}
