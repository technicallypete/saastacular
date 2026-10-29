package tenancy

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/text/unicode/norm"
)

// InvitationID is the canonical UUID of an invitation (core-generated; keyed insert without RETURNING).
type InvitationID [16]byte

func (id InvitationID) String() string { return formatUUID(id) }

// ParseInvitationID parses an invitation ID; see ParseOrganizationID.
func ParseInvitationID(s string) (InvitationID, error) { return parseUUID[InvitationID](s) }

// Invitation lifetimes (Plan r1 §1.4): default 7 days, minimum 1 hour, maximum 30 days (the CHECK
// invitations_lifetime pins the maximum in the database too).
const (
	InvitationDefaultLifetime = 7 * 24 * time.Hour
	InvitationMinLifetime     = time.Hour
	InvitationMaxLifetime     = 30 * 24 * time.Hour
	// MaxPendingInvitations is the per-scope bound on pending unexpired invitations (invitation_limit).
	MaxPendingInvitations = 100
)

// Address bounds (RFC 5321 local part 64 octets; 254 octets in total), matching invitations_address_normalized.
const (
	maxAddressBytes   = 254
	maxLocalPartBytes = 64
)

// Invitation status values, derived from the timestamps (no status column).
const (
	InvitationPending  = "pending"
	InvitationExpired  = "expired"
	InvitationRevoked  = "revoked"
	InvitationConsumed = "consumed"
)

// Invitation is one row of app.invitations as the runtime role may read it: never the token digest. IntendedAddress is
// present only on rows read through the issuing scope's administrator branch.
type Invitation struct {
	ID               InvitationID
	OrganizationID   OrganizationID
	ScopeType        ScopeType
	ScopeID          [16]byte
	GroupID          *GroupID
	Role             string
	IssuerUserID     UserID
	IntendedUserID   *UserID
	IntendedAddress  *string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	ConsumedAt       *time.Time
	ConsumedByUserID *UserID
	ConsumedBinding  *BindingID
}

// Status derives the lifecycle state at now: consumed and revoked are final; otherwise expired after ExpiresAt.
func (inv Invitation) Status(now time.Time) string {
	switch {
	case inv.ConsumedAt != nil:
		return InvitationConsumed
	case inv.RevokedAt != nil:
		return InvitationRevoked
	case !now.Before(inv.ExpiresAt):
		return InvitationExpired
	default:
		return InvitationPending
	}
}

// InvitationRoleValid reports whether role is an invitable role of the scope: admin or member only (never an owner
// role; the CHECK invitations_scope_consistent repeats it).
func InvitationRoleValid(scope ScopeType, role string) bool {
	switch scope {
	case ScopeOrganization:
		return role == string(OrgAdmin) || role == string(OrgMember)
	case ScopeGroup:
		return role == string(GroupAdmin) || role == string(GroupMember)
	}
	return false
}

// NormalizeAddress is the single address normalization (Plan r1 §1.4 C4; correction-01 B-1(ii)): trim leading and
// trailing ASCII whitespace; require exactly one '@' with a non-empty local part and domain; refuse any remaining
// whitespace or control character, invalid UTF-8, a local part over 64 bytes or an address over 254 bytes. The local
// part is preserved byte-exact (no case folding, NFC or dot/plus rewriting); the domain is NFC-normalized and then
// lowercased per Unicode default case mapping, IDN labels kept as given (no punycode conversion). The address is a
// delivery hint and a match key, never an identity.
func NormalizeAddress(s string) (string, error) {
	s = strings.Trim(s, " \t\r\n\v\f")
	if s == "" || !utf8.ValidString(s) {
		return "", newError(CodeInvalidAddress)
	}
	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return "", newError(CodeInvalidAddress)
		}
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 || strings.IndexByte(s[at+1:], '@') >= 0 {
		return "", newError(CodeInvalidAddress)
	}
	local := s[:at]
	domain := strings.ToLower(norm.NFC.String(s[at+1:]))
	out := local + "@" + domain
	if len(local) > maxLocalPartBytes || len(out) > maxAddressBytes || strings.IndexByte(domain, '@') >= 0 {
		return "", newError(CodeInvalidAddress)
	}
	return out, nil
}

