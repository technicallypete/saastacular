package domains

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// The TestDNSWire tests drive SystemChecker.LookupCNAME against scripted DNS servers on real loopback UDP and TCP
// sockets (no external network). Every case scripts the exact reply bytes, so the outcomes are deterministic.

const (
	flagsOK = 0x8180 // QR, RD, RA, NOERROR
	flagTC  = 0x0200
)

// wireQuery is a query as the fake server received it.
type wireQuery struct {
	transport string // "udp" or "tcp"
	id        uint16
	name      string // absolute, as sent
	question  []byte // the question section bytes, for echoing
	at        time.Time
}

// wireHandler returns the replies to one query: UDP datagrams, or raw bytes written to the TCP stream (use frame for
// a well-formed frame). No reply leaves the server mute (a TCP connection then stays open until the client closes it).
type wireHandler func(q wireQuery) [][]byte

// fakeDNS is a scripted DNS server listening on one loopback port over UDP and TCP.
type fakeDNS struct {
	addr   string
	handle wireHandler
	udp    *net.UDPConn
	tcp    *net.TCPListener
	wg     sync.WaitGroup
	mu     sync.Mutex
	got    []wireQuery
	conns  map[net.Conn]bool
	recv   chan wireQuery // every received query
	tcpEOF chan struct{}  // one per TCP connection the client closed
}

func newFakeDNS(t *testing.T, h wireHandler) *fakeDNS {
	t.Helper()
	f := &fakeDNS{handle: h, conns: map[net.Conn]bool{}, recv: make(chan wireQuery, 64), tcpEOF: make(chan struct{}, 64)}
	for try := 0; ; try++ {
		u, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			t.Fatal(err)
		}
		l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: u.LocalAddr().(*net.UDPAddr).Port})
		if err != nil {
			_ = u.Close()
			if try < 20 {
				continue
			}
			t.Fatal(err)
		}
		f.udp, f.tcp, f.addr = u, l, u.LocalAddr().String()
		break
	}
	f.wg.Add(2)
	go f.serveUDP()
	go f.serveTCP()
	t.Cleanup(func() {
		_ = f.udp.Close()
		_ = f.tcp.Close()
		f.mu.Lock()
		for c := range f.conns {
			_ = c.Close()
		}
		f.mu.Unlock()
		f.wg.Wait()
	})
	return f
}

func (f *fakeDNS) record(transport string, msg []byte) (wireQuery, bool) {
	var p dnsmessage.Parser
	h, err := p.Start(msg)
	if err != nil {
		return wireQuery{}, false
	}
	q, err := p.Question()
	if err != nil {
		return wireQuery{}, false
	}
	w := wireQuery{transport: transport, id: h.ID, name: q.Name.String(), question: append([]byte(nil), msg[dnsHeaderLen:]...), at: time.Now()}
	f.mu.Lock()
	f.got = append(f.got, w)
	f.mu.Unlock()
	select {
	case f.recv <- w:
	default:
	}
	return w, true
}

func (f *fakeDNS) queries(transport string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, q := range f.got {
		if q.transport == transport {
			n++
		}
	}
	return n
}

func (f *fakeDNS) first() wireQuery {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.got[0]
}

func (f *fakeDNS) serveUDP() {
	defer f.wg.Done()
	buf := make([]byte, 65536)
	for {
		n, from, err := f.udp.ReadFromUDP(buf)
		if err != nil {
			return
		}
		q, ok := f.record("udp", buf[:n])
		if !ok {
			continue
		}
		for _, r := range f.handle(q) {
			_, _ = f.udp.WriteToUDP(r, from)
		}
	}
}

func (f *fakeDNS) serveTCP() {
	defer f.wg.Done()
	for {
		c, err := f.tcp.AcceptTCP()
		if err != nil {
			return
		}
		f.mu.Lock()
		f.conns[c] = true
		f.mu.Unlock()
		f.wg.Add(1)
		go func() {
			defer f.wg.Done()
			defer c.Close()
			var l [2]byte
			if _, err := io.ReadFull(c, l[:]); err != nil {
				return
			}
			msg := make([]byte, binary.BigEndian.Uint16(l[:]))
			if _, err := io.ReadFull(c, msg); err != nil {
				return
			}
			q, ok := f.record("tcp", msg)
			if !ok {
				return
			}
			if replies := f.handle(q); len(replies) > 0 {
				for _, r := range replies {
					_, _ = c.Write(r)
				}
				_ = c.CloseWrite() // no further bytes: a short frame ends in EOF, not a hang
			}
			if _, err := io.Copy(io.Discard, c); err == nil {
				select {
				case f.tcpEOF <- struct{}{}: // the client closed its side
				default:
				}
			}
		}()
	}
}

// Reply construction (raw bytes; no name compression, so owner case and malformations stay exactly as scripted).

func header(id, flags, qd, an, ns, ar uint16) []byte {
	b := binary.BigEndian.AppendUint16(nil, id)
	for _, v := range []uint16{flags, qd, an, ns, ar} {
		b = binary.BigEndian.AppendUint16(b, v)
	}
	return b
}

