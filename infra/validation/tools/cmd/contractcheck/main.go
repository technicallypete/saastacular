// Command contractcheck is the independent Go validator for the intelligence contract corpus
// (G003 Plan §8.3). It is test tooling only and is never linked into core.
//
//	contractcheck corpus          --schemas DIR --corpus DIR --out FILE
//	contractcheck ndjson          --schemas DIR --corpus DIR --file F --as event|envelope [--require-types a,b,...]
//	contractcheck schema-audit    --schema FILE --as envelope|event
//	contractcheck schema-negative --schemas DIR --corpus DIR --out FILE
//	contractcheck sn-derive       --schemas DIR --corpus DIR
//	contractcheck produce-envelopes --out FILE
//
// Exit status: 0 pass; 1 a check failed (case mismatch, invalid stream, audit class other than
// ok, wrong schema-negative class, derivation mismatch); 2 usage, I/O, invalid expectations or
// corpus layout; 3 normal mode refused before evaluating any case (schema digest mismatch,
// missing schema or a canonical schema failing the audit).
//
// Every command prints one JSON summary line on stdout. Output never contains schema, case or
// record content: only IDs, classes, counts and the paths given on the command line.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"validationtools/internal/conformance"
)

const (
	exitPass    = 0
	exitFail    = 1
	exitUsage   = 2
	exitRefused = 3
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: contractcheck corpus|ndjson|schema-audit|schema-negative|sn-derive|produce-envelopes [flags]")
		return exitUsage
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "corpus":
		return cmdCorpus(rest, stdout, stderr)
	case "ndjson":
		return cmdNDJSON(rest, stdout, stderr)
	case "schema-audit":
		return cmdSchemaAudit(rest, stdout, stderr)
	case "schema-negative":
		return cmdSchemaNegative(rest, stdout, stderr)
	case "sn-derive":
		return cmdSNDerive(rest, stdout, stderr)
	case "produce-envelopes":
		return cmdProduce(rest, stdout, stderr)
	}
	fmt.Fprintln(stderr, "contractcheck: unknown command")
	return exitUsage
}

