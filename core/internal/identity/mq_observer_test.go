package identity_test

// The least-privilege observer helpers of the manual qualification test (manual_qualification_test.go, build tag
// workos_manual; G057 Plan r5 §3.7, D-14/D-16). They are untagged so that their pure rules run in every core check
// (go test ./internal/identity -run '^TestMQ') and so that mq_observer_integration_test.go (tag integration) executes
// the same reading path against a real database under row-level security.
//
// Every database read of a G057 step is one READ ONLY transaction as <db>_auth_reader (SET LOCAL ROLE from the
// migrator login, the SET chain every migration already uses) executing a member of the operational statement
// registry mqOperationalSQL through mqRead. The registry names only columns that role holds: created_at,
// revoke_reason and provider_revocation are not granted and are never read; the credential columns are grantable but
// excluded (TestMQSessionColumnsAreSanitized). Sessions are discovered by baseline set difference on granted columns,
// never by a timestamp look-back: a pre-existing session is in the baseline, a merely renewed session keeps its id and
// stays in the baseline (reported as renewed, never as new); only an INSERT (a completed callback) produces a new id.
// The two origins known before any real flow (P, the platform root; T, the custom host) come from MQ_ORIGINS; the
// tenant roles A/B (and C) are bound from rows discovered after a baseline taken before the first identified sign-in.

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/hostname"
	"core/internal/migrate"
)

// mqSessionColumns is the observer's complete read set on app.sessions: every member is in the <db>_auth_reader
// column grant (00006 + kind 00010 + origin/organization_id 00012). revoked_at and provider_logout_returned_at are
// read as IS NOT NULL flags only.
var mqSessionColumns = []string{"id", "user_id", "kind", "origin", "organization_id", "auth_time", "renewed_at",
	"revoked_at", "provider_logout_returned_at", "mfa_assurance"}

// The operational statements. mqRead executes nothing else.
const (
	mqSQLGuard    = `SELECT current_user::text, current_setting('transaction_read_only')`
	mqSQLSessions = `SELECT id::text, user_id::text, kind, coalesce(origin, ''), coalesce(organization_id::text, ''), auth_time, renewed_at,
  revoked_at IS NOT NULL, provider_logout_returned_at IS NOT NULL, mfa_assurance FROM app.sessions ORDER BY auth_time, id`
	mqSQLOrganizations = `SELECT id::text, coalesce(slug, ''), coalesce(platform_label, '') FROM app.organizations ORDER BY id`
	mqSQLRevocation    = `SELECT revoked_at IS NOT NULL, provider_logout_returned_at IS NOT NULL FROM app.sessions WHERE id = $1::text::uuid`
	mqSQLLatestAuth    = `SELECT max(auth_time) FROM app.sessions WHERE user_id = $1::text::uuid`
	mqSQLLoginFlags    = `SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = session_user`
	mqSQLRoleFlags     = `SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = $1`
	mqSQLForcedRLS     = `SELECT count(*)::int FROM pg_class WHERE oid IN ('app.users'::regclass, 'app.sessions'::regclass)
  AND relrowsecurity AND relforcerowsecurity`
	mqSQLReaderPolicies = `SELECT count(*)::int FROM pg_policy WHERE polname IN ('sessions_auth_reader', 'users_auth_reader')
  AND $1::text::regrole = ANY (polroles)`
)

// mqOperationalSQL is the registry of every statement a G057 step executes (Plan r5 §3.7, r4): the session select
// behind mqBaseline/mqNewSessions/mqWaitNewSession, the Organization select behind the O0/Q-4 discovery, the
// prior-session revocation probe, the kept latest-auth_time select, and the start-of-run preconditions.
var mqOperationalSQL = map[string]string{
	"guard":           mqSQLGuard,
	"sessions":        mqSQLSessions,
	"organizations":   mqSQLOrganizations,
	"revocation":      mqSQLRevocation,
	"latest_auth":     mqSQLLatestAuth,
	"login_flags":     mqSQLLoginFlags,
	"role_flags":      mqSQLRoleFlags,
	"forced_rls":      mqSQLForcedRLS,
	"reader_policies": mqSQLReaderPolicies,
}

var errMQNotOperational = errors.New("mq observer: the statement is not a member of mqOperationalSQL")

func mqOperational(sql string) bool {
	for _, s := range mqOperationalSQL {
		if s == sql {
			return true
		}
	}
	return false
}

// mqTx is the querier mqRead hands to its callback: it refuses any statement outside the registry before touching the
// transaction.
type mqTx struct{ tx pgx.Tx }

func (m mqTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if !mqOperational(sql) {
		return nil, errMQNotOperational
	}
	return m.tx.Query(ctx, sql, args...)
}

func (m mqTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if !mqOperational(sql) {
		return mqRefusedRow{}
	}
	return m.tx.QueryRow(ctx, sql, args...)
}

type mqRefusedRow struct{}

func (mqRefusedRow) Scan(...any) error { return errMQNotOperational }

