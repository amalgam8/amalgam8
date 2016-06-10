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

package uptime

import "github.com/ant0ine/go-json-rest/rest"

// RouteHandlers returns an array of uptime route handlers
func RouteHandlers() []*rest.Route {
	return []*rest.Route{
		rest.Get(URL(), uptimeHandler),
		rest.Get(HealthyURL(), healthyHandler),
	}
}
