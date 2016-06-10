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
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCloneInstance(t *testing.T) {

	original := &ServiceInstance{
		ID:          "1",
		ServiceName: "Calc",
		Endpoint:    &Endpoint{Value: "localhost" + ":9080", Type: "tcp"},
		Status:      "UP",
		Metadata:    []byte("Metadata"),
		LastRenewal: time.Now(),
		TTL:         time.Duration(30) * time.Second,
	}

	cloned := original.DeepClone()

	assert.Equal(t, original, cloned)
	assert.False(t, &original == &cloned)

	cloned.ServiceName = "Adder"
	assert.NotEqual(t, original, cloned)

	cloned = original.DeepClone()
	cloned.Metadata = []byte("Cloned metadata")
	assert.NotEqual(t, original, cloned)

	cloned = original.DeepClone()
	cloned.Status = "DOWN"
	assert.NotEqual(t, original, cloned)

	cloned = original.DeepClone()
	cloned.LastRenewal = time.Now().Add(time.Hour)
	assert.NotEqual(t, original, cloned)

	cloned = original.DeepClone()
	cloned.Endpoint.Type = "udp"
	assert.NotEqual(t, original, cloned)
	assert.False(t, reflect.DeepEqual(original, cloned))

}

/*
func TestHasTag(t *testing.T) {

	cases := []struct {
		tags 		[]string 	// Actual instance tags
		input	 	string		// Tag to check for existence
		expected	bool		// Expected check result
	}{
		{ tags: []string{}, input: "", expected: true },
		{ tags: []string{}, input: "tag", expected: false },
		{ tags: []string{ "tag" }, input: "", expected: true },
		{ tags: []string{ "tag" }, input: "tag", expected: true },
		{ tags: []string{ "tag1" }, input: "tag2", expected: false },
		{ tags: []string{ "tag1", "tag2" }, input: "tag1", expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: "tag2", expected: true },
	}

	for i, c := range cases {
		si := ServiceInstance{ Tags: c.tags }
		actual := si.HasTag(c.input)
		assert.Equal(t, c.expected, actual, "Failure in case %v", i)
	}

}

func TestHasAllTags(t *testing.T) {

	cases := []struct {
		tags 		[]string 	// Actual instance tags
		input	 	[]string	// Tags to check for existence
		expected	bool		// Expected check result
	}{
		{ tags: []string{}, input: []string{}, expected: true },
		{ tags: []string{}, input: []string{ "" }, expected: true },
		{ tags: []string{}, input: []string{ "tag" }, expected: false },

		{ tags: []string{ "tag1" }, input: []string{}, expected: true },
		{ tags: []string{ "tag1" }, input: []string{ "" }, expected: true },
		{ tags: []string{ "tag1" }, input: []string{ "tag1" }, expected: true },
		{ tags: []string{ "tag1" }, input: []string{ "tag2" }, expected: false },
		{ tags: []string{ "tag1" }, input: []string{ "tag1", "tag2" }, expected: false },
		{ tags: []string{ "tag1" }, input: []string{ "tag2", "tag1" }, expected: false },

		{ tags: []string{ "tag1", "tag2" }, input: []string{}, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag1" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag2" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag1", "tag2" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag2", "tag1" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag1", "tag2", "tag3" }, expected: false },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag1", "tag2", "tag1" }, expected: true },

	}

	for i, c := range cases {
		si := ServiceInstance{ Tags: c.tags }
		actual := si.HasAllTags(c.input)
		assert.Equal(t, c.expected, actual, "Failure in case %v", i)
	}

}

func TestHasAnyTag(t *testing.T) {

	cases := []struct {
		tags 		[]string 	// Actual instance tags
		input	 	[]string	// Tags to check for existence
		expected	bool		// Expected check result
	}{
		{ tags: []string{}, input: []string{}, expected: true },
		{ tags: []string{}, input: []string{ "" }, expected: true },
		{ tags: []string{}, input: []string{ "tag" }, expected: false },

		{ tags: []string{ "tag1" }, input: []string{}, expected: true },
		{ tags: []string{ "tag1" }, input: []string{ "" }, expected: true },
		{ tags: []string{ "tag1" }, input: []string{ "tag1" }, expected: true },
		{ tags: []string{ "tag1" }, input: []string{ "tag2" }, expected: false },
		{ tags: []string{ "tag1" }, input: []string{ "tag1", "tag2" }, expected: true },
		{ tags: []string{ "tag1" }, input: []string{ "tag2", "tag1" }, expected: true },

		{ tags: []string{ "tag1", "tag2" }, input: []string{}, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag1" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag2" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag3" }, expected: false },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag3", "tag4" }, expected: false },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag1", "tag2" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag2", "tag1" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag1", "tag2", "tag3" }, expected: true },
		{ tags: []string{ "tag1", "tag2" }, input: []string{ "tag1", "tag2", "tag1" }, expected: true },

	}

	for i, c := range cases {
		si := ServiceInstance{ Tags: c.tags }
		actual := si.HasAnyTag(c.input)
		assert.Equal(t, c.expected, actual,
					"Failure in case %v: Instances tags are %v, and checked tags are %v",
					i, si.Tags, c.input)
	}

}

*/
