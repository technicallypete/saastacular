package signup

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"core/internal/hostname"
)

// declText returns the source text of the named top-level declaration (type, const, func) of the Go file at path.
func declText(t *testing.T, path, name string) string {
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
				switch sp := s.(type) {
				case *ast.TypeSpec:
					if sp.Name.Name == name {
						n = x
					}
				case *ast.ValueSpec:
					for _, id := range sp.Names {
						if id.Name == name {
							n = x
						}
					}
				}
			}
		}
		if n != nil {
			start := fset.Position(n.Pos()).Offset
			if g, ok := n.(*ast.GenDecl); ok && g.Doc != nil {
				start = fset.Position(g.Doc.Pos()).Offset
			}
			if fd, ok := n.(*ast.FuncDecl); ok && fd.Doc != nil {
				start = fset.Position(fd.Doc.Pos()).Offset
			}
			return string(src[start:fset.Position(n.End()).Offset])
		}
	}
	t.Fatalf("%s: %s not declared", path, name)
	return ""
}

// TestNameRuleCopyIsExact (G062 Plan r5 §2, D-9): signup validates Organization names with a byte-pinned copy of
// tenancy's rule, not an import.
func TestNameRuleCopyIsExact(t *testing.T) {
	for _, name := range []string{"MaxNameLength", "ValidateName"} {
		if a, b := declText(t, "../tenancy/names.go", name), declText(t, "names_copy.go", name); a != b {
			t.Errorf("%s differs from core/internal/tenancy/names.go:\n--- tenancy\n%s\n--- signup\n%s", name, a, b)
		}
	}
}

// TestSlugRuleCopyIsExact: the advisory base and candidate rule is a byte-pinned copy of tenancy's.
func TestSlugRuleCopyIsExact(t *testing.T) {
	for _, name := range []string{"MaxSlugLength", "maxSuffix", "suggestBase", "cut", "slugCandidates"} {
		if a, b := declText(t, "../tenancy/slug.go", name), declText(t, "slug.go", name); a != b {
			t.Errorf("%s differs from core/internal/tenancy/slug.go:\n--- tenancy\n%s\n--- signup\n%s", name, a, b)
		}
	}
}

// TestSignupSlug (SUP-6, unit part): the slug grammar and the reserved labels (floor, www, auth, the CNAME-target and
// auth-host labels under the platform domain, configured labels); nested names and the platform domain never qualify;
// the target origin comes from configuration only.
func TestSignupSlug(t *testing.T) {
	cfg := hostname.Config{BaseDomain: "localhost", CNAMETarget: "tenants.localhost", AuthHost: "auth.localhost", Port: "543",
		ReservedLabels: []string{"status"}}
	for _, s := range []string{"acme", "acme-2", "a", "0rg", strings.Repeat("a", 63)} {
		if !ValidSlug(s, cfg) {
			t.Errorf("ValidSlug(%q) = false", s)
		}
	}
	for _, s := range []string{"", "www", "auth", "tenants", "status", "api", "admin", "sign-in", "Acme", "acme_co", "-acme",
		"acme-", "a--b", "a.b", "acme.localhost", "localhost.", strings.Repeat("a", 64), "acme/x", "%61cme", "ácme"} {
		if ValidSlug(s, cfg) {
			t.Errorf("ValidSlug(%q) = true", s)
		}
	}
	s := NewService(Config{Hostname: cfg, RootOrigin: "https://localhost:543"})
	if got := s.TargetOrigin("acme"); got != "https://acme.localhost:543" {
		t.Errorf("TargetOrigin = %q", got)
	}
	prod := NewService(Config{Hostname: hostname.Config{BaseDomain: "example.com"}, RootOrigin: "https://example.com"})
	if got := prod.TargetOrigin("acme"); got != "https://acme.example.com" {
		t.Errorf("TargetOrigin (no port) = %q", got)
	}
	if got := suggestBase("  Acme & Sons, Ltd.  "); got != "acme-sons-ltd" {
		t.Errorf("suggestBase = %q", got)
	}
	if got := suggestBase("日本"); got != "" {
		t.Errorf("suggestBase(non-ASCII) = %q", got)
	}
	if c := slugCandidates("acme"); len(c) != 99 || c[0] != "acme" || c[1] != "acme-2" || c[98] != "acme-99" {
		t.Errorf("slugCandidates = %d %v", len(c), c[:2])
	}
}
