package domains

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"
	"time"
)

// Checker looks up the two DNS facts of an ownership proof. Implementations return ErrNXDomain when the name or record
// is confirmed absent and ErrDNSTransient for any other failure (timeout, SERVFAIL, refused, cancellation).
type Checker interface {
	LookupTXT(ctx context.Context, name string) ([]string, error)
	LookupCNAME(ctx context.Context, name string) (string, error)
}

// DNS lookup errors.
var (
	ErrNXDomain     = errors.New("domains: name not found")
	ErrDNSTransient = errors.New("domains: transient DNS failure")
)

// Proof bounds (Plan r4 §1.5).
const (
	maxCNAMEHops  = 3
	maxTXTAnswers = 16
)

// ProofConfig is the configuration of an ownership proof.
type ProofConfig struct {
	ChallengeLabel string        // CUSTOM_DOMAIN_CHALLENGE_LABEL
	CNAMETarget    string        // CUSTOM_DOMAIN_CNAME_TARGET
	BaseDomain     string        // PLATFORM_BASE_DOMAIN (refused before any lookup)
	Timeout        time.Duration // per lookup
}

// refusedSuffixes are special-use and internal names that are never looked up (Plan r4 §1.5).
var refusedSuffixes = []string{"localhost", "local", "internal", "test", "example", "invalid", "onion", "home.arpa", "arpa"}

// Refused reports whether a hostname is refused before any lookup: special-use names, the platform domain and names
// under it, and the CNAME target itself.
func Refused(host string, cfg ProofConfig) bool {
	if host == cfg.CNAMETarget || host == cfg.BaseDomain || strings.HasSuffix(host, "."+cfg.BaseDomain) ||
		net.ParseIP(host) != nil {
		return true
	}
	for _, s := range refusedSuffixes {
		if host == s || strings.HasSuffix(host, "."+s) {
			return true
		}
	}
	return false
}

// Proof checks a custom hostname's ownership proof: (a) a TXT record at <label>.<host> equal to the row's challenge
// value and (b) the host's CNAME chain (at most three hops) ending exactly at the CNAME target. A shared target alone
// is never sufficient. Confirmed absence is nxdomain, a wrong value or target is mismatch, anything else transient.
// Refused names are a mismatch without any lookup.
func Proof(ctx context.Context, c Checker, cfg ProofConfig, host, value string) Outcome {
	if Refused(host, cfg) || value == "" {
		return OutcomeMismatch
	}
	lookup := func(f func(context.Context) error) error {
		lctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
		return f(lctx)
	}
	var txt []string
	if err := lookup(func(lctx context.Context) (err error) {
		txt, err = c.LookupTXT(lctx, cfg.ChallengeLabel+"."+host)
		return err
	}); err != nil {
		return outcomeOf(err)
	}
	match := false
	for i, v := range txt {
		if i >= maxTXTAnswers {
			break
		}
		if v == value {
			match = true
		}
	}
	if !match {
		return OutcomeMismatch
	}
	name := host
	for hop := 0; ; hop++ {
		var next string
		if err := lookup(func(lctx context.Context) (err error) {
			next, err = c.LookupCNAME(lctx, name)
			return err
		}); err != nil {
			if hop > 0 && errors.Is(err, ErrNXDomain) {
				return OutcomeMismatch // the chain ended at a name that is not the target
			}
			return outcomeOf(err)
		}
		next = strings.TrimSuffix(strings.ToLower(next), ".")
		if next == cfg.CNAMETarget {
			return OutcomeOK
		}
		if next == name || next == "" || hop+1 >= maxCNAMEHops {
			return OutcomeMismatch
		}
		name = next
	}
}

func outcomeOf(err error) Outcome {
	if errors.Is(err, ErrNXDomain) {
		return OutcomeNXDomain
	}
	return OutcomeTransient
}

// SystemChecker is the production checker on the system resolver configuration (no caller-supplied resolver).
type SystemChecker struct {
	r     *net.Resolver
	cname cnameClient
}

// NewSystemChecker returns the production checker over the default dialer: TXT through Go's native resolver on the
// system configuration, CNAME edges through a typed CNAME query to the system configuration's nameservers (dnswire.go).
func NewSystemChecker() SystemChecker { return newSystemChecker(nil, nil) }

