package audit

import (
	"context"
	"errors"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// U-AUDIT (G011 Plan §9; G014 Plan r8 §5): the closed catalog is identical in Go, the latest migration defining each list
// (00008, widened by 00010 and again by 00011) and docs/foundation/AUDIT.md; Append
// refuses trigger-owned actions before sending anything; the statement texts are pinned.

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// sqlList returns the quoted literals of the IN (...) list following marker in sql.
func sqlList(t *testing.T, sql, marker string) []string {
	t.Helper()
	i := strings.Index(sql, marker)
	if i < 0 {
		t.Fatalf("migration: %q not found", marker)
	}
	rest := sql[i+len(marker):]
	loc := regexp.MustCompile(`\bIN\s*\(`).FindStringIndex(rest)
	if loc == nil {
		t.Fatalf("migration: no IN list after %q", marker)
	}
	j := loc[1]
	k, depth := j, 1
	for ; k < len(rest) && depth > 0; k++ {
		switch rest[k] {
		case '(':
			depth++
		case ')':
			depth--
		}
	}
	var out []string
	for _, m := range regexp.MustCompile(`'([a-z_.]+)'`).FindAllStringSubmatch(rest[j:k], -1) {
		out = append(out, m[1])
	}
	return out
}

func sorted[T ~string](in []T) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = string(v)
	}
	slices.Sort(out)
	return out
}

func sortedStrings(in []string) []string { out := slices.Clone(in); slices.Sort(out); return out }

