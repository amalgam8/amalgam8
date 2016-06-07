package amalgam8

// InstancesList type is returned in response to a request to list the instances
type InstancesList struct {
	Instances []*ServiceInstance `json:"instances"`
}
