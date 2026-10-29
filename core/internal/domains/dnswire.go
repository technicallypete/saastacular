package domains

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// The CNAME edge lookup: one typed CNAME query for the exact absolute owner, sent to the trusted nameservers of the
// system configuration, answered by that owner's own immediate edge. It does not reproduce the native resolver's
// policy: of /etc/resolv.conf it interprets only the first three "nameserver" lines (port 53; Go's loopback defaults
// when there are none) and "options use-vc" (also "usevc", "tcp": TCP only). Everything else (search, domain, ndots,
// rotate, attempts, timeout, single-request, edns0, trust-ad, sortlist and any other option) is deliberately ignored:
// the query is absolute, each server gets one attempt in order, and the time budget below replaces timeout/attempts.
// No EDNS is sent; a truncated UDP answer is retried over TCP. There is no DNSSEC validation.

const (
	resolvConfPath     = "/etc/resolv.conf"
	maxNameservers     = 3
	defaultLookupBound = 5 * time.Second // only when the caller's context has no deadline
	maxDNSMessage      = 4096            // UDP read buffer and largest accepted TCP frame
	dnsHeaderLen       = 12
)

// defaultNameservers are Go's own defaults when resolv.conf is unreadable or lists no usable nameserver.
var defaultNameservers = []string{"127.0.0.1:53", "[::1]:53"}

// resolvConf is the interpreted part of the system resolver configuration.
type resolvConf struct {
	servers []string // host:port, at most maxNameservers
	useVC   bool     // TCP only
}

func readResolvConf(path string) resolvConf {
	b, err := os.ReadFile(path)
	if err != nil {
		return resolvConf{servers: defaultNameservers}
	}
	return parseResolvConf(string(b))
}

func parseResolvConf(s string) resolvConf {
	var c resolvConf
	for _, line := range strings.Split(s, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 || strings.HasPrefix(f[0], "#") || strings.HasPrefix(f[0], ";") {
			continue
		}
		switch f[0] {
		case "nameserver":
			if a, err := netip.ParseAddr(f[1]); err == nil && len(c.servers) < maxNameservers {
				c.servers = append(c.servers, netip.AddrPortFrom(a, 53).String())
			}
		case "options":
			for _, o := range f[1:] {
				if o == "use-vc" || o == "usevc" || o == "tcp" {
					c.useVC = true
				}
			}
		}
	}
	if len(c.servers) == 0 {
		c.servers = defaultNameservers
	}
	return c
}

// cnameClient performs the CNAME edge lookup over an injectable transport (tests only; production uses the system
// dialer and /etc/resolv.conf).
type cnameClient struct {
	conf  func() resolvConf
	dial  func(ctx context.Context, network, address string) (net.Conn, error)
	bound time.Duration // total bound applied only when the caller's context has no deadline
}

// Internal exchange results. errServer and errForeign never leave the package: a failed server yields its remaining
// budget to the next one, and when none is left the lookup is ErrDNSTransient.
var (
	errServer    = errors.New("domains: DNS server failure")
	errForeign   = errors.New("domains: DNS message does not answer the query")
	errTruncated = errors.New("domains: DNS answer truncated")
)

// lookup returns the immediate CNAME edge of name as an absolute lowercase name, the name itself when it has no CNAME
// (NODATA), ErrNXDomain when the name does not exist, or ErrDNSTransient.
func (c cnameClient) lookup(ctx context.Context, name string) (string, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.bound)
		defer cancel()
	}
	owner := lowerASCII(name)
	if !strings.HasSuffix(owner, ".") {
		owner += "."
	}
	qname, err := dnsmessage.NewName(owner)
	if err != nil {
		return "", ErrDNSTransient
	}
	q := dnsmessage.Question{Name: qname, Type: dnsmessage.TypeCNAME, Class: dnsmessage.ClassINET}
	conf := c.conf()
	deadline, _ := ctx.Deadline()
	for i, server := range conf.servers {
		if ctx.Err() != nil {
			return "", ErrDNSTransient
		}
		// Each server gets an equal share of what is left, so a mute first server cannot consume the fallback's time.
		sctx, cancel := context.WithTimeout(ctx, time.Until(deadline)/time.Duration(len(conf.servers)-i))
		target, err := c.ask(sctx, server, conf.useVC, q)
		cancel()
		if !errors.Is(err, errServer) {
			return target, err
		}
	}
	return "", ErrDNSTransient
}

