// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

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
