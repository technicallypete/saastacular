package access

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/identity"
	"core/internal/observability"
	"core/internal/tenancy"
)

// U-CTX-1..9 (G007 Plan r1 §6.1): the request scope and the B1-B6 binding check. A spy database counts BeginTx (it
// never opens a transaction) and a spy identity counts its calls, so every rejection also proves that it happened
// before identity or the database was reached. The reason is read from the access.decision record.

type spyDB struct{ begins atomic.Int64 }

var errSpyDB = errors.New("spy database: no transaction")

func (d *spyDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	d.begins.Add(1)
	return nil, errSpyDB
}

type spyIdentity struct {
	mu       sync.Mutex
	sessions map[string]identity.SessionInfo
	fail     error
	resolves atomic.Int64
	checks   atomic.Int64
}

func (s *spyIdentity) Session(_ context.Context, credential, _ string) (identity.SessionInfo, error) {
	s.resolves.Add(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail != nil {
		return identity.SessionInfo{}, s.fail
	}
	if info, ok := s.sessions[credential]; ok {
		return info, nil
	}
	return identity.SessionInfo{}, identity.ErrUnauthenticated
}

func (s *spyIdentity) LockLiveSession(context.Context, pgx.Tx, string, string) error {
	s.checks.Add(1)
	return nil
}

func (s *spyIdentity) CheckLiveSession(context.Context, pgx.Tx, string, string) error {
	s.checks.Add(1)
	return nil
}

type logSink struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *logSink) Write(p []byte) (int, error) { l.mu.Lock(); defer l.mu.Unlock(); return l.b.Write(p) }

// last returns the last access.decision record.
func (l *logSink) last(t *testing.T) map[string]string {
	t.Helper()
	l.mu.Lock()
	defer l.mu.Unlock()
	lines := strings.Split(strings.TrimSpace(l.b.String()), "\n")
	var rec map[string]string
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &rec); err != nil {
		t.Fatalf("decision record: %v", err)
	}
	return rec
}

// The two fixture subjects: A and B, each with one session; credential A2 is a second credential of A's session (two
// requests of one browser session share subject and session). G016 security amendment: both sessions are bound to
// orgA's tenant host (testOrigin).
const testOrigin = "https://org-a.localhost:543"

var (
	subjA, subjB = uid[tenancy.UserID](0xaa), uid[tenancy.UserID](0xbb)
	infoA        = identity.SessionInfo{SessionID: "5e55a000-0000-4000-8000-00000000000a", UserID: subjA.String(), MFAAssurance: "unknown",
		OrganizationID: orgA.String()}
	infoB = identity.SessionInfo{SessionID: "5e55b000-0000-4000-8000-00000000000b", UserID: subjB.String(), MFAAssurance: "unknown",
		OrganizationID: orgA.String()}
)

type ctxHarness struct {
	db   *spyDB
	id   *spyIdentity
	logs *logSink
	svc  *Service
}

func newCtxHarness() *ctxHarness {
	h := &ctxHarness{db: &spyDB{}, id: &spyIdentity{sessions: map[string]identity.SessionInfo{"cred-A": infoA, "cred-A2": infoA, "cred-B": infoB}},
		logs: &logSink{}}
	h.svc = newService(h.db, h.id, slog.New(slog.NewJSONHandler(h.logs, nil)))
	return h
}

func (h *ctxHarness) begin(t *testing.T, ctx context.Context, cred string) (context.Context, func()) {
	t.Helper()
	rctx, end, err := h.svc.BeginRequest(ctx, cred, testOrigin)
	if err != nil {
		t.Fatalf("BeginRequest: %v", err)
	}
	return rctx, end
}

// resolved builds, in-package, the context Resolve returns for the scope of rctx (Resolve itself needs a database).
func (h *ctxHarness) resolved(t *testing.T, rctx context.Context, group tenancy.GroupID) *TenantContext {
	t.Helper()
	sc := scopeOf(rctx)
	if sc == nil {
		t.Fatal("no scope")
	}
	return &TenantContext{issuer: h.svc, scope: sc, subject: sc.subject, session: sc.session, organization: orgA, group: group}
}

type facadeOp struct {
	name string
	call func(ctx context.Context, tc *TenantContext) error
}

