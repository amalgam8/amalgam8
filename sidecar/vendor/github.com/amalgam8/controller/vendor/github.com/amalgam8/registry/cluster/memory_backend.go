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

import (
	"fmt"
	"sync"
)

func newMemoryBackend() *memoryBackend {
	return &memoryBackend{
		members: make(map[MemberID]*member),
	}
}

type memoryBackend struct {
	members map[MemberID]*member
	mutex   sync.RWMutex
}

func (b *memoryBackend) WriteMember(m *member) error {

	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.members[m.ID()] = m
	return nil

}

func (b *memoryBackend) DeleteMember(id MemberID) error {

	b.mutex.Lock()
	defer b.mutex.Unlock()

	_, exists := b.members[id]
	if !exists {
		return fmt.Errorf("Member %v does not exist", id)
	}
	delete(b.members, id)
	return nil

}

func (b *memoryBackend) ReadMember(id MemberID) (*member, error) {

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.members[id], nil

}

func (b *memoryBackend) ReadMembers() (map[MemberID]*member, error) {

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	snapshot := make(map[MemberID]*member, len(b.members))
	for id, member := range b.members {
		snapshot[id] = member
	}
	return snapshot, nil

}

func (b *memoryBackend) ReadMemberIDs() (map[MemberID]struct{}, error) {

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	ids := make(map[MemberID]struct{})
	for id := range b.members {
		ids[id] = struct{}{}
	}
	return ids, nil

}
