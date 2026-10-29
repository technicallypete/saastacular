package logrecord

import (
	"encoding/json"
	"strings"
	"testing"
)

func load(t *testing.T) *Catalogue {
	t.Helper()
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return c
}

const (
	traceA  = "4bf92f3577b34da6a3ce929d0e0e4736"
	traceB  = "0af7651916cd43dd8448eb211c80319c"
	parentX = "b7ad6b7169203331"
	spanWeb = "00f067aa0ba902b7"
	spanCor = "1a2b3c4d5e6f7081"
	spanInt = "9f8e7d6c5b4a3921"
)

// completion returns a valid completion record for svc with the given overrides; a nil
// override value deletes the key.
func completion(svc string, over map[string]any) map[string]any {
	m := map[string]any{
		"time": "2026-10-06T12:00:00.123Z", "level": "info", "service": svc, "event": CompletionEvent,
		"method": "GET", "status": 200, "duration_ms": 3, "request_id": "req-chain-0001",
		"request_id_source": "inbound", "trace_id": traceA, "trace_source": "inbound",
		"span_id": spanWeb, "parent_span_id": parentX, "trace_flags": "01", "outcome": "completed",
	}
	switch svc {
	case "web":
		m["route"], m["core_probe"] = "/api/ready", "ok"
	case "core":
		m["route"], m["span_id"], m["parent_span_id"] = "/readyz", spanCor, spanWeb
	case "intelligence":
		m["route"], m["span_id"], m["parent_span_id"], m["logger"] = "/healthz", spanInt, spanCor, "intelligence.http"
	}
	for k, v := range over {
		if v == nil {
			delete(m, k)
		} else {
			m[k] = v
		}
	}
	return m
}

