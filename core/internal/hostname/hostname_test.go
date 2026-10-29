package hostname

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestHostnameNormalize(t *testing.T) {
	long := strings.Repeat("a", 63)
	// 3×63 + 59 + 1 bytes of labels and 4 dots: exactly 253 bytes.
	name253 := strings.Join([]string{long, long, long, strings.Repeat("b", 59), "z"}, ".")
	cases := []struct {
		raw, port string
		want      Host
		err       error
	}{
		{"org-a.example.net", "", Host{Name: "org-a.example.net"}, nil},
		{"ORG-A.Example.NET", "", Host{Name: "org-a.example.net"}, nil},
		{"org-a.example.net.", "", Host{Name: "org-a.example.net"}, nil},
		{"org-a.example.net..", "", Host{}, ErrMalformed},
		{"org-a.localhost:543", "543", Host{Name: "org-a.localhost", Port: "543"}, nil},
		{"org-a.localhost.:543", "543", Host{Name: "org-a.localhost", Port: "543"}, nil},
		{"org-a.localhost:544", "543", Host{}, ErrPort},
		{"org-a.example.net:443", "", Host{}, ErrPort},
		{"org-a.example.net:", "", Host{}, ErrMalformed},
		{"org-a.example.net:+43", "543", Host{}, ErrMalformed},
		{"org-a.example.net:123456", "543", Host{}, ErrMalformed},
		{"localhost", "", Host{Name: "localhost"}, nil},
		{name253, "", Host{Name: name253}, nil},
		{name253 + "z", "", Host{}, ErrMalformed},
		{long + "a.example.net", "", Host{}, ErrMalformed},
		{"", "", Host{}, ErrMalformed},
		{strings.Repeat("a", MaxHeaderBytes+1), "", Host{}, ErrMalformed},
		{"bücher.example", "", Host{}, ErrNonASCII},
		{"xn--bcher-kva.example", "", Host{}, ErrNonASCII},
		{"a.XN--p1ai", "", Host{}, ErrNonASCII},
		{"192.168.0.1", "", Host{}, ErrMalformed},
		{"192.168.0.1:543", "543", Host{}, ErrMalformed},
		{"[::1]", "", Host{}, ErrMalformed},
		{"[::1]:543", "543", Host{}, ErrMalformed},
		{"user@org-a.example.net", "", Host{}, ErrMalformed},
		{"org-a.example.net/path", "", Host{}, ErrMalformed},
		{"https://org-a.example.net", "", Host{}, ErrMalformed},
		{"org-a.example.net ", "", Host{}, ErrMalformed},
		{" org-a.example.net", "", Host{}, ErrMalformed},
		{"org_a.example.net", "", Host{}, ErrMalformed},
		{"-org.example.net", "", Host{}, ErrMalformed},
		{"org-.example.net", "", Host{}, ErrMalformed},
		{"org..example.net", "", Host{}, ErrMalformed},
		{".example.net", "", Host{}, ErrMalformed},
		{"org-a.example.net,org-b.example.net", "", Host{}, ErrMalformed},
		{"org-a.example.net\r\nx", "", Host{}, ErrMalformed},
	}
	for _, c := range cases {
		got, err := Normalize(c.raw, c.port)
		if !errors.Is(err, c.err) || got != c.want {
			t.Errorf("Normalize(%q, %q) = %+v, %v; want %+v, %v", c.raw, c.port, got, err, c.want, c.err)
		}
	}
	h := Host{Name: "org-a.localhost", Port: "543"}
	if h.HostPort() != "org-a.localhost:543" || h.Origin() != "https://org-a.localhost:543" {
		t.Errorf("HostPort/Origin = %q %q", h.HostPort(), h.Origin())
	}
	if (Host{Name: "portal.acme.test"}).Origin() != "https://portal.acme.test" {
		t.Error("portless origin")
	}
}

