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
	"github.com/amalgam8/amalgam8/controller/metrics"
	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/controller/util/i18n"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/ant0ine/go-json-rest/rest"
)

// Rule API.
type Rule struct {
	manager  rules.Manager
	reporter metrics.Reporter
}

// NewRule constructs a new Rule API.
func NewRule(m rules.Manager, r metrics.Reporter) *Rule {
	return &Rule{
		manager:  m,
		reporter: r,
	}
}

// Routes returns this API's routes wrapped by the middlewares.
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
	namespace := GetNamespace(req)

	rulesSet := api.RulesSet{}
	if err := req.DecodeJsonPayload(&rulesSet); err != nil {
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorInvalidJSON)
		return err
	}

	if len(rulesSet.Rules) == 0 {
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorNoRulesProvided)
		return errors.New("no_rules_provided")
	}

	for i := range rulesSet.Rules {
		if rulesSet.Rules[i].Tags == nil {
			rulesSet.Rules[i].Tags = []string{}
		}
	}

	newRules, err := r.manager.AddRules(namespace, rulesSet.Rules)
	if err != nil {
		handleManagerError(w, req, err)
		return err
	}

	resp := struct {
		IDs []string `json:"ids"`
	}{
		IDs: newRules.IDs,
	}

	w.WriteHeader(http.StatusCreated)
	w.WriteJson(&resp)
	return nil
}

func (r *Rule) list(w rest.ResponseWriter, req *rest.Request) error {
	namespace := GetNamespace(req)
	ruleIDs := getQueries("id", req)
	tags := getQueries("tag", req)
	destinations := getQueries("destination", req)

	filter := api.RuleFilter{
		IDs:          ruleIDs,
		Tags:         tags,
		Destinations: destinations,
		RuleType:     api.RuleAny,
	}

	return r.get(namespace, filter, w, req)
}

func (r *Rule) get(ns string, f api.RuleFilter, w rest.ResponseWriter, req *rest.Request) error {
	res, err := r.manager.GetRules(ns, f)
	if err != nil {
		handleManagerError(w, req, err)
		return err
	}

	resp := api.RulesSet{
		Rules:    res.Rules,
		Revision: res.Revision,
	}

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&resp)
	return nil
}

