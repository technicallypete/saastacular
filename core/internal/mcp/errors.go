package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"

	"core/internal/tenancy/access"
)

// The closed tool error codes (Plan r1 §3.5; contracts/mcp/control-plane.md). A tool error is a result with
// isError true, structuredContent {error, field?} and the code as text: no identifier, name, role or driver text.
const (
	CodeContextDenied      = "context_denied"
	CodeGroupNotSelectable = "group_not_selectable"
	CodeInvalidInput       = "invalid_input"
	CodeUnavailable        = "unavailable"
)

// The invalid_input fields.
const (
	FieldInvalidArguments = "invalid_arguments"
	FieldInvalidCursor    = "invalid_cursor"
	FieldInvalidPageSize  = "invalid_page_size"
	FieldInvalidGroupID   = "invalid_group_id"
)

// The application-defined JSON-RPC error codes of a refused tools/list (the message is the closed code).
const (
	RPCCodeContextDenied = -32040
	RPCCodeUnavailable   = -32041
)

// toolCode maps a facade error onto the closed tool codes: a lost context (or a session that ended during the call) is
// context_denied; an unselectable target group_not_selectable; invalid input invalid_input; everything else unavailable.
func toolCode(err error) string {
	switch access.CodeOf(err) {
	case access.CodeContextDenied, access.CodeUnauthenticated, access.CodeNotFound, access.CodeForbidden:
		return CodeContextDenied
	case access.CodeGroupNotSelectable:
		return CodeGroupNotSelectable
	case access.CodeInvalidInput:
		return CodeInvalidInput
	}
	return CodeUnavailable
}

// rpcError maps a refused tools/list onto its JSON-RPC error.
func rpcError(err error) *jsonrpc.Error {
	if toolCode(err) == CodeUnavailable {
		return &jsonrpc.Error{Code: RPCCodeUnavailable, Message: CodeUnavailable}
	}
	return &jsonrpc.Error{Code: RPCCodeContextDenied, Message: CodeContextDenied}
}

// methodNotFound refuses an unlisted method without echoing its name.
func methodNotFound() *jsonrpc.Error {
	return &jsonrpc.Error{Code: jsonrpc.CodeMethodNotFound, Message: "method not found"}
}
