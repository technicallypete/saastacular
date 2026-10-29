// Package logrecord classifies service log output against the G003 log contract (Plan §5):
// the §5.1 envelope, the per-service event catalogue (events.json), the completion record
// schema (completion.schema.json) and web's exact non-JSON startup allow-list (§5.4). It also
// checks cross-service chains and scans for canaries. Diagnostics carry classes, file names
// and line numbers only, never log content.
package logrecord

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed events.json
var eventsJSON []byte

//go:embed completion.schema.json
var completionSchemaJSON []byte

// CompletionEvent is the completion record's event constant.
const CompletionEvent = "http.request.completed"

// Services in chain order (web -> core -> intelligence).
var Services = []string{"web", "core", "intelligence"}

// Line failure classes.
const (
	ClassNonJSON         = "non-json"
	ClassDuplicateKey    = "duplicate-key"
	ClassEnvelope        = "envelope-invalid"
	ClassUncatalogued    = "uncatalogued-event"
	ClassErrorType       = "error-type-invalid"
	ClassEventKeys       = "event-keys-invalid"
	ClassCompletion      = "completion-invalid"
	ClassDuplicateRecord = "duplicate-record"
)

var (
	timeRe      = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$`)
	classNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.]{0,63}$`)
	// RequestIDRe is the §4.1 request ID profile (full-string match).
	RequestIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)
	traceIDRe   = regexp.MustCompile(`^[0-9a-f]{32}$`)
	spanIDRe    = regexp.MustCompile(`^[0-9a-f]{16}$`)
	levels      = map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
)

type catalogueFile struct {
	Version  int                    `json:"version"`
	Services map[string]serviceFile `json:"services"`
}

type serviceFile struct {
	Logger string               `json:"logger"`
	Events map[string]eventFile `json:"events"`
}

// WebStartupAllowList is the complete set of non-JSON lines any service may print: web's Next.js
// production startup banner, as observed from the real Next 16.3.8 standalone server (Plan
// §5.4). It deviates from the Plan's anticipated patterns (recorded as deviation R14 for
// review): no leading space or indentation, one more padding space after "Local:" (9) and
// "Network:" (7), no "Starting..." line, an added "Running next.config took" line and no empty
// line. Core and intelligence allow no non-JSON line at all.
//
// Matching rules: a line is compared after removing a single trailing "\r" only; each pattern
// may match at most once per process start, where a "▲ Next.js" line begins a new process
// start; a further occurrence is unclassified (banner-repeated).
var WebStartupAllowList = []string{
	`^▲ Next\.js [0-9]+\.[0-9]+\.[0-9]+$`,
	`^- Local:         http://[^ ]+$`,
	`^- Network:       http://[^ ]+$`,
	`^✓ Ready in [0-9]+(\.[0-9]+)?m?s$`,
	`^✓ Running next\.config took [0-9]+(\.[0-9]+)?m?s$`,
}

// ClassBannerRepeated is a startup-banner line seen twice within one process start.
const ClassBannerRepeated = "banner-repeated"

type eventFile struct {
	Completion     bool           `json:"completion"`
	ErrorTypes     []string       `json:"error_types"`
	ErrorTypeClass bool           `json:"error_type_class"`
	Levels         []string       `json:"levels"`
	Logger         string         `json:"logger"`
	LoggerPattern  string         `json:"logger_pattern"`
	Required       []string       `json:"required"`
	Properties     map[string]any `json:"properties"`
	Rules          []any          `json:"rules"`
}

// Catalogue is the compiled log contract.
type Catalogue struct {
	services   map[string]*service
	completion *jsonschema.Schema
}

type service struct {
	name   string
	logger *regexp.Regexp
	allow  []*regexp.Regexp
	events map[string]*event
}

type event struct {
	completion bool
	errorTypes map[string]bool
	errorClass bool
	logger     string
	schema     *jsonschema.Schema
	errorType  *jsonschema.Schema
}

type refusingLoader struct{}

func (refusingLoader) Load(string) (any, error) { return nil, errors.New("schema loading is disabled") }