func TestCatalogParity(t *testing.T) {
	mig := readFile(t, "../../migrations/00008_audit_events.sql")
	// 00010 (G014) replaces the action, reason, operation and resource-type constraints and the trigger-owned list of
	// audit.events_attribute with widened lists; its up section is read first. The outcome list is 00008's.
	// 00011 (G016) replaces them again (the four domain and slug trigger-owned actions, the domain reasons, operations and
	// resource type); its up section is read first, so the lists below are the version-11 ones.
	// 00013 (G017) replaces them again (organization.branding_changed, its four reasons and the five branding
	// operations); its up section is read first, so the lists below are the version-13 ones.
	mig10 := readFile(t, "../../migrations/00013_organization_branding.sql")
	// 00014 (G018) replaces only the operation constraint (the seven Group onboarding operations); its up section is read
	// first, so the operation list below is the version-14 one. Every other list keeps its version-13 source.
	// 00015 (G019) replaces the action, operation and resource-type constraints and the trigger-owned list (the three
	// invitation.* captures, the seventeen membership/ownership/invitation/retirement operations and the invitation
	// resource type); its up section is read first. The reasons keep their version-13 source.
	mig15 := readFile(t, "../../migrations/00015_membership_invitations.sql")
	// 00016 (G058 slice 1a, Codex F5 and the round-2 audit decision) replaces the reason and operation constraints (the
	// seven enrollment refusal reasons and the eleven enrollment operations); its up section is read first, so the reason and operation lists below are the
	// version-16 ones. Actions, resource types and the trigger-owned list keep their version-15 source.
	// Slice 1b replaces the action constraint and the trigger-owned list too (the three enrollment policy/readiness
	// captures). The request slice adds three more trigger-owned actions (enrollment.requested, enrollment.decided,
	// enrollment.revoked) and replaces the resource-type constraint (enrollment_request): from it every list but the
	// outcomes is read from 00016.
	mig16 := readFile(t, "../../migrations/00016_organization_enrollment.sql")
	// 00017 (G060) replaces the action, reason and operation constraints (access.override_used, the two override reasons
	// of organization.policy_changed, organization.set_group_override) and the attribution body; its up section is read
	// first, so the action, reason, operation and trigger-owned lists below are the version-17 ones. Resource types keep
	// their version-16 source; the outcomes their 00008 source.
	mig17 := readFile(t, "../../migrations/00017_group_authority.sql")
	// 00018 (G062) replaces the action, reason and operation constraints again (no new action; the nine signup reasons and
	// the eight signup operations) and the attribution body; its up section is read first, so the action, reason,
	// operation and trigger-owned lists below are the version-18 ones. Resource types keep their version-16 source.
	mig18 := readFile(t, "../../migrations/00018_platform_signup.sql")
	// 00019 (G063) replaces the action constraint (+ mcp.connection_renamed) and the attribution body (the action is
	// trigger-owned); its up section is read first, so the action and trigger-owned lists below are the version-19 ones.
	// Reasons and operations keep their version-18 source.
	mig19 := readFile(t, "../../migrations/00019_mcp_connection_names.sql")
	// 00020 (G020) replaces all four catalog constraints (+ the five label.* captures, the two label.updated reasons, the
	// seven label operations and the label resource type) and the attribution body (the five actions are trigger-owned);
	// its up section is read first, so the action, reason, operation, resource-type and trigger-owned lists below are the
	// version-20 ones. The outcomes keep their 00008 source.
	mig20 := readFile(t, "../../migrations/00020_labels.sql")
	all := append(slices.Clone(TriggerActions), CoreActions...)
	if len(TriggerActions) != 45 || len(CoreActions) != 7 || len(all) != 52 || len(ResourceTypes) != 12 {
		t.Fatalf("catalog: %d trigger + %d core actions, %d resource types, want 45 + 7 = 52 and 12", len(TriggerActions), len(CoreActions),
			len(ResourceTypes))
	}
	for name, c := range map[string]struct{ goList, migList []string }{
		"actions":        {sorted(all), sortedStrings(sqlList(t, mig20, "ADD CONSTRAINT events_action_valid"))},
		"reasons":        {sorted(Reasons), sortedStrings(sqlList(t, mig20, "ADD CONSTRAINT events_reason_valid"))},
		"operations":     {sorted(Operations), sortedStrings(sqlList(t, mig20, "ADD CONSTRAINT events_operation_valid"))},
		"outcomes":       {sorted(Outcomes), sortedStrings(sqlList(t, mig, "CONSTRAINT events_outcome_valid"))},
		"resource types": {sorted(ResourceTypes), sortedStrings(sqlList(t, mig20, "ADD CONSTRAINT events_resource_type_valid"))},
		"trigger-owned":  {sorted(TriggerActions), sortedStrings(sqlList(t, mig20, "trig := NEW.action"))},
	} {
		if !slices.Equal(c.goList, c.migList) || len(slices.Compact(slices.Clone(c.goList))) != len(c.goList) {
			t.Errorf("%s differ:\n go        %v\n migration %v", name, c.goList, c.migList)
		}
	}
	// The 00020 lists are the reviewed version-19 sources (actions and trigger-owned 00019, reasons and operations 00018,
	// resource types 00016) plus exactly the label values; no value is dropped or reordered away.
	labelActions := []string{string(LabelCreated), string(LabelUpdated), string(LabelDeleted), string(LabelAssigned), string(LabelUnassigned)}
	for _, l := range []struct {
		marker string
		prev   string
		plus   []string
	}{
		{"ADD CONSTRAINT events_action_valid", mig19, labelActions},
		{"trig := NEW.action", mig19, labelActions},
		{"ADD CONSTRAINT events_reason_valid", mig18, []string{string(ReasonLabelRenamed), string(ReasonLabelVisibilityChanged)}},
		{"ADD CONSTRAINT events_operation_valid", mig18, []string{string(OpLabelDefinitionList), string(OpLabelDefinitionCreate),
			string(OpLabelDefinitionUpdate), string(OpLabelDefinitionDelete), string(OpLabelAssignmentList), string(OpLabelAssign),
			string(OpLabelUnassign)}},
		{"ADD CONSTRAINT events_resource_type_valid", mig16, []string{string(ResourceLabel)}},
	} {
		prev, v20 := sortedStrings(sqlList(t, l.prev, l.marker)), sortedStrings(sqlList(t, mig20, l.marker))
		if want := sortedStrings(append(slices.Clone(prev), l.plus...)); !slices.Equal(v20, want) {
			t.Errorf("%s: 00020 list %v, want the version-19 list plus exactly %v", l.marker, v20, l.plus)
		}
	}
	// The 00018 action and trigger-owned lists stay the reviewed version-18 sources the 00019 lists extend by exactly
	// mcp.connection_renamed.
	for _, marker := range []string{"ADD CONSTRAINT events_action_valid", "trig := NEW.action"} {
		v18, v19 := sortedStrings(sqlList(t, mig18, marker)), sortedStrings(sqlList(t, mig19, marker))
		if want := sortedStrings(append(slices.Clone(v18), string(MCPConnectionRenamed))); !slices.Equal(v19, want) {
			t.Errorf("%s: 00019 list %v, want the 00018 list plus mcp.connection_renamed", marker, v19)
		}
	}
	// Codex round-2 audit decision: +3 reasons (recipient_proof_unavailable, recipient_ineligible, session_invalid);
	// round-3 decision 1: +1 operation (enrollment.start, the callback T1 policy/intent denial).
	// Slice 1b: +3 reasons (operator_verified_absent, operator_backfilled, history_unresolved).
	// G060: +2 reasons (group_management_override_enabled/disabled), +1 operation (organization.set_group_override).
	// G062: +9 reasons (signup_authenticated, tenant_created, signup_origin_rejected, signup_user_mismatch, signup_invalid,
	// intent_used, slug_taken, continuation_invalid, signup_context_mismatch), +8 signup operations.
	// G020: +2 reasons (renamed, visibility_changed), +7 label operations.
	if len(Reasons) != 81 || len(Operations) != 80 {
		t.Errorf("%d reasons, %d operations: want 81 and 80", len(Reasons), len(Operations))
	}
	// The 00017 lists stay the reviewed version-17 sources the 00018 lists extend (no value dropped).
	for _, l := range []struct {
		marker string
		ok     func(string) bool
	}{
		{"ADD CONSTRAINT events_action_valid", func(v string) bool { return slices.Contains(all, Action(v)) }},
		{"ADD CONSTRAINT events_reason_valid", func(v string) bool { return slices.Contains(Reasons, Reason(v)) }},
		{"ADD CONSTRAINT events_operation_valid", func(v string) bool { return slices.Contains(Operations, Operation(v)) }},
		{"trig := NEW.action", func(v string) bool { return slices.Contains(TriggerActions, Action(v)) }},
	} {
		for _, v := range sqlList(t, mig17, l.marker) {
			if !l.ok(v) {
				t.Errorf("%s value %s of 00017 missing at version 18", l.marker, v)
			}
		}
	}
	// The 00016 lists stay the reviewed version-16 sources the 00017 lists extend (no value dropped).
	for _, l := range []struct {
		marker string
		ok     func(string) bool
	}{
		{"ADD CONSTRAINT events_action_valid", func(v string) bool { return slices.Contains(all, Action(v)) }},
		{"ADD CONSTRAINT events_reason_valid", func(v string) bool { return slices.Contains(Reasons, Reason(v)) }},
		{"ADD CONSTRAINT events_operation_valid", func(v string) bool { return slices.Contains(Operations, Operation(v)) }},
	} {
		for _, v := range sqlList(t, mig16, l.marker) {
			if !l.ok(v) {
				t.Errorf("%s value %s of 00016 missing at version 17", l.marker, v)
			}
		}
	}
	for _, a := range sqlList(t, mig15, "ADD CONSTRAINT events_action_valid") {
		if !slices.Contains(all, Action(a)) {
			t.Errorf("action %s of 00015 missing at version 16", a)
		}
	}
	// The 00013 and 00015 lists stay the reviewed version-13/15 sources the 00016 lists extend (no value dropped).
	for _, r := range sqlList(t, mig10, "ADD CONSTRAINT events_reason_valid") {
		if !slices.Contains(Reasons, Reason(r)) {
			t.Errorf("reason %s of 00013 missing at version 16", r)
		}
	}
	for _, o := range sqlList(t, mig15, "ADD CONSTRAINT events_operation_valid") {
		if !slices.Contains(Operations, Operation(o)) {
			t.Errorf("operation %s of 00015 missing at version 16", o)
		}
	}
	for _, rt := range sqlList(t, mig15, "ADD CONSTRAINT events_resource_type_valid") {
		if !slices.Contains(ResourceTypes, ResourceType(rt)) {
			t.Errorf("resource type %s of 00015 missing at version 16", rt)
		}
	}
	// The AUDIT.md half of the parity (its catalog table against the migration) runs in B-AUDIT-CATALOG: the core
	// check image holds the core module only, never docs/.
}

