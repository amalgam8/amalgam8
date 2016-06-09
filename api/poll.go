package api

import (
	"net/http"

	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/metrics"
	"github.com/ant0ine/go-json-rest/rest"
)

// Poll handles poll API
type Poll struct {
	checker  checker.Checker
	reporter metrics.Reporter
}

// NewPoll create struct
func NewPoll(reporter metrics.Reporter, checker checker.Checker) *Poll {
	return &Poll{
		reporter: reporter,
		checker:  checker,
	}
}

// Routes for poll API
func (p *Poll) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/poll", reportMetric(p.reporter, p.Poll, "poll")),
	}
}

// Poll Registry for latest changes
func (p *Poll) Poll(w rest.ResponseWriter, req *rest.Request) error {
	if err := p.checker.Check(nil); err != nil {
		RestError(w, req, http.StatusInternalServerError, "failed")
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