func decodeUseNumber(b []byte, v any, strict bool) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if strict {
		dec.DisallowUnknownFields()
	}
	return dec.Decode(v)
}

// Load compiles the embedded catalogue and completion schema (Draft 2020-12, format
// assertions on, every URL refused).
func Load() (*Catalogue, error) {
	var cf catalogueFile
	if err := decodeUseNumber(eventsJSON, &cf, true); err != nil || cf.Version != 1 {
		return nil, errors.New("events.json is invalid")
	}
	comp := jsonschema.NewCompiler()
	comp.DefaultDraft(jsonschema.Draft2020)
	comp.AssertFormat()
	comp.UseLoader(refusingLoader{})
	add := func(url string, doc any) (*jsonschema.Schema, error) {
		if err := comp.AddResource(url, doc); err != nil {
			return nil, err
		}
		return comp.Compile(url)
	}
	var completionDoc any
	if err := decodeUseNumber(completionSchemaJSON, &completionDoc, false); err != nil {
		return nil, errors.New("completion.schema.json is invalid")
	}
	completion, err := add("file:///validationtools/logrecord/completion.schema.json", completionDoc)
	if err != nil {
		return nil, fmt.Errorf("completion.schema.json does not compile: %w", err)
	}
	cat := &Catalogue{services: map[string]*service{}, completion: completion}
	if len(cf.Services) != len(Services) {
		return nil, errors.New("events.json must define exactly web, core and intelligence")
	}
	for _, name := range Services {
		sf, ok := cf.Services[name]
		if !ok {
			return nil, fmt.Errorf("events.json lacks service %s", name)
		}
		svc := &service{name: name, events: map[string]*event{}}
		if sf.Logger != "" {
			if svc.logger, err = regexp.Compile(sf.Logger); err != nil {
				return nil, err
			}
		}
		if name == "web" {
			for _, p := range WebStartupAllowList {
				re, err := regexp.Compile(p)
				if err != nil {
					return nil, err
				}
				svc.allow = append(svc.allow, re)
			}
		}
		for i, evName := range sortedKeys(sf.Events) {
			ef := sf.Events[evName]
			ev, err := buildEvent(name, evName, ef, svc, sf.Logger, func(kind string, doc any) (*jsonschema.Schema, error) {
				return add(fmt.Sprintf("file:///validationtools/logrecord/events/%s/%d-%s.json", name, i, kind), doc)
			})
			if err != nil {
				return nil, fmt.Errorf("events.json %s/%q: %w", name, evName, err)
			}
			svc.events[evName] = ev
		}
		cat.services[name] = svc
	}
	return cat, nil
}

