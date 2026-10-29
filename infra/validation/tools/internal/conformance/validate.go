// Package conformance is the independent Go validator for the intelligence contract corpus
// (G003 Plan §8). It is test tooling only and is never linked into core.
//
// Normal mode (RunCorpus, CheckStreamFile) checks the pinned schema digests before anything
// else, then audits and compiles both schemas, then evaluates cases. Audit mode (AuditFile)
// classifies one schema file and never computes digests or evaluates cases.
package conformance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Global refusal classes that stop a normal-mode run before any case is evaluated.
const (
	ClassExpectationsInvalid = "expectations-invalid"
	ClassCorpusInvalid       = "corpus-invalid"
	ClassFileMissing         = "file-missing"
	ClassMissingTypes        = "missing-types"
)

// ClassError is a classified failure. It never carries schema, case or record content: Detail
// is a schema name, a case ID or a corpus-relative path.
type ClassError struct {
	Class  string
	Detail []string
}

func (e *ClassError) Error() string {
	if len(e.Detail) == 0 {
		return e.Class
	}
	return e.Class + ": " + strings.Join(e.Detail, ", ")
}

func classErr(class string, detail ...string) *ClassError {
	return &ClassError{Class: class, Detail: detail}
}

// SchemaPin is one pinned schema in expectations.json.
type SchemaPin struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

// Profile is the declared keyword and format profile.
type Profile struct {
	Keywords []string `json:"keywords"`
	Formats  []string `json:"formats"`
}

// Case is one declared corpus case.
type Case struct {
	ID     string `json:"id"`
	File   string `json:"file"`
	Schema string `json:"schema"`
	Kind   string `json:"kind"`
	Valid  *bool  `json:"valid"`
	Class  string `json:"class"`
}

// Mutation is a declared schema-negative mutation.
type Mutation struct {
	Op     string          `json:"op"`
	Path   *string         `json:"path"`
	Value  json.RawMessage `json:"value"`
	Length *int            `json:"length"`
}

// SNEntry is one declared schema-negative entry.
type SNEntry struct {
	ID       string   `json:"id"`
	Base     string   `json:"base"`
	File     string   `json:"file"`
	Mutation Mutation `json:"mutation"`
	Class    string   `json:"class"`
}

// DigestControl is one declared normal-mode digest control (no fixture file).
type DigestControl struct {
	ID         string `json:"id"`
	Base       string `json:"base"`
	Derivation string `json:"derivation"`
	Class      string `json:"class"`
}

// AuditPositive is one declared positive audit control.
type AuditPositive struct {
	ID     string `json:"id"`
	Schema string `json:"schema"`
	Class  string `json:"class"`
}

// Expectations is the parsed expectations.json.
type Expectations struct {
	Version        int                  `json:"version"`
	Schemas        map[string]SchemaPin `json:"schemas"`
	Profile        Profile              `json:"profile"`
	Cases          []Case               `json:"cases"`
	SchemaNegative []SNEntry            `json:"schema_negative"`
	DigestControls []DigestControl      `json:"digest_controls"`
	AuditPositive  []AuditPositive      `json:"audit_positive"`
}

// ExpectationsFile is the expectations file name at the corpus root.
const ExpectationsFile = "expectations.json"

var (
	caseIDRe = regexp.MustCompile(`^(EV|EI|XV|XI|NV|NI)[0-9]{2}$`)
	snIDRe   = regexp.MustCompile(`^SN[0-9]{2}$`)
	sdIDRe   = regexp.MustCompile(`^SD[0-9]{2}$`)
	saIDRe   = regexp.MustCompile(`^SA[0-9]{2}$`)
	sha256Re = regexp.MustCompile(`^[0-9a-f]{64}$`)
	slugRe   = regexp.MustCompile(`^[a-z0-9-]+$`)
	// declaredIDRe bounds identifiers echoed in diagnostics.
	declaredIDRe = regexp.MustCompile(`^[A-Z]{2}[0-9]{2}$`)
	caseClass    = setOf([]string{ClassValid, ClassInvalidJSON, ClassSchemaInvalid, ClassFraming})
	auditClass   = setOf([]string{ClassOK, ClassSchemaMissing, ClassMalformedJSON, ClassWrongDraft, ClassMetaInvalid,
		ClassUnsupportedKeyword, ClassUnsupportedFormat, ClassRemoteRef, ClassCompileFailed})
)

