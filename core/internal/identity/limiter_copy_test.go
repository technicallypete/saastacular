package identity

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"testing"
	"time"
)

// limiterDecl returns the source text of the named top-level type, function or method of the Go file at path.
func limiterDecl(t *testing.T, path, name string) string {
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

// TestLimiterCopyIsExact (G062 Plan r5 §4.5): the root limiters use a byte-pinned copy of the enrollment/api limiter, not
// a shared helper; the type, its constructor, allow and reclaim are identical and the key bound is the same.
func TestLimiterCopyIsExact(t *testing.T) {
	for _, name := range []string{"limiter", "newLimiter", "allow", "reclaim"} {
		if a, b := limiterDecl(t, "../enrollment/api/http.go", name), limiterDecl(t, "limiter_copy.go", name); a != b {
			t.Errorf("%s differs from enrollment/api:\n--- enrollment\n%s\n--- identity\n%s", name, a, b)
		}
	}
	if maxLimiters != 10000 {
		t.Errorf("maxLimiters = %d, want 10000", maxLimiters)
	}
}

// TestLimiterBounds: a depleted bucket is not restored by churning other keys; a full map refuses a new key until a
// bucket is fully refilled.
func TestLimiterBounds(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	l := newLimiter(1, 2)
	l.now = func() time.Time { return now }
	if !l.allow("a") || !l.allow("a") || l.allow("a") {
		t.Fatal("burst 2 not enforced")
	}
	for i := range maxLimiters - 1 {
		l.allow(string(rune(0x10000 + i)))
	}
	if l.allow("new") {
		t.Fatal("a new key was admitted at capacity with no refilled bucket")
	}
	if l.allow("a") {
		t.Fatal("churn restored the depleted bucket")
	}
	now = now.Add(3 * time.Second)
	if !l.allow("new") {
		t.Fatal("a new key was refused after every bucket refilled")
	}
}
