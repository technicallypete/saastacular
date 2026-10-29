package tenancy

import (
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

// The BusinessProfile context bounds, in Unicode code points (the 00004_business_profiles.sql CHECKs; U-PARITY).
const (
	MaxDisplayNameLength = 255
	MaxLegalNameLength   = 255
	MaxDescriptionLength = 4000
	MaxIndustryLength    = 255
	MaxWebsiteLength     = 2048
)

// BusinessContext is the editable context of a BusinessProfile. DisplayName is required; for the optional fields nil
// means absent (stored as NULL). Values are stored as given after the trim of ValidateBusinessContext, never fetched,
// rendered or interpreted.
type BusinessContext struct {
	DisplayName string
	LegalName   *string
	Description *string
	Industry    *string
	Website     *string
}

// ValidateBusinessContext returns the stored form of c or the code of the first failing field, in the order display
// name, legal name, description, industry, website. Every field must be valid UTF-8; Unicode whitespace is trimmed at
// both ends with no other normalization; an optional field that is nil, empty or whitespace-only is absent; no C0, DEL
// or C1 control character (U+0000-U+001F, U+007F, U+0080-U+009F) is allowed, except TAB and LF inside the description
// (CR is refused everywhere); lengths are 1-255, 1-255, 1-4,000, 1-255 and 1-2,048 code points. The website must also
// satisfy ValidateWebsite. The database CHECKs are never stricter than these rules.
func ValidateBusinessContext(c BusinessContext) (BusinessContext, error) {
	var out BusinessContext
	var err error
	if out.DisplayName, err = contextText(c.DisplayName, MaxDisplayNameLength, false, CodeInvalidDisplayName); err != nil {
		return BusinessContext{}, err
	}
	if out.LegalName, err = optionalText(c.LegalName, MaxLegalNameLength, false, CodeInvalidLegalName); err != nil {
		return BusinessContext{}, err
	}
	if out.Description, err = optionalText(c.Description, MaxDescriptionLength, true, CodeInvalidDescription); err != nil {
		return BusinessContext{}, err
	}
	if out.Industry, err = optionalText(c.Industry, MaxIndustryLength, false, CodeInvalidIndustry); err != nil {
		return BusinessContext{}, err
	}
	if out.Website, err = optionalText(c.Website, MaxWebsiteLength, false, CodeInvalidWebsite); err != nil {
		return BusinessContext{}, err
	}
	if out.Website != nil {
		if err := ValidateWebsite(*out.Website); err != nil {
			return BusinessContext{}, err
		}
	}
	return out, nil
}

// contextText validates one required field: UTF-8, trimmed, no control character (TAB and LF allowed internally when
// multiline), 1-max code points.
func contextText(s string, max int, multiline bool, code Code) (string, error) {
	if !utf8.ValidString(s) {
		return "", newError(code)
	}
	t := strings.TrimSpace(s)
	n := 0
	for _, r := range t {
		if (r <= 0x1f || (r >= 0x7f && r <= 0x9f)) && !(multiline && (r == '\t' || r == '\n')) {
			return "", newError(code)
		}
		n++
	}
	if n < 1 || n > max {
		return "", newError(code)
	}
	return t, nil
}

// optionalText validates one optional field: nil, empty or whitespace-only is absent (nil); anything else follows
// contextText.
func optionalText(s *string, max int, multiline bool, code Code) (*string, error) {
	if s == nil {
		return nil, nil
	}
	if !utf8.ValidString(*s) {
		return nil, newError(code)
	}
	if strings.TrimSpace(*s) == "" {
		return nil, nil
	}
	t, err := contextText(*s, max, multiline, code)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ValidateWebsite accepts a stored website value: 1-2,048 bytes of printable ASCII without space or backslash; an exact
// lowercase http:// or https:// prefix; a URL that parses with no opaque part, no userinfo and a non-empty host; a host
// without % that is an LDH name (which also admits dotted IPv4) or a bracketed IPv6 literal; and a port, when present,
// of 1-65535. Anything else is invalid_website. The value is never fetched, resolved, rendered or treated as domain
// ownership.
func ValidateWebsite(w string) error {
	bad := newError(CodeInvalidWebsite)
	if len(w) < 1 || len(w) > MaxWebsiteLength {
		return bad
	}
	for i := 0; i < len(w); i++ {
		if b := w[i]; b < 0x21 || b > 0x7e || b == '\\' {
			return bad
		}
	}
	if !strings.HasPrefix(w, "http://") && !strings.HasPrefix(w, "https://") {
		return bad
	}
	u, err := url.Parse(w)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Opaque != "" || u.User != nil || u.Host == "" {
		return bad
	}
	if strings.Contains(u.Host, "%") {
		return bad
	}
	host, port := u.Hostname(), u.Port()
	if strings.HasPrefix(u.Host, "[") {
		a, err := netip.ParseAddr(host)
		if err != nil || !a.Is6() || a.Zone() != "" {
			return bad
		}
	} else if !ldhName(host) {
		return bad
	}
	if port != "" {
		p, err := strconv.Atoi(port)
		if err != nil || p < 1 || p > 65535 || strconv.Itoa(p) != port {
			return bad
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return bad
	}
	return nil
}

// ldhName reports an LDH host name: dot-separated labels of letters, digits and inner hyphens, 1-63 bytes each, an
// optional trailing dot, at most 253 bytes. Dotted IPv4 addresses are LDH names too.
func ldhName(h string) bool {
	h = strings.TrimSuffix(h, ".")
	if h == "" || len(h) > 253 {
		return false
	}
	for _, label := range strings.Split(h, ".") {
		if len(label) < 1 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}
