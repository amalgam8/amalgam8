package uptime

import "github.com/ant0ine/go-json-rest/rest"

// RouteHandlers returns an array of uptime route handlers
func RouteHandlers() []*rest.Route {
	return []*rest.Route{
		rest.Get(URL(), uptimeHandler),
		rest.Get(HealthyURL(), healthyHandler),
	}
}
