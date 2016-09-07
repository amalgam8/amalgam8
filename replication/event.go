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
)

type event interface {
	ID() string
	Event() string
	Data() string
	Retry() int64
}

type sse struct {
	id, event, data string
	retry           int64
}

func (s *sse) ID() string    { return s.id }
func (s *sse) Event() string { return s.event }
func (s *sse) Data() string  { return s.data }
func (s *sse) Retry() int64  { return s.retry }
func (s *sse) String() string {
	return fmt.Sprintf("Id: %s, Event: %s, Retry: %d, Data: %s", s.id, s.event, s.retry, s.data)
}