func buildEvent(svcName, evName string, ef eventFile, svc *service, loggerPattern string, compile func(string, any) (*jsonschema.Schema, error)) (*event, error) {
	if evName == "" || (svc.logger == nil && (ef.Logger != "" || ef.LoggerPattern != "")) || (ef.Logger != "" && ef.LoggerPattern != "") {
		return nil, errors.New("invalid event")
	}
	if ef.Completion {
		if evName != CompletionEvent || len(ef.Levels)+len(ef.Required)+len(ef.Properties)+len(ef.Rules) != 0 || ef.LoggerPattern != "" ||
			(len(ef.ErrorTypes) == 0) == !ef.ErrorTypeClass || (svcName == "intelligence") != (ef.Logger == "intelligence.http") {
			return nil, errors.New("invalid completion entry")
		}
		ev := &event{completion: true, errorClass: ef.ErrorTypeClass, logger: ef.Logger}
		if !ef.ErrorTypeClass {
			ev.errorTypes = map[string]bool{}
			for _, t := range ef.ErrorTypes {
				if !classNameRe.MatchString(t) {
					return nil, errors.New("invalid error type category")
				}
				ev.errorTypes[t] = true
			}
		}
		return ev, nil
	}
	if len(ef.ErrorTypes) != 0 || ef.ErrorTypeClass || len(ef.Levels) == 0 {
		return nil, errors.New("invalid entry")
	}
	lv := make([]any, 0, len(ef.Levels))
	for _, l := range ef.Levels {
		if !levels[l] {
			return nil, errors.New("invalid level")
		}
		lv = append(lv, l)
	}
	props := map[string]any{
		"time":    map[string]any{"type": "string"},
		"level":   map[string]any{"enum": lv},
		"service": map[string]any{"const": svcName},
		"event":   map[string]any{"const": evName},
	}
	required := []any{"time", "level", "service", "event"}
	if svc.logger != nil {
		switch {
		case ef.Logger != "":
			props["logger"] = map[string]any{"const": ef.Logger}
		case ef.LoggerPattern != "":
			props["logger"] = map[string]any{"type": "string", "pattern": ef.LoggerPattern}
		default:
			props["logger"] = map[string]any{"type": "string", "pattern": loggerPattern}
		}
		required = append(required, "logger")
	}
	for k, v := range ef.Properties {
		if _, envelope := props[k]; envelope {
			return nil, errors.New("event property shadows the envelope")
		}
		props[k] = v
	}
	for _, r := range ef.Required {
		if _, ok := ef.Properties[r]; !ok {
			return nil, errors.New("required key without a property schema")
		}
		required = append(required, r)
	}
	doc := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"additionalProperties": false,
		"required":             required,
		"properties":           props,
	}
	if len(ef.Rules) > 0 {
		doc["allOf"] = ef.Rules
	}
	sch, err := compile("record", doc)
	if err != nil {
		return nil, err
	}
	ev := &event{schema: sch, logger: ef.Logger}
	if et, ok := ef.Properties["error_type"].(map[string]any); ok {
		frag := map[string]any{"$schema": "https://json-schema.org/draft/2020-12/schema"}
		for k, v := range et {
			frag[k] = v
		}
		if ev.errorType, err = compile("error_type", frag); err != nil {
			return nil, err
		}
	}
	return ev, nil
}

// EventNames lists a service's catalogued events, sorted.
func (c *Catalogue) EventNames(svc string) []string {
	s, ok := c.services[svc]
	if !ok {
		return nil
	}
	return sortedKeys(s.events)
}

// HasService reports whether svc is catalogued.
func (c *Catalogue) HasService(svc string) bool {
	_, ok := c.services[svc]
	return ok
}

// Line is one classified line. Class is empty when the line is classified.
type Line struct {
	N       int
	Raw     []byte
	Allowed bool
	Class   string
	Event   string
	Fields  map[string]any
	// allowIdx is the matched WebStartupAllowList index when Allowed.
	allowIdx int
}

// Classify classifies one line of svc's raw output on its own. ClassifyAll additionally
// enforces that each startup-banner pattern occurs at most once per process start.
func (c *Catalogue) Classify(svc string, raw []byte) Line {
	s := c.services[svc]
	l := Line{Raw: raw}
	if len(raw) == 0 || raw[0] != '{' {
		cmp := bytes.TrimSuffix(raw, []byte{'\r'})
		for i, re := range s.allow {
			if utf8.Valid(cmp) && re.Match(cmp) {
				l.Allowed, l.allowIdx = true, i
				return l
			}
		}
		l.Class = ClassNonJSON
		return l
	}
	fields, err := decodeObject(raw)
	if err != nil {
		l.Class = ClassNonJSON
		if errors.Is(err, errDuplicateKey) {
			l.Class = ClassDuplicateKey
		}
		return l
	}
	if !envelopeOK(s, svc, fields) {
		l.Class = ClassEnvelope
		return l
	}
	name := fields["event"].(string)
	ev, ok := s.events[name]
	if !ok {
		l.Class = ClassUncatalogued
		return l
	}
	et, hasET := fields["error_type"]
	if ev.completion {
		if hasET && !ev.errorTypeOK(et) {
			l.Class = ClassErrorType
			return l
		}
		if err := c.completion.Validate(any(fields)); err != nil {
			l.Class = ClassCompletion
			return l
		}
	} else {
		if hasET && ev.errorType != nil && ev.errorType.Validate(et) != nil {
			l.Class = ClassErrorType
			return l
		}
		if err := ev.schema.Validate(any(fields)); err != nil {
			l.Class = ClassEventKeys
			return l
		}
	}
	l.Event, l.Fields = name, fields
	return l
}

