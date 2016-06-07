package store

import (
	"time"
)

const (
	defaultDefaultTTL        = time.Duration(30) * time.Second
	defaultMinimumTTL        = time.Duration(5) * time.Second
	defaultMaximumTTL        = time.Duration(10) * time.Minute
	defaultSyncTimeout       = time.Duration(30) * time.Second
	defaultNamespaceCapacity = 50
)

// DefaultConfig is the default configuration parameters for the registry
var DefaultConfig = NewConfig(defaultDefaultTTL, defaultMinimumTTL, defaultMaximumTTL, defaultNamespaceCapacity)

// Config encapsulates the registry configuration parameters
type Config struct {
	DefaultTTL time.Duration
	MinimumTTL time.Duration
	MaximumTTL time.Duration

	NamespaceCapacity int

	SyncWaitTime time.Duration
}

// NewConfig creates a new registry configuration according to the specified TTL values
func NewConfig(defaultTTL, minimumTTL, maximumTTL time.Duration, namespaceCapacity int) *Config {
	validate(defaultTTL, minimumTTL, maximumTTL, namespaceCapacity)
	return &Config{
		DefaultTTL:        defaultTTL,
		MinimumTTL:        minimumTTL,
		MaximumTTL:        maximumTTL,
		SyncWaitTime:      defaultSyncTimeout,
		NamespaceCapacity: namespaceCapacity,
	}
}

func validate(defaultTTL, minimumTTL, maximumTTL time.Duration, namespaceCapacity int) {
	if minimumTTL > maximumTTL {
		panic("Maximum TTL must be larger or equal to minimum TTL")
	}
	if defaultTTL < minimumTTL {
		panic("Default TTL must be larger or equal to minimum TTL")
	}
	if defaultTTL > maximumTTL {
		panic("Default TTL must be smaller or equal to maximum TTL")
	}
	if namespaceCapacity < -1 {
		panic("Namespace capacity must be greater than or equal to -1")
	}
}
