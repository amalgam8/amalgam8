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

package datastructures

// StringSet is a set of strings, supporting basic set and inter-set operations.
// The StringSet is abstracted on top of a map[string]struct{}, so it also supports range-style iteration,
// as well as direct map operations.
// The implementation is not thread-safe.
type StringSet map[string]struct{}

// NewDefaultStringSet creates a new StringSet with a default initial capacity.
func NewDefaultStringSet() StringSet {
	return make(StringSet)
}

// NewStringSet creates a new StringSet with the specified initial capacity.
func NewStringSet(initialCapacity int) StringSet {
	return make(StringSet, initialCapacity)
}

// Add the given string to the set.
// Returns true if the set has changed, and false otherwise.
func (set StringSet) Add(s string) bool {
	_, exist := set[s]
	set[s] = struct{}{}
	return !exist
}

// Remove the given string from the set.
// Returns true if the set has changed, and false otherwise.
func (set StringSet) Remove(s string) bool {
	_, exist := set[s]
	delete(set, s)
	return exist
}

// Exists returns whether the given string is in the set.
func (set StringSet) Exists(s string) bool {
	_, exist := set[s]
	return exist
}

// Union creates a set containing all elements in either the receiver set or the given set.
func (set StringSet) Union(otherSet StringSet) StringSet {
	newSet := NewStringSet(max(len(set), len(otherSet)))

	for elem := range set {
		newSet.Add(elem)
	}

	for elem := range otherSet {
		newSet.Add(elem)
	}

	return newSet
}

// Intersection creates a set containing all elements in both the receiver set and the given set.
func (set StringSet) Intersection(otherSet StringSet) StringSet {
	newSet := NewStringSet(min(len(set), len(otherSet)))

	if set == nil || otherSet == nil {
		return newSet
	}

	for elem := range set {
		if otherSet.Exists(elem) {
			newSet.Add(elem)
		}
	}

	return newSet
}

// Difference creates a set containing all elements in the receiver set which are not in the given set.
func (set StringSet) Difference(otherSet StringSet) StringSet {
	newSet := NewStringSet(len(set))

	if set == nil {
		return newSet
	}

	for elem := range set {
		if newSet == nil || !otherSet.Exists(elem) {
			newSet.Add(elem)
		}
	}

	return newSet
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