// facadeOps are the eight guarded operations with well-formed targets and payloads.
func facadeOps(s *Service) []facadeOp {
	return []facadeOp{
		{"GetOrganization", func(ctx context.Context, tc *TenantContext) error { _, err := s.GetOrganization(ctx, tc); return err }},
		{"GetOrganizationPolicy", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.GetOrganizationPolicy(ctx, tc)
			return err
		}},
		{"SetOrganizationPolicy", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.SetOrganizationPolicy(ctx, tc, tenancy.OrganizationPolicy{GroupAdminsEditBusinessProfiles: false})
			return err
		}},
		{"GetGroup", func(ctx context.Context, tc *TenantContext) error { _, err := s.GetGroup(ctx, tc, g1); return err }},
		{"RenameGroup", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.RenameGroup(ctx, tc, g1, "Renamed")
			return err
		}},
		{"ChangeGroupSlug", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.ChangeGroupSlug(ctx, tc, g1, "renamed")
			return err
		}},
		{"GetBusinessProfile", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.GetBusinessProfile(ctx, tc, p1)
			return err
		}},
		{"UpdateBusinessProfile", func(ctx context.Context, tc *TenantContext) error {
			_, err := s.UpdateBusinessProfile(ctx, tc, p1, tenancy.BusinessContext{DisplayName: "Renamed"})
			return err
		}},
	}
}

// wantRejected asserts that call fails with code and reason and reached neither identity nor the database.
func (h *ctxHarness) wantRejected(t *testing.T, what string, code Code, reason Reason, call func() error) {
	t.Helper()
	begins, resolves, checks := h.db.begins.Load(), h.id.resolves.Load(), h.id.checks.Load()
	err := call()
	if CodeOf(err) != code {
		t.Errorf("%s: %v, want %s", what, err, code)
		return
	}
	if got := h.logs.last(t)["reason"]; got != string(reason) {
		t.Errorf("%s: reason %s, want %s", what, got, reason)
	}
	if h.db.begins.Load() != begins || h.id.resolves.Load() != resolves || h.id.checks.Load() != checks {
		t.Errorf("%s: reached BeginTx (%d) or identity (%d, %d)", what, h.db.begins.Load()-begins, h.id.resolves.Load()-resolves, h.id.checks.Load()-checks)
	}
}

// wantReachesTx asserts that call passes the binding check: exactly one BeginTx (the spy refuses it: unavailable).
func (h *ctxHarness) wantReachesTx(t *testing.T, what string, call func() error) {
	t.Helper()
	begins := h.db.begins.Load()
	if err := call(); CodeOf(err) != CodeUnavailable || h.db.begins.Load() != begins+1 {
		t.Errorf("%s: %v with %d BeginTx, want unavailable after exactly one BeginTx", what, err, h.db.begins.Load()-begins)
	}
}

// requestCtx returns a context from the real observability middleware for one request with the given inbound
// X-Request-Id.
func requestCtx(t *testing.T, requestID string) context.Context {
	t.Helper()
	var got context.Context
	mw := observability.Middleware(slog.New(slog.DiscardHandler), http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got = r.Context() }))
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/probe", nil)
	req.Header.Set("X-Request-Id", requestID)
	mw.ServeHTTP(httptest.NewRecorder(), req)
	if observability.RequestID(got) != requestID {
		t.Fatalf("middleware request ID: %q", observability.RequestID(got))
	}
	return got
}

