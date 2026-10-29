package conformance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

// DeriveResult is the derivation check for one schema-negative entry.
type DeriveResult struct {
	ID    string `json:"id"`
	Match bool   `json:"match"`
	// Reason is a fixed code when Match is false.
	Reason string `json:"reason,omitempty"`
}

// Derive applies each declared mutation to the parsed canonical base and requires structural
// equality with the parsed fixture; truncate-bytes requires byte equality with the prefix, and
// absent requires that no file exists. The canonical bases must match their pinned digests.
func Derive(schemasDir, corpusDir string, exp *Expectations) ([]DeriveResult, bool, error) {
	bases := map[string][]byte{}
	for _, name := range sortedKeys(exp.Schemas) {
		raw, err := os.ReadFile(filepath.Join(schemasDir, exp.Schemas[name].File))
		if err != nil {
			return nil, false, classErr(ClassSchemaMissing, name)
		}
		sum := sha256.Sum256(raw)
		if hex.EncodeToString(sum[:]) != exp.Schemas[name].SHA256 {
			return nil, false, classErr(ClassDigestMismatch, name)
		}
		bases[name] = raw
	}
	all := true
	out := make([]DeriveResult, 0, len(exp.SchemaNegative))
	for _, s := range exp.SchemaNegative {
		r := DeriveResult{ID: s.ID, Match: true}
		if reason := deriveOne(bases[s.Base], filepath.Join(corpusDir, filepath.FromSlash(s.File)), s.Mutation); reason != "" {
			r.Match, r.Reason = false, reason
			all = false
		}
		out = append(out, r)
	}
	return out, all, nil
}

func deriveOne(base []byte, fixture string, m Mutation) string {
	if m.Op == "absent" {
		if _, err := os.Lstat(fixture); !errors.Is(err, os.ErrNotExist) {
			return "fixture-present"
		}
		return ""
	}
	raw, err := os.ReadFile(fixture)
	if err != nil {
		return "fixture-missing"
	}
	switch m.Op {
	case "truncate-bytes":
		if *m.Length >= len(base) || !bytes.Equal(raw, base[:*m.Length]) {
			return "prefix-mismatch"
		}
		return ""
	case "add", "replace":
		doc, err := ParseStrict(base)
		if err != nil {
			return "base-malformed"
		}
		value, err := ParseStrict(m.Value)
		if err != nil {
			return "mutation-invalid"
		}
		want, err := applyPatch(doc, m.Op, *m.Path, value)
		if err != nil {
			return "mutation-invalid"
		}
		got, err := ParseStrict(raw)
		if err != nil {
			return "fixture-malformed"
		}
		if !jsonEqual(want, got) {
			return "structure-mismatch"
		}
		return ""
	}
	return "mutation-invalid"
}

var errPatch = errors.New("patch")

// applyPatch applies one RFC 6902 add or replace operation to a deep copy of doc.
func applyPatch(doc any, op, path string, value any) (any, error) {
	tokens, ok := splitPointer(path)
	if !ok {
		return nil, errPatch
	}
	if len(tokens) == 0 {
		return clone(value), nil
	}
	root := clone(doc)
	parent := root
	for _, tok := range tokens[:len(tokens)-1] {
		switch c := parent.(type) {
		case map[string]any:
			next, ok := c[tok]
			if !ok {
				return nil, errPatch
			}
			parent = next
		case []any:
			i, ok := arrayIndex(tok, len(c))
			if !ok || i >= len(c) {
				return nil, errPatch
			}
			parent = c[i]
		default:
			return nil, errPatch
		}
	}
	last := tokens[len(tokens)-1]
	switch c := parent.(type) {
	case map[string]any:
		if _, exists := c[last]; op == "replace" && !exists {
			return nil, errPatch
		}
		c[last] = clone(value)
	case []any:
		// Arrays are replaced inside their parent, so locate the parent container again.
		return patchArray(root, tokens, op, value)
	default:
		return nil, errPatch
	}
	return root, nil
}