func rrClass(name string, typ, class uint16, data []byte) []byte {
	b := wireName(name)
	b = binary.BigEndian.AppendUint16(b, typ)
	b = binary.BigEndian.AppendUint16(b, class)
	b = binary.BigEndian.AppendUint32(b, 60)
	b = binary.BigEndian.AppendUint16(b, uint16(len(data)))
	return append(b, data...)
}

func cnameRR(owner, target string) []byte { return rr(owner, 5, wireName(target)) }
func aRR(owner string) []byte             { return rr(owner, 1, []byte{192, 0, 2, 1}) }

// reply answers q with the given flags and answer records, echoing its question.
func reply(q wireQuery, flags uint16, answers ...[]byte) []byte {
	m := append(header(q.id, flags, 1, uint16(len(answers)), 0, 0), q.question...)
	for _, a := range answers {
		m = append(m, a...)
	}
	return m
}

func frame(m []byte) []byte { return append(binary.BigEndian.AppendUint16(nil, uint16(len(m))), m...) }

func one(m []byte) [][]byte { return [][]byte{m} }

// over sends a well-formed message over the query's transport (framed on TCP).
func over(q wireQuery, m []byte) [][]byte {
	if q.transport == "tcp" {
		return one(frame(m))
	}
	return one(m)
}

// question builds a question section.
func question(name string, typ, class uint16) []byte {
	b := wireName(name)
	b = binary.BigEndian.AppendUint16(b, typ)
	return binary.BigEndian.AppendUint16(b, class)
}

func loopbackChecker(useVC bool, servers ...string) SystemChecker {
	return newSystemChecker(nil, func() resolvConf { return resolvConf{servers: servers, useVC: useVC} })
}

func shortCtx(t *testing.T, d time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}

// txtOK isolates the CNAME path in Proof: the TXT challenge always matches.
type txtOK struct{ SystemChecker }

func (txtOK) LookupTXT(context.Context, string) ([]string, error) { return []string{testValue}, nil }

// chainHandler answers a CNAME query like the defect's resolver: the whole chain from the queried name plus the
// terminal address, in the given order; a chain end without a CNAME is NODATA, an unknown name NXDOMAIN.
func chainHandler(chain map[string]string, order string) wireHandler {
	return func(q wireQuery) [][]byte {
		n := strings.TrimSuffix(q.name, ".")
		var edges [][]byte
		end := n
		for hops := 0; hops < 16; hops++ {
			t, ok := chain[end]
			if !ok {
				break
			}
			edges = append(edges, cnameRR(end, t))
			end = t
		}
		known := len(edges) > 0
		for _, t := range chain {
			known = known || t == n
		}
		if !known {
			return over(q, reply(q, flagsOK|3))
		}
		var answers [][]byte
		switch order {
		case "forward":
			answers = append(edges, aRR(end))
		case "terminal-first":
			answers = append([][]byte{aRR(end)}, edges...)
		case "reversed":
			answers = [][]byte{aRR(end)}
			for i := len(edges) - 1; i >= 0; i-- {
				answers = append(answers, edges[i])
			}
		case "rotated":
			if len(edges) > 1 {
				answers = append(append(append([][]byte{}, edges[1:]...), aRR(end)), edges[0])
			} else {
				answers = append([][]byte{aRR(end)}, edges...)
			}
		}
		if len(edges) == 0 {
			answers = nil // NODATA at the chain's end
		}
		return over(q, reply(q, flagsOK, answers...))
	}
}

func TestDNSWireChains(t *testing.T) {
	chain := map[string]string{
		"one.example.com":   testTarget,
		"chain.example.com": "hop1.example.org", "hop1.example.org": "hop2.example.org", "hop2.example.org": testTarget,
		"long.example.com": "l1.example.org", "l1.example.org": "l2.example.org", "l2.example.org": "l3.example.org",
		"l3.example.org": testTarget,
	}
	cfg := ProofConfig{ChallengeLabel: "_platform-verify", CNAMETarget: testTarget, BaseDomain: "example.net", Timeout: 2 * time.Second}
	for _, order := range []string{"forward", "terminal-first", "reversed", "rotated"} {
		for _, transport := range []string{"udp", "tcp"} {
			f := newFakeDNS(t, chainHandler(chain, order))
			sys := loopbackChecker(transport == "tcp", f.addr)
			for name, want := range map[string]string{"long.example.com": "l1.example.org.", "l1.example.org": "l2.example.org.",
				"chain.example.com": "hop1.example.org.", "one.example.com": testTarget + "."} {
				if got, err := sys.LookupCNAME(shortCtx(t, 2*time.Second), name); err != nil || got != want {
					t.Errorf("%s/%s: LookupCNAME(%s) = %q, %v; want the owner's edge %s", order, transport, name, got, err, want)
				}
			}
			for _, k := range []struct {
				host string
				want Outcome
			}{
				{"one.example.com", OutcomeOK},        // one edge
				{"chain.example.com", OutcomeOK},      // three edges
				{"long.example.com", OutcomeMismatch}, // four edges: over the bound although the chain reaches the target
			} {
				if got := Proof(context.Background(), txtOK{sys}, cfg, k.host, testValue); got != k.want {
					t.Errorf("%s/%s: Proof(%s) = %s; want %s", order, transport, k.host, got, k.want)
				}
			}
		}
	}
}

