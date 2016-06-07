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
