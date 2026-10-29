package conformance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// Small, self-contained schemas in the audited profile. They mirror the shapes of the real
// contracts (local $defs references, oneOf of $refs, date-time and uuid formats, anchored
// patterns) without copying them; the real corpus is exercised by the harness.
const testEnvelopeSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "urn:test:envelope",
  "title": "Test envelope",
  "type": "object",
  "additionalProperties": false,
  "required": ["request_id", "deadline", "organization_id", "authorization_ref"],
  "properties": {
    "request_id": { "$ref": "#/$defs/request_id" },
    "deadline": {
      "type": "string",
      "format": "date-time",
      "pattern": "^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\\.[0-9]{1,9})?(Z|[+-][0-9]{2}:[0-9]{2})$"
    },
    "organization_id": { "$ref": "#/$defs/uuid" },
    "authorization_ref": { "type": "string", "minLength": 1, "maxLength": 16 },
    "tools": { "type": "array", "maxItems": 2, "items": { "enum": ["read", "commit"] } }
  },
  "$defs": {
    "request_id": { "type": "string", "pattern": "^[A-Za-z0-9._-]{8,64}$" },
    "uuid": { "type": "string", "format": "uuid" }
  }
}
`

const testEventSchema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "urn:test:event",
  "oneOf": [
    { "$ref": "#/$defs/text_delta" },
    { "$ref": "#/$defs/done" }
  ],
  "$defs": {
    "request_id": { "type": "string", "pattern": "^[A-Za-z0-9._-]{8,64}$" },
    "text_delta": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type", "request_id", "text"],
      "properties": {
        "type": { "const": "text_delta" },
        "request_id": { "$ref": "#/$defs/request_id" },
        "text": { "type": "string" }
      }
    },
    "done": {
      "type": "object",
      "additionalProperties": false,
      "required": ["type", "request_id"],
      "properties": {
        "type": { "const": "done" },
        "request_id": { "$ref": "#/$defs/request_id" }
      }
    }
  }
}
`

const validEnvelope = `{"request_id":"req-12345678","deadline":"2028-02-29T23:59:60Z","organization_id":"6F1C2A1E-6D3B-4D0E-9A51-0C9F5A7E2B10","authorization_ref":"a"}`

type testCase struct {
	id, file, schema, kind, class string
	body                          string
}

var testCases = []testCase{
	{"EV01", "cases/envelope/ev01-minimal.json", SchemaEnvelope, "document", ClassValid, validEnvelope},
	{"EV02", "cases/envelope/ev02-offset-fraction.json", SchemaEnvelope, "document", ClassValid,
		`{"request_id":"a.b_c-12","deadline":"2027-06-01T12:00:00.123456789+02:00","organization_id":"00000000-0000-0000-0000-000000000000","authorization_ref":"x","tools":["read"]}`},
	{"EI01", "cases/envelope/ei01-truncated.json", SchemaEnvelope, "document", ClassInvalidJSON, `{"request_id":"req-1234`},
	{"EI02", "cases/envelope/ei02-nan.json", SchemaEnvelope, "document", ClassInvalidJSON, `{"request_id":NaN}`},
	{"EI03", "cases/envelope/ei03-top-level-array.json", SchemaEnvelope, "document", ClassSchemaInvalid, `[]`},
	{"EI04", "cases/envelope/ei04-request-id-newline.json", SchemaEnvelope, "document", ClassSchemaInvalid,
		strings.Replace(validEnvelope, `"req-12345678"`, `"req-12345678\n"`, 1)},
	{"EI05", "cases/envelope/ei05-february-30.json", SchemaEnvelope, "document", ClassSchemaInvalid,
		strings.Replace(validEnvelope, "2028-02-29T23:59:60Z", "2027-02-30T00:00:00Z", 1)},
	{"EI06", "cases/envelope/ei06-not-uuid.json", SchemaEnvelope, "document", ClassSchemaInvalid,
		strings.Replace(validEnvelope, "6F1C2A1E-6D3B-4D0E-9A51-0C9F5A7E2B10", "6F1C2A1E-6D3B-4D0E-9A51-0C9F5A7E2B1G", 1)},
	{"EI07", "cases/envelope/ei07-trailing-data.json", SchemaEnvelope, "document", ClassInvalidJSON, validEnvelope + " {}"},
	{"EI08", "cases/envelope/ei08-leap-second-not-2359.json", SchemaEnvelope, "document", ClassSchemaInvalid,
		strings.Replace(validEnvelope, "2028-02-29T23:59:60Z", "2028-02-29T12:00:60Z", 1)},
	{"EI09", "cases/envelope/ei09-trailing-comma.json", SchemaEnvelope, "document", ClassInvalidJSON, `{"request_id":"req-12345678",}`},
	{"XV01", "cases/event/xv01-text-delta.json", SchemaEvent, "document", ClassValid, `{"type":"text_delta","request_id":"req-12345678","text":""}`},
	{"XV02", "cases/event/xv02-done.json", SchemaEvent, "document", ClassValid, `{"type":"done","request_id":"req-12345678"}`},
	{"XI01", "cases/event/xi01-unknown-type.json", SchemaEvent, "document", ClassSchemaInvalid, `{"type":"vision","request_id":"req-12345678"}`},
	{"NV01", "cases/ndjson/nv01-text-done.ndjson", SchemaEvent, "ndjson", ClassValid,
		`{"type":"text_delta","request_id":"req-12345678","text":"hi"}` + "\n" + `{"type":"done","request_id":"req-12345678"}` + "\n"},
	{"NI01", "cases/ndjson/ni01-no-final-lf.ndjson", SchemaEvent, "ndjson", ClassFraming, `{"type":"done","request_id":"req-12345678"}`},
	{"NI02", "cases/ndjson/ni02-blank-line.ndjson", SchemaEvent, "ndjson", ClassFraming,
		`{"type":"done","request_id":"req-12345678"}` + "\n\n" + `{"type":"done","request_id":"req-12345678"}` + "\n"},
	{"NI03", "cases/ndjson/ni03-garbage-line.ndjson", SchemaEvent, "ndjson", ClassInvalidJSON,
		`{"type":"done","request_id":"req-12345678"}` + "\n" + "not json\n" + `{"type":"done","request_id":"req-12345678"}` + "\n"},
	{"NI04", "cases/ndjson/ni04-schema-invalid-line.ndjson", SchemaEvent, "ndjson", ClassSchemaInvalid,
		`{"type":"done","request_id":"req-12345678"}` + "\n" + `{"type":"done"}` + "\n"},
	{"NI05", "cases/ndjson/ni05-empty.ndjson", SchemaEvent, "ndjson", ClassFraming, ""},
}

