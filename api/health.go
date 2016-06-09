package api

import (
	"net/http"

	"github.com/amalgam8/controller/metrics"
	"github.com/ant0ine/go-json-rest/rest"
)

// Health handles health API calls
type Health struct {
	reporter metrics.Reporter
}

// NewHealth creates struct
func NewHealth(reporter metrics.Reporter) *Health {
	return &Health{
		reporter: reporter,
	}
}

// Routes for health check API
func (h *Health) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Get("/health", reportMetric(h.reporter, h.GetHealth, "controller_health")),
	}
}

// GetHealth performs health check on controller and dependencies
func (h *Health) GetHealth(w rest.ResponseWriter, req *rest.Request) error {
	// TODO: perform checks on cloudant, optionally SD and MH
	w.WriteHeader(http.StatusOK)
	return nil
}
