//go:build integration

// G011 B-AUDIT-IT (Plan r2 §9): TestAudit against real PostgreSQL at version 8 (the attested runtime login), the fake
// WorkOS through the real G004 services, and the private MinIO of the migrations suite (the scoped credential). Every
// producer of the closed catalog runs for real: guarded writes and denials through the facade, identity sign-in,
// logout and lifecycle, storage Put/Delete/Sign with provider faults, and privileged fixtures. Rows are read by the
// attested superuser (setup); isolation assertions use the runtime pool or the facade only. Faults are planted by the
// superuser as NOT VALID check constraints or a revoked column privilege and are always removed, with the grants
// re-verified. Failure messages print labels and hashes, never names.
package access

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"core/internal/audit"
	"core/internal/domains"
	"core/internal/identity"
	"core/internal/identity/identitytest"
	"core/internal/identity/workos"
	"core/internal/objectstore"
	"core/internal/observability"
	"core/internal/tenancy"
)

// audWorld is TestAudit's fixture: two Organizations and their users, labelled for readable expectations.
type audWorld struct {
	e      *objEnv
	orgA   tenancy.OrganizationID
	orgB   tenancy.OrganizationID
	g1, g2 tenancy.Group
	prof   tenancy.BusinessProfile
	owner  *itUser // org_owner A
	admin  *itUser // org_admin A
	member *itUser // org_member A, group_member g1
	gadmin *itUser // org_member A, group_admin g1
	multi  *itUser // org_owner B, org_member A
	labels map[string]string
	mark   time.Time
}

func (w *audWorld) label(id string, name string) { w.labels[id] = name }

func audNewWorld(t *testing.T, e *objEnv) *audWorld {
	t.Helper()
	w := &audWorld{e: e, labels: map[string]string{}}
	w.orgA, w.orgB = e.org(t, "Audit A "+e.canaries[0]), e.org(t, "Audit B "+e.canaries[1])
	w.owner, w.admin, w.member, w.gadmin, w.multi = e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t), e.newUser(t)
	e.grantOrg(t, w.orgA, w.owner, tenancy.OrgOwner)
	e.grantOrg(t, w.orgA, w.admin, tenancy.OrgAdmin)
	e.grantOrg(t, w.orgA, w.member, tenancy.OrgMember)
	e.grantOrg(t, w.orgA, w.gadmin, tenancy.OrgMember)
	e.grantOrg(t, w.orgB, w.multi, tenancy.OrgOwner)
	e.grantOrg(t, w.orgA, w.multi, tenancy.OrgMember)
	w.g1 = e.group(t, w.orgA, w.gadmin, "Audit One", fmt.Sprintf("audit-one-%d", itSeq.Add(1)))
	w.g2 = e.group(t, w.orgA, w.owner, "Audit Two", fmt.Sprintf("audit-two-%d", itSeq.Add(1)))
	e.grantGroup(t, w.orgA, w.g1.ID, w.member, tenancy.GroupMember)
	e.changeGroupRole(t, w.orgA, w.g1.ID, w.gadmin, tenancy.GroupAdmin)
	w.prof = e.profile(t, w.orgA, "Audit Profile "+e.canaries[2])
	e.link(t, w.orgA, w.g1.ID, w.prof.ID)
	// G016 A2: one prepared credential per (user, Organization), before any observation window; multi holds a distinct
	// credential for each of its Organizations.
	for _, u := range []*itUser{w.owner, w.admin, w.member, w.gadmin, w.multi} {
		e.bind(t, u, w.orgA)
	}
	e.bind(t, w.multi, w.orgB)
	for id, n := range map[string]string{w.orgA.String(): "A", w.orgB.String(): "B", w.g1.ID.String(): "g1", w.g2.ID.String(): "g2",
		w.prof.ID.String(): "p", w.owner.id.String(): "owner", w.admin.id.String(): "admin", w.member.id.String(): "member",
		w.gadmin.id.String(): "gadmin", w.multi.id.String(): "multi"} {
		w.label(id, n)
	}
	return w
}

func (e *itEnv) changeGroupRole(t *testing.T, org tenancy.OrganizationID, g tenancy.GroupID, u *itUser, r tenancy.GroupRole) {
	t.Helper()
	if _, err := e.model.ChangeGroupRole(context.Background(), org, g, u.id, r); err != nil {
		t.Fatalf("change group role: %v", err)
	}
}

