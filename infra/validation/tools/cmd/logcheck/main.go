// Command logcheck classifies service log output against the G003 log contract (Plan §5, §9).
// It never prints log content, canaries or header values: only file names, line numbers,
// classes, counts and catalogued event names.
//
//	logcheck validate --service web|core|intelligence --file LOG [--file LOG ...] [--allow-repeated-request-ids]
//	    Classifies every line. Fails on any unclassified line (non-JSON outside web's startup
//	    allow-list, duplicate JSON key, invalid envelope, uncatalogued event, error_type outside
//	    its categories, invalid event keys, invalid completion record) and on a second completion
//	    record with the same request_id in the service (unless --allow-repeated-request-ids).
//	logcheck startup --service S --case FILE=RC:EVENT[:CATEGORY | :KEY[=VALUE]+KEY[=VALUE]...] [--case ...]
//	                 [--exactly-one-line FILE ...] [--canaries CANARYFILE]
//	    For each case: the exit status in FILE.rc equals RC, every line of FILE is classified, and
//	    exactly one record has EVENT. A third part without "+" is the expected error_type; a part
//	    with "+" is the exact set of keys beyond the envelope (time, level, service, event and,
//	    for intelligence, logger), optionally with values ("port=8180+app_env=test"; a single key
//	    is written "port+"). FILE must not contain "=". --exactly-one-line FILE requires exactly
//	    one non-empty line. --canaries fails on any canary in any case or --exactly-one-line file.
//	logcheck scan --canaries CANARYFILE PATH...
//	    Prints "path:line" for every line containing a canary (never the canary). Directories are
//	    walked (regular files only). Exit 0 clean, 1 hits, 2 error (unreadable path, canary file
//	    empty or with a canary shorter than 8 bytes).
//	logcheck chain --cases CASES.json --log web=LOG --log core=LOG [--log intelligence=LOG]
//	    Checks each case's cross-service chain (logrecord.CheckChain); --log may repeat per service.
//	logcheck records --service S --file LOG --request-id ID
//	    Prints the valid completion records for ID as JSON lines.
//
// Exit status: 0 pass, 1 check failed, 2 usage or I/O error.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"validationtools/internal/logrecord"
)

const (
	exitPass  = 0
	exitFail  = 1
	exitUsage = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type multi []string

func (m *multi) String() string     { return strings.Join(*m, ",") }
func (m *multi) Set(v string) error { *m = append(*m, v); return nil }

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: logcheck validate|startup|scan|chain|records [flags]")
		return exitUsage
	}
	cat, err := logrecord.Load()
	if err != nil {
		fmt.Fprintln(stderr, "logcheck: embedded catalogue does not load")
		return exitUsage
	}
	switch args[0] {
	case "validate":
		return cmdValidate(cat, args[1:], stdout, stderr)
	case "startup":
		return cmdStartup(cat, args[1:], stdout, stderr)
	case "scan":
		return cmdScan(args[1:], stdout, stderr)
	case "chain":
		return cmdChain(cat, args[1:], stdout, stderr)
	case "records":
		return cmdRecords(cat, args[1:], stdout, stderr)
	}
	fmt.Fprintln(stderr, "logcheck: unknown command")
	return exitUsage
}

func newFlags(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

func emit(w io.Writer, v any) {
	b, err := json.Marshal(v)
	if err == nil {
		fmt.Fprintln(w, string(b))
	}
}

func usage(stderr io.Writer, cmd, msg string) int {
	fmt.Fprintf(stderr, "%s: %s\n", cmd, msg)
	return exitUsage
}

func cmdValidate(cat *logrecord.Catalogue, args []string, stdout, stderr io.Writer) int {
	fs := newFlags("validate", stderr)
	svc := fs.String("service", "", "web|core|intelligence")
	var files multi
	fs.Var(&files, "file", "log file (repeatable)")
	allowRepeated := fs.Bool("allow-repeated-request-ids", false, "do not fail on repeated request IDs")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		return exitUsage
	}
	if !cat.HasService(*svc) || len(files) == 0 {
		return usage(stderr, "validate", "--service web|core|intelligence and at least one --file are required")
	}
	var logs []logrecord.Log
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			return usage(stderr, "validate", "cannot read "+f)
		}
		logs = append(logs, logrecord.Log{Name: f, Data: b})
	}
	sum := cat.Validate(*svc, logs, *allowRepeated)
	emit(stdout, sum)
	if !sum.Pass {
		return exitFail
	}
	return exitPass
}

