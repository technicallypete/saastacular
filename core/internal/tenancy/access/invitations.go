package access

import (
	"context"
	"time"

	"core/internal/identity"
	"core/internal/tenancy"
)

// G019 invitations (Plan r1 §1.4 "Invitations" and "Consumption"; r2 addendum §1–§2). Issue and revoke are scope
// administration (organization.manage_members or group.manage_members); a recipient reads and accepts its own Group
// invitation without a capability; the Organization transition is the bounded consumer ConsumeOrganizationInvitation,
// which takes identity-owned address evidence and no tenant context. The raw token is returned once by IssueInvitation
// and is never stored, logged or audited.

// Recipient is an invitation's recipient: a current Organization member for a Group invitation (UserID), a normalized
// address for an Organization invitation (Address). Exactly one is set.
type Recipient struct {
	UserID  *tenancy.UserID
	Address *string
}

// Invitation is one invitation as its reader may see it: IntendedAddress only in the issuing scope's administrator
// views; never the token digest. Status is derived from the timestamps (pending, expired, revoked or consumed).
// IssuerDisplayName is optional (D-A3, Result addendum A1-01 §4): nil whenever no authorized reader supplies it (no
// reader is added for an optional label); consumers show a truthful fallback, never a guessed name.
type Invitation struct {
	ID                tenancy.InvitationID
	Scope             Scope
	Role              string
	IssuerDisplayName *string
	IntendedUserID    *tenancy.UserID
	IntendedAddress   *string
	ExpiresAt         time.Time
	Status            string
	ConsumedBy        *tenancy.UserID
}

// InvitationPage is one keyset page of invitations; Next is the cursor of the following page (nil on the last page).
type InvitationPage struct {
	Invitations []Invitation
	Next        *tenancy.InvitationID
}

// IssuedInvitation is a new invitation and its raw token, returned once.
type IssuedInvitation struct {
	Invitation Invitation
	Token      string
}

// AcceptOutcome is a consumption's outcome: joined, upgraded or already_member with the resulting binding, or consumed
// (the token was already used by this recipient and no qualifying binding is current; nothing is restored) with none.
type AcceptOutcome struct {
	Outcome string
	Binding *tenancy.Binding
}

// AcceptConsumed is the consumed replay outcome (no binding).
const AcceptConsumed = "consumed"

// Candidate is one Group invitation candidate: a current Organization member without a binding in the Group.
type Candidate struct {
	UserID      tenancy.UserID
	DisplayName *string
}

// CandidatePage is one keyset page of candidates.
type CandidatePage struct {
	Candidates []Candidate
	Next       *tenancy.UserID
}

// fieldEvidence is the invalid_input field of a zero or malformed VerifiedAddress (refused before any statement).
const fieldEvidence = "evidence"

// invitationView maps a stored invitation for its reader; admin keeps the intended address.
func invitationView(inv tenancy.Invitation, admin bool, now time.Time) Invitation {
	v := Invitation{ID: inv.ID, Scope: Scope{Kind: inv.ScopeType, Group: inv.GroupID}, Role: inv.Role,
		IntendedUserID: inv.IntendedUserID, ExpiresAt: inv.ExpiresAt, Status: inv.Status(now), ConsumedBy: inv.ConsumedByUserID}
	if admin {
		v.IntendedAddress = inv.IntendedAddress
	}
	return v
}

func invitationPage(rows []tenancy.Invitation, limit int, admin bool) InvitationPage {
	var out InvitationPage
	if len(rows) > limit {
		next := rows[limit-1].ID
		out.Next, rows = &next, rows[:limit]
	}
	now := time.Now()
	out.Invitations = make([]Invitation, 0, len(rows))
	for _, r := range rows {
		out.Invitations = append(out.Invitations, invitationView(r, admin, now))
	}
	return out
}

