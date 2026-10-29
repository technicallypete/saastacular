package mcp

import "strings"

// The MCP route constants (Plan r8 §3.2). The two public paths are served under the tenant hosts of the trusted
// registry only, through Handler.Wrap, before the server mux sees them; the internal setup routes are mounted on the mux
// with these exact patterns, only when MCP is configured. B-MCP-SOURCE pins them.
const (
	// ConnectionPathPrefix is the per-connection MCP endpoint: https://{host}/mcp/c/{uuid}.
	ConnectionPathPrefix = "/mcp/c/"
	// MetadataPathPrefix is the per-connection protected-resource metadata:
	// https://{host}/.well-known/oauth-protected-resource/mcp/c/{uuid}.
	MetadataPathPrefix = "/.well-known/oauth-protected-resource/mcp/c/"

	// The internal setup routes; G063 adds the rename route. (No trailing comment on these lines: B-MCP-SOURCE
	// reads each pattern as the whole quoted value at the end of its line.)
	RouteConnectionsList   = "GET /internal/v1/mcp/connections"
	RouteConnectionsCreate = "POST /internal/v1/mcp/connections"
	RouteConnectionsRevoke = "POST /internal/v1/mcp/connections/revoke"
	RouteConnectionsRename = "POST /internal/v1/mcp/connections/rename"
)

// connectionID returns the canonical lowercase UUID of an exact "<prefix><uuid>" raw path, false for anything else
// (trailing slash, extra segments, percent-encoding, upper case, a malformed or zero UUID).
func connectionID(rawPath, prefix string) (string, bool) {
	id, ok := strings.CutPrefix(rawPath, prefix)
	if !ok || !canonicalUUID(id) {
		return "", false
	}
	return id, true
}

// canonicalUUID accepts a nonzero lowercase hyphenated UUID (8-4-4-4-12 hexadecimal digits).
func canonicalUUID(s string) bool {
	if len(s) != 36 || s == "00000000-0000-0000-0000-000000000000" {
		return false
	}
	for i := 0; i < 36; i++ {
		c := s[i]
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
				return false
			}
		}
	}
	return true
}

// resourceURL is the exact resource of a connection: https://{host}/mcp/c/{uuid}.
func resourceURL(host, id string) string { return "https://" + host + ConnectionPathPrefix + id }

// metadataURL is the exact protected-resource metadata URL of a connection.
func metadataURL(host, id string) string { return "https://" + host + MetadataPathPrefix + id }