func TestRequestScopeBinding(t *testing.T) {
	t.Run("U-CTX-1 nil, zero-valued, caller-built and foreign contexts", func(t *testing.T) {
		h := newCtxHarness()
		rctx, end := h.begin(t, context.Background(), "cred-A")
		defer end()
		sc := scopeOf(rctx)
		other := newCtxHarness()
		octx, oend := other.begin(t, context.Background(), "cred-A")
		defer oend()
		for _, op := range facadeOps(h.svc) {
			h.wantRejected(t, op.name+" nil", CodeContextDenied, reasonContextForged, func() error { return op.call(rctx, nil) })
			h.wantRejected(t, op.name+" zero", CodeContextDenied, reasonContextForged, func() error { return op.call(rctx, &TenantContext{}) })
			h.wantRejected(t, op.name+" no issuer", CodeContextDenied, reasonContextForged, func() error {
				return op.call(rctx, &TenantContext{scope: sc, subject: sc.subject, session: sc.session, organization: orgA})
			})
			h.wantRejected(t, op.name+" no scope", CodeContextDenied, reasonContextForged, func() error {
				return op.call(rctx, &TenantContext{issuer: h.svc, subject: sc.subject, session: sc.session, organization: orgA})
			})
			h.wantRejected(t, op.name+" other issuer", CodeContextDenied, reasonContextForged, func() error {
				return op.call(rctx, other.resolved(t, octx, tenancy.GroupID{}))
			})
		}
	})

	t.Run("U-CTX-2 no scope or another Service's scope", func(t *testing.T) {
		h := newCtxHarness()
		rctx, end := h.begin(t, context.Background(), "cred-A")
		defer end()
		tc := h.resolved(t, rctx, tenancy.GroupID{})
		other := newCtxHarness()
		octx, oend := other.begin(t, context.Background(), "cred-A")
		defer oend()
		req := ResolveRequest{OrganizationID: orgA.String()}
		for name, ctx := range map[string]context.Context{"Background": context.Background(), "fresh request": requestCtx(t, "fresh-request-0001"),
			"other Service": octx} {
			h.wantRejected(t, "Resolve "+name, CodeContextDenied, reasonScopeMissing, func() error { _, err := h.svc.Resolve(ctx, req); return err })
			for _, op := range facadeOps(h.svc) {
				h.wantRejected(t, op.name+" "+name, CodeContextDenied, reasonScopeMissing, func() error { return op.call(ctx, tc) })
			}
		}
	})

	t.Run("U-CTX-3 same subject and session, distinct requests", func(t *testing.T) {
		for name, base := range map[string][2]context.Context{
			"identical inbound X-Request-Id": {requestCtx(t, "shared-request-id-1"), requestCtx(t, "shared-request-id-1")},
			"absent request IDs":             {context.Background(), context.Background()},
		} {
			h := newCtxHarness()
			r1, end1 := h.begin(t, base[0], "cred-A")
			r2, end2 := h.begin(t, base[1], "cred-A2")
			tc1, tc2 := h.resolved(t, r1, tenancy.GroupID{}), h.resolved(t, r2, tenancy.GroupID{})
			if tc1.subject != tc2.subject || tc1.session != tc2.session || observability.RequestID(r1) != observability.RequestID(r2) {
				t.Fatalf("%s: the two requests must share subject, session and request ID", name)
			}
			for _, op := range facadeOps(h.svc) {
				h.wantRejected(t, op.name+" R1 context under R2 ("+name+")", CodeContextDenied, reasonScopeMismatch, func() error { return op.call(r2, tc1) })
				h.wantRejected(t, op.name+" R2 context under R1 ("+name+")", CodeContextDenied, reasonScopeMismatch, func() error { return op.call(r1, tc2) })
			}
			end1()
			end2()
		}
	})

	t.Run("U-CTX-4 cross-subject substitution", func(t *testing.T) {
		for name, base := range map[string][2]context.Context{
			"equal request IDs":  {requestCtx(t, "equal-request-id-01"), requestCtx(t, "equal-request-id-01")},
			"absent request IDs": {context.Background(), context.Background()},
		} {
			h := newCtxHarness()
			ra, enda := h.begin(t, base[0], "cred-A")
			rb, endb := h.begin(t, base[1], "cred-B")
			tca, tcb := h.resolved(t, ra, tenancy.GroupID{}), h.resolved(t, rb, tenancy.GroupID{})
			for _, op := range facadeOps(h.svc) {
				h.wantRejected(t, op.name+" A under B ("+name+")", CodeContextDenied, reasonScopeMismatch, func() error { return op.call(rb, tca) })
				h.wantRejected(t, op.name+" B under A ("+name+")", CodeContextDenied, reasonScopeMismatch, func() error { return op.call(ra, tcb) })
			}
			enda()
			endb()
		}
	})

	t.Run("U-CTX-5 the subject and session check on its own", func(t *testing.T) {
		h := newCtxHarness()
		rctx, end := h.begin(t, context.Background(), "cred-A")
		defer end()
		sc := scopeOf(rctx)
		for name, tc := range map[string]*TenantContext{
			"other subject": {issuer: h.svc, scope: sc, subject: subjB, session: sc.session, organization: orgA},
			"other session": {issuer: h.svc, scope: sc, subject: sc.subject, session: infoB.SessionID, organization: orgA},
		} {
			for _, op := range facadeOps(h.svc) {
				h.wantRejected(t, op.name+" "+name, CodeContextDenied, reasonSubjectMismatch, func() error { return op.call(rctx, tc) })
			}
		}
	})

	t.Run("U-CTX-6 ended scopes", func(t *testing.T) {
		h := newCtxHarness()
		rctx, end := h.begin(t, context.Background(), "cred-A")
		r2, end2 := h.begin(t, context.Background(), "cred-A")
		defer end2()
		tc, tc2 := h.resolved(t, rctx, tenancy.GroupID{}), h.resolved(t, r2, tenancy.GroupID{})
		end()
		end() // idempotent
		var wg sync.WaitGroup
		for range 8 {
			wg.Add(1)
			go func() { defer wg.Done(); end() }() // safe for concurrent use
		}
		wg.Wait()
		req := ResolveRequest{OrganizationID: orgA.String()}
		detached := context.WithoutCancel(rctx)
		for name, ctx := range map[string]context.Context{"ended": rctx, "WithoutCancel child": detached} {
			h.wantRejected(t, "Resolve "+name, CodeContextDenied, reasonScopeEnded, func() error { _, err := h.svc.Resolve(ctx, req); return err })
			for _, op := range facadeOps(h.svc) {
				h.wantRejected(t, op.name+" "+name, CodeContextDenied, reasonScopeEnded, func() error { return op.call(ctx, tc) })
			}
		}
		// Ending one scope leaves another usable.
		for _, op := range facadeOps(h.svc) {
			h.wantReachesTx(t, op.name+" other scope", func() error { return op.call(r2, tc2) })
		}
		h.wantReachesTx(t, "Resolve other scope", func() error { _, err := h.svc.Resolve(r2, req); return err })
	})

	t.Run("U-CTX-7 nested BeginRequest", func(t *testing.T) {
		h := newCtxHarness()
		rctx, end := h.begin(t, context.Background(), "cred-A")
		defer end()
		other := newCtxHarness()
		octx, oend := other.begin(t, context.Background(), "cred-B")
		defer oend()
		for name, ctx := range map[string]context.Context{"own scope": rctx, "another Service's scope": octx,
			"derived child": context.WithValue(rctx, struct{ k int }{1}, "v")} {
			h.wantRejected(t, "nested "+name, CodeContextDenied, reasonScopeNested, func() error {
				nctx, nend, err := h.svc.BeginRequest(ctx, "cred-B", testOrigin)
				nend()
				if err == nil || nctx != ctx {
					return errors.New("nested scope created")
				}
				return err
			})
		}
	})

	t.Run("U-CTX-8 derived contexts of the same scope pass the binding", func(t *testing.T) {
		h := newCtxHarness()
		rctx, end := h.begin(t, requestCtx(t, "derived-request-01"), "cred-A")
		defer end()
		tc := h.resolved(t, rctx, tenancy.GroupID{})
		deadline, cancel := context.WithTimeout(rctx, time.Minute)
		defer cancel()
		for name, ctx := range map[string]context.Context{"same": rctx, "deadline child": deadline,
			"value child": context.WithValue(rctx, struct{ k int }{2}, "v"), "WithoutCancel child": context.WithoutCancel(rctx)} {
			for _, op := range facadeOps(h.svc) {
				h.wantReachesTx(t, op.name+" "+name, func() error { return op.call(ctx, tc) })
			}
			h.wantReachesTx(t, "Resolve "+name, func() error { _, err := h.svc.Resolve(ctx, ResolveRequest{OrganizationID: orgA.String()}); return err })
		}
	})

	t.Run("U-CTX-9 no correlation, credential or authority field", func(t *testing.T) {
		forbidden := regexp.MustCompile(`(?i)request|trace|span|correlation|credential|token|csrf|cookie|header|claim|assurance|mfa|role|capab|allow|display|provider`)
		for _, v := range []any{TenantContext{}, requestScope{}} {
			rt := reflect.TypeOf(v)
			for i := range rt.NumField() {
				fd := rt.Field(i)
				if forbidden.MatchString(fd.Name) || fd.Type == reflect.TypeOf(observability.TraceContext{}) || fd.IsExported() {
					t.Errorf("%s.%s (%s)", rt.Name(), fd.Name, fd.Type)
				}
			}
		}
	})
}

