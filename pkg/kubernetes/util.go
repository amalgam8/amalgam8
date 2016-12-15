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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/amalgam8/amalgam8/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
)

// BuildServiceInstance builds an api.ServiceInstance from the given serviceName, pod, address and port.
func BuildServiceInstance(serviceName string, pod *v1.Pod, address *v1.EndpointAddress, port *v1.EndpointPort) (*api.ServiceInstance, error) {

	// Parse the service endpoint
	endpoint, err := BuildServiceEndpoint(address, port)
	if err != nil {
		return nil, err
	}

	// Determine the ID of the service instance.
	// For a pod, that would be the pod UID, followed by the port number.
	// For an externalName, that would be the IP address, followed by the port number.
	var id string
	if pod != nil {
		id = fmt.Sprintf("%s-%d", pod.UID, port.Port)
	} else {
		id = fmt.Sprintf("%s-%d", address.IP, port.Port)
	}

	// Extract the pod labels as instance tags
	var tags []string
	if pod != nil {
		tags, err = BuildServiceTags(pod.Labels)
		if err != nil {
			return nil, err
		}
	}

	// Extract the pod annotations as instance metadata
	var metadata json.RawMessage
	if pod != nil {
		metadata, err = BuildServiceMetadata(pod.Annotations)
		if err != nil {
			return nil, err
		}
	}

	var status string
	if pod != nil {
		status, err = BuildServiceStatus(&pod.Status)
		if err != nil {
			return nil, err
		}
	}

	return &api.ServiceInstance{
		ID:          id,
		ServiceName: serviceName,
		Endpoint:    *endpoint,
		Tags:        tags,
		Metadata:    metadata,
		Status:      status,
	}, nil
}

// BuildServiceEndpoint builds an api.ServiceEndpoint from the given address and port Kubernetes objects.
func BuildServiceEndpoint(address *v1.EndpointAddress, port *v1.EndpointPort) (*api.ServiceEndpoint, error) {
	var endpointType string
	endpointValue := fmt.Sprintf("%s:%d", address.IP, port.Port)

	switch port.Protocol {
	case v1.ProtocolUDP:
		endpointType = "udp"
	case v1.ProtocolTCP:
		portName := strings.ToLower(port.Name)
		switch portName {
		case "http":
			fallthrough
		case "https":
			endpointType = portName
		default:
			endpointType = "tcp"
		}
	default:
		return nil, fmt.Errorf("unsupported kubernetes endpoint port protocol: %s", port.Protocol)
	}

	return &api.ServiceEndpoint{
		Type:  endpointType,
		Value: endpointValue,
	}, nil
}

// BuildServiceTags builds a slice of string tags from the given Kubernetes resource labels.
// Each label is converted into a "key=value" string.
func BuildServiceTags(labels map[string]string) ([]string, error) {
	tags := make([]string, 0, len(labels))
	for key, value := range labels {
		tags = append(tags, fmt.Sprintf("%s=%s", key, value))
	}
	return tags, nil
}

// BuildServiceMetadata builds a serialized JSON object from the given Kubernetes resource annotations.
func BuildServiceMetadata(annotations map[string]string) (json.RawMessage, error) {
	bytes, err := json.Marshal(annotations)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(bytes), nil
}

// BuildServiceStatus builds a service instance status string for the given pod.
func BuildServiceStatus(podStatus *v1.PodStatus) (string, error) {
	// Stub implementation that assumes that only healthy pods get here,
	// and otherwise would have been removed by the endpoint controller
	// TODO: implement for real
	return "UP", nil
}
