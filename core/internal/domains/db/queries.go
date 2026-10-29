// Package db holds the G016 domain queries as handwritten pgx statements (Plan r4 §5.7 fallback). The one bounded sqlc
// trial was deferred: sqlc 1.29.0 cannot parse the authoritative Goose set alone, because the app and audit schemas are
// created by the role script (infra/postgres/init/10-roles.sh), not by a migration, so generation would need a second
// schema file (the recorded deferral criterion; exact error in the G016 Result). The names and shapes are the ones the
// generator would have produced, so the facade is unchanged.
//
// UUIDs travel as text and are cast in SQL (the repository convention). Every statement runs on the access facade's
// guarded transaction as the runtime role, under row-level security; the lock order user -> session -> Organization
// (facts) -> domain row is the facade's. Every update carries version + 1 and names the version it read (the optimistic
// guard: a raced change matches no row). Nothing here opens a transaction.
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX is the caller's transaction.
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// Queries runs the domain statements on the DBTX each method receives.
type Queries struct{}

// New returns the query set.
func New() *Queries { return &Queries{} }

const domainColumns = `id::text AS id, organization_id::text AS organization_id, hostname, generation, state, challenge_value,
          challenge_expires_at, verified_at, verified_until, last_check_at, last_check_outcome, check_failures, tls_state,
          tls_checked_at, tls_not_after, tls_reference, removed_at, created_at, version`

// GetDomainRow is one organization_domains row as the facade reads it.
type GetDomainRow struct {
	ID                 string
	OrganizationID     string
	Hostname           string
	Generation         int64
	State              string
	ChallengeValue     string
	ChallengeExpiresAt time.Time
	VerifiedAt         *time.Time
	VerifiedUntil      *time.Time
	LastCheckAt        *time.Time
	LastCheckOutcome   *string
	CheckFailures      int32
	TlsState           string
	TlsCheckedAt       *time.Time
	TlsNotAfter        *time.Time
	TlsReference       *string
	RemovedAt          *time.Time
	CreatedAt          time.Time
	Version            int64
}

// The per-statement row types share GetDomainRow's shape (convertible), as generated code would.
type (
	GetDomainForUpdateRow     GetDomainRow
	InsertDomainRow           GetDomainRow
	ListDomainsRow            GetDomainRow
	RotateChallengeRow        GetDomainRow
	RecordCheckRow            GetDomainRow
	RemoveDomainRow           GetDomainRow
	ReserveWithdrawalRetryRow GetDomainRow
	RecordWithdrawalRow       GetDomainRow
)

func scanDomain(row pgx.Row) (GetDomainRow, error) {
	var i GetDomainRow
	err := row.Scan(&i.ID, &i.OrganizationID, &i.Hostname, &i.Generation, &i.State, &i.ChallengeValue, &i.ChallengeExpiresAt,
		&i.VerifiedAt, &i.VerifiedUntil, &i.LastCheckAt, &i.LastCheckOutcome, &i.CheckFailures, &i.TlsState, &i.TlsCheckedAt,
		&i.TlsNotAfter, &i.TlsReference, &i.RemovedAt, &i.CreatedAt, &i.Version)
	return i, err
}

// InsertDomain: the generation is core's opaque CSPRNG claim fence (deviation #3 closure). Only its own (hostname,
// generation) collision is absorbed (no row: pgx.ErrNoRows, the caller regenerates once); a live-hostname conflict still
// raises 23505.
const insertDomain = `INSERT INTO app.organization_domains (id, organization_id, hostname, generation, challenge_value,
  challenge_expires_at, created_by)
VALUES ($1::text::uuid, $2::text::uuid, $3::text, $4::bigint, $5::text, $6::timestamptz, $7::text::uuid)
ON CONFLICT ON CONSTRAINT organization_domains_hostname_generation_key DO NOTHING
RETURNING ` + domainColumns

type InsertDomainParams struct {
	ID                 string
	OrganizationID     string
	Hostname           string
	Generation         int64
	ChallengeValue     string
	ChallengeExpiresAt time.Time
	CreatedBy          string
}

