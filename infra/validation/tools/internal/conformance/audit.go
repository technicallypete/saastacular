package conformance

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Audit classes in stage order (Plan §8.2a). The first failing stage sets the class.
const (
	ClassOK                 = "ok"
	ClassSchemaMissing      = "schema-missing"
	ClassMalformedJSON      = "schema-malformed-json"
	ClassWrongDraft         = "schema-wrong-draft"
	ClassMetaInvalid        = "schema-meta-invalid"
	ClassUnsupportedKeyword = "schema-unsupported-keyword"
	ClassUnsupportedFormat  = "schema-unsupported-format"
	ClassRemoteRef          = "schema-remote-ref"
	ClassCompileFailed      = "schema-compile-failed"
	// ClassDigestMismatch is the normal-mode refusal; audit mode never computes digests.
	ClassDigestMismatch = "schema-digest-mismatch"
)

// Draft202012 is the only accepted `$schema` value.
const Draft202012 = "https://json-schema.org/draft/2020-12/schema"

// ProfileKeywords is the audited keyword profile (Plan §8.3), implemented independently of the
// Python validator. expectations.json must declare exactly this set.
var ProfileKeywords = []string{
	"$schema", "$id", "$defs", "$ref", "title", "description", "type", "properties", "required",
	"additionalProperties", "items", "minItems", "maxItems", "minLength", "maxLength", "pattern",
	"enum", "const", "oneOf", "minimum", "format",
}

// ProfileFormats are the only formats a schema may use.
var ProfileFormats = []string{"date-time", "uuid"}

// Schema names, their audit aliases and the canonical resource URLs used with AddResource.
const (
	SchemaEnvelope = "execution-envelope"
	SchemaEvent    = "stream-event"
)

var (
	asToSchema    = map[string]string{"envelope": SchemaEnvelope, "event": SchemaEvent}
	schemaFile    = map[string]string{SchemaEnvelope: "execution-envelope.schema.json", SchemaEvent: "stream-event.schema.json"}
	auditFileName = regexp.MustCompile(`^sn[0-9]{2}-[a-z0-9-]+\.json$`)
)

// SchemaForAs maps an --as value (envelope|event) to its schema name.
func SchemaForAs(as string) (string, bool) {
	s, ok := asToSchema[as]
	return s, ok
}

// ResourceURL is the file:///contracts/... URL a schema is registered under. Nothing is ever
// fetched from it: the compiler's loader refuses every URL.
func ResourceURL(schema string) string {
	return "file:///contracts/intelligence/" + schemaFile[schema]
}

// AuditPathAllowed reports whether audit mode may read path: a schema-negative fixture name
// inside a schema-negative directory, or a canonical schema file name (positive control).
func AuditPathAllowed(path string) bool {
	clean := filepath.Clean(path)
	base := filepath.Base(clean)
	if base == schemaFile[SchemaEnvelope] || base == schemaFile[SchemaEvent] {
		return true
	}
	return filepath.Base(filepath.Dir(clean)) == "schema-negative" && auditFileName.MatchString(base)
}

// refusingLoader fails every URL, so no schema resolution can reach a network or the file
// system. Draft meta-schemas are embedded in the library and never use the loader.
type refusingLoader struct{}

var errLoadRefused = errors.New("schema loading is disabled")

func (refusingLoader) Load(string) (any, error) { return nil, errLoadRefused }

var (
	metaOnce   sync.Once
	metaSchema *jsonschema.Schema
	metaErr    error
)

// draftMetaSchema compiles the embedded Draft 2020-12 meta-schema once.
func draftMetaSchema() (*jsonschema.Schema, error) {
	metaOnce.Do(func() {
		c := jsonschema.NewCompiler()
		c.UseLoader(refusingLoader{})
		metaSchema, metaErr = c.Compile(Draft202012)
	})
	return metaSchema, metaErr
}

// AuditResult is an audit classification and, when the class is ok, the compiled schema.
type AuditResult struct {
	Class  string
	Raw    []byte
	Schema *jsonschema.Schema
}

// AuditFile runs the eight audit stages on the file at path, registered as schema.
func AuditFile(path, schema string) AuditResult {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return AuditResult{Class: ClassSchemaMissing}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return AuditResult{Class: ClassSchemaMissing}
	}
	res := AuditBytes(raw, schema)
	res.Raw = raw
	return res
}