type keyExpect struct {
	key      string
	value    string
	hasValue bool
}

type caseSpec struct {
	file     string
	rc       int
	event    string
	category string
	keys     []keyExpect
	hasKeys  bool
}

func parseCaseSpec(spec string) (caseSpec, error) {
	var cs caseSpec
	file, rest, ok := strings.Cut(spec, "=")
	if !ok || file == "" {
		return cs, errors.New("invalid --case")
	}
	parts := strings.SplitN(rest, ":", 3)
	if len(parts) < 2 || parts[1] == "" {
		return cs, errors.New("invalid --case")
	}
	rc, err := strconv.Atoi(parts[0])
	if err != nil {
		return cs, errors.New("invalid --case exit status")
	}
	cs.file, cs.rc, cs.event = file, rc, parts[1]
	if len(parts) == 3 {
		if !strings.Contains(parts[2], "+") {
			if parts[2] == "" {
				return cs, errors.New("invalid --case category")
			}
			cs.category = parts[2]
			return cs, nil
		}
		cs.hasKeys = true
		seen := map[string]bool{}
		for _, item := range strings.Split(parts[2], "+") {
			if item == "" {
				continue
			}
			k, v, hasV := strings.Cut(item, "=")
			if k == "" || seen[k] {
				return cs, errors.New("invalid --case key set")
			}
			seen[k] = true
			cs.keys = append(cs.keys, keyExpect{key: k, value: v, hasValue: hasV})
		}
		if len(cs.keys) == 0 {
			return cs, errors.New("invalid --case key set")
		}
	}
	return cs, nil
}

type lineFailure struct {
	Line  int    `json:"line"`
	Class string `json:"class"`
}

type caseResult struct {
	File          string        `json:"file"`
	ExpectedRC    int           `json:"expected_rc"`
	RC            int           `json:"rc"`
	Event         string        `json:"event"`
	Records       int           `json:"records"`
	Unclassified  int           `json:"unclassified"`
	Failures      []lineFailure `json:"failures"`
	CategoryMatch *bool         `json:"category_match,omitempty"`
	KeysMatch     *bool         `json:"keys_match,omitempty"`
	Pass          bool          `json:"pass"`
}

type oneLineResult struct {
	File          string `json:"file"`
	NonEmptyLines int    `json:"non_empty_lines"`
	Pass          bool   `json:"pass"`
}

