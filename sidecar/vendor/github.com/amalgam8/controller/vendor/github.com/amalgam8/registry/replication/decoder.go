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
	"bufio"
	"io"
	"strconv"
	"strings"
)

type decoder struct {
	*bufio.Reader
}

func newDecoder(r io.Reader) *decoder {
	dec := &decoder{bufio.NewReader(r)}
	return dec
}

func (dec *decoder) Decode() (event, error) {

	// peek ahead before we start a new event so we can return EOFs
	_, err := dec.Peek(1)
	if err == io.ErrUnexpectedEOF {
		err = io.EOF
	}
	if err != nil {
		return nil, err
	}
	ev := new(sse)
	for {
		line, err := dec.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if line == "\n" {
			break
		}
		line = strings.TrimSuffix(line, "\n")
		if strings.HasPrefix(line, ":") {
			continue
		}
		sections := strings.SplitN(line, ":", 2)
		field, value := sections[0], ""
		if len(sections) == 2 {
			value = strings.TrimPrefix(sections[1], " ")
		}
		switch field {
		case "event":
			ev.event = value
		case "data":
			ev.data += value + "\n"
		case "id":
			ev.id = value
		case "retry":
			ev.retry, _ = strconv.ParseInt(value, 10, 64)
		}
	}
	ev.data = strings.TrimSuffix(ev.data, "\n")
	return ev, nil
}