// ListInvitations reads one keyset page of the scope's invitations, all statuses, with intended addresses
// (organization.manage_members or group.manage_members). limit 0 is MaxMembersPage.
func (s *Service) ListInvitations(ctx context.Context, tc *TenantContext, sc Scope, after string, limit int) (InvitationPage, error) {
	var out InvitationPage
	var from tenancy.InvitationID
	pre := scopePre(sc, func() *failure {
		var f *failure
		if limit, f = pageLimit(limit); f != nil {
			return f
		}
		from, f = cursor(after, tenancy.ParseInvitationID)
		return f
	})
	err := s.guard(ctx, tc, opInvitationList, sc.manageMembers(), false, sc.target(), sc.query(false, nil, false, false), pre,
		func(tx *trackedTx, f Facts) *failure {
			rows, err := tenancy.InvitationsPage(ctx, tx, f.Organization.ID, sc.group(), from, limit)
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			out = invitationPage(rows, limit, true)
			return nil
		})
	if err != nil {
		return InvitationPage{}, err
	}
	return out, nil
}

// ListMyInvitations reads one keyset page of the subject's own open invitations in the active Organization (no
// capability), never with an intended address. limit 0 is MaxMembersPage.
func (s *Service) ListMyInvitations(ctx context.Context, tc *TenantContext, after string, limit int) (InvitationPage, error) {
	var out InvitationPage
	var from tenancy.InvitationID
	pre := func() *failure {
		var f *failure
		if limit, f = pageLimit(limit); f != nil {
			return f
		}
		from, f = cursor(after, tenancy.ParseInvitationID)
		return f
	}
	err := s.guard(ctx, tc, opInvitationMine, noCapability, false, Target{Kind: kindOrganization}, organizationQuery, pre,
		func(tx *trackedTx, f Facts) *failure {
			rows, err := tenancy.MyInvitationsPage(ctx, tx, f.Organization.ID, tc.subject, from, limit)
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			out = invitationPage(rows, limit, false)
			return nil
		})
	if err != nil {
		return InvitationPage{}, err
	}
	return out, nil
}

// ListInvitationCandidates reads one keyset page of the Group recipient picker (group.manage_members): current
// Organization members not yet in the Group, identifiers and display names only (D-13). limit 0 is MaxMembersPage.
func (s *Service) ListInvitationCandidates(ctx context.Context, tc *TenantContext, group tenancy.GroupID, after string, limit int) (CandidatePage, error) {
	var out CandidatePage
	var from tenancy.UserID
	pre := func() *failure {
		if f := targetID(group)(); f != nil {
			return f
		}
		var f *failure
		if limit, f = pageLimit(limit); f != nil {
			return f
		}
		from, f = cursor(after, tenancy.ParseUserID)
		return f
	}
	err := s.guard(ctx, tc, opInvitationCandidates, capGroupManageMembers, false, Target{Kind: kindGroup, Group: group},
		groupQuery(group, false), pre,
		func(tx *trackedTx, _ Facts) *failure {
			rows, err := tenancy.InvitationCandidatesPage(ctx, tx, group, from, limit)
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			if len(rows) > limit {
				next := rows[limit-1].UserID
				out.Next, rows = &next, rows[:limit]
			}
			out.Candidates = make([]Candidate, 0, len(rows))
			for _, r := range rows {
				out.Candidates = append(out.Candidates, Candidate{UserID: r.UserID, DisplayName: r.DisplayName})
			}
			return nil
		})
	if err != nil {
		return CandidatePage{}, err
	}
	return out, nil
}

