package tenancy

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func ptr(s string) *string { return &s }

// U-CTX: the BusinessProfile context contract (Plan r2 §4.3): per-field codes, trimming, absence, control characters,
// code-point bounds and first-failure order.
func TestValidateBusinessContext(t *testing.T) {
	long := func(n int) string { return strings.Repeat("é", n) } // two bytes, one code point
	ok := BusinessContext{DisplayName: "Acme"}
	for _, c := range []struct {
		name string
		in   BusinessContext
		want Code
	}{
		{"display required", BusinessContext{}, CodeInvalidDisplayName},
		{"display whitespace only", BusinessContext{DisplayName: " \t  "}, CodeInvalidDisplayName},
		{"display NUL", BusinessContext{DisplayName: "a\x00b"}, CodeInvalidDisplayName},
		{"display TAB", BusinessContext{DisplayName: "a\tb"}, CodeInvalidDisplayName},
		{"display LF", BusinessContext{DisplayName: "a\nb"}, CodeInvalidDisplayName},
		{"display DEL", BusinessContext{DisplayName: "a\x7fb"}, CodeInvalidDisplayName},
		{"display C1", BusinessContext{DisplayName: "a\u009fb"}, CodeInvalidDisplayName},
		{"display NEL inside", BusinessContext{DisplayName: "a\u0085b"}, CodeInvalidDisplayName},
		{"display invalid UTF-8", BusinessContext{DisplayName: "a\xffb"}, CodeInvalidDisplayName},
		{"display 256", BusinessContext{DisplayName: long(256)}, CodeInvalidDisplayName},
		{"legal 256", BusinessContext{DisplayName: "A", LegalName: ptr(long(256))}, CodeInvalidLegalName},
		{"legal CR", BusinessContext{DisplayName: "A", LegalName: ptr("a\rb")}, CodeInvalidLegalName},
		{"legal invalid UTF-8", BusinessContext{DisplayName: "A", LegalName: ptr("\xc3")}, CodeInvalidLegalName},
		{"description 4001", BusinessContext{DisplayName: "A", Description: ptr(long(4001))}, CodeInvalidDescription},
		{"description CR", BusinessContext{DisplayName: "A", Description: ptr("a\r\nb")}, CodeInvalidDescription},
		{"description VT", BusinessContext{DisplayName: "A", Description: ptr("a\vb")}, CodeInvalidDescription},
		{"industry 256", BusinessContext{DisplayName: "A", Industry: ptr(long(256))}, CodeInvalidIndustry},
		{"industry LF", BusinessContext{DisplayName: "A", Industry: ptr("a\nb")}, CodeInvalidIndustry},
		{"website 2049", BusinessContext{DisplayName: "A", Website: ptr("https://e.com/" + strings.Repeat("a", 2035))}, CodeInvalidWebsite},
		{"website scheme", BusinessContext{DisplayName: "A", Website: ptr("ftp://example.com")}, CodeInvalidWebsite},
		// First failure in table order: display name before website.
		{"order", BusinessContext{DisplayName: "", Website: ptr("ftp://x")}, CodeInvalidDisplayName},
		{"order legal before description", BusinessContext{DisplayName: "A", LegalName: ptr("a\x00"), Description: ptr("b\x00")}, CodeInvalidLegalName},
	} {
		if _, err := ValidateBusinessContext(c.in); CodeOf(err) != c.want {
			t.Errorf("%s: %v, want %s", c.name, err, c.want)
		}
	}
	// Valid values: trimmed at both ends only; absent optional fields; bounds; TAB and LF inside the description.
	got, err := ValidateBusinessContext(BusinessContext{
		DisplayName: "  Acme  Holdings ", LegalName: ptr(""), Description: ptr("\n line one\n\tline two \n"),
		Industry: ptr("   "), Website: ptr(" https://example.com/x "),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.DisplayName != "Acme  Holdings" || got.LegalName != nil || got.Industry != nil || got.Description == nil ||
		*got.Description != "line one\n\tline two" || got.Website == nil || *got.Website != "https://example.com/x" {
		t.Fatalf("stored form: %+v", got)
	}
	for _, n := range []struct {
		field string
		c     BusinessContext
	}{
		{"display 255", BusinessContext{DisplayName: long(255)}},
		{"legal 255", BusinessContext{DisplayName: "A", LegalName: ptr(long(255))}},
		{"description 4000", BusinessContext{DisplayName: "A", Description: ptr(long(4000))}},
		{"industry 255", BusinessContext{DisplayName: "A", Industry: ptr(long(255))}},
		{"display 1", BusinessContext{DisplayName: "x"}},
		{"format characters are data", BusinessContext{DisplayName: "a\u202eb\u200d"}},
	} {
		if _, err := ValidateBusinessContext(n.c); err != nil {
			t.Errorf("%s: %v", n.field, err)
		}
	}
	if _, err := ValidateBusinessContext(ok); err != nil {
		t.Fatal(err)
	}
}

// U-URL: the website contract (Plan r2 §4.4; the r1 §5.2 lists).
func TestValidateWebsite(t *testing.T) {
	valid := []string{
		"https://example.com", "http://example.com:8080/p?q=1#f", "https://xn--bcher-kva.example/", "https://192.0.2.1/",
		"https://[2001:db8::1]/", "https://example.com/" + strings.Repeat("a", 2048-len("https://example.com/")),
		"https://example.com.", "http://a-b.example:1/", "https://example.com:65535",
	}
	for _, w := range valid {
		if err := ValidateWebsite(w); err != nil {
			t.Errorf("valid %q: %v", h12(w), err)
		}
		got, err := ValidateBusinessContext(BusinessContext{DisplayName: "A", Website: &w})
		if err != nil || got.Website == nil || *got.Website != w {
			t.Errorf("valid %q not stored unchanged: %v", h12(w), err)
		}
	}
	invalid := []string{
		"javascript:alert(1)", "data:text/html,x", "file:///etc/passwd", "ftp://example.com", "//example.com", "example.com",
		"HTTPS://example.com", "http:example.com", "https://", "https://user:pass@example.com", "https://user@example.com",
		"https://@example.com", "https://example.com\\@evil.test", "https://exa mple.com", "https://bücher.example",
		"https://exa_mple.com", "https://example.com:0/", "https://example.com:65536/", "https://[fe80::1%25eth0]/",
		"https://example.com/" + strings.Repeat("a", 2049-len("https://example.com/")), "https://-example.com/",
		"https://example-.com/", "https://example..com/", "https://example.com:/", "https://example.com:08080/",
		"https://[192.0.2.1]/", "https://" + strings.Repeat("a", 64) + ".com/",
	}
	for _, w := range invalid {
		if err := ValidateWebsite(w); CodeOf(err) != CodeInvalidWebsite {
			t.Errorf("invalid %q: %v", h12(w), err)
		}
	}
}

func h12(s string) string {
	if len(s) > 40 {
		return s[:40] + "..."
	}
	return s
}

// noDB fails the test if any operation reaches the database.
type noDB struct{ t *testing.T }

func (d noDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	d.t.Error("an invalid call reached the database")
	return nil, errors.New("no database")
}

// U-CTX (pre-SQL): every one of the nine G006 methods refuses zero IDs and invalid context before any transaction.
func TestBusinessProfileRejectionsBeforeSQL(t *testing.T) {
	svc := NewService(noDB{t}, nil)
	ctx := context.Background()
	var org OrganizationID
	var g GroupID
	var p BusinessProfileID
	var u UserID
	org1, g1, p1, u1 := OrganizationID{1}, GroupID{1}, BusinessProfileID{1}, UserID{1}
	valid := BusinessContext{DisplayName: "A"}
	for name, c := range map[string]struct {
		call func() error
		want Code
	}{
		"create zero org":    {func() error { _, err := svc.CreateBusinessProfile(ctx, org, valid); return err }, CodeInvalidID},
		"create bad context": {func() error { _, err := svc.CreateBusinessProfile(ctx, org1, BusinessContext{}); return err }, CodeInvalidDisplayName},
		"get zero profile":   {func() error { _, err := svc.GetBusinessProfile(ctx, org1, p); return err }, CodeInvalidID},
		"get zero org":       {func() error { _, err := svc.GetBusinessProfile(ctx, org, p1); return err }, CodeInvalidID},
		"update zero":        {func() error { _, err := svc.UpdateBusinessProfile(ctx, org1, p, valid); return err }, CodeInvalidID},
		"update bad website": {func() error {
			_, err := svc.UpdateBusinessProfile(ctx, org1, p1, BusinessContext{DisplayName: "A", Website: ptr("ftp://x")})
			return err
		}, CodeInvalidWebsite},
		"link zero group":   {func() error { _, err := svc.LinkGroupBusinessProfile(ctx, org1, g, p1); return err }, CodeInvalidID},
		"link zero profile": {func() error { _, err := svc.LinkGroupBusinessProfile(ctx, org1, g1, p); return err }, CodeInvalidID},
		"get association":   {func() error { _, _, err := svc.GetGroupBusinessProfile(ctx, org1, g); return err }, CodeInvalidID},
		"list groups":       {func() error { _, err := svc.ListGroupsForBusinessProfile(ctx, org1, p); return err }, CodeInvalidID},
		"unlink zero":       {func() error { return svc.UnlinkGroupBusinessProfile(ctx, org, g1, p1) }, CodeInvalidID},
		"profile paths":     {func() error { _, err := svc.ListBusinessProfileAccessPaths(ctx, org1, p); return err }, CodeInvalidID},
		"user paths":        {func() error { _, err := svc.ListUserBusinessProfileAccessPaths(ctx, org1, u); return err }, CodeInvalidID},
		"user paths org":    {func() error { _, err := svc.ListUserBusinessProfileAccessPaths(ctx, org, u1); return err }, CodeInvalidID},
	} {
		if got := CodeOf(c.call()); got != c.want {
			t.Errorf("%s: %s, want %s", name, got, c.want)
		}
	}
	for _, s := range []string{"", "00000000-0000-0000-0000-000000000000", "not-a-uuid"} {
		if _, err := ParseBusinessProfileID(s); CodeOf(err) != CodeInvalidID {
			t.Errorf("ParseBusinessProfileID(%q): %v", s, err)
		}
		if _, err := ParseAssociationID(s); CodeOf(err) != CodeInvalidID {
			t.Errorf("ParseAssociationID(%q): %v", s, err)
		}
	}
	id, err := ParseBusinessProfileID("0F8FAD5B-D9CB-469F-A165-70867728950E")
	if err != nil || id.String() != "0f8fad5b-d9cb-469f-a165-70867728950e" {
		t.Fatalf("ParseBusinessProfileID round trip: %v %s", err, id)
	}
}
