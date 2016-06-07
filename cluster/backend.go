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
