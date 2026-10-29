package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"core/internal/identity"
	"core/internal/signup"
)

// G062 (Plan r5 §4.2, §4.5; SUP-3, SUP-9): the signup HTTP adapter over a fake signup service — bounds, the closed
// error vocabulary, the credential boundary, CSRF, no-store/nosniff/no-CORS, one closed signup.request record per
// request, the pinned routes and every bucket of §4.5 (iii)-(iv). The service and its SQL are proven against PostgreSQL
// by TestSignupProvisioning/TestSignupSlugRace/TestSignupContinuation (core/internal/signup, integration). Since G057 the
// tenant origin is a generated platform label host, never the slug's (D-5: the DTO keeps slug and target_origin).

const (
	root      = "https://localhost:543"
	tenant    = "https://mfrggzdfmztwq2lknnwg23tpoa.localhost:543"
	goodCred  = "cred-good"
	otherCred = "cred-other"
	binding   = "binding-value"
)

var csrf = bytes.Repeat([]byte{7}, 32)

// fakeService records the inputs the adapter forwards and answers from its fields.
type fakeService struct {
	unconfigured bool
	rootOrigin   string
	creds        map[string]identity.SignupCredentialInfo // credential -> resolved info
	err          error                                    // the next operation's error (after resolution)
	calls        []string
	gotCred      string
	gotOrigin    string
	gotBinding   string
	gotArgs      []string
}

func newFake() *fakeService {
	info := identity.SignupCredentialInfo{CSRFToken: csrf, ExpiresAt: time.Unix(1_800_000_000, 0)}
	prov := info
	prov.Provisioned, prov.Slug, prov.TargetOrigin, prov.OrganizationID, prov.Generation = true, "acme", tenant, "org-id", 1
	return &fakeService{rootOrigin: root, creds: map[string]identity.SignupCredentialInfo{goodCred: info, otherCred: prov}}
}

func (f *fakeService) Configured() bool   { return !f.unconfigured }
func (f *fakeService) RootOrigin() string { return f.rootOrigin }

func (f *fakeService) Resolve(_ context.Context, credential, origin string) (identity.SignupCredentialInfo, error) {
	f.calls = append(f.calls, "resolve")
	f.gotCred, f.gotOrigin = credential, origin
	info, ok := f.creds[credential]
	if !ok {
		return identity.SignupCredentialInfo{}, &signup.Error{Code: signup.CodeUnauthenticated}
	}
	return info, nil
}

func (f *fakeService) SuggestSlug(_ context.Context, _ identity.SignupCredentialInfo, name string) (signup.Suggestion, error) {
	f.calls, f.gotArgs = append(f.calls, "suggest"), []string{name}
	if f.err != nil {
		return signup.Suggestion{}, f.err
	}
	if name == "!!!" {
		return signup.Suggestion{}, nil
	}
	return signup.Suggestion{Suggested: "acme", Available: true}, nil
}

func (f *fakeService) Provision(_ context.Context, _ identity.SignupCredentialInfo, name, slug string) (signup.Provisioned, error) {
	f.calls, f.gotArgs = append(f.calls, "provision"), []string{name, slug}
	if f.err != nil {
		return signup.Provisioned{}, f.err
	}
	return signup.Provisioned{Outcome: signup.OutcomeCreated, OrganizationID: "org-id", Slug: slug, Generation: 1,
		Continue: tenant + "/signup/continue?ref=R"}, nil
}

func (f *fakeService) Challenge(_ context.Context, origin, binding, ref string) (signup.Challenge, error) {
	f.calls, f.gotOrigin, f.gotBinding, f.gotArgs = append(f.calls, "challenge"), origin, binding, []string{ref}
	if f.err != nil {
		return signup.Challenge{}, f.err
	}
	return signup.Challenge{Value: "C", MaxAge: 120, Prove: root + "/signup/prove?n=N"}, nil
}

func (f *fakeService) Prove(_ context.Context, _ identity.SignupCredentialInfo, nonce string) (string, error) {
	f.calls, f.gotArgs = append(f.calls, "prove"), []string{nonce}
	if f.err != nil {
		return "", f.err
	}
	return tenant + "/signup/complete?p=P", nil
}

func (f *fakeService) Complete(_ context.Context, origin, binding, challenge, proof string) (identity.LoginOutput, error) {
	f.calls, f.gotOrigin, f.gotBinding, f.gotArgs = append(f.calls, "complete"), origin, binding, []string{challenge, proof}
	if f.err != nil {
		return identity.LoginOutput{}, f.err
	}
	return identity.LoginOutput{AuthorizationURL: "https://provider.test/authorize?x", Binding: "B2", BindingMaxAge: 600}, nil
}

func (f *fakeService) Abandon(context.Context, identity.SignupCredentialInfo) error {
	f.calls = append(f.calls, "abandon")
	return f.err
}