// now returns the database clock (the time base of occurred_at).
func (w *audWorld) now(t *testing.T) time.Time {
	t.Helper()
	var n time.Time
	if err := w.e.setup.QueryRow(context.Background(), `SELECT clock_timestamp()`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// since starts a new observation window.
func (w *audWorld) since(t *testing.T) { t.Helper(); w.mark = w.now(t) }

// audRow is one stored event in a fixed, labelled form:
// action|outcome|reason|operation|actor_kind|actor|org|group|resource_type|resource|subject|role|previous_role|ttl.
type audRow struct {
	text    string
	xmin    string
	reqID   string
	traceID string
	id      string
}

func (w *audWorld) lab(s *string) string {
	if s == nil {
		return "-"
	}
	if n, ok := w.labels[*s]; ok {
		return n
	}
	return "#" + *s
}

// rows returns the events of the window matching where (on audit.events e), oldest first, read by the superuser.
func (w *audWorld) rows(t *testing.T, where string, args ...any) []audRow {
	t.Helper()
	q := `SELECT action, outcome, reason, operation, actor_kind, actor_user_id::text, organization_id::text, group_id::text,
  resource_type, resource_id::text, subject_user_id::text, role, previous_role, ttl_seconds, request_id, trace_id, xmin::text, id::text
FROM audit.events e WHERE occurred_at > $1`
	if where != "" {
		q += " AND (" + where + ")"
	}
	rows, err := w.e.setup.Query(context.Background(), q+" ORDER BY occurred_at, id", append([]any{w.mark}, args...)...)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []audRow
	for rows.Next() {
		var action, outcome, kind, xmin, id string
		var reason, op, actor, org, grp, rtype, res, subj, role, prev, rid, tid *string
		var ttl *int
		if err := rows.Scan(&action, &outcome, &reason, &op, &kind, &actor, &org, &grp, &rtype, &res, &subj, &role, &prev, &ttl,
			&rid, &tid, &xmin, &id); err != nil {
			t.Fatal(err)
		}
		tt := "-"
		if ttl != nil {
			tt = fmt.Sprint(*ttl)
		}
		plain := func(s *string) string {
			if s == nil {
				return "-"
			}
			return *s
		}
		out = append(out, audRow{text: strings.Join([]string{action, outcome, plain(reason), plain(op), kind, w.lab(actor), w.lab(org),
			w.lab(grp), plain(rtype), w.lab(res), w.lab(subj), plain(role), plain(prev), tt}, "|"), xmin: xmin, reqID: plain(rid),
			traceID: plain(tid), id: id})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func texts(rs []audRow) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.text
	}
	return out
}

func (w *audWorld) wantRows(t *testing.T, what string, got []audRow, want ...string) {
	t.Helper()
	if g := texts(got); !slices.Equal(g, want) {
		t.Errorf("%s:\n got %q\nwant %q", what, g, want)
	}
}

// fault plants a superuser NOT VALID check constraint on audit.events (new rows only) and returns its removal.
func (w *audWorld) fault(t *testing.T, check string) func() {
	t.Helper()
	name := fmt.Sprintf("audit_fault_%d", itSeq.Add(1))
	if _, err := w.e.setup.Exec(context.Background(), `ALTER TABLE audit.events ADD CONSTRAINT `+name+` CHECK (`+check+`) NOT VALID`); err != nil {
		t.Fatal(err)
	}
	return func() {
		if _, err := w.e.setup.Exec(context.Background(), `ALTER TABLE audit.events DROP CONSTRAINT `+name); err != nil {
			t.Fatal(err)
		}
	}
}

// audCanonACL is the canonical text of one aclitem[] expression (G013 amendment 01): every complete entry (grantee,
// privileges with their grant options, grantor) in sorted order, duplicates kept (no DISTINCT), "null" for a NULL ACL and
// "{}" for an explicit empty one. Equal privileges compare equal whatever their stored order (a revoke and re-grant
// reorders the items when another role holds rights on the same column); any changed entry compares unequal.
func audCanonACL(expr string) string {
	return `CASE WHEN ` + expr + ` IS NULL THEN 'null' ELSE '{' || coalesce((SELECT string_agg(x::text, ',' ORDER BY x::text) FROM unnest(` +
		expr + `) x), '') || '}' END`
}

// auditACL is the audit.events table and column ACL (CT-AUD-3's inventory), read by the superuser: the table ACL and
// every column with an ACL, keyed by attname, each in canonical form.
func (w *audWorld) auditACL(t *testing.T) string {
	t.Helper()
	var s string
	if err := w.e.setup.QueryRow(context.Background(), `SELECT `+audCanonACL("relacl")+` || '|' || coalesce((SELECT string_agg(attname || '=' || `+
		audCanonACL("attacl")+`, ',' ORDER BY attname)
  FROM pg_attribute WHERE attrelid = 'audit.events'::regclass AND attacl IS NOT NULL), '') FROM pg_class WHERE oid = 'audit.events'::regclass`).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

// audCanonACLControl is the helper's focused control (amendment 01): reordered equal entries compare equal; a changed
// privilege, a changed grant option or a lost duplicate compares unequal; NULL and an explicit empty ACL stay distinct.
func (w *audWorld) audCanonACLControl(t *testing.T) {
	t.Helper()
	ow, rt, rn := w.e.cfg.OwnerRole(), w.e.cfg.RuntimeRole(), w.e.cfg.RetentionRole()
	canon := func(lit *string) string {
		var s string
		if err := w.e.setup.QueryRow(context.Background(), `SELECT `+audCanonACL("$1::aclitem[]"), lit).Scan(&s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	lit := func(items ...string) *string { s := "{" + strings.Join(items, ",") + "}"; return &s }
	a, b := rt+"=a/"+ow, rn+"=r/"+ow
	if canon(lit(a, b)) != canon(lit(b, a)) {
		t.Error("ACL control: reordered equal entries compare unequal")
	}
	for name, other := range map[string]*string{
		"changed privilege":    lit(rt+"=r/"+ow, b),
		"changed grant option": lit(rt+"=a*/"+ow, b),
		"lost duplicate":       lit(a),
		"null":                 nil,
		"explicit empty":       lit(),
	} {
		if canon(lit(a, b)) == canon(other) {
			t.Errorf("ACL control: %s compares equal", name)
		}
	}
	if canon(lit(a, a)) == canon(lit(a)) || canon(nil) == canon(lit()) {
		t.Error("ACL control: multiplicity or the NULL/explicit distinction is lost")
	}
}

// audAdmit is the audit suite's G016 D-2(a) fixture admission (integration_test.go admit): the subject's user, principal
// and sign-in Organization membership are provisioned once and kept (idempotent), so a test that admits a subject before
// its capture window or planted fault sees only the sign-in's own rows. An empty subject (a provider error) is a no-op.
func (e *itEnv) audAdmit(t *testing.T, sub string) {
	t.Helper()
	if sub == "" {
		return
	}
	var n int
	if err := e.setupDB.QueryRow(context.Background(), `SELECT count(*)::int FROM app.role_bindings b
JOIN app.identity_principals p ON p.user_id = b.subject_user_id
WHERE p.issuer = $1 AND p.subject = $2 AND b.organization_id = $3::text::uuid AND b.scope_type = 'organization'`,
		e.fake.Issuer, sub, e.loginOrg.String()).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		_ = e.admit(t, e.loginOrg, sub, "Audit Fixture") // kept: the release is not called
	}
}

// callback completes one login of sub through id with the given extra query (code or error) and live credential.
func (e *itEnv) audCallback(t *testing.T, id *identity.Service, sub, sid, extra, live string) (identity.CallbackOutput, error) {
	t.Helper()
	ctx := context.Background()
	e.audAdmit(t, sub)
	out, err := id.Login(ctx, identity.LoginInput{Origin: itOrigin, ReturnTo: "/account"})
	if err != nil {
		t.Fatalf("login start: %v", err)
	}
	u, err := url.Parse(out.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	q := "state=" + url.QueryEscape(u.Query().Get("state"))
	switch extra {
	case "code":
		code := e.fake.IssueCode(u.Query().Get("code_challenge"), identitytest.Grant{UserID: sub, FirstName: "Audit", LastName: "Fixture", SID: sid})
		q += "&code=" + url.QueryEscape(code)
	default:
		q += "&" + extra
	}
	return id.Callback(ctx, identity.CallbackInput{Origin: itOrigin, RawQuery: q, Binding: out.Binding, Credential: live})
}

func (e *itEnv) sessionOf(t *testing.T, cred string) identity.SessionInfo {
	t.Helper()
	info, err := e.id.Session(context.Background(), cred, itOrigin)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	return info
}

// logs returns the decoded log records written since n (the byte offset of the log buffer).
func (e *itEnv) recordsSince(t *testing.T, n int) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(e.logs.String()[n:]), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("log record: %v", err)
		}
		out = append(out, m)
	}
	return out
}

func msgs(recs []map[string]any) []string {
	var out []string
	for _, r := range recs {
		m, _ := r["event"].(string)
		out = append(out, m)
	}
	return out
}

// eAuditTrigger prints one escalation line (E-AUDIT-TRIGGER) for an unexpected cost observation.
func eAuditTrigger(fields map[string]any) {
	fields["evidence"] = "E-AUDIT-TRIGGER"
	b, _ := json.Marshal(fields)
	fmt.Println("E-AUDIT-TRIGGER " + string(b))
}

func TestAudit(t *testing.T) {
	e := setupObjects(t)
	ctx := context.Background()
	w := audNewWorld(t, e)
	observed := map[string]bool{}
	note := func(rs []audRow) {
		for _, r := range rs {
			observed[strings.SplitN(r.text, "|", 2)[0]] = true
		}
	}
	// tc resolves (org, g) in a request scope of u's credential for org (G016 A2).
	tc := func(t *testing.T, s *Service, u *itUser, org tenancy.OrganizationID, g tenancy.GroupID) (context.Context, *TenantContext) {
		t.Helper()
		rctx, _ := beginAs(t, s, ctx, u, org)
		return rctx, resolve(t, s, rctx, org, g)
	}

	t.Run("AT-ATTEST runtime login and scoped credential", func(t *testing.T) {
		if err := e.rlsAttest(e.pool); err != nil {
			t.Fatal(err)
		}
		for _, k := range []string{"MINIO_ROOT_USER", "MINIO_ROOT_PASSWORD"} {
			if os.Getenv(k) != "" {
				t.Errorf("%s is present in the test environment", k)
			}
		}
		var super bool
		if err := e.setup.QueryRow(ctx, `SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).Scan(&super); err != nil || !super {
			t.Fatal("setup is not the attested superuser")
		}
	})

	t.Run("AT-FMT exact fields, attribution and refused forgeries", func(t *testing.T) {
		// A user event (a guarded denial), a system event (disable) and unknown events (privileged fixtures).
		w.since(t)
		rctx, mtc := tc(t, e.osvc, w.member, w.orgA, tenancy.GroupID{})
		_, err := e.osvc.RenameGroup(rctx, mtc, w.g2.ID, "Member Rename")
		wantCode(t, "member renames g2", err, CodeNotFound)
		_, err = e.osvc.RenameGroup(rctx, mtc, w.g1.ID, "Member Rename")
		wantCode(t, "member renames g1", err, CodeForbidden)
		got := w.rows(t, "action = 'access.denied'")
		note(got)
		w.wantRows(t, "denials", got,
			"access.denied|denied|target_not_found|group.rename|user|member|A|-|-|-|-|-|-|-",
			"access.denied|denied|role_not_permitted|group.rename|user|member|A|g1|group|g1|-|-|-|-")
		// A selector of another Organization never becomes a recorded UUID; an unestablished Organization is global.
		w.since(t)
		_, err = e.osvc.GetGroup(rctx, mtc, mustParse[tenancy.GroupID](w.orgB.String()))
		wantCode(t, "foreign selector", err, CodeNotFound)
		rctx2, _ := beginAs(t, e.osvc, ctx, w.member, w.orgA) // G016 A2: member's credential for A; both refusals kept
		_, err = e.osvc.Resolve(rctx2, ResolveRequest{OrganizationID: w.orgB.String()})
		wantCode(t, "resolve B as member", err, CodeContextDenied)
		_, err = e.osvc.Resolve(rctx2, ResolveRequest{OrganizationID: w.orgA.String(), GroupID: w.g2.ID.String()})
		wantCode(t, "resolve g2 as member", err, CodeContextDenied)
		got = w.rows(t, "action = 'access.denied'")
		// A Group the subject cannot select makes the Organization gate itself fail under row-level security
		// (rls_org_role() is NULL for an unselectable Group): the refusal is organization_not_permitted, global.
		w.wantRows(t, "unestablished scope", got,
			"access.denied|denied|target_not_found|group.get|user|member|A|-|-|-|-|-|-|-",
			"access.denied|denied|organization_not_permitted|access.resolve|user|member|-|-|-|-|-|-|-|-",
			"access.denied|denied|organization_not_permitted|access.resolve|user|member|-|-|-|-|-|-|-|-")
		// Privileged writes with no application context: unknown, kept apart from user attribution; lifecycle actions
		// without the lifecycle producer's context are unknown too (Q3).
		w.since(t)
		tmp := e.org(t, "Audit Temp")
		w.label(tmp.String(), "tmp")
		if _, err := e.setup.Exec(ctx, `UPDATE app.users SET status = 'disabled', disabled_at = now() WHERE id = $1::text::uuid`, w.multi.id.String()); err != nil {
			t.Fatal(err)
		}
		if _, err := e.setup.Exec(ctx, `UPDATE app.users SET status = 'active', disabled_at = NULL WHERE id = $1::text::uuid`, w.multi.id.String()); err != nil {
			t.Fatal(err)
		}
		got = w.rows(t, "")
		note(got)
		w.wantRows(t, "privileged", got,
			"organization.created|success|-|-|unknown|-|tmp|-|organization|tmp|-|-|-|-",
			"user.disabled|success|-|-|unknown|-|-|-|user|multi|multi|-|-|-",
			"user.enabled|success|-|-|unknown|-|-|-|user|multi|multi|-|-|-")
		gB := e.group(t, w.orgB, w.multi, "Audit B One", fmt.Sprintf("audit-b-one-%d", itSeq.Add(1)))
		// Forged direct appends through the attested runtime login, each refused with its detector code and rolled back.
		rt := e.pool
		for _, c := range []struct {
			what, ctxUser, ctxOrg, sql, state, constraint string
			args                                          []any
		}{
			{"forged actor column", w.owner.id.String(), w.orgA.String(),
				`INSERT INTO audit.events (action, outcome, actor_kind, actor_user_id) VALUES ('auth.sign_in', 'failed', 'user', $1::text::uuid)`, "42501", "",
				[]any{w.admin.id.String()}},
			{"foreign Organization", w.owner.id.String(), w.orgA.String(),
				`INSERT INTO audit.events (action, outcome, reason, operation, organization_id) VALUES ('access.denied', 'denied', 'target_not_found', 'group.get', $1::text::uuid)`,
				"42501", "audit_event_scope", []any{w.orgB.String()}},
			{"unestablished Organization", w.member.id.String(), w.orgB.String(),
				`INSERT INTO audit.events (action, outcome, reason, operation, organization_id) VALUES ('access.denied', 'denied', 'target_not_found', 'group.get', $1::text::uuid)`,
				"42501", "audit_event_scope", []any{w.orgB.String()}},
			{"Group of another Organization (owner appender)", w.multi.id.String(), w.orgB.String(),
				`INSERT INTO audit.events (action, outcome, reason, operation, organization_id, group_id) VALUES ('access.denied', 'denied', 'target_not_found', 'group.get', $1::text::uuid, $2::text::uuid)`,
				"42501", "audit_event_group", []any{w.orgB.String(), w.g1.ID.String()}},
			{"nonexistent Group (admin appender)", w.admin.id.String(), w.orgA.String(),
				`INSERT INTO audit.events (action, outcome, reason, operation, organization_id, group_id) VALUES ('access.denied', 'denied', 'target_not_found', 'group.get', $1::text::uuid, gen_random_uuid())`,
				"42501", "audit_event_group", []any{w.orgA.String()}},
			// S-review F1: with a Group, the refusal rests on groups_select requiring app.rls_org_role() IS NOT NULL and the
			// appender's own Group visibility.
			{"Group of the context Organization, appender without a role in it", w.member.id.String(), w.orgB.String(),
				`INSERT INTO audit.events (action, outcome, reason, operation, organization_id, group_id) VALUES ('access.denied', 'denied', 'target_not_found', 'group.get', $1::text::uuid, $2::text::uuid)`,
				"42501", "audit_event_group", []any{w.orgB.String(), gB.ID.String()}},
			{"same-Organization Group invisible to a Group-role appender", w.gadmin.id.String(), w.orgA.String(),
				`INSERT INTO audit.events (action, outcome, reason, operation, organization_id, group_id) VALUES ('access.denied', 'denied', 'target_not_found', 'group.get', $1::text::uuid, $2::text::uuid)`,
				"42501", "audit_event_group", []any{w.orgA.String(), w.g2.ID.String()}},
			{"trigger-owned action", w.owner.id.String(), w.orgA.String(),
				`INSERT INTO audit.events (action, outcome, organization_id, resource_type, resource_id) VALUES ('organization.renamed', 'success', $1::text::uuid, 'organization', $1::text::uuid)`,
				"42501", "audit_event_source", []any{w.orgA.String()}},
			{"bad action", w.owner.id.String(), w.orgA.String(),
				`INSERT INTO audit.events (action, outcome) VALUES ('nope', 'failed')`, "23514", "audit_event_shape", nil},
			{"bad reason", w.owner.id.String(), "",
				`INSERT INTO audit.events (action, outcome, reason) VALUES ('auth.sign_in', 'failed', 'login_canceled')`, "23514", "audit_event_shape", nil},
			{"bad shape", w.owner.id.String(), "",
				`INSERT INTO audit.events (action, outcome, reason, ttl_seconds) VALUES ('auth.sign_in', 'failed', 'provider_error', 5)`, "23514", "audit_event_shape", nil},
		} {
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`, c.ctxUser, c.ctxOrg); err != nil {
				t.Fatal(err)
			}
			_, err = tx.Exec(ctx, c.sql, c.args...)
			_ = tx.Rollback(ctx)
			code, constraint := "", ""
			if p := pgErr(err); p != nil {
				code, constraint = p.Code, p.ConstraintName
			}
			if code != c.state || (c.constraint != "" && constraint != c.constraint) {
				t.Errorf("%s: %s/%s (%v), want %s/%s", c.what, code, constraint, err, c.state, c.constraint)
			}
		}
		// Malformed correlation is stored as NULL; well-formed correlation is stored (a direct append by the owner).
		for _, c := range []struct{ rid, tid, wantRID, wantTID string }{
			{"bad id with spaces", "XYZ", "-", "-"},
			{"req-0123456789", "0123456789abcdef0123456789abcdef", "req-0123456789", "0123456789abcdef0123456789abcdef"},
		} {
			w.since(t)
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('audit.request_id', $2, true), set_config('audit.trace_id', $3, true)`,
				w.owner.id.String(), c.rid, c.tid); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO audit.events (action, outcome, reason) VALUES ('auth.sign_in', 'failed', 'provider_error')`); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			got := w.rows(t, "action = 'auth.sign_in'")
			if len(got) != 1 || got[0].reqID != c.wantRID || got[0].traceID != c.wantTID ||
				got[0].text != "auth.sign_in|failed|provider_error|-|user|owner|-|-|-|-|-|-|-|-" {
				t.Errorf("correlation %q: %+v", c.rid, got)
			}
		}
		// The positive: a signed URL of a Group object carries that Group. G060 (Plan r0 A9): an Organization admin with no
		// binding in g1 no longer signs g1's objects (not_found, no url row) under either setting; g1's admin signs.
		w.since(t)
		rctx3, atc := tc(t, e.osvc, w.admin, w.orgA, tenancy.GroupID{})
		info := audPut(t, e, w.gadmin, w.orgA, w.g1.ID, "fmt")
		w.label(info.ID.String(), "obj")
		if _, err := e.osvc.SignedGetURL(rctx3, atc, SignedGetURLRequest{Object: info.ID, DefaultTTL: true}); CodeOf(err) != CodeNotFound {
			t.Errorf("Organization admin signs a Group object: %v, want not_found (G060)", err)
		}
		grctx, gtc := tc(t, e.osvc, w.gadmin, w.orgA, tenancy.GroupID{})
		if _, err := e.osvc.SignedGetURL(grctx, gtc, SignedGetURLRequest{Object: info.ID, DefaultTTL: true}); err != nil {
			t.Fatalf("Group admin sign: %v", err)
		}
		got = w.rows(t, "action = 'object.url_issued'")
		note(got)
		w.wantRows(t, "Group admin url", got, fmt.Sprintf("object.url_issued|success|-|-|user|gadmin|A|g1|stored_object|obj|-|-|-|%d",
			int(e.ocfg.DefaultTTL/time.Second)))
	})

	idlog := &itSyncBuf{}
	idl := e.audIdentity(t, idlog)
	defer func() { _, _ = e.logs.Write([]byte(idlog.String())) }() // the identity records join the validated log

	t.Run("AT-ATOM same-transaction writes, strict faults and revocation priority", func(t *testing.T) {
		rctx, otc := tc(t, e.osvc, w.owner, w.orgA, tenancy.GroupID{})
		// M1: the guarded rename and its event are one transaction.
		w.since(t)
		if _, err := e.osvc.RenameGroup(rctx, otc, w.g2.ID, "Audit Two Renamed"); err != nil {
			t.Fatal(err)
		}
		var gx string
		if err := e.setup.QueryRow(ctx, `SELECT xmin::text FROM app.groups WHERE id = $1::text::uuid`, w.g2.ID.String()).Scan(&gx); err != nil {
			t.Fatal(err)
		}
		got := w.rows(t, "")
		note(got)
		w.wantRows(t, "rename", got, "group.renamed|success|-|-|user|owner|A|g2|group|g2|-|-|-|-")
		if len(got) == 1 && got[0].xmin != gx {
			t.Errorf("rename event xmin %s, group row xmin %s: not one transaction", got[0].xmin, gx)
		}
		// G063 (Plan r1 §6, r1 C1/C5): the rename and mcp.connection_renamed are one transaction. Under an audit fault the
		// strict capture refuses the rename: the name stays NULL, no rename event exists, and neither the connection row
		// (its xmin) nor any security fact of the connection or its 'mcp' session changes. Without the fault the event's
		// xmin equals the connection row's.
		atomConn, err := identity.NewMCPConnectionID()
		if err != nil {
			t.Fatal(err)
		}
		w.label(atomConn, "aconn")
		if err := e.osvc.CreateMCPConnection(rctx, otc, atomConn, "org-a.localhost:543", "https://org-a.localhost:543/mcp/c/"+atomConn,
			identity.MCPLifetimes{Pending: 30 * time.Minute, Idle: 24 * time.Hour, Absolute: 48 * time.Hour}, nil); err != nil {
			t.Fatal(err)
		}
		connFacts := func() string {
			t.Helper()
			var s string
			if err := e.setup.QueryRow(ctx, `SELECT concat_ws('|', c.xmin::text, c.state, c.client_id, c.selected_group_id::text,
  c.context_version::text, c.activated_at::text, coalesce(c.display_name, '<null>'), s.revoked_at::text, s.idle_expires_at::text,
  s.absolute_expires_at::text, s.renewed_at::text) FROM app.mcp_connections c JOIN app.sessions s ON s.id = c.session_id
  WHERE c.id = $1::text::uuid`, atomConn).Scan(&s); err != nil {
				t.Fatal(err)
			}
			return s
		}
		beforeFault := connFacts()
		removeFault := w.fault(t, "false")
		atomName := "Atomic laptop"
		ferr := e.osvc.RenameMCPConnection(rctx, otc, atomConn, &atomName)
		removeFault()
		if ferr == nil {
			t.Fatal("rename under an audit fault succeeded")
		}
		if got := connFacts(); got != beforeFault {
			t.Errorf("the refused rename changed the connection:\n before %s\n after  %s", beforeFault, got)
		}
		if got := w.rows(t, "action = 'mcp.connection_renamed' AND resource_id = $2::text::uuid", atomConn); len(got) != 0 {
			t.Errorf("rename events after the refused rename: %q", texts(got))
		}
		w.since(t)
		if err := e.osvc.RenameMCPConnection(rctx, otc, atomConn, &atomName); err != nil {
			t.Fatalf("rename after the fault: %v", err)
		}
		var cx string
		if err := e.setup.QueryRow(ctx, `SELECT xmin::text FROM app.mcp_connections WHERE id = $1::text::uuid`, atomConn).Scan(&cx); err != nil {
			t.Fatal(err)
		}
		got = w.rows(t, "")
		note(got)
		w.wantRows(t, "connection rename", got, "mcp.connection_renamed|success|-|-|user|owner|A|-|mcp_connection|aconn|owner|-|-|-")
		if len(got) == 1 && got[0].xmin != cx {
			t.Errorf("rename event xmin %s, connection row xmin %s: not one transaction", got[0].xmin, cx)
		}
		// M2: sign-in (both branches) through inTxAudited: the session row and auth.sign_in share the transaction.
		// G016 D-2(a): the account is provisioned first (an operator's act, outside the window), so the first sign-in
		// issues a session for an existing user and creates none.
		u := &itUser{sub: fmt.Sprintf("user_01AUDIT%08d", itSeq.Add(1))}
		e.audAdmit(t, u.sub)
		w.since(t)
		out, err := e.audCallback(t, e.id, u.sub, fmt.Sprintf("session_01AU%08d", itSeq.Add(1)), "code", "")
		if err != nil || out.Outcome != identity.OutcomeSignedIn {
			t.Fatalf("sign in: %+v %v", out, err)
		}
		u.cred = out.Credential
		info := e.sessionOf(t, u.cred)
		u.id = mustParse[tenancy.UserID](info.UserID)
		w.label(info.UserID, "u")
		w.label(info.SessionID, "s1")
		got = w.rows(t, "")
		note(got)
		w.wantRows(t, "first sign-in", got,
			"session.issued|success|-|-|user|u|-|-|session|s1|u|-|-|-",
			"auth.sign_in|success|signed_in|-|user|u|-|-|-|-|u|-|-|-")
		// The session row, its event and auth.sign_in share the sign-in's top-level transaction id.
		var sx string
		if err := e.setup.QueryRow(ctx, `SELECT xmin::text FROM app.sessions WHERE id = $1::text::uuid`, info.SessionID).Scan(&sx); err != nil {
			t.Fatal(err)
		}
		if len(got) == 2 && (got[0].xmin != sx || got[1].xmin != sx) {
			t.Error("first sign-in events are not in the sign-in transaction")
		}
		w.since(t)
		out, err = e.audCallback(t, e.id, u.sub, fmt.Sprintf("session_01AU%08d", itSeq.Add(1)), "code", u.cred)
		if err != nil || out.Outcome != identity.OutcomeSessionRotated {
			t.Fatalf("rotation: %+v %v", out, err)
		}
		info2 := e.sessionOf(t, out.Credential)
		w.label(info2.SessionID, "s2")
		u.cred = out.Credential
		got = w.rows(t, "")
		note(got)
		w.wantRows(t, "rotation", got,
			"session.revoked|success|superseded|-|user|u|-|-|session|s1|u|-|-|-",
			"session.issued|success|-|-|user|u|-|-|session|s2|u|-|-|-",
			"auth.sign_in|success|session_rotated|-|user|u|-|-|-|-|u|-|-|-")
		if len(got) == 3 && (got[0].xmin != got[2].xmin || got[1].xmin != got[2].xmin) {
			t.Error("rotation events are not one transaction")
		}

		// Strict paths: a planted fault fails the operation with neither effect, then the operation succeeds.
		v := &itUser{sub: fmt.Sprintf("user_01AUDIT%08d", itSeq.Add(1))}
		e.audAdmit(t, v.sub) // G016 D-2(a): provisioned before the fault (the sign-in under the fault is the strict path)
		remove := w.fault(t, "false")
		before := e.state(t, w.orgA)
		_, err = e.osvc.RenameGroup(rctx, otc, w.g2.ID, "Faulted Rename")
		if err == nil || e.state(t, w.orgA) != before {
			t.Errorf("rename under an audit fault: %v (state changed %v)", err, e.state(t, w.orgA) != before)
		}
		var n0, n1 int
		count := func() int {
			var n int
			if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM app.sessions WHERE user_id = $1::text::uuid`, info.UserID).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		n0 = count()
		if _, err := e.audCallback(t, e.id, u.sub, fmt.Sprintf("session_01AU%08d", itSeq.Add(1)), "code", u.cred); err == nil {
			t.Error("rotation under an audit fault succeeded")
		}
		if live := e.sessionOf(t, u.cred); live.SessionID != info2.SessionID {
			t.Error("the faulted rotation superseded the live session")
		}
		if _, err := e.audCallback(t, e.id, v.sub, fmt.Sprintf("session_01AU%08d", itSeq.Add(1)), "code", ""); err == nil {
			t.Error("sign-in under an audit fault succeeded")
		}
		if err := e.id.DisableUser(ctx, info.UserID); err != nil {
			t.Fatalf("disable under an audit fault: %v (revocation priority)", err)
		}
		if err := e.id.EnableUser(ctx, info.UserID); err == nil {
			t.Error("enable under an audit fault succeeded")
		}
		if _, err := e.model.GrantOrganizationRole(ctx, w.orgB, u.id, tenancy.OrgMember); err == nil {
			t.Error("grant under an audit fault succeeded")
		}
		n1 = count()
		remove()
		if n1 != n0 {
			t.Errorf("sessions %d -> %d under the strict faults", n0, n1)
		}
		if err := e.id.EnableUser(ctx, info.UserID); err != nil {
			t.Fatalf("enable after the fault: %v", err)
		}
		if _, err := e.osvc.RenameGroup(rctx, otc, w.g2.ID, "Audit Two Again"); err != nil {
			t.Fatalf("rename after the fault: %v", err)
		}

		// Revocation priority: a fault on the revocation events only. Logout (current, all, provider) and disable commit;
		// revocation.unaudited is recorded and audit.gap reports it.
		revFault := "action NOT IN ('session.revoked', 'user.disabled')"
		cases := []struct {
			name, scope, reason, resource string
		}{
			{"logout", identity.ScopeCurrent, "logout", "session"},
			{"logout_all", identity.ScopeAll, "logout_all", "user"},
			{"provider_logout", identity.ScopeProvider, "provider_logout", "session"},
			{"disabled", "", "disabled", "user"},
		}
		for _, mode := range []struct {
			name, check, independent string
		}{{"revocation fault", revFault, "recorded"}, {"every insert faulted", "false", "failed"}} {
			for _, c := range cases {
				x := e.newUser(t)
				w.label(x.id.String(), "x")
				w.label(x.session, "xs")
				cred2 := e.relogin(t, x)
				w.label(e.sessionOf(t, cred2).SessionID, "xs2")
				remove := w.fault(t, mode.check)
				mark := len(idlog.String())
				w.since(t)
				var err error
				if c.scope == "" {
					err = idl.DisableUser(ctx, x.id.String())
				} else {
					_, err = idl.Logout(ctx, identity.LogoutInput{Scope: c.scope, CSRF: e.csrf(t, x.cred), Origin: itOrigin, Credential: x.cred})
				}
				remove()
				// Every insert faulted also fails provider logout's strict bookkeeping step (session.provider_revocation_
				// recorded) after the committed revocation: identity_unavailable, as for any failure of that step.
				bookkeeping := mode.independent == "failed" && c.scope == identity.ScopeProvider
				if (err != nil) != bookkeeping || (bookkeeping && !errors.Is(err, identity.ErrUnavailable)) {
					t.Errorf("%s/%s: %v (revocation priority)", mode.name, c.name, err)
					continue
				}
				if _, err := e.id.Session(ctx, x.cred, itOrigin); !errors.Is(err, identity.ErrUnauthenticated) {
					t.Errorf("%s/%s: the session is still live", mode.name, c.name)
				}
				got := w.rows(t, "subject_user_id = $2::text::uuid AND action IN ('session.revoked', 'user.disabled', 'revocation.unaudited')", x.id.String())
				note(got)
				actor := "user|x"
				if c.reason == "disabled" {
					actor = "system|-"
				}
				res := "xs"
				if c.resource == "user" {
					res = "x"
				}
				want := []string{}
				if mode.independent == "recorded" {
					want = []string{fmt.Sprintf("revocation.unaudited|success|%s|-|%s|-|-|%s|%s|x|-|-|-", c.reason, actor, c.resource, res)}
				}
				w.wantRows(t, mode.name+"/"+c.name, got, want...)
				if gaps := gapsSince(t, idlog, mark); !slices.Equal(gaps, []string{c.reason + "/" + mode.independent}) {
					t.Errorf("%s/%s: audit.gap %v", mode.name, c.name, gaps)
				}
			}
		}

		// Privilege fault: the runtime loses its column INSERT privilege on audit.events (the app UPDATE privileges stay).
		w.audCanonACLControl(t)
		acl := w.auditACL(t)
		x := e.newUser(t)
		w.label(x.id.String(), "px")
		cred2 := e.relogin(t, x)
		info3 := e.sessionOf(t, cred2)
		if _, err := e.setup.Exec(ctx, `UPDATE app.sessions SET idle_expires_at = now() - interval '1 minute' WHERE id = $1::text::uuid`, info3.SessionID); err != nil {
			t.Fatal(err)
		}
		y, z, r := e.newUser(t), e.newUser(t), e.newUser(t)
		rotLive := e.sessionOf(t, r.cred).SessionID
		csrfX, csrfY := e.csrf(t, x.cred), e.csrf(t, y.cred)
		rt := pgx.Identifier{e.cfg.RuntimeRole()}.Sanitize()
		if _, err := e.setup.Exec(ctx, `REVOKE INSERT (action, outcome, reason, operation, resource_type, resource_id, subject_user_id, organization_id, `+
			`group_id, role, previous_role, ttl_seconds) ON audit.events FROM `+rt); err != nil {
			t.Fatal(err)
		}
		restored := false
		restore := func() {
			if restored {
				return
			}
			restored = true
			if _, err := e.setup.Exec(ctx, `GRANT INSERT (action, outcome, reason, operation, resource_type, resource_id, subject_user_id, organization_id, `+
				`group_id, role, previous_role, ttl_seconds) ON audit.events TO `+rt); err != nil {
				t.Fatal(err)
			}
		}
		t.Cleanup(restore)
		mark := len(idlog.String())
		_, errAll := idl.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeAll, CSRF: csrfX, Origin: itOrigin, Credential: x.cred})
		_, errProv := idl.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeProvider, CSRF: csrfY, Origin: itOrigin, Credential: y.cred})
		errDis := idl.DisableUser(ctx, z.id.String())
		_, errRot := e.audCallback(t, e.id, r.sub, fmt.Sprintf("session_01AU%08d", itSeq.Add(1)), "code", r.cred)
		signedIn := e.newUserMaybe(t)
		restore()
		// The revocations commit; provider logout's later strict bookkeeping step fails (identity_unavailable) after its
		// committed revocation.
		if errAll != nil || errDis != nil || !errors.Is(errProv, identity.ErrUnavailable) {
			t.Errorf("revocations under the privilege fault: all %v, provider %v, disable %v", errAll, errProv, errDis)
		}
		for _, u := range []*itUser{x, y, z} {
			if _, err := e.id.Session(ctx, u.cred, itOrigin); !errors.Is(err, identity.ErrUnauthenticated) {
				t.Errorf("a revocation under the privilege fault did not commit (%s)", itH(u.sub))
			}
		}
		if errRot == nil || signedIn || e.sessionOf(t, r.cred).SessionID != rotLive {
			t.Errorf("strict paths under the privilege fault: rotation %v, new sign-in succeeded %v", errRot, signedIn)
		}
		var revoked int
		if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM app.sessions WHERE user_id = $1::text::uuid AND revoke_reason = 'logout_all'`,
			x.id.String()).Scan(&revoked); err != nil || revoked != 2 {
			t.Errorf("logout-all revoked %d sessions (incl. the expired-but-unrevoked one), want 2 (%v)", revoked, err)
		}
		if gaps := gapsSince(t, idlog, mark); !slices.Equal(gaps, []string{"logout_all/failed", "provider_logout/failed", "disabled/failed"}) {
			t.Errorf("privilege fault audit.gap %v", gaps)
		}
		if got := w.auditACL(t); got != acl {
			t.Errorf("audit ACL after restoring the grants:\n%s\nwant\n%s", got, acl)
		}

		// A stalled recorder rollback is bounded by the cleanup budget and its connection is not reused.
		audStalledRecorder(t, e)

		// Business rollback, denial with an unchanged target, and a recording fault.
		w.since(t)
		_, err = e.osvc.ChangeGroupSlug(rctx, otc, w.g2.ID, w.g1.Slug)
		wantCode(t, "slug conflict", err, CodeConflict)
		if got := w.rows(t, ""); len(got) != 0 {
			t.Errorf("a business rollback left events %q", texts(got))
		}
		mtx, mtc := tc(t, e.osvc, w.member, w.orgA, tenancy.GroupID{})
		remove = w.fault(t, "action <> 'access.denied'")
		mark = len(e.logs.String())
		before = e.state(t, w.orgA)
		_, err = e.osvc.RenameGroup(mtx, mtc, w.g1.ID, "Denied Rename")
		remove()
		wantCode(t, "denial under a recording fault", err, CodeForbidden)
		if ms := msgs(e.recordsSince(t, mark)); !slices.Equal(ms, []string{"audit.record_failed", "access.decision"}) || e.state(t, w.orgA) != before {
			t.Errorf("recording fault records %v", ms)
		}
	})

	t.Run("AT-ID identity events and fan-out", func(t *testing.T) {
		x := e.newUser(t)
		w.label(x.id.String(), "x")
		c2 := e.relogin(t, x)
		c3 := e.relogin(t, x)
		s3 := e.sessionOf(t, c3).SessionID
		if _, err := e.setup.Exec(ctx, `UPDATE app.sessions SET idle_expires_at = now() - interval '1 minute' WHERE id = $1::text::uuid`, s3); err != nil {
			t.Fatal(err)
		}
		w.since(t)
		e.logout(t, c2, identity.ScopeAll)
		got := w.rows(t, "subject_user_id = $2::text::uuid", x.id.String())
		note(got)
		if len(got) != 3 || !strings.HasPrefix(got[0].text, "session.revoked|success|logout_all|-|user|x|") {
			t.Errorf("logout-all over 3 unrevoked sessions (1 expired): %q", texts(got))
		}
		// Disable over sessions incl. an expired-but-unrevoked one; enable.
		d := e.newUser(t)
		w.label(d.id.String(), "d")
		cd := e.relogin(t, d)
		if _, err := e.setup.Exec(ctx, `UPDATE app.sessions SET idle_expires_at = now() - interval '1 minute' WHERE id = $1::text::uuid`,
			e.sessionOf(t, cd).SessionID); err != nil {
			t.Fatal(err)
		}
		w.since(t)
		if err := e.id.DisableUser(ctx, d.id.String()); err != nil {
			t.Fatal(err)
		}
		if err := e.id.EnableUser(ctx, d.id.String()); err != nil {
			t.Fatal(err)
		}
		got = w.rows(t, "subject_user_id = $2::text::uuid", d.id.String())
		note(got)
		var sys []string
		for _, r := range got {
			f := strings.Split(r.text, "|")
			sys = append(sys, f[0]+"|"+f[2]+"|"+f[4])
		}
		if !slices.Equal(sys, []string{"user.disabled|-|system", "session.revoked|disabled|system", "session.revoked|disabled|system", "user.enabled|-|system"}) {
			t.Errorf("disable/enable: %q", sys)
		}
		// Provider logout, confirmed and failed.
		for _, failRevoke := range []bool{false, true} {
			p := e.newUser(t)
			w.label(p.id.String(), "pv")
			w.label(p.session, "pvs")
			if failRevoke {
				e.fake.FailRevoke(500)
			}
			w.since(t)
			e.logout(t, p.cred, identity.ScopeProvider)
			if failRevoke {
				e.fake.FailRevoke(0)
			}
			got := w.rows(t, "subject_user_id = $2::text::uuid", p.id.String())
			note(got)
			outcome := "success"
			if failRevoke {
				outcome = "failed"
			}
			w.wantRows(t, fmt.Sprintf("provider logout (fail=%v)", failRevoke), got,
				"session.revoked|success|provider_logout|-|user|pv|-|-|session|pvs|pv|-|-|-",
				"session.provider_revocation_recorded|"+outcome+"|-|-|user|pv|-|-|session|pvs|pv|-|-|-")
		}
		// Current logout; idempotent repeat records nothing.
		c := e.newUser(t)
		w.label(c.id.String(), "c")
		w.label(c.session, "cs")
		w.since(t)
		e.logout(t, c.cred, identity.ScopeCurrent)
		if _, err := e.id.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, Origin: itOrigin, Credential: c.cred}); err != nil {
			t.Fatal(err)
		}
		w.wantRows(t, "current logout", w.rows(t, "subject_user_id = $2::text::uuid", c.id.String()),
			"session.revoked|success|logout|-|user|c|-|-|session|cs|c|-|-|-")
		// Post-consume failures: unattributed login_canceled; attributed user_disabled and identity_conflict. A
		// pre-consume refusal records nothing.
		o := e.newUser(t)
		w.label(o.id.String(), "o")
		w.since(t)
		if _, err := e.audCallback(t, e.id, "", "", "error=access_denied", ""); !errors.Is(err, identity.ErrLoginCanceled) {
			t.Fatalf("canceled login: %v", err)
		}
		if err := e.id.DisableUser(ctx, c.id.String()); err != nil {
			t.Fatal(err)
		}
		if _, err := e.audCallback(t, e.id, c.sub, fmt.Sprintf("session_01AU%08d", itSeq.Add(1)), "code", ""); !errors.Is(err, identity.ErrUserDisabled) {
			t.Fatalf("disabled sign-in: %v", err)
		}
		if _, err := e.audCallback(t, e.id, x.sub, fmt.Sprintf("session_01AU%08d", itSeq.Add(1)), "code", o.cred); !errors.Is(err, identity.ErrIdentityConflict) {
			t.Fatalf("conflicting sign-in: %v", err)
		}
		if _, err := e.id.Callback(ctx, identity.CallbackInput{Origin: itOrigin, RawQuery: "state=unknown&code=x", Binding: "bad"}); err == nil {
			t.Fatal("pre-consume refusal succeeded")
		}
		got = w.rows(t, "action = 'auth.sign_in'")
		note(got)
		w.wantRows(t, "post-consume failures", got,
			"auth.sign_in|denied|login_canceled|-|unknown|-|-|-|-|-|-|-|-|-",
			"auth.sign_in|denied|user_disabled|-|user|c|-|-|-|-|c|-|-|-",
			"auth.sign_in|denied|identity_conflict|-|user|o|-|-|-|-|o|-|-|-")
		_ = e.id.EnableUser(ctx, c.id.String())
		// Global rows are invisible to every tenant reader.
		rctx, otc := tc(t, e.osvc, w.owner, w.orgA, tenancy.GroupID{})
		page, err := e.osvc.ReadAuditEvents(rctx, otc, AuditReadRequest{Limit: 100})
		if err != nil {
			t.Fatal(err)
		}
		for _, ev := range page.Events {
			if ev.Organization != audit.UUID(w.orgA) {
				t.Errorf("tenant read returned an event of another scope (%s)", ev.Action)
			}
		}
		// A revocation on an unavailable (frozen) database is identity_unavailable, never success, with no G6.
		audFrozenRevocation(t, e)
		// A lost COMMIT reply on a guarded write: at most one event, no retry.
		h := &hookDB{pool: e.pool}
		s := NewService(h, e.id, e.svc.logger)
		lctx, ltc := tc(t, s, w.owner, w.orgA, tenancy.GroupID{})
		w.since(t)
		h.setCommit(commitLostReply)
		_, err = s.RenameGroup(lctx, ltc, w.g2.ID, fmt.Sprintf("Lost Reply %d", itSeq.Add(1)))
		h.setCommit(commitNormal)
		wantCode(t, "lost commit reply", err, CodeUnavailable)
		if got := w.rows(t, "action = 'group.renamed'"); len(got) > 1 {
			t.Errorf("lost reply: %d rename events", len(got))
		}
	})

	t.Run("AT-OBJ storage events", func(t *testing.T) {
		rctx, gtc := tc(t, e.osvc, w.gadmin, w.orgA, tenancy.GroupID{})
		w.since(t)
		info := audPut(t, e, w.gadmin, w.orgA, w.g1.ID, "obj")
		w.label(info.ID.String(), "o1")
		su, err := e.osvc.SignedGetURL(rctx, gtc, SignedGetURLRequest{Object: info.ID, TTLSeconds: 7})
		if err != nil {
			t.Fatal(err)
		}
		res, err := e.osvc.DeleteObject(rctx, gtc, info.ID)
		if err != nil || res.Outcome != DeleteOutcomeDeleted {
			t.Fatalf("delete: %+v %v", res, err)
		}
		got := w.rows(t, "resource_id = $2::text::uuid", info.ID.String())
		note(got)
		w.wantRows(t, "put/sign/delete", got,
			"object.upload_started|success|-|-|user|gadmin|A|g1|stored_object|o1|-|-|-|-",
			"object.ready|success|-|-|user|gadmin|A|g1|stored_object|o1|-|-|-|-",
			"object.url_issued|success|-|-|user|gadmin|A|g1|stored_object|o1|-|-|-|7",
			"object.delete_started|success|-|-|user|gadmin|A|g1|stored_object|o1|-|-|-|-",
			"object.deleted|success|-|-|user|gadmin|A|g1|stored_object|o1|-|-|-|-")
		// No URL, signature or key in any row.
		var leak int
		if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM audit.events e WHERE e::text ~* '(https?:|x-amz|signature|objects/)' OR e::text LIKE '%' || $1 || '%'`,
			su.URL()).Scan(&leak); err != nil || leak != 0 {
			t.Errorf("%d rows hold a URL, signature or key (%v)", leak, err)
		}
		// Issuance fault: no URL.
		info2 := audPut(t, e, w.gadmin, w.orgA, w.g1.ID, "obj2")
		remove := w.fault(t, "action <> 'object.url_issued'")
		_, err = e.osvc.SignedGetURL(rctx, gtc, SignedGetURLRequest{Object: info2.ID, DefaultTTL: true})
		remove()
		wantCode(t, "issuance fault", err, CodeUnavailable)
		if rec := e.lastStorage(t); rec["stage"] != "audit" {
			t.Errorf("issuance fault record %v", rec)
		}
		// Provider faults after T1 and on delete: upload_failed (compensated) and deletion_pending.
		f := e.faultService(e.ocfg)
		f.spy.before = func(op string, _ objectstore.Key) error {
			if op == "put" {
				return errors.New("injected provider put failure")
			}
			return nil
		}
		w.since(t)
		before := e.objectsOf(t, w.orgA)
		_, err = f.putAs(t, w.gadmin, w.orgA, w.g1.ID, 16, io.NopCloser(strings.NewReader("0123456789abcdef")))
		wantCode(t, "provider put fault", err, CodeUnavailable)
		id, _ := e.one(t, w.orgA, before)
		w.label(id.String(), "o3")
		f.spy.before = nil
		got = w.rows(t, "resource_id = $2::text::uuid", id.String())
		note(got)
		w.wantRows(t, "provider put fault", got,
			"object.upload_started|success|-|-|user|gadmin|A|g1|stored_object|o3|-|-|-|-",
			"object.upload_failed|failed|compensated|-|user|gadmin|A|g1|stored_object|o3|-|-|-|-")
		f.spy.before = func(op string, _ objectstore.Key) error {
			if op == "delete" {
				return errors.New("injected provider delete failure")
			}
			return nil
		}
		w.since(t)
		res, err = f.deleteAs(t, w.gadmin, w.orgA, info2.ID)
		f.spy.before = nil
		if err != nil || res.Outcome != DeleteOutcomeDeletionPending {
			t.Fatalf("delete under a provider fault: %+v %v", res, err)
		}
		w.label(info2.ID.String(), "o2")
		got = w.rows(t, "resource_id = $2::text::uuid", info2.ID.String())
		note(got)
		w.wantRows(t, "provider delete fault", got,
			"object.delete_started|success|-|-|user|gadmin|A|g1|stored_object|o2|-|-|-|-",
			"object.deletion_pending|failed|provider_unavailable|-|user|gadmin|A|g1|stored_object|o2|-|-|-|-")
		e.repairDelete(t, &objOrderWorld{admin: w.gadmin, org: w.orgA}, info2.ID)
		// T2 lost reply: reconcile finds the row ready (object.ready committed with T2), or records upload_failed.
		f = e.faultService(e.ocfg)
		f.hook.stage = func(op, stage string) {
			if op == opObjectPut && stage == stageT2 {
				f.h.setCommit(commitLostReply)
			}
			if op == opObjectPut && stage == stageReconcile {
				f.h.setCommit(commitNormal)
			}
		}
		w.since(t)
		before = e.objectsOf(t, w.orgA)
		_, err = f.putAs(t, w.gadmin, w.orgA, w.g1.ID, 4, io.NopCloser(strings.NewReader("abcd")))
		id, _ = e.one(t, w.orgA, before)
		w.label(id.String(), "o4")
		got = w.rows(t, "resource_id = $2::text::uuid", id.String())
		switch {
		case err == nil:
			w.wantRows(t, "T2 lost reply (reconciled ready)", got,
				"object.upload_started|success|-|-|user|gadmin|A|g1|stored_object|o4|-|-|-|-",
				"object.ready|success|-|-|user|gadmin|A|g1|stored_object|o4|-|-|-|-")
		default:
			if len(got) < 2 || !strings.HasPrefix(got[len(got)-1].text, "object.upload_failed|") {
				t.Errorf("T2 lost reply: %v, events %q", err, texts(got))
			}
		}
		e.cleanupOrgObjects(t, w.orgA)
	})

	t.Run("AT-READ owner-only reads", func(t *testing.T) {
		rctx, otc := tc(t, e.osvc, w.owner, w.orgA, tenancy.GroupID{})
		page, err := e.osvc.ReadAuditEvents(rctx, otc, AuditReadRequest{Limit: 100})
		if err != nil || len(page.Events) == 0 {
			t.Fatalf("owner read: %d events, %v", len(page.Events), err)
		}
		for _, u := range []*itUser{w.admin, w.member, w.gadmin} {
			uctx, utc := tc(t, e.osvc, u, w.orgA, tenancy.GroupID{})
			_, err := e.osvc.ReadAuditEvents(uctx, utc, AuditReadRequest{})
			wantCode(t, w.labels[u.id.String()]+" read", err, CodeForbidden)
		}
		// Group roles in selected-Group context read nothing; the owner in selected-Group context reads the whole
		// Organization; a shared profile confers nothing.
		gctx, gtc := tc(t, e.osvc, w.gadmin, w.orgA, w.g1.ID)
		_, err = e.osvc.ReadAuditEvents(gctx, gtc, AuditReadRequest{})
		wantCode(t, "group admin in g1", err, CodeForbidden)
		octx, ogtc := tc(t, e.osvc, w.owner, w.orgA, w.g1.ID)
		gpage, err := e.osvc.ReadAuditEvents(octx, ogtc, AuditReadRequest{Limit: 100})
		whole, werr := e.osvc.ReadAuditEvents(rctx, otc, AuditReadRequest{Limit: 100}) // successful reads append nothing
		if err != nil || werr != nil || fmt.Sprint(gpage.Events) != fmt.Sprint(whole.Events) {
			t.Errorf("owner in g1: %d events, want the whole Organization's %d (%v, %v)", len(gpage.Events), len(whole.Events), err, werr)
		}
		// The multi-tenant user: owner of B reads B only; member of A is forbidden in A.
		bctx, btc := tc(t, e.osvc, w.multi, w.orgB, tenancy.GroupID{})
		bpage, err := e.osvc.ReadAuditEvents(bctx, btc, AuditReadRequest{Limit: 100})
		if err != nil {
			t.Fatal(err)
		}
		for _, ev := range bpage.Events {
			if ev.Organization != audit.UUID(w.orgB) {
				t.Error("owner of B read an event of A")
			}
		}
		actx, atc := tc(t, e.osvc, w.multi, w.orgA, tenancy.GroupID{})
		_, err = e.osvc.ReadAuditEvents(actx, atc, AuditReadRequest{})
		wantCode(t, "member of A", err, CodeForbidden)
		// A cursor of another Organization is invalid_cursor.
		bfull, err := e.osvc.ReadAuditEvents(bctx, btc, AuditReadRequest{Limit: 1})
		if err != nil || !bfull.More {
			t.Fatalf("B first page: %v more=%v", err, bfull.More)
		}
		_, err = e.osvc.ReadAuditEvents(rctx, otc, AuditReadRequest{Limit: 1, Cursor: bfull.Next})
		if CodeOf(err) != CodeInvalidInput || err.(*Error).Field != audit.CodeInvalidCursor {
			t.Errorf("cross-tenant cursor: %v", err)
		}
		// Raw runtime SQL per context: exact visible sets.
		visible := func(u, org, group string) int {
			tx, err := e.pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true), set_config('app.group_id', $3, true)`,
				u, org, group); err != nil {
				t.Fatal(err)
			}
			var n int
			if err := tx.QueryRow(ctx, `SELECT count(*)::int FROM audit.events`).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		var ownA int
		if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM audit.events WHERE organization_id = $1::text::uuid`, w.orgA.String()).Scan(&ownA); err != nil {
			t.Fatal(err)
		}
		for _, c := range []struct {
			name, u, org, group string
			want                int
		}{
			{"owner A", w.owner.id.String(), w.orgA.String(), "", ownA},
			{"owner A in g1", w.owner.id.String(), w.orgA.String(), w.g1.ID.String(), ownA},
			{"admin A", w.admin.id.String(), w.orgA.String(), "", 0},
			{"member A", w.member.id.String(), w.orgA.String(), "", 0},
			{"group admin g1", w.gadmin.id.String(), w.orgA.String(), w.g1.ID.String(), 0},
			{"owner B in A context", w.multi.id.String(), w.orgA.String(), "", 0},
			{"identity only", w.owner.id.String(), "", "", 0},
			{"no context", "", "", "", 0},
		} {
			if got := visible(c.u, c.org, c.group); got != c.want {
				t.Errorf("runtime %s sees %d events, want %d", c.name, got, c.want)
			}
		}
		// Pooled reuse: one connection serves the owner, then the admin, with no carried authority.
		sp := e.singlePool(t)
		ss := NewService(sp, e.id, e.svc.logger)
		s1, t1 := tc(t, ss, w.owner, w.orgA, tenancy.GroupID{})
		if _, err := ss.ReadAuditEvents(s1, t1, AuditReadRequest{}); err != nil {
			t.Fatal(err)
		}
		s2, t2 := tc(t, ss, w.admin, w.orgA, tenancy.GroupID{})
		_, err = ss.ReadAuditEvents(s2, t2, AuditReadRequest{})
		wantCode(t, "admin after the owner on one connection", err, CodeForbidden)
		// Downgrade and revoked session.
		dn := e.newUser(t)
		e.grantOrg(t, w.orgA, dn, tenancy.OrgOwner)
		e.bind(t, dn, w.orgA) // G016 A2
		dctx, dtc := tc(t, e.osvc, dn, w.orgA, tenancy.GroupID{})
		if _, err := e.osvc.ReadAuditEvents(dctx, dtc, AuditReadRequest{}); err != nil {
			t.Fatal(err)
		}
		if _, err := e.model.ChangeOrganizationRole(ctx, w.orgA, dn.id, tenancy.OrgAdmin); err != nil {
			t.Fatal(err)
		}
		_, err = e.osvc.ReadAuditEvents(dctx, dtc, AuditReadRequest{})
		wantCode(t, "downgraded owner", err, CodeForbidden)
		e.logout(t, dn.in(t, w.orgA), identity.ScopeCurrent) // G016 A2: revokes the credential of the scope under test
		_, err = e.osvc.ReadAuditEvents(dctx, dtc, AuditReadRequest{})
		wantCode(t, "revoked session", err, CodeUnauthenticated)
	})

	t.Run("AT-LIM bounds, cursor binding and continuation", func(t *testing.T) {
		rctx, otc := tc(t, e.osvc, w.owner, w.orgA, tenancy.GroupID{})
		for _, c := range []struct {
			req   AuditReadRequest
			field string
		}{
			{AuditReadRequest{Limit: -1}, audit.CodeInvalidPageSize},
			{AuditReadRequest{Limit: 101}, audit.CodeInvalidPageSize},
			{AuditReadRequest{Since: time.Now(), Until: time.Now().Add(-time.Hour)}, audit.CodeInvalidTimeRange},
		} {
			_, err := e.osvc.ReadAuditEvents(rctx, otc, c.req)
			if CodeOf(err) != CodeInvalidInput || err.(*Error).Field != c.field {
				t.Errorf("%+v: %v, want %s", c.req, err, c.field)
			}
		}
		first, err := e.osvc.ReadAuditEvents(rctx, otc, AuditReadRequest{Limit: 2})
		if err != nil || len(first.Events) != 2 || !first.More {
			t.Fatalf("first page: %d %v %v", len(first.Events), first.More, err)
		}
		if _, err := e.osvc.ReadAuditEvents(rctx, otc, AuditReadRequest{Limit: 2, Since: time.Unix(1, 0), Cursor: first.Next}); CodeOf(err) != CodeInvalidInput ||
			err.(*Error).Field != audit.CodeInvalidCursor {
			t.Errorf("cursor of another window: %v", err)
		}
		def, err := e.osvc.ReadAuditEvents(rctx, otc, AuditReadRequest{})
		if err != nil || len(def.Events) > defaultAuditPage {
			t.Errorf("default page: %d %v", len(def.Events), err)
		}
		// Deterministic continuation under concurrent appends: strictly decreasing (time, id), no duplicates.
		stop := make(chan struct{})
		var appends atomic.Int64
		var wg sync.WaitGroup
		mctx, mtc := tc(t, e.osvc, w.gadmin, w.orgA, tenancy.GroupID{})
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				_, _ = e.osvc.SetOrganizationPolicy(mctx, mtc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: true})
				appends.Add(1)
			}
		}()
		seen := map[audit.UUID]bool{}
		var prev *AuditEvent
		cur := AuditCursor{}
		for pages := 0; pages < 200; pages++ {
			p, err := e.osvc.ReadAuditEvents(rctx, otc, AuditReadRequest{Limit: 3, Cursor: cur})
			if err != nil {
				t.Fatal(err)
			}
			for i := range p.Events {
				ev := p.Events[i]
				if seen[ev.ID] {
					t.Fatalf("duplicate event across pages")
				}
				seen[ev.ID] = true
				if prev != nil && !(ev.OccurredAt.Before(prev.OccurredAt) || (ev.OccurredAt.Equal(prev.OccurredAt) && bytesLess(ev.ID[:], prev.ID[:]))) {
					t.Fatalf("keys are not strictly decreasing")
				}
				prev = &ev
			}
			if !p.More {
				break
			}
			cur = p.Next
		}
		close(stop)
		wg.Wait()
		if appends.Load() == 0 {
			t.Error("no concurrent append ran")
		}
		// E-AUDIT-1: statement and transaction counts per mechanism (descriptive).
		audFootprint(t, e, w)
		// E-AUDIT-2: the tenant read plan on the two-tenant fixture.
		audExplain(t, e, w)
	})

	// G020 (Plan r2 §3 step 5, T8, AC6): the five trigger-owned label actions through the real facade. Every row names the
	// resource label and the definition id, the Organization, and as group_id the owning Group (a Group-owned definition)
	// or the target Group (an assignment); no row carries a label name. label.updated carries renamed or
	// visibility_changed; a no-op update records nothing; a definition deletion records label.unassigned per assignment and
	// then label.deleted in one transaction; a planted audit fault refuses the creation and the assignment with no effect.
	t.Run("AT-LAB label producers: attribution, reasons, no-op silence, one-transaction deletion and audit faults", func(t *testing.T) {
		zeroG := tenancy.GroupID{}
		octx, otc := tc(t, e.osvc, w.owner, w.orgA, zeroG)
		gctx, gtc := tc(t, e.osvc, w.gadmin, w.orgA, zeroG)
		orgSc := Scope{Kind: tenancy.ScopeOrganization}
		g1Sc := Scope{Kind: tenancy.ScopeGroup, Group: &w.g1.ID}
		on := func(g tenancy.GroupID) LabelTargetSelector { return LabelTargetSelector{Group: g.String()} }
		labelName := func(what string) string { return fmt.Sprintf("Audit %s label %d", what, itSeq.Add(1)) }
		owner, renamed, group := labelName("organization"), labelName("renamed"), labelName("group")
		w.since(t)
		lo, err := e.osvc.CreateLabelDefinition(octx, otc, orgSc, owner, string(tenancy.LabelVisibilityOwnerScope))
		if err != nil {
			t.Fatalf("owner creates an Organization label: %v", err)
		}
		w.label(lo.ID.String(), "lo")
		lg, err := e.osvc.CreateLabelDefinition(gctx, gtc, g1Sc, group, string(tenancy.LabelVisibilityOwnerScope))
		if err != nil {
			t.Fatalf("g1 admin creates a Group label: %v", err)
		}
		w.label(lg.ID.String(), "lg")
		if _, err := e.osvc.UpdateLabelDefinition(octx, otc, lo.ID.String(), renamed, string(tenancy.LabelVisibilityOwnerScope)); err != nil {
			t.Fatalf("rename: %v", err)
		}
		if _, err := e.osvc.UpdateLabelDefinition(gctx, gtc, lg.ID.String(), group, string(tenancy.LabelVisibilityTargetMembers)); err != nil {
			t.Fatalf("visibility change: %v", err)
		}
		// No-op updates (the current values) succeed and record nothing.
		if _, err := e.osvc.UpdateLabelDefinition(gctx, gtc, lg.ID.String(), group, string(tenancy.LabelVisibilityTargetMembers)); err != nil {
			t.Fatalf("no-op update of lg: %v", err)
		}
		if _, err := e.osvc.UpdateLabelDefinition(octx, otc, lo.ID.String(), renamed, string(tenancy.LabelVisibilityOwnerScope)); err != nil {
			t.Fatalf("no-op update of lo: %v", err)
		}
		for _, g := range []tenancy.GroupID{w.g1.ID, w.g2.ID} {
			if err := e.osvc.AssignLabel(octx, otc, lo.ID.String(), on(g)); err != nil {
				t.Fatalf("assign: %v", err)
			}
		}
		if err := e.osvc.UnassignLabel(octx, otc, lo.ID.String(), on(w.g2.ID)); err != nil {
			t.Fatalf("unassign: %v", err)
		}
		got := w.rows(t, "action LIKE 'label.%'")
		note(got)
		w.wantRows(t, "label producers", got,
			"label.created|success|-|-|user|owner|A|-|label|lo|-|-|-|-",
			"label.created|success|-|-|user|gadmin|A|g1|label|lg|-|-|-|-",
			"label.updated|success|renamed|-|user|owner|A|-|label|lo|-|-|-|-",
			"label.updated|success|visibility_changed|-|user|gadmin|A|g1|label|lg|-|-|-|-",
			"label.assigned|success|-|-|user|owner|A|g1|label|lo|-|-|-|-",
			"label.assigned|success|-|-|user|owner|A|g2|label|lo|-|-|-|-",
			"label.unassigned|success|-|-|user|owner|A|g2|label|lo|-|-|-|-")
		// An audit fault on the label rows only: the creation and the assignment fail with neither the row nor an event.
		w.since(t)
		faulted := labelName("faulted")
		remove := w.fault(t, "action NOT LIKE 'label.%'")
		_, cerr := e.osvc.CreateLabelDefinition(octx, otc, orgSc, faulted, string(tenancy.LabelVisibilityOwnerScope))
		aerr := e.osvc.AssignLabel(octx, otc, lo.ID.String(), on(w.g2.ID))
		remove()
		if cerr == nil || aerr == nil {
			t.Errorf("under a label audit fault: create %v, assign %v; want both refused", cerr, aerr)
		}
		var defs, asg int
		if err := e.setup.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.label_definitions WHERE organization_id = $1::text::uuid AND name = $2)::int,
  (SELECT count(*) FROM app.label_assignments WHERE label_id = $3::text::uuid AND target_group_id = $4::text::uuid)::int`,
			w.orgA.String(), faulted, lo.ID.String(), w.g2.ID.String()).Scan(&defs, &asg); err != nil {
			t.Fatal(err)
		}
		if defs != 0 || asg != 0 {
			t.Errorf("the refused writes left %d definitions and %d assignments", defs, asg)
		}
		if got := w.rows(t, "action LIKE 'label.%'"); len(got) != 0 {
			t.Errorf("label events under the fault: %q", texts(got))
		}
		// Paired positive after the fault: the same assignment commits with its event.
		if err := e.osvc.AssignLabel(octx, otc, lo.ID.String(), on(w.g2.ID)); err != nil {
			t.Fatalf("assign after the fault: %v", err)
		}
		w.wantRows(t, "assign after the fault", w.rows(t, "action LIKE 'label.%'"),
			"label.assigned|success|-|-|user|owner|A|g2|label|lo|-|-|-|-")
		// Deletion: the Organization owner (Organization-wide context) deletes lo with its two assignments; the two
		// label.unassigned rows (any order) and then label.deleted share one transaction. The Group admin deletes lg.
		w.since(t)
		res, err := e.osvc.DeleteLabelDefinition(octx, otc, lo.ID.String())
		if err != nil || res.AssignmentsRemoved != 2 {
			t.Fatalf("delete lo: %+v %v, want 2 assignments removed", res, err)
		}
		del := w.rows(t, "action LIKE 'label.%'")
		note(del)
		if len(del) != 3 {
			t.Fatalf("deletion events: %q", texts(del))
		}
		unassigned := texts(del[:2])
		slices.Sort(unassigned)
		if want := []string{"label.unassigned|success|-|-|user|owner|A|g1|label|lo|-|-|-|-",
			"label.unassigned|success|-|-|user|owner|A|g2|label|lo|-|-|-|-"}; !slices.Equal(unassigned, want) ||
			del[2].text != "label.deleted|success|-|-|user|owner|A|-|label|lo|-|-|-|-" {
			t.Errorf("deletion events %q, want both label.unassigned then label.deleted", texts(del))
		}
		if del[0].xmin != del[2].xmin || del[1].xmin != del[2].xmin {
			t.Error("the deletion events are not one transaction")
		}
		w.since(t)
		if res, err := e.osvc.DeleteLabelDefinition(gctx, gtc, lg.ID.String()); err != nil || res.AssignmentsRemoved != 0 {
			t.Fatalf("delete lg: %+v %v", res, err)
		}
		got = w.rows(t, "action LIKE 'label.%'")
		note(got)
		w.wantRows(t, "Group label deletion", got, "label.deleted|success|-|-|user|gadmin|A|g1|label|lg|-|-|-|-")
		// No label name in any audit row of the run.
		var leaks int
		if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM audit.events e
  WHERE position($1 in e::text) > 0 OR position($2 in e::text) > 0 OR position($3 in e::text) > 0 OR position($4 in e::text) > 0`,
			owner, renamed, group, faulted).Scan(&leaks); err != nil {
			t.Fatal(err)
		}
		if leaks != 0 {
			t.Errorf("%d audit rows carry a label name", leaks)
		}
	})

	t.Run("AT-INV every catalog action observed from real calls", func(t *testing.T) {
		rctx, otc := tc(t, e.osvc, w.owner, w.orgA, tenancy.GroupID{})
		w.since(t)
		// The remaining producers: Organization rename and policy, membership change and removal, profile update and
		// unlink, and slug change.
		if _, err := e.model.RenameOrganization(ctx, w.orgB, "Audit B Renamed"); err != nil {
			t.Fatal(err)
		}
		if _, err := e.osvc.SetOrganizationPolicy(rctx, otc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false}); err != nil {
			t.Fatal(err)
		}
		if _, err := e.osvc.ChangeGroupSlug(rctx, otc, w.g2.ID, fmt.Sprintf("audit-slug-%d", itSeq.Add(1))); err != nil {
			t.Fatal(err)
		}
		// G060 (Plan r0 A4, r1 C5): the Organization owner updates the profile through the Group-management override, which
		// also observes the new core action access.override_used (atomic with the write); the override is then disabled.
		if _, err := e.osvc.SetGroupManagementOverride(rctx, otc, true); err != nil {
			t.Fatal(err)
		}
		if _, err := e.osvc.UpdateBusinessProfile(rctx, otc, w.prof.ID, tenancy.BusinessContext{DisplayName: "Audit Profile Updated"}); err != nil {
			t.Fatal(err)
		}
		if _, err := e.osvc.SetGroupManagementOverride(rctx, otc, false); err != nil {
			t.Fatal(err)
		}
		if _, err := e.model.ChangeOrganizationRole(ctx, w.orgA, w.member.id, tenancy.OrgAdmin); err != nil {
			t.Fatal(err)
		}
		if err := e.model.RemoveGroupMembership(ctx, w.orgA, w.g1.ID, w.member.id); err != nil {
			t.Fatal(err)
		}
		if err := e.model.UnlinkGroupBusinessProfile(ctx, w.orgA, w.g1.ID, w.prof.ID); err != nil {
			t.Fatal(err)
		}
		// G014: an MCP connection's activation and Group selection (mcp.connection_established, mcp.context_changed)
		// through the facade, from setup to the verified scope.
		connID, err := identity.NewMCPConnectionID()
		if err != nil {
			t.Fatal(err)
		}
		host := "org-a.localhost:543"
		if err := e.osvc.CreateMCPConnection(rctx, otc, connID, host, "https://"+host+"/mcp/c/"+connID,
			identity.MCPLifetimes{Pending: 30 * time.Minute, Idle: 24 * time.Hour, Absolute: 48 * time.Hour}, nil); err != nil {
			t.Fatal(err)
		}
		mconn, err := e.id.ReadMCPConnection(ctx, w.owner.id.String(), w.orgA.String(), connID)
		if err != nil {
			t.Fatal(err)
		}
		mctx, mend, err := e.osvc.BeginVerifiedRequest(ctx, mconn)
		if err != nil {
			t.Fatal(err)
		}
		if err := e.osvc.ActivateConnection(mctx, "client_01AUDITINVENTORY"); err != nil {
			t.Fatal(err)
		}
		if _, err := e.osvc.SelectConnectionGroup(mctx, w.g2.ID.String()); err != nil {
			t.Fatal(err)
		}
		mend()
		// G063: the owner names the active connection through the setup facade (mcp.connection_renamed; no name in the row).
		w.label(connID, "invconn")
		invName := "Audit inventory laptop"
		if err := e.osvc.RenameMCPConnection(rctx, otc, connID, &invName); err != nil {
			t.Fatal(err)
		}
		w.wantRows(t, "connection rename", w.rows(t, "action = 'mcp.connection_renamed'"),
			"mcp.connection_renamed|success|-|-|user|owner|A|-|mcp_connection|invconn|owner|-|-|-")
		// G016: the Organization slug change and a custom domain's claim, check and state change through the facade.
		if _, err := e.osvc.SetOrganizationSlug(rctx, otc, fmt.Sprintf("audit-org-%d", itSeq.Add(1))); err != nil {
			t.Fatal(err)
		}
		dc, err := e.osvc.ClaimDomain(rctx, otc, fmt.Sprintf("audit-%d.acme-domains.example", itSeq.Add(1)), time.Hour, "_platform-challenge")
		if err != nil {
			t.Fatal(err)
		}
		dr, err := e.osvc.BeginDomainCheck(rctx, otc, dc.ID, domains.CheckVerify)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.osvc.RecordDomainCheck(rctx, otc, dr, domains.CheckVerify, domains.Observation{Proof: domains.OutcomeOK}, 24*time.Hour); err != nil {
			t.Fatal(err)
		}
		// G017: one guarded branding write by the Organization owner (organization.branding_changed).
		if _, err := e.osvc.UpdateBranding(rctx, otc, BrandingUpdate{Label: strp("Audit Inventory")}); err != nil {
			t.Fatal(err)
		}
		// G058 slice 1b (addendum §A.2, correction 01 §C.2): the three trigger-owned enrollment producers. The owner's
		// guarded policy write (enrollment.policy_changed, the owner's); then the superuser operations channel, each call
		// its own audited procedure with the operator reference as request_id and actor unknown: close A's readiness
		// (enrollment.readiness_changed history_unresolved), backfill one departure of a never-bound user
		// (enrollment.restriction_backfilled; readiness is already closed, so no second readiness row) and attest the
		// backfilled history (enrollment.readiness_changed operator_backfilled).
		if _, err := e.osvc.UpdateEnrollmentPolicy(rctx, otc, tenancy.EnrollmentRequestAccess); err != nil {
			t.Fatal(err)
		}
		nonmember := e.newUser(t)
		w.label(nonmember.id.String(), "nonmember")
		for _, c := range []struct{ q, want string }{
			{`SELECT app.enrollment_readiness_set($1::text::uuid, NULL, 'audit-inv.close')`, "closed"},
			{`SELECT app.enrollment_history_backfill($1::text::uuid, $2::text::uuid, 'l1.audit-inv')`, "inserted"},
			{`SELECT app.enrollment_readiness_set($1::text::uuid, 'operator_backfilled', 'l2.audit-inv')`, "resolved"},
		} {
			args := []any{w.orgA.String()}
			if strings.Contains(c.q, "$2") {
				args = append(args, nonmember.id.String())
			}
			var got string
			if err := e.setup.QueryRow(ctx, c.q, args...).Scan(&got); err != nil || got != c.want {
				t.Fatalf("%s: %q %v, want %s", c.q, got, err, c.want)
			}
		}
		w.wantRows(t, "enrollment producers", w.rows(t, "action LIKE 'enrollment.%'"),
			"enrollment.policy_changed|success|-|-|user|owner|A|-|organization|A|-|-|-|-",
			"enrollment.readiness_changed|success|history_unresolved|-|unknown|-|A|-|organization|A|-|-|-|-",
			"enrollment.restriction_backfilled|success|operator_backfilled|-|unknown|-|A|-|organization|A|nonmember|-|-|-",
			"enrollment.readiness_changed|success|operator_backfilled|-|unknown|-|A|-|organization|A|-|-|-|-")
		// G058 request slice: the three request producers through the real facades (A is request_access since the policy
		// write above). The nonmember applies with its own enrollment credential (enrollment.requested, the applicant's,
		// appended by the claim writer inside enrollment_request; its backfilled departure does not refuse a request), the
		// owner approves (enrollment.decided, reason NULL, the decider's) and then revokes the unconsumed approval
		// (enrollment.revoked, the revoker's); each names the request as its resource and the applicant as its subject.
		req, err := e.osvc.RequestEnrollment(ctx, e.enrHandle(t, nonmember.id, w.orgA))
		if err != nil {
			t.Fatalf("enrollment request: %v", err)
		}
		w.label(req.ID.String(), "req1")
		if _, err := e.osvc.DecideEnrollmentRequest(rctx, otc, req.ID, DecisionApprove); err != nil {
			t.Fatalf("approve: %v", err)
		}
		if _, err := e.osvc.RevokeEnrollmentApproval(rctx, otc, req.ID); err != nil {
			t.Fatalf("revoke approval: %v", err)
		}
		w.wantRows(t, "request producers", w.rows(t, "action IN ('enrollment.requested', 'enrollment.decided', 'enrollment.revoked')"),
			"enrollment.requested|success|-|-|user|nonmember|A|-|enrollment_request|req1|nonmember|-|-|-",
			"enrollment.decided|success|-|-|user|owner|A|-|enrollment_request|req1|nonmember|-|-|-",
			"enrollment.revoked|success|-|-|user|owner|A|-|enrollment_request|req1|nonmember|-|-|-")
		// G019 (Plan r1 §6.2 "AT-INV (dynamic) extend"): the invitation producers through the real facade. The owner
		// (group_owner of g2) issues and revokes one Group invitation (gadmin), issues a second one that its intended,
		// enabled Organization member accepts in the member's own request scope (invitation.consumed with the binding in
		// one transaction), and issues and revokes one address invitation. Every row is attributed to its actor, scope
		// and subject; no token or address reaches the ledger.
		issue := func(sc Scope, role string, r Recipient, key string) IssuedInvitation {
			t.Helper()
			iss, err := e.osvc.IssueInvitation(rctx, otc, sc, role, r, 0)
			if err != nil {
				t.Fatalf("issue %s: %v", key, err)
			}
			w.label(iss.Invitation.ID.String(), key)
			return iss
		}
		inv1 := issue(Scope{Kind: tenancy.ScopeGroup, Group: &w.g2.ID}, string(tenancy.GroupMember), Recipient{UserID: &w.gadmin.id}, "inv1")
		if err := e.osvc.RevokeInvitation(rctx, otc, inv1.Invitation.ID); err != nil {
			t.Fatal(err)
		}
		inv2 := issue(Scope{Kind: tenancy.ScopeGroup, Group: &w.g2.ID}, string(tenancy.GroupMember), Recipient{UserID: &w.member.id}, "inv2")
		mctx, mtc := tc(t, e.osvc, w.member, w.orgA, tenancy.GroupID{})
		out, err := e.osvc.AcceptGroupInvitation(mctx, mtc, inv2.Invitation.ID, inv2.Token)
		if err != nil || out.Outcome != "joined" {
			t.Fatalf("accept inv2: outcome %q, %v", out.Outcome, err)
		}
		addr3 := memAddr("audit-inv")
		inv3 := issue(Scope{Kind: tenancy.ScopeOrganization}, string(tenancy.OrgMember), Recipient{Address: &addr3}, "inv3")
		if err := e.osvc.RevokeInvitation(rctx, otc, inv3.Invitation.ID); err != nil {
			t.Fatal(err)
		}
		invRows := w.rows(t, "action LIKE 'invitation.%'")
		w.wantRows(t, "invitation producers", invRows,
			"invitation.issued|success|-|-|user|owner|A|g2|invitation|inv1|gadmin|group_member|-|-",
			"invitation.revoked|success|-|-|user|owner|A|g2|invitation|inv1|gadmin|group_member|-|-",
			"invitation.issued|success|-|-|user|owner|A|g2|invitation|inv2|member|group_member|-|-",
			"invitation.consumed|success|-|-|user|member|A|g2|invitation|inv2|member|group_member|-|-",
			"invitation.issued|success|-|-|user|owner|A|-|invitation|inv3|-|org_member|-|-",
			"invitation.revoked|success|-|-|user|owner|A|-|invitation|inv3|-|org_member|-|-")
		// The consumption and the granted binding share one transaction (Spec R4; MEM-10's shape for the Group path).
		granted := w.rows(t, "action = 'membership.granted' AND subject_user_id = $2::text::uuid", w.member.id.String())
		if len(granted) != 1 || len(invRows) != 6 {
			t.Fatalf("membership.granted rows for member: %q", texts(granted))
		}
		if !strings.HasPrefix(granted[0].text, "membership.granted|success|-|-|user|member|A|g2|") || granted[0].xmin != invRows[3].xmin {
			t.Errorf("membership.granted %q (xmin %s) is not the consumption's transaction (xmin %s)", granted[0].text, granted[0].xmin, invRows[3].xmin)
		}
		var leaks int
		if err := e.setup.QueryRow(ctx, `SELECT count(*) FROM audit.events e WHERE occurred_at > $1
  AND (position($2 in e::text) > 0 OR position($3 in e::text) > 0 OR position($4 in e::text) > 0 OR position($5 in e::text) > 0)`,
			w.mark, inv1.Token, inv2.Token, inv3.Token, addr3).Scan(&leaks); err != nil {
			t.Fatal(err)
		}
		if leaks != 0 {
			t.Errorf("%d audit rows carry an invitation token or address", leaks)
		}
		note(w.rows(t, ""))
		// The fixtures created before this subtest's windows (Organizations, Groups, grants, profiles, links, users
		// and sessions) are part of the same run: count them from their Organization and subjects.
		w.mark = time.Time{}
		note(w.rows(t, "organization_id::text = ANY($2::text[]) OR subject_user_id::text = ANY($3::text[])",
			[]string{w.orgA.String(), w.orgB.String()}, []string{w.owner.id.String(), w.member.id.String()}))
		// G020 (Result S0.2): the catalog is 52 actions, 45 of them trigger-owned. The five label.* actions are observed
		// from AT-LAB's real facade calls in A (noted there and again by the A-wide read above). Label rows that other
		// tests of this package commit into the shared database (TestRLSIsolation's and TestIsolationAttacks' superuser
		// fixtures) belong to their own Organizations and stay outside both reads, so they neither add nor hide an action.
		if nt, na := len(audit.TriggerActions), len(audit.TriggerActions)+len(audit.CoreActions); nt != 45 || na != 52 {
			t.Errorf("catalog: %d trigger actions of %d, want 45 of 52", nt, na)
		}
		var obs []string
		for a := range observed {
			obs = append(obs, a)
		}
		slices.Sort(obs)
		var missing []string
		for _, a := range append(slices.Clone(audit.TriggerActions), audit.CoreActions...) {
			if !observed[string(a)] {
				missing = append(missing, string(a))
			}
		}
		b, _ := json.Marshal(map[string]int{"actions": len(audit.TriggerActions) + len(audit.CoreActions), "observed": len(obs) - len(extra(obs))})
		fmt.Println("AUDIT-INV " + string(b))
		if len(missing) != 0 {
			t.Errorf("actions never observed: %v", missing)
		}
	})

	t.Run("AT-PRIV no canary, no deletion path", func(t *testing.T) {
		var rows int
		if err := e.setup.QueryRow(ctx, `SELECT count(*)::int FROM audit.events e WHERE `+canaryPredicate(len(e.canaries)), anyArgs(e.canaries)...).Scan(&rows); err != nil || rows != 0 {
			t.Errorf("%d audit rows hold a canary (%v)", rows, err)
		}
		rctx, otc := tc(t, e.osvc, w.owner, w.orgA, tenancy.GroupID{})
		page, err := e.osvc.ReadAuditEvents(rctx, otc, AuditReadRequest{Limit: 100})
		if err != nil {
			t.Fatal(err)
		}
		dump := fmt.Sprintf("%+v", page)
		for _, c := range e.canaries {
			if strings.Contains(dump, c) {
				t.Error("ReadAuditEvents output holds a canary")
			}
		}
		// The scan detects a planted canary (a superuser row with the canary as its request correlation, rolled back).
		tx, err := e.setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `SELECT set_config('audit.request_id', $1, true)`, e.canaries[0]); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO audit.events (action, outcome, reason) VALUES ('auth.sign_in', 'failed', 'provider_error')`); err != nil {
			t.Fatal(err)
		}
		if err := tx.QueryRow(ctx, `SELECT count(*)::int FROM audit.events e WHERE `+canaryPredicate(len(e.canaries)), anyArgs(e.canaries)...).Scan(&rows); err != nil || rows != 1 {
			t.Errorf("the planted canary was not detected (%d, %v)", rows, err)
		}
		_ = tx.Rollback(ctx)
		var del bool
		if err := e.setup.QueryRow(ctx, `SELECT has_table_privilege($1, 'audit.events', 'DELETE') OR has_table_privilege($1, 'audit.events', 'UPDATE')
  OR has_table_privilege($1, 'audit.events', 'TRUNCATE')`, e.cfg.RuntimeRole()).Scan(&del); err != nil || del {
			t.Errorf("runtime holds a deletion path (%v)", err)
		}
	})
}

