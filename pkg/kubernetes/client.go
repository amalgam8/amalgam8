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

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	k8serrors "k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/serializer"
	"k8s.io/client-go/rest"

	"github.com/amalgam8/amalgam8/pkg/errors"
)

// Config stores configurable attributes of a Kubernetes client.
type Config struct {
	URL   string
	Token string
}

// TPRConfig stores configurable attributes of a Kubernetes ThirdPartyResource object.
type TPRConfig struct {
	Name        string
	GroupName   string
	Version     string
	Description string
	Type        runtime.Object
	ListType    runtime.Object
}

// NewClient creates a new Kubernetes client using the specified configuration.
// If no URL and Token are specified, then these values are attempted to be read from the
// service account (if running within a Kubernetes pod).
func NewClient(config Config) (kubernetes.Interface, error) {
	kubeConfig, err := newKubeConfig(&config)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating Kubernetes REST client")
	}
	return client, nil
}

// NewTPRClient creates a new Kubernetes client using the specified configuration for
// working with the specified ThirdPartyResource.
// If no URL and Token are specified, then these values are attempted to be read from the
// service account (if running within a Kubernetes pod).
func NewTPRClient(config Config, tprConfig *TPRConfig) (*rest.RESTClient, error) {
	kubeConfig, err := newKubeConfig(&config)
	if err != nil {
		return nil, err
	}

	groupversion := unversioned.GroupVersion{
		Group:   tprConfig.GroupName,
		Version: tprConfig.Version,
	}

	kubeConfig.GroupVersion = &groupversion
	kubeConfig.APIPath = "/apis"
	kubeConfig.ContentType = runtime.ContentTypeJSON
	kubeConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}

	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				groupversion,
				tprConfig.Type,
				tprConfig.ListType,
				&api.ListOptions{},
				&api.DeleteOptions{},
			)
			return nil
		})
	schemeBuilder.AddToScheme(api.Scheme)

	client, err := rest.RESTClientFor(kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating Kubernetes REST client")
	}

	return client, nil
}

// InitThirdPartyResource initialize third party resource if it does not exist
func InitThirdPartyResource(tprConfig *TPRConfig) error {
	client, err := NewClient(Config{})
	if err != nil {
		return err
	}

	resName := fmt.Sprintf("%s.%s", tprConfig.Name, tprConfig.GroupName)
	_, err = client.Extensions().ThirdPartyResources().Get(resName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			tpr := &v1beta1.ThirdPartyResource{
				ObjectMeta: v1.ObjectMeta{
					Name: resName,
				},
				Versions: []v1beta1.APIVersion{
					{Name: tprConfig.Version},
				},
				Description: tprConfig.Description,
			}

			result, err := client.Extensions().ThirdPartyResources().Create(tpr)
			if err != nil {
				return errors.Wrap(err, "Failed creating ThirdPartyResource")
			}
			logger.Infof("The ThridPartyResource %#v created", result)
		} else {
			return err
		}
	} else {
		logger.Infof("The ThirdPartyResource %s already exists", resName)
	}
	return nil
}

func newKubeConfig(config *Config) (*rest.Config, error) {
	var kubeConfig *rest.Config
	if config.URL != "" {
		kubeConfig = &rest.Config{
			Host:        config.URL,
			BearerToken: config.Token,
		}
	} else {
		logger.Debugf("No Kubernetes credentials provided. Attempting to load from service account")
		var err error
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "Failed loading Kubernetes credentials from service account")
		}
	}

	return kubeConfig, nil
}
