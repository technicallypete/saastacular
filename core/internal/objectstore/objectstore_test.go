package objectstore

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// U-OBJ-KEY and U-OBJ-ADAPTER (G009 Plan r1 §5.1): key validation, and the adapter's request shape, error
// classification and presign shape against httptest servers. These classify; they are never provider evidence (the
// real MinIO cases are B-OBJ-IT).

const (
	orgA = "11111111-2222-4333-8444-555555555555"
	objA = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
	grpA = "99999999-8888-4777-8666-555555555555"
	// canary is provider text that must never escape an error.
	canary = "Cx5providertextcanary"
)

func TestKeys(t *testing.T) {
	k, err := NewKey(orgA, objA)
	if err != nil || k.String() != "objects/"+orgA+"/"+objA || !k.Valid() {
		t.Fatalf("NewKey: %v %v", k, err)
	}
	if (Key{}).Valid() {
		t.Fatal("zero Key is valid")
	}
	for _, bad := range []string{
		"", "objects/", "objects/" + orgA, "objects/" + orgA + "/", "objects/" + orgA + "/" + objA + "/x",
		"objects/../" + objA, "objects/" + orgA + "/../" + objA, "objects/" + orgA + "%2F" + objA, "objects%2F" + orgA + "/" + objA,
		"objects/" + orgA + "/" + strings.ToUpper(objA), "Objects/" + orgA + "/" + objA, "other/" + orgA + "/" + objA,
		"/objects/" + orgA + "/" + objA, "objects/" + orgA + "/" + objA + " ", "objects//" + objA,
		"objects/00000000-0000-0000-0000-000000000000/" + objA, "objects/" + orgA + "/00000000-0000-0000-0000-000000000000",
		"objects/" + orgA + "/" + objA + "\n", "objects/" + orgA + "\\" + objA,
	} {
		if _, err := ParseKey(bad); !errors.Is(err, ErrInvalid) {
			t.Errorf("ParseKey(%q): %v, want ErrInvalid", bad, err)
		}
	}
}

// fakeS3 is an httptest S3 endpoint: it records the requests and answers with status (and an XML error body that
// carries the canary).
type fakeS3 struct {
	srv      *httptest.Server
	status   atomic.Int64
	requests atomic.Int64
	last     atomic.Value // *http.Request clone (headers, URL)
	body     atomic.Value // []byte
	delay    time.Duration
}

