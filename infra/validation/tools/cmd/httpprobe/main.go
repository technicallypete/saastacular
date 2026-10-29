// Command httpprobe is the harness's HTTP client (G003 Plan §6.2, §9). It runs inside a
// service's network namespace and times each request with Go's monotonic clock.
//
//		httpprobe (--url URL | --url-file PATH) [--method GET] [--header 'Name: value']...
//		          [--header-file 'Name=PATH']... [--body-file PATH] [--timeout 10s]
//		          [--abort-after DURATION] [--parallel N] [--out FILE]
//
//	  - --header is repeatable; repeating a name sends separate field lines (duplicates are kept).
//	    Leading and trailing spaces and tabs around the value are removed.
//	  - --header-file sends a header whose value is the file's content minus one trailing LF
//	    (or CRLF); --url-file reads the URL (surrounding whitespace removed); --body-file sends
//	    the file's bytes. Values read from files never appear in argv, the output or diagnostics,
//	    and the request URL is never printed.
//	  - --parallel N sends N concurrent requests. Each gets its own generated valid context
//	    (X-Request-Id: 32 lowercase hex; traceparent: 00-<32 hex>-<16 hex>-01) unless that header
//	    is supplied, and the generated values are reported under "sent".
//	  - --abort-after cancels the request from the client side after the duration; the result
//	    is then "error":"aborted" with status 0. --timeout (default 10s) bounds each request.
//
// Output (stdout, or --out FILE): without --parallel a single JSON object; with --parallel N
// a JSON array of N objects in request order. Each object has:
//
//	status        int     HTTP status; 0 when no response was received or the request aborted
//	duration_ms   int     monotonic, floored, from send to the end of the body (or the error)
//	x_request_id  string  the response's X-Request-Id when it is exactly one value matching
//	                      ^[A-Za-z0-9._-]{8,64}$, else null
//	x_request_id_count int number of X-Request-Id field values in the response
//	cache_control string  the response's Cache-Control (values joined with ", ") when printable
//	                      ASCII of at most 256 bytes, else null
//	body_sha256   string  SHA-256 of the full body (present when a body was read)
//	body_bytes    int     body length
//	body          string  the body when it is at most 4096 bytes of valid UTF-8, else omitted
//	error         string  class only: timeout|aborted|refused|reset|dns|eof|invalid_request|body_too_large|other
//	sent          object  generated context (only with --parallel and only for generated headers)
//
// Exit status: 0 output written (request errors are reported in the output); 2 usage or I/O.
package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"
)

const (
	exitPass  = 0
	exitUsage = 2

	maxBodyRead  = 16 << 20
	maxBodyShown = 4096
	maxParallel  = 1024
)

var (
	requestIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)
	tokenRe     = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$")
	errTimeout  = errors.New("timeout")
	errAbort    = errors.New("aborted")
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

type multi []string

func (m *multi) String() string     { return strings.Join(*m, ",") }
func (m *multi) Set(v string) error { *m = append(*m, v); return nil }

type header struct{ name, value string }

// Result is one request's outcome. It never holds request URLs or supplied header values.
type Result struct {
	Status          int          `json:"status"`
	DurationMS      int64        `json:"duration_ms"`
	XRequestID      *string      `json:"x_request_id"`
	XRequestIDCount int          `json:"x_request_id_count"`
	CacheControl    *string      `json:"cache_control"`
	BodySHA256      string       `json:"body_sha256,omitempty"`
	BodyBytes       int          `json:"body_bytes"`
	Body            *string      `json:"body,omitempty"`
	Error           string       `json:"error,omitempty"`
	Sent            *SentContext `json:"sent,omitempty"`
}

// SentContext holds generated (never supplied) diagnostic headers.
type SentContext struct {
	XRequestID  string `json:"x_request_id,omitempty"`
	Traceparent string `json:"traceparent,omitempty"`
}

type probe struct {
	method     string
	target     string
	headers    []header
	body       []byte
	hasBody    bool
	timeout    time.Duration
	abortAfter time.Duration
	client     *http.Client
}

