package domains

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

const (
	testValue  = "abcdefghijklmnopqrstuvwxyz234567abcdefghijklmnopqrst"
	testOther  = "zzzzefghijklmnopqrstuvwxyz234567abcdefghijklmnopqrst"
	testTarget = "custom.example.net"
)

func TestDNSChecker(t *testing.T) {
	fixture := strings.Join([]string{
		"# proof fixture",
		"_platform-verify.good.example.com\tTXT\tnoise",
		"_platform-verify.good.example.com\tTXT\t" + testValue,
		"good.example.com\tCNAME\t" + testTarget + ".",
		"_platform-verify.shared.example.com\tTXT\t" + testOther, // shared target, foreign value
		"shared.example.com\tCNAME\t" + testTarget,
		"target-only.example.com\tCNAME\t" + testTarget, // no TXT at all
		"_platform-verify.wrong.example.com\tTXT\t" + testValue,
		"wrong.example.com\tCNAME\tedge.elsewhere.example.org",
		"_platform-verify.chain.example.com\tTXT\t" + testValue,
		"chain.example.com\tCNAME\thop1.example.org",
		"hop1.example.org\tCNAME\thop2.example.org",
		"hop2.example.org\tCNAME\t" + testTarget,
		"_platform-verify.long.example.com\tTXT\t" + testValue,
		"long.example.com\tCNAME\tl1.example.org",
		"l1.example.org\tCNAME\tl2.example.org",
		"l2.example.org\tCNAME\tl3.example.org",
		"l3.example.org\tCNAME\t" + testTarget,
		"_platform-verify.flaky.example.com\tSERVFAIL",
		"_platform-verify.gone.example.com\tNXDOMAIN",
		"_platform-verify.cname-flaky.example.com\tTXT\t" + testValue,
		"cname-flaky.example.com\tSERVFAIL",
	}, "\n")
	c, err := ParseFixtureChecker([]byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	cfg := ProofConfig{ChallengeLabel: "_platform-verify", CNAMETarget: testTarget, BaseDomain: "example.net", Timeout: time.Second}
	ctx := context.Background()
	for _, k := range []struct {
		host, value string
		want        Outcome
	}{
		{"good.example.com", testValue, OutcomeOK},
		{"good.example.com", testOther, OutcomeMismatch},        // a foreign, stale (rotated) or replayed value
		{"shared.example.com", testValue, OutcomeMismatch},      // the shared target alone is never sufficient
		{"target-only.example.com", testValue, OutcomeNXDomain}, // CNAME without the TXT proof
		{"wrong.example.com", testValue, OutcomeMismatch},
		{"chain.example.com", testValue, OutcomeOK},
		{"long.example.com", testValue, OutcomeMismatch}, // more than three hops
		{"flaky.example.com", testValue, OutcomeTransient},
		{"gone.example.com", testValue, OutcomeNXDomain},
		{"cname-flaky.example.com", testValue, OutcomeTransient},
		{"unknown.example.com", testValue, OutcomeNXDomain},
		{"good.example.com", "", OutcomeMismatch},
	} {
		if got := Proof(ctx, c, cfg, k.host, k.value); got != k.want {
			t.Errorf("Proof(%s, %.4s) = %s; want %s", k.host, k.value, got, k.want)
		}
	}

	// Refused names are never looked up.
	rec := &recordingChecker{}
	for _, h := range []string{"localhost", "a.localhost", "printer.local", "svc.internal", "portal.acme.test", "x.example",
		"x.invalid", "x.onion", "router.home.arpa", "1.0.0.10.in-addr.arpa", "example.net", "org-a.example.net",
		testTarget, "10.0.0.1"} {
		if !Refused(h, cfg) {
			t.Errorf("%s must be refused", h)
		}
		if got := Proof(ctx, rec, cfg, h, testValue); got != OutcomeMismatch {
			t.Errorf("Proof(%s) = %s", h, got)
		}
	}
	if rec.n != 0 {
		t.Errorf("refused names were looked up %d times", rec.n)
	}
	if Refused("portal.example.com", cfg) {
		t.Error("an ordinary name must not be refused")
	}

	// Deadline and cancellation are transient.
	slow := slowChecker{}
	if got := Proof(ctx, slow, ProofConfig{ChallengeLabel: "_v", CNAMETarget: testTarget, BaseDomain: "example.net", Timeout: 20 * time.Millisecond}, "slow.example.com", testValue); got != OutcomeTransient {
		t.Errorf("deadline: %s", got)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if got := Proof(canceled, c, cfg, "good.example.com", testValue); got != OutcomeTransient {
		t.Errorf("canceled: %s", got)
	}

	for _, bad := range []string{"name-only", "x\tAAAA\t1.2.3.4"} {
		if _, err := ParseFixtureChecker([]byte(bad)); err == nil {
			t.Errorf("fixture %q accepted", bad)
		}
	}
}

type recordingChecker struct{ n int }

func (r *recordingChecker) LookupTXT(context.Context, string) ([]string, error) {
	r.n++
	return nil, ErrNXDomain
}
func (r *recordingChecker) LookupCNAME(context.Context, string) (string, error) {
	r.n++
	return "", ErrNXDomain
}

type slowChecker struct{}

func (slowChecker) LookupTXT(ctx context.Context, _ string) ([]string, error) {
	<-ctx.Done()
	return nil, ErrDNSTransient
}
func (slowChecker) LookupCNAME(ctx context.Context, _ string) (string, error) {
	<-ctx.Done()
	return "", ErrDNSTransient
}

// TestDNSSystemChecker drives the production adapter (behind NewSystemChecker) over a disposable in-process DNS
// transport answering like a recursive resolver: chains in order, no network. It pins that LookupCNAME yields one
// actual edge per call, so Proof's three-hop bound holds on wire answers: three edges ok, four mismatch. This stream
// transport exercises the length-framed path; dnswire_test.go covers real loopback UDP and TCP and adverse answers.
func TestDNSSystemChecker(t *testing.T) {
	z := wireZone{
		txt: map[string]string{},
		cname: map[string]string{
			"one.example.com":   testTarget,
			"chain.example.com": "hop1.example.org", "hop1.example.org": "hop2.example.org", "hop2.example.org": testTarget,
			"long.example.com": "l1.example.org", "l1.example.org": "l2.example.org", "l2.example.org": "l3.example.org",
			"l3.example.org":    testTarget,
			"wrong.example.com": "edge.elsewhere.example.org",
		},
		a: map[string]bool{testTarget: true, "edge.elsewhere.example.org": true, "plain.example.com": true},
	}
	for _, h := range []string{"one.example.com", "chain.example.com", "long.example.com", "wrong.example.com", "plain.example.com"} {
		z.txt["_platform-verify."+h] = testValue
	}
	sys := newSystemChecker(z.dial, pipeConf)
	ctx := context.Background()
	// One edge per call, on the production adapter itself.
	for name, want := range map[string]string{"long.example.com": "l1.example.org", "l1.example.org": "l2.example.org",
		"chain.example.com": "hop1.example.org", "plain.example.com": "plain.example.com"} {
		got, err := sys.LookupCNAME(ctx, name)
		if err != nil || strings.TrimSuffix(got, ".") != want {
			t.Errorf("LookupCNAME(%s) = %q, %v; want the edge %s", name, got, err, want)
		}
	}
	if v, err := sys.LookupTXT(ctx, "_platform-verify.chain.example.com"); err != nil || len(v) != 1 || v[0] != testValue {
		t.Errorf("LookupTXT = %v, %v", v, err)
	}
	if _, err := sys.LookupCNAME(ctx, "unknown.example.com"); err != ErrNXDomain {
		t.Errorf("unknown name: %v, want ErrNXDomain", err)
	}
	cfg := ProofConfig{ChallengeLabel: "_platform-verify", CNAMETarget: testTarget, BaseDomain: "example.net", Timeout: 2 * time.Second}
	for _, k := range []struct {
		host string
		want Outcome
	}{
		{"one.example.com", OutcomeOK},
		{"chain.example.com", OutcomeOK},      // three edges
		{"long.example.com", OutcomeMismatch}, // four edges: over the bound although the chain reaches the target
		{"wrong.example.com", OutcomeMismatch},
		{"plain.example.com", OutcomeMismatch}, // no CNAME at all
		{"unknown.example.com", OutcomeNXDomain},
	} {
		if got := Proof(ctx, sys, cfg, k.host, testValue); got != k.want {
			t.Errorf("system Proof(%s) = %s; want %s", k.host, got, k.want)
		}
	}
	// The per-lookup deadline holds on the system adapter: a server that never answers is transient.
	mute := newSystemChecker(func(context.Context, string, string) (net.Conn, error) {
		c, s := net.Pipe()
		go func() { _, _ = io.Copy(io.Discard, s) }()
		return c, nil
	}, pipeConf)
	start := time.Now()
	if got := Proof(ctx, mute, ProofConfig{ChallengeLabel: "_v", CNAMETarget: testTarget, BaseDomain: "example.net",
		Timeout: 100 * time.Millisecond}, "slow.example.com", testValue); got != OutcomeTransient || time.Since(start) > 2*time.Second {
		t.Errorf("mute server: %s after %s", got, time.Since(start))
	}
}

// pipeConf names one documentation-range server; the injected stream dialers ignore the address.
func pipeConf() resolvConf { return resolvConf{servers: []string{"192.0.2.53:53"}} }

// wireZone answers DNS queries on a stream transport (2-byte length framing) the way a recursive resolver does: an
// A or AAAA answer is the name's CNAME chain in order followed by the end's address, a CNAME answer is the name's own
// edge, a TXT answer the name's value; an unknown name is NXDOMAIN, a known name without the type NODATA.
type wireZone struct {
	txt   map[string]string
	cname map[string]string
	a     map[string]bool
}

func (z wireZone) dial(context.Context, string, string) (net.Conn, error) {
	c, s := net.Pipe()
	go z.serve(s)
	return c, nil
}

func (z wireZone) serve(s net.Conn) {
	defer s.Close()
	for {
		var l [2]byte
		if _, err := io.ReadFull(s, l[:]); err != nil {
			return
		}
		q := make([]byte, binary.BigEndian.Uint16(l[:]))
		if _, err := io.ReadFull(s, q); err != nil {
			return
		}
		r := z.answer(q)
		if _, err := s.Write(append(binary.BigEndian.AppendUint16(nil, uint16(len(r))), r...)); err != nil {
			return
		}
	}
}

func (z wireZone) answer(q []byte) []byte {
	i, name := 12, ""
	for i < len(q) && q[i] != 0 {
		n := int(q[i])
		name += string(q[i+1:i+1+n]) + "."
		i += 1 + n
	}
	name = strings.ToLower(strings.TrimSuffix(name, "."))
	qtype := binary.BigEndian.Uint16(q[i+1 : i+3])
	question := q[12 : i+5]
	_, isTXT := z.txt[name]
	_, isCNAME := z.cname[name]
	var rrs [][]byte
	switch qtype {
	case 5: // CNAME
		if t, ok := z.cname[name]; ok {
			rrs = append(rrs, rr(name, 5, wireName(t)))
		}
	case 16: // TXT
		if v, ok := z.txt[name]; ok {
			rrs = append(rrs, rr(name, 16, append([]byte{byte(len(v))}, v...)))
		}
	case 1, 28: // A, AAAA
		n := name
		for hops := 0; hops < 16; hops++ {
			t, ok := z.cname[n]
			if !ok {
				break
			}
			rrs = append(rrs, rr(n, 5, wireName(t)))
			n = t
		}
		if qtype == 1 && z.a[n] {
			rrs = append(rrs, rr(n, 1, []byte{192, 0, 2, 1}))
		}
	}
	flags := uint16(0x8180) // response, recursion desired and available
	if !isTXT && !isCNAME && !z.a[name] {
		flags |= 3 // NXDOMAIN
	}
	m := append([]byte{q[0], q[1]}, binary.BigEndian.AppendUint16(nil, flags)...)
	m = binary.BigEndian.AppendUint16(m, 1)
	m = binary.BigEndian.AppendUint16(m, uint16(len(rrs)))
	m = append(m, 0, 0, 0, 0)
	m = append(m, question...)
	for _, r := range rrs {
		m = append(m, r...)
	}
	return m
}

func wireName(name string) []byte {
	var b []byte
	for _, l := range strings.Split(strings.TrimSuffix(name, "."), ".") {
		b = append(append(b, byte(len(l))), l...)
	}
	return append(b, 0)
}

func rr(name string, typ uint16, data []byte) []byte {
	b := wireName(name)
	b = binary.BigEndian.AppendUint16(b, typ)
	b = binary.BigEndian.AppendUint16(b, 1)  // class IN
	b = binary.BigEndian.AppendUint32(b, 60) // TTL
	b = binary.BigEndian.AppendUint16(b, uint16(len(data)))
	return append(b, data...)
}
