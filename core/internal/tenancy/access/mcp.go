package access

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"core/internal/audit"
	"core/internal/identity"
	"core/internal/tenancy"
)

// G014: the access facade of tenant-bound MCP connections (docs/foundation/MCP.md; Plan r8 §4). The MCP transport
// (core/internal/mcp) is the only importer. Every operation authorizes and acts in one database transaction, rechecks
// the connection's 'mcp' session, its owner and every relied-on fact, and writes only on an allow, ordered against
// conflicting writers by row locks. Selectors come only from the server-side connection record. Nothing is cached.

// mcpIdentity is the identity boundary of the MCP operations; *identity.Service satisfies it. Its transaction-scoped
// helpers run only inside this facade's transactions.
type mcpIdentity interface {
	InsertMCPConnectionTx(ctx context.Context, tx pgx.Tx, userID, organizationID, connectionID, host, resource string, l identity.MCPLifetimes,
		displayName *string) error
	ReadMCPConnectionSessionTx(ctx context.Context, tx pgx.Tx, userID, organizationID, connectionID string) (string, error)
	RenameMCPConnectionTx(ctx context.Context, tx pgx.Tx, userID, organizationID, connectionID string, name *string) error
	ActivateMCPConnectionTx(ctx context.Context, tx pgx.Tx, c identity.MCPConnection, clientID string) (identity.MCPActivation, error)
	ReadMCPSelectionTx(ctx context.Context, tx pgx.Tx, c identity.MCPConnection) (string, int64, bool, error)
	SelectMCPGroupTx(ctx context.Context, tx pgx.Tx, c identity.MCPConnection, expected, group string) (int64, error)
	ListMCPConnectionsTx(ctx context.Context, tx pgx.Tx, userID, organizationID string, after *identity.MCPListPosition) (identity.MCPConnectionPage, error)
	UserDisplayNameTx(ctx context.Context, tx pgx.Tx, userID string) (string, error)
}

// mcpHooks are inert test seams (nil in every production construction): selectLocked runs inside a selection write
// after its fact locks and before the connection row lock (O-MCP-1..4); activateLocked inside an activation after the
// Organization binding lock.
type mcpHooks struct {
	selectDecided   func() // after the switch's decision snapshot, before its guarded write
	selectLocked    func()
	activateLocked  func()
	catalogComputed func() // G015: inside the eligibility snapshot, after the eligible names are computed
}

// The logged MCP operations (access.decision; no capability: the MCP control surface evaluates membership and
// selected-Group validity, not a capability of the matrix).
const (
	opBeginVerifiedRequest = "access.begin_verified_request"
	opMCPCreate            = "mcp.connection.create"
	opMCPList              = "mcp.connection.list"
	opMCPRename            = "mcp.connection.rename" // G063
	opMCPActivate          = "mcp.connection.activate"
	opMCPAuthorize         = "mcp.authorize"
	opMCPContext           = "mcp.get_context"
	opMCPListGroups        = "mcp.list_groups"
	opMCPSwitchGroup       = "mcp.switch_group"
	// opMCPRequest is the access.denied operation of every protected MCP call (audit.OpMCPRequest).
	opMCPRequest = "mcp.request"
)

// ConnectionContext is the current context of the request's connection: the owner (with its untrusted display name),
// the immutable Organization (with its name), the optional selected Group and the context version.
type ConnectionContext struct {
	UserID           tenancy.UserID
	DisplayName      string
	OrganizationID   tenancy.OrganizationID
	OrganizationName string
	Group            *tenancy.Group // nil in Organization-only context
	ContextVersion   int64
}

// SelectableGroup is one Group of a listing page; Selected marks the connection's current selection.
type SelectableGroup struct {
	Group    tenancy.Group
	Selected bool
}

// GroupPage is one listing page, ordered by Group UUID; More reports whether another page follows its last Group.
type GroupPage struct {
	Groups []SelectableGroup
	More   bool
}

