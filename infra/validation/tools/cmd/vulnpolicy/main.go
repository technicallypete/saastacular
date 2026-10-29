// Command vulnpolicy applies the G012 dependency vulnerability policy (Plan r1 §5.3–§5.6) to the outputs of the pinned
// scanners (osv-scanner v2.6.0 JSON, govulncheck v1.8.0 streaming JSON) and the scanned lockfiles. It is test tooling:
// standard library only, run with go run in the pinned golang image without network.
//
//	vulnpolicy evaluate --date YYYY-MM-DD --exceptions FILE [--osv FILE] [--govulncheck FILE]
//	    [--gomod FILE] [--npm-lock FILE] [--uv-lock FILE] [--report FILE]
//	vulnpolicy fixture-db --out ZIP FILE...
//
// evaluate reconciles package identities (ecosystem, normalized name, resolved version) between each lockfile and the
// scanner result, merges Go findings of both tools by advisory ID and alias, classifies severity and reachability,
// applies the exact exception rows and prints a summary (the JSON report goes to --report). Exit 0 pass; 1 policy
// failure (a blocking finding, a stale, unused or invalid exception); 2 invalid, missing or incomplete input.
// fixture-db writes the given OSV JSON records into an all.zip for the offline scanner fixtures.
package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: vulnpolicy evaluate|fixture-db ...")
		return exitInvalidInput
	}
	switch args[0] {
	case "evaluate":
		return runEvaluate(args[1:], stdout, stderr)
	case "fixture-db":
		return runFixtureDB(args[1:], stderr)
	}
	fmt.Fprintln(stderr, "usage: vulnpolicy evaluate|fixture-db ...")
	return exitInvalidInput
}

func runEvaluate(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("evaluate", flag.ContinueOnError)
	fl.SetOutput(stderr)
	date := fl.String("date", "", "scan date in UTC, YYYY-MM-DD")
	excf := fl.String("exceptions", "", "dependency-exceptions.tsv")
	osvf := fl.String("osv", "", "osv-scanner JSON output")
	gvf := fl.String("govulncheck", "", "govulncheck JSON output")
	gomod := fl.String("gomod", "", "go.mod of the scanned Go module")
	npm := fl.String("npm-lock", "", "package-lock.json")
	uv := fl.String("uv-lock", "", "uv.lock")
	repf := fl.String("report", "", "JSON report path")
	if err := fl.Parse(args); err != nil || fl.NArg() != 0 {
		return exitInvalidInput
	}
	fail := func(f string, a ...any) int {
		fmt.Fprintf(stderr, "vulnpolicy: "+f+"\n", a...)
		return exitInvalidInput
	}
	var in inputs
	d, err := time.Parse("2006-01-02", *date)
	if err != nil {
		return fail("--date must be YYYY-MM-DD")
	}
	in.Date = d
	if *excf == "" {
		return fail("--exceptions is required")
	}
	b, err := os.ReadFile(*excf)
	if err != nil {
		return fail("cannot read the exceptions file")
	}
	if in.Exceptions, in.ExcProblem, err = parseExceptions(b); err != nil {
		return fail("%v", err)
	}
	for _, lf := range []struct {
		path  string
		parse func([]byte) (*graph, error)
	}{{*gomod, parseGoMod}, {*npm, parseNPMLock}, {*uv, parseUVLock}} {
		if lf.path == "" {
			continue
		}
		b, err := os.ReadFile(lf.path)
		if err != nil {
			return fail("cannot read %s", filepath.Base(lf.path))
		}
		g, err := lf.parse(b)
		if err != nil {
			return fail("%v", err)
		}
		in.Graphs = append(in.Graphs, g)
	}
	if len(in.Graphs) == 0 {
		return fail("no lockfile given")
	}
	if *osvf != "" {
		b, err := os.ReadFile(*osvf)
		if err != nil {
			return fail("cannot read the osv-scanner output")
		}
		if in.OSV, err = parseOSV(b); err != nil {
			return fail("%v", err)
		}
	}
	if *gvf != "" {
		b, err := os.ReadFile(*gvf)
		if err != nil {
			return fail("cannot read the govulncheck output")
		}
		if in.Govuln, err = parseGovulncheck(b); err != nil {
			return fail("%v", err)
		}
	}
	if in.OSV == nil && in.Govuln == nil {
		return fail("no scanner output given")
	}
	rep := evaluate(in)
	printSummary(stdout, rep)
	if *repf != "" {
		out, _ := json.MarshalIndent(rep, "", "  ")
		if err := os.WriteFile(*repf, append(out, '\n'), 0o644); err != nil {
			return fail("cannot write the report")
		}
	}
	return rep.Exit
}

