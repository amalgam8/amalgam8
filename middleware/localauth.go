package middleware

import "errors"

type LocalAuth struct {}

func (l *LocalAuth) Authenticate(token string) (string, error) {
	if token != "" {
		return token, nil
	}

	return token, errors.New("Invalid token")
}