// newSystemChecker is NewSystemChecker with a transport dialer and a nameserver configuration. Production passes nil
// for both (the system dialer and /etc/resolv.conf); the package's tests inject disposable in-process DNS transports
// to pin the per-edge semantics of LookupCNAME against actual wire answers. It is unexported: no caller selects a
// resolver or a server.
func newSystemChecker(dial func(ctx context.Context, network, address string) (net.Conn, error), conf func() resolvConf) SystemChecker {
	c := cnameClient{dial: dial, conf: conf, bound: defaultLookupBound}
	if c.dial == nil {
		c.dial = (&net.Dialer{}).DialContext
	}
	if c.conf == nil {
		c.conf = func() resolvConf { return readResolvConf(resolvConfPath) }
	}
	return SystemChecker{r: &net.Resolver{PreferGo: true, Dial: dial}, cname: c}
}

// LookupTXT implements Checker.
func (s SystemChecker) LookupTXT(ctx context.Context, name string) ([]string, error) {
	v, err := s.r.LookupTXT(ctx, name)
	return v, classify(err)
}

// LookupCNAME implements Checker: the next name of the chain, one actual edge per call, so Proof's loop bounds actual
// edges unconditionally. It sends one typed CNAME query for the exact absolute name and returns the target of the
// unique CNAME record owned by that name, whatever the order of the answer section; a name without a CNAME returns
// itself (the chain's end, which Proof treats as a mismatch unless the target was reached), a nonexistent name is
// ErrNXDomain, and an ambiguous or malformed answer is ErrDNSTransient. TestDNSSystemChecker and the TestDNSWire tests
// pin this on wire answers (three edges ok, four edges mismatch, any record order).
func (s SystemChecker) LookupCNAME(ctx context.Context, name string) (string, error) {
	return s.cname.lookup(ctx, name)
}

func classify(err error) error {
	if err == nil {
		return nil
	}
	var de *net.DNSError
	if errors.As(err, &de) && de.IsNotFound {
		return ErrNXDomain
	}
	return ErrDNSTransient
}

// FixtureChecker answers from a table (CUSTOM_DOMAIN_DNS_FIXTURE_FILE; APP_ENV=test only): one
// "name<TAB>type<TAB>value" per line with type TXT, CNAME, NXDOMAIN or SERVFAIL; '#' comments allowed. A name without
// an entry of the asked type is NXDOMAIN.
type FixtureChecker struct {
	txt   map[string][]string
	cname map[string]string
	fail  map[string]error
}

// ParseFixtureChecker parses a fixture table.
func ParseFixtureChecker(b []byte) (FixtureChecker, error) {
	f := FixtureChecker{txt: map[string][]string{}, cname: map[string]string{}, fail: map[string]error{}}
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	for sc.Scan() {
		line := strings.TrimSuffix(sc.Text(), "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.Split(line, "\t")
		if len(p) < 2 || p[0] == "" {
			return FixtureChecker{}, errors.New("domains: malformed DNS fixture line")
		}
		name := strings.ToLower(strings.TrimSuffix(p[0], "."))
		val := ""
		if len(p) > 2 {
			val = p[2]
		}
		switch p[1] {
		case "TXT":
			f.txt[name] = append(f.txt[name], val)
		case "CNAME":
			f.cname[name] = val
		case "NXDOMAIN":
			f.fail[name] = ErrNXDomain
		case "SERVFAIL":
			f.fail[name] = ErrDNSTransient
		default:
			return FixtureChecker{}, errors.New("domains: unknown DNS fixture record type")
		}
	}
	return f, sc.Err()
}

func (f FixtureChecker) answer(ctx context.Context, name string) (string, error) {
	if ctx.Err() != nil {
		return "", ErrDNSTransient
	}
	name = strings.ToLower(strings.TrimSuffix(name, "."))
	if err, ok := f.fail[name]; ok {
		return name, err
	}
	return name, nil
}

// LookupTXT implements Checker.
func (f FixtureChecker) LookupTXT(ctx context.Context, name string) ([]string, error) {
	n, err := f.answer(ctx, name)
	if err != nil {
		return nil, err
	}
	if v, ok := f.txt[n]; ok {
		return v, nil
	}
	return nil, ErrNXDomain
}

// LookupCNAME implements Checker: the next name of the chain, or the name itself at the chain's end.
func (f FixtureChecker) LookupCNAME(ctx context.Context, name string) (string, error) {
	n, err := f.answer(ctx, name)
	if err != nil {
		return "", err
	}
	if v, ok := f.cname[n]; ok {
		return v, nil
	}
	if _, ok := f.txt[n]; ok {
		return n, nil
	}
	return "", ErrNXDomain
}
