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

// Applications is an array of application objects
type Applications struct {
	VersionDelta int64
	Hashcode     string
	Application  []*Application
}

// This struct is used ONLY for the specific UnmarshalJSON
type singleApplications struct {
	VersionDelta int64        `json:"versions__delta,omitempty"`
	Hashcode     string       `json:"apps__hashcode,omitempty"`
	Application  *Application `json:"application,omitempty"`
}

// This struct is used ONLY for the specific UnmarshalJSON
type multiApplications struct {
	VersionDelta int64          `json:"versions__delta,omitempty"`
	Hashcode     string         `json:"apps__hashcode,omitempty"`
	Application  []*Application `json:"application,omitempty"`
}

// UnmarshalJSON parses the JSON object of Applications struct.
// We need this specific implementation because the Eureka server
// marshals differently single application (object) and multiple applications (array).
func (apps *Applications) UnmarshalJSON(b []byte) error {
	var mApps multiApplications
	err := json.Unmarshal(b, &mApps)
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(b[:2048]))
		// error probably means that we have a single Application object.
		// Thus, we try to unmarshal to a different object type
		var sApps singleApplications
		err = json.Unmarshal(b, &sApps)
		if err != nil {
			return err
		}
		apps.Hashcode = sApps.Hashcode
		apps.VersionDelta = sApps.VersionDelta
		if sApps.Application != nil {
			apps.Application = []*Application{sApps.Application}
		}
		return nil
	}

	apps.Hashcode = mApps.Hashcode
	apps.VersionDelta = mApps.VersionDelta
	apps.Application = mApps.Application
	return nil
}
