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

package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloneEndpoint(t *testing.T) {

	original := &Endpoint{
		Value: "127.0.0.1:9080",
		Type:  "tcp",
	}

	cloned := original.DeepClone()

	assert.Equal(t, original, cloned)
	assert.False(t, &original == &cloned)

	cloned.Value = "127.0.0.1:9443"
	assert.NotEqual(t, original, cloned)

}