// BeginVerifiedRequest authenticates one protected MCP call from the identity-minted connection binding c (Plan r8
// §4.3 step 5): it returns a derived context carrying a new request scope for c's 'mcp' session and owner, and the end
// function the caller defers. It never resolves a credential; each operation of the scope rechecks the session, the
// owner and the facts. A zero binding or a context that already carries a scope is refused.
func (s *Service) BeginVerifiedRequest(ctx context.Context, c identity.MCPConnection) (context.Context, func(), error) {
	sc, f := s.beginVerified(ctx, c)
	s.logContext(ctx, opBeginVerifiedRequest, f)
	if f != nil {
		return ctx, func() {}, f.err()
	}
	return context.WithValue(ctx, scopeKey{}, sc), func() { sc.ended.Store(true) }, nil
}

func (s *Service) beginVerified(ctx context.Context, c identity.MCPConnection) (*requestScope, *failure) {
	if ctx.Value(scopeKey{}) != nil {
		return nil, fail(CodeContextDenied, reasonScopeNested)
	}
	if !c.Valid() || s.mcp == nil {
		return nil, fail(CodeUnauthenticated, reasonSessionInvalid)
	}
	subject, err := tenancy.ParseUserID(c.UserID())
	if err != nil {
		return nil, fail(CodeUnauthenticated, reasonSessionInvalid)
	}
	org, err := tenancy.ParseOrganizationID(c.OrganizationID())
	if err != nil {
		return nil, fail(CodeUnauthenticated, reasonSessionInvalid)
	}
	conn := c
	return &requestScope{issuer: s, session: c.SessionID(), subject: subject, subjectText: c.UserID(), conn: &conn, org: org}, nil
}

// mcpScope is the binding check of every MCP connection operation: a live scope of this Service begun by
// BeginVerifiedRequest.
func (s *Service) mcpScope(ctx context.Context) (*requestScope, *failure) {
	sc := scopeOf(ctx)
	switch {
	case sc == nil || sc.issuer != s:
		return nil, fail(CodeContextDenied, reasonScopeMissing)
	case sc.ended.Load():
		return nil, fail(CodeContextDenied, reasonScopeEnded)
	case sc.conn == nil || s.mcp == nil:
		return nil, fail(CodeContextDenied, reasonScopeMismatch)
	}
	return sc, nil
}

// logMCP writes the access.decision record of one MCP operation (outcome allow, deny or failed; no capability).
func (s *Service) logMCP(ctx context.Context, op string, f *failure) {
	s.logDecision(ctx, op, "", outcomeAllow, reasonAllowed, f)
}

// recordMCPDenial records a marked denial of a protected MCP call (operation mcp.request).
func (s *Service) recordMCPDenial(ctx context.Context, sc *requestScope, f *failure) {
	if sc != nil {
		s.recordDenial(ctx, sc.subject, sc.org, opMCPRequest, f)
	}
}

// RecordResourceMismatch records one global access.denied event (operation mcp.request, reason resource_mismatch,
// actor the connection's owner) when a verified, mapped subject presented a token at a host, connection or audience its
// own connection c is not bound to (Plan r8 §6.1). It is an independent record: a failure only writes the diagnostic.
func (s *Service) RecordResourceMismatch(ctx context.Context, c identity.MCPConnection) {
	subject, err := tenancy.ParseUserID(c.UserID())
	if err != nil {
		return
	}
	s.record(ctx, dbContext{user: subject}, audit.Event{Action: audit.AccessDenied, Outcome: audit.Denied,
		Reason: audit.ReasonResourceMismatch, Operation: audit.OpMCPRequest})
}

// mcpFailure maps an identity error of an MCP helper: unauthenticated, a concurrent selection change (a conflict: a
// denial with reason write_conflict, which the tool reports as unavailable and the client reconciles), or unavailable
// with the transaction's reason.
func mcpFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	switch {
	case errors.Is(err, identity.ErrUnauthenticated):
		return fail(CodeUnauthenticated, reasonSessionInvalid)
	case errors.Is(err, identity.ErrMCPSelectionChanged):
		return fail(CodeConflict, reasonWriteConflict)
	}
	return fail(CodeUnavailable, failureReason(ctx, tx.cause))
}

