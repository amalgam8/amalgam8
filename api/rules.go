package api

import (
	"net/http"

	"errors"

	"github.com/amalgam8/controller/metrics"
	"github.com/amalgam8/controller/rules"
	"github.com/ant0ine/go-json-rest/rest"
)

type TenantRules struct {
	Rules []rules.Rule `json:"rules"`
}

type Rule struct {
	manager rules.Manager
}

func NewRule(m rules.Manager) *Rule {
	return &Rule{
		manager: m,
	}
}

func (r *Rule) Routes(middlewares ...rest.Middleware) []*rest.Route {
	reporter := metrics.NewReporter()

	routes := []*rest.Route{
		rest.Post("/v1/rules", reportMetric(reporter, r.add, "add_rules")),
		rest.Get("/v1/rules", reportMetric(reporter, r.list, "get_rules")),
		rest.Delete("/v1/rules", reportMetric(reporter, r.remove, "delete_rules")),
	}

	for _, route := range routes {
		route.Func = rest.WrapMiddlewares(middlewares, route.Func)
	}

	return routes
}

func (r *Rule) add(w rest.ResponseWriter, req *rest.Request) error {
	tenantID := req.PathParam("id")

	tenantRules := TenantRules{}
	if err := req.DecodeJsonPayload(&tenantRules); err != nil {
		RestError(w, req, http.StatusBadRequest, "invalid_json")
		return err
	}

	if len(tenantRules.Rules) == 0 {
		RestError(w, req, http.StatusBadRequest, "no_rules_provided")
		return errors.New("no_rules_provided")
	}

	if err := r.manager.AddRules(tenantID, tenantRules.Rules); err != nil {
		// TODO: more informative error parsing
		RestError(w, req, http.StatusInternalServerError, "request_failed")
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

func (r *Rule) list(w rest.ResponseWriter, req *rest.Request) error {
	tenantID := req.PathParam("id")
	ruleIDs := getQueries("id", req)
	tags := getQueries("tag", req)

	filter := rules.Filter{
		IDs:  ruleIDs,
		Tags: tags,
	}

	rules, err := r.manager.GetRules(tenantID, filter)
	if err != nil {
		// TODO: more informative error parsing
		RestError(w, req, http.StatusInternalServerError, "could_not_get_rules")
		return err
	}

	tenantRules := TenantRules{
		Rules: rules,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&tenantRules)
	return nil
}

func (r *Rule) remove(w rest.ResponseWriter, req *rest.Request) error {
	tenantID := req.PathParam("id")
	ruleIDs := getQueries("id", req)
	tags := getQueries("tag", req)

	filter := rules.Filter{
		IDs:  ruleIDs,
		Tags: tags,
	}

	if err := r.manager.DeleteRules(tenantID, filter); err != nil {
		// TODO: more informative error parsing
		RestError(w, req, http.StatusInternalServerError, "could_not_delete_rules")
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func getQueries(key string, req *rest.Request) []string {
	queries := req.URL.Query()
	values, ok := queries[key]
	if !ok || len(values) == 0 {
		return []string{}
	}
	return values
}
