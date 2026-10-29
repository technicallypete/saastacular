package gotest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Fixture streams for the converter's unit tests and for the harness control T-6 (`gotestreport selftest`), which
// runs them through the binary in the validation tools image. The passing stream has the shape of B-IDENTITY-IT
// (Plan r2 §4 #53): two test packages, interleaved, and one package without test files.

const (
	fxPI = "core/internal/identity"
	fxPW = "core/internal/identity/workos"
	fxPN = "core/internal/identity/identitytest"
)

// Fixture is one stream with its expected outcome.
type Fixture struct {
	Name        string
	Stream      string
	RC          int
	Interrupted bool
	Status      string // expected Result.Status
	Finding     string // a finding the result must contain ("" for a passing fixture)
}

func fxEv(action, pkg, test, out string, el ...float64) string {
	e := struct {
		Action  string   `json:"Action"`
		Package string   `json:"Package"`
		Test    string   `json:"Test,omitempty"`
		Elapsed *float64 `json:"Elapsed,omitempty"`
		Output  string   `json:"Output,omitempty"`
	}{Action: action, Package: pkg, Test: test, Output: out}
	if len(el) > 0 {
		e.Elapsed = &el[0]
	}
	b, _ := json.Marshal(e)
	return string(b) + "\n"
}

func fxRun(pkg, test string) string {
	return fxEv("run", pkg, test, "") + fxEv("output", pkg, test, "=== RUN   "+test+"\n")
}

func fxEnd(action, pkg, test, dur string, el float64) string {
	return fxEv("output", pkg, test, "--- "+strings.ToUpper(action)+": "+test+" ("+dur+")\n") + fxEv(action, pkg, test, "", el)
}

func fxPkgEnd(pkg, dur string, el float64) string {
	return fxEv("output", pkg, "", "PASS\n") + fxEv("output", pkg, "", "ok  \t"+pkg+"\t"+dur+"\n") + fxEv("pass", pkg, "", "", el)
}

// fxList is the expectation block for the given package -> tests ("-" for a package without test files).
func fxList(rc int, pkgs ...[]string) string {
	var b strings.Builder
	for _, p := range pkgs {
		b.WriteString("GO-IT-LIST " + fxEv("start", p[0], "", ""))
	}
	for _, p := range pkgs {
		if len(p) > 1 && p[1] == "-" {
			b.WriteString("GO-IT-LIST " + fxEv("output", p[0], "", "?   \t"+p[0]+"\t[no test files]\n"))
			b.WriteString("GO-IT-LIST " + fxEv("skip", p[0], "", "", 0))
			continue
		}
		for _, t := range p[1:] {
			b.WriteString("GO-IT-LIST " + fxEv("output", p[0], "", t+"\n"))
		}
		b.WriteString("GO-IT-LIST " + fxEv("output", p[0], "", "ok  \t"+p[0]+"\t0.001s\n"))
		b.WriteString("GO-IT-LIST " + fxEv("pass", p[0], "", "", 0.001))
	}
	fmt.Fprintf(&b, "GO-IT-LIST-EXIT %d\n", rc)
	return b.String()
}

var fxListOK = fxList(0, []string{fxPI, "TestA", "TestB"}, []string{fxPW, "TestW"}, []string{fxPN, "-"})

// Stream parts of the passing fixture, so negatives differ from it in exactly one place.
var (
	fxHead = fxEv("start", fxPI, "", "") + fxEv("start", fxPW, "", "") + fxEv("start", fxPN, "", "") +
		fxEv("output", fxPN, "", "?   \t"+fxPN+"\t[no test files]\n") + fxEv("skip", fxPN, "", "", 0)
	fxTestA = fxRun(fxPI, "TestA") + fxRun(fxPI, "TestA/one") + fxEv("output", fxPI, "TestA/one", "    a_test.go:9: hello\n") +
		fxRun(fxPI, "TestA/one/deep") + fxEnd("pass", fxPI, "TestA/one/deep", "0.00s", 0) +
		fxEnd("pass", fxPI, "TestA/one", "0.01s", 0.01) + fxRun(fxPI, "TestA/two") + fxEnd("pass", fxPI, "TestA/two", "0.00s", 0) +
		fxEv("output", fxPI, "TestA", "E-PG-1 {\"x\":1}\n") + fxEnd("pass", fxPI, "TestA", "0.02s", 0.02)
	fxTestW = fxRun(fxPW, "TestW") + fxEnd("pass", fxPW, "TestW", "0.00s", 0)
	fxTestB = fxRun(fxPI, "TestB") + fxEnd("pass", fxPI, "TestB", "0.00s", 0)
	fxEndPI = fxPkgEnd(fxPI, "0.030s", 0.03)
	fxEndPW = fxPkgEnd(fxPW, "0.010s", 0.01)
	fxTail  = "RLS-ACCESSLOG {\"event\":\"access.decision\"}\n"
)

