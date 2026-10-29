// Package branding holds the pure validation of Organization branding (G017 Plan r2 §1.3, §1.4.1): the display label,
// the colour-override grammar, the effective palette with its contrast oracle, and (image.go) the logo image checks
// and normalization. It has no I/O, no SQL and no dependency beyond the standard library; core/internal/tenancy/access
// calls it before any write.
package branding

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

// MaxLabelLength is the bound of the display label, in Unicode code points.
const MaxLabelLength = 80

// Error is a closed validation error. Field names the offending input (never its value).
type Error struct {
	Code  string // invalid_input or contrast_failed
	Field string
}

func (e *Error) Error() string { return "branding: " + e.Code + ": " + e.Field }

// The closed codes.
const (
	CodeInvalidInput   = "invalid_input"
	CodeContrastFailed = "contrast_failed"
)

func invalid(field string) error { return &Error{Code: CodeInvalidInput, Field: field} }

// AsError returns the closed validation error, if err is one.
func AsError(err error) (*Error, bool) {
	var e *Error
	ok := errors.As(err, &e)
	return e, ok
}

// Tokens are the 15 overridable semantic colour tokens: web/lib/tokens.ts SEMANTIC_COLOR_TOKENS without overlay (an
// rgba scrim, not a brand colour).
var Tokens = []string{
	"background", "foreground", "primary", "primary-foreground", "accent", "muted-foreground", "border", "popover",
	"popover-foreground", "muted", "secondary", "secondary-foreground", "destructive", "input", "ring",
}

var colorGrammar = regexp.MustCompile(`^#[0-9a-f]{6}$`)

// Overrides are one appearance's token overrides: token name -> lowercase #rrggbb.
type Overrides map[string]string

// ValidateLabel returns the stored label from an NFC-normalized input: nil for "no label" (empty after Unicode
// whitespace trimming), otherwise valid UTF-8 without C0, DEL or C1 control characters and 1-80 code points. The
// accepted contract is trim plus NFC normalization: the standard library has no normalizer, so the guarded operation
// normalizes with PostgreSQL's normalize(text, NFC) between PrepareLabel and this check (the length bound applies to
// the normalized label); the database constraint (IS NFC NORMALIZED) stays as defence in depth.
func ValidateLabel(s *string) (*string, error) {
	t, err := PrepareLabel(s)
	if err != nil || t == nil {
		return nil, err
	}
	if utf8.RuneCountInString(*t) > MaxLabelLength {
		return nil, invalid("display_label")
	}
	return t, nil
}

// PrepareLabel is the part of the label rule that precedes normalization: nil for "no label" (absent, or empty after
// Unicode whitespace trimming), otherwise the trimmed input, which must be valid UTF-8 without C0, DEL or C1 control
// characters (NFC normalization neither adds nor removes those). The guarded operation normalizes its result with
// PostgreSQL's built-in normalize(text, NFC) (tenancy.NormalizeLabelTx) and then applies ValidateLabel.
func PrepareLabel(s *string) (*string, error) {
	if s == nil {
		return nil, nil
	}
	if !utf8.ValidString(*s) {
		return nil, invalid("display_label")
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil, nil
	}
	for _, r := range t {
		if r <= 0x1f || (r >= 0x7f && r <= 0x9f) {
			return nil, invalid("display_label")
		}
	}
	return &t, nil
}

// ValidateOverrides returns the stored form of one appearance's overrides: known token names only, values lowercased
// and matching ^#[0-9a-f]{6}$. field is "light" or "dark".
func ValidateOverrides(field string, in map[string]string) (Overrides, error) {
	out := Overrides{}
	if len(in) > len(Tokens) {
		return nil, invalid(field)
	}
	for k, v := range in {
		if !isToken(k) {
			return nil, invalid(field)
		}
		v = strings.ToLower(v)
		if !colorGrammar.MatchString(v) {
			return nil, invalid(field)
		}
		out[k] = v
	}
	return out, nil
}

func isToken(k string) bool {
	for _, t := range Tokens {
		if t == k {
			return true
		}
	}
	return false
}

// ValidateColors checks both appearances' grammar and every pinned contrast pair of both effective palettes, returning
// the stored overrides. The label is checked separately (PrepareLabel, normalization, ValidateLabel).
func ValidateColors(light, dark map[string]string) (Overrides, Overrides, error) {
	lo, err := ValidateOverrides("light", light)
	if err != nil {
		return nil, nil, err
	}
	do, err := ValidateOverrides("dark", dark)
	if err != nil {
		return nil, nil, err
	}
	if err := CheckContrast(Light, lo); err != nil {
		return nil, nil, err
	}
	if err := CheckContrast(Dark, do); err != nil {
		return nil, nil, err
	}
	return lo, do, nil
}

// Validate checks a complete, already NFC-normalized branding configuration: the label, then ValidateColors. It
// returns the stored label and overrides.
func Validate(label *string, light, dark map[string]string) (*string, Overrides, Overrides, error) {
	l, err := ValidateLabel(label)
	if err != nil {
		return nil, nil, nil, err
	}
	lo, do, err := ValidateColors(light, dark)
	if err != nil {
		return nil, nil, nil, err
	}
	return l, lo, do, nil
}