func cmdStartup(cat *logrecord.Catalogue, args []string, stdout, stderr io.Writer) int {
	fs := newFlags("startup", stderr)
	svc := fs.String("service", "", "web|core|intelligence")
	var specs, oneLine multi
	fs.Var(&specs, "case", "FILE=RC:EVENT[:CATEGORY|:KEY[=VALUE]+...] (repeatable)")
	fs.Var(&oneLine, "exactly-one-line", "file that must hold exactly one non-empty line (repeatable)")
	canaryFile := fs.String("canaries", "", "canary file (one per line)")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		return exitUsage
	}
	if !cat.HasService(*svc) || len(specs) == 0 {
		return usage(stderr, "startup", "--service web|core|intelligence and at least one --case are required")
	}
	known := map[string]bool{}
	for _, e := range cat.EventNames(*svc) {
		known[e] = true
	}
	var cases []caseSpec
	for _, s := range specs {
		cs, err := parseCaseSpec(s)
		if err != nil {
			return usage(stderr, "startup", err.Error())
		}
		if !known[cs.event] {
			return usage(stderr, "startup", "--case names an event outside the service catalogue")
		}
		cases = append(cases, cs)
	}
	var canaries [][]byte
	if *canaryFile != "" {
		b, err := os.ReadFile(*canaryFile)
		if err != nil {
			return usage(stderr, "startup", "cannot read --canaries")
		}
		if canaries, err = logrecord.ReadCanaries(b); err != nil {
			return usage(stderr, "startup", err.Error())
		}
	}
	pass := true
	contents := map[string][]byte{}
	var scanOrder []string
	read := func(path string) ([]byte, error) {
		if b, ok := contents[path]; ok {
			return b, nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		contents[path] = b
		scanOrder = append(scanOrder, path)
		return b, nil
	}
	results := []caseResult{}
	for _, cs := range cases {
		data, err := read(cs.file)
		if err != nil {
			return usage(stderr, "startup", "cannot read "+cs.file)
		}
		rcRaw, err := os.ReadFile(cs.file + ".rc")
		if err != nil {
			return usage(stderr, "startup", "cannot read "+cs.file+".rc")
		}
		rc, err := strconv.Atoi(strings.TrimSpace(string(rcRaw)))
		if err != nil {
			return usage(stderr, "startup", cs.file+".rc does not hold an integer")
		}
		r := caseResult{File: cs.file, ExpectedRC: cs.rc, RC: rc, Event: cs.event, Failures: []lineFailure{}}
		var match []logrecord.Line
		for _, l := range cat.ClassifyAll(*svc, data) {
			switch {
			case l.Class != "":
				r.Unclassified++
				r.Failures = append(r.Failures, lineFailure{Line: l.N, Class: l.Class})
			case !l.Allowed && l.Event == cs.event:
				match = append(match, l)
			}
		}
		r.Records = len(match)
		r.Pass = rc == cs.rc && r.Unclassified == 0 && len(match) == 1
		if cs.category != "" {
			ok := len(match) == 1 && match[0].Fields["error_type"] == cs.category
			r.CategoryMatch = &ok
			r.Pass = r.Pass && ok
		}
		if cs.hasKeys {
			ok := len(match) == 1 && keysMatch(*svc, match[0].Fields, cs.keys)
			r.KeysMatch = &ok
			r.Pass = r.Pass && ok
		}
		pass = pass && r.Pass
		results = append(results, r)
	}
	oneLines := []oneLineResult{}
	for _, f := range oneLine {
		data, err := read(f)
		if err != nil {
			return usage(stderr, "startup", "cannot read "+f)
		}
		n := 0
		for _, l := range logrecord.SplitLines(data) {
			if len(strings.TrimSpace(string(l))) > 0 {
				n++
			}
		}
		r := oneLineResult{File: f, NonEmptyLines: n, Pass: n == 1}
		pass = pass && r.Pass
		oneLines = append(oneLines, r)
	}
	hits := []string{}
	if canaries != nil {
		for _, path := range scanOrder {
			for _, n := range logrecord.ScanLines(contents[path], canaries) {
				hits = append(hits, fmt.Sprintf("%s:%d", path, n))
			}
		}
	}
	pass = pass && len(hits) == 0
	emit(stdout, map[string]any{
		"service": *svc, "pass": pass, "cases": results, "exactly_one_line": oneLines,
		"canaries_checked": len(canaries), "canary_hits": hits,
	})
	if !pass {
		return exitFail
	}
	return exitPass
}

func keysMatch(svc string, fields map[string]any, want []keyExpect) bool {
	got := logrecord.ExtraKeys(svc, fields)
	keys := make([]string, 0, len(want))
	for _, k := range want {
		keys = append(keys, k.key)
		if k.hasValue && logrecord.ValueString(fields[k.key]) != k.value {
			return false
		}
	}
	sort.Strings(keys)
	return strings.Join(got, "\x00") == strings.Join(keys, "\x00")
}

