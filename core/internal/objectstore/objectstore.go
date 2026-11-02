// Package objectstore is core's owned object-storage provider boundary (G009; docs/foundation/OBJECT-STORAGE.md).
//
// The Store interface is provider-neutral: Put, Get, Delete, SignedGetURL and Metadata (Head) of exactly one
// server-derived key. It accepts only Key values of two families, objects/<organization uuid>/<object uuid> (G009) and
// avatars/<user uuid>/<avatar uuid> (G074: identity-owned account avatars, never presigned), carries the object's
// binding metadata, and reports failures as a closed set of errors without provider or SDK text. It makes no
// authorization decision: only core/internal/tenancy/access imports it (U-ACCESS-BOUNDARY), after its guard. The key
// path names the owning Organization or user; it is never the authorization boundary.
package objectstore

import (
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
)

// The closed error categories. No error returned by this package carries provider, SDK, URL or credential text.
var (
	ErrNotFound       = errors.New("objectstore: not_found")       // the key holds no object
	ErrDenied         = errors.New("objectstore: denied")          // the provider refused the credential or request
	ErrUnavailable    = errors.New("objectstore: unavailable")     // the provider could not be reached or failed
	ErrOutcomeUnknown = errors.New("objectstore: outcome_unknown") // a write may or may not have taken effect
	ErrInvalid        = errors.New("objectstore: invalid")         // a malformed key, binding, size or option
)

const uuidPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`

var (
	keyRe  = regexp.MustCompile(`^(objects|avatars)/` + uuidPattern + `/` + uuidPattern + `$`)
	uuidRe = regexp.MustCompile(`^` + uuidPattern + `$`)
)

const zeroUUID = "00000000-0000-0000-0000-000000000000"

// Family is a key's family: the Organization-bound G009 objects or the identity-owned G074 avatars.
type Family string

// The key families.
const (
	FamilyObjects Family = "objects"
	FamilyAvatars Family = "avatars"
)

// Key is a validated provider key: objects/<organization uuid>/<object uuid> or avatars/<user uuid>/<avatar uuid>,
// lowercase canonical UUIDs only, neither of them zero. The zero Key is invalid and refused by every Store method before
// any request.
type Key struct{ s string }

// ParseKey validates a key. Traversal, encoded separators, uppercase, other prefixes and every other form are refused
// with ErrInvalid.
func ParseKey(s string) (Key, error) {
	if !keyRe.MatchString(s) {
		return Key{}, ErrInvalid
	}
	owner, _, _ := strings.Cut(s[strings.IndexByte(s, '/')+1:], "/")
	if owner == zeroUUID || s[len(s)-36:] == zeroUUID {
		return Key{}, ErrInvalid
	}
	return Key{s: s}, nil
}

// NewKey derives the key of an object from its Organization and object UUIDs (canonical lowercase strings).
func NewKey(organization, object string) (Key, error) {
	return ParseKey("objects/" + organization + "/" + object)
}

// NewAvatarKey derives the key of an account avatar from its user and avatar UUIDs (canonical lowercase strings).
func NewAvatarKey(user, object string) (Key, error) {
	return ParseKey("avatars/" + user + "/" + object)
}

// String returns the key.
func (k Key) String() string { return k.s }

// Valid reports whether k came from ParseKey, NewKey or NewAvatarKey.
func (k Key) Valid() bool { return k.s != "" }

// Family returns the key's family ("" for the zero Key).
func (k Key) Family() Family {
	switch {
	case strings.HasPrefix(k.s, "objects/"):
		return FamilyObjects
	case strings.HasPrefix(k.s, "avatars/"):
		return FamilyAvatars
	}
	return ""
}

// Binding is the object's identity as stored with its bytes (user metadata object-id, organization-id, group-id and,
// G074, user-id): canonical lowercase UUIDs with exactly one owner, either an Organization (Group optional, User empty)
// for an objects/ key or a User (Organization and Group empty) for an avatars/ key. Reads and the post-upload
// verification compare it with the metadata row.
type Binding struct {
	Object       string
	Organization string
	Group        string
	User         string
}

func (b Binding) valid() bool {
	if !uuidRe.MatchString(b.Object) {
		return false
	}
	if b.User != "" {
		return uuidRe.MatchString(b.User) && b.Organization == "" && b.Group == ""
	}
	return uuidRe.MatchString(b.Organization) && (b.Group == "" || uuidRe.MatchString(b.Group))
}

// fits reports whether the binding's owner kind is the key family's: an Organization binding for objects/, a user
// binding for avatars/.
func (b Binding) fits(k Key) bool {
	switch k.Family() {
	case FamilyObjects:
		return b.User == ""
	case FamilyAvatars:
		return b.User != ""
	}
	return false
}

// Info is what the provider reports for a stored object: its exact length and binding.
type Info struct {
	Size    int64
	Binding Binding
}

// Store is the provider-neutral object store. Every method takes one validated key. Put streams exactly size bytes of
// body with an explicit length (no buffering, no automatic retry); Get returns the object's info and its body, which
// the caller closes; Delete is idempotent (an absent object is success); Metadata is a Head; SignedGetURL presigns a GET
// of the key offline for ttl with fixed attachment, content-type and no-store response overrides.
type Store interface {
	Put(ctx context.Context, key Key, body io.Reader, size int64, contentType string, b Binding) error
	Get(ctx context.Context, key Key) (Info, io.ReadCloser, error)
	Delete(ctx context.Context, key Key) error
	Metadata(ctx context.Context, key Key) (Info, error)
	SignedGetURL(ctx context.Context, key Key, ttl time.Duration, contentType string) (string, error)
}