// extra lists observed names that are not catalog actions (none expected).
func extra(obs []string) []string {
	var out []string
	for _, a := range obs {
		if !slices.Contains(audit.TriggerActions, audit.Action(a)) && !slices.Contains(audit.CoreActions, audit.Action(a)) {
			out = append(out, a)
		}
	}
	return out
}

func canaryPredicate(n int) string {
	var parts []string
	for i := 1; i <= n; i++ {
		parts = append(parts, fmt.Sprintf("e::text LIKE '%%' || $%d || '%%'", i))
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func anyArgs(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

func bytesLess(a, b []byte) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// audPut uploads a small object through the facade as u (G016 A2: in a request scope of u's credential for org).
func audPut(t *testing.T, e *objEnv, u *itUser, org tenancy.OrganizationID, g tenancy.GroupID, seed string) ObjectInfo {
	t.Helper()
	rctx, end := beginAs(t, e.osvc, context.Background(), u, org)
	defer end()
	tc := resolve(t, e.osvc, rctx, org, tenancy.GroupID{})
	data := payload(64, seed)
	info, err := e.osvc.PutObject(rctx, tc, PutObjectRequest{Group: g, Size: int64(len(data)), Body: io.NopCloser(strings.NewReader(string(data)))})
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	return info
}

// cleanupOrgObjects removes the remaining objects of org (rows and keys) through the fixtures.
func (e *objEnv) cleanupOrgObjects(t *testing.T, org tenancy.OrganizationID) {
	t.Helper()
	e.cleanupOrg(t, org)
}

// newUserMaybe reports whether a new sign-in succeeds (the privilege-fault strict check).
func (e *itEnv) newUserMaybe(t *testing.T) bool {
	t.Helper()
	_, err := e.audCallback(t, e.id, fmt.Sprintf("user_01AUDIT%08d", itSeq.Add(1)), fmt.Sprintf("session_01AU%08d", itSeq.Add(1)), "code", "")
	return err == nil
}

// audStalledRecorder: the identity recorder's rollback is frozen after its failed insert; the call returns within the
// cleanup budget and the connection is destroyed (never reused).
func audStalledRecorder(t *testing.T, e *objEnv) {
	t.Helper()
	ctx := context.Background()
	fr := newFreezer(t, fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port))
	defer fr.close()
	host, port := fr.addr()
	var freezeNext atomic.Bool
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = host, port, e.cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = e.cfg.RuntimeRole(), e.pw
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
	pc.ConnConfig.Tracer = &freezeTracer{after: "INSERT INTO audit.events", fr: fr, armed: &freezeNext}
	pc.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	clk := newITClock()
	id := e.identity(t, clk, pool)
	w := &audWorld{e: e, labels: map[string]string{}}
	remove := w.fault(t, "action <> 'auth.sign_in'")
	defer remove()
	var pid1 int
	if err := pool.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid1); err != nil {
		t.Fatal(err)
	}
	freezeNext.Store(true)
	start := time.Now()
	_, err = e.audCallback(t, id, "", "", "error=access_denied", "")
	took := time.Since(start)
	fr.thaw()
	if !errors.Is(err, identity.ErrLoginCanceled) {
		t.Errorf("stalled recorder changed the outcome: %v", err)
	}
	if took > 2*time.Second+2*time.Second+3*time.Second {
		t.Errorf("stalled recorder took %v, want within the record and cleanup budgets", took)
	}
	var pid2 int
	if err := pool.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid2); err != nil {
		t.Fatal(err)
	}
	if pid2 == pid1 {
		t.Error("the stalled connection was reused")
	}
}

