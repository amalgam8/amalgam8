package api

import (
	"github.com/amalgam8/controller/checker"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/cactus/go-statsd-client/statsd"
	"net/http"
)

// Poll TODO
type Poll struct {
	checker checker.Checker
	statsd  statsd.Statter
}

// NewPoll TODO
func NewPoll(statsd statsd.Statter, checker checker.Checker) *Poll {
	return &Poll{
		statsd:  statsd,
		checker: checker,
	}
}

// Routes TODO
func (p *Poll) Routes() []*rest.Route {
	return []*rest.Route{
		rest.Post("/v1/poll", ReportMetric(p.statsd, p.Poll, "poll")),
	}
}

// Poll TODO
func (p *Poll) Poll(w rest.ResponseWriter, req *rest.Request) error {
	if err := p.checker.Check(nil); err != nil {
		RestError(w, req, http.StatusInternalServerError, "failed")
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
