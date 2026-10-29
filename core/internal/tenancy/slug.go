package tenancy

import (
	"slices"
	"strconv"
)

// MaxSlugLength is the bound of a stored Group slug, in ASCII bytes.
const MaxSlugLength = 63

// maxSuffix bounds the advisory suggestion search: base, then base-2 through base-99.
const maxSuffix = 99

// reservedSlugs are top-level application and system route segments that no Group slug may take (pinned identically in
// the groups_slug_reserved CHECK of 00003_tenancy.sql; the parity test compares them).
var reservedSlugs = map[string]bool{
	"account": true, "admin": true, "api": true, "auth": true, "billing": true, "blueprint": true,
	"fonts": true, "health": true, "healthz": true, "internal": true, "login": true, "logout": true,
	"mcp": true, "members": true, "message": true, "messages": true, "new": true, "onboarding": true,
	"readyz": true, "rpc": true, "settings": true, "sign-in": true, "sse": true, "system": true,
}

// ReservedSlugs returns the reserved set, sorted.
func ReservedSlugs() []string {
	out := make([]string, 0, len(reservedSlugs))
	for s := range reservedSlugs {
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}

// ValidateSlug checks a stored or explicit slug: 1-63 ASCII bytes matching [a-z0-9]+(-[a-z0-9]+)* (invalid_slug) and
// not reserved (slug_reserved). Explicit values are validated, never normalized: no case folding, percent-decoding,
// Unicode folding or transliteration happens anywhere.
func ValidateSlug(s string) error {
	if !canonicalSlug(s) {
		return newError(CodeInvalidSlug)
	}
	if reservedSlugs[s] {
		return newError(CodeSlugReserved)
	}
	return nil
}

func canonicalSlug(s string) bool {
	if len(s) < 1 || len(s) > MaxSlugLength || s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '-':
			if s[i-1] == '-' {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// suggestBase derives the deterministic base suggestion from a valid name: ASCII letters are lowercased, ASCII letters
// and digits are kept, every other rune (including non-ASCII, '_', '.', '/', '%' and whitespace) is a separator, runs
// of separators collapse to one '-', leading and trailing '-' are trimmed, and the result is cut to 63 bytes with a
// trailing '-' trimmed. An empty result means the name has no base and an explicit slug is required.
func suggestBase(name string) string {
	out := make([]byte, 0, len(name))
	sep := false
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			r += 'a' - 'A'
			fallthrough
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			if sep && len(out) > 0 {
				out = append(out, '-')
			}
			sep = false
			out = append(out, byte(r))
		default:
			sep = true
		}
	}
	return cut(string(out), MaxSlugLength)
}

// cut shortens s to at most n bytes and trims trailing hyphens.
func cut(s string, n int) string {
	if len(s) > n {
		s = s[:n]
	}
	for len(s) > 0 && s[len(s)-1] == '-' {
		s = s[:len(s)-1]
	}
	return s
}

// slugCandidates returns the advisory candidates for a base: base, then base-2 through base-99, each cut so that the
// whole candidate is at most 63 bytes.
func slugCandidates(base string) []string {
	out := []string{base}
	for n := 2; n <= maxSuffix; n++ {
		suffix := "-" + strconv.Itoa(n)
		b := cut(base, MaxSlugLength-len(suffix))
		if b == "" {
			continue
		}
		out = append(out, b+suffix)
	}
	return out
}
