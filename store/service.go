package store

import "fmt"

// Service represents a runtime service group.
type Service struct {
	ServiceName string
}

func (s *Service) String() string {
	return fmt.Sprintf("%s", s.ServiceName)
}