// ask queries one server: UDP first (unless use-vc), the same query over TCP when the UDP answer is truncated. UDP and
// its TCP retry share the server's budget.
func (c cnameClient) ask(ctx context.Context, server string, useVC bool, q dnsmessage.Question) (string, error) {
	if !useVC {
		p, h, err := c.exchange(ctx, "udp", server, q)
		if err == nil {
			return edge(&p, h, q)
		}
		if !errors.Is(err, errTruncated) {
			return "", err
		}
		if ctx.Err() != nil {
			return "", errServer
		}
	}
	p, h, err := c.exchange(ctx, "tcp", server, q)
	if err != nil {
		return "", errServer // including a truncated answer over TCP
	}
	return edge(&p, h, q)
}

// exchange sends one query and returns the parser positioned after the validated question of the matching answer.
func (c cnameClient) exchange(ctx context.Context, network, server string, q dnsmessage.Question) (dnsmessage.Parser, dnsmessage.Header, error) {
	id := newQueryID()
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{ID: id, RecursionDesired: true})
	if err := b.StartQuestions(); err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, ErrDNSTransient
	}
	if err := b.Question(q); err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, ErrDNSTransient // an unpackable name: not a server failure
	}
	query, err := b.Finish()
	if err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, ErrDNSTransient
	}
	if ctx.Err() != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errServer
	}
	conn, err := c.dial(ctx, network, server)
	if err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errServer
	}
	release := closeOnDone(ctx, conn)
	defer func() {
		release()
		_ = conn.Close()
	}()
	if d, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(d)
	}
	if _, ok := conn.(net.PacketConn); ok {
		return udpRoundTrip(conn, query, id, q)
	}
	return streamRoundTrip(conn, query, id, q)
}

// closeOnDone closes conn as soon as ctx is done, which unblocks a read or write in progress. The returned release
// unregisters the callback; if the callback has already started, release waits until it has finished (the stop
// function of context.AfterFunc does not wait for a running callback), so the callback never outlives release.
func closeOnDone(ctx context.Context, conn net.Conn) (release func()) {
	done := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(done)
		_ = conn.Close()
	})
	return func() {
		if !stop() {
			<-done
		}
	}
}

// udpRoundTrip reads datagrams until one answers the query. Datagrams with another ID, no response bit, another
// opcode or another question are discarded (stale or spoofed traffic); a matching datagram with a malformed question
// section is a server failure. Truncation (TC, or a datagram filling the buffer) is only acted on after that
// validation.
func udpRoundTrip(conn net.Conn, query []byte, id uint16, q dnsmessage.Question) (dnsmessage.Parser, dnsmessage.Header, error) {
	if _, err := conn.Write(query); err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errServer
	}
	buf := make([]byte, maxDNSMessage)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return dnsmessage.Parser{}, dnsmessage.Header{}, errServer
		}
		p, h, err := envelope(buf[:n], id, q)
		if errors.Is(err, errForeign) {
			continue
		}
		if err != nil {
			return dnsmessage.Parser{}, dnsmessage.Header{}, err
		}
		if h.Truncated || n == len(buf) {
			return dnsmessage.Parser{}, dnsmessage.Header{}, errTruncated
		}
		if !recordsFramed(buf[:n]) {
			return dnsmessage.Parser{}, dnsmessage.Header{}, errServer
		}
		return p, h, nil
	}
}

// streamRoundTrip sends the length-framed query and reads exactly one length-framed answer of 12..4096 bytes.
func streamRoundTrip(conn net.Conn, query []byte, id uint16, q dnsmessage.Question) (dnsmessage.Parser, dnsmessage.Header, error) {
	framed := binary.BigEndian.AppendUint16(make([]byte, 0, 2+len(query)), uint16(len(query)))
	if _, err := conn.Write(append(framed, query...)); err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errServer
	}
	var l [2]byte
	if _, err := io.ReadFull(conn, l[:]); err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errServer
	}
	n := int(binary.BigEndian.Uint16(l[:]))
	if n < dnsHeaderLen || n > maxDNSMessage {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errServer
	}
	msg := make([]byte, n)
	if _, err := io.ReadFull(conn, msg); err != nil {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errServer
	}
	p, h, err := envelope(msg, id, q)
	if err != nil || h.Truncated || !recordsFramed(msg) {
		return dnsmessage.Parser{}, dnsmessage.Header{}, errServer // a stream carries one answer: no discarding
	}
	return p, h, nil
}

// recordsFramed checks record framing through the pinned parser's public record API: every record of every section
// lies inside msg, and every answer CNAME's encoded name occupies exactly its declared RDLENGTH. The parser's body path
// (AllAnswers) advances by RDLENGTH without comparing it to the message length or to the bytes the CNAME name actually
// consumed (it decodes the name from the whole message), so an oversized, undersized or padded CNAME body would
// otherwise supply an edge. Together with edge's complete body parse, an answer is used only when both passes succeed.
func recordsFramed(msg []byte) bool {
	var p dnsmessage.Parser
	if _, err := p.Start(msg); err != nil {
		return false
	}
	if err := p.SkipAllQuestions(); err != nil {
		return false
	}
	for {
		h, err := p.AnswerHeader()
		if errors.Is(err, dnsmessage.ErrSectionDone) {
			break
		}
		if err != nil {
			return false
		}
		if h.Type != dnsmessage.TypeCNAME {
			if err := p.SkipAnswer(); err != nil {
				return false
			}
			continue
		}
		raw, err := p.UnknownResource() // exactly the declared RDATA, bounds-checked by the parser
		if err != nil || nameExtent(raw.Data) != len(raw.Data) {
			return false
		}
	}
	if err := p.SkipAllAuthorities(); err != nil {
		return false
	}
	return p.SkipAllAdditionals() == nil
}

