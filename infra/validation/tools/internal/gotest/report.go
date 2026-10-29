// Package gotest converts one harness Go test invocation (G064 Plan r2 §5.1-§5.2) into supplemental per-test evidence.
//
// The input is the invocation's saved stream: the `go test -json` events of the real run, interleaved with non-JSON
// lines (merged stderr, RLS-ACCESSLOG records, harness lines), followed by the expectation block the container prints
// after the run: one `GO-IT-LIST <event>` line per event of `go test -json -list RUN PACKAGE` (the expected top-level
// Test* names of the bound source, keyed by the event's Package) and one `GO-IT-LIST-EXIT <status>` line.
//
// Convert reconstructs the readable `-v` text (results laid out as `go test -v` prints them: a subtest's result line
// follows its parent's, indented four spaces per level, and each package's text is contiguous; non-JSON lines are kept
// verbatim in place; the expectation block is rendered as `GO-IT-EXPECTED package<TAB>test` lines; the `=== NAME`
// attribution lines that test2json consumes are not re-synthesized), one TSV row per invocation, per selected package
// and per test or subtest, and a JUnit fragment. It fails closed (Status "fail") on every completeness class of §5.2:
// a selected package without exactly one terminal event, a build failure, a package with expected tests but no run,
// an omitted expected test, an unfinished test, duplicate or contradictory terminals, a skipped test, an undeclared
// test or package, a malformed JSON line, a pass with a non-zero exit status, an empty or failed expectation and a
// selected package absent from the events.
package gotest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Header is the column header of the per-invocation and aggregated TSV rows.
const Header = "kind\tinvocation\tpackage\ttest\tstatus\trc\texpected\tpassed\tfailed\tunfinished\telapsed_s\tlog_file\tline_from\tline_to\tdetail"

// Event is one cmd/test2json event (build events carry ImportPath instead of Package).
type Event struct {
	Action      string   `json:"Action"`
	Package     string   `json:"Package"`
	ImportPath  string   `json:"ImportPath"`
	Test        string   `json:"Test"`
	Elapsed     *float64 `json:"Elapsed"`
	Output      string   `json:"Output"`
	FailedBuild string   `json:"FailedBuild"`
}

// Options are the harness-side facts of one invocation.
type Options struct {
	Name        string      // invocation name
	RC          int         // the container's exit status, captured before any conversion
	Interrupted bool        // finalized from the EXIT trap: the invocation did not end
	LogFile     string      // the published readable log the rows reference (go-tests/<Name>.log)
	Provenance  [][2]string // ordered key=value facts written into the invocation row
}

// Result is the converted invocation.
type Result struct {
	Text     []byte
	Rows     []byte // TSV rows without the header
	JUnit    []byte
	Status   string // pass, fail or interrupted
	Findings []string
}

var testName = regexp.MustCompile(`^Test[\p{L}\p{N}_]*$`)

type chunk struct {
	text string
	pkg  string
	test string
}

type testState struct {
	pkg, name, parent string
	order             int
	started           bool
	terminals         []string
	elapsed           *float64
	from, to          int
}

type pkgState struct {
	name      string
	chunks    []chunk
	buf       map[string][]string // buffered result lines per parent test (go test -v layout)
	bufOrder  []string
	terminals []string
	elapsed   *float64
	noFiles   bool
	buildFail bool
	from, to  int
}

type segment struct {
	pkg      string // a package's contiguous text
	text     string // a verbatim line
	expected bool   // the rendered expectation block
}

type converter struct {
	opt      Options
	findings []string
	seen     map[string]bool
	segs     []segment
	pkgs     map[string]*pkgState
	tests    map[string]*testState
	norder   int
	running  map[string]map[string]bool // package -> known test names
	list     []Event
	listRC   *int
	listSeen bool
	malform  int
}

func (c *converter) find(f string) {
	if !c.seen[f] {
		c.seen[f] = true
		c.findings = append(c.findings, f)
	}
}

