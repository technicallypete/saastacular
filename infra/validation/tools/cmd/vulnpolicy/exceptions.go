package main

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// exceptionHeader is the exact header of infra/security/dependency-exceptions.tsv (G012 Plan r1 §5.4).
const exceptionHeader = "ecosystem\tpackage\tversion\tadvisory\trationale\towner\texpires\treview_condition\tcompensating_control\treview_record"

// exception is one exact, expiring, reviewed row. The format has no scope or group field, so a whole development or
// test dependency group can never be exempted; a row must name one package, one exact version and one advisory.
type exception struct {
	Line                                        int
	Eco, Package, Version, Advisory             string
	Rationale, Owner                            string
	Expires                                     time.Time
	ReviewCondition, Compensating, ReviewRecord string
	used                                        bool
}

var (
	exactVersionRe = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z.+_-]*$`)
	advisoryRe     = regexp.MustCompile(`^(GHSA(-[23456789cfghjmpqrvwx]{4}){3}|CVE-\d{4}-\d{4,}|GO-\d{4}-\d{4,}|PYSEC-\d{4}-\d+|OSV-\d{4}-\d+)$`)
	packageRe      = regexp.MustCompile(`^(@[a-z0-9][a-z0-9._-]*/)?[A-Za-z0-9][A-Za-z0-9._/~-]*$`)
	reviewRecordRe = regexp.MustCompile(`^management/v0/results/[A-Za-z0-9._-]+\.md$`)
	wildcardChars  = "*?^~<>|=, "
)

// parseExceptions validates every row; any invalid row is reported (the policy fails, exit 1).
func parseExceptions(b []byte) ([]*exception, []string, error) {
	var rows []*exception
	var problems []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	n := 0
	seen := map[string]int{}
	for sc.Scan() {
		n++
		line := strings.TrimRight(sc.Text(), "\r")
		if n == 1 {
			if line != exceptionHeader {
				return nil, nil, fmt.Errorf("exceptions: the header is not exactly %q", exceptionHeader)
			}
			continue
		}
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 10 {
			problems = append(problems, fmt.Sprintf("exception line %d: want 10 tab-separated fields, got %d", n, len(f)))
			continue
		}
		bad := ""
		for i, v := range f {
			if strings.TrimSpace(v) == "" || v != strings.TrimSpace(v) {
				bad = fmt.Sprintf("field %d is empty or padded", i+1)
				break
			}
		}
		e := &exception{Line: n, Eco: f[0], Package: f[1], Version: f[2], Advisory: f[3], Rationale: f[4], Owner: f[5],
			ReviewCondition: f[7], Compensating: f[8], ReviewRecord: f[9]}
		if bad == "" {
			switch {
			case e.Eco != ecoGo && e.Eco != ecoNPM && e.Eco != ecoPyPI:
				bad = "ecosystem must be Go, npm or PyPI"
			case strings.ContainsAny(e.Package, wildcardChars) || !packageRe.MatchString(e.Package):
				bad = "package must be one exact name (no wildcard, range or group)"
			case strings.ContainsAny(e.Version, wildcardChars) || !exactVersionRe.MatchString(e.Version) ||
				strings.Contains(strings.ToLower(e.Version), ".x") || e.Version == "x" || e.Version == "latest":
				bad = "version must be one exact version (no wildcard or range)"
			case !advisoryRe.MatchString(e.Advisory):
				bad = "advisory must be one exact GHSA, CVE, GO, PYSEC or OSV identifier"
			case !reviewRecordRe.MatchString(e.ReviewRecord):
				bad = "review_record must be a committed management/v0/results/*.md review path"
			}
		}
		if bad == "" {
			t, err := time.Parse("2006-01-02", f[6])
			if err != nil {
				bad = "expires must be a YYYY-MM-DD date"
			} else {
				e.Expires = t
			}
		}
		if bad != "" {
			problems = append(problems, fmt.Sprintf("exception line %d: %s", n, bad))
			continue
		}
		key := e.Eco + "\t" + normName(e.Eco, e.Package) + "\t" + e.Version + "\t" + e.Advisory
		if prev, dup := seen[key]; dup {
			problems = append(problems, fmt.Sprintf("exception line %d duplicates line %d", n, prev))
			continue
		}
		seen[key] = n
		rows = append(rows, e)
	}
	// A read error (for example a line over the scanner limit) must never leave a parsed prefix in use: a later
	// invalid or expired row could be hidden behind it.
	if err := sc.Err(); err != nil {
		return nil, nil, fmt.Errorf("exceptions: line %d cannot be read (%v)", n+1, err)
	}
	if n == 0 {
		return nil, nil, fmt.Errorf("exceptions: empty file (the header is required)")
	}
	return rows, problems, nil
}