// readSelection reads the connection's current selection in tx and parses it.
func (s *Service) readSelection(ctx context.Context, tx *trackedTx, sc *requestScope) (tenancy.GroupID, string, int64, *failure) {
	raw, version, active, err := s.mcp.ReadMCPSelectionTx(ctx, tx, *sc.conn)
	if err != nil {
		return tenancy.GroupID{}, "", 0, mcpFailure(ctx, tx, err)
	}
	if !active {
		return tenancy.GroupID{}, "", 0, fail(CodeUnauthenticated, reasonSessionInvalid)
	}
	var g tenancy.GroupID
	if raw != "" {
		if g, err = tenancy.ParseGroupID(raw); err != nil {
			return tenancy.GroupID{}, "", 0, fail(CodeInternal, reasonFactInconsistent)
		}
	}
	return g, raw, version, nil
}

// snapshot runs fn in one REPEATABLE READ READ ONLY snapshot in the connection's Organization-only (TENANT-ORG) shape,
// after the session liveness check, the current selection read and the facts of the Organization and the selected
// Group, and only when the context still holds: a current Organization membership and, with a selection, a selected
// Group the owner may still select. A failed context is context_denied (recorded as access.denied mcp.request).
func (s *Service) snapshot(ctx context.Context, op string, fn func(tx *trackedTx, sc *requestScope, f Facts, sel contextSelection, version int64) *failure) error {
	return s.snapshotFor(ctx, op, nil, fn)
}

// snapshotFor is snapshot with the fact options of a catalog's requirements (factQueryFor). With no capability
// requirement it reads exactly what snapshot reads.
func (s *Service) snapshotFor(ctx context.Context, op string, reqs map[string]ToolRequirement, fn func(tx *trackedTx, sc *requestScope, f Facts, sel contextSelection, version int64) *failure) error {
	sc, f := s.mcpScope(ctx)
	if f == nil {
		f = s.inContextTx(ctx, false, dbContext{user: sc.subject, organization: sc.org}, cleanupUnbounded, func(tx *trackedTx) *failure {
			if err := s.id.CheckLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
				return sessionFailure(ctx, tx, err)
			}
			g, _, version, f := s.readSelection(ctx, tx, sc)
			if f != nil {
				return f
			}
			sel := contextSelection{Organization: sc.org, Group: g}
			q := factQueryFor(reqs, sel)
			q.User = sc.subject
			facts, err := tenancy.ReadFacts(ctx, tx, q)
			if err != nil {
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			}
			if r := contextCheck(sel, facts); r != "" {
				return markDenied(fail(CodeContextDenied, r), sel, facts, "", audit.UUID{}, tenancy.GroupID{})
			}
			return fn(tx, sc, facts, sel, version)
		})
		s.recordMCPDenial(ctx, sc, f)
	}
	s.logMCP(ctx, op, f)
	if f != nil {
		return f.err()
	}
	return nil
}

// AuthorizeConnection checks the context of the request's connection for tools/list and the read tools (Plan r8 §4.2):
// the 'mcp' session is live, the owner is active and is a current member of the Organization, and any selected Group is
// still selectable. It writes nothing.
func (s *Service) AuthorizeConnection(ctx context.Context) error {
	return s.snapshot(ctx, opMCPAuthorize, func(*trackedTx, *requestScope, Facts, contextSelection, int64) *failure { return nil })
}

// G015: requirement-aware discovery (docs/foundation/MCP.md "Discovery"; Plan r4 §1, §5.1). A tool declares one
// requirement. The binding, membership and context requirements are the G014 control tools' own handler checks (the
// transport registers those tools on every request and runs no check in front of them); a capability requirement is
// one of the eleven canonical capabilities and makes a tool discoverable only when AuthorizeCatalog finds it eligible.

// requirementKind is the kind of a ToolRequirement; zero is no requirement (invalid).
type requirementKind uint8

const (
	reqNone requirementKind = iota
	reqBinding
	reqMembership
	reqContext
	reqCapability
)

// ToolRequirement is the opaque, comparable requirement of one MCP tool. Its zero value is invalid. It carries no
// authority: it names what a later operation must check, never a result.
type ToolRequirement struct {
	kind requirementKind
	cap  Capability
}

// RequireBinding is the exact-binding requirement (logout: the G014 termination exception; no membership).
func RequireBinding() ToolRequirement { return ToolRequirement{kind: reqBinding} }

