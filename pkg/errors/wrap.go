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

package errors

import (
	"bytes"
	"fmt"
)

// Wrap decorates the given error with the given message
func Wrap(cause error, message string) error {
	return &wrapper{
		cause:   cause,
		message: message,
	}
}

// Wrapf decorates the given error with a message built from the Sprintf-style format and arguments.
func Wrapf(cause error, format string, args ...interface{}) error {
	return Wrap(cause, fmt.Sprintf(format, args...))
}

type wrapper struct {
	cause   error
	message string
}

// Error returns a textual representation of the error
func (w *wrapper) Error() string {
	var buf bytes.Buffer
	if w.message != "" {
		buf.WriteString(w.message)
		buf.WriteString(": ")
	}
	if w.cause != nil {
		buf.WriteString(w.cause.Error())
	}
	return buf.String()
}