func printSummary(w io.Writer, rep report) {
	for _, g := range rep.Graphs {
		fmt.Fprintf(w, "GRAPH %s identities=%d locations=%d direct=%d transitive=%d development_only=%d roots_excluded=%s unsupported=%d missing=%d scanner_only=%d allowed=%s%s\n",
			g.Ecosystem, g.Identities, g.Locations, g.Direct, g.Transitive, g.Development, strings.Join(g.Roots, ","),
			len(g.Unsupported), len(g.Missing), len(g.ScannerOnly), strings.Join(g.Allowed, ","),
			map[bool]string{true: " toolchain=" + g.Toolchain, false: ""}[g.Toolchain != ""])
	}
	for _, f := range rep.Findings {
		state := "reported"
		switch {
		case f.Excepted:
			state = "excepted"
		case f.Blocking:
			state = "BLOCKING"
		}
		fmt.Fprintf(w, "FINDING %s %s:%s@%s ids=%s aliases=%s class=%s reachability=%s sources=%s groups=%s fixed=%s\n", state,
			f.Eco, f.Name, f.Version, strings.Join(f.IDs, ","), strings.Join(f.Aliases, ","), f.Class, nonEmpty(f.Reach, "-"),
			strings.Join(f.Sources, ","), strings.Join(f.Groups, ","), strings.Join(f.Fixed, ","))
	}
	for _, p := range rep.Incomplete {
		fmt.Fprintf(w, "INCOMPLETE %s\n", p)
	}
	for _, p := range rep.Problems {
		fmt.Fprintf(w, "PROBLEM %s\n", p)
	}
	fmt.Fprintf(w, "RESULT %s exit=%d findings=%d\n", rep.Result, rep.Exit, len(rep.Findings))
}

// runFixtureDB zips OSV JSON records (each must parse as an object with an id) into one all.zip, test support for
// the offline scanner fixtures.
func runFixtureDB(args []string, stderr io.Writer) int {
	fl := flag.NewFlagSet("fixture-db", flag.ContinueOnError)
	fl.SetOutput(stderr)
	out := fl.String("out", "", "all.zip to write")
	if err := fl.Parse(args); err != nil || *out == "" || fl.NArg() == 0 {
		fmt.Fprintln(stderr, "usage: vulnpolicy fixture-db --out ZIP FILE...")
		return exitInvalidInput
	}
	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(stderr, "vulnpolicy: cannot create the zip")
		return exitInvalidInput
	}
	zw := zip.NewWriter(f)
	for _, p := range fl.Args() {
		b, err := os.ReadFile(p)
		var rec struct {
			ID string `json:"id"`
		}
		if err != nil || json.Unmarshal(b, &rec) != nil || rec.ID == "" {
			fmt.Fprintf(stderr, "vulnpolicy: %s is not an OSV record\n", filepath.Base(p))
			return exitInvalidInput
		}
		w, err := zw.Create(rec.ID + ".json")
		if err == nil {
			_, err = w.Write(b)
		}
		if err != nil {
			fmt.Fprintln(stderr, "vulnpolicy: cannot write the zip")
			return exitInvalidInput
		}
	}
	if zw.Close() != nil || f.Close() != nil {
		fmt.Fprintln(stderr, "vulnpolicy: cannot finish the zip")
		return exitInvalidInput
	}
	return exitPass
}
