package identity

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// G014: tenant-bound MCP connections (docs/foundation/MCP.md; contracts/mcp/control-plane.md; migration 00010).
//
// An MCP connection is a server-side record created by an authenticated browser setup. It owns exactly one 'mcp'
// session row (no issued bearer: the credential hash is the hash of a discarded random value and the CSRF value is
// never returned), which carries liveness (idle and absolute expiry), revocation and G013 restore invalidation. Core
// never issues, stores or refreshes MCP bearer values: a verified WorkOS Connect access token only authenticates the
// subject, which maps to the canonical user through the existing G004 principal (E2), and every protected call reads the
// connection, its session and the user again. Nothing here is cached.

// MCPRenewEvery bounds idle renewal of an MCP session (Plan r8 §4.6 M5): at most one renewal per hour per connection.
const MCPRenewEvery = time.Hour

// MCP connection states (app.mcp_connections.state).
const (
	MCPStatePending = "pending"
	MCPStateActive  = "active"
)

// MCPLifetimes are the configured connection lifetimes (MCP_PENDING_TTL, MCP_CONNECTION_IDLE, MCP_CONNECTION_ABSOLUTE).
type MCPLifetimes struct {
	Pending  time.Duration
	Idle     time.Duration
	Absolute time.Duration
}

// MCPConnection is one connection binding read by identity (unexported fields): the record, its session's liveness and
// the owner's status, as of one read. Only identity mints a non-zero value; it is a binding, not an authority snapshot.
type MCPConnection struct {
	id, session, user, organization string
	host, resource, state, clientID string
	selectedGroup                   string // "" when no Group is selected
	contextVersion                  int64
	pendingExpires                  time.Time
	renewedAt                       time.Time
}

// ID returns the connection UUID.
func (c MCPConnection) ID() string { return c.id }

// SessionID returns the UUID of the connection's 'mcp' session.
func (c MCPConnection) SessionID() string { return c.session }

// UserID returns the canonical owner UUID.
func (c MCPConnection) UserID() string { return c.user }

// OrganizationID returns the immutable Organization UUID.
func (c MCPConnection) OrganizationID() string { return c.organization }

// Host returns the exact tenant host the connection was created on.
func (c MCPConnection) Host() string { return c.host }

// Resource returns the exact resource URL (the only accepted token audience).
func (c MCPConnection) Resource() string { return c.resource }

// Active reports whether the connection completed activation.
func (c MCPConnection) Active() bool { return c.state == MCPStateActive }

// ClientID returns the client bound at activation ("" while pending).
func (c MCPConnection) ClientID() string { return c.clientID }

// SelectedGroupID returns the stored selected Group UUID, "" for Organization-only context.
func (c MCPConnection) SelectedGroupID() string { return c.selectedGroup }

// ContextVersion returns the stored context version.
func (c MCPConnection) ContextVersion() int64 { return c.contextVersion }

// RenewalDue reports whether the session's last renewal is at least MCPRenewEvery before now.
func (c MCPConnection) RenewalDue(now time.Time) bool {
	return !now.Before(c.renewedAt.Add(MCPRenewEvery))
}

// Valid reports whether c was minted by identity.
func (c MCPConnection) Valid() bool {
	return c.id != "" && c.session != "" && c.user != "" && c.organization != ""
}

// MCPConnectionInfo is one own live connection as the setup page lists it: no session, token or user field.
type MCPConnectionInfo struct {
	ID          string
	Resource    string
	State       string
	ClientID    string
	CreatedAt   time.Time
	ActivatedAt time.Time // zero while pending
	// DisplayName is the owner's friendly name (G063), nil when unnamed (every legacy row). Presentation metadata only:
	// never identity, client binding, tenant selection or authority.
	DisplayName *string
}

// MCPNameMax is the largest friendly name in Unicode code points after trimming (G063 Plan r1 §4.1).
const MCPNameMax = 80

// ErrInvalidMCPName reports a friendly name that is not valid UTF-8, contains a control character or a line/paragraph
// separator anywhere, or is longer than MCPNameMax code points after trimming. It is refused, never sanitized.
var ErrInvalidMCPName = errors.New("identity: invalid mcp connection name")

// NormalizeMCPDisplayName is the authoritative friendly-name rule (G063 Plan r1 §4.1, D-6), in this order: raw must be
// valid UTF-8; any control character (Cc: C0 including tab, newline, vertical tab, form feed and carriage return, DEL,
// C1 including U+0085) or line/paragraph separator (Zl/Zp: U+2028, U+2029) anywhere is refused before trimming; the
// surrounding Unicode White_Space is then trimmed (with Cc/Zl/Zp already refused the effective set is SP, U+00A0,
// U+1680, U+2000-U+200A, U+202F, U+205F and U+3000; U+FEFF is not White_Space and counts); an empty result is nil (clear
// or unnamed); otherwise 1..MCPNameMax code points. Markup is kept verbatim (it is never interpreted). The wire layer
// proves that malformed UTF-8 and unpaired surrogate escapes were refused rather than replaced before raw exists.
func NormalizeMCPDisplayName(raw string) (*string, error) {
	if !utf8.ValidString(raw) {
		return nil, ErrInvalidMCPName
	}
	for _, r := range raw {
		if unicode.IsControl(r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return nil, ErrInvalidMCPName
		}
	}
	name := strings.TrimFunc(raw, unicode.IsSpace)
	if name == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(name) > MCPNameMax {
		return nil, ErrInvalidMCPName
	}
	return &name, nil
}

// validMCPName reports whether name is nil or already in its normalized form (defense in depth for the Tx helpers).
func validMCPName(name *string) bool {
	if name == nil {
		return true
	}
	n, err := NormalizeMCPDisplayName(*name)
	return err == nil && n != nil && *n == *name
}

