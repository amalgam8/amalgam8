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
	"log"
	"time"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/serializer"
	"k8s.io/client-go/pkg/util/json"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	kuberules "github.com/amalgam8/amalgam8/pkg/adapters/rules/kubernetes"
	a8api "github.com/amalgam8/amalgam8/pkg/api"
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
	validator a8api.Validator

	// namespace from which to sync endpoints/pods
	namespace string

	// stopChan for stop signals
	stopChan chan struct{}
}

// New creates and starts the controller.
func new(ctx context.Context, namespace string) (*controller, error) {
	// If no namespace is specified, fallback to default namespace
	if namespace == "" {
		namespace = "default"
	}

	if err := initThirdPartyResource(); err != nil {
		return nil, err
	}

	// Create the kubernetes client for the third party resource
	client, err := createTPRClient()
	if err != nil {
		return nil, err
	}

	validator, err := a8api.NewValidator()
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
		cache.NewListWatchFromClient(client, "rules", namespace, nil),
		&kuberules.Rule{},
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

// initialize third party resource if it does not exist
func initThirdPartyResource() error {
	client, err := kubepkg.NewClient(kubepkg.Config{})
	if err != nil {
		return err
	}

	resName := fmt.Sprintf("%s.%s", kuberules.ResourceName, kuberules.ResourceGroupName)
	_, err = client.Extensions().ThirdPartyResources().Get(resName)
	if err != nil {
		if errors.IsNotFound(err) {
			tpr := &v1beta1.ThirdPartyResource{
				ObjectMeta: v1.ObjectMeta{
					Name: resName,
				},
				Versions: []v1beta1.APIVersion{
					{Name: kuberules.ResourceVersion},
				},
				Description: kuberules.ResourceDescription,
			}

			result, err := client.Extensions().ThirdPartyResources().Create(tpr)
			if err != nil {
				return err
			}
			log.Printf("The ThridPartYresource %#v created\n", result)
		} else {
			return err
		}
	} else {
		log.Printf("The ThirdPartyResource %s already exists\n", resName)
	}
	return nil
}

func createTPRClient() (*rest.RESTClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Failed to create the in-cluster config. %s\n", err)
		return nil, err
	}

	groupversion := unversioned.GroupVersion{
		Group:   kuberules.ResourceGroupName,
		Version: kuberules.ResourceVersion,
	}

	config.GroupVersion = &groupversion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}

	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				groupversion,
				&kuberules.Rule{},
				&kuberules.RuleList{},
				&api.ListOptions{},
				&api.DeleteOptions{},
			)
			return nil
		})
	schemeBuilder.AddToScheme(api.Scheme)

	client, err := rest.RESTClientFor(config)
	if err != nil {
		log.Printf("Failed to create the in-cluster client. %s\n", err)
		return nil, err
	}

	return client, nil
}

func (c *controller) validateRule(rule *kuberules.Rule) error {
	var updRule kuberules.Rule
	var result kuberules.Rule

	log.Printf("Rule %s is being validated\n", rule.Metadata.Name)

	if rule.Spec.ID == "" {
		updRule.Spec.ID = rule.Metadata.Name
	}

	err := c.validator.Validate(rule.Spec)
	if err != nil {
		log.Printf("Rule %#v is not valid. %s\n", rule, err)
		updRule.Status.State = kuberules.RuleStateInvalid
		updRule.Status.Message = err.Error()
	} else {
		log.Printf("Rule %s is valid\n", rule.Metadata.Name)
		updRule.Status.State = kuberules.RuleStateValid
	}

	data, err := json.Marshal(&updRule)
	if err != nil {
		log.Printf("Failed to encode status for rule %s. %s\n", rule.Metadata.Name, err)
		return err
	}

	err = c.client.Patch(api.MergePatchType).
		Resource(fmt.Sprintf("%ss", kuberules.ResourceName)).
		Namespace(rule.Metadata.Namespace).
		Name(rule.Metadata.Name).
		Body(data).
		Do().
		Into(&result)

	if err != nil {
		log.Printf("Failed to update status for rule %s. %s\n", rule.Metadata.Name, err)
	}

	return err
}

// addRule is the callback invoked by the Kubernetes cache when a rule API resource is added.
func (c *controller) addRule(obj interface{}) {
	rule, ok := obj.(*kuberules.Rule)
	if !ok {
		log.Printf("Invalid rule added: object is of type %T\n", obj)
		return
	}

	log.Printf("Rule object added: %s (state: %s, version: %s)\n",
		rule.Metadata.Name,
		rule.Status.State,
		rule.Metadata.ResourceVersion)

	if rule.Status.State == "" {
		c.validateRule(rule)
	}
}

// updateRule is the callback invoked by the Kubernetes cache when a rule API resource is updated.
func (c *controller) updateRule(oldObj, newObj interface{}) {
	oldRule, ok := oldObj.(*kuberules.Rule)
	if !ok {
		log.Printf("Invalid rule update: old object is of type %T\n", oldObj)
		return
	}
	newRule, ok := newObj.(*kuberules.Rule)
	if !ok {
		log.Printf("Invalid rule update: new object is of type %T\n", newObj)
		return
	}

	log.Printf("Rule object updated: %s (state: %s, version: %s -> %s)\n",
		newRule.Metadata.Name,
		newRule.Status.State,
		oldRule.Metadata.ResourceVersion,
		newRule.Metadata.ResourceVersion)

	if newRule.Status.State == "" {
		c.validateRule(newRule)
	}
}
