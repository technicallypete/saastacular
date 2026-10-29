package tenancy

import (
	"strings"
	"unicode/utf8"
)

// MaxNameLength is the bound of Organization and Group display names, in Unicode code points.
const MaxNameLength = 200

// ValidateName returns the stored form of an Organization or Group display name: valid UTF-8, Unicode whitespace
// trimmed, no C0, DEL or C1 control character (U+0000-U+001F, U+007F, U+0080-U+009F) and 1-200 code points.
// Anything else is invalid_name. The database rule (length, ASCII-space trim, the same control characters) is never
// stricter than this rule.
func ValidateName(s string) (string, error) {
	if !utf8.ValidString(s) {
		return "", newError(CodeInvalidName)
	}
	t := strings.TrimSpace(s)
	n := 0
	for _, r := range t {
		if r <= 0x1f || (r >= 0x7f && r <= 0x9f) {
			return "", newError(CodeInvalidName)
		}
		n++
	}
	if n < 1 || n > MaxNameLength {
		return "", newError(CodeInvalidName)
	}
	return t, nil
}
