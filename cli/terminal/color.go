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

package terminal

import (
	"fmt"
)

const (
	//Black .
	Black uint = 30
	// Red .
	Red uint = 31
	// Green .
	Green uint = 32
	// Yellow .
	Yellow uint = 33
	// Blue .
	Blue uint = 34
	// Magenta .
	Magenta uint = 35
	// Cyan .
	Cyan uint = 36
	// Grey .
	Grey uint = 37
	// White .
	White uint = 38
)

// Weight .
type Weight uint

const (
	// Normal .
	Normal uint = 0
	// Bold .
	Bold uint = 1
	// Lighter .
	Lighter uint = 2
)

// FontColor returns a colorized string.
func (t *term) FontColor(color uint, weight uint, message interface{}) string {
	return fmt.Sprintf("\033[%d;%dm%s\033[0m", weight, color, message)
}