func line(t *testing.T, m map[string]any) string {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestCatalogueMatchesTheBrief(t *testing.T) {
	c := load(t)
	want := map[string][]string{
		"core": {
			// G004 Plan r5 §4.7: the five identity events. G005 Plan r4 §4.6: the one tenancy event. G007 Plan r1 §3.6: the one
			// access event. G009 Plan r1 §4.8: the one storage event. G011 Plan r2 §6: the two audit diagnostics.
			// G016 (scope amendment 04): auth.relay (identity/relay.go) and domain.request (domains/api/http.go).
			"access.decision", "audit.gap", "audit.record_failed", "auth.callback.completed", "auth.login.started", "auth.provider.revocation", "auth.relay", "auth.session.revoked",
			// G017 Plan r2 §5.3: branding.public and branding.request (branding/api/http.go).
			"branding.public", "branding.request",
			// G014 Plan r8 §6.2: the three MCP operational events (mcp.*).
			"core listening", "core.exit",
			// G071 Plan r4 S14/S19: the default-tenant startup selection record (S3).
			"default_tenant",
			"domain.request",
			// G058 Plan r0 §5.2 (HTTP slice): enrollment.request (enrollment/api/http.go).
			"enrollment.request",
			"http.request.completed", "http.server.error", "identity.namespace_mismatch",
			// G020 Plan r2 §4.2: labels.request (labels/api/http.go).
			"labels.request",
			"mcp.connection", "mcp.request", "mcp.setup",
			// G019 Plan r1 §5.3: membership.request (membership/api/http.go).
			"membership.request",
			// G018 Plan r5 §5.3: onboarding.request (onboarding/api/http.go).
			"onboarding.request",
			// G071 Plan r4 S14/S19: the invalid PLATFORM_ROOT_PAGE startup warning (S15).
			"platform_root_page",
			// G062 Plan r5 §2/SUP-13: signup.request (signup/log.go; logged by the signup/api HTTP adapter).
			"signup.request",
			"storage.operation", "tenancy.operation",
		},
		"web": {"http.request.completed"},
		"intelligence": {
			"INTELLIGENCE_SERVICE_TOKEN is not set; /internal/* refuses all requests", "configured", "execution",
			"http.request.completed", "intelligence.message", "intelligence.startup.failed", "library.log", "python.warning",
			"uvicorn.app_load_failed", "uvicorn.asgi_exception", "uvicorn.invalid_request", "uvicorn.listening",
			"uvicorn.other", "uvicorn.process_finished", "uvicorn.process_started", "uvicorn.server_error",
			"uvicorn.shutdown_complete", "uvicorn.shutdown_waiting", "uvicorn.shutting_down",
			"uvicorn.startup_complete", "uvicorn.startup_waiting",
		},
	}
	for svc, events := range want {
		if got := strings.Join(c.EventNames(svc), "|"); got != strings.Join(events, "|") {
			t.Errorf("%s events:\n got %s\nwant %s", svc, got, strings.Join(events, "|"))
		}
	}
}

func TestClassifyAccepts(t *testing.T) {
	c := load(t)
	for _, tc := range []struct{ svc, line, event string }{
		// Key order is irrelevant: core's slog emits time, level, event, service, ...
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"core listening","service":"core","port":8180,"app_env":"test"}`, "core listening"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"core.exit","service":"core","error_type":"config_invalid","config_keys":["DB_PASSWORD","REDIS_PASSWORD"]}`, "core.exit"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"core.exit","service":"core","error_type":"config_invalid","config_keys":[]}`, "core.exit"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"core.exit","service":"core","error_type":"config_invalid","config_keys":["IDENTITY_WORKOS_API_KEY_FILE","IDENTITY_WORKOS_ISSUER"]}`, "core.exit"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"core.exit","service":"core","error_type":"listen_failed"}`, "core.exit"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"auth.login.started","service":"core","outcome":"started","request_id":"req-12345678","trace_id":"` + traceA + `"}`, "auth.login.started"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"auth.callback.completed","service":"core","outcome":"signed_in"}`, "auth.callback.completed"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"auth.session.revoked","service":"core","outcome":"local_revoked","scope":"current"}`, "auth.session.revoked"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"auth.provider.revocation","service":"core","outcome":"provider_revoke_failed","scope":"provider"}`, "auth.provider.revocation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"identity.namespace_mismatch","service":"core"}`, "identity.namespace_mismatch"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"http.server.error","service":"core","error_type":"handler_panic"}`, "http.server.error"},
		// G005: one record per tenancy mutation; bounded codes only, correlation when present.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"tenancy.operation","service":"core","operation":"organization.create","outcome":"ok"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"tenancy.operation","service":"core","operation":"group.create","outcome":"ok","request_id":"req-12345678","trace_id":"` + traceA + `"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"tenancy.operation","service":"core","operation":"group.change_slug","outcome":"rejected","error_code":"slug_conflict"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"tenancy.operation","service":"core","operation":"membership.grant","scope_type":"group","outcome":"rejected","error_code":"membership_exists"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"tenancy.operation","service":"core","operation":"membership.remove","scope_type":"organization","outcome":"ok"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"tenancy.operation","service":"core","operation":"membership.change","scope_type":"organization","outcome":"failed","error_code":"unavailable","request_id":"req-12345678","trace_id":"` + traceA + `"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"tenancy.operation","service":"core","operation":"organization.rename","outcome":"failed","error_code":"internal"}`, "tenancy.operation"},
		// G006: BusinessProfile and association operations carry no scope_type.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"tenancy.operation","service":"core","operation":"business_profile.create","outcome":"ok"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"tenancy.operation","service":"core","operation":"business_profile.update","outcome":"rejected","error_code":"invalid_website"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"tenancy.operation","service":"core","operation":"association.link","outcome":"rejected","error_code":"association_exists","request_id":"req-12345678","trace_id":"` + traceA + `"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"tenancy.operation","service":"core","operation":"association.unlink","outcome":"failed","error_code":"unavailable"}`, "tenancy.operation"},
		// G007: one access.decision per call; the context operations carry no capability.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"access.begin_request","outcome":"established","reason":"context_established"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"access.resolve","outcome":"established","reason":"context_established","request_id":"req-12345678","trace_id":"` + traceA + `"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"access.begin_request","outcome":"deny","reason":"session_invalid"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"access.resolve","outcome":"deny","reason":"context_scope_mismatch"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"access.decision","service":"core","operation":"access.resolve","outcome":"failed","reason":"store_unavailable"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"organization.get","capability":"organization.read","outcome":"allow","reason":"allowed"}`, "access.decision"},
		// G019 (Result addendum A1-01 §3, D-A2): revoke/get carry the capability actually evaluated, or none for a
		// recipient-only read or a failure before that decision; the scoped operations carry their own capability.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"invitation.revoke","capability":"group.manage_members","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"invitation.get","capability":"organization.manage_members","outcome":"deny","reason":"role_not_permitted"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"invitation.get","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"invitation.revoke","outcome":"deny","reason":"target_not_found"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"membership.leave","capability":"organization.leave","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"ownership.transfer","capability":"group.manage_owners","outcome":"deny","reason":"write_conflict"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"retirement.eligibility","outcome":"allow","reason":"allowed"}`, "access.decision"},
		// G058: the enrollment-intent callback's outcomes (the enrollment answer and the T1/T2 refusals).
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"auth.callback.completed","service":"core","outcome":"enrollment_status"}`, "auth.callback.completed"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"auth.callback.completed","service":"core","outcome":"enrollment_closed"}`, "auth.callback.completed"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"auth.callback.completed","service":"core","outcome":"invitation_mismatch"}`, "auth.callback.completed"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"auth.callback.completed","service":"core","outcome":"recipient_proof_unavailable"}`, "auth.callback.completed"},
		// G062: the root signup callback (T1), the pinned tenant callback and its denials; LoginPinned's start refusals.
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"info","event":"auth.callback.completed","service":"core","outcome":"signup"}`, "auth.callback.completed"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"info","event":"auth.callback.completed","service":"core","outcome":"tenant_created"}`, "auth.callback.completed"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"info","event":"auth.callback.completed","service":"core","outcome":"signup_origin_rejected"}`, "auth.callback.completed"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"info","event":"auth.callback.completed","service":"core","outcome":"signup_user_mismatch"}`, "auth.callback.completed"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"info","event":"auth.login.started","service":"core","outcome":"continuation_invalid"}`, "auth.login.started"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"info","event":"auth.login.started","service":"core","outcome":"host_changed"}`, "auth.login.started"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"info","event":"auth.login.started","service":"core","outcome":"organization_not_member"}`, "auth.login.started"},
		// G058: the enrollment admission decision carries no capability (IDENTITY shape).
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"enrollment.admit","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.admit","outcome":"deny","reason":"write_conflict"}`, "access.decision"},
		// G058 Codex round-2 log decision: the admission refusals name their semantic reasons.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.admit","outcome":"deny","reason":"enrollment_closed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.admit","outcome":"deny","reason":"enrollment_restricted"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.admit","outcome":"deny","reason":"recipient_ineligible"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.admit","outcome":"deny","reason":"host_changed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.admit","outcome":"deny","reason":"session_invalid"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.admit","outcome":"deny","reason":"target_not_found"}`, "access.decision"},
		// G058 slice 1b: the guarded enrollment policy read and write; open on unresolved history is enrollment_not_ready.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"enrollment.policy.get","capability":"organization.read","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"enrollment.policy.update","capability":"organization.manage_policy","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.policy.update","capability":"organization.manage_policy","outcome":"deny","reason":"enrollment_not_ready"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.policy.update","capability":"organization.manage_policy","outcome":"deny","reason":"role_not_permitted"}`, "access.decision"},
		// G058 request slice: the request admission's approval_revoked; the applicant's own-credential operations carry
		// no capability (enrollment.request records the admission's applicable reasons); the three manager operations
		// carry organization.manage_members.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.admit","outcome":"deny","reason":"approval_revoked"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"enrollment.request","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.request","outcome":"deny","reason":"enrollment_closed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.request","outcome":"deny","reason":"write_conflict"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"access.decision","service":"core","operation":"enrollment.request","outcome":"failed","reason":"aborted"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"enrollment.status","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.status","outcome":"deny","reason":"session_invalid"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"enrollment.abandon","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"enrollment.requests.list","capability":"organization.manage_members","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"enrollment.requests.decide","capability":"organization.manage_members","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.requests.revoke","capability":"organization.manage_members","outcome":"deny","reason":"write_conflict"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"enrollment.requests.decide","capability":"organization.manage_members","outcome":"deny","reason":"role_not_permitted"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"organization.set_policy","capability":"organization.manage_policy","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"group.rename","capability":"group.update","outcome":"deny","reason":"role_not_permitted"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"business_profile.update","capability":"business_profile.profile.update","outcome":"deny","reason":"edit_policy_disabled","request_id":"req-12345678","trace_id":"` + traceA + `"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"access.decision","service":"core","operation":"business_profile.update","capability":"business_profile.profile.update","outcome":"failed","reason":"commit_outcome_unknown"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"access.decision","service":"core","operation":"group.change_slug","capability":"group.update","outcome":"failed","reason":"aborted"}`, "access.decision"},
		// G011: the guarded audit read, the issuance-record stage and the two audit diagnostics.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"audit.read","capability":"audit.read","outcome":"deny","reason":"role_not_permitted"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"storage.operation","service":"core","operation":"object.sign","outcome":"failed","reason":"store_unavailable","stage":"audit"}`, "storage.operation"},
		// G017: one record per branding route; closed operations and reasons, the hostname only as its tag.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"branding.request","service":"core","operation":"update","outcome":"ok","request_id":"req-12345678"}`, "branding.request"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"branding.request","service":"core","operation":"set_logo","outcome":"refused","reason":"payload_too_large"}`, "branding.request"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"branding.request","service":"core","operation":"set_logo","outcome":"failed","reason":"publication_outcome_unknown"}`, "branding.request"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"branding.public","service":"core","operation":"logo","outcome":"ok","host_tag":"0123456789ab"}`, "branding.public"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"branding.public","service":"core","operation":"projection","outcome":"refused","reason":"unknown_host"}`, "branding.public"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"branding.public","service":"core","operation":"logo","outcome":"failed","reason":"branding_unavailable","host_tag":"0123456789ab"}`, "branding.public"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"audit.record_failed","service":"core","action":"access.denied","request_id":"req-12345678","trace_id":"` + traceA + `"}`, "audit.record_failed"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"audit.record_failed","service":"core","action":"auth.sign_in"}`, "audit.record_failed"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"audit.gap","service":"core","operation":"logout_all","independent":"recorded"}`, "audit.gap"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"audit.gap","service":"core","operation":"disabled","independent":"not_attempted"}`, "audit.gap"},
		{"core", line(t, completion("core", map[string]any{"route": "/healthz"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "unmatched", "status": 404, "method": "OTHER"})), CompletionEvent},
		// G004 (checkpoint-C4 logging amendment): exactly the five core identity routes.
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/session"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/callback", "method": "POST", "status": 401})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/login", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/logout", "method": "POST", "status": 503, "level": "warn"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/provider-logout/return", "method": "POST"})), CompletionEvent},
		// G014: exactly the two internal MCP setup route paths (list and create share one path).
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/mcp/connections"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/mcp/connections/revoke", "method": "POST"})), CompletionEvent},
		// G016 (scope amendment 05): exactly the ten registered relay, host-resolution and domain route paths (list and
		// claim share /internal/v1/domains).
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/relay", "status": 303})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/hosts/resolve"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains", "method": "POST", "status": 201})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains/tls-ask", "status": 403})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains/rotate", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains/verify", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains/recheck", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains/remove", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains/retry-withdrawal", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/organization/slug", "method": "POST", "status": 409})), CompletionEvent},
		// G017 Plan r2 §5.3: exactly the six registered branding route paths (GET/POST share /branding and /branding/logo).
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding/public"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding/logo", "status": 404})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding/logo", "method": "POST", "status": 413})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding", "method": "POST", "status": 409})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding/reset", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding/logo/remove", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding/objects/delete", "method": "POST"})), CompletionEvent},
		// G018 Plan r5 §5.3: exactly the six registered onboarding route paths (eight method/path pairs).
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/groups"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/groups", "method": "POST", "status": 201})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/groups/slug-suggestion", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/groups/{id}", "status": 404})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/groups/{id}/business-profile", "method": "POST", "status": 409})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/business-profiles"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/business-profiles/{id}"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/business-profiles/{id}", "method": "PUT", "status": 403})), CompletionEvent},
		// G018: one onboarding.request per route; closed operations and reasons, the origin host only as its tag; the
		// access.decision record of each onboarding operation with its capability.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"onboarding.request","service":"core","operation":"create_group","outcome":"ok","host_tag":"0123456789ab","request_id":"req-12345678"}`, "onboarding.request"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"onboarding.request","service":"core","operation":"method","outcome":"refused","reason":"method_not_allowed"}`, "onboarding.request"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"onboarding.request","service":"core","operation":"complete_onboarding","outcome":"failed","reason":"onboarding_unavailable","host_tag":"0123456789ab"}`, "onboarding.request"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"group.suggest_slug","capability":"group.create","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"group.create","capability":"group.create","outcome":"deny","reason":"role_not_permitted"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"group.list","capability":"group.read","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"group.onboarding.get","capability":"group.read","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"group.onboarding.complete_new","capability":"group.complete_onboarding","outcome":"deny","reason":"write_conflict"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"group.onboarding.complete_existing","capability":"group.complete_onboarding","outcome":"allow","reason":"allowed"}`, "access.decision"},
		// G060 (r1 C2): the Business Profile listing is a metadata read guarded by business_profile.profile.read; an
		// override-based management allow carries reason allowed_override on its management capability only.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"business_profile.list","capability":"business_profile.profile.read","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-19T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"membership.change_role","capability":"group.manage_members","outcome":"allow","reason":"allowed_override"}`, "access.decision"},
		// G019 Plan r1 §5.3, r2 addendum §7: exactly the fifteen registered membership route paths (sixteen pairs).
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/members"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/members/self", "status": 404})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/members/{user_id}/role", "method": "POST", "status": 403})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/members/{user_id}/remove", "method": "POST", "status": 409})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/members/{user_id}/promote", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/members/{user_id}/demote", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/members/leave", "method": "POST", "status": 409})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/ownership/transfer", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/invitations"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/invitations", "method": "POST", "status": 201})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/invitations/mine"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/invitations/candidates"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/invitations/{id}", "status": 404})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/invitations/{id}/revoke", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/invitations/{id}/accept", "method": "POST", "status": 404})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/account/retirement-eligibility"})), CompletionEvent},
		// G060 (scope amendment 02): the one Group-authority route path; the page read (GET) and the write (POST, allowed
		// or denied) share it.
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/organization/group-authority"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/organization/group-authority", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/organization/group-authority", "method": "POST", "status": 403})), CompletionEvent},
		// G019: one membership.request per route; closed operations and reasons, the origin host only as its tag. The
		// trusted retirement's one tenancy.operation record (retirement.retire) and the owner bootstrap's refusal.
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"membership.request","service":"core","operation":"issue_invitation","outcome":"ok","host_tag":"0123456789ab","request_id":"req-12345678"}`, "membership.request"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"membership.request","service":"core","operation":"remove_member","outcome":"refused","reason":"conflict","host_tag":"0123456789ab"}`, "membership.request"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"membership.request","service":"core","operation":"path","outcome":"refused","reason":"not_found"}`, "membership.request"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"membership.request","service":"core","operation":"accept_invitation","outcome":"failed","reason":"membership_unavailable","host_tag":"0123456789ab"}`, "membership.request"},
		// G020 Plan r2 §4.2: one labels.request per route (closed operations and reasons, the origin host only as its tag),
		// the six labels route paths (seven pairs), and the seven label operations' access.decision records with the
		// three label capabilities (the capability actually evaluated, absent for a call that ended before that decision).
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"info","event":"labels.request","service":"core","operation":"create_label","outcome":"ok","host_tag":"0123456789ab"}`, "labels.request"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"labels.request","service":"core","operation":"assign_label","outcome":"refused","reason":"conflict","host_tag":"0123456789ab"}`, "labels.request"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"labels.request","service":"core","operation":"list_group_labels","outcome":"refused","reason":"not_found","host_tag":"0123456789ab"}`, "labels.request"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"labels.request","service":"core","operation":"path","outcome":"refused","reason":"not_found"}`, "labels.request"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"labels.request","service":"core","operation":"method","outcome":"refused","reason":"method_not_allowed"}`, "labels.request"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"error","event":"labels.request","service":"core","operation":"update_label","outcome":"failed","reason":"labels_unavailable","host_tag":"0123456789ab"}`, "labels.request"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"info","event":"labels.request","service":"core","operation":"list_labels","outcome":"ok","host_tag":"0123456789ab"}`, "labels.request"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"labels.request","service":"core","operation":"delete_label","outcome":"refused","reason":"forbidden","host_tag":"0123456789ab"}`, "labels.request"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"labels.request","service":"core","operation":"unassign_label","outcome":"refused","reason":"invalid_request","host_tag":"0123456789ab"}`, "labels.request"},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/labels"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/labels", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/labels/{id}/update", "method": "POST", "status": 409})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/labels/{id}/delete", "method": "POST", "status": 404})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/groups/{id}/labels", "status": 403})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/groups/{id}/labels/assign", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/groups/{id}/labels/unassign", "method": "POST", "status": 409})), CompletionEvent},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"label.definition.list","capability":"label.organization.manage","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"label.definition.list","capability":"label.group.read","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"label.definition.create","capability":"label.group.manage","outcome":"deny","reason":"role_not_permitted"}`, "access.decision"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"label.definition.update","capability":"label.group.manage","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"label.definition.update","outcome":"deny","reason":"target_not_found"}`, "access.decision"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"label.definition.delete","capability":"label.organization.manage","outcome":"deny","reason":"input_invalid"}`, "access.decision"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"info","event":"access.decision","service":"core","operation":"label.assignment.list","capability":"label.organization.manage","outcome":"allow","reason":"allowed"}`, "access.decision"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"label.assign","capability":"label.organization.manage","outcome":"deny","reason":"write_conflict"}`, "access.decision"},
		// G020: a group-target assignment decides group.read on the exact Group after label.organization.manage; its denial (a
		// Group outside the selected context) records group.read.
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"warn","event":"access.decision","service":"core","operation":"label.assign","capability":"group.read","outcome":"deny","reason":"target_outside_context"}`, "access.decision"},
		{"core", `{"time":"2026-10-22T12:00:00.000Z","level":"error","event":"access.decision","service":"core","operation":"label.unassign","capability":"label.organization.manage","outcome":"failed","reason":"aborted"}`, "access.decision"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"tenancy.operation","service":"core","operation":"retirement.retire","outcome":"ok"}`, "tenancy.operation"},
		// G058 HTTP slice: one enrollment.request per route (the origin host only as its tag) and the enrollment routes'
		// completion records.
		{"core", `{"time":"2026-10-18T12:00:00.000Z","level":"info","event":"enrollment.request","service":"core","operation":"status","outcome":"ok","host_tag":"0123456789ab","request_id":"req-12345678"}`, "enrollment.request"},
		{"core", `{"time":"2026-10-18T12:00:00.000Z","level":"warn","event":"enrollment.request","service":"core","operation":"admit","outcome":"refused","reason":"enrollment_closed","host_tag":"0123456789ab"}`, "enrollment.request"},
		{"core", `{"time":"2026-10-18T12:00:00.000Z","level":"warn","event":"enrollment.request","service":"core","operation":"public","outcome":"refused","reason":"rate_limited"}`, "enrollment.request"},
		{"core", `{"time":"2026-10-18T12:00:00.000Z","level":"warn","event":"enrollment.request","service":"core","operation":"path","outcome":"refused","reason":"not_found"}`, "enrollment.request"},
		{"core", `{"time":"2026-10-18T12:00:00.000Z","level":"error","event":"enrollment.request","service":"core","operation":"request","outcome":"failed","reason":"enrollment_unavailable","host_tag":"0123456789ab"}`, "enrollment.request"},
		// G062: one signup.request per signup operation (no host tag, credential, secret, name, slug or identifier).
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"info","event":"signup.request","service":"core","operation":"create","outcome":"ok","request_id":"req-12345678","trace_id":"` + traceA + `"}`, "signup.request"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"warn","event":"signup.request","service":"core","operation":"create","outcome":"refused","reason":"slug_taken"}`, "signup.request"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"warn","event":"signup.request","service":"core","operation":"prove","outcome":"refused","reason":"signup_context_mismatch"}`, "signup.request"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"warn","event":"signup.request","service":"core","operation":"challenge","outcome":"refused","reason":"rate_limited"}`, "signup.request"},
		{"core", `{"time":"2026-10-20T12:00:00.000Z","level":"error","event":"signup.request","service":"core","operation":"complete","outcome":"failed","reason":"signup_unavailable"}`, "signup.request"},
		// G062 (Plan r5 §4.2): exactly the seven registered signup route paths (core/internal/signup/api Paths).
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/status"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/slug", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/organizations", "method": "POST", "status": 201})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/organizations", "method": "POST", "status": 409})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/challenge", "method": "POST", "status": 401})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/prove", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/complete", "method": "POST", "status": 503, "level": "warn"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/abandon", "method": "POST"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/status", "method": "PUT", "status": 405})), CompletionEvent},
		// G071 (Plan r4 S14/S19): the default-tenant selection records (a canonical slug and the pinned UUID only when
		// active; a slug only for unknown_slug; nothing configured for invalid_slug), the invalid root-page warning and
		// the lookup-outage exit category.
		{"core", `{"time":"2026-10-27T12:00:00.000Z","level":"info","event":"default_tenant","service":"core","outcome":"active","slug":"alpha-agency","organization_id":"0f71a000-0000-4000-8000-0000000000a1"}`, "default_tenant"},
		{"core", `{"time":"2026-10-27T12:00:00.000Z","level":"warn","event":"default_tenant","service":"core","outcome":"inactive","reason":"unknown_slug","slug":"no-such-agency"}`, "default_tenant"},
		{"core", `{"time":"2026-10-27T12:00:00.000Z","level":"warn","event":"default_tenant","service":"core","outcome":"inactive","reason":"invalid_slug"}`, "default_tenant"},
		{"core", `{"time":"2026-10-27T12:00:00.000Z","level":"warn","event":"platform_root_page","service":"core","outcome":"invalid_value","fallback":"signup_redirect"}`, "platform_root_page"},
		{"core", `{"time":"2026-10-27T12:00:00.000Z","level":"error","event":"core.exit","service":"core","error_type":"default_tenant_unavailable"}`, "core.exit"},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/enrollment/status"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"route": "/internal/v1/enrollment/requests/{id}/revoke", "method": "POST"})), CompletionEvent},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"tenancy.operation","service":"core","operation":"retirement.retire","outcome":"rejected","error_code":"retirement_blocked"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"tenancy.operation","service":"core","operation":"retirement.retire","outcome":"rejected","error_code":"last_owner"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"tenancy.operation","service":"core","operation":"retirement.retire","outcome":"failed","error_code":"unavailable"}`, "tenancy.operation"},
		{"core", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","event":"tenancy.operation","service":"core","operation":"organization.create","outcome":"rejected","error_code":"recipient_ineligible"}`, "tenancy.operation"},
		{"core", line(t, completion("core", map[string]any{"status": 0, "outcome": "cancelled"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"status": 0, "outcome": "failed", "error_type": "panic", "level": "error"})), CompletionEvent},
		{"core", line(t, completion("core", map[string]any{"status": 503, "level": "warn", "trace_source": "generated", "parent_span_id": nil, "trace_flags": "00", "request_id_source": "generated"})), CompletionEvent},
		{"web", "▲ Next.js 16.3.8", ""},
		{"web", "- Local:         http://localhost:3000", ""},
		{"web", "- Network:       http://0.0.0.0:3000", ""},
		{"web", "✓ Ready in 512ms", ""},
		{"web", "✓ Ready in 1.2s", ""},
		{"web", "✓ Ready in 0ms\r", ""},
		{"web", "✓ Running next.config took 12ms", ""},
		{"web", "✓ Running next.config took 1.05s", ""},
		{"web", line(t, completion("web", nil)), CompletionEvent},
		{"web", line(t, completion("web", map[string]any{"route": "/api/health", "core_probe": nil})), CompletionEvent},
		{"web", line(t, completion("web", map[string]any{"outcome": "failed", "error_type": "TypeError", "level": "error", "status": 500, "core_probe": nil})), CompletionEvent},
		{"web", line(t, completion("web", map[string]any{"outcome": "cancelled", "status": 0, "core_probe": nil})), CompletionEvent},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"configured","logger":"intelligence","capabilities":["chat"]}`, "configured"},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","service":"intelligence","event":"INTELLIGENCE_SERVICE_TOKEN is not set; /internal/* refuses all requests","logger":"intelligence"}`, "INTELLIGENCE_SERVICE_TOKEN is not set; /internal/* refuses all requests"},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"execution","logger":"intelligence.observability","request_id":"req-12345678","trace_id":"` + traceA + `","capability":"chat","outcome":"ok","usage":{"configured_provider":"openai","configured_model":"m","gateway":"openrouter","actual_provider":"unknown","actual_model":"m","input_tokens":"[REDACTED]","output_tokens":5,"latency_ms":12,"estimated_cost":null}}`, "execution"},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"execution","logger":"intelligence.observability","request_id":"req-12345678","trace_id":"` + traceA + `","capability":"chat","outcome":"capability_not_configured"}`, "execution"},
		{"intelligence", line(t, completion("intelligence", nil)), CompletionEvent},
		{"intelligence", line(t, completion("intelligence", map[string]any{"outcome": "failed", "error_type": "RuntimeError", "level": "error", "status": 500})), CompletionEvent},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"error","service":"intelligence","event":"intelligence.startup.failed","logger":"intelligence","error_type":"ConfigError"}`, "intelligence.startup.failed"},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"uvicorn.startup_complete","logger":"uvicorn.error"}`, "uvicorn.startup_complete"},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"error","service":"intelligence","event":"uvicorn.asgi_exception","logger":"uvicorn.error","error_type":"ValueError"}`, "uvicorn.asgi_exception"},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","service":"intelligence","event":"library.log","logger":"httpx"}`, "library.log"},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"intelligence.message","logger":"intelligence"}`, "intelligence.message"},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","service":"intelligence","event":"intelligence.message","logger":"intelligence.observability"}`, "intelligence.message"},
		{"intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","service":"intelligence","event":"python.warning","logger":"py.warnings","category":"DeprecationWarning"}`, "python.warning"},
	} {
		l := c.Classify(tc.svc, []byte(tc.line))
		if l.Class != "" || l.Event != tc.event {
			t.Errorf("%s %q: class %q event %q", tc.svc, tc.line, l.Class, l.Event)
		}
	}
}

func TestClassifyRejects(t *testing.T) {
	c := load(t)
	envelope := `"time":"2026-10-06T12:00:00.000Z","service":"core"`
	for _, tc := range []struct{ name, svc, line, class string }{
		{"planted non-JSON line", "core", "listen tcp: lookup bad.invalid: no such host", ClassNonJSON},
		{"empty line outside web", "core", "", ClassNonJSON},
		{"panic stack text", "core", "goroutine 1 [running]:", ClassNonJSON},
		{"banner with trailing space", "web", "✓ Ready in 5ms ", ClassNonJSON},
		{"banner with other version shape", "web", "▲ Next.js 16.3", ClassNonJSON},
		{"Plan-anticipated indented banner", "web", " ▲ Next.js 16.3.8", ClassNonJSON},
		{"Plan-anticipated starting line", "web", " ✓ Starting...", ClassNonJSON},
		{"Starting line", "web", "✓ Starting...", ClassNonJSON},
		{"Local with eight spaces", "web", "- Local:        http://localhost:3000", ClassNonJSON},
		{"Network with spaces in URL", "web", "- Network:       http://0.0.0.0:3000 x", ClassNonJSON},
		{"two carriage returns", "web", "✓ Ready in 5ms\r\r", ClassNonJSON},
		{"empty line on web", "web", "", ClassNonJSON},
		{"banner on core", "core", "✓ Ready in 5ms", ClassNonJSON},
		{"next error text", "web", "Error: boom", ClassNonJSON},
		{"next error stack", "web", "    at handler (/app/.next/server/app/api/route.js:1:1)", ClassNonJSON},
		{"top-level array", "core", `[1]`, ClassNonJSON},
		{"trailing space after JSON", "core", `{"a":1} `, ClassNonJSON},
		{"NaN", "core", `{"time":NaN}`, ClassNonJSON},
		{"truncated JSON", "core", `{"time":"2026-10-06T12:00:00.000Z"`, ClassNonJSON},
		{"duplicate key", "core", `{` + envelope + `,"level":"info","level":"info","event":"core listening","port":1,"app_env":"test"}`, ClassDuplicateKey},
		{"uncatalogued event", "core", `{` + envelope + `,"level":"info","event":"server started"}`, ClassUncatalogued},
		{"uncatalogued identity-like event", "core", `{` + envelope + `,"level":"info","event":"auth.session.created","outcome":"started"}`, ClassUncatalogued},
		{"identity event outside core", "web", `{"time":"2026-10-06T12:00:00.000Z","service":"web","level":"info","event":"auth.login.started","outcome":"started"}`, ClassUncatalogued},
		{"event carries an exception message", "core", `{` + envelope + `,"level":"error","event":"listen tcp: lookup bad-x.invalid: no such host"}`, ClassUncatalogued},
		{"intelligence event carries an exception message", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"error","service":"intelligence","event":"ConfigError: APP_ENV is required","logger":"intelligence"}`, ClassUncatalogued},
		{"free-text error_type", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"listen tcp: lookup x"}`, ClassErrorType},
		{"uncategorized error_type", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"fatal"}`, ClassErrorType},
		{"core completion error_type other than panic", "core", line(t, completion("core", map[string]any{"outcome": "failed", "error_type": "Error", "level": "error"})), ClassErrorType},
		{"web error_type outside the closed set", "web", line(t, completion("web", map[string]any{"outcome": "failed", "error_type": "MyCustomError", "level": "error", "core_probe": nil})), ClassErrorType},
		{"web error_type free text", "web", line(t, completion("web", map[string]any{"outcome": "failed", "error_type": "boom: x", "level": "error", "core_probe": nil})), ClassErrorType},
		{"intelligence free-text error_type", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"error","service":"intelligence","event":"intelligence.startup.failed","logger":"intelligence","error_type":"APP_ENV is required"}`, ClassErrorType},
		{"core listening with address", "core", `{` + envelope + `,"level":"info","event":"core listening","port":8180,"app_env":"test","addr":":8180"}`, ClassEventKeys},
		{"core listening without port", "core", `{` + envelope + `,"level":"info","event":"core listening","app_env":"test"}`, ClassEventKeys},
		{"core listening wrong level", "core", `{` + envelope + `,"level":"warn","event":"core listening","port":8180,"app_env":"test"}`, ClassEventKeys},
		{"core listening string port", "core", `{` + envelope + `,"level":"info","event":"core listening","port":"8180","app_env":"test"}`, ClassEventKeys},
		{"config_keys on listen_failed", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"listen_failed","config_keys":[]}`, ClassEventKeys},
		{"config_invalid without config_keys", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"config_invalid"}`, ClassEventKeys},
		{"config_keys not allow-listed", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"config_invalid","config_keys":["SOMETHING_ELSE"]}`, ClassEventKeys},
		{"config_keys identity near miss", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"config_invalid","config_keys":["IDENTITY_WORKOS_API_KEY_FILES"]}`, ClassEventKeys},
		{"config_keys unknown identity name", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"config_invalid","config_keys":["IDENTITY_UNKNOWN"]}`, ClassEventKeys},
		{"exit message key", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"unknown","error":"boom"}`, ClassEventKeys},
		// G005: the tenancy operation record is closed and its outcome, level, code and scope are coupled.
		{"tenancy uncatalogued name", "core", `{` + envelope + `,"level":"info","event":"tenancy.op","operation":"group.create","outcome":"ok"}`, ClassUncatalogued},
		{"tenancy event outside core", "web", `{"time":"2026-10-06T12:00:00.000Z","service":"web","level":"info","event":"tenancy.operation","operation":"group.create","outcome":"ok"}`, ClassUncatalogued},
		{"tenancy rejected without error_code", "core", `{` + envelope + `,"level":"warn","event":"tenancy.operation","operation":"group.create","outcome":"rejected"}`, ClassEventKeys},
		{"tenancy failed without error_code", "core", `{` + envelope + `,"level":"error","event":"tenancy.operation","operation":"group.create","outcome":"failed"}`, ClassEventKeys},
		{"tenancy ok with error_code", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"group.create","outcome":"ok","error_code":"internal"}`, ClassEventKeys},
		{"tenancy extra key", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"group.create","outcome":"ok","slug":"acme"}`, ClassEventKeys},
		{"tenancy identifier key", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"group.create","outcome":"ok","organization_id":"00000000-0000-0000-0000-000000000001"}`, ClassEventKeys},
		{"tenancy unknown operation", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"organization.delete","outcome":"ok"}`, ClassEventKeys},
		{"tenancy unknown outcome", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"group.create","outcome":"done"}`, ClassEventKeys},
		{"tenancy unknown error_code", "core", `{` + envelope + `,"level":"warn","event":"tenancy.operation","operation":"group.create","outcome":"rejected","error_code":"duplicate key value"}`, ClassEventKeys},
		{"tenancy rejected at info", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"group.create","outcome":"rejected","error_code":"slug_conflict"}`, ClassEventKeys},
		{"tenancy failed with a rejection code", "core", `{` + envelope + `,"level":"error","event":"tenancy.operation","operation":"group.create","outcome":"failed","error_code":"slug_conflict"}`, ClassEventKeys},
		{"tenancy rejected with a failure code", "core", `{` + envelope + `,"level":"warn","event":"tenancy.operation","operation":"group.create","outcome":"rejected","error_code":"unavailable"}`, ClassEventKeys},
		{"tenancy membership without scope_type", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"membership.grant","outcome":"ok"}`, ClassEventKeys},
		{"tenancy scope_type on a Group operation", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"group.rename","scope_type":"group","outcome":"ok"}`, ClassEventKeys},
		{"tenancy BusinessProfile scope", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"membership.grant","scope_type":"business_profile","outcome":"ok"}`, ClassEventKeys},
		{"tenancy scope_type on a profile operation", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"association.link","scope_type":"group","outcome":"ok"}`, ClassEventKeys},
		{"tenancy unknown profile operation", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"business_profile.delete","outcome":"ok"}`, ClassEventKeys},
		{"tenancy unknown profile code", "core", `{` + envelope + `,"level":"warn","event":"tenancy.operation","operation":"association.link","outcome":"rejected","error_code":"association_replaced"}`, ClassEventKeys},
		{"tenancy request_id without trace_id", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"group.create","outcome":"ok","request_id":"req-12345678"}`, ClassEventKeys},
		{"tenancy invalid request_id", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"group.create","outcome":"ok","request_id":"short","trace_id":"` + traceA + `"}`, ClassEventKeys},
		// G007: the access.decision record is closed; context operations never carry a capability or a policy allow.
		{"access context operation with a capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"access.resolve","capability":"organization.read","outcome":"established","reason":"context_established"}`, ClassEventKeys},
		{"access context operation allowed", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"access.begin_request","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access context operation with reason allowed", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"access.resolve","outcome":"established","reason":"allowed"}`, ClassEventKeys},
		// G018: each onboarding operation carries exactly its capability; onboarding.request is closed.
		{"onboarding create with group.read", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"group.create","capability":"group.read","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"onboarding list with group.create", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"business_profile.list","capability":"group.create","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		// G060 (D-A3-1): the pre-G060 selection capability is no longer the listing's capability; allowed_override stays
		// on allow decisions of the management capabilities only.
		{"profile list with the obsolete selection capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"business_profile.list","capability":"business_profile.select_for_onboarding","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"override reason on the metadata listing", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"business_profile.list","capability":"business_profile.profile.read","outcome":"allow","reason":"allowed_override"}`, ClassEventKeys},
		{"override reason on a denial", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"membership.change_role","capability":"group.manage_members","outcome":"deny","reason":"allowed_override"}`, ClassEventKeys},
		{"override reason on owner administration", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"ownership.promote","capability":"group.manage_owners","outcome":"allow","reason":"allowed_override"}`, ClassEventKeys},
		{"onboarding request with an identifier", "core", `{` + envelope + `,"level":"info","event":"onboarding.request","operation":"get_group","outcome":"ok","group_id":"0f18a000-0000-4000-8000-0000000000a1"}`, ClassEventKeys},
		{"onboarding request refused as unavailable", "core", `{` + envelope + `,"level":"warn","event":"onboarding.request","operation":"create_group","outcome":"refused","reason":"onboarding_unavailable"}`, ClassEventKeys},
		{"onboarding request unknown operation", "core", `{` + envelope + `,"level":"info","event":"onboarding.request","operation":"delete_group","outcome":"ok"}`, ClassEventKeys},
		// G019: membership.request is closed (no identifier, token or address; no foreign unavailable code); the retirement
		// codes and recipient_ineligible stay on their own tenancy operations; retirement.retire carries no scope_type.
		{"membership request with a token", "core", `{` + envelope + `,"level":"info","event":"membership.request","operation":"issue_invitation","outcome":"ok","token":"abc"}`, ClassEventKeys},
		{"membership request with an address", "core", `{` + envelope + `,"level":"warn","event":"membership.request","operation":"issue_invitation","outcome":"refused","reason":"conflict","intended_address":"a@example.invalid"}`, ClassEventKeys},
		{"membership request refused as onboarding unavailable", "core", `{` + envelope + `,"level":"warn","event":"membership.request","operation":"list_members","outcome":"refused","reason":"onboarding_unavailable"}`, ClassEventKeys},
		{"membership request unknown operation", "core", `{` + envelope + `,"level":"info","event":"membership.request","operation":"retire_user","outcome":"ok"}`, ClassEventKeys},
		// G020: labels.request is closed (no name, label or Group id, token or payload; no foreign unavailable code; no
		// unknown operation); the label operations carry only a label capability; no labels route outside the six paths.
		{"labels request with a name", "core", `{` + envelope + `,"level":"info","event":"labels.request","operation":"create_label","outcome":"ok","name":"Priority"}`, ClassEventKeys},
		{"labels request with a label id", "core", `{` + envelope + `,"level":"warn","event":"labels.request","operation":"assign_label","outcome":"refused","reason":"conflict","label_id":"0f20a000-0000-4000-8000-00000000e001"}`, ClassEventKeys},
		{"labels request with a Group id", "core", `{` + envelope + `,"level":"warn","event":"labels.request","operation":"list_group_labels","outcome":"refused","reason":"not_found","group_id":"0f20a000-0000-4000-8000-00000000b001"}`, ClassEventKeys},
		{"labels request with a target id", "core", `{` + envelope + `,"level":"info","event":"labels.request","operation":"assign_label","outcome":"ok","target_id":"0f20a000-0000-4000-8000-00000000b001"}`, ClassEventKeys},
		{"labels request refused as membership unavailable", "core", `{` + envelope + `,"level":"warn","event":"labels.request","operation":"list_labels","outcome":"refused","reason":"membership_unavailable"}`, ClassEventKeys},
		{"labels request refused as labels unavailable", "core", `{` + envelope + `,"level":"warn","event":"labels.request","operation":"list_labels","outcome":"refused","reason":"labels_unavailable"}`, ClassEventKeys},
		{"labels request failed with a refusal reason", "core", `{` + envelope + `,"level":"error","event":"labels.request","operation":"delete_label","outcome":"failed","reason":"forbidden"}`, ClassEventKeys},
		{"labels request unknown operation", "core", `{` + envelope + `,"level":"info","event":"labels.request","operation":"assign_fixture_label","outcome":"ok"}`, ClassEventKeys},
		{"labels request ok with a reason", "core", `{` + envelope + `,"level":"info","event":"labels.request","operation":"create_label","outcome":"ok","reason":"conflict"}`, ClassEventKeys},
		{"access label operation with a membership capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"label.assign","capability":"organization.manage_members","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access label list with a Group read capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"label.definition.list","capability":"group.read","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access label operation allowed by override", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"label.definition.create","capability":"label.group.manage","outcome":"allow","reason":"allowed_override"}`, ClassEventKeys},
		{"access guarded operation with a label capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"group.get","capability":"label.group.read","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access unknown label operation", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"label.fixture.assign","capability":"label.group.manage","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access unknown label capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"label.assign","capability":"label.group.assign","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"labels sub-path route", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/labels/{id}"})), ClassCompletion},
		// G058 HTTP slice: enrollment.request is closed (no credential, token, identifier or address; no foreign
		// unavailable code; no unknown operation).
		{"enrollment request with a credential", "core", `{` + envelope + `,"level":"info","event":"enrollment.request","operation":"status","outcome":"ok","credential":"abc"}`, ClassEventKeys},
		{"enrollment request with a user id", "core", `{` + envelope + `,"level":"warn","event":"enrollment.request","operation":"admit","outcome":"refused","reason":"conflict","user_id":"0f58a000-0000-4000-8000-0000000000a1"}`, ClassEventKeys},
		{"enrollment request refused as membership unavailable", "core", `{` + envelope + `,"level":"warn","event":"enrollment.request","operation":"request","outcome":"refused","reason":"membership_unavailable"}`, ClassEventKeys},
		{"enrollment request refused as enrollment unavailable", "core", `{` + envelope + `,"level":"warn","event":"enrollment.request","operation":"request","outcome":"refused","reason":"enrollment_unavailable"}`, ClassEventKeys},
		{"enrollment request unknown operation", "core", `{` + envelope + `,"level":"info","event":"enrollment.request","operation":"issue_session","outcome":"ok"}`, ClassEventKeys},
		// G062: signup.request is closed (no credential, continuation secret, slug or identifier; no foreign code; failed
		// only as signup_unavailable); the identity events refuse signup outcomes they never emit.
		{"signup request with a credential", "core", `{` + envelope + `,"level":"info","event":"signup.request","operation":"status","outcome":"ok","credential":"abc"}`, ClassEventKeys},
		{"signup request with a slug", "core", `{` + envelope + `,"level":"warn","event":"signup.request","operation":"create","outcome":"refused","reason":"slug_taken","slug":"acme"}`, ClassEventKeys},
		{"signup request with a user id", "core", `{` + envelope + `,"level":"warn","event":"signup.request","operation":"prove","outcome":"refused","reason":"signup_context_mismatch","user_id":"0f58a000-0000-4000-8000-0000000000a1"}`, ClassEventKeys},
		{"signup request refused as signup unavailable", "core", `{` + envelope + `,"level":"warn","event":"signup.request","operation":"create","outcome":"refused","reason":"signup_unavailable"}`, ClassEventKeys},
		{"signup request failed as slug taken", "core", `{` + envelope + `,"level":"error","event":"signup.request","operation":"create","outcome":"failed","reason":"slug_taken"}`, ClassEventKeys},
		{"signup request refused as enrollment unavailable", "core", `{` + envelope + `,"level":"warn","event":"signup.request","operation":"create","outcome":"refused","reason":"enrollment_unavailable"}`, ClassEventKeys},
		{"signup request unknown operation", "core", `{` + envelope + `,"level":"info","event":"signup.request","operation":"issue_session","outcome":"ok"}`, ClassEventKeys},
		{"signup request ok with a reason", "core", `{` + envelope + `,"level":"info","event":"signup.request","operation":"create","outcome":"ok","reason":"slug_taken"}`, ClassEventKeys},
		{"login start as tenant_created", "core", `{` + envelope + `,"level":"info","event":"auth.login.started","outcome":"tenant_created"}`, ClassEventKeys},
		{"callback as signup_context_mismatch", "core", `{` + envelope + `,"level":"info","event":"auth.callback.completed","outcome":"signup_context_mismatch"}`, ClassEventKeys},
		// G071 (Plan r4 S14/S19): the startup records cannot carry a raw configured value; outcome, level and members are
		// coupled; default_tenant_unavailable carries no config_keys and is the only new exit category.
		{"default tenant invalid slug with the value", "core", `{` + envelope + `,"level":"warn","event":"default_tenant","outcome":"inactive","reason":"invalid_slug","slug":"alpha-agency"}`, ClassEventKeys},
		{"default tenant unknown slug noncanonical value", "core", `{` + envelope + `,"level":"warn","event":"default_tenant","outcome":"inactive","reason":"unknown_slug","slug":"Alpha-Agency"}`, ClassEventKeys},
		{"default tenant unknown slug without slug", "core", `{` + envelope + `,"level":"warn","event":"default_tenant","outcome":"inactive","reason":"unknown_slug"}`, ClassEventKeys},
		{"default tenant active without organization_id", "core", `{` + envelope + `,"level":"info","event":"default_tenant","outcome":"active","slug":"alpha-agency"}`, ClassEventKeys},
		{"default tenant active at warn", "core", `{` + envelope + `,"level":"warn","event":"default_tenant","outcome":"active","slug":"alpha-agency","organization_id":"0f71a000-0000-4000-8000-0000000000a1"}`, ClassEventKeys},
		{"root page with the value", "core", `{` + envelope + `,"level":"warn","event":"platform_root_page","outcome":"invalid_value","fallback":"signup_redirect","value":"Landing"}`, ClassEventKeys},
		{"default tenant outage with config_keys", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"default_tenant_unavailable","config_keys":["PLATFORM_BASE_DOMAIN"]}`, ClassEventKeys},
		{"default tenant uncategorized exit", "core", `{` + envelope + `,"level":"error","event":"core.exit","error_type":"default_tenant_missing"}`, ClassErrorType},
		{"tenancy retirement code on another operation", "core", `{` + envelope + `,"level":"warn","event":"tenancy.operation","operation":"group.create","outcome":"rejected","error_code":"retirement_blocked"}`, ClassEventKeys},
		{"tenancy recipient_ineligible on a Group", "core", `{` + envelope + `,"level":"warn","event":"tenancy.operation","operation":"group.rename","outcome":"rejected","error_code":"recipient_ineligible"}`, ClassEventKeys},
		{"tenancy retirement with scope_type", "core", `{` + envelope + `,"level":"info","event":"tenancy.operation","operation":"retirement.retire","scope_type":"organization","outcome":"ok"}`, ClassEventKeys},
		{"access guarded operation without a capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"group.get","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access invitation revoke with a foreign capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"invitation.revoke","capability":"organization.manage_owners","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access invitation get with a read capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"invitation.get","capability":"group.read","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access membership leave with manage_members", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"membership.leave","capability":"organization.manage_members","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access retirement eligibility with a capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"retirement.eligibility","capability":"organization.leave","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access membership list without a capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"membership.list","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access guarded operation established", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"group.get","capability":"group.read","outcome":"established","reason":"context_established"}`, ClassEventKeys},
		{"access guarded operation with another capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"group.get","capability":"group.update","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access unknown operation", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"business_profile.delete","capability":"business_profile.profile.update","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access unknown capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"organization.get","capability":"organization.admin","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"access unknown reason", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"organization.get","capability":"organization.read","outcome":"deny","reason":"denied"}`, ClassEventKeys},
		// G058: the admission refusal reasons belong to enrollment.admit denials only.
		{"enrollment reason on another operation", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"invitation.consume","outcome":"deny","reason":"recipient_ineligible"}`, ClassEventKeys},
		{"host_changed on a guarded operation", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"membership.remove","capability":"organization.manage_members","outcome":"deny","reason":"host_changed"}`, ClassEventKeys},
		{"enrollment reason as failed", "core", `{` + envelope + `,"level":"error","event":"access.decision","operation":"enrollment.admit","outcome":"failed","reason":"enrollment_closed"}`, ClassEventKeys},
		{"enrollment reason as allow", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"enrollment.admit","outcome":"allow","reason":"enrollment_restricted"}`, ClassEventKeys},
		// G058 request slice: approval_revoked and enrollment_restricted belong to enrollment.admit denials only; the
		// applicant operations carry no capability and the manager operations exactly organization.manage_members.
		{"approval_revoked on the request", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"enrollment.request","outcome":"deny","reason":"approval_revoked"}`, ClassEventKeys},
		{"restricted on the request", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"enrollment.request","outcome":"deny","reason":"enrollment_restricted"}`, ClassEventKeys},
		{"approval_revoked as failed", "core", `{` + envelope + `,"level":"error","event":"access.decision","operation":"enrollment.admit","outcome":"failed","reason":"approval_revoked"}`, ClassEventKeys},
		{"enrollment_closed on the status read", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"enrollment.status","outcome":"deny","reason":"enrollment_closed"}`, ClassEventKeys},
		{"request with a capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"enrollment.request","capability":"organization.manage_members","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"request list without capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"enrollment.requests.list","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"decide with the policy capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"enrollment.requests.decide","capability":"organization.manage_policy","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		// G058 slice 1b: enrollment_not_ready belongs to enrollment.policy.update denials only; the policy operations carry
		// their own capability.
		{"not_ready on another operation", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"organization.set_policy","capability":"organization.manage_policy","outcome":"deny","reason":"enrollment_not_ready"}`, ClassEventKeys},
		{"not_ready as failed", "core", `{` + envelope + `,"level":"error","event":"access.decision","operation":"enrollment.policy.update","capability":"organization.manage_policy","outcome":"failed","reason":"enrollment_not_ready"}`, ClassEventKeys},
		{"policy update with the read capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"enrollment.policy.update","capability":"organization.read","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		{"policy read without capability", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"enrollment.policy.get","outcome":"allow","reason":"allowed"}`, ClassEventKeys},
		// G058: the enrollment callback logs its one enrollment_status outcome, never an admission status.
		{"callback admission status as outcome", "core", `{` + envelope + `,"level":"info","event":"auth.callback.completed","outcome":"admitted"}`, ClassEventKeys},
		{"access extra key", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"organization.get","capability":"organization.read","outcome":"allow","reason":"allowed","role":"org_owner"}`, ClassEventKeys},
		{"access identifier key", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"organization.get","capability":"organization.read","outcome":"allow","reason":"allowed","organization_id":"00000000-0000-0000-0000-000000000001"}`, ClassEventKeys},
		{"access UUID-shaped value", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"organization.get","capability":"organization.read","outcome":"allow","reason":"00000000-0000-0000-0000-000000000001"}`, ClassEventKeys},
		{"access deny with a failure reason", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"organization.get","capability":"organization.read","outcome":"deny","reason":"store_unavailable"}`, ClassEventKeys},
		{"access failed with a denial reason", "core", `{` + envelope + `,"level":"error","event":"access.decision","operation":"organization.get","capability":"organization.read","outcome":"failed","reason":"role_not_permitted"}`, ClassEventKeys},
		{"access allow with a denial reason", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"organization.get","capability":"organization.read","outcome":"allow","reason":"role_not_permitted"}`, ClassEventKeys},
		{"access deny at info", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"organization.get","capability":"organization.read","outcome":"deny","reason":"role_not_permitted"}`, ClassEventKeys},
		{"access request_id without trace_id", "core", `{` + envelope + `,"level":"info","event":"access.decision","operation":"organization.get","capability":"organization.read","outcome":"allow","reason":"allowed","request_id":"req-12345678"}`, ClassEventKeys},
		// G011: the audit diagnostics carry closed values only; a trigger-owned action, a reason or an identifier is refused.
		{"audit read with another capability", "core", `{` + envelope + `,"level":"warn","event":"access.decision","operation":"audit.read","capability":"organization.read","outcome":"deny","reason":"role_not_permitted"}`, ClassEventKeys},
		{"audit stage on another operation", "core", `{` + envelope + `,"level":"error","event":"storage.operation","operation":"object.put","outcome":"failed","reason":"store_unavailable","stage":"audit"}`, ClassEventKeys},
		{"record_failed trigger-owned action", "core", `{` + envelope + `,"level":"error","event":"audit.record_failed","action":"session.revoked"}`, ClassEventKeys},
		{"record_failed at warn", "core", `{` + envelope + `,"level":"warn","event":"audit.record_failed","action":"access.denied"}`, ClassEventKeys},
		{"record_failed with an identifier", "core", `{` + envelope + `,"level":"error","event":"audit.record_failed","action":"access.denied","organization_id":"0f8fad5b-d9cb-469f-a165-70867728950e"}`, ClassEventKeys},
		{"gap unknown operation", "core", `{` + envelope + `,"level":"error","event":"audit.gap","operation":"superseded","independent":"recorded"}`, ClassEventKeys},
		{"gap unknown independent", "core", `{` + envelope + `,"level":"error","event":"audit.gap","operation":"logout","independent":"retried"}`, ClassEventKeys},
		{"gap request_id without trace_id", "core", `{` + envelope + `,"level":"error","event":"audit.gap","operation":"logout","independent":"failed","request_id":"req-12345678"}`, ClassEventKeys},
		{"missing span", "core", line(t, completion("core", map[string]any{"span_id": nil})), ClassCompletion},
		{"zero span", "core", line(t, completion("core", map[string]any{"span_id": "0000000000000000"})), ClassCompletion},
		{"zero trace", "core", line(t, completion("core", map[string]any{"trace_id": "00000000000000000000000000000000"})), ClassCompletion},
		{"error_type on completed", "core", line(t, completion("core", map[string]any{"error_type": "panic"})), ClassCompletion},
		{"failed without error_type", "core", line(t, completion("core", map[string]any{"outcome": "failed", "level": "error"})), ClassCompletion},
		{"extra key", "core", line(t, completion("core", map[string]any{"path": "/readyz?token=x"})), ClassCompletion},
		{"info on 503", "core", line(t, completion("core", map[string]any{"status": 503})), ClassCompletion},
		{"parent on generated trace", "core", line(t, completion("core", map[string]any{"trace_source": "generated"})), ClassCompletion},
		{"inbound trace without parent", "core", line(t, completion("core", map[string]any{"parent_span_id": nil})), ClassCompletion},
		{"status 0 completed", "core", line(t, completion("core", map[string]any{"status": 0})), ClassCompletion},
		{"raw path route", "core", line(t, completion("core", map[string]any{"route": "/readyz?x=1"})), ClassCompletion},
		{"web route on core", "core", line(t, completion("core", map[string]any{"route": "/api/ready"})), ClassCompletion},
		// G004: only the five exact core identity routes; never another, a near match, a raw path, another service or
		// an extra field.
		{"unknown identity route", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/token"})), ClassCompletion},
		{"identity route prefix", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth"})), ClassCompletion},
		{"identity route trailing slash", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/session/"})), ClassCompletion},
		// G014: only the two exact MCP setup routes; never a prefix or a trailing slash.
		{"MCP setup route prefix", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/mcp"})), ClassCompletion},
		{"MCP setup route trailing slash", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/mcp/connections/"})), ClassCompletion},
		// G016: only the ten exact domain/host/relay routes; never a prefix, a trailing slash, a query or another service.
		{"domain route prefix", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/domain"})), ClassCompletion},
		{"domain route trailing slash", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains/"})), ClassCompletion},
		{"host resolve with query", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/hosts/resolve?origin=https://a.example"})), ClassCompletion},
		{"unknown domain route", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains/claim", "method": "POST"})), ClassCompletion},
		{"host resolve on web", "web", line(t, completion("web", map[string]any{"route": "/internal/v1/hosts/resolve", "core_probe": nil})), ClassCompletion},
		{"relay on intelligence", "intelligence", line(t, completion("intelligence", map[string]any{"route": "/internal/v1/auth/relay"})), ClassCompletion},
		{"domain record with hostname", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/domains", "hostname": "portal.example.com"})), ClassCompletion},
		// G017: only the six exact branding routes; never a prefix, a trailing slash, a query or another service.
		{"branding route trailing slash", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding/"})), ClassCompletion},
		{"branding logo with query", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding/logo?v=0123456789ab"})), ClassCompletion},
		{"unknown branding route", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/branding/objects", "method": "POST"})), ClassCompletion},
		{"branding public on web", "web", line(t, completion("web", map[string]any{"route": "/internal/v1/branding/public", "core_probe": nil})), ClassCompletion},
		// G062: only the seven exact signup routes; never a prefix, a trailing slash, a query, another path or another
		// service.
		{"signup route prefix", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup"})), ClassCompletion},
		{"signup route trailing slash", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/status/"})), ClassCompletion},
		{"signup status with query", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/status?origin=https://a.example"})), ClassCompletion},
		{"unknown signup route", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/signup/organizations/retry", "method": "POST"})), ClassCompletion},
		{"signup route on web", "web", line(t, completion("web", map[string]any{"route": "/internal/v1/signup/status", "core_probe": nil})), ClassCompletion},
		// G060 (scope amendment 02): the Group-authority route is a core route only.
		{"group authority on web", "web", line(t, completion("web", map[string]any{"route": "/internal/v1/organization/group-authority", "core_probe": nil})), ClassCompletion},
		{"group authority on intelligence", "intelligence", line(t, completion("intelligence", map[string]any{"route": "/internal/v1/organization/group-authority", "method": "POST"})), ClassCompletion},
		{"branding.request with a host tag", "core", `{` + envelope + `,"level":"info","event":"branding.request","operation":"get","outcome":"ok","host_tag":"0123456789ab"}`, ClassEventKeys},
		{"branding.request unknown operation", "core", `{` + envelope + `,"level":"info","event":"branding.request","operation":"set_slug","outcome":"ok"}`, ClassEventKeys},
		{"branding.request refused as unavailable", "core", `{` + envelope + `,"level":"warn","event":"branding.request","operation":"update","outcome":"refused","reason":"branding_unavailable"}`, ClassEventKeys},
		{"branding.request failure as a refusal code", "core", `{` + envelope + `,"level":"error","event":"branding.request","operation":"set_logo","outcome":"failed","reason":"version_conflict"}`, ClassEventKeys},
		{"branding.public with a raw hostname", "core", `{` + envelope + `,"level":"info","event":"branding.public","operation":"logo","outcome":"ok","hostname":"portal.example.com"}`, ClassEventKeys},
		{"branding.public refusal without a reason", "core", `{` + envelope + `,"level":"warn","event":"branding.public","operation":"projection","outcome":"refused"}`, ClassEventKeys},
		{"branding.public session reason", "core", `{` + envelope + `,"level":"warn","event":"branding.public","operation":"projection","outcome":"refused","reason":"unauthenticated"}`, ClassEventKeys},
		{"identity route case", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/Login"})), ClassCompletion},
		{"identity route partial", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/provider-logout"})), ClassCompletion},
		{"identity route with query", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/callback?code=c&state=s"})), ClassCompletion},
		{"identity route on web", "web", line(t, completion("web", map[string]any{"route": "/internal/v1/auth/login", "core_probe": nil})), ClassCompletion},
		{"identity route on intelligence", "intelligence", line(t, completion("intelligence", map[string]any{"route": "/internal/v1/auth/session"})), ClassCompletion},
		{"identity record with user", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/session", "user_id": "00000000-0000-0000-0000-000000000000"})), ClassCompletion},
		{"identity record with session", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/logout", "method": "POST", "session": "s"})), ClassCompletion},
		{"identity record with code", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/callback", "method": "POST", "code": "c"})), ClassCompletion},
		{"identity record with email", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/login", "method": "POST", "email": "a@b.example"})), ClassCompletion},
		{"identity record with authorization", "core", line(t, completion("core", map[string]any{"route": "/internal/v1/auth/session", "authorization": "Session x"})), ClassCompletion},
		{"web ready without core_probe", "web", line(t, completion("web", map[string]any{"core_probe": nil})), ClassCompletion},
		{"core_probe on core", "core", line(t, completion("core", map[string]any{"core_probe": "ok"})), ClassCompletion},
		{"invalid request_id", "core", line(t, completion("core", map[string]any{"request_id": "short"})), ClassCompletion},
		{"seconds precision time", "core", line(t, completion("core", map[string]any{"time": "2026-10-06T12:00:00Z"})), ClassEnvelope},
		{"offset time", "core", line(t, completion("core", map[string]any{"time": "2026-10-06T12:00:00.000+00:00"})), ClassEnvelope},
		{"impossible date", "core", line(t, completion("core", map[string]any{"time": "2026-02-30T12:00:00.000Z"})), ClassEnvelope},
		{"python level name", "core", line(t, completion("core", map[string]any{"level": "WARNING"})), ClassEnvelope},
		{"wrong service", "core", line(t, completion("web", nil)), ClassEnvelope},
		{"logger on core", "core", line(t, completion("core", map[string]any{"logger": "core"})), ClassEnvelope},
		{"intelligence without logger", "intelligence", line(t, completion("intelligence", map[string]any{"logger": nil})), ClassEnvelope},
		{"intelligence bad logger name", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","service":"intelligence","event":"library.log","logger":"bad logger name"}`, ClassEnvelope},
		{"intelligence completion wrong logger", "intelligence", line(t, completion("intelligence", map[string]any{"logger": "intelligence"})), ClassCompletion},
		{"configured wrong logger", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"configured","logger":"intelligence.x","capabilities":[]}`, ClassEventKeys},
		{"execution with content", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"execution","logger":"intelligence.observability","request_id":"req-12345678","trace_id":"` + traceA + `","capability":"chat","outcome":"ok","content":{"completion":"x"}}`, ClassEventKeys},
		{"uvicorn listening with host", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"uvicorn.listening","logger":"uvicorn.error","host":"0.0.0.0","port":8000}`, ClassEventKeys},
		{"warning with message", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","service":"intelligence","event":"python.warning","logger":"py.warnings","category":"UserWarning","message":"x"}`, ClassEventKeys},
		{"message from a third-party logger", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"intelligence.message","logger":"httpx"}`, ClassEventKeys},
		{"message with extra key", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"intelligence","event":"intelligence.message","logger":"intelligence","msg":"x"}`, ClassEventKeys},
		{"exception key", "intelligence", `{"time":"2026-10-06T12:00:00.000Z","level":"warn","service":"intelligence","event":"library.log","logger":"httpx","exception":"ValueError"}`, ClassEventKeys},
	} {
		if l := c.Classify(tc.svc, []byte(tc.line)); l.Class != tc.class {
			t.Errorf("%s: class %q, want %q", tc.name, l.Class, tc.class)
		}
	}
}

func TestWebStartupAllowList(t *testing.T) {
	want := []string{
		"^▲ Next\\.js [0-9]+\\.[0-9]+\\.[0-9]+$",
		"^- Local:" + strings.Repeat(" ", 9) + "http://[^ ]+$",
		"^- Network:" + strings.Repeat(" ", 7) + "http://[^ ]+$",
		"^✓ Ready in [0-9]+(\\.[0-9]+)?m?s$",
		"^✓ Running next\\.config took [0-9]+(\\.[0-9]+)?m?s$",
	}
	if strings.Join(WebStartupAllowList, "\n") != strings.Join(want, "\n") {
		t.Fatalf("the allow-list must be exactly the reviewed R14 list")
	}
	c := load(t)
	banner := "▲ Next.js 16.3.8\n- Local:         http://localhost:3000\n- Network:       http://0.0.0.0:3000\n✓ Running next.config took 3ms\n✓ Ready in 40ms\n"
	if sum := c.Validate("web", []Log{{Name: "web.log", Data: []byte(banner)}}, false); !sum.Pass || sum.AllowedNonJSON != 5 {
		t.Errorf("one banner: %+v", sum)
	}
	// A second process start may print the banner again.
	if sum := c.Validate("web", []Log{{Name: "web.log", Data: []byte(banner + banner)}}, false); !sum.Pass || sum.AllowedNonJSON != 10 {
		t.Errorf("two process starts: %+v", sum)
	}
	// Within one process start every pattern occurs at most once.
	repeated := banner + "✓ Ready in 41ms\n"
	sum := c.Validate("web", []Log{{Name: "web.log", Data: []byte(repeated)}}, false)
	if sum.Pass || sum.Failures[0].Class != ClassBannerRepeated || sum.Failures[0].Line != 6 {
		t.Errorf("repeated banner line: %+v", sum)
	}
}

func TestValidate(t *testing.T) {
	c := load(t)
	healthz := completion("core", map[string]any{"route": "/healthz", "request_id": "req-health-1"})
	ready := completion("core", map[string]any{"status": 503, "level": "warn", "request_id": "req-ready-1"})
	data := line(t, healthz) + "\n" + line(t, ready) + "\n" + `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"core","event":"core listening","port":8180,"app_env":"test"}` + "\n"
	sum := c.Validate("core", []Log{{Name: "core.log", Data: []byte(data)}}, false)
	if !sum.Pass || sum.Lines != 3 || sum.Events[CompletionEvent] != 2 || sum.Completions.ByRouteStatus["GET /healthz 200"] != 1 || sum.Completions.ByRouteStatus["GET /readyz 503"] != 1 {
		t.Errorf("summary: %+v", sum)
	}
	// A duplicate record for the same request fails, unless repeated IDs are allowed.
	dup := data + line(t, healthz) + "\n"
	if sum := c.Validate("core", []Log{{Name: "core.log", Data: []byte(dup)}}, false); sum.Pass || sum.Failures[0].Class != ClassDuplicateRecord || sum.Failures[0].Line != 4 {
		t.Errorf("duplicate record must fail: %+v", sum)
	}
	if sum := c.Validate("core", []Log{{Name: "core.log", Data: []byte(dup)}}, true); !sum.Pass {
		t.Errorf("repeated IDs allowed: %+v", sum)
	}
	// A planted line fails and the summary never echoes content.
	canary := "Cx0123456789abcdef"
	planted := data + "panic: " + canary + "\n" + `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"core","event":"` + canary + `"}` + "\n"
	sum = c.Validate("core", []Log{{Name: "core.log", Data: []byte(planted)}}, false)
	b, _ := json.Marshal(sum)
	if sum.Pass || sum.Unclassified != 2 || strings.Contains(string(b), canary) {
		t.Errorf("planted lines: %s", b)
	}
}

func chainLogs(t *testing.T, web, core, intel []map[string]any) map[string][]Log {
	t.Helper()
	join := func(recs []map[string]any) []byte {
		var b strings.Builder
		for _, r := range recs {
			b.WriteString(line(t, r) + "\n")
		}
		return []byte(b.String())
	}
	return map[string][]Log{
		"web":          {{Name: "web.log", Data: append([]byte("▲ Next.js 16.3.8\n✓ Ready in 1s\n"), join(web)...)}},
		"core":         {{Name: "core.log", Data: join(core)}},
		"intelligence": {{Name: "intelligence.log", Data: join(intel)}},
	}
}

func chainFile(t *testing.T, s string) *ChainFile {
	t.Helper()
	cf, err := ParseChainFile([]byte(s), map[string]bool{"web": true, "core": true, "intelligence": true})
	if err != nil {
		t.Fatalf("ParseChainFile: %v", err)
	}
	return cf
}

const supplied = `{"version":1,"cases":[{"id":"supplied","request_id":"req-chain-0001","expect":{"web":1,"core":1,"intelligence":"0-1"},"trace_id":"` + traceA + `","parent_span_id":"` + parentX + `","trace_flags":"01","request_id_source":"inbound","trace_source":"inbound"}]}`

func TestChain(t *testing.T) {
	c := load(t)
	w, co, in := completion("web", nil), completion("core", nil), completion("intelligence", nil)
	if res := c.CheckChain(chainFile(t, supplied), chainLogs(t, []map[string]any{w}, []map[string]any{co}, []map[string]any{in})); !res[0].Pass {
		t.Fatalf("valid chain: %+v", res)
	}
	// intelligence "0-1": absent is allowed.
	if res := c.CheckChain(chainFile(t, supplied), chainLogs(t, []map[string]any{w}, []map[string]any{co}, nil)); !res[0].Pass {
		t.Errorf("absent optional intelligence: %+v", res)
	}
	// Generated context at the entry: no parent on web.
	gen := `{"version":1,"cases":[{"id":"generated","request_id":"req-chain-0001","expect":{"web":1,"core":1,"intelligence":0},"trace_flags":"00","request_id_source":"inbound","trace_source":"generated"}]}`
	wg := completion("web", map[string]any{"trace_source": "generated", "parent_span_id": nil, "trace_flags": "00"})
	cg := completion("core", map[string]any{"trace_flags": "00"})
	if res := c.CheckChain(chainFile(t, gen), chainLogs(t, []map[string]any{wg}, []map[string]any{cg}, nil)); !res[0].Pass {
		t.Errorf("generated context: %+v", res)
	}

	for _, tc := range []struct {
		name      string
		web, core []map[string]any
		intel     []map[string]any
		failure   string
	}{
		{"duplicate record", []map[string]any{w}, []map[string]any{co, co}, []map[string]any{in}, "count:core"},
		{"missing record", []map[string]any{w}, nil, []map[string]any{in}, "count:core"},
		{"missing span", []map[string]any{w}, []map[string]any{completion("core", map[string]any{"span_id": nil})}, []map[string]any{in}, "record-invalid:core"},
		{"broken parent", []map[string]any{w}, []map[string]any{completion("core", map[string]any{"parent_span_id": "1111111111111111"})}, []map[string]any{in}, "parent:core"},
		{"broken intelligence parent", []map[string]any{w}, []map[string]any{co}, []map[string]any{completion("intelligence", map[string]any{"parent_span_id": spanWeb})}, "parent:intelligence"},
		{"mismatched trace", []map[string]any{w}, []map[string]any{completion("core", map[string]any{"trace_id": traceB})}, []map[string]any{in}, "trace-mismatch:core"},
		{"unexpected trace", []map[string]any{completion("web", map[string]any{"trace_id": traceB})}, []map[string]any{completion("core", map[string]any{"trace_id": traceB})}, nil, "trace-expected:web"},
		{"flags not propagated", []map[string]any{w}, []map[string]any{completion("core", map[string]any{"trace_flags": "00"})}, nil, "flags-mismatch:core"},
		{"web parent not the supplied one", []map[string]any{completion("web", map[string]any{"parent_span_id": "2222222222222222"})}, []map[string]any{co}, nil, "parent:web"},
		{"span reused", []map[string]any{w}, []map[string]any{completion("core", map[string]any{"span_id": spanWeb})}, nil, "span-duplicate:core"},
		{"upstream missing", []map[string]any{w}, nil, []map[string]any{in}, "upstream-missing:intelligence"},
		{"downstream generated", []map[string]any{w}, []map[string]any{completion("core", map[string]any{"request_id_source": "generated"})}, nil, "source:core"},
		{"entry source", []map[string]any{completion("web", map[string]any{"request_id_source": "generated"})}, []map[string]any{co}, nil, "source:web"},
	} {
		res := c.CheckChain(chainFile(t, supplied), chainLogs(t, tc.web, tc.core, tc.intel))
		if res[0].Pass || !contains(res[0].Failures, tc.failure) {
			t.Errorf("%s: failures %v, want %s", tc.name, res[0].Failures, tc.failure)
		}
	}
	// No parent supplied: a web parent is unexpected.
	noParent := strings.Replace(supplied, `"parent_span_id":"`+parentX+`",`, "", 1)
	if res := c.CheckChain(chainFile(t, noParent), chainLogs(t, []map[string]any{w}, []map[string]any{co}, nil)); res[0].Pass || !contains(res[0].Failures, "parent-unexpected:web") {
		t.Errorf("unexpected parent: %+v", res)
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func TestParseChainFile(t *testing.T) {
	all := map[string]bool{"web": true, "core": true}
	for name, s := range map[string]string{
		"no cases":         `{"version":1,"cases":[]}`,
		"bad version":      `{"version":2,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":1,"core":1}}]}`,
		"bad request id":   `{"version":1,"cases":[{"id":"a","request_id":"x","expect":{"web":1,"core":1}}]}`,
		"missing service":  `{"version":1,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":1}}]}`,
		"service w/o log":  `{"version":1,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":1,"core":1,"intelligence":0}}]}`,
		"bad count":        `{"version":1,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":2,"core":1}}]}`,
		"zero parent":      `{"version":1,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":1,"core":1},"parent_span_id":"0000000000000000"}]}`,
		"unknown field":    `{"version":1,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":1,"core":1},"parent":"x"}]}`,
		"duplicate id":     `{"version":1,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":1,"core":1}},{"id":"a","request_id":"req-12345678","expect":{"web":1,"core":1}}]}`,
		"bad flags":        `{"version":1,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":1,"core":1},"trace_flags":"ff"}]}`,
		"uppercase trace":  `{"version":1,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":1,"core":1},"trace_id":"4BF92F3577B34DA6A3CE929D0E0E4736"}]}`,
		"not json at all":  `nope`,
		"bad source value": `{"version":1,"cases":[{"id":"a","request_id":"req-12345678","expect":{"web":1,"core":1},"trace_source":"x"}]}`,
	} {
		if _, err := ParseChainFile([]byte(s), all); err == nil {
			t.Errorf("%s: must be rejected", name)
		}
	}
}

func TestCanaries(t *testing.T) {
	if _, err := ReadCanaries([]byte("\n \n")); err == nil {
		t.Errorf("an empty canary list must be an error")
	}
	if _, err := ReadCanaries([]byte("short\n")); err == nil {
		t.Errorf("a short canary must be an error")
	}
	cs, err := ReadCanaries([]byte("  canary-0001aa \r\nCANARY0002bbbb\n\n"))
	if err != nil || len(cs) != 2 || string(cs[0]) != "canary-0001aa" {
		t.Fatalf("ReadCanaries: %v %q", err, cs)
	}
	hits := ScanLines([]byte("clean\nxx canary-0001aa yy\nclean\nCANARY0002bbbb and canary-0001aa\n"), cs)
	if len(hits) != 2 || hits[0] != 2 || hits[1] != 4 {
		t.Errorf("hits = %v", hits)
	}
}

func TestCount(t *testing.T) {
	for in, want := range map[string]Count{`0`: {0, 0}, `1`: {1, 1}, `"0-1"`: {0, 1}, `"1"`: {1, 1}} {
		var c Count
		if err := json.Unmarshal([]byte(in), &c); err != nil || c != want {
			t.Errorf("%s: %+v %v", in, c, err)
		}
	}
	for _, in := range []string{`2`, `"1-2"`, `null`, `-1`} {
		var c Count
		if err := json.Unmarshal([]byte(in), &c); err == nil {
			t.Errorf("%s must be rejected", in)
		}
	}
}

func TestExtraKeysAndValues(t *testing.T) {
	f := map[string]any{"time": 1, "level": 1, "service": 1, "event": 1, "logger": 1, "port": json.Number("8180"), "app_env": "test"}
	if got := strings.Join(ExtraKeys("intelligence", f), ","); got != "app_env,port" {
		t.Errorf("intelligence extra keys %s", got)
	}
	if got := strings.Join(ExtraKeys("core", f), ","); got != "app_env,logger,port" {
		t.Errorf("core extra keys %s", got)
	}
	if ValueString(json.Number("8180")) != "8180" || ValueString("test") != "test" || ValueString(true) != "true" || ValueString(nil) != "null" {
		t.Errorf("ValueString")
	}
}
