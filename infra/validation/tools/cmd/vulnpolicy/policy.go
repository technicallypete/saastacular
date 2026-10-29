package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Pinned scanner contract (G012 Plan r1 §5.1, §5.3).
const (
	wantProtocol       = "v1.0.0"
	wantGovulncheck    = "v1.8.0"
	wantGovulnScanner  = "govulncheck"
	wantScanLevel      = "symbol"
	stdlibModule       = "stdlib"
	classCritical      = "Critical"
	classHigh          = "High"
	classMedium        = "Medium"
	classLow           = "Low"
	classUnclassified  = "Unclassified"
	reachCalled        = "called"
	reachImported      = "imported"
	reachRequired      = "required"
	sourceOSVScanner   = "osv-scanner"
	sourceGovulncheck  = "govulncheck"
	exitPass           = 0
	exitPolicyFailure  = 1
	exitInvalidInput   = 2
	unsupportedPrefixG = "unsupported record: "
)

// inputs are the parsed evaluator inputs; nil members were not given.
type inputs struct {
	Graphs     []*graph
	OSV        *osvResult
	Govuln     *gvStream
	Exceptions []*exception
	ExcProblem []string
	Date       time.Time
}

// finding is one advisory on one package identity after merging both tools.
type finding struct {
	Eco      string   `json:"ecosystem"`
	Name     string   `json:"package"`
	Version  string   `json:"version"`
	IDs      []string `json:"ids"`
	Aliases  []string `json:"aliases,omitempty"`
	Class    string   `json:"class"`
	Reach    string   `json:"reachability,omitempty"`
	Sources  []string `json:"sources"`
	Groups   []string `json:"dependency_groups,omitempty"`
	Fixed    []string `json:"fixed,omitempty"`
	Blocking bool     `json:"blocking"`
	Excepted bool     `json:"excepted,omitempty"`
}

// graphReport summarizes one reconciled graph.
type graphReport struct {
	Ecosystem   string   `json:"ecosystem"`
	Identities  int      `json:"identities"`
	Locations   int      `json:"locations"`
	Direct      int      `json:"direct"`
	Transitive  int      `json:"transitive"`
	Development int      `json:"development_only"`
	Roots       []string `json:"roots_excluded"`
	Toolchain   string   `json:"toolchain,omitempty"`
	Unsupported []string `json:"unsupported"`
	Missing     []string `json:"missing_from_scanner"`
	ScannerOnly []string `json:"unexplained_scanner_only"`
	Allowed     []string `json:"allowed_scanner_only,omitempty"`
	RootsLogged []string `json:"scanner_reported_roots,omitempty"`
}

// report is the evaluator output.
type report struct {
	Result     string        `json:"result"`
	Exit       int           `json:"exit"`
	Graphs     []graphReport `json:"graphs"`
	Findings   []finding     `json:"findings"`
	Problems   []string      `json:"problems"`
	Incomplete []string      `json:"incomplete"`
	Scanner    *gvConfig     `json:"govulncheck_config,omitempty"`
}

// severityClass maps a CVSS base score to its class (CVSS v3/v4 qualitative bands).
func severityClass(score float64) string {
	switch {
	case score >= 9.0:
		return classCritical
	case score >= 7.0:
		return classHigh
	case score >= 4.0:
		return classMedium
	case score > 0:
		return classLow
	}
	return classUnclassified
}

// labelClass maps an advisory's database_specific.severity label.
func labelClass(l string) string {
	switch strings.ToUpper(strings.TrimSpace(l)) {
	case "CRITICAL":
		return classCritical
	case "HIGH":
		return classHigh
	case "MODERATE", "MEDIUM":
		return classMedium
	case "LOW":
		return classLow
	}
	return ""
}

func trimV(v string) string { return strings.TrimPrefix(v, "v") }

