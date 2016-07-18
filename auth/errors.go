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

import "errors"

var (
	// ErrUnrecognizedToken is returned when a token has been provided to an authenticator which does not recognize it
	ErrUnrecognizedToken = errors.New("unrecognized token")

	// ErrUnauthorized is returned when the token is not valid
	ErrUnauthorized = errors.New("unauthorized")

	// ErrEmptyToken is returned when an empty token has been provided to an authenticator which does not support it
	ErrEmptyToken = errors.New("empty token")

	// ErrCommunicationError is returned when an authenticator is unable to communicate with the token issuer,
	// and hence unavailable to authorize (or unauthorize) a token.
	ErrCommunicationError = errors.New("communication error")
)