// freezeTracer freezes the forwarder after the first failed statement whose SQL starts with after, once armed.
type freezeTracer struct {
	after string
	fr    *freezer
	armed *atomic.Bool
}

type freezeKey struct{}

func (f *freezeTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, freezeKey{}, strings.HasPrefix(d.SQL, f.after))
}

func (f *freezeTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryEndData) {
	if match, _ := ctx.Value(freezeKey{}).(bool); match && d.Err != nil && f.armed.CompareAndSwap(true, false) {
		f.fr.freeze()
	}
}

// audFrozenRevocation: a logout against a frozen database fails as identity_unavailable within its deadline, with no
// revocation.unaudited and no success.
func audFrozenRevocation(t *testing.T, e *objEnv) {
	t.Helper()
	fr := newFreezer(t, fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port))
	defer fr.close()
	host, port := fr.addr()
	pc, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	pc.ConnConfig.Host, pc.ConnConfig.Port, pc.ConnConfig.Database = host, port, e.cfg.Database
	pc.ConnConfig.User, pc.ConnConfig.Password = e.cfg.RuntimeRole(), e.pw
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pc.ConnConfig.TLSConfig, pc.ConnConfig.Fallbacks = nil, nil
	pool, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	id := e.identity(t, e.clk, pool)
	u := e.newUser(t)
	csrf := e.csrf(t, u.cred)
	info, err := id.Session(context.Background(), u.cred, itOrigin)
	if err != nil {
		t.Fatal(err)
	}
	w := &audWorld{e: e, labels: map[string]string{}}
	w.since(t)
	fr.freeze()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, err = id.Logout(ctx, identity.LogoutInput{Scope: identity.ScopeCurrent, CSRF: csrf, Origin: itOrigin, Credential: u.cred})
	cancel()
	fr.thaw()
	if err == nil {
		t.Error("a logout on a frozen database reported success")
	}
	if got := w.rows(t, "action = 'revocation.unaudited' AND subject_user_id = $2::text::uuid", info.UserID); len(got) != 0 {
		t.Errorf("revocation.unaudited after a failed revocation: %q", texts(got))
	}
}

