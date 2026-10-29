package tenancy

import (
	"strings"
	"testing"
)

// The pinned 24-entry reserved set (Plan r4 §4.4).
var pinnedReserved = []string{
	"account", "admin", "api", "auth", "billing", "blueprint", "fonts", "health", "healthz", "internal", "login", "logout",
	"mcp", "members", "message", "messages", "new", "onboarding", "readyz", "rpc", "settings", "sign-in", "sse", "system",
}

// slugGrammarNegatives are refused as stored or explicit slugs (IT-SLUG-2 uses the same list against the database).
var slugGrammarNegatives = []string{
	"", strings.Repeat("a", 64), "Acme", "acme_revenue", "acme.revenue", "-acme", "acme-", "acme--revenue",
	"acme%2frevenue", "%2e%2e", "..", "a/b", "a b", "é", "ａｃｍｅ", "асme", " acme", "acme\n",
}

// U-SLUG: the reserved set, the grammar and the deterministic base suggestion table.
func TestReservedSlugs(t *testing.T) {
	if got := strings.Join(ReservedSlugs(), ","); got != strings.Join(pinnedReserved, ",") {
		t.Fatalf("reserved set = %s", got)
	}
	for _, s := range pinnedReserved {
		if err := ValidateSlug(s); CodeOf(err) != CodeSlugReserved {
			t.Errorf("%s: %v, want slug_reserved", s, err)
		}
	}
}

func TestValidateSlug(t *testing.T) {
	for _, s := range []string{"a", "2026", "acme-revenue", "acme-revenue-2026", strings.Repeat("a", 63), "a-b-c", "api-2", "sign-in-2"} {
		if err := ValidateSlug(s); err != nil {
			t.Errorf("%q: %v", s, err)
		}
	}
	for _, s := range slugGrammarNegatives {
		if err := ValidateSlug(s); CodeOf(err) != CodeInvalidSlug {
			t.Errorf("%q: %v, want invalid_slug", s, err)
		}
	}
}

func TestSuggestBase(t *testing.T) {
	for _, c := range []struct{ name, want string }{
		{"Acme Revenue", "acme-revenue"},
		{"  Acme   Revenue  ", "acme-revenue"},
		{"ACME_Revenue.2026", "acme-revenue-2026"},
		{"-Acme--Revenue-", "acme-revenue"},
		{"../Acme%2FRevenue", "acme-2frevenue"},
		{"Café Ops", "caf-ops"},
		{"Ｒｅｖｅｎｕｅ", ""},
		{"東京チーム", ""},
		{"🚀", ""},
		{"API", "api"},
		{strings.Repeat("a", 70), strings.Repeat("a", 63)},
		{strings.Repeat("a", 62) + " b", strings.Repeat("a", 62)},
		{"2026", "2026"},
		{"a", "a"},
	} {
		if got := suggestBase(c.name); got != c.want {
			t.Errorf("suggestBase(%q) = %q, want %q", c.name, got, c.want)
		}
		if got := suggestBase(c.name); got != "" && !canonicalSlug(got) {
			t.Errorf("suggestBase(%q) = %q is not canonical", c.name, got)
		}
	}
}

func TestSlugCandidates(t *testing.T) {
	c := slugCandidates("acme")
	if len(c) != 99 || c[0] != "acme" || c[1] != "acme-2" || c[98] != "acme-99" {
		t.Fatalf("candidates: %d %v", len(c), c[:3])
	}
	long := slugCandidates(strings.Repeat("a", 63))
	if long[0] != strings.Repeat("a", 63) || long[1] != strings.Repeat("a", 61)+"-2" || long[98] != strings.Repeat("a", 60)+"-99" {
		t.Fatalf("long candidates: %q %q %q", long[0], long[1], long[98])
	}
	// A cut that would end in a hyphen is trimmed before the suffix.
	edge := slugCandidates(strings.Repeat("a", 60) + "-bc")
	if edge[1] != strings.Repeat("a", 60)+"-2" {
		t.Fatalf("edge candidate: %q", edge[1])
	}
	for _, list := range [][]string{c, long, edge} {
		for _, s := range list {
			if !canonicalSlug(s) {
				t.Errorf("candidate %q is not canonical", s)
			}
		}
	}
}