// RequireMembership is the current Organization membership requirement (switch_group: an explicit clear stays
// reachable after the selected Group is lost; a select rechecks the current and target Group in its guarded write).
func RequireMembership() ToolRequirement { return ToolRequirement{kind: reqMembership} }

// RequireContext is the membership and valid current selection requirement (the G014 read tools).
func RequireContext() ToolRequirement { return ToolRequirement{kind: reqContext} }

// RequireCapability is the requirement of one canonical capability, decided with the matrix in the current context.
// Any other name is invalid_input: no prefix, case or alias matching.
func RequireCapability(name string) (ToolRequirement, error) {
	c := Capability(name)
	if _, ok := capabilityKind(c); !ok {
		return ToolRequirement{}, fail(CodeInvalidInput, reasonCapabilityUnknown).err()
	}
	return ToolRequirement{kind: reqCapability, cap: c}, nil
}

// AuthorizeCatalog returns the names of reqs that are eligible for the request's connection now, in no particular
// order. It is discovery only: one read-only snapshot of AuthorizeConnection's shape with the fact options of the
// requirements (factQueryFor), which writes nothing and returns names, never a context, fact, decision or anything a
// later operation accepts. Every tool still authorizes its own execution. Its refusals are AuthorizeConnection's: a
// context that no longer holds is context_denied (recorded as access.denied mcp.request), liveness unauthenticated, a
// store failure unavailable. Eligibility of a capability is the eligible hint: decide allows the Organization or a
// candidate Group's target, or the role-level Organization owner/admin path holds, which asserts no Group or profile
// target exists. A zero requirement is never eligible.
func (s *Service) AuthorizeCatalog(ctx context.Context, reqs map[string]ToolRequirement) ([]string, error) {
	var out []string
	err := s.snapshotFor(ctx, opMCPAuthorize, reqs, func(_ *trackedTx, _ *requestScope, f Facts, sel contextSelection, _ int64) *failure {
		for name, r := range reqs {
			if eligible(r, sel, f) {
				out = append(out, name)
			}
		}
		if s.mcpHooks != nil && s.mcpHooks.catalogComputed != nil {
			s.mcpHooks.catalogComputed()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetConnectionContext returns the current context of the request's connection (account.get_current_user and
// account.get_current_context) on one snapshot.
func (s *Service) GetConnectionContext(ctx context.Context) (ConnectionContext, error) {
	var out ConnectionContext
	err := s.snapshot(ctx, opMCPContext, func(tx *trackedTx, sc *requestScope, f Facts, sel contextSelection, version int64) *failure {
		name, err := s.mcp.UserDisplayNameTx(ctx, tx, sc.subjectText)
		if err != nil {
			return mcpFailure(ctx, tx, err)
		}
		out = ConnectionContext{UserID: sc.subject, DisplayName: name, OrganizationID: f.Organization.ID,
			OrganizationName: f.Organization.Name, ContextVersion: version}
		if !zero(sel.Group) {
			g := findGroup(f, sel.Group)
			out.Group = &g
		}
		return nil
	})
	if err != nil {
		return ConnectionContext{}, err
	}
	return out, nil
}

// ListSelectableGroups returns one page (at most limit, ordered by Group UUID after afterGroupID) of the Groups the owner may
// select, evaluated in Organization-only context after the context check (Plan r1 §3.5): an Organization owner or admin
// every Group of the Organization, any other member exactly its own Groups. The current selection is marked.
func (s *Service) ListSelectableGroups(ctx context.Context, afterGroupID string, limit int) (GroupPage, error) {
	var out GroupPage
	var after tenancy.GroupID
	var pf *failure
	switch {
	case limit < 1 || limit > tenancy.MaxSelectableGroupsPage:
		pf = &failure{code: CodeInvalidInput, field: "invalid_page_size", reason: reasonInputInvalid}
	case afterGroupID != "":
		g, err := tenancy.ParseGroupID(afterGroupID)
		if err != nil {
			pf = &failure{code: CodeInvalidInput, field: "invalid_cursor", reason: reasonInputInvalid}
		}
		after = g
	}
	if pf != nil {
		s.logMCP(ctx, opMCPListGroups, pf)
		return GroupPage{}, pf.err()
	}
	err := s.snapshot(ctx, opMCPListGroups, func(tx *trackedTx, sc *requestScope, f Facts, sel contextSelection, _ int64) *failure {
		role, r := organizationRole(sel, f)
		if r != "" {
			return fail(CodeContextDenied, r)
		}
		manager := role == tenancy.OrgOwner || role == tenancy.OrgAdmin
		groups, err := tenancy.SelectableGroupsPage(ctx, tx, sc.org, sc.subject, manager, after, limit)
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		if len(groups) > limit {
			groups, out.More = groups[:limit], true
		}
		out.Groups = make([]SelectableGroup, 0, len(groups))
		for _, g := range groups {
			out.Groups = append(out.Groups, SelectableGroup{Group: g, Selected: g.ID == sel.Group})
		}
		return nil
	})
	if err != nil {
		return GroupPage{}, err
	}
	return out, nil
}

// CodeGroupNotSelectable is the switch refusal of an unknown, foreign or unselectable Group (identical for all three).
const CodeGroupNotSelectable Code = "group_not_selectable"

// SelectConnectionGroup selects the Group groupID for the request's connection, or clears the selection when groupID is "" (Plan
// r8 §4.6 M3; r1 §4.4), and returns the resulting context. A clear needs a current Organization membership only (the
// Spec's explicit-clear exception); a select also needs the current selection (if any) to be still selectable and the
// target to be selectable. Selecting the current Group or clearing an empty selection changes nothing. The write is
// one READ COMMITTED transaction in the target's TENANT-GROUP shape (TENANT-ORG for a clear) with the lock order users
// -> sessions -> organizations -> Organization binding -> target Group binding and row -> connection row; a concurrent
// selection committed between the plain read and the row lock writes nothing (unavailable).
func (s *Service) SelectConnectionGroup(ctx context.Context, groupID string) (ConnectionContext, error) {
	var group tenancy.GroupID
	sc, f := s.mcpScope(ctx)
	if f == nil && groupID != "" {
		g, err := tenancy.ParseGroupID(groupID)
		if err != nil {
			f = &failure{code: CodeInvalidInput, field: "invalid_group_id", reason: reasonInputInvalid}
		}
		group = g
	}
	if f == nil {
		f = s.selectGroup(ctx, sc, group)
		s.recordMCPDenial(ctx, sc, f)
	}
	s.logMCP(ctx, opMCPSwitchGroup, f)
	if f != nil {
		return ConnectionContext{}, f.err()
	}
	return s.GetConnectionContext(ctx)
}

func (s *Service) selectGroup(ctx context.Context, sc *requestScope, target tenancy.GroupID) *failure {
	// The decision snapshot (Organization-only shape): membership, the current selection's validity and the target's
	// selectability; it decides no-ops and records policy refusals. The guarded write below rechecks all of them on
	// locked facts in the same Organization-only shape, where the subject's own bindings in both Groups are visible
	// (G008), and passes the snapshot's selection as the expected value: any selection change committed in between makes
	// it write nothing (write_conflict).
	noop := false
	expected := ""
	var current tenancy.GroupID
	pre := s.inContextTx(ctx, false, dbContext{user: sc.subject, organization: sc.org}, cleanupUnbounded, func(tx *trackedTx) *failure {
		if err := s.id.CheckLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
			return sessionFailure(ctx, tx, err)
		}
		cur, currentRaw, _, f := s.readSelection(ctx, tx, sc)
		if f != nil {
			return f
		}
		current, expected = cur, currentRaw
		groups := []tenancy.GroupID{}
		for _, g := range []tenancy.GroupID{target, current} {
			if !zero(g) && !containsGroupID(groups, g) {
				groups = append(groups, g)
			}
		}
		facts, err := tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: sc.org, User: sc.subject, Groups: groups})
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		orgOnly := contextSelection{Organization: sc.org}
		if r := contextCheck(orgOnly, facts); r != "" {
			return markDenied(fail(CodeContextDenied, r), orgOnly, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		if zero(target) {
			noop = zero(current)
			return nil
		}
		if !zero(current) && current != target {
			cur := contextSelection{Organization: sc.org, Group: current}
			if r := contextCheck(cur, facts); r != "" {
				return markDenied(fail(CodeContextDenied, r), orgOnly, facts, "", audit.UUID{}, tenancy.GroupID{})
			}
		}
		sel := contextSelection{Organization: sc.org, Group: target}
		if r := contextCheck(sel, facts); r != "" {
			if r == reasonGroupNotSelectable {
				d := markDenied(fail(CodeContextDenied, r), orgOnly, facts, "", audit.UUID{}, tenancy.GroupID{})
				d.code = CodeGroupNotSelectable
				return d
			}
			return fail(CodeContextDenied, r)
		}
		noop = current == target
		return nil
	})
	if pre != nil || noop {
		return pre
	}
	if s.mcpHooks != nil && s.mcpHooks.selectDecided != nil {
		s.mcpHooks.selectDecided()
	}
	// The guarded write, Organization-only for a select and a clear alike (Plan r8 §4.4 lock order): the user and the
	// session, then the Organization, the Organization binding, the subject's bindings in the previous and target Groups
	// and those Group rows (ReadFacts lock mode), then the connection row. The guard re-validates both Groups.
	return s.inContextTx(ctx, true, dbContext{user: sc.subject, organization: sc.org}, cleanupUnbounded, func(tx *trackedTx) *failure {
		if err := s.id.LockLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
			return sessionFailure(ctx, tx, err)
		}
		groups := []tenancy.GroupID{} // a clear relies on the membership only
		if !zero(target) {
			for _, g := range []tenancy.GroupID{target, current} {
				if !zero(g) && !containsGroupID(groups, g) {
					groups = append(groups, g)
				}
			}
		}
		facts, err := tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: sc.org, User: sc.subject, Groups: groups, Lock: true})
		if err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
		// The decision on the locked facts; refusals are recorded like the snapshot's. The guard re-validates as
		// defense in depth only, so a raced refusal is decided here and never surfaces as a constraint failure.
		orgOnly := contextSelection{Organization: sc.org}
		if r := contextCheck(orgOnly, facts); r != "" {
			return markDenied(fail(CodeContextDenied, r), orgOnly, facts, "", audit.UUID{}, tenancy.GroupID{})
		}
		if !zero(target) {
			// Spec R5: after the selected Group is lost only a clear (or logout) succeeds.
			if !zero(current) && current != target {
				if r := contextCheck(contextSelection{Organization: sc.org, Group: current}, facts); r != "" {
					return markDenied(fail(CodeContextDenied, r), orgOnly, facts, "", audit.UUID{}, tenancy.GroupID{})
				}
			}
			if r := contextCheck(contextSelection{Organization: sc.org, Group: target}, facts); r != "" {
				if r == reasonGroupNotSelectable {
					d := markDenied(fail(CodeContextDenied, r), orgOnly, facts, "", audit.UUID{}, tenancy.GroupID{})
					d.code = CodeGroupNotSelectable
					return d
				}
				return fail(CodeContextDenied, r)
			}
		}
		if s.mcpHooks != nil && s.mcpHooks.selectLocked != nil {
			s.mcpHooks.selectLocked()
		}
		next := ""
		if !zero(target) {
			next = target.String()
		}
		if _, err := s.mcp.SelectMCPGroupTx(ctx, tx, *sc.conn, expected, next); err != nil {
			return mcpFailure(ctx, tx, err)
		}
		return nil
	})
}

