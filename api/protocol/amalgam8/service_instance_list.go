package amalgam8

// InstanceList type is returned in response to a request to list the instances of a service name
type InstanceList struct {
	ServiceName string             `json:"service_name,omitempty"`
	Instances   []*ServiceInstance `json:"instances"`
}