func TestDNSWireOwnerSelection(t *testing.T) {
	const owner = "chain.example.com"
	cases := []struct {
		name    string
		answers func(q wireQuery) [][]byte
		want    string
		wantErr error
	}{
		{"only another owner's CNAME", func(wireQuery) [][]byte {
			return [][]byte{cnameRR("hop1.example.org", testTarget)}
		}, owner + ".", nil},
		{"owner differing only in case", func(wireQuery) [][]byte {
			return [][]byte{cnameRR("CHAIN.Example.COM", "Hop1.EXAMPLE.org")}
		}, "hop1.example.org.", nil},
		{"two CNAMEs at the owner", func(wireQuery) [][]byte {
			return [][]byte{cnameRR(owner, testTarget), cnameRR(owner, "elsewhere.example.org")}
		}, "", ErrDNSTransient},
		{"two CNAMEs at the owner, reversed", func(wireQuery) [][]byte {
			return [][]byte{cnameRR(owner, "elsewhere.example.org"), cnameRR(owner, testTarget)}
		}, "", ErrDNSTransient},
		{"duplicate identical CNAMEs", func(wireQuery) [][]byte {
			return [][]byte{cnameRR(owner, testTarget), cnameRR(owner, testTarget)}
		}, "", ErrDNSTransient},
		{"CH-class CNAME at the owner", func(wireQuery) [][]byte {
			return [][]byte{rrClass(owner, 5, 3, wireName(testTarget))}
		}, owner + ".", nil},
		{"address record only", func(wireQuery) [][]byte { return [][]byte{aRR(owner)} }, owner + ".", nil},
		{"unrelated records around the edge", func(wireQuery) [][]byte {
			return [][]byte{aRR(owner), cnameRR("other.example.org", "x.example.org"), cnameRR(owner, "hop1.example.org")}
		}, "hop1.example.org.", nil},
		{"bad target characters", func(wireQuery) [][]byte {
			return [][]byte{cnameRR(owner, "bad*name.example.org")}
		}, "", ErrDNSTransient},
		{"root target", func(wireQuery) [][]byte { return [][]byte{rr(owner, 5, []byte{0})} }, "", ErrDNSTransient},
	}
	for _, k := range cases {
		f := newFakeDNS(t, func(q wireQuery) [][]byte { return one(reply(q, flagsOK, k.answers(q)...)) })
		got, err := loopbackChecker(false, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), "Chain.Example.com")
		if got != k.want || err != k.wantErr {
			t.Errorf("%s: LookupCNAME = %q, %v; want %q, %v", k.name, got, err, k.want, k.wantErr)
		}
	}
	// Mismatched owner and ambiguity end the proof without an OK.
	cfg := ProofConfig{ChallengeLabel: "_v", CNAMETarget: testTarget, BaseDomain: "example.net", Timeout: time.Second}
	for _, k := range []struct {
		answers [][]byte
		want    Outcome
	}{
		{[][]byte{cnameRR("hop1.example.org", testTarget)}, OutcomeMismatch},
		{[][]byte{cnameRR(owner, testTarget), cnameRR(owner, "x.example.org")}, OutcomeTransient},
	} {
		f := newFakeDNS(t, func(q wireQuery) [][]byte { return one(reply(q, flagsOK, k.answers...)) })
		if got := Proof(context.Background(), txtOK{loopbackChecker(false, f.addr)}, cfg, owner, testValue); got != k.want {
			t.Errorf("Proof = %s; want %s", got, k.want)
		}
	}
}

func TestDNSWireRCodes(t *testing.T) {
	for _, k := range []struct {
		rcode   uint16
		want    string
		wantErr error
	}{
		{0, "plain.example.com.", nil}, // NODATA: the name itself
		{3, "", ErrNXDomain},
		{1, "", ErrDNSTransient}, // FORMERR
		{2, "", ErrDNSTransient}, // SERVFAIL
		{4, "", ErrDNSTransient}, // NOTIMP
		{5, "", ErrDNSTransient}, // REFUSED
	} {
		for _, useVC := range []bool{false, true} {
			f := newFakeDNS(t, func(q wireQuery) [][]byte {
				m := reply(q, flagsOK|k.rcode)
				if q.transport == "tcp" {
					m = frame(m)
				}
				return one(m)
			})
			got, err := loopbackChecker(useVC, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), "plain.example.com")
			if got != k.want || err != k.wantErr {
				t.Errorf("rcode %d (use-vc %v): %q, %v; want %q, %v", k.rcode, useVC, got, err, k.want, k.wantErr)
			}
		}
	}
	// NXDOMAIN after the first hop is a mismatch, at the first hop nxdomain.
	cfg := ProofConfig{ChallengeLabel: "_v", CNAMETarget: testTarget, BaseDomain: "example.net", Timeout: time.Second}
	f := newFakeDNS(t, func(q wireQuery) [][]byte {
		if q.name == "dangling.example.com." {
			return one(reply(q, flagsOK, cnameRR("dangling.example.com", "gone.example.org")))
		}
		return one(reply(q, flagsOK|3))
	})
	sys := txtOK{loopbackChecker(false, f.addr)}
	if got := Proof(context.Background(), sys, cfg, "dangling.example.com", testValue); got != OutcomeMismatch {
		t.Errorf("dangling chain: %s", got)
	}
	if got := Proof(context.Background(), sys, cfg, "nothing.example.com", testValue); got != OutcomeNXDomain {
		t.Errorf("missing name: %s", got)
	}
}

