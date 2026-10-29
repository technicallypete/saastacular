package identity

import "testing"

// U-RET (Plan r5 §8.2): hostile return_to values never escape the closed allow-list {/, /account, /signup/details}
// (G062 Plan r5 §3.1 adds /signup/details; its look-alikes reduce to /).
func TestReturnPathAllowList(t *testing.T) {
	if len(returnPaths) != 3 {
		t.Fatalf("returnPaths has %d members, want 3", len(returnPaths))
	}
	for _, ok := range []string{"/", "/account", "/signup/details"} {
		if got := ReturnPath(ok); got != ok {
			t.Errorf("ReturnPath(%q) = %q", ok, got)
		}
	}
	for _, hostile := range []string{
		"", "//evil.example", "/\\evil.example", "\\\\evil.example", "%2f%2fevil.example", "/%2f%2fevil.example",
		"%5cevil.example", "/%5c%5cevil.example", "/account/../../evil", "/../account", "https://evil.example/account",
		"javascript:alert(1)", "data:text/html,x", "/account\r\nSet-Cookie: x=y", "/account%0d%0a", "/account?x=1",
		"/account#frag", "/Account", "/account/", " /account", "/account ", "\t/", string(make([]byte, 4096)),
		"/sign-in", "/auth/callback", "/api/account",
		// G062 look-alikes of /signup/details.
		"/signup", "/signup/details/", "/signup/details/../account", "/signup/details?status=x", "/signup/details#f",
		"/Signup/Details", "//signup/details", "/\\signup/details", "%2Fsignup%2Fdetails", "/%2Fevil", "/join/status",
		"https://evil.example/signup/details", "/signup/details\r\n",
	} {
		if got := ReturnPath(hostile); got != "/" {
			t.Errorf("ReturnPath(%q) = %q, want /", hostile, got)
		}
	}
}

func TestOriginsAreExact(t *testing.T) {
	o := NewOrigins([]string{"https://org-a.localhost:543", "https://org-b.localhost:543"})
	for _, ok := range []string{"https://org-a.localhost:543", "https://org-b.localhost:543"} {
		if !o.Allowed(ok) {
			t.Errorf("%q refused", ok)
		}
	}
	for _, bad := range []string{
		"", "null", "https://org-a.localhost", "https://org-a.localhost:543/", "http://org-a.localhost:543",
		"https://attacker.localhost:543", "https://ORG-A.localhost:543", "https://org-a.localhost:5430",
		"https://evil.org-a.localhost:543", "https://org-a.localhost:543.evil.example",
	} {
		if o.Allowed(bad) {
			t.Errorf("%q allowed", bad)
		}
	}
	if CallbackURL("https://org-a.localhost:543") != "https://org-a.localhost:543/auth/callback" ||
		SignInURL("https://org-a.localhost:543") != "https://org-a.localhost:543/sign-in" {
		t.Fatal("derived URLs")
	}
	list := o.List()
	list[0] = "https://changed.example"
	if !o.Allowed("https://org-a.localhost:543") {
		t.Fatal("List exposed internal state")
	}
}

// G016 (Plan r4 §1.7 leg 1/2): a custom-host login state carries the initiating origin as an opaque base64url tag; the
// relay decodes only exact https origins (no path, query, fragment, user info or trailing slash) and nothing else.
func TestStateOriginTag(t *testing.T) {
	for _, o := range []string{"https://portal.example.com", "https://portal.example.com:543"} {
		st := taggedState("AAAAAAAAAAAAAAAAAAAAAA", o)
		if got := stateOrigin(st); got != o {
			t.Errorf("stateOrigin(%q) = %q, want %q", st, got, o)
		}
		if st[:22] != "AAAAAAAAAAAAAAAAAAAAAA" || st[22:23] != stateTagSep {
			t.Errorf("tag layout %q", st)
		}
	}
	for _, bad := range []string{
		"", "AAAA", "AAAA.", ".AAAA", "AAAA.!!!!", "AAAA.aHR0cDovL3BvcnRhbC5leGFtcGxlLmNvbQ", // http://portal.example.com
		taggedState("A", "https://portal.example.com/"), taggedState("A", "https://portal.example.com/x"),
		taggedState("A", "https://portal.example.com?x=1"), taggedState("A", "https://u@portal.example.com"),
		taggedState("A", "https://portal.example.com#f"), taggedState("A", "javascript:alert(1)"), taggedState("A", "https://"),
		taggedState("A", "https://"+string(make([]byte, 300))),
	} {
		if got := stateOrigin(bad); got != "" {
			t.Errorf("stateOrigin(%q) = %q, want empty", bad, got)
		}
	}
}
