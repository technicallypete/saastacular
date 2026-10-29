package access

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"core/internal/objectstore"
	"core/internal/tenancy"
)

// G009 storage unit tests (Plan r1 §5.1): U-OBJ-POLICY, U-OBJ-INPUT, U-OBJ-READER, U-OBJ-EVENT and U-OBJ-SEAM. A spy
// database and a spy store prove that rejections happen before any transaction or provider call.

var testObjectConfig = ObjectConfig{MaxObjectBytes: 26214400, DefaultTTL: 60 * time.Second, MaxTTL: 300 * time.Second,
	TransferTimeout: 120 * time.Second}

// --- spy store ---------------------------------------------------------------------------------------------------------

type spyStore struct {
	calls atomic.Int64
	put   func(ctx context.Context, key objectstore.Key, body io.Reader, size int64) error
}

func (s *spyStore) Put(ctx context.Context, key objectstore.Key, body io.Reader, size int64, _ string, _ objectstore.Binding) error {
	s.calls.Add(1)
	if s.put != nil {
		return s.put(ctx, key, body, size)
	}
	return objectstore.ErrUnavailable
}

func (s *spyStore) Get(context.Context, objectstore.Key) (objectstore.Info, io.ReadCloser, error) {
	s.calls.Add(1)
	return objectstore.Info{}, nil, objectstore.ErrUnavailable
}

func (s *spyStore) Delete(context.Context, objectstore.Key) error {
	s.calls.Add(1)
	return objectstore.ErrUnavailable
}

func (s *spyStore) Metadata(context.Context, objectstore.Key) (objectstore.Info, error) {
	s.calls.Add(1)
	return objectstore.Info{}, objectstore.ErrUnavailable
}

func (s *spyStore) SignedGetURL(context.Context, objectstore.Key, time.Duration, string) (string, error) {
	s.calls.Add(1)
	return "", objectstore.ErrUnavailable
}

// countingBody is an upload source that counts Close calls.
type countingBody struct {
	io.Reader
	closes atomic.Int64
}

func (b *countingBody) Close() error { b.closes.Add(1); return nil }

// --- U-OBJ-INPUT ---------------------------------------------------------------------------------------------------------

func TestObjectInputs(t *testing.T) {
	h := newCtxHarness()
	store := &spyStore{}
	h.svc.objects, h.svc.ocfg = store, testObjectConfig
	rctx, end := h.begin(t, context.Background(), "cred-A")
	defer end()
	tc := h.resolved(t, rctx, tenancy.GroupID{})
	obj := uid[tenancy.ObjectID](0x51)

	check := func(name string, err error, code Code, field string, reason Reason) {
		t.Helper()
		var e *Error
		if !errors.As(err, &e) || e.Code != code || e.Field != field {
			t.Errorf("%s: %v (field %q), want %s/%s", name, err, func() string {
				if e != nil {
					return e.Field
				}
				return ""
			}(), code, field)
		}
		rec := h.logs.last(t)
		if rec["msg"] != eventStorage || rec["reason"] != string(reason) {
			t.Errorf("%s: record %v, want storage.operation reason %s", name, rec, reason)
		}
	}
	put := func(size int64, ct string, body *countingBody) error {
		req := PutObjectRequest{ContentType: ct, Size: size}
		if body != nil {
			req.Body = body
		}
		_, err := h.svc.PutObject(rctx, tc, req)
		return err
	}
	for _, c := range []struct {
		name   string
		size   int64
		ct     string
		body   *countingBody
		field  string
		reason Reason
	}{
		{"negative size", -1, "", &countingBody{Reader: strings.NewReader("x")}, "object_size_required", reasonInputInvalid},
		{"min int size", -1 << 63, "", &countingBody{Reader: strings.NewReader("x")}, "object_size_required", reasonInputInvalid},
		{"above the maximum", 26214401, "", &countingBody{Reader: strings.NewReader("x")}, "object_too_large", reasonInputInvalid},
		{"above the hard cap", 1073741825, "", &countingBody{Reader: strings.NewReader("x")}, "object_too_large", reasonInputInvalid},
		{"content type with parameters", 1, "text/plain; charset=utf-8", &countingBody{Reader: strings.NewReader("x")}, "invalid_content_type", reasonInputInvalid},
		{"content type without subtype", 1, "text", &countingBody{Reader: strings.NewReader("x")}, "invalid_content_type", reasonInputInvalid},
		{"missing body", 1, "", nil, "missing_body", reasonInputInvalid},
		{"zero size with data", 0, "", &countingBody{Reader: strings.NewReader("x")}, "object_size_mismatch", reasonInputInvalid},
	} {
		err := put(c.size, c.ct, c.body)
		check("Put "+c.name, err, CodeInvalidInput, c.field, c.reason)
		if c.body != nil && c.body.closes.Load() != 1 {
			t.Errorf("Put %s: Body closed %d times, want exactly once", c.name, c.body.closes.Load())
		}
	}
	for _, c := range []struct {
		name string
		req  SignedGetURLRequest
	}{
		{"neither form", SignedGetURLRequest{Object: obj}},
		{"both forms", SignedGetURLRequest{Object: obj, DefaultTTL: true, TTLSeconds: 60}},
		{"zero", SignedGetURLRequest{Object: obj, TTLSeconds: 0}},
		{"negative", SignedGetURLRequest{Object: obj, TTLSeconds: -1}},
		{"above the maximum", SignedGetURLRequest{Object: obj, TTLSeconds: 301}},
		{"overflow", SignedGetURLRequest{Object: obj, TTLSeconds: 9223372037}},
		{"max int64", SignedGetURLRequest{Object: obj, TTLSeconds: 1<<63 - 1}},
	} {
		_, err := h.svc.SignedGetURL(rctx, tc, c.req)
		check("Sign "+c.name, err, CodeInvalidInput, "invalid_ttl", reasonInputInvalid)
	}
	_, err := h.svc.SignedGetURL(rctx, tc, SignedGetURLRequest{DefaultTTL: true})
	check("Sign zero object", err, CodeInvalidInput, "invalid_id", reasonInputInvalid)
	_, err = h.svc.ObjectMetadata(rctx, tc, tenancy.ObjectID{})
	check("Metadata zero object", err, CodeInvalidInput, "invalid_id", reasonInputInvalid)
	_, _, err = h.svc.GetObject(rctx, tc, tenancy.ObjectID{})
	check("Get zero object", err, CodeInvalidInput, "invalid_id", reasonInputInvalid)
	_, err = h.svc.DeleteObject(rctx, tc, tenancy.ObjectID{})
	check("Delete zero object", err, CodeInvalidInput, "invalid_id", reasonInputInvalid)

	// A forged context is context_denied, and the Body is still closed exactly once.
	b := &countingBody{Reader: strings.NewReader("x")}
	_, err = h.svc.PutObject(rctx, &TenantContext{}, PutObjectRequest{Size: 1, Body: b})
	check("Put forged context", err, CodeContextDenied, "", reasonContextForged)
	if b.closes.Load() != 1 {
		t.Errorf("forged context: Body closed %d times", b.closes.Load())
	}
	if rec := h.logs.last(t); rec["actor_id"] != "" || rec["organization_id"] != "" {
		t.Errorf("an unbound call carries attribution: %v", rec)
	}

	if n, c := h.db.begins.Load(), store.calls.Load(); n != 0 || c != 0 {
		t.Errorf("validation reached the database (%d BeginTx) or the provider (%d calls)", n, c)
	}

	// Without a store every storage operation fails closed.
	plain := newCtxHarness()
	prctx, pend := plain.begin(t, context.Background(), "cred-A")
	defer pend()
	ptc := plain.resolved(t, prctx, tenancy.GroupID{})
	if _, err := plain.svc.ObjectMetadata(prctx, ptc, obj); CodeOf(err) != CodeUnavailable || plain.logs.last(t)["reason"] != "objects_unconfigured" {
		t.Errorf("unconfigured: %v %v", err, plain.logs.last(t))
	}
	b = &countingBody{Reader: strings.NewReader("x")}
	if _, err := plain.svc.PutObject(prctx, ptc, PutObjectRequest{Size: 1, Body: b}); CodeOf(err) != CodeUnavailable || b.closes.Load() != 1 {
		t.Errorf("unconfigured put: %v, closes %d", err, b.closes.Load())
	}
	if plain.db.begins.Load() != 0 {
		t.Error("unconfigured operations reached the database")
	}
}

