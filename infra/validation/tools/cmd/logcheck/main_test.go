package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const canary = "Cx5f3a9c0e1d7b2468"

// rawDefaultTenant is a malformed DEFAULT_TENANT_SLUG value (G071 S3(b)): the invalid_slug record must never carry it.
const rawDefaultTenant = "Raw_Default_Tenant_7c41"

func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func runCLI(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

const (
	coreExitConfig = `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"core.exit","service":"core","error_type":"config_invalid","config_keys":["DB_PASSWORD","REDIS_PASSWORD"]}`
	coreExitListen = `{"time":"2026-10-06T12:00:00.000Z","level":"error","event":"core.exit","service":"core","error_type":"listen_failed"}`
	coreListening  = `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"core listening","service":"core","port":8180,"app_env":"test"}`
	coreHealthz    = `{"time":"2026-10-06T12:00:00.000Z","level":"info","event":"http.request.completed","service":"core","method":"GET","route":"/healthz","status":200,"duration_ms":0,"request_id":"req-cli-00001","request_id_source":"inbound","trace_id":"4bf92f3577b34da6a3ce929d0e0e4736","trace_source":"generated","span_id":"00f067aa0ba902b7","trace_flags":"00","outcome":"completed"}`
)

// G071 (Plan r4 S14/S19): the malformed-slug selection record and the lookup-outage exit.
const (
	coreDefaultTenantInvalid = `{"time":"2026-10-27T12:00:00.000Z","level":"warn","event":"default_tenant","service":"core","outcome":"inactive","reason":"invalid_slug"}`
	coreExitDefaultTenant    = `{"time":"2026-10-27T12:00:00.000Z","level":"error","event":"core.exit","service":"core","error_type":"default_tenant_unavailable"}`
)

func TestValidateCLI(t *testing.T) {
	dir := t.TempDir()
	good := write(t, dir, "core.log", coreListening+"\n"+coreHealthz+"\n")
	code, out, _ := runCLI("validate", "--service", "core", "--file", good)
	if code != exitPass || !strings.Contains(out, `"pass":true`) {
		t.Errorf("good log: %d %s", code, out)
	}
	bad := write(t, dir, "bad.log", coreListening+"\nhttp: panic serving 10.0.0.1: "+canary+"\n")
	code, out, errOut := runCLI("validate", "--service", "core", "--file", good, "--file", bad)
	if code != exitFail || !strings.Contains(out, `"class":"non-json"`) || strings.Contains(out+errOut, canary) {
		t.Errorf("planted line: %d %s", code, out)
	}
	dup := write(t, dir, "dup.log", coreHealthz+"\n"+coreHealthz+"\n")
	if code, _, _ := runCLI("validate", "--service", "core", "--file", dup); code != exitFail {
		t.Errorf("duplicate record must fail")
	}
	if code, _, _ := runCLI("validate", "--service", "core", "--file", dup, "--allow-repeated-request-ids"); code != exitPass {
		t.Errorf("repeated IDs allowed")
	}
	for _, args := range [][]string{
		{"validate", "--service", "db", "--file", good},
		{"validate", "--service", "core"},
		{"validate", "--service", "core", "--file", filepath.Join(dir, "missing.log")},
		{"validate", "--service", "core", "--file", good, "extra"},
		{"nope"},
		{},
	} {
		if code, _, _ := runCLI(args...); code != exitUsage {
			t.Errorf("%v: exit %d, want %d", args, code, exitUsage)
		}
	}
}

func startupFixture(t *testing.T) (string, []string) {
	dir := t.TempDir()
	write(t, dir, "ce1.out", coreExitConfig+"\n")
	write(t, dir, "ce1.out.rc", "1\n")
	write(t, dir, "ce2.out", coreExitListen+"\n")
	write(t, dir, "ce2.out.rc", "1")
	write(t, dir, "ce3.out", coreListening+"\n")
	write(t, dir, "ce3.out.rc", "0\n")
	write(t, dir, "dt1.out", coreDefaultTenantInvalid+"\n")
	write(t, dir, "dt1.out.rc", "0\n")
	write(t, dir, "dt2.out", coreExitDefaultTenant+"\n")
	write(t, dir, "dt2.out.rc", "1\n")
	write(t, dir, "canaries.txt", canary+"\n"+rawDefaultTenant+"\n")
	args := []string{"startup", "--service", "core",
		"--case", filepath.Join(dir, "ce1.out") + "=1:core.exit:config_invalid",
		"--case", filepath.Join(dir, "ce2.out") + "=1:core.exit:listen_failed",
		"--exactly-one-line", filepath.Join(dir, "ce2.out"),
		"--case", filepath.Join(dir, "ce3.out") + "=0:core listening:port=8180+app_env=test",
		"--exactly-one-line", filepath.Join(dir, "ce3.out"),
		"--case", filepath.Join(dir, "dt1.out") + "=0:default_tenant:outcome=inactive+reason=invalid_slug",
		"--case", filepath.Join(dir, "dt2.out") + "=1:core.exit:default_tenant_unavailable",
		"--canaries", filepath.Join(dir, "canaries.txt"),
	}
	return dir, args
}

func TestStartupCLI(t *testing.T) {
	_, args := startupFixture(t)
	code, out, _ := runCLI(args...)
	if code != exitPass {
		t.Fatalf("startup cases: %d %s", code, out)
	}
	var rep struct {
		Pass            bool `json:"pass"`
		CanariesChecked int  `json:"canaries_checked"`
		Cases           []struct {
			Pass bool `json:"pass"`
		} `json:"cases"`
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil || !rep.Pass || rep.CanariesChecked != 2 || len(rep.Cases) != 5 {
		t.Errorf("report: %s", out)
	}

	for name, mutate := range map[string]func(dir string){
		"wrong exit status": func(dir string) { write(t, dir, "ce1.out.rc", "0\n") },
		"wrong category": func(dir string) {
			write(t, dir, "ce2.out", strings.Replace(coreExitListen, "listen_failed", "serve_failed", 1)+"\n")
		},
		"listening before exit": func(dir string) { write(t, dir, "ce2.out", coreListening+"\n"+coreExitListen+"\n") },
		"raw exit print":        func(dir string) { write(t, dir, "ce2.out", "listen tcp: lookup bad-"+canary+".invalid\n") },
		"canary in a record": func(dir string) {
			write(t, dir, "ce3.out", strings.Replace(coreListening, `"test"`, `"`+canary+`"`, 1)+"\n")
		},
		"address key": func(dir string) {
			write(t, dir, "ce3.out", strings.Replace(coreListening, `"port"`, `"addr":":8180","port"`, 1)+"\n")
		},
		"wrong port":       func(dir string) { write(t, dir, "ce3.out", strings.Replace(coreListening, "8180", "8181", 1)+"\n") },
		"two exit records": func(dir string) { write(t, dir, "ce1.out", coreExitConfig+"\n"+coreExitConfig+"\n") },
		"empty output":     func(dir string) { write(t, dir, "ce3.out", "") },
	} {
		dir, args := startupFixture(t)
		mutate(dir)
		code, out, errOut := runCLI(args...)
		if code != exitFail || strings.Contains(out+errOut, canary) {
			t.Errorf("%s: exit %d (want %d), canary echoed=%t: %s", name, code, exitFail, strings.Contains(out+errOut, canary), out)
		}
	}
	// A canary hit is reported by file and line only.
	dir, args := startupFixture(t)
	write(t, dir, "ce1.out", coreExitConfig+"\n"+`{"oops":"`+canary+`"}`+"\n")
	_, out, _ = runCLI(args...)
	if !strings.Contains(out, filepath.Join(dir, "ce1.out")+":2") || strings.Contains(out, canary) {
		t.Errorf("canary hit report: %s", out)
	}
	// G071 (S14/S19): the raw configured value in the invalid_slug record fails the catalogue and the canary scan, and
	// is reported by file and line only.
	dir, args = startupFixture(t)
	dt1 := filepath.Join(dir, "dt1.out")
	write(t, dir, "dt1.out", strings.Replace(coreDefaultTenantInvalid, `"reason":"invalid_slug"`, `"reason":"invalid_slug","slug":"`+rawDefaultTenant+`"`, 1)+"\n")
	code, out, errOut := runCLI(args...)
	if code != exitFail || !strings.Contains(out, dt1+":1") || !strings.Contains(out, `"class":"event-keys-invalid"`) || strings.Contains(out+errOut, rawDefaultTenant) {
		t.Errorf("raw default-tenant value: exit %d (want %d): %s", code, exitFail, out)
	}

	// Intelligence startup failure (IE1 shape).
	dir = t.TempDir()
	write(t, dir, "ie1.out", `{"time":"2026-10-06T12:00:00.000Z","level":"error","service":"intelligence","event":"intelligence.startup.failed","logger":"intelligence","error_type":"ConfigError"}`+"\n")
	write(t, dir, "ie1.out.rc", "3\n")
	if code, out, _ := runCLI("startup", "--service", "intelligence", "--case", filepath.Join(dir, "ie1.out")+"=3:intelligence.startup.failed:ConfigError"); code != exitPass {
		t.Errorf("ie1: %d %s", code, out)
	}
	write(t, dir, "ie1.out", "Traceback (most recent call last):\n")
	if code, _, _ := runCLI("startup", "--service", "intelligence", "--case", filepath.Join(dir, "ie1.out")+"=3:intelligence.startup.failed:ConfigError"); code != exitFail {
		t.Errorf("a traceback must fail")
	}

	dir, _ = startupFixture(t)
	ce1 := filepath.Join(dir, "ce1.out")
	for name, args := range map[string][]string{
		"missing rc":         {"startup", "--service", "core", "--case", filepath.Join(dir, "canaries.txt") + "=1:core.exit"},
		"uncatalogued event": {"startup", "--service", "core", "--case", ce1 + "=1:server.exit"},
		"bad spec":           {"startup", "--service", "core", "--case", ce1},
		"bad rc in spec":     {"startup", "--service", "core", "--case", ce1 + "=x:core.exit"},
		"empty key set":      {"startup", "--service", "core", "--case", ce1 + "=1:core.exit:+"},
		"empty canary file":  {"startup", "--service", "core", "--case", ce1 + "=1:core.exit", "--canaries", write(t, dir, "empty.txt", "\n")},
		"no case":            {"startup", "--service", "core"},
	} {
		if code, _, _ := runCLI(args...); code != exitUsage {
			t.Errorf("%s: exit %d, want %d", name, code, exitUsage)
		}
	}
}

func TestParseCaseSpec(t *testing.T) {
	cs, err := parseCaseSpec("/in/ce3.out=0:core listening:port+app_env")
	if err != nil || cs.file != "/in/ce3.out" || cs.rc != 0 || cs.event != "core listening" || !cs.hasKeys || len(cs.keys) != 2 {
		t.Errorf("key set spec: %+v %v", cs, err)
	}
	cs, err = parseCaseSpec("/in/ce1.out=1:core.exit:config_invalid")
	if err != nil || cs.category != "config_invalid" || cs.hasKeys {
		t.Errorf("category spec: %+v %v", cs, err)
	}
	cs, err = parseCaseSpec("/in/x=0:core listening:port=8180+")
	if err != nil || len(cs.keys) != 1 || !cs.keys[0].hasValue || cs.keys[0].value != "8180" {
		t.Errorf("single key with value: %+v %v", cs, err)
	}
	for _, bad := range []string{"", "=0:e", "/f", "/f=0", "/f=0:", "/f=a:e", "/f=0:e:", "/f=0:e:a+a"} {
		if _, err := parseCaseSpec(bad); err == nil {
			t.Errorf("%q must be rejected", bad)
		}
	}
}

func TestScanCLI(t *testing.T) {
	dir := t.TempDir()
	cf := write(t, dir, "canaries.txt", canary+"\n")
	logs := filepath.Join(dir, "logs")
	if err := os.MkdirAll(filepath.Join(logs, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	clean := write(t, logs, "clean.log", "nothing here\n")
	if code, out, _ := runCLI("scan", "--canaries", cf, clean, logs); code != exitPass || out != "" {
		t.Errorf("clean: %d %q", code, out)
	}
	write(t, filepath.Join(logs, "sub"), "hit.log", "a\nb "+canary+" c\n")
	code, out, errOut := runCLI("scan", "--canaries", cf, logs)
	if code != exitFail || strings.TrimSpace(out) != filepath.Join(logs, "sub", "hit.log")+":2" || strings.Contains(out+errOut, canary) {
		t.Errorf("hit: %d %q", code, out)
	}
	for name, args := range map[string][]string{
		"empty canary list": {"scan", "--canaries", write(t, dir, "empty.txt", ""), clean},
		"short canary":      {"scan", "--canaries", write(t, dir, "short.txt", "abc\n"), clean},
		"missing path":      {"scan", "--canaries", cf, filepath.Join(dir, "missing")},
		"no path":           {"scan", "--canaries", cf},
		"no canaries":       {"scan", clean},
	} {
		if code, _, _ := runCLI(args...); code != exitUsage {
			t.Errorf("%s: exit %d, want %d", name, code, exitUsage)
		}
	}
}

func completionLine(svc, route, span, parent, extra string) string {
	s := `{"time":"2026-10-06T12:00:00.000Z","level":"info","service":"` + svc + `","event":"http.request.completed","method":"GET","route":"` + route +
		`","status":200,"duration_ms":1,"request_id":"req-chain-cli1","request_id_source":"inbound","trace_id":"4bf92f3577b34da6a3ce929d0e0e4736",` +
		`"trace_source":"inbound","span_id":"` + span + `","parent_span_id":"` + parent + `","trace_flags":"01","outcome":"completed"` + extra + `}`
	return s + "\n"
}

func TestChainCLI(t *testing.T) {
	dir := t.TempDir()
	web := write(t, dir, "web.log", "▲ Next.js 16.3.8\n"+completionLine("web", "/api/ready", "00f067aa0ba902b7", "b7ad6b7169203331", `,"core_probe":"ok"`))
	core := write(t, dir, "core.log", completionLine("core", "/readyz", "1a2b3c4d5e6f7081", "00f067aa0ba902b7", ""))
	intel := write(t, dir, "intel.log", completionLine("intelligence", "/healthz", "9f8e7d6c5b4a3921", "1a2b3c4d5e6f7081", `,"logger":"intelligence.http"`))
	cases := write(t, dir, "cases.json", `{"version":1,"cases":[{"id":"i","request_id":"req-chain-cli1","expect":{"web":1,"core":1,"intelligence":1},"parent_span_id":"b7ad6b7169203331","trace_flags":"01"}]}`)
	code, out, _ := runCLI("chain", "--cases", cases, "--log", "web="+web, "--log", "core="+core, "--log", "intelligence="+intel)
	if code != exitPass || !strings.Contains(out, `"pass":true`) {
		t.Errorf("valid chain: %d %s", code, out)
	}
	broken := write(t, dir, "core2.log", completionLine("core", "/readyz", "1a2b3c4d5e6f7081", "2222222222222222", ""))
	code, out, _ = runCLI("chain", "--cases", cases, "--log", "web="+web, "--log", "core="+broken, "--log", "intelligence="+intel)
	if code != exitFail || !strings.Contains(out, "parent:core") {
		t.Errorf("broken parent: %d %s", code, out)
	}
	for name, args := range map[string][]string{
		"bad log arg":            {"chain", "--cases", cases, "--log", "db=" + core},
		"expect without log":     {"chain", "--cases", cases, "--log", "web=" + web, "--log", "core=" + core},
		"missing cases":          {"chain", "--log", "web=" + web},
		"unreadable cases":       {"chain", "--cases", filepath.Join(dir, "nope.json"), "--log", "web=" + web},
		"unreadable service log": {"chain", "--cases", cases, "--log", "web=" + filepath.Join(dir, "nope.log")},
	} {
		if code, _, _ := runCLI(args...); code != exitUsage {
			t.Errorf("%s: exit %d, want %d", name, code, exitUsage)
		}
	}
}

func TestRecordsCLI(t *testing.T) {
	dir := t.TempDir()
	rec := completionLine("core", "/readyz", "1a2b3c4d5e6f7081", "00f067aa0ba902b7", "")
	log := write(t, dir, "core.log", coreListening+"\n"+rec+coreHealthz+"\n")
	code, out, _ := runCLI("records", "--service", "core", "--file", log, "--request-id", "req-chain-cli1")
	if code != exitPass || out != rec {
		t.Errorf("records: %d %q", code, out)
	}
	invalid := write(t, dir, "bad.log", strings.Replace(rec, `"span_id":"1a2b3c4d5e6f7081",`, "", 1))
	if code, out, _ := runCLI("records", "--service", "core", "--file", invalid, "--request-id", "req-chain-cli1"); code != exitFail || out != "" {
		t.Errorf("invalid matching record: %d %q", code, out)
	}
	if code, _, _ := runCLI("records", "--service", "core", "--file", log, "--request-id", "bad id"); code != exitUsage {
		t.Errorf("an invalid request ID must be refused")
	}
}