// patchArray handles a final array token: add inserts (or appends with "-"), replace overwrites.
func patchArray(root any, tokens []string, op string, value any) (any, error) {
	arr, _ := lookupTokens(root, tokens[:len(tokens)-1])
	c := arr.([]any)
	last := tokens[len(tokens)-1]
	var updated []any
	switch {
	case op == "add" && last == "-":
		updated = append(append([]any{}, c...), clone(value))
	case op == "add":
		i, ok := arrayIndex(last, len(c))
		if !ok {
			return nil, errPatch
		}
		updated = append(append(append([]any{}, c[:i]...), clone(value)), c[i:]...)
	default:
		i, ok := arrayIndex(last, len(c))
		if !ok || i >= len(c) {
			return nil, errPatch
		}
		updated = append([]any{}, c...)
		updated[i] = clone(value)
	}
	if len(tokens) == 1 {
		return updated, nil
	}
	grand, _ := lookupTokens(root, tokens[:len(tokens)-2])
	key := tokens[len(tokens)-2]
	switch g := grand.(type) {
	case map[string]any:
		g[key] = updated
	case []any:
		i, _ := arrayIndex(key, len(g))
		g[i] = updated
	}
	return root, nil
}

func lookupTokens(doc any, tokens []string) (any, bool) {
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

func clone(v any) any {
	switch c := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(c))
		for k, x := range c {
			m[k] = clone(x)
		}
		return m
	case []any:
		a := make([]any, len(c))
		for i, x := range c {
			a[i] = clone(x)
		}
		return a
	}
	return v
}

// jsonEqual is structural equality of parsed JSON; numbers compare by value.
func jsonEqual(a, b any) bool {
	switch x := a.(type) {
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for k, v := range x {
			w, ok := y[k]
			if !ok || !jsonEqual(v, w) {
				return false
			}
		}
		return true
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !jsonEqual(x[i], y[i]) {
				return false
			}
		}
		return true
	case json.Number:
		y, ok := b.(json.Number)
		if !ok {
			return false
		}
		rx, okx := new(big.Rat).SetString(string(x))
		ry, oky := new(big.Rat).SetString(string(y))
		return okx && oky && rx.Cmp(ry) == 0
	case string:
		y, ok := b.(string)
		return ok && x == y
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case nil:
		return b == nil
	}
	return false
}

// SNResult is one schema-negative audit outcome.
type SNResult struct {
	ID            string `json:"id"`
	Class         string `json:"class"`
	ExpectedClass string `json:"expected_class"`
	Match         bool   `json:"match"`
}

// PositiveResult is one positive audit control outcome.
type PositiveResult struct {
	ID     string `json:"id"`
	Schema string `json:"schema"`
	Class  string `json:"class"`
}

// DigestResult is one normal-mode digest control outcome.
type DigestResult struct {
	ID        string `json:"id"`
	Class     string `json:"class"`
	Evaluated int    `json:"evaluated"`
}

// SchemaNegativeReport is schema-negative-go.json.
type SchemaNegativeReport struct {
	Validator      string           `json:"validator"`
	Results        []SNResult       `json:"results"`
	Derive         string           `json:"derive"`
	AuditPositive  []PositiveResult `json:"audit_positive"`
	DigestControls []DigestResult   `json:"digest_controls"`

	// Pass and the details below are not serialized; the command reports them separately.
	Pass          bool           `json:"-"`
	DeriveResults []DeriveResult `json:"-"`
	Failures      []string       `json:"-"`
}

