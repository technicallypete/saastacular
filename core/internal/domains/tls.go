package domains

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TLSAdapter is the generation-bound TLS readiness boundary (Plan r4 §1.5). Every call names one DomainRef and has a
// bounded deadline; a timeout or ambiguous reply is the explicit unknown state. There is no name-only operation.
type TLSAdapter interface {
	Request(ctx context.Context, ref DomainRef) TLSResult
	Status(ctx context.Context, ref DomainRef) TLSResult
	Remove(ctx context.Context, ref DomainRef) TLSResult
}

// EdgeProbe is the provider-neutral E1 adapter (ADR-007: Caddy on-demand TLS with the core-owned tls-ask answer).
// Request is a no-op (requested): the edge asks tls-ask before issuing. Status dials the edge with SNI = hostname and
// verifies the served chain, its SAN and expiry. Remove is withdrawn immediately: withdrawal is the issuance predicate
// answering 403 from the removal commit on (an already issued certificate stays cached until its expiry while routing
// already refuses).
type EdgeProbe struct {
	Addr    string         // DOMAIN_TLS_EDGE_ADDR
	Timeout time.Duration  // DOMAIN_TLS_TIMEOUT
	Roots   *x509.CertPool // nil: the system roots
	now     func() time.Time
}

// Request implements TLSAdapter.
func (e EdgeProbe) Request(_ context.Context, ref DomainRef) TLSResult {
	return TLSResult{Ref: ref, State: TLSRequested}
}

// Remove implements TLSAdapter.
func (e EdgeProbe) Remove(_ context.Context, ref DomainRef) TLSResult {
	return TLSResult{Ref: ref, State: TLSWithdrawn}
}

// Status implements TLSAdapter.
func (e EdgeProbe) Status(ctx context.Context, ref DomainRef) TLSResult {
	res := TLSResult{Ref: ref, State: TLSUnknown}
	ctx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()
	d := tls.Dialer{NetDialer: &net.Dialer{}, Config: &tls.Config{ServerName: ref.Hostname, RootCAs: e.Roots,
		MinVersion: tls.VersionTLS12, Time: e.now}}
	conn, err := d.DialContext(ctx, "tcp", e.Addr)
	if err == nil {
		defer conn.Close()
		state := conn.(*tls.Conn).ConnectionState()
		if len(state.PeerCertificates) == 0 {
			return res
		}
		res.State, res.NotAfter = TLSReady, state.PeerCertificates[0].NotAfter
		return res
	}
	var hostErr x509.HostnameError
	var authErr x509.UnknownAuthorityError
	var invalid x509.CertificateInvalidError
	var verifyErr *tls.CertificateVerificationError
	var alert tls.AlertError
	var opErr *net.OpError
	switch {
	case errors.As(err, &invalid) && invalid.Reason == x509.Expired:
		res.State = TLSExpired
		if invalid.Cert != nil {
			res.NotAfter = invalid.Cert.NotAfter
		}
	case errors.As(err, &hostErr), errors.As(err, &authErr), errors.As(err, &invalid):
		res.State = TLSError
	case errors.As(err, &verifyErr):
		res.State = TLSError
	case errors.As(err, &alert):
		res.State = TLSPending // the edge refused the handshake for this name: no certificate yet
	case errors.As(err, &opErr) && opErr.Op == "remote error":
		res.State = TLSPending
	}
	return res
}

// FixtureAdapter scripts outcomes per hostname and operation (DOMAIN_TLS_FIXTURE_FILE; APP_ENV=test only): one
// "hostname<TAB>generation-or-*<TAB>operation<TAB>state[<TAB>not-after RFC 3339]" per line; repeated lines for the same
// key form a sequence consumed in order, the last repeating. An unscripted Status is pending, Request requested and
// Remove withdrawn. It is also the programmable adapter of the integration tests (Script).
type FixtureAdapter struct {
	mu      sync.Mutex
	scripts map[string][]TLSResult
	calls   []string
	hook    func(op string, ref DomainRef)
}

// NewFixtureAdapter returns an empty fixture adapter.
func NewFixtureAdapter() *FixtureAdapter { return &FixtureAdapter{scripts: map[string][]TLSResult{}} }

// ParseFixtureAdapter parses a fixture table.
func ParseFixtureAdapter(b []byte) (*FixtureAdapter, error) {
	f := NewFixtureAdapter()
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	for sc.Scan() {
		line := strings.TrimSuffix(sc.Text(), "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.Split(line, "\t")
		if len(p) < 4 {
			return nil, errors.New("domains: malformed TLS fixture line")
		}
		gen := int64(0)
		if p[1] != "*" {
			n, err := strconv.ParseInt(p[1], 10, 64)
			if err != nil || n < 1 {
				return nil, errors.New("domains: malformed TLS fixture generation")
			}
			gen = n
		}
		r := TLSResult{State: TLSState(p[3])}
		if len(p) > 4 && p[4] != "" {
			t, err := time.Parse(time.RFC3339, p[4])
			if err != nil {
				return nil, errors.New("domains: malformed TLS fixture time")
			}
			r.NotAfter = t
		}
		f.Script(p[0], gen, p[2], r)
	}
	return f, sc.Err()
}

func fixtureKey(host string, gen int64, op string) string {
	return host + "\x00" + strconv.FormatInt(gen, 10) + "\x00" + op
}

// Script appends one scripted result for host, generation (0 = any) and operation (request, status or remove).
func (f *FixtureAdapter) Script(host string, gen int64, op string, r TLSResult) {
	f.mu.Lock()
	defer f.mu.Unlock()
	k := fixtureKey(host, gen, op)
	f.scripts[k] = append(f.scripts[k], r)
}

// OnCall installs a hook run at the start of every call (tests: a late completion or a removal mid-call).
func (f *FixtureAdapter) OnCall(h func(op string, ref DomainRef)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hook = h
}

// Calls returns the recorded calls as "op hostname generation".
func (f *FixtureAdapter) Calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.calls...)
}

func (f *FixtureAdapter) next(op string, ref DomainRef, def TLSState) TLSResult {
	f.mu.Lock()
	f.calls = append(f.calls, op+" "+ref.Hostname+" "+strconv.FormatInt(ref.Generation, 10))
	hook := f.hook
	var r TLSResult
	found := false
	for _, k := range []string{fixtureKey(ref.Hostname, ref.Generation, op), fixtureKey(ref.Hostname, 0, op)} {
		if s := f.scripts[k]; len(s) > 0 {
			r, found = s[0], true
			if len(s) > 1 {
				f.scripts[k] = s[1:]
			}
			break
		}
	}
	f.mu.Unlock()
	if hook != nil {
		hook(op, ref)
	}
	if !found {
		r = TLSResult{State: def}
	}
	r.Ref = ref
	return r
}

// Request implements TLSAdapter.
func (f *FixtureAdapter) Request(_ context.Context, ref DomainRef) TLSResult {
	return f.next("request", ref, TLSRequested)
}

// Status implements TLSAdapter.
func (f *FixtureAdapter) Status(_ context.Context, ref DomainRef) TLSResult {
	return f.next("status", ref, TLSPending)
}

// Remove implements TLSAdapter.
func (f *FixtureAdapter) Remove(_ context.Context, ref DomainRef) TLSResult {
	return f.next("remove", ref, TLSWithdrawn)
}