func TestObjectConfigBounds(t *testing.T) {
	good := testObjectConfig
	if !good.valid() {
		t.Fatal("the default configuration is invalid")
	}
	for name, c := range map[string]ObjectConfig{
		"zero max":             {MaxObjectBytes: 0, DefaultTTL: time.Minute, MaxTTL: time.Minute, TransferTimeout: time.Minute},
		"max above hard cap":   {MaxObjectBytes: tenancy.MaxObjectBytes + 1, DefaultTTL: time.Minute, MaxTTL: time.Minute, TransferTimeout: time.Minute},
		"max TTL above 300 s":  {MaxObjectBytes: 1, DefaultTTL: time.Minute, MaxTTL: 301 * time.Second, TransferTimeout: time.Minute},
		"default above max":    {MaxObjectBytes: 1, DefaultTTL: 2 * time.Minute, MaxTTL: time.Minute, TransferTimeout: time.Minute},
		"fractional TTL":       {MaxObjectBytes: 1, DefaultTTL: 1500 * time.Millisecond, MaxTTL: time.Minute, TransferTimeout: time.Minute},
		"zero transfer":        {MaxObjectBytes: 1, DefaultTTL: time.Minute, MaxTTL: time.Minute},
		"transfer above 900 s": {MaxObjectBytes: 1, DefaultTTL: time.Minute, MaxTTL: time.Minute, TransferTimeout: 901 * time.Second},
	} {
		if c.valid() {
			t.Errorf("%s accepted", name)
		}
		if s := NewServiceWithObjects(&spyDB{}, nil, nil, &spyStore{}, c); s.objects != nil {
			t.Errorf("%s: NewServiceWithObjects kept the store", name)
		}
	}
	if s := NewServiceWithObjects(&spyDB{}, nil, nil, nil, good); s.objects != nil {
		t.Error("a nil store is configured")
	}
	if s := NewServiceWithObjects(&spyDB{}, nil, nil, &spyStore{}, good); s.objects == nil || s.hooks != nil {
		t.Error("a valid configuration is refused, or test seams are set in a production construction")
	}
}

// --- U-OBJ-READER -----------------------------------------------------------------------------------------------------

// countingReader counts the bytes read from src.
type countingReader struct {
	src io.Reader
	n   atomic.Int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.src.Read(p)
	c.n.Add(int64(n))
	return n, err
}

type infinite struct{}

func (infinite) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}