func containsGroupID(gs []tenancy.GroupID, g tenancy.GroupID) bool {
	for _, x := range gs {
		if x == g {
			return true
		}
	}
	return false
}

// ActivateConnection activates the request's pending connection for clientID (Plan r8 §4.6 M2): one TENANT-ORG write
// transaction that locks the owner, the connection's 'mcp' session (live), the Organization and the owner's
// Organization binding (a current membership is required), then the connection row, and activates it only while it is
// still pending and unexpired; the capture trigger records mcp.connection_established in the same transaction (strict).
// Exactly one of several concurrent first tokens wins; a later one succeeds only for the same client. Any refusal is
// unauthenticated and the connection stays pending.
func (s *Service) ActivateConnection(ctx context.Context, clientID string) error {
	sc, f := s.mcpScope(ctx)
	if f == nil {
		f = s.inContextTx(ctx, true, dbContext{user: sc.subject, organization: sc.org}, cleanupUnbounded, func(tx *trackedTx) *failure {
			if err := s.id.LockLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
				return sessionFailure(ctx, tx, err)
			}
			facts, err := tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: sc.org, User: sc.subject, Groups: []tenancy.GroupID{}, Lock: true})
			if err != nil {
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			}
			if _, r := organizationRole(contextSelection{Organization: sc.org}, facts); r != "" {
				return fail(CodeUnauthenticated, r)
			}
			if s.mcpHooks != nil && s.mcpHooks.activateLocked != nil {
				s.mcpHooks.activateLocked()
			}
			a, err := s.mcp.ActivateMCPConnectionTx(ctx, tx, *sc.conn, clientID)
			if err != nil {
				return mcpFailure(ctx, tx, err)
			}
			if !a.Activated && (a.State != identity.MCPStateActive || a.ClientID != clientID) {
				return fail(CodeUnauthenticated, reasonSessionInvalid)
			}
			return nil
		})
	}
	s.logMCP(ctx, opMCPActivate, f)
	if f != nil {
		return f.err()
	}
	return nil
}

