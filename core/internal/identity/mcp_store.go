package identity

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// The MCP connection statements (G014; migration 00010). The binding read runs in its own TENANT-ORG-shaped transaction
// (initTenantOrg) and the renewal in an IDENTITY one; the transaction-scoped helpers (the methods ending in Tx) run only
// inside the access facade's tenant-shaped transaction, whose context statement is already first. Every statement names
// its row explicitly in addition to row-level security.
const (
	sqlMCPBinding = `SELECT c.id::text, c.session_id::text, c.user_id::text, c.organization_id::text, c.host, c.resource, c.state, ` +
		`c.pending_expires_at, c.client_id, c.selected_group_id::text, c.context_version, s.kind, s.revoked_at IS NULL, ` +
		`s.idle_expires_at, s.absolute_expires_at, s.renewed_at, u.status ` +
		`FROM app.mcp_connections c JOIN app.sessions s ON s.id = c.session_id JOIN app.users u ON u.id = c.user_id ` +
		`WHERE c.id = $1 AND c.user_id = $2 AND c.organization_id = $3`
	// M5: the conditional renewal and, when it updates nothing, the same liveness without the throttle.
	sqlMCPRenew = `UPDATE app.sessions SET idle_expires_at = LEAST($3::timestamptz, absolute_expires_at), renewed_at = $2 ` +
		`WHERE id = $1 AND user_id = $4 AND kind = 'mcp' AND revoked_at IS NULL AND idle_expires_at > $2 AND absolute_expires_at > $2 ` +
		`AND renewed_at <= $5 AND EXISTS (SELECT 1 FROM app.users u WHERE u.id = $4 AND u.status = 'active') RETURNING idle_expires_at`
	sqlMCPLive = `SELECT EXISTS (SELECT 1 FROM app.sessions WHERE id = $1 AND user_id = $3 AND kind = 'mcp' AND revoked_at IS NULL ` +
		`AND idle_expires_at > $2 AND absolute_expires_at > $2 AND EXISTS (SELECT 1 FROM app.users u WHERE u.id = $3 AND u.status = 'active'))`
	// M1: the 'mcp' session (no issued bearer) and the pending connection, in the caller's TENANT-ORG transaction.
	sqlMCPInsertSession = `INSERT INTO app.sessions (user_id, kind, credential_hash, csrf_token, auth_time, created_at, renewed_at, ` +
		`idle_expires_at, absolute_expires_at) VALUES ($1, 'mcp', $2, $3, $4, $4, $4, $5, $6) RETURNING id::text`
	sqlMCPInsertConnection = `INSERT INTO app.mcp_connections (id, session_id, user_id, organization_id, host, resource, pending_expires_at, ` +
		`created_at, display_name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	// G063 rename (Plan r1 §4.2): the plain read of the connection's (immutable) 'mcp' session, then, after that session
	// is locked, the connection row lock (last) and the name-only write.
	sqlMCPConnectionSession = `SELECT session_id::text FROM app.mcp_connections WHERE id = $1 AND user_id = $2 AND organization_id = $3`
	sqlMCPNameLock          = `SELECT state, pending_expires_at FROM app.mcp_connections WHERE id = $1 AND user_id = $2 AND organization_id = $3 ` +
		`FOR NO KEY UPDATE`
	sqlMCPRename = `UPDATE app.mcp_connections SET display_name = $2 WHERE id = $1`
	// The plain row read and the exclusive row lock (the last lock of M2 and M3).
	mcpRowColumns  = `state, client_id, selected_group_id::text, context_version`
	sqlMCPRow      = `SELECT ` + mcpRowColumns + ` FROM app.mcp_connections WHERE id = $1 AND user_id = $2 AND organization_id = $3`
	sqlMCPRowLock  = sqlMCPRow + ` FOR NO KEY UPDATE`
	sqlMCPActivate = `UPDATE app.mcp_connections SET state = 'active', client_id = $2, activated_at = $3 ` +
		`WHERE id = $1 AND state = 'pending' AND pending_expires_at > $3`
	sqlMCPSelect = `UPDATE app.mcp_connections SET selected_group_id = $2::uuid, context_version = context_version + 1 ` +
		`WHERE id = $1 RETURNING context_version`
	// The setup listing (Duo follow-up correction): the owner's live connections in the Organization, newest first on
	// (created_at, id), one page of MCPListPage plus one look-ahead row; the next page resumes strictly after a keyset
	// position. Two constant statements share the unchanged owner, Organization, session and pending-expiry filters.
	sqlMCPListFrom = `SELECT c.id::text, c.resource, c.state, c.client_id, c.created_at, c.activated_at, c.display_name ` +
		`FROM app.mcp_connections c JOIN app.sessions s ON s.id = c.session_id ` +
		`WHERE c.user_id = $1 AND c.organization_id = $2 AND s.revoked_at IS NULL AND s.idle_expires_at > $3 AND s.absolute_expires_at > $3 ` +
		`AND (c.state = 'active' OR c.pending_expires_at > $3)`
	sqlMCPListOrder   = ` ORDER BY c.created_at DESC, c.id DESC LIMIT 101`
	sqlMCPList        = sqlMCPListFrom + sqlMCPListOrder
	sqlMCPListAfter   = sqlMCPListFrom + ` AND (c.created_at, c.id) < ($4::timestamptz, $5::uuid)` + sqlMCPListOrder
	sqlMCPDisplayName = `SELECT display_name FROM app.users WHERE id = $1`
)

// ErrMCPSelectionChanged reports that a concurrent selection committed between the plain read and the row lock of a
// selection write (Plan r1 §4.4): nothing was written; the client reconciles with get_current_context.
var ErrMCPSelectionChanged = errors.New("identity: mcp selection changed concurrently")

// readMCPBinding reads one connection with its session liveness and owner status in a TENANT-ORG transaction of userID
// in org. live reports the evaluated liveness at now; a missing row is (zero, false, nil).
func (s *Store) readMCPBinding(ctx context.Context, userID, org, id string, now time.Time) (MCPConnection, bool, error) {
	var (
		c                                 MCPConnection
		client, group                     *string
		kind, status                      string
		unrevoked                         bool
		idleExp, absoluteExp, renewed, pe time.Time
		found                             bool
	)
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initTenantOrg(ctx, tx, userID, org); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, sqlMCPBinding, id, userID, org).Scan(&c.id, &c.session, &c.user, &c.organization, &c.host, &c.resource,
			&c.state, &pe, &client, &group, &c.contextVersion, &kind, &unrevoked, &idleExp, &absoluteExp, &renewed, &status)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return unavailable(err)
		}
		found = true
		return nil
	})
	if err != nil || !found {
		return MCPConnection{}, false, err
	}
	if client != nil {
		c.clientID = *client
	}
	if group != nil {
		c.selectedGroup = *group
	}
	c.pendingExpires, c.renewedAt = pe, renewed
	live := kind == SessionKindMCP && unrevoked && idleExp.After(now) && absoluteExp.After(now) && status == "active" &&
		(c.state == MCPStateActive || (c.state == MCPStatePending && c.pendingExpires.After(now)))
	return c, live, nil
}

// renewMCPSession is M5 in an IDENTITY transaction of userID.
func (s *Store) renewMCPSession(ctx context.Context, userID, sessionID string, now time.Time, idle time.Duration) error {
	return s.inTx(ctx, func(tx pgx.Tx) error {
		if err := initIdentity(ctx, tx, userID); err != nil {
			return err
		}
		var exp time.Time
		err := tx.QueryRow(ctx, sqlMCPRenew, sessionID, now, now.Add(idle), userID, now.Add(-MCPRenewEvery)).Scan(&exp)
		switch {
		case err == nil:
			return nil
		case !errors.Is(err, pgx.ErrNoRows):
			return unavailable(err)
		}
		var live bool
		if err := tx.QueryRow(ctx, sqlMCPLive, sessionID, now, userID).Scan(&live); err != nil {
			return unavailable(err)
		}
		if !live {
			return ErrUnauthenticated
		}
		return nil
	})
}

// InsertMCPConnectionTx inserts the 'mcp' session and the pending connection (M1) in the access facade's TENANT-ORG
// transaction for the facade's resolved subject, after the facade locked the user, the browser session and the
// Organization binding. Row-level security requires a current Organization role for both rows again. displayName is the
// optional friendly name (G063), nil for an unnamed connection, already normalized by NormalizeMCPDisplayName.
func (s *Service) InsertMCPConnectionTx(ctx context.Context, tx pgx.Tx, userID, organizationID, connectionID, host, resource string, l MCPLifetimes,
	displayName *string) error {
	if !canonicalUserID(userID) || !canonicalUserID(organizationID) || !canonicalUserID(connectionID) || host == "" || resource == "" ||
		l.Pending <= 0 || l.Idle <= 0 || l.Absolute < l.Idle {
		return ErrInvalidRequest
	}
	if !validMCPName(displayName) {
		return ErrInvalidMCPName
	}
	discarded, err := newSecret()
	if err != nil {
		return ErrUnavailable
	}
	csrf, err := newSecret()
	if err != nil {
		return ErrUnavailable
	}
	now := s.now()
	var sessionID string
	if err := tx.QueryRow(ctx, sqlMCPInsertSession, userID, hashSecret(discarded), csrf, now, now.Add(l.Idle), now.Add(l.Absolute)).
		Scan(&sessionID); err != nil {
		return unavailable(err)
	}
	if _, err := tx.Exec(ctx, sqlMCPInsertConnection, connectionID, sessionID, userID, organizationID, host, resource, now.Add(l.Pending), now,
		displayName); err != nil {
		return unavailable(err)
	}
	return nil
}

// ReadMCPConnectionSessionTx is step 2 of the rename (G063 Plan r1 §4.2), in the access facade's TENANT-ORG write
// transaction after the user and the browser session were locked: the plain read of the 'mcp' session of userID's own
// connection in organizationID. The session is immutable (the guard), so the unlocked read cannot be stale for the
// target-session lock that follows. Unknown, foreign and malformed are unauthenticated (the facade's not_found).
func (s *Service) ReadMCPConnectionSessionTx(ctx context.Context, tx pgx.Tx, userID, organizationID, connectionID string) (string, error) {
	if !canonicalUserID(userID) || !canonicalUserID(organizationID) || !canonicalUserID(connectionID) {
		return "", ErrUnauthenticated
	}
	var sessionID string
	err := tx.QueryRow(ctx, sqlMCPConnectionSession, connectionID, userID, organizationID).Scan(&sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUnauthenticated
	}
	if err != nil {
		return "", unavailable(err)
	}
	return sessionID, nil
}

// RenameMCPConnectionTx is the last step of the rename (G063 Plan r1 §4.2), after the user, the browser session, the
// connection's 'mcp' session (live, FOR SHARE) and the Organization binding were locked: it locks the connection row
// (FOR NO KEY UPDATE, the last lock, as in M2/M3) and, only while the row exists and a pending setup has not expired at
// the service clock, writes display_name (nil clears) and nothing else. Every other column, the session and its expiries
// are untouched; the strict capture trigger records mcp.connection_renamed only for a distinct value. name must already
// be normalized (ErrInvalidMCPName otherwise).
func (s *Service) RenameMCPConnectionTx(ctx context.Context, tx pgx.Tx, userID, organizationID, connectionID string, name *string) error {
	if !canonicalUserID(userID) || !canonicalUserID(organizationID) || !canonicalUserID(connectionID) {
		return ErrUnauthenticated
	}
	if !validMCPName(name) {
		return ErrInvalidMCPName
	}
	var (
		state   string
		expires time.Time
	)
	err := tx.QueryRow(ctx, sqlMCPNameLock, connectionID, userID, organizationID).Scan(&state, &expires)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUnauthenticated
	}
	if err != nil {
		return unavailable(err)
	}
	if state != MCPStateActive && !(state == MCPStatePending && expires.After(s.now())) {
		return ErrUnauthenticated // an expired setup is never usable again: not one of the caller's live connections
	}
	tag, err := tx.Exec(ctx, sqlMCPRename, connectionID, name)
	if err != nil {
		return unavailable(err)
	}
	if tag.RowsAffected() != 1 {
		return ErrUnauthenticated
	}
	return nil
}

// mcpRow is the mutable part of a connection row.
type mcpRow struct {
	state, client, group string
	version              int64
}

func scanMCPRow(row pgx.Row) (mcpRow, error) {
	var r mcpRow
	var client, group *string
	if err := row.Scan(&r.state, &client, &group, &r.version); err != nil {
		return mcpRow{}, err
	}
	if client != nil {
		r.client = *client
	}
	if group != nil {
		r.group = *group
	}
	return r, nil
}

// ActivateMCPConnectionTx is the last step of M2 in the facade's TENANT-ORG write transaction (after the user, the
// connection's 'mcp' session and the Organization binding were locked): it locks the connection row and, only while it
// is still pending and unexpired at the service clock, activates it and binds clientID. Concurrent first tokens
// serialize on the row lock: exactly one activates; a later one observes the bound client.
func (s *Service) ActivateMCPConnectionTx(ctx context.Context, tx pgx.Tx, c MCPConnection, clientID string) (MCPActivation, error) {
	if !c.Valid() || clientID == "" || len(clientID) > 2048 {
		return MCPActivation{}, ErrUnauthenticated
	}
	r, err := scanMCPRow(tx.QueryRow(ctx, sqlMCPRowLock, c.id, c.user, c.organization))
	if errors.Is(err, pgx.ErrNoRows) {
		return MCPActivation{}, ErrUnauthenticated
	}
	if err != nil {
		return MCPActivation{}, unavailable(err)
	}
	if r.state != MCPStatePending {
		return MCPActivation{State: r.state, ClientID: r.client}, nil
	}
	tag, err := tx.Exec(ctx, sqlMCPActivate, c.id, clientID, s.now())
	if err != nil {
		return MCPActivation{}, unavailable(err)
	}
	if tag.RowsAffected() != 1 {
		return MCPActivation{State: MCPStatePending}, nil // expired setup: it stays pending until it is never usable
	}
	return MCPActivation{State: MCPStateActive, ClientID: clientID, Activated: true}, nil
}

// ReadMCPSelectionTx is the plain (unlocked) read of c's current selection and context version in the facade's
// transaction.
func (s *Service) ReadMCPSelectionTx(ctx context.Context, tx pgx.Tx, c MCPConnection) (group string, version int64, active bool, err error) {
	if !c.Valid() {
		return "", 0, false, ErrUnauthenticated
	}
	r, err := scanMCPRow(tx.QueryRow(ctx, sqlMCPRow, c.id, c.user, c.organization))
	if errors.Is(err, pgx.ErrNoRows) {
		return "", 0, false, ErrUnauthenticated
	}
	if err != nil {
		return "", 0, false, unavailable(err)
	}
	return r.group, r.version, r.state == MCPStateActive, nil
}

// SelectMCPGroupTx is the last step of M3 in the facade's guarded write (after the user, the session, the Organization,
// the Organization binding and the subject's bindings in the previous and target Groups and those rows were locked): it locks the connection row, writes
// nothing if the selection differs from expected (ErrMCPSelectionChanged), and otherwise stores group ("" clears) with
// the next context version. It runs in the facade's Organization-only write; the guard trigger requires that shape and
// a target (and previous selection) visible to the subject under row-level security.
func (s *Service) SelectMCPGroupTx(ctx context.Context, tx pgx.Tx, c MCPConnection, expected, group string) (int64, error) {
	if !c.Valid() {
		return 0, ErrUnauthenticated
	}
	r, err := scanMCPRow(tx.QueryRow(ctx, sqlMCPRowLock, c.id, c.user, c.organization))
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrUnauthenticated
	}
	if err != nil {
		return 0, unavailable(err)
	}
	if r.state != MCPStateActive || r.group != expected {
		return 0, ErrMCPSelectionChanged
	}
	var arg *string
	if group != "" {
		arg = &group
	}
	var v int64
	if err := tx.QueryRow(ctx, sqlMCPSelect, c.id, arg).Scan(&v); err != nil {
		return 0, unavailable(err)
	}
	return v, nil
}

// ListMCPConnectionsTx lists one page of the subject's own live connections in the context Organization (at most
// MCPListPage, newest first), after the untrusted keyset position when one is given, in the facade's TENANT-ORG
// transaction.
func (s *Service) ListMCPConnectionsTx(ctx context.Context, tx pgx.Tx, userID, organizationID string, after *MCPListPosition) (MCPConnectionPage, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if after == nil {
		rows, err = tx.Query(ctx, sqlMCPList, userID, organizationID, s.now())
	} else {
		if !canonicalUserID(after.ID) || after.CreatedAt.IsZero() {
			return MCPConnectionPage{}, ErrInvalidRequest
		}
		rows, err = tx.Query(ctx, sqlMCPListAfter, userID, organizationID, s.now(), after.CreatedAt, after.ID)
	}
	if err != nil {
		return MCPConnectionPage{}, unavailable(err)
	}
	defer rows.Close()
	out := []MCPConnectionInfo{}
	for rows.Next() {
		var i MCPConnectionInfo
		var client *string
		var activated *time.Time
		if err := rows.Scan(&i.ID, &i.Resource, &i.State, &client, &i.CreatedAt, &activated, &i.DisplayName); err != nil {
			return MCPConnectionPage{}, unavailable(err)
		}
		if client != nil {
			i.ClientID = *client
		}
		if activated != nil {
			i.ActivatedAt = *activated
		}
		out = append(out, i)
	}
	if err := rows.Err(); err != nil {
		return MCPConnectionPage{}, unavailable(err)
	}
	if len(out) > MCPListPage {
		return MCPConnectionPage{Connections: out[:MCPListPage], More: true}, nil
	}
	return MCPConnectionPage{Connections: out}, nil
}

// UserDisplayNameTx reads the subject's untrusted display name ("" when absent) in the facade's transaction.
func (s *Service) UserDisplayNameTx(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	var name *string
	err := tx.QueryRow(ctx, sqlMCPDisplayName, userID).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUnauthenticated
	}
	if err != nil {
		return "", unavailable(err)
	}
	if name == nil {
		return "", nil
	}
	return *name, nil
}
