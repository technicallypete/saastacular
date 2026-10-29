package access

import (
	"context"
	"crypto/sha256"
	"errors"
	"hash"
	"io"
	"log/slog"
	"sync"
	"time"
)

// --- the owned upload source (G009 Plan r1 §4.1, correction 1) ------------------------------------------------------

// errSourceClosed is what a Read of a closed owned source returns; it never reaches a caller.
var errSourceClosed = errors.New("access: upload source closed")

// ownedSource is PutObject's ownership of the caller's Body: Close is idempotent and closes Body exactly once; a Read
// after Close returns errSourceClosed without touching Body; in-flight Reads are counted, so PutObject can wait (for at
// most the reap grace) until none is still inside Body.Read after closing it. Body must make a blocked Read return
// promptly on Close (the documented upload source contract); a source that does not is residual X6.
type ownedSource struct {
	rc       io.ReadCloser
	once     sync.Once
	mu       sync.Mutex
	closed   bool // guarded by mu: a Read checks it and registers as in flight in one critical section
	inflight int
	idle     chan struct{} // closed when inflight drops to zero; nil while idle
}

func newOwnedSource(rc io.ReadCloser) *ownedSource { return &ownedSource{rc: rc} }

func (o *ownedSource) Read(p []byte) (int, error) {
	o.mu.Lock()
	if o.closed {
		o.mu.Unlock()
		return 0, errSourceClosed
	}
	if o.inflight == 0 {
		o.idle = make(chan struct{})
	}
	o.inflight++
	o.mu.Unlock()
	n, err := o.rc.Read(p)
	o.mu.Lock()
	o.inflight--
	if o.inflight == 0 {
		close(o.idle)
		o.idle = nil
	}
	closed := o.closed
	o.mu.Unlock()
	if closed && err == nil && n == 0 {
		err = errSourceClosed
	}
	return n, err
}

// Close closes Body exactly once.
func (o *ownedSource) Close() error {
	o.once.Do(func() {
		o.mu.Lock()
		o.closed = true
		o.mu.Unlock()
		_ = o.rc.Close()
	})
	return nil
}

// waitReleased waits up to grace until no Read is inside Body.Read; false means the source did not release (X6).
func (o *ownedSource) waitReleased(grace time.Duration) bool {
	o.mu.Lock()
	ch := o.idle
	o.mu.Unlock()
	if ch == nil {
		return true
	}
	t := time.NewTimer(grace)
	defer t.Stop()
	select {
	case <-ch:
		return true
	case <-t.C:
		return false
	}
}

// probeEmpty is P1 for a zero-length upload: one Read of up to one byte under ctx (the transfer deadline), run in its
// own goroutine so that a blocked source is abandoned to Close and the reap grace. It returns verdictExact for EOF,
// verdictLong for data, verdictNotReleased when the Read never returned, and verdictSourceError otherwise.
func probeEmpty(ctx context.Context, src *ownedSource) uploadVerdict {
	type result struct {
		n   int
		err error
	}
	for {
		ch := make(chan result, 1)
		go func() {
			var b [1]byte
			n, err := src.Read(b[:])
			ch <- result{n, err}
		}()
		select {
		case r := <-ch:
			switch {
			case r.n > 0:
				return verdictLong
			case errors.Is(r.err, io.EOF):
				return verdictExact
			case r.err != nil:
				return verdictSourceError
			}
			// 0, nil: read again (bounded by ctx, which closes the source).
		case <-ctx.Done():
			_ = src.Close()
			if !src.waitReleased(sourceReapGrace) {
				return verdictNotReleased
			}
			return verdictSourceError
		}
	}
}

// --- the exact upload reader with final-byte holdback ---------------------------------------------------------------

// uploadVerdict is the exact reader's account of the stream.
type uploadVerdict int

const (
	verdictPending     uploadVerdict = iota // the transport has not read to the end
	verdictExact                            // exactly Size bytes, then EOF
	verdictShort                            // EOF before Size bytes
	verdictLong                             // more than Size bytes
	verdictSourceError                      // the source failed (including cancellation or the deadline closing it)
	verdictNotReleased                      // the source did not return from Read within the reap grace
)

var (
	errUploadShort  = errors.New("access: upload shorter than its size")
	errUploadLong   = errors.New("access: upload longer than its size")
	errUploadSource = errors.New("access: upload source failed")
)

// exactReader releases bytes 1..Size-1 as they arrive and the final byte only after a lookahead Read returned EOF.
// Extra data or an early EOF makes Read fail instead, so the transport has received at most Size-1 body bytes against
// Content-Length = Size when the stream is not exact. It digests every released byte. It is used by one transport
// goroutine at a time, which alone touches hash. The digest leaves the reader only as a snapshot taken by that goroutine
// after the final byte's hash write and published with verdictExact under mu (correction C1): an asynchronous transport
// may return on cancellation while its goroutine is still inside Read, so the caller never reads hash itself.
type exactReader struct {
	src      io.Reader
	size     int64
	released int64
	hash     hash.Hash // transport goroutine only
	mu       sync.Mutex
	v        uploadVerdict // guarded by mu
	digest   [32]byte      // guarded by mu; set only together with verdictExact
	// afterRead is an inert test seam (nil in production): it runs on the transport goroutine after each source Read
	// returned, before the bytes are digested.
	afterRead func()
}

func newExactReader(src io.Reader, size int64) *exactReader {
	return &exactReader{src: src, size: size, hash: sha256.New()}
}

