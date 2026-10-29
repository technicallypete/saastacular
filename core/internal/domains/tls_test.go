package domains

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"testing"
	"time"
)

// TestTLSAdapterContract: the edge-probe against a disposable local TLS listener (pending, ready, wrong host, untrusted,
// expired, timeout -> unknown, idempotent repeat) and the scripted fixture adapter's generation binding.
func TestTLSAdapterContract(t *testing.T) {
	now := time.Now()
	ca, caKey := mustCA(t, "probe CA")
	other, otherKey := mustCA(t, "foreign CA")
	roots := x509.NewCertPool()
	roots.AddCert(ca)
	certs := map[string]tls.Certificate{
		"ready.example.com":     leaf(t, ca, caKey, "ready.example.com", now.Add(-time.Hour), now.Add(48*time.Hour)),
		"wronghost.example.com": leaf(t, ca, caKey, "someone-else.example.com", now.Add(-time.Hour), now.Add(48*time.Hour)),
		"untrusted.example.com": leaf(t, other, otherKey, "untrusted.example.com", now.Add(-time.Hour), now.Add(48*time.Hour)),
		"expired.example.com":   leaf(t, ca, caKey, "expired.example.com", now.Add(-72*time.Hour), now.Add(-time.Hour)),
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{GetCertificate: func(h *tls.ClientHelloInfo) (*tls.Certificate, error) {
		c, ok := certs[h.ServerName]
		if !ok {
			return nil, errors.New("no certificate for this name") // the edge refuses the handshake: pending
		}
		return &c, nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() { _ = c.(*tls.Conn).Handshake(); c.Close() }()
		}
	}()
	probe := EdgeProbe{Addr: ln.Addr().String(), Timeout: 2 * time.Second, Roots: roots}
	ctx := context.Background()
	for _, c := range []struct {
		host string
		want TLSState
	}{
		{"ready.example.com", TLSReady}, {"ready.example.com", TLSReady}, // idempotent repeat
		{"pending.example.com", TLSPending}, {"wronghost.example.com", TLSError}, {"untrusted.example.com", TLSError},
		{"expired.example.com", TLSExpired},
	} {
		ref := DomainRef{ID: "id-" + c.host, Generation: 2, Hostname: c.host}
		got := probe.Status(ctx, ref)
		if got.State != c.want || got.Ref != ref {
			t.Errorf("Status(%s) = %+v; want %s", c.host, got, c.want)
		}
		if c.want == TLSReady && !got.NotAfter.After(now) {
			t.Errorf("ready without a future not_after: %v", got.NotAfter)
		}
	}
	ref := DomainRef{ID: "x", Generation: 1, Hostname: "ready.example.com"}
	if r := probe.Request(ctx, ref); r.State != TLSRequested || r.Ref != ref {
		t.Errorf("E1 Request: %+v", r)
	}
	if r := probe.Remove(ctx, ref); r.State != TLSWithdrawn || r.Ref != ref {
		t.Errorf("E1 Remove: %+v", r)
	}

	// A listener that never answers: the deadline yields unknown, never a readiness claim.
	silent, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer silent.Close()
	go func() {
		for {
			c, err := silent.Accept()
			if err != nil {
				return
			}
			defer c.Close()
		}
	}()
	slow := EdgeProbe{Addr: silent.Addr().String(), Timeout: 100 * time.Millisecond, Roots: roots}
	if r := slow.Status(ctx, ref); r.State != TLSUnknown {
		t.Errorf("timeout: %+v", r)
	}
	closedAddr := silent.Addr().String()
	if r := (EdgeProbe{Addr: "127.0.0.1:1", Timeout: time.Second}).Status(ctx, ref); r.State != TLSUnknown {
		t.Errorf("refused connection: %+v (%s)", r, closedAddr)
	}

	// The fixture adapter: per-generation scripts, sequences, defaults and the call record.
	f, err := ParseFixtureAdapter([]byte("# tls fixture\nportal.example.com\t2\tstatus\tready\t2030-01-01T00:00:00Z\n" +
		"portal.example.com\t*\tremove\tunknown\nportal.example.com\t*\tremove\twithdrawn\n"))
	if err != nil {
		t.Fatal(err)
	}
	g1 := DomainRef{ID: "a", Generation: 1, Hostname: "portal.example.com"}
	g2 := DomainRef{ID: "b", Generation: 2, Hostname: "portal.example.com"}
	if r := f.Status(ctx, g1); r.State != TLSPending || r.Ref != g1 {
		t.Errorf("unscripted generation: %+v", r)
	}
	if r := f.Status(ctx, g2); r.State != TLSReady || r.Ref != g2 || r.NotAfter.Year() != 2030 {
		t.Errorf("scripted generation: %+v", r)
	}
	if a, b, c := f.Remove(ctx, g1), f.Remove(ctx, g1), f.Remove(ctx, g1); a.State != TLSUnknown || b.State != TLSWithdrawn || c.State != TLSWithdrawn {
		t.Errorf("remove sequence: %s %s %s", a.State, b.State, c.State)
	}
	if r := f.Request(ctx, g2); r.State != TLSRequested {
		t.Errorf("default request: %+v", r)
	}
	if n := len(f.Calls()); n != 6 {
		t.Errorf("calls: %v", f.Calls())
	}
	for _, bad := range []string{"h\t*\tstatus", "h\t0\tstatus\tready", "h\t1\tstatus\tready\tyesterday"} {
		if _, err := ParseFixtureAdapter([]byte(bad)); err == nil {
			t.Errorf("fixture %q accepted", bad)
		}
	}
}

func mustCA(t *testing.T, cn string) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: cn}, NotBefore: time.Now().Add(-time.Hour),
		NotAfter: time.Now().Add(240 * time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	c, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return c, key
}

func leaf(t *testing.T, ca *x509.Certificate, caKey *ecdsa.PrivateKey, name string, from, to time.Time) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{SerialNumber: big.NewInt(time.Now().UnixNano()), Subject: pkix.Name{CommonName: name},
		DNSNames: []string{name}, NotBefore: from, NotAfter: to, KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, tpl, ca, &key.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}
