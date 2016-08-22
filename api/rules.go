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
	manager  rules.Manager
	reporter metrics.Reporter
}

func NewRule(m rules.Manager, r metrics.Reporter) *Rule {
	return &Rule{
		manager:  m,
		reporter: r,
	}
}

func (r *Rule) Routes(middlewares ...rest.Middleware) []*rest.Route {

	routes := []*rest.Route{
		rest.Post("/v1/rules", reportMetric(r.reporter, r.add, "add_rules")),
		rest.Get("/v1/rules", reportMetric(r.reporter, r.list, "get_rules")),
		rest.Delete("/v1/rules", reportMetric(r.reporter, r.remove, "delete_rules")),

		rest.Put("/v1/rules/#destination", reportMetric(r.reporter, r.setDestination, "put_rule_destination")),
		rest.Put("/v1/rules/#destination/routes", reportMetric(r.reporter, r.setRouteDestination, "put_rule_route_destination")),
		rest.Put("/v1/rules/#destination/actions", reportMetric(r.reporter, r.setActionDestination, "put_rule_action_destination")),
		rest.Get("/v1/rules/#destination/routes", reportMetric(r.reporter, r.getRouteDestination, "get_rule_route_destination")),
		rest.Get("/v1/rules/#destination/actions", reportMetric(r.reporter, r.getActionDestination, "get_rule_action_destination")),
		rest.Delete("/v1/rules/#destination/routes", reportMetric(r.reporter, r.deleteRouteDestination, "delete_rule_route_destination")),
		rest.Delete("/v1/rules/#destination/actions", reportMetric(r.reporter, r.deleteActionDestination, "delete_rule_action_destination")),
	}

	for _, route := range routes {
		route.Func = rest.WrapMiddlewares(middlewares, route.Func)
	}

	return routes
}

func (r *Rule) add(w rest.ResponseWriter, req *rest.Request) error {
	tenantID := GetTenantID(req)

	tenantRules := TenantRules{}
	if err := req.DecodeJsonPayload(&tenantRules); err != nil {
		RestError(w, req, http.StatusBadRequest, "invalid_json")
		return err
	}

	if len(tenantRules.Rules) == 0 {
		RestError(w, req, http.StatusBadRequest, "no_rules_provided")
		return errors.New("no_rules_provided")
	}

	for i := range tenantRules.Rules {
		if tenantRules.Rules[i].Tags == nil {
			tenantRules.Rules[i].Tags = []string{}
		}
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
	tenantID := GetTenantID(req)
	ruleIDs := getQueries("id", req)
	tags := getQueries("tag", req)
	destinations := getQueries("destination", req)

	filter := rules.Filter{
		IDs:          ruleIDs,
		Tags:         tags,
		Destinations: destinations,
		RuleType:     rules.RuleAny,
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
	tenantID := GetTenantID(req)
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

func (r *Rule) setByDestination(ruleType int, w rest.ResponseWriter, req *rest.Request) error {
	tenantID := GetTenantID(req)
	destination := req.PathParam("destination")

	tenantRules := TenantRules{}
	if err := req.DecodeJsonPayload(&tenantRules); err != nil {
		RestError(w, req, http.StatusBadRequest, "invalid_json")
		return err
	}

	for i := range tenantRules.Rules {
		if tenantRules.Rules[i].Tags == nil {
			tenantRules.Rules[i].Tags = []string{}
		}
	}

	filter := rules.Filter{
		Destinations: []string{destination},
		RuleType:     ruleType,
	}

	if err := r.manager.SetRulesByDestination(tenantID, filter, tenantRules.Rules); err != nil {
		// TODO: more informative error parsing
		RestError(w, req, http.StatusInternalServerError, "request_failed")
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
}

func (r *Rule) setDestination(w rest.ResponseWriter, req *rest.Request) error {
	return r.setByDestination(rules.RuleAny, w, req)
}

func (r *Rule) setRouteDestination(w rest.ResponseWriter, req *rest.Request) error {
	return r.setByDestination(rules.RuleRoute, w, req)
}

func (r *Rule) setActionDestination(w rest.ResponseWriter, req *rest.Request) error {
	return r.setByDestination(rules.RuleAction, w, req)
}

func (r *Rule) getByDestination(ruleType int, w rest.ResponseWriter, req *rest.Request) error {
	tenantID := GetTenantID(req)
	destination := req.PathParam("destination")

	filter := rules.Filter{
		Destinations: []string{destination},
		RuleType:     ruleType,
	}

	entries, err := r.manager.GetRules(tenantID, filter)
	if err != nil {
		RestError(w, req, http.StatusInternalServerError, "request_failed")
		return err
	}

	tenantRules := TenantRules{
		Rules: entries,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&tenantRules)
	return nil
}

func (r *Rule) getRouteDestination(w rest.ResponseWriter, req *rest.Request) error {
	return r.getByDestination(rules.RuleRoute, w, req)
}

func (r *Rule) getActionDestination(w rest.ResponseWriter, req *rest.Request) error {
	return r.getByDestination(rules.RuleAction, w, req)
}

func (r *Rule) deleteByDestination(ruleType int, w rest.ResponseWriter, req *rest.Request) error {
	tenantID := GetTenantID(req)
	destination := req.PathParam("destination")

	filter := rules.Filter{
		Destinations: []string{destination},
		RuleType:     ruleType,
	}

	if err := r.manager.SetRulesByDestination(tenantID, filter, []rules.Rule{}); err != nil {
		// TODO: more informative error parsing
		RestError(w, req, http.StatusInternalServerError, "request_failed")
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (r *Rule) deleteRouteDestination(w rest.ResponseWriter, req *rest.Request) error {
	return r.deleteByDestination(rules.RuleRoute, w, req)
}

func (r *Rule) deleteActionDestination(w rest.ResponseWriter, req *rest.Request) error {
	return r.deleteByDestination(rules.RuleAction, w, req)
}

func getQueries(key string, req *rest.Request) []string {
	queries := req.URL.Query()
	values, ok := queries[key]
	if !ok || len(values) == 0 {
		return []string{}
	}
	return values
}