type corpusFixture struct {
	schemas, corpus string
	exp             map[string]any
}

func mustParse(t *testing.T, s string) map[string]any {
	t.Helper()
	v, err := ParseStrict([]byte(s))
	if err != nil {
		t.Fatalf("test schema does not parse")
	}
	return v.(map[string]any)
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(b, '\n')
}

func writeFile(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
}

func sha(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func obj(v any) map[string]any { return v.(map[string]any) }

// newCorpus writes canonical schemas, cases, schema-negative fixtures and expectations.json.
func newCorpus(t *testing.T) *corpusFixture {
	t.Helper()
	root := t.TempDir()
	f := &corpusFixture{schemas: filepath.Join(root, "schemas"), corpus: filepath.Join(root, "corpus")}
	writeFile(t, filepath.Join(f.schemas, "execution-envelope.schema.json"), []byte(testEnvelopeSchema))
	writeFile(t, filepath.Join(f.schemas, "stream-event.schema.json"), []byte(testEventSchema))

	var cases []any
	for _, c := range testCases {
		writeFile(t, filepath.Join(f.corpus, filepath.FromSlash(c.file)), []byte(c.body))
		cases = append(cases, map[string]any{"id": c.id, "file": c.file, "schema": c.schema, "kind": c.kind, "valid": c.class == ClassValid, "class": c.class})
	}

	// Schema-negative fixtures are built here by direct map edits, independently of applyPatch.
	sn03 := mustParse(t, testEnvelopeSchema)
	obj(obj(sn03["properties"])["authorization_ref"])["type"] = 5
	sn04 := mustParse(t, testEnvelopeSchema)
	sn04["dependentSchemas"] = map[string]any{"organization_id": map[string]any{"required": []any{"request_id"}}}
	sn05 := mustParse(t, testEnvelopeSchema)
	obj(obj(sn05["properties"])["authorization_ref"])["format"] = "ipv4"
	sn06 := mustParse(t, testEventSchema)
	obj(sn06["oneOf"].([]any)[0])["$ref"] = "https://example.invalid/x.json"
	sn07 := mustParse(t, testEnvelopeSchema)
	sn07["$schema"] = "http://json-schema.org/draft-07/schema#"
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn02-truncated.json"), []byte(testEnvelopeSchema)[:64])
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn03-meta-invalid-type.json"), mustMarshal(t, sn03))
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn04-dependent-schemas.json"), mustMarshal(t, sn04))
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn05-format-ipv4.json"), mustMarshal(t, sn05))
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn06-remote-ref.json"), mustMarshal(t, sn06))
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn07-draft-07.json"), mustMarshal(t, sn07))

	f.exp = map[string]any{
		"version": 1,
		"schemas": map[string]any{
			SchemaEnvelope: map[string]any{"file": "execution-envelope.schema.json", "sha256": sha([]byte(testEnvelopeSchema))},
			SchemaEvent:    map[string]any{"file": "stream-event.schema.json", "sha256": sha([]byte(testEventSchema))},
		},
		"profile": map[string]any{"keywords": ProfileKeywords, "formats": ProfileFormats},
		"cases":   cases,
		"schema_negative": []any{
			map[string]any{"id": "SN01", "base": SchemaEnvelope, "file": "schema-negative/sn01-missing.json", "mutation": map[string]any{"op": "absent"}, "class": ClassSchemaMissing},
			map[string]any{"id": "SN02", "base": SchemaEnvelope, "file": "schema-negative/sn02-truncated.json", "mutation": map[string]any{"op": "truncate-bytes", "length": 64}, "class": ClassMalformedJSON},
			map[string]any{"id": "SN03", "base": SchemaEnvelope, "file": "schema-negative/sn03-meta-invalid-type.json", "mutation": map[string]any{"op": "replace", "path": "/properties/authorization_ref/type", "value": 5}, "class": ClassMetaInvalid},
			map[string]any{"id": "SN04", "base": SchemaEnvelope, "file": "schema-negative/sn04-dependent-schemas.json", "mutation": map[string]any{"op": "add", "path": "/dependentSchemas", "value": map[string]any{"organization_id": map[string]any{"required": []any{"request_id"}}}}, "class": ClassUnsupportedKeyword},
			map[string]any{"id": "SN05", "base": SchemaEnvelope, "file": "schema-negative/sn05-format-ipv4.json", "mutation": map[string]any{"op": "add", "path": "/properties/authorization_ref/format", "value": "ipv4"}, "class": ClassUnsupportedFormat},
			map[string]any{"id": "SN06", "base": SchemaEvent, "file": "schema-negative/sn06-remote-ref.json", "mutation": map[string]any{"op": "replace", "path": "/oneOf/0/$ref", "value": "https://example.invalid/x.json"}, "class": ClassRemoteRef},
			map[string]any{"id": "SN07", "base": SchemaEnvelope, "file": "schema-negative/sn07-draft-07.json", "mutation": map[string]any{"op": "replace", "path": "/$schema", "value": "http://json-schema.org/draft-07/schema#"}, "class": ClassWrongDraft},
		},
		"digest_controls": []any{
			map[string]any{"id": "SD01", "base": SchemaEnvelope, "derivation": "canonical-plus-lf", "class": ClassDigestMismatch},
			map[string]any{"id": "SD02", "base": SchemaEnvelope, "derivation": "bytes-of:SN03", "class": ClassDigestMismatch},
		},
		"audit_positive": []any{
			map[string]any{"id": "SA01", "schema": SchemaEnvelope, "class": ClassOK},
			map[string]any{"id": "SA02", "schema": SchemaEvent, "class": ClassOK},
		},
	}
	f.writeExpectations(t)
	return f
}

func (f *corpusFixture) writeExpectations(t *testing.T) {
	t.Helper()
	writeFile(t, filepath.Join(f.corpus, ExpectationsFile), mustMarshal(t, f.exp))
}

func (f *corpusFixture) setCase(t *testing.T, id string, valid bool, class string) {
	t.Helper()
	for _, c := range f.exp["cases"].([]any) {
		if m := c.(map[string]any); m["id"] == id {
			m["valid"], m["class"] = valid, class
		}
	}
	f.writeExpectations(t)
}

func classOf(err error) string {
	var ce *ClassError
	if errors.As(err, &ce) {
		return ce.Class
	}
	return ""
}

func TestParseStrict(t *testing.T) {
	for _, tc := range []struct {
		in string
		ok bool
	}{
		{`{"a":1}`, true},
		{" \t\r\n{\"a\":[1,2.5e3,null,true]} \n", true},
		{`"x"`, true},
		{`{"a":NaN}`, false},
		{`{"a":Infinity}`, false},
		{`{"a":-Infinity}`, false},
		{`{"a":1,}`, false},
		{`[1,2,]`, false},
		{`{"a":1`, false},
		{`{"a":1}{}`, false},
		{`{"a":1} x`, false},
		{``, false},
		{`   `, false},
		{"{\"a\":\"\xff\"}", false},
		{`{'a':1}`, false},
	} {
		_, err := ParseStrict([]byte(tc.in))
		if (err == nil) != tc.ok {
			t.Errorf("ParseStrict(%q) ok=%t, want %t", tc.in, err == nil, tc.ok)
		}
	}
	v, err := ParseStrict([]byte(`{"n":12345678901234567890}`))
	if err != nil || obj(v)["n"] != json.Number("12345678901234567890") {
		t.Errorf("numbers must be kept exactly as json.Number")
	}
}

func TestCheckNDJSON(t *testing.T) {
	accept := func(any) error { return nil }
	for _, tc := range []struct {
		in    string
		class string
		line  int
	}{
		{"{}\n", ClassValid, 0},
		{"{}\n{}\n", ClassValid, 0},
		{"", ClassFraming, 0},
		{"{}", ClassFraming, 0},
		{"{}\n\n{}\n", ClassFraming, 0},
		{"{}\n \t\n", ClassInvalidJSON, 2},
		{"{}\n\r\n", ClassInvalidJSON, 2},
		{"\n", ClassFraming, 0},
		{"{}\nnope\n{}\n", ClassInvalidJSON, 2},
		{"{}\n{\"a\":NaN}\n", ClassInvalidJSON, 2},
	} {
		got := CheckNDJSON([]byte(tc.in), accept)
		if got.Class != tc.class || got.Line != tc.line {
			t.Errorf("CheckNDJSON(%q) = %s line %d, want %s line %d", tc.in, got.Class, got.Line, tc.class, tc.line)
		}
	}
	reject := func(v any) error {
		if _, ok := obj(v)["bad"]; ok {
			return errors.New("bad")
		}
		return nil
	}
	if got := CheckNDJSON([]byte("{}\n{\"bad\":1}\nnope\n"), reject); got.Class != ClassSchemaInvalid || got.Line != 2 {
		t.Errorf("the first failing line decides: got %s line %d", got.Class, got.Line)
	}
}

func TestRunCorpusMatchesEveryCase(t *testing.T) {
	f := newCorpus(t)
	res, mismatches, err := RunCorpus(f.schemas, f.corpus)
	if err != nil {
		t.Fatalf("RunCorpus: %v", err)
	}
	for _, m := range mismatches {
		t.Errorf("mismatch %s: expected %s got %s", m.ID, m.ExpectedClass, m.Class)
	}
	if res.Counts.Total != len(testCases) || len(res.Results) != len(testCases) || res.Skipped != 0 {
		t.Errorf("counts: total %d results %d skipped %d", res.Counts.Total, len(res.Results), res.Skipped)
	}
	if got := strings.Join(res.Counts.EventTypes, ","); got != "done,text_delta" {
		t.Errorf("event_types = %q", got)
	}
	if res.SchemaSHA256[SchemaEnvelope] != sha([]byte(testEnvelopeSchema)) || res.SchemaSHA256[SchemaEvent] != sha([]byte(testEventSchema)) {
		t.Errorf("schema_sha256 must report the loaded digests")
	}
	want, _ := CorpusSHA256(f.corpus)
	if res.CorpusSHA256 != want || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(want) {
		t.Errorf("corpus_sha256 mismatch")
	}
	if !strings.HasPrefix(res.Validator, "contractcheck ") {
		t.Errorf("validator = %q", res.Validator)
	}
	b, _ := json.Marshal(res)
	var shape map[string]any
	_ = json.Unmarshal(b, &shape)
	for _, k := range []string{"validator", "schema_sha256", "corpus_sha256", "results", "counts", "skipped"} {
		if _, ok := shape[k]; !ok {
			t.Errorf("result lacks %q", k)
		}
	}
}

// The format assertion is effective: a case that only fails format: date-time, declared
// valid in a scratch copy, is named as a mismatch (the EI39 control).
func TestFlippedFormatOnlyCaseIsNamed(t *testing.T) {
	for _, id := range []string{"EI05", "EI06", "EI08", "EI04"} {
		f := newCorpus(t)
		f.setCase(t, id, true, ClassValid)
		_, mismatches, err := RunCorpus(f.schemas, f.corpus)
		if err != nil {
			t.Fatalf("RunCorpus: %v", err)
		}
		if len(mismatches) != 1 || mismatches[0].ID != id || mismatches[0].Class != ClassSchemaInvalid {
			t.Errorf("%s flipped: mismatches = %+v", id, mismatches)
		}
	}
	// A declaration whose flag and class disagree can never match.
	f := newCorpus(t)
	f.setCase(t, "EV01", false, ClassValid)
	if _, mismatches, _ := RunCorpus(f.schemas, f.corpus); len(mismatches) != 1 || mismatches[0].ID != "EV01" {
		t.Errorf("inconsistent declaration must be named: %+v", mismatches)
	}
}

func TestDigestIsCheckedFirst(t *testing.T) {
	for name, mutate := range map[string]func(f *corpusFixture) []byte{
		"canonical plus LF": func(*corpusFixture) []byte { return []byte(testEnvelopeSchema + "\n") },
		"SN03 bytes": func(f *corpusFixture) []byte {
			b, _ := os.ReadFile(filepath.Join(f.corpus, "schema-negative/sn03-meta-invalid-type.json"))
			return b
		},
		"truncated": func(*corpusFixture) []byte { return []byte(testEnvelopeSchema)[:10] },
	} {
		f := newCorpus(t)
		writeFile(t, filepath.Join(f.schemas, "execution-envelope.schema.json"), mutate(f))
		res, _, err := RunCorpus(f.schemas, f.corpus)
		if classOf(err) != ClassDigestMismatch || res != nil {
			t.Errorf("%s: class %q, want %s with no result", name, classOf(err), ClassDigestMismatch)
		}
		if _, err := CheckStreamFile(f.schemas, f.corpus, filepath.Join(f.corpus, "cases/ndjson/nv01-text-done.ndjson"), SchemaEvent, nil); classOf(err) != ClassDigestMismatch {
			t.Errorf("%s: ndjson must refuse with %s", name, ClassDigestMismatch)
		}
	}
	f := newCorpus(t)
	os.Remove(filepath.Join(f.schemas, "stream-event.schema.json"))
	if _, _, err := RunCorpus(f.schemas, f.corpus); classOf(err) != ClassSchemaMissing {
		t.Errorf("missing schema: class %q", classOf(err))
	}
}

func TestAuditClasses(t *testing.T) {
	f := newCorpus(t)
	sn := filepath.Join(f.corpus, "schema-negative")
	for _, tc := range []struct {
		path, schema, class string
	}{
		{filepath.Join(sn, "sn01-missing.json"), SchemaEnvelope, ClassSchemaMissing},
		{filepath.Join(sn, "sn02-truncated.json"), SchemaEnvelope, ClassMalformedJSON},
		{filepath.Join(sn, "sn03-meta-invalid-type.json"), SchemaEnvelope, ClassMetaInvalid},
		{filepath.Join(sn, "sn04-dependent-schemas.json"), SchemaEnvelope, ClassUnsupportedKeyword},
		{filepath.Join(sn, "sn05-format-ipv4.json"), SchemaEnvelope, ClassUnsupportedFormat},
		{filepath.Join(sn, "sn06-remote-ref.json"), SchemaEvent, ClassRemoteRef},
		{filepath.Join(sn, "sn07-draft-07.json"), SchemaEnvelope, ClassWrongDraft},
		{filepath.Join(f.schemas, "execution-envelope.schema.json"), SchemaEnvelope, ClassOK},
		{filepath.Join(f.schemas, "stream-event.schema.json"), SchemaEvent, ClassOK},
		{sn, SchemaEnvelope, ClassSchemaMissing},
	} {
		if got := AuditFile(tc.path, tc.schema).Class; got != tc.class {
			t.Errorf("audit %s = %s, want %s", filepath.Base(tc.path), got, tc.class)
		}
	}
	env := func(edit func(m map[string]any)) []byte {
		m := mustParse(t, testEnvelopeSchema)
		edit(m)
		return mustMarshal(t, m)
	}
	for name, tc := range map[string]struct {
		raw   []byte
		class string
	}{
		"top-level array":  {[]byte(`[]`), ClassMalformedJSON},
		"NaN":              {[]byte(`{"$schema":NaN}`), ClassMalformedJSON},
		"no $schema":       {env(func(m map[string]any) { delete(m, "$schema") }), ClassWrongDraft},
		"2019-09":          {env(func(m map[string]any) { m["$schema"] = "https://json-schema.org/draft/2019-09/schema" }), ClassWrongDraft},
		"nested keyword":   {env(func(m map[string]any) { obj(obj(m["properties"])["tools"])["uniqueItems"] = true }), ClassUnsupportedKeyword},
		"keyword in $defs": {env(func(m map[string]any) { obj(obj(m["$defs"])["uuid"])["contentEncoding"] = "base64" }), ClassUnsupportedKeyword},
		"format in items": {env(func(m map[string]any) {
			obj(obj(m["properties"])["tools"])["items"] = map[string]any{"format": "email"}
		}), ClassUnsupportedFormat},
		"dangling local ref":     {env(func(m map[string]any) { obj(obj(m["properties"])["request_id"])["$ref"] = "#/$defs/nope" }), ClassRemoteRef},
		"relative file ref":      {env(func(m map[string]any) { obj(obj(m["properties"])["request_id"])["$ref"] = "other.json#/$defs/x" }), ClassRemoteRef},
		"root ref":               {env(func(m map[string]any) { obj(obj(m["properties"])["request_id"])["$ref"] = "#" }), ClassRemoteRef},
		"invalid RE2 pattern":    {env(func(m map[string]any) { obj(obj(m["$defs"])["request_id"])["pattern"] = "(?=x)" }), ClassCompileFailed},
		"meta-invalid minLength": {env(func(m map[string]any) { obj(obj(m["properties"])["authorization_ref"])["minLength"] = -1 }), ClassMetaInvalid},
	} {
		if got := AuditBytes(tc.raw, SchemaEnvelope).Class; got != tc.class {
			t.Errorf("%s: class %s, want %s", name, got, tc.class)
		}
	}
}

func TestAuditPathAllowed(t *testing.T) {
	for path, ok := range map[string]bool{
		"/c/conformance/intelligence/schema-negative/sn03-meta-invalid-type.json": true,
		"/c/conformance/intelligence/schema-negative/sn01-missing.json":           true,
		"/contracts/intelligence/execution-envelope.schema.json":                  true,
		"/contracts/intelligence/stream-event.schema.json":                        true,
		"/c/conformance/intelligence/cases/envelope/ev01-minimal.json":            false,
		"/etc/passwd":                        false,
		"/c/schema-negative/../cases/x.json": false,
	} {
		if AuditPathAllowed(path) != ok {
			t.Errorf("AuditPathAllowed(%s) = %t", path, !ok)
		}
	}
}

func TestDerive(t *testing.T) {
	f := newCorpus(t)
	load := func() *Expectations {
		exp, err := LoadExpectations(f.corpus)
		if err != nil {
			t.Fatalf("LoadExpectations: %v", err)
		}
		return exp
	}
	results, ok, err := Derive(f.schemas, f.corpus, load())
	if err != nil || !ok || len(results) != 7 {
		t.Fatalf("derive on faithful fixtures: ok=%t err=%v results=%+v", ok, err, results)
	}
	// A second, undeclared difference fails the derivation.
	sn05 := mustParse(t, testEnvelopeSchema)
	obj(obj(sn05["properties"])["authorization_ref"])["format"] = "ipv4"
	obj(obj(sn05["properties"])["authorization_ref"])["maxLength"] = 17
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn05-format-ipv4.json"), mustMarshal(t, sn05))
	// SN01 must be absent.
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn01-missing.json"), []byte("{}"))
	// SN02 must be the exact prefix.
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn02-truncated.json"), []byte(testEnvelopeSchema)[:63])
	results, ok, err = Derive(f.schemas, f.corpus, load())
	if err != nil || ok {
		t.Fatalf("derive must fail: ok=%t err=%v", ok, err)
	}
	want := map[string]string{"SN01": "fixture-present", "SN02": "prefix-mismatch", "SN05": "structure-mismatch"}
	for _, r := range results {
		if r.Reason != want[r.ID] {
			t.Errorf("%s reason %q, want %q", r.ID, r.Reason, want[r.ID])
		}
	}
	// Numbers compare by value: 5 and 5.0 are the same JSON number.
	sn03 := strings.Replace(string(mustMarshal(t, func() map[string]any {
		m := mustParse(t, testEnvelopeSchema)
		obj(obj(m["properties"])["authorization_ref"])["type"] = 5
		return m
	}())), `"type": 5`, `"type": 5.0`, 1)
	f2 := newCorpus(t)
	writeFile(t, filepath.Join(f2.corpus, "schema-negative/sn03-meta-invalid-type.json"), []byte(sn03))
	if _, ok, _ := Derive(f2.schemas, f2.corpus, func() *Expectations { e, _ := LoadExpectations(f2.corpus); return e }()); !ok {
		t.Errorf("5.0 must equal 5")
	}
}

func TestApplyPatch(t *testing.T) {
	doc := mustParse(t, `{"a":[1,2,3],"b":{"c":"d"}}`)
	for _, tc := range []struct {
		op, path, value, want string
	}{
		{"add", "/a/1", `9`, `{"a":[1,9,2,3],"b":{"c":"d"}}`},
		{"add", "/a/-", `9`, `{"a":[1,2,3,9],"b":{"c":"d"}}`},
		{"replace", "/a/0", `9`, `{"a":[9,2,3],"b":{"c":"d"}}`},
		{"add", "/b/e", `{"f":true}`, `{"a":[1,2,3],"b":{"c":"d","e":{"f":true}}}`},
		{"replace", "/b/c", `null`, `{"a":[1,2,3],"b":{"c":null}}`},
		{"add", "/b~1x", `1`, `{"a":[1,2,3],"b":{"c":"d"},"b/x":1}`},
	} {
		v, _ := ParseStrict([]byte(tc.value))
		got, err := applyPatch(doc, tc.op, tc.path, v)
		want, _ := ParseStrict([]byte(tc.want))
		if err != nil || !jsonEqual(got, want) {
			t.Errorf("%s %s: err=%v", tc.op, tc.path, err)
		}
	}
	for _, tc := range [][2]string{{"replace", "/b/missing"}, {"replace", "/a/3"}, {"add", "/a/5"}, {"add", "/missing/x"}, {"add", "a"}, {"replace", "/a/01"}} {
		if _, err := applyPatch(doc, tc[0], tc[1], json.Number("1")); err == nil {
			t.Errorf("%s %s must fail", tc[0], tc[1])
		}
	}
	if !jsonEqual(doc, mustParse(t, `{"a":[1,2,3],"b":{"c":"d"}}`)) {
		t.Errorf("applyPatch must not modify its input")
	}
}

func TestRunSchemaNegative(t *testing.T) {
	f := newCorpus(t)
	rep, err := RunSchemaNegative(f.schemas, f.corpus, t.TempDir())
	if err != nil {
		t.Fatalf("RunSchemaNegative: %v", err)
	}
	if !rep.Pass || rep.Derive != "pass" || len(rep.Results) != 7 || len(rep.AuditPositive) != 2 || len(rep.DigestControls) != 2 {
		t.Fatalf("report: %+v", rep)
	}
	for _, d := range rep.DigestControls {
		if d.Class != ClassDigestMismatch || d.Evaluated != 0 {
			t.Errorf("%s: %+v", d.ID, d)
		}
	}
	b, _ := json.Marshal(rep)
	var top map[string]any
	_ = json.Unmarshal(b, &top)
	if strings.Join(sortedKeys(top), ",") != "audit_positive,derive,digest_controls,results,validator" {
		t.Errorf("top-level keys: %v", sortedKeys(top))
	}
	shape := map[string][]map[string]any{}
	for _, k := range []string{"results", "audit_positive", "digest_controls"} {
		var arr []map[string]any
		raw, _ := json.Marshal(top[k])
		_ = json.Unmarshal(raw, &arr)
		shape[k] = arr
	}
	keys := map[string]string{"results": "class,expected_class,id,match", "audit_positive": "class,id,schema", "digest_controls": "class,evaluated,id"}
	for k, want := range keys {
		if got := strings.Join(sortedKeys(shape[k][0]), ","); got != want {
			t.Errorf("%s keys %s, want %s", k, got, want)
		}
	}

	// A wrong declared class is a failure even though the audit itself "fails" as expected.
	f.exp["schema_negative"].([]any)[2].(map[string]any)["class"] = ClassUnsupportedKeyword
	f.writeExpectations(t)
	rep, err = RunSchemaNegative(f.schemas, f.corpus, t.TempDir())
	if err != nil || rep.Pass || rep.Results[2].Match {
		t.Errorf("a wrong class must fail: %+v", rep)
	}
}

func TestLayout(t *testing.T) {
	f := newCorpus(t)
	writeFile(t, filepath.Join(f.corpus, "cases/envelope/ev99-undeclared.json"), []byte(validEnvelope))
	if _, _, err := RunCorpus(f.schemas, f.corpus); classOf(err) != ClassCorpusInvalid {
		t.Errorf("undeclared file: class %q", classOf(err))
	}
	f = newCorpus(t)
	os.Remove(filepath.Join(f.corpus, "cases/event/xv02-done.json"))
	if _, _, err := RunCorpus(f.schemas, f.corpus); classOf(err) != ClassCorpusInvalid {
		t.Errorf("missing file: class %q", classOf(err))
	}
	f = newCorpus(t)
	writeFile(t, filepath.Join(f.corpus, "schema-negative/sn01-missing.json"), []byte("{}"))
	if _, _, err := RunCorpus(f.schemas, f.corpus); classOf(err) != ClassCorpusInvalid {
		t.Errorf("present SN01: class %q", classOf(err))
	}
}

func TestExpectationsChecks(t *testing.T) {
	for name, edit := range map[string]func(e map[string]any){
		"profile widened": func(e map[string]any) {
			obj(e["profile"])["keywords"] = append(append([]string{}, ProfileKeywords...), "allOf")
		},
		"format widened":    func(e map[string]any) { obj(e["profile"])["formats"] = []string{"date-time", "uuid", "email"} },
		"bad digest":        func(e map[string]any) { obj(obj(e["schemas"])[SchemaEnvelope])["sha256"] = "ABC" },
		"misplaced file":    func(e map[string]any) { obj(e["cases"].([]any)[0])["file"] = "cases/event/ev01-minimal.json" },
		"wrong name":        func(e map[string]any) { obj(e["cases"].([]any)[0])["file"] = "cases/envelope/ev02-minimal.json" },
		"missing valid":     func(e map[string]any) { delete(obj(e["cases"].([]any)[0]), "valid") },
		"duplicate id":      func(e map[string]any) { obj(e["cases"].([]any)[1])["id"] = "EV01" },
		"version":           func(e map[string]any) { e["version"] = 2 },
		"bad derivation":    func(e map[string]any) { obj(e["digest_controls"].([]any)[1])["derivation"] = "bytes-of:SN01" },
		"ok as SN class":    func(e map[string]any) { obj(e["schema_negative"].([]any)[0])["class"] = ClassOK },
		"absent with value": func(e map[string]any) { obj(obj(e["schema_negative"].([]any)[0])["mutation"])["value"] = 1 },
	} {
		f := newCorpus(t)
		edit(f.exp)
		f.writeExpectations(t)
		if _, err := LoadExpectations(f.corpus); classOf(err) != ClassExpectationsInvalid {
			t.Errorf("%s: class %q", name, classOf(err))
		}
	}
	f := newCorpus(t)
	writeFile(t, filepath.Join(f.corpus, ExpectationsFile), []byte(`{"version":1,"x":NaN}`))
	if _, err := LoadExpectations(f.corpus); classOf(err) != ClassExpectationsInvalid {
		t.Errorf("strict parse of expectations.json")
	}
}

func TestCheckStreamFile(t *testing.T) {
	f := newCorpus(t)
	file := filepath.Join(f.corpus, "cases/ndjson/nv01-text-done.ndjson")
	rep, err := CheckStreamFile(f.schemas, f.corpus, file, SchemaEvent, []string{"text_delta", "done"})
	if err != nil || rep.Class != ClassValid || strings.Join(rep.Types, ",") != "done,text_delta" {
		t.Errorf("valid stream: %+v %v", rep, err)
	}
	rep, _ = CheckStreamFile(f.schemas, f.corpus, file, SchemaEvent, []string{"usage", "done"})
	if rep.Class != ClassMissingTypes || strings.Join(rep.MissingTypes, ",") != "usage" {
		t.Errorf("missing type: %+v", rep)
	}
	rep, _ = CheckStreamFile(f.schemas, f.corpus, filepath.Join(f.corpus, "cases/ndjson/ni03-garbage-line.ndjson"), SchemaEvent, nil)
	if rep.Class != ClassInvalidJSON || rep.Line != 2 {
		t.Errorf("garbage line: %+v", rep)
	}
	if _, err := CheckStreamFile(f.schemas, f.corpus, filepath.Join(f.corpus, "nope.ndjson"), SchemaEvent, nil); classOf(err) != ClassFileMissing {
		t.Errorf("missing file: %v", err)
	}
}

func TestCorpusSHA256(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "b.txt"), []byte("b\n"))
	writeFile(t, filepath.Join(dir, "a/z.txt"), []byte("z"))
	writeFile(t, filepath.Join(dir, "B.txt"), []byte(""))
	if err := os.Symlink("b.txt", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	// Bytewise order: "B.txt" < "a/z.txt" < "b.txt"; the symlink is not a regular file.
	text := sha([]byte("")) + "  B.txt\n" + sha([]byte("z")) + "  a/z.txt\n" + sha([]byte("b\n")) + "  b.txt\n"
	got, err := CorpusSHA256(dir)
	if err != nil || got != sha([]byte(text)) {
		t.Errorf("CorpusSHA256 = %s, want %s", got, sha([]byte(text)))
	}
}