// LoadExpectations reads and checks corpusDir/expectations.json.
func LoadExpectations(corpusDir string) (*Expectations, error) {
	raw, err := os.ReadFile(filepath.Join(corpusDir, ExpectationsFile))
	if err != nil {
		return nil, classErr(ClassExpectationsInvalid, ExpectationsFile)
	}
	if _, err := ParseStrict(raw); err != nil {
		return nil, classErr(ClassExpectationsInvalid, "strict-parse")
	}
	var exp Expectations
	if err := json.Unmarshal(raw, &exp); err != nil {
		return nil, classErr(ClassExpectationsInvalid, "structure")
	}
	if problems := exp.check(); len(problems) > 0 {
		return nil, classErr(ClassExpectationsInvalid, problems...)
	}
	return &exp, nil
}

// check validates declarations that would otherwise make a run meaningless. A case whose
// declared valid flag and class disagree is not rejected here: it can never match, so the run
// names it as a mismatch after evaluation.
func (e *Expectations) check() []string {
	var p []string
	if e.Version != 1 {
		p = append(p, "version")
	}
	if len(e.Schemas) != 2 {
		p = append(p, "schemas")
	}
	for _, name := range []string{SchemaEnvelope, SchemaEvent} {
		pin, ok := e.Schemas[name]
		if !ok || pin.File == "" || pin.File != filepath.Base(pin.File) || strings.HasPrefix(pin.File, ".") || !sha256Re.MatchString(pin.SHA256) {
			p = append(p, "schemas."+name)
		}
	}
	if !sameSet(e.Profile.Keywords, ProfileKeywords) {
		p = append(p, "profile.keywords")
	}
	if !sameSet(e.Profile.Formats, ProfileFormats) {
		p = append(p, "profile.formats")
	}
	ids := map[string]bool{}
	seen := func(id string) {
		if ids[id] {
			p = append(p, "duplicate-id:"+id)
		}
		ids[id] = true
	}
	if len(e.Cases) == 0 {
		p = append(p, "cases")
	}
	for _, c := range e.Cases {
		seen(c.ID)
		if !caseIDRe.MatchString(c.ID) || c.Valid == nil || !caseClass[c.Class] || (c.Kind != "document" && c.Kind != "ndjson") {
			p = append(p, "case:"+safeID(c.ID))
			continue
		}
		if _, ok := e.Schemas[c.Schema]; !ok {
			p = append(p, "case-schema:"+c.ID)
			continue
		}
		if !caseFileOK(c) {
			p = append(p, "case-file:"+c.ID)
		}
	}
	sn := map[string]SNEntry{}
	for _, s := range e.SchemaNegative {
		seen(s.ID)
		sn[s.ID] = s
		if !snIDRe.MatchString(s.ID) || !auditClass[s.Class] || s.Class == ClassOK {
			p = append(p, "schema-negative:"+safeID(s.ID))
			continue
		}
		if _, ok := e.Schemas[s.Base]; !ok || !snFileOK(s) || !mutationOK(s.Mutation) {
			p = append(p, "schema-negative:"+s.ID)
		}
	}
	for _, d := range e.DigestControls {
		seen(d.ID)
		_, baseOK := e.Schemas[d.Base]
		derivOK := d.Derivation == "canonical-plus-lf"
		if ref, ok := strings.CutPrefix(d.Derivation, "bytes-of:"); ok {
			s, found := sn[ref]
			derivOK = found && s.Base == d.Base && s.Mutation.Op != "absent"
		}
		if !sdIDRe.MatchString(d.ID) || !baseOK || !derivOK || d.Class != ClassDigestMismatch {
			p = append(p, "digest-control:"+safeID(d.ID))
		}
	}
	for _, a := range e.AuditPositive {
		seen(a.ID)
		if _, ok := e.Schemas[a.Schema]; !ok || !saIDRe.MatchString(a.ID) || !auditClass[a.Class] {
			p = append(p, "audit-positive:"+safeID(a.ID))
		}
	}
	return p
}