// MCPListPage is the setup listing's page size (newest first; the listing reads one more row to detect a next page).
const MCPListPage = 100

// MCPListPosition is a keyset position of the setup listing: the next page holds the rows strictly older than
// (CreatedAt, ID). It is untrusted and confers nothing; the listing's owner, Organization and liveness filters bound
// every page whatever position is presented, and the anchor row need not still exist.
type MCPListPosition struct {
	CreatedAt time.Time
	ID        string
}

// MCPConnectionPage is one page of the setup listing; More reports that older live connections follow its last row.
type MCPConnectionPage struct {
	Connections []MCPConnectionInfo
	More        bool
}

// MCPActivation is the outcome of an activation attempt (Plan r8 §4.6 M2): the state after the locked decision and the
// bound client. Activated is true only for the one winning transaction.
type MCPActivation struct {
	State     string
	ClientID  string
	Activated bool
}

// NewMCPConnectionID returns a random (version 4) canonical UUID for a new connection: core mints it, never a caller.
func NewMCPConnectionID() (string, error) { return newUserID() }

// CSRFMatches reports, in constant time, whether token is the base64url form of the session's CSRF value (the
// server-rendered setup forms carry it, as the account forms do).
func CSRFMatches(info SessionInfo, token string) bool {
	got, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(info.CSRFToken) != secretLen {
		return false
	}
	return len(got) == len(info.CSRFToken) && subtle.ConstantTimeCompare(got, info.CSRFToken) == 1
}

// Now returns the service clock (microsecond precision, UTC): the one time source of the MCP lifecycle.
func (s *Service) Now() time.Time { return s.now() }

// MapConnectSubject maps the subject of a verified Connect access token to the canonical user through the existing
// principal of the configured G004 issuer (E2; Plan r8 §3.3 namespace rule). There is no user creation and no email
// linking: an unmapped subject is unauthenticated.
func (s *Service) MapConnectSubject(ctx context.Context, subject string) (string, error) {
	if subject == "" || s.cfg.Issuer == "" {
		return "", ErrUnauthenticated
	}
	id, found, err := s.store.lookupUser(ctx, s.cfg.Issuer, subject)
	if err != nil {
		return "", err
	}
	if !found {
		return "", ErrUnauthenticated
	}
	return id, nil
}

// ReadMCPConnection reads one connection of userID (the mapped subject) in org (the trusted registry's Organization) in
// a TENANT-ORG-shaped transaction and evaluates its liveness at the service clock: the 'mcp' session is unrevoked and
// before both expiries, the owner is active and, for a pending connection, the setup has not expired. Unknown, foreign,
// terminal and malformed all give unauthenticated.
func (s *Service) ReadMCPConnection(ctx context.Context, userID, organizationID, connectionID string) (MCPConnection, error) {
	if !canonicalUserID(userID) || !canonicalUserID(organizationID) || !canonicalUserID(connectionID) {
		return MCPConnection{}, ErrUnauthenticated
	}
	c, live, err := s.store.readMCPBinding(ctx, userID, organizationID, connectionID, s.now())
	if err != nil {
		return MCPConnection{}, err
	}
	if !live {
		return MCPConnection{}, ErrUnauthenticated
	}
	return c, nil
}

// RenewMCPConnection renews the idle expiry of c's session when renewal is due (Plan r8 §4.6 M5), before any tool runs:
// one conditional UPDATE in the IDENTITY shape that re-checks kind, revocation, both expiries, the throttle and the active
// owner in the write itself, capped at the absolute expiry. A throttled live session is a no-op; anything else (expired,
// revoked, unknown, wrong kind, disabled owner) is unauthenticated and nothing is extended.
func (s *Service) RenewMCPConnection(ctx context.Context, c MCPConnection, idle time.Duration) error {
	if !c.Valid() || idle <= 0 {
		return ErrUnauthenticated
	}
	return s.store.renewMCPSession(ctx, c.user, c.session, s.now(), idle)
}

// RevokeMCPConnection terminates c (the logout tool; Plan r8 §4.6 M4): the existing revokeSession transaction on its
// 'mcp' session with reason logout (revocation priority; G011). No membership is needed and the connection can never be
// used again; the provider grant is not touched (normal logout is local).
func (s *Service) RevokeMCPConnection(ctx context.Context, c MCPConnection) error {
	if !c.Valid() {
		return ErrUnauthenticated
	}
	if err := s.store.revokeSession(ctx, c.user, c.session, revokeLogout, s.now()); err != nil {
		return err
	}
	s.event(ctx, slog.LevelInfo, "auth.session.revoked", slog.String("outcome", OutcomeLocalRevoked), slog.String("scope", ScopeMCP))
	return nil
}

// RevokeOwnMCPConnection terminates one own connection of userID in org (the setup page's revoke; M4): exact ownership
// only (no membership) and connection-local. An unknown, foreign or already terminal connection is unauthenticated.
func (s *Service) RevokeOwnMCPConnection(ctx context.Context, userID, organizationID, connectionID string) error {
	if !canonicalUserID(userID) || !canonicalUserID(organizationID) || !canonicalUserID(connectionID) {
		return ErrUnauthenticated
	}
	c, live, err := s.store.readMCPBinding(ctx, userID, organizationID, connectionID, s.now())
	if err != nil {
		return err
	}
	if !live {
		return ErrUnauthenticated // unknown, foreign or already terminal: not one of the caller's live connections
	}
	return s.RevokeMCPConnection(ctx, c)
}

// ScopeMCP is the auth.session.revoked scope of an MCP connection termination (logout tool or setup-page revoke).
const ScopeMCP = "mcp"
