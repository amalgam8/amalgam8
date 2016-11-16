package api

const (
	servicesPath = "/api/v1/services"
)

// ServiceList .
type ServiceList struct {
	Services []string `json:"services" yaml:"services"`
}

// InstanceList .
type InstanceList struct {
	Name     string `json:"service_name" yaml:"service_name"`
	Instance []struct {
		ID          string `json:"id" yaml:"id"`
		ServiceName string `json:"service_name" yaml:"service_name"`
		Endpoint    struct {
			Type  string `json:"type" yaml:"type"`
			Value string `json:"value" yaml:"value"`
		} `json:"endpoint" yaml:"endpoint"`
		TTL           int      `json:"ttl" yaml:"ttl"`
		Status        string   `json:"status" yaml:"status"`
		LastHeartbeat string   `json:"last_heartbeat" yaml:"last_heartbeat"`
		Tags          []string `json:"tags" yaml:"tags"`
	} `json:"instances" yaml:"instances"`
}

// ServiceInstancesList .
type ServiceInstancesList struct {
	Service   string   `json:"service" yaml:"service"`
	Instances []string `json:"instances" yaml:"instances"`
}
