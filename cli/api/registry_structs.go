package api

const (
	servicesPath = "/api/v1/services"
)

// ServiceList .
type ServiceList struct {
	Services []string `json:"services"`
}

// InstanceList .
type InstanceList struct {
	Name     string `json:"service_name"`
	Instance []struct {
		ID          string `json:"id"`
		ServiceName string `json:"service_name"`
		Endpoint    struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"endpoint"`
		TTL           int      `json:"ttl"`
		Status        string   `json:"status"`
		LastHeartbeat string   `json:"last_heartbeat"`
		Tags          []string `json:"tags"`
	} `json:"instances"`
}

// ServiceInstancesList .
type ServiceInstancesList struct {
	Service   string   `json:"service"`
	Instances []string `json:"instances"`
}