func (ev *event) errorTypeOK(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	if ev.errorClass {
		return classNameRe.MatchString(s)
	}
	return ev.errorTypes[s]
}

func envelopeOK(s *service, svc string, f map[string]any) bool {
	t, ok := f["time"].(string)
	if !ok || !timeRe.MatchString(t) {
		return false
	}
	if _, err := time.Parse(time.RFC3339, t); err != nil {
		return false
	}
	if l, ok := f["level"].(string); !ok || !levels[l] {
		return false
	}
	if v, ok := f["service"].(string); !ok || v != svc {
		return false
	}
	if e, ok := f["event"].(string); !ok || e == "" {
		return false
	}
	lg, has := f["logger"]
	if s.logger == nil {
		return !has
	}
	name, ok := lg.(string)
	return ok && s.logger.MatchString(name)
}

var (
	errNotJSON      = errors.New("not json")
	errDuplicateKey = errors.New("duplicate key")
)

// decodeObject parses one complete JSON object filling the whole line: valid UTF-8, no
// duplicate keys, nothing before or after it. Numbers are json.Number.
func decodeObject(raw []byte) (map[string]any, error) {
	if !utf8.Valid(raw) {
		return nil, errNotJSON
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	v, err := readValue(dec, 0)
	if err != nil {
		return nil, err
	}
	if dec.InputOffset() != int64(len(raw)) {
		return nil, errNotJSON
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil, errNotJSON
	}
	return obj, nil
}

func readValue(dec *json.Decoder, depth int) (any, error) {
	if depth > 64 {
		return nil, errNotJSON
	}
	tok, err := dec.Token()
	if err != nil {
		return nil, errNotJSON
	}
	d, ok := tok.(json.Delim)
	if !ok {
		return tok, nil
	}
	switch d {
	case '{':
		m := map[string]any{}
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return nil, errNotJSON
			}
			k, ok := kt.(string)
			if !ok {
				return nil, errNotJSON
			}
			if _, dup := m[k]; dup {
				return nil, errDuplicateKey
			}
			v, err := readValue(dec, depth+1)
			if err != nil {
				return nil, err
			}
			m[k] = v
		}
		if _, err := dec.Token(); err != nil {
			return nil, errNotJSON
		}
		return m, nil
	case '[':
		a := []any{}
		for dec.More() {
			v, err := readValue(dec, depth+1)
			if err != nil {
				return nil, err
			}
			a = append(a, v)
		}
		if _, err := dec.Token(); err != nil {
			return nil, errNotJSON
		}
		return a, nil
	}
	return nil, errNotJSON
}

// SplitLines splits raw output into lines; a final LF does not start another line.
func SplitLines(b []byte) [][]byte {
	if len(b) == 0 {
		return nil
	}
	if b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}
	return bytes.Split(b, []byte{'\n'})
}

// ClassifyAll classifies every line of one log.
func (c *Catalogue) ClassifyAll(svc string, data []byte) []Line {
	raw := SplitLines(data)
	out := make([]Line, len(raw))
	seen := map[int]bool{}
	for i, r := range raw {
		out[i] = c.Classify(svc, r)
		out[i].N = i + 1
		if !out[i].Allowed {
			continue
		}
		if out[i].allowIdx == 0 {
			// "▲ Next.js" begins a new process start.
			seen = map[int]bool{}
		}
		if seen[out[i].allowIdx] {
			out[i].Allowed, out[i].Class = false, ClassBannerRepeated
		}
		seen[out[i].allowIdx] = true
	}
	return out
}

// Log is a named log file's content.
type Log struct {
	Name string
	Data []byte
}

// Failure locates a failing line by file and line number, with its class.
type Failure struct {
	File  string `json:"file"`
	Line  int    `json:"line"`
	Class string `json:"class"`
}

