package eureka

// Application is an array of instances
type Application struct {
	Name      string      `json:"name,omitempty"`
	Instances []*Instance `json:"instance,omitempty"`
}
