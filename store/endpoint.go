package store

import (
	"bytes"
)

// Endpoint represents a network endpoint.
// Immutable by convention.
type Endpoint struct {
	Type  string
	Value string
}

func (e *Endpoint) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(e.Value)
	return buffer.String()
}

// DeepClone creates a deep copy of the receiver
func (e *Endpoint) DeepClone() *Endpoint {
	cloned := *e
	return &cloned
}
