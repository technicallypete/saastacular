//go:build integration

// G058 B-ENROLLMENT-CATALOG, slice 1a subset (Plan r3 §4, §6.2 CT-ENR-1 / ENR-21 / ENR-22 negatives; selected-design
// addendum r0 §A.1/§A.3 and correction 01; Codex selected-design review F1/F4/F5): TestEnrollmentCatalog.
//   - CT-ENR-1_legacy_postcondition an Organization present at version 15 is invite_only with NULL readiness after the up
//     (nothing fabricated), cannot be opened (organizations_open_requires_resolved_history), and an Organization
//     inserted afterwards is ready by construction (created_after_ledger).
//   - CT-ENR-1_tables the three tables (columns, constraints, foreign keys, indexes, RLS/FORCE), the organizations and
//     login_transactions deltas.
//   - CT-ENR-1_policies the exact policy inventory on the three tables, the four new organizations policies and the
//     claim writer's audit policy (name, command, roles, load-bearing predicates); the organizations policy count grows
//     by exactly four.
//   - CT-ENR-1_privileges the ledger has no INSERT for rt/cw/pr/ar, no DELETE for any of them, no DELETE or UPDATE policy
//     and exactly one INSERT policy (the owner's, trigger-bound); its identity sequence is owner-only; the claim writer's
//     organizations columns are SELECT (id, enrollment_mode) and UPDATE (enrollment_mode) only.
//   - CT-ENR-1_definers the eight SECURITY DEFINER functions of 00016 (owner, prosecdef, volatility, search_path, EXECUTE
//     holders, reviewed bodies), the D-15(i) runtime EXECUTE, the owned-function inventories (claim writer 7) and the
//     departure and decision triggers.
//   - CT-ENR-1_audit the reason (68), operation (64), action (45) and resource-type (11) constraints and
//     events_attribute = the 00015 body + the reviewed enrollment delta (hash-diff), with positive/negative shape
//     controls on the runtime append path.
//   - CT-ENR-1_negatives (rolled back): owner ledger INSERT at depth 0, runtime ledger INSERT, a direct call of the
//     departure body, a runtime call of the lock helper, a claim-writer mode UPDATE and name read; positive controls:
//     the claim writer's FOR SHARE and a superuser binding delete recording exactly one departed row; atomicity: without
//     the owner policy the delete itself fails.
//   - CT-ENR-1_requests (request slice, every case rolled back): the request/decision/revocation/consumption state
//     machine through the real runtime and claim-writer paths (enrollment_request idempotent while pending, an expired
//     pending request kept and never consumable, approve/deny/revoke with their ledger rows and exactly one audited row
//     each, authority_seq derived and a newer restriction refusing with zero writes, a departed decider's approval
//     refused, the covered restriction lifted on admission), the privilege and transition negatives, and the mandatory
//     atomic append (a refused audit row rolls the request or decision back).
//   - CT-ENR-1_replay_pre16 / CT-ENR-1_replay (ENR-16 (a)-(h); Plan r3 §1.4, addendum §A.6, correction 01 §C.3/§C.6): the
//     bound infra/recovery/sql/replay-restrict.sql (ENROLLMENT_REPLAY_SQL_B64, SHA-256 pinned) as the superuser on
//     emulated restored states: ledger absent at version 15; trigger row present; restored row present; materialized
//     newer than a restored approval (admission refused, zero writes); backfill restored present; restore before the
//     backfill materialized and still closed until the explicit operator_backfilled re-attestation on the replay-only
//     history (restriction kept, X still refused after an authorized open); the ready-checkpoint/later-backfill gap materialized (open admission
//     refused, control admitted); an unmaterializable entry (absent subject, role mismatch) closes readiness, idempotently.
//   - CT-ENR-1_updown a clean down to the exact version-15 catalog and up again; the down refused, without losing
//     history, while a ledger row exists.
//
// The request slice adds the third table (13 columns as Plan r3 §4.2 enumerates them), its five policies and the claim
// writer's audit policy, the 7th claim-writer body and the owner-owned decision body (8 DEFINER), the state machine and
// the request capture, actions 42 -> 45 and resource types 10 -> 11.
package migrate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	enrOrgL    = "0f58e000-0000-4000-8000-000000000001" // present at version 15 (legacy)
	enrOrgN    = "0f58e000-0000-4000-8000-000000000002" // inserted after 00016
	enrOwnerL  = "0f58e000-0000-4000-8000-000000000011"
	enrMemberL = "0f58e000-0000-4000-8000-000000000012"
)

// enrShapeNarrowing is the reviewed events_attribute clause pair narrowing the enrollment refusal reasons to enrollment
// operations without an Organization, Group or resource (Codex round-2 audit decision), and enrollment.start and
// enrollment.admit to no Organization, Group or resource for every reason, start to its two reasons and admit to its
// reasons with an established actor (round-3 decisions 1 and the operation-level scope rule), exactly as 00016 writes it.
const enrShapeNarrowing = `
      AND (NEW.reason NOT IN ('enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch',
                              'recipient_ineligible', 'session_invalid', 'host_changed')
           OR (NEW.operation IN ('enrollment.start', 'enrollment.public', 'enrollment.status', 'enrollment.request',
                                 'enrollment.admit', 'enrollment.abandon', 'enrollment.policy.get', 'enrollment.policy.update',
                                 'enrollment.requests.list', 'enrollment.requests.decide', 'enrollment.requests.revoke')
               AND NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL))
      AND (NEW.operation NOT IN ('enrollment.start', 'enrollment.admit')
           OR (NEW.organization_id IS NULL AND NEW.group_id IS NULL AND NEW.resource_type IS NULL
               AND CASE WHEN NEW.operation = 'enrollment.start'
                        THEN NEW.reason IN ('enrollment_closed', 'invitation_mismatch')
                        ELSE NEW.actor_user_id IS NOT NULL AND NEW.reason IN ('target_not_found', 'enrollment_closed',
                             'enrollment_restricted', 'approval_revoked', 'recipient_ineligible', 'session_invalid',
                             'host_changed') END))`

// enrReadinessShape is the reviewed events_attribute branch of the three slice-1b captures (addendum §A.2): each names
// its Organization as the resource; readiness and backfill rows are the operator channel's (actor unknown).
const enrReadinessShape = `    WHEN NEW.action IN ('enrollment.policy_changed', 'enrollment.readiness_changed', 'enrollment.restriction_backfilled') THEN
      NEW.outcome = 'success' AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL AND NEW.group_id IS NULL
      AND NEW.organization_id IS NOT NULL AND NEW.resource_type = 'organization' AND NEW.resource_id = NEW.organization_id
      AND CASE NEW.action
            WHEN 'enrollment.policy_changed' THEN NEW.reason IS NULL AND NEW.subject_user_id IS NULL
            WHEN 'enrollment.readiness_changed' THEN NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NULL
              AND NEW.reason IN ('operator_verified_absent', 'operator_backfilled', 'history_unresolved')
            ELSE NEW.actor_kind = 'unknown' AND NEW.subject_user_id IS NOT NULL AND NEW.reason = 'operator_backfilled' END
`

// enrRequestShape is the reviewed events_attribute branch of the three request captures (Plan r3 §4.5; Codex round-8
// disposition 2: enrollment.decided carries no reason): the request is the resource, the applicant the subject, a user
// the actor, and enrollment.requested is attributed to the applicant itself.
const enrRequestShape = `    WHEN NEW.action IN ('enrollment.requested', 'enrollment.decided', 'enrollment.revoked') THEN
      NEW.outcome = 'success' AND NEW.reason IS NULL AND plain AND NEW.role IS NULL AND NEW.previous_role IS NULL
      AND NEW.group_id IS NULL AND NEW.organization_id IS NOT NULL AND NEW.subject_user_id IS NOT NULL
      AND NEW.resource_type = 'enrollment_request' AND NEW.resource_id IS NOT NULL AND NEW.actor_kind = 'user'
      AND (NEW.action <> 'enrollment.requested' OR NEW.actor_user_id = NEW.subject_user_id)
`

// enrReasons and enrOps are the seven reasons and eleven operations 00016 adds (Plan r3 §4.5, §11.2; Codex round-2
// audit decision +3 reasons; round-3 decision 1 enrollment.start).
var (
	enrReasons = []string{"enrollment_closed", "enrollment_restricted", "approval_revoked", "invitation_mismatch",
		"recipient_proof_unavailable", "recipient_ineligible", "session_invalid",
		"operator_verified_absent", "operator_backfilled", "history_unresolved"} // slice 1b: the readiness reasons
	// enrActions are the six trigger-owned actions 00016 adds: slice 1b's policy/readiness/backfill captures and the
	// request slice's three request captures.
	enrActions = []string{"enrollment.policy_changed", "enrollment.readiness_changed", "enrollment.restriction_backfilled",
		"enrollment.requested", "enrollment.decided", "enrollment.revoked"}
	enrOps = []string{"enrollment.start", "enrollment.public", "enrollment.status", "enrollment.request", "enrollment.admit", "enrollment.abandon",
		"enrollment.policy.get", "enrollment.policy.update", "enrollment.requests.list", "enrollment.requests.decide",
		"enrollment.requests.revoke"}
)

