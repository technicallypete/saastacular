package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const envelopeSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "urn:test:envelope",
  "type": "object",
  "additionalProperties": false,
  "required": ["request_id", "deadline"],
  "properties": {
    "request_id": { "$ref": "#/$defs/request_id" },
    "deadline": { "type": "string", "format": "date-time" },
    "authorization_ref": { "type": "string", "maxLength": 16 }
  },
  "$defs": { "request_id": { "type": "string", "pattern": "^[A-Za-z0-9._-]{8,64}$" } }
}
`

const eventSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "urn:test:event",
  "oneOf": [{ "$ref": "#/$defs/text_delta" }, { "$ref": "#/$defs/done" }],
  "$defs": {
    "text_delta": {
      "type": "object", "additionalProperties": false, "required": ["type", "request_id", "text"],
      "properties": { "type": { "const": "text_delta" }, "request_id": { "type": "string" }, "text": { "type": "string" } }
    },
    "done": {
      "type": "object", "additionalProperties": false, "required": ["type", "request_id"],
      "properties": { "type": { "const": "done" }, "request_id": { "type": "string" } }
    }
  }
}
`

const (
	delta = `{"type":"text_delta","request_id":"req-12345678","text":"hi"}`
	done  = `{"type":"done","request_id":"req-12345678"}`
)

type fixture struct{ schemas, corpus string }

