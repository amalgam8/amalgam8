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

package register

import (
	"time"

	"github.com/amalgam8/amalgam8/pkg/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Registration agent", func() {
	mockClient := &mockRegistryClient{}
	mockProvider := &mockIdentityProvider{}

	config := RegistrationConfig{
		Registry: mockClient,
		Identity: mockProvider,
	}

	var agent *RegistrationAgent
	var err error

	BeforeEach(func() {
		mockClient.Reset()

		agent, err = NewRegistrationAgent(config)
		Expect(err).To((BeNil()))

		agent.Start()
	})

	Context("When registration agent is started", func() {

		It("Registers the service with the registry", func() {
			// Avoid race condition on registration
			time.Sleep(100 * time.Millisecond)

			Expect(mockClient.registered).To(BeTrue())
		})

		It("Continuously renews the registration", func() {
			// Avoid race condition on registration
			time.Sleep(100 * time.Millisecond)

			ttl := time.Duration(mockProvider.MustGetIdentity().TTL) * time.Second
			for i := 0; i < 3; i++ {
				// Assert that last heartbeat took place at most <TTL> ago
				Expect(mockClient.lastHeartbeat).To(BeTemporally("~", time.Now(), ttl))
				time.Sleep(ttl)
			}
		})
	})

	Context("When registration agent is stopped", func() {

		BeforeEach(func() {
			agent.Stop()
		})

		It("Deregisters the service with the registry", func() {
			Expect(mockClient.registered).To(BeFalse())
		})

		It("Can be started again", func() {
			agent.Start()
			time.Sleep(250 * time.Millisecond)

			Expect(mockClient.registered).To(BeTrue())
		})
	})

})

type mockRegistryClient struct {
	registered    bool
	lastHeartbeat time.Time
}

func (c *mockRegistryClient) Register(instance *api.ServiceInstance) (*api.ServiceInstance, error) {
	c.registered = true
	c.lastHeartbeat = time.Now()
	return &api.ServiceInstance{
		ID:            "1234567890",
		ServiceName:   instance.ServiceName,
		Endpoint:      instance.Endpoint,
		Status:        instance.Status,
		Tags:          instance.Tags,
		Metadata:      instance.Metadata,
		TTL:           instance.TTL,
		LastHeartbeat: c.lastHeartbeat,
	}, nil
}

func (c *mockRegistryClient) Deregister(id string) error {
	c.registered = false
	return nil
}

func (c *mockRegistryClient) Renew(id string) error {
	c.lastHeartbeat = time.Now()
	return nil
}

func (c *mockRegistryClient) Reset() {
	c.registered = false
}

type mockIdentityProvider struct{}

func (mip mockIdentityProvider) GetIdentity() (*api.ServiceInstance, error) {
	return mip.MustGetIdentity(), nil
}

func (mip mockIdentityProvider) MustGetIdentity() *api.ServiceInstance {
	return &api.ServiceInstance{
		ServiceName: "test_service",
		Endpoint: api.ServiceEndpoint{
			Type:  "http",
			Value: "http://172.17.0.10:8080",
		},
		TTL: 1,
	}
}
