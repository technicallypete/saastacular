package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---- synthetic inputs ----

const testGoMod = `module example.test/app

go 1.27.2

require (
	example.test/direct v1.2.0
	example.test/indirect v0.3.1 // indirect
)
`

const testNPMLock = `{
  "name": "app", "version": "1.0.0", "lockfileVersion": 3, "requires": true,
  "packages": {
    "": {"name": "app", "version": "1.0.0", "dependencies": {"left": "1.0.0"}, "devDependencies": {"devtool": "2.0.0"}},
    "node_modules/left": {"version": "1.0.0", "resolved": "https://registry.npmjs.org/left/-/left-1.0.0.tgz"},
    "node_modules/devtool": {"version": "2.0.0", "dev": true, "resolved": "https://registry.npmjs.org/devtool/-/devtool-2.0.0.tgz"},
    "node_modules/devtool/node_modules/left": {"version": "1.0.0", "dev": true, "resolved": "https://registry.npmjs.org/left/-/left-1.0.0.tgz"},
    "node_modules/@scope/pkg": {"version": "0.1.0", "resolved": "https://registry.npmjs.org/@scope/pkg/-/pkg-0.1.0.tgz"}
  }
}`

const testUVLock = `version = 1
revision = 5

[[package]]
name = "app"
version = "0.1.0"
source = { virtual = "." }
dependencies = [
    { name = "Web_Lib" },
]

[package.dev-dependencies]
dev = [
    { name = "testtool" },
]

[[package]]
name = "helper"
version = "1.1.0"
source = { registry = "https://pypi.org/simple" }

[[package]]
name = "testtool"
version = "9.0.0"
source = { registry = "https://pypi.org/simple" }
dependencies = [
    { name = "helper" },
]

[[package]]
name = "web-lib"
version = "3.0.0"
source = { registry = "https://pypi.org/simple" }
`

type pkg struct {
	eco, name, version string
	groups             []string
	vulns              []map[string]any
	osvGroups          []map[string]any
}

func osvJSON(pkgs ...pkg) []byte {
	var ps []map[string]any
	for _, p := range pkgs {
		m := map[string]any{"package": map[string]any{"name": p.name, "version": p.version, "ecosystem": p.eco}}
		if p.groups != nil {
			m["dependency_groups"] = p.groups
		}
		if p.vulns != nil {
			m["vulnerabilities"] = p.vulns
		}
		if p.osvGroups != nil {
			m["groups"] = p.osvGroups
		}
		ps = append(ps, m)
	}
	b, _ := json.Marshal(map[string]any{"results": []any{map[string]any{"source": map[string]any{"path": "/in/x", "type": "lockfile"}, "packages": ps}}})
	return b
}

// cleanPkgs is the complete scanner side of the three test lockfiles.
func cleanPkgs() []pkg {
	return []pkg{
		{eco: "Go", name: "example.test/direct", version: "1.2.0"}, {eco: "Go", name: "example.test/indirect", version: "0.3.1"},
		{eco: "npm", name: "left", version: "1.0.0"}, {eco: "npm", name: "devtool", version: "2.0.0", groups: []string{"dev"}},
		{eco: "npm", name: "@scope/pkg", version: "0.1.0"},
		{eco: "PyPI", name: "helper", version: "1.1.0"}, {eco: "PyPI", name: "testtool", version: "9.0.0"},
		{eco: "PyPI", name: "web_lib", version: "3.0.0"},
	}
}

func gvJSON(goVersion string, findings ...string) []byte {
	var b bytes.Buffer
	cfg := map[string]any{"config": map[string]any{"protocol_version": "v1.0.0", "scanner_name": "govulncheck", "scanner_version": "v1.8.0",
		"db": "file:///db/govulndb", "db_last_modified": "2026-10-08T22:31:09Z", "go_version": goVersion, "scan_level": "symbol", "scan_mode": "source"}}
	enc := json.NewEncoder(&b)
	enc.SetIndent("", "  ")
	_ = enc.Encode(cfg)
	_ = enc.Encode(map[string]any{"SBOM": map[string]any{"go_version": goVersion, "modules": []any{
		map[string]any{"path": "example.test/app"}, map[string]any{"path": "example.test/direct", "version": "v1.2.0"},
		map[string]any{"path": "stdlib", "version": "v1.27.2"}}}})
	for _, f := range findings {
		b.WriteString(f + "\n")
	}
	return b.Bytes()
}

