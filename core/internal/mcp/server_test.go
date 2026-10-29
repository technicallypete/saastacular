package mcp

import (
	"context"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// connect runs the SDK's initialize handshake in process: a synthetic test
// client over the SDK's in-memory transport. No network is involved.
func connect(t *testing.T, s *sdk.Server) (*sdk.ClientSession, *sdk.ServerSession) {
	t.Helper()
	ctx := context.Background()
	st, ct := sdk.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	cs, err := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "0.0.0"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client initialize: %v", err)
	}
	return cs, ss
}

func TestInitializeAndEmptyCatalog(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cs, ss := connect(t, NewServer(Options{Title: "Configured Title", Version: "0.0.0-test"}))

	init := cs.InitializeResult()
	if init == nil || init.ServerInfo == nil {
		t.Fatal("no initialize result")
	}
	if init.ServerInfo.Name != ServerName || init.ServerInfo.Title != "Configured Title" || init.ServerInfo.Version != "0.0.0-test" {
		t.Fatalf("server info = %+v", init.ServerInfo)
	}
	if init.Capabilities == nil || init.Capabilities.Tools == nil {
		t.Fatalf("tools capability not declared: %+v", init.Capabilities)
	}
	res, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	if len(res.Tools) != 0 {
		t.Fatalf("catalog must be empty, got %d tools", len(res.Tools))
	}
	if err := cs.Ping(ctx, nil); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if _, err := cs.CallTool(ctx, &sdk.CallToolParams{Name: "account.get_current_user"}); err == nil {
		t.Fatal("calling an unregistered tool must fail")
	}
	if err := cs.Close(); err != nil {
		t.Fatalf("client close: %v", err)
	}
	if err := ss.Wait(); err != nil {
		t.Fatalf("server session did not end cleanly: %v", err)
	}
}

// Each construction is independent: there is no shared, process-global
// catalog that a later session could inherit.
func TestServersAreIndependent(t *testing.T) {
	ctx := context.Background()
	a, b := NewServer(Options{Title: "A"}), NewServer(Options{Title: "B"})
	if a == b {
		t.Fatal("NewServer returned a shared instance")
	}
	a.AddTool(&sdk.Tool{Name: "probe.only_on_a", InputSchema: map[string]any{"type": "object"}},
		func(context.Context, *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
			return &sdk.CallToolResult{}, nil
		})
	csA, _ := connect(t, a)
	csB, _ := connect(t, b)
	ra, err := csA.ListTools(ctx, nil)
	if err != nil || len(ra.Tools) != 1 {
		t.Fatalf("server A tools = %v (%v)", ra, err)
	}
	rb, err := csB.ListTools(ctx, nil)
	if err != nil || len(rb.Tools) != 0 {
		t.Fatalf("server B must keep an empty catalog, got %v (%v)", rb, err)
	}
	if fresh, err := connectList(t, NewServer(Options{})); err != nil || fresh != 0 {
		t.Fatalf("a new server after others were modified must be empty: %d (%v)", fresh, err)
	}
	_ = csA.Close()
	_ = csB.Close()
}

func connectList(t *testing.T, s *sdk.Server) (int, error) {
	t.Helper()
	cs, _ := connect(t, s)
	defer cs.Close()
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		return -1, err
	}
	return len(res.Tools), nil
}