func TestDNSWireTruncation(t *testing.T) {
	const name = "trunc.example.com"
	// UDP TC → the same query over TCP, whose answer is used.
	f := newFakeDNS(t, func(q wireQuery) [][]byte {
		if q.transport == "udp" {
			return one(reply(q, flagsOK|flagTC))
		}
		return one(frame(reply(q, flagsOK, cnameRR(name, "tcp.example.org"))))
	})
	if got, err := loopbackChecker(false, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name); err != nil || got != "tcp.example.org." {
		t.Errorf("TC fallback: %q, %v", got, err)
	}
	if f.queries("udp") != 1 || f.queries("tcp") != 1 {
		t.Errorf("TC fallback queries: udp %d tcp %d", f.queries("udp"), f.queries("tcp"))
	}
	// A datagram filling (or exceeding) the read buffer may be locally truncated: never parsed for an edge, retried
	// over TCP.
	for _, size := range []int{maxDNSMessage, maxDNSMessage + 904} {
		f := newFakeDNS(t, func(q wireQuery) [][]byte {
			if q.transport == "udp" {
				m := reply(q, flagsOK, cnameRR(name, "udp.example.org"))
				return one(append(m, make([]byte, size-len(m))...))
			}
			return one(frame(reply(q, flagsOK, cnameRR(name, "tcp.example.org"))))
		})
		if got, err := loopbackChecker(false, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name); err != nil || got != "tcp.example.org." {
			t.Errorf("full buffer %d: %q, %v", size, got, err)
		}
	}
	// The envelope is validated before TC is acted on: TC datagrams with a wrong ID, a wrong question or no response
	// bit are discarded, not a reason to fall back; the following valid UDP answer is used and TCP is never dialed.
	f = newFakeDNS(t, func(q wireQuery) [][]byte {
		if q.transport == "tcp" {
			return one(frame(reply(q, flagsOK, cnameRR(name, "tcp.example.org"))))
		}
		wrongID := reply(q, flagsOK|flagTC)
		wrongID[0] ^= 0xff
		wrongQ := append(header(q.id, flagsOK|flagTC, 1, 0, 0, 0), question("other.example.com", 5, 1)...)
		query := reply(q, 0x0100|flagTC) // QR clear
		return [][]byte{wrongID, wrongQ, query, reply(q, flagsOK, cnameRR(name, "udp.example.org"))}
	})
	if got, err := loopbackChecker(false, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name); err != nil || got != "udp.example.org." {
		t.Errorf("foreign TC datagrams: %q, %v", got, err)
	}
	if f.queries("tcp") != 0 {
		t.Errorf("foreign TC datagrams caused a TCP retry")
	}
	// TC over TCP, and bad TCP framing, are transient.
	for label, raw := range map[string]func(q wireQuery) []byte{
		"TC over TCP":            func(q wireQuery) []byte { return frame(reply(q, flagsOK|flagTC, cnameRR(name, "x.example.org"))) },
		"zero length":            func(wireQuery) []byte { return []byte{0, 0} },
		"shorter than a header":  func(wireQuery) []byte { return append([]byte{0, 5}, 1, 2, 3, 4, 5) },
		"longer than 4096":       func(q wireQuery) []byte { return frame(append(reply(q, flagsOK), make([]byte, 4100)...)) },
		"length beyond the data": func(q wireQuery) []byte { m := reply(q, flagsOK); return append([]byte{0, byte(len(m) + 20)}, m...) },
		"truncated length":       func(wireQuery) []byte { return []byte{0} },
	} {
		f := newFakeDNS(t, func(q wireQuery) [][]byte { return one(raw(q)) })
		start := time.Now()
		got, err := loopbackChecker(true, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name)
		if err != ErrDNSTransient || got != "" {
			t.Errorf("%s: %q, %v; want transient", label, got, err)
		}
		if time.Since(start) > time.Second {
			t.Errorf("%s: took %s; a framing failure must not wait for the deadline", label, time.Since(start))
		}
	}
}

