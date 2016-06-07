package auth

import "errors"

var (
	// ErrUnrecognizedToken is returned when a token has been provided to an authenticator which does not recognize it
	ErrUnrecognizedToken = errors.New("unrecognized token")

	// ErrUnauthorized is returned when the token is not valid.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrEmptyToken is returned when an empty token has been provided to an authenticator which does not support it
	ErrEmptyToken = errors.New("empty token")
)