func (q *Queries) InsertDomain(ctx context.Context, db DBTX, arg InsertDomainParams) (InsertDomainRow, error) {
	i, err := scanDomain(db.QueryRow(ctx, insertDomain, arg.ID, arg.OrganizationID, arg.Hostname, arg.Generation,
		arg.ChallengeValue, arg.ChallengeExpiresAt, arg.CreatedBy))
	return InsertDomainRow(i), err
}

const getDomain = `SELECT ` + domainColumns + `
FROM app.organization_domains
WHERE id = $1::text::uuid AND organization_id = $2::text::uuid`

type GetDomainParams struct {
	ID             string
	OrganizationID string
}

func (q *Queries) GetDomain(ctx context.Context, db DBTX, arg GetDomainParams) (GetDomainRow, error) {
	return scanDomain(db.QueryRow(ctx, getDomain, arg.ID, arg.OrganizationID))
}

const getDomainForUpdate = getDomain + `
FOR NO KEY UPDATE`

type GetDomainForUpdateParams struct {
	ID             string
	OrganizationID string
}

func (q *Queries) GetDomainForUpdate(ctx context.Context, db DBTX, arg GetDomainForUpdateParams) (GetDomainForUpdateRow, error) {
	i, err := scanDomain(db.QueryRow(ctx, getDomainForUpdate, arg.ID, arg.OrganizationID))
	return GetDomainForUpdateRow(i), err
}

const listDomains = `SELECT ` + domainColumns + `
FROM app.organization_domains
WHERE organization_id = $1::text::uuid
ORDER BY created_at, id
LIMIT 500`