func (c *converter) pkg(name string) *pkgState {
	p := c.pkgs[name]
	if p == nil {
		p = &pkgState{name: name, buf: map[string][]string{}}
		c.pkgs[name] = p
		c.segs = append(c.segs, segment{pkg: name})
	}
	return p
}

func key(pkg, test string) string { return pkg + "\x00" + test }

// parentOf is the longest known test of the package of which name is a subtest ("" for a top-level test).
func (c *converter) parentOf(pkg, name string) string {
	best := ""
	for n := range c.running[pkg] {
		if len(n) > len(best) && strings.HasPrefix(name, n+"/") {
			best = n
		}
	}
	return best
}

func (c *converter) test(pkg, name string) *testState {
	k := key(pkg, name)
	t := c.tests[k]
	if t == nil {
		c.norder++
		t = &testState{pkg: pkg, name: name, parent: c.parentOf(pkg, name), order: c.norder}
		c.tests[k] = t
		if c.running[pkg] == nil {
			c.running[pkg] = map[string]bool{}
		}
		c.running[pkg][name] = true
	}
	return t
}

func isResult(out, test string) bool {
	for _, s := range []string{"PASS", "FAIL", "SKIP"} {
		if strings.HasPrefix(out, "--- "+s+": "+test+" (") {
			return true
		}
	}
	return false
}

func indent(lines []string) []string {
	r := make([]string, len(lines))
	for i, l := range lines {
		r[i] = "    " + l
	}
	return r
}

func (c *converter) event(e Event) {
	switch {
	case e.Package == "" && e.ImportPath != "":
		if e.Action == "build-fail" {
			c.find("build-failed:" + e.ImportPath)
		}
		if e.Output != "" {
			c.segs = append(c.segs, segment{text: e.Output})
		}
		return
	case e.Package == "":
		c.malform++
		c.find("malformed-event")
		return
	}
	p := c.pkg(e.Package)
	if e.Test == "" {
		switch e.Action {
		case "output":
			p.chunks = append(p.chunks, chunk{text: e.Output, pkg: e.Package})
			if strings.Contains(e.Output, "[no test files]") {
				p.noFiles = true
			}
			if strings.Contains(e.Output, "[build failed]") || strings.Contains(e.Output, "[setup failed]") {
				p.buildFail = true
			}
		case "pass", "fail", "skip":
			p.terminals = append(p.terminals, e.Action)
			p.elapsed = e.Elapsed
			if e.FailedBuild != "" {
				p.buildFail = true
			}
		}
		return
	}
	t := c.test(e.Package, e.Test)
	switch e.Action {
	case "run":
		if t.started {
			c.find("duplicate-run:" + e.Package + "." + e.Test)
		}
		t.started = true
	case "pass", "fail", "skip":
		t.terminals = append(t.terminals, e.Action)
		t.elapsed = e.Elapsed
	case "output":
		if isResult(e.Output, e.Test) {
			lines := append([]string{e.Output}, p.buf[e.Test]...)
			delete(p.buf, e.Test)
			if t.parent != "" {
				if _, ok := p.buf[t.parent]; !ok {
					p.bufOrder = append(p.bufOrder, t.parent)
				}
				p.buf[t.parent] = append(p.buf[t.parent], tag(indent(lines), e.Test)...)
				return
			}
			for i, l := range lines {
				name := e.Test
				if i > 0 {
					name, l = untag(l)
				}
				p.chunks = append(p.chunks, chunk{text: l, pkg: e.Package, test: name})
			}
			return
		}
		p.chunks = append(p.chunks, chunk{text: e.Output, pkg: e.Package, test: e.Test})
	}
}

// Buffered lines carry their test name after a NUL so line ranges stay attributable after re-layout.
func tag(lines []string, test string) []string {
	r := make([]string, len(lines))
	for i, l := range lines {
		if strings.Contains(l, "\x00") {
			r[i] = l
		} else {
			r[i] = l + "\x00" + test
		}
	}
	return r
}

