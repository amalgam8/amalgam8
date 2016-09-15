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

package eureka

import (
	"encoding/json"
	"fmt"
)

// Application is an array of instances
type Application struct {
	Name      string      `json:"name,omitempty"`
	Instances []*Instance `json:"instance,omitempty"`
}

// This struct is used ONLY for the specific UnmarshalJSON
type singleApplication struct {
	Name     string    `json:"name,omitempty"`
	Instance *Instance `json:"instance,omitempty"`
}

// This struct is used ONLY for the specific UnmarshalJSON
type multiApplication struct {
	Name      string      `json:"name,omitempty"`
	Instances []*Instance `json:"instance,omitempty"`
}

// UnmarshalJSON parses the JSON object of Application struct.
// We need this specific implementation because the Eureka server
// marshals differently single instance (object) and multiple instances (array).
func (app *Application) UnmarshalJSON(b []byte) error {
	var mApp multiApplication
	err := json.Unmarshal(b, &mApp)
	if err != nil {
		// error probably means that we have a single instance object.
		// Thus, we try to unmarshal to a different object type
		var sApp singleApplication
		err = json.Unmarshal(b, &sApp)
		if err != nil {
			fmt.Println(string(b[:20]))
			return err
		}
		app.Name = sApp.Name
		if sApp.Instance != nil {
			app.Instances = []*Instance{sApp.Instance}
		}
		return nil
	}

	app.Name = mApp.Name
	app.Instances = mApp.Instances
	return nil
}
