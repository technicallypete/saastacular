package identity

import "slices"

// Origins is the exact allow-list of non-tenant origins (IDENTITY_PUBLIC_ORIGINS; since G016 the shared auth host and
// nothing else). Tenant-host origins are not configuration: they resolve through Organization slugs and verified
// domains (HostResolver). Each tenant origin has one callback (<origin>/auth/callback, or the shared relay's for a custom
// host) and one provider sign-out return (<origin>/sign-in, or the auth host's for a custom host). No shared cookie,
// session or cross-host bearer exists (G016 Plan r4 §1.7).
type Origins struct{ list []string }

// NewOrigins builds the allow-list from already validated exact origins.
func NewOrigins(list []string) Origins { return Origins{list: slices.Clone(list)} }

// Allowed reports whether origin is exactly one of the configured origins.
func (o Origins) Allowed(origin string) bool { return origin != "" && slices.Contains(o.list, origin) }

// List returns a copy of the configured origins.
func (o Origins) List() []string { return slices.Clone(o.list) }

// CallbackURL is the registered callback for an allowed origin.
func CallbackURL(origin string) string { return origin + "/auth/callback" }

// SignInURL is the registered provider sign-out return for an allowed origin.
func SignInURL(origin string) string { return origin + "/sign-in" }

// returnPaths is the closed set of requestable post-login destinations. G062 (Plan r5 §3.1): /signup/details is the
// signup-intent login's destination; the database CHECK (00018) holds the same members plus /join/status, which only
// code sets.
var returnPaths = []string{"/", "/account", "/signup/details"}

// ReturnPath maps a requested return_to onto the allow-list. Anything that is not exactly an allowed local path
// (external, scheme-relative, encoded, traversal, backslash, CRLF, overlong) becomes "/". The requested value is
// never echoed, logged or interpolated.
func ReturnPath(requested string) string {
	if slices.Contains(returnPaths, requested) {
		return requested
	}
	return "/"
}
