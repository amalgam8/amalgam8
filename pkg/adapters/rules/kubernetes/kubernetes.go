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

package kubernetes

import (
	"fmt"
	"strings"
	"sync"
	"time"

	a8api "github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/pkg/auth"
	kubepkg "github.com/amalgam8/amalgam8/pkg/kubernetes"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	// RulesCacheResyncPeriod is the period in which we do a full resync of the Rules cache.
	RulesCacheResyncPeriod = time.Duration(60) * time.Second
)

// Make sure we implement the RulesStore interface
var _ a8api.RulesService = (*Adapter)(nil)

// Package global logger
var logger = logging.GetLogger("KUBERNETES_RULES")

// Config stores configurable attributes of the Kubernetes adapter.
type Config struct {
	URL       string
	Token     string
	Namespace auth.Namespace
}

// Adapter for Kubernetes Service Discovery.
type Adapter struct {

	// kubernetes REST client
	client *rest.RESTClient

	// workqueue is used to queue and process events send from the cache controllers
	workqueue *kubepkg.Workqueue

	// rulesCache caches rules from Kubernetes API
	rulesCache      cache.Store
	rulesController cache.ControllerInterface

	// rules maps a rule unique name to a api rule object.
	rules map[string]*a8api.Rule

	// namespace from which to sync rules
	namespace string

	// stopChan for stop signals
	stopChan chan struct{}

	// revision . We assume that sidecar and rules fetcher runs always together
	revision int

	// mutex is used to synchronize access to the 'rules' map,
	// which is read by the ListRules() method (externally),
	// and is written by the cache event handlers (internally).
	// Given that we expect a single reader only, we use a regular sync.Mutex rather than a sync.RWMutex.
	mutex sync.Mutex
}

// New creates and starts a new Kubernetes Rules adapter.
func New(config Config) (*Adapter, error) {
	namespace := config.Namespace.String()
	// If no namespace is specified, fallback to default namespace
	if namespace == "" {
		namespace = "default"
	}

	tprConfig := &kubepkg.TPRConfig{Name: ResourceName,
		GroupName:   ResourceGroupName,
		Version:     ResourceVersion,
		Description: ResourceDescription,
		Type:        &RoutingRule{},
		ListType:    &RoutingRuleList{}}

	kubePkgConfig := kubepkg.Config{URL: config.URL,
		Token: config.Token}

	// Create the kubernetes client for the rules resource
	client, err := kubepkg.NewTPRClient(kubePkgConfig, tprConfig)

	if err != nil {
		return nil, err
	}
	workqueue := kubepkg.NewWorkqueue()

	adapter := &Adapter{
		client:    client,
		workqueue: workqueue,
		rules:     make(map[string]*a8api.Rule),
		namespace: namespace,
		revision:  0,
	}
	adapter.rulesCache, adapter.rulesController = cache.NewInformer(
		// In Kubernetes the kind of the ThirdPartyResource takes the form <kind name>.<domain>.
		// Kind names will be converted to CamelCase when creating instances of the ThirdPartyResource.
		// Hyphens in the kind are assumed to be word breaks and are converted by kubernetes to CamelCase.
		cache.NewListWatchFromClient(client, fmt.Sprintf("%ss", strings.Replace(ResourceName, "-", "", -1)), namespace, nil),
		&RoutingRule{},
		RulesCacheResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    workqueue.EnqueueingAddFunc(adapter.addRule),
			UpdateFunc: workqueue.EnqueueingUpdateFunc(adapter.updateRule),
			DeleteFunc: workqueue.EnqueueingDeleteFunc(adapter.deleteRule),
		},
	)
	return adapter, adapter.Start()
}

// Start synchronizing the Kubernetes adapter.
func (a *Adapter) Start() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.stopChan != nil {
		err := fmt.Errorf("kubernetes adapter already started")

		logger.WithError(err).Errorf("Failed starting Kubernetes adapter")
		return err
	}
	a.stopChan = make(chan struct{})
	a.workqueue.Start()
	go a.rulesController.Run(a.stopChan)
	return nil
}

// Stop synchronizing the Kubernetes adapter.
func (a *Adapter) Stop() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.stopChan == nil {
		err := fmt.Errorf("kubernetes adapter not started")
		logger.WithError(err).Errorf("Failed stopping Kubernetes adapter")
		return err
	}
	close(a.stopChan)
	a.stopChan = nil
	a.workqueue.Stop()
	return nil
}

// ListRules queries for the list of rules currently exist.
func (a *Adapter) ListRules(f *a8api.RuleFilter) (*a8api.RulesSet, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	rules := make([]a8api.Rule, 0, len(a.rules))
	for _, rule := range a.rules {
		rules = append(rules, *rule)
	}
	rules = f.Apply(rules)
	rulesSet := &a8api.RulesSet{
		Rules:    rules,
		Revision: int64(a.revision),
	}

	return rulesSet, nil
}

// addRule is the callback invoked by the Kubernetes cache when a rule API resource is added.
func (a *Adapter) addRule(obj interface{}) {
	a.storeRule(obj)
}

// updateRule is the callback invoked by the Kubernetes cache when a rule API resource is updated.
func (a *Adapter) updateRule(oldObj, newObj interface{}) {
	a.storeRule(newObj)

}

// storeRules is a helper function called by updateRules() and addRule() callback functions. is stores or updates rules
// int the rules map of the adapter.
func (a *Adapter) storeRule(obj interface{}) {
	newRule, ok := (obj).(*RoutingRule)
	if !ok {
		logger.Warnf("Invalid rule : object is of type %T", obj)
		return
	}
	if newRule.Status.State == RuleStateInvalid {
		a.revision++
		delete(a.rules, newRule.Metadata.Name)
		return
	} else if newRule.Status.State == RuleStateValid {
		a.revision++
		a8Rule := &newRule.Spec
		ruleName := newRule.Metadata.Name
		a.rules[ruleName] = a8Rule
		return
	} else {
		logger.Warnf("Rule state %s is undefined . Rule %s will not be updated or created",
			newRule.Status.State, newRule.Metadata.Name)
		return
	}
}

// deleteRule is the callback invoked by the Kubernetes cache when a rule API resource is deleted.
func (a *Adapter) deleteRule(obj interface{}) {
	rule, ok := extractDeletedObject(obj).(*RoutingRule)
	if !ok {
		logger.Warnf("Trying to delete Invalid rule : object is of type %T", obj)
		return
	}
	ruleName := rule.Metadata.Name
	delete(a.rules, ruleName)
	a.revision++

}

// extractDeletedObject is used within "deleteXXX" cache callbacks, where the provided
// object may be a wrapper (DeletedFinalStateUnknown) around the actual deleted object.
func extractDeletedObject(obj interface{}) interface{} {
	deleted, ok := obj.(cache.DeletedFinalStateUnknown)
	if ok {
		return deleted.Obj
	}
	return obj
}
