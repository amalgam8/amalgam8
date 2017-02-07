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
	"io"
	"os"
)

// UI defines the UI Interface
type UI interface {
	NewTable() Table
}

type term struct {
	Input  io.Reader
	Output io.Writer
}

// NewUI returns a reader for inputs and writer for outputs.
func NewUI(input io.Reader, output io.Writer) UI {
	if input == nil {
		input = os.Stdin
	}

	if output == nil {
		output = os.Stdout
	}

	return &term{
		Input:  input,
		Output: output,
	}
}

func (t term) NewTable() Table {
	return NewTable(t.Output)
}
