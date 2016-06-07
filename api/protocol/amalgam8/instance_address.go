package amalgam8

import (
	"fmt"
)

const (
	//EndpointTypeTCP Denotes a TCP endpoint type
	EndpointTypeTCP = "tcp"
	//EndpointTypeUDP Denotes a UDP endpoint type
	EndpointTypeUDP = "udp"
	//EndpointTypeHTTP Denotes an HTTP endpoint type
	EndpointTypeHTTP = "http"
	//EndpointTypeUser Denotes a user-defined endpoint type
	EndpointTypeUser = "user"
)

// InstanceAddress encapsulates a service network endpoint
type InstanceAddress struct {
	Type  string `json:"type,omitempty"` // possible values: { tcp, udp, http, user}
	Value string `json:"value"`          // can't be empty string, or consists of only spaces

}

// String output the structure
func (a *InstanceAddress) String() string {
	return fmt.Sprintf("%s:%s", a.Type, a.Value)
}
