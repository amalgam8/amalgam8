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

package cluster

import "errors"

// BackendType defines the type of backend used for the cluster
type BackendType int

// Available backend types
const (
	UnspecifiedBackend BackendType = iota // Zero-value
	FilesystemBackend
	MemoryBackend
)

type backend interface {
	WriteMember(m *member) error
	DeleteMember(id MemberID) error

	ReadMember(id MemberID) (*member, error)
	ReadMembers() (map[MemberID]*member, error)
	ReadMemberIDs() (map[MemberID]struct{}, error)
}

func newBackend(conf *Config) (backend, error) {
	switch conf.BackendType {
	case FilesystemBackend:
		return newFilesystemBackend(conf.Directory, conf.TTL*3)
	case MemoryBackend:
		return newMemoryBackend(), nil
	default:
		return nil, errors.New("No backend type configured")
	}
}