func TestEnrollmentCatalog(t *testing.T) {
	cfg := itConfig(t)
	ctx := context.Background()
	reset(t, cfg)
	r15 := runnerFor(t, cfg, subsetFS(t, 15))
	objRunUp(t, r15, "current_version\t15\n")
	setup := rlsSetupConn(t, cfg)
	cat := catalogConn(t, cfg)
	rt := rlsRuntimeConn(t, cfg)
	own := rlsOwnerConn(t, cfg)
	o, rtr, pr, ar, cw := cfg.OwnerRole(), cfg.RuntimeRole(), cfg.Database+"_policy_reader", cfg.Database+"_auth_reader",
		cfg.Database+"_claim_writer"
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := setup.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("fixture: %v", err)
		}
	}
	one := func(sql string, args ...any) string {
		t.Helper()
		var s *string
		if err := cat.QueryRow(ctx, sql, args...).Scan(&s); err != nil {
			t.Fatalf("catalog: %v", err)
		}
		if s == nil {
			return ""
		}
		return *s
	}
	// state is the catalog an 00016 up or down changes (the TestMembershipCatalog state, restated for this file):
	// relations/ACLs/history, every app and audit function, policy, column ACL, RLS flag, schema ACL, constraint and
	// non-internal trigger.
	state := func(r *Runner) string {
		t.Helper()
		return snapshot(t, r) + "\n" + one(`SELECT concat_ws(E'\n--\n',
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s|%s|%s|%s', p.oid::regprocedure, p.proowner::regrole,
     p.prosecdef, coalesce(p.proacl::text, ''), md5(p.prosrc)) AS x FROM pg_proc p
     WHERE p.pronamespace IN ('app'::regnamespace, 'audit'::regnamespace)) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s|%s|%s|%s|%s', p.polrelid::regclass, p.polname,
     p.polcmd, p.polroles::regrole[]::text, coalesce(pg_get_expr(p.polqual, p.polrelid), ''),
     coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '')) AS x FROM pg_policy p) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s.%s|%s', c.oid::regclass, a.attname, a.attacl::text) AS x
     FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
     WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND cardinality(a.attacl) > 0) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s|%s', c.oid::regclass, c.relrowsecurity,
     c.relforcerowsecurity) AS x FROM pg_class c
     WHERE c.relnamespace IN ('app'::regnamespace, 'audit'::regnamespace) AND c.relkind = 'r') s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s', n.nspname, coalesce(n.nspacl::text, '')) AS x
     FROM pg_namespace n WHERE n.nspname IN ('app', 'audit')) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s|%s', c.conrelid::regclass, c.conname,
     md5(pg_get_constraintdef(c.oid))) AS x FROM pg_constraint c
     WHERE c.connamespace IN ('app'::regnamespace, 'audit'::regnamespace)) s),
  (SELECT string_agg(x, E'\n' ORDER BY x COLLATE "C") FROM (SELECT format('%s|%s', t.tgrelid::regclass, t.tgname) AS x
     FROM pg_trigger t WHERE NOT t.tgisinternal) s))`)
	}
	labels := func(col string) string {
		return `CASE ` + col + ` WHEN 0 THEN 'public' WHEN $1::regrole THEN 'rt' WHEN $2::regrole THEN 'ow' WHEN $3::regrole THEN 'pr'
  WHEN $4::regrole THEN 'cw' WHEN $5::regrole THEN 'ar' ELSE 'other' END`
	}
	roles := []any{rtr, o, pr, cw, ar}

	// The version-15 fixture: one owned Organization with a member (the legacy install 00016 must close, not resolve).
	exec(`INSERT INTO app.users (id, display_name) VALUES ('` + enrOwnerL + `', 'Enr Owner L'), ('` + enrMemberL + `', 'Enr Member L')`)
	exec(`INSERT INTO app.organizations (id, name) VALUES ('` + enrOrgL + `', 'Enrollment Legacy')`)
	exec(`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + enrOrgL + `', '` + enrOwnerL + `', 'organization', '` + enrOrgL + `', NULL, 'org_owner', 'membership'),
  ('` + enrOrgL + `', '` + enrMemberL + `', 'organization', '` + enrOrgL + `', NULL, 'org_member', 'membership')`)
	orgPolicies15 := one(`SELECT count(*)::text FROM pg_policy WHERE polrelid = 'app.organizations'::regclass`)
	s15 := state(r15)
	// ENR-16 (d) (addendum §A.6; correction 01 §C.3/§C.4): a pre-00016 checkpoint has no ledger. The replay of an
	// Organization-scope membership.remove still applies the G013 delete and reports ledger absent: the guarded block writes
	// nothing and does not fail on the missing table and columns. The mandatory migrate then closes every Organization
	// (CT-ENR-1_legacy_postcondition below). Rolled back: the version-15 state stays the fixture's.
	t.Run("CT-ENR-1_replay_pre16", func(t *testing.T) {
		stmts := replayRestrictScript(t, replayVars("membership.remove", enrOrgL, enrMemberL, "organization", enrOrgL, "org_member"))
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		lines, err := psqlRun(ctx, tx, stmts)
		if got := strings.Join(lines, ";"); err != nil || got != "updated\t1;state\tnone;ledger\tabsent" {
			t.Errorf("replay on a version-15 checkpoint: %q %v, want updated 1, state none, ledger absent", got, err)
		}
		var reg *string
		if err := tx.QueryRow(ctx, `SELECT to_regclass('app.enrollment_restrictions')::text`).Scan(&reg); err != nil || reg != nil {
			t.Errorf("the ledger exists at version 15: %v %v", reg, err)
		}
	})
	r := runnerFor(t, cfg, subsetFS(t, 16))
	objRunUp(t, r, "applied\t00016_organization_enrollment.sql\ncurrent_version\t16\n")
	s16 := state(r)

	t.Run("CT-ENR-1_legacy_postcondition", func(t *testing.T) {
		var got string
		if err := setup.QueryRow(ctx, `SELECT format('%s|%s|%s', enrollment_mode, enrollment_history_resolved_at IS NULL,
  coalesce(enrollment_history_basis, '-')) FROM app.organizations WHERE id = $1`, enrOrgL).Scan(&got); err != nil || got != "invite_only|t|-" {
			t.Errorf("legacy Organization after 00016: %q %v, want invite_only with NULL readiness", got, err)
		}
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(ctx, `UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = $1`, enrOrgL)
		_ = tx.Rollback(ctx)
		if sqlstateOf(err) != "23514" || !strings.Contains(err.Error(), "organizations_open_requires_resolved_history") {
			t.Errorf("opening an unresolved Organization: %v, want 23514 organizations_open_requires_resolved_history", err)
		}
		exec(`INSERT INTO app.organizations (id, name) VALUES ('` + enrOrgN + `', 'Enrollment New')`)
		if err := setup.QueryRow(ctx, `SELECT format('%s|%s|%s', enrollment_mode, enrollment_history_resolved_at IS NOT NULL,
  enrollment_history_basis) FROM app.organizations WHERE id = $1`, enrOrgN).Scan(&got); err != nil || got != "invite_only|t|created_after_ledger" {
			t.Errorf("Organization inserted after 00016: %q %v, want ready by construction", got, err)
		}
	})

	t.Run("CT-ENR-1_tables", func(t *testing.T) {
		metadataThroughVersion16Only(t, cfg, r)
		cols := func(rel string) string {
			return one(`SELECT string_agg(a.attname || ':' || format_type(a.atttypid, a.atttypmod) || ':' || a.attnotnull::text, ',' ORDER BY a.attnum)
FROM pg_attribute a WHERE a.attrelid = $1::regclass AND a.attnum > 0 AND NOT a.attisdropped`, rel)
		}
		if got := cols("app.enrollment_sessions"); got != "id:uuid:true,user_id:uuid:true,organization_id:uuid:true,origin:text:true,"+
			"credential_hash:bytea:true,csrf_token:bytea:true,provider_session_id:text:true,created_at:timestamp with time zone:true,"+
			"expires_at:timestamp with time zone:true,admitted_at:timestamp with time zone:false,admitted_binding_id:uuid:false,"+
			"revoked_at:timestamp with time zone:false,revoke_reason:text:false" {
			t.Errorf("enrollment_sessions columns (13): %s", got)
		}
		// Plan r3 §4.2 enumerates twelve ledger columns (its "13" count does not match its own list; the list is bound).
		if got := cols("app.enrollment_restrictions"); got != "id:uuid:true,seq:bigint:true,organization_id:uuid:true,user_id:uuid:true,"+
			"kind:text:true,recorded_via:text:true,created_at:timestamp with time zone:true,created_by_user_id:uuid:false,"+
			"source_request_id:uuid:false,lifted_at:timestamp with time zone:false,lifted_by_user_id:uuid:false,lifted_via:text:false" {
			t.Errorf("enrollment_restrictions columns (12): %s", got)
		}
		// Plan r3 §4.2 enumerates thirteen request columns incl. authority_seq (Codex round-8 disposition 1: the list is
		// bound; its "15" prose is wrong).
		if got := cols("app.enrollment_requests"); got != "id:uuid:true,organization_id:uuid:true,user_id:uuid:true,status:text:true,"+
			"created_at:timestamp with time zone:true,expires_at:timestamp with time zone:true,decided_at:timestamp with time zone:false,"+
			"decided_by_user_id:uuid:false,revoked_at:timestamp with time zone:false,revoked_by_user_id:uuid:false,"+
			"consumed_at:timestamp with time zone:false,consumed_binding_id:uuid:false,authority_seq:bigint:true" {
			t.Errorf("enrollment_requests columns (13): %s", got)
		}
		if got := one(`SELECT pg_get_expr(d.adbin, d.adrelid) FROM pg_attrdef d
WHERE d.adrelid = 'app.enrollment_restrictions'::regclass
  AND d.adnum = (SELECT attnum FROM pg_attribute WHERE attrelid = 'app.enrollment_restrictions'::regclass AND attname = 'created_at')`); got != "clock_timestamp()" {
			t.Errorf("ledger created_at default: %s, want clock_timestamp()", got)
		}
		if got := one(`SELECT attidentity::text FROM pg_attribute WHERE attrelid = 'app.enrollment_restrictions'::regclass AND attname = 'seq'`); got != "a" {
			t.Errorf("ledger seq identity: %q, want GENERATED ALWAYS (a)", got)
		}
		if got := one(`SELECT string_agg(relname || ':' || relrowsecurity::text || ':' || relforcerowsecurity::text, ',' ORDER BY relname)
FROM pg_class WHERE oid IN ('app.enrollment_sessions'::regclass, 'app.enrollment_restrictions'::regclass, 'app.organizations'::regclass,
  'app.login_transactions'::regclass, 'app.enrollment_requests'::regclass)`); got != "enrollment_requests:true:true,"+
			"enrollment_restrictions:true:true,enrollment_sessions:true:true,login_transactions:true:true,organizations:true:true" {
			t.Errorf("RLS/FORCE at version 16: %s", got)
		}
		cons := func(rel string) string {
			return one(`SELECT string_agg(conname || ':' || contype::text, ',' ORDER BY conname) FROM pg_constraint WHERE conrelid = $1::regclass`, rel)
		}
		if got := cons("app.enrollment_sessions"); got != "enrollment_sessions_admission_shape:c,enrollment_sessions_credential_hash_key:u,"+
			"enrollment_sessions_credential_hash_length:c,enrollment_sessions_csrf_token_length:c,enrollment_sessions_lifetime:c,"+
			"enrollment_sessions_organization_fkey:f,enrollment_sessions_origin_length:c,enrollment_sessions_pkey:p,"+
			"enrollment_sessions_provider_session_id_length:c,enrollment_sessions_revocation_shape:c,enrollment_sessions_user_fkey:f" {
			t.Errorf("enrollment_sessions constraints: %s", got)
		}
		if got := cons("app.enrollment_restrictions"); got != "enrollment_restrictions_kind_valid:c,enrollment_restrictions_lift_shape:c,"+
			"enrollment_restrictions_lifted_via_valid:c,enrollment_restrictions_organization_fkey:f,enrollment_restrictions_pkey:p,"+
			"enrollment_restrictions_recorded_via_valid:c,enrollment_restrictions_replay_shape:c,enrollment_restrictions_seq_key:u,"+
			"enrollment_restrictions_source_request_fkey:f,enrollment_restrictions_user_fkey:f" {
			t.Errorf("enrollment_restrictions constraints: %s", got)
		}
		if got := cons("app.enrollment_requests"); got != "enrollment_requests_authority_shape:c,enrollment_requests_consumption_shape:c,"+
			"enrollment_requests_decided_by_fkey:f,enrollment_requests_decision_shape:c,enrollment_requests_lifetime:c,"+
			"enrollment_requests_organization_fkey:f,enrollment_requests_pkey:p,enrollment_requests_revocation_shape:c,"+
			"enrollment_requests_revoked_by_fkey:f,enrollment_requests_status_valid:c,enrollment_requests_user_fkey:f" {
			t.Errorf("enrollment_requests constraints: %s", got)
		}
		if got := one(`SELECT string_agg(conname || ':' || confdeltype::text, ',' ORDER BY conname) FROM pg_constraint
WHERE conrelid IN ('app.enrollment_sessions'::regclass, 'app.enrollment_restrictions'::regclass, 'app.enrollment_requests'::regclass)
  AND contype = 'f'`); got !=
			"enrollment_requests_decided_by_fkey:r,enrollment_requests_organization_fkey:r,enrollment_requests_revoked_by_fkey:r,"+
				"enrollment_requests_user_fkey:r,enrollment_restrictions_organization_fkey:r,enrollment_restrictions_source_request_fkey:r,"+
				"enrollment_restrictions_user_fkey:r,enrollment_sessions_organization_fkey:r,enrollment_sessions_user_fkey:r" {
			t.Errorf("enrollment foreign keys: %s", got)
		}
		def := func(name string) string {
			return one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conname = $1`, name)
		}
		for name, frags := range map[string][]string{
			"enrollment_sessions_lifetime":                 {"00:30:00"},
			"enrollment_restrictions_recorded_via_valid":   {"'trigger'", "'replay'", "'backfill'"},
			"enrollment_restrictions_kind_valid":           {"'departed'", "'denied'", "'revoked'"},
			"enrollment_restrictions_replay_shape":         {"'trigger'", "'departed'", "source_request_id IS NULL", "created_by_user_id IS NULL"},
			"organizations_enrollment_mode_valid":          {"'invite_only'", "'request_access'", "'open'"},
			"organizations_enrollment_history_basis_valid": {"'created_after_ledger'", "'operator_verified_absent'", "'operator_backfilled'"},
			"organizations_open_requires_resolved_history": {"'open'", "enrollment_history_resolved_at IS NOT NULL"},
			"login_transactions_return_path_check":         {"'/'", "'/account'", "'/join/status'"},
			"enrollment_requests_status_valid":             {"'pending'", "'approved'", "'denied'", "'revoked'"},
			"enrollment_requests_lifetime":                 {"expires_at > created_at", "30 days"},
			"enrollment_requests_decision_shape":           {"status = 'pending'", "decided_at IS NULL", "decided_by_user_id IS NULL"},
			"enrollment_requests_revocation_shape":         {"status = 'revoked'", "revoked_at IS NOT NULL", "revoked_by_user_id IS NULL"},
			"enrollment_requests_consumption_shape":        {"consumed_at IS NULL", "consumed_binding_id IS NULL", "status = 'approved'"},
			"enrollment_requests_authority_shape":          {"authority_seq >= 0", "authority_seq = 0", "'approved'", "'revoked'"},
		} {
			got := def(name)
			for _, f := range frags {
				if !strings.Contains(got, f) {
					t.Errorf("constraint %s lacks %q: %s", name, f, got)
				}
			}
		}
		if got := one(`SELECT string_agg(c.relname, ',' ORDER BY c.relname) FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
WHERE i.indrelid IN ('app.enrollment_sessions'::regclass, 'app.enrollment_restrictions'::regclass, 'app.enrollment_requests'::regclass)`); got !=
			"enrollment_requests_one_pending,enrollment_requests_pkey,enrollment_requests_scope_idx,enrollment_requests_user_idx,"+
				"enrollment_restrictions_open_idx,enrollment_restrictions_pkey,enrollment_restrictions_seq_key,"+
				"enrollment_sessions_credential_hash_key,enrollment_sessions_pkey,enrollment_sessions_provider_idx,enrollment_sessions_user_idx" {
			t.Errorf("enrollment indexes: %s", got)
		}
		// one_pending: unique, partial on the pending status, over (organization_id, user_id).
		if got := one(`SELECT pg_get_indexdef('app.enrollment_requests_one_pending'::regclass)`); !strings.Contains(got, "CREATE UNIQUE INDEX") ||
			!strings.Contains(got, "(organization_id, user_id) WHERE (status = 'pending'::text)") {
			t.Errorf("one_pending index: %s", got)
		}
		if got := one(`SELECT string_agg(a.attname || ':' || format_type(a.atttypid, a.atttypmod) || ':' || a.attnotnull::text || ':' ||
  coalesce(pg_get_expr(d.adbin, d.adrelid), '-'), ',' ORDER BY a.attname) FROM pg_attribute a
LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = 'app.organizations'::regclass AND a.attname LIKE 'enrollment\_%' AND NOT a.attisdropped`); got !=
			"enrollment_history_basis:text:false:'created_after_ledger'::text,enrollment_history_resolved_at:timestamp with time zone:false:now(),"+
				"enrollment_mode:text:true:'invite_only'::text" {
			t.Errorf("organizations enrollment columns: %s", got)
		}
		if got := one(`SELECT string_agg(a.attname || ':' || format_type(a.atttypid, a.atttypmod) || ':' || a.attnotnull::text || ':' ||
  coalesce(pg_get_expr(d.adbin, d.adrelid), '-'), ',' ORDER BY a.attname) FROM pg_attribute a
LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
WHERE a.attrelid = 'app.login_transactions'::regclass AND a.attname IN ('enrollment_intent', 'invitation_id', 'request_on_arrival')
  AND NOT a.attisdropped`); got != "enrollment_intent:boolean:true:false,invitation_id:uuid:false:-,request_on_arrival:boolean:true:false" {
			t.Errorf("login_transactions enrollment columns: %s", got)
		}
	})

	t.Run("CT-ENR-1_policies", func(t *testing.T) {
		if got := one(`SELECT string_agg(format('%s:%s:%s', p.polname, p.polcmd, (SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C")
  FROM (SELECT `+labels("x")+` AS v FROM unnest(p.polroles) x) l)), ';' ORDER BY p.polname)
FROM pg_policy p WHERE p.polrelid IN ('app.enrollment_sessions'::regclass, 'app.enrollment_restrictions'::regclass,
  'app.enrollment_requests'::regclass)
  OR p.polname IN ('organizations_policy_reader', 'organizations_claim_select', 'organizations_claim_lock',
                   'organizations_enrollment_update', 'events_insert_enrollment')`, roles...); got !=
			"enrollment_requests_claim_insert:a:cw;enrollment_requests_claim_select:r:cw;enrollment_requests_claim_update:w:cw;"+
				"enrollment_requests_decide:w:ow+rt;enrollment_requests_select:r:ow+rt;"+
				"enrollment_restrictions_claim_lift:w:cw;enrollment_restrictions_claim_select:r:cw;enrollment_restrictions_owner_insert:a:ow;"+
				"enrollment_restrictions_select:r:ow+rt;enrollment_sessions_auth_reader:r:ar;enrollment_sessions_claim_select:r:cw;"+
				"enrollment_sessions_claim_update:w:cw;enrollment_sessions_expire:d:ow+rt;enrollment_sessions_insert:a:ow+rt;"+
				"enrollment_sessions_select:r:ow+rt;enrollment_sessions_update:w:ow+rt;events_insert_enrollment:a:cw;organizations_claim_lock:w:cw;"+
				"organizations_claim_select:r:cw;organizations_enrollment_update:w:ow+rt;organizations_policy_reader:r:pr" {
			t.Errorf("enrollment policies:\n got %s", got)
		}
		// Slice 1b: +1 (organizations_enrollment_update, the runtime/owner mode write of an Organization manager).
		if got := one(`SELECT (count(*) - $1::int)::text FROM pg_policy WHERE polrelid = 'app.organizations'::regclass`, orgPolicies15); got != "4" {
			t.Errorf("organizations policies added by 00016: %s, want 4", got)
		}
		pol := func(name string) string {
			return one(`SELECT regexp_replace(coalesce(pg_get_expr(polqual, polrelid), ''), '\s+', ' ', 'g') || ' || ' ||
  regexp_replace(coalesce(pg_get_expr(polwithcheck, polrelid), ''), '\s+', ' ', 'g') FROM pg_policy WHERE polname = $1`, name)
		}
		for name, want := range map[string][]string{
			"organizations_policy_reader":          {"true || "},
			"organizations_claim_select":           {"true || "},
			"organizations_claim_lock":             {"true || false"},
			"organizations_enrollment_update":      {"id = ( SELECT app.ctx_org()", "rls_org_role()", "'org_owner'", "'org_admin'"},
			"enrollment_sessions_select":           {"user_id = ( SELECT app.ctx_user()"},
			"enrollment_sessions_insert":           {"user_id = ( SELECT app.ctx_user()", "ctx_org()", "revoked_at IS NULL", "admitted_at IS NULL"},
			"enrollment_sessions_update":           {"user_id = ( SELECT app.ctx_user()", "ctx_org()"},
			"enrollment_sessions_expire":           {"expires_at <= now()", "ctx_user()", "ctx_login_binding()"},
			"enrollment_sessions_auth_reader":      {"true || "},
			"enrollment_sessions_claim_select":     {"true || "},
			"enrollment_sessions_claim_update":     {"true || (user_id = app.ctx_user())"},
			"enrollment_restrictions_select":       {"user_id = ( SELECT app.ctx_user()", "organization_id = ( SELECT app.ctx_org()", "rls_org_role()", "'org_owner'", "'org_admin'"},
			"enrollment_restrictions_claim_select": {"true || "},
			"enrollment_restrictions_claim_lift":   {"true || ", "user_id = app.ctx_user()", "lifted_at IS NOT NULL"},
			"enrollment_restrictions_owner_insert": {"pg_trigger_depth() >= 1", "'departed'", "'denied'", "'revoked'", "recorded_via = 'trigger'",
				"lifted_at IS NULL", "lifted_by_user_id IS NULL", "lifted_via IS NULL"},
			"enrollment_requests_select": {"user_id = ( SELECT app.ctx_user()", "organization_id = ( SELECT app.ctx_org()", "rls_org_role()",
				"'org_owner'", "'org_admin'"},
			"enrollment_requests_decide": {"organization_id = ( SELECT app.ctx_org()", "rls_org_role()", "'pending'", "'approved'", " || ",
				"'denied'", "'revoked'"},
			"enrollment_requests_claim_insert": {" || ((user_id = app.ctx_user()) AND (status = 'pending'"},
			"enrollment_requests_claim_select": {"true || "},
			"enrollment_requests_claim_update": {"true || (user_id = app.ctx_user())"},
			"events_insert_enrollment": {"pg_trigger_depth() >= 1", "actor_kind = 'user'", "actor_user_id = app.ctx_user()",
				"organization_id IS NOT NULL", "action = 'enrollment.requested'"},
		} {
			got := pol(name)
			for _, f := range want {
				if !strings.Contains(got, f) {
					t.Errorf("policy %s lacks %q: %s", name, f, got)
				}
			}
		}
	})

	t.Run("CT-ENR-1_privileges", func(t *testing.T) {
		for _, role := range []string{rtr, cw, pr, ar} {
			if got := one(`SELECT format('%s:%s:%s', has_table_privilege($1, 'app.enrollment_restrictions', 'INSERT'),
  has_any_column_privilege($1, 'app.enrollment_restrictions', 'INSERT'), has_table_privilege($1, 'app.enrollment_restrictions', 'DELETE'))`,
				role); got != "f:f:f" {
				t.Errorf("%s INSERT/column INSERT/DELETE on the ledger: %s, want none", role, got)
			}
		}
		// The owner holds the ACL of its table, but FORCE RLS leaves it exactly one (trigger-bound) INSERT policy and no
		// UPDATE or DELETE policy; no policy of any role deletes a ledger row.
		if got := one(`SELECT string_agg(polname || ':' || polcmd::text, ',' ORDER BY polname) FROM pg_policy
WHERE polrelid = 'app.enrollment_restrictions'::regclass AND (polcmd IN ('a', 'd', '*') OR $1::regrole = ANY (polroles) AND polcmd = 'w')`, o); got !=
			"enrollment_restrictions_owner_insert:a" {
			t.Errorf("ledger write policies beyond the claim writer's lift: %s", got)
		}
		if got := one(`SELECT coalesce(string_agg(a.grantee::regrole::text || ':' || a.privilege_type, ','), '') FROM pg_class c
CROSS JOIN LATERAL aclexplode(c.relacl) a WHERE c.oid = pg_get_serial_sequence('app.enrollment_restrictions', 'seq')::regclass
  AND a.grantee <> c.relowner`); got != "" {
			t.Errorf("ledger identity sequence grantees: %s, want owner only", got)
		}
		colgrants := func(grantee, rel string) string {
			return one(`SELECT coalesce(string_agg(x, ',' ORDER BY x COLLATE "C"), '') FROM (SELECT at.attname || ':' || a.privilege_type AS x
FROM pg_attribute at CROSS JOIN LATERAL aclexplode(at.attacl) a
WHERE at.attrelid = $2::regclass AND at.attnum > 0 AND NOT at.attisdropped AND a.grantee = $1::regrole) s`, grantee, rel)
		}
		for _, c := range []struct{ role, rel, want string }{
			{cw, "app.organizations", "enrollment_mode:SELECT,enrollment_mode:UPDATE,id:SELECT"},
			{pr, "app.organizations", "enrollment_mode:SELECT,id:SELECT"},
			{rtr, "app.enrollment_sessions", "created_at:INSERT,credential_hash:INSERT,csrf_token:INSERT,expires_at:INSERT,id:INSERT," +
				"organization_id:INSERT,origin:INSERT,provider_session_id:INSERT,revoke_reason:UPDATE,revoked_at:UPDATE,user_id:INSERT"},
			{cw, "app.enrollment_sessions", "admitted_at:UPDATE,admitted_binding_id:UPDATE"},
			{cw, "app.enrollment_restrictions", "lifted_at:UPDATE,lifted_by_user_id:UPDATE,lifted_via:UPDATE"},
			// Request slice: the runtime decides/revokes through five columns (never authority_seq or the consumption pair);
			// the claim writer inserts five columns (never a decision or authority_seq) and consumes through two.
			{rtr, "app.enrollment_requests", "decided_at:UPDATE,decided_by_user_id:UPDATE,revoked_at:UPDATE,revoked_by_user_id:UPDATE,status:UPDATE"},
			{cw, "app.enrollment_requests", "consumed_at:UPDATE,consumed_binding_id:UPDATE,expires_at:INSERT,id:INSERT,organization_id:INSERT," +
				"status:INSERT,user_id:INSERT"},
			{pr, "app.enrollment_requests", ""}, {ar, "app.enrollment_requests", ""},
		} {
			if got := colgrants(c.role, c.rel); got != c.want {
				t.Errorf("%s column grants on %s:\n got %s\nwant %s", c.role, c.rel, got, c.want)
			}
		}
		if got := one(`SELECT coalesce(string_agg(c.oid::regclass::text || ':' || a.grantee::regrole::text || ':' || a.privilege_type, ','
  ORDER BY c.oid::regclass::text, a.grantee::regrole::text, a.privilege_type), '') FROM pg_class c CROSS JOIN LATERAL aclexplode(c.relacl) a
WHERE c.oid IN ('app.enrollment_sessions'::regclass, 'app.enrollment_restrictions'::regclass, 'app.enrollment_requests'::regclass)
  AND a.grantee <> c.relowner`); got !=
			"app.enrollment_requests:"+cw+":SELECT,app.enrollment_requests:"+rtr+":SELECT,"+
				"app.enrollment_restrictions:"+cw+":SELECT,app.enrollment_restrictions:"+rtr+":SELECT,"+
				"app.enrollment_sessions:"+ar+":SELECT,app.enrollment_sessions:"+cw+":SELECT,"+
				"app.enrollment_sessions:"+rtr+":DELETE,app.enrollment_sessions:"+rtr+":SELECT" {
			t.Errorf("enrollment table-level grants: %s", got)
		}
		for _, c := range []struct {
			role, col, priv string
			want            bool
		}{
			{cw, "name", "SELECT", false}, {cw, "name", "UPDATE", false}, {cw, "enrollment_history_resolved_at", "UPDATE", false},
			{pr, "name", "SELECT", false}, {rtr, "enrollment_history_resolved_at", "UPDATE", false},
			// Slice 1b: the runtime writes the mode column only (the readiness pair has no session-role writer).
			{rtr, "enrollment_mode", "UPDATE", true}, {rtr, "enrollment_history_basis", "UPDATE", false},
			{cw, "enrollment_history_basis", "UPDATE", false}, {pr, "enrollment_history_basis", "UPDATE", false},
		} {
			if got := one(`SELECT has_column_privilege($1, 'app.organizations', $2, $3)::text`, c.role, c.col, c.priv); got != map[bool]string{true: "true", false: "false"}[c.want] {
				t.Errorf("%s %s on app.organizations.%s: %s", c.role, c.priv, c.col, got)
			}
		}
	})

	t.Run("CT-ENR-1_definers", func(t *testing.T) {
		if got := one(`SELECT string_agg(p.proname || ':' || `+labels("p.proowner")+` || ':' || p.prosecdef::text || ':' || p.provolatile::text
  || ':' || p.proisstrict::text || ':' || coalesce(array_to_string(p.proconfig, '|'), '-') || ':' || coalesce((SELECT string_agg(l.v, '+' ORDER BY l.v COLLATE "C")
  FROM (SELECT `+labels("a.grantee")+` || '=' || a.privilege_type AS v FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner) l), '-'),
  ';' ORDER BY p.proname)
FROM pg_proc p WHERE p.oid IN ('app.rls_enrollment_mode(uuid)'::regprocedure, 'app.rls_invitation_pending(uuid, uuid)'::regprocedure,
  'app.auth_resolve_enrollment(bytea, timestamptz)'::regprocedure, 'app.enrollment_lock_organization(uuid)'::regprocedure,
  'app.enrollment_admit(bytea, uuid, text, uuid)'::regprocedure, 'app.enrollment_restrict_on_departure()'::regprocedure,
  'app.enrollment_request(bytea, uuid)'::regprocedure, 'app.enrollment_restrict_on_decision()'::regprocedure,
  'app.invitation_claim_organization(uuid, bytea, text, text, text)'::regprocedure)`, roles...); got !=
			"auth_resolve_enrollment:ar:true:s:false:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
				"enrollment_admit:cw:true:v:false:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
				"enrollment_lock_organization:cw:true:v:true:search_path=pg_catalog, pg_temp:ow=EXECUTE;"+
				"enrollment_request:cw:true:v:false:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
				"enrollment_restrict_on_decision:ow:true:v:false:search_path=pg_catalog, pg_temp:-;"+
				"enrollment_restrict_on_departure:ow:true:v:false:search_path=pg_catalog, pg_temp:-;"+
				"invitation_claim_organization:cw:true:v:false:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
				"rls_enrollment_mode:pr:true:s:false:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE;"+
				"rls_invitation_pending:cw:true:s:false:search_path=pg_catalog, pg_temp:ow=EXECUTE+rt=EXECUTE" {
			t.Errorf("enrollment functions:\n got %s", got)
		}
		// Owned-function inventories: claim writer 3 -> 7 (1a's three and the request slice's enrollment_request); policy
		// reader +1; auth reader +1.
		for _, c := range []struct{ role, want string }{
			{cw, "enrollment_admit,enrollment_lock_organization,enrollment_request,invitation_claim_group,invitation_claim_organization," +
				"rls_invitation_pending,rls_user_enabled_lock"},
			{pr, "rls_dependent_group_bindings,rls_enrollment_mode,rls_group_invitation_candidates,rls_group_role,rls_org_role," +
				"rls_own_organizations,rls_owner_count,rls_profile_eligible,rls_sole_owner_scopes"},
			{ar, "auth_logout_ref,auth_principal_subject,auth_resolve_enrollment,auth_resolve_session,hostname_state,is_organization_member," +
				"public_branding"},
		} {
			if got := one(`SELECT string_agg(proname, ',' ORDER BY proname) FROM pg_proc WHERE proowner = $1::regrole`, c.role); got != c.want {
				t.Errorf("functions owned by %s: %s, want %s", c.role, got, c.want)
			}
		}
		// The owner-owned SECURITY DEFINER bodies: the departure and decision triggers.
		if got := one(`SELECT string_agg(proname, ',' ORDER BY proname) FROM pg_proc
WHERE prosecdef AND proowner = $1::regrole AND pronamespace = 'app'::regnamespace`, o); got != "enrollment_restrict_on_decision,enrollment_restrict_on_departure" {
			t.Errorf("owner-owned DEFINER functions: %s", got)
		}
		for _, f := range []struct{ sig, marker string }{
			{"app.rls_enrollment_mode(uuid)", "CREATE FUNCTION app.rls_enrollment_mode("},
			{"app.rls_invitation_pending(uuid, uuid)", "CREATE FUNCTION app.rls_invitation_pending("},
			{"app.auth_resolve_enrollment(bytea, timestamptz)", "CREATE FUNCTION app.auth_resolve_enrollment("},
			{"app.enrollment_lock_organization(uuid)", "CREATE FUNCTION app.enrollment_lock_organization("},
			{"app.enrollment_admit(bytea, uuid, text, uuid)", "CREATE FUNCTION app.enrollment_admit("},
			{"app.enrollment_restrict_on_departure()", "CREATE FUNCTION app.enrollment_restrict_on_departure("},
			{"app.enrollment_request(bytea, uuid)", "CREATE FUNCTION app.enrollment_request("},
			{"app.enrollment_restrict_on_decision()", "CREATE FUNCTION app.enrollment_restrict_on_decision("},
		} {
			if got := one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = $1::regprocedure`, f.sig); got !=
				md5hex(memBody(t, "00016_organization_enrollment.sql", f.marker)) {
				t.Errorf("%s differs from the reviewed migration body", f.sig)
			}
		}
		// Slice 1b (addendum §A.2): the readiness setter and backfill, the guard and the two captures are owner-owned
		// SECURITY INVOKER bodies with no grantee at all (PUBLIC revoked; no session role executes them); bodies pinned.
		for _, f := range []struct{ sig, marker string }{
			{"app.enrollment_readiness_set(uuid, text, text)", "CREATE FUNCTION app.enrollment_readiness_set("},
			{"app.enrollment_history_backfill(uuid, uuid, text)", "CREATE FUNCTION app.enrollment_history_backfill("},
			{"app.organizations_enrollment_guard()", "CREATE FUNCTION app.organizations_enrollment_guard("},
			{"audit.capture_organization_enrollment()", "CREATE FUNCTION audit.capture_organization_enrollment("},
			{"audit.capture_enrollment_restrictions()", "CREATE FUNCTION audit.capture_enrollment_restrictions("},
			// Request slice: the state machine and the request capture.
			{"app.enrollment_requests_immutable()", "CREATE FUNCTION app.enrollment_requests_immutable("},
			{"audit.capture_enrollment_requests()", "CREATE FUNCTION audit.capture_enrollment_requests("},
		} {
			if got := one(`SELECT format('%s|%s|%s', p.proowner = $2::regrole, p.prosecdef, coalesce((SELECT string_agg(a.grantee::regrole::text, '+')
  FROM aclexplode(p.proacl) a WHERE a.grantee <> p.proowner), '-')) FROM pg_proc p WHERE p.oid = $1::regprocedure`, f.sig, o); got != "t|f|-" {
				t.Errorf("%s owner/definer/grantees: %s, want owner-owned invoker without grantees", f.sig, got)
			}
			if got := one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = $1::regprocedure`, f.sig); got !=
				md5hex(memBody(t, "00016_organization_enrollment.sql", f.marker)) {
				t.Errorf("%s differs from the reviewed migration body", f.sig)
			}
		}
		for rel, want := range map[string]string{
			"app.organizations": "BEFORE UPDATE OF enrollment_mode, enrollment_history_resolved_at, enrollment_history_basis ON app.organizations FOR EACH ROW EXECUTE FUNCTION app.organizations_enrollment_guard()|" +
				"AFTER UPDATE OF enrollment_mode, enrollment_history_resolved_at, enrollment_history_basis ON app.organizations FOR EACH ROW WHEN",
			"app.enrollment_restrictions": "AFTER INSERT ON app.enrollment_restrictions FOR EACH ROW WHEN ((new.recorded_via = 'backfill'::text)) EXECUTE FUNCTION audit.capture_enrollment_restrictions()",
		} {
			defs := one(`SELECT string_agg(pg_get_triggerdef(oid), '|' ORDER BY tgname) FROM pg_trigger
WHERE tgrelid = $1::regclass AND tgname IN ('organizations_enrollment_guard', 'audit_capture_enrollment', 'audit_capture_backfill')`, rel)
			for _, frag := range strings.Split(want, "|") {
				if !strings.Contains(defs, frag) {
					t.Errorf("%s enrollment triggers lack %q: %s", rel, frag, defs)
				}
			}
		}
		// Request slice: exactly four triggers on app.enrollment_requests (the capture of the insert, the capture of a status
		// change, the decision body on denied/revoked, the BEFORE state machine on every update).
		if got := one(`SELECT string_agg(pg_get_triggerdef(oid), '|' ORDER BY tgname) FROM pg_trigger
WHERE tgrelid = 'app.enrollment_requests'::regclass AND NOT tgisinternal`); got !=
			"CREATE TRIGGER audit_capture_request_insert AFTER INSERT ON app.enrollment_requests FOR EACH ROW EXECUTE FUNCTION audit.capture_enrollment_requests()|"+
				"CREATE TRIGGER audit_capture_request_status AFTER UPDATE OF status ON app.enrollment_requests FOR EACH ROW WHEN ((old.status IS DISTINCT FROM new.status)) EXECUTE FUNCTION audit.capture_enrollment_requests()|"+
				"CREATE TRIGGER enrollment_requests_immutable BEFORE UPDATE ON app.enrollment_requests FOR EACH ROW EXECUTE FUNCTION app.enrollment_requests_immutable()|"+
				"CREATE TRIGGER enrollment_requests_restrict AFTER UPDATE OF status ON app.enrollment_requests FOR EACH ROW WHEN (((new.status = ANY (ARRAY['denied'::text, 'revoked'::text])) AND (old.status IS DISTINCT FROM new.status))) EXECUTE FUNCTION app.enrollment_restrict_on_decision()" {
			t.Errorf("enrollment_requests triggers:\n got %s", got)
		}
		// Load-bearing body fragments (the hash pins the rest).
		body := func(sig string) string {
			return one(`SELECT prosrc FROM pg_proc WHERE oid = $1::regprocedure`, sig)
		}
		for sig, frags := range map[string][]string{
			"app.enrollment_restrict_on_departure()": {"pg_trigger_depth() < 1", "TG_WHEN IS DISTINCT FROM 'AFTER'", "TG_LEVEL IS DISTINCT FROM 'ROW'",
				"TG_TABLE_SCHEMA IS DISTINCT FROM 'app'", "TG_TABLE_NAME IS DISTINCT FROM 'role_bindings'", "TG_OP IS DISTINCT FROM 'DELETE'",
				"app.enrollment_lock_organization(OLD.organization_id)", "'departed', 'trigger', app.ctx_user()"},
			"app.enrollment_lock_organization(uuid)": {"FOR NO KEY UPDATE;\n  RETURN FOUND;"},
			"app.rls_invitation_pending(uuid, uuid)": {"i.id = p_id AND i.organization_id = p_org AND i.scope_type = 'organization'",
				"i.consumed_at IS NULL AND i.revoked_at IS NULL AND i.expires_at > now()"},
			"app.auth_resolve_enrollment(bytea, timestamptz)": {"es.credential_hash = p_hash AND es.revoked_at IS NULL AND es.expires_at > p_at AND u.status = 'active'"},
			"app.enrollment_admit(bytea, uuid, text, uuid)": {"app.ctx_org() IS NOT NULL OR app.ctx_group() IS NOT NULL",
				"app.rls_user_enabled_lock(u, 'organization', p_org)", "FOR SHARE;", "p_kind NOT IN ('lift', 'open', 'request')",
				"RAISE EXCEPTION 'enrollment_restricted' USING ERRCODE = 'P0001'", "r.created_at >= inv_created", "r.created_at < inv_created",
				"'org_member', 'membership'",
				// request kind: own row locked, decider still org_owner/org_admin, a newer restriction refuses, the covered rows
				// lifted with the decider as the lifter.
				"r.id = p_request AND r.organization_id = p_org AND r.user_id = u\n    FOR UPDATE;", "rb.role IN ('org_owner', 'org_admin')",
				"req_expires <= now()", "r.seq > req_seq", "r.seq <= req_seq", "lifted_by_user_id = req_decider, lifted_via = 'approved_request'",
				"SET consumed_at = now(), consumed_binding_id = bid"},
			"app.enrollment_request(bytea, uuid)": {"app.ctx_org() IS NOT NULL OR app.ctx_group() IS NOT NULL",
				"app.rls_user_enabled_lock(u, 'organization', p_org)", "FOR UPDATE;", "FOR SHARE;", "IS DISTINCT FROM 'request_access'",
				"r.status = 'pending'", "now() + interval '30 days'", "ON CONFLICT (organization_id, user_id) WHERE status = 'pending' DO NOTHING"},
			"app.enrollment_restrict_on_decision()": {"pg_trigger_depth() < 1", "TG_WHEN IS DISTINCT FROM 'AFTER'", "TG_LEVEL IS DISTINCT FROM 'ROW'",
				"TG_TABLE_NAME IS DISTINCT FROM 'enrollment_requests'", "TG_OP IS DISTINCT FROM 'UPDATE'",
				"app.enrollment_lock_organization(NEW.organization_id)",
				"NEW.status, 'trigger', coalesce(NEW.revoked_by_user_id, NEW.decided_by_user_id), NEW.id"},
			"app.enrollment_requests_immutable()": {"FOR NO KEY UPDATE;", "NEW.authority_seq := coalesce((SELECT max(r.seq)",
				"NEW.decided_by_user_id = app.ctx_user()", "NEW.revoked_by_user_id = app.ctx_user()", "OLD.consumed_at IS NULL",
				"'enrollment_request_transition'"},
			// Re-attestation correction: the backfill attestation's evidence is a departed backfill or replay row of p_org.
			"app.enrollment_readiness_set(uuid, text, text)": {"WHERE r.organization_id = p_org AND r.kind = 'departed'",
				"AND r.recorded_via IN ('backfill', 'replay')) THEN\n    RETURN 'no_backfill';"},
		} {
			got := body(sig)
			for _, f := range frags {
				if !strings.Contains(got, f) {
					t.Errorf("%s lacks %q", sig, f)
				}
			}
		}
		// The admission statement order equals the lock order: user lock, own credential FOR UPDATE, Organization FOR SHARE,
		// ledger rows FOR UPDATE, binding INSERT, admitted_* UPDATE.
		adm := body("app.enrollment_admit(bytea, uuid, text, uuid)")
		// Request slice: the request row FOR UPDATE after the Organization and before the ledger rows; its consumption after
		// the binding INSERT and before the admitted_* UPDATE.
		order := []string{"app.rls_user_enabled_lock(", "FROM app.enrollment_sessions es", "FOR SHARE;",
			"FROM app.enrollment_requests r WHERE r.id = p_request", "ORDER BY r.seq FOR UPDATE;",
			"INSERT INTO app.role_bindings", "UPDATE app.enrollment_requests r SET consumed_at", "UPDATE app.enrollment_sessions es SET admitted_at"}
		for i := 1; i < len(order); i++ {
			if a, b := strings.Index(adm, order[i-1]), strings.Index(adm, order[i]); a < 0 || b < 0 || a > b {
				t.Errorf("enrollment_admit: %q (%d) must precede %q (%d)", order[i-1], a, order[i], b)
			}
		}
		// The request statement order: user lock, own credential FOR UPDATE, Organization FOR SHARE, the one INSERT.
		req := body("app.enrollment_request(bytea, uuid)")
		reqOrder := []string{"app.rls_user_enabled_lock(", "FROM app.enrollment_sessions es", "FOR SHARE;", "INSERT INTO app.enrollment_requests"}
		for i := 1; i < len(reqOrder); i++ {
			if a, b := strings.Index(req, reqOrder[i-1]), strings.Index(req, reqOrder[i]); a < 0 || b < 0 || a > b {
				t.Errorf("enrollment_request: %q (%d) must precede %q (%d)", reqOrder[i-1], a, reqOrder[i], b)
			}
		}
		if got := one(`SELECT pg_get_triggerdef(oid) FROM pg_trigger WHERE tgrelid = 'app.role_bindings'::regclass
  AND tgname = 'role_bindings_enrollment_departure'`); !strings.Contains(got, "AFTER DELETE ON app.role_bindings FOR EACH ROW WHEN") ||
			!strings.Contains(got, "old.scope_type = 'organization'") || !strings.Contains(got, "old.source = 'membership'") ||
			!strings.HasSuffix(got, "EXECUTE FUNCTION app.enrollment_restrict_on_departure()") {
			t.Errorf("departure trigger: %s", got)
		}
		if got := one(`SELECT string_agg(t.tgname, ',' ORDER BY t.tgname) FROM pg_trigger t
WHERE t.tgrelid = 'app.role_bindings'::regclass AND NOT t.tgisinternal`); got !=
			"audit_capture_delete,audit_capture_insert,audit_capture_update,role_bindings_authority_guard,role_bindings_enrollment_departure,"+
				"role_bindings_keys_immutable" {
			t.Errorf("role_bindings triggers: %s", got)
		}
	})

	t.Run("CT-ENR-1_audit", func(t *testing.T) {
		conDef := func(name string) string {
			return one(`SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'audit.events'::regclass AND conname = $1`, name)
		}
		re := regexp.MustCompile(`'([a-z_.]+)'::text`)
		for _, c := range []struct {
			name string
			want int
			adds []string
		}{{"events_reason_valid", 68, enrReasons}, {"events_operation_valid", 64, append(slicesClone(enrOps), memOps...)},
			{"events_action_valid", 45, enrActions}, {"events_resource_type_valid", 11, []string{"invitation", "enrollment_request"}}} {
			def := conDef(c.name)
			got := map[string]bool{}
			for _, m := range re.FindAllStringSubmatch(def, -1) {
				got[m[1]] = true
			}
			if len(got) != c.want {
				t.Errorf("%s lists %d values, want %d", c.name, len(got), c.want)
			}
			for _, v := range c.adds {
				if !got[v] {
					t.Errorf("%s lacks %s", c.name, v)
				}
			}
		}
		b16 := memBody(t, "00016_organization_enrollment.sql", "CREATE OR REPLACE FUNCTION audit.events_attribute()")
		if got := one(`SELECT md5(prosrc) FROM pg_proc WHERE oid = 'audit.events_attribute()'::regprocedure`); got != md5hex(b16) {
			t.Error("events_attribute differs from the reviewed 00016 body")
		}
		// The reviewed delta, removed piece by piece (each exactly once), leaves the 00015 body: recipient_proof_unavailable
		// on auth.sign_in denied, the seven enrollment refusal reasons on access.denied, and the two narrowing clauses.
		stripped := b16
		for _, d := range []struct{ from, to string }{
			{"'organization_not_member', 'host_changed', 'recipient_proof_unavailable'))", "'organization_not_member', 'host_changed'))"},
			{"'resource_mismatch',\n        'enrollment_closed', 'enrollment_restricted', 'approval_revoked', 'invitation_mismatch', 'recipient_ineligible',\n" +
				"        'session_invalid', 'host_changed')", "'resource_mismatch')"},
			{enrShapeNarrowing, ""},
			// Slice 1b and the request slice: the six trigger-owned captures and their two shape branches.
			{"'invitation.issued', 'invitation.revoked', 'invitation.consumed',\n    'enrollment.policy_changed', 'enrollment.readiness_changed', " +
				"'enrollment.restriction_backfilled',\n    'enrollment.requested', 'enrollment.decided', 'enrollment.revoked');",
				"'invitation.issued', 'invitation.revoked', 'invitation.consumed');"},
			{enrReadinessShape, ""},
			{enrRequestShape, ""},
		} {
			if strings.Count(stripped, d.from) != 1 {
				t.Errorf("events_attribute delta piece found %d times, want 1: %q", strings.Count(stripped, d.from), d.from)
			}
			stripped = strings.Replace(stripped, d.from, d.to, 1)
		}
		if md5hex(stripped) != md5hex(memBody(t, "00015_membership_invitations.sql", "CREATE OR REPLACE FUNCTION audit.events_attribute()")) {
			t.Error("events_attribute without the reviewed enrollment delta differs from the 00015 body")
		}
		// Shape controls through the runtime role's real append path, each rolled back: the reviewed producers' rows are
		// admitted; an enrollment reason on another operation, with an Organization, on enrollment.admit without an
		// established actor, invitation_mismatch on enrollment.admit, and recipient_proof_unavailable outside auth.sign_in
		// denied are refused (23514 audit_event_shape).
		shape := func(user, action, outcome, reason, op, org string) string {
			t.Helper()
			tx, err := rt.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			rlsContext(t, tx, user, org, "", "")
			_, err = tx.Exec(ctx, `INSERT INTO audit.events (action, outcome, reason, operation, organization_id)
VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, '')::uuid)`, action, outcome, reason, op, org)
			return sqlstateOf(err)
		}
		for _, c := range []struct {
			name, user, action, outcome, reason, op, org, want string
		}{
			{"T2 session_invalid", enrMemberL, "access.denied", "denied", "session_invalid", "enrollment.admit", "", ""},
			{"T2 recipient_ineligible", enrMemberL, "access.denied", "denied", "recipient_ineligible", "enrollment.admit", "", ""},
			{"T2 host_changed", enrMemberL, "access.denied", "denied", "host_changed", "enrollment.admit", "", ""},
			{"T2 target_not_found", enrMemberL, "access.denied", "denied", "target_not_found", "enrollment.admit", "", ""},
			{"T1 start enrollment_closed attributed", enrMemberL, "access.denied", "denied", "enrollment_closed", "enrollment.start", "", ""},
			{"T1 start enrollment_closed unattributed", "", "access.denied", "denied", "enrollment_closed", "enrollment.start", "", ""},
			{"T1 start invitation_mismatch", enrMemberL, "access.denied", "denied", "invitation_mismatch", "enrollment.start", "", ""},
			{"T2 target_not_found with an Organization", enrOwnerL, "access.denied", "denied", "target_not_found", "enrollment.admit", enrOrgL, "23514"},
			{"T1 start with an Organization", enrOwnerL, "access.denied", "denied", "enrollment_closed", "enrollment.start", enrOrgL, "23514"},
			{"T1 start with a non-start reason", enrMemberL, "access.denied", "denied", "session_invalid", "enrollment.start", "", "23514"},
			{"T1 start with target_not_found", enrMemberL, "access.denied", "denied", "target_not_found", "enrollment.start", "", "23514"},
			{"proof failure", enrMemberL, "auth.sign_in", "denied", "recipient_proof_unavailable", "", "", ""},
			{"proof failure unattributed", "", "auth.sign_in", "denied", "recipient_proof_unavailable", "", "", ""},
			{"enrollment reason on another operation", enrMemberL, "access.denied", "denied", "session_invalid", "membership.remove", "", "23514"},
			{"host_changed on another operation", enrMemberL, "access.denied", "denied", "host_changed", "invitation.consume", "", "23514"},
			{"enrollment reason with an Organization", enrOwnerL, "access.denied", "denied", "enrollment_closed", "enrollment.admit", enrOrgL, "23514"},
			{"admit without an actor", "", "access.denied", "denied", "session_invalid", "enrollment.admit", "", "23514"},
			{"invitation_mismatch on admit", enrMemberL, "access.denied", "denied", "invitation_mismatch", "enrollment.admit", "", "23514"},
			{"admit with a non-enrollment reason", enrMemberL, "access.denied", "denied", "role_not_permitted", "enrollment.admit", "", "23514"},
			{"proof reason as failed", enrMemberL, "auth.sign_in", "failed", "recipient_proof_unavailable", "", "", "23514"},
			{"proof reason on access.denied", enrMemberL, "access.denied", "denied", "recipient_proof_unavailable", "enrollment.admit", "", "23514"},
			// Slice 1b: the captures are trigger-owned; a direct append is refused before its shape (42501 audit_event_source).
			{"direct enrollment.readiness_changed", enrMemberL, "enrollment.readiness_changed", "success", "operator_verified_absent", "", "", "42501"},
			{"direct enrollment.policy_changed", enrOwnerL, "enrollment.policy_changed", "success", "", "", enrOrgL, "42501"},
			// Request slice: the three request captures are trigger-owned as well.
			{"direct enrollment.requested", enrMemberL, "enrollment.requested", "success", "", "", "", "42501"},
			{"direct enrollment.decided", enrOwnerL, "enrollment.decided", "success", "", "", enrOrgL, "42501"},
			{"direct enrollment.revoked", enrOwnerL, "enrollment.revoked", "success", "", "", enrOrgL, "42501"},
		} {
			if got := shape(c.user, c.action, c.outcome, c.reason, c.op, c.org); got != c.want {
				t.Errorf("shape %s: sqlstate %q, want %q", c.name, got, c.want)
			}
		}
		// ENR-17b: the applicant's IDENTITY shape (a user, NULL ctx_org()) cannot append an Organization-attributed row
		// through the runtime role: an enrollment.requested row is refused by the trigger-owned source guard (42501
		// audit_event_source) and an ordinary descriptive row naming L by the scope guard (42501 audit_event_scope), both
		// ahead of events_insert's organization_id = ctx_org() check (00008, unchanged defense in depth); no row of L is
		// written.
		auditOfL := func() int {
			t.Helper()
			var n int
			if err := setup.QueryRow(ctx, `SELECT count(*)::int FROM audit.events WHERE organization_id = $1::uuid`, enrOrgL).Scan(&n); err != nil {
				t.Fatal(err)
			}
			return n
		}
		before := auditOfL()
		for action, guard := range map[string]string{"enrollment.requested": "audit_event_source", "access.denied": "audit_event_scope"} {
			func() {
				tx, err := rt.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				rlsContext(t, tx, enrMemberL, "", "", "")
				reason, op := "", ""
				if action == "access.denied" {
					reason, op = "target_not_found", "group.get"
				}
				_, err = tx.Exec(ctx, `INSERT INTO audit.events (action, outcome, reason, operation, organization_id)
VALUES ($1, CASE WHEN $1 = 'access.denied' THEN 'denied' ELSE 'success' END, NULLIF($2, ''), NULLIF($3, ''), $4::uuid)`, action, reason, op, enrOrgL)
				var pe *pgconn.PgError
				if !errors.As(err, &pe) || pe.Code != "42501" || pe.ConstraintName != guard {
					t.Errorf("IDENTITY-shape %s with organization_id set: %v, want 42501 %s", action, err, guard)
				}
			}()
		}
		if after := auditOfL(); after != before {
			t.Errorf("audit rows of L after the refused IDENTITY-shape appends: %d, want %d", after, before)
		}
	})

	t.Run("CT-ENR-1_negatives", func(t *testing.T) {
		// inTx runs f in a transaction of c that is always rolled back; role (if any) is SET LOCAL first (superuser conn).
		inTx := func(c *pgx.Conn, role, user string, f func(tx pgx.Tx)) {
			t.Helper()
			tx, err := c.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if role != "" {
				if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+quote(role)); err != nil {
					t.Fatal(err)
				}
			}
			if user != "" {
				rlsContext(t, tx, user, "", "", "")
			}
			f(tx)
		}
		ledgerInsert := `INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via) VALUES ($1, $2, 'departed', 'trigger')`
		inTx(own, "", "", func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, ledgerInsert, enrOrgL, enrMemberL); sqlstateOf(err) != "42501" || !strings.Contains(err.Error(), "row-level security") {
				t.Errorf("owner ledger INSERT at depth 0: %v, want 42501 policy violation", err)
			}
		})
		inTx(rt, "", enrMemberL, func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, ledgerInsert, enrOrgL, enrMemberL); sqlstateOf(err) != "42501" {
				t.Errorf("runtime ledger INSERT: %v, want 42501", err)
			}
		})
		// ENR-22b: no other session role inserts either (no INSERT grant or policy for cw, pr or ar).
		for _, role := range []string{cw, pr, ar} {
			inTx(setup, role, enrMemberL, func(tx pgx.Tx) {
				if _, err := tx.Exec(ctx, ledgerInsert, enrOrgL, enrMemberL); sqlstateOf(err) != "42501" {
					t.Errorf("%s ledger INSERT: %v, want 42501", role, err)
				}
			})
		}
		inTx(own, "", "", func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, `SELECT app.enrollment_restrict_on_departure()`); err == nil ||
				(sqlstateOf(err) != "0A000" && sqlstateOf(err) != "42501") {
				t.Errorf("direct departure body call: %v, want a refusal (0A000/42501)", err)
			}
		})
		inTx(rt, "", enrMemberL, func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, `SELECT app.enrollment_restrict_on_departure()`); sqlstateOf(err) != "42501" {
				t.Errorf("runtime departure body call: %v, want 42501", err)
			}
		})
		inTx(rt, "", enrMemberL, func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, `SELECT app.enrollment_lock_organization($1)`, enrOrgL); sqlstateOf(err) != "42501" {
				t.Errorf("runtime lock helper call: %v, want 42501", err)
			}
		})
		inTx(setup, cw, enrMemberL, func(tx pgx.Tx) {
			// Slice 1b: a mode change is refused earlier by the BEFORE guard (owner rule), so the lock-only policy's WITH
			// CHECK (false) is observed with a same-value write the guard lets through.
			if _, err := tx.Exec(ctx, `UPDATE app.organizations SET enrollment_mode = enrollment_mode WHERE id = $1`, enrOrgL); sqlstateOf(err) != "42501" ||
				!strings.Contains(err.Error(), "row-level security") {
				t.Errorf("claim-writer mode UPDATE: %v, want 42501 policy violation (WITH CHECK false)", err)
			}
		})
		inTx(setup, cw, enrMemberL, func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, `SELECT name FROM app.organizations WHERE id = $1`, enrOrgL); sqlstateOf(err) != "42501" {
				t.Errorf("claim-writer name read: %v, want 42501", err)
			}
		})
		inTx(setup, cw, enrMemberL, func(tx pgx.Tx) {
			var n int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM (SELECT id, enrollment_mode FROM app.organizations WHERE id = $1 FOR SHARE) s`,
				enrOrgL).Scan(&n); err != nil || n != 1 {
				t.Errorf("claim-writer FOR SHARE: %d %v, want one row", n, err)
			}
		})
		// The owner path (superuser delete, as the G013 replay): exactly one departed row, trigger-recorded, no creator.
		deleteMember := `DELETE FROM app.role_bindings WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = 'organization'`
		inTx(setup, "", "", func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, deleteMember, enrOrgL, enrMemberL); err != nil {
				t.Fatalf("superuser departure: %v", err)
			}
			var got string
			if err := tx.QueryRow(ctx, `SELECT string_agg(format('%s|%s|%s|%s|%s', kind, recorded_via, created_by_user_id IS NULL,
  lifted_at IS NULL, seq > 0), ',') FROM app.enrollment_restrictions WHERE organization_id = $1 AND user_id = $2`, enrOrgL, enrMemberL).Scan(&got); err != nil ||
				got != "departed|trigger|t|t|t" {
				t.Errorf("departure ledger row: %q %v", got, err)
			}
			// ENR-22e: the owner (bound by FORCE RLS) sees the subject's row under the subject's context but has no UPDATE
			// or DELETE policy: both affect no row (or are refused) and the row is unchanged.
			if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+quote(o)); err != nil {
				t.Fatal(err)
			}
			rlsContext(t, tx, enrMemberL, "", "", "")
			var seen int
			if err := tx.QueryRow(ctx, `SELECT count(*)::int FROM app.enrollment_restrictions WHERE organization_id = $1 AND user_id = $2`,
				enrOrgL, enrMemberL).Scan(&seen); err != nil || seen != 1 {
				t.Errorf("owner read of the subject's ledger row: %d %v, want 1", seen, err)
			}
			for _, w := range []string{
				`UPDATE app.enrollment_restrictions SET lifted_at = now(), lifted_via = 'invitation' WHERE organization_id = $1 AND user_id = $2`,
				`DELETE FROM app.enrollment_restrictions WHERE organization_id = $1 AND user_id = $2`,
			} {
				sp, err := tx.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				tag, err := sp.Exec(ctx, w, enrOrgL, enrMemberL)
				if (err == nil && tag.RowsAffected() != 0) || (err != nil && sqlstateOf(err) != "42501") {
					t.Errorf("owner ledger write %q: %d rows %v, want 0 rows or 42501", w, tag.RowsAffected(), err)
				}
				_ = sp.Rollback(ctx)
			}
			if _, err := tx.Exec(ctx, `RESET ROLE`); err != nil {
				t.Fatal(err)
			}
			if err := tx.QueryRow(ctx, `SELECT string_agg(format('%s|%s|%s|%s|%s', kind, recorded_via, created_by_user_id IS NULL,
  lifted_at IS NULL, seq > 0), ',') FROM app.enrollment_restrictions WHERE organization_id = $1 AND user_id = $2`, enrOrgL, enrMemberL).Scan(&got); err != nil ||
				got != "departed|trigger|t|t|t" {
				t.Errorf("ledger row after the owner writes: %q %v, want it unchanged", got, err)
			}
		})
		// Atomicity: without the owner policy the deleting statement itself fails and nothing is removed.
		inTx(setup, "", "", func(tx pgx.Tx) {
			if _, err := tx.Exec(ctx, `DROP POLICY enrollment_restrictions_owner_insert ON app.enrollment_restrictions`); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, deleteMember, enrOrgL, enrMemberL); sqlstateOf(err) != "42501" {
				t.Errorf("departure without the owner policy: %v, want 42501 (the delete rolls back)", err)
			}
		})
		var n int
		if err := setup.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.role_bindings WHERE organization_id = $1 AND subject_user_id = $2)
  + 10 * (SELECT count(*) FROM app.enrollment_restrictions)`, enrOrgL, enrMemberL).Scan(&n); err != nil || n != 1 {
			t.Errorf("after the rolled-back negatives: binding+10*ledger = %d %v, want 1", n, err)
		}
	})

	// Slice 1b (addendum §A.6 ENR-3 / ENR-12/13 / ENR-22 SQL extensions, correction 01 §C.2/§C.5/§C.6(i)): the readiness
	// guard, the operator setter and backfill on the legacy Organization L (closed at 00016) and the new Organization N
	// (ready by construction). Every case runs in its own transaction that is always rolled back, so the up/down state
	// below is unchanged.
	t.Run("CT-ENR-1_readiness", func(t *testing.T) {
		rb := func(c *pgx.Conn, role, user, org string, f func(tx pgx.Tx)) {
			t.Helper()
			tx, err := c.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if role != "" {
				if _, err := tx.Exec(ctx, "SET LOCAL ROLE "+quote(role)); err != nil {
					t.Fatal(err)
				}
			}
			if user != "" {
				rlsContext(t, tx, user, org, "", "")
			}
			f(tx)
		}
		val := func(tx pgx.Tx, sql string, args ...any) (string, string) {
			t.Helper()
			var s *string
			err := tx.QueryRow(ctx, sql, args...).Scan(&s)
			if err != nil || s == nil {
				return "", sqlstateOf(err)
			}
			return *s, ""
		}
		events := func(tx pgx.Tx, org string) string {
			t.Helper()
			s, code := val(tx, `SELECT coalesce(string_agg(format('%s|%s|%s|%s|%s', action, coalesce(reason, '-'), actor_kind,
  coalesce(request_id, '-'), coalesce(subject_user_id::text, '-')), ',' ORDER BY occurred_at, action), '')
FROM audit.events WHERE organization_id = $1::uuid AND action LIKE 'enrollment.%'`, org)
			if code != "" {
				t.Fatalf("events: %s", code)
			}
			return s
		}
		state := func(tx pgx.Tx, org string) string {
			t.Helper()
			s, _ := val(tx, `SELECT format('%s|%s', enrollment_mode, coalesce(enrollment_history_basis, '-')) FROM app.organizations WHERE id = $1::uuid`, org)
			return s
		}
		set := `SELECT app.enrollment_readiness_set($1::uuid, $2, $3)`
		backfill := `SELECT app.enrollment_history_backfill($1::uuid, $2::uuid, $3)`

		// ENR-3: no_backfill; resolved + one audited row (actor unknown, request_id = the reference); unchanged on repeat;
		// the owner may then open; a close of the open Organization forces invite_only and writes two rows.
		rb(setup, "", "", "", func(tx pgx.Tx) {
			if got, code := val(tx, set, enrOrgL, "operator_backfilled", "l1.ct-enr-1"); got != "no_backfill" {
				t.Errorf("attest backfilled without a backfill row: %q %s", got, code)
			}
			// Re-attestation correction: only L's own departed backfill/replay rows count. Another Organization's backfill and
			// replay rows and L's own trigger row (a live departure) still give no_backfill; rolled back to the empty L.
			sp, err := tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := sp.Exec(ctx, `INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via) VALUES
  ($1, $3, 'departed', 'backfill'), ($1, $3, 'departed', 'replay'), ($2, $3, 'departed', 'trigger')`, enrOrgN, enrOrgL, enrMemberL); err != nil {
				t.Fatalf("foreign/trigger rows: %v", err)
			}
			if got, code := val(sp, set, enrOrgL, "operator_backfilled", "l1.ct-enr-1"); got != "no_backfill" {
				t.Errorf("attest backfilled with foreign-Organization and trigger rows only: %q %s", got, code)
			}
			if err := sp.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			if got, code := val(tx, set, enrOrgL, "operator_verified_absent", "ct-enr-ref"); got != "resolved" {
				t.Fatalf("attest verified absence: %q %s", got, code)
			}
			if got := events(tx, enrOrgL); got != "enrollment.readiness_changed|operator_verified_absent|unknown|ct-enr-ref|-" {
				t.Errorf("attestation audit: %s", got)
			}
			if got, _ := val(tx, set, enrOrgL, "operator_verified_absent", "ct-enr-ref"); got != "unchanged" {
				t.Errorf("repeat attestation: %q", got)
			}
			if _, err := tx.Exec(ctx, `UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = $1`, enrOrgL); err != nil {
				t.Fatalf("open after attestation: %v", err)
			}
			if got, _ := val(tx, set, enrOrgL, nil, "ct-enr-close"); got != "closed" || state(tx, enrOrgL) != "invite_only|-" {
				t.Errorf("close of an open Organization: %q, state %s", got, state(tx, enrOrgL))
			}
			if got := events(tx, enrOrgL); !strings.HasSuffix(got, ",enrollment.policy_changed|-|unknown|ct-enr-close|-,"+
				"enrollment.readiness_changed|history_unresolved|unknown|ct-enr-close|-") {
				t.Errorf("close audit (want the policy and readiness rows): %s", got)
			}
			if got, _ := val(tx, set, enrOrgL, nil, "ct-enr-close"); got != "unchanged" {
				t.Errorf("repeat close: %q", got)
			}
			if got, _ := val(tx, set, "0f58e000-0000-4000-8000-0000000000ff", "operator_verified_absent", "ct-enr-ref"); got != "absent" {
				t.Errorf("unknown Organization: %q", got)
			}
		})
		// The guard: a superuser readiness write outside the procedure, the owner (FORCE-bound, in L's owner context) and the
		// runtime (no column privilege) are refused; the CHECK refuses open while unresolved for the runtime owner, the
		// owner role and the superuser alike; a runtime member may not change the mode; the runtime owner may.
		for _, c := range []struct {
			name, sql string
			conn      *pgx.Conn
			user      string
			want      string
		}{
			{"superuser readiness without the marker", `UPDATE app.organizations SET enrollment_history_resolved_at = now(), enrollment_history_basis = 'operator_verified_absent' WHERE id = $1`, setup, "", "42501"},
			{"owner readiness", `UPDATE app.organizations SET enrollment_history_resolved_at = now(), enrollment_history_basis = 'operator_verified_absent' WHERE id = $1`, own, enrOwnerL, "42501"},
			{"runtime readiness", `UPDATE app.organizations SET enrollment_history_resolved_at = now() WHERE id = $1`, rt, enrOwnerL, "42501"},
			{"runtime owner opens unresolved", `UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = $1`, rt, enrOwnerL, "23514"},
			{"owner role opens unresolved", `UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = $1`, own, enrOwnerL, "23514"},
			{"superuser opens unresolved", `UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = $1`, setup, "", "23514"},
			{"runtime member changes the mode", `UPDATE app.organizations SET enrollment_mode = 'request_access' WHERE id = $1`, rt, enrMemberL, "42501"},
			{"runtime owner request_access", `UPDATE app.organizations SET enrollment_mode = 'request_access' WHERE id = $1`, rt, enrOwnerL, ""},
		} {
			rb(c.conn, "", c.user, enrOrgL, func(tx pgx.Tx) {
				_, err := tx.Exec(ctx, c.sql, enrOrgL)
				if got := sqlstateOf(err); got != c.want {
					t.Errorf("%s: sqlstate %q (%v), want %q", c.name, got, err, c.want)
				}
				if c.want == "" && events(tx, enrOrgL) != "enrollment.policy_changed|-|user|-|-" {
					t.Errorf("%s: audit %s", c.name, events(tx, enrOrgL))
				}
			})
		}
		// EXECUTE held by no session role (42501 before the body); the owner may call it but is refused in the body; NULL
		// arguments raise 22004 before the grammar; a malformed reference or a plain reference for a backfill basis 22023.
		for _, c := range []struct {
			name, role, sql string
			args            []any
			want            string
		}{
			{"runtime setter", rtr, set, []any{enrOrgL, "operator_verified_absent", "ct-enr-ref"}, "42501"},
			{"claim writer backfill", cw, backfill, []any{enrOrgL, enrMemberL, "l1.ct-enr"}, "42501"},
			{"policy reader setter", pr, set, []any{enrOrgL, nil, "ct-enr-ref"}, "42501"},
			{"auth reader backfill", ar, backfill, []any{enrOrgL, enrMemberL, "l1.ct-enr"}, "42501"},
			{"owner setter", o, set, []any{enrOrgL, "operator_verified_absent", "ct-enr-ref"}, "42501"},
			{"setter NULL reference", "", set, []any{enrOrgL, "operator_verified_absent", nil}, "22004"},
			{"setter NULL Organization", "", set, []any{nil, nil, "ct-enr-ref"}, "22004"},
			{"setter malformed reference", "", set, []any{enrOrgL, "operator_verified_absent", "bad ref"}, "22023"},
			{"setter plain reference for a backfill basis", "", set, []any{enrOrgL, "operator_backfilled", "plain-ref"}, "22023"},
			{"setter unknown basis", "", set, []any{enrOrgL, "created_after_ledger", "ct-enr-ref"}, "22023"},
			{"backfill NULL reference", "", backfill, []any{enrOrgL, enrMemberL, nil}, "22004"},
			{"backfill NULL user", "", backfill, []any{enrOrgL, nil, "l1.ct-enr"}, "22004"},
			{"backfill reference without a ledger sequence", "", backfill, []any{enrOrgL, enrMemberL, "noseq-ref"}, "22023"},
		} {
			rb(setup, c.role, "", "", func(tx pgx.Tx) {
				if _, code := val(tx, c.sql, c.args...); code != c.want {
					t.Errorf("%s: sqlstate %q, want %q", c.name, code, c.want)
				}
			})
		}
		// ENR-12/13 backfill outcomes and the live-side close (correction 01 §C.2): absent, not_legacy (N), member (a
		// current binding, no write), inserted (one departed/backfill row without a creator, one audited row carrying the
		// ledger reference, and an attested open L closed to invite_only with two rows), present on repeat.
		rb(setup, "", "", "", func(tx pgx.Tx) {
			const ux = "0f58e000-0000-4000-8000-000000000013"
			if _, err := tx.Exec(ctx, `INSERT INTO app.users (id, display_name) VALUES ($1, 'Enr Departed X')`, ux); err != nil {
				t.Fatal(err)
			}
			for _, c := range []struct{ name, org, user, want string }{
				{"unknown Organization", "0f58e000-0000-4000-8000-0000000000ff", ux, "absent"},
				{"unknown user", enrOrgL, "0f58e000-0000-4000-8000-0000000000fe", "absent"},
				{"created after the ledger", enrOrgN, ux, "not_legacy"},
				{"current member", enrOrgL, enrMemberL, "member"},
			} {
				if got, code := val(tx, backfill, c.org, c.user, "l3.ct-enr"); got != c.want {
					t.Errorf("backfill %s: %q %s, want %s", c.name, got, code, c.want)
				}
			}
			if got, _ := val(tx, set, enrOrgL, "operator_verified_absent", "ct-enr-ref"); got != "resolved" {
				t.Fatalf("attest before the backfill: %q", got)
			}
			if _, err := tx.Exec(ctx, `UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = $1`, enrOrgL); err != nil {
				t.Fatal(err)
			}
			before := events(tx, enrOrgL)
			if got, code := val(tx, backfill, enrOrgL, ux, "l7.ct-enr"); got != "inserted" {
				t.Fatalf("backfill: %q %s", got, code)
			}
			if got, _ := val(tx, `SELECT format('%s|%s|%s|%s|%s', kind, recorded_via, created_by_user_id IS NULL, lifted_at IS NULL, count(*) OVER ())
FROM app.enrollment_restrictions WHERE organization_id = $1::uuid AND user_id = $2::uuid`, enrOrgL, ux); got != "departed|backfill|t|t|1" {
				t.Errorf("backfilled row: %s", got)
			}
			if got := strings.TrimPrefix(events(tx, enrOrgL), before+","); got != "enrollment.policy_changed|-|unknown|l7.ct-enr|-,"+
				"enrollment.readiness_changed|history_unresolved|unknown|l7.ct-enr|-,enrollment.restriction_backfilled|operator_backfilled|unknown|l7.ct-enr|"+ux &&
				got != "enrollment.restriction_backfilled|operator_backfilled|unknown|l7.ct-enr|"+ux+",enrollment.policy_changed|-|unknown|l7.ct-enr|-,"+
					"enrollment.readiness_changed|history_unresolved|unknown|l7.ct-enr|-" {
				t.Errorf("backfill audit: %s", got)
			}
			if state(tx, enrOrgL) != "invite_only|-" {
				t.Errorf("live-side close: %s", state(tx, enrOrgL))
			}
			if got, _ := val(tx, backfill, enrOrgL, ux, "l7.ct-enr"); got != "present" {
				t.Errorf("repeat backfill: %q", got)
			}
			if got, _ := val(tx, set, enrOrgL, "operator_backfilled", "l7.ct-enr"); got != "resolved" {
				t.Errorf("attest backfilled after the pair: %q", got)
			}
		})
		// ENR-22 extension: a backfill row never passes the owner's trigger-bound INSERT policy, and no session role inserts.
		for _, c := range []struct {
			name string
			conn *pgx.Conn
		}{{"owner", own}, {"runtime", rt}} {
			rb(c.conn, "", enrOwnerL, enrOrgL, func(tx pgx.Tx) {
				_, err := tx.Exec(ctx, `INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via)
VALUES ($1, $2, 'departed', 'backfill')`, enrOrgL, enrMemberL)
				if got := sqlstateOf(err); got != "42501" {
					t.Errorf("%s backfill INSERT: %q (%v), want 42501", c.name, got, err)
				}
			})
		}
	})

	// Request slice (Plan r3 §1.4/§4.3-§4.5; Codex round-8 dispositions 1-3): the SQL request lifecycle through the real
	// runtime and claim-writer paths. Every case runs in its own transaction of the superuser fixture connection that is
	// always rolled back (the role and context are switched inside it; expected refusals run under a savepoint), so the
	// up/down state below is unchanged. The application facades and their ordering proofs are c2's.
	t.Run("CT-ENR-1_requests", func(t *testing.T) {
		const (
			ua  = "0f58e000-0000-4000-8000-000000000021" // the applicant (not a member of L)
			uad = "0f58e000-0000-4000-8000-000000000022" // an org_admin of L
		)
		cred := bytes.Repeat([]byte{0xa1}, 32)
		type env struct {
			tx pgx.Tx
			// as switches to role (none: the superuser) with the given user/Organization context.
			as func(role, user, org string)
			// try runs one statement under a savepoint: rows affected and the SQLSTATE ("" on success).
			try func(sql string, args ...any) (int64, string)
			// val reads one text value as the current role ("" when NULL or refused).
			val func(sql string, args ...any) string
		}
		run := func(name string, f func(e env)) {
			t.Helper()
			tx, err := setup.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			e := env{tx: tx}
			e.as = func(role, user, org string) {
				t.Helper()
				if _, err := tx.Exec(ctx, `RESET ROLE`); err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				if role != "" {
					if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+quote(role)); err != nil {
						t.Fatalf("%s: %v", name, err)
					}
				}
				rlsContext(t, tx, user, org, "", "")
			}
			e.try = func(sql string, args ...any) (int64, string) {
				t.Helper()
				sp, err := tx.Begin(ctx)
				if err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				tag, err := sp.Exec(ctx, sql, args...)
				if err != nil {
					_ = sp.Rollback(ctx)
					return 0, sqlstateOf(err)
				}
				if err := sp.Commit(ctx); err != nil {
					t.Fatalf("%s: %v", name, err)
				}
				return tag.RowsAffected(), ""
			}
			e.val = func(sql string, args ...any) string {
				t.Helper()
				var s *string
				if err := tx.QueryRow(ctx, sql, args...).Scan(&s); err != nil {
					t.Fatalf("%s: %s: %v", name, sql, err)
				}
				if s == nil {
					return ""
				}
				return *s
			}
			// The fixture: the applicant and an org_admin of L, L in request_access, and the applicant's live enrollment
			// credential for L.
			e.as("", "", "")
			for _, s := range []string{
				`INSERT INTO app.users (id, display_name) VALUES ('` + ua + `', 'Enr Applicant'), ('` + uad + `', 'Enr Admin L')`,
				`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + enrOrgL + `', '` + uad + `', 'organization', '` + enrOrgL + `', NULL, 'org_admin', 'membership')`,
				`UPDATE app.organizations SET enrollment_mode = 'request_access' WHERE id = '` + enrOrgL + `'`,
			} {
				if _, err := tx.Exec(ctx, s); err != nil {
					t.Fatalf("%s fixture: %v", name, err)
				}
			}
			if _, err := tx.Exec(ctx, `INSERT INTO app.enrollment_sessions (id, user_id, organization_id, origin, credential_hash, csrf_token,
  provider_session_id, expires_at) VALUES (gen_random_uuid(), $1, $2, 'https://l.example.test', $3, $4, 'ct-enr-ps', now() + interval '10 minutes')`,
				ua, enrOrgL, cred, bytes.Repeat([]byte{0xb2}, 32)); err != nil {
				t.Fatalf("%s fixture session: %v", name, err)
			}
			f(e)
		}
		// request returns "outcome|request_id" for the applicant (runtime role, IDENTITY shape).
		request := func(e env) string {
			e.as(rtr, ua, "")
			return e.val(`SELECT outcome || '|' || coalesce(request_id::text, '-') FROM app.enrollment_request($1, $2::uuid)`, cred, enrOrgL)
		}
		admit := func(e env, rid string) string {
			e.as(rtr, ua, "")
			return e.val(`SELECT outcome || '|' || coalesce(binding_id::text, '-') FROM app.enrollment_admit($1, $2::uuid, 'request', $3::uuid)`,
				cred, enrOrgL, rid)
		}
		decide := func(e env, decider, rid, status string) (int64, string) {
			e.as(rtr, decider, enrOrgL)
			if status == "revoked" {
				return e.try(`UPDATE app.enrollment_requests SET status = 'revoked', revoked_at = now(), revoked_by_user_id = $2 WHERE id = $1`, rid, decider)
			}
			return e.try(`UPDATE app.enrollment_requests SET status = $3, decided_at = now(), decided_by_user_id = $2 WHERE id = $1`, rid, decider, status)
		}
		// events lists the request events of L (superuser read): action|reason|actor_kind|actor|resource_type|resource=rid.
		events := func(e env, rid string) string {
			e.as("", "", "")
			return e.val(`SELECT coalesce(string_agg(format('%s|%s|%s|%s|%s|%s', action, coalesce(reason, '-'), actor_kind,
  coalesce(actor_user_id::text, '-'), resource_type, resource_id = $2::uuid), ',' ORDER BY occurred_at), '')
FROM audit.events WHERE organization_id = $1::uuid AND action IN ('enrollment.requested', 'enrollment.decided', 'enrollment.revoked')`, enrOrgL, rid)
		}
		ridOf := func(t *testing.T, got string) string {
			t.Helper()
			out, id, ok := strings.Cut(got, "|")
			if !ok || out != "pending" || id == "-" {
				t.Fatalf("request: %q, want pending with an id", got)
			}
			return id
		}

		// Approval and admission: idempotent while pending; authority_seq derived from the ledger (the supplied decision
		// carries none); the approval consumed once, the covered restriction lifted by the decider, one binding; the
		// consumption is not captured; a second admission reconciles.
		run("approve", func(e env) {
			e.as("", "", "")
			s1 := e.val(`INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via) VALUES ($1, $2, 'departed', 'replay')
RETURNING seq::text`, enrOrgL, ua)
			rid := ridOf(t, request(e))
			if got := request(e); got != "pending|"+rid {
				t.Errorf("repeat request: %q, want the same pending id", got)
			}
			if n, code := decide(e, enrOwnerL, rid, "approved"); n != 1 || code != "" {
				t.Fatalf("approve: %d %s", n, code)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s|%s', status, authority_seq, decided_by_user_id) FROM app.enrollment_requests WHERE id = $1`, rid); got !=
				"approved|"+s1+"|"+enrOwnerL {
				t.Errorf("approved row: %s, want authority_seq %s", got, s1)
			}
			out := admit(e, rid)
			bid, ok := strings.CutPrefix(out, "admitted|")
			if !ok {
				t.Fatalf("admit: %q", out)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s|%s|%s', (SELECT role FROM app.role_bindings WHERE id = $2::uuid AND subject_user_id = $3),
  (SELECT consumed_binding_id = $2::uuid FROM app.enrollment_requests WHERE id = $1::uuid),
  (SELECT string_agg(format('%s/%s', lifted_via, lifted_by_user_id), ',') FROM app.enrollment_restrictions WHERE user_id = $3),
  (SELECT admitted_binding_id = $2::uuid FROM app.enrollment_sessions WHERE user_id = $3))`, rid, bid, ua); got !=
				"org_member|t|approved_request/"+enrOwnerL+"|t" {
				t.Errorf("admission writes: %s", got)
			}
			if got := events(e, rid); got != "enrollment.requested|-|user|"+ua+"|enrollment_request|t,"+
				"enrollment.decided|-|user|"+enrOwnerL+"|enrollment_request|t" {
				t.Errorf("request events: %s", got)
			}
			if got := e.val(`SELECT count(*)::text FROM audit.events WHERE action = 'membership.granted' AND subject_user_id = $1`, ua); got != "1" {
				t.Errorf("membership.granted rows: %s", got)
			}
			if got := admit(e, rid); got != "already_member|"+bid {
				t.Errorf("second admission: %q", got)
			}
		})
		// Denial: a ledger row (trigger, the decider, the request), one decided event without a reason (disposition 2),
		// admission denied; a new request is admitted as a new pending row (a denial refuses open signup, not a request).
		run("deny", func(e env) {
			rid := ridOf(t, request(e))
			if n, code := decide(e, enrOwnerL, rid, "denied"); n != 1 || code != "" {
				t.Fatalf("deny: %d %s", n, code)
			}
			e.as("", "", "")
			if got := e.val(`SELECT string_agg(format('%s|%s|%s|%s|%s', kind, recorded_via, created_by_user_id, source_request_id = $2::uuid,
  lifted_at IS NULL), ',') FROM app.enrollment_restrictions WHERE user_id = $1`, ua, rid); got != "denied|trigger|"+enrOwnerL+"|t|t" {
				t.Errorf("denial ledger: %s", got)
			}
			if got := events(e, rid); got != "enrollment.requested|-|user|"+ua+"|enrollment_request|t,"+
				"enrollment.decided|-|user|"+enrOwnerL+"|enrollment_request|t" {
				t.Errorf("denial events: %s", got)
			}
			if got := admit(e, rid); got != "denied|-" {
				t.Errorf("admit after denial: %q", got)
			}
			if got := ridOf(t, request(e)); got == rid {
				t.Errorf("request after denial reused the denied id")
			}
		})
		// Revocation of an unconsumed approval: a revoked ledger row by the revoker, one revoked event, admission refused.
		run("revoke", func(e env) {
			rid := ridOf(t, request(e))
			if n, code := decide(e, enrOwnerL, rid, "approved"); n != 1 || code != "" {
				t.Fatalf("approve: %d %s", n, code)
			}
			if n, code := decide(e, uad, rid, "revoked"); n != 1 || code != "" {
				t.Fatalf("revoke: %d %s", n, code)
			}
			e.as("", "", "")
			if got := e.val(`SELECT string_agg(format('%s|%s|%s', kind, created_by_user_id, source_request_id = $2::uuid), ',')
FROM app.enrollment_restrictions WHERE user_id = $1`, ua, rid); got != "revoked|"+uad+"|t" {
				t.Errorf("revocation ledger: %s", got)
			}
			if got := events(e, rid); !strings.HasSuffix(got, ",enrollment.revoked|-|user|"+uad+"|enrollment_request|t") ||
				strings.Count(got, ",") != 2 {
				t.Errorf("revocation events: %s", got)
			}
			if got := admit(e, rid); got != "approval_revoked|-" {
				t.Errorf("admit after revocation: %q", got)
			}
		})
		// A restriction newer than the approval's authority version refuses with zero writes.
		run("newer restriction", func(e env) {
			rid := ridOf(t, request(e))
			if n, code := decide(e, enrOwnerL, rid, "approved"); n != 1 || code != "" {
				t.Fatalf("approve: %d %s", n, code)
			}
			e.as("", "", "")
			e.val(`INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via) VALUES ($1, $2, 'departed', 'replay')
RETURNING seq::text`, enrOrgL, ua)
			if got := admit(e, rid); got != "enrollment_restricted|-" {
				t.Errorf("admit under a newer restriction: %q", got)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s|%s|%s', (SELECT count(*) FROM app.role_bindings WHERE subject_user_id = $1),
  (SELECT consumed_at IS NULL FROM app.enrollment_requests WHERE id = $2::uuid),
  (SELECT bool_and(lifted_at IS NULL) FROM app.enrollment_restrictions WHERE user_id = $1),
  (SELECT admitted_at IS NULL FROM app.enrollment_sessions WHERE user_id = $1))`, ua, rid); got != "0|t|t|t" {
				t.Errorf("writes after the refusal: %s, want none", got)
			}
		})
		// An approval whose decider no longer manages the Organization is not current authority.
		run("departed decider", func(e env) {
			rid := ridOf(t, request(e))
			if n, code := decide(e, uad, rid, "approved"); n != 1 || code != "" {
				t.Fatalf("approve as admin: %d %s", n, code)
			}
			e.as("", "", "")
			if _, err := e.tx.Exec(ctx, `DELETE FROM app.role_bindings WHERE subject_user_id = $1`, uad); err != nil {
				t.Fatal(err)
			}
			if got := admit(e, rid); got != "approval_revoked|-" {
				t.Errorf("admit after the decider left: %q", got)
			}
		})
		// Disposition 3: an expired pending request keeps occupying one_pending, is returned as pending (never renewed:
		// same id and expiry, no new row or event), may still be decided, and its approval is never consumable.
		run("expired pending", func(e env) {
			e.as("", ua, "")
			rid := e.val(`INSERT INTO app.enrollment_requests (id, organization_id, user_id, status, created_at, expires_at)
VALUES (gen_random_uuid(), $1, $2, 'pending', now() - interval '30 days', now() - interval '1 second') RETURNING id::text`, enrOrgL, ua)
			if got := request(e); got != "pending|"+rid {
				t.Errorf("request with an expired pending row: %q, want the same id", got)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s', count(*), bool_and(expires_at < now())) FROM app.enrollment_requests WHERE user_id = $1`, ua); got != "1|t" {
				t.Errorf("expired request after the repeat: %s, want one unrenewed row", got)
			}
			if n, code := decide(e, enrOwnerL, rid, "approved"); n != 1 || code != "" {
				t.Fatalf("approve the expired request: %d %s", n, code)
			}
			if got := admit(e, rid); got != "approval_revoked|-" {
				t.Errorf("admit an expired approval: %q", got)
			}
		})
		// Privilege and transition negatives (each refused, nothing changed).
		run("negatives", func(e env) {
			rid := ridOf(t, request(e))
			for _, c := range []struct {
				name, role, user, org, sql string
				args                       []any
				want                       string
			}{
				{"runtime INSERT", rtr, ua, "", `INSERT INTO app.enrollment_requests (id, organization_id, user_id, status, expires_at)
VALUES (gen_random_uuid(), $1, $2, 'pending', now() + interval '1 day')`, []any{enrOrgL, ua}, "42501"},
				{"runtime DELETE", rtr, enrOwnerL, enrOrgL, `DELETE FROM app.enrollment_requests WHERE id = $1`, []any{rid}, "42501"},
				{"runtime authority_seq", rtr, enrOwnerL, enrOrgL, `UPDATE app.enrollment_requests SET authority_seq = 5 WHERE id = $1`, []any{rid}, "42501"},
				{"runtime consumption", rtr, enrOwnerL, enrOrgL, `UPDATE app.enrollment_requests SET consumed_at = now() WHERE id = $1`, []any{rid}, "42501"},
				{"claim-writer decision", cw, ua, "", `UPDATE app.enrollment_requests SET status = 'approved' WHERE id = $1`, []any{rid}, "42501"},
				{"decision attributed to another manager", rtr, enrOwnerL, enrOrgL,
					`UPDATE app.enrollment_requests SET status = 'approved', decided_at = now(), decided_by_user_id = $2 WHERE id = $1`, []any{rid, uad}, "42501"},
				{"pending -> revoked", rtr, enrOwnerL, enrOrgL,
					`UPDATE app.enrollment_requests SET status = 'revoked', revoked_at = now(), revoked_by_user_id = $2 WHERE id = $1`, []any{rid, enrOwnerL}, "42501"},
				{"request with an Organization context", rtr, ua, enrOrgL, `SELECT * FROM app.enrollment_request($1, $2::uuid)`, []any{cred, enrOrgL}, "42501"},
				{"runtime decision body", rtr, ua, "", `SELECT app.enrollment_restrict_on_decision()`, nil, "42501"},
			} {
				e.as(c.role, c.user, c.org)
				if _, code := e.try(c.sql, c.args...); code != c.want {
					t.Errorf("%s: sqlstate %q, want %q", c.name, code, c.want)
				}
			}
			e.as("", "", "")
			if _, code := e.try(`SELECT app.enrollment_restrict_on_decision()`); code != "0A000" && code != "42501" {
				t.Errorf("direct decision body call: %q, want a refusal (0A000/42501)", code)
			}
			// A plain member's decision matches no row (the decide policy's manager predicate).
			if n, code := decide(e, enrMemberL, rid, "approved"); n != 0 || code != "" {
				t.Errorf("member decision: %d %s, want 0 rows", n, code)
			}
			if n, code := decide(e, enrOwnerL, rid, "approved"); n != 1 || code != "" {
				t.Fatalf("approve: %d %s", n, code)
			}
			if _, code := decide(e, enrOwnerL, rid, "denied"); code != "42501" {
				t.Errorf("approved -> denied: %q, want 42501", code)
			}
			e.as(rtr, enrOwnerL, enrOrgL)
			if _, code := e.try(`UPDATE app.enrollment_requests SET status = 'pending', decided_at = NULL, decided_by_user_id = NULL WHERE id = $1`, rid); code != "42501" {
				t.Errorf("approved -> pending: %q, want 42501", code)
			}
			if out := admit(e, rid); !strings.HasPrefix(out, "admitted|") {
				t.Fatalf("admit: %q", out)
			}
			if _, code := decide(e, enrOwnerL, rid, "revoked"); code != "42501" {
				t.Errorf("revoke a consumed approval: %q, want 42501", code)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s|%s', status, consumed_at IS NOT NULL, (SELECT count(*) FROM app.enrollment_restrictions WHERE user_id = $2))
FROM app.enrollment_requests WHERE id = $1`, rid, ua); got != "approved|t|0" {
				t.Errorf("after the refused transitions: %s", got)
			}
		})
		// Mandatory atomic append: a refused request capture leaves no request; a refused decision capture leaves the request
		// pending and writes no ledger row.
		run("atomic append", func(e env) {
			e.as("", "", "")
			if _, err := e.tx.Exec(ctx, `DROP POLICY events_insert_enrollment ON audit.events`); err != nil {
				t.Fatal(err)
			}
			e.as(rtr, ua, "")
			if _, code := e.try(`SELECT * FROM app.enrollment_request($1, $2::uuid)`, cred, enrOrgL); code != "42501" {
				t.Errorf("request without the capture policy: %q, want 42501", code)
			}
			e.as("", "", "")
			if got := e.val(`SELECT count(*)::text FROM app.enrollment_requests WHERE user_id = $1`, ua); got != "0" {
				t.Errorf("requests after the refused capture: %s", got)
			}
		})
		run("atomic decision", func(e env) {
			rid := ridOf(t, request(e))
			e.as("", "", "")
			if _, err := e.tx.Exec(ctx, `ALTER TABLE audit.events ADD CONSTRAINT ct_enr_fault CHECK (action <> 'enrollment.decided') NOT VALID`); err != nil {
				t.Fatal(err)
			}
			if _, code := decide(e, enrOwnerL, rid, "denied"); code != "23514" {
				t.Errorf("denial under a refused capture: %q, want 23514", code)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s', status, (SELECT count(*) FROM app.enrollment_restrictions WHERE user_id = $2))
FROM app.enrollment_requests WHERE id = $1`, rid, ua); got != "pending|0" {
				t.Errorf("after the refused decision capture: %s", got)
			}
		})
		// ENR-22f (atomic rollback, r3 correction 1): with enrollment_restrictions_owner_insert dropped inside a savepoint,
		// each runtime path that must record a restriction fails in its own statement — a manager's RemoveMember delete of
		// a member binding, the member's self-leave of its last (only) Organization binding, a manager's denial — and,
		// after the savepoint rolls back, the binding is present, the request pending and the ledger has no new row. The
		// same three statements with the policy restored commit and each writes exactly one ledger row (positive controls:
		// the refusals are the missing owner policy's, not the runtime's delete/decide policies).
		run("atomic without the owner policy", func(e env) {
			rid := ridOf(t, request(e))
			state := func() string {
				e.as("", "", "")
				return e.val(`SELECT format('%s|%s|%s', (SELECT count(*) FROM app.role_bindings WHERE organization_id = $1 AND subject_user_id = $2),
  (SELECT status FROM app.enrollment_requests WHERE id = $3::uuid), (SELECT count(*) FROM app.enrollment_restrictions))`, enrOrgL, enrMemberL, rid)
			}
			before := state()
			if before != "1|pending|0" {
				t.Fatalf("before: %s", before)
			}
			if got := e.val(`SELECT count(*)::text FROM app.role_bindings WHERE subject_user_id = $1`, enrMemberL); got != "1" {
				t.Fatalf("member bindings: %s, want exactly the Organization binding of L", got)
			}
			removeMember := `DELETE FROM app.role_bindings WHERE organization_id = $1 AND subject_user_id = $2 AND scope_type = 'organization'`
			cases := []struct {
				name string
				run  func() (int64, string)
			}{
				{"manager RemoveMember", func() (int64, string) { e.as(rtr, enrOwnerL, enrOrgL); return e.try(removeMember, enrOrgL, enrMemberL) }},
				{"self-leave of the last Organization binding", func() (int64, string) {
					e.as(rtr, enrMemberL, enrOrgL)
					return e.try(removeMember, enrOrgL, enrMemberL)
				}},
				{"manager denial", func() (int64, string) { return decide(e, enrOwnerL, rid, "denied") }},
			}
			sp, err := e.tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			e.as("", "", "")
			if _, err := e.tx.Exec(ctx, `DROP POLICY enrollment_restrictions_owner_insert ON app.enrollment_restrictions`); err != nil {
				t.Fatal(err)
			}
			for _, c := range cases {
				if n, code := c.run(); code != "42501" {
					t.Errorf("%s without the owner policy: %d rows %q, want 42501 (the statement fails)", c.name, n, code)
				}
			}
			if err := sp.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			if got := state(); got != before {
				t.Errorf("after the rolled-back savepoint: %s, want %s", got, before)
			}
			for _, c := range cases {
				func() {
					sp, err := e.tx.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					defer func() { _ = sp.Rollback(ctx) }()
					if n, code := c.run(); n != 1 || code != "" {
						t.Errorf("%s with the owner policy: %d rows %q, want 1 row", c.name, n, code)
					}
					e.as("", "", "")
					if got := e.val(`SELECT count(*)::text FROM app.enrollment_restrictions WHERE organization_id = $1 AND recorded_via = 'trigger'`,
						enrOrgL); got != "1" {
						t.Errorf("%s with the owner policy: %s ledger rows, want 1", c.name, got)
					}
				}()
			}
			if got := state(); got != before {
				t.Errorf("after the positive controls: %s, want %s", got, before)
			}
		})
		// Nothing above was committed (superuser read: FORCE RLS would hide rows from the owner).
		var left string
		if err := setup.QueryRow(ctx, `SELECT format('%s|%s', (SELECT count(*) FROM app.enrollment_requests), (SELECT count(*) FROM audit.events
  WHERE action IN ('enrollment.requested', 'enrollment.decided', 'enrollment.revoked')))`).Scan(&left); err != nil || left != "0|0" {
			t.Errorf("request rows/events after the rolled-back cases: %s %v", left, err)
		}
	})

	// ENR-16 (Plan r3 §1.4/§6.2 (a)-(c); addendum §A.6 (e)/(f); correction 01 §C.3/§C.6 (g)/(h)): the bound
	// replay-restrict.sql run as the superuser, as recovery.sh runs it, on restored states emulated in one database. Each
	// case is one transaction of the superuser fixture connection that is always rolled back; inside a case, a savepoint
	// rolled back after live operator steps stands for restoring a checkpoint taken before them (stored values come back,
	// as pg_restore restores them). (d) is CT-ENR-1_replay_pre16 above; (i) is CT-ENR-1_readiness (22004, 22023,
	// not_legacy). The restore, invalidation and replay gates of recovery.sh itself are the unchanged G013 drill's.
	t.Run("CT-ENR-1_replay", func(t *testing.T) {
		const (
			ux = "0f58e000-0000-4000-8000-000000000031" // the departed subject
			uv = "0f58e000-0000-4000-8000-000000000032" // an unrestricted applicant (positive control)
			ug = "0f58e000-0000-4000-8000-0000000000fd" // a subject absent from the checkpoint (no users row)
			zz = "0f58e000-0000-4000-8000-0000000000fc" // an Organization or Group absent from the checkpoint
		)
		credX, credV := bytes.Repeat([]byte{0xc1}, 32), bytes.Repeat([]byte{0xc2}, 32)
		type env struct {
			t   *testing.T
			tx  pgx.Tx
			as  func(role, user, org string)
			try func(sql string, args ...any) (int64, string)
			val func(sql string, args ...any) string
			// replay runs the bound script for one ledger entry as the superuser; its printed lines joined by ';'.
			replay func(vars map[string]string) string
		}
		rcase := func(name string, f func(e env)) {
			t.Run(name, func(t *testing.T) {
				tx, err := setup.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				e := env{t: t, tx: tx}
				e.as = func(role, user, org string) {
					t.Helper()
					if _, err := tx.Exec(ctx, `RESET ROLE`); err != nil {
						t.Fatal(err)
					}
					if role != "" {
						if _, err := tx.Exec(ctx, `SET LOCAL ROLE `+quote(role)); err != nil {
							t.Fatal(err)
						}
					}
					rlsContext(t, tx, user, org, "", "")
				}
				e.try = func(sql string, args ...any) (int64, string) {
					t.Helper()
					sp, err := tx.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					tag, err := sp.Exec(ctx, sql, args...)
					if err != nil {
						_ = sp.Rollback(ctx)
						return 0, sqlstateOf(err)
					}
					if err := sp.Commit(ctx); err != nil {
						t.Fatal(err)
					}
					return tag.RowsAffected(), ""
				}
				e.val = func(sql string, args ...any) string {
					t.Helper()
					var s *string
					if err := tx.QueryRow(ctx, sql, args...).Scan(&s); err != nil {
						t.Fatalf("%s: %v", sql, err)
					}
					if s == nil {
						return ""
					}
					return *s
				}
				e.replay = func(vars map[string]string) string {
					t.Helper()
					e.as("", "", "")
					lines, err := psqlRun(ctx, tx, replayRestrictScript(t, vars))
					if err != nil {
						t.Fatalf("replay: %v (printed %q)", err, lines)
					}
					return strings.Join(lines, ";")
				}
				// The fixture: X and V, each with a live enrollment credential for L.
				e.as("", "", "")
				if _, err := tx.Exec(ctx, `INSERT INTO app.users (id, display_name) VALUES ($1, 'Enr Replay X'), ($2, 'Enr Replay V')`, ux, uv); err != nil {
					t.Fatalf("fixture: %v", err)
				}
				for _, c := range []struct {
					user string
					cred []byte
				}{{ux, credX}, {uv, credV}} {
					if _, err := tx.Exec(ctx, `INSERT INTO app.enrollment_sessions (id, user_id, organization_id, origin, credential_hash, csrf_token,
  provider_session_id, expires_at) VALUES (gen_random_uuid(), $1, $2, 'https://l.example.test', $3, $4, $5, now() + interval '10 minutes')`,
						c.user, enrOrgL, c.cred, bytes.Repeat([]byte{0xd3}, 32), "ct-enr-replay-"+c.user); err != nil {
						t.Fatalf("fixture session: %v", err)
					}
				}
				f(e)
			})
		}
		remove := func(org, subject, scopeType, scopeID, role string) map[string]string {
			return replayVars("membership.remove", org, subject, scopeType, scopeID, role)
		}
		// rows lists the subject's ledger rows in L: kind|recorded_via|no creator|unlifted (superuser read).
		rows := func(e env, user string) string {
			e.as("", "", "")
			return e.val(`SELECT coalesce(string_agg(format('%s|%s|%s|%s', kind, recorded_via, created_by_user_id IS NULL, lifted_at IS NULL), ','
  ORDER BY seq), '') FROM app.enrollment_restrictions WHERE organization_id = $1 AND user_id = $2`, enrOrgL, user)
		}
		// readiness is L's mode|basis|resolved_at IS NULL.
		readiness := func(e env) string {
			e.as("", "", "")
			return e.val(`SELECT format('%s|%s|%s', enrollment_mode, coalesce(enrollment_history_basis, '-'), enrollment_history_resolved_at IS NULL)
FROM app.organizations WHERE id = $1`, enrOrgL)
		}
		// events is the multiset of L's enrollment audit rows (action|reason|actor kind), sorted.
		events := func(e env) string {
			e.as("", "", "")
			return e.val(`SELECT coalesce(string_agg(x, ',' ORDER BY x COLLATE "C"), '') FROM (SELECT format('%s|%s|%s', action, coalesce(reason, '-'),
  actor_kind) AS x FROM audit.events WHERE organization_id = $1 AND action LIKE 'enrollment.%') s`, enrOrgL)
		}
		admitOpen := func(e env, user string, cred []byte) string {
			e.as(rtr, user, "")
			return e.val(`SELECT outcome || '|' || coalesce(binding_id::text, '-') FROM app.enrollment_admit($1, $2::uuid, 'open', NULL)`, cred, enrOrgL)
		}
		set := `SELECT app.enrollment_readiness_set($1::uuid, $2, $3)`
		backfill := `SELECT app.enrollment_history_backfill($1::uuid, $2::uuid, $3)`
		// readyOpen: L attested (operator_verified_absent) and opened, the state a checkpoint captures in (g) and (h).
		readyOpen := func(e env) {
			e.t.Helper()
			e.as("", "", "")
			if got := e.val(set, enrOrgL, "operator_verified_absent", "ct-enr-replay"); got != "resolved" {
				e.t.Fatalf("attest L: %q", got)
			}
			if n, code := e.try(`UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = $1`, enrOrgL); n != 1 || code != "" {
				e.t.Fatalf("open L: %d %s", n, code)
			}
		}

		// (a) The checkpoint holds the binding: the G013 delete fires the departure trigger; the ledger line reports the
		// trigger's row; a second replay is already applied; readiness untouched.
		rcase("ENR-16a_binding_in_checkpoint", func(e env) {
			t := e.t
			v := remove(enrOrgL, enrMemberL, "organization", enrOrgL, "org_member")
			if got := e.replay(v); got != "updated\t1;state\tnone;ledger\tpresent" {
				t.Errorf("replay: %q", got)
			}
			if got := rows(e, enrMemberL); got != "departed|trigger|t|t" {
				t.Errorf("ledger rows: %q, want one trigger row", got)
			}
			if got := e.replay(v); got != "updated\t0;state\tnone;ledger\tpresent" {
				t.Errorf("second replay: %q", got)
			}
			if got := rows(e, enrMemberL); got != "departed|trigger|t|t" {
				t.Errorf("ledger rows after the second replay: %q", got)
			}
			if got := readiness(e); got != "invite_only|-|t" {
				t.Errorf("readiness: %s", got)
			}
		})
		// (b) The checkpoint was taken after the removal: it holds the restriction row, so the replay is already applied.
		rcase("ENR-16b_restriction_in_checkpoint", func(e env) {
			t := e.t
			e.as("", "", "")
			for _, s := range []string{
				`INSERT INTO app.role_bindings (organization_id, subject_user_id, scope_type, scope_id, group_id, role, source) VALUES
  ('` + enrOrgL + `', '` + ux + `', 'organization', '` + enrOrgL + `', NULL, 'org_member', 'membership')`,
				`DELETE FROM app.role_bindings WHERE organization_id = '` + enrOrgL + `' AND subject_user_id = '` + ux + `'`,
			} {
				if n, code := e.try(s); n != 1 || code != "" {
					t.Fatalf("fixture %q: %d %s", s, n, code)
				}
			}
			if got := e.replay(remove(enrOrgL, ux, "organization", enrOrgL, "org_member")); got != "updated\t0;state\tnone;ledger\tpresent" {
				t.Errorf("replay: %q", got)
			}
			if got := rows(e, ux); got != "departed|trigger|t|t" {
				t.Errorf("ledger rows: %q, want the restored trigger row only", got)
			}
		})
		// (c) The checkpoint predates both the grant and the removal: X exists with an approval decided before the replay
		// (authority version 0) and no binding or restriction. The replay materializes one departed row newer than that
		// approval, so admission under it is refused with zero writes (a restored invitation is refused by the same
		// created_at rule, ENR-23); a second replay finds it present. Group scope, an Organization absent from the
		// checkpoint and another kind materialize nothing.
		rcase("ENR-16c_checkpoint_predates_grant_and_removal", func(e env) {
			t := e.t
			e.as("", "", "")
			if n, code := e.try(`UPDATE app.organizations SET enrollment_mode = 'request_access' WHERE id = $1`, enrOrgL); n != 1 || code != "" {
				t.Fatalf("request_access: %d %s", n, code)
			}
			e.as(rtr, ux, "")
			rid := e.val(`SELECT request_id::text FROM app.enrollment_request($1, $2::uuid) WHERE outcome = 'pending'`, credX, enrOrgL)
			if rid == "" {
				t.Fatal("no pending request")
			}
			e.as(rtr, enrOwnerL, enrOrgL)
			if n, code := e.try(`UPDATE app.enrollment_requests SET status = 'approved', decided_at = now(), decided_by_user_id = $2 WHERE id = $1`,
				rid, enrOwnerL); n != 1 || code != "" {
				t.Fatalf("approve: %d %s", n, code)
			}
			v := remove(enrOrgL, ux, "organization", enrOrgL, "org_member")
			if got := e.replay(v); got != "updated\t0;state\tnone;ledger\tmaterialized" {
				t.Errorf("replay: %q", got)
			}
			if got := rows(e, ux); got != "departed|replay|t|t" {
				t.Errorf("ledger rows: %q, want one replay row without a creator", got)
			}
			e.as("", "", "")
			if got := e.val(`SELECT (r.created_at > q.decided_at AND r.created_at > q.created_at AND r.seq > q.authority_seq)::text
FROM app.enrollment_restrictions r, app.enrollment_requests q WHERE r.organization_id = $1 AND r.user_id = $2 AND q.id = $3::uuid`, enrOrgL, ux, rid); got != "true" {
				t.Errorf("the materialized row is not newer than the restored approval: %s", got)
			}
			if got := e.val(`SELECT count(*)::text FROM audit.events WHERE action = 'enrollment.restriction_backfilled'`); got != "0" {
				t.Errorf("backfill events from a replay: %s", got)
			}
			e.as(rtr, ux, "")
			if got := e.val(`SELECT outcome || '|' || coalesce(binding_id::text, '-') FROM app.enrollment_admit($1, $2::uuid, 'request', $3::uuid)`,
				credX, enrOrgL, rid); got != "enrollment_restricted|-" {
				t.Errorf("admission under the materialized restriction: %q", got)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s|%s', (SELECT count(*) FROM app.role_bindings WHERE subject_user_id = $1),
  (SELECT consumed_at IS NULL FROM app.enrollment_requests WHERE id = $2::uuid),
  (SELECT bool_and(lifted_at IS NULL) FROM app.enrollment_restrictions WHERE user_id = $1))`, ux, rid); got != "0|t|t" {
				t.Errorf("writes after the refusal: %s, want none", got)
			}
			if got := e.replay(v); got != "updated\t0;state\tnone;ledger\tpresent" {
				t.Errorf("second replay: %q", got)
			}
			if got := e.replay(remove(enrOrgL, ux, "group", zz, "group_member")); got != "updated\t0;state\tabsent;ledger\tunavailable" {
				t.Errorf("Group-scope replay: %q", got)
			}
			if got := e.replay(remove(zz, ux, "organization", zz, "org_member")); got != "updated\t0;state\tabsent;ledger\tunavailable" {
				t.Errorf("replay for an absent Organization: %q", got)
			}
			if got := e.replay(replayVars("policy.disable", enrOrgL, "-", "-", "-", "-")); !strings.HasSuffix(got, ";ledger\tunavailable") {
				t.Errorf("policy.disable replay: %q", got)
			}
			if got := rows(e, ux); got != "departed|replay|t|t" {
				t.Errorf("ledger rows after the other entries: %q", got)
			}
			if got := readiness(e); got != "request_access|-|t" {
				t.Errorf("readiness: %s", got)
			}
		})
		// (e) A checkpoint taken after an attestation and a backfill restores both; the backfill's entry replays present.
		rcase("ENR-16e_attested_and_backfilled_before_checkpoint", func(e env) {
			t := e.t
			e.as("", "", "")
			if got := e.val(backfill, enrOrgL, ux, "l5.ct-enr-replay"); got != "inserted" {
				t.Fatalf("backfill: %q", got)
			}
			if got := e.val(set, enrOrgL, "operator_backfilled", "l5.ct-enr-replay"); got != "resolved" {
				t.Fatalf("attest: %q", got)
			}
			before := events(e)
			if got := e.replay(remove(enrOrgL, ux, "organization", enrOrgL, "org_member")); got != "updated\t0;state\tnone;ledger\tpresent" {
				t.Errorf("replay: %q", got)
			}
			if got := rows(e, ux); got != "departed|backfill|t|t" {
				t.Errorf("ledger rows: %q", got)
			}
			if got := readiness(e); got != "invite_only|operator_backfilled|f" {
				t.Errorf("readiness: %s", got)
			}
			if got := events(e); got != before {
				t.Errorf("events changed by the replay: %s, was %s", got, before)
			}
		})
		// (f) A checkpoint taken before them: the restored L is closed without the row; the operator's ledger entry
		// materializes it; open stays refused (23514) until the procedure is repeated, and the explicit re-attestation then
		// accepts the replay-only history without lifting the restriction.
		rcase("ENR-16f_checkpoint_before_attestation_and_backfill", func(e env) {
			t := e.t
			e.as("", "", "")
			sp, err := e.tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if got := e.val(backfill, enrOrgL, ux, "l6.ct-enr-replay"); got != "inserted" {
				t.Fatalf("backfill: %q", got)
			}
			if got := e.val(set, enrOrgL, "operator_backfilled", "l6.ct-enr-replay"); got != "resolved" {
				t.Fatalf("attest: %q", got)
			}
			if err := sp.Rollback(ctx); err != nil { // restore the earlier checkpoint
				t.Fatal(err)
			}
			if got, r := readiness(e), rows(e, ux); got != "invite_only|-|t" || r != "" {
				t.Fatalf("restored state: %s rows %q", got, r)
			}
			if got := e.replay(remove(enrOrgL, ux, "organization", enrOrgL, "org_member")); got != "updated\t0;state\tnone;ledger\tmaterialized" {
				t.Errorf("replay: %q", got)
			}
			if got := rows(e, ux); got != "departed|replay|t|t" {
				t.Errorf("ledger rows: %q", got)
			}
			if got := readiness(e); got != "invite_only|-|t" {
				t.Errorf("readiness after the replay: %s", got)
			}
			if _, code := e.try(`UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = $1`, enrOrgL); code != "23514" {
				t.Errorf("open before the procedure is repeated: %q, want 23514", code)
			}
			// Re-attestation (correction 01 §C.4; Codex re-attestation correction): the repeated backfill finds the replay row
			// present and still leaves L closed (no automatic readiness); the explicit, audited operator_backfilled attestation
			// accepts the replay-only history. The row keeps its replay provenance, seq and times; after an authorized open
			// the departed X is still refused (zero writes) while V is admitted.
			e.as("", "", "")
			row := `SELECT format('%s|%s|%s|%s', seq, recorded_via, created_at, lifted_at IS NULL) FROM app.enrollment_restrictions
WHERE organization_id = $1 AND user_id = $2`
			replayed, before := e.val(row, enrOrgL, ux), events(e)
			beforeN := e.val(`SELECT count(*)::text FROM audit.events WHERE organization_id = $1 AND action LIKE 'enrollment.%'`, enrOrgL)
			if got := e.val(backfill, enrOrgL, ux, "l6.ct-enr-replay"); got != "present" {
				t.Errorf("repeated backfill: %q, want present", got)
			}
			if got := readiness(e); got != "invite_only|-|t" {
				t.Errorf("readiness before the explicit re-attestation: %s", got)
			}
			if got := events(e); got != before {
				t.Errorf("events before the re-attestation: %s, was %s", got, before)
			}
			if got := e.val(set, enrOrgL, "operator_backfilled", "l6.ct-enr-replay"); got != "resolved" {
				t.Fatalf("re-attest on replay-only history: %q, want resolved", got)
			}
			if got := readiness(e); got != "invite_only|operator_backfilled|f" {
				t.Errorf("readiness after the re-attestation: %s", got)
			}
			if got := e.val(`SELECT format('%s|%s', count(*) - $2::text::bigint, count(*) FILTER (WHERE action = 'enrollment.readiness_changed'
  AND reason = 'operator_backfilled' AND actor_kind = 'unknown' AND request_id = 'l6.ct-enr-replay'))
FROM audit.events WHERE organization_id = $1 AND action LIKE 'enrollment.%'`, enrOrgL, beforeN); got != "1|1" {
				t.Errorf("re-attestation audit (want exactly one new readiness row): %s, was %q", got, before)
			}
			if got := e.val(row, enrOrgL, ux); got != replayed {
				t.Errorf("replayed restriction rewritten by the re-attestation: %s, was %s", got, replayed)
			}
			e.as(rtr, enrOwnerL, enrOrgL)
			if n, code := e.try(`UPDATE app.organizations SET enrollment_mode = 'open' WHERE id = $1`, enrOrgL); n != 1 || code != "" {
				t.Fatalf("authorized open after the re-attestation: %d %s", n, code)
			}
			if got := admitOpen(e, ux, credX); got != "enrollment_restricted|-" {
				t.Errorf("X's open admission after the re-attestation: %q", got)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s', (SELECT count(*) FROM app.role_bindings WHERE subject_user_id = $1),
  (SELECT bool_and(admitted_at IS NULL) FROM app.enrollment_sessions WHERE user_id = $1))`, ux); got != "0|t" {
				t.Errorf("writes after X's refusal: %s", got)
			}
			if got := admitOpen(e, uv, credV); !strings.HasPrefix(got, "admitted|") {
				t.Errorf("V's open admission (control): %q", got)
			}
			e.as("", "", "")
			if got := e.val(row, enrOrgL, ux); got != replayed {
				t.Errorf("replayed restriction after the open: %s, was %s", got, replayed)
			}
		})
		// (g) The reviewed gap: a checkpoint of a ready and open L; afterwards the mandatory ledger entry (seq 9) and its
		// backfill close L on the live database. Restored, L is ready and open again without the row; the replay of the
		// entry materializes it, so X's open admission is refused (V's still admitted) though the restored ready fact and
		// open survive; a second replay finds it present.
		rcase("ENR-16g_ready_checkpoint_later_backfill", func(e env) {
			t := e.t
			readyOpen(e)
			sp, err := e.tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			e.as("", "", "")
			if got := e.val(backfill, enrOrgL, ux, "l9.ct-enr-replay"); got != "inserted" {
				t.Fatalf("backfill: %q", got)
			}
			if got := readiness(e); got != "invite_only|-|t" {
				t.Errorf("live-side close: %s", got)
			}
			if got := events(e); got != "enrollment.policy_changed|-|unknown,enrollment.policy_changed|-|unknown,"+
				"enrollment.readiness_changed|history_unresolved|unknown,enrollment.readiness_changed|operator_verified_absent|unknown,"+
				"enrollment.restriction_backfilled|operator_backfilled|unknown" {
				t.Errorf("live backfill events: %s", got)
			}
			if err := sp.Rollback(ctx); err != nil { // restore the checkpoint
				t.Fatal(err)
			}
			if got, r := readiness(e), rows(e, ux); got != "open|operator_verified_absent|f" || r != "" {
				t.Fatalf("restored state: %s rows %q", got, r)
			}
			v := remove(enrOrgL, ux, "organization", enrOrgL, "org_member")
			if got := e.replay(v); got != "updated\t0;state\tnone;ledger\tmaterialized" {
				t.Errorf("replay: %q", got)
			}
			if got := rows(e, ux); got != "departed|replay|t|t" {
				t.Errorf("ledger rows: %q", got)
			}
			if got := readiness(e); got != "open|operator_verified_absent|f" {
				t.Errorf("readiness after a materializing replay: %s", got)
			}
			if got := admitOpen(e, ux, credX); got != "enrollment_restricted|-" {
				t.Errorf("X's open admission: %q", got)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s', (SELECT count(*) FROM app.role_bindings WHERE subject_user_id = $1),
  (SELECT bool_and(admitted_at IS NULL) FROM app.enrollment_sessions WHERE user_id = $1))`, ux); got != "0|t" {
				t.Errorf("writes after X's refusal: %s", got)
			}
			if got := admitOpen(e, uv, credV); !strings.HasPrefix(got, "admitted|") {
				t.Errorf("V's open admission (control): %q", got)
			}
			if got := e.replay(v); got != "updated\t0;state\tnone;ledger\tpresent" {
				t.Errorf("second replay: %q", got)
			}
			if got := rows(e, ux); got != "departed|replay|t|t" {
				t.Errorf("ledger rows after the second replay: %q", got)
			}
		})
		// (h) The same ready and open checkpoint, an entry that cannot be materialized: its subject is absent (state absent,
		// still a conflict for recovery.sh) or its binding remains (role mismatch). Each closes L (NULL readiness,
		// invite_only, a policy and a readiness row, operator channel); a rerun is idempotent.
		rcase("ENR-16h_unmaterializable_entry_closes", func(e env) {
			t := e.t
			readyOpen(e)
			v := remove(enrOrgL, ug, "organization", enrOrgL, "org_member")
			if got := e.replay(v); got != "updated\t0;state\tabsent;ledger\tclosed" {
				t.Errorf("replay: %q", got)
			}
			if got := readiness(e); got != "invite_only|-|t" {
				t.Errorf("readiness after the close: %s", got)
			}
			closed := "enrollment.policy_changed|-|unknown,enrollment.policy_changed|-|unknown," +
				"enrollment.readiness_changed|history_unresolved|unknown,enrollment.readiness_changed|operator_verified_absent|unknown"
			if got := events(e); got != closed {
				t.Errorf("close events: %s", got)
			}
			if got := e.replay(v); got != "updated\t0;state\tabsent;ledger\tclosed" {
				t.Errorf("rerun: %q", got)
			}
			if got := events(e); got != closed {
				t.Errorf("events after the rerun: %s", got)
			}
			readyOpen(e)
			if got := e.replay(remove(enrOrgL, enrMemberL, "organization", enrOrgL, "org_admin")); got != "updated\t0;state\torg_member;ledger\tclosed" {
				t.Errorf("role-mismatch replay: %q", got)
			}
			if got := readiness(e); got != "invite_only|-|t" {
				t.Errorf("readiness after the role-mismatch close: %s", got)
			}
			e.as("", "", "")
			if got := e.val(`SELECT format('%s|%s', (SELECT count(*) FROM app.role_bindings WHERE organization_id = $1 AND subject_user_id = $2
  AND scope_type = 'organization'), (SELECT count(*) FROM app.enrollment_restrictions WHERE organization_id = $1))`, enrOrgL, enrMemberL); got != "1|0" {
				t.Errorf("binding/ledger after the closes: %s, want the binding kept and no row", got)
			}
		})
	})

	t.Run("CT-ENR-1_updown", func(t *testing.T) {
		down := func() (string, error) {
			var b bytes.Buffer
			err := r.Down(ctx, &b)
			return b.String(), err
		}
		out, err := down()
		if err != nil || out != "rolled_back\t00016_organization_enrollment.sql\ncurrent_version\t15\n" {
			t.Fatalf("down: %v %q", err, out)
		}
		metadataThroughVersion15Only(t, cfg, r)
		if got := state(r); got != s15 {
			t.Fatalf("catalog after down differs from version 15:\n got %s\nwant %s", got, s15)
		}
		objRunUp(t, r, "applied\t00016_organization_enrollment.sql\ncurrent_version\t16\n")
		if got := state(r); got != s16 {
			t.Fatalf("catalog after down/up differs from version 16:\n got %s\nwant %s", got, s16)
		}
		// Refused while a ledger row exists, transactionally: catalog, row, FORCE and version unchanged.
		exec(`INSERT INTO app.enrollment_restrictions (organization_id, user_id, kind, recorded_via) VALUES ($1, $2, 'departed', 'replay')`,
			enrOrgL, enrMemberL)
		if out, err := down(); err == nil || !strings.Contains(err.Error(),
			"refusing to roll back 00016: 0 enrollment sessions / 1 restrictions / 0 requests / 0 login transactions / 0 audit rows (history is never discarded)") {
			t.Fatalf("down with a ledger row: %v %q", err, out)
		}
		if got := state(r); got != s16 {
			t.Fatal("the refused down changed the catalog")
		}
		if got := one(`SELECT string_agg(relname || ':' || relforcerowsecurity::text, ',' ORDER BY relname) FROM pg_class
WHERE oid IN ('app.enrollment_restrictions'::regclass, 'app.enrollment_sessions'::regclass, 'app.login_transactions'::regclass,
  'audit.events'::regclass, 'app.enrollment_requests'::regclass)`); got !=
			"enrollment_requests:true,enrollment_restrictions:true,enrollment_sessions:true,events:true,login_transactions:true" {
			t.Fatalf("FORCE after the refused down: %s", got)
		}
		if out := run(t, "version", func(ctx context.Context, b *bytes.Buffer) error { return r.Version(ctx, b) }); out != "current_version\t16\n" {
			t.Fatalf("version after the refused down: %q", out)
		}
		// Request slice: a committed request and its audited row are history as well.
		tx, err := setup.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		rlsContext(t, tx, enrMemberL, "", "", "")
		if _, err := tx.Exec(ctx, `INSERT INTO app.enrollment_requests (id, organization_id, user_id, status, expires_at)
VALUES (gen_random_uuid(), $1, $2, 'pending', now() + interval '1 day')`, enrOrgL, enrMemberL); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		if out, err := down(); err == nil || !strings.Contains(err.Error(),
			"refusing to roll back 00016: 0 enrollment sessions / 1 restrictions / 1 requests / 0 login transactions / 1 audit rows (history is never discarded)") {
			t.Fatalf("down with a request row: %v %q", err, out)
		}
		if got := state(r); got != s16 {
			t.Fatal("the refused down (request row) changed the catalog")
		}
		reset(t, cfg)
	})
}

// replayVars are the psql variables recovery.sh passes replay-restrict.sql for one entry (unused ones are "-").
func replayVars(kind, org, subject, scopeType, scopeID, role string) map[string]string {
	return map[string]string{"kind": kind, "org": org, "subject": subject, "scope_type": scopeType, "scope_id": scopeID, "role": role,
		"from_role": "-", "to_role": "-", "grp": "-", "profile": "-"}
}

// replayRestrictSHA256 binds the replay script these cases run to the reviewed infra/recovery/sql/replay-restrict.sql.
const replayRestrictSHA256 = "8ef8f9d9589508f9bdc2dd5be74b248c98de7bad7ce16e7e411234e2aaeab665"

// replayRestrictScript is infra/recovery/sql/replay-restrict.sql as the harness passes it (ENROLLMENT_REPLAY_SQL_B64,
// the MCP_INVALIDATE_SQL_B64 precedent: the core check image carries no infra file), split as psql sends it with its
// BEGIN and COMMIT removed: each case runs it inside its own always-rolled-back transaction.
func replayRestrictScript(t *testing.T, vars map[string]string) []psqlStmt {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(os.Getenv("ENROLLMENT_REPLAY_SQL_B64"))
	if err != nil || len(raw) == 0 {
		t.Fatal("ENROLLMENT_REPLAY_SQL_B64 is required (infra/recovery/sql/replay-restrict.sql)")
	}
	if sum := sha256.Sum256(raw); hex.EncodeToString(sum[:]) != replayRestrictSHA256 {
		t.Fatalf("the replay script is not the bound infra/recovery/sql/replay-restrict.sql (sha256 %s)", hex.EncodeToString(sum[:]))
	}
	stmts := psqlStatements(t, string(raw), vars)
	if len(stmts) < 3 || stmts[0].sql != "BEGIN;" || stmts[len(stmts)-1].sql != "COMMIT;" {
		t.Fatal("the replay script is not one BEGIN ... COMMIT transaction")
	}
	return stmts[1 : len(stmts)-1]
}

// psqlStmt is one statement psql sends; discard marks one ended by "\g /dev/null" (its rows are not printed).
type psqlStmt struct {
	sql     string
	discard bool
}

// psqlStatements splits a psql script the way psql -X -q -tA -v ON_ERROR_STOP=1 runs it, for the constructs the G013
// replay scripts use: outside single-quoted literals (E-prefixed escapes included), quoted identifiers, dollar quotes and
// comments (dropped), ';' or a "\g /dev/null" meta-command ends a statement and :'name' becomes the escaped literal of
// vars[name]. Any other meta-command, an unknown variable or a value containing a backslash fails the test.
func psqlStatements(t *testing.T, src string, vars map[string]string) []psqlStmt {
	t.Helper()
	var out []psqlStmt
	var b strings.Builder
	flush := func(discard bool) {
		if s := strings.TrimSpace(b.String()); s != "" {
			out = append(out, psqlStmt{sql: s, discard: discard})
		}
		b.Reset()
	}
	ident := func(c byte) bool {
		return c == '_' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
	}
	dollar := regexp.MustCompile(`^\$([A-Za-z_][A-Za-z0-9_]*)?\$`)
	for i := 0; i < len(src); {
		c := src[i]
		switch {
		case strings.HasPrefix(src[i:], "--"):
			j := strings.IndexByte(src[i:], '\n')
			if j < 0 {
				j = len(src) - i
			}
			i += j
		case strings.HasPrefix(src[i:], "/*"):
			j := strings.Index(src[i+2:], "*/")
			if j < 0 {
				t.Fatal("psql script: unterminated comment")
			}
			i += j + 4
		case c == '\'' || c == '"':
			esc := c == '\'' && i > 0 && (src[i-1] == 'E' || src[i-1] == 'e') && (i < 2 || !ident(src[i-2]))
			j := i + 1
			for ; j < len(src); j++ {
				if esc && src[j] == '\\' {
					j++
					continue
				}
				if src[j] == c {
					if j+1 < len(src) && src[j+1] == c {
						j++
						continue
					}
					break
				}
			}
			if j >= len(src) {
				t.Fatal("psql script: unterminated quote")
			}
			b.WriteString(src[i : j+1])
			i = j + 1
		case c == '$' && dollar.MatchString(src[i:]):
			tag := dollar.FindString(src[i:])
			j := strings.Index(src[i+len(tag):], tag)
			if j < 0 {
				t.Fatal("psql script: unterminated dollar quote")
			}
			b.WriteString(src[i : i+len(tag)+j+len(tag)])
			i += len(tag) + j + len(tag)
		case strings.HasPrefix(src[i:], ":'"):
			j := strings.IndexByte(src[i+2:], '\'')
			if j < 0 {
				t.Fatal("psql script: unterminated variable reference")
			}
			name := src[i+2 : i+2+j]
			v, ok := vars[name]
			if !ok || strings.ContainsRune(v, '\\') {
				t.Fatalf("psql script: variable %q unset or not a plain value", name)
			}
			b.WriteString("'" + strings.ReplaceAll(v, "'", "''") + "'")
			i += j + 3
		case c == '\\':
			j := strings.IndexByte(src[i:], '\n')
			if j < 0 {
				j = len(src) - i
			}
			if strings.TrimSpace(src[i:i+j]) != `\g /dev/null` {
				t.Fatalf("psql script: unsupported meta-command %q", src[i:i+j])
			}
			flush(true)
			i += j
		case c == ';':
			b.WriteByte(';')
			flush(false)
			i++
		default:
			b.WriteByte(c)
			i++
		}
	}
	flush(false)
	return out
}

// psqlRun sends each statement to tx's connection with the simple protocol, as psql does, and returns the printed rows
// (-tA: columns joined by '|'), stopping at the first error (ON_ERROR_STOP). Each G013 replay runs in a session of its
// own, so the transaction-local operator settings an earlier step of the same case set are cleared first.
func psqlRun(ctx context.Context, tx pgx.Tx, stmts []psqlStmt) ([]string, error) {
	if _, err := tx.Exec(ctx, `SELECT set_config('audit.request_id', '', true), set_config('app.enrollment_readiness_setter', '', true)`); err != nil {
		return nil, err
	}
	var lines []string
	for _, s := range stmts {
		res, err := tx.Conn().PgConn().Exec(ctx, s.sql).ReadAll()
		if err != nil {
			return lines, err
		}
		if s.discard {
			continue
		}
		for _, r := range res {
			for _, row := range r.Rows {
				cols := make([]string, len(row))
				for k, v := range row {
					cols[k] = string(v)
				}
				lines = append(lines, strings.Join(cols, "|"))
			}
		}
	}
	return lines, nil
}
