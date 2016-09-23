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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBKeyString(t *testing.T) {
	dbk := DBKey{Namespace: "namespace", InstanceID: "inst-id"}
	parts := []string{registryRecordKey, namespaceKey, "namespace", instanceKey, "inst-id"}
	expectedString := strings.Join(parts, keySeparator)

	assert.Equal(t, expectedString, dbk.String())
}

func TestDBKeyParse(t *testing.T) {
	invalidPrefix := "somestringvalue" + keySeparator + instanceKey + keySeparator + "instvalue"
	invalidInstKey := registryRecordKey + keySeparator + namespaceKey + keySeparator + "somestringvalue"
	invalidBoth := "somestringvalue"

	// Validate the DecodeString function
	actualValue, err := parseStringIntoDBKey(invalidPrefix)
	assert.Nil(t, actualValue)
	assert.Error(t, err)

	actualValue, err = parseStringIntoDBKey(invalidInstKey)
	assert.Nil(t, actualValue)
	assert.Error(t, err)

	actualValue, err = parseStringIntoDBKey(invalidBoth)
	assert.Nil(t, actualValue)
	assert.Error(t, err)

	// Test a valid key
	parts := []string{registryRecordKey, namespaceKey, "namespace", instanceKey, "instid"}
	validString := strings.Join(parts, keySeparator)
	expectedValue := DBKey{Namespace: "namespace", InstanceID: "instid"}
	actualValue, err = parseStringIntoDBKey(validString)
	assert.Equal(t, expectedValue, *actualValue)
	assert.NoError(t, err)
}
