package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/cactus/go-statsd-client/statsd"
	"net/http"
)

// Health handles health API calls
type Health struct {
	statsdClient statsd.Statter
}

// NewHealth creates struct
func NewHealth(statter statsd.Statter) *Health {
	return &Health{
		statsdClient: statter,
	}
}

// Routes for health check API
func (h *Health) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/health", ReportMetric(h.statsdClient, h.GetHealth, "controller_health")),
	}
}

// GetHealth performs health check on controller and dependencies
func (h *Health) GetHealth(w rest.ResponseWriter, req *rest.Request) error {
	// TODO: perform checks on cloudant, optionally SD and MH
	w.WriteHeader(http.StatusOK)
	return nil
}
