// Package mcp is the core service's Model Context Protocol boundary (G014; ADR-005; contracts/mcp/control-plane.md;
// docs/foundation/MCP.md).
//
// Every MCP connection has its own tenant-hosted resource URL, https://{host}/mcp/c/{uuid}, created by an authenticated,
// membership-checked browser setup. WorkOS AuthKit Connect is the authorization server; core is the resource server and
// the local authority. Each POST verifies the Connect access token for the exact resource, maps its subject through the
// existing G004 principal, reads the connection, checks the binding (host, audience, client, live session, active owner)
// and then builds one SDK server for that call only, whose five tools act through the access facade in a fresh request
// scope. There is no process-global catalog, no session store and no cache: the transport is stateless (no
// Mcp-Session-Id), and every protected call reads current authority.
package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"core/internal/identity"
)

// ServerName identifies the core service to MCP clients.
const ServerName = "core"

// ProtocolVersions are the claimed MCP revisions (S0): 2025-11-25 and 2025-06-18. 2026-07-28 is not claimed: the SDK
// answers it with the SEP-2575 unsupported-version error, and the official client falls back to 2025-11-25.
var ProtocolVersions = []string{"2025-11-25", "2025-06-18"}

// Options configure a server. Title comes from configuration (the product name), never from a literal.
type Options struct {
	Title   string
	Version string
}

// NewServer returns a new server with the tools capability and the claimed protocol revisions. listChanged is false,
// truthfully: the stateless JSON transport has no stream that could deliver notifications/tools/list_changed, so a
// client re-lists instead (G015 notification rule, docs/foundation/MCP.md). It registers no tool; the per-call server of
// a verified connection adds its catalog (catalog.go).
func NewServer(opts Options) *sdk.Server {
	return sdk.NewServer(
		&sdk.Implementation{Name: ServerName, Title: opts.Title, Version: opts.Version},
		&sdk.ServerOptions{
			Capabilities:              &sdk.ServerCapabilities{Tools: &sdk.ToolCapabilities{ListChanged: false}},
			SupportedProtocolVersions: ProtocolVersions,
		},
	)
}

// The methods a connection may call (C9 lifecycle methods, the catalog and tool calls). Everything else is
// method-not-found: no resources, prompts, completions, logging, sampling, elicitation or roots.
var allowedMethods = map[string]bool{
	"initialize": true, "notifications/initialized": true, "notifications/cancelled": true, "ping": true,
	"tools/list": true, "tools/call": true,
}

// newServer builds the server of one verified call (G015 requirement-aware registration; Plan r4 §1): the control tools
// bound to the call's access scope (actx) and connection, registered with no check in front of them, and a receiving
// middleware that refuses unlisted methods and decides discovery on the call's own snapshot. tools/list runs one
// AuthorizeCatalog snapshot (refused as a whole when the context no longer holds) and registers the eligible
// capability-mapped tools. A tools/call naming anything but a control tool runs the same snapshot whether the name is a
// capability entry or absent: a context refusal registers nothing, so the SDK answers its unknown-tool error for the
// requested name, recorded once by the snapshot like a refused list; a store failure is unavailable for both; an allow
// registers the eligible entries and the SDK dispatches. A registered tool still authorizes its own execution.
func (h *Handler) newServer(actx context.Context, conn identity.MCPConnection) *sdk.Server {
	s := NewServer(Options{Title: h.cfg.Title, Version: h.cfg.Version})
	t := tools{h: h, ctx: actx, conn: conn}
	c := h.catalog()
	s.AddReceivingMiddleware(func(next sdk.MethodHandler) sdk.MethodHandler {
		return func(ctx context.Context, method string, req sdk.Request) (sdk.Result, error) {
			if !allowedMethods[method] {
				return nil, methodNotFound()
			}
			switch method {
			case "tools/list":
				names, err := h.acc.AuthorizeCatalog(actx, c.reqs)
				if err != nil {
					return nil, rpcError(err)
				}
				c.registerEligible(s, t, names)
			case "tools/call":
				if !c.isControl(callName(req)) {
					names, err := h.acc.AuthorizeCatalog(actx, c.reqs)
					switch {
					case err == nil:
						c.registerEligible(s, t, names)
					case toolCode(err) == CodeUnavailable:
						return nil, rpcError(err)
					}
				}
			}
			return next(ctx, method, req)
		}
	})
	c.registerControl(s, t)
	return s
}

// callName is the requested tool name of a tools/call request ("" when absent).
func callName(req sdk.Request) string {
	if r, ok := req.(*sdk.CallToolRequest); ok && r != nil && r.Params != nil {
		return r.Params.Name
	}
	return ""
}
