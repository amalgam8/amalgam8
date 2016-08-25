// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package api

import (
	"net/http"

	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/metrics"
	"github.com/amalgam8/controller/rules"
	"github.com/amalgam8/controller/util/i18n"
	"github.com/ant0ine/go-json-rest/rest"
)

type TenantRules struct {
	Rules []rules.Rule `json:"rules"`
}

type ServiceRules struct {
	ServiceName string       `json:"service"`
	Rules       []rules.Rule `json:"rules"`
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
		rest.Put("/v1/rules", reportMetric(r.reporter, r.update, "update_rules")),
		rest.Delete("/v1/rules", reportMetric(r.reporter, r.remove, "delete_rules")),

		rest.Get("/v1/rules/routes", reportMetric(r.reporter, r.getRoutes, "get_all_routes")),
		rest.Get("/v1/rules/actions", reportMetric(r.reporter, r.getActions, "get_all_actions")),

		rest.Put("/v1/rules/routes/#destination", reportMetric(r.reporter, r.setRouteDestination, "put_rule_route_destination")),
		rest.Put("/v1/rules/actions/#destination", reportMetric(r.reporter, r.setActionDestination, "put_rule_action_destination")),
		rest.Get("/v1/rules/routes/#destination", reportMetric(r.reporter, r.getRouteDestination, "get_rule_route_destination")),
		rest.Get("/v1/rules/actions/#destination", reportMetric(r.reporter, r.getActionDestination, "get_rule_action_destination")),
		rest.Delete("/v1/rules/routes/#destination", reportMetric(r.reporter, r.deleteRouteDestination, "delete_rule_route_destination")),
		rest.Delete("/v1/rules/actions/#destination", reportMetric(r.reporter, r.deleteActionDestination, "delete_rule_action_destination")),
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
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorInvalidJSON)
		return err
	}

	if len(tenantRules.Rules) == 0 {
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorNoRulesProvided)
		return errors.New("no_rules_provided")
	}

	for i := range tenantRules.Rules {
		if tenantRules.Rules[i].Tags == nil {
			tenantRules.Rules[i].Tags = []string{}
		}
	}

	if err := r.manager.AddRules(tenantID, tenantRules.Rules); err != nil {
		// TODO: more informative error parsing
		handleManagerError(w, req, err)
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
		handleManagerError(w, req, err)
		return err
	}

	tenantRules := TenantRules{
		Rules: rules,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&tenantRules)
	return nil
}

// TODO: ensure all IDs have been set
func (r *Rule) update(w rest.ResponseWriter, req *rest.Request) error {
	tenantID := GetTenantID(req)

	tenantRules := TenantRules{}
	if err := req.DecodeJsonPayload(&tenantRules); err != nil {
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorInvalidJSON)
		return err
	}

	if len(tenantRules.Rules) == 0 {
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorNoRulesProvided)
		return errors.New("no_rules_provided")
	}

	for i := range tenantRules.Rules {
		if tenantRules.Rules[i].Tags == nil {
			tenantRules.Rules[i].Tags = []string{}
		}
	}

	if err := r.manager.UpdateRules(tenantID, tenantRules.Rules); err != nil {
		// TODO: more informative error parsing
		handleManagerError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (r *Rule) getRoutes(w rest.ResponseWriter, req *rest.Request) error {
	return r.getByRuleType(rules.RuleRoute, w, req)
}

func (r *Rule) getActions(w rest.ResponseWriter, req *rest.Request) error {
	return r.getByRuleType(rules.RuleAction, w, req)
}

func (r *Rule) getByRuleType(ruleType int, w rest.ResponseWriter, req *rest.Request) error {
	tenantID := GetTenantID(req)
	ruleIDs := getQueries("id", req)
	tags := getQueries("tag", req)
	destinations := getQueries("destination", req)

	filter := rules.Filter{
		IDs:          ruleIDs,
		Tags:         tags,
		Destinations: destinations,
		RuleType:     ruleType,
	}

	entries, err := r.manager.GetRules(tenantID, filter)
	if err != nil {
		handleManagerError(w, req, err)
		return err
	}

	respJSON := struct {
		Services map[string][]rules.Rule `json:"services"`
	}{
		Services: make(map[string][]rules.Rule),
	}

	services := make(map[string][]rules.Rule)
	for _, rule := range entries {
		if _, ok := services[rule.Destination]; ok {
			rulesByService := services[rule.Destination]
			rulesByService = append(rulesByService, rule)
			services[rule.Destination] = rulesByService
		} else {
			rulesByService := []rules.Rule{rule}
			services[rule.Destination] = rulesByService
		}
	}

	respJSON.Services = services

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&respJSON)

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
		handleManagerError(w, req, err)
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
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorInvalidJSON)
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

	if err := r.manager.SetRules(tenantID, filter, tenantRules.Rules); err != nil {
		handleManagerError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return nil
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
		handleManagerError(w, req, err)
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

	if err := r.manager.SetRules(tenantID, filter, []rules.Rule{}); err != nil {
		handleManagerError(w, req, err)
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

func handleManagerError(w rest.ResponseWriter, req *rest.Request, err error, args ...interface{}) {
	switch e := err.(type) {
	case *rules.InvalidRuleError:
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorInvalidRule, args)
	case *rules.JSONMarshallError:
		i18n.RestError(w, req, http.StatusInternalServerError, i18n.ErrorInternalServer, args)
	default:
		logrus.WithError(e).Warn("Unknown error")
		i18n.RestError(w, req, http.StatusInternalServerError, i18n.ErrorInternalServer, args)
	}

}
