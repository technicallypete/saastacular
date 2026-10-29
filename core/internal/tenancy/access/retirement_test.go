package access

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
	"core/internal/tenancy"
)

// fakeFinalizer models identity's finalization contract: it calls the check calls times (once in identity) with a nil
// transaction, forwards a check error unchanged, and otherwise returns result.
type fakeFinalizer struct {
	calls  int
	result error
	users  []string
}

func (f *fakeFinalizer) FinalizeRetirement(ctx context.Context, userID string, check identity.RetirementCheck) error {
	f.users = append(f.users, userID)
	for range f.calls {
		if err := check(ctx, nil); err != nil {
			return err
		}
	}
	return f.result
}

func fixedRemaining(scopes []ScopeRef, orgs []tenancy.OrganizationID, err error) remainingReader {
	return func(context.Context, pgx.Tx) ([]ScopeRef, []tenancy.OrganizationID, error) { return scopes, orgs, err }
}

func newProgress() *RetirementProgress {
	return &RetirementProgress{Blockers: blockers(nil), Departed: []tenancy.OrganizationID{uid[tenancy.OrganizationID](0x01)},
		AlreadyDeparted: []tenancy.OrganizationID{}, Remaining: []tenancy.OrganizationID{}}
}

// TestRetireUserFinalize (G019 F-01, MEM-14 unit half; NOT EXECUTED while the native test hold applies): the final
// check's outcome applied to the progress — remaining memberships or sole-owned scopes refuse (conflict,
// retirement_blocked) with the complete locked lists and no disable; an unknown COMMIT or any other error is
// unavailable with Disabled false and the departures kept; a re-run that confirms disables; a success without exactly one
// check call is never claimed.
func TestRetireUserFinalize(t *testing.T) {
	ctx := context.Background()
	user := uid[tenancy.UserID](0x77)
	c, d := uid[tenancy.OrganizationID](0x30), uid[tenancy.OrganizationID](0x31)
	scope := ScopeRef{OrganizationID: c, ScopeType: tenancy.ScopeOrganization, ScopeID: c}
	many := make([]tenancy.OrganizationID, retirementPage+5)
	for i := range many {
		many[i][15] = byte(i + 1)
		many[i][14] = 0x40
	}

	for name, tc := range map[string]struct {
		fin       *fakeFinalizer
		read      remainingReader
		code      Code
		field     string
		disabled  bool
		remaining []tenancy.OrganizationID
		blockers  int
	}{
		"late membership":      {&fakeFinalizer{calls: 1}, fixedRemaining(nil, []tenancy.OrganizationID{c}, nil), CodeConflict, fieldRetirementBlocked, false, []tenancy.OrganizationID{c}, 0},
		"late sole ownership":  {&fakeFinalizer{calls: 1}, fixedRemaining([]ScopeRef{scope}, []tenancy.OrganizationID{c}, nil), CodeConflict, fieldRetirementBlocked, false, []tenancy.OrganizationID{c}, 1},
		"complete remaining":   {&fakeFinalizer{calls: 1}, fixedRemaining(nil, many, nil), CodeConflict, fieldRetirementBlocked, false, many, 0},
		"reader failure":       {&fakeFinalizer{calls: 1}, fixedRemaining(nil, nil, errors.New("read")), CodeUnavailable, "", false, nil, 0},
		"page order":           {&fakeFinalizer{calls: 1}, fixedRemaining(nil, nil, errRetirementOrder), CodeInternal, "", false, nil, 0},
		"unknown commit":       {&fakeFinalizer{calls: 1, result: identity.ErrUnavailable}, fixedRemaining(nil, nil, nil), CodeUnavailable, "", false, nil, 0},
		"user gone":            {&fakeFinalizer{calls: 1, result: identity.ErrInvalidRequest}, fixedRemaining(nil, nil, nil), CodeUnavailable, "", false, nil, 0},
		"check never called":   {&fakeFinalizer{calls: 0}, fixedRemaining(nil, nil, nil), CodeUnavailable, "", false, nil, 0},
		"check called twice":   {&fakeFinalizer{calls: 2}, fixedRemaining(nil, nil, nil), CodeUnavailable, "", false, nil, 0},
		"confirmed":            {&fakeFinalizer{calls: 1}, fixedRemaining(nil, nil, nil), "", "", true, nil, 0},
		"remaining, two calls": {&fakeFinalizer{calls: 2}, fixedRemaining(nil, []tenancy.OrganizationID{d}, nil), CodeConflict, fieldRetirementBlocked, false, []tenancy.OrganizationID{d}, 0},
	} {
		p := newProgress()
		f := finalizeRetirement(ctx, tc.fin, user, tc.read, p)
		switch {
		case tc.code == "" && f != nil:
			t.Errorf("%s: %+v, want success", name, f)
		case tc.code != "" && (f == nil || f.code != tc.code || f.field != tc.field):
			t.Errorf("%s: %+v, want %s %s", name, f, tc.code, tc.field)
		}
		if p.Disabled != tc.disabled || !slices.Equal(p.Remaining, append([]tenancy.OrganizationID{}, tc.remaining...)) ||
			p.Blockers.Count != tc.blockers || len(p.Departed) != 1 {
			t.Errorf("%s progress: disabled %t, %d remaining, %d blockers, %d departed", name, p.Disabled, len(p.Remaining),
				p.Blockers.Count, len(p.Departed))
		}
		if !slices.Equal(tc.fin.users, []string{user.String()}) {
			t.Errorf("%s: finalizer called for %v", name, tc.fin.users)
		}
	}

	// Resume: an unconfirmed finalization, then a re-run whose finalization confirms.
	fin := &fakeFinalizer{calls: 1, result: identity.ErrUnavailable}
	p := newProgress()
	if f := finalizeRetirement(ctx, fin, user, fixedRemaining(nil, nil, nil), p); f == nil || p.Disabled {
		t.Fatalf("unconfirmed: %+v, disabled %t", f, p.Disabled)
	}
	fin.result = nil
	p = newProgress()
	if f := finalizeRetirement(ctx, fin, user, fixedRemaining(nil, nil, nil), p); f != nil || !p.Disabled {
		t.Fatalf("resumed: %+v, disabled %t", f, p.Disabled)
	}
}

