package tenancy

import (
	"regexp"
	"strings"
	"testing"
)

// U-OBJ-INPUT (tenancy part, G009 Plan r1 §5.1): the object selector parser, the content-type validator and the key
// derivation.

func TestParseObjectID(t *testing.T) {
	ok := "3f2a6c1e-9b4d-4e7a-8c5f-0123456789ab"
	id, err := ParseObjectID(ok)
	if err != nil || id.String() != ok {
		t.Fatalf("ParseObjectID(%q) = %v, %v", ok, id, err)
	}
	if up, err := ParseObjectID(strings.ToUpper(ok)); err != nil || up != id {
		t.Errorf("case-insensitive parse: %v %v", up, err)
	}
	for _, bad := range []string{"", "00000000-0000-0000-0000-000000000000", "3f2a6c1e9b4d4e7a8c5f0123456789ab",
		"{3f2a6c1e-9b4d-4e7a-8c5f-0123456789ab}", "urn:uuid:3f2a6c1e-9b4d-4e7a-8c5f-0123456789ab", "3f2a6c1e-9b4d-4e7a-8c5f-0123456789ag",
		"../3f2a6c1e-9b4d-4e7a-8c5f-0123456789", "3f2a6c1e-9b4d-4e7a-8c5f%2F123456789a"} {
		if _, err := ParseObjectID(bad); CodeOf(err) != CodeInvalidID {
			t.Errorf("ParseObjectID(%q): %v, want invalid_id", bad, err)
		}
	}
}

func TestValidateContentType(t *testing.T) {
	for in, want := range map[string]string{
		"":                         DefaultContentType,
		"text/plain":               "text/plain",
		"Image/PNG":                "image/png",
		"application/vnd.api+json": "application/vnd.api+json",
		"a/b":                      "a/b",
	} {
		if got, err := ValidateContentType(in); err != nil || got != want {
			t.Errorf("ValidateContentType(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	long := strings.Repeat("a", 63) + "/" + strings.Repeat("b", 63)
	if got, err := ValidateContentType(long); err != nil || got != long {
		t.Errorf("127-byte type: %q %v", got, err)
	}
	for _, bad := range []string{"text/plain; charset=utf-8", "text", "/plain", "text/", " text/plain", "text/plain ", "text /plain",
		"text/pla\x00in", "text/plain\n", "text/pl ain", strings.Repeat("a", 64) + "/b", "a/" + strings.Repeat("b", 64), "té/xt",
		"text/plain/extra", "-text/plain", "text/*"} {
		if _, err := ValidateContentType(bad); CodeOf(err) != CodeInvalidContentType {
			t.Errorf("ValidateContentType(%q): %v, want invalid_content_type", bad, err)
		}
	}
}

func TestObjectStorageKey(t *testing.T) {
	org, _ := ParseOrganizationID("11111111-2222-4333-8444-555555555555")
	id, _ := ParseObjectID("aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee")
	k := ObjectStorageKey(org, id)
	if k != "objects/11111111-2222-4333-8444-555555555555/aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee" {
		t.Fatalf("key %q", k)
	}
	if !regexp.MustCompile(`^objects/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`).MatchString(k) {
		t.Fatalf("key shape %q", k)
	}
}
