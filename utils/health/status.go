package health

// Status is the result of a health check run.
type Status struct {
	Healthy    bool                   `json:"healthy"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

var (
	// Healthy is a healthy status with no additional properties.
	Healthy = Status{Healthy: true}
)

// StatusHealthy creates a new healthy status with given message property.
// To return a default healthy status, with no properties, just use health.Healthy
func StatusHealthy(message string) Status {
	s := Status{
		Healthy: true,
	}
	if len(message) > 0 {
		s.Properties = map[string]interface{}{"message": message}
	}
	return s
}

// StatusHealthyWithProperties creates a new healthy status with given properties.
func StatusHealthyWithProperties(properties map[string]interface{}) Status {
	return Status{
		Healthy:    true,
		Properties: properties,
	}
}

// StatusUnhealthy creates a new unhealthy status with the given message and error properties.
func StatusUnhealthy(message string, cause error) Status {
	s := Status{
		Healthy: false,
	}
	if len(message) > 0 {
		s.Properties = map[string]interface{}{"message": message}
	}
	if cause != nil && len(cause.Error()) > 0 {
		if s.Properties == nil {
			s.Properties = make(map[string]interface{})
		}
		s.Properties["cause"] = cause.Error()
	}
	return s
}

// StatusUnhealthyWithProperties creates a new unhealthy status with given properties.
func StatusUnhealthyWithProperties(properties map[string]interface{}) Status {
	return Status{
		Healthy:    false,
		Properties: properties,
	}
}
