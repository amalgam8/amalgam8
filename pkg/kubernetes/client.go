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
	"github.com/amalgam8/amalgam8/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Config stores configurable attributes of a Kubernetes client.
type Config struct {
	URL   string
	Token string
}

// NewClient creates a new Kubernetes client using the specified configuration.
// If no URL and Token are specified, then these values are attempted to be read from the
// service account (if running within a Kubernetes pod).
func NewClient(config Config) (kubernetes.Interface, error) {
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

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating Kubernetes REST client")
	}
	return client, nil
}
