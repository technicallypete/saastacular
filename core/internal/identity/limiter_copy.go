package identity

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// G062 (Plan r5 §4.5; D-12): a byte-pinned copy of core/internal/enrollment/api's limiter (TestLimiterCopyIsExact), used
// by the identity HTTP adapter for the root signup login start and the root-origin callback only. Every bucket is per
// process and in memory; a multi-process deployment multiplies capacity by the process count (IDENTITY.md).

// maxLimiters bounds the keys of one limiter (the enrollment/api value).
const maxLimiters = 10000

// limiter is a bounded set of token buckets keyed by a digest. The map is never cleared and a bucket that is not fully
// refilled is never dropped, so filling or churning the map with unrelated keys cannot restore a depleted bucket. When
// the map holds maxLimiters keys a new key triggers a reclaim scan that deletes only fully refilled buckets (dropping
// one is indistinguishable from recreating it); if the scan frees nothing the new key is refused (rate_limited) and no
// existing bucket is touched. Every scan, successful or not, defers the next one by the full-refill time burst/rate, so
// new keys at capacity inside that window are refused without scanning (at most one O(maxLimiters) scan per window).
// With no refill (rate <= 0) a used bucket can never become fully refilled: there is no scan and no window, and a new
// key at capacity is refused.
type limiter struct {
	mu    sync.Mutex
	m     map[string]*rate.Limiter
	r     rate.Limit
	burst int
	now   func() time.Time // the clock (time.Now; a test injects a fake), read once per allow under mu
	full  time.Duration    // burst/rate, the full-refill time and scan window; 0 when rate <= 0 (no reclaim)
	next  time.Time        // the earliest instant the next reclaim scan may run
}

func newLimiter(perSecond, burst int) *limiter {
	l := &limiter{m: map[string]*rate.Limiter{}, r: rate.Limit(perSecond), burst: burst, now: time.Now}
	if perSecond > 0 {
		l.full = time.Duration(float64(burst) / float64(perSecond) * float64(time.Second))
	}
	return l
}

// allow takes one token from key's bucket. A new key at capacity is refused unless reclaim frees room.
func (l *limiter) allow(key string) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b, ok := l.m[key]
	if !ok {
		if len(l.m) >= maxLimiters && !l.reclaim(now) {
			return false
		}
		b = rate.NewLimiter(l.r, l.burst)
		l.m[key] = b
	}
	return b.AllowN(now, 1)
}

// reclaim (l.mu held) deletes the fully refilled buckets at now and reports whether any was freed. It scans only at or
// after l.next and then moves l.next to now+full whatever the outcome; with no refill it never scans.
func (l *limiter) reclaim(now time.Time) bool {
	if l.full <= 0 || now.Before(l.next) {
		return false
	}
	l.next = now.Add(l.full)
	burst, freed := float64(l.burst), false
	for k, b := range l.m {
		if b.TokensAt(now) >= burst {
			delete(l.m, k)
			freed = true
		}
	}
	return freed
}