// CompletionCounts summarizes completion records.
type CompletionCounts struct {
	Total         int            `json:"total"`
	ByRouteStatus map[string]int `json:"by_route_status"`
	ByOutcome     map[string]int `json:"by_outcome"`
}

// Summary is the result of validating a service's logs. It holds only catalogued event names,
// counts and validated enum values.
type Summary struct {
	Service        string           `json:"service"`
	Files          int              `json:"files"`
	Lines          int              `json:"lines"`
	JSONRecords    int              `json:"json_records"`
	AllowedNonJSON int              `json:"allowed_non_json"`
	Unclassified   int              `json:"unclassified"`
	Events         map[string]int   `json:"events"`
	Completions    CompletionCounts `json:"completions"`
	Failures       []Failure        `json:"failures"`
	FailuresTotal  int              `json:"failures_total"`
	Pass           bool             `json:"pass"`
}

const maxListedFailures = 200

// Validate classifies every line of svc's logs. Unless allowRepeatedIDs is set, two
// completion records with the same request_id in one service fail as duplicate-record.
func (c *Catalogue) Validate(svc string, logs []Log, allowRepeatedIDs bool) *Summary {
	sum := &Summary{
		Service: svc, Files: len(logs), Events: map[string]int{}, Failures: []Failure{},
		Completions: CompletionCounts{ByRouteStatus: map[string]int{}, ByOutcome: map[string]int{}},
	}
	seen := map[string]bool{}
	fail := func(file string, line int, class string) {
		sum.FailuresTotal++
		if len(sum.Failures) < maxListedFailures {
			sum.Failures = append(sum.Failures, Failure{File: file, Line: line, Class: class})
		}
	}
	for _, lg := range logs {
		for _, l := range c.ClassifyAll(svc, lg.Data) {
			sum.Lines++
			switch {
			case l.Allowed:
				sum.AllowedNonJSON++
			case l.Class != "":
				sum.Unclassified++
				fail(lg.Name, l.N, l.Class)
			default:
				sum.JSONRecords++
				sum.Events[l.Event]++
				if l.Event != CompletionEvent {
					continue
				}
				sum.Completions.Total++
				key := fmt.Sprintf("%s %s %s", l.Fields["method"], l.Fields["route"], l.Fields["status"])
				sum.Completions.ByRouteStatus[key]++
				sum.Completions.ByOutcome[l.Fields["outcome"].(string)]++
				id := l.Fields["request_id"].(string)
				if seen[id] && !allowRepeatedIDs {
					fail(lg.Name, l.N, ClassDuplicateRecord)
				}
				seen[id] = true
			}
		}
	}
	sum.Pass = sum.FailuresTotal == 0
	return sum
}

// CompletionRecords returns the classified completion records with the given request ID and
// the line numbers of matching records that fail classification.
func (c *Catalogue) CompletionRecords(svc string, data []byte, requestID string) (records []Line, invalid []int) {
	for _, l := range c.ClassifyAll(svc, data) {
		if l.Class == "" && l.Event == CompletionEvent {
			if l.Fields["request_id"] == requestID {
				records = append(records, l)
			}
			continue
		}
		if l.Class == "" || len(l.Raw) == 0 || l.Raw[0] != '{' {
			continue
		}
		// An invalid line that claims to be this request's completion record.
		var probe map[string]any
		if json.Unmarshal(l.Raw, &probe) == nil && probe["event"] == CompletionEvent && probe["request_id"] == requestID {
			invalid = append(invalid, l.N)
		}
	}
	return records, invalid
}