// evaluate applies the Spec R4 policy. Exit 2 (incomplete or invalid input) dominates exit 1 (policy failure).
func evaluate(in inputs) report {
	rep := report{Problems: []string{}, Incomplete: []string{}, Findings: []finding{}}
	incomplete := func(f string, a ...any) { rep.Incomplete = append(rep.Incomplete, fmt.Sprintf(f, a...)) }

	// ---- scanner contract ----
	var goGraph *graph
	for _, g := range in.Graphs {
		if g.Eco == ecoGo {
			goGraph = g
		}
	}
	if in.Govuln != nil {
		c := in.Govuln.Config
		rep.Scanner = c
		switch {
		case c.ProtocolVersion != wantProtocol:
			incomplete("govulncheck protocol_version %q, want %s", c.ProtocolVersion, wantProtocol)
		case c.ScannerName != wantGovulnScanner || c.ScannerVersion != wantGovulncheck:
			incomplete("govulncheck scanner %s %s, want %s %s", c.ScannerName, c.ScannerVersion, wantGovulnScanner, wantGovulncheck)
		case c.ScanLevel != wantScanLevel:
			incomplete("govulncheck scan_level %q, want %s", c.ScanLevel, wantScanLevel)
		case !strings.HasPrefix(c.DB, "file://"):
			incomplete("govulncheck db %q is not a local file database", c.DB)
		case goGraph == nil:
			incomplete("govulncheck output without the go.mod of the scanned module")
		case c.GoVersion != "go"+goGraph.GoDirective:
			incomplete("govulncheck ran on %s, but the module's go directive is %s", c.GoVersion, goGraph.GoDirective)
		}
	}

	// ---- identity reconciliation ----
	scannerIDs := map[string]map[identity]bool{}
	if in.OSV != nil {
		for _, r := range in.OSV.Results {
			for _, p := range r.Packages {
				eco := p.Package.Ecosystem
				id := identity{eco, normName(eco, p.Package.Name), trimV(p.Package.Version)}
				if scannerIDs[eco] == nil {
					scannerIDs[eco] = map[identity]bool{}
				}
				scannerIDs[eco][id] = true
			}
		}
		for eco := range scannerIDs {
			known := false
			for _, g := range in.Graphs {
				known = known || g.Eco == eco
			}
			if !known {
				incomplete("osv-scanner reports ecosystem %q, which no given lockfile declares", eco)
			}
		}
	}
	for _, g := range in.Graphs {
		gr := graphReport{Ecosystem: g.Eco, Unsupported: append([]string{}, g.Unsupported...), Missing: []string{}, ScannerOnly: []string{}}
		sort.Strings(gr.Unsupported)
		for id, n := range g.Identities {
			gr.Identities++
			gr.Locations += n
			if g.Direct[id.Name] {
				gr.Direct++
			} else {
				gr.Transitive++
			}
			if g.Dev[id] {
				gr.Development++
			}
		}
		for _, r := range g.Roots {
			gr.Roots = append(gr.Roots, r.Name)
		}
		if g.Eco == ecoGo {
			gr.Toolchain = "go" + g.GoDirective + " standard library"
		}
		for _, u := range g.Unsupported {
			incomplete("%s%s", unsupportedPrefixG, u)
		}
		switch {
		case g.Eco == ecoGo && in.OSV == nil && in.Govuln == nil, g.Eco != ecoGo && in.OSV == nil:
			incomplete("%s graph given without a scanner result", g.Eco)
		case in.OSV != nil:
			got := scannerIDs[g.Eco]
			for id := range g.Identities {
				if !got[id] {
					gr.Missing = append(gr.Missing, id.String())
				}
			}
			for id := range got {
				if _, ok := g.Identities[id]; ok {
					continue
				}
				isRoot := false
				for _, r := range g.Roots {
					isRoot = isRoot || r.Name == id.Name
				}
				switch {
				case isRoot:
					gr.RootsLogged = append(gr.RootsLogged, id.String())
				case g.Eco == ecoGo && id.Name == stdlibModule && id.Version == g.GoDirective:
					gr.Allowed = append(gr.Allowed, id.String())
				default:
					gr.ScannerOnly = append(gr.ScannerOnly, id.String())
				}
			}
			sort.Strings(gr.Missing)
			sort.Strings(gr.ScannerOnly)
			for _, m := range gr.Missing {
				incomplete("%s identity missing from the scanner result: %s", g.Eco, m)
			}
			for _, m := range gr.ScannerOnly {
				incomplete("%s identity reported by the scanner but absent from the lockfile: %s", g.Eco, m)
			}
		}
		if g.Eco == ecoGo && in.Govuln != nil {
			for _, m := range in.Govuln.Modules {
				id := identity{ecoGo, m.Path, trimV(m.Version)}
				_, req := g.Identities[id]
				isMain := len(g.Roots) == 1 && m.Path == g.Roots[0].Name && m.Version == ""
				if !req && !isMain && m.Path != stdlibModule {
					incomplete("govulncheck SBOM module %s is not a go.mod requirement", id.String())
				}
			}
		}
		rep.Graphs = append(rep.Graphs, gr)
	}

	// ---- findings ----
	var fs []*finding
	if in.OSV != nil {
		for _, r := range in.OSV.Results {
			for _, p := range r.Packages {
				eco := p.Package.Ecosystem
				vulns := map[string]osvVuln{}
				for _, v := range p.Vulnerabilities {
					vulns[v.ID] = v
				}
				// Group and advisory references must be complete in both directions, with non-empty IDs.
				grouped := map[string]bool{}
				for _, gp := range p.Groups {
					for _, id := range gp.IDs {
						grouped[id] = true
						if _, ok := vulns[id]; !ok || id == "" {
							incomplete("osv-scanner group id %q on %s:%s@%s has no advisory record in that package", id, eco, p.Package.Name, p.Package.Version)
						}
					}
				}
				for _, v := range p.Vulnerabilities {
					if v.ID == "" {
						incomplete("osv-scanner advisory record without an id on %s:%s@%s", eco, p.Package.Name, p.Package.Version)
						continue
					}
					if !grouped[v.ID] {
						incomplete("osv-scanner vulnerability %s on %s:%s@%s is in no group of that package", v.ID, eco, p.Package.Name, p.Package.Version)
					}
				}
				for _, gp := range p.Groups {
					if len(gp.IDs) == 0 {
						incomplete("osv-scanner group without ids on %s", p.Package.Name)
						continue
					}
					f := &finding{Eco: eco, Name: normName(eco, p.Package.Name), Version: trimV(p.Package.Version),
						IDs: slices.Clone(gp.IDs), Aliases: slices.Clone(gp.Aliases), Sources: []string{sourceOSVScanner},
						Groups: slices.Clone(p.DependencyGroups)}
					if s, err := strconv.ParseFloat(strings.TrimSpace(gp.MaxSeverity), 64); err == nil && s > 0 && s <= 10 {
						f.Class = severityClass(s)
					} else {
						for _, id := range gp.IDs {
							var ds struct {
								Severity string `json:"severity"`
							}
							if v, ok := vulns[id]; ok && len(v.DatabaseSpecific) > 0 && json.Unmarshal(v.DatabaseSpecific, &ds) == nil {
								if c := labelClass(ds.Severity); c != "" && (f.Class == "" || rank(c) > rank(f.Class)) {
									f.Class = c
								}
							}
						}
						if f.Class == "" {
							f.Class = classUnclassified
						}
					}
					for _, id := range gp.IDs {
						for _, a := range vulns[id].Affected {
							for _, rg := range a.Ranges {
								for _, ev := range rg.Events {
									if x, ok := ev["fixed"]; ok && !slices.Contains(f.Fixed, x) {
										f.Fixed = append(f.Fixed, x)
									}
								}
							}
						}
					}
					fs = append(fs, f)
				}
			}
		}
	}
	if in.Govuln != nil {
		type gvKey struct{ osv, module, version string }
		levels := map[gvKey]string{}
		fixed := map[gvKey]string{}
		var order []gvKey
		for _, gf := range in.Govuln.Findings {
			fr := gf.Trace[0]
			lvl := reachRequired
			if fr.Function != "" {
				lvl = reachCalled
			} else if fr.Package != "" {
				lvl = reachImported
			}
			k := gvKey{gf.OSV, fr.Module, trimV(fr.Version)}
			if _, ok := levels[k]; !ok {
				order = append(order, k)
			}
			if rank(lvl) > rank(levels[k]) {
				levels[k] = lvl
			}
			if gf.FixedVersion != "" {
				fixed[k] = gf.FixedVersion
			}
		}
		for _, k := range order {
			o := in.Govuln.OSVs[k.osv]
			names := append([]string{k.osv}, o.Aliases...)
			var match *finding
			for _, f := range fs {
				if f.Eco == ecoGo && f.Name == k.module && f.Version == k.version && overlaps(names, append(slices.Clone(f.IDs), f.Aliases...)) {
					match = f
					break
				}
			}
			if match == nil {
				match = &finding{Eco: ecoGo, Name: k.module, Version: k.version, IDs: []string{k.osv}, Aliases: slices.Clone(o.Aliases),
					Class: classUnclassified}
				fs = append(fs, match)
			}
			match.Sources = appendUnique(match.Sources, sourceGovulncheck)
			if rank(levels[k]) > rank(match.Reach) {
				match.Reach = levels[k]
			}
			if fx := fixed[k]; fx != "" && !slices.Contains(match.Fixed, fx) {
				match.Fixed = append(match.Fixed, fx)
			}
		}
	}
	for _, f := range fs {
		f.Blocking = f.Class == classCritical || f.Class == classHigh || f.Class == classUnclassified || f.Reach == reachCalled
	}

	// ---- exceptions ----
	rep.Problems = append(rep.Problems, in.ExcProblem...)
	for _, e := range in.Exceptions {
		if in.Date.After(e.Expires) {
			rep.Problems = append(rep.Problems, fmt.Sprintf("exception line %d (%s %s %s %s) expired on %s (stale)", e.Line, e.Eco, e.Package, e.Version, e.Advisory, e.Expires.Format("2006-01-02")))
			e.used = true // reported as stale, not also as unused
			continue
		}
		for _, f := range fs {
			if f.Eco == e.Eco && f.Name == normName(e.Eco, e.Package) && f.Version == e.Version && slices.Contains(append(slices.Clone(f.IDs), f.Aliases...), e.Advisory) {
				e.used = true
				// An npm or PyPI exception covers a development-only identity; production inclusion revokes it.
				if (f.Eco == ecoNPM || f.Eco == ecoPyPI) && !devOnly(in.Graphs, identity{f.Eco, f.Name, f.Version}) {
					rep.Problems = append(rep.Problems, fmt.Sprintf("exception line %d (%s %s %s %s) covers %s:%s@%s, which is not development-only (production inclusion revokes it)",
						e.Line, e.Eco, e.Package, e.Version, e.Advisory, f.Eco, f.Name, f.Version))
					continue
				}
				if f.Blocking {
					f.Excepted = true
				}
			}
		}
		if !e.used {
			rep.Problems = append(rep.Problems, fmt.Sprintf("exception line %d (%s %s %s %s) matches no current finding (unused)", e.Line, e.Eco, e.Package, e.Version, e.Advisory))
		}
	}
	for _, f := range fs {
		if f.Blocking && !f.Excepted {
			rep.Problems = append(rep.Problems, fmt.Sprintf("blocking finding: %s:%s@%s %s class %s%s", f.Eco, f.Name, f.Version,
				strings.Join(f.IDs, ","), f.Class, map[bool]string{true: " reachability " + f.Reach, false: ""}[f.Reach != ""]))
		}
	}
	sort.Slice(fs, func(i, j int) bool {
		a, b := fs[i], fs[j]
		if a.Eco != b.Eco {
			return a.Eco < b.Eco
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return strings.Join(a.IDs, ",") < strings.Join(b.IDs, ",")
	})
	for _, f := range fs {
		rep.Findings = append(rep.Findings, *f)
	}
	switch {
	case len(rep.Incomplete) > 0:
		rep.Result, rep.Exit = "incomplete", exitInvalidInput
	case len(rep.Problems) > 0:
		rep.Result, rep.Exit = "fail", exitPolicyFailure
	default:
		rep.Result, rep.Exit = "pass", exitPass
	}
	return rep
}

func rank(s string) int {
	switch s {
	case reachRequired, classLow:
		return 1
	case reachImported, classMedium:
		return 2
	case reachCalled, classHigh:
		return 3
	case classCritical:
		return 4
	}
	return 0
}

// devOnly reports whether the identity is development-only in its lockfile graph.
func devOnly(graphs []*graph, id identity) bool {
	for _, g := range graphs {
		if g.Eco == id.Eco && g.Dev[id] {
			return true
		}
	}
	return false
}

func overlaps(a, b []string) bool {
	for _, x := range a {
		if slices.Contains(b, x) {
			return true
		}
	}
	return false
}

func appendUnique(s []string, v string) []string {
	if slices.Contains(s, v) {
		return s
	}
	return append(s, v)
}