func newFlags(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

// parse parses flags, rejects positional arguments and requires the named flags.
func parse(fs *flag.FlagSet, args []string, required map[string]*string) bool {
	if err := fs.Parse(args); err != nil {
		return false
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(fs.Output(), "%s: unexpected arguments\n", fs.Name())
		return false
	}
	for name, v := range required {
		if *v == "" {
			fmt.Fprintf(fs.Output(), "%s: --%s is required\n", fs.Name(), name)
			return false
		}
	}
	return true
}

func emit(w io.Writer, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	fmt.Fprintln(w, string(b))
}

func writeJSONFile(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

// refusal reports a classified error. Normal-mode schema refusals exit 3 with zero cases
// evaluated; everything else is a usage-class error (2).
func refusal(command string, err error, stdout, stderr io.Writer) int {
	var ce *conformance.ClassError
	if !errors.As(err, &ce) {
		emit(stdout, map[string]any{"command": command, "status": "error"})
		fmt.Fprintf(stderr, "%s: error\n", command)
		return exitUsage
	}
	status, code := "error", exitUsage
	switch ce.Class {
	case conformance.ClassExpectationsInvalid, conformance.ClassCorpusInvalid, conformance.ClassFileMissing:
	default:
		status, code = "refused", exitRefused
	}
	emit(stdout, map[string]any{"command": command, "status": status, "class": ce.Class, "detail": ce.Detail, "cases_evaluated": 0})
	fmt.Fprintf(stderr, "%s: %s\n", command, ce.Error())
	return code
}

func cmdCorpus(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("corpus", stderr)
	schemas := fs.String("schemas", "", "directory holding the two canonical schemas")
	corpus := fs.String("corpus", "", "corpus root holding expectations.json")
	out := fs.String("out", "", "result file (shared format)")
	if !parse(fs, args, map[string]*string{"schemas": schemas, "corpus": corpus, "out": out}) {
		return exitUsage
	}
	res, mismatches, err := conformance.RunCorpus(*schemas, *corpus)
	if err != nil {
		return refusal("corpus", err, stdout, stderr)
	}
	if err := writeJSONFile(*out, res); err != nil {
		fmt.Fprintln(stderr, "corpus: cannot write --out")
		return exitUsage
	}
	ids := make([]string, 0, len(mismatches))
	for _, m := range mismatches {
		ids = append(ids, m.ID)
		fmt.Fprintf(stderr, "corpus: MISMATCH %s expected valid=%t class=%s got valid=%t class=%s\n",
			m.ID, m.ExpectedValid, m.ExpectedClass, m.Valid, m.Class)
	}
	status, code := "pass", exitPass
	if len(mismatches) > 0 {
		status, code = "fail", exitFail
	}
	emit(stdout, map[string]any{
		"command": "corpus", "status": status, "cases_evaluated": len(res.Results), "mismatches": ids,
		"schema_sha256": res.SchemaSHA256, "corpus_sha256": res.CorpusSHA256, "skipped": res.Skipped,
	})
	return code
}

func cmdNDJSON(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("ndjson", stderr)
	schemas := fs.String("schemas", "", "directory holding the two canonical schemas")
	corpus := fs.String("corpus", "", "corpus root holding expectations.json (schema pins)")
	file := fs.String("file", "", "NDJSON file to check")
	as := fs.String("as", "", "record schema: event|envelope")
	require := fs.String("require-types", "", "comma-separated event types that must occur (events only)")
	if !parse(fs, args, map[string]*string{"schemas": schemas, "corpus": corpus, "file": file, "as": as}) {
		return exitUsage
	}
	schema, ok := conformance.SchemaForAs(*as)
	if !ok {
		fmt.Fprintln(stderr, "ndjson: --as must be event or envelope")
		return exitUsage
	}
	var types []string
	if *require != "" {
		if schema != conformance.SchemaEvent {
			fmt.Fprintln(stderr, "ndjson: --require-types applies to --as event only")
			return exitUsage
		}
		for _, t := range strings.Split(*require, ",") {
			if t = strings.TrimSpace(t); t != "" {
				types = append(types, t)
			}
		}
	}
	rep, err := conformance.CheckStreamFile(*schemas, *corpus, *file, schema, types)
	if err != nil {
		return refusal("ndjson", err, stdout, stderr)
	}
	status, code := "pass", exitPass
	if rep.Class != conformance.ClassValid {
		status, code = "fail", exitFail
	}
	emit(stdout, map[string]any{"command": "ndjson", "status": status, "file": *file, "as": *as, "result": rep})
	return code
}

func cmdSchemaAudit(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("schema-audit", stderr)
	schemaPath := fs.String("schema", "", "schema file (a schema-negative fixture or a canonical schema)")
	as := fs.String("as", "", "envelope|event")
	if !parse(fs, args, map[string]*string{"schema": schemaPath, "as": as}) {
		return exitUsage
	}
	schema, ok := conformance.SchemaForAs(*as)
	if !ok {
		fmt.Fprintln(stderr, "schema-audit: --as must be envelope or event")
		return exitUsage
	}
	if !conformance.AuditPathAllowed(*schemaPath) {
		fmt.Fprintln(stderr, "schema-audit: only schema-negative fixtures or canonical schema files are accepted")
		return exitUsage
	}
	class := conformance.AuditFile(*schemaPath, schema).Class
	emit(stdout, map[string]any{"file": *schemaPath, "class": class})
	if class != conformance.ClassOK {
		return exitFail
	}
	return exitPass
}

func cmdSchemaNegative(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("schema-negative", stderr)
	schemas := fs.String("schemas", "", "directory holding the two canonical schemas")
	corpus := fs.String("corpus", "", "corpus root holding expectations.json")
	out := fs.String("out", "", "result file")
	if !parse(fs, args, map[string]*string{"schemas": schemas, "corpus": corpus, "out": out}) {
		return exitUsage
	}
	rep, err := conformance.RunSchemaNegative(*schemas, *corpus, os.TempDir())
	if err != nil {
		return refusal("schema-negative", err, stdout, stderr)
	}
	if err := writeJSONFile(*out, rep); err != nil {
		fmt.Fprintln(stderr, "schema-negative: cannot write --out")
		return exitUsage
	}
	for _, f := range rep.Failures {
		fmt.Fprintf(stderr, "schema-negative: FAIL %s\n", f)
	}
	status, code := "pass", exitPass
	if !rep.Pass {
		status, code = "fail", exitFail
	}
	emit(stdout, map[string]any{"command": "schema-negative", "status": status, "derive": rep.Derive, "failures": nonNil(rep.Failures)})
	return code
}

func cmdSNDerive(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("sn-derive", stderr)
	schemas := fs.String("schemas", "", "directory holding the two canonical schemas")
	corpus := fs.String("corpus", "", "corpus root holding expectations.json")
	if !parse(fs, args, map[string]*string{"schemas": schemas, "corpus": corpus}) {
		return exitUsage
	}
	exp, err := conformance.LoadExpectations(*corpus)
	if err != nil {
		return refusal("sn-derive", err, stdout, stderr)
	}
	results, ok, err := conformance.Derive(*schemas, *corpus, exp)
	if err != nil {
		return refusal("sn-derive", err, stdout, stderr)
	}
	status, code := "pass", exitPass
	if !ok {
		status, code = "fail", exitFail
	}
	emit(stdout, map[string]any{"command": "sn-derive", "derive": status, "results": results})
	return code
}

func cmdProduce(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("produce-envelopes", stderr)
	out := fs.String("out", "", "NDJSON output file")
	if !parse(fs, args, map[string]*string{"out": out}) {
		return exitUsage
	}
	envs := conformance.ProduceEnvelopes(time.Now())
	b, err := conformance.EncodeNDJSON(envs)
	if err != nil {
		fmt.Fprintln(stderr, "produce-envelopes: encode failed")
		return exitUsage
	}
	if err := os.WriteFile(*out, b, 0o644); err != nil {
		fmt.Fprintln(stderr, "produce-envelopes: cannot write --out")
		return exitUsage
	}
	ids := make([]string, 0, len(envs))
	for _, e := range envs {
		ids = append(ids, e.RequestID)
	}
	emit(stdout, map[string]any{"command": "produce-envelopes", "status": "pass", "envelopes": len(envs), "request_ids": ids})
	return exitPass
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
