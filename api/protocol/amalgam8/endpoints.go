package amalgam8

import (
	"strings"
)

// InstanceCreateURL returns URL path used for creating new instance registrations
func InstanceCreateURL() string {
	return instancesPath
}

// InstancesURL returns URL path used for querying instances
func InstancesURL() string {
	return instancesPath
}

// InstanceURL returns (client side) URL path used for interacting with the specified instance
func InstanceURL(id string) string {
	return strings.Join([]string{instancesPath, "/", id}, "")
}

// instanceTemplateURL returns the router's (server side) URL path used for interacting with an instance
func instanceTemplateURL() string {
	return instanceTemplate
}

// InstanceHeartbeatURL returns (client side) URL path used for renewing registration of the identified instance
func InstanceHeartbeatURL(id string) string {
	return strings.Join([]string{instancesPath, "/", id, heartbeat}, "")
}

// instanceHeartbeatTemplateURL returns router (server side) URL template for instance heartbeat
func instanceHeartbeatTemplateURL() string {
	return instanceHeartbeatTemplate
}

// ServiceNamesURL returns (client side) URL path used for querying service names
func ServiceNamesURL() string {
	return servicesPath
}

// ServiceInstancesURL returns the (client side) URL path corresponding to the instance list for the named service
func ServiceInstancesURL(name string) string {
	return strings.Join([]string{servicesPath, "/", name}, "")
}

// serviceInstancesTemplateURL returns the router (server side) URL template for querying for service instances
func serviceInstancesTemplateURL() string {
	return serviceInstanceTemplate
}

// API parameter names
const (
	RouteParamServiceName = "sname"
	RouteParamInstanceID  = "iid"
)

const ( // API related constants
	apiPath                   = "/api"
	apiVer                    = "/v1"
	heartbeat                 = "/heartbeat"
	instancesPath             = apiPath + apiVer + "/instances"
	servicesPath              = apiPath + apiVer + "/services"
	instanceTemplate          = instancesPath + "/#" + RouteParamInstanceID
	instanceHeartbeatTemplate = instanceTemplate + heartbeat
	serviceInstanceTemplate   = servicesPath + "/#" + RouteParamServiceName
)
