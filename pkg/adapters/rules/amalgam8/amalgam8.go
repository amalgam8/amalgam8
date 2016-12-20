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

package amalgam8

import (
	"time"

	"github.com/amalgam8/amalgam8/controller/client"
	"github.com/amalgam8/amalgam8/pkg/adapters/rules/cache"
	"github.com/amalgam8/amalgam8/pkg/api"
)

// ControllerConfig stores the configurable attributes of the Amalgam8 Controller adapter.
type ControllerConfig client.Config

// NewRulesAdapter constructs a new RuleService adapter
// for the Amalgam8 Controller using the given configuration.
func NewRulesAdapter(config ControllerConfig) (api.RulesService, error) {
	return client.New(client.Config(config))
}

// NewCachedRulesAdapter constructs a new RuleService adapter
// for the Amalgam8 Controller using the given configuration, and a local
// cache refreshed at the frequency specified by the given poll interval.
func NewCachedRulesAdapter(config ControllerConfig, pollInterval time.Duration) (api.RulesService, error) {
	controller, err := NewRulesAdapter(config)
	if err != nil {
		return nil, err
	}

	return cache.New(controller, pollInterval)
}