// IssueInvitation issues an admin or member invitation of the scope (organization.manage_members or
// group.manage_members): a Group invitation names a current Organization member, an Organization invitation a normalized
// address. lifetime 0 is the default (7 days); 1 hour to 30 days otherwise. The scope row is locked at exclusive
// strength before the pending count, so the bound (100 pending unexpired per scope) is serialized per scope
// (invitation_limit); the insert trigger repeats it. The raw token is returned once.
func (s *Service) IssueInvitation(ctx context.Context, tc *TenantContext, sc Scope, role string, r Recipient, lifetime time.Duration) (IssuedInvitation, error) {
	var out IssuedInvitation
	pre := scopePre(sc, func() *failure {
		if !tenancy.InvitationRoleValid(sc.Kind, role) {
			return invalid(string(tenancy.CodeInvalidRole))
		}
		switch sc.Kind {
		case tenancy.ScopeGroup:
			if r.UserID == nil || zero(*r.UserID) || r.Address != nil {
				return invalid(fieldRecipient)
			}
			// G060 (r1 C3): nobody invites itself into a Group, under any authority; its own Group role comes only from
			// another authorized issuer (the claim body also never consumes such an invitation).
			if tc != nil && *r.UserID == tc.subject {
				return invalid(fieldSelfTarget)
			}
		default:
			if r.UserID != nil || r.Address == nil {
				return invalid(fieldRecipient)
			}
			if _, err := tenancy.NormalizeAddress(*r.Address); err != nil {
				return inputFailure(err, reasonInputInvalid)
			}
		}
		if lifetime != 0 && (lifetime < tenancy.InvitationMinLifetime || lifetime > tenancy.InvitationMaxLifetime) {
			return invalid(string(tenancy.CodeInvalidLifetime))
		}
		return nil
	})
	err := s.guard(ctx, tc, opInvitationIssue, sc.manageMembers(), true, sc.target(), sc.query(true, nil, false, false), pre,
		func(tx *trackedTx, f Facts) *failure {
			n, err := tenancy.PendingInvitationCount(ctx, tx, f.Organization.ID, sc.group())
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			if n >= tenancy.MaxPendingInvitations {
				return conflict(string(tenancy.CodeInvitationLimit))
			}
			id, ok := newKey[tenancy.InvitationID]()
			if !ok {
				return fail(CodeUnavailable, reasonStoreUnavailable)
			}
			spec := tenancy.InvitationSpec{ID: id, OrganizationID: f.Organization.ID, Group: sc.group(), Role: role,
				Issuer: tc.subject, Lifetime: lifetime}
			if r.UserID != nil {
				spec.IntendedUser = *r.UserID
			}
			if r.Address != nil {
				spec.IntendedAddress = *r.Address
			}
			inv, token, err := tenancy.InsertInvitationTx(ctx, tx, spec)
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			out = IssuedInvitation{Invitation: invitationView(inv, true, time.Now()), Token: token}
			return s.scopeOverrideUsed(ctx, tx, tc, f, sc, opInvitationIssue)
		})
	if err != nil {
		return IssuedInvitation{}, err
	}
	return out, nil
}

// RevokeInvitation revokes a pending invitation (the manage_members capability of the invitation's scope, decided after
// the row is read). A row the subject cannot see or manage is not_found; one no longer pending is conflict
// (invitation_not_pending).
func (s *Service) RevokeInvitation(ctx context.Context, tc *TenantContext, id tenancy.InvitationID) error {
	return s.guardEvaluated(ctx, tc, opInvitationRevoke, true, Target{Kind: kindOrganization}, organizationQuery, targetID(id),
		func(tx *trackedTx, f Facts, evaluated *evaluation) *failure {
			inv, err := tenancy.InvitationTx(ctx, tx, f.Organization.ID, id)
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			if fl := s.invitationAdmin(ctx, tx, tc, f, inv, true, evaluated); fl != nil {
				return fl
			}
			if _, err := tenancy.RevokeInvitationTx(ctx, tx, f.Organization.ID, id); err != nil {
				return membershipFailure(ctx, tx, err)
			}
			if inv.GroupID != nil && evaluated.reason == reasonAllowedOverride {
				return s.overrideUsed(ctx, tx, tc, evaluated.facts, capGroupManageMembers, Target{Kind: kindGroup, Group: *inv.GroupID},
					opInvitationRevoke, tenancy.GroupID{})
			}
			return nil
		})
}

// GetInvitation reads one invitation: the intended recipient sees its own invitation without the address; the issuing
// scope's administrators (the manage_members capability of its scope) see it with the address; anyone else not_found.
func (s *Service) GetInvitation(ctx context.Context, tc *TenantContext, id tenancy.InvitationID) (Invitation, error) {
	var out Invitation
	err := s.guardEvaluated(ctx, tc, opInvitationGet, false, Target{Kind: kindOrganization}, organizationQuery, targetID(id),
		func(tx *trackedTx, f Facts, evaluated *evaluation) *failure {
			inv, err := tenancy.InvitationTx(ctx, tx, f.Organization.ID, id)
			if err != nil {
				return membershipFailure(ctx, tx, err)
			}
			if inv.IntendedUserID != nil && *inv.IntendedUserID == tc.subject {
				out = invitationView(inv, false, time.Now())
				return nil
			}
			if fl := s.invitationAdmin(ctx, tx, tc, f, inv, false, evaluated); fl != nil {
				return fl
			}
			out = invitationView(inv, true, time.Now())
			return nil
		})
	if err != nil {
		return Invitation{}, err
	}
	return out, nil
}