// safeID returns id only when it is shaped like a declared identifier.
func safeID(id string) string {
	if declaredIDRe.MatchString(id) {
		return id
	}
	return "?"
}

// caseFileOK enforces the corpus naming rule: lowercase ID, "-", slug, extension, in the
// directory for its kind and schema.
func caseFileOK(c Case) bool {
	var dir, ext string
	switch {
	case c.Kind == "ndjson":
		dir, ext = "cases/ndjson/", ".ndjson"
	case c.Schema == SchemaEnvelope:
		dir, ext = "cases/envelope/", ".json"
	default:
		dir, ext = "cases/event/", ".json"
	}
	return namedFile(c.File, dir, c.ID, ext)
}

func snFileOK(s SNEntry) bool { return namedFile(s.File, "schema-negative/", s.ID, ".json") }

func namedFile(file, dir, id, ext string) bool {
	rest, ok := strings.CutPrefix(file, dir+strings.ToLower(id)+"-")
	if !ok {
		return false
	}
	slug, ok := strings.CutSuffix(rest, ext)
	return ok && slugRe.MatchString(slug)
}

func mutationOK(m Mutation) bool {
	switch m.Op {
	case "absent":
		return m.Path == nil && m.Value == nil && m.Length == nil
	case "truncate-bytes":
		return m.Path == nil && m.Value == nil && m.Length != nil && *m.Length > 0
	case "add", "replace":
		return m.Path != nil && m.Value != nil && m.Length == nil
	}
	return false
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sa := setOf(a)
	if len(sa) != len(a) {
		return false
	}
	for _, s := range b {
		if !sa[s] {
			return false
		}
	}
	return true
}

// PinnedSchemas are the normal-mode schemas: digests checked, audited and compiled.
type PinnedSchemas struct {
	SHA256 map[string]string
	byName map[string]*jsonschema.Schema
}

// Validate validates v against the named schema.
func (p *PinnedSchemas) Validate(schema string, v any) error {
	sch, ok := p.byName[schema]
	if !ok {
		return classErr(ClassExpectationsInvalid, "unknown-schema")
	}
	return sch.Validate(v)
}

// LoadPinnedSchemas is the normal-mode schema step. The digest of every schema is checked
// first, before any schema is parsed; only then do the audit stages 1-8 run. There is no
// option that disables the digest step.
func LoadPinnedSchemas(schemasDir string, exp *Expectations) (*PinnedSchemas, error) {
	names := sortedKeys(exp.Schemas)
	raws := map[string][]byte{}
	sums := map[string]string{}
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(schemasDir, exp.Schemas[name].File))
		if err != nil {
			return nil, classErr(ClassSchemaMissing, name)
		}
		sum := sha256.Sum256(raw)
		sums[name] = hex.EncodeToString(sum[:])
		if sums[name] != exp.Schemas[name].SHA256 {
			return nil, classErr(ClassDigestMismatch, name)
		}
		raws[name] = raw
	}
	out := &PinnedSchemas{SHA256: sums, byName: map[string]*jsonschema.Schema{}}
	for _, name := range names {
		res := AuditFile(filepath.Join(schemasDir, exp.Schemas[name].File), name)
		if res.Class != ClassOK {
			return nil, classErr(res.Class, name)
		}
		if string(res.Raw) != string(raws[name]) {
			// The file changed between the digest step and the audit.
			return nil, classErr(ClassDigestMismatch, name)
		}
		out.byName[name] = res.Schema
	}
	return out, nil
}