func fxPass(testW, endPW, list string) string {
	return fxHead + fxTestA + testW + fxTestB + fxEndPI + endPW + fxTail + list
}

// PassingText is the exact readable text of the passing fixture.
const PassingText = `=== RUN   TestA
=== RUN   TestA/one
    a_test.go:9: hello
=== RUN   TestA/one/deep
=== RUN   TestA/two
E-PG-1 {"x":1}
--- PASS: TestA (0.02s)
    --- PASS: TestA/one (0.01s)
        --- PASS: TestA/one/deep (0.00s)
    --- PASS: TestA/two (0.00s)
=== RUN   TestB
--- PASS: TestB (0.00s)
PASS
ok  	core/internal/identity	0.030s
=== RUN   TestW
--- PASS: TestW (0.00s)
PASS
ok  	core/internal/identity/workos	0.010s
?   	core/internal/identity/identitytest	[no test files]
RLS-ACCESSLOG {"event":"access.decision"}
GO-IT-EXPECTED core/internal/identity	TestA
GO-IT-EXPECTED core/internal/identity	TestB
GO-IT-EXPECTED core/internal/identity/workos	TestW
GO-IT-EXPECTED core/internal/identity/identitytest	-
GO-IT-LIST-EXIT 0
`

// Fixtures returns the passing fixture first, then one fixture per fail-closed class of Plan r2 §5.2 / §6 T-6.
func Fixtures() []Fixture {
	pass := fxPass(fxTestW, fxEndPW, fxListOK)
	wPass := fxEnd("pass", fxPW, "TestW", "0.00s", 0)
	return []Fixture{
		{Name: "pass-multi-package", Stream: pass, Status: "pass"},
		{Name: "truncated", Stream: fxHead + fxTestA + fxRun(fxPW, "TestW") + `{"Action":"output","Package":"core/inte`, Status: "fail", Finding: "truncated-final-line"},
		{Name: "skip", Stream: fxPass(fxRun(fxPW, "TestW")+fxEnd("skip", fxPW, "TestW", "0.00s", 0), fxEndPW, fxListOK), Status: "fail", Finding: "skipped:" + fxPW + ".TestW"},
		{Name: "zero-tests", Stream: fxPass("", fxPkgEnd(fxPW, "0.010s", 0.01), fxListOK), Status: "fail", Finding: "zero-tests:" + fxPW},
		{Name: "build-fail", Stream: fxPass("", fxEv("output", fxPW, "", "FAIL\t"+fxPW+" [build failed]\n")+fxEv("fail", fxPW, "", "", 0), fxListOK), Status: "fail", Finding: "build-failed:" + fxPW},
		{Name: "malformed-line", Stream: fxHead + "{\"Action\":\"output\",\"Pack\n" + fxTestA + fxTestW + fxTestB + fxEndPI + fxEndPW + fxListOK, Status: "fail", Finding: "malformed-json-line"},
		{Name: "pass-with-rc-1", Stream: pass, RC: 1, Status: "fail", Finding: "package-pass-with-exit:" + fxPI},
		{Name: "omitted-expected", Stream: fxPass(fxTestW, fxEndPW, fxList(0, []string{fxPI, "TestA", "TestB"}, []string{fxPW, "TestW", "TestW2"}, []string{fxPN, "-"})), Status: "fail", Finding: "omitted:" + fxPW + ".TestW2"},
		{Name: "unfinished", Stream: fxPass(fxTestW+fxRun(fxPW, "TestW/late"), fxEndPW, fxListOK), Status: "fail", Finding: "unfinished:" + fxPW + ".TestW/late"},
		{Name: "duplicate-terminal", Stream: fxPass(fxTestW+fxEv("pass", fxPW, "TestW", "", 0), fxEndPW, fxListOK), Status: "fail", Finding: "duplicate-terminal:" + fxPW + ".TestW"},
		{Name: "contradictory-terminal", Stream: fxPass(fxTestW+fxEv("fail", fxPW, "TestW", "", 0), fxEndPW, fxListOK), Status: "fail", Finding: "contradictory:" + fxPW + ".TestW"},
		{Name: "undeclared-test", Stream: fxPass(fxTestW+fxRun(fxPI, "TestExtra")+fxEnd("pass", fxPI, "TestExtra", "0.00s", 0), fxEndPW, fxListOK), Status: "fail", Finding: "undeclared-test:" + fxPI + ".TestExtra"},
		{Name: "second-package-unterminated", Stream: fxPass(fxTestW, fxEv("output", fxPW, "", "PASS\n"), fxListOK), Status: "fail", Finding: "package-unterminated:" + fxPW},
		{Name: "wrong-package", Stream: fxPass(fxRun(fxPI, "TestW")+fxEnd("pass", fxPI, "TestW", "0.00s", 0), fxEndPW, fxListOK), Status: "fail", Finding: "undeclared-test:" + fxPI + ".TestW"},
		{Name: "package-absent", Stream: fxEv("start", fxPI, "", "") + fxEv("start", fxPW, "", "") + fxTestA + fxTestW + fxTestB + fxEndPI + fxEndPW + fxListOK, Status: "fail", Finding: "package-absent:" + fxPN},
		{Name: "expectation-missing", Stream: fxPass(fxTestW, fxEndPW, ""), Status: "fail", Finding: "expectation-missing"},
		{Name: "expectation-exit", Stream: fxPass(fxTestW, fxEndPW, fxList(1, []string{fxPI, "TestA", "TestB"}, []string{fxPW, "TestW"}, []string{fxPN, "-"})), Status: "fail", Finding: "expectation-exit:1"},
		{Name: "expectation-empty", Stream: fxPass(fxTestW, fxEndPW, fxList(0, []string{fxPI, "TestA", "TestB"}, []string{fxPW}, []string{fxPN, "-"})), Status: "fail", Finding: "expectation-empty:" + fxPW},
		{Name: "failing-test", Stream: fxPass(fxRun(fxPW, "TestW")+fxEnd("fail", fxPW, "TestW", "0.00s", 0), fxEv("output", fxPW, "", "FAIL\n")+fxEv("fail", fxPW, "", "", 0.01), fxListOK), RC: 1, Status: "fail", Finding: "package-failed:" + fxPW},
		{Name: "interrupted", Stream: fxHead + fxTestA + fxRun(fxPW, "TestW") + fxListOK, Interrupted: true, RC: 137, Status: "interrupted", Finding: "interrupted"},
		{Name: "no-test-files-not-expected", Stream: fxPass(fxTestW, fxEndPW, fxList(0, []string{fxPI, "TestA", "TestB"}, []string{fxPW, "TestW"}, []string{fxPN, "TestN"})), Status: "fail", Finding: "package-skipped:" + fxPN},
		{Name: "duplicate-package-terminal", Stream: fxPass(fxTestW, fxEndPW+fxEv("pass", fxPW, "", "", 0), fxListOK), Status: "fail", Finding: "package-duplicate-terminal:" + fxPW},
		{Name: "undeclared-package", Stream: fxPass(fxTestW, fxEndPW, fxListOK) + fxEv("start", "core/internal/other", "", "") + fxEv("pass", "core/internal/other", "", "", 0), Status: "fail", Finding: "undeclared-package:core/internal/other"},
		{Name: "pass-subtest-skip", Stream: fxPass(fxRun(fxPW, "TestW")+fxRun(fxPW, "TestW/s")+fxEnd("skip", fxPW, "TestW/s", "0.00s", 0)+wPass, fxEndPW, fxListOK), Status: "fail", Finding: "skipped:" + fxPW + ".TestW/s"},
	}
}

