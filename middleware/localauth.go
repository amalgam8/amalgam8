package middleware

import "errors"

// LocalAuth implements a simple passthrough
type LocalAuth struct{}

// Authenticate any token and return the token as an ID
func (l *LocalAuth) Authenticate(token string) (string, error) {
	if token != "" {
		return token, nil
	}

	return token, errors.New("Invalid token")
}
