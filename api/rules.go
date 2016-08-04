package api

import (
	"net/http"

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
	routes := []*rest.Route{
		rest.Post("/v1/rules", r.add),
		rest.Get("/v1/rules", r.list),
		rest.Delete("/v1/rules", r.remove),
	}

	for _, route := range routes {
		route.Func = rest.WrapMiddlewares(middlewares, route.Func)
	}

	return routes
}

func (r *Rule) add(w rest.ResponseWriter, req *rest.Request) {
	tenantID := req.PathParam("id")

	tenantRules := TenantRules{}
	if err := req.DecodeJsonPayload(&tenantRules); err != nil {
		RestError(w, req, http.StatusBadRequest, "invalid_json")
		return
	}

	if len(tenantRules.Rules) == 0 {
		RestError(w, req, http.StatusBadRequest, "no_rules_provided")
		return
	}

	if err := r.manager.AddRules(tenantID, tenantRules.Rules); err != nil {
		// TODO: more informative error parsing
		RestError(w, req, http.StatusInternalServerError, "request_failed")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (r *Rule) list(w rest.ResponseWriter, req *rest.Request) {
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
		return
	}

	tenantRules := TenantRules{
		Rules: rules,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&tenantRules)
}

func (r *Rule) remove(w rest.ResponseWriter, req *rest.Request) {
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
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getQueries(key string, req *rest.Request) []string {
	queries := req.URL.Query()
	values, ok := queries[key]
	if !ok || len(values) == 0 {
		return []string{}
	}
	return values
}