// BeginRequest and Resolve outcomes before any transaction: authentication failures, identity outages and malformed
// selectors.
func TestBeginRequestAndSelectors(t *testing.T) {
	h := newCtxHarness()
	for _, cred := range []string{"", "garbage", "cred-unknown"} {
		begins := h.db.begins.Load()
		_, end, err := h.svc.BeginRequest(context.Background(), cred, testOrigin)
		end()
		if CodeOf(err) != CodeUnauthenticated || h.logs.last(t)["reason"] != string(reasonSessionInvalid) || h.db.begins.Load() != begins {
			t.Errorf("credential %q: %v", cred, err)
		}
	}
	h.id.fail = identity.ErrUnavailable
	if _, end, err := h.svc.BeginRequest(context.Background(), "cred-A", testOrigin); CodeOf(err) != CodeUnavailable || h.logs.last(t)["reason"] != string(reasonStoreUnavailable) {
		t.Errorf("identity outage: %v", err)
	} else {
		end()
	}
	h.id.fail = nil
	// G016 security amendment: a session without an Organization binding fails closed before any transaction.
	h.id.mu.Lock()
	h.id.sessions["cred-unbound"] = identity.SessionInfo{SessionID: infoA.SessionID, UserID: infoA.UserID, MFAAssurance: "unknown"}
	h.id.mu.Unlock()
	if _, end, err := h.svc.BeginRequest(context.Background(), "cred-unbound", testOrigin); CodeOf(err) != CodeUnavailable ||
		h.logs.last(t)["reason"] != string(reasonStoreUnavailable) {
		t.Errorf("unbound session: %v", err)
	} else {
		end()
	}
	rctx, end := h.begin(t, context.Background(), "cred-A")
	defer end()
	if rec := h.logs.last(t); rec["operation"] != opBeginRequest || rec["outcome"] != outcomeEstablished || rec["reason"] != string(reasonContextEstablished) {
		t.Errorf("BeginRequest record: %v", rec)
	} else if _, ok := rec["capability"]; ok {
		t.Error("BeginRequest record carries a capability")
	}
	// The browser scope resolves only its session's Organization: another one is organization_not_permitted before any
	// context snapshot (the denial is recorded without the requested Organization).
	checks := h.id.checks.Load()
	if _, err := h.svc.Resolve(rctx, ResolveRequest{OrganizationID: orgB.String()}); CodeOf(err) != CodeContextDenied ||
		h.logs.last(t)["reason"] != string(reasonOrganizationNotPermitted) || h.id.checks.Load() != checks {
		t.Errorf("Resolve of another Organization in a browser scope: %v", err)
	}
	for name, req := range map[string]ResolveRequest{
		"empty":        {},
		"all zero":     {OrganizationID: "00000000-0000-0000-0000-000000000000"},
		"braced":       {OrganizationID: "{" + orgA.String() + "}"},
		"unhyphenated": {OrganizationID: strings.ReplaceAll(orgA.String(), "-", "")},
		"urn":          {OrganizationID: "urn:uuid:" + orgA.String()},
		"bad Group":    {OrganizationID: orgA.String(), GroupID: "not-a-uuid"},
		"zero Group":   {OrganizationID: orgA.String(), GroupID: "00000000-0000-0000-0000-000000000000"},
	} {
		h.wantRejected(t, "Resolve "+name, CodeInvalidInput, reasonSelectorInvalid, func() error {
			_, err := h.svc.Resolve(rctx, req)
			var e *Error
			if errors.As(err, &e) && e.Field != string(tenancy.CodeInvalidID) {
				return errors.New("field " + e.Field)
			}
			return err
		})
	}
	tc := h.resolved(t, rctx, tenancy.GroupID{})
	for name, call := range map[string]func() error{
		"zero Group target":   func() error { _, err := h.svc.GetGroup(rctx, tc, tenancy.GroupID{}); return err },
		"zero profile target": func() error { _, err := h.svc.GetBusinessProfile(rctx, tc, tenancy.BusinessProfileID{}); return err },
		"invalid name":        func() error { _, err := h.svc.RenameGroup(rctx, tc, g1, " "); return err },
		"invalid slug":        func() error { _, err := h.svc.ChangeGroupSlug(rctx, tc, g1, "Not A Slug"); return err },
		"reserved slug":       func() error { _, err := h.svc.ChangeGroupSlug(rctx, tc, g1, "admin"); return err },
		"invalid profile": func() error {
			_, err := h.svc.UpdateBusinessProfile(rctx, tc, p1, tenancy.BusinessContext{})
			return err
		},
	} {
		h.wantRejected(t, name, CodeInvalidInput, reasonInputInvalid, call)
	}
	var e *Error
	if _, err := h.svc.ChangeGroupSlug(rctx, tc, g1, "admin"); !errors.As(err, &e) || e.Field != string(tenancy.CodeSlugReserved) {
		t.Errorf("reserved slug field: %v", err)
	}
	if _, err := h.svc.UpdateBusinessProfile(rctx, tc, p1, tenancy.BusinessContext{}); !errors.As(err, &e) || e.Field != string(tenancy.CodeInvalidDisplayName) {
		t.Errorf("profile field: %v", err)
	}
	if got := (&Error{Code: CodeForbidden, Field: "x"}).Error(); got != "access: forbidden" {
		t.Errorf("Error(): %q", got)
	}
}
