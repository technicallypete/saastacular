package access

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
	"core/internal/tenancy"
)

// U-CTX-MCP (G014 Plan r8 §10.1; G015 Plan r4 §6.2 adds AuthorizeCatalog): the MCP scope binding before any
// transaction. A zero connection binding is refused; a context that already carries a scope is refused; every MCP
// operation outside an MCP scope of this Service, after the scope ended, or in a browser scope is context_denied
// without touching the database (the nil pool would panic).

// stubMCP satisfies the identity boundary; it fails the test on any call.
type stubMCP struct{ t *testing.T }

func (s stubMCP) fail() { s.t.Helper(); s.t.Error("identity MCP helper called") }
func (s stubMCP) InsertMCPConnectionTx(context.Context, pgx.Tx, string, string, string, string, string, identity.MCPLifetimes, *string) error {
	s.fail()
	return nil
}
func (s stubMCP) ReadMCPConnectionSessionTx(context.Context, pgx.Tx, string, string, string) (string, error) {
	s.fail()
	return "", nil
}
func (s stubMCP) RenameMCPConnectionTx(context.Context, pgx.Tx, string, string, string, *string) error {
	s.fail()
	return nil
}
func (s stubMCP) ActivateMCPConnectionTx(context.Context, pgx.Tx, identity.MCPConnection, string) (identity.MCPActivation, error) {
	s.fail()
	return identity.MCPActivation{}, nil
}
func (s stubMCP) ReadMCPSelectionTx(context.Context, pgx.Tx, identity.MCPConnection) (string, int64, bool, error) {
	s.fail()
	return "", 0, false, nil
}
func (s stubMCP) SelectMCPGroupTx(context.Context, pgx.Tx, identity.MCPConnection, string, string) (int64, error) {
	s.fail()
	return 0, nil
}
func (s stubMCP) ListMCPConnectionsTx(context.Context, pgx.Tx, string, string, *identity.MCPListPosition) (identity.MCPConnectionPage, error) {
	s.fail()
	return identity.MCPConnectionPage{}, nil
}
func (s stubMCP) UserDisplayNameTx(context.Context, pgx.Tx, string) (string, error) {
	s.fail()
	return "", nil
}

func mcpOps(s *Service) map[string]func(context.Context) error {
	return map[string]func(context.Context) error{
		"ActivateConnection":  func(ctx context.Context) error { return s.ActivateConnection(ctx, "client_01X") },
		"AuthorizeConnection": s.AuthorizeConnection,
		"AuthorizeCatalog": func(ctx context.Context) error {
			names, err := s.AuthorizeCatalog(ctx, map[string]ToolRequirement{"test.audit_read": {kind: reqCapability, cap: capAuditRead}})
			if names != nil {
				return &Error{Code: "names returned on a refusal"}
			}
			return err
		},
		"GetConnectionContext": func(ctx context.Context) error {
			_, err := s.GetConnectionContext(ctx)
			return err
		},
		"ListSelectableGroups": func(ctx context.Context) error {
			_, err := s.ListSelectableGroups(ctx, "", 10)
			return err
		},
		"SelectConnectionGroup": func(ctx context.Context) error {
			_, err := s.SelectConnectionGroup(ctx, "")
			return err
		},
	}
}

func TestMCPScopeBinding(t *testing.T) {
	s := newService(nil, nil, nil)
	s.mcp = stubMCP{t}
	ctx := context.Background()

	if _, _, err := s.BeginVerifiedRequest(ctx, identity.MCPConnection{}); CodeOf(err) != CodeUnauthenticated {
		t.Fatalf("zero binding: %v", err)
	}
	nested := context.WithValue(ctx, scopeKey{}, &requestScope{issuer: s})
	if _, _, err := s.BeginVerifiedRequest(nested, identity.MCPConnection{}); CodeOf(err) != CodeContextDenied {
		t.Fatalf("nested scope: %v", err)
	}

	other := newService(nil, nil, nil)
	browser := &requestScope{issuer: s, session: "s", subjectText: "u"}
	ended := &requestScope{issuer: s, conn: &identity.MCPConnection{}}
	ended.ended.Store(true)
	foreign := &requestScope{issuer: other, conn: &identity.MCPConnection{}}
	for name, c := range map[string]context.Context{
		"no scope":      ctx,
		"browser scope": context.WithValue(ctx, scopeKey{}, browser),
		"ended scope":   context.WithValue(ctx, scopeKey{}, ended),
		"foreign scope": context.WithValue(ctx, scopeKey{}, foreign),
	} {
		for op, f := range mcpOps(s) {
			if err := f(c); CodeOf(err) != CodeContextDenied {
				t.Errorf("%s, %s: %v, want context_denied", name, op, err)
			}
		}
	}
	// A Service without the identity MCP helpers refuses every MCP operation even in an MCP-shaped scope.
	bare := newService(nil, nil, nil)
	sc := &requestScope{issuer: bare, conn: &identity.MCPConnection{}}
	for op, f := range mcpOps(bare) {
		if err := f(context.WithValue(ctx, scopeKey{}, sc)); CodeOf(err) != CodeContextDenied {
			t.Errorf("no helpers, %s: %v", op, err)
		}
	}
	if _, _, err := bare.BeginVerifiedRequest(ctx, identity.MCPConnection{}); CodeOf(err) != CodeUnauthenticated {
		t.Errorf("no helpers: %v", err)
	}

	// G063: the rename is Organization-only and needs the identity MCP helpers; both refusals are scope mismatches
	// before any transaction (the nil pool would panic; the stub fails on any helper call).
	subject, err := tenancy.ParseUserID("0f63a000-0000-4000-8000-0000000000a1")
	if err != nil {
		t.Fatal(err)
	}
	org, err := tenancy.ParseOrganizationID("0f63a000-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	group, err := tenancy.ParseGroupID("0f63a000-0000-4000-8000-000000000011")
	if err != nil {
		t.Fatal(err)
	}
	name := "Laptop"
	for label, svc := range map[string]*Service{"group-selected context": s, "no identity MCP helpers": bare} {
		rs := &requestScope{issuer: svc, session: "s", subject: subject, subjectText: subject.String()}
		tc := &TenantContext{issuer: svc, scope: rs, subject: subject, session: "s", organization: org}
		if label == "group-selected context" {
			tc.group = group
		}
		err := svc.RenameMCPConnection(context.WithValue(ctx, scopeKey{}, rs), tc, "0f63a000-0000-4000-8000-000000000c01", &name)
		var e *Error
		if !asError(err, &e) || e.Code != CodeContextDenied {
			t.Errorf("rename, %s: %v, want context_denied", label, err)
		}
	}
}

func TestMCPListBounds(t *testing.T) {
	s := newService(nil, nil, nil)
	s.mcp = stubMCP{t}
	for _, c := range []struct {
		after string
		limit int
		field string
	}{{"", 0, "invalid_page_size"}, {"", 101, "invalid_page_size"}, {"not-a-uuid", 10, "invalid_cursor"}} {
		_, err := s.ListSelectableGroups(context.Background(), c.after, c.limit)
		var e *Error
		if !asError(err, &e) || e.Code != CodeInvalidInput || e.Field != c.field {
			t.Errorf("%+v: %v", c, err)
		}
	}
}

func asError(err error, target **Error) bool {
	e, ok := err.(*Error)
	if ok {
		*target = e
	}
	return ok
}