func cmdScan(args []string, stdout, stderr io.Writer) int {
	set := newFlags("scan", stderr)
	canaryFile := set.String("canaries", "", "canary file (one per line)")
	if err := set.Parse(args); err != nil {
		return exitUsage
	}
	if *canaryFile == "" || set.NArg() == 0 {
		return usage(stderr, "scan", "--canaries and at least one PATH are required")
	}
	b, err := os.ReadFile(*canaryFile)
	if err != nil {
		return usage(stderr, "scan", "cannot read --canaries")
	}
	canaries, err := logrecord.ReadCanaries(b)
	if err != nil {
		return usage(stderr, "scan", err.Error())
	}
	hits, failed := 0, false
	scanFile := func(path string) {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "scan: cannot read %s\n", path)
			failed = true
			return
		}
		for _, n := range logrecord.ScanLines(data, canaries) {
			fmt.Fprintf(stdout, "%s:%d\n", path, n)
			hits++
		}
	}
	for _, root := range set.Args() {
		info, err := os.Stat(root)
		switch {
		case err != nil:
			fmt.Fprintf(stderr, "scan: cannot read %s\n", root)
			failed = true
		case info.IsDir():
			err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					fmt.Fprintf(stderr, "scan: cannot read %s\n", path)
					failed = true
					return nil
				}
				if d.Type().IsRegular() {
					scanFile(path)
				}
				return nil
			})
			if err != nil {
				failed = true
			}
		case info.Mode().IsRegular():
			scanFile(root)
		default:
			fmt.Fprintf(stderr, "scan: not a regular file or directory: %s\n", root)
			failed = true
		}
	}
	fmt.Fprintf(stderr, "scan: %d canaries, %d hit lines\n", len(canaries), hits)
	switch {
	case failed:
		return exitUsage
	case hits > 0:
		return exitFail
	}
	return exitPass
}

func cmdChain(cat *logrecord.Catalogue, args []string, stdout, stderr io.Writer) int {
	fs := newFlags("chain", stderr)
	casesFile := fs.String("cases", "", "chain cases file")
	var logArgs multi
	fs.Var(&logArgs, "log", "SERVICE=LOG (repeatable)")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		return exitUsage
	}
	if *casesFile == "" || len(logArgs) == 0 {
		return usage(stderr, "chain", "--cases and at least one --log are required")
	}
	logs := map[string][]logrecord.Log{}
	for _, a := range logArgs {
		svc, path, ok := strings.Cut(a, "=")
		if !ok || !cat.HasService(svc) || path == "" {
			return usage(stderr, "chain", "--log must be web=LOG, core=LOG or intelligence=LOG")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return usage(stderr, "chain", "cannot read "+path)
		}
		logs[svc] = append(logs[svc], logrecord.Log{Name: path, Data: b})
	}
	raw, err := os.ReadFile(*casesFile)
	if err != nil {
		return usage(stderr, "chain", "cannot read --cases")
	}
	have := map[string]bool{}
	for svc := range logs {
		have[svc] = true
	}
	cf, err := logrecord.ParseChainFile(raw, have)
	if err != nil {
		return usage(stderr, "chain", err.Error())
	}
	results := cat.CheckChain(cf, logs)
	pass := true
	for _, r := range results {
		pass = pass && r.Pass
	}
	emit(stdout, map[string]any{"pass": pass, "cases": results})
	if !pass {
		return exitFail
	}
	return exitPass
}

func cmdRecords(cat *logrecord.Catalogue, args []string, stdout, stderr io.Writer) int {
	fs := newFlags("records", stderr)
	svc := fs.String("service", "", "web|core|intelligence")
	file := fs.String("file", "", "log file")
	id := fs.String("request-id", "", "request ID")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		return exitUsage
	}
	if !cat.HasService(*svc) || *file == "" || !logrecord.RequestIDRe.MatchString(*id) {
		return usage(stderr, "records", "--service, --file and a valid --request-id are required")
	}
	data, err := os.ReadFile(*file)
	if err != nil {
		return usage(stderr, "records", "cannot read "+*file)
	}
	records, invalid := cat.CompletionRecords(*svc, data, *id)
	for _, r := range records {
		fmt.Fprintln(stdout, string(r.Raw))
	}
	fmt.Fprintf(stderr, "records: %d valid, %d invalid\n", len(records), len(invalid))
	if len(invalid) > 0 {
		return exitFail
	}
	return exitPass
}