// CreateMCPConnection creates a pending connection for the subject of tc in tc's Organization (Plan r8 §3.4, §4.6 M1):
// one TENANT-ORG write transaction that locks the user and the browser session (live), the Organization and the
// subject's Organization binding (a current membership is required), then inserts the 'mcp' session and the connection
// row, whose policies require a current Organization role again. tc must be Organization-only. connectionID, host and
// resource are minted and built by the trusted MCP layer, never request input. displayName is the optional friendly name
// (G063; nil for unnamed), already normalized by identity.NormalizeMCPDisplayName; it grants nothing.
func (s *Service) CreateMCPConnection(ctx context.Context, tc *TenantContext, connectionID, host, resource string, l identity.MCPLifetimes,
	displayName *string) error {
	sc, f := s.bind(ctx, tc)
	if f == nil && (!zero(tc.group) || s.mcp == nil) {
		f = fail(CodeContextDenied, reasonScopeMismatch)
	}
	if f == nil {
		f = s.inContextTx(ctx, true, dbContext{user: tc.subject, organization: tc.organization}, cleanupUnbounded, func(tx *trackedTx) *failure {
			if err := s.id.LockLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
				return sessionFailure(ctx, tx, err)
			}
			facts, err := tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: tc.organization, User: tc.subject, Groups: []tenancy.GroupID{}, Lock: true})
			if err != nil {
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			}
			if r := contextCheck(tc.selection(), facts); r != "" {
				return fail(CodeContextDenied, r)
			}
			if err := s.mcp.InsertMCPConnectionTx(ctx, tx, sc.subjectText, tc.organization.String(), connectionID, host, resource, l,
				displayName); err != nil {
				if errors.Is(err, identity.ErrInvalidRequest) || errors.Is(err, identity.ErrInvalidMCPName) {
					return fail(CodeInvalidInput, reasonInputInvalid)
				}
				return mcpFailure(ctx, tx, err)
			}
			return nil
		})
	}
	s.logMCP(ctx, opMCPCreate, f)
	if f != nil {
		return f.err()
	}
	return nil
}