// audFootprint prints E-AUDIT-1: statement and transaction counts of one operation per mechanism, and the revocation
// fan-out as one statement-level insert of N rows.
func audFootprint(t *testing.T, e *objEnv, w *audWorld) {
	t.Helper()
	ctx := context.Background()
	h := &hookDB{pool: e.pool}
	s := NewServiceWithObjects(h, e.id, e.svc.logger, e.spy, e.ocfg)
	line := func(fields map[string]any) {
		fields["evidence"] = "E-AUDIT-1"
		b, _ := json.Marshal(fields)
		fmt.Println("E-AUDIT-1 " + string(b))
	}
	octx, otc := func() (context.Context, *TenantContext) {
		rctx, _ := beginAs(t, s, ctx, w.owner, w.orgA) // G016 A2
		return rctx, resolve(t, s, rctx, w.orgA, tenancy.GroupID{})
	}()
	mctx, mtc := func() (context.Context, *TenantContext) {
		rctx, _ := beginAs(t, s, ctx, w.gadmin, w.orgA) // G016 A2
		return rctx, resolve(t, s, rctx, w.orgA, tenancy.GroupID{})
	}()
	h.reset()
	if _, err := s.RenameGroup(octx, otc, w.g2.ID, fmt.Sprintf("Footprint %d", itSeq.Add(1))); err != nil {
		t.Fatal(err)
	}
	b, st := h.counts()
	line(map[string]any{"mechanism": "M1 trigger (guarded rename)", "transactions": b, "statements": st})
	h.reset()
	_, _ = s.SetOrganizationPolicy(mctx, mtc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: true})
	b, st = h.counts()
	line(map[string]any{"mechanism": "M3 independent (denial)", "transactions": b, "statements": st})
	if b != 2 {
		eAuditTrigger(map[string]any{"mechanism": "M3", "reason": "a denial did not take exactly one extra transaction", "transactions": b})
	}
	info := audPut(t, e, w.gadmin, w.orgA, w.g1.ID, "fp")
	h.reset()
	if _, err := s.SignedGetURL(mctx, mtc, SignedGetURLRequest{Object: info.ID, DefaultTTL: true}); err != nil {
		t.Fatal(err)
	}
	b, st = h.counts()
	line(map[string]any{"mechanism": "M4 issuance", "transactions": b, "statements": st})
	// Revocation fan-out: replay logout-all of a user with 3 unrevoked sessions (1 expired) in a rolled-back runtime
	// transaction and count the capture function and attribution calls of that one statement (track_functions = 'all'
	// on the disposable database, seen by new runtime connections, as in TestRLSFootprint).
	db := pgx.Identifier{e.cfg.Database}.Sanitize()
	if _, err := e.setup.Exec(ctx, `ALTER DATABASE `+db+` SET track_functions = 'all'`); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = e.setup.Exec(ctx, `ALTER DATABASE `+db+` RESET track_functions`) }()
	fp := e.runtimePool(t, nil)
	x := e.newUser(t)
	e.relogin(t, x)
	c3 := e.relogin(t, x)
	if _, err := e.setup.Exec(ctx, `UPDATE app.sessions SET idle_expires_at = now() - interval '1 minute' WHERE id = $1::text::uuid`, e.sessionOf(t, c3).SessionID); err != nil {
		t.Fatal(err)
	}
	tx, err := fp.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true)`, x.id.String()); err != nil {
		t.Fatal(err)
	}
	calls := func() (int64, int64) {
		var a, r int64
		if err := tx.QueryRow(ctx, `SELECT coalesce(pg_stat_get_xact_function_calls('audit.events_attribute()'::regprocedure), 0),
  coalesce(pg_stat_get_xact_function_calls('audit.capture_session_revocations()'::regprocedure), 0)`).Scan(&a, &r); err != nil {
			t.Fatal(err)
		}
		return a, r
	}
	a0, r0 := calls()
	tag, err := tx.Exec(ctx, `UPDATE app.sessions SET revoked_at = now(), revoke_reason = 'logout_all' WHERE user_id = $1 AND revoked_at IS NULL`, x.id.String())
	if err != nil {
		t.Fatal(err)
	}
	var gap string
	if err := tx.QueryRow(ctx, `SELECT coalesce(current_setting('audit.revocation_gap', true), '')`).Scan(&gap); err != nil {
		t.Fatal(err)
	}
	a1, r1 := calls()
	line(map[string]any{"mechanism": "M5 revocation fan-out (logout_all)", "revoked_sessions": tag.RowsAffected(),
		"capture_statement_calls": r1 - r0, "event_rows": a1 - a0, "gap": gap, "extra_statements": 1})
	if tag.RowsAffected() != 3 || r1-r0 != 1 || a1-a0 != 3 || gap != "" {
		eAuditTrigger(map[string]any{"mechanism": "M5", "reason": "revocation fan-out is not one statement-level insert of N rows",
			"revoked": tag.RowsAffected(), "capture_calls": r1 - r0, "rows": a1 - a0})
		t.Errorf("fan-out: %d revoked, %d capture calls, %d rows", tag.RowsAffected(), r1-r0, a1-a0)
	}
	// A renewal-shaped update runs the statement trigger and no audit statement.
	if _, err := tx.Exec(ctx, `UPDATE app.sessions SET renewed_at = now() WHERE user_id = $1`, x.id.String()); err != nil {
		t.Fatal(err)
	}
	a2, r2 := calls()
	if a2 != a1 || r2-r1 != 1 {
		eAuditTrigger(map[string]any{"mechanism": "renewal", "reason": "a renewal executed an audit statement", "rows": a2 - a1})
		t.Errorf("renewal: %d audit rows, %d capture calls", a2-a1, r2-r1)
	}
	line(map[string]any{"mechanism": "sessions UPDATE without revocation (renewal)", "capture_statement_calls": r2 - r1, "event_rows": a2 - a1})
	// Reader calls of a direct append with a Group (bound ≤ 3).
	tx2, err := fp.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx2.Rollback(ctx) }()
	if _, err := tx2.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`, w.gadmin.id.String(), w.orgA.String()); err != nil {
		t.Fatal(err)
	}
	readers := func() int64 {
		var n int64
		if err := tx2.QueryRow(ctx, `SELECT coalesce(pg_stat_get_xact_function_calls('app.rls_org_role()'::regprocedure), 0) +
  coalesce(pg_stat_get_xact_function_calls('app.rls_group_role(uuid)'::regprocedure), 0)`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	before := readers()
	if _, err := tx2.Exec(ctx, `INSERT INTO audit.events (action, outcome, reason, operation, organization_id, group_id, resource_type, resource_id)
VALUES ('access.denied', 'denied', 'role_not_permitted', 'group.rename', $1::text::uuid, $2::text::uuid, 'group', $2::text::uuid)`, w.orgA.String(), w.g1.ID.String()); err != nil {
		t.Fatal(err)
	}
	n := readers() - before
	line(map[string]any{"mechanism": "direct append with a Group (reader calls)", "reader_calls": n})
	if n > 3 {
		eAuditTrigger(map[string]any{"mechanism": "direct append", "reason": "more than 3 reader calls", "reader_calls": n})
	}
}

// audExplain prints E-AUDIT-2: EXPLAIN (ANALYZE, BUFFERS) of the tenant read as the Organization owner.
func audExplain(t *testing.T, e *objEnv, w *audWorld) {
	t.Helper()
	ctx := context.Background()
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true), set_config('app.organization_id', $2, true)`, w.owner.id.String(), w.orgA.String()); err != nil {
		t.Fatal(err)
	}
	var raw []byte
	if err := tx.QueryRow(ctx, `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) SELECT id, occurred_at, action FROM audit.events
