package gotest

import (
	"strings"
	"testing"
)

// Every fixture (the passing multi-package stream with exact text and per-package rows, and one stream per
// fail-closed class) yields its expected outcome.
func TestFixtures(t *testing.T) {
	fx := Fixtures()
	if len(fx) < 20 || fx[0].Status != "pass" {
		t.Fatalf("fixture table: %d entries, first %q", len(fx), fx[0].Status)
	}
	for _, f := range fx {
		t.Run(f.Name, func(t *testing.T) {
			if msg := f.Check(); msg != "" {
				t.Fatal(msg)
			}
		})
	}
}

func rowsOf(t *testing.T, r *Result, kind string) [][]string {
	t.Helper()
	var out [][]string
	for _, l := range strings.Split(strings.TrimSuffix(string(r.Rows), "\n"), "\n") {
		c := strings.Split(l, "\t")
		if len(c) != len(strings.Split(Header, "\t")) {
			t.Fatalf("row has %d columns: %q", len(c), l)
		}
		if c[0] == kind {
			out = append(out, c)
		}
	}
	return out
}

// Per-test rows: status, elapsed capture, parent and the line range of each test in the reconstructed text;
// provenance pass-through and per-package counts in the invocation row.
func TestRowsAndProvenance(t *testing.T) {
	r, err := Convert(strings.NewReader(Fixtures()[0].Stream), Options{Name: "identity-it", Provenance: [][2]string{
		{"run_id", "r1"}, {"tested_tree", "abc"}, {"command", "sh -c 'go test; exit $rc'\n"}}})
	if err != nil || r.Status != "pass" {
		t.Fatalf("status %v %v", r, err)
	}
	inv := rowsOf(t, r, "invocation")
	if len(inv) != 1 || inv[0][4] != "pass" || inv[0][6] != "3" || inv[0][7] != "3" || inv[0][11] != "go-tests/identity-it.log" {
		t.Fatalf("invocation row %q", inv)
	}
	for _, want := range []string{"run_id=r1", "tested_tree=abc", "command=sh -c 'go test%3B exit $rc'%0A",
		"expected_by_package=core/internal/identity:2,core/internal/identity/workos:1,core/internal/identity/identitytest:0"} {
		if !strings.Contains(inv[0][14], want) {
			t.Errorf("detail lacks %q: %s", want, inv[0][14])
		}
	}
	lines := strings.Split(string(r.Text), "\n")
	tests := map[string][]string{}
	for _, c := range rowsOf(t, r, "test") {
		tests[c[3]] = c
	}
	if len(tests) != 6 {
		t.Fatalf("test rows %d", len(tests))
	}
	deep := tests["TestA/one/deep"]
	if deep[4] != "pass" || deep[10] != "0.000" || deep[14] != "parent=TestA/one" {
		t.Fatalf("deep row %q", deep)
	}
	if one := tests["TestA/one"]; one[10] != "0.010" {
		t.Fatalf("elapsed %q", one)
	}
	// The line range of TestA/one/deep covers its RUN line and its re-laid-out result line.
	from, to := atoi(deep[12]), atoi(deep[13])
	if !strings.HasPrefix(lines[from-1], "=== RUN   TestA/one/deep") || lines[to-1] != "        --- PASS: TestA/one/deep (0.00s)" {
		t.Fatalf("range %d-%d: %q .. %q", from, to, lines[from-1], lines[to-1])
	}
	if !strings.Contains(string(r.JUnit), `<testcase classname="core/internal/identity" name="TestA/one" time="0.010"/>`) ||
		strings.Contains(string(r.JUnit), "><![CDATA[") {
		t.Fatalf("junit %s", r.JUnit)
	}
}

// An interrupted invocation keeps every terminal row observed so far and marks the started test unfinished.
func TestInterruptedRows(t *testing.T) {
	var f Fixture
	for _, x := range Fixtures() {
		if x.Name == "interrupted" {
			f = x
		}
	}
	r, err := Convert(strings.NewReader(f.Stream), Options{Name: "n", RC: 137, Interrupted: true})
	if err != nil || r.Status != "interrupted" {
		t.Fatalf("%v %v", r, err)
	}
	got := map[string]string{}
	for _, c := range rowsOf(t, r, "test") {
		got[c[3]] = c[4]
	}
	if got["TestA"] != "pass" || got["TestW"] != "unfinished" || got["TestB"] != "omitted" {
		t.Fatalf("rows %v", got)
	}
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
