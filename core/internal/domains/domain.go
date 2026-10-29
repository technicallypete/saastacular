// Package domains owns the G016 hostname facts that are not authorization: the custom-domain record model and its pure
// lifecycle table (Plan r4 §1.4), the DNS ownership checker, the generation-bound TLS adapter (§1.5) and the durable
// hostname resolver (§1.3, §5.5). It imports nothing from tenancy/access or identity: the guarded operations live in
// the access facade and the HTTP surface in domains/api (§5.0). Hostnames select an Organization; they never authorize.
package domains

import (
	"errors"
	"time"
)

// State is the finite custom-domain state.
type State string

// The states (00011 organization_domains_state_valid).
const (
	StatePending   State = "pending"
	StateVerified  State = "verified"
	StateActive    State = "active"
	StateSuspended State = "suspended"
	StateRemoved   State = "removed"
)

// TLSState is the observed TLS readiness of a hostname (00011 organization_domains_tls_state_valid).
type TLSState string

// The TLS states.
const (
	TLSNone        TLSState = "none"
	TLSRequested   TLSState = "requested"
	TLSPending     TLSState = "pending"
	TLSReady       TLSState = "ready"
	TLSError       TLSState = "error"
	TLSExpired     TLSState = "expired"
	TLSWithdrawing TLSState = "withdrawing"
	TLSWithdrawn   TLSState = "withdrawn"
	TLSUnknown     TLSState = "unknown"
)

// Outcome is a closed check outcome: the DNS proof outcomes, the TLS-only failure and the removed row's withdrawal
// outcomes (00011 organization_domains_outcome_valid; the domain.checked audit reason).
type Outcome string

// The outcomes.
const (
	OutcomeOK              Outcome = "ok"
	OutcomeMismatch        Outcome = "mismatch"
	OutcomeNXDomain        Outcome = "nxdomain"
	OutcomeTransient       Outcome = "transient"
	OutcomeTLSError        Outcome = "tls_error"
	OutcomeWithdrawn       Outcome = "withdrawn"
	OutcomeWithdrawError   Outcome = "withdraw_error"
	OutcomeWithdrawUnknown Outcome = "withdraw_unknown"
)

// Record is one organization_domains row as the facade reads it.
type Record struct {
	ID                 string
	OrganizationID     string
	Hostname           string
	Generation         int64
	State              State
	ChallengeValue     string
	ChallengeExpiresAt time.Time
	VerifiedAt         *time.Time
	VerifiedUntil      *time.Time
	LastCheckAt        *time.Time
	LastCheckOutcome   *string
	CheckFailures      int32
	TLSState           TLSState
	TLSCheckedAt       *time.Time
	TLSNotAfter        *time.Time
	TLSReference       *string
	RemovedAt          *time.Time
	CreatedAt          time.Time
	Version            int64
}

// Ref returns the generation-bound adapter reference of the row.
func (r Record) Ref() DomainRef {
	ref := DomainRef{ID: r.ID, Generation: r.Generation, Hostname: r.Hostname}
	if r.TLSReference != nil {
		ref.Reference = *r.TLSReference
	}
	return ref
}

// DomainRef binds every TLS adapter call to one row and claim generation (Plan r4 §1.5). Results are applied only to
// the row with this ID and Generation under the version guard; there is no name-only operation.
type DomainRef struct {
	ID         string
	Generation int64
	Hostname   string
	Reference  string // the provider resource of this generation (E2); empty under E1
}

// TLSResult is the outcome of one adapter call for one DomainRef.
type TLSResult struct {
	Ref       DomainRef
	State     TLSState
	NotAfter  time.Time // set when State is ready (or expired, when known)
	Reference string    // a provider resource created by Request (E2); empty under E1
}

// Errors of the package.
var (
	// ErrUnknownHost: the host does not currently route to any Organization (malformed, reserved, unknown, removed or
	// not routing-ready). It is never returned for a store failure.
	ErrUnknownHost = errors.New("domains: unknown host")
	// ErrStateInvalid: the operation does not apply to the row's current state (zero effects).
	ErrStateInvalid = errors.New("domains: domain state invalid")
	// ErrChallengeExpired: a pending claim's challenge has expired; it must be rotated before it can verify.
	ErrChallengeExpired = errors.New("domains: challenge expired")
)