func usage(stderr io.Writer, msg string) int {
	fmt.Fprintln(stderr, "httpprobe: "+msg)
	return exitUsage
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("httpprobe", flag.ContinueOnError)
	fs.SetOutput(stderr)
	rawURL := fs.String("url", "", "request URL (must not carry secrets or canaries)")
	urlFile := fs.String("url-file", "", "file holding the request URL")
	method := fs.String("method", http.MethodGet, "request method")
	var headerArgs, headerFiles multi
	fs.Var(&headerArgs, "header", "'Name: value' (repeatable; duplicates kept)")
	fs.Var(&headerFiles, "header-file", "'Name=PATH': header value read from PATH (repeatable)")
	bodyFile := fs.String("body-file", "", "request body file")
	timeout := fs.Duration("timeout", 10*time.Second, "per-request timeout")
	abortAfter := fs.Duration("abort-after", 0, "client-side abort after this duration")
	parallel := fs.Int("parallel", 0, "number of concurrent requests with generated contexts")
	out := fs.String("out", "", "output file (default stdout)")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	parallelSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "parallel" {
			parallelSet = true
		}
	})
	if fs.NArg() != 0 {
		return usage(stderr, "unexpected arguments")
	}
	if (*rawURL == "") == (*urlFile == "") {
		return usage(stderr, "exactly one of --url or --url-file is required")
	}
	if *timeout <= 0 || *abortAfter < 0 || (parallelSet && (*parallel < 1 || *parallel > maxParallel)) {
		return usage(stderr, "invalid --timeout, --abort-after or --parallel")
	}
	if !tokenRe.MatchString(*method) {
		return usage(stderr, "invalid --method")
	}
	target := *rawURL
	if *urlFile != "" {
		b, err := os.ReadFile(*urlFile)
		if err != nil {
			return usage(stderr, "cannot read --url-file")
		}
		target = strings.TrimSpace(string(b))
	}
	if u, err := url.Parse(target); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return usage(stderr, "invalid URL")
	}
	p := &probe{method: *method, target: target, timeout: *timeout, abortAfter: *abortAfter}
	for _, h := range headerArgs {
		name, value, ok := strings.Cut(h, ":")
		if !ok || !tokenRe.MatchString(name) {
			return usage(stderr, "invalid --header (expected 'Name: value')")
		}
		p.headers = append(p.headers, header{name: name, value: strings.Trim(value, " \t")})
	}
	for _, hf := range headerFiles {
		name, path, ok := strings.Cut(hf, "=")
		if !ok || !tokenRe.MatchString(name) || path == "" {
			return usage(stderr, "invalid --header-file (expected 'Name=PATH')")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return usage(stderr, "cannot read a --header-file")
		}
		v := strings.TrimSuffix(strings.TrimSuffix(string(b), "\n"), "\r")
		p.headers = append(p.headers, header{name: name, value: v})
	}
	if *bodyFile != "" {
		b, err := os.ReadFile(*bodyFile)
		if err != nil {
			return usage(stderr, "cannot read --body-file")
		}
		p.body, p.hasBody = b, true
	}
	p.client = &http.Client{
		Transport: &http.Transport{
			Proxy:               nil,
			DialContext:         (&net.Dialer{Timeout: *timeout}).DialContext,
			DisableKeepAlives:   true,
			DisableCompression:  true,
			TLSHandshakeTimeout: *timeout,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	var output any
	if !parallelSet {
		output = p.do(nil)
	} else {
		results := make([]Result, *parallel)
		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := range results {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				gen := p.generateContext()
				<-start
				results[i] = p.do(gen)
			}(i)
		}
		close(start)
		wg.Wait()
		output = results
	}
	b, err := json.Marshal(output)
	if err != nil {
		return usage(stderr, "cannot encode output")
	}
	b = append(b, '\n')
	if *out != "" {
		if err := os.WriteFile(*out, b, 0o644); err != nil {
			return usage(stderr, "cannot write --out")
		}
		return exitPass
	}
	if _, err := stdout.Write(b); err != nil {
		return exitUsage
	}
	return exitPass
}

// generateContext returns fresh valid diagnostic headers for the headers not supplied.
func (p *probe) generateContext() *SentContext {
	s := &SentContext{}
	if !p.hasHeader("X-Request-Id") {
		s.XRequestID = randomHex(16)
	}
	if !p.hasHeader("Traceparent") {
		s.Traceparent = "00-" + randomHex(16) + "-" + randomHex(8) + "-01"
	}
	if s.XRequestID == "" && s.Traceparent == "" {
		return nil
	}
	return s
}

