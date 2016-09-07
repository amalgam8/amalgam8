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

package replication

import (
	"fmt"
	"io"
	"strings"
)

var (
	encFields = []struct {
		prefix string
		value  func(event) string
	}{
		{"id: ", event.ID},
		{"event: ", event.Event},
		{"data: ", event.Data},
	}
)

type encoder struct {
	w io.Writer
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{w: w}
}

func (enc *encoder) Encode(ev event) (err error) {
	for _, field := range encFields {
		prefix, value := field.prefix, field.value(ev)
		if len(value) == 0 {
			continue
		}
		value = strings.Replace(value, "\n", "\n"+prefix, -1)
		if _, err = io.WriteString(enc.w, prefix+value+"\n"); err != nil {
			err = fmt.Errorf("Encode: %s", err)
			return
		}
	}
	if _, err = io.WriteString(enc.w, "\n"); err != nil {
		err = fmt.Errorf("Fail Encode: %s", err)
	}
	return
}
