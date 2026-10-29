package workos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

// JWKS bounds (Plan r5 §3.6).
const (
	jwksRefreshEvery   = time.Hour
	jwksUnknownKIDGap  = 30 * time.Second
	jwksFetchDeadline  = 5 * time.Second
	jwksMaxBody        = 256 << 10
	authenticateMaxBdy = 64 << 10
)

var (
	// errKeysUnavailable: no trusted key set could be obtained (outage). Never reported as a signature failure.
	errKeysUnavailable = errors.New("workos: signing keys unavailable")
	// errUnknownKey: the token names a key that a trusted, current key set does not contain.
	errUnknownKey = errors.New("workos: unknown signing key")
)

// keySource holds the client-specific JWKS fetched only from the configured base URL. It refreshes every hour (lazily,
// when used), refreshes for an unknown kid at most once per 30 s (non-blocking, de-duplicated) and keeps serving the
// last trusted set when a refresh fails. Unknown keys fail until a trusted refresh contains them.
type keySource struct {
	url     string
	client  *http.Client
	now     func() time.Time
	limiter *rate.Limiter
	timeout time.Duration

	mu       sync.Mutex // serializes fetches
	current  atomic.Pointer[keyfunc.Keyfunc]
	loadedAt atomic.Int64 // unix nanoseconds of the last trusted fetch
}

func newKeySource(url string, client *http.Client, now func() time.Time) *keySource {
	return &keySource{url: url, client: client, now: now, limiter: rate.NewLimiter(rate.Every(jwksUnknownKIDGap), 1),
		timeout: jwksFetchDeadline}
}

// fetch downloads and parses the key set, replacing the current one on success.
func (k *keySource) fetch(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, k.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := k.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("workos: jwks status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body) // bounded by the client's transport
	if err != nil {
		return err
	}
	var set jwkset.JWKSMarshal
	if err := json.Unmarshal(body, &set); err != nil {
		return err
	}
	if len(set.Keys) == 0 {
		return errors.New("workos: empty key set")
	}
	store, err := set.ToStorage()
	if err != nil {
		return err
	}
	kf, err := keyfunc.New(keyfunc.Options{Ctx: context.Background(), Storage: store})
	if err != nil {
		return err
	}
	k.current.Store(&kf)
	k.loadedAt.Store(k.now().UnixNano())
	return nil
}

// refresh fetches unless another caller fetched within minGap; it reports whether a trusted set is now current.
func (k *keySource) refresh(ctx context.Context, minGap time.Duration) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if last := k.loadedAt.Load(); last != 0 && k.now().Sub(time.Unix(0, last)) < minGap {
		return nil
	}
	return k.fetch(ctx)
}

// keyfunc returns the jwt.Keyfunc for one validation. Algorithm allow-listing happens in the parser before this is
// called; the JWK's own alg (if any) must match the token's (keyfunc enforces it).
func (k *keySource) keyfunc(ctx context.Context) jwt.Keyfunc {
	return func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok || kid == "" || len(kid) > 256 {
			return nil, errUnknownKey
		}
		// Lazy periodic refresh; a failed refresh keeps the last trusted set.
		if last := k.loadedAt.Load(); last == 0 || k.now().Sub(time.Unix(0, last)) >= jwksRefreshEvery {
			_ = k.refresh(ctx, jwksRefreshEvery)
		}
		kf := k.current.Load()
		if kf == nil {
			return nil, errKeysUnavailable
		}
		key, err := (*kf).KeyfuncCtx(ctx)(t)
		if err == nil {
			return key, nil
		}
		if !errors.Is(err, jwkset.ErrKeyNotFound) {
			return nil, errUnknownKey
		}
		// Unknown kid: at most one de-duplicated, non-blocking refresh per 30 s.
		if !k.limiter.Allow() {
			return nil, errUnknownKey
		}
		if err := k.refresh(ctx, 0); err != nil {
			return nil, errKeysUnavailable
		}
		key, err = (*k.current.Load()).KeyfuncCtx(ctx)(t)
		if err != nil {
			return nil, errUnknownKey
		}
		return key, nil
	}
}

// boundedClient is an HTTP client that never follows redirects, has a total deadline and limits response bodies.
func boundedClient(next http.RoundTripper, timeout time.Duration, maxBody int64) *http.Client {
	if next == nil {
		next = http.DefaultTransport
	}
	return &http.Client{
		Transport: limitTransport{next: next, limit: maxBody},
		Timeout:   timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// errBodyTooLarge is returned when a provider response exceeds its bound.
var errBodyTooLarge = errors.New("workos: response body too large")

type limitTransport struct {
	next  http.RoundTripper
	limit int64
}

func (t limitTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := t.next.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	if resp.ContentLength > t.limit {
		resp.Body.Close()
		return nil, errBodyTooLarge
	}
	resp.Body = &limitedBody{rc: resp.Body, remaining: t.limit}
	return resp, nil
}

type limitedBody struct {
	rc        io.ReadCloser
	remaining int64
}

func (b *limitedBody) Read(p []byte) (int, error) {
	if b.remaining <= 0 {
		var one [1]byte
		if n, _ := b.rc.Read(one[:]); n > 0 {
			return 0, errBodyTooLarge
		}
		return 0, io.EOF
	}
	if int64(len(p)) > b.remaining {
		p = p[:b.remaining]
	}
	n, err := b.rc.Read(p)
	b.remaining -= int64(n)
	return n, err
}

func (b *limitedBody) Close() error { return b.rc.Close() }