func newFakeS3(t *testing.T) *fakeS3 {
	t.Helper()
	f := &fakeS3{}
	f.status.Store(200)
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requests.Add(1)
		b, _ := io.ReadAll(r.Body)
		f.body.Store(b)
		f.last.Store(r.Clone(context.Background()))
		if f.delay > 0 {
			time.Sleep(f.delay)
		}
		st := int(f.status.Load())
		if st == 0 { // reset the connection
			hj, _ := w.(http.Hijacker)
			c, _, _ := hj.Hijack()
			_ = c.Close()
			return
		}
		w.Header().Set("x-amz-meta-object-id", objA)
		w.Header().Set("x-amz-meta-organization-id", orgA)
		if st < 300 && (r.Method == http.MethodGet || r.Method == http.MethodHead) {
			w.Header().Set("Content-Length", "3")
		}
		if st >= 300 {
			w.Header().Set("Content-Type", "application/xml")
			body := "<Error><Code>" + map[int]string{404: "NoSuchKey", 403: "AccessDenied"}[st] + "</Code><Message>" + canary + "</Message></Error>"
			w.WriteHeader(st)
			if r.Method != http.MethodHead {
				_, _ = io.WriteString(w, body)
			}
			return
		}
		w.WriteHeader(st)
		if r.Method == http.MethodGet {
			_, _ = io.WriteString(w, "abc")
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeS3) store(t *testing.T) *S3 {
	t.Helper()
	s, err := New(Config{Endpoint: f.srv.URL, Region: "us-east-1", Bucket: "bucket", ForcePathStyle: true, AccessKey: "AKTEST",
		SecretKey: "secret-" + canary}, "test")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func noText(t *testing.T, what string, err error) {
	t.Helper()
	if err == nil {
		return
	}
	for _, e := range []error{ErrNotFound, ErrDenied, ErrUnavailable, ErrOutcomeUnknown, ErrInvalid} {
		if err == e {
			if strings.Contains(err.Error(), canary) || strings.Contains(err.Error(), "127.0.0.1") {
				t.Errorf("%s: error text leaks: %v", what, err)
			}
			return
		}
	}
	t.Errorf("%s: %v is not one of the closed errors", what, err)
}

func TestAdapterClassification(t *testing.T) {
	ctx := context.Background()
	key, _ := NewKey(orgA, objA)
	b := Binding{Object: objA, Organization: orgA}
	for _, c := range []struct {
		status            int
		put, get, del, hd error
	}{
		{200, nil, nil, nil, nil},
		{404, ErrOutcomeUnknown, ErrNotFound, nil, ErrNotFound},
		{403, ErrDenied, ErrDenied, ErrDenied, ErrDenied},
		{500, ErrOutcomeUnknown, ErrUnavailable, ErrUnavailable, ErrUnavailable},
		{503, ErrOutcomeUnknown, ErrUnavailable, ErrUnavailable, ErrUnavailable},
		{0, ErrOutcomeUnknown, ErrUnavailable, ErrUnavailable, ErrUnavailable},
	} {
		f := newFakeS3(t)
		f.status.Store(int64(c.status))
		s := f.store(t)
		err := s.Put(ctx, key, strings.NewReader("abc"), 3, "text/plain", b)
		if err != c.put {
			t.Errorf("status %d Put: %v, want %v", c.status, err, c.put)
		}
		noText(t, "Put", err)
		_, rc, err := s.Get(ctx, key)
		if err != c.get {
			t.Errorf("status %d Get: %v, want %v", c.status, err, c.get)
		}
		if rc != nil {
			_ = rc.Close()
		}
		noText(t, "Get", err)
		if err := s.Delete(ctx, key); err != c.del {
			t.Errorf("status %d Delete: %v, want %v", c.status, err, c.del)
		}
		_, err = s.Metadata(ctx, key)
		if err != c.hd {
			t.Errorf("status %d Head: %v, want %v", c.status, err, c.hd)
		}
		noText(t, "Head", err)
	}
	// A timeout is a closed category too.
	f := newFakeS3(t)
	f.delay = 300 * time.Millisecond
	s := f.store(t)
	tctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	if err := s.Put(tctx, key, strings.NewReader("abc"), 3, "text/plain", b); err != ErrOutcomeUnknown {
		t.Errorf("Put timeout: %v", err)
	}
	if _, err := s.Metadata(tctx, key); err != ErrUnavailable {
		t.Errorf("Head timeout: %v", err)
	}
}

func TestAdapterRequestShape(t *testing.T) {
	ctx := context.Background()
	f := newFakeS3(t)
	s := f.store(t)
	key, _ := NewKey(orgA, objA)
	if err := s.Put(ctx, key, strings.NewReader("abc"), 3, "text/plain", Binding{Object: objA, Organization: orgA, Group: grpA}); err != nil {
		t.Fatal(err)
	}
	r := f.last.Load().(*http.Request)
	if r.Method != http.MethodPut || r.URL.Path != "/bucket/"+key.String() || r.ContentLength != 3 ||
		r.Header.Get("X-Amz-Content-Sha256") != "UNSIGNED-PAYLOAD" || r.Header.Get("Content-Disposition") != "attachment" ||
		r.Header.Get("Content-Type") != "text/plain" || r.Header.Get("X-Amz-Meta-Object-Id") != objA ||
		r.Header.Get("X-Amz-Meta-Organization-Id") != orgA || r.Header.Get("X-Amz-Meta-Group-Id") != grpA ||
		r.Header.Get("X-Amz-Sdk-Checksum-Algorithm") != "" || r.Header.Get("X-Amz-Trailer") != "" {
		t.Errorf("put request: %s %s %d %v", r.Method, r.URL.Path, r.ContentLength, r.Header)
	}
	if string(f.body.Load().([]byte)) != "abc" {
		t.Errorf("put body %q", f.body.Load())
	}
	if n := f.requests.Load(); n != 1 {
		t.Errorf("%d requests, want exactly one (no retry)", n)
	}
	info, rc, err := s.Get(ctx, key)
	if err != nil || info.Size != 3 || info.Binding != (Binding{Object: objA, Organization: orgA}) {
		t.Fatalf("get: %+v %v", info, err)
	}
	_ = rc.Close()
	// A failing provider is never retried.
	f.status.Store(503)
	before := f.requests.Load()
	_ = s.Delete(ctx, key)
	if n := f.requests.Load() - before; n != 1 {
		t.Errorf("%d delete attempts, want 1", n)
	}
	// Invalid inputs never reach the provider.
	before = f.requests.Load()
	for _, err := range []error{
		s.Put(ctx, Key{}, strings.NewReader(""), 0, "text/plain", Binding{Object: objA, Organization: orgA}),
		s.Put(ctx, key, strings.NewReader(""), -1, "text/plain", Binding{Object: objA, Organization: orgA}),
		s.Put(ctx, key, strings.NewReader(""), 0, "text/plain", Binding{Object: "x", Organization: orgA}),
		s.Delete(ctx, Key{}),
	} {
		if err != ErrInvalid {
			t.Errorf("invalid input: %v", err)
		}
	}
	if _, err := s.SignedGetURL(ctx, key, 0, "text/plain"); err != ErrInvalid {
		t.Errorf("zero TTL: %v", err)
	}
	if n := f.requests.Load() - before; n != 0 {
		t.Errorf("%d requests for invalid inputs", n)
	}
}

func TestPresignShape(t *testing.T) {
	ctx := context.Background()
	s, err := New(Config{Endpoint: "http://minio:9000", SigningEndpoint: "http://127.0.0.1:9100", Region: "us-east-1", Bucket: "bucket",
		ForcePathStyle: true, AccessKey: "AKTEST", SecretKey: "secret"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	key, _ := NewKey(orgA, objA)
	raw, err := s.SignedGetURL(ctx, key, 120*time.Second, "image/png")
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if u.Scheme != "http" || u.Host != "127.0.0.1:9100" || u.Path != "/bucket/"+key.String() || q.Get("X-Amz-Expires") != "120" ||
		q.Get("response-content-disposition") != "attachment" || q.Get("response-content-type") != "image/png" ||
		q.Get("response-cache-control") != "no-store" || q.Get("X-Amz-Signature") == "" || !strings.Contains(q.Get("X-Amz-SignedHeaders"), "host") ||
		q.Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" {
		t.Errorf("presigned URL shape: host %s path %s query keys %v", u.Host, u.Path, keysOf(q))
	}
}

func keysOf(q url.Values) []string {
	var out []string
	for k := range q {
		out = append(out, k)
	}
	return out
}

// transportCases mirror config's U-OBJ-CONFIG transport table (Plan r1 §4.7): New refuses the same configurations.
var transportCases = []struct {
	name, appEnv, data, signing string
	ok                          bool
}{
	{"production https both", "production", "https://objects.example.net", "https://signing.example.net", true},
	{"production https data, default signing", "production", "https://objects.example.net", "", true},
	{"production http data", "production", "http://minio:9000", "", false},
	{"production http signing", "production", "https://objects.example.net", "http://127.0.0.1:9100", false},
	{"production mixed", "production", "http://minio:9000", "https://signing.example.net", false},
	{"development mixed", "development", "http://minio:9000", "https://signing.example.net", false},
	{"test mixed", "test", "http://minio:9000", "https://signing.example.net", false},
	{"development http both", "development", "http://minio:9000", "http://127.0.0.1:9100", true},
	{"test http both", "test", "http://minio.invalid:9000", "", true},
	{"test https data, http signing", "test", "https://objects.example.net", "http://127.0.0.1:9100", true},
	{"data path", "test", "http://minio:9000/bucket", "", false},
	{"data userinfo", "test", "http://u:p@minio:9000", "", false},
	{"data query", "test", "http://minio:9000?x=1", "", false},
	{"data trailing slash", "test", "http://minio:9000/", "", false},
	{"data relative", "test", "minio:9000", "", false},
	{"signing path", "test", "http://minio:9000", "http://127.0.0.1:9100/x", false},
}

func TestConstructorTransportRules(t *testing.T) {
	for _, c := range transportCases {
		_, err := New(Config{Endpoint: c.data, SigningEndpoint: c.signing, Region: "us-east-1", Bucket: "b", AccessKey: "a",
			SecretKey: "secret-" + canary}, c.appEnv)
		if (err == nil) != c.ok {
			t.Errorf("%s: %v, want ok=%v", c.name, err, c.ok)
		}
		if err != nil && (strings.Contains(err.Error(), canary) || strings.Contains(err.Error(), "minio") || strings.Contains(err.Error(), "example.net")) {
			t.Errorf("%s: problem echoes a value: %v", c.name, err)
		}
	}
	for _, missing := range []Config{
		{Endpoint: "http://minio:9000", Bucket: "b", AccessKey: "a", SecretKey: "s"},
		{Endpoint: "http://minio:9000", Region: "r", AccessKey: "a", SecretKey: "s"},
		{Endpoint: "http://minio:9000", Region: "r", Bucket: "b", SecretKey: "s"},
		{Endpoint: "http://minio:9000", Region: "r", Bucket: "b", AccessKey: "a"},
	} {
		if _, err := New(missing, "test"); err == nil {
			t.Errorf("incomplete config accepted: region=%q bucket=%q", missing.Region, missing.Bucket)
		}
	}
}

// The transport is bounded: a dial to an unroutable endpoint fails within the dial timeout bound, not the OS default.
func TestTransportDialBound(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close() // nothing listens: connection refused, immediately
	s, err := New(Config{Endpoint: "http://" + addr, Region: "us-east-1", Bucket: "b", ForcePathStyle: true, AccessKey: "a", SecretKey: "s"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	key, _ := NewKey(orgA, objA)
	start := time.Now()
	if _, err := s.Metadata(context.Background(), key); err != ErrUnavailable {
		t.Errorf("refused endpoint: %v", err)
	}
	if time.Since(start) > dialTimeout+time.Second {
		t.Errorf("refused connection took %v", time.Since(start))
	}
}