func gvOSVMsg(id string, aliases ...string) string {
	b, _ := json.Marshal(map[string]any{"osv": map[string]any{"id": id, "aliases": aliases, "summary": "synthetic"}})
	return string(b)
}

func gvFindingMsg(id, module, version, pkgPath, function string) string {
	fr := map[string]any{"module": module, "version": version}
	if pkgPath != "" {
		fr["package"] = pkgPath
	}
	if function != "" {
		fr["function"] = function
	}
	b, _ := json.Marshal(map[string]any{"finding": map[string]any{"osv": id, "fixed_version": "v9.9.9", "trace": []any{fr}}})
	return string(b)
}

const testHeader = exceptionHeader + "\n"

func excRow(eco, name, version, adv, expires string) string {
	return strings.Join([]string{eco, name, version, adv, "synthetic rationale", "platform-owner", expires,
		"a fixed release or a usage change", "absent from production artifacts", "management/v0/results/G012-fixture-review.md"}, "\t") + "\n"
}

type tcase struct {
	osv, gv, exc []byte
	gomod        string
	npm, uv      string
}

func mustEval(t *testing.T, c tcase) report {
	t.Helper()
	var in inputs
	in.Date = time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	if c.exc == nil {
		c.exc = []byte(testHeader)
	}
	var err error
	if in.Exceptions, in.ExcProblem, err = parseExceptions(c.exc); err != nil {
		t.Fatalf("exceptions: %v", err)
	}
	add := func(s string, p func([]byte) (*graph, error)) {
		if s == "" {
			return
		}
		g, err := p([]byte(s))
		if err != nil {
			t.Fatalf("lockfile: %v", err)
		}
		in.Graphs = append(in.Graphs, g)
	}
	add(c.gomod, parseGoMod)
	add(c.npm, parseNPMLock)
	add(c.uv, parseUVLock)
	if c.osv != nil {
		if in.OSV, err = parseOSV(c.osv); err != nil {
			t.Fatalf("osv: %v", err)
		}
	}
	if c.gv != nil {
		if in.Govuln, err = parseGovulncheck(c.gv); err != nil {
			t.Fatalf("govulncheck: %v", err)
		}
	}
	return evaluate(in)
}

func all(osv []byte) tcase {
	return tcase{osv: osv, gv: gvJSON("go1.27.2"), gomod: testGoMod, npm: testNPMLock, uv: testUVLock}
}

func withVuln(base []pkg, eco, name string, id string, aliases []string, maxSev string, label string, fixed string) []pkg {
	out := append([]pkg(nil), base...)
	for i := range out {
		if out[i].eco == eco && out[i].name == name {
			v := map[string]any{"id": id, "aliases": aliases}
			if label != "" {
				v["database_specific"] = map[string]any{"severity": label}
			}
			if fixed != "" {
				v["affected"] = []any{map[string]any{"ranges": []any{map[string]any{"events": []any{map[string]any{"introduced": "0"}, map[string]any{"fixed": fixed}}}}}}
			}
			out[i].vulns = append(out[i].vulns, v)
			out[i].osvGroups = append(out[i].osvGroups, map[string]any{"ids": []string{id}, "aliases": aliases, "max_severity": maxSev})
		}
	}
	return out
}

// ---- tests ----

func TestCleanGraphsPassWithFullReport(t *testing.T) {
	r := mustEval(t, all(osvJSON(cleanPkgs()...)))
	if r.Exit != exitPass {
		t.Fatalf("exit %d: %v %v", r.Exit, r.Incomplete, r.Problems)
	}
	byEco := map[string]graphReport{}
	for _, g := range r.Graphs {
		byEco[g.Ecosystem] = g
	}
	if g := byEco["Go"]; g.Identities != 2 || g.Direct != 1 || g.Transitive != 1 || g.Toolchain != "go1.27.2 standard library" || g.Roots[0] != "example.test/app" {
		t.Errorf("Go graph %+v", g)
	}
	// Repeated npm locations collapse to one identity; the "" root is excluded.
	if g := byEco["npm"]; g.Identities != 3 || g.Locations != 4 || g.Development != 1 || g.Direct != 2 || g.Roots[0] != "app" {
		t.Errorf("npm graph %+v", g)
	}
	// The virtual uv root is excluded; testtool and its helper are development-only.
	if g := byEco["PyPI"]; g.Identities != 3 || g.Development != 2 || g.Direct != 2 || g.Roots[0] != "app" {
		t.Errorf("PyPI graph %+v", g)
	}
}

