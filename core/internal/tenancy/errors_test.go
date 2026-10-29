package tenancy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// U-ERR: the constraint/SQLSTATE mapping table, driven from synthetic PostgreSQL errors, and the bounded error text.
func TestClassifyMapping(t *testing.T) {
	const detail = "Key (slug)=(secret-canary-slug) already exists."
	pg := func(code, constraint string) error {
		return fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: code, ConstraintName: constraint, Detail: detail,
			Message: "row value secret-canary-name"})
	}
	for _, c := range []struct {
		err  error
		want Code
	}{
		{pg("23505", "groups_organization_slug_key"), CodeSlugConflict},
		{pg("23505", "role_bindings_subject_scope_key"), CodeMembershipExists},
		{pg("23505", "groups_pkey"), CodeInternal},
		{pg("23503", "groups_organization_fkey"), CodeOrganizationNotFound},
		{pg("23503", "role_bindings_organization_fkey"), CodeOrganizationNotFound},
		{pg("23503", "role_bindings_group_fkey"), CodeGroupNotFound},
		{pg("23503", "role_bindings_subject_fkey"), CodeUserNotFound},
		{pg("23503", "sessions_user_id_fkey"), CodeInternal},
		{pg("23514", "organizations_name_valid"), CodeInvalidName},
		{pg("23514", "groups_name_valid"), CodeInvalidName},
		{pg("23514", "groups_slug_format"), CodeInvalidSlug},
		{pg("23514", "groups_slug_reserved"), CodeSlugReserved},
		{pg("23514", "organizations_id_nonzero"), CodeInvalidID},
		{pg("23514", "groups_id_nonzero"), CodeInvalidID},
		{pg("23514", "role_bindings_id_nonzero"), CodeInvalidID},
		{pg("23514", "role_bindings_subject_nonzero"), CodeInvalidID},
		{pg("23514", "role_bindings_scope_consistent"), CodeInvalidRole},
		{pg("23514", "users_check"), CodeInternal},
		// G006.
		{pg("23505", "group_business_profiles_group_key"), CodeAssociationExists},
		{pg("23505", "business_profiles_pkey"), CodeInternal},
		{pg("23503", "business_profiles_organization_fkey"), CodeOrganizationNotFound},
		{pg("23503", "group_business_profiles_organization_fkey"), CodeOrganizationNotFound},
		{pg("23503", "group_business_profiles_group_fkey"), CodeGroupNotFound},
		{pg("23503", "group_business_profiles_business_profile_fkey"), CodeBusinessProfileNotFound},
		{pg("23514", "business_profiles_display_name_valid"), CodeInvalidDisplayName},
		{pg("23514", "business_profiles_legal_name_valid"), CodeInvalidLegalName},
		{pg("23514", "business_profiles_description_valid"), CodeInvalidDescription},
		{pg("23514", "business_profiles_industry_valid"), CodeInvalidIndustry},
		{pg("23514", "business_profiles_website_valid"), CodeInvalidWebsite},
		{pg("23514", "business_profiles_id_nonzero"), CodeInvalidID},
		{pg("23514", "group_business_profiles_id_nonzero"), CodeInvalidID},
		{pg("23000", "business_profiles_keys_immutable"), CodeInternal},
		{pg("23000", "group_business_profiles_keys_immutable"), CodeInternal},
		{pg("08006", ""), CodeUnavailable},
		{pg("08P01", ""), CodeUnavailable},
		{pg("40001", ""), CodeUnavailable},
		{pg("40P01", ""), CodeUnavailable},
		{pg("53300", ""), CodeUnavailable},
		{pg("57014", ""), CodeUnavailable},
		{pg("57P01", ""), CodeUnavailable},
		{pg("42501", ""), CodeInternal},
		{pg("23000", "groups_keys_immutable"), CodeInternal},
		{pg("42P01", ""), CodeInternal},
		{context.Canceled, CodeUnavailable},
		{fmt.Errorf("x: %w", context.DeadlineExceeded), CodeUnavailable},
		{io.ErrUnexpectedEOF, CodeUnavailable},
		{&pgconn.ConnectError{}, CodeUnavailable},
		{errors.New("can't scan into dest"), CodeInternal},
		{newError(CodeSlugRequired), CodeSlugRequired},
	} {
		e := classify(c.err)
		if e == nil || e.Code != c.want {
			t.Errorf("classify(%v) = %v, want %s", c.err, e, c.want)
			continue
		}
		msg := e.Error()
		if msg != "tenancy: "+string(c.want) || strings.Contains(msg, "secret") {
			t.Errorf("Error() = %q", msg)
		}
		// Only context errors are exposed through Unwrap.
		if u := errors.Unwrap(e); u != nil && u != context.Canceled && u != context.DeadlineExceeded {
			t.Errorf("Unwrap exposes %v", u)
		}
	}
	if !errors.Is(classify(context.Canceled), context.Canceled) || !errors.Is(unavailable(context.DeadlineExceeded), context.DeadlineExceeded) {
		t.Fatal("context cause not exposed")
	}
	var pgErr *pgconn.PgError
	if errors.As(classify(pg("23505", "groups_organization_slug_key")), &pgErr) {
		t.Fatal("driver error exposed")
	}
	if classify(nil) != nil || CodeOf(errors.New("x")) != "" {
		t.Fatal("nil or foreign error")
	}
	for _, c := range []Code{CodeUnavailable, CodeInternal} {
		if rejected(c) {
			t.Errorf("%s counted as rejected", c)
		}
	}
	for _, c := range []Code{CodeInvalidID, CodeMembershipExists, CodeGroupNotFound, CodeSlugConflict, CodeAssociationExists,
		CodeAssociationNotFound, CodeBusinessProfileNotFound, CodeInvalidWebsite} {
		if !rejected(c) {
			t.Errorf("%s counted as failed", c)
		}
	}
}
