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