func TestSeverityClassBoundaries(t *testing.T) {
	for _, c := range []struct {
		score float64
		want  string
	}{{10, classCritical}, {9.0, classCritical}, {8.9, classHigh}, {7.0, classHigh}, {6.9, classMedium}, {4.0, classMedium},
		{3.9, classLow}, {0.1, classLow}, {0, classUnclassified}} {
		if got := severityClass(c.score); got != c.want {
			t.Errorf("score %v: %s, want %s", c.score, got, c.want)
		}
	}
	for _, c := range []struct {
		sev, label, want string
		exit             int
	}{
		{"8.7", "", classHigh, exitPolicyFailure}, {"9.1", "", classCritical, exitPolicyFailure},
		{"5.5", "", classMedium, exitPass}, {"2.0", "", classLow, exitPass},
		{"", "HIGH", classHigh, exitPolicyFailure}, {"", "MODERATE", classMedium, exitPass}, {"", "LOW", classLow, exitPass},
		{"", "", classUnclassified, exitPolicyFailure}, {"not-a-score", "", classUnclassified, exitPolicyFailure},
	} {
		r := mustEval(t, all(osvJSON(withVuln(cleanPkgs(), "npm", "left", "GHSA-2222-3333-4444", []string{"CVE-2026-0001"}, c.sev, c.label, "1.0.1")...)))
		if len(r.Findings) != 1 || r.Findings[0].Class != c.want || r.Exit != c.exit {
			t.Errorf("max_severity %q label %q: findings %+v exit %d, want class %s exit %d", c.sev, c.label, r.Findings, r.Exit, c.want, c.exit)
		}
		if len(r.Findings) == 1 && (r.Findings[0].Fixed[0] != "1.0.1" || r.Findings[0].Groups != nil) {
			t.Errorf("fixed/groups not reported: %+v", r.Findings[0])
		}
	}
}

func TestGoMergeByAliasAndReachability(t *testing.T) {
	// osv-scanner names the advisory by its GHSA, govulncheck by its GO ID; the alias merges them.
	osv := osvJSON(withVuln(cleanPkgs(), "Go", "example.test/direct", "GHSA-2222-3333-5555", []string{"GO-2026-0001"}, "5.0", "", "")...)
	c := all(osv)
	c.gv = gvJSON("go1.27.2", gvOSVMsg("GO-2026-0001", "GHSA-2222-3333-5555"),
		gvFindingMsg("GO-2026-0001", "example.test/direct", "v1.2.0", "example.test/direct/x", "Do"),
		gvFindingMsg("GO-2026-0001", "example.test/direct", "v1.2.0", "example.test/direct/x", ""))
	r := mustEval(t, c)
	if len(r.Findings) != 1 {
		t.Fatalf("findings %+v", r.Findings)
	}
	f := r.Findings[0]
	if f.Class != classMedium || f.Reach != reachCalled || !f.Blocking || len(f.Sources) != 2 || r.Exit != exitPolicyFailure {
		t.Fatalf("a called Go finding blocks whatever its severity: %+v exit %d", f, r.Exit)
	}
	// Imported only, Medium: not blocking.
	c.gv = gvJSON("go1.27.2", gvOSVMsg("GO-2026-0001", "GHSA-2222-3333-5555"),
		gvFindingMsg("GO-2026-0001", "example.test/direct", "v1.2.0", "example.test/direct/x", ""))
	if r := mustEval(t, c); r.Exit != exitPass || r.Findings[0].Reach != reachImported {
		t.Fatalf("imported Medium: %+v exit %d", r.Findings, r.Exit)
	}
}

