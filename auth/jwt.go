package auth

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
)

// JWT related constants
const (
	SigningAlgorithm = "HS256"
	NamespaceClaim   = "namespace"
)

type jwtAuthenticator struct {
	key []byte
}

// NewJWTAuthenticator creates a new Json-Web-Token authenticator based on the provided configuration options.
// Returns a valid Authenticator interface on success or an error on failure
func NewJWTAuthenticator(key []byte) (Authenticator, error) {
	if key == nil || len(key) == 0 {
		return nil, errors.New("Secret key is required")
	}
	return &jwtAuthenticator{key: key}, nil
}

func (aut *jwtAuthenticator) Authenticate(token string) (*Namespace, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	t, err := aut.parseToken(token)
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, ErrUnrecognizedToken
			}
		}
		return nil, ErrUnauthorized
	}

	claim, exists := t.Claims[NamespaceClaim]
	if !exists || claim.(string) == "" {
		return nil, ErrUnauthorized
	}

	namespace := Namespace(claim.(string))
	return &namespace, nil
}

func (aut *jwtAuthenticator) parseToken(token string) (*jwt.Token, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(SigningAlgorithm) != token.Method {
			return nil, ErrUnauthorized
		}
		return aut.key, nil
	})
}
