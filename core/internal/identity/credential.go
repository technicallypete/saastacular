package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// secretLen is the size of every bearer or binding value: session credentials, CSRF tokens, login bindings,
// state, PKCE verifiers and logout references all carry 256 bits from crypto/rand.
const secretLen = 32

// encodedLen is the base64url (no padding) length of a secretLen value.
var encodedLen = base64.RawURLEncoding.EncodedLen(secretLen)

// newSecret returns secretLen random bytes.
func newSecret() ([]byte, error) {
	b := make([]byte, secretLen)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("identity: randomness unavailable: %w", err)
	}
	return b, nil
}

// encodeSecret renders a secret for transport (cookie, header, URL parameter, PKCE verifier).
func encodeSecret(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// decodeSecret accepts exactly one canonical base64url encoding of exactly secretLen bytes.
func decodeSecret(s string) ([]byte, bool) {
	if len(s) != encodedLen {
		return nil, false
	}
	b, err := base64.RawURLEncoding.Strict().DecodeString(s)
	if err != nil || len(b) != secretLen || encodeSecret(b) != s {
		return nil, false
	}
	return b, true
}

// hashSecret is the stored one-way representation of a bearer or binding value.
func hashSecret(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}

// hashEncoded hashes the transport form (used for state, which travels as the encoded string).
func hashEncoded(s string) []byte {
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}

// equalSecret compares two values in constant time.
func equalSecret(a, b []byte) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare(a, b) == 1
}

// s256Challenge is the PKCE S256 code challenge for a verifier: BASE64URL(SHA-256(ASCII(verifier))).
func s256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