// RenameMCPConnection sets (or, with a nil name, clears) the friendly name of one own live pending or active connection
// of tc's subject in tc's Organization (G063 Plan r1 §4.2; D-4). tc must be Organization-only. One TENANT-ORG write
// transaction, in this statement order: (1) the live user and the browser session FOR SHARE (as CreateMCPConnection);
// (2) the plain read of the connection's immutable 'mcp' session; (3) that target session FOR SHARE through the same
// liveness statements, which rereads its revocation, both expiries and the owner's status under the lock; (4) the
// locked facts with a current Organization membership; (5) the connection row FOR NO KEY UPDATE, last, and the
// name-only write. Every lock before the row is FOR SHARE and taken user -> session -> row, the order of DisableUser,
// revocation, activation (M2) and selection (M3), so a concurrent revocation either commits first (and is observed in
// step 3) or waits for the rename; nothing deadlocks and nothing is revived or renewed. An unknown, foreign, terminal or
// expired target is not_found (the caller's own session is live; the target is not one of its live connections); a
// dead caller session is unauthenticated; a non-member is context_denied; a name that is not normalized is invalid_input.
// connectionID and name are request input; neither grants anything.
func (s *Service) RenameMCPConnection(ctx context.Context, tc *TenantContext, connectionID string, name *string) error {
	sc, f := s.bind(ctx, tc)
	if f == nil && (!zero(tc.group) || s.mcp == nil) {
		f = fail(CodeContextDenied, reasonScopeMismatch)
	}
	if f == nil {
		f = s.inContextTx(ctx, true, dbContext{user: tc.subject, organization: tc.organization}, cleanupUnbounded, func(tx *trackedTx) *failure {
			if err := s.id.LockLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
				return sessionFailure(ctx, tx, err)
			}
			target, err := s.mcp.ReadMCPConnectionSessionTx(ctx, tx, sc.subjectText, tc.organization.String(), connectionID)
			if err != nil {
				return renameFailure(ctx, tx, err)
			}
			if err := s.id.LockLiveSession(ctx, tx, target, sc.subjectText); err != nil {
				return renameFailure(ctx, tx, err)
			}
			facts, err := tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: tc.organization, User: tc.subject, Groups: []tenancy.GroupID{}, Lock: true})
			if err != nil {
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			}
			if r := contextCheck(tc.selection(), facts); r != "" {
				return fail(CodeContextDenied, r)
			}
			if err := s.mcp.RenameMCPConnectionTx(ctx, tx, sc.subjectText, tc.organization.String(), connectionID, name); err != nil {
				if errors.Is(err, identity.ErrInvalidMCPName) {
					return fail(CodeInvalidInput, reasonInputInvalid)
				}
				return renameFailure(ctx, tx, err)
			}
			return nil
		})
	}
	s.logMCP(ctx, opMCPRename, f)
	if f != nil {
		return f.err()
	}
	return nil
}