func TestDNSWireEnvelope(t *testing.T) {
	const name = "env.example.com"
	good := func(q wireQuery) []byte { return reply(q, flagsOK, cnameRR(name, "good.example.org")) }
	foreign := map[string]func(q wireQuery) []byte{
		"wrong ID": func(q wireQuery) []byte { m := good(q); m[0] ^= 0xff; return m },
		"wrong name": func(q wireQuery) []byte {
			return append(header(q.id, flagsOK, 1, 1, 0, 0), append(question("other.example.com", 5, 1), cnameRR("other.example.com", "bad.example.org")...)...)
		},
		"wrong type":     func(q wireQuery) []byte { return append(header(q.id, flagsOK, 1, 0, 0, 0), question(name, 1, 1)...) },
		"wrong class":    func(q wireQuery) []byte { return append(header(q.id, flagsOK, 1, 0, 0, 0), question(name, 5, 3)...) },
		"not a response": func(q wireQuery) []byte { return reply(q, 0x0100, cnameRR(name, "bad.example.org")) },
		"opcode STATUS":  func(q wireQuery) []byte { return reply(q, flagsOK|2<<11, cnameRR(name, "bad.example.org")) },
		"short datagram": func(wireQuery) []byte { return []byte{1, 2, 3} },
	}
	for label, bad := range foreign {
		// UDP: discarded, the following valid datagram is used.
		f := newFakeDNS(t, func(q wireQuery) [][]byte { return [][]byte{bad(q), good(q)} })
		if got, err := loopbackChecker(false, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name); err != nil || got != "good.example.org." {
			t.Errorf("UDP %s then valid: %q, %v", label, got, err)
		}
		// TCP: one answer per stream; a foreign answer is transient.
		f = newFakeDNS(t, func(q wireQuery) [][]byte { return one(frame(bad(q))) })
		if got, err := loopbackChecker(true, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name); err != ErrDNSTransient {
			t.Errorf("TCP %s: %q, %v; want transient", label, got, err)
		}
	}
	// UDP with only a wrong-ID answer: transient at the deadline.
	f := newFakeDNS(t, func(q wireQuery) [][]byte { return one(foreign["wrong ID"](q)) })
	if got, err := loopbackChecker(false, f.addr).LookupCNAME(shortCtx(t, 200*time.Millisecond), name); err != ErrDNSTransient {
		t.Errorf("wrong ID only: %q, %v", got, err)
	}
}

func TestDNSWireMalformed(t *testing.T) {
	const name = "mal.example.com"
	edgeRR := cnameRR(name, "x.example.org")
	withHeader := func(q wireQuery, an, ns, ar uint16, rest ...[]byte) []byte {
		m := append(header(q.id, flagsOK, 1, an, ns, ar), q.question...)
		for _, r := range rest {
			m = append(m, r...)
		}
		return m
	}
	owner := wireName(name)
	// declared is the edge record with its RDLENGTH set to n, the encoded target bytes left in place after it.
	declared := func(n int) func(q wireQuery) []byte {
		return func(q wireQuery) []byte {
			r := append([]byte(nil), edgeRR...)
			binary.BigEndian.PutUint16(r[len(owner)+8:], uint16(n))
			return withHeader(q, 1, 0, 0, r)
		}
	}
	exampleCom := byte(dnsHeaderLen + len("mal.")) // offset of "example.com" inside the echoed question name
	cases := map[string]func(q wireQuery) []byte{
		"CNAME RDLENGTH 0, target after it":      declared(0),
		"CNAME RDLENGTH 1, target after it":      declared(1),
		"CNAME RDLENGTH one short of the target": declared(len(wireName("x.example.org")) - 1),
		"CNAME name padded inside its RDATA": func(q wireQuery) []byte {
			return withHeader(q, 1, 0, 0, rr(name, 5, append(wireName("x.example.org"), 0, 0, 0)))
		},
		"CNAME pointer padded inside its RDATA": func(q wireQuery) []byte {
			return withHeader(q, 1, 0, 0, rr(name, 5, []byte{1, 'x', 0xc0, exampleCom, 0}))
		},
		"cut mid-record": func(q wireQuery) []byte { return withHeader(q, 1, 0, 0, edgeRR[:len(edgeRR)-4]) },
		"RDLENGTH beyond the message": func(q wireQuery) []byte {
			r := append([]byte(nil), edgeRR...)
			binary.BigEndian.PutUint16(r[len(owner)+8:], uint16(len(r)))
			return withHeader(q, 1, 0, 0, r)
		},
		"compression pointer loop": func(q wireQuery) []byte {
			at := dnsHeaderLen + len(q.question)
			loop := []byte{0xc0 | byte(at>>8), byte(at)}
			return withHeader(q, 1, 0, 0, append(loop, rr(name, 5, wireName("x.example.org"))[len(owner):]...))
		},
		"dotted label in the target": func(q wireQuery) []byte {
			return withHeader(q, 1, 0, 0, rr(name, 5, append([]byte{3, 'a', '.', 'b'}, wireName("example.org")...)))
		},
		"label longer than 63": func(q wireQuery) []byte {
			return withHeader(q, 1, 0, 0, rr(name, 5, append(append([]byte{64}, make([]byte, 64)...), 0)))
		},
		"no question": func(q wireQuery) []byte { return append(header(q.id, flagsOK, 0, 1, 0, 0), edgeRR...) },
		"two questions": func(q wireQuery) []byte {
			return append(append(header(q.id, flagsOK, 2, 1, 0, 0), append(q.question, q.question...)...), edgeRR...)
		},
		"answer count beyond the records": func(q wireQuery) []byte { return withHeader(q, 2, 0, 0, edgeRR) },
		"malformed unrelated authority": func(q wireQuery) []byte {
			soa := rr("example.com", 6, []byte{1, 2, 3})
			return withHeader(q, 1, 1, 0, edgeRR, soa)
		},
		"malformed unrelated additional": func(q wireQuery) []byte {
			extra := aRR("glue.example.org")
			return withHeader(q, 1, 0, 1, edgeRR, extra[:len(extra)-2])
		},
	}
	for label, bad := range cases {
		for _, useVC := range []bool{false, true} {
			f := newFakeDNS(t, func(q wireQuery) [][]byte {
				m := bad(q)
				if q.transport == "tcp" {
					m = frame(m)
				}
				return one(m)
			})
			start := time.Now()
			got, err := loopbackChecker(useVC, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name)
			if err != ErrDNSTransient || got != "" {
				t.Errorf("%s (use-vc %v): %q, %v; want transient", label, useVC, got, err)
			}
			if time.Since(start) > time.Second {
				t.Errorf("%s (use-vc %v): took %s; a matching malformed answer fails at once", label, useVC, time.Since(start))
			}
		}
	}
	// The fixtures are well-formed apart from the scripted defect: the same shapes without it are accepted.
	f := newFakeDNS(t, func(q wireQuery) [][]byte {
		return one(withHeader(q, 1, 1, 1, edgeRR, rr("example.com", 1, []byte{192, 0, 2, 9}), aRR("glue.example.org")))
	})
	if got, err := loopbackChecker(false, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name); err != nil || got != "x.example.org." {
		t.Errorf("well-formed control: %q, %v", got, err)
	}
	// Legal compression is kept: an owner pointing at the question name and a target ending in a pointer (its RDATA
	// is exactly label + two-byte pointer) supply the edge on both transports.
	compressed := func(q wireQuery) []byte {
		r := []byte{0xc0, dnsHeaderLen, 0, 5, 0, 1, 0, 0, 0, 60, 0, 4, 1, 'x', 0xc0, exampleCom}
		return withHeader(q, 1, 0, 0, r)
	}
	for _, useVC := range []bool{false, true} {
		f := newFakeDNS(t, func(q wireQuery) [][]byte { return over(q, compressed(q)) })
		if got, err := loopbackChecker(useVC, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name); err != nil || got != "x.example.com." {
			t.Errorf("compressed control (use-vc %v): %q, %v", useVC, got, err)
		}
	}
}