// AuditBytes runs stages 2-8 on schema bytes.
func AuditBytes(raw []byte, schema string) AuditResult {
	// Stage 2: strict JSON with an object at the top level.
	doc, err := ParseStrict(raw)
	if err != nil {
		return AuditResult{Class: ClassMalformedJSON}
	}
	root, ok := doc.(map[string]any)
	if !ok {
		return AuditResult{Class: ClassMalformedJSON}
	}
	// Stage 3: exactly Draft 2020-12.
	if s, ok := root["$schema"].(string); !ok || s != Draft202012 {
		return AuditResult{Class: ClassWrongDraft}
	}
	// Stage 4: meta-validation against the embedded meta-schema.
	meta, err := draftMetaSchema()
	if err != nil {
		return AuditResult{Class: ClassMetaInvalid}
	}
	if err := meta.Validate(doc); err != nil {
		return AuditResult{Class: ClassMetaInvalid}
	}
	// Stage 5: every keyword in the profile.
	keywords := setOf(ProfileKeywords)
	if !walkSchemas(root, func(obj map[string]any) bool {
		for k := range obj {
			if !keywords[k] {
				return false
			}
		}
		return true
	}) {
		return AuditResult{Class: ClassUnsupportedKeyword}
	}
	// Stage 6: only the profile's formats.
	formats := setOf(ProfileFormats)
	if !walkSchemas(root, func(obj map[string]any) bool {
		f, ok := obj["format"]
		if !ok {
			return true
		}
		s, ok := f.(string)
		return ok && formats[s]
	}) {
		return AuditResult{Class: ClassUnsupportedFormat}
	}
	// Stage 7: every $ref is local (#/$defs/...) and resolves inside this document.
	if !walkSchemas(root, func(obj map[string]any) bool {
		r, ok := obj["$ref"]
		if !ok {
			return true
		}
		s, ok := r.(string)
		if !ok || !strings.HasPrefix(s, "#/$defs/") {
			return false
		}
		_, found := resolveFragment(root, s[1:])
		return found
	}) {
		return AuditResult{Class: ClassRemoteRef}
	}
	// Stage 8: compile with format assertions on and every URL refused.
	sch, err := compileSchema(schema, doc)
	if err != nil {
		return AuditResult{Class: ClassCompileFailed}
	}
	return AuditResult{Class: ClassOK, Schema: sch}
}

func compileSchema(schema string, doc any) (*jsonschema.Schema, error) {
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	c.AssertFormat()
	c.UseLoader(refusingLoader{})
	u := ResourceURL(schema)
	if u == "file:///contracts/intelligence/" {
		return nil, errors.New("unknown schema")
	}
	if err := c.AddResource(u, doc); err != nil {
		return nil, err
	}
	return c.Compile(u)
}

// walkSchemas visits every schema object reachable through the profile's applicators
// (properties, $defs, items, additionalProperties, oneOf). Boolean schemas are leaves. A
// keyword outside the profile stops the descent at its parent, which stage 5 rejects.
func walkSchemas(v any, fn func(map[string]any) bool) bool {
	obj, ok := v.(map[string]any)
	if !ok {
		return true
	}
	if !fn(obj) {
		return false
	}
	for _, kw := range []string{"properties", "$defs"} {
		if m, ok := obj[kw].(map[string]any); ok {
			for _, k := range sortedKeys(m) {
				if !walkSchemas(m[k], fn) {
					return false
				}
			}
		}
	}
	for _, kw := range []string{"items", "additionalProperties"} {
		if sub, ok := obj[kw]; ok && !walkSchemas(sub, fn) {
			return false
		}
	}
	if arr, ok := obj["oneOf"].([]any); ok {
		for _, s := range arr {
			if !walkSchemas(s, fn) {
				return false
			}
		}
	}
	return true
}

// resolveFragment resolves a URI-fragment JSON pointer ("/a/b") inside doc.
func resolveFragment(doc any, frag string) (any, bool) {
	ptr, err := url.PathUnescape(frag)
	if err != nil {
		return nil, false
	}
	tokens, ok := splitPointer(ptr)
	if !ok {
		return nil, false
	}
	cur := doc
	for _, tok := range tokens {
		switch c := cur.(type) {
		case map[string]any:
			next, ok := c[tok]
			if !ok {
				return nil, false
			}
			cur = next
		case []any:
			i, ok := arrayIndex(tok, len(c))
			if !ok || i >= len(c) {
				return nil, false
			}
			cur = c[i]
		default:
			return nil, false
		}
	}
	return cur, true
}

// splitPointer splits an RFC 6901 pointer into unescaped reference tokens.
func splitPointer(ptr string) ([]string, bool) {
	if ptr == "" {
		return nil, true
	}
	if !strings.HasPrefix(ptr, "/") {
		return nil, false
	}
	parts := strings.Split(ptr[1:], "/")
	for i, p := range parts {
		if strings.Contains(strings.ReplaceAll(strings.ReplaceAll(p, "~0", ""), "~1", ""), "~") {
			return nil, false
		}
		parts[i] = strings.ReplaceAll(strings.ReplaceAll(p, "~1", "/"), "~0", "~")
	}
	return parts, true
}

// arrayIndex parses an RFC 6901 array index (no leading zeros, no sign).
func arrayIndex(tok string, n int) (int, bool) {
	if tok == "" || (len(tok) > 1 && tok[0] == '0') {
		return 0, false
	}
	for _, r := range tok {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	i, err := strconv.Atoi(tok)
	if err != nil || i < 0 || i > n {
		return 0, false
	}
	return i, true
}

func setOf(list []string) map[string]bool {
	m := make(map[string]bool, len(list))
	for _, s := range list {
		m[s] = true
	}
	return m
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
