package channels

import (
	"errors"
	"sync"
	"time"
)

// ChannelTimeout represents a go channel with timeout
type ChannelTimeout interface {
	// Receive returns an object from the channel or an error if the timeout expires.
	// if timeout is zero then the receive block until an object is available.
	Receive(timeout time.Duration) (interface{}, error)

	// Send adds the given object into the channel or returns an error if the timeout expires.
	// if timeout is zero then the send block until the object is added.
	Send(obj interface{}, timeout time.Duration) error

	// Close the channel
	Close() error

	// Channel returns the underlying channel
	Channel() chan interface{}
}

type chTimeout struct {
	ch       chan interface{}
	isClosed bool
	sync.Mutex
}

var (
	errChannelFullTimeout  = errors.New("channel full timeout")
	errChannelEmptyTimeout = errors.New("channel empty timeout")
	errChannelClosed       = errors.New("channel is closed")
)

// NewChannelTimeout creates a channel with timeout
func NewChannelTimeout(capacity int) ChannelTimeout {
	return &chTimeout{
		ch: make(chan interface{}, capacity),
	}
}

// Receive returns an object from the channel or an error if the timeout expires
func (ct *chTimeout) Receive(timeout time.Duration) (interface{}, error) {
	if timeout == 0 {
		obj := <-ct.ch
		return obj, nil
	}

	select {
	case obj := <-ct.ch:
		return obj, nil
	case <-time.After(timeout):
		return nil, errChannelEmptyTimeout
	}
}

// Send adds the given object into the channel or returns an error if the timeout expires
func (ct *chTimeout) Send(obj interface{}, timeout time.Duration) error {
	if timeout == 0 {
		ct.ch <- obj
		return nil
	}

	select {
	case ct.ch <- obj:
		return nil
	case <-time.After(timeout):
		return errChannelFullTimeout
	}
}

// Close the channel
func (ct *chTimeout) Close() error {
	ct.Lock()
	defer ct.Unlock()
	if ct.isClosed {
		return errChannelClosed
	}
	close(ct.ch)
	ct.isClosed = true
	return nil
}

// Channel returns the underlying channel
func (ct *chTimeout) Channel() chan interface{} {
	return ct.ch
}
