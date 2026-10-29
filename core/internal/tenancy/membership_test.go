package tenancy

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// G019 (Plan r1 §1.4 failure semantics; r2 §4): 00015's named guard and claim refusals map onto bounded codes; any
// other 23000/42501 stays internal; every class-40 abort (deadlock, serialization) stays unavailable, never a conflict.
func TestMembershipRefusalCodes(t *testing.T) {
	pg := func(code, constraint string) error {
		return fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: code, ConstraintName: constraint, Message: "secret-canary row"})
	}
	for _, c := range []struct {
		err  error
		want Code
	}{
		{pg("23000", "last_owner"), CodeLastOwner},
		{pg("23000", "recipient_ineligible"), CodeRecipientIneligible},
		{pg("23000", "dependent_bindings"), CodeDependentBindings},
		{pg("23000", "issuer_authority_lost"), CodeIssuerAuthorityLost},
		{pg("23000", "invitation_limit"), CodeInvitationLimit},
		{pg("23000", "invitations_keys_immutable"), CodeInvitationNotPending},
		{pg("42501", "owner_authority"), CodeOwnerAuthority},
		{pg("42501", "context_required"), CodeContextRequired},
		{pg("42501", "invitation_context"), CodeContextRequired},
		// G060 (00017): the override guard, step (3b) and (amendment01) the departure cascade's owner-subject refusal.
		{pg("42501", "group_override_owner"), CodeOwnerAuthority},
		{pg("42501", "self_escalation"), CodeSelfEscalation},
		{pg("42501", "owner_subject"), CodeOwnerSubject},
		{pg("23000", "owner_subject"), CodeInternal}, // the class must match too
		// Unnamed or foreign refusals of the same classes stay internal.
		{pg("23000", "groups_keys_immutable"), CodeInternal},
		{pg("23000", "role_bindings_keys_immutable"), CodeInternal},
		{pg("23000", "legacy_owner_preflight"), CodeInternal},
		{pg("23000", ""), CodeInternal},
		{pg("42501", ""), CodeInternal},
		{pg("42501", "audit_event_scope"), CodeInternal},
		{pg("42501", "last_owner"), CodeInternal},      // the class must match too
		{pg("23000", "owner_authority"), CodeInternal}, // the class must match too
		// Class 40 (r2 §4): a whole-transaction abort, answered as unavailable.
		{pg("40P01", ""), CodeUnavailable},
		{pg("40001", ""), CodeUnavailable},
		{pg("40P01", "last_owner"), CodeUnavailable},
		// The invitation table's named constraints.
		{pg("23503", "invitations_organization_fkey"), CodeOrganizationNotFound},
		{pg("23503", "invitations_group_fkey"), CodeGroupNotFound},
		{pg("23503", "invitations_issuer_fkey"), CodeUserNotFound},
		{pg("23503", "invitations_intended_user_fkey"), CodeUserNotFound},
		{pg("23503", "invitations_consumed_by_fkey"), CodeUserNotFound},
		{pg("23514", "invitations_id_nonzero"), CodeInvalidID},
		{pg("23514", "invitations_scope_consistent"), CodeInvalidRole},
		{pg("23514", "invitations_address_normalized"), CodeInvalidAddress},
		{pg("23514", "invitations_lifetime"), CodeInvalidLifetime},
		{pg("23514", "invitations_token_digest_length"), CodeInternal},
		{pg("23505", "invitations_token_digest_key"), CodeInternal},
	} {
		e := classify(c.err)
		if e == nil || e.Code != c.want {
			t.Errorf("classify(%v) = %v, want %s", c.err, e, c.want)
			continue
		}
		if e.Error() != "tenancy: "+string(c.want) {
			t.Errorf("Error() = %q", e.Error())
		}
	}
	for _, c := range []Code{CodeLastOwner, CodeOwnerAuthority, CodeRecipientIneligible, CodeDependentBindings, CodeContextRequired,
		CodeIssuerAuthorityLost, CodeInvitationLimit, CodeInvitationNotPending, CodeInvitationNotFound, CodeSelfEscalation, CodeOwnerSubject} {
		if !rejected(c) {
			t.Errorf("%s counted as a failure, want a rejection", c)
		}
	}
}

