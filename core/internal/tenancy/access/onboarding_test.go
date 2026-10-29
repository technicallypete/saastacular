package access

import (
	"strings"
	"testing"

	"core/internal/tenancy"
)

// TestOnboardingPageBounds (Plan r5 §1.5, ONB-10 unit part): the list page input is the G014 keyset contract — limit
// 1..100 (invalid_page_size), an empty or canonical UUID cursor (invalid_cursor) — decided before any transaction; and
// the response-size arithmetic of a full page stays below the pinned client bounds.
func TestOnboardingPageBounds(t *testing.T) {
	if tenancy.MaxOnboardingPage != 100 {
		t.Fatalf("page bound %d, want 100", tenancy.MaxOnboardingPage)
	}
	const id = "0f18a000-0000-4000-8000-0000000000a1"
	for _, c := range []struct {
		after string
		limit int
		field string
	}{
		{"", 1, ""}, {"", 100, ""}, {id, 50, ""},
		{"", 0, "invalid_page_size"}, {"", 101, "invalid_page_size"}, {"", -1, "invalid_page_size"},
		{"x", 10, "invalid_cursor"}, {strings.ToUpper(id) + "0", 10, "invalid_cursor"}, {id[:35], 10, "invalid_cursor"},
	} {
		cur, f := pageInput(c.after, c.limit, tenancy.ParseGroupID)
		switch {
		case c.field == "" && (f != nil || (c.after != "" && cur.String() != c.after)):
			t.Errorf("%q/%d: %v %v", c.after, c.limit, cur, f)
		case c.field != "" && (f == nil || f.code != CodeInvalidInput || f.field != c.field || f.reason != reasonInputInvalid):
			t.Errorf("%q/%d: %+v, want invalid_input %s", c.after, c.limit, f, c.field)
		}
		if _, pf := pageInput(c.after, c.limit, tenancy.ParseBusinessProfileID); (pf == nil) != (c.field == "") {
			t.Errorf("profile cursor %q/%d: %v", c.after, c.limit, pf)
		}
	}
	// Response bounds (§1.5): a Group row at most 1,500 bytes and a profile summary at most 3,200 (Go's encoder: at most
	// 6 bytes per code point), so 100 rows plus the envelope stay below 160 KiB and 330 KiB.
	group := 2*36 + 6*tenancy.MaxNameLength + tenancy.MaxSlugLength + 160
	profile := 36 + 6*(tenancy.MaxDisplayNameLength+tenancy.MaxIndustryLength) + 60
	if group > 1500 || 100*group+1024 > 160<<10 || profile > 3200 || 100*profile+1024 > 330<<10 {
		t.Errorf("page bounds: group row %d, profile row %d", group, profile)
	}
}
