// Command gotestreport converts one harness Go test invocation stream (G064 Plan r2 §5.2) read from stdin.
//
//	gotestreport --name NAME --rc N [--interrupted] [--log-file go-tests/NAME.log] [--prov key=value]... <stream
//
// writes a tar archive to stdout with NAME.log (the reconstructed readable text), NAME.tsv (the invocation, package
// and test rows, no header), NAME.junit.xml (a JUnit fragment) and NAME.status (one summary line). Exit status: 0 when
// the invocation passed every completeness rule, 1 when it failed or was interrupted (the archive is still complete),
// 2 on a usage or read error (no archive).
//
//	gotestreport selftest          runs every fixture of the converter (harness control T-6)
//	gotestreport fixture NAME      prints a fixture's stream (NAME "-text" prints the passing fixture's expected text)
package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"validationtools/internal/gotest"
)

var safeName = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,120}$`)

type provFlag [][2]string

func (p *provFlag) String() string { return "" }
func (p *provFlag) Set(s string) error {
	k, v, ok := strings.Cut(s, "=")
	if !ok || k == "" || strings.ContainsAny(k, ";=\t\n") {
		return fmt.Errorf("--prov wants key=value")
	}
	*p = append(*p, [2]string{k, v})
	return nil
}

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

func run(args []string, in io.Reader, out, errw io.Writer) int {
	if len(args) > 0 && args[0] == "selftest" {
		bad := 0
		for _, f := range gotest.Fixtures() {
			if msg := f.Check(); msg != "" {
				bad++
				fmt.Fprintf(out, "FAIL %s: %s\n", f.Name, msg)
			} else {
				fmt.Fprintf(out, "PASS %s\n", f.Name)
			}
		}
		if bad > 0 {
			return 1
		}
		return 0
	}
	if len(args) == 2 && args[0] == "fixture" {
		if args[1] == "-text" {
			_, _ = io.WriteString(out, gotest.PassingText)
			return 0
		}
		for _, f := range gotest.Fixtures() {
			if f.Name == args[1] {
				_, _ = io.WriteString(out, f.Stream)
				return 0
			}
		}
		fmt.Fprintln(errw, "gotestreport: unknown fixture")
		return 2
	}
	fs := flag.NewFlagSet("gotestreport", flag.ContinueOnError)
	fs.SetOutput(errw)
	var opt gotest.Options
	var prov provFlag
	fs.StringVar(&opt.Name, "name", "", "invocation name")
	fs.IntVar(&opt.RC, "rc", -1, "the invocation's exit status")
	fs.BoolVar(&opt.Interrupted, "interrupted", false, "the invocation did not end")
	fs.StringVar(&opt.LogFile, "log-file", "", "published log the rows reference")
	fs.Var(&prov, "prov", "provenance key=value (repeatable)")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 || !safeName.MatchString(opt.Name) || opt.RC < 0 {
		fmt.Fprintln(errw, "gotestreport: usage: --name NAME --rc N [--interrupted] [--log-file F] [--prov k=v]... <stream")
		return 2
	}
	opt.Provenance = prov
	r, err := gotest.Convert(in, opt)
	if err != nil {
		fmt.Fprintln(errw, "gotestreport: read:", err)
		return 2
	}
	tw := tar.NewWriter(out)
	for _, f := range []struct {
		name string
		data []byte
	}{
		{opt.Name + ".log", r.Text}, {opt.Name + ".tsv", r.Rows}, {opt.Name + ".junit.xml", r.JUnit},
		{opt.Name + ".status", []byte(r.Summary() + "\n")},
	} {
		if err := tw.WriteHeader(&tar.Header{Name: f.name, Mode: 0o600, Size: int64(len(f.data)), Typeflag: tar.TypeReg}); err != nil {
			fmt.Fprintln(errw, "gotestreport: write:", err)
			return 2
		}
		if _, err := tw.Write(f.data); err != nil {
			fmt.Fprintln(errw, "gotestreport: write:", err)
			return 2
		}
	}
	if err := tw.Close(); err != nil {
		fmt.Fprintln(errw, "gotestreport: write:", err)
		return 2
	}
	if r.Status != "pass" {
		return 1
	}
	return 0
}
