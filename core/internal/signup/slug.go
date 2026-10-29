package signup

import (
	"strconv"

	"core/internal/hostname"
)

// The advisory slug suggestion (G062 Plan r5 §4.3): a pinned copy of tenancy's deterministic base and candidate rule
// (TestSlugRuleCopyIsExact compares suggestBase, cut, slugCandidates, MaxSlugLength and maxSuffix with
// core/internal/tenancy/slug.go byte for byte). A suggestion never reserves anything; the provisioning body's unique
// index decides.

// MaxSlugLength is the bound of a stored Group slug, in ASCII bytes.
const MaxSlugLength = 63

// maxSuffix bounds the advisory suggestion search: base, then base-2 through base-99.
const maxSuffix = 99

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

// ValidSlug reports whether slug may name a new platform tenant host under cfg: the Organization slug grammar and not a
// reserved first label (the floor, the CNAME-target and auth-host labels under the platform domain, and the configured
// PLATFORM_RESERVED_LABELS). A nested name or the platform domain itself never matches the grammar.
func ValidSlug(slug string, cfg hostname.Config) bool {
	return hostname.SlugGrammar(slug) && !hostname.IsReservedLabel(slug, cfg)
}
