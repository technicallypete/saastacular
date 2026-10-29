package signup

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// The signup statements (G062 Plan r5 §3.4). Identity owns every transaction shape (the pinned request-context
// initialization lives only in identity and tenancy): these run inside identity.Service.SignupIdentityTx (the IDENTITY
// shape of a resolved signup credential's user) or SignupContinuationTx (the LOGIN shape of a tenant hop). Every secret
// reaches SQL as its SHA-256 digest only.
const (
	// sqlProvision is the provision writer's constrained provisioning body (IDENTITY shape; 00021: the platform label
	// is the fourth argument, and the answer carries the target origin: this call's for created, the stored one for
	// already_created).
	sqlProvision = `SELECT outcome, organization_id::text, slug, generation, target_origin ` +
		`FROM app.signup_provision_organization($1, $2, $3, $4, $5, $6, $7, $8)`
	// sqlChallengeIssue is the provision writer's tenant challenge body (LOGIN shape): it consumes R once.
	sqlChallengeIssue = `SELECT ok FROM app.signup_challenge_issue($1, $2, $3::uuid, $4, $5, $6)`
	// sqlProve is step 3, runtime SQL under signup_sessions_prove: the credential holder's own row whose nonce is
	// unconsumed and whose generation is live; anything else matches no row and consumes nothing.
	sqlProve = `UPDATE app.signup_sessions SET nonce_consumed_at = $1, proof_hash = $2 WHERE credential_hash = $3 ` +
		`AND user_id = app.ctx_user() AND nonce_hash = $4 AND nonce_consumed_at IS NULL AND continuation_expires_at > $1 ` +
		`AND revoked_at IS NULL`
	// sqlAbandon revokes the caller's own credential (and with it the current generation) under signup_sessions_revoke.
	sqlAbandon = `UPDATE app.signup_sessions SET revoked_at = $1, revoke_reason = 'abandoned' WHERE credential_hash = $2 ` +
		`AND user_id = app.ctx_user() AND revoked_at IS NULL`
	// sqlSlugTaken is the auth reader's slug availability body (00021, D-3): true when an Organization holds the slug.
	// The slug is metadata; app.hostname_state stays a pure routing reader keyed by the platform label.
	sqlSlugTaken = `SELECT app.signup_slug_taken($1)`
)

// The provisioning outcomes of app.signup_provision_organization.
const (
	OutcomeCreated        = "created"
	OutcomeAlreadyCreated = "already_created"
	outcomeIntentUsed     = "intent_used"
	outcomeSlugTaken      = "slug_taken"
	outcomeUserDisabled   = "user_disabled"
	outcomeSignupInvalid  = "signup_invalid"
	// outcomeLabelTaken (00021) is consumed only by Provision's bounded label retry; it never reaches a client.
	outcomeLabelTaken = "label_taken"
)

// provisionRow is one answer row of the provisioning body.
type provisionRow struct {
	Outcome        string
	OrganizationID string
	Slug           string
	Generation     int
	TargetOrigin   string
}

func provision(ctx context.Context, tx pgx.Tx, digest []byte, name, slug, label string, intent, ref []byte, target string, now time.Time) (provisionRow, error) {
	var r provisionRow
	var org, sl, origin *string
	var gen *int32
	err := tx.QueryRow(ctx, sqlProvision, digest, name, slug, label, intent, ref, target, now).Scan(&r.Outcome, &org, &sl, &gen, &origin)
	if err != nil {
		return provisionRow{}, storeError(err)
	}
	if org != nil {
		r.OrganizationID = *org
	}
	if sl != nil {
		r.Slug = *sl
	}
	if gen != nil {
		r.Generation = int(*gen)
	}
	if origin != nil {
		r.TargetOrigin = *origin
	}
	return r, nil
}

func challengeIssue(ctx context.Context, tx pgx.Tx, ref []byte, origin, org string, challenge, nonce []byte, now time.Time) error {
	var ok bool
	if err := tx.QueryRow(ctx, sqlChallengeIssue, ref, origin, org, challenge, nonce, now).Scan(&ok); err != nil {
		return storeError(err)
	}
	if !ok {
		return newError(CodeContinuationInvalid)
	}
	return nil
}

func prove(ctx context.Context, tx pgx.Tx, digest, nonce, proof []byte, now time.Time) error {
	tag, err := tx.Exec(ctx, sqlProve, now, proof, digest, nonce)
	if err != nil {
		return storeError(err)
	}
	if tag.RowsAffected() != 1 {
		return newError(CodeContextMismatch)
	}
	return nil
}

func abandon(ctx context.Context, tx pgx.Tx, digest []byte, now time.Time) error {
	if _, err := tx.Exec(ctx, sqlAbandon, now, digest); err != nil {
		return storeError(err)
	}
	return nil
}

func slugTaken(ctx context.Context, tx pgx.Tx, slug string) (bool, error) {
	var taken bool
	if err := tx.QueryRow(ctx, sqlSlugTaken, slug).Scan(&taken); err != nil {
		return false, storeError(err)
	}
	return taken, nil
}

// storeError maps a driver error: a CHECK violation of the provisioning INSERT (the 00011 slug grammar/reserved or the
// name rule, which Go validates first) is invalid_request naming its field; everything else is signup_unavailable.
// Driver text never leaves the package.
func storeError(err error) error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23514" {
		if strings.Contains(pg.ConstraintName, "slug") {
			return invalid(FieldSlug)
		}
		if strings.Contains(pg.ConstraintName, "name") {
			return invalid(FieldName)
		}
	}
	return newError(CodeUnavailable)
}
