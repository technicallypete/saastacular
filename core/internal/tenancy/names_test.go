package tenancy

import (
	"strings"
	"testing"
)

// U-NAME: valid UTF-8, Unicode whitespace trimmed, no C0/DEL/C1 control character, 1-200 code points.
func TestValidateName(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"Acme Revenue", "Acme Revenue"},
		{"  Acme Revenue \t\n", "Acme Revenue"},
		{"　Acme ", "Acme"},
		{"a", "a"},
		{strings.Repeat("é", 200), strings.Repeat("é", 200)}, // 200 code points, 400 bytes
		{"東京チーム", "東京チーム"},
		{"Café Ops", "Café Ops"},
		{"../Acme%2FRevenue", "../Acme%2FRevenue"},
	} {
		got, err := ValidateName(c.in)
		if err != nil || got != c.want {
			t.Errorf("ValidateName(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	for _, in := range []string{
		"", " ", "\t\n", "　", strings.Repeat("a", 201), strings.Repeat("é", 201),
		"a\x00b", "a\tb", "a\nb", "a\x7fb", "a\u0085b", "a\u009fb", "a\u0080b", "\xff", "a\xc3",
	} {
		if _, err := ValidateName(in); CodeOf(err) != CodeInvalidName {
			t.Errorf("ValidateName(%q): %v, want invalid_name", in, err)
		}
	}
}