// harness is one handler on a mux with a captured log.
type harness struct {
	f   *fakeService
	h   *Handler
	mux *http.ServeMux
	log *bytes.Buffer
}

func newHarness(f *fakeService) *harness {
	buf := &bytes.Buffer{}
	h := newHandler(f, slog.New(slog.NewJSONHandler(buf, nil)))
	h.userOf = func(info identity.SignupCredentialInfo) string { return "user-1" }
	mux := http.NewServeMux()
	h.Register(mux)
	return &harness{f: f, h: h, mux: mux, log: buf}
}

type req struct {
	method, path, body string
	header             map[string][]string
}

func (x *harness) do(t *testing.T, r req) (int, http.Header, map[string]any) {
	t.Helper()
	hr := httptest.NewRequest(r.method, r.path, strings.NewReader(r.body))
	if r.body != "" {
		hr.Header.Set("Content-Type", "application/json")
	}
	for k, vs := range r.header {
		hr.Header.Del(k)
		for _, v := range vs {
			hr.Header.Add(k, v)
		}
	}
	w := httptest.NewRecorder()
	x.mux.ServeHTTP(w, hr)
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("%s %s: non-JSON answer %q", r.method, r.path, w.Body.String())
	}
	return w.Code, w.Header(), out
}

func auth(cred string) map[string][]string {
	return map[string][]string{"Authorization": {"Signup " + cred}}
}

func csrfToken() string { return base64.RawURLEncoding.EncodeToString(csrf) }

func js(v map[string]any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// records returns the signup.request records written so far and resets the buffer.
func (x *harness) records(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(x.log.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("log line %q: %v", line, err)
		}
		out = append(out, m)
	}
	x.log.Reset()
	return out
}

