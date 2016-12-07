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

package client

import (
	"time"

	"sync"

	"github.com/amalgam8/amalgam8/pkg/api"
)

// Make sure we implement the ServiceRules  interface.
var _ api.RulesService = (*Cache)(nil)

// CacheConfig stores the configurable attributes of the caching client.
type CacheConfig struct {
	Config

	// PollInterval is the time interval in which the caching client refreshes its local cache
	PollInterval time.Duration
}

// Cache implements the ServiceRules interface using a local cache.
// The cache is refreshed periodically using the non-caching, REST API-based Amalagam8 Controller client.
type Cache struct {
	client api.RulesService
	cache api.RulesSet
	mutex sync.RWMutex
	pollCount int64
}

// NewCache constructs a new Caching Client using the given configuration.
func NewCache(config CacheConfig) (*Cache, error) {

	cl, err := New(config.Config)
	if err != nil {
		return nil, err
	}

	c := &Cache{
		client: cl,
		cache:  api.RulesSet{},
	}

	go c.maintain(config.PollInterval)
	return c, nil
}

// ListRules queries for the list of rules.
func (c *Cache) ListRules(filter *api.RuleFilter) (*api.RulesSet, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	filteredRules :=api.FilterRules(*filter, c.cache.Rules)

	filteredRuleSet :=api.RulesSet {Rules:filteredRules,
		                        Revision:c.cache.Revision}

	return &filteredRuleSet, nil
}


func (c *Cache) maintain(pollInterval time.Duration) {
	go c.refresh()
	for range time.Tick(pollInterval) {
		go c.refresh()
	}
}

func (c *Cache) refresh() {
	noFilter := api.RuleFilter{}
	ruleList, err := c.client.ListRules(&noFilter)
	if err != nil {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = *ruleList
	c.pollCount+=1
}
