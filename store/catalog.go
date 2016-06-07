package store

// Predicate for filtering returned instances
type Predicate func(si *ServiceInstance) bool

// Catalog for managing instances within a registry namespace
type Catalog interface {
	Register(si *ServiceInstance) (*ServiceInstance, error)
	Deregister(instanceID string) error
	Renew(instanceID string) error
	SetStatus(instanceID, status string) error

	Instance(instanceID string) (*ServiceInstance, error)
	List(serviceName string, predicate Predicate) ([]*ServiceInstance, error)
	ListServices(predicate Predicate) []*Service
}