WHERE organization_id = $1 AND occurred_at >= '-infinity'::timestamptz AND (occurred_at, id) < ('infinity'::timestamptz, '00000000-0000-0000-0000-000000000000'::uuid)
ORDER BY occurred_at DESC, id DESC LIMIT 101`, w.orgA.String()).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var doc []map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil || len(doc) != 1 {
		t.Fatalf("explain: %v", err)
	}
	var version string
	_ = e.setup.QueryRow(ctx, `SHOW server_version`).Scan(&version)
	var rows int
	_ = e.setup.QueryRow(ctx, `SELECT count(*)::int FROM audit.events`).Scan(&rows)
	plan, _ := json.Marshal(doc[0]["Plan"])
	b, _ := json.Marshal(map[string]any{"evidence": "E-AUDIT-2", "plan": summarize(doc[0]["Plan"].(map[string]any)),
		"execution_ms": doc[0]["Execution Time"], "planning_ms": doc[0]["Planning Time"],
		"environment": map[string]any{"server_version": version, "audit_rows": rows, "tenants": 2}})
	fmt.Println("E-AUDIT-2 " + string(b))
	if regexp.MustCompile(`"Node Type":"Seq Scan","Parallel Aware":false,"Async Capable":false,"Relation Name":"events"`).Match(plan) {
		eAuditTrigger(map[string]any{"mechanism": "tenant read", "reason": "sequential scan of audit.events", "rows": rows})
	}
}

func pgErr(err error) *pgconn.PgError {
	var p *pgconn.PgError
	if errors.As(err, &p) {
		return p
	}
	return nil
}

// audIdentity is an identity service over the environment's pool, fake WorkOS and clock whose records (the G011 audit
// diagnostics included) go to buf.
func (e *itEnv) audIdentity(t *testing.T, buf *itSyncBuf) *identity.Service {
	t.Helper()
	a, err := workos.New(workos.Config{APIKey: e.fake.APIKey, ClientID: itClient, Issuer: e.fake.Issuer, BaseURL: e.fake.URL(),
		Algorithms: []string{"RS256"}, Now: e.clk.Now})
	if err != nil {
		t.Fatal(err)
	}
	return identity.NewService(identity.NewStore(e.pool), identity.Config{Provider: a, Issuer: e.fake.Issuer, ClientID: itClient,
		Hosts: itHost{org: e.loginOrg.String()}, Idle: itIdle, Max: itMax, Now: e.clk.Now,
		Logger: observability.NewLogger(buf, slog.LevelInfo, nil)})
}

// gapsSince returns the audit.gap records (operation/independent) of buf after byte offset n.
func gapsSince(t *testing.T, buf *itSyncBuf, n int) []string {
	t.Helper()
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(buf.String()[n:]), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("identity record: %v", err)
		}
		if m["event"] == "audit.gap" {
			out = append(out, fmt.Sprint(m["operation"], "/", m["independent"]))
		}
	}
	return out
}