// mqRead runs fn inside one `BEGIN READ ONLY; SET LOCAL ROLE <reader>; …; COMMIT` transaction on pool (a migrator
// login). It fails closed unless current_user inside the transaction is the reader and the transaction is read-only.
func mqRead(ctx context.Context, pool *pgxpool.Pool, reader string, fn func(context.Context, mqTx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return fmt.Errorf("mq observer: begin read only: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE "+pgx.Identifier{reader}.Sanitize()); err != nil {
		return fmt.Errorf("mq observer: set local role %s: %w", reader, err)
	}
	var current, readOnly string
	if err := tx.QueryRow(ctx, mqSQLGuard).Scan(&current, &readOnly); err != nil {
		return fmt.Errorf("mq observer: transaction guard: %w", err)
	}
	if current != reader || readOnly != "on" {
		return fmt.Errorf("mq observer: the transaction runs as %q with transaction_read_only %q, want %q and on", current, readOnly, reader)
	}
	if err := fn(ctx, mqTx{tx: tx}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// mqMigratorPool opens the observer's login: the migrator role, which may SET ROLE to the owner and, through it, to
// the auth reader (infra/postgres/provision-roles.sql). DB_RUNTIME_PASSWORD is not needed.
func mqMigratorPool(ctx context.Context, cfg migrate.Config) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, err
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = cfg.Host, uint16(cfg.Port), cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = cfg.MigratorRole(), cfg.Password
	pc.ConnConfig.ConnectTimeout = 3 * time.Second
	pc.MaxConns = 4
	return pgxpool.NewWithConfig(ctx, pc)
}

// mqPreflight asserts the start-of-run preconditions once, fail closed: neither the login nor the reader is a superuser
// or bypasses row-level security; app.users and app.sessions FORCE row-level security (the inverse of the legacy
// observer's gate: RLS must be on) and the two auth-reader SELECT policies name the reader.
func mqPreflight(ctx context.Context, pool *pgxpool.Pool, reader string) error {
	return mqRead(ctx, pool, reader, func(ctx context.Context, q mqTx) error {
		var super, bypass bool
		if err := q.QueryRow(ctx, mqSQLLoginFlags).Scan(&super, &bypass); err != nil {
			return fmt.Errorf("login role flags: %w", err)
		}
		if super || bypass {
			return errors.New("the login role is a superuser or bypasses row-level security")
		}
		if err := q.QueryRow(ctx, mqSQLRoleFlags, reader).Scan(&super, &bypass); err != nil {
			return fmt.Errorf("reader role flags: %w", err)
		}
		if super || bypass {
			return errors.New("the reader role is a superuser or bypasses row-level security")
		}
		var forced, policies int
		if err := q.QueryRow(ctx, mqSQLForcedRLS).Scan(&forced); err != nil {
			return fmt.Errorf("forced row-level security: %w", err)
		}
		if forced != 2 {
			return fmt.Errorf("app.users/app.sessions with FORCE row-level security: %d, want 2", forced)
		}
		if err := q.QueryRow(ctx, mqSQLReaderPolicies, reader).Scan(&policies); err != nil {
			return fmt.Errorf("reader policies: %w", err)
		}
		if policies != 2 {
			return fmt.Errorf("auth-reader SELECT policies naming %s: %d, want 2", reader, policies)
		}
		return nil
	})
}

// mqSessionState is one sanitized session row: granted columns only, flags in place of the revocation timestamps.
// Identifiers stay in memory and are never written to the result file.
type mqSessionState struct {
	ID, UserID, Kind, Origin, OrganizationID string
	AuthTime, RenewedAt                      time.Time
	Revoked, ProviderLogoutReturned          bool
	Assurance                                string
}

func mqScanSessions(ctx context.Context, q mqTx) ([]mqSessionState, error) {
	rows, err := q.Query(ctx, mqSQLSessions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []mqSessionState
	for rows.Next() {
		var s mqSessionState
		if err := rows.Scan(&s.ID, &s.UserID, &s.Kind, &s.Origin, &s.OrganizationID, &s.AuthTime, &s.RenewedAt, &s.Revoked,
			&s.ProviderLogoutReturned, &s.Assurance); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// mqBaseline is the set of every session id readable under the reader (the auth-reader policy admits all rows), keyed
// by id with the state kept for the renewed/revoked comparisons.
func mqBaseline(ctx context.Context, pool *pgxpool.Pool, reader string) (map[string]mqSessionState, error) {
	base := map[string]mqSessionState{}
	err := mqRead(ctx, pool, reader, func(ctx context.Context, q mqTx) error {
		rows, err := mqScanSessions(ctx, q)
		for _, s := range rows {
			base[s.ID] = s
		}
		return err
	})
	return base, err
}

// mqDelta is one set difference against a baseline.
type mqDelta struct {
	New     []mqSessionState // ids absent from the baseline, ordered by auth_time then id
	Renewed int              // baseline ids whose renewed_at advanced: never counted as new
}

// mqDiff is the pure set difference: a row is new only when its id is absent from the baseline. A baseline id whose
// renewed_at advanced is counted as renewed; a baseline id whose revoked flag flipped is neither.
func mqDiff(baseline map[string]mqSessionState, rows []mqSessionState) mqDelta {
	var d mqDelta
	for _, s := range rows {
		b, known := baseline[s.ID]
		switch {
		case !known:
			d.New = append(d.New, s)
		case s.RenewedAt.After(b.RenewedAt):
			d.Renewed++
		}
	}
	sort.SliceStable(d.New, func(i, j int) bool {
		if !d.New[i].AuthTime.Equal(d.New[j].AuthTime) {
			return d.New[i].AuthTime.Before(d.New[j].AuthTime)
		}
		return d.New[i].ID < d.New[j].ID
	})
	return d
}

// mqNewSessions reads the current rows under the reader and returns their set difference against baseline.
func mqNewSessions(ctx context.Context, pool *pgxpool.Pool, reader string, baseline map[string]mqSessionState) (mqDelta, error) {
	var d mqDelta
	err := mqRead(ctx, pool, reader, func(ctx context.Context, q mqTx) error {
		rows, err := mqScanSessions(ctx, q)
		if err != nil {
			return err
		}
		d = mqDiff(baseline, rows)
		return nil
	})
	return d, err
}

// mqWaitNewSession polls mqNewSessions until a new row satisfies pred or limit elapses. It returns the first matching
// row and the delta of the poll that found it.
func mqWaitNewSession(ctx context.Context, pool *pgxpool.Pool, reader string, baseline map[string]mqSessionState,
	pred func(mqSessionState) bool, limit, poll time.Duration) (mqSessionState, mqDelta, bool, error) {
	deadline := time.Now().Add(limit)
	for {
		d, err := mqNewSessions(ctx, pool, reader, baseline)
		if err != nil {
			return mqSessionState{}, d, false, err
		}
		for _, s := range d.New {
			if pred(s) {
				return s, d, true, nil
			}
		}
		if !time.Now().Before(deadline) {
			return mqSessionState{}, d, false, nil
		}
		time.Sleep(poll)
	}
}

// mqRevocation reads the two revocation flags of one session id (the allowed-column form of the local-revocation
// probe: revoke_reason is never read).
func mqRevocation(ctx context.Context, pool *pgxpool.Pool, reader, id string) (revoked, returned bool, err error) {
	err = mqRead(ctx, pool, reader, func(ctx context.Context, q mqTx) error {
		return q.QueryRow(ctx, mqSQLRevocation, id).Scan(&revoked, &returned)
	})
	return revoked, returned, err
}

// mqWaitRevoked polls mqRevocation until the session shows revoked_at IS NOT NULL or limit elapses; it returns the
// provider_logout_returned_at flag observed with the revocation.
func mqWaitRevoked(ctx context.Context, pool *pgxpool.Pool, reader, id string, limit, poll time.Duration) (returned, ok bool, err error) {
	deadline := time.Now().Add(limit)
	for {
		revoked, ret, err := mqRevocation(ctx, pool, reader, id)
		if err != nil {
			return false, false, err
		}
		if revoked {
			return ret, true, nil
		}
		if !time.Now().Before(deadline) {
			return false, false, nil
		}
		time.Sleep(poll)
	}
}

// mqLatestAuth is the latest signed authentication time over all of a user's sessions (zero when there is none).
func mqLatestAuth(ctx context.Context, pool *pgxpool.Pool, reader, userID string) (time.Time, error) {
	var latest *time.Time
	err := mqRead(ctx, pool, reader, func(ctx context.Context, q mqTx) error {
		return q.QueryRow(ctx, mqSQLLatestAuth, userID).Scan(&latest)
	})
	if err != nil || latest == nil {
		return time.Time{}, err
	}
	return *latest, nil
}

// mqOrg is one Organization row as the reader sees it (id, slug, platform_label are its only granted columns).
type mqOrg struct{ ID, Slug, PlatformLabel string }

func mqOrganizations(ctx context.Context, pool *pgxpool.Pool, reader string) ([]mqOrg, error) {
	var out []mqOrg
	err := mqRead(ctx, pool, reader, func(ctx context.Context, q mqTx) error {
		rows, err := q.Query(ctx, mqSQLOrganizations)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var o mqOrg
			if err := rows.Scan(&o.ID, &o.Slug, &o.PlatformLabel); err != nil {
				return err
			}
			out = append(out, o)
		}
		return rows.Err()
	})
	return out, err
}

// mqOrgIDs is the Organization-ID baseline (O0) of a row set.
func mqOrgIDs(orgs []mqOrg) map[string]bool {
	ids := map[string]bool{}
	for _, o := range orgs {
		ids[o.ID] = true
	}
	return ids
}

// mqOrigin is one configured origin by role.
type mqOrigin struct {
	Role, Origin, Host, Port string
}

func (o mqOrigin) hostPort() string {
	if o.Port == "" {
		return o.Host
	}
	return o.Host + ":" + o.Port
}

// labelOrigin is the exact origin of a platform label host under this origin's host and port.
func (o mqOrigin) labelOrigin(label string) string { return "https://" + label + "." + o.hostPort() }

// mqOrigins are the two origins known before any real flow: P (the platform root) and T (the custom host).
type mqOrigins struct{ P, T mqOrigin }

// mqHostName reports whether s is a lowercase DNS name (one or more labels of [a-z0-9-], no leading or trailing
// hyphen, each at most 63 bytes, at most 253 in all).
func mqHostName(s string) bool {
	if s == "" || len(s) > 253 {
		return false
	}
	for _, l := range strings.Split(s, ".") {
		if l == "" || len(l) > 63 || l[0] == '-' || l[len(l)-1] == '-' || strings.Trim(l, "abcdefghijklmnopqrstuvwxyz0123456789-") != "" {
			return false
		}
	}
	return true
}

// mqParseOrigins parses MQ_ORIGINS (`P=https://B[:port],T=https://T`): exact https origins, no userinfo, path, query or
// fragment; a port only when platformPort (PLATFORM_HOST_PORT) is set, and then exactly that port on every origin; the
// roles P and T mandatory, each once, with distinct origins. The tenant roles A and B are refused: they are bound from
// the database by discovery, never taken from the environment.
func mqParseOrigins(spec, platformPort string) (mqOrigins, error) {
	var out mqOrigins
	seen := map[string]bool{}
	for _, item := range strings.Split(spec, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		role, raw, found := strings.Cut(item, "=")
		if !found {
			return out, fmt.Errorf("MQ_ORIGINS: entry without a role")
		}
		switch role {
		case "P", "T":
		case "A", "B", "C":
			return out, fmt.Errorf("MQ_ORIGINS: role %s is bound by discovery and must not be configured", role)
		default:
			return out, fmt.Errorf("MQ_ORIGINS: unknown role %q", role)
		}
		if seen[role] {
			return out, fmt.Errorf("MQ_ORIGINS: role %s given twice", role)
		}
		seen[role] = true
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.User != nil || u.Opaque != "" || u.Path != "" || u.RawQuery != "" ||
			u.Fragment != "" || raw != "https://"+u.Host || !mqHostName(u.Hostname()) || strings.HasSuffix(u.Host, ":") {
			return out, fmt.Errorf("MQ_ORIGINS: role %s is not an exact lowercase https origin", role)
		}
		if u.Port() != platformPort {
			if platformPort == "" {
				return out, fmt.Errorf("MQ_ORIGINS: role %s carries a port but PLATFORM_HOST_PORT is not set", role)
			}
			return out, fmt.Errorf("MQ_ORIGINS: role %s must carry the configured PLATFORM_HOST_PORT", role)
		}
		o := mqOrigin{Role: role, Origin: raw, Host: u.Hostname(), Port: u.Port()}
		if role == "P" {
			out.P = o
		} else {
			out.T = o
		}
	}
	if !seen["P"] || !seen["T"] {
		return out, errors.New("MQ_ORIGINS: roles P and T are both required")
	}
	if out.P.Origin == out.T.Origin {
		return out, errors.New("MQ_ORIGINS: P and T must be distinct origins")
	}
	return out, nil
}

// mqTenant is a tenant role bound by discovery.
type mqTenant struct {
	Role, Origin, Label, OrganizationID, SessionID, UserID string
}

// mqBindTenant binds role to the one new session of a step (Plan r5 §3.7 Q-A binding; the same match Q-4 uses for C).
// It is pure: it consumes rows the operational helpers already read. It binds only when all hold: exactly one new
// row; a browser session; its origin is https://<label>.<host of P>[:port of P] with one label of the platform
// grammar; its Organization is absent from the step's Organization-ID baseline o0 and names a readable Organization
// whose platform_label equals that label; and, for a later role, its origin and Organization differ from every
// already bound tenant's.
func mqBindTenant(role string, fresh []mqSessionState, orgs []mqOrg, o0 map[string]bool, p mqOrigin, bound []mqTenant) (mqTenant, error) {
	if role != "A" && role != "B" && role != "C" {
		return mqTenant{}, fmt.Errorf("role %q is not a tenant role", role)
	}
	if len(fresh) != 1 {
		return mqTenant{}, fmt.Errorf("%d new sessions since the step baseline, want exactly 1", len(fresh))
	}
	row := fresh[0]
	if row.Kind != "browser" {
		return mqTenant{}, errors.New("the new session is not a browser session")
	}
	host, ok := strings.CutPrefix(row.Origin, "https://")
	label, under, found := strings.Cut(host, ".")
	if !ok || !found || under != p.hostPort() || !hostname.SlugGrammar(label) || row.Origin != p.labelOrigin(label) {
		return mqTenant{}, errors.New("the new session's origin is not a single-label platform host under P")
	}
	if row.OrganizationID == "" || o0[row.OrganizationID] {
		return mqTenant{}, errors.New("the new session's Organization is not new since the step baseline")
	}
	var org *mqOrg
	for i := range orgs {
		if orgs[i].ID == row.OrganizationID {
			org = &orgs[i]
		}
	}
	if org == nil {
		return mqTenant{}, errors.New("the new session's Organization is not readable")
	}
	if org.PlatformLabel != label {
		return mqTenant{}, errors.New("the Organization's platform_label differs from the origin label")
	}
	for _, b := range bound {
		if b.Role == role || b.Origin == row.Origin || b.OrganizationID == row.OrganizationID {
			return mqTenant{}, fmt.Errorf("the new session repeats the origin or Organization bound as %s", b.Role)
		}
	}
	return mqTenant{Role: role, Origin: row.Origin, Label: label, OrganizationID: row.OrganizationID, SessionID: row.ID, UserID: row.UserID}, nil
}

// mqSSOObservation is the boolean form of mqSSOInput (mq_eval_test.go) for the G057 roles: ordering is proven by the
// post-revocation re-baseline (a session absent from a baseline taken after the revocation was inserted after it),
// not by a created_at comparison, which the reader cannot make.
type mqSSOObservation struct {
	Attestation        string    // "no-prompt" or "prompt"
	SameUser           bool      // the new session maps to the same canonical user
	NewAfterRevocation bool      // the new id appeared after the prior session showed revoked_at IS NOT NULL
	NewAuth            time.Time // its signed auth_time
	PriorLatestAuth    time.Time // the latest signed authentication before the new session
}

// mqSSOReuseObserved is mqSSOReuse over the boolean form: unprompted, same user, new after the local revocation, prior
// signed authentication retained. A prompted sign-in is an ordinary sign-in and never counts as reuse.
func mqSSOReuseObserved(in mqSSOObservation) (bool, string) {
	switch {
	case in.Attestation != "no-prompt":
		return false, "a credential prompt was attested: an ordinary sign-in, not reuse"
	case !in.SameUser:
		return false, "the new session belongs to another user"
	case !in.NewAfterRevocation:
		return false, "no new application session after the local revocation"
	case !in.NewAuth.Equal(in.PriorLatestAuth):
		return false, "the signed authentication time was not retained"
	}
	return true, ""
}

// --- unit tests (go test ./internal/identity -run '^TestMQ') ---

func TestMQOriginsFromEnv(t *testing.T) {
	good, err := mqParseOrigins("P=https://platform.example,T=https://portal.customer.example", "")
	if err != nil || good.P.Origin != "https://platform.example" || good.T.Host != "portal.customer.example" || good.P.Port != "" {
		t.Fatalf("the P/T form was rejected: %+v %v", good, err)
	}
	if good.P.labelOrigin("abc") != "https://abc.platform.example" {
		t.Fatalf("label origin %q", good.P.labelOrigin("abc"))
	}
	withPort, err := mqParseOrigins(" T=https://portal.acme.test:543 , P=https://localhost:543 ", "543")
	if err != nil || withPort.P.Origin != "https://localhost:543" || withPort.P.labelOrigin("org-a") != "https://org-a.localhost:543" {
		t.Fatalf("the configured-port form was rejected: %+v %v", withPort, err)
	}
	for name, c := range map[string]struct{ spec, port string }{
		"http":                 {"P=http://platform.example,T=https://portal.customer.example", ""},
		"path":                 {"P=https://platform.example/,T=https://portal.customer.example", ""},
		"query":                {"P=https://platform.example?x=1,T=https://portal.customer.example", ""},
		"fragment":             {"P=https://platform.example#f,T=https://portal.customer.example", ""},
		"userinfo":             {"P=https://u@platform.example,T=https://portal.customer.example", ""},
		"upper case":           {"P=https://Platform.example,T=https://portal.customer.example", ""},
		"duplicate role":       {"P=https://platform.example,P=https://other.example,T=https://portal.customer.example", ""},
		"missing P":            {"T=https://portal.customer.example", ""},
		"missing T":            {"P=https://platform.example", ""},
		"A configured":         {"P=https://platform.example,T=https://portal.customer.example,A=https://la.platform.example", ""},
		"B configured":         {"P=https://platform.example,T=https://portal.customer.example,B=https://lb.platform.example", ""},
		"unknown role":         {"P=https://platform.example,T=https://portal.customer.example,X=https://x.example", ""},
		"no role":              {"https://platform.example,T=https://portal.customer.example", ""},
		"port without config":  {"P=https://platform.example:543,T=https://portal.customer.example", ""},
		"port other than conf": {"P=https://platform.example:8443,T=https://portal.customer.example:8443", "543"},
		"port missing on T":    {"P=https://localhost:543,T=https://portal.acme.test", "543"},
		"identical origins":    {"P=https://platform.example,T=https://platform.example", ""},
		"empty":                {"", ""},
		"trailing colon":       {"P=https://platform.example:,T=https://portal.customer.example", ""},
	} {
		if _, err := mqParseOrigins(c.spec, c.port); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

// TestMQSessionColumnsAreSanitized pins the registry (and only the registry, r4): the session read set is a subset of
// the auth-reader column grant, disjoint from the credential columns and without created_at; every operational
// statement names no column of app.sessions outside that set (auth_time and user_id included in it), none of the
// forbidden identifiers, nothing of app.organizations beyond id/slug/platform_label; the registry is non-empty; and
// mqRead's querier refuses a non-member before any SQL. The negative privilege probes (integration, mqProbe) and the
// emulator-era runtime bodies are outside this scope by construction.
func TestMQSessionColumnsAreSanitized(t *testing.T) {
	arSessions := []string{"id", "user_id", "credential_hash", "csrf_token", "provider_session_id", "mfa_assurance", "auth_time",
		"idle_expires_at", "absolute_expires_at", "renewed_at", "revoked_at", "provider_logout_ref_hash", "provider_logout_ref_expires_at",
		"provider_logout_returned_at", "kind", "origin", "organization_id"}
	forbidden := []string{"created_at", "revoke_reason", "provider_revocation", "credential_hash", "csrf_token", "provider_session_id",
		"provider_logout_ref_hash"}
	allSessions := append(append([]string{}, arSessions...), "created_at", "revoke_reason", "provider_revocation")
	set := func(xs []string) map[string]bool {
		m := map[string]bool{}
		for _, x := range xs {
			m[x] = true
		}
		return m
	}
	granted, banned, columns, allowed := set(arSessions), set(forbidden), set(allSessions), set(mqSessionColumns)
	for _, c := range mqSessionColumns {
		if !granted[c] {
			t.Errorf("%s is not in the auth-reader column grant", c)
		}
		if banned[c] {
			t.Errorf("%s is a forbidden column", c)
		}
	}
	if allowed["created_at"] {
		t.Error("created_at is in the read set")
	}
	if len(mqOperationalSQL) == 0 {
		t.Fatal("the operational registry is empty")
	}
	// The exact vocabulary of the registry: SQL words, catalog names and the allowed columns. Any other identifier
	// (a new column, a function reading a wider set) fails here before it could fail live as 42501.
	vocabulary := set([]string{"select", "from", "where", "and", "in", "is", "not", "null", "order", "by", "any", "as", "coalesce",
		"count", "max", "int", "text", "uuid", "regclass", "regrole", "app", "sessions", "organizations", "users", "slug",
		"platform_label", "current_user", "current_setting", "transaction_read_only", "session_user", "pg_roles", "rolname",
		"rolsuper", "rolbypassrls", "pg_class", "oid", "relrowsecurity", "relforcerowsecurity", "pg_policy", "polname", "polroles",
		"sessions_auth_reader", "users_auth_reader"})
	ident := regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)
	for name, stmt := range mqOperationalSQL {
		for _, tok := range ident.FindAllString(stmt, -1) {
			tok = strings.ToLower(tok)
			switch {
			case banned[tok]:
				t.Errorf("registry %s names the forbidden column %s", name, tok)
			case columns[tok] && !allowed[tok]:
				t.Errorf("registry %s names the app.sessions column %s outside the read set", name, tok)
			case !vocabulary[tok] && !allowed[tok]:
				t.Errorf("registry %s names %q, outside the pinned vocabulary", name, tok)
			}
		}
	}
	for _, c := range mqSessionColumns {
		if !strings.Contains(mqSQLSessions, c) {
			t.Errorf("the session select does not read %s", c)
		}
	}
	// mqRead's querier refuses a non-member before touching the transaction (a nil transaction would panic).
	var q mqTx
	if _, err := q.Query(context.Background(), `SELECT created_at FROM app.sessions`); !errors.Is(err, errMQNotOperational) {
		t.Errorf("non-member Query: %v", err)
	}
	if err := q.QueryRow(context.Background(), `SELECT revoke_reason FROM app.sessions WHERE id = $1`, "x").Scan(new(string)); !errors.Is(err, errMQNotOperational) {
		t.Errorf("non-member QueryRow: %v", err)
	}
	if !mqOperational(mqSQLSessions) || mqOperational(mqSQLSessions+" ") {
		t.Error("registry membership is not exact")
	}
}

// TestMQStepsReadThroughTheRegistry pins, over the tagged file's source text, that no G057 step function references the
// emulator-era runtime bodies (o.sessions(, o.session(, o.waitSession(, o.count() or the runtime pool, and that the
// region reads through mqRead-backed helpers. It is the legacy-call guard only: the forbidden-column assertion is scoped
// to the mqOperationalSQL registry (TestMQSessionColumnsAreSanitized, Plan r5 §3.7 r4), never to source text, because
// step code may legitimately name a column in a sanitized observation key without reading it.
func TestMQStepsReadThroughTheRegistry(t *testing.T) {
	src, err := os.ReadFile("manual_qualification_test.go")
	if err != nil {
		t.Fatal(err)
	}
	const begin, end = "// G057 steps begin", "// G057 steps end"
	text := string(src)
	i, j := strings.Index(text, begin), strings.Index(text, end)
	if i < 0 || j < i {
		t.Fatal("the G057 step region markers are missing")
	}
	region := text[i:j]
	for _, legacy := range []string{"o.sessions(", "o.session(", "o.waitSession(", "o.count(", "o.pool", "o.svc", "o.returned(", "o.latestAuth("} {
		if strings.Contains(region, legacy) {
			t.Errorf("a G057 step references the emulator-era %s", legacy)
		}
	}
	for _, helper := range []string{"mqBaseline(", "mqNewSessions(", "mqWaitNewSession(", "mqBindTenant(", "mqWaitRevoked(", "mqOrganizations("} {
		if !strings.Contains(region, helper) {
			t.Errorf("the G057 steps do not use %s", helper)
		}
	}
}

func mqRow(id, user, origin, org string, auth time.Time, revoked bool) mqSessionState {
	return mqSessionState{ID: id, UserID: user, Kind: "browser", Origin: origin, OrganizationID: org, AuthTime: auth, RenewedAt: auth,
		Revoked: revoked, Assurance: "unknown"}
}

// TestMQNewSessionsAreSetDifference: over in-memory rows, a baseline id with a later renewed_at is reported renewed,
// not new; a baseline id whose revoked flag flipped is not new; an id absent from the baseline is new; ordering is
// auth_time then id; and a post-revocation re-baseline excludes a row inserted before it.
func TestMQNewSessionsAreSetDifference(t *testing.T) {
	const (
		p    = "https://platform.example"
		orgA = "0f16a000-0000-4000-8000-0000000000a1"
	)
	t0 := mqClock("17:35:00")
	s1, s2 := mqRow("s1", "u1", p, orgA, t0, false), mqRow("s2", "u2", p, orgA, t0.Add(time.Minute), false)
	base := mqDiffBaseline(s1, s2)

	renewed := s1
	renewed.RenewedAt = t0.Add(5 * time.Minute)
	revoked := s2
	revoked.Revoked = true
	n1 := mqRow("n1", "u1", p, orgA, t0.Add(3*time.Minute), false)
	d := mqDiff(base, []mqSessionState{renewed, revoked, n1})
	if len(d.New) != 1 || d.New[0].ID != "n1" || d.Renewed != 1 {
		t.Fatalf("delta %+v", d)
	}
	if d := mqDiff(base, []mqSessionState{s1, s2}); len(d.New) != 0 || d.Renewed != 0 {
		t.Fatalf("unchanged rows: %+v", d)
	}
	// A renewal with an identical renewed_at is not a renewal; a baseline row that disappeared is nothing.
	if d := mqDiff(base, []mqSessionState{s1}); len(d.New) != 0 || d.Renewed != 0 {
		t.Fatalf("missing row: %+v", d)
	}
	// Ordering: auth_time, then id.
	n2, n3 := mqRow("b", "u1", p, orgA, t0.Add(2*time.Minute), false), mqRow("a", "u1", p, orgA, t0.Add(2*time.Minute), false)
	d = mqDiff(base, []mqSessionState{n1, n2, n3})
	if got := d.New[0].ID + d.New[1].ID + d.New[2].ID; got != "abn1" {
		t.Fatalf("order %q", got)
	}
	// The Q-6 ordering proof: a row inserted before the revocation is in the re-baseline taken after it; only a row
	// inserted after the re-baseline is new.
	before := mqRow("before", "u2", p, orgA, t0.Add(4*time.Minute), false)
	revokedPrior := s2
	revokedPrior.Revoked = true
	rebase := mqDiffBaseline(s1, revokedPrior, before)
	after := mqRow("after", "u2", p, orgA, t0.Add(6*time.Minute), false)
	d = mqDiff(rebase, []mqSessionState{s1, revokedPrior, before, after})
	if len(d.New) != 1 || d.New[0].ID != "after" {
		t.Fatalf("post-revocation delta %+v", d)
	}
}

func mqDiffBaseline(rows ...mqSessionState) map[string]mqSessionState {
	m := map[string]mqSessionState{}
	for _, r := range rows {
		m[r.ID] = r
	}
	return m
}

// TestMQTenantBinding: a new browser row on https://<label>.P-host[:port] whose Organization is absent from O0 and whose
// platform_label equals the label binds; every other shape is rejected.
func TestMQTenantBinding(t *testing.T) {
	p := mqOrigin{Role: "P", Origin: "https://localhost:543", Host: "localhost", Port: "543"}
	const (
		orgOld = "0f16a000-0000-4000-8000-0000000000a1"
		orgA   = "0f16a000-0000-4000-8000-0000000000a2"
		orgB   = "0f16a000-0000-4000-8000-0000000000b2"
	)
	o0 := map[string]bool{orgOld: true}
	orgs := []mqOrg{{ID: orgOld, Slug: "alpha-agency", PlatformLabel: "org-a"}, {ID: orgA, Slug: "new-agency", PlatformLabel: "k5mx7q2d"},
		{ID: orgB, Slug: "other-agency", PlatformLabel: "p9ab3c4e"}}
	at := mqClock("18:00:00")
	rowA := mqRow("sa", "ua", "https://k5mx7q2d.localhost:543", orgA, at, false)
	a, err := mqBindTenant("A", []mqSessionState{rowA}, orgs, o0, p, nil)
	if err != nil || a.Label != "k5mx7q2d" || a.OrganizationID != orgA || a.Origin != rowA.Origin || a.SessionID != "sa" || a.UserID != "ua" {
		t.Fatalf("binding A: %+v %v", a, err)
	}
	rowB := mqRow("sb", "ub", "https://p9ab3c4e.localhost:543", orgB, at.Add(time.Minute), false)
	b, err := mqBindTenant("B", []mqSessionState{rowB}, orgs, o0, p, []mqTenant{a})
	if err != nil || b.Role != "B" || b.OrganizationID != orgB {
		t.Fatalf("binding B: %+v %v", b, err)
	}
	if _, err := mqBindTenant("C", []mqSessionState{rowB}, orgs, o0, p, nil); err != nil {
		t.Fatalf("the Q-4 match for C: %v", err)
	}
	// A port-less platform root binds a port-less label origin.
	bare := mqOrigin{Role: "P", Origin: "https://platform.example", Host: "platform.example"}
	rowBare := mqRow("sc", "uc", "https://k5mx7q2d.platform.example", orgA, at, false)
	if _, err := mqBindTenant("A", []mqSessionState{rowBare}, orgs, o0, bare, nil); err != nil {
		t.Fatalf("port-less binding: %v", err)
	}
	rows := func(r mqSessionState) []mqSessionState { return []mqSessionState{r} }
	with := func(f func(*mqSessionState)) []mqSessionState { r := rowA; f(&r); return rows(r) }
	for name, c := range map[string]struct {
		fresh []mqSessionState
		role  string
		bound []mqTenant
	}{
		"slug-host origin":                 {with(func(r *mqSessionState) { r.Origin = "https://new-agency.localhost:543" }), "A", nil},
		"two-label host":                   {with(func(r *mqSessionState) { r.Origin = "https://x.k5mx7q2d.localhost:543" }), "A", nil},
		"another port":                     {with(func(r *mqSessionState) { r.Origin = "https://k5mx7q2d.localhost:8443" }), "A", nil},
		"no port":                          {with(func(r *mqSessionState) { r.Origin = "https://k5mx7q2d.localhost" }), "A", nil},
		"http":                             {with(func(r *mqSessionState) { r.Origin = "http://k5mx7q2d.localhost:543" }), "A", nil},
		"the platform root itself":         {with(func(r *mqSessionState) { r.Origin = "https://localhost:543" }), "A", nil},
		"another base domain":              {with(func(r *mqSessionState) { r.Origin = "https://k5mx7q2d.evil.example:543" }), "A", nil},
		"label outside the grammar":        {with(func(r *mqSessionState) { r.Origin = "https://K5MX.localhost:543" }), "A", nil},
		"Organization present in O0":       {with(func(r *mqSessionState) { r.Origin = "https://org-a.localhost:543"; r.OrganizationID = orgOld }), "A", nil},
		"platform_label differs":           {with(func(r *mqSessionState) { r.Origin = "https://p9ab3c4e.localhost:543" }), "A", nil},
		"unreadable Organization":          {with(func(r *mqSessionState) { r.OrganizationID = "0f16a000-0000-4000-8000-0000000000ff" }), "A", nil},
		"unbound session":                  {with(func(r *mqSessionState) { r.OrganizationID = "" }), "A", nil},
		"mcp kind":                         {with(func(r *mqSessionState) { r.Kind = "mcp" }), "A", nil},
		"two new rows":                     {[]mqSessionState{rowA, rowB}, "A", nil},
		"zero new rows":                    {nil, "A", nil},
		"B equal to A's origin":            {rows(rowA), "B", []mqTenant{a}},
		"B equal to A's Organization":      {with(func(r *mqSessionState) { r.ID = "sx" }), "B", []mqTenant{a}},
		"A bound twice":                    {rows(rowB), "A", []mqTenant{a}},
		"not a tenant role":                {rows(rowA), "T", nil},
		"P-prefixed foreign host (suffix)": {with(func(r *mqSessionState) { r.Origin = "https://k5mx7q2d.localhost:543.evil.example" }), "A", nil},
	} {
		if got, err := mqBindTenant(c.role, c.fresh, orgs, o0, p, c.bound); err == nil {
			t.Errorf("%s bound: %+v", name, got)
		}
	}
}

func TestMQSSOReuseObservedForm(t *testing.T) {
	prior := mqClock("17:44:05")
	good := mqSSOObservation{Attestation: "no-prompt", SameUser: true, NewAfterRevocation: true, NewAuth: prior, PriorLatestAuth: prior}
	if ok, note := mqSSOReuseObserved(good); !ok {
		t.Fatalf("the observed reuse does not qualify: %s", note)
	}
	for name, change := range map[string]func(*mqSSOObservation){
		"prompted":                      func(in *mqSSOObservation) { in.Attestation = "prompt" },
		"another user":                  func(in *mqSSOObservation) { in.SameUser = false },
		"not after the revocation":      func(in *mqSSOObservation) { in.NewAfterRevocation = false },
		"a new authentication":          func(in *mqSSOObservation) { in.NewAuth = prior.Add(time.Second) },
		"prompted with a retained time": func(in *mqSSOObservation) { in.Attestation = "prompt"; in.NewAuth = prior },
	} {
		in := good
		change(&in)
		if ok, _ := mqSSOReuseObserved(in); ok {
			t.Errorf("%s qualified as reuse", name)
		}
	}
	// The legacy time form and the boolean form agree on the linked run-2 evidence of TestMQSSOReuseVerdict.
	legacy, _ := mqSSOReuse(mqSSOInput{Attestation: "no-prompt", SameUser: true, RevokedAt: mqClock("17:45:30"), NewCreated: mqClock("17:45:34"),
		NewAuth: prior, PriorLatestAuth: prior})
	if observed, _ := mqSSOReuseObserved(good); legacy != observed {
		t.Fatal("the two forms disagree")
	}
}