func untag(l string) (string, string) {
	if i := strings.LastIndexByte(l, 0); i >= 0 {
		return l[i+1:], l[:i]
	}
	return "", l
}

func (c *converter) line(l string) {
	switch {
	case strings.HasPrefix(l, "GO-IT-LIST "):
		var e Event
		if err := json.Unmarshal([]byte(strings.TrimSuffix(l[len("GO-IT-LIST "):], "\n")), &e); err != nil || e.Action == "" {
			c.find("malformed-expectation")
			return
		}
		if !c.listSeen {
			c.listSeen = true
			c.segs = append(c.segs, segment{expected: true})
		}
		c.list = append(c.list, e)
		return
	case strings.HasPrefix(l, "GO-IT-LIST-EXIT "):
		n, err := strconv.Atoi(strings.TrimSpace(l[len("GO-IT-LIST-EXIT "):]))
		if err != nil {
			c.find("malformed-expectation")
		} else {
			c.listRC = &n
		}
		c.segs = append(c.segs, segment{text: l})
		return
	case strings.HasPrefix(l, "{"):
		var e Event
		if err := json.Unmarshal([]byte(strings.TrimSuffix(l, "\n")), &e); err != nil || e.Action == "" {
			c.malform++
			c.find("malformed-json-line")
			c.segs = append(c.segs, segment{text: l})
			return
		}
		c.event(e)
		return
	}
	c.segs = append(c.segs, segment{text: l})
}

type expectation struct {
	pkgs    []string
	tests   map[string][]string
	noFiles map[string]bool
}

func (c *converter) expectation() expectation {
	x := expectation{tests: map[string][]string{}, noFiles: map[string]bool{}}
	terms := map[string][]string{}
	for _, e := range c.list {
		if e.Package == "" {
			if e.Action == "build-fail" {
				c.find("expectation-build-failed:" + e.ImportPath)
			}
			continue
		}
		if _, ok := x.tests[e.Package]; !ok {
			x.pkgs = append(x.pkgs, e.Package)
			x.tests[e.Package] = nil
		}
		switch e.Action {
		case "output":
			s := strings.TrimSuffix(e.Output, "\n")
			if testName.MatchString(s) {
				x.tests[e.Package] = append(x.tests[e.Package], s)
			}
			if strings.Contains(s, "[no test files]") {
				x.noFiles[e.Package] = true
			}
		case "pass", "fail", "skip":
			terms[e.Package] = append(terms[e.Package], e.Action)
			if e.FailedBuild != "" {
				c.find("expectation-build-failed:" + e.Package)
			}
		}
	}
	if !c.listSeen {
		c.find("expectation-missing")
	}
	if c.listRC == nil {
		c.find("expectation-exit-missing")
	} else if *c.listRC != 0 {
		c.find("expectation-exit:" + strconv.Itoa(*c.listRC))
	}
	for _, p := range x.pkgs {
		t := terms[p]
		switch {
		case len(t) != 1:
			c.find("expectation-unterminated:" + p)
		case t[0] == "fail":
			c.find("expectation-failed:" + p)
		case t[0] == "skip" && !x.noFiles[p]:
			c.find("expectation-skipped:" + p)
		case t[0] == "pass" && len(x.tests[p]) == 0:
			c.find("expectation-empty:" + p)
		}
		seen := map[string]bool{}
		for _, n := range x.tests[p] {
			if seen[n] {
				c.find("expectation-duplicate:" + p + "." + n)
			}
			seen[n] = true
		}
	}
	return x
}

