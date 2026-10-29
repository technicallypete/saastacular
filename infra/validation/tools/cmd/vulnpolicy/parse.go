package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

// Ecosystems reconciled by identity (G012 Plan r1 §5.3).
const (
	ecoGo   = "Go"
	ecoNPM  = "npm"
	ecoPyPI = "PyPI"
)

// identity is (ecosystem, normalized name, resolved version).
type identity struct{ Eco, Name, Version string }

func (i identity) String() string { return i.Eco + ":" + i.Name + "@" + i.Version }

// graph is the expected side of one lockfile.
type graph struct {
	Eco         string
	Identities  map[identity]int // identity -> lockfile locations
	Roots       []identity       // project roots, excluded from coverage
	Direct      map[string]bool  // direct dependency names
	Dev         map[identity]bool
	Unsupported []string
	GoDirective string // Go only: the go directive version (stdlib coverage)
}

func newGraph(eco string) *graph {
	return &graph{Eco: eco, Identities: map[identity]int{}, Direct: map[string]bool{}, Dev: map[identity]bool{}}
}

var pep503Re = regexp.MustCompile(`[-_.]+`)

// normPyPI normalizes a PyPI name per PEP 503.
func normPyPI(n string) string { return strings.ToLower(pep503Re.ReplaceAllString(n, "-")) }

// normName normalizes a package name for its ecosystem.
func normName(eco, n string) string {
	if eco == ecoPyPI {
		return normPyPI(n)
	}
	return n
}

var (
	goModuleRe  = regexp.MustCompile(`^module\s+(\S+)\s*$`)
	goDirRe     = regexp.MustCompile(`^go\s+(\d+\.\d+(?:\.\d+)?)\s*$`)
	goReqLineRe = regexp.MustCompile(`^(?:require\s+)?(\S+)\s+v(\S+?)(\s*//\s*indirect)?\s*$`)
)

