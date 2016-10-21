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

import "context"

const (
	// Module name to be used in logging
	module = "AUTH"

	// ContextHeadersKey is the key in the context for the headers passed to the Authenticators from the auth middleware
	ContextHeadersKey = "headers"
)

// Authenticator is an interface for token authentication
type Authenticator interface {

	// Authenticate resolves an arbitrary string token into a namespace.
	Authenticate(ctx context.Context, token string) (*Namespace, error)
}
