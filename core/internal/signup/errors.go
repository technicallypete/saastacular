package signup

import "errors"

// Code is the closed signup error vocabulary (G062 Plan r5 §4.2). Codes never carry IDs, names, slugs, secrets or
// driver text.
type Code string

// The signup codes. invalid_request carries the offending Field (name or slug); slug_taken and intent_used are the
// conflict answers of provisioning; continuation_invalid is a tenant hop that matched no row; signup_context_mismatch a
// root proof whose credential is not the row's own; signup_unavailable a store failure or an unknown commit (an identical
// retry answers already_created).
const (
	CodeInvalidRequest      Code = "invalid_request"
	CodeUnauthenticated     Code = "unauthenticated"
	CodeCSRFFailed          Code = "csrf_failed"
	CodeSlugTaken           Code = "slug_taken"
	CodeIntentUsed          Code = "intent_used"
	CodeUserDisabled        Code = "user_disabled"
	CodeContinuationInvalid Code = "continuation_invalid"
	CodeContextMismatch     Code = "signup_context_mismatch"
	CodeUnconfigured        Code = "signup_unconfigured"
	CodeUnavailable         Code = "signup_unavailable"

	// CodeInvalidName is the pinned name rule's refusal (names_copy.go, identical to tenancy's); Provision and SuggestSlug
	// report it as invalid_request with Field "name".
	CodeInvalidName Code = "invalid_name"
)

// The invalid_request fields.
const (
	FieldName = "name"
	FieldSlug = "slug"
)

// Error is a signup refusal or failure.
type Error struct {
	Code  Code
	Field string
}

func (e *Error) Error() string {
	if e.Field != "" {
		return "signup: " + string(e.Code) + " (" + e.Field + ")"
	}
	return "signup: " + string(e.Code)
}

func newError(c Code) *Error { return &Error{Code: c} }

func invalid(field string) *Error { return &Error{Code: CodeInvalidRequest, Field: field} }

// CodeOf returns the code of a signup error, "" for nil; any other error is signup_unavailable.
func CodeOf(err error) Code {
	if err == nil {
		return ""
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return CodeUnavailable
}

// AsError maps any error onto the closed vocabulary; unknown errors become signup_unavailable.
func AsError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return newError(CodeUnavailable)
}