// nameExtent returns how many leading bytes of b an encoded name occupies at its own position: its labels through
// the root terminator, or through a two-byte compression pointer (which ends the local encoding wherever it points),
// or -1 when b ends first or a reserved label type appears. It neither decodes nor follows the name; dnsmessage does.
func nameExtent(b []byte) int {
	for i := 0; i < len(b); {
		switch l := int(b[i]); {
		case l == 0:
			return i + 1
		case l&0xc0 == 0xc0:
			if i+2 > len(b) {
				return -1
			}
			return i + 2
		case l&0xc0 != 0:
			return -1
		default:
			i += 1 + l
		}
	}
	return -1
}

// envelope validates the header and the single question echo of msg against the query.
func envelope(msg []byte, id uint16, q dnsmessage.Question) (dnsmessage.Parser, dnsmessage.Header, error) {
	var p dnsmessage.Parser
	h, err := p.Start(msg)
	if err != nil || h.ID != id || !h.Response || h.OpCode != 0 {
		return p, h, errForeign
	}
	got, err := p.Question()
	if err != nil {
		return p, h, errServer // no or a malformed question in an answer bearing our ID
	}
	if got.Type != q.Type || got.Class != q.Class || lowerASCII(got.Name.String()) != q.Name.String() {
		return p, h, errForeign
	}
	if _, err := p.Question(); !errors.Is(err, dnsmessage.ErrSectionDone) {
		return p, h, errServer // more than one question, or malformed
	}
	return p, h, nil
}

// edge parses every remaining section of the validated answer, then selects the owner's immediate edge: the unique
// IN CNAME record whose owner equals the queried owner, independent of record order. Records of other owners, types or
// classes are never used as the edge.
func edge(p *dnsmessage.Parser, h dnsmessage.Header, q dnsmessage.Question) (string, error) {
	answers, err := p.AllAnswers()
	if err != nil {
		return "", errServer
	}
	if _, err := p.AllAuthorities(); err != nil {
		return "", errServer
	}
	if _, err := p.AllAdditionals(); err != nil {
		return "", errServer
	}
	switch h.RCode {
	case dnsmessage.RCodeSuccess:
	case dnsmessage.RCodeNameError:
		return "", ErrNXDomain
	default:
		return "", errServer
	}
	owner := q.Name.String()
	target, n := "", 0
	for _, rr := range answers {
		if rr.Header.Type != dnsmessage.TypeCNAME || rr.Header.Class != dnsmessage.ClassINET ||
			lowerASCII(rr.Header.Name.String()) != owner {
			continue
		}
		c, ok := rr.Body.(*dnsmessage.CNAMEResource)
		if !ok {
			return "", errServer
		}
		target, n = lowerASCII(c.CNAME.String()), n+1
	}
	switch {
	case n == 0:
		return owner, nil // NODATA: the chain ends at the queried name
	case n > 1:
		return "", ErrDNSTransient // ambiguous (RFC 2181 §10.1): never an edge
	case !hostnameShaped(target):
		return "", ErrDNSTransient
	}
	return target, nil
}

// hostnameShaped reports whether an absolute lowercase name is a hostname: labels of 1–63 [a-z0-9_-], ≤253 characters.
func hostnameShaped(fqdn string) bool {
	s, ok := strings.CutSuffix(fqdn, ".")
	if !ok || s == "" || len(s) > 253 {
		return false
	}
	for _, l := range strings.Split(s, ".") {
		if l == "" || len(l) > 63 {
			return false
		}
		for i := 0; i < len(l); i++ {
			if c := l[i]; !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
				return false
			}
		}
	}
	return true
}

// lowerASCII lowercases ASCII letters only (DNS names compare case-insensitively in ASCII only).
func lowerASCII(s string) string {
	b := []byte(s)
	for i, c := range b {
		if 'A' <= c && c <= 'Z' {
			b[i] = c + 'a' - 'A'
		}
	}
	return string(b)
}

func newQueryID() uint16 {
	var b [2]byte
	_, _ = rand.Read(b[:]) // crypto/rand.Read does not fail
	return binary.BigEndian.Uint16(b[:])
}