// tripTx fails the test when any statement is sent.
type tripTx struct {
	pgx.Tx
	t *testing.T
}

func (x tripTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	x.t.Error("a statement was sent")
	return pgconn.CommandTag{}, nil
}

func TestAppendRefusesTriggerActions(t *testing.T) {
	for _, a := range TriggerActions {
		if err := Append(context.Background(), tripTx{t: t}, Event{Action: a, Outcome: Success}); !errors.Is(err, ErrRefused) {
			t.Errorf("%s: %v, want ErrRefused before any statement", a, err)
		}
	}
	for _, c := range []struct {
		err  error
		want error
	}{
		{&pgconn.PgError{Code: "42501", Message: "secret text"}, ErrRefused},
		{&pgconn.PgError{Code: "23514", Message: "secret text"}, ErrRefused},
		{&pgconn.PgError{Code: "40001"}, ErrUnavailable},
		{context.DeadlineExceeded, ErrUnavailable},
		{nil, nil},
	} {
		if got := classify(c.err); got != c.want {
			t.Errorf("classify(%v) = %v, want %v", c.err, got, c.want)
		}
	}
}

func TestStatements(t *testing.T) {
	if sqlAppend != `INSERT INTO audit.events (action, outcome, reason, operation, resource_type, resource_id, subject_user_id, `+
		`organization_id, group_id, ttl_seconds) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)` {
		t.Errorf("append statement changed: %s", sqlAppend)
	}
	for _, c := range []string{"actor_kind", "actor_user_id", "occurred_at", "request_id", "trace_id", "schema_version", " id,"} {
		if strings.Contains(strings.SplitN(sqlAppend, "VALUES", 2)[0], c) {
			t.Errorf("append names a server-owned column %q", c)
		}
	}
	if !strings.Contains(sqlReadTenant, "WHERE organization_id = $1 AND occurred_at >= $2 AND (occurred_at, id) < ($3, $4) "+
		"ORDER BY occurred_at DESC, id DESC LIMIT $5") || strings.Count(sqlReadTenant, "$") != 5 {
		t.Errorf("tenant read changed: %s", sqlReadTenant)
	}
	if sqlRevocationGap != `SELECT coalesce(pg_catalog.current_setting('audit.revocation_gap', true), '')` {
		t.Errorf("gap read changed: %s", sqlRevocationGap)
	}
}