// CaseResult is one evaluated case in the shared result format.
type CaseResult struct {
	ID    string `json:"id"`
	Valid bool   `json:"valid"`
	Class string `json:"class"`
}

// Counts summarizes evaluated results.
type Counts struct {
	Total      int            `json:"total"`
	Valid      int            `json:"valid"`
	Invalid    int            `json:"invalid"`
	ByClass    map[string]int `json:"by_class"`
	EventTypes []string       `json:"event_types"`
}

// CorpusResult is the shared result format (Plan §8.3).
type CorpusResult struct {
	Validator    string            `json:"validator"`
	SchemaSHA256 map[string]string `json:"schema_sha256"`
	CorpusSHA256 string            `json:"corpus_sha256"`
	Results      []CaseResult      `json:"results"`
	Counts       Counts            `json:"counts"`
	Skipped      int               `json:"skipped"`
}

// Mismatch is a case whose evaluated verdict differs from its declaration.
type Mismatch struct {
	ID            string
	ExpectedValid bool
	ExpectedClass string
	Valid         bool
	Class         string
}

// ValidatorName is "<name> <version>": the tool and the schema library it uses.
func ValidatorName() string {
	version := "unknown"
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, d := range bi.Deps {
			if d.Path == "github.com/santhosh-tekuri/jsonschema/v6" {
				version = d.Version
			}
		}
	}
	return "contractcheck santhosh-tekuri/jsonschema/v6@" + version
}

// RunCorpus is normal mode: expectations, schema digests (first), schema audit, corpus
// layout, then every case. A *ClassError means no case was evaluated.
func RunCorpus(schemasDir, corpusDir string) (*CorpusResult, []Mismatch, error) {
	exp, err := LoadExpectations(corpusDir)
	if err != nil {
		return nil, nil, err
	}
	schemas, err := LoadPinnedSchemas(schemasDir, exp)
	if err != nil {
		return nil, nil, err
	}
	if problems := checkLayout(corpusDir, exp); len(problems) > 0 {
		return nil, nil, classErr(ClassCorpusInvalid, problems...)
	}
	corpusSum, err := CorpusSHA256(corpusDir)
	if err != nil {
		return nil, nil, classErr(ClassCorpusInvalid, "digest")
	}
	res := &CorpusResult{
		Validator:    ValidatorName(),
		SchemaSHA256: schemas.SHA256,
		CorpusSHA256: corpusSum,
		Results:      make([]CaseResult, 0, len(exp.Cases)),
		Counts:       Counts{ByClass: map[string]int{}, EventTypes: []string{}},
	}
	types := map[string]bool{}
	var mismatches []Mismatch
	for _, c := range exp.Cases {
		raw, err := os.ReadFile(filepath.Join(corpusDir, filepath.FromSlash(c.File)))
		if err != nil {
			return nil, nil, classErr(ClassCorpusInvalid, c.File)
		}
		validate := func(v any) error { return schemas.Validate(c.Schema, v) }
		var class string
		if c.Kind == "ndjson" {
			sr := CheckNDJSON(raw, validate)
			class = sr.Class
			if class == ClassValid && c.Schema == SchemaEvent {
				for _, t := range sr.Types {
					types[t] = true
				}
			}
		} else {
			var doc any
			class, doc = classifyDocument(raw, validate)
			if class == ClassValid && c.Schema == SchemaEvent {
				if obj, ok := doc.(map[string]any); ok {
					if t, ok := obj["type"].(string); ok {
						types[t] = true
					}
				}
			}
		}
		r := CaseResult{ID: c.ID, Valid: class == ClassValid, Class: class}
		res.Results = append(res.Results, r)
		res.Counts.Total++
		if r.Valid {
			res.Counts.Valid++
		} else {
			res.Counts.Invalid++
		}
		res.Counts.ByClass[class]++
		if r.Valid != *c.Valid || r.Class != c.Class {
			mismatches = append(mismatches, Mismatch{ID: c.ID, ExpectedValid: *c.Valid, ExpectedClass: c.Class, Valid: r.Valid, Class: r.Class})
		}
	}
	res.Counts.EventTypes = sortedKeys(types)
	return res, mismatches, nil
}

