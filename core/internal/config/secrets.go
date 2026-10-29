package config

import (
	"errors"
	"strings"
)

// Sentinel errors a SecretProvider wraps to classify a failed lookup (G012 Plan r1 §2.2). Load reports only the
// variable name and a fixed reason for each class; the provider's own error text and type are always discarded, so
// provider messages can never carry a value, a path or a backend detail into an error or a log.
var (
	// ErrSecretNotSet: the source has no value for the name.
	ErrSecretNotSet = errors.New("secret not set")
	// ErrSecretDenied: the source refused access to the name.
	ErrSecretDenied = errors.New("secret access denied")
	// ErrSecretUnavailable: the source could not be reached or answered.
	ErrSecretUnavailable = errors.New("secret source unavailable")
	// ErrSecretAmbiguous: the source holds more than one candidate value for the name.
	ErrSecretAmbiguous = errors.New("secret ambiguous")
)

// resolveSecret asks p for name exactly once and returns either the exact value with an empty problem, or an empty
// value with one closed problem text that names only the variable. There is no fallback to the environment, a
// default or an earlier value, and a valid value is returned byte-exact (never trimmed).
func resolveSecret(p SecretProvider, name string) (string, string) {
	if p == nil {
		return "", name + " is unavailable from the secret provider"
	}
	v, err := p.Secret(name)
	switch {
	case err == nil && v == "", errors.Is(err, ErrSecretNotSet):
		return "", name + " is not set"
	case errors.Is(err, ErrSecretDenied):
		return "", name + " is denied by the secret provider"
	case errors.Is(err, ErrSecretAmbiguous):
		return "", name + " is ambiguous"
	case err != nil:
		return "", name + " is unavailable from the secret provider"
	case strings.TrimSpace(v) == "":
		return "", name + " is empty"
	case strings.ContainsAny(v, "\x00\r\n"):
		return "", name + " is invalid"
	}
	return v, ""
}
