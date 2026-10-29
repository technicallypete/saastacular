// Package signup is the platform root signup service of G062 (Plan r5 §1.3-§1.4, §3.4, §4.3): constrained provisioning
// of one Organization with its creator's org_owner binding through the provision writer's DEFINER body, the advisory
// slug suggestion, and the automatic continuation into the new tenant host (tenant challenge, root proof with the
// original signup credential, tenant completion through identity's pinned login). Since G057 (Plan r5 §3.2) the tenant
// host is the generated opaque platform label (label.go), never the slug. It takes no actor, role or tenant parameter:
// the user comes only from a resolved root signup credential (identity.SignupCredential), and every database context is
// identity's. It does not import tenancy (D-9).
package signup

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"github.com/jackc/pgx/v5"

	"core/internal/audit"
	"core/internal/hostname"
	"core/internal/identity"
)

// ContinuationTTL is the lifetime of one continuation generation's secrets (Plan r5 §1.4; the bodies' 120 s).
const ContinuationTTL = 120 * time.Second

// secretLen is the size of every continuation secret (R, C, N, P): 256 bits.
const secretLen = 32

// Config is the signup service's configuration.
type Config struct {
	Identity   *identity.Service
	Hostname   hostname.Config // the platform domain, port and reserved labels (target origins come from here only)
	RootOrigin string          // the exact platform root origin ("" disables signup)
	Now        func() time.Time
	LabelRand  io.Reader // the platform label entropy source (nil: crypto/rand; tests inject a deterministic reader)
}

// Service is the signup service.
type Service struct {
	cfg Config
	// afterProvision, when set (package tests only), runs after a committed provisioning and before the answer; an error
	// stands for a lost commit acknowledgement (SUP-7 (a)).
	afterProvision func() error
	// provisionStmt, when set (package unit tests only), stands for one provisioning transaction: it answers the row the
	// provisioning body would for the drawn label and its target origin (the label budget tests count the store calls).
	provisionStmt func(label, target string) (provisionRow, error)
}

// NewService returns the signup service for cfg.
func NewService(cfg Config) *Service {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Service{cfg: cfg}
}

// Configured reports whether root signup is enabled (identity, the root origin and the platform domain are set).
func (s *Service) Configured() bool {
	return s != nil && s.cfg.Identity != nil && s.cfg.RootOrigin != "" && s.cfg.Hostname.BaseDomain != ""
}

// RootOrigin is the exact platform root origin the service accepts root requests on ("" when root signup is disabled).
func (s *Service) RootOrigin() string {
	if s == nil {
		return ""
	}
	return s.cfg.RootOrigin
}

func (s *Service) now() time.Time { return s.cfg.Now().UTC() }

// The status of a resolved signup credential (GET /signup/status).
const (
	StatusPending   = "pending"
	StatusCreated   = "created"
	StatusCompleted = "completed"
)

// StatusOf is the closed status of a resolved credential: pending until provisioned, created while a continuation is
// outstanding, completed once the current generation completed.
func StatusOf(info identity.SignupCredentialInfo) string {
	switch {
	case !info.Provisioned:
		return StatusPending
	case info.Completed:
		return StatusCompleted
	}
	return StatusCreated
}

// Resolve resolves a forwarded root signup credential at the request's exact origin (identity.SignupCredential): a
// malformed, unknown, expired, revoked or foreign-origin credential is unauthenticated.
func (s *Service) Resolve(ctx context.Context, credential, origin string) (identity.SignupCredentialInfo, error) {
	if !s.Configured() {
		return identity.SignupCredentialInfo{}, newError(CodeUnconfigured)
	}
	info, err := s.cfg.Identity.SignupCredential(ctx, credential, origin)
	if err != nil {
		return identity.SignupCredentialInfo{}, identityError(err)
	}
	return info, nil
}

// Suggestion is the advisory slug answer: the first candidate no Organization holds now, and whether one was found (an
// empty base means the name has no ASCII letter or digit and an explicit slug is required).
type Suggestion struct {
	Suggested string
	Available bool
}