// TODO: ensure all IDs have been set
func (r *Rule) update(w rest.ResponseWriter, req *rest.Request) error {
	namespace := GetNamespace(req)

	rulesSet := api.RulesSet{}
	if err := req.DecodeJsonPayload(&rulesSet); err != nil {
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorInvalidJSON)
		return err
	}

	if len(rulesSet.Rules) == 0 {
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorNoRulesProvided)
		return errors.New("no_rules_provided")
	}

	for i := range rulesSet.Rules {
		if rulesSet.Rules[i].Tags == nil {
			rulesSet.Rules[i].Tags = []string{}
		}
	}

	if err := r.manager.UpdateRules(namespace, rulesSet.Rules); err != nil {
		handleManagerError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (r *Rule) getRoutes(w rest.ResponseWriter, req *rest.Request) error {
	return r.getByRuleType(api.RuleRoute, w, req)
}

func (r *Rule) getActions(w rest.ResponseWriter, req *rest.Request) error {
	return r.getByRuleType(api.RuleAction, w, req)
}

func (r *Rule) getByRuleType(ruleType int, w rest.ResponseWriter, req *rest.Request) error {
	namespace := GetNamespace(req)
	ruleIDs := getQueries("id", req)
	tags := getQueries("tag", req)
	destinations := getQueries("destination", req)

	filter := api.RuleFilter{
		IDs:          ruleIDs,
		Tags:         tags,
		Destinations: destinations,
		RuleType:     ruleType,
	}

	retrievedRules, err := r.manager.GetRules(namespace, filter)
	if err != nil {
		handleManagerError(w, req, err)
		return err
	}

	respJSON := struct {
		Services map[string][]api.Rule `json:"services"`
	}{
		Services: make(map[string][]api.Rule),
	}

	services := make(map[string][]api.Rule)
	for _, rule := range retrievedRules.Rules {
		if _, ok := services[rule.Destination]; ok {
			rulesByService := services[rule.Destination]
			rulesByService = append(rulesByService, rule)
			services[rule.Destination] = rulesByService
		} else {
			rulesByService := []api.Rule{rule}
			services[rule.Destination] = rulesByService
		}
	}

	respJSON.Services = services

	w.WriteHeader(http.StatusOK)
	w.WriteJson(&respJSON)

	return nil
}

func (r *Rule) delete(ns string, f api.RuleFilter, w rest.ResponseWriter, req *rest.Request) error {
	if err := r.manager.DeleteRules(ns, f); err != nil {
		handleManagerError(w, req, err)
		return err
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (r *Rule) remove(w rest.ResponseWriter, req *rest.Request) error {
	ns := GetNamespace(req)
	ruleIDs := getQueries("id", req)
	tags := getQueries("tag", req)
	dests := getQueries("destination", req)

	f := api.RuleFilter{
		IDs:          ruleIDs,
		Tags:         tags,
		Destinations: dests,
	}

	return r.delete(ns, f, w, req)
}

func (r *Rule) set(ns string, f api.RuleFilter, w rest.ResponseWriter, req *rest.Request) error {
	rulesSet := api.RulesSet{}
	if err := req.DecodeJsonPayload(&rulesSet); err != nil {
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorInvalidJSON)
		return err
	}

	for i := range rulesSet.Rules {
		if rulesSet.Rules[i].Tags == nil {
			rulesSet.Rules[i].Tags = []string{}
		}
	}

	newRules, err := r.manager.SetRules(ns, f, rulesSet.Rules)
	if err != nil {
		handleManagerError(w, req, err)
		return err
	}

	resp := struct {
		IDs []string `json:"ids"`
	}{
		IDs: newRules.IDs,
	}

	w.WriteHeader(http.StatusCreated)
	w.WriteJson(&resp)
	return nil
}

func (r *Rule) setRouteDestination(w rest.ResponseWriter, req *rest.Request) error {
	ns := GetNamespace(req)
	dest := req.PathParam("destination")

	f := api.RuleFilter{
		Destinations: []string{dest},
		RuleType:     api.RuleRoute,
	}

	return r.set(ns, f, w, req)
}

func (r *Rule) setActionDestination(w rest.ResponseWriter, req *rest.Request) error {
	ns := GetNamespace(req)
	dest := req.PathParam("destination")

	f := api.RuleFilter{
		Destinations: []string{dest},
		RuleType:     api.RuleAction,
	}

	return r.set(ns, f, w, req)
}

func (r *Rule) getRouteDestination(w rest.ResponseWriter, req *rest.Request) error {
	ns := GetNamespace(req)
	dest := req.PathParam("destination")

	f := api.RuleFilter{
		Destinations: []string{dest},
		RuleType:     api.RuleRoute,
	}

	return r.get(ns, f, w, req)
}

func (r *Rule) getActionDestination(w rest.ResponseWriter, req *rest.Request) error {
	ns := GetNamespace(req)
	dest := req.PathParam("destination")

	f := api.RuleFilter{
		Destinations: []string{dest},
		RuleType:     api.RuleAction,
	}

	return r.get(ns, f, w, req)
}

func (r *Rule) deleteRouteDestination(w rest.ResponseWriter, req *rest.Request) error {
	ns := GetNamespace(req)
	dest := req.PathParam("destination")

	f := api.RuleFilter{
		Destinations: []string{dest},
		RuleType:     api.RuleRoute,
	}

	return r.delete(ns, f, w, req)
}

func (r *Rule) deleteActionDestination(w rest.ResponseWriter, req *rest.Request) error {
	ns := GetNamespace(req)
	dest := req.PathParam("destination")

	f := api.RuleFilter{
		Destinations: []string{dest},
		RuleType:     api.RuleAction,
	}

	return r.delete(ns, f, w, req)
}

func getQueries(key string, req *rest.Request) []string {
	queries := req.URL.Query()
	values, ok := queries[key]
	if !ok || len(values) == 0 {
		return []string{}
	}
	return values
}

// handleManagerError interprets errors from the manager and outputs REST error messages.
func handleManagerError(w rest.ResponseWriter, req *rest.Request, err error, args ...interface{}) {
	switch e := err.(type) {
	case *rules.InvalidRuleError:
		i18n.RestError(w, req, http.StatusBadRequest, i18n.ErrorInvalidRule, args)
	case *rules.JSONMarshalError:
		i18n.RestError(w, req, http.StatusInternalServerError, i18n.ErrorInternalServer, args)
	default:
		logrus.WithError(e).Warn("Unknown error")
		i18n.RestError(w, req, http.StatusInternalServerError, i18n.ErrorInternalServer, args)
	}

}