func TestDNSWireUseVC(t *testing.T) {
	const name = "vc.example.com"
	f := newFakeDNS(t, func(q wireQuery) [][]byte {
		if q.transport == "udp" {
			return one(reply(q, flagsOK, cnameRR(name, "udp.example.org")))
		}
		return one(frame(reply(q, flagsOK, cnameRR(name, "tcp.example.org"))))
	})
	if got, err := loopbackChecker(true, f.addr).LookupCNAME(shortCtx(t, 2*time.Second), name); err != nil || got != "tcp.example.org." {
		t.Errorf("use-vc: %q, %v", got, err)
	}
	if f.queries("udp") != 0 || f.queries("tcp") != 1 {
		t.Errorf("use-vc queries: udp %d tcp %d", f.queries("udp"), f.queries("tcp"))
	}
}

func TestDNSWireResolvConf(t *testing.T) {
	c := parseResolvConf(strings.Join([]string{
		"# comment",
		"; comment",
		"search corp.example.com",
		"nameserver 192.0.2.1",
		"nameserver not-an-address",
		"nameserver 2001:db8::1",
		"nameserver fe80::1%eth0",
		"nameserver 192.0.2.4", // a fourth server is ignored
		"options ndots:5 rotate timeout:1 attempts:9 edns0 trust-ad",
	}, "\n"))
	want := []string{"192.0.2.1:53", "[2001:db8::1]:53", "[fe80::1%eth0]:53"}
	if !reflect.DeepEqual(c.servers, want) || c.useVC {
		t.Errorf("parse = %+v; want %v without use-vc", c, want)
	}
	for _, opt := range []string{"use-vc", "usevc", "tcp"} {
		if c := parseResolvConf("nameserver 192.0.2.1\noptions rotate " + opt); !c.useVC {
			t.Errorf("options %s not honored", opt)
		}
	}
	if c := parseResolvConf("search example.com\n"); !reflect.DeepEqual(c.servers, defaultNameservers) {
		t.Errorf("no nameserver: %v; want Go's defaults", c.servers)
	}
	if c := readResolvConf(filepath.Join(t.TempDir(), "missing")); !reflect.DeepEqual(c.servers, defaultNameservers) {
		t.Errorf("missing file: %v; want Go's defaults", c.servers)
	}
	p := filepath.Join(t.TempDir(), "resolv.conf")
	if err := os.WriteFile(p, []byte("nameserver 192.0.2.7\noptions use-vc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if c := readResolvConf(p); !reflect.DeepEqual(c.servers, []string{"192.0.2.7:53"}) || !c.useVC {
		t.Errorf("read = %+v", c)
	}
	// Only the configured (here: default) servers are ever dialed, in order, over UDP.
	var mu sync.Mutex
	var dialed []string
	sys := newSystemChecker(func(_ context.Context, network, addr string) (net.Conn, error) {
		mu.Lock()
		defer mu.Unlock()
		dialed = append(dialed, network+" "+addr)
		return nil, errors.New("unreachable")
	}, func() resolvConf { return parseResolvConf("") })
	if _, err := sys.LookupCNAME(shortCtx(t, time.Second), "x.example.com"); err != ErrDNSTransient {
		t.Errorf("unreachable defaults: %v", err)
	}
	if !reflect.DeepEqual(dialed, []string{"udp 127.0.0.1:53", "udp [::1]:53"}) {
		t.Errorf("dialed %v", dialed)
	}
}

func TestDNSWireServerFallback(t *testing.T) {
	const name = "fb.example.com"
	answer := func(q wireQuery) [][]byte { return one(reply(q, flagsOK, cnameRR(name, "ok.example.org"))) }
	mute := func(wireQuery) [][]byte { return nil }
	// A mute first server gets half of the total; the second answers within the rest.
	a, b := newFakeDNS(t, mute), newFakeDNS(t, answer)
	start := time.Now()
	got, err := loopbackChecker(false, a.addr, b.addr).LookupCNAME(shortCtx(t, time.Second), name)
	if err != nil || got != "ok.example.org." {
		t.Errorf("mute then answering: %q, %v", got, err)
	}
	if b.queries("udp") != 1 {
		t.Fatalf("second server queried %d times", b.queries("udp"))
	}
	if d := b.first().at.Sub(start); d < 400*time.Millisecond || d > 900*time.Millisecond {
		t.Errorf("second server asked after %s; want about half of the 1s total", d)
	}
	// SERVFAIL and an unreachable (closed) port yield to the next server at once.
	sf := newFakeDNS(t, func(q wireQuery) [][]byte { return one(reply(q, flagsOK|2)) })
	closed, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	closedAddr := closed.LocalAddr().String()
	_ = closed.Close()
	for label, first := range map[string]string{"SERVFAIL": sf.addr, "closed port": closedAddr} {
		b := newFakeDNS(t, answer)
		got, err := loopbackChecker(false, first, b.addr).LookupCNAME(shortCtx(t, 2*time.Second), name)
		if err != nil || got != "ok.example.org." {
			t.Errorf("%s then answering: %q, %v", label, got, err)
		}
	}
	// NXDOMAIN from the first server is an answer: the second is not asked.
	nx, other := newFakeDNS(t, func(q wireQuery) [][]byte { return one(reply(q, flagsOK|3)) }), newFakeDNS(t, answer)
	if _, err := loopbackChecker(false, nx.addr, other.addr).LookupCNAME(shortCtx(t, 2*time.Second), name); err != ErrNXDomain || other.queries("udp") != 0 {
		t.Errorf("NXDOMAIN first: %v, second asked %d", err, other.queries("udp"))
	}
	// Three mute servers: transient within the total, each asked once with its share.
	m1, m2, m3 := newFakeDNS(t, mute), newFakeDNS(t, mute), newFakeDNS(t, mute)
	start = time.Now()
	if _, err := loopbackChecker(false, m1.addr, m2.addr, m3.addr).LookupCNAME(shortCtx(t, 600*time.Millisecond), name); err != ErrDNSTransient {
		t.Errorf("three mute servers: %v", err)
	}
	if d := time.Since(start); d < 550*time.Millisecond || d > 1500*time.Millisecond {
		t.Errorf("three mute servers took %s; want the 600ms total", d)
	}
	for i, m := range []*fakeDNS{m1, m2, m3} {
		if m.queries("udp") != 1 {
			t.Errorf("mute server %d asked %d times", i+1, m.queries("udp"))
		}
	}
	if d := m3.first().at.Sub(start); d < 300*time.Millisecond {
		t.Errorf("third server asked after %s; the first two must each have had their share", d)
	}
}

func TestDNSWireDeadlines(t *testing.T) {
	const name = "slow.example.com"
	mute := func(wireQuery) [][]byte { return nil }
	// Mute UDP and TCP servers are transient within the proof's per-lookup timeout.
	cfg := ProofConfig{ChallengeLabel: "_v", CNAMETarget: testTarget, BaseDomain: "example.net", Timeout: 200 * time.Millisecond}
	for _, useVC := range []bool{false, true} {
		f := newFakeDNS(t, mute)
		start := time.Now()
		if got := Proof(context.Background(), txtOK{loopbackChecker(useVC, f.addr)}, cfg, name, testValue); got != OutcomeTransient {
			t.Errorf("mute (use-vc %v): %s", useVC, got)
		}
		if d := time.Since(start); d > time.Second {
			t.Errorf("mute (use-vc %v): took %s", useVC, d)
		}
	}
	// Without a caller deadline the default bound applies (shortened here); with one, the caller's deadline is kept even
	// when it is longer than the default.
	f := newFakeDNS(t, mute)
	sys := loopbackChecker(false, f.addr)
	sys.cname.bound = 300 * time.Millisecond
	start := time.Now()
	if _, err := sys.LookupCNAME(context.Background(), name); err != ErrDNSTransient {
		t.Errorf("no deadline, mute: %v", err)
	}
	if d := time.Since(start); d < 250*time.Millisecond || d > 1500*time.Millisecond {
		t.Errorf("no deadline, mute: took %s; want the 300ms default bound", d)
	}
	late := newFakeDNS(t, func(q wireQuery) [][]byte {
		time.Sleep(600 * time.Millisecond)
		return one(reply(q, flagsOK, cnameRR(name, "late.example.org")))
	})
	sys = loopbackChecker(false, late.addr)
	sys.cname.bound = 300 * time.Millisecond
	if got, err := sys.LookupCNAME(shortCtx(t, 3*time.Second), name); err != nil || got != "late.example.org." {
		t.Errorf("caller deadline longer than the default: %q, %v", got, err)
	}
	if defaultLookupBound != 5*time.Second || NewSystemChecker().cname.bound != defaultLookupBound {
		t.Error("the production default bound must be 5s")
	}
	// A canceled context sends nothing.
	idle := newFakeDNS(t, mute)
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := loopbackChecker(false, idle.addr).LookupCNAME(canceled, name); err != ErrDNSTransient {
		t.Errorf("canceled: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if idle.queries("udp")+idle.queries("tcp") != 0 {
		t.Error("a canceled lookup sent a query")
	}
}

// trackedUDP and trackedTCP record that the exchange closed its connection; trackedUDP keeps the PacketConn shape.
type trackedUDP struct {
	*net.UDPConn
	closed *atomic.Bool
}

func (c trackedUDP) Close() error { c.closed.Store(true); return c.UDPConn.Close() }

type trackedTCP struct {
	net.Conn
	closed *atomic.Bool
}

func (c trackedTCP) Close() error { c.closed.Store(true); return c.Conn.Close() }

func trackingDial(closed *atomic.Bool) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		c, err := (&net.Dialer{}).DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		if u, ok := c.(*net.UDPConn); ok {
			return trackedUDP{u, closed}, nil
		}
		return trackedTCP{c, closed}, nil
	}
}