// Convert reads one invocation stream and returns its evidence.
func Convert(r io.Reader, opt Options) (*Result, error) {
	c := &converter{opt: opt, seen: map[string]bool{}, pkgs: map[string]*pkgState{}, tests: map[string]*testState{},
		running: map[string]map[string]bool{}}
	br := bufio.NewReader(r)
	for {
		l, err := br.ReadString('\n')
		if l != "" {
			if !strings.HasSuffix(l, "\n") {
				l += "\n" // a final line without its newline (a stopped writer) is kept and closed
				if strings.HasPrefix(l, "{") || strings.HasPrefix(l, "GO-IT-LIST") {
					c.find("truncated-final-line")
				}
			}
			c.line(l)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	x := c.expectation()
	c.check(x)
	return c.render(x), nil
}

func (c *converter) check(x expectation) {
	selected := map[string]bool{}
	for _, p := range x.pkgs {
		selected[p] = true
	}
	for name := range c.pkgs {
		if !selected[name] {
			c.find("undeclared-package:" + name)
		}
	}
	for _, p := range x.pkgs {
		ps := c.pkgs[p]
		if ps == nil {
			c.find("package-absent:" + p)
			continue
		}
		switch {
		case len(ps.terminals) == 0:
			c.find("package-unterminated:" + p)
		case len(ps.terminals) > 1:
			c.find("package-duplicate-terminal:" + p)
		case ps.terminals[0] == "fail":
			c.find("package-failed:" + p)
		case ps.terminals[0] == "skip" && !(x.noFiles[p] && ps.noFiles && len(x.tests[p]) == 0):
			c.find("package-skipped:" + p)
		case ps.terminals[0] == "pass" && c.opt.RC != 0:
			c.find("package-pass-with-exit:" + p)
		}
		if ps.buildFail {
			c.find("build-failed:" + p)
		}
		exp := map[string]bool{}
		for _, n := range x.tests[p] {
			exp[n] = true
		}
		runs := 0
		for _, t := range c.tests {
			if t.pkg != p || t.parent != "" {
				continue
			}
			if t.started {
				runs++
			}
			if !exp[t.name] {
				c.find("undeclared-test:" + p + "." + t.name)
			}
		}
		if len(x.tests[p]) > 0 && runs == 0 {
			c.find("zero-tests:" + p)
		}
		for _, n := range x.tests[p] {
			if t := c.tests[key(p, n)]; t == nil || !t.started {
				c.find("omitted:" + p + "." + n)
			}
		}
	}
	for _, t := range c.sortedTests() {
		id := t.pkg + "." + t.name
		switch {
		case !t.started:
			c.find("not-started:" + id)
		case len(t.terminals) == 0:
			if !c.opt.Interrupted {
				c.find("unfinished:" + id)
			}
		case len(t.terminals) > 1:
			contra := false
			for _, s := range t.terminals[1:] {
				if s != t.terminals[0] {
					contra = true
				}
			}
			if contra {
				c.find("contradictory:" + id)
			} else {
				c.find("duplicate-terminal:" + id)
			}
		case t.terminals[0] == "skip":
			c.find("skipped:" + id)
		}
	}
	if c.opt.RC != 0 {
		c.find("exit:" + strconv.Itoa(c.opt.RC))
	}
	if c.opt.Interrupted {
		c.find("interrupted")
	}
}

func (c *converter) sortedTests() []*testState {
	r := make([]*testState, 0, len(c.tests))
	for _, t := range c.tests {
		r = append(r, t)
	}
	sort.Slice(r, func(i, j int) bool { return r[i].order < r[j].order })
	return r
}

func (t *testState) status(interrupted bool) string {
	switch {
	case !t.started:
		return "not-started"
	case len(t.terminals) == 0:
		return "unfinished"
	case len(t.terminals) > 1:
		for _, s := range t.terminals[1:] {
			if s != t.terminals[0] {
				return "contradictory"
			}
		}
		return "duplicate"
	}
	return t.terminals[0]
}

func (p *pkgState) terminal() string {
	switch {
	case p == nil || len(p.terminals) == 0:
		return "missing"
	case len(p.terminals) > 1:
		return "multiple"
	case p.terminals[0] == "skip" && p.noFiles:
		return "skip-no-test-files"
	}
	return p.terminals[0]
}

var cell = strings.NewReplacer("\t", " ", "\n", " ", "\r", " ")
var val = strings.NewReplacer("%", "%25", ";", "%3B", "\t", "%09", "\n", "%0A", "\r", "%0D")

func elapsed(e *float64) string {
	if e == nil {
		return "-"
	}
	return strconv.FormatFloat(*e, 'f', 3, 64)
}

func (c *converter) render(x expectation) *Result {
	var text strings.Builder
	line := 1
	emit := func(s, pkg, test string) {
		if s == "" {
			return
		}
		from := line
		text.WriteString(s)
		line += strings.Count(s, "\n")
		to := line - 1
		if !strings.HasSuffix(s, "\n") {
			to = line
		}
		if pkg != "" {
			if p := c.pkgs[pkg]; p != nil {
				if p.from == 0 {
					p.from = from
				}
				p.to = to
			}
			if test != "" {
				if t := c.tests[key(pkg, test)]; t != nil {
					if t.from == 0 {
						t.from = from
					}
					t.to = to
				}
			}
		}
	}
	for _, s := range c.segs {
		switch {
		case s.expected:
			for _, p := range x.pkgs {
				if len(x.tests[p]) == 0 {
					emit("GO-IT-EXPECTED "+p+"\t-\n", "", "")
				}
				for _, n := range x.tests[p] {
					emit("GO-IT-EXPECTED "+p+"\t"+n+"\n", "", "")
				}
			}
		case s.pkg != "":
			p := c.pkgs[s.pkg]
			for _, ch := range p.chunks {
				emit(ch.text, ch.pkg, ch.test)
			}
			for _, parent := range p.bufOrder { // results whose parent never reported (an interrupted package)
				for _, l := range p.buf[parent] {
					n, t := untag(l)
					emit(t, s.pkg, n)
				}
			}
		default:
			emit(s.text, "", "")
		}
	}
	status := "pass"
	if c.opt.Interrupted {
		status = "interrupted"
	} else if len(c.findings) > 0 {
		status = "fail"
	}
	logFile := c.opt.LogFile
	if logFile == "" {
		logFile = "go-tests/" + c.opt.Name + ".log"
	}
	nlines := line - 1
	var rows bytes.Buffer
	row := func(cols ...string) {
		for i, s := range cols {
			if i > 0 {
				rows.WriteByte('\t')
			}
			if s == "" {
				s = "-"
			}
			rows.WriteString(cell.Replace(s))
		}
		rows.WriteByte('\n')
	}
	type counts struct{ exp, pass, fail, unf int }
	per := map[string]*counts{}
	total := counts{}
	for _, p := range x.pkgs {
		per[p] = &counts{exp: len(x.tests[p])}
		total.exp += len(x.tests[p])
	}
	for _, t := range c.tests {
		k := per[t.pkg]
		if k == nil {
			k = &counts{}
			per[t.pkg] = k
		}
		st := t.status(c.opt.Interrupted)
		if st == "unfinished" {
			k.unf++
			total.unf++
		}
		if t.parent == "" {
			switch st {
			case "pass":
				k.pass++
				total.pass++
			case "fail":
				k.fail++
				total.fail++
			}
		}
	}
	var expBy, passBy []string
	for _, p := range x.pkgs {
		expBy = append(expBy, p+":"+strconv.Itoa(per[p].exp))
		passBy = append(passBy, p+":"+strconv.Itoa(per[p].pass))
	}
	detail := []string{"findings=" + val.Replace(strings.Join(c.findings, ",")),
		"expected_by_package=" + val.Replace(strings.Join(expBy, ",")), "passed_by_package=" + val.Replace(strings.Join(passBy, ","))}
	for _, kv := range c.opt.Provenance {
		detail = append(detail, kv[0]+"="+val.Replace(kv[1]))
	}
	row("invocation", c.opt.Name, strings.Join(x.pkgs, ","), "-", status, strconv.Itoa(c.opt.RC), strconv.Itoa(total.exp),
		strconv.Itoa(total.pass), strconv.Itoa(total.fail), strconv.Itoa(total.unf), "-", logFile, "1", strconv.Itoa(nlines),
		strings.Join(detail, ";"))
	for _, p := range x.pkgs {
		ps := c.pkgs[p]
		from, to, el := "-", "-", "-"
		if ps != nil {
			from, to, el = strconv.Itoa(ps.from), strconv.Itoa(ps.to), elapsed(ps.elapsed)
		}
		var pf []string
		for _, f := range c.findings {
			if strings.HasSuffix(f, ":"+p) || strings.Contains(f, ":"+p+".") {
				pf = append(pf, f)
			}
		}
		k := per[p]
		row("package", c.opt.Name, p, "-", ps.terminal(), "-", strconv.Itoa(k.exp), strconv.Itoa(k.pass), strconv.Itoa(k.fail),
			strconv.Itoa(k.unf), el, logFile, from, to, "findings="+val.Replace(strings.Join(pf, ",")))
	}
	var j bytes.Buffer
	nt, nf, ne, ns := 0, 0, 0, 0
	var cases bytes.Buffer
	tc := func(pkg, name, el, kind, msg string) {
		nt++
		cases.WriteString(`<testcase classname="` + attr(pkg) + `" name="` + attr(name) + `"`)
		if el != "-" {
			cases.WriteString(` time="` + el + `"`)
		}
		if kind == "" {
			cases.WriteString("/>\n")
			return
		}
		cases.WriteString("><" + kind + ` message="` + attr(msg) + `"/></testcase>` + "\n")
	}
	for _, t := range c.sortedTests() {
		st := t.status(c.opt.Interrupted)
		row("test", c.opt.Name, t.pkg, t.name, st, "-", "-", "-", "-", "-", elapsed(t.elapsed), logFile, strconv.Itoa(t.from),
			strconv.Itoa(t.to), "parent="+val.Replace(t.parent))
		switch st {
		case "pass":
			tc(t.pkg, t.name, elapsed(t.elapsed), "", "")
		case "fail":
			nf++
			tc(t.pkg, t.name, elapsed(t.elapsed), "failure", "fail")
		case "skip":
			ns++
			tc(t.pkg, t.name, elapsed(t.elapsed), "skipped", "skip")
		default:
			ne++
			tc(t.pkg, t.name, "-", "error", st)
		}
	}
	for _, p := range x.pkgs {
		for _, n := range x.tests[p] {
			if t := c.tests[key(p, n)]; t == nil || !t.started {
				row("test", c.opt.Name, p, n, "omitted", "-", "-", "-", "-", "-", "-", logFile, "-", "-", "parent=")
				ne++
				tc(p, n, "-", "error", "omitted")
			}
		}
		if term := c.pkgs[p].terminal(); term != "pass" && term != "skip-no-test-files" {
			ne++
			tc(p, "(package)", "-", "error", "package "+term)
		}
	}
	j.WriteString(`<testsuite name="` + attr(c.opt.Name) + `" tests="` + strconv.Itoa(nt) + `" failures="` + strconv.Itoa(nf) +
		`" errors="` + strconv.Itoa(ne) + `" skipped="` + strconv.Itoa(ns) + `">` + "\n")
	j.Write(cases.Bytes())
	j.WriteString("</testsuite>\n")
	return &Result{Text: []byte(text.String()), Rows: rows.Bytes(), JUnit: j.Bytes(), Status: status, Findings: c.findings}
}

func attr(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

// Summary is the one-line status the harness logs.
func (r *Result) Summary() string {
	return fmt.Sprintf("%s findings=%s", r.Status, strings.Join(r.Findings, ","))
}
