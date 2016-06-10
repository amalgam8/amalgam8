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

package reflection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type inner struct {
	S string
}

type entity struct {
	Str   string
	N     int
	Arr   []string
	B     bool
	Inner inner
}

func TestFilterNilReturnsSame(t *testing.T) {
	var entityToFilter = entity{"ssss", 10, []string{"a", "b", "c"}, true, inner{"innerVal"}}
	var expectedCopy = entity{}
	assert.NoError(t, FilterStructByFields(&entityToFilter, &expectedCopy, nil))
	assert.Equal(t, entityToFilter, expectedCopy, "Got different struct, but should be the same")
}

func TestNoFilterReturnsZeroStruct(t *testing.T) {
	var entityToFilter = entity{"ssss", 10, []string{"a", "b", "c"}, true, inner{"innerVal"}}
	var theSameExpectedResult = entity{}
	err := FilterStructByFields(&entityToFilter, &theSameExpectedResult, []string{})
	assert.NoError(t, err)
	assert.NotEqual(t, theSameExpectedResult, entityToFilter, "Returned entity is as the entity passed to the filter")
}

func TestFilteringBadFields(t *testing.T) {
	var entityToFilter = entity{"ssss", 10, []string{"a", "b", "c"}, true, inner{"innerVal"}}
	filteredEntity := entity{}
	assert.Error(t, FilterStructByFields(&entityToFilter, &filteredEntity, []string{"U"}), "Should have returned an error")
}

func TestFilteringMissingFields(t *testing.T) {
	var entityToFilter = entity{"ssss", 10, []string{"a", "b", "c"}, true, inner{"innerVal"}}
	filteredEntity := entity{}
	err := FilterStructByFields(&entityToFilter, &filteredEntity, []string{"Str", "Arr", "Timestamp", "Inner"})
	assert.Error(t, err, "Timestamp is not a field in entity")
}

func TestNormalFiltering(t *testing.T) {
	var entityToFilter = entity{"ssss", 10, []string{"a", "b", "c"}, true, inner{"innerVal"}}
	filteredEntity := entity{}
	err := FilterStructByFields(&entityToFilter, &filteredEntity, []string{"Str", "Arr", "Inner"})
	assert.NoError(t, err)
	assert.Equal(t, 0, filteredEntity.N, "N value wasn't filtered out")
	assert.Equal(t, false, filteredEntity.B, "B value wasn't filtered out")
	assert.Equal(t, []string{"a", "b", "c"}, filteredEntity.Arr, "Arr wasn't copied correctly")
	assert.Equal(t, inner{"innerVal"}, filteredEntity.Inner, "Inner struct wasn't deep-copied")
}

func TestMatchingPrimitives(t *testing.T) {
	var entityToFilter = entity{"ssss", 10, []string{"a", "b", "c"}, true, inner{"innerVal"}}

	matches, err := StructFieldMatchesValue(entityToFilter, "Str", "ssss")
	assert.NoError(t, err)
	assert.True(t, matches)

	matches, err = StructFieldMatchesValue(entityToFilter, "Str", "sssssssss")
	assert.NoError(t, err)
	assert.False(t, matches)

	matches, err = StructFieldMatchesValue(entityToFilter, "N", 10)
	assert.NoError(t, err)
	assert.True(t, matches)

	matches, err = StructFieldMatchesValue(entityToFilter, "N", 11)
	assert.NoError(t, err)
	assert.False(t, matches)

	matches, err = StructFieldMatchesValue(entityToFilter, "Inner", entityToFilter.Inner)
	assert.NoError(t, err)
	assert.True(t, matches)
}

func TestMatchingFailsOnErrors(t *testing.T) {
	var entityToFilter = entity{"ssss", 10, []string{"a", "b", "c"}, true, inner{"innerVal"}}

	_, err := StructFieldMatchesValue(entityToFilter, "Strrrrrrr", "ssss")
	assert.Error(t, err, "Didn't return an error, but no such field should exist in the entity")

	_, err = StructFieldMatchesValue(entityToFilter, "Str", 101)
	assert.Error(t, err, "Didn't return an error, but the fields have different field types")

	// Tests different types
	_, err = StructFieldMatchesValue(entityToFilter, "Arr", []int{5, 5})
	assert.Error(t, err, "Given array has different types from Arr's array")
}

func TestMatchingArrays(t *testing.T) {
	var entityToFilter = entity{"ssss", 10, []string{"a", "b", "c"}, true, inner{"innerVal"}}

	// Tests subset-equal case
	matches, err := StructFieldMatchesValue(entityToFilter, "Arr", []string{"a", "b", "c"})
	assert.NoError(t, err)
	assert.True(t, matches)

	// Tests subset equal case
	matches, err = StructFieldMatchesValue(entityToFilter, "Arr", []string{"a", "c"})
	assert.NoError(t, err)
	assert.True(t, matches)

	// Tests not-subset case
	matches, err = StructFieldMatchesValue(entityToFilter, "Arr", []string{"a", "d"})
	assert.NoError(t, err)
	assert.False(t, matches)
}

type dummy struct {
	NoJSON        string `json:"-"`
	SameName      string `json:"SameName"`
	DifferentName string `json:"different_name,omitempty"`
	NoJSONTag     string `yml:"no_json_tag"`
	NoTag         string
}

func TestGetJSONToFieldsMap(t *testing.T) {
	expected := map[string]string{
		"SameName":       "SameName",
		"different_name": "DifferentName",
		"NoJSONTag":      "NoJSONTag",
		"NoTag":          "NoTag"}

	var d dummy
	m := GetJSONToFieldsMap(d)
	assert.Equal(t, 4, len(m))
	assert.Equal(t, expected, m)

	m = GetJSONToFieldsMap(&d)
	assert.Equal(t, 4, len(m))
	assert.Equal(t, expected, m)
}
