package amalgam8

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/amalgam8/registry/utils/reflection"
)

// InstanceRegistration encapsulates information needed for a service instance registration request
type InstanceRegistration struct {
	ServiceName string           `json:"service_name,omitempty"`
	Endpoint    *InstanceAddress `json:"endpoint,omitempty"`
	TTL         uint32           `json:"ttl,omitempty"`
	Status      string           `json:"status,omitempty"`
	Metadata    json.RawMessage  `json:"metadata,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
}

// String output the structure
func (ir *InstanceRegistration) String() string {
	mtlen := 0
	if ir.Metadata != nil {
		mtlen = len(ir.Metadata)
	}
	return fmt.Sprintf("service_name: %s, endpoint: %s, ttl: %d, status: %s, metadata: %d",
		ir.ServiceName, ir.Endpoint, ir.TTL, ir.Status, mtlen)
}

// ServiceInstance defines the response of a successful instance registration request
type ServiceInstance struct {
	ID            string           `json:"id,omitempty"`
	ServiceName   string           `json:"service_name,omitempty"`
	Endpoint      *InstanceAddress `json:"endpoint,omitempty"`
	TTL           uint32           `json:"ttl,omitempty"`
	Status        string           `json:"status,omitempty"`
	Metadata      json.RawMessage  `json:"metadata,omitempty"`
	LastHeartbeat *time.Time       `json:"last_heartbeat,omitempty"`
	Links         *InstanceLinks   `json:"links,omitempty"`
	Tags          []string         `json:"tags,omitempty"`
}

// String output the structure
func (si *ServiceInstance) String() string {
	mtlen := 0
	if si.Metadata != nil {
		mtlen = len(si.Metadata)
	}
	return fmt.Sprintf("id: %s, serviceName: %s, endpoint: %s, ttl: %d, status: %s, lastHeartbeat: %v, metadata: %d, links: %s, tags:%s",
		si.ID, si.ServiceName, si.Endpoint, si.TTL, si.Status, si.LastHeartbeat, mtlen, si.Links, si.Tags)
}

// GetJSONToFieldsMap returns a map from JSON fields to struct field names
func (si *ServiceInstance) GetJSONToFieldsMap() map[string]string {
	return reflection.GetJSONToFieldsMap(si)
}