// The membership primitives validate before any statement: a nil transaction is never touched by a rejected call.
func TestMembershipRejectsBeforeSQL(t *testing.T) {
	ctx := t.Context()
	org, _ := ParseOrganizationID("0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	u, _ := ParseUserID("1f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	g, _ := ParseGroupID("2f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	id, _ := ParseBindingID("3f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	var zu UserID
	orgBinding := Binding{ID: id, OrganizationID: org, ScopeType: ScopeOrganization, ScopeID: org}
	groupBinding := Binding{ID: id, OrganizationID: org, ScopeType: ScopeGroup, ScopeID: g, GroupID: &g}
	s := NewService(nil, nil)
	for _, c := range []struct {
		name string
		err  error
		want Code
	}{
		{"change zero binding", second(ChangeRoleTx(ctx, nil, Binding{OrganizationID: org}, string(OrgAdmin))), CodeInvalidID},
		{"change zero org", second(ChangeRoleTx(ctx, nil, Binding{ID: id}, string(OrgAdmin))), CodeInvalidID},
		{"change group role on org binding", second(ChangeRoleTx(ctx, nil, orgBinding, string(GroupAdmin))), CodeInvalidRole},
		{"change org role on group binding", second(ChangeRoleTx(ctx, nil, groupBinding, string(OrgOwner))), CodeInvalidRole},
		{"change empty role", second(ChangeRoleTx(ctx, nil, orgBinding, "")), CodeInvalidRole},
		{"change unknown scope", second(ChangeRoleTx(ctx, nil, Binding{ID: id, OrganizationID: org, ScopeType: "tenant"}, string(OrgAdmin))), CodeInvalidRole},
		{"remove zero binding", RemoveBindingTx(ctx, nil, Binding{OrganizationID: org}), CodeInvalidID},
		{"owner count bad scope", second(OwnerCount(ctx, nil, "tenant", org)), CodeInvalidRole},
		{"owner count zero scope", second(OwnerCount(ctx, nil, ScopeGroup, [16]byte{})), CodeInvalidID},
		{"members zero org", second(MembersPage(ctx, nil, OrganizationID{}, GroupID{}, zu, 10, false)), CodeInvalidID},
		{"members limit 0", second(MembersPage(ctx, nil, org, GroupID{}, zu, 0, false)), CodeInvalidID},
		{"members limit 101", second(MembersPage(ctx, nil, org, g, zu, MaxMembersPage+1, true)), CodeInvalidID},
		{"candidates zero group", second(InvitationCandidatesPage(ctx, nil, GroupID{}, zu, 10)), CodeInvalidID},
		{"candidates limit 101", second(InvitationCandidatesPage(ctx, nil, g, zu, MaxMembersPage+1)), CodeInvalidID},
		{"facts subject without lock", second(ReadFacts(ctx, nil, FactQuery{Organization: org, User: u, Subject: &u})), CodeInvalidID},
		{"facts zero subject", second(ReadFacts(ctx, nil, FactQuery{Organization: org, User: u, Subject: &zu, Lock: true})), CodeInvalidID},
		{"create with owner zero owner", third(s.CreateOrganizationWithOwner(ctx, "Acme", zu)), CodeInvalidID},
		{"create with owner bad name", third(s.CreateOrganizationWithOwner(ctx, " ", u)), CodeInvalidName},
	} {
		if got := CodeOf(c.err); got != c.want {
			t.Errorf("%s: %v, want %s", c.name, c.err, c.want)
		}
	}
}

// The Subject and OwnerCount facts address the target Group's scope when a target is set, else the Organization's.
func TestFactScope(t *testing.T) {
	org, _ := ParseOrganizationID("0f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	g, _ := ParseGroupID("2f1e2d3c-4b5a-4978-8695-a4b3c2d1e0f9")
	if s, id := (FactQuery{Organization: org}).FactScope(); s != ScopeOrganization || id != org {
		t.Errorf("Organization fact scope = %s %x", s, id)
	}
	if s, id := (FactQuery{Organization: org, TargetGroup: g}).FactScope(); s != ScopeGroup || id != g {
		t.Errorf("Group fact scope = %s %x", s, id)
	}
	for _, c := range []struct {
		scope ScopeType
		role  string
		want  bool
	}{
		{ScopeOrganization, "org_owner", true}, {ScopeOrganization, "org_admin", true}, {ScopeOrganization, "org_member", true},
		{ScopeOrganization, "group_owner", false}, {ScopeGroup, "group_owner", true}, {ScopeGroup, "group_member", true},
		{ScopeGroup, "org_admin", false}, {"", "org_admin", false},
	} {
		if roleValid(c.scope, c.role) != c.want {
			t.Errorf("roleValid(%s, %s) != %v", c.scope, c.role, c.want)
		}
	}
}