func (q *Queries) ListDomains(ctx context.Context, db DBTX, organizationID string) ([]ListDomainsRow, error) {
	rows, err := db.Query(ctx, listDomains, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListDomainsRow
	for rows.Next() {
		i, err := scanDomain(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, ListDomainsRow(i))
	}
	return items, rows.Err()
}

const rotateChallenge = `UPDATE app.organization_domains
SET challenge_value = $1::text, challenge_expires_at = $2::timestamptz, version = version + 1
WHERE id = $3::text::uuid AND version = $4::bigint
RETURNING ` + domainColumns

type RotateChallengeParams struct {
	ChallengeValue     string
	ChallengeExpiresAt time.Time
	ID                 string
	Version            int64
}

func (q *Queries) RotateChallenge(ctx context.Context, db DBTX, arg RotateChallengeParams) (RotateChallengeRow, error) {
	i, err := scanDomain(db.QueryRow(ctx, rotateChallenge, arg.ChallengeValue, arg.ChallengeExpiresAt, arg.ID, arg.Version))
	return RotateChallengeRow(i), err
}

const recordCheck = `UPDATE app.organization_domains
SET state = $1::text,
    verified_at = coalesce($2::timestamptz, verified_at),
    verified_until = CASE WHEN $3::boolean THEN now() ELSE coalesce($4::timestamptz, verified_until) END,
    last_check_at = $5::timestamptz,
    last_check_outcome = $6::text,
    check_failures = $7::integer,
    tls_state = coalesce($8::text, tls_state),
    tls_checked_at = coalesce($9::timestamptz, tls_checked_at),
    tls_not_after = coalesce($10::timestamptz, tls_not_after),
    tls_reference = coalesce($11::text, tls_reference),
    version = version + 1
WHERE id = $12::text::uuid AND version = $13::bigint
RETURNING ` + domainColumns

type RecordCheckParams struct {
	State            string
	VerifiedAt       *time.Time
	ProofLost        bool
	VerifiedUntil    *time.Time
	LastCheckAt      time.Time
	LastCheckOutcome string
	CheckFailures    int32
	TlsState         *string
	TlsCheckedAt     *time.Time
	TlsNotAfter      *time.Time
	TlsReference     *string
	ID               string
	Version          int64
}

func (q *Queries) RecordCheck(ctx context.Context, db DBTX, arg RecordCheckParams) (RecordCheckRow, error) {
	i, err := scanDomain(db.QueryRow(ctx, recordCheck, arg.State, arg.VerifiedAt, arg.ProofLost, arg.VerifiedUntil, arg.LastCheckAt,
		arg.LastCheckOutcome, arg.CheckFailures, arg.TlsState, arg.TlsCheckedAt, arg.TlsNotAfter, arg.TlsReference, arg.ID, arg.Version))
	return RecordCheckRow(i), err
}

const removeDomain = `UPDATE app.organization_domains
SET state = 'removed', removed_at = now(), tls_state = 'withdrawing', version = version + 1
WHERE id = $1::text::uuid AND version = $2::bigint
RETURNING ` + domainColumns

type RemoveDomainParams struct {
	ID      string
	Version int64
}

func (q *Queries) RemoveDomain(ctx context.Context, db DBTX, arg RemoveDomainParams) (RemoveDomainRow, error) {
	i, err := scanDomain(db.QueryRow(ctx, removeDomain, arg.ID, arg.Version))
	return RemoveDomainRow(i), err
}

const reserveWithdrawalRetry = `UPDATE app.organization_domains
SET check_failures = check_failures + 1, version = version + 1
WHERE id = $1::text::uuid AND version = $2::bigint
RETURNING ` + domainColumns

type ReserveWithdrawalRetryParams struct {
	ID      string
	Version int64
}

func (q *Queries) ReserveWithdrawalRetry(ctx context.Context, db DBTX, arg ReserveWithdrawalRetryParams) (ReserveWithdrawalRetryRow, error) {
	i, err := scanDomain(db.QueryRow(ctx, reserveWithdrawalRetry, arg.ID, arg.Version))
	return ReserveWithdrawalRetryRow(i), err
}

const recordWithdrawal = `UPDATE app.organization_domains
SET tls_state = $1::text, tls_checked_at = now(), last_check_at = now(), last_check_outcome = $2::text,
    check_failures = check_failures + $3::integer, version = version + 1
WHERE id = $4::text::uuid AND generation = $5::bigint AND version = $6::bigint
RETURNING ` + domainColumns

type RecordWithdrawalParams struct {
	TlsState         string
	LastCheckOutcome string
	FailureIncrement int32
	ID               string
	Generation       int64
	Version          int64
}

func (q *Queries) RecordWithdrawal(ctx context.Context, db DBTX, arg RecordWithdrawalParams) (RecordWithdrawalRow, error) {
	i, err := scanDomain(db.QueryRow(ctx, recordWithdrawal, arg.TlsState, arg.LastCheckOutcome, arg.FailureIncrement, arg.ID,
		arg.Generation, arg.Version))
	return RecordWithdrawalRow(i), err
}

const setOrganizationSlug = `UPDATE app.organizations SET slug = $1::text
WHERE id = $2::text::uuid
RETURNING slug`

type SetOrganizationSlugParams struct {
	Slug *string
	ID   string
}

func (q *Queries) SetOrganizationSlug(ctx context.Context, db DBTX, arg SetOrganizationSlugParams) (*string, error) {
	var slug *string
	err := db.QueryRow(ctx, setOrganizationSlug, arg.Slug, arg.ID).Scan(&slug)
	return slug, err
}

const getOrganizationSlug = `SELECT slug FROM app.organizations WHERE id = $1::text::uuid`

func (q *Queries) GetOrganizationSlug(ctx context.Context, db DBTX, id string) (*string, error) {
	var slug *string
	err := db.QueryRow(ctx, getOrganizationSlug, id).Scan(&slug)
	return slug, err
}

// hostnameProven (G016 security amendment, M-1): whether a proven (verified, active or suspended) row of the hostname
// exists in any Organization, through the definer reader (runtime has EXECUTE; it ignores ambient settings and reads
// proven rows only). Pending claims are never disclosed.
const hostnameProven = `SELECT count(*) > 0 FROM app.hostname_state('custom', $1::text)`

func (q *Queries) HostnameProven(ctx context.Context, db DBTX, hostname string) (bool, error) {
	var proven bool
	err := db.QueryRow(ctx, hostnameProven, hostname).Scan(&proven)
	return proven, err
}