// RunSchemaNegative runs SN01-SN07 in audit mode, the positive audit controls, the derivation
// check, and SD01/SD02 in normal mode against scratch schema directories under scratchRoot.
// A wrong class is a failure regardless of any exit status.
func RunSchemaNegative(schemasDir, corpusDir, scratchRoot string) (*SchemaNegativeReport, error) {
	exp, err := LoadExpectations(corpusDir)
	if err != nil {
		return nil, err
	}
	rep := &SchemaNegativeReport{
		Validator:      ValidatorName(),
		Results:        []SNResult{},
		AuditPositive:  []PositiveResult{},
		DigestControls: []DigestResult{},
	}
	snByID := map[string]SNEntry{}
	for _, s := range exp.SchemaNegative {
		snByID[s.ID] = s
		got := AuditFile(filepath.Join(corpusDir, filepath.FromSlash(s.File)), s.Base).Class
		r := SNResult{ID: s.ID, Class: got, ExpectedClass: s.Class, Match: got == s.Class}
		if !r.Match {
			rep.Failures = append(rep.Failures, "class:"+s.ID)
		}
		rep.Results = append(rep.Results, r)
	}
	for _, a := range exp.AuditPositive {
		got := AuditFile(filepath.Join(schemasDir, exp.Schemas[a.Schema].File), a.Schema).Class
		rep.AuditPositive = append(rep.AuditPositive, PositiveResult{ID: a.ID, Schema: a.Schema, Class: got})
		if got != a.Class || got != ClassOK {
			rep.Failures = append(rep.Failures, "class:"+a.ID)
		}
	}
	derived, ok, err := Derive(schemasDir, corpusDir, exp)
	if err != nil {
		ok = false
		rep.Failures = append(rep.Failures, "derive:"+err.(*ClassError).Class)
	}
	rep.DeriveResults = derived
	rep.Derive = "pass"
	if !ok {
		rep.Derive = "fail"
		for _, d := range derived {
			if !d.Match {
				rep.Failures = append(rep.Failures, "derive:"+d.ID+":"+d.Reason)
			}
		}
	}
	for _, d := range exp.DigestControls {
		class, evaluated, err := runDigestControl(schemasDir, corpusDir, scratchRoot, exp, snByID, d)
		if err != nil {
			return nil, err
		}
		rep.DigestControls = append(rep.DigestControls, DigestResult{ID: d.ID, Class: class, Evaluated: evaluated})
		if class != d.Class || evaluated != 0 {
			rep.Failures = append(rep.Failures, "class:"+d.ID)
		}
	}
	if len(exp.SchemaNegative) == 0 || len(exp.AuditPositive) == 0 || len(exp.DigestControls) == 0 {
		rep.Failures = append(rep.Failures, "empty-declarations")
	}
	rep.Pass = len(rep.Failures) == 0
	return rep, nil
}

// runDigestControl writes the canonical schemas into a fresh scratch directory, replaces the
// control's base with its derivation and runs normal mode against the real corpus.
func runDigestControl(schemasDir, corpusDir, scratchRoot string, exp *Expectations, sn map[string]SNEntry, d DigestControl) (string, int, error) {
	dir, err := os.MkdirTemp(scratchRoot, "contractcheck-sd-")
	if err != nil {
		return "", 0, classErr("scratch-unavailable")
	}
	defer os.RemoveAll(dir)
	for _, name := range sortedKeys(exp.Schemas) {
		raw, err := os.ReadFile(filepath.Join(schemasDir, exp.Schemas[name].File))
		if err != nil {
			return "", 0, classErr(ClassSchemaMissing, name)
		}
		if name == d.Base {
			if ref, ok := strings.CutPrefix(d.Derivation, "bytes-of:"); ok {
				raw, err = os.ReadFile(filepath.Join(corpusDir, filepath.FromSlash(sn[ref].File)))
				if err != nil {
					return "", 0, classErr(ClassCorpusInvalid, sn[ref].File)
				}
			} else {
				raw = append(append([]byte{}, raw...), '\n')
			}
		}
		if err := os.WriteFile(filepath.Join(dir, exp.Schemas[name].File), raw, 0o600); err != nil {
			return "", 0, classErr("scratch-unavailable")
		}
	}
	res, _, err := RunCorpus(dir, corpusDir)
	if err != nil {
		var ce *ClassError
		if errors.As(err, &ce) {
			return ce.Class, 0, nil
		}
		return "error", 0, nil
	}
	return ClassOK, len(res.Results), nil
}
