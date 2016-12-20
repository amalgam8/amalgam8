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

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/util/json"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	kuberules "github.com/amalgam8/amalgam8/pkg/adapters/rules/kubernetes"
	rulesapi "github.com/amalgam8/amalgam8/pkg/api"
	kubepkg "github.com/amalgam8/amalgam8/pkg/kubernetes"
)

const (
	// cacheResyncPeriod is the period in which we do a full resync of the rules cache.
	cacheResyncPeriod = time.Duration(60) * time.Second
)

type controller struct {
	// kubernetes REST client
	client *rest.RESTClient

	// rulesCache caches service rules resources from Kubernetes API
	rulesCache      cache.Store
	rulesController cache.ControllerInterface

	// workqueue is used to queue and process events send from the cache controllers
	workqueue *kubepkg.Workqueue

	// validator validates rules against the rule schema
	validator rulesapi.Validator

	// namespace from which to sync endpoints/pods
	namespace string
}

// New creates and starts the controller.
func new(ctx context.Context, namespace string) (*controller, error) {
	// If no namespace is specified, fallback to default namespace
	if namespace == "" {
		namespace = "default"
	}

	tprConfig := &kubepkg.TPRConfig{Name: kuberules.ResourceName,
		GroupName:   kuberules.ResourceGroupName,
		Version:     kuberules.ResourceVersion,
		Description: kuberules.ResourceDescription,
		Type:        &kuberules.RoutingRule{},
		ListType:    &kuberules.RoutingRuleList{}}

	if err := kubepkg.InitThirdPartyResource(tprConfig); err != nil {
		return nil, err
	}

	// Create the kubernetes client for the third party resource
	client, err := kubepkg.NewTPRClient(kubepkg.Config{}, tprConfig)
	if err != nil {
		return nil, err
	}

	validator, err := rulesapi.NewValidator()
	if err != nil {
		return nil, err
	}

	workqueue := kubepkg.NewWorkqueue()
	controller := &controller{
		client:    client,
		workqueue: workqueue,
		validator: validator,
		namespace: namespace,
	}

	controller.rulesCache, controller.rulesController = cache.NewInformer(
		cache.NewListWatchFromClient(client, fmt.Sprintf("%ss", strings.Replace(kuberules.ResourceName, "-", "", -1)), namespace, nil),
		&kuberules.RoutingRule{},
		cacheResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    workqueue.EnqueueingAddFunc(controller.addRule),
			UpdateFunc: workqueue.EnqueueingUpdateFunc(controller.updateRule),
		},
	)

	controller.workqueue.Start()
	go controller.rulesController.Run(ctx.Done())

	return controller, nil
}

// Stop synchronizing the rules.
func (c *controller) Stop() error {
	c.workqueue.Stop()

	return nil
}

func (c *controller) validateRule(rule *kuberules.RoutingRule) error {
	var updRule kuberules.RoutingRule
	var result kuberules.RoutingRule

	if rule.Spec.ID == "" {
		updRule.Spec.ID = rule.Metadata.Name
	}

	err := c.validator.Validate(rule.Spec)
	if err != nil {
		logger.WithError(err).Warnf("Rule %#v is not valid", rule)
		updRule.Status.State = kuberules.RuleStateInvalid
		updRule.Status.Message = err.Error()
	} else {
		logger.Infof("Rule %s is valid", rule.Metadata.Name)
		updRule.Status.State = kuberules.RuleStateValid
	}

	data, err := json.Marshal(&updRule)
	if err != nil {
		logger.WithError(err).Errorf("Failed encoding status for rule %s", rule.Metadata.Name)
		return err
	}

	err = c.client.Patch(api.MergePatchType).
		Resource(fmt.Sprintf("%ss", strings.Replace(kuberules.ResourceName, "-", "", -1))).
		Namespace(rule.Metadata.Namespace).
		Name(rule.Metadata.Name).
		Body(data).
		Do().
		Into(&result)

	if err != nil {
		logger.WithError(err).Errorf("Failed updating status for rule %s", rule.Metadata.Name)
	}

	return err
}

// addRule is the callback invoked by the Kubernetes cache when a rule API resource is added.
func (c *controller) addRule(obj interface{}) {
	rule, ok := obj.(*kuberules.RoutingRule)
	if !ok {
		logger.Errorf("Invalid rule added: object is of type %T", obj)
		return
	}

	logger.Infof("Rule object added: %s (state: %s, version: %s)",
		rule.Metadata.Name,
		rule.Status.State,
		rule.Metadata.ResourceVersion)

	if rule.Status.State == "" {
		c.validateRule(rule)
	}
}

// updateRule is the callback invoked by the Kubernetes cache when a rule API resource is updated.
func (c *controller) updateRule(oldObj, newObj interface{}) {
	oldRule, ok := oldObj.(*kuberules.RoutingRule)
	if !ok {
		logger.Errorf("Invalid rule update: old object is of type %T", oldObj)
		return
	}
	newRule, ok := newObj.(*kuberules.RoutingRule)
	if !ok {
		logger.Errorf("Invalid rule update: new object is of type %T", newObj)
		return
	}

	logger.Infof("Rule object updated: %s (state: %s, version: %s -> %s)",
		newRule.Metadata.Name,
		newRule.Status.State,
		oldRule.Metadata.ResourceVersion,
		newRule.Metadata.ResourceVersion)

	if newRule.Status.State == "" {
		c.validateRule(newRule)
	}
}
