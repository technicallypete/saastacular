package mcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"core/internal/identity"
)

// The five tools of the tenant-bound control surface (Plan r1 §3.5, r8 §3.5). The names contain dots as fixed by the
// Architecture; compatibility is claimed only for tested clients. There is no Organization or agency switching tool and
// no argument carries an Organization selector.
const (
	ToolGetCurrentUser    = "account.get_current_user"
	ToolGetCurrentContext = "account.get_current_context"
	ToolListGroups        = "context.list_groups"
	ToolSwitchGroup       = "context.switch_group"
	ToolLogout            = "logout"
)

// ToolNames are the exact catalog, in registration order.
var ToolNames = []string{ToolGetCurrentUser, ToolGetCurrentContext, ToolListGroups, ToolSwitchGroup, ToolLogout}

// The list_groups page bounds.
const (
	defaultGroupPage = 50
	maxGroupPage     = 100
)

func object(props map[string]any, required ...string) map[string]any {
	s := map[string]any{"type": "object", "properties": props, "additionalProperties": false}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

var (
	uuidSchema     = map[string]any{"type": "string", "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"}
	groupSchema    = object(map[string]any{"group_id": uuidSchema, "slug": map[string]any{"type": "string"}, "name": map[string]any{"type": "string"}}, "group_id", "slug", "name")
	contextSchema  = object(map[string]any{"organization_id": uuidSchema, "organization_name": map[string]any{"type": "string"}, "group": map[string]any{"anyOf": []any{map[string]any{"type": "null"}, groupSchema}}, "context_version": map[string]any{"type": "integer", "minimum": 0}}, "organization_id", "organization_name", "group", "context_version")
	emptySchema    = object(map[string]any{})
	toolSchemaSets = map[string][2]map[string]any{
		ToolGetCurrentUser:    {emptySchema, object(map[string]any{"user_id": uuidSchema, "display_name": map[string]any{"type": []any{"string", "null"}}}, "user_id", "display_name")},
		ToolGetCurrentContext: {emptySchema, contextSchema},
		ToolListGroups: {
			object(map[string]any{"cursor": map[string]any{"type": "string", "maxLength": 128}, "limit": map[string]any{"type": "integer", "minimum": 1, "maximum": maxGroupPage}}),
			object(map[string]any{"groups": map[string]any{"type": "array", "items": object(map[string]any{"group_id": uuidSchema, "slug": map[string]any{"type": "string"}, "name": map[string]any{"type": "string"}, "selected": map[string]any{"type": "boolean"}}, "group_id", "slug", "name", "selected")}, "next_cursor": map[string]any{"type": []any{"string", "null"}}}, "groups", "next_cursor"),
		},
		ToolSwitchGroup: {object(map[string]any{"group_id": map[string]any{"anyOf": []any{map[string]any{"type": "null"}, uuidSchema}}}, "group_id"), contextSchema},
		ToolLogout:      {emptySchema, object(map[string]any{"outcome": map[string]any{"const": "terminated"}}, "outcome")},
	}
)

// ToolSchemas returns the input and output schema of a tool (U-MCP-TOOLS).
func ToolSchemas(name string) (input, output map[string]any, ok bool) {
	s, ok := toolSchemaSets[name]
	return s[0], s[1], ok
}

// tools binds the catalog's tools to one verified call: its access scope (ctx) and connection. Registration is the
// catalog's (catalog.go): the control tools on every request, capability-mapped tools only when eligible.
type tools struct {
	h    *Handler
	ctx  context.Context
	conn identity.MCPConnection
}

func (t tools) handler(name string) sdk.ToolHandler {
	return func(_ context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		var raw json.RawMessage
		if req != nil && req.Params != nil {
			raw = req.Params.Arguments
		}
		switch name {
		case ToolGetCurrentUser:
			return t.getCurrentUser(raw), nil
		case ToolGetCurrentContext:
			return t.getCurrentContext(raw), nil
		case ToolListGroups:
			return t.listGroups(raw), nil
		case ToolSwitchGroup:
			return t.switchGroup(raw), nil
		}
		return t.logout(raw), nil
	}
}

// toolError is a closed tool error result.
func toolError(code, field string) *sdk.CallToolResult {
	sc := map[string]any{"error": code}
	if field != "" {
		sc["field"] = field
	}
	return &sdk.CallToolResult{IsError: true, StructuredContent: sc, Content: []sdk.Content{&sdk.TextContent{Text: code}}}
}

// result is a successful structured result with its JSON text.
func result(v any) *sdk.CallToolResult {
	b, err := json.Marshal(v)
	if err != nil {
		return toolError(CodeUnavailable, "")
	}
	return &sdk.CallToolResult{StructuredContent: json.RawMessage(b), Content: []sdk.Content{&sdk.TextContent{Text: string(b)}}}
}

// decodeArgs strictly decodes the arguments object (absent means {}) into dst: unknown fields, trailing data and
// non-object values are invalid.
func decodeArgs(raw json.RawMessage, dst any) bool {
	if len(bytes.TrimSpace(raw)) == 0 || string(bytes.TrimSpace(raw)) == "null" {
		raw = json.RawMessage("{}")
	}
	if t := bytes.TrimSpace(raw); len(t) == 0 || t[0] != '{' {
		return false
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return false
	}
	_, err := dec.Token()
	return errors.Is(err, io.EOF)
}

type noArgs struct{}

func contextOutput(c ContextView) map[string]any {
	var group any
	if c.Group != nil {
		group = map[string]any{"group_id": c.Group.ID, "slug": c.Group.Slug, "name": c.Group.Name}
	}
	return map[string]any{"organization_id": c.OrganizationID, "organization_name": c.OrganizationName, "group": group,
		"context_version": c.ContextVersion}
}

// ContextView is the wire view of the connection context.
type ContextView struct {
	UserID           string
	DisplayName      string
	OrganizationID   string
	OrganizationName string
	Group            *GroupView
	ContextVersion   int64
}

// GroupView is the wire view of a Group.
type GroupView struct{ ID, Slug, Name string }

func (t tools) context() (ContextView, string) {
	c, err := t.h.acc.GetConnectionContext(t.ctx)
	if err != nil {
		return ContextView{}, toolCode(err)
	}
	v := ContextView{UserID: c.UserID.String(), DisplayName: c.DisplayName, OrganizationID: c.OrganizationID.String(),
		OrganizationName: c.OrganizationName, ContextVersion: c.ContextVersion}
	if c.Group != nil {
		v.Group = &GroupView{ID: c.Group.ID.String(), Slug: c.Group.Slug, Name: c.Group.Name}
	}
	return v, ""
}

func (t tools) getCurrentUser(raw json.RawMessage) *sdk.CallToolResult {
	if !decodeArgs(raw, &noArgs{}) {
		return toolError(CodeInvalidInput, FieldInvalidArguments)
	}
	c, code := t.context()
	if code != "" {
		return toolError(code, "")
	}
	var name any
	if c.DisplayName != "" {
		name = c.DisplayName
	}
	return result(map[string]any{"user_id": c.UserID, "display_name": name})
}

func (t tools) getCurrentContext(raw json.RawMessage) *sdk.CallToolResult {
	if !decodeArgs(raw, &noArgs{}) {
		return toolError(CodeInvalidInput, FieldInvalidArguments)
	}
	c, code := t.context()
	if code != "" {
		return toolError(code, "")
	}
	return result(contextOutput(c))
}

// The opaque list cursor: base64url (no padding) of "1|<bound Organization UUID>|<last Group UUID>". It confers nothing:
// listing is filtered by current authority, and a cursor of another Organization or of another shape is invalid.
const cursorVersion = "1"

func encodeCursor(org, last string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(cursorVersion + "|" + org + "|" + last))
}

