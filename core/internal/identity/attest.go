package identity

import (
	"context"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// VerifiedAddress is identity-owned evidence that the canonical user mapped to an exact provider principal
// (issuer, subject) controls an email address, as attested by a reviewed provider adapter (G019 Plan r2 addendum §2;
// the G058 contract). Its fields are unexported: no JSON decoding, struct literal or constructor outside this package
// can produce a non-zero value, and the user ID is the result of the canonical principal resolution, never a caller
// argument. core/internal/tenancy/access consumes it for Organization invitation consumption (D-15 pending: no
// production caller in G019).
type VerifiedAddress struct {
	userID     string
	issuer     string
	subject    string
	address    string
	verifiedAt time.Time
	source     string
}

// UserID is the canonical user mapped to the principal.
func (v VerifiedAddress) UserID() string { return v.userID }

// Issuer is the provider principal's issuer.
func (v VerifiedAddress) Issuer() string { return v.issuer }

// Subject is the provider principal's subject.
func (v VerifiedAddress) Subject() string { return v.subject }

// Address is the verified address, normalized exactly as tenancy.NormalizeAddress normalizes invitation addresses.
func (v VerifiedAddress) Address() string { return v.address }

// IsZero reports whether v is the zero value (no evidence). Consumers refuse it before any statement.
func (v VerifiedAddress) IsZero() bool { return v == VerifiedAddress{} }

// AttestVerifiedAddress builds a VerifiedAddress for the exact provider principal (issuer, subject) and an address the
// provider verified at verifiedAt, recording source (the adapter's evidence kind). It resolves the canonical user
// itself through the E2 principal reader (app.auth_principal_subject); an unknown principal is unauthenticated. An
// empty issuer, subject or source, a zero verifiedAt or an address that does not normalize is invalid_request; a store
// failure is identity_unavailable. Production call sites are pinned to core/internal/identity (TestAccessBoundary);
// G058's reviewed adapter adds the first, after obtaining current provider evidence for the exact session subject.
func (s *Service) AttestVerifiedAddress(ctx context.Context, issuer, subject, address string, verifiedAt time.Time, source string) (VerifiedAddress, error) {
	if issuer == "" || subject == "" || source == "" || verifiedAt.IsZero() {
		return VerifiedAddress{}, ErrInvalidRequest
	}
	addr, ok := normalizeAddress(address)
	if !ok {
		return VerifiedAddress{}, ErrInvalidRequest
	}
	if s == nil || s.store == nil {
		return VerifiedAddress{}, ErrUnavailable
	}
	mapped, err := principalSubject(ctx, s.store.db, issuer, subject)
	if err != nil {
		return VerifiedAddress{}, ErrUnavailable
	}
	if mapped == "" {
		return VerifiedAddress{}, ErrUnauthenticated
	}
	return VerifiedAddress{userID: mapped, issuer: issuer, subject: subject, address: addr, verifiedAt: verifiedAt.UTC(), source: source}, nil
}

// Address bounds of tenancy's invitations_address_normalized (RFC 5321: 64-octet local part, 254 octets in total).
const (
	attestMaxAddressBytes   = 254
	attestMaxLocalPartBytes = 64
)

// normalizeAddress is tenancy.NormalizeAddress re-implemented locally so that identity does not import tenancy (the
// cycle-free contract, r2 §2); TestVerifiedAddress pins the two equal on a shared table. Trim ASCII whitespace; exactly
// one '@' with a non-empty local part and domain; no whitespace, control character or invalid UTF-8; the local part
// byte-exact and at most 64 bytes; the domain NFC-normalized then lowercased; at most 254 bytes in total.
func normalizeAddress(s string) (string, bool) {
	s = strings.Trim(s, " \t\r\n\v\f")
	if s == "" || !utf8.ValidString(s) {
		return "", false
	}
	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return "", false
		}
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 || strings.IndexByte(s[at+1:], '@') >= 0 {
		return "", false
	}
	local := s[:at]
	domain := strings.ToLower(norm.NFC.String(s[at+1:]))
	out := local + "@" + domain
	if len(local) > attestMaxLocalPartBytes || len(out) > attestMaxAddressBytes || strings.IndexByte(domain, '@') >= 0 {
		return "", false
	}
	return out, true
}