func (r *exactReader) set(v uploadVerdict) {
	r.mu.Lock()
	if r.v == verdictPending {
		r.v = v
	}
	r.mu.Unlock()
}

func (r *exactReader) verdict() uploadVerdict {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.v
}

func (r *exactReader) seam() {
	if r.afterRead != nil {
		r.afterRead()
	}
}

// result returns the verdict and, only for verdictExact, the digest snapshot published with it (zero otherwise).
func (r *exactReader) result() (uploadVerdict, [32]byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.v != verdictExact {
		return r.v, [32]byte{}
	}
	return r.v, r.digest
}

func (r *exactReader) fail(v uploadVerdict, err error) (int, error) {
	r.set(v)
	return 0, err
}

func (r *exactReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	switch r.verdict() {
	case verdictExact:
		return 0, io.EOF
	case verdictShort:
		return 0, errUploadShort
	case verdictLong:
		return 0, errUploadLong
	case verdictSourceError, verdictNotReleased:
		return 0, errUploadSource
	}
	remaining := r.size - r.released
	if remaining > 1 {
		want := int64(len(p))
		if want > remaining-1 {
			want = remaining - 1
		}
		n, err := r.src.Read(p[:want])
		r.seam()
		if n > 0 {
			r.hash.Write(p[:n])
			r.released += int64(n)
		}
		switch {
		case err == nil:
			return n, nil
		case errors.Is(err, io.EOF):
			r.set(verdictShort)
			if n > 0 {
				return n, nil
			}
			return 0, errUploadShort
		default:
			r.set(verdictSourceError)
			if n > 0 {
				return n, nil
			}
			return 0, errUploadSource
		}
	}
	// The final byte: read it together with a lookahead, and release it only after EOF.
	var buf [2]byte
	got := 0
	for got < 1 {
		n, err := r.src.Read(buf[:2])
		r.seam()
		got += n
		if got >= 2 {
			return r.fail(verdictLong, errUploadLong)
		}
		if err != nil {
			if got == 1 && errors.Is(err, io.EOF) {
				return r.final(p, buf[0])
			}
			if errors.Is(err, io.EOF) {
				return r.fail(verdictShort, errUploadShort)
			}
			return r.fail(verdictSourceError, errUploadSource)
		}
	}
	for {
		n, err := r.src.Read(buf[1:2])
		r.seam()
		if n > 0 {
			return r.fail(verdictLong, errUploadLong)
		}
		if errors.Is(err, io.EOF) {
			return r.final(p, buf[0])
		}
		if err != nil {
			return r.fail(verdictSourceError, errUploadSource)
		}
	}
}

// final releases the last byte: its hash write, then the digest snapshot, both on the transport goroutine, then the
// snapshot and verdictExact published together under mu (unless another verdict was already set).
func (r *exactReader) final(p []byte, b byte) (int, error) {
	p[0] = b
	r.hash.Write(p[:1])
	r.released++
	var d [32]byte
	copy(d[:], r.hash.Sum(nil))
	r.mu.Lock()
	if r.v == verdictPending {
		r.v, r.digest = verdictExact, d
	}
	r.mu.Unlock()
	return 1, nil
}

// --- the bounded, verifying download reader -------------------------------------------------------------------------

// verifyingReader is GetObject's returned body: bounded to the stored size, it verifies the stored SHA-256 at EOF and
// reports a mismatch (internal) or a short or failed transfer (unavailable) there, after any delivered bytes. Close
// cancels the transfer.
type verifyingReader struct {
	body      io.ReadCloser
	cancel    context.CancelFunc
	remaining int64
	want      [32]byte
	hash      hash.Hash
	err       error
	closeOnce sync.Once
}

func (v *verifyingReader) Read(p []byte) (int, error) {
	if v.err != nil {
		return 0, v.err
	}
	if v.remaining == 0 {
		var got [32]byte
		copy(got[:], v.hash.Sum(nil))
		if got != v.want {
			v.err = &Error{Code: CodeInternal}
			return 0, v.err
		}
		v.err = io.EOF
		return 0, io.EOF
	}
	if int64(len(p)) > v.remaining {
		p = p[:v.remaining]
	}
	n, err := v.body.Read(p)
	if n > 0 {
		v.hash.Write(p[:n])
		v.remaining -= int64(n)
	}
	if err != nil && (v.remaining > 0 || !errors.Is(err, io.EOF)) {
		v.err = &Error{Code: CodeUnavailable}
		if n > 0 {
			return n, nil
		}
		return 0, v.err
	}
	return n, nil
}

func (v *verifyingReader) Close() error {
	v.closeOnce.Do(func() {
		_ = v.body.Close()
		v.cancel()
	})
	return nil
}

// --- the signed URL --------------------------------------------------------------------------------------------------

// SignedURL is an issued presigned GET. The URL and its query are a bearer secret: URL returns it; String, GoString and
// LogValue redact it, so it is never printed or logged by accident. It is never stored or logged by core.
type SignedURL struct {
	url       string
	ExpiresAt time.Time
	TTL       time.Duration
}

const redactedURL = "[redacted signed URL]"

// URL returns the presigned URL.
func (u SignedURL) URL() string { return u.url }

// String redacts the URL.
func (u SignedURL) String() string { return redactedURL }

// GoString redacts the URL.
func (u SignedURL) GoString() string { return redactedURL }

// LogValue redacts the URL.
func (u SignedURL) LogValue() slog.Value { return slog.StringValue(redactedURL) }