// checkLayout requires the corpus file set to equal expectations.json plus every declared
// case and schema-negative fixture, with no other entry; an "absent" fixture must not exist.
func checkLayout(corpusDir string, exp *Expectations) []string {
	want := map[string]bool{ExpectationsFile: true}
	var problems []string
	for _, c := range exp.Cases {
		want[c.File] = true
	}
	for _, s := range exp.SchemaNegative {
		if s.Mutation.Op == "absent" {
			if _, err := os.Lstat(filepath.Join(corpusDir, filepath.FromSlash(s.File))); err == nil {
				problems = append(problems, "present:"+s.File)
			}
			continue
		}
		want[s.File] = true
	}
	have := map[string]bool{}
	err := filepath.WalkDir(corpusDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(corpusDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if !d.Type().IsRegular() {
			problems = append(problems, "not-regular:"+rel)
			return nil
		}
		have[rel] = true
		return nil
	})
	if err != nil {
		return append(problems, "walk")
	}
	for _, f := range sortedKeys(want) {
		if !have[f] {
			problems = append(problems, "missing:"+f)
		}
	}
	for _, f := range sortedKeys(have) {
		if !want[f] {
			problems = append(problems, "undeclared:"+f)
		}
	}
	return problems
}

// CorpusSHA256 hashes every regular file under root: lines "<sha256>  <relative path>\n"
// sorted bytewise by path, then the SHA-256 of that text. It equals
// `cd ROOT && find . -type f -printf '%P\n' | LC_ALL=C sort | xargs -d '\n' sha256sum | sha256sum`.
func CorpusSHA256(root string) (string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			paths = append(paths, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)
	var b strings.Builder
	for _, p := range paths {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
		if err != nil {
			return "", err
		}
		sum := sha256.Sum256(raw)
		fmt.Fprintf(&b, "%s  %s\n", hex.EncodeToString(sum[:]), p)
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:]), nil
}

// StreamReport is the outcome of the ndjson command.
type StreamReport struct {
	Class        string   `json:"class"`
	Line         int      `json:"line,omitempty"`
	Records      int      `json:"records"`
	Types        []string `json:"types,omitempty"`
	MissingTypes []string `json:"missing_types,omitempty"`
}

// CheckStreamFile is normal-mode stream checking: digests first, audit, then the file. With
// requireTypes (events only) every listed type must occur in a valid stream.
func CheckStreamFile(schemasDir, corpusDir, file, schema string, requireTypes []string) (*StreamReport, error) {
	exp, err := LoadExpectations(corpusDir)
	if err != nil {
		return nil, err
	}
	schemas, err := LoadPinnedSchemas(schemasDir, exp)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, classErr(ClassFileMissing)
	}
	sr := CheckNDJSON(raw, func(v any) error { return schemas.Validate(schema, v) })
	rep := &StreamReport{Class: sr.Class, Line: sr.Line, Records: sr.Records}
	if sr.Class != ClassValid || schema != SchemaEvent {
		return rep, nil
	}
	seen := setOf(sr.Types)
	rep.Types = sortedKeys(seen)
	for _, t := range requireTypes {
		if !seen[t] {
			rep.MissingTypes = append(rep.MissingTypes, t)
		}
	}
	if len(rep.MissingTypes) > 0 {
		rep.Class = ClassMissingTypes
	}
	return rep, nil
}
