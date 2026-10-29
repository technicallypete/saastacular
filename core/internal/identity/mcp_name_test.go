package identity

import (
	"errors"
	"strings"
	"testing"
)

// G063 (Plan r1 §4.1, §6; D-6): the authoritative friendly-name rule on the shared vector table. The same vectors run
// against web/lib/mcp-client.ts normalizeDisplayNameInput (web/tests/unit/mcp-connections.test.mjs). This is a Go-string
// test of the normalizer only; the wire refusal of malformed UTF-8 and unpaired surrogate escapes is
// TestSetupNameDecoding (core/internal/mcp) and IT-MCP-15.

type nameVector struct {
	label string
	in    string
	want  string // the stored value; "" with clear true means nil
	clear bool
	bad   bool
}

func mcpNameVectors() []nameVector {
	emoji80 := strings.Repeat("\U0001F600", 80)
	return []nameVector{
		{label: "80 BMP", in: strings.Repeat("a", 80), want: strings.Repeat("a", 80)},
		{label: "80 non-BMP", in: emoji80, want: emoji80},
		{label: "padded non-BMP", in: strings.Repeat("\u3000", 8) + emoji80 + "\u00A0\u2000 ", want: emoji80},
		{label: "81 BMP", in: strings.Repeat("a", 81), bad: true},
		{label: "81 non-BMP", in: emoji80 + "\U0001F600", bad: true},
		{label: "empty", in: "", clear: true},
		{label: "spaces only", in: "    ", clear: true},
		{label: "Unicode whitespace only", in: "\u00A0\u3000 \u2000\u200A\u202F\u205F\u1680", clear: true},
		{label: "outer NBSP", in: "\u00A0Laptop\u00A0", want: "Laptop"},
		{label: "outer ideographic space", in: "\u3000Laptop\u3000", want: "Laptop"},
		{label: "outer en quad and hair space", in: "\u2000Laptop\u200A", want: "Laptop"},
		{label: "outer narrow NBSP and medium math space", in: "\u202FLaptop\u205F", want: "Laptop"},
		{label: "outer ogham space", in: "\u1680Laptop ", want: "Laptop"},
		{label: "inner spaces kept", in: "Work  laptop\u00A0two", want: "Work  laptop\u00A0two"},
		{label: "BOM kept", in: "\uFEFFx", want: "\uFEFFx"},
		{label: "BOM counted", in: "\uFEFF" + strings.Repeat("a", 80), bad: true},
		{label: "BOM not trimmed", in: " \uFEFF ", want: "\uFEFF"},
		{label: "markup kept", in: "<script>alert(1)</script>", want: "<script>alert(1)</script>"},
		{label: "duplicate-friendly plain", in: "Claude Desktop", want: "Claude Desktop"},
		{label: "inner newline", in: "a\nb", bad: true},
		{label: "only newline", in: "\n", bad: true},
		{label: "trailing newline", in: "Laptop\n", bad: true},
		{label: "leading tab", in: "\tLaptop", bad: true},
		{label: "leading vertical tab", in: "\vLaptop", bad: true},
		{label: "trailing form feed", in: "Laptop\f", bad: true},
		{label: "trailing carriage return", in: "Laptop\r", bad: true},
		{label: "NUL", in: "Lap\x00top", bad: true},
		{label: "leading NEL", in: "\u0085Laptop", bad: true},
		{label: "trailing line separator", in: "Laptop\u2028", bad: true},
		{label: "paragraph separator only", in: "\u2029", bad: true},
		{label: "DEL", in: "Laptop\x7f", bad: true},
		{label: "C1 APC", in: "\u009fLaptop", bad: true},
		{label: "invalid UTF-8", in: "Lap\xfftop", bad: true},
		{label: "encoded surrogate", in: "\xed\xa0\x80", bad: true},
	}
}

func TestNormalizeMCPDisplayName(t *testing.T) {
	for _, v := range mcpNameVectors() {
		got, err := NormalizeMCPDisplayName(v.in)
		switch {
		case v.bad:
			if !errors.Is(err, ErrInvalidMCPName) || got != nil {
				t.Errorf("%s: %v, %v; want ErrInvalidMCPName", v.label, got, err)
			}
		case v.clear:
			if err != nil || got != nil {
				t.Errorf("%s: %v, %v; want nil (clear)", v.label, got, err)
			}
		default:
			if err != nil || got == nil || *got != v.want {
				t.Errorf("%s: %v, %v; want %q", v.label, got, err, v.want)
			}
		}
		if err == nil && !validMCPName(got) {
			t.Errorf("%s: the normalized value is not itself valid", v.label)
		}
	}
	// validMCPName refuses a value that is not already normalized (the Tx helpers' defense in depth).
	for _, s := range []string{" x", "x ", "", "a\nb", strings.Repeat("a", 81)} {
		if validMCPName(&s) {
			t.Errorf("validMCPName(%q) = true", s)
		}
	}
	if !validMCPName(nil) {
		t.Error("validMCPName(nil) = false")
	}
}