func TestExactReader(t *testing.T) {
	data := bytes.Repeat([]byte("0123456789"), 1000)
	for _, size := range []int64{1, 2, 10, int64(len(data))} {
		src := &countingReader{src: bytes.NewReader(data[:size])}
		r := newExactReader(src, size)
		got, err := io.ReadAll(r)
		v, sum := r.result()
		if err != nil || int64(len(got)) != size || v != verdictExact || sum != sha256.Sum256(data[:size]) {
			t.Errorf("exact %d: %d bytes, %v, verdict %d", size, len(got), err, v)
		}
	}
	// Short: an error, never more than Size-1 bytes released.
	r := newExactReader(bytes.NewReader(data[:9]), 10)
	got, err := io.ReadAll(r)
	if err == nil || r.verdict() != verdictShort || len(got) > 9 {
		t.Errorf("short: %d bytes, %v, verdict %d", len(got), err, r.verdict())
	}
	r = newExactReader(bytes.NewReader(nil), 1)
	if got, err := io.ReadAll(r); err == nil || len(got) != 0 || r.verdict() != verdictShort {
		t.Errorf("empty source for size 1: %d %v %d", len(got), err, r.verdict())
	}
	// Long: an error with exactly Size-1 bytes released; an infinite source is read for exactly Size+1 bytes.
	for _, size := range []int64{1, 2, 10, 70000} {
		src := &countingReader{src: infinite{}}
		r := newExactReader(src, size)
		got, err := io.ReadAll(r)
		if err == nil || r.verdict() != verdictLong || int64(len(got)) != size-1 || src.n.Load() != size+1 {
			t.Errorf("long %d: released %d, read %d, %v, verdict %d", size, len(got), src.n.Load(), err, r.verdict())
		}
	}
	// Holdback: the final byte waits for the lookahead's EOF.
	pr, pw := io.Pipe()
	r = newExactReader(pr, 3)
	go func() { _, _ = pw.Write([]byte("abc")) }()
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if err != nil || n != 2 || string(buf[:2]) != "ab" {
		t.Fatalf("first read %d %v", n, err)
	}
	done := make(chan struct{})
	var fn int
	var ferr error
	go func() { fn, ferr = r.Read(buf); close(done) }()
	select {
	case <-done:
		t.Fatal("the final byte was released before EOF")
	case <-time.After(100 * time.Millisecond):
	}
	_ = pw.Close()
	<-done
	if fn != 1 || ferr != nil || buf[0] != 'c' {
		t.Fatalf("final read %d %v", fn, ferr)
	}
	if n, err := r.Read(buf); n != 0 || err != io.EOF || r.verdict() != verdictExact {
		t.Fatalf("after the final byte: %d %v %d", n, err, r.verdict())
	}
	// A failing source is source_error.
	r = newExactReader(io.MultiReader(strings.NewReader("ab"), errReader{}), 5)
	if _, err := io.ReadAll(r); err == nil || r.verdict() != verdictSourceError {
		t.Errorf("source error: %v %d", err, r.verdict())
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("source broke") }

// pipeBody is an io.Pipe reader whose Close unblocks a blocked Read (a conforming source), counting Close calls.
type pipeBody struct {
	*io.PipeReader
	closes atomic.Int64
}

func (p *pipeBody) Close() error { p.closes.Add(1); return p.PipeReader.Close() }

// stuckBody ignores Close: a blocked Read returns only when the test releases it (residual X6).
type stuckBody struct {
	release chan struct{}
	closes  atomic.Int64
}

func (s *stuckBody) Read([]byte) (int, error) { <-s.release; return 0, io.EOF }
func (s *stuckBody) Close() error             { s.closes.Add(1); return nil }

// sdkLikeStore consumes the body in its own goroutine, as an HTTP transport does, and returns when the body ends or ctx
// is done (possibly while a Read is still blocked).
func sdkLikeStore() *spyStore {
	return &spyStore{put: func(ctx context.Context, _ objectstore.Key, body io.Reader, _ int64) error {
		done := make(chan error, 1)
		go func() {
			_, err := io.Copy(io.Discard, body)
			done <- err
		}()
		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}}
}

func waitGoroutines(t *testing.T, baseline int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for runtime.NumGoroutine() > baseline && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > baseline {
		t.Errorf("%d goroutines, baseline %d: a transfer worker leaked", n, baseline)
	}
}

func TestOwnedUploadSource(t *testing.T) {
	key, _ := objectstore.NewKey("11111111-2222-4333-8444-555555555555", "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
	b := objectstore.Binding{Object: "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", Organization: "11111111-2222-4333-8444-555555555555"}
	s := newService(&spyDB{}, nil, nil)
	s.objects, s.ocfg = sdkLikeStore(), testObjectConfig
	const deadline = 300 * time.Millisecond

	run := func(name string, cancelParent bool, size int64, body io.ReadCloser, closes func() int64, want uploadVerdict) {
		t.Helper()
		base := runtime.NumGoroutine()
		parent, cancel := context.WithCancel(context.Background())
		src := newOwnedSource(body)
		d := deadline
		if cancelParent {
			d = time.Minute
			time.AfterFunc(deadline, cancel)
		}
		tctx, release := transferContext(parent, src, d)
		start := time.Now()
		v, _, _ := s.upload(tctx, src, key, size, "text/plain", b)
		elapsed := time.Since(start)
		release()
		cancel()
		_ = src.Close()
		if v != want {
			t.Errorf("%s: verdict %d, want %d", name, v, want)
		}
		if elapsed > deadline+sourceReapGrace+time.Second {
			t.Errorf("%s: took %v", name, elapsed)
		}
		if closes() != 1 {
			t.Errorf("%s: Body closed %d times, want exactly once", name, closes())
		}
		waitGoroutines(t, base)
	}

	// (a) A blocked first Read (an io.Pipe without writer), unblocked by the deadline and by cancellation.
	for _, cancelParent := range []bool{false, true} {
		pr, _ := io.Pipe()
		p := &pipeBody{PipeReader: pr}
		run(fmt.Sprintf("blocked first read (cancel=%v)", cancelParent), cancelParent, 10, p, p.closes.Load, verdictSourceError)
	}
	// (b) A blocked final lookahead: all Size bytes, then neither EOF nor more data.
	pr, pw := io.Pipe()
	go func() { _, _ = pw.Write([]byte("0123456789")) }()
	p := &pipeBody{PipeReader: pr}
	run("blocked lookahead", false, 10, p, p.closes.Load, verdictSourceError)
	// An exact stream through the same path.
	eb := &countingBody{Reader: strings.NewReader("0123456789")}
	run("exact", false, 10, eb, eb.closes.Load, verdictExact)
	// (d) A non-conforming source that ignores Close: source_not_released within the grace; then released by the test.
	st := &stuckBody{release: make(chan struct{})}
	base := runtime.NumGoroutine()
	src := newOwnedSource(st)
	tctx, release := transferContext(context.Background(), src, deadline)
	start := time.Now()
	v, _, _ := s.upload(tctx, src, key, 10, "text/plain", b)
	release()
	if v != verdictNotReleased || time.Since(start) > deadline+sourceReapGrace+time.Second || st.closes.Load() != 1 {
		t.Errorf("non-conforming source: verdict %d after %v, closes %d", v, time.Since(start), st.closes.Load())
	}
	close(st.release)
	waitGoroutines(t, base)
}

// releasingBody is a conforming upload source that returns data as Close releases it: the first Read delivers head at
// once; the second blocks until Close and then returns tail with tailErr; later Reads return EOF. It is read by one
// goroutine at a time (the transport's).
type releasingBody struct {
	head, tail []byte
	tailErr    error
	stage      int
	blocked    chan struct{} // closed when the blocking Read starts
	release    chan struct{}
	once       sync.Once
	closes     atomic.Int64
}

func newReleasingBody(head, tail []byte, tailErr error) *releasingBody {
	return &releasingBody{head: head, tail: tail, tailErr: tailErr, blocked: make(chan struct{}), release: make(chan struct{})}
}

func (b *releasingBody) Read(p []byte) (int, error) {
	switch b.stage {
	case 0:
		b.stage = 1
		return copy(p, b.head), nil
	case 1:
		b.stage = 2
		close(b.blocked)
		<-b.release
		return copy(p, b.tail), b.tailErr
	}
	return 0, io.EOF
}

func (b *releasingBody) Close() error {
	b.closes.Add(1)
	b.once.Do(func() { close(b.release) })
	return nil
}

// asyncStore consumes the body in its own goroutine and returns as soon as ctx is done, without waiting for that
// goroutine (an asynchronous transport); finished is closed when the goroutine's last Read has returned.
func asyncStore(finished chan struct{}) *spyStore {
	return &spyStore{put: func(ctx context.Context, _ objectstore.Key, body io.Reader, _ int64) error {
		done := make(chan error, 1)
		go func() {
			_, err := io.Copy(io.Discard, body)
			done <- err
			close(finished)
		}()
		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}}
}

// TestUploadDigestSnapshot (correction C1): the digest leaves exactReader only with verdictExact, as a snapshot taken
// after the final hash write; an asynchronous transport that returns on cancellation while a conforming source releases
// data never makes upload read the hash concurrently with its write (run under -race), never publishes a digest without
// the exact verdict, and keeps close-once and the bounded return.
func TestUploadDigestSnapshot(t *testing.T) {
	data := []byte("0123456789abcdefghij")

	// No premature publication: while the final byte's lookahead is blocked, the verdict is pending and no digest is
	// published; after the data-bearing release with EOF, exact and the digest of all Size bytes.
	body := newReleasingBody(data[:9], data[9:10], io.EOF)
	r := newExactReader(body, 10)
	buf := make([]byte, 16)
	if n, err := r.Read(buf); n != 9 || err != nil {
		t.Fatalf("first read: %d %v", n, err)
	}
	done := make(chan struct{})
	var fn int
	var ferr error
	go func() { fn, ferr = r.Read(buf); close(done) }()
	<-body.blocked
	if v, d := r.result(); v != verdictPending || d != ([32]byte{}) {
		t.Fatalf("before the final byte: verdict %d, digest published %v", v, d != [32]byte{})
	}
	_ = body.Close()
	<-done
	if v, d := r.result(); fn != 1 || ferr != nil || v != verdictExact || d != sha256.Sum256(data[:10]) {
		t.Fatalf("after the final byte: %d %v, verdict %d, digest ok %v", fn, ferr, v, d == sha256.Sum256(data[:10]))
	}

	// Asynchronous transport, cancelled while the source is blocked; the source releases data as it is closed.
	key, _ := objectstore.NewKey("11111111-2222-4333-8444-555555555555", "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
	b := objectstore.Binding{Object: "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", Organization: "11111111-2222-4333-8444-555555555555"}
	s := newService(&spyDB{}, nil, nil)
	s.ocfg = testObjectConfig
	for _, c := range []struct {
		name       string
		size       int
		head, tail []byte
		tailErr    error
		mayExact   bool // the release completes the stream, so exact (with the full digest) is a permitted outcome
	}{
		{"mid-stream data released on close", 20, data[:4], data[4:7], nil, false},
		{"final byte released on close, then closed", 10, data[:9], data[9:10], nil, false},
		{"final byte released on close with EOF", 10, data[:9], data[9:10], io.EOF, true},
	} {
		exact := 0
		for i := 0; i < 50; i++ {
			base := runtime.NumGoroutine()
			finished := make(chan struct{})
			s.objects = asyncStore(finished)
			body := newReleasingBody(c.head, c.tail, c.tailErr)
			parent, cancel := context.WithCancel(context.Background())
			src := newOwnedSource(body)
			tctx, release := transferContext(parent, src, time.Minute)
			go func() { <-body.blocked; cancel() }()
			start := time.Now()
			v, sum, perr := s.upload(tctx, src, key, int64(c.size), "text/plain", b)
			elapsed := time.Since(start)
			release()
			cancel()
			select {
			case <-finished: // the transport goroutine's last hash write has happened during this test
			case <-time.After(3 * time.Second):
				t.Fatalf("%s: the transport goroutine did not finish", c.name)
			}
			switch {
			case v == verdictExact && c.mayExact:
				exact++
				if sum != sha256.Sum256(data[:c.size]) {
					t.Errorf("%s: exact with a wrong digest", c.name)
				}
			case v == verdictExact:
				t.Errorf("%s: exact verdict for an incomplete stream", c.name)
			case sum != [32]byte{}:
				t.Errorf("%s: verdict %d published a digest", c.name, v)
			}
			if perr == nil && v != verdictExact {
				t.Errorf("%s: Put succeeded with verdict %d", c.name, v)
			}
			if body.closes.Load() != 1 {
				t.Errorf("%s: Body closed %d times, want exactly once", c.name, body.closes.Load())
			}
			if elapsed > sourceReapGrace+time.Second {
				t.Errorf("%s: upload took %v", c.name, elapsed)
			}
			waitGoroutines(t, base)
		}
		t.Logf("%s: %d/50 exact", c.name, exact)
	}

	// The C1 window, deterministically: after cancellation the source releases data and its Read returns (so the reap
	// check passes), and the transport goroutine is then held by the sourceRead seam, before the digest write, until a
	// timer armed before upload fires (so nothing orders upload's own digest access before that write; under -race the
	// pre-C1 upload, which read the hash here unconditionally, is reported). upload must return within the hold with
	// neither the exact verdict nor a digest, and the late write must not change what it returned.
	const hold = 300 * time.Millisecond
	for _, c := range []struct {
		name       string
		size       int
		head, tail []byte
		tailErr    error
	}{
		{"held mid-stream release", 20, data[:4], data[4:7], nil},
		{"held final-byte release with EOF", 10, data[:9], data[9:10], io.EOF},
	} {
		var armed atomic.Bool
		gate := make(chan struct{})
		s.hooks = &objectHooks{sourceRead: func() {
			if armed.Load() {
				<-gate
			}
		}}
		finished := make(chan struct{})
		s.objects = asyncStore(finished)
		body := newReleasingBody(c.head, c.tail, c.tailErr)
		parent, cancel := context.WithCancel(context.Background())
		src := newOwnedSource(body)
		tctx, release := transferContext(parent, src, time.Minute)
		go func() { <-body.blocked; armed.Store(true); cancel() }()
		timer := time.AfterFunc(hold, func() { close(gate) })
		start := time.Now()
		v, sum, _ := s.upload(tctx, src, key, int64(c.size), "text/plain", b)
		elapsed := time.Since(start)
		release()
		cancel()
		if elapsed >= hold {
			t.Errorf("%s: upload took %v, so the held window was not exercised", c.name, elapsed)
		}
		if v == verdictExact || sum != ([32]byte{}) {
			t.Errorf("%s: upload returned verdict %d with a digest %v while the final write was pending", c.name, v, sum != [32]byte{})
		}
		select {
		case <-finished:
		case <-time.After(hold + 3*time.Second):
			t.Fatalf("%s: the transport goroutine did not finish", c.name)
		}
		timer.Stop()
		if body.closes.Load() != 1 {
			t.Errorf("%s: Body closed %d times, want exactly once", c.name, body.closes.Load())
		}
		s.hooks = nil
	}

	// Data-bearing conforming sources (final data together with EOF, one byte per Read, half reads) through the reader
	// and through upload with an SDK-like transport: exact, with the digest of exactly the Size bytes.
	big := bytes.Repeat([]byte("0123456789"), 7000)
	s.objects = sdkLikeStore()
	for _, mk := range []struct {
		name string
		f    func(io.Reader) io.Reader
	}{
		{"data with EOF", iotest.DataErrReader},
		{"one byte per read", iotest.OneByteReader},
		{"half reads", iotest.HalfReader},
	} {
		for _, size := range []int{1, 2, 10, len(big)} {
			r := newExactReader(mk.f(bytes.NewReader(big[:size])), int64(size))
			got, err := io.ReadAll(r)
			if v, d := r.result(); err != nil || len(got) != size || v != verdictExact || d != sha256.Sum256(big[:size]) {
				t.Errorf("reader %s %d: %d bytes, %v, verdict %d", mk.name, size, len(got), err, v)
			}
			cb := &countingBody{Reader: mk.f(bytes.NewReader(big[:size]))}
			src := newOwnedSource(cb)
			tctx, release := transferContext(context.Background(), src, time.Minute)
			v, sum, perr := s.upload(tctx, src, key, int64(size), "text/plain", b)
			release()
			if perr != nil || v != verdictExact || sum != sha256.Sum256(big[:size]) || cb.closes.Load() != 1 {
				t.Errorf("upload %s %d: %v, verdict %d, digest ok %v, closes %d", mk.name, size, perr, v,
					sum == sha256.Sum256(big[:size]), cb.closes.Load())
			}
		}
	}
}

func TestEmptyUploadProbe(t *testing.T) {
	for _, c := range []struct {
		name string
		body func() (io.ReadCloser, func())
		want uploadVerdict
	}{
		{"empty", func() (io.ReadCloser, func()) { return &countingBody{Reader: strings.NewReader("")}, func() {} }, verdictExact},
		{"extra data", func() (io.ReadCloser, func()) { return &countingBody{Reader: strings.NewReader("x")}, func() {} }, verdictLong},
		{"blocked", func() (io.ReadCloser, func()) {
			pr, _ := io.Pipe()
			return &pipeBody{PipeReader: pr}, func() {}
		}, verdictSourceError},
		{"non-conforming", func() (io.ReadCloser, func()) {
			st := &stuckBody{release: make(chan struct{})}
			return st, func() { close(st.release) }
		}, verdictNotReleased},
	} {
		base := runtime.NumGoroutine()
		body, cleanup := c.body()
		src := newOwnedSource(body)
		tctx, release := transferContext(context.Background(), src, 200*time.Millisecond)
		start := time.Now()
		v := probeEmpty(tctx, src)
		release()
		_ = src.Close()
		if v != c.want || time.Since(start) > 200*time.Millisecond+sourceReapGrace+time.Second {
			t.Errorf("%s: verdict %d after %v", c.name, v, time.Since(start))
		}
		cleanup()
		waitGoroutines(t, base)
	}
}

// Streaming 25 MiB through a discarding store allocates far less than the object: there is no buffering.
func TestUploadStreamsWithoutBuffering(t *testing.T) {
	key, _ := objectstore.NewKey("11111111-2222-4333-8444-555555555555", "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
	b := objectstore.Binding{Object: "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", Organization: "11111111-2222-4333-8444-555555555555"}
	s := newService(&spyDB{}, nil, nil)
	s.objects, s.ocfg = sdkLikeStore(), testObjectConfig
	const size = 25 << 20
	body := &countingBody{Reader: io.LimitReader(infinite{}, size)}
	src := newOwnedSource(body)
	tctx, release := transferContext(context.Background(), src, time.Minute)
	defer release()
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	v, _, err := s.upload(tctx, src, key, size, "application/octet-stream", b)
	runtime.ReadMemStats(&after)
	if v != verdictExact || err != nil {
		t.Fatalf("upload: verdict %d, %v", v, err)
	}
	if grown := after.TotalAlloc - before.TotalAlloc; grown > 4<<20 {
		t.Errorf("streaming 25 MiB allocated %d bytes (bound 4 MiB)", grown)
	}
}

func TestVerifyingReader(t *testing.T) {
	data := []byte("stored object bytes")
	sum := sha256.Sum256(data)
	mk := func(b []byte, size int64, want [32]byte) (*verifyingReader, *bool) {
		canceled := false
		return &verifyingReader{body: io.NopCloser(bytes.NewReader(b)), cancel: func() { canceled = true }, remaining: size, want: want,
			hash: sha256.New()}, &canceled
	}
	r, canceled := mk(data, int64(len(data)), sum)
	got, err := io.ReadAll(r)
	if err != nil || !bytes.Equal(got, data) {
		t.Errorf("verified read: %q %v", got, err)
	}
	_ = r.Close()
	_ = r.Close()
	if !*canceled {
		t.Error("Close does not cancel the transfer")
	}
	r, _ = mk(data, int64(len(data)), sha256.Sum256([]byte("other")))
	got, err = io.ReadAll(r)
	if CodeOf(err) != CodeInternal || !bytes.Equal(got, data) {
		t.Errorf("integrity mismatch: %d bytes, %v (want the bytes, then internal at EOF)", len(got), err)
	}
	r, _ = mk(data[:5], int64(len(data)), sum)
	if _, err := io.ReadAll(r); CodeOf(err) != CodeUnavailable {
		t.Errorf("short body: %v", err)
	}
	r, _ = mk(append(append([]byte{}, data...), "extra"...), int64(len(data)), sum)
	if got, err := io.ReadAll(r); err != nil || len(got) != len(data) {
		t.Errorf("bounded read: %d %v", len(got), err)
	}
	r, _ = mk(nil, 0, sha256.Sum256(nil))
	if got, err := io.ReadAll(r); err != nil || len(got) != 0 {
		t.Errorf("empty object: %v", err)
	}
}

func TestSignedURLRedaction(t *testing.T) {
	secret := "http://127.0.0.1:9100/bucket/objects/x/y?X-Amz-Signature=Cx3signaturecanary"
	u := SignedURL{url: secret, ExpiresAt: time.Now(), TTL: time.Minute}
	var logs bytes.Buffer
	slog.New(slog.NewJSONHandler(&logs, nil)).Info("x", "url", u)
	for _, out := range []string{u.String(), u.GoString(), fmt.Sprintf("%v", u), fmt.Sprintf("%+v", u), fmt.Sprintf("%#v", u), fmt.Sprint(u),
		fmt.Sprintf("%s", u), logs.String()} {
		if strings.Contains(out, "Cx3signaturecanary") {
			t.Errorf("the signed URL leaks: %s", out)
		}
	}
	if u.URL() != secret {
		t.Error("URL does not return the URL")
	}
}

// --- U-OBJ-POLICY --------------------------------------------------------------------------------------------------------

func TestObjectStoragePolicyMapping(t *testing.T) {
	org, g1, g2 := uid[tenancy.OrganizationID](0x01), uid[tenancy.GroupID](0x11), uid[tenancy.GroupID](0x12)
	me, other := uid[tenancy.UserID](0x21), uid[tenancy.UserID](0x22)
	facts := func(orgRole tenancy.OrganizationRole, groupRoles map[tenancy.GroupID]tenancy.GroupRole) Facts {
		f := Facts{Organization: tenancy.Organization{ID: org}, Groups: []tenancy.Group{{ID: g1, OrganizationID: org}, {ID: g2, OrganizationID: org}}}
		f.OrganizationBindings = []tenancy.BindingFact{{OrganizationID: org, Role: string(orgRole), Source: tenancy.SourceMembership}}
		for g, r := range groupRoles {
			f.GroupBindings = append(f.GroupBindings, tenancy.BindingFact{OrganizationID: org, GroupID: g, Role: string(r), Source: tenancy.SourceMembership})
		}
		return f
	}
	sel := contextSelection{Organization: org}
	type c struct {
		a       objectAction
		orgRole tenancy.OrganizationRole
		gr      tenancy.GroupRole
		t       objectTarget
		sel     contextSelection
		ok      bool
		reason  Reason
	}
	orgObj := objectTarget{creator: other}
	mine, others := objectTarget{group: g1, creator: me}, objectTarget{group: g1, creator: other}
	for i, x := range []c{
		{actionPut, tenancy.OrgOwner, "", orgObj, sel, true, reasonAllowed},
		{actionPut, tenancy.OrgAdmin, "", orgObj, sel, true, reasonAllowed},
		{actionPut, tenancy.OrgMember, tenancy.GroupOwner, orgObj, sel, false, reasonRoleNotPermitted},
		{actionPut, tenancy.OrgMember, tenancy.GroupMember, objectTarget{group: g1}, sel, true, reasonAllowed},
		{actionPut, tenancy.OrgMember, "", objectTarget{group: g1}, sel, false, reasonRoleNotPermitted},
		{actionPut, tenancy.OrgAdmin, "", objectTarget{group: g1}, contextSelection{Organization: org, Group: g2}, false, reasonTargetOutsideContext},
		{actionDelete, tenancy.OrgOwner, "", orgObj, sel, true, reasonAllowed},
		{actionDelete, tenancy.OrgMember, tenancy.GroupAdmin, orgObj, sel, false, reasonRoleNotPermitted},
		{actionDelete, tenancy.OrgMember, tenancy.GroupMember, mine, sel, true, reasonAllowed},
		{actionDelete, tenancy.OrgMember, tenancy.GroupMember, others, sel, false, reasonNotCreator},
		{actionDelete, tenancy.OrgMember, tenancy.GroupAdmin, others, sel, true, reasonAllowed},
		{actionDelete, tenancy.OrgMember, tenancy.GroupOwner, others, sel, true, reasonAllowed},
		// G060 (Spec R3): a Group object is content; an Organization owner or admin without a role in its Group is denied
		// every action under either value of the Group-management override.
		{actionDelete, tenancy.OrgAdmin, "", others, sel, false, reasonRoleNotPermitted},
		{actionDelete, tenancy.OrgOwner, "", mine, sel, false, reasonRoleNotPermitted},
		{actionPut, tenancy.OrgOwner, "", objectTarget{group: g1}, sel, false, reasonRoleNotPermitted},
		{actionPut, tenancy.OrgAdmin, "", objectTarget{group: g1}, sel, false, reasonRoleNotPermitted},
		{actionRead, tenancy.OrgOwner, "", others, sel, false, reasonRoleNotPermitted},
		{actionRead, tenancy.OrgAdmin, "", others, contextSelection{Organization: org, Group: g1}, false, reasonRoleNotPermitted},
		{actionRead, tenancy.OrgAdmin, tenancy.GroupMember, others, sel, true, reasonAllowed},
		{actionPut, tenancy.OrgAdmin, tenancy.GroupMember, objectTarget{group: g1}, sel, true, reasonAllowed},
		{actionDelete, tenancy.OrgAdmin, tenancy.GroupMember, others, sel, false, reasonNotCreator},
		{actionDelete, tenancy.OrgOwner, tenancy.GroupAdmin, others, sel, true, reasonAllowed},
		{actionRead, tenancy.OrgOwner, "", orgObj, sel, true, reasonAllowed},
		{actionDelete, tenancy.OrgMember, "", mine, sel, false, reasonRoleNotPermitted},
		{actionDelete, tenancy.OrgMember, tenancy.GroupMember, objectTarget{group: g1}, sel, false, reasonNotCreator},
		{actionRead, tenancy.OrgMember, "", orgObj, sel, true, reasonAllowed},
		{actionRead, tenancy.OrgMember, "", others, sel, false, reasonRoleNotPermitted},
		{actionRead, tenancy.OrgMember, tenancy.GroupMember, others, sel, true, reasonAllowed},
		{actionRead, tenancy.OrgMember, tenancy.GroupMember, others, contextSelection{Organization: org, Group: g2}, false, reasonTargetOutsideContext},
		{actionRead, tenancy.OrgMember, "", objectTarget{group: uid[tenancy.GroupID](0x99)}, sel, false, reasonTargetNotFound},
	} {
		roles := map[tenancy.GroupID]tenancy.GroupRole{}
		if x.gr != "" {
			roles[g1] = x.gr
		}
		ok, r := decideObject(x.a, x.sel, facts(x.orgRole, roles), x.t, me)
		if ok != x.ok || r != x.reason {
			t.Errorf("case %d (%s %s/%s): %v %s, want %v %s", i, x.a, x.orgRole, x.gr, ok, r, x.ok, x.reason)
		}
	}
	// Creator equality never grants scope: the creator without a current role is denied, and a member's own-delete
	// needs the exact creator.
	if ok, _ := decideObject(actionDelete, sel, facts(tenancy.OrgMember, nil), mine, me); ok {
		t.Error("creator equality granted scope")
	}
	// G060: the Group-management override never confers Group content (read, put or delete) on Organization staff.
	for _, role := range []tenancy.OrganizationRole{tenancy.OrgOwner, tenancy.OrgAdmin} {
		for _, a := range []objectAction{actionRead, actionPut, actionDelete} {
			f := facts(role, nil)
			f.Policy.GroupManagementOverride = true
			if ok, r := decideObject(a, sel, f, others, me); ok || r != reasonRoleNotPermitted {
				t.Errorf("%s %s of a Group object under the override: %v %s", role, a, ok, r)
			}
		}
	}
	// The objects facts are bounded: never every binding (Groups is never nil).
	tc := &TenantContext{organization: org, subject: me}
	if q := objectFactQuery(tc, tenancy.GroupID{}, false); q.Groups == nil || len(q.Groups) != 0 {
		t.Errorf("Organization-object facts: %+v", q.Groups)
	}
	tc.group = g2
	if q := objectFactQuery(tc, g1, true); len(q.Groups) != 2 || q.Groups[0] != g1 || q.Groups[1] != g2 || !q.Lock {
		t.Errorf("Group-object facts: %+v", q)
	}
	if q := objectFactQuery(tc, g2, false); len(q.Groups) != 1 {
		t.Errorf("selected owning Group read twice: %+v", q.Groups)
	}
}

// TestObjectPolicyOracle (U-OBJ-POLICY): the storage decision and the context check on synthetic facts derived from
// the independent oracle's fixture (objects_expected_test.go), for every service case that the facts alone decide:
// binding faults, session and user liveness and raw selector parsing need the real services and run in B-OBJ-IT.
// Synthetic facts model what the facade reads under row-level security: the Organization row and its binding only
// in a valid context, Group rows only when selectable/visible, and the object row only within the read scope.
func TestObjectPolicyOracle(t *testing.T) {
	fx := objectsExpectedFixture()
	m := newOracleModel(fx)
	n, skipped := 0, 0
	for _, c := range objectsExpectedCases() {
		if c.DBRole != oeDBNone || c.Ctx.Fault != oeFaultNone || c.RawObjectID != "" || c.RawGroupID != "" || (c.Action != oeActPut && c.Object == "") {
			skipped++
			continue
		}
		u := m.users[c.Actor]
		if u.State != oeUserActive || u.Session != oeSessionLive {
			skipped++
			continue
		}
		n++
		got := m.decide(t, c)
		if got != c.Expect {
			t.Errorf("ORACLE-DISAGREEMENT %s (%s): decision %s, oracle %s (%s)", c.ID, c.Action, got, c.Expect, c.Why)
		}
	}
	t.Logf("U-OBJ-POLICY: %d oracle cases decided on synthetic facts, %d need the real services", n, skipped)
	if n < 150 {
		t.Errorf("only %d oracle cases evaluated", n)
	}
}

// oracleModel maps the oracle fixture onto synthetic facts.
type oracleModel struct {
	fx     oeFixture
	orgs   map[string]tenancy.OrganizationID
	groups map[string]oeGroup
	gids   map[string]tenancy.GroupID
	users  map[string]oeUser
	uids   map[string]tenancy.UserID
	objs   map[string]oeObject
}

func mustParse[T ~[16]byte](s string) T {
	var id T
	b := strings.ReplaceAll(s, "-", "")
	for i := 0; i < 16; i++ {
		var v byte
		_, _ = fmt.Sscanf(b[2*i:2*i+2], "%02x", &v)
		id[i] = v
	}
	return id
}

func newOracleModel(fx oeFixture) *oracleModel {
	m := &oracleModel{fx: fx, orgs: map[string]tenancy.OrganizationID{}, groups: map[string]oeGroup{}, gids: map[string]tenancy.GroupID{},
		users: map[string]oeUser{}, uids: map[string]tenancy.UserID{}, objs: map[string]oeObject{}}
	for _, o := range fx.Organizations {
		m.orgs[o.Key] = mustParse[tenancy.OrganizationID](o.ID)
	}
	for _, g := range fx.Groups {
		m.groups[g.Key], m.gids[g.Key] = g, mustParse[tenancy.GroupID](g.ID)
	}
	for _, u := range fx.Users {
		m.users[u.Key], m.uids[u.Key] = u, mustParse[tenancy.UserID](u.ID)
	}
	for _, o := range fx.Objects {
		m.objs[o.Key] = o
	}
	return m
}

// role returns the actor's current role in an Organization or Group scope key, or "".
func (m *oracleModel) role(u oeUser, scope oeScopeType, key string) string {
	for _, b := range u.Bindings {
		if b.Scope == scope && b.ScopeKey == key {
			return string(b.Role)
		}
	}
	return ""
}

// facts returns what the facade's fact read sees under row-level security for the actor in the context, with the
// candidate Groups of the query.
func (m *oracleModel) facts(u oeUser, orgKey, selKey string, candidates []string) (Facts, bool) {
	var f Facts
	orgRole := m.role(u, oeScopeOrganization, orgKey)
	_, orgExists := m.orgs[orgKey]
	valid := orgExists && orgRole != ""
	if valid && selKey != "" {
		g, ok := m.groups[selKey]
		valid = ok && g.OrganizationKey == orgKey && (orgRole == "org_owner" || orgRole == "org_admin" || m.role(u, oeScopeGroup, selKey) != "")
	}
	if !valid { // rls_org_role is NULL: no tenant row is visible
		return f, false
	}
	org := m.orgs[orgKey]
	f.Organization = tenancy.Organization{ID: org}
	f.OrganizationBindings = []tenancy.BindingFact{{OrganizationID: org, Role: orgRole, Source: tenancy.SourceMembership}}
	mgr := orgRole == "org_owner" || orgRole == "org_admin"
	for _, key := range candidates {
		g, ok := m.groups[key]
		if !ok || g.OrganizationKey != orgKey || (selKey != "" && key != selKey) {
			continue
		}
		gr := m.role(u, oeScopeGroup, key)
		if gr != "" {
			f.GroupBindings = append(f.GroupBindings, tenancy.BindingFact{OrganizationID: org, GroupID: m.gids[key], Role: gr, Source: tenancy.SourceMembership})
		}
		if mgr || gr != "" {
			f.Groups = append(f.Groups, tenancy.Group{ID: m.gids[key], OrganizationID: org})
		}
	}
	return f, true
}

func (m *oracleModel) decide(t *testing.T, c oeCase) oeOutcome {
	t.Helper()
	u := m.users[c.Actor]
	if c.Ctx.Kind == oeCtxNone {
		return oeContextDenied
	}
	orgKey, selKey := c.Ctx.OrganizationKey, c.Ctx.GroupKey
	sel := contextSelection{Organization: m.orgs[orgKey]}
	if selKey != "" {
		sel.Group = m.gids[selKey]
	}
	subject := m.uids[c.Actor]
	if c.Action == oeActSign && c.TTL.Set {
		if _, ok := testObjectConfig.signedTTL(SignedGetURLRequest{Object: uid[tenancy.ObjectID](1), TTLSeconds: c.TTL.Seconds}); !ok {
			return oeInvalidInput
		}
	}
	var target objectTarget
	var owningKey string
	var obj oeObject
	found := false
	switch c.Action {
	case oeActPut:
		owningKey = c.PutGroupKey
		if owningKey != "" {
			if _, ok := m.gids[owningKey]; !ok {
				t.Fatalf("%s: unknown Group %s", c.ID, owningKey)
			}
			target.group = m.gids[owningKey]
		}
		target.creator = subject
	default:
		obj = m.objs[c.Object]
		// Scope discovery: the row is visible only in the active Organization and within the read scope.
		if obj.Stored && obj.OrganizationKey == orgKey {
			f0, valid := m.facts(u, orgKey, selKey, nil)
			if valid {
				orgRole := f0.OrganizationBindings[0].Role
				mgr := orgRole == "org_owner" || orgRole == "org_admin"
				found = obj.GroupKey == "" || ((selKey == "" || obj.GroupKey == selKey) && (mgr || m.role(u, oeScopeGroup, obj.GroupKey) != ""))
			}
		}
		if found {
			owningKey = obj.GroupKey
			if owningKey != "" {
				target.group = m.gids[owningKey]
			}
			target.creator = m.uids[obj.CreatorKey]
		}
	}
	candidates := []string{}
	if owningKey != "" {
		candidates = append(candidates, owningKey)
	}
	if selKey != "" && selKey != owningKey {
		candidates = append(candidates, selKey)
	}
	f, _ := m.facts(u, orgKey, selKey, candidates)
	if r := contextCheck(sel, f); r != "" {
		return oeContextDenied
	}
	action := map[oeAction]objectAction{oeActMetadata: actionRead, oeActGet: actionRead, oeActSign: actionRead, oeActPut: actionPut,
		oeActDelete: actionDelete}[c.Action]
	if c.Action != oeActPut && !found {
		return oeNotFound
	}
	if ok, r := decideObject(action, sel, f, target, subject); !ok {
		return oeOutcome(objectDenial(action, sel, f, target, subject, r).code)
	}
	if action == actionRead && obj.State != oeStateReady {
		return oeNotFound
	}
	return oeAllow
}

// --- U-OBJ-EVENT --------------------------------------------------------------------------------------------------------

// The storage.operation record contract (Plan r1 §4.8), mirrored from events.json: the catalogued fields and their
// closed values. infra/ is outside the core-check build context, so the harness validates the real records against
// events.json itself (logcheck in B-OBJ-IT); this pins the producer.
var (
	storageFields = []string{"time", "level", "msg", "operation", "outcome", "reason", "stage", "actor_id", "organization_id", "group_id",
		"object_id", "ttl_seconds", "physical", "request_id", "trace_id"}
	storageOps     = []string{opObjectPut, opObjectGet, opObjectMetadata, opObjectSign, opObjectDelete}
	storageStages  = []string{stageT1, stageProvider, stageVerify, stageT2, stageReconcile, stageCompensate}
	storageReasons = []Reason{reasonAllowed, reasonNotCreator, reasonObjectNotReady, reasonObjectsUnconfigured, reasonProviderUnavailable,
		reasonProviderOutcomeUnknown, reasonObjectBytesMissing, reasonBindingMismatch, reasonIntegrityMismatch, reasonCompensated,
		reasonCompensationFailed, reasonReconciledReady, reasonSourceNotReleased,
		reasonSessionInvalid, reasonContextForged, reasonScopeMissing, reasonScopeEnded, reasonScopeMismatch, reasonSubjectMismatch,
		reasonOrganizationNotPermitted, reasonGroupNotSelectable, reasonTargetNotFound, reasonTargetOutsideContext, reasonCapabilityUnknown,
		reasonRoleNotPermitted, reasonFactInconsistent, reasonInputInvalid, reasonStoreUnavailable, reasonAborted, reasonCanceled,
		reasonCommitOutcomeUnknown}
	uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

func TestStorageRecord(t *testing.T) {
	var logs bytes.Buffer
	s := newService(&spyDB{}, nil, slog.New(slog.NewJSONHandler(&logs, nil)))
	org, g, actor, obj := uid[tenancy.OrganizationID](1), uid[tenancy.GroupID](2), uid[tenancy.UserID](3), uid[tenancy.ObjectID](4)
	bound := func(op string) *storageRecord {
		return &storageRecord{op: op, bound: true, actor: actor, org: org, group: g, object: obj}
	}
	cases := []struct {
		rec   *storageRecord
		f     *failure
		level string
	}{
		{bound(opObjectPut), nil, "INFO"},
		{&storageRecord{op: opObjectPut, bound: true, actor: actor, org: org, stage: stageT1}, fail(CodeForbidden, reasonRoleNotPermitted), "WARN"},
		{&storageRecord{op: opObjectPut, bound: true, actor: actor, org: org, object: obj, stage: stageProvider},
			&failure{code: CodeInvalidInput, field: "object_size_mismatch", reason: reasonCompensated}, "WARN"},
		{&storageRecord{op: opObjectPut, bound: true, actor: actor, org: org, object: obj, stage: stageReconcile, reason: reasonReconciledReady}, nil, "INFO"},
		{&storageRecord{op: opObjectSign, bound: true, actor: actor, org: org, object: obj, ttl: 60}, nil, "INFO"},
		{&storageRecord{op: opObjectGet, bound: true, actor: actor, org: org, object: obj}, fail(CodeNotFound, reasonObjectBytesMissing), "ERROR"},
		{&storageRecord{op: opObjectDelete, bound: true, actor: actor, org: org, object: obj, physical: physicalConfirmed}, nil, "INFO"},
		{&storageRecord{op: opObjectDelete, bound: true, actor: actor, org: org, object: obj, stage: stageProvider, physical: physicalPending,
			reason: reasonProviderUnavailable}, nil, "INFO"},
		{&storageRecord{op: opObjectMetadata}, fail(CodeContextDenied, reasonContextForged), "WARN"},
		{&storageRecord{op: opObjectMetadata, bound: true, actor: actor, org: org}, fail(CodeUnavailable, reasonObjectsUnconfigured), "ERROR"},
	}
	for _, c := range cases {
		logs.Reset()
		s.logStorage(context.Background(), c.rec, c.f)
		var rec map[string]any
		if err := json.Unmarshal(logs.Bytes(), &rec); err != nil {
			t.Fatal(err)
		}
		for k, v := range rec {
			if !slices.Contains(storageFields, k) {
				t.Errorf("uncatalogued field %q", k)
			}
			if strings.HasSuffix(k, "_id") && k != "request_id" && k != "trace_id" && !uuidPattern.MatchString(fmt.Sprint(v)) {
				t.Errorf("%s is not a UUID: %v", k, v)
			}
		}
		if rec["msg"] != eventStorage || rec["level"] != c.level || !slices.Contains(storageOps, rec["operation"].(string)) ||
			!slices.Contains(storageReasons, Reason(rec["reason"].(string))) {
			t.Errorf("record %v (level %s)", rec, c.level)
		}
		if st, ok := rec["stage"].(string); ok && !slices.Contains(storageStages, st) {
			t.Errorf("stage %q", st)
		}
		switch rec["outcome"] {
		case outcomeAllow:
			if c.f != nil {
				t.Errorf("allow for a failure: %v", rec)
			}
		case outcomeDeny, outcomeFailed:
			if c.f == nil {
				t.Errorf("deny for a success: %v", rec)
			}
		default:
			t.Errorf("outcome %v", rec["outcome"])
		}
		if _, ok := rec["ttl_seconds"]; ok != (c.rec.op == opObjectSign && c.f == nil) {
			t.Errorf("ttl_seconds presence: %v", rec)
		}
		if _, ok := rec["physical"]; ok != (c.rec.op == opObjectDelete && c.f == nil) {
			t.Errorf("physical presence: %v", rec)
		}
		if _, ok := rec["actor_id"]; ok != c.rec.bound {
			t.Errorf("attribution presence: %v", rec)
		}
	}
}

// --- U-OBJ-SEAM: transaction budgets -------------------------------------------------------------------------------------

// budgetTx is a fake transaction: the context statement answers its one row; it records the deadlines of the
// statement and rollback contexts; Rollback may block until its context ends.
type budgetTx struct {
	pgx.Tx
	mu                sync.Mutex
	stmtDeadline      time.Time
	stmtHasDeadline   bool
	rbDeadline        time.Time
	rbHasDeadline     bool
	rbCanceledAtStart bool
	blockRollback     bool
	rollbacks, commit atomic.Int64
}

type budgetRow struct{}

func (budgetRow) Scan(dest ...any) error {
	for _, d := range dest {
		if p, ok := d.(*string); ok {
			*p = ""
		}
	}
	return nil
}

func (t *budgetTx) QueryRow(ctx context.Context, _ string, _ ...any) pgx.Row {
	t.mu.Lock()
	t.stmtDeadline, t.stmtHasDeadline = ctx.Deadline()
	t.mu.Unlock()
	return budgetRow{}
}

func (t *budgetTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (t *budgetTx) Conn() *pgx.Conn { return nil }

func (t *budgetTx) Commit(context.Context) error { t.commit.Add(1); return nil }

func (t *budgetTx) Rollback(ctx context.Context) error {
	t.rollbacks.Add(1)
	t.mu.Lock()
	t.rbDeadline, t.rbHasDeadline = ctx.Deadline()
	t.rbCanceledAtStart = ctx.Err() != nil
	block := t.blockRollback
	t.mu.Unlock()
	if block {
		<-ctx.Done()
		return ctx.Err()
	}
	return nil
}

type budgetDB struct{ tx *budgetTx }

func (d *budgetDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) { return d.tx, nil }

func TestStorageTransactionBudgets(t *testing.T) {
	ctxs := dbContext{user: uid[tenancy.UserID](1), organization: uid[tenancy.OrganizationID](2)}

	// A storage transaction's statements carry the 5 s decision deadline; its deferred rollback context carries the
	// 5 s cleanup deadline and is not canceled by the caller.
	tx := &budgetTx{}
	s := newService(&budgetDB{tx: tx}, nil, nil)
	parent, cancel := context.WithCancel(context.Background())
	start := time.Now()
	f := s.storageTx(parent, true, ctxs, func(ctx context.Context, ttx *trackedTx) *failure {
		tx.mu.Lock()
		tx.stmtHasDeadline, tx.stmtDeadline = false, time.Time{} // forget the context statement's deadline
		tx.mu.Unlock()
		_ = ttx.QueryRow(ctx, "statement of the transaction").Scan() // a statement inside fn carries the decision deadline
		cancel()                                                     // the caller goes away mid-transaction
		return fail(CodeUnavailable, reasonCanceled)
	})
	if f == nil || !tx.stmtHasDeadline || tx.stmtDeadline.Sub(start) > storageDecisionBudget+time.Second || tx.stmtDeadline.Sub(start) < storageDecisionBudget-time.Second {
		t.Errorf("decision budget: statement deadline %v (has %v)", tx.stmtDeadline.Sub(start), tx.stmtHasDeadline)
	}
	if !tx.rbHasDeadline || tx.rbCanceledAtStart || tx.rbDeadline.Sub(start) > storageCleanupBudget+time.Second || tx.rbDeadline.Sub(start) < storageCleanupBudget-time.Second {
		t.Errorf("cleanup budget: rollback deadline %v (has %v), canceled at start %v", tx.rbDeadline.Sub(start), tx.rbHasDeadline, tx.rbCanceledAtStart)
	}

	// A rollback that blocks until its context ends returns within the cleanup budget.
	tx = &budgetTx{blockRollback: true}
	s = newService(&budgetDB{tx: tx}, nil, nil)
	start = time.Now()
	_ = s.storageTx(context.Background(), true, ctxs, func(context.Context, *trackedTx) *failure { return fail(CodeUnavailable, reasonStoreUnavailable) })
	if el := time.Since(start); el > storageCleanupBudget+time.Second || el < storageCleanupBudget-time.Second {
		t.Errorf("a blocked rollback held the call for %v (budget %v)", el, storageCleanupBudget)
	}

	// The G007/G008 path (cleanupUnbounded) is unchanged: an undeadlined WithoutCancel rollback context.
	tx = &budgetTx{}
	s = newService(&budgetDB{tx: tx}, nil, nil)
	parent, cancel = context.WithCancel(context.Background())
	_ = s.inContextTx(parent, true, ctxs, cleanupUnbounded, func(*trackedTx) *failure {
		cancel()
		return fail(CodeUnavailable, reasonCanceled)
	})
	if tx.rbHasDeadline || tx.rbCanceledAtStart || tx.stmtHasDeadline {
		t.Errorf("cleanupUnbounded: rollback deadline %v, canceled %v; statement deadline %v", tx.rbHasDeadline, tx.rbCanceledAtStart, tx.stmtHasDeadline)
	}
	// After a successful COMMIT the deferred rollback still runs once (a no-op without I/O for a committed pgx.Tx).
	tx = &budgetTx{}
	s = newService(&budgetDB{tx: tx}, nil, nil)
	if f := s.storageTx(context.Background(), false, ctxs, func(context.Context, *trackedTx) *failure { return nil }); f != nil || tx.commit.Load() != 1 || tx.rollbacks.Load() != 1 {
		t.Errorf("commit path: %+v, commits %d, rollbacks %d", f, tx.commit.Load(), tx.rollbacks.Load())
	}
}
