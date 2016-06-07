package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/cactus/go-statsd-client/statsd"
	"net/http"
)

// Health TODO
type Health struct {
	statsdClient statsd.Statter
}

// NewHealth TODO
func NewHealth(statter statsd.Statter) *Health {
	return &Health{
		statsdClient: statter,
	}
}

// Routes TODO
func (h *Health) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/health", ReportMetric(h.statsdClient, h.GetHealth, "controller_health")),
	}
}

// GetHealth TODO
func (h *Health) GetHealth(w rest.ResponseWriter, req *rest.Request) error {
	// TODO: perform checks on cloudant, optionally SD and MH
	w.WriteHeader(http.StatusOK)
	return nil
}