// renameFailure maps an identity error about the rename's target: unauthenticated (unknown, foreign, terminal or
// expired target) is not_found with reason target_not_found; anything else is unavailable with the transaction's reason.
func renameFailure(ctx context.Context, tx *trackedTx, err error) *failure {
	if errors.Is(err, identity.ErrUnauthenticated) {
		return fail(CodeNotFound, reasonTargetNotFound)
	}
	return fail(CodeUnavailable, failureReason(ctx, tx.cause))
}

// ListMCPConnections lists one page of the subject's own live connections in tc's Organization (the setup page; Plan r8
// §3.2; newest first, after the untrusted keyset position when one is given) on one snapshot that also requires a
// current Organization membership. Every page re-runs every check; the position never narrows or replaces one.
func (s *Service) ListMCPConnections(ctx context.Context, tc *TenantContext, after *identity.MCPListPosition) (identity.MCPConnectionPage, error) {
	var out identity.MCPConnectionPage
	sc, f := s.bind(ctx, tc)
	if f == nil && s.mcp == nil {
		f = fail(CodeContextDenied, reasonScopeMismatch)
	}
	if f == nil {
		f = s.inContextTx(ctx, false, dbContext{user: tc.subject, organization: tc.organization}, cleanupUnbounded, func(tx *trackedTx) *failure {
			if err := s.id.CheckLiveSession(ctx, tx, sc.session, sc.subjectText); err != nil {
				return sessionFailure(ctx, tx, err)
			}
			facts, err := tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: tc.organization, User: tc.subject, Groups: []tenancy.GroupID{}})
			if err != nil {
				return fail(CodeUnavailable, failureReason(ctx, tx.cause))
			}
			if r := contextCheck(contextSelection{Organization: tc.organization}, facts); r != "" {
				return fail(CodeContextDenied, r)
			}
			list, err := s.mcp.ListMCPConnectionsTx(ctx, tx, sc.subjectText, tc.organization.String(), after)
			if err != nil {
				return mcpFailure(ctx, tx, err)
			}
			out = list
			return nil
		})
	}
	s.logMCP(ctx, opMCPList, f)
	if f != nil {
		return identity.MCPConnectionPage{}, f.err()
	}
	return out, nil
}
