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

type trustedAuthenticator struct{}

var trustedAuth = &trustedAuthenticator{}

// NewTrustedAuthenticator creates a trusted authenticator instance
func NewTrustedAuthenticator() Authenticator {
	return trustedAuth
}

func (aut *trustedAuthenticator) Authenticate(token string) (*Namespace, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	namespace := Namespace(token)
	return &namespace, nil
}
