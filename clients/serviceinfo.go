package clients

// Endpoint TODO
type Endpoint struct {
	Type  string `json:"type" encrypt:"true"`
	Value string `json:"value" encrypt:"true"`
}

// MetaData service instance metadata
type MetaData struct {
	Version string `json:"version"`
}

// Instance TODO
type Instance struct {
	Endpoint    Endpoint `json:"endpoint"`
	ServiceName string   `json:"service_name,omitempty"`
	MetaData    MetaData `json:"metadata"`
	// Also has TTL and last_heartbeat, but we don't use them
}

// ByInstance TODO
type ByInstance []Instance

// Len TODO
func (a ByInstance) Len() int {
	return len(a)
}

// Swap TODO
func (a ByInstance) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less TODO
func (a ByInstance) Less(i, j int) bool {
	return a[i].Endpoint.Value < a[j].Endpoint.Value
}

// ServiceInfo TODO
type ServiceInfo struct {
	ServiceName string     `json:"service_name" encrypt:"true"`
	Instances   []Instance `json:"instances"`
}

// ByService TODO
type ByService []ServiceInfo

// Len TODO
func (a ByService) Len() int {
	return len(a)
}

// Swap TODO
func (a ByService) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less TODO
func (a ByService) Less(i, j int) bool {
	return a[i].ServiceName < a[j].ServiceName
}

// Links TODO
type Links struct {
	Heartbeat string `json:"heartbeat"`
}

// RegisterResponse TODO
type RegisterResponse struct {
	Links Links `json:"links"`
}

// RegisteredService TODO
type RegisteredService struct {
	ServiceName string   `json:"service_name"`
	Endpoint    Endpoint `json:"endpoint"`
	TTL         int      `json:"ttl"`
}
