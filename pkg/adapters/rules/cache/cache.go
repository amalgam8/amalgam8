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

package cache

import (
	"time"

	"sync"

	"github.com/amalgam8/amalgam8/pkg/api"
)

// Make sure we implement the ServiceRules  interface.
var _ api.RulesService = (*Cache)(nil)

// An empty filter used to query for all rules
var emptyFilter = api.RuleFilter{}

// Cache implements the ServiceRules interface using a local cache.
// The cache is refreshed periodically using the non-caching, REST API-based Amalagam8 Controller client.
type Cache struct {
	rules api.RulesService
	cache api.RulesSet
	mutex sync.RWMutex
}

// New constructs a new Cache.
// The cache is refreshed at the frequency specified by pollInterval using the provided RulesService object.
func New(rules api.RulesService, pollInterval time.Duration) (*Cache, error) {
	c := &Cache{
		rules: rules,
		cache: api.RulesSet{},
	}

	go c.maintain(pollInterval)
	return c, nil
}

// ListRules queries for the list of rules.
func (c *Cache) ListRules(filter *api.RuleFilter) (*api.RulesSet, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if filter == nil {
		filter = &emptyFilter
	}
	filteredRules := filter.Apply(c.cache.Rules)

	filteredRuleSet := api.RulesSet{
		Rules:    filteredRules,
		Revision: c.cache.Revision,
	}

	return &filteredRuleSet, nil
}

func (c *Cache) maintain(pollInterval time.Duration) {
	go c.refresh()
	for range time.Tick(pollInterval) {
		go c.refresh()
	}
}

func (c *Cache) refresh() {
	ruleList, err := c.rules.ListRules(&emptyFilter)
	if err != nil {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = *ruleList
}