// invitationTokenTag domain-separates the invitation digest from every other digest in the system.
const invitationTokenTag = "invitation/1\x00"

// invitationTokenLen is the length of a token: 32 random bytes, base64url without padding.
const invitationTokenLen = 43

// NewInvitationToken returns a fresh token (32 bytes from crypto/rand, base64url without padding) and its digest. The
// raw token is returned to the issuer once and is never stored, logged or audited.
func NewInvitationToken() (token string, digest []byte, err error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", nil, newError(CodeInternal)
	}
	token = base64.RawURLEncoding.EncodeToString(b[:])
	return token, InvitationDigest(token), nil
}

// ValidInvitationToken reports whether token has the exact shape NewInvitationToken produces.
func ValidInvitationToken(token string) bool {
	if len(token) != invitationTokenLen {
		return false
	}
	b, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(b) == 32
}

// InvitationDigest is sha256("invitation/1\x00" || token): the only form in which a token is stored or compared.
func InvitationDigest(token string) []byte {
	h := sha256.New()
	h.Write([]byte(invitationTokenTag))
	h.Write([]byte(token))
	return h.Sum(nil)
}

// DigestEqual compares two digests in constant time.
func DigestEqual(a, b []byte) bool { return len(a) == len(b) && subtle.ConstantTimeCompare(a, b) == 1 }

// InvitationSpec is one invitation to issue (Plan r1 §1.4 "Invitations"). Group zero selects the Organization scope,
// whose recipient form is IntendedAddress (normalized here) with IntendedUser zero; a Group invitation names
// IntendedUser and no address. Lifetime zero is InvitationDefaultLifetime; it is counted in whole seconds.
type InvitationSpec struct {
	ID              InvitationID
	OrganizationID  OrganizationID
	Group           GroupID
	Role            string
	Issuer          UserID
	IntendedUser    UserID
	IntendedAddress string
	Lifetime        time.Duration
}

// validate returns the scope, the normalized address (nil for a Group invitation) and the lifetime in seconds.
func (sp InvitationSpec) validate() (ScopeType, *string, int64, error) {
	if zero(sp.ID) || zero(sp.OrganizationID) || zero(sp.Issuer) {
		return "", nil, 0, newError(CodeInvalidID)
	}
	scope := ScopeOrganization
	if !zero(sp.Group) {
		scope = ScopeGroup
	}
	if !InvitationRoleValid(scope, sp.Role) {
		return "", nil, 0, newError(CodeInvalidRole)
	}
	var addr *string
	switch scope {
	case ScopeOrganization:
		if !zero(sp.IntendedUser) {
			return "", nil, 0, newError(CodeInvalidID)
		}
		a, err := NormalizeAddress(sp.IntendedAddress)
		if err != nil {
			return "", nil, 0, err
		}
		addr = &a
	case ScopeGroup:
		if zero(sp.IntendedUser) {
			return "", nil, 0, newError(CodeInvalidID)
		}
		if sp.IntendedAddress != "" {
			return "", nil, 0, newError(CodeInvalidAddress)
		}
		// G060 (r1 C3): an issuer never invites itself into a Group; its own Group role comes only from another authorized
		// issuer. The claim body refuses to consume such an invitation as well.
		if sp.IntendedUser == sp.Issuer {
			return "", nil, 0, newError(CodeSelfTarget)
		}
	}
	life := sp.Lifetime
	if life == 0 {
		life = InvitationDefaultLifetime
	}
	life = life.Truncate(time.Second)
	if life < InvitationMinLifetime || life > InvitationMaxLifetime {
		return "", nil, 0, newError(CodeInvalidLifetime)
	}
	return scope, addr, int64(life / time.Second), nil
}

