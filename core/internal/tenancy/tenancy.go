// Package tenancy is the Organization/Group tenancy model and its single RoleBinding authority store (G005).
//
// Organization is the tenant boundary; a Group belongs to exactly one Organization and has an editable,
// Organization-unique slug. Role bindings are the only writable role store: Organization and Group memberships are
// read through the read-only projections of migration 00003_tenancy.sql. The Service methods are model primitives for
// trusted in-process callers and tests; they are not an authorization system. No HTTP, MCP, CLI or bootstrap path
// reaches them, and this package is linked into neither the server nor the migrate binary. Request context, policy and
// row-level security arrive with G007/G008 (docs/foundation/TENANCY.md).
package tenancy

import (
	"encoding/hex"
	"time"
)

// OrganizationID, GroupID, BindingID and UserID are canonical UUIDs. They are distinct types so that a Group ID can
// never be passed where an Organization or user ID is expected. The zero value is never a valid ID.
type (
	OrganizationID [16]byte
	GroupID        [16]byte
	BindingID      [16]byte
	UserID         [16]byte
)

func (id OrganizationID) String() string { return formatUUID(id) }
func (id GroupID) String() string        { return formatUUID(id) }
func (id BindingID) String() string      { return formatUUID(id) }
func (id UserID) String() string         { return formatUUID(id) }

// ParseOrganizationID parses a canonical hyphenated UUID (case-insensitive). Braces, URNs, the unhyphenated form, the
// empty string and the all-zero UUID are refused with invalid_id.
func ParseOrganizationID(s string) (OrganizationID, error) { return parseUUID[OrganizationID](s) }

// ParseGroupID parses a Group ID; see ParseOrganizationID.
func ParseGroupID(s string) (GroupID, error) { return parseUUID[GroupID](s) }

// ParseBindingID parses a binding ID; see ParseOrganizationID.
func ParseBindingID(s string) (BindingID, error) { return parseUUID[BindingID](s) }

// ParseUserID parses a global user ID (the G004 users.id UUID); see ParseOrganizationID.
func ParseUserID(s string) (UserID, error) { return parseUUID[UserID](s) }

func parseUUID[T ~[16]byte](s string) (T, error) {
	var id T
	if len(s) != 36 || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return id, newError(CodeInvalidID)
	}
	hexDigits := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	b, err := hex.DecodeString(hexDigits)
	if err != nil || len(b) != 16 {
		return id, newError(CodeInvalidID)
	}
	copy(id[:], b)
	if id == (T{}) {
		return T{}, newError(CodeInvalidID)
	}
	return id, nil
}

func formatUUID[T ~[16]byte](id T) string {
	var buf [36]byte
	hex.Encode(buf[0:8], id[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], id[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], id[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], id[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], id[10:16])
	return string(buf[:])
}

// OrganizationRole is an Organization-scope membership role. These are internal semantic codes, not client-facing
// labels; presentation names resolve through configuration in later Goals.
type OrganizationRole string

// The Organization roles.
const (
	OrgOwner  OrganizationRole = "org_owner"
	OrgAdmin  OrganizationRole = "org_admin"
	OrgMember OrganizationRole = "org_member"
)

// GroupRole is a Group-scope membership role. A Group's creator starts as GroupOwner (CreateGroup).
type GroupRole string

// The Group roles.
const (
	GroupOwner  GroupRole = "group_owner"
	GroupAdmin  GroupRole = "group_admin"
	GroupMember GroupRole = "group_member"
)

// ScopeType is the scope of a role binding: an Organization or a Group. There is no other scope.
type ScopeType string

// The supported scopes.
const (
	ScopeOrganization ScopeType = "organization"
	ScopeGroup        ScopeType = "group"
)

// SourceMembership is the only binding source in G005. Invitation and system sources belong to later Goals.
const SourceMembership = "membership"

// Valid reports whether r is one of the three Organization roles.
func (r OrganizationRole) Valid() bool { return r == OrgOwner || r == OrgAdmin || r == OrgMember }

// Valid reports whether r is one of the three Group roles.
func (r GroupRole) Valid() bool { return r == GroupOwner || r == GroupAdmin || r == GroupMember }

// Organization is a tenant. Its ID is immutable; its display name is editable and need not be unique.
type Organization struct {
	ID        OrganizationID
	Name      string
	CreatedAt time.Time
}

// Group belongs to exactly one Organization for its whole life. Name and slug are editable; neither is an identity.
type Group struct {
	ID             GroupID
	OrganizationID OrganizationID
	Name           string
	Slug           string
	CreatedAt      time.Time
}

// Binding is one row of the single authority store. GroupID is set exactly for Group scope; ScopeID equals the
// Organization ID for Organization scope and the Group ID for Group scope.
type Binding struct {
	ID             BindingID
	OrganizationID OrganizationID
	SubjectUserID  UserID
	ScopeType      ScopeType
	ScopeID        [16]byte
	GroupID        *GroupID
	Role           string
	Source         string
	CreatedAt      time.Time
}

// OrganizationMembership is a row of the read-only organization_memberships projection.
type OrganizationMembership struct {
	UserID         UserID
	OrganizationID OrganizationID
	Role           OrganizationRole
}

// GroupMembership is a row of the read-only group_memberships projection. OrganizationID is the Group's tenant anchor.
type GroupMembership struct {
	UserID         UserID
	OrganizationID OrganizationID
	GroupID        GroupID
	Role           GroupRole
}
