package workos

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// leeway is the clock leeway for every time claim (Plan r5 §3.3; at most 60 s per Spec R3).
const leeway = 30 * time.Second

var (
	subjectRe = regexp.MustCompile(`^user_[A-Za-z0-9]{1,250}$`)
	sidRe     = regexp.MustCompile(`^[A-Za-z0-9_-]{1,255}$`)
)

// errTokenInvalid is any failure of the signed access-token contract.
var errTokenInvalid = errors.New("workos: access token invalid")

// errImpersonation marks an impersonated session (refused, never mapped).
var errImpersonation = errors.New("workos: impersonation refused")

// accessClaims are the only access-token claims read. Vendor claims (org_id, role(s), permissions, entitlements,
// feature_flags) are not even decoded: they can never be persisted, logged or returned.
type accessClaims struct {
	jwt.RegisteredClaims
	ClientID string           `json:"client_id"`
	SID      string           `json:"sid"`
	AuthTime *jwt.NumericDate `json:"auth_time"`
	Act      json.RawMessage  `json:"act"`
}

// tokenValidator implements the access-token checks of Plan r5 §3.3. Only the access_token of an authenticate
// response obtained by core is ever validated; no browser-supplied token is accepted anywhere.
type tokenValidator struct {
	issuer     string
	clientID   string
	algorithms []string
	keys       *keySource
	now        func() time.Time
}

// validate verifies signature (allow-listed asymmetric algorithms only, via the trusted client JWKS), exact issuer,
// exact client_id, sub equal to the response user ID, required sid, exp/iat/nbf with leeway and required auth_time
// not in the future. It returns errKeysUnavailable for key outages, errImpersonation for an act claim, and
// errTokenInvalid otherwise.
func (v *tokenValidator) validate(ctx context.Context, raw, userID string) (accessClaims, error) {
	var c accessClaims
	parser := jwt.NewParser(
		jwt.WithValidMethods(v.algorithms),
		jwt.WithIssuer(v.issuer),
		jwt.WithLeeway(leeway),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(v.now),
		jwt.WithStrictDecoding(),
	)
	if _, err := parser.ParseWithClaims(raw, &c, v.keys.keyfunc(ctx)); err != nil {
		if errors.Is(err, errKeysUnavailable) {
			return accessClaims{}, errKeysUnavailable
		}
		return accessClaims{}, errTokenInvalid
	}
	now := v.now()
	switch {
	case c.Issuer != v.issuer, c.ClientID != v.clientID:
		return accessClaims{}, errTokenInvalid
	case c.Subject == "" || c.Subject != userID || !subjectRe.MatchString(c.Subject):
		return accessClaims{}, errTokenInvalid
	case !sidRe.MatchString(c.SID):
		return accessClaims{}, errTokenInvalid
	case c.IssuedAt == nil:
		return accessClaims{}, errTokenInvalid
	case c.AuthTime == nil || c.AuthTime.After(now.Add(leeway)) || c.AuthTime.Time.Unix() <= 0:
		return accessClaims{}, errTokenInvalid
	}
	if len(c.Act) > 0 && string(c.Act) != "null" {
		return accessClaims{}, errImpersonation
	}
	return c, nil
}