// invitationKeyCollisions: a collision of the core-generated id or of the 32-byte random token digest is a store
// failure (unavailable), never a conflict the caller could act on.
var invitationKeyCollisions = map[string]bool{"invitations_pkey": true, "invitations_token_digest_key": true}

// InsertInvitationTx validates sp and issues it inside the caller's transaction, after the facade has taken the scope
// row at exclusive strength and checked the pending count (C5): a fresh token, one keyed INSERT of its digest without
// RETURNING (app.invitations_bound repeats lock-then-count: invitation_limit), then the read of the new row. The raw
// token is returned once and never stored, logged or audited. A row the read cannot see is unavailable.
func InsertInvitationTx(ctx context.Context, tx pgx.Tx, sp InvitationSpec) (Invitation, string, error) {
	scope, addr, secs, err := sp.validate()
	if err != nil {
		return Invitation{}, "", err
	}
	token, digest, err := NewInvitationToken()
	if err != nil {
		return Invitation{}, "", err
	}
	scopeID, group, intended := [16]byte(sp.OrganizationID), pgtype.UUID{}, pgtype.UUID{}
	if scope == ScopeGroup {
		scopeID, group, intended = sp.Group, uuidArg(sp.Group), uuidArg(sp.IntendedUser)
	}
	if _, err := tx.Exec(ctx, sqlInsertInvitation, uuidArg(sp.OrganizationID), uuidArg(sp.ID), string(scope),
		pgtype.UUID{Bytes: scopeID, Valid: true}, group, sp.Role, uuidArg(sp.Issuer), intended, addr, digest, secs); err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" && invitationKeyCollisions[pg.ConstraintName] {
			return Invitation{}, "", newError(CodeUnavailable)
		}
		return Invitation{}, "", classify(err)
	}
	inv, err := scanInvitation(tx.QueryRow(ctx, sqlSelectInvitation, uuidArg(sp.OrganizationID), uuidArg(sp.ID)))
	switch {
	case isNoRows(err):
		return Invitation{}, "", newError(CodeUnavailable)
	case err != nil:
		return Invitation{}, "", classify(err)
	}
	return inv, token, nil
}

// RevokeInvitationTx revokes the pending invitation id of org inside the caller's transaction and returns the revoked
// row. A row that exists for the context but is no longer pending (revoked or consumed) is invitation_not_pending; a
// row the context cannot see or manage is invitation_not_found.
func RevokeInvitationTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id InvitationID) (Invitation, error) {
	if zero(org) || zero(id) {
		return Invitation{}, newError(CodeInvalidID)
	}
	inv, err := scanInvitation(tx.QueryRow(ctx, sqlRevokeInvitation, uuidArg(org), uuidArg(id)))
	if err == nil {
		return inv, nil
	}
	if !isNoRows(err) {
		return Invitation{}, classify(err)
	}
	if err := exists(ctx, tx, CodeInvitationNotFound, sqlInvitationExists, uuidArg(org), uuidArg(id)); err != nil {
		return Invitation{}, classify(err)
	}
	return Invitation{}, newError(CodeInvitationNotPending)
}

// InvitationTx reads invitation id of org inside the caller's transaction, as row-level security admits it: the issuing
// scope's administrators (with intended_address) and the intended recipient. A row the context cannot see is
// invitation_not_found; the digest is never read.
func InvitationTx(ctx context.Context, tx pgx.Tx, org OrganizationID, id InvitationID) (Invitation, error) {
	if zero(org) || zero(id) {
		return Invitation{}, newError(CodeInvalidID)
	}
	inv, err := scanInvitation(tx.QueryRow(ctx, sqlSelectInvitation, uuidArg(org), uuidArg(id)))
	if e := classify(one(err, CodeInvitationNotFound)); e != nil {
		return Invitation{}, e
	}
	return inv, nil
}

// Claim outcomes (00015 invitation_claim_group / invitation_claim_organization): a new binding, an upgraded one, or a
// replay by the recipient who already holds a qualifying binding.
const (
	ClaimJoined        = "joined"
	ClaimUpgraded      = "upgraded"
	ClaimAlreadyMember = "already_member"
)