func TestGoFindingsWithoutSeverityBlockAtEveryLevel(t *testing.T) {
	for _, lvl := range []struct{ pkg, fn, want string }{
		{"net/http", "Serve", reachCalled}, {"net/http", "", reachImported}, {"", "", reachRequired},
	} {
		c := all(osvJSON(cleanPkgs()...))
		c.gv = gvJSON("go1.27.2", gvOSVMsg("GO-2026-6603", "CVE-2026-78659"), gvFindingMsg("GO-2026-6603", "stdlib", "v1.27.2", lvl.pkg, lvl.fn))
		r := mustEval(t, c)
		if len(r.Findings) != 1 || r.Findings[0].Class != classUnclassified || r.Findings[0].Reach != lvl.want || !r.Findings[0].Blocking || r.Exit != exitPolicyFailure {
			t.Errorf("%s: %+v exit %d", lvl.want, r.Findings, r.Exit)
		}
	}
}

func TestExceptions(t *testing.T) {
	osv := osvJSON(withVuln(cleanPkgs(), "npm", "devtool", "GHSA-vfj7-8cjw-p6xm", []string{"CVE-2026-93687"}, "8.7", "HIGH", "")...)
	for _, c := range []struct {
		name string
		rows string
		exit int
		want string
	}{
		{"none", "", exitPolicyFailure, "blocking finding"},
		{"valid exact row", excRow("npm", "devtool", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-29"), exitPass, ""},
		{"valid by alias", excRow("npm", "devtool", "2.0.0", "CVE-2026-93687", "2026-10-15"), exitPass, ""},
		{"expired", excRow("npm", "devtool", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-14"), exitPolicyFailure, "stale"},
		{"unused", excRow("npm", "devtool", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-29") + excRow("npm", "left", "1.0.0", "GHSA-2222-3333-4444", "2026-10-29"), exitPolicyFailure, "unused"},
		{"mismatched version", excRow("npm", "devtool", "2.0.1", "GHSA-vfj7-8cjw-p6xm", "2026-10-29"), exitPolicyFailure, "unused"},
		{"wildcard version", excRow("npm", "devtool", "2.*", "GHSA-vfj7-8cjw-p6xm", "2026-10-29"), exitPolicyFailure, "exact version"},
		{"range version", excRow("npm", "devtool", ">=2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-29"), exitPolicyFailure, "exact version"},
		{"group instead of package", excRow("npm", "dev:*", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-29"), exitPolicyFailure, "one exact name"},
		{"missing field", "npm\tdevtool\t2.0.0\tGHSA-vfj7-8cjw-p6xm\n", exitPolicyFailure, "10 tab-separated"},
		{"empty owner", strings.Replace(excRow("npm", "devtool", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-29"), "platform-owner", " ", 1), exitPolicyFailure, "empty or padded"},
		{"no review record", strings.Replace(excRow("npm", "devtool", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-29"), "management/v0/results/G012-fixture-review.md", "approved", 1), exitPolicyFailure, "review_record"},
		{"bad date", excRow("npm", "devtool", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "29.10.2026"), exitPolicyFailure, "YYYY-MM-DD"},
		{"duplicate", excRow("npm", "devtool", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-29") + excRow("npm", "devtool", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-28"), exitPolicyFailure, "duplicates"},
	} {
		t.Run(c.name, func(t *testing.T) {
			tc := all(osv)
			tc.exc = []byte(testHeader + c.rows)
			r := mustEval(t, tc)
			joined := strings.Join(r.Problems, "\n")
			if r.Exit != c.exit || (c.want != "" && !strings.Contains(joined, c.want)) {
				t.Fatalf("exit %d, problems %q", r.Exit, joined)
			}
		})
	}
	if _, _, err := parseExceptions([]byte("ecosystem\tpackage\n")); err == nil {
		t.Fatal("a wrong header must be refused")
	}
	if _, _, err := parseExceptions(nil); err == nil {
		t.Fatal("an empty exceptions file must be refused")
	}
}

// The same exact row on an identity that is not development-only (here a runtime npm package) must fail: production
// inclusion revokes an npm or PyPI exception.
func TestExceptionRequiresDevelopmentOnlyIdentity(t *testing.T) {
	adv := "GHSA-vfj7-8cjw-p6xm"
	dev := all(osvJSON(withVuln(cleanPkgs(), "npm", "devtool", adv, nil, "8.7", "", "")...))
	dev.exc = []byte(testHeader + excRow("npm", "devtool", "2.0.0", adv, "2026-10-29"))
	if r := mustEval(t, dev); r.Exit != exitPass || !r.Findings[0].Excepted {
		t.Fatalf("development-only identity: exit %d %v", r.Exit, r.Problems)
	}
	prod := all(osvJSON(withVuln(cleanPkgs(), "npm", "left", adv, nil, "8.7", "", "")...))
	prod.exc = []byte(testHeader + excRow("npm", "left", "1.0.0", adv, "2026-10-29"))
	r := mustEval(t, prod)
	joined := strings.Join(r.Problems, "\n")
	if r.Exit != exitPolicyFailure || r.Findings[0].Excepted || !strings.Contains(joined, "not development-only") || !strings.Contains(joined, "blocking finding") {
		t.Fatalf("runtime identity: exit %d excepted %v problems %q", r.Exit, r.Findings[0].Excepted, joined)
	}
	// PyPI: helper is development-only (reached only from the dev group); web-lib is runtime.
	py := all(osvJSON(withVuln(cleanPkgs(), "PyPI", "web_lib", adv, nil, "8.7", "", "")...))
	py.exc = []byte(testHeader + excRow("PyPI", "web-lib", "3.0.0", adv, "2026-10-29"))
	if r := mustEval(t, py); r.Exit != exitPolicyFailure {
		t.Fatalf("runtime PyPI identity: exit %d", r.Exit)
	}
}

// A vulnerability the scanner lists for a package but in none of its groups is an incomplete result.
func TestUngroupedVulnerabilityIsIncomplete(t *testing.T) {
	pk := withVuln(cleanPkgs(), "npm", "left", "GHSA-2222-3333-4444", nil, "5.0", "", "")
	for i := range pk {
		if pk[i].name == "left" {
			pk[i].osvGroups = nil
		}
	}
	if r := mustEval(t, all(osvJSON(pk...))); r.Exit != exitInvalidInput || !strings.Contains(strings.Join(r.Incomplete, "\n"), "in no group") {
		t.Fatalf("exit %d incomplete %v", r.Exit, r.Incomplete)
	}
}

// A read error after a valid prefix (a line over the scanner limit) is invalid input in every line-scanned file, so a
// later invalid or expired row, requirement or package can never hide behind it; the CLI exits 2.
func TestScannerErrorsRejectParsedPrefix(t *testing.T) {
	long := strings.Repeat("x", 70000)
	excWithHiddenExpired := testHeader + excRow("npm", "devtool", "2.0.0", "GHSA-vfj7-8cjw-p6xm", "2026-10-29") +
		"# " + long + "\n" + excRow("npm", "left", "1.0.0", "GHSA-2222-3333-4444", "2026-01-01")
	if _, _, err := parseExceptions([]byte(excWithHiddenExpired)); err == nil || !strings.Contains(err.Error(), "cannot be read") {
		t.Fatalf("exceptions with an over-limit line: %v", err)
	}
	goModLong := testGoMod + "// " + long + "\nrequire example.test/hidden v9.9.9\n"
	if _, err := parseGoMod([]byte(goModLong)); err == nil || !strings.Contains(err.Error(), "cannot be read") {
		t.Fatalf("go.mod with an over-limit line: %v", err)
	}
	uvLong := testUVLock + "# " + strings.Repeat("y", 2<<20) + "\n[[package]]\nname = \"hidden\"\nversion = \"1.0.0\"\nsource = { git = \"https://example.test/h.git\" }\n"
	if _, err := parseUVLock([]byte(uvLong)); err == nil || !strings.Contains(err.Error(), "cannot be read") {
		t.Fatalf("uv.lock with an over-limit line: %v", err)
	}
	dir := t.TempDir()
	write := func(n, s string) string {
		p := filepath.Join(dir, n)
		if err := os.WriteFile(p, []byte(s), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	gm, nl, ul := write("go.mod", testGoMod), write("package-lock.json", testNPMLock), write("uv.lock", testUVLock)
	ex := write("exc.tsv", testHeader)
	osvf := write("osv.json", string(osvJSON(withVuln(cleanPkgs(), "npm", "devtool", "GHSA-vfj7-8cjw-p6xm", nil, "8.7", "", "")...)))
	gv := write("gv.json", string(gvJSON("go1.27.2")))
	for name, args := range map[string][]string{
		"exceptions": {"--exceptions", write("exc-long.tsv", excWithHiddenExpired), "--gomod", gm, "--npm-lock", nl, "--uv-lock", ul},
		"go.mod":     {"--exceptions", ex, "--gomod", write("go-long.mod", goModLong), "--npm-lock", nl, "--uv-lock", ul},
		"uv.lock":    {"--exceptions", ex, "--gomod", gm, "--npm-lock", nl, "--uv-lock", write("uv-long.lock", uvLong)},
	} {
		var out, errb bytes.Buffer
		full := append([]string{"evaluate", "--date", "2026-10-15", "--osv", osvf, "--govulncheck", gv}, args...)
		if code := run(full, &out, &errb); code != exitInvalidInput {
			t.Errorf("%s with an over-limit line: CLI exit %d, want 2 (%s)", name, code, errb.String())
		}
	}
}

// Group and advisory references must resolve both ways with non-empty IDs, and the output must be exactly one JSON
// document (trailing whitespace allowed).
func TestOSVReferencesAndSingleDocument(t *testing.T) {
	// A group ID without its advisory record (the group would otherwise carry its Low score and pass).
	pk := withVuln(cleanPkgs(), "npm", "left", "GHSA-2222-3333-4444", nil, "2.0", "", "")
	for i := range pk {
		if pk[i].name == "left" {
			pk[i].vulns = nil
		}
	}
	if r := mustEval(t, all(osvJSON(pk...))); r.Exit != exitInvalidInput || !strings.Contains(strings.Join(r.Incomplete, "\n"), "has no advisory record") {
		t.Fatalf("group without record: exit %d %v", r.Exit, r.Incomplete)
	}
	// An empty advisory ID in both the record and the group.
	pk = withVuln(cleanPkgs(), "npm", "left", "", nil, "2.0", "", "")
	r := mustEval(t, all(osvJSON(pk...)))
	if joined := strings.Join(r.Incomplete, "\n"); r.Exit != exitInvalidInput || !strings.Contains(joined, "without an id") || !strings.Contains(joined, "has no advisory record") {
		t.Fatalf("empty advisory id: exit %d %v", r.Exit, r.Incomplete)
	}
	clean := osvJSON(cleanPkgs()...)
	for name, tail := range map[string]string{"closing brace": "}", "closing bracket": "]", "second document": "{}"} {
		if _, err := parseOSV(append(append([]byte(nil), clean...), []byte("\n"+tail)...)); err == nil {
			t.Errorf("trailing %s: want a parse error", name)
		}
	}
	if _, err := parseOSV(append(append([]byte(nil), clean...), []byte("\n \t\n")...)); err != nil {
		t.Fatalf("trailing whitespace must stay valid: %v", err)
	}
	// The CLI returns 2 for a trailing delimiter.
	dir := t.TempDir()
	p := filepath.Join(dir, "osv.json")
	_ = os.WriteFile(p, append(append([]byte(nil), clean...), ']'), 0o600)
	ex := filepath.Join(dir, "exc.tsv")
	_ = os.WriteFile(ex, []byte(testHeader), 0o600)
	nl := filepath.Join(dir, "package-lock.json")
	_ = os.WriteFile(nl, []byte(testNPMLock), 0o600)
	var out, errb bytes.Buffer
	if code := run([]string{"evaluate", "--date", "2026-10-15", "--exceptions", ex, "--osv", p, "--npm-lock", nl}, &out, &errb); code != exitInvalidInput {
		t.Fatalf("trailing bracket: CLI exit %d", code)
	}
}

func TestIdentityReconciliation(t *testing.T) {
	base := cleanPkgs()
	drop := func(name string) []pkg {
		var out []pkg
		for _, p := range base {
			if p.name != name {
				out = append(out, p)
			}
		}
		return out
	}
	stdlibOK := append(append([]pkg(nil), base...), pkg{eco: "Go", name: "stdlib", version: "1.27.2"})
	rootLogged := append(append([]pkg(nil), base...), pkg{eco: "npm", name: "app", version: "1.0.0"})
	for _, c := range []struct {
		name string
		pkgs []pkg
		exit int
		want string
	}{
		{"stdlib at the go directive is allowed", stdlibOK, exitPass, ""},
		{"a reported project root is logged, not counted", rootLogged, exitPass, ""},
		{"omission", drop("helper"), exitInvalidInput, "missing from the scanner result: PyPI:helper@1.1.0"},
		{"equal-count omission plus duplication", append(drop("helper"), pkg{eco: "PyPI", name: "web-lib", version: "3.0.0"}), exitInvalidInput, "PyPI:helper@1.1.0"},
		{"equal-count substitution", append(drop("left"), pkg{eco: "npm", name: "left", version: "1.0.1"}), exitInvalidInput, "absent from the lockfile: npm:left@1.0.1"},
		{"stdlib at another version", append(append([]pkg(nil), base...), pkg{eco: "Go", name: "stdlib", version: "1.27.1"}), exitInvalidInput, "Go:stdlib@1.27.1"},
		{"unknown ecosystem", append(append([]pkg(nil), base...), pkg{eco: "crates.io", name: "x", version: "1"}), exitInvalidInput, "crates.io"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := mustEval(t, all(osvJSON(c.pkgs...)))
			joined := strings.Join(r.Incomplete, "\n")
			if r.Exit != c.exit || (c.want != "" && !strings.Contains(joined, c.want)) {
				t.Fatalf("exit %d, incomplete %q", r.Exit, joined)
			}
		})
	}
	// Totals equal, identities differ: the result must still be incomplete.
	r := mustEval(t, all(osvJSON(append(drop("helper"), pkg{eco: "PyPI", name: "other", version: "1.1.0"})...)))
	if r.Exit != exitInvalidInput {
		t.Fatal("equal totals with a substituted identity must be incomplete")
	}
}

func TestUnsupportedRecordsAreIncomplete(t *testing.T) {
	link := strings.Replace(testNPMLock, `"node_modules/@scope/pkg": {"version": "0.1.0", "resolved": "https://registry.npmjs.org/@scope/pkg/-/pkg-0.1.0.tgz"}`,
		`"node_modules/@scope/pkg": {"resolved": "packages/pkg", "link": true}`, 1)
	git := testUVLock + "\n[[package]]\nname = \"forked\"\nversion = \"1.0.0\"\nsource = { git = \"https://example.test/forked.git\" }\n"
	repl := testGoMod + "\nreplace example.test/direct => ../direct\n"
	for name, c := range map[string]tcase{
		"npm link":   {osv: osvJSON(cleanPkgs()...), gv: gvJSON("go1.27.2"), gomod: testGoMod, npm: link, uv: testUVLock},
		"uv git":     {osv: osvJSON(cleanPkgs()...), gv: gvJSON("go1.27.2"), gomod: testGoMod, npm: testNPMLock, uv: git},
		"go replace": {osv: osvJSON(cleanPkgs()...), gv: gvJSON("go1.27.2"), gomod: repl, npm: testNPMLock, uv: testUVLock},
	} {
		if r := mustEval(t, c); r.Exit != exitInvalidInput || !strings.Contains(strings.Join(r.Incomplete, "\n"), unsupportedPrefixG) {
			t.Errorf("%s: exit %d, incomplete %v", name, r.Exit, r.Incomplete)
		}
	}
}

func TestGovulncheckContract(t *testing.T) {
	good := gvJSON("go1.27.2")
	for name, gv := range map[string][]byte{
		"toolchain differs from the go directive": gvJSON("go1.27.1"),
		"scanner version":                         bytes.Replace(good, []byte(`"v1.8.0"`), []byte(`"v1.9.0"`), 1),
		"protocol":                                bytes.Replace(good, []byte(`"v1.0.0"`), []byte(`"v0.1.0"`), 1),
		"scan level":                              bytes.Replace(good, []byte(`"symbol"`), []byte(`"module"`), 1),
		"online database":                         bytes.Replace(good, []byte(`"file:///db/govulndb"`), []byte(`"https://vuln.go.dev"`), 1),
		"SBOM module outside go.mod":              bytes.Replace(good, []byte(`"example.test/direct"`), []byte(`"example.test/other"`), 1),
	} {
		c := all(osvJSON(cleanPkgs()...))
		c.gv = gv
		if r := mustEval(t, c); r.Exit != exitInvalidInput {
			t.Errorf("%s: exit %d", name, r.Exit)
		}
	}
	for name, b := range map[string]string{
		"no config":   `{"SBOM": {"modules": []}}`,
		"no SBOM":     string(good[:bytes.Index(good, []byte(`{`+"\n"+`  "SBOM"`))]),
		"truncated":   string(good[:len(good)-10]),
		"bad finding": string(good) + `{"finding": {"osv": "GO-1", "trace": []}}`,
	} {
		if _, err := parseGovulncheck([]byte(b)); err == nil {
			t.Errorf("%s: want a parse error", name)
		}
	}
}

func TestMalformedOSVAndExitCodes(t *testing.T) {
	for name, b := range map[string]string{"truncated": `{"results": [`, "trailing": `{"results": []} {}`, "no results": `{}`, "not json": `x`} {
		if _, err := parseOSV([]byte(b)); err == nil {
			t.Errorf("%s: want a parse error", name)
		}
	}
	dir := t.TempDir()
	write := func(n string, b []byte) string {
		p := filepath.Join(dir, n)
		if err := os.WriteFile(p, b, 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	gm, nl, ul := write("go.mod", []byte(testGoMod)), write("package-lock.json", []byte(testNPMLock)), write("uv.lock", []byte(testUVLock))
	ex := write("exc.tsv", []byte(testHeader))
	clean := write("osv.json", osvJSON(cleanPkgs()...))
	high := write("osv-high.json", osvJSON(withVuln(cleanPkgs(), "npm", "left", "GHSA-2222-3333-4444", nil, "7.5", "", "")...))
	gv := write("gv.json", gvJSON("go1.27.2"))
	bad := write("bad.json", []byte(`{"results": [`))
	common := []string{"--date", "2026-10-15", "--exceptions", ex, "--gomod", gm, "--npm-lock", nl, "--uv-lock", ul, "--govulncheck", gv}
	var out, errb bytes.Buffer
	rep := filepath.Join(dir, "report.json")
	for _, c := range []struct {
		args []string
		want int
	}{
		{append([]string{"evaluate", "--osv", clean, "--report", rep}, common...), exitPass},
		{append([]string{"evaluate", "--osv", high}, common...), exitPolicyFailure},
		{append([]string{"evaluate", "--osv", bad}, common...), exitInvalidInput},
		{[]string{"evaluate", "--date", "2026-10-15", "--exceptions", ex, "--npm-lock", nl}, exitInvalidInput},
		{[]string{"evaluate", "--date", "15/10/2026", "--exceptions", ex, "--npm-lock", nl, "--osv", clean}, exitInvalidInput},
		{[]string{"evaluate", "--date", "2026-10-15", "--npm-lock", nl, "--osv", clean}, exitInvalidInput},
		{[]string{"nonsense"}, exitInvalidInput},
	} {
		out.Reset()
		errb.Reset()
		if got := run(c.args, &out, &errb); got != c.want {
			t.Errorf("%v: exit %d, want %d (%s %s)", c.args[:2], got, c.want, out.String(), errb.String())
		}
	}
	var r report
	b, err := os.ReadFile(rep)
	if err != nil || json.Unmarshal(b, &r) != nil || r.Result != "pass" || len(r.Graphs) != 3 {
		t.Fatalf("report: %v %+v", err, r)
	}
	// Incomplete dominates a policy failure.
	tc := all(osvJSON(withVuln(cleanPkgs()[1:], "npm", "left", "GHSA-2222-3333-4444", nil, "9.5", "", "")...))
	if r := mustEval(t, tc); r.Exit != exitInvalidInput || len(r.Problems) == 0 {
		t.Fatalf("incomplete and blocking: exit %d", r.Exit)
	}
}

func TestFixtureDB(t *testing.T) {
	dir := t.TempDir()
	rec := filepath.Join(dir, "r.json")
	if err := os.WriteFile(rec, []byte(`{"id": "G012-SYN-0001", "affected": []}`), 0o600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "all.zip")
	var errb bytes.Buffer
	if code := run([]string{"fixture-db", "--out", out, rec}, &errb, &errb); code != exitPass {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	zr, err := zip.OpenReader(out)
	if err != nil || len(zr.File) != 1 || zr.File[0].Name != "G012-SYN-0001.json" {
		t.Fatalf("zip: %v", err)
	}
	zr.Close()
	bad := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(bad, []byte(`{"summary": "no id"}`), 0o600)
	if code := run([]string{"fixture-db", "--out", out, bad}, &errb, &errb); code != exitInvalidInput {
		t.Fatal("a record without an id must be refused")
	}
}