// SuggestSlug derives the advisory slug for name (Plan r5 §4.3): the pinned base and candidate rule, skipping reserved
// labels, with availability read through the app.signup_slug_taken reader (G057 D-3: the slug is metadata, so its
// availability is not a routing question) in the credential's IDENTITY shape. Nothing is reserved.
func (s *Service) SuggestSlug(ctx context.Context, info identity.SignupCredentialInfo, name string) (Suggestion, error) {
	stored, err := ValidateName(name)
	if err != nil {
		return Suggestion{}, invalid(FieldName)
	}
	base := suggestBase(stored)
	if base == "" {
		return Suggestion{}, nil
	}
	var out Suggestion
	err = s.cfg.Identity.SignupIdentityTx(ctx, info.Handle, func(tx pgx.Tx) error {
		for _, c := range slugCandidates(base) {
			if !ValidSlug(c, s.cfg.Hostname) {
				continue
			}
			taken, err := slugTaken(ctx, tx, c)
			if err != nil {
				return err
			}
			if !taken {
				out = Suggestion{Suggested: c, Available: true}
				return nil
			}
		}
		out = Suggestion{Suggested: base}
		return nil
	})
	if err != nil {
		return Suggestion{}, identityError(err)
	}
	return out, nil
}

// Provisioned is a successful provisioning answer: created (generation 1) or already_created (the identical intent
// re-minted generation g+1). Continue is the tenant continuation URL carrying the generation's initial reference R, on
// the Organization's stored target origin (its platform label host).
type Provisioned struct {
	Outcome        string
	OrganizationID string
	Slug           string
	Generation     int
	Continue       string
}

// Intent is the create-intent digest SHA-256(name || 0x00 || slug) of a validated name and slug.
func Intent(name, slug string) []byte {
	h := sha256.New()
	h.Write([]byte(name))
	h.Write([]byte{0})
	h.Write([]byte(slug))
	return h.Sum(nil)
}

// TargetOrigin is the canonical tenant origin of a platform label: https://<label>.<PLATFORM_BASE_DOMAIN>
// [:PLATFORM_HOST_PORT], from configuration only. It is never derived from the friendly slug (G057 Plan r5 §3.2).
func (s *Service) TargetOrigin(label string) string {
	return hostname.Host{Name: label + "." + s.cfg.Hostname.BaseDomain, Port: s.cfg.Hostname.Port}.Origin()
}

// Provision creates the Organization named name with slug for the credential's user, or re-mints the continuation of
// an identical earlier intent (Plan r5 §1.3, §3.4): one IDENTITY transaction calling app.signup_provision_organization.
// A changed payload is intent_used, a taken slug slug_taken (nothing retained), a disabled creator user_disabled, a
// credential that is no longer live unauthenticated; an unknown commit is signup_unavailable (an identical retry then
// answers already_created).
//
// G057 (Plan r5 §3.2): each attempt draws a fresh platform label (GenerateLabel) and passes it with its target origin;
// a label reserved under the active configuration, or equal to the submitted slug (the web continuation guard refuses a host equal to the Organization's own slug), is redrawn without any statement, and a label_taken outcome (another
// Organization holds the label; nothing retained) is redrawn and retried in a new transaction. Both share one budget
// of maxLabelDraws draws per call; exhaustion is signup_unavailable. The label is not part of the intent: an identical
// retry answers already_created with the stored target origin, whatever label that retry drew.
func (s *Service) Provision(ctx context.Context, info identity.SignupCredentialInfo, name, slug string) (Provisioned, error) {
	if !s.Configured() {
		return Provisioned{}, newError(CodeUnconfigured)
	}
	stored, err := ValidateName(name)
	if err != nil {
		return Provisioned{}, invalid(FieldName)
	}
	if !ValidSlug(slug, s.cfg.Hostname) {
		return Provisioned{}, invalid(FieldSlug)
	}
	ref, err := newSecret()
	if err != nil {
		return Provisioned{}, newError(CodeUnavailable)
	}
	intent := Intent(stored, slug)
	for draw := 0; draw < maxLabelDraws; draw++ {
		label, err := GenerateLabel(s.labelReader())
		if err != nil {
			return Provisioned{}, newError(CodeUnavailable)
		}
		if label == slug || !s.labelIssuable(label) {
			continue // equal to the submitted slug (web refuses that host) or reserved under this configuration: redrawn before any statement
		}
		row, err := s.provisionOnce(ctx, info, stored, slug, label, intent, ref)
		if err != nil {
			return Provisioned{}, identityError(err)
		}
		if row.Outcome == outcomeLabelTaken {
			continue // the label index refused this draw; the body retained nothing
		}
		return s.provisioned(ctx, info, row, ref)
	}
	return Provisioned{}, newError(CodeUnavailable)
}

// provisionOnce runs one provisioning transaction for label (or the unit-test statement double).
func (s *Service) provisionOnce(ctx context.Context, info identity.SignupCredentialInfo, name, slug, label string, intent, ref []byte) (provisionRow, error) {
	target := s.TargetOrigin(label)
	if s.provisionStmt != nil {
		return s.provisionStmt(label, target)
	}
	var row provisionRow
	err := s.cfg.Identity.SignupIdentityTx(ctx, info.Handle, func(tx pgx.Tx) error {
		var err error
		row, err = provision(ctx, tx, info.Handle.Digest(), name, slug, label, intent, digest(ref), target, s.now())
		return err
	})
	return row, err
}