// ExtraKeys returns a record's keys beyond the envelope (time, level, service, event and, for
// intelligence, logger), sorted.
func ExtraKeys(svc string, fields map[string]any) []string {
	var keys []string
	for k := range fields {
		switch k {
		case "time", "level", "service", "event":
			continue
		case "logger":
			if svc == "intelligence" {
				continue
			}
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ValueString renders a parsed JSON scalar for comparison with an expected value.
func ValueString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return string(x)
	case bool:
		return strconv.FormatBool(x)
	case nil:
		return "null"
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// ReadCanaries parses a canary file: one canary per line, surrounding whitespace ignored,
// empty lines skipped. Every canary must be at least 8 bytes, and at least one is required.
func ReadCanaries(b []byte) ([][]byte, error) {
	var out [][]byte
	for _, l := range bytes.Split(b, []byte{'\n'}) {
		l = bytes.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		if len(l) < 8 {
			return nil, errors.New("canary shorter than 8 bytes")
		}
		out = append(out, l)
	}
	if len(out) == 0 {
		return nil, errors.New("empty canary list")
	}
	return out, nil
}

// ScanLines returns the 1-based numbers of lines containing any canary.
func ScanLines(data []byte, canaries [][]byte) []int {
	var hits []int
	for i, l := range bytes.Split(data, []byte{'\n'}) {
		for _, c := range canaries {
			if bytes.Contains(l, c) {
				hits = append(hits, i+1)
				break
			}
		}
	}
	return hits
}

// Count is an expected record count: 0, 1 or "0-1".
type Count struct{ Min, Max int }

// UnmarshalJSON accepts 0, 1, "0", "1" or "0-1".
func (c *Count) UnmarshalJSON(b []byte) error {
	switch strings.TrimSpace(string(b)) {
	case "0", `"0"`:
		*c = Count{0, 0}
	case "1", `"1"`:
		*c = Count{1, 1}
	case `"0-1"`:
		*c = Count{0, 1}
	default:
		return errors.New("count must be 0, 1 or \"0-1\"")
	}
	return nil
}

// ChainCase is one case of a chain cases file. Omitted parent_span_id means the entry record
// must carry none; the other optional fields are checked only when present.
type ChainCase struct {
	ID              string           `json:"id"`
	RequestID       string           `json:"request_id"`
	Expect          map[string]Count `json:"expect"`
	TraceID         *string          `json:"trace_id"`
	ParentSpanID    *string          `json:"parent_span_id"`
	TraceFlags      *string          `json:"trace_flags"`
	RequestIDSource *string          `json:"request_id_source"`
	TraceSource     *string          `json:"trace_source"`
}

// ChainFile is the chain cases file.
type ChainFile struct {
	Version int         `json:"version"`
	Cases   []ChainCase `json:"cases"`
}

var caseIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// ParseChainFile parses and checks a chain cases file against the services that have logs.
func ParseChainFile(b []byte, services map[string]bool) (*ChainFile, error) {
	var cf ChainFile
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cf); err != nil {
		return nil, errors.New("cases file does not parse")
	}
	if cf.Version != 1 || len(cf.Cases) == 0 {
		return nil, errors.New("cases file needs version 1 and at least one case")
	}
	ids := map[string]bool{}
	for i, c := range cf.Cases {
		bad := !caseIDRe.MatchString(c.ID) || ids[c.ID] || !RequestIDRe.MatchString(c.RequestID) ||
			(c.TraceID != nil && (!traceIDRe.MatchString(*c.TraceID) || allZero(*c.TraceID))) ||
			(c.ParentSpanID != nil && (!spanIDRe.MatchString(*c.ParentSpanID) || allZero(*c.ParentSpanID))) ||
			(c.TraceFlags != nil && *c.TraceFlags != "00" && *c.TraceFlags != "01") ||
			(c.RequestIDSource != nil && *c.RequestIDSource != "inbound" && *c.RequestIDSource != "generated") ||
			(c.TraceSource != nil && *c.TraceSource != "inbound" && *c.TraceSource != "generated") ||
			len(c.Expect) != len(services)
		for svc := range c.Expect {
			if !services[svc] {
				bad = true
			}
		}
		if bad {
			return nil, fmt.Errorf("case %d is invalid (fields, or expect must name exactly the services with logs)", i+1)
		}
		ids[c.ID] = true
	}
	return &cf, nil
}

func allZero(s string) bool { return strings.Trim(s, "0") == "" }

// ChainResult is one case's outcome. Failures are fixed codes naming services only.
type ChainResult struct {
	ID       string         `json:"id"`
	Pass     bool           `json:"pass"`
	Records  map[string]int `json:"records"`
	Failures []string       `json:"failures"`
}

// CheckChain checks each case against the services' logs (§4.3): exactly the expected number
// of completion records per service, equal request_id and trace_id, the entry record's parent
// equal to the supplied parent or absent, each downstream parent equal to its caller's span,
// distinct spans, equal (and, when given, expected) flags and inbound sources downstream.
func (c *Catalogue) CheckChain(cf *ChainFile, logs map[string][]Log) []ChainResult {
	var out []ChainResult
	for _, cs := range cf.Cases {
		r := ChainResult{ID: cs.ID, Records: map[string]int{}, Failures: []string{}}
		recs := map[string]map[string]any{}
		var order []string
		for _, svc := range Services {
			exp, ok := cs.Expect[svc]
			if !ok {
				continue
			}
			order = append(order, svc)
			var found []Line
			for _, lg := range logs[svc] {
				got, invalid := c.CompletionRecords(svc, lg.Data, cs.RequestID)
				found = append(found, got...)
				if len(invalid) > 0 {
					r.Failures = append(r.Failures, "record-invalid:"+svc)
				}
			}
			r.Records[svc] = len(found)
			if len(found) < exp.Min || len(found) > exp.Max {
				r.Failures = append(r.Failures, "count:"+svc)
			}
			if len(found) == 1 {
				recs[svc] = found[0].Fields
			}
		}
		// The entry is the first service in chain order that is expected to have a record;
		// records of services before it are already count failures.
		entryIdx := len(order)
		for i, svc := range order {
			if cs.Expect[svc].Max > 0 {
				entryIdx = i
				break
			}
		}
		spans := map[string]bool{}
		var traceID, flags string
		for i, svc := range order {
			f, ok := recs[svc]
			if !ok || i < entryIdx {
				continue
			}
			tid, _ := f["trace_id"].(string)
			fl, _ := f["trace_flags"].(string)
			if traceID == "" {
				traceID, flags = tid, fl
			} else {
				if tid != traceID {
					r.Failures = append(r.Failures, "trace-mismatch:"+svc)
				}
				if fl != flags {
					r.Failures = append(r.Failures, "flags-mismatch:"+svc)
				}
			}
			if cs.TraceID != nil && tid != *cs.TraceID {
				r.Failures = append(r.Failures, "trace-expected:"+svc)
			}
			if cs.TraceFlags != nil && fl != *cs.TraceFlags {
				r.Failures = append(r.Failures, "flags-expected:"+svc)
			}
			span, _ := f["span_id"].(string)
			if spans[span] {
				r.Failures = append(r.Failures, "span-duplicate:"+svc)
			}
			spans[span] = true
			parent, hasParent := f["parent_span_id"].(string)
			if i == entryIdx {
				if cs.ParentSpanID == nil && hasParent {
					r.Failures = append(r.Failures, "parent-unexpected:"+svc)
				}
				if cs.ParentSpanID != nil && parent != *cs.ParentSpanID {
					r.Failures = append(r.Failures, "parent:"+svc)
				}
				if cs.RequestIDSource != nil && f["request_id_source"] != *cs.RequestIDSource {
					r.Failures = append(r.Failures, "source:"+svc)
				}
				if cs.TraceSource != nil && f["trace_source"] != *cs.TraceSource {
					r.Failures = append(r.Failures, "source:"+svc)
				}
				continue
			}
			// Downstream of the entry: the direct caller's record must exist and be the parent.
			caller, ok := recs[order[i-1]]
			if !ok {
				r.Failures = append(r.Failures, "upstream-missing:"+svc)
			} else if callerSpan, _ := caller["span_id"].(string); !hasParent || parent != callerSpan {
				r.Failures = append(r.Failures, "parent:"+svc)
			}
			if f["request_id_source"] != "inbound" || f["trace_source"] != "inbound" {
				r.Failures = append(r.Failures, "source:"+svc)
			}
		}
		r.Pass = len(r.Failures) == 0
		out = append(out, r)
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