// TestSignupHTTP (SUP-9): every route's success shape and its closed refusals; headers; one closed log record per
// request that never carries a credential, a secret, a name, a slug or an origin.
func TestSignupHTTP(t *testing.T) {
	createBody := js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "Acme Inc", "slug": "acme"})
	cases := []struct {
		name   string
		r      req
		err    error // the fake's operation error
		code   int
		want   map[string]any
		reason string // the logged reason ("" = ok)
		op     string
		calls  []string
	}{
		{"status pending", req{"GET", "/internal/v1/signup/status?origin=" + root, "", auth(goodCred)}, nil, 200,
			map[string]any{"status": "pending", "csrf_token": csrfToken()}, "", "status", []string{"resolve"}},
		{"status provisioned", req{"GET", "/internal/v1/signup/status?origin=" + root, "", auth(otherCred)}, nil, 200,
			map[string]any{"status": "created", "csrf_token": csrfToken(), "organization": map[string]any{"slug": "acme", "target_origin": tenant}},
			"", "status", []string{"resolve"}},
		{"status no origin", req{"GET", "/internal/v1/signup/status", "", auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "status", nil},
		{"status extra parameter", req{"GET", "/internal/v1/signup/status?origin=" + root + "&x=1", "", auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "status", nil},
		{"status repeated origin", req{"GET", "/internal/v1/signup/status?origin=" + root + "&origin=" + root, "", auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "status", nil},
		{"status www origin", req{"GET", "/internal/v1/signup/status?origin=https://www.localhost:543", "", auth(goodCred)}, nil, 403,
			map[string]any{"error": "origin_rejected"}, "origin_rejected", "status", nil},
		{"status tenant origin", req{"GET", "/internal/v1/signup/status?origin=" + tenant, "", auth(goodCred)}, nil, 403,
			map[string]any{"error": "origin_rejected"}, "origin_rejected", "status", nil},
		{"status no credential", req{"GET", "/internal/v1/signup/status?origin=" + root, "", nil}, nil, 401,
			map[string]any{"error": "unauthenticated"}, "unauthenticated", "status", nil},
		{"status session scheme", req{"GET", "/internal/v1/signup/status?origin=" + root, "", map[string][]string{"Authorization": {"Session " + goodCred}}}, nil, 401,
			map[string]any{"error": "unauthenticated"}, "unauthenticated", "status", nil},
		{"status enrollment scheme", req{"GET", "/internal/v1/signup/status?origin=" + root, "", map[string][]string{"Authorization": {"Enrollment " + goodCred}}}, nil, 401,
			map[string]any{"error": "unauthenticated"}, "unauthenticated", "status", nil},
		{"status repeated credential", req{"GET", "/internal/v1/signup/status?origin=" + root, "", map[string][]string{"Authorization": {"Signup " + goodCred, "Signup " + goodCred}}}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "status", nil},
		{"status dead credential", req{"GET", "/internal/v1/signup/status?origin=" + root, "", auth("dead")}, nil, 401,
			map[string]any{"error": "unauthenticated"}, "unauthenticated", "status", []string{"resolve"}},

		{"slug", req{"POST", "/internal/v1/signup/slug", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "Acme"}), auth(goodCred)}, nil, 200,
			map[string]any{"suggested": "acme", "available": true}, "", "slug", []string{"resolve", "suggest"}},
		{"slug without base", req{"POST", "/internal/v1/signup/slug", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "!!!"}), auth(goodCred)}, nil, 200,
			map[string]any{"suggested": nil, "available": false}, "", "slug", []string{"resolve", "suggest"}},
		{"slug invalid name", req{"POST", "/internal/v1/signup/slug", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": " "}), auth(goodCred)},
			&signup.Error{Code: signup.CodeInvalidRequest, Field: "name"}, 400, map[string]any{"error": "invalid_request", "field": "name"}, "invalid_request", "slug", []string{"resolve", "suggest"}},
		{"slug csrf", req{"POST", "/internal/v1/signup/slug", js(map[string]any{"origin": root, "csrf_token": "AAAA", "name": "Acme"}), auth(goodCred)}, nil, 403,
			map[string]any{"error": "csrf_failed"}, "csrf_failed", "slug", []string{"resolve"}},
		{"slug missing csrf", req{"POST", "/internal/v1/signup/slug", js(map[string]any{"origin": root, "name": "Acme"}), auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "slug", nil},

		{"create", req{"POST", "/internal/v1/signup/organizations", createBody, auth(goodCred)}, nil, 201,
			map[string]any{"outcome": "created", "organization": map[string]any{"id": "org-id", "slug": "acme"}, "continue": tenant + "/signup/continue?ref=R"},
			"", "create", []string{"resolve", "provision"}},
		{"create slug taken", req{"POST", "/internal/v1/signup/organizations", createBody, auth(goodCred)}, &signup.Error{Code: signup.CodeSlugTaken}, 409,
			map[string]any{"error": "conflict", "field": "slug_taken"}, "slug_taken", "create", []string{"resolve", "provision"}},
		{"create intent used", req{"POST", "/internal/v1/signup/organizations", createBody, auth(goodCred)}, &signup.Error{Code: signup.CodeIntentUsed}, 409,
			map[string]any{"error": "conflict", "field": "intent_used"}, "intent_used", "create", []string{"resolve", "provision"}},
		{"create invalid slug", req{"POST", "/internal/v1/signup/organizations", createBody, auth(goodCred)}, &signup.Error{Code: signup.CodeInvalidRequest, Field: "slug"}, 400,
			map[string]any{"error": "invalid_request", "field": "slug"}, "invalid_request", "create", []string{"resolve", "provision"}},
		{"create user disabled", req{"POST", "/internal/v1/signup/organizations", createBody, auth(goodCred)}, &signup.Error{Code: signup.CodeUserDisabled}, 403,
			map[string]any{"error": "user_disabled"}, "user_disabled", "create", []string{"resolve", "provision"}},
		{"create credential gone", req{"POST", "/internal/v1/signup/organizations", createBody, auth(goodCred)}, &signup.Error{Code: signup.CodeUnauthenticated}, 401,
			map[string]any{"error": "unauthenticated"}, "unauthenticated", "create", []string{"resolve", "provision"}},
		{"create unknown commit", req{"POST", "/internal/v1/signup/organizations", createBody, auth(goodCred)}, errors.New("driver text"), 503,
			map[string]any{"error": "signup_unavailable"}, "signup_unavailable", "create", []string{"resolve", "provision"}},
		{"create unknown field", req{"POST", "/internal/v1/signup/organizations", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "A", "slug": "a", "role": "owner"}), auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "create", nil},
		{"create missing slug", req{"POST", "/internal/v1/signup/organizations", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "A"}), auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "create", nil},
		{"create organization id refused", req{"POST", "/internal/v1/signup/organizations", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "A", "slug": "a", "organization_id": "x"}), auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "create", nil},
		{"create query refused", req{"POST", "/internal/v1/signup/organizations?origin=" + root, createBody, auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "create", nil},
		{"create trailing value", req{"POST", "/internal/v1/signup/organizations", createBody + " {}", auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "create", nil},
		{"create over 8 KiB", req{"POST", "/internal/v1/signup/organizations", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": strings.Repeat("a", 8<<10), "slug": "a"}), auth(goodCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "create", nil},
		{"create session scheme", req{"POST", "/internal/v1/signup/organizations", createBody, map[string][]string{"Authorization": {"Session " + goodCred}}}, nil, 401,
			map[string]any{"error": "unauthenticated"}, "unauthenticated", "create", nil},
		{"create wrong origin", req{"POST", "/internal/v1/signup/organizations", js(map[string]any{"origin": "https://www.localhost:543", "csrf_token": csrfToken(), "name": "A", "slug": "a"}), auth(goodCred)}, nil, 403,
			map[string]any{"error": "origin_rejected"}, "origin_rejected", "create", nil},

		{"prove", req{"POST", "/internal/v1/signup/prove", js(map[string]any{"origin": root, "nonce": "N"}), auth(otherCred)}, nil, 200,
			map[string]any{"complete": tenant + "/signup/complete?p=P"}, "", "prove", []string{"resolve", "prove"}},
		{"prove context mismatch", req{"POST", "/internal/v1/signup/prove", js(map[string]any{"origin": root, "nonce": "N"}), auth(otherCred)}, &signup.Error{Code: signup.CodeContextMismatch}, 403,
			map[string]any{"error": "signup_context_mismatch"}, "signup_context_mismatch", "prove", []string{"resolve", "prove"}},
		{"prove without credential", req{"POST", "/internal/v1/signup/prove", js(map[string]any{"origin": root, "nonce": "N"}), nil}, nil, 401,
			map[string]any{"error": "unauthenticated"}, "unauthenticated", "prove", nil},
		{"prove session scheme", req{"POST", "/internal/v1/signup/prove", js(map[string]any{"origin": root, "nonce": "N"}), map[string][]string{"Authorization": {"Session " + otherCred}}}, nil, 401,
			map[string]any{"error": "unauthenticated"}, "unauthenticated", "prove", nil},
		{"prove csrf member refused", req{"POST", "/internal/v1/signup/prove", js(map[string]any{"origin": root, "nonce": "N", "csrf_token": csrfToken()}), auth(otherCred)}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "prove", nil},
		{"prove tenant origin", req{"POST", "/internal/v1/signup/prove", js(map[string]any{"origin": tenant, "nonce": "N"}), auth(otherCred)}, nil, 403,
			map[string]any{"error": "origin_rejected"}, "origin_rejected", "prove", nil},

		{"abandon", req{"POST", "/internal/v1/signup/abandon", js(map[string]any{"origin": root, "csrf_token": csrfToken()}), auth(goodCred)}, nil, 200,
			map[string]any{}, "", "abandon", []string{"resolve", "abandon"}},
		{"abandon csrf", req{"POST", "/internal/v1/signup/abandon", js(map[string]any{"origin": root, "csrf_token": ""}), auth(goodCred)}, nil, 403,
			map[string]any{"error": "csrf_failed"}, "csrf_failed", "abandon", []string{"resolve"}},

		{"challenge", req{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": tenant, "ref": "R"}), map[string][]string{"X-Login-Binding": {binding}}}, nil, 200,
			map[string]any{"challenge": map[string]any{"value": "C", "max_age": float64(120)}, "prove": root + "/signup/prove?n=N"}, "", "challenge", []string{"challenge"}},
		{"challenge refused", req{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": tenant, "ref": "R"}), nil}, &signup.Error{Code: signup.CodeContinuationInvalid}, 404,
			map[string]any{"error": "continuation_invalid"}, "continuation_invalid", "challenge", []string{"challenge"}},
		{"challenge malformed origin", req{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": "http://mfrggzdfmztwq2lknnwg23tpoa.localhost:543", "ref": "R"}), nil}, nil, 404,
			map[string]any{"error": "continuation_invalid"}, "continuation_invalid", "challenge", nil},
		{"challenge origin with path", req{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": tenant + "/x", "ref": "R"}), nil}, nil, 404,
			map[string]any{"error": "continuation_invalid"}, "continuation_invalid", "challenge", nil},
		{"challenge repeated binding", req{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": tenant, "ref": "R"}), map[string][]string{"X-Login-Binding": {"a", "b"}}}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "challenge", nil},
		{"challenge missing ref", req{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": tenant}), nil}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "challenge", nil},
		{"challenge unavailable", req{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": tenant, "ref": "R"}), nil}, &signup.Error{Code: signup.CodeUnavailable}, 503,
			map[string]any{"error": "signup_unavailable"}, "signup_unavailable", "challenge", []string{"challenge"}},

		{"complete", req{"POST", "/internal/v1/signup/complete", js(map[string]any{"origin": tenant, "challenge": "C", "proof": "P"}), map[string][]string{"X-Login-Binding": {binding}}}, nil, 200,
			map[string]any{"authorization_url": "https://provider.test/authorize?x", "binding": map[string]any{"value": "B2", "max_age": float64(600)}}, "", "complete", []string{"complete"}},
		{"complete refused", req{"POST", "/internal/v1/signup/complete", js(map[string]any{"origin": tenant, "challenge": "C", "proof": "P"}), nil}, &signup.Error{Code: signup.CodeContinuationInvalid}, 404,
			map[string]any{"error": "continuation_invalid"}, "continuation_invalid", "complete", []string{"complete"}},
		{"complete missing proof", req{"POST", "/internal/v1/signup/complete", js(map[string]any{"origin": tenant, "challenge": "C"}), nil}, nil, 400,
			map[string]any{"error": "invalid_request"}, "invalid_request", "complete", nil},
	}
	for _, c := range cases {
		x := newHarness(newFake())
		x.f.err = c.err
		code, hd, out := x.do(t, c.r)
		if code != c.code || !equalJSON(out, c.want) {
			t.Errorf("%s: %d %v, want %d %v", c.name, code, out, c.code, c.want)
		}
		if hd.Get("Cache-Control") != "no-store" || hd.Get("X-Content-Type-Options") != "nosniff" ||
			hd.Get("Content-Type") != "application/json" {
			t.Errorf("%s: headers %v", c.name, hd)
		}
		for k := range hd {
			if strings.HasPrefix(strings.ToLower(k), "access-control-") {
				t.Errorf("%s: CORS header %s", c.name, k)
			}
		}
		if !slices.Equal(x.f.calls, c.calls) {
			t.Errorf("%s: service calls %v, want %v", c.name, x.f.calls, c.calls)
		}
		raw := x.log.String()
		recs := x.records(t)
		if len(recs) != 1 || recs[0]["msg"] != signup.EventRequest || recs[0]["operation"] != c.op {
			t.Errorf("%s: log records %v", c.name, recs)
			continue
		}
		if c.reason == "" {
			if recs[0]["outcome"] != "ok" || recs[0]["reason"] != nil {
				t.Errorf("%s: log %v, want ok", c.name, recs[0])
			}
		} else if recs[0]["reason"] != c.reason {
			t.Errorf("%s: logged reason %v, want %s", c.name, recs[0]["reason"], c.reason)
		}
		for _, secret := range []string{goodCred, otherCred, binding, csrfToken(), "Acme", "acme", "localhost", "mfrggzdfmztwq2lknnwg23tpoa", "?ref=", "\"N\"", "\"P\"", "\"C\""} {
			if strings.Contains(raw, secret) {
				t.Errorf("%s: the log carries %q: %s", c.name, secret, raw)
			}
		}
	}
}

// TestSignupForwarding: the root routes forward exactly the Signup credential and the root origin; the continuation hops
// forward the exact origin and the login binding and never consult a credential (an Authorization header is ignored).
func TestSignupForwarding(t *testing.T) {
	x := newHarness(newFake())
	x.do(t, req{"POST", "/internal/v1/signup/organizations", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "Acme Inc", "slug": "acme"}), auth(goodCred)})
	if x.f.gotCred != goodCred || x.f.gotOrigin != root || !slices.Equal(x.f.gotArgs, []string{"Acme Inc", "acme"}) {
		t.Errorf("create forwarded %q %q %v", x.f.gotCred, x.f.gotOrigin, x.f.gotArgs)
	}
	for _, path := range []string{"/internal/v1/signup/challenge", "/internal/v1/signup/complete"} {
		x := newHarness(newFake())
		body := js(map[string]any{"origin": tenant, "ref": "R"})
		if strings.HasSuffix(path, "complete") {
			body = js(map[string]any{"origin": tenant, "challenge": "C", "proof": "P"})
		}
		code, _, _ := x.do(t, req{"POST", path, body, map[string][]string{"Authorization": {"Signup " + goodCred}, "X-Login-Binding": {binding}}})
		if code != 200 || slices.Contains(x.f.calls, "resolve") || x.f.gotCred != "" || x.f.gotOrigin != tenant || x.f.gotBinding != binding {
			t.Errorf("%s: %d calls %v cred %q origin %q binding %q", path, code, x.f.calls, x.f.gotCred, x.f.gotOrigin, x.f.gotBinding)
		}
	}
	// No binding header: the hop forwards none (identity then uses a fresh random binding: the shape only).
	x = newHarness(newFake())
	x.do(t, req{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": tenant, "ref": "R"}), nil})
	if x.f.gotBinding != "" {
		t.Errorf("binding %q forwarded without a header", x.f.gotBinding)
	}
}

// TestSignupMediaType: exactly one application/json Content-Type (charset utf-8 only) or invalid_request, before any
// credential is read.
func TestSignupMediaType(t *testing.T) {
	body := js(map[string]any{"origin": root, "csrf_token": csrfToken()})
	for _, ct := range [][]string{{"text/plain"}, {"application/json; charset=latin1"}, {"application/json", "application/json"}, {"application/json; x=y"}, {}} {
		x := newHarness(newFake())
		hr := httptest.NewRequest("POST", "/internal/v1/signup/abandon", strings.NewReader(body))
		for _, v := range ct {
			hr.Header.Add("Content-Type", v)
		}
		hr.Header.Set("Authorization", "Signup "+goodCred)
		w := httptest.NewRecorder()
		x.mux.ServeHTTP(w, hr)
		if w.Code != 400 || !strings.Contains(w.Body.String(), `"invalid_request"`) || len(x.f.calls) != 0 {
			t.Errorf("%v: %d %s calls %v", ct, w.Code, w.Body.String(), x.f.calls)
		}
	}
	x := newHarness(newFake())
	code, _, _ := x.do(t, req{"POST", "/internal/v1/signup/abandon", body, map[string][]string{"Authorization": {"Signup " + goodCred}, "Content-Type": {"application/json; charset=UTF-8"}}})
	if code != 200 {
		t.Errorf("charset utf-8: %d", code)
	}
}

// TestSignupUnconfigured: an unconfigured service (no identity, root origin or platform domain) answers
// signup_unconfigured on every route without a statement.
func TestSignupUnconfigured(t *testing.T) {
	for _, r := range []req{
		{"GET", "/internal/v1/signup/status?origin=" + root, "", auth(goodCred)},
		{"POST", "/internal/v1/signup/organizations", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "A", "slug": "a"}), auth(goodCred)},
		{"POST", "/internal/v1/signup/prove", js(map[string]any{"origin": root, "nonce": "N"}), auth(goodCred)},
		{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": tenant, "ref": "R"}), nil},
		{"POST", "/internal/v1/signup/complete", js(map[string]any{"origin": tenant, "challenge": "C", "proof": "P"}), nil},
	} {
		f := newFake()
		f.unconfigured = true
		x := newHarness(f)
		if code, _, out := x.do(t, r); code != 503 || out["error"] != "signup_unconfigured" || len(f.calls) != 0 {
			t.Errorf("%s %s: %d %v calls %v", r.method, r.path, code, out, f.calls)
		}
	}
	// A configured service with an empty root origin refuses every root origin.
	f := newFake()
	f.rootOrigin = ""
	x := newHarness(f)
	if code, _, out := x.do(t, req{"GET", "/internal/v1/signup/status?origin=" + root, "", auth(goodCred)}); code != 403 || out["error"] != "origin_rejected" {
		t.Errorf("empty root origin: %d %v", code, out)
	}
	if _, err := New(Config{}); err == nil {
		t.Error("New accepted a missing service")
	}
	if h, err := New(Config{Service: signup.NewService(signup.Config{})}); err != nil || h == nil || h.svc.Configured() {
		t.Errorf("New(unconfigured service) = %v, %v", h, err)
	}
}

// TestSignupRoutesPinned: exactly the seven method/path pairs of Plan r5 §4.2 are served; every other method on a
// served path is method_not_allowed and any other path at or below the root is not_found, each with one closed record.
func TestSignupRoutesPinned(t *testing.T) {
	want := []string{
		"GET /internal/v1/signup/status", "POST /internal/v1/signup/abandon", "POST /internal/v1/signup/challenge",
		"POST /internal/v1/signup/complete", "POST /internal/v1/signup/organizations", "POST /internal/v1/signup/prove",
		"POST /internal/v1/signup/slug",
	}
	got := []string{RouteStatus, RouteSlug, RouteOrganizations, RouteChallenge, RouteProve, RouteComplete, RouteAbandon}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Fatalf("routes %v, want %v", got, want)
	}
	if len(Paths) != len(want) {
		t.Fatalf("paths %v", Paths)
	}
	for _, r := range want {
		method, path, _ := strings.Cut(r, " ")
		if !slices.Contains(Paths, path) {
			t.Errorf("%s not in Paths", path)
		}
		for _, m := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
			if m == method {
				continue
			}
			x := newHarness(newFake())
			code, _, out := x.do(t, req{m, path, "", auth(goodCred)})
			recs := x.records(t)
			if code != 405 || out["error"] != "method_not_allowed" || len(x.f.calls) != 0 || len(recs) != 1 || recs[0]["operation"] != "method" {
				t.Errorf("%s %s: %d %v %v", m, path, code, out, recs)
			}
		}
	}
	for _, p := range []string{"/internal/v1/signup", "/internal/v1/signup/", "/internal/v1/signup/sessions", "/internal/v1/signup/status/x"} {
		x := newHarness(newFake())
		code, _, out := x.do(t, req{"GET", p, "", nil})
		recs := x.records(t)
		if code != 404 || out["error"] != "not_found" || len(recs) != 1 || recs[0]["operation"] != "path" {
			t.Errorf("%s: %d %v %v", p, code, out, recs)
		}
	}
	for code, st := range status {
		if st < 400 || st > 503 {
			t.Errorf("%s: status %d", code, st)
		}
	}
	if len(status) != 13 {
		t.Errorf("%d refusal codes, want 13", len(status))
	}
}

// TestSignupCredentialBoundary (SUP-3, stack-free part): the signup credential is not a session and a session is not a
// signup credential. identity's session route answers a Signup scheme as no credential (401 before any statement); the
// signup routes answer a Session scheme as no credential. (The enrollment adapter and MCP accept only their own
// "Enrollment " and bearer schemes; their refusal of a Signup credential is proven in their integration suites.)
func TestSignupCredentialBoundary(t *testing.T) {
	auth := identity.NewHandler(identity.NewService(nil, identity.Config{RootOrigin: root}))
	for _, v := range []string{"Signup " + goodCred, "signup " + goodCred, "Bearer " + goodCred} {
		hr := httptest.NewRequest("GET", "/internal/v1/auth/session?origin="+root, nil)
		hr.Header.Set("Authorization", v)
		w := httptest.NewRecorder()
		auth.Session(w, hr)
		if w.Code != 401 || !strings.Contains(w.Body.String(), `"unauthenticated"`) {
			t.Errorf("identity session with %q: %d %s", strings.Fields(v)[0], w.Code, w.Body.String())
		}
	}
	x := newHarness(newFake())
	for _, v := range []string{"Session " + goodCred, "signup " + goodCred, "Signup", "Signup  " + goodCred, "Signup " + goodCred + " x"} {
		code, _, out := x.do(t, req{"GET", "/internal/v1/signup/status?origin=" + root, "", map[string][]string{"Authorization": {v}}})
		if code != 401 || out["error"] != "unauthenticated" || slices.Contains(x.f.calls, "resolve") {
			t.Errorf("signup status with %q: %d %v %v", v, code, out, x.f.calls)
		}
	}
}

// TestSignupRateLimited (SUP-9, §4.5 (iii)-(iv)): the per-credential bucket (burst 10) covers status, slug, create,
// prove and abandon together; the per-user create bucket (burst 5) is not replenished by fresh credentials of the same
// user; challenge and complete each hold a per-tenant-origin bucket (burst 20); exhaustion is 429 rate_limited before
// any statement.
func TestSignupRateLimited(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	clock := func() time.Time { return now }
	f := newFake()
	x := newHarness(f)
	x.h.own.now, x.h.user.now, x.h.challenge.now, x.h.complete.now = clock, clock, clock, clock
	rootReqs := []req{
		{"GET", "/internal/v1/signup/status?origin=" + root, "", auth(goodCred)},
		{"POST", "/internal/v1/signup/slug", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "Acme"}), auth(goodCred)},
		{"POST", "/internal/v1/signup/prove", js(map[string]any{"origin": root, "nonce": "N"}), auth(goodCred)},
		{"POST", "/internal/v1/signup/abandon", js(map[string]any{"origin": root, "csrf_token": csrfToken()}), auth(goodCred)},
		{"POST", "/internal/v1/signup/organizations", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "A", "slug": "a"}), auth(goodCred)},
	}
	for i := range credentialBurst {
		if code, _, out := x.do(t, rootReqs[i%len(rootReqs)]); code == 429 {
			t.Fatalf("request %d limited early: %v", i, out)
		}
	}
	for _, r := range rootReqs {
		f.calls = nil
		code, _, out := x.do(t, r)
		if code != 429 || out["error"] != "rate_limited" || len(f.calls) != 0 {
			t.Errorf("%s %s after the credential burst: %d %v calls %v", r.method, r.path, code, out, f.calls)
		}
	}
	if recs := x.records(t); recs[len(recs)-1]["reason"] != "rate_limited" {
		t.Errorf("last record %v", recs[len(recs)-1])
	}
	// Another credential keeps its own bucket.
	if code, _, _ := x.do(t, req{"GET", "/internal/v1/signup/status?origin=" + root, "", auth(otherCred)}); code != 200 {
		t.Errorf("another credential limited: %d", code)
	}
	now = now.Add(time.Duration(credentialBurst/credentialRate) * time.Second)
	if code, _, _ := x.do(t, rootReqs[0]); code != 200 {
		t.Errorf("credential bucket not refilled: %d", code)
	}

	// Per user: fresh credentials of one user (userOf maps every credential to user-1) cannot exceed the create burst.
	f = newFake()
	x = newHarness(f)
	x.h.own.now, x.h.user.now = clock, clock
	create := func(cred string) int {
		f.creds[cred] = f.creds[goodCred]
		code, _, _ := x.do(t, req{"POST", "/internal/v1/signup/organizations", js(map[string]any{"origin": root, "csrf_token": csrfToken(), "name": "A", "slug": "a"}), auth(cred)})
		return code
	}
	for i := range userBurst {
		if code := create("fresh-" + string(rune('a'+i))); code != 201 {
			t.Fatalf("create %d: %d", i, code)
		}
	}
	f.calls = nil
	if code := create("fresh-z"); code != 429 || slices.Contains(f.calls, "provision") {
		t.Errorf("a re-minted credential replenished the user bucket: %d %v", code, f.calls)
	}
	// Status of the same user stays available (the user bucket is create-only).
	if code, _, _ := x.do(t, req{"GET", "/internal/v1/signup/status?origin=" + root, "", auth("fresh-z")}); code != 200 {
		t.Errorf("status limited by the create bucket: %d", code)
	}
	now = now.Add(userEvery)
	if code := create("fresh-y"); code != 201 {
		t.Errorf("user bucket not refilled after %s: %d", userEvery, code)
	}

	// Per tenant origin, per hop.
	for _, hop := range []struct {
		path string
		body func(origin string) string
	}{
		{"/internal/v1/signup/challenge", func(o string) string { return js(map[string]any{"origin": o, "ref": "R"}) }},
		{"/internal/v1/signup/complete", func(o string) string { return js(map[string]any{"origin": o, "challenge": "C", "proof": "P"}) }},
	} {
		f := newFake()
		x := newHarness(f)
		x.h.challenge.now, x.h.complete.now = clock, clock
		for i := range hopBurst {
			if code, _, _ := x.do(t, req{"POST", hop.path, hop.body(tenant), nil}); code != 200 {
				t.Fatalf("%s %d: %d", hop.path, i, code)
			}
		}
		f.calls = nil
		if code, _, out := x.do(t, req{"POST", hop.path, hop.body(tenant), nil}); code != 429 || out["error"] != "rate_limited" || len(f.calls) != 0 {
			t.Errorf("%s after the burst: %d %v %v", hop.path, code, out, f.calls)
		}
		if code, _, _ := x.do(t, req{"POST", hop.path, hop.body("https://other.localhost:543"), nil}); code != 200 {
			t.Errorf("%s: another tenant origin limited: %d", hop.path, code)
		}
	}
	// The two hops do not share a bucket.
	f = newFake()
	x = newHarness(f)
	for range hopBurst {
		x.do(t, req{"POST", "/internal/v1/signup/challenge", js(map[string]any{"origin": tenant, "ref": "R"}), nil})
	}
	if code, _, _ := x.do(t, req{"POST", "/internal/v1/signup/complete", js(map[string]any{"origin": tenant, "challenge": "C", "proof": "P"}), nil}); code != 200 {
		t.Errorf("complete limited by the challenge bucket: %d", code)
	}
}

// TestSignupLimiterChurn: churning credential headers or tenant origins fills each map to its bound without restoring a
// depleted bucket, and a full map refuses a new key until buckets refill; the per-user bucket's slow refill reclaims
// only after its full-refill time.
func TestSignupLimiterChurn(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	for name, l := range map[string]*limiter{"credential": newLimiter(credentialRate, credentialBurst), "hop": newLimiter(hopRate, hopBurst),
		"user": newSlowLimiter(userEvery, userBurst)} {
		l.now = func() time.Time { return now }
		victim := keyOf(name, "victim")
		for l.allow(victim) {
		}
		for i := range maxLimiters - 1 {
			l.allow(keyOf(name, string(rune(0x10000+i))))
		}
		if l.allow(keyOf(name, "new")) {
			t.Errorf("%s: a new key was admitted at capacity with no refilled bucket", name)
		}
		if l.allow(victim) {
			t.Errorf("%s: churn restored the depleted bucket", name)
		}
		now = now.Add(l.full + time.Second)
		if !l.allow(keyOf(name, "new")) {
			t.Errorf("%s: a new key was refused after every bucket refilled", name)
		}
	}
	if l := newSlowLimiter(userEvery, userBurst); l.full != 50*time.Second || l.burst != userBurst {
		t.Errorf("slow limiter full %s burst %d", l.full, l.burst)
	}
	if keyOf("credential", goodCred) == keyOf("user", goodCred) || strings.Contains(keyOf("credential", goodCred), goodCred) {
		t.Error("limiter keys are not scoped digests")
	}
}

// decl returns the source text of the named top-level type, function or method of the Go file at path.
func decl(t *testing.T, path, name string) string {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		var n ast.Node
		switch x := d.(type) {
		case *ast.FuncDecl:
			if x.Name.Name == name {
				n = x
			}
		case *ast.GenDecl:
			for _, s := range x.Specs {
				if ts, ok := s.(*ast.TypeSpec); ok && ts.Name.Name == name {
					n = x
				}
			}
		}
		if n != nil {
			return string(src[fset.Position(n.Pos()).Offset:fset.Position(n.End()).Offset])
		}
	}
	t.Fatalf("%s: %s not declared", path, name)
	return ""
}

// TestSignupCopyIsExact (Plan r5 §4.5): the signup buckets use a byte-pinned copy of the enrollment/api limiter, not a
// shared helper; the type, its constructor, allow and reclaim are identical and the key bound is the same.
func TestSignupCopyIsExact(t *testing.T) {
	for _, name := range []string{"limiter", "newLimiter", "allow", "reclaim"} {
		if a, b := decl(t, "../../enrollment/api/http.go", name), decl(t, "limiter_copy.go", name); a != b {
			t.Errorf("%s differs from enrollment/api:\n--- enrollment\n%s\n--- signup/api\n%s", name, a, b)
		}
	}
	if maxLimiters != 10000 {
		t.Errorf("maxLimiters = %d, want 10000", maxLimiters)
	}
}

// equalJSON compares a decoded answer with the expected map (numbers as float64).
func equalJSON(got, want map[string]any) bool {
	a, _ := json.Marshal(got)
	b, _ := json.Marshal(want)
	return bytes.Equal(a, b)
}
