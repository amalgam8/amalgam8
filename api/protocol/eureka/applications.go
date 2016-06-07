package eureka

// Applications is an array of application objects
type Applications struct {
	Application []*Application `json:"application,omitempty"`
}