// The two refusal outcomes of the claim bodies: the one closed not_found and the consumed replay without a qualifying
// binding (never restored).
const (
	claimNotFound = "not_found"
	claimConsumed = "consumed"
)

// ClaimOutcome is a successful consumption: the outcome, the scope and the resulting binding (and its role).
type ClaimOutcome struct {
	Outcome        string
	OrganizationID OrganizationID
	ScopeType      ScopeType
	ScopeID        [16]byte
	GroupID        *GroupID
	Role           string
	BindingID      BindingID
}

// OrganizationClaim carries the provider address evidence of an Organization invitation claim (r2 addendum D-14a):
// built only by the access facade from identity's VerifiedAddress, never from request input.
type OrganizationClaim struct {
	Issuer  string
	Subject string
	Address string
}

// ClaimInvitationTx consumes invitation id with token inside the caller's transaction, as the context user. With org
// nil it is the Group transition (app.invitation_claim_group; Group-scope context); otherwise the Organization
// transition (app.invitation_claim_organization; IDENTITY-shape context, owner role only in G019, D-15 pending). A
// malformed token is invalid_token before any statement; every mismatch is the closed invitation_not_found; a consumed
// replay without a qualifying binding is invitation_not_pending; the bodies' guard refusals arrive as their codes.
func ClaimInvitationTx(ctx context.Context, tx pgx.Tx, id InvitationID, token string, org *OrganizationClaim) (ClaimOutcome, error) {
	if zero(id) {
		return ClaimOutcome{}, newError(CodeInvalidID)
	}
	if !ValidInvitationToken(token) {
		return ClaimOutcome{}, newError(CodeInvalidToken)
	}
	var row pgx.Row
	if org == nil {
		row = tx.QueryRow(ctx, sqlClaimGroupInvitation, uuidArg(id), InvitationDigest(token))
	} else {
		if org.Issuer == "" || org.Subject == "" {
			return ClaimOutcome{}, newError(CodeInvalidID)
		}
		addr, err := NormalizeAddress(org.Address)
		if err != nil {
			return ClaimOutcome{}, err
		}
		row = tx.QueryRow(ctx, sqlClaimOrganizationInvitation, uuidArg(id), InvitationDigest(token), org.Issuer, org.Subject, addr)
	}
	out, err := scanClaim(row)
	if err != nil {
		return ClaimOutcome{}, classify(one(err, CodeInvitationNotFound))
	}
	return out, nil
}

// scanClaim maps one claim row: the success outcomes to a ClaimOutcome, the refusal outcomes to their codes, and any
// other value (or a success row without its scope or binding) to internal.
func scanClaim(row pgx.Row) (ClaimOutcome, error) {
	var outcome string
	var org, scope, group, binding pgtype.UUID
	var scopeType, role pgtype.Text
	if err := row.Scan(&outcome, &org, &scopeType, &scope, &group, &role, &binding); err != nil {
		return ClaimOutcome{}, err
	}
	switch outcome {
	case claimNotFound:
		return ClaimOutcome{}, newError(CodeInvitationNotFound)
	case claimConsumed:
		return ClaimOutcome{}, newError(CodeInvitationNotPending)
	case ClaimJoined, ClaimUpgraded, ClaimAlreadyMember:
	default:
		return ClaimOutcome{}, newError(CodeInternal)
	}
	if !org.Valid || !scope.Valid || !binding.Valid || !scopeType.Valid || !role.Valid {
		return ClaimOutcome{}, newError(CodeInternal)
	}
	out := ClaimOutcome{Outcome: outcome, OrganizationID: OrganizationID(org.Bytes), ScopeType: ScopeType(scopeType.String),
		ScopeID: scope.Bytes, Role: role.String, BindingID: BindingID(binding.Bytes)}
	if group.Valid {
		g := GroupID(group.Bytes)
		out.GroupID = &g
	}
	return out, nil
}