func TestDNSWireCancelMidRead(t *testing.T) {
	for _, useVC := range []bool{false, true} {
		f := newFakeDNS(t, func(wireQuery) [][]byte { return nil })
		var closed atomic.Bool
		sys := newSystemChecker(trackingDial(&closed), func() resolvConf { return resolvConf{servers: []string{f.addr}, useVC: useVC} })
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		go func() {
			<-f.recv // the query is in flight and the exchange is blocked reading
			cancel()
		}()
		start := time.Now()
		_, err := sys.LookupCNAME(ctx, "cancel.example.com")
		cancel()
		if err != ErrDNSTransient {
			t.Errorf("use-vc %v: %v; want transient", useVC, err)
		}
		if d := time.Since(start); d > 2*time.Second {
			t.Errorf("use-vc %v: cancellation took %s; the blocked read must be unblocked at once", useVC, d)
		}
		if !closed.Load() {
			t.Errorf("use-vc %v: connection not closed when the lookup returned", useVC)
		}
		if useVC {
			select {
			case <-f.tcpEOF: // the server saw the client close the stream
			case <-time.After(2 * time.Second):
				t.Error("TCP server did not observe the close")
			}
		}
	}
}

// blockingCloseConn's Close blocks until released, to observe whether closeOnDone's release waits for a running callback.
type blockingCloseConn struct {
	net.Conn
	entered  chan struct{}
	unblock  chan struct{}
	calls    atomic.Int32
	finished atomic.Bool
}