func TestHostnameClassify(t *testing.T) {
	cfg := Config{
		BaseDomain: "example.net", CNAMETarget: "custom.example.net", AuthHost: "login.example.net",
		ReservedLabels: []string{"blog"},
	}
	cases := []struct {
		name string
		want Classification
	}{
		{"example.net", Classification{Kind: KindPublicEntry}},
		{"www.example.net", Classification{Kind: KindPublicEntry}},
		{"custom.example.net", Classification{Kind: KindReserved}},
		{"login.example.net", Classification{Kind: KindReserved}},
		{"auth.example.net", Classification{Kind: KindReserved}},
		{"api.example.net", Classification{Kind: KindReserved}},
		{"status.example.net", Classification{Kind: KindReserved}},
		{"blog.example.net", Classification{Kind: KindReserved}},
		{"org-a.example.net", Classification{Kind: KindPlatform, Slug: "org-a"}},
		{"7.example.net", Classification{Kind: KindPlatform, Slug: "7"}},
		// a generated platform label (G057: 26 lowercase base32 characters, 128 bits) is an ordinary platform label
		{"mfrggzdfmztwq2lknnwg23tpoa.example.net", Classification{Kind: KindPlatform, Slug: "mfrggzdfmztwq2lknnwg23tpoa"}},
		{"mfrggzdfmztwq2lknnwg23tpoa.org-a.example.net", Classification{Kind: KindNone}},
		{"a.b.example.net", Classification{Kind: KindNone}},
		{"www.org-a.example.net", Classification{Kind: KindNone}},
		{"a--b.example.net", Classification{Kind: KindNone}},
		{"portal.acme.test", Classification{Kind: KindCustom}},
		{"acme-example.net", Classification{Kind: KindCustom}},
		{"example.net.evil.test", Classification{Kind: KindCustom}},
		{"xexample.net", Classification{Kind: KindCustom}},
		{"localhost", Classification{Kind: KindCustom}},
	}
	for _, c := range cases {
		if got := Classify(Host{Name: c.name}, cfg); got != c.want {
			t.Errorf("Classify(%q) = %+v; want %+v", c.name, got, c.want)
		}
	}
	// A CNAME target or auth host outside the platform domain reserves no label under it.
	other := Config{BaseDomain: "example.net", CNAMETarget: "edge.provider.test"}
	if got := Classify(Host{Name: "edge.example.net"}, other); got.Kind != KindPlatform {
		t.Errorf("edge.example.net under a foreign target = %+v", got)
	}
	if IsReservedLabel("custom", Config{BaseDomain: "example.net", CNAMETarget: "custom.example.net"}) != true {
		t.Error("floor label custom must be reserved")
	}
	if !IsReservedLabel("login", cfg) || IsReservedLabel("org-a", cfg) {
		t.Error("configured auth host label must be reserved; org-a must not")
	}
	// A configured 26-character value is reserved like any other configured label (the label generator rejects it).
	generated := Config{BaseDomain: "example.net", ReservedLabels: []string{"mfrggzdfmztwq2lknnwg23tpoa"}}
	if got := Classify(Host{Name: "mfrggzdfmztwq2lknnwg23tpoa.example.net"}, generated); got.Kind != KindReserved {
		t.Errorf("configured reserved 26-character label = %+v", got)
	}

	for _, c := range []struct {
		name string
		ok   bool
	}{
		{"portal.acme.test", true}, {"acme.io", true}, {"example.net", false}, {"custom.example.net", false},
		{"login.example.net", false}, {"org-a.example.net", false}, {"deep.org-a.example.net", false},
		{"localhost", false}, {"example.net.evil.test", true},
	} {
		if got := CustomClaimAllowed(c.name, cfg); got != c.ok {
			t.Errorf("CustomClaimAllowed(%q) = %v; want %v", c.name, got, c.ok)
		}
	}
}

func TestReservedLabelsFloor(t *testing.T) {
	got := ReservedLabels()
	if !slices.IsSorted(got) || len(slices.Compact(slices.Clone(got))) != len(got) {
		t.Fatalf("ReservedLabels not sorted/unique: %v", got)
	}
	for _, l := range append(slices.Clone(groupReserved), hostReserved...) {
		if !slices.Contains(got, l) {
			t.Errorf("floor lacks %q", l)
		}
	}
	if len(got) != 35 {
		t.Errorf("floor size = %d; want 35 (24 Group + 12 host, auth shared)", len(got))
	}
	for _, l := range got {
		if !SlugGrammar(l) {
			t.Errorf("floor label %q outside the slug grammar", l)
		}
	}
}

// TestPlatformRoot (G062 Plan r5 §4.1): the root origin is https://<domain>[:port] from configuration only, "" without a
// platform domain; only the platform domain itself is the root host (never www.<domain>, a tenant or a look-alike).
func TestPlatformRoot(t *testing.T) {
	for _, c := range []struct {
		cfg  Config
		want string
	}{
		{Config{BaseDomain: "example.net"}, "https://example.net"},
		{Config{BaseDomain: "localhost", Port: "543"}, "https://localhost:543"},
		{Config{}, ""},
		{Config{Port: "543"}, ""},
	} {
		if got := PlatformRootOrigin(c.cfg); got != c.want {
			t.Errorf("PlatformRootOrigin(%+v) = %q, want %q", c.cfg, got, c.want)
		}
	}
	cfg := Config{BaseDomain: "localhost", CNAMETarget: "edge.localhost", AuthHost: "auth.localhost", Port: "543"}
	if !IsPlatformRoot(Host{Name: "localhost", Port: "543"}, cfg) || !IsPlatformRoot(Host{Name: "localhost"}, cfg) {
		t.Error("the platform domain is not the root")
	}
	for _, n := range []string{"www.localhost", "org-a.localhost", "auth.localhost", "edge.localhost", "localhost.example",
		"xlocalhost", ""} {
		if IsPlatformRoot(Host{Name: n, Port: "543"}, cfg) {
			t.Errorf("%q classified as the platform root", n)
		}
	}
	if IsPlatformRoot(Host{Name: ""}, Config{}) {
		t.Error("an empty platform domain matched")
	}
	// Both public entry names classify as public_entry; only the platform domain is the root.
	for _, n := range []string{"localhost", "www.localhost"} {
		if k := Classify(Host{Name: n}, cfg).Kind; k != KindPublicEntry {
			t.Errorf("%s: kind %s", n, k)
		}
	}
}
