package rules

import (
	"fmt"

	kuberules "github.com/amalgam8/amalgam8/pkg/adapters/rules/kubernetes"
	"github.com/amalgam8/amalgam8/pkg/api"
	kubepkg "github.com/amalgam8/amalgam8/pkg/kubernetes"
	"github.com/pborman/uuid"
	kubeapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/rest"
)

var (
	// BulkNotSupportedError is returned when unsupported bulk operations are attempted.
	BulkNotSupportedError = fmt.Errorf("bulk operations not supported")
)

// K8S controller. Currently no bulk insert/update operations are supported and
// the controller is bound to a single namespace.
type K8S struct {
	// kubernetes REST client
	client *rest.RESTClient

	// validator validates rules against the rule schema
	validator api.Validator

	// namespace from which to sync endpoints/pods
	namespace string
}

// NewK8S creates a kubernetes controller for the namespace.
func NewK8S(ns string) (*K8S, error) {
	// fallback to default namespace
	if ns == "" {
		ns = "default"
	}

	// init the TPR
	tprConfig := &kubepkg.TPRConfig{Name: kuberules.ResourceName,
		GroupName:   kuberules.ResourceGroupName,
		Version:     kuberules.ResourceVersion,
		Description: kuberules.ResourceDescription,
		Type:        &kuberules.RoutingRule{},
		ListType:    &kuberules.RoutingRuleList{}}

	if err := kubepkg.InitThirdPartyResource(tprConfig); err != nil {
		return nil, err
	}

	// create the TPR client
	client, err := kubepkg.NewTPRClient(kubepkg.Config{}, tprConfig)
	if err != nil {
		return nil, err
	}

	validator, err := api.NewValidator()
	if err != nil {
		return nil, err
	}

	return &K8S{
		client:    client,
		validator: validator,
		namespace: ns,
	}, nil
}

// AddRules validates the rules and adds them to the collection for the namespace.
// TODO: bulk operations
func (k8s *K8S) AddRules(_ string, rules []api.Rule) (out NewRules, err error) {
	if len(rules) != 1 {
		return out, BulkNotSupportedError
	}

	rule := rules[0]
	if err = k8s.validator.Validate(rule); err != nil {
		return
	}

	return k8s.addRule(rule)
}

// GetRules returns a collection of filtered rules from the namespace.
func (k8s *K8S) GetRules(_ string, f api.RuleFilter) (RetrievedRules, error) {
	return k8s.getRules(f)
}

// UpdateRules updates rules by ID in the namespace.
// TODO: bulk operations
func (k8s *K8S) UpdateRules(_ string, rules []api.Rule) (err error) {
	if len(rules) != 1 {
		return BulkNotSupportedError
	}

	rule := rules[0]
	if err = k8s.validator.Validate(rule); err != nil {
		return
	}

	if rule.ID == "" {
		return fmt.Errorf("rule ID is missing")
	}

	in := k8s.buildRoutingRule(rule)
	if err = k8s.client.Put().Body(&in).Namespace(k8s.namespace).Resource(kuberules.ResourceKind + "s").
		Name(rule.ID).Do().Error(); err != nil {
		return
	}

	return nil
}

// DeleteRules deletes rules that match the filter in the namespace.
// FIXME: retrieve, filter and delete are not atomic.
func (k8s *K8S) DeleteRules(_ string, f api.RuleFilter) (err error) {
	retrieved, err := k8s.getRules(f)
	if err != nil {
		return
	}

	for _, rule := range retrieved.Rules {
		if err = k8s.client.Delete().Namespace(k8s.namespace).Resource(kuberules.ResourceKind + "s").
			Name(rule.ID).Do().Error(); err != nil {
			return
		}
	}

	return nil
}

func (k8s *K8S) getRules(f api.RuleFilter) (out RetrievedRules, err error) {
	var res kuberules.RoutingRuleList
	if err = k8s.client.Get().Namespace(k8s.namespace).Resource(kuberules.ResourceKind + "s").
		Do().Into(&res); err != nil {
		return
	}

	// convert
	out.Rules = make([]api.Rule, 0, len(res.Items))
	for _, item := range res.Items {
		out.Rules = append(out.Rules, item.Spec)
	}

	// filter
	out.Rules = f.Apply(out.Rules)
	return out, nil
}

func (k8s *K8S) addRule(rule api.Rule) (out NewRules, err error) {
	// Generate an ID for each rule if none provided
	if rule.ID == "" {
		rule.ID = uuid.New()
	}

	in := k8s.buildRoutingRule(rule)
	if err = k8s.client.Post().Body(&in).Namespace(k8s.namespace).Resource(kuberules.ResourceKind + "s").
		Do().Error(); err != nil {
		return
	}

	out.IDs = []string{rule.ID}
	return out, nil
}

func (k8s *K8S) buildRoutingRule(rule api.Rule) kuberules.RoutingRule {
	return kuberules.RoutingRule{
		Spec: rule,
		TypeMeta: unversioned.TypeMeta{
			APIVersion: kuberules.ResourceName + "/" + kuberules.ResourceVersion,
			Kind:       kuberules.ResourceKind,
		},
		Metadata: kubeapi.ObjectMeta{
			Name:      rule.ID,
			Namespace: k8s.namespace,
		},
		Status: kuberules.StatusSpec{
			State: kuberules.RuleStateValid,
		},
	}
}
