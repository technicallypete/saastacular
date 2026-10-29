package audit

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// The statements of the package: the one append (ten of the 12 descriptive columns runtime may insert; the two role
// columns are written only by the membership capture trigger), the tenant read and the revocation-gap marker read.
const (
	sqlAppend = `INSERT INTO audit.events (action, outcome, reason, operation, resource_type, resource_id, subject_user_id, ` +
		`organization_id, group_id, ttl_seconds) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	// The tenant read is sargable for a backward range scan of events_tenant_time_idx: the window lower bound, the
	// exclusive (time, id) upper bound (the cursor, or the window end with the zero UUID), newest first.
	sqlReadTenant = `SELECT id, occurred_at, action, outcome, coalesce(reason, ''), coalesce(operation, ''), actor_kind, ` +
		`actor_user_id, organization_id, group_id, coalesce(resource_type, ''), resource_id, subject_user_id, coalesce(role, ''), ` +
		`coalesce(previous_role, ''), coalesce(ttl_seconds, 0), coalesce(request_id, ''), coalesce(trace_id, '') ` +
		`FROM audit.events WHERE organization_id = $1 AND occurred_at >= $2 AND (occurred_at, id) < ($3, $4) ` +
		`ORDER BY occurred_at DESC, id DESC LIMIT $5`
	// The transaction-local revocation gap marker set by a capture trigger's exception handler (closed values
	// session.revoked and user.disabled, comma-separated); empty when no audit insert of a revocation was lost.
	sqlRevocationGap = `SELECT coalesce(pg_catalog.current_setting('audit.revocation_gap', true), '')`
)

// The append errors. Neither carries driver text.
var (
	// ErrRefused: the database or Append refused the event (privilege, source, scope, Group, shape or constraint).
	ErrRefused = errors.New("audit: event refused")
	// ErrUnavailable: the statement failed for any other reason (connection, cancellation, deadline, abort).
	ErrUnavailable = errors.New("audit: unavailable")
)

func uuidArg(u UUID) pgtype.UUID { return pgtype.UUID{Bytes: u, Valid: u != UUID{}} }

func textArg[T ~string](s T) *string {
	if s == "" {
		return nil
	}
	v := string(s)
	return &v
}

// Append records one core-owned event with one INSERT on the caller's transaction (never its own connection). A
// trigger-owned action is refused before anything is sent.
func Append(ctx context.Context, tx pgx.Tx, e Event) error {
	if TriggerOwned(e.Action) {
		return ErrRefused
	}
	var ttl *int
	if e.TTLSeconds != 0 {
		ttl = &e.TTLSeconds
	}
	_, err := tx.Exec(ctx, sqlAppend, string(e.Action), string(e.Outcome), textArg(e.Reason), textArg(e.Operation),
		textArg(e.ResourceType), uuidArg(e.ResourceID), uuidArg(e.Subject), uuidArg(e.Organization), uuidArg(e.Group), ttl)
	return classify(err)
}

// classify maps a driver error onto the two append errors: a privilege (42501) or integrity (class 23) refusal is
// ErrRefused, anything else ErrUnavailable.
func classify(err error) error {
	if err == nil {
		return nil
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) && (pg.Code == "42501" || (len(pg.Code) == 5 && pg.Code[:2] == "23")) {
		return ErrRefused
	}
	return ErrUnavailable
}

// Record is one stored event as the tenant reader returns it. Zero values mean NULL.
type Record struct {
	ID           UUID
	OccurredAt   time.Time
	Action       Action
	Outcome      Outcome
	Reason       Reason
	Operation    Operation
	ActorKind    string
	ActorUserID  UUID
	Organization UUID
	Group        UUID
	ResourceType ResourceType
	ResourceID   UUID
	Subject      UUID
	Role         string
	PreviousRole string
	TTLSeconds   int
	RequestID    string
	TraceID      string
}

// Cursor sentinels: no lower bound is -infinity; no cursor and no window end is (infinity, zero UUID).
func lowerBound(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{InfinityModifier: pgtype.NegativeInfinity, Valid: true}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func upperBound(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{InfinityModifier: pgtype.Infinity, Valid: true}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// ReadTenant returns at most limit events of org in one statement on the caller's transaction (row-level security
// filters them further): occurred_at at or after since (zero: no bound) and (occurred_at, id) strictly before
// (before, beforeID) (zero time: no bound), newest first.
func ReadTenant(ctx context.Context, tx pgx.Tx, org UUID, since, before time.Time, beforeID UUID, limit int) ([]Record, error) {
	rows, err := tx.Query(ctx, sqlReadTenant, pgtype.UUID{Bytes: org, Valid: true}, lowerBound(since), upperBound(before),
		pgtype.UUID{Bytes: beforeID, Valid: true}, limit)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer rows.Close()
	var out []Record
	for rows.Next() {
		var r Record
		var id, actor, o, g, res, subj pgtype.UUID
		var action, outcome, reason, op, rtype string
		if err := rows.Scan(&id, &r.OccurredAt, &action, &outcome, &reason, &op, &r.ActorKind, &actor, &o, &g, &rtype, &res, &subj,
			&r.Role, &r.PreviousRole, &r.TTLSeconds, &r.RequestID, &r.TraceID); err != nil {
			return nil, ErrUnavailable
		}
		r.ID, r.ActorUserID, r.Organization, r.Group, r.ResourceID, r.Subject = id.Bytes, actor.Bytes, o.Bytes, g.Bytes, res.Bytes, subj.Bytes
		r.Action, r.Outcome, r.Reason, r.Operation, r.ResourceType = Action(action), Outcome(outcome), Reason(reason), Operation(op), ResourceType(rtype)
		out = append(out, r)
	}
	if rows.Err() != nil {
		return nil, ErrUnavailable
	}
	return out, nil
}

// The revocation gap markers a capture trigger records.
const (
	GapSessionRevoked = "session.revoked"
	GapUserDisabled   = "user.disabled"
)

// RevocationGap reads the transaction-local gap marker on tx: the closed markers whose revocation insert was lost to an
// audit-only fault in this transaction (empty when none). A value outside the closed set is ErrUnavailable.
func RevocationGap(ctx context.Context, tx pgx.Tx) ([]string, error) {
	var v string
	if err := tx.QueryRow(ctx, sqlRevocationGap).Scan(&v); err != nil {
		return nil, ErrUnavailable
	}
	if v == "" {
		return nil, nil
	}
	var out []string
	start := 0
	for i := 0; i <= len(v); i++ {
		if i < len(v) && v[i] != ',' {
			continue
		}
		m := v[start:i]
		if m != GapSessionRevoked && m != GapUserDisabled {
			return nil, ErrUnavailable
		}
		out = append(out, m)
		start = i + 1
	}
	return out, nil
}