// provisioned maps one final provisioning row: created and already_created answer the continuation URL on the row's
// target origin (created: this call's label origin; already_created: the origin stored at creation, the only truth once
// labels are random, whatever label the retry drew); every refusal maps to its closed code.
func (s *Service) provisioned(ctx context.Context, info identity.SignupCredentialInfo, row provisionRow, ref []byte) (Provisioned, error) {
	switch row.Outcome {
	case OutcomeCreated, OutcomeAlreadyCreated:
	case outcomeIntentUsed:
		s.cfg.Identity.RecordSignupDenial(ctx, info.Handle, audit.OpSignupCreate, audit.ReasonIntentUsed)
		return Provisioned{}, newError(CodeIntentUsed)
	case outcomeSlugTaken:
		s.cfg.Identity.RecordSignupDenial(ctx, info.Handle, audit.OpSignupCreate, audit.ReasonSlugTaken)
		return Provisioned{}, newError(CodeSlugTaken)
	case outcomeUserDisabled:
		return Provisioned{}, newError(CodeUserDisabled)
	case outcomeSignupInvalid:
		s.cfg.Identity.RecordSignupDenial(ctx, info.Handle, audit.OpSignupCreate, audit.ReasonSignupInvalid)
		return Provisioned{}, newError(CodeUnauthenticated)
	default:
		return Provisioned{}, newError(CodeUnavailable)
	}
	if row.TargetOrigin == "" {
		return Provisioned{}, newError(CodeUnavailable) // fail closed: no continuation without a stored origin
	}
	if s.afterProvision != nil {
		if err := s.afterProvision(); err != nil {
			return Provisioned{}, newError(CodeUnavailable)
		}
	}
	return Provisioned{Outcome: row.Outcome, OrganizationID: row.OrganizationID, Slug: row.Slug, Generation: row.Generation,
		Continue: row.TargetOrigin + "/signup/continue?ref=" + encode(ref)}, nil
}

// Challenge is a tenant challenge answer: C for the host-only tenant challenge cookie, its lifetime, and the root proof
// URL carrying N.
type Challenge struct {
	Value  string
	MaxAge int
	Prove  string
}

// Challenge is continuation step 2 on the tenant host (Plan r5 §1.4): in the LOGIN shape of the browser's binding, the
// exact origin resolves to its Organization at the runtime, and app.signup_challenge_issue consumes R once for that
// origin and Organization, storing the digests of fresh C and N. Any mismatch, replay or expiry is continuation_invalid.
// No credential is consulted.
func (s *Service) Challenge(ctx context.Context, origin, binding, ref string) (Challenge, error) {
	if !s.Configured() {
		return Challenge{}, newError(CodeUnconfigured)
	}
	r, ok := decode(ref)
	if !ok || origin == "" || len(origin) > 300 {
		return Challenge{}, newError(CodeContinuationInvalid)
	}
	c, err := newSecret()
	if err != nil {
		return Challenge{}, newError(CodeUnavailable)
	}
	n, err := newSecret()
	if err != nil {
		return Challenge{}, newError(CodeUnavailable)
	}
	err = s.cfg.Identity.SignupContinuationTx(ctx, origin, binding, func(tx pgx.Tx, org string) error {
		return challengeIssue(ctx, tx, digest(r), origin, org, digest(c), digest(n), s.now())
	})
	if err != nil {
		return Challenge{}, s.hopDenial(ctx, audit.OpSignupChallenge, err)
	}
	return Challenge{Value: encode(c), MaxAge: int(ContinuationTTL / time.Second), Prove: s.cfg.RootOrigin + "/signup/prove?n=" + encode(n)}, nil
}

// Prove is continuation step 3 on the root host (Plan r5 §1.4, §3.4): the browser's original signup credential spends
// its own row's nonce and receives the proof P; a credential that is not the row's own (another user, another row,
// revoked, expired) or a spent/expired nonce matches no row and consumes nothing (signup_context_mismatch). The answer
// is the tenant completion URL carrying P.
func (s *Service) Prove(ctx context.Context, info identity.SignupCredentialInfo, nonce string) (string, error) {
	if !s.Configured() {
		return "", newError(CodeUnconfigured)
	}
	n, ok := decode(nonce)
	if !ok || !info.Provisioned || info.TargetOrigin == "" {
		return "", newError(CodeContextMismatch) // refused before any statement (not recorded)
	}
	p, err := newSecret()
	if err != nil {
		return "", newError(CodeUnavailable)
	}
	err = s.cfg.Identity.SignupIdentityTx(ctx, info.Handle, func(tx pgx.Tx) error {
		return prove(ctx, tx, info.Handle.Digest(), digest(n), digest(p), s.now())
	})
	if err != nil {
		if CodeOf(err) == CodeContextMismatch {
			s.cfg.Identity.RecordSignupDenial(ctx, info.Handle, audit.OpSignupProve, audit.ReasonSignupContextMismatch)
		}
		return "", identityError(err)
	}
	return info.TargetOrigin + "/signup/complete?p=" + encode(p), nil
}