func write(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func digest(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func edited(t *testing.T, base string, edit func(m map[string]any)) []byte {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(base), &m); err != nil {
		t.Fatal(err)
	}
	edit(m)
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func sub(m map[string]any, keys ...string) map[string]any {
	for _, k := range keys {
		m = m[k].(map[string]any)
	}
	return m
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	root := t.TempDir()
	f := fixture{schemas: filepath.Join(root, "schemas"), corpus: filepath.Join(root, "corpus")}
	write(t, filepath.Join(f.schemas, "execution-envelope.schema.json"), []byte(envelopeSchema))
	write(t, filepath.Join(f.schemas, "stream-event.schema.json"), []byte(eventSchema))
	cases := []map[string]any{
		{"id": "EV01", "file": "cases/envelope/ev01-minimal.json", "schema": "execution-envelope", "kind": "document", "valid": true, "class": "valid", "body": `{"request_id":"req-12345678","deadline":"2027-01-01T00:00:00Z"}`},
		{"id": "EI01", "file": "cases/envelope/ei01-calendar.json", "schema": "execution-envelope", "kind": "document", "valid": false, "class": "schema-invalid", "body": `{"request_id":"req-12345678","deadline":"2027-02-30T00:00:00Z"}`},
		{"id": "EI02", "file": "cases/envelope/ei02-truncated.json", "schema": "execution-envelope", "kind": "document", "valid": false, "class": "invalid-json", "body": `{"request_id":`},
		{"id": "XV01", "file": "cases/event/xv01-done.json", "schema": "stream-event", "kind": "document", "valid": true, "class": "valid", "body": done},
		{"id": "NV01", "file": "cases/ndjson/nv01-delta-done.ndjson", "schema": "stream-event", "kind": "ndjson", "valid": true, "class": "valid", "body": delta + "\n" + done + "\n"},
		{"id": "NI01", "file": "cases/ndjson/ni01-blank-line.ndjson", "schema": "stream-event", "kind": "ndjson", "valid": false, "class": "ndjson-framing", "body": delta + "\n\n" + done + "\n"},
	}
	var declared []any
	for _, c := range cases {
		write(t, filepath.Join(f.corpus, c["file"].(string)), []byte(c["body"].(string)))
		delete(c, "body")
		declared = append(declared, c)
	}
	sn := filepath.Join(f.corpus, "schema-negative")
	write(t, filepath.Join(sn, "sn02-truncated.json"), []byte(envelopeSchema)[:64])
	write(t, filepath.Join(sn, "sn03-meta-invalid.json"), edited(t, envelopeSchema, func(m map[string]any) { sub(m, "properties", "authorization_ref")["type"] = 5 }))
	write(t, filepath.Join(sn, "sn04-keyword.json"), edited(t, envelopeSchema, func(m map[string]any) {
		m["dependentSchemas"] = map[string]any{"deadline": map[string]any{"required": []any{"request_id"}}}
	}))
	write(t, filepath.Join(sn, "sn05-format.json"), edited(t, envelopeSchema, func(m map[string]any) { sub(m, "properties", "authorization_ref")["format"] = "ipv4" }))
	write(t, filepath.Join(sn, "sn06-remote-ref.json"), edited(t, eventSchema, func(m map[string]any) {
		m["oneOf"].([]any)[0].(map[string]any)["$ref"] = "https://example.invalid/x.json"
	}))
	write(t, filepath.Join(sn, "sn07-draft.json"), edited(t, envelopeSchema, func(m map[string]any) { m["$schema"] = "http://json-schema.org/draft-07/schema#" }))
	exp := map[string]any{
		"version": 1,
		"schemas": map[string]any{
			"execution-envelope": map[string]any{"file": "execution-envelope.schema.json", "sha256": digest(envelopeSchema)},
			"stream-event":       map[string]any{"file": "stream-event.schema.json", "sha256": digest(eventSchema)},
		},
		"profile": map[string]any{
			"keywords": []string{"$schema", "$id", "$defs", "$ref", "title", "description", "type", "properties", "required", "additionalProperties", "items", "minItems", "maxItems", "minLength", "maxLength", "pattern", "enum", "const", "oneOf", "minimum", "format"},
			"formats":  []string{"date-time", "uuid"},
		},
		"cases": declared,
		"schema_negative": []any{
			map[string]any{"id": "SN01", "base": "execution-envelope", "file": "schema-negative/sn01-missing.json", "mutation": map[string]any{"op": "absent"}, "class": "schema-missing"},
			map[string]any{"id": "SN02", "base": "execution-envelope", "file": "schema-negative/sn02-truncated.json", "mutation": map[string]any{"op": "truncate-bytes", "length": 64}, "class": "schema-malformed-json"},
			map[string]any{"id": "SN03", "base": "execution-envelope", "file": "schema-negative/sn03-meta-invalid.json", "mutation": map[string]any{"op": "replace", "path": "/properties/authorization_ref/type", "value": 5}, "class": "schema-meta-invalid"},
			map[string]any{"id": "SN04", "base": "execution-envelope", "file": "schema-negative/sn04-keyword.json", "mutation": map[string]any{"op": "add", "path": "/dependentSchemas", "value": map[string]any{"deadline": map[string]any{"required": []any{"request_id"}}}}, "class": "schema-unsupported-keyword"},
			map[string]any{"id": "SN05", "base": "execution-envelope", "file": "schema-negative/sn05-format.json", "mutation": map[string]any{"op": "add", "path": "/properties/authorization_ref/format", "value": "ipv4"}, "class": "schema-unsupported-format"},
			map[string]any{"id": "SN06", "base": "stream-event", "file": "schema-negative/sn06-remote-ref.json", "mutation": map[string]any{"op": "replace", "path": "/oneOf/0/$ref", "value": "https://example.invalid/x.json"}, "class": "schema-remote-ref"},
			map[string]any{"id": "SN07", "base": "execution-envelope", "file": "schema-negative/sn07-draft.json", "mutation": map[string]any{"op": "replace", "path": "/$schema", "value": "http://json-schema.org/draft-07/schema#"}, "class": "schema-wrong-draft"},
		},
		"digest_controls": []any{
			map[string]any{"id": "SD01", "base": "execution-envelope", "derivation": "canonical-plus-lf", "class": "schema-digest-mismatch"},
			map[string]any{"id": "SD02", "base": "execution-envelope", "derivation": "bytes-of:SN03", "class": "schema-digest-mismatch"},
		},
		"audit_positive": []any{
			map[string]any{"id": "SA01", "schema": "execution-envelope", "class": "ok"},
			map[string]any{"id": "SA02", "schema": "stream-event", "class": "ok"},
		},
	}
	b, _ := json.MarshalIndent(exp, "", "  ")
	write(t, filepath.Join(f.corpus, "expectations.json"), b)
	return f
}

func runCLI(args ...string) (int, map[string]any, string) {
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	var summary map[string]any
	_ = json.Unmarshal(out.Bytes(), &summary)
	return code, summary, errb.String()
}

func TestCorpus(t *testing.T) {
	f := newFixture(t)
	out := filepath.Join(t.TempDir(), "results-go.json")
	code, sum, stderr := runCLI("corpus", "--schemas", f.schemas, "--corpus", f.corpus, "--out", out)
	if code != exitPass || sum["status"] != "pass" || sum["cases_evaluated"] != float64(6) {
		t.Fatalf("corpus: %d %v %s", code, sum, stderr)
	}
	var res map[string]any
	b, _ := os.ReadFile(out)
	if err := json.Unmarshal(b, &res); err != nil {
		t.Fatalf("results file: %v", err)
	}
	counts := res["counts"].(map[string]any)
	if counts["total"] != float64(6) || res["skipped"] != float64(0) || len(res["corpus_sha256"].(string)) != 64 ||
		strings.Join(toStrings(counts["event_types"]), ",") != "done,text_delta" {
		t.Errorf("results: %s", b)
	}
	if code, _, _ := runCLI("corpus", "--schemas", f.schemas, "--corpus", f.corpus, "--out", os.DevNull); code != exitPass {
		t.Errorf("--out %s must be accepted", os.DevNull)
	}

	// The EI39-style control: a format-only failure declared valid is named.
	expPath := filepath.Join(f.corpus, "expectations.json")
	raw, _ := os.ReadFile(expPath)
	var exp map[string]any
	if err := json.Unmarshal(raw, &exp); err != nil {
		t.Fatal(err)
	}
	for _, c := range exp["cases"].([]any) {
		if m := c.(map[string]any); m["id"] == "EI01" {
			m["valid"], m["class"] = true, "valid"
		}
	}
	flipped, _ := json.Marshal(exp)
	write(t, expPath, flipped)
	code, sum, stderr = runCLI("corpus", "--schemas", f.schemas, "--corpus", f.corpus, "--out", out)
	if code != exitFail || strings.Join(toStrings(sum["mismatches"]), ",") != "EI01" || !strings.Contains(stderr, "MISMATCH EI01") {
		t.Errorf("flipped EI01: %d %v %s", code, sum, stderr)
	}
}

func toStrings(v any) []string {
	var out []string
	arr, _ := v.([]any)
	for _, x := range arr {
		s, _ := x.(string)
		out = append(out, s)
	}
	return out
}

func TestDigestRefusal(t *testing.T) {
	f := newFixture(t)
	write(t, filepath.Join(f.schemas, "execution-envelope.schema.json"), []byte(envelopeSchema+"\n"))
	out := filepath.Join(t.TempDir(), "results-go.json")
	code, sum, _ := runCLI("corpus", "--schemas", f.schemas, "--corpus", f.corpus, "--out", out)
	if code != exitRefused || sum["class"] != "schema-digest-mismatch" || sum["cases_evaluated"] != float64(0) {
		t.Errorf("digest refusal: %d %v", code, sum)
	}
	if _, err := os.Stat(out); err == nil {
		t.Errorf("a refused run must not write a results file")
	}
	code, sum, _ = runCLI("ndjson", "--schemas", f.schemas, "--corpus", f.corpus, "--file", filepath.Join(f.corpus, "cases/ndjson/nv01-delta-done.ndjson"), "--as", "event")
	if code != exitRefused || sum["class"] != "schema-digest-mismatch" {
		t.Errorf("ndjson digest refusal: %d %v", code, sum)
	}
}

func TestNDJSON(t *testing.T) {
	f := newFixture(t)
	file := filepath.Join(f.corpus, "cases/ndjson/nv01-delta-done.ndjson")
	base := []string{"ndjson", "--schemas", f.schemas, "--corpus", f.corpus, "--as", "event", "--file"}
	if code, sum, _ := runCLI(append(base, file, "--require-types", "text_delta,done")...); code != exitPass || sum["status"] != "pass" {
		t.Errorf("valid stream: %d %v", code, sum)
	}
	if code, sum, _ := runCLI(append(base, file, "--require-types", "text_delta,usage,error,done")...); code != exitFail ||
		sum["result"].(map[string]any)["class"] != "missing-types" {
		t.Errorf("missing types: %d %v", code, sum)
	}
	dir := t.TempDir()
	for name, content := range map[string]string{
		"truncated last line": delta + "\n" + done[:10],
		"garbage line":        delta + "\nnot json\n" + done + "\n",
		"invalid event":       delta + "\n" + `{"type":"usage"}` + "\n",
		"blank line":          delta + "\n\n" + done + "\n",
		"whitespace line":     delta + "\n \n" + done + "\n",
		"empty":               "",
	} {
		p := filepath.Join(dir, strings.ReplaceAll(name, " ", "-")+".ndjson")
		write(t, p, []byte(content))
		if code, _, _ := runCLI(append(base, p)...); code != exitFail {
			t.Errorf("%s: exit %d, want %d", name, code, exitFail)
		}
	}
	if code, _, _ := runCLI(append(base, filepath.Join(dir, "missing.ndjson"))...); code != exitUsage {
		t.Errorf("missing file must be a usage error")
	}
}

func TestSchemaAuditAndNegative(t *testing.T) {
	f := newFixture(t)
	if code, sum, _ := runCLI("schema-audit", "--schema", filepath.Join(f.schemas, "stream-event.schema.json"), "--as", "event"); code != exitPass || sum["class"] != "ok" {
		t.Errorf("positive audit: %d %v", code, sum)
	}
	if code, sum, _ := runCLI("schema-audit", "--schema", filepath.Join(f.corpus, "schema-negative/sn05-format.json"), "--as", "envelope"); code != exitFail || sum["class"] != "schema-unsupported-format" {
		t.Errorf("SN05 audit: %d %v", code, sum)
	}
	if code, _, _ := runCLI("schema-audit", "--schema", filepath.Join(f.corpus, "cases/envelope/ev01-minimal.json"), "--as", "envelope"); code != exitUsage {
		t.Errorf("audit mode must refuse a non-audit path")
	}
	out := filepath.Join(t.TempDir(), "schema-negative-go.json")
	if code, sum, stderr := runCLI("schema-negative", "--schemas", f.schemas, "--corpus", f.corpus, "--out", out); code != exitPass || sum["derive"] != "pass" {
		t.Fatalf("schema-negative: %d %v %s", code, sum, stderr)
	}
	var rep map[string]any
	b, _ := os.ReadFile(out)
	if err := json.Unmarshal(b, &rep); err != nil || len(rep["results"].([]any)) != 7 || len(rep["digest_controls"].([]any)) != 2 || rep["derive"] != "pass" {
		t.Errorf("report: %s", b)
	}
	for _, d := range rep["digest_controls"].([]any) {
		if m := d.(map[string]any); m["class"] != "schema-digest-mismatch" || m["evaluated"] != float64(0) {
			t.Errorf("digest control: %v", m)
		}
	}
	if code, sum, _ := runCLI("sn-derive", "--schemas", f.schemas, "--corpus", f.corpus); code != exitPass || sum["derive"] != "pass" {
		t.Errorf("sn-derive: %d %v", code, sum)
	}
	// A fixture with a second difference fails the derivation and therefore schema-negative.
	write(t, filepath.Join(f.corpus, "schema-negative/sn05-format.json"), edited(t, envelopeSchema, func(m map[string]any) {
		sub(m, "properties", "authorization_ref")["format"] = "ipv4"
		sub(m, "properties", "authorization_ref")["maxLength"] = 17
	}))
	if code, sum, _ := runCLI("sn-derive", "--schemas", f.schemas, "--corpus", f.corpus); code != exitFail || sum["derive"] != "fail" {
		t.Errorf("sn-derive with a second difference: %d %v", code, sum)
	}
	if code, _, stderr := runCLI("schema-negative", "--schemas", f.schemas, "--corpus", f.corpus, "--out", out); code != exitFail || !strings.Contains(stderr, "derive:SN05") {
		t.Errorf("schema-negative with a derivation failure: %d %s", code, stderr)
	}
}

func TestProduceEnvelopes(t *testing.T) {
	out := filepath.Join(t.TempDir(), "go-envelopes.ndjson")
	code, sum, _ := runCLI("produce-envelopes", "--out", out)
	b, _ := os.ReadFile(out)
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	if code != exitPass || sum["envelopes"] != float64(3) || len(lines) != 3 || !strings.HasSuffix(string(b), "\n") {
		t.Fatalf("produce-envelopes: %d %v", code, sum)
	}
	for i, l := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("line %d: %v", i+1, err)
		}
		if c := m["capability"]; c != "chat" && c != "fast" && c != "reasoning" && c != "extraction" {
			t.Errorf("line %d: capability %v is not a text capability", i+1, c)
		}
	}
}

func TestUsage(t *testing.T) {
	f := newFixture(t)
	for _, args := range [][]string{
		{},
		{"nope"},
		{"corpus", "--schemas", f.schemas, "--corpus", f.corpus},
		{"corpus", "--schemas", f.schemas, "--corpus", f.corpus, "--out", "x", "extra"},
		{"ndjson", "--schemas", f.schemas, "--corpus", f.corpus, "--file", "x", "--as", "stream"},
		{"ndjson", "--schemas", f.schemas, "--corpus", f.corpus, "--file", "x", "--as", "envelope", "--require-types", "done"},
		{"schema-audit", "--schema", filepath.Join(f.schemas, "execution-envelope.schema.json"), "--as", "x"},
		{"schema-negative", "--schemas", f.schemas, "--corpus", f.corpus},
		{"sn-derive", "--schemas", f.schemas},
		{"produce-envelopes"},
		{"corpus", "--schemas", f.schemas, "--corpus", filepath.Join(f.corpus, "missing"), "--out", os.DevNull},
	} {
		if code, _, _ := runCLI(args...); code != exitUsage {
			t.Errorf("%v: exit %d, want %d", args, code, exitUsage)
		}
	}
}
