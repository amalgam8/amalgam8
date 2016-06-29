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

// Package health provides functionality to add application-level health checks to individual application components.
// Health check results can be exposed using an HTTP route and handler.
// Components are expected to register health checkers, either as Checker interfaces or CheckerFunc using Register()
// and RegisterFunc() respectively.
// Registration can be done in package init() function, or explicitly when the component is created.
// The health check returns a "binary" healthy/unhealthy status and may add additional message.
// An unhealthy component may optional add a root cause, typically an error value returned by some internal check
// procedure.
// The HTTP handler is added by attaching health.Handler() to a route. The returned body is a JSON encoding of a map
// of components to their corresponding health status.
package health