func (p *probe) hasHeader(name string) bool {
	for _, h := range p.headers {
		if http.CanonicalHeaderKey(h.name) == name {
			return true
		}
	}
	return false
}

// randomHex returns n random bytes as lowercase hex, never all zeros.
func randomHex(n int) string {
	b := make([]byte, n)
	for {
		if _, err := rand.Read(b); err != nil {
			panic("crypto/rand unavailable")
		}
		for _, x := range b {
			if x != 0 {
				return hex.EncodeToString(b)
			}
		}
	}
}

// validHeaderValue rejects control characters other than horizontal tab.
func validHeaderValue(v string) bool {
	for i := 0; i < len(v); i++ {
		if c := v[i]; (c < 0x20 && c != '\t') || c == 0x7f {
			return false
		}
	}
	return true
}

func (p *probe) do(gen *SentContext) Result {
	res := Result{Sent: gen}
	ctx, cancelTimeout := context.WithTimeoutCause(context.Background(), p.timeout, errTimeout)
	defer cancelTimeout()
	ctx, cancelAbort := context.WithCancelCause(ctx)
	defer cancelAbort(nil)

	var body io.Reader
	if p.hasBody {
		body = strings.NewReader(string(p.body))
	}
	req, err := http.NewRequestWithContext(ctx, p.method, p.target, body)
	if err != nil {
		res.Error = "invalid_request"
		return res
	}
	for _, h := range p.headers {
		if !validHeaderValue(h.value) {
			res.Error = "invalid_request"
			return res
		}
		if http.CanonicalHeaderKey(h.name) == "Host" {
			req.Host = h.value
			continue
		}
		req.Header[http.CanonicalHeaderKey(h.name)] = append(req.Header[http.CanonicalHeaderKey(h.name)], h.value)
	}
	if gen != nil {
		if gen.XRequestID != "" {
			req.Header.Set("X-Request-Id", gen.XRequestID)
		}
		if gen.Traceparent != "" {
			req.Header.Set("Traceparent", gen.Traceparent)
		}
	}
	if p.abortAfter > 0 {
		timer := time.AfterFunc(p.abortAfter, func() { cancelAbort(errAbort) })
		defer timer.Stop()
	}

	start := time.Now()
	resp, err := p.client.Do(req)
	if err != nil {
		res.DurationMS = time.Since(start).Milliseconds()
		res.Error = classify(ctx, err)
		return res
	}
	defer resp.Body.Close()
	data, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodyRead+1))
	res.DurationMS = time.Since(start).Milliseconds()
	if readErr != nil {
		res.Error = classify(ctx, readErr)
		if res.Error == "aborted" {
			return res
		}
	}
	if len(data) > maxBodyRead {
		res.Error = "body_too_large"
		data = data[:maxBodyRead]
	}
	res.Status = resp.StatusCode
	ids := resp.Header.Values("X-Request-Id")
	res.XRequestIDCount = len(ids)
	if len(ids) == 1 && requestIDRe.MatchString(ids[0]) {
		res.XRequestID = &ids[0]
	}
	if cc := resp.Header.Values("Cache-Control"); len(cc) > 0 {
		joined := strings.Join(cc, ", ")
		if printableASCII(joined, 256) {
			res.CacheControl = &joined
		}
	}
	sum := sha256.Sum256(data)
	res.BodySHA256 = hex.EncodeToString(sum[:])
	res.BodyBytes = len(data)
	if len(data) <= maxBodyShown && utf8.Valid(data) {
		s := string(data)
		res.Body = &s
	}
	return res
}

func printableASCII(s string, max int) bool {
	if len(s) > max {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] > 0x7e {
			return false
		}
	}
	return true
}

// classify maps a transport error to a class; error text is never reported.
func classify(ctx context.Context, err error) string {
	if ctx.Err() != nil {
		switch cause := context.Cause(ctx); {
		case errors.Is(cause, errAbort):
			return "aborted"
		case errors.Is(cause, errTimeout):
			return "timeout"
		}
	}
	var ne net.Error
	var dnsErr *net.DNSError
	switch {
	case errors.As(err, &dnsErr):
		return "dns"
	case errors.As(err, &ne) && ne.Timeout():
		return "timeout"
	case errors.Is(err, syscall.ECONNREFUSED):
		return "refused"
	case errors.Is(err, syscall.ECONNRESET):
		return "reset"
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return "eof"
	}
	return "other"
}