// Check runs one fixture and returns "" when its outcome is the expected one.
func (f Fixture) Check() string {
	r, err := Convert(strings.NewReader(f.Stream), Options{Name: "fx-" + f.Name, RC: f.RC, Interrupted: f.Interrupted})
	if err != nil {
		return "convert: " + err.Error()
	}
	if r.Status != f.Status {
		return fmt.Sprintf("status %s, want %s (%s)", r.Status, f.Status, strings.Join(r.Findings, ","))
	}
	if f.Finding != "" {
		for _, x := range r.Findings {
			if x == f.Finding {
				return ""
			}
		}
		return fmt.Sprintf("finding %s missing (%s)", f.Finding, strings.Join(r.Findings, ","))
	}
	if !bytes.Equal(r.Text, []byte(PassingText)) {
		return "reconstructed text differs from PassingText"
	}
	want := map[string]string{fxPI: "pass", fxPW: "pass", fxPN: "skip-no-test-files"}
	got := map[string]string{}
	for _, l := range strings.Split(strings.TrimSuffix(string(r.Rows), "\n"), "\n") {
		c := strings.Split(l, "\t")
		if len(c) == 15 && c[0] == "package" {
			got[c[2]] = c[4] + "/" + c[6] + "/" + c[7]
		}
	}
	for p, term := range want {
		exp := map[string]string{fxPI: "2/2", fxPW: "1/1", fxPN: "0/0"}[p]
		if got[p] != term+"/"+exp {
			return fmt.Sprintf("package row %s = %q, want %s/%s", p, got[p], term, exp)
		}
	}
	return ""
}
