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

package client

import (
	"bytes"
	"fmt"
)

// clientError .
type clientError struct {
	StatusCode int
	Body       string
}

// Error constructs an error based on the status code and body
func (e clientError) Error() string {
	var buffer bytes.Buffer
	buffer.WriteString("Request response:\n")

	if e.StatusCode != 0 {
		buffer.WriteString(fmt.Sprintf("status_code=%v\n", e.StatusCode))
	}

	if len(e.Body) > 0 {
		buffer.WriteString(fmt.Sprintf("body=%v\n", e.Body))
	}

	return buffer.String()
}