// scopeOf returns the scope addressed by org and group (group zero: the Organization scope).
func scopeOf(org OrganizationID, group GroupID) (ScopeType, pgtype.UUID) {
	if zero(group) {
		return ScopeOrganization, uuidArg(org)
	}
	return ScopeGroup, uuidArg(group)
}

// InvitationsPage reads at most limit+1 invitations of a scope of org (group zero: the Organization scope) ordered by
// id after after (zero: from the start), with intended_address, so the caller can tell whether a next page exists.
// Row-level security limits the rows to the issuing scope's administrators (and recipients of their own rows).
func InvitationsPage(ctx context.Context, tx pgx.Tx, org OrganizationID, group GroupID, after InvitationID, limit int) ([]Invitation, error) {
	if zero(org) || limit < 1 || limit > MaxMembersPage {
		return nil, newError(CodeInvalidID)
	}
	scope, scopeID := scopeOf(org, group)
	rows, err := tx.Query(ctx, sqlInvitationsPage, uuidArg(org), string(scope), scopeID, uuidArg(after), limit+1)
	out, err := collect(rows, err, scanInvitation)
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

// MyInvitationsPage reads at most limit+1 open (neither consumed nor revoked) invitations of org addressed to
// recipient, ordered by id after after, never with intended_address; expiry stays derived status.
func MyInvitationsPage(ctx context.Context, tx pgx.Tx, org OrganizationID, recipient UserID, after InvitationID, limit int) ([]Invitation, error) {
	if zero(org) || zero(recipient) || limit < 1 || limit > MaxMembersPage {
		return nil, newError(CodeInvalidID)
	}
	rows, err := tx.Query(ctx, sqlMyInvitationsPage, uuidArg(org), uuidArg(recipient), uuidArg(after), limit+1)
	out, err := collect(rows, err, scanInvitation)
	if e := classify(err); e != nil {
		return nil, e
	}
	return out, nil
}

// PendingInvitationCount returns the pending unexpired invitations of a scope of org (group zero: the Organization
// scope): the facade's invitation_limit precheck after the scope lock; app.invitations_bound is the guarantee.
func PendingInvitationCount(ctx context.Context, tx pgx.Tx, org OrganizationID, group GroupID) (int, error) {
	if zero(org) {
		return 0, newError(CodeInvalidID)
	}
	scope, scopeID := scopeOf(org, group)
	var n int64
	if err := tx.QueryRow(ctx, sqlInvitationsPendingCount, uuidArg(org), string(scope), scopeID).Scan(&n); err != nil {
		return 0, classify(err)
	}
	return int(n), nil
}

func scanInvitation(row pgx.Row) (Invitation, error) {
	var inv Invitation
	var id, org, scope, group, issuer, intended, consumedBy, binding pgtype.UUID
	var scopeType string
	if err := row.Scan(&id, &org, &scopeType, &scope, &group, &inv.Role, &issuer, &intended, &inv.IntendedAddress,
		&inv.CreatedAt, &inv.ExpiresAt, &inv.RevokedAt, &inv.ConsumedAt, &consumedBy, &binding); err != nil {
		return Invitation{}, err
	}
	inv.ID, inv.OrganizationID, inv.ScopeType, inv.ScopeID = InvitationID(id.Bytes), OrganizationID(org.Bytes), ScopeType(scopeType), scope.Bytes
	inv.IssuerUserID = UserID(issuer.Bytes)
	if group.Valid {
		g := GroupID(group.Bytes)
		inv.GroupID = &g
	}
	if intended.Valid {
		u := UserID(intended.Bytes)
		inv.IntendedUserID = &u
	}
	if consumedBy.Valid {
		u := UserID(consumedBy.Bytes)
		inv.ConsumedByUserID = &u
	}
	if binding.Valid {
		b := BindingID(binding.Bytes)
		inv.ConsumedBinding = &b
	}
	return inv, nil
}