func decodeCursor(s, org string) (string, bool) {
	b, err := base64.RawURLEncoding.Strict().DecodeString(s)
	if err != nil {
		return "", false
	}
	parts := strings.Split(string(b), "|")
	if len(parts) != 3 || parts[0] != cursorVersion || parts[1] != org || !canonicalUUID(parts[2]) {
		return "", false
	}
	return parts[2], true
}

func (t tools) listGroups(raw json.RawMessage) *sdk.CallToolResult {
	var in struct {
		Cursor *string `json:"cursor"`
		Limit  *int    `json:"limit"`
	}
	if !decodeArgs(raw, &in) {
		return toolError(CodeInvalidInput, FieldInvalidArguments)
	}
	limit := defaultGroupPage
	if in.Limit != nil {
		if *in.Limit < 1 || *in.Limit > maxGroupPage {
			return toolError(CodeInvalidInput, FieldInvalidPageSize)
		}
		limit = *in.Limit
	}
	org := t.conn.OrganizationID()
	after := ""
	if in.Cursor != nil {
		var ok bool
		if after, ok = decodeCursor(*in.Cursor, org); !ok {
			return toolError(CodeInvalidInput, FieldInvalidCursor)
		}
	}
	page, err := t.h.acc.ListSelectableGroups(t.ctx, after, limit)
	if err != nil {
		return toolError(toolCode(err), "")
	}
	groups := make([]map[string]any, 0, len(page.Groups))
	for _, g := range page.Groups {
		groups = append(groups, map[string]any{"group_id": g.Group.ID.String(), "slug": g.Group.Slug, "name": g.Group.Name,
			"selected": g.Selected})
	}
	var next any
	if page.More && len(page.Groups) > 0 {
		next = encodeCursor(org, page.Groups[len(page.Groups)-1].Group.ID.String())
	}
	return result(map[string]any{"groups": groups, "next_cursor": next})
}

func (t tools) switchGroup(raw json.RawMessage) *sdk.CallToolResult {
	var in map[string]json.RawMessage
	if !decodeArgs(raw, &in) || len(in) != 1 {
		return toolError(CodeInvalidInput, FieldInvalidArguments)
	}
	v, ok := in["group_id"]
	if !ok {
		return toolError(CodeInvalidInput, FieldInvalidArguments)
	}
	target := ""
	if string(bytes.TrimSpace(v)) != "null" {
		if err := json.Unmarshal(v, &target); err != nil || !canonicalUUID(target) {
			return toolError(CodeInvalidInput, FieldInvalidGroupID)
		}
	}
	c, err := t.h.acc.SelectConnectionGroup(t.ctx, target)
	if err != nil {
		return toolError(toolCode(err), "")
	}
	view := ContextView{OrganizationID: c.OrganizationID.String(), OrganizationName: c.OrganizationName, ContextVersion: c.ContextVersion}
	if c.Group != nil {
		view.Group = &GroupView{ID: c.Group.ID.String(), Slug: c.Group.Slug, Name: c.Group.Name}
	}
	return result(contextOutput(view))
}

// logout terminates this connection (exact-binding self-termination: the Spec exception that needs no membership). A
// later call with any token for this URL is 401.
func (t tools) logout(raw json.RawMessage) *sdk.CallToolResult {
	if !decodeArgs(raw, &noArgs{}) {
		return toolError(CodeInvalidInput, FieldInvalidArguments)
	}
	if err := t.h.id.RevokeMCPConnection(t.ctx, t.conn); err != nil {
		return toolError(CodeUnavailable, "")
	}
	return result(map[string]any{"outcome": "terminated"})
}