// Complete is continuation step 4 on the tenant host (Plan r5 §1.4): identity.LoginPinned completes the generation only
// when C (from the host-only tenant challenge cookie) and P belong to the same row for this origin and its runtime
// Organization, then starts the pinned login (expected user, current membership). The answer is the ordinary login
// start (authorization URL and binding).
func (s *Service) Complete(ctx context.Context, origin, binding, challenge, proof string) (identity.LoginOutput, error) {
	if !s.Configured() {
		return identity.LoginOutput{}, newError(CodeUnconfigured)
	}
	c, okC := decode(challenge)
	p, okP := decode(proof)
	if !okC || !okP { // refused before any statement (not recorded)
		return identity.LoginOutput{}, newError(CodeContinuationInvalid)
	}
	out, err := s.cfg.Identity.LoginPinned(ctx, identity.PinnedLoginInput{Origin: origin, Binding: binding, Challenge: digest(c), Proof: digest(p)})
	if err != nil {
		return identity.LoginOutput{}, s.hopDenial(ctx, audit.OpSignupComplete, err)
	}
	return out, nil
}

// Abandon revokes the caller's own signup credential and with it the current continuation generation.
func (s *Service) Abandon(ctx context.Context, info identity.SignupCredentialInfo) error {
	if !s.Configured() {
		return newError(CodeUnconfigured)
	}
	err := s.cfg.Identity.SignupIdentityTx(ctx, info.Handle, func(tx pgx.Tx) error {
		return abandon(ctx, tx, info.Handle.Digest(), s.now())
	})
	if err != nil {
		return identityError(err)
	}
	return nil
}

// hopDenial maps a tenant hop's failure: a moved or non-routing host is host_changed and every other refusal is
// continuation_invalid (both recorded unattributed); a store failure is signup_unavailable (not recorded).
func (s *Service) hopDenial(ctx context.Context, op audit.Operation, err error) error {
	var ie *identity.Error
	if errors.As(err, &ie) {
		switch ie.Code {
		case identity.ErrHostChanged.Code:
			s.cfg.Identity.RecordSignupDenial(ctx, identity.SignupHandle{}, op, audit.ReasonHostChanged)
			return newError(CodeContinuationInvalid)
		case identity.ErrContinuationInvalid.Code, identity.ErrNotMember.Code, identity.ErrOriginRejected.Code:
			s.cfg.Identity.RecordSignupDenial(ctx, identity.SignupHandle{}, op, audit.ReasonContinuationInvalid)
			return newError(CodeContinuationInvalid)
		}
		return newError(CodeUnavailable)
	}
	if CodeOf(err) == CodeContinuationInvalid {
		s.cfg.Identity.RecordSignupDenial(ctx, identity.SignupHandle{}, op, audit.ReasonContinuationInvalid)
	}
	return identityError(err)
}

// identityError maps identity's errors onto the signup vocabulary (a signup error passes through).
func identityError(err error) error {
	var se *Error
	if errors.As(err, &se) {
		return se
	}
	var ie *identity.Error
	if errors.As(err, &ie) {
		switch ie.Code {
		case identity.ErrUnauthenticated.Code:
			return newError(CodeUnauthenticated)
		case identity.ErrUserDisabled.Code:
			return newError(CodeUserDisabled)
		case identity.ErrContinuationInvalid.Code, identity.ErrHostChanged.Code, identity.ErrNotMember.Code:
			return newError(CodeContinuationInvalid)
		case identity.ErrUnconfigured.Code:
			return newError(CodeUnconfigured)
		}
	}
	return newError(CodeUnavailable)
}

func newSecret() ([]byte, error) {
	b := make([]byte, secretLen)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func digest(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}

func encode(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// decode accepts exactly one canonical base64url encoding of exactly secretLen bytes.
func decode(s string) ([]byte, bool) {
	if len(s) != base64.RawURLEncoding.EncodedLen(secretLen) {
		return nil, false
	}
	b, err := base64.RawURLEncoding.Strict().DecodeString(s)
	if err != nil || len(b) != secretLen || encode(b) != s {
		return nil, false
	}
	return b, true
}