func TestProduceEnvelopes(t *testing.T) {
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	b, err := EncodeNDJSON(ProduceEnvelopes(now))
	if err != nil {
		t.Fatal(err)
	}
	lines, ok := SplitNDJSON(b)
	if !ok || len(lines) != 3 {
		t.Fatalf("want 3 LF-terminated records")
	}
	required := []string{"request_id", "trace_id", "capability", "authorization_ref", "organization_id", "user_id", "allowed_tools", "deadline", "context_refs", "input"}
	optional := []string{"group_id", "business_entity_id", "callback_grant_ref"}
	idRe := regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)
	seen := map[string]bool{}
	for i, l := range lines {
		v, err := ParseStrict(l)
		if err != nil {
			t.Fatalf("line %d does not parse", i+1)
		}
		m := obj(v)
		for _, k := range required {
			if _, ok := m[k]; !ok {
				t.Errorf("line %d lacks %s", i+1, k)
			}
		}
		id := m["request_id"].(string)
		if !idRe.MatchString(id) || seen[id] {
			t.Errorf("line %d request_id invalid or repeated", i+1)
		}
		seen[id] = true
		deadline, err := time.Parse(time.RFC3339Nano, m["deadline"].(string))
		if err != nil || !deadline.After(now) {
			t.Errorf("line %d deadline must be a future RFC 3339 instant", i+1)
		}
		present := 0
		for _, k := range optional {
			if _, ok := m[k]; ok {
				present++
			}
		}
		switch i {
		case 0:
			if present != 0 || len(m["allowed_tools"].([]any)) != 0 || len(m["context_refs"].([]any)) != 0 {
				t.Errorf("minimal envelope carries optional data")
			}
		case 1:
			if present != 3 || len(obj(m["input"])["messages"].([]any)) != 3 {
				t.Errorf("full envelope lacks optional data")
			}
		case 2:
			if len(id) != 64 || len(m["authorization_ref"].(string)) != 512 || !strings.HasSuffix(m["deadline"].(string), "+02:00") {
				t.Errorf("boundary envelope is not at the bounds")
			}
		}
	}
}
