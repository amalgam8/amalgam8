package amalgam8

// ServicesList is the response type returned from a request to query service names
type ServicesList struct {
	Services []string `json:"services"`
}
