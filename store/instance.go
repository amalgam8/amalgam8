package store

import (
	"fmt"
	"time"

	"github.com/amalgam8/registry/api/protocol"
)

// Registered instance status related constants
const (
	Starting     = "STARTING"
	Up           = "UP"
	OutOfService = "OUT_OF_SERVICE"
	All          = "ALL" // ALL is only a valid status for the query string param and not for the register
)

// ServiceInstance represents a runtime instance of a service.
type ServiceInstance struct {
	ID               string
	Protocol         protocol.Type
	ServiceName      string
	Endpoint         *Endpoint
	Status           string
	Metadata         []byte
	RegistrationTime time.Time
	LastRenewal      time.Time
	TTL              time.Duration
	Tags             []string
	Extension        map[string]interface{}
}

// String output the structure
func (si *ServiceInstance) String() string {
	return fmt.Sprintf("id: %s, protocol: %d, service_name: %s, endpoint: %s, status: %s, registrationTime: %v, lastRenewal: %v, ttl: %d, tags: %v",
		si.ID, si.Protocol, si.ServiceName, si.Endpoint, si.Status, si.RegistrationTime, si.LastRenewal, si.TTL, si.Tags)
}

// DeepClone creates a deep copy of the receiver
func (si *ServiceInstance) DeepClone() *ServiceInstance {
	cloned := *si
	cloned.Endpoint = si.Endpoint.DeepClone()
	if si.Metadata == nil || len(si.Metadata) == 0 {
		cloned.Metadata = nil
	} else {
		cloned.Metadata = make([]byte, len(si.Metadata))
		copy(cloned.Metadata, si.Metadata)
	}
	if len(si.Extension) == 0 {
		cloned.Extension = nil
	} else {
		cloned.Extension = make(map[string]interface{}, len(si.Extension))
		for k, v := range si.Extension {
			cloned.Extension[k] = v
		}
	}
	return &cloned
}
