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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAuthenticatorEmptyToken(t *testing.T) {
	auth := DefaultAuthenticator()
	namespace, err := auth.Authenticate("")
	assert.NoError(t, err)
	assert.EqualValues(t, defaultNamespace, *namespace)
}

func TestDefaultAuthenticatorDefaultToken(t *testing.T) {
	auth := DefaultAuthenticator()
	namespace, err := auth.Authenticate(defaultNamespace.String())
	assert.NoError(t, err)
	assert.EqualValues(t, defaultNamespace, *namespace)
}

func TestDefaultAuthenticatorInvalidToken(t *testing.T) {
	auth := DefaultAuthenticator()
	namespace, err := auth.Authenticate("invalid-token")
	assert.Error(t, err)
	assert.Nil(t, namespace)
}