type gapTx struct {
	pgx.Tx
	v string
}

type gapRow struct{ v string }

func (r gapRow) Scan(dest ...any) error { *dest[0].(*string) = r.v; return nil }

func (x gapTx) QueryRow(context.Context, string, ...any) pgx.Row { return gapRow{x.v} }

func TestRevocationGapAndUUID(t *testing.T) {
	for v, want := range map[string]string{
		"": "", "session.revoked": "session.revoked", "user.disabled,session.revoked": "user.disabled,session.revoked",
	} {
		got, err := RevocationGap(context.Background(), gapTx{v: v})
		if err != nil || strings.Join(got, ",") != want {
			t.Errorf("RevocationGap(%q) = %v %v", v, got, err)
		}
	}
	for _, v := range []string{"x", "session.revoked,", ",user.disabled", "session.revoked,other"} {
		if _, err := RevocationGap(context.Background(), gapTx{v: v}); !errors.Is(err, ErrUnavailable) {
			t.Errorf("RevocationGap(%q): %v, want ErrUnavailable", v, err)
		}
	}
	u, ok := ParseUUID("0f8fad5b-d9cb-469f-a165-70867728950e")
	if !ok || u[0] != 0x0f || u[15] != 0x0e {
		t.Errorf("ParseUUID: %x %v", u, ok)
	}
	for _, bad := range []string{"", "0F8FAD5B-D9CB-469F-A165-70867728950E", "0f8fad5bd9cb469fa16570867728950e", "0f8fad5b-d9cb-469f-a165-70867728950g"} {
		if _, ok := ParseUUID(bad); ok {
			t.Errorf("ParseUUID(%q) accepted", bad)
		}
	}
}
