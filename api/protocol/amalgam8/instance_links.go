package amalgam8

import "strings"

// InstanceLinks type defines the REST links relating to a service instance
type InstanceLinks struct {
	Self      string `json:"self,omitempty"`
	Heartbeat string `json:"heartbeat,omitempty"`
}

// BuildLinks composes URL values in the InstanceLinks structure for the instance identifier.
// URL's are based off of the given base URL value.
func BuildLinks(baseURL, id string) *InstanceLinks {
	return &InstanceLinks{
		Self:      strings.Join([]string{baseURL, InstanceURL(id)}, ""),
		Heartbeat: strings.Join([]string{baseURL, InstanceHeartbeatURL(id)}, ""),
	}
}
