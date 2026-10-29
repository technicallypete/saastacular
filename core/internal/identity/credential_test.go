package identity

import (
	"bytes"
	"strings"
	"testing"
)

func TestSecretsAre256BitAndDistinct(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		b, err := newSecret()
		if err != nil {
			t.Fatal(err)
		}
		if len(b) != 32 {
			t.Fatalf("secret length %d", len(b))
		}
		s := encodeSecret(b)
		if len(s) != 43 || seen[s] {
			t.Fatalf("encoding %q (len %d) or repeat", s, len(s))
		}
		seen[s] = true
		back, ok := decodeSecret(s)
		if !ok || !bytes.Equal(back, b) {
			t.Fatal("round trip failed")
		}
	}
}

func TestDecodeSecretIsStrict(t *testing.T) {
	b, _ := newSecret()
	good := encodeSecret(b)
	for _, bad := range []string{
		"", "short", good + "A", good[:42], good + "=",
		strings.Replace(good, good[:1], "+", 1), strings.Replace(good, good[:1], "/", 1), good[:42] + " ",
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAB", // non-canonical trailing bits
	} {
		if bad == good {
			continue
		}
		if _, ok := decodeSecret(bad); ok {
			t.Errorf("accepted %q", bad)
		}
	}
}

// RFC 7636 appendix B vector.
func TestS256Challenge(t *testing.T) {
	if got := s256Challenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"); got != "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM" {
		t.Fatalf("challenge %q", got)
	}
}

func TestEqualSecret(t *testing.T) {
	a, _ := newSecret()
	b := append([]byte(nil), a...)
	if !equalSecret(a, b) || equalSecret(a, b[:31]) || equalSecret(a, make([]byte, 32)) {
		t.Fatal("equalSecret")
	}
	if !bytes.Equal(hashSecret(a), hashSecret(b)) || len(hashSecret(a)) != 32 || len(hashEncoded("x")) != 32 {
		t.Fatal("hash")
	}
}