func (c *blockingCloseConn) Close() error {
	c.calls.Add(1)
	close(c.entered)
	<-c.unblock
	c.finished.Store(true)
	return nil
}

func TestDNSWireCloseOnDone(t *testing.T) {
	// A callback that has started is joined: release returns only after it finished.
	c := &blockingCloseConn{entered: make(chan struct{}), unblock: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	release := closeOnDone(ctx, c)
	cancel()
	<-c.entered
	returned := make(chan struct{})
	go func() {
		release()
		close(returned)
	}()
	select {
	case <-returned:
		t.Fatal("release returned while the close callback was still running")
	case <-time.After(100 * time.Millisecond):
	}
	close(c.unblock)
	select {
	case <-returned:
	case <-time.After(2 * time.Second):
		t.Fatal("release did not return after the callback finished")
	}
	if !c.finished.Load() || c.calls.Load() != 1 {
		t.Errorf("callback finished %v, calls %d", c.finished.Load(), c.calls.Load())
	}
	// Released before the context ends: the callback never runs.
	c2 := &blockingCloseConn{entered: make(chan struct{}), unblock: make(chan struct{})}
	ctx2, cancel2 := context.WithCancel(context.Background())
	closeOnDone(ctx2, c2)()
	cancel2()
	time.Sleep(50 * time.Millisecond)
	if c2.calls.Load() != 0 {
		t.Error("callback ran after release")
	}
}