// parseGoMod reads the requirements of a go.mod: the main module is the root; replace directives are unsupported.
func parseGoMod(b []byte) (*graph, error) {
	g := newGraph(ecoGo)
	sc := bufio.NewScanner(bytes.NewReader(b))
	block := ""
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if i := strings.Index(line, "//"); i >= 0 && !strings.Contains(line[i:], "indirect") {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		switch {
		case block != "" && line == ")":
			block = ""
			continue
		case strings.HasSuffix(line, "(") && block == "":
			block = strings.Fields(line)[0]
			continue
		}
		kw := block
		if kw == "" {
			kw = strings.Fields(line)[0]
		}
		switch kw {
		case "module":
			if m := goModuleRe.FindStringSubmatch(line); m != nil {
				g.Roots = append(g.Roots, identity{ecoGo, m[1], ""})
			}
		case "go":
			if m := goDirRe.FindStringSubmatch(line); m != nil {
				g.GoDirective = m[1]
			}
		case "require":
			m := goReqLineRe.FindStringSubmatch(line)
			if m == nil {
				return nil, fmt.Errorf("go.mod: unparseable requirement")
			}
			id := identity{ecoGo, m[1], m[2]}
			g.Identities[id]++
			if m[3] == "" {
				g.Direct[m[1]] = true
			}
		case "replace":
			g.Unsupported = append(g.Unsupported, "Go replace directive: "+strings.TrimSpace(strings.TrimPrefix(line, "replace")))
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("go.mod: cannot be read completely (%v)", err)
	}
	if len(g.Roots) != 1 || g.GoDirective == "" {
		return nil, fmt.Errorf("go.mod: module or go directive missing")
	}
	return g, nil
}

// parseNPMLock reads a package-lock.json (lockfileVersion 2 or 3): the "" entry is the root; link entries, entries
// without a version and non-registry sources are unsupported; repeated locations collapse to one identity.
func parseNPMLock(b []byte) (*graph, error) {
	var lock struct {
		LockfileVersion int `json:"lockfileVersion"`
		Packages        map[string]struct {
			Name            string            `json:"name"`
			Version         string            `json:"version"`
			Resolved        string            `json:"resolved"`
			Link            bool              `json:"link"`
			Dev             bool              `json:"dev"`
			DevOptional     bool              `json:"devOptional"`
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(b, &lock); err != nil {
		return nil, fmt.Errorf("package-lock.json: %v", err)
	}
	if lock.LockfileVersion < 2 || len(lock.Packages) == 0 {
		return nil, fmt.Errorf("package-lock.json: lockfileVersion 2 or 3 with a packages map required")
	}
	g := newGraph(ecoNPM)
	root, ok := lock.Packages[""]
	if !ok {
		return nil, fmt.Errorf("package-lock.json: root entry missing")
	}
	g.Roots = append(g.Roots, identity{ecoNPM, root.Name, root.Version})
	for n := range root.Dependencies {
		g.Direct[n] = true
	}
	for n := range root.DevDependencies {
		g.Direct[n] = true
	}
	devLocs := map[identity]int{}
	for path, e := range lock.Packages {
		if path == "" {
			continue
		}
		name := e.Name
		if name == "" {
			i := strings.LastIndex(path, "node_modules/")
			if i < 0 {
				g.Unsupported = append(g.Unsupported, "npm entry outside node_modules: "+path)
				continue
			}
			name = path[i+len("node_modules/"):]
		}
		switch {
		case e.Link:
			g.Unsupported = append(g.Unsupported, "npm link entry: "+path)
			continue
		case e.Version == "":
			g.Unsupported = append(g.Unsupported, "npm entry without a version: "+path)
			continue
		case e.Resolved != "" && !strings.HasPrefix(e.Resolved, "https://registry.npmjs.org/"):
			g.Unsupported = append(g.Unsupported, "npm non-registry source: "+path)
			continue
		}
		id := identity{ecoNPM, name, e.Version}
		g.Identities[id]++
		if e.Dev || e.DevOptional {
			devLocs[id]++
		}
	}
	// Development-only: every location of the identity is a development entry.
	for id, n := range devLocs {
		if n == g.Identities[id] {
			g.Dev[id] = true
		}
	}
	return g, nil
}

// uvPackage is one [[package]] table of a uv.lock.
type uvPackage struct {
	Name, Version, Source string
	Deps, DevDeps         []string
}

var (
	tomlStrRe  = regexp.MustCompile(`^([A-Za-z0-9_-]+)\s*=\s*"([^"]*)"\s*$`)
	tomlSrcRe  = regexp.MustCompile(`^source\s*=\s*\{\s*([A-Za-z0-9_-]+)\s*=`)
	tomlNameRe = regexp.MustCompile(`\{\s*name\s*=\s*"([^"]+)"`)
)

// parseUVLock reads the [[package]] tables of a uv.lock with a small line parser (stdlib only): name, version, the
// source kind and the dependency names of the package table and its dev-dependencies table. A virtual or editable
// source is the project root; git, url, path and directory sources are unsupported.
func parseUVLock(b []byte) (*graph, error) {
	var pkgs []*uvPackage
	var cur *uvPackage
	table, list := "", ""
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && list == "" {
			table = line
			if line == "[[package]]" {
				cur = &uvPackage{}
				pkgs = append(pkgs, cur)
			}
			continue
		}
		if cur == nil {
			continue
		}
		if list != "" {
			if m := tomlNameRe.FindStringSubmatch(line); m != nil {
				if list == "deps" {
					cur.Deps = append(cur.Deps, normPyPI(m[1]))
				} else {
					cur.DevDeps = append(cur.DevDeps, normPyPI(m[1]))
				}
			}
			if line == "]" {
				list = ""
			}
			continue
		}
		switch table {
		case "[[package]]":
			if m := tomlStrRe.FindStringSubmatch(line); m != nil {
				switch m[1] {
				case "name":
					cur.Name = m[2]
				case "version":
					cur.Version = m[2]
				}
			} else if m := tomlSrcRe.FindStringSubmatch(line); m != nil {
				cur.Source = m[1]
			} else if strings.HasPrefix(line, "dependencies = [") {
				list = "deps"
				if strings.HasSuffix(line, "]") {
					for _, n := range tomlNameRe.FindAllStringSubmatch(line, -1) {
						cur.Deps = append(cur.Deps, normPyPI(n[1]))
					}
					list = ""
				}
			}
		case "[package.dev-dependencies]":
			if strings.Contains(line, "= [") {
				list = "dev"
				if strings.HasSuffix(line, "]") {
					for _, n := range tomlNameRe.FindAllStringSubmatch(line, -1) {
						cur.DevDeps = append(cur.DevDeps, normPyPI(n[1]))
					}
					list = ""
				}
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("uv.lock: cannot be read completely (%v)", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("uv.lock: no [[package]] table")
	}
	g := newGraph(ecoPyPI)
	byName := map[string]*uvPackage{}
	var rootPkgs []*uvPackage
	for _, p := range pkgs {
		if p.Name == "" {
			return nil, fmt.Errorf("uv.lock: package without a name")
		}
		switch p.Source {
		case "virtual", "editable":
			g.Roots = append(g.Roots, identity{ecoPyPI, normPyPI(p.Name), p.Version})
			rootPkgs = append(rootPkgs, p)
			continue
		case "registry":
		default:
			g.Unsupported = append(g.Unsupported, fmt.Sprintf("PyPI %s source: %s", nonEmpty(p.Source, "unknown"), p.Name))
			continue
		}
		if p.Version == "" {
			g.Unsupported = append(g.Unsupported, "PyPI entry without a version: "+p.Name)
			continue
		}
		id := identity{ecoPyPI, normPyPI(p.Name), p.Version}
		g.Identities[id]++
		byName[id.Name] = p
	}
	// Development membership: reachable from a root's dev dependencies and not from its runtime dependencies.
	walk := func(start []string) map[string]bool {
		seen := map[string]bool{}
		stack := append([]string(nil), start...)
		for len(stack) > 0 {
			n := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[n] {
				continue
			}
			seen[n] = true
			if p := byName[n]; p != nil {
				stack = append(stack, p.Deps...)
			}
		}
		return seen
	}
	var prodStart, devStart []string
	for _, r := range rootPkgs {
		prodStart = append(prodStart, r.Deps...)
		devStart = append(devStart, r.DevDeps...)
		for _, d := range r.Deps {
			g.Direct[d] = true
		}
		for _, d := range r.DevDeps {
			g.Direct[d] = true
		}
	}
	prod, dev := walk(prodStart), walk(devStart)
	for id := range g.Identities {
		if dev[id.Name] && !prod[id.Name] {
			g.Dev[id] = true
		}
	}
	return g, nil
}

func nonEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// ---- scanner outputs ----

// osvResult is the subset of the osv-scanner v2.6.0 JSON output read here (pkg/models/results.go).
type osvResult struct {
	Results []struct {
		Source struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"source"`
		Packages []osvPackage `json:"packages"`
	} `json:"results"`
}

type osvPackage struct {
	Package struct {
		Name      string `json:"name"`
		Version   string `json:"version"`
		Ecosystem string `json:"ecosystem"`
	} `json:"package"`
	DependencyGroups []string   `json:"dependency_groups"`
	Vulnerabilities  []osvVuln  `json:"vulnerabilities"`
	Groups           []osvGroup `json:"groups"`
}

type osvVuln struct {
	ID               string          `json:"id"`
	Aliases          []string        `json:"aliases"`
	Summary          string          `json:"summary"`
	DatabaseSpecific json.RawMessage `json:"database_specific"`
	Affected         []struct {
		Ranges []struct {
			Events []map[string]string `json:"events"`
		} `json:"ranges"`
	} `json:"affected"`
}

type osvGroup struct {
	IDs         []string `json:"ids"`
	Aliases     []string `json:"aliases"`
	MaxSeverity string   `json:"max_severity"`
}

func parseOSV(b []byte) (*osvResult, error) {
	var r osvResult
	dec := json.NewDecoder(bytes.NewReader(b))
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("osv-scanner output: %v", err)
	}
	// Exactly one JSON document: anything after it but whitespace (including a stray closing delimiter) is refused.
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("osv-scanner output: trailing data after the result document")
	}
	if r.Results == nil {
		return nil, fmt.Errorf("osv-scanner output: no results array")
	}
	return &r, nil
}

// govulncheck streaming JSON protocol v1.0.0 (internal/govulncheck).
type gvConfig struct {
	ProtocolVersion string `json:"protocol_version"`
	ScannerName     string `json:"scanner_name"`
	ScannerVersion  string `json:"scanner_version"`
	DB              string `json:"db"`
	DBLastModified  string `json:"db_last_modified"`
	GoVersion       string `json:"go_version"`
	ScanLevel       string `json:"scan_level"`
	ScanMode        string `json:"scan_mode"`
}

type gvFrame struct {
	Module   string `json:"module"`
	Version  string `json:"version"`
	Package  string `json:"package"`
	Function string `json:"function"`
	Receiver string `json:"receiver"`
}

type gvFinding struct {
	OSV          string    `json:"osv"`
	FixedVersion string    `json:"fixed_version"`
	Trace        []gvFrame `json:"trace"`
}

type gvOSV struct {
	ID      string   `json:"id"`
	Aliases []string `json:"aliases"`
	Summary string   `json:"summary"`
}

type gvStream struct {
	Config   *gvConfig
	Modules  []struct{ Path, Version string }
	SBOMSeen bool
	OSVs     map[string]gvOSV
	Findings []gvFinding
}

// parseGovulncheck reads the stream of JSON values (one object per message, possibly pretty-printed).
func parseGovulncheck(b []byte) (*gvStream, error) {
	s := &gvStream{OSVs: map[string]gvOSV{}}
	dec := json.NewDecoder(bytes.NewReader(b))
	for {
		var m struct {
			Config *gvConfig `json:"config"`
			SBOM   *struct {
				GoVersion string `json:"go_version"`
				Modules   []struct {
					Path    string `json:"path"`
					Version string `json:"version"`
				} `json:"modules"`
			} `json:"SBOM"`
			OSV     *gvOSV     `json:"osv"`
			Finding *gvFinding `json:"finding"`
		}
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("govulncheck output: %v", err)
		}
		switch {
		case m.Config != nil:
			if s.Config != nil {
				return nil, fmt.Errorf("govulncheck output: more than one config message")
			}
			s.Config = m.Config
		case m.SBOM != nil:
			s.SBOMSeen = true
			for _, x := range m.SBOM.Modules {
				s.Modules = append(s.Modules, struct{ Path, Version string }{x.Path, x.Version})
			}
		case m.OSV != nil:
			s.OSVs[m.OSV.ID] = *m.OSV
		case m.Finding != nil:
			if len(m.Finding.Trace) == 0 || m.Finding.OSV == "" {
				return nil, fmt.Errorf("govulncheck output: finding without an advisory or trace")
			}
			s.Findings = append(s.Findings, *m.Finding)
		}
	}
	if s.Config == nil {
		return nil, fmt.Errorf("govulncheck output: no config message")
	}
	if !s.SBOMSeen {
		return nil, fmt.Errorf("govulncheck output: no SBOM message")
	}
	return s, nil
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