// TestRetireUserNeedsFinalizer: an identity boundary without FinalizeRetirement refuses before any statement (the
// Service has no pool here, so the preflight would panic), and the call writes one retirement.retire record.
func TestRetireUserNeedsFinalizer(t *testing.T) {
	var logs bytes.Buffer
	s := newService(nil, nil, slog.New(slog.NewJSONHandler(&logs, nil)))
	p, err := s.RetireUser(context.Background(), uid[tenancy.UserID](0x77))
	var e *Error
	if !errors.As(err, &e) || e.Code != CodeUnavailable || p.Disabled || len(p.Departed) != 0 {
		t.Errorf("no finalizer: %v, progress %+v", err, p)
	}
	out := logs.String()
	if strings.Count(out, `"msg":"tenancy.operation"`) != 1 || !strings.Contains(out, `"operation":"retirement.retire"`) ||
		!strings.Contains(out, `"outcome":"failed"`) || !strings.Contains(out, `"error_code":"unavailable"`) {
		t.Errorf("retirement record: %s", out)
	}
}

// TestRetirementLogCodes: the record's code is the refusal field for a conflict (retirement_blocked or the departure's
// tenancy code), the closed access code otherwise, and absent on success.
func TestRetirementLogCodes(t *testing.T) {
	for _, c := range []struct {
		f       *failure
		outcome string
		code    string
	}{
		{nil, "ok", ""},
		{&failure{code: CodeConflict, field: fieldRetirementBlocked}, "rejected", fieldRetirementBlocked},
		{conflict(string(tenancy.CodeLastOwner)), "rejected", string(tenancy.CodeLastOwner)},
		{fail(CodeUnavailable, reasonStoreUnavailable), "failed", "unavailable"},
		{fail(CodeInternal, reasonFactInconsistent), "failed", "internal"},
	} {
		var logs bytes.Buffer
		logRetirement(context.Background(), slog.New(slog.NewJSONHandler(&logs, nil)), c.f)
		out := logs.String()
		if !strings.Contains(out, `"outcome":"`+c.outcome+`"`) || (c.code == "") != !strings.Contains(out, "error_code") ||
			(c.code != "" && !strings.Contains(out, `"error_code":"`+c.code+`"`)) {
			t.Errorf("%+v: %s", c.f, out)
		}
	}
}