// invitationAdmin decides the manage_members capability of an invitation's scope on the operation's facts; for a Group
// invitation it first reads that Group's facts in the same transaction (locked for a write, after the Organization and
// own-binding locks the facts already hold). A denial is the ordinary not_found/forbidden mapping, recorded.
func (s *Service) invitationAdmin(ctx context.Context, tx *trackedTx, tc *TenantContext, f Facts, inv tenancy.Invitation, write bool,
	evaluated *evaluation) *failure {
	sel := tc.selection()
	c, t := capOrganizationManageMembers, Target{Kind: kindOrganization, Organization: f.Organization.ID}
	if inv.GroupID != nil {
		g := *inv.GroupID
		c, t = capGroupManageMembers, Target{Kind: kindGroup, Organization: f.Organization.ID, Group: g}
		groups := []tenancy.GroupID{g}
		if !zero(sel.Group) && sel.Group != g {
			groups = append(groups, sel.Group)
		}
		var err error
		if f, err = tenancy.ReadFacts(ctx, tx, tenancy.FactQuery{Organization: f.Organization.ID, User: tc.subject, Groups: groups, Lock: write}); err != nil {
			return fail(CodeUnavailable, failureReason(ctx, tx.cause))
		}
	}
	evaluated.capability = c // D-A2: the decision below is the one the access.decision record reports
	ok, r := decide(c, sel, f, t)
	if !ok {
		res, id, g := targetScope(t)
		return markDenied(denial(c, sel, f, t, r), sel, f, res, id, g)
	}
	evaluated.reason, evaluated.facts = r, f
	return nil
}

// evaluation is what guardEvaluated's act decided: the capability the access.decision record reports, its allow reason
// (G060: reasonAllowedOverride for the Organization override path) and the facts it was decided on.
type evaluation struct {
	capability Capability
	reason     Reason
	facts      Facts
}

// guardEvaluated is guard for an operation whose capability is known only after its row is read (revoke and get;
// D-A2, Result addendum A1-01 §3): the one access.decision record carries the capability actually evaluated
// (organization.manage_members or group.manage_members), or none for a recipient-only read or a call that ended before
// that decision (invalid id, context or session failure, invitation not found, Group facts unavailable).
func (s *Service) guardEvaluated(ctx context.Context, tc *TenantContext, op string, write bool, t Target,
	query func(*TenantContext) tenancy.FactQuery, pre func() *failure, act func(tx *trackedTx, f Facts, evaluated *evaluation) *failure) error {
	var evaluated evaluation
	f := s.guarded(ctx, tc, noCapability, write, t, query, pre, func(tx *trackedTx, facts Facts) *failure {
		return act(tx, facts, &evaluated)
	})
	s.recordGuarded(ctx, tc, op, f)
	s.logGuardedAllow(ctx, op, evaluated.capability, evaluated.reason, f)
	if f != nil {
		return f.err()
	}
	return nil
}

// acceptOutcome maps a claim to its outcome and resulting binding.
func acceptOutcome(user tenancy.UserID, c tenancy.ClaimOutcome) AcceptOutcome {
	return AcceptOutcome{Outcome: c.Outcome, Binding: &tenancy.Binding{ID: c.BindingID, OrganizationID: c.OrganizationID,
		SubjectUserID: user, ScopeType: c.ScopeType, ScopeID: c.ScopeID, GroupID: c.GroupID, Role: c.Role,
		Source: tenancy.SourceMembership}}
}

