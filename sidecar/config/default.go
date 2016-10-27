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

package config

import "time"

// DefaultConfig defines default values for the various configuration options
var DefaultConfig = Config{
	Register: false,
	Proxy:    false,

	Service: Service{
		Name: "",
		Tags: nil,
	},
	Endpoint: Endpoint{
		Host: "",
		Port: 0,
		Type: "http",
	},

	Registry: Registry{
		URL:   "",
		Token: "",
		Poll:  time.Duration(15 * time.Second),
	},
	Controller: Controller{
		URL:   "",
		Token: "",
		Poll:  time.Duration(15 * time.Second),
	},

	HealthChecks: nil,

	LogLevel: "info",

	Commands: nil,
}
