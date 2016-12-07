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

package math

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	cases := []struct {
		x, y, min int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{1, 1, 1},
		{0, 0, 0},
		{-2, 1, -2},
		{1, -2, -2},
	}

	for _, c := range cases {
		assert.Equal(t, c.min, Min(c.x, c.y), "Failed to compute Min(%d, %d)", c.x, c.y)
	}
}

func TestMax(t *testing.T) {
	cases := []struct {
		x, y, max int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{1, 1, 1},
		{0, 0, 0},
		{-2, 1, 1},
		{1, -2, 1},
	}

	for _, c := range cases {
		assert.Equal(t, c.max, Max(c.x, c.y), "Failed to compute Min(%d, %d)", c.x, c.y)
	}
}