// claimFailure maps a claim refusal: a consumed replay without a qualifying binding is the consumed outcome (no error,
// nothing written); everything else is membershipFailure (one closed not_found for every mismatch). Consumed requires
// the local status mapping (a successful claim Scan, tx.cause nil): a raised SQL error that maps to the same code
// (23000 invitations_keys_immutable) has aborted the transaction, so it falls through to the conflict and its rollback.
func claimFailure(ctx context.Context, tx *trackedTx, err error) (bool, *failure) {
	if tenancy.CodeOf(err) == tenancy.CodeInvitationNotPending && tx.cause == nil {
		return true, nil
	}
	return false, membershipFailure(ctx, tx, err)
}

// tokenPre validates the invitation id and the token's form before any statement.
func tokenPre(id tenancy.InvitationID, token string) func() *failure {
	return func() *failure {
		if f := targetID(id)(); f != nil {
			return f
		}
		if !tenancy.ValidInvitationToken(token) {
			return invalid(string(tenancy.CodeInvalidToken))
		}
		return nil
	}
}

// AcceptGroupInvitation consumes the subject's own Group invitation with its token (no capability; the recipient must
// be a current Organization member, re-enforced by app.invitation_claim_group, which also verifies the token, the
// recipient, the pending state, the issuer's current authority and the selected context). Every mismatch is not_found.
func (s *Service) AcceptGroupInvitation(ctx context.Context, tc *TenantContext, id tenancy.InvitationID, token string) (AcceptOutcome, error) {
	var out AcceptOutcome
	err := s.guard(ctx, tc, opInvitationAccept, noCapability, true, Target{Kind: kindOrganization}, organizationQuery, tokenPre(id, token),
		func(tx *trackedTx, _ Facts) *failure {
			c, err := tenancy.ClaimInvitationTx(ctx, tx, id, token, nil)
			if err != nil {
				consumed, fl := claimFailure(ctx, tx, err)
				if consumed {
					out = AcceptOutcome{Outcome: AcceptConsumed}
				}
				return fl
			}
			out = acceptOutcome(tc.subject, c)
			return nil
		})
	if err != nil {
		return AcceptOutcome{}, err
	}
	return out, nil
}

// ConsumeOrganizationInvitation is the bounded Organization-invitation consumer (Plan r1 §5.5; r2 addendum §1–§2): no
// tenant context and no session; the recipient is the canonical user identity resolved from (issuer, subject) when it
// attested ev, and the address reaches SQL only from ev. It runs in the IDENTITY-shape context of that user (no
// Organization) and calls app.invitation_claim_organization, which re-verifies the principal, the address, the token,
// the pending state and the issuer's current authority; in G019 only the owner role may execute it (D-15 pending), so a
// runtime-pool Service is refused (internal). A zero ev is refused before any statement. A not_found refusal is
// recorded as access.denied without an Organization; session issuance is not part of this call (G058).
func (s *Service) ConsumeOrganizationInvitation(ctx context.Context, ev identity.VerifiedAddress, id tenancy.InvitationID, token string) (AcceptOutcome, error) {
	var out AcceptOutcome
	var user tenancy.UserID
	f := func() *failure {
		if ev.IsZero() {
			return invalid(fieldEvidence)
		}
		var err error
		if user, err = tenancy.ParseUserID(ev.UserID()); err != nil || zero(user) {
			return invalid(fieldEvidence)
		}
		if f := tokenPre(id, token)(); f != nil {
			return f
		}
		claim := &tenancy.OrganizationClaim{Issuer: ev.Issuer(), Subject: ev.Subject(), Address: ev.Address()}
		return s.inContextTx(ctx, true, dbContext{user: user}, cleanupUnbounded, func(tx *trackedTx) *failure {
			c, err := tenancy.ClaimInvitationTx(ctx, tx, id, token, claim)
			if err != nil {
				consumed, fl := claimFailure(ctx, tx, err)
				if consumed {
					out = AcceptOutcome{Outcome: AcceptConsumed}
				}
				return fl
			}
			out = acceptOutcome(user, c)
			return nil
		})
	}()
	if f != nil && f.code == CodeNotFound && !zero(user) {
		f.denied = &denialScope{}
		s.recordDenial(ctx, user, tenancy.OrganizationID{}, opInvitationConsume, f)
	}
	s.logDecision(ctx, opInvitationConsume, noCapability, outcomeAllow, reasonAllowed, f)
	if f != nil {
		return AcceptOutcome{}, f.err()
	}
	return out, nil
}